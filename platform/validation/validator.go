// FILE: platform/validation/validator.go
package validation

import (
	"context"
	"fmt"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/kafka"
	"go.uber.org/zap"
)

// Validator wraps workflow validation and other validation logic
type Validator struct {
	workflowValidator *WorkflowValidator
	kafkaProducer     *kafka.KafkaProducer
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
	if headers["client_id"] == "" ||
		headers["correlation_id"] == "" ||
		headers["orchestration_id"] == "" ||
		headers["from_agent_type"] == "" ||
		headers["step_name"] == "" {

		v.logger.Error("Outgoing message missing required fields",
			zap.String("client_id", headers["client_id"]),
			zap.String("correlation_id", headers["correlation_id"]),
			zap.String("orchestration_id", headers["orchestration_id"]),
			zap.String("from_agent_type", headers["from_agent_type"]),
			zap.String("step_name", headers["step_name"]))
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

func (v Validator) ProduceWithValidation(ctx context.Context, topic string, headers map[string]string, key, value []byte) error {
	// Check if it's an error message - always send those
	if headers["is_error"] == "true" {
		v.logger.Info("Sending error message without validation",
			zap.String("topic", topic),
			zap.String("correlation_id", headers["correlation_id"]))
		return v.kafkaProducer.Produce(ctx, topic, headers, key, value)
	}

	// Validate required fields
	validator := NewValidator(v.logger)
	if !validator.ValidateOutgoingMessage(headers) {
		v.logger.Error("Message validation failed, not sending",
			zap.String("topic", topic))
		return fmt.Errorf("message missing required fields")
	}

	// Send the message
	return v.kafkaProducer.Produce(ctx, topic, headers, key, value)
}
