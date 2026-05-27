// FILE: platform/orchestration/actions/discovery_checks/check_tool_recreation_needed.go
//
// Discovery check: tool_recreation_needed
//
// Detects interactive pages (page_type tool/game) that have NO widget, and —
// when adoption captured interactive_features for them — emits
// needs_tool_recreation so tool-recreation-handler rebuilds the widget.
//
// Why this exists: adoption routes interactive pages to tool-recreation-handler
// only when the per-page feature lookup resolves (apply_adoption_plan ->
// buildPageFeatureMap -> pageFeatures[pageName]). A canonical-name desync in
// buildPageFeatureMap (fixed separately) caused tool pages to be misrouted to
// page-build-handler and rebuilt as static description pages with no widget.
// This check is the safety net: it catches any interactive page that ended up
// without a widget — from that bug, a failed recreation, or a later rebuild
// that dropped it — and re-requests recreation. It is also the BACKFILL
// mechanism for sites adopted before the routing fix: their existing
// widget-less tool pages are picked up on the next discovery run.
//
// Detection (per site): page_type IN ('tool','game'), status='active', and the
// page has no body component that is either a tool/game component or carries a
// <script> (the widget signature). The <script> arm makes detection robust to
// the exact component_level a recreated widget lands at.
//
// Source of truth for "what to recreate": the adoption findings'
// interactive_features, matched to the page by CANONICAL name (the same
// CanonicalisePage transform the router uses — the read-only twin of
// buildPageFeatureMap). Interactive-typed pages with no captured features are
// reported as findings but NOT auto-recreated: there is nothing to recreate
// from. Generating a brand-new tool is the missing_tools / tool-suggester path,
// not this check.
//
// Cooldown: skips pages with a needs_tool_recreation item in the last 7 days.
// Dedup across runs is also enforced by the runner (insertWorkItem two-strike
// rule + idx_swi_dedup on item_key).

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&ToolRecreationNeededCheck{}) }

type ToolRecreationNeededCheck struct{}

func (c *ToolRecreationNeededCheck) Name() string { return "tool_recreation_needed" }

