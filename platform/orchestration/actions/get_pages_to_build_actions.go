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

	// PAGE-OWNERSHIP EXCLUSION (bugs_open/208). A rebuild_policy='owned' page
	// belongs to a tool/widget or is a runtime-fill shell, so it is NOT the
	// generic builder's to compose — and every consumer of this action feeds a
	// loop whose next steps are assemble_page -> deploy_page (git_commit), which
	// replaces the served file BEFORE save_page_sections gets to refuse the save.
	// Excluding here is what keeps the destructive state unreachable rather than
	// caught after the commit.
	//
	// The unsafe branch is the one that must be asked for by name (owner ruling
	// 2026-08-02 §2: a seam licensed by "callers must all be X" gets a field with
	// the unsafe default OFF, because a comment is not a control on a tree this
	// many sessions share). Nothing live sets it: measured 2026-08-06, both live
	// consumers (page-rebuild, pageflow-builder) carry only include_all and
	// build_statuses.
	includeOwned := datahelpers.GetBoolField(config, "include_owned", false)

	// Query pages from database
	pages, err := queryPagesForBuild(ctx, params.DB, siteID, statusFilter, includeAll, includeOwned, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}

	// Name what was excluded, and record it where an operator will find it. A
	// silent exclusion would leave the page at needs_rebuild for ever with nobody
	// told why it never rebuilt.
	var excludedOwned []string
	if !includeOwned {
		excludedOwned = censusExcludedRefusedPages(ctx, params.DB, siteID, statusFilter, includeAll,
			"get_pages_to_build", logger)
	}

	logger.Info("GetPagesToBuildAction: Complete",
		zap.Int("pages_found", len(pages)),
		zap.String("site_id", siteIDStr),
		zap.Strings("status_filter", statusFilter),
		zap.Bool("include_owned", includeOwned),
		zap.Int("owned_pages_excluded", len(excludedOwned)),
	)

	// Convert to interface slice for loop iteration
	// This becomes the array that build_pages_loop iterates over
	pagesInterface := make([]interface{}, len(pages))
	for i, p := range pages {
		pagesInterface[i] = p
	}

	return map[string]interface{}{
		"pages":      pagesInterface, // items_field: "pages_to_build.pages" points here
		"page_count": len(pages),
		// Observability, not control flow: a reader of the run can see that pages
		// were withheld and which ones, without joining back to the pages table.
		"owned_pages_excluded":       excludedOwned,
		"owned_pages_excluded_count": len(excludedOwned),
		"site_id":                    siteIDStr,
		"filter_used":                statusFilter,
	}, nil
}

// ownedPageExclusionSQL is the ownership predicate shared by both branches of
// queryPagesForBuild (bugs_open/208).
//
// COALESCE is deliberate even though pages.rebuild_policy is NOT NULL DEFAULT
// 'generic' today (migration 164): a bare `rebuild_policy <> 'owned'` would drop
// every row with a NULL policy if that default is ever relaxed, which is the same
// shape of silent-omission bug this predicate exists to prevent.
//
// Kept BYTE-IDENTICAL when the tool-shell arm was added (bugs_open/450): the new
// disjunct is appended by genericBuildExclusionSQL rather than folded in here,
// so anything pinning this literal keeps matching and the owned arm can be read
// on its own.
const ownedPageExclusionSQL = ` AND COALESCE(rebuild_policy, 'generic') <> 'owned'`

// genericBuildExclusionSQL is what the selection actually applies: the ownership
// exclusion above, plus — while the arm is armed — the tool-shell exclusion from
// owned_page_guard.go. Together they are the SQL spelling of
// genericBuildRefusal, and censusExcludedRefusedPages is their exact inverse.
//
// A tool page with no tool must not be handed to a generic build loop for the
// same reason an owned page must not: what comes back is prose where an
// interactive tool belongs. The difference is only that this one is derived and
// therefore lifts by itself when the tool arrives.
func genericBuildExclusionSQL() string {
	if !toolShellRefusalArmed() {
		return ownedPageExclusionSQL
	}
	return ownedPageExclusionSQL + ` AND NOT ` + toolShellPredicateFor("pages")
}

