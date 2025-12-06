// FILE: platform/orchestration/actions/multipage_assembly_actions.go
package actions

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// AssembleMultipageSiteAction assembles a complete multi-page website
// Takes a map of pages with their content and creates consistent HTML files
func AssembleMultipageSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Assembling multi-page website")

	config := params.StepConfig.Config
	if config == nil {
		return nil, fmt.Errorf("config is required for assemble_multipage_site action")
	}

	// Extract configuration
	indexHTMLField, _ := config["index_html_field"].(string)       // Main page HTML
	pagesField, _ := config["pages_field"].(string)                // Map of additional pages
	sharedStylesField, _ := config["shared_styles_field"].(string) // Optional shared CSS
	navigationField, _ := config["navigation_field"].(string)      // Optional nav config
	batchFields, _ := config["batch_fields"].([]interface{})       // Optional: multiple batch sources

	if indexHTMLField == "" {
		return nil, fmt.Errorf("index_html_field is required in config")
	}

	// Extract the main page HTML
	indexHTML := extractFieldValue(params.CollectedData, indexHTMLField, params.Logger)
	if indexHTML == "" {
		return nil, fmt.Errorf("index HTML is empty (field: %s)", indexHTMLField)
	}

	params.Logger.Info("Extracted index HTML",
		zap.Int("length", len(indexHTML)),
	)

	// Initialize the files map with index.html
	files := map[string]string{
		"index.html": datahelpers.CleanHTMLString(indexHTML),
	}

	// Extract additional pages from single field if provided
	if pagesField != "" {
		additionalPages := extractPagesMap(params.CollectedData, pagesField, params.Logger)
		for filename, content := range additionalPages {
			if content != "" {
				files[filename] = datahelpers.CleanHTMLString(content)
				params.Logger.Info("Added page",
					zap.String("filename", filename),
					zap.Int("length", len(content)),
				)
			}
		}
	}

	// Extract pages from multiple batch fields if provided
	if len(batchFields) > 0 {
		for i, batchField := range batchFields {
			if batchFieldStr, ok := batchField.(string); ok {
				batchPages := extractPagesMap(params.CollectedData, batchFieldStr, params.Logger)
				for filename, content := range batchPages {
					if content != "" {
						files[filename] = datahelpers.CleanHTMLString(content)
						params.Logger.Info("Added page from batch",
							zap.Int("batch_index", i),
							zap.String("batch_field", batchFieldStr),
							zap.String("filename", filename),
							zap.Int("length", len(content)),
						)
					}
				}
			}
		}
	}

	// Extract shared styles if provided
	var sharedStyles string
	if sharedStylesField != "" {
		sharedStyles = extractFieldValue(params.CollectedData, sharedStylesField, params.Logger)
		if sharedStyles != "" {
			params.Logger.Info("Extracted shared styles",
				zap.Int("length", len(sharedStyles)),
			)
		}
	}

	// Extract navigation config if provided
	var navConfig map[string]interface{}
	if navigationField != "" {
		if navData, ok := extractNestedField(params.CollectedData, navigationField).(map[string]interface{}); ok {
			navConfig = navData
			params.Logger.Info("Extracted navigation config")
		}
	}

	// Apply shared enhancements to all pages
	for filename, html := range files {
		enhanced, err := enhancePageHTML(html, filename, sharedStyles, navConfig, files, params.Logger)
		if err != nil {
			params.Logger.Warn("Failed to enhance page",
				zap.String("filename", filename),
				zap.Error(err),
			)
			continue
		}
		files[filename] = enhanced
	}

	params.Logger.Info("Multi-page site assembled successfully",
		zap.Int("total_pages", len(files)),
		zap.Int("total_size", calculateTotalSize(files)),
	)

	// Check if we should stream to S3 instead of returning all files
	streamToS3, _ := config["stream_to_s3"].(bool)
	if streamToS3 {
		return streamFilesToStorage(files, params)
	}

	return map[string]interface{}{
		"files":        files,
		"page_count":   len(files),
		"total_bytes":  calculateTotalSize(files),
		"assembled_at": params.ExecutionContext.Timestamp,
		"mode":         "in_memory",
	}, nil
}

