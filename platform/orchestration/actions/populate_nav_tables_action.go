// FILE: platform/orchestration/actions/populate_nav_tables_action.go
//
// PopulateNavTablesAction reads pages from the database (after sync_pages_to_db)
// and populates site_nav_groups + site_nav_items. Runs once during the build
// workflow, after pages are synced but before content writing or rendering.
//
// Standalone mode:  { "site_id": "uuid" }
// Integrated mode:  called from pageflow-builder with input_mapping
//
// This replaces the implicit nav derivation that previously happened at query
// time via scattered functions. Nav structure is now explicit and stored.

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Input spec and registration
// ---------------------------------------------------------------------------

var PopulateNavTablesInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"max_header_items"},
	Defaults: map[string]interface{}{
		"max_header_items": 6,
	},
	Deprecated: map[string]string{
		"site_id_field": "site_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("populate_nav_tables", PopulateNavTablesInputSpec)
}

// ---------------------------------------------------------------------------
// pageNavInfo — fields we need from a page record to classify into nav groups
// ---------------------------------------------------------------------------

type pageNavInfo struct {
	ID       uuid.UUID
	Name     string
	Title    string
	URL      string
	NavLabel string
	NavOrder int
	InHeader bool
	InFooter bool
}

// ---------------------------------------------------------------------------
// Action entry point
// ---------------------------------------------------------------------------

// PopulateNavTablesAction reads pages for this site and populates nav tables.
//
// Input contract:
//   - site_id       (required) - UUID of the site
//   - max_header_items (optional, default 6) - max items in primary nav group
//
// Output contract:
//   - success    bool
//   - site_id    string
//   - groups     int    - number of nav groups created
//   - items      int    - total nav items created
//   - navigation *NavigationStructure - header nav in the standard shape
//     for downstream consumers (content writer, webdesign agent)
func PopulateNavTablesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("PopulateNavTablesAction: Starting")

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// Handle initialization
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// --- Standardised input extraction ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		PopulateNavTablesInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w (got %q)", err, siteIDStr)
	}

	maxHeaderItems := inputs.GetInt("max_header_items", 6)

	// --- Load pages ---
	pages, err := loadPagesForNav(ctx, params.DB, siteID, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load pages: %w", err)
	}

	if len(pages) == 0 {
		logger.Warn("PopulateNavTablesAction: No pages found")
		return map[string]interface{}{
			"success":    true,
			"site_id":    siteIDStr,
			"groups":     0,
			"items":      0,
			"navigation": &NavigationStructure{Items: []NavigationItem{}},
		}, nil
	}

	// --- Classify and populate ---
	primaryPages, legalPages, utilityPages := classifyPagesForNav(pages, logger)

	if len(primaryPages) > maxHeaderItems {
		primaryPages = primaryPages[:maxHeaderItems]
	}

	// Upsert in a transaction (full rebuild each time)
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err = tx.ExecContext(ctx, `DELETE FROM site_nav_items WHERE site_id = $1`, siteID); err != nil {
		return nil, fmt.Errorf("failed to clear nav items: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM site_nav_groups WHERE site_id = $1`, siteID); err != nil {
		return nil, fmt.Errorf("failed to clear nav groups: %w", err)
	}

	totalItems := 0
	groupCount := 0

	if len(primaryPages) > 0 {
		gid, err := upsertNavGroup(ctx, tx, siteID, "primary", "Main Navigation", NavGroupPrimary, 0)
		if err != nil {
			return nil, fmt.Errorf("failed to create primary group: %w", err)
		}
		for i, page := range primaryPages {
			if err := insertNavItem(ctx, tx, siteID, gid, page, i); err != nil {
				logger.Warn("Failed to insert primary nav item", zap.String("page", page.Name), zap.Error(err))
				continue
			}
			totalItems++
		}
		groupCount++
	}

	if len(legalPages) > 0 {
		gid, err := upsertNavGroup(ctx, tx, siteID, "legal", "Legal", NavGroupLegal, 10)
		if err != nil {
			return nil, fmt.Errorf("failed to create legal group: %w", err)
		}
		for i, page := range legalPages {
			if err := insertNavItem(ctx, tx, siteID, gid, page, i); err != nil {
				logger.Warn("Failed to insert legal nav item", zap.String("page", page.Name), zap.Error(err))
				continue
			}
			totalItems++
		}
		groupCount++
	}

	if len(utilityPages) > 0 {
		gid, err := upsertNavGroup(ctx, tx, siteID, "utility", "Utility", NavGroupUtility, 5)
		if err != nil {
			return nil, fmt.Errorf("failed to create utility group: %w", err)
		}
		for i, page := range utilityPages {
			if err := insertNavItem(ctx, tx, siteID, gid, page, i); err != nil {
				logger.Warn("Failed to insert utility nav item", zap.String("page", page.Name), zap.Error(err))
				continue
			}
			totalItems++
		}
		groupCount++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit nav tables: %w", err)
	}

	logger.Info("PopulateNavTablesAction: Complete",
		zap.Int("groups", groupCount),
		zap.Int("items", totalItems),
		zap.Int("primary", len(primaryPages)),
		zap.Int("legal", len(legalPages)),
		zap.Int("utility", len(utilityPages)),
	)

	// Build navigation structure for downstream (same JSON shape as db_sync.navigation)
	nav := buildNavStructureFromClassified(primaryPages)

	return map[string]interface{}{
		"success":    true,
		"site_id":    siteIDStr,
		"groups":     groupCount,
		"items":      totalItems,
		"navigation": nav,
	}, nil
}

