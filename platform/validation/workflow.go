// FILE: platform/validation/workflow.go
package validation

import (
	"fmt"
	"github.com/gqls/agentchassis/pkg/models"
)

// WorkflowValidator provides validation for workflow plans
type WorkflowValidator struct{}

// NewWorkflowValidator creates a new workflow validator
func NewWorkflowValidator() *WorkflowValidator {
	return &WorkflowValidator{}
}

// ValidateWorkflowPlan validates a workflow plan for correctness
func (v *WorkflowValidator) ValidateWorkflowPlan(plan models.WorkflowPlan) error {
	if plan.StartStep == "" {
		return fmt.Errorf("workflow plan must have a start step")
	}

	if len(plan.Steps) == 0 {
		return fmt.Errorf("workflow plan must have at least one step")
	}

	// Check start step exists
	if _, ok := plan.Steps[plan.StartStep]; !ok {
		return fmt.Errorf("start step '%s' not found in steps", plan.StartStep)
	}

	// Validate each step
	for stepName, step := range plan.Steps {
		if err := v.validateStep(stepName, step, plan); err != nil {
			return err
		}
	}

	// Check for cycles
	if err := v.checkForCycles(plan); err != nil {
		return err
	}

	// Check all dependencies exist
	if err := v.validateDependencies(plan); err != nil {
		return err
	}

	return nil
}

// validateStep validates an individual step
func (v *WorkflowValidator) validateStep(name string, step models.Step, plan models.WorkflowPlan) error {
	if step.Action == "" {
		return fmt.Errorf("step '%s' must have an action", name)
	}

	// Validate based on action type
	switch step.Action {
	case "fan_out":
		if len(step.SubTasks) == 0 {
			return fmt.Errorf("fan_out step '%s' must have at least one sub-task", name)
		}
		for i, subTask := range step.SubTasks {
			if subTask.StepName == "" {
				return fmt.Errorf("sub-task %d in step '%s' must have a step name", i, name)
			}
			if subTask.Topic == "" {
				return fmt.Errorf("sub-task %d in step '%s' must have a topic", i, name)
			}
		}
	case "complete_workflow":
		if step.NextStep != "" {
			return fmt.Errorf("complete_workflow step '%s' should not have a next step", name)
		}
	default:
		// For standard actions, ensure topic is set if not internal actions
		if step.Topic == "" && !v.isInternalAction(step.Action) {
			return fmt.Errorf("step '%s' with action '%s' requires a topic", name, step.Action)
		}
	}

	// Validate next step exists
	if step.NextStep != "" {
		if _, ok := plan.Steps[step.NextStep]; !ok {
			return fmt.Errorf("step '%s' references non-existent next step '%s'", name, step.NextStep)
		}
	}

	return nil
}

// isInternalAction checks if an action is handled internally
func (v *WorkflowValidator) isInternalAction(action string) bool {
	internalActions := map[string]bool{
		"complete_workflow":     true,
		"pause_for_human_input": true,
		"store_memory":          true,
		"retrieve_memory":       true,
	}
	return internalActions[action]
}

// validateDependencies ensures all dependencies exist
func (v *WorkflowValidator) validateDependencies(plan models.WorkflowPlan) error {
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

	var hasCycle func(string) bool
	hasCycle = func(stepName string) bool {
		visited[stepName] = true
		recStack[stepName] = true

		step, ok := plan.Steps[stepName]
		if !ok {
			return false
		}

		// Check next step
		if step.NextStep != "" {
			if !visited[step.NextStep] {
				if hasCycle(step.NextStep) {
					return true
				}
			} else if recStack[step.NextStep] {
				return true
			}
		}

		// Check dependencies
		for _, dep := range step.Dependencies {
			if !visited[dep] {
				if hasCycle(dep) {
					return true
				}
			} else if recStack[dep] {
				return true
			}
		}

		recStack[stepName] = false
		return false
	}

	// Start from the start step
	if hasCycle(plan.StartStep) {
		return fmt.Errorf("workflow contains a cycle")
	}

	// Check any unvisited steps (disconnected components)
	for stepName := range plan.Steps {
		if !visited[stepName] {
			if hasCycle(stepName) {
				return fmt.Errorf("workflow contains a cycle")
			}
		}
	}

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

	var calculateDepth func(string) int
	calculateDepth = func(stepName string) int {
		if depth, ok := depths[stepName]; ok {
			return depth
		}

		step, ok := plan.Steps[stepName]
		if !ok {
			return 0
		}

		maxDepth := 0

		// Check dependencies
		for _, dep := range step.Dependencies {
			depDepth := calculateDepth(dep)
			if depDepth > maxDepth {
				maxDepth = depDepth
			}
		}

		// Check next step
		if step.NextStep != "" {
			nextDepth := calculateDepth(step.NextStep)
			if nextDepth > maxDepth {
				maxDepth = nextDepth
			}
		}

		depths[stepName] = maxDepth + 1
		return maxDepth + 1
	}

	return calculateDepth(plan.StartStep)
}
