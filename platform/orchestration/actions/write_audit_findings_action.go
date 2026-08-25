// FILE: platform/orchestration/actions/write_audit_findings_action.go
//
// WriteAuditFindingsAction takes structured findings from an LLM audit
// and creates properly classified site_work_items.
//
// The LLM audit produces raw observations: category, description, severity,
// affected pages. This action does the classification:
//
//   1. Loads existing pages for the site (one query, cached for all findings)
//   2. For each finding, determines:
//      - Does the referenced page exist? → content fix on existing page
//      - Is the page name a placeholder ("new page needed", "site-wide")? → classify by category
//      - Is it a gap that needs a new page? → route to content-gap-planner
//      - Is it a metadata/config issue? → route to spec updater
//      - Is it a design issue? → route to design handler
//   3. Creates work items with correct item_type, handler_agent, and page_id
//
// Registration:
//   "write_audit_findings": {
//       Handler:     WriteAuditFindingsAction,
//       Category:    "site",
//       Description: "Classify and store LLM audit findings as work items",
//       IsLocal:     true,
//   },

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var WriteAuditFindingsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id", "audit_source"},
	Optional:    []string{"findings_field"},
	Deprecated:  map[string]string{},
	// filing_mode is a SETTING, not a data reference, so it lives in ConfigKeys
	// (the conditional_branch precedent) and does not move this action's
	// optional-key budget (RFC_022 counts len(Optional) only). Declaring it opts
	// the action into unknown-config-key detection; every live step carries
	// exactly site_id / audit_source / findings_field [MEASURED 2026-08-25, 5 live
	// steps], so nothing is newly reported by that.
	ConfigKeys: []string{"filing_mode"},
	ConditionalKeys: map[string]string{
		"filing_mode": "\"record\" files every routable finding as a VERDICT row (status 'deferred', handler_agent '', routing preserved in spec.routed_handler) that no promoter can dispatch; absent or \"dispatch\" is the historical behaviour",
	},
}

// ============================================================================
// filing_mode — record a verdict without dispatching a repair (RFC_056)
// ============================================================================
//
// WHAT IT IS. An opt-in step setting. Absent (or "dispatch"), this action
// behaves exactly as it always has: a finding becomes a 'detected' work item
// naming the handler that will regenerate the page. "record" files the SAME
// row — same item_type, same dedup key, same spec — but parked: status
// 'deferred', handler_agent '', with the handler it WOULD have been routed to
// kept in spec.routed_handler. Both promoters refuse such a row by construction
// (detected-item-promoter: COALESCE(handler_agent,'') <> ''; triage_detected_items:
// workItemRoutableSQL), so nothing can dispatch it until a human or a later
// migration releases it with the recipe in spec.release_recipe.
//
// WHY IT EXISTS (owner, 2026-08-25). The LLM audit seats of the improvement
// loop — visual-design, content-quality, strategic review, offer analysis,
// brief fidelity — file findings that are OPINIONS about pages that already
// work ("aspirational improvements"), and this router turns them into
// content_rewrite / needs_content_page / needs_content_planning items whose
// handlers REGENERATE the page. [MEASURED 2026-08-25, live+archive] design-audit
// alone filed 976 content_rewrite, 399 needs_content_page and 964
// needs_content_planning over its life, and bugs_open/238 records what one such
// rewrite did to a page that was fine (five <img src=""> on a live homepage,
// asked for by nobody). The owner's ruling: keep the seats — they are the site
// acceptance council — but stop the rewrites. IMP-006 (register) is the same
// remedy proposed four times since March and never built; this is its first
// shipped piece, as a filing mode rather than a per-site approval column.
//
// WHY 'deferred' + '' AND NOT 'detected'. IMP-054 believed 'detected' was a
// safe parking status ("a lone discovery run files findings nothing can ever
// see"). That stopped being true on 2026-08-15: detected-item-promoter promotes
// any 'detected' row whose (item_type, handler) pair has ever completed, every
// 15 minutes, whether or not the improvement sweep is on. [MEASURED 2026-08-25]
// 26 LLM-audit rows were promoted between 2026-08-20 and 2026-08-24 while the
// sweep was disabled. 'detected' is a queue, not a shelf. 'deferred' + empty
// handler is the estate's one parking convention nothing promotes
// (capability_gap, bugs_closed/077; remit.go), and bugs_open/396 is exactly the
// warning against the other shape — 'deferred' WITH a named handler, no
// provenance — which is why the routing moves into spec rather than staying in
// the column.
//
// WHAT IT DOES NOT DO. It does not change what the seat OBSERVES or how it is
// classified; it does not touch findings the router already parks
// (capability_gap rows are 'deferred' + '' already and pass through unchanged);
// and it does not release anything. A record row holds the finding's dedup slot
// like any open row, so a later dispatch-mode filing of the same finding is a
// duplicate until the record row is released or closed — one row per finding,
// whichever mode filed it.

const (
	filingModeDispatch = "dispatch"
	filingModeRecord   = "record"
)

// parseFilingMode reads the literal setting. Absent or empty means dispatch.
// Anything other than the two known values is an ERROR, deliberately: a typo
// ("recrod") must not silently dispatch the rewrites the setting exists to
// stop. Pure, so the precedence is unit-testable without an orchestrator.
func parseFilingMode(config map[string]interface{}) (string, error) {
	raw, present := config["filing_mode"]
	if !present || raw == nil {
		return filingModeDispatch, nil
	}
	mode, ok := raw.(string)
	if !ok {
		return "", fmt.Errorf("write_audit_findings: filing_mode must be a string, got %T", raw)
	}
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", filingModeDispatch:
		return filingModeDispatch, nil
	case filingModeRecord:
		return filingModeRecord, nil
	}
	return "", fmt.Errorf("write_audit_findings: filing_mode %q is not one of dispatch|record — refusing rather than guessing, because the wrong guess dispatches a page rewrite", mode)
}

