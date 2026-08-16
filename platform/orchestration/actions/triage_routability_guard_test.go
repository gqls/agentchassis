// FILE: platform/orchestration/actions/triage_routability_guard_test.go
//
// bugs_open/284 — the promoter promoted rows the claim path could only block.
//
// The platform files two kinds of finding. Most are JOBS: a named agent can fix
// them, the row carries a handler_agent, the dispatch loop routes it. Some are
// FLAGS: nothing on the platform can act on them — a brand palette nothing can
// repaint, a VM nothing can restart, an image reference nothing can repoint — so
// the row names no handler ON PURPOSE and waits for a human.
//
// TriageDetectedItemsAction promoted both, because it filtered on neither field.
// The dispatch loop then claimed the flag rows, found nobody to send them to, and
// stamped them `blocked` with "No handler_agent set — item cannot be routed to any
// agent": a correct finding rewritten as a routing failure. Permanently, because
// `blocked` is not terminal (it holds the dedup slot, so the check can never file
// that finding again) and feasibility-recheck can only release a row whose handler
// is a registered agent — and no agent type is the empty string. Measured
// 2026-08-16: 60 rows, 4 item_types, 15+ sites.
//
// These tests pin the guard and, more importantly, pin that the promoter and the
// claim path apply THE SAME test. Both render it from workItemHandlerRegisteredSQL,
// so they cannot drift by editing one; what these tests catch is someone
// hand-writing either one back into existence.
package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestPromotionRefusesWhatTheClaimPathWouldBlock is the guard itself. The
// promoting UPDATE must carry both halves of the routability test; a row that
// names no handler, or names an unregistered one, must not be promoted.
//
// Load-bearing check: delete `AND %s` from the UPDATE in
// triage_detect_items_action.go and this test fails on both clauses.
func TestPromotionRefusesWhatTheClaimPathWouldBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()

	// The expectation IS the assertion. sqlmock matches the query it is given as
	// a regexp against the SQL the action actually runs, so a promoting UPDATE
	// that has lost either half of the routability test no longer matches and the
	// call fails here — which is the only way to assert this that a mutation can
	// falsify. (Re-rendering the predicate from its own helper and comparing would
	// pass with the guard deleted from the query entirely; that was the first
	// version of this test and it proved nothing.)
	mock.ExpectExec(`UPDATE site_work_items[\s\S]*status = 'detected'[\s\S]*`+
		`COALESCE\(wi\.handler_agent, ''\) <> ''[\s\S]*FROM agent_definitions`).
		WithArgs(siteID, "build").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectQuery(`SELECT wi\.item_type, count\(\*\)`).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"item_type", "count"}).
			AddRow("capability_gap", int64(1)).
			AddRow("image_url_404", int64(3)))
	mock.ExpectQuery(`SELECT count\(\*\)`).
		WithArgs(siteID, "build").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	// sqlmock's default matcher is a regexp over the query text, which is also
	// how we get to READ the query the action actually ran.
	mock.MatchExpectationsInOrder(true)

	params := ActionParams{
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "triage_findings"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    map[string]interface{}{"site_id": siteID.String()},
		Logger:           zap.NewNop(),
		DB:               db,
	}
	out, err := TriageDetectedItemsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("action returned an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}

	result, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("expected a map result, got %T", out)
	}
	// Held-back rows are reported, never silently dropped: a filter that quietly
	// promotes fewer rows is indistinguishable from a site with less work.
	if result["not_promotable"] != int64(4) {
		t.Errorf("not_promotable = %v, want 4 (1 capability_gap + 3 image_url_404)", result["not_promotable"])
	}
	byType, ok := result["not_promotable_by_type"].(map[string]int64)
	if !ok {
		t.Fatalf("not_promotable_by_type missing or wrong type: %T", result["not_promotable_by_type"])
	}
	if byType["capability_gap"] != 1 || byType["image_url_404"] != 3 {
		t.Errorf("the breakdown must name the types being held back, got %v", byType)
	}
}

// TestClaimAndPromoterAskTheSameQuestion pins the coupling. The claim path's
// registration check and the promoter's guard must be the SAME predicate — if
// they diverge, one of two things happens and both are bad: a stricter promoter
// strands rows nothing else will ever promote (the scheduled detected-item-promoter
// is stricter still), and a stricter CLAIM re-opens exactly this bug.
//
// Load-bearing check: hand-write claim's EXISTS query back (e.g. add
// `AND is_active`) and this test fails, because the expectation is built from the
// shared renderer.
func TestClaimAndPromoterAskTheSameQuestion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	itemID := uuid.New()

	mock.ExpectQuery(`UPDATE site_work_items`).
		WithArgs(itemID, "dispatch-loop").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(itemID.String()))
	mock.ExpectQuery(`SELECT handler_agent FROM site_work_items`).
		WithArgs(itemID).
		WillReturnRows(sqlmock.NewRows([]string{"handler_agent"}).AddRow("page-build-handler"))

	// THE assertion: the registration check claim runs must be byte-identical to
	// what workItemHandlerRegisteredSQL renders — the same function the promoter's
	// guard is built from.
	expected := "SELECT " + workItemHandlerRegisteredSQL("$1")
	mock.ExpectQuery(regexp.QuoteMeta(expected)).
		WithArgs("page-build-handler").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

	// The AI-endpoint health lookup that follows on the happy path.
	mock.ExpectQuery(`SELECT default_config FROM agent_definitions`).
		WithArgs("page-build-handler").
		WillReturnRows(sqlmock.NewRows([]string{"default_config"}).AddRow([]byte(`{}`)))

	params := ActionParams{
		ExecutionContext: &orchtypes.ExecutionContext{StepName: "claim_item"},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    map[string]interface{}{"work_item_id": itemID.String()},
		Logger:           zap.NewNop(),
		DB:               db,
	}
	out, err := ClaimWorkItemAction(context.Background(), params)
	if err != nil {
		t.Fatalf("claim returned an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("claim did not run the shared registration predicate: %v", err)
	}

	result, _ := out.(map[string]interface{})
	if result["claimed"] != true {
		t.Errorf("a registered handler must claim; got %+v", result)
	}
}

// TestRoutablePredicateNamesBothFailureModes is the small, direct one: the guard
// has to cover BOTH ways claim refuses a row, and the empty-handler half is the
// one this bug was made of. A predicate that only checked registration would still
// promote every flag-only row, because `EXISTS (… type = ”)` is false — it would
// LOOK like it worked while blocking on the wrong reason.
func TestRoutablePredicateNamesBothFailureModes(t *testing.T) {
	p := workItemRoutableSQL("wi")

	if !strings.Contains(p, "COALESCE(wi.handler_agent, '') <> ''") {
		t.Errorf("missing the empty-handler half:\n%s", p)
	}
	if !strings.Contains(p, "ad.type = wi.handler_agent") {
		t.Errorf("the registration half must be keyed on the row's own handler:\n%s", p)
	}
	if !strings.Contains(p, "ad.deleted_at IS NULL") {
		t.Errorf("a deleted definition is not a handler:\n%s", p)
	}
	// Deliberately absent, and the comment on workItemHandlerRegisteredSQL says
	// why: claim does not filter on these, so neither may we. If this ever
	// changes it must change in claim first.
	if strings.Contains(p, "is_active") || strings.Contains(p, "is_snapshot") {
		t.Errorf("narrower than the claim path — this holds back rows claim would accept:\n%s", p)
	}
}
