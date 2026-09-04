// FILE: platform/orchestration/actions/html_actions.go
// HTML generation with smart content-aware data extraction

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// GenerateHTMLAction generates HTML using LLM based on input_fields configuration
func GenerateHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Generating HTML content with smart extraction",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	// Get configuration
	config := params.StepConfig.Config
	generationType, _ := config["generation_type"].(string)
	maxTokens := getMaxTokens(config, 16000)

	// Get input_fields from config
	inputFieldsRaw, hasInputFields := config["input_fields"]
	var inputFields []string
	if hasInputFields {
		if fields, ok := inputFieldsRaw.([]interface{}); ok {
			for _, f := range fields {
				if fieldStr, ok := f.(string); ok {
					inputFields = append(inputFields, fieldStr)
				}
			}
		}
	}

	params.Logger.Info("Using input fields",
		zap.Strings("input_fields", inputFields),
		zap.String("generation_type", generationType),
	)

	// Build context with SMART extraction
	context := buildContextSmart(params.CollectedData, inputFields, params.Logger)

	params.Logger.Info("Context built",
		zap.Any("context_keys", datahelpers.GetMapKeys(context)),
	)

	// Build prompt
	var prompt string
	switch generationType {
	case "structure":
		prompt = buildStructurePrompt(context)
	case "styles":
		prompt = buildStylesPrompt(context)
	case "content":
		prompt = buildContentPrompt(context)
	default:
		prompt = buildFullHTMLPrompt(context)
	}

	params.Logger.Info("Generated prompt",
		zap.Int("prompt_length", len(prompt)),
		zap.String("prompt_preview", prompt[:min(500, len(prompt))]),
	)

	// Call LLM
	llmParams := params
	// IMPORTANT: Set prompt at CollectedData["prompt"] where getPromptWithPriority looks for it
	llmParams.CollectedData["prompt"] = prompt
	// Also set input_data for template rendering
	llmParams.CollectedData["input_data"] = map[string]interface{}{
		"prompt": prompt,
	}
	llmParams.StepConfig.Config = map[string]interface{}{
		"ai_service": map[string]interface{}{
			"model":           "claude-sonnet-4-5-20250514",
			"provider":        "anthropic",
			"api_key_env_var": "ANTHROPIC_API_KEY",
			"max_tokens":      maxTokens,
		},
		"prompt": prompt, // Also set in StepConfig.Config for Priority 1
	}

	result, err := ExecuteLLMPromptAction(ctx, llmParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}

	htmlContent := datahelpers.ExtractHTMLFromResponse(result)

	params.Logger.Info("HTML generation complete",
		zap.Int("html_length", len(htmlContent)),
	)

	return map[string]interface{}{
		"raw_html":        htmlContent,
		"generation_type": generationType,
		"generated_at":    time.Now().UTC(),
		"tokens_used":     maxTokens,
	}, nil
}

// ProcessHTMLAction processes and enhances generated HTML
func ProcessHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Processing HTML content")

	var rawHTML string

	// Check generate_html step
	if genStep, ok := params.CollectedData["generate_html"]; ok {
		extracted := datahelpers.ExtractStepData(genStep)
		if genResult, ok := extracted.(map[string]interface{}); ok {
			rawHTML, _ = genResult["raw_html"].(string)
		}
	}

	// Fallback to direct raw_html field
	if rawHTML == "" {
		if htmlData, ok := params.CollectedData["raw_html"]; ok {
			extracted := datahelpers.ExtractStepData(htmlData)
			if str, ok := extracted.(string); ok {
				rawHTML = str
			}
		}
	}

	if rawHTML == "" {
		return nil, fmt.Errorf("no HTML content to process")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	businessInfo := datahelpers.ExtractBusinessInfo(params.CollectedData)
	processingSteps := []string{}

	datahelpers.EnsureHTMLStructure(doc)
	processingSteps = append(processingSteps, "structure_validation")

	datahelpers.AddMetaTags(doc, businessInfo)
	processingSteps = append(processingSteps, "meta_tags")

	datahelpers.EnsureResponsiveDesign(doc)
	processingSteps = append(processingSteps, "responsive_design")

	datahelpers.OptimizeImages(doc)
	processingSteps = append(processingSteps, "image_optimization")

	datahelpers.AddStructuredData(doc, businessInfo)
	processingSteps = append(processingSteps, "structured_data")

	processedHTML, _ := doc.Html()

	if shouldMinify(params) {
		processedHTML = datahelpers.MinifyHTML(processedHTML, params.Logger)
		processingSteps = append(processingSteps, "minification")
	}

	return map[string]interface{}{
		"processed_html":   processedHTML,
		"original_size":    len(rawHTML),
		"processed_size":   len(processedHTML),
		"processing_steps": processingSteps,
		"business_info":    businessInfo,
	}, nil
}