// recordOnlyFinding turns a routable classified finding into its verdict-row
// form. A finding the router already parks (empty handler, e.g. capability_gap)
// is returned unchanged — parking it harder would only erase the provenance the
// fallback wrote. Pure; the transform is what the tests pin.
func recordOnlyFinding(c classifiedFinding, auditSource string) classifiedFinding {
	if c.HandlerAgent == "" {
		return c
	}
	spec := make(map[string]interface{}, len(c.Spec)+6)
	for k, v := range c.Spec {
		spec[k] = v
	}
	routedStatus := c.Status
	if routedStatus == "" {
		routedStatus = "detected"
	}
	spec["filing_mode"] = filingModeRecord
	spec["routed_handler"] = c.HandlerAgent
	spec["routed_status"] = routedStatus
	spec["deferred_by"] = auditSource
	spec["deferred_reason"] = "filing_mode=record (RFC_056): a verdict row — the seat's finding is recorded, the repair it would have dispatched is not"
	spec["not_dispatchable"] = "status 'deferred' + empty handler_agent — deliberate; the routing is preserved in spec.routed_handler and nothing promotes this row (bugs_closed/077 convention, bugs_open/396 provenance)"
	spec["release_recipe"] = "UPDATE site_work_items SET status = spec->>'routed_status', handler_agent = spec->>'routed_handler', updated_at = now() WHERE id = <id> AND status = 'deferred' AND spec->>'filing_mode' = 'record'"
	out := c
	out.Spec = spec
	out.HandlerAgent = ""
	out.Status = "deferred"
	out.Summary = "[verdict, not dispatched] " + c.Summary
	return out
}

func init() {
	datahelpers.RegisterActionInputSpec("write_audit_findings", WriteAuditFindingsInputSpec)
}

// ============================================================================
// Classification rules — deterministic routing based on category + page existence
// ============================================================================

// Design categories route to design/CSS handlers regardless of page existence
var designCategories = map[string]struct{}{
	"colour": {}, "color": {}, "spacing": {}, "typography": {},
	"header_footer": {}, "dark_section": {}, "responsive": {},
}

// Categories that indicate site-wide config/metadata issues, not page content
var metadataCategories = map[string]struct{}{
	"metadata": {}, "config": {}, "identity": {}, "spec": {},
	"contact_mismatch": {},
}

// Categories whose ONLY route was a handler that refuses them by design.
//
// Until 2026-08-19 these filed cta_improvement / nav_restructure at
// component-template-fixer, whose dispatch table has returned
// {fixed:false, action:"needs_review", reason:"fix_type requires LLM-driven
// changes, not programmatic HTML edits"} for both since 2026-03-14
// (fix_component_template_action.go, fixTypesRefusedByDesign). Nothing read the
// flag, so [MEASURED 2026-08-19, archive-inclusive] 993 cta_improvement items
// across 22 sites closed 'complete' with 0 ever fixed (bugs_open/323). Rule 3
// below now files them as the estate's "found work I have no handler for" record
// (capability_gap, bugs_closed/077) until a handler that can do LLM-driven CTA /
// navigation copy work exists — at which point this map's categories move to a
// real route and TestAuditRoutingNeverTargetsAFixerRefusalArm keeps that route
// honest.
var noHandlerCategories = map[string]struct{}{
	"cta": {}, "nav_restructure": {},
}

// Page names that are not real pages — they're audit shorthand
var placeholderPageNames = map[string]bool{
	"new page needed": true,
	"new page":        true,
	"site-wide":       true,
	"general":         true,
	"all pages":       true,
	"multiple pages":  true,
	"":                true,
}

// Normalise common page name aliases
var pageNameAliases = map[string]string{
	"homepage": "index",
	"home":     "index",
}

// Design category → handler agent
//
// Read at FILING time only (Rule 1 below), so a change here affects newly-filed
// items and never re-routes rows already carrying a handler_agent.
var designRouting = map[string]string{
	"colour":        "webdesign-agent",
	"color":         "webdesign-agent",
	"spacing":       "component-template-fixer",
	"typography":    "webdesign-agent",
	"header_footer": "site-component-linker",

	// dark_section was routed to color-variable-fixer until 2026-08-19. OWNER RULING
	// that date, on the choice this lane put to him: route it to a handler that can
	// actually make these changes.
	//
	// WHY THE OLD ROUTE COULD NEVER WORK, measured rather than argued. These findings
	// ask for SCOPED CSS BLOCKS DEFINING SECTION TOKENS — the live suggestions read
	// "add a scoped style block for .cta-section that defines --section-text: #ffffff;
	// --section-text-muted: …", "add a .site-footer scoped block defining
	// --section-text, --section-heading, --section-surface and --section-bg". That is
	// an ADDITIVE STYLESHEET RULE. color-variable-fixer's transform
	// (checks.ReplaceHardcodedColors) does something else entirely: it rewrites dark
	// hex literals found INSIDE component bodies, and its whole output alphabet is
	// var(--color-primary) and var(--color-secondary). A transform with two words
	// cannot satisfy a criterion naming five others, on any input — so [MEASURED
	// 2026-08-19, archive-inclusive] all 26 completions of this type across 17 sites
	// reported a repair count of ZERO, and not one ever reported a non-zero fix.
	//
	// WHY css-patch-agent. Its workflow IS this job — load_current_css → plan_css_fix
	// → save_css_to_db → deploy_css — and it already carries the sibling colour
	// concern (contrast_failure, 287 items). Capability confirmed at the ARTEFACT, not
	// from its config: [MEASURED 2026-08-19] 39 of its completions carry a git-adapter
	// response deploying assets/css/styles.css, most recent 2026-08-18.
	//
	// WHY NOT webdesign-agent, which takes every other colour category: for its main
	// type (needs_design_review) the owner ruled the same day that THE ANALYSIS IS THE
	// DELIVERABLE, and [MEASURED] all 1,268 of its completions carry no response
	// envelope at all. Routing a repair need at an analysis producer is the same
	// mistake this line is fixing, in different clothes.
	//
	// ⚠ READ complete_work_item_no_change.go BEFORE RE-ENABLING A CARRIER FOR THIS
	// TYPE. Gate 1b's roster entry for dark_section_audit is licensed by a measurement
	// about color-variable-fixer's transform, and this line has just voided it. The
	// entry is deliberately left in place — its failure direction is a REFUSAL, which
	// is safe — but it must be re-measured against css-patch-agent, or removed, before
	// items flow again. Both carriers are enabled=false today (owner, 2026-08-19:
	// keep them off for now), so this route is INERT until that changes.
	"dark_section": "css-patch-agent",

	"responsive": "component-template-fixer",
}

// Category → fix_type mapping for component-template-fixer
// Ensures fix_type is always set in spec when routing to that handler.
//
// Every value here must be a fix_type the handler ACCEPTS — not one in its
// fixTypesRefusedByDesign set (bugs_open/323; pinned by
// TestAuditRoutingNeverTargetsAFixerRefusalArm). "cta" → "cta_improvement" and
// "nav_restructure" lived here until 2026-08-19 and were refused on every run.
var categoryToFixType = map[string]string{
	"spacing":    "inject_nav_flex_css",
	"responsive": "responsive_fix",
}

