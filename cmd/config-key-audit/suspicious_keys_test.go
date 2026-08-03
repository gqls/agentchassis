// FILE: cmd/config-key-audit/suspicious_keys_test.go
//
// bugs_open/134: a documentation convention leaked into two real config key
// names. Seed 156 wrote "{site_id, category?}" in a comment — correct there —
// and then wrote "category?" and "limit?" into the actual step config, where a
// key is just a key. `refresh_product_specs` reads "category" and "limit", so
// neither resolved, and because an unrecognised config key is silently ignored
// at execution nothing ever said so.
//
// The fixtures are the REAL shape first (the live product-spec-refresher step as
// bug 134 found it) and then the three other characters in the set, because a
// detector tested only on the case it was written for cannot tell you it still
// covers the rest. Both arms are here: what must fire, and the clean keys that
// must stay silent.
package main

import (
	"encoding/json"
	"testing"
)

// The live refresh_specs step exactly as bugs_open/134 found it, alongside the
// clean neighbour ("site_id") that must not be flagged.
const optionalMarkerFleet = `[
	{"type": "product-spec-refresher", "workflow": {"start_step": "ensure_site_record", "steps": {
		"ensure_site_record": {"action": "ensure_site_record", "next_step": "refresh_specs",
			"config": {"input_fields": ["site_id", "domain"]}},
		"refresh_specs": {"action": "refresh_product_specs", "next_step": "complete", "config": {
			"site_id":   "site_record.site_id",
			"category?": "input_data.category",
			"limit?":    "input_data.limit"
		}},
		"complete": {"action": "complete_workflow", "config": {"output_fields": ["refresh_result"]}}
	}}}
]`

