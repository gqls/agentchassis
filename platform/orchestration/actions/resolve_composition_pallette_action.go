// FILE: platform/orchestration/actions/resolve_composition_palette_action.go
//
// ResolveCompositionPaletteAction is site-design-planner's palette picker.
// Runs after resolve_composition_layout and resolve_composition_typography,
// before install_site_composition.
//
// Palette philosophy:
//   Palettes are ALWAYS site-specific. Unlike layouts (which are shared
//   library grammar) and typography (which matches by font_family across
//   sites), every site gets its own palettes row. Library reuse of
//   palettes would be lying — "this site's brand colour is #c8102e" is a
//   statement about one specific site, not a reusable template.
//
// Signal priority cascade:
//   1. mission.preferred_palette                    (human pre-specified)
//   2. design_intent.palette.reference_values       (semantic brief; for adopted
//                                                    sites this carries the
//                                                    fingerprint colours PLUS the
//                                                    accent/secondary enrichment)
//   3. design_reference fingerprint                 (suggested_mapping / css_variables
//                                                    via paletteFromDesignReference —
//                                                    FALLBACK, only when design_intent
//                                                    produced no palette; adopted sites
//                                                    only)
//   4. Inherit from chosen layout's default library palette
//      (e.g. brochure-formal → "default" palette, high-energy → "boxing")
//      — only the CORE slots (primary/secondary/accent/background/
//      surface/text/text_muted/border) — so the site has a coherent
//      starting point. Core slots can still be overridden later by
//      design_spec merge at render time.
//   5. The "default" library palette (last resort).
//
// NOTE: design_intent stays ahead of the design_reference fingerprint on
// purpose. The fingerprint's suggested_mapping is typically 6 colour slots
// with no accent; the generated design_intent adds accent/secondary. Reading
// the fingerprint first would strip an adopted site's accent colour.
//
// Separation of concerns:
//   - The INSERT logic lives in `createPalette` (fork_theme_composition.go)
//     — shared with createPaletteForFork.
//   - This action is the thin workflow wrapper: extract colours via the
//     cascade, then delegate.
//
// Inputs:
//   - site_id              (path-resolved, required)
//   - selected_layout_id   (path-resolved, required) — id of the layout
//                           chosen by resolve_composition_layout, used
//                           to inherit a library palette as fallback
//
// Config literals:
//   - none currently
//
// Returns:
//   {
//     "palette_id":   "uuid-string",
//     "palette_name": "palette-<slug>",
//     "source":       "design_reference" | "mission_hint" |
//                     "design_intent" | "layout_library_inherit" |
//                     "fallback_default",
//     "colour_count": 8,
//     "colours":      { "primary": "#...", ... },  // what was written
//   }
//
// Registration (add to registry.go):
//
//   "resolve_composition_palette": {
//       Handler:     ResolveCompositionPaletteAction,
//       Category:    "site",
//       Description: "Extract palette colours via signal cascade and create palette row",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ResolveCompositionPaletteInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id", "selected_layout_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec(
		"resolve_composition_palette",
		ResolveCompositionPaletteInputSpec,
	)
}

