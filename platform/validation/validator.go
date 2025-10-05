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

type RequiredFields struct {
	ClientID        string
	CorrelationID   string
	OrchestrationID string
	AgentType       string
	StepName        string
}

// NewValidator creates a new validator instance
func NewValidator(logger *zap.Logger) *Validator {
	return &Validator{
		workflowValidator: NewWorkflowValidator(logger),
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

// ValidateOutgoingMessage checks if all required fields are present
func (v *Validator) ValidateOutgoingMessage(headers map[string]string) bool {
	v.logger.Info("In ValidateOutgoingMessage",
		zap.Any("headers", headers),
	)

	stepName := headers["step_name"]
	if stepName == "" {
		stepName = headers["in_response_to_step_name"]
	}

	if headers["client_id"] == "" ||
		headers["correlation_id"] == "" ||
		headers["orchestration_id"] == "" ||
		headers["sender_agent_type"] == "" ||
		stepName == "" {

		v.logger.Error("Outgoing message missing required fields",
			zap.String("client_id", headers["client_id"]),
			zap.String("correlation_id", headers["correlation_id"]),
			zap.String("orchestration_id", headers["orchestration_id"]),
			zap.String("sender_agent_type", headers["sender_agent_type"]),
			zap.String("in_response_to_step_name", stepName))
		return false
	}
	return true
}

// ValidateIncomingMessage checks required fields for incoming messages
func (v *Validator) ValidateIncomingMessage(headers map[string]string) bool {
	// If it's an error message, let it through
	if headers["is_error"] == "true" {
		return true
	}

	if headers["client_id"] == "" ||
		headers["correlation_id"] == "" ||
		headers["orchestration_id"] == "" {

		v.logger.Warn("Incoming message missing required fields",
			zap.String("client_id", headers["client_id"]),
			zap.String("correlation_id", headers["correlation_id"]),
			zap.String("orchestration_id", headers["orchestration_id"]))
		return false
	}
	return true
}
