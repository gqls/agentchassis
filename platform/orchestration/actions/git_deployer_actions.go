// git_deployer_actions_updated.go
// Updates to GitCommitAction to support files_field extraction from CollectedData

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
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
	RequestsTopic string                 `json:"requests_topic"`
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GitCommitAction sends a commit request to the configured git-adapter
// Supports both direct files in config and files_field path to extract from CollectedData
func GitCommitAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing GitCommitAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.Any("config_keys", datahelpers.GetMapKeys(params.StepConfig.Config)),
	)

	// 1. Extract configuration
	config := params.StepConfig.Config

	// Get client_id and response topic
	// Get execution context values - fall back to CollectedData if params.ExecutionContext is empty
	// This handles loop iterations where ExecutionContext may not be fully propagated
	correlationID := getExecutionContextValue(params, "correlation_id", params.ExecutionContext.CorrelationID)
	orchestrationID := getExecutionContextValue(params, "orchestration_id", params.ExecutionContext.OrchestrationID)
	orchestrationName := getExecutionContextValue(params, "orchestration_name", params.ExecutionContext.OrchestrationName)
	parentOrchestrationID := getExecutionContextValue(params, "parent_orchestration_id", params.ExecutionContext.ParentOrchestrationID)
	clientID := getExecutionContextValue(params, "client_id", params.ExecutionContext.ClientID)
	myResponsesTopic := getExecutionContextValue(params, "responses_topic", params.ExecutionContext.ResponsesTopic)

	// Also check __my_responses_topic__ as alternative
	if myResponsesTopic == "" {
		if alt, ok := params.CollectedData["__my_responses_topic__"].(string); ok && alt != "" {
			myResponsesTopic = alt
		}
	}

	// Final fallback to parent responses topic
	if myResponsesTopic == "" {
		if parent, ok := params.CollectedData["__parent_responses_topic__"].(string); ok && parent != "" {
			myResponsesTopic = parent
			params.Logger.Info("Using __parent_responses_topic__ as responses_topic fallback")
		}
	}

	if myResponsesTopic == "" {
		params.Logger.Warn("ResponsesTopic not set - message may fail validation")
	}

	params.Logger.Info("Resolved execution context for git commit",
		zap.String("correlation_id", correlationID),
		zap.String("orchestration_id", orchestrationID),
		zap.String("client_id", clientID),
		zap.String("responses_topic_preview", datahelpers.TruncateString(myResponsesTopic, 60)))

	// Adapter topic
	adapterTopic := "system.adapter.git.requests"

	// Extract repo name (can be from config or from CollectedData)
	repoName, _ := config["repo_name"].(string)
	if repoName == "" {
		repoName = "sites" // default
	}

	// Extract domain - supports field path extraction
	domain := extractDomainForGit(params.CollectedData, config, params.Logger)

	// Extract files - NEW: supports files_field path
	filesMap := extractFilesForGit(params.CollectedData, config, domain, params.Logger)

	if len(filesMap) == 0 {
		return nil, fmt.Errorf("no files to commit")
	}

	// Build commit message with template support
	commitMessage := buildCommitMessage(config, domain, len(filesMap))

	params.Logger.Info("Git commit prepared",
		zap.String("repo_name", repoName),
		zap.String("domain", domain),
		zap.Int("file_count", len(filesMap)),
		zap.String("commit_message", commitMessage),
		zap.Any("file_names", getFileNames(filesMap)),
	)

	// Build the git data payload
	gitData := map[string]interface{}{
		"repo_name":      repoName,
		"domain":         domain,
		"files":          filesMap,
		"commit_message": commitMessage,
	}

	// Generate request ID
	newRequestID := uuid.New().String()

	// Build full adapter request
	adapterRequest := map[string]interface{}{
		"headers": map[string]interface{}{
			"correlation_id":          correlationID,
			"orchestration_id":        orchestrationID,
			"orchestration_name":      orchestrationName,
			"parent_orchestration_id": parentOrchestrationID,
			"client_id":               clientID,
			"step_name":               params.ExecutionContext.StepName,
			"step_id":                 params.ExecutionContext.StepID,
			"request_id":              newRequestID,
			"message_type":            "request",
			"sender_agent_type":       params.ExecutionContext.Sender.AgentType,
			"sender_agent_id":         orchestrationID,
			"sender_pod_name":         params.ExecutionContext.Sender.PodName,
			"sender_agent_version":    params.ExecutionContext.Sender.AgentVersion,
			"sender_role":             params.ExecutionContext.Sender.Role,
			"responses_topic":         myResponsesTopic,
			"parent_responses_topic":  myResponsesTopic,
			"timestamp":               time.Now().UTC().Format(time.RFC3339),
			"action":                  "commit",
		},
		"body": map[string]interface{}{
			"action":                 "commit",
			"data":                   gitData,
			"reply_to_topic":         myResponsesTopic,
			"parent_responses_topic": myResponsesTopic,
			"metadata": map[string]interface{}{
				"requesting_agent_id":   params.ExecutionContext.OrchestrationID,
				"requesting_agent_type": params.ExecutionContext.Sender.AgentType,
				"requesting_step":       params.ExecutionContext.StepName,
				"client_id":             clientID,
				"domain":                domain,
				"file_count":            len(filesMap),
			},
			"request_context": map[string]interface{}{
				"correlation_id":   params.ExecutionContext.CorrelationID,
				"orchestration_id": params.ExecutionContext.OrchestrationID,
				"request_id":       newRequestID,
			},
		},
	}

	// Marshal and send
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
		params.Logger.Error("Failed to send to git-adapter", zap.Error(err))
		return nil, fmt.Errorf("failed to send message to adapter: %w", err)
	}

	result := GitCommitResult{
		Success:       true,
		RequestID:     newRequestID,
		TopicSentTo:   adapterTopic,
		RequestsTopic: adapterTopic,
		AwaitResponse: true,
		Metadata: map[string]interface{}{
			"repo_name":  repoName,
			"domain":     domain,
			"file_count": len(filesMap),
			"files":      getFileNames(filesMap),
		},
	}

	return result, nil
}

