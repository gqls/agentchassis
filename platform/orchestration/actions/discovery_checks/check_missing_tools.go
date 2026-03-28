// FILE: platform/orchestration/actions/discovery_checks/check_missing_tools.go
//
// Discovery check: missing_tools
//
// Structural check — asks two questions:
//   1. Does this site have any tools deployed?
//   2. Has a tool evaluation happened in the last 7 days?
//
// If both are no, creates an evaluate_tools work item for tool-suggester.
// The LLM in tool-suggester makes the actual decision about which tools
// are appropriate for the site's industry and audience.
//
// This check does NOT do tag matching or direct tool deployment.
// The old version had a matchToolToSite function that classified
// "security" and "password" as universal tags, causing every site
// to get a password checker. Tool selection is an LLM judgment call,
// not a tag lookup.

package discovery_checks

import (
	"fmt"

	"go.uber.org/zap"
)

func init() { Register(&MissingToolsCheck{}) }

type MissingToolsCheck struct{}

func (c *MissingToolsCheck) Name() string { return "missing_tools" }

func (c *MissingToolsCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// 1. Does this site have any tools deployed?
	var deployedToolCount int
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COUNT(*)
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND cc.component_level = 'tool'
		  AND cc.is_active = true
	`, dctx.SiteID).Scan(&deployedToolCount)
	if err != nil {
		return nil, fmt.Errorf("missing_tools: count deployed tools: %w", err)
	}

	if deployedToolCount > 0 {
		dctx.Logger.Info("missing_tools: site has tools, skipping",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("deployed_tools", deployedToolCount))
		return result, nil
	}

	// 2. Has a tool evaluation been done recently?
	var recentEvaluation bool
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT EXISTS (
			SELECT 1 FROM site_work_items
			WHERE site_id = $1
			  AND item_type = 'evaluate_tools'
			  AND created_at > NOW() - INTERVAL '7 days'
		)
	`, dctx.SiteID).Scan(&recentEvaluation)
	if err != nil {
		return nil, fmt.Errorf("missing_tools: check recent evaluation: %w", err)
	}

	if recentEvaluation {
		dctx.Logger.Info("missing_tools: recent evaluation exists, skipping",
			zap.String("site_id", dctx.SiteID.String()))
		return result, nil
	}

	// 3. No tools and no recent evaluation — create evaluate_tools item
	// tool-suggester will use LLM judgment to decide what's appropriate
	dctx.Logger.Info("missing_tools: requesting tool evaluation",
		zap.String("site_id", dctx.SiteID.String()))

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":          "missing_tools",
		"deployed_tools": 0,
		"action":         "requesting_evaluation",
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "evaluate_tools",
		Severity:     "low",
		Summary:      "Evaluate tool needs for site",
		SpecJSON:     `{"check": "missing_tools", "reason": "no tools deployed, no recent evaluation"}`,
		Priority:     130,
		HandlerAgent: "tool-suggester",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("evaluate_tools:%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})

	return result, nil
}
