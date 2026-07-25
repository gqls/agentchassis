// FILE: platform/orchestration/actions/v3_site_actions.go
// Additional actions needed for the v3 multipage website builder component-based architecture.
// These complement existing actions in site_db_actions.go and component_library.go.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// SelectStyleCollectionAction selects a style collection for a site
// Either by explicit ID, by site_id lookup, or by domain/industry matching
// Config:
//   - style_collection_id: explicit UUID (optional)
//   - site_id_field: path to site_id in collected_data (optional)
//   - domain_field: path to domain for industry matching (optional)
func SelectStyleCollectionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SelectStyleCollectionAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	if params.DB == nil {
		params.Logger.Warn("SelectStyleCollectionAction: No database, returning default style")
		return getDefaultStyleCollection(), nil
	}

	// Resolve site_id early — needed for persist step at the end
	var siteID uuid.UUID
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = parsed
			}
		}
	}

	// Helper: persist and return
	persistAndReturn := func(coll *StyleCollection, source string) (interface{}, error) {
		params.Logger.Info("SelectStyleCollectionAction: Resolved style",
			zap.String("name", coll.Name),
			zap.String("source", source),
			zap.String("id", coll.ID.String()))

		// Persist style_collection_id to sites table so downstream agents
		// (webdesign-agent, maintenance agents) can find it via DB lookup
		if siteID != uuid.Nil {
			query := `UPDATE sites SET style_collection_id = $2, updated_at = NOW() WHERE id = $1`
			if err := execDB(ctx, params.DB, query, siteID, coll.ID); err != nil {
				params.Logger.Warn("SelectStyleCollectionAction: Failed to persist style_collection_id",
					zap.Error(err))
				// Non-fatal — continue with the result
			} else {
				params.Logger.Info("SelectStyleCollectionAction: Persisted style_collection_id to sites table",
					zap.String("site_id", siteID.String()),
					zap.String("style_collection_id", coll.ID.String()))
			}
		}

		return styleCollectionToResult(coll), nil
	}

	// Priority 1: Explicit style_collection_id in config
	if scID, ok := config["style_collection_id"].(string); ok && scID != "" {
		scUUID, err := uuid.Parse(scID)
		if err == nil {
			coll, err := getStyleCollectionByID(ctx, params.DB, scUUID, params.Logger)
			if err == nil {
				return persistAndReturn(coll, "explicit_id")
			}
		}
	}

	// Priority 2 (NEW): Planner's style choice via style_from config
	// Config example: "style_from": "site_plan.style_collection"
	// The planner writes a style name (e.g. "professional-dark") to that path
	if styleFromField, ok := config["style_from"].(string); ok && styleFromField != "" {
		styleName := datahelpers.ExtractNestedFieldString(params.CollectedData, styleFromField)
		if styleName != "" {
			params.Logger.Info("SelectStyleCollectionAction: Trying planner style choice",
				zap.String("style_from", styleFromField),
				zap.String("style_name", styleName))
			coll, err := GetStyleCollectionByName(ctx, params.DB, styleName, params.Logger)
			if err == nil && coll != nil {
				return persistAndReturn(coll, "planner_style_from")
			}
			params.Logger.Warn("SelectStyleCollectionAction: Planner style not found in DB",
				zap.String("style_name", styleName))
		}
	}

	// Priority 3: Look up by site_id (existing style_collection_id on sites table)
	if siteID != uuid.Nil {
		coll, err := GetStyleCollectionForSite(ctx, params.DB, siteID, params.Logger)
		if err == nil && coll != nil {
			// Already persisted — just return
			return styleCollectionToResult(coll), nil
		}
	}

	// Priority 4: Match by domain keywords
	domainField := "input_data.domain"
	if df, ok := config["domain_field"].(string); ok && df != "" {
		domainField = df
	}
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)
	if domain != "" {
		coll, err := SelectStyleCollectionByDomain(ctx, params.DB, domain, params.Logger)
		if err == nil && coll != nil {
			return persistAndReturn(coll, "domain_keywords")
		}
	}

	// Fallback: Return default (not persisted — it's a synthetic default)
	params.Logger.Info("SelectStyleCollectionAction: No matching style found, using default")
	return getDefaultStyleCollection(), nil
}

func styleCollectionToResult(coll *StyleCollection) map[string]interface{} {
	result := map[string]interface{}{
		"style_collection_id": coll.ID.String(),
		"name":                coll.Name,
		"display_name":        coll.DisplayName,
		"category":            coll.Category,
		"color_palette":       coll.ColorPalette,
		"typography":          coll.Typography,
	}
	if coll.HeaderComponentID != nil {
		result["header_component_id"] = coll.HeaderComponentID.String()
	}
	if coll.FooterComponentID != nil {
		result["footer_component_id"] = coll.FooterComponentID.String()
	}
	if coll.CSSThemeID != nil {
		result["css_theme_id"] = coll.CSSThemeID.String()
	}
	return result
}

func getDefaultStyleCollection() map[string]interface{} {
	return map[string]interface{}{
		"style_collection_id": "",
		"name":                "default",
		"display_name":        "Default Style",
		"category":            "general",
		"color_palette": map[string]string{
			"primary":    "#1a1a2e",
			"secondary":  "#16213e",
			"accent":     "#0f3460",
			"text":       "#333333",
			"background": "#ffffff",
		},
		"typography": map[string]string{
			"heading": "system-ui, sans-serif",
			"body":    "system-ui, sans-serif",
		},
	}
}

// ============================================================================
// ACTION: update_site_content
// ============================================================================

// UpdateSiteContentAction updates the sites.content_data JSONB column
// Used to store the site plan, brand DNA, or other structured content
// Config:
//   - site_id_field: path to site_id in collected_data
//   - content_field: path to content data to store
//   - merge: boolean - if true, merges with existing content_data
func UpdateSiteContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteContentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get content to store
	contentField := "page_plan"
	if f, ok := config["content_field"].(string); ok && f != "" {
		contentField = f
	}
	contentValue := datahelpers.ExtractNestedField(params.CollectedData, contentField)
	if contentValue == nil {
		return nil, fmt.Errorf("content not found at %s", contentField)
	}

	contentJSON, err := json.Marshal(contentValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteContentAction: No database, skipping update")
		return map[string]interface{}{
			"updated":     false,
			"site_id":     siteIDStr,
			"reason":      "no database connection",
			"content_key": contentField,
		}, nil
	}

	// Determine if merge or replace
	merge, _ := config["merge"].(bool)

	var query string
	if merge {
		query = `
			UPDATE sites 
			SET content_data = COALESCE(content_data, '{}'::jsonb) || $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	} else {
		query = `
			UPDATE sites 
			SET content_data = $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	}

	if err := execDB(ctx, params.DB, query, siteID, string(contentJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site content: %w", err)
	}

	// --- Sync key columns to sites table ---
	// When storing brief/identity data, also populate the sites columns
	// so loadSiteDataFull and RenderSiteComponentsAction can find them.
	syncColumns, _ := config["sync_columns"].(bool)
	if syncColumns {
		if contentMap, ok := contentValue.(map[string]interface{}); ok {
			companyName := getFirstNonEmpty(contentMap, "company_name")
			tagline := getFirstNonEmpty(contentMap, "tagline")
			email := getFirstNonEmpty(contentMap, "contact_email", "email")
			phone := getFirstNonEmpty(contentMap, "contact_phone", "phone")

			if companyName != "" || tagline != "" || email != "" || phone != "" {
				syncErr := execDB(ctx, params.DB, `
					UPDATE sites SET
						company_name = CASE WHEN COALESCE(company_name, '') IN ('', domain) AND $2 != '' THEN $2 ELSE company_name END,
						tagline      = CASE WHEN COALESCE(tagline, '')      = '' AND $3 != '' THEN $3 ELSE tagline END,
						email        = CASE WHEN COALESCE(email, '')        = '' AND $4 != '' THEN $4 ELSE email END,
						phone        = CASE WHEN COALESCE(phone, '')        = '' AND $5 != '' THEN $5 ELSE phone END,
						updated_at   = now()
					WHERE id = $1
				`, siteID, companyName, tagline, email, phone)
				if syncErr != nil {
					params.Logger.Warn("UpdateSiteContentAction: column sync failed", zap.Error(syncErr))
				} else {
					params.Logger.Info("UpdateSiteContentAction: synced columns",
						zap.String("company_name", companyName),
						zap.String("email", email),
					)
				}
			}
		}
	}

	params.Logger.Info("UpdateSiteContentAction: Site content updated",
		zap.String("site_id", siteIDStr),
		zap.String("content_field", contentField),
		zap.Bool("merged", merge),
	)

	return map[string]interface{}{
		"updated":      true,
		"site_id":      siteIDStr,
		"content_key":  contentField,
		"content_size": len(contentJSON),
		"merged":       merge,
	}, nil
}

// getFirstNonEmpty returns the first non-empty string value found at any of
// the given keys in the map.
func getFirstNonEmpty(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// ============================================================================
// ACTION: update_site_status
// ============================================================================

// UpdateSiteStatusAction updates the sites.status column
// Config:
//   - site_id_field: path to site_id
//   - status: new status value (draft, building, review, published, archived)
func UpdateSiteStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get new status
	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required in config")
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft": true, "building": true, "review": true,
		"published": true, "deployed": true, "archived": true, "error": true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s (valid: draft, building, review, published, deployed, archived, error)", newStatus)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	// Check if deployed_at should be set
	var query string
	deployedAt, hasDeployedAt := config["deployed_at"].(string)
	if hasDeployedAt && (deployedAt == "now" || deployedAt == "NOW()") && newStatus == "deployed" {
		query = `UPDATE sites SET status = $2, last_deployed_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else {
		query = `UPDATE sites SET status = $2, updated_at = NOW() WHERE id = $1`
	}

	if err := execDB(ctx, params.DB, query, siteID, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update site status: %w", err)
	}

	params.Logger.Info("UpdateSiteStatusAction: Status updated",
		zap.String("site_id", siteIDStr),
		zap.String("status", newStatus),
	)

	return map[string]interface{}{
		"updated":   true,
		"site_id":   siteIDStr,
		"status":    newStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ============================================================================
// ACTION: update_site_defaults
// ============================================================================

// UpdateSiteDefaultsAction updates the sites.default_components JSONB column
// Used to store default header/footer component IDs
// Config:
//   - site_id_field: path to site_id in collected_data
//   - header_component_id: UUID of header component (optional)
//   - footer_component_id: UUID of footer component (optional)
//   - css_theme_id: UUID of CSS theme (optional)
//   - defaults_field: path to a map containing all defaults (alternative to individual fields)
func UpdateSiteDefaultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteDefaultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Build defaults map
	defaults := make(map[string]interface{})

	// Try getting from defaults_field first
	if defaultsField, ok := config["defaults_field"].(string); ok && defaultsField != "" {
		if defaultsData := datahelpers.ExtractNestedField(params.CollectedData, defaultsField); defaultsData != nil {
			if m, ok := defaultsData.(map[string]interface{}); ok {
				defaults = m
			}
		}
	}

	// Override/add from explicit config fields
	if headerID, ok := config["header_component_id"].(string); ok && headerID != "" {
		defaults["header_component_id"] = headerID
	}
	if footerID, ok := config["footer_component_id"].(string); ok && footerID != "" {
		defaults["footer_component_id"] = footerID
	}
	if cssThemeID, ok := config["css_theme_id"].(string); ok && cssThemeID != "" {
		defaults["css_theme_id"] = cssThemeID
	}

	// Also check collected data for style_collection results
	if sc := datahelpers.ExtractNestedField(params.CollectedData, "style_collection"); sc != nil {
		if scMap, ok := sc.(map[string]interface{}); ok {
			if hID, ok := scMap["header_component_id"].(string); ok && hID != "" {
				if defaults["header_component_id"] == nil {
					defaults["header_component_id"] = hID
				}
			}
			if fID, ok := scMap["footer_component_id"].(string); ok && fID != "" {
				if defaults["footer_component_id"] == nil {
					defaults["footer_component_id"] = fID
				}
			}
			if cID, ok := scMap["css_theme_id"].(string); ok && cID != "" {
				if defaults["css_theme_id"] == nil {
					defaults["css_theme_id"] = cID
				}
			}
		}
	}

	if len(defaults) == 0 {
		params.Logger.Warn("UpdateSiteDefaultsAction: No defaults to set")
		return map[string]interface{}{
			"updated": false,
			"site_id": siteIDStr,
			"reason":  "no defaults provided",
		}, nil
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteDefaultsAction: No database")
		return map[string]interface{}{
			"updated":  false,
			"site_id":  siteIDStr,
			"defaults": defaults,
			"reason":   "no database connection",
		}, nil
	}

	defaultsJSON, err := json.Marshal(defaults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal defaults: %w", err)
	}

	query := `
		UPDATE sites 
		SET default_components = COALESCE(default_components, '{}'::jsonb) || $2::jsonb,
		    updated_at = NOW()
		WHERE id = $1
	`

	if err := execDB(ctx, params.DB, query, siteID, string(defaultsJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site defaults: %w", err)
	}

	params.Logger.Info("UpdateSiteDefaultsAction: Defaults updated",
		zap.String("site_id", siteIDStr),
		zap.Any("defaults", defaults),
	)

	return map[string]interface{}{
		"updated":  true,
		"site_id":  siteIDStr,
		"defaults": defaults,
	}, nil
}

// ============================================================================
// ACTION: update_page_status
// ============================================================================

// UpdatePageStatusAction updates a single page's build_status
// Config:
//   - page_id_field: path to page_id OR
//   - site_id_field + page_name_field: to look up page
//   - status: new build_status value (e.g., "deployed", "failed", "building")
func UpdatePageStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageStatusAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("UpdatePageStatusAction: No database connection available")
		return map[string]interface{}{"updated": false, "reason": "no database"}, nil
	}

	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required")
	}

	var pageID uuid.UUID

	// Try direct page_id from config field
	if pageIDField, ok := config["page_id_field"].(string); ok && pageIDField != "" {
		pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pageIDField)
		if pageIDStr != "" {
			var err error
			pageID, err = uuid.Parse(pageIDStr)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Invalid page_id format",
					zap.String("page_id_field", pageIDField),
					zap.String("value", pageIDStr),
					zap.Error(err))
				return nil, fmt.Errorf("invalid page_id: %w", err)
			}
		}
	}

	// Alternative: try current_page.id (common in loop iterations)
	if pageID == uuid.Nil {
		if pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.id"); pageIDStr != "" {
			if parsed, err := uuid.Parse(pageIDStr); err == nil {
				pageID = parsed
			}
		}
	}

	// Alternative: look up by site_id + page_name
	if pageID == uuid.Nil {
		siteIDField, _ := config["site_id_field"].(string)
		pageNameField, _ := config["page_name_field"].(string)

		if siteIDField != "" && pageNameField != "" {
			siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
			pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, pageNameField)

			if siteIDStr != "" && pageName != "" {
				siteUUID, _ := uuid.Parse(siteIDStr)
				var err error
				pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
				if err != nil {
					return nil, fmt.Errorf("page lookup failed: %w", err)
				}
			}
		}
	}

	// Last resort: try current_page.name with site_record.site_id
	if pageID == uuid.Nil {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.name")

		if siteIDStr != "" && pageName != "" {
			siteUUID, _ := uuid.Parse(siteIDStr)
			var err error
			pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Page lookup by name failed",
					zap.String("site_id", siteIDStr),
					zap.String("page_name", pageName),
					zap.Error(err))
			}
		}
	}

	if pageID == uuid.Nil {
		params.Logger.Error("UpdatePageStatusAction: Could not determine page_id",
			zap.Any("config", config))
		return nil, fmt.Errorf("could not determine page_id")
	}

	// Option B guard: never mark a page "deployed" without rendered components.
	// build_status='deployed' is what the reconciler trusts to skip a page, so a
	// 0-component page marked deployed becomes permanently fileless (gamesdesign
	// homepage, 2026-06-04: its needs_content_page was auto-completed on a lost
	// response with zero components, then a deploy path stamped it deployed, so
	// the reconciler never rebuilt it). Refuse the deploy mark and flip to
	// needs_rebuild (clearing the stamp) so the reconciler re-emits a build.
	// Fail-open on a check error: a transient check failure must not halt
	// legitimate deploys; Option A (the claimed-item-timeout evidence check) is
	// the other layer of protection.
	//
	// bugs_open/040-partial-build: the same reasoning applies one step up, to a
	// build that wrote SOME but not ALL of its planned sections. dartsonline
	// index (2026-07-20) reached this deploy mark with 5 of 6 planned sections
	// (testimonials never written) and was stamped deployed + built_from_plan_version
	// = current, so decideEmit returns skip_built and the reconciler will never
	// revisit the missing section — a permanent five-sixths page that no longer
	// asks to be built. A partial build must be treated exactly like a 0-component
	// one: refuse the mark, flip to needs_rebuild, clear the stamp. This runs
	// AFTER save_page_sections has written the components (page-build-handler,
	// page-rerender, section-editor and tool-recreation-handler all write their
	// components before this step), so a shortfall here is a real one, not a race.
	if newStatus == "deployed" {
		hasComponents, checkErr := pageHasComponents(ctx, params.DB, pageID)
		switch {
		case checkErr != nil:
			params.Logger.Warn("UpdatePageStatusAction: component check failed; proceeding with deploy",
				zap.String("page_id", pageID.String()),
				zap.Error(checkErr))
		case !hasComponents:
			params.Logger.Warn("UpdatePageStatusAction: refusing to mark page deployed with no rendered components; setting needs_rebuild",
				zap.String("page_id", pageID.String()))
			const rebuildQuery = `UPDATE pages SET build_status = 'needs_rebuild', built_from_plan_version = NULL, updated_at = NOW() WHERE id = $1`
			if rbErr := execDB(ctx, params.DB, rebuildQuery, pageID); rbErr != nil {
				params.Logger.Error("UpdatePageStatusAction: failed to set needs_rebuild after refusing deploy",
					zap.String("page_id", pageID.String()),
					zap.Error(rbErr))
				return nil, fmt.Errorf("failed to set needs_rebuild for 0-component page: %w", rbErr)
			}
			return map[string]interface{}{
				"updated":      false,
				"page_id":      pageID.String(),
				"build_status": "needs_rebuild",
				"reason":       "refused deploy: page has no rendered components",
			}, nil
		default:
			// Page has >= 1 component but may still be short of its plan.
			// Fail-open on a check error, same as the 0-component guard above.
			// suppressed_sections (subtracted inside pageSectionShortfall) is
			// maintained by plan_sections' persistSectionSkips: an
			// on_missing=skip_section name is added there and removed again the
			// build it plans ready — so a legitimately data-gated section does
			// not count as a shortfall here (bugs_open/040 skip-not-recorded).
			planned, rendered, shErr := pageSectionShortfall(ctx, params.DB, pageID)
			if shErr != nil {
				params.Logger.Warn("UpdatePageStatusAction: section-shortfall check failed; proceeding with deploy",
					zap.String("page_id", pageID.String()),
					zap.Error(shErr))
			} else if rendered < planned {
				params.Logger.Warn("UpdatePageStatusAction: refusing to mark page deployed; build is short of its plan; setting needs_rebuild",
					zap.String("page_id", pageID.String()),
					zap.Int("planned_sections", planned),
					zap.Int("rendered_components", rendered))
				const rebuildQuery = `UPDATE pages SET build_status = 'needs_rebuild', built_from_plan_version = NULL, updated_at = NOW() WHERE id = $1`
				if rbErr := execDB(ctx, params.DB, rebuildQuery, pageID); rbErr != nil {
					params.Logger.Error("UpdatePageStatusAction: failed to set needs_rebuild after refusing partial deploy",
						zap.String("page_id", pageID.String()),
						zap.Error(rbErr))
					return nil, fmt.Errorf("failed to set needs_rebuild for partial-build page: %w", rbErr)
				}
				return map[string]interface{}{
					"updated":      false,
					"page_id":      pageID.String(),
					"build_status": "needs_rebuild",
					"reason":       fmt.Sprintf("refused deploy: only %d of %d planned sections rendered", rendered, planned),
				}, nil
			}
		}
	}

	// Build the query - use build_status column (not status)
	var query string
	if newStatus == "deployed" {
		// Also set deployed_at, and stamp built_from_plan_version with the site's
		// current plan id. This is the build-time drift stamp the reconciler
		// compares against (029/030 design; the deferred item in HANDOFF_2026-05-07
		// #5). COALESCE keeps any existing value when no current plan exists yet —
		// e.g. tool-recreation deploys before build-site-planner has written the
		// plan — and SyncPagesToDBAction then fills it on its first pass. With this
		// stamp in place the reconciler detects genuine drift (built_from_plan_version
		// != current) rather than relying on the blunt deployed->needs_rebuild flip.
		query = `UPDATE pages
		         SET build_status = $2,
		             deployed_at = NOW(),
		             built_from_plan_version = COALESCE(
		                 (SELECT sp.id FROM site_plans sp
		                   WHERE sp.site_id = pages.site_id AND sp.is_current = true),
		                 built_from_plan_version
		             ),
		             updated_at = NOW()
		         WHERE id = $1`
	} else {
		query = `UPDATE pages SET build_status = $2, updated_at = NOW() WHERE id = $1`
	}

	result, err := params.DB.ExecContext(ctx, query, pageID, newStatus)
	if err != nil {
		params.Logger.Error("UpdatePageStatusAction: Failed to update page",
			zap.String("page_id", pageID.String()),
			zap.String("build_status", newStatus),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update page build_status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	params.Logger.Info("UpdatePageStatusAction: Updated page",
		zap.String("page_id", pageID.String()),
		zap.String("build_status", newStatus),
		zap.Int64("rows_affected", rowsAffected))

	// Mirror the deploy mark onto one page_component when the caller names it via
	// config page_component_id_field. Every discovery check matches
	// page_components.build_status = 'deployed', and apply_section_edit leaves its
	// row at 'approved', so without this an edited section silently disappears from
	// the whole audit surface (check_empty_sections, check_image_url_404,
	// check_undeployed_assets, check_placeholder_image_in_use, check_component_standards).
	// Non-fatal: the page row is already committed, and failing here would re-run the
	// step and re-deploy.
	componentUpdated := false
	if newStatus == "deployed" {
		if field, ok := config["page_component_id_field"].(string); ok && field != "" {
			pcIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, field)
			switch {
			case pcIDStr == "":
				params.Logger.Warn("UpdatePageStatusAction: page_component_id_field configured but empty",
					zap.String("page_component_id_field", field))
			default:
				pcID, parseErr := uuid.Parse(pcIDStr)
				if parseErr != nil {
					params.Logger.Warn("UpdatePageStatusAction: invalid page_component_id",
						zap.String("value", pcIDStr),
						zap.Error(parseErr))
					break
				}
				const pcQuery = `UPDATE page_components SET build_status = $2, updated_at = NOW() WHERE id = $1`
				if pcErr := execDB(ctx, params.DB, pcQuery, pcID, newStatus); pcErr != nil {
					params.Logger.Error("UpdatePageStatusAction: failed to mark page_component deployed",
						zap.String("page_component_id", pcID.String()),
						zap.Error(pcErr))
					break
				}
				componentUpdated = true
				params.Logger.Info("UpdatePageStatusAction: Marked page_component deployed",
					zap.String("page_component_id", pcID.String()))
			}
		}
	}

	return map[string]interface{}{
		"updated":                true,
		"page_id":                pageID.String(),
		"build_status":           newStatus,
		"rows_affected":          rowsAffected,
		"page_component_updated": componentUpdated,
	}, nil
}

// pageHasComponents reports whether a page has at least one real rendered
// component (non-null component_id, non-empty rendered_html). This is the
// "positive evidence" check from FOCUS_page_build_handler_silent_completion.md,
// used by Option B to stop a 0-component page being marked deployed. Mirrors the
// db type switch used by lookupPageID/execDB so it works with both *sql.DB and
// *pgxpool.Pool.
func pageHasComponents(ctx context.Context, db interface{}, pageID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM page_components
			WHERE page_id = $1
			  AND component_id IS NOT NULL
			  AND rendered_html IS NOT NULL
			  AND rendered_html <> ''
		)`
	var exists bool
	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, pageID).Scan(&exists)
		return exists, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, pageID).Scan(&exists)
		return exists, err
	default:
		return false, fmt.Errorf("unsupported database type: %T", db)
	}
}

