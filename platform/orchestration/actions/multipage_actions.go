// FILE: platform/orchestration/actions/multipage_actions.go
package actions

import (
	"context"
	"fmt"
	"regexp"
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
	html = cleanHTMLStructure(html)

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
	params.Logger.Info("Assembling multi-page site with consistent headers")

	config := params.StepConfig.Config
	if config == nil {
		return nil, fmt.Errorf("config is required for assemble_multipage_site action")
	}

	pagesField, _ := config["pages_field"].(string)
	if pagesField == "" {
		return nil, fmt.Errorf("pages_field is required in config")
	}

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

	// Build list of all page names for link fixing
	pageNamesList := make([]string, 0, len(pages))
	for name := range pages {
		pageNamesList = append(pageNamesList, name)
	}

	// Build header config once (shared across all pages except current page indicator)
	baseHeaderConfig := buildHeaderConfig(params.CollectedData, "", params.Logger)

	// POST-PROCESS ALL PAGES
	for name, html := range pages {
		// Step 1: Clean HTML structure (fix double DOCTYPE)
		html = cleanHTMLStructure(html)

		// Step 2: Fix anchor links to page links
		html = fixAnchorLinks(html, pageNamesList)

		// Step 3: Build page-specific header config (with active state)
		headerConfig := *baseHeaderConfig // Copy
		headerConfig.CurrentPage = strings.TrimSuffix(name, ".html")
		headerConfig.IsHomePage = name == "index.html" || name == "home.html"

		// Update active states in nav items
		for i := range headerConfig.NavItems {
			urlPage := strings.TrimSuffix(strings.TrimPrefix(headerConfig.NavItems[i].URL, "/"), ".html")
			headerConfig.NavItems[i].IsActive = urlPage == headerConfig.CurrentPage ||
				(headerConfig.CurrentPage == "index" && (urlPage == "home" || urlPage == "index"))
		}

		// Step 4: Inject consistent header
		html = injectConsistentHeader(html, &headerConfig, params.Logger)

		pages[name] = html

		params.Logger.Debug("Post-processed page",
			zap.String("page", name),
			zap.Int("nav_items", len(headerConfig.NavItems)),
			zap.Int("final_length", len(html)),
		)
	}

	params.Logger.Info("Multi-page site assembled with consistent headers",
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
func cleanHTMLStructure(html string) string {
	html = strings.TrimSpace(html)

	// Remove duplicate DOCTYPEs (case-insensitive)
	// Pattern: <!DOCTYPE html>...<!doctype html> or variations
	doctypeRe := regexp.MustCompile(`(?i)<!doctype\s+html\s*>`)
	matches := doctypeRe.FindAllStringIndex(html, -1)

	if len(matches) > 1 {
		// Keep only the first DOCTYPE, remove others
		// Work backwards to preserve indices
		for i := len(matches) - 1; i > 0; i-- {
			start := matches[i][0]
			end := matches[i][1]
			html = html[:start] + html[end:]
		}
	}

	// Remove duplicate <html> tags
	htmlTagRe := regexp.MustCompile(`(?i)<html[^>]*>`)
	htmlMatches := htmlTagRe.FindAllStringIndex(html, -1)
	if len(htmlMatches) > 1 {
		for i := len(htmlMatches) - 1; i > 0; i-- {
			start := htmlMatches[i][0]
			end := htmlMatches[i][1]
			html = html[:start] + html[end:]
		}
	}

	// Remove duplicate <head> sections (keep the more complete one)
	headCount := strings.Count(strings.ToLower(html), "<head")
	if headCount > 1 {
		// Find both head sections
		lowerHTML := strings.ToLower(html)
		firstHeadStart := strings.Index(lowerHTML, "<head")
		firstHeadEnd := strings.Index(lowerHTML, "</head>")

		if firstHeadEnd > firstHeadStart {
			// Check if there's a second head after the first
			secondHeadStart := strings.Index(lowerHTML[firstHeadEnd:], "<head")
			if secondHeadStart >= 0 {
				// Find the second head's end
				secondHeadStart += firstHeadEnd
				secondHeadEnd := strings.Index(lowerHTML[secondHeadStart:], "</head>")
				if secondHeadEnd >= 0 {
					secondHeadEnd += secondHeadStart + 7 // include </head>

					// Compare sizes - keep the larger one
					firstHeadLen := firstHeadEnd - firstHeadStart
					secondHeadLen := secondHeadEnd - secondHeadStart

					if secondHeadLen > firstHeadLen {
						// Remove first head, keep second
						html = html[:firstHeadStart] + html[firstHeadEnd+7:]
					} else {
						// Remove second head, keep first
						html = html[:secondHeadStart] + html[secondHeadEnd:]
					}
				}
			}
		}
	}

	// Remove duplicate <body> tags
	bodyTagRe := regexp.MustCompile(`(?i)<body[^>]*>`)
	bodyMatches := bodyTagRe.FindAllStringIndex(html, -1)
	if len(bodyMatches) > 1 {
		for i := len(bodyMatches) - 1; i > 0; i-- {
			start := bodyMatches[i][0]
			end := bodyMatches[i][1]
			html = html[:start] + html[end:]
		}
	}

	// Ensure we have a DOCTYPE at the start
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(html)), "<!DOCTYPE") {
		html = "<!DOCTYPE html>\n" + html
	}

	return html
}

