package actions

// FILE: platform/orchestration/actions/render_css_from_spec_action.go
//
// Renders CSS deterministically from a palette + layout + typography
// composition, merged with the design_spec produced by analyze_design.
//
// This is the Phase 4.3 cutover of 025_palette_layout_typography_migration.
// Before cutover: the action loaded css_themes.css_template (a monolithic
// Go template) and fed it a struct with hardcoded fields. After cutover:
// the action loads the three independently-versioned rows via the
// css_themes FKs (palette_id, layout_id, typography_set_id) and feeds
// the layout's css_template three FuncMap helpers ({{palette}},
// {{typo}}, {{token}}) backed by merged string maps.
//
// What the caller still sees, unchanged:
//   - Action name, config keys (theme_name, theme_id)
//   - Return shape: {"result": "<css>", "type": "text"}
//   - Component-snippet append (css_snippets table)
//   - Renderer-owned --section-* defaults append (buildSectionDefaults)
//
// What's gone after cutover:
//   - cssTemplateData struct with hardcoded fields
//   - loadCSSGoTemplate (css_themes.css_template direct read)
//
// Merge rules applied here (implemented in
// render_css_composition_helpers.go):
//   - Palette core slots  (primary/secondary/accent/background/surface/
//                         text/text_muted/border): spec wins
//   - Palette specialised slots (heading/hero_title/cta_bg/etc.):
//                         theme wins
//   - Typography:         spec wins across the board
//   - Structure tokens:   layout-only; spec does not contribute

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

// sectionStyleEntry is the per-component entry passed as SectionStyles
// in the template data map. Exported fields only — text/template cannot
// access unexported struct fields, so this type stays as-is.
type sectionStyleEntry struct {
	Function  string // e.g. "hero"
	ClassName string // e.g. "hero-section"
	IsDark    bool
}

