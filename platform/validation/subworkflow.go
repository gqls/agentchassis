// FILE: platform/validation/subworkflow.go
//
// bugs_open/144 — ValidateWorkflow ran once, on the top-level workflow. Steps
// nested inside a loop's sub-workflow are pulled straight out of config at
// execution time (platform/orchestration/actions/loop_actions.go:70-88), so every
// invariant this package enforces was unenforced for them. MEASURED 2026-07-29 over
// live agent_definitions (is_active, non-snapshot, not deleted): 85 nested steps
// across 18 agents, carrying 25 (action, config-key) pairs that exist ONLY nested.
//
// Why it stayed invisible: the offline audit (scripts/audit-config-keys.sh) read the
// same top-level-only path, so the two halves were blind in the same direction and
// agreed with each other. Consistent blindness reads exactly like correctness.
//
// TWO RULES SHAPE EVERYTHING BELOW.
//
//  1. VALIDATE WHAT RUNS, NOT WHAT IS WRITTEN. A nested step is decoded here by the
//     same function the executor uses — models.DecodeSubWorkflowStep — which reads
//     seven fields and drops the rest, so a nested `dependencies` or `sub_tasks` has
//     no effect whatever. Decoding by JSON round-trip into models.Step would populate
//     fields the executor never sees, and the validator would then vouch for
//     behaviour that does not happen. The dropped keys are REPORTED rather than
//     pretend-enforced.
//
//  2. A REFERENCE OUT OF A SUB-WORKFLOW IS LEGITIMATE, so an unresolved one WARNS
//     rather than fails. Loop expansion prefixes a next_step/error_step only when it
//     names a sibling substep; anything else passes through untouched
//     (coordinator.go:4002-4014, loop_expansion_handler.go:134-140) and is expected to
//     resolve against the enclosing plan — or against a step the expander itself
//     injects (`<loop>_complete`, loop_expansion_handler.go:192), which is not in the
//     definition at all and therefore cannot be checked here. A hard error on a
//     reference this file cannot see the target of would reject workflows that run.
//
// EVERY hard error below was measured against the live fleet before it shipped: 0 of
// the 85 nested steps fails any of them, and 0 of the 66 nested (action, key) pairs
// trips the strict-config rule. That measurement is re-runnable and ships with the
// change — platform/orchestration/actions/subworkflow_live_dryrun_test.go replays this
// validator over an export of the live definitions rather than over fixtures.
package validation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	"go.uber.org/zap"
)

// maxSubWorkflowDepth is a backstop against a pathological definition, not a design
// limit. Live maximum is 1. A definition nested deeper than this is rejected rather
// than partially validated, because "validated to depth 8" is the kind of qualified
// assurance that gets quoted without its qualifier.
const maxSubWorkflowDepth = 8

// The nested-step decode is models.DecodeSubWorkflowStep — the SAME function the loop
// action executes with, not a copy of it. It reads seven fields and reports the rest
// as dropped; see that file for why the set is not models.Step's json tags.
//
// This file used to hold a second implementation, pinned to the executor's by a
// lockstep test. The council's reuse seat objected that a test proving two copies
// agree is a backstop rather than single-sourcing, and that the founding incident
// (bugs_open/144: two hand-written traversals, blind in the same direction, agreeing
// with each other) is exactly what the backstop is supposed to prevent. It was right.

// WalkSteps calls fn for every step in a plan: each top-level step, then every step
// nested inside a loop sub-workflow, at any depth. `path` locates the step in the
// definition; `nested` is false only for the top level.
//
// Exported so that anything asking "which definitions carry action X?" or "which
// (action, key) pairs are live?" can ask it of the SAME step set the validator
// enforces against. That question used to be answered by a hand-written
// `->'workflow'->'steps'` query, which sees only the top level — the offline audit and
// the runtime validator were therefore blind in the same direction and agreed with
// each other (bugs_open/144). subWorkflowsOf is the single definition of where a
// nested step lives; this function and the validator both use it, so the two cannot
// disagree about what exists.
//
// Iteration order is sorted by step name, so a caller that prints results gets a
// stable diff between runs.
func WalkSteps(plan models.WorkflowPlan, fn func(path string, step models.Step, nested bool)) {
	names := make([]string, 0, len(plan.Steps))
	for name := range plan.Steps {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		step := plan.Steps[name]
		path := "steps." + name
		fn(path, step, false)
		walkNested(path, step, fn, 0)
	}
}

func walkNested(path string, step models.Step, fn func(string, models.Step, bool), depth int) {
	if depth >= maxSubWorkflowDepth {
		return
	}
	for _, sub := range subWorkflowsOf(path, step) {
		names := make([]string, 0, len(sub.Steps))
		for name := range sub.Steps {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			nestedStep := sub.Steps[name]
			nestedPath := sub.Path + "." + name
			fn(nestedPath, nestedStep, true)
			walkNested(nestedPath, nestedStep, fn, depth+1)
		}
	}
}

