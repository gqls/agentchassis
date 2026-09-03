// FILE: platform/orchestration/actions/tool_item_sources.go
//
// bugs_open/450, the SUPPLY half — the tool-page item-source gate: ONE answer to
// "is this planned page a tool page, and does a tool that could fill it actually
// exist for THIS site?", asked at plan time by ValidateSitePlanAction so a tool
// page whose tool does not exist is held out of the plan and filed as a
// capability_gap instead of becoming a page row that nothing will ever fill.
//
// THE DEFECT THIS CUTS OFF AT SOURCE. The planner names tool pages with generic
// sections (hero-tool, generic-text-block) before their tools exist, because
// tools arrive from the design rotation hours-to-days later and under names the
// planner never saw — seotools.co.uk: 0 of 7 planned names matched what
// tool-deployer eventually built. The page row is minted at plan time, links
// point at it, and generic producers fill it with prose about a tool that is not
// there. The guard in owned_page_guard.go (PBP-053) stops the FILLING; this
// stops the SUPPLY, and the two are independent — either alone is an improvement.
//
// WHY THIS IS A SIBLING OF listing_item_sources.go AND NOT A WIDENING OF IT.
// The bugs_open/444 session, which built that gate, wrote the request into
// bugs_open/450 itself: arm a tool case behind its OWN optional key rather than
// widening what enforce_listing_sources means, so a tool-arm misfire is
// switchable without losing the live listing gate, and because a tool page is
// not semantically a listing page. This file therefore reuses that gate's
// FRAME — the preserve-guard rule, the fail-open posture, the capability_gap
// shape through the shared work-item writer, the durable drop findings — and
// shares its plan-page primitives (planPageView, pageViewFromMap,
// droppedSectionName, LogActionFindings), while keeping its own resolver
// vocabulary. listing_item_sources.go is not edited at all.
//
// THE ORDER-SAFETY QUESTION THAT GATED THIS, AND ITS ANSWER. 444's CONTRIB was
// right that the property making its own gate safe does NOT transfer: a listing
// hub's children are IN the plan, so post-plan builders cannot false-positive,
// whereas a tool page's producer arrives from OUTSIDE the plan, later, by
// design. "Hold if no tool exists" would therefore hold every tool page on every
// fresh build — and whether that STARVES the rotation or merely removes dead
// rows was the open question. Answered at the rows 2026-09-03 (bugs_open/450 §7):
// tool-deployer CREATES ITS OWN page rows, its names are disjoint from the
// planner's, and nothing reads planned tool pages to decide what to build. The
// held page was never the producer's input, so holding it starves nothing. That
// measurement is this gate's licence to exist; if it is ever falsified, this
// gate is the thing to switch off.
//
// POLICY: fail OPEN on infrastructure, CLOSED on a positive answer. Unreadable
// DB, unparseable site id, or a failed tool census → the gate stands down
// entirely. But an EMPTY tool set is not ambiguity: it is a positive answer
// ("this site owns no tools"), and pages are held. Unlike the listing family
// there is no unknown-shape middle ground — a tool page's item source is one
// positively-understood thing, a live component_level='tool' row.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// toolProducerSlug is the capability_gap grouping key for "this page needs a
// tool that does not exist".
//
// Held as a const with a LOCKSTEP TEST against builderForPageType("tool")
// rather than read from it at runtime, deliberately: if page-type routing ever
// gains a real tool builder, that map's value changes and a runtime read would
// silently change what this gate writes into spec.builder_needed — the field
// the roadmap sweep groups on. A failing test with a human deciding is the
// better failure.
const toolProducerSlug = "tool-builder"

// toolPageRole is the canonical page_type this gate is scoped to. Only 'tool':
// no component_level='game' exists anywhere in the estate (measured
// 2026-09-03 — the levels are element/footer/fragment/head/header/section/site/
// tool), so a game arm could never resolve and would hold such pages for ever;
// entity-page already reaches a capability_gap through builderForPageType.
const toolPageRole = "tool"

// ToolSourceResolution reports whether a planned page is a tool page and whether
// a tool that could fill it exists for the site being planned.
type ToolSourceResolution struct {
	ToolFamily     bool   // page_type is 'tool'
	Producible     bool   // meaningful only when ToolFamily: a live tool component matches
	ProducerNeeded string // when !Producible: the slug capability_gap consumers group on
	Evidence       string // one line saying what was checked and what it found
}

// isToolRole is the family test. Role only — NOT a section-name heuristic:
// hero-tool, tool-guide-intro and tool-cta are ordinary section components that
// share the `tool-` prefix with tool functions, so a prefix test cannot
// discriminate, and a non-tool page that embeds a tool section is not this
// bug's class (it has no /tools/ URL to serve as a shell).
func isToolRole(role string) bool { return role == toolPageRole }

