package actions

import (
	"encoding/json"
	"strings"
	"testing"
)

// F1.1b(c): the allowlist is the safety boundary between the LLM's file output
// and the write surface. These tests call validateImplementation — the REAL
// core the action runs — not a mirror of it.

func mkPlan(t *testing.T) planLite {
	t.Helper()
	raw := `{
		"summary": "Fix the sectionless-page silent success. Second sentence ignored in titles.",
		"edits": [
			{"file": "platform/orchestration/actions/apply_gap_plan_action.go", "symbol": "defaultSectionsForPage", "operation": "modify", "rationale": "r1"},
			{"file": "platform/orchestration/actions/populate_nav_tables_action.go", "symbol": "loadPagesForNav", "operation": "modify", "rationale": "r2"},
			{"file": "page-build-handler.json", "symbol": "complete_error", "operation": "config_change", "rationale": "r3"}
		],
		"grounded_in": ["quote"],
		"risks": "risks text"
	}`
	var plan planLite
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		t.Fatalf("test plan invalid: %v", err)
	}
	return plan
}

func TestValidateImplementationHappyPath(t *testing.T) {
	impl := implWire{Files: []implFile{
		{Path: "platform/orchestration/actions/apply_gap_plan_action.go", Content: "package actions // new"},
		{Path: "platform/orchestration/actions/populate_nav_tables_action.go", Content: "package actions // new2"},
	}}
	files, violations := validateImplementation(mkPlan(t), impl, nil)
	if len(violations) != 0 {
		t.Fatalf("want pass, got: %v", violations)
	}
	if len(files) != 2 {
		t.Fatalf("want 2 files, got %d", len(files))
	}
	// GitCommitData shape: {content, encoding}
	entry, ok := files["platform/orchestration/actions/apply_gap_plan_action.go"].(map[string]interface{})
	if !ok || entry["encoding"] != "utf-8" || entry["content"] == "" {
		t.Fatalf("files map not GitCommitData-shaped: %#v", entry)
	}
}

func TestValidateImplementationRejectsOutsideAllowlist(t *testing.T) {
	impl := implWire{Files: []implFile{
		{Path: "platform/orchestration/actions/apply_gap_plan_action.go", Content: "x"},
		{Path: "platform/orchestration/actions/populate_nav_tables_action.go", Content: "y"},
		{Path: "platform/messaging/processor.go", Content: "sneaky"},
	}}
	_, violations := validateImplementation(mkPlan(t), impl, nil)
	if !hasViolation(violations, "processor.go is OUTSIDE") {
		t.Fatalf("want outside-allowlist rejection, got: %v", violations)
	}
}

func TestValidateImplementationRejectsConfigChangeAsFile(t *testing.T) {
	// config_change targets agent_definitions, not the repo — a fabricated
	// file for it must land outside the allowlist.
	impl := implWire{Files: []implFile{
		{Path: "platform/orchestration/actions/apply_gap_plan_action.go", Content: "x"},
		{Path: "platform/orchestration/actions/populate_nav_tables_action.go", Content: "y"},
		{Path: "page-build-handler.json", Content: "{}"},
	}}
	_, violations := validateImplementation(mkPlan(t), impl, nil)
	if !hasViolation(violations, "page-build-handler.json is OUTSIDE") {
		t.Fatalf("want config_change-as-file rejection, got: %v", violations)
	}
}

func TestValidateImplementationRejectsIncomplete(t *testing.T) {
	impl := implWire{Files: []implFile{
		{Path: "platform/orchestration/actions/apply_gap_plan_action.go", Content: "x"},
	}}
	_, violations := validateImplementation(mkPlan(t), impl, nil)
	if !hasViolation(violations, "populate_nav_tables_action.go was NOT implemented") {
		t.Fatalf("want incomplete-implementation rejection, got: %v", violations)
	}
}

func TestValidateImplementationRejectsNoOpAndEmptyAndDuplicate(t *testing.T) {
	impl := implWire{Files: []implFile{
		{Path: "platform/orchestration/actions/apply_gap_plan_action.go", Content: "same"},
		{Path: "platform/orchestration/actions/populate_nav_tables_action.go", Content: ""},
		{Path: "platform/orchestration/actions/apply_gap_plan_action.go", Content: "same"},
	}}
	originals := map[string]string{
		"platform/orchestration/actions/apply_gap_plan_action.go": "same",
	}
	_, violations := validateImplementation(mkPlan(t), impl, originals)
	if !hasViolation(violations, "no-op") {
		t.Fatalf("want no-op rejection, got: %v", violations)
	}
	if !hasViolation(violations, "empty body") {
		t.Fatalf("want empty-body rejection, got: %v", violations)
	}
}

func TestValidateImplementationConfigChangeOnlyPlan(t *testing.T) {
	plan := planLite{Edits: []planEditLite{
		{File: "page-build-handler.json", Operation: "config_change"},
	}}
	_, violations := validateImplementation(plan, implWire{}, nil)
	if !hasViolation(violations, "no modify/add edits") {
		t.Fatalf("want config_change-only rejection, got: %v", violations)
	}
}

func TestFirstSentence(t *testing.T) {
	if got := firstSentence("Fix the thing. And more."); got != "Fix the thing" {
		t.Fatalf("firstSentence: %q", got)
	}
	long := strings.Repeat("a", 150)
	if got := firstSentence(long); len(got) != 100 {
		t.Fatalf("want 100-char cap, got %d", len(got))
	}
}

func hasViolation(violations []string, substr string) bool {
	for _, v := range violations {
		if strings.Contains(v, substr) {
			return true
		}
	}
	return false
}
