// FILE: platform/orchestration/actions/install_site_composition_action.go
//
// InstallSiteCompositionAction is site-design-planner's final step. It ties
// the three resolved references (palette_id, layout_id, typography_set_id)
// into a css_themes row, creates a style_collections row pointing at it,
// updates sites.style_collection_id, AND writes the resolved_composition
// site_specs row — all in a single transaction.
//
// Why the spec write is inline (not a separate write_site_spec step):
//   resolved_composition IS the install record. If they commit separately,
//   we can end up with a live composition and no spec row explaining it
//   (or vice versa). One transaction, one atomic outcome. Deep-merge
//   semantics are irrelevant here — resolved_composition is a fresh
//   record on first install, and re-resolve is deferred. write_site_spec's
//   merge logic would be dead weight for this aspect.
//
// Runs in a single transaction:
//   1. INSERT INTO css_themes with all three FKs populated
//   2. INSERT INTO style_collections pointing at the new theme
//   3. UPDATE sites SET style_collection_id = new collection (guarded)
//   4. Supersede any existing resolved_composition spec row
//   5. INSERT INTO site_specs for the new resolved_composition
//
// Idempotency, and the one flag that changes it:
//   If sites.style_collection_id is already set, this action errors out
//   rather than overwriting — UNCHANGED default behaviour.
//   `allow_reinstall: true` on the step opts into replacing it, in the same
//   transaction, so the site is never left uncomposed. Off by default because
//   the permissive branch re-points a LIVE site's stylesheet (owner ruling
//   2026-08-02, RFC_010: new authority on a shared seam ships as an opt-in
//   field with the unsafe default OFF, not as a documented contract).
//
//   Before the flag, the only repair was an operator nulling the column by
//   hand — which opens a window where the composition loader's emergency
//   fallback can deploy `standard-brochure` over a live site
//   (render_css_composition_loader.go:144-158). See bugs_open/113.
//
// Inputs:
//   - site_id                          (required)
//   - selected_palette_id              (required) — from resolve_composition_palette
//   - selected_layout_id               (required) — from resolve_composition_layout
//   - selected_typography_set_id       (required) — from resolve_composition_typography
//   - allow_reinstall                  (optional, default false) — replace an
//     existing composition instead of erroring. The displaced id comes back as
//     `previous_collection_id`; that is the rollback value and the only record
//     of it, since the UPDATE overwrites the column in place.
//
//     TWO SOURCES, checked in this order, both defaulting to false:
//       1. the step's own config        — an AGENT-DEFINITION edit, so it applies
//                                         to EVERY install this agent performs
//       2. the work item's `spec`       — PER-REQUEST; one dispatch opts in and
//                                         nobody else's behaviour changes
//     Prefer (2). (1) exists for a workflow whose every install is a re-install,
//     and there is no such workflow today. Setting (1) on site-design-planner
//     would turn re-install on fleet-wide, which is what this flag exists to
//     prevent — see council b8e341b9 round 1, and `bugs_open/113`.
//
// The resolver outputs carried in collected_data also contribute to the
// resolved_composition spec's lineage block. Read by path:
//   - composition_palette.source    (e.g. "design_reference")
//   - composition_layout.is_fallback (bool)
//   - composition_layout.reason
//   - composition_layout.candidates  (array of layout names)
//   - composition_typography.source (e.g. "design_reference")
//   - composition_typography.matched_existing (bool)
//
// Returns:
//   {
//     "css_theme_id":         "uuid-string",
//     "css_theme_name":       "theme-<slug>",
//     "style_collection_id":  "uuid-string",
//     "collection_name":      "collection-<slug>",
//     "spec_id":              "uuid-string",       // resolved_composition row
//     "installed":            true,
//     "previous_collection_id": "uuid-string|\"\"", // rollback value; "" on first install
//     "replaced_existing":      false,              // true only on an allow_reinstall swap
//     "reinstall_approved_by":  "sentinel|name|\"\"", // who approved the replace; "" on first install
//   }
//
// APPROVAL (owner ruling 2026-08-12: "yes approval needed but for now default
// that the human approves"). A replace ALWAYS records an approver. Sources, in
// order: step config `reinstall_approved_by`, then the work item spec's
// `reinstall_approved_by` or `approved_by`, then the standing grant sentinel
// `reinstallDefaultApprover`. The sentinel and a real name are deliberately
// distinguishable, so the day the default is tightened, its blast radius is a
// query rather than a guess:
//
//   SELECT result->>'reinstall_approved_by', count(*)
//     FROM site_work_items WHERE result ? 'reinstall_approved_by' GROUP BY 1;
//
// Nothing BLOCKS on approval today — that is the "default that the human
// approves" half, and tightening it is a one-line change in the resolver.
//
// Registration (add to registry.go):
//
//   "install_site_composition": {
//       Handler:     InstallSiteCompositionAction,
//       Category:    "site",
//       Description: "Install composition into css_themes + style_collections + resolved_composition spec",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/input_contracts"
	"go.uber.org/zap"
)