// pageSectionShortfall reports how many sections the page's plan promised
// (planned) versus how many page_components rows exist for it (rendered). It is
// the 040-partial-build companion to pageHasComponents: a build that reaches the
// deploy mark having written a row for only 5 of 6 planned sections must NOT be
// stamped deployed + built_from_plan_version, or decideEmit
// (reconcile_site_plan_action.go) returns skip_built for it forever and the
// reconciler never revisits the shortfall (dartsonline index, 2026-07-20:
// testimonials dropped by a build that reported complete; caller flips it to
// needs_rebuild and clears the stamp instead).
//
// It compares ROW COUNTS, deliberately NOT section names. pages.sections names
// do NOT reliably equal page_components.slot_name / content_components.function
// across templates — gaswholesalers services planned ["services-hero", …,
// "call_to_action"] against live slots ["hero-services", …, "call-to-action"]
// (word order and _/- both differ) — so per-name matching produces false
// positives that would refuse a healthy page and drive it into a rebuild loop.
// The row count is the signal the bugs_open/040 fleet sweep validated; a genuine
// missing section (no row at all) is what it catches, while a present-but-hollow
// section (a row that exists but rendered nothing) is bugs_open/039's concern and
// still counts as rendered here. suppressed_sections are excluded from the
// planned count so a deliberately-dropped section is never read as a shortfall.
// Mirrors the db type switch used by pageHasComponents/execDB.
func pageSectionShortfall(ctx context.Context, db interface{}, pageID uuid.UUID) (planned int, rendered int, err error) {
	const query = `
		SELECT
			(SELECT count(*)
			   FROM jsonb_array_elements_text(COALESCE(p.sections, '[]'::jsonb)) AS sec
			  WHERE sec NOT IN (
			      SELECT jsonb_array_elements_text(COALESCE(p.suppressed_sections, '[]'::jsonb)))),
			(SELECT count(*) FROM page_components pc WHERE pc.page_id = p.id)
		FROM pages p
		WHERE p.id = $1`
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, pageID).Scan(&planned, &rendered)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, pageID).Scan(&planned, &rendered)
	default:
		err = fmt.Errorf("unsupported database type: %T", db)
	}
	return planned, rendered, err
}

// BuildRenderContextAction assembles a RenderContext from multiple sources
// Used before rendering components with templates
func BuildRenderContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildRenderContextAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Start with empty context
	renderCtx := &RenderContext{
		Year: fmt.Sprintf("%d", time.Now().Year()),
	}

	sourcesMerged := 0

	// Handle sources config - can be array or map format
	// Array format: ["input_data", "site_record", ...]
	// Map format: {"page": "input_data.current_page", "site": "input_data.site_record", ...}

	if sourcesMap, ok := config["sources"].(map[string]interface{}); ok {
		// Map format: keys are logical names, values are paths to data
		params.Logger.Info("Using map-format sources config",
			zap.Int("source_count", len(sourcesMap)))

		for logicalName, pathVal := range sourcesMap {
			path, ok := pathVal.(string)
			if !ok {
				continue
			}

			sourceData := datahelpers.ExtractNestedField(params.CollectedData, path)
			if sourceData == nil {
				params.Logger.Debug("Source not found",
					zap.String("name", logicalName),
					zap.String("path", path))
				continue
			}

			if m, ok := sourceData.(map[string]interface{}); ok {
				params.Logger.Debug("Merging source",
					zap.String("name", logicalName),
					zap.String("path", path))
				mergeIntoRenderContextEnhanced(renderCtx, m, logicalName, params.Logger)
				sourcesMerged++
			}
		}
	} else if sourcesArray, ok := config["sources"].([]interface{}); ok {
		// Array format: direct paths
		for _, src := range sourcesArray {
			if srcStr, ok := src.(string); ok {
				sourceData := datahelpers.ExtractNestedField(params.CollectedData, srcStr)
				if sourceData == nil {
					continue
				}
				if m, ok := sourceData.(map[string]interface{}); ok {
					mergeIntoRenderContextEnhanced(renderCtx, m, srcStr, params.Logger)
					sourcesMerged++
				}
			}
		}
	} else {
		// Default sources
		defaultSources := []string{"input_data", "site_record", "style_collection", "reviewed_brief", "page_plan"}
		for _, source := range defaultSources {
			sourceData := datahelpers.ExtractNestedField(params.CollectedData, source)
			if sourceData == nil {
				// Try with input_data prefix
				sourceData = datahelpers.ExtractNestedField(params.CollectedData, "input_data."+source)
			}
			if sourceData == nil {
				continue
			}
			if m, ok := sourceData.(map[string]interface{}); ok {
				mergeIntoRenderContextEnhanced(renderCtx, m, source, params.Logger)
				sourcesMerged++
			}
		}
	}

	// Try to load navigation from DB if we have site_id
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if siteUUID, err := uuid.Parse(siteIDStr); err == nil && params.DB != nil {
				renderCtx.SiteID = siteUUID
				/*nav, _ := getNavigationFromDB(ctx, params.DB, siteUUID, "header", params.Logger)*/
				headerNav := GetNavItems(ctx, params.DB, siteUUID, []string{NavGroupPrimary}, false, 0, params.Logger)
				if len(headerNav) > 0 {
					renderCtx.NavItems = headerNav
				}
			}
		}
	}

	// Extract image URLs from deploy_image_asset output
	// (adds logo_deployed block between hero_deployed and fallback)
	// =========================================================================
	if heroDeployed, ok := params.CollectedData["hero_deployed"].(map[string]interface{}); ok {
		if imageURL, ok := heroDeployed["image_url"].(string); ok && imageURL != "" {
			if renderCtx.ContentData == nil {
				renderCtx.ContentData = make(map[string]interface{})
			}
			renderCtx.ContentData["hero_url"] = imageURL
			params.Logger.Info("Set hero_url from hero_deployed.image_url",
				zap.String("url", imageURL))
		}
	}

	if logoDeployed, ok := params.CollectedData["logo_deployed"].(map[string]interface{}); ok {
		if imageURL, ok := logoDeployed["image_url"].(string); ok && imageURL != "" {
			if renderCtx.ContentData == nil {
				renderCtx.ContentData = make(map[string]interface{})
			}
			renderCtx.ContentData["logo_url"] = imageURL
			renderCtx.LogoURL = imageURL
			params.Logger.Info("Set logo_url from logo_deployed.image_url",
				zap.String("url", imageURL))
		}
	}

	// Also check for direct hero_url in collected_data (fallback)
	for _, field := range []string{"hero_url", "hero_home_url", "logo_url"} {
		if url := datahelpers.ExtractNestedFieldString(params.CollectedData, field); url != "" {
			if renderCtx.ContentData == nil {
				renderCtx.ContentData = make(map[string]interface{})
			}
			renderCtx.ContentData[field] = url
			if field == "logo_url" {
				renderCtx.LogoURL = url
			}
		}
	}

	params.Logger.Info("BuildRenderContextAction: Context built",
		zap.String("domain", renderCtx.Domain),
		zap.String("company_name", renderCtx.CompanyName),
		zap.String("logo_text", renderCtx.LogoText),
		zap.Int("nav_items", len(renderCtx.NavItems)),
		zap.Int("sources_merged", sourcesMerged),
	)

	// Return the context directly, not wrapped in "render_context" key
	// The workflow output_field already specifies where to store it
	// Adding metadata fields at same level
	result := renderCtxToMap(renderCtx)
	result["_sources_merged"] = sourcesMerged
	result["_built_at"] = time.Now().Format(time.RFC3339)

	return result, nil
}

// mergeIntoRenderContextEnhanced extracts data from various source formats
func mergeIntoRenderContextEnhanced(ctx *RenderContext, data map[string]interface{}, sourceName string, logger *zap.Logger) {
	// =========================================================================
	// STEP 1: Unwrap .response wrapper if present (common in agent responses)
	// reviewed_brief has: {"response": {...actual data...}, "response_status": "complete"}
	// This recursively processes the unwrapped data first
	// =========================================================================
	if response, ok := data["response"].(map[string]interface{}); ok {
		// Check if this looks like a wrapped response (has response_status sibling)
		if _, hasStatus := data["response_status"]; hasStatus {
			logger.Debug("Unwrapping .response wrapper for source",
				zap.String("source", sourceName))
			// Recursively merge the unwrapped response data FIRST
			mergeIntoRenderContextEnhanced(ctx, response, sourceName+".response", logger)
			// Then continue to process the outer data for any additional fields
		}
	}

	// =========================================================================
	// STEP 2: Direct field extraction from current data level
	// =========================================================================

	// Domain
	if v, ok := data["domain"].(string); ok && v != "" {
		ctx.Domain = v
	}

	// Company name (sets logo_text as fallback)
	if v, ok := data["company_name"].(string); ok && v != "" {
		ctx.CompanyName = v
		if ctx.LogoText == "" {
			ctx.LogoText = v
		}
	}

	// Logo text (explicit override)
	if v, ok := data["logo_text"].(string); ok && v != "" {
		ctx.LogoText = v
	}

	// Tagline
	if v, ok := data["tagline"].(string); ok && v != "" {
		ctx.Tagline = v
	}

	// Email - check both "email" and "contact_email" (reviewed_brief uses contact_email)
	if v, ok := data["email"].(string); ok && v != "" {
		ctx.Email = v
	}
	if v, ok := data["contact_email"].(string); ok && v != "" {
		ctx.Email = v
	}

	// Phone - check both "phone" and "contact_phone"
	if v, ok := data["phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["contact_phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["contact_email"].(string); ok && v != "" {
		ctx.Email = v
	}

	// Extract image URLs from content_data or collected_data
	imageURLFields := []string{
		"hero_url",      // pageflow-builder uses this
		"hero_home_url", // multipage-website-builder uses this
		"hero_about_url",
		"hero_services_url",
		"logo_url",
	}

	for _, field := range imageURLFields {
		if v, ok := data[field].(string); ok && v != "" {
			if ctx.ContentData == nil {
				ctx.ContentData = make(map[string]interface{})
			}
			ctx.ContentData[field] = v
			logger.Debug("Extracted image URL",
				zap.String("field", field),
				zap.String("url", v))
		}
	}

	// Colors - direct fields
	if v, ok := data["primary_color"].(string); ok && v != "" {
		ctx.PrimaryColor = v
	}
	if v, ok := data["secondary_color"].(string); ok && v != "" {
		ctx.SecondaryColor = v
	}
	if v, ok := data["accent_color"].(string); ok && v != "" {
		ctx.AccentColor = v
	}
	if v, ok := data["text_color"].(string); ok && v != "" {
		ctx.TextColor = v
	}
	if v, ok := data["background_color"].(string); ok && v != "" {
		ctx.BackgroundColor = v
	}

	// =========================================================================
	// STEP 3: Check nested color_palette (from style_collection)
	// =========================================================================
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok && v != "" {
			ctx.AccentColor = v
		}
		if v, ok := palette["background"].(string); ok && v != "" {
			ctx.BackgroundColor = v
		}
		if v, ok := palette["text"].(string); ok && v != "" {
			ctx.TextColor = v
		}
	}

	// =========================================================================
	// STEP 4: Extract from nested structures (business_context, contact_info, brand)
	// =========================================================================

	// business_context (alternative structure)
	if brief, ok := data["business_context"].(map[string]interface{}); ok {
		if v, ok := brief["company_name"].(string); ok && v != "" {
			ctx.CompanyName = v
			if ctx.LogoText == "" {
				ctx.LogoText = v
			}
		}
		if v, ok := brief["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
		if v, ok := brief["industry"].(string); ok && v != "" {
			ctx.Industry = v
		}
	}

	// contact_info (nested contact structure)
	if contact, ok := data["contact_info"].(map[string]interface{}); ok {
		if v, ok := contact["email"].(string); ok && v != "" {
			ctx.Email = v
		}
		if v, ok := contact["phone"].(string); ok && v != "" {
			ctx.Phone = v
		}
	}

	// brand (nested brand/visual settings)
	if brand, ok := data["brand"].(map[string]interface{}); ok {
		if v, ok := brand["primary_color"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := brand["secondary_color"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := brand["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
	}

	// =========================================================================
	// STEP 5: Content generation context (tone, target_audience, industry)
	// =========================================================================
	if v, ok := data["tone"].(string); ok && v != "" {
		ctx.Tone = v
	}
	if v, ok := data["target_audience"].(string); ok && v != "" {
		ctx.TargetAudience = v
	}
	if v, ok := data["industry"].(string); ok && v != "" {
		ctx.Industry = v
	}

	// =========================================================================
	// STEP 6: Site/page identifiers
	// =========================================================================
	if v, ok := data["site_id"].(string); ok && v != "" {
		if siteUUID, err := uuid.Parse(v); err == nil {
			ctx.SiteID = siteUUID
		}
	}

	// =========================================================================
	// STEP 7: CTA settings
	// =========================================================================
	if v, ok := data["cta_text"].(string); ok && v != "" {
		ctx.CTAText = v
	}
	if v, ok := data["cta_url"].(string); ok && v != "" {
		ctx.CTAUrl = v
	}

	// =========================================================================
	// STEP 8: Extract services array (for footer and services sections)
	// Services appear in reviewed_brief.response.services as []interface{}
	// Each service is {"name": "...", "description": "..."}
	//
	// ctx.Services is []string (just names)
	// ctx.ContentData["services"] is []interface{} (full objects for {{range .services}})
	// =========================================================================
	if services, ok := data["services"].([]interface{}); ok && len(services) > 0 && len(ctx.Services) == 0 {
		// Store full services in ContentData for template access via {{range .services}}
		// Only if not already populated (avoid duplicates from brief + brief.response)
		if ctx.ContentData == nil {
			ctx.ContentData = make(map[string]interface{})
		}
		// ctx.ContentData["services"] = services
		ctx.ContentData["services"] = normaliseToNameDescArray(services)

		// Also extract just the names to ctx.Services ([]string)
		for _, svc := range services {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				if name, ok := svcMap["name"].(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				}
			}
		}

		logger.Info("Extracted services array",
			zap.String("source", sourceName),
			zap.Int("full_count", len(services)),
			zap.Int("names_count", len(ctx.Services)))
	}

	// =========================================================================
	// STEP 9: Extract navigation from db_sync source
	// db_sync contains: {"navigation": {"items": [{"label": "Home", "url": "/index.html"}, ...]}}
	// =========================================================================
	if sourceName == "db_sync" || sourceName == "db_sync.response" {
		if navigation, ok := data["navigation"].(map[string]interface{}); ok {
			if items, ok := navigation["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)
						if label != "" && url != "" {
							ctx.NavItems = append(ctx.NavItems, NavItem{
								Label: label,
								URL:   url,
							})
						}
					}
				}
				if len(ctx.NavItems) > 0 {
					logger.Info("Extracted navigation items from db_sync",
						zap.Int("count", len(ctx.NavItems)))
				}
			}
		}
	}

	// Handle site_record.content_data - recursively merge nested content_data
	if sourceName == "site_record" || sourceName == "site" {
		if contentData, ok := data["content_data"].(map[string]interface{}); ok {
			logger.Debug("Processing site_record.content_data",
				zap.Int("fields", len(contentData)))
			mergeIntoRenderContextEnhanced(ctx, contentData, "site_record.content_data", logger)
		}
	}

	// Handle services array (for services_html generation)
	if len(ctx.Services) == 0 {
		if services, ok := data["services"].([]interface{}); ok && len(services) > 0 {
			for _, svc := range services {
				if name, ok := svc.(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				} else if svcMap, ok := svc.(map[string]interface{}); ok {
					if name, ok := svcMap["name"].(string); ok && name != "" {
						ctx.Services = append(ctx.Services, name)
					}
				}
			}
		}
	}

	// =========================================================================
	// STEP 10: Log final state for debugging
	// =========================================================================
	logger.Info("Merged source into render context",
		zap.String("source", sourceName),
		zap.String("company_name", ctx.CompanyName),
		zap.String("domain", ctx.Domain),
		zap.String("tagline", ctx.Tagline),
		zap.String("email", ctx.Email),
		zap.Int("nav_items", len(ctx.NavItems)),
		zap.Int("services", len(ctx.Services)))
}

// renderCtxToMap converts RenderContext to map for template substitution
func renderCtxToMap(ctx *RenderContext) map[string]interface{} {
	result := map[string]interface{}{
		"domain":           ctx.Domain,
		"logo_text":        ctx.LogoText,
		"company_name":     ctx.CompanyName,
		"tagline":          ctx.Tagline,
		"email":            ctx.Email,
		"phone":            ctx.Phone,
		"primary_color":    ctx.PrimaryColor,
		"secondary_color":  ctx.SecondaryColor,
		"accent_color":     ctx.AccentColor,
		"text_color":       ctx.TextColor,
		"background_color": ctx.BackgroundColor,
		"year":             ctx.Year,
		"cta_text":         ctx.CTAText,
		"cta_url":          ctx.CTAUrl,
		"industry":         ctx.Industry,
		"tone":             ctx.Tone,
		"target_audience":  ctx.TargetAudience,
	}

	if ctx.SiteID != uuid.Nil {
		result["site_id"] = ctx.SiteID.String()
	}

	// =========================================================================
	// Generate nav_items and nav_items_html from NavItems
	// Templates may use either:
	//   {{range .nav_items}} for iteration
	//   {{.nav_items_html}} for pre-rendered HTML
	// =========================================================================
	if len(ctx.NavItems) > 0 {
		// nav_items as array (for {{range .nav_items}})
		navItems := make([]map[string]interface{}, len(ctx.NavItems))
		for i, item := range ctx.NavItems {
			navItems[i] = map[string]interface{}{
				"label":     item.Label,
				"url":       item.URL,
				"is_active": item.IsActive,
			}
		}
		result["nav_items"] = navItems

		// nav_items_html as pre-rendered string (for {{.nav_items_html}})
		// Format: <li><a href="/page.html">Label</a></li>
		var htmlParts []string
		for _, item := range ctx.NavItems {
			activeClass := ""
			if item.IsActive {
				activeClass = ` class="active"`
			}
			htmlParts = append(htmlParts, fmt.Sprintf(
				`<li><a href="%s"%s>%s</a></li>`,
				item.URL, activeClass, item.Label,
			))
		}
		result["nav_items_html"] = strings.Join(htmlParts, "\n                ")
	} else {
		// Ensure empty values don't render as "<no value>"
		result["nav_items"] = []map[string]interface{}{}
		result["nav_items_html"] = ""
	}

	// =========================================================================
	// Generate services list HTML for footer
	// Templates use {{.services_html}} for footer services list
	// ctx.Services is []string (service names extracted from reviewed_brief)
	// =========================================================================
	if len(ctx.Services) > 0 {
		result["services"] = ctx.Services

		// services_html as pre-rendered string
		var servicesParts []string
		for _, serviceName := range ctx.Services {
			if serviceName != "" {
				servicesParts = append(servicesParts, fmt.Sprintf(
					`<li><a href="/services.html">%s</a></li>`,
					serviceName,
				))
			}
		}
		result["services_html"] = strings.Join(servicesParts, "\n                ")
	} else {
		result["services"] = []string{}
		result["services_html"] = ""
	}

	// Add image URLs if present in ContentData
	if ctx.ContentData != nil {
		for _, field := range []string{"hero_url", "hero_home_url", "hero_about_url", "logo_url"} {
			if url, ok := ctx.ContentData[field].(string); ok && url != "" {
				result[field] = url
			}
		}
	}

	// =========================================================================
	// Merge ContentData fields for additional template access
	// This includes full service objects for {{range .services}} iteration
	// =========================================================================
	if ctx.ContentData != nil {
		for key, value := range ctx.ContentData {
			// Don't overwrite explicit fields
			if _, exists := result[key]; !exists {
				result[key] = value
			}
		}
	}

	return result
}

func mergeIntoRenderContext(ctx *RenderContext, data map[string]interface{}) {
	if v, ok := data["domain"].(string); ok && v != "" {
		ctx.Domain = v
	}
	if v, ok := data["company_name"].(string); ok && v != "" {
		ctx.CompanyName = v
		if ctx.LogoText == "" {
			ctx.LogoText = v
		}
	}
	if v, ok := data["logo_text"].(string); ok && v != "" {
		ctx.LogoText = v
	}
	if v, ok := data["tagline"].(string); ok && v != "" {
		ctx.Tagline = v
	}
	if v, ok := data["email"].(string); ok && v != "" {
		ctx.Email = v
	}
	if v, ok := data["phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["primary_color"].(string); ok && v != "" {
		ctx.PrimaryColor = v
	}
	if v, ok := data["secondary_color"].(string); ok && v != "" {
		ctx.SecondaryColor = v
	}
	if v, ok := data["accent_color"].(string); ok && v != "" {
		ctx.AccentColor = v
	}

	// Check nested color_palette
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok {
			ctx.AccentColor = v
		}
	}

	// Capture ALL fields into ContentData for template access
	if ctx.ContentData == nil {
		ctx.ContentData = make(map[string]interface{})
	}
	for key, value := range data {
		ctx.ContentData[key] = value
	}
}

