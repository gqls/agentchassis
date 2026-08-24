// FILE: platform/orchestration/actions/action_request_producers_recurrence_test.go
//
// bugs_open/326, option E (owner ruling 2026-08-24, "D + E now"): the Go producers
// that file an ACTION REQUEST — a pipeline stage handoff, a re-render after an
// asset landed, a build-queue seed — declare recurrenceExpected, so the anti-churn
// brake in writeWorkItem does not drop a legitimate repeat inside 3h or bury the
// third one in a terminal status.
//
// WHY A SOURCE RATCHET, and what it does and does not prove. The flag is one line
// in a struct literal at eight sites across seven files. writeWorkItem reads it
// directly, and nav_rebuild_request_test.go's TestNavRebuildRequestSkipsTheTwoStrikeRule
// already proves — mutation-proven, the awkward way — that the flag causes the
// effect. What this file adds is the other half of that chain: that each named
// site actually SETS it. Remove the line from any one site and this fails naming
// the site. The effect test at the bottom drives one real action end to end in
// the same worked pattern, so the chain is closed at least once at runtime.
//
// THE COMMENT-STRIPPING IS LOAD-BEARING. A source-scanning test makes comments
// part of the contract (the a-source-scanning-test-makes-comments-load-bearing
// trap): every site here carries a prose comment that SAYS "recurrenceExpected",
// so a naive scan passes on the comment alone with the code line deleted. Comments
// go first; only then is the struct literal read.
//
// SITES DELIBERATELY NOT HERE, so nobody "completes" the list:
//   - create_tool_cross_link_items.go — decided FALSE in its own comment ("a
//     cross-link is a detected gap, not a repeatable action request"). Respected.
//   - emit_content_card_derive.go — an action request, but its item_key is shared
//     with a discovery CHECK (ContentImageItemKey), so the decision belongs to the
//     lane that owns that coupling.
//   - rerender_page_sections_action.go:~1419 — a "full rebuild" ESCALATION after a
//     re-render could not proceed; whether a repeat is a legitimate re-request or
//     "the remedy did not work" is genuinely ambiguous. Left unflagged, listed.
//   - discovery_checks.go and every check that emits page_rerender — those are
//     DETECTORS filing a remedy, and when the remedy completes without fixing the
//     fault the brake is RIGHT to stop the churn. That population is bugs_open/352's.

package actions

import (
	"context"
	"database/sql/driver"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// actionRequestProducerSite names one struct literal by a line unique to it.
// The anchor is the item_key expression, because item_type alone is not unique
// (two sites file needs_page, two file needs_domain_research).
type actionRequestProducerSite struct {
	file   string
	anchor string
}

var actionRequestProducerSites = []actionRequestProducerSite{
	{"apply_adoption_plan_action.go", `itemKey:      "needs_domain_research",`},
	{"emit_design_items_action.go", `itemKey:      "needs_composition",`},
	{"emit_design_items_action.go", `itemKey:      "needs_design",`},
	{"emit_imagery_items_action.go", `itemType:     "needs_imagery",`},
	{"flag_page_image_rebuild_action.go", `summary:      fmt.Sprintf("Re-render %s after its image asset landed", pageName),`},
	{"reconcile_section_data_action.go", `itemKey:      fmt.Sprintf("page_rerender:%s", page),`},
	{"seed_build_queue_action.go", `itemKey:      fmt.Sprintf("seed_%s_%s", itemType, domain),`},
	{"validate_composition_inputs_action.go", `itemKey:      "backfill_classification_for_composition",`},
}

var lineComment = regexp.MustCompile(`//.*$`)

// structLiteralAfterAnchor returns the lines from the anchor to the struct
// literal's closing brace, comments stripped. The closing brace is the first
// subsequent line whose first non-tab character is '}' — struct-literal fields
// are one per line in this package, so that is the literal's end.
func structLiteralAfterAnchor(t *testing.T, src, anchor string) []string {
	t.Helper()
	lines := strings.Split(src, "\n")
	start := -1
	for i, l := range lines {
		if strings.Contains(l, anchor) {
			if start != -1 {
				t.Fatalf("anchor %q is not unique — pick a line that is", anchor)
			}
			start = i
		}
	}
	if start == -1 {
		t.Fatalf("anchor %q not found — the site moved or was rewritten; update the table rather than deleting the row", anchor)
	}
	var out []string
	for _, l := range lines[start:] {
		stripped := lineComment.ReplaceAllString(l, "")
		out = append(out, stripped)
		if strings.HasPrefix(strings.TrimLeft(stripped, "\t "), "}") {
			break
		}
	}
	return out
}

func TestActionRequestProducers_DeclareRecurrenceExpected(t *testing.T) {
	for _, site := range actionRequestProducerSites {
		t.Run(site.file+" :: "+site.anchor, func(t *testing.T) {
			raw, err := os.ReadFile(site.file)
			if err != nil {
				t.Fatalf("read %s: %v", site.file, err)
			}
			body := structLiteralAfterAnchor(t, string(raw), site.anchor)
			found := false
			for _, l := range body {
				if strings.Contains(l, "recurrenceExpected:") && strings.Contains(l, "true") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s: the struct literal at %q does not set recurrenceExpected: true — "+
					"this producer files an ACTION REQUEST, and without the flag the anti-churn "+
					"brake drops its second request inside 3h and buries its third (bugs_open/326)",
					site.file, site.anchor)
			}
		})
	}
}

