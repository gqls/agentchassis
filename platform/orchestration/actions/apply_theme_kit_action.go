// FILE: platform/orchestration/actions/apply_theme_kit_action.go
//
// ApplyThemeKitAction materializes a theme_kits row's defaults into ONE
// site's own rows. It never creates a live binding — see the theme-kit
// plan, session "themes" (docs026 register entry: design-composition.md,
// theme_kits vs css_themes).
//
// What it writes, each independently, each a DEFAULT not a constraint:
//   1. site_specs aspect 'theme_kit_adoption' — lineage record (always).
//   2. site_specs aspect 'design_intent' — merges the kit's resolved
//      palette/typography colours into .palette.reference_values /
//      .typography.reference_values, ONLY if the site has none there yet,
//      or mode="reapply". This is the field resolve_composition_palette_
//      action.go / resolve_composition_typography_action.go already read
//      (see extractPaletteSignal / extractTypographySignal) — writing here
//      needs NO change to either resolver.
//   3. Queues needs_composition (site-design-planner) — composition install
//      stays owned by site-design-planner (the "Choice B" precedent this
//      platform already settled once: a composition-adjacent mechanism
//      consults the resolvers, it does not install anything itself).
//
// Layout is different: resolve_composition_layout does NOT consult
// design_intent at all (it matches layouts by tag/scheme, never reads a
// site spec). The kit's layout_id/header/footer are read directly by
// resolve_composition_layout_action.go and install_site_composition_
// action.go via loadSiteThemeKitDefaults (theme_kit_defaults.go) — a plain
// siteID-keyed DB lookup, not something threaded through collected_data, so
// it works however the resolver was invoked (not just via this action).
//
// mode:
//   "fill_gaps" (default) — never overwrites anything the site already has.
//   "reapply"              — overwrites design_intent's reference_values too;
//                             the downstream needs_composition item still
//                             needs its own allow_reinstall approval before
//                             install_site_composition will replace a live
//                             composition (unchanged existing guard).
//
// A kit with needs_review=true or is_active=false is refused outright — not
// selectable until reviewed/published.
//
// Registration (add to registry.go):
//   "apply_theme_kit": {
//       Handler:     ApplyThemeKitAction,
//       Category:    "site",
//       Description: "Materialize a theme_kits row's defaults into one site",
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

var ApplyThemeKitInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "theme_kit"},
	Optional: []string{"mode"},
	Defaults: map[string]interface{}{
		"mode": "fill_gaps",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("apply_theme_kit", ApplyThemeKitInputSpec)
}

