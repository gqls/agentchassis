// FILE: platform/orchestration/actions/create_blog_posts_action.go
//
// Takes an LLM-planned blog post list and creates page records +
// needs_content_page work items for each post. Also creates a
// needs_rerender item for the blog index page.
//
// This avoids reusing sync_pages_to_db (which expects planner-specific
// data shapes) and write_build_items (which has similar assumptions).

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var CreateBlogPostsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"plan_field"},
	Defaults:   map[string]interface{}{"plan_field": "site_plan.result"},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_blog_posts", CreateBlogPostsInputSpec)
}

type blogPostPlan struct {
	Name     string   `json:"name"`
	Title    string   `json:"title"`
	Purpose  string   `json:"purpose"`
	PageType string   `json:"page_type"`
	Sections []string `json:"sections"`
}

func CreateBlogPostsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "create_blog_posts"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, CreateBlogPostsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Extract the plan — try multiple paths
	planField := inputs.Get("plan_field")
	if planField == "" {
		planField = "site_plan.result"
	}

	var posts []blogPostPlan
	var planRaw interface{}

	// Try the configured path first
	planRaw = datahelpers.ExtractNestedField(params.CollectedData, planField)

	// Fallback paths
	if planRaw == nil {
		for _, path := range []string{
			"site_plan.result",
			"site_plan",
			"plan_posts.result",
			"blog_plan.result",
			"blog_plan",
		} {
			planRaw = datahelpers.ExtractNestedField(params.CollectedData, path)
			if planRaw != nil {
				logger.Info("Found plan at fallback path", zap.String("path", path))
				break
			}
		}
	}

	if planRaw == nil {
		return nil, fmt.Errorf("no blog plan found at '%s' or fallback paths", planField)
	}

	// Parse the plan — could be a JSON string, a map, or already have a pages/posts array
	switch v := planRaw.(type) {
	case string:
		// Strip markdown fences
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)

		if cleaned == "" {
			return nil, fmt.Errorf("blog plan is empty string")
		}

		// Try parsing as {pages: [...]} or {posts: [...]}
		var wrapper map[string]json.RawMessage
		if err := json.Unmarshal([]byte(cleaned), &wrapper); err == nil {
			if pagesJSON, ok := wrapper["pages"]; ok {
				json.Unmarshal(pagesJSON, &posts)
			} else if postsJSON, ok := wrapper["posts"]; ok {
				json.Unmarshal(postsJSON, &posts)
			}
		}

		// Try parsing as direct array
		if len(posts) == 0 {
			json.Unmarshal([]byte(cleaned), &posts)
		}

	case map[string]interface{}:
		// Look for pages or posts key
		var arr interface{}
		if p, ok := v["pages"]; ok {
			arr = p
		} else if p, ok := v["posts"]; ok {
			arr = p
		}
		if arr != nil {
			b, _ := json.Marshal(arr)
			json.Unmarshal(b, &posts)
		}

	case []interface{}:
		b, _ := json.Marshal(v)
		json.Unmarshal(b, &posts)
	}

	if len(posts) == 0 {
		logger.Warn("No blog posts found in plan",
			zap.String("plan_field", planField),
			zap.String("plan_type", fmt.Sprintf("%T", planRaw)))
		return map[string]interface{}{
			"pages_created": 0,
			"items_created": 0,
			"reason":        "no posts in plan",
		}, nil
	}

	logger.Info("Creating blog posts",
		zap.Int("count", len(posts)),
		zap.String("site_id", siteIDStr))

	// Get the site domain for work item summaries
	var domain string
	params.DB.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)

	pagesCreated := 0
	itemsCreated := 0
	skipped := 0
	batchID := uuid.New()

	for i, post := range posts {
		// Sanitise name
		name := post.Name
		if name == "" {
			name = fmt.Sprintf("blog-post-%d", i+1)
		}
		name = strings.ToLower(strings.ReplaceAll(name, " ", "-"))

		// Default sections
		sections := post.Sections
		if len(sections) == 0 {
			sections = []string{"hero", "article-body", "call-to-action"}
		}
		sectionsJSON, _ := json.Marshal(sections)

		// Default page type
		pageType := post.PageType
		if pageType == "" {
			pageType = "blog-post"
		}
		
		// Check growth budget for blog posts
		budget, budgetErr := CheckPageGrowthBudget(ctx, params.DB, siteID, pageType, logger)
		if budgetErr != nil {
			logger.Warn("Growth budget check failed, allowing by default", zap.Error(budgetErr))
		} else if !budget.Allowed {
			logger.Info("Blog post throttled by growth budget",
				zap.String("title", post.Title),
				zap.String("reason", budget.Reason),
				zap.Int("recent_blog", budget.RecentBlog),
				zap.Int("max", budget.Config.WeeklyBlogPostsMax))
			skipped++
			continue
		}

		// URL
		url := fmt.Sprintf("/blog/%s.html", name)

		// Create page record
		var pageID uuid.UUID
		pageSpec, _ := json.Marshal(map[string]interface{}{
			"purpose": post.Purpose,
		})
		err := params.DB.QueryRowContext(ctx, `
			INSERT INTO pages (site_id, name, url, title, page_type,
			                   build_status, nav_order, in_header, in_footer, sections, page_spec)
			VALUES ($1, $2, $3, $4, $5, 'planned', $6, false, false, $7::jsonb, $8::jsonb)
			ON CONFLICT (site_id, name) DO UPDATE SET
				title = EXCLUDED.title,
				page_type = EXCLUDED.page_type,
				sections = EXCLUDED.sections,
				page_spec = EXCLUDED.page_spec,
				updated_at = NOW()
			RETURNING id
		`, siteID, name, url, post.Title, pageType,
			20+i, string(sectionsJSON), string(pageSpec)).Scan(&pageID)

		if err != nil {
			logger.Warn("Failed to create page", zap.String("name", name), zap.Error(err))
			continue
		}
		pagesCreated++

		// Create work item
		spec := map[string]interface{}{
			"page_name": name,
			"page_type": pageType,
			"title":     post.Title,
			"purpose":   post.Purpose,
			"sections":  sections,
		}
		specJSON, _ := json.Marshal(spec)

		itemKey := fmt.Sprintf("blog_post_%s_%s", name, siteID)

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, domain, item_type, severity, summary,
				page_id, priority, handler_agent, status, created_by,
				spec, item_key, batch_id
			) VALUES ($1, 'blog-content-planner', 'build', 'needs_content_page',
			          'medium', $2, $3, 55, 'page-build-handler', 'triaged',
			          'blog-content-planner', $4::jsonb, $5, $6)
			ON CONFLICT DO NOTHING
		`, siteID, "Write blog post: "+post.Title, pageID,
			string(specJSON), itemKey, batchID)

		if err != nil {
			logger.Warn("Failed to create work item", zap.String("name", name), zap.Error(err))
			continue
		}
		itemsCreated++
	}

	// Create rerender for blog index
	var blogPageID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		SELECT id FROM pages WHERE site_id = $1 AND (name = 'blog' OR page_type = 'blog-index') LIMIT 1
	`, siteID).Scan(&blogPageID)

	if err == nil {
		params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, domain, item_type, severity, summary,
				page_id, priority, handler_agent, status, created_by,
				spec, item_key, batch_id
			) VALUES ($1, 'blog-content-planner', 'build', 'needs_rerender',
			          'medium', 'Re-render blog index after blog posts created',
			          $2, 60, 'rerender-pages', 'triaged',
			          'blog-content-planner',
			          '{"refresh_site_components": false, "reason": "new blog posts"}'::jsonb,
			          $3, $4)
			ON CONFLICT DO NOTHING
		`, siteID, blogPageID, fmt.Sprintf("rerender_blog_%s", siteID), batchID)
	}

	logger.Info("CreateBlogPostsAction: Complete",
		zap.Int("pages_created", pagesCreated),
		zap.Int("items_created", itemsCreated),
		zap.String("batch_id", batchID.String()))

	return map[string]interface{}{
		"pages_created":  pagesCreated,
		"items_created":  itemsCreated,
		"budget_skipped": skipped,
		"batch_id":       batchID.String(),
	}, nil
}
