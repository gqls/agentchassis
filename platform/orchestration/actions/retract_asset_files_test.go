// FILE: platform/orchestration/actions/retract_asset_files_test.go
//
// The guard chain is the design (file header of the action), so the tests
// exercise the guards, not the plumbing. Construction mirrors the sibling's
// deploy-path refusal tests: Producer is nil on purpose, so any test that
// finishes without a "kafka producer not available" error has ALSO proven no
// dispatch was reached — and the one test that arms the deletion proves the
// opposite ordering by hitting exactly that error after every guard passed.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

const testSiteID = "5fe15466-4e2e-4ff2-981e-98c1b7074002"

// TestAssetPathScopeViolation pins guard 1 as a table, because the scope
// check is the only guard that runs on the raw string and every bypass class
// has its own spelling.
func TestAssetPathScopeViolation(t *testing.T) {
	cases := []struct {
		path   string
		wantOK bool
	}{
		{"/assets/images/logo.jpg", true},
		{"/assets/data/x.json", true},
		{"/index.html", false},                 // a page
		{"/data/adoption-tracker.json", false}, // a feed
		{"assets/images/logo.jpg", false},      // not site-absolute
		{"/assets/../index.html", false},       // traversal
		{"/assets//images/logo.jpg", false},    // does not survive cleaning
		{"/assets/images/", false},             // a directory
		{"/assets", false},                     // the root itself
		{"https://x.com/assets/a.jpg", false},
		{"/assets/images/logo.jpg?v=2", false},
		{"/assets/images/logo.jpg#f", false},
	}
	for _, c := range cases {
		reason := assetPathScopeViolation(c.path)
		if c.wantOK && reason != "" {
			t.Errorf("%q refused: %s — this is an admissible asset path", c.path, reason)
		}
		if !c.wantOK && reason == "" {
			t.Errorf("%q ADMITTED — scope guard 1 has a hole; this class must never reach the adapter", c.path)
		}
	}
}

// standardMockPreamble expects the site-domain resolve and the owned-assets
// load that every run performs, in order.
func standardMockPreamble(mock sqlmock.Sqlmock, ownedRows *sqlmock.Rows) {
	mock.ExpectQuery(`SELECT domain FROM sites`).
		WillReturnRows(sqlmock.NewRows([]string{"domain"}).AddRow("gaswholesalers.com"))
	mock.ExpectQuery(`FROM assets`).WillReturnRows(ownedRows)
}

// TestRetractAssetFilesRefusesAnOwnedDerivedPath pins guard 2: a path that a
// non-deleted asset row derives is the platform's artefact, and the row is
// the authority. logo.png is exactly the live case this action was built
// beside — deleting it would be the regression the 248 drain twice avoided.
func TestRetractAssetFilesRefusesAnOwnedDerivedPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	standardMockPreamble(mock, sqlmock.NewRows([]string{"id", "asset_key", "purpose"}).
		AddRow("b99c5355-4b3a-430c-9294-56482726be34", "logo", "logo"))

	out, err := RetractAssetFilesAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		DB:               db,
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": testSiteID,
			"paths":   []interface{}{"/assets/images/logo.png"},
		}},
		CollectedData: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("action errored: %v", err)
	}
	res := out.(map[string]interface{})
	refusals := res["refusals"].([]assetRetractionRefusal)
	if len(refusals) != 1 || !strings.Contains(refusals[0].Refused, "b99c5355") {
		t.Fatalf("want one refusal naming the owning row, got %+v", refusals)
	}
	if got := res["retracted"].(int); got != 0 {
		t.Errorf("retracted = %d, want 0 — an owned path must never survive to the batch", got)
	}
}

// TestRetractAssetFilesRefusesABrandHeadPath pins guard 3. Brand-head files
// may legitimately exist with NO asset row (the deriver writes fixed names),
// so guard 2 alone would admit them — that is why this guard exists
// separately, and why the mock returns zero owned rows.
func TestRetractAssetFilesRefusesABrandHeadPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	standardMockPreamble(mock, sqlmock.NewRows([]string{"id", "asset_key", "purpose"}))

	out, err := RetractAssetFilesAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		DB:               db,
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": testSiteID,
			"paths":   []interface{}{"/assets/images/favicon.png"},
		}},
		CollectedData: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("action errored: %v", err)
	}
	res := out.(map[string]interface{})
	refusals := res["refusals"].([]assetRetractionRefusal)
	if len(refusals) != 1 || !strings.Contains(refusals[0].Refused, "brand-head") {
		t.Fatalf("favicon.png not refused as brand-head with no row present, got %+v", refusals)
	}
}

