// FILE: cmd/scheduler/main_test.go
//
// bugs_open/074 — a scheduled task's inline workflow is silently ignored.
//
// The detector below is what stops the silence. It has to fire on the shapes
// that were actually authored in production, and it has to stay off ordinary
// payloads: a false positive here refuses a task that works today, which is
// worse than the bug.

package main

import (
	"encoding/json"
	"testing"
)

func TestDiscardedInlineWorkflow(t *testing.T) {
	cases := []struct {
		name          string
		inputData     string
		wantFound     bool
		wantStartStep string
	}{
		{
			// Verbatim shape of `evidence-freshness` (claims_verification), which
			// fired daily and never once ran its sweep.
			name: "the authored trap — workflow under config",
			inputData: `{"config":{"agent_type":"generic","workflow":{
				"start_step":"refresh_evidence","processing_mode":"orchestrator","timeout_seconds":600,
				"steps":{"refresh_evidence":{"action":"refresh_evidence_base","next_step":"complete"},
				         "complete":{"action":"complete_workflow"}}}}}`,
			wantFound:     true,
			wantStartStep: "refresh_evidence",
		},
		{
			// A workflow with no start_step is just as undeliverable. Report it,
			// with an empty start_step rather than a miss.
			name:          "malformed workflow — object without start_step",
			inputData:     `{"config":{"workflow":{"steps":{}}}}`,
			wantFound:     true,
			wantStartStep: "",
		},
		{
			// Presence is structural: a workflow that is not even an object is
			// discarded exactly as silently, so it must still be caught.
			name:          "malformed workflow — not an object",
			inputData:     `{"config":{"workflow":"refresh_evidence"}}`,
			wantFound:     true,
			wantStartStep: "",
		},
		{
			// `ch-fetch-accounts` — envelope-ish, but carries no workflow. Legal.
			name:      "config without a workflow is left alone",
			inputData: `{"config":{"agent_type":"ch-accounts-fetcher"}}`,
			wantFound: false,
		},
		{
			// `diagnose-pipeline-trigger` — the bugs_closed/054 family. Harmless
			// (its inner payload is empty) and must keep firing.
			name:      "the 054 envelope family keeps firing",
			inputData: `{"action":"orchestrate","config":{"agent_type":"diagnose-dispatch-loop"},"input_data":{}}`,
			wantFound: false,
		},
		{
			// The overwhelming majority: a plain payload.
			name:      "ordinary payload",
			inputData: `{"batch_size":20,"vertical_slug":"veterinary"}`,
			wantFound: false,
		},
		{
			// A payload field that merely mentions a workflow one level away is
			// not the trap — only config.workflow is undeliverable.
			name:      "workflow elsewhere in the payload is not the trap",
			inputData: `{"workflow":"nightly","settings":{"workflow":{"start_step":"x"}}}`,
			wantFound: false,
		},
		{
			name:      "empty input_data",
			inputData: `{}`,
			wantFound: false,
		},
		{
			name:      "config is not an object",
			inputData: `{"config":"generic"}`,
			wantFound: false,
		},
		{
			name:      "input_data is not an object",
			inputData: `"just a string"`,
			wantFound: false,
		},
		{
			name:      "input_data is not valid JSON",
			inputData: `{not json`,
			wantFound: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			startStep, found := discardedInlineWorkflow(json.RawMessage(tc.inputData))
			if found != tc.wantFound {
				t.Fatalf("found = %v, want %v (input_data: %s)", found, tc.wantFound, tc.inputData)
			}
			if found && startStep != tc.wantStartStep {
				t.Fatalf("start_step = %q, want %q", startStep, tc.wantStartStep)
			}
		})
	}
}

// A nil/absent input_data column must not panic — the column is nullable in
// spirit even though it defaults to '{}'.
func TestDiscardedInlineWorkflowNilInput(t *testing.T) {
	if _, found := discardedInlineWorkflow(nil); found {
		t.Fatal("nil input_data reported as carrying a workflow")
	}
}
