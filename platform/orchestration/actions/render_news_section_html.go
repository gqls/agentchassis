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
// > CORRECTED 2026-07-24 — the first version of this emitter was MIS-ROUTED,
// > and the mistake was expensive: it emitted item_type `needs_page` →
// > page-build-handler, copied from reconcile_section_data /
// > flag_page_image_rebuild on the belief that spec.reason selected a scoped
// > no-LLM branch there. It does not. Those two emitters use needs_page
// > DELIBERATELY because their cases need plan_sections + the writer (a
// > deferred field is absent from content_data, so only a full build can
// > backfill it). For THIS emitter the items already live in content_data, and
// > `needs_page` meant a FULL LLM REBUILD of every news page on every feed
// > cycle — copy-regeneration roulette 4x/day. On 2026-07-24 a roll of that
// > roulette re-invented two phantom links and FABRICATED A CONTACT EMAIL on
// > the relojistas homepage. The council's round-2 objections (bug_historian's
// > escalation concern, prior_art's "reuse shape unverified") were pointing at
// > exactly this and were triaged too generously.
//
// The correct route is the LIGHT one: item_type `page_rerender` →
// page-rerender. With spec.reason ∈ {section_data_resolved, image_landed}
// that handler runs rerender_page_sections — re-resolves query.* fields
// (including our news queries) against stored content_data, re-renders
// templates, assembles and deploys. No LLM anywhere. 002's routing table
// says exactly this; I mis-copied from the wrong row.
//
// item_key follows pageRerenderItemKey (create_rerender_items_action.go:111):
// the reason is part of the key so a reason-stamped item can never be
// dedup-suppressed by an assemble-only item or vice versa (bugs_open/024
// defect 6).
//
// Best-effort by design: the JSON path has already succeeded when this runs,
// and a duplicate-suppressed insert (open item with the same key) is normal,
// not an error. Returns the number of items actually queued.
func queueNewsPageRerenders(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) int {
	if db == nil {
		return 0
	}

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.id, p.name, s.domain
		FROM pages p
		JOIN sites s ON s.id = p.site_id
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

	type newsPage struct {
		id     uuid.UUID
		name   string
		domain string
	}
	var pages []newsPage
	for rows.Next() {
		var np newsPage
		if err := rows.Scan(&np.id, &np.name, &np.domain); err != nil {
			logger.Warn("queueNewsPageRerenders: scan failed", zap.Error(err))
			continue
		}
		pages = append(pages, np)
	}
	if len(pages) == 0 {
		return 0
	}

	batchID := uuid.New()
	queued := 0
	for _, page := range pages {
		// The canonical page_rerender shape (create_rerender_items_action.go
		// ~:282): item_type page_rerender → page-rerender, reason in spec so
		// the handler takes rerender_page_sections, key from
		// pageRerenderItemKey so the reason-stamped mode can never be
		// dedup-suppressed by an assemble-only item.
		spec := fmt.Sprintf(
			`{"reason":"section_data_resolved","page_name":%q,"page_id":%q,"domain":%q}`,
			page.name, page.id.String(), page.domain)
		itemKey := pageRerenderItemKey(page.name, siteID, "section_data_resolved")

		res, err := db.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				page_id, priority, handler_agent, status, created_by,
				spec, item_key, batch_id
			) VALUES ($1, 'render_news_section', 'build', 'page_rerender',
			          'low', $2, $3, 80, 'page-rerender', 'triaged',
			          'render_news_section', $4::jsonb, $5, $6)
			ON CONFLICT DO NOTHING
		`, siteID,
			fmt.Sprintf("Re-render %s — fresh news items available", page.name),
			page.id, spec, itemKey, batchID)
		if err != nil {
			logger.Warn("queueNewsPageRerenders: insert failed",
				zap.String("page", page.name), zap.Error(err))
			continue
		}
		if n, _ := res.RowsAffected(); n > 0 {
			queued++
		}
	}

	if queued > 0 {
		logger.Info("queueNewsPageRerenders: queued scoped re-renders",
			zap.Int("queued", queued), zap.Int("news_pages", len(pages)))
	}
	return queued
}
