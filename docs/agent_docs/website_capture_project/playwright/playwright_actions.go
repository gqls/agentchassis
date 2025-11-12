// in playwright_actions.go
package playwright

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
	// ... other imports
)

// This tells the orchestrator to wait for the response
type CaptureSiteResult struct {
	Success       bool   `json:"success"`
	RequestID     string `json:"request_id"`
	TopicSentTo   string `json:"topic_sent_to"`
	AwaitResponse bool   `json:"await_response"` // This will be true
}

func CaptureSiteAction(ctx context.Context, params actions.ActionParams) (interface{}, error) {
	params.Logger.Info("Executing CaptureSiteAction", zap.String("step_name", params.ExecutionContext.StepName))

	// 1. Get data from the workflow
	// The workflow JSON will pass the URL in the config
	urlToCapture, _ := params.StepConfig.Config["url"].(string)
	if urlToCapture == "" {
		return nil, fmt.Errorf("url is required for CaptureSiteAction")
	}

	// 2. Define our topics
	adapterRequestTopic := "system.adapter.playwright.requests"
	// The orchestrator will listen on this dynamic topic for the response
	myResponsesTopic := params.ExecutionContext.ResponsesTopic

	// 3. Build the message body for the Python adapter
	requestBody := map[string]interface{}{
		"action": "capture_site",
		"data": map[string]interface{}{
			"url":            urlToCapture,
			"s3_bucket":      "my-backblaze-bucket", // Or get from env/config
			"s3_path_prefix": fmt.Sprintf("captures/%s", params.ExecutionContext.CorrelationID),
		},
		"reply_to_topic": myResponsesTopic,
	}

	// 4. Build the full Kafka message (just like in image_actions.go)
	newRequestID := uuid.NewString()
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":    params.ExecutionContext.CorrelationID,
			"orchestration_id":  params.ExecutionContext.OrchestrationID,
			"step_name":         params.ExecutionContext.StepName,
			"request_id":        newRequestID,
			"message_type":      "request",
			"responses_topic":   myResponsesTopic,
			"sender_agent_type": params.ExecutionContext.Sender.AgentType,
			// ... all other required headers
		},
		"body": requestBody,
	}

	// 5. Send the message
	headers := datahelpers.BuildKafkaHeaders(adapterRequest["headers"]) // Assuming you have a helper
	messageBytes, _ := json.Marshal(adapterRequest)

	if err := params.Producer.ProduceWithValidation(
		ctx,
		adapterRequestTopic,
		headers,
		[]byte(params.ExecutionContext.CorrelationID),
		messageBytes,
	); err != nil {
		return nil, fmt.Errorf("failed to send to playwright adapter: %w", err)
	}

	// 6. Tell the orchestrator to wait
	result := CaptureSiteResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   adapterRequestTopic,
		AwaitResponse: true,
	}

	return result, nil
}
