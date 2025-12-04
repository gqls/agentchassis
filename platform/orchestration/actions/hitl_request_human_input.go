// platform/orchestration/actions/request_human_input.go
// Extended HITL actions supporting questionnaires, confirmations, and skip conditions.
// This complements hitl_actions.go with additional request types while following
// the same patterns and conventions.
package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// RequestHumanInputAction pauses the workflow and sends a notification for human input.
// This extends AwaitApprovalAction with support for:
// - Multiple request types: confirmation, questionnaire, review
// - Skip conditions for automated workflows
// - Field defaults populated from collected data
func RequestHumanInputAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RequestHumanInputAction: Starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	config := params.StepConfig.Config

	// Check skip condition first (extension over AwaitApprovalAction)
	if shouldSkip, reason := checkInputSkipCondition(config, params.CollectedData, params.Logger); shouldSkip {
		params.Logger.Info("RequestHumanInputAction: Skipping due to condition",
			zap.String("reason", reason),
		)

		// Return defaults as if user confirmed them
		result := map[string]interface{}{
			"skipped":     true,
			"skip_reason": reason,
			"status":      "auto_confirmed",
			"message":     "HITL skipped - using defaults from classification",
		}

		// Populate field defaults as the "response"
		if fields, ok := config["fields"].([]interface{}); ok {
			defaults := extractFieldDefaults(fields, params.CollectedData, params.Logger)

			// Log the populated defaults for traceability
			defaultKeys := make([]string, 0, len(defaults))
			for k := range defaults {
				defaultKeys = append(defaultKeys, k)
			}
			params.Logger.Info("RequestHumanInputAction: Populated defaults for skipped HITL",
				zap.Strings("populated_fields", defaultKeys),
				zap.Any("default_values", defaults),
			)

			for k, v := range defaults {
				result[k] = v
			}

			// If no defaults were populated, log a warning
			if len(defaults) == 0 {
				params.Logger.Warn("RequestHumanInputAction: No defaults could be populated - downstream steps may fail",
					zap.Int("field_count", len(fields)),
				)
			}
		}

		return result, nil
	}

	// Generate unique request token (same pattern as AwaitApprovalAction)
	requestToken := uuid.New().String()

	// Get request type (extension: multiple types)
	requestType := "confirmation"
	if rt, ok := config["request_type"].(string); ok {
		requestType = rt
	}

	// Extract data for the request
	dataForInput := extractDataForInput(params.CollectedData, config, params.Logger)

	// Get notification topic from config or use default
	notificationTopic := "system.notifications.ui"
	if topic, ok := config["notification_topic"].(string); ok {
		notificationTopic = topic
	}

	// Build input request notification
	notification := buildInputNotification(
		params.ExecutionContext,
		requestToken,
		requestType,
		dataForInput,
		config,
		params.CollectedData,
		*params.Logger,
	)

	// Send notification to HITL service
	notificationBytes, err := json.Marshal(notification)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input notification: %w", err)
	}

	headers := map[string]string{
		"correlation_id": params.ExecutionContext.CorrelationID,
		"request_id":     requestToken,
		"message_type":   "notification",
		"action":         "input_required",
		"request_type":   requestType,
	}

	key := []byte(params.ExecutionContext.CorrelationID)

	err = params.Producer.Produce(ctx, notificationTopic, headers, key, notificationBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to send input notification: %w", err)
	}

	params.Logger.Info("RequestHumanInputAction: Sent input request",
		zap.String("request_token", requestToken),
		zap.String("request_type", requestType),
		zap.String("notification_topic", notificationTopic),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.String("correlation_id", params.ExecutionContext.CorrelationID),
	)

	// Store request in database if DB is available
	if params.DB != nil {
		err = storeInputRequest(ctx, params.DB, requestToken, requestType, params.ExecutionContext, dataForInput, params.Logger)
		if err != nil {
			params.Logger.Error("Failed to store input request in DB", zap.Error(err))
			// Continue anyway - notification was sent
		}
	}

	// Return with AwaitResponse flag to pause the workflow (same pattern as AwaitApprovalAction)
	return map[string]interface{}{
		"request_token":  requestToken,
		"request_id":     requestToken, // Alias for compatibility
		"request_type":   requestType,
		"status":         "awaiting_input",
		"message":        "Workflow paused for human input",
		"await_response": true, // This tells SagaCoordinator to pause
		"reply_to_topic": params.ExecutionContext.ResponsesTopic,
		"data_for_input": dataForInput,
	}, nil
}

