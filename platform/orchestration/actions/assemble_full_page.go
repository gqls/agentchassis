// FILE: platform/orchestration/actions/assemble_full_page.go
package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AssembleFullPageAction assembles a complete HTML page from:
// - Content JSON (from content-creator)
// - HTML templates (from architect)
// - CSS theme (from database)
// - CSS snippets (from database)
// - JS snippets (from database)
func AssembleFullPageAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("AssembleFullPageAction starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	config := params.StepConfig.Config

	// 1. Extract content data (from content-creator)
	contentData, err := extractContentData(params)
	if err != nil {
		return nil, fmt.Errorf("failed to extract content data: %w", err)
	}

	// 2. Extract template data (from architect)
	templateData, err := extractTemplateData(params)
	if err != nil {
		return nil, fmt.Errorf("failed to extract template data: %w", err)
	}

	// 3. Get database connection
	db, err := getDBConnection(params)
	if err != nil {
		params.Logger.Warn("No database connection, using embedded CSS", zap.Error(err))
		// Continue without DB - will use fallback CSS
	}

	// 4. Determine theme
	themeName := determineTheme(contentData, params)
	params.Logger.Info("Selected theme", zap.String("theme", themeName))

	// 5. Get CSS theme from database
	var themeCSS string
	if db != nil {
		themeCSS, err = getCSSTheme(ctx, db, themeName)
		if err != nil {
			params.Logger.Warn("Failed to get theme, using default", zap.Error(err))
			themeCSS = getDefaultThemeCSS()
		}
	} else {
		themeCSS = getDefaultThemeCSS()
	}

	// 6. Get CSS snippets based on tags
	var snippetCSS string
	if db != nil && getConfigBoolValue(config, "include_css_snippets", true) {
		tags := getThemeTags(contentData)
		snippetCSS, _ = getCSSSnippets(ctx, db, tags)
	}

	// 7. Get JS snippets
	var snippetJS string
	if db != nil && getConfigBoolValue(config, "include_js_snippets", true) {
		snippetJS, _ = getJSSnippets(ctx, db, []string{"nav-mobile-toggle", "nav-smooth-scroll", "interaction-accordion"})
	}

	// 8. Get base CSS from head component or use default
	baseCSS := getBaseCSS()

	// 9. Render HTML template with content
	renderedBody, err := renderTemplateWithContent(templateData, contentData, params.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
	}

	// 10. Get meta information
	meta := getMetaInfo(contentData, params)

	// 11. Assemble final HTML document
	finalHTML := assembleDocument(meta, baseCSS, themeCSS, snippetCSS, renderedBody, snippetJS)

	// 12. Optional minification
	if getConfigBoolValue(config, "minify", false) {
		finalHTML = minifyHTMLContent(finalHTML)
	}

	params.Logger.Info("AssembleFullPageAction complete",
		zap.Int("html_length", len(finalHTML)),
		zap.String("theme", themeName),
	)

	return map[string]interface{}{
		"final_html":   finalHTML,
		"theme_used":   themeName,
		"html_length":  len(finalHTML),
		"css_length":   len(baseCSS) + len(themeCSS) + len(snippetCSS),
		"js_length":    len(snippetJS),
		"assembled_at": time.Now().UTC(),
		"success":      true,
	}, nil
}

// ============================================================================
// DATA EXTRACTION HELPERS
// ============================================================================

func extractContentData(params ActionParams) (map[string]interface{}, error) {
	// Try to find content data in CollectedData
	possiblePaths := []string{
		"content_data",
		"input_data.content_data",
	}

	for _, path := range possiblePaths {
		if data, ok := getValueByDotPath(params.CollectedData, path); ok {
			if contentMap, ok := data.(map[string]interface{}); ok {
				// Check if it has a nested result (from LLM output)
				if genContent, ok := contentMap["generate_content"].(map[string]interface{}); ok {
					if result, ok := genContent["result"].(string); ok {
						// Parse JSON string
						var parsed map[string]interface{}
						// Clean markdown if present
						cleanResult := cleanJSONFromMarkdown(result)
						if err := json.Unmarshal([]byte(cleanResult), &parsed); err == nil {
							return parsed, nil
						}
					}
				}
				return contentMap, nil
			}
		}
	}

	return nil, fmt.Errorf("content data not found in CollectedData")
}

