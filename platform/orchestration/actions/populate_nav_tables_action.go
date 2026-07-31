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
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Input spec and registration
// ---------------------------------------------------------------------------

var PopulateNavTablesInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"max_header_items"},
	Defaults: map[string]interface{}{
		"max_header_items": 8,
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
	PageType string
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
//   - max_header_items (optional, default 8) - max items in primary nav group
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

	maxHeaderItems := inputs.GetInt("max_header_items", 8)

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
		// Overflow primary items go to utility so they still appear in footer nav.
		// Without this, items beyond the limit simply vanish from all nav.
		overflowPages := primaryPages[maxHeaderItems:]
		primaryPages = primaryPages[:maxHeaderItems]
		utilityPages = append(overflowPages, utilityPages...)
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
			COALESCE(page_type, 'content') as page_type,
			COALESCE(in_header, false) as in_header,
			COALESCE(in_footer, false) as in_footer
		FROM pages
		WHERE site_id = $1
		  AND status IN ('active', 'deployed', 'pending')
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
		if err := rows.Scan(&p.ID, &p.Name, &p.Title, &p.URL, &p.NavLabel, &p.NavOrder, &p.PageType, &p.InHeader, &p.InFooter); err != nil {
			logger.Warn("loadPagesForNav: scan failed", zap.Error(err))
			continue
		}
		pages = append(pages, p)
	}
	return pages, nil
}

// classifyPagesForNav sorts pages into primary, legal, and utility groups.
//
// Primary nav is filled by priority tier — core pages first, then content hubs,
// then secondary pages. Pages that don't make the cut go to utility (footer).
// This produces sensible navigation across all sites without per-site tuning.
//
// Tier 1 (core): index, services, about, contact — the pages every site needs visible.
// Tier 2 (content hubs): blog, case-studies, use-cases, pricing, how-we-work, portfolio.
// Tier 3 (secondary): faq, approach, careers, guides, insights, news, resources.
// Tier 4 (never primary): individual tool pages, blog posts, guide pages, entity pages.
//
// Within each tier, pages are ordered by nav_order (from the planner), then creation date.
// Pages without in_header=true skip primary entirely and go to utility if in_footer=true.
func classifyPagesForNav(pages []pageNavInfo, logger *zap.Logger) (primary, legal, utility []pageNavInfo) {
	legalNames := map[string]bool{
		"privacy": true, "terms": true, "cookies": true,
		"disclaimer": true, "privacy-policy": true, "terms-of-service": true,
		"cookie-policy": true, "terms-and-conditions": true,
	}

	systemNames := map[string]bool{
		"404": true, "sitemap": true, "robots": true,
	}

	// isChildPageURL returns true for pages that live under a category path,
	// like /tools/something.html or /blog/something.html. These are child pages
	// regardless of their page_type and should never appear in primary nav —
	// the parent /tools.html or /blog.html represents them in navigation.
	isChildPageURL := func(url string) bool {
		lower := strings.ToLower(url)
		childPrefixes := []string{
			"/tools/", "/blog/", "/guides/", "/articles/",
			"/case-studies/", "/news/", "/resources/", "/insights/",
		}
		for _, prefix := range childPrefixes {
			if strings.HasPrefix(lower, prefix) {
				return true
			}
		}
		return false
	}

	// Page types that should never appear in primary nav regardless of in_header flag.
	// These are child pages that belong under a parent listing, not top-level nav.
	neverPrimaryTypes := map[string]bool{
		"blog-post": true, "tool": true, "entity-page": true,
	}

	// Tiered primary nav candidates — collected then sorted by tier
	type tieredPage struct {
		page pageNavInfo
		tier int // 1=core, 2=hub, 3=secondary
	}
	var candidates []tieredPage

	for _, page := range pages {
		nameLower := strings.ToLower(page.Name)

		if systemNames[nameLower] {
			continue
		}

		if legalNames[nameLower] || isLegalPage(nameLower) {
			legal = append(legal, page)
			continue
		}

		// A page can be barred from the PRIMARY nav two ways, and until
		// 2026-07-31 the two were answering different questions:
		//
		//   - by page_type   (blog-post, tool, entity-page) — went to utility if
		//     it had declared a nav flag, i.e. barred from the main menu but still
		//     reachable from the footer;
		//   - by URL shape   (/tools/, /blog/, /guides/ …) — `continue`d out
		//     entirely, before either flag was read.
		//
		// The URL-shaped rule ran first, so it decided WHETHER a page appeared at
		// all, which was never its job. Consequence, measured on 2026-07-31: a
		// `nav_drift` work item for gamesdesign.co.uk naming four /tools/ pages
		// completed on 07-29 and all four were still absent from site_nav_items two
		// days later — while the same check, handler and action repaired
		// robot-hands.com's /learning-center.html and /news.html, whose only
		// difference is the URL prefix. nav-updater rebuilds nav then re-renders
		// chrome, so it had everything it needed except a page to place. See
		// bugs_open/149 A2 and LANDMINES.md "nav-updater DELETES every nav link
		// whose URL sits under /tools/ …", found independently by the leopardess
		// lane the day before.
		//
		// So the two notions are one notion — NEVER PRIMARY — and the flags decide
		// presence for both. The invariant, pinned in nav_membership_test.go:
		// pages.in_header/in_footer declares nav membership; a page's URL shape may
		// decide WHERE it appears, never WHETHER it appears.
		//
		// EXCEPTION, unchanged: a section-index page (e.g. games-index at
		// /games/index.html) is the section's PARENT, not one of its children. It
		// sits under the section prefix but is primary-eligible. Keyed on page_type,
		// not URL, so tool/blog-post/entity-page leaves under the same prefix are
		// still correctly barred from primary.
		neverPrimary := neverPrimaryTypes[page.PageType] ||
			(isChildPageURL(page.URL) && !isSectionIndexType(page.PageType))

		if neverPrimary {
			if page.InHeader || page.InFooter {
				// Barred from the main menu, kept in the footer. Utility is where
				// the platform already put never-primary types that declared a
				// flag; child-path pages now land in the same place rather than
				// being dropped.
				utility = append(utility, page)
			} else {
				// No flag on either side: the page has declared no nav membership,
				// so leaving it out is the honest answer, not a silent drop. This
				// is the branch a tool page created with in_header=false takes, and
				// it is why A3 (nothing keeps a parent listing in sync) is still
				// open — a page in this branch is reachable only from its listing.
				logger.Info("classifyPagesForNav: page declares no nav membership, omitted from nav",
					zap.String("name", page.Name),
					zap.String("url", page.URL),
					zap.String("page_type", page.PageType))
			}
			continue
		}

		// Pages without in_header go to utility (if in_footer) or are skipped
		if !page.InHeader {
			if page.InFooter {
				utility = append(utility, page)
			}
			continue
		}

		// Assign tier based on page name and type
		tier := navPriorityTier(nameLower, page.PageType)
		candidates = append(candidates, tieredPage{page: page, tier: tier})
	}

	// Sort candidates: tier ascending, then nav_order ascending
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].tier != candidates[j].tier {
			return candidates[i].tier < candidates[j].tier
		}
		return candidates[i].page.NavOrder < candidates[j].page.NavOrder
	})

	for _, c := range candidates {
		primary = append(primary, c.page)
	}

	logger.Info("classifyPagesForNav: classified",
		zap.Int("primary_candidates", len(primary)),
		zap.Int("legal", len(legal)),
		zap.Int("utility", len(utility)),
	)
	return
}

