// FILE: platform/orchestration/actions/discovery_checks/check_unrendered_page_imagery.go
//
// Discovery check: a page-scoped content-hero asset exists — generated, active,
// deployed under a deterministic path — and the page it was made for renders it
// nowhere. bugs_open/114's remaining gap: every other stage of that pipeline now
// converges (generation → store → deploy → event-filed card derivation, 193/193
// measured 2026-09-02), and NOTHING detected the one state the bug is named
// after. check_undeployed_assets cannot: its evidence is purpose-prefix and
// site-wide, so one wired sibling vouches for every asset of that purpose, and
// its remedy is a deploy, which re-commits bytes and cannot wire a page — the
// recurrence brake then parks the refiled items (1,651 born-`unresolved` when
// this was written).
//
// THREE STATES, because the remedies differ and lumping them is how the wrong
// medicine got prescribed for a year:
//
//   unwired       — an image-capable, non-fragment component exists on the page.
//                   The asset is deliverable; delivery is bugs_open/412's fix
//                   candidate 1 (deploy-time wiring). Actionable.
//   fragment_slot — the only image-capable component row stores an interactive
//                   fragment (bugs_open/357 / RFC_046: the row declares `hero`
//                   while holding the tool shell). Imagery cannot land until the
//                   row's identity is repaired; NOT actionable from imagery.
//   no_image_slot — no component template on the page carries an image branch.
//                   A composition gap (bugs_open/412 §7). The asset still feeds
//                   listing-card derivation, so generation was not waste — which
//                   is also why this check deliberately does NOT gate the
//                   generator (decision recorded in bugfix_114_imagery_wiring
//                   PLAN, 2026-09-02 revision).
//
// ONE ROLLUP ITEM PER (site, state), NOT one per page. Per-page flags would land
// hundreds of rows in a review queue with no working surface (bugs_open/033);
// the remedies above are per-mechanism, not per-page. The rollup is flag-only
// (`needs_human_review`, no handler), carries the census count with the date it
// was counted (owner ruling 2026-08-22: a count without its date reads as
// current for ever), and up to unrenderedImageryMaxExamples example pages. The
// dedup key keeps one open rollup per state; a standing rollup's spec is the
// census AT FILING — the re-runnable query lives in
// docs/agent_docs/docs024_key_docs_latest/bugfix_114_imagery_wiring/RUNBOOK_imagery_wiring.md.
//
// RETRACTION (RFC_010, narrow branch only): when a sweep completes and a state
// has NO members, the check resolves that state's rollup by ItemKey. That is a
// positive observation — the population was computed and found empty — never an
// inference from an errored run (the runner already skips Resolved on error).
//
// THE KEY CONVENTION IS MIRRORED IN SQL, DELIBERATELY, AND PINNED. The sweep
// matches assets with `'content_hero_' || replace(p.name, '-', '_')` — the SQL
// spelling of imageryplan.ContentHeroKey. A set-based sweep cannot call the Go
// helper per row; TestContentHeroKeySQLMirrorStaysTrue pins the two spellings
// together so they cannot drift apart silently.
//
// THE SERVED PATH COMES FROM storage.DeployedWebPath — the deployer's own
// derivation. Re-deriving it here (or in SQL) is exactly the two-derivations
// defect that poisoned sites.content_data for a month (bugs_open/114, GAP 1,
// register IMG-072). One rule, one writer.
//
// ⚠ THE FUNCTION'S OWN LANDMINE, READ BEFORE TRUSTING IT (council round 1's
// gating objection — rightly: it was cited here unread). LANDMINES.md carries
// an entry keyed on `storage.DeployedWebPath` ("silently WRONG for og_card,
// right for favicon by coincidence"). Its status header: FIXED AND LIVE on
// chassis v1.0.1229, pod-verified — "THE TRAP BELOW IS HISTORY"
// (bugs_closed/168, IMG-067). Two further reasons it cannot bite this check
// even on an older binary: the historical defect covered ONLY the two
// brand-head purposes ("It is the correct answer for every purpose except the
// two brand-head artefacts"), and this check's population is content-hero
// keys, never brand-head. The entry's surviving guidance — brand-head
// references live in site_components, everything else compares against
// page_components — is exactly the comparison surface used below.
//
// WHY A FOURTH IMAGE-STATE CHECK DOES NOT OVERLAP THE TWO GENERATION
// NEIGHBOURS (the reuse seat's question, answered by premise rather than by
// census): check_placeholder_image_in_use fires when the site's canonical hero
// asset is ABSENT or a placeholder, and files generation;
// check_unfulfilled_image_prompt fires when a site_specs prompt has NO asset,
// and files generation. Both trigger on an asset that does not exist. This
// check triggers ONLY on an asset that DOES exist (active, under the page's
// own key) — the existence probe that admits a page here is the same fact
// that silences them, so the populations are disjoint by construction.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

func init() { Register(&UnrenderedPageImageryCheck{}) }

type UnrenderedPageImageryCheck struct{}

func (c *UnrenderedPageImageryCheck) Name() string { return "unrendered_page_imagery" }

