// FILE: platform/validation/workflow.go
package validation

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// WorkflowValidator provides validation for workflow plans
type WorkflowValidator struct {
	logger *zap.Logger

	// warnedConfigKeys dedupes the unknown-config-key warning, which would
	// otherwise fire on every message for the life of the pod.
	warnedConfigKeys sync.Map
}

func NewWorkflowValidator(logger *zap.Logger) *WorkflowValidator {
	return &WorkflowValidator{
		logger: logger.Named("workflow_validator"),
	}
}

// ValidateWorkflow validates a workflow configuration
func (v *WorkflowValidator) ValidateWorkflow(workflow models.WorkflowPlan) error {
	v.logger.Info(" in ValidateWorkflow workflow.go",
		zap.Any("workflow", workflow),
	)

	if workflow.StartStep == "" {
		return fmt.Errorf("workflow must have a start_step")
	}

	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// Check if start step exists
	if _, exists := workflow.Steps[workflow.StartStep]; !exists {
		return fmt.Errorf("start_step '%s' not found in steps", workflow.StartStep)
	}

	// Validate each step
	for stepName, step := range workflow.Steps {
		if err := v.validateStep(stepName, step, workflow.Steps); err != nil {
			return err
		}
	}

	// Check for cycles
	if err := v.checkForCycles(workflow); err != nil {
		return err
	}

	// Check all dependencies exist (though validateStep already does this)
	if err := v.validateDependencies(workflow); err != nil {
		return err
	}

	// Validate the steps nested inside loop sub-workflows (bugs_open/144).
	//
	// Until 2026-07-29 validation stopped here, at the top level. A nested step is
	// extracted from config and executed directly by the loop action, so it reached
	// production having been checked by nothing at all — 85 such steps across 18
	// live agents. Everything specific to that traversal is in subworkflow.go; the
	// visible set starts as the top-level step names because a reference out of a
	// sub-workflow resolves against the enclosing plan.
	visible := make(map[string]bool, len(workflow.Steps))
	for stepName := range workflow.Steps {
		visible[stepName] = true
	}
	for stepName, step := range workflow.Steps {
		if err := v.validateSubWorkflowsOf("steps."+stepName, step, visible, 0); err != nil {
			return err
		}
	}

	return nil
}

func (v *WorkflowValidator) validateStep(name string, step models.Step, allSteps map[string]models.Step) error {
	// Validate action
	if step.Action == "" {
		return fmt.Errorf("step '%s' must have an action", name)
	}

	// Use the global registry to check if it's a local action
	isLocal := actioncheck.IsLocalAction(step.Action)

	v.logger.Info("Validating step, workflow.go",
		zap.String("step", name),
		zap.String("action", step.Action),
		zap.Bool("is_local", isLocal),
		zap.Bool("has_topic", step.Topic != ""),
	)

	// Remote actions need a topic unless they're local
	if !isLocal && step.Topic == "" {
		return fmt.Errorf("step '%s' with action '%s' requires a topic", name, step.Action)
	}

	// Validate next step exists (unless it's empty, which means end of workflow)
	if step.NextStep != "" {
		if _, exists := allSteps[step.NextStep]; !exists {
			return fmt.Errorf("step '%s' references non-existent next_step '%s'", name, step.NextStep)
		}
	}

	// Validate dependencies exist
	for _, dep := range step.Dependencies {
		if _, exists := allSteps[dep]; !exists {
			return fmt.Errorf("step '%s' has dependency on non-existent step '%s'", name, dep)
		}
	}

	if err := v.checkStepConfigKeys(name, step); err != nil {
		return err
	}

	// Validate fan-out sub-tasks
	if step.Action == "fan_out" {
		if len(step.SubTasks) == 0 {
			return fmt.Errorf("fan_out step '%s' must have at least one sub-task", name)
		}
		for _, subTask := range step.SubTasks {
			if subTask.Topic == "" {
				return fmt.Errorf("fan_out sub-task '%s' must have a topic", subTask.StepName)
			}
		}
	}

	return nil
}

