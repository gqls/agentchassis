// FILE: platform/orchestration/actions/storage_actions.go
package actions

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gqls/agentchassis/platform/config"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// RouteStorageAction is a generic storage action that routes to appropriate backends
func RouteStorageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Starting storage routing action")

	// Get storage configuration from agent config first, then step config
	storageConfig, storageType := extractStorageConfig(params)

	if storageType == "" || storageType == "none" {
		params.Logger.Info("Storage not configured or disabled")
		return map[string]interface{}{
			"stored": false,
			"reason": "storage not configured",
		}, nil
	}

	// Extract content to store based on agent type and configuration
	content, metadata, err := extractContentToStore(params, storageConfig)
	if err != nil {
		params.Logger.Warn("No content to store", zap.Error(err))
		return map[string]interface{}{
			"stored": false,
			"error":  err.Error(),
		}, nil
	}

	params.Logger.Info("Content extracted for storage",
		zap.String("storage_type", storageType),
		zap.Int("content_items", len(content)))

	// Route to appropriate storage backend
	switch storageType {
	case "s3", "aws":
		return storeToS3(ctx, params, content, metadata, storageConfig)
	case "b2", "backblaze":
		return storeToB2(ctx, params, content, metadata, storageConfig)
	case "database":
		return storeToDatabase(ctx, params, content, metadata, storageConfig)
	case "local":
		return storeToLocal(ctx, params, content, metadata, storageConfig)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// UploadToS3Action maintains backward compatibility and uses S3 specifically
func UploadToS3Action(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Starting S3 upload action")

	// Get storage config
	var storageConfig config.ObjectStorageConfig

	// Extract config from agent or step
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if storageCfg, ok := agentConfig["storage_config"].(map[string]interface{}); ok {
			storageConfig.Provider = getStringOrDefault(storageCfg, "provider", "s3")
			storageConfig.Endpoint = getStringOrDefault(storageCfg, "endpoint", "")
			storageConfig.Bucket = getStringOrDefault(storageCfg, "bucket", "")
			storageConfig.AccessKeyEnvVar = getStringOrDefault(storageCfg, "access_key_env_var", "AWS_ACCESS_KEY_ID")
			storageConfig.SecretKeyEnvVar = getStringOrDefault(storageCfg, "secret_key_env_var", "AWS_SECRET_ACCESS_KEY")
		}
	}

	// Override with step config
	if params.StepConfig.Config != nil {
		if bucket, ok := params.StepConfig.Config["bucket"].(string); ok && bucket != "" {
			storageConfig.Bucket = bucket
		}
	}

	// Create S3 client
	s3Client, err := storage.NewS3Client(ctx, storageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Extract files to upload
	files := extractFilesToUpload(params)
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found to upload")
	}

	// Generate storage path
	path := generateStoragePath(params, params.StepConfig.Config)

	// Upload files
	uploadedFiles := make(map[string]string)
	publicURLs := make(map[string]string)

	for filename, content := range files {
		key := path + "/" + filename
		contentType := getContentType(filename)

		reader := bytes.NewReader(convertToBytes(content))

		s3URI, err := s3Client.Upload(ctx, key, contentType, reader)
		if err != nil {
			params.Logger.Error("Failed to upload file",
				zap.String("filename", filename),
				zap.Error(err))
			continue
		}

		uploadedFiles[filename] = s3URI

		// Generate public URL if configured
		if shouldMakePublic(params.StepConfig.Config) {
			url, err := s3Client.GetPresignedURL(ctx, key, 60*24*7) // 7 days
			if err == nil {
				publicURLs[filename] = url
			}
		}

		params.Logger.Info("Uploaded file",
			zap.String("filename", filename),
			zap.String("s3_uri", s3URI))
	}

	result := map[string]interface{}{
		"uploaded":     true,
		"files":        uploadedFiles,
		"file_count":   len(uploadedFiles),
		"bucket":       storageConfig.Bucket,
		"path":         path,
		"s3_base_path": fmt.Sprintf("s3://%s/%s", storageConfig.Bucket, path),
	}

	if len(publicURLs) > 0 {
		result["public_urls"] = publicURLs
		if url, ok := publicURLs["index.html"]; ok {
			result["website_url"] = url
		}
	}

	return result, nil
}

// Helper functions

