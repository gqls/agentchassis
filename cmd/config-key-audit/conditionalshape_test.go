// FILE: cmd/config-key-audit/conditionalshape_test.go
//
// bugs_open/313. findArrayProducerConditions is the class-closing half of the
// internal-linker fix: migration 488 repaired the one live instance, and this
// check is what makes the next producer/consumer shape mismatch loud at config
// time instead of four months into silent complete_no_candidates runs.
//
// The first fixture is the REAL pre-488 internal-linker shape, not an invented
// one — a detector whose only test is the state it was written to produce
// cannot tell you it still works, so the repaired shape is here as the
// negative arm.
package main

import (
	"testing"
)

// The exact pre-488 internal-linker pairing: array producer, .count consumer.
// check_target_found is the in-workflow control — same consumer idiom against
// an OBJECT producer, and it must NOT be flagged.
const preFix313Fleet = `[
	{"type": "internal-linker", "workflow": {"start_step": "ensure_site_record", "steps": {
		"load_target_page":     {"action": "query_database", "output_field": "target_page",
		                         "config": {"output_format": "object", "query": "SELECT 1"}},
		"check_target_found":   {"action": "conditional",
		                         "config": {"condition": "target_page.page_id != null",
		                                    "then_step": "load_candidate_pages", "else_step": "complete_not_found"}},
		"load_candidate_pages": {"action": "query_database", "output_field": "candidate_pages",
		                         "config": {"output_format": "array", "query": "SELECT 1"}},
		"check_candidates":     {"action": "conditional",
		                         "config": {"condition": "candidate_pages.count > 0",
		                                    "then_step": "load_specs", "else_step": "complete_no_candidates"}}
	}}}
]`

func TestArrayProducerConditionFlagsThe313Shape(t *testing.T) {
	agents, failed, err := decodeLiveAgents([]byte(preFix313Fleet), "test")
	if err != nil || failed != 0 {
		t.Fatalf("decodeLiveAgents: err=%v failed=%d", err, failed)
	}
	findings, conditionals := findArrayProducerConditions(agents)
	if conditionals != 2 {
		t.Fatalf("expected 2 conditional steps counted, got %d", conditionals)
	}
	if len(findings) != 1 {
		t.Fatalf("expected exactly 1 finding (the object-producer probe must NOT flag), got %d: %+v",
			len(findings), findings)
	}
	f := findings[0]
	if f.ConditionPath != "candidate_pages.count" {
		t.Errorf("condition_path = %q, want candidate_pages.count", f.ConditionPath)
	}
	if f.ProducerPath == "" || f.OutputFormat != "array" {
		t.Errorf("finding must name the producer and its declared format, got %+v", f)
	}
}

// The post-488 shape: the same workflow with output_format object. Zero
// findings — the fix is what this detector certifies from now on.
const postFix313Fleet = `[
	{"type": "internal-linker", "workflow": {"start_step": "ensure_site_record", "steps": {
		"load_candidate_pages": {"action": "query_database", "output_field": "candidate_pages",
		                         "config": {"output_format": "object", "query": "SELECT 1"}},
		"check_candidates":     {"action": "conditional",
		                         "config": {"condition": "candidate_pages.count > 0",
		                                    "then_step": "load_specs", "else_step": "complete_no_candidates"}}
	}}}
]`

func TestArrayProducerConditionPassesTheRepairedShape(t *testing.T) {
	agents, _, err := decodeLiveAgents([]byte(postFix313Fleet), "test")
	if err != nil {
		t.Fatalf("decodeLiveAgents: %v", err)
	}
	findings, conditionals := findArrayProducerConditions(agents)
	if conditionals != 1 {
		t.Fatalf("expected 1 conditional counted, got %d", conditionals)
	}
	if len(findings) != 0 {
		t.Fatalf("repaired shape must produce 0 findings, got %+v", findings)
	}
}

// An ABSENT output_format defaults to array (database_actions.go) and must be
// flagged exactly like a declared one — reported with the raw empty format so
// the reader can tell defaulted from declared.
func TestArrayProducerConditionCatchesDefaultedFormat(t *testing.T) {
	fleet := `[
		{"type": "a", "workflow": {"steps": {
			"load": {"action": "query_database", "output_field": "res", "config": {"query": "SELECT 1"}},
			"chk":  {"action": "conditional", "config": {"condition": "res.count > 0", "then_step": "x", "else_step": "y"}}
		}}}
	]`
	agents, _, _ := decodeLiveAgents([]byte(fleet), "test")
	findings, _ := findArrayProducerConditions(agents)
	if len(findings) != 1 {
		t.Fatalf("defaulted output_format must be treated as array, got %+v", findings)
	}
	if findings[0].OutputFormat != "" {
		t.Errorf("raw format must stay empty for a defaulted array, got %q", findings[0].OutputFormat)
	}
}

