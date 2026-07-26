// FILE: platform/orchestration/actions/rebuild_blog_listing_action.go
//
// RebuildBlogListingAction queries deployed blog-post pages for a site and
// renders a blog-listing page_component using the template from content_components.
//
// Data layer only — presentation comes from the component library template.
// Uses the existing content-listing content_component's html_template.
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
//
// Changes from previous version:
//   - findBlogListingSlot: discovers actual slot name from page_components
//     instead of hardcoding "blog-listing". Handles sites where the blog page
//     was planned with article-grid, featured-article, content-listing, etc.
//   - Loads content-listing template (has proper input_schema with {{range}})
//     instead of blog-listing (which was CSS-only, empty input_schema).
//   - ensureArticleLinks: patches article titles missing <a href>.
//   - Writes content_data alongside rendered_html (source-of-truth principle).
//   - estimateReadTime: computed from blog post rendered content length.
//   - Removed loadBlogListingTemplate (was loading CSS-only template).
//
// v2 changes (April 2026):
//   - findBlogPage: falls back to name='blog' when page_type != 'blog-index'.
//     Fixes sites where the planner created the blog page as type 'content'
//     because blog-index was in the unavailableBuilders map.
//     When found by name fallback, corrects page_type to 'blog-index' so
//     future runs find it directly.
//   - Blog post query widened from build_status='deployed' to
//     build_status IN ('deployed','needs_rebuild'). Posts with needs_rebuild
//     are already in git with valid content — they just have pending rerender
//     items. Excluding them left the listing empty.
//
// v3 changes (2026-07-26, bugs_open/052):
//   - The v2 widening above was right about the symptom and wrong about the
//     test. `needs_rebuild` does not mean "still serves" — 4 fleet pages are
//     `needs_rebuild` AND never deployed, and they 404. build_status alone
//     cannot answer "will this URL serve"; `deployed_at` can. The query now
//     carries the shared floor instead of a hand-written build_status list —
//     see blogPostsQuery.
//   - Both findBlogPage strategies gained a `status` filter. They had none, so
//     an archived blog-index page could be selected as the listing target.

package actions

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// blogPostsQuery selects the posts this listing advertises.
//
// The floor is queryresolve.ListedPageEligibilitySQL — the SAME constant
// queryresolve's own `blog_posts` source uses — because this action and that
// resolver both derive the article set for the same blog page and must not
// disagree about it. Two hand-maintained copies of one predicate is the
// drift class bugs_closed/023 documents; sharing the constant is what stops
// it recurring here.
//
// Why the floor at all (bugs_open/052): a listing regenerated from the page
// set must never advertise a page that would 404. The previous predicate,
// `build_status IN ('deployed','needs_rebuild')`, admitted two populations it
// should not have — pages that are `needs_rebuild` but were never deployed
// (they 404), and, because there was no `status` filter at all, pages that had
// been deliberately `archived`. That second gap defeated archiving, which is
// the containment route the fleet relies on for a dead page.
//
// Alias contract: `p` for pages, as the shared constants require.
const blogPostsQuery = `
		SELECT p.id, p.name, p.url, p.title,
		       COALESCE(p.meta_description, ''),
		       p.created_at,
		       COALESCE(
		           (SELECT SUM(LENGTH(COALESCE(pc.rendered_html, '')))
		            FROM page_components pc
		            WHERE pc.page_id = p.id
		              AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')),
		           0
		       ) as content_length
		FROM pages p
		WHERE p.site_id = $1
		  AND p.page_type = 'blog-post'
		  AND p.status IN ('active', 'deployed')` + queryresolve.ListedPageEligibilitySQL + `
		ORDER BY p.created_at DESC
	`

var RebuildBlogListingInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("rebuild_blog_listing", RebuildBlogListingInputSpec)
}

