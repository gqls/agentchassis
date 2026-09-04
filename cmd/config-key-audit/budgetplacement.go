// FILE: cmd/config-key-audit/budgetplacement.go
//
// WHERE DOES EACH LIVE STEP'S OUTPUT-TOKEN BUDGET ACTUALLY COME FROM?
// (bugs_open/257 round 3, owner decision 4, 2026-09-04.)
//
// An operator can write `max_tokens` in four places that all look right, because
// they all sit next to keys that ARE read. Round 3 gave the estate one precedence
// ladder so every one of them is honoured. This is the other half: the report that
// notices the next misplacement, and the one after that.
//
// IT IS HERE RATHER THAN IN A GO TEST BECAUSE A GO TEST CANNOT SEE THIS CLASS, and
// the lane's own guard says so in its header —
// `platform/orchestration/actions/llm_budget_call_sites_test.go`: "WHAT THIS AUDIT
// CANNOT SEE ... Config that is declared and read by nobody. [MEASURED 2026-09-03]
// four `site-adoption-agent` steps declare a top-level `config.max_tokens`
// (32000/4000/6000/4000) and every one of them runs at the root 16000, because no
// reader looks there. That is the same defect from the config end, and no Go test
// can see it." A package-scoped offline test watches the CODE; the defect lives in
// the join between the code and 208 live agent definitions.
//
// IT CALLS PRODUCTION'S OWN LADDER — `actions.ResolveStepBudget` — and carries no
// copy of the precedence rule. A detector that re-implements the rule it is
// checking can only ever confirm itself; `componentsourcevocabulary.go` calls the
// birth gate's own findings function for exactly this reason, and this file
// follows it. If the ladder changes, this report changes with it, with nothing to
// keep in step.
//
// THE THREE FINDING KINDS, and what each one costs if nobody looks:
//
//  1. UNCONFIGURED — a step that declares an `ai_service` block (the operator's own
//     statement that this step talks to a model) and NO budget at any level, so it
//     runs at the provider floor of 2048. That is bugs_open/205's shape, which cost
//     64 truncations before anything said so. [MEASURED 2026-09-04] exactly one live
//     step is in this state: provocation-generator-manual.gate. FAILS the run.
//
//  2. AMBIGUOUS — ONE level declares the budget in BOTH spellings with DIFFERENT
//     numbers (say `root.max_tokens: 2000` beside `root.ai_service.max_tokens: 4000`).
//     The canonical ladder picks the ai_service one; the direct-caller ladder in
//     llmOptionsFromConfig picks the bare one at the step level. So this is the ONE
//     configuration state in which the estate's two readers genuinely disagree, and
//     which of an operator's two numbers is sent depends on which action the step
//     runs. Nothing else in this report depends on knowing that, which is why the
//     mode does not need a list of direct-caller actions. [MEASURED 2026-09-04] zero
//     live instances — this is also migration 769's guard 3. FAILS the run.
//
//  3. NON_CANONICAL — the effective budget is declared outside an `ai_service` block.
//     HONOURED, not broken: the ladder reads it. Reported so the fleet converges on
//     one spelling, because the bare key is where both round-3 failures started and
//     it sits beside model/provider keys that ARE read. ADVISORY — it never fails
//     the run, because an advisory that changes an exit code stops being read as
//     advisory within a week.
//
// WHAT IS DELIBERATELY *NOT* A FINDING, corrected the first time this ran against
// the live fleet: a root declaration beaten by a step declaration. That is the
// documented overlay design — resolveAIServiceConfig's own comment says "the root
// block is the fleet default, the current step's block overrides it key-by-key" —
// and feed-triage is a live agent doing it on purpose (root 4000, steps 8000/8192).
// The first cut reported it and produced 18 findings, every one healthy. A report
// whose first run cries on the healthy majority is one nobody opens twice, and the
// same correction was made to the runtime log line in ai_actions.go in the same
// breath. Two levels holding the SAME number is likewise not a finding: for a
// top-level step the runtime StepConfig and the definition's own block are one
// declaration arriving by two routes.
//
// WHY "declares an ai_service block" IS THE MODEL-STEP TEST, and not a list of
// action names. A hardcoded list of model-calling actions would be a second
// vocabulary to keep in step with the registry, and drift between two lists that
// must agree is the failure class this whole bug is about. The `ai_service` block
// is the operator's own declaration, it lives in the same row being audited, and
// it cannot go stale relative to the code. What it costs is stated rather than
// hidden: a step that calls a model WITHOUT declaring an ai_service block is
// invisible here (it inherits the agent's), and so is every direct caller outside
// the orchestration layer — that is bugs_open/479's territory, not this report's.
//
// VACUITY. Zero findings is a healthy reading; zero SCANNED steps never is. The
// mode refuses to print a clean report over an export with no ai_service blocks
// and no budget declarations at all, because "the check could not run" must never
// read as "the check passed" (016b §9). The stdin export must also carry the
// `agent_config` projection: without the agent's root config the two lowest rungs
// of the ladder are missing and EVERY verdict would be silently wrong, so an
// export without it is refused rather than graded.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/validation"
)

