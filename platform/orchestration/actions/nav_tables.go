// FILE: platform/orchestration/actions/nav_tables.go
//
// Unified navigation queries from site_nav_groups + site_nav_items tables.
// Falls back to pages table for sites that predate the nav tables.
//
// REPLACES these functions (all deprecated, kept temporarily for reference):
//   - loadNavItems              (render_site_components_action.go)
//   - loadFooterNavItems        (render_site_components_action.go)
//   - GetHeaderNavFromPages     (inject_header_footer.go)
//   - GetFooterNavFromPages     (inject_header_footer.go)
//   - rerenderGetHeaderNavFromDB (rerender_site_pages_action.go)
//   - rerenderGetFooterNavFromDB (rerender_site_pages_action.go) [dead code]
//   - getNavigationFromDB       (sync_pages_to_db_action.go)
//   - buildNavigationFromDB     (sync_pages_to_db_action.go)
//   - buildNavigationFromPages  (sync_pages_to_db_action.go)
//   - cacheNavigation           (sync_pages_to_db_action.go)

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

// Nav group type constants for consumer queries.
const (
	NavGroupPrimary = "primary"
	NavGroupLegal   = "legal"
	NavGroupUtility = "utility"
	NavGroupContent = "content"
)

// NavVisibility says what a caller intends to do with the items it gets back.
//
// It replaced a `deployedOnly bool` on 2026-07-26 (bugs_open/049 mechanism 2).
// Deliberately a named type and not a bool: the defect was a `false` at three
// chrome call sites that read as harmless, and a bool RENAME does not break
// compilation, so nobody would have re-read them. A new type makes the compiler
// stop at every call site, which is the structural half of the fix — the
// predicate is only the arithmetic half.
//
// Both former settings of that bool were wrong. `false` shipped links to pages
// that had never been built (a 404 on every page of the site); `true` filtered
// on `build_status = 'deployed'`, which is measured wrong for the 34 fleet pages
// that are needs_rebuild but still serve their old artefact. See
// datahelpers.NeverDeployedPagePredicate for that measurement.
type NavVisibility int

const (
	// NavAllItems returns every active item regardless of whether its target
	// exists yet. For planning and prompt context, where a page about to be
	// built SHOULD appear — telling a writer the page does not exist is worse
	// than naming one that is coming.
	NavAllItems NavVisibility = iota

	// NavFetchableOnly drops items whose target would 404. For anything that
	// becomes an <a href> in shipped HTML.
	NavFetchableOnly
)

// ---------------------------------------------------------------------------
// Public API — two entry points, same underlying logic
// ---------------------------------------------------------------------------

