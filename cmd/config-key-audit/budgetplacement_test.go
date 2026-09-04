// FILE: cmd/config-key-audit/budgetplacement_test.go
//
// The cases here are the live fleet's own shapes, measured 2026-09-04, not
// invented ones — including the two the FIRST cut of this check got wrong, which
// are asserted as NON-findings so they cannot come back.
package main

import (
	"encoding/json"
	"testing"

	"github.com/gqls/agentchassis/pkg/models"
)

func budgetAgent(t *testing.T, root map[string]interface{}, steps map[string]models.Step) liveAgent {
	t.Helper()
	raw, err := json.Marshal(root)
	if err != nil {
		t.Fatalf("marshalling root config: %v", err)
	}
	return liveAgent{Type: "test-agent", AgentConfig: raw, Workflow: models.WorkflowPlan{Steps: steps}}
}

func budgetService(maxTokens interface{}) map[string]interface{} {
	m := map[string]interface{}{"model": "claude-sonnet-5", "provider": "anthropic"}
	if maxTokens != nil {
		m["max_tokens"] = maxTokens
	}
	return m
}

func budgetKinds(t *testing.T, agent liveAgent) map[string]int {
	t.Helper()
	report := budgetPlacementFindings([]liveAgent{agent})
	kinds := map[string]int{}
	for _, f := range report.Findings {
		kinds[f.Kind]++
	}
	return kinds
}

// TestAStepWithNoBudgetAnywhereIsUnconfigured — provocation-generator-manual.gate,
// the one live instance on 2026-09-04. It declares a model and runs at 2048.
func TestAStepWithNoBudgetAnywhereIsUnconfigured(t *testing.T) {
	agent := budgetAgent(t, map[string]interface{}{}, map[string]models.Step{
		"gate": {Action: "gate_provocation", Config: map[string]interface{}{"ai_service": budgetService(nil)}},
	})
	report := budgetPlacementFindings([]liveAgent{agent})
	if len(report.Findings) != 1 || report.Findings[0].Kind != "unconfigured" {
		t.Fatalf("findings = %+v, want exactly one unconfigured", report.Findings)
	}
	if report.Findings[0].Effective != 0 || report.Findings[0].From != "" {
		t.Errorf("an unconfigured step must report no effective level, got %d from %q",
			report.Findings[0].Effective, report.Findings[0].From)
	}
}

// TestARootDefaultOverriddenByAStepIsNotAFinding is the correction the first run
// forced, and it is asserted rather than merely written down.
//
// feed-triage, live: root ai_service.max_tokens 4000, steps 8000 and 8192. That is
// the documented overlay design — resolveAIServiceConfig's own comment says the
// root block is the fleet default and the step overrides it key-by-key. The first
// cut of this check called it "shadowed" and produced 18 findings across the fleet,
// every one healthy.
func TestARootDefaultOverriddenByAStepIsNotAFinding(t *testing.T) {
	agent := budgetAgent(t,
		map[string]interface{}{"ai_service": budgetService(4000.0)},
		map[string]models.Step{
			"extract_event_facts": {Action: "execute_llm_prompt", Config: map[string]interface{}{"ai_service": budgetService(8000.0)}},
			"score_relevance":     {Action: "execute_llm_prompt", Config: map[string]interface{}{"ai_service": budgetService(8192.0)}},
		})
	if kinds := budgetKinds(t, agent); len(kinds) != 0 {
		t.Errorf("findings %v — a root default beaten by a step declaration is the overlay design working, "+
			"not a defect; flagging it fires on the healthy majority", kinds)
	}
}

// TestOneNumberWrittenAtTwoLevelsIsNotAFinding. A top-level step's config reaches
// the ladder twice (runtime StepConfig and the definition's own block). 149 live
// steps are in that state and every one is healthy.
func TestOneNumberWrittenAtTwoLevelsIsNotAFinding(t *testing.T) {
	agent := budgetAgent(t,
		map[string]interface{}{"ai_service": budgetService(16000.0)},
		map[string]models.Step{
			"write": {Action: "execute_llm_prompt", Config: map[string]interface{}{"ai_service": budgetService(16000.0)}},
		})
	if kinds := budgetKinds(t, agent); len(kinds) != 0 {
		t.Errorf("findings %v — the same number at two levels is one declaration, not a disagreement", kinds)
	}
}

