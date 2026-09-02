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
// ⚠ WHAT "VERIFIED" DOES NOT MEAN HERE — read this before working the backlog
// below (bugs_open/375, added 2026-08-24).
//
// "Has a registered verifier" is not "is verified". The verifier is consulted by
// the code that stamps the row, and there are THREE such writers, of which only
// one has always asked:
//
//   - CompleteWorkItemAction — consults it (complete_work_item_verification.go).
//   - UpdateWorkItemStatusAction — consults it only when that STEP's config sets
//     `verify_before_complete: true`. Unarmed, it completes and records the bypass
//     at result._verification (status "verifier_not_consulted"). Opt-in per step
//     because register CQ-023 shows one router's arms want different answers.
//   - the claimed-item-timeout sweep — consults NEITHER gate; it writes the row
//     directly, and is held off a type only by livespec.ClaimedItemTimeoutExclusions
//     (bugs_closed/317, lockstep-guarded in claim_timeout_exclusion_lockstep_test.go).
//
// So registering a verifier turns THIS list green, and protects the type only on
// the paths that ask. [MEASURED 2026-08-24, live DB UNION ARCHIVE] 4 of 200 live
// agent definitions complete through update_work_item_status across 6 arms, over
// SEVEN item types — needs_imagery, required_fields_missing, needs_hero_image,
// unfulfilled_hero_variant, needs_logo, image_url_404,
// image_source_unsatisfiable; 578 completions all-history. NONE of the seven has
// a verifier today, so the gap is LATENT rather than active.
// ⚠ Read that census over `site_work_items` UNION `site_work_items_archive`: the
// live table is a ROLLING WINDOW, and over it alone this said FIVE types and 134
// completions — two types had been completed entirely into the archive. Two of them
// (required_fields_missing :237, image_source_unsatisfiable :240) sit in the
// catMechanical backlog below, which means the first person to work that backlog
// is the person this bites. Census + controls:
// docs024_key_docs_latest/bugfix_375_completion_verifier_gap/RUNBOOK_completion_verifier_gap.md.
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
	// itself declines to rewrite when the stored url is already a valid non-self
	// page, and when the recomputed target is empty or not a valid page. So even
	// inside the ctaFieldNames set the handler deliberately leaves fields alone.
	// A whole-page verifier would have been MORE wrong than the hold assumed.
	//
	// bugs_open/248 (cta_recompute_clobbers_authored_contact_links) WIDENED that
	// decline by one case, so this description no longer says "non-excluded": a
	// stored, valid destination in a utility area (contact/about/...) is now KEPT
	// rather than overwritten, because no resolver path can produce one and it is
	// therefore authored. Anyone writing the held verifier must model that too, or
	// it will report an authored contact link as an unresolved misdirect.
	"page_rerender": {catMechanical, "verifier written and held: a whole-page predicate is stricter than the handler's ctaFieldNames remit and would destroy the designed two-strike escalation — needs component→spec-function scoping first"},

	// hardcoded_section_colors got its verifier 2026-07-24 (bugs_open/021
	// INSTANCE 2) — check_hardcoded_section_colors.go, the first written against
	// the widened contract and the pattern to copy for the entry below: the
	// verdict is "the HANDLER's transform is at a fixed point", never the
	// detector's broader predicate.
	"undeployed_asset": {catNoTarget, "site-scoped asset sweep; no per-item target id. Unblocked by VerifyTarget.SiteID — next in line; read the handler's remit first"},
	// contrast_failure's reason was REWRITTEN 2026-08-12, and the exemption kept.
	// What stood here was: "the dedup key plus the NEXT render audit is the
	// verifier, and the two-strike rule catches a persistent pairing". Three
	// things were wrong with it, and they are worth keeping because the same
	// sentence is still live one entry below:
	//
	//   - It is the argument RFC_017 REFUTED on 2026-08-08, six days after this
	//     line was authored (f2a222964) and never revisited. Re-detection is not
	//     verification: it grades the NEXT run's finding, not THIS row.
	//   - It infers resolution from ABSENCE, which CheckResult.Resolved's own
	//     contract forbids in writing — a pairing stops being re-filed when the
	//     page is fixed, when it is unreachable, and when a cap dropped it.
	//   - It had never once been exercised: 0 of 226 rows dispatched, completed
	//     or re-detected, max(attempt_count)=0, measured 2026-08-12. An
	//     unregistered type completes UNTOUCHED, so the claim was not merely
	//     unproven, it described a mechanism that does not run.
	//
	// It is now true, by construction rather than by hope: since 2026-08-12
	// write_render_audit_findings RETRACTS these rows on a positive
	// re-observation — scoped to summary.pages_audited, the pages the adapter
	// successfully measured — inside the transaction that files the run's
	// findings. That is the same measurement a verifier would fetch, taken on
	// the discovery path where the browser probe is already precedented.
	"contrast_failure": {catMechanical, "minted by write_render_audit_findings (actions pkg, outside this sensor's glob). Deliberately NOT a verifier candidate — verification needs a browser, i.e. an outbound probe on the completion path, the same standing objection as image_url_404, backend_entry_orphaned and asset_reference_404. It does not NEED one: since 2026-08-12 the producer retracts its own findings on a positive re-observation, scoped to the pages the audit successfully measured (summary.pages_audited) and never to the pages it merely requested, stamping result.resolved_by/reason as the evidence. asset_reference_404's posture, for its reasons"},
	// dark_section_audit: minted by write_audit_findings (actions pkg, outside this
	// sensor's glob) — bugs_open/213 split it OUT of hardcoded_section_colors, whose
	// verifier grades a different predicate entirely (the discovery check's site
	// aggregate, not a named section's defect). Classified on the way IN rather than
	// waiting for the ratchet refresh, the mistyped_deployed_page precedent: the
	// sensor cannot see a type minted outside this package, so a new one is invisible
	// until someone remembers, which is the failure mode this file exists to stop.
	// Its pass condition is spec.acceptance_test — free text over computed styles —
	// so verification needs a browser. Candidate verifier: criteria_check (RFC_002)
	// over acceptance_test, which nothing on the completion path reads today.
	//
	// ⚠ ITS REASON IS THE ONE CONTRAST_FAILURE JUST RETIRED, AND IT NO LONGER HAS
	// COMPANY (2026-08-12). This entry used to say "same posture as contrast_failure
	// above" and inherit its argument; contrast_failure has since moved to
	// retraction-on-positive-re-observation and this type has NOT, so the
	// re-detection claim now stands here alone — including the part RFC_017 refuted
	// on 2026-08-08 (re-detection grades the next run's finding, not this row) and
	// the part that infers resolution from absence. The remedy that worked for
	// contrast_failure is available in principle — write_audit_findings could
	// retract on the design audit's own positive re-observation — but whether
	// dark_section_audit should be verified at all is bugs_open/213's call or the
	// owner's, NOT ESTABLISHED here, and this note deliberately decides nothing.
	// It records that the citation moved out from under it.
	"dark_section_audit":           {catMechanical, "minted by write_audit_findings (actions pkg, outside this sensor's glob); verification needs a browser — pass condition is spec.acceptance_test free text over computed styles; the NEXT design audit plus the dedup key is the re-detection and two-strike escalates a persistent pairing — an absence-inference RFC_017 refuted, and since 2026-08-12 NO LONGER contrast_failure's posture, which now retracts on positive re-observation; candidate verifier is criteria_check (RFC_002) over acceptance_test"},
	"needs_rerender":               {catMechanical, "43 of 142 carry component_id; predicate is the section-drift check"},
	"needs_component_regeneration": {catMechanical, "12 of 57 carry component_id"},
	"phantom_internal_link":        {catMechanical, "all 65 carry page_id; predicate is check_phantom_internal_links"},
	// unbuilt_internal_link: VERIFIER WRITTEN 2026-08-08 (bugs_open/220) —
	// VerifyUnbuiltInternalLinkResolved in check_phantom_internal_links.go. Entry
	// removed from this map rather than edited, per the literal_markdown precedent.
	// The verdict is judged by the SAME shared predicate the detector minted the
	// finding by (datahelpers.NeverDeployedPagePredicate), with link-removal as the
	// accepted alternative remedy — see the verifier's own header for the remit
	// argument, and bugs_open/220 for why completion had to stop trusting the saga.
	"needs_internal_links":          {catMechanical, "all 49 carry page_id"},
	"link_resolution_rebuild":       {catMechanical, "all 26 carry page_id"},
	"needs_section_data":            {catMechanical, "all 69 carry component_id"},
	"deactivated_component":         {catMechanical, "all 41 carry component_id"},
	"empty_internal_href":           {catMechanical, "all 21 carry page_id"},
	"cta_names_unknown_destination": {catMechanical, "all 47 carry page_id; sibling finding of misdirected_cta"},
	"cta_names_nonpage_destination": {catMechanical, "review-only, no handler (bugs_open/299 round 1): repair deliberately withheld until the non-page keep branches are proven live — a cta_links_stale item here IS the LANDMINES recompute clobber; predicate is classifyNonPageAnchor, re-runnable when a verifier is wanted"},
	"cta_tel_malformed":             {catMechanical, "review-only, no handler: predicate is datahelpers.NormalizeTelHref, trivially re-runnable; the unambiguous forms self-heal via the keep branches, the refused residue (+440… collapsed trunk) NEEDS a human, so auto-completion would be wrong by design (bugs_open/299)"},
	"voice_tells":                   {catMechanical, "all 25 carry page_id; predicate is check_voice_tells"},
	// literal_markdown: VERIFIER WRITTEN 2026-08-06 (bugs_open/201 symptom 2) —
	// VerifyLiteralMarkdownResolved in check_literal_markdown.go. Entry removed from
	// this map rather than edited, which is what the guard above demands.
	// The deferral note that used to sit here said "write it against the REPAIRING
	// agent's rewrite remit, not the detector's predicate (page_rerender's trap)". That
	// instruction was followed: the handler is page-build-handler, whose build_pages_loop
	// rewrites ALL of the page's spec sections, so whole-page scope IS its remit and the
	// verifier is not stricter than the thing it judges.
	// required_fields_missing REMOVED 2026-08-26: it now HAS a verifier
	// (verify_required_fields_missing.go, bugs_open/375 / WII-032). It was the first entry
	// taken off this catMechanical backlog — the list that calls itself "the actionable
	// backlog, not an excuse list".
	"dead_control":               {catMechanical, "all 6 carry page_id"},
	"unresolved_cta":             {catMechanical, "66 items, none completed yet"},
	"image_source_unsatisfiable": {catMechanical, "predicate is the imagery source check"},
	"image_url_404":              {catMechanical, "deliberately NOT a verifier candidate: verification would add an outbound HTTP call to the completion path"},
	"backend_entry_orphaned":     {catMechanical, "live-probe check (GET → 405); deliberately NOT a verifier candidate — verification would put an outbound HTTP probe in the completion path, same reason as image_url_404"},
	// page_content_divergence: bugs_open/315 candidate 4 / PLAN D5. Classified on
	// the way IN rather than waiting for a ratchet refresh — the mistyped_deployed_page
	// and dark_section_audit precedent.
	"page_content_divergence":     {catMechanical, "bugs_open/315 / RFC_038; live-probe check (GET the page with a cache-buster, sha256 the body, compare against pages.content_hash) and carries page_id on every finding. Deliberately NOT a verifier candidate — verification would put an outbound HTTP probe in the completion path, the standing objection shared with asset_reference_404, image_url_404 and backend_entry_orphaned. It does not NEED one, and this is the posture contrast_failure moved to on 2026-08-12: the check retracts its own findings through CheckResult.Resolved on a positive re-observation (the served bytes now hash to the stored fingerprint), which is the SAME comparison a verifier would make, taken on the discovery path where the probe is already precedented. Note the item is flag-only by design (no handler agent, D5), so nothing on the completion path claims it done in the first place"},
	"archived_page_still_serving": {catMechanical, "bugs_open/359; live-probe check (GET the RETIRED page's own url, cache-busted, and file only a 2xx that is confirmed twice AND lands at the page's own final url); carries page_id on every finding; single-producer by design (the co-dedup landmine). Deliberately NOT a verifier candidate — verification would put an outbound HTTP probe in the completion path, the standing objection shared with asset_reference_404, image_url_404, page_content_divergence and backend_entry_orphaned. It does not NEED one on two grounds: the item is flag-only (handler_agent \"\" — the remedy is retract_page_deployment OR un-archiving, a judgement about intent that no generator can make), so nothing on the completion path claims it done; and it self-clears through CheckResult.Resolved on any of three POSITIVE observations — a twice-confirmed 404/410, a 2xx that lands somewhere other than the page's own url (a legitimate redirected retirement), or pages.status read back as 'active'. ⚠ THE PROPERTY THAT MAKES ITS ZERO READABLE, and the one place it differs from every sibling above: this check's FINDING is a 200, so its blinded state reports ZERO — which is exactly what a healthy estate reports, the opposite profile from asset_reference_404's under-reporting. It therefore runs two per-domain instrument controls (an invented url that must not answer 2xx, an active sibling that must) BEFORE judging anything, and RETURNS AN ERROR when either fails — filing nothing and, because the runner skips Resolved on an errored check, retracting nothing either."},
	"asset_reference_404":         {catMechanical, "bugs_open/084; live-probe check (GET a referenced <script src>/stylesheet, 404|410 only); carries page_id on the page surface, nil on chrome. Deliberately NOT a verifier candidate — verification would put an outbound HTTP probe in the completion path, same reason as image_url_404 and backend_entry_orphaned. It does not NEED one: the check retracts its own findings through CheckResult.Resolved on a positive 2xx/3xx re-observation, which is the same information a verifier would fetch, taken on the discovery path where the probe is already precedented"},
	// stylesheet_gutted: bugs_open/198 (three clobber waves — relojistas 08-04,
	// six sites 08-17/18, remortgagecalculator.uk + loanzy.uk 08-21) and
	// bugs_open/211 (the subtler shape: a 26KB stylesheet missing only the
	// renderer's alias block, --color-heading defined zero times). Classified on
	// the way IN, before the check is enabled on any live agent, per this file's
	// own precedent for the five structural checks directly below.
	"stylesheet_gutted":           {catMechanical, "bugs_open/198 + /211; live-probe check (GET every same-host <link rel=stylesheet> the deployed corpus references, then compare the custom properties DEFINED there against those REFERENCED without a fallback by page_components/site_components CSS, plus css_snippets). Deliberately NOT a verifier candidate — verification would put an outbound HTTP probe in the completion path, the standing objection shared with asset_reference_404, image_url_404, page_content_divergence and backend_entry_orphaned. It does not NEED one on two independent grounds: the item is flag-only by design (no handler agent — the repair is restore-from-git or a webdesign-agent run, both judgements, and auto-routing at webdesign-agent would re-roll the palette), so nothing on the completion path claims it done; and the check self-clears through CheckResult.Resolved on a positive re-observation (every referenced property defined, every stylesheet 2xx), which is the SAME comparison a verifier would make, taken on the discovery path where the probe is already precedented. Note it declines to judge — filing AND retracting nothing — whenever a same-host stylesheet fails to fetch or returns non-2xx, so a blinded run can never be mistaken for a healthy one"},
	"contact_form_undeliverable":  {catMechanical, "needs_human_review queue — since cc2cff79b only the address-less branch files this type (resolvable sites route to page_rerender). Predicate re-runs on DEPLOYED html, which re-renders only after a fix lands, so a completion-time re-check would false-fail during the render lag; resolve delivery timing before writing one"},
	"section_source_drift":        {catMechanical, "predicate is check_section_source_drift"},
	"page_component_status_drift": {catMechanical, "predicate is check_page_component_status_drift"},
	"generic_theme":               {catMechanical, "site-scoped theme marker check"},
	"nav_drift":                   {catMechanical, "site-scoped nav comparison"},
	"missing_css":                 {catMechanical, "site-scoped asset existence"},
	"needs_sprite_css":            {catMechanical, "site-scoped asset existence"},
	"missing_news_page":           {catMechanical, "page existence by page_type"},
	"missing_news_sources":        {catMechanical, "site config existence"},
	"incomplete_page_group":       {catMechanical, "page-group completeness"},
	"orphan_blog_posts":           {catMechanical, "link-graph reachability"},
	"claims_unverified":           {catMechanical, "carries page_id"},
	"acceptance_run":              {catMechanical, "carries page_id and component_id"},
	"audit_tool":                  {catMechanical, "all 19 carry page_id and component_id"},
	"improve_tool":                {catMechanical, "17 of 18 carry page_id and component_id"},
	// ported_tool_fix (bugs_open/281): a PORTED tool's structural/acceptance
	// finding, filed by check_tool_health and check_tool_acceptance with NO
	// handler (needs_human_review) — the same posture as orphan_element_refs.
	// The predicate is re-runnable (auditTool / the criteria evaluator over the
	// instance's rendered_html, keyed component_id + page_id), but resolution is
	// a human's until ported tools have PLANs and a per-instance repair path;
	// graduation is decomposition to a real fork, which then draws ordinary
	// improve_tool items.
	"ported_tool_fix":          {catMechanical, "carries page_id and component_id; instance-keyed, human-routed"},
	"component_quality_scan":   {catMechanical, "quality score recomputation"},
	"missing_style_collection": {catMechanical, "site-scoped existence check"},
	"unfulfilled_hero_variant": {catMechanical, "imagery plan completeness"},
	"silent_failure":           {catMechanical, "meta item; predicate is the silent-check sweep"},

	// The five check_site_structural_validity.go checks (generalising
	// bugs_open/251's verify_site.py --live catch fleet-wide). Classified on the
	// way IN, before the first row exists, per this file's own 07-20 correction —
	// none is enabled on a live discovery agent yet (that wiring is deliberate
	// follow-up work, not this commit's job).
	//
	// All five are deliberately NOT verifier candidates, same posture and same
	// reason as asset_reference_404/site_unreachable directly above/below in this
	// map: each is itself a live-probe check (a GET of the served page, or for
	// sitemap_entry_dead_live a GET of sitemap.xml, and for dead_internal_link_live/
	// sitemap_entry_dead_live a second GET of the link target), so a completion-time
	// verifier would re-run an outbound HTTP call on the completion path — the
	// exact thing those two entries decline. Each SELF-CLEARS through
	// CheckResult.Resolved on a positive re-observation on its next run, which is
	// the same information a verifier would have had to fetch.
	"dead_internal_link_live": {catMechanical, "bugs_open/251-adjacent; live-probe check (GET a referenced <a href> target, 404|410 only, confirm-before-file — same discipline as asset_reference_404); carries page_id of the FIRST page observed to link there. Deliberately NOT a verifier candidate, same reason as asset_reference_404/backend_entry_orphaned above: verification would put an outbound HTTP probe on the completion path. Self-clears via CheckResult.Resolved on the next run's positive 2xx/3xx re-observation"},
	"canonical_mismatch":      {catMechanical, "bugs_open/251 fleet-wide regression guard; live-probe check (GET the served page, compare its <link rel=canonical> against preferredStructuralURL's expected value, GET the canonical target itself). Deliberately NOT a verifier candidate — same outbound-HTTP-on-completion-path reason as its siblings in this map. Self-clears via CheckResult.Resolved when the next run's live re-probe finds the canonical correct"},
	"structured_data_invalid": {catMechanical, "live-probe check (GET the served page, json.Unmarshal every <script type=application/ld+json> block). The parse itself is deterministic given the body, but obtaining the body is the same outbound GET the siblings above decline to repeat on the completion path. Self-clears via CheckResult.Resolved when the next run's re-probe finds every block parses (including the vacuous zero-block case)"},
	"head_essentials_missing": {catMechanical, "live-probe check (GET the served page, assert non-empty <title>, a skip-link, a <footer>). Same outbound-HTTP-on-completion-path reason as its three siblings above. Self-clears via CheckResult.Resolved when the next run's re-probe finds all three present"},
	"sitemap_entry_dead_live": {catMechanical, "the narrow, safe half of 'every page appears in the sitemap' (the file header's WHAT IS DELIBERATELY NOT GATED explains why the wide half stays out): GET this site's own /sitemap.xml, parse every <loc>, GET each on-domain entry, 404|410 only, confirm-before-file — same probeInternalLinkTargets discipline as dead_internal_link_live above, reused verbatim. No page_id (a sitemap entry is the site's own stated URL, not a link found ON a page, so there is no linking page to carry — matches undeployed_asset's 'no per-item target id' shape above). Deliberately NOT a verifier candidate — same outbound-HTTP-on-completion-path reason as its four siblings in this map. Self-clears via CheckResult.Resolved when the next run's re-probe finds the entry live again"},

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

	// Phase B finance directories (2026-08-13) — the SAME two checks again,
	// instantiated from three further directoryCheckProfiles entries
	// (mortgage-lender, savings-provider, health-insurer). Not a new mechanism:
	// existence checks against page_components.function / pages.page_type,
	// handled by content-gap-planner. Registered but not in any discovery
	// agent's checks array yet (enablement is Phase B3f, ordered after the
	// image roll), so never observed live — a statement about the config, not
	// a claim that they work.
	"missing_mortgage_lender_directory_section":  {catMechanical, "page_component existence by function (mortgage-lender-directory); profile sibling of missing_model_directory_section; registered, not yet in the discovery checks array"},
	"missing_mortgage_lender_directory_page":     {catMechanical, "page existence by page_type (mortgage-lenders); profile sibling of missing_model_directory_page; registered, not yet in the discovery checks array"},
	"missing_savings_provider_directory_section": {catMechanical, "page_component existence by function (savings-provider-directory); profile sibling of missing_model_directory_section; registered, not yet in the discovery checks array"},
	"missing_savings_provider_directory_page":    {catMechanical, "page existence by page_type (savings-providers); profile sibling of missing_model_directory_page; registered, not yet in the discovery checks array"},
	"missing_health_insurer_directory_section":   {catMechanical, "page_component existence by function (health-insurer-directory); profile sibling of missing_model_directory_section; registered, not yet in the discovery checks array"},
	"missing_health_insurer_directory_page":      {catMechanical, "page existence by page_type (health-insurers); profile sibling of missing_model_directory_page; registered, not yet in the discovery checks array"},
	"unrendered_template":                        {catMechanical, "[INFERRED] check_integrity; never observed live"},
	"cross_site_contamination":                   {catMechanical, "[INFERRED] check_integrity; never observed live"},
	"forced_text_colors":                         {catMechanical, "[INFERRED] check_forced_text_colors — sibling of bugs_open/017's action; never observed live"},
	"duplicate_palette":                          {catMechanical, "[INFERRED] check_duplicate_palette; never observed live"},
	"placeholder_contact":                        {catMechanical, "[INFERRED] check_placeholder_contact; never observed live"},
	"broken_nav_links":                           {catMechanical, "[INFERRED] check_broken_nav_links; never observed live"},
	"backend_unreachable":                        {catMechanical, "[INFERRED] check_backend_unreachable, which already SELF-CLEARS on a live health probe — a verifier may be redundant here; check before writing one"},
	"site_unreachable":                           {catMechanical, "check_site_unreachable (bugs_open/236, 522 half) SELF-CLEARS via Resolved{AllOfType} on a serving probe — the same posture as backend_unreachable, decided at birth rather than inferred later; a completion verifier would re-run the identical probe the check already re-runs every rotation pass"},
	// ---- site acceptance council seats (RFC_056, loanzy_uk_example_site lane, 2026-08-25) ----
	// All three are flag-only VERDICT rows (HandlerAgent ""), filed 'detected' and never
	// promoted: the promoter and triage both require a handler. A verdict is graded by the
	// next pass of the same check, which retracts it through CheckResult.Resolved on a
	// positive re-observation — the same comparison a verifier would make, taken where the
	// probe already runs. Nothing on the completion path ever claims one done.
	"prerequisite_missing":  {catMechanical, "RFC_056 prerequisites seat (check_build_prerequisites.go): one row per missing prerequisite KIND (vertical_landscape / page_research / evidence_base / feed_sources), keyed prerequisite_missing:<kind>:<site>. Flag-only by design — bugs_open/380 D1 forbids minting a register to satisfy it, so there is no handler to verify; self-clears via Resolved when the kind is positively observed present"},
	"heading_promise_unmet": {catMechanical, "RFC_056 promise seat (check_heading_promise.go): a served page whose own <h1>-<h3> promises a structure (calendar, checklist, comparison, top-N) the non-anchor body does not contain. Flag-only — the repair is a planner/writer judgement, and the rule NOMINATES candidates, it does not adjudicate (a keyword is not a promise). Self-clears via Resolved when the page is re-observed keeping the promise; declines to judge on a parked domain (invented-path control 200s)"},
	"structure_floor_unmet": {catMechanical, "RFC_056 structure seat (check_structure_floor.go): fewer than N (owner ruling 2026-08-25: N = 6) distinct DELIVERED reader-facing structures across the served site and no recorded refusal. Flag-only — below the floor the seat records the shortfall and the delivered set; the refusal is a recorded planner/human verdict, not something a fixer can produce. Self-clears via Resolved when the count reaches N or a refusal is recorded"},
	// decision_blocked_change (RFC_015 §5b, save_sections_decision_gate.go): a
	// rebuild tried to overwrite a decision-protected slot without citing the
	// decision; the stored content was kept and this records it. Classified on the
	// way IN, before the first row exists, per this file's own 07-20 correction.
	//
	// catJudgement rather than catMechanical, and the distinction is real: the item
	// does not describe a DEFECT that could be re-checked. It records an EVENT that
	// already happened and was already handled correctly — the page is intact.
	// "Resolved" would have to mean "a human decided whether the blocked change
	// should now be made by citing the decision", which is a judgement, not a
	// predicate. Contrast decision_regression, which IS mechanical and now has a
	// verifier: there, the assertion can be re-run against the page.
	"decision_blocked_change": {catJudgement, "save_sections_decision_gate.go (RFC_015 §5b, owner ruling 2026-08-10) — records a PREVENTED overwrite, not a live defect: the stored content stands and the page is intact, so there is no predicate to re-run. Resolution is a human deciding whether to re-dispatch WITH a citation. Same shape as lock_blocked_change"},

	// ⚠ CONTRIBUTED FINDING, 2026-08-10, not this lane's to fix: **lock_blocked_change
	// is in NEITHER half of this guard and has 37 live rows.** It is produced by
	// lock_helpers.go in package `actions`, so the SENSOR (which scans only
	// discovery_checks source) cannot see it, and it is absent from liveItemTypes
	// below — so it is neither verified nor an acknowledged gap, and its
	// completions are taken on the handler's word with nothing recording that
	// choice. It is the direct sibling of the entry above and would classify the
	// same way. Left for whoever refreshes the ratchet rather than silently
	// adopted here, because the union rule means adding a type is a commitment
	// about someone else's producer. Measured: SELECT item_type, count(*) FROM
	// site_work_items WHERE item_type='lock_blocked_change' → 37.

	// ⚠ CONTRIBUTED CENSUS, 2026-08-12, from the bugs_open/213 D3 lane — not this
	// lane's to adopt, for the same reason the lock_blocked_change note above says:
	// adding a type here is a commitment about somebody else's producer.
	//
	// lock_blocked_change is NOT the only one. [MEASURED 2026-08-12, conservative] of
	// the 98 live item_type values in site_work_items, TWELVE (89 rows) appear NOWHERE
	// in this file — not verified, not an acknowledged gap, not in liveItemTypes — so
	// their completions are taken on the handler's word with nothing recording that
	// choice:
	//
	//	37  lock_blocked_change            (already known, see the note above)
	//	19  chrome_divergence_overwritten   16  save_refused_incomplete
	//	 5  stale_evidence                   2  content_edit
	//	 2  grounded_draft_review            2  page_divergence_overwritten
	//	 2  vision_finding                   1  alias_witness_136
	//	 1  citation_unverified              1  nav_rebuild_refused_incomplete
	//	 1  stale_directory_claim
	//
	// Every one is produced OUTSIDE this package, so the SENSOR half cannot see them
	// and only a liveItemTypes refresh would. Refresh query:
	//	SELECT item_type, count(*) FROM site_work_items GROUP BY 1 ORDER BY 1;
	// then subtract what this file names. ⚠ METHOD, because it nearly went out wrong:
	// a strict regex over the two maps reported FOURTEEN — it missed entries this file
	// spells differently. The 12 above come from over-collecting every quoted
	// lower-snake string in the whole file as "known", which biases toward FEWER blind
	// types, so each of the 12 is blind for certain.

	// decision_regression: GAP CLOSED 2026-08-10 by the RFC_015 lane, whose debt
	// this was. VerifyDecisionRegressionResolved re-runs the guard predicate over
	// the same stored assembly (both extracted into decisionGuardViolated /
	// storedPageAssemblySQL so check and verifier cannot drift), fail-closed on
	// anything it cannot evaluate. Removed from this map because the test rightly
	// refuses to let a type be both verified and listed as an acknowledged gap.
	// The 236 lane's classification note — which named the debt and left the
	// decision here — is what made it findable; thank you.

	// Produced OUTSIDE this package by refresh_evidence_fact_drift.go in package
	// `actions` (Piece 3 of PLAN_2026-08-09_facts_into_tool_acceptance.md, the
	// bugs_closed/225 class fix), so neither half of this guard would see it on its
	// own: the SENSOR scans discovery_checks source only, and the RATCHET is a
	// snapshot of types already in the database. Classified here on the way IN, in
	// the same commit that ships the producer — this file's own 07-20 correction.
	//
	// JUDGEMENT, and the distinction is real. The item says: a fact this tool
	// declares it encodes has moved, or its evidence has — AND the tool may not be
	// rewritten automatically (no tool-level component owns the code, or the fence
	// says no_auto_fix, or the evidence itself is what shifted). Every one of those
	// resolutions is a person's: whether the law actually changed, whether this
	// tool's arithmetic must follow, whether a lost citation means anything at all.
	// There is no predicate to re-run — the register's own next daily sweep is the
	// only re-check, and it is the producer. Contrast the improve_tool arm of the
	// same producer, which IS mechanical and inherits improve_tool's classification
	// above: there a fixer can act and the criteria fence can be re-driven.
	"fact_drift_review": {catJudgement, "refresh_evidence_fact_drift.go (bugs_closed/225 class) — a declared fact moved but the tool must not be auto-rewritten (not_a_fork / no_auto_fix / evidence_drift); carries page_id, and component_id when a fork exists. Resolution is a human ruling on whether the law moved and what the tool must now compute; the re-check is the daily evidence sweep itself"},

	// Produced OUTSIDE this package, by applyNewPage in apply_gap_plan_action.go
	// (bugs_open/081, 2026-07-31), so neither half of this guard can see it yet:
	// the SENSOR only scans discovery_checks source, and the RATCHET below is a
	// snapshot of item types already in the DB. Classified on the way IN rather
	// than waiting for the first row — which is the whole point of the 07-20
	// correction in this file's header.
	//
	// A verifier IS writable here (pages.page_type = the requested type once a
	// human has ruled), so this is mechanical rather than judgement: it is the
	// DECISION that needs a person, not the check that it was carried out.
	"mistyped_deployed_page": {catMechanical, "pages.page_type = spec.wanted_type for spec.page_name; produced by applyNewPage (bugs_open/081) and, since 2026-08-02, by UpsertPageForRole for the constant-role arms — tool-deployer, tool-generator, report-builder (bugs_open/175); read spec.source to tell them apart; never observed live"},

	// Produced OUTSIDE this package by cmd/verifier-remit-check, the daily class
	// detector for bugs_open/213 (owner ruling D3, 2026-08-11), so neither half of
	// this guard would see it on its own: the SENSOR scans discovery_checks source
	// only, and the RATCHET is a snapshot of types already in the database.
	// Classified here on the way IN, in the same commit that ships the producer.
	//
	// MECHANICAL, and deliberately unregistered — the distinction matters because
	// this map's header calls catMechanical "the actionable backlog, not an excuse
	// list". The predicate is exact and re-runnable (VerifierDeclaresRemit(subject)
	// is true, or the subject resolves to one producer family), but the check
	// RETRACTS its own items on a positive observation each run — WII-009's rule —
	// so the loop is already closed from the producer side. Registering a verifier
	// as well would also have to move the sql_for_agents/220 claim-timeout exclusion
	// in lockstep (TestRegisteredVerifiersMatchClaimTimeoutExclusion enforces both
	// directions). Worth doing; not folded into the commit that mints the type.
	//
	// The row is undispatchable by construction (status 'deferred' + empty
	// handler_agent, remit.go's double lock), because the remedy is a code change
	// by a person, and there is no handler that could ever claim it.
	"verifier_remit_gap":      {catMechanical, "a VERIFIED item_type has accumulated rows from more than one producer shape while its verifier declares no remit — bugs_open/213 one level up, from the class detector (cmd/verifier-remit-check). Predicate: discovery_checks.VerifierDeclaresRemit(spec.subject_type), or the subject collapsing back to one producer family. The detector closes its own findings on a positive observation, so completion-side verification is a follow-on rather than the only guard"},
	"brief_supplies_negation": {catJudgement, "a site's writer-visible brief HANDS the writer a phrase built on define-by-negation, from cmd/brief-negation-check (bugs_open/305). Filed against the site, terminal at needs_human_review with no handler, and deliberately so: the remedy is the SITE OWNER editing their own brief, and the writer-seam gate exempts brief-supplied phrases precisely because a site's voice specification outranks the fleet rules. There IS a mechanical re-observation — the phrase is either still in the visible surface or it is not — and the detector performs it, closing its own finding on a positive result; what is judgement is whether the phrase SHOULD go, which is why it is not catMechanical"},
	"spec_supplies_claim":     {catJudgement, "a site's specs hand a GENERATOR (any live agent, not just the writer) a first-person claim about the business that no evidence register can adjudicate — from the second detector in cmd/brief-negation-check (bugs_open/414). Filed against the site, terminal at needs_human_review with NO handler, and that is the design rather than a gap: an automated handler rewriting a spec off a spec-content finding is exactly how the audit fleet canonised the planted marker this check exists for. Sibling of brief_supplies_negation and classified the same way for the same reason — the re-observation IS mechanical (the claim is either still in the visible surface or it is not, and the detector closes its own finding on a positive result), while whether a brief SHOULD say it is a human call"},
	"unrendered_page_imagery": {catJudgement, "per-(site,state) rollup census of pages holding an active content-hero asset the page renders nowhere — bugs_open/114's closing detector (check_unrendered_page_imagery.go). needs_human_review with NO handler, deliberately: the three states route to three different owners (unwired → bugs_open/412 deploy-time wiring; fragment_slot → bugs_open/357 identity repair; no_image_slot → a composition decision), and any single automated remedy here is how undeployed_asset built its 1,651-row parked backlog prescribing deploys for a wiring problem. The re-observation IS mechanical and the check performs it — an emptied state retracts its own rollup via Resolved on a positive census; what is judgement is which remedy each site's owner applies"},

	// ---- creation: "make X exist" ----
	"needs_page":          {catCreation, "page existence; 49 of 365 carry page_id"},
	"needs_content_page":  {catCreation, "page existence; 13 of 196 carry page_id"},
	"needs_imagery":       {catCreation, "image asset existence"},
	"needs_content_image": {catCreation, "image asset existence"},
	"needs_hero_image":    {catCreation, "image asset existence"},
	"needs_logo":          {catCreation, "asset existence"},
	// needs_brand_head_assets: VERIFIER REGISTERED 2026-08-18 (bugs_open/131) —
	// the entry that stood here is CORRECTED, not just deleted. It read:
	// "'Exists' is a genuinely weak proof here — the artefact can be committed
	// and still be wrong (illegible favicon source) — so this stays catCreation
	// rather than becoming a verifier." That guarded against over-trusting
	// existence as a proof of QUALITY, and that half still stands (the eyeball
	// on the produced PNG stays; a verifier cannot see an illegible favicon).
	// What it missed is the other direction, and the other direction was the
	// live failure: items routed to deploy_image_asset (spec had no mode),
	// refused by its brand-head guard, and the refusal-as-result stamped 21
	// items 'complete' with zero artefacts derived — ABSENCE completing as
	// success, which an existence verifier blocks and nothing else did.
	// Resolved:false can only ever BLOCK a completion; weak proof is only a
	// hazard on the Resolved:true side, where it merely reproduces the
	// no-verifier behaviour this type had anyway.
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
	"audit_finding_audience":  {catJudgement, "audience opinion; HISTORICAL — see audit_finding_brief_fidelity"},
	"capability_gap":          {catJudgement, "capability opinion"},
	"owned_page_review":       {catJudgement, "human review item"},
	"evaluate_tools":          {catJudgement, "tool evaluation opinion"},

	// ---- first observed in site_work_items by the 2026-07-24 refresh ----
	"needs_human_review":            {catJudgement, "generic HITL checkpoint item (checkpoint_for_review_action.go); resolution IS the human's ruling"},
	"directory_citation_unverified": {catJudgement, "human ruling on citation candidates that failed live verification (directory_claims.go); nothing mechanical to re-run"},
	"audit_finding_brief_fidelity":  {catJudgement, "HISTORICAL: the audit_finding_* minting fallback was removed 2026-08-15 (bugs_open/279 — unknown categories now file capability_gap); rows predating that survive, so the type stays listed per the union rule. Sibling of audit_finding_audience"},
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
	// missing_{model_directory,adoption_tracker,protocol_tracker,
	// mortgage_lender_directory,savings_provider_directory,
	// health_insurer_directory}_{section,page}.
	// Note the scan is NOT blind here by luck — directoryCheckProfiles spells
	// each type as a literal in a field whose name ends "ItemType:", so
	// itemTypeLiteralRe picks all twelve up anyway and they are classified
	// below. Adding a further profile therefore still fails this guard until
	// it is classified, which is the behaviour we want.
	"check_directory.go": "p.SectionItemType / p.PageItemType — one profile per register kind (model, company, protocol, mortgage-lender, savings-provider, health-insurer)",
	// fl.f.Kind — exactly two values, both spelled as literals in
	// classifyNonPageAnchor and both classified in itemTypesWithoutVerifiers:
	// cta_names_nonpage_destination, cta_tel_malformed. A third Kind fails
	// this guard until it too is classified, which is the behaviour we want.
	"check_cta_nonpage.go": "fl.f.Kind — cta_names_nonpage_destination | cta_tel_malformed (bugs_open/299)",
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

