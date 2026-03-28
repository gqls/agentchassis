// FILE: platform/orchestration/actions/rebuild_blog_listing_action.go
//
// RebuildBlogListingAction queries deployed blog-post pages for a site and
// writes a blog-listing page_component to the blog-index page. This replaces
// the query.* template-based approach with a simple rebuild-on-change pattern.
//
// No LLM needed — purely algorithmic. Runs before rerender, or after blog
// post publishing.
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
	"context"
	"database/sql"
	"fmt"
	"html"
	"strings"
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

type blogPost struct {
	ID        uuid.UUID
	Name      string
	URL       string
	Title     string
	PageType  string
	CreatedAt time.Time
	MetaDesc  string
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
		SELECT id, name, url, title, COALESCE(meta_description, ''), created_at
		FROM pages
		WHERE site_id = $1
		  AND page_type = 'blog-post'
		  AND build_status = 'deployed'
		ORDER BY created_at DESC
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query blog posts: %w", err)
	}
	defer rows.Close()

	var posts []blogPost
	for rows.Next() {
		var p blogPost
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Title, &p.MetaDesc, &p.CreatedAt); err != nil {
			logger.Warn("Failed to scan blog post", zap.Error(err))
			continue
		}
		posts = append(posts, p)
	}

	if len(posts) == 0 {
		logger.Info("RebuildBlogListingAction: No deployed blog posts found")
		return map[string]interface{}{
			"rebuilt": false,
			"reason":  "no deployed blog posts",
		}, nil
	}

	// ── Build listing HTML ──────────────────────────────────────────────
	listingHTML := buildBlogListingHTML(posts)

	// ── Upsert page_component ───────────────────────────────────────────
	var componentID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		SELECT id FROM page_components
		WHERE page_id = $1 AND slot_name = 'blog-listing'
		LIMIT 1
	`, blogPageID).Scan(&componentID)

	if err == sql.ErrNoRows {
		// Insert new
		err = params.DB.QueryRowContext(ctx, `
			INSERT INTO page_components (page_id, slot_name, position, rendered_html, build_status)
			VALUES ($1, 'blog-listing', 3, $2, 'deployed')
			RETURNING id
		`, blogPageID, listingHTML).Scan(&componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert blog-listing component: %w", err)
		}
		logger.Info("RebuildBlogListingAction: Created blog-listing component",
			zap.String("component_id", componentID.String()))
	} else if err != nil {
		return nil, fmt.Errorf("failed to check existing blog-listing: %w", err)
	} else {
		// Update existing
		_, err = params.DB.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $1, updated_at = NOW()
			WHERE id = $2
		`, listingHTML, componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to update blog-listing component: %w", err)
		}
		logger.Info("RebuildBlogListingAction: Updated blog-listing component",
			zap.String("component_id", componentID.String()))
	}

	logger.Info("RebuildBlogListingAction: Complete",
		zap.String("blog_page", blogPageName),
		zap.Int("post_count", len(posts)),
	)

	return map[string]interface{}{
		"rebuilt":      true,
		"blog_page_id": blogPageID.String(),
		"post_count":   len(posts),
		"component_id": componentID.String(),
	}, nil
}

// buildBlogListingHTML produces the article grid HTML from blog post records.
// Uses CSS variables for consistent theming with the rest of the site.
func buildBlogListingHTML(posts []blogPost) string {
	var b strings.Builder

	b.WriteString(`<section class="blog-listing" data-component="blog-listing">
    <div class="blog-container">
        <div class="blog-grid">
`)

	for _, p := range posts {
		// Clean the title — strip " | Company Name" suffix that some titles have
		title := p.Title
		if idx := strings.LastIndex(title, " | "); idx > 0 {
			title = title[:idx]
		}

		// Use meta_description as excerpt, fall back to empty
		excerpt := p.MetaDesc
		if len(excerpt) > 200 {
			excerpt = excerpt[:197] + "..."
		}

		date := p.CreatedAt.Format("Jan 2, 2006")

		b.WriteString(fmt.Sprintf(`
            <a href="%s" class="blog-card">
                <div class="blog-card__meta">
                    <time>%s</time>
                </div>
                <h3>%s</h3>
                <p>%s</p>
            </a>
`,
			html.EscapeString(p.URL),
			html.EscapeString(date),
			html.EscapeString(title),
			html.EscapeString(excerpt),
		))
	}

	b.WriteString(`
        </div>
    </div>
</section>
<style>
.blog-listing {
    padding: 4rem 2rem;
    background: var(--background-color, #0a0a1a);
}
.blog-container {
    max-width: 1200px;
    margin: 0 auto;
}
.blog-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: 2rem;
}
.blog-card {
    display: block;
    background: rgba(255,255,255,0.04);
    border: 1px solid rgba(255,255,255,0.08);
    border-radius: 8px;
    padding: 2rem;
    text-decoration: none;
    color: rgba(255,255,255,0.9);
    transition: all 0.2s ease;
}
.blog-card:hover {
    background: rgba(255,255,255,0.08);
    border-color: rgba(255,255,255,0.15);
    transform: translateY(-2px);
}
.blog-card__meta {
    font-size: 0.8rem;
    color: rgba(255,255,255,0.5);
    margin-bottom: 0.75rem;
}
.blog-card h3 {
    font-size: 1.25rem;
    font-weight: 600;
    line-height: 1.3;
    margin-bottom: 0.75rem;
    color: #fff;
}
.blog-card p {
    font-size: 0.95rem;
    line-height: 1.6;
    color: rgba(255,255,255,0.6);
    margin: 0;
}
@media (max-width: 768px) {
    .blog-listing { padding: 2rem 1rem; }
    .blog-grid { grid-template-columns: 1fr; gap: 1.5rem; }
}
</style>`)

	return b.String()
}
