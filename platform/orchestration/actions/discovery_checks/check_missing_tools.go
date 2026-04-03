// FILE: platform/orchestration/actions/discovery_checks/check_missing_tools.go
//
// Discovery check: missing_tools
//
// Triggers periodic tool evaluation for sites. Tiered cooldown:
//   - Sites with 0 tools: evaluate every 7 days (needs tools)
//   - Sites with 1+ tools: evaluate every 30 days (periodic re-evaluation
//     as content grows — a site with 10 pages might need more tools than
//     when it had 4)
//
// Tool-suggester handles the actual decision: it loads existing tools,
// avoids duplicates, and can suggest 0 new tools if the site is well-served.
//
// This check does NOT evaluate tool fit or match by tags. The old version
// had a matchToolToSite function that classified "security" and "password"
// as universal tags, causing every site to get a password checker. Tool
// selection is an LLM judgment call, not a tag lookup.

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

	// 1. Count deployed tools for this site
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

	// 2. Tiered cooldown — sites with no tools get evaluated sooner
	//    0 tools: 7 days (needs tools urgently)
	//    1+ tools: 30 days (periodic re-evaluation as content grows)
	cooldownDays := 7
	if deployedToolCount > 0 {
		cooldownDays = 30
	}

	var recentEvaluation bool
	err = dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT EXISTS (
			SELECT 1 FROM site_work_items
			WHERE site_id = $1
			  AND item_type = 'evaluate_tools'
			  AND created_at > NOW() - make_interval(days => $2)
		)
	`, dctx.SiteID, cooldownDays).Scan(&recentEvaluation)
	if err != nil {
		return nil, fmt.Errorf("missing_tools: check recent evaluation: %w", err)
	}

	if recentEvaluation {
		dctx.Logger.Info("missing_tools: recent evaluation exists, skipping",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("deployed_tools", deployedToolCount),
			zap.Int("cooldown_days", cooldownDays))
		return result, nil
	}

	// 3. No recent evaluation — create evaluate_tools item
	// tool-suggester loads existing tools and avoids duplicates
	dctx.Logger.Info("missing_tools: requesting tool evaluation",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("deployed_tools", deployedToolCount),
		zap.Int("cooldown_days", cooldownDays))

	reason := "no tools deployed, no recent evaluation"
	if deployedToolCount > 0 {
		reason = fmt.Sprintf("%d tools deployed, periodic re-evaluation (every %d days)", deployedToolCount, cooldownDays)
	}

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":          "missing_tools",
		"deployed_tools": deployedToolCount,
		"action":         "requesting_evaluation",
		"cooldown_days":  cooldownDays,
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "evaluate_tools",
		Severity:     "low",
		Summary:      "Evaluate tool needs for site",
		SpecJSON:     fmt.Sprintf(`{"check": "missing_tools", "reason": "%s", "existing_tools": %d}`, reason, deployedToolCount),
		Priority:     130,
		HandlerAgent: "tool-suggester",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("evaluate_tools:%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})

	return result, nil
}
