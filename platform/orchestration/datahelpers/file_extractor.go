// FILE: datahelpers/file_extractor.go
// Unified file extraction helper that supports multiple patterns
//
// This consolidates logic from:
// - extractFilesFromParams (registry/actions)
// - extractFilesForGit (git_deployer_actions)
// - extractFilesToUpload (cf_deploy_actions)
// - extractWebsiteFiles (cf_deploy_actions)

package datahelpers

import (
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// FileExtractionConfig controls how files are extracted
type FileExtractionConfig struct {
	// Primary extraction methods (tried in order)
	FilesField   string // Path to map of filename->content (e.g., "site_files.files")
	ContentField string // Path to single content string (saved as filename from PageFrom or "index.html")
	HTMLFrom     string // Path to HTML content string (e.g., "assembled_page.html")
	PageFrom     string // Path to page data for filename (e.g., "current_page")

	// Direct files embedded in config
	DirectFiles map[string]interface{}

	// Fallback behavior
	DefaultFilesField string // Fallback if FilesField not set (e.g., "site_files.files")
	DefaultFilename   string // Default filename for single content (default: "index.html")

	// Search behavior
	SearchForHTMLKeys bool // Scan collected data for html/site keys
}

// FileExtractionResult contains extracted files and metadata
type FileExtractionResult struct {
	Files       map[string]string // filename -> content
	Method      string            // which method succeeded
	SourceField string            // which field the content came from
}

// ExtractFiles extracts files from collected data using multiple strategies
// This is the main entry point that tries methods in priority order
func ExtractFiles(data map[string]interface{}, cfg FileExtractionConfig, logger *zap.Logger) *FileExtractionResult {
	result := &FileExtractionResult{
		Files: make(map[string]string),
	}

	// Set defaults
	if cfg.DefaultFilename == "" {
		cfg.DefaultFilename = "index.html"
	}

	// Method 1: files_field - map of filename -> content
	if cfg.FilesField != "" {
		if files := extractFilesMap(data, cfg.FilesField, logger); len(files) > 0 {
			result.Files = files
			result.Method = "files_field"
			result.SourceField = cfg.FilesField
			logger.Debug("ExtractFiles: found via files_field",
				zap.String("field", cfg.FilesField),
				zap.Int("count", len(files)))
			return result
		}
	}

	// Method 2: Direct files in config
	if len(cfg.DirectFiles) > 0 {
		for filename, content := range cfg.DirectFiles {
			if contentStr, ok := content.(string); ok {
				result.Files[filename] = contentStr
			}
		}
		if len(result.Files) > 0 {
			result.Method = "direct_files"
			logger.Debug("ExtractFiles: found via direct_files",
				zap.Int("count", len(result.Files)))
			return result
		}
	}

	// Method 3: html_from + page_from (loop-based single page)
	if cfg.HTMLFrom != "" {
		htmlContent := ExtractNestedFieldString(data, cfg.HTMLFrom)
		if htmlContent != "" {
			filename := determineFilename(data, cfg.PageFrom, cfg.DefaultFilename, logger)
			result.Files[filename] = htmlContent
			result.Method = "html_from"
			result.SourceField = cfg.HTMLFrom
			logger.Debug("ExtractFiles: found via html_from",
				zap.String("field", cfg.HTMLFrom),
				zap.String("filename", filename))
			return result
		}
	}

	// Method 4: content_field - single content string
	if cfg.ContentField != "" {
		content := ExtractNestedFieldString(data, cfg.ContentField)
		if content != "" {
			filename := determineFilename(data, cfg.PageFrom, cfg.DefaultFilename, logger)
			result.Files[filename] = content
			result.Method = "content_field"
			result.SourceField = cfg.ContentField
			logger.Debug("ExtractFiles: found via content_field",
				zap.String("field", cfg.ContentField),
				zap.String("filename", filename))
			return result
		}
	}

	// Method 5: Default files_field fallback
	if cfg.DefaultFilesField != "" && cfg.FilesField != cfg.DefaultFilesField {
		if files := extractFilesMap(data, cfg.DefaultFilesField, logger); len(files) > 0 {
			result.Files = files
			result.Method = "default_files_field"
			result.SourceField = cfg.DefaultFilesField
			logger.Debug("ExtractFiles: found via default_files_field",
				zap.String("field", cfg.DefaultFilesField),
				zap.Int("count", len(files)))
			return result
		}
	}

	// Method 6: Search for common HTML keys
	if cfg.SearchForHTMLKeys {
		if files := searchForHTMLContent(data, logger); len(files) > 0 {
			result.Files = files
			result.Method = "html_search"
			logger.Debug("ExtractFiles: found via html_search",
				zap.Int("count", len(files)))
			return result
		}
	}

	logger.Warn("ExtractFiles: no files found",
		zap.String("files_field", cfg.FilesField),
		zap.String("html_from", cfg.HTMLFrom),
		zap.String("content_field", cfg.ContentField))

	return result
}

// ExtractFilesFromConfig creates config from a typical action config map
func ExtractFilesFromConfig(actionConfig map[string]interface{}) FileExtractionConfig {
	cfg := FileExtractionConfig{}

	if v, ok := actionConfig["files_field"].(string); ok {
		cfg.FilesField = v
	}
	if v, ok := actionConfig["content_field"].(string); ok {
		cfg.ContentField = v
	}
	if v, ok := actionConfig["html_from"].(string); ok {
		cfg.HTMLFrom = v
	}
	if v, ok := actionConfig["page_from"].(string); ok {
		cfg.PageFrom = v
	}
	if v, ok := actionConfig["files"].(map[string]interface{}); ok {
		cfg.DirectFiles = v
	}

	return cfg
}

// extractFilesMap extracts a map of filename->content from a field path
func extractFilesMap(data map[string]interface{}, fieldPath string, logger *zap.Logger) map[string]string {
	files := make(map[string]string)

	val := ExtractNestedField(data, fieldPath)
	if val == nil {
		return files
	}

	switch v := val.(type) {
	case map[string]interface{}:
		for filename, content := range v {
			if contentStr, ok := content.(string); ok && contentStr != "" {
				files[filename] = contentStr
			}
		}
	case map[string]string:
		return v
	}

	return files
}

// determineFilename determines the output filename from page data
func determineFilename(data map[string]interface{}, pageFrom string, defaultName string, logger *zap.Logger) string {
	if pageFrom == "" {
		return defaultName
	}

	pageData := ExtractNestedField(data, pageFrom)
	if pageData == nil {
		return defaultName
	}

	pageMap, ok := pageData.(map[string]interface{})
	if !ok {
		return defaultName
	}

	// Try url field first (e.g., "/index.html" -> "index.html")
	if url, ok := pageMap["url"].(string); ok && url != "" {
		filename := strings.TrimPrefix(url, "/")
		if !strings.HasSuffix(filename, ".html") {
			filename = filename + ".html"
		}
		return filepath.Clean(filename)
	}

	// Fallback to name field (e.g., "index" -> "index.html")
	if name, ok := pageMap["name"].(string); ok && name != "" {
		filename := name
		if !strings.HasSuffix(filename, ".html") {
			filename = filename + ".html"
		}
		return filepath.Clean(filename)
	}

	// Fallback to slug field
	if slug, ok := pageMap["slug"].(string); ok && slug != "" {
		filename := slug
		if !strings.HasSuffix(filename, ".html") {
			filename = filename + ".html"
		}
		return filepath.Clean(filename)
	}

	return defaultName
}

// searchForHTMLContent searches collected data for common HTML content patterns
func searchForHTMLContent(data map[string]interface{}, logger *zap.Logger) map[string]string {
	files := make(map[string]string)

	for key, value := range data {
		// Skip system keys
		if strings.HasPrefix(key, "__") {
			continue
		}

		// Try to extract website files from this value
		extracted := extractWebsiteFiles(value)
		if len(extracted) > 0 {
			for filename, content := range extracted {
				files[filename] = content
			}
			// If we found a files map or HTML, return it
			if len(files) > 0 {
				return files
			}
		}
	}

	return files
}

// extractWebsiteFiles recursively extracts website files from nested data structures
// Supports: files map, html, final_html, page_html, css, javascript
// Recursively unwraps "result" fields
func extractWebsiteFiles(data interface{}) map[string]string {
	files := make(map[string]string)

	switch v := data.(type) {
	case map[string]interface{}:
		// Direct file mapping - if we find a "files" map, return it directly
		if filesMap, ok := v["files"].(map[string]interface{}); ok {
			for filename, content := range filesMap {
				if contentStr, ok := content.(string); ok {
					files[filename] = contentStr
				}
			}
			if len(files) > 0 {
				return files
			}
		}

		// Look for HTML content (multiple possible keys)
		if html, ok := v["html"].(string); ok && html != "" {
			files["index.html"] = html
		}
		if html, ok := v["final_html"].(string); ok && html != "" {
			files["index.html"] = html
		}
		if html, ok := v["page_html"].(string); ok && html != "" {
			files["index.html"] = html
		}

		// Look for CSS
		if css, ok := v["css"].(string); ok && css != "" {
			files["styles.css"] = css
		}

		// Look for JavaScript
		if js, ok := v["javascript"].(string); ok && js != "" {
			files["script.js"] = js
		}
		if js, ok := v["js"].(string); ok && js != "" {
			files["script.js"] = js
		}

		// If we found files at this level, return them
		if len(files) > 0 {
			return files
		}

		// Recursive check in "result" field - common wrapper pattern
		if result, ok := v["result"]; ok {
			return extractWebsiteFiles(result)
		}

		// Also check "body" field - another common wrapper
		if body, ok := v["body"]; ok {
			return extractWebsiteFiles(body)
		}

		// Check "response" field
		if response, ok := v["response"]; ok {
			return extractWebsiteFiles(response)
		}
	}

	return files
}

// ExtractFilesAsBytes is a convenience wrapper that returns []byte values
// Use with SearchForHTMLKeys: true to get the recursive extractWebsiteFiles behavior
func ExtractFilesAsBytes(data map[string]interface{}, cfg FileExtractionConfig, logger *zap.Logger) map[string][]byte {
	result := ExtractFiles(data, cfg, logger)
	bytesMap := make(map[string][]byte, len(result.Files))
	for filename, content := range result.Files {
		bytesMap[filename] = []byte(content)
	}
	return bytesMap
}

// ExtractWebsiteFilesFromValue is exported for direct use when you have a single value
// to extract from (not searching across collected data)
func ExtractWebsiteFilesFromValue(data interface{}) map[string]string {
	return extractWebsiteFiles(data)
}