// ValidateHTMLAction validates the processed HTML
func ValidateHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Validating HTML content")

	var htmlContent string

	if procStep, ok := params.CollectedData["process_html"]; ok {
		extracted := datahelpers.ExtractStepData(procStep)
		if procResult, ok := extracted.(map[string]interface{}); ok {
			htmlContent, _ = procResult["processed_html"].(string)
		}
	}

	if htmlContent == "" {
		return nil, fmt.Errorf("no HTML content to validate")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return map[string]interface{}{
			"valid":  false,
			"errors": []string{fmt.Sprintf("Failed to parse HTML: %v", err)},
		}, nil
	}

	errors := []string{}
	warnings := []string{}

	if doc.Find("html").Length() == 0 {
		errors = append(errors, "Missing <html> element")
	}
	if doc.Find("head").Length() == 0 {
		errors = append(errors, "Missing <head> element")
	}
	if doc.Find("body").Length() == 0 {
		errors = append(errors, "Missing <body> element")
	}
	if doc.Find("title").Length() == 0 {
		warnings = append(warnings, "Missing <title> element")
	}
	if doc.Find("meta[charset]").Length() == 0 {
		warnings = append(warnings, "Missing charset meta tag")
	}
	if doc.Find("meta[name='viewport']").Length() == 0 {
		warnings = append(warnings, "Missing viewport meta tag")
	}

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		alt, hasAlt := s.Attr("alt")
		if src == "" {
			errors = append(errors, fmt.Sprintf("Image %d has no src attribute", i+1))
		}
		if !hasAlt || alt == "" {
			warnings = append(warnings, fmt.Sprintf("Image %d missing alt text", i+1))
		}
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" || href == "#" {
			warnings = append(warnings, fmt.Sprintf("Link %d has empty or placeholder href", i+1))
		}
	})

	isValid := len(errors) == 0
	finalHTML := htmlContent
	if isValid {
		params.CollectedData["final_html"] = finalHTML
	}

	return map[string]interface{}{
		"valid":         isValid,
		"errors":        errors,
		"warnings":      warnings,
		"html_size":     len(finalHTML),
		"element_count": datahelpers.CountElements(doc),
		"final_html":    finalHTML,
	}, nil
}

// ============================================================================
// SMART CONTEXT EXTRACTION
// ============================================================================

// buildContextSmart uses intelligent search to find data regardless of nesting
func buildContextSmart(collectedData map[string]interface{}, inputFields []string, logger *zap.Logger) map[string]interface{} {
	if len(inputFields) == 0 {
		logger.Warn("No input_fields specified, using fallback field names")
		inputFields = []string{"input_data", "site_architecture", "site_content", "domain_analysis"}
	}

	logger.Info("Building context with UNIFIED EXTRACTOR",
		zap.Strings("input_fields", inputFields),
	)

	// USE THE UNIFIED EXTRACTOR
	context := datahelpers.ExtractFields(collectedData, inputFields, logger)

	logger.Info("Context built successfully",
		zap.Int("field_count", len(context)),
		zap.Strings("context_keys", datahelpers.GetMapKeys(context)),
	)

	return context
}

// findStringField recursively searches for a string field by name
func findStringField(data interface{}, fieldName string, depth int, logger *zap.Logger) string {
	if depth > 15 {
		return ""
	}

	if m, ok := data.(map[string]interface{}); ok {
		// Check direct field
		if val, ok := m[fieldName]; ok {
			if str, ok := val.(string); ok && str != "" {
				logger.Debug("Found field",
					zap.String("field", fieldName),
					zap.Int("depth", depth),
				)
				return str
			}
		}

		// Recurse into all values
		for _, val := range m {
			if result := findStringField(val, fieldName, depth+1, logger); result != "" {
				return result
			}
		}
	}

	if slice, ok := data.([]interface{}); ok {
		for _, val := range slice {
			if result := findStringField(val, fieldName, depth+1, logger); result != "" {
				return result
			}
		}
	}

	return ""
}