// DEPRECATED
func convertNavigationItems(items []NavigationItem) []NavItem {
	result := make([]NavItem, len(items))
	for i, item := range items {
		result[i] = NavItem{
			Label: item.Label,
			URL:   item.URL,
		}
	}
	return result
}

// RenderComponentAction renders a single component template with context
// Config:
//   - component_function: function name to look up (e.g., "hero-banner")
//   - component_id: explicit component UUID (alternative to function)
//   - component_from: path to object containing function/id (e.g., "current_section")
//   - context_field: path to render context in collected_data
//   - content_field: path to additional content data
//   - content_from: alias for content_field (for consistency)
func RenderComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderComponentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required for component rendering")
	}

	// Get component - now with support for component_from indirection
	var comp *Component
	var err error
	var componentFunction string
	var componentID string

	// Priority 1: Direct component_id in config
	if compID, ok := config["component_id"].(string); ok && compID != "" {
		componentID = compID
	}

	// Priority 2: Direct component_function in config
	if compFunc, ok := config["component_function"].(string); ok && compFunc != "" {
		componentFunction = compFunc
	}

	// Priority 3: Extract from component_from field (indirection)
	if componentID == "" && componentFunction == "" {
		if componentFrom, ok := config["component_from"].(string); ok && componentFrom != "" {
			componentData := datahelpers.ExtractNestedField(params.CollectedData, componentFrom)
			if componentData != nil {
				// Case 1: component_from points to a map (e.g., "current_section")
				if compMap, ok := componentData.(map[string]interface{}); ok {
					// Try to get function first (most common)
					if fn, ok := compMap["function"].(string); ok && fn != "" {
						componentFunction = fn
						params.Logger.Debug("RenderComponentAction: Extracted function from component_from",
							zap.String("component_from", componentFrom),
							zap.String("function", componentFunction),
						)
					}
					// Also check for component_function key
					if fn, ok := compMap["component_function"].(string); ok && fn != "" {
						componentFunction = fn
					}
					// Check for id/component_id
					if id, ok := compMap["id"].(string); ok && id != "" {
						componentID = id
					}
					if id, ok := compMap["component_id"].(string); ok && id != "" {
						componentID = id
					}
					// Check for name as fallback (some components use name as function)
					if componentFunction == "" {
						if name, ok := compMap["name"].(string); ok && name != "" {
							componentFunction = name
							params.Logger.Debug("RenderComponentAction: Using name as function fallback",
								zap.String("name", name),
							)
						}
					}
				} else if compStr, ok := componentData.(string); ok && compStr != "" {
					// Case 2: component_from points directly to a string value
					// e.g., "current_section.function" resolves to "hero"
					componentFunction = compStr
					params.Logger.Debug("RenderComponentAction: Extracted function string directly from component_from",
						zap.String("component_from", componentFrom),
						zap.String("function", componentFunction),
					)
				}
			} else {
				params.Logger.Warn("RenderComponentAction: component_from field not found",
					zap.String("component_from", componentFrom),
				)
			}
		}
	}

	// Enforce naming contract: normalize to kebab-case before lookup
	if componentFunction != "" {
		normalized := NormalizeComponentFunction(componentFunction)
		if normalized != componentFunction {
			params.Logger.Info("RenderComponentAction: Normalized component function to kebab-case",
				zap.String("original", componentFunction),
				zap.String("normalized", normalized),
			)
			componentFunction = normalized
		}
	}

	// Now resolve the component
	if componentID != "" {
		compUUID, parseErr := uuid.Parse(componentID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid component_id: %w", parseErr)
		}
		comp, err = GetComponentByID(ctx, params.DB, compUUID, params.Logger)
	} else if componentFunction != "" {
		comp, err = GetComponentWithFallback(ctx, params.DB, componentFunction, params.Logger)
	} else {
		// Log available info for debugging
		params.Logger.Error("RenderComponentAction: No component identifier found",
			zap.Any("config_keys", datahelpers.GetMapKeys(config)),
			zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		)
		return nil, fmt.Errorf("component_function, component_id, or component_from required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get component '%s': %w", componentFunction, err)
	}

	// Get render context
	contextField := "render_context"
	if cf, ok := config["context_field"].(string); ok && cf != "" {
		contextField = cf
	}
	// Also support context_from as an alias
	if cf, ok := config["context_from"].(string); ok && cf != "" {
		contextField = cf
	}

	renderCtxData := datahelpers.ExtractNestedField(params.CollectedData, contextField)
	renderCtx := &RenderContext{}
	if m, ok := renderCtxData.(map[string]interface{}); ok {
		mergeIntoRenderContext(renderCtx, m)
	}

	// Merge additional content if specified (support both content_field and content_from)
	contentField := ""
	if cf, ok := config["content_field"].(string); ok && cf != "" {
		contentField = cf
	}
	if cf, ok := config["content_from"].(string); ok && cf != "" {
		contentField = cf
	}

	var sectionContentData map[string]interface{}

	if contentField != "" {
		// Try to extract content with fallback paths
		// LLM responses sometimes have .result wrapper, sometimes not
		contentData := extractContentWithFallbacks(params.CollectedData, contentField, params.Logger)
		if contentData != nil {
			// Safety net: reconcile any array-item keys the LLM invented
			// (e.g. title/body) against the keys the component template reads
			// (e.g. name/description), before the content reaches the template
			// or is persisted. Expected keys are sourced from the component's
			// own input_schema (reloaded fresh above from the component store),
			// which is the authoritative contract the html_template is built
			// against — so reconciliation does not depend on section-plan
			// freshness or on the prompt. Scoped to source:"llm" array fields so
			// reach matches the writer loop; a no-op on the template-only path
			// (content is render_context, whose keys don't match the schema's
			// llm array fields).
			if comp != nil && len(comp.InputSchema) > 0 {
				reconcileGeneratedItemKeys(contentData, expectedItemFieldsFromComponentSchema(comp.InputSchema), componentFunction, params.Logger)
			}
			sectionContentData = contentData // ← capture before merge
			params.Logger.Info("RenderComponentAction: Merging content data",
				zap.String("content_field", contentField),
				zap.Int("field_count", len(contentData)),
				zap.Any("keys", datahelpers.GetMapKeys(contentData)))
			mergeIntoRenderContext(renderCtx, contentData)
		} else {
			params.Logger.Warn("RenderComponentAction: No content data found at any path",
				zap.String("content_field", contentField),
				zap.Any("available_top_keys", datahelpers.GetMapKeys(params.CollectedData)))
		}
	}

	// Step 3: optional merge_with — overlay pre-resolved data on top of the
	// LLM/content output. Used by page-content-writer's loop with
	// `merge_with: current_section.resolved_data` so query-resolved items,
	// static fallback values, and other authoritative data land in both the
	// rendered HTML AND the persisted content_data. The merge happens AFTER
	// the content_from block so resolved_data wins on conflicts — by design,
	// because it's database-derived and authoritative; the LLM should never
	// be writing items/urls/labels that the resolver already produced.
	if mw, ok := config["merge_with"].(string); ok && mw != "" {
		mergeData := datahelpers.ExtractNestedField(params.CollectedData, mw)
		if mergeMap, ok := mergeData.(map[string]interface{}); ok && len(mergeMap) > 0 {
			params.Logger.Info("RenderComponentAction: Merging resolved data",
				zap.String("merge_with", mw),
				zap.Int("merge_field_count", len(mergeMap)),
				zap.Any("merge_keys", datahelpers.GetMapKeys(mergeMap)))
			if sectionContentData == nil {
				sectionContentData = make(map[string]interface{})
			}
			// Overlay merge data onto section content data so it lands in both
			// the render context AND the persisted content_data output.
			// Last write wins → resolved_data overrides LLM duplicates.
			for k, v := range mergeMap {
				sectionContentData[k] = v
			}
			mergeIntoRenderContext(renderCtx, mergeMap)
		} else if mergeData != nil {
			params.Logger.Warn("RenderComponentAction: merge_with did not resolve to a map",
				zap.String("merge_with", mw),
				zap.String("type", fmt.Sprintf("%T", mergeData)))
		}
		// If mergeData is nil, that's fine — the path simply wasn't populated
		// for this section (e.g. an all-LLM section with no resolved_data).
	}

	// Before calling RenderTemplate, set ComponentID in context
	renderCtx.ContentData["ComponentID"] = comp.ID

	// Fail loud rather than ship a silently-empty section. If the component's
	// schema marks a content field required (source:"llm") and it never arrived
	// — the LLM was truncated, or its response was unparseable and fell back to
	// a raw-text envelope that carries no such field — then missingkey=zero
	// would render {{.field}} as empty, page assembly would drop the visually
	// empty section, and the article would silently vanish. That is exactly the
	// mechanism that blanked 9 live article bodies. Refuse the render instead so
	// the step fails and the good content is left in place (the content
	// regression guard in save_page_sections blocks the overwrite).
	if len(comp.InputSchema) > 0 {
		datahelpers.WarnIfLegacyDialect(comp.InputSchema, params.Logger, "render-gate", comp.Function)
		if missing := missingRequiredLLMFields(comp.InputSchema, renderCtx.ContentData); len(missing) > 0 {
			params.Logger.Error("RenderComponentAction: required content field(s) missing — refusing to render an empty section",
				zap.String("component_function", comp.Function),
				zap.String("component_name", comp.Name),
				zap.Strings("missing_fields", missing),
			)
			return nil, fmt.Errorf(
				"component %q is missing required content field(s) %v — refusing to render an empty section "+
					"(likely LLM truncation or an unparseable response); leaving existing content untouched",
				comp.Function, missing)
		}
	}

	// Render template
	rendered := RenderTemplate(comp.HTMLTemplate, renderCtx, params.Logger)

	// Dark section contract validation (warning only, non-blocking)
	// Uses is_dark_section from DB when available, falls back to CSS auto-detection.
	// See validate_dark_section.go and 014_section_context_contract.md.
	if missing := ValidateDarkSectionContract(rendered, comp.IsDarkSection, params.Logger); len(missing) > 0 {
		params.Logger.Warn("RenderComponentAction: Dark section missing --section-* variables",
			zap.String("component_function", comp.Function),
			zap.Bool("is_dark_section", comp.IsDarkSection),
			zap.Strings("missing_vars", missing),
		)
	}

	params.Logger.Info("RenderComponentAction: Component rendered",
		zap.String("component", comp.Name),
		zap.String("function", comp.Function),
		zap.Int("output_length", len(rendered)),
	)

	result := map[string]interface{}{
		"rendered_html":      rendered,
		"component_id":       comp.ID,
		"component_name":     comp.Name,
		"component_function": comp.Function,
	}
	if sectionContentData != nil {
		result["content_data"] = sectionContentData
	}
	return result, nil
}

// ============================================================================
// ACTION: compile_page_sections
// ============================================================================

