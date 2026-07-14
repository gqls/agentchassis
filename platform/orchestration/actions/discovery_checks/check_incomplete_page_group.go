// FILE: platform/orchestration/actions/discovery_checks/check_incomplete_page_group.go
//
// Detects PARTIALLY-BUILT page groups: the current site plan promises a set
// of sibling pages (same parent_section, or the implicit "interactive" group
// of tool/game roles) and some siblings are deployed while others were never
// built. A half-built group is worse than an unbuilt one — the deployed
// siblings link to and promise the missing ones (the vonc case: two tools
// deployed, copy selling a third that has no page anywhere).
//
// Group rule: a group is INCOMPLETE when >= 1 sibling is deployed AND >= 1
// sibling has no pages row or is not deployed. Classification mirrors
// reconcile_site_plan's decideEmit (reimplemented here — package actions
// imports this package, so importing it back would cycle):
//   missing   — no pages row
//   not_built — row exists, build_status != deployed
//   stale     — deployed but built_from_plan_version missing/older
// Stale pages count as REALISED (the destination exists for a visitor) and
// are surfaced as findings only: reconcile_site_plan owns stale rebuilds, and
// emitting needs_page here too would double-queue every page on a plan bump.
//
// Routing per gap:
//   - ordinary content roles -> needs_page / page-build-handler, the same
//     item_key ("needs_page:<name>") and spec shape reconcile_site_plan
//     emits, so the two sources co-dedup via (site_id, item_key).
//   - tool/game roles -> needs_human_review, NO handler. Known gap TP-004:
//     the generic page builder would produce a widget-less prose page where
//     an interactive tool belongs; building a tool is a tool-generator task
//     a human must dispatch.
//
// Registration: automatic via init() -> Register(&IncompletePageGroupCheck{}).
// Enable by adding "incomplete_page_group" to a discovery agent's checks array.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func init() { Register(&IncompletePageGroupCheck{}) }

type IncompletePageGroupCheck struct{}

func (c *IncompletePageGroupCheck) Name() string { return "incomplete_page_group" }

type groupPlanPage struct {
	Name          string
	Role          string
	Title         string
	ParentSection string
}