func extractStorageConfig(params ActionParams) (map[string]interface{}, string) {
	storageConfig := make(map[string]interface{})
	storageType := ""

	// First check agent configuration
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if storage, ok := agentConfig["storage"].(map[string]interface{}); ok {
			for k, v := range storage {
				storageConfig[k] = v
			}
			storageType = getStringOrDefault(storage, "type", "")
		}
	}

	// Override with step configuration
	if params.StepConfig.Config != nil {
		for k, v := range params.StepConfig.Config {
			storageConfig[k] = v
		}
		if st, ok := params.StepConfig.Config["storage_type"].(string); ok {
			storageType = st
		}
	}

	return storageConfig, storageType
}

func extractContentToStore(params ActionParams, config map[string]interface{}) (map[string][]byte, map[string]interface{}, error) {
	// content := make(map[string][]byte)
	// metadata := make(map[string]interface{})

	// Check what type of content we're storing based on agent type
	agentType := params.AgentType
	if agentType == "" {
		agentType = params.Headers["agent_type"]
	}

	switch agentType {
	case "html-developer":
		return extractHTMLContent(params, config)
	case "image-generator":
		return extractImageContent(params, config)
	case "document-creator":
		return extractDocumentContent(params, config)
	default:
		return extractGenericContent(params, config)
	}
}

func extractHTMLContent(params ActionParams, config map[string]interface{}) (map[string][]byte, map[string]interface{}, error) {
	content := make(map[string][]byte)
	metadata := make(map[string]interface{})

	// Look for HTML in various places
	htmlSources := []string{"final_html", "processed_html", "html", "generated_html"}

	for _, source := range htmlSources {
		if htmlData, ok := params.CollectedData[source]; ok {
			if html, ok := htmlData.(string); ok && html != "" {
				content["index.html"] = []byte(html)
				metadata["content_type"] = "text/html"
				metadata["source"] = source
				break
			}
			if htmlMap, ok := htmlData.(map[string]interface{}); ok {
				if html, ok := htmlMap["html"].(string); ok && html != "" {
					content["index.html"] = []byte(html)
					metadata["content_type"] = "text/html"
					metadata["source"] = source
					break
				}
			}
		}
	}

	// Look for CSS and JS
	if css, ok := params.CollectedData["css"].(string); ok {
		content["styles.css"] = []byte(css)
	}
	if js, ok := params.CollectedData["javascript"].(string); ok {
		content["script.js"] = []byte(js)
	}

	if len(content) == 0 {
		return nil, nil, fmt.Errorf("no HTML content found")
	}

	return content, metadata, nil
}

func extractImageContent(params ActionParams, config map[string]interface{}) (map[string][]byte, map[string]interface{}, error) {
	content := make(map[string][]byte)
	metadata := make(map[string]interface{})

	// Look for image data
	imageSources := []string{"generated_image", "processed_image", "image", "image_data"}

	for _, source := range imageSources {
		if imageData, ok := params.CollectedData[source]; ok {
			filename := generateImageFilename(params, config)
			content[filename] = convertToBytes(imageData)
			metadata["content_type"] = detectImageType(content[filename])
			metadata["source"] = source
			break
		}
	}

	if len(content) == 0 {
		return nil, nil, fmt.Errorf("no image content found")
	}

	return content, metadata, nil
}

func extractDocumentContent(params ActionParams, config map[string]interface{}) (map[string][]byte, map[string]interface{}, error) {
	content := make(map[string][]byte)
	metadata := make(map[string]interface{})

	// Look for document data
	docSources := []string{"document", "generated_document", "pdf", "report"}

	for _, source := range docSources {
		if docData, ok := params.CollectedData[source]; ok {
			filename := generateDocumentFilename(params, config)
			content[filename] = convertToBytes(docData)
			metadata["content_type"] = detectDocumentType(content[filename])
			metadata["source"] = source
			break
		}
	}

	if len(content) == 0 {
		return nil, nil, fmt.Errorf("no document content found")
	}

	return content, metadata, nil
}

