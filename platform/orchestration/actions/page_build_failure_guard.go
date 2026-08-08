// FILE: platform/orchestration/actions/page_build_failure_guard.go
//
// Deploy-stamp refusal for content-failed page builds (bugs_open/210).
//
// UpdatePageStatusAction's deploy guards (pageHasComponents, pageSectionShortfall)
// are satisfied by the PREVIOUS build's output, so a rebuild whose content
// generation failed — assemble_page returned {"skipped":true}, deploy_page
// committed nothing — used to be stamped build_status='deployed' with
// built_from_plan_version = the current plan. That stamp makes decideEmit
// (reconcile_site_plan_action.go) return skip_built forever: the rebuild request
// was not failed, not retried, not queued — forgotten, while the page kept
// serving its old content. bugs_open/208 refused the stamp for its own
// OWNED_PAGE_GUARD skips and deliberately filed the general case as 210 because
// widening changes retry behaviour on the fleet's main build path. This file is
// that widening, WITH the retry bound 210 demanded.
//
// The shape of the bound, and why each piece is what it is:
//
//   - Every refusal writes an agent_error_log row (DEPLOY_STAMP_REFUSED_ON_SKIP)
//     FIRST, before the status flip — a failed flip must still leave the trace.
//     This is also the permanent counter 210's measurement section asks for: the
//     obvious proxies were measured and rejected there (a healthy page-rerender
//     produces the same deployed_at-after-components signature), so this row is
//     the only honest frequency signal the fleet has.
//   - The page flips to needs_rebuild + built_from_plan_version NULL — the same
//     statement the sibling 0-component and shortfall guards use. The request
//     stays alive; the existing producers re-emit and retry.
//   - The third refusal since the page's last successful deploy (7-day cap)
//     parks the page: an OPEN page_build_failed item, status='needs_human_review',
//     NO handler, item_key 'needs_page:<name>' — the page-slot key that
//     ReconcileSitePlanAction (raw INSERT), WriteBuildItemsAction (insertWorkItem)
//     and the tool-recreation lane already share — so idx_swi_dedup holds the
//     slot and every automatic producer stops paying for LLM retries until a
//     human closes the item. A later successful deploy stamp closes it
//     automatically (closePageBuildFailureItems).
//
// THE PARK MUST STAY A RAW INSERT. insertWorkItem's two-strike block counts
// terminal ('complete','failed') predecessors on the same (site_id, item_key) —
// and the prior failed builds are exactly such predecessors, dishonestly
// 'complete' (that dishonesty IS bug 210). Routed through insertWorkItem the park
// would be branded 'unresolved' at birth: a TERMINAL status, outside
// idx_swi_dedup's predicate, holding nothing — the bound would silently not
// exist. emitOwnedPageReviewItem (owned_page_guard.go) is the reviewed precedent
// for this exact raw-insert-plus-dedup-index shape.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// pageBuildFailedItemType marks a page parked after repeated content-failed
// builds. No automated consumer, by design: the item exists to hold the page's
// needs_page:<name> dedup slot open and to be read by a human.
const pageBuildFailedItemType = "page_build_failed"

// deployStampRefusedErrorCode is the agent_error_log code every refusal writes.
// It is both the audit trail and the strike counter's key. Pod-greppable to
// prove the guard is on a running binary.
const deployStampRefusedErrorCode = "DEPLOY_STAMP_REFUSED_ON_SKIP"

// pageBuildFailureStrikeLimit: the refusal that reaches this count (since the
// last successful deploy, 7-day cap) parks the page. Three mirrors the estate's
// two-strike convention: two full retries, the third failure escalates.
const pageBuildFailureStrikeLimit = 3

