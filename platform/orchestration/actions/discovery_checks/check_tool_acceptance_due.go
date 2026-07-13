// FILE: platform/orchestration/actions/discovery_checks/check_tool_acceptance_due.go
//
// Discovery check: tool_acceptance_due — the periodic trigger that makes
// Tier 4 continuous (RUNBOOK Stage 6 trigger points). For every active tool
// on the site whose page is DEPLOYED and whose current PLAN carries a
// ```criteria fence, it emits ONE acceptance_run work item (handler:
// tool-acceptance-agent) unless a run is already recorded or pending:
//
//   - cooldown: an acceptance-run OR acceptance-fail note for the subject in
//     the last N days (default 7) means the tool was recently verified — skip;
//   - an OPEN acceptance_run item for the function means one is queued — skip.
//
// Items are emitted at status 'triaged' + pipeline 'build' (the
// create_rerender_items precedent): acceptance needs no human judgment and
// this site's triage pass is not guaranteed to sweep — 'detected' items can
// sit indefinitely (observed on the first Tier-2 items). Priority 90 places
// runs AFTER page builds/rerenders in the queue, so acceptance tests the NEW
// page, not the one about to be replaced.
//
// The dispatch loop's input_mapping passes `spec` whole, so the handler
// receives input_data.spec.function — exactly tool-acceptance-agent's
// contract (migration 145).

package discovery_checks

import (
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&ToolAcceptanceDueCheck{}) }

type ToolAcceptanceDueCheck struct{}

func (c *ToolAcceptanceDueCheck) Name() string { return "tool_acceptance_due" }

func (c *ToolAcceptanceDueCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT
			cc.id::text   AS component_id,
			cc.function,
			p.id::text    AS page_id,
			p.name        AS page_name
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND cc.component_level = 'tool'
		  AND cc.is_active = true
		  AND p.status = 'active'
		  AND p.build_status = 'deployed'
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("tool_acceptance_due: tool query failed: %w", err)
	}
	defer rows.Close()

	type toolRow struct{ ComponentID, Function, PageID, PageName string }
	var tools []toolRow
	for rows.Next() {
		var t toolRow
		if err := rows.Scan(&t.ComponentID, &t.Function, &t.PageID, &t.PageName); err != nil {
			dctx.Logger.Warn("tool_acceptance_due: scan error", zap.Error(err))
			continue
		}
		tools = append(tools, t)
	}

	for _, tool := range tools {
		// A current PLAN with a criteria fence is the precondition — a tool
		// without one is Tier-2's needs_criteria concern, not an acceptance run.
		criteria, _ := loadCurrentCriteria(dctx, tool.Function)
		if criteria == "" {
			continue
		}

		// Cooldown: recently verified (either verdict counts — a fail already
		// has its improve_tool item in flight; re-running before the fix lands
		// would just duplicate the verdict).
		var recentVerdict bool
		_ = dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT EXISTS (
				SELECT 1 FROM doc_notes
				WHERE subject_type = 'tool' AND subject_key = $1
				  AND (categories ? 'acceptance-run' OR categories ? 'acceptance-fail')
				  AND created_at > NOW() - INTERVAL '7 days')
		`, tool.Function).Scan(&recentVerdict)
		if recentVerdict {
			continue
		}

		// An open run for this function is already queued/claimed.
		var openRun bool
		_ = dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT EXISTS (
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND item_type = 'acceptance_run'
				  AND spec->>'function' = $2
				  AND status NOT IN ('complete', 'cancelled', 'failed'))
		`, dctx.SiteID, tool.Function).Scan(&openRun)
		if openRun {
			continue
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "tool_acceptance_due",
			"tool":  tool.Function,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "acceptance_run",
			Severity: "low",
			Summary:  fmt.Sprintf("Tier-4 acceptance run due: %s", tool.Function),
			SpecJSON: fmt.Sprintf(
				`{"function": "%s", "component_id": "%s", "page_id": "%s", "page_name": "%s", "check": "tool_acceptance_due"}`,
				tool.Function, tool.ComponentID, tool.PageID, tool.PageName,
			),
			Priority:     90, // after builds/rerenders: test the NEW page
			HandlerAgent: "tool-acceptance-agent",
			Status:       "triaged", // no human judgment needed; triage may not sweep
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("acceptance_run:%s:%s", tool.Function, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})

		dctx.Logger.Info("tool_acceptance_due: acceptance run queued",
			zap.String("tool", tool.Function))
	}

	return result, nil
}
