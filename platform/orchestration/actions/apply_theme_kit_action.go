// FILE: platform/orchestration/actions/apply_theme_kit_action.go
//
// ApplyThemeKitAction materializes a theme_kits row's defaults into ONE
// site's own rows. It never creates a live binding — see the theme-kit
// plan, session "themes" (docs026 register entry: design-composition.md,
// theme_kits vs css_themes).
//
// What it writes, each independently, each a DEFAULT not a constraint:
//   1. site_specs aspect 'theme_kit_adoption' — lineage record (always).
//   2. site_specs aspect 'design_intent' — writes the kit's resolved
//      palette/typography colours into .palette.reference_values /
//      .typography.reference_values. In the DEFAULT mode ("start") this
//      SUPERSEDES what is already there; see the mode block below, and note
//      that this comment used to say "ONLY if the site has none there yet",
//      which was the pre-ruling behaviour and the opposite of what ships.
//      This is the field resolve_composition_palette_action.go /
//      resolve_composition_typography_action.go already read (see
//      extractPaletteSignal / extractTypographySignal) — writing here needs
//      NO change to either resolver.
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
//   "start" (DEFAULT)      — WRITES the kit's palette/typography, superseding
//                             what is there. Per OWNER RULING 2026-09-02:
//                             "by default it can start with a theme and change
//                             it if it wishes, but it must have full authority
//                             to ignore our set of themes if it chooses." It
//                             first shipped defaulting to "fill_gaps", which
//                             was a no-op on the 33 of 57 sites the classifier
//                             had already touched — a theme that never started
//                             anything.
//   "fill_gaps"            — never overwrites anything the site already has.
//   "reapply"              — like "start", and also replaces an INSTALLED
//                             composition; the downstream needs_composition
//                             item still needs its own allow_reinstall
//                             approval before install_site_composition will
//                             replace a live composition (unchanged guard).
//
// Written values are marked reference_source: "theme_kit:<name>" and
// reference_is_default: true, so a later reader can tell a kit's default from
// a decision. The ONE thing no mode overwrites is design_intent.<dim>.locked
// (see designIntentLocked) — a deliberate human pin nothing sets automatically.
//
// ⚠ ORDERING HAZARD — A KIT APPLIED BEFORE CLASSIFICATION LOSES PALETTE AND
// TYPOGRAPHY, SILENTLY. There is no guard here, by omission not by design.
// On the FRESH path (082 with no --from) domain-research-classifier writes
// design_intent AFTER this action runs, and write_site_spec supersedes the
// current row after a deep merge in which SCALAR KEYS ARE OVERWRITTEN BY THE
// INCOMING VALUE. So the classifier discards the kit's reference_values.
// Measured on gamedesign.uk: a manual design_intent at 17:04:35 with pinned=t
// was is_current=f by 17:11:32, carrying a different hex.
//   · layout SURVIVES — it is read from aspect 'theme_kit_adoption', which the
//     classifier does not write.
//   · palette is discarded, but that is moot for APPEARANCE: no design_intent
//     palette reaches the 8 core slots anyway (bugs_open/438 §6a-ter).
//   · TYPOGRAPHY is discarded AND typography is the dimension that renders —
//     this is the one that actually costs something.
//   · `locked: true` does NOT protect you. It is read when THIS action writes;
//     nothing makes the classifier respect it. Worse, the key survives the
//     merge while the values do not, so the row ends up claiming a human pin
//     over a classifier's values.
// So a kit works on an ALREADY-CLASSIFIED site and is defeated on a new one —
// the inverse of the ruling's framing. Found by the council gate, correlation
// bed139b2-f512-436a-9ba8-ff2fbfade8ef round 2; candidates and why none is
// applied here: bugs_open/438 §6d.
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
	// OWNER RULING 2026-09-02: "by default it can start with a theme and change
	// it if it wishes, but it must have full authority to ignore our set of
	// themes if it chooses." So a kit is the STARTING POINT, not a deferential
	// filler — the default mode WRITES the kit's values. Whatever runs later
	// (the classifier, the design overlay, a human) is free to supersede them,
	// and nothing here freezes anything.
	//
	//   start      (default) — write the kit's palette/typography, superseding
	//                          what is there. This is what makes a theme the
	//                          starting point rather than a no-op on any site
	//                          the classifier has already touched.
	//   fill_gaps            — conservative: write only dimensions the site has
	//                          nothing for. Was the default until the ruling
	//                          above; kept because "top up what's missing" is a
	//                          real, if narrower, thing to want.
	//   reapply              — start, AND replace an installed composition
	//                          (carries allow_reinstall downstream).
	//
	// The ONE thing no mode overwrites is an explicit human lock (see
	// designIntentLocked) — a person saying "these values, deliberately" is not
	// a default to be started from.
	mode := inputs.Get("mode")
	if mode == "" {
		mode = "start"
	}
	if mode != "start" && mode != "fill_gaps" && mode != "reapply" {
		return nil, fmt.Errorf("mode must be one of %q, %q, %q — got %q", "start", "fill_gaps", "reapply", mode)
	}
	reapply := mode == "reapply"
	writeOverExisting := mode == "start" || mode == "reapply"

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
		if err := mergeThemeKitDesignIntent(ctx, tx, siteID, kitName, paletteID, typoID, writeOverExisting, applied, skipped); err != nil {
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
	//
	// But only when it can actually succeed. install_site_composition REFUSES
	// (hard error) whenever sites.style_collection_id is already set and
	// allow_reinstall is false — so on an already-composed site, fill_gaps used
	// to queue an item that was GUARANTEED to fail, and still report
	// composition_queued:true. That left a failed work item, an adoption record,
	// and a suppressed generic_theme check, for a site whose layout and chrome
	// never changed. Now: composed + fill_gaps means we say so and queue nothing.
	var existingCollection sql.NullString
	if err := tx.QueryRowContext(ctx,
		`SELECT style_collection_id::text FROM sites WHERE id = $1`, siteID,
	).Scan(&existingCollection); err != nil {
		return nil, fmt.Errorf("read style_collection_id: %w", err)
	}
	alreadyComposed := existingCollection.Valid && existingCollection.String != ""

	compInserted := false
	compQueueNote := ""
	switch {
	case alreadyComposed && !reapply:
		compQueueNote = "site already has a composition; fill_gaps will not replace it — re-run with mode=reapply to change layout/chrome"
		skipped["composition"] = compQueueNote
	default:
		batchID := uuid.New()
		compSpec, _ := json.Marshal(map[string]interface{}{
			"reason":          "theme_kit_apply",
			"theme_kit_id":    kitID.String(),
			"theme_kit_name":  kitName,
			"allow_reinstall": reapply,
		})
		compInserted, err = insertWorkItem(ctx, tx, workItem{
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
		if !compInserted {
			// idx_swi_dedup dropped it: an open needs_composition already
			// exists for this site, and it carries the PREVIOUS spec — so this
			// apply's allow_reinstall/theme_kit_id never reach the planner.
			// Silent-false would read as "queued fine" to a caller checking
			// only the error.
			compQueueNote = "an open needs_composition item already exists for this site — this apply's spec (including allow_reinstall) did NOT replace it; resolve or cancel that item first"
			skipped["composition"] = compQueueNote
			logger.Warn("ApplyThemeKitAction: needs_composition dropped on conflict with an open item",
				zap.String("site_id", siteID.String()),
				zap.String("theme_kit", kitName),
				zap.Bool("reapply", reapply),
			)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	logger.Info("ApplyThemeKitAction: applied",
		zap.String("site_id", siteID.String()),
		zap.String("theme_kit", kitName),
		zap.String("mode", mode),
		zap.Bool("composition_queued", compInserted),
		zap.Int("dimensions_applied", len(applied)),
		zap.Int("dimensions_skipped", len(skipped)),
	)

	out := map[string]interface{}{
		"applied_theme_kit":  kitName,
		"theme_kit_id":       kitID.String(),
		"mode":               mode,
		"applied":            applied,
		"skipped":            skipped,
		"composition_queued": compInserted,
		// An apply that changed NOTHING is a real outcome and must be visible
		// to a caller that only checks for an error. On an already-composed,
		// already-classified site (the common case) fill_gaps legitimately has
		// nothing to do — saying so is the difference between a no-op and a
		// no-op that looks like success.
		"changed_anything": len(applied) > 0 || compInserted,
	}
	if compQueueNote != "" {
		out["composition_note"] = compQueueNote
	}
	return out, nil
}

// mergeThemeKitDesignIntent writes the kit's resolved palette/typography
// colours into design_intent.palette.reference_values /
// .typography.reference_values — the exact shape
// resolve_composition_palette_action.go / resolve_composition_typography_
// action.go already read via extractReferenceValuesFromSpec. Only writes a
// dimension the site does not already have, unless reapply=true.
func mergeThemeKitDesignIntent(
	ctx context.Context, tx *sql.Tx, siteID uuid.UUID, kitName string,
	paletteID, typoID sql.NullString, writeOverExisting bool,
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

	// Presence MUST be asked the same way the READER asks it. The resolvers use
	// extractReferenceValuesFromSpec (resolve_composition_helpers.go), which
	// accepts BOTH a nested {palette:{reference_values:{…}}} and a FLAT
	// {palette:{primary:…}} shape, preferring nested when both exist. An
	// independent nested-only check here read a flat-shaped site as "has none",
	// wrote the kit's values nested, and the resolver then preferred the kit's —
	// silently overwriting the site's own palette, which is the exact guarantee
	// fill_gaps exists to keep. Latent when found ([MEASURED 2026-09-02]: 0 flat
	// palette rows, 1 flat typography row) — fixed before it wasn't.
	hasPalette := len(extractReferenceValuesFromSpec(current, "palette")) > 0
	hasTypo := len(extractReferenceValuesFromSpec(current, "typography")) > 0

	// A human's explicit mission hint outranks a kit default. For PALETTE the
	// cascade already guarantees that (mission is rung 1, design_intent rung 2
	// — resolve_composition_pallette_action.go), so writing design_intent
	// cannot displace it. TYPOGRAPHY is the reverse: design_intent is rung 1
	// and mission rung 2 (resolve_composition_typography_action.go:12-14), an
	// existing asymmetry in this codebase — so writing the kit's fonts into
	// design_intent WOULD silently outrank a human's mission.preferred_typography.
	// Treat that as locked so the kit stays a default, not an override.
	typoMissionLocked := missionPrefersTypography(ctx, tx, siteID)
	if typoMissionLocked {
		skipped["typography_mission_hint"] = "site has mission.preferred_typography — a human's explicit choice outranks a kit default"
	}

	// A deliberate human lock is the ONE thing no mode overwrites. A
	// classifier's guess is not a lock: per the owner's 2026-09-02 ruling the
	// theme is the starting point and the machine may change it afterwards, so
	// starting from the theme is not overriding anyone.
	paletteLocked := designIntentLocked(current, "palette")
	typoLocked := designIntentLocked(current, "typography") || typoMissionLocked
	if paletteLocked {
		skipped["palette_locked"] = "design_intent.palette.locked is true — a deliberate human pin is never overwritten by a kit"
	}
	if designIntentLocked(current, "typography") {
		skipped["typography_locked"] = "design_intent.typography.locked is true — a deliberate human pin is never overwritten by a kit"
	}

	writePalette := paletteID.Valid && !paletteLocked && (writeOverExisting || !hasPalette)
	writeTypo := typoID.Valid && !typoLocked && (writeOverExisting || !hasTypo)

	changed := false

	if writePalette {
		var coloursJSON []byte
		if err := tx.QueryRowContext(ctx, `SELECT colours FROM palettes WHERE id = $1`, paletteID.String).Scan(&coloursJSON); err != nil {
			return fmt.Errorf("load kit palette colours: %w", err)
		}
		var colours map[string]interface{}
		if err := json.Unmarshal(coloursJSON, &colours); err != nil {
			return fmt.Errorf("unmarshal kit palette colours: %w", err)
		}
		setReferenceValues(current, "palette", colours)
		markThemeKitStartingPoint(current, "palette", kitName)
		applied["palette"] = map[string]interface{}{"replaced_existing": hasPalette}
		changed = true
	} else if paletteID.Valid && !paletteLocked {
		skipped["palette"] = "site already has design_intent.palette.reference_values and mode=fill_gaps was requested"
	}

	if writeTypo {
		var fontsJSON []byte
		if err := tx.QueryRowContext(ctx, `SELECT fonts FROM typography_sets WHERE id = $1`, typoID.String).Scan(&fontsJSON); err != nil {
			return fmt.Errorf("load kit typography fonts: %w", err)
		}
		var fonts map[string]interface{}
		if err := json.Unmarshal(fontsJSON, &fonts); err != nil {
			return fmt.Errorf("unmarshal kit typography fonts: %w", err)
		}
		setReferenceValues(current, "typography", fonts)
		markThemeKitStartingPoint(current, "typography", kitName)
		applied["typography"] = map[string]interface{}{"replaced_existing": hasTypo}
		changed = true
	} else if typoID.Valid && !typoLocked {
		skipped["typography"] = "site already has design_intent.typography.reference_values and mode=fill_gaps was requested"
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

// missionPrefersTypography reports whether the site's mission spec carries an
// explicit preferred_typography. Best-effort: a read failure returns false
// (the kit then applies), because refusing to theme a site over an unreadable
// mission spec is the worse of the two errors.
func missionPrefersTypography(ctx context.Context, tx *sql.Tx, siteID uuid.UUID) bool {
	var dataJSON []byte
	err := tx.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'mission' AND is_current = true
	`, siteID).Scan(&dataJSON)
	if err != nil {
		return false
	}
	var data map[string]interface{}
	if json.Unmarshal(dataJSON, &data) != nil {
		return false
	}
	switch v := data["preferred_typography"].(type) {
	case map[string]interface{}:
		return len(v) > 0
	case string:
		return v != ""
	default:
		return false
	}
}

// designIntentLocked reports a DELIBERATE human pin on one dimension —
// `design_intent.<dimension>.locked: true`. It is the only thing a theme kit
// will not write over.
//
// Why an in-data key rather than site_specs.pinned: `pinned` is a per-ROW flag
// and design_intent is superseded-then-inserted on every write, with neither
// WriteSiteSpecAction nor this action carrying `pinned` forward — so a pin set
// that way survives exactly until the next write of any kind (2 of the 4 rows
// ever pinned are already superseded). A key inside `data` rides the document.
//
// Nothing writes this key automatically, and that is the point: under the
// owner's 2026-09-02 ruling a classifier's palette is a starting guess the
// machine may revise, NOT a pin. Only a person marking a value deliberate
// makes it one.
func designIntentLocked(data map[string]interface{}, dimension string) bool {
	dim, ok := data[dimension].(map[string]interface{})
	if !ok {
		return false
	}
	locked, _ := dim["locked"].(bool)
	return locked
}

// markThemeKitStartingPoint records WHERE these values came from and, by
// saying so, that they are a default rather than a decision. Two readers
// benefit: a human asking "is this site on a kit's palette or its own?" (which
// was otherwise unanswerable — a kit-written design_intent is byte-identical
// in shape to a classifier-written one), and anything downstream that wants to
// know it is free to override. Deep-merge carries unknown keys forward, so it
// survives later writes until something replaces the dimension wholesale.
func markThemeKitStartingPoint(data map[string]interface{}, dimension, kitName string) {
	dim, ok := data[dimension].(map[string]interface{})
	if !ok {
		return
	}
	dim["reference_source"] = "theme_kit:" + kitName
	dim["reference_is_default"] = true
	data[dimension] = dim
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
