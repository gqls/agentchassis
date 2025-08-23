// platform/orchestration/actions/storage_actions.go
package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// UploadToS3Action handles S3 uploads using the actual S3 client
func UploadToS3Action(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Starting S3 upload action")

	// Get storage config from agent config or step config
	var storageConfig config.ObjectStorageConfig

	// First try agent config
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if storageCfg, ok := agentConfig["storage_config"].(map[string]interface{}); ok {
			storageConfig.Provider = getStringOrDefault(storageCfg, "provider", "s3")
			storageConfig.Endpoint = getStringOrDefault(storageCfg, "endpoint", "")
			storageConfig.Bucket = getStringOrDefault(storageCfg, "bucket", "")
			storageConfig.AccessKeyEnvVar = getStringOrDefault(storageCfg, "access_key_env_var", "AWS_ACCESS_KEY_ID")
			storageConfig.SecretKeyEnvVar = getStringOrDefault(storageCfg, "secret_key_env_var", "AWS_SECRET_ACCESS_KEY")
		}
	}

	// Override with step config if present
	if params.StepConfig.Config != nil {
		if bucket, ok := params.StepConfig.Config["bucket"].(string); ok && bucket != "" {
			storageConfig.Bucket = bucket
		}
		if bucketOverride, ok := params.StepConfig.Config["bucket_override"].(string); ok && bucketOverride != "" {
			storageConfig.Bucket = bucketOverride
		}
	}

	// Create S3 client
	s3Client, err := storage.NewS3Client(ctx, storageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Get the website data to upload
	var websiteData map[string]interface{}

	// Look for website data from previous steps
	if htmlData, ok := params.CollectedData["develop_site"]; ok {
		websiteData = extractWebsiteFiles(htmlData)
	} else if htmlData, ok := params.CollectedData["html_developer"]; ok {
		websiteData = extractWebsiteFiles(htmlData)
	} else {
		// Try to find any HTML content in collected data
		for key, value := range params.CollectedData {
			if strings.Contains(key, "html") || strings.Contains(key, "site") {
				websiteData = extractWebsiteFiles(value)
				if len(websiteData) > 0 {
					break
				}
			}
		}
	}

	if len(websiteData) == 0 {
		return nil, fmt.Errorf("no website data found to upload")
	}

	params.Logger.Info("Found website files to upload",
		zap.Int("file_count", len(websiteData)))

	// Upload each file
	uploadedFiles := make(map[string]string)
	correlationID := params.Headers["correlation_id"]
	prefix := fmt.Sprintf("sites/%s/", correlationID[:8])

	for filename, content := range websiteData {
		key := prefix + filename
		contentType := getContentType(filename)

		// Convert content to reader
		var contentBytes []byte
		switch v := content.(type) {
		case string:
			contentBytes = []byte(v)
		case []byte:
			contentBytes = v
		default:
			// Try to marshal as JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				params.Logger.Warn("Skipping file, cannot convert to bytes",
					zap.String("filename", filename))
				continue
			}
			contentBytes = jsonBytes
		}

		reader := bytes.NewReader(contentBytes)

		// Upload to S3
		s3URI, err := s3Client.Upload(ctx, key, contentType, reader)
		if err != nil {
			params.Logger.Error("Failed to upload file",
				zap.String("filename", filename),
				zap.Error(err))
			continue
		}

		uploadedFiles[filename] = s3URI
		params.Logger.Info("Uploaded file",
			zap.String("filename", filename),
			zap.String("s3_uri", s3URI))
	}

	// Generate public URL if requested
	publicURL := ""
	if makePublic, ok := params.StepConfig.Config["make_public"].(bool); ok && makePublic {
		// For index.html, generate a presigned URL
		if indexURI, ok := uploadedFiles["index.html"]; ok {
			// Extract key from URI
			key := strings.TrimPrefix(indexURI, fmt.Sprintf("s3://%s/", storageConfig.Bucket))
			url, err := s3Client.GetPresignedURL(ctx, key, 60*24*7) // 7 days
			if err == nil {
				publicURL = url
			}
		}
	}

	result := map[string]interface{}{
		"uploaded":     true,
		"files":        uploadedFiles,
		"file_count":   len(uploadedFiles),
		"bucket":       storageConfig.Bucket,
		"prefix":       prefix,
		"s3_base_path": fmt.Sprintf("s3://%s/%s", storageConfig.Bucket, prefix),
	}

	if publicURL != "" {
		result["public_url"] = publicURL
	}

	params.Logger.Info("S3 upload completed",
		zap.Int("uploaded_count", len(uploadedFiles)),
		zap.String("bucket", storageConfig.Bucket))

	return result, nil
}

// Helper functions

func extractWebsiteFiles(data interface{}) map[string]interface{} {
	files := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		// Look for common keys that contain website files
		if htmlContent, ok := v["html"]; ok {
			files["index.html"] = htmlContent
		}
		if cssContent, ok := v["css"]; ok {
			files["styles.css"] = cssContent
		}
		if jsContent, ok := v["javascript"]; ok {
			files["script.js"] = jsContent
		}
		if result, ok := v["result"]; ok {
			// Recursively extract from result
			return extractWebsiteFiles(result)
		}
		if files, ok := v["files"].(map[string]interface{}); ok {
			return files
		}
		// If the whole thing looks like files, return it
		if len(files) == 0 && containsWebFiles(v) {
			return v
		}
	}

	return files
}

func containsWebFiles(data map[string]interface{}) bool {
	for key := range data {
		if strings.HasSuffix(key, ".html") ||
			strings.HasSuffix(key, ".css") ||
			strings.HasSuffix(key, ".js") ||
			strings.HasSuffix(key, ".json") {
			return true
		}
	}
	return false
}

func getContentType(filename string) string {
	switch {
	case strings.HasSuffix(filename, ".html"):
		return "text/html"
	case strings.HasSuffix(filename, ".css"):
		return "text/css"
	case strings.HasSuffix(filename, ".js"):
		return "application/javascript"
	case strings.HasSuffix(filename, ".json"):
		return "application/json"
	case strings.HasSuffix(filename, ".png"):
		return "image/png"
	case strings.HasSuffix(filename, ".jpg"), strings.HasSuffix(filename, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(filename, ".svg"):
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}

func getStringOrDefault(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}
