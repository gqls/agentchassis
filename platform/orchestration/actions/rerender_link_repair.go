// FILE: platform/orchestration/actions/rerender_link_repair.go
//
// The shared dead-internal-link repair for RERENDER paths (bugs_open/097).
//
// RepairPageLinks (bugs_open/079) resolves every internal <a href> in an
// assembled page against pages.url — rewriting fixable ones, unlinking
// phantoms. Until this file it had exactly ONE call site: the initial-build
// gate (ValidatePageContentAction), wired to page-build-handler's
// validate_content step. Neither rerender path invokes that step, so a page
// was protected on first build and unprotected on every rebuild — diagnosis
// corr 9543aaf1 confirmed the robot-hands.com learning-center rebuild of
// 2026-07-25 went through the bulk path and redeployed six 404s the gate
// would have repaired.
//
// This is the sibling-path drift the rerender file already records twice (the
// bugs_open/072 CSS call, the collectJSAssets asymmetry note): every
// post-processing improvement has had to be hand-copied between the bulk and
// single-page paths. So the repair lives HERE, in a file neither path owns,
// and both call one function at the same altitude — the last step before the
// page's HTML leaves for deploy.

package actions

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// repairOutboundPageLinks applies the build gate's dead-internal-link repair to
// a fully assembled page at the LAST step before it leaves a rerender path.
//
// Fail-open: with no usable page index the HTML ships unchanged and the skip is
// logged loudly — a rerender must not be blocked by a failed metadata read,
// which is the same degrade the gate takes (its pageIndexOK guard). "Unchanged"
// is no worse than every rerender before this file existed.
func repairOutboundPageLinks(ctx context.Context, params ActionParams, siteID uuid.UUID, domain, pageName, pageURL, html string, pageIndex datahelpers.PageURLIndex, indexOK bool, logger *zap.Logger) string {
	if !indexOK {
		logger.Warn("rerender link repair SKIPPED — page index unavailable; deploying unrepaired",
			zap.String("page", pageName),
			zap.String("domain", domain))
		// A pod log line is not a record (071 gap 3, and the council's round-2
		// objection to this very branch): a page that shipped UNREPAIRED because
		// the index failed to load must be findable a day later. Best-effort,
		// like the repair log itself — the skip record must not fail the deploy
		// it describes.
		if params.DB != nil {
			skipCtx, _ := json.Marshal(map[string]string{"page_name": pageName, "page_url": pageURL})
			if _, err := params.DB.ExecContext(ctx, `
				INSERT INTO agent_error_log
				    (site_id, domain, agent_type, step_name, action,
				     error_message, error_code, severity, context)
				VALUES (NULLIF($1,'')::uuid, NULLIF($2, ''), 'page-rerender', $3, 'rerender_page',
				        $4, 'CONTENT_LINK_REPAIR_SKIPPED', 'warning', $5::jsonb)`,
				siteID.String(), domain, skipStepName(params),
				"Rerender link repair SKIPPED — page index unavailable; page deployed unrepaired",
				string(skipCtx),
			); err != nil {
				logger.Warn("rerender link repair: failed to write skip record", zap.Error(err))
			}
		}
		return html
	}
	repaired, repairs := datahelpers.RepairPageLinks(html, pageIndex)
	if len(repairs) == 0 {
		return html
	}
	rewritten, unlinked := countLinkRepairs(repairs)
	logger.Info("rerender: repaired dead internal links before deploy",
		zap.String("page", pageName),
		zap.String("domain", domain),
		zap.Int("rewritten", rewritten),
		zap.Int("unlinked", unlinked))
	writeLinkRepairLog(ctx, params, siteID.String(), domain,
		linkRepairOrigin{
			AgentType:  "page-rerender",
			StepName:   skipStepName(params),
			ActionName: "rerender_page",
			PageName:   pageName,
			PageURL:    pageURL,
		},
		repairs, rewritten, unlinked, logger)
	return repaired
}

// skipStepName names the workflow step for the durable record, degrading to a
// stable label when the execution context is absent (unit tests, odd adoptions).
func skipStepName(params ActionParams) string {
	if params.ExecutionContext != nil && params.ExecutionContext.StepName != "" {
		return params.ExecutionContext.StepName
	}
	return "rerender"
}