// TestTwoSpellingsAtOneLevelWithDifferentNumbersIsAmbiguous is the ONE state in
// which the estate's two readers send different numbers for the same step, so it
// is the one that must fail the run rather than be reported as advice.
func TestTwoSpellingsAtOneLevelWithDifferentNumbersIsAmbiguous(t *testing.T) {
	stepCfg := map[string]interface{}{"ai_service": budgetService(16000.0), "max_tokens": 15999.0}
	agent := budgetAgent(t, map[string]interface{}{}, map[string]models.Step{
		"write": {Action: "execute_llm_prompt", Config: stepCfg},
	})
	report := budgetPlacementFindings([]liveAgent{agent})
	if len(report.Findings) != 1 || report.Findings[0].Kind != "ambiguous" {
		t.Fatalf("findings = %+v, want exactly one ambiguous", report.Findings)
	}

	// Equal numbers in both spellings is NOT ambiguous: whichever reader runs, the
	// same request goes out, so there is nothing for an operator to decide.
	same := map[string]interface{}{"ai_service": budgetService(16000.0), "max_tokens": 16000.0}
	agent = budgetAgent(t, map[string]interface{}{}, map[string]models.Step{
		"write": {Action: "execute_llm_prompt", Config: same},
	})
	if kinds := budgetKinds(t, agent); kinds["ambiguous"] != 0 {
		t.Errorf("findings %v — two spellings agreeing on one number is not a disagreement", kinds)
	}
}

// TestABareEffectiveDeclarationIsAdvisoryOnly — site-adoption-agent's four steps
// and html-developer-chunked's three, before migrations 769/770. Honoured by the
// ladder, so it must be reported without failing the run.
func TestABareEffectiveDeclarationIsAdvisoryOnly(t *testing.T) {
	agent := budgetAgent(t,
		map[string]interface{}{"ai_service": budgetService(16000.0)},
		map[string]models.Step{
			"analyze_site": {Action: "execute_llm_prompt", Config: map[string]interface{}{
				"ai_service": budgetService(nil), "max_tokens": 32000.0,
			}},
		})
	report := budgetPlacementFindings([]liveAgent{agent})
	if len(report.Findings) != 1 || report.Findings[0].Kind != "non_canonical" {
		t.Fatalf("findings = %+v, want exactly one non_canonical", report.Findings)
	}
	// The number the operator wrote must be the one reported as effective — the
	// whole defect was that it was not.
	if report.Findings[0].Effective != 32000 {
		t.Errorf("effective = %d, want 32000 — the bare step key must beat the root ai_service block",
			report.Findings[0].Effective)
	}
}

// TestANestedStepIsGradedAtTheLevelItActuallyOccupies. A loop body arrives as the
// runtime StepConfig, never as a workflow step, so its reported level must say so
// — otherwise the report and the pod log disagree about the same call.
func TestANestedStepIsGradedAtTheLevelItActuallyOccupies(t *testing.T) {
	agent := budgetAgent(t, map[string]interface{}{}, map[string]models.Step{
		"loop": {Action: "loop", Config: map[string]interface{}{
			"sub_workflow": map[string]interface{}{
				"start_step": "write",
				"steps": map[string]interface{}{
					"write": map[string]interface{}{
						"action": "execute_llm_prompt",
						"config": map[string]interface{}{"ai_service": budgetService(16000.0)},
					},
				},
			},
		}},
	})
	report := budgetPlacementFindings([]liveAgent{agent})
	if report.StepsScanned != 1 {
		t.Fatalf("steps_scanned = %d, want 1 — the walk is not descending into the loop body", report.StepsScanned)
	}
	if len(report.Findings) != 0 {
		t.Errorf("findings = %+v, want none for a canonically configured nested step", report.Findings)
	}
}

// TestAnExportWithoutTheAgentConfigProjectionIsRefused. Without the agent level the
// ladder is missing its two lowest rungs, so every inheriting step would be graded
// UNCONFIGURED — a confidently wrong report, which is worse than no report.
func TestAnExportWithoutTheAgentConfigProjectionIsRefused(t *testing.T) {
	bare := []liveAgent{{Type: "a", Workflow: models.WorkflowPlan{Steps: map[string]models.Step{}}}}
	if err := requireAgentConfigProjection(bare); err == nil {
		t.Error("an export with no agent_config projection was accepted; every verdict would be wrong in the same direction")
	}
	// A JSON null IS a projection — the agent genuinely has no root config.
	withNull := []liveAgent{{Type: "a", AgentConfig: json.RawMessage("null")}}
	if err := requireAgentConfigProjection(withNull); err != nil {
		t.Errorf("a projected JSON null was refused: %v — that is an agent with no root config, not a missing projection", err)
	}
}