// checkStepConfigKeys reports step-config keys the action does not recognise
// (bugs_open/101).
//
// An unknown key is silently ignored at execution, so a stale or aspirational key is
// indistinguishable BY INSPECTION from a live one — the config reads as a
// specification of behaviour while being evidence of nothing. This is the only place
// that sees the action name and its config together on every run.
//
// Warn, don't fail, unless the action has declared its key set complete
// (StrictConfig). ValidateWorkflow rejecting a workflow fleet-wide on a key list that
// is merely incomplete would be a far worse defect than the one being reported.
//
// `where` is the step name at the top level and the full path inside a sub-workflow,
// so the same rule reads the same way at both depths. Shared with the nested
// validator deliberately: two copies of this rule is how the top level and the nested
// level came to disagree in the first place (bugs_open/144).
func (v *WorkflowValidator) checkStepConfigKeys(where string, step models.Step) error {
	if len(step.Config) == 0 {
		return nil
	}
	unknown, checked := datahelpers.UnknownConfigKeys(step.Action, step.Config)
	if !checked || len(unknown) == 0 {
		return nil
	}
	if datahelpers.IsStrictConfigAction(step.Action) {
		return fmt.Errorf("step '%s' (action '%s') has unrecognised config keys %v — "+
			"this action declares its config contract as complete, so an unknown key is "+
			"a definition error, not a no-op", where, step.Action, unknown)
	}
	v.warnUnknownConfigKeysOnce(where, step.Action, unknown)
	return nil
}

// warnUnknownConfigKeysOnce logs each distinct (step, action, keys) combination a
// single time per process. ValidateWorkflow runs on EVERY message, so an undeduped
// warning would repeat for the life of the pod on a busy lane — and a warning that
// scrolls is a warning nobody reads, which would leave the defect just as invisible
// as the silent ignore it replaced.
func (v *WorkflowValidator) warnUnknownConfigKeysOnce(stepName, action string, unknown []string) {
	key := stepName + "\x00" + action + "\x00" + strings.Join(unknown, ",")
	if _, seen := v.warnedConfigKeys.LoadOrStore(key, struct{}{}); seen {
		return
	}
	v.logger.Warn("Step config carries keys this action does not read — they are silently ignored at execution",
		zap.String("step", stepName),
		zap.String("action", action),
		zap.Strings("unrecognised_keys", unknown),
		zap.String("consequence", "the config describes behaviour that does not happen"),
		zap.String("fix", "implement the key, delete it from the definition, or add it to the action's ActionInputSpec.ConfigKeys if it is read elsewhere"),
		zap.String("ref", "bugs_open/101"),
	)
}

// warnOnce is the same dedupe applied to the sub-workflow warnings, for the same
// reason: ValidateWorkflow runs on every message, and a warning that repeats for the
// life of the pod is one nobody reads. `kind` namespaces the key so two different
// findings about the same step cannot silence each other.
func (v *WorkflowValidator) warnOnce(kind, key, msg string, fields ...zap.Field) {
	if _, seen := v.warnedConfigKeys.LoadOrStore(kind+"\x00"+key, struct{}{}); seen {
		return
	}
	v.logger.Warn(msg, fields...)
}

// IsLocalAction checks if an action is executed locally
func (v *WorkflowValidator) IsLocalAction(action string) bool {

	v.logger.Debug(" in IsLocalAction workflow.go",
		zap.Any("action", action),
	)

	return actioncheck.IsLocalAction(action)
}

// RequiresTopic checks if action needs a Kafka topic
func (v *WorkflowValidator) RequiresTopic(action string) bool {
	// Local and built-in actions don't need topics
	v.logger.Debug(" in RequiresTopic workflow.go",
		zap.Any("action", action),
	)

	return actioncheck.IsLocalAction(action)
}

// validateDependencies ensures all dependencies exist
func (v *WorkflowValidator) validateDependencies(plan models.WorkflowPlan) error {
	v.logger.Debug(" in ValidateDependencies workflow.go",
		zap.Any("plan", plan),
	)

	for stepName, step := range plan.Steps {
		for _, dep := range step.Dependencies {
			if _, ok := plan.Steps[dep]; !ok {
				return fmt.Errorf("step '%s' has dependency on non-existent step '%s'", stepName, dep)
			}
		}
	}
	return nil
}

