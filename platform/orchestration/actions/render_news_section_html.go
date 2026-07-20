// FILE: platform/orchestration/actions/render_news_section_html.go
//
// Freshness delivery for server-rendered news (bugs_open/027, reworked).
//
// HISTORY, kept because the wrong version shipped: the first implementation
// in this file injected news HTML into page_components.rendered_html by
// string-anchor surgery (persistNewsSectionHTML / injectNewsItems). It went
// live in v1.0.1140 and was then removed: 003's source-of-truth contract
// rejects HTML patching outright — a scoped rerender regenerates
// rendered_html from html_template + content_data, so anything living in
// neither is silently wiped. Council REVISE 4b91237a (render_guardian) caught
// it; the operator's guidelines check confirmed it.
//
// The replacement: news items are query.latest_news / query.news_archive
// schema fields (queryresolve/news_items.go), rendered by the components' own
// html_template into content_data-backed HTML. This file now holds only the
// freshness trigger — after a feed refresh, queue the scoped re-render that
// makes the static HTML track the feed.

package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// queueNewsPageRerenders emits one scoped re-render work item per page that
// carries a news component, so freshly ingested items reach the deployed HTML.
//
// Shape copied from reconcile_section_data / flag_page_image_rebuild — the
// two existing "data resolved, re-render the page" emitters: item_type
// needs_page with spec.reason section_data_resolved routes page-build-handler
// down its scoped path (re-resolve queries, re-render from stored
// content_data, no LLM), and item_key page_rerender:<page> collapses
// concurrent triggers through idx_swi_dedup.
//
// Best-effort by design: the JSON path has already succeeded when this runs,
// and a duplicate-suppressed insert (two-strike rule, open item with the same
// key) is normal, not an error. Returns the number of items actually queued.
func queueNewsPageRerenders(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) int {
	if db == nil {
		return 0
	}

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.name
		FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND cc.function IN ('latest-news', 'news-listing')
		  AND p.build_status = 'deployed'
	`, siteID)
	if err != nil {
		logger.Warn("queueNewsPageRerenders: page lookup failed", zap.Error(err))
		return 0
	}
	defer rows.Close()

	var pages []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			logger.Warn("queueNewsPageRerenders: scan failed", zap.Error(err))
			continue
		}
		pages = append(pages, name)
	}
	if len(pages) == 0 {
		return 0
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("queueNewsPageRerenders: begin tx failed", zap.Error(err))
		return 0
	}
	defer tx.Rollback()

	batchID := uuid.New()
	queued := 0
	for _, page := range pages {
		item := workItem{
			siteID:       siteID,
			source:       "render_news_section",
			pipeline:     "build",
			itemType:     "needs_page",
			severity:     "low",
			summary:      fmt.Sprintf("Re-render %s — fresh news items available", page),
			spec:         fmt.Sprintf(`{"reason":"section_data_resolved","page_name":%q}`, page),
			priority:     99,
			handlerAgent: "page-build-handler",
			status:       "triaged",
			createdBy:    "render_news_section",
			itemKey:      fmt.Sprintf("page_rerender:%s", page),
			batchID:      batchID,

			// A re-request every feed cycle is the design, not a failed fix:
			// fresh items are a recurring event, so suppress the two-strike
			// unresolved labelling for these.
			recurrenceExpected: true,
		}
		inserted, err := insertWorkItem(ctx, tx, item, logger)
		if err != nil {
			logger.Warn("queueNewsPageRerenders: insert failed",
				zap.String("page", page), zap.Error(err))
			continue
		}
		if inserted {
			queued++
		}
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("queueNewsPageRerenders: commit failed", zap.Error(err))
		return 0
	}

	if queued > 0 {
		logger.Info("queueNewsPageRerenders: queued scoped re-renders",
			zap.Int("queued", queued), zap.Int("news_pages", len(pages)))
	}
	return queued
}