func extractTemplateData(params ActionParams) (map[string]interface{}, error) {
	possiblePaths := []string{
		"template_data",
		"input_data.template_data",
	}

	for _, path := range possiblePaths {
		if data, ok := getValueByDotPath(params.CollectedData, path); ok {
			if templateMap, ok := data.(map[string]interface{}); ok {
				// Check for assemble_template result
				if assembled, ok := templateMap["assemble_template"].(map[string]interface{}); ok {
					return assembled, nil
				}
				return templateMap, nil
			}
		}
	}

	return nil, fmt.Errorf("template data not found in CollectedData")
}

func getValueByDotPath(data map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	current := interface{}(data)

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
			} else {
				return nil, false
			}
		} else {
			return nil, false
		}
	}

	return current, true
}

func cleanJSONFromMarkdown(s string) string {
	// Remove ```json and ``` markers
	re := regexp.MustCompile("(?s)```json\\s*(.*)```")
	if matches := re.FindStringSubmatch(s); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	re = regexp.MustCompile("(?s)```\\s*(.*)```")
	if matches := re.FindStringSubmatch(s); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return strings.TrimSpace(s)
}

// ============================================================================
// THEME DETERMINATION
// ============================================================================

func determineTheme(contentData map[string]interface{}, params ActionParams) string {
	// 1. Check if content specifies a theme
	if theme, ok := contentData["theme"].(string); ok && theme != "" {
		return theme
	}

	// 2. Check brief data for theme recommendation
	if briefData, ok := params.CollectedData["brief_data"].(map[string]interface{}); ok {
		if themeInfo, ok := briefData["theme"].(map[string]interface{}); ok {
			if recommended, ok := themeInfo["recommended"].(string); ok && recommended != "" {
				return recommended
			}
		}
	}

	// 3. Default
	return "default"
}

