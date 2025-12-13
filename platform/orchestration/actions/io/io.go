// FILE: platform/orchestration/actions/io/io.go
// Package io provides external I/O actions: git, http, notifications
package io

import (
	"context"
	"fmt"

	"github.com/aqls/agentchassis/platform/orchestration/actions/registry"
)

func init() {
	// Git operations
	registry.Register("git_commit", registry.ActionDefinition{
		Func:        GitCommitAction,
		Category:    registry.CategoryIO,
		Description: "Commits files to a Git repository via the git-adapter",
		Status:      registry.StatusActive,
	})

	registry.Register("git_commit_action", registry.ActionDefinition{
		Func:        GitCommitAction, // Alias
		Category:    registry.CategoryIO,
		Description: "Alias for git_commit",
		Status:      registry.StatusDeprecated,
	})

	// HTTP operations
	registry.Register("http_request", registry.ActionDefinition{
		Func:        HTTPRequestAction,
		Category:    registry.CategoryIO,
		Description: "Makes an HTTP request to an external endpoint",
		Status:      registry.StatusActive,
	})

	// Notifications
	registry.Register("send_notification", registry.ActionDefinition{
		Func:        SendNotificationAction,
		Category:    registry.CategoryIO,
		Description: "Sends a notification (email, slack, webhook)",
		Status:      registry.StatusActive,
	})
}

// GitCommitAction commits files to a git repository
// Migrated from git_deployer_actions.go
func GitCommitAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	repoName, _ := params.Config["repo_name"].(string)
	if repoName == "" {
		repoName = "sites" // default
	}

	// Extract domain using field path
	domain := extractDomainFromParams(params)
	if domain == "" {
		return nil, fmt.Errorf("git_commit: could not extract domain")
	}

	// Extract files to commit
	files, err := extractFilesFromParams(params)
	if err != nil {
		return nil, fmt.Errorf("git_commit: %w", err)
	}

	commitMessage, _ := params.Config["commit_message"].(string)
	if commitMessage == "" {
		commitMessage = fmt.Sprintf("Update site: %s", domain)
	}

	params.Logger.Info("Committing to git") // zap.String("repo", repoName),
	// zap.String("domain", domain),
	// zap.Int("file_count", len(files)),

	// TODO: Migrate actual git adapter call from git_deployer_actions.go
	// This includes:
	// - Calling the git-adapter service
	// - Handling authentication
	// - Processing response

	return map[string]interface{}{
		"status":     "committed",
		"repo":       repoName,
		"domain":     domain,
		"file_count": len(files),
	}, nil
}

// HTTPRequestAction makes an HTTP request
func HTTPRequestAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	url, ok := params.Config["url"].(string)
	if !ok {
		return nil, fmt.Errorf("http_request requires 'url' in config")
	}

	method, _ := params.Config["method"].(string)
	if method == "" {
		method = "GET"
	}

	params.Logger.Info("Making HTTP request") // zap.String("method", method),
	// zap.String("url", url),

	// TODO: Migrate from generic_actions.go or implement
	// Make the actual HTTP request

	return map[string]interface{}{
		"status": "requested",
		"url":    url,
		"method": method,
	}, nil
}

// SendNotificationAction sends notifications
func SendNotificationAction(ctx context.Context, params registry.ActionParams) (interface{}, error) {
	notificationType, _ := params.Config["type"].(string)
	if notificationType == "" {
		notificationType = "webhook"
	}

	params.Logger.Info("Sending notification") // zap.String("type", notificationType),

	// TODO: Implement notification logic

	return map[string]interface{}{
		"status": "sent",
		"type":   notificationType,
	}, nil
}

// Helper functions

func extractDomainFromParams(params registry.ActionParams) string {
	// Try domain_field first
	if domainField, ok := params.Config["domain_field"].(string); ok {
		if val := extractNestedField(params.CollectedData, domainField); val != nil {
			if s, ok := val.(string); ok {
				return s
			}
		}
	}

	// Try direct domain in collected data
	if domain, ok := params.CollectedData["domain"].(string); ok {
		return domain
	}

	// Try input_data.domain
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok {
			return domain
		}
	}

	return ""
}

func extractFilesFromParams(params registry.ActionParams) (map[string]string, error) {
	files := make(map[string]string)

	// Try files_field first (multi-file mode)
	if filesField, ok := params.Config["files_field"].(string); ok {
		if val := extractNestedField(params.CollectedData, filesField); val != nil {
			if filesMap, ok := val.(map[string]interface{}); ok {
				for name, content := range filesMap {
					if s, ok := content.(string); ok {
						files[name] = s
					}
				}
				return files, nil
			}
		}
	}

	// Try content_field (single file mode)
	if contentField, ok := params.Config["content_field"].(string); ok {
		if val := extractNestedField(params.CollectedData, contentField); val != nil {
			if s, ok := val.(string); ok {
				files["index.html"] = s
				return files, nil
			}
		}
	}

	// Try direct files in config
	if configFiles, ok := params.Config["files"].(map[string]interface{}); ok {
		for name, content := range configFiles {
			if s, ok := content.(string); ok {
				files[name] = s
			}
		}
		return files, nil
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no files found to commit")
	}

	return files, nil
}

func extractNestedField(data map[string]interface{}, fieldPath string) interface{} {
	// Split path and traverse
	// This is a simplified version - use the shared helper from data package
	parts := splitFieldPath(fieldPath)
	current := interface{}(data)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		default:
			return nil
		}
		if current == nil {
			return nil
		}
	}

	return current
}

func splitFieldPath(path string) []string {
	var parts []string
	var current string
	for _, c := range path {
		if c == '.' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
