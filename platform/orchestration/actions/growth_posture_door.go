// FILE: platform/orchestration/actions/growth_posture_door.go
//
// The GROWTH POSTURE door (owner decision 5 of 2026-08-31, register WDS-020):
// a growth-gated item filed against a site whose
// sites.settings->maintenance_profile->>growth_posture is 'hold' is written in
// the record shape — status 'deferred', handler_agent '' — instead of being
// dispatched. FILED, never skipped: the signal that the machinery wanted to
// grow the site is kept as a durable, releasable row; release is a one-UPDATE
// human verb from the recipe stamped on the row's own spec.
//
// WHY A DOOR IN writeWorkItem AND NOT A GUARD PER PRODUCER. The first cut of
// this change guarded the two producers a 30-day census had found
// (check_missing_tools for evaluate_tools; create_work_item for tool-suggester's
// add_tool). The council's round-1 objections (corr 1e735fa2) were right twice
// over: the census ran over site_work_items, a rolling-window table that cannot
// prove producer completeness, and per-producer guards silently miss the next
// producer — the exact class LANDMINES records as "a guard only guards the door
// you walk through". writeWorkItem is the seam every filing crosses
// (insertWorkItem wraps it; discovery sweeps, config-driven steps and direct Go
// callers all pass through), and it already hosts two policy doors of exactly
// this shape — the owned-page door and the unregistered-handler demotion. The
// workflow-config alternative (a conditional_branch inside tool-suggester's own
// workflow) was considered and rejected for the same reason stated at those
// doors: it binds ONE producer, not the TYPE, and a config step is a
// convention, not a control, on a tree this many sessions share.
//
// ORDERING (mirrors the owned-page door's argument verbatim): the door runs
// BEFORE the anti-churn brake, so the strike count is never consulted for a
// row we are not dispatching; 291's registration block then sees `deferred`
// and correctly skips (workItemStatusRequiresRegisteredHandler excludes it).
//
// THE HELD ITEM KEEPS ITS item_type AND item_key — the ownedPageParkedItem
// lesson, learned via bugs_open/342: `deferred` is not in
// workItemClosedStatuses, so the held row is retracted normally if its
// detector stops reproducing it, and it holds its dedup slot so re-finds
// collapse onto it instead of stacking. recurrenceExpected is set for the
// same reason it is set there: a held row's eventual close is a hold ending
// or a retraction, not a strike against a handler that never ran.
//
// Kill-switch DISABLE_GROWTH_POSTURE_DOOR ships ARMED (the owner has ruled
// against default-OFF switches that rot unexercised); it exists to be
// disarmed in anger, not to gate the feature.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

const growthHeldPrefix = "[growth held] "

// applyGrowthPostureDoor decides whether this write is held, and returns the
// (possibly transformed) item. Fail-open, loudly: GROWTH_DOOR_PROBE_FAILED is
// a stable literal so a fail-open is countable in logs (the owned-page door's
// bug_historian convention) — deliberately a log line and not an
// agent_error_log row, which would ride this same transaction.
func applyGrowthPostureDoor(ctx context.Context, tx *sql.Tx, item workItem, logger *zap.Logger) workItem {
	if !datahelpers.GrowthGateApplies(item.itemType, item.source) ||
		!workItemStatusHeadsForDispatch(item.status) ||
		os.Getenv("DISABLE_GROWTH_POSTURE_DOOR") != "" {
		return item
	}
	posture, err := datahelpers.SiteGrowthPosture(ctx, tx, item.siteID.String())
	if err != nil {
		logger.Warn("writeWorkItem: GROWTH_DOOR_PROBE_FAILED growth_posture unreadable — "+
			"door standing down, item files as if the site were open",
			zap.String("site_id", item.siteID.String()), zap.Error(err))
		return item
	}
	if posture != datahelpers.GrowthPostureHold {
		return item
	}

	logger.Warn("writeWorkItem: GROWTH HELD — site posture is 'hold', filing as a record, not dispatching",
		zap.String("site_id", item.siteID.String()),
		zap.String("item_type", item.itemType),
		zap.String("item_key", item.itemKey),
		zap.String("would_have_handled", item.handlerAgent))

	held := item
	heldHandler := item.handlerAgent
	held.status = "deferred"
	held.handlerAgent = ""
	held.recurrenceExpected = true

	// Truncate the ORIGINAL, then prefix — ownedPageParkedItem's rule, so the
	// marker that tells a human why the row is here survives truncation.
	summary := item.summary
	if max := workItemSummaryMaxLen - len(growthHeldPrefix); len(summary) > max {
		summary = summary[:max-3] + "..."
	}
	held.summary = growthHeldPrefix + summary

	// The producer's spec is preserved at the top level; an unparsable spec is
	// kept verbatim rather than dropped (same idiom as the owned-page park).
	spec := map[string]interface{}{}
	if item.spec != "" {
		if err := json.Unmarshal([]byte(item.spec), &spec); err != nil {
			spec = map[string]interface{}{"spec_raw": item.spec}
		}
	}
	spec["growth_held"] = true
	spec["growth_handler"] = heldHandler
	spec["growth_release_recipe"] = "owner release: UPDATE site_work_items SET status='detected', " +
		"handler_agent=spec->>'growth_handler' WHERE id='<this row>'"
	if b, err := json.Marshal(spec); err != nil {
		// Marshal of a just-unmarshalled map plus three scalars cannot
		// realistically fail; if it somehow does, keep the original spec and
		// still hold the row — losing the stamps is better than dispatching.
		logger.Warn("writeWorkItem: growth door spec marshal failed — holding with original spec",
			zap.Error(err))
	} else {
		held.spec = string(b)
	}
	return held
}