// refuseDeployStampOnSkip handles a deploy-stamp request that arrived after a
// non-ownership assembly skip. It records the refusal, flips the page to
// needs_rebuild, and parks the page on the third consecutive failure. It never
// returns an error: refusal must fail toward retrying, not toward stamping, and
// not toward failing the workflow (which would strand every page after this one
// in the loop — the same posture as AssemblePageAction's own skip).
func refuseDeployStampOnSkip(ctx context.Context, params ActionParams, pageID uuid.UUID, skipReason string) map[string]interface{} {
	logger := params.Logger

	result := map[string]interface{}{
		"updated":      false,
		"page_id":      pageID.String(),
		"build_status": "needs_rebuild",
		"skipped":      true,
		"reason":       "refused deploy stamp: assembly was skipped — " + skipReason,
	}

	var siteID uuid.UUID
	var pageName, domain string
	var deployedAt sql.NullTime
	if err := params.DB.QueryRowContext(ctx, `
		SELECT p.site_id, p.name, s.domain, p.deployed_at
		FROM pages p JOIN sites s ON s.id = p.site_id
		WHERE p.id = $1
	`, pageID).Scan(&siteID, &pageName, &domain, &deployedAt); err != nil {
		// Identity unknown: still refuse the stamp and flip the status (the
		// refusal itself must not depend on the lookup), but the strike count
		// and park have no key to work with.
		logger.Error("refuseDeployStampOnSkip: page identity lookup failed — refusing without strike accounting",
			zap.String("page_id", pageID.String()), zap.Error(err))
		if rbErr := execDB(ctx, params.DB, `UPDATE pages SET build_status = 'needs_rebuild', built_from_plan_version = NULL, updated_at = NOW() WHERE id = $1`, pageID); rbErr != nil {
			logger.Error("refuseDeployStampOnSkip: needs_rebuild flip failed", zap.String("page_id", pageID.String()), zap.Error(rbErr))
		}
		return result
	}

	// 1. The trace, before anything that could fail. Best-effort by contract;
	// a lost row undercounts strikes and grants at most one extra retry.
	LogActionError(ctx, params, siteID.String(), domain, "update_page_status",
		deployStampRefusedErrorCode, "warning",
		fmt.Sprintf("deploy stamp refused for page %s: assembly was skipped — %s", pageName, skipReason),
		map[string]interface{}{
			"page_id":     pageID.String(),
			"page_name":   pageName,
			"skip_reason": skipReason,
		}, logger)

	// 2. The honest state: same statement as the sibling guards above this one.
	if rbErr := execDB(ctx, params.DB, `UPDATE pages SET build_status = 'needs_rebuild', built_from_plan_version = NULL, updated_at = NOW() WHERE id = $1`, pageID); rbErr != nil {
		logger.Error("refuseDeployStampOnSkip: needs_rebuild flip failed",
			zap.String("page_id", pageID.String()), zap.Error(rbErr))
	}

	// 3. Strikes since the last successful deploy, capped at 7 days (matching
	// the two-strike window). pages.deployed_at is only ever written by a
	// successful stamp — this guard refuses before it — so it IS the
	// last-success marker, and an intervening success resets the count.
	var strikes int
	if err := params.DB.QueryRowContext(ctx, `
		SELECT count(*) FROM agent_error_log
		WHERE error_code = $1
		  AND context->>'page_id' = $2
		  AND occurred_at > NOW() - INTERVAL '7 days'
		  AND ($3::timestamptz IS NULL OR occurred_at > $3)
	`, deployStampRefusedErrorCode, pageID.String(), deployedAt).Scan(&strikes); err != nil {
		logger.Warn("refuseDeployStampOnSkip: strike count unavailable — refusing without park decision",
			zap.String("page_id", pageID.String()), zap.Error(err))
		return result
	}
	result["refusals"] = strikes

	if strikes >= pageBuildFailureStrikeLimit {
		parkPageBuildFailure(ctx, params.DB, siteID, pageID, pageName, skipReason, strikes, logger)
		result["parked"] = true
	}

	logger.Warn("refuseDeployStampOnSkip: deploy stamp refused after assembly skip",
		zap.String("page_id", pageID.String()),
		zap.String("page_name", pageName),
		zap.String("skip_reason", skipReason),
		zap.Int("refusals_since_last_success", strikes))
	return result
}

// parkPageBuildFailure files the strike-limit escalation. RAW insert, arbitrated
// by idx_swi_dedup via bare ON CONFLICT DO NOTHING — see the file header for why
// insertWorkItem must not be used here. While this row is open it holds the
// page's needs_page:<name> slot against ReconcileSitePlanAction,
// WriteBuildItemsAction and tool-recreation inserts alike; a human closing it —
// or any later successful deploy stamp — reopens the slot.
func parkPageBuildFailure(ctx context.Context, db *sql.DB, siteID, pageID uuid.UUID, pageName, skipReason string, strikes int, logger *zap.Logger) {
	spec, err := json.Marshal(map[string]interface{}{
		// "bug" is the queue-drain convention mistyped_deployed_page established
		// (LANDMINES): a needs_human_review row with no handler must say which
		// decision it is, so a drain never routes it to a model to guess at.
		"bug":         "bugs_open/210",
		"page_name":   pageName,
		"page_id":     pageID.String(),
		"skip_reason": skipReason,
		"refusals":    strikes,
		"fix": fmt.Sprintf("Content generation for this page failed %d times; automatic rebuilds are "+
			"parked so they stop spending LLM calls on a failing input. Fix the cause (page spec, "+
			"provider quota, unsatisfiable component — see skip_reason), then close this item; the "+
			"next successful deploy also closes it automatically.", strikes),
	})
	if err != nil {
		logger.Warn("parkPageBuildFailure: could not marshal spec", zap.Error(err))
		return
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, page_id, priority, status, created_by, item_key
		) VALUES ($1, 'update_page_status', 'build', 'page_build_failed',
		          'high', $2, $3::jsonb, $4, 40,
		          'needs_human_review', 'update_page_status', $5)
		ON CONFLICT DO NOTHING
	`, siteID,
		fmt.Sprintf("Page %s: content generation failed %d times — rebuild parked for a human", pageName, strikes),
		string(spec),
		pageID,
		"needs_page:"+pageName,
	); err != nil {
		logger.Warn("parkPageBuildFailure: park insert failed — retries stay unbounded for this page",
			zap.String("page_name", pageName), zap.Error(err))
		return
	}

	logger.Warn("parkPageBuildFailure: page parked after repeated content failures",
		zap.String("page_name", pageName),
		zap.String("site_id", siteID.String()),
		zap.Int("refusals", strikes))
}

// closePageBuildFailureItems completes any open park for the page after a
// successful deploy stamp. Success is the definitive evidence the parked
// condition is resolved; an open park would otherwise outlive its truth and
// keep the page's work-item slot blocked. Errors are logged and swallowed: the
// deploy itself is already committed and must not be re-run for a bookkeeping
// failure (same posture as the page_component mirror below the stamp).
func closePageBuildFailureItems(ctx context.Context, db *sql.DB, pageID uuid.UUID, logger *zap.Logger) {
	res, err := db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete', updated_at = NOW()
		WHERE item_type = 'page_build_failed'
		  AND page_id = $1
		  AND status = 'needs_human_review'
	`, pageID)
	if err != nil {
		logger.Warn("closePageBuildFailureItems: close failed — an open park may outlive its truth",
			zap.String("page_id", pageID.String()), zap.Error(err))
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		logger.Info("closePageBuildFailureItems: park closed by successful deploy",
			zap.String("page_id", pageID.String()), zap.Int64("items_closed", n))
	}
}
