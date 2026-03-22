// FILE: platform/orchestration/actions/discovery_checks/check_empty_blog.go
//
// Detects blog/blog-index pages that have no blog-post pages.
// Creates needs_blog_posts work item routed to blog-content-planner.
//
// Registration: automatic via init() → Register(&BlogEmptyCheck{})

package discovery_checks

import (
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&BlogEmptyCheck{}) }

type BlogEmptyCheck struct{}

func (c *BlogEmptyCheck) Name() string { return "empty_blog" }

func (c *BlogEmptyCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Check if a blog/blog-index page exists
	var blogPageID string
	var blogPageName string
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT id::text, name FROM pages
		WHERE site_id = $1
		  AND (page_type = 'blog-index' OR name = 'blog')
		  AND build_status IN ('deployed', 'planned', 'active')
		LIMIT 1
	`, dctx.SiteID).Scan(&blogPageID, &blogPageName)

	if err != nil {
		// No blog page — nothing to check
		return &CheckResult{}, nil
	}

	// Count blog-post pages
	var postCount int
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*) FROM pages
		WHERE site_id = $1
		  AND page_type = 'blog-post'
		  AND build_status IN ('deployed', 'planned', 'active')
	`, dctx.SiteID).Scan(&postCount)

	if err != nil {
		dctx.Logger.Warn("BlogEmptyCheck: failed to count posts", zap.Error(err))
		return &CheckResult{}, nil
	}

	if postCount > 0 {
		// Has posts, blog is populated
		return &CheckResult{}, nil
	}

	// Blog page exists but no posts
	pageID, _ := uuid.Parse(blogPageID)

	dctx.Logger.Info("BlogEmptyCheck: blog page found with no posts",
		zap.String("site_id", dctx.SiteID.String()),
		zap.String("blog_page", blogPageName),
		zap.Int("post_count", 0))

	specJSON := fmt.Sprintf(`{"check":"empty_blog","blog_page_id":"%s","blog_page_name":"%s","post_count":0}`,
		blogPageID, blogPageName)

	return &CheckResult{
		Findings: []map[string]interface{}{{
			"check":          "empty_blog",
			"blog_page_id":   blogPageID,
			"blog_page_name": blogPageName,
			"post_count":     0,
		}},
		WorkItems: []WorkItemSpec{{
			SiteID:       dctx.SiteID,
			Source:       "discovery",
			Pipeline:     "content",
			ItemType:     "needs_blog_posts",
			Severity:     "medium",
			Summary:      "Blog page exists but no blog posts — needs initial content planned by blog-content-planner",
			SpecJSON:     specJSON,
			PageID:       &pageID,
			Priority:     50,
			HandlerAgent: "blog-content-planner",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("empty_blog:%s", dctx.SiteID),
			BatchID:      dctx.BatchID,
		}},
	}, nil
}
