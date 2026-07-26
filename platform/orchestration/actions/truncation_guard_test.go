package actions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
)

// The live shape on 2026-07-26: every step that opts into tolerance sits in a
// council workflow whose decider reads the marker. The guard must pass these
// unchanged — a fix for bugs_open/076 that breaks the councils the mechanism was
// built for is not a fix.
func TestCouncilWorkflowIsGuarded(t *testing.T) {
	steps := map[string]models.Step{
		"review_editquality": {Action: "execute_llm_prompt",
			Config: map[string]interface{}{"tolerate_truncation": true}},
		"review_guardian": {Action: "execute_llm_prompt",
			Config: map[string]interface{}{"tolerate_truncation": true}},
		"council_decide": {Action: "diagnose_council_decide"},
		"complete":       {Action: "complete_workflow"},
	}

	got, ok := findTruncationAwareConsumer(steps, "review_editquality")
	if !ok {
		t.Fatal("a council workflow containing diagnose_council_decide must be treated as guarded")
	}
	if got != "council_decide" {
		t.Fatalf("want the decider named as the guard, got %q", got)
	}
}

// The case the bug is about: tolerance opted into where nothing reads the marker.
// A page renderer consuming a fragment is the failure bugs_open/076 describes.
func TestUnguardedWorkflowIsRefused(t *testing.T) {
	steps := map[string]models.Step{
		"write_section": {Action: "execute_llm_prompt",
			Config: map[string]interface{}{"tolerate_truncation": true}},
		"render_page": {Action: "render_page_html"},
		"complete":    {Action: "complete_workflow"},
	}

	if name, ok := findTruncationAwareConsumer(steps, "write_section"); ok {
		t.Fatalf("no step here reads __truncated; want refusal, got guard %q", name)
	}
}

// A step must not certify its own truncation — otherwise accepts_truncated on the
// producing step would wave the fragment straight through to an unguarded reader.
func TestProducerCannotCertifyItself(t *testing.T) {
	steps := map[string]models.Step{
		"write_section": {Action: "execute_llm_prompt", Config: map[string]interface{}{
			"tolerate_truncation": true,
			"accepts_truncated":   true,
		}},
		"render_page": {Action: "render_page_html"},
	}

	if name, ok := findTruncationAwareConsumer(steps, "write_section"); ok {
		t.Fatalf("the producing step certified itself as its own guard (%q)", name)
	}

	// Both name forms are excluded: CurrentStep and ExecutionContext.StepName are
	// populated independently, and only one is guaranteed non-empty.
	if name, ok := findTruncationAwareConsumer(steps, "", "write_section"); ok {
		t.Fatalf("producer not excluded when named via the second form (%q)", name)
	}
}

// The config escape hatch, for a consumer whose action is too generic for the Go
// registry to speak for.
func TestConfigDeclarationCountsAsAGuard(t *testing.T) {
	steps := map[string]models.Step{
		"summarise": {Action: "execute_llm_prompt",
			Config: map[string]interface{}{"tolerate_truncation": true}},
		"store": {Action: "update_work_item_status",
			Config: map[string]interface{}{"accepts_truncated": true}},
	}

	got, ok := findTruncationAwareConsumer(steps, "summarise")
	if !ok || got != "store" {
		t.Fatalf("accepts_truncated must declare a guard; got %q ok=%v", got, ok)
	}
}

// An unknown plan fails closed. This is unreachable through the coordinator (it
// resolves the step from the plan before dispatching, so the map is populated),
// but the default has to be refusal: the whole point of bugs_open/076 is that a
// missing guard must not read as a present one.
func TestEmptyPlanFailsClosed(t *testing.T) {
	if name, ok := findTruncationAwareConsumer(nil, "write_section"); ok {
		t.Fatalf("an unknown plan must fail closed, got guard %q", name)
	}
	if name, ok := findTruncationAwareConsumer(map[string]models.Step{}, "x"); ok {
		t.Fatalf("an empty plan must fail closed, got guard %q", name)
	}
}

// Lockstep: truncationAwareActions is a claim about code, and a name added
// without a reader silently re-opens the bug for every workflow containing that
// action. Same shape as the provider/TruncatedError scan in
// platform/aiservice/stop_signal_test.go, and as the dedup-index-to-Go-list
// contract: two things that must agree, held together by a test rather than by
// remembering.
func TestTruncationAwareActionsReadTheMarker(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read package dir: %v", err)
	}

	// The registry's own file is EXCLUDED, and that exclusion is the test.
	// truncation_guard.go names "__truncated" in its doc comment, so scanning it
	// would let every entry satisfy the check by being declared — the check would
	// share ground with the thing it is meant to falsify and could never fail. A
	// falsification probe on 2026-07-26 proved exactly that: a bogus
	// "render_page_html" entry passed. The reader must live in the action's own
	// implementation file, not beside the claim.
	const registryFile = "truncation_guard.go"

	var sources []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		if name == registryFile {
			continue
		}
		sources = append(sources, name)
	}

	for action, mechanism := range truncationAwareActions {
		if strings.TrimSpace(mechanism) == "" {
			t.Errorf("action %q is registered truncation-aware with no mechanism recorded — the value is what makes the claim checkable", action)
		}

		var found bool
		for _, src := range sources {
			body, err := os.ReadFile(filepath.Clean(src))
			if err != nil {
				t.Fatalf("read %s: %v", src, err)
			}
			text := string(body)
			// The file that implements the action must also read the marker.
			if strings.Contains(text, action) && strings.Contains(text, "__truncated") {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("action %q is registered as truncation-aware, but no non-test file in this package both names it and reads \"__truncated\".\n"+
				"Either its guard was removed (bugs_open/076 is re-opened for every workflow using it) or the entry was added without one.", action)
		}
	}
}