// TestStepsWithNoModelDeclarationAreNotScanned keeps the report about model steps:
// a create_work_item step has no ai_service block and no budget, and reporting it
// as unconfigured would bury the one step that really is.
func TestStepsWithNoModelDeclarationAreNotScanned(t *testing.T) {
	agent := budgetAgent(t, map[string]interface{}{}, map[string]models.Step{
		"file_it":  {Action: "create_work_item", Config: map[string]interface{}{"item_type": "x"}},
		"complete": {Action: "complete_workflow"},
	})
	report := budgetPlacementFindings([]liveAgent{agent})
	if report.StepsScanned != 0 || len(report.Findings) != 0 {
		t.Errorf("scanned %d steps, %d findings — a step that never declares a model is not this report's business",
			report.StepsScanned, len(report.Findings))
	}
}

// TestBudgetTokensAgainstARejectingModelIsFatal — the trap a peer lane hit
// first-hand on 2026-09-04, four times, replaying a stored prompt against
// claude-sonnet-5. Live exposure is ZERO (no active agent declares the key), so
// this test IS the exercise: without it the arm would ship never having fired.
func TestBudgetTokensAgainstARejectingModelIsFatal(t *testing.T) {
	agent := budgetAgent(t,
		map[string]interface{}{"ai_service": budgetService(16000.0)},
		map[string]models.Step{
			"write": {Action: "execute_llm_prompt", Config: map[string]interface{}{
				"ai_service": map[string]interface{}{
					"model": "claude-sonnet-5", "provider": "anthropic",
					"max_tokens": 16000.0, "budget_tokens": 10000.0,
				},
			}},
		})
	report := budgetPlacementFindings([]liveAgent{agent})
	if len(report.Findings) != 1 || report.Findings[0].Kind != "thinking_unsupported" {
		t.Fatalf("findings = %+v, want exactly one thinking_unsupported — a budget_tokens "+
			"declaration against claude-sonnet-5 is a 400 on every call for this step", report.Findings)
	}
	if report.Findings[0].Effective != 10000 {
		t.Errorf("effective = %d, want the declared 10000", report.Findings[0].Effective)
	}
}

// TestBudgetTokensAgainstAnAcceptingModelIsNotAFinding is the other half, and it
// is why the arm is model-aware rather than a blanket refusal: claude-haiku-4-5
// carries 32 model declarations across 24 live agents and REQUIRES budget_tokens
// for thinking to happen at all. Flagging it would be telling operators to break
// the only models where the key still works.
func TestBudgetTokensAgainstAnAcceptingModelIsNotAFinding(t *testing.T) {
	for _, model := range []string{"claude-haiku-4-5", "claude-sonnet-4-6", "claude-opus-4-6"} {
		t.Run(model, func(t *testing.T) {
			agent := budgetAgent(t, map[string]interface{}{}, map[string]models.Step{
				"write": {Action: "execute_llm_prompt", Config: map[string]interface{}{
					"ai_service": map[string]interface{}{
						"model": model, "provider": "anthropic",
						"max_tokens": 8000.0, "budget_tokens": 4000.0,
					},
				}},
			})
			if kinds := budgetKinds(t, agent); kinds["thinking_unsupported"] != 0 {
				t.Errorf("%s was flagged; it accepts a manual thinking budget (4.6 deprecated-but-functional, "+
					"4.5 and older require it)", model)
			}
		})
	}
}

// TestTheThinkingArmInheritsTheAgentModel — the model is resolved by the same
// overlay production uses, so a step that names no model of its own is judged
// against the agent root one. Getting this wrong would silently exempt every
// step that inherits, which is most of them.
func TestTheThinkingArmInheritsTheAgentModel(t *testing.T) {
	agent := budgetAgent(t,
		map[string]interface{}{"ai_service": map[string]interface{}{
			"model": "claude-sonnet-5", "provider": "anthropic", "max_tokens": 16000.0,
		}},
		map[string]models.Step{
			"write": {Action: "execute_llm_prompt", Config: map[string]interface{}{
				"ai_service": map[string]interface{}{"budget_tokens": 4000.0},
			}},
		})
	if kinds := budgetKinds(t, agent); kinds["thinking_unsupported"] != 1 {
		t.Errorf("findings %v — a step declaring no model of its own must be judged against the "+
			"agent root model, which is what the ai_service overlay does at runtime", kinds)
	}
}