// checkForCycles detects cycles in the workflow
func (v *WorkflowValidator) checkForCycles(plan models.WorkflowPlan) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := []string{} // Track the path for debugging

	var hasCycle func(string) bool
	hasCycle = func(stepName string) bool {
		visited[stepName] = true
		recStack[stepName] = true
		path = append(path, stepName) // Add to path

		// Log current path
		v.logger.Debug("Checking step for cycle",
			zap.String("step_name", stepName),
			zap.Strings("path", path),
		)

		step, ok := plan.Steps[stepName]
		if !ok {
			path = path[:len(path)-1] // Remove from path
			recStack[stepName] = false
			return false
		}

		// Check next step
		if step.NextStep != "" {
			v.logger.Debug("Following next_step",
				zap.String("from", stepName),
				zap.String("to next_step", step.NextStep),
			)
			if !visited[step.NextStep] {
				if hasCycle(step.NextStep) {
					return true
				}
			} else if recStack[step.NextStep] {
				v.logger.Error("CYCLE DETECTED in workflow via next_step",
					zap.String("from", stepName),
					zap.String("to", step.NextStep),
					zap.Strings("cycle_path", append(path, step.NextStep)),
				)
				return true
			}
		}

		// Check dependencies
		for _, dep := range step.Dependencies {
			v.logger.Debug(" Step depends on",
				zap.String("Step", stepName),
				zap.String("depends on", dep),
			)
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				v.logger.Error("CYCLE DETECTED in workflow - step dependencies",
					zap.String("from", stepName),
					zap.String("depends on", dep),
					zap.Strings("cycle_path", append(path, dep)),
				)
				return true
			}
		}

		path = path[:len(path)-1] // Remove from path when backtracking
		recStack[stepName] = false
		return false
	}

	// Start from the start step
	v.logger.Debug("Starting cycle check",
		zap.String("start_step", plan.StartStep),
	)

	if hasCycle(plan.StartStep) {
		return fmt.Errorf("workflow contains a cycle")
	}

	// Check any unvisited steps (disconnected components)
	for stepName := range plan.Steps {
		if !visited[stepName] {
			// Was fmt.Printf to stdout. Left alone it would now fire for every step of
			// every sub-workflow on every message — a sub-workflow usually has no
			// start_step, so ALL of its steps take this branch (bugs_open/144).
			// Unstructured stdout from a hot path is not a log, it is noise; nothing
			// greps for it (checked across the tree).
			v.logger.Debug("Checking disconnected step for cycles", zap.String("step_name", stepName))
			if hasCycle(stepName) {
				return fmt.Errorf("workflow contains a cycle")
			}
		}
	}

	v.logger.Debug("No cycles detected")
	return nil
}

// GetWorkflowMetrics calculates metrics about a workflow
func (v *WorkflowValidator) GetWorkflowMetrics(plan models.WorkflowPlan) map[string]interface{} {
	metrics := map[string]interface{}{
		"total_steps":    len(plan.Steps),
		"fan_out_steps":  0,
		"external_calls": 0,
		"max_depth":      0,
	}

	for _, step := range plan.Steps {
		if step.Action == "fan_out" {
			metrics["fan_out_steps"] = metrics["fan_out_steps"].(int) + 1
		}
		if step.Topic != "" {
			metrics["external_calls"] = metrics["external_calls"].(int) + 1
		}
	}

	// Calculate max depth
	metrics["max_depth"] = v.calculateMaxDepth(plan)

	return metrics
}

// calculateMaxDepth calculates the maximum depth of the workflow
func (v *WorkflowValidator) calculateMaxDepth(plan models.WorkflowPlan) int {
	depths := make(map[string]int)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var calculateDepth func(string) int
	calculateDepth = func(stepName string) int {
		// If already calculated, return cached result
		if depth, ok := depths[stepName]; ok {
			return depth
		}

		// Cycle detection - if we're already calculating this step
		if recStack[stepName] {
			return 0 // Return 0 to break the cycle
		}

		step, ok := plan.Steps[stepName]
		if !ok {
			return 0
		}

		// Mark as being calculated
		visited[stepName] = true
		recStack[stepName] = true

		maxDepth := 0

		// Check dependencies
		for _, dep := range step.Dependencies {
			depDepth := calculateDepth(dep)
			if depDepth > maxDepth {
				maxDepth = depDepth
			}
		}

		// Check next step
		if step.NextStep != "" && step.NextStep != stepName {
			nextDepth := calculateDepth(step.NextStep)
			if nextDepth > maxDepth {
				maxDepth = nextDepth
			}
		}

		// Unmark from recursion stack
		recStack[stepName] = false

		// Cache the result
		depths[stepName] = maxDepth + 1

		return maxDepth + 1
	}

	return calculateDepth(plan.StartStep)
}
