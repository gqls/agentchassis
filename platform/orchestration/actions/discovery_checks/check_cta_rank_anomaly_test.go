// FILE: platform/orchestration/actions/discovery_checks/check_cta_rank_anomaly_test.go
//
// The predicate is tested over ALREADY-RANKED input because the ranking it
// consumes is datahelpers.RankCTAPositionalCandidates — the writers' own,
// tested where it lives — and this check must add no ordering opinion of its
// own. Each negative case here is a REAL estate state, named, so a threshold
// edit that un-silences one of them has to argue with the state it re-flags.

package discovery_checks

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func rankedTools(navOrders ...int) []datahelpers.CTAPositionalCandidate {
	names := []string{"tool-a", "tool-b", "tool-c", "tool-d"}
	out := make([]datahelpers.CTAPositionalCandidate, 0, len(navOrders))
	for i, n := range navOrders {
		out = append(out, datahelpers.CTAPositionalCandidate{
			Name: names[i], URL: "/tools/" + names[i] + ".html", Area: "tools", NavOrder: n,
		})
	}
	return out
}

func TestCTARankAnomalyPredicate(t *testing.T) {
	cases := []struct {
		name string
		in   []datahelpers.CTAPositionalCandidate
		want bool
	}{
		// 391's fossil: nav_order 1 set at page creation 2026-03-13, pack at
		// the default. The case this check exists for.
		{"the 391 fossil fires", rankedTools(1, 100, 100), true},
		// The same sites after the owner's demotion (1 -> 900): pack leads.
		{"the demoted state is silent", rankedTools(100, 100, 900), false},
		// webdesign's shape: every tool at the default, winner is alphabetical.
		// Arbitrary but NOT anomalous — candidate 3's business, out of scope.
		{"all-default is silent", rankedTools(100, 100, 100, 100), false},
		// A deliberate curated ordering must not be second-guessed.
		{"a curated ladder is silent", rankedTools(10, 20, 30), false},
		// Two curated leaders: unique minimum but no fossil-sized lead.
		{"two close leaders are silent", rankedTools(1, 2, 100), false},
		// "Anomalous against its siblings" needs siblings.
		{"two candidates are too few to judge", rankedTools(1, 100), false},
		{"an empty ranking is silent", nil, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, detail := ctaRankAnomaly(tc.in)
			if got != tc.want {
				t.Errorf("ctaRankAnomaly(%v) = %v (%s), want %v", tc.in, got, detail, tc.want)
			}
			if detail == "" {
				t.Error("detail must always be populated — it is the Resolved reason and the work item body")
			}
		})
	}
}

// TestCTARankAnomalySilencedByTheLever pins the pairing the owner approved:
// setting eligible_as_cta_target=false on the fossil removes it from the
// RANKING (the lever), and the alarm then observes a healthy rank-1 — correct
// silencing, because the flagged page can no longer be the primary button.
// The ranking is the real one, not a re-implementation, so this test also
// fails if the check ever grows its own candidate filtering.
func TestCTARankAnomalySilencedByTheLever(t *testing.T) {
	supply := []datahelpers.CTAPositionalCandidate{
		{Name: "tool-fossil", URL: "/tools/fossil.html", Area: "tools", NavOrder: 1},
		{Name: "tool-b", URL: "/tools/b.html", Area: "tools", NavOrder: 100},
		{Name: "tool-c", URL: "/tools/c.html", Area: "tools", NavOrder: 100},
	}
	if got, _ := ctaRankAnomaly(datahelpers.RankCTAPositionalCandidates("", supply)); !got {
		t.Fatal("fixture is inert: the fossil must fire before the lever is pulled")
	}
	supply[0].IneligibleAsCTATarget = true
	anomalous, detail := ctaRankAnomaly(datahelpers.RankCTAPositionalCandidates("", supply))
	if anomalous {
		t.Errorf("opting the fossil out must silence the alarm (the condition is gone, "+
			"not blinded): %s", detail)
	}
	if !strings.Contains(detail, "candidate") {
		t.Errorf("detail should describe the healthy observation, got %q", detail)
	}
}