// budgetPlacementExportQuery adds the agent's root config to the projection the
// other wrappers use. `default_config - 'workflow'` rather than the whole thing:
// the workflow is already projected, and shipping it twice doubles an export that
// is already the largest thing this binary reads.
//
// Deliberately NOT the shared fleetExportQuery — that one is character-for-
// character identical to audit-shared-output-fields.sh's and must stay that way.
// --template-input-fields set the precedent for a mode needing its own projection.
const budgetPlacementExportQuery = `
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow',
                                    'agent_config', default_config - 'workflow'))
FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(is_snapshot,false) = false
  AND is_active
  AND default_config ? 'workflow';`

// budgetDeclaration is one place a number was written, and whether it took effect.
type budgetDeclaration struct {
	Level     string `json:"level"`
	Value     int    `json:"value"`
	Canonical bool   `json:"canonical"`
	Effective bool   `json:"effective"`
}

type budgetFinding struct {
	Agent string `json:"agent"`
	Path  string `json:"path"`
	Kind  string `json:"kind"` // shadowed | non_canonical | unconfigured
	// Effective is the number the ladder sends, and From is the level it came
	// from — both taken from actions.ResolveStepBudget, never re-derived here.
	Effective    int                 `json:"effective"`
	From         string              `json:"from"`
	Declarations []budgetDeclaration `json:"declarations"`
	Detail       string              `json:"detail"`
}

// budgetPlacementReport is the whole output. The counts are part of the answer,
// not decoration: "0 findings" means nothing without the size of what was looked at.
type budgetPlacementReport struct {
	AgentsScanned   int             `json:"agents_scanned"`
	AgentsUndecoded int             `json:"agents_undecoded"`
	StepsScanned    int             `json:"steps_scanned"`
	Declarations    int             `json:"declarations"`
	Findings        []budgetFinding `json:"findings"`
}