// ApplyThemeKitAction is the workflow / admin entry point.
func ApplyThemeKitAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "apply_theme_kit"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, ApplyThemeKitInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	kitRef := inputs.Get("theme_kit")
	if kitRef == "" {
		return nil, fmt.Errorf("theme_kit is required (name or id)")
	}
	mode := inputs.Get("mode")
	if mode == "" {
		mode = "fill_gaps"
	}
	if mode != "fill_gaps" && mode != "reapply" {
		return nil, fmt.Errorf("mode must be %q or %q, got %q", "fill_gaps", "reapply", mode)
	}
	reapply := mode == "reapply"

	// ── 1. Load the kit ──
	var kitID uuid.UUID
	var kitName string
	var paletteID, typoID sql.NullString
	var isActive, needsReview bool
	err = params.DB.QueryRowContext(ctx, `
		SELECT id, name, palette_id::text, typography_set_id::text, is_active, needs_review
		FROM theme_kits WHERE name = $1 OR id::text = $1
	`, kitRef).Scan(&kitID, &kitName, &paletteID, &typoID, &isActive, &needsReview)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no theme_kit matches %q", kitRef)
		}
		return nil, fmt.Errorf("load theme_kit %q: %w", kitRef, err)
	}
	if !isActive {
		return nil, fmt.Errorf("theme_kit %q is not active", kitName)
	}
	if needsReview {
		return nil, fmt.Errorf("theme_kit %q still needs_review=true — not selectable until reviewed", kitName)
	}

	applied := map[string]interface{}{}
	skipped := map[string]interface{}{}

	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// ── 2. design_intent: merge the kit's resolved palette/typography values ──
	if paletteID.Valid || typoID.Valid {
		if err := mergeThemeKitDesignIntent(ctx, tx, siteID, paletteID, typoID, reapply, applied, skipped); err != nil {
			return nil, fmt.Errorf("merge design_intent: %w", err)
		}
	}

	// ── 3. theme_kit_adoption: lineage record, superseded-then-inserted whole ──
	adoptionData := map[string]interface{}{
		"theme_kit_id":   kitID.String(),
		"theme_kit_name": kitName,
		"mode":           mode,
		"applied":        applied,
		"skipped":        skipped,
	}
	if err := supersedeAndInsertSpecWhole(ctx, tx, siteID, "theme_kit_adoption", adoptionData, "apply_theme_kit", "apply_theme_kit"); err != nil {
		return nil, fmt.Errorf("write theme_kit_adoption spec: %w", err)
	}

	// ── 4. Queue needs_composition — site-design-planner owns the actual install ──
	batchID := uuid.New()
	compSpec, _ := json.Marshal(map[string]interface{}{
		"reason":          "theme_kit_apply",
		"theme_kit_id":    kitID.String(),
		"theme_kit_name":  kitName,
		"allow_reinstall": reapply,
	})
	compInserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:             siteID,
		source:             "planner",
		pipeline:           "build",
		itemType:           "needs_composition",
		severity:           "high",
		summary:            fmt.Sprintf("Resolve composition for theme kit %q", kitName),
		spec:               string(compSpec),
		priority:           7,
		handlerAgent:       "site-design-planner",
		status:             "triaged",
		createdBy:          "apply_theme_kit",
		itemKey:            "needs_composition",
		batchID:            batchID,
		recurrenceExpected: true,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("queue needs_composition: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("ApplyThemeKitAction: applied",
		zap.String("site_id", siteID.String()),
		zap.String("theme_kit", kitName),
		zap.String("mode", mode),
		zap.Bool("composition_queued", compInserted),
	)

	return map[string]interface{}{
		"applied_theme_kit":  kitName,
		"theme_kit_id":       kitID.String(),
		"mode":               mode,
		"applied":            applied,
		"skipped":            skipped,
		"composition_queued": compInserted,
	}, nil
}