// ResolveCompositionPaletteAction is the workflow entry point.
func ResolveCompositionPaletteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "resolve_composition_palette"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ResolveCompositionPaletteInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	layoutIDStr := inputs.Get("selected_layout_id")
	layoutID, err := uuid.Parse(layoutIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid selected_layout_id %q: %w", layoutIDStr, err)
	}

	// Domain — needed for palette naming and source_domain column.
	var domain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)

	// Walk the priority cascade.
	colours, source, err := extractPaletteSignal(
		ctx, params, siteID, layoutID, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("extract palette signal: %w", err)
	}
	if len(colours) == 0 {
		return nil, fmt.Errorf(
			"no palette signal available — neither spec cascade nor layout library produced colours for site %s",
			siteID,
		)
	}

	// Classification — for category + industry_tags on the new row.
	category, industryTags := readClassificationFromContext(ctx, params, siteID, logger)

	baseName := slugifyForCompositionName(domain)
	if baseName == "" {
		baseName = "site-" + siteID.String()[:8]
	}
	displayName := fmt.Sprintf("Palette for %s", domain)

	// Delegate to createPalette inside a transaction. Composition-installed
	// palettes don't need HITL review by default — the site is using its
	// own colours, no library promotion is implied.
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	paletteID, err := createPalette(
		ctx, tx,
		colours,
		baseName, displayName,
		category, industryTags,
		siteID, domain,
		nil,              // no parent palette — fresh composition
		"adopted", false, // origin=adopted, needs_review=false
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("createPalette failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Read back the name so the output is complete.
	var paletteName string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT name FROM palettes WHERE id = $1`, paletteID,
	).Scan(&paletteName)

	logger.Info("ResolveCompositionPaletteAction: resolved",
		zap.String("site_id", siteID.String()),
		zap.String("palette_id", paletteID.String()),
		zap.String("palette_name", paletteName),
		zap.String("source", source),
		zap.Int("colour_count", len(colours)),
	)

	return map[string]interface{}{
		"palette_id":   paletteID.String(),
		"palette_name": paletteName,
		"source":       source,
		"colour_count": len(colours),
		"colours":      colours,
	}, nil
}

// extractPaletteSignal walks the cascade to find palette colours.
// Returns the colours map and the name of the source slot that won.
// Never returns an empty map unless every source (including layout
// library inherit) was blank — which would be a data gap the caller
// should error on.
func extractPaletteSignal(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	layoutID uuid.UUID,
	logger *zap.Logger,
) (map[string]string, string, error) {

	// 1. mission.preferred_palette (human hint — most authoritative)
	if mission, ok := loadSpecAspectFromContext(ctx, params, siteID, "mission", logger); ok {
		if raw, exists := mission["preferred_palette"]; exists {
			if m, ok := raw.(map[string]interface{}); ok {
				colours := mapInterfaceToStrings(m)
				if len(colours) > 0 {
					logger.Info("extractPaletteSignal: using mission.preferred_palette",
						zap.Int("colour_count", len(colours)))
					return colours, "mission_hint", nil
				}
			}
		}
	}

	// 2. design_intent.palette.reference_values (semantic brief).
	//    For an adopted site this is the fingerprint colours PLUS the
	//    accent/secondary the enrichment step (generate_design_intent) added,
	//    so it MUST stay ahead of the raw design_reference fingerprint below —
	//    otherwise an adopted site would lose its accent colour to the leaner
	//    suggested_mapping. (For a fresh site this carries the classifier's
	//    structured palette, once the classifier emits one.)
	if intent, ok := loadSpecAspectFromContext(ctx, params, siteID, "design_intent", logger); ok {
		if colours := extractReferenceValuesFromSpec(intent, "palette"); len(colours) > 0 {
			logger.Info("extractPaletteSignal: using design_intent",
				zap.Int("colour_count", len(colours)))
			return colours, "design_intent", nil
		}
	}

	// 3. design_reference fingerprint (suggested_mapping / css_variables).
	//    FALLBACK: only reached when design_intent produced no palette — e.g.
	//    generate_design_intent failed or returned an empty palette. Reads the
	//    fingerprint's real shape via paletteFromDesignReference; the
	//    palette.reference_values key this slot used to read is never written
	//    by extract_design_fingerprint, so the old slot could never fire.
	//    Has no effect on fresh sites (they have no design_reference).
	if ref, ok := loadSpecAspectFromContext(ctx, params, siteID, "design_reference", logger); ok {
		if colours := paletteFromDesignReference(ref); len(colours) > 0 {
			logger.Info("extractPaletteSignal: using design_reference fingerprint (design_intent had no palette)",
				zap.Int("colour_count", len(colours)))
			return colours, "design_reference", nil
		}
	}

	// 4. Inherit from the chosen layout's default library palette.
	//    Many layouts have a "default library palette" mapping inherent
	//    to their character: high-energy → boxing palette, soft-editorial
	//    → bakery/calm-minimal, technical-precise → premium-elegant.
	//
	//    The mapping lives in css_themes: each layout is linked by
	//    existing seed css_themes rows to its canonical palette. We find
	//    a seed theme that uses this layout and read its palette.
	colours, paletteName, err := inheritPaletteFromLayout(ctx, params.DB, layoutID, logger)
	if err != nil {
		logger.Warn("extractPaletteSignal: layout inherit failed",
			zap.Error(err),
			zap.String("layout_id", layoutID.String()))
	}
	if len(colours) > 0 {
		logger.Info("extractPaletteSignal: inheriting from layout's library palette",
			zap.String("inherited_palette_name", paletteName),
			zap.Int("colour_count", len(colours)))
		return colours, "layout_library_inherit", nil
	}

	// 5. Final fallback: the "default" palette.
	colours, err = loadDefaultLibraryPalette(ctx, params.DB, logger)
	if err != nil {
		return nil, "", fmt.Errorf("load default palette fallback: %w", err)
	}
	logger.Warn("extractPaletteSignal: falling through to default library palette",
		zap.Int("colour_count", len(colours)))
	return colours, "fallback_default", nil
}

// inheritPaletteFromLayout looks for a seed css_themes row that uses the
// given layout and returns its linked palette's colours. This gives new
// compositions a "sensible default pairing" when the site itself has no
// palette signal — boxing sites end up looking boxing-y even when they
// haven't told us their colours yet.
//
// Returns (nil, "", nil) if no seed theme uses this layout — that's not
// an error, it just means the fallback cascade continues.
func inheritPaletteFromLayout(
	ctx context.Context,
	db *sql.DB,
	layoutID uuid.UUID,
	logger *zap.Logger,
) (map[string]string, string, error) {

	var paletteName string
	var coloursJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT p.name, p.colours
		FROM css_themes t
		JOIN palettes p ON p.id = t.palette_id
		WHERE t.layout_id = $1
		  AND t.origin = 'seed'
		  AND t.is_active = true
		  AND p.is_active = true
		ORDER BY t.created_at ASC
		LIMIT 1
	`, layoutID).Scan(&paletteName, &coloursJSON)

	if err == sql.ErrNoRows {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", fmt.Errorf("query seed theme for layout: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(coloursJSON, &raw); err != nil {
		return nil, "", fmt.Errorf("unmarshal inherited palette: %w", err)
	}
	return mapInterfaceToStrings(raw), paletteName, nil
}

// loadDefaultLibraryPalette reads the "default" palette directly. Last-
// resort fallback used when every other signal source is blank. If this
// fails the whole action errors out — the data model is broken if there
// is no default palette to fall back to.
func loadDefaultLibraryPalette(
	ctx context.Context,
	db *sql.DB,
	logger *zap.Logger,
) (map[string]string, error) {

	var coloursJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT colours FROM palettes
		WHERE name = 'default' AND is_active = true
		LIMIT 1
	`).Scan(&coloursJSON)
	if err != nil {
		return nil, fmt.Errorf("default palette not found: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(coloursJSON, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal default palette: %w", err)
	}
	return mapInterfaceToStrings(raw), nil
}
