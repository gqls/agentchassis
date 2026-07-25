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
// TWO HALVES, and why. A council bug_historian seat warned that a coverage helper
// iterating the check registry would under-report, because an item_type created by
// a path that never registered a discovery check would be invisible. Correct: the
// registry keys on check NAME (registry.go:102-104) and item types are literals
// inside each Run, so they are not enumerable AT RUNTIME; and the highest-volume
// types — cta_improvement (313 completions), needs_content_planning (387),
// spacing_fix (116) — come from planner/auditor paths with no check at all.
//
//  1. SENSOR — TestEveryCheckProducedItemTypeIsClassified scans this package's
//     SOURCE for `ItemType: "literal"`. Needs no refresh: a new check with a new
//     item type fails the build the moment it is written. Covers 56 types.
//  2. RATCHET — liveItemTypes, a hand-refreshed snapshot of item_type values seen
//     in the database, for the types produced OUTSIDE this package. Refresh query
//     in RUNBOOK_work_item_completion_integrity.md; last refreshed 2026-07-20.
//
// > **CORRECTED 2026-07-20, same day it was written.** The first version had only
// > half 2, and this header said so plainly — "a ratchet, not a sensor" — as though
// > naming the weakness discharged it. The council's deciding objection
// > (bug_historian) pointed out what that admission was actually conceding: a guard
// > against "the mechanism relies on someone remembering" that *itself* relies on
// > someone remembering reproduces the exact failure mode this workstream had just
// > finished correcting, one level down. I had also over-claimed "not enumerable at
// > all" — true of runtime, but a TEST CAN READ THE SOURCE, which I had not
// > considered.
// >
// > Adding the sensor immediately found **17 item types the snapshot could not
// > see** (slot_name_mismatch, forced_text_colors, backend_unreachable, …): all
// > declared by checks, none ever written to site_work_items, so NO refresh cadence
// > of the DB query would ever have caught them. Each would have completed
// > unverified the first time it fired, invisibly, because there is no data to look
// > at until it does. **Caught by:** the council objection, acted on rather than
// > answered in prose.