// ============================================================================
// PROMPT BUILDING
// ============================================================================

// extractStructuredContent extracts content that has proper website structure (hero/sections/meta/footer)
// Returns empty string if the content is just prose text (like sections[0].content)
// This prevents the alias system from returning the wrong data
func extractStructuredContent(data interface{}, logger *zap.Logger) string {
	if data == nil {
		return ""
	}

	// If it's a string, check if it looks like JSON with structure
	if str, ok := data.(string); ok {
		cleaned := datahelpers.CleanHTMLString(str)
		// Try to parse as JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(cleaned), &parsed); err == nil {
			// Check if parsed JSON has expected structure
			if isStructuredWebContent(parsed) {
				return cleaned
			}
		}
		// It's just a string - probably prose text, not structured content
		return ""
	}

	// If it's a map, check for nested content patterns
	if m, ok := data.(map[string]interface{}); ok {
		// Check if THIS map is the structured content
		if isStructuredWebContent(m) {
			if contentJSON, err := json.Marshal(m); err == nil {
				return string(contentJSON)
			}
		}

		// Pattern 1: content_result.result (from content-creator)
		if contentResult, ok := m["content_result"].(map[string]interface{}); ok {
			if result, ok := contentResult["result"]; ok {
				return extractStructuredContent(result, logger)
			}
		}

		// Pattern 2: create_content.result
		if createContent, ok := m["create_content"].(map[string]interface{}); ok {
			if result, ok := createContent["result"]; ok {
				return extractStructuredContent(result, logger)
			}
		}

		// Pattern 3: Nested in input_data (call_agent response pattern)
		if inputData, ok := m["input_data"].(map[string]interface{}); ok {
			if found := extractStructuredContent(inputData, logger); found != "" {
				return found
			}
		}

		// Pattern 4: Direct result field
		if result, ok := m["result"]; ok {
			return extractStructuredContent(result, logger)
		}
	}

	return ""
}

// isStructuredWebContent checks if a map looks like website content structure
// Must have at least one of: hero, sections, meta, footer
func isStructuredWebContent(m map[string]interface{}) bool {
	_, hasHero := m["hero"]
	_, hasSections := m["sections"]
	_, hasMeta := m["meta"]
	_, hasFooter := m["footer"]

	// Must have at least 2 of these to be considered structured content
	count := 0
	if hasHero {
		count++
	}
	if hasSections {
		count++
	}
	if hasMeta {
		count++
	}
	if hasFooter {
		count++
	}

	return count >= 2
}