// RenderCSSFromSpecAction renders CSS deterministically from the theme's
// palette + layout + typography composition, merged with the design_spec
// from analyze_design.
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

	// 1. Resolve the theme composition (palette + layout + typography
	// in one query). Hard-errors on missing theme, NULL FKs, or
	// unparseable JSONB — migration gaps must be loud.
	// Surface site_id from collectedData into config so the loader's
	// site-composition resolution branch can find it. Does nothing if
	// config already carries an explicit theme_id or theme_name (those
	// paths short-circuit the site_id lookup in loadThemeComposition).
	//
	// Two candidate paths: site_context.site_id (standard for design
	// agents) and input_data.site_id (less common). Config-supplied
	// theme_id still wins over both.
	if _, hasID := config["site_id"].(string); !hasID {
		if sid := datahelpers.ExtractNestedFieldString(
			params.CollectedData, "site_context.site_id",
		); sid != "" {
			config["site_id"] = sid
		} else if sid := datahelpers.ExtractNestedFieldString(
			params.CollectedData, "input_data.site_id",
		); sid != "" {
			config["site_id"] = sid
		}
	}

	themeName := datahelpers.GetStringField(config, "theme_name", "standard-brochure")
	comp, err := loadThemeComposition(ctx, params.DB, config, themeName, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load theme composition: %w", err)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to load theme composition: %w", err)
	}

	// 2. Extract raw spec maps from collectedData. These are the
	// partial overrides the design_spec step produced.
	specPalette, specTypo, specSpacing := extractDesignSpecRawMaps(params.CollectedData, logger)

	// 3. Merge spec with theme, core-vs-specialised rules applied.
	// The merged palette is what the rest of the render path uses —
	// template lookups, dark-section detection, buildSectionDefaults.
	mergedPalette := buildPaletteMap(comp.Palette, specPalette)
	mergedTypo := buildTypographyMap(comp.Typography, specTypo)

	// 4. Surface the override deltas for observability. Operators
	// debugging a site's render can see at a glance which slots the
	// spec claimed from the theme.
	logOverrides(logger, comp, specPalette, specTypo, mergedPalette, mergedTypo)

	// 5. Extract component list + dark-section flags for the template
	// data (separate concern from palette/typography — these drive
	// per-section iteration, not colour selection).
	components := extractCSSComponents(params.CollectedData, logger)
	darkSections := queryDarkSectionsForCSS(ctx, params.DB, components, logger)
	sectionStyles := buildCSSsectionStyles(components, darkSections)

	// 6. Compute dark-palette flags for the renderer-owned
	// --section-* defaults. The merged palette drives these (not the
	// raw theme palette) because a site-supplied light/dark background
	// override must flip the section defaults accordingly.
	bgHex := lookupOrFallback(mergedPalette, "background", "#ffffff")
	surfaceHex := lookupOrFallback(mergedPalette, "surface", "#f7fafc")
	backgroundIsDark := isDarkHex(bgHex)
	surfaceIsDark := isDarkHex(surfaceHex)

	// 7. Build the template data map. Layout templates use the three
	// helper funcs registered below plus these top-level keys.
	tmplData := map[string]interface{}{
		"Components":       components,
		"SectionStyles":    sectionStyles,
		"SurfaceIsDark":    surfaceIsDark,
		"BackgroundIsDark": backgroundIsDark,
		// Spacing block for layouts that want a spec-driven container
		// width or section padding. Not used by the 15 seeded layouts
		// (they read from structure_tokens via {{token}}), but harmless
		// to include and useful for adopted-from-site templates that
		// might look for it. NOTE: spec-sourced only.
		"Spacing": specSpacing,
	}

	logger.Info("RenderCSSFromSpecAction: Template data built",
		zap.String("theme", comp.ThemeName),
		zap.String("palette", comp.PaletteName),
		zap.String("layout", comp.LayoutName),
		zap.String("typography", comp.TypographyName),
		zap.String("primary", lookupOrFallback(mergedPalette, "primary", "")),
		zap.String("background", bgHex),
		zap.Int("palette_keys", len(mergedPalette)),
		zap.Int("typography_keys", len(mergedTypo)),
		zap.Int("structure_keys", len(comp.Structure)),
		zap.Int("components", len(components)),
		zap.Int("dark_sections", len(darkSections)),
	)

	// 8. Parse and render the layout template with the three FuncMap
	// helpers registered. Helper signature is func(key, fallback) string
	// — the fallback lets layouts stay valid CSS even when a map omits
	// a slot (Template Helper Fallback contract).
	funcMap := template.FuncMap{
		"palette": makeMapLookupFunc(mergedPalette),
		"typo":    makeMapLookupFunc(mergedTypo),
		"token":   makeMapLookupFunc(comp.Structure),
	}

	tmpl, err := template.New("css").Funcs(funcMap).Parse(comp.LayoutTemplate)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to parse layout %q template: %w",
			comp.LayoutName, err,
		)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tmplData); err != nil {
		return nil, fmt.Errorf(
			"failed to render layout %q template: %w",
			comp.LayoutName, err,
		)
	}

	renderedCSS := buf.String()

	// 9. Append component-specific CSS snippets (unchanged from pre-
	// cutover behaviour — the snippets table and logic don't depend on
	// the theme composition model).
	snippetCSS := loadComponentCSSSnippets(ctx, params.DB, components, logger)
	if snippetCSS != "" {
		renderedCSS = renderedCSS + snippetCSS
	}

	// 10. Append renderer-enforced --section-* defaults based on the
	// MERGED palette's luminance. buildSectionDefaults picks readable
	// colours from the palette so every theme gets correct dark-section
	// text without declaring --section-* variables itself.
	sectionDefaults := buildSectionDefaults(
		bgHex,
		surfaceHex,
		mergedPalette,
		backgroundIsDark,
		surfaceIsDark,
		logger,
	)
	if sectionDefaults != "" {
		renderedCSS = renderedCSS + sectionDefaults
	}

	logger.Info("RenderCSSFromSpecAction: CSS rendered",
		zap.Int("css_length", len(renderedCSS)),
		zap.Int("snippet_css_length", len(snippetCSS)),
		zap.Int("section_defaults_length", len(sectionDefaults)),
		zap.String("theme", comp.ThemeName),
		zap.String("layout", comp.LayoutName),
	)

	// Return format matches execute_llm_prompt output so deploy_css
	// works unchanged.
	return map[string]interface{}{
		"result": renderedCSS,
		"type":   "text",
	}, nil
}

// --- private helpers ---

// lookupOrFallback is a non-template-context wrapper around the same
// lookup rule makeMapLookupFunc implements, for use from Go code.
func lookupOrFallback(m map[string]string, key, fallback string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return fallback
}