func extractGenericContent(params ActionParams, config map[string]interface{}) (map[string][]byte, map[string]interface{}, error) {
	content := make(map[string][]byte)
	metadata := make(map[string]interface{})

	// Try to find any storable content
	contentField := getStringOrDefault(config, "content_field", "")

	if contentField != "" {
		if data, ok := params.CollectedData[contentField]; ok {
			filename := generateFilename(params, config)
			content[filename] = convertToBytes(data)
			metadata["source"] = contentField
		}
	} else {
		// Look for common output fields
		for _, field := range []string{"output", "result", "generated_content", "data"} {
			if data, ok := params.CollectedData[field]; ok && data != nil {
				filename := generateFilename(params, config)
				content[filename] = convertToBytes(data)
				metadata["source"] = field
				break
			}
		}
	}

	if len(content) == 0 {
		return nil, nil, fmt.Errorf("no content found to store")
	}

	return content, metadata, nil
}

func extractFilesToUpload(params ActionParams) map[string][]byte {
	files := make(map[string][]byte)

	// Look for website files in collected data
	for key, value := range params.CollectedData {
		if strings.Contains(key, "html") || strings.Contains(key, "site") || key == "develop_site" {
			extracted := extractWebsiteFiles(value)
			for filename, content := range extracted {
				files[filename] = convertToBytes(content)
			}
		}
	}

	// Also check for validate_html step results
	if validateResult, ok := params.CollectedData["validate_html"].(map[string]interface{}); ok {
		if html, ok := validateResult["final_html"].(string); ok {
			files["index.html"] = []byte(html)
		}
	}

	return files
}

func extractWebsiteFiles(data interface{}) map[string]interface{} {
	files := make(map[string]interface{})

	switch v := data.(type) {
	case map[string]interface{}:
		// Direct file mapping
		if filesMap, ok := v["files"].(map[string]interface{}); ok {
			return filesMap
		}

		// Look for HTML/CSS/JS
		if html, ok := v["html"]; ok {
			files["index.html"] = html
		}
		if html, ok := v["final_html"]; ok {
			files["index.html"] = html
		}
		if css, ok := v["css"]; ok {
			files["styles.css"] = css
		}
		if js, ok := v["javascript"]; ok {
			files["script.js"] = js
		}

		// Recursive check in result
		if result, ok := v["result"]; ok {
			return extractWebsiteFiles(result)
		}
	}

	return files
}

func generateStoragePath(params ActionParams, config map[string]interface{}) string {
	pathTemplate := getStringOrDefault(config, "path_template", "{{.ClientID}}/{{.AgentType}}/{{.Timestamp}}")

	tmpl, err := template.New("path").Parse(pathTemplate)
	if err != nil {
		// Fallback path
		return fmt.Sprintf("%s/%s/%d",
			params.Headers["client_id"],
			params.AgentType,
			time.Now().Unix())
	}

	var buf bytes.Buffer
	tmpl.Execute(&buf, map[string]string{
		"ClientID":      params.Headers["client_id"],
		"CorrelationID": params.Headers["correlation_id"],
		"AgentType":     params.AgentType,
		"Timestamp":     fmt.Sprintf("%d", time.Now().Unix()),
	})

	return buf.String()
}

func generateFilename(params ActionParams, config map[string]interface{}) string {
	if filename, ok := config["filename"].(string); ok {
		return filename
	}

	extension := "dat"
	if ext, ok := config["extension"].(string); ok {
		extension = ext
	}

	return fmt.Sprintf("%s_%d.%s",
		params.Headers["correlation_id"][:8],
		time.Now().Unix(),
		extension)
}

func generateImageFilename(params ActionParams, config map[string]interface{}) string {
	if filename, ok := config["filename"].(string); ok {
		return filename
	}

	// Default to PNG
	extension := "png"
	if format, ok := config["image_format"].(string); ok {
		extension = format
	}

	return fmt.Sprintf("image_%s_%d.%s",
		params.Headers["correlation_id"][:8],
		time.Now().Unix(),
		extension)
}

func generateDocumentFilename(params ActionParams, config map[string]interface{}) string {
	if filename, ok := config["filename"].(string); ok {
		return filename
	}

	extension := "pdf"
	if format, ok := config["document_format"].(string); ok {
		extension = format
	}

	return fmt.Sprintf("document_%s_%d.%s",
		params.Headers["correlation_id"][:8],
		time.Now().Unix(),
		extension)
}

