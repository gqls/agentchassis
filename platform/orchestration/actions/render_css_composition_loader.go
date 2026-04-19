package actions

// FILE: platform/orchestration/actions/render_css_composition_loader.go
//
// Composition loader for the 025_palette_layout_typography_migration
// renderer (Phase 4.2). Reads a css_themes row plus its three linked
// rows — palette, layout, typography_set — in a single SQL query and
// returns them as typed Go maps plus the layout's CSS template string.
//
// Phase 4.2 of the migration: this loader lands without being wired
// into RenderCSSFromSpecAction. Phase 4.3 replaces the legacy
// loadCSSGoTemplate call with this one.
//
// Error contract (per migration plan):
//   - Theme row missing           → hard error
//   - Any of palette_id / layout_id / typography_set_id is NULL
//                                 → hard error (migration gaps are
//                                   audit events, not render-path
//                                   silent fallbacks)
//   - JSONB parse failure on any of colours / structure_tokens /
//     fonts                       → hard error (data shape bug)
//
// The intent is to make failures loud and easy to track back to a
// specific css_themes row. Silent fallback to standard-brochure (as
// loadCSSGoTemplate did) hides the underlying problem.

import (
	"context"
	"database/sql"
	"fmt"

	"go.uber.org/zap"
)

// themeComposition is the result of resolving a css_themes row through
// its FKs to palette + layout + typography_set. All five maps are
// non-nil after a successful load (possibly empty, never nil).
type themeComposition struct {
	// Identifying names (for logging and debugging; no caller depends
	// on them for rendering).
	ThemeName      string
	PaletteName    string
	LayoutName     string
	TypographyName string

	// The layout's CSS Go-template string. Passed into text/template
	// by the caller.
	LayoutTemplate string

	// Merged-ready inputs. Callers pass Palette and Typography through
	// buildPaletteMap/buildTypographyMap with the design_spec before
	// handing the result to the template; Structure goes straight to
	// the template unchanged.
	Palette    map[string]string
	Structure  map[string]string
	Typography map[string]string

	// Typography scale (h1_ratio, line_height, base_size) is returned
	// but not currently consumed — reserved for a later iteration that
	// may feed heading scale into the layout template.
	TypographyScale map[string]string
}