func fixAnchorLinks(html string, pageNames []string) string {
	// Build set of known pages
	knownPages := make(map[string]bool)
	for _, name := range pageNames {
		cleanName := strings.TrimSuffix(strings.ToLower(name), ".html")
		knownPages[cleanName] = true
	}

	// Add common page names
	commonPages := []string{
		"home", "index", "about", "services", "contact", "insights", "blog",
		"careers", "team", "portfolio", "case-studies", "privacy", "terms",
		"pricing", "faq", "support", "features", "solutions", "resources",
		"testimonials", "clients", "work", "projects",
	}
	for _, name := range commonPages {
		knownPages[name] = true
	}

	result := html

	for pageName := range knownPages {
		targetURL := "/" + pageName + ".html"
		if pageName == "home" || pageName == "index" {
			targetURL = "/index.html"
		}

		// Replace various patterns
		patterns := []struct {
			old string
			new string
		}{
			// Double quotes
			{fmt.Sprintf(`href="#%s"`, pageName), fmt.Sprintf(`href="%s"`, targetURL)},
			{fmt.Sprintf(`href="#%s"`, strings.Title(pageName)), fmt.Sprintf(`href="%s"`, targetURL)},
			// Single quotes
			{fmt.Sprintf(`href='#%s'`, pageName), fmt.Sprintf(`href='%s'`, targetURL)},
			{fmt.Sprintf(`href='#%s'`, strings.Title(pageName)), fmt.Sprintf(`href='%s'`, targetURL)},
			// No quotes (minified)
			{fmt.Sprintf(`href=#%s>`, pageName), fmt.Sprintf(`href=%s>`, targetURL)},
			{fmt.Sprintf(`href=#%s `, pageName), fmt.Sprintf(`href=%s `, targetURL)},
		}

		for _, p := range patterns {
			result = strings.ReplaceAll(result, p.old, p.new)
		}
	}

	return result
}

func extractCanonicalNavigation(collectedData map[string]interface{}, logger *zap.Logger) []map[string]string {
	var navItems []map[string]string

	// Priority 1: db_sync.navigation (from sync_pages_to_db)
	if dbSync, ok := collectedData["db_sync"].(map[string]interface{}); ok {
		if nav, ok := dbSync["navigation"].(map[string]interface{}); ok {
			if items, ok := nav["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)
						if label != "" && url != "" {
							navItems = append(navItems, map[string]string{
								"label": label,
								"url":   url,
							})
						}
					}
				}
				if len(navItems) > 0 {
					logger.Info("Using navigation from db_sync",
						zap.Int("nav_items", len(navItems)),
					)
					return navItems
				}
			}
		}
	}

	// Priority 2: page_plan.plan_data.sitemap
	if pagePlan, ok := collectedData["page_plan"].(map[string]interface{}); ok {
		var sitemap []interface{}

		if planData, ok := pagePlan["plan_data"].(map[string]interface{}); ok {
			if sm, ok := planData["sitemap"].([]interface{}); ok {
				sitemap = sm
			}
		}
		if sitemap == nil {
			if sm, ok := pagePlan["sitemap"].([]interface{}); ok {
				sitemap = sm
			}
		}

		for _, entry := range sitemap {
			if e, ok := entry.(map[string]interface{}); ok {
				label, _ := e["label"].(string)
				if label == "" {
					label, _ = e["title"].(string)
				}
				if label == "" {
					label, _ = e["name"].(string)
				}

				url, _ := e["url"].(string)
				if url == "" {
					if name, ok := e["name"].(string); ok {
						if name == "index" || name == "home" {
							url = "/index.html"
						} else {
							url = "/" + name + ".html"
						}
					}
				}

				inHeader := true
				if ih, ok := e["in_header"].(bool); ok {
					inHeader = ih
				}

				// Skip footer-only pages
				nameLower := strings.ToLower(label)
				if nameLower == "privacy" || nameLower == "terms" ||
					strings.Contains(nameLower, "privacy") || strings.Contains(nameLower, "terms") {
					inHeader = false
				}

				if inHeader && label != "" && url != "" {
					navItems = append(navItems, map[string]string{
						"label": label,
						"url":   url,
					})
				}
			}
		}

		if len(navItems) > 0 {
			logger.Info("Using navigation from page_plan sitemap",
				zap.Int("nav_items", len(navItems)),
			)
			return navItems
		}
	}

	logger.Warn("No canonical navigation found")
	return nil
}

