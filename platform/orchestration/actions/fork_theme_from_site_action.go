package actions

// FILE: platform/orchestration/actions/fork_theme_from_site_action.go
//
// ForkThemeFromSiteAction forks an adopted site's design into the
// reusable theme library using the post-025 composable model.
//
// Rewritten in Phase 5 of 025_palette_layout_typography_migration.
// The pre-025 version produced a single css_themes row carrying a
// monolithic css_template with {{.Primary}}-style placeholders. The
// new renderer ignores css_themes.css_template; it resolves palette,
// layout, and typography via FK columns on css_themes.
//
// What changed vs pre-025:
//   - Inside the transaction, calls three helpers BEFORE the
//     css_themes INSERT:
//       createPaletteForFork        → new palettes row
//       resolveTypographySetForFork → matched or new typography_sets row
//       resolveLayoutByTags        → matched layout from existing library
//   - The resulting three FK ids are written into css_themes:
//     palette_id, layout_id, typography_set_id are populated.
//   - Legacy columns (css_content, css_template, color_palette,
//     typography) stay populated for backward compat with
//     getThemeByID / HTML-assembly path until Phase 7 drops them.
//   - TemplateCSSFromSpec is no longer called. The new renderer doesn't
//     read css_themes.css_template; producing templated CSS with
//     legacy {{.Primary}} placeholders would be busy-work that nobody
//     reads. An empty string is written to that column instead. The
//     css_templating.go helper file stays compile-clean as an inert
//     neighbour until Phase 7 removes it.
//   - The HITL work item spec gains a layout_resolution block with the
//     reason and candidate layout list so the reviewer can override
//     the automated layout pick.
//
// What stayed the same:
//   - Action name, config keys (design_spec_field, rendered_css_field,
//     current_collection_id_field)
//   - Return shape: { forked: bool, theme_id, collection_id, ... }
//   - Lineage resolution (forked_from_*_id fields)
//   - Classification loading (loadClassificationForFork)
//   - Name collision handling
//   - Non-fatal skip behaviour — fork never aborts parent workflow

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ForkThemeFromSiteInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{
		"domain",
		"design_spec_field",
		"rendered_css_field",
		"current_collection_id_field",
	},
	Defaults: map[string]interface{}{
		"design_spec_field":           "design_spec.result",
		"rendered_css_field":          "generated_css.result",
		"current_collection_id_field": "site_context.style_collection_id",
	},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("fork_theme_from_site", ForkThemeFromSiteInputSpec)
}