// extractDesignSpecRawMaps pulls the three design_spec sub-blocks out
// of collectedData and returns them as raw interface-valued maps (not
// yet merged or type-narrowed). The merge helpers do the narrowing.
//
// The LLM step stores design_spec = {"result": {...}, "type": "json"}
// but may be auto-unwrapped in some code paths — we try both shapes.
func extractDesignSpecRawMaps(
	collectedData map[string]interface{},
	logger *zap.Logger,
) (palette, typo, spacing map[string]interface{}) {

	extract := func(subField string) map[string]interface{} {
		// Standard execute_llm_prompt shape
		v := datahelpers.ExtractNestedField(collectedData, "design_spec.result."+subField)
		if v == nil {
			// Fallback: auto-unwrapped shape
			v = datahelpers.ExtractNestedField(collectedData, "design_spec."+subField)
		}
		if m, ok := v.(map[string]interface{}); ok {
			return m
		}
		return nil
	}

	palette = extract("color_scheme")
	typo = extract("typography")
	spacing = extract("spacing")

	if palette == nil {
		logger.Warn("RenderCSSFromSpecAction: design_spec.color_scheme not found or not a map")
	}

	return palette, typo, spacing
}

// logOverrides emits a single structured log line naming every palette
// and typography slot where the design_spec overrode the theme's value.
// Useful when debugging "why did my primary change?" and for
// release-train visibility into how much a site is diverging from the
// library theme.
//
// Behaviour: a key is considered "overridden" if the spec has a
// non-empty string value for it AND the merged result differs from the
// theme's original value for that key. Keys the spec couldn't win
// (e.g. specialised palette slots that the theme retained) are
// reported in the "claimed_but_ignored" list so the caller can see
// what the spec wanted but didn't get.
func logOverrides(
	logger *zap.Logger,
	comp *themeComposition,
	specPalette, specTypo map[string]interface{},
	mergedPalette, mergedTypo map[string]string,
) {
	paletteApplied, paletteIgnored := diffOverrides(comp.Palette, specPalette, mergedPalette)
	typoApplied, typoIgnored := diffOverrides(comp.Typography, specTypo, mergedTypo)

	if len(paletteApplied) == 0 && len(paletteIgnored) == 0 &&
		len(typoApplied) == 0 && len(typoIgnored) == 0 {
		return // nothing to say — spec contributed nothing
	}

	logger.Info("RenderCSSFromSpecAction: spec → theme overrides",
		zap.String("theme", comp.ThemeName),
		zap.Strings("palette_applied", paletteApplied),
		zap.Strings("palette_claimed_but_ignored", paletteIgnored),
		zap.Strings("typography_applied", typoApplied),
		zap.Strings("typography_claimed_but_ignored", typoIgnored),
	)
}

// diffOverrides computes which spec-supplied keys won and which were
// dropped. Returns sorted slices of keys (so log output is stable
// across runs).
//
// Applied:           spec had non-empty value AND merged differs from theme
// ClaimedButIgnored: spec had non-empty value AND merged matches theme
//
//	(i.e. spec wanted to override but wasn't allowed)
func diffOverrides(
	themeMap map[string]string,
	specMap map[string]interface{},
	mergedMap map[string]string,
) (applied, claimedButIgnored []string) {

	for k, raw := range specMap {
		specVal, ok := raw.(string)
		if !ok || specVal == "" {
			continue // spec didn't actually supply this key
		}
		themeVal := themeMap[k]
		mergedVal := mergedMap[k]

		if mergedVal == specVal && mergedVal != themeVal {
			applied = append(applied, k)
		} else if specVal != themeVal {
			// Spec wanted something different but theme won
			claimedButIgnored = append(claimedButIgnored, k)
		}
	}

	sort.Strings(applied)
	sort.Strings(claimedButIgnored)
	return applied, claimedButIgnored
}

// extractCSSComponents gets the all_component_functions slice from
// site_context. Unchanged from pre-cutover — the component list drives
// iteration and dark-section queries, not palette/typography.
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

// queryDarkSectionsForCSS queries content_components.is_dark_section
// for the given function names. Unchanged from pre-cutover.
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
// Unchanged from pre-cutover.
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

// loadComponentCSSSnippets queries css_snippets for entries whose
// applies_to array overlaps with the site's component list. Returns
// concatenated CSS. Unchanged from pre-cutover.
func loadComponentCSSSnippets(ctx context.Context, db *sql.DB, components []string, logger *zap.Logger) string {
	if len(components) == 0 || db == nil {
		return ""
	}

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
