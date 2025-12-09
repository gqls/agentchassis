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
	AwaitResponse bool                   `json:"await_response"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GitCommitAction sends a commit request to the configured git-adapter
// Supports both direct files in config and files_field path to extract from CollectedData
func GitCommitAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Executing GitCommitAction",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.String("orchestration_id", params.ExecutionContext.OrchestrationID),
		zap.Any("config_keys", getMapKeysGit(params.StepConfig.Config)),
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

// extractDomainForGit extracts domain from CollectedData using field path
func extractDomainForGit(data map[string]interface{}, config map[string]interface{}, logger *zap.Logger) string {
	// Get configured domain_field
	domainField, _ := config["domain_field"].(string)

	logger.Info("Attempting to extract domain",
		zap.String("configured_field", domainField))

	// Build list of paths to try
	pathsToTry := []string{}

	// If domain_field is configured, try it and variations
	if domainField != "" {
		pathsToTry = append(pathsToTry,
			domainField,               // Original configured field
			"input_data."+domainField, // With input_data prefix
		)
	}

	// Add common fallback paths with increasing nesting levels
	// These handle the case where agent calls wrap data in additional input_data layers
	pathsToTry = append(pathsToTry,
		"domain",                                                                                                                                                // Top-level
		"input_data.domain",                                                                                                                                     // 1 level deep
		"input_data.input_data.domain",                                                                                                                          // 2 levels deep
		"input_data.input_data.input_data.domain",                                                                                                               // 3 levels deep
		"input_data.input_data.input_data.input_data.domain",                                                                                                    // 4 levels deep
		"input_data.input_data.input_data.input_data.input_data.domain",                                                                                         // 5 levels deep
		"input_data.input_data.input_data.input_data.input_data.input_data.domain",                                                                              // 6 levels deep
		"input_data.input_data.input_data.input_data.input_data.input_data.input_data.domain",                                                                   // 7 levels deep
		"input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.domain",                                                        // 8 levels deep
		"input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.domain",                                             // 9 levels deep
		"input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.domain",                                  // 10 levels deep
		"input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.domain",                       // 11 levels deep
		"input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.input_data.domain", // 12 levels deep
	)

	// Remove duplicates
	seen := make(map[string]bool)
	uniquePaths := []string{}
	for _, path := range pathsToTry {
		if !seen[path] {
			seen[path] = true
			uniquePaths = append(uniquePaths, path)
		}
	}

	// Try each path
	for _, path := range uniquePaths {
		logger.Debug("Trying domain path", zap.String("path", path))

		domainData := datahelpers.ExtractNestedField(data, path)
		if domainData != nil {
			if domainStr, ok := domainData.(string); ok && domainStr != "" {
				logger.Info("Successfully extracted domain",
					zap.String("path_used", path),
					zap.String("domain", domainStr))
				return domainStr
			}
		}
	}

	// Last resort: recursive search for "domain" key anywhere in the structure
	logger.Info("Trying recursive search for domain")
	if foundDomain := recursiveFindDomain(data, logger); foundDomain != "" {
		logger.Info("Found domain via recursive search", zap.String("domain", foundDomain))
		return foundDomain
	}

	logger.Warn("Failed to extract domain from any path",
		zap.Strings("tried_paths", uniquePaths))

	return "unknown-domain"
}

// recursiveFindDomain searches recursively for a "domain" key in nested maps
func recursiveFindDomain(data interface{}, logger *zap.Logger) string {
	switch v := data.(type) {
	case map[string]interface{}:
		// Check if this map has a "domain" key directly
		if domain, ok := v["domain"].(string); ok && domain != "" {
			return domain
		}
		// Recursively search nested maps
		for key, value := range v {
			// Skip special internal keys
			if strings.HasPrefix(key, "__") {
				continue
			}
			if found := recursiveFindDomain(value, logger); found != "" {
				return found
			}
		}
	case []interface{}:
		// Search arrays
		for _, item := range v {
			if found := recursiveFindDomain(item, logger); found != "" {
				return found
			}
		}
	}
	return ""
}

// extractFilesForGit extracts files map from CollectedData or config
func extractFilesForGit(data map[string]interface{}, config map[string]interface{}, domain string, logger *zap.Logger) map[string]string {
	filesMap := make(map[string]string)

	// Method 1: Check files_field config (NEW - for multi-page support)
	if filesField, ok := config["files_field"].(string); ok && filesField != "" {
		logger.Info("Attempting to extract files from field path",
			zap.String("files_field", filesField))

		if filesData := datahelpers.ExtractNestedField(data, filesField); filesData != nil {
			if files, ok := filesData.(map[string]interface{}); ok {
				for filename, content := range files {
					if contentStr, ok := content.(string); ok {
						// Prepend domain to create path: domain/filename
						fullPath := filename
						/*if !strings.HasPrefix(filename, domain+"/") && !strings.HasPrefix(filename, domain+"\\") {
							fullPath = filepath.Join(domain, filename)
						}*/
						filesMap[fullPath] = contentStr
						logger.Info("Added file from files_field",
							zap.String("path", fullPath),
							zap.Int("size", len(contentStr)))
					}
				}
				if len(filesMap) > 0 {
					return filesMap
				}
			}
		}

		// Try further multiple possible paths for files_field
		pathsToTry := []string{
			filesField,                 // Original path: "site_files.files"
			"input_data." + filesField, // input_data.site_files.files
			"input_data.site_files.wrap_multipage.files", // Known multipage path
		}

		for _, path := range pathsToTry {
			if filesData := datahelpers.ExtractNestedField(data, path); filesData != nil {
				logger.Info("Found files from further search at path",
					zap.String("path", path),
				)

				if files, ok := filesData.(map[string]interface{}); ok {
					for filename, content := range files {
						if contentStr, ok := content.(string); ok {
							// Prepend domain to create path: domain/filename
							// fullPath := filepath.Join(domain, filename)
							filesMap[filename] = contentStr
							logger.Info("Successfully added file from files_field",
								zap.String("path", filename),
								zap.Int("size", len(contentStr)))
						}
					}
					if len(filesMap) > 0 {
						return filesMap
					}
				}
			}
		}
	}

	// Method 2: Check direct files in config (legacy support)
	if filesRaw, ok := config["files"].(map[string]interface{}); ok {
		for filename, content := range filesRaw {
			if contentStr, ok := content.(string); ok {
				fullPath := filename
				/*if !strings.HasPrefix(filename, domain+"/") && !strings.HasPrefix(filename, domain+"\\") {
					fullPath = filepath.Join(domain, filename)
				}*/
				filesMap[fullPath] = contentStr
			}
		}
		if len(filesMap) > 0 {
			return filesMap
		}
	}

	// Method 3: Check content_field for single file (backward compatibility)
	if contentField, ok := config["content_field"].(string); ok && contentField != "" {
		if content := datahelpers.ExtractNestedField(data, contentField); content != nil {
			if contentStr, ok := content.(string); ok && contentStr != "" {
				//fullPath := filepath.Join(domain, "index.html")
				fullPath := "index.html"
				filesMap[fullPath] = contentStr
				logger.Info("Using content_field for single file",
					zap.String("path", fullPath))
				return filesMap
			}
		}
	}

	// Returns first one that works

	logger.Warn("No files found via any method",
		zap.Any("config_keys", getMapKeysGit(config)),
		zap.Any("data_keys", getMapKeysGit(data)))

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

// getMapKeysGit returns keys of a map (helper)
func getMapKeysGit(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
