package actions

// FILE: platform/orchestration/actions/render_css_from_spec_action.go
//
// Replaces the LLM-based generate_css step in webdesign-agent with deterministic
// Go template rendering. The analyze_design LLM step picks industry-appropriate
// colors/fonts; this action just slots those values into a CSS template.
//
// Workflow wiring (webdesign-agent):
//   analyze_design (LLM) → render_css_from_spec (this) → deploy_css (git_commit)
//
// Output matches what deploy_css expects: {"result": "<css>", "type": "text"}
//
// Hot-swappable: css_themes rows hold different Go templates (e.g.
// "standard-brochure", "animated-portfolio"). Config selects which one.
//
// Step Zero search:
//   - execute_llm_prompt: renders template then calls LLM. No "skip LLM" mode.
//   - transform_data: key-value transforms, no template or DB queries.
//   - render_site_components: renders HTML components, not CSS from design specs.
//   - RenderTemplateWithMap: utility function, no DB or data reshaping.
//   Decision: New action. No existing action or small patch covers the need.
//
// Registration (add to registry.go):
//   "render_css_from_spec": { Handler: RenderCSSFromSpecAction, Category: "site",
//       Description: "Render CSS from Go template + design spec", IsLocal: true },
// Also add to local_actions.go:
//   "render_css_from_spec": true,

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// cssTemplateData is passed to the Go CSS template. Fields are exported
// because text/template requires it; the type itself stays unexported.
type cssTemplateData struct {
	Primary    string
	Secondary  string
	Accent     string
	Background string
	Surface    string
	Text       string
	TextMuted  string
	Border     string

	FontFamily  string
	HeadingFont string
	BaseSize    string
	LineHeight  string

	SectionPadding    string
	ContainerMaxWidth string

	Components    []string
	SectionStyles []sectionStyleEntry
}

type sectionStyleEntry struct {
	Function  string // e.g. "hero"
	ClassName string // e.g. "hero-section"
	IsDark    bool
}

// RenderCSSFromSpecAction renders CSS deterministically from a Go template
// stored in css_themes, merged with the design_spec from analyze_design.
//
// Config:
//   - theme_name: css_themes.name to load (default: "standard-brochure")
//   - theme_id:   explicit UUID override (optional)
//
// Reads from collectedData (populated by prior workflow steps):
//   - design_spec.result.color_scheme / typography / spacing
//   - site_context.all_component_functions
func RenderCSSFromSpecAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("RenderCSSFromSpecAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	// 1. Extract design_spec from collectedData
	spec := extractDesignColors(params.CollectedData, logger)

	// 2. Extract component list from site_context
	components := extractCSSComponents(params.CollectedData, logger)

	// 3. Query which components are dark sections
	darkSections := queryDarkSectionsForCSS(ctx, params.DB, components, logger)

	// 4. Build sorted section style entries
	sectionStyles := buildCSSsectionStyles(components, darkSections)

	// 5. Assemble template data
	tmplData := cssTemplateData{
		Primary:           spec.color("primary", "#1a365d"),
		Secondary:         spec.color("secondary", "#2c5282"),
		Accent:            spec.color("accent", "#3182ce"),
		Background:        spec.color("background", "#ffffff"),
		Surface:           spec.color("surface", "#f7fafc"),
		Text:              spec.color("text", "#2d3748"),
		TextMuted:         spec.color("text_muted", "#718096"),
		Border:            spec.color("border", "#e2e8f0"),
		FontFamily:        spec.typo("font_family", "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif"),
		HeadingFont:       spec.typo("heading_font", "inherit"),
		BaseSize:          spec.typo("base_size", "16px"),
		LineHeight:        spec.typo("line_height", "1.6"),
		SectionPadding:    spec.space("section_padding", "4rem 0"),
		ContainerMaxWidth: spec.space("container_max_width", "1200px"),
		Components:        components,
		SectionStyles:     sectionStyles,
	}

	logger.Info("RenderCSSFromSpecAction: Template data built",
		zap.String("primary", tmplData.Primary),
		zap.String("accent", tmplData.Accent),
		zap.String("font_family", tmplData.FontFamily),
		zap.Int("components", len(components)),
		zap.Int("dark_sections", len(darkSections)),
	)

	// 6. Load Go template from css_themes table
	themeName := datahelpers.GetStringField(config, "theme_name", "standard-brochure")

	cssTemplate, err := loadCSSGoTemplate(ctx, params.DB, config, themeName, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load CSS template '%s': %w", themeName, err)
	}

	// 7. Parse and render
	tmpl, err := template.New("css").Parse(cssTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSS template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return nil, fmt.Errorf("failed to render CSS template: %w", err)
	}

	renderedCSS := buf.String()

	// 7.5 Append component-specific CSS snippets
	snippetCSS := loadComponentCSSSnippets(ctx, params.DB, components, logger)
	if snippetCSS != "" {
		renderedCSS = renderedCSS + snippetCSS
	}

	logger.Info("RenderCSSFromSpecAction: CSS rendered",
		zap.Int("css_length", len(renderedCSS)),
		zap.Int("snippet_css_length", len(snippetCSS)),
		zap.String("theme", themeName),
	)

	// Return format matches execute_llm_prompt output so deploy_css works unchanged
	return map[string]interface{}{
		"result": renderedCSS,
		"type":   "text",
	}, nil
}