// checks for both site_content and page_content
func buildFullHTMLPrompt(context map[string]interface{}) string {
	var domainInfo, architectureInfo, contentInfo, sitemapInfo, designInfo string

	// Extract domain/business info - check root level first (after flattening)
	if domain, ok := context["domain"].(string); ok && domain != "" {
		domainInfo = domain
	} else if bizName, ok := context["business_name"].(string); ok && bizName != "" {
		domainInfo = bizName
	} else if inputData, ok := context["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			domainInfo = domain
		} else if bizName, ok := inputData["business_name"].(string); ok && bizName != "" {
			domainInfo = bizName
		}
	}

	// Extract architecture
	if arch, ok := context["site_architecture"]; ok {
		if archStr, ok := arch.(string); ok {
			architectureInfo = archStr
		} else if archJSON, err := json.Marshal(arch); err == nil {
			architectureInfo = string(archJSON)
		}
	}

	// Extract sitemap for navigation links (CANONICAL from database)
	sitemapInfo = extractSitemapInfo(context)

	// Extract design context for consistency
	designInfo = extractDesignContext(context)

	// Extract content - Priority: page_content > site_content > content
	contentFound := false

	if content, ok := context["page_content"]; ok {
		if extracted := extractStructuredContent(content, nil); extracted != "" {
			if len(extracted) > 8000 {
				contentInfo = extracted[:8000] + "... [truncated]"
			} else {
				contentInfo = extracted
			}
			contentFound = true
		}
	}

	if !contentFound {
		if content, ok := context["site_content"]; ok {
			if extracted := extractStructuredContent(content, nil); extracted != "" {
				if len(extracted) > 8000 {
					contentInfo = extracted[:8000] + "... [truncated]"
				} else {
					contentInfo = extracted
				}
				contentFound = true
			}
		}
	}

	if !contentFound {
		if content, ok := context["content"]; ok {
			if extracted := extractStructuredContent(content, nil); extracted != "" {
				if len(extracted) > 8000 {
					contentInfo = extracted[:8000] + "... [truncated]"
				} else {
					contentInfo = extracted
				}
				contentFound = true
			}
		}
	}

	// Extract page info
	pageInfo := extractPageInfo(context)

	return fmt.Sprintf(`Generate a complete, production-ready HTML5 page.

Business/Domain: %s
%s

%s
%s
%s

%s

STRICT REQUIREMENTS:

1. DOCUMENT STRUCTURE:
   - Complete HTML5 document starting with <!DOCTYPE html>
   - Single <html>, <head>, and <body> element
   - All CSS must be inline in <style> tags (no external files)
   - All JS must be inline in <script> tags (no external files)

2. NAVIGATION (MANDATORY):
   - Copy the EXACT navigation items listed above into the header <nav>
   - Use the EXACT URLs shown (e.g., /about.html NOT #about)
   - Do NOT use anchor links (#about) for page navigation
   - Do NOT add or remove navigation items
   - EVERY page must have IDENTICAL header navigation

3. DESIGN CONSISTENCY:
   - Use the color scheme specified above consistently
   - Maintain the same header/footer design as other pages
   - Keep typography and spacing consistent
   - Use CSS variables for colors: :root { --primary: #xxx; --secondary: #xxx; }

4. RESPONSIVE & ACCESSIBLE:
   - Mobile-first responsive design
   - Proper viewport meta tag
   - Semantic HTML5 elements (header, nav, main, section, footer)
   - Alt text on images, ARIA labels where appropriate

5. INTERACTIVE ELEMENTS:
   - PREFER CSS-only effects (transitions, animations, scroll-snap)
   - For carousels/sliders: Use CSS animation, NOT complex JavaScript
   - AVOID: filtering, sorting, pagination, AJAX loading
   - If using JS: Keep under 30 lines, vanilla only

Output ONLY the HTML code, starting with <!DOCTYPE html>.`,
		ifEmpty(domainInfo, "Professional website"),
		pageInfo,
		ifNotEmpty(sitemapInfo, sitemapInfo),
		ifNotEmpty(designInfo, designInfo),
		ifNotEmpty(architectureInfo, "Site Architecture:\n"+architectureInfo),
		ifNotEmpty(contentInfo, "Content:\n"+contentInfo),
	)
}

// extractDesignContext pulls design info from brief_data or input_data
func extractDesignContext(context map[string]interface{}) string {
	var designParts []string

	// Try to get from input_data paths
	var briefData map[string]interface{}

	if inputData, ok := context["input_data"].(map[string]interface{}); ok {
		// Try reviewed_brief first
		if rb, ok := inputData["reviewed_brief"].(map[string]interface{}); ok {
			briefData = rb
		} else if bd, ok := inputData["brief_data"].(map[string]interface{}); ok {
			// Try brief_data.infer_via_llm.result
			if llmResult, ok := bd["infer_via_llm"].(map[string]interface{}); ok {
				if result, ok := llmResult["result"].(map[string]interface{}); ok {
					briefData = result
				}
			}
		}
	}

	// Also check page_plan for design info
	if pagePlan, ok := context["page_plan"].(map[string]interface{}); ok {
		if planData, ok := pagePlan["plan_data"].(map[string]interface{}); ok {
			if design, ok := planData["design"].(map[string]interface{}); ok {
				if colors, ok := design["color_scheme"].(string); ok && colors != "" {
					designParts = append(designParts, fmt.Sprintf("Color Palette: %s", colors))
				}
				if fonts, ok := design["typography"].(string); ok && fonts != "" {
					designParts = append(designParts, fmt.Sprintf("Typography: %s", fonts))
				}
			}
		}
	}

	if briefData != nil {
		if colorScheme, ok := briefData["color_scheme"].(string); ok && colorScheme != "" {
			// Only add if not already present
			hasColor := false
			for _, part := range designParts {
				if strings.HasPrefix(part, "Color") {
					hasColor = true
					break
				}
			}
			if !hasColor {
				designParts = append(designParts, fmt.Sprintf("Color Palette: %s", colorScheme))
			}
		}
		if tone, ok := briefData["tone"].(string); ok && tone != "" {
			designParts = append(designParts, fmt.Sprintf("Brand Tone: %s", tone))
		}
		if tagline, ok := briefData["tagline"].(string); ok && tagline != "" {
			designParts = append(designParts, fmt.Sprintf("Tagline: %s", tagline))
		}
	}

	if len(designParts) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("=== DESIGN GUIDELINES (use consistently across ALL pages) ===\n")
	for _, part := range designParts {
		result.WriteString(part + "\n")
	}
	result.WriteString("\nMaintain visual consistency:\n")
	result.WriteString("- Same color scheme on every page\n")
	result.WriteString("- Same header/footer layout\n")
	result.WriteString("- Same typography and spacing\n")
	result.WriteString("================================================================\n")

	return result.String()
}