// The claimed-item-timeout lockstep guard USED TO LIVE HERE. It moved on 2026-08-19
// to platform/orchestration/actions/claim_timeout_exclusion_lockstep_test.go, as
// TestClaimTimeoutExclusionCoversBothCompletionGates (bugs_open/317).
//
// WHY IT HAD TO LEAVE THIS PACKAGE, rather than being widened in place. It pinned the
// exclusion list against the verifier registry alone — migration 220's stated contract,
// "the LOCKSTEP TWIN of the RegisterVerifier() calls", which was complete while gate 2
// was the only completion gate. Gate 1b (noChangeGates) arrived 2026-08-13 with a second
// opt-in roster, and a type on that roster with no verifier was therefore NOT excluded:
// for it the sweep completed the item with neither gate running. Widening the guard means
// reading BOTH rosters, and noChangeGates lives in package `actions`, which imports this
// package — so it can be read from there and never from here.
//
// Nothing is un-guarded by the move: the successor asserts the same two directions over
// the union of the rosters, and was proven to fail in both directions by mutation before
// this one was deleted.

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
	// dark_section_audit is listed from the moment it is minted (bugs_open/213), not
	// at the next hand-refresh: the union rule means the ratchet protects it from
	// day one, and a type that only appears here after someone remembers is the very
	// gap the CORRECTED note in this file's header is about.
	"dark_section_audit",
	// decision_blocked_change is listed from the moment it is minted, same rule and
	// same reason as dark_section_audit above (RFC_015 §5b, 2026-08-10): the union
	// rule means the ratchet protects it from day one, and a type that only appears
	// here once someone remembers is exactly the gap this file's header corrects.
	"decision_blocked_change",
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
	// ported_tool_fix is listed from the moment it is minted (bugs_open/281), same
	// union rule and same reason as dark_section_audit above.
	"ported_tool_fix",
	"required_fields_missing", "responsive_fix", "section_edit",
	"section_source_drift",
	"silent_failure", "spacing_fix", "tone_shift", "truncated_component",
	"undeployed_asset",
	// verifier_remit_gap is listed from the moment it is minted (bugs_open/213 D3),
	// same union rule and same reason as dark_section_audit above: the ratchet
	// protects it from day one rather than from whenever somebody next refreshes
	// this snapshot. Its producer is cmd/verifier-remit-check, outside this package,
	// so the SENSOR half cannot see it at all.
	"verifier_remit_gap",
	// brief_supplies_negation is listed from the moment it is minted, same rule
	// (bugs_open/305): a type this file cannot see is a type whose completions
	// nobody is recording a choice about.
	"brief_supplies_negation",
	"spec_supplies_claim",
	"unfulfilled_hero_variant", "unresolved_cta", "voice_tells",
}