// Design category → item type
var designItemTypes = map[string]string{
	"colour":        "needs_design_review",
	"color":         "needs_design_review",
	"spacing":       "spacing_fix",
	"typography":    "needs_design_review",
	"header_footer": "header_footer_fix",
	// dark_section gets its OWN item_type (bugs_open/213). It used to file under
	// hardcoded_section_colors, whose registered verifier belongs to the DISCOVERY
	// check of that name and re-runs that check's predicate — "the
	// color-variable-fixer's transform has nothing left to do on this site", a
	// site-level aggregate. A design-audit finding is not an aggregate: it names one
	// section's defect and carries its own spec.acceptance_test. The verifier
	// answered its own question correctly and returned Resolved:true, so every
	// design-audit item on this route closed 'complete' untouched — 11 of 11, none
	// of which has ever failed to close, while every item that ever DID fail to
	// close was the discovery check's.
	//
	// An item_type is the join between who filed an item and what predicate regrades
	// it before it closes. Two producers under one name is two predicates behind one
	// join, and the registry cannot see the difference.
	"dark_section": "dark_section_audit",
	"responsive":   "responsive_fix",
}

// ============================================================================
// Page info cache — loaded once per invocation
// ============================================================================

type pageInfo struct {
	ID       uuid.UUID
	Name     string
	PageType string
	Sections string
}

func loadSitePages(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]pageInfo, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, name, COALESCE(page_type, ''), COALESCE(sections::text, '[]')
		FROM pages
		WHERE site_id = $1 AND status = 'active'
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pages := make(map[string]pageInfo)
	for rows.Next() {
		var p pageInfo
		if err := rows.Scan(&p.ID, &p.Name, &p.PageType, &p.Sections); err != nil {
			continue
		}
		pages[p.Name] = p
	}
	return pages, nil
}

// ============================================================================
// Finding struct — what the LLM produces
// ============================================================================

type auditFinding struct {
	Category          string   `json:"category"`
	Severity          string   `json:"severity"`
	Description       string   `json:"description"`
	Suggestion        string   `json:"suggestion"`
	Page              string   `json:"page"`
	AffectedPages     []string `json:"affected_pages"`
	FixType           string   `json:"fix_type"`
	AffectedComponent string   `json:"affected_component"`
	// Structured findings fields (migration 084)
	CurrentValue   string `json:"current_value"`
	AcceptanceTest string `json:"acceptance_test"`
	MaxFixAttempts int    `json:"max_fix_attempts"`
	// The optional machine-checkable half of the acceptance test, and the
	// record of one this platform refused (features_open/030 v2(d)).
	//
	// PASSTHROUGH ONLY. This action neither validates nor evaluates either key:
	// verify_acceptance_predicates does that, upstream, and a producer whose
	// workflow does not carry that step will never populate them. Absent → both
	// keys are omitted from the spec and this file behaves byte-for-byte as it
	// did before. That is the 2026-08-02 shared-seam shape: an opt-in field with
	// the unsafe side off, on an action with many producers.
	AcceptancePredicate         map[string]interface{} `json:"acceptance_predicate"`
	AcceptancePredicateRejected map[string]interface{} `json:"acceptance_predicate_rejected"`
}

// ============================================================================
// Classified result — what we insert as a work item
// ============================================================================

type classifiedFinding struct {
	ItemType     string
	HandlerAgent string
	Severity     string
	Priority     int
	PageID       *uuid.UUID
	PageName     string
	Spec         map[string]interface{}
	DedupKey     string
	// Status is the row's initial status; empty means 'detected'. Only the
	// capability_gap fallback sets it ('deferred' — a roadmap row must not be
	// promoted by the site-wide triage pass, which keys on 'detected').
	Status string
	// Summary overrides the finding's own description as the row summary;
	// empty means the description. Set where the description alone would
	// read as dispatchable work rather than what the row actually is.
	Summary string
}

// ============================================================================
// classify — the core algorithmic routing
// ============================================================================

