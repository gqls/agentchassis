// platform/orchestration/actions/hitl_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// AwaitApprovalAction pauses the workflow and sends a notification for human approval
func AwaitApprovalAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("AwaitApprovalAction: Starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	// Generate unique approval token
	approvalToken := uuid.New().String()

	// Extract data to be approved from CollectedData
	dataForApproval := extractDataForApproval(params.CollectedData, params.StepConfig.Config, params.Logger)

	// Get notification topic from config or use default
	notificationTopic := "system.notifications.ui"
	if topic, ok := params.StepConfig.Config["notification_topic"].(string); ok {
		notificationTopic = topic
	}

	// Build approval request notification
	notification := buildApprovalNotification(
		params.ExecutionContext,
		approvalToken,
		dataForApproval,
		params.StepConfig.Config,
	)

	// Send notification to HITL service
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal approval notification: %w", err)
	}

	headers := map[string]string{
		"correlation_id": params.ExecutionContext.CorrelationID,
		"request_id":     approvalToken,
		"message_type":   "notification",
		"action":         "approval_required",
	}

	key := []byte(params.ExecutionContext.CorrelationID)

	err = params.Producer.Produce(ctx, notificationTopic, headers, key, notificationBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send approval notification: %w", err)
	}

	params.Logger.Info("AwaitApprovalAction: Sent approval request",
		zap.String("approval_token", approvalToken),
		zap.String("notification_topic", notificationTopic),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// Store approval request in database if DB is available
	if params.DB != nil {
		err = storeApprovalRequest(ctx, params.DB, approvalToken, params.ExecutionContext, dataForApproval, params.Logger)
		if err != nil {
			params.Logger.Error("Failed to store approval request in DB", zap.Error(err))
			// Continue anyway - notification was sent
		}
	}

	// Return with AwaitResponse flag to pause the workflow
	return map[string]interface{}{
		"approval_token":    approvalToken,
		"status":            "awaiting_approval",
		"message":           "Workflow paused for human approval",
		"await_response":    true, // This tells SagaCoordinator to pause
		"request_id":        approvalToken,
		"reply_to_topic":    "system.commands.workflow.resume",
		"data_for_approval": dataForApproval,
	}, nil
}

// ProcessApprovalDecisionAction processes the approval/rejection response from human
func ProcessApprovalDecisionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ProcessApprovalDecisionAction: Starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// Extract approval response from CollectedData
	approvalResponse := extractApprovalResponse(params.CollectedData, params.Logger)

	if approvalResponse == nil {
		return nil, fmt.Errorf("no approval response found in collected data")
	}

	approved, _ := approvalResponse["approved"].(bool)
	comments, _ := approvalResponse["comments"].(string)
	approvedBy, _ := approvalResponse["approved_by"].(string)

	params.Logger.Info("ProcessApprovalDecisionAction: Processing decision",
		zap.Bool("approved", approved),
		zap.String("comments", comments),
		zap.String("approved_by", approvedBy),
	)

	// Update approval request in database if available
	if params.DB != nil && params.ExecutionContext.RequestID != "" {
		err := updateApprovalRequest(ctx, params.DB, params.ExecutionContext.RequestID, approved, comments, approvedBy, params.Logger)
		if err != nil {
			params.Logger.Error("Failed to update approval request in DB", zap.Error(err))
		}
	}

	// Determine next action based on approval
	result := map[string]interface{}{
		"approved":    approved,
		"comments":    comments,
		"approved_by": approvedBy,
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
	}

	if !approved {
		result["status"] = "rejected"
		result["message"] = "Request was rejected by human approval"

		// If rejection should stop the workflow
		if stopOnReject, ok := params.StepConfig.Config["stop_on_reject"].(bool); ok && stopOnReject {
			result["stop_workflow"] = true
		}
	} else {
		result["status"] = "approved"
		result["message"] = "Request was approved"
	}

	// Merge any data modifications from the approval
	if modifiedData, ok := approvalResponse["modified_data"].(map[string]interface{}); ok {
		result["modified_data"] = modifiedData
	}

	return result, nil
}

// CreateApprovalRequestAction creates an approval request record
func CreateApprovalRequestAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CreateApprovalRequestAction: Starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// Generate approval request ID
	requestID := uuid.New().String()

	// Extract input data
	inputData := datahelpers.GetInputData(params.CollectedData, params.Logger)

	// Build approval request
	approvalRequest := map[string]interface{}{
		"request_id":       requestID,
		"orchestration_id": params.ExecutionContext.OrchestrationID,
		"correlation_id":   params.ExecutionContext.CorrelationID,
		"agent_type":       params.ExecutionContext.Sender.AgentType,
		"agent_id":         params.ExecutionContext.Sender.AgentID,
		"step_name":        params.ExecutionContext.StepName,
		"data":             inputData,
		"created_at":       time.Now().UTC().Format(time.RFC3339),
		"status":           "pending",
	}

	// Add any metadata from config
	if metadata, ok := params.StepConfig.Config["metadata"].(map[string]interface{}); ok {
		approvalRequest["metadata"] = metadata
	}

	// Store in database if available
	if params.DB != nil {
		err := storeApprovalRequest(ctx, params.DB, requestID, params.ExecutionContext, inputData, params.Logger)
		if err != nil {
			return nil, fmt.Errorf("failed to store approval request: %w", err)
		}
		params.Logger.Info("CreateApprovalRequestAction: Stored in database",
			zap.String("request_id", requestID),
		)
	}

	return approvalRequest, nil
}

