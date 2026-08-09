// FILE: platform/orchestration/actions/discovery_checks/check_premise_incomplete_test.go
//
// The three states the predicate must separate, plus the two guards that make
// retraction safe: greenfield is neither a finding nor a retraction, and a read
// error is an error (never an empty result — the runner's skip-Resolved-on-error
// safety depends on checks not swallowing this).

package discovery_checks

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func premiseCtx(t *testing.T) (DiscoveryCheckContext, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return DiscoveryCheckContext{
		Ctx:       context.Background(),
		DB:        db,
		SiteID:    uuid.New(),
		Pipeline:  "content",
		AgentType: "quality-discovery-agent",
		BatchID:   uuid.New(),
		Logger:    zap.NewNop(),
	}, mock
}

func premiseRows(domain string, deployed int, hasStrategy bool, primary interface{}) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"domain", "count", "exists", "primary_model"}).
		AddRow(domain, deployed, hasStrategy, primary)
}

func TestPremiseIncomplete_MissingStrategyOnDeployedSiteFiles(t *testing.T) {
	dctx, mock := premiseCtx(t)
	mock.ExpectQuery(`SELECT s\.domain`).
		WillReturnRows(premiseRows("loancalculator.co.uk", 27, false, nil))

	res, err := (&PremiseIncompleteCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}
	wi := res.WorkItems[0]
	if wi.ItemType != "needs_strategy" || wi.HandlerAgent != "domain-strategist" {
		t.Errorf("routing: got %s -> %s", wi.ItemType, wi.HandlerAgent)
	}
	// The SHARED key shape (RFC_010 §1): must match vertical-exemplar-researcher's
	// strategy_<domain>, so the dedup index holds one open request per site
	// whichever producer files first.
	if wi.ItemKey != "strategy_loancalculator.co.uk" {
		t.Errorf("item_key = %q, want the shared strategy_<domain> shape", wi.ItemKey)
	}
	if len(res.Resolved) != 0 {
		t.Errorf("a finding must not also retract; Resolved = %v", res.Resolved)
	}
}

func TestPremiseIncomplete_OldShapeWithoutPrimaryModelFiles(t *testing.T) {
	dctx, mock := premiseCtx(t)
	// The gaswholesalers case: strategy row exists, revenue_models block absent.
	mock.ExpectQuery(`SELECT s\.domain`).
		WillReturnRows(premiseRows("gaswholesalers.com", 12, true, nil))

	res, err := (&PremiseIncompleteCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 1 {
		t.Fatalf("want 1 work item, got %d", len(res.WorkItems))
	}
}

func TestPremiseIncomplete_CompletePremiseRetractsByKey(t *testing.T) {
	dctx, mock := premiseCtx(t)
	mock.ExpectQuery(`SELECT s\.domain`).
		WillReturnRows(premiseRows("relojistas.com", 21, true, "display_advertising"))

	res, err := (&PremiseIncompleteCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 {
		t.Fatalf("complete premise must file nothing, got %d items", len(res.WorkItems))
	}
	if len(res.Resolved) != 1 {
		t.Fatalf("want 1 narrow retraction, got %d", len(res.Resolved))
	}
	r := res.Resolved[0]
	if r.ItemType != "needs_strategy" || r.ItemKey != "strategy_relojistas.com" || r.AllOfType {
		t.Errorf("retraction must be narrow by shared key: %+v", r)
	}
	if r.Reason == "" {
		t.Error("retraction without a reason is indistinguishable from a hand-close later")
	}
}

func TestPremiseIncomplete_GreenfieldIsNeitherFindingNorRetraction(t *testing.T) {
	dctx, mock := premiseCtx(t)
	// Site row exists, no deployed pages, no strategy: the build chain owns this.
	mock.ExpectQuery(`SELECT s\.domain`).
		WillReturnRows(premiseRows("lendzy.co.uk", 0, false, nil))

	res, err := (&PremiseIncompleteCheck{}).Run(dctx)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.WorkItems) != 0 || len(res.Resolved) != 0 {
		t.Errorf("greenfield must be silent both ways: items=%d resolved=%d",
			len(res.WorkItems), len(res.Resolved))
	}
}

func TestPremiseIncomplete_ReadErrorIsAnErrorNotAnEmptyResult(t *testing.T) {
	dctx, mock := premiseCtx(t)
	mock.ExpectQuery(`SELECT s\.domain`).WillReturnError(context.DeadlineExceeded)

	_, err := (&PremiseIncompleteCheck{}).Run(dctx)
	if err == nil {
		t.Fatal("a failed read must return error — an empty result here would look like a healthy site AND license retraction skipping")
	}
}