func classifyFinding(f auditFinding, pages map[string]pageInfo, siteID uuid.UUID, auditSource string) classifiedFinding {
	category := strings.ToLower(strings.TrimSpace(f.Category))
	pageName := strings.TrimSpace(f.Page)

	// Normalise page name aliases
	if alias, ok := pageNameAliases[strings.ToLower(pageName)]; ok {
		pageName = alias
	}

	// Use first affected_page if page is empty
	if pageName == "" && len(f.AffectedPages) > 0 {
		pageName = strings.TrimSpace(f.AffectedPages[0])
		if alias, ok := pageNameAliases[strings.ToLower(pageName)]; ok {
			pageName = alias
		}
	}

	severity := f.Severity
	if severity == "" {
		severity = "medium"
	}
	priority := severityToPriority(severity)

	// Build base spec (always included)
	spec := map[string]interface{}{
		"audit_source":    auditSource,
		"category":        f.Category,
		"description":     f.Description,
		"original_domain": "build",
	}
	if f.Suggestion != "" {
		spec["suggestion"] = f.Suggestion
	}
	if f.AffectedComponent != "" {
		spec["affected_component"] = f.AffectedComponent
	}
	if f.FixType != "" {
		spec["fix_type"] = f.FixType
	}

	// Structured findings fields (migration 084)
	if f.CurrentValue != "" {
		spec["current_value"] = f.CurrentValue
	}
	if f.AcceptanceTest != "" {
		spec["acceptance_test"] = f.AcceptanceTest
	}
	if f.MaxFixAttempts > 0 {
		spec["max_fix_attempts"] = f.MaxFixAttempts
	}
	if len(f.AcceptancePredicate) > 0 {
		spec["acceptance_predicate"] = f.AcceptancePredicate
	}
	if len(f.AcceptancePredicateRejected) > 0 {
		spec["acceptance_predicate_rejected"] = f.AcceptancePredicateRejected
	}

	// If fix_type still not set, derive from category for component-template-fixer compatibility
	if _, hasFixType := spec["fix_type"]; !hasFixType {
		if derived, ok := categoryToFixType[category]; ok {
			spec["fix_type"] = derived
		}
	}

	// ── Rule 1: Design categories → design handlers (page existence irrelevant)
	if _, isDesign := designCategories[category]; isDesign {
		handler := designRouting[category]
		itemType := designItemTypes[category]
		if handler == "" {
			handler = "component-template-fixer"
		}
		if itemType == "" {
			itemType = "needs_design_review"
		}

		var pageID *uuid.UUID
		if pageName != "" {
			if p, exists := pages[pageName]; exists {
				pageID = &p.ID
			}
		}
		spec["page_name"] = pageName

		return classifiedFinding{
			ItemType:     itemType,
			HandlerAgent: handler,
			Severity:     severity,
			Priority:     priority,
			PageID:       pageID,
			PageName:     pageName,
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_%s_%s_%s", auditSource, itemType, pageName, siteID),
		}
	}

	// ── Rule 2: Metadata/config categories → spec update (not a page issue)
	if _, isMeta := metadataCategories[category]; isMeta {
		spec["page_name"] = ""
		return classifiedFinding{
			ItemType:     "needs_spec_update",
			HandlerAgent: "spec-updater",
			Severity:     severity,
			Priority:     priority,
			PageID:       nil,
			PageName:     "",
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_needs_spec_update_%s_%s", auditSource, category, siteID),
		}
	}

	// ── Rule 3: Categories with NO capable handler → capability_gap (bugs_open/323)
	//
	// These used to dispatch cta_improvement / nav_restructure at
	// component-template-fixer, which refuses both by design (see
	// noHandlerCategories). A finding the estate cannot act on is still worth
	// recording — it is the demand signal for the missing handler — so it takes
	// the bugs_closed/077 shape: status 'deferred', empty handler_agent, low
	// severity, priority 200, gap_kind handler_missing. The finding's own
	// severity, description, suggestion and acceptance_test stay in spec for
	// whoever builds the handler. One open row per site per category (dedup on
	// the category, not the page): the missing thing is a HANDLER, however many
	// pages or producers report it.
	//
	// Distinct from the unknown-category fallback at the bottom of this function
	// (gap_kind rule_missing): there the ROUTER lacks a rule; here the router
	// knows exactly what the finding is and the estate has nothing to send it to.
	if _, noHandler := noHandlerCategories[category]; noHandler {
		spec["page_name"] = pageName
		spec["finding_severity"] = severity
		spec["gap_kind"] = checks.GapHandlerMissing
		spec["builder_needed"] = fmt.Sprintf(
			"a handler for LLM-driven %s work (CTA / navigation copy: rewrite labels and destinations "+
				"on one named component via section-editor field_updates); component-template-fixer "+
				"refuses this fix_type by design (bugs_open/323)", category)
		spec["capability"] = fmt.Sprintf(
			"%s findings from %s have no handler: the only programmatic fixer declines them "+
				"(fix_component_template_action.go fixTypesRefusedByDesign) and no LLM copy editor is routed",
			category, auditSource)
		spec["not_dispatchable"] = "status 'deferred' + empty handler_agent — deliberate; " +
			"promoting this row dispatches work no handler can do (bugs_open/077, bugs_open/323)"
		return classifiedFinding{
			ItemType:     "capability_gap",
			HandlerAgent: "",
			Severity:     "low",
			Priority:     200,
			Status:       "deferred",
			PageID:       nil,
			PageName:     pageName,
			Summary: fmt.Sprintf("no handler for audit category %q (%s): %s",
				category, auditSource, f.Description),
			Spec:     spec,
			DedupKey: fmt.Sprintf("capability_gap:no_handler_for_audit_category:%s", category),
		}
	}

	// ── Rule 4: Content categories — depends on whether the page exists
	isPlaceholder := placeholderPageNames[strings.ToLower(pageName)]

	if !isPlaceholder {
		if p, exists := pages[pageName]; exists {
			// Page exists. Routing depends on the nature of the finding:
			//   - gap: missing content on an existing page → rebuild (not rewrite)
			//   - tone: stylistic adjustment to existing content → tone_shift
			//   - other (content, differentiation, structure): rewrite existing content
			pageID := p.ID
			spec["page_name"] = pageName

			switch category {
			case "gap":
				// Missing content on an existing page — trigger a rebuild.
				// The page-build-handler will regenerate sections that are empty
				// or have missing data. This is structurally different from
				// rewriting existing content — empty sections need building, not editing.
				return classifiedFinding{
					ItemType:     "needs_content_page",
					HandlerAgent: "page-build-handler",
					Severity:     severity,
					Priority:     priority,
					PageID:       &pageID,
					PageName:     pageName,
					Spec:         spec,
					DedupKey:     fmt.Sprintf("%s_needs_content_page_%s_%s", auditSource, pageName, siteID),
				}

			case "tone":
				// A tone finding is a STYLISTIC adjustment to copy that already
				// exists, and it used to be filed `tone_shift` at
				// page-build-handler — which regenerates the page. That is the
				// wrong instrument, and the estate has the incident to prove it:
				// `save_page_sections` DELETEs and re-INSERTs the row, so
				// finetuning.uk's homepage "lost all 11 of its non-llm URL keys to
				// one tone_shift and served five <img src=""> plus six vanished
				// controls" (plan_sections_action.go's carry-stored comment,
				// bugs_open/238). Asking for a better sentence should not be able
				// to empty an image tag.
				//
				// copy-editor (CQ-024, stage 2) is the surgical alternative: it
				// reads the whole page, emits `field_updates` for named fields on
				// named components, and STRUCTURALLY cannot write to a page —
				// migration 447 RAISEs if any page-writing step is added to it.
				// Its output parks at `copy_edit_proposed`/`needs_human_review`
				// for a human, so this routes an auto-PROPOSAL, never an
				// auto-edit, and owner decision D2 is untouched.
				//
				// A NEW item_type rather than re-pointing `tone_shift`: this is a
				// new (item_type, handler) pair and is held for a human canary,
				// and silently changing what an existing type means would misread
				// the 33 `tone_shift` rows already in history.
				//
				// Volume is why this is safe to land as a canary: `tone` is rare —
				// 33 `tone_shift` items lifetime across live and archive as of
				// 2026-08-24, versus 1,893 `content_rewrite`. Roughly one a week
				// cannot flood the review queue that `bugs_open/033` is about
				// (1,079 items parked as of 2026-08-24). Do NOT extend this to
				// `content_rewrite` on the strength of this comment.
				return classifiedFinding{
					ItemType:     "needs_copy_edit",
					HandlerAgent: "copy-editor",
					Severity:     severity,
					Priority:     priority,
					PageID:       &pageID,
					PageName:     pageName,
					Spec:         spec,
					DedupKey:     fmt.Sprintf("%s_needs_copy_edit_%s_%s", auditSource, pageName, siteID),
				}

			default:
				// content, differentiation, structure, etc. → rewrite existing content
				return classifiedFinding{
					ItemType:     "content_rewrite",
					HandlerAgent: "page-build-handler",
					Severity:     severity,
					Priority:     priority,
					PageID:       &pageID,
					PageName:     pageName,
					Spec:         spec,
					DedupKey:     fmt.Sprintf("%s_content_rewrite_%s_%s", auditSource, pageName, siteID),
				}
			}
		}
	}

	// ── Rule 5: Gap — page doesn't exist or is a placeholder
	// Route to content-gap-planner which decides what to create
	if category == "gap" || category == "content" || category == "differentiation" || category == "structure" {
		spec["page_name"] = pageName

		return classifiedFinding{
			ItemType:     "needs_content_planning",
			HandlerAgent: "content-gap-planner",
			Severity:     severity,
			Priority:     priority + 5,
			PageID:       nil,
			PageName:     "",
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_needs_content_planning_%s_%s", auditSource, sanitiseDedupSegment(f.Description), siteID),
		}
	}

	// ── Rule 6: Tone/content on placeholder page → also goes to planner
	if isPlaceholder && (category == "tone" || category == "content_rewrite") {
		spec["page_name"] = ""

		return classifiedFinding{
			ItemType:     "needs_content_planning",
			HandlerAgent: "content-gap-planner",
			Severity:     severity,
			Priority:     priority + 5,
			PageID:       nil,
			PageName:     "",
			Spec:         spec,
			DedupKey:     fmt.Sprintf("%s_needs_content_planning_%s_%s", auditSource, sanitiseDedupSegment(f.Description), siteID),
		}
	}

	// ── Fallback: unknown category → capability_gap for the roadmap
	//
	// Until 2026-08-15 this minted item_type "audit_finding_"+category — a type
	// registered nowhere, so every such item died in 'detected' (bugs_open/115,
	// bugs_open/279; brief-fidelity-auditor's entire output took this path). An
	// unknown category means THIS ROUTER lacks a rule, not that the finding is
	// work some handler can do — so it files the platform's "found work I have
	// no handler for" shape (bugs_closed/077), which the triage sweep surfaces
	// as a roadmap entry and cannot silently rot. Field choices mirror
	// CapabilityGapItem (discovery_checks/remit.go): status 'deferred', empty
	// handler_agent, severity 'low', priority 200, spec.builder_needed — the
	// row is a signal, not a dispatch. The finding's own severity, category and
	// page stay in spec for whoever writes the missing rule.
	spec["page_name"] = pageName
	spec["finding_severity"] = severity
	spec["gap_kind"] = checks.GapRuleMissing
	spec["builder_needed"] = fmt.Sprintf("write_audit_findings: no route for category %q", category)
	spec["capability"] = fmt.Sprintf(
		"classifyFinding needs a routing rule for category %q before %s findings can reach a handler",
		category, auditSource)
	spec["not_dispatchable"] = "status 'deferred' + empty handler_agent — deliberate; " +
		"promoting this row dispatches work no handler can do (bugs_open/077)"
	return classifiedFinding{
		ItemType:     "capability_gap",
		HandlerAgent: "",
		Severity:     "low",
		Priority:     200,
		Status:       "deferred",
		PageID:       nil,
		PageName:     pageName,
		Summary:      fmt.Sprintf("no route for audit category %q (%s): %s", category, auditSource, f.Description),
		Spec:         spec,
		// One open row per site per unknown category — the missing thing is a
		// rule for the CATEGORY, however many findings or producers hit it.
		DedupKey: fmt.Sprintf("capability_gap:unrouted_audit_category:%s", category),
	}
}