// GetNavItems returns []NavItem for the requested group types.
// Tries nav tables first; falls back to pages table for backward compatibility.
//
// groupTypes: which groups to include
//   - header callers pass: []string{"primary"}
//   - footer callers pass: []string{"primary", "utility", "legal"}
//
// visibility: NavFetchableOnly for anything that ships as an <a href>;
//
//	NavAllItems for planning/prompt context. See NavVisibility.
//
// maxItems:   0 means no limit. Applied AFTER the visibility filter, so it means
//
//	"N usable items" — a SQL LIMIT would cap first and then lose more
//	to filtering, silently returning a short nav.
func GetNavItems(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	groupTypes []string,
	visibility NavVisibility,
	maxItems int,
	logger *zap.Logger,
) []NavItem {
	if db == nil || siteID == uuid.Nil {
		logger.Debug("GetNavItems: No DB or site_id, returning empty")
		return []NavItem{}
	}

	items := getNavItemsFromTables(ctx, db, siteID, groupTypes, logger)
	if len(items) > 0 {
		return applyNavVisibility(ctx, db, siteID, items, visibility, groupTypes, maxItems, logger)
	}

	// An empty result from the nav tables is ambiguous: it can mean "this site
	// predates the nav tables" (the backward-compat case the fallback exists for)
	// or "this site HAS nav tables and genuinely has no items in the requested
	// groups" — the correct, expected answer for most sites (e.g. no legal pages).
	// Only fall back when the site has no nav-table rows at all; otherwise respect
	// the truthful empty answer. Gating on the per-group row count instead let a
	// site with a full nav but no legal pages take the fallback, which then filled
	// the footer's legal slot with every footer page. See /bugs_open/053.
	if siteHasAnyNavItems(ctx, db, siteID, logger) {
		// The site uses nav tables and simply has no items in the requested
		// groups. Respect that truthful empty answer rather than papering over
		// it with the pages fallback.
		//
		// But distinguish a LEGITIMATELY-empty group from an ANOMALOUS one. A
		// lone legal/utility/content group is empty for most sites by design. A
		// request that includes the PRIMARY group, however, should never come
		// back empty on a site that has nav rows — primary is the main menu, so
		// an empty result means the primary nav is missing (a bad sync/write),
		// not a truthful "no such pages". Returning empty silently there would
		// be a fresh instance of the platform's silent-empty-render class
		// (016b §9): required-but-missing rendered as empty with no error. So
		// fail LOUD for that case and quiet for the expected one, at no cost
		// until an anomaly actually occurs (today every live site has primary
		// nav rows, so this Warn never fires).
		if containsGroupType(groupTypes, NavGroupPrimary) {
			logger.Warn("GetNavItems: nav-table site returned NO items for a request including the primary group — primary nav appears to be missing (expected non-empty); serving empty nav",
				zap.String("site_id", siteID.String()),
				zap.Strings("group_types", groupTypes),
			)
		} else {
			logger.Debug("GetNavItems: nav-table site has no items for these groups; returning truthful empty",
				zap.String("site_id", siteID.String()),
				zap.Strings("group_types", groupTypes),
			)
		}
		return []NavItem{}
	}

	// No nav table entries for this site at all — fall back to pages table
	logger.Debug("GetNavItems: No nav table entries, using pages fallback",
		zap.String("site_id", siteID.String()),
		zap.Strings("group_types", groupTypes),
	)
	fallback := getNavItemsFromPagesFallback(ctx, db, siteID, groupTypes, logger)
	return applyNavVisibility(ctx, db, siteID, fallback, visibility, groupTypes, maxItems, logger)
}