// extractPageInfo extracts current page context
func extractPageInfo(context map[string]interface{}) string {
	if currentPage := context["current_page"]; currentPage != nil {
		switch cp := currentPage.(type) {
		case string:
			if cp != "" {
				return fmt.Sprintf("\nPage: %s", cp)
			}
		case map[string]interface{}:
			var parts []string
			if name, ok := cp["name"].(string); ok && name != "" {
				parts = append(parts, fmt.Sprintf("Page Name: %s", name))
			}
			if title, ok := cp["title"].(string); ok && title != "" {
				parts = append(parts, fmt.Sprintf("Page Title: %s", title))
			}
			if purpose, ok := cp["purpose"].(string); ok && purpose != "" {
				parts = append(parts, fmt.Sprintf("Page Purpose: %s", purpose))
			}
			if sections, ok := cp["sections"].([]interface{}); ok && len(sections) > 0 {
				var sectionNames []string
				for _, s := range sections {
					if sName, ok := s.(string); ok {
						sectionNames = append(sectionNames, sName)
					}
				}
				if len(sectionNames) > 0 {
					parts = append(parts, fmt.Sprintf("Sections: %s", strings.Join(sectionNames, ", ")))
				}
			}
			if metaDesc, ok := cp["meta_description"].(string); ok && metaDesc != "" {
				parts = append(parts, fmt.Sprintf("Meta Description: %s", metaDesc))
			}
			if len(parts) > 0 {
				return "\n" + strings.Join(parts, "\n")
			}
		}
	}
	return ""
}

func buildStructurePrompt(context map[string]interface{}) string {
	domainInfo := extractDomainInfo(context)
	return fmt.Sprintf(`Generate HTML structure skeleton.

Domain: %s

Create:
- <!DOCTYPE html>
- <html lang="en">
- <head> with charset, viewport, title
- <body> with semantic placeholders (<!-- CONTENT_HERE -->)

DO NOT include actual content or CSS.
Return ONLY the HTML structure.`, domainInfo)
}

func buildStylesPrompt(context map[string]interface{}) string {
	domainInfo := extractDomainInfo(context)
	return fmt.Sprintf(`Generate complete CSS for a website.

Domain: %s

Include:
- CSS reset and variables
- Mobile-first responsive design
- Modern typography
- Layout utilities
- Professional color scheme

Return ONLY CSS (or wrapped in <style> tags).`, domainInfo)
}