// getExecutionContextValue extracts a value from ExecutionContext,
// falling back to CollectedData["__execution_context__"] if empty.
// This handles loop iterations where params.ExecutionContext may not be fully populated.
func getExecutionContextValue(params ActionParams, field string, directValue string) string {
	// If direct value from params.ExecutionContext is present, use it
	if directValue != "" {
		return directValue
	}

	// Fall back to __execution_context__ in CollectedData
	execCtx, ok := params.CollectedData["__execution_context__"].(map[string]interface{})
	if !ok {
		params.Logger.Debug("__execution_context__ not found in CollectedData",
			zap.String("field", field))
		return ""
	}

	if val, ok := execCtx[field].(string); ok {
		params.Logger.Debug("Using execution context from CollectedData",
			zap.String("field", field),
			zap.String("value_preview", datahelpers.TruncateString(val, 50)))
		return val
	}

	return ""
}

// extractDomainForGit extracts domain from CollectedData using unified extractor
func extractDomainForGit(data map[string]interface{}, config map[string]interface{}, logger *zap.Logger) string {
	// Get configured domain_field, default to "domain"
	domainField, _ := config["domain_field"].(string)
	if domainField == "" {
		domainField = "domain"
	}

	logger.Info("Extracting domain using unified extractor",
		zap.String("domain_field", domainField))

	// Use the unified extractor for consistent field extraction
	extracted := datahelpers.ExtractFields(data, []string{domainField}, logger)

	// The extracted key will be the last part of the path
	pathParts := strings.Split(domainField, ".")
	extractedKey := pathParts[len(pathParts)-1]

	if domainData, ok := extracted[extractedKey]; ok {
		if domainStr, ok := domainData.(string); ok && domainStr != "" {
			logger.Info("Successfully extracted domain",
				zap.String("domain", domainStr))
			return domainStr
		}
	}

	logger.Warn("Failed to extract domain", zap.String("domain_field", domainField))
	return "unknown-domain"
}