// applyNavVisibility drops items whose target would 404, then caps to maxItems.
// It runs on BOTH source paths (nav tables and pages fallback) so there is one
// filter and one safety net rather than two that can drift.
//
// The invariant is: filtering can SHRINK a nav, never EMPTY one. Three cases
// depend on it, and the loud/quiet split mirrors the one just above for empty
// groups — fail loud for the anomalous case, stay quiet for the expected one:
//
//   - First build: nothing is deployed yet, so every item is unfetchable. Serve
//     the unfiltered nav (identical to the old deployedOnly=false behaviour) and
//     Warn. This matters because RenderSiteComponentsAction's idempotence gate
//     skips re-rendering once rendered_html is non-empty, so a nav emptied at
//     first build would persist indefinitely.
//   - Partial: the real defect — some targets exist, some never shipped. Drop
//     the dead ones and name them.
//   - Loader failure: never let an infrastructure error empty a site's chrome.
//
// Judging the URL rather than page_id is deliberate. The URL is what ships;
// page_id is metadata that goes NULL on ON DELETE SET NULL (016_nav_tables.sql),
// which is exactly how a nav item outlives its page — and the old
// `ni.page_id IS NULL OR ...` predicate KEPT those. Matching on URL through
// PageURLSet also means this loader and check_phantom_internal_links agree on
// "same target" by one implementation, not two that resemble each other.
func applyNavVisibility(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	items []NavItem,
	visibility NavVisibility,
	groupTypes []string,
	maxItems int,
	logger *zap.Logger,
) []NavItem {
	if visibility == NavFetchableOnly && len(items) > 0 {
		fetchable, deployedPages, err := loadFetchablePageSet(ctx, db, siteID)
		switch {
		case err != nil:
			logger.Warn("GetNavItems: fetchable-page lookup failed; serving the UNFILTERED nav rather than risk empty chrome",
				zap.String("site_id", siteID.String()),
				zap.Strings("group_types", groupTypes),
				zap.Error(err),
			)
		case deployedPages == 0:
			// The site has not deployed a single page, so "never deployed" is
			// true of everything and carries no signal — this is a first build,
			// not a site full of dead links. Filtering here would freeze a
			// near-empty nav into the chrome (the idempotence gate means it may
			// never be re-rendered). Note this cannot be detected from the
			// surviving-item count: loadFetchablePageSet always injects the site
			// root, so a "Home" item survives even when nothing is deployed, and
			// the rest would be silently dropped.
			logger.Warn("GetNavItems: site has NO deployed pages — serving the UNFILTERED nav (expected during a first build, anomalous on an established site)",
				zap.String("site_id", siteID.String()),
				zap.Strings("group_types", groupTypes),
				zap.Int("items", len(items)),
			)
		default:
			kept := make([]NavItem, 0, len(items))
			var dropped []string
			for _, item := range items {
				// Only page links can 404 against the pages table. External
				// links, mailto and in-page anchors are out of scope.
				if datahelpers.ClassifyLinkScope(item.URL) != datahelpers.LinkScopePage ||
					fetchable.Contains(item.URL) {
					kept = append(kept, item)
					continue
				}
				dropped = append(dropped, item.URL)
			}

			if len(kept) == 0 {
				// Belt and braces alongside the deployedPages==0 branch: the
				// site HAS deployed pages, yet not one nav item points at them.
				// That is a broken nav rather than a young one, and empty chrome
				// helps nobody.
				logger.Warn("GetNavItems: EVERY nav item's target is unfetchable on a site that HAS deployed pages — serving the UNFILTERED nav rather than empty chrome",
					zap.String("site_id", siteID.String()),
					zap.Strings("group_types", groupTypes),
					zap.Int("items", len(items)),
				)
			} else {
				if len(dropped) > 0 {
					logger.Warn("GetNavItems: dropped nav items whose target page has never been deployed (would 404 on every page)",
						zap.String("site_id", siteID.String()),
						zap.Strings("group_types", groupTypes),
						zap.Strings("dropped_urls", dropped),
					)
				}
				items = kept
			}
		}
	}

	if maxItems > 0 && len(items) > maxItems {
		items = items[:maxItems]
	}
	return items
}