// subWorkflow is one sub-workflow definition found inside a step's config.
type subWorkflow struct {
	// Path locates it in the definition — "steps.build_pages_loop.sub_workflow".
	// Carried through every message because "step 'write_page' must have an action"
	// is unactionable in a definition that holds three steps of that name at
	// different depths.
	Path      string
	StartStep string
	Steps     map[string]models.Step
	// Dropped maps a nested step name to the config keys the runtime decoder ignores.
	Dropped map[string][]string
	// AlsoCarriedSubWorkflow records that the step carried BOTH shapes. The runtime
	// takes substeps and ignores the other, so the ignored half is inert config.
	AlsoCarriedSubWorkflow bool
}

// subWorkflowsOf returns the sub-workflow a step defines, or nothing.
//
// Precedence mirrors loop_actions.go:74-88 exactly: `substeps` wins, and
// `sub_workflow` is consulted only when substeps is absent or empty. Validating both
// when both are present would report a step that never runs.
func subWorkflowsOf(path string, step models.Step) []subWorkflow {
	if len(step.Config) == 0 {
		return nil
	}

	rawSubsteps, hasSubsteps := step.Config["substeps"].(map[string]interface{})
	rawSubWorkflow, hasSubWorkflow := step.Config["sub_workflow"].(map[string]interface{})

	if hasSubsteps && len(rawSubsteps) > 0 {
		sub := decodeSubWorkflow(path+".substeps", rawSubsteps, "")
		sub.AlsoCarriedSubWorkflow = hasSubWorkflow
		return []subWorkflow{sub}
	}

	if hasSubWorkflow {
		startStep, _ := rawSubWorkflow["start_step"].(string)
		rawSteps, _ := rawSubWorkflow["steps"].(map[string]interface{})
		return []subWorkflow{decodeSubWorkflow(path+".sub_workflow", rawSteps, startStep)}
	}

	// `substeps` present but empty is reported by the caller, not silently skipped:
	// the loop action fails at execution on exactly this.
	if hasSubsteps {
		return []subWorkflow{decodeSubWorkflow(path+".substeps", nil, "")}
	}

	return nil
}

func decodeSubWorkflow(path string, rawSteps map[string]interface{}, startStep string) subWorkflow {
	sub := subWorkflow{
		Path:      path,
		StartStep: startStep,
		Steps:     make(map[string]models.Step, len(rawSteps)),
		Dropped:   make(map[string][]string),
	}
	for name, raw := range rawSteps {
		rawStep, ok := raw.(map[string]interface{})
		if !ok {
			// Not a step definition at all. Recorded as a step with no action so the
			// caller's "must have an action" error names it, rather than skipping it
			// — skipping is what produced this bug in the first place.
			sub.Steps[name] = models.Step{}
			continue
		}
		step, dropped := models.DecodeSubWorkflowStep(rawStep)
		sub.Steps[name] = step
		if len(dropped) > 0 {
			sub.Dropped[name] = dropped
		}
	}
	return sub
}

// validateSubWorkflowsOf validates every sub-workflow nested inside one step, and
// recurses. `visible` is every step name in scope from the enclosing plans, used to
// decide whether a reference out of the sub-workflow can be resolved here.
func (v *WorkflowValidator) validateSubWorkflowsOf(path string, step models.Step, visible map[string]bool, depth int) error {
	subs := subWorkflowsOf(path, step)
	if len(subs) == 0 {
		return nil
	}
	if depth >= maxSubWorkflowDepth {
		return fmt.Errorf("sub-workflow nesting at '%s' exceeds the maximum supported depth of %d", path, maxSubWorkflowDepth)
	}
	for _, sub := range subs {
		if err := v.validateSubWorkflow(sub, visible, depth+1); err != nil {
			return err
		}
	}
	return nil
}

