// FILE: platform/orchestration/actions/html_actions_configurable.go
// CORRECT VERSION: Uses input_fields config instead of hardcoded field names

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
// This properly uses workflow field mappings instead of hardcoding field names
func GenerateHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Generating HTML content",
		zap.Any("collected_data_keys", GetMapKeys(params.CollectedData)),
		//zap.Any("DEBUGaa: collected_data", params.CollectedData),
	)

	params.Logger.Info("Checking for input fields",
		zap.Bool("has_input_data", params.CollectedData["input_data"] != nil),
		zap.Bool("has_site_architecture", params.CollectedData["site_architecture"] != nil),
		zap.Bool("has_site_content", params.CollectedData["site_content"] != nil),
	)

	// Get configuration
	config := params.StepConfig.Config
	generationType, _ := config["generation_type"].(string)
	maxTokens := getMaxTokens(config, 16000)

	// Get input_fields from config (like call_agent does)
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

	params.Logger.Info("generate html action Using input fields",
		zap.Strings("input_fields", inputFields),
		zap.String("generation_type", generationType),
	)

	// Build context from specified input_fields
	context := buildContextFromInputFields(params.CollectedData, inputFields, params.Logger)

	// Build prompt based on generation type
	var prompt string
	switch generationType {
	case "structure":
		prompt = buildStructurePrompt(context)
		params.Logger.Info("Generating HTML structure")
	case "styles":
		prompt = buildStylesPrompt(context)
		params.Logger.Info("Generating CSS styles")
	case "content":
		prompt = buildContentPrompt(context)
		params.Logger.Info("Generating HTML content")
	default:
		prompt = buildFullHTMLPrompt(context)
		params.Logger.Info("Generating complete HTML")
	}

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

	// Extract HTML from response
	htmlContent := datahelpers.ExtractHTMLFromResponse(result)

	if htmlContent == "" {
		params.Logger.Warn("LLM returned empty HTML",
			zap.String("generation_type", generationType),
		)
	}

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

	// Get HTML from previous step (checking for response field)
	var rawHTML string

	// First check generate_html step
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

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(rawHTML))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Get business context
	businessInfo := datahelpers.ExtractBusinessInfo(params.CollectedData)

	// Process based on configuration
	processingSteps := []string{}

	// Ensure proper structure
	datahelpers.EnsureHTMLStructure(doc)
	processingSteps = append(processingSteps, "structure_validation")

	// Add meta tags
	datahelpers.AddMetaTags(doc, businessInfo)
	processingSteps = append(processingSteps, "meta_tags")

	// Ensure responsive design
	datahelpers.EnsureResponsiveDesign(doc)
	processingSteps = append(processingSteps, "responsive_design")

	// Optimize images
	datahelpers.OptimizeImages(doc)
	processingSteps = append(processingSteps, "image_optimization")

	// Add structured data
	datahelpers.AddStructuredData(doc, businessInfo)
	processingSteps = append(processingSteps, "structured_data")

	// Get processed HTML
	processedHTML, _ := doc.Html()

	// Minify if needed
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

	// Get processed HTML (checking for response field)
	var htmlContent string

	// Check process_html step
	if procStep, ok := params.CollectedData["process_html"]; ok {
		extracted := datahelpers.ExtractStepData(procStep)
		if procResult, ok := extracted.(map[string]interface{}); ok {
			htmlContent, _ = procResult["processed_html"].(string)
		}
	}

	if htmlContent == "" {
		return nil, fmt.Errorf("no HTML content to validate")
	}

	// Parse for validation
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return map[string]interface{}{
			"valid":  false,
			"errors": []string{fmt.Sprintf("Failed to parse HTML: %v", err)},
		}, nil
	}

	errors := []string{}
	warnings := []string{}

	// Required elements
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

	// Check meta tags
	if doc.Find("meta[charset]").Length() == 0 {
		warnings = append(warnings, "Missing charset meta tag")
	}

	if doc.Find("meta[name='viewport']").Length() == 0 {
		warnings = append(warnings, "Missing viewport meta tag")
	}

	// Check images
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

	// Check links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if href == "" || href == "#" {
			warnings = append(warnings, fmt.Sprintf("Link %d has empty or placeholder href", i+1))
		}
	})

	isValid := len(errors) == 0

	// Store the final HTML if valid
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

// buildContextFromInputFields extracts data from CollectedData based on input_fields
// This is the proper way - let the workflow configure which fields to use
func buildContextFromInputFields(collectedData map[string]interface{}, inputFields []string, logger *zap.Logger) map[string]interface{} {
	context := make(map[string]interface{})

	if len(inputFields) == 0 {
		// Fallback: if no input_fields specified, try common names
		// But log a warning - this shouldn't happen in properly configured workflows
		logger.Warn("No input_fields specified, using fallback field names")
		inputFields = []string{"input_data", "site_architecture", "site_content", "domain_analysis"}
	}

	for _, fieldPath := range inputFields {
		value := extractFieldByPath(collectedData, fieldPath)
		if value != nil {
			// Store by the field name (last part of path)
			parts := strings.Split(fieldPath, ".")
			fieldName := parts[len(parts)-1]
			context[fieldName] = value

			logger.Debug("Added field to context",
				zap.String("field_path", fieldPath),
				zap.String("stored_as", fieldName),
			)
		}
	}

	return context
}

// extractFieldByPath navigates nested field paths like "step_name.result"
func extractFieldByPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
			} else {
				// Try ExtractStepData for wrapped responses
				if extracted := datahelpers.ExtractStepData(v[part]); extracted != nil {
					current = extracted
				} else {
					return nil
				}
			}
		default:
			return nil
		}
	}

	return current
}

// buildFullHTMLPrompt creates prompt for complete HTML generation
func buildFullHTMLPrompt(context map[string]interface{}) string {
	// Extract available data (might have different field names depending on workflow)
	var domainInfo, architectureInfo, contentInfo string

	// Try to find domain/business info
	if inputData, ok := context["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok {
			domainInfo = domain
		} else if bizName, ok := inputData["business_name"].(string); ok {
			domainInfo = bizName
		}
	}

	// Try to find architecture/structure
	if arch, ok := context["site_architecture"]; ok {
		if archStr, ok := arch.(string); ok {
			architectureInfo = archStr
		} else if archJSON, err := json.Marshal(arch); err == nil {
			architectureInfo = string(archJSON)
		}
	}

	// Try to find content
	if content, ok := context["site_content"]; ok {
		if contentStr, ok := content.(string); ok {
			contentInfo = contentStr
		} else if contentJSON, err := json.Marshal(content); err == nil {
			// Truncate if too large
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

// buildStructurePrompt creates prompt for HTML structure only
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

// buildStylesPrompt creates prompt for CSS only
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

// buildContentPrompt creates prompt for HTML content only
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

// Helper functions

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
	// Try various field names that might contain domain info
	possibleFields := []string{"input_data", "domain", "business", "domain_analysis"}

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
	// Try to find content field
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
	return true // Default to minifying
}
