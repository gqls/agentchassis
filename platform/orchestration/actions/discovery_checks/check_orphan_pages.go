// FILE: platform/orchestration/actions/discovery_checks/check_orphan_pages.go
//
// Detects deployed pages that have no inbound links from:
//   - site_nav_items (navigation)
//   - site_components rendered_html (header/footer links)
//   - page_components rendered_html (inline links from other pages)
//
// Orphan pages are live but unreachable — users can't find them.
//
// Three routing paths:
//   1. Blog posts → orphan_blog_posts → rerender-pages (rebuild listing)
//   2. Pages with in_header/in_footer flags → nav_drift → nav-updater
//      (one item per site — nav tables need rebuilding, not per-page fixes)
//   3. Pages without nav flags → needs_internal_links → (future handler)
//      (sub-pages that need linking from parent content, not nav)
//
// Registration: automatic via init() → Register(&OrphanPagesCheck{})

package discovery_checks

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&OrphanPagesCheck{}) }

type OrphanPagesCheck struct{}

func (c *OrphanPagesCheck) Name() string { return "orphan_pages" }

func (c *OrphanPagesCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	orphans, err := findOrphanPages(dctx)
	if err != nil {
		return nil, err
	}
	if len(orphans) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{}

	// Group by routing path
	blogOrphans := 0
	var navDriftPages []orphanPageFinding // have nav flags but aren't in nav
	var unlinkPages []orphanPageFinding   // no nav flags, need internal links

	for _, o := range orphans {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "orphan_pages",
			"page_name": o.Name,
			"page_url":  o.URL,
			"page_type": o.PageType,
			"page_id":   o.ID,
			"in_header": o.InHeader,
			"in_footer": o.InFooter,
		})

		switch o.PageType {
		case "blog-post":
			blogOrphans++
		default:
			if o.InHeader || o.InFooter {
				navDriftPages = append(navDriftPages, o)
			} else {
				unlinkPages = append(unlinkPages, o)
			}
		}
	}

	// ── Blog orphans: rebuild listing (one item per site) ──
	if blogOrphans > 0 {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":        "orphan_pages",
			"orphan_type":  "blog-post",
			"orphan_count": blogOrphans,
			"fix":          "Blog posts are deployed but not linked from the blog listing page. Rebuild the listing.",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "orphan_blog_posts",
			Severity:     "high",
			Summary:      fmt.Sprintf("%d blog posts deployed but not linked from blog listing page", blogOrphans),
			SpecJSON:     string(specJSON),
			Priority:     15,
			HandlerAgent: "rerender-pages",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("orphan_blog_posts:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
	}

	// ── Nav drift: pages have nav flags but aren't in nav (one item per site) ──
	// The fix is to re-run populate_nav_tables + re-render site components.
	// nav-updater already does this.
	if len(navDriftPages) > 0 {
		pageNames := make([]string, len(navDriftPages))
		for i, p := range navDriftPages {
			pageNames[i] = p.Name
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":          "orphan_pages",
			"orphan_type":    "nav_drift",
			"affected_pages": pageNames,
			"affected_count": len(navDriftPages),
			"fix":            "Pages have in_header/in_footer flags set but are missing from site_nav_items. Nav tables need rebuilding.",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "nav_drift",
			Severity:     "high",
			Summary:      fmt.Sprintf("%d pages have nav flags but missing from navigation — nav tables need rebuild", len(navDriftPages)),
			SpecJSON:     string(specJSON),
			Priority:     20,
			HandlerAgent: "nav-updater",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("nav_drift:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
	}

	// ── Needs internal links: sub-pages with no nav flags ──
	// These are service sub-pages, product details, etc. that should be
	// linked from parent content pages, not from top-level navigation.
	for _, o := range unlinkPages {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "orphan_pages",
			"page_name": o.Name,
			"page_url":  o.URL,
			"page_type": o.PageType,
			"page_id":   o.ID,
			"fix":       "Page has no nav flags and no inbound links. Needs linking from a related content page.",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "needs_internal_links",
			Severity:     "low",
			Summary:      fmt.Sprintf("Page %s (%s) has no inbound links — needs linking from related content", o.Name, o.URL),
			SpecJSON:     string(specJSON),
			Priority:     70,
			HandlerAgent: "internal-linker",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("needs_links:%s:%s", o.Name, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
	}

	dctx.Logger.Info("OrphanPagesCheck: complete",
		zap.Int("total_orphans", len(orphans)),
		zap.Int("blog_orphans", blogOrphans),
		zap.Int("nav_drift", len(navDriftPages)),
		zap.Int("needs_links", len(unlinkPages)),
	)

	return result, nil
}

type orphanPageFinding struct {
	ID       string
	Name     string
	URL      string
	PageType string
	InHeader bool
	InFooter bool
}

func findOrphanPages(dctx DiscoveryCheckContext) ([]orphanPageFinding, error) {
	// Find deployed pages that have no inbound links from:
	// 1. site_nav_items (direct nav reference by URL or page_id)
	// 2. site_components rendered_html (header/footer containing the URL)
	// 3. page_components rendered_html on OTHER pages (inline links)
	//
	// Exclusions:
	// - index/home page (always reachable as /)
	// - blog-index page (in nav, the listing is the entry point)
	// - tool pages WITHOUT nav flags (may be linked externally or from JS).
	//   A tool page WITH in_header/in_footer set has declared it belongs in
	//   nav — its absence from site_nav_items is nav_drift like any other
	//   page's, so nav-flagged pages are always considered.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name, p.url, COALESCE(p.page_type, 'content'),
		       COALESCE(p.in_header, false), COALESCE(p.in_footer, false)
		FROM pages p
		WHERE p.site_id = $1
		  AND p.build_status = 'deployed'
		  AND p.url IS NOT NULL
		  AND p.url != ''
		  AND p.name NOT IN ('index', 'home')
		  AND (COALESCE(p.page_type, 'content') NOT IN ('blog-index', 'tool')
		       OR COALESCE(p.in_header, false) OR COALESCE(p.in_footer, false))
		  -- Not linked from nav (by URL)
		  AND NOT EXISTS (
		      SELECT 1 FROM site_nav_items sni
		      WHERE sni.site_id = p.site_id
		        AND sni.url = p.url
		        AND sni.status = 'active'
		  )
		  -- Not linked from nav (by page_id)
		  AND NOT EXISTS (
		      SELECT 1 FROM site_nav_items sni
		      WHERE sni.page_id = p.id
		        AND sni.status = 'active'
		  )
		  -- Not linked from site_components (header/footer)
		  AND NOT EXISTS (
		      SELECT 1 FROM site_components sc
		      WHERE sc.site_id = p.site_id
		        AND sc.rendered_html IS NOT NULL
		        AND sc.rendered_html LIKE '%' || p.url || '%'
		  )
		  -- Not linked from page_components on OTHER pages
		  AND NOT EXISTS (
		      SELECT 1 FROM page_components pc
		      JOIN pages p2 ON pc.page_id = p2.id
		      WHERE p2.site_id = p.site_id
		        AND p2.id != p.id
		        AND pc.rendered_html IS NOT NULL
		        AND pc.rendered_html LIKE '%' || p.url || '%'
		  )
		ORDER BY p.page_type, p.name
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("orphan_pages query failed: %w", err)
	}
	defer rows.Close()

	var findings []orphanPageFinding
	for rows.Next() {
		var f orphanPageFinding
		if err := rows.Scan(&f.ID, &f.Name, &f.URL, &f.PageType, &f.InHeader, &f.InFooter); err != nil {
			dctx.Logger.Warn("Failed to scan orphan page", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
