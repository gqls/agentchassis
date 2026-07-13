// FILE: platform/orchestration/actions/fork_theme_from_site_action.go
//
// ForkThemeFromSiteAction persists a site's generated CSS as a css_themes +
// style_collections pair in the library, pending human review.
//
// Single mode: library contribution.
//   Called when should_fork_theme is true on the agent's input (currently
//   never routinely triggered — the `fork_theme` step in webdesign-agent is
//   the only caller and it runs only if input_data.should_fork_theme == true).
//   Inserts theme + collection with needs_review = true, creates a
//   needs_theme_review HITL work item. The site's own style_collection_id
//   is NOT modified. Failure is non-fatal and returns `{forked: false,
//   reason: ...}`.
//
// Post-merge (2026-04-19): the `install_on_site` mode is gone. Composition
// installation is owned exclusively by `site-design-planner` via the
// `install_site_composition` action (see 026). Any caller that wants to
// install a composition onto a site must go through site-design-planner
// (queue a `needs_composition` work item). Removed code:
//   - `install_on_site` flag on InputSpec
//   - "install mode" branch that UPDATE'd sites.style_collection_id
//   - Conditional needs_review flag (was `!installOnSite`; now hardcoded true)
//
// Either way:
//   1. Reads design_spec, rendered_css from collected data
//   2. Loads the site's current style_collection (for lineage, if any)
//   3. Produces a templated css_template from the rendered CSS
//   4. Inserts css_themes + style_collections rows (one transaction)
//   5. Inserts a needs_theme_review work item
//
// Registration: unchanged. Still registered as "fork_theme_from_site".
//   "fork_theme_from_site": {
//       Handler:     ForkThemeFromSiteAction,
//       Category:    "site",
//       Description: "Fork an adopted site's generated theme into the reusable library",
//       IsLocal:     true,
//   }

package actions

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
	Deprecated: map[string]string{
		// Kept here so any stale workflow config referencing this key
		// gets a clear log line rather than silently being ignored.
		"install_on_site": "Removed 2026-04-19. Composition install moved to site-design-planner " +
			"(install_site_composition action). Queue a needs_composition work item instead.",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("fork_theme_from_site", ForkThemeFromSiteInputSpec)
}

func ForkThemeFromSiteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "fork_theme_from_site"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		// Non-fatal: return a result indicating skip.
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

	// Warn if stale workflow config still references the removed flag — helps
	// surface any agent_definitions rows that weren't updated during the merge.
	if _, hasStaleFlag := config["install_on_site"]; hasStaleFlag {
		logger.Error("fork_theme_from_site: workflow config references removed 'install_on_site' flag " +
			"— site-design-planner owns composition install now; remove this key from the step config")
	}

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
	// These feed forked_from_collection_id and forked_from_theme_id.
	// If the site has no collection assigned, both remain nil and
	// origin stays "adopted" (not "fork_of_adopted").
	var (
		parentCollectionID *uuid.UUID
		parentThemeID      *uuid.UUID
	)
	parentCollectionStr := datahelpers.ExtractNestedFieldString(params.CollectedData, currentCollectionField)
	if parentCollectionStr == "" {
		// Fall back to DB lookup against sites.style_collection_id.
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
			// Look up the theme the parent collection points at.
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

	// ── Classification: used for collection category and industry_tags ──
	category, industryTags := loadClassificationForFork(ctx, params.DB, siteID, logger)

	// ── Extract palette + typography from design_spec for the theme row ──
	palette := extractStringMap(designSpec, "color_scheme")
	typography := extractStringMap(designSpec, "typography")

	// ── Produce templated CSS ──
	templatedCSS := TemplateCSSFromSpec(renderedCSS, designSpec)

	// ── Name resolution with collision handling ──
	baseName := "adopted-" + domainSlug(domain)
	themeName, collectionName := resolveForkNames(ctx, params.DB, baseName, logger)

	// ── Origin ──
	origin := "adopted"
	if parentThemeID != nil {
		origin = "fork_of_adopted"
	}

	// Library-contribution mode is the only mode now. Every fork needs
	// human review before it joins the usable library.
	needsReview := true

	// ── Transaction: themes + collections + HITL review item ──
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("fork_theme_from_site: failed to begin tx, skipping", zap.Error(err))
		return forkSkipped("transaction begin failed"), nil
	}
	defer tx.Rollback() // safe: no-op after commit

	// Insert theme
	var newThemeID uuid.UUID
	paletteJSON, _ := json.Marshal(palette)
	typographyJSON, _ := json.Marshal(typography)
	displayName := "Adopted from " + domain
	forkedAt := time.Now()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO css_themes (
			name, display_name, css_content, css_template,
			color_palette, typography, is_active,
			origin, needs_review,
			forked_from_theme_id, source_site_id, source_domain, forked_at
		) VALUES (
			$1, $2, $3, $4,
			$5::jsonb, $6::jsonb, true,
			$7, $12,
			$8, $9, $10, $11
		) RETURNING id
	`, themeName, displayName, renderedCSS, templatedCSS,
		string(paletteJSON), string(typographyJSON),
		origin,
		parentThemeID, siteID, domain, forkedAt,
		needsReview,
	).Scan(&newThemeID)
	if err != nil {
		logger.Warn("fork_theme_from_site: theme insert failed, skipping", zap.Error(err))
		return forkSkipped("theme insert failed: " + err.Error()), nil
	}

	// Look up parent collection's header/footer if we have one, otherwise null
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

	// Insert collection
	var newCollectionID uuid.UUID
	// industry_tags is text[] not jsonb; see nullable_helpers.go for the helper.
	industryTagsLiteral := datahelpers.PGTextArrayLiteral(industryTags)
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
			$6::jsonb, $7::jsonb, $8, $9::text[],
			true, $10, $15,
			$11, $12, $13, $14
		) RETURNING id
	`, collectionName, displayName, newThemeID,
		headerComponentID, footerComponentID,
		string(paletteJSON), string(typographyJSON), category, industryTagsLiteral,
		origin,
		parentCollectionID, siteID, domain, forkedAt,
		needsReview,
	).Scan(&newCollectionID)
	if err != nil {
		// logger.Error (not Warn) because fork_theme silently absorbs the error
		// via forkSkipped — without an error-level log we have no signal in
		// production that this path is broken. See the industry_tags text[]
		// type mismatch that hid here for months.
		logger.Error("fork_theme_from_site: collection insert failed, skipping fork",
			zap.Error(err),
			zap.String("collection_name", collectionName),
			zap.String("domain", domain),
		)
		return forkSkipped("collection insert failed: " + err.Error()), nil
	}

	// Library-contribution mode: create HITL review work item. Inside the
	// same transaction so theme + collection + item all commit or all roll back.
	var workItemID uuid.UUID
	spec := map[string]interface{}{
		"theme_id":             newThemeID.String(),
		"collection_id":        newCollectionID.String(),
		"theme_name":           themeName,
		"collection_name":      collectionName,
		"source_domain":        domain,
		"source_site_id":       siteID.String(),
		"forked_from_theme_id": uuidPtrString(parentThemeID),
		"palette":              palette,
		"typography":           typography,
		"preview_url":          "https://" + domain,
		"on_approve": map[string]interface{}{
			"update_theme":      map[string]interface{}{"needs_review": false},
			"update_collection": map[string]interface{}{"needs_review": false},
		},
		"on_reject": map[string]interface{}{
			"update_theme":      map[string]interface{}{"is_active": false, "needs_review": false},
			"update_collection": map[string]interface{}{"is_active": false, "needs_review": false},
		},
	}
	specJSON, _ := json.Marshal(spec)

	workItemKey := fmt.Sprintf("theme_review:%s", newThemeID.String())
	summary := fmt.Sprintf("Review adopted theme from %s", domain)

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

	logger.Info("fork_theme_from_site: persisted theme for library review",
		zap.String("theme_name", themeName),
		zap.String("theme_id", newThemeID.String()),
		zap.String("collection_id", newCollectionID.String()),
		zap.String("work_item_id", workItemID.String()),
		zap.String("origin", origin),
		zap.String("source_domain", domain),
	)

	return map[string]interface{}{
		"forked":          true,
		"theme_id":        newThemeID.String(),
		"theme_name":      themeName,
		"collection_id":   newCollectionID.String(),
		"collection_name": collectionName,
		"work_item_id":    workItemID.String(),
		"origin":          origin,
	}, nil
}

// forkSkipped returns a structured non-error result indicating the fork
// was skipped for the given reason. Used so the parent workflow can branch
// on `forked == true` without treating a skip as a failure.
func forkSkipped(reason string) map[string]interface{} {
	return map[string]interface{}{
		"forked": false,
		"reason": reason,
	}
}

// resolveForkNames resolves the theme and collection names, appending a short
// timestamp suffix if the base name already exists in css_themes OR style_collections.
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
		industryTags = append(industryTags, strings.ToLower(ind))
	}
	if sub, ok := parsed["sub_industry"].(string); ok && sub != "" {
		industryTags = append(industryTags, strings.ToLower(sub))
	}
	if bst, ok := parsed["site_type"].(string); ok && bst != "" {
		category = bst
	}

	return category, industryTags
}

// extractStringMap returns a map[string]string from a nested map[string]interface{}
// under the given key. Non-string values are skipped silently.
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
