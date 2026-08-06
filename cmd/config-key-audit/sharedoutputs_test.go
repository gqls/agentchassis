// FILE: cmd/config-key-audit/sharedoutputs_test.go
//
// The acceptance bar is fixed by RFC_012's owner ruling: both naive detectors
// return 0 on the known bug (bugs_open/192), so a candidate implementation
// must PROVE ITSELF AGAINST THAT CASE before it counts. The first fixture IS
// that case: producer and refiner share output_field "section_plan" with
// different actions, and the only path between them runs through a
// conditional's config.then_step — invisible to a direct-edge check and to a
// transitive walk over next_step/error_step alone.
package main

import (
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
)

// bug192Workflow reproduces the pre-fix page-build-handler shape from
// RFC_012 addendum 1: plan_sections writes section_plan; a conditional routes
// via config.then_step; the step it routes to overwrites section_plan with a
// DIFFERENT action.
func bug192Workflow() liveAgent {
	return liveAgent{
		Type: "page-build-handler",
		Workflow: models.WorkflowPlan{
			StartStep: "plan_sections",
			Steps: map[string]models.Step{
				"plan_sections": {
					Action:      "plan_sections",
					OutputField: "section_plan",
					NextStep:    "check_has_ready_sections",
				},
				"check_has_ready_sections": {
					Action: "conditional",
					Config: map[string]interface{}{
						"then_step": "load_current_section_content",
						"else_step": "complete",
					},
				},
				"load_current_section_content": {
					Action:      "load_current_section_content",
					OutputField: "section_plan", // the overwrite
				},
				"complete": {Action: "complete_workflow"},
			},
		},
	}
}

func TestSharedOutputs_FiresOnTheKnownBug(t *testing.T) {
	findings := findSharedOutputFields([]liveAgent{bug192Workflow()})
	if len(findings) != 1 {
		t.Fatalf("got %d findings on bugs_open/192's own shape, want exactly 1 — "+
			"a detector that misses the known bug is the naive form the ruling forbids: %+v", len(findings), findings)
	}
	f := findings[0]
	if f.Producer != "plan_sections" || f.Refiner != "load_current_section_content" || f.OutputField != "section_plan" {
		t.Fatalf("wrong finding: %+v", f)
	}
}

// TestSharedOutputs_TheConfigEdgeIsLoadBearing is the mutation the ruling
// implies: remove the conditional's config routing and the pair becomes
// unreachable — if the finding SURVIVES this mutation, the detector is not
// actually reading the graph it claims to read.
func TestSharedOutputs_TheConfigEdgeIsLoadBearing(t *testing.T) {
	agent := bug192Workflow()
	cond := agent.Workflow.Steps["check_has_ready_sections"]
	cond.Config = map[string]interface{}{} // sever the then_step edge
	agent.Workflow.Steps["check_has_ready_sections"] = cond

	if findings := findSharedOutputFields([]liveAgent{agent}); len(findings) != 0 {
		t.Fatalf("finding survived severing the only reaching edge — the graph walk is not real: %+v", findings)
	}
}

// TestSharedOutputs_SameActionRetryLoopIsBenign pins the discriminator: the
// propose -> repropose retry loops share an output_field and ARE sequentially
// reachable, and are structurally benign — same action, same shape. Reporting
// them is how 24 pairs became noise.
func TestSharedOutputs_SameActionRetryLoopIsBenign(t *testing.T) {
	agent := liveAgent{
		Type: "experience-planner",
		Workflow: models.WorkflowPlan{
			Steps: map[string]models.Step{
				"propose": {
					Action:      "execute_llm_prompt",
					OutputField: "proposal",
					NextStep:    "review",
				},
				"review": {
					Action: "conditional",
					Config: map[string]interface{}{"then_step": "repropose"},
				},
				"repropose": {
					Action:      "execute_llm_prompt", // SAME action
					OutputField: "proposal",
				},
			},
		},
	}
	if findings := findSharedOutputFields([]liveAgent{agent}); len(findings) != 0 {
		t.Fatalf("same-action retry loop reported — the discriminator is gone: %+v", findings)
	}
}

// TestSharedOutputs_UnreachablePairIsNotAFinding: two branches that both
// write a key but can never both run in one pass are the mutually-exclusive
// case the addendum's census counted as benign.
func TestSharedOutputs_UnreachablePairIsNotAFinding(t *testing.T) {
	agent := liveAgent{
		Type: "brancher",
		Workflow: models.WorkflowPlan{
			Steps: map[string]models.Step{
				"fork": {
					Action: "conditional",
					Config: map[string]interface{}{"then_step": "left", "else_step": "right"},
				},
				"left":  {Action: "a", OutputField: "result"},
				"right": {Action: "b", OutputField: "result"},
			},
		},
	}
	if findings := findSharedOutputFields([]liveAgent{agent}); len(findings) != 0 {
		t.Fatalf("mutually-exclusive branches reported: %+v", findings)
	}
}

// TestSharedOutputs_NestedSubWorkflowOverApproximates: a nested step sharing
// an output_field with a step after its container must surface — the
// containment rule errs toward a finding, never toward silence.
func TestSharedOutputs_NestedSubWorkflowOverApproximates(t *testing.T) {
	agent := liveAgent{
		Type: "looper",
		Workflow: models.WorkflowPlan{
			Steps: map[string]models.Step{
				"loop": {
					Action:   "loop",
					NextStep: "after",
					Config: map[string]interface{}{
						"sub_workflow": map[string]interface{}{
							"steps": map[string]interface{}{
								"inner": map[string]interface{}{
									"action":       "producer_action",
									"output_field": "shared_key",
								},
							},
						},
					},
				},
				"after": {Action: "consumer_action", OutputField: "shared_key"},
			},
		},
	}
	findings := findSharedOutputFields([]liveAgent{agent})
	if len(findings) != 1 || findings[0].Producer != "inner" || findings[0].Refiner != "after" {
		t.Fatalf("nested producer -> post-container refiner not surfaced: %+v", findings)
	}
}
