// FILE: platform/orchestration/actions/render_news_section_action.go
//
// RenderNewsSectionAction queries recent relevant feed items for a site,
// loads the latest-news component template, renders it, and upserts the
// page_component. Follows the same pattern as rebuild_blog_listing.
//
// This is a data-driven render — no LLM call. The template is populated
// from content_feed_items rows. The action is idempotent: re-running it
// with the same data produces the same HTML.
//
// Registration:
//   "render_news_section": {
//       Handler:     RenderNewsSectionAction,
//       Category:    "feed",
//       Description: "Render latest-news component from content_feed_items",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "render_news": {
//       "action": "render_news_section",
//       "config": {
//           "site_id": "input_data.site_id",
//           "page_name": "index",
//           "max_items": 6,
//           "max_age_hours": 72
//       },
//       "output_field": "news_render_result",
//       "next_step": "complete"
//   }

package actions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenderNewsSectionInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"page_name", "max_items", "max_age_hours", "headline"},
}

func init() {
	datahelpers.RegisterActionInputSpec("render_news_section", RenderNewsSectionInputSpec)
}

func RenderNewsSectionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "render_news_section"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		RenderNewsSectionInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	pageName := inputs.Get("page_name")
	if pageName == "" {
		pageName = "index"
	}

	maxItems := inputs.GetInt("max_items", 6)
	maxAgeHours := inputs.GetInt("max_age_hours", 72)

	headline := "Latest News"
	if h, ok := params.StepConfig.Config["headline"].(string); ok && h != "" {
		headline = h
	}

	// -----------------------------------------------------------------------
	// 1. Find the target page
	// -----------------------------------------------------------------------
	var pageID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		SELECT id FROM pages
		WHERE site_id = $1 AND name = $2 AND status = 'active'
	`, siteID, pageName).Scan(&pageID)
	if err == sql.ErrNoRows {
		logger.Info("RenderNewsSectionAction: target page not found, skipping",
			zap.String("page_name", pageName))
		return map[string]interface{}{
			"rendered": false,
			"reason":   "page not found",
			"page":     pageName,
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query page: %w", err)
	}

	// -----------------------------------------------------------------------
	// 2. Load relevant feed items
	// -----------------------------------------------------------------------
	rows, err := params.DB.QueryContext(ctx, `
		SELECT 
			cfi.source_title,
			cfi.source_summary,
			cfi.source_url,
			cfi.source_published_at,
			COALESCE(cs.name, '') as source_name
		FROM content_feed_items cfi
		LEFT JOIN content_sources cs ON cs.id = cfi.source_id
		WHERE cfi.site_id = $1 
		  AND cfi.status IN ('relevant', 'ingested')
		  AND cfi.created_at > NOW() - make_interval(hours => $2)
		ORDER BY cfi.source_published_at DESC NULLS LAST, cfi.created_at DESC
		LIMIT $3
	`, siteID, maxAgeHours, maxItems)
	if err != nil {
		return nil, fmt.Errorf("query feed items: %w", err)
	}
	defer rows.Close()

	var newsItems []map[string]interface{}
	for rows.Next() {
		var title, summary, url, sourceName sql.NullString
		var publishedAt sql.NullTime

		if err := rows.Scan(&title, &summary, &url, &publishedAt, &sourceName); err != nil {
			logger.Warn("RenderNewsSectionAction: scan error", zap.Error(err))
			continue
		}

		item := map[string]interface{}{
			"source_title":   title.String,
			"source_summary": truncateSummary(summary.String, 200),
			"source_url":     url.String,
			"source_name":    sourceName.String,
		}

		if publishedAt.Valid {
			item["published_display"] = formatNewsDate(publishedAt.Time)
		}

		newsItems = append(newsItems, item)
	}

	logger.Info("RenderNewsSectionAction: loaded items",
		zap.Int("count", len(newsItems)),
		zap.String("page", pageName))

	// -----------------------------------------------------------------------
	// 3. Load the latest-news component template
	// -----------------------------------------------------------------------
	var componentID uuid.UUID
	var htmlTemplate string

	err = params.DB.QueryRowContext(ctx, `
		SELECT id, html_template FROM content_components
		WHERE function = 'latest-news' AND is_active = true
		LIMIT 1
	`).Scan(&componentID, &htmlTemplate)
	if err == sql.ErrNoRows {
		logger.Warn("RenderNewsSectionAction: latest-news component not found in content_components")
		return map[string]interface{}{
			"rendered": false,
			"reason":   "latest-news component template not found",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query component: %w", err)
	}

	// -----------------------------------------------------------------------
	// 4. Render the template
	// -----------------------------------------------------------------------
	templateData := map[string]interface{}{
		"headline":   headline,
		"news_items": newsItems,
	}

	renderedHTML := RenderTemplateWithMap(htmlTemplate, templateData, logger)
	if renderedHTML == "" && len(newsItems) > 0 {
		logger.Warn("RenderNewsSectionAction: template rendered empty despite having items")
	}

	// Content hash for change detection
	contentHash := fmt.Sprintf("%x", sha256.Sum256([]byte(renderedHTML)))[:16]

	// Content data for the page_component
	contentDataJSON, _ := json.Marshal(templateData)

	// -----------------------------------------------------------------------
	// 5. Upsert page_component
	// -----------------------------------------------------------------------
	// Find existing latest-news component on this page
	var existingID uuid.UUID
	err = params.DB.QueryRowContext(ctx, `
		SELECT pc.id FROM page_components pc
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE pc.page_id = $1 AND cc.function = 'latest-news'
		LIMIT 1
	`, pageID).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Insert new — find the highest position and add after it
		var maxPosition int
		params.DB.QueryRowContext(ctx, `
			SELECT COALESCE(MAX(position), 0) FROM page_components WHERE page_id = $1
		`, pageID).Scan(&maxPosition)

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO page_components (
				page_id, component_id, position, slot_name,
				rendered_html, content_data, content_hash, build_status
			) VALUES ($1, $2, $3, 'latest-news', $4, $5::jsonb, $6, 'deployed')
		`, pageID, componentID, maxPosition+1, renderedHTML, string(contentDataJSON), contentHash)
		if err != nil {
			return nil, fmt.Errorf("insert page_component: %w", err)
		}

		logger.Info("RenderNewsSectionAction: inserted new latest-news component",
			zap.String("page_id", pageID.String()),
			zap.Int("position", maxPosition+1),
			zap.Int("item_count", len(newsItems)))
	} else if err == nil {
		// Update existing
		_, err = params.DB.ExecContext(ctx, `
			UPDATE page_components
			SET rendered_html = $1,
			    content_data = $2::jsonb,
			    content_hash = $3,
			    build_status = 'deployed',
			    updated_at = NOW()
			WHERE id = $4
		`, renderedHTML, string(contentDataJSON), contentHash, existingID)
		if err != nil {
			return nil, fmt.Errorf("update page_component: %w", err)
		}

		logger.Info("RenderNewsSectionAction: updated existing latest-news component",
			zap.String("component_id", existingID.String()),
			zap.Int("item_count", len(newsItems)))
	} else {
		return nil, fmt.Errorf("query existing component: %w", err)
	}

	return map[string]interface{}{
		"rendered":     true,
		"page_id":      pageID.String(),
		"page_name":    pageName,
		"component_id": componentID.String(),
		"item_count":   len(newsItems),
		"content_hash": contentHash,
	}, nil
}

// truncateSummary truncates a summary to maxLen characters at a word boundary.
func truncateSummary(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Find last space before maxLen
	idx := strings.LastIndex(s[:maxLen], " ")
	if idx < maxLen/2 {
		idx = maxLen
	}
	return s[:idx] + "..."
}

// formatNewsDate formats a time for display in the news card.
// Shows relative time for recent items, date for older ones.
func formatNewsDate(t time.Time) string {
	hours := time.Since(t).Hours()
	if hours < 1 {
		return "Just now"
	}
	if hours < 24 {
		return fmt.Sprintf("%dh ago", int(hours))
	}
	if hours < 48 {
		return "Yesterday"
	}
	if hours < 168 { // 7 days
		return fmt.Sprintf("%dd ago", int(hours/24))
	}
	return t.Format("2 Jan 2006")
}