// ProcessHumanInputResponseAction processes the response from human input.
// This mirrors ProcessApprovalDecisionAction for input requests.
func ProcessHumanInputResponseAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ProcessHumanInputResponseAction: Starting",
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.Any("DEBUG: CollectedData", params.CollectedData),
	)

	// Extract input response from CollectedData
	inputResponse := extractInputResponse(params.CollectedData, params.Logger)

	if inputResponse == nil {
		return nil, fmt.Errorf("no input response found in collected data")
	}

	// Check status
	status, _ := inputResponse["status"].(string)
	if status == "" {
		if cancelled, ok := inputResponse["cancelled"].(bool); ok && cancelled {
			status = "cancelled"
		} else {
			status = "completed"
		}
	}

	respondedBy, _ := inputResponse["responded_by"].(string)

	params.Logger.Info("ProcessHumanInputResponseAction: Processing response",
		zap.String("status", status),
		zap.String("responded_by", respondedBy),
		zap.Any("input_response", inputResponse),
	)

	// Update request in database if available
	if params.DB != nil && params.ExecutionContext.RequestID != "" {
		err := updateInputRequest(ctx, params.DB, params.ExecutionContext.RequestID, status, inputResponse, params.Logger)
		if err != nil {
			params.Logger.Error("Failed to update input request in DB", zap.Error(err))
		}
	}

	// Build result
	result := map[string]interface{}{
		"status":       status,
		"responded_by": respondedBy,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	// Copy response data to result
	if data, ok := inputResponse["data"].(map[string]interface{}); ok {
		for k, v := range data {
			result[k] = v
		}
	} else {
		// Response might be flat - copy non-meta fields
		for k, v := range inputResponse {
			if k != "status" && k != "request_id" && k != "request_token" && k != "responded_by" {
				result[k] = v
			}
		}
	}

	// Handle cancellation
	if status == "cancelled" || status == "timeout" {
		result["message"] = fmt.Sprintf("Human input %s", status)

		// If cancellation should stop the workflow
		if stopOnCancel, ok := params.StepConfig.Config["stop_on_cancel"].(bool); ok && stopOnCancel {
			result["stop_workflow"] = true
		}
	} else {
		result["message"] = "Input received"
	}

	return result, nil
}

// Helper functions

func checkInputSkipCondition(config map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) (bool, string) {
	skipIf, ok := config["skip_if"].(string)
	if !ok || skipIf == "" {
		return false, ""
	}

	// Parse condition: "field == value" or "field != value"
	var fieldPath, operator, expectedValue string

	if idx := strings.Index(skipIf, "=="); idx > 0 {
		fieldPath = strings.TrimSpace(skipIf[:idx])
		expectedValue = strings.TrimSpace(skipIf[idx+2:])
		operator = "=="
	} else if idx := strings.Index(skipIf, "!="); idx > 0 {
		fieldPath = strings.TrimSpace(skipIf[:idx])
		expectedValue = strings.TrimSpace(skipIf[idx+2:])
		operator = "!="
	} else {
		logger.Warn("Invalid skip_if condition format", zap.String("skip_if", skipIf))
		return false, ""
	}

	// Get actual value using dot notation (with JSON string support)
	actualValue, err := getNestedFieldValue(collectedData, fieldPath)
	if err != nil {
		logger.Debug("Skip condition field not found",
			zap.String("field_path", fieldPath),
			zap.Error(err),
		)
		return false, ""
	}

	actualStr := fmt.Sprintf("%v", actualValue)

	switch operator {
	case "==":
		if actualStr == expectedValue {
			return true, fmt.Sprintf("%s == %s", fieldPath, expectedValue)
		}
	case "!=":
		if actualStr != expectedValue {
			return true, fmt.Sprintf("%s != %s", fieldPath, expectedValue)
		}
	}

	return false, ""
}

