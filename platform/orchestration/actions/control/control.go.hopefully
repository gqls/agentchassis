// FILE: platform/orchestration/actions/control/control.go
// Package control provides workflow control actions: complete, await, branch, route
package control

import (
	"context"
	"fmt"

	"github.com/aqls/personae/platform/orchestration/actions/registry"
)

func init() {
	// Register all control actions
	registry.Register("complete_workflow", registry.ActionDefinition{
		Func:        CompleteWorkflowAction,
		Category:    registry.CategoryControl,
		Description: "Marks the current workflow as complete and returns collected data",
		Status:      registry.StatusActive,
	})

	registry.Register("await_response", registry.ActionDefinition{
		Func:        AwaitResponseAction,
		Category:    registry.CategoryControl,
		Description: "Pauses workflow execution waiting for an external response",
		Status:      registry.StatusActive,
	})

	registry.Register("conditional_branch", registry.ActionDefinition{
		Func:        ConditionalBranchAction,
		Category:    registry.CategoryControl,
		Description: "Branches workflow based on a condition evaluation",
		Status:      registry.StatusActive,
	})

	registry.Register("conditional_route", registry.ActionDefinition{
		Func:        ConditionalRouteAction,
		Category:    registry.CategoryControl,
		Description: "Routes to different steps based on conditions",
		Status:      registry.StatusActive,
	})
}

// CompleteWorkflowAction marks the workflow as complete
func CompleteWorkflowAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	params.Logger.Info("Completing workflow")// Add structured logging

	// Return collected data as the result
	result := make(map[string]interface{})
	result["status"] = "completed"
	result["step"] = params.StepName

	// Include any output field data if configured
	if outputField, ok := params.Config["output_field"].(string); ok {
		if data, exists := params.CollectedData[outputField]; exists {
			result["output"] = data
		}
	}

	return result, nil
}

// AwaitResponseAction pauses for external input
func AwaitResponseAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	// Extract timeout from config
	timeoutSeconds := 30
	if ts, ok := params.Config["timeout_seconds"].(float64); ok {
		timeoutSeconds = int(ts)
	}

	params.Logger.Info("Awaiting response")// zap.Int("timeout_seconds", timeoutSeconds),

	return map[string]interface{}{
		"status":          "awaiting",
		"timeout_seconds": timeoutSeconds,
	}, nil
}

// ConditionalBranchAction evaluates a condition and returns the branch to take
func ConditionalBranchAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	condition, ok := params.Config["condition"].(string)
	if !ok {
		return nil, fmt.Errorf("conditional_branch requires 'condition' in config")
	}

	trueBranch, _ := params.Config["true_branch"].(string)
	falseBranch, _ := params.Config["false_branch"].(string)

	// Evaluate the condition (simplified - real implementation would parse expressions)
	result := evaluateCondition(condition, params.CollectedData)

	branch := falseBranch
	if result {
		branch = trueBranch
	}

	return map[string]interface{}{
		"condition_result": result,
		"next_step":        branch,
	}, nil
}

// ConditionalRouteAction routes based on multiple conditions
func ConditionalRouteAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	routes, ok := params.Config["routes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("conditional_route requires 'routes' array in config")
	}

	defaultRoute, _ := params.Config["default"].(string)

	for _, r := range routes {
		route, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		condition, _ := route["condition"].(string)
		target, _ := route["target"].(string)

		if evaluateCondition(condition, params.CollectedData) {
			return map[string]interface{}{
				"next_step": target,
				"matched":   condition,
			}, nil
		}
	}

	return map[string]interface{}{
		"next_step": defaultRoute,
		"matched":   "default",
	}, nil
}

// evaluateCondition is a simple condition evaluator
// In production, this would use a proper expression parser
func evaluateCondition(condition string, data map[string]interface{}) bool {
	// Simplified: just check if a field exists and is truthy
	// Real implementation would parse expressions like "data.count > 5"
	if val, exists := data[condition]; exists {
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return v != ""
		case int, int64, float64:
			return v != 0
		default:
			return val != nil
		}
	}
	return false
}