// CompilePageSectionsAction combines multiple rendered sections into a page
// Config:
//   - sections_from OR sections_field: path to array of section results
//   - page_from: path to page data for name/title
//   - page_name: explicit name for the page (fallback)
//   - inject_header: boolean
//   - inject_footer: boolean
func CompilePageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompilePageSectionsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Accept both "sections_from" and "sections_field" for compatibility
	sectionsField := "rendered_sections"
	if sf, ok := config["sections_from"].(string); ok && sf != "" {
		sectionsField = sf
	} else if sf, ok := config["sections_field"].(string); ok && sf != "" {
		sectionsField = sf
	}

	params.Logger.Info("CompilePageSectionsAction: Looking for sections",
		zap.String("sections_field", sectionsField))

	sectionsData := datahelpers.ExtractNestedField(params.CollectedData, sectionsField)
	if sectionsData == nil {
		// Try with .results suffix (loop action output format)
		sectionsData = datahelpers.ExtractNestedField(params.CollectedData, sectionsField+".results")
		if sectionsData != nil {
			params.Logger.Info("CompilePageSectionsAction: Found sections at .results path")
		}
	}

	if sectionsData == nil {
		params.Logger.Error("CompilePageSectionsAction: Sections not found",
			zap.String("tried_path", sectionsField),
			zap.String("also_tried", sectionsField+".results"),
			zap.Strings("available_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return nil, fmt.Errorf("sections not found at %s", sectionsField)
	}

	var sections []string
	var sectionsMetadata []map[string]interface{}

	switch v := sectionsData.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				sections = append(sections, s)
				// String-only item: no metadata available
				sectionsMetadata = append(sectionsMetadata, map[string]interface{}{
					"rendered_html": s,
				})
			} else if m, ok := item.(map[string]interface{}); ok {
				html, meta := extractSectionFromMap(m, params.Logger)
				if html != "" {
					sections = append(sections, html)
					sectionsMetadata = append(sectionsMetadata, meta)
				}
			}
		}
	case map[string]interface{}:
		// Check if this is a loop output with "results" array
		if results, ok := v["results"].([]interface{}); ok {
			for _, item := range results {
				if m, ok := item.(map[string]interface{}); ok {
					html, meta := extractSectionFromMap(m, params.Logger)
					if html != "" {
						sections = append(sections, html)
						sectionsMetadata = append(sectionsMetadata, meta)
					}
				}
			}
		} else {
			// Ordered by keys (section_0, section_1, etc.)
			for i := 0; i < len(v); i++ {
				key := fmt.Sprintf("section_%d", i)
				if section, ok := v[key]; ok {
					if s, ok := section.(string); ok {
						sections = append(sections, s)
						sectionsMetadata = append(sectionsMetadata, map[string]interface{}{
							"rendered_html": s,
						})
					} else if m, ok := section.(map[string]interface{}); ok {
						html, meta := extractSectionFromMap(m, params.Logger)
						if html != "" {
							sections = append(sections, html)
							sectionsMetadata = append(sectionsMetadata, meta)
						}
					}
				}
			}
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Extracted sections",
		zap.Int("count", len(sections)))

	if len(sections) == 0 {
		params.Logger.Warn("CompilePageSectionsAction: No sections to compile, returning placeholder")

		// Get page info for context
		pageName := "page"
		if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
			if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
				if pm, ok := pageData.(map[string]interface{}); ok {
					if name, ok := pm["name"].(string); ok && name != "" {
						pageName = name
					}
				}
			}
		}

		return map[string]interface{}{
			"page_body":     "",
			"page_name":     pageName,
			"section_count": 0,
			"skipped":       true,
			"reason":        "no sections defined for page",
		}, nil
	}

	// Build page body
	pageBody := strings.Join(sections, "\n\n")

	// Get page name - try page_from first, then page_name config
	pageName := "index"
	if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
		if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
			if pm, ok := pageData.(map[string]interface{}); ok {
				if name, ok := pm["name"].(string); ok && name != "" {
					pageName = name
				}
			}
		}
	}
	if pn, ok := config["page_name"].(string); ok && pn != "" {
		pageName = pn
	}

	// Build full HTML page
	pageHTML := buildPageHTML(pageName, pageBody)

	// Optionally inject head/header/footer from component library.
	// Head is injected here (same time as header/footer) rather than deferred
	// to a later assemble_page step — deferring caused <head> to end up inside
	// <body> when cleanHTMLStructure's dedup logic picked the wrong block.
	injectHead, _ := config["inject_head"].(bool)
	injectHeader, _ := config["inject_header"].(bool)
	injectFooter, _ := config["inject_footer"].(bool)

	if params.DB != nil && (injectHead || injectHeader || injectFooter) {
		// Get site_id for component lookup
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		siteUUID := uuid.Nil
		if siteIDStr != "" {
			siteUUID, _ = uuid.Parse(siteIDStr)
		}

		renderCtx := &RenderContext{}
		if rc := datahelpers.ExtractNestedField(params.CollectedData, "render_context"); rc != nil {
			if m, ok := rc.(map[string]interface{}); ok {
				mergeIntoRenderContext(renderCtx, m)
			}
		}

		// Ensure page-specific title/description are set for head component.
		// The head template uses {{.title}} and {{.description}} — these must
		// reflect the current page, not just the site-level defaults.
		if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
			if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
				if pm, ok := pageData.(map[string]interface{}); ok {
					if t, ok := pm["title"].(string); ok && t != "" {
						renderCtx.Title = t
					} else if n, ok := pm["name"].(string); ok && n != "" && renderCtx.Title == "" {
						renderCtx.Title = strings.Title(strings.ReplaceAll(n, "-", " "))
					}
					if d, ok := pm["meta_description"].(string); ok && d != "" {
						renderCtx.Description = d
					}
				}
			}
		}

		if injectHead {
			pageHTML = InjectHead(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
		if injectHeader {
			pageHTML = InjectHeader(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
		if injectFooter {
			pageHTML = InjectFooter(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Page compiled",
		zap.String("page_name", pageName),
		zap.Int("section_count", len(sections)),
		zap.Int("html_length", len(pageHTML)),
	)

	return map[string]interface{}{
		"page_html":         pageHTML,
		"page_name":         pageName,
		"section_count":     len(sections),
		"sections_metadata": sectionsMetadata,
	}, nil
}

// extractSectionFromMap reads HTML and component metadata from a section result map.
// It handles two shapes:
//
//  1. Direct RenderComponentAction output: metadata lives at top level alongside
//     rendered_html. This is what CompilePageSectionsAction used to assume.
//
//  2. Loop-wrapped output: LoopAction's completion step (Strategy 1,
//     substep_output_fields) promotes HTML to top-level rendered_html/page_html,
//     but component_id/component_name/component_function stay nested inside the
//     substep output key (section_output, render_section, render_from_template).
//     This is what content-writer's process_sections_loop produces in practice.
//
// Lookup order: top-level first, then nested substep keys (earliest-wins). This
// preserves the historical behaviour for callers with the flat shape while
// recovering metadata from the nested shape — without which save_page_sections
// ends up with ComponentName = "section" (the default from extractSectionsFromMetadata)
// and enrichSectionsWithComponentIDs skips every section, leaving page_components.component_id NULL.
//
// Returned meta always contains rendered_html. When available, also:
// component_id, component_name, component_function, content_data.
// Returns ("", nil) if no HTML could be found in any known position.
func extractSectionFromMap(m map[string]interface{}, logger *zap.Logger) (string, map[string]interface{}) {
	// Extract HTML — check top-level first, then common substep keys.
	html := extractHTMLFromSectionMap(m)
	if html == "" {
		return "", nil
	}

	meta := map[string]interface{}{
		"rendered_html": html,
	}

	// Collect component metadata from top level first.
	if id, ok := m["component_id"]; ok && id != nil {
		meta["component_id"] = fmt.Sprintf("%v", id)
	}
	if name, ok := m["component_name"].(string); ok && name != "" {
		meta["component_name"] = name
	}
	if fn, ok := m["component_function"].(string); ok && fn != "" {
		meta["component_function"] = fn
	}
	if cd, ok := m["content_data"]; ok && cd != nil {
		meta["content_data"] = cd
	}

	// Remember whether top-level already had the name, so we only log recovery
	// when the nested fallback actually contributed it.
	_, hadTopName := m["component_name"].(string)

	if meta["component_id"] == nil || meta["component_name"] == nil ||
		meta["component_function"] == nil || meta["content_data"] == nil {

		for _, subKey := range []string{"section_output", "render_section", "render_from_template"} {
			nested, ok := m[subKey].(map[string]interface{})
			if !ok {
				continue
			}
			if meta["component_id"] == nil {
				if id, ok := nested["component_id"]; ok && id != nil {
					meta["component_id"] = fmt.Sprintf("%v", id)
				}
			}
			if meta["component_name"] == nil {
				if name, ok := nested["component_name"].(string); ok && name != "" {
					meta["component_name"] = name
				}
			}
			if meta["component_function"] == nil {
				if fn, ok := nested["component_function"].(string); ok && fn != "" {
					meta["component_function"] = fn
				}
			}
			if meta["content_data"] == nil {
				if cd, ok := nested["content_data"]; ok && cd != nil {
					meta["content_data"] = cd
				}
			}
			if meta["component_id"] != nil && meta["component_name"] != nil &&
				meta["component_function"] != nil && meta["content_data"] != nil {
				break
			}
		}

		// Signal that the nested-lookup fallback fired — the primary diagnostic
		// signal that this fix is taking effect in production logs.
		if !hadTopName {
			if n, ok := meta["component_name"].(string); ok && n != "" {
				logger.Info("CompilePageSectionsAction: recovered component_name from nested substep output",
					zap.String("component_name", n))
			}
		}
	}

	return html, meta
}

// extractHTMLFromSectionMap pulls the rendered HTML string from a section-result
// map, checking top-level keys first, then common substep output keys where
// LoopAction may have nested the RenderComponentAction result.
func extractHTMLFromSectionMap(m map[string]interface{}) string {
	if h, ok := m["rendered_html"].(string); ok && h != "" {
		return h
	}
	if h, ok := m["page_html"].(string); ok && h != "" {
		return h
	}
	if h, ok := m["html"].(string); ok && h != "" {
		return h
	}
	// Also try nested substep keys (loop-wrapped shape where top-level HTML
	// promotion didn't happen for some reason).
	for _, subKey := range []string{"section_output", "render_section", "render_from_template"} {
		if nested, ok := m[subKey].(map[string]interface{}); ok {
			if h, ok := nested["rendered_html"].(string); ok && h != "" {
				return h
			}
			if h, ok := nested["page_html"].(string); ok && h != "" {
				return h
			}
		}
	}
	return ""
}

func buildPageHTML(pageName, body string) string {
	title := strings.Title(strings.ReplaceAll(pageName, "-", " "))
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body>
%s
</body>
</html>`, title, body)
}

// ============================================================================
// ACTION: insert_research_result
// ============================================================================

// InsertResearchResultAction stores research findings in the database
// Config:
//   - table: target table name (default: "research_results")
//   - fields: map of column_name -> data_path for dynamic field mapping
//   - site_id_field: path to site_id (fallback if not in fields)
//   - result_type: type of research (fallback if not in fields)
func InsertResearchResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("InsertResearchResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("InsertResearchResultAction: No database")
		return map[string]interface{}{"inserted": false, "reason": "no database"}, nil
	}

	// Get table name (default to research_results)
	tableName := "research_results"
	if tn, ok := config["table"].(string); ok && tn != "" {
		tableName = tn
	}

	// Get field mappings from config
	fieldMappings, hasFields := config["fields"].(map[string]interface{})

	// Generate new ID
	resultID := uuid.New()

	// Build dynamic INSERT based on field mappings
	if hasFields && len(fieldMappings) > 0 {
		// Dynamic field-based insert
		columns := []string{"id"}
		placeholders := []string{"$1"}
		values := []interface{}{resultID}
		paramIdx := 2

		for column, dataPath := range fieldMappings {
			dataPathStr, ok := dataPath.(string)
			if !ok {
				continue
			}

			// Extract the value from collected data
			value := datahelpers.ExtractNestedField(params.CollectedData, dataPathStr)

			// Special handling for site_id (needs UUID conversion)
			if column == "site_id" {
				if siteIDStr, ok := value.(string); ok && siteIDStr != "" {
					if siteUUID, err := uuid.Parse(siteIDStr); err == nil {
						value = siteUUID
					} else {
						value = nil
					}
				} else {
					value = nil
				}
			}

			// Skip nil values for optional fields, but include empty strings
			if value == nil {
				continue
			}

			// For complex types, marshal to JSON
			/*switch v := value.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
						zap.String("column", column),
						zap.Error(err))
					continue
				}
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
				values = append(values, string(jsonBytes))
			default:
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d", paramIdx))
				values = append(values, value)
			}*/

			// make it all json
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
					zap.String("column", column),
					zap.String("value_type", fmt.Sprintf("%T", value)),
					zap.Error(err))
				continue
			}
			columns = append(columns, column)
			placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
			values = append(values, string(jsonBytes))
			paramIdx++
		}

		// Add created_at
		columns = append(columns, "created_at")
		placeholders = append(placeholders, "NOW()")

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)

		params.Logger.Info("InsertResearchResultAction: Executing dynamic insert",
			zap.String("table", tableName),
			zap.Strings("columns", columns),
			zap.Int("value_count", len(values)))

		if err := execDB(ctx, params.DB, query, values...); err != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed",
				zap.String("query", query),
				zap.Error(err))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": "general",
				"error":       err.Error(),
			}, nil
		}

		return map[string]interface{}{
			"inserted":  true,
			"id":        resultID.String(),
			"result_id": resultID.String(),
			"table":     tableName,
			"columns":   columns,
		}, nil
	}

	// Fallback: Legacy mode with hardcoded columns
	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	siteID := uuid.Nil
	if siteIDStr != "" {
		siteID, _ = uuid.Parse(siteIDStr)
	}

	// Get result type
	resultType := "general"
	if rt, ok := config["result_type"].(string); ok && rt != "" {
		resultType = rt
	}

	// Get research data - try multiple field paths
	dataField := "research_result"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	researchData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// If no data found, try to build from synthesis
	if researchData == nil {
		researchData = map[string]interface{}{
			"summary":  datahelpers.ExtractNestedField(params.CollectedData, "synthesis.summary"),
			"findings": datahelpers.ExtractNestedField(params.CollectedData, "synthesis"),
			"query":    datahelpers.ExtractNestedField(params.CollectedData, "search_query.result"),
			"topic":    datahelpers.ExtractNestedField(params.CollectedData, "extracted.topic"),
		}
	}

	dataJSON, err := json.Marshal(researchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal research data: %w", err)
	}

	// Try insert with 'findings' column first (new schema), fall back to 'data' (old schema)
	var siteIDArg interface{} = nil
	if siteID != uuid.Nil {
		siteIDArg = siteID
	}

	// Try new schema first
	query := `
		INSERT INTO research_results (id, site_id, result_type, findings, created_at)
		VALUES ($1, $2, $3, $4::jsonb, NOW())
	`

	if err := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err != nil {
		// Try with 'data' column (old schema)
		query = `
			INSERT INTO research_results (id, site_id, result_type, data, created_at)
			VALUES ($1, $2, $3, $4::jsonb, NOW())
		`
		if err2 := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err2 != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed with both schemas",
				zap.Error(err),
				zap.Error(err2))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": resultType,
				"error":       err2.Error(),
			}, nil
		}
	}

	return map[string]interface{}{
		"inserted":    true,
		"result_id":   resultID.String(),
		"result_type": resultType,
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// ACTION: store_asset
// ============================================================================

// StoreAssetAction stores an asset (image, font, file) in the assets table
// This is a LOCAL action - it does NOT require a topic
// Config:
//   - asset_type: type of asset (image, font, css, js, etc.)
//   - site_id_field: path to site_id in collected_data
//   - data_field: path to asset data (URL, base64, or content)
//   - name_field: path to asset name
//   - metadata_field: optional path to additional metadata
func StoreAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StoreAssetAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get asset type
	assetType := "image"
	if at, ok := config["asset_type"].(string); ok && at != "" {
		assetType = at
	}

	// Get purpose from config — supports literal OR field lookup.
	// Phase 2H: purpose_field added so the new needs_imagery branch can pass
	// spec.purpose through dynamically (logo, hero, illustration, icon,
	// infographic) without hardcoding in workflow step config. Mirrors the
	// asset_key / asset_key_field pattern below.
	//
	// Resolution priority:
	//   1. config["purpose"]        — literal string (e.g. "hero", "logo")
	//   2. config["purpose_field"]  — JSONPath into collected_data
	//   3. ""                       — empty; downstream asset_key resolution
	//                                 may still backfill via asset_key_field
	purpose := ""
	if p, ok := config["purpose"].(string); ok && p != "" {
		purpose = p
	}
	if purpose == "" {
		if pf, ok := config["purpose_field"].(string); ok && pf != "" {
			purpose = datahelpers.ExtractNestedFieldString(params.CollectedData, pf)
		}
	}

	// Phase 2C: extract asset_key from config, defaulting to purpose.
	// Phase 2E: also support asset_key_field for JSONPath lookup so
	// per-item variants can be passed through the workflow without
	// hardcoded literals.
	//
	// Resolution priority:
	//   1. config["asset_key"]        — literal string (e.g. "logo")
	//   2. config["asset_key_field"]  — JSONPath into collected_data
	//   3. purpose (default — Phase 2C backward-compat)
	assetKey := ""
	if k, ok := config["asset_key"].(string); ok && k != "" {
		assetKey = k
	}
	if assetKey == "" {
		if kf, ok := config["asset_key_field"].(string); ok && kf != "" {
			assetKey = datahelpers.ExtractNestedFieldString(params.CollectedData, kf)
		}
	}
	if assetKey == "" {
		assetKey = purpose
	}

	// Get site_id (optional - assets can be global)
	var siteID *uuid.UUID
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = &parsed
			}
		}
	}

	// Get asset data
	dataField := "asset_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	assetData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// Get asset name
	nameField := "asset_name"
	if nf, ok := config["name_field"].(string); ok && nf != "" {
		nameField = nf
	}
	assetName := datahelpers.ExtractNestedFieldString(params.CollectedData, nameField)
	if assetName == "" {
		assetName = fmt.Sprintf("%s_%s", assetType, uuid.New().String()[:8])
	}

	// Extract URL or content from asset data
	var assetURL string
	switch v := assetData.(type) {
	case string:
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "s3://") {
			assetURL = v
		}
	case map[string]interface{}:
		if url, ok := v["url"].(string); ok {
			assetURL = url
		}
		if url, ok := v["image_url"].(string); ok {
			assetURL = url
		}
	}

	if assetURL == "" {
		params.Logger.Warn("StoreAssetAction: No asset URL found")
		return map[string]interface{}{
			"stored":     false,
			"asset_name": assetName,
			"asset_type": assetType,
			"reason":     "no asset URL found",
		}, nil
	}

	// If no DB, return success without persisting
	if params.DB == nil {
		params.Logger.Warn("StoreAssetAction: No database, returning without persistence")
		return map[string]interface{}{
			"stored":     true,
			"persisted":  false,
			"asset_id":   uuid.New().String(),
			"asset_name": assetName,
			"asset_type": assetType,
			"asset_url":  assetURL,
		}, nil
	}

	// Determine origin_type based on URL
	originType := "uploaded"
	if strings.HasPrefix(assetURL, "s3://") || strings.Contains(assetURL, "backblazeb2.com") {
		originType = "generated"
	}

	// Phase 0.2: extract origin_prompt and origin_model from step config.
	// Bug fix — origin_prompt_field has been silently passed by workflows
	// without this action ever reading it. After this lands, new generations
	// populate the column. New — origin_model accepts a literal (used today,
	// e.g. "sdxl") or a path (origin_model_field) for future provider routing.
	// Literal wins if both are set.
	originPrompt := ""
	if pf, ok := config["origin_prompt_field"].(string); ok && pf != "" {
		originPrompt = datahelpers.ExtractNestedFieldString(params.CollectedData, pf)
	}
	originModel := ""
	if m, ok := config["origin_model"].(string); ok && m != "" {
		originModel = m
	} else if mf, ok := config["origin_model_field"].(string); ok && mf != "" {
		originModel = datahelpers.ExtractNestedFieldString(params.CollectedData, mf)
	}

	// Insert into assets table - matches actual schema
	assetID := uuid.New()

	query := `
		INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url, origin_type,
		                    origin_prompt, origin_model, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active' DO UPDATE SET
			purpose = EXCLUDED.purpose,
			url = EXCLUDED.url,
			name = EXCLUDED.name,
			origin_type = EXCLUDED.origin_type,
			origin_prompt = COALESCE(EXCLUDED.origin_prompt, assets.origin_prompt),
			origin_model = COALESCE(EXCLUDED.origin_model, assets.origin_model),
			updated_at = NOW()
		WHERE assets.locked_at IS NULL
		RETURNING id
	`

	var returnedID uuid.UUID
	err := queryRowScanUUID(ctx, params.DB, query, &returnedID,
		assetID, siteID, assetName, assetType, nullString(purpose),
		nullString(assetKey), assetURL, originType,
		nullString(originPrompt), nullString(originModel))

	// Phase I1 (D5, logo permanence): the conflict target matched a LOCKED
	// asset — the DO UPDATE ... WHERE suppressed the write and RETURNING
	// produced no row. Approved assets (locked_at set, e.g. an
	// approve-and-locked logo) must never be silently replaced by a fresh
	// generation. Report it as a refusal, not an error, so callers complete.
	if err != nil && strings.Contains(err.Error(), "no rows") {
		params.Logger.Warn("StoreAssetAction: target asset is LOCKED — refusing to overwrite",
			zap.String("asset_key", assetKey),
			zap.String("purpose", purpose))
		return map[string]interface{}{
			"stored":     false,
			"locked":     true,
			"asset_key":  assetKey,
			"asset_name": assetName,
			"reason":     "asset is locked (locked_at set) — approved assets are never overwritten",
		}, nil
	}

	if err != nil {
		// Try simpler insert without upsert if constraint doesn't exist
		params.Logger.Warn("StoreAssetAction: Upsert failed, trying simple insert",
			zap.Error(err))

		simpleQuery := `
			INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url, origin_type,
			                    origin_prompt, origin_model, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
			RETURNING id
		`
		err = queryRowScanUUID(ctx, params.DB, simpleQuery, &returnedID,
			assetID, siteID, assetName, assetType, nullString(purpose),
			nullString(assetKey), assetURL, originType,
			nullString(originPrompt), nullString(originModel))

		if err != nil {
			params.Logger.Warn("StoreAssetAction: Insert failed",
				zap.Error(err))
			return map[string]interface{}{
				"stored":     true,
				"persisted":  false,
				"asset_id":   assetID.String(),
				"asset_name": assetName,
				"asset_type": assetType,
				"asset_url":  assetURL,
				"error":      err.Error(),
			}, nil
		}
	}

	// If purpose is set and we have a site_id, update sites.content_data
	// This stores the storage URI (for download) and relative URL (for templates)
	storageURI := ""
	if purpose != "" && siteID != nil {
		// Find storage URI from asset data
		if assetDataMap, ok := assetData.(map[string]interface{}); ok {
			if uri, ok := assetDataMap["image_uri"].(string); ok {
				storageURI = uri
			}
		}
		// Also check collected_data for {purpose}_result.image_uri pattern
		if storageURI == "" {
			uriField := strings.TrimSuffix(dataField, ".image_url") + ".image_uri"
			if uri := datahelpers.ExtractNestedFieldString(params.CollectedData, uriField); uri != "" {
				storageURI = uri
			}
		}
		// Use assetURL if it's a storage URI
		if storageURI == "" && storage.IsS3URI(assetURL) {
			storageURI = assetURL
		}

		// Generate paths using storage package helper (use correct extension for purpose)
		_, _, _, purposeExt := storage.GetImageConfig(purpose)
		paths := storage.BuildAssetPaths(purpose, purposeExt)

		// Update sites.content_data
		if storageURI != "" {
			// Store URI for deploy_image_asset to download from
			updateContentDataField(ctx, params.DB, *siteID, purpose+"_uri", storageURI, params.Logger)
			params.CollectedData[purpose+"_uri"] = storageURI
		}

		// Store relative URL for templates
		updateContentDataField(ctx, params.DB, *siteID, purpose+"_url", paths.RelativeURL, params.Logger)
		params.CollectedData[purpose+"_url"] = paths.RelativeURL

		params.Logger.Info("StoreAssetAction: Updated content_data for purpose",
			zap.String("purpose", purpose),
			zap.String("storage_uri", storageURI),
			zap.String("relative_url", paths.RelativeURL))
	}

	params.Logger.Info("StoreAssetAction: Asset stored",
		zap.String("asset_id", returnedID.String()),
		zap.String("asset_name", assetName),
		zap.String("asset_type", assetType),
	)

	result := map[string]interface{}{
		"stored":     true,
		"persisted":  true,
		"asset_id":   returnedID.String(),
		"asset_name": assetName,
		"asset_type": assetType,
		"asset_url":  assetURL,
	}

	// Add purpose-specific fields if set
	if purpose != "" {
		result["purpose"] = purpose
		_, _, _, purposeExt := storage.GetImageConfig(purpose)
		paths := storage.BuildAssetPaths(purpose, purposeExt)
		result[purpose+"_url"] = paths.RelativeURL
	}

	// Add storage URI to result for downstream deploy step
	if storageURI != "" {
		result["image_uri"] = storageURI
		result["s3_uri"] = storageURI
	}

	return result, nil
}

// updateContentDataField updates a single field in sites.content_data
func updateContentDataField(ctx context.Context, db interface{}, siteID uuid.UUID, field, value string, logger *zap.Logger) {
	query := `
        UPDATE sites 
        SET content_data = jsonb_set(
            COALESCE(content_data, '{}'::jsonb),
            $2::text[],
            to_jsonb($3::text),
            true
        ),
        updated_at = NOW()
        WHERE id = $1
    `
	jsonPath := fmt.Sprintf("{%s}", field)

	if err := execDB(ctx, db, query, siteID, jsonPath, value); err != nil {
		logger.Warn("Failed to update content_data field",
			zap.String("field", field),
			zap.Error(err))
	} else {
		logger.Debug("Updated content_data field",
			zap.String("field", field),
			zap.String("value", value))
	}
}

func ValidateSitePlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ValidateSitePlanAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	planField := "llm_plan"
	if pf, ok := config["plan_field"].(string); ok && pf != "" {
		planField = pf
	}

	// Extract the plan data
	planData := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planData == nil {
		return nil, fmt.Errorf("plan not found at '%s'", planField)
	}

	planData = datahelpers.UnwrapDeep(planData, params.Logger)

	params.Logger.Info("ValidateSitePlanAction: After UnwrapDeep",
		zap.String("planData_type", fmt.Sprintf("%T", planData)))

	var plan map[string]interface{}
	switch v := planData.(type) {
	case map[string]interface{}:
		plan = v
	case string:
		if v == "" {
			return nil, fmt.Errorf("plan is empty string - LLM may have returned no content. Check template rendering logs for <no value> placeholders")
		}
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			return nil, fmt.Errorf("plan is empty after cleaning markdown")
		}
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			// Include content preview in error for debugging
			preview := cleaned
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return nil, fmt.Errorf("failed to parse plan JSON: %w (preview: %s)", err, preview)
		}
	default:
		return nil, fmt.Errorf("plan must be object or JSON string, got %T", planData)
	}

	pagesRaw, ok := plan["pages"]
	if !ok {
		// Log available keys for debugging
		keys := make([]string, 0, len(plan))
		for k := range plan {
			keys = append(keys, k)
		}
		params.Logger.Error("ValidateSitePlanAction: plan missing 'pages'",
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("plan must have 'pages' array (available keys: %v)", keys)
	}

	pages, ok := pagesRaw.([]interface{})
	if !ok || len(pages) == 0 {
		return nil, fmt.Errorf("pages must be non-empty array")
	}

	// ── Deterministic convergence with realised pages ───────────────────────
	// existing_pages is loaded by the load_existing_pages workflow step and
	// carries site_has_no_current_plan and build_status per page. reconcilePlanWithRealised
	// force-preserves every page on the site's first plan OR built, so a re-plan
	// can no longer silently redesign or drop a built page (bugs_open/001). It is
	// a no-op only for a genuinely from-scratch build. See
	// FOCUS_adoption_faithfulness_via_locks.md.
	existingField := "existing_pages"
	if ef, ok := config["existing_pages_field"].(string); ok && ef != "" {
		existingField = ef
	}
	var existingPages []interface{}
	if ev := datahelpers.ExtractNestedField(params.CollectedData, existingField); ev != nil {
		switch vv := ev.(type) {
		case []interface{}:
			existingPages = vv
		case []map[string]interface{}:
			// query_database (output_format=array) returns []map[string]interface{},
			// which does NOT satisfy a []interface{} assertion in Go. Convert it so
			// the convergence actually sees the realised pages. Without this the
			// assertion silently fails, existingPages stays empty, and
			// reconcilePlanWithRealised no-ops for every site (adopted pages never
			// preserved, planner siblings never dropped).
			existingPages = make([]interface{}, len(vv))
			for i := range vv {
				existingPages[i] = vv[i]
			}
		}
	}
	// Explicit redesign intent (bugs_open/037 fix step 4 / features_open/012).
	// Pages named in the trigger spec's optional `recompose_pages` list are
	// RELEASED from the preserve guard for THIS re-plan only, so the LLM's proposed
	// composition governs them (a page may be recomposed or, if the LLM omits it,
	// dropped). This is the sanctioned way to deliberately redesign a preserved
	// (deployed / needs_rebuild) page — without an entry here the guard preserves
	// it. Filtering the realised set HERE, before both reconcilePlanWithRealised
	// and the truncation must-keep read `existingPages`, makes a recompose page
	// uniformly from-scratch. Ordinary re-plans carry no such field and are
	// unaffected.
	if recompose := recomposePagesFromSpec(params.CollectedData, params.Logger); len(recompose) > 0 {
		existingPages = filterOutRecomposePages(existingPages, recompose, params.Logger)
	}

	// Surface the convergence input size so an empty set is never silent again.
	params.Logger.Info("ValidateSitePlanAction: existing pages loaded for convergence",
		zap.Int("existing_pages", len(existingPages)),
		zap.String("existing_pages_field", existingField))
	var unionedIn, droppedCollision, snappedRename, snappedSections int
	pages, unionedIn, droppedCollision, snappedRename, snappedSections =
		reconcilePlanWithRealised(pages, existingPages, params.Logger)
	plan["pages"] = pages
	params.Logger.Info("ValidateSitePlanAction: reconciled with realised pages",
		zap.Int("unioned_in", unionedIn),
		zap.Int("dropped_collision", droppedCollision),
		zap.Int("snapped_rename", snappedRename),
		zap.Int("snapped_sections", snappedSections),
		zap.Int("pages_after", len(pages)))

	// ── Truncate, preserving first-plan AND built pages ─────────────────────
	maxPages := 20
	if mp, ok := config["max_pages"].(float64); ok {
		maxPages = int(mp)
	}
	if len(pages) > maxPages {
		// Must-keep mirrors reconcilePlanWithRealised's preservation set: a
		// built page must survive truncation for exactly the reason it must
		// survive the plan — the LLM re-proposing 80 pages must not be able to
		// evict a page that is live on the site (bugs_open/001, fix step 3).
		var mustKeep []interface{}
		for _, rp := range existingPages {
			if rm, ok := rp.(map[string]interface{}); ok {
				if noCurrentPlanFlag(rm) || realisedPageCompositionIsPreserved(rm) {
					mustKeep = append(mustKeep, rp)
				}
			}
		}
		pages = truncatePreservingRealised(pages, mustKeep, maxPages, params.Logger)
		plan["pages"] = pages
	}

	pageNames := make(map[string]bool)
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			if name, ok := pm["name"].(string); ok {
				pageNames[name] = true
			}
		}
	}

	if ensurePages, ok := config["ensure_pages"].([]interface{}); ok {
		for _, req := range ensurePages {
			if reqName, ok := req.(string); ok && !pageNames[reqName] {
				pages = append(pages, map[string]interface{}{
					"name": reqName, "title": strings.Title(reqName),
					"nav_label": strings.Title(reqName), "nav_order": len(pages) + 1,
					"in_header": true, "in_footer": true, "sections": []interface{}{},
				})
			}
		}
		plan["pages"] = pages
	}

	if _, ok := plan["style_collection"].(string); !ok {
		if ds, ok := config["default_style"].(string); ok {
			plan["style_collection"] = ds
		}
	}

	if plan["needs_logo"] == nil {
		plan["needs_logo"] = false
	}
	if plan["needs_images"] == nil {
		plan["needs_images"] = false
	}
	if plan["image_prompts"] == nil {
		plan["image_prompts"] = map[string]interface{}{}
	}

	// ── Strip site-chrome components from page sections ──────────────────
	// The LLM sometimes includes header/footer components in page sections
	// arrays (e.g. "header-bold-gradient", "footer-standard"). These are
	// site-level components injected during assembly — not page content.
	// If left in, plan_sections creates bogus HITL items for them.
	if params.DB != nil {
		siteChrome := loadSiteChromeNames(ctx, params.DB, params.Logger)
		if len(siteChrome) > 0 {
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				var filtered []interface{}
				for _, s := range sectionsRaw {
					name, ok := s.(string)
					if !ok {
						filtered = append(filtered, s)
						continue
					}
					if siteChrome[name] {
						params.Logger.Info("ValidateSitePlanAction: stripped site-chrome component from page sections",
							zap.Any("page", pm["name"]),
							zap.String("component", name))
					} else {
						filtered = append(filtered, s)
					}
				}
				pm["sections"] = filtered
			}
		}
	}

	// ── Resolve section names to canonical component functions ───────────
	// Implements config flag `validate_components`. Each section name must
	// map to a real content_components.function. Display names ("FAQ
	// Section"), wrong case, and underscore variants are resolved;
	// unresolvable names are dropped + logged. This does NOT deduplicate or
	// make content-intent decisions — it only guarantees every surviving
	// section name is a valid component function.
	validateComponents := false
	if vc, ok := config["validate_components"].(bool); ok {
		validateComponents = vc
	}
	if validateComponents && params.DB != nil {
		resolver := loadComponentNameResolver(ctx, params.DB, params.Logger)
		if len(resolver.validFunctions) > 0 { // only act if components actually loaded
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				resolved := make([]interface{}, 0, len(sectionsRaw))
				for _, s := range sectionsRaw {
					name, ok := s.(string)
					if !ok {
						resolved = append(resolved, s) // brief objects pass through
						continue
					}
					fn, ok := resolver.resolve(name)
					if !ok {
						params.Logger.Warn("ValidateSitePlanAction: dropped unresolvable section name",
							zap.Any("page", pm["name"]),
							zap.String("section", name))
						continue
					}
					if fn != name {
						params.Logger.Info("ValidateSitePlanAction: resolved section name to function",
							zap.Any("page", pm["name"]),
							zap.String("from", name),
							zap.String("to", fn))
					}
					resolved = append(resolved, fn)
				}
				pm["sections"] = resolved
			}
		} else {
			params.Logger.Warn("ValidateSitePlanAction: validate_components set but no components loaded — skipping name resolution")
		}
	}

	params.Logger.Info("ValidateSitePlanAction: Complete", zap.Int("pages", len(pages)))
	return plan, nil
}

// nullString returns nil for empty strings, otherwise the string pointer
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// loadSiteChromeNames returns a set of component names that are site-level
// chrome (headers, footers, head) — not page content sections.
func loadSiteChromeNames(ctx context.Context, db *sql.DB, logger *zap.Logger) map[string]bool {
	result := make(map[string]bool)
	rows, err := db.QueryContext(ctx,
		`SELECT name FROM content_components WHERE component_level = 'site' AND is_active = true`)
	if err != nil {
		logger.Warn("loadSiteChromeNames: query failed", zap.Error(err))
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if rows.Scan(&name) == nil {
			result[name] = true
		}
	}
	return result
}

// componentNameResolver resolves plan section names to canonical
// content_components.function values. Used to implement the
// validate_components flag and to normalise gap-planner section names.
type componentNameResolver struct {
	validFunctions map[string]bool   // function -> true
	displayToFunc  map[string]string // lower(display_name) -> function
	nameToFunc     map[string]string // lower(name) -> function
}

// loadComponentNameResolver loads section/element component identity so
// plan section names can be resolved to a canonical function. Returns an
// empty (non-nil) resolver on error so callers can no-op safely.
func loadComponentNameResolver(ctx context.Context, db *sql.DB, logger *zap.Logger) *componentNameResolver {
	r := &componentNameResolver{
		validFunctions: make(map[string]bool),
		displayToFunc:  make(map[string]string),
		nameToFunc:     make(map[string]string),
	}
	if db == nil {
		return r
	}
	rows, err := db.QueryContext(ctx,
		`SELECT "function", name, COALESCE(display_name, '')
		   FROM content_components
		  WHERE component_level IN ('section','element')
		    AND is_active = true
		    AND "function" <> ''`)
	if err != nil {
		logger.Warn("loadComponentNameResolver: query failed", zap.Error(err))
		return r
	}
	defer rows.Close()
	for rows.Next() {
		var fn, name, display string
		if err := rows.Scan(&fn, &name, &display); err != nil {
			continue
		}
		r.validFunctions[fn] = true
		if name != "" {
			r.nameToFunc[strings.ToLower(name)] = fn
		}
		if display != "" {
			r.displayToFunc[strings.ToLower(display)] = fn
		}
	}
	return r
}

// resolve maps a raw section name to a canonical component function.
// Returns (function, true) if resolved, ("", false) if not. It does NOT
// deduplicate or make content-intent decisions — only name resolution.
func (r *componentNameResolver) resolve(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	// 1. Already a valid function.
	if r.validFunctions[raw] {
		return raw, true
	}
	// 2. Normalise (underscore->hyphen, camelCase->kebab) and re-check.
	norm := NormalizeComponentFunction(raw)
	if norm != raw && r.validFunctions[norm] {
		return norm, true
	}
	// 3. Display-name lookup (handles "FAQ Section" -> "faq").
	if fn, ok := r.displayToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 4. Component name lookup (row name differing from function).
	if fn, ok := r.nameToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 5. Display lookup on the normalised form.
	if fn, ok := r.displayToFunc[strings.ToLower(norm)]; ok {
		return fn, true
	}
	return "", false
}

// ============================================================================
// ACTION: db_sync
// ============================================================================

// DBSyncAction is a general-purpose database sync action
// Can be used to insert/update records in any table
// Config:
//   - operation: insert, update, upsert
//   - table: table name
//   - data_field: path to data to sync
//   - key_fields: array of fields that form the primary/unique key
func DBSyncAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DBSyncAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("DBSyncAction: No database connection")
		return map[string]interface{}{"synced": false, "reason": "no database"}, nil
	}

	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "upsert"
	}

	table, ok := config["table"].(string)
	if !ok || table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// Get data to sync
	dataField := "sync_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	syncData := datahelpers.ExtractNestedField(params.CollectedData, dataField)
	if syncData == nil {
		return nil, fmt.Errorf("no data found at %s", dataField)
	}

	dataMap, ok := syncData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("sync data must be a map")
	}

	// Build and execute query based on operation
	// This is a simplified implementation - in production you'd want more robust SQL building

	params.Logger.Info("DBSyncAction: Syncing data",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int("field_count", len(dataMap)),
	)

	// For now, just marshal and log - actual implementation would build SQL
	dataJSON, _ := json.Marshal(dataMap)

	return map[string]interface{}{
		"synced":      true,
		"operation":   operation,
		"table":       table,
		"field_count": len(dataMap),
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func execDB(ctx context.Context, db interface{}, query string, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, query, args...)
		return err
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, query, args...)
		return err
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

func lookupPageID(ctx context.Context, db interface{}, siteID uuid.UUID, pageName string, logger *zap.Logger) (uuid.UUID, error) {
	query := `SELECT id FROM pages WHERE site_id = $1 AND name = $2`
	var pageID uuid.UUID

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	default:
		return uuid.Nil, fmt.Errorf("unsupported database type: %T", db)
	}
}

// queryRowScanUUID executes a query and scans the result into a UUID
func queryRowScanUUID(ctx context.Context, db interface{}, query string, dest *uuid.UUID, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		return d.QueryRowContext(ctx, query, args...).Scan(dest)
	case *pgxpool.Pool:
		return d.QueryRow(ctx, query, args...).Scan(dest)
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

// getStyleCollectionByID looks up a style collection by UUID
// This is a local helper since component_library.go doesn't have this function yet
func getStyleCollectionByID(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*StyleCollection, error) {
	query := `
		SELECT id, name, display_name, category, color_palette, typography,
		       header_component_id, footer_component_id, css_theme_id
		FROM style_collections
		WHERE id = $1 AND is_active = true
	`

	var coll StyleCollection
	var colorPaletteJSON, typographyJSON []byte
	var headerID, footerID, cssThemeID *uuid.UUID

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if colorPaletteJSON != nil {
		json.Unmarshal(colorPaletteJSON, &coll.ColorPalette)
	}
	if typographyJSON != nil {
		json.Unmarshal(typographyJSON, &coll.Typography)
	}
	coll.HeaderComponentID = headerID
	coll.FooterComponentID = footerID
	coll.CSSThemeID = cssThemeID

	return &coll, nil
}

func BuildReviewResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildReviewResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := map[string]interface{}{
		"reviewed_at": time.Now().UTC().Format(time.RFC3339),
		"review_mode": "unknown",
		"approved":    false,
		"issues":      []interface{}{},
		"edits":       map[string]interface{}{},
	}

	if approved, ok := config["approved"].(bool); ok {
		result["approved"] = approved
	} else if field, ok := config["approved_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			if b, ok := val.(bool); ok {
				result["approved"] = b
			}
		}
	}

	if reviewer, ok := config["reviewer"].(string); ok && reviewer != "" {
		result["reviewed_by"] = reviewer
	} else if field, ok := config["reviewer_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			result["reviewed_by"] = val
		}
	}
	if result["reviewed_by"] == nil {
		result["reviewed_by"] = "system"
	}

	if mode, ok := config["review_mode"].(string); ok {
		result["review_mode"] = mode
	}

	if field, ok := config["eval_score"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["eval_score"] = val
		}
	}

	if field, ok := config["edits_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["edits"] = val
		}
	}

	if field, ok := config["auto_eval_issues"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["auto_eval_issues"] = val
			if issues, ok := val.([]interface{}); ok {
				result["issues"] = issues
			}
		}
	}

	if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
		result["content"] = content
	}

	params.Logger.Info("BuildReviewResultAction: Complete",
		zap.Bool("approved", result["approved"].(bool)),
		zap.String("mode", result["review_mode"].(string)))
	return result, nil
}

// ============================================================================
// ACTION: prepare_review_data
// ============================================================================

func PrepareReviewDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PrepareReviewDataAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	reviewData := map[string]interface{}{
		"prepared_at": time.Now().UTC().Format(time.RFC3339),
		"fields":      map[string]interface{}{},
	}

	if includeFields, ok := config["include_fields"].([]interface{}); ok {
		fieldsMap := reviewData["fields"].(map[string]interface{})
		for _, field := range includeFields {
			if fieldName, ok := field.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, fieldName); value != nil {
					fieldsMap[fieldName] = value
				}
			}
		}
	}

	if formatForDisplay, ok := config["format_for_display"].(bool); ok && formatForDisplay {
		if page := datahelpers.ExtractNestedField(params.CollectedData, "current_page"); page != nil {
			if pm, ok := page.(map[string]interface{}); ok {
				reviewData["page_name"] = pm["name"]
				reviewData["page_title"] = pm["title"]
			}
		}
		if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
			reviewData["content"] = content
		}
		if brief := datahelpers.ExtractNestedField(params.CollectedData, "reviewed_brief"); brief != nil {
			if bm, ok := brief.(map[string]interface{}); ok {
				reviewData["company_name"] = bm["company_name"]
				reviewData["tone"] = bm["tone"]
			}
		}
	}

	params.Logger.Info("PrepareReviewDataAction: Complete")
	return reviewData, nil
}

// ============================================================================
// ACTION: update_page_components_status
// ============================================================================

func UpdatePageComponentsStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageComponentsStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	newStatus := "approved"
	if status, ok := config["status"].(string); ok && status != "" {
		newStatus = status
	}

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	pageData := datahelpers.ExtractNestedField(params.CollectedData, pageField)
	if pageData == nil {
		return map[string]interface{}{"updated": false, "reason": "no page data", "status_set": newStatus}, nil
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	var pageID uuid.UUID
	if idStr, ok := page["id"].(string); ok && idStr != "" {
		pageID, _ = uuid.Parse(idStr)
	}

	reviewedAt := time.Now().UTC()
	if field, ok := config["reviewed_at_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				reviewedAt = t
			}
		}
	}

	reviewedBy := "system"
	if field, ok := config["reviewed_by_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			reviewedBy = val
		}
	}

	if params.DB != nil && pageID != uuid.Nil {
		query := `UPDATE page_components SET build_status = $1, reviewed_at = $2, reviewed_by = $3, updated_at = NOW() WHERE page_id = $4`
		result, err := params.DB.ExecContext(ctx, query, newStatus, reviewedAt, reviewedBy, pageID)
		if err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
		rows, _ := result.RowsAffected()
		return map[string]interface{}{
			"updated": true, "rows_affected": rows, "page_id": pageID.String(),
			"status_set": newStatus, "reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
		}, nil
	}

	return map[string]interface{}{
		"updated": false, "reason": "no db or page_id", "status_set": newStatus,
		"reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
	}, nil
}

// ============================================================================
// SHARED SECTION-COMPONENT LOADER
// ============================================================================
//
// loadSectionComponents is the canonical loader for component rows used by
// section-level callers. Extracted from LoadPageSectionComponentsAction so
// plan_sections can reuse the same logic without a second SQL path. Both
// callers get the same component-row shape; differences in behaviour are
// expressed by what each caller does with the returned data, not by what
// each caller queries.
//
// Behaviour matches the previous in-action implementation:
//   - Match by name first, fall back to function (DISTINCT ON, newest first)
//   - Stubs for sections with no matching component
//   - Order preserved relative to sectionNames input
//   - When pageID != "", content_brief is attached per slot
//   - When activeOnly is true, only is_active=true rows are returned
//
// activeOnly preserves the historical behaviour difference between the two
// callers: plan_sections used to filter `is_active = true` inline;
// LoadPageSectionComponentsAction did not. Passing the flag explicitly keeps
// both callers' behaviour intact while sharing one query path.
//
// The returned per-component map carries: component_id (when from DB),
// name, function, display_name, category, semantic_tags (when set),
// description (when set), html_template (when set), input_schema (when
// set, as raw JSON string), render_mode, agent_type (when set),
// component_level, needs_llm, and content_brief (when found).
func loadSectionComponents(
	ctx context.Context,
	db *sql.DB,
	sectionNames []string,
	pageID string,
	activeOnly bool,
	logger *zap.Logger,
) []map[string]interface{} {
	if len(sectionNames) == 0 {
		return []map[string]interface{}{}
	}
	if db == nil {
		// No DB available — return name-stubs so callers can still proceed.
		return buildStubSectionComponents(sectionNames)
	}

	// Match each requested section against BOTH its raw name and its kebab-
	// normalised form (bugs_open/041): the library stores kebab-case, but plans
	// may emit snake_case/CamelCase ("call_to_action"). This value set is a
	// strict superset of the raw names, so nothing that resolved before stops
	// resolving — including the few components whose *name* is itself snake_case.
	lookupValues := sectionLookupValueSet(sectionNames)
	placeholders := make([]string, len(lookupValues))
	args := make([]interface{}, len(lookupValues))
	for i, v := range lookupValues {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = v
	}

	var components []map[string]interface{}

	// Pass 1: lookup by name
	activeFilter := ""
	if activeOnly {
		activeFilter = " AND is_active = true"
	}
	nameQuery := fmt.Sprintf(`
		SELECT
			id,
			name,
			COALESCE(display_name, name) AS display_name,
			function,
			COALESCE(category, '') AS category,
			semantic_tags,
			description,
			html_template,
			input_schema,
			COALESCE(render_mode, 'template') AS render_mode,
			agent_type,
			COALESCE(component_level, 'section') AS component_level
		FROM content_components
		WHERE name IN (%s)%s`, strings.Join(placeholders, ", "), activeFilter)

	rows, err := db.QueryContext(ctx, nameQuery, args...)
	if err != nil {
		logger.Error("loadSectionComponents: name query failed",
			zap.Error(err),
			zap.Strings("sections", sectionNames))
	} else {
		for rows.Next() {
			comp, scanErr := scanSectionComponentRow(rows)
			if scanErr != nil {
				logger.Error("loadSectionComponents: name row scan failed",
					zap.Error(scanErr))
				continue
			}
			components = append(components, comp)
		}
		rows.Close()
	}

	// Track which inputs are already satisfied by name or function
	foundNames := make(map[string]bool)
	for _, comp := range components {
		if n, ok := comp["name"].(string); ok {
			foundNames[n] = true
		}
		if fn, ok := comp["function"].(string); ok {
			foundNames[fn] = true
		}
	}

	var missing []string
	for _, name := range sectionNames {
		if !sectionResolvedByFound(foundNames, name) {
			missing = append(missing, name)
		}
	}

	// Pass 2: lookup by function for anything still missing
	if len(missing) > 0 {
		logger.Info("loadSectionComponents: trying function lookup for missing",
			zap.Strings("missing", missing))

		// Same raw+normalised superset as Pass 1 (bugs_open/041).
		funcValues := sectionLookupValueSet(missing)
		funcValueSet := make(map[string]bool, len(funcValues))
		funcPlaceholders := make([]string, len(funcValues))
		funcArgs := make([]interface{}, len(funcValues))
		for i, v := range funcValues {
			funcValueSet[v] = true
			funcPlaceholders[i] = fmt.Sprintf("$%d", i+1)
			funcArgs[i] = v
		}

		funcQuery := fmt.Sprintf(`
			SELECT DISTINCT ON (function)
				id,
				name,
				COALESCE(display_name, name) AS display_name,
				function,
				COALESCE(category, '') AS category,
				semantic_tags,
				description,
				html_template,
				input_schema,
				COALESCE(render_mode, 'template') AS render_mode,
				agent_type,
				COALESCE(component_level, 'section') AS component_level
			FROM content_components
			WHERE function IN (%s)%s
			ORDER BY function, created_at DESC
		`, strings.Join(funcPlaceholders, ", "), activeFilter)

		funcRows, ferr := db.QueryContext(ctx, funcQuery, funcArgs...)
		if ferr != nil {
			logger.Warn("loadSectionComponents: function lookup failed",
				zap.Error(ferr))
		} else {
			for funcRows.Next() {
				comp, scanErr := scanSectionComponentRow(funcRows)
				if scanErr != nil {
					continue
				}
				function, _ := comp["function"].(string)
				if !funcValueSet[function] {
					continue
				}
				components = append(components, comp)
				foundNames[function] = true
				logger.Info("loadSectionComponents: found component by function",
					zap.String("function", function),
					zap.String("name", comp["name"].(string)))
			}
			funcRows.Close()
		}
	}

	// Stubs for anything still not found
	var stillMissing []string
	for _, name := range sectionNames {
		if !sectionResolvedByFound(foundNames, name) {
			stillMissing = append(stillMissing, name)
		}
	}
	if len(stillMissing) > 0 {
		logger.Warn("loadSectionComponents: stubs for unresolved sections",
			zap.Strings("missing", stillMissing))
		for _, name := range stillMissing {
			components = append(components, map[string]interface{}{
				"name":         name,
				"display_name": name,
				"function":     name,
				"category":     "",
				"needs_llm":    true,
				"description":  "",
			})
		}
	}

	// Reorder to match sectionNames input order. Match a component to a requested
	// section under either the raw or normalised form (bugs_open/041), mirroring
	// the lookup above so a resolved "call_to_action" lands in its slot.
	ordered := make([]map[string]interface{}, 0, len(components))
	for _, sectionName := range sectionNames {
		keys := sectionLookupKeys(sectionName)
		for _, comp := range components {
			name, _ := comp["name"].(string)
			function, _ := comp["function"].(string)
			if containsString(keys, name) || containsString(keys, function) {
				ordered = append(ordered, comp)
				break
			}
		}
	}

	// Optional: content_brief enrichment from page_components
	if pageID != "" {
		enrichSectionComponentsWithBriefs(ctx, db, pageID, ordered, logger)
	}

	return ordered
}

// scanSectionComponentRow turns one SQL row into the per-component map shape.
// Centralised so the by-name and by-function passes produce identical shapes.
func scanSectionComponentRow(rows *sql.Rows) (map[string]interface{}, error) {
	var id, name, function string
	var displayName, category sql.NullString
	var semanticTags, description, htmlTemplate, inputSchema sql.NullString
	var renderMode, agentType, componentLevel sql.NullString

	if err := rows.Scan(
		&id, &name, &displayName, &function, &category,
		&semanticTags, &description, &htmlTemplate, &inputSchema,
		&renderMode, &agentType, &componentLevel,
	); err != nil {
		return nil, err
	}

	comp := map[string]interface{}{
		"component_id": id,
		"name":         name,
		"function":     function,
	}
	if displayName.Valid {
		comp["display_name"] = displayName.String
	} else {
		comp["display_name"] = name
	}
	if category.Valid {
		comp["category"] = category.String
	} else {
		comp["category"] = ""
	}
	if semanticTags.Valid {
		comp["semantic_tags"] = semanticTags.String
	}
	if description.Valid && description.String != "" {
		comp["description"] = description.String
	}
	if htmlTemplate.Valid && htmlTemplate.String != "" {
		comp["html_template"] = htmlTemplate.String
	}
	if inputSchema.Valid && inputSchema.String != "" {
		comp["input_schema"] = inputSchema.String
	}
	if renderMode.Valid && renderMode.String != "" {
		comp["render_mode"] = renderMode.String
	} else {
		comp["render_mode"] = "template"
	}
	if agentType.Valid && agentType.String != "" {
		comp["agent_type"] = agentType.String
	}
	if componentLevel.Valid && componentLevel.String != "" {
		comp["component_level"] = componentLevel.String
	} else {
		comp["component_level"] = "section"
	}
	comp["needs_llm"] = detectNeedsLLMContent(htmlTemplate.String, inputSchema.String)
	return comp, nil
}

// buildStubSectionComponents returns minimal stubs for the no-DB code path.
func buildStubSectionComponents(sectionNames []string) []map[string]interface{} {
	stubs := make([]map[string]interface{}, len(sectionNames))
	for i, name := range sectionNames {
		stubs[i] = map[string]interface{}{
			"name":         name,
			"function":     name,
			"display_name": name,
			"description":  "",
			"needs_llm":    true,
		}
	}
	return stubs
}

// enrichSectionComponentsWithBriefs attaches per-section admin content briefs
// (from page_components.content_brief) onto the components in-place.
func enrichSectionComponentsWithBriefs(
	ctx context.Context,
	db *sql.DB,
	pageID string,
	components []map[string]interface{},
	logger *zap.Logger,
) {
	briefRows, briefErr := db.QueryContext(ctx, `
		SELECT COALESCE(slot_name, ''), content_brief
		FROM page_components
		WHERE page_id = $1
		  AND content_brief IS NOT NULL
		  AND build_status != 'removed'
	`, pageID)
	if briefErr != nil {
		logger.Warn("enrichSectionComponentsWithBriefs: query failed",
			zap.Error(briefErr))
		return
	}
	defer briefRows.Close()

	briefMap := make(map[string]interface{})
	for briefRows.Next() {
		var slotName string
		var briefJSON []byte
		if err := briefRows.Scan(&slotName, &briefJSON); err != nil {
			continue
		}
		if len(briefJSON) > 0 && slotName != "" {
			var brief interface{}
			if err := json.Unmarshal(briefJSON, &brief); err == nil {
				briefMap[slotName] = brief
			}
		}
	}
	if len(briefMap) == 0 {
		return
	}

	for _, comp := range components {
		name, _ := comp["name"].(string)
		function, _ := comp["function"].(string)
		if brief, ok := briefMap[name]; ok {
			comp["content_brief"] = brief
		} else if brief, ok := briefMap[function]; ok {
			comp["content_brief"] = brief
		}
	}
	logger.Info("enrichSectionComponentsWithBriefs: attached briefs",
		zap.Int("briefs_found", len(briefMap)))
}

// ============================================================================
// ACTION: load_page_section_components
// ============================================================================
//
// LoadPageSectionComponentsAction loads component definitions for a page's
// sections. Thin workflow wrapper around loadSectionComponents.
//
// Config:
//   - page_from: collected_data path to the page record (default "current_page")
//   - include_templates, include_input_schema: kept for back-compat; the shared
//     loader always returns both, so these flags are not consulted. See
//     doc 019_tool_library.md for the rationale.
//
// When call_agent passes input_fields, they arrive under input_data.*
// extractWithInputDataFallback handles both root and input_data locations.
func LoadPageSectionComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadPageSectionComponentsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	pageData := extractWithInputDataFallback(params.CollectedData, pageField, params.Logger)
	if pageData == nil {
		keys := make([]string, 0, len(params.CollectedData))
		for k := range params.CollectedData {
			keys = append(keys, k)
		}
		params.Logger.Error("Page not found",
			zap.String("page_field", pageField),
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("page not found at '%s'", pageField)
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	params.Logger.Info("Found page data",
		zap.String("page_field", pageField),
		zap.Any("page_name", page["name"]),
		zap.Any("page_title", page["title"]))

	sectionsRaw := page["sections"]
	if sectionsRaw == nil {
		sectionsRaw = extractWithInputDataFallback(params.CollectedData, "sections", params.Logger)
	}

	sectionNames := datahelpers.ExtractSectionNamesHelper(sectionsRaw)
	if len(sectionNames) == 0 {
		params.Logger.Warn("No sections found for page")
		return map[string]interface{}{
			"components":    []interface{}{},
			"count":         0,
			"from_database": false,
		}, nil
	}

	// Enforce naming contract: LLM site plans may output "social_proof" or "SocialProof"
	NormalizeSectionNames(sectionNames, params.Logger)

	params.Logger.Info("Loading components for sections",
		zap.Strings("sections", sectionNames))

	pageID, _ := page["id"].(string)
	// activeOnly=false: this action historically returned components regardless
	// of is_active. Preserved here so existing callers (page-content-writer's
	// load_page_components step, audit flows, admin tools) behave identically.
	components := loadSectionComponents(ctx, params.DB, sectionNames, pageID, false, params.Logger)

	// Detect whether any row carries a real component_id (i.e. came from DB)
	// to preserve the from_database signal callers rely on.
	fromDB := false
	for _, c := range components {
		if _, has := c["component_id"]; has {
			fromDB = true
			break
		}
	}

	return map[string]interface{}{
		"components":    components,
		"count":         len(components),
		"from_database": fromDB,
		"requested":     sectionNames,
	}, nil
}

// Order: direct path -> input_data.{path} -> FindByPath helper
// extractWithInputDataFallback tries to extract a field, falling back to input_data prefix
// This handles the common case where workflows specify paths like "current_section.name"
// but the data is actually at "input_data.current_section.name"
func extractWithInputDataFallback(data map[string]interface{}, path string, logger *zap.Logger) interface{} {
	// Try direct path first
	if value := datahelpers.ExtractNestedField(data, path); value != nil {
		logger.Debug("extractWithInputDataFallback: Found at direct path",
			zap.String("path", path),
		)
		return value
	}

	// If path doesn't already start with input_data, try with prefix
	if !strings.HasPrefix(path, "input_data.") {
		prefixedPath := "input_data." + path
		if value := datahelpers.ExtractNestedField(data, prefixedPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via input_data prefix",
				zap.String("original_path", path),
				zap.String("actual_path", prefixedPath),
			)
			return value
		}
	}

	// Try in __raw_message__.body.input_data (deeply nested case from child agents)
	if !strings.HasPrefix(path, "__raw_message__") {
		rawMsgPath := "__raw_message__.body.input_data." + path
		if value := datahelpers.ExtractNestedField(data, rawMsgPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via __raw_message__.body.input_data",
				zap.String("original_path", path),
			)
			return value
		}
	}

	// Also try agent_config location (for workflow config data)
	if !strings.HasPrefix(path, "agent_config") {
		agentConfigPath := "agent_config." + path
		if value := datahelpers.ExtractNestedField(data, agentConfigPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via agent_config",
				zap.String("original_path", path),
			)
			return value
		}
	}

	logger.Debug("extractWithInputDataFallback: Not found anywhere",
		zap.String("path", path),
	)
	return nil
}

// ============================================================================
// ACTION: filter_search_results
// ============================================================================

// FilterSearchResultsAction filters search results based on criteria
// Handles various response formats: direct array, {results: []}, {data: {results: []}}, etc.
func FilterSearchResultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("FilterSearchResultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	resultsField := "search_results"
	if rf, ok := config["results_field"].(string); ok && rf != "" {
		resultsField = rf
	}

	// Try to find results array - handles various response formats
	results := datahelpers.FindResultsArray(params.CollectedData, resultsField, params.Logger)
	if results == nil {
		params.Logger.Warn("FilterSearchResultsAction: No results found",
			zap.String("results_field", resultsField),
			zap.Strings("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return map[string]interface{}{"filtered_results": []interface{}{}, "count": 0, "original_count": 0}, nil
	}

	params.Logger.Info("FilterSearchResultsAction: Found results",
		zap.Int("count", len(results)))

	// Support both max_results and max_sources config keys
	maxResults := 10
	if mr, ok := config["max_results"].(float64); ok {
		maxResults = int(mr)
	} else if ms, ok := config["max_sources"].(float64); ok {
		maxResults = int(ms)
	}

	excludePatterns := datahelpers.ExtractStringListHelper(config["exclude_patterns"])
	requiredKeywords := datahelpers.ExtractStringListHelper(config["required_keywords"])
	preferDomains := datahelpers.ExtractStringListHelper(config["prefer_domains"])

	var filtered []interface{}
	var preferred []interface{}

	for _, r := range results {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := result["title"].(string)
		content, _ := result["content"].(string)
		snippet, _ := result["snippet"].(string)
		url, _ := result["url"].(string)
		searchText := strings.ToLower(title + " " + content + " " + snippet + " " + url)

		// Check exclusions
		excluded := false
		for _, pattern := range excludePatterns {
			if strings.Contains(searchText, strings.ToLower(pattern)) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check required keywords
		if len(requiredKeywords) > 0 {
			hasKeyword := false
			for _, kw := range requiredKeywords {
				if strings.Contains(searchText, strings.ToLower(kw)) {
					hasKeyword = true
					break
				}
			}
			if !hasKeyword {
				continue
			}
		}

		// Check if from preferred domain
		isPreferred := false
		for _, domain := range preferDomains {
			if strings.Contains(strings.ToLower(url), strings.ToLower(domain)) {
				isPreferred = true
				break
			}
		}

		if isPreferred {
			preferred = append(preferred, result)
		} else {
			filtered = append(filtered, result)
		}
	}

	// Combine: preferred first, then others, up to maxResults
	combined := append(preferred, filtered...)
	if len(combined) > maxResults {
		combined = combined[:maxResults]
	}

	params.Logger.Info("FilterSearchResultsAction: Complete",
		zap.Int("original", len(results)),
		zap.Int("preferred", len(preferred)),
		zap.Int("filtered", len(combined)))

	return map[string]interface{}{
		"filtered_results": combined,
		"count":            len(combined),
		"original_count":   len(results),
		"preferred_count":  len(preferred),
	}, nil
}

// ============================================================================
// ACTION: extract_fields
// ============================================================================

func ExtractFieldsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ExtractFieldsAction: Starting",
		zap.Any("config_keys", getConfigKeys(params.StepConfig.Config)))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := make(map[string]interface{})

	// Handle fields array format: ["field1", "field2"]
	if fields, ok := config["fields"].([]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as array")
		for _, f := range fields {
			switch field := f.(type) {
			case string:
				if value := datahelpers.ExtractNestedField(params.CollectedData, field); value != nil {
					parts := strings.Split(field, ".")
					result[parts[len(parts)-1]] = value
				}
			case map[string]interface{}:
				source, _ := field["source"].(string)
				target, _ := field["target"].(string)
				if source == "" {
					continue
				}
				if target == "" {
					parts := strings.Split(source, ".")
					target = parts[len(parts)-1]
				}
				if value := datahelpers.ExtractNestedField(params.CollectedData, source); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Handle fields as map-of-arrays format (fallback paths)
	// Example: {"topic": ["path1", "path2"], "company": ["path3"]}
	if fieldsMap, ok := config["fields"].(map[string]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as map-of-arrays",
			zap.Int("field_count", len(fieldsMap)))

		for targetField, pathsRaw := range fieldsMap {
			var found bool

			// Handle array of paths
			if paths, ok := pathsRaw.([]interface{}); ok {
				for _, pathRaw := range paths {
					if path, ok := pathRaw.(string); ok {
						// Try direct path first
						if value := datahelpers.ExtractNestedField(params.CollectedData, path); value != nil {
							result[targetField] = value
							found = true
							params.Logger.Info("ExtractFieldsAction: Found via direct path",
								zap.String("target", targetField),
								zap.String("path", path))
							break
						}

						// Try with input_data prefix
						if !strings.HasPrefix(path, "input_data.") {
							prefixedPath := "input_data." + path
							if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
								result[targetField] = value
								found = true
								params.Logger.Info("ExtractFieldsAction: Found via input_data prefix",
									zap.String("target", targetField),
									zap.String("original_path", path),
									zap.String("prefixed_path", prefixedPath))
								break
							}
						}
					}
				}
			}

			// Handle single string path (not array)
			if singlePath, ok := pathsRaw.(string); ok && !found {
				if value := datahelpers.ExtractNestedField(params.CollectedData, singlePath); value != nil {
					result[targetField] = value
					found = true
				} else if !strings.HasPrefix(singlePath, "input_data.") {
					prefixedPath := "input_data." + singlePath
					if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
						result[targetField] = value
						found = true
					}
				}
			}

			if !found {
				params.Logger.Warn("ExtractFieldsAction: Field not found in any path",
					zap.String("target", targetField),
					zap.Any("tried_paths", pathsRaw))
			}
		}
	}

	// Handle field_map format: {"target": "source"}
	if fieldMap, ok := config["field_map"].(map[string]interface{}); ok {
		for target, source := range fieldMap {
			if sourceStr, ok := source.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, sourceStr); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Apply defaults
	if defaults, ok := config["defaults"].(map[string]interface{}); ok {
		for key, val := range defaults {
			if result[key] == nil {
				result[key] = val
			}
		}
	}

	params.Logger.Info("ExtractFieldsAction: Complete",
		zap.Int("fields_extracted", len(result)),
		zap.Strings("result_keys", datahelpers.GetMapKeys(result)))
	return result, nil
}

// Priority 5: Try to build query from section context (for research-agent)
// Check both root level and inside input_data
func getSearchQueryFromSectionContext(params ActionParams) string {
	var currentSection map[string]interface{}

	// Try root level first
	if cs, ok := params.CollectedData["current_section"].(map[string]interface{}); ok {
		currentSection = cs
	}
	// Try input_data.current_section
	if currentSection == nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if cs, ok := inputData["current_section"].(map[string]interface{}); ok {
				currentSection = cs
			}
		}
	}

	if currentSection == nil {
		return ""
	}

	// Get section name/function
	sectionName := ""
	if fn, ok := currentSection["function"].(string); ok && fn != "" {
		sectionName = fn
	} else if name, ok := currentSection["name"].(string); ok && name != "" {
		sectionName = name
	}

	if sectionName == "" {
		return ""
	}

	// Get domain
	domain := ""
	if d, ok := params.CollectedData["domain"].(string); ok {
		domain = d
	} else if siteRecord, ok := params.CollectedData["site_record"].(map[string]interface{}); ok {
		if d, ok := siteRecord["domain"].(string); ok {
			domain = d
		}
	} else if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if siteRecord, ok := inputData["site_record"].(map[string]interface{}); ok {
			if d, ok := siteRecord["domain"].(string); ok {
				domain = d
			}
		}
	}

	// Build query
	query := sectionName
	if domain != "" {
		query = fmt.Sprintf("%s %s", sectionName, domain)
	}

	return query
}

// containsString checks if a string slice contains a specific string
// extractContentWithFallbacks tries multiple paths to find content data
// This handles different output formats from execute_llm_prompt (with/without .result wrapper)
func extractContentWithFallbacks(data map[string]interface{}, contentField string, logger *zap.Logger) map[string]interface{} {
	// Build list of paths to try
	pathsToTry := []string{contentField}

	// If path ends with ".result", also try without it
	if strings.HasSuffix(contentField, ".result") {
		base := strings.TrimSuffix(contentField, ".result")
		pathsToTry = append(pathsToTry, base)
		pathsToTry = append(pathsToTry, base+".response")
		pathsToTry = append(pathsToTry, base+".content")
	} else {
		// If path doesn't end with .result, also try with it
		pathsToTry = append(pathsToTry, contentField+".result")
		pathsToTry = append(pathsToTry, contentField+".response")
		pathsToTry = append(pathsToTry, contentField+".content")
	}

	// Try each path
	for _, path := range pathsToTry {
		if extracted := datahelpers.ExtractNestedField(data, path); extracted != nil {
			if m, ok := extracted.(map[string]interface{}); ok && len(m) > 0 {
				logger.Debug("extractContentWithFallbacks: Found content",
					zap.String("path", path),
					zap.Int("field_count", len(m)))
				return m
			}
		}
	}

	// Last resort: check if the base field exists and contains the content directly
	// Sometimes LLM output is stored as the field value itself
	parts := strings.Split(contentField, ".")
	if len(parts) > 0 {
		baseField := parts[0]
		if baseData := datahelpers.ExtractNestedField(data, baseField); baseData != nil {
			if m, ok := baseData.(map[string]interface{}); ok && len(m) > 0 {
				// Check if this map contains content-like fields
				if hasContentFields(m) {
					logger.Debug("extractContentWithFallbacks: Found content at base field",
						zap.String("base_field", baseField),
						zap.Int("field_count", len(m)))
					return m
				}
			}
		}
	}

	logger.Debug("extractContentWithFallbacks: No content found",
		zap.String("original_path", contentField),
		zap.Strings("tried_paths", pathsToTry))

	return nil
}

// hasContentFields checks if a map contains typical content field names
func hasContentFields(m map[string]interface{}) bool {
	contentFieldNames := []string{
		"headline", "subheadline", "body", "content", "heading",
		"title", "description", "text", "features", "items",
		"primary_cta", "cta_text", "button_text",
	}
	for _, fieldName := range contentFieldNames {
		if _, exists := m[fieldName]; exists {
			return true
		}
	}
	return false
}

// detectNeedsLLMContent determines if a component template needs LLM-generated content
// Returns true if template has content placeholders that need dynamic content
// Returns false if template only has structural placeholders (domain, logo_text, etc.)
func detectNeedsLLMContent(htmlTemplate, inputSchema string) bool {
	// No template = needs LLM to generate something
	if htmlTemplate == "" {
		return true
	}

	// If has input_schema, it defines what content is needed
	if inputSchema != "" && inputSchema != "{}" && inputSchema != "null" {
		return true
	}

	// Content placeholders that indicate LLM content is needed
	contentPlaceholders := []string{
		"{{.headline}}", "{{.subheadline}}", "{{.body}}", "{{.content}}",
		"{{.description}}", "{{.text}}", "{{.paragraph}}",
		"{{.primary_cta}}", "{{.secondary_cta}}", "{{.cta_text}}",
		"{{.features}}", "{{.services}}", "{{.benefits}}", "{{.items}}",
		"{{.testimonial}}", "{{.quote}}", "{{.author}}",
		"{{.heading}}", "{{.subtitle}}",
		"{{ .headline}}", "{{ .body}}", "{{ .content}}",
		"{{range .features}}", "{{range .services}}", "{{range .items}}",
		"{{range .benefits}}", "{{range .testimonials}}",
	}

	// Check for content placeholders
	for _, placeholder := range contentPlaceholders {
		if strings.Contains(htmlTemplate, placeholder) {
			return true
		}
	}

	// Structural fields that DON'T require LLM (filled from render_context)
	structuralFields := map[string]bool{
		"domain": true, "logo_text": true, "company_name": true,
		"tagline": true, "current_page": true, "year": true,
		"primary_color": true, "secondary_color": true, "accent_color": true,
		"background_color": true, "text_color": true,
		"email": true, "phone": true, "address": true,
		"nav_items_html": true, "theme_css": true, "title": true,
		"cta_url": true, "industry": true, "tone": true,
	}

	// Check for any {{.X}} where X is not a known structural field
	re := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(htmlTemplate, -1)
	for _, match := range matches {
		if len(match) > 1 {
			fieldName := match[1]
			if !structuralFields[fieldName] {
				return true
			}
		}
	}

	// Only structural placeholders - doesn't need LLM
	return false
}

// UpdateWorkItemStatusAction updates a site_work_items row's status.
// Config:
//   - work_item_id_field: path to work_item_id in collected_data
//     (default: "input_data.work_item_id")
//   - status:             new status — default "complete"
//     (valid: complete, failed, claimed, executing,
//     detected, wont_fix, needs_human_review, unresolved)
//   - skip_if_missing:    bool — when true (default), gracefully no-op if
//     work_item_id absent. When false, error.
//   - error_message:      optional literal recorded in the error column so
//     triage can see why a handler parked the item.
//     When omitted and the status is not 'complete',
//     the routed step error (__step_error) is
//     recorded instead — see below.
//   - result_fields:      optional map of extra fields to merge into the
//     row's result JSONB. Values are literals; the
//     action always adds orchestration_id and step
//     metadata automatically.
func UpdateWorkItemStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateWorkItemStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get work_item_id
	workItemIDField := "input_data.work_item_id"
	if f, ok := config["work_item_id_field"].(string); ok && f != "" {
		workItemIDField = f
	}
	workItemIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, workItemIDField)

	// Skip gracefully if missing — supports manual triggers without work_item_id
	skipIfMissing := true
	if v, ok := config["skip_if_missing"].(bool); ok {
		skipIfMissing = v
	}

	if workItemIDStr == "" {
		if skipIfMissing {
			params.Logger.Info("UpdateWorkItemStatusAction: no work_item_id in collected_data; skipping",
				zap.String("looked_at", workItemIDField))
			return map[string]interface{}{
				"updated": false,
				"skipped": true,
				"reason":  "work_item_id not present",
			}, nil
		}
		return nil, fmt.Errorf("work_item_id not found at %s and skip_if_missing=false", workItemIDField)
	}

	workItemID, err := uuid.Parse(workItemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id %q: %w", workItemIDStr, err)
	}

	// Status (default complete)
	newStatus := "complete"
	if s, ok := config["status"].(string); ok && s != "" {
		newStatus = s
	}
	validStatuses := map[string]bool{
		"complete":  true,
		"failed":    true,
		"claimed":   true,
		"executing": true,
		"detected":  true,
		"wont_fix":  true,
		// Handler no-op flags: a handler that cannot address its work item
		// (no ready sections, writer skipped) parks the item visibly here.
		// Without a flagged status the dispatch loop would stamp the item
		// complete on saga success — the false-completion bug caught on
		// robot-hands' gripper-detail (2026-07-10).
		"needs_human_review": true,
		"unresolved":         true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid work item status: %s (valid: complete, failed, claimed, executing, detected, wont_fix, needs_human_review, unresolved)", newStatus)
	}

	// Optional error message — recorded in the error column so triage can see
	// why a handler parked the item.
	errorMessage, _ := config["error_message"].(string)

	// Fall back to the REAL step error when the workflow supplied no literal.
	//
	// Why (bugs_closed/040-partial-build, candidate 2): a literal only fits a
	// static reason ("writer skipped this page"). The genuinely failing path —
	// `mark_item_failed`, reached via error_step — has a *dynamic* reason and so
	// carries no literal, which left `site_work_items.error` EMPTY on every such
	// item while the coordinator had already written the real message to
	// agent_error_log 1s earlier from the very same routeToErrorStep call
	// (coordinator.go: __step_error and logAgentError are set together). 21 of 75
	// failed items fleet-wide carried a blank error on 2026-07-25; 20 of those 21
	// had exactly one agent_error_log row waiting to be joined. Triage looks at
	// the item, not the log, so the reason was effectively invisible.
	//
	// Never for 'complete': __step_error is never cleared once set, so a workflow
	// that recovers from a routed error and then stamps the item complete would
	// otherwise be given a stale failure. Fleet census 2026-07-25 — the only
	// literal-less update_work_item_status steps are 2×'failed'
	// (page-build-handler, image-build-handler) and 1×'complete'
	// (image-build-handler); the 2×'needs_human_review' steps carry literals and
	// are unaffected because a configured literal always wins.
	if errorMessage == "" && newStatus != "complete" {
		if stepErr := datahelpers.ExtractNestedFieldString(params.CollectedData, "__step_error.message"); stepErr != "" {
			errorMessage = stepErr
			// Name the step unless the message already does. The routed message
			// has two shapes: "step X failed: …" (action errors) and a bare
			// "Request <id> timed out after 3 retries" (awaited-request
			// timeouts), which alone does not say WHAT timed out. Converge on
			// the prefix the column already uses rather than inventing a
			// second format.
			if failedStep := datahelpers.ExtractNestedFieldString(params.CollectedData, "__step_error.failed_step"); failedStep != "" &&
				!strings.HasPrefix(errorMessage, "step ") {
				errorMessage = "step " + failedStep + " failed: " + errorMessage
			}
			params.Logger.Info("UpdateWorkItemStatusAction: no error_message literal — recording the routed step error",
				zap.String("work_item_id", workItemIDStr),
				zap.String("status", newStatus),
				zap.String("error", errorMessage))
		}
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateWorkItemStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	// Build result payload — always includes orchestration tracking; merges
	// any caller-supplied extras under result_fields.
	resultPayload := map[string]interface{}{
		"completed_by_orchestration_id": params.ExecutionContext.OrchestrationID,
		"completed_by_step":             params.ExecutionContext.StepName,
		"completed_at_iso":              time.Now().UTC().Format(time.RFC3339),
	}
	if extras, ok := config["result_fields"].(map[string]interface{}); ok {
		for k, v := range extras {
			resultPayload[k] = v
		}
	}
	resultJSON, err := json.Marshal(resultPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result payload: %w", err)
	}

	// Build query. completed_at only set when transitioning to complete; for
	// other statuses (failed, etc.) leave it alone and just update status.
	var query string
	if newStatus == "complete" {
		query = `UPDATE site_work_items
		            SET status = $2,
		                completed_at = NOW(),
		                updated_at = NOW(),
		                attempt_count = attempt_count + 1,
		                result = COALESCE(result, '{}'::jsonb) || $3::jsonb,
		                error = COALESCE(NULLIF($4, ''), error)
		          WHERE id = $1`
	} else {
		query = `UPDATE site_work_items
		            SET status = $2,
		                updated_at = NOW(),
		                attempt_count = attempt_count + 1,
		                result = COALESCE(result, '{}'::jsonb) || $3::jsonb,
		                error = COALESCE(NULLIF($4, ''), error)
		          WHERE id = $1`
	}

	if err := execDB(ctx, params.DB, query, workItemID, newStatus, resultJSON, errorMessage); err != nil {
		return nil, fmt.Errorf("failed to update work item status: %w", err)
	}

	params.Logger.Info("UpdateWorkItemStatusAction: status updated",
		zap.String("work_item_id", workItemIDStr),
		zap.String("status", newStatus),
	)

	return map[string]interface{}{
		"updated":      true,
		"work_item_id": workItemIDStr,
		"status":       newStatus,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ============================================================================
// Adoption-faithfulness convergence (doc 029 Phase 1 / FOCUS_adoption_faithfulness_via_locks.md)
//
// These helpers make ValidateSitePlanAction deterministically preserve the
// pages that are currently under a live adoption lock, so the planner LLM
// cannot drop, rename, or duplicate adopted pages during the faithful-first-
// pass window. They become no-ops once the adoption lock has expired (or for
// from-scratch builds), letting the site develop normally thereafter.
// ============================================================================

// isSectionIndexType reports whether a page_type is a directory/section index.
func isSectionIndexType(pageType string) bool {
	switch pageType {
	case "blog-index", "entity-directory", "section-index":
		return true
	}
	return false
}

// sectionStemOf returns the section stem for a realised section-index page, or
// "" if it isn't a section index. e.g. ("games-index", "/games/index.html",
// "entity-directory") -> "games". Prefers the URL path segment; falls back to
// the name minus the "-index" suffix.
func sectionStemOf(name, url, pageType string) string {
	isIndex := isSectionIndexType(pageType)
	if !isIndex {
		// Also treat any non-root URL ending in /index.html as a section index.
		if strings.HasSuffix(url, "/index.html") && url != "/index.html" {
			isIndex = true
		}
	}
	if !isIndex {
		return ""
	}
	trimmed := strings.Trim(url, "/")
	if i := strings.Index(trimmed, "/"); i > 0 {
		return trimmed[:i]
	}
	return strings.TrimSuffix(name, "-index")
}

// slugOf derives a comparison slug from an LLM page's name/url. A flat page URL
// like /games.html yields "games"; falls back to the name.
func slugOf(name, url string) string {
	if url != "" {
		t := strings.Trim(url, "/")
		t = strings.TrimSuffix(t, ".html")
		t = strings.TrimSuffix(t, "/index")
		if t != "" {
			if i := strings.Index(t, "/"); i > 0 {
				return t[:i]
			}
			return t
		}
	}
	return name
}

// itemStemOf returns the topic stem of an item page name by stripping the
// role prefixes that CanonicalisePage adds (tool-, guide-, game-): e.g.
// "guide-economy-basics" -> "economy-basics", "economy-basics" ->
// "economy-basics". Mirrors the TrimPrefix calls in CanonicalisePage's
// tool/guide/game cases - keep this prefix list in sync with them. Returns
// the input unchanged when no role prefix is present, so two adopted pages on
// the same topic share a stem and a re-proposed bare sibling collides with
// them. Unlike sectionStemOf, this is name-based rather than URL/hub-based.
func itemStemOf(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, p := range []string{"tool-", "guide-", "game-"} {
		if strings.HasPrefix(n, p) {
			return strings.TrimPrefix(n, p)
		}
	}
	return n
}

// normaliseRealisedToPlanPage converts a realised pages-table row (as returned
// by the load_existing_pages query) into the plan-page shape the downstream
// write_site_plan / ValidateRoles expects. Carries a from_realised marker so
// logging/debugging can distinguish these from LLM-proposed pages.
func normaliseRealisedToPlanPage(rm map[string]interface{}) map[string]interface{} {
	name, _ := rm["name"].(string)
	pageType, _ := rm["page_type"].(string)
	// Carry the realised page's sections so a unioned adopted page keeps them.
	// load_existing_pages runs via query_database, which stringifies jsonb
	// columns, so rm["sections"] arrives as a JSON string (e.g. ["hero",
	// "guide-list"]); tolerate a native []interface{} too. Empty/missing -> [].
	// Without carrying these, the union emits empty values and the page sync's
	// "<col> = EXCLUDED.<col>" clobbers the adopted page's real sections,
	// meta_description, and nav_order. (nav_label is COALESCE-preserved by the
	// upsert, so it is safe without carrying.)
	sections := []interface{}{}
	switch v := rm["sections"].(type) {
	case []interface{}:
		sections = v
	case string:
		if v != "" {
			var parsed []interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				sections = parsed
			}
		}
	}
	return map[string]interface{}{
		"name":             name,
		"page_type":        pageType,
		"url":              rm["url"],
		"title":            rm["title"],
		"nav_label":        rm["nav_label"],
		"in_header":        rm["in_header"],
		"in_footer":        rm["in_footer"],
		"meta_description": rm["meta_description"],
		"nav_order":        rm["nav_order"],
		"sections":         sections,
		"from_realised":    true,
	}
}

// reconcilePlanWithRealised enforces preservation of and convergence on the
// realised pages a re-plan must not silently redesign or drop.
//
// PRESERVATION SET (widened 2026-07-19, bugs_open/001). Formerly this was the
// first-plan subset alone, which made the whole function a no-op on every
// re-plan. NOTE what that flag actually is (renamed 2026-07-22 from the
// misleading "adoption_locked", bugs_open/051): NOT a per-page or 90-day lock —
// there never was one. The live load_existing_pages query surfaces it as
// site_has_no_current_plan, derived per SITE, so it is true for every page on a
// site's FIRST plan and false for every page on every re-plan after that. The two-branch design in 053 §054 (branch (b): a live
// timed per-page preserve-directive) is absent from the live query and has zero
// rows behind it fleet-wide, so no per-page lock has ever existed. The realised
// composition of a BUILT page was then carried by nothing: a page the LLM
// re-proposed under the same name was silently re-composed to whatever the LLM
// proposed that run, and a page the LLM omitted was dropped from the plan
// outright. Proven on idea.uk 2026-07-14 (plan 32be2797 -> ff03bdef): four
// built pages regressed, two of which re-rendered and re-deployed the regressed
// artefact.
//
// A built page deserves preservation whether or not it is on the first plan, so
// the set is now:
//
//	site_has_no_current_plan == true  OR  build_status IN ("deployed", "needs_rebuild")
//
// needs_rebuild joined the set 2026-07-21 (bugs_open/037). A needs_rebuild page
// still holds its intended composition in pages.sections, and EVERY writer of
// that status keeps those sections and means "re-render this page as planned",
// never "recompose it from scratch": a refused 0-component or partial deploy
// (UpdatePageStatusAction — clears built_from_plan_version but keeps sections),
// an image/maintenance rebuild (flagPagesForRebuild), or a now-available
// component the sections already name (markPagesForRebuild,
// check_unresolved_sections — those two would be actively DEFEATED by
// recomposition, since the sections name the very components the rebuild exists
// to pick up). So letting a re-plan take the LLM's composition for a
// needs_rebuild page was silent loss, not an honoured redesign request. This
// widens only MEMBERSHIP of the preserved set (realisedPageCompositionIsPreserved);
// the empty-sections classification in Pass B/B2 still keys on realisedPageIsBuilt
// (== deployed), because a needs_rebuild page with empty sections may be either
// rendered-elsewhere OR genuinely awaiting composition, which Pass B2's non-empty
// gate already routes correctly (see bugs_open/050 for the deployed-empty case).
//
// All flags come from the load_existing_pages query. build_status is only
// surfaced by that query as of migration 173 — if it is absent both status terms
// are empty and behaviour falls back to the first-plan set, so the Go change
// and the query change are safe to land in either order.
//
// Passes over the preservation set:
//
//	Pass C  — section-collision dedup: drop an LLM page whose slug equals the
//	          stem of a realised section index ("games" vs "games-index").
//	Pass C2 — item-topic dedup. Deliberately still scoped to the FIRST-PLAN
//	          subset, not the widened set: it is a name-stem heuristic, and a
//	          false positive suppresses a legitimately new page (a new
//	          "tool-pricing" beside a built "guide-pricing" shares the stem
//	          "pricing"). Made permanent for every built page that risk is not
//	          acceptable, and it is not needed for this bug — invented pages carry
//	          new topics and so collide with nothing. See bugs_open/001 "pages
//	          invented", which this does not claim to fix.
//	          CORRECTED 2026-07-20 (bugs_open/051): this used to read "bounded to
//	          the 90-day window that risk is acceptable". There is no 90-day
//	          window — see the preservation-set note above. Because noCurrentPlanPages
//	          is empty whenever the site has a current plan, Pass C2 can fire ONLY on
//	          a site's first plan and never on a re-plan. The scoping decision
//	          stands; the reason given for it did not exist.
//	Pass B  — rename snap-back: same URL as a realised page, different name ->
//	          replace with the realised identity. Its sections are carried too,
//	          EXCEPT when the realised page is NOT deployed and its sections are
//	          empty: that is a catalogued page that has never been composed, so
//	          keep the realised identity but take the LLM's proposed sections so
//	          the re-plan can finally compose it (bugs_open/050).
//	Pass B2 — composition snap-back: same NAME as a realised page. Reconciles the
//	          LLM's sections against the realised composition, keyed on the
//	          realised sections and deployed-ness (bugs_open/050):
//	            NON-EMPTY            -> restore those sections over the LLM's; a
//	                                    built page must not be re-composed.
//	            EMPTY + deployed     -> force the LLM's proposal back to empty; the
//	                                    page renders through another subsystem (a
//	                                    tool or blog-index page) and must not
//	                                    receive an injected generic layout.
//	            EMPTY + not-deployed -> keep the LLM's proposal; a catalogued page
//	                                    is finally composed.
//	          Carrying emptiness forward unconditionally is what made "re-plan to
//	          compose the missing pages" structurally impossible (bugs_open/001,
//	          second defect); composing onto a deployed sectionless page is the
//	          injection risk bugs_open/050 closes. For a deployed page, empty
//	          sections is a positive statement ("not section-composed here"), not
//	          an absence awaiting composition.
//	Pass A  — union: append every preserved realised page not already present.
//
// Returns the reconciled page slice plus counts for logging.
func reconcilePlanWithRealised(
	llmPages []interface{},
	existingPages []interface{},
	logger *zap.Logger,
) ([]interface{}, int, int, int, int) {
	// Force-preserve first-plan (no-current-plan) OR built pages (see header).
	var preserved []interface{}
	var noCurrentPlanPages []interface{}
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		noCurrentPlan := noCurrentPlanFlag(rm)
		if noCurrentPlan {
			noCurrentPlanPages = append(noCurrentPlanPages, rp)
		}
		if noCurrentPlan || realisedPageCompositionIsPreserved(rm) {
			preserved = append(preserved, rp)
		}
	}
	if len(preserved) == 0 {
		// Nothing realised worth converging on: a genuinely from-scratch build.
		// Leave the LLM plan untouched.
		return llmPages, 0, 0, 0, 0
	}
	existingPages = preserved

	realisedByURL := make(map[string]map[string]interface{})
	realisedByName := make(map[string]map[string]interface{})
	sectionStems := make(map[string]string) // stem -> realised index name
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		url, _ := rm["url"].(string)
		pageType, _ := rm["page_type"].(string)
		if url != "" {
			realisedByURL[url] = rm
		}
		if name != "" {
			realisedByName[name] = rm
		}
		if stem := sectionStemOf(name, url, pageType); stem != "" {
			sectionStems[stem] = name
		}
	}

	// Item-topic stems: the role-prefix-stripped name stem of each realised
	// page (guide-economy-basics -> economy-basics). Keyed to a SET of realised
	// names so a topic legitimately covered by two adopted pages (e.g. a tool
	// and a guide) does not false-positive on either of them. Lets Pass C2 drop
	// an LLM page that re-proposes an adopted item under a different
	// prefix/role/URL.
	//
	// Built deliberately from noCurrentPlanPages, NOT the widened preservation
	// set — see the Pass C2 note in the header for why this one heuristic stays
	// narrow. In practice that makes it first-plan-only: noCurrentPlanPages is
	// empty whenever the site has a current plan (bugs_open/051).
	itemStemSets := make(map[string]map[string]bool)
	for _, rp := range noCurrentPlanPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		if name == "" {
			continue
		}
		stem := itemStemOf(name)
		if stem == "" {
			continue
		}
		if itemStemSets[stem] == nil {
			itemStemSets[stem] = make(map[string]bool)
		}
		itemStemSets[stem][name] = true
	}

	// Pass C (collision) + Pass B (rename) + Pass B2 (composition) over the
	// LLM pages.
	var kept []interface{}
	droppedCollision, snappedRename, snappedSections := 0, 0, 0
	for _, lp := range llmPages {
		lm, ok := lp.(map[string]interface{})
		if !ok {
			kept = append(kept, lp)
			continue
		}
		lname, _ := lm["name"].(string)
		lurl, _ := lm["url"].(string)
		ltype, _ := lm["page_type"].(string)
		lslug := slugOf(lname, lurl)

		// Pass C: flat page colliding with a realised section index.
		if idxName, isStem := sectionStems[lslug]; isStem &&
			!isSectionIndexType(ltype) && lname != idxName {
			logger.Info("validate: dropped flat page colliding with realised section index",
				zap.String("dropped", lname), zap.String("kept_index", idxName))
			droppedCollision++
			continue
		}

		// Pass C2: item-topic collision - the LLM re-proposes an adopted item
		// under a different name/prefix/role (e.g. "economy-basics" beside the
		// adopted "guide-economy-basics"; different URL, so Pass B misses it).
		// Drop it; the adopted page already covers the topic. Skips when the LLM
		// name IS one of the realised names for that stem (a preserved page).
		if names, isStem := itemStemSets[itemStemOf(lname)]; isStem && !names[lname] {
			logger.Info("validate: dropped page duplicating an adopted item topic",
				zap.String("dropped", lname),
				zap.String("stem", itemStemOf(lname)))
			droppedCollision++
			continue
		}

		// Pass B: same URL as a realised page, different name -> snap back to the
		// realised identity, carrying its sections. Exception (bugs_open/050): a
		// NOT-deployed realised page with empty sections is a catalogued page that
		// has never been composed, so keep the realised identity but take the LLM's
		// proposed sections. A DEPLOYED empty page renders through another subsystem
		// and its emptiness is authoritative — carry it (as normalise already does).
		if rp, ok := realisedByURL[lurl]; ok {
			if rname, _ := rp["name"].(string); rname != "" && rname != lname {
				snapped := normaliseRealisedToPlanPage(rp)
				if !realisedPageIsBuilt(rp) && len(realisedSectionsOf(rp)) == 0 {
					if ls, ok := lm["sections"].([]interface{}); ok && len(ls) > 0 {
						snapped["sections"] = ls
					}
				}
				logger.Info("validate: snapped renamed page back to realised identity",
					zap.String("llm_name", lname), zap.String("realised_name", rname))
				kept = append(kept, snapped)
				snappedRename++
				continue
			}
		}

		// Pass B2: same NAME as a preserved realised page -> reconcile the LLM's
		// sections against the realised composition (bugs_open/050). Only
		// "sections" is touched — title/meta/nav stay the LLM's, so a re-plan can
		// still refresh copy and navigation without touching the layout.
		//   - realised NON-EMPTY: restore those sections over the LLM's; a page
		//     built through the section composer must not be re-composed.
		//   - realised EMPTY + deployed: force the LLM's proposal back to empty; the
		//     page renders through another subsystem (a tool or blog-index page)
		//     and must not receive an injected generic layout.
		//   - realised EMPTY + not-deployed: keep the LLM's proposal (fall through);
		//     a catalogued page is finally composed.
		if rp, ok := realisedByName[lname]; ok {
			if rs := realisedSectionsOf(rp); len(rs) > 0 {
				if !sameSectionList(lm["sections"], rs) {
					logger.Info("validate: snapped built page composition back to realised sections",
						zap.String("page", lname),
						zap.Int("realised_sections", len(rs)))
					lm["sections"] = rs
					snappedSections++
				}
			} else if realisedPageIsBuilt(rp) {
				if ls, ok := lm["sections"].([]interface{}); ok && len(ls) > 0 {
					logger.Info("validate: forced deployed sectionless page back to empty",
						zap.String("page", lname),
						zap.Int("llm_sections", len(ls)))
					lm["sections"] = []interface{}{}
					snappedSections++
				}
			}
		}
		kept = append(kept, lm)
	}

	// Pass A: union — add preserved realised pages not present by name.
	presentName := make(map[string]bool)
	for _, p := range kept {
		if pm, ok := p.(map[string]interface{}); ok {
			if n, _ := pm["name"].(string); n != "" {
				presentName[n] = true
			}
		}
	}
	unioned := 0
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		if name == "" || presentName[name] {
			continue
		}
		kept = append(kept, normaliseRealisedToPlanPage(rm))
		presentName[name] = true
		unioned++
	}

	return kept, unioned, droppedCollision, snappedRename, snappedSections
}

// realisedPageIsBuilt reports whether a realised pages-table row (as returned by
// the load_existing_pages query) represents a page that is actually built and
// deployed. Mirrors decideEmit's "skip_built" test in
// reconcile_site_plan_action.go — build_status='deployed' is the platform's
// single definition of "built", so keep the two in step.
//
// Returns false when build_status is absent, which is what makes the widened
// preservation set degrade safely to the first-plan set on a chassis whose
// load_existing_pages query has not yet been updated to surface the column.
func realisedPageIsBuilt(rm map[string]interface{}) bool {
	status, _ := rm["build_status"].(string)
	return status == "deployed"
}

// noCurrentPlanFlag reports the load_existing_pages "site_has_no_current_plan"
// flag for a realised page: true when the page's site has no current plan yet —
// uniquely the site's FIRST plan after adoption (bugs_open/051). It force-preserves
// adopted pages through that one plan and is empty on every re-plan thereafter.
//
// Renamed 2026-07-22 from the misleading "adoption_locked" (there was never a
// per-page or 90-day lock — bugs_open/051). The live query now emits ONLY
// site_has_no_current_plan: migration 193 added it beside the old alias, the
// renamed chassis (v1.0.1151) went fleet-live reading it, then migration 194
// dropped the adoption_locked alias. The adoption_locked read below is KEPT as a
// defensive compat path — it is dead against the current query, costs nothing,
// resolves a snapshot rollback of 194, and is what the reconcile tests exercise
// (their fixtures set the old key). An absent flag degrades to false, matching
// realisedPageIsBuilt's treatment of a missing column.
func noCurrentPlanFlag(rm map[string]interface{}) bool {
	if v, _ := rm["site_has_no_current_plan"].(bool); v {
		return true
	}
	v, _ := rm["adoption_locked"].(bool) // legacy alias, kept as a defensive compat read
	return v
}

// realisedPageCompositionIsPreserved reports whether a realised pages-table row
// carries a composition a re-plan must not silently discard. Two build states
// qualify, for the same reason — the page already holds an intended composition
// in pages.sections that machinery expects to survive a re-plan:
//
//	"deployed"      -- built and live.
//	"needs_rebuild" -- awaiting a re-render, but its sections ARE its intended
//	    composition. Every writer of needs_rebuild keeps pages.sections and means
//	    "re-render as planned", never "recompose from scratch" (bugs_open/037; see
//	    the reconcilePlanWithRealised header for the enumeration of writers).
//
// This is the preservation-MEMBERSHIP predicate, used by reconcilePlanWithRealised
// and the truncation guard. It is deliberately DISTINCT from realisedPageIsBuilt
// (== deployed), which the empty-sections gate in Pass B/B2 keeps using: an empty
// needs_rebuild page may be genuinely awaiting composition rather than rendered
// elsewhere, so it must not be force-emptied. Pass B2's non-empty gate routes both
// kinds correctly — a needs_rebuild page with real sections is snapped back; one
// with empty sections falls through to the LLM's proposal, exactly as before this
// change (bugs_open/050 owns the deployed-empty classification).
//
// Returns false when build_status is absent, so the preservation set degrades
// safely to the first-plan set on a chassis whose load_existing_pages query
// has not been updated to surface the column.
func realisedPageCompositionIsPreserved(rm map[string]interface{}) bool {
	switch status, _ := rm["build_status"].(string); status {
	case "deployed", "needs_rebuild":
		return true
	default:
		return false
	}
}

// recomposePagesFromSpec reads the OPTIONAL `recompose_pages` list from the
// needs_site_plan trigger spec, at input_data.spec.recompose_pages (the work
// item's spec travels there unchanged — see features_open/012). It names realised
// pages the caller has DELIBERATELY asked to redesign, the explicit-intent escape
// hatch for bugs_open/037: `needs_rebuild`/`deployed` pages are otherwise
// preserved, so this is the only way a re-plan may recompose one. Returns nil when
// the field is absent (every ordinary re-plan), so behaviour is unchanged unless a
// caller opts in. Names match realised page names (pages.name), case-sensitive as
// elsewhere in the planner.
func recomposePagesFromSpec(collectedData map[string]interface{}, logger *zap.Logger) map[string]bool {
	raw := datahelpers.ExtractNestedField(collectedData, "input_data.spec.recompose_pages")
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}
	set := make(map[string]bool, len(list))
	for _, v := range list {
		if name, ok := v.(string); ok && name != "" {
			set[name] = true
		}
	}
	if len(set) == 0 {
		return nil
	}
	names := make([]string, 0, len(set))
	for n := range set {
		names = append(names, n)
	}
	logger.Info("ValidateSitePlanAction: recompose_pages requested (explicit redesign intent)",
		zap.Strings("pages", names))
	return set
}

// filterOutRecomposePages drops every realised page named in the recompose set
// from the convergence input, so reconcilePlanWithRealised — and the truncation
// must-keep, which reads the same slice — treat those pages as from-scratch: the
// LLM's proposed composition governs, and the page may be redesigned or dropped
// per the LLM's plan. A named page matching no realised page is a harmless no-op.
func filterOutRecomposePages(existingPages []interface{}, recompose map[string]bool, logger *zap.Logger) []interface{} {
	if len(recompose) == 0 {
		return existingPages
	}
	kept := make([]interface{}, 0, len(existingPages))
	var released []string
	for _, rp := range existingPages {
		if rm, ok := rp.(map[string]interface{}); ok {
			if name, _ := rm["name"].(string); recompose[name] {
				released = append(released, name)
				continue
			}
		}
		kept = append(kept, rp)
	}
	if len(released) > 0 {
		logger.Info("ValidateSitePlanAction: recompose — realised pages released from the preserve guard for this re-plan",
			zap.Strings("pages", released))
	}
	return kept
}

// realisedSectionsOf extracts a realised page's section list. The
// load_existing_pages query runs via query_database, which stringifies jsonb, so
// "sections" normally arrives as a JSON string; a native []interface{} is
// tolerated too. Returns nil for missing, empty, or unparseable values — callers
// treat nil/empty as "no realised composition to preserve".
func realisedSectionsOf(rm map[string]interface{}) []interface{} {
	switch v := rm["sections"].(type) {
	case []interface{}:
		return v
	case string:
		if v == "" {
			return nil
		}
		var parsed []interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil
		}
		return parsed
	}
	return nil
}

// sameSectionList reports whether an LLM-proposed sections value already equals
// the realised one, so Pass B2 only logs and counts a snap-back that actually
// changed something.
func sameSectionList(proposed interface{}, realised []interface{}) bool {
	pl, ok := proposed.([]interface{})
	if !ok || len(pl) != len(realised) {
		return false
	}
	for i := range pl {
		if fmt.Sprintf("%v", pl[i]) != fmt.Sprintf("%v", realised[i]) {
			return false
		}
	}
	return true
}

// truncatePreservingRealised caps the plan at maxPages but never drops a
// must-keep page — on the site's first plan or built (see the caller). Must-keep
// pages are kept first; net-new proposed pages fill the remaining budget in order.
func truncatePreservingRealised(
	pages, mustKeep []interface{},
	maxPages int,
	logger *zap.Logger,
) []interface{} {
	keepNames := make(map[string]bool)
	for _, rp := range mustKeep {
		if rm, ok := rp.(map[string]interface{}); ok {
			if n, _ := rm["name"].(string); n != "" {
				keepNames[n] = true
			}
		}
	}
	var keep, netNew []interface{}
	for _, p := range pages {
		name := ""
		if pm, ok := p.(map[string]interface{}); ok {
			name, _ = pm["name"].(string)
		}
		if keepNames[name] {
			keep = append(keep, p)
		} else {
			netNew = append(netNew, p)
		}
	}
	if len(keep) >= maxPages {
		logger.Warn("validate: preserved pages exceed max_pages; keeping all preserved, dropping all net-new",
			zap.Int("preserved", len(keep)), zap.Int("max_pages", maxPages))
		return keep
	}
	budget := maxPages - len(keep)
	if budget > len(netNew) {
		budget = len(netNew)
	}
	return append(keep, netNew[:budget]...)
}

// ----------------------------------------------------------------------------
// LLM array-item key reconciliation (page-content-writer safety net)
//
// Array components declare the per-element field names in their input_schema
// (items / item_schema); plan_sections carries those onto each llm_field_spec
// as ItemFields, and the prompt now asks the LLM for exactly those keys. This
// is the belt-and-braces second line: if the model still emits a different
// spelling (title/body where the template reads name/description), repair it
// before it reaches the template or is persisted. See plan_sections_action.go.
// ----------------------------------------------------------------------------

// itemKeySynonyms groups field names that mean the same thing across component
// templates. Matching is case- and separator-insensitive, so "Title" and
// "card_title" both match "title". Keep common spellings first; extend a group
// rather than scattering new pairs elsewhere.
var itemKeySynonyms = [][]string{
	{"title", "name", "heading", "header", "label", "headline"},
	{"description", "body", "text", "content", "desc", "detail", "details", "summary", "copy", "caption"},
	{"icon_svg", "icon", "image", "img", "svg"},
	{"url", "href", "link"},
	{"cta_text", "cta", "button_text", "button_label", "action_text"},
}

func synonymsFor(key string) []string {
	for _, group := range itemKeySynonyms {
		for _, k := range group {
			if k == key {
				return group
			}
		}
	}
	return nil
}

func normaliseKeyForMatch(s string) string {
	return strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(s))
}

// expectedItemFieldsFromComponentSchema reads the component's own input_schema
// and returns, per source:"llm" array field that declares an item shape, the
// field names each element must contain. The schema is the authoritative
// contract the html_template is built against and is reloaded fresh on every
// render, so reconciliation no longer depends on the section plan carrying
// item_fields or on the prompt. Scoped to source:"llm" so the reconciler's
// reach matches the writer loop; query-resolved/static arrays (already keyed
// correctly by the system) are left untouched. Reuses extractArrayItemFields
// (plan_sections_action.go) so item-field extraction stays identical to how the
// plan and prompt derive it. Empty — reconcile becomes a no-op — when the
// schema has no fields map or no llm array fields (e.g. render_component called
// outside the writer loop, or on a non-array component).
func expectedItemFieldsFromComponentSchema(inputSchema map[string]interface{}) map[string][]string {
	out := map[string][]string{}
	fields, ok := inputSchema["fields"].(map[string]interface{})
	if !ok {
		return out
	}
	for name, raw := range fields {
		fieldDef, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if src, _ := fieldDef["source"].(string); src != "llm" {
			continue
		}
		if itemFields := extractArrayItemFields(fieldDef); len(itemFields) > 0 {
			out[name] = itemFields
		}
	}
	return out
}

// reconcileGeneratedItemKeys repairs LLM array output whose per-item keys don't
// match the keys the component template reads. Per element: an expected key
// that is missing but present under a case/separator variant or a known synonym
// is moved onto the expected key (WARN, so disobedience is visible); an expected
// key still missing afterwards is logged at ERROR with the element's actual
// keys, so a silent empty card cannot pass unseen. Modifies content in place.
// Non-fatal by design.
func reconcileGeneratedItemKeys(content map[string]interface{}, expected map[string][]string, componentFn string, logger *zap.Logger) {
	if len(content) == 0 || len(expected) == 0 {
		return
	}
	for fieldName, wantFields := range expected {
		raw, present := content[fieldName]
		if !present {
			continue
		}
		items, ok := raw.([]interface{})
		if !ok {
			logger.Warn("reconcileGeneratedItemKeys: array field is not a list — skipping",
				zap.String("component", componentFn), zap.String("field", fieldName),
				zap.String("got_type", fmt.Sprintf("%T", raw)))
			continue
		}
		wantSet := make(map[string]bool, len(wantFields))
		for _, w := range wantFields {
			wantSet[w] = true
		}
		for idx, itemRaw := range items {
			item, ok := itemRaw.(map[string]interface{})
			if !ok {
				logger.Warn("reconcileGeneratedItemKeys: array element is not an object — skipping",
					zap.String("component", componentFn), zap.String("field", fieldName),
					zap.Int("index", idx), zap.String("got_type", fmt.Sprintf("%T", itemRaw)))
				continue
			}
			norm := make(map[string]string, len(item))
			for k := range item {
				norm[normaliseKeyForMatch(k)] = k
			}
			for _, want := range wantFields {
				if _, has := item[want]; has {
					continue
				}
				wantNorm := normaliseKeyForMatch(want)
				if actual, ok := norm[wantNorm]; ok && actual != want {
					item[want] = item[actual]
					delete(item, actual)
					norm[wantNorm] = want
					logger.Warn("reconcileGeneratedItemKeys: normalised LLM item key",
						zap.String("component", componentFn), zap.String("field", fieldName),
						zap.Int("index", idx), zap.String("from", actual), zap.String("to", want))
					continue
				}
				remapped := false
				for _, syn := range synonymsFor(want) {
					if syn == want || wantSet[syn] {
						continue
					}
					if actual, ok := norm[normaliseKeyForMatch(syn)]; ok {
						item[want] = item[actual]
						delete(item, actual)
						norm[wantNorm] = want
						logger.Warn("reconcileGeneratedItemKeys: remapped LLM item key",
							zap.String("component", componentFn), zap.String("field", fieldName),
							zap.Int("index", idx), zap.String("from", actual), zap.String("to", want))
						remapped = true
						break
					}
				}
				if !remapped {
					logger.Error("reconcileGeneratedItemKeys: expected item field missing and unrecoverable",
						zap.String("component", componentFn), zap.String("field", fieldName),
						zap.Int("index", idx), zap.String("expected_key", want),
						zap.Any("item_keys", datahelpers.GetMapKeys(item)))
				}
			}
		}
	}
}