// --- the acknowledgement (migration 750, owner ruling 2026-09-03) ---------
//
// These test ctaRankAcknowledged against a mocked DB rather than the pure
// predicate, because the whole mechanism IS the SQL: the self-expiry lives in
// `cta_rank_deliberate_nav_order = COALESCE(nav_order, 100)`, and a test that
// re-implemented that comparison in Go would pass while the column silenced
// the wrong shapes.

func ackContext(t *testing.T, db *sql.DB, siteID uuid.UUID) DiscoveryCheckContext {
	t.Helper()
	return DiscoveryCheckContext{
		Ctx: context.Background(), DB: db, SiteID: siteID, Logger: zap.NewNop(),
	}
}

// TestCTARankAcknowledgedHonoursAMatchingAcknowledgement — the boxingonline
// case: a person accepted the fight calendar at nav_order 3, so the row comes
// back and the caller must not file.
func TestCTARankAcknowledgedHonoursAMatchingAcknowledgement(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	mock.ExpectQuery(`cta_rank_deliberate_nav_order`).
		WithArgs(site, "tool-fight-calendar").
		WillReturnRows(sqlmock.NewRows([]string{"cta_rank_deliberate_nav_order"}).AddRow(3))

	nav, ok, err := ctaRankAcknowledged(ackContext(t, db, site), "tool-fight-calendar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok || nav != 3 {
		t.Errorf("got (%d, %v), want (3, true)", nav, ok)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// TestCTARankAcknowledgedIsAbsentByDefault — the state of every row in the
// estate the day 750 applies. If absence silenced the check, applying the
// migration would blind the alarm fleet-wide with no diff to show it.
func TestCTARankAcknowledgedIsAbsentByDefault(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	mock.ExpectQuery(`cta_rank_deliberate_nav_order`).
		WithArgs(site, "tool-example").
		WillReturnError(sql.ErrNoRows)

	nav, ok, err := ctaRankAcknowledged(ackContext(t, db, site), "tool-example")
	if err != nil {
		t.Fatalf("ErrNoRows must not surface as an error: %v", err)
	}
	if ok || nav != 0 {
		t.Errorf("got (%d, %v), want (0, false) — an unacknowledged page must stay flagged", nav, ok)
	}
}

// TestCTARankAcknowledgedRefusesToSwallowALookupFailure — a DB error must NOT
// read as "not acknowledged". That direction is the safe-looking one and it is
// wrong: it re-files a finding a human already retired, on every pass, with
// nothing in the item to say the lookup failed.
func TestCTARankAcknowledgedRefusesToSwallowALookupFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	mock.ExpectQuery(`cta_rank_deliberate_nav_order`).
		WithArgs(site, "tool-x").
		WillReturnError(errors.New("connection reset"))

	if _, ok, err := ctaRankAcknowledged(ackContext(t, db, site), "tool-x"); err == nil || ok {
		t.Errorf("a lookup failure must return an error and ok=false, got ok=%v err=%v", ok, err)
	}
}

// TestCTARankAcknowledgementSelfExpiresInSQL is the one that protects the
// DESIGN rather than the plumbing. The acknowledgement is scoped to the
// nav_order it was granted at, and that scoping is expressed ONLY in the
// query's WHERE clause. A future "simplification" to a bare
// `cta_rank_deliberate_nav_order IS NOT NULL` would pass every other test in
// this file while silently converting a shape-specific acknowledgement into a
// permanent per-page mute — an acknowledgement outliving what it acknowledged.
func TestCTARankAcknowledgementSelfExpiresInSQL(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	site := uuid.New()
	// The regexp IS the assertion: the query must compare the stored value
	// against the page's live COALESCE(nav_order, 100).
	mock.ExpectQuery(`cta_rank_deliberate_nav_order\s*=\s*COALESCE\(nav_order,\s*100\)`).
		WithArgs(site, "tool-fight-calendar").
		WillReturnError(sql.ErrNoRows)

	if _, ok, err := ctaRankAcknowledged(ackContext(t, db, site), "tool-fight-calendar"); err != nil || ok {
		t.Fatalf("got ok=%v err=%v", ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the acknowledgement must be scoped to the reviewed nav_order in SQL: %v", err)
	}
}
