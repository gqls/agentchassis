package actions

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// Triage is a deterministic router; its dedup key, symptom, and report are pure
// and safety-relevant (a wrong key would flood the loop with duplicates). Test
// the real functions.

func TestTriageItemKeyStableAndPatternScoped(t *testing.T) {
	k1 := triageItemKey("needs_content_page", "page-build-handler", "boom at step X")
	k2 := triageItemKey("needs_content_page", "page-build-handler", "boom at step X")
	if k1 != k2 {
		t.Fatalf("same pattern must yield same key: %s vs %s", k1, k2)
	}
	if triageItemKey("needs_content_page", "page-build-handler", "DIFFERENT error") == k1 {
		t.Fatal("different error signature must yield a different key")
	}
	if triageItemKey("other_type", "page-build-handler", "boom at step X") == k1 {
		t.Fatal("different item_type must yield a different key")
	}
	if !strings.HasPrefix(k1, "triage-diag:needs_content_page:") {
		t.Fatalf("key not readable/prefixed: %s", k1)
	}
}

func TestTriageSpecJSONShape(t *testing.T) {
	s := triageSpecJSON("sym", "gqls", "agentchassis", "main", "diagnose-orchestrator")
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		t.Fatalf("spec not valid JSON: %v", err)
	}
	for _, k := range []string{"symptom", "owner", "repo", "ref", "correlation_id", "source"} {
		if _, ok := m[k]; !ok {
			t.Fatalf("spec missing key %q: %s", k, s)
		}
	}
	if m["correlation_id"] == "" {
		t.Fatal("correlation_id must be populated")
	}
}

func TestRenderTriageEmptyIsExplicit(t *testing.T) {
	out := renderTriage(336, time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC), nil, nil, 0, 0, 0, 3, false)
	for _, want := range []string{
		"No failed work items in this window.",
		"No capability_gap / deferred items in this window.",
		"no model",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("triage report missing %q:\n%s", want, out)
		}
	}
}

func TestRenderTriageReportsCapAndGaps(t *testing.T) {
	patterns := []failurePattern{
		{ItemType: "needs_content_page", Handler: "page-build-handler", ErrSig: "kafka write error", Count: 12, Sites: 4},
		{ItemType: "page_rerender", Handler: "page-rerender", ErrSig: "timeout", Count: 3, Sites: 1},
	}
	gaps := []capabilityGap{{Builder: "tool-builder", Count: 5, Sites: 2}}
	out := renderTriage(336, time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC), patterns, gaps, 1, 0, 1, 1, true)

	for _, want := range []string{
		"needs_content_page` via `page-build-handler` — 12 item(s), 4 site(s): kafka write error",
		"1 pattern(s) NOT escalated this sweep (cap=1)", // capped, named — never silently dropped
		"tool-builder** needed — 5 page(s) across 2 site(s)",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("triage report missing %q:\n%s", want, out)
		}
	}
}

func TestTriageSymptomHandlesEmptyError(t *testing.T) {
	s := triageSymptom(failurePattern{ItemType: "x", Handler: "h", ErrSig: "  ", Count: 1, Sites: 1})
	if !strings.Contains(s, "(no error text recorded)") {
		t.Fatalf("empty error not handled: %s", s)
	}
}