// --- private helpers ---

// designColorMaps wraps the nested design_spec maps for safe field access.
type designColorMaps struct {
	colorScheme map[string]interface{}
	typography  map[string]interface{}
	spacing     map[string]interface{}
}

func (d *designColorMaps) color(key, fallback string) string {
	return getMapString(d.colorScheme, key, fallback)
}

func (d *designColorMaps) typo(key, fallback string) string {
	return getMapString(d.typography, key, fallback)
}

func (d *designColorMaps) space(key, fallback string) string {
	return getMapString(d.spacing, key, fallback)
}

func getMapString(m map[string]interface{}, key, fallback string) string {
	if m == nil {
		return fallback
	}
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return fallback
}

// extractDesignColors pulls color_scheme, typography, spacing from design_spec
// in collectedData. The LLM step stores: design_spec = {"result": {...}, "type": "json"}
func extractDesignColors(collectedData map[string]interface{}, logger *zap.Logger) *designColorMaps {
	helper := &designColorMaps{}

	// Try design_spec.result.color_scheme (standard execute_llm_prompt output)
	colorScheme := datahelpers.ExtractNestedField(collectedData, "design_spec.result.color_scheme")
	if colorScheme == nil {
		// Fallback: design_spec.color_scheme (if .response was auto-unwrapped)
		colorScheme = datahelpers.ExtractNestedField(collectedData, "design_spec.color_scheme")
	}
	if cs, ok := colorScheme.(map[string]interface{}); ok {
		helper.colorScheme = cs
	} else {
		logger.Warn("RenderCSSFromSpecAction: color_scheme not found or not a map, using defaults")
	}

	typography := datahelpers.ExtractNestedField(collectedData, "design_spec.result.typography")
	if typography == nil {
		typography = datahelpers.ExtractNestedField(collectedData, "design_spec.typography")
	}
	if tp, ok := typography.(map[string]interface{}); ok {
		helper.typography = tp
	}

	spacing := datahelpers.ExtractNestedField(collectedData, "design_spec.result.spacing")
	if spacing == nil {
		spacing = datahelpers.ExtractNestedField(collectedData, "design_spec.spacing")
	}
	if sp, ok := spacing.(map[string]interface{}); ok {
		helper.spacing = sp
	}

	return helper
}

// extractCSSComponents gets the all_component_functions slice from site_context.
func extractCSSComponents(collectedData map[string]interface{}, logger *zap.Logger) []string {
	raw := datahelpers.ExtractNestedField(collectedData, "site_context.all_component_functions")
	if raw == nil {
		logger.Warn("RenderCSSFromSpecAction: no component list, using defaults")
		return []string{"hero", "services-grid", "differentiators", "social-proof", "call-to-action"}
	}

	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	default:
		logger.Warn("RenderCSSFromSpecAction: component list unexpected type",
			zap.String("type", fmt.Sprintf("%T", raw)))
		return []string{"hero", "services-grid", "differentiators", "social-proof", "call-to-action"}
	}
}

// queryDarkSectionsForCSS queries content_components.is_dark_section for the
// given function names. Returns a map of function→true for dark sections.
func queryDarkSectionsForCSS(ctx context.Context, db *sql.DB, functions []string, logger *zap.Logger) map[string]bool {
	darkSections := make(map[string]bool)
	if len(functions) == 0 {
		return darkSections
	}

	placeholders := make([]string, len(functions))
	args := make([]interface{}, len(functions))
	for i, f := range functions {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = f
	}

	query := fmt.Sprintf(`
		SELECT function, is_dark_section
		FROM content_components
		WHERE function IN (%s)
		  AND is_active = true
		  AND forked_from IS NULL
	`, strings.Join(placeholders, ","))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Warn("RenderCSSFromSpecAction: dark section query failed, using common defaults",
			zap.Error(err))
		for _, f := range []string{"hero", "social-proof", "call-to-action", "testimonials"} {
			darkSections[f] = true
		}
		return darkSections
	}
	defer rows.Close()

	for rows.Next() {
		var function string
		var isDark bool
		if err := rows.Scan(&function, &isDark); err == nil && isDark {
			darkSections[function] = true
		}
	}

	logger.Info("RenderCSSFromSpecAction: Dark sections loaded",
		zap.Any("dark_sections", darkSections))

	return darkSections
}

