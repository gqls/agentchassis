// FILE: platform/orchestration/actions/discovery_checks/verifier_coverage_test.go
//
// The coverage guard for bugs_open/021 §INSTANCE 2.
//
// THE GAP. CompleteWorkItemAction consults a per-item_type verifier before
// stamping 'complete'. For a long time RegisterVerifier had been called exactly
// ONCE, for "empty_section", while the platform ran 69 item types — so 4,644 of
// 4,644 completions bar 5 were taken on the handler saga's own word. That is the
// same class as bugs_open/017 (a saga reporting success without touching the
// defect), one level up: 017 stops a saga that says it FAILED from being stamped
// complete; a verifier is what stops one that says it SUCCEEDED but did nothing.
//
// WHAT THIS TEST DOES. Every item_type the platform produces must be either
// verified or listed below with a category and a reason. Adding a new item type
// without doing one of those breaks the build. The gap becomes a decision on the
// record instead of an accident nobody can see.
//
// WHY THE LIST IS HAND-MAINTAINED, and the objection it answers. A council
// bug_historian seat warned that a coverage helper iterating the check registry
// would under-report, because an item_type created by a path that never
// registered a discovery check would be invisible to the guard. That is correct
// and it is worse than stated: the check registry keys on check NAME and each
// check's item types are string literals inside its Run method, so they are not
// enumerable at runtime AT ALL. And the highest-volume item types are exactly
// the ones the objection describes — cta_improvement (313 completions),
// needs_content_planning (387), spacing_fix (116) come from planner/auditor
// paths with no discovery check. So the denominator is sourced from the live
// database instead. Refresh query in RUNBOOK_work_item_completion_integrity.md;
// last refreshed 2026-07-20 (69 item types).
//
// THE HONEST WEAKNESS: this list can go stale between refreshes, so the guard
// catches a new item_type only once someone reruns the query. It is a ratchet,
// not a sensor. It still converts the silent default (no verifier, nobody
// notices) into a loud one (build fails until you classify it).

package discovery_checks

import (
	"fmt"
	"sort"
	"testing"
)

// verificationCategory records WHY an item type has no verifier.
type verificationCategory string

const (
	// catMechanical: the defect has a re-runnable predicate and the item carries
	// enough identity to locate it. These SHOULD get verifiers — this is the
	// actionable backlog, not an excuse list.
	catMechanical verificationCategory = "mechanical"

	// catJudgement: "resolved" is a quality opinion (is this copy better? is
	// this design good?). There is no predicate to re-run; verifying would mean
	// another LLM call, which is a different design decision with its own cost.
	catJudgement verificationCategory = "judgement"

	// catCreation: the item asks for something to be brought into existence.
	// Verifiable in principle by an existence check, but "exists" is a weak
	// proof of "was done well", so these are called out separately rather than
	// hidden inside mechanical.
	catCreation verificationCategory = "creation"

	// catNoTarget: the item is a site-level aggregate carrying no target id, so
	// the predicate cannot be scoped to what the item actually described.
	// VerifyTarget now supplies SiteID, so these are newly UNBLOCKED — they were
	// impossible before 2026-07-20 and are merely unwritten now.
	catNoTarget verificationCategory = "no_target"
)

type verificationGap struct {
	Category verificationCategory
	Why      string
}

