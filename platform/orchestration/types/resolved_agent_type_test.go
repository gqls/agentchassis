// FILE: platform/orchestration/types/resolved_agent_type_test.go
//
// The hoisted ladder's contract, pinned where the ladder now lives.
//
// This file is written to FAIL against a mutated build, not to be green against
// a correct one — the defect it guards is precisely a ladder that looked right
// and answered one rung short, and a suite that never drove the rungs could not
// see that. Each case below names the rung it kills.
package types

import "testing"

func TestResolvedAgentType_RungOrder(t *testing.T) {
	cases := []struct {
		name string
		ec   *ExecutionContext
		want string
		// why names the rung this case exists to kill, so a failure reads as a
		// diagnosis rather than as a diff.
		why string
	}{
		{
			name: "the resolved run agent wins over the dispatch-path sender",
			ec: &ExecutionContext{
				RunAgentType: "vet-practice-verifier",
				Sender:       AgentIdentity{AgentType: "generic"},
			},
			want: "vet-practice-verifier",
			why:  "this is the whole point: on the generic dispatch path the sender IS 'generic', and reading it is what filed real findings under a filler",
		},
		{
			name: "the sender is used when the run agent type is absent",
			ec: &ExecutionContext{
				Sender: AgentIdentity{AgentType: "page-content-writer"},
			},
			want: "page-content-writer",
			why:  "a dedicated pod, and any message predating the run_agent_type header, has only this rung — dropping it would regress every one of them",
		},
		{
			name: "an empty run agent type does not shadow the sender",
			ec: &ExecutionContext{
				RunAgentType: "",
				Sender:       AgentIdentity{AgentType: "page-build-handler"},
			},
			want: "page-build-handler",
			why:  "a rung that returned early on presence-of-field rather than non-emptiness would blank every row",
		},
		{
			name: "a context that cannot answer says so, rather than guessing",
			ec:   &ExecutionContext{},
			want: "",
			why:  "the empty answer is load-bearing: each caller's own fallback differs, and a filler returned from here would overwrite the better one",
		},
		{
			name: "a nil context answers instead of panicking",
			ec:   nil,
			want: "",
			why:  "ActionParams.ExecutionContext is a pointer and is nil in unit tests and in at least one spawn path; the call site should not need a nil check to ask this question",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.ec.ResolvedAgentType(); got != tc.want {
				t.Errorf("ResolvedAgentType() = %q, want %q\n  rung under test: %s", got, tc.want, tc.why)
			}
		})
	}
}

// The ladder must not read the process environment, however tempting the
// coordinator's third rung is. AGENT_TYPE is a property of the pod, not of the
// message, and a context deserialised from Kafka headers must answer the same
// on every pod that reads it.
//
// t.Setenv is the assertion: it fails the test if the value ever leaks in.
func TestResolvedAgentType_IgnoresTheProcessEnvironment(t *testing.T) {
	t.Setenv("AGENT_TYPE", "the-pods-own-type")

	if got := (&ExecutionContext{}).ResolvedAgentType(); got != "" {
		t.Errorf("ResolvedAgentType() = %q with AGENT_TYPE set — the ladder read the process environment; that rung belongs to the coordinator, whose fallback the actions door must not inherit", got)
	}
}

// A context that survives the header round-trip must still answer. The rung
// this protects is the wire, not the ladder: run_agent_type has to be in BOTH
// ToHeaders and FromHeaders or the resolved type is silently dropped the moment
// a message crosses a topic, and every consumer falls back to the sender
// without anything looking wrong.
func TestResolvedAgentType_SurvivesTheHeaderRoundTrip(t *testing.T) {
	original := &ExecutionContext{
		CorrelationID: "corr-1",
		MessageType:   "request",
		RunAgentType:  "council-gate",
		Sender:        AgentIdentity{AgentType: "generic"},
	}

	restored, err := FromHeaders(original.ToHeaders())
	if err != nil {
		t.Fatalf("FromHeaders: %v", err)
	}

	if got := restored.ResolvedAgentType(); got != "council-gate" {
		t.Errorf("after a ToHeaders/FromHeaders round trip ResolvedAgentType() = %q, want %q — run_agent_type is missing from one of the two header maps, so the resolved type dies at the topic boundary", got, "council-gate")
	}
}
