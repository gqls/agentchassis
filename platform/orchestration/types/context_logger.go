package types

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogContext creates structured log fields from ExecutionContext
func (ec *ExecutionContext) LogContext() []zapcore.Field {
	fields := []zapcore.Field{
		zap.String("correlation_id", ec.CorrelationID),
		zap.String("orchestration_id", ec.OrchestrationID),
		zap.String("orchestration_name", ec.OrchestrationName),

		zap.String("parent_orchestration_id", ec.ParentOrchestrationID),
		zap.String("parent_orchestration_name", ec.ParentOrchestrationName),
		zap.String("parent_request_id", ec.ParentRequestID),
		zap.String("ec.StepID", ec.StepID),
		zap.String("ec.StepName", ec.StepName),
		zap.String("ec.Action", ec.Action),
		zap.Any("ec.InResponseTo", ec.InResponseTo),
		zap.String("ec.Status", ec.Status),
		zap.Any("Sender", ec.Sender),

		zap.String("message_id", ec.MessageID),
		zap.String("request_id", ec.RequestID),
		zap.String("message_type", ec.MessageType),
		zap.String("client_id", ec.ClientID),
		zap.Int("fuel_budget", ec.FuelBudget),
		zap.Time("timestamp", ec.Timestamp),
	}

	// Add optional fields only if present
	if ec.ParentOrchestrationID != "" {
		fields = append(fields, zap.String("parent_orchestration_id", ec.ParentOrchestrationID))
	}
	if ec.InResponseTo != nil {
		if ec.InResponseTo.RequestID != "" {
			fields = append(fields, zap.String("in_response_to_request_id", ec.InResponseTo.RequestID))
		}
		if ec.InResponseTo.StepID != "" {
			fields = append(fields, zap.String("in_response_to_step_id", ec.InResponseTo.StepID))
		}
	}
	if ec.GroupID != "" {
		fields = append(fields, zap.String("group_id", ec.GroupID))
	}
	if ec.FromAgentID != "" {
		fields = append(fields, zap.String("from_agent_id", ec.FromAgentID))
	}
	if ec.ToAgentID != "" {
		fields = append(fields, zap.String("to_agent_id", ec.ToAgentID))
	}

	return fields
}

// LogCompact creates a compact set of critical fields for less verbose logging
func (ec *ExecutionContext) LogCompact() []zapcore.Field {
	return []zapcore.Field{
		zap.String("corr_id", ec.CorrelationID[:8]), // First 8 chars
		zap.String("orch_id", ec.OrchestrationID[:8]),
		zap.String("msg_type", ec.MessageType),
		zap.String("req_id", ec.RequestID[:8]),
	}
}
