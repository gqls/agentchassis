// FILE: platform/orchestration/actions/plan_sections_contract_test.go
//
// Tests for the query-list cardinality contract that plan_sections must honour
// (bugs_open/054 fix-candidate 2). The load-bearing property: a query-sourced
// list field that resolves EMPTY (or shorter than its declared min_items) fails
// its required/min_items contract and must be routed through on_missing instead
// of being stored as an empty slice the template ranges over to nothing. These
// are DB-free tests of the pure decision function; the routing itself (below
// contract -> handleMissingField) is a trivial branch verified by reading.
package actions

import "testing"

func TestQueryListBelowContract(t *testing.T) {
	empty := []map[string]interface{}{}
	one := []map[string]interface{}{{"name": "x"}}
	three := []map[string]interface{}{{}, {}, {}}

	cases := []struct {
		name     string
		value    interface{}
		required bool
		minItems int
		want     bool
	}{
		// The bug: a required list (the five list components carry
		// required:true, min_items:1) that resolves empty must fail the contract.
		{"required empty list fails", empty, true, 0, true},
		{"required empty list with min_items=1 fails", empty, true, 1, true},
		{"required one item satisfies the default floor", one, true, 0, false},
		{"required one item satisfies min_items=1", one, true, 1, false},

		// min_items is honoured on its own, independent of required.
		{"min_items=3 with one item fails", one, false, 3, true},
		{"min_items=3 with three items satisfies", three, false, 3, false},

		// An optional list with no floor is legitimately allowed to be empty
		// (news components: required:false, on_missing:skip_field).
		{"optional empty list with no min_items passes", empty, false, 0, false},
		{"optional empty list with min_items=1 fails", empty, false, 1, true},

		// Non-list values are never a cardinality failure — a scalar (e.g. a URL
		// from section_index_for) or a nil resolve keeps its prior handling, so
		// the fix cannot accidentally divert scalar fields into on_missing.
		{"scalar string never fails", "https://example/tools", true, 1, false},
		{"nil is not a list and never fails", nil, true, 1, false},

		// The defensive []interface{} shape is treated as a list too.
		{"[]interface{} empty required fails", []interface{}{}, true, 0, true},
		{"[]interface{} with one item satisfies", []interface{}{"x"}, true, 1, false},
	}

	for _, c := range cases {
		if got := queryListBelowContract(c.value, c.required, c.minItems); got != c.want {
			t.Errorf("%s: queryListBelowContract(%#v, required=%v, min=%d) = %v, want %v",
				c.name, c.value, c.required, c.minItems, got, c.want)
		}
	}
}

func TestQueryResultLen(t *testing.T) {
	if n, isList := queryResultLen([]map[string]interface{}{{}, {}}); !isList || n != 2 {
		t.Errorf("[]map len: got (%d, %v), want (2, true)", n, isList)
	}
	if n, isList := queryResultLen([]interface{}{1, 2, 3}); !isList || n != 3 {
		t.Errorf("[]interface len: got (%d, %v), want (3, true)", n, isList)
	}
	if _, isList := queryResultLen("a-scalar"); isList {
		t.Error("a scalar string must report isList=false")
	}
	if _, isList := queryResultLen(nil); isList {
		t.Error("nil must report isList=false")
	}
}

// hasItems (reconcile_section_data_action.go) was consolidated onto queryResultLen
// so the list-shape type-switch lives in one place (bugs_open/054, council R1).
// Pin its behaviour is preserved: non-empty list → true, everything else → false.
func TestHasItemsBehaviourPreserved(t *testing.T) {
	cases := []struct {
		value interface{}
		want  bool
	}{
		{[]map[string]interface{}{{"a": 1}}, true},
		{[]interface{}{"x"}, true},
		{[]map[string]interface{}{}, false},
		{[]interface{}{}, false},
		{"scalar", false},
		{nil, false},
	}
	for _, c := range cases {
		if got := hasItems(c.value); got != c.want {
			t.Errorf("hasItems(%#v) = %v, want %v", c.value, got, c.want)
		}
	}
}