// loadFetchablePageSet returns the site's page URLs that would actually serve —
// every non-deleted page except those that have never been deployed — together
// with the count of REAL rows behind it.
//
// The count is returned separately because the set is not a proxy for it: the
// site root is always injected (matching check_phantom_internal_links), so the
// set is never empty and cannot tell a first build from an established site.
func loadFetchablePageSet(ctx context.Context, db *sql.DB, siteID uuid.UUID) (datahelpers.PageURLSet, int, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT url
		FROM pages
		WHERE site_id = $1
		  AND status NOT IN ('deleted', 'archived')
		  AND NOT (`+datahelpers.NeverDeployedPagePredicate+`)
	`, siteID)
	if err != nil {
		return nil, 0, fmt.Errorf("fetchable pages query failed: %w", err)
	}
	defer rows.Close()

	urls := []string{"/", "/index.html"}
	deployedPages := 0
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		urls = append(urls, url)
		deployedPages++
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("fetchable pages iteration failed: %w", err)
	}
	return datahelpers.NewPageURLSet(urls), deployedPages, nil
}

// GetNavigationStructure returns *NavigationStructure for callers that
// need that type (SyncPagesToDBAction, BuildRenderContextAction,
// GetNavigationFromDBAction). All callers pass ActionParams.DB which is *sql.DB.
func GetNavigationStructure(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	navType string,
	logger *zap.Logger,
) (*NavigationStructure, error) {
	groupTypes := navTypeToGroupTypes(navType)

	// NavAllItems, deliberately. The only live consumer of this structure is
	// extractSitemapInfo, which builds LLM PROMPT context — at planning time the
	// intended sitemap should include pages that are about to be built, because
	// telling a writer those pages do not exist is worse than naming one that is
	// coming. Nothing here becomes an <a href> directly.
	items := GetNavItems(ctx, db, siteID, groupTypes, NavAllItems, 0, logger)

	nav := &NavigationStructure{
		Items: make([]NavigationItem, len(items)),
	}
	for i, item := range items {
		nav.Items[i] = NavigationItem{
			Label: item.Label,
			URL:   item.URL,
		}
	}
	return nav, nil
}

// ---------------------------------------------------------------------------
// Nav tables query
// ---------------------------------------------------------------------------

func getNavItemsFromTables(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	groupTypes []string,
	logger *zap.Logger,
) []NavItem {
	if len(groupTypes) == 0 {
		return nil
	}

	// Build positional params: $1=site_id, $2..$N+1=group types
	args := make([]interface{}, 0, len(groupTypes)+1)
	args = append(args, siteID)

	placeholders := make([]string, len(groupTypes))
	for i, gt := range groupTypes {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, gt)
	}
	inClause := strings.Join(placeholders, ", ")

	var qb strings.Builder
	qb.WriteString(`
		SELECT ni.label, ni.url
		FROM site_nav_items ni
		JOIN site_nav_groups ng ON ni.group_id = ng.id`)

	fmt.Fprintf(&qb, `
		WHERE ni.site_id = $1
		  AND ng.group_type IN (%s)
		  AND ni.status = 'active'`, inClause)

	// No deployment predicate here any more, and no LIMIT. Both moved to
	// applyNavVisibility: the old `ni.page_id IS NULL OR p.build_status =
	// 'deployed'` was wrong twice over (it kept items orphaned to a NULL
	// page_id, and it dropped pages that are needs_rebuild but still serving),
	// and a SQL LIMIT capped before filtering, so a nav could come back short.
	qb.WriteString(`
		ORDER BY ng.position ASC, ni.position ASC`)

	rows, err := db.QueryContext(ctx, qb.String(), args...)
	if err != nil {
		// Table might not exist on older deployments
		if strings.Contains(err.Error(), "does not exist") {
			logger.Debug("getNavItemsFromTables: Nav tables not created yet")
			return nil
		}
		logger.Warn("getNavItemsFromTables: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			logger.Warn("getNavItemsFromTables: Scan failed", zap.Error(err))
			continue
		}
		items = append(items, NavItem{Label: label, URL: url})
	}

	if len(items) > 0 {
		logger.Debug("getNavItemsFromTables: Loaded from nav tables",
			zap.Int("items", len(items)),
			zap.Strings("group_types", groupTypes),
		)
	}

	return items
}

// ---------------------------------------------------------------------------
// Pages-table fallback (backward compatibility for pre-nav-table sites)
// ---------------------------------------------------------------------------