// queryPagesForBuild queries pages that need building from the database
// Returns page maps with fields needed by content writer and HTML builder
//
// includeOwned=false (the default for every caller) excludes rebuild_policy='owned'
// pages: they are tool/widget-owned and must never be composed by a generic
// builder. See GetPagesToBuildAction for why the exclusion lives at selection.
func queryPagesForBuild(ctx context.Context, db interface{}, siteID uuid.UUID, statuses []string, includeAll bool, includeOwned bool, logger *zap.Logger) ([]map[string]interface{}, error) {
	var query string
	var args []interface{}

	// Applied to BOTH branches. include_all drops the status filter, so without
	// this it would sweep every owned page on the site, including the ~189 sitting
	// at 'deployed' (measured 2026-08-06) — a wider blast radius than the
	// needs_rebuild case that surfaced the bug.
	ownershipClause := genericBuildExclusionSQL()
	if includeOwned {
		ownershipClause = ""
	}

	if includeAll {
		query = fmt.Sprintf(`
			SELECT id, site_id, name, url, title, page_type, status,
			       COALESCE(build_status, 'planned') as build_status,
			       COALESCE(sections, '[]'::jsonb) as sections,
			       nav_label, nav_order, in_header, in_footer,
			       COALESCE(version, 1) as version,
			       meta_description,
			       content_direction::text
			FROM pages
			WHERE site_id = $1 AND status = 'active'%s
			ORDER BY nav_order ASC, name ASC
		`, ownershipClause)
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
			       meta_description,
			       content_direction::text
			FROM pages
			WHERE site_id = $1
			  AND status = 'active'
			  AND COALESCE(build_status, 'planned') IN (%s)%s
			ORDER BY nav_order ASC, name ASC
		`, strings.Join(placeholders, ", "), ownershipClause)
	}

	logger.Debug("Querying pages for build",
		zap.String("site_id", siteID.String()),
		zap.Int("status_count", len(statuses)),
		zap.Bool("include_owned", includeOwned),
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
			id               uuid.UUID
			siteID           uuid.UUID
			name             string
			url              string
			title            string
			pageType         string
			status           string
			buildStatus      string
			sectionsJSON     []byte
			navLabel         sql.NullString
			navOrder         int
			inHeader         bool
			inFooter         bool
			version          int
			metaDescription  sql.NullString
			contentDirection sql.NullString
		)

		err := rows.Scan(
			&id, &siteID, &name, &url, &title, &pageType, &status,
			&buildStatus, &sectionsJSON,
			&navLabel, &navOrder, &inHeader, &inFooter,
			&version, &metaDescription, &contentDirection,
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
			"page_type":        pageType,    // "landing", "content", "tool", "blog-post", ...
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

		// Per-page content_direction (bug 025): optional writer steering, nullable jsonb.
		// Only set when present so the writer's {{if .current_page.content_direction}} guard
		// stays false for pages without it.
		if contentDirection.Valid {
			addContentDirection(page, contentDirection.String, logger)
		}

		pages = append(pages, page)
	}

	return pages, rows.Err()
}

// addContentDirection parses the raw per-page content_direction jsonb text and, when
// it is a non-null value, stores it under "content_direction" on the page map so it
// reaches the writer as .current_page.content_direction (bug 025). A parse failure is
// logged and skipped — the page still builds, just without the extra steering.
func addContentDirection(page map[string]interface{}, raw string, logger *zap.Logger) {
	if raw == "" {
		return
	}
	var cd interface{}
	if err := json.Unmarshal([]byte(raw), &cd); err != nil {
		logger.Warn("Failed to parse page content_direction JSON",
			zap.String("content_direction_raw", raw), zap.Error(err))
		return
	}
	if cd != nil {
		page["content_direction"] = cd
	}
}

// scanPageRowsForBuildPgx scans pgx rows into page maps
func scanPageRowsForBuildPgx(rows pgx.Rows, logger *zap.Logger) ([]map[string]interface{}, error) {
	var pages []map[string]interface{}

	for rows.Next() {
		var (
			id               uuid.UUID
			siteID           uuid.UUID
			name             string
			url              string
			title            string
			pageType         string
			status           string
			buildStatus      string
			sectionsJSON     []byte
			navLabel         *string
			navOrder         int
			inHeader         bool
			inFooter         bool
			version          int
			metaDescription  *string
			contentDirection *string
		)

		err := rows.Scan(
			&id, &siteID, &name, &url, &title, &pageType, &status,
			&buildStatus, &sectionsJSON,
			&navLabel, &navOrder, &inHeader, &inFooter,
			&version, &metaDescription, &contentDirection,
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

		// Per-page content_direction (bug 025): optional writer steering, nullable jsonb.
		if contentDirection != nil {
			addContentDirection(page, *contentDirection, logger)
		}

		pages = append(pages, page)
	}

	return pages, rows.Err()
}