// buildCSSsectionStyles creates a sorted list of section style entries.
func buildCSSsectionStyles(components []string, darkSections map[string]bool) []sectionStyleEntry {
	styles := make([]sectionStyleEntry, 0, len(components))
	for _, f := range components {
		styles = append(styles, sectionStyleEntry{
			Function:  f,
			ClassName: f + "-section",
			IsDark:    darkSections[f],
		})
	}
	sort.Slice(styles, func(i, j int) bool {
		return styles[i].Function < styles[j].Function
	})
	return styles
}

// loadCSSGoTemplate loads the Go template string from css_themes.
// Priority: config["theme_id"] (UUID) → themeName → "standard-brochure"
func loadCSSGoTemplate(ctx context.Context, db *sql.DB, config map[string]interface{}, themeName string, logger *zap.Logger) (string, error) {
	// Try explicit theme_id first
	if themeIDStr, ok := config["theme_id"].(string); ok && themeIDStr != "" {
		var tmpl sql.NullString
		err := db.QueryRowContext(ctx,
			`SELECT css_template FROM css_themes WHERE id = $1 AND is_active = true`,
			themeIDStr,
		).Scan(&tmpl)
		if err == nil && tmpl.Valid && tmpl.String != "" {
			logger.Info("RenderCSSFromSpecAction: Loaded template by ID",
				zap.String("theme_id", themeIDStr))
			return tmpl.String, nil
		}
	}

	// Try by name
	var tmpl sql.NullString
	err := db.QueryRowContext(ctx,
		`SELECT css_template FROM css_themes WHERE name = $1 AND is_active = true`,
		themeName,
	).Scan(&tmpl)
	if err == nil && tmpl.Valid && tmpl.String != "" {
		logger.Info("RenderCSSFromSpecAction: Loaded template by name",
			zap.String("theme_name", themeName))
		return tmpl.String, nil
	}

	// Fallback if a different name was requested
	if themeName != "standard-brochure" {
		err := db.QueryRowContext(ctx,
			`SELECT css_template FROM css_themes WHERE name = 'standard-brochure' AND is_active = true`,
		).Scan(&tmpl)
		if err == nil && tmpl.Valid && tmpl.String != "" {
			logger.Warn("RenderCSSFromSpecAction: Requested theme not found, falling back",
				zap.String("requested", themeName))
			return tmpl.String, nil
		}
	}

	return "", fmt.Errorf("no CSS template found for theme '%s'", themeName)
}

// loadComponentCSSSnippets queries css_snippets for entries whose applies_to
// array overlaps with the site's component list. Returns concatenated CSS.
func loadComponentCSSSnippets(ctx context.Context, db *sql.DB, components []string, logger *zap.Logger) string {
	if len(components) == 0 || db == nil {
		return ""
	}

	// Build a JSONB array from the component list for the overlap operator (&&)
	componentsJSON, err := json.Marshal(components)
	if err != nil {
		logger.Warn("loadComponentCSSSnippets: marshal failed", zap.Error(err))
		return ""
	}

	rows, err := db.QueryContext(ctx, `
		SELECT name, css_content 
		FROM css_snippets 
		WHERE applies_to && $1::jsonb
		ORDER BY name
	`, string(componentsJSON))
	if err != nil {
		logger.Warn("loadComponentCSSSnippets: query failed", zap.Error(err))
		return ""
	}
	defer rows.Close()

	var parts []string
	for rows.Next() {
		var name, cssContent string
		if err := rows.Scan(&name, &cssContent); err != nil {
			logger.Warn("loadComponentCSSSnippets: scan error", zap.Error(err))
			continue
		}
		parts = append(parts, cssContent)
		logger.Info("loadComponentCSSSnippets: included snippet",
			zap.String("name", name),
			zap.Int("css_bytes", len(cssContent)))
	}

	if len(parts) == 0 {
		return ""
	}

	return "\n\n/* === Component-specific styles === */\n" + strings.Join(parts, "\n\n")
}