// WaitForApprovalResponseAction checks for approval response
// This is similar to AwaitApprovalAction but without sending notification
func WaitForApprovalResponseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("WaitForApprovalResponseAction: Starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// Check if we have a request_id to wait for
	var requestID string
	if rid, ok := params.CollectedData["approval_token"].(string); ok {
		requestID = rid
	} else if rid, ok := params.CollectedData["request_id"].(string); ok {
		requestID = rid
	} else {
		// Generate new request ID
		requestID = uuid.New().String()
	}

	params.Logger.Info("WaitForApprovalResponseAction: Waiting for response",
		zap.String("request_id", requestID),
	)

	// Check database for existing approval if available
	if params.DB != nil {
		approval, err := getApprovalStatus(ctx, params.DB, requestID, params.Logger)
		if err == nil && approval != nil {
			status, _ := approval["status"].(string)
			if status == "approved" || status == "rejected" {
				params.Logger.Info("WaitForApprovalResponseAction: Found existing decision",
					zap.String("status", status),
				)
				return approval, nil
			}
		}
	}

	// Return with await flag to pause workflow
	return map[string]interface{}{
		"request_id":     requestID,
		"status":         "waiting",
		"message":        "Waiting for approval response",
		"await_response": true,
		"reply_to_topic": "system.commands.workflow.resume",
	}, nil
}

// Helper functions

func extractDataForApproval(collectedData map[string]interface{}, config map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Get the data fields to include in approval
	includeFields, _ := config["approval_fields"].([]string)

	if len(includeFields) == 0 {
		// If no specific fields, get input_data
		return datahelpers.GetInputData(collectedData, logger)
	}

	result := make(map[string]interface{})
	for _, field := range includeFields {
		if value, ok := collectedData[field]; ok {
			result[field] = value
		}
	}

	return result
}

func buildApprovalNotification(
	execCtx *types.ExecutionContext,
	approvalToken string,
	dataForApproval map[string]interface{},
	config map[string]interface{},
) map[string]interface{} {

	notification := map[string]interface{}{
		"type":             "approval_request",
		"request_id":       approvalToken,
		"orchestration_id": execCtx.OrchestrationID,
		"correlation_id":   execCtx.CorrelationID,
		"agent_type":       execCtx.Sender.AgentType,
		"agent_id":         execCtx.Sender.AgentID,
		"step_name":        execCtx.StepName,
		"reply_to_topic":   "system.commands.workflow.resume",
		"data":             dataForApproval,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	// Add approval type if specified
	if approvalType, ok := config["approval_type"].(string); ok {
		notification["approval_type"] = approvalType
	}

	// Add timeout if specified
	if timeout, ok := config["timeout_seconds"].(int); ok {
		notification["timeout_seconds"] = timeout
	}

	// Add any UI hints
	if uiConfig, ok := config["ui_config"].(map[string]interface{}); ok {
		notification["ui_config"] = uiConfig
	}

	return notification
}

func extractApprovalResponse(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Check for approval response in various locations
	if response, ok := collectedData["approval_response"].(map[string]interface{}); ok {
		return response
	}

	if response, ok := collectedData["await_approval"].(map[string]interface{}); ok {
		if body, ok := response["body"].(map[string]interface{}); ok {
			return body
		}
	}

	// Check in input_data
	inputData := datahelpers.GetInputData(collectedData, logger)
	if _, exists := inputData["approved"]; exists {
		return inputData
	}

	return nil
}

// Database helper functions (stubs - implement based on your schema)

func storeApprovalRequest(ctx context.Context, db interface{}, requestID string, execCtx *types.ExecutionContext, data map[string]interface{}, logger *zap.Logger) error {
	// Implementation depends on your database structure
	// This is a stub that should be implemented based on your schema
	logger.Info("Storing approval request (stub)",
		zap.String("request_id", requestID),
	)
	return nil
}

func updateApprovalRequest(ctx context.Context, db interface{}, requestID string, approved bool, comments string, approvedBy string, logger *zap.Logger) error {
	// Implementation depends on your database structure
	logger.Info("Updating approval request (stub)",
		zap.String("request_id", requestID),
		zap.Bool("approved", approved),
	)
	return nil
}

func getApprovalStatus(ctx context.Context, db interface{}, requestID string, logger *zap.Logger) (map[string]interface{}, error) {
	// Implementation depends on your database structure
	logger.Info("Getting approval status (stub)",
		zap.String("request_id", requestID),
	)
	return nil, fmt.Errorf("not implemented")
}