func buildContentPrompt(context map[string]interface{}) string {
	domainInfo := extractDomainInfo(context)
	contentData := extractContentInfo(context)
	return fmt.Sprintf(`Generate HTML body content.

Domain: %s
Content: %s

Create semantic HTML5 content:
- <header> with navigation
- <main> with sections
- <footer>

Return ONLY body content (no DOCTYPE, html, head, or body tags).`,
		domainInfo,
		contentData,
	)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getMaxTokens resolves this step's output budget through the package's one
// budget ladder (llm_options.go), falling back to defaultVal.
//
// It used to read the BARE `config["max_tokens"]` and nothing else, which made it
// the only reader in the estate that honoured the non-canonical spelling and
// ignored the canonical one — so `html-developer-chunked`'s three steps work
// today only because they happen to be written the way this one function reads.
// Move those numbers into the `ai_service` block they sit beside, as every other
// agent writes them, and this function would silently have returned 16000 for all
// three (bugs_open/257 round 3, owner decision 4). Reading both spellings, in the
// same order as every other call site, is what makes that migration safe.
//
// The value is then handed to ExecuteLLMPromptAction inside a synthesised
// ai_service block (see GenerateHTMLAction), so the canonical ladder sees it at
// its most specific level and agrees. ⚠ That synthesis also hardcodes the model
// and provider, discarding whatever the step declared — a separate defect, out of
// this change's scope, recorded here because the next reader of this function
// will wonder.
func getMaxTokens(config map[string]interface{}, defaultVal int) int {
	if value, from, _ := resolveBudgetKey("max_tokens", []budgetLevel{
		{subConfigMap(config, "ai_service"), "step_config.ai_service"},
		{config, "step_config"},
	}); from != "" {
		return value
	}
	return defaultVal
}

func extractDomainInfo(context map[string]interface{}) string {
	// Check root level first (after flattening from unified extractor)
	if domain, ok := context["domain"].(string); ok && domain != "" {
		return domain
	}
	if bizName, ok := context["business_name"].(string); ok && bizName != "" {
		return bizName
	}

	// Check nested fields
	possibleFields := []string{"input_data", "business", "domain_analysis"}
	for _, fieldName := range possibleFields {
		if val, ok := context[fieldName]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				if domain, ok := m["domain"].(string); ok && domain != "" {
					return domain
				}
				if bizName, ok := m["business_name"].(string); ok && bizName != "" {
					return bizName
				}
			}
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}
	return "this website"
}

func extractContentInfo(context map[string]interface{}) string {
	if content, ok := context["site_content"]; ok {
		if str, ok := content.(string); ok {
			return str
		}
		if m, ok := content.(map[string]interface{}); ok {
			if contentJSON, err := json.Marshal(m); err == nil {
				if len(contentJSON) > 2000 {
					return string(contentJSON[:2000]) + "..."
				}
				return string(contentJSON)
			}
		}
	}
	return "professional content appropriate for the business"
}

func ifEmpty(s, defaultVal string) string {
	if s == "" {
		return defaultVal
	}
	return s
}

func ifNotEmpty(s, prefix string) string {
	if s == "" {
		return ""
	}
	return prefix
}

// extractSitemapInfo extracts navigation links from database sync or sitemap
// Sources checked (in order of priority):
// 1. db_sync.navigation (from sync_pages_to_db - CANONICAL DATABASE SOURCE)
// 2. link_data.navigation (from link-manager agent)
// 3. page_plan.plan_data.sitemap (from strategist)
// 4. sitemap (direct field)
func extractSitemapInfo(context map[string]interface{}) string {
	// PRIORITY 1: Check for db_sync.navigation (from sync_pages_to_db)
	// This is the CANONICAL source - generated from database, all pages should use this
	if dbSync, ok := context["db_sync"].(map[string]interface{}); ok {
		if nav := extractNavigationFromDBSync(dbSync); nav != "" {
			return nav
		}
	}

	// PRIORITY 2: Check for link_data.navigation (from link-manager)
	if linkData, ok := context["link_data"].(map[string]interface{}); ok {
		// Try navigation structure first
		if nav, ok := linkData["navigation"].(map[string]interface{}); ok {
			return formatNavigationFromLinkManager(nav)
		}
		// Try link_registry as fallback
		if registry, ok := linkData["link_registry"].(map[string]interface{}); ok {
			return formatNavigationFromRegistry(registry)
		}
	}

	// PRIORITY 3: Check for sitemap from various sources
	var sitemap []interface{}

	// Try direct sitemap field first
	if sm, ok := context["sitemap"].([]interface{}); ok {
		sitemap = sm
	}

	// Try nested in page_plan.plan_data.sitemap
	if sitemap == nil {
		if pagePlan, ok := context["page_plan"].(map[string]interface{}); ok {
			if planData, ok := pagePlan["plan_data"].(map[string]interface{}); ok {
				if sm, ok := planData["sitemap"].([]interface{}); ok {
					sitemap = sm
				}
			}
			// Also try direct result under page_plan
			if sitemap == nil {
				if sm, ok := pagePlan["sitemap"].([]interface{}); ok {
					sitemap = sm
				}
			}
		}
	}

	// Try nested in plan_data.sitemap (if page_plan was flattened)
	if sitemap == nil {
		if planData, ok := context["plan_data"].(map[string]interface{}); ok {
			if sm, ok := planData["sitemap"].([]interface{}); ok {
				sitemap = sm
			}
		}
	}

	if sitemap == nil || len(sitemap) == 0 {
		return ""
	}

	return formatNavigationFromSitemap(sitemap)
}

