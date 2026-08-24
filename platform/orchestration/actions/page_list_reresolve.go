// FILE: platform/orchestration/actions/page_list_reresolve.go
//
// bugs_open/384. A page's listing card lands, is entity-linked and joinable —
// and the listing that renders it keeps showing a text-only card. The listing's
// items live in a STORED array (page_components.content_data, filled from a
// `query.*` source at the last section resolve), and every assemble-mode
// re-render re-ships that array verbatim. Only a re-render carrying
// spec.reason='section_data_resolved' (page-rerender → rerender_page_sections)
// re-runs the query. Nothing in the card-landing chain asked for that mode, so
// the listing was re-rendered three times after the cards landed and stayed
// stale each time.
//
// THE RULE THIS FILE IMPLEMENTS
//   A producer that changes the data behind a page-list query source files a
//   section_data_resolved re-render for every page on the site that consumes
//   one. Producers name the CAUSE; the consumer set is derived, once, by
//   queryresolve.PageListConsumerPages from content_components.input_schema —
//   never hard-coded per producer (render_news_section and render_directory
//   each name their one consumer page; a third such name would be the one that
//   goes stale when a site grows a second listing).
//
// SHAPE
//   item_type page_rerender → page-rerender, via the canonical
//   insertPageRerenderItem (a raw INSERT ... ON CONFLICT DO NOTHING against
//   idx_swi_dedup; no writeWorkItem, so no anti-churn brake and no
//   recurrenceExpected — an action request from a render-side producer, the
//   render_news_section posture, agreed with the bugs_open/326 lane 2026-08-24).
//   Key: pageRerenderItemKey(page, site, "section_data_resolved"), shared with
//   the page_list_stale sweep through discovery_checks.PageRerenderItemKey.
//   Owned pages are excluded at the lookup (they fail save_sections' ownership
//   refusal; the per-agent owned-page door cannot see spec.reason —
//   bugs_open/333 lane, 2026-08-24).
//
// POSTURE: NEVER FAILS THE CALLER. By the time either caller reaches this the
// artefact side effects have happened (the card is committed, the hero is
// deployed). Failing the action would retry those, not this. Every outcome is
// a disposition string in the caller's return map and log line, so a skip is
// visible rather than silent — the same construction as emitContentCardDerive.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"go.uber.org/zap"
)

// pageListReresolveReason is the ONE reason value that makes page-rerender
// re-run query.* sources for a stored array (check_rerender_mode; STY-048).
const pageListReresolveReason = "section_data_resolved"

// maxPageListReresolvePerEvent bounds what ONE landing may file (council round
// c2873f56, guardian seat: "per-event consumer-count … acknowledged or
// bounded"). The structural bound is the lookup's renders-image predicate
// (0–3 consumer pages per site on 2026-08-24, max robot-hands.com at 3); this
// is the belt behind it, at queryresolve's own hard cap — one page-list's
// worth. Exceeding it is NOT silent: the remainder is logged at Warn with the
// page names and reported as Capped, and the page_list_stale sweep files the
// rest on its next visit.
const maxPageListReresolvePerEvent = 24