// mergeThemeKitDesignIntent writes the kit's resolved palette/typography
// colours into design_intent.palette.reference_values /
// .typography.reference_values — the exact shape
// resolve_composition_palette_action.go / resolve_composition_typography_
// action.go already read via extractReferenceValuesFromSpec. Only writes a
// dimension the site does not already have, unless reapply=true.
func mergeThemeKitDesignIntent(
	ctx context.Context, tx *sql.Tx, siteID uuid.UUID,
	paletteID, typoID sql.NullString, reapply bool,
	applied, skipped map[string]interface{},
) error {
	var currentID *uuid.UUID
	var currentJSON []byte
	err := tx.QueryRowContext(ctx, `
		SELECT id, data FROM site_specs
		WHERE site_id = $1 AND aspect = 'design_intent' AND is_current = true
	`, siteID).Scan(&currentID, &currentJSON)
	var current map[string]interface{}
	switch {
	case err == sql.ErrNoRows:
		current = map[string]interface{}{}
	case err != nil:
		return fmt.Errorf("read current design_intent: %w", err)
	default:
		current = map[string]interface{}{}
		if jerr := json.Unmarshal(currentJSON, &current); jerr != nil {
			return fmt.Errorf("unmarshal design_intent: %w", jerr)
		}
	}

	hasPalette := referenceValuesPresent(current, "palette")
	hasTypo := referenceValuesPresent(current, "typography")
	changed := false

	if paletteID.Valid && (!hasPalette || reapply) {
		var coloursJSON []byte
		if err := tx.QueryRowContext(ctx, `SELECT colours FROM palettes WHERE id = $1`, paletteID.String).Scan(&coloursJSON); err != nil {
			return fmt.Errorf("load kit palette colours: %w", err)
		}
		var colours map[string]interface{}
		if err := json.Unmarshal(coloursJSON, &colours); err != nil {
			return fmt.Errorf("unmarshal kit palette colours: %w", err)
		}
		setReferenceValues(current, "palette", colours)
		applied["palette"] = true
		changed = true
	} else if paletteID.Valid {
		skipped["palette"] = "site already has design_intent.palette.reference_values"
	}

	if typoID.Valid && (!hasTypo || reapply) {
		var fontsJSON []byte
		if err := tx.QueryRowContext(ctx, `SELECT fonts FROM typography_sets WHERE id = $1`, typoID.String).Scan(&fontsJSON); err != nil {
			return fmt.Errorf("load kit typography fonts: %w", err)
		}
		var fonts map[string]interface{}
		if err := json.Unmarshal(fontsJSON, &fonts); err != nil {
			return fmt.Errorf("unmarshal kit typography fonts: %w", err)
		}
		setReferenceValues(current, "typography", fonts)
		applied["typography"] = true
		changed = true
	} else if typoID.Valid {
		skipped["typography"] = "site already has design_intent.typography.reference_values"
	}

	if !changed {
		return nil
	}

	mergedJSON, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("marshal design_intent: %w", err)
	}
	if currentID != nil {
		if _, err := tx.ExecContext(ctx, `
			UPDATE site_specs SET is_current = false, superseded_at = now() WHERE id = $1
		`, *currentID); err != nil {
			return fmt.Errorf("supersede design_intent: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, source_agent, notes, is_current, created_by)
		VALUES ($1, 'design_intent', $2::jsonb, 'apply_theme_kit', 'apply_theme_kit',
		        'theme-kit default — see theme_kit_adoption spec for lineage', true, 'apply_theme_kit')
	`, siteID, string(mergedJSON)); err != nil {
		return fmt.Errorf("insert design_intent: %w", err)
	}
	return nil
}

func referenceValuesPresent(data map[string]interface{}, dimension string) bool {
	dim, ok := data[dimension].(map[string]interface{})
	if !ok {
		return false
	}
	ref, ok := dim["reference_values"].(map[string]interface{})
	return ok && len(ref) > 0
}

func setReferenceValues(data map[string]interface{}, dimension string, values map[string]interface{}) {
	dim, ok := data[dimension].(map[string]interface{})
	if !ok {
		dim = map[string]interface{}{}
	}
	dim["reference_values"] = values
	data[dimension] = dim
}

// supersedeAndInsertSpecWhole is a plain (non-merging) supersede-then-insert
// for a spec aspect that should be REPLACED wholesale on each write, not
// deep-merged — a lineage/point-in-time record (theme_kit_adoption), unlike
// an accumulating spec (WriteSiteSpecAction's deep-merge semantics, which
// would be wrong here: each apply is a fresh fact, not an addition to the
// last one).
func supersedeAndInsertSpecWhole(
	ctx context.Context, tx *sql.Tx, siteID uuid.UUID, aspect string,
	data map[string]interface{}, source, createdBy string,
) error {
	var oldID *uuid.UUID
	if err := tx.QueryRowContext(ctx, `
		SELECT id FROM site_specs WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect).Scan(&oldID); err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("read current %s spec: %w", aspect, err)
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s spec: %w", aspect, err)
	}
	if oldID != nil {
		if _, err := tx.ExecContext(ctx, `
			UPDATE site_specs SET is_current = false, superseded_at = now() WHERE id = $1
		`, *oldID); err != nil {
			return fmt.Errorf("supersede %s spec: %w", aspect, err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current)
		VALUES ($1, $2, $3::jsonb, $4, $5, true)
	`, siteID, aspect, string(dataJSON), source, createdBy); err != nil {
		return fmt.Errorf("insert %s spec: %w", aspect, err)
	}
	return nil
}
