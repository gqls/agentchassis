package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// GetPagesToBuildAction queries pages from DB that need content generation
func GetPagesToBuildAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "get_pages_to_build"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	logger.Info("GetPagesToBuildAction: Starting")

	// Handle initialization
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	// Check DB availability
	if params.DB == nil {
		logger.Error("Database not available")
		return nil, fmt.Errorf("database not available - get_pages_to_build requires DB")
	}

	// Extract site_id using existing helper (from site_actions.go)
	siteIDStr := extractSiteID(params.CollectedData, logger)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found in collected data")
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get config options using existing datahelpers
	config := params.StepConfig.Config

	// Which statuses to include (default: planned, needs_rebuild)
	statusFilter := []string{"planned", "needs_rebuild"}
	if statuses, ok := config["build_statuses"].([]interface{}); ok {
		statusFilter = make([]string, len(statuses))
		for i, s := range statuses {
			statusFilter[i] = fmt.Sprintf("%v", s)
		}
	}

	// Whether to include all pages or just those needing build
	includeAll := datahelpers.GetBoolField(config, "include_all", false)

	// Query pages from database
	pages, err := queryPagesForBuild(ctx, params.DB, siteID, statusFilter, includeAll, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}

	logger.Info("GetPagesToBuildAction: Complete",
		zap.Int("pages_found", len(pages)),
		zap.String("site_id", siteIDStr),
		zap.Strings("status_filter", statusFilter),
	)

	// Convert to interface slice for loop iteration
	// This becomes the array that build_pages_loop iterates over
	pagesInterface := make([]interface{}, len(pages))
	for i, p := range pages {
		pagesInterface[i] = p
	}

	return map[string]interface{}{
		"pages":       pagesInterface, // items_field: "pages_to_build.pages" points here
		"page_count":  len(pages),
		"site_id":     siteIDStr,
		"filter_used": statusFilter,
	}, nil
}

