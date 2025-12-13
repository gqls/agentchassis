// FILE: platform/orchestration/actions/multipage_actions.go
package actions

import (
	"context"
	"fmt"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// AssemblePageAction builds a single complete HTML page
// Takes HTML content and ensures it's a valid, complete page
func AssemblePageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Assembling single page")

	config := params.StepConfig.Config
	if config == nil {
		return nil, fmt.Errorf("config is required for assemble_page action")
	}

	contentField, _ := config["content_field"].(string)
	if contentField == "" {
		return nil, fmt.Errorf("content_field is required in config")
	}

	addNav, _ := config["add_navigation"].(bool)

	// Extract content
	content := extractFieldValue(params.CollectedData, contentField, params.Logger)
	if content == "" {
		return nil, fmt.Errorf("no content found at %s", contentField)
	}

	params.Logger.Info("Extracted content",
		zap.String("field", contentField),
		zap.Int("length", len(content)),
	)

	// Clean HTML (remove markdown code blocks, etc.)
	html := datahelpers.CleanHTMLString(content)

	// Ensure valid HTML structure
	html = ensureValidHTML(html)

	// Add navigation if requested
	if addNav {
		html = addSimpleNavigation(html, "index")
	}

	params.Logger.Info("Page assembled successfully",
		zap.Int("final_length", len(html)),
		zap.Bool("added_navigation", addNav),
	)

	return map[string]interface{}{
		"html":         html,
		"assembled_at": params.ExecutionContext.Timestamp,
	}, nil
}

// AssembleMultipageSiteAction creates a complete multi-page site
// Takes pages from loop output, adds navigation, generates standard pages
func AssembleMultipageSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Assembling multi-page site")

	config := params.StepConfig.Config
	if config == nil {
		return nil, fmt.Errorf("config is required for assemble_multipage_site action")
	}

	pagesField, _ := config["pages_field"].(string)
	if pagesField == "" {
		return nil, fmt.Errorf("pages_field is required in config")
	}

	addNav, _ := config["add_navigation"].(bool)
	generateStandard, _ := config["generate_standard_pages"].(bool)

	// Extract pages from loop output
	pages := extractPagesFromLoop(params.CollectedData, pagesField, params.Logger)

	if len(pages) == 0 {
		return nil, fmt.Errorf("no pages found at %s", pagesField)
	}

	params.Logger.Info("Extracted pages from loop",
		zap.Int("page_count", len(pages)),
		zap.Strings("page_names", getPageNames(pages)),
	)

	// Generate standard pages if requested and missing
	if generateStandard {
		domain := extractDomainFromData(params.CollectedData)

		if _, hasAbout := pages["about.html"]; !hasAbout {
			pages["about.html"] = generateAboutPage(domain)
			params.Logger.Info("Generated about page")
		}

		if _, hasContact := pages["contact.html"]; !hasContact {
			pages["contact.html"] = generateContactPage(domain)
			params.Logger.Info("Generated contact page")
		}
	}

	// Add navigation to all pages if requested
	if addNav {
		pages = addNavigationToAllPages(pages, params.Logger)
		params.Logger.Info("Added navigation to all pages")
	}

	// Ensure all pages are valid HTML
	for name, html := range pages {
		pages[name] = ensureValidHTML(html)
	}

	params.Logger.Info("Multi-page site assembled successfully",
		zap.Int("total_pages", len(pages)),
		zap.Int("total_bytes", calculateTotalSize(pages)),
	)

	return map[string]interface{}{
		"files":        pages,
		"page_count":   len(pages),
		"total_bytes":  calculateTotalSize(pages),
		"page_names":   getPageNames(pages),
		"assembled_at": params.ExecutionContext.Timestamp,
	}, nil
}

// ============================================================================
// Helper Functions
// ============================================================================