// agentRootConfig decodes the agent's default_config minus the workflow. An absent
// projection and a JSON null are distinguished by the caller, not here.
func (a liveAgent) agentRootConfig() map[string]interface{} {
	if len(a.AgentConfig) == 0 {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(a.AgentConfig, &m); err != nil {
		return nil
	}
	return m
}

// requireAgentConfigProjection refuses an export that cannot answer the question.
// Without the agent's root config the ladder is missing its two lowest rungs, so
// every step inheriting an agent-level budget would be reported UNCONFIGURED and
// every root-level shadowing would be invisible — a confidently wrong report,
// which is worse than no report. Modelled on requireAgentPromptProjection.
func requireAgentConfigProjection(agents []liveAgent) error {
	for _, a := range agents {
		if len(a.AgentConfig) > 0 {
			return nil
		}
	}
	return fmt.Errorf("config-key-audit --budget-placement: no agent row carries the `agent_config` " +
		"projection.\nThis mode needs `'agent_config', default_config - 'workflow'` in the export: " +
		"without it the two lowest levels of the budget ladder are absent and every verdict would be " +
		"wrong in the same direction. Use scripts/audit-budget-placement.sh, or add the projection.")
}

// subMap is the local read of a nested config block. It is three lines rather than
// an import because package actions keeps its own unexported; what must not be
// duplicated is the PRECEDENCE, and that is called, not copied.
func subMap(m map[string]interface{}, key string) map[string]interface{} {
	if m == nil {
		return nil
	}
	sub, _ := m[key].(map[string]interface{})
	return sub
}

func numeric(m map[string]interface{}, key string) (int, bool) {
	if m == nil {
		return 0, false
	}
	switch v := m[key].(type) {
	case float64:
		if v > 0 {
			return int(v), true
		}
	case int:
		if v > 0 {
			return v, true
		}
	}
	return 0, false
}

// budgetPlacementFindings grades every live step against the shipped ladder.
func budgetPlacementFindings(agents []liveAgent) budgetPlacementReport {
	report := budgetPlacementReport{AgentsScanned: len(agents), Findings: []budgetFinding{}}

	for _, agent := range agents {
		root := agent.agentRootConfig()
		rootService := subMap(root, "ai_service")

		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			stepCfg := step.Config
			stepService := subMap(stepCfg, "ai_service")

			// Only steps that state they talk to a model, or that declare a budget
			// anyway (a budget on a non-model step is itself worth seeing).
			_, bareDecl := numeric(stepCfg, "max_tokens")
			_, canonDecl := numeric(stepService, "max_tokens")
			if stepService == nil && !bareDecl && !canonDecl {
				return
			}
			_ = rootService // read below by the thinking arm and the ladder
			report.StepsScanned++

			// A top-level step's config occupies the workflow_step rung at runtime;
			// a nested one arrives as the runtime StepConfig. Passing it at the rung
			// it actually occupies is what makes `from` below match what the pod logs.
			var workflowStepCfg, runtimeStepCfg map[string]interface{}
			if nested {
				runtimeStepCfg = stepCfg
			} else {
				workflowStepCfg = stepCfg
			}
			effective, from, _ := actions.ResolveStepBudget("max_tokens", root, workflowStepCfg, runtimeStepCfg)

			// Every declaration, in ladder order, with the winner marked.
			stepRung := "workflow_step"
			if nested {
				stepRung = "step_config"
			}
			var decls []budgetDeclaration
			for _, d := range []struct {
				cfg       map[string]interface{}
				level     string
				canonical bool
			}{
				{stepService, stepRung + ".ai_service", true},
				{stepCfg, stepRung, false},
				{rootService, "root.ai_service", true},
				{root, "root", false},
			} {
				if v, ok := numeric(d.cfg, "max_tokens"); ok {
					decls = append(decls, budgetDeclaration{
						Level: d.level, Value: v, Canonical: d.canonical,
						Effective: d.level == from,
					})
				}
			}
			report.Declarations += len(decls)

			finding := budgetFinding{
				Agent: agent.Type, Path: path, Effective: effective, From: from, Declarations: decls,
			}

			// budget_tokens BEFORE max_tokens, because this one is not a sizing
			// question at all — it is a request the provider refuses outright.
			if btValue, btFrom, _ := actions.ResolveStepBudget("budget_tokens", root, workflowStepCfg, runtimeStepCfg); btFrom != "" {
				model := resolveStepModel(stepService, rootService)
				if !aiservice.AcceptsThinkingBudget(model) {
					report.Findings = append(report.Findings, budgetFinding{
						Agent: agent.Type, Path: path, Kind: "thinking_unsupported",
						Effective: btValue, From: btFrom,
						Detail: fmt.Sprintf("declares budget_tokens=%d at %q against model %q, which REJECTS a manual "+
							"thinking budget with a 400. anthropic.go emits thinking:{type:enabled,budget_tokens:N} "+
							"whenever this key resolves to a positive number, so this fails EVERY call for this step, "+
							"not one. Anthropic replaced the fixed budget with adaptive thinking; the key is still "+
							"correct on 4.6 (deprecated) and REQUIRED on 4.5 and older, which is why the reader was "+
							"not simply changed to drop it (bugs_open/257)", btValue, btFrom, model),
					})
				}
			}

			switch {
			case from == "":
				finding.Kind = "unconfigured"
				finding.Detail = "declares an ai_service block and no max_tokens at any level — this step " +
					"runs at the provider floor (aiservice.DefaultMaxOutputTokens, 2048), the smallest number " +
					"in the estate, and the first oversized reply meets a silent cliff (bugs_open/205)"
				report.Findings = append(report.Findings, finding)

			case ambiguousLevel(decls) != "":
				finding.Kind = "ambiguous"
				finding.Detail = fmt.Sprintf("the %q level declares max_tokens in BOTH spellings with different "+
					"numbers. This is the one state in which the estate's two readers disagree: the canonical "+
					"ladder takes the ai_service one, llmOptionsFromConfig takes the bare one at the step level, "+
					"so which of your numbers is sent depends on which action the step runs. Delete one",
					ambiguousLevel(decls))
				report.Findings = append(report.Findings, finding)

			case !declaredCanonically(decls, from):
				finding.Kind = "non_canonical"
				finding.Detail = fmt.Sprintf("the effective budget is declared at %q, outside an ai_service block. "+
					"Honoured by the ladder, and advisory only — but the bare spelling is where both round-3 "+
					"failures started, and it sits beside model/provider keys that ARE read", from)
				report.Findings = append(report.Findings, finding)
			}
		})
	}

	sort.Slice(report.Findings, func(i, j int) bool {
		if report.Findings[i].Kind != report.Findings[j].Kind {
			return report.Findings[i].Kind < report.Findings[j].Kind
		}
		if report.Findings[i].Agent != report.Findings[j].Agent {
			return report.Findings[i].Agent < report.Findings[j].Agent
		}
		return report.Findings[i].Path < report.Findings[j].Path
	})
	return report
}