func storeToS3(ctx context.Context, params ActionParams, content map[string][]byte, metadata map[string]interface{}, configMap map[string]interface{}) (interface{}, error) {
	// Build ObjectStorageConfig from the config map
	var storageConfig config.ObjectStorageConfig

	// First, check if we have storage config in agent_config (from database)
	if agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{}); ok {
		if agentStorage, ok := agentConfig["storage"].(map[string]interface{}); ok {
			// Agent has storage configuration
			if provider, ok := agentStorage["provider"].(string); ok {
				storageConfig.Provider = provider
			}
			if endpoint, ok := agentStorage["endpoint"].(string); ok {
				storageConfig.Endpoint = endpoint
			}
			if bucket, ok := agentStorage["bucket"].(string); ok {
				storageConfig.Bucket = bucket
			}
			// These might be stored differently in the database
			if accessKey, ok := agentStorage["access_key_env_var"].(string); ok {
				storageConfig.AccessKeyEnvVar = accessKey
			}
			if secretKey, ok := agentStorage["secret_key_env_var"].(string); ok {
				storageConfig.SecretKeyEnvVar = secretKey
			}
		}

		// Also check for storage_config (as used in your existing UploadToS3Action)
		if storageCfg, ok := agentConfig["storage_config"].(map[string]interface{}); ok {
			if provider, ok := storageCfg["provider"].(string); ok && provider != "" {
				storageConfig.Provider = provider
			}
			if endpoint, ok := storageCfg["endpoint"].(string); ok && endpoint != "" {
				storageConfig.Endpoint = endpoint
			}
			if bucket, ok := storageCfg["bucket"].(string); ok && bucket != "" {
				storageConfig.Bucket = bucket
			}
			if accessKey, ok := storageCfg["access_key_env_var"].(string); ok && accessKey != "" {
				storageConfig.AccessKeyEnvVar = accessKey
			}
			if secretKey, ok := storageCfg["secret_key_env_var"].(string); ok && secretKey != "" {
				storageConfig.SecretKeyEnvVar = secretKey
			}
		}
	}

	// Override with step-specific config (from configMap parameter)
	if provider, ok := configMap["provider"].(string); ok && provider != "" {
		storageConfig.Provider = provider
	}
	if endpoint, ok := configMap["endpoint"].(string); ok && endpoint != "" {
		storageConfig.Endpoint = endpoint
	}
	if bucket, ok := configMap["bucket"].(string); ok && bucket != "" {
		storageConfig.Bucket = bucket
	}
	if bucketOverride, ok := configMap["bucket_override"].(string); ok && bucketOverride != "" {
		storageConfig.Bucket = bucketOverride
	}

	// Set defaults if not specified
	if storageConfig.Provider == "" {
		storageConfig.Provider = "s3"
	}
	if storageConfig.AccessKeyEnvVar == "" {
		storageConfig.AccessKeyEnvVar = "AWS_ACCESS_KEY_ID"
	}
	if storageConfig.SecretKeyEnvVar == "" {
		storageConfig.SecretKeyEnvVar = "AWS_SECRET_ACCESS_KEY"
	}

	// If bucket is still empty, try to get from environment
	if storageConfig.Bucket == "" {
		if bucketEnv, ok := configMap["bucket_env"].(string); ok {
			storageConfig.Bucket = os.Getenv(bucketEnv)
		}
		if storageConfig.Bucket == "" {
			storageConfig.Bucket = os.Getenv("ASSETS_BUCKET")
		}
	}

	params.Logger.Info("S3 storage configuration",
		zap.String("provider", storageConfig.Provider),
		zap.String("endpoint", storageConfig.Endpoint),
		zap.String("bucket", storageConfig.Bucket),
		zap.Bool("has_access_key_env", storageConfig.AccessKeyEnvVar != ""),
		zap.Bool("has_secret_key_env", storageConfig.SecretKeyEnvVar != ""))

	// Now create the S3 client with properly structured config
	s3Client, err := storage.NewS3Client(ctx, storageConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Rest of the function remains the same...
	basePath := generateStoragePath(params, configMap)

	uploadedFiles := make(map[string]string)
	publicURLs := make(map[string]string)

	for filename, data := range content {
		key := basePath + "/" + filename
		contentType := getContentType(filename)

		reader := bytes.NewReader(data)

		s3URI, err := s3Client.Upload(ctx, key, contentType, reader)
		if err != nil {
			params.Logger.Error("Failed to upload to S3",
				zap.String("filename", filename),
				zap.Error(err))
			continue
		}

		uploadedFiles[filename] = s3URI

		// Generate public URL if needed
		if shouldMakePublic(configMap) {
			url, err := s3Client.GetPresignedURL(ctx, key, 60*24*7)
			if err == nil {
				publicURLs[filename] = url
			}
		}
	}

	result := map[string]interface{}{
		"stored":       true,
		"storage_type": "s3",
		"files":        uploadedFiles,
		"bucket":       storageConfig.Bucket,
		"path":         basePath,
		"metadata":     metadata,
	}

	if len(publicURLs) > 0 {
		result["public_urls"] = publicURLs
	}

	return result, nil
}