// extractPagesFromLoop handles different loop output formats
func extractPagesFromLoop(data map[string]interface{}, fieldPath string, logger *zap.Logger) map[string]string {
	pages := make(map[string]string)

	value := extractNestedField(data, fieldPath)
	if value == nil {
		logger.Warn("No value found at field path", zap.String("path", fieldPath))
		return pages
	}

	logger.Info("Extracting pages from loop output",
		zap.String("value_type", fmt.Sprintf("%T", value)),
	)

	// Format 1: Map of pages {"index": "...", "about": "...", ...}
	if pagesMap, ok := value.(map[string]interface{}); ok {
		for name, content := range pagesMap {
			html := extractHTMLFromValue(content, logger)
			if html != "" {
				filename := name
				if !strings.HasSuffix(filename, ".html") {
					filename = filename + ".html"
				}
				pages[filename] = html
				logger.Debug("Extracted page from map",
					zap.String("name", filename),
					zap.Int("length", len(html)),
				)
			}
		}
		return pages
	}

	// Format 2: Array of page objects [{"name": "index", "page_html": "..."}, ...]
	if pagesArray, ok := value.([]interface{}); ok {
		for i, item := range pagesArray {
			if itemMap, ok := item.(map[string]interface{}); ok {
				name := fmt.Sprintf("page_%d", i)

				// Try to get name from item
				if itemName, ok := itemMap["name"].(string); ok && itemName != "" {
					name = itemName
				} else if itemName, ok := itemMap["page_name"].(string); ok && itemName != "" {
					name = itemName
				}

				// Extract HTML
				html := extractHTMLFromValue(itemMap, logger)
				if html != "" {
					filename := name
					if !strings.HasSuffix(filename, ".html") {
						filename = filename + ".html"
					}
					pages[filename] = html
					logger.Debug("Extracted page from array",
						zap.String("name", filename),
						zap.Int("index", i),
						zap.Int("length", len(html)),
					)
				}
			}
		}
		return pages
	}

	logger.Warn("Could not extract pages from value",
		zap.String("type", fmt.Sprintf("%T", value)))
	return pages
}

// extractHTMLFromValue tries to extract HTML string from various value types
func extractHTMLFromValue(value interface{}, logger *zap.Logger) string {
	// Direct string
	if html, ok := value.(string); ok {
		return html
	}

	// Map with HTML field
	if m, ok := value.(map[string]interface{}); ok {
		// Try common field names
		fieldNames := []string{"html", "page_html", "content", "result", "output"}
		for _, fieldName := range fieldNames {
			if html, ok := m[fieldName].(string); ok && html != "" {
				return html
			}
		}
	}

	return ""
}

// ensureValidHTML ensures HTML has proper structure
func ensureValidHTML(html string) string {
	html = strings.TrimSpace(html)

	// Add DOCTYPE if missing
	if !strings.HasPrefix(html, "<!DOCTYPE") {
		html = "<!DOCTYPE html>\n" + html
	}

	// Ensure has <html> tags
	if !strings.Contains(html, "<html") {
		html = strings.Replace(html, "<!DOCTYPE html>", "<!DOCTYPE html>\n<html lang=\"en\">", 1)
		if !strings.Contains(html, "</html>") {
			html = html + "\n</html>"
		}
	}

	// Ensure has <head>
	if !strings.Contains(html, "<head>") {
		headContent := `<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Page</title>
</head>`

		// Insert after <html> tag
		htmlTagEnd := strings.Index(html, "<html")
		if htmlTagEnd >= 0 {
			htmlTagEnd = strings.Index(html[htmlTagEnd:], ">")
			if htmlTagEnd >= 0 {
				insertPoint := strings.Index(html, "<html") + htmlTagEnd + 1
				html = html[:insertPoint] + "\n" + headContent + "\n" + html[insertPoint:]
			}
		}
	}

	// Ensure has <body>
	if !strings.Contains(html, "<body>") {
		html = strings.Replace(html, "</head>", "</head>\n<body>\n", 1)
		html = strings.Replace(html, "</html>", "</body>\n</html>", 1)
	}

	return html
}

// addSimpleNavigation adds a simple navigation bar
func addSimpleNavigation(html string, currentPage string) string {
	// Determine current page for active state
	pageClass := func(page string) string {
		if page == currentPage {
			return ` class="active"`
		}
		return ""
	}

	nav := fmt.Sprintf(`<nav style="padding: 20px; background: #f5f5f5; margin-bottom: 20px;">
    <a href="index.html"%s style="margin-right: 20px; color: #0066cc; text-decoration: none;">Home</a>
    <a href="about.html"%s style="margin-right: 20px; color: #0066cc; text-decoration: none;">About</a>
    <a href="contact.html"%s style="color: #0066cc; text-decoration: none;">Contact</a>
</nav>`,
		pageClass("index"),
		pageClass("about"),
		pageClass("contact"),
	)

	// Insert after <body> tag
	bodyIdx := strings.Index(html, "<body>")
	if bodyIdx >= 0 {
		insertPoint := bodyIdx + 6 // len("<body>")
		html = html[:insertPoint] + "\n" + nav + "\n" + html[insertPoint:]
	}

	return html
}