package discovery_checks

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
	//
	// VERIFIED IN THE HANDLER 2026-07-20 (previously this rested on the detector's
	// header comment, which is not evidence about the handler).
	// rerender_page_sections_action.go:283-296 gates the recompute:
	//     if reason == "cta_links_stale" {
	//         fn := comp.Function; if fn == "" { fn = s.slotName }
	//         if fields, isCTA := ctaFieldNames[fn]; isCTA { applyCTARecompute(...) }
	//     }
	// The remit is in fact NARROWER than the comment implies: applyCTARecompute
	// itself returns early, leaving the field untouched, when the stored url is
	// already a valid non-excluded non-self page ("authored link ... keep it") and
	// when the recomputed target is empty or not a valid page. So even inside the
	// ctaFieldNames set the handler deliberately declines to rewrite. A whole-page
	// verifier would have been MORE wrong than the hold assumed.
	"page_rerender": {catMechanical, "verifier written and held: a whole-page predicate is stricter than the handler's ctaFieldNames remit and would destroy the designed two-strike escalation — needs component→spec-function scoping first"},

	// hardcoded_section_colors got its verifier 2026-07-24 (bugs_open/021
	// INSTANCE 2) — check_hardcoded_section_colors.go, the first written against
	// the widened contract and the pattern to copy for the entry below: the
	// verdict is "the HANDLER's transform is at a fixed point", never the
	// detector's broader predicate.
	"undeployed_asset":              {catNoTarget, "site-scoped asset sweep; no per-item target id. Unblocked by VerifyTarget.SiteID — next in line; read the handler's remit first"},
	"needs_rerender":                {catMechanical, "43 of 142 carry component_id; predicate is the section-drift check"},
	"needs_component_regeneration":  {catMechanical, "12 of 57 carry component_id"},
	"phantom_internal_link":         {catMechanical, "all 65 carry page_id; predicate is check_phantom_internal_links"},
	"unbuilt_internal_link":         {catMechanical, "bugs_open/049 mechanism 2; carries the TARGET page_id (never-deployed page the link 404s at); predicate is check_phantom_internal_links deployed_at IS NULL"},
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
	"backend_entry_orphaned":        {catMechanical, "live-probe check (GET → 405); deliberately NOT a verifier candidate — verification would put an outbound HTTP probe in the completion path, same reason as image_url_404"},
	"contact_form_undeliverable":    {catMechanical, "needs_human_review queue — since cc2cff79b only the address-less branch files this type (resolvable sites route to page_rerender). Predicate re-runs on DEPLOYED html, which re-renders only after a fix lands, so a completion-time re-check would false-fail during the render lag; resolve delivery timing before writing one"},
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

	// ---- FOUND BY THE SOURCE SCAN, not by the live-DB snapshot ----
	//
	// These 17 are declared by checks but have never appeared in site_work_items,
	// so the hand-refreshed liveItemTypes list below could not see them — no
	// possible refresh cadence would have. They are the proof that
	// TestEveryCheckProducedItemTypeIsClassified is worth having: each is an
	// item type that would have completed unverified the FIRST time it ever
	// fired, silently, with nobody able to notice from the data because there is
	// no data yet.
	//
	// Categories are [INFERRED] from check names — I have not opened these checks,
	// and per this session's own wrong call that means the classification could be
	// wrong in the direction that matters (a handler whose remit is narrower than
	// its detector). Read the handler before writing any verifier from this block.
	"slot_name_mismatch":      {catMechanical, "[INFERRED] check_component_standards; never observed live"},
	"unlinked_site_component": {catMechanical, "[INFERRED] check_component_standards; never observed live"},
	"stacked_nav":             {catMechanical, "[INFERRED] check_component_standards; never observed live"},
	"missing_logo_in_header":  {catMechanical, "[INFERRED] check_component_standards; never observed live"},
	"broken_template_slots":   {catMechanical, "[INFERRED] check_component_standards; never observed live"},
	"missing_site_metadata":   {catMechanical, "[INFERRED] check_component_standards; never observed live"},
	"unwanted_nav_element":    {catMechanical, "[INFERRED] check_component_standards; never observed live"},
	"stale_news_section":      {catMechanical, "[INFERRED] check_news_feed; never observed live"},
	"missing_news_section":    {catMechanical, "[INFERRED] check_news_feed; never observed live"},
	"all_sources_erroring":    {catMechanical, "[INFERRED] check_news_feed; never observed live"},

	// model_directory_pipeline Phase D (2026-07-22) — not [INFERRED]: I wrote
	// and read both checks myself. Same mechanical shape as their news-feed
	// siblings (missing_news_section/missing_news_page): existence checks
	// (page_component by function / page by page_type), handled by the same
	// content-gap-planner. Not yet enabled (migration 194 pending an image
	// roll), so genuinely never observed live, not just unrefreshed.
	"missing_model_directory_section": {catMechanical, "page_component existence by function (model-directory); handler is content-gap-planner, same as missing_news_section; enabled 2026-07-24 (migration 194), has fired and been serviced on ai-agent-orchestration.com"},
	"missing_model_directory_page":    {catMechanical, "page existence by page_type (model-directory); handler is content-gap-planner, same as missing_news_page; enabled 2026-07-24 (migration 194), has fired and been serviced on ai-agent-orchestration.com"},

	// model_directory_pipeline Phase E (2026-07-25) — the SAME two checks as
	// the four lines above, instantiated from a different profile
	// (check_directory.go's directoryCheckProfiles). Not a new mechanism and
	// not a new shape: existence checks against page_components.function /
	// pages.page_type, handled by content-gap-planner. Registered but NOT in
	// completeness-discovery's checks array yet, so never observed live —
	// which is a statement about the config, not a claim that they work.
	"missing_adoption_tracker_section": {catMechanical, "page_component existence by function (adoption-tracker); profile sibling of missing_model_directory_section; registered, not yet in the discovery checks array"},
	"missing_adoption_tracker_page":    {catMechanical, "page existence by page_type (adoption-tracker); profile sibling of missing_model_directory_page; registered, not yet in the discovery checks array"},
	"missing_protocol_tracker_section": {catMechanical, "page_component existence by function (protocol-tracker); profile sibling of missing_model_directory_section; registered, not yet in the discovery checks array"},
	"missing_protocol_tracker_page":    {catMechanical, "page existence by page_type (protocol-tracker); profile sibling of missing_model_directory_page; registered, not yet in the discovery checks array"},
	"unrendered_template":              {catMechanical, "[INFERRED] check_integrity; never observed live"},
	"cross_site_contamination":         {catMechanical, "[INFERRED] check_integrity; never observed live"},
	"forced_text_colors":               {catMechanical, "[INFERRED] check_forced_text_colors — sibling of bugs_open/017's action; never observed live"},
	"duplicate_palette":                {catMechanical, "[INFERRED] check_duplicate_palette; never observed live"},
	"placeholder_contact":              {catMechanical, "[INFERRED] check_placeholder_contact; never observed live"},
	"broken_nav_links":                 {catMechanical, "[INFERRED] check_broken_nav_links; never observed live"},
	"backend_unreachable":              {catMechanical, "[INFERRED] check_backend_unreachable, which already SELF-CLEARS on a live health probe — a verifier may be redundant here; check before writing one"},

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

	// ---- first observed in site_work_items by the 2026-07-24 refresh ----
	"needs_human_review":            {catJudgement, "generic HITL checkpoint item (checkpoint_for_review_action.go); resolution IS the human's ruling"},
	"directory_citation_unverified": {catJudgement, "human ruling on citation candidates that failed live verification (directory_claims.go); nothing mechanical to re-run"},
	"audit_finding_brief_fidelity":  {catJudgement, "audit_finding_* is computed in write_audit_findings_action.go; auditor opinion, sibling of audit_finding_audience"},
	"section_edit":                  {catJudgement, "[INFERRED] owner-directed section edit, created outside Go (agent workflow config — no ItemType literal anywhere in platform/); 'applied as intended' has no stored predicate"},
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