// ===========================================================================
// NEW FUNCTION: injectCanonicalNavigation
// ===========================================================================
// Replaces header navigation with canonical navigation

func injectCanonicalNavigation(html string, navItems []map[string]string, currentPageName string, logger *zap.Logger) string {
	// Build the navigation HTML
	var navLinks []string
	currentPage := strings.TrimSuffix(currentPageName, ".html")

	for _, item := range navItems {
		label := item["label"]
		url := item["url"]

		// Add active class for current page
		activeClass := ""
		itemPage := strings.TrimSuffix(strings.TrimPrefix(url, "/"), ".html")
		if itemPage == currentPage || (currentPage == "index" && (itemPage == "home" || itemPage == "index")) {
			activeClass = ` class="active"`
		}

		navLinks = append(navLinks, fmt.Sprintf(`<a href="%s"%s>%s</a>`, url, activeClass, label))
	}

	navHTML := strings.Join(navLinks, "\n            ")

	// Try to find and replace existing nav content
	// Pattern: <nav...>...</nav>
	navRe := regexp.MustCompile(`(?is)<nav[^>]*>.*?</nav>`)

	if navRe.MatchString(html) {
		// Replace the nav content but keep any nav attributes
		html = navRe.ReplaceAllStringFunc(html, func(match string) string {
			// Extract opening nav tag
			openTagEnd := strings.Index(match, ">")
			if openTagEnd < 0 {
				return match
			}
			openTag := match[:openTagEnd+1]

			return fmt.Sprintf(`%s
        <ul>
            %s
        </ul>
    </nav>`, strings.TrimSuffix(openTag, ">")+" id=\"main-nav\">",
				strings.ReplaceAll(navHTML, "<a ", "<li><a "))
		})

		logger.Debug("Replaced nav content",
			zap.String("page", currentPageName),
		)
	} else {
		// No nav found, try to insert after header tag
		headerEnd := strings.Index(strings.ToLower(html), "<header")
		if headerEnd >= 0 {
			// Find the end of header opening tag
			closeIdx := strings.Index(html[headerEnd:], ">")
			if closeIdx >= 0 {
				insertPoint := headerEnd + closeIdx + 1
				newNav := fmt.Sprintf(`
    <nav id="main-nav">
        <ul>
            %s
        </ul>
    </nav>`, strings.ReplaceAll(navHTML, "<a ", "<li><a "))
				html = html[:insertPoint] + newNav + html[insertPoint:]
			}
		}

		logger.Debug("Inserted new nav",
			zap.String("page", currentPageName),
		)
	}

	return html
}

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

	// Format 1: Map - could be direct pages OR loop_complete output
	if pagesMap, ok := value.(map[string]interface{}); ok {
		// Check for loop_complete output format: {iterations: N, results: [...]}
		if results, hasResults := pagesMap["results"].([]interface{}); hasResults {
			logger.Info("Detected loop_complete format, extracting from results array",
				zap.Int("results_count", len(results)))

			// Process the results array (same as Format 2)
			for i, item := range results {
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
						logger.Debug("Extracted page from loop_complete results",
							zap.String("name", filename),
							zap.Int("index", i),
							zap.Int("length", len(html)),
						)
					}
				}
			}
			return pages
		}

		// Otherwise, treat as direct map of pages {"index": "...", "about": "...", ...}
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
	// Just use cleanHTMLStructure - it handles everything
	return cleanHTMLStructure(html)
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
/*func extractFieldValue(data map[string]interface{}, fieldPath string, logger *zap.Logger) string {
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
}*/

