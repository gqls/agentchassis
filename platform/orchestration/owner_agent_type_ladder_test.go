// FILE: platform/orchestration/owner_agent_type_ladder_test.go
//
// The coordinator half of the hoisted ladder (2026-08-08).
//
// determineOwnerAgentType had no test at all, which is why this file exists.
// The change it guards moves the two CONTEXT rungs of its ladder onto
// types.ExecutionContext so the actions-package error-log door climbs the same
// ones, and leaves the two PROCESS rungs — os.Getenv("AGENT_TYPE") and the
// "generic" filler — here. A claim of the form "one ladder, two consumers" is
// only worth as much as its least-pinned consumer: with the actions side
// covered and this side bare, the drift could simply come back through here and
// every test would stay green.
package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func TestDetermineOwnerAgentType_Ladder(t *testing.T) {
	s := &SagaCoordinator{logger: zap.NewNop()}

	cases := []struct {
		name    string
		envVar  string
		execCtx *types.ExecutionContext
		want    string
		why     string
	}{
		{
			name: "the resolved run agent wins over the dispatch-path sender",
			execCtx: &types.ExecutionContext{
				RunAgentType: "council-gate",
				Sender:       types.AgentIdentity{AgentType: "generic"},
			},
			envVar: "the-pods-own-type",
			want:   "council-gate",
			why:    "bugs_open/060: on the generic dispatch path the sender is literally 'generic', and owner_agent_type is the column every investigation of a run starts from",
		},
		{
			name: "the sender is used when the run agent type is absent",
			execCtx: &types.ExecutionContext{
				Sender: types.AgentIdentity{AgentType: "page-build-handler"},
			},
			envVar: "the-pods-own-type",
			want:   "page-build-handler",
			why:    "a dedicated pod, and any message predating the run_agent_type header, has only this rung",
		},
		{
			name:    "the POD's own type is the rung below the context, and it stays here",
			execCtx: &types.ExecutionContext{},
			envVar:  "the-pods-own-type",
			want:    "the-pods-own-type",
			why:     "this rung is deliberately NOT on types.ExecutionContext: it is a property of the process, and the actions door has a better floor (state.OwnerAgentType) that must not be displaced by it",
		},
		{
			name:    "with nothing to go on the filler is used, because the column is NOT NULL",
			execCtx: &types.ExecutionContext{},
			envVar:  "",
			want:    "generic",
			why:     "the filler is the reason 'generic' is a live value with real traffic — which is why the actions door must never reach for it, and why 'unattributed' rather than 'generic' is that door's sentinel",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("AGENT_TYPE", tc.envVar)
			if got := s.determineOwnerAgentType(tc.execCtx); got != tc.want {
				t.Errorf("determineOwnerAgentType() = %q, want %q\n  rung under test: %s", got, tc.want, tc.why)
			}
		})
	}
}

// TestEnsureFullExecutionContext_BackfillsRunAgentType pins the resumed-step half
// of the same ladder (RFC_019 §7, the declared limitation).
//
// Rung 1 reaches an action unaided only on the first-step / same-message path:
// the processor sets RunAgentType, ToHeaders carries it, FromHeaders restores it.
// A step resumed after an await rebuilds execCtx from the RESPONSE message's
// headers, which never carry it — so without this backfill every consumer of the
// ladder silently drops to rung 2 (the dispatch-path sender, often 'generic') for
// the remainder of the run, and the fix reads as a partial no-op in production
// while every existing test stays green.
//
// state.OwnerAgentType is determineOwnerAgentType's own durable output from run
// start, which is why it is the source rather than the pod's environment.
func TestEnsureFullExecutionContext_BackfillsRunAgentType(t *testing.T) {
	cases := []struct {
		name           string
		execCtx        *types.ExecutionContext
		state          *OrchestrationState
		wantRunAgent   string
		wantSenderType string
		why            string
	}{
		{
			name: "a resumed step gets the run's own resolved agent type back",
			execCtx: &types.ExecutionContext{
				Sender: types.AgentIdentity{AgentType: "generic"},
			},
			state:          &OrchestrationState{OwnerAgentType: "council-gate"},
			wantRunAgent:   "council-gate",
			wantSenderType: "generic",
			why:            "this is the whole case: the response message carried a sender, so the Sender backfill below is a no-op and rung 1 stayed empty — the RFC_019 §7 gap, and the bugs_open/093 shape of one guarded call site with an unguarded sibling",
		},
		{
			name: "a populated run agent type is never overwritten",
			execCtx: &types.ExecutionContext{
				RunAgentType: "page-build-handler",
				Sender:       types.AgentIdentity{AgentType: "generic"},
			},
			state:          &OrchestrationState{OwnerAgentType: "council-gate"},
			wantRunAgent:   "page-build-handler",
			wantSenderType: "generic",
			why:            "the message's own resolved answer is more specific than the run-start record; this function backfills what is MISSING and must not restate what is present",
		},
		{
			name:           "nothing is invented when the state has no answer either",
			execCtx:        &types.ExecutionContext{},
			state:          &OrchestrationState{},
			wantRunAgent:   "",
			wantSenderType: "",
			why:            "an empty RunAgentType falls through to the rungs below it; filling it with a placeholder here would hand the actions door the 'generic' filler that RSH-008 chose 'unattributed' to avoid colliding with",
		},
		{
			name:           "the two siblings are backfilled together when both are empty",
			execCtx:        &types.ExecutionContext{},
			state:          &OrchestrationState{OwnerAgentType: "council-gate", OwnerAgentID: "agent-1"},
			wantRunAgent:   "council-gate",
			wantSenderType: "council-gate",
			why:            "the pre-existing Sender backfill keeps working unchanged; this is the no-op half of the check, and it is what would catch the new block being written in place of the old one rather than beside it",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ensureFullExecutionContext(tc.execCtx, tc.state, "test-pod", zap.NewNop())

			if got := tc.execCtx.RunAgentType; got != tc.wantRunAgent {
				t.Errorf("RunAgentType = %q, want %q\n  case under test: %s", got, tc.wantRunAgent, tc.why)
			}
			if got := tc.execCtx.Sender.AgentType; got != tc.wantSenderType {
				t.Errorf("Sender.AgentType = %q, want %q\n  case under test: %s", got, tc.wantSenderType, tc.why)
			}
		})
	}
}