// A conditional nested in a loop's sub_workflow must be walked (the
// bugs_open/144 32%-undercount class), and its read of a TOP-LEVEL producer's
// field must be judged against that producer.
func TestArrayProducerConditionDescendsIntoSubWorkflows(t *testing.T) {
	fleet := `[
		{"type": "a", "workflow": {"steps": {
			"load": {"action": "query_database", "output_field": "res",
			         "config": {"output_format": "array", "query": "SELECT 1"}},
			"lp":   {"action": "loop", "config": {"items_field": "res", "sub_workflow": {"steps": {
				"chk": {"action": "conditional", "config": {"condition": "res.count > 0", "then_step": "x", "else_step": "y"}}
			}}}}
		}}}
	]`
	agents, _, _ := decodeLiveAgents([]byte(fleet), "test")
	findings, conditionals := findArrayProducerConditions(agents)
	if conditionals != 1 {
		t.Fatalf("the nested conditional was not walked (counted %d)", conditionals)
	}
	if len(findings) != 1 {
		t.Fatalf("nested conditional against a top-level array producer must be flagged, got %+v", findings)
	}
}

// The three shapes the mode must deliberately SKIP: a numeric index into the
// array (WFA-012 — resolvable), a non-query_database producer (its emitted
// keys are invisible to config), and a field with a second, object-format
// producer (some run of the workflow can satisfy the path).
func TestArrayProducerConditionSkipsTheSatisfiableShapes(t *testing.T) {
	fleet := `[
		{"type": "numeric-index", "workflow": {"steps": {
			"load": {"action": "query_database", "output_field": "res",
			         "config": {"output_format": "array", "query": "SELECT 1"}},
			"chk":  {"action": "conditional", "config": {"condition": "res.0.url != null", "then_step": "x", "else_step": "y"}}
		}}},
		{"type": "go-action-producer", "workflow": {"steps": {
			"load": {"action": "load_unswept_areas", "output_field": "res", "config": {}},
			"chk":  {"action": "conditional", "config": {"condition": "res.count > 0", "then_step": "x", "else_step": "y"}}
		}}},
		{"type": "mixed-producers", "workflow": {"steps": {
			"load_a": {"action": "query_database", "output_field": "res",
			           "config": {"output_format": "array", "query": "SELECT 1"}},
			"load_b": {"action": "query_database", "output_field": "res",
			           "config": {"output_format": "object", "query": "SELECT 1"}},
			"chk":    {"action": "conditional", "config": {"condition": "res.count > 0", "then_step": "x", "else_step": "y"}}
		}}}
	]`
	agents, _, _ := decodeLiveAgents([]byte(fleet), "test")
	findings, conditionals := findArrayProducerConditions(agents)
	if conditionals != 3 {
		t.Fatalf("expected 3 conditionals counted, got %d", conditionals)
	}
	if len(findings) != 0 {
		t.Fatalf("all three shapes are satisfiable and must not be flagged, got %+v", findings)
	}
}

// AND/OR clauses are parsed the way the runtime parses them: the dead path
// inside a compound condition is still a finding, and the healthy clause
// beside it is not.
func TestArrayProducerConditionParsesCompoundConditions(t *testing.T) {
	fleet := `[
		{"type": "a", "workflow": {"steps": {
			"load_obj": {"action": "query_database", "output_field": "target",
			             "config": {"output_format": "object", "query": "SELECT 1"}},
			"load_arr": {"action": "query_database", "output_field": "res",
			             "config": {"output_format": "array", "query": "SELECT 1"}},
			"chk": {"action": "conditional",
			        "config": {"condition": "target.id != null AND res.count > 0", "then_step": "x", "else_step": "y"}}
		}}}
	]`
	agents, _, _ := decodeLiveAgents([]byte(fleet), "test")
	findings, _ := findArrayProducerConditions(agents)
	if len(findings) != 1 {
		t.Fatalf("expected exactly the res.count clause flagged, got %+v", findings)
	}
	if findings[0].ConditionPath != "res.count" {
		t.Errorf("condition_path = %q, want res.count", findings[0].ConditionPath)
	}
}

// conditionLeftPaths mirrors the runtime's probe order — "  >= " must not be
// split at " > ", and a bare truthy path is returned whole.
func TestConditionLeftPaths(t *testing.T) {
	got := conditionLeftPaths("a.count >= 3 OR (b.flag == true AND c.items)")
	want := []string{"a.count", "b.flag", "c.items"}
	if len(got) != len(want) {
		t.Fatalf("paths = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("paths[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