// pageListReresolveExec is the slice of *sql.DB / *sql.Tx this needs: the
// consumer lookup (QueryContext) and the canonical INSERT (ExecContext).
type pageListReresolveExec interface {
	rerenderItemExec
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// pageListReresolve is what one invalidation did, for the caller's return map.
type pageListReresolve struct {
	// Disposition: lookup_failed | no_consumers | queued | deduped | partial | insert_failed
	Disposition string
	Consumers   int
	Queued      int
	Deduped     int
	Failed      int
	// Capped counts consumer pages NOT filed because the per-event bound was
	// hit; they are named in the Warn log and left to the sweep.
	Capped int
	Pages  []string
}

func (r pageListReresolve) fields() map[string]interface{} {
	return map[string]interface{}{
		"page_list_reresolve":           r.Disposition,
		"page_list_reresolve_consumers": r.Consumers,
		"page_list_reresolve_queued":    r.Queued,
		"page_list_reresolve_deduped":   r.Deduped,
		"page_list_reresolve_capped":    r.Capped,
		"page_list_reresolve_pages":     r.Pages,
	}
}

// requestPageListReresolve files one section_data_resolved page_rerender per
// consumer page on the site. `source` is the producer (it becomes the row's
// source AND created_by, so page_rerender attribution stays per-producer);
// `cause` is the forensic why, carried in spec.cause (e.g. "card_landed:<page>").
func requestPageListReresolve(
	ctx context.Context,
	exec pageListReresolveExec,
	siteID uuid.UUID,
	source string,
	cause string,
	batchID uuid.UUID,
	logger *zap.Logger,
) pageListReresolve {
	if logger == nil {
		logger = zap.NewNop()
	}
	consumers, err := queryresolve.PageListConsumerPages(ctx, exec, siteID, logger)
	if err != nil {
		logger.Error("page_list_reresolve: consumer lookup FAILED — no listing will be told; the stored arrays on this site stay stale until an unrelated section re-resolve (bugs_open/384)",
			zap.String("site_id", siteID.String()), zap.String("cause", cause), zap.Error(err))
		return pageListReresolve{Disposition: "lookup_failed"}
	}
	out := pageListReresolve{Consumers: len(consumers)}
	if len(consumers) == 0 {
		out.Disposition = "no_consumers"
		return out
	}
	if len(consumers) > maxPageListReresolvePerEvent {
		dropped := make([]string, 0, len(consumers)-maxPageListReresolvePerEvent)
		for _, c := range consumers[maxPageListReresolvePerEvent:] {
			dropped = append(dropped, c.Name)
		}
		out.Capped = len(dropped)
		logger.Warn("page_list_reresolve: consumer count exceeds the per-event bound — the remainder is NOT filed here and is left to the page_list_stale sweep",
			zap.String("site_id", siteID.String()), zap.String("cause", cause),
			zap.Int("consumers", len(consumers)), zap.Int("bound", maxPageListReresolvePerEvent),
			zap.Strings("not_filed", dropped))
		consumers = consumers[:maxPageListReresolvePerEvent]
	}

	for _, c := range consumers {
		specJSON, err := json.Marshal(map[string]interface{}{
			"reason":    pageListReresolveReason,
			"page_name": c.Name,
			"page_id":   c.ID.String(),
			"domain":    c.Domain,
			"cause":     cause,
			"consumes":  c.Sources(),
		})
		if err != nil {
			out.Failed++
			continue
		}
		itemKey := pageRerenderItemKey(c.Name, siteID, pageListReresolveReason)
		inserted, err := insertPageRerenderItem(ctx, exec, siteID, c.ID,
			source, "low",
			fmt.Sprintf("Re-render %s — page-list data changed (%s)", c.Name, cause),
			string(specJSON), itemKey, batchID)
		switch {
		case err != nil:
			out.Failed++
			logger.Warn("page_list_reresolve: insert failed for one consumer page",
				zap.String("page", c.Name), zap.String("cause", cause), zap.Error(err))
		case inserted:
			out.Queued++
			out.Pages = append(out.Pages, c.Name)
		default:
			// An open item with the same key already holds the slot — a
			// re-resolve is already queued for this page and will run against
			// the data as it is when it runs, which includes this change.
			out.Deduped++
		}
	}

	switch {
	case out.Queued > 0 && out.Failed == 0:
		out.Disposition = "queued"
	case out.Queued > 0:
		out.Disposition = "partial"
	case out.Failed > 0:
		out.Disposition = "insert_failed"
	default:
		out.Disposition = "deduped"
	}
	logger.Info("page_list_reresolve: told the listings",
		zap.String("site_id", siteID.String()), zap.String("cause", cause),
		zap.String("disposition", out.Disposition), zap.Int("consumers", out.Consumers),
		zap.Int("queued", out.Queued), zap.Int("deduped", out.Deduped), zap.Int("failed", out.Failed),
		zap.Strings("pages", out.Pages))
	return out
}

// reresolvePageListsAfterCard is derive_card_asset's call. The card is what
// queryresolve's pageImageJoins projects into every page-list item on the
// site, so its landing is exactly the event the listings must hear about.
// Skipped when the provenance upsert was lock-suppressed: the join reads the
// row, and the row did not change. Returns the fields for the action's result.
func reresolvePageListsAfterCard(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageName string,
	provenanceRecorded bool,
	logger *zap.Logger,
) map[string]interface{} {
	if !provenanceRecorded {
		return pageListReresolve{Disposition: "skipped_provenance_not_recorded"}.fields()
	}
	return requestPageListReresolve(ctx, db, siteID, "derive_card_asset", "card_landed:"+pageName, uuid.New(), logger).fields()
}

// reresolvePageListsAfterPageImage is flag_page_image_rebuild's call, inside
// its transaction. A landed page image is the listings' fallback image (the
// `ha` arm of pageImageJoins) — UNLESS a card derive was just raised for the
// page, in which case the card will supersede it and derive_card_asset
// requests the re-resolve when the card lands; re-resolving now as well would
// show the hero for a few minutes and then re-render again. Returns the
// disposition for the caller's log line and result map.
func reresolvePageListsAfterPageImage(
	ctx context.Context,
	tx *sql.Tx,
	siteID uuid.UUID,
	pageName string,
	cardEmit string,
	batchID uuid.UUID,
	logger *zap.Logger,
) string {
	if cardEmit == "raised" {
		return "deferred_to_card_derive"
	}
	return requestPageListReresolve(ctx, tx, siteID, "image-build-handler", "page_image_landed:"+pageName, batchID, logger).Disposition
}