// TestRetractAssetFilesRefusesAReferencedPath pins guard 4 — the check that
// caught both of the 248 drain's would-be regressions: a linked 200 is live
// whatever the assets table says.
func TestRetractAssetFilesRefusesAReferencedPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	standardMockPreamble(mock, sqlmock.NewRows([]string{"id", "asset_key", "purpose"}))
	mock.ExpectQuery(`FROM page_components`).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("index"))

	out, err := RetractAssetFilesAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		DB:               db,
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": testSiteID,
			"paths":   []interface{}{"/assets/images/orphan.jpg"},
		}},
		CollectedData: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("action errored: %v", err)
	}
	res := out.(map[string]interface{})
	refusals := res["refusals"].([]assetRetractionRefusal)
	if len(refusals) != 1 || !strings.Contains(refusals[0].Refused, "index") {
		t.Fatalf("referenced path not refused with its referrer named, got %+v", refusals)
	}
}

// TestRetractAssetFilesDryRunIsTheDefault pins guard 5's default direction.
// Producer is nil, so if the silent default were "delete", this test would
// error at the dispatch — finishing cleanly with dry_run:true in the result
// is the proof that saying nothing audits and touches nothing.
func TestRetractAssetFilesDryRunIsTheDefault(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	standardMockPreamble(mock, sqlmock.NewRows([]string{"id", "asset_key", "purpose"}))
	mock.ExpectQuery(`FROM page_components`).
		WillReturnRows(sqlmock.NewRows([]string{"name"}))

	out, err := RetractAssetFilesAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		DB:               db,
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": testSiteID,
			"paths":   []interface{}{"/assets/images/orphan.jpg"},
			// no dry_run key at all — absence must mean TRUE
		}},
		CollectedData: map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("a config with no dry_run key reached the dispatch: %v\n"+
			"The deleting branch must be opt-in (owner ruling 2026-08-02 §2) — absence arms the AUDIT, never the deletion.", err)
	}
	res := out.(map[string]interface{})
	if res["dry_run"] != true {
		t.Errorf("dry_run = %v in the result, want true by default", res["dry_run"])
	}
	if got := res["retracted"].(int); got != 1 {
		t.Errorf("retracted = %d, want 1 — the dry run must still report what a real run WOULD delete", got)
	}
}

// TestRetractAssetFilesArmedRunReachesDispatchOnlyAfterEveryGuard is the
// ordering proof, the same nil-dependency construction as the sibling's: with
// dry_run explicitly false and a clean orphan path, the FIRST error must be
// the nil producer — meaning every guard, the audit write and the batch
// assembly all ran first, and nothing before the dispatch can delete.
func TestRetractAssetFilesArmedRunReachesDispatchOnlyAfterEveryGuard(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	standardMockPreamble(mock, sqlmock.NewRows([]string{"id", "asset_key", "purpose"}))
	mock.ExpectQuery(`FROM page_components`).
		WillReturnRows(sqlmock.NewRows([]string{"name"}))
	// The audit INSERTs are best-effort and unexpected by the mock — that is
	// fine: agenterrors.Write swallows the mock's refusal into `false`.

	_, err = RetractAssetFilesAction(context.Background(), ActionParams{
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		DB:               db,
		StepConfig: models.Step{Config: map[string]interface{}{
			"site_id": testSiteID,
			"paths":   []interface{}{"/assets/images/orphan.jpg"},
			"dry_run": false,
		}},
		CollectedData: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("an armed run with a nil producer returned nil error — the dispatch was never attempted, or something swallowed it")
	}
	if !strings.Contains(err.Error(), "producer not available") {
		t.Fatalf("first error is %q, want the nil-producer error — anything else means a guard or the audit errored AFTER arming, i.e. the ordering moved", err)
	}
}