// slotPriority is the ordered list of slot names we look for when finding
// the blog listing page_component. The first match wins.
// These cover the various names planners have used for blog listing sections.
var slotPriority = []string{
	"blog-listing",
	"article-grid",
	"content-listing",
	"guide-list",
	"featured-article",
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

	// ── Find blog page ──────────────────────────────────────────────────
	blogPageID, blogPageName, err := findBlogPage(ctx, params.DB, siteID, logger)
	if err != nil {
		return nil, err
	}
	if blogPageID == uuid.Nil {
		logger.Info("RebuildBlogListingAction: No blog page found, skipping")
		return map[string]interface{}{
			"rebuilt": false,
			"reason":  "no blog page",
		}, nil
	}

	// ── Find the correct slot name ──────────────────────────────────────
	slotName, existingComponentID := findBlogListingSlot(ctx, params.DB, blogPageID, logger)
	logger.Info("RebuildBlogListingAction: Resolved slot",
		zap.String("slot_name", slotName),
		zap.Bool("has_existing_component", existingComponentID != uuid.Nil),
	)

	// ── Load blog posts ─────────────────────────────────────────────────
	rows, err := params.DB.QueryContext(ctx, blogPostsQuery, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query blog posts: %w", err)
	}
	defer rows.Close()

	var articles []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, url, title, metaDesc string
		var createdAt time.Time
		var contentLength int
		if err := rows.Scan(&id, &name, &url, &title, &metaDesc, &createdAt, &contentLength); err != nil {
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
			"category":  "",
			"image":     "",
			"read_time": estimateReadTime(contentLength),
		})
	}

	if len(articles) == 0 {
		// An empty set leaves any EXISTING listing untouched, so the page keeps
		// advertising whatever it advertised before. That is the safe choice —
		// blanking a live listing on a transient empty read would be worse —
		// but it is not a no-op, so say so loudly when there is a component
		// standing: the eligibility floor above can newly empty a set that used
		// to fill, and the stale listing is then the thing serving the 404s
		// this fix exists to stop (bugs_open/052).
		if existingComponentID != uuid.Nil {
			logger.Warn("RebuildBlogListingAction: no eligible blog posts, but a listing component exists — it keeps its previous contents and may now be stale",
				zap.String("blog_page", blogPageName),
				zap.String("slot_name", slotName),
				zap.String("component_id", existingComponentID.String()),
			)
		} else {
			logger.Info("RebuildBlogListingAction: No blog posts found")
		}
		return map[string]interface{}{
			"rebuilt": false,
			"reason":  "no blog posts",
		}, nil
	}

	// ── Load component template ─────────────────────────────────────────
	// Use content-listing function (has proper input_schema with {{range}}).
	// The old blog-listing component was CSS-only with empty input_schema.
	htmlTemplate := loadContentListingTemplate(ctx, params.DB, logger)

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

	// Patch: ensure article titles are clickable links
	rendered = ensureArticleLinks(rendered, articles)

	// ── Build content_data (source of truth) ────────────────────────────
	contentData := map[string]interface{}{
		"section_title":    "Latest Articles",
		"section_subtitle": "",
		"articles":         articles,
		"show_load_more":   false,
		"load_more_text":   "Load More",
	}
	contentDataJSON, err := json.Marshal(contentData)
	if err != nil {
		logger.Warn("Failed to marshal content_data", zap.Error(err))
		contentDataJSON = []byte("{}")
	}

	// ── Upsert page_component ───────────────────────────────────────────
	var componentID uuid.UUID

	if existingComponentID != uuid.Nil {
		// Update existing component in the correct slot — unless a human has
		// locked it (bugs_open/058): the lock predicate on the WHERE makes the
		// refusal race-free, and the blocked refresh is surfaced as a work
		// item rather than silently skipped.
		res, updErr := params.DB.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $1, content_data = $2::jsonb, updated_at = NOW()
			WHERE id = $3 AND `+pageComponentAgentWritableSQL("")+`
		`, rendered, string(contentDataJSON), existingComponentID)
		if updErr != nil {
			return nil, fmt.Errorf("failed to update blog listing component: %w", updErr)
		}
		if n, raErr := res.RowsAffected(); raErr == nil && n == 0 {
			logger.Warn("RebuildBlogListingAction: blog listing component is human-locked — refresh refused (bugs_open/058)",
				zap.String("component_id", existingComponentID.String()),
				zap.String("slot_name", slotName),
			)
			lock, lockErr := CheckComponentLock(ctx, params.DB, existingComponentID, logger)
			lockedBy, lockType := "", ""
			if lockErr == nil && lock.IsLocked {
				lockedBy, lockType = lock.LockedBy, lock.LockType
			}
			pcID := existingComponentID
			emitLockBlockedChangeItem(ctx, params.DB, siteID, &blogPageID, &pcID,
				blogPageName, slotName, lockedBy, lockType,
				"overwrite", "rebuild_blog_listing", logger)
			return map[string]interface{}{
				"rebuilt": false,
				"skipped": true,
				"locked":  true,
				"reason":  fmt.Sprintf("blog listing component %s is locked by %q — auto-refresh refused", existingComponentID, lockedBy),
			}, nil
		}
		componentID = existingComponentID
		logger.Info("RebuildBlogListingAction: Updated existing component",
			zap.String("component_id", componentID.String()),
			zap.String("slot_name", slotName),
		)
	} else {
		// No existing component found — insert into the discovered slot
		err = params.DB.QueryRowContext(ctx, `
			INSERT INTO page_components (page_id, slot_name, position, rendered_html, content_data, build_status)
			VALUES ($1, $2, 3, $3, $4::jsonb, 'deployed')
			RETURNING id
		`, blogPageID, slotName, rendered, string(contentDataJSON)).Scan(&componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert blog listing component: %w", err)
		}
		logger.Info("RebuildBlogListingAction: Created new component",
			zap.String("component_id", componentID.String()),
			zap.String("slot_name", slotName),
		)
	}

	logger.Info("RebuildBlogListingAction: Complete",
		zap.String("blog_page", blogPageName),
		zap.Int("post_count", len(articles)),
		zap.String("slot_name", slotName),
	)

	return map[string]interface{}{
		"rebuilt":      true,
		"blog_page_id": blogPageID.String(),
		"post_count":   len(articles),
		"component_id": componentID.String(),
		"slot_name":    slotName,
	}, nil
}

// findBlogPage locates the blog listing page for a site.
// Strategy 1: Look for page_type = 'blog-index' (canonical).
// Strategy 2: Fall back to name = 'blog' with page_type = 'content'
//
//	(created when blog-index was in unavailableBuilders).
//	When found by fallback, corrects page_type to 'blog-index'
//	so future runs and other actions (e.g. BlogEmptyCheck) behave consistently.
//
// Returns uuid.Nil, "", nil if no blog page exists (not an error).
func findBlogPage(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) (uuid.UUID, string, error) {
	var blogPageID uuid.UUID
	var blogPageName string

	// Strategy 1: canonical page_type.
	// The status filter is load-bearing (bugs_open/052): without it an archived
	// blog-index page is a valid listing target, so the action would rebuild a
	// listing on a page the site has deliberately retired.
	err := db.QueryRowContext(ctx, `
		SELECT id, name FROM pages
		WHERE site_id = $1 AND page_type = 'blog-index'
		  AND status IN ('active', 'deployed')
		LIMIT 1
	`, siteID).Scan(&blogPageID, &blogPageName)

	if err == nil {
		return blogPageID, blogPageName, nil
	}
	if err != sql.ErrNoRows {
		return uuid.Nil, "", fmt.Errorf("failed to query blog-index page: %w", err)
	}

	// Strategy 2: page named 'blog' created as content type.
	// Same status filter, and here it also guards a WRITE: on a hit this
	// strategy stamps page_type='blog-index' onto the row it found.
	err = db.QueryRowContext(ctx, `
		SELECT id, name FROM pages
		WHERE site_id = $1 AND name = 'blog' AND page_type = 'content'
		  AND status IN ('active', 'deployed')
		LIMIT 1
	`, siteID).Scan(&blogPageID, &blogPageName)

	if err == sql.ErrNoRows {
		return uuid.Nil, "", nil
	}
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to query blog page by name: %w", err)
	}

	// Found by fallback — fix page_type for future consistency
	_, fixErr := db.ExecContext(ctx, `
		UPDATE pages SET page_type = 'blog-index' WHERE id = $1
	`, blogPageID)
	if fixErr != nil {
		logger.Warn("findBlogPage: failed to fix page_type to blog-index",
			zap.String("page_id", blogPageID.String()),
			zap.Error(fixErr))
	} else {
		logger.Info("findBlogPage: corrected page_type from content to blog-index",
			zap.String("page_id", blogPageID.String()),
			zap.String("page_name", blogPageName))
	}

	return blogPageID, blogPageName, nil
}

// findBlogListingSlot checks existing page_components for the blog-index page
// and returns the first slot name that matches our priority list.
// If no match is found, falls back to checking the page's sections array,
// then defaults to "blog-listing".
// Also returns the existing component ID if found (uuid.Nil if not).
func findBlogListingSlot(ctx context.Context, db *sql.DB, blogPageID uuid.UUID, logger *zap.Logger) (string, uuid.UUID) {
	// Strategy 1: Check existing page_components for a known listing slot
	for _, candidate := range slotPriority {
		var pcID uuid.UUID
		err := db.QueryRowContext(ctx, `
			SELECT id FROM page_components
			WHERE page_id = $1 AND slot_name = $2
			LIMIT 1
		`, blogPageID, candidate).Scan(&pcID)
		if err == nil {
			logger.Info("findBlogListingSlot: Found existing component by slot_name",
				zap.String("slot_name", candidate),
				zap.String("component_id", pcID.String()),
			)
			return candidate, pcID
		}
	}

	// Strategy 2: Check the page's sections array (the planner's plan)
	var sectionsJSON []byte
	err := db.QueryRowContext(ctx, `
		SELECT sections FROM pages WHERE id = $1 AND sections IS NOT NULL
	`, blogPageID).Scan(&sectionsJSON)
	if err == nil && len(sectionsJSON) > 0 {
		var sections []string
		if json.Unmarshal(sectionsJSON, &sections) == nil {
			for _, candidate := range slotPriority {
				for _, section := range sections {
					if section == candidate {
						logger.Info("findBlogListingSlot: Found slot in page sections array",
							zap.String("slot_name", candidate),
						)
						return candidate, uuid.Nil
					}
				}
			}
			// If the sections array has entries not in our priority list,
			// use the first non-standard section (skip hero, header, footer, head, call-to-action)
			skipSections := map[string]bool{
				"hero": true, "header": true, "footer": true,
				"head": true, "call-to-action": true, "cta": true,
			}
			for _, section := range sections {
				if !skipSections[section] {
					logger.Info("findBlogListingSlot: Using first content section from page plan",
						zap.String("slot_name", section),
					)
					return section, uuid.Nil
				}
			}
		}
	}

	// Strategy 3: Default
	logger.Info("findBlogListingSlot: No existing slot found, defaulting to blog-listing")
	return "blog-listing", uuid.Nil
}

// loadContentListingTemplate loads the content-listing template from
// content_components. This component has a proper input_schema with
// {{range .articles}} and article card markup.
// Falls back to a minimal default if not found.
func loadContentListingTemplate(ctx context.Context, db *sql.DB, logger *zap.Logger) string {
	var tmpl string

	// Primary: content-listing function (has proper input_schema)
	err := db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components
		WHERE function = 'content-listing'
		  AND is_active = true
		  AND html_template IS NOT NULL
		  AND html_template != ''
		  AND html_template LIKE '%range%'
		ORDER BY created_at DESC
		LIMIT 1
	`).Scan(&tmpl)
	if err == nil && tmpl != "" {
		logger.Info("RebuildBlogListingAction: Using content-listing template")
		return tmpl
	}

	// Fallback: article_grid by name
	err = db.QueryRowContext(ctx, `
		SELECT html_template FROM content_components
		WHERE name = 'article_grid'
		  AND is_active = true
		  AND html_template LIKE '%range%'
		LIMIT 1
	`).Scan(&tmpl)
	if err == nil && tmpl != "" {
		logger.Info("RebuildBlogListingAction: Using article_grid template as fallback")
		return tmpl
	}

	// Last resort — minimal template
	logger.Warn("RebuildBlogListingAction: No listing template found, using built-in default")
	return defaultBlogListingTemplate
}