func getThemeTags(contentData map[string]interface{}) []string {
	if tags, ok := contentData["theme_tags"].([]interface{}); ok {
		result := make([]string, 0, len(tags))
		for _, t := range tags {
			if s, ok := t.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return []string{"professional", "modern"}
}

// ============================================================================
// DATABASE QUERIES
// ============================================================================

func getDBConnection(params ActionParams) (*sql.DB, error) {
	// Try to get from params or environment
	// This should be injected by the agent framework
	if db, ok := params.CollectedData["__db_connection__"].(*sql.DB); ok {
		return db, nil
	}

	// The agent framework should provide this - return nil if not available
	return nil, fmt.Errorf("database connection not available")
}

func getCSSTheme(ctx context.Context, db *sql.DB, themeName string) (string, error) {
	var cssContent string
	err := db.QueryRowContext(ctx,
		"SELECT css_content FROM css_themes WHERE name = $1 AND is_active = true",
		themeName,
	).Scan(&cssContent)

	if err != nil {
		return "", err
	}
	return cssContent, nil
}

func getCSSSnippets(ctx context.Context, db *sql.DB, tags []string) (string, error) {
	if len(tags) == 0 {
		return "", nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT css_content FROM css_snippets 
		 WHERE semantic_tags && $1
		 ORDER BY category, name`,
		tags,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var snippets []string
	for rows.Next() {
		var css string
		if err := rows.Scan(&css); err == nil {
			snippets = append(snippets, css)
		}
	}

	return strings.Join(snippets, "\n\n"), nil
}

func getJSSnippets(ctx context.Context, db *sql.DB, functions []string) (string, error) {
	if len(functions) == 0 {
		return "", nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT js_content FROM js_snippets 
		 WHERE function = ANY($1)
		 ORDER BY category, name`,
		functions,
	)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var snippets []string
	for rows.Next() {
		var js string
		if err := rows.Scan(&js); err == nil {
			snippets = append(snippets, js)
		}
	}

	return strings.Join(snippets, "\n\n"), nil
}

// ============================================================================
// TEMPLATE RENDERING
// ============================================================================

func renderTemplateWithContent(templateData, contentData map[string]interface{}, logger *zap.Logger) (string, error) {
	// Get the stitched HTML template
	htmlTemplate, ok := templateData["stitched_html_template"].(string)
	if !ok {
		return "", fmt.Errorf("stitched_html_template not found")
	}

	// Get sections from content data
	sections, ok := contentData["sections"].(map[string]interface{})
	if !ok {
		logger.Warn("No sections in content data, using content_requirements from template")
		// Fall back to content_requirements defaults
		if reqs, ok := templateData["content_requirements"].(map[string]interface{}); ok {
			sections = reqs
		} else {
			sections = make(map[string]interface{})
		}
	}

	// Flatten sections into a single data map for template rendering
	flatData := make(map[string]interface{})
	for componentID, componentData := range sections {
		if data, ok := componentData.(map[string]interface{}); ok {
			for key, value := range data {
				flatData[key] = value
			}
		}
		// Also store by component ID in case template uses it
		flatData[componentID] = componentData
	}

	// Convert Go template syntax {{ .var }} - ensure it's proper Go template
	// The templates use {{.var_name}} syntax
	tmpl, err := template.New("page").Parse(htmlTemplate)
	if err != nil {
		// Try with looser parsing - replace {{ with {{
		htmlTemplate = strings.ReplaceAll(htmlTemplate, "{{ ", "{{")
		htmlTemplate = strings.ReplaceAll(htmlTemplate, " }}", "}}")
		tmpl, err = template.New("page").Parse(htmlTemplate)
		if err != nil {
			return "", fmt.Errorf("failed to parse template: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, flatData); err != nil {
		logger.Warn("Template execution had errors, using partial result", zap.Error(err))
		// Return what we have even if incomplete
		return htmlTemplate, nil
	}

	return buf.String(), nil
}

// ============================================================================
// HTML ASSEMBLY
// ============================================================================

func getMetaInfo(contentData map[string]interface{}, params ActionParams) map[string]string {
	meta := map[string]string{
		"title":       "Website",
		"description": "",
		"lang":        "en",
	}

	// From content data
	if metaData, ok := contentData["meta"].(map[string]interface{}); ok {
		if title, ok := metaData["title"].(string); ok {
			meta["title"] = title
		}
		if desc, ok := metaData["description"].(string); ok {
			meta["description"] = desc
		}
	}

	// From input data
	if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if domain, ok := inputData["domain"].(string); ok && meta["title"] == "Website" {
			meta["title"] = domain
		}
	}

	return meta
}

func assembleDocument(meta map[string]string, baseCSS, themeCSS, snippetCSS, body, js string) string {
	var sb strings.Builder

	sb.WriteString("<!DOCTYPE html>\n")
	sb.WriteString(fmt.Sprintf("<html lang=\"%s\">\n", meta["lang"]))
	sb.WriteString("<head>\n")
	sb.WriteString("  <meta charset=\"UTF-8\">\n")
	sb.WriteString("  <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", template.HTMLEscapeString(meta["title"])))

	if meta["description"] != "" {
		sb.WriteString(fmt.Sprintf("  <meta name=\"description\" content=\"%s\">\n", template.HTMLEscapeString(meta["description"])))
	}

	// CSS
	sb.WriteString("  <style>\n")
	sb.WriteString("    /* Base Styles */\n")
	sb.WriteString(baseCSS)
	sb.WriteString("\n\n    /* Theme Variables */\n")
	sb.WriteString(themeCSS)
	if snippetCSS != "" {
		sb.WriteString("\n\n    /* CSS Snippets */\n")
		sb.WriteString(snippetCSS)
	}
	sb.WriteString("\n  </style>\n")

	sb.WriteString("</head>\n")
	sb.WriteString("<body>\n")
	sb.WriteString(body)
	sb.WriteString("\n")

	// JavaScript
	if js != "" {
		sb.WriteString("<script>\n")
		sb.WriteString(js)
		sb.WriteString("\n</script>\n")
	}

	sb.WriteString("</body>\n")
	sb.WriteString("</html>")

	return sb.String()
}

// ============================================================================
// FALLBACK CSS
// ============================================================================

func getBaseCSS() string {
	return `/* CSS Reset & Base Styles */
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
  line-height: 1.6;
  color: var(--color-text, #333);
  background-color: var(--color-background, #fff);
}

.container { max-width: 1200px; margin: 0 auto; padding: 0 1rem; }
.container--narrow { max-width: 800px; }

.grid { display: grid; gap: 2rem; }
.grid--2 { grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); }
.grid--3 { grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); }
.grid--4 { grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); }

.section { padding: 4rem 1rem; }
.section__title { font-size: 2.5rem; margin-bottom: 1rem; color: var(--color-heading, #111); }
.section__title--center { text-align: center; }

.card { 
  padding: 2rem; 
  background: var(--color-card-bg, #fff);
  border-radius: var(--border-radius, 0.5rem);
  box-shadow: var(--shadow, 0 1px 3px rgba(0,0,0,0.1));
}

.button {
  display: inline-block;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: var(--border-radius, 0.5rem);
  font-size: 1rem;
  font-weight: 600;
  text-decoration: none;
  cursor: pointer;
  transition: all 0.2s;
}
.button--primary { background: var(--color-primary, #3b82f6); color: var(--color-primary-text, #fff); }
.button--primary:hover { background: var(--color-primary-hover, #2563eb); }
.button--secondary { background: var(--color-secondary, #64748b); color: var(--color-secondary-text, #fff); }
.button--large { padding: 1rem 2rem; font-size: 1.125rem; }
.button--small { padding: 0.5rem 1rem; font-size: 0.875rem; }
.button--full-width { display: block; width: 100%; text-align: center; }
.button--primary-inverse { background: var(--color-background, #fff); color: var(--color-primary, #3b82f6); }
.button--secondary-inverse { background: transparent; color: var(--color-background, #fff); border: 2px solid var(--color-background, #fff); }

.site-header { background: var(--color-header-bg, #1e293b); color: var(--color-header-text, #fff); position: sticky; top: 0; z-index: 1000; }
.site-header__nav { display: flex; justify-content: space-between; align-items: center; padding: 1rem; }
.site-header__brand { font-size: 1.5rem; font-weight: bold; }
.site-header__menu { display: flex; gap: 2rem; list-style: none; }
.site-header__link { color: var(--color-header-text, #fff); text-decoration: none; }
.site-header__link:hover { opacity: 0.8; }

.hero { text-align: center; padding: 4rem 1rem; }
.hero__title { font-size: 3rem; margin-bottom: 1.5rem; color: var(--color-hero-title, #111); }
.hero__subtitle { font-size: 1.5rem; margin-bottom: 2rem; color: var(--color-hero-subtitle, #666); }
.hero__actions { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap; }

.section--hero { background: var(--color-header-bg, #1e293b); color: var(--color-header-text, #fff); }
.section--hero .hero__title { color: var(--color-header-text, #fff); }
.section--hero .hero__subtitle { color: var(--color-header-text, #fff); opacity: 0.9; }

.section--cta { background: var(--color-cta-bg, #3b82f6); color: var(--color-cta-text, #fff); }
.cta { text-align: center; }
.cta__title { font-size: 2.5rem; margin-bottom: 1rem; color: inherit; }
.cta__subtitle { font-size: 1.25rem; margin-bottom: 2rem; }
.cta__actions { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap; }

.feature { text-align: center; }
.feature__icon { font-size: 3rem; margin-bottom: 1rem; }
.feature__title { font-size: 1.25rem; margin-bottom: 0.5rem; }
.feature__description { color: var(--color-text-muted, #666); }

.pricing-tier { position: relative; }
.pricing-tier--featured { border: 3px solid var(--color-primary, #3b82f6); transform: scale(1.05); }
.pricing-tier__badge { position: absolute; top: -1rem; right: 1rem; background: var(--color-primary, #3b82f6); color: #fff; padding: 0.25rem 1rem; border-radius: 0.25rem; font-size: 0.875rem; }
.pricing-tier__name { font-size: 1.5rem; margin-bottom: 1rem; }
.pricing-tier__price { font-size: 2.5rem; font-weight: bold; margin-bottom: 1.5rem; color: var(--color-primary, #3b82f6); }
.pricing-tier__features { list-style: none; margin-bottom: 2rem; }
.pricing-tier__feature { padding: 0.5rem 0; border-bottom: 1px solid var(--color-border, #e5e7eb); }

.stat-highlight { text-align: center; margin-bottom: 3rem; }
.stat-highlight__number { font-size: 3rem; font-weight: bold; color: var(--color-primary, #3b82f6); }
.stat-highlight__label { font-size: 1.25rem; color: var(--color-text-muted, #666); }

.testimonial { text-align: center; }
.testimonial__rating { color: var(--color-accent, #fbbf24); font-size: 1.25rem; margin-bottom: 1rem; }
.testimonial__text { font-style: italic; margin-bottom: 1rem; }
.testimonial__author { font-weight: bold; color: var(--color-text-muted, #666); }

.faq-item { padding: 1.5rem; background: var(--color-card-bg, #fff); border-radius: var(--border-radius, 0.5rem); margin-bottom: 1rem; }
.faq-item__question { font-size: 1.125rem; font-weight: 600; cursor: pointer; }
.faq-item__answer { margin-top: 1rem; color: var(--color-text-muted, #666); }

.site-footer { background: var(--color-footer-bg, #1e293b); color: var(--color-footer-text, #fff); padding: 3rem 1rem 1rem; }
.site-footer__brand { font-size: 1.5rem; font-weight: bold; margin-bottom: 0.5rem; }
.site-footer__tagline { opacity: 0.8; }
.site-footer__heading { font-size: 1.125rem; margin-bottom: 1rem; }
.site-footer__links { list-style: none; }
.site-footer__links li { margin-bottom: 0.5rem; }
.site-footer__link { color: inherit; text-decoration: none; opacity: 0.8; }
.site-footer__link:hover { opacity: 1; }
.site-footer__bottom { margin-top: 3rem; padding-top: 2rem; border-top: 1px solid rgba(255,255,255,0.1); text-align: center; opacity: 0.6; }

@media (max-width: 768px) {
  .site-header__menu { display: none; }
  .site-header__menu.is-open { display: flex; flex-direction: column; position: absolute; top: 100%; left: 0; right: 0; background: var(--color-header-bg, #1e293b); padding: 1rem; }
  .hero__title { font-size: 2rem; }
  .hero__subtitle { font-size: 1.25rem; }
  .section__title { font-size: 2rem; }
  .pricing-tier--featured { transform: none; }
}`
}

func getDefaultThemeCSS() string {
	return `:root {
  --color-primary: #3b82f6;
  --color-primary-hover: #2563eb;
  --color-primary-text: #ffffff;
  --color-secondary: #64748b;
  --color-secondary-hover: #475569;
  --color-secondary-text: #ffffff;
  --color-accent: #fbbf24;

  --color-text: #1e293b;
  --color-text-muted: #64748b;
  --color-heading: #0f172a;
  --color-background: #ffffff;
  --color-border: #e2e8f0;

  --color-header-bg: #1e293b;
  --color-header-text: #ffffff;
  --color-hero-title: #0f172a;
  --color-hero-subtitle: #475569;
  --color-card-bg: #ffffff;
  --color-cta-bg: linear-gradient(135deg, #3b82f6 0%, #2563eb 100%);
  --color-cta-text: #ffffff;
  --color-footer-bg: #1e293b;
  --color-footer-text: #ffffff;

  --border-radius: 0.5rem;
  --shadow: 0 1px 3px rgba(0,0,0,0.1);
}`
}

// ============================================================================
// UTILITY HELPERS
// ============================================================================

func getConfigBoolValue(config map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := config[key].(bool); ok {
		return val
	}
	return defaultVal
}

func minifyHTMLContent(html string) string {
	// Basic minification - remove excess whitespace
	// For production, use a proper minifier library
	re := regexp.MustCompile(`>\s+<`)
	html = re.ReplaceAllString(html, "><")
	re = regexp.MustCompile(`\s+`)
	html = re.ReplaceAllString(html, " ")
	return strings.TrimSpace(html)
}