var InstallSiteCompositionInputSpec = datahelpers.ActionInputSpec{
	Required: []string{
		"site_id",
		"selected_palette_id",
		"selected_layout_id",
		"selected_typography_set_id",
	},
	// allow_reinstall is NOT here on purpose. ExtractActionInputs resolves
	// config strings as PATH REFERENCES into collected_data, which is right for
	// data inputs and wrong for an author's literal switch. It is read straight
	// off StepConfig.Config by GetBoolFieldLoud below.
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec(
		"install_site_composition",
		InstallSiteCompositionInputSpec,
	)
}

// InstallSiteCompositionAction is the workflow entry point.
func InstallSiteCompositionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "install_site_composition"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		InstallSiteCompositionInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	paletteID, err := uuid.Parse(inputs.Get("selected_palette_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid selected_palette_id: %w", err)
	}
	layoutID, err := uuid.Parse(inputs.Get("selected_layout_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid selected_layout_id: %w", err)
	}
	typoID, err := uuid.Parse(inputs.Get("selected_typography_set_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid selected_typography_set_id: %w", err)
	}

	// Load site record: domain and current style_collection_id (for the
	// idempotency guard). One round-trip.
	var domain string
	var existingCollectionID sql.NullString
	err = params.DB.QueryRowContext(ctx, `
		SELECT domain, style_collection_id::text
		FROM sites
		WHERE id = $1
	`, siteID).Scan(&domain, &existingCollectionID)
	if err != nil {
		return nil, fmt.Errorf("load site record: %w", err)
	}

	// Declared-bool read: a malformed declaration (a string "true", say) does
	// NOT switch the unsafe direction on — it warns and falls back to false.
	// For a flag whose permissive branch replaces a live site's stylesheet,
	// "we could not parse it" must mean "do not do it".
	//
	// TWO SOURCES, and the second is the one that makes this usable (council
	// b8e341b9, round 1, editquality: "safe but inert"). Step config alone is
	// an AGENT-DEFINITION edit, so switching it on there turns re-install on for
	// EVERY composition install fleet-wide — the exact unsafe-default-ON state
	// this flag exists to prevent. Reading it per-request as well lets ONE
	// work item opt in and changes nobody else's behaviour.
	//
	// Both default false and both go through the loud reader, so the widest
	// branch still needs a well-formed, deliberate `true` from someone.
	// `resolvedFrom` exists because a SILENT no-op and a SAFE refusal look
	// identical from outside (council b8e341b9, round 2, guardian: "ship this
	// with a log line naming which branch resolved the flag so a silent no-op is
	// diagnosable, not just safe"). GetBoolFieldLoud is deliberately quiet when
	// a key is ABSENT — correct for it, but it means "the spec never arrived"
	// and "the spec arrived without the key" are indistinguishable in the log,
	// and those two have completely different fixes.
	resolvedFrom := "default(false): no declaration in step config or work item spec"
	allowReinstall := datahelpers.GetBoolFieldLoud(
		params.StepConfig.Config, "allow_reinstall", false, logger,
		zap.String("action", "install_site_composition"),
		zap.String("source", "step_config"),
		zap.String("site_id", siteID.String()),
	)
	if allowReinstall {
		resolvedFrom = "step_config"
	} else {
		spec, whyNoSpec := requestSpecFromCollected(params.CollectedData)
		switch {
		case spec == nil:
			resolvedFrom = "default(false): " + whyNoSpec
		case datahelpers.GetBoolFieldLoud(
			spec, "allow_reinstall", false, logger,
			zap.String("action", "install_site_composition"),
			zap.String("source", "work_item_spec"),
			zap.String("site_id", siteID.String()),
		):
			allowReinstall = true
			resolvedFrom = requestSpecPath
		default:
			resolvedFrom = "default(false): " + requestSpecPath + " present but declares no true allow_reinstall"
		}
	}
	logger.Info("InstallSiteCompositionAction: allow_reinstall resolved",
		zap.Bool("allow_reinstall", allowReinstall),
		zap.String("resolved_from", resolvedFrom),
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
	)

	// Who approved the re-compose. Empty on a first install (nothing to
	// approve); set only on the replace path.
	reinstallApprover := ""

	// Idempotency guard. A site that already has a collection is only
	// re-composed when the CALLER has explicitly asked for it.
	//
	// Why this is a flag and not a doc comment (owner ruling 2026-08-02,
	// RFC_010): the permissive branch re-points a live site's stylesheet, and
	// "callers must know what they are doing" is not a control on a tree this
	// many sessions share. Default OFF keeps the historical behaviour exactly.
	//
	// Before this flag existed the only repair was an operator nulling
	// sites.style_collection_id by hand, which opens a window where the
	// composition loader's emergency fallback can deploy a `standard-brochure`
	// stylesheet over a live site (render_css_composition_loader.go:144-158).
	// That is the hazard this closes: with allow_reinstall the swap happens
	// inside one transaction and the site is never left uncomposed.
	if existingCollectionID.Valid && existingCollectionID.String != "" {
		if !allowReinstall {
			logger.Error("InstallSiteCompositionAction: site already has style_collection_id — re-resolve not requested",
				zap.String("site_id", siteID.String()),
				zap.String("domain", domain),
				zap.String("existing_collection_id", existingCollectionID.String),
				zap.String("recommendation", "set allow_reinstall=true on this step to replace the existing composition"),
			)
			return nil, fmt.Errorf(
				"site %s already has style_collection_id=%s; re-resolve not requested (set allow_reinstall=true)",
				siteID, existingCollectionID.String,
			)
		}
		// APPROVAL (owner ruling 2026-08-12). A re-compose is an approved act,
		// not merely a requested one — but the default today is that approval is
		// GRANTED, so this records an approver rather than blocking on one.
		reinstallApprover = resolveReinstallApprover(params, logger, siteID, domain)

		// Loud on the way through: this is a live site changing its whole
		// look, and it must be greppable in the logs after the fact.
		logger.Warn("InstallSiteCompositionAction: REPLACING an existing composition (allow_reinstall=true)",
			zap.String("site_id", siteID.String()),
			zap.String("domain", domain),
			zap.String("previous_collection_id", existingCollectionID.String),
			zap.String("approved_by", reinstallApprover),
			zap.Bool("approval_was_explicit", reinstallApprover != reinstallDefaultApprover),
			zap.String("rollback", "UPDATE sites SET style_collection_id='"+existingCollectionID.String+"' WHERE id='"+siteID.String()+"'"),
		)
	}

	// Load classification for category + industry_tags on the collection row.
	category, industryTags := readClassificationFromContext(ctx, params, siteID, logger)

	// Build a slug stem used for both theme and collection names.
	slug := slugifyForCompositionName(domain)
	if slug == "" {
		slug = "site-" + siteID.String()[:8]
	}
	displayName := fmt.Sprintf("Composition for %s", domain)

	// Load legacy palette/typography JSON from the linked rows, so the
	// css_themes.color_palette and .typography columns stay populated for
	// backward-compat with the pre-composition getThemeByID / HTML-assembly
	// path. Phase 7 will drop these columns; until then the renderer's
	// composition FKs win but the legacy columns remain readable.
	var legacyPaletteJSON, legacyTypoJSON []byte
	err = params.DB.QueryRowContext(ctx,
		`SELECT colours FROM palettes WHERE id = $1`, paletteID,
	).Scan(&legacyPaletteJSON)
	if err != nil {
		return nil, fmt.Errorf("load palette colours: %w", err)
	}
	err = params.DB.QueryRowContext(ctx,
		`SELECT fonts FROM typography_sets WHERE id = $1`, typoID,
	).Scan(&legacyTypoJSON)
	if err != nil {
		return nil, fmt.Errorf("load typography fonts: %w", err)
	}

	// Transaction boundary: all three writes (theme, collection, site
	// update) atomic. If any fails, none are committed.
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// ── 1. css_themes row ──
	themeName, err := resolveUniqueNameInTx(
		ctx, tx,
		"css_themes", "name",
		"theme-"+slug,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve theme name: %w", err)
	}

	// css_content is empty — the renderer reads composition via FKs.
	// css_template is empty — no {{.Primary}}-style placeholders are used
	// by the post-025 renderer. The legacy columns stay for one more phase.
	var themeID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO css_themes (
			name, display_name,
			css_content, css_template,
			color_palette, typography,
			palette_id, layout_id, typography_set_id,
			is_active, origin, needs_review,
			source_site_id, source_domain, forked_at
		) VALUES (
			$1, $2,
			'', '',
			$3::jsonb, $4::jsonb,
			$5, $6, $7,
			true, 'adopted', false,
			$8, $9, NOW()
		) RETURNING id
	`,
		themeName, displayName,
		string(legacyPaletteJSON), string(legacyTypoJSON),
		paletteID, layoutID, typoID,
		siteID, domain,
	).Scan(&themeID)
	if err != nil {
		return nil, fmt.Errorf("insert css_themes: %w", err)
	}

	// ── 2. style_collections row ──
	collectionName, err := resolveUniqueNameInTx(
		ctx, tx,
		"style_collections", "name",
		"collection-"+slug,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("resolve collection name: %w", err)
	}

	// industry_tags is a PostgreSQL text[] column — NOT jsonb. Marshalling
	// to JSON would pass bytes with the wrong shape and the INSERT fails
	// with "column is of type text[] but expression is of type jsonb".
	// datahelpers.PGTextArrayLiteral produces a PG array literal string
	// that pairs with a $N::text[] cast in the SQL.
	industryTagsLiteral := datahelpers.PGTextArrayLiteral(industryTags)

	// Note on columns: style_collections does NOT have palette_id/layout_id/
	// typography_set_id FK columns (Phase 2 migration only added those to
	// css_themes). Composition is tracked via css_theme_id which has the
	// three FKs. The renderer joins through css_themes to reach them.
	//
	// header_component_id and footer_component_id are left NULL at install
	// time by default — webdesign-agent populates these later when it
	// renders the header and footer components. The one exception: a
	// theme-kit default pin, applied here (not deferred) exactly like
	// fork_theme_from_site_action.go's own chrome pins — same
	// chromePinEligibleSQL guard, so a kit can never pin an ineligible row.
	var kitHeaderID, kitFooterID interface{}
	if kit, ok, kerr := loadSiteThemeKitDefaults(ctx, params.DB, siteID); kerr == nil && ok {
		if kit.HeaderComponentID.Valid {
			if eligible, eerr := chromeComponentEligible(ctx, tx, kit.HeaderComponentID.UUID); eerr == nil && eligible {
				kitHeaderID = kit.HeaderComponentID.UUID
			} else if eerr != nil {
				logger.Warn("InstallSiteCompositionAction: theme-kit header eligibility check failed",
					zap.Error(eerr), zap.String("theme_kit", kit.ThemeKitName))
			}
		}
		if kit.FooterComponentID.Valid {
			if eligible, eerr := chromeComponentEligible(ctx, tx, kit.FooterComponentID.UUID); eerr == nil && eligible {
				kitFooterID = kit.FooterComponentID.UUID
			} else if eerr != nil {
				logger.Warn("InstallSiteCompositionAction: theme-kit footer eligibility check failed",
					zap.Error(eerr), zap.String("theme_kit", kit.ThemeKitName))
			}
		}
		if kitHeaderID != nil || kitFooterID != nil {
			logger.Info("InstallSiteCompositionAction: pinning theme-kit chrome",
				zap.String("theme_kit", kit.ThemeKitName),
				zap.Bool("header_pinned", kitHeaderID != nil),
				zap.Bool("footer_pinned", kitFooterID != nil),
			)
		}
	}
	var collectionID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO style_collections (
			name, display_name, css_theme_id,
			header_component_id, footer_component_id,
			color_palette, typography, category, industry_tags,
			is_active, origin, needs_review,
			source_site_id, source_domain, forked_at
		) VALUES (
			$1, $2, $3,
			$10, $11,
			$4::jsonb, $5::jsonb, $6, $7::text[],
			true, 'adopted', false,
			$8, $9, NOW()
		) RETURNING id
	`,
		collectionName, displayName, themeID,
		string(legacyPaletteJSON), string(legacyTypoJSON),
		category, industryTagsLiteral,
		siteID, domain,
		kitHeaderID, kitFooterID,
	).Scan(&collectionID)
	if err != nil {
		return nil, fmt.Errorf("insert style_collections: %w", err)
	}

	// ── 3. Link to site ──
	// Guarded UPDATE — only writes if the column still holds exactly what we
	// read during the idempotency check above. Defensive against a race where
	// another path (e.g. legacy install_theme in webdesign-agent) linked or
	// re-linked a collection between that check and now.
	//
	// `IS NOT DISTINCT FROM` rather than `IS NULL` so the guard keeps working
	// in BOTH modes: NULL matches NULL on a first install, and the observed id
	// matches itself on an allow_reinstall. Weakening this to an unguarded
	// UPDATE would silently clobber a concurrent install — the reinstall path
	// needs the race check MORE than the first-install path, not less.
	result, err := tx.ExecContext(ctx, `
		UPDATE sites
		SET style_collection_id = $1,
		    updated_at = NOW()
		WHERE id = $2
		  AND style_collection_id IS NOT DISTINCT FROM $3::uuid
	`, collectionID, siteID, existingCollectionID)
	if err != nil {
		return nil, fmt.Errorf("update sites.style_collection_id: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		// Lost the race. Someone set style_collection_id between our
		// load and this UPDATE. We won't write it — the earlier install
		// wins. The theme + collection rows we already inserted will
		// roll back (we're still in the transaction — no commit yet).
		logger.Error("InstallSiteCompositionAction: sites.style_collection_id was set during install — aborting link",
			zap.String("site_id", siteID.String()),
			zap.String("domain", domain),
			zap.String("our_collection_id", collectionID.String()),
		)
		return nil, fmt.Errorf(
			"lost race on sites.style_collection_id for %s — another install ran concurrently",
			siteID,
		)
	}

	// ── 4. Write resolved_composition spec ──
	// Inline because the spec IS the install record. Splitting it into
	// a separate workflow step would break atomicity — composition could
	// exist without a spec row or vice versa.
	//
	// Preserves the site_specs history contract: any existing current row
	// for (site_id, 'resolved_composition') gets is_current=false +
	// superseded_at; new row inserted with is_current=true. Same shape as
	// WriteSiteSpecAction, minus the deep-merge (not meaningful here —
	// resolved_composition is a full-replacement spec).
	specBody, err := buildResolvedCompositionSpec(
		params.CollectedData,
		themeID, paletteID, layoutID, typoID,
	)
	if err != nil {
		return nil, fmt.Errorf("build resolved_composition spec body: %w", err)
	}
	// The names read-back to populate the spec are still the ones we just
	// inserted — attach them here so the spec is self-contained.
	specBody["css_theme_name"] = themeName
	specBody["palette_name"], err = readPaletteNameInTx(ctx, tx, paletteID)
	if err != nil {
		return nil, fmt.Errorf("read palette name: %w", err)
	}
	specBody["layout_name"], err = readLayoutNameInTx(ctx, tx, layoutID)
	if err != nil {
		return nil, fmt.Errorf("read layout name: %w", err)
	}
	specBody["typography_name"], err = readTypographyNameInTx(ctx, tx, typoID)
	if err != nil {
		return nil, fmt.Errorf("read typography name: %w", err)
	}

	specJSON, err := json.Marshal(specBody)
	if err != nil {
		return nil, fmt.Errorf("marshal resolved_composition spec: %w", err)
	}

	// Supersede any existing current row (defensive — should not exist
	// given the style_collection_id guard above, but history contract
	// requires we handle it)
	_, err = tx.ExecContext(ctx, `
		UPDATE site_specs
		SET is_current = false, superseded_at = NOW()
		WHERE site_id = $1 AND aspect = 'resolved_composition' AND is_current = true
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("supersede old resolved_composition spec: %w", err)
	}

	var specID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_specs (
			site_id, aspect, data,
			source, source_agent,
			is_current, created_by
		) VALUES (
			$1, 'resolved_composition', $2::jsonb,
			'site-design-planner', 'site-design-planner',
			true, 'site-design-planner'
		) RETURNING id
	`, siteID, string(specJSON)).Scan(&specID)
	if err != nil {
		return nil, fmt.Errorf("insert resolved_composition spec: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit install tx: %w", err)
	}

	logger.Info("InstallSiteCompositionAction: installed",
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
		zap.String("css_theme_id", themeID.String()),
		zap.String("css_theme_name", themeName),
		zap.String("style_collection_id", collectionID.String()),
		zap.String("collection_name", collectionName),
		zap.String("spec_id", specID.String()),
		zap.String("palette_id", paletteID.String()),
		zap.String("layout_id", layoutID.String()),
		zap.String("typography_set_id", typoID.String()),
	)

	return map[string]interface{}{
		"css_theme_id":        themeID.String(),
		"css_theme_name":      themeName,
		"style_collection_id": collectionID.String(),
		"collection_name":     collectionName,
		"spec_id":             specID.String(),
		"installed":           true,
		// Empty string on a first install; the replaced id on a reinstall.
		// This is the rollback value — the ONLY record of what was displaced,
		// since the UPDATE overwrites it in place.
		"previous_collection_id": existingCollectionID.String,
		"replaced_existing":      allowReinstall && existingCollectionID.Valid && existingCollectionID.String != "",
		// Who approved the replacement. "" on a first install. On a replace it
		// is either a named approver or reinstallDefaultApprover — and the two
		// are deliberately distinguishable, so a later audit can separate
		// "a human said yes" from "the standing default said yes for them".
		"reinstall_approved_by": reinstallApprover,
	}, nil
}