// loadThemeComposition fetches a css_themes row plus its three linked
// rows in one JOIN, returns a themeComposition struct populated with
// parsed JSONB. Hard-errors on any missing FK or unparseable JSONB.
//
// Resolution order for the theme row:
//  1. config["theme_id"]   (UUID string, if present and non-empty)
//  2. config["theme_name"] (text name)
//  3. fallback arg themeName (e.g. "standard-brochure")
//
// This mirrors loadCSSGoTemplate's resolution order for behavioural
// parity at the theme-selection boundary. The difference is what
// happens after resolution: the old loader returned just css_template
// text; this loader returns the full composition.
func loadThemeComposition(
	ctx context.Context,
	db *sql.DB,
	config map[string]interface{},
	fallbackThemeName string,
	logger *zap.Logger,
) (*themeComposition, error) {

	// Build the WHERE clause based on which identifier we have.
	// The SELECT is identical; only the filter differs.
	const selectSQL = `
		SELECT
			t.name                AS theme_name,
			p.id                  AS palette_id,
			p.name                AS palette_name,
			p.colours             AS palette_colours,
			l.id                  AS layout_id,
			l.name                AS layout_name,
			l.css_template        AS layout_css_template,
			l.structure_tokens    AS layout_structure_tokens,
			ts.id                 AS typo_id,
			ts.name               AS typo_name,
			ts.fonts              AS typo_fonts,
			ts.scale              AS typo_scale
		FROM css_themes t
		LEFT JOIN palettes        p  ON p.id  = t.palette_id         AND p.is_active  = true
		LEFT JOIN layouts         l  ON l.id  = t.layout_id          AND l.is_active  = true
		LEFT JOIN typography_sets ts ON ts.id = t.typography_set_id  AND ts.is_active = true
		WHERE t.is_active = true
	`

	var (
		row           *sql.Row
		resolvedBy    string
		resolvedValue string
	)

	if themeIDStr, ok := config["theme_id"].(string); ok && themeIDStr != "" {
		row = db.QueryRowContext(ctx, selectSQL+" AND t.id = $1", themeIDStr)
		resolvedBy = "theme_id"
		resolvedValue = themeIDStr
	} else {
		themeName := fallbackThemeName
		if n, ok := config["theme_name"].(string); ok && n != "" {
			themeName = n
		}
		row = db.QueryRowContext(ctx, selectSQL+" AND t.name = $1", themeName)
		resolvedBy = "theme_name"
		resolvedValue = themeName
	}

	// Problem:
	//   The action currently resolves a theme via:
	//     1. config["theme_id"]   (explicit override)
	//     2. config["theme_name"] (explicit override)
	//     3. fallback arg "standard-brochure" (silent default)
	//
	//   Step 3 is the "first render wrong layout" bug site-design-planner
	//   was built to eliminate: if a site has composition installed but the
	//   workflow config doesn't say so, the renderer silently falls back to
	//   standard-brochure and produces the wrong CSS.
	//
	//   webdesign-agent's workflow still has `config: { "theme_name": "standard-brochure" }`
	//   hardcoded on its generate_css step. That's config carried from before
	//   composition existed. Leaving it forces standard-brochure every time,
	//   regardless of what site-design-planner installed. No amount of
	//   site_context plumbing helps as long as that config literal is present.
	//
	// this fix:
	//   Insert a new resolution step 3 AHEAD of the existing fallback:
	//     3. site_context.site_id → look up sites.style_collection_id →
	//        style_collections.css_theme_id → use that theme_id
	//
	//   If that fails (no site_id in context, or no composition for the site):
	//     4. Old step 3 behaviour as LAST-RESORT emergency fallback, with
	//        a logger.Error ("site-design-planner did not run for this site")
	//        distinguishing it from normal operation.
	//
	//   Explicit config overrides (theme_id, theme_name) STILL win over
	//   site_id lookup. That preserves:
	//     - HITL flow: operator picks a theme to render against
	//     - Library preview: render an adopted theme without a live site
	//     - Testing / dev workflows
	//
	//   webdesign-agent's workflow will need its hardcoded
	//   "theme_name": "standard-brochure" REMOVED (tracked as a separate
	//   workflow edit — once this patch lands, deleting that config literal
	//   makes the resolver find site_context.site_id and pick up the real
	//   composition).

	if themeIDStr, ok := config["theme_id"].(string); ok && themeIDStr != "" {
		row = db.QueryRowContext(ctx, selectSQL+" AND t.id = $1", themeIDStr)
		resolvedBy = "theme_id"
		resolvedValue = themeIDStr
	} else if themeName, ok := config["theme_name"].(string); ok && themeName != "" {
		row = db.QueryRowContext(ctx, selectSQL+" AND t.name = $1", themeName)
		resolvedBy = "theme_name"
		resolvedValue = themeName
	} else if siteThemeID := resolveThemeIDFromSiteContext(ctx, db, config, logger); siteThemeID != "" {
		// NEW: site-composition lookup. Uses site_id from config (which
		// the caller populates from collectedData via path). Traces
		// sites.style_collection_id → style_collections.css_theme_id.
		row = db.QueryRowContext(ctx, selectSQL+" AND t.id = $1", siteThemeID)
		resolvedBy = "site_composition"
		resolvedValue = siteThemeID
	} else {
		// EMERGENCY fallback to the provided default (historically
		// "standard-brochure"). Loud-error — this path means either
		// site-design-planner didn't run, or the renderer was called
		// without a site_id, or the sites row isn't linked to a
		// style_collection. All three are worth investigating.
		logger.Error(
			"loadThemeComposition: no theme resolution path succeeded — using emergency fallback",
			zap.String("emergency_fallback", fallbackThemeName),
			zap.String("recommendation", "verify site-design-planner ran for this site and sites.style_collection_id is set"),
		)
		row = db.QueryRowContext(ctx, selectSQL+" AND t.name = $1", fallbackThemeName)
		resolvedBy = "emergency_fallback"
		resolvedValue = fallbackThemeName
	}

	var (
		themeName                         string
		paletteID, layoutID, typoID       sql.NullString
		paletteName, layoutName, typoName sql.NullString
		paletteColoursJSON                []byte
		layoutCSSTemplate                 sql.NullString
		layoutStructureTokensJSON         []byte
		typoFontsJSON, typoScaleJSON      []byte
	)

	err := row.Scan(
		&themeName,
		&paletteID, &paletteName, &paletteColoursJSON,
		&layoutID, &layoutName, &layoutCSSTemplate, &layoutStructureTokensJSON,
		&typoID, &typoName, &typoFontsJSON, &typoScaleJSON,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf(
			"theme not found: %s=%q",
			resolvedBy, resolvedValue,
		)
	}
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load theme composition (%s=%q): %w",
			resolvedBy, resolvedValue, err,
		)
	}

	// Hard-error on any missing FK. Per migration plan: "migration
	// gaps are audit events, not render-path silent fallbacks." A
	// nullable FK column on a row that should be fully linked is a
	// data integrity problem, not a render-time concern.
	if !paletteID.Valid {
		return nil, fmt.Errorf(
			"theme %q (%s=%q) has NULL palette_id — migration gap; run Phase 3 mapping",
			themeName, resolvedBy, resolvedValue,
		)
	}
	if !layoutID.Valid || !layoutCSSTemplate.Valid {
		return nil, fmt.Errorf(
			"theme %q (%s=%q) has NULL layout_id or inactive layout — migration gap; run Phase 3 mapping",
			themeName, resolvedBy, resolvedValue,
		)
	}
	if !typoID.Valid {
		return nil, fmt.Errorf(
			"theme %q (%s=%q) has NULL typography_set_id — migration gap; run Phase 3 mapping",
			themeName, resolvedBy, resolvedValue,
		)
	}

	// Parse each JSONB payload. jsonbToStringMap tolerates empty input
	// (returns an empty non-nil map) but returns an error on malformed
	// JSON — which is a data integrity problem worth surfacing.
	paletteColours, err := jsonbToStringMap(paletteColoursJSON)
	if err != nil {
		return nil, fmt.Errorf(
			"theme %q: failed to parse palette %q colours JSONB: %w",
			themeName, paletteName.String, err,
		)
	}

	structureTokens, err := jsonbToStringMap(layoutStructureTokensJSON)
	if err != nil {
		return nil, fmt.Errorf(
			"theme %q: failed to parse layout %q structure_tokens JSONB: %w",
			themeName, layoutName.String, err,
		)
	}

	typoFonts, err := jsonbToStringMap(typoFontsJSON)
	if err != nil {
		return nil, fmt.Errorf(
			"theme %q: failed to parse typography_set %q fonts JSONB: %w",
			themeName, typoName.String, err,
		)
	}

	typoScale, err := jsonbToStringMap(typoScaleJSON)
	if err != nil {
		return nil, fmt.Errorf(
			"theme %q: failed to parse typography_set %q scale JSONB: %w",
			themeName, typoName.String, err,
		)
	}

	comp := &themeComposition{
		ThemeName:       themeName,
		PaletteName:     paletteName.String,
		LayoutName:      layoutName.String,
		TypographyName:  typoName.String,
		LayoutTemplate:  layoutCSSTemplate.String,
		Palette:         paletteColours,
		Structure:       structureTokens,
		Typography:      typoFonts,
		TypographyScale: typoScale,
	}

	logger.Info("loadThemeComposition: loaded",
		zap.String("theme", comp.ThemeName),
		zap.String("palette", comp.PaletteName),
		zap.String("layout", comp.LayoutName),
		zap.String("typography", comp.TypographyName),
		zap.Int("palette_keys", len(comp.Palette)),
		zap.Int("structure_keys", len(comp.Structure)),
		zap.Int("typography_keys", len(comp.Typography)),
		zap.Int("layout_template_bytes", len(comp.LayoutTemplate)),
	)

	return comp, nil
}