func extractDataForInput(collectedData map[string]interface{}, config map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	// Get the data fields to include (same pattern as extractDataForApproval)
	includeFields, _ := config["input_fields"].([]string)

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

func buildInputNotification(
	execCtx *types.ExecutionContext,
	requestToken string,
	requestType string,
	dataForInput map[string]interface{},
	config map[string]interface{},
	collectedData map[string]interface{},
	logger zap.Logger,
) map[string]interface{} {

	replyTopic := execCtx.ResponsesTopic
	if replyTopic == "" {
		replyTopic = os.Getenv("RESPONSES_TOPIC")
		if replyTopic == "" {
			err := errors.New("missing required env var RESPONSES_TOPIC")
			logger.Error("reply topic missing in buildInputNotification", zap.Error(err))
			replyTopic = "system.generic.responses"
		}
	}

	notification := map[string]interface{}{
		"type":             "input_request",
		"request_type":     requestType,
		"request_id":       requestToken,
		"orchestration_id": execCtx.OrchestrationID,
		"correlation_id":   execCtx.CorrelationID,
		"agent_type":       execCtx.Sender.AgentType,
		"agent_id":         execCtx.Sender.AgentID,
		"step_name":        execCtx.StepName,
		"reply_to_topic":   replyTopic,
		"data":             dataForInput,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}

	// Add title and message if specified
	if title, ok := config["title"].(string); ok {
		notification["title"] = title
	}
	if message, ok := config["message"].(string); ok {
		notification["message"] = message
	}

	// Add timeout if specified
	if timeout, ok := config["timeout_seconds"].(int); ok {
		notification["timeout_seconds"] = timeout
	} else if timeout, ok := config["timeout_seconds"].(float64); ok {
		notification["timeout_seconds"] = int(timeout)
	}

	// Handle request type specific fields
	switch requestType {
	case "questionnaire", "confirmation":
		if fields, ok := config["fields"].([]interface{}); ok {
			// Populate default values from collected data
			populatedFields := populateInputFieldDefaults(fields, collectedData, &logger)
			notification["fields"] = populatedFields
		}

	case "review":
		if editable, ok := config["editable"].(bool); ok {
			notification["editable"] = editable
		} else {
			notification["editable"] = true
		}
		if dataField, ok := config["data_field"].(string); ok {
			if data, err := getNestedFieldValue(collectedData, dataField); err == nil {
				notification["data_to_review"] = data
			}
		}
	}

	// Add any UI hints (same pattern as buildApprovalNotification)
	if uiConfig, ok := config["ui_config"].(map[string]interface{}); ok {
		notification["ui_config"] = uiConfig
	}

	return notification
}

func extractInputResponse(collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	logger.Info("In extractInputResponse")

	// Check for input response in various locations (same pattern as extractApprovalResponse)
	if response, ok := collectedData["input_response"].(map[string]interface{}); ok {
		logger.Info("Found input data in 'input_response' key")
		return response
	}

	if response, ok := collectedData["human_input_response"].(map[string]interface{}); ok {
		logger.Info("Found input data in 'human_input_response' key")
		return response
	}

	if response, ok := collectedData["request_human_input"].(map[string]interface{}); ok {
		logger.Info("Found 'request_human_input' key", zap.Any("data", response))
		// Check for a nested 'body'
		if body, ok := response["body"].(map[string]interface{}); ok {
			logger.Info("Found nested 'body' key")
			return body
		}
		return response
	}

	// Check in input_data (same pattern as extractApprovalResponse)
	inputData := datahelpers.GetInputData(collectedData, logger)
	if len(inputData) > 0 {
		// Check if it looks like a response
		if _, hasData := inputData["data"]; hasData {
			logger.Info("Found input response in 'input_data' key")
			return inputData
		}
		if _, hasStatus := inputData["status"]; hasStatus {
			logger.Info("Found input response in 'input_data' key")
			return inputData
		}
	}

	logger.Warn("Could not find input response in expected locations")
	return nil
}

func populateInputFieldDefaults(fields []interface{}, collectedData map[string]interface{}, logger *zap.Logger) []interface{} {
	result := make([]interface{}, len(fields))

	for i, field := range fields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			result[i] = field
			continue
		}

		// Copy the field
		newField := make(map[string]interface{})
		for k, v := range fieldMap {
			newField[k] = v
		}

		// If there's a default_from, populate the default value
		if defaultFrom, ok := fieldMap["default_from"].(string); ok {
			if value, err := getNestedFieldValue(collectedData, defaultFrom); err == nil {
				newField["default"] = value
				logger.Debug("Populated field default",
					zap.Any("field", fieldMap["name"]),
					zap.String("from", defaultFrom),
					zap.Any("value", value),
				)
			} else {
				logger.Debug("Could not populate field default",
					zap.Any("field", fieldMap["name"]),
					zap.String("from", defaultFrom),
					zap.Error(err),
				)
			}
		}

		result[i] = newField
	}

	return result
}