// addNavigationToAllPages adds navigation to each page
func addNavigationToAllPages(pages map[string]string, logger *zap.Logger) map[string]string {
	result := make(map[string]string)

	for name, html := range pages {
		// Determine current page name (remove .html)
		currentPage := strings.TrimSuffix(name, ".html")
		result[name] = addSimpleNavigation(html, currentPage)

		logger.Debug("Added navigation to page",
			zap.String("page", name),
			zap.String("current", currentPage),
		)
	}

	return result
}

// generateAboutPage creates a standard about page
func generateAboutPage(domain string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>About - %s</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6; 
            padding: 40px 20px; 
            max-width: 800px; 
            margin: 0 auto;
        }
        h1 { margin-bottom: 20px; color: #333; }
        p { margin-bottom: 15px; color: #666; }
    </style>
</head>
<body>
    <h1>About %s</h1>
    <p>Learn more about our company and what we do.</p>
    <p>We're dedicated to providing quality solutions for our customers.</p>
</body>
</html>`, domain, domain)
}

// generateContactPage creates a standard contact page
func generateContactPage(domain string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Contact - %s</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6; 
            padding: 40px 20px; 
            max-width: 800px; 
            margin: 0 auto;
        }
        h1 { margin-bottom: 20px; color: #333; }
        p { margin-bottom: 20px; color: #666; }
        form { margin-top: 30px; }
        label { display: block; margin: 15px 0 5px; font-weight: 500; }
        input, textarea { 
            width: 100%%; 
            padding: 10px; 
            border: 1px solid #ddd;
            border-radius: 4px;
            font-family: inherit;
        }
        button { 
            margin-top: 15px; 
            padding: 10px 24px; 
            background: #0066cc;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
        }
        button:hover { background: #0052a3; }
    </style>
</head>
<body>
    <h1>Contact Us</h1>
    <p>Get in touch with us for more information.</p>
    <form>
        <label for="name">Name</label>
        <input type="text" id="name" name="name" required>
        
        <label for="email">Email</label>
        <input type="email" id="email" name="email" required>
        
        <label for="message">Message</label>
        <textarea id="message" name="message" rows="5" required></textarea>
        
        <button type="submit">Send Message</button>
    </form>
</body>
</html>`, domain)
}

// extractDomainFromData extracts domain from collected data
func extractDomainFromData(data map[string]interface{}) string {
	// Try input_data.domain first
	if inputData, ok := data["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			return domain
		}
	}

	// Try direct domain field
	if domain, ok := data["domain"].(string); ok && domain != "" {
		return domain
	}

	// Fallback
	return "Our Company"
}

// extractNestedField navigates nested field paths like "step.result.html"
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
			// Try ExtractStepData for step results
			if extracted := datahelpers.ExtractStepData(v[part]); extracted != nil {
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

// extractFieldValue extracts a string value from nested field path
func extractFieldValue(data map[string]interface{}, fieldPath string, logger *zap.Logger) string {
	value := extractNestedField(data, fieldPath)
	if value == nil {
		logger.Warn("Field not found",
			zap.String("path", fieldPath),
		)
		return ""
	}

	// Direct string
	if str, ok := value.(string); ok {
		return str
	}

	// Map with common field names
	if m, ok := value.(map[string]interface{}); ok {
		fieldNames := []string{"html", "result", "content", "output", "page_html"}
		for _, fieldName := range fieldNames {
			if str, ok := m[fieldName].(string); ok && str != "" {
				return str
			}
		}
	}

	logger.Warn("Could not extract string from value",
		zap.String("path", fieldPath),
		zap.String("type", fmt.Sprintf("%T", value)),
	)
	return ""
}

// calculateTotalSize calculates total bytes across all pages
func calculateTotalSize(pages map[string]string) int {
	total := 0
	for _, content := range pages {
		total += len(content)
	}
	return total
}

// getPageNames returns sorted list of page names
func getPageNames(pages map[string]string) []string {
	names := make([]string, 0, len(pages))
	for name := range pages {
		names = append(names, name)
	}
	return names
}
