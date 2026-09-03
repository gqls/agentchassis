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

	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
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

	// ⚠ THE PAGE-STATUS RULE THIS FUNCTION LEARNED THE HARD WAY NOW LIVES IN THE
	// SHARED LOOKUP (queryresolve.consumerSQL). Keeping the history here because
	// the next person to touch either place needs it:
	//
	// `p.status = 'active'` is NOT redundant beside `build_status = 'deployed'`,
	// and leaving it out RESURRECTS RETIRED PAGES (bugs_open/098). The two
	// columns answer different questions and nothing keeps them in step:
	// `build_status` records whether the page ever shipped, `status` records
	// whether the platform still wants it served. Archiving a page sets
	// `status='archived'` and leaves `build_status='deployed'` untouched — so a
	// selector keyed on build_status alone keeps choosing it for ever.
	//
	// Observed, not theorised: robot-hands.com/learning-center-index is archived
	// and was re-rendered and re-committed to the sites repo TWICE A DAY
	// (2026-08-01 08:07 and 20:06, 08-02 08:05 and 20:15, 08-03 08:15 — six
	// page_rerender items raised by this function since 07-31). 098 was filed on
	// 07-26 describing archived pages as FROZEN; this path had made that false
	// by the time it was fixed. It also makes a retraction self-undoing: delete
	// the file and the next news refresh republishes it.
	//
	// The same defect is STILL LIVE one seam along, in
	// component-template-fixer.create_rerender, which has no page-status filter
	// at all — see bugs_open/. Do not copy that query.
	//
	// MIGRATED 2026-08-25 onto the shared consumer lookup (RFC_052, owner ruling
	// "generalise it now"). This used to run its own SQL selecting pages by
	// COMPONENT FUNCTION — `cc.function IN ('latest-news','news-listing')` — a
	// hand-kept component list that goes stale the day a component is renamed or
	// a second component starts consuming query.latest_news / query.news_archive.
	// queryresolve derives the same set from what the components DECLARE, so
	// there is now one answer to "who consumes my data" rather than three
	// spellings of it.
	//
	// The predicates that were hand-written here live in the shared lookup now
	// and are NOT lost: PageHasShippedPredicateFor was added to consumerSQL with
	// this migration precisely so it could not weaken this call site, and the
	// live-page half is covered by the lookup's own status filter. The
	// owned-page exclusion is NEW here and is correct — page-rerender's reasoned
	// branch runs save_sections, which refuses an owned page (bugs_open/208).
	//
	// MEASURED BEFORE MIGRATING [2026-08-25], fleet-wide: the function route and
	// the schema route select the SAME 16 pages — 16 in both, 0 function-only,
	// 0 schema-only. So this is a no-op on today's fleet, which is what makes it
	// reviewable rather than a leap.
	//
	// ROUTE AND ITEM SHAPE ARE DELIBERATELY UNCHANGED. Only the page SELECTION
	// moved. This still files page_rerender via insertPageRerenderItem with
	// reason=section_data_resolved. Changing the route is what made this
	// emitter's first version expensive — see the CORRECTED block above.
	pages, err := queryresolve.ConsumerPages(ctx, db, siteID, queryresolve.DepFeedItems, logger)
	if err != nil {
		logger.Warn("queueNewsPageRerenders: consumer lookup failed", zap.Error(err))
		return 0
	}
	if len(pages) == 0 {
		return 0
	}

	batchID := uuid.New()
	queued := 0
	for _, page := range pages {
		// The canonical page_rerender shape via the shared helper (ONE
		// literal INSERT, two emitters — see insertPageRerenderItem). The
		// reason in spec makes page-rerender take rerender_page_sections
		// (its gate is spec.reason alone — verified in the live workflow's
		// conditional); the pageRerenderItemKey reason-suffix means the
		// reason-stamped mode can never be dedup-suppressed by an
		// assemble-only item.
		spec := fmt.Sprintf(
			`{%s"page_name":%q,"page_id":%q,"domain":%q}`,
			livespec.RerenderReasonJSONPrefix(livespec.ReasonSectionDataResolved),
			page.Name, page.ID.String(), page.Domain)
		itemKey := pageRerenderItemKey(page.Name, siteID, "section_data_resolved")

		inserted, err := insertPageRerenderItem(ctx, db, siteID, page.ID,
			"render_news_section", "low",
			fmt.Sprintf("Re-render %s — fresh news items available", page.Name),
			spec, itemKey, batchID)
		if err != nil {
			logger.Warn("queueNewsPageRerenders: insert failed",
				zap.String("page", page.Name), zap.Error(err))
			continue
		}
		if inserted {
			queued++
		}
	}

	if queued > 0 {
		logger.Info("queueNewsPageRerenders: queued scoped re-renders",
			zap.Int("queued", queued), zap.Int("news_pages", len(pages)))
	}
	return queued
}
