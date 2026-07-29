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
	"sort"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// structuralPageTypes lists page types that are site infrastructure rather
// than content. These have their own weekly budget so they don't compete
// with content pages for growth slots.
var structuralPageTypes = map[string]bool{
	"news-index":       true,
	"blog-index":       true,
	"sitemap":          true,
	"privacy":          true,
	"terms":            true,
	"error-404":        true,
	"faq":              true,
	"search":           true,
	"tag-index":        true,
	"category":         true,
	"model-directory":  true,
	"adoption-tracker": true,
	"protocol-tracker": true,
}

// isStructuralPageType returns true if the page type is infrastructure
// rather than content.
func isStructuralPageType(pageType string) bool {
	return structuralPageTypes[pageType]
}

// structuralPageTypeList returns the same vocabulary as a sorted slice, for
// passing to SQL as a text[] parameter.
//
// This exists to KILL a drift risk this file used to carry in a comment:
// the budget query below repeated the type list twice as SQL literals, so
// adding a structural page type in the map above and forgetting the SQL
// meant the Go-side classification and the SQL-side count disagreed
// silently — a page treated as structural by one and as content by the
// other. Adding two Phase E types to three hand-maintained copies is exactly
// the moment that stops being hypothetical, so the copies are gone: the map
// is now the single source and the query reads from it.
func structuralPageTypeList() []string {
	out := make([]string, 0, len(structuralPageTypes))
	for t := range structuralPageTypes {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

type GrowthConfig struct {
	InitialTarget            int `json:"initial_target"`
	WeeklyContentPagesMax    int `json:"weekly_content_pages_max"`
	WeeklyBlogPostsMax       int `json:"weekly_blog_posts_max"`
	WeeklyStructuralPagesMax int `json:"weekly_structural_pages_max"`
	AbsoluteMax              int `json:"absolute_max"`

	// ContentToolsRatio: how many PUBLISHED articles/guides justify one tool
	// (6 => a site with 18 guides should have about 3). 0 or absent means OFF,
	// which is the default and the state of 31 of 32 sites.
	//
	// Nothing in THIS file reads it. It is modelled here because this struct is
	// the canonical picture of the growth_config aspect, and the field's only
	// live reader —- discovery_checks/check_missing_tools.go -- cannot use this
	// struct: package actions imports discovery_checks in 6 files and
	// discovery_checks imports actions in 0, so calling loadGrowthConfig from
	// there would close an import cycle (and loadGrowthConfig is unexported
	// besides). It therefore reads the one key with its own inline SQL.
	//
	// Added 2026-07-29 on the council gate's advice: three seats (reuse_agent,
	// architecture, prior_art_librarian) independently objected that a field
	// living only at its call site invites a THIRD ad-hoc reader, because a
	// future author reading this struct would not know the key exists. Keeping
	// it modelled here costs nothing at runtime — loadGrowthConfig unmarshals
	// over defaults — and makes the surface discoverable.
	ContentToolsRatio int `json:"content_tools_ratio"`
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
	//
	// The structural vocabulary comes from structuralPageTypes (one source,
	// see structuralPageTypeList) rather than being repeated here as SQL
	// literals. 'blog-post' is deliberately NOT in that map — it is counted
	// against the blog budget, and the content bucket is "neither blog nor
	// structural", so it appears here on its own.
	var recentContent, recentBlog, recentStructural int
	db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE page_type = 'blog-post'),
			COUNT(*) FILTER (WHERE page_type = ANY($2::text[])),
			COUNT(*) FILTER (WHERE COALESCE(page_type, 'content') <> 'blog-post'
			                   AND NOT (COALESCE(page_type, 'content') = ANY($2::text[])))
		FROM pages
		WHERE site_id = $1
		  AND status IN ('active', 'deployed', 'planned')
		  AND created_at > NOW() - INTERVAL '7 days'
	`, siteID, datahelpers.PGTextArrayLiteral(structuralPageTypeList())).Scan(&recentBlog, &recentStructural, &recentContent)

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