const unrenderedPageImageryItemType = "unrendered_page_imagery"

// unrenderedImageryMaxCandidates bounds the per-site sweep. It is a sanity
// ceiling, not a paging cursor: the largest live population when written was 66
// candidate pages on one site (webdesign.co.uk, measured 2026-09-02).
const unrenderedImageryMaxCandidates = 200

// unrenderedImageryMaxExamples bounds the example list carried in each rollup's
// spec. The COUNT is always the full census; only the examples are truncated.
const unrenderedImageryMaxExamples = 12

// InteractiveStructuralMarkers mirrors interactiveStructuralMarkers in
// actions/save_page_sections_action.go — the markers bugs_open/357's machinery
// uses to recognise a stored interactive fragment. Exported so the actions
// package (which already imports this one; the reverse import would cycle) can
// pin the two lists together: TestInteractiveStructuralMarkersLockstep fails if
// they drift. Single-sourcing means moving that private list down the
// dependency graph into this package — offered to the 357 lane rather than done
// under them while their lane is mid-flight.
var InteractiveStructuralMarkers = []string{"<canvas", "game-container", "tool-page", "data-tool"}

// unrenderedImageryStates, in the order rollups are filed. Order is stable so
// tests and readers see a deterministic sequence.
var unrenderedImageryStates = []string{"unwired", "fragment_slot", "no_image_slot"}

// unrenderedImageryPage is one census member.
type unrenderedImageryPage struct {
	PageName string `json:"page"`
	AssetKey string `json:"asset_key"`
	WebPath  string `json:"web_path"`
}

// classifyUnrenderedImagery is the pure per-page decision. "" means fulfilled —
// the page renders its asset and no state applies.
func classifyUnrenderedImagery(referenced, capable, capableNonFragment bool) string {
	switch {
	case referenced:
		return ""
	case capableNonFragment:
		return "unwired"
	case capable:
		return "fragment_slot"
	default:
		return "no_image_slot"
	}
}

// unrenderedImageryRemedies routes each state to the mechanism that owns it.
// These strings travel in the item spec so a reader of the ITEM knows where the
// remedy lives without re-deriving the split.
var unrenderedImageryRemedies = map[string]string{
	"unwired":       "deliverable but undelivered — delivery is bugs_open/412 fix candidate 1 (deploy-time wiring); where content_data already holds the path, a reason=template_changed rerender delivers today",
	"fragment_slot": "the image-capable component row stores an interactive fragment (bugs_open/357 / RFC_046) — imagery cannot land until the row identity is repaired; not actionable from the imagery side",
	"no_image_slot": "no component template on this page carries an image branch. WHY the planner composed it that way is an open question, not a finding (bugs_open/412 may hold the answer — its s1 says a page cannot gain a hero without a full copy rebuild); the asset still feeds listing-card derivation, so do not read this as pure waste",
}

var unrenderedImagerySeverity = map[string]string{
	"unwired":       "medium",
	"fragment_slot": "low",
	"no_image_slot": "low",
}

func (c *UnrenderedPageImageryCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	// Population: pages holding an ACTIVE asset under the resolver's own key
	// convention (SQL mirror of imageryplan.ContentHeroKey — pinned by test),
	// restricted to pages that have at least one non-removed component: a page
	// that has never been composed has nothing to render anything with, and
	// flagging it would be a plan state, not a defect. Lifecycle: the platform
	// must still WANT the page served (PostureArmed) — a retired page's
	// unrendered asset is not actionable and would misdirect the census.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT DISTINCT ON (p.id) p.id::text, p.name, a.asset_key, COALESCE(a.purpose, 'content_hero')
		  FROM pages p
		  JOIN assets a ON a.site_id = p.site_id
		               AND a.status = 'active'
		               AND a.asset_key = 'content_hero_' || replace(p.name, '-', '_')
		 WHERE p.site_id = $1
		   AND `+datahelpers.PageWantedLivePredicateFor("p")+`
		   AND EXISTS (SELECT 1 FROM page_components pc0
		                WHERE pc0.page_id = p.id
		                  AND `+datahelpers.NotRemoved("pc0")+`)
		 ORDER BY p.id
		 LIMIT `+fmt.Sprint(unrenderedImageryMaxCandidates), dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("unrendered_page_imagery: population sweep failed: %w", err)
	}
	defer rows.Close()

	type candidate struct {
		pageID   string
		pageName string
		assetKey string
		purpose  string
	}
	var candidates []candidate
	for rows.Next() {
		var cand candidate
		if err := rows.Scan(&cand.pageID, &cand.pageName, &cand.assetKey, &cand.purpose); err != nil {
			return nil, fmt.Errorf("unrendered_page_imagery: population scan failed: %w", err)
		}
		candidates = append(candidates, cand)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unrendered_page_imagery: population iter failed: %w", err)
	}

	byState := map[string][]unrenderedImageryPage{}
	for _, cand := range candidates {
		webPath := storage.DeployedWebPath(cand.assetKey, cand.purpose)
		if webPath == "" {
			// DeployedWebPath contract: "" is unresolvable — skip loudly, never guess.
			dctx.Logger.Warn("unrendered_page_imagery: no deployable web path for asset — skipped",
				zap.String("asset_key", cand.assetKey), zap.String("purpose", cand.purpose))
			continue
		}

		var referenced, capable, capableNonFragment bool
		err := dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT
			  EXISTS (SELECT 1 FROM page_components rc
			           WHERE rc.page_id = $1::uuid
			             AND `+datahelpers.NotRemoved("rc")+`
			             AND rc.rendered_html LIKE '%' || $2 || '%'),
			  EXISTS (SELECT 1 FROM page_components pc
			            JOIN content_components cc ON cc.id = pc.component_id
			           WHERE pc.page_id = $1::uuid
			             AND `+datahelpers.NotRemoved("pc")+`
			             AND (cc.html_template LIKE '%hero_url%' OR cc.html_template LIKE '%background_image%')),
			  EXISTS (SELECT 1 FROM page_components pc
			            JOIN content_components cc ON cc.id = pc.component_id
			           WHERE pc.page_id = $1::uuid
			             AND `+datahelpers.NotRemoved("pc")+`
			             AND (cc.html_template LIKE '%hero_url%' OR cc.html_template LIKE '%background_image%')
			             AND NOT (`+fragmentMarkerPredicateSQL("pc.rendered_html")+`))
		`, cand.pageID, webPath).Scan(&referenced, &capable, &capableNonFragment)
		if err != nil {
			return nil, fmt.Errorf("unrendered_page_imagery: classify %q failed: %w", cand.pageName, err)
		}

		state := classifyUnrenderedImagery(referenced, capable, capableNonFragment)
		if state == "" {
			continue
		}
		byState[state] = append(byState[state], unrenderedImageryPage{
			PageName: cand.pageName,
			AssetKey: cand.assetKey,
			WebPath:  webPath,
		})
	}

	return buildUnrenderedImageryResult(dctx, byState, time.Now().UTC().Format("2006-01-02"))
}