// extractFilesForGit extracts files map from CollectedData or config
func extractFilesForGit(data map[string]interface{}, config map[string]interface{}, domain string, logger *zap.Logger) map[string]string {
	filesMap := make(map[string]string)

	// Method 1: Use files_field config with unified extractor (for multi-page support)
	filesField, hasFilesField := config["files_field"].(string)

	if !hasFilesField || filesField == "" {
		filesField = "site_files.files"
		logger.Debug("files_field not configured, using default multipage path",
			zap.String("default_path", filesField))
	}

	logger.Debug("Extracting files",
		zap.String("files_field", filesField))

	extracted := datahelpers.ExtractFields(data, []string{filesField}, logger)

	pathParts := strings.Split(filesField, ".")
	extractedKey := pathParts[len(pathParts)-1]

	if filesData, ok := extracted[extractedKey]; ok {
		if files, ok := filesData.(map[string]interface{}); ok {
			for filename, content := range files {
				if contentStr, ok := content.(string); ok {
					filesMap[filename] = contentStr
				}
			}
			if len(filesMap) > 0 {
				logger.Info("Extracted files from files_field",
					zap.Int("count", len(filesMap)))
				return filesMap
			}
		}
	}

	// Method 2: Check direct files in config (legacy support)
	if filesRaw, ok := config["files"].(map[string]interface{}); ok {
		for filename, content := range filesRaw {
			if contentStr, ok := content.(string); ok {
				filesMap[filename] = contentStr
			}
		}
		if len(filesMap) > 0 {
			logger.Info("Using direct files from config", zap.Int("count", len(filesMap)))
			return filesMap
		}
	}

	// Method 3: Check content_field for single file
	if contentField, ok := config["content_field"].(string); ok && contentField != "" {
		extracted := datahelpers.ExtractFields(data, []string{contentField}, logger)

		contentPathParts := strings.Split(contentField, ".")
		contentKey := contentPathParts[len(contentPathParts)-1]

		if content, ok := extracted[contentKey]; ok {
			if contentStr, ok := content.(string); ok && contentStr != "" {
				// Determine filename from page_field
				filename := determinePageFilename(data, config, logger)
				filesMap[filename] = contentStr
				logger.Info("Using content_field for single file",
					zap.String("filename", filename))
				return filesMap
			}
		}
	}

	logger.Warn("No files found via any method",
		zap.Any("config_keys", datahelpers.GetMapKeys(config)),
		zap.String("files_field", filesField))

	return filesMap
}

// determinePageFilename determines the HTML filename for a single page commit
// Uses page_field config to get the page name, defaults to "index.html"
func determinePageFilename(data map[string]interface{}, config map[string]interface{}, logger *zap.Logger) string {
	pageField, ok := config["page_field"].(string)
	if !ok || pageField == "" {
		logger.Debug("No page_field configured, using default filename")
		return "index.html"
	}

	// Extract page data using unified extractor
	extracted := datahelpers.ExtractFields(data, []string{pageField}, logger)

	// Get the last part of the page field path (e.g., "current_page" -> "current_page")
	pathParts := strings.Split(pageField, ".")
	extractedKey := pathParts[len(pathParts)-1]

	pageData, ok := extracted[extractedKey]
	if !ok {
		logger.Warn("page_field not found in data",
			zap.String("page_field", pageField))
		return "index.html"
	}

	// Handle different page data formats
	switch p := pageData.(type) {
	case map[string]interface{}:
		// Try common field names for page identifier
		// Priority: slug > name > id
		if slug, ok := p["slug"].(string); ok && slug != "" {
			return ensureHTMLExtension(slug)
		}
		if name, ok := p["name"].(string); ok && name != "" {
			return ensureHTMLExtension(name)
		}
		if id, ok := p["id"].(string); ok && id != "" {
			// ID might be a UUID, not ideal but usable
			return ensureHTMLExtension(id)
		}
		logger.Warn("page data found but no name/slug/id field",
			zap.Any("page_keys", datahelpers.GetMapKeys(p)))
	case string:
		// Direct string value (unlikely but handle it)
		return ensureHTMLExtension(p)
	default:
		logger.Warn("Unexpected page data type",
			zap.String("type", fmt.Sprintf("%T", pageData)))
	}

	return "index.html"
}

// ensureHTMLExtension ensures filename ends with .html
func ensureHTMLExtension(name string) string {
	if name == "" {
		return "index.html"
	}
	// Handle "index" specially
	if name == "index" || name == "home" {
		return "index.html"
	}
	// Already has .html
	if strings.HasSuffix(name, ".html") {
		return name
	}
	// Already has other extension - replace
	if idx := strings.LastIndex(name, "."); idx > 0 {
		return name[:idx] + ".html"
	}
	// No extension - add .html
	return name + ".html"
}

// buildCommitMessage creates commit message with template support
func buildCommitMessage(config map[string]interface{}, domain string, fileCount int) string {
	messageTemplate, _ := config["commit_message"].(string)
	if messageTemplate == "" {
		messageTemplate = "Update site: {{.domain}}"
	}

	// Simple template replacement
	tmpl, err := template.New("commit").Parse(messageTemplate)
	if err != nil {
		return fmt.Sprintf("Update site: %s", domain)
	}

	var buf strings.Builder
	data := map[string]interface{}{
		"domain":     domain,
		"file_count": fileCount,
	}
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Sprintf("Update site: %s", domain)
	}

	return buf.String()
}

// getFileNames returns list of file names from files map
func getFileNames(files map[string]string) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
