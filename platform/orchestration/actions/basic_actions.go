// platform/orchestration/actions/basic_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ValidateInputAction validates the input data
func ValidateInputAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Get input_data from CollectedData
	inputData, ok := params.CollectedData["input_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no input_data in CollectedData")
	}

	// Validate specific fields within input_data
	// Example: if the input should have a "message" field
	if _, ok := inputData["message"]; !ok {
		// Check if it might be nested differently
		if data, ok := inputData["data"].(map[string]interface{}); ok {
			if _, ok := data["message"]; !ok {
				return nil, fmt.Errorf("missing required field: message")
			}
		}
	}

	return map[string]interface{}{
		"validated": true,
		"input":     inputData,
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

	// Get data from input_data
	data, ok := params.CollectedData["input_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no input_data to transform")
	}

	// Check if we have collected data from previous steps
	if validatedData, ok := params.CollectedData["validate_input"]; ok {
		// Use the validated data from the previous step
		if validated, ok := validatedData.(map[string]interface{}); ok {
			if input, ok := validated["input"].(map[string]interface{}); ok {
				data = input
			}
		}
	}

	// If no collected data, try to parse from input
	if data == nil && params.InputData != nil {
		var payload map[string]interface{}
		if err := json.Unmarshal(params.InputData, &payload); err == nil {
			if d, ok := payload["data"].(map[string]interface{}); ok {
				data = d
			}
		}
	}

	if data == nil {
		return nil, fmt.Errorf("no data to transform")
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
		"type":      "workflow_completed",
		"data":      params.CollectedData,
		"step":      params.CurrentStep,
		"timestamp": time.Now().UTC(),
	}

	notificationBytes, _ := json.Marshal(notification)

	// Use a fixed topic or get from step config
	topic := "system.agent.generic.responses" // Fixed topic for generic agent

	// Or get from step config if you want it configurable
	if params.StepConfig.Config != nil {
		if customTopic, ok := params.StepConfig.Config["topic"].(string); ok {
			topic = customTopic
		}
	}

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