// fragmentMarkerPredicateSQL builds the OR-chain over
// InteractiveStructuralMarkers for the given column expression. Built from the
// slice — not hand-spelled — so the lockstep test's subject is also the SQL's.
func fragmentMarkerPredicateSQL(column string) string {
	parts := make([]string, len(InteractiveStructuralMarkers))
	for i, m := range InteractiveStructuralMarkers {
		parts[i] = column + " LIKE '%" + m + "%'"
	}
	return strings.Join(parts, " OR ")
}

// buildUnrenderedImageryResult turns the census into rollup work items and
// retractions. Pure apart from reading dctx identity fields, so tests exercise
// the item shapes without a DB.
func buildUnrenderedImageryResult(dctx DiscoveryCheckContext, byState map[string][]unrenderedImageryPage, measuredAt string) (*CheckResult, error) {
	result := &CheckResult{}
	counts := map[string]int{}

	for _, state := range unrenderedImageryStates {
		members := byState[state]
		counts[state] = len(members)

		if len(members) == 0 {
			// Positive observation: the population was computed and this state
			// has no members. Retract the standing rollup, narrowly.
			result.Resolved = append(result.Resolved, ResolvedFinding{
				// The literal spelling here and below is what the verifier
				// coverage sensor scans for; the const carries the same string
				// (pinned by the shapes test) for key construction.
				ItemType: "unrendered_page_imagery",
				ItemKey:  unrenderedPageImageryItemType + ":" + state,
				Reason: fmt.Sprintf("census of %s found no page in state %s (active content-hero asset, page renders it nowhere)",
					measuredAt, state),
			})
			continue
		}

		examples := members
		if len(examples) > unrenderedImageryMaxExamples {
			examples = examples[:unrenderedImageryMaxExamples]
		}
		specJSON, err := json.Marshal(map[string]interface{}{
			"check":       "unrendered_page_imagery",
			"state":       state,
			"count":       len(members),
			"measured_at": measuredAt,
			"examples":    examples,
			"remedy":      unrenderedImageryRemedies[state],
			"census_query": "docs/agent_docs/docs024_key_docs_latest/bugfix_114_imagery_wiring/" +
				"RUNBOOK_imagery_wiring.md — re-run for current membership; the count above is as of measured_at",
		})
		if err != nil {
			return nil, fmt.Errorf("unrendered_page_imagery: spec marshal for %s failed: %w", state, err)
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "unrendered_page_imagery",
			Severity: unrenderedImagerySeverity[state],
			Summary: fmt.Sprintf("%d page(s) hold a deployed content-hero the page never renders (state %s, counted %s)",
				len(members), state, measuredAt),
			SpecJSON:     string(specJSON),
			Priority:     40,
			HandlerAgent: "", // flag-only: the dispatch loop must never claim these
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      unrenderedPageImageryItemType + ":" + state,
			BatchID:      dctx.BatchID,
		})
	}

	if counts["unwired"]+counts["fragment_slot"]+counts["no_image_slot"] > 0 {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":       "unrendered_page_imagery",
			"measured_at": measuredAt,
			"counts":      counts,
		})
	}
	return result, nil
}