// itemTypesWithoutVerifiers is the maintained denominator. Sourced from the live
// database 2026-07-20; see the file header for why it cannot be derived.
//
// EVIDENCE STATUS OF THE REASONS BELOW — the two kinds are not equally solid, and
// stating them in the same voice would be the error CLAUDE.md's marker rule exists
// to stop:
//   - The COUNTS ("all 47 carry page_id", "392 completions") are MEASURED, from
//     SQL over site_work_items on 2026-07-20. Refresh query in the RUNBOOK.
//   - The CATEGORIES and predicate descriptions are [INFERRED] from item-type and
//     check-file names, EXCEPT for empty_section, page_rerender/misdirected_cta and
//     hardcoded_section_colors, whose checks I actually opened and read.
//
// That distinction matters more than it looks: this session's own wrong call
// (WRONG_CALLS.md, 2026-07-20) was assuming an item type was verifiable from its
// detector's predicate without reading the HANDLER's remit. Every [INFERRED]
// category below could be wrong the same way — a "mechanical" entry may turn out
// to have a handler whose responsibility is narrower than its detector. So treat
// this map as a triage list to check, not a settled classification: before writing
// any verifier from it, read that item type's handler remit first.
var itemTypesWithoutVerifiers = map[string]verificationGap{
	// ---- mechanical: should get verifiers, in roughly volume order ----

	// page_rerender is the single biggest prize AND the cautionary tale: 1,849 of
	// ~4,644 completions, with page_id on 1,914 of 1,929 rows, so it looks like
	// the obvious first verifier. A working one was written and tested on
	// 2026-07-20 and then deliberately HELD, because re-running the
	// misdirected-CTA predicate over the whole page is STRICTER THAN THE HANDLER'S
	// REMIT: check_misdirected_cta's own header records that the rerender only
	// rewrites CTA url fields of components in the actions package's ctaFieldNames
	// set (hero, call-to-action, archetype-grid, archetype-combinations,
	// gauntlet-cta, content-block-about), and that a misdirect in any other
	// component — e.g. prose — is deliberately left for the next discovery pass to
	// re-detect and escalate via the two-strike rule to human review. A whole-page
	// verifier would therefore mark a correctly-handled rerender unresolved, burn
	// its attempts and strand it in 'failed', destroying that designed escalation
	// across 1,849 items. Writing it needs the rendered component mapped back to
	// its spec section's component.function so the verdict is scoped to what the
	// handler is actually responsible for. Do that first; do not re-derive the
	// trap. Design notes: work_item_completion_integrity/PLAN + NOTES 2026-07-20.
	"page_rerender": {catMechanical, "verifier written and held: a whole-page predicate is stricter than the handler's ctaFieldNames remit and would destroy the designed two-strike escalation — needs component→spec-function scoping first"},

	"hardcoded_section_colors":      {catNoTarget, "site-aggregate item; predicate is a site-wide component count. UNBLOCKED by VerifyTarget.SiteID — this is the next verifier to write"},
	"undeployed_asset":              {catNoTarget, "site-scoped asset sweep; no per-item target id. Unblocked by VerifyTarget.SiteID"},
	"needs_rerender":                {catMechanical, "43 of 142 carry component_id; predicate is the section-drift check"},
	"needs_component_regeneration":  {catMechanical, "12 of 57 carry component_id"},
	"phantom_internal_link":         {catMechanical, "all 65 carry page_id; predicate is check_phantom_internal_links"},
	"needs_internal_links":          {catMechanical, "all 49 carry page_id"},
	"link_resolution_rebuild":       {catMechanical, "all 26 carry page_id"},
	"needs_section_data":            {catMechanical, "all 69 carry component_id"},
	"deactivated_component":         {catMechanical, "all 41 carry component_id"},
	"empty_internal_href":           {catMechanical, "all 21 carry page_id"},
	"cta_names_unknown_destination": {catMechanical, "all 47 carry page_id; sibling finding of misdirected_cta"},
	"voice_tells":                   {catMechanical, "all 25 carry page_id; predicate is check_voice_tells"},
	"required_fields_missing":       {catMechanical, "carries page_id and component_id"},
	"dead_control":                  {catMechanical, "all 6 carry page_id"},
	"unresolved_cta":                {catMechanical, "66 items, none completed yet"},
	"image_source_unsatisfiable":    {catMechanical, "predicate is the imagery source check"},
	"image_url_404":                 {catMechanical, "deliberately NOT a verifier candidate: verification would add an outbound HTTP call to the completion path"},
	"section_source_drift":          {catMechanical, "predicate is check_section_source_drift"},
	"page_component_status_drift":   {catMechanical, "predicate is check_page_component_status_drift"},
	"generic_theme":                 {catMechanical, "site-scoped theme marker check"},
	"nav_drift":                     {catMechanical, "site-scoped nav comparison"},
	"missing_css":                   {catMechanical, "site-scoped asset existence"},
	"needs_sprite_css":              {catMechanical, "site-scoped asset existence"},
	"missing_news_page":             {catMechanical, "page existence by page_type"},
	"missing_news_sources":          {catMechanical, "site config existence"},
	"incomplete_page_group":         {catMechanical, "page-group completeness"},
	"orphan_blog_posts":             {catMechanical, "link-graph reachability"},
	"claims_unverified":             {catMechanical, "carries page_id"},
	"acceptance_run":                {catMechanical, "carries page_id and component_id"},
	"audit_tool":                    {catMechanical, "all 19 carry page_id and component_id"},
	"improve_tool":                  {catMechanical, "17 of 18 carry page_id and component_id"},
	"component_quality_scan":        {catMechanical, "quality score recomputation"},
	"missing_style_collection":      {catMechanical, "site-scoped existence check"},
	"unfulfilled_hero_variant":      {catMechanical, "imagery plan completeness"},
	"silent_failure":                {catMechanical, "meta item; predicate is the silent-check sweep"},

	// ---- creation: "make X exist" ----
	"needs_page":                 {catCreation, "page existence; 49 of 365 carry page_id"},
	"needs_content_page":         {catCreation, "page existence; 13 of 196 carry page_id"},
	"needs_imagery":              {catCreation, "image asset existence"},
	"needs_content_image":        {catCreation, "image asset existence"},
	"needs_hero_image":           {catCreation, "image asset existence"},
	"needs_logo":                 {catCreation, "asset existence"},
	"needs_brand_head_assets":    {catCreation, "asset existence"},
	"needs_new_component":        {catCreation, "component existence"},
	"needs_composition":          {catCreation, "composition existence"},
	"needs_tool_recreation":      {catCreation, "tool existence"},
	"add_tool":                   {catCreation, "tool existence"},
	"needs_blog_posts":           {catCreation, "post count"},
	"needs_site_plan":            {catCreation, "site plan existence"},
	"needs_experience_plan":      {catCreation, "plan existence"},
	"needs_new_layout_candidate": {catCreation, "layout candidate existence"},

	// ---- judgement: no mechanical predicate exists ----
	"needs_design_review":     {catJudgement, "392 completions — an LLM design opinion; nothing to re-run"},
	"needs_content_planning":  {catJudgement, "387 completions — a planning judgement"},
	"cta_improvement":         {catJudgement, "313 completions — 'is this CTA better' is an opinion"},
	"content_rewrite":         {catJudgement, "256 completions — prose quality"},
	"spacing_fix":             {catJudgement, "116 completions — visual judgement, no stored predicate"},
	"responsive_fix":          {catJudgement, "57 completions — visual judgement across breakpoints"},
	"needs_design":            {catJudgement, "design judgement"},
	"needs_strategy":          {catJudgement, "strategy judgement"},
	"needs_briefing":          {catJudgement, "briefing judgement"},
	"needs_diagnosis":         {catJudgement, "diagnosis loop output; graded by its own rubric"},
	"needs_domain_research":   {catJudgement, "research judgement"},
	"needs_vertical_research": {catJudgement, "research judgement"},
	"tone_shift":              {catJudgement, "tone opinion"},
	"audit_finding_audience":  {catJudgement, "audience opinion"},
	"capability_gap":          {catJudgement, "capability opinion"},
	"owned_page_review":       {catJudgement, "human review item"},
	"evaluate_tools":          {catJudgement, "tool evaluation opinion"},
}