// ensureArticleLinks checks rendered HTML for article titles that lack
// <a href> links and patches them using the article data.
// This handles the case where the content-listing template has
// <h3 class="article-card__title">{{.title}}</h3> without a link.
func ensureArticleLinks(html string, articles []map[string]interface{}) string {
	for _, article := range articles {
		title, _ := article["title"].(string)
		url, _ := article["url"].(string)
		if title == "" || url == "" {
			continue
		}

		// Look for title text NOT already wrapped in an <a> tag
		// Pattern: <h3...>Title</h3> where Title is not inside <a>
		escapedTitle := regexp.QuoteMeta(title)
		// Match: >Title</h3  (title directly inside heading, no <a>)
		pattern := regexp.MustCompile(`(>[^<]*)` + escapedTitle + `([^<]*</h[1-6])`)
		if pattern.MatchString(html) {
			// Check it's not already linked
			linkedPattern := regexp.MustCompile(`<a[^>]*>` + escapedTitle + `</a>`)
			if !linkedPattern.MatchString(html) {
				// Replace bare title with linked title
				old := `>` + title + `</h`
				new := `><a href="` + url + `">` + title + `</a></h`
				html = strings.Replace(html, old, new, 1)
			}
		}
	}
	return html
}

// estimateReadTime returns a human-readable read time estimate
// based on approximate content length in characters.
// Assumes ~5 chars per word, 200 words per minute.
func estimateReadTime(contentLengthChars int) string {
	if contentLengthChars == 0 {
		return ""
	}
	words := contentLengthChars / 5
	minutes := words / 200
	if minutes < 1 {
		minutes = 1
	}
	return fmt.Sprintf("%d min read", minutes)
}