// extractNavigationFromDBSync extracts navigation from the db_sync output
// The db_sync.navigation structure is: {"items": [{"label": "Home", "url": "/index.html"}, ...]}
func extractNavigationFromDBSync(dbSync map[string]interface{}) string {
	navData, ok := dbSync["navigation"]
	if !ok {
		return ""
	}

	var items []interface{}

	// Handle map with items array (NavigationStructure format after JSON serialization)
	if nav, ok := navData.(map[string]interface{}); ok {
		if itemsRaw, ok := nav["items"].([]interface{}); ok {
			items = itemsRaw
		}
	}

	if len(items) == 0 {
		return ""
	}

	// Format as navigation string for the prompt
	var headerNav []string
	for _, item := range items {
		if itemMap, ok := item.(map[string]interface{}); ok {
			label, _ := itemMap["label"].(string)
			url, _ := itemMap["url"].(string)
			if label != "" && url != "" {
				headerNav = append(headerNav, fmt.Sprintf("%s -> %s", label, url))
			}
		}
	}

	if len(headerNav) == 0 {
		return ""
	}

	// Build navigation string with STRONG enforcement language
	var result strings.Builder
	result.WriteString("=== SITE NAVIGATION (MANDATORY - USE EXACTLY AS PROVIDED) ===\n")
	result.WriteString("Header navigation items (use in this exact order):\n")
	for _, nav := range headerNav {
		result.WriteString("  • " + nav + "\n")
	}
	result.WriteString("\n")
	result.WriteString("STRICT NAVIGATION REQUIREMENTS:\n")
	result.WriteString("• Include ALL of these navigation items in the header\n")
	result.WriteString("• Use them in EXACTLY this order (left to right)\n")
	result.WriteString("• Do NOT add any additional navigation items\n")
	result.WriteString("• Do NOT remove any navigation items\n")
	result.WriteString("• Use the EXACT URLs shown (do not modify)\n")
	result.WriteString("• This ensures all pages have consistent navigation\n")
	result.WriteString("=============================================================\n")

	return result.String()
}

// formatNavigationFromLinkManager formats nav from link-manager's navigation structure
// Expected format: {"header": {"items": [...]}, "footer": {"columns": [...]}}
func formatNavigationFromLinkManager(nav map[string]interface{}) string {
	var headerNav []string
	var footerNav []string

	// Extract header items
	if header, ok := nav["header"].(map[string]interface{}); ok {
		if items, ok := header["items"].([]interface{}); ok {
			for _, item := range items {
				if itemMap, ok := item.(map[string]interface{}); ok {
					label, _ := itemMap["label"].(string)
					url, _ := itemMap["url"].(string)
					if label != "" && url != "" {
						headerNav = append(headerNav, fmt.Sprintf("%s -> %s", label, url))
					}
				}
			}
		}
	}

	// Extract footer items from columns
	if footer, ok := nav["footer"].(map[string]interface{}); ok {
		if columns, ok := footer["columns"].([]interface{}); ok {
			for _, col := range columns {
				if colMap, ok := col.(map[string]interface{}); ok {
					if links, ok := colMap["links"].([]interface{}); ok {
						for _, link := range links {
							if linkMap, ok := link.(map[string]interface{}); ok {
								label, _ := linkMap["label"].(string)
								url, _ := linkMap["url"].(string)
								if label != "" && url != "" {
									footerNav = append(footerNav, fmt.Sprintf("%s -> %s", label, url))
								}
							}
						}
					}
				}
			}
		}
	}

	if len(headerNav) == 0 && len(footerNav) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("=== SITE NAVIGATION (FROM LINK MANAGER) ===\n")

	if len(headerNav) > 0 {
		result.WriteString("Header navigation (use in this exact order):\n")
		for _, nav := range headerNav {
			result.WriteString("  • " + nav + "\n")
		}
	}

	if len(footerNav) > 0 {
		result.WriteString("Footer navigation:\n")
		for _, nav := range footerNav {
			result.WriteString("  • " + nav + "\n")
		}
	}

	result.WriteString("============================================\n")
	return result.String()
}