// resolveStepModel is the effective model for a step, by the same overlay order
// resolveAIServiceConfig uses: the step's ai_service block wins over the agent's
// root one. Returns "" when neither declares a model, which AcceptsThinkingBudget
// treats as unknown-and-permissive.
func resolveStepModel(stepService, rootService map[string]interface{}) string {
	for _, m := range []map[string]interface{}{stepService, rootService} {
		if m == nil {
			continue
		}
		if model, ok := m["model"].(string); ok && model != "" {
			return model
		}
	}
	return ""
}

// ambiguousLevel names the level, if any, that declares the budget in both
// spellings with different numbers — the only state in which the canonical ladder
// and the direct-caller ladder send different numbers for one step.
//
// Equal numbers are NOT ambiguous: one declaration written twice sends the same
// request whichever reader runs, so there is nothing for an operator to decide.
func ambiguousLevel(decls []budgetDeclaration) string {
	byLevel := map[string]budgetDeclaration{}
	for _, d := range decls {
		base := strings.TrimSuffix(d.Level, ".ai_service")
		if prev, seen := byLevel[base]; seen && prev.Value != d.Value {
			return base
		}
		byLevel[base] = d
	}
	return ""
}

func declaredCanonically(decls []budgetDeclaration, from string) bool {
	for _, d := range decls {
		if d.Level == from {
			return d.Canonical
		}
	}
	return true
}