// ForkThemeFromSiteAction forks an adopted site's composition into the
// library. Producing granular rows (palette + typography + theme) lets
// later sites reuse the typography stack independently of the palette,
// and lets a HITL reviewer see which layout was picked and why.
func ForkThemeFromSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "fork_theme_from_site"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		logger.Warn("fork_theme_from_site: no database connection, skipping")
		return forkSkipped("no database connection"), nil
	}

	config := params.StepConfig.Config

	// ── Resolve inputs ──
	siteIDStr := resolveConfigString(config, "site_id", params.CollectedData, logger)
	if siteIDStr == "" {
		siteIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		logger.Warn("fork_theme_from_site: invalid site_id, skipping", zap.Error(err))
		return forkSkipped("invalid site_id"), nil
	}

	domain := resolveConfigString(config, "domain", params.CollectedData, logger)
	if domain == "" {
		domain = datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.domain")
	}
	if domain == "" {
		logger.Warn("fork_theme_from_site: domain missing, skipping")
		return forkSkipped("domain not found in collected data"), nil
	}

	designSpecField := datahelpers.GetStringField(config, "design_spec_field", "design_spec.result")
	renderedCSSField := datahelpers.GetStringField(config, "rendered_css_field", "generated_css.result")
	currentCollectionField := datahelpers.GetStringField(config, "current_collection_id_field", "site_context.style_collection_id")

	designSpec, _ := datahelpers.ExtractNestedField(params.CollectedData, designSpecField).(map[string]interface{})
	if designSpec == nil {
		logger.Warn("fork_theme_from_site: design_spec not found, skipping",
			zap.String("field", designSpecField))
		return forkSkipped("design_spec not in collected data"), nil
	}

	renderedCSS, _ := datahelpers.ExtractNestedField(params.CollectedData, renderedCSSField).(string)
	if renderedCSS == "" {
		logger.Warn("fork_theme_from_site: rendered_css not found, skipping",
			zap.String("field", renderedCSSField))
		return forkSkipped("rendered_css not in collected data"), nil
	}

	// ── Lineage: find the site's current style_collection and theme ──
	// Feeds forked_from_collection_id and forked_from_theme_id.
	var (
		parentCollectionID *uuid.UUID
		parentThemeID      *uuid.UUID
	)
	parentCollectionStr := datahelpers.ExtractNestedFieldString(params.CollectedData, currentCollectionField)
	if parentCollectionStr == "" {
		var collID sql.NullString
		err := params.DB.QueryRowContext(ctx,
			`SELECT style_collection_id::text FROM sites WHERE id = $1`,
			siteID).Scan(&collID)
		if err == nil && collID.Valid {
			parentCollectionStr = collID.String
		}
	}
	if parentCollectionStr != "" {
		if cid, err := uuid.Parse(parentCollectionStr); err == nil {
			parentCollectionID = &cid
			var tid sql.NullString
			err := params.DB.QueryRowContext(ctx,
				`SELECT css_theme_id::text FROM style_collections WHERE id = $1`,
				cid).Scan(&tid)
			if err == nil && tid.Valid {
				if parsed, err := uuid.Parse(tid.String); err == nil {
					parentThemeID = &parsed
				}
			}
		}
	}

	// Also look up the parent theme's palette_id so the new palette
	// can record forked_from_palette_id lineage. Post-025 themes have
	// palette_id populated; pre-025 themes in the library may not.
	var parentPaletteID *uuid.UUID
	if parentThemeID != nil {
		var pid sql.NullString
		err := params.DB.QueryRowContext(ctx,
			`SELECT palette_id::text FROM css_themes WHERE id = $1`,
			*parentThemeID,
		).Scan(&pid)
		if err == nil && pid.Valid {
			if parsed, err := uuid.Parse(pid.String); err == nil {
				parentPaletteID = &parsed
			}
		}
	}

	// ── Classification: used for collection category and industry_tags ──
	category, industryTags := loadClassificationForFork(ctx, params.DB, siteID, logger)

	// ── Extract palette + typography from design_spec for legacy columns ──
	// These are still written to css_themes for backward-compat with
	// getThemeByID and the HTML-assembly path. The authoritative copies
	// now live in the palettes and typography_sets tables.
	legacyPalette := extractStringMap(designSpec, "color_scheme")
	legacyTypography := extractStringMap(designSpec, "typography")

	// ── Name resolution with collision handling ──
	baseName := "adopted-" + domainSlug(domain)
	themeName, collectionName := resolveForkNames(ctx, params.DB, baseName, logger)

	// ── Origin ──
	origin := "adopted"
	if parentThemeID != nil {
		origin = "fork_of_adopted"
	}

	// ── Transaction: palette + typography + layout + themes + collections + work item ──
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("fork_theme_from_site: failed to begin tx, skipping", zap.Error(err))
		return forkSkipped("transaction begin failed"), nil
	}
	defer tx.Rollback() // safe: no-op after commit

	displayName := "Adopted from " + domain
	forkedAt := time.Now()

	// ── 1. Create palette row ──
	newPaletteID, err := createPaletteForFork(
		ctx, tx,
		baseName, displayName, category, industryTags,
		designSpec, siteID, domain, parentPaletteID,
		logger,
	)
	if err != nil {
		logger.Warn("fork_theme_from_site: palette fork failed, skipping", zap.Error(err))
		return forkSkipped("palette fork failed: " + err.Error()), nil
	}

	// ── 2. Match or create typography_set ──
	newTypoID, typoMatched, err := resolveTypographySetForFork(
		ctx, tx,
		baseName, displayName, category, industryTags,
		designSpec, siteID, domain,
		logger,
	)
	if err != nil {
		logger.Warn("fork_theme_from_site: typography_set resolution failed, skipping", zap.Error(err))
		return forkSkipped("typography_set resolution failed: " + err.Error()), nil
	}

	// ── 3. Resolve layout (match from library, or fall back) ──
	layoutRes, err := resolveLayoutByTags(
		ctx, tx,
		category, industryTags,
		logger,
	)
	if err != nil {
		logger.Warn("fork_theme_from_site: layout resolution failed, skipping", zap.Error(err))
		return forkSkipped("layout resolution failed: " + err.Error()), nil
	}

	// ── 4. Insert theme ──
	// Legacy columns (css_content, css_template, color_palette, typography)
	// are still populated for backward-compat with getThemeByID and
	// the HTML-assembly path. The new renderer reads from the three
	// FKs instead. Phase 7 drops the legacy columns.
	//
	// css_template gets an empty string: the new renderer ignores it,
	// and calling TemplateCSSFromSpec to produce {{.Primary}}-style
	// placeholders nobody reads would be wasted work.
	var newThemeID uuid.UUID
	legacyPaletteJSON, _ := json.Marshal(legacyPalette)
	legacyTypoJSON, _ := json.Marshal(legacyTypography)

	err = tx.QueryRowContext(ctx, `
		INSERT INTO css_themes (
			name, display_name,
			css_content, css_template,
			color_palette, typography,
			palette_id, layout_id, typography_set_id,
			is_active, origin, needs_review,
			forked_from_theme_id, source_site_id, source_domain, forked_at
		) VALUES (
			$1, $2,
			$3, '',
			$4::jsonb, $5::jsonb,
			$6, $7, $8,
			true, $9, true,
			$10, $11, $12, $13
		) RETURNING id
	`,
		themeName, displayName,
		renderedCSS,
		string(legacyPaletteJSON), string(legacyTypoJSON),
		newPaletteID, layoutRes.LayoutID, newTypoID,
		origin,
		parentThemeID, siteID, domain, forkedAt,
	).Scan(&newThemeID)
	if err != nil {
		logger.Warn("fork_theme_from_site: theme insert failed, skipping", zap.Error(err))
		return forkSkipped("theme insert failed: " + err.Error()), nil
	}

	// ── 5. Inherit parent collection's header/footer if any ──
	var headerComponentID, footerComponentID *uuid.UUID
	if parentCollectionID != nil {
		var hID, fID sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT header_component_id::text, footer_component_id::text
			FROM style_collections WHERE id = $1
		`, *parentCollectionID).Scan(&hID, &fID)
		if err == nil {
			if hID.Valid {
				if parsed, err := uuid.Parse(hID.String); err == nil {
					headerComponentID = &parsed
				}
			}
			if fID.Valid {
				if parsed, err := uuid.Parse(fID.String); err == nil {
					footerComponentID = &parsed
				}
			}
		}
	}

	// ── 6. Insert collection ──
	var newCollectionID uuid.UUID
	industryTagsJSON, _ := json.Marshal(industryTags)
	err = tx.QueryRowContext(ctx, `
		INSERT INTO style_collections (
			name, display_name, css_theme_id,
			header_component_id, footer_component_id,
			color_palette, typography, category, industry_tags,
			is_active, origin, needs_review,
			forked_from_collection_id, source_site_id, source_domain, forked_at
		) VALUES (
			$1, $2, $3,
			$4, $5,
			$6::jsonb, $7::jsonb, $8, $9::jsonb,
			true, $10, true,
			$11, $12, $13, $14
		) RETURNING id
	`, collectionName, displayName, newThemeID,
		headerComponentID, footerComponentID,
		string(legacyPaletteJSON), string(legacyTypoJSON),
		category, string(industryTagsJSON),
		origin,
		parentCollectionID, siteID, domain, forkedAt,
	).Scan(&newCollectionID)
	if err != nil {
		logger.Warn("fork_theme_from_site: collection insert failed, skipping", zap.Error(err))
		return forkSkipped("collection insert failed: " + err.Error()), nil
	}

	// ── 7. Insert HITL work item ──
	// Spec now includes layout_resolution so a reviewer can see why
	// that layout was picked and which alternatives matched. On
	// approve, the on_approve update clears needs_review. A reviewer
	// who wants to change the layout edits the theme row directly
	// before approving — no separate workflow action needed for that.
	spec := map[string]interface{}{
		"theme_id":                    newThemeID.String(),
		"collection_id":               newCollectionID.String(),
		"theme_name":                  themeName,
		"collection_name":             collectionName,
		"source_domain":               domain,
		"source_site_id":              siteID.String(),
		"forked_from_theme_id":        uuidPtrString(parentThemeID),
		"palette_id":                  newPaletteID.String(),
		"palette":                     legacyPalette,
		"typography_set_id":           newTypoID.String(),
		"typography":                  legacyTypography,
		"typography_matched_existing": typoMatched,
		"layout_resolution": map[string]interface{}{
			"layout_id":   layoutRes.LayoutID.String(),
			"layout_name": layoutRes.LayoutName,
			"reason":      layoutRes.Reason,
			"candidates":  layoutRes.Candidates,
		},
		"preview_url": "https://" + domain,
		"on_approve": map[string]interface{}{
			"update_theme":      map[string]interface{}{"needs_review": false},
			"update_collection": map[string]interface{}{"needs_review": false},
			"update_palette":    map[string]interface{}{"needs_review": false},
		},
		"on_reject": map[string]interface{}{
			"update_theme":      map[string]interface{}{"is_active": false, "needs_review": false},
			"update_collection": map[string]interface{}{"is_active": false, "needs_review": false},
			"update_palette":    map[string]interface{}{"is_active": false, "needs_review": false},
		},
	}
	// If the typography_set was freshly created (not a library match),
	// include it in the on_approve/on_reject update set so it also
	// clears needs_review.
	if !typoMatched {
		if onApprove, ok := spec["on_approve"].(map[string]interface{}); ok {
			onApprove["update_typography_set"] = map[string]interface{}{"needs_review": false}
		}
		if onReject, ok := spec["on_reject"].(map[string]interface{}); ok {
			onReject["update_typography_set"] = map[string]interface{}{"is_active": false, "needs_review": false}
		}
	}

	specJSON, _ := json.Marshal(spec)

	workItemKey := fmt.Sprintf("theme_review:%s", newThemeID.String())
	summary := fmt.Sprintf("Review adopted theme from %s (layout: %s)", domain, layoutRes.LayoutName)

	var workItemID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by, item_key
		) VALUES (
			$1, 'theme-fork', 'design', 'needs_theme_review', 'medium',
			$2, $3::jsonb, 60, 'theme-review-handler', 'needs_human_review',
			'fork_theme_from_site', $4
		) ON CONFLICT DO NOTHING
		RETURNING id
	`, siteID, summary, string(specJSON), workItemKey).Scan(&workItemID)
	if err != nil && err != sql.ErrNoRows {
		logger.Warn("fork_theme_from_site: work item insert failed, skipping", zap.Error(err))
		return forkSkipped("work item insert failed: " + err.Error()), nil
	}

	if err := tx.Commit(); err != nil {
		logger.Warn("fork_theme_from_site: commit failed, skipping", zap.Error(err))
		return forkSkipped("commit failed: " + err.Error()), nil
	}

	logger.Info("fork_theme_from_site: forked theme into library (composable)",
		zap.String("theme_name", themeName),
		zap.String("theme_id", newThemeID.String()),
		zap.String("palette_id", newPaletteID.String()),
		zap.String("typography_set_id", newTypoID.String()),
		zap.Bool("typography_matched_existing", typoMatched),
		zap.String("layout_id", layoutRes.LayoutID.String()),
		zap.String("layout_name", layoutRes.LayoutName),
		zap.String("layout_resolution_reason", layoutRes.Reason),
		zap.String("collection_id", newCollectionID.String()),
		zap.String("work_item_id", workItemID.String()),
		zap.String("origin", origin),
		zap.String("source_domain", domain),
	)

	return map[string]interface{}{
		"forked":            true,
		"theme_id":          newThemeID.String(),
		"theme_name":        themeName,
		"palette_id":        newPaletteID.String(),
		"typography_set_id": newTypoID.String(),
		"layout_id":         layoutRes.LayoutID.String(),
		"layout_name":       layoutRes.LayoutName,
		"collection_id":     newCollectionID.String(),
		"collection_name":   collectionName,
		"work_item_id":      workItemID.String(),
		"origin":            origin,
	}, nil
}

