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
// Idempotency:
//   If sites.style_collection_id is already set, this action errors out
//   rather than overwriting. Re-resolution (changing a site's composition
//   after first install) is a deliberate future feature behind HITL.
//
// Inputs:
//   - site_id                          (required)
//   - selected_palette_id              (required) — from resolve_composition_palette
//   - selected_layout_id               (required) — from resolve_composition_layout
//   - selected_typography_set_id       (required) — from resolve_composition_typography
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
//   }
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
	"go.uber.org/zap"
)

var InstallSiteCompositionInputSpec = datahelpers.ActionInputSpec{
	Required: []string{
		"site_id",
		"selected_palette_id",
		"selected_layout_id",
		"selected_typography_set_id",
	},
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

	// Idempotency guard: re-resolution is a future feature. If a site
	// already has a collection, loud-fail rather than silently duplicate.
	if existingCollectionID.Valid && existingCollectionID.String != "" {
		logger.Error("InstallSiteCompositionAction: site already has style_collection_id — re-resolve is not supported",
			zap.String("site_id", siteID.String()),
			zap.String("domain", domain),
			zap.String("existing_collection_id", existingCollectionID.String),
			zap.String("recommendation", "clear sites.style_collection_id manually to force re-install"),
		)
		return nil, fmt.Errorf(
			"site %s already has style_collection_id=%s; re-resolve not supported",
			siteID, existingCollectionID.String,
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
	// time — webdesign-agent populates these later when it renders the
	// header and footer components.
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
			NULL, NULL,
			$4::jsonb, $5::jsonb, $6, $7::text[],
			true, 'adopted', false,
			$8, $9, NOW()
		) RETURNING id
	`,
		collectionName, displayName, themeID,
		string(legacyPaletteJSON), string(legacyTypoJSON),
		category, industryTagsLiteral,
		siteID, domain,
	).Scan(&collectionID)
	if err != nil {
		return nil, fmt.Errorf("insert style_collections: %w", err)
	}

	// ── 3. Link to site ──
	// Guarded UPDATE — only writes if still NULL. Defensive against a race
	// where another path (e.g. legacy install_theme in webdesign-agent)
	// linked a collection between our idempotency check and now.
	result, err := tx.ExecContext(ctx, `
		UPDATE sites
		SET style_collection_id = $1,
		    updated_at = NOW()
		WHERE id = $2
		  AND style_collection_id IS NULL
	`, collectionID, siteID)
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
	layoutSource := "library_match"
	if readBoolFromContext(collectedData, "composition_layout.is_fallback") {
		layoutSource = "library_fallback"
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
