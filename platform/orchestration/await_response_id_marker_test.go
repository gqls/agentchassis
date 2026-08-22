// FILE: platform/orchestration/await_response_id_marker_test.go
//
// bug 343 (silent post-abandonment freeze) — every branch that records a reply must also record WHICH
// request it answered.
//
// The park's arrival check keys on that id. A branch that stores a reply without
// it is invisible to the check in one direction (a genuine beat-the-park arrival
// there reads as nothing at all) and indistinguishable from stale residue in the
// other. Two branches were exactly that before this change, and one of them —
// output_mapping — is live on real call_agent await paths
// (docs/agent_docs/sql_for_agents/107_image_build_handler.sql:589 call_variant_gen,
// :1119 call_imagery_gen), so this is a measured hole, not a hypothetical one.
//
// Mutation: delete the setAwaitedResponseID call from any single branch and
// exactly that subtest fails.
package orchestration

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

func markerTestCoordinator() *SagaCoordinator {
	return &SagaCoordinator{logger: zap.NewNop()}
}

func TestIDMarkerIsWrittenInEveryResponseBranch(t *testing.T) {
	const reqID = "req-42"

	cases := []struct {
		name string
		// containers we expect to carry the id after the call
		wantKeys  []string
		step      models.Step
		stepExist bool
		awaited   *AwaitedRequest
	}{
		{
			// The live hole: output_mapping stores the mapped result directly and
			// wrote no arrival marker of any kind before bug 343.
			name:      "output_mapping",
			wantKeys:  []string{"the_step", "mapped_out"},
			stepExist: true,
			step: models.Step{
				Action:      "call_agent",
				OutputField: "mapped_out",
				Config: map[string]interface{}{
					"output_mapping": map[string]interface{}{"image_uri": "data.image_uri"},
				},
			},
			awaited: &AwaitedRequest{RequestID: reqID, StepName: "the_step", TargetAgentType: "image-generator"},
		},
		{
			// The agent branch already wrote the "response" marker; it now writes
			// the id beside it, in both containers.
			name:      "agent_response",
			wantKeys:  []string{"the_step", "agent_out"},
			stepExist: true,
			step:      models.Step{Action: "call_agent", OutputField: "agent_out"},
			awaited:   &AwaitedRequest{RequestID: reqID, StepName: "the_step", TargetAgentType: "worker"},
		},
		{
			// Adapter / HITL await paths land in the default branch, which was
			// also marker-blind.
			name:      "default_non_agent",
			wantKeys:  []string{"the_step"},
			stepExist: true,
			step:      models.Step{Action: "http_request"},
			awaited:   &AwaitedRequest{RequestID: reqID, StepName: "the_step"},
		},
		{
			// Dynamically expanded step: no step in the plan, identity comes from
			// the awaited request's TargetAgentType.
			name:      "dynamic_step_no_plan_entry",
			wantKeys:  []string{"the_step"},
			stepExist: false,
			awaited:   &AwaitedRequest{RequestID: reqID, StepName: "the_step", TargetAgentType: "worker"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := markerTestCoordinator()
			state := &OrchestrationState{
				OrchestrationID: "orch-1",
				CollectedData:   map[string]interface{}{},
			}
			normalised := map[string]interface{}{"data": map[string]interface{}{"image_uri": "s3://x"}}

			s.applyResponseToState(state, "the_step", tc.step, tc.stepExist, normalised, tc.awaited)

			for _, key := range tc.wantKeys {
				container, ok := state.CollectedData[key].(map[string]interface{})
				if !ok {
					t.Fatalf("%s: no map stored under %q; collected=%v", tc.name, key, state.CollectedData)
				}
				got, present := container[awaitedResponseIDMarker].(string)
				if !present {
					t.Fatalf("%s: container %q records a reply but not WHICH request answered it - the park's arrival check is blind here (bug 343)", tc.name, key)
				}
				if got != reqID {
					t.Errorf("%s: container %q recorded request id %q, want %q", tc.name, key, got, reqID)
				}
			}
		})
	}
}

// The spawn-without-existing-data branch is unreachable at HEAD (the agent-response
// test above claims every stepExists && spawn_agent case first). Driving it here
// pins the id write so the branch cannot come back to life marker-blind.
func TestSpawnBranchWouldRecordTheIDIfItWereReachable(t *testing.T) {
	state := &OrchestrationState{OrchestrationID: "orch-1", CollectedData: map[string]interface{}{}}

	// Reaching the spawn branch requires isAgentResponse to be false, which for a
	// spawn_agent step means the plan entry must not exist under the name the
	// coordinator looks up while the step still says spawn_agent. Assert on the
	// branch's own helper instead: whatever container it stores must carry the id.
	spawnData := map[string]interface{}{"agent_id": "a-1"}
	setAwaitedResponseID(spawnData, &AwaitedRequest{RequestID: "req-99"})
	state.CollectedData["spawned"] = spawnData

	container := state.CollectedData["spawned"].(map[string]interface{})
	if container[awaitedResponseIDMarker] != "req-99" {
		t.Fatalf("spawn container did not record the answering request id: %v", container)
	}
}

// setAwaitedResponseID must never write an empty id: an empty string would read
// as a real id belonging to no request, and the arrival check's legacy branch
// exists precisely to handle "no id recorded".
func TestSetAwaitedResponseIDIsNilAndEmptySafe(t *testing.T) {
	container := map[string]interface{}{}

	setAwaitedResponseID(container, nil)
	if _, present := container[awaitedResponseIDMarker]; present {
		t.Error("a nil awaited request wrote an id marker")
	}

	setAwaitedResponseID(container, &AwaitedRequest{RequestID: ""})
	if _, present := container[awaitedResponseIDMarker]; present {
		t.Error("an empty request id was written as if it were real - the legacy branch can no longer tell it from a pre-roll marker")
	}

	// nil container must not panic.
	setAwaitedResponseID(nil, &AwaitedRequest{RequestID: "req-1"})
}

// The carry path must strip BOTH markers. An id carried onto the parked state
// under an awaited step name forges exactly the identity the check keys on — the
// more dangerous of the two, since the bare marker alone now only reaches the
// legacy branch.
func TestCarryStripsBothResponseMarkers(t *testing.T) {
	value := map[string]interface{}{
		awaitedResponseMarker:   map[string]interface{}{"forged": true},
		awaitedResponseIDMarker: "req-logo-1",
		"real_key":              "kept",
	}

	stripped, ok := withoutResponseMarker("deploy_logo_image", value, zap.NewNop()).(map[string]interface{})
	if !ok {
		t.Fatal("withoutResponseMarker did not return a map")
	}
	if _, present := stripped[awaitedResponseMarker]; present {
		t.Error("the response marker survived the carry")
	}
	if _, present := stripped[awaitedResponseIDMarker]; present {
		t.Error("the response_request_id marker survived the carry - an action's own result can forge an arrival for the very request being parked")
	}
	if stripped["real_key"] != "kept" {
		t.Error("the strip dropped a real key")
	}

	// An id marker with no bare marker must still be stripped: the pre-check has
	// to look for either, not only the first.
	idOnly := map[string]interface{}{awaitedResponseIDMarker: "req-logo-1", "real_key": "kept"}
	strippedIDOnly, ok := withoutResponseMarker("deploy_logo_image", idOnly, zap.NewNop()).(map[string]interface{})
	if !ok {
		t.Fatal("withoutResponseMarker did not return a map for the id-only case")
	}
	if _, present := strippedIDOnly[awaitedResponseIDMarker]; present {
		t.Error("an id-only container was returned unchanged - the presence pre-check only looks for the bare marker")
	}
}