// forkSkipped returns a structured non-error result indicating the fork
// was skipped for the given reason. Used so the parent workflow can
// branch on `forked == true` without treating a skip as a failure.
func forkSkipped(reason string) map[string]interface{} {
	return map[string]interface{}{
		"forked": false,
		"reason": reason,
	}
}

// resolveForkNames resolves the theme and collection names, appending
// a short timestamp suffix if the base name already exists in
// css_themes OR style_collections. Unchanged from pre-025.
//
// Note: palette and typography_set name resolution happens inside the
// transaction via resolveUniqueNameInTx (in fork_theme_composition.go)
// — those names don't collide with theme/collection namespaces since
// they live in different tables.
func resolveForkNames(ctx context.Context, db *sql.DB, baseName string, logger *zap.Logger) (themeName, collectionName string) {
	themeName = baseName
	collectionName = baseName

	var themeExists, collectionExists bool
	_ = db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM css_themes WHERE name = $1)`, baseName).Scan(&themeExists)
	_ = db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM style_collections WHERE name = $1)`, baseName).Scan(&collectionExists)

	if themeExists || collectionExists {
		suffix := time.Now().UTC().Format("-20060102-1504")
		themeName = baseName + suffix
		collectionName = baseName + suffix
		logger.Info("fork_theme_from_site: name collision, suffixing",
			zap.String("base", baseName),
			zap.String("resolved", themeName))
	}
	return themeName, collectionName
}

