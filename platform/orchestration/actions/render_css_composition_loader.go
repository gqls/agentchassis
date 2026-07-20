package actions

// FILE: platform/orchestration/actions/render_css_composition_loader.go
//
// Composition loader for the 025_palette_layout_typography_migration
// renderer (Phase 4.2). Reads a css_themes row plus its three linked
// rows — palette, layout, typography_set — in a single SQL query and
// returns them as typed Go maps plus the layout's CSS template string.
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
// Theme resolution order (first match wins):
//  1. config["theme_id"]   — explicit operator/HITL override
//  2. config["theme_name"] — explicit library preview / test override
//  3. siteID parameter      — live site composition lookup
//     (sites.style_collection_id → style_collections.css_theme_id)
//  4. fallbackThemeName arg — emergency fallback, logged as Error

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

	// The layout's declared colour scheme ("light", "dark", "neutral",
	// or "" when the layouts.scheme column is NULL). The renderer uses
	// this to refuse a design_spec background that contradicts the
	// user's layout choice (bugs_open/022).
	LayoutScheme string

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
// Resolution order for the theme row (first match wins):
//  1. config["theme_id"]    — explicit UUID override
//  2. config["theme_name"]  — explicit name override
//  3. siteID                — live site composition lookup
//  4. fallbackThemeName     — emergency fallback (loud-error)
//
// Explicit config overrides (theme_id, theme_name) STILL win over
// siteID lookup. That preserves:
//   - HITL flow: operator picks a theme to render against
//   - Library preview: render an adopted theme without a live site
//   - Testing / dev workflows
//
// siteID is passed by the caller as a plain string (not written into
// config). An empty string disables the site-composition lookup path.
func loadThemeComposition(
	ctx context.Context,
	db *sql.DB,
	config map[string]interface{},
	siteID string,
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
			l.scheme              AS layout_scheme,
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

	// NB: config can legitimately be nil here (workflow JSON with no
	// "config" key unmarshals Step.Config to nil). Reading from a nil
	// map is safe in Go — returns zero values — so we don't need to
	// allocate. We never WRITE to config in this function.
	if themeIDStr, ok := config["theme_id"].(string); ok && themeIDStr != "" {
		row = db.QueryRowContext(ctx, selectSQL+" AND t.id = $1", themeIDStr)
		resolvedBy = "theme_id"
		resolvedValue = themeIDStr
	} else if themeName, ok := config["theme_name"].(string); ok && themeName != "" {
		row = db.QueryRowContext(ctx, selectSQL+" AND t.name = $1", themeName)
		resolvedBy = "theme_name"
		resolvedValue = themeName
	} else if siteThemeID := resolveThemeIDFromSiteContext(ctx, db, siteID, logger); siteThemeID != "" {
		// Site-composition lookup. Traces
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
		layoutScheme                      sql.NullString
		layoutStructureTokensJSON         []byte
		typoFontsJSON, typoScaleJSON      []byte
	)

	err := row.Scan(
		&themeName,
		&paletteID, &paletteName, &paletteColoursJSON,
		&layoutID, &layoutName, &layoutCSSTemplate, &layoutStructureTokensJSON, &layoutScheme,
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
		LayoutScheme:    layoutScheme.String,
		Palette:         paletteColours,
		Structure:       structureTokens,
		Typography:      typoFonts,
		TypographyScale: typoScale,
	}

	logger.Info("loadThemeComposition: loaded",
		zap.String("theme", comp.ThemeName),
		zap.String("palette", comp.PaletteName),
		zap.String("layout", comp.LayoutName),
		zap.String("layout_scheme", comp.LayoutScheme),
		zap.String("typography", comp.TypographyName),
		zap.String("resolved_by", resolvedBy),
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
//	sites (id = siteID) → style_collection_id → style_collections.css_theme_id
//
// Returns the theme id as a string (for direct use in a UUID-typed WHERE),
// or empty string if siteID is empty or no resolution path succeeds.
// Never errors — lookup failures are logged and treated as "no resolution",
// letting loadThemeComposition fall through to the emergency fallback.
func resolveThemeIDFromSiteContext(
	ctx context.Context,
	db *sql.DB,
	siteID string,
	logger *zap.Logger,
) string {

	if siteID == "" {
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
	`, siteID).Scan(&themeID)

	if err == sql.ErrNoRows {
		logger.Warn("resolveThemeIDFromSiteContext: no composition linked to site",
			zap.String("site_id", siteID),
		)
		return ""
	}
	if err != nil {
		logger.Warn("resolveThemeIDFromSiteContext: lookup failed",
			zap.String("site_id", siteID),
			zap.Error(err),
		)
		return ""
	}
	if !themeID.Valid || themeID.String == "" {
		return ""
	}

	logger.Info("resolveThemeIDFromSiteContext: resolved from site composition",
		zap.String("site_id", siteID),
		zap.String("css_theme_id", themeID.String),
	)
	return themeID.String
}