// renderBlogTemplate renders a Go template with blog data.
func renderBlogTemplate(htmlTemplate string, data map[string]interface{}, logger *zap.Logger) string {
	tmpl, err := template.New("blog_listing").Parse(htmlTemplate)
	if err != nil {
		logger.Warn("RebuildBlogListingAction: Failed to parse template, using fallback",
			zap.Error(err))
		tmpl, _ = template.New("blog_listing").Parse(defaultBlogListingTemplate)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		logger.Warn("RebuildBlogListingAction: Failed to execute template",
			zap.Error(err))
		return ""
	}

	return buf.String()
}

var defaultBlogListingTemplate = `<section class="blog-listing" data-component="content-listing">
  <div class="container">
    {{if .section_title}}<h2 class="section-title">{{.section_title}}</h2>{{end}}
    {{if .section_subtitle}}<p class="section-subtitle">{{.section_subtitle}}</p>{{end}}
    <div class="article-grid">
      {{range .articles}}
      <article class="article-card">
        <div class="article-card__content">
          <h3 class="article-card__title"><a href="{{.url}}">{{.title}}</a></h3>
          {{if .date}}<time class="article-card__date">{{.date}}</time>{{end}}
          {{if .read_time}}<span class="article-card__read-time">{{.read_time}}</span>{{end}}
          {{if .excerpt}}<p class="article-card__excerpt">{{.excerpt}}</p>{{end}}
        </div>
      </article>
      {{end}}
    </div>
  </div>
</section>`
