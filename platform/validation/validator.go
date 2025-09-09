// FILE: platform/validation/validator.go
package validation

import (
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

// Validator wraps workflow validation and other validation logic
type Validator struct {
	workflowValidator *WorkflowValidator
	logger            *zap.Logger
}

// NewValidator creates a new validator instance
func NewValidator(logger *zap.Logger) *Validator {
	return &Validator{
		workflowValidator: NewWorkflowValidator(),
		logger:            logger,
	}
}

// ValidateWorkflow validates a workflow plan
func (v *Validator) ValidateWorkflow(workflow models.WorkflowPlan) error {
	return v.workflowValidator.ValidateWorkflow(workflow)
}

// IsLocalAction checks if an action is executed locally
func (v *Validator) IsLocalAction(action string) bool {
	return v.workflowValidator.IsLocalAction(action)
}