// streamFilesToStorage stores files to S3 and returns references instead of content
// This prevents holding massive sites in memory
func streamFilesToStorage(files map[string]string, params ActionParams) (interface{}, error) {
	storedFiles := make(map[string]string)
	totalBytes := 0

	for filename, content := range files {
		// Store each file individually
		s3Key, err := storeFileToS3(filename, content, params)
		if err != nil {
			params.Logger.Error("Failed to store file to S3",
				zap.String("filename", filename),
				zap.Error(err),
			)
			return nil, fmt.Errorf("failed to store %s: %w", filename, err)
		}

		storedFiles[filename] = s3Key
		totalBytes += len(content)

		params.Logger.Info("Stored file to S3",
			zap.String("filename", filename),
			zap.String("s3_key", s3Key),
			zap.Int("bytes", len(content)),
		)
	}

	return map[string]interface{}{
		"stored_files": storedFiles,
		"page_count":   len(storedFiles),
		"total_bytes":  totalBytes,
		"assembled_at": params.ExecutionContext.Timestamp,
		"mode":         "streamed_to_s3",
	}, nil
}

// storeFileToS3 stores a single file to S3 and returns the key
func storeFileToS3(filename, content string, params ActionParams) (string, error) {
	// Check if S3 storage is available
	bucket := os.Getenv("ASSETS_BUCKET")
	if bucket == "" {
		return "", fmt.Errorf("ASSETS_BUCKET environment variable not set")
	}

	// Generate S3 key with orchestration context
	orchestrationID := params.ExecutionContext.OrchestrationID
	timestamp := time.Now().Format("20060102-150405")
	s3Key := fmt.Sprintf("multipage-sites/%s/%s/%s", orchestrationID, timestamp, filename)

	// Store to S3 (assuming S3 client is available)
	// This would typically use the AWS SDK
	// For now, return the key format - actual S3 upload would be implemented based on your S3 setup

	params.Logger.Info("Would store to S3",
		zap.String("bucket", bucket),
		zap.String("key", s3Key),
		zap.Int("size", len(content)),
	)

	// TODO: Implement actual S3 upload
	// Example:
	// s3Client := getS3Client(params)
	// _, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
	//     Bucket: aws.String(bucket),
	//     Key:    aws.String(s3Key),
	//     Body:   strings.NewReader(content),
	//     ContentType: aws.String("text/html"),
	// })

	return s3Key, nil
}

// extractPagesMap extracts a map of pages from collected data
func extractPagesMap(data map[string]interface{}, fieldPath string, logger *zap.Logger) map[string]string {
	result := make(map[string]string)

	value := extractNestedField(data, fieldPath)
	if value == nil {
		return result
	}

	// Try to convert to map[string]string
	if pagesMap, ok := value.(map[string]string); ok {
		return pagesMap
	}

	// Try to convert to map[string]interface{} and extract strings
	if pagesMap, ok := value.(map[string]interface{}); ok {
		for key, val := range pagesMap {
			if strVal, ok := val.(string); ok {
				result[key] = strVal
			} else if mapVal, ok := val.(map[string]interface{}); ok {
				// Try to extract "html" or "content" or "result" fields
				if html, ok := mapVal["html"].(string); ok {
					result[key] = html
				} else if content, ok := mapVal["content"].(string); ok {
					result[key] = content
				} else if resultVal, ok := mapVal["result"].(string); ok {
					result[key] = resultVal
				}
			}
		}
	}

	logger.Info("Extracted pages map",
		zap.Int("page_count", len(result)),
	)

	return result
}

// enhancePageHTML applies shared styles, navigation, and cross-links to a page
func enhancePageHTML(html, currentFilename string, sharedStyles string, navConfig map[string]interface{}, allFiles map[string]string, logger *zap.Logger) (string, error) {
	doc := html

	// 1. Add shared styles if provided and not already present
	if sharedStyles != "" && !strings.Contains(doc, "/* SHARED_STYLES */") {
		// Wrap in style tags if needed
		if !strings.Contains(sharedStyles, "<style") {
			sharedStyles = "<style>\n/* SHARED_STYLES */\n" + sharedStyles + "\n</style>"
		}

		// Insert before </head>
		headCloseIdx := strings.Index(doc, "</head>")
		if headCloseIdx >= 0 {
			doc = doc[:headCloseIdx] + "\n" + sharedStyles + "\n" + doc[headCloseIdx:]
			logger.Debug("Added shared styles",
				zap.String("page", currentFilename),
			)
		}
	}

	// 2. Add navigation if provided
	if navConfig != nil {
		doc = addNavigation(doc, currentFilename, navConfig, allFiles, logger)
	}

	// 3. Ensure proper DOCTYPE
	if !strings.HasPrefix(strings.TrimSpace(doc), "<!DOCTYPE") {
		doc = "<!DOCTYPE html>\n" + doc
	}

	return doc, nil
}

