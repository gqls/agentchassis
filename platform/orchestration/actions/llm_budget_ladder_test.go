// FILE: platform/orchestration/actions/llm_budget_ladder_test.go
//
// THE PRECEDENCE CONTRACT FOR A REQUEST-SIZING KEY, asserted rather than
// described (bugs_open/257 round 3, owner decision 4, 2026-09-04).
//
// `llm_budget_call_sites_test.go` is the sibling of this file and answers a
// different question: it refuses a hardcoded budget anywhere in the package. It
// cannot see the failure THIS file is about, and says so in its own header —
// config that is declared and read by nobody. A literal check watches the code;
// this watches the rule the code implements.
//
// The two cases named LIVE below are the ones that were actually wrong on
// 2026-09-04, taken from `agent_definitions` and not invented. They are here so
// that a future edit to the ladder cannot re-break them quietly: each of them
// FAILS against the code this round replaced, which is what makes them a
// regression test rather than a restatement.
package actions

import (
	"encoding/json"
	"testing"

	"go.uber.org/zap"
)

func TestBudgetLadderPrefersTheMostSpecificLevel(t *testing.T) {
	// Every level declares a distinct number, so the winner names itself.
	agent := map[string]interface{}{
		"max_tokens": 6.0,
		"ai_service": map[string]interface{}{"max_tokens": 5.0},
	}
	workflowStep := map[string]interface{}{
		"max_tokens": 4.0,
		"ai_service": map[string]interface{}{"max_tokens": 3.0},
	}
	runtimeStep := map[string]interface{}{
		"max_tokens": 2.0,
		"ai_service": map[string]interface{}{"max_tokens": 1.0},
	}

	cases := []struct {
		name                               string
		agentCfg, workflowStepCfg, stepCfg map[string]interface{}
		wantValue                          int
		wantFrom                           string
	}{
		{"runtime step ai_service outranks everything", agent, workflowStep, runtimeStep, 1, "step_config.ai_service"},
		{"runtime step bare key beats the workflow step",
			agent, workflowStep, map[string]interface{}{"max_tokens": 2.0}, 2, "step_config"},
		{"workflow step ai_service beats the agent",
			agent, workflowStep, nil, 3, "workflow_step.ai_service"},
		{"workflow step bare key beats the agent",
			agent, map[string]interface{}{"max_tokens": 4.0}, nil, 4, "workflow_step"},
		{"agent ai_service beats the agent bare key",
			agent, nil, nil, 5, "root.ai_service"},
		{"agent bare key is the last resort",
			map[string]interface{}{"max_tokens": 6.0}, nil, nil, 6, "root"},
		{"nothing declared anywhere resolves to nothing", nil, nil, nil, 0, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, from, _ := resolveBudgetKey("max_tokens",
				budgetLevelsForStep(tc.agentCfg, tc.workflowStepCfg, tc.stepCfg))
			if got != tc.wantValue || from != tc.wantFrom {
				t.Errorf("resolveBudgetKey = (%d, %q), want (%d, %q)", got, from, tc.wantValue, tc.wantFrom)
			}
		})
	}
}

// TestLiveShadowedRootKeyNoLongerCapsItsOwnStep is failure (1) of the two the
// ladder was written for, in the exact shape ten live agents carried.
//
// content-creator-about, measured 2026-09-04: a leftover top-level `max_tokens`
// of 2000 on the agent, and 8000 declared on `generate_about_content` — its only
// LLM step — in the canonical place. The reader this replaced took the agent's
// key FIRST, so the step ran at 2000 and the configuration read correctly to
// anyone inspecting it.
func TestLiveShadowedRootKeyNoLongerCapsItsOwnStep(t *testing.T) {
	agent := map[string]interface{}{
		"max_tokens": 2000.0,
		"ai_service": map[string]interface{}{
			"model": "claude-sonnet-5", "provider": "anthropic",
		},
	}
	step := map[string]interface{}{
		"ai_service": map[string]interface{}{"max_tokens": 8000.0},
	}

	got, from, shadowed := resolveBudgetKey("max_tokens", budgetLevelsForStep(agent, step, nil))
	if got != 8000 {
		t.Fatalf("the step declared 8000 in the canonical place and resolved to %d from %q — "+
			"an agent-level leftover is capping its own step again", got, from)
	}
	if from != "workflow_step.ai_service" {
		t.Errorf("resolved from %q, want workflow_step.ai_service", from)
	}
	// The overridden declaration must be reported, not merely lost: an operator
	// reading 2000 in the definition needs to be told it is inert.
	if len(shadowed) != 1 || shadowed[0] != "root" {
		t.Errorf("shadowed = %v, want [root] — the losing declaration must be named", shadowed)
	}
}

// TestLiveMisplacedStepKeyIsNowRead is failure (2): the key one brace outside the
// block it was meant for.
//
// site-adoption-agent, measured 2026-09-04. All four of its LLM steps carry an
// `ai_service` block holding model/provider/api_key_env_var — which ARE read —
// and `max_tokens` as a SIBLING of that block, which was read by nothing. The
// four asked for 32000/6000/4000/4000 and every one of them ran at the agent's
// root 16000.
func TestLiveMisplacedStepKeyIsNowRead(t *testing.T) {
	agent := map[string]interface{}{
		"ai_service": map[string]interface{}{
			"model": "claude-sonnet-4-6", "provider": "anthropic", "max_tokens": 16000.0,
		},
	}
	for _, tc := range []struct {
		step string
		want int
	}{
		{"analyze_site", 32000},
		{"derive_content_direction", 6000},
		{"classify_archetype", 4000},
		{"generate_design_intent", 4000},
	} {
		t.Run(tc.step, func(t *testing.T) {
			stepCfg := map[string]interface{}{
				"max_tokens": float64(tc.want),
				"ai_service": map[string]interface{}{
					"model": "claude-sonnet-4-6", "provider": "anthropic",
				},
			}
			got, from, _ := resolveBudgetKey("max_tokens", budgetLevelsForStep(agent, stepCfg, nil))
			if got != tc.want {
				t.Errorf("step declared %d and resolved to %d from %q — the bare spelling is unread again",
					tc.want, got, from)
			}
		})
	}
}

