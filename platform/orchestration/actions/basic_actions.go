// platform/orchestration/actions/basic_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ValidateInputAction validates the input data
func ValidateInputAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Get input_data from CollectedData
	var inputData map[string]interface{}

	// Check if input_data is a step with a response
	if stepData, ok := params.CollectedData["input_data"]; ok {
		extractedData := datahelpers.ExtractStepData(stepData)
		inputData, ok = extractedData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no valid input_data in CollectedData")
		}
	} else {
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
	params.Logger.Info("In TransformDataAction ")

	// Get transformation config from step
	var transformation string
	if params.StepConfig.Config != nil {
		transformation, _ = params.StepConfig.Config["transformation"].(string)
	}

	// Get data to transform
	var data map[string]interface{}

	// First, check if we have data from a previous validation step
	if validatedStepData, ok := params.CollectedData["validate_input"]; ok {
		// Extract data from the step (checking for response field)
		extractedData := datahelpers.ExtractStepData(validatedStepData)

		// If it's a validated response, extract the input field
		if validated, ok := extractedData.(map[string]interface{}); ok {
			if input, ok := validated["input"].(map[string]interface{}); ok {
				data = input
			} else {
				// Maybe the whole extracted data is what we need
				data = validated
			}
		}
	}

	// If no validated data, try input_data
	if data == nil {
		if inputStepData, ok := params.CollectedData["input_data"]; ok {
			extractedData := datahelpers.ExtractStepData(inputStepData)
			if inputMap, ok := extractedData.(map[string]interface{}); ok {
				data = inputMap
			}
		}
	}

	// If still no data, try to parse from raw InputData
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

	params.Logger.Info("In TransformDataAction at the end",
		zap.String("transformation", transformation),
		zap.Any("the result of the transformation is:", result),
	)

	return result, nil
}

// SendNotificationAction sends a message to the response topic
func SendNotificationAction(ctx context.Context, params ActionParams) (interface{}, error) {
	// Prepare collected data with responses extracted
	collectedResults := make(map[string]interface{})

	// Extract responses from all steps
	for key, value := range params.CollectedData {
		// Skip internal fields
		if strings.HasPrefix(key, "__") {
			continue
		}
		// Extract data from step (checking for response field)
		collectedResults[key] = datahelpers.ExtractStepData(value)
	}

	// Prepare notification
	notification := map[string]interface{}{
		"type":      "workflow_completed",
		"data":      collectedResults,
		"step":      params.CurrentStep,
		"timestamp": time.Now().UTC(),
	}

	notificationBytes, _ := json.Marshal(notification)

	parentsResponsesTopic := os.Getenv("PARENT_RESPONSES_TOPIC")
	// Use ResponsesTopic from ExecutionContext
	topic := parentsResponsesTopic

	// Allow override from step config if needed
	if params.StepConfig.Config != nil {
		if customTopic, ok := params.StepConfig.Config["topic"].(string); ok {
			topic = customTopic
		}
	}

	// No fallback - must have a topic
	if topic == "" {
		return nil, fmt.Errorf("no response topic configured for notification")
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