func (c *ToolRecreationNeededCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// 1. Interactive pages on this site with NO widget.
	//    "Has a widget" = a body component that is a tool/game component OR
	//    carries a <script>. The <script> arm avoids depending on the exact
	//    component_level a recreated widget uses.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name, p.page_type
		FROM pages p
		WHERE p.site_id = $1
		  AND p.page_type IN ('tool', 'game')
		  AND p.status = 'active'
		  AND NOT EXISTS (
			SELECT 1
			FROM page_components pc
			LEFT JOIN content_components cc ON cc.id = pc.component_id
			WHERE pc.page_id = p.id
			  AND (
				(cc.component_level IN ('tool', 'game') AND cc.is_active = true)
				OR pc.rendered_html LIKE '%<script%'
			  )
		  )
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("tool_recreation_needed: query widget-less pages: %w", err)
	}

	type targetPage struct {
		ID       uuid.UUID
		Name     string
		PageType string
	}
	var targets []targetPage
	for rows.Next() {
		var idStr, name, ptype string
		if err := rows.Scan(&idStr, &name, &ptype); err != nil {
			dctx.Logger.Warn("tool_recreation_needed: scan error", zap.Error(err))
			continue
		}
		pid, perr := uuid.Parse(idStr)
		if perr != nil {
			dctx.Logger.Warn("tool_recreation_needed: bad page id", zap.String("id", idStr))
			continue
		}
		targets = append(targets, targetPage{ID: pid, Name: name, PageType: ptype})
	}
	rows.Close()

	if len(targets) == 0 {
		return result, nil
	}

	// 2. Cooldown: pages with a recent needs_tool_recreation item (7 days).
	recent := map[string]bool{}
	crows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT COALESCE(spec->>'page_name', '')
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'needs_tool_recreation'
		  AND created_at > NOW() - INTERVAL '7 days'
	`, dctx.SiteID)
	if err == nil {
		for crows.Next() {
			var pn string
			if crows.Scan(&pn) == nil && pn != "" {
				recent[pn] = true
			}
		}
		crows.Close()
	}

	// 3. Adoption interactive_features, keyed by CANONICAL page name (matches
	//    p.name above and the router's pageFeatures[pageName] lookup).
	featuresByCanon := loadAdoptionFeaturesByCanon(dctx)

	// 4. Emit needs_tool_recreation for widget-less pages we have features for.
	for _, t := range targets {
		if recent[t.Name] {
			continue
		}

		feats := featuresByCanon[t.Name]
		if len(feats) == 0 {
			// Interactive-typed page with no captured features: nothing to
			// recreate from. Surface it; generation is the tool-suggester path.
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":     "tool_recreation_needed",
				"page":      t.Name,
				"page_type": t.PageType,
				"status":    "no_captured_features_deferred_to_generation",
			})
			continue
		}

		// Spec mirrors apply_adoption_plan's needs_tool_recreation branch so
		// tool-recreation-handler receives identical input.
		spec := map[string]interface{}{
			"page_name":            t.Name,
			"page_type":            t.PageType,
			"source":               "discovery",
			"mode":                 "recreate", // load_existing_content checks for this
			"interactive_features": feats,
		}
		specJSON, _ := json.Marshal(spec)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "tool_recreation_needed",
			"page":      t.Name,
			"page_type": t.PageType,
			"features":  len(feats),
			"action":    "requesting_recreation",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       t.ID,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "needs_tool_recreation",
			Severity:     "medium",
			Summary:      fmt.Sprintf("Recreate missing widget for %s", t.Name),
			SpecJSON:     string(specJSON),
			Priority:     20,
			HandlerAgent: "tool-recreation-handler",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			// Distinct from adoption's "needs_page:<name>" so this never
			// collides with a completed adoption content-page item for the same
			// page, while still deduping this check's own re-runs.
			ItemKey: fmt.Sprintf("needs_tool_recreation:%s", t.Name),
			BatchID: dctx.BatchID,
		})
	}

	return result, nil
}

// loadAdoptionFeaturesByCanon reads the latest adoption findings and returns
// interactive_features grouped by CANONICAL page name, using the same
// CanonicalisePage transform the adoption router uses. Read-only twin of
// actions.buildPageFeatureMap (which lives in a different package, so the logic
// is mirrored rather than shared — both reuse datahelpers.CanonicalisePage).
func loadAdoptionFeaturesByCanon(dctx DiscoveryCheckContext) map[string][]map[string]interface{} {
	out := map[string][]map[string]interface{}{}

	var findingsJSON []byte
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT findings FROM research_results
		WHERE site_id = $1 AND result_type = 'adoption_crawl'
		ORDER BY created_at DESC
		LIMIT 1
	`, dctx.SiteID).Scan(&findingsJSON)
	if err != nil || findingsJSON == nil {
		return out
	}

	var plan map[string]interface{}
	if json.Unmarshal(findingsJSON, &plan) != nil {
		return out
	}

	// raw page name -> role, so feature refs can be canonicalised with the
	// correct role (the prefix CanonicalisePage adds depends on the role).
	roleByRaw := map[string]string{}
	if pages, ok := plan["pages"].([]interface{}); ok {
		for _, pr := range pages {
			pm, ok := pr.(map[string]interface{})
			if !ok {
				continue
			}
			rawName, _ := pm["name"].(string)
			rawType, _ := pm["page_type"].(string)
			if rawName != "" {
				roleByRaw[strings.ToLower(strings.TrimSpace(rawName))] = rawType
			}
		}
	}

	feats, ok := plan["interactive_features"].([]interface{})
	if !ok {
		return out
	}
	for _, f := range feats {
		fm, ok := f.(map[string]interface{})
		if !ok {
			continue
		}
		rawPage, _ := fm["page"].(string)
		if rawPage == "" {
			continue
		}
		role := roleByRaw[strings.ToLower(strings.TrimSpace(rawPage))]
		cname, _, _ := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
			Role: role,
			Slug: rawPage,
		})
		if cname == "" {
			cname = strings.ToLower(strings.TrimSpace(rawPage))
		}
		out[cname] = append(out[cname], fm)
	}
	return out
}