func emitBudgetPlacement(args []string) {
	report := false
	for _, a := range args {
		switch a {
		case "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --budget-placement: unknown argument %q\n", a)
			os.Exit(2)
		}
	}

	var agents []liveAgent
	var failed int
	var err error
	if db, derr := dbConn(); derr != nil || db != nil {
		if derr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --budget-placement: %v\n", derr)
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDBWithQuery(db, "--budget-placement", budgetPlacementExportQuery)
	} else {
		var raw []byte
		raw, err = io.ReadAll(os.Stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --budget-placement: reading stdin: %v\n", err)
			os.Exit(2)
		}
		agents, failed, err = decodeLiveAgents(raw, "--budget-placement")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintln(os.Stderr, "config-key-audit --budget-placement: 0 live agents decoded — "+
			"refusing to print a clean report over an empty fleet.")
		os.Exit(2)
	}
	if err := requireAgentConfigProjection(agents); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	out := budgetPlacementFindings(agents)
	out.AgentsUndecoded = failed

	// Vacuity: a fleet of 208 agents that declares no budget anywhere means the key
	// has been renamed or the export lost a projection, not that all is well.
	if out.StepsScanned == 0 || out.Declarations == 0 {
		fmt.Fprintf(os.Stderr, "config-key-audit --budget-placement: %d steps scanned, %d declarations found "+
			"across %d agents — this report is watching nothing. The key has been renamed or the export is "+
			"missing a projection; repoint it rather than reading this as clean.\n",
			out.StepsScanned, out.Declarations, len(agents))
		os.Exit(2)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)

	if report {
		writeDocNote("budget-placement", budgetPlacementRunSummary(out), "config-integrity", "budget_placement_check")
	}

	// Only shadowed and unconfigured fail the run. non_canonical is honoured config
	// and must never turn a clean fleet red — an advisory that fails the exit code
	// stops being read as advisory within a week.
	// non_canonical is the ONLY advisory kind. thinking_unsupported is the most
	// urgent of the rest: it is not a sizing mistake, it is a request the provider
	// refuses, and it takes down every call for the step.
	fatal := 0
	for _, f := range out.Findings {
		if f.Kind != "non_canonical" {
			fatal++
		}
	}
	if fatal > 0 {
		fmt.Fprintf(os.Stderr, "%d step(s) send a token budget nobody chose, or one of two numbers an operator wrote at the same level.\n", fatal)
		os.Exit(1)
	}
}

// budgetPlacementRunSummary is the doc_notes body: prose stating the SCOPE as well
// as the result, because "0 findings over 3 agents" and "0 findings over 208" have
// opposite meanings.
func budgetPlacementRunSummary(r budgetPlacementReport) string {
	var b strings.Builder
	byKind := map[string]int{}
	for _, f := range r.Findings {
		byKind[f.Kind]++
	}
	fmt.Fprintf(&b, "budget-placement check: %d model steps across %d live agents, %d max_tokens declarations. ",
		r.StepsScanned, r.AgentsScanned, r.Declarations)
	if len(r.Findings) == 0 {
		b.WriteString("CLEAN: every step's budget is declared once, canonically, and takes effect.")
	} else {
		fmt.Fprintf(&b, "%d thinking_unsupported (budget_tokens against a model that 400s), %d unconfigured "+
			"(running at the 2048 floor), %d ambiguous (two spellings at one level, different numbers, the two "+
			"readers disagree), %d non-canonical (honoured, declared outside ai_service): ",
			byKind["thinking_unsupported"], byKind["unconfigured"], byKind["ambiguous"], byKind["non_canonical"])
		for i, f := range r.Findings {
			if i > 0 {
				b.WriteString("; ")
			}
			fmt.Fprintf(&b, "%s %s [%s] effective=%d from=%s", f.Agent, f.Path, f.Kind, f.Effective, f.From)
		}
		b.WriteString(".")
	}
	if r.AgentsUndecoded > 0 {
		fmt.Fprintf(&b, " %d agent row(s) failed to decode and were not scanned.", r.AgentsUndecoded)
	}
	return b.String()
}
