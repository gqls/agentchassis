package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// TestResumedStepKeepsTheRunAgentTypeThroughTheRealResponseRoundTrip is the
// INDUCED version of RFC_019 §7's acceptance test, and it exists because the
// live version cannot be run.
//
// §12 established that the production measurement is unfalsifiable: the three
// actions that used to file `generic` rows went dormant on 2026-08-05, four days
// before the fix rolled, so a post-roll count of zero is guaranteed by absent
// demand rather than by working code. Worse, the discriminating case — a step
// resumed after an await whose response carries a `generic` dispatch sender —
// has not appeared anywhere in `agent_error_log` since **2026-07-26**
// (`diagnose_persist_fix_plan`, an inheriting door: 17 rows `generic` up to
// 07-26, then `feature-designer` and `council-gate` thereafter). Waiting for one
// is therefore not a plan, so the condition is induced here instead.
//
// What makes this different from the unit test beside it: that one hands
// ensureFullExecutionContext a struct built by hand, which assumes the very
// thing in question — that a resumed context arrives without RunAgentType. This
// one BUILDS the context the way a resume actually builds it, by round-tripping
// through ToResponseHeaders/FromResponseHeaders, so the premise is exercised
// rather than asserted.
func TestResumedStepKeepsTheRunAgentTypeThroughTheRealResponseRoundTrip(t *testing.T) {
	const realAgent = "council-gate"

	// The request side, as the processor leaves it: RunAgentType resolved from
	// the loaded agent definition (messaging/processor.go:1828).
	outbound := &types.ExecutionContext{
		CorrelationID:   "corr-1",
		OrchestrationID: "orch-1",
		RunAgentType:    realAgent,
		Sender:          types.AgentIdentity{AgentType: realAgent, AgentID: "agent-1"},
	}

	// The await: the responder answers, and the coordinator rebuilds the context
	// from the RESPONSE. This is the real path, not a hand-built struct.
	resumed := types.FromResponseHeaders(outbound.ToResponseHeaders())
	if resumed == nil {
		t.Fatal("FromResponseHeaders returned nil; the round trip under test does not exist")
	}

	// THE PREMISE, exercised rather than assumed. ResponseHeaders carries no
	// run_agent_type field at all, so this is structural: no amount of care on
	// the sending side can preserve rung 1 across an await. If a future change
	// adds the field to the response, this assertion fails and tells its author
	// that the backfill below became redundant — which is the useful failure.
	if resumed.RunAgentType != "" {
		t.Fatalf("premise broken: a context rebuilt from a response now carries RunAgentType=%q. "+
			"The §7 backfill may be redundant — re-read RFC_019 §7 before deleting it", resumed.RunAgentType)
	}

	// The discriminating condition, induced. A generic-dispatch response names
	// `generic` as its sender, which is why the pre-fix ladder answered
	// `generic`: rung 1 was empty and rung 2 was the dispatch sender. This is
	// the case that has not occurred in production since 2026-07-26 — set
	// explicitly, because ensureFullExecutionContext only backfills Sender when
	// it is EMPTY, so a populated `generic` sender survives untouched and the
	// old code had nothing better to reach for.
	resumed.Sender.AgentType = "generic"

	state := &OrchestrationState{
		OrchestrationID: "orch-1",
		CorrelationID:   "corr-1",
		OwnerAgentType:  realAgent, // determineOwnerAgentType's durable answer from run start
		OwnerAgentID:    "agent-1",
	}

	ensureFullExecutionContext(resumed, state, "test-pod", zap.NewNop())

	// The claim: the row an inheriting RSH-008 door writes on this resumed step
	// names the real agent. runningStepProvenance asks exactly this method.
	if got := resumed.ResolvedAgentType(); got != realAgent {
		t.Errorf("ResolvedAgentType() = %q, want %q\n"+
			"This is the RFC_019 §7 defect reproduced: a step resumed after an await filed under the "+
			"dispatch sender instead of the agent whose workflow was running.", got, realAgent)
	}

	// The sender is deliberately NOT rewritten. The two fields answer different
	// questions — who dispatched, and whose workflow is executing — and a fix
	// that quietly made them agree would destroy the distinction rather than
	// resolve it.
	if resumed.Sender.AgentType != "generic" {
		t.Errorf("Sender.AgentType = %q, want %q: the backfill must not overwrite the dispatch sender",
			resumed.Sender.AgentType, "generic")
	}
}