// itemTypeLiteralRe matches `ItemType: "foo"` in a check's WorkItemSpec.
var itemTypeLiteralRe = regexp.MustCompile(`ItemType:\s*"([a-z_0-9]+)"`)

// itemTypeComputedRe matches `ItemType: <expression>` — a type this test cannot
// read from source. Each must be acknowledged in computedItemTypeSites.
var itemTypeComputedRe = regexp.MustCompile(`ItemType:\s*([^"\s][^,\n]*)`)

// computedItemTypeSites are the check sites that build item_type at RUNTIME, so
// no source scan can enumerate what they produce. Listed explicitly so a NEW one
// is a build failure rather than a silent hole in the sensor below.
var computedItemTypeSites = map[string]string{
	"check_image_url_404.go":            "mapping.itemType — per-surface mapping table",
	"check_phantom_internal_links.go":   "f.IssueType — finding carries its own type",
	"check_placeholder_image_in_use.go": "mapping.itemType — per-surface mapping table",
	"check_unfulfilled_image_prompt.go": "itemType — computed from the prompt's surface",
	// p.SectionItemType / p.PageItemType, one profile per register kind:
	// missing_{model_directory,adoption_tracker,protocol_tracker}_{section,page}.
	// Note the scan is NOT blind here by luck — directoryCheckProfiles spells
	// each type as a literal in a field whose name ends "ItemType:", so
	// itemTypeLiteralRe picks all six up anyway and they are classified below.
	// Adding a seventh profile therefore still fails this guard until it is
	// classified, which is the behaviour we want.
	"check_directory.go": "p.SectionItemType / p.PageItemType — one profile per register kind (model, company, protocol)",
}

