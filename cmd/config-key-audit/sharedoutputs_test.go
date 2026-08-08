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
	"strings"
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

// TestSharedOutputs_SubstepsShapeIsSeenToo is the same fixture in the OTHER
// nesting shape, and it is the reason this file's descent is no longer its own.
//
// A loop declares its body as EITHER `substeps` or `sub_workflow`, and at
// execution `substeps` WINS: loop_actions.go:91 reads it first and consults
// sub_workflow only when it is absent or empty (the precedence
// validation.subWorkflowsOf mirrors exactly). This detector's original descent
// read `sub_workflow` ONLY — so it was blind to the shape that takes priority
// at runtime, which is bugs_open/144's founding failure ("two hand-written
// traversals, blind in the same direction, agreeing with each other") in the
// direction that matters most.
//
// It cost nothing when written: measured 2026-08-08 over live agent_definitions
// (is_active, non-snapshot, not deleted) with a recursive jsonpath
// `$.** ? (@.substeps != null)`, ZERO live definitions carry the shape at any
// depth — the only two rows that do are soft-deleted multipage-website-builder
// definitions. That is precisely why a test is the right place for this: the
// gap was invisible to every live run and would have stayed invisible until the
// first author who preferred `substeps` silently dropped out of view.
func TestSharedOutputs_SubstepsShapeIsSeenToo(t *testing.T) {
	agent := liveAgent{
		Type: "looper-substeps",
		Workflow: models.WorkflowPlan{
			Steps: map[string]models.Step{
				"loop": {
					Action:   "loop",
					NextStep: "after",
					Config: map[string]interface{}{
						"substeps": map[string]interface{}{
							"inner": map[string]interface{}{
								"action":       "producer_action",
								"output_field": "shared_key",
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
		t.Fatalf("substeps-nested producer -> post-container refiner not surfaced (%d findings): %+v — "+
			"substeps is the shape the loop executor PREFERS, so a descent that cannot see it "+
			"is blind exactly where the runtime looks first", len(findings), findings)
	}
}

// TestSharedOutputs_BothShapesResolveToTheExecutedOne pins the precedence
// itself, not merely the presence of each shape. A step carrying BOTH must be
// walked as `substeps`, because that is the half that runs; walking the
// sub_workflow half would make this detector report on config that never
// executes. The two halves are given DIFFERENT output_fields so the assertion
// can only pass for one of them.
func TestSharedOutputs_BothShapesResolveToTheExecutedOne(t *testing.T) {
	agent := liveAgent{
		Type: "looper-both",
		Workflow: models.WorkflowPlan{
			Steps: map[string]models.Step{
				"loop": {
					Action:   "loop",
					NextStep: "after",
					Config: map[string]interface{}{
						"substeps": map[string]interface{}{
							"executed_inner": map[string]interface{}{
								"action":       "producer_action",
								"output_field": "shared_key",
							},
						},
						"sub_workflow": map[string]interface{}{
							"steps": map[string]interface{}{
								"inert_inner": map[string]interface{}{
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
	if len(findings) != 1 {
		t.Fatalf("want exactly 1 finding (the EXECUTED half only), got %d: %+v", len(findings), findings)
	}
	if findings[0].Producer != "executed_inner" {
		t.Fatalf("reported producer %q — the inert sub_workflow half was walked instead of the "+
			"substeps half that actually executes: %+v", findings[0].Producer, findings[0])
	}
}

// TestParseSharedOutputArgs pins the argument handling, which became
// load-bearing when this check went on a schedule.
//
// The old code read --ack at exactly os.Args[2]/[3]. Passed anywhere else it was
// ignored SILENTLY, so the ratchet reverted to a raw check that exits 1 for ever
// — and a scheduled job that always fails is one everybody learns to ignore,
// which is the same blindness this check exists to prevent, reached by a
// different route. Position-independence and refusal-on-typo are the fix, so
// they get a test rather than a comment.
func TestParseSharedOutputArgs(t *testing.T) {
	cases := []struct {
		name    string
		argv    []string
		wantAck string
		wantRep bool
		wantErr bool
	}{
		{
			name:    "canonical order, as the runbook documents it",
			argv:    []string{"--shared-output-fields", "--ack", "scripts/shared_output_fields_ack.txt"},
			wantAck: "scripts/shared_output_fields_ack.txt",
		},
		{
			// The case the old positional read got WRONG: --report first pushes
			// --ack past argv[2], and the ack file was then silently ignored.
			name:    "--ack AFTER --report is still honoured",
			argv:    []string{"--shared-output-fields", "--report", "--ack", "/app/ack.txt"},
			wantAck: "/app/ack.txt",
			wantRep: true,
		},
		{
			name:    "--report alone",
			argv:    []string{"--shared-output-fields", "--report"},
			wantRep: true,
		},
		{
			name:    "no ack file is legal — that is the raw form the wrapper runs",
			argv:    []string{"--shared-output-fields"},
			wantAck: "",
		},
		{
			// A typo must REFUSE, not proceed with the flag dropped: proceeding
			// is how a ratchet loosens with nobody choosing to loosen it.
			name:    "a typo'd flag is refused, not ignored",
			argv:    []string{"--shared-output-fields", "--akc", "scripts/shared_output_fields_ack.txt"},
			wantErr: true,
		},
		{
			name:    "--ack with no value is refused",
			argv:    []string{"--shared-output-fields", "--ack"},
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSharedOutputArgs(tc.argv)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want an error, got %+v — a silently-dropped flag is the defect this test exists for", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ackPath != tc.wantAck {
				t.Errorf("ackPath = %q, want %q", got.ackPath, tc.wantAck)
			}
			if got.report != tc.wantRep {
				t.Errorf("report = %v, want %v", got.report, tc.wantRep)
			}
		})
	}
}

// TestSharedOutputRunSummary_StatesScopeNotJustResult: the doc_notes body must
// carry the SCOPE of the run, because a finding count alone cannot tell "looked
// at 177 agents and found nothing" from "looked at 3 and found nothing", and the
// second is a broken export reporting success. Also pins that an undecodable
// agent row is disclosed — a clean result that silently excludes rows is the
// pass-bought-by-going-blind case.
func TestSharedOutputRunSummary_StatesScopeNotJustResult(t *testing.T) {
	clean := sharedOutputRunSummary(177, 0, nil, []sharedOutputFinding{{Agent: "a"}, {Agent: "b"}}, nil)
	for _, want := range []string{"CLEAN", "177 live agents", "routing keys", "2 acknowledged"} {
		if !strings.Contains(clean, want) {
			t.Errorf("clean summary missing %q: %s", want, clean)
		}
	}

	withUndecoded := sharedOutputRunSummary(3, 4, nil, nil, nil)
	if !strings.Contains(withUndecoded, "WARNING") || !strings.Contains(withUndecoded, "does not cover them") {
		t.Errorf("a run with undecodable rows must say its clean result does not cover them: %s", withUndecoded)
	}

	withFinding := sharedOutputRunSummary(177, 0, []sharedOutputFinding{{
		Agent: "page-build-handler", OutputField: "section_plan",
		Producer: "plan_sections", ProducerAct: "plan_sections",
		Refiner: "load_current_section_content", RefinerAct: "load_current_section_content",
	}}, nil, []string{"stale-agent field prod refine"})
	for _, want := range []string{"1 NEW hazard", "page-build-handler", "section_plan", "STALE ACK", "shared_output_fields_ack.txt"} {
		if !strings.Contains(withFinding, want) {
			t.Errorf("finding summary missing %q: %s", want, withFinding)
		}
	}
}
