// FILE: platform/orchestration/actions/discovery_checks/check_missing_tools.go
//
// Discovery check: missing_tools
//
// Triggers periodic tool evaluation for sites. Tiered cooldown:
//   - Sites with 0 tools: evaluate every 7 days (needs tools)
//   - Sites with 1+ tools: evaluate every 30 days (periodic re-evaluation
//     as content grows — a site with 10 pages might need more tools than
//     when it had 4)
//   - Sites behind their configured content-to-tools ratio: 7 days, however
//     many tools they already have.
//
// The content-to-tools ratio (site_specs aspect growth_config, key
// `content_tools_ratio`) is how many PUBLISHED articles/guides justify one
// tool — 6 means a site with 18 guides should have about 3. It exists because
// the cooldown alone asks "has it been a while?", which cannot distinguish a
// site that published thirty guides last fortnight from a dormant one; the
// third bullet is the only tier that responds to a site actually growing.
//
// The key is OPT-IN and defaults to OFF: absent or <= 0 leaves the original
// two-tier behaviour byte-for-byte unchanged, so no existing site is affected
// until someone writes it. Only DEPLOYED pages count — sizing a tool budget on
// planned pages buys tools for content that may never ship.
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
	"encoding/json"
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

	// 2. Content-to-tools ratio (opt-in, per site).
	//
	// The cooldown below re-evaluates on a CLOCK, which answers "has it been a while?"
	// and not "has this site outgrown its tools?" — a site that publishes thirty guides
	// in a fortnight looks identical to a dormant one. `content_tools_ratio` in the
	// growth_config spec says how many published articles/guides justify one tool
	// (e.g. 6 => a site with 18 guides should have ~3). When the site is behind that
	// ratio the 30-day cooldown is cut to the same 7 days a tool-less site gets, so the
	// suggester is asked sooner.
	//
	// ABSENT OR <=0 MEANS OFF, and that is the default: no site changes behaviour
	// until someone writes the key. Read directly rather than via
	// actions.loadGrowthConfig — discovery_checks is imported BY actions, so importing
	// it back would be a cycle, and only one field is needed here.
	var contentToolsRatio int
	if err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT COALESCE((data->>'content_tools_ratio')::int, 0)
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'growth_config' AND is_current = true
	`, dctx.SiteID).Scan(&contentToolsRatio); err != nil {
		contentToolsRatio = 0 // no spec row, or an unparseable value: stay off
	}

	// Published articles/guides. Deployed only: a planned page is an intention, and
	// sizing a tool budget on intentions is how a site ends up with tools for content
	// that never shipped.
	var articleCount int
	if contentToolsRatio > 0 {
		if err := dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT COUNT(*)
			FROM pages
			WHERE site_id = $1
			  AND page_type IN ('blog-post', 'content')
			  AND status = 'active'
			  AND deployed_at IS NOT NULL
		`, dctx.SiteID).Scan(&articleCount); err != nil {
			return nil, fmt.Errorf("missing_tools: count published articles: %w", err)
		}
	}

	toolsExpected := 0
	if contentToolsRatio > 0 {
		toolsExpected = articleCount / contentToolsRatio
	}
	behindRatio := toolsExpected > deployedToolCount

	// 3. Tiered cooldown — sites with no tools get evaluated sooner
	//    0 tools: 7 days (needs tools urgently)
	//    1+ tools: 30 days (periodic re-evaluation as content grows)
	//    behind the configured ratio: 7 days, however many tools it already has
	cooldownDays := 7
	if deployedToolCount > 0 && !behindRatio {
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

	// 4. No recent evaluation — create evaluate_tools item
	// tool-suggester loads existing tools and avoids duplicates
	dctx.Logger.Info("missing_tools: requesting tool evaluation",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("deployed_tools", deployedToolCount),
		zap.Int("cooldown_days", cooldownDays))

	reason := "no tools deployed, no recent evaluation"
	if deployedToolCount > 0 {
		reason = fmt.Sprintf("%d tools deployed, periodic re-evaluation (every %d days)", deployedToolCount, cooldownDays)
	}
	if behindRatio {
		reason = fmt.Sprintf(
			"%d published articles/guides at 1 tool per %d implies ~%d tools; %d deployed",
			articleCount, contentToolsRatio, toolsExpected, deployedToolCount)
	}

	finding := map[string]interface{}{
		"check":          "missing_tools",
		"deployed_tools": deployedToolCount,
		"action":         "requesting_evaluation",
		"cooldown_days":  cooldownDays,
	}
	if contentToolsRatio > 0 {
		finding["content_tools_ratio"] = contentToolsRatio
		finding["published_articles"] = articleCount
		finding["tools_expected"] = toolsExpected
		finding["behind_ratio"] = behindRatio
	}
	result.Findings = append(result.Findings, finding)

	summary := "Evaluate tool needs for site"
	spec := map[string]interface{}{
		"check":          "missing_tools",
		"reason":         reason,
		"existing_tools": deployedToolCount,
	}
	if behindRatio {
		summary = fmt.Sprintf("Site is behind its content-to-tools ratio (%d articles, %d tools, target 1 per %d)",
			articleCount, deployedToolCount, contentToolsRatio)
		spec["content_tools_ratio"] = contentToolsRatio
		spec["published_articles"] = articleCount
		spec["tools_expected"] = toolsExpected
		spec["tools_short"] = toolsExpected - deployedToolCount
	}
	// Marshal rather than Sprintf a JSON literal: `reason` is assembled from config
	// values, and a quote in one of them would have produced a malformed spec that
	// only shows up when something tries to read it.
	specBytes, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("missing_tools: marshal spec: %w", err)
	}

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:       dctx.SiteID,
		Source:       "discovery",
		Pipeline:     "build",
		ItemType:     "evaluate_tools",
		Severity:     "low",
		Summary:      summary,
		SpecJSON:     string(specBytes),
		Priority:     130,
		HandlerAgent: "tool-suggester",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("evaluate_tools:%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})

	return result, nil
}
