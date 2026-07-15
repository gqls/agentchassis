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

func TestTriageRouteLoopWorthiness(t *testing.T) {
	cases := map[string]string{
		"":                                     "hold",
		"   ":                                  "hold",
		"Claim timed out (attempts exhausted)": "requeue",
		"Claim timed out — handler pod likely died":                        "requeue",
		"CONSUMER REBALANCE in progress":                                   "requeue",
		"step store_component failed: new row for relation violates check": "loop",
		"template rejected by pre-store validation":                        "loop",
		"kafka write error":                                                "loop",
	}
	for errSig, want := range cases {
		if got := triageRoute(errSig, defaultTransientSignatures); got != want {
			t.Fatalf("triageRoute(%q) = %q, want %q", errSig, got, want)
		}
	}
}

func TestRenderTriageEmptyIsExplicit(t *testing.T) {
	out := renderTriage(336, time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC), nil, nil, nil, nil, 0, 0, 0, 3, nil, false)
	for _, want := range []string{
		"No code-bug failure patterns in this window.",
		"Transient / infra → re-queue, NOT the loop (0 pattern(s))",
		"No parked escalation's pattern has resolved — all remain open.",
		"No capability_gap / deferred items in this window.",
		"no model",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("triage report missing %q:\n%s", want, out)
		}
	}
}

func TestRenderTriageGroupsCapAndGaps(t *testing.T) {
	loop := []failurePattern{
		{ItemType: "needs_content_page", Handler: "page-build-handler", ErrSig: "kafka write error", Count: 12, Sites: 4},
	}
	requeue := []failurePattern{
		{ItemType: "needs_page", Handler: "page-build-handler", ErrSig: "Claim timed out", Count: 6, Sites: 2},
	}
	hold := []failurePattern{
		{ItemType: "content_rewrite", Handler: "page-build-handler", ErrSig: "", Count: 1, Sites: 1},
	}
	gaps := []capabilityGap{{Builder: "tool-builder", Count: 5, Sites: 2}}
	out := renderTriage(336, time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC), loop, requeue, hold, gaps, 0, 0, 1, 1, nil, true)

	for _, want := range []string{
		"Code bugs → fix loop (1 pattern(s)",
		"needs_content_page` via `page-build-handler` — 12 item(s), 4 site(s): kafka write error",
		"1 pattern(s) NOT escalated this sweep (cap=1)",
		"Transient / infra → re-queue, NOT the loop (1 pattern(s))",
		"Claim timed out",
		"No error signal → hold for a human (1 pattern(s))",
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

// Close-out must only close what has genuinely resolved: an open escalation
// closes iff its pattern key no longer exists among live failure keys.
func TestTriageResolvedKeys(t *testing.T) {
	live := map[string]bool{"triage-diag:a:111111": true, "triage-diag:b:222222": true}
	open := []string{"triage-diag:a:111111", "triage-diag:gone:333333", "triage-diag:b:222222"}
	got := triageResolvedKeys(open, live)
	if len(got) != 1 || got[0] != "triage-diag:gone:333333" {
		t.Fatalf("want only the vanished pattern to close, got %v", got)
	}
	if r := triageResolvedKeys(nil, live); len(r) != 0 {
		t.Fatalf("no open escalations must mean no closures, got %v", r)
	}
	if r := triageResolvedKeys(open, map[string]bool{}); len(r) != 3 {
		t.Fatalf("all patterns gone must close all open escalations, got %v", r)
	}
}

func TestRenderTriageCloseOutSection(t *testing.T) {
	// Live sweep with a closure: the closed key must be named.
	out := renderTriage(336, time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC), nil, nil, nil, nil, 0, 0, 0, 3,
		[]string{"triage-diag:gone:333333"}, false)
	for _, want := range []string{
		"Close-out — escalations resolved (1)",
		"`triage-diag:gone:333333`",
		"re-escalates automatically if it returns",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("triage report missing %q:\n%s", want, out)
		}
	}
	// Dry-run must not render a close-out section at all — nothing was closed.
	dry := renderTriage(336, time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC), nil, nil, nil, nil, 0, 0, 0, 3, nil, true)
	if strings.Contains(dry, "Close-out") {
		t.Fatal("dry-run report must not show a close-out section")
	}
}
