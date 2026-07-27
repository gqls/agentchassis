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
//   - Scheme guard (bugs_open/022): a layout with a declared scheme
//     (light/dark) overrules a spec background that contradicts it —
//     the theme's background AND text are restored together, and the
//     render hard-fails when the theme cannot supply both (violating
//     CSS must never ship on a Warn alone)

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/google/uuid"
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
//   - site_context.site_id (or input_data.site_id) for site-composition
//     resolution inside loadThemeComposition
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

	// Resolve the site_id from collectedData to a plain local variable.
	// Two candidate paths: site_context.site_id (standard for design
	// agents) and input_data.site_id (less common). Explicit config
	// overrides (theme_id, theme_name) still win over this inside
	// loadThemeComposition.
	//
	// NB: we deliberately do NOT write site_id back into the step's
	// config map. Step.Config may be nil (workflow JSON without a
	// "config" key unmarshals to a nil map), and mutating shared
	// workflow state from an action is a bad pattern regardless.
	siteID := datahelpers.ExtractNestedFieldString(
		params.CollectedData, "site_context.site_id",
	)
	if siteID == "" {
		siteID = datahelpers.ExtractNestedFieldString(
			params.CollectedData, "input_data.site_id",
		)
	}

	// 1. Resolve the theme composition (palette + layout + typography
	// in one query). Hard-errors on missing theme, NULL FKs, or
	// unparseable JSONB — migration gaps must be loud.
	themeName := datahelpers.GetStringField(config, "theme_name", "standard-brochure")
	comp, err := loadThemeComposition(ctx, params.DB, config, siteID, themeName, logger)
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

	// 3b. Scheme guard: the layout's declared scheme is a user decision;
	// the spec's color_scheme is a per-run LLM guess. When they
	// contradict, the layout wins (bugs_open/022). A detected violation
	// the theme cannot repair hard-fails the render — scheme-violating
	// CSS must never ship, and a failed step is what the fleet-wide
	// failure sweep consumes (a Warn alone goes unwatched).
	if err := enforceLayoutScheme(comp.LayoutScheme, comp.Palette, mergedPalette, logger); err != nil {
		return nil, fmt.Errorf(
			"scheme guard (theme %q, layout %q): %w",
			comp.ThemeName, comp.LayoutName, err,
		)
	}

	// 3c. Fill the specialised slots a dark palette omits. Without this the
	// layout's own light literals (card_bg #ffffff, header_bg #ffffff,
	// cta_bg #1a365d) ship onto a dark site and carry the site's light text
	// colour on top of them. Inert on light sites and on any slot the
	// palette defines — see palette_specialised_slots.go for the measurement.
	mergedPalette = fillDarkSchemeSpecialisedSlots(mergedPalette, logger)

	// 3d. A primary that cannot be read on its own background makes every
	// eyebrow, link and card title invisible while every fill still looks
	// right. Warn, don't fail: the repair is a palette-authoring decision.
	warnUnusablePrimary(mergedPalette, logger)

	// 4. Surface the override deltas for observability. Operators
	// debugging a site's render can see at a glance which slots the
	// spec claimed from the theme.
	logOverrides(logger, comp, specPalette, specTypo, mergedPalette, mergedTypo)

	// 5. Extract component list + dark-section flags for the template
	// data (separate concern from palette/typography — these drive
	// per-section iteration, not colour selection).
	components := extractCSSComponents(params.CollectedData, logger)

	// 5b. An EMPTY list is not the same as a missing one, and only the
	// missing case has a fallback: extractCSSComponents returns defaults
	// when the field is nil, but an empty non-nil array passes straight
	// through, and loadComponentCSSSnippets early-returns "" on a
	// zero-length list. A stylesheet then ships with NO component CSS at
	// all and nothing says so — which is how ai-agent-orchestration.com
	// got a styles.css containing not one snippet (bugs_open/072); its
	// news cards have been bare ever since, because nothing re-renders a
	// site stylesheet once it is written.
	//
	// Resolve from the DB instead, exactly as the JS sibling already does
	// (loadSiteComponentFunctionsForJS, render_js_snippets_for_site_action.go).
	if len(components) == 0 {
		if siteUUID, err := uuid.Parse(siteID); err == nil {
			components = loadSiteComponentFunctionsForJS(ctx, params.DB, siteUUID, logger)
		}
		logger.Warn("RenderCSSFromSpecAction: component list was empty — resolved from the DB. "+
			"An empty list would have written a stylesheet with no component CSS (bugs_open/072).",
			zap.String("site_id", siteID),
			zap.Int("recovered_components", len(components)))
	}

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

	// 11. Append renderer-enforced compatibility aliases for custom-property
	// names that component templates consume but this theme does not define
	// (R6f vocabulary audit 2026-07-06: synonym drift such as --border-radius
	// vs --radius, legacy names such as --primary-color, and orphan tokens
	// such as --hero-ink). Only names absent from the CSS built so far are
	// added, so a layout that defines its own value always wins. Same
	// renderer-enforced pattern as buildSectionDefaults in step 10.
	aliasCSS := buildTokenAliases(renderedCSS, logger)
	if aliasCSS != "" {
		renderedCSS = renderedCSS + aliasCSS
	}

	logger.Info("RenderCSSFromSpecAction: CSS rendered",
		zap.Int("css_length", len(renderedCSS)),
		zap.Int("snippet_css_length", len(snippetCSS)),
		zap.Int("section_defaults_length", len(sectionDefaults)),
		zap.Int("token_alias_length", len(aliasCSS)),
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

// enforceLayoutScheme rejects a merged background that contradicts the
// layout's declared colour scheme (bugs_open/022: analyze_design emits a
// fresh color_scheme every run, and the core-slot spec-wins rule let a
// light background ship onto a scheme=dark site with no warning).
//
// The layout's scheme is a user decision; the spec's palette is a
// per-run LLM guess — on contradiction the layout wins. A violating
// background is replaced by the THEME's background AND text together:
// restoring only the background would pair it with a spec-chosen text
// colour and break contrast (never half-swap).
//
// Outcomes:
//   - scheme "", "neutral", or anything but light/dark → inert, no
//     signal (15 of 18 seeded layouts declare no scheme — logging here
//     would be per-render noise);
//   - merged background absent or not parseable hex (gradient, var())
//     → cannot be judged; passes through with an Info naming what was
//     skipped, so a scheme-declaring site never goes unexamined
//     silently (council round 2, bug_historian);
//   - violation, theme supplies background AND text → both restored
//     together, one Warn naming rejected and kept values;
//   - violation, theme missing either slot → ERROR. Restoring one slot
//     is the forbidden half-swap (council round 1), and shipping the
//     violating merge with only a Warn is the unwatched-signal shape
//     behind this class of incident (council round 2). A complete
//     theme palette is a Phase 3 data invariant, so this is the same
//     migration-gap-must-be-loud contract the composition loader
//     already enforces; the failed step is what the fleet-wide failure
//     sweep consumes, and the site keeps its last-good CSS.
//
// The luminance threshold is symmetric at 0.5: a dark layout rejects a
// background lighter than mid, a light layout rejects one darker. Real
// backgrounds cluster near the extremes (#F4F5F7 ≈ 0.91, #0f172a ≈ 0.01),
// so mid-tone false positives are a theoretical concern, not an observed
// one.
//
// PLACEMENT (council trail 0328ddc7, rounds 1+5): the guard lives HERE,
// at buildPaletteMap's single non-test call site, not inside the merge
// primitive. buildPaletteMap is a pure helper (its file's contract: no
// logger, no side effects) and scheme enforcement needs both logging and
// a failure path; today this call site IS the whole mechanism (caller
// count re-verified 2026-07-20). If a second caller of buildPaletteMap
// ever appears, it bypasses this guard — move the enforcement to that
// boundary or wrap the merge, and re-read bugs_open/022 first.
//
// mergedPalette is modified in place.
func enforceLayoutScheme(
	layoutScheme string,
	themePalette map[string]string,
	mergedPalette map[string]string,
	logger *zap.Logger,
) error {
	if layoutScheme != "light" && layoutScheme != "dark" {
		return nil
	}

	bgHex := mergedPalette["background"]
	if bgHex == "" {
		logger.Info("RenderCSSFromSpecAction: layout declares a scheme but merged palette has no background — scheme guard has nothing to judge",
			zap.String("layout_scheme", layoutScheme),
		)
		return nil
	}
	r, g, b, err := parseHexColor(bgHex)
	if err != nil {
		logger.Info("RenderCSSFromSpecAction: layout declares a scheme but merged background is not parseable hex — scheme guard cannot judge it",
			zap.String("layout_scheme", layoutScheme),
			zap.String("background", bgHex),
		)
		return nil
	}
	lum := relativeLuminance(r, g, b)

	violates := (layoutScheme == "dark" && lum > 0.5) ||
		(layoutScheme == "light" && lum < 0.5)
	if !violates {
		return nil
	}

	themeBG := themePalette["background"]
	themeText := themePalette["text"]
	if themeBG == "" || themeText == "" {
		return fmt.Errorf(
			"merged background %s (luminance %.3f) contradicts declared layout scheme %q, and the theme palette cannot supply both background and text to restore (has background: %t, has text: %t) — refusing to render scheme-violating CSS; repair the theme palette's core slots",
			bgHex, lum, layoutScheme, themeBG != "", themeText != "",
		)
	}

	rejectedText := mergedPalette["text"]
	mergedPalette["background"] = themeBG
	mergedPalette["text"] = themeText

	logger.Warn("RenderCSSFromSpecAction: spec background contradicts layout scheme — restoring theme background and text",
		zap.String("layout_scheme", layoutScheme),
		zap.String("rejected_background", bgHex),
		zap.Float64("rejected_luminance", lum),
		zap.String("kept_background", themeBG),
		zap.String("rejected_text", rejectedText),
		zap.String("kept_text", themeText),
	)
	return nil
}

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

	// Overlap check between two jsonb arrays. The `&&` operator does NOT
	// exist for jsonb (only for Postgres arrays like text[]). EXISTS +
	// jsonb_array_elements_text is the pure-jsonb pattern.
	rows, err := db.QueryContext(ctx, `
		SELECT name, css_content
		FROM css_snippets
		WHERE EXISTS (
		  SELECT 1
		  FROM jsonb_array_elements_text(applies_to) AS a(elem)
		  WHERE a.elem IN (SELECT jsonb_array_elements_text($1::jsonb))
		)
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

// tokenAliases maps component-consumed custom-property names that some
// themes never define onto canonical tokens (or safe literals). Order is
// stable so output is deterministic. Sourced from the R6f vocabulary
// audit (2026-07-06): synonym drift, legacy names, orphan tokens. Extend
// here when a new template vocabulary variant appears in the wild.
var tokenAliases = []struct{ Name, Value string }{
	{"--border-radius", "var(--radius, 8px)"},
	{"--shadow", "var(--shadow-md, 0 2px 12px rgba(0, 0, 0, 0.15))"},
	{"--spacing-section", "var(--section-pad-y, 4rem)"},
	{"--container-max-width", "var(--container-max, 1200px)"},
	{"--primary-color", "var(--color-primary)"},
	{"--secondary-color", "var(--color-secondary)"},
	{"--accent-color", "var(--color-accent)"},
	{"--color-heading", "var(--color-text)"},
	{"--color-white", "#ffffff"},
	{"--color-error", "#d64545"},
	{"--hero-ink", "var(--color-text)"},
}

// buildTokenAliases returns a :root block defining every tokenAliases
// name that is not already DEFINED anywhere in css. A definition is the
// name followed by a colon ("--shadow:"), which cannot match a var()
// usage ("var(--shadow)") or a longer sibling name ("--shadow-md:").
// Returns "" when nothing is missing.
func buildTokenAliases(css string, logger *zap.Logger) string {
	var missing []string
	var b strings.Builder
	for _, a := range tokenAliases {
		if strings.Contains(css, a.Name+":") {
			continue
		}
		missing = append(missing, a.Name)
		b.WriteString("  " + a.Name + ": " + a.Value + ";\n")
	}
	if len(missing) == 0 {
		return ""
	}
	logger.Info("RenderCSSFromSpecAction: appending compatibility token aliases",
		zap.Strings("aliases", missing))
	return "\n\n/* renderer-enforced compatibility aliases (component vocabulary bridge) */\n:root {\n" +
		b.String() + "}\n"
}