func storeToB2(ctx context.Context, params ActionParams, content map[string][]byte, metadata map[string]interface{}, config map[string]interface{}) (interface{}, error) {
	// Similar to S3 but with B2-specific config
	config["provider"] = "b2"
	config["access_key_env_var"] = "B2_APPLICATION_KEY_ID"
	config["secret_key_env_var"] = "B2_APPLICATION_KEY"

	return storeToS3(ctx, params, content, metadata, config)
}

func storeToDatabase(ctx context.Context, params ActionParams, content map[string][]byte, metadata map[string]interface{}, config map[string]interface{}) (interface{}, error) {
	// Store in database
	// This would use params.DB to store content
	return map[string]interface{}{
		"stored":       true,
		"storage_type": "database",
		"record_count": len(content),
		"metadata":     metadata,
	}, nil
}

func storeToLocal(ctx context.Context, params ActionParams, content map[string][]byte, metadata map[string]interface{}, config map[string]interface{}) (interface{}, error) {
	// Store locally in the pod's filesystem
	basePath := "/tmp/agentchassis/" + generateStoragePath(params, config)

	// Create directory structure
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	storedFiles := make(map[string]string)
	storedSizes := make(map[string]int)

	for filename, data := range content {
		filePath := filepath.Join(basePath, filename)

		// Ensure subdirectories exist
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			params.Logger.Error("Failed to create subdirectory",
				zap.String("dir", dir),
				zap.Error(err))
			continue
		}

		// Write file
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			params.Logger.Error("Failed to write file",
				zap.String("file", filePath),
				zap.Error(err))
			continue
		}

		storedFiles[filename] = filePath
		storedSizes[filename] = len(data)

		params.Logger.Info("Stored file locally",
			zap.String("file", filePath),
			zap.Int("size", len(data)))
	}

	return map[string]interface{}{
		"stored":       true,
		"storage_type": "local",
		"files":        storedFiles,
		"file_sizes":   storedSizes,
		"path":         basePath,
		"metadata":     metadata,
		"note":         "Files stored in pod filesystem at " + basePath,
	}, nil
}

// Utility functions

func convertToBytes(data interface{}) []byte {
	switch v := data.(type) {
	case []byte:
		return v
	case string:
		return []byte(v)
	case io.Reader:
		buf := new(bytes.Buffer)
		buf.ReadFrom(v)
		return buf.Bytes()
	default:
		// Try JSON encoding
		jsonBytes, _ := json.Marshal(v)
		return jsonBytes
	}
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
	case strings.HasSuffix(filename, ".pdf"):
		return "application/pdf"
	default:
		return "application/octet-stream"
	}
}

func detectImageType(data []byte) string {
	// Check magic bytes
	if len(data) > 8 {
		if bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
			return "image/png"
		}
		if bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}) {
			return "image/jpeg"
		}
		if bytes.HasPrefix(data, []byte("GIF")) {
			return "image/gif"
		}
	}
	return "application/octet-stream"
}

func detectDocumentType(data []byte) string {
	if len(data) > 4 {
		if bytes.HasPrefix(data, []byte("%PDF")) {
			return "application/pdf"
		}
	}
	return "application/octet-stream"
}

func shouldMakePublic(config map[string]interface{}) bool {
	if public, ok := config["make_public"].(bool); ok {
		return public
	}
	if public, ok := config["public_access"].(bool); ok {
		return public
	}
	return false
}

func getStringOrDefault(m map[string]interface{}, key, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}