// extractFieldValue navigates nested field paths like "base_structure.result"
func extractFieldValue(data map[string]interface{}, fieldPath string, logger *zap.Logger) string {
	parts := strings.Split(fieldPath, ".")

	var current interface{} = data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			// First try direct access
			if val, ok := v[part]; ok {
				current = val
				continue
			}
			// Then try ExtractStepData if it looks like a step result
			if extracted := datahelpers.ExtractStepData(v[part]); extracted != nil {
				current = extracted
				continue
			}
			logger.Warn("Field not found in path",
				zap.String("field", part),
				zap.String("full_path", fieldPath),
			)
			return ""
		default:
			// If we're at a terminal value and still have more parts, something's wrong
			if len(parts) > 1 {
				logger.Warn("Cannot traverse further, value is not a map",
					zap.String("field", part),
					zap.String("full_path", fieldPath),
				)
				return ""
			}
		}
	}

	// Convert final value to string
	switch v := current.(type) {
	case string:
		return v
	case map[string]interface{}:
		// If it's still a map, try to get "result" or "html" or "content"
		if result, ok := v["result"].(string); ok {
			return result
		}
		if html, ok := v["html"].(string); ok {
			return html
		}
		if content, ok := v["content"].(string); ok {
			return content
		}
		logger.Warn("Final value is a map but couldn't extract string",
			zap.String("full_path", fieldPath),
		)
		return ""
	default:
		logger.Warn("Final value is not a string",
			zap.String("full_path", fieldPath),
			zap.String("type", fmt.Sprintf("%T", current)),
		)
		return ""
	}
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

// extractDomainFromCollectedData searches for domain in collected data
func extractDomainFromCollectedData(collectedData map[string]interface{}) string {
	// Try input_data first
	if inputData, ok := collectedData["input_data"]; ok {
		if inputMap, ok := inputData.(map[string]interface{}); ok {
			if domain, ok := inputMap["domain"].(string); ok && domain != "" {
				return domain
			}
		}
	}

	// Search recursively
	domain := findStringInMap(collectedData, "domain", 0)
	if domain != "" {
		return domain
	}

	return "Our Company"
}

// extractBusinessInfoMap extracts business info from collected data
func extractBusinessInfoMap(collectedData map[string]interface{}) map[string]interface{} {
	businessInfo := make(map[string]interface{})

	// Try to find domain
	domain := extractDomainFromCollectedData(collectedData)
	if domain != "" {
		businessInfo["domain"] = domain
	}

	// Try to find objective
	if inputData, ok := collectedData["input_data"]; ok {
		if inputMap, ok := inputData.(map[string]interface{}); ok {
			if objective, ok := inputMap["objective"].(string); ok && objective != "" {
				businessInfo["objective"] = objective
			}
		}
	}

	// Fallback: search recursively
	if businessInfo["objective"] == nil {
		objective := findStringInMap(collectedData, "objective", 0)
		if objective != "" {
			businessInfo["objective"] = objective
		}
	}

	return businessInfo
}

// findStringInMap recursively searches for a string field
func findStringInMap(data interface{}, key string, depth int) string {
	if depth > 10 {
		return ""
	}

	m, ok := data.(map[string]interface{})
	if !ok {
		return ""
	}

	// Direct match
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok && str != "" {
			return str
		}
	}

	// Recurse into values
	for _, val := range m {
		if result := findStringInMap(val, key, depth+1); result != "" {
			return result
		}
	}

	return ""
}

// extractDomain extracts domain from business info map

func getConfigKeys(config map[string]interface{}) []string {
	keys := make([]string, 0, len(config))
	for k := range config {
		keys = append(keys, k)
	}
	return keys
}

func getCollectedDataKeys(data map[string]interface{}) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		// Skip internal fields
		if !strings.HasPrefix(k, "__") {
			keys = append(keys, k)
		}
	}
	return keys
}

// NavItem represents a single nav link
type NavItem struct {
	Label    string
	URL      string
	IsActive bool
}

// HeaderConfig holds configuration for header generation
type HeaderConfig struct {
	LogoText     string
	LogoAccent   string
	NavItems     []NavItem
	PrimaryColor string
	AccentColor  string
	CurrentPage  string
	IsHomePage   bool
}