// TestEveryCheckProducedItemTypeIsClassified is the SENSOR half of the guard.
//
// Answers the council's deciding objection (bug_historian, 2026-07-20): the
// liveItemTypes list below is a hand-refreshed snapshot, so on its own this guard
// "relies on someone remembering" — reproducing, one level down, the exact failure
// mode that 021 and this workstream's own correction identify as the WRONG
// diagnosis for the verifier gap. Naming that in a comment was not fixing it.
//
// So: scan the package SOURCE for ItemType literals. I had claimed item types are
// "not enumerable at runtime at all" — true (Register keys on check.Name(),
// registry.go:102-104, and the types are literals inside each Run), but a TEST can
// read the files. 62 of them are source-visible today and need no refresh; adding a
// check with a new literal item type now fails the build the moment it is written.
//
// Residual, stated rather than hidden: this cannot see the 4 computed sites above,
// nor item types created outside discovery checks (cta_improvement,
// needs_content_planning, spacing_fix — the highest-volume ones). Those still rely
// on liveItemTypes. So the guard is now a sensor for the majority and a ratchet for
// the rest, which is strictly better than a ratchet for all of it.
func TestEveryCheckProducedItemTypeIsClassified(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob: %v", err)
	}

	verified := map[string]bool{}
	for _, itemType := range RegisteredVerifierItemTypes() {
		verified[itemType] = true
	}

	seen := map[string]string{} // item type -> file it came from
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("read %s: %v", f, err)
		}
		for _, m := range itemTypeLiteralRe.FindAllStringSubmatch(string(src), -1) {
			seen[m[1]] = f
		}
		// A computed ItemType is invisible to the scan — require it be declared.
		for _, m := range itemTypeComputedRe.FindAllStringSubmatch(string(src), -1) {
			expr := strings.TrimSpace(m[1])
			if expr == "" {
				continue
			}
			if _, declared := computedItemTypeSites[f]; !declared {
				t.Errorf("%s computes ItemType at runtime (%s) but is not in computedItemTypeSites.\n"+
					"The source scan cannot see what it produces, so add it there with what it emits —\n"+
					"otherwise this guard has a silent hole exactly where it claims coverage.", f, expr)
			}
		}
	}

	if len(seen) == 0 {
		t.Fatal("scanned 0 item types — the regex or the layout changed, and this guard is now vacuous")
	}

	for itemType, file := range seen {
		if verified[itemType] {
			continue
		}
		if _, listed := itemTypesWithoutVerifiers[itemType]; listed {
			continue
		}
		t.Errorf("item_type %q (produced by %s) has NO verifier and is NOT an acknowledged gap.\n"+
			"Register a verifier, or add it to itemTypesWithoutVerifiers with a category and reason.\n"+
			"Detected by source scan — no list refresh was needed, and none will excuse it.", itemType, file)
	}

	t.Logf("source scan: %d check-produced item types across %d files, %d computed sites acknowledged",
		len(seen), len(files), len(computedItemTypeSites))
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
// Refreshed 2026-07-24 (66 live rows, UNIONed with the previous list):
//
//	SELECT DISTINCT item_type FROM site_work_items ORDER BY 1;
//
// Refresh by UNION, never replacement: site_work_items rows get pruned, so a
// type can vanish from the live query while its producer is still deployed
// (10 of the 2026-07-20 types had no rows left by 07-24 — e.g. silent_failure,
// tone_shift). Dropping one would drop its protection.
var liveItemTypes = []string{
	"acceptance_run", "add_tool", "audit_finding_audience",
	"audit_finding_brief_fidelity", "audit_tool",
	"capability_gap", "claims_unverified", "component_quality_scan",
	"contact_form_undeliverable",
	"content_rewrite", "cta_improvement", "cta_names_unknown_destination",
	"dead_control", "deactivated_component", "directory_citation_unverified",
	"empty_internal_href",
	"empty_section", "evaluate_tools", "generic_theme",
	"hardcoded_section_colors", "image_source_unsatisfiable", "image_url_404",
	"improve_tool", "incomplete_page_group", "link_resolution_rebuild",
	"missing_css", "missing_model_directory_page",
	"missing_model_directory_section", "missing_news_page",
	"missing_news_sources",
	"missing_style_collection", "nav_drift", "needs_blog_posts",
	"needs_brand_head_assets", "needs_briefing", "needs_component_regeneration",
	"needs_composition", "needs_content_image", "needs_content_page",
	"needs_content_planning", "needs_design", "needs_design_review",
	"needs_diagnosis", "needs_domain_research", "needs_experience_plan",
	"needs_hero_image", "needs_human_review", "needs_imagery",
	"needs_internal_links",
	"needs_logo", "needs_new_component", "needs_new_layout_candidate",
	"needs_page", "needs_rerender", "needs_section_data", "needs_site_plan",
	"needs_sprite_css", "needs_strategy", "needs_tool_recreation",
	"needs_vertical_research", "orphan_blog_posts", "owned_page_review",
	"page_component_status_drift", "page_rerender", "phantom_internal_link",
	"required_fields_missing", "responsive_fix", "section_edit",
	"section_source_drift",
	"silent_failure", "spacing_fix", "tone_shift", "truncated_component",
	"undeployed_asset",
	"unfulfilled_hero_variant", "unresolved_cta", "voice_tells",
}