// loadClassificationForFork reads the site's classification spec for
// category and industry_tags. Best-effort: returns sensible defaults
// if the spec isn't present or doesn't contain the fields.
// Unchanged from pre-025.
func loadClassificationForFork(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (category string, industryTags []string) {
	category = "general"
	industryTags = []string{}

	var dataJSON sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT data::text
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
		ORDER BY created_at DESC
		LIMIT 1
	`, siteID).Scan(&dataJSON)
	if err != nil || !dataJSON.Valid {
		return category, industryTags
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(dataJSON.String), &parsed); err != nil {
		return category, industryTags
	}

	if ind, ok := parsed["industry"].(string); ok && ind != "" {
		industryTags = append(industryTags, strings.ToLower(strings.TrimSpace(ind)))
	}
	if sub, ok := parsed["sub_industry"].(string); ok && sub != "" {
		industryTags = append(industryTags, strings.ToLower(strings.TrimSpace(sub)))
	}
	if bst, ok := parsed["site_type"].(string); ok && bst != "" {
		category = bst
	}

	return category, industryTags
}

// extractStringMap returns a map[string]string from a nested map[string]interface{}
// under the given key. Non-string values are skipped silently.
// Unchanged from pre-025.
func extractStringMap(parent map[string]interface{}, key string) map[string]string {
	out := make(map[string]string)
	nested, ok := parent[key].(map[string]interface{})
	if !ok {
		return out
	}
	for k, v := range nested {
		if s, ok := v.(string); ok && s != "" {
			out[k] = s
		}
	}
	return out
}

func uuidPtrString(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}
