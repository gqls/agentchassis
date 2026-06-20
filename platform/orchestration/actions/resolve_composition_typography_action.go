// FILE: platform/orchestration/actions/resolve_composition_typography_action.go
//
// ResolveCompositionTypographyAction is site-design-planner's typography picker.
// Runs after resolve_composition_layout, before resolve_composition_palette.
//
// Separation of concerns:
//   - The match-or-insert LOGIC lives in `resolveTypographySet`
//     (fork_theme_composition.go) — shared with fork_theme_from_site.
//   - This action is the thin workflow wrapper: extract typography signal
//     from the site's specs using a priority cascade, then delegate.
//
// Signal priority cascade:
//   1. design_intent.typography.reference_values     (semantic brief)
//   2. mission.preferred_typography                  (if human pre-specified)
//   3. design_reference fingerprint                  (suggested_mapping.font_family /
//                                                    css_variables via
//                                                    typographyFromDesignReference —
//                                                    FALLBACK, only when design_intent
//                                                    produced no typography; adopted
//                                                    sites only)
//   4. (fall through to resolveTypographySet's own default: sans-modern)
//
// design_intent stays ahead of the design_reference fingerprint: the generated
// design_intent carries the font stack plus a heading_font, the raw fingerprint
// only the family.
//
// The chosen layout may also imply typography (docs layouts often pair
// with mono-technical, editorial with serif-editorial). That's a later
// refinement — for now the resolver trusts the site's own signals over
// layout defaults.
//
// Inputs:
//   - site_id                    (path-resolved, required)
//   - classification_data        (optional — read from collected_data or specs)
//   - selected_layout_id         (optional — future: layout-aware default)
//
// Config literals:
//   - classification_source (optional) — path to classification data,
//                                         default "validated_inputs.classification"
//
// Returns:
//   {
//     "typography_set_id":   "uuid-string",
//     "typography_name":     "serif-editorial" | "sans-modern" | "typography-xxxxx",
//     "matched_existing":    true | false,
//     "source":              "design_reference" | "design_intent" |
//                            "mission_hint" | "fallback_sans_modern",
//     "font_family":         "the font_family string resolved" | "",
//   }
//
// Registration (add to registry.go):
//
//   "resolve_composition_typography": {
//       Handler:     ResolveCompositionTypographyAction,
//       Category:    "site",
//       Description: "Pick a typography_set by font-family match with spec cascade",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ResolveCompositionTypographyInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec(
		"resolve_composition_typography",
		ResolveCompositionTypographyInputSpec,
	)
}

// ResolveCompositionTypographyAction is the workflow entry point.
func ResolveCompositionTypographyAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "resolve_composition_typography"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ResolveCompositionTypographyInputSpec,
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

	// Domain is needed if we insert a new row; also used for logging.
	var domain string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)

	// Build the typography signal via the priority cascade.
	fonts, source := extractTypographySignal(ctx, params, siteID, logger)

	// Classification — used for industry_tags and category on any new row.
	category, industryTags := readClassificationFromContext(
		ctx, params, siteID, logger,
	)

	// baseName for the typography-set's name if we end up inserting one.
	baseName := slugifyForCompositionName(domain)
	if baseName == "" {
		baseName = "site-" + siteID.String()[:8]
	}
	displayName := fmt.Sprintf("Typography from %s", domain)

	// Delegate to the shared resolver inside a transaction.
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	typoID, matched, err := resolveTypographySet(
		ctx, tx,
		fonts,
		baseName, displayName,
		category, industryTags,
		siteID, domain,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("resolveTypographySet failed: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	// Read back the name so the output is complete.
	var typoName string
	_ = params.DB.QueryRowContext(ctx,
		`SELECT name FROM typography_sets WHERE id = $1`, typoID,
	).Scan(&typoName)

	// When we fell back to sans-modern because no signal was present, the
	// source records that rather than the cascade slot it came from.
	effectiveSource := source
	if len(fonts) == 0 || fonts["font_family"] == "" {
		effectiveSource = "fallback_sans_modern"
	}

	logger.Info("ResolveCompositionTypographyAction: resolved",
		zap.String("site_id", siteID.String()),
		zap.String("typography_set_id", typoID.String()),
		zap.String("typography_name", typoName),
		zap.Bool("matched_existing", matched),
		zap.String("source", effectiveSource),
		zap.String("font_family", fonts["font_family"]),
	)

	return map[string]interface{}{
		"typography_set_id": typoID.String(),
		"typography_name":   typoName,
		"matched_existing":  matched,
		"source":            effectiveSource,
		"font_family":       fonts["font_family"],
	}, nil
}

// extractTypographySignal walks the priority cascade to find a typography
// signal for the site. Returns the fonts map (possibly empty) and the name
// of the source that provided it.
//
// Cascade:
//  1. design_intent.typography.reference_values
//  2. mission.preferred_typography
//  3. design_reference fingerprint (suggested_mapping.font_family / css_variables)
//
// Empty-or-missing at every step → returns empty map and "none". The
// caller then lets resolveTypographySet apply its sans-modern default.
func extractTypographySignal(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	logger *zap.Logger,
) (map[string]string, string) {

	// 1. design_intent.typography.reference_values (semantic). For an adopted
	//    site this carries the fingerprint's font stack plus a heading_font, so
	//    it stays ahead of the raw design_reference fingerprint below.
	if intent, ok := loadSpecAspectFromContext(ctx, params, siteID, "design_intent", logger); ok {
		if fonts := extractReferenceValuesFromSpec(intent, "typography"); fonts["font_family"] != "" {
			return fonts, "design_intent"
		}
	}

	// 2. mission.preferred_typography
	if mission, ok := loadSpecAspectFromContext(ctx, params, siteID, "mission", logger); ok {
		if raw, exists := mission["preferred_typography"]; exists {
			if m, ok := raw.(map[string]interface{}); ok {
				fonts := mapInterfaceToStrings(m)
				if fonts["font_family"] != "" {
					return fonts, "mission_hint"
				}
			}
		}
	}

	// 3. design_reference fingerprint (suggested_mapping.font_family / css_variables).
	//    FALLBACK: reached only when design_intent produced no typography. Reads the
	//    fingerprint's real shape via typographyFromDesignReference; the
	//    typography.reference_values key this slot used to read is never written by
	//    extract_design_fingerprint. Has no effect on fresh sites.
	if ref, ok := loadSpecAspectFromContext(ctx, params, siteID, "design_reference", logger); ok {
		if fonts := typographyFromDesignReference(ref); fonts["font_family"] != "" {
			return fonts, "design_reference"
		}
	}

	return map[string]string{}, "none"
}