// ---------------------------------------------------------------------------
// Page loading and classification
// ---------------------------------------------------------------------------

func loadPagesForNav(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]pageNavInfo, error) {
	query := `
		SELECT
			id,
			name,
			COALESCE(title, name) as title,
			COALESCE(url, '/' || name || '.html') as url,
			COALESCE(nav_label, '') as nav_label,
			COALESCE(nav_order, 0) as nav_order,
			COALESCE(in_header, false) as in_header,
			COALESCE(in_footer, false) as in_footer
		FROM pages
		WHERE site_id = $1
		  AND status IN ('active', 'deployed', 'pending')
		  AND deleted_at IS NULL
		ORDER BY nav_order ASC, created_at ASC
	`

	rows, err := db.QueryContext(ctx, query, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []pageNavInfo
	for rows.Next() {
		var p pageNavInfo
		if err := rows.Scan(&p.ID, &p.Name, &p.Title, &p.URL, &p.NavLabel, &p.NavOrder, &p.InHeader, &p.InFooter); err != nil {
			logger.Warn("loadPagesForNav: scan failed", zap.Error(err))
			continue
		}
		pages = append(pages, p)
	}
	return pages, nil
}

// classifyPagesForNav sorts pages into primary, legal, and utility groups.
func classifyPagesForNav(pages []pageNavInfo, logger *zap.Logger) (primary, legal, utility []pageNavInfo) {
	legalNames := map[string]bool{
		"privacy": true, "terms": true, "cookies": true,
		"disclaimer": true, "privacy-policy": true, "terms-of-service": true,
		"cookie-policy": true, "terms-and-conditions": true,
	}

	systemNames := map[string]bool{
		"404": true, "sitemap": true, "robots": true,
	}

	for _, page := range pages {
		nameLower := strings.ToLower(page.Name)

		if systemNames[nameLower] {
			continue
		}

		if legalNames[nameLower] || isLegalPage(nameLower) {
			legal = append(legal, page)
			continue
		}

		if page.InHeader {
			primary = append(primary, page)
		}

		// Footer-only (not header, not legal) → utility
		if page.InFooter && !page.InHeader {
			utility = append(utility, page)
		}
	}

	logger.Debug("classifyPagesForNav",
		zap.Int("primary", len(primary)),
		zap.Int("legal", len(legal)),
		zap.Int("utility", len(utility)),
	)
	return
}

func isLegalPage(nameLower string) bool {
	for _, prefix := range []string{"privacy", "terms", "cookie", "disclaimer", "legal"} {
		if strings.HasPrefix(nameLower, prefix) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// DB operations
// ---------------------------------------------------------------------------

func upsertNavGroup(ctx context.Context, tx *sql.Tx, siteID uuid.UUID, groupKey, groupLabel, groupType string, position int) (uuid.UUID, error) {
	var groupID uuid.UUID
	err := tx.QueryRowContext(ctx, `
		INSERT INTO site_nav_groups (site_id, group_key, group_label, group_type, position)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (site_id, group_key) DO UPDATE SET
			group_label = EXCLUDED.group_label,
			group_type = EXCLUDED.group_type,
			position = EXCLUDED.position,
			updated_at = NOW()
		RETURNING id
	`, siteID, groupKey, groupLabel, groupType, position).Scan(&groupID)
	return groupID, err
}

func insertNavItem(ctx context.Context, tx *sql.Tx, siteID, groupID uuid.UUID, page pageNavInfo, position int) error {
	label := navLabelForPage(page)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position)
		VALUES ($1, $2, $3, $4, $5, 'page_link', $6)
	`, siteID, groupID, label, page.URL, page.ID, position)
	return err
}

func navLabelForPage(page pageNavInfo) string {
	if page.NavLabel != "" {
		return navSimplifyLabel(page.NavLabel, page.URL)
	}
	return navSimplifyLabel(page.Title, page.URL)
}

// buildNavStructureFromClassified creates a NavigationStructure from primary pages.
// The JSON shape {items: [{label, url, page_id}]} is consumed by:
//   - extractNavigationFromDBSync (content writer prompt)
//   - extractNavItemsForHeader    (webdesign agent)
func buildNavStructureFromClassified(primaryPages []pageNavInfo) *NavigationStructure {
	nav := &NavigationStructure{
		Items: make([]NavigationItem, len(primaryPages)),
	}
	for i, page := range primaryPages {
		nav.Items[i] = NavigationItem{
			PageID: page.ID.String(),
			Label:  navLabelForPage(page),
			URL:    page.URL,
		}
	}
	return nav
}
