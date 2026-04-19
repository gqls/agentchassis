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
//   1. design_reference.typography.reference_values  (from fingerprint, if adopted)
//   2. design_intent.typography.reference_values     (semantic brief)
//   3. mission.preferred_typography                  (if human pre-specified)
//   4. (fall through to resolveTypographySet's own default: sans-modern)
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
	"strings"

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
	category, industryTags := readClassificationFieldsFromContext(
		ctx, params, siteID, logger,
	)

	// baseName for the typography-set's name if we end up inserting one.
	baseName := slugifyForTypography(domain)
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
//  1. design_reference.typography.reference_values
//  2. design_intent.typography.reference_values
//  3. mission.preferred_typography
//
// Empty-or-missing at every step → returns empty map and "none". The
// caller then lets resolveTypographySet apply its sans-modern default.
func extractTypographySignal(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	logger *zap.Logger,
) (map[string]string, string) {

	// 1. design_reference (from fingerprint)
	if ref, ok := readSpecAspectMap(ctx, params, siteID, "design_reference", logger); ok {
		if fonts := extractReferenceValues(ref, "typography"); len(fonts) > 0 {
			if fonts["font_family"] != "" {
				return fonts, "design_reference"
			}
		}
	}

	// 2. design_intent (semantic)
	if intent, ok := readSpecAspectMap(ctx, params, siteID, "design_intent", logger); ok {
		if fonts := extractReferenceValues(intent, "typography"); len(fonts) > 0 {
			if fonts["font_family"] != "" {
				return fonts, "design_intent"
			}
		}
	}

	// 3. mission.preferred_typography
	if mission, ok := readSpecAspectMap(ctx, params, siteID, "mission", logger); ok {
		if raw, exists := mission["preferred_typography"]; exists {
			if m, ok := raw.(map[string]interface{}); ok {
				fonts := mapInterfaceToString(m)
				if fonts["font_family"] != "" {
					return fonts, "mission_hint"
				}
			}
		}
	}

	return map[string]string{}, "none"
}

// readSpecAspectMap loads a spec aspect. Tries collected_data first
// (if the workflow already loaded it via read_site_spec), then falls back
// to a direct DB read.
func readSpecAspectMap(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	aspect string,
	logger *zap.Logger,
) (map[string]interface{}, bool) {

	// Check collected_data under conventional paths used by read_site_spec
	candidates := []string{
		"site_specs.specs." + aspect,
		aspect + "_spec.data",
		aspect,
	}
	for _, p := range candidates {
		raw := datahelpers.ExtractNestedField(params.CollectedData, p)
		if raw != nil {
			unwrapped := datahelpers.UnwrapDeep(raw, logger)
			if m, ok := unwrapped.(map[string]interface{}); ok && len(m) > 0 {
				return m, true
			}
		}
	}

	// Fall back to DB read. loadCurrentSpecData lives in
	// validate_composition_inputs_action.go.
	data, found, err := loadCurrentSpecData(ctx, params.DB, siteID, aspect)
	if err != nil {
		logger.Warn("readSpecAspectMap: DB read failed",
			zap.String("aspect", aspect), zap.Error(err),
		)
		return nil, false
	}
	return data, found
}

// extractReferenceValues pulls the fonts map out of a design spec that
// follows the documented schema:
//
//	{
//	  "typography": {
//	    "character": "...",
//	    "reference_values": {
//	      "font_family": "Inter, ...",
//	      "heading_font": "..."
//	    }
//	  }
//	}
//
// Falls through to direct keys on typography if reference_values is absent
// (so shapes that are simpler than the full schema still work).
func extractReferenceValues(spec map[string]interface{}, key string) map[string]string {
	typoRaw, ok := spec[key]
	if !ok {
		return nil
	}
	typoMap, ok := typoRaw.(map[string]interface{})
	if !ok {
		return nil
	}
	// Prefer nested reference_values
	if refRaw, has := typoMap["reference_values"]; has {
		if refMap, ok := refRaw.(map[string]interface{}); ok {
			return mapInterfaceToString(refMap)
		}
	}
	// Fall back: flat shape
	return mapInterfaceToString(typoMap)
}

// mapInterfaceToString converts a map[string]interface{} to
// map[string]string, keeping only string values and trimming whitespace.
// Non-string values (arrays, numbers, nested maps) are dropped.
func mapInterfaceToString(in map[string]interface{}) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		if s, ok := v.(string); ok {
			s = strings.TrimSpace(s)
			if s != "" {
				out[k] = s
			}
		}
	}
	return out
}

// readClassificationFieldsFromContext pulls category and industry_tags
// following the same logic as extractClassificationTags in
// resolve_composition_layout_action.go. Kept as a small separate function
// here to avoid cross-file dependency coupling.
func readClassificationFieldsFromContext(
	ctx context.Context,
	params ActionParams,
	siteID uuid.UUID,
	logger *zap.Logger,
) (string, []string) {

	classPath := "validated_inputs.classification"
	if cs, ok := params.StepConfig.Config["classification_source"].(string); ok && cs != "" {
		classPath = cs
	}

	var classData map[string]interface{}
	classRaw := datahelpers.ExtractNestedField(params.CollectedData, classPath)
	if classRaw != nil {
		unwrapped := datahelpers.UnwrapDeep(classRaw, logger)
		if m, ok := unwrapped.(map[string]interface{}); ok {
			classData = m
		}
	}

	if len(classData) == 0 {
		// Fall back to spec read — validate_composition_inputs should
		// have caught a missing classification earlier, so this is just
		// a safety net.
		if data, found, err := loadCurrentSpecData(ctx, params.DB, siteID, "classification"); err == nil && found {
			classData = data
		}
	}

	category, _ := classData["category"].(string)
	var tags []string
	if raw, ok := classData["industry_tags"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, t := range v {
				if s, ok := t.(string); ok {
					tags = append(tags, s)
				}
			}
		case []string:
			tags = append(tags, v...)
		}
	}
	return category, tags
}

// slugifyForTypography makes a domain safe for use inside a typography_set
// name. Postgres identifiers allow lowercase/digits/underscores; we also
// replace dots and hyphens to keep the name readable.
func slugifyForTypography(domain string) string {
	d := strings.ToLower(strings.TrimSpace(domain))
	d = strings.ReplaceAll(d, ".", "-")
	// Keep a-z, 0-9, hyphen, underscore. Drop anything else.
	out := make([]byte, 0, len(d))
	for i := 0; i < len(d); i++ {
		c := d[i]
		switch {
		case c >= 'a' && c <= 'z',
			c >= '0' && c <= '9',
			c == '-' || c == '_':
			out = append(out, c)
		}
	}
	// Postgres identifiers cap at 63 chars; plus we prefix with "typography-"
	// which is 11 chars. Leave some headroom for resolveUniqueNameInTx's
	// collision suffix.
	const maxLen = 40
	if len(out) > maxLen {
		out = out[:maxLen]
	}
	return string(out)
}