// formatNavigationFromRegistry formats nav from link registry
// This is a fallback when navigation structure isn't available
func formatNavigationFromRegistry(registry map[string]interface{}) string {
	var navLinks []string

	// Registry format: {"links": [...]} or direct array of link objects
	var links []interface{}

	if linkArr, ok := registry["links"].([]interface{}); ok {
		links = linkArr
	} else if linkArr, ok := registry["internal_links"].([]interface{}); ok {
		links = linkArr
	}

	for _, link := range links {
		if linkMap, ok := link.(map[string]interface{}); ok {
			// Check if it's a navigation link
			linkType, _ := linkMap["link_type"].(string)
			if linkType != "navigation" && linkType != "" {
				continue
			}

			anchorText, _ := linkMap["anchor_text"].(string)
			targetURL, _ := linkMap["target_url"].(string)

			if anchorText == "" {
				anchorText, _ = linkMap["label"].(string)
			}
			if targetURL == "" {
				targetURL, _ = linkMap["url"].(string)
			}

			if anchorText != "" && targetURL != "" {
				navLinks = append(navLinks, fmt.Sprintf("%s -> %s", anchorText, targetURL))
			}
		}
	}

	if len(navLinks) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("NAVIGATION (from link registry):\n")
	result.WriteString("Navigation links: ")
	result.WriteString(strings.Join(navLinks, " | "))
	result.WriteString("\n")

	return result.String()
}

// formatNavigationFromSitemap formats navigation from strategist's sitemap
// Expected format: [{"label": "Home", "url": "/index.html", "in_header": true}, ...]
func formatNavigationFromSitemap(sitemap []interface{}) string {
	var headerNav []string
	var footerNav []string

	for _, entry := range sitemap {
		if e, ok := entry.(map[string]interface{}); ok {
			label, _ := e["label"].(string)
			url, _ := e["url"].(string)

			// Default to header if not specified
			inHeader := true
			if ih, ok := e["in_header"].(bool); ok {
				inHeader = ih
			}
			inFooter, _ := e["in_footer"].(bool)

			if label == "" {
				// Try alternative field names
				label, _ = e["name"].(string)
			}
			if label == "" {
				label, _ = e["title"].(string)
			}

			if url == "" {
				// Try to construct URL from name
				if name, ok := e["name"].(string); ok && name != "" {
					if name == "index" || name == "home" {
						url = "/index.html"
					} else {
						url = "/" + name + ".html"
					}
				}
			}

			if label == "" || url == "" {
				continue
			}

			linkStr := fmt.Sprintf("%s -> %s", label, url)
			if inHeader {
				headerNav = append(headerNav, linkStr)
			}
			if inFooter {
				footerNav = append(footerNav, linkStr)
			}
		}
	}

	if len(headerNav) == 0 && len(footerNav) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("=== SITE NAVIGATION (USE EXACTLY AS PROVIDED) ===\n")

	if len(headerNav) > 0 {
		result.WriteString("Header navigation (use in this exact order):\n")
		for _, nav := range headerNav {
			result.WriteString("  • " + nav + "\n")
		}
	}

	if len(footerNav) > 0 {
		result.WriteString("Footer navigation:\n")
		for _, nav := range footerNav {
			result.WriteString("  • " + nav + "\n")
		}
	}

	result.WriteString("\n")
	result.WriteString("STRICT REQUIREMENTS:\n")
	result.WriteString("• Include ALL navigation items in the header\n")
	result.WriteString("• Use EXACTLY this order (left to right)\n")
	result.WriteString("• Do NOT add or remove items\n")
	result.WriteString("================================================\n")

	return result.String()
}

func shouldMinify(params ActionParams) bool {
	if config, ok := params.StepConfig.Config["minify"].(bool); ok {
		return config
	}
	return true
}