// siteToolFunctions reads, ONCE per gate run, the set of tool functions this
// site already owns. Deliberately the same predicate create_tool_component uses
// for its own "does this site already have this tool" probe
// (create_tool_component_action.go: cc.function + component_level='tool' +
// p.site_id + is_active), because these two must agree about what "the site has
// this tool" means or the gate holds pages the tool pipeline believes it has
// already built.
//
// A LIBRARY tool that is not deployed to this site does NOT count, and that is
// the point: §7 established that nothing consumes planned tool pages, so a page
// kept on the strength of a library match would be a page nothing will ever
// fill — the shell door reopened one level along.
func siteToolFunctions(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]bool, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT cc.function
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND cc.component_level = 'tool'
		  AND cc.is_active = true
		  AND cc.function IS NOT NULL AND cc.function <> ''
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	set := map[string]bool{}
	for rows.Next() {
		var fn string
		if scanErr := rows.Scan(&fn); scanErr != nil {
			return nil, scanErr
		}
		set[strings.ToLower(strings.TrimSpace(fn))] = true
	}
	return set, rows.Err()
}

// toolFunctionCandidates lists the names under which a planned tool page's tool
// could exist, mirroring the tool pipeline's own naming contract:
//
//   - every SECTION the plan names — a realised tool page carries its widget as a
//     section whose slot_name IS the tool function, which is how advertise.co.uk's
//     plan came to name `tool-ab-test-calculator` and why it must resolve;
//   - the page NAME itself — create_tool_component sets pages.name = cc.function
//     (the acceptance coupling), so a tool page built by the pipeline matches here;
//   - "tool-" + the page name — sanitiseFunction guarantees the `tool-` prefix on
//     every function, while resolveToolPageIdentity still accepts a legacy page
//     named for the bare slug, so both spellings must be tried.
//
// Matching on ANY candidate keeps the page. This arm is what stops the gate
// filing false gaps against the tool pipeline's OWN pages when a replan echoes
// them back through the plan.
func toolFunctionCandidates(page planPageView) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(page.Sections)+2)
	add := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, s := range page.Sections {
		add(s)
	}
	add(page.Name)
	if !strings.HasPrefix(strings.ToLower(page.Name), "tool-") {
		add("tool-" + page.Name)
	}
	return out
}

// ResolveToolItemSource classifies one planned page against the site's tool set.
//
// PURE, unlike its listing sibling: the set is read once per gate run and passed
// in, so this costs one query per plan rather than one per page and the whole
// decision is testable without a database.
func ResolveToolItemSource(page planPageView, siteTools map[string]bool) ToolSourceResolution {
	if !isToolRole(page.Role) {
		return ToolSourceResolution{}
	}
	candidates := toolFunctionCandidates(page)
	for _, c := range candidates {
		if siteTools[c] {
			return ToolSourceResolution{
				ToolFamily: true, Producible: true,
				Evidence: fmt.Sprintf("site owns tool component %q", c),
			}
		}
	}
	return ToolSourceResolution{
		ToolFamily:     true,
		Producible:     false,
		ProducerNeeded: toolProducerSlug,
		Evidence: fmt.Sprintf("no active tool component on this site matches any of %d candidate function name(s) (%s); "+
			"tools are created by the tool pipeline (design-discovery evaluate_tools → tool-suggester → tool-deployer), "+
			"which names its own pages — a planned tool page is not its input",
			len(candidates), strings.Join(candidates, ", ")),
	}
}

