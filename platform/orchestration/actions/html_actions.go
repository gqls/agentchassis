// FILE: platform/orchestration/actions/html_actions.go
package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"go.uber.org/zap"
)

// GenerateHTMLAction generates HTML using LLM based on collected data
func GenerateHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Generating HTML content")

	// Gather context from previous steps
	context := gatherHTMLContext(params.CollectedData)

	// Now, it gets its *own* configuration from the standard location.
	agentConfig, ok := params.CollectedData["agent_config"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("agent_config not found in CollectedData")
	}

	// Build the prompt
	prompt := buildHTMLPrompt(context, agentConfig)

	// Call LLM
	// It then calls the LLM action, passing the generated prompt in the
	// standardized "input_data" field for the next action.
	llmParams := params // copy params to avoid mutation
	llmParams.CollectedData["input_data"] = map[string]interface{}{
		"prompt": prompt,
	}
	llmParams.StepConfig.Config = map[string]interface{}{
		"model":      "claude-3-opus-20240229",
		"max_tokens": 8000,
	}

	result, err := ExecuteLLMPromptAction(ctx, llmParams)
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTML: %w", err)
	}

	// Extract HTML from response
	htmlContent := extractHTMLFromResponse(result)

	return map[string]interface{}{
		"raw_html":     htmlContent,
		"generated_at": time.Now().UTC(),
		"prompt_used":  prompt,
	}, nil
}

// ProcessHTMLAction processes and enhances generated HTML
func ProcessHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Processing HTML content")

	// Get HTML from previous step
	var rawHTML string
	if genResult, ok := params.CollectedData["generate_html"].(map[string]interface{}); ok {
		rawHTML, _ = genResult["raw_html"].(string)
	} else if content, ok := params.CollectedData["raw_html"].(string); ok {
		rawHTML = content
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
	businessInfo := extractBusinessInfo(params.CollectedData)

	// Process based on configuration
	processingSteps := []string{}

	// Ensure proper structure
	ensureHTMLStructure(doc)
	processingSteps = append(processingSteps, "structure_validation")

	// Add meta tags
	addMetaTags(doc, businessInfo)
	processingSteps = append(processingSteps, "meta_tags")

	// Ensure responsive design
	ensureResponsiveDesign(doc)
	processingSteps = append(processingSteps, "responsive_design")

	// Optimize images
	optimizeImages(doc)
	processingSteps = append(processingSteps, "image_optimization")

	// Add structured data
	addStructuredData(doc, businessInfo)
	processingSteps = append(processingSteps, "structured_data")

	// Get processed HTML
	processedHTML, _ := doc.Html()

	// Minify if needed
	if shouldMinify(params) {
		processedHTML = minifyHTML(processedHTML, params.Logger)
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

	// Get processed HTML
	var htmlContent string
	if procResult, ok := params.CollectedData["process_html"].(map[string]interface{}); ok {
		htmlContent, _ = procResult["processed_html"].(string)
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
		"element_count": countElements(doc),
		"final_html":    finalHTML,
	}, nil
}

// Helper functions

func extractHTMLFromResponse(result interface{}) string {
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return ""
	}

	response, _ := resultMap["result"].(string)

	// Look for HTML in code blocks
	htmlBlockRe := regexp.MustCompile("```html\\s*([\\s\\S]*?)```")
	matches := htmlBlockRe.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Look for DOCTYPE or html tag
	if strings.Contains(response, "<!DOCTYPE") || strings.Contains(response, "<html") {
		startIdx := strings.Index(response, "<!DOCTYPE")
		if startIdx == -1 {
			startIdx = strings.Index(response, "<html")
		}
		if startIdx >= 0 {
			return response[startIdx:]
		}
	}

	return response
}

func gatherHTMLContext(collectedData map[string]interface{}) map[string]interface{} {
	context := make(map[string]interface{})

	// Extract domain analysis
	if domainData, ok := collectedData["analyze_domain"].(map[string]interface{}); ok {
		context["domain_analysis"] = domainData
	}

	// Extract site architecture
	if archData, ok := collectedData["architect_site"].(map[string]interface{}); ok {
		context["site_structure"] = archData
	}

	// Extract content
	if contentData, ok := collectedData["create_content"].(map[string]interface{}); ok {
		context["content"] = contentData
	}

	// Extract business info
	context["business"] = extractBusinessInfo(collectedData)

	return context
}

func extractBusinessInfo(collectedData map[string]interface{}) map[string]interface{} {
	info := make(map[string]interface{})

	// Try to get from input data
	if inputData, ok := collectedData["input_data"].(map[string]interface{}); ok {
		if businessName, ok := inputData["business_name"].(string); ok {
			info["business_name"] = businessName
		}
		if domain, ok := inputData["domain"].(string); ok {
			info["domain"] = domain
		}
		if desc, ok := inputData["description"].(string); ok {
			info["description"] = desc
		}
	}

	// Try to get from headers
	if headers, ok := collectedData["headers"].(map[string]interface{}); ok {
		if clientID, ok := headers["client_id"].(string); ok {
			info["client_id"] = clientID
		}
	}

	return info
}