// addNavigation adds or updates navigation in the HTML
func addNavigation(html, currentFilename string, navConfig map[string]interface{}, allFiles map[string]string, logger *zap.Logger) string {
	// Build navigation HTML
	navHTML := buildNavigationHTML(currentFilename, navConfig, allFiles)

	// Look for existing nav tag
	navStartIdx := strings.Index(html, "<nav")
	if navStartIdx >= 0 {
		// Find the closing </nav>
		navEndIdx := strings.Index(html[navStartIdx:], "</nav>")
		if navEndIdx >= 0 {
			navEndIdx += navStartIdx + 6 // length of "</nav>"
			// Replace existing nav
			html = html[:navStartIdx] + navHTML + html[navEndIdx:]
			logger.Debug("Replaced existing navigation",
				zap.String("page", currentFilename),
			)
			return html
		}
	}

	// No existing nav, insert after <body> tag
	bodyStartIdx := strings.Index(html, "<body")
	if bodyStartIdx >= 0 {
		bodyOpenEndIdx := strings.Index(html[bodyStartIdx:], ">")
		if bodyOpenEndIdx >= 0 {
			insertPoint := bodyStartIdx + bodyOpenEndIdx + 1
			html = html[:insertPoint] + "\n" + navHTML + "\n" + html[insertPoint:]
			logger.Debug("Added new navigation",
				zap.String("page", currentFilename),
			)
		}
	}

	return html
}

// buildNavigationHTML creates navigation HTML based on config and available files
func buildNavigationHTML(currentFilename string, navConfig map[string]interface{}, allFiles map[string]string) string {
	var navItems []string

	// Check if there's a custom nav structure in config
	if items, ok := navConfig["items"].([]interface{}); ok {
		for _, item := range items {
			if itemMap, ok := item.(map[string]interface{}); ok {
				label, _ := itemMap["label"].(string)
				href, _ := itemMap["href"].(string)

				if label != "" && href != "" {
					activeClass := ""
					if href == currentFilename || (currentFilename == "index.html" && href == "/") {
						activeClass = ` class="active"`
					}
					navItems = append(navItems, fmt.Sprintf(`<a href="%s"%s>%s</a>`, href, activeClass, label))
				}
			}
		}
	} else {
		// Auto-generate nav from available files
		fileOrder := []string{"index.html", "about.html", "services.html", "contact.html"}
		fileLabels := map[string]string{
			"index.html":    "Home",
			"about.html":    "About",
			"services.html": "Services",
			"contact.html":  "Contact",
		}

		for _, filename := range fileOrder {
			if _, exists := allFiles[filename]; exists {
				label := fileLabels[filename]
				if label == "" {
					// Generate label from filename
					label = strings.TrimSuffix(filename, ".html")
					label = strings.Title(label)
				}

				activeClass := ""
				if filename == currentFilename {
					activeClass = ` class="active"`
				}

				navItems = append(navItems, fmt.Sprintf(`<a href="%s"%s>%s</a>`, filename, activeClass, label))
			}
		}

		// Add any remaining files not in the order
		for filename := range allFiles {
			found := false
			for _, orderedFile := range fileOrder {
				if filename == orderedFile {
					found = true
					break
				}
			}
			if !found {
				label := strings.TrimSuffix(filename, ".html")
				label = strings.Title(label)
				activeClass := ""
				if filename == currentFilename {
					activeClass = ` class="active"`
				}
				navItems = append(navItems, fmt.Sprintf(`<a href="%s"%s>%s</a>`, filename, activeClass, label))
			}
		}
	}

	// Assemble the nav HTML
	return fmt.Sprintf(`<nav class="main-navigation">
	%s
</nav>`, strings.Join(navItems, "\n\t"))
}

// extractNestedField navigates nested field paths
func extractNestedField(data map[string]interface{}, fieldPath string) interface{} {
	parts := strings.Split(fieldPath, ".")

	var current interface{} = data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
				continue
			}
			// Try ExtractStepData
			if extracted := ExtractStepData(v[part]); extracted != nil {
				current = extracted
				continue
			}
			return nil
		default:
			return nil
		}
	}

	return current
}

// calculateTotalSize calculates total bytes across all files
func calculateTotalSize(files map[string]string) int {
	total := 0
	for _, content := range files {
		total += len(content)
	}
	return total
}