// resolveThemeIDFromSiteContext tries to find the site's installed
// css_theme_id via the path:
//
//	sites (id = site_id) → style_collection_id → style_collections.css_theme_id
//
// Returns the theme id as a string (for direct use in a UUID-typed WHERE),
// or empty string if no resolution path succeeds. Never errors — lookup
// failures are logged and treated as "no resolution", letting
// loadThemeComposition fall through to the emergency fallback.
//
// site_id is read from config["site_id"] — the caller (the action) is
// responsible for populating it from collectedData before calling
// loadThemeComposition. Config-based plumbing (rather than a new param)
// keeps the loader's function signature unchanged.
func resolveThemeIDFromSiteContext(
	ctx context.Context,
	db *sql.DB,
	config map[string]interface{},
	logger *zap.Logger,
) string {

	siteIDStr, ok := config["site_id"].(string)
	if !ok || siteIDStr == "" {
		return ""
	}

	var themeID sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT sc.css_theme_id::text
		FROM sites s
		JOIN style_collections sc ON sc.id = s.style_collection_id
		WHERE s.id = $1
		  AND s.style_collection_id IS NOT NULL
		  AND sc.is_active = true
		  AND sc.css_theme_id IS NOT NULL
	`, siteIDStr).Scan(&themeID)

	if err == sql.ErrNoRows {
		logger.Warn("resolveThemeIDFromSiteContext: no composition linked to site",
			zap.String("site_id", siteIDStr),
		)
		return ""
	}
	if err != nil {
		logger.Warn("resolveThemeIDFromSiteContext: lookup failed",
			zap.String("site_id", siteIDStr),
			zap.Error(err),
		)
		return ""
	}
	if !themeID.Valid || themeID.String == "" {
		return ""
	}

	logger.Info("resolveThemeIDFromSiteContext: resolved from site composition",
		zap.String("site_id", siteIDStr),
		zap.String("css_theme_id", themeID.String),
	)
	return themeID.String
}
