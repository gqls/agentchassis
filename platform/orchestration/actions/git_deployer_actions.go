// internal/backend/agent-chassis/platform/orchestration/actions/deployer_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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
//  1. Direct files map: config["files"] = {"filename": "content", ...}
//  2. Content field reference: config["content_field"] = "path.to.content" + config["filename"] = "index.html"
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
	adapterTopic := "system.adapter.git.requests"

	// 4. Extract specific parameters from the config
	// Build template data using existing function
	templateData := extractDataForGitAgent(params).(map[string]interface{})

	// Extract and resolve template variables
	repoName, _ := config["repo_name"].(string)
	commitMessage, _ := config["commit_message"].(string)

	// Resolve templates
	if strings.Contains(repoName, "{{") {
		resolved, err := datahelpers.RenderPromptTemplate(repoName, templateData, *params.Logger)
		if err != nil {
			params.Logger.Warn("Failed to resolve repo_name template", zap.Error(err))
		} else {
			repoName = resolved
		}
	}

	if strings.Contains(commitMessage, "{{") {
		resolved, err := datahelpers.RenderPromptTemplate(commitMessage, templateData, *params.Logger)
		if err != nil {
			params.Logger.Warn("Failed to resolve commit_message template", zap.Error(err))
		} else {
			commitMessage = resolved
		}
	}

	params.Logger.Info("Resolved template values",
		zap.String("repo_name", repoName),
		zap.String("commit_message", commitMessage),
	)
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
		contentStr = datahelpers.CleanMarkdownJSON(contentStr)

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

// extractDataForGitAgent merges data from multiple sources specified in the step's 'input_fields' config.
// very similar to extractDataForAgent and extractDataForAIAgent - needs consolidation
func extractDataForGitAgent(params ActionParams) interface{} {
	params.Logger.Info("Extracting data for AI agent",
		zap.Any("available_keys", GetMapKeys(params.CollectedData)),
	)

	templateData := make(map[string]interface{})

	// 1. Determine inputs to fetch
	var inputFields []string
	if fields, ok := params.StepConfig.Config["input_fields"].([]interface{}); ok {
		for _, fieldInterface := range fields {
			if field, ok := fieldInterface.(string); ok {
				inputFields = append(inputFields, field)
			}
		}
	} else {
		params.Logger.Warn("No 'input_fields' found in config, defaulting to ['input_data']")
		inputFields = []string{"input_data"}
	}

	params.Logger.Info("Processing input_fields", zap.Strings("fields", inputFields))

	// Get input_data once for reuse
	inputDataMap := datahelpers.GetInputData(params.CollectedData, params.Logger)

	// Check for double-nesting and unwrap if needed
	if nestedInputData, ok := inputDataMap["input_data"].(map[string]interface{}); ok {
		params.Logger.Warn("Detected double-nested input_data, unwrapping")
		inputDataMap = nestedInputData
	}

	// 2. Smart Extraction Loop
	for _, fieldName := range inputFields {

		// Scenario A: "input_data" keyword (Legacy/Bulk behavior)
		// Flattens the entire input_data map into the template root
		if fieldName == "input_data" {
			for key, val := range inputDataMap {
				templateData[key] = val
			}
			params.Logger.Info("Flattened input_data to root", zap.Int("field_count", len(inputDataMap)))
			continue
		}

		// Scenario B: Specific Field Lookup
		var foundValue interface{}
		var found bool
		var foundPath string

		// Check 1: Direct lookup in CollectedData
		if val, ok := datahelpers.GetValueByPath(params.CollectedData, fieldName, params.Logger); ok {
			foundValue = val
			found = true
			foundPath = fieldName
		}

		// Check 2: Look in the unwrapped input_data map directly
		if !found && inputDataMap != nil {
			if val, ok := inputDataMap[fieldName]; ok {
				foundValue = val
				found = true
				foundPath = "input_data." + fieldName
				params.Logger.Debug("Found field in input_data", zap.String("field", fieldName))
			}
		}

		// Check 3: Try dot notation in CollectedData
		if !found {
			if val, ok := datahelpers.GetValueByPath(params.CollectedData, "input_data."+fieldName, params.Logger); ok {
				foundValue = val
				found = true
				foundPath = "input_data." + fieldName
			}
		}

		// Check 4: Look inside __raw_message__
		if !found {
			if raw, ok := params.CollectedData["__raw_message__"].(map[string]interface{}); ok {
				if val, ok := datahelpers.GetValueByPath(raw, fieldName, params.Logger); ok {
					foundValue = val
					found = true
					foundPath = "__raw_message__." + fieldName
				}
			}
		}

		if found {
			// Use the simple field name for the template key
			// e.g. if we found "input_data.domain", store as "domain"
			keyParts := strings.Split(fieldName, ".")
			simpleKey := keyParts[len(keyParts)-1]

			templateData[simpleKey] = foundValue
			params.Logger.Info("Extracted field",
				zap.String("field", fieldName),
				zap.String("template_key", simpleKey),
				zap.String("found_at", foundPath),
				zap.Any("value", foundValue),
			)
		} else {
			params.Logger.Warn("Requested input_field not found",
				zap.String("field", fieldName),
				zap.Strings("checked_paths", []string{
					fieldName,
					"input_data." + fieldName,
					"CollectedData[input_data][" + fieldName + "]",
					"__raw_message__." + fieldName,
				}),
			)
		}
	}

	params.Logger.Info("Final template data",
		zap.Any("template_data", templateData),
		zap.Int("field_count", len(templateData)),
	)

	return templateData
}
