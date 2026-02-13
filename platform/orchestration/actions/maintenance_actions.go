// FILE: platform/orchestration/actions/maintenance_actions.go
// Maintenance agent actions for page rebuilds, content refresh, etc.
//
// LoadSiteForRebuildAction provides supplementary context that
// the page-content-writer needs but that ensure_site_record and
// select_style_collection don't cover:
//   - reviewed_brief (extracted from sites.content_data)
//   - db_sync (navigation + all active pages for link context)
//   - brand asset URLs (logo, hero from content_data)
//   - site_plan (from content_data, stored during original build)
//
// Designed to run AFTER ensure_site_record so site_record is already
// in collectedData. Reads from that rather than re-querying.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// LoadSiteForRebuildAction loads the supplementary context needed by
// page-content-writer when rebuilding pages on an existing site.
//
// Expects in collectedData:
//   - site_record.site_id (from ensure_site_record)
//   - site_record.content_data (from ensure_site_record)
//
// Config:
//   - site_id_field: override path to site_id (default: "site_record.site_id")
//   - task_id_field: optional maintenance_queue task_id (future use)
//
// Returns:
//   - reviewed_brief: company info extracted from content_data
//   - site_plan: plan data from content_data (sections, page structure)
//   - db_sync: { navigation: {...}, pages: [...all active pages...] }
//   - logo_url: brand logo URL if available
//   - hero_url: hero image URL if available
//   - task_id: echoed back if provided (for future queue integration)
func LoadSiteForRebuildAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadSiteForRebuildAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required for rebuild context loading")
	}

	config := params.StepConfig.Config

	// --- Extract site_id from collectedData (set by prior ensure_site_record step) ---
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s — ensure_site_record must run first", siteIDField)
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}

	// --- Extract content_data from collectedData ---
	// ensure_site_record returns content_data as site_record.content_data
	contentData, _ := datahelpers.ExtractNestedField(params.CollectedData, "site_record.content_data").(map[string]interface{})
	if contentData == nil {
		// Fallback: load directly from DB
		params.Logger.Warn("LoadSiteForRebuildAction: content_data not in collectedData, loading from DB")
		contentData, err = loadContentDataFromDB(ctx, params.DB, siteID)
		if err != nil {
			return nil, fmt.Errorf("failed to load content_data: %w", err)
		}
	}

	// --- Build reviewed_brief from content_data ---
	// During original build, store_reviewed_brief merged brief fields into content_data
	// at the top level (company_name, tagline, services, contact_email, etc.)
	// The content writer expects reviewed_brief as a map with these fields.
	// content_data IS the brief (plus plan data merged in). Pass it as-is.
	reviewedBrief := contentData

	// --- Build site_plan from content_data ---
	// store_site_plan also merged plan data into content_data.
	// The content writer uses site_plan for page structure context.
	sitePlan := contentData

	// --- Extract brand asset URLs ---
	logoURL, _ := contentData["logo_url"].(string)
	heroURL, _ := contentData["hero_url"].(string)

	// --- Load all active pages for link context (db_sync.pages) ---
	allPages, err := loadActivePagesForLinkContext(ctx, params.DB, siteID, params.Logger)
	if err != nil {
		params.Logger.Warn("LoadSiteForRebuildAction: Failed to load pages for link context",
			zap.Error(err))
		allPages = []interface{}{}
	}

	// --- Load navigation structure ---
	var navResult interface{}
	nav, err := GetNavigationStructure(ctx, params.DB, siteID, "header", params.Logger)
	if err != nil {
		params.Logger.Warn("LoadSiteForRebuildAction: Failed to load navigation", zap.Error(err))
		navResult = map[string]interface{}{"items": []interface{}{}}
	} else {
		// Convert to map for JSON compatibility in collected_data
		navJSON, _ := json.Marshal(nav)
		var navMap map[string]interface{}
		json.Unmarshal(navJSON, &navMap)
		navResult = navMap
	}

	// --- Build db_sync equivalent ---
	// This mimics what sync_pages_to_db returns, providing the navigation
	// and pages list that prepare_link_context expects
	dbSync := map[string]interface{}{
		"pages_synced": len(allPages),
		"navigation":   navResult,
		"pages":        allPages,
		"site_id":      siteIDStr,
		"db_available": true,
	}

	// --- Optional: task_id for future maintenance queue integration ---
	taskIDField := "input_data.task_id"
	if f, ok := config["task_id_field"].(string); ok && f != "" {
		taskIDField = f
	}
	taskID := datahelpers.ExtractNestedFieldString(params.CollectedData, taskIDField)

	params.Logger.Info("LoadSiteForRebuildAction: Context loaded",
		zap.String("site_id", siteIDStr),
		zap.Int("pages_for_linking", len(allPages)),
		zap.Bool("has_logo", logoURL != ""),
		zap.Bool("has_hero", heroURL != ""),
		zap.String("task_id", taskID),
	)

	result := map[string]interface{}{
		"reviewed_brief": reviewedBrief,
		"site_plan":      sitePlan,
		"db_sync":        dbSync,
		"logo_url":       logoURL,
		"hero_url":       heroURL,
		"site_id":        siteIDStr,
	}
	if taskID != "" {
		result["task_id"] = taskID
	}

	return result, nil
}

// loadContentDataFromDB loads content_data directly from sites table.
// Fallback when ensure_site_record didn't provide it in collectedData.
func loadContentDataFromDB(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]interface{}, error) {
	var contentJSON []byte
	err := db.QueryRowContext(ctx,
		`SELECT COALESCE(content_data, '{}'::jsonb) FROM sites WHERE id = $1`,
		siteID,
	).Scan(&contentJSON)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(contentJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content_data: %w", err)
	}
	return result, nil
}

// loadActivePagesForLinkContext queries all active pages for a site.
// Returns them in the format that extractPagesForLinking expects:
// []interface{} of map[string]interface{} with name, url, title, description.
func loadActivePagesForLinkContext(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]interface{}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name, url, COALESCE(title, name) as title,
		       COALESCE(meta_description, '') as description,
		       COALESCE(nav_order, 100) as nav_order
		FROM pages
		WHERE site_id = $1 AND status = 'active'
		ORDER BY nav_order ASC, name ASC
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}
	defer rows.Close()

	var pages []interface{}
	for rows.Next() {
		var name, url, title, description string
		var navOrder int
		if err := rows.Scan(&name, &url, &title, &description, &navOrder); err != nil {
			logger.Warn("loadActivePagesForLinkContext: Failed to scan page", zap.Error(err))
			continue
		}
		// Build URL from name if missing (same logic as extractPagesForLinking)
		if url == "" {
			if name == "index" || name == "home" {
				url = "/index.html"
			} else {
				url = "/" + name + ".html"
			}
		}
		// Title fallback
		if title == name {
			title = strings.Title(strings.ReplaceAll(name, "-", " "))
		}

		pages = append(pages, map[string]interface{}{
			"name":        name,
			"url":         url,
			"title":       title,
			"description": description,
			"nav_order":   navOrder,
		})
	}

	logger.Info("loadActivePagesForLinkContext: Loaded pages",
		zap.Int("count", len(pages)),
		zap.String("site_id", siteID.String()),
	)

	return pages, nil
}