// enforceToolItemSources is the gate. Structure mirrors enforceListingItemSources
// deliberately — same preserve-guard rule, same drop-recording door, same
// fail-open posture — so the two arms cannot drift in the properties that make a
// plan-time hold safe. What differs is only the resolver and one census read.
//
// HOLDS EMPTY-SECTIONED TOOL PAGES TOO, and that is a decision rather than an
// oversight. A sectionless tool page does not "park harmlessly": bugs_open/450's
// websitepromotion instance is the worked case — the link repair parked all 7 of
// its unbuilt_internal_link items at mark_no_ready_sections (a HUMAN queue) and a
// needs_content_page joined them, recurring per remake, on a page row §7 proved
// no producer will ever fill. "No shell" converts a served-prose bug into a
// permanent HITL tax plus a live phantom-link source. Holding the page at plan
// time means no pages row, no nav entry, no link — neither branch of the fork.
func enforceToolItemSources(ctx context.Context, params ActionParams, pages []interface{}, existingPages []interface{}) []interface{} {
	if params.DB == nil {
		params.Logger.Warn("tool_item_sources: enforce_tool_sources set but no DB — gate skipped (fail-open)")
		return pages
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		params.Logger.Warn("tool_item_sources: no parseable site id — gate skipped (fail-open)",
			zap.String("site_id", siteIDStr))
		return pages
	}

	planViews := make([]planPageView, 0, len(pages))
	anyToolPage := false
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			v := pageViewFromMap(pm)
			planViews = append(planViews, v)
			if isToolRole(v.Role) {
				anyToolPage = true
			}
		}
	}
	// No tool page in the plan is the common case; do not spend the census on it.
	if !anyToolPage {
		return pages
	}

	siteTools, err := siteToolFunctions(ctx, params.DB, siteID)
	if err != nil {
		// Infrastructure failure, not a positive answer — stand down entirely.
		// Holding every tool page on an unreadable census would be the starvation
		// failure the §7 measurement exists to rule out, arriving by accident.
		params.Logger.Warn("tool_item_sources: site tool census failed — gate skipped (fail-open)",
			zap.String("site_id", siteIDStr), zap.Error(err))
		return pages
	}

	realisedKeys := make(map[string]bool, len(existingPages))
	for _, p := range existingPages {
		if pm, ok := p.(map[string]interface{}); ok {
			v := pageViewFromMap(pm)
			if v.Name != "" {
				realisedKeys[v.Name] = true
			}
			if v.URL != "" {
				realisedKeys[v.URL] = true
			}
		}
	}

	kept := make([]interface{}, 0, len(pages))
	var drops []droppedSectionName
	viewIdx := 0
	for _, p := range pages {
		if _, ok := p.(map[string]interface{}); !ok {
			kept = append(kept, p)
			continue
		}
		view := planViews[viewIdx]
		viewIdx++
		res := ResolveToolItemSource(view, siteTools)
		if !res.ToolFamily || res.Producible {
			kept = append(kept, p)
			continue
		}
		// PRESERVE GUARD (bugs_open/001 owns built pages): a realised page is never
		// dropped from the plan — including the 450 shells, whose removal is
		// instance work and not a plan-time decision. The gap receipt is still
		// filed, so the page's state is recorded rather than silently accepted.
		if realisedKeys[view.Name] || (view.URL != "" && realisedKeys[view.URL]) {
			params.Logger.Warn("tool_item_sources: realised tool page has no tool component — kept (preserve guard); capability_gap filed",
				zap.String("page", view.Name),
				zap.String("evidence", res.Evidence))
			fileToolCapabilityGap(ctx, params, siteID, view, res)
			kept = append(kept, p)
			continue
		}
		params.Logger.Warn("tool_item_sources: held tool page with no tool component",
			zap.String("page", view.Name),
			zap.String("producer_needed", res.ProducerNeeded),
			zap.String("evidence", res.Evidence))
		fileToolCapabilityGap(ctx, params, siteID, view, res)
		drops = append(drops, droppedSectionName{Page: view.Name, Name: res.ProducerNeeded})
	}
	if len(drops) > 0 {
		attempted, recorded := LogActionFindings(ctx, params, siteIDStr, "", "validate_plan",
			toolDropFindings(drops), params.Logger)
		warnUnrecordedDrops(attempted, recorded, params.Logger)
	}
	return kept
}

// fileToolCapabilityGap files ONE deferred capability_gap row for a held tool
// page, through the SHARED work-item writer.
//
// The item_key is `capability_gap:tool:<page>` — the SAME key builderForPageType's
// unavailable-builder arm already mints for a tool page (builder_routing.go), so
// the two paths CO-DEDUP on the partial unique index rather than filing two rows
// for one page. That is deliberate: one page, one gap, whichever gate sees it first.
func fileToolCapabilityGap(ctx context.Context, params ActionParams, siteID uuid.UUID, view planPageView, res ToolSourceResolution) {
	gapKey := fmt.Sprintf("capability_gap:%s:%s", view.Role, view.Name)
	gapSpec, _ := json.Marshal(map[string]interface{}{
		"gap_kind":       "producer_missing",
		"page_name":      view.Name,
		"page_type":      view.Role,
		"builder_needed": res.ProducerNeeded,
		"reason":         res.Evidence,
	})
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		params.Logger.Error("tool_item_sources: capability_gap tx open failed",
			zap.String("page", view.Name), zap.Error(err))
		return
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "validate_site_plan",
		pipeline:     "build",
		itemType:     "capability_gap",
		severity:     "low",
		summary:      fmt.Sprintf("Tool page '%s' has no tool component on this site — held from the plan", view.Name),
		spec:         string(gapSpec),
		priority:     200,
		handlerAgent: "",
		status:       "deferred",
		createdBy:    "validate_site_plan",
		itemKey:      gapKey,
	}, params.Logger)
	if err != nil {
		_ = tx.Rollback()
		params.Logger.Error("tool_item_sources: capability_gap insert failed",
			zap.String("page", view.Name), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		params.Logger.Error("tool_item_sources: capability_gap commit failed",
			zap.String("page", view.Name), zap.Error(err))
		return
	}
	if !inserted {
		params.Logger.Info("tool_item_sources: capability_gap already on file (dedup)",
			zap.String("item_key", gapKey))
	}
}

// toolDropFindings converts gate holds to durable findings rows, so "the gate
// fired" and "the planner proposed no tool pages" can never produce identical
// evidence.
func toolDropFindings(drops []droppedSectionName) []agenterrors.Finding {
	findings := make([]agenterrors.Finding, 0, len(drops))
	for _, d := range drops {
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "TOOL_PAGE_HELD_NO_TOOL_SOURCE",
			Severity:  "warning",
			Message: fmt.Sprintf("tool page %q held from plan: no tool component for this site (producer needed %q)",
				d.Page, d.Name),
			Context: map[string]interface{}{
				"page":            d.Page,
				"producer_needed": d.Name,
			},
		})
	}
	return findings
}
