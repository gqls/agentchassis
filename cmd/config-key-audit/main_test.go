// FILE: cmd/config-key-audit/main_test.go
//
// bugs_open/148: findUnregisteredActions replays the exact "unregistered action
// with no topic" rule platform/validation/workflow.go and subworkflow.go enforce
// at runtime, but offline. These fixtures are the same shapes bug 148 measured
// live: a local action (registered, no topic needed), a remote action correctly
// routed by topic, and the three concrete unregistered-and-untopic'd cases —
// one of them nested inside a loop, since that is exactly the traversal
// bugs_open/144 showed a hand-written walk could miss.
package main

import (
	"encoding/json"
	"testing"
)

func TestFindUnregisteredActions(t *testing.T) {
	const input = `[
		{"type": "clean-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "complete_workflow", "next_step": "s2"},
			"s2": {"action": "generate_text", "topic": "system.llm.requests"}
		}}},
		{"type": "html-assembler", "workflow": {"start_step": "assemble_html", "steps": {
			"assemble_html": {"action": "assemble_full_page", "next_step": "complete"}
		}}},
		{"type": "loop-agent", "workflow": {"start_step": "build_pages", "steps": {
			"build_pages": {"action": "loop", "config": {"substeps": {
				"wrap": {"action": "wrap_multipage", "next_step": ""}
			}}}
		}}}
	]`

	agents, failed, err := decodeLiveAgents([]byte(input), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}
	if failed != 0 {
		t.Fatalf("expected 0 undecodable rows, got %d", failed)
	}
	if len(agents) != 3 {
		t.Fatalf("expected 3 decoded agents, got %d", len(agents))
	}

	findings := findUnregisteredActions(agents)

	want := map[string]bool{
		"html-assembler\tsteps.assemble_html\tassemble_full_page\tfalse":    true,
		"loop-agent\tsteps.build_pages.substeps.wrap\twrap_multipage\ttrue": true,
	}
	if len(findings) != len(want) {
		t.Fatalf("expected %d findings, got %d: %+v", len(want), len(findings), findings)
	}
	for _, f := range findings {
		key := f.Agent + "\t" + f.Path + "\t" + f.Action + "\t" + boolString(f.Nested)
		if !want[key] {
			t.Errorf("unexpected finding: %+v", f)
		}
		delete(want, key)
	}
	for k := range want {
		t.Errorf("missing expected finding: %s", k)
	}

	// "loop" itself is a registered local action with no topic and must NOT be
	// flagged — only its unregistered nested substep should be.
	for _, f := range findings {
		if f.Action == "loop" {
			t.Errorf("'loop' (a registered local action) was flagged: %+v", f)
		}
	}
}

func TestFindUnregisteredActionsEncodesEmptyAsArray(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(`[
		{"type": "clean-agent", "workflow": {"start_step": "s1", "steps": {
			"s1": {"action": "complete_workflow"}
		}}}
	]`), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}

	findings := findUnregisteredActions(agents)
	if len(findings) != 0 {
		t.Fatalf("expected no findings, got %+v", findings)
	}

	out, err := json.Marshal(findings)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if string(out) != "[]" {
		t.Errorf("expected zero findings to encode as '[]' (not null, so a consumer can iterate without a nil check), got %q", out)
	}
}

func boolString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
