// FILE: platform/validation/workflow.go
package validation

import (
	"fmt"
	"github.com/gqls/agentchassis/pkg/models"
)

// WorkflowValidator provides validation for workflow plans
type WorkflowValidator struct {
	localActions   map[string]bool // Actions executed within orchestrator
	builtInActions map[string]bool // Built-in orchestration control actions
}

func NewWorkflowValidator() *WorkflowValidator {
	// Actions that execute custom code locally
	localActions := map[string]bool{
		"validate_input":    true,
		"transform_data":    true,
		"send_notification": true,
	}

	// Built-in orchestration control actions
	builtInActions := map[string]bool{
		"complete_workflow":     true,
		"fan_out":               true,
		"pause_for_human_input": true,
		"store_memory":          true,
		"retrieve_memory":       true,
	}

	return &WorkflowValidator{
		localActions:   localActions,
		builtInActions: builtInActions,
	}
}

// ValidateWorkflow validates a workflow configuration
func (v *WorkflowValidator) ValidateWorkflow(workflow models.WorkflowPlan) error {
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

	return nil
}

func (v *WorkflowValidator) validateStep(name string, step models.Step, allSteps map[string]models.Step) error {
	// Validate action
	if step.Action == "" {
		return fmt.Errorf("step '%s' must have an action", name)
	}

	// Check if it's a local action
	isLocalAction := v.localActions[step.Action]

	// Remote actions need a topic (unless they're local or built-in)
	if !isLocalAction && step.Topic == "" {
		// Special cases that don't need topics
		switch step.Action {
		case "complete_workflow", "fan_out", "pause_for_human_input":
			// These are OK without topics
		default:
			return fmt.Errorf("step '%s' with action '%s' requires a topic", name, step.Action)
		}
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

// IsLocalAction checks if an action is executed locally
func (v *WorkflowValidator) IsLocalAction(action string) bool {
	return v.localActions[action]
}

// RequiresTopic checks if action needs a Kafka topic
func (v *WorkflowValidator) RequiresTopic(action string) bool {
	// Local and built-in actions don't need topics
	return !v.IsLocalAction(action)
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
// calculateMaxDepth calculates the maximum depth of the workflow
func (v *WorkflowValidator) calculateMaxDepth(plan models.WorkflowPlan) int {
	depths := make(map[string]int)
	visiting := make(map[string]bool) // Add cycle detection

	var calculateDepth func(string) int
	calculateDepth = func(stepName string) int {
		// Check if already calculated
		if depth, ok := depths[stepName]; ok {
			return depth
		}

		// Check for cycles
		if visiting[stepName] {
			return 0 // Cycle detected, return 0 to avoid infinite recursion
		}

		step, ok := plan.Steps[stepName]
		if !ok {
			return 0
		}

		visiting[stepName] = true // Mark as visiting
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

		visiting[stepName] = false // Unmark
		depths[stepName] = maxDepth + 1
		return maxDepth + 1
	}

	return calculateDepth(plan.StartStep)
}