func (c *IncompletePageGroupCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Current plan — a site without one has nothing promised, nothing to check.
	var planID uuid.UUID
	err := dctx.DB.QueryRowContext(dctx.Ctx, `
		SELECT id FROM site_plans WHERE site_id = $1 AND is_current = true
	`, dctx.SiteID).Scan(&planID)
	if errors.Is(err, sql.ErrNoRows) {
		dctx.Logger.Info("incomplete_page_group: no current site plan, skipping",
			zap.String("site_id", dctx.SiteID.String()))
		return &CheckResult{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("incomplete_page_group plan lookup failed: %w", err)
	}

	planPages, err := loadGroupPlanPages(dctx, planID)
	if err != nil {
		return nil, err
	}

	// Realised state per plan page name.
	status, err := loadRealisedStatus(dctx, planID)
	if err != nil {
		return nil, err
	}

	// Build groups: explicit parent_section groups + the implicit interactive
	// group (tool/game roles). A page may appear in both; gaps are emitted
	// once (dedup by name below).
	groups := map[string][]groupPlanPage{}
	for _, p := range planPages {
		if p.ParentSection != "" {
			groups["section:"+p.ParentSection] = append(groups["section:"+p.ParentSection], p)
		}
		if p.Role == "tool" || p.Role == "game" {
			groups["interactive"] = append(groups["interactive"], p)
		}
	}

	result := &CheckResult{}
	emittedGap := map[string]bool{}

	groupNames := make([]string, 0, len(groups))
	for g := range groups {
		groupNames = append(groupNames, g)
	}
	sort.Strings(groupNames)

	for _, groupName := range groupNames {
		members := groups[groupName]
		if len(members) < 2 {
			continue // a group of one cannot be "partially" built
		}

		deployed := 0
		var gaps []groupPlanPage
		var staleNames []string
		for _, m := range members {
			switch status[m.Name] {
			case "deployed":
				deployed++
			case "stale":
				deployed++ // realised for a visitor; reconcile owns the rebuild
				staleNames = append(staleNames, m.Name)
			default: // "missing" or "not_built"
				gaps = append(gaps, m)
			}
		}
		if deployed == 0 || len(gaps) == 0 {
			continue // fully unbuilt or fully built — not this check's problem
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":    "incomplete_page_group",
			"group":    groupName,
			"members":  len(members),
			"deployed": deployed,
			"gaps":     pageNames(gaps),
			"stale":    staleNames,
		})

		for _, gap := range gaps {
			if emittedGap[gap.Name] {
				continue
			}
			emittedGap[gap.Name] = true

			gapState := status[gap.Name]
			if gapState == "" {
				gapState = "missing"
			}

			if gap.Role == "tool" || gap.Role == "game" {
				// TP-004: generic builder makes prose pages, not tools.
				specJSON, _ := json.Marshal(map[string]interface{}{
					"check":     "incomplete_page_group",
					"group":     groupName,
					"page_name": gap.Name,
					"page_role": gap.Role,
					"plan_id":   planID.String(),
					"state":     gapState,
					"fix": "Planned interactive page missing while siblings are live. " +
						"Do NOT route to the generic page builder (it produces a widget-less " +
						"prose page). Build via tool-generator/create_tool_component, or " +
						"remove the page from the plan.",
				})
				result.WorkItems = append(result.WorkItems, WorkItemSpec{
					SiteID:    dctx.SiteID,
					Source:    "discovery",
					Pipeline:  "build",
					ItemType:  "incomplete_page_group",
					Severity:  "high",
					Summary:   fmt.Sprintf("Planned %s page %q is %s while sibling pages are deployed (%s)", gap.Role, gap.Name, gapState, groupName),
					SpecJSON:  string(specJSON),
					Priority:  30,
					Status:    "needs_human_review",
					CreatedBy: dctx.AgentType,
					ItemKey:   fmt.Sprintf("incomplete_group_tool:%s", gap.Name),
					BatchID:   dctx.BatchID,
				})
				continue
			}

			// Ordinary content page: same key + spec shape as reconcile's
			// needs_page items, so the two sources co-dedup.
			specJSON, _ := json.Marshal(map[string]interface{}{
				"page_name": gap.Name,
				"page_role": gap.Role,
				"plan_id":   planID.String(),
				"reason":    "incomplete_page_group",
			})
			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:       dctx.SiteID,
				Source:       "discovery",
				Pipeline:     "build",
				ItemType:     "needs_page",
				Severity:     "medium",
				Summary:      fmt.Sprintf("Build %s page (%s sibling gap in %s)", gap.Name, gapState, groupName),
				SpecJSON:     string(specJSON),
				Priority:     50,
				HandlerAgent: "page-build-handler",
				Status:       "detected",
				CreatedBy:    dctx.AgentType,
				ItemKey:      "needs_page:" + gap.Name,
				BatchID:      dctx.BatchID,
			})
		}
	}

	if len(result.WorkItems) > 0 {
		dctx.Logger.Warn("incomplete_page_group: found partially-built page groups",
			zap.Int("gaps", len(result.WorkItems)),
			zap.String("site_id", dctx.SiteID.String()))
	}

	return result, nil
}

func pageNames(pages []groupPlanPage) []string {
	names := make([]string, len(pages))
	for i, p := range pages {
		names[i] = p.Name
	}
	return names
}

func loadGroupPlanPages(dctx DiscoveryCheckContext, planID uuid.UUID) ([]groupPlanPage, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT name, role, COALESCE(title, ''), COALESCE(parent_section, '')
		FROM site_plan_pages
		WHERE plan_id = $1
	`, planID)
	if err != nil {
		return nil, fmt.Errorf("incomplete_page_group plan pages query failed: %w", err)
	}
	defer rows.Close()

	var out []groupPlanPage
	for rows.Next() {
		var p groupPlanPage
		if err := rows.Scan(&p.Name, &p.Role, &p.Title, &p.ParentSection); err != nil {
			continue
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// loadRealisedStatus classifies every realised page: deployed | not_built |
// stale (name absent => missing). Mirrors reconcile_site_plan's decideEmit.
func loadRealisedStatus(dctx DiscoveryCheckContext, planID uuid.UUID) (map[string]string, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT name, COALESCE(build_status, ''), built_from_plan_version
		FROM pages
		WHERE site_id = $1 AND status NOT IN ('deleted', 'archived')
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("incomplete_page_group pages query failed: %w", err)
	}
	defer rows.Close()

	out := map[string]string{}
	for rows.Next() {
		var name, buildStatus string
		var builtFrom uuid.NullUUID
		if err := rows.Scan(&name, &buildStatus, &builtFrom); err != nil {
			continue
		}
		switch {
		case buildStatus != "deployed":
			out[name] = "not_built"
		case !builtFrom.Valid || builtFrom.UUID != planID:
			out[name] = "stale"
		default:
			out[name] = "deployed"
		}
	}
	return out, rows.Err()
}
