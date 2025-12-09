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
		zap.Any("collected_data_keys", GetMapKeys(params.CollectedData)),
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
		zap.Any("context_keys", GetMapKeys(context)),
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
		zap.Strings("context_keys", getMapKeys(context)),
	)

	return context
}

// ensureCoreDomainInfo searches for domain/objective anywhere in the structure
func ensureCoreDomainInfo(context map[string]interface{}, collectedData map[string]interface{}, logger *zap.Logger) {
	// Check if input_data already has domain/objective
	if inputData, ok := context["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && domain != "" {
			if objective, ok := inputData["objective"].(string); ok && objective != "" {
				logger.Info("Domain and objective already in input_data")
				return // Already have both
			}
		}
	}

	// Use aggressive search to find domain and objective
	logger.Info("Using aggressive search to find core data")
	coreData := datahelpers.ExtractCoreInputData(collectedData, logger)

	if len(coreData) > 0 {
		logger.Info("Found core data via aggressive search",
			zap.Any("found_fields", getMapKeys(coreData)),
		)

		// Ensure input_data exists in context
		if context["input_data"] == nil {
			context["input_data"] = make(map[string]interface{})
		}

		// Merge found data into input_data
		if inputDataMap, ok := context["input_data"].(map[string]interface{}); ok {
			for key, val := range coreData {
				if inputDataMap[key] == nil || inputDataMap[key] == "" {
					inputDataMap[key] = val
					logger.Info("Added field to input_data",
						zap.String("field", key),
						zap.Int("value_length", len(fmt.Sprintf("%v", val))),
					)
				}
			}
		}
	} else {
		logger.Warn("Aggressive search found no core data")
	}
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

func buildFullHTMLPrompt(context map[string]interface{}) string {
	var domainInfo, architectureInfo, contentInfo string

	// Extract domain/business info - check root level first (after flattening)
	if domain, ok := context["domain"].(string); ok && domain != "" {
		domainInfo = domain
	} else if bizName, ok := context["business_name"].(string); ok && bizName != "" {
		domainInfo = bizName
	} else if inputData, ok := context["input_data"].(map[string]interface{}); ok {
		// Fallback: check inside input_data if not flattened
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

	// Extract content
	if content, ok := context["site_content"]; ok {
		if contentStr, ok := content.(string); ok {
			contentInfo = contentStr
		} else if contentJSON, err := json.Marshal(content); err == nil {
			if len(contentJSON) > 3000 {
				contentInfo = string(contentJSON[:3000]) + "... [truncated]"
			} else {
				contentInfo = string(contentJSON)
			}
		}
	}

	return fmt.Sprintf(`Generate a complete, modern, responsive HTML5 website.

Business/Domain: %s

%s

%s

Requirements:
1. Complete HTML5 document (<!DOCTYPE html> to </html>)
2. Modern, clean design with inline CSS
3. Fully responsive (mobile-first)
4. Proper meta tags (charset, viewport, description)
5. Semantic HTML5 elements
6. Professional, production-ready

Output ONLY the HTML code, starting with <!DOCTYPE html>.`,
		ifEmpty(domainInfo, "Professional website"),
		ifNotEmpty(architectureInfo, "Site Architecture:\n"+architectureInfo),
		ifNotEmpty(contentInfo, "Content:\n"+contentInfo),
	)
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

func getMaxTokens(config map[string]interface{}, defaultVal int) int {
	if maxTokensRaw, ok := config["max_tokens"]; ok {
		if tokens, ok := maxTokensRaw.(float64); ok {
			return int(tokens)
		} else if tokens, ok := maxTokensRaw.(int); ok {
			return tokens
		}
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

func shouldMinify(params ActionParams) bool {
	if config, ok := params.StepConfig.Config["minify"].(bool); ok {
		return config
	}
	return true
}