// navPriorityTier returns 1-3 based on how important a page is for primary nav.
// Lower tier = higher priority. Used to fill primary nav slots with the most
// valuable pages when there are more in_header pages than maxHeaderItems allows.
func navPriorityTier(nameLower, pageType string) int {
	// Tier 1 — core pages every site needs in primary nav
	tier1 := map[string]bool{
		"index": true, "services": true, "tools": true, "about": true, "contact": true,
	}
	if tier1[nameLower] {
		return 1
	}

	// Tier 2 — content hubs and key conversion pages
	tier2 := map[string]bool{
		"blog": true, "news": true, "case-studies": true, "use-cases": true,
		"pricing": true, "how-we-work": true, "portfolio": true,
		"products": true, "solutions": true, "industries": true,
		// The directory registers (model_directory_pipeline). These are hub
		// pages in exactly the sense this tier means — a listing of many
		// entries, continuously updated, that a visitor navigates TO rather
		// than lands on. Ranked explicitly because the alternative is not
		// neutral: as tier 3 they sort behind every tier-2 page and are the
		// first thing dropped when max_header_items truncates, so a site that
		// deliberately promoted its directory into the header would silently
		// lose it again the next time any other page gained in_header. That
		// is a real sequence — ai-agent-orchestration.com's header was exactly
		// full on 2026-07-25 and the directory only fitted because the owner
		// moved Pricing down to make room.
		"model-directory": true, "adoption-tracker": true, "protocol-tracker": true,
	}
	// Also tier 2: the section-index family (blog-index, entity-directory,
	// section-index) — listing/hub pages.
	if tier2[nameLower] || isSectionIndexType(pageType) {
		return 2
	}

	// Tier 3 — secondary pages
	return 3
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
		INSERT INTO site_nav_items (site_id, group_id, label, url, page_id, item_type, position, status)
		VALUES ($1, $2, $3, $4, $5, 'page_link', $6, 'active')
	`, siteID, groupID, label, page.URL, page.ID, position)
	return err
}

func navLabelForPage(page pageNavInfo) string {
	// Prefer explicit nav_label from the page — the planner set this
	// intentionally short for nav display.
	if page.NavLabel != "" {
		// Trust nav_label if it's a reasonable nav length.
		// Only simplify if the planner set something unreasonably long.
		if len(page.NavLabel) <= 30 {
			return page.NavLabel
		}
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
