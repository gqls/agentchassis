// FILE: platform/orchestration/actions/rebuild_blog_listing_action.go
//
// RebuildBlogListingAction queries deployed blog-post pages for a site and
// renders a blog-listing page_component using the template from content_components.
//
// Data layer only — presentation comes from the component library template.
// Uses the existing article_grid or blog-listing content_component's html_template.
// Falls back to a minimal template if no component is found.
//
// No LLM needed — purely algorithmic. Runs as a step in the rerender-pages
// workflow (before get_pages), or triggered after blog post publishing.
//
// Registration:
//   "rebuild_blog_listing": {
//       Handler:     RebuildBlogListingAction,
//       Category:    "site",
//       Description: "Rebuild blog listing page_component from published posts",
//       IsLocal:     true,
//   },
//
// Data inputs (via ActionInputSpec):
//   - site_id (required)

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RebuildBlogListingInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("rebuild_blog_listing", RebuildBlogListingInputSpec)
}

func RebuildBlogListingAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "rebuild_blog_listing"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		RebuildBlogListingInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// ── Find blog-index page ────────────────────────────────────────────
	var blogPageID uuid.UUID
	var blogPageName string
	err = params.DB.QueryRowContext(ctx, `
		SELECT id, name FROM pages
		WHERE site_id = $1 AND page_type = 'blog-index'
		LIMIT 1
	`, siteID).Scan(&blogPageID, &blogPageName)

	if err == sql.ErrNoRows {
		logger.Info("RebuildBlogListingAction: No blog-index page found, skipping")
		return map[string]interface{}{
			"rebuilt": false,
			"reason":  "no blog-index page",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find blog-index page: %w", err)
	}

	// ── Load deployed blog posts ────────────────────────────────────────
	rows, err := params.DB.QueryContext(ctx, `
		SELECT p.id, p.name, p.url, p.title,
		       COALESCE(p.meta_description, ''),
		       p.created_at
		FROM pages p
		WHERE p.site_id = $1
		  AND p.page_type = 'blog-post'
		  AND p.build_status = 'deployed'
		ORDER BY p.created_at DESC
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query blog posts: %w", err)
	}
	defer rows.Close()

	var articles []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, url, title, metaDesc string
		var createdAt time.Time
		if err := rows.Scan(&id, &name, &url, &title, &metaDesc, &createdAt); err != nil {
			logger.Warn("Failed to scan blog post", zap.Error(err))
			continue
		}

		// Strip " | Company Name" suffix from titles
		cleanTitle := title
		if idx := strings.LastIndex(cleanTitle, " | "); idx > 0 {
			cleanTitle = cleanTitle[:idx]
		}

		// Truncate excerpt
		excerpt := metaDesc
		if len(excerpt) > 200 {
			excerpt = excerpt[:197] + "..."
		}

		articles = append(articles, map[string]interface{}{
			"title":     cleanTitle,
			"url":       url,
			"excerpt":   excerpt,
			"date":      createdAt.Format("Jan 2, 2006"),
			"category":  "", // Enrichable from page metadata later
			"image":     "", // Enrichable from assets later
			"read_time": "", // Computable from content length later
		})
	}

	if len(articles) == 0 {
		logger.Info("RebuildBlogListingAction: No deployed blog posts found")
		return map[string]interface{}{
			"rebuilt": false,
			"reason":  "no deployed blog posts",
		}, nil
	}

	// ── Load component template ─────────────────────────────────────────
	htmlTemplate := loadBlogListingTemplate(ctx, params.DB, logger)

	// ── Render template with post data ──────────────────────────────────
	templateData := map[string]interface{}{
		"section_title":    "Latest Articles",
		"section_subtitle": "",
		"articles":         articles,
		"show_load_more":   false,
		"load_more_text":   "Load More",
	}

	rendered := renderBlogTemplate(htmlTemplate, templateData, logger)
	if rendered == "" {
		return nil, fmt.Errorf("template rendering produced empty output")
	}

	// ── Upsert page_component ───────────────────────────────────────────
	var componentID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		SELECT id FROM page_components
		WHERE page_id = $1 AND slot_name = 'blog-listing'
		LIMIT 1
	`, blogPageID).Scan(&componentID)

	if err == sql.ErrNoRows {
		err = params.DB.QueryRowContext(ctx, `
			INSERT INTO page_components (page_id, slot_name, position, rendered_html, build_status)
			VALUES ($1, 'blog-listing', 3, $2, 'deployed')
			RETURNING id
		`, blogPageID, rendered).Scan(&componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert blog-listing component: %w", err)
		}
		logger.Info("RebuildBlogListingAction: Created blog-listing component",
			zap.String("component_id", componentID.String()))
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing blog-listing: %w", err)
	} else {
		_, err = params.DB.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $1, updated_at = NOW()
			WHERE id = $2
		`, rendered, componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to update blog-listing component: %w", err)
		}
		logger.Info("RebuildBlogListingAction: Updated blog-listing component",
			zap.String("component_id", componentID.String()))
	}

	logger.Info("RebuildBlogListingAction: Complete",
		zap.String("blog_page", blogPageName),
		zap.Int("post_count", len(articles)),
	)

	return map[string]interface{}{
		"rebuilt":      true,
		"blog_page_id": blogPageID.String(),
		"post_count":   len(articles),
		"component_id": componentID.String(),
	}, nil
}

// loadBlogListingTemplate finds the best template for the blog listing.
// Priority: blog-listing component → article_grid component → fallback
func loadBlogListingTemplate(ctx context.Context, db *sql.DB, logger *zap.Logger) string {
	// Try specific blog-listing component first
	var tmpl string
	err := db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components
		WHERE (name = 'blog-listing' OR function = 'blog-listing')
		  AND is_active = true
		LIMIT 1
	`).Scan(&tmpl)
	if err == nil && tmpl != "" {
		logger.Info("Using blog-listing component template")
		return tmpl
	}

	// Fall back to article_grid (existing component with content-listing function)
	err = db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components
		WHERE (name = 'article_grid' OR function = 'content-listing')
		  AND is_active = true
		LIMIT 1
	`).Scan(&tmpl)
	if err == nil && tmpl != "" {
		logger.Info("Using article_grid component template as fallback")
		return tmpl
	}

	// Last resort — minimal template using generic CSS class names
	logger.Warn("No blog listing template found in content_components, using default")
	return defaultBlogListingTemplate
}

// renderBlogTemplate renders a Go template with blog data.
func renderBlogTemplate(templateStr string, data map[string]interface{}, logger *zap.Logger) string {
	tmpl, err := template.New("blog-listing").Parse(templateStr)
	if err != nil {
		logger.Warn("Blog listing template parse error", zap.Error(err))
		return ""
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		logger.Warn("Blog listing template execute error", zap.Error(err))
		return ""
	}

	return buf.String()
}

// defaultBlogListingTemplate is used only when no content_component template
// exists. Uses generic CSS class names — the site's stylesheet provides styling.
const defaultBlogListingTemplate = `<section class="section section--articles" data-component="blog-listing">
  <div class="container">
    <div class="section__header">
      <h2 class="section__title">{{.section_title}}</h2>
      {{if .section_subtitle}}<p class="section__subtitle">{{.section_subtitle}}</p>{{end}}
    </div>
    <div class="article-grid grid grid--3">
      {{range .articles}}
      <article class="article-card hover-lift">
        <div class="article-card__content">
          <div class="article-card__meta">
            <span class="article-card__date">{{.date}}</span>
            {{if .read_time}}<span class="article-card__read-time">{{.read_time}}</span>{{end}}
          </div>
          <h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>
          {{if .excerpt}}<p class="article-card__excerpt">{{.excerpt}}</p>{{end}}
        </div>
      </article>
      {{end}}
    </div>
  </div>
</section>`
