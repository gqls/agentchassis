// FILE: platform/orchestration/actions/page_growth_budget.go
//
// Checks whether a site is within its page growth budget before allowing
// new page creation. Called by apply_gap_plan (new_page) and blog-content-planner.
//
// Three-tier budget:
//   - Content pages (about, services, guides, etc.)
//   - Blog posts
//   - Structural pages (news-index, blog-index, privacy, terms, sitemap, etc.)
//
// Budget is configured per-site via site_specs aspect "growth_config":
//   {
//     "initial_target": 12,
//     "weekly_content_pages_max": 3,
//     "weekly_blog_posts_max": 2,
//     "weekly_structural_pages_max": 3,
//     "absolute_max": 60
//   }
//
// If no growth_config spec exists, defaults are used. Missing fields in a
// partial spec keep their default values (JSON unmarshal over defaults).

package actions

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// structuralPageTypes lists page types that are site infrastructure rather
// than content. These have their own weekly budget so they don't compete
// with content pages for growth slots.
var structuralPageTypes = map[string]bool{
	"news-index": true,
	"blog-index": true,
	"sitemap":    true,
	"privacy":    true,
	"terms":      true,
	"error-404":  true,
	"faq":        true,
	"search":     true,
	"tag-index":  true,
	"category":   true,
}

// isStructuralPageType returns true if the page type is infrastructure
// rather than content.
func isStructuralPageType(pageType string) bool {
	return structuralPageTypes[pageType]
}

type GrowthConfig struct {
	InitialTarget            int `json:"initial_target"`
	WeeklyContentPagesMax    int `json:"weekly_content_pages_max"`
	WeeklyBlogPostsMax       int `json:"weekly_blog_posts_max"`
	WeeklyStructuralPagesMax int `json:"weekly_structural_pages_max"`
	AbsoluteMax              int `json:"absolute_max"`
}

var DefaultGrowthConfig = GrowthConfig{
	InitialTarget:            12,
	WeeklyContentPagesMax:    3,
	WeeklyBlogPostsMax:       2,
	WeeklyStructuralPagesMax: 3,
	AbsoluteMax:              60,
}

type GrowthBudgetResult struct {
	Allowed          bool         `json:"allowed"`
	Reason           string       `json:"reason"`
	CurrentTotal     int          `json:"current_total"`
	RecentContent    int          `json:"recent_content_pages"`
	RecentBlog       int          `json:"recent_blog_posts"`
	RecentStructural int          `json:"recent_structural_pages"`
	Config           GrowthConfig `json:"config"`
}

// CheckPageGrowthBudget determines whether a new page can be created for a site.
// pageType is checked against three categories:
//   - "blog-post" → blog budget
//   - structural types (news-index, blog-index, etc.) → structural budget
//   - everything else → content budget
func CheckPageGrowthBudget(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageType string, logger *zap.Logger) (*GrowthBudgetResult, error) {
	config := loadGrowthConfig(ctx, db, siteID)

	// Count total active pages
	var totalPages int
	db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pages
		WHERE site_id = $1 AND status IN ('active', 'deployed', 'planned')
	`, siteID).Scan(&totalPages)

	// Count pages created in the last 7 days, split by type
	// Blog: page_type = 'blog-post'
	// Structural: page_type IN (news-index, blog-index, privacy, terms, sitemap, etc.)
	// Content: everything else
	var recentContent, recentBlog, recentStructural int
	db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE page_type = 'blog-post'),
			COUNT(*) FILTER (WHERE page_type IN (
				'news-index', 'blog-index', 'sitemap', 'privacy', 'terms',
				'error-404', 'faq', 'search', 'tag-index', 'category'
			)),
			COUNT(*) FILTER (WHERE COALESCE(page_type, 'content') NOT IN (
				'blog-post', 'news-index', 'blog-index', 'sitemap', 'privacy', 'terms',
				'error-404', 'faq', 'search', 'tag-index', 'category'
			))
		FROM pages
		WHERE site_id = $1
		  AND status IN ('active', 'deployed', 'planned')
		  AND created_at > NOW() - INTERVAL '7 days'
	`, siteID).Scan(&recentBlog, &recentStructural, &recentContent)

	result := &GrowthBudgetResult{
		CurrentTotal:     totalPages,
		RecentContent:    recentContent,
		RecentBlog:       recentBlog,
		RecentStructural: recentStructural,
		Config:           config,
	}

	// Check absolute max
	if totalPages >= config.AbsoluteMax {
		result.Reason = "absolute_max_reached"
		logger.Info("PageGrowthBudget: absolute max reached",
			zap.String("site_id", siteID.String()),
			zap.Int("total", totalPages),
			zap.Int("max", config.AbsoluteMax))
		return result, nil
	}

	// If still under initial target, allow freely
	if totalPages < config.InitialTarget {
		result.Allowed = true
		result.Reason = "under_initial_target"
		return result, nil
	}

	// Past initial target — check weekly rate limits by category
	isBlogPost := pageType == "blog-post"
	isStructural := isStructuralPageType(pageType)

	if isBlogPost {
		if recentBlog >= config.WeeklyBlogPostsMax {
			result.Reason = "weekly_blog_limit_reached"
			logger.Info("PageGrowthBudget: weekly blog limit reached",
				zap.String("site_id", siteID.String()),
				zap.Int("recent_blog", recentBlog),
				zap.Int("max", config.WeeklyBlogPostsMax))
			return result, nil
		}
	} else if isStructural {
		if recentStructural >= config.WeeklyStructuralPagesMax {
			result.Reason = "weekly_structural_limit_reached"
			logger.Info("PageGrowthBudget: weekly structural limit reached",
				zap.String("site_id", siteID.String()),
				zap.String("page_type", pageType),
				zap.Int("recent_structural", recentStructural),
				zap.Int("max", config.WeeklyStructuralPagesMax))
			return result, nil
		}
	} else {
		if recentContent >= config.WeeklyContentPagesMax {
			result.Reason = "weekly_content_limit_reached"
			logger.Info("PageGrowthBudget: weekly content limit reached",
				zap.String("site_id", siteID.String()),
				zap.Int("recent_content", recentContent),
				zap.Int("max", config.WeeklyContentPagesMax))
			return result, nil
		}
	}

	result.Allowed = true
	result.Reason = "within_budget"
	return result, nil
}

func loadGrowthConfig(ctx context.Context, db *sql.DB, siteID uuid.UUID) GrowthConfig {
	config := DefaultGrowthConfig

	var dataJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT data FROM site_specs
		WHERE site_id = $1 AND aspect = 'growth_config' AND is_current = true
	`, siteID).Scan(&dataJSON)

	if err != nil || len(dataJSON) == 0 {
		return config
	}

	// Unmarshal over defaults — missing fields keep their default values
	json.Unmarshal(dataJSON, &config)

	// Sanity: ensure minimums
	if config.InitialTarget < 3 {
		config.InitialTarget = 3
	}
	if config.WeeklyContentPagesMax < 1 {
		config.WeeklyContentPagesMax = 1
	}
	if config.WeeklyBlogPostsMax < 1 {
		config.WeeklyBlogPostsMax = 1
	}
	if config.WeeklyStructuralPagesMax < 1 {
		config.WeeklyStructuralPagesMax = 1
	}
	if config.AbsoluteMax < 10 {
		config.AbsoluteMax = 10
	}

	return config
}