// buildHeaderConfig extracts header configuration from collected data
func buildHeaderConfig(collectedData map[string]interface{}, currentPageName string, logger *zap.Logger) *HeaderConfig {
	config := &HeaderConfig{
		LogoText:     "Company",
		LogoAccent:   "",
		PrimaryColor: "#1a1a2e",
		AccentColor:  "#16a085",
		CurrentPage:  strings.TrimSuffix(currentPageName, ".html"),
		IsHomePage:   currentPageName == "index.html" || currentPageName == "home.html",
	}

	// Try to get domain/business name for logo
	if inputData, ok := collectedData["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			// Extract business name from domain
			parts := strings.Split(domain, ".")
			if len(parts) > 0 {
				name := parts[0]
				// Capitalize first letter, handle common suffixes
				if len(name) > 0 {
					config.LogoText = strings.ToUpper(name[:1]) + name[1:]
				}
			}
		}

		// Try to get colors from reviewed_brief
		if brief, ok := inputData["reviewed_brief"].(map[string]interface{}); ok {
			if colors, ok := brief["color_scheme"].(string); ok && colors != "" {
				config.PrimaryColor, config.AccentColor = parseColorScheme(colors)
			}
		}
	}

	// Get navigation items
	config.NavItems = extractNavItemsForHeader(collectedData, config.CurrentPage, logger)

	return config
}

// parseColorScheme extracts primary and accent colors from a description
func parseColorScheme(scheme string) (primary, accent string) {
	primary = "#1a1a2e"
	accent = "#16a085"

	schemeLower := strings.ToLower(scheme)

	if strings.Contains(schemeLower, "dark") {
		primary = "#1a1a2e"
	}
	if strings.Contains(schemeLower, "navy") {
		primary = "#1e3a5f"
	}
	if strings.Contains(schemeLower, "teal") {
		accent = "#16a085"
	}
	if strings.Contains(schemeLower, "gold") {
		accent = "#d4af37"
	}
	if strings.Contains(schemeLower, "blue") {
		accent = "#2563eb"
	}
	if strings.Contains(schemeLower, "green") {
		accent = "#059669"
	}
	if strings.Contains(schemeLower, "purple") {
		accent = "#7c3aed"
	}

	return primary, accent
}

// extractNavItemsForHeader gets navigation items with simple labels
func extractNavItemsForHeader(collectedData map[string]interface{}, currentPage string, logger *zap.Logger) []NavItem {
	var items []NavItem

	// Priority 1: db_sync.navigation
	if dbSync, ok := collectedData["db_sync"].(map[string]interface{}); ok {
		if nav, ok := dbSync["navigation"].(map[string]interface{}); ok {
			if navItems, ok := nav["items"].([]interface{}); ok {
				for _, item := range navItems {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)

						if label != "" && url != "" {
							urlPage := strings.TrimSuffix(strings.TrimPrefix(url, "/"), ".html")
							isActive := urlPage == currentPage ||
								(currentPage == "index" && (urlPage == "home" || urlPage == "index"))

							items = append(items, NavItem{
								Label:    label, // Already simplified by buildNavigationFromPages
								URL:      url,
								IsActive: isActive,
							})
						}
					}
				}
			}
		}
	}

	// Fallback: default navigation
	if len(items) == 0 {
		logger.Warn("No navigation found, using defaults")
		items = []NavItem{
			{Label: "Home", URL: "/index.html", IsActive: currentPage == "index"},
			{Label: "About", URL: "/about.html", IsActive: currentPage == "about"},
			{Label: "Services", URL: "/services.html", IsActive: currentPage == "services"},
			{Label: "Contact", URL: "/contact.html", IsActive: currentPage == "contact"},
		}
	}

	return items
}

// generateConsistentHeader creates the header HTML
func generateConsistentHeader(config *HeaderConfig) string {
	var navLinks []string
	for _, item := range config.NavItems {
		activeClass := ""
		if item.IsActive {
			activeClass = ` class="active"`
		}
		navLinks = append(navLinks, fmt.Sprintf(
			`<li><a href="%s"%s>%s</a></li>`,
			item.URL, activeClass, item.Label,
		))
	}

	navHTML := strings.Join(navLinks, "\n                ")

	logoHTML := fmt.Sprintf(`<span class="logo-text">%s</span>`, config.LogoText)
	if config.LogoAccent != "" {
		logoHTML = fmt.Sprintf(`<span class="logo-text">%s</span><span class="logo-accent">%s</span>`,
			config.LogoText, config.LogoAccent)
	}

	return fmt.Sprintf(`<header class="site-header">
    <div class="header-container">
        <a href="/index.html" class="logo">
            %s
        </a>
        <button class="mobile-menu-toggle" aria-label="Toggle menu">
            <span></span><span></span><span></span>
        </button>
        <nav class="main-nav">
            <ul>
                %s
            </ul>
        </nav>
    </div>
</header>`, logoHTML, navHTML)
}