func TestEveryItemTypeIsVerifiedOrAnAcknowledgedGap(t *testing.T) {
	verified := map[string]bool{}
	for _, itemType := range RegisteredVerifierItemTypes() {
		verified[itemType] = true
	}

	// A verifier and a gap entry for the same item type means the list is lying.
	for itemType := range verified {
		if gap, listed := itemTypesWithoutVerifiers[itemType]; listed {
			t.Errorf("item_type %q HAS a verifier but is still listed as a gap (%s: %s) — remove it from itemTypesWithoutVerifiers",
				itemType, gap.Category, gap.Why)
		}
	}

	// Every gap must carry a real reason, not a placeholder.
	for itemType, gap := range itemTypesWithoutVerifiers {
		if gap.Why == "" {
			t.Errorf("item_type %q is listed as a gap with no reason — say why it has no verifier", itemType)
		}
		switch gap.Category {
		case catMechanical, catJudgement, catCreation, catNoTarget:
		default:
			t.Errorf("item_type %q has unknown category %q", itemType, gap.Category)
		}
	}

	// The load-bearing assertion: nothing is silently unverified. This fires for
	// a NEW item type only after the denominator below is refreshed — see the
	// file header on why that is a ratchet rather than a sensor.
	for _, itemType := range liveItemTypes {
		if verified[itemType] {
			continue
		}
		if _, listed := itemTypesWithoutVerifiers[itemType]; !listed {
			t.Errorf("item_type %q has NO verifier and is NOT an acknowledged gap.\n"+
				"Either register a verifier (discovery_checks.RegisterVerifier) or add it to\n"+
				"itemTypesWithoutVerifiers with a category and the reason it cannot have one.\n"+
				"Silently completing on the handler's self-report is what bugs_open/021 filed.", itemType)
		}
	}
}

