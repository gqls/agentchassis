package actions

import (
	"encoding/json"
	"go/format"
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

// TestFormatGeneratedGoCanonicalisesBodies pins bugs_open/013. The implementer
// committed the LLM's bytes verbatim, so a cosmetically-unformatted file failed
// the build gate's first step (`gofmt -l`) and spent the whole run — LLM
// generation, git push and a k8s build Job — for no PR. This is the exact shape
// that killed BUG A's first implementer run: a new struct field inserted without
// re-aligning its sibling.
func TestFormatGeneratedGoCanonicalisesBodies(t *testing.T) {
	unformatted := "package aiservice\n\ntype response struct {\n" +
		"StopReason string `json:\"stop_reason\"`\n" +
		"Usage struct{ InputTokens int } `json:\"usage\"`\n" +
		"}\n"

	files := map[string]interface{}{
		"platform/aiservice/anthropic.go": unformatted,
		"docs/notes.md":                   "# not go, must be left alone\n   ragged   ",
	}

	if err := formatGeneratedGo(files); err != nil {
		t.Fatalf("valid-but-unformatted Go must format, not error: %v", err)
	}

	got, _ := files["platform/aiservice/anthropic.go"].(string)
	want, err := format.Source([]byte(unformatted))
	if err != nil {
		t.Fatalf("fixture is not valid Go: %v", err)
	}
	if got != string(want) {
		t.Errorf("body not canonicalised.\n got: %q\nwant: %q", got, string(want))
	}
	if got == unformatted {
		t.Error("body unchanged — the gate would still reject it")
	}
	// Non-Go files must pass through untouched: the gate only gofmts .go, and
	// rewriting anything else would be an unrequested edit to a committed file.
	if files["docs/notes.md"] != "# not go, must be left alone\n   ragged   " {
		t.Errorf("non-Go file was modified: %q", files["docs/notes.md"])
	}
}

// TestFormatGeneratedGoAcceptsCommitEntryShape pins the B4 first-fire failure
// (2026-07-24): validateImplementation emits GitCommitData entries —
// {"content": <string>, "encoding": "utf-8"} — while formatGeneratedGo asserted
// a bare string, so the FIRST .go file through the implementer died at
// commit-prep with "non-string body (map[string]interface {})". The two
// functions were tested separately with different shapes and never run
// together; this test runs the real chain.
func TestFormatGeneratedGoAcceptsCommitEntryShape(t *testing.T) {
	unformatted := "package main\n\nfunc main() {\nprintln(\"x\")\n}\n"
	plan := planLite{Edits: []planEditLite{{File: "cmd/tools-api/main.go", Operation: "add"}}}
	impl := implWire{Files: []implFile{{Path: "cmd/tools-api/main.go", Content: unformatted}}}

	files, violations := validateImplementation(plan, impl, map[string]string{})
	if len(violations) != 0 {
		t.Fatalf("fixture should validate cleanly, got: %v", violations)
	}
	if err := formatGeneratedGo(files); err != nil {
		t.Fatalf("commit-entry-shaped .go body must format, not error: %v", err)
	}
	entry, ok := files["cmd/tools-api/main.go"].(map[string]interface{})
	if !ok {
		t.Fatalf("commit entry shape must be preserved, got %T", files["cmd/tools-api/main.go"])
	}
	got, _ := entry["content"].(string)
	want, err := format.Source([]byte(unformatted))
	if err != nil {
		t.Fatalf("fixture is not valid Go: %v", err)
	}
	if got != string(want) {
		t.Errorf("wrapped body not canonicalised.\n got: %q\nwant: %q", got, string(want))
	}
	if entry["encoding"] != "utf-8" {
		t.Errorf("encoding field must survive formatting, got %v", entry["encoding"])
	}
}

// A truncated body inside the commit-entry shape must still fail loud — the
// wrapped path must keep the bugs_open/013 posture, not silently bypass it.
func TestFormatGeneratedGoFailsLoudOnUnparseableWrappedBody(t *testing.T) {
	files := map[string]interface{}{
		"platform/aiservice/anthropic.go": map[string]interface{}{
			"content":  "package aiservice\n\nfunc GenerateText() {\n\tif x {\n",
			"encoding": "utf-8",
		},
	}
	err := formatGeneratedGo(files)
	if err == nil {
		t.Fatal("an unparseable wrapped body must fail commit-prep, not reach the gate")
	}
	for _, want := range []string{"platform/aiservice/anthropic.go", "truncated"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must name %q", err.Error(), want)
		}
	}
}

// A body that will not parse is a different failure and must stay loud — it is
// most often a max_tokens truncation. Silently committing the raw bytes would
// push known-broken Go and defer the error to the build Job.
func TestFormatGeneratedGoFailsLoudOnUnparseableBody(t *testing.T) {
	files := map[string]interface{}{
		// Truncated mid-function, the classic max_tokens shape.
		"platform/aiservice/anthropic.go": "package aiservice\n\nfunc GenerateText() {\n\tif x {\n",
	}

	err := formatGeneratedGo(files)
	if err == nil {
		t.Fatal("an unparseable body must fail commit-prep, not reach the gate")
	}
	for _, want := range []string{"platform/aiservice/anthropic.go", "truncated"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q must name %q", err.Error(), want)
		}
	}
}