// queryPagesForBuild queries pages that need building from the database
// Returns page maps with fields needed by content writer and HTML builder
func queryPagesForBuild(ctx context.Context, db interface{}, siteID uuid.UUID, statuses []string, includeAll bool, logger *zap.Logger) ([]map[string]interface{}, error) {
	var query string
	var args []interface{}

	if includeAll {
		query = `
			SELECT id, site_id, name, url, title, page_type, status, 
			       COALESCE(build_status, 'planned') as build_status,
			       COALESCE(sections, '[]'::jsonb) as sections,
			       nav_label, nav_order, in_header, in_footer,
			       COALESCE(version, 1) as version,
			       meta_description
			FROM pages 
			WHERE site_id = $1 AND status = 'active'
			ORDER BY nav_order ASC, name ASC
		`
		args = []interface{}{siteID}
	} else {
		// Build status filter with parameterized placeholders
		placeholders := make([]string, len(statuses))
		args = []interface{}{siteID}
		for i, s := range statuses {
			placeholders[i] = fmt.Sprintf("$%d", i+2)
			args = append(args, s)
		}

		query = fmt.Sprintf(`
			SELECT id, site_id, name, url, title, page_type, status,
			       COALESCE(build_status, 'planned') as build_status,
			       COALESCE(sections, '[]'::jsonb) as sections,
			       nav_label, nav_order, in_header, in_footer,
			       COALESCE(version, 1) as version,
			       meta_description
			FROM pages 
			WHERE site_id = $1 
			  AND status = 'active'
			  AND COALESCE(build_status, 'planned') IN (%s)
			ORDER BY nav_order ASC, name ASC
		`, strings.Join(placeholders, ", "))
	}

	logger.Debug("Querying pages for build",
		zap.String("site_id", siteID.String()),
		zap.Int("status_count", len(statuses)),
	)

	var pages []map[string]interface{}

	switch d := db.(type) {
	case *sql.DB:
		rows, err := d.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()
		pages, err = scanPageRowsForBuild(rows, logger)
		if err != nil {
			return nil, err
		}

	case *pgxpool.Pool:
		rows, err := d.Query(ctx, query, args...)
		if err != nil {
			return nil, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()
		pages, err = scanPageRowsForBuildPgx(rows, logger)
		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	return pages, nil
}

// scanPageRowsForBuild scans sql.Rows into page maps
// Output fields match what write_page_content and other substeps expect
func scanPageRowsForBuild(rows *sql.Rows, logger *zap.Logger) ([]map[string]interface{}, error) {
	var pages []map[string]interface{}

	for rows.Next() {
		var (
			id              uuid.UUID
			siteID          uuid.UUID
			name            string
			url             string
			title           string
			pageType        string
			status          string
			buildStatus     string
			sectionsJSON    []byte
			navLabel        sql.NullString
			navOrder        int
			inHeader        bool
			inFooter        bool
			version         int
			metaDescription sql.NullString
		)

		err := rows.Scan(
			&id, &siteID, &name, &url, &title, &pageType, &status,
			&buildStatus, &sectionsJSON,
			&navLabel, &navOrder, &inHeader, &inFooter,
			&version, &metaDescription,
		)
		if err != nil {
			logger.Error("Failed to scan page row", zap.Error(err))
			continue
		}

		// Parse sections JSON
		var sections []interface{}
		if err := json.Unmarshal(sectionsJSON, &sections); err != nil {
			sections = []interface{}{}
		}

		// Build page map with fields that current_page variable needs
		// These match what write_page_content expects in its input_fields
		page := map[string]interface{}{
			"id":               id.String(), // For DB updates
			"page_id":          id.String(), // Alias used by some actions
			"site_id":          siteID.String(),
			"name":             name,        // Page slug: "index", "about", etc.
			"page_name":        name,        // Alias: dispatch loop maps this to handler
			"url":              url,         // Full path: "/index.html"
			"title":            title,       // Page title for content
			"page_type":        pageType,    // "index", "content", "landing"
			"status":           status,      // "active", "draft"
			"build_status":     buildStatus, // "planned", "deployed"
			"sections":         sections,    // Section names to generate
			"nav_label":        navLabel.String,
			"nav_order":        navOrder,
			"in_header":        inHeader,
			"in_footer":        inFooter,
			"version":          version,
			"meta_description": metaDescription.String,
		}

		pages = append(pages, page)
	}

	return pages, rows.Err()
}

// scanPageRowsForBuildPgx scans pgx rows into page maps
func scanPageRowsForBuildPgx(rows pgx.Rows, logger *zap.Logger) ([]map[string]interface{}, error) {
	var pages []map[string]interface{}

	for rows.Next() {
		var (
			id              uuid.UUID
			siteID          uuid.UUID
			name            string
			url             string
			title           string
			pageType        string
			status          string
			buildStatus     string
			sectionsJSON    []byte
			navLabel        *string
			navOrder        int
			inHeader        bool
			inFooter        bool
			version         int
			metaDescription *string
		)

		err := rows.Scan(
			&id, &siteID, &name, &url, &title, &pageType, &status,
			&buildStatus, &sectionsJSON,
			&navLabel, &navOrder, &inHeader, &inFooter,
			&version, &metaDescription,
		)
		if err != nil {
			logger.Error("Failed to scan page row", zap.Error(err))
			continue
		}

		// Parse sections JSON
		var sections []interface{}
		if err := json.Unmarshal(sectionsJSON, &sections); err != nil {
			sections = []interface{}{}
		}

		// Handle nullable strings
		navLabelStr := ""
		if navLabel != nil {
			navLabelStr = *navLabel
		}
		metaDescStr := ""
		if metaDescription != nil {
			metaDescStr = *metaDescription
		}

		page := map[string]interface{}{
			"id":               id.String(),
			"page_id":          id.String(),
			"site_id":          siteID.String(),
			"name":             name,
			"page_name":        name,
			"url":              url,
			"title":            title,
			"page_type":        pageType,
			"status":           status,
			"build_status":     buildStatus,
			"sections":         sections,
			"nav_label":        navLabelStr,
			"nav_order":        navOrder,
			"in_header":        inHeader,
			"in_footer":        inFooter,
			"version":          version,
			"meta_description": metaDescStr,
		}

		pages = append(pages, page)
	}

	return pages, rows.Err()
}