// buildResolvedCompositionSpec assembles the resolved_composition spec body
// from the resolver outputs in collected_data and the install's own IDs.
// Name fields are left blank here — caller fills them after we've committed
// the theme/collection rows (the names are read inside the same transaction).
//
// Sources are mapped to the enum values defined in the resolved_composition
// spec schema (see 004_spec_schemas.sql / validate_resolved_composition_spec).
func buildResolvedCompositionSpec(
	collectedData map[string]interface{},
	themeID, paletteID, layoutID, typoID uuid.UUID,
) (map[string]interface{}, error) {

	paletteSource := mapPaletteSourceToLineageEnum(
		readStringFromContext(collectedData, "composition_palette.source"),
	)
	typographySource := mapTypographySourceToLineageEnum(
		readStringFromContext(collectedData, "composition_typography.source"),
	)
	// A theme-kit layout is NOT a library tag match, and recording it as one
	// writes a false structured fact into the lineage the enum exists to make
	// queryable: nothing could then separate "the matcher scored this highest"
	// from "a human chose this kit". The free-text `reasoning` below carries
	// the truth but is not queryable, which is the whole point of the enum.
	// Requires 689_theme_kits.sql's validator widening to be LIVE first —
	// validate_resolved_composition_spec REFUSES an unknown layout_source.
	//
	// The resolver now REPORTS its source on every branch (bugs_open/445); the
	// is_fallback inference below is kept only for replaying older
	// collected_data shapes that carry no `source` key. Trust the reported
	// value first — an inference cannot represent a scheme gap or a weak fit,
	// and recorded both as a clean `library_match`.
	layoutSource := "library_match"
	switch reported := readStringFromContext(collectedData, "composition_layout.source"); reported {
	case "theme_kit_default", "library_fallback", "library_match", "mission_hint":
		layoutSource = reported
	default:
		if readBoolFromContext(collectedData, "composition_layout.is_fallback") {
			layoutSource = "library_fallback"
		}
	}

	layoutReason := readStringFromContext(collectedData, "composition_layout.reason")
	paletteRationale := readStringFromContext(collectedData, "composition_palette.source")
	typoRationale := readStringFromContext(collectedData, "composition_typography.source")

	lineage := map[string]interface{}{
		"palette_source":    paletteSource,
		"layout_source":     layoutSource,
		"typography_source": typographySource,
	}
	if cands := readStringSliceFromContext(collectedData, "composition_layout.candidates"); len(cands) > 0 {
		lineage["layout_candidates"] = cands
	}

	// --- fit evidence (bugs_open/445) ---
	//
	// `layout_match_score` is not a new invention: migration 103 specified it in
	// April 2026 as "(float 0-1) — tag-overlap score for chosen layout", and it
	// was never computed. Measured 2026-09-03, 0 of 33 current
	// resolved_composition rows carry the key, and the only trace of a score was
	// a prose sentence in `reasoning` that had to be parsed with a regex to be
	// read at all. This writes it, plus the context needed to act on it.
	//
	// The validator (689_theme_kits.sql) is permissive on unknown keys, so these
	// need no migration — but `layout_source` IS enum-checked, and
	// 'needs_new_layout_candidate' is already among its allowed values.
	if coverage, ok := readFloatFromContext(collectedData, "composition_layout.tag_coverage"); ok {
		lineage["layout_match_score"] = coverage

		fit := map[string]interface{}{
			"tag_coverage":    coverage,
			"matched_terms":   readStringSliceFromContext(collectedData, "composition_layout.matched_terms"),
			"unmatched_terms": readStringSliceFromContext(collectedData, "composition_layout.unmatched_terms"),
			"runner_up":       readStringFromContext(collectedData, "composition_layout.runner_up"),
		}
		// The threshold in force is recorded ALONGSIDE the score, so that
		// changing the threshold later cannot silently re-interpret rows written
		// under the old one. A bare score plus a constant that moved is how a
		// historical comparison quietly stops meaning anything.
		if thr, ok := readFloatFromContext(collectedData, "composition_layout.fit_threshold"); ok {
			fit["threshold"] = thr
		}
		if s, ok := readFloatFromContext(collectedData, "composition_layout.score"); ok {
			fit["score"] = s
		}
		if ts, ok := readFloatFromContext(collectedData, "composition_layout.tag_score"); ok {
			fit["tag_score"] = ts
		}
		if m, ok := readFloatFromContext(collectedData, "composition_layout.margin"); ok {
			fit["margin"] = m
		}
		if rus, ok := readFloatFromContext(collectedData, "composition_layout.runner_up_score"); ok {
			fit["runner_up_score"] = rus
		}
		lineage["layout_fit"] = fit
	}

	// A scheme gap used to be invisible in the structured record: the resolver
	// computed is_scheme_mismatch, nothing persisted it, and the row read as a
	// clean `library_match`. Recorded as its own key rather than folded into the
	// enum, so one field carries one fact.
	if readBoolFromContext(collectedData, "composition_layout.is_scheme_mismatch") {
		lineage["layout_scheme_mismatch"] = true
	}
	if readBoolFromContext(collectedData, "composition_layout.library_gap") {
		lineage["layout_gap_flagged"] = true
		if gr := readStringFromContext(collectedData, "composition_layout.gap_reason"); gr != "" {
			lineage["layout_gap_reason"] = gr
		}
		// Promote the enum only for a gap on a real match. A hard fallback keeps
		// `library_fallback`, which says what was actually applied; the gap flag
		// above is what marks it for review either way.
		if layoutSource == "library_match" {
			lineage["layout_source"] = "needs_new_layout_candidate"
		}
	}

	reasoning := layoutReason
	if reasoning == "" {
		reasoning = fmt.Sprintf(
			"composition resolved: palette=%s, layout=%s, typography=%s",
			paletteRationale, layoutSource, typoRationale,
		)
	}

	return map[string]interface{}{
		"css_theme_id":      themeID.String(),
		"palette_id":        paletteID.String(),
		"layout_id":         layoutID.String(),
		"typography_set_id": typoID.String(),

		// *_name fields filled by caller after read-back inside the transaction
		"css_theme_name":  "", // overwritten
		"palette_name":    "", // overwritten
		"layout_name":     "", // overwritten
		"typography_name": "", // overwritten

		"lineage":     lineage,
		"reasoning":   reasoning,
		"resolved_by": "site-design-planner",
		"resolved_at": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// mapPaletteSourceToLineageEnum maps the resolver's source string to the
// enum values the resolved_composition spec schema requires
// (fingerprint | library_reuse | mission_hint | design_intent_values |
// archetype_default). Unknown values fall through to archetype_default.
func mapPaletteSourceToLineageEnum(src string) string {
	switch src {
	case "design_reference":
		return "fingerprint"
	case "mission_hint":
		return "mission_hint"
	case "design_intent":
		return "design_intent_values"
	case "layout_library_inherit", "fallback_default":
		return "archetype_default"
	default:
		return "archetype_default"
	}
}

// mapTypographySourceToLineageEnum maps the resolver's source string to the
// enum values the resolved_composition spec schema requires
// (fingerprint_font_family_match | archetype_default | layout_default |
// mission_hint | fallback_sans_modern).
func mapTypographySourceToLineageEnum(src string) string {
	switch src {
	case "design_reference", "design_intent":
		return "fingerprint_font_family_match"
	case "mission_hint":
		return "mission_hint"
	case "fallback_sans_modern":
		return "fallback_sans_modern"
	default:
		return "fallback_sans_modern"
	}
}

// readStringFromContext / readBoolFromContext / readStringSliceFromContext
// are tiny convenience wrappers around ExtractNestedField. Kept private to
// this file because they're only useful for reading already-known resolver
// output paths.

func readStringFromContext(data map[string]interface{}, path string) string {
	return datahelpers.ExtractNestedFieldString(data, path)
}

func readBoolFromContext(data map[string]interface{}, path string) bool {
	raw := datahelpers.ExtractNestedField(data, path)
	if b, ok := raw.(bool); ok {
		return b
	}
	return false
}

// readFloatFromContext returns (value, true) only when the path actually holds
// a number. The `ok` is load-bearing: an ABSENT fit and a measured-zero fit
// must not both land in the lineage as 0.0, because a reader cannot then tell
// "this layout matched nothing" from "this row predates fit evidence".
func readFloatFromContext(data map[string]interface{}, path string) (float64, bool) {
	switch v := datahelpers.ExtractNestedField(data, path).(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case json.Number:
		f, err := v.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

func readStringSliceFromContext(data map[string]interface{}, path string) []string {
	raw := datahelpers.ExtractNestedField(data, path)
	out := []string{}
	switch v := raw.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
	case []string:
		out = append(out, v...)
	}
	return out
}

// Three small helpers to read back just-inserted row names from within the
// transaction. Kept here because they're install-specific — the install
// action is the only caller that needs all three names in one place.

func readPaletteNameInTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (string, error) {
	var name string
	err := tx.QueryRowContext(ctx, `SELECT name FROM palettes WHERE id = $1`, id).Scan(&name)
	return name, err
}

func readLayoutNameInTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (string, error) {
	var name string
	err := tx.QueryRowContext(ctx, `SELECT name FROM layouts WHERE id = $1`, id).Scan(&name)
	return name, err
}

func readTypographyNameInTx(ctx context.Context, tx *sql.Tx, id uuid.UUID) (string, error) {
	var name string
	err := tx.QueryRowContext(ctx, `SELECT name FROM typography_sets WHERE id = $1`, id).Scan(&name)
	return name, err
}

// requestSpecPath is the ONE collected_data path the dispatching work item's
// `spec` arrives at. Named so the "we did not find it" log line can quote the
// path it looked at, rather than leaving a reader to grep for it.
const requestSpecPath = "input_data.spec"

// requestSpecFromCollected returns the dispatching work item's `spec` object
// from collected_data, or (nil, reason) naming why there isn't one.
//
// IT RETURNS THE REASON, NOT JUST THE ABSENCE (council b8e341b9 round 3,
// editquality): "nothing at that path" and "something at that path that is not
// an object" are different failures with different fixes — a dispatch/shape
// problem versus a malformed spec from whatever queued the item — and an earlier
// draft collapsed both into one nil, so the caller reported "no input_data.spec
// in collected_data" for a spec that was demonstrably present. A diagnostic that
// lies about the rarer case is worse than none, because it sends the next reader
// to the wrong half of the system.
//
// Why this exists rather than a path in the step's config: the config map is
// resolved as PATH REFERENCES into collected_data, so an author cannot put a
// literal switch there without it being read as a lookup. The work item's spec
// is the per-request channel, and it is the only one a single dispatch can set.
//
// THE SHAPE IS MEASURED, NOT GUESSED (2026-08-12). Round 2 of council
// b8e341b9 objected — five seats, on the documented envelope-unwrap fault line —
// that this was a shape guess that could silently never populate. Over 30 days
// of orchestration_states carrying input_data: `input_data.spec` is present on
// 2,363 runs and `input_data.body.spec` on ZERO. Two live
// needs_composition → site-design-planner dispatches carry the work item's spec
// VERBATIM under input_data.spec — null-valued keys included, which is what
// shows the dispatcher passes the spec whole rather than projecting a subset.
// The `body.spec` branch this function used to carry was dormant machinery and
// has been removed; if a third shape ever appears the action refuses (safe) AND
// the caller's `resolved_from` log line names the path it looked at, so the
// diagnosis is one grep rather than a code read.
//
// Uses input_contracts.GetValueAtExactPath — the platform's existing exact-path
// reader — rather than datahelpers.ExtractNestedField, which auto-unwraps
// through a `.response` envelope. That auto-unwrap is right for reading data and
// wrong here: this is an AUTHORITY switch, and it must not be satisfiable by a
// `true` that arrived inside some other agent's reply.
func requestSpecFromCollected(collected map[string]interface{}) (map[string]interface{}, string) {
	value, found := input_contracts.GetValueAtExactPath(collected, requestSpecPath)
	if !found {
		return nil, "no " + requestSpecPath + " in collected_data"
	}
	spec, isMap := value.(map[string]interface{})
	if !isMap {
		// %T only — the TYPE, never the value. What arrived here is
		// caller-supplied and can carry anything.
		return nil, fmt.Sprintf("%s is present but is %T, not an object", requestSpecPath, value)
	}
	return spec, ""
}

// reinstallDefaultApprover is what a re-compose records when nobody named an
// approver. It is a SENTINEL, not a person, and it is written into the result
// so an audit can tell a real approval from an inherited one.
//
// OWNER RULING 2026-08-12: "yes approval needed but for now default that the
// human approves." So approval is a first-class field on the replace path from
// today, and the default is GRANT. Tightening later is a one-line change here —
// return "" and have the caller refuse — and every call site already records
// which of the two it got, so the blast radius of that flip is measurable
// BEFORE it is made:
//
//	SELECT result->>'reinstall_approved_by', count(*)
//	  FROM site_work_items WHERE result ? 'reinstall_approved_by' GROUP BY 1;
//
// Do NOT reword this constant casually — it is a stored value, so a changed
// string silently splits that GROUP BY into two populations.
const reinstallDefaultApprover = "default-grant/owner-2026-08-12"

// resolveReinstallApprover names who approved replacing a live site's
// composition. Checked in order, first non-empty wins:
//
//  1. step config      `reinstall_approved_by`  — an agent-definition edit
//  2. work item spec   `reinstall_approved_by`  — per-request, the usual one
//  3. work item spec   `approved_by`            — the generic field a HITL
//     approval flow already writes
//  4. reinstallDefaultApprover                  — the standing grant
//
// (3) exists so that when a real approval queue is wired, it needs no change
// here: `site_work_items.approved_by` is the column that flow already fills.
func resolveReinstallApprover(params ActionParams, logger *zap.Logger, siteID uuid.UUID, domain string) string {
	if v := datahelpers.GetStringField(params.StepConfig.Config, "reinstall_approved_by", ""); v != "" {
		return v
	}
	if spec, _ := requestSpecFromCollected(params.CollectedData); spec != nil {
		for _, key := range []string{"reinstall_approved_by", "approved_by"} {
			if v := datahelpers.GetStringField(spec, key, ""); v != "" {
				return v
			}
		}
	}
	logger.Warn("InstallSiteCompositionAction: re-compose approved by the STANDING DEFAULT, not by a named approver",
		zap.String("site_id", siteID.String()),
		zap.String("domain", domain),
		zap.String("approved_by", reinstallDefaultApprover),
		zap.String("owner_ruling", "2026-08-12: approval required, default GRANTED"),
		zap.String("to_name_one", "set reinstall_approved_by (or approved_by) in the work item spec"),
	)
	return reinstallDefaultApprover
}