func severityToPriority(severity string) int {
	switch severity {
	case "high":
		return 10
	case "low":
		return 60
	default:
		return 30
	}
}

func sanitiseDedupSegment(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '_'
		}
		return -1
	}, s)
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

// ============================================================================
// Findings parsing — the LLM result reaches collected_data as a JSON string,
// an already-parsed array, or an already-parsed object wrapping the array
// under "findings" (the shape site-review-agent's prompt asks for —
// bugs_open/272: the object case was missing, so a prompt-compliant response
// silently produced zero work items).
// ============================================================================

// findingsFromList maps already-parsed finding objects into auditFinding.
//
// It also reports whether the list was RECOGNISED — every element decoded into
// a finding object. An empty input list is recognised (an auditor legitimately
// reporting nothing wrong); a non-empty list that produced no findings is NOT,
// because every element was some other shape. That distinction is only needed
// by the retraction path (write_audit_findings_retraction.go), where "the
// auditor found nothing" is evidence and "I did not understand the reply" must
// be inert — see bugs_open/213 D1 half two.
func findingsFromList(items []interface{}) (findings []auditFinding, recognised bool) {
	for _, item := range items {
		if m, ok := item.(map[string]interface{}); ok {
			f := auditFinding{
				Category:          getStringFromMap(m, "category"),
				Severity:          getStringFromMap(m, "severity"),
				Description:       getStringFromMap(m, "description"),
				Suggestion:        getStringFromMap(m, "suggestion"),
				Page:              getStringFromMap(m, "page"),
				FixType:           getStringFromMap(m, "fix_type"),
				AffectedComponent: getStringFromMap(m, "affected_component"),
				CurrentValue:      getStringFromMap(m, "current_value"),
				AcceptanceTest:    getStringFromMap(m, "acceptance_test"),
				MaxFixAttempts:    getIntFromMap(m, "max_fix_attempts"),
			}
			// ⚠ TWO DECODE PATHS, AND THIS IS THE ONE THAT MATTERS HERE.
			// findingsFromString unmarshals into the struct, so its json tags
			// are enough; this path maps keys BY HAND, so a field added to the
			// struct and not to this block is silently dropped — and native
			// maps are exactly what an upstream ACTION hands over, which is how
			// verify_acceptance_predicates' output arrives. Held by
			// TestFindingsFromListPopulatesEveryTaggedField.
			if p, ok := m["acceptance_predicate"].(map[string]interface{}); ok {
				f.AcceptancePredicate = p
			}
			if p, ok := m["acceptance_predicate_rejected"].(map[string]interface{}); ok {
				f.AcceptancePredicateRejected = p
			}
			if ap, ok := m["affected_pages"].([]interface{}); ok {
				for _, p := range ap {
					if ps, ok := p.(string); ok {
						f.AffectedPages = append(f.AffectedPages, ps)
					}
				}
			}
			findings = append(findings, f)
		}
	}
	return findings, len(findings) == len(items)
}