func buildHTMLPrompt(context map[string]interface{}, agentConfig map[string]interface{}) string {
	// todo bring in agentConfig
	contextJSON, _ := json.MarshalIndent(context, "", "  ")

	return fmt.Sprintf(`Generate a complete, modern, responsive HTML website based on the following context:

%s

Requirements:
1. Create a complete HTML5 document with proper structure
2. Include inline CSS for styling (modern, clean design)
3. Make it fully responsive with mobile-first approach
4. Include proper meta tags for SEO
5. Use semantic HTML elements
6. Include a navigation menu
7. Create sections based on the site structure provided
8. Make it production-ready

Output only the HTML code, starting with <!DOCTYPE html>.`, string(contextJSON))
}

func ensureHTMLStructure(doc *goquery.Document) {
	// Ensure DOCTYPE (goquery doesn't preserve it, we'll add it back later)

	// Ensure html element has lang attribute
	html := doc.Find("html")
	if html.Length() > 0 {
		if lang, exists := html.Attr("lang"); !exists || lang == "" {
			html.SetAttr("lang", "en")
		}
	}

	// Ensure head exists
	if doc.Find("head").Length() == 0 {
		doc.Find("html").PrependHtml("<head></head>")
	}

	// Ensure body exists
	if doc.Find("body").Length() == 0 {
		doc.Find("html").AppendHtml("<body></body>")
	}
}

func addMetaTags(doc *goquery.Document, businessInfo map[string]interface{}) {
	head := doc.Find("head")

	// Charset
	if doc.Find("meta[charset]").Length() == 0 {
		head.PrependHtml(`<meta charset="UTF-8">`)
	}

	// Viewport
	if doc.Find("meta[name='viewport']").Length() == 0 {
		head.AppendHtml(`<meta name="viewport" content="width=device-width, initial-scale=1.0">`)
	}

	// Description
	if desc, ok := businessInfo["description"].(string); ok && desc != "" {
		if doc.Find("meta[name='description']").Length() == 0 {
			head.AppendHtml(fmt.Sprintf(`<meta name="description" content="%s">`, desc))
		}
	}

	// Open Graph
	if name, ok := businessInfo["business_name"].(string); ok && name != "" {
		head.AppendHtml(fmt.Sprintf(`<meta property="og:title" content="%s">`, name))
		head.AppendHtml(`<meta property="og:type" content="website">`)

		if desc, ok := businessInfo["description"].(string); ok {
			head.AppendHtml(fmt.Sprintf(`<meta property="og:description" content="%s">`, desc))
		}
	}
}

func ensureResponsiveDesign(doc *goquery.Document) {
	// Check if viewport meta exists
	if doc.Find("meta[name='viewport']").Length() == 0 {
		doc.Find("head").AppendHtml(`<meta name="viewport" content="width=device-width, initial-scale=1.0">`)
	}

	// Add responsive CSS if not present
	style := doc.Find("style")
	if style.Length() == 0 {
		doc.Find("head").AppendHtml(`
        <style>
            * { box-sizing: border-box; margin: 0; padding: 0; }
            body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif; line-height: 1.6; }
            img { max-width: 100%; height: auto; }
            .container { max-width: 1200px; margin: 0 auto; padding: 0 20px; }
            @media (max-width: 768px) {
                .container { padding: 0 15px; }
            }
        </style>`)
	}
}

func optimizeImages(doc *goquery.Document) {
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		// Add loading="lazy" for images not in viewport
		if loading, exists := s.Attr("loading"); !exists || loading == "" {
			s.SetAttr("loading", "lazy")
		}

		// Ensure images have dimensions to prevent layout shift
		if _, exists := s.Attr("width"); !exists {
			s.SetAttr("width", "auto")
		}
		if _, exists := s.Attr("height"); !exists {
			s.SetAttr("height", "auto")
		}
	})
}

func addStructuredData(doc *goquery.Document, businessInfo map[string]interface{}) {
	if name, ok := businessInfo["business_name"].(string); ok {
		structuredData := map[string]interface{}{
			"@context": "https://schema.org",
			"@type":    "Organization",
			"name":     name,
		}

		if desc, ok := businessInfo["description"].(string); ok {
			structuredData["description"] = desc
		}

		if domain, ok := businessInfo["domain"].(string); ok {
			structuredData["url"] = "https://" + domain
		}

		jsonLD, _ := json.Marshal(structuredData)
		doc.Find("head").AppendHtml(fmt.Sprintf(`<script type="application/ld+json">%s</script>`, string(jsonLD)))
	}
}

func minifyHTML(htmlContent string, logger *zap.Logger) string {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("application/javascript", js.Minify)

	minified, err := m.String("text/html", htmlContent)
	if err != nil {
		logger.Warn("Failed to minify HTML", zap.Error(err))
		return htmlContent
	}

	// Ensure DOCTYPE is preserved
	if !strings.HasPrefix(minified, "<!DOCTYPE") {
		minified = "<!DOCTYPE html>" + minified
	}

	return minified
}

func shouldMinify(params ActionParams) bool {
	if config, ok := params.StepConfig.Config["minify"].(bool); ok {
		return config
	}
	return true // Default to minifying
}

func countElements(doc *goquery.Document) map[string]int {
	return map[string]int{
		"images":   doc.Find("img").Length(),
		"links":    doc.Find("a").Length(),
		"headings": doc.Find("h1,h2,h3,h4,h5,h6").Length(),
		"sections": doc.Find("section,article,main,aside").Length(),
	}
}
