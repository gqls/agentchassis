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
