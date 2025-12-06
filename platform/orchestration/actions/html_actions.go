// FILE: platform/orchestration/actions/html_actions_enhanced.go
// ENHANCEMENT: Adds generation_type support to GenerateHTMLAction
// This allows chunked generation: structure, styles, or content separately

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
// Enhanced to support chunked generation via generation_type config
func GenerateHTMLAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("Generating HTML content")

	// Get configuration
	config := params.StepConfig.Config
	generationType, _ := config["generation_type"].(string)
	maxTokensRaw, hasMaxTokens := config["max_tokens"]

	// Default max tokens based on generation type
	maxTokens := 16000
	if hasMaxTokens {
		if tokens, ok := maxTokensRaw.(float64); ok {
			maxTokens = int(tokens)
		} else if tokens, ok := maxTokensRaw.(int); ok {
			maxTokens = tokens
		}
	}

	// Gather context from previous steps
	context := gatherHTMLContext(params.CollectedData)

	// Get agent config
	var agentConfig map[string]interface{}
	if configData, ok := params.CollectedData["agent_config"]; ok {
		extracted := ExtractStepData(configData)
		agentConfig, ok = extracted.(map[string]interface{})
		if !ok {
			agentConfig = make(map[string]interface{})
		}
	} else {
		agentConfig = make(map[string]interface{})
	}

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
		prompt = buildHTMLPrompt(context, agentConfig)
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
	htmlContent := extractHTMLFromResponse(result)

	return map[string]interface{}{
		"raw_html":        htmlContent,
		"generation_type": generationType,
		"generated_at":    time.Now().UTC(),
		"prompt_used":     prompt,
		"tokens_used":     maxTokens,
	}, nil
}

// buildStructurePrompt creates a prompt for generating HTML structure only
func buildStructurePrompt(context map[string]interface{}) string {
	domainInfo := extractDomainFromContext(context)

	return fmt.Sprintf(`Generate the HTML structure skeleton for a website.

Domain/Business: %s

Create a basic HTML5 structure with:
- DOCTYPE declaration
- HTML element with lang attribute
- Head section with:
  - charset meta tag (UTF-8)
  - viewport meta tag
  - title tag
  - Link to external stylesheet (style.css)
- Body element with semantic structure:
  - header with nav placeholder
  - main with section placeholders
  - footer placeholder

IMPORTANT: 
- Include comment placeholders like <!-- CONTENT_HERE -->
- DO NOT include any actual content or CSS
- Keep it minimal and semantic

Return ONLY the HTML structure, nothing else.`, domainInfo)
}

// buildStylesPrompt creates a prompt for generating CSS only
func buildStylesPrompt(context map[string]interface{}) string {
	domainInfo := extractDomainFromContext(context)
	archInfo := extractArchitectureFromContext(context)

	return fmt.Sprintf(`Generate complete CSS for a website.

Domain/Business: %s
Architecture: %s

Create modern, production-ready CSS with:
- CSS reset and box-sizing
- CSS custom properties for colors, fonts, spacing
- Mobile-first responsive design
- Typography system
- Layout utilities (flexbox/grid)
- Component styles (header, nav, footer, sections)
- Responsive breakpoints (mobile, tablet, desktop)

Requirements:
- Modern, clean aesthetic
- Professional color scheme
- Accessible (WCAG AA)
- Performance-optimized

Return ONLY CSS (no HTML, no markdown code blocks).
If you want to wrap it, use <style> tags.`, domainInfo, archInfo)
}

// buildContentPrompt creates a prompt for generating HTML content only
func buildContentPrompt(context map[string]interface{}) string {
	domainInfo := extractDomainFromContext(context)
	contentData := extractContentFromContext(context)
	archInfo := extractArchitectureFromContext(context)

	return fmt.Sprintf(`Generate HTML content for a website.

Domain/Business: %s
Architecture: %s
Content to include: %s

Create semantic HTML5 content including:
- Header with navigation menu
- Hero section with compelling value proposition
- Main content sections based on architecture
- Footer with links and information

Requirements:
- Use semantic HTML5 elements (header, nav, main, section, article, footer)
- Include appropriate headings (h1-h6)
- Make content engaging and professional
- Include calls-to-action where appropriate

IMPORTANT:
- Return ONLY the body content (no <!DOCTYPE>, <html>, <head>, or <body> tags)
- Start with <header> and end with </footer>
- DO NOT include CSS (styles will be added separately)

Return ONLY the HTML content elements.`, domainInfo, archInfo, contentData)
}

// Helper functions to extract specific data from context

func extractDomainFromContext(context map[string]interface{}) string {
	// Try to get domain from various places
	if business, ok := context["business"].(map[string]interface{}); ok {
		if domain, ok := business["domain"].(string); ok {
			return domain
		}
		if name, ok := business["business_name"].(string); ok {
			return name
		}
	}

	if domainAnalysis, ok := context["domain_analysis"].(map[string]interface{}); ok {
		if domain, ok := domainAnalysis["domain"].(string); ok {
			return domain
		}
	}

	return "this website"
}

func extractArchitectureFromContext(context map[string]interface{}) string {
	if siteStructure, ok := context["site_structure"].(map[string]interface{}); ok {
		// Try to get a string representation
		if arch, ok := siteStructure["architecture"].(string); ok {
			return arch
		}
		// Try to stringify the whole structure
		if archJSON, err := json.Marshal(siteStructure); err == nil {
			return string(archJSON)
		}
	}
	return "standard website structure with header, main content, and footer"
}

func extractContentFromContext(context map[string]interface{}) string {
	if content, ok := context["content"].(map[string]interface{}); ok {
		// Try to get string representation
		if contentStr, ok := content["content"].(string); ok {
			return contentStr
		}
		// Try to stringify
		if contentJSON, err := json.Marshal(content); err == nil {
			// Truncate if too long
			if len(contentJSON) > 2000 {
				return string(contentJSON[:2000]) + "... [truncated]"
			}
			return string(contentJSON)
		}
	}
	return "professional content appropriate for the business"
}

// NOTE: The rest of the functions (ProcessHTMLAction, ValidateHTMLAction, etc.)
// remain the same as in the original html_actions.go file

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

	// Extract domain analysis (checking for response)
	if domainStep, ok := collectedData["analyze_domain"]; ok {
		context["domain_analysis"] = ExtractStepData(domainStep)
	}

	// Extract site architecture (checking for response)
	if archStep, ok := collectedData["architect_site"]; ok {
		context["site_structure"] = ExtractStepData(archStep)
	}

	// Extract content (checking for response)
	if contentStep, ok := collectedData["create_content"]; ok {
		context["content"] = ExtractStepData(contentStep)
	}

	// Extract business info
	context["business"] = extractBusinessInfo(collectedData)

	return context
}

func extractBusinessInfo(collectedData map[string]interface{}) map[string]interface{} {
	info := make(map[string]interface{})

	// Try to get from input data (checking for response field)
	if inputStep, ok := collectedData["input_data"]; ok {
		extracted := ExtractStepData(inputStep)
		if inputData, ok := extracted.(map[string]interface{}); ok {
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
	}

	// Try to get from headers (checking for response field)
	if headersStep, ok := collectedData["headers"]; ok {
		extracted := ExtractStepData(headersStep)
		if headers, ok := extracted.(map[string]interface{}); ok {
			if clientID, ok := headers["client_id"].(string); ok {
				info["client_id"] = clientID
			}
		}
	}

	return info
}

func buildHTMLPrompt(context map[string]interface{}, agentConfig map[string]interface{}) string {
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