func TestOptionalMarkerKeysAreFlagged(t *testing.T) {
	agents, failed, err := decodeLiveAgents([]byte(optionalMarkerFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}
	if failed != 0 {
		t.Fatalf("expected 0 undecodable rows, got %d", failed)
	}

	findings := findSuspiciousKeys(agents)
	if len(findings) != 2 {
		t.Fatalf("expected the two marker keys to be flagged, got %d: %+v", len(findings), findings)
	}

	// Sorted by (agent, path, key), so "category?" precedes "limit?".
	want := []suspiciousKeyFinding{
		{Agent: "product-spec-refresher", Path: "steps.refresh_specs", Action: "refresh_product_specs", Key: "category?", Nested: false},
		{Agent: "product-spec-refresher", Path: "steps.refresh_specs", Action: "refresh_product_specs", Key: "limit?", Nested: false},
	}
	for i, w := range want {
		if findings[i] != w {
			t.Errorf("findings[%d] = %+v, want %+v", i, findings[i], w)
		}
	}
}

// The other three characters in the set. "?" is the one that bit us; a check
// that only catches the recorded instance would be a fix, not a detector.
func TestTheOtherDocNotationCharactersAreFlagged(t *testing.T) {
	const mixedFleet = `[
		{"type": "glob-agent", "workflow": {"start_step": "s", "steps": {
			"s": {"action": "scrape_web", "config": {"page_*": "input_data.pages"}}
		}}},
		{"type": "spaced-agent", "workflow": {"start_step": "s", "steps": {
			"s": {"action": "scrape_web", "config": {"max pages": "3"}}
		}}},
		{"type": "colon-agent", "workflow": {"start_step": "s", "steps": {
			"s": {"action": "scrape_web", "config": {"site_id: uuid": "site_record.site_id"}}
		}}}
	]`

	agents, _, err := decodeLiveAgents([]byte(mixedFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	findings := findSuspiciousKeys(agents)
	if len(findings) != 3 {
		t.Fatalf("expected one finding per character, got %d: %+v", len(findings), findings)
	}
	wantKeys := map[string]string{
		"glob-agent":   "page_*",
		"spaced-agent": "max pages",
		"colon-agent":  "site_id: uuid",
	}
	for _, f := range findings {
		want, ok := wantKeys[f.Agent]
		if !ok {
			t.Errorf("unexpected agent in findings: %+v", f)
			continue
		}
		if f.Key != want {
			t.Errorf("%s: key = %q, want %q", f.Agent, f.Key, want)
		}
		delete(wantKeys, f.Agent)
	}
	for agent, key := range wantKeys {
		t.Errorf("missing finding: %s carrying %q", agent, key)
	}
}

// A key inside a loop sub-workflow is a real key, and this is the arm that
// asserts the traversal rather than the character set. bugs_open/144: the audit
// and the runtime validator both walked the top level only, went blind in the
// same direction and agreed with each other — 25 (action, key) pairs lived only
// inside loop bodies and were invisible to the whole system. Every mode here
// walks with validation.WalkSteps for that reason; if this one ever regresses to
// a top-level range, a marker key hidden in a loop body reads as a clean fleet.
func TestANestedKeyIsFlaggedAndMarkedNested(t *testing.T) {
	const nestedFleet = `[
		{"type": "loop-agent", "workflow": {"start_step": "per_product", "steps": {
			"per_product": {"action": "loop", "config": {"substeps": {
				"refresh": {"action": "refresh_product_specs", "next_step": "",
					"config": {"site_id": "site_record.site_id", "limit?": "input_data.limit"}}
			}}}
		}}}
	]`

	agents, _, err := decodeLiveAgents([]byte(nestedFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	findings := findSuspiciousKeys(agents)
	if len(findings) != 1 {
		t.Fatalf("expected the nested marker key to be found, got %d findings: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.Key != "limit?" {
		t.Errorf("key = %q, want limit?", f.Key)
	}
	if !f.Nested {
		t.Errorf("nested = false for %s — a finding inside a loop body must say so, or the reader "+
			"looks for it at the top level and concludes the report is wrong", f.Path)
	}
	if f.Path != "steps.per_product.substeps.refresh" {
		t.Errorf("path = %q, want steps.per_product.substeps.refresh (the path is what makes a finding "+
			"actionable without a second query)", f.Path)
	}
}

// The negative control. A check that fires on ordinary keys is worse than no
// check: the report gets ignored, and the next real marker key goes through it
// unread. Underscores and dots are the two shapes almost every live key has.
func TestOrdinaryKeysAreNotFlagged(t *testing.T) {
	const cleanFleet = `[
		{"type": "clean-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "refresh_product_specs", "next_step": "s2", "config": {
				"site_id":       "site_record.site_id",
				"category":      "input_data.category",
				"limit":         "input_data.limit",
				"a.b":           "collected.a.b",
				"max_pages":     "3",
				"input_fields":  ["site_id", "domain"],
				"upload-result": "true"
			}},
			"s2": {"action": "complete_workflow", "config": {"output_fields": ["refresh_result"]}}
		}}}
	]`

	agents, _, err := decodeLiveAgents([]byte(cleanFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	findings := findSuspiciousKeys(agents)
	if len(findings) != 0 {
		t.Fatalf("expected no findings over ordinary keys, got %+v", findings)
	}

	out, err := json.Marshal(findings)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(out) != "[]" {
		t.Errorf("zero findings must encode as '[]', not null, so a consumer can iterate without a "+
			"nil check and cannot mistake it for a missing key; got %q", out)
	}
}

// The order is part of the contract: scripts/audit-config-keys.sh prints these
// straight out, and a report that reorders itself between runs over an unchanged
// fleet cannot be diffed or held as a baseline. Map iteration over step.Config
// is randomised by the runtime, so without the sort this test fails
// intermittently rather than never — which is the point of asserting it.
func TestFindingsAreSortedDeterministically(t *testing.T) {
	const unsortedFleet = `[
		{"type": "zeta-agent", "workflow": {"start_step": "s", "steps": {
			"s": {"action": "scrape_web", "config": {"b?": "1", "a?": "2"}}
		}}},
		{"type": "alpha-agent", "workflow": {"start_step": "z_step", "steps": {
			"z_step": {"action": "scrape_web", "config": {"k?": "1"}},
			"a_step": {"action": "scrape_web", "config": {"k?": "1"}}
		}}}
	]`

	agents, _, err := decodeLiveAgents([]byte(unsortedFleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	want := []string{
		"alpha-agent\tsteps.a_step\tk?",
		"alpha-agent\tsteps.z_step\tk?",
		"zeta-agent\tsteps.s\ta?",
		"zeta-agent\tsteps.s\tb?",
	}

	// Repeated, because a single pass can agree with the wanted order by luck:
	// Go randomises map iteration per range, so an unsorted implementation
	// produces a different order on different runs of the SAME input.
	for attempt := 0; attempt < 20; attempt++ {
		findings := findSuspiciousKeys(agents)
		if len(findings) != len(want) {
			t.Fatalf("attempt %d: expected %d findings, got %d: %+v", attempt, len(want), len(findings), findings)
		}
		for i, w := range want {
			got := findings[i].Agent + "\t" + findings[i].Path + "\t" + findings[i].Key
			if got != w {
				t.Fatalf("attempt %d: findings[%d] = %q, want %q (sorted by agent, path, key)", attempt, i, got, w)
			}
		}
	}
}
