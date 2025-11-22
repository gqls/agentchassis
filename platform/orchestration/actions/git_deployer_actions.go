// internal/backend/agent-chassis/platform/orchestration/actions/deployer_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GitCommitResult represents the result of a Git commit operation.
type GitCommitResult struct {
	Success       bool                   `json:"success"`
	RequestID     string                 `json:"request_id"`
	TopicSentTo   string                 `json:"topic_sent_to"`
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GitCommitAction sends a commit request to the configured git-adapter.
// Supports two modes:
//   1. Direct files map: config["files"] = {"filename": "content", ...}
//   2. Content field reference: config["content_field"] = "path.to.content" + config["filename"] = "index.html"
func GitCommitAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing GitCommitAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
	)

	// 1. Extract configuration
	config := params.StepConfig.Config
	action := "commit"

	// 2. Get client_id and response topic
	clientID := params.ExecutionContext.ClientID
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		params.Logger.Warn("ResponsesTopic not set in ExecutionContext, responses may be lost",
			zap.String("step_name", params.ExecutionContext.StepName),
		)
	}

	// 3. Get adapter topic
	adapterTopic := "system.adapter.gitcommit.requests"

	// 4. Extract specific parameters from the config
	repoName, _ := config["repo_name"].(string)
	commitMessage, _ := config["commit_message"].(string)

	// 5. Build filesMap - support two modes
	filesMap := make(map[string]string)

	// Mode 1: Check for content_field + filename (new simplified approach)
	if contentField, ok := config["content_field"].(string); ok && contentField != "" {
		filename, _ := config["filename"].(string)
		if filename == "" {
			filename = "index.html" // Default filename
		}

		// Extract content from CollectedData using the field path
		content := extractNestedFieldForGit(params.CollectedData, contentField)
		if content == nil {
			params.Logger.Error("Content field not found in CollectedData",
				zap.String("content_field", contentField),
				zap.Any("available_keys", getMapKeys(params.CollectedData)),
			)
			return nil, fmt.Errorf("content_field '%s' not found in CollectedData", contentField)
		}

		contentStr, ok := content.(string)
		if !ok {
			params.Logger.Error("Content is not a string",
				zap.String("content_field", contentField),
				zap.String("actual_type", fmt.Sprintf("%T", content)),
			)
			return nil, fmt.Errorf("content at '%s' is not a string, got %T", contentField, content)
		}

		// Strip markdown code fences if present (LLM often wraps HTML in ```html ... ```)
		contentStr = stripCodeFences(contentStr)

		filesMap[filename] = contentStr

		params.Logger.Info("Built files map from content_field",
			zap.String("content_field", contentField),
			zap.String("filename", filename),
			zap.Int("content_length", len(contentStr)),
		)

	} else if filesRaw, ok := config["files"].(map[string]interface{}); ok {
		// Mode 2: Direct files map (existing behavior)
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
	} else {
		return nil, fmt.Errorf("GitCommitAction requires either 'content_field' or 'files' in config")
	}

	if len(filesMap) == 0 {
		return nil, fmt.Errorf("no files to commit")
	}

	// 6. Build the specific 'data' payload for the adapter's body
	gitData := map[string]interface{}{
		"repo_name":      repoName,
		"files":          filesMap,
		"commit_message": commitMessage,
	}

	// 7. Generate a unique request ID for this specific adapter call
	newRequestID := uuid.New().String()

	// 8. Build the full adapterRequest
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":          params.ExecutionContext.CorrelationID,
			"orchestration_id":        params.ExecutionContext.OrchestrationID,
			"orchestration_name":      params.ExecutionContext.OrchestrationName,
			"parent_orchestration_id": params.ExecutionContext.ParentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",
			"sender_agent_type":       params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":         params.ExecutionContext.OrchestrationID,
			"sender_pod_name":         params.ExecutionContext.Sender.PodName,
			"sender_agent_version":    params.ExecutionContext.Sender.AgentVersion,
			"sender_role":             params.ExecutionContext.Sender.Role,
			"responses_topic":         myResponsesTopic,
			"parent_responses_topic":  myResponsesTopic,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
			"action":                  action,
		},
		"body": map[string]interface{}{
			"action":                 action,
			"data":                   gitData,
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
			"metadata": map[string]interface{}{
				"requesting_agent_id":   params.ExecutionContext.OrchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
			},
			"request_context": map[string]interface{}{
				"correlation_id":   params.ExecutionContext.CorrelationID,
				"orchestration_id": params.ExecutionContext.OrchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	// 9. Marshal and send
	messageBytes, err := json.Marshal(adapterRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal adapter request: %w", err)
	}

	rawHeaders := adapterRequest["headers"].(map[string]interface{})
	headers := make(map[string]string)
	for k, v := range rawHeaders {
		if str, ok := v.(string); ok {
			headers[k] = str
		} else {
			headers[k] = fmt.Sprintf("%v", v)
		}
	}

	key := []byte(params.ExecutionContext.CorrelationID)

	if err := params.Producer.ProduceWithValidation(ctx, adapterTopic, headers, key, messageBytes); err != nil {
		params.Logger.Error("Failed to send Kafka message to git-adapter", zap.Error(err), zap.String("topic", adapterTopic))
		return nil, fmt.Errorf("failed to send message to adapter: %w", err)
	}

	// 10. Return result
	result := GitCommitResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   adapterTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"repo_name":  repoName,
			"file_count": len(filesMap),
		},
	}

	return result, nil
}

// extractNestedFieldForGit extracts a value from a nested map using dot notation
// e.g., "final_site_data.generate_content.result" extracts data["final_site_data"]["generate_content"]["result"]
func extractNestedFieldForGit(data map[string]interface{}, fieldPath string) interface{} {
	parts := strings.Split(fieldPath, ".")
	var current interface{} = data

	for _, part := range parts {
		if currentMap, ok := current.(map[string]interface{}); ok {
			current = currentMap[part]
			if current == nil {
				return nil
			}
		} else {
			return nil
		}
	}

	return current
}

// stripCodeFences removes markdown code fences (```html ... ``` or ``` ... ```) from content
func stripCodeFences(content string) string {
	content = strings.TrimSpace(content)

	// Check for opening fence with language hint (```html, ```xml, etc.)
	if strings.HasPrefix(content, "```") {
		// Find end of first line (the opening fence)
		if idx := strings.Index(content, "\n"); idx != -1 {
			content = content[idx+1:]
		}
	}

	// Check for closing fence
	if strings.HasSuffix(content, "```") {
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	return content
}
