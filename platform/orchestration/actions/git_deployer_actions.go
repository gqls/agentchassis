// internal/backend/agent-chassis/platform/orchestration/actions/deployer_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GitCommitResult represents the result of a Git commit operation.
// (This matches the structure of your WebscrapeResult)
type GitCommitResult struct {
	Success       bool                   `json:"success"`
	RequestID     string                 `json:"request_id"`
	TopicSentTo   string                 `json:"topic_sent_to"`
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GitCommitAction sends a commit request to the configured git-adapter,
func GitCommitAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing GitCommitAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// 1. Extract configuration
	config := params.StepConfig.Config
	action := "commit" // This is our specific action

	// 2. Get client_id and response topic
	clientID := params.ExecutionContext.ClientID
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		params.Logger.Warn("ResponsesTopic not set in ExecutionContext, responses may be lost",
			zap.String("step_name", params.ExecutionContext.StepName),
		)
		// Define a default or just log, depending on your system's strictness
	}

	// 3. Get adapter topic
	adapterTopic := "system.adapter.gitcommit.requests"

	// 4. Extract specific parameters from the config
	repoName, _ := config["repo_name"].(string)
	commitMessage, _ := config["commit_message"].(string)

	// Convert 'files' map from map[string]interface{} to map[string]string
	filesRaw, _ := config["files"].(map[string]interface{})
	filesMap := make(map[string]string)
	for filename, content := range filesRaw {
		if contentStr, ok := content.(string); ok {
			filesMap[filename] = contentStr
		} else {
			params.Logger.Warn("Skipping non-string file content in commit",
				zap.String("file", filename),
				zap.Any("content_type", fmt.Sprintf("%T", content)),
			)
		}
	}

	// 5. Build the specific 'data' payload for the adapter's body
	gitData := map[string]interface{}{
		"repo_name":      repoName,
		"files":          filesMap,
		"commit_message": commitMessage,
	}

	// 6. Generate a unique request ID for this specific adapter call
	newRequestID := uuid.New().String()

	// 7. Build the full adapterRequest, matching your example
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			// Core message identification
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",

			// Sender information
			"sender_agent_type":    params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":      params.ExecutionContext.OrchestrationID,
			"sender_pod_name":      params.ExecutionContext.Sender.PodName,
			"sender_agent_version": params.ExecutionContext.Sender.AgentVersion,
			"sender_role":          params.ExecutionContext.Sender.Role,

			// Response routing
			"responses_topic":        myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,

			// Additional metadata
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"action":    action,
		},
		"body": map[string]interface{}{
			"action": action,
			"data":   gitData, // This is our Git-specific payload

			// Response routing in body as well
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,

			// Additional metadata for the adapter
			"metadata": map[string]interface{}{
				"requesting_agent_id":   params.ExecutionContext.OrchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
			},

			// Include original request context
			"request_context": map[string]interface{}{
				"correlation_id":   params.ExecutionContext.CorrelationID,
				"orchestration_id": params.ExecutionContext.OrchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	// 8. Marshal the *entire* adapterRequest map
	// Convert entire request to JSON
	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal adapter request: %w", err)
	}

	// Convert headers to map[string]string for validation
	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string)

	for k, v := range rawHeaders {
		if str, ok := v.(string); ok {
			headers[k] = str
		} else {
			headers[k] = fmt.Sprintf("%v", v) // fallback stringify for non-string values
		}
	}

	key := []byte(params.ExecutionContext.CorrelationID)

	// 9. Send the message to the adapter topic
	if err := params.Producer.ProduceWithValidation(
		ctx,
		adapterTopic,
		headers,
		key,
		messageBytes,
	); err != nil {
		return nil, fmt.Errorf("failed to send to webscrape adapter: %w", err)
	}

	if err != nil {
		params.Logger.Error("Failed to send Kafka message to git-adapter", zap.Error(err), zap.String("topic", adapterTopic))
		return nil, fmt.Errorf("failed to send message to adapter: %w", err)
	}

	// 10. Return the result struct
	result := GitCommitResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   adapterTopic,
		AwaitResponse: true, // We expect the orchestrator to wait for the adapter's response
		Metadata: map[string]interface{}{
			"repo_name":  repoName,
			"file_count": len(filesMap),
		},
	}

	return result, nil
}
