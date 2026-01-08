// git_deployer_actions_updated.go
// Updates to GitCommitAction to support files_field extraction from CollectedData

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
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
	clientID := params.ExecutionContext.ClientID
	myResponsesTopic := params.ExecutionContext.ResponsesTopic
	if myResponsesTopic == "" {
		params.Logger.Warn("ResponsesTopic not set in ExecutionContext")
	}

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

// recursiveFindDomain searches recursively for a "domain" key in nested maps
// extractFilesForGit extracts files map from CollectedData or config
// Uses the unified extractor infrastructure for consistent field extraction
func extractFilesForGit(data map[string]interface{}, config map[string]interface{}, domain string, logger *zap.Logger) map[string]string {
	filesMap := make(map[string]string)

	// Method 1: Use files_field config with unified extractor (for multi-page batch)
	filesField, hasFilesField := config["files_field"].(string)

	if hasFilesField && filesField != "" {
		logger.Info("Extracting files using files_field",
			zap.String("files_field", filesField))

		extracted := datahelpers.ExtractFields(data, []string{filesField}, logger)
		pathParts := strings.Split(filesField, ".")
		extractedKey := pathParts[len(pathParts)-1]

		if filesData, ok := extracted[extractedKey]; ok {
			if files, ok := filesData.(map[string]interface{}); ok {
				for filename, content := range files {
					if contentStr, ok := content.(string); ok {
						filesMap[filename] = contentStr
						logger.Info("Added file from files_field",
							zap.String("filename", filename),
							zap.Int("size", len(contentStr)))
					}
				}
				if len(filesMap) > 0 {
					return filesMap
				}
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

	// Method 3: Check content_field for single file (backward compatibility)
	if contentField, ok := config["content_field"].(string); ok && contentField != "" {
		extracted := datahelpers.ExtractFields(data, []string{contentField}, logger)
		contentPathParts := strings.Split(contentField, ".")
		contentKey := contentPathParts[len(contentPathParts)-1]

		if content, ok := extracted[contentKey]; ok {
			if contentStr, ok := content.(string); ok && contentStr != "" {
				filesMap["index.html"] = contentStr
				logger.Info("Using content_field for single file")
				return filesMap
			}
		}
	}

	// Method 4: NEW - html_from + page_from pattern (for loop-based single page commits)
	// This is used by pageflow-builder's build_pages_loop
	htmlFrom, hasHtmlFrom := config["html_from"].(string)
	pageFrom, hasPageFrom := config["page_from"].(string)

	if hasHtmlFrom && htmlFrom != "" {
		logger.Info("Trying html_from + page_from pattern",
			zap.String("html_from", htmlFrom),
			zap.String("page_from", pageFrom))

		// Extract HTML content using the helper that handles nested paths
		htmlContent := datahelpers.ExtractNestedFieldString(data, htmlFrom)

		if htmlContent == "" {
			logger.Warn("html_from path did not yield content",
				zap.String("html_from", htmlFrom))
		} else {
			// Determine filename from page_from
			filename := "index.html" // default

			if hasPageFrom && pageFrom != "" {
				pageData := datahelpers.ExtractNestedField(data, pageFrom)
				if pageMap, ok := pageData.(map[string]interface{}); ok {
					// Try url field first (e.g., "/index.html")
					if url, ok := pageMap["url"].(string); ok && url != "" {
						filename = strings.TrimPrefix(url, "/")
						// Ensure it has .html extension
						if !strings.HasSuffix(filename, ".html") {
							filename = filename + ".html"
						}
					} else if name, ok := pageMap["name"].(string); ok && name != "" {
						// Fallback to name field
						filename = name + ".html"
					}
				}
			}

			// Clean the filename
			filename = filepath.Clean(filename)
			if strings.HasPrefix(filename, "/") {
				filename = strings.TrimPrefix(filename, "/")
			}

			filesMap[filename] = htmlContent
			logger.Info("Using html_from + page_from for single file",
				zap.String("filename", filename),
				zap.Int("content_size", len(htmlContent)))
			return filesMap
		}
	}

	// No files_field configured, try default path as last resort
	if !hasFilesField {
		defaultPath := "site_files.files"
		logger.Info("No files_field configured, trying default multipage path",
			zap.String("default_path", defaultPath))

		extracted := datahelpers.ExtractFields(data, []string{defaultPath}, logger)
		if filesData, ok := extracted["files"]; ok {
			if files, ok := filesData.(map[string]interface{}); ok {
				for filename, content := range files {
					if contentStr, ok := content.(string); ok {
						filesMap[filename] = contentStr
					}
				}
				if len(filesMap) > 0 {
					return filesMap
				}
			}
		}
	}

	// Log available keys for debugging
	datahelpers.LogCollectedDataStructure(data, logger, "error_")

	logger.Warn("No files found via any method",
		zap.Any("config_keys", datahelpers.GetMapKeys(config)),
		zap.String("files_field", filesField),
		zap.String("html_from", htmlFrom),
		zap.String("page_from", pageFrom))

	return filesMap
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