// TestBudgetLadderAcceptsEveryShapeTheKeyArrivesIn. jsonb decodes to float64 and
// viper to int; the reader this replaced was float64-ONLY, so a YAML-configured
// budget was silently dropped at this call site while platform/aiservice honoured
// it. Two readers of one key that disagree about what counts as configured is the
// drift bugs_open/257 is about, so they are asserted to agree.
func TestBudgetLadderAcceptsEveryShapeTheKeyArrivesIn(t *testing.T) {
	for name, value := range map[string]interface{}{
		"jsonb float64": 4096.0,
		"viper int":     4096,
		"int64":         int64(4096),
		"json.Number":   json.Number("4096"),
	} {
		t.Run(name, func(t *testing.T) {
			got, from, _ := resolveBudgetKey("max_tokens",
				budgetLevelsForStep(map[string]interface{}{"max_tokens": value}, nil, nil))
			if got != 4096 || from != "root" {
				t.Errorf("%v (%T) resolved to (%d, %q), want (4096, \"root\")", value, value, got, from)
			}
		})
	}

	// Zero and negative are UNSET, not "no limit": max_tokens: 0 is a hard 400
	// from Anthropic. A level holding one must fall through to the next.
	for name, value := range map[string]interface{}{"zero": 0.0, "negative": -1.0, "string": "8000"} {
		t.Run("unusable/"+name, func(t *testing.T) {
			got, from, _ := resolveBudgetKey("max_tokens", budgetLevelsForStep(
				map[string]interface{}{"max_tokens": 5000.0}, // the next level down
				map[string]interface{}{"max_tokens": value},
				nil))
			if got != 5000 || from != "root" {
				t.Errorf("%v at the step level resolved to (%d, %q); an unusable value must fall through, want (5000, \"root\")",
					value, got, from)
			}
		})
	}
}

// TestDirectCallerLadderStaysTwoLevels pins the deliberate difference between the
// two ladders (bugs_open/257 candidate 2, ruled option (c) on 2026-09-04: the
// rules differ permanently and the difference is written down). A direct caller
// has the step's config and an already-merged ai_service map, and NO agent-level
// config — so it must not acquire an agent level by accident.
func TestDirectCallerLadderStaysTwoLevels(t *testing.T) {
	stepCfg := map[string]interface{}{"max_tokens": 8000.0}
	aiCfg := map[string]interface{}{"max_tokens": 2000.0}

	if got, ok := intFromConfig(stepCfg, aiCfg, "max_tokens"); !ok || got != 8000 {
		t.Errorf("intFromConfig = (%d, %v), want (8000, true) — the step must beat the ai_service block", got, ok)
	}
	if got, ok := intFromConfig(nil, aiCfg, "max_tokens"); !ok || got != 2000 {
		t.Errorf("intFromConfig = (%d, %v), want (2000, true) — the ai_service block is the fallback", got, ok)
	}
	if _, ok := intFromConfig(nil, nil, "max_tokens"); ok {
		t.Error("intFromConfig reported a budget with nothing configured")
	}

	opts := llmOptionsFromConfig(stepCfg, aiCfg, zap.NewNop(), "test")
	if opts["max_tokens"] != 8000 {
		t.Errorf("llmOptionsFromConfig sent max_tokens=%v, want 8000", opts["max_tokens"])
	}
	// budget_tokens is forwarded only when present: the client enables extended
	// thinking on the PRESENCE of the key, so an absent one must stay absent.
	if _, present := opts["budget_tokens"]; present {
		t.Error("budget_tokens was forwarded although nothing declared it — that turns thinking on")
	}
}

// TestWorkflowStepConfigFindsOnlyATopLevelStep. A loop body's step name is
// synthesised at runtime and never appears in workflow.steps, so this traversal
// must return nil for it — the nested config arrives as the runtime StepConfig,
// which is a higher level in the ladder. Asserted because a silent non-nil here
// would feed the ladder a DIFFERENT step's config, which reads as working.
func TestWorkflowStepConfigFindsOnlyATopLevelStep(t *testing.T) {
	agent := map[string]interface{}{
		"workflow": map[string]interface{}{
			"steps": map[string]interface{}{
				"process_sections_loop": map[string]interface{}{
					"config": map[string]interface{}{
						"sub_workflow": map[string]interface{}{
							"steps": map[string]interface{}{
								"rewrite_negations": map[string]interface{}{
									"config": map[string]interface{}{"max_tokens": 15999.0},
								},
							},
						},
					},
				},
			},
		},
	}

	if cfg := workflowStepConfig(agent, "process_sections_loop"); cfg == nil {
		t.Error("workflowStepConfig did not find a top-level step that exists")
	}
	for _, absent := range []string{
		"process_sections_loop_iter_0_rewrite_negations", // the synthesised runtime name
		"rewrite_negations", // the nested name, not a top-level step
		"no_such_step",
	} {
		if cfg := workflowStepConfig(agent, absent); cfg != nil {
			t.Errorf("workflowStepConfig(%q) returned %v; a step that is not top-level must resolve to nil, "+
				"or the ladder reads a config the runtime never used", absent, cfg)
		}
	}
	if cfg := workflowStepConfig(nil, "anything"); cfg != nil {
		t.Errorf("workflowStepConfig(nil, …) = %v, want nil", cfg)
	}
}