// generateHeaderStyles creates CSS for the header
func generateHeaderStyles(config *HeaderConfig) string {
	return fmt.Sprintf(`
/* ========== CONSISTENT HEADER STYLES ========== */
.site-header {
    background: %s;
    padding: 1rem 0;
    position: sticky;
    top: 0;
    z-index: 1000;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
}
.header-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 2rem;
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.logo {
    text-decoration: none;
    font-size: 1.5rem;
    font-weight: 700;
    color: white;
}
.logo-accent { color: %s; }
.main-nav ul {
    display: flex;
    list-style: none;
    margin: 0;
    padding: 0;
    gap: 2rem;
}
.main-nav a {
    color: rgba(255,255,255,0.9);
    text-decoration: none;
    font-weight: 500;
    padding: 0.5rem 0;
    transition: color 0.2s;
}
.main-nav a:hover,
.main-nav a.active { color: %s; }
.mobile-menu-toggle {
    display: none;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
}
.mobile-menu-toggle span {
    display: block;
    width: 24px;
    height: 2px;
    background: white;
    margin: 5px 0;
}
@media (max-width: 768px) {
    .mobile-menu-toggle { display: block; }
    .main-nav {
        position: absolute;
        top: 100%%;
        left: 0;
        right: 0;
        background: %s;
        padding: 1rem;
        display: none;
    }
    .main-nav.active { display: block; }
    .main-nav ul { flex-direction: column; gap: 0; }
    .main-nav a { display: block; padding: 0.75rem 0; border-bottom: 1px solid rgba(255,255,255,0.1); }
}
/* ========== END HEADER STYLES ========== */
`, config.PrimaryColor, config.AccentColor, config.AccentColor, config.PrimaryColor)
}

// injectConsistentHeader replaces the existing header with a consistent one
func injectConsistentHeader(html string, config *HeaderConfig, logger *zap.Logger) string {
	headerHTML := generateConsistentHeader(config)
	headerCSS := generateHeaderStyles(config)

	// Step 1: Remove existing header element
	headerRe := regexp.MustCompile(`(?is)<header[^>]*>.*?</header>`)
	html = headerRe.ReplaceAllString(html, "<!-- HEADER_PLACEHOLDER -->")

	// Step 2: Insert new header after <body> tag
	bodyRe := regexp.MustCompile(`(?i)(<body[^>]*>)`)
	if bodyRe.MatchString(html) {
		html = bodyRe.ReplaceAllString(html, "$1\n"+headerHTML)
		html = strings.Replace(html, "<!-- HEADER_PLACEHOLDER -->", "", 1)
	} else {
		html = strings.Replace(html, "<!-- HEADER_PLACEHOLDER -->", headerHTML, 1)
	}

	// Step 3: Inject CSS into <style> or <head>
	styleRe := regexp.MustCompile(`(?i)(</style>)`)
	if styleRe.MatchString(html) {
		// Insert before first </style>
		replaced := false
		html = styleRe.ReplaceAllStringFunc(html, func(match string) string {
			if !replaced {
				replaced = true
				return headerCSS + "\n" + match
			}
			return match
		})
	} else {
		// No style tag, add one in head
		headCloseRe := regexp.MustCompile(`(?i)(</head>)`)
		if headCloseRe.MatchString(html) {
			html = headCloseRe.ReplaceAllString(html, "<style>"+headerCSS+"</style>\n$1")
		}
	}

	// Step 4: Add mobile menu JS
	if !strings.Contains(html, "mobile-menu-toggle") || !strings.Contains(strings.ToLower(html), "addeventlistener") {
		mobileJS := `<script>
document.addEventListener('DOMContentLoaded', function() {
    var toggle = document.querySelector('.mobile-menu-toggle');
    var nav = document.querySelector('.main-nav');
    if (toggle && nav) {
        toggle.addEventListener('click', function() {
            nav.classList.toggle('active');
        });
    }
});
</script>`
		bodyCloseRe := regexp.MustCompile(`(?i)(</body>)`)
		if bodyCloseRe.MatchString(html) {
			html = bodyCloseRe.ReplaceAllString(html, mobileJS+"\n$1")
		}
	}

	logger.Debug("Injected consistent header",
		zap.Int("nav_items", len(config.NavItems)),
		zap.String("primary_color", config.PrimaryColor),
	)

	return html
}
