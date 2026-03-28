// FILE: platform/orchestration/actions/discovery_checks/check_orphan_pages.go
//
// Detects deployed pages that have no inbound links from:
//   - site_nav_items (navigation)
//   - site_components rendered_html (header/footer links)
//   - page_components rendered_html (inline links from other pages)
//
// Orphan pages are live but unreachable — users can't find them.
// The most common cause is blog posts published without updating the
// blog listing page, or new pages added without nav entries.
//
// Creates work items routed to rebuild_blog_listing (for blog posts)
// or nav-agent (for content pages missing from nav).
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

	// Group by page_type for targeted work items
	blogOrphans := 0
	contentOrphans := 0

	for _, o := range orphans {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "orphan_pages",
			"page_name": o.Name,
			"page_url":  o.URL,
			"page_type": o.PageType,
			"page_id":   o.ID,
		})

		switch o.PageType {
		case "blog-post":
			blogOrphans++
		default:
			contentOrphans++

			specJSON, _ := json.Marshal(map[string]interface{}{
				"check":     "orphan_pages",
				"page_name": o.Name,
				"page_url":  o.URL,
				"page_type": o.PageType,
				"page_id":   o.ID,
				"fix":       "Page is deployed but not linked from navigation or any other page",
			})

			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:       dctx.SiteID,
				Source:       "discovery",
				Pipeline:     "build",
				ItemType:     "orphan_page",
				Severity:     "medium",
				Summary:      fmt.Sprintf("Page %s (%s) is deployed but unreachable — no inbound links", o.Name, o.URL),
				SpecJSON:     string(specJSON),
				Priority:     45,
				HandlerAgent: "content-gap-planner",
				Status:       "detected",
				CreatedBy:    dctx.AgentType,
				ItemKey:      fmt.Sprintf("orphan_page:%s:%s", o.Name, dctx.SiteID),
				BatchID:      dctx.BatchID,
			})
		}
	}

	// Blog orphans get a single work item — the fix is to rebuild the listing,
	// not to fix each post individually
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

	dctx.Logger.Info("OrphanPagesCheck: complete",
		zap.Int("total_orphans", len(orphans)),
		zap.Int("blog_orphans", blogOrphans),
		zap.Int("content_orphans", contentOrphans),
	)

	return result, nil
}

type orphanPageFinding struct {
	ID       string
	Name     string
	URL      string
	PageType string
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
	// - tool pages (may be linked externally or from JS)
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name, p.url, COALESCE(p.page_type, 'content')
		FROM pages p
		WHERE p.site_id = $1
		  AND p.build_status = 'deployed'
		  AND p.url IS NOT NULL
		  AND p.url != ''
		  AND p.name NOT IN ('index', 'home')
		  AND COALESCE(p.page_type, 'content') NOT IN ('blog-index', 'tool')
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
		if err := rows.Scan(&f.ID, &f.Name, &f.URL, &f.PageType); err != nil {
			dctx.Logger.Warn("Failed to scan orphan page", zap.Error(err))
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
