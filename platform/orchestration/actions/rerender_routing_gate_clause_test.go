// FILE: platform/orchestration/actions/rerender_routing_gate_clause_test.go
//
// bugs_open/440 / RFC_062 phase 3: the guard clause livespec renders is PASTED
// into a gate migration, where the workflow evaluator — not Go — decides what
// it means. This test lives HERE, in the package that owns that evaluator, and
// runs the rendered clause through it for every state an item can be in.
//
// WHY IT EXISTS. The first cut of CheckRoutingKnownConditionClause tested only
// `== ''` for "no routing key", on the assumption that a missing key and an
// empty one compare alike. MEASURED 2026-09-03 by executing compareValues:
// they do NOT — its nil branch runs BEFORE quote-stripping, so a quoted `''`
// never equals nil. The clause would have evaluated FALSE for every item minted
// before phase 2 and sent the entire legacy population to human review on the
// day the gate flipped. The blocker written into the module header, RFC_062 and
// two approved submissions is what made someone check; this test is so nobody
// has to check again.
//
// MUTATION CHECK: delete the `== null` disjunct from
// CheckRoutingKnownConditionClause and TestCheckRoutingKnownConditionClause_
// CoversEveryEvaluatorState must fail on the absent case.

package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/livespec"
	"go.uber.org/zap"
)

// evalGateClause runs a rendered OR-clause through the real evaluator the way
// the live gate does, against a spec built for one item state.
func evalGateClause(t *testing.T, clause string, spec map[string]interface{}) bool {
	t.Helper()
	data := map[string]interface{}{"input_data": map[string]interface{}{"spec": spec}}
	lg := zap.NewNop()
	for _, disjunct := range strings.Split(clause, " OR ") {
		idx := strings.Index(disjunct, " == ")
		if idx < 0 {
			t.Fatalf("clause disjunct is not an equality test: %q", disjunct)
		}
		field := strings.TrimSpace(disjunct[:idx])
		expected := strings.TrimSpace(disjunct[idx+4:])
		if compareValues(resolveFieldValue(field, data, lg), expected, lg) {
			return true
		}
	}
	return false
}

func TestCheckRoutingKnownConditionClause_CoversEveryEvaluatorState(t *testing.T) {
	clause := livespec.CheckRoutingKnownConditionClause()

	cases := []struct {
		name      string
		spec      map[string]interface{}
		wantAllow bool
		why       string
	}{
		{
			name:      "absent routing key (every item minted before phase 2)",
			spec:      map[string]interface{}{"page_name": "about"},
			wantAllow: true,
			why: "a missing key must ALLOW. This is the whole legacy population; refusing it " +
				"routes the fleet's re-renders to human review the day the gate flips. It needs " +
				"the `== null` disjunct — `== ''` does NOT match a missing key",
		},
		{
			name:      "present but empty",
			spec:      map[string]interface{}{"page_name": "about", "routing_reason": ""},
			wantAllow: true,
			why:       "an explicitly empty key is the assemble-only case and must ALLOW",
		},
		{
			name:      "present and in the vocabulary",
			spec:      map[string]interface{}{"page_name": "about", "routing_reason": livespec.ReasonCTALinksStale},
			wantAllow: true,
			why:       "a declared routing value must ALLOW and go on to route",
		},
		{
			name:      "present and NOT in the vocabulary",
			spec:      map[string]interface{}{"page_name": "about", "routing_reason": "tool_retirement"},
			wantAllow: false,
			why: "the ONLY refusing state, and the point of the whole bug: a routing key nobody " +
				"understands must not quietly assemble",
		},
	}

	for _, c := range cases {
		got := evalGateClause(t, clause, c.spec)
		if got != c.wantAllow {
			t.Errorf("%s: clause allowed=%v, want %v — %s\n  clause: %s",
				c.name, got, c.wantAllow, c.why, clause)
		}
	}
}

func TestTransitionClause_RoutesUnderEitherKeyDuringTheDrain(t *testing.T) {
	// The drain window: producers converted at different times, so the same
	// value must route whether it arrives under the new key or the old one.
	clause := livespec.TransitionRerenderModeConditionClause()
	for _, r := range livespec.RerenderSectionReasons {
		if !evalGateClause(t, clause, map[string]interface{}{"routing_reason": r.Name}) {
			t.Errorf("%q under routing_reason did not route — a converted producer's items would "+
				"assemble during the drain", r.Name)
		}
		if !evalGateClause(t, clause, map[string]interface{}{"reason": r.Name}) {
			t.Errorf("%q under reason did not route — an UNconverted producer's items would "+
				"assemble during the drain, which is the regression the transition clause exists "+
				"to prevent", r.Name)
		}
	}
	if evalGateClause(t, clause, map[string]interface{}{"reason": "tool_retirement"}) {
		t.Error("an out-of-vocabulary value routed to the sections branch under the transition clause")
	}
}
