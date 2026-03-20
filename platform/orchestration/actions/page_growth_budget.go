// FILE: platform/orchestration/actions/page_growth_budget.go
//
// Checks whether a site is within its page growth budget before allowing
// new page creation. Called by apply_gap_plan (new_page) and blog-content-planner.
//
// Budget is configured via a site_specs aspect "growth_config":
//   {
//     "initial_target": 12,
//     "weekly_content_pages_max": 3,
//     "weekly_blog_posts_max": 2,
//     "absolute_max": 60
//   }
//
// If no growth_config spec exists, defaults are used.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type GrowthConfig struct {
	InitialTarget         int `json:"initial_target"`
	WeeklyContentPagesMax int `json:"weekly_content_pages_max"`
	WeeklyBlogPostsMax    int `json:"weekly_blog_posts_max"`
	AbsoluteMax           int `json:"absolute_max"`
}

var DefaultGrowthConfig = GrowthConfig{
	InitialTarget:         12,
	WeeklyContentPagesMax: 3,
	WeeklyBlogPostsMax:    2,
	AbsoluteMax:           60,
}

type GrowthBudgetResult struct {
	Allowed       bool         `json:"allowed"`
	Reason        string       `json:"reason"`
	CurrentTotal  int          `json:"current_total"`
	RecentContent int          `json:"recent_content_pages"`
	RecentBlog    int          `json:"recent_blog_posts"`
	Config        GrowthConfig `json:"config"`
}

// CheckPageGrowthBudget determines whether a new page can be created for a site.
// pageType should be "blog-post" for blog posts, anything else for content pages.
func CheckPageGrowthBudget(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageType string, logger *zap.Logger) (*GrowthBudgetResult, error) {
	config := loadGrowthConfig(ctx, db, siteID)

	// Count total active pages
	var totalPages int
	db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM pages
		WHERE site_id = $1 AND status IN ('active', 'deployed', 'planned')
	`, siteID).Scan(&totalPages)

	// Count pages created in the last 7 days, split by type
	var recentContent, recentBlog int
	db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE COALESCE(page_type, 'content') != 'blog-post'),
			COUNT(*) FILTER (WHERE page_type = 'blog-post')
		FROM pages
		WHERE site_id = $1
		  AND status IN ('active', 'deployed', 'planned')
		  AND created_at > NOW() - INTERVAL '7 days'
	`, siteID).Scan(&recentContent, &recentBlog)

	result := &GrowthBudgetResult{
		CurrentTotal:  totalPages,
		RecentContent: recentContent,
		RecentBlog:    recentBlog,
		Config:        config,
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

	// Past initial target — check weekly rate limits
	isBlogPost := pageType == "blog-post"

	if isBlogPost {
		if recentBlog >= config.WeeklyBlogPostsMax {
			result.Reason = "weekly_blog_limit_reached"
			logger.Info("PageGrowthBudget: weekly blog limit reached",
				zap.String("site_id", siteID.String()),
				zap.Int("recent_blog", recentBlog),
				zap.Int("max", config.WeeklyBlogPostsMax))
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
	if config.AbsoluteMax < config.InitialTarget {
		config.AbsoluteMax = config.InitialTarget * 5
	}

	return config
}
