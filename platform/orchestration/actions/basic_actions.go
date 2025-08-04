// platform/orchestration/actions/basic_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ValidateInputAction validates the input data
func ValidateInputAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Check if we have input data
	if params.InputData == nil || len(params.InputData) == 0 {
		return nil, fmt.Errorf("no input data provided")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(params.InputData, &data); err != nil {
		return nil, fmt.Errorf("invalid input format: %w", err)
	}

	// Example validation: check for required fields
	if _, ok := data["message"]; !ok {
		return nil, fmt.Errorf("missing required field: message")
	}

	return map[string]interface{}{
		"validated": true,
		"input":     data,
	}, nil
}

// TransformDataAction transforms data (e.g., uppercase)
func TransformDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Get transformation config from step
	var transformation string
	if params.StepConfig.Config != nil {
		transformation, _ = params.StepConfig.Config["transformation"].(string)
	}

	// Get data to transform
	var data map[string]interface{}
	if params.InputData != nil {
		json.Unmarshal(params.InputData, &data)
	} else if len(params.CollectedData) > 0 {
		// Use collected data if no input data
		data = params.CollectedData
	}

	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = v
	}

	switch transformation {
	case "uppercase":
		// Transform any string values to uppercase
		for k, v := range data {
			if str, ok := v.(string); ok {
				result[k+"_transformed"] = strings.ToUpper(str)
			}
		}
	case "reverse":
		for k, v := range data {
			if str, ok := v.(string); ok {
				result[k+"_transformed"] = reverseString(str)
			}
		}
	default:
		return nil, fmt.Errorf("unknown transformation: %s", transformation)
	}

	return result, nil
}

// SendNotificationAction sends a message to the response topic
func SendNotificationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Prepare notification with all collected data
	notification := map[string]interface{}{
		"type":       "workflow_completed",
		"data":       params.CollectedData,
		"step":       params.CurrentStep,
		"agent_type": params.AgentType,
	}

	notificationBytes, _ := json.Marshal(notification)

	// Send to responses topic
	topic := fmt.Sprintf("system.responses.%s", params.AgentType)
	err := params.Producer.Produce(ctx, topic, params.Headers,
		[]byte(params.Headers["correlation_id"]), notificationBytes)

	if err != nil {
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	return map[string]interface{}{
		"notification_sent": true,
		"topic":             topic,
	}, nil
}

func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