func extractFieldDefaults(fields []interface{}, collectedData map[string]interface{}, logger *zap.Logger) map[string]interface{} {
	result := make(map[string]interface{})

	for _, field := range fields {
		fieldMap, ok := field.(map[string]interface{})
		if !ok {
			continue
		}

		fieldName, _ := fieldMap["name"].(string)
		if fieldName == "" {
			continue
		}

		// Try default_from first
		if defaultFrom, ok := fieldMap["default_from"].(string); ok {
			logger.Debug("Attempting to extract default",
				zap.String("field", fieldName),
				zap.String("default_from", defaultFrom),
			)

			if value, err := getNestedFieldValue(collectedData, defaultFrom); err == nil {
				result[fieldName] = value
				logger.Info("Successfully extracted default for field",
					zap.String("field", fieldName),
					zap.String("from", defaultFrom),
					zap.Any("value", value),
				)
				continue
			} else {
				logger.Warn("Failed to extract default for field",
					zap.String("field", fieldName),
					zap.String("from", defaultFrom),
					zap.Error(err),
				)
			}
		}

		// Fall back to static default
		if defaultVal, ok := fieldMap["default"]; ok {
			result[fieldName] = defaultVal
			logger.Debug("Using static default for field",
				zap.String("field", fieldName),
				zap.Any("value", defaultVal),
			)
		}
	}

	return result
}

// cleanMarkdownJSON removes markdown code fences from JSON strings
// Handles cases where there's extra text after the closing ```
func cleanMarkdownJSON(s string) string {
	s = strings.TrimSpace(s)

	// Check for ```json or ``` prefix
	hasJSONFence := strings.HasPrefix(s, "```json")
	hasPlainFence := strings.HasPrefix(s, "```")

	if hasJSONFence {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimSpace(s)
	} else if hasPlainFence {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSpace(s)
	}

	// If we removed a prefix fence, look for the closing fence
	// It might not be at the very end (LLM might add extra text after)
	if hasJSONFence || hasPlainFence {
		if idx := strings.Index(s, "```"); idx >= 0 {
			s = s[:idx]
			s = strings.TrimSpace(s)
		}
	} else if strings.HasSuffix(s, "```") {
		// No opening fence but has closing - just trim it
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}

	return s
}

// tryParseJSONString attempts to parse a string as JSON (including markdown-wrapped JSON)
// Returns the parsed map and true if successful, nil and false otherwise
func tryParseJSONString(s string) (map[string]interface{}, bool) {
	// Clean markdown fences if present
	cleaned := cleanMarkdownJSON(s)

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(cleaned), &result); err == nil {
		return result, true
	}

	return nil, false
}

func getNestedFieldValue(data map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return nil, fmt.Errorf("key '%s' not found at position %d", part, i)
			}
			current = val

		case string:
			// Try to parse the string as JSON (handles LLM outputs with markdown-wrapped JSON)
			if parsed, ok := tryParseJSONString(v); ok {
				val, exists := parsed[part]
				if !exists {
					return nil, fmt.Errorf("key '%s' not found in parsed JSON at position %d", part, i)
				}
				current = val
			} else {
				return nil, fmt.Errorf("cannot navigate into string at '%s' (position %d) - not valid JSON", part, i)
			}

		default:
			return nil, fmt.Errorf("cannot navigate into type %T at '%s'", current, part)
		}
	}

	return current, nil
}

// Database helper functions (stubs - same pattern as hitl_actions.go)

func storeInputRequest(ctx context.Context, db interface{}, requestID string, requestType string, execCtx *types.ExecutionContext, data map[string]interface{}, logger *zap.Logger) error {
	// Implementation depends on your database structure
	// This is a stub that should be implemented based on your schema
	logger.Info("Storing input request (stub)",
		zap.String("request_id", requestID),
		zap.String("request_type", requestType),
	)
	return nil
}

func updateInputRequest(ctx context.Context, db interface{}, requestID string, status string, response map[string]interface{}, logger *zap.Logger) error {
	// Implementation depends on your database structure
	logger.Info("Updating input request (stub)",
		zap.String("request_id", requestID),
		zap.String("status", status),
	)
	return nil
}
