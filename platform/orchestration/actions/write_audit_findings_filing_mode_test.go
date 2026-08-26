// FILE: platform/orchestration/actions/write_audit_findings_filing_mode_test.go
//
// Pins the filing_mode=record seam (RFC_056). Two properties matter and both
// are tested by MUTATION-shaped assertions rather than by bookkeeping: a record
// row must be one that NEITHER promoter can dispatch (empty handler AND a parked
// status — either alone is not enough, see the promoter's doors), and the
// routing must survive in spec so the row can be released later.

package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestParseFilingMode_AbsentOrDispatchIsHistoricalBehaviour(t *testing.T) {
	for name, cfg := range map[string]map[string]interface{}{
		"absent":   {},
		"nil":      {"filing_mode": nil},
		"empty":    {"filing_mode": ""},
		"dispatch": {"filing_mode": "dispatch"},
		"spaced":   {"filing_mode": "  Dispatch "},
	} {
		got, err := parseFilingMode(cfg)
		if err != nil || got != filingModeDispatch {
			t.Errorf("%s: got (%q, %v), want (dispatch, nil)", name, got, err)
		}
	}
}

func TestParseFilingMode_RecordAndRefusals(t *testing.T) {
	got, err := parseFilingMode(map[string]interface{}{"filing_mode": "record"})
	if err != nil || got != filingModeRecord {
		t.Fatalf("record: got (%q, %v)", got, err)
	}
	// A typo must be an ERROR, never a silent dispatch — the wrong guess is a
	// page rewrite, which is the thing the setting exists to stop.
	for name, cfg := range map[string]map[string]interface{}{
		"typo":   {"filing_mode": "recrod"},
		"bool":   {"filing_mode": true},
		"number": {"filing_mode": 1},
	} {
		if _, err := parseFilingMode(cfg); err == nil {
			t.Errorf("%s: expected an error, got nil", name)
		}
	}
}

func TestRecordOnlyFinding_IsUndispatchableByBothPromoters(t *testing.T) {
	pageID := uuid.New()
	in := classifiedFinding{
		ItemType:     "content_rewrite",
		HandlerAgent: "page-build-handler",
		Severity:     "medium",
		Priority:     50,
		PageID:       &pageID,
		PageName:     "about",
		Spec:         map[string]interface{}{"category": "content", "page_name": "about"},
		DedupKey:     "offer-analysis_content_rewrite_about_site",
		Summary:      "the about page is mostly about us",
	}
	out := recordOnlyFinding(in, "offer-analysis")

	// The two doors. detected-item-promoter: COALESCE(handler_agent,'') <> ''.
	// triage_detected_items: status = 'detected' AND routable. A row must fail
	// BOTH, and each assertion is the one that would catch a half-fix.
	if out.HandlerAgent != "" {
		t.Errorf("handler_agent must be empty (promoter door), got %q", out.HandlerAgent)
	}
	if out.Status != "deferred" {
		t.Errorf("status must be 'deferred' (triage door keys on 'detected'), got %q", out.Status)
	}
	// Routing preserved, so the row can be released.
	if out.Spec["routed_handler"] != "page-build-handler" || out.Spec["routed_status"] != "detected" {
		t.Errorf("routing not preserved in spec: %v", out.Spec)
	}
	if out.Spec["filing_mode"] != "record" || out.Spec["deferred_by"] != "offer-analysis" {
		t.Errorf("provenance stamps missing (bugs_open/396): %v", out.Spec)
	}
	if _, ok := out.Spec["release_recipe"]; !ok {
		t.Errorf("release_recipe missing — a park with no release is bugs_open/396's shape")
	}
	// Identity unchanged: same type, same dedup key, same page.
	if out.ItemType != in.ItemType || out.DedupKey != in.DedupKey || out.PageID != in.PageID {
		t.Errorf("identity changed: %+v", out)
	}
	if !strings.HasPrefix(out.Summary, "[verdict, not dispatched] ") {
		t.Errorf("summary should read as a verdict, got %q", out.Summary)
	}
	// The anti-churn skip (round 04a3ce1f, debug_historian HIGH): a verdict
	// re-observed per audit is expected recurrence; without this flag two
	// retractions inside seven days brand the third re-file `unresolved`
	// (bugs_open/033's shape) and the self-correction licence is inexact.
	if !out.RecurrenceExpected {
		t.Errorf("record rows must set RecurrenceExpected — the two-strike arm otherwise brands the re-file unresolved")
	}
	// The input is not mutated (spec is copied, not aliased).
	if _, leaked := in.Spec["routed_handler"]; leaked {
		t.Errorf("input spec was mutated")
	}
	if in.HandlerAgent != "page-build-handler" {
		t.Errorf("input handler was mutated")
	}
}

func TestRecordOnlyFinding_LeavesAlreadyParkedRowsAlone(t *testing.T) {
	// A capability_gap fallback is already 'deferred' + '' with its own
	// provenance; re-parking it would overwrite that provenance with ours.
	in := classifiedFinding{
		ItemType:     "capability_gap",
		HandlerAgent: "",
		Status:       "deferred",
		Spec:         map[string]interface{}{"builder_needed": "x"},
		Summary:      "no route for audit category",
	}
	out := recordOnlyFinding(in, "site-review")
	if out.Status != "deferred" || out.HandlerAgent != "" {
		t.Fatalf("parked row changed: %+v", out)
	}
	if _, stamped := out.Spec["filing_mode"]; stamped {
		t.Errorf("an already-parked row must not be re-stamped: %v", out.Spec)
	}
	if out.RecurrenceExpected {
		t.Errorf("an already-parked row must not gain the recurrence flag")
	}
	if out.Summary != in.Summary {
		t.Errorf("summary of an already-parked row must be untouched, got %q", out.Summary)
	}
}
