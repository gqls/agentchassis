// FILE: platform/orchestration/actions/hitl/hitl.go
// Package hitl provides human-in-the-loop approval actions
package hitl

import (
	"context"

	"github.com/aqls/personae/platform/orchestration/actions/registry"
)

func init() {
	registry.Register("await_approval", registry.ActionDefinition{
		Func:        AwaitApprovalAction,
		Category:    registry.CategoryHITL,
		Description: "Pauses workflow awaiting human approval",
		Status:      registry.StatusActive,
	})

	registry.Register("process_approval_decision", registry.ActionDefinition{
		Func:        ProcessApprovalDecisionAction,
		Category:    registry.CategoryHITL,
		Description: "Processes an approval decision",
		Status:      registry.StatusActive,
	})

	registry.Register("process_data", registry.ActionDefinition{
		Func:        ProcessApprovalDecisionAction, // Alias
		Category:    registry.CategoryHITL,
		Description: "Alias for process_approval_decision",
		Status:      registry.StatusDeprecated,
	})

	registry.Register("create_approval_request", registry.ActionDefinition{
		Func:        CreateApprovalRequestAction,
		Category:    registry.CategoryHITL,
		Description: "Creates an approval request for human review",
		Status:      registry.StatusActive,
	})

	registry.Register("wait_for_approval_response", registry.ActionDefinition{
		Func:        WaitForApprovalResponseAction,
		Category:    registry.CategoryHITL,
		Description: "Waits for approval response",
		Status:      registry.StatusActive,
	})
}

// TODO: Migrate implementations from hitl_actions.go

func AwaitApprovalAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	return map[string]interface{}{"status": "awaiting_approval"}, nil
}

func ProcessApprovalDecisionAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	return map[string]interface{}{"status": "processed"}, nil
}

func CreateApprovalRequestAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	return map[string]interface{}{"status": "created"}, nil
}

func WaitForApprovalResponseAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	return map[string]interface{}{"status": "waiting"}, nil
}