// TestActionRequestProducers_TheStripperIsNotFooledByProse is the mutation guard
// for the ratchet itself: a comment line saying "recurrenceExpected: true" must
// NOT satisfy it. Without this, deleting the code line and leaving the comment
// passes — which is the trap the header names.
func TestActionRequestProducers_TheStripperIsNotFooledByProse(t *testing.T) {
	src := "\tx := workItem{\n" +
		"\t\titemKey: \"k\",\n" +
		"\t\t// recurrenceExpected: true — prose only, the field is NOT set\n" +
		"\t}\n"
	body := structLiteralAfterAnchor(t, src, `itemKey: "k",`)
	for _, l := range body {
		if strings.Contains(l, "recurrenceExpected:") {
			t.Fatalf("comment survived stripping: %q", l)
		}
	}
}

// ---------------------------------------------------------------------------
// The effect, closed once at runtime: EmitDesignItemsAction driven through
// sqlmock in nav_rebuild_request_test.go's worked pattern.
//
// Supply a two-strike history that WOULD brand both items 'unresolved', then
// require each INSERT to carry status = 'triaged' at $12. If either site loses
// the flag, the brake's COUNT runs, returns 2, the status becomes 'unresolved',
// the WithArgs mismatch fails the Exec, and the action returns an error.
// ExpectationsWereMet is deliberately NOT asserted — on the correct path the
// COUNT is never issued and that expectation is legitimately unused.
//
// expectWorkItemDoorStandsDown is bugs_open/333's owned-page door, consulted in
// writeWorkItem AFTER the brake for every dispatchable row; unscripted, the door
// fails open and the test silently stops exercising its own statement sequence.
// ---------------------------------------------------------------------------

func emitDesignInsertArgsRequiringStatus(status string) []driver.Value {
	const cols = 16
	args := make([]driver.Value, cols)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	args[11] = status // $12
	return args
}

func TestEmitDesignItems_StageHandoffsSurviveATwoStrikeHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	siteID := uuid.New()

	// No composition installed yet, so the action proceeds to emit.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT style_collection_id::text FROM sites")).
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"style_collection_id"}).AddRow(nil))

	// A history that WOULD trip the two-strike rule on either key: 2 prior
	// terminal items, newest 100h old (so within-cycle suppression does not
	// also apply). Registered so that if a flag is dropped the COUNT SUCCEEDS
	// and the branding actually happens, rather than erroring and being
	// swallowed by `if err == nil && terminalCount > 0`.
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*)")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age_hours"}).AddRow(2, 100.0))

	mock.ExpectBegin()
	// needs_composition → site-design-planner
	expectWorkItemDoorStandsDown(mock)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WithArgs(emitDesignInsertArgsRequiringStatus("triaged")...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// The composition-dependency lookup between the two inserts; an open
	// composition item to depend on.
	mock.ExpectQuery(regexp.QuoteMeta("FROM site_work_items")).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New().String()))
	// needs_design → webdesign-agent
	expectWorkItemDoorStandsDown(mock)
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO site_work_items")).
		WithArgs(emitDesignInsertArgsRequiringStatus("triaged")...).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	out, err := EmitDesignItemsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData:    map[string]interface{}{"site_id": siteID.String()},
	})
	if err != nil {
		t.Fatalf("EmitDesignItemsAction: %v — an INSERT did not carry status='triaged'. With a "+
			"2-strike history that means the brake branded a stage handoff 'unresolved', i.e. "+
			"recurrenceExpected is no longer set on it and a re-run of the design cascade would be "+
			"born terminal and never dispatched", err)
	}
	if m, ok := out.(map[string]interface{}); !ok || m["design_emitted"] != true {
		t.Fatalf("expected design_emitted=true, got %#v", out)
	}
}
