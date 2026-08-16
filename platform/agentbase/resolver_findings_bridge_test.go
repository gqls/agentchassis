// FILE: platform/agentbase/resolver_findings_bridge_test.go
//
// RFC_029 Phase 1 revision: the chassis is the ONE place that turns a
// datahelpers.ResolverFinding into an agent_error_log row. The mapping is pure
// and pinned here without a DB — the INSERT itself is agenterrors' contract,
// tested in that package; the resolver's one-finding-per-occurrence contract
// is datahelpers'. What THIS test guards is the join between them: the row
// carries the code the observation-window query keys on, the severity that
// keeps it out of the error/fatal populations, and the pod identity that is
// the only attribution the resolver can give.
package agentbase

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestResolverFindingEntryCarriesCodeSeverityAndPodIdentity(t *testing.T) {
	a := &Agent{AgentType: "agent-chassis", AgentID: "agent-1", PodName: "agent-chassis-abc12"}
	f := datahelpers.ResolverFinding{
		Code:    datahelpers.ResolverFindingConflictingCandidates,
		Field:   "purpose",
		Message: "aggressive search: conflicting candidates",
		Context: map[string]interface{}{"field": "purpose", "winner_path": "alpha.purpose"},
	}

	e := a.resolverFindingEntry(f)

	if e.ErrorCode != "RESOLVER_CONFLICTING_CANDIDATES" {
		t.Errorf("ErrorCode = %q — the observation-window query keys on the literal code", e.ErrorCode)
	}
	if e.Severity != "warning" {
		t.Errorf("Severity = %q, want warning: Phase 1 observes, it does not fail", e.Severity)
	}
	if e.AgentType != "agent-chassis" || e.PodName != "agent-chassis-abc12" || e.AgentID != "agent-1" {
		t.Errorf("identity = %s/%s/%s — pod-level attribution is the only attribution these rows have", e.AgentType, e.AgentID, e.PodName)
	}
	if e.OrchestrationID != "" || e.StepName != "" {
		t.Errorf("orchestration_id/step_name = %q/%q, want empty: the resolver cannot know them, and inventing one would be a false join", e.OrchestrationID, e.StepName)
	}
	if e.Action != resolverFindingAction {
		t.Errorf("Action = %q, want %q", e.Action, resolverFindingAction)
	}
	if e.ErrorMessage != f.Message || e.Context["winner_path"] != "alpha.purpose" {
		t.Errorf("message/context not carried through: %q %v", e.ErrorMessage, e.Context)
	}
}

// With no pool the bridge is a no-op — it must not panic on a nil db, because
// SetResolverFindingRecorder is only reached after the pool exists, but a
// zero Agent in a test or a shutdown race must still be safe.
func TestRecordResolverFindingWithoutDBIsANoOp(t *testing.T) {
	a := &Agent{}
	a.recordResolverFinding(datahelpers.ResolverFinding{Code: datahelpers.ResolverFindingMappingBypassed})
}