// findingsFromString fence-trims and unmarshals a JSON string holding either
// a bare findings array or an object wrapping one under "findings". A
// non-empty parseError means the string held no findings and was not
// parseable as an array.
//
// `recognised` is true only when a findings array was actually decoded — see
// findingsFromList's note. A string that parses as an ARRAY of zero findings
// is recognised; one that fails both unmarshals is not, and neither is a
// wrapper object carrying no "findings" key.
func findingsFromString(s string, logger *zap.Logger) (findings []auditFinding, parseError string, recognised bool) {
	cleaned := strings.TrimSpace(s)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	if err := json.Unmarshal([]byte(cleaned), &findings); err != nil {
		var wrapper map[string]json.RawMessage
		if err2 := json.Unmarshal([]byte(cleaned), &wrapper); err2 == nil {
			if findingsJSON, ok := wrapper["findings"]; ok {
				if uErr := json.Unmarshal(findingsJSON, &findings); uErr == nil {
					return findings, "", true
				}
			}
		}
		if len(findings) == 0 {
			logger.Warn("Failed to parse findings JSON",
				zap.Error(err),
				zap.String("raw_preview", cleaned[:min(200, len(cleaned))]))
			return nil, err.Error(), false
		}
		return findings, "", false
	}
	return findings, "", true
}

// parseAuditFindings accepts all three shapes. Only an unparseable string
// sets parseError; an unrecognised shape, or an object with no "findings"
// key, returns (nil, "", false) and the caller reports the zero-findings
// reason together with the offending type.
//
// ⚠ THE THIRD RETURN IS WHY THIS FUNCTION CHANGED (bugs_open/213 D1 half two).
// `len(findings)==0 && parseError==""` IS AMBIGUOUS and always was — this
// function's own contract above says so — because it is returned BOTH for an
// auditor that looked and found nothing AND for a reply whose shape was never
// recognised. Reporting zero work items is a correct response to either, so
// nothing before now had to tell them apart. A retraction rule does: silence
// is only evidence when the instrument was read successfully, and "I did not
// understand the reply" must be inert. Do not collapse this back into
// len(findings)==0.
func parseAuditFindings(findingsRaw interface{}, logger *zap.Logger) (findings []auditFinding, parseError string, recognised bool) {
	switch v := findingsRaw.(type) {
	case string:
		return findingsFromString(v, logger)
	case []interface{}:
		f, ok := findingsFromList(v)
		return f, "", ok
	case map[string]interface{}:
		// The prompt-compliant wrapped shape: mirror the string case's
		// wrapper unwrap.
		switch inner := v["findings"].(type) {
		case []interface{}:
			f, ok := findingsFromList(inner)
			return f, "", ok
		case string:
			return findingsFromString(inner, logger)
		}
	}
	return nil, "", false
}

// ============================================================================
// Main action
// ============================================================================