func getNavItemsFromPagesFallback(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	groupTypes []string,
	logger *zap.Logger,
) []NavItem {
	isHeaderOnly := len(groupTypes) == 1 && groupTypes[0] == NavGroupPrimary
	includesLegal := containsGroupType(groupTypes, NavGroupLegal)

	var qb strings.Builder
	args := []interface{}{siteID}

	qb.WriteString(`
		SELECT
			COALESCE(nav_label, title, name) as label,
			COALESCE(url, '/' || name || '.html') as url
		FROM pages
		WHERE site_id = $1`)

	if isHeaderOnly {
		// Header: in_header=true, exclude legal/system pages
		qb.WriteString(`
		  AND in_header = true
		  AND name NOT IN ('privacy', 'terms', 'cookies', '404', 'sitemap')`)
	} else {
		// Footer: in_footer=true, optionally include legal pages by name
		qb.WriteString(`
		  AND (in_footer = true`)
		if includesLegal {
			qb.WriteString(` OR LOWER(name) IN ('privacy', 'terms', 'cookies', 'disclaimer')`)
		}
		qb.WriteString(`)
		  AND name NOT IN ('index', '404', 'sitemap')`)
	}

	qb.WriteString(`
		  AND status IN ('deployed', 'active')`)

	// That is pages.status (the row's lifecycle), a different column from
	// build_status. The build_status filter that used to sit here moved to
	// applyNavVisibility — see getNavItemsFromTables for why.

	if isHeaderOnly {
		qb.WriteString(`
		ORDER BY
			COALESCE(nav_order, 99),
			CASE name
				WHEN 'index' THEN 1 WHEN 'home' THEN 1
				WHEN 'services' THEN 2 WHEN 'about' THEN 3
				WHEN 'contact' THEN 10 ELSE 5
			END`)
	} else {
		qb.WriteString(`
		ORDER BY
			CASE WHEN LOWER(name) IN ('privacy', 'terms', 'cookies', 'disclaimer') THEN 1 ELSE 0 END,
			COALESCE(nav_order, 99),
			CASE name
				WHEN 'services' THEN 1 WHEN 'about' THEN 2
				WHEN 'contact' THEN 3 WHEN 'privacy' THEN 8
				WHEN 'terms' THEN 9 ELSE 5
			END`)
	}

	// No LIMIT — applyNavVisibility caps after filtering, so maxItems means
	// "N usable items".

	rows, err := db.QueryContext(ctx, qb.String(), args...)
	if err != nil {
		logger.Warn("getNavItemsFromPagesFallback: Query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var items []NavItem
	for rows.Next() {
		var label, url string
		if err := rows.Scan(&label, &url); err != nil {
			continue
		}
		label = navSimplifyLabel(label, url)
		items = append(items, NavItem{Label: label, URL: url})
	}

	logger.Debug("getNavItemsFromPagesFallback: Built from pages table",
		zap.Int("items", len(items)),
		zap.Bool("header_only", isHeaderOnly),
	)

	return items
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// navTypeToGroupTypes maps the old "header"/"footer" navType to group type arrays.
func navTypeToGroupTypes(navType string) []string {
	switch navType {
	case "header":
		return []string{NavGroupPrimary}
	case "footer":
		return []string{NavGroupPrimary, NavGroupUtility, NavGroupLegal}
	default:
		return []string{NavGroupPrimary}
	}
}

func containsGroupType(types []string, target string) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}

// siteHasAnyNavItems reports whether the site has ANY rows in site_nav_items
// (any group, any status). It is the gate that distinguishes a pre-nav-table
// site — no rows, so the pages-table fallback is the only source of nav — from
// a nav-table site that has simply no items in the requested group, whose empty
// answer is truthful and must be respected. If the nav tables do not exist yet
// (older deployments) the query errors and we treat the site as pre-nav-table.
// See /bugs_open/053.
func siteHasAnyNavItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) bool {
	var exists bool
	err := db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM site_nav_items WHERE site_id = $1)`, siteID,
	).Scan(&exists)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			logger.Debug("siteHasAnyNavItems: Nav tables not created yet, treating as pre-nav-table")
			return false
		}
		// On any other error, prefer the fallback rather than silently
		// suppressing all nav — the previous behaviour on a nav-table miss.
		logger.Warn("siteHasAnyNavItems: existence check failed, using pages fallback", zap.Error(err))
		return false
	}
	return exists
}

// navSimplifyLabel consolidates rerenderSimplifyNavLabel and
// datahelpers.SimplifyNavLabel into one function.
// Only applied in the pages-table fallback path; nav table entries
// are expected to have clean labels set by PopulateNavTablesAction.
func navSimplifyLabel(label, url string) string {
	if len(label) <= 15 {
		return label
	}

	name := strings.TrimPrefix(url, "/")
	name = strings.TrimSuffix(name, ".html")
	nameLower := strings.ToLower(name)

	switch {
	case nameLower == "index" || nameLower == "home":
		return "Home"
	case strings.HasPrefix(nameLower, "about"):
		return "About"
	case strings.HasPrefix(nameLower, "service"):
		return "Services"
	case strings.HasPrefix(nameLower, "contact"):
		return "Contact"
	case strings.HasPrefix(nameLower, "work") || strings.HasPrefix(nameLower, "portfolio") || strings.HasPrefix(nameLower, "case"):
		return "Work"
	case strings.HasPrefix(nameLower, "team"):
		return "Team"
	case strings.HasPrefix(nameLower, "privacy"):
		return "Privacy"
	case strings.HasPrefix(nameLower, "terms"):
		return "Terms"
	case strings.HasPrefix(nameLower, "blog"):
		return "Blog"
	default:
		return strings.Title(strings.ReplaceAll(name, "-", " "))
	}
}