func (v *WorkflowValidator) validateSubWorkflow(sub subWorkflow, outer map[string]bool, depth int) error {
	if len(sub.Steps) == 0 {
		return fmt.Errorf("sub-workflow '%s' defines no steps — the loop action fails at execution with "+
			"\"requires 'substeps' or 'sub_workflow.steps'\"", sub.Path)
	}

	if sub.StartStep != "" {
		if _, exists := sub.Steps[sub.StartStep]; !exists {
			// The loop action warns and auto-detects a start step instead
			// (buildSubstepOrder). That tolerance is what makes this invisible: the
			// loop runs, in an order nobody chose.
			return fmt.Errorf("sub-workflow '%s' declares start_step '%s', which is not one of its steps", sub.Path, sub.StartStep)
		}
	}

	if sub.AlsoCarriedSubWorkflow {
		v.warnOnce("sub_workflow_ignored", sub.Path, "Step carries both 'substeps' and 'sub_workflow'; only 'substeps' is executed",
			zap.String("sub_workflow", sub.Path),
			zap.String("consequence", "the sub_workflow block is inert config that describes behaviour which does not happen"),
		)
	}

	visible := make(map[string]bool, len(outer)+len(sub.Steps))
	for name := range outer {
		visible[name] = true
	}
	for name := range sub.Steps {
		visible[name] = true
	}

	names := make([]string, 0, len(sub.Steps))
	for name := range sub.Steps {
		names = append(names, name)
	}
	// Deterministic order: map iteration would make the FIRST error reported for a
	// definition with two faults vary between pods, and a validation error that
	// changes identity between restarts is one nobody can dedupe or fix confidently.
	sort.Strings(names)

	for _, name := range names {
		if err := v.validateSubWorkflowStep(sub, name, sub.Steps[name], visible); err != nil {
			return err
		}
	}

	if err := v.checkForCycles(models.WorkflowPlan{StartStep: sub.StartStep, Steps: sub.Steps}); err != nil {
		return fmt.Errorf("sub-workflow '%s': %w", sub.Path, err)
	}

	for _, name := range names {
		if err := v.validateSubWorkflowsOf(sub.Path+"."+name, sub.Steps[name], visible, depth); err != nil {
			return err
		}
	}

	return nil
}

func (v *WorkflowValidator) validateSubWorkflowStep(sub subWorkflow, name string, step models.Step, visible map[string]bool) error {
	where := sub.Path + "." + name

	if step.Action == "" {
		return fmt.Errorf("nested step '%s' must have an action", where)
	}

	if !actioncheck.IsLocalAction(step.Action) && step.Topic == "" {
		return fmt.Errorf("nested step '%s' with action '%s' requires a topic", where, step.Action)
	}

	// fan_out cannot work nested, and failing here is the only place that says so.
	// The runtime decoder does not read sub_tasks, so a nested fan_out reaches the
	// executor with none — and the top-level rule ("must have at least one sub-task")
	// would pass on the definition while the executed step has zero.
	if step.Action == "fan_out" {
		return fmt.Errorf("nested step '%s' uses fan_out, which cannot work inside a sub-workflow: "+
			"the loop decoder does not read sub_tasks, so the step would execute with none", where)
	}

	if err := v.checkStepConfigKeys(where, step); err != nil {
		return err
	}

	if len(sub.Dropped[name]) > 0 {
		v.warnOnce("dropped_fields", where, "Nested step declares fields the loop decoder does not read — they are silently dropped before execution",
			zap.String("nested_step", where),
			zap.String("action", step.Action),
			zap.Strings("dropped_fields", sub.Dropped[name]),
			zap.String("consequence", "the definition describes behaviour that does not happen"),
			zap.String("honoured_fields", strings.Join(sortedKeys(models.SubWorkflowStepFields), ", ")),
			zap.String("ref", "bugs_open/144"),
		)
	}

	// Unresolved references WARN — see rule 2 in the file header. An external target
	// may be a step of an enclosing plan (visible here) or one the expander injects
	// at runtime (`<loop>_complete`, which is not in any definition), so this cannot
	// be decided with certainty at validation time.
	if step.NextStep != "" && !visible[step.NextStep] && !isInjectedLoopStep(step.NextStep) {
		v.warnOnce("dangling_next_step", where+"->"+step.NextStep,
			"Nested step's next_step names no step in scope",
			zap.String("nested_step", where),
			zap.String("next_step", step.NextStep),
			zap.String("note", "legitimate if it names a step injected at runtime; otherwise the iteration chain breaks at execution"),
		)
	}
	if step.ErrorStep != "" && !visible[step.ErrorStep] && !isInjectedLoopStep(step.ErrorStep) {
		v.warnOnce("dangling_error_step", where+"->"+step.ErrorStep,
			"Nested step's error_step names no step in scope",
			zap.String("nested_step", where),
			zap.String("error_step", step.ErrorStep),
			zap.String("note", "the coordinator routes to it and then fails with 'step not found'"),
		)
	}

	return nil
}

// isInjectedLoopStep reports whether a name can only be resolved after loop
// expansion. The expander creates "<loop>_complete" (loop_expansion_handler.go:192),
// which exists in no definition — so a definition referencing one is correct and
// unverifiable here. Deliberately a suffix test rather than a match against the loop
// names in scope: the reference may name a loop in a SIBLING branch, and a false
// "dangling" warning on a working workflow is worse than a missed one.
func isInjectedLoopStep(name string) bool {
	return strings.HasSuffix(name, "_complete")
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