// pendingCopyEditForPage reports whether the page already carries an open
// copy-edit item or an un-reviewed stage-2 proposal. It is the tone route's
// convergence bound (copy_quality_two_stage handoff 2026-08-25, item 3):
// stage 2 keeps proposing on repeated runs over the same page — run 5
// re-edited 2 of the 3 components run 4 had just changed (deeper cuts, not
// oscillation, but unbounded) — and idx_swi_dedup only bounds CONCURRENT
// needs_copy_edit rows: the slot frees when the run completes, while the
// un-reviewed proposal it parked is a different item_type holding nothing.
// The dedup key also embeds audit_source, so two auditors could otherwise
// file the same page in parallel.
//
// So the bound is structural, not a clock: while a page has EITHER an open
// needs_copy_edit (any producer) OR an un-reviewed copy_edit_proposed, no
// new one is filed. It drains when a human acts on the parked proposal —
// which is owner decision D2's posture exactly: the human is the rate
// limiter, so the queue can never outrun the person reviewing it.
// The status list rides as a PARAMETER (`<> ALL(string_to_array($3,','))`),
// not interpolated SQL — the constitution's "parameterised always" applies to
// package constants too (council 754dcffd round 1, constitution + guardian
// seats), and string_to_array keeps this free of lib/pq, which this package
// deliberately avoids (resolve_composition_helpers.go, asset_lock_guard.go).
// None of the statuses contains a comma; the lockstep test asserts the joined
// form against workItemTerminalStatuses so drift in either direction fails.
func pendingCopyEditForPage(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageID uuid.UUID) (bool, error) {
	var pending bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM site_work_items
			WHERE site_id = $1 AND page_id = $2
			  AND item_type IN ('needs_copy_edit', 'copy_edit_proposed')
			  AND status <> ALL(string_to_array($3, ','))
		)
	`, siteID, pageID, strings.Join(workItemTerminalStatuses, ",")).Scan(&pending)
	return pending, err
}

func WriteAuditFindingsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "write_audit_findings"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	config := params.StepConfig.Config

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, config, WriteAuditFindingsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// audit_source is Required with no Default (bugs_open/264): a config that
	// fails to resolve now fails ExtractActionInputs outright instead of
	// silently landing every producer's findings under "design-audit".
	auditSource := inputs.Get("audit_source")

	// ── Load existing pages for classification ──
	pages, err := loadSitePages(ctx, params.DB, siteID)
	if err != nil {
		logger.Warn("Failed to load pages for classification, continuing without",
			zap.Error(err))
		pages = make(map[string]pageInfo)
	}
	logger.Info("Loaded site pages for classification",
		zap.Int("page_count", len(pages)),
		zap.String("site_id", siteIDStr))

	// ── Extract findings from LLM response ──
	findingsField := "audit_result.result"
	if f, ok := config["findings_field"].(string); ok && f != "" {
		findingsField = f
	}

	filingMode, fmErr := parseFilingMode(config)
	if fmErr != nil {
		return nil, fmErr
	}
	recordOnly := filingMode == filingModeRecord
	recorded := 0

	findingsRaw := datahelpers.ExtractNestedField(params.CollectedData, findingsField)
	if findingsRaw == nil {
		for _, alt := range []string{"audit_result.response.result", "audit_result", "review_result.result"} {
			findingsRaw = datahelpers.ExtractNestedField(params.CollectedData, alt)
			if findingsRaw != nil {
				break
			}
		}
	}

	if findingsRaw == nil {
		logger.Warn("No findings found", zap.String("field", findingsField))
		return map[string]interface{}{
			"items_created": 0,
			"reason":        "no findings in " + findingsField,
		}, nil
	}

	// ── Parse findings ──
	findings, parseError, shapeRecognised := parseAuditFindings(findingsRaw, logger)
	if parseError != "" {
		return map[string]interface{}{
			"items_created": 0,
			"parse_error":   parseError,
		}, nil
	}

	batchID := uuid.New()

	// ── Classify EVERY finding, before any filter ──
	// The retraction pass below asks "did this run report anything of type X for
	// this site", and that question must be answered from what the audit
	// OBSERVED, never from what this action FILED. A finding dropped by the
	// blocked check, the dedup check or a failed insert was still observed —
	// reading "not filed" as "fixed" is the one mistake that closes live
	// defects (WII-016's landmine, inherited).
	classified := make([]classifiedFinding, 0, len(findings))
	observedItemTypes := make(map[string]bool, len(findings))
	// Counted here, before the blocked/dedup/insert filters, for the same
	// reason observedItemTypes is: the question is what the audit OBSERVED
	// that this router could not place, not what happened to be filed.
	unroutedCategories := make(map[string]int)
	for _, f := range findings {
		c := classifyFinding(f, pages, siteID, auditSource)
		if recordOnly && c.HandlerAgent != "" {
			// The summary override below reuses the finding's description the
			// way the insert loop would, so a record row reads as a verdict
			// rather than as work.
			if c.Summary == "" {
				c.Summary = f.Description
			}
			c = recordOnlyFinding(c, auditSource)
			recorded++
		}
		classified = append(classified, c)
		observedItemTypes[c.ItemType] = true
		if c.ItemType == "capability_gap" {
			unroutedCategories[strings.ToLower(strings.TrimSpace(f.Category))]++
			logger.Warn("write_audit_findings: no route for finding category — filed as capability_gap",
				zap.String("category", f.Category),
				zap.String("audit_source", auditSource),
				zap.String("page", f.Page))
		}
	}

	if len(findings) == 0 {
		// A RECOGNISED empty findings list is the audit saying "nothing wrong
		// here", and that is exactly the observation retraction runs on — so
		// this early return must not skip it. An UNRECOGNISED reply reaches
		// here too and retractSilentAuditFindings refuses it on
		// shapeRecognised, which is why that flag is threaded rather than
		// inferred from this branch.
		retraction, rErr := retractSilentAuditFindings(ctx, params.DB, siteID, auditSource,
			batchID, observedItemTypes, shapeRecognised, logger)
		if rErr != nil {
			logger.Warn("Silence retraction failed", zap.Error(rErr))
		}
		logger.Warn("No valid findings extracted",
			zap.String("findings_field", findingsField),
			zap.String("findings_type", fmt.Sprintf("%T", findingsRaw)),
			zap.Bool("shape_recognised", shapeRecognised))
		out := map[string]interface{}{
			"items_created":    0,
			"reason":           "no valid findings",
			"findings_field":   findingsField,
			"findings_type":    fmt.Sprintf("%T", findingsRaw),
			"shape_recognised": shapeRecognised,
		}
		if retraction != nil {
			out["retraction"] = retraction
		}
		return out, nil
	}

	logger.Info("Parsed audit findings",
		zap.Int("count", len(findings)),
		zap.String("source", auditSource))

	// ── Load blocked items for filtering ──
	blockedKeys := make(map[string]bool)
	blockedRows, err := params.DB.QueryContext(ctx, `
		SELECT item_key FROM site_work_items
		WHERE site_id = $1 AND status = 'blocked' AND item_key IS NOT NULL
	`, siteID)
	if err == nil {
		for blockedRows.Next() {
			var key string
			if blockedRows.Scan(&key) == nil {
				blockedKeys[key] = true
			}
		}
		blockedRows.Close()
	}

	// ── Insert ── (classification already done, above, deliberately)
	created := 0
	skipped := 0
	skippedBlocked := 0
	skippedPendingProposal := 0
	copyEditBoundUnevaluated := 0
	parkedOwned := 0
	classificationStats := make(map[string]int)

	for i := range classified {
		classified, f := classified[i], findings[i]
		classificationStats[classified.ItemType]++

		logger.Info("Classified finding",
			zap.String("category", f.Category),
			zap.String("page", f.Page),
			zap.String("→item_type", classified.ItemType),
			zap.String("→handler", classified.HandlerAgent),
			zap.String("→page_name", classified.PageName),
			zap.Bool("→has_page_id", classified.PageID != nil))

		// Skip blocked
		if blockedKeys[classified.DedupKey] {
			skippedBlocked++
			continue
		}

		// Broader blocked check — scoped to THIS producer's own blocked rows.
		// The producer scope exists because item_type does not name its producer
		// (the retraction helper refuses the same co-filing trap structurally,
		// and this check used to be the one reader of the shared type with no
		// such guard): capability_gap is co-filed by the discovery/remit path,
		// and when PageID is nil the page clause is TRUE for every row, so an
		// unscoped check collapsed to "ANY blocked row of this type on the
		// site" — 18 discovery-filed blocked capability_gap rows on 14 sites
		// would have muted every unrouted audit category fleet-wide
		// (bugs_open/279, the bugfix_213 lane's contribution).
		var isBlocked bool
		params.DB.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND status = 'blocked'
				  AND item_type = $2
				  AND ($3::uuid IS NULL OR page_id = $3::uuid)
				  AND spec->>'audit_source' = $4
			)
		`, siteID, classified.ItemType, classified.PageID, auditSource).Scan(&isBlocked)

		if isBlocked {
			skippedBlocked++
			continue
		}

		// One un-reviewed proposal per page: the tone route's convergence bound
		// (see pendingCopyEditForPage). Fails OPEN on a query error, like the
		// neighbouring checks — a missed bound costs one extra parked proposal
		// (tone runs at roughly one a week); failing closed would silently mute
		// the route, which is the armed-but-inert shape.
		if classified.ItemType == "needs_copy_edit" {
			if classified.PageID == nil {
				// Structurally unreachable today — classifyFinding mints
				// needs_copy_edit only inside the pages[pageName] branch, with
				// &pageID (Rule 4), and TestNeedsCopyEditAlwaysCarriesPageID
				// holds that invariant. LOUD anyway (council 754dcffd round 1,
				// editquality gating objection): "any producer can file the
				// type" (CQ-030), so the day this becomes reachable the bound
				// must SAY it has gone blind — counted and reported, never a
				// silent no-op that files unbound exactly as before.
				copyEditBoundUnevaluated++
				logger.Warn("needs_copy_edit filed WITHOUT page_id — the pending-proposal bound cannot see it; filing unbound",
					zap.String("page", classified.PageName))
			} else if pending, pErr := pendingCopyEditForPage(ctx, params.DB, siteID, *classified.PageID); pErr != nil {
				logger.Warn("pending copy-edit bound check failed; filing anyway",
					zap.Error(pErr))
			} else if pending {
				skippedPendingProposal++
				logger.Info("needs_copy_edit withheld: the page already has an open copy-edit item or un-reviewed proposal",
					zap.String("page", classified.PageName))
				continue
			}
		}

		// Dedup check. 'deferred' is an OPEN status — idx_swi_dedup's exclusion
		// list is terminal-only — and capability_gap rows file as 'deferred';
		// without it here every later run of the same producer would re-insert
		// into the unique index and take a logged error instead of a clean skip.
		var exists bool
		params.DB.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM site_work_items
				WHERE site_id = $1 AND item_key = $2
				  AND status IN ('detected', 'triaged', 'claimed', 'blocked', 'deferred')
			)
		`, siteID, classified.DedupKey).Scan(&exists)

		if exists {
			skipped++
			continue
		}

		specJSON, _ := json.Marshal(classified.Spec)

		status := classified.Status
		if status == "" {
			status = "detected"
		}
		summary := classified.Summary
		if summary == "" {
			summary = f.Description
		}

		// Through the shared write seam, not a raw INSERT (bugs_open/333, owner
		// ruling 2026-08-25). This action was the last content-finding producer
		// walking round writeWorkItem, so its rows met none of the seam's
		// guards: three offer-analysis findings died wont_fix on the owned-page
		// refusal within hours of that door going live. The filters above keep
		// their jobs (skip bookkeeping, the producer-scoped mutes); the seam
		// now decides what the ROW becomes — parked for an owned page at a
		// declaring handler, deferred/branded by the anti-churn brake. (The
		// registration probe keys on triaged/approved/claimed, so rows born
		// 'detected' here still meet it at promotion, unchanged.)
		//
		// Log-and-continue per finding is unchanged — one bad finding must not
		// cost the batch — which is why each finding gets its OWN transaction:
		// a failed statement poisons a Postgres tx, so one shared tx would turn
		// the first failure into all of them.
		tx, txErr := params.DB.BeginTx(ctx, nil)
		if txErr != nil {
			logger.Warn("Failed to insert finding work item",
				zap.String("category", f.Category),
				zap.String("item_type", classified.ItemType),
				zap.Error(txErr))
			continue
		}
		w, err := writeWorkItem(ctx, tx, workItem{
			siteID:       siteID,
			source:       "discovery",
			pipeline:     "build",
			itemType:     classified.ItemType,
			severity:     classified.Severity,
			summary:      summary,
			spec:         string(specJSON),
			pageID:       classified.PageID,
			priority:     classified.Priority,
			handlerAgent: classified.HandlerAgent,
			status:       status,
			createdBy:    auditSource,
			itemKey:      classified.DedupKey,
			batchID:      batchID,
		}, dropOnConflict, logger)
		if err == nil {
			err = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
		if err != nil {
			logger.Warn("Failed to insert finding work item",
				zap.String("category", f.Category),
				zap.String("item_type", classified.ItemType),
				zap.Error(err))
			continue
		}
		if !w.Inserted {
			// An open row already holds this key: the seam-level dedup caught a
			// row the EXISTS pre-check above ran too early to see.
			skipped++
			continue
		}
		if w.OwnedPageParked {
			parkedOwned++
		}
		logger.Info("write_audit_findings: finding routed via the work-item seam",
			zap.String("item_key", classified.DedupKey),
			zap.String("requested_status", status),
			zap.Bool("owned_page_parked", w.OwnedPageParked),
			zap.Bool("born_blocked", w.BornBlocked),
			zap.Bool("deferred", w.Deferred))
		created++
	}

	// ── Retract what this run is silent about (bugs_open/213 D1 half two) ──
	// After the inserts, so a type this run DID report is already on record,
	// and inside its own transaction: a retraction failure must not lose the
	// filings, which are this action's primary job.
	retraction, rErr := retractSilentAuditFindings(ctx, params.DB, siteID, auditSource,
		batchID, observedItemTypes, shapeRecognised, logger)
	if rErr != nil {
		logger.Warn("Silence retraction failed", zap.Error(rErr))
	}

	logger.Info("WriteAuditFindingsAction: Complete",
		zap.Int("created", created),
		zap.Int("skipped_duplicates", skipped),
		zap.Int("skipped_blocked", skippedBlocked),
		zap.Int("skipped_pending_proposal", skippedPendingProposal),
		zap.Int("copy_edit_bound_unevaluated", copyEditBoundUnevaluated),
		zap.Int("parked_owned_page", parkedOwned),
		zap.Int("total_findings", len(findings)),
		zap.Any("classification_stats", classificationStats))

	out := map[string]interface{}{
		"items_created":         created,
		"items_skipped":         skipped,
		"items_skipped_blocked": skippedBlocked,
		"total_findings":        len(findings),
		"batch_id":              batchID.String(),
		"audit_source":          auditSource,
		"classification_stats":  classificationStats,
	}
	if recordOnly {
		// Present only in record mode so the historical output stays
		// byte-identical. `items_recorded_only` counts findings TRANSFORMED,
		// not rows inserted — a recorded finding can still be a dedup skip.
		out["filing_mode"] = filingModeRecord
		out["items_recorded_only"] = recorded
	}
	if parkedOwned > 0 {
		// Parked-not-routed is a different outcome from created-and-dispatchable;
		// present only when non-zero so the common case stays byte-identical.
		out["items_parked_owned_page"] = parkedOwned
	}
	if skippedPendingProposal > 0 {
		// Withheld behind a pending proposal is a different outcome from a dedup
		// skip; present only when non-zero so the common case stays byte-identical.
		out["items_skipped_pending_proposal"] = skippedPendingProposal
	}
	if copyEditBoundUnevaluated > 0 {
		// The bound went BLIND on these (needs_copy_edit with no page_id): they
		// filed unbound, and this receipt is the visibility. Non-zero here means
		// the Rule-4 invariant broke or a new producer files the type — read the
		// Warn lines and fix the producer, not this counter.
		out["items_copy_edit_bound_unevaluated"] = copyEditBoundUnevaluated
	}
	if len(unroutedCategories) > 0 {
		// The categories this router had no rule for — filed as capability_gap
		// rather than dispatched (bugs_open/279). Present only when non-empty
		// so the common case stays byte-identical.
		out["unrouted_categories"] = unroutedCategories
	}
	if retraction != nil {
		out["retraction"] = retraction
	}
	return out, nil
}

func getStringFromMap(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getIntFromMap(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	}
	return 0
}