// TestVerifierCoverageIsReported prints the coverage picture so a reader sees
// the shape of the gap rather than only its violations. Never fails.
func TestVerifierCoverageIsReported(t *testing.T) {
	verified := map[string]bool{}
	for _, itemType := range RegisteredVerifierItemTypes() {
		verified[itemType] = true
	}

	byCat := map[verificationCategory][]string{}
	for itemType, gap := range itemTypesWithoutVerifiers {
		byCat[gap.Category] = append(byCat[gap.Category], itemType)
	}

	names := make([]string, 0, len(verified))
	for itemType := range verified {
		names = append(names, itemType)
	}
	sort.Strings(names)

	t.Logf("verified item types (%d): %v", len(names), names)
	for _, cat := range []verificationCategory{catMechanical, catNoTarget, catCreation, catJudgement} {
		list := byCat[cat]
		sort.Strings(list)
		t.Logf("%s gaps (%d): %s", cat, len(list), fmt.Sprint(list))
	}
	t.Logf("coverage: %d verified / %d classified item types",
		len(names), len(names)+len(itemTypesWithoutVerifiers))
}

// liveItemTypes is every item_type observed in site_work_items.
// Refreshed 2026-07-20 (69 rows):
//
//	SELECT DISTINCT item_type FROM site_work_items ORDER BY 1;
var liveItemTypes = []string{
	"acceptance_run", "add_tool", "audit_finding_audience", "audit_tool",
	"capability_gap", "claims_unverified", "component_quality_scan",
	"content_rewrite", "cta_improvement", "cta_names_unknown_destination",
	"dead_control", "deactivated_component", "empty_internal_href",
	"empty_section", "evaluate_tools", "generic_theme",
	"hardcoded_section_colors", "image_source_unsatisfiable", "image_url_404",
	"improve_tool", "incomplete_page_group", "link_resolution_rebuild",
	"missing_css", "missing_news_page", "missing_news_sources",
	"missing_style_collection", "nav_drift", "needs_blog_posts",
	"needs_brand_head_assets", "needs_briefing", "needs_component_regeneration",
	"needs_composition", "needs_content_image", "needs_content_page",
	"needs_content_planning", "needs_design", "needs_design_review",
	"needs_diagnosis", "needs_domain_research", "needs_experience_plan",
	"needs_hero_image", "needs_imagery", "needs_internal_links",
	"needs_logo", "needs_new_component", "needs_new_layout_candidate",
	"needs_page", "needs_rerender", "needs_section_data", "needs_site_plan",
	"needs_sprite_css", "needs_strategy", "needs_tool_recreation",
	"needs_vertical_research", "orphan_blog_posts", "owned_page_review",
	"page_component_status_drift", "page_rerender", "phantom_internal_link",
	"required_fields_missing", "responsive_fix", "section_source_drift",
	"silent_failure", "spacing_fix", "tone_shift", "undeployed_asset",
	"unfulfilled_hero_variant", "unresolved_cta", "voice_tells",
}
