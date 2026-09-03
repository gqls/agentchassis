// FILE: platform/orchestration/actions/discovery_checks/page_lifecycle_posture_test.go
//
// Build-enforced guard: a discovery check that queries `pages` must DECLARE
// which lifecycle posture it takes, and say why.
//
// WHY THIS EXISTS
// ---------------
// bugs_open/356. `pages` carries two independent axes, and the predicate family
// (platform/orchestration/datahelpers/links.go) states in terms that each
// consumer must pair the arms its own question needs:
//
//	BUILD     PageHasShippedPredicateFor / PageMayBeLinkedPredicateFor
//	          "has this page ever been served"
//	LIFECYCLE PageWantedLivePredicateFor  -> `status = 'active'`
//	          "does the platform still want it served"
//
// `check_orphan_pages` took the build arm alone. So an ARCHIVED page that
// shipped before it was retired was enumerated as an orphan and filed as work
// telling a handler to make it MORE REACHABLE — and all three handlers it routes
// to already refuse retired pages on their own account. Measured 2026-08-22: all
// 17 no-target `internal-linker` completions named an archived page, 34 archived
// pages were being filed across the three branches, and the same keys had
// recurred since April. The no-op completions then consumed the two-strike ladder
// (load_work_item_actions.go:1468-1504) and parked the queue at `unresolved`,
// which is terminal — so the defect retired legitimate work as collateral.
//
// The class is an ADOPTION GAP, not eighteen independent mistakes. On the day this
// was written only 3 files in this package called PageWantedLivePredicateFor; the
// rest hand-spelled the intent in at least four ways, two of which do not do what
// they appear to:
//
//	COALESCE(p.status,'') <> 'deleted'   excludes NOTHING — no `deleted` row exists
//	p.status IN ('active','deployed')    works by accident — `deployed` is a
//	                                     build_status value, not a pages.status one
//
// `pages.status` had exactly two live values on 2026-08-22: active 759 /
// archived 65.
//
// THE POSTURE IS NOT BINARY, and a two-valued guard would be actively harmful.
// bugs_open/266's note of 2026-08-14 warns by name against "fixing" audits with a
// blanket page-status filter: an archived page can still be serving 200 to the
// public, so a check that only OBSERVES is right to look at it.
// check_unverified_claims.go:458 therefore runs a THIRD posture — keep archived
// pages that are still serving, drop the ones that never shipped — and has a
// mutation test defending exactly that. A guard that could not express it would
// push authors back to hand-spelling, which is the disease.
//
// The discriminator is what the finding's REMEDY does, not whether the page is
// archived:
//
//	remedy mutates / re-links the page  -> PostureArmed
//	finding is observed and flagged only -> PostureObserves
//	both, conditionally                  -> PostureNuanced
//
// THE SENSOR / RATCHET SPLIT, and its honest limit
// ------------------------------------------------
// Deliberately modelled on verifier_coverage_test.go and handler_coverage_test.go
// — same package, same two-halves shape — because a second, differently-shaped
// list is the drift this whole class is about.
//
//  1. SENSOR — TestEveryPagesQueryingCheckDeclaresItsLifecyclePosture reads this
//     package's SOURCE for `FROM pages` / `JOIN pages`. It needs no refresh: a
//     check written tomorrow that queries pages fails the build until its author
//     declares a posture. This half cannot go stale.
//
//  2. ASSERTION — TestArmedChecksActuallyCarryALifecycleArm re-checks every
//     PostureArmed claim on every build: the file must carry a lifecycle arm
//     BOUND TO THE PAGES ALIAS the entry declares. This is what stops the
//     registry becoming a rubber stamp.
//
//     MUTATION-PROVEN 2026-08-22, with the decoys deliberately left in place:
//     deleting the arm from check_orphan_pages.go fails this test even though the
//     file still contains the string twice — once in the arm's own comment, once
//     as `sni.status = 'active'` on the nav join. Breaking the shared
//     toolEligibilityWhere fragment fails TestSharedEligibilityFragmentCarriesTheArm.
//
//     > ⚠ CORRECTED 2026-08-22, before this file was first committed. This
//     > paragraph previously read "IT IS A NECESSARY CONDITION, NOT A SUFFICIENT
//     > ONE … it cannot prove the predicate binds to the PAGES row", and treated
//     > that as an inherent limit to be documented. It was not inherent — it was
//     > an unaliased needle, and the fix was to bind the needle to a declared
//     > alias. The admission was written in the same breath as the weakness it
//     > described, which is exactly the shape verifier_coverage_test.go's own
//     > 2026-07-20 correction warns about: naming a guard's blind spot in its
//     > header does not discharge it, and a stated limit is where to look for a
//     > fix, not a place to stop.
//
//     WHAT IT STILL CANNOT DO, measured rather than assumed: it cannot tell that
//     the declared alias is really bound to `pages` rather than to some other
//     table in the same query. Nothing in the file's text distinguishes those, so
//     that one IS a human read, and the Reason string is where it is recorded.
//
//  3. BACKLOG — TestKnownLifecycleGapsAreReportedNotSilent prints the
//     PostureKnownGap count. Those entries are 356's measured, unfixed debt.
//
// ⚠ WHY THE GAPS ARE LISTED RATHER THAN FIXED HERE, and why that is not a false
// all-clear. LANDMINES records the trap that an allow-list silences the detector
// it was written for, and COMPONENT_WRITE_ALLOWED's own note says adding a known
// gap to quieten a check "converts a live debt into a false all-clear". The
// defence is that PostureKnownGap is NOT a pass: it is a distinct value from
// PostureObserves, every entry names bugs_open/356, and test 3 prints the running
// count so a backlog of seventeen reads as a backlog of seventeen. Fixing one is
// a one-line change plus moving its entry to PostureArmed — at which point test 2
// starts enforcing it. **Never move an entry to PostureArmed or PostureObserves
// to quieten anything; only move it because you read the FROM clause.**
package discovery_checks

import (
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

type lifecyclePosture int

const (
	// PostureArmed: the query constrains the PAGES row to the live lifecycle.
	// Required whenever the finding's remedy mutates, re-renders or re-links the
	// page — acting on a retired page is the bugs_closed/098 shape.
	PostureArmed lifecyclePosture = iota

	// PostureObserves: the check deliberately sees retired pages, because it
	// only flags and never prescribes a change to the page. An archived page can
	// still be serving to the public and such a check is right to look at it.
	PostureObserves

	// PostureNuanced: a bespoke predicate that is neither of the above — e.g.
	// "keep archived pages that are still serving, drop the ones that never
	// shipped". The Reason must state the predicate and where its test lives.
	PostureNuanced

	// PostureKnownGap: measured, unfixed debt from bugs_open/356. NOT a pass.
	// Counted and printed by test 3.
	PostureKnownGap
)

type posture struct {
	Posture lifecyclePosture

	// Alias is the SQL alias this file binds the PAGES table to — "p" in
	// `FROM pages p`, "" for an unaliased `FROM pages`. Required for
	// PostureArmed, and it is what makes the assertion discriminating rather
	// than decorative: nearly every check here also joins
	// `site_nav_items sni ... AND sni.status = 'active'`, so an unqualified
	// search for the arm is satisfied by a predicate on the wrong table. That
	// is not theoretical — the first version of this test PASSED on a
	// check_orphan_pages.go whose real arm had been deleted, because the nav
	// join's identical string was still there.
	Alias string

	// InheritsFrom names a shared SQL fragment that carries the arm on this
	// file's behalf (today: toolEligibilityWhere). Set it INSTEAD of Alias;
	// the fragment's own arm is asserted by
	// TestSharedEligibilityFragmentCarriesTheArm.
	InheritsFrom string

	// Reason must say WHY, in terms of what the finding's remedy does. A reason
	// that only restates the posture ("no filter needed") is not one.
	Reason string
}

// pageLifecyclePostures is keyed by FILE NAME, not by check name: a file is what
// an author adds, and several files hold more than one pages query. Where a
// file's queries differ, the entry records the WEAKEST posture present and the
// Reason says which query is the weak one — check_phantom_internal_links.go is
// the worked example.
//
// Classification date: 2026-08-22, from a read of all 71 non-test files in this
// package (bugs_open/356 §5). A sample was re-verified by hand in BOTH
// directions — three claimed-unarmed and two claimed-clean — because checking
// only the side that supports the thesis cannot detect a systematically
// over-reporting audit. Every PostureArmed claim is additionally re-checked
// mechanically by test 2 on every build.
//
// THE ARITHMETIC, stated because a reader must be able to see that no file was
// silently dropped (council 4cf291a2, editquality, low). 71 non-test files in the
// package; **51 of them query `pages` and are listed here; the other 20 do not
// query it at all** and are therefore out of scope — they are not omitted, they
// are not members. Nothing hand-maintains that split: the SENSOR recomputes it
// from the directory on every run, so a file that STARTS querying pages joins the
// population automatically and fails the build until it is classified.
//
//	20 Armed + 1 Nuanced + 7 Observes + 23 KnownGap = 51
//
// And the 23 gaps against bugs_open/356 §5's headline "18 checks can route an
// archived page AT A HANDLER": **17 of the 23 route at a handler, 6 do not**
// (they file at `handler_agent: ""`, or mutate/suppress without filing). The
// eighteenth was check_orphan_pages.go itself, fixed in the same commit as this
// file — so 17 remaining + 1 fixed = 18, and the two numbers agree.
var pageLifecyclePostures = map[string]posture{
	// ── Armed: the remedy touches the page, so a retired page must not qualify ──
	"check_orphan_pages.go":             {Posture: PostureArmed, Alias: "p", Reason: "bugs_open/356's subject. All three remedies (internal-linker, nav-updater, rerender-pages) make the page more reachable, and all three already refuse retired pages; the producer was the sole outlier."},
	"check_asset_reference_404.go":      {Posture: PostureArmed, Alias: "p", Reason: "PageWantedLivePredicateFor(\"p\") + shipped arm; the file's own comment names the two axes as independent."},
	"check_unrendered_page_imagery.go":  {Posture: PostureArmed, Alias: "p", Reason: "PageWantedLivePredicateFor(\"p\") on the census population: the check only flags, but a retired page's unrendered asset is not actionable by any of the three states' owners, and a census that lists it misdirects the reader (bugs_open/114 closing detector)."},
	"check_backend_entry_orphaned.go":   {Posture: PostureArmed, Alias: "p", Reason: "literal p.status = 'active' plus the shipped arm."},
	"check_componentless_pages.go":      {Posture: PostureArmed, Alias: "p", Reason: "files needs_content_page at page-build-handler — builds the page, so lifecycle is mandatory."},
	"check_content_image_missing.go":    {Posture: PostureArmed, Alias: "p", Reason: "routes needs_imagery/needs_content_image, which write to the page. NB the arm is spelled `status IN ('active','deployed')` — correct only by accident, since `deployed` is not a pages.status value. Normalise to PageWantedLivePredicateFor when next edited."},
	"check_dead_controls.go":            {Posture: PostureArmed, Alias: "p", Reason: "literal p.status = 'active' on the joined pages row."},
	"check_incomplete_page_group.go":    {Posture: PostureArmed, Alias: "", Reason: "spelled `status NOT IN ('deleted','archived')`; files needs_page at page-build-handler."},
	"check_misdirected_cta.go":          {Posture: PostureArmed, Alias: "p", Reason: "both queries spelled `status NOT IN ('deleted','archived')`; routes page_rerender."},
	"check_missing_prose_links.go":      {Posture: PostureArmed, Alias: "p", Reason: "files content_rewrite at page-build-handler with mode=edit_live, which RE-RUNS THE WRITER over the page's live prose — the most mutating remedy in this table, so the lifecycle arm is mandatory. Carries PageWantedLivePredicateFor(\"p\") + PageHasShippedPredicateFor(\"p\"); a retired page that is still serving belongs to retraction (bugs_closed/098), never to a rewrite."},
	"check_missing_structure.go":        {Posture: PostureArmed, Alias: "p", Reason: "p.status = 'active' on the EXISTS gate; routes needs_rerender."},
	"check_missing_tools.go":            {Posture: PostureArmed, Alias: "", Reason: "the routing query (evaluate_tools) carries status = 'active' + deployed_at IS NOT NULL. The other pages read in this file is a bare COUNT that routes nothing."},
	"check_orphan_element_refs.go":      {Posture: PostureArmed, Alias: "p", Reason: "literal p.status = 'active'."},
	"check_page_content_divergence.go":  {Posture: PostureArmed, Alias: "p", Reason: "literal p.status = 'active', with a comment naming archived explicitly."},
	"check_site_structural_validity.go": {Posture: PostureArmed, Alias: "p", Reason: "PageWantedLivePredicateFor(\"p\") + shipped arm."},
	"check_stylesheet_gutted.go":        {Posture: PostureArmed, Alias: "p", Reason: "PageWantedLivePredicateFor(\"p\") + shipped arm."},
	"check_tool_acceptance.go":          {Posture: PostureArmed, InheritsFrom: "toolEligibilityWhere", Reason: "inherits p.status = 'active' from toolEligibilityWhere (tool_eligibility.go)."},
	"check_tool_acceptance_due.go":      {Posture: PostureArmed, InheritsFrom: "toolEligibilityWhere", Reason: "inherits the arm from toolEligibilityWhere, plus the shipped arm."},
	"check_tool_health.go":              {Posture: PostureArmed, InheritsFrom: "toolEligibilityWhere", Reason: "inherits the arm from toolEligibilityWhere; routes improve_tool/audit_tool."},
	"check_tool_recreation_needed.go":   {Posture: PostureArmed, Alias: "p", Reason: "literal p.status = 'active'; routes needs_tool_recreation."},
	"check_unresolved_sections.go":      {Posture: PostureArmed, Alias: "p", Reason: "p.status = 'active' + build_status='deployed'. Files no work item — it UPDATEs pages directly, which is precisely why the arm is mandatory."},
	"check_voice_tells.go":              {Posture: PostureArmed, Alias: "p", Reason: "spelled `status IN ('active','deployed')` — see the check_content_image_missing note on that spelling."},

	// ── Nuanced ──
	"check_unverified_claims.go": {Posture: PostureNuanced, Reason: "NOT (p.status='archived' AND <never deployed>): KEEPS archived pages that are still serving, because an archived page can serve 200 and unsupported public claims must still be audited. Defended by a mutation test in check_unverified_claims_archivedskip_test.go. Do not 'simplify' to a plain status filter — bugs_open/266's 2026-08-14 note warns against exactly that."},

	// ── Observes only: no handler, so nothing acts on the page ──
	"check_archived_page_still_serving.go": {Posture: PostureObserves, Reason: "bugs_open/359: the check's entire SUBJECT is archived pages — \"is a page we RETIRED still answering 200 to the public?\" — so arming the lifecycle axis would blind it to its own population. That is the ⚠ case this file's header names when it warns that a two-valued guard would be actively harmful. It prescribes nothing: archived_page_still_serving files at handler_agent \"\", flag-only, because a wrongly-archived page serving correctly is indistinguishable ON THE WIRE from a rightly-archived page serving wrongly, and the choice between retracting and un-archiving is a human's. The file's ACTIVE-page arm exists only to emit Resolved (the page is no longer retired, so the finding's premise is gone) and to build the file-path collision guard it shares with retract_page_deployment; it routes nothing."},
	"check_cta_rank_anomaly.go":            {Posture: PostureObserves, Reason: "bugs_open/436: cta_rank_anomaly files at handler_agent \"\", flag-only — whether a fossil-ranked primary button is deliberate needs a human who knows the site's premise, so nothing is routed and no page is mutated. Its `pages` query is ONLY the migration-755 acknowledgement lookup (cta_rank_deliberate_nav_order = COALESCE(nav_order,100)), keyed on a page the RANKING has already chosen — and the ranking's own supply (datahelpers.CTAPositionalInteractiveSQL) carries the lifecycle arm already, via status IN ('active','deployed') plus PageMayBeLinkedPredicateFor. Arming this second query would add nothing reachable: an archived page cannot be rank-1, so the lookup can only ever run for a page that already passed the arm."},
	"check_decision_guards.go":             {Posture: PostureObserves, Reason: "decision_regression, handler_agent \"\" — flag only."},
	"check_forced_text_colors.go":          {Posture: PostureObserves, Reason: "files capability_gap at handler_agent \"\" (remit.go) — flag only."},
	"check_image_source_unsatisfiable.go":  {Posture: PostureObserves, Reason: "image_source_unsatisfiable, handler_agent \"\" — flag only."},
	"check_image_url_404.go":               {Posture: PostureObserves, Reason: "image_url_404, handler_agent \"\" — flag only."},
	"check_site_unreachable.go":            {Posture: PostureObserves, Reason: "site-scoped, handler_agent \"\"; the pages read is a title lookup."},
	"check_heading_promise.go":             {Posture: PostureObserves, Alias: "p", Reason: "heading_promise_unmet, handler_agent \"\" — flag only (RFC_056 promise seat). Carries PageWantedLivePredicateFor(\"p\") — the promise is judged over pages the platform still WANTS served; no build arm, the GET itself establishes shipped."},
	"check_structure_floor.go":             {Posture: PostureObserves, Alias: "p", Reason: "structure_floor_unmet, handler_agent \"\" — flag only (RFC_056 structure seat). Site-scoped verdict; carries PageWantedLivePredicateFor(\"p\") because the floor is judged over the pages the platform still WANTS served; no build arm — the GET itself establishes shipped."},
	"verify_required_fields_missing.go":    {Posture: PostureObserves, Alias: "p", Reason: "NOT a check — a completion VERIFIER, so it prescribes nothing and only decides whether a defect is gone. Arming the lifecycle filter here would be the WRONG direction and unsafe: an archived page can still be serving, so a retired page whose component still renders empty required fields must be GRADED, not have its ErrNoRows read as \"page gone, resolved\" — that would certify a live defect as fixed, which is the fail-open shape RFC_017 closed. The page lookup exists to establish positive absence, and every resolved arm names a live fact rather than a missing row."},
	"check_truncated_component.go":         {Posture: PostureObserves, Reason: "truncated_component, handler_agent \"\" — flag only."},
	"check_directory.go":                   {Posture: PostureObserves, Reason: "reads pages as COUNTs only; routes missing_* items at the SITE, never at a page row."},

	// ── Known gaps: bugs_open/356 §5, measured and unfixed. NOT a pass. ──
	// Routes an archived page AT A HANDLER — the harmful subset.
	"check_sectionless_pages.go":            {Posture: PostureKnownGap, Reason: "bugs_open/356: `COALESCE(p.status,'') <> 'deleted'` excludes NOTHING (no such status value exists), so it reads as a lifecycle arm and is not one. Files needs_content_page at page-build-handler, priority 90, with the archived PageID. Highest blast radius in the package."},
	"check_component_standards.go":          {Posture: PostureKnownGap, Reason: "bugs_open/356: build_status IN ('deployed','active') only; files needs_content_page at page-build-handler."},
	"check_empty_sections.go":               {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm on the pages join; files empty_section at page-build-handler with the archived PageID."},
	"check_literal_markdown.go":             {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm; routes page-rerender / section-editor, i.e. re-renders a retired page."},
	"check_contact_form_undeliverable.go":   {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm; routes page_rerender."},
	"check_required_fields_missing.go":      {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm; routes required-fields-missing-handler."},
	"check_placeholder_contact.go":          {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm; routes page-build-handler."},
	"check_content_duplication.go":          {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm; routes deduplicate-sections with the archived PageID."},
	"check_integrity.go":                    {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm on either pages join; both file content_rewrite at page-build-handler."},
	"check_phantom_internal_links.go":       {Posture: PostureKnownGap, Reason: "bugs_open/356: MIXED — the link TARGET set is correctly armed, the CONTAINER scan is not, so an archived container page files phantom_internal_link at page-build-handler. The weak query is the container scan; fix that one."},
	"check_hardcoded_section_colors.go":     {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm; routes color-variable-fixer."},
	"check_component_template_corrupted.go": {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm; routes component-creator, regenerating a component only an archived page uses."},
	"check_page_component_status_drift.go":  {Posture: PostureKnownGap, Reason: "bugs_open/356: p.build_status='deployed' only, so an archived-but-deployed page routes to component-template-fixer."},
	"check_revenue_shape.go":                {Posture: PostureKnownGap, Reason: "bugs_open/356: PageHasShippedPredicateFor only on both queries; routes component-template-fixer / content-gap-planner."},
	"check_empty_blog.go":                   {Posture: PostureKnownGap, Reason: "bugs_open/356: build_status only; files needs_blog_posts with an archived blog page's PageID."},
	"check_news_feed.go":                    {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm on the section queries or findStrandedNavPages; routes content-feed-orchestrator / content-gap-planner."},
	"check_placeholder_image_in_use.go":     {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm on the in-use gate, so an archived page's HTML can satisfy it and route image-build-handler. Weaker than the others: a gate, not per-page routing."},

	// Selects archived pages but files nothing at a handler — still debt, lower priority.
	"check_page_canonical_collision.go":         {Posture: PostureKnownGap, Reason: "bugs_open/356: the finder does not filter status (it SELECTs it); handler_agent is \"\" and a Go gate requires activeCount>=2, so the exposure is a group MEMBER, not a routed page. The verifier query IS armed."},
	"check_section_source_drift.go":             {Posture: PostureKnownGap, Reason: "bugs_open/356: `COALESCE(status,'') <> 'deleted'` — the same no-op spelling as check_sectionless_pages. handler_agent \"\", so flag-only today."},
	"check_premise_incomplete.go":               {Posture: PostureKnownGap, Reason: "bugs_open/356: PageHasShippedPredicateFor only on a COUNT subquery, so archived pages inflate the count that decides needs_strategy. No page is routed."},
	"check_phantom_internal_links_fragments.go": {Posture: PostureKnownGap, Reason: "bugs_open/356: MIXED — the whole-site concat scan has no arm and contributes archived pages' HTML to the fragment id set; the sibling query IS armed. Files no items of its own."},
	"check_unlinked_components.go":              {Posture: PostureKnownGap, Reason: "bugs_open/356: no arm on an `UPDATE page_components ... JOIN pages`. Files no work item — it MUTATES components on archived pages directly, which makes this a side-effect gap rather than a routing one."},
	"check_undeployed_assets.go":                {Posture: PostureKnownGap, Reason: "bugs_open/356: an archived page's rendered_html can SUPPRESS an undeployed_asset finding. Fail-quiet — the opposite direction from the rest of this list, and it means the gap hides findings rather than inventing them."},
}

// lifecycleArmSpellings are every spelling of the lifecycle axis in live use in
// this package on 2026-08-22, enumerated from the source rather than imagined.
// The two accidental ones are included because test 2 asks "is an arm present",
// not "is it well spelled" — their misuse is recorded in the Reason strings and
// in bugs_open/356 §5.
var lifecycleArmSpellings = []string{
	"PageWantedLivePredicateFor",
	"toolEligibilityWhere",
	"status = 'active'",
	"status='active'",
	"status IN ('active'",
	"status NOT IN ('deleted'",
}

var pagesQueryRE = regexp.MustCompile(`(?i)(FROM|JOIN)\s+pages\b`)

// pagesQueryingSourceFiles is the SENSOR: it reads the package directory, so it
// sees a check written tomorrow. Nothing here is hand-maintained.
func pagesQueryingSourceFiles(t *testing.T) []string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("cannot read package directory: %v", err)
	}

	var found []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		src, err := os.ReadFile(filepath.Clean(name))
		if err != nil {
			t.Fatalf("cannot read %s: %v", name, err)
		}
		if pagesQueryRE.Match(src) {
			found = append(found, name)
		}
	}
	sort.Strings(found)

	// A sensor that finds nothing passes every assertion built on it. This is
	// the 016b §9 "a gate's 0 findings has two causes" guard: the package
	// demonstrably queries pages in dozens of files, so an empty or tiny result
	// means the regex or the working directory broke, not that the debt is gone.
	if len(found) < 20 {
		t.Fatalf("sensor found only %d pages-querying files — it is broken, not the package clean", len(found))
	}
	return found
}

// TestEveryPagesQueryingCheckDeclaresItsLifecyclePosture is the SENSOR half.
func TestEveryPagesQueryingCheckDeclaresItsLifecyclePosture(t *testing.T) {
	for _, file := range pagesQueryingSourceFiles(t) {
		if _, ok := pageLifecyclePostures[file]; !ok {
			t.Errorf(`%s queries `+"`pages`"+` but declares no lifecycle posture.

Add an entry to pageLifecyclePostures in this file. Decide by asking what your
finding's REMEDY does, not whether archived pages feel relevant:

  the remedy mutates / re-renders / re-links the page -> PostureArmed, and add
      "AND " + datahelpers.PageWantedLivePredicateFor("<your pages alias>")
  the check only flags, handler_agent is ""           -> PostureObserves
  neither, e.g. keep archived-but-still-serving pages -> PostureNuanced

Full reasoning, and why this is not "audits should skip archived pages":
bugs_open/356, and the header of this file.`, file)
		}
	}
}

// TestPageLifecyclePostureRegistryHasNoStaleEntries stops the registry outliving
// the files it describes: a deleted or renamed check must not leave a declaration
// behind vouching for source that is gone.
func TestPageLifecyclePostureRegistryHasNoStaleEntries(t *testing.T) {
	live := map[string]bool{}
	for _, f := range pagesQueryingSourceFiles(t) {
		live[f] = true
	}
	for file := range pageLifecyclePostures {
		if !live[file] {
			t.Errorf("pageLifecyclePostures names %s, which no longer queries pages (renamed, deleted, or its query was removed). Delete the entry — a declaration about absent source is worse than none.", file)
		}
	}
}

// stripGoComments returns only the CODE of a Go source file, so a source scan
// cannot be satisfied by PROSE.
//
// ⚠ THIS IS NOT TIDINESS — THE FIRST VERSION OF THIS TEST WAS BLIND WITHOUT IT,
// AND THE MUTATION RUN IS WHAT FOUND THAT. Deleting the lifecycle arm from
// check_orphan_pages.go left the test PASSING, because the arm's own explanatory
// comment names `PageWantedLivePredicateFor` in prose — so the string was still
// in the file while the predicate was gone. A guard that reads its own
// documentation as evidence of the thing documented vouches for nothing, and it
// fails in the one direction that matters: silently, on a tree that has just
// lost the protection. This is LANDMINES' "a source-scanning test makes your
// COMMENTS load-bearing" firing in practice.
//
// IT USES go/scanner RATHER THAN STRING SURGERY, AND THAT IS THE SECOND LESSON.
// The hand-rolled version this replaces scanned for `/*` and `//` textually. It
// was wrong within minutes of being written: `req.Header.Set("Accept", "*/*")`
// appears in two checks in this package, and a naive scanner reads the `/*`
// inside that STRING LITERAL as the start of a block comment and discards the
// rest of the file — including the real lifecycle arm 400 lines later. Both
// check_asset_reference_404.go and check_page_content_divergence.go failed on a
// tree where they were correctly armed.
//
// The comment on the discarded version claimed its only weakness was
// UNDER-stripping, "which could only ever cause a false PASS". It over-stripped
// and caused a false FAIL — the stated limit was wrong about its own direction,
// which is the more interesting half: a hand-reasoned bound on a hand-rolled
// parser is worth about as much as the parser.
//
// ⚠ A SIBLING IMPLEMENTATION EXISTS AND IS DELIBERATELY NOT REUSED — it carries
// the bug described above. `actions/diagnose_load_runtime_schema_test.go:252`
// declares a function of the same name, byte-scanning for `//` and `/*`. Run
// against `req.Header.Set("Accept", "*/*")` it returns everything up to the `*`
// and DISCARDS THE REST OF THE FILE [MEASURED 2026-08-22, both inputs executed].
// Reusing it — which would also mean hoisting a test helper into non-test code
// across a package boundary — would have imported the exact defect this version
// was written to fix. Not filed as a bug and not edited from here: it scans
// exactly one file, `diagnose_load_runtime_action.go`, which contains no `*/*`
// [MEASURED: 0 occurrences], so it is LATENT there, not blind today. Its owning
// lane should know; the finding is recorded in bugs_open/356 rather than acted on
// uninvited. Raised by the council's reuse_agent seat (corr 4cf291a2, low) asking
// whether an existing helper had been searched for — it had not, and searching
// found one worth NOT using.
//
// go/scanner tokenises Go properly,
// so string literals, raw strings and comments are distinguished by the same code
// the compiler uses, and there is no bound left to reason about.
func stripGoComments(src string) string {
	var fset token.FileSet
	file := fset.AddFile("", fset.Base(), len(src))

	var sc scanner.Scanner
	// The nil error handler and mode 0 (no ScanComments) mean comments are
	// skipped outright and a malformed file degrades to fewer tokens rather
	// than panicking. Errors are counted, not ignored — see the check below.
	sc.Init(file, []byte(src), nil, 0)

	var out strings.Builder
	out.Grow(len(src))
	for {
		_, tok, lit := sc.Scan()
		if tok == token.EOF {
			break
		}
		if lit != "" {
			out.WriteString(lit)
		} else {
			out.WriteString(tok.String())
		}
		out.WriteByte('\n')
	}
	return out.String()
}

// TestStripGoCommentsIsNotItselfBlind guards the guard, in BOTH directions.
//
// Under-stripping puts TestArmedChecksActuallyCarryALifecycleArm back to reading
// prose as evidence, and it would still PASS — the silent failure this whole file
// exists to prevent, one level down. Over-stripping is loud rather than silent,
// but it is what actually happened here, so it is tested too: the last two cases
// are regressions from the hand-rolled version this replaced.
//
// ⚠ The spellings in lifecycleArmSpellings must each be a SINGLE Go token — an
// identifier, or text inside one string literal. stripGoComments emits one token
// per line, so a spelling spanning two tokens can never match. Every current
// spelling satisfies this; a new one must too, and the last case below is what
// will tell you when it does not.
func TestStripGoCommentsIsNotItselfBlind(t *testing.T) {
	cases := []struct {
		name   string
		src    string
		needle string
		want   bool // should the needle survive stripping?
	}{
		{"line comment is stripped", "// mentions status = 'active' in prose\nvar x = 1\n", "status = 'active'", false},
		{"block comment is stripped", "/*\n status = 'active'\n*/\nvar x = 1\n", "status = 'active'", false},
		{"trailing comment is stripped", "var x = 1 // status = 'active'\n", "status = 'active'", false},
		{"real code survives", "q := `WHERE p.status = 'active'`\n", "status = 'active'", true},
		{"code after a block comment survives", "/* noise */ q := `p.status = 'active'`\n", "status = 'active'", true},

		// REGRESSION (2026-08-22): `*/*` in an HTTP Accept header is not a
		// comment. The hand-rolled stripper read its `/*` as a block-comment
		// opener and discarded everything to end of file, so the real arm 400
		// lines below went missing and two correctly-armed checks failed.
		{
			"a */ inside a string literal does not eat the file",
			"req.Header.Set(\"Accept\", \"*/*\")\nq := `WHERE p.status = 'active'`\n",
			"status = 'active'", true,
		},
		// A `//` inside a string literal is not a comment either — the failure
		// mode the discarded version claimed was its ONLY one.
		{
			"a // inside a string literal does not truncate the line",
			"u := \"https://example.com\"\nq := `WHERE p.status = 'active'`\n",
			"status = 'active'", true,
		},
		// The single-token rule, made visible: an identifier survives, and a
		// needle spanning two tokens does not.
		{"an identifier survives", "x := datahelpers.PageWantedLivePredicateFor(\"p\")\n", "PageWantedLivePredicateFor", true},
		{"a needle spanning two tokens cannot match", "x := datahelpers.PageWantedLivePredicateFor(\"p\")\n", "datahelpers.PageWantedLivePredicateFor", false},
	}
	for _, c := range cases {
		got := strings.Contains(stripGoComments(c.src), c.needle)
		if got != c.want {
			t.Errorf("%s: needle %q survived=%v, want %v\nstripped: %q", c.name, c.needle, got, c.want, stripGoComments(c.src))
		}
	}
}

// lifecycleArmFor builds the alias-bound needles for one declared alias.
//
// Binding to the alias is what makes this assertion discriminating. Searching
// for a bare `status = 'active'` is satisfied by the `site_nav_items sni` join
// that nearly every check in this package carries — the exact confusion that
// produced a false clean reading on check_orphan_pages.go twice: once as a
// grep-count while filing bugs_open/356 (WRONG_CALLS 2026-08-22), and once from
// the first version of THIS TEST, which passed on a mutated tree with the real
// arm deleted.
//
// `[^.\w]` before the unaliased form is what stops `sni.status` matching when
// the pages table is unaliased. RE2 has no lookbehind, so the guard is a
// character class and the alternation starts at a boundary.
func lifecycleArmFor(alias string) []*regexp.Regexp {
	pred := `\s*(=\s*'active'|IN\s*\('active'|NOT\s+IN\s*\('deleted')`

	if alias == "" {
		return []*regexp.Regexp{
			regexp.MustCompile(`(^|[^.\w])status` + pred),
			// The helper called with the unaliased form.
			regexp.MustCompile(`PageWantedLivePredicateFor\s*\(\s*""\s*\)`),
		}
	}
	q := regexp.QuoteMeta(alias)
	return []*regexp.Regexp{
		regexp.MustCompile(`\b` + q + `\.status` + pred),
		regexp.MustCompile(`PageWantedLivePredicateFor\s*\(\s*"` + q + `"\s*\)`),
	}
}

// TestSharedEligibilityFragmentCarriesTheArm closes the hole under
// InheritsFrom. Three checks inherit their lifecycle arm from
// toolEligibilityWhere rather than spelling one; without this, deleting the arm
// from the shared fragment would leave all three still naming the identifier and
// still passing. A guard that can be defeated one level up is not a guard.
func TestSharedEligibilityFragmentCarriesTheArm(t *testing.T) {
	const fragmentFile = "tool_eligibility.go"

	src, err := os.ReadFile(fragmentFile)
	if err != nil {
		t.Fatalf("cannot read %s — if the shared fragment moved, retarget this test rather than deleting it: %v", fragmentFile, err)
	}
	body := stripGoComments(string(src))

	if !strings.Contains(body, "toolEligibilityWhere") {
		t.Fatalf("%s no longer defines toolEligibilityWhere; the three checks declaring InheritsFrom point at nothing.", fragmentFile)
	}

	armed := false
	for _, re := range lifecycleArmFor("p") {
		if re.MatchString(body) {
			armed = true
			break
		}
	}
	if !armed {
		t.Errorf(`%s defines toolEligibilityWhere but it carries no lifecycle arm on the pages row.

check_tool_acceptance.go, check_tool_acceptance_due.go and check_tool_health.go all
declare InheritsFrom: "toolEligibilityWhere" and rely on this fragment for it. Either
restore the arm here, or give each of those three its own and drop InheritsFrom.`, fragmentFile)
	}
}

// TestArmedChecksActuallyCarryALifecycleArm is the ASSERTION half: it re-checks
// every PostureArmed claim against the source on every build, so the registry
// cannot become a rubber stamp.
//
// MUTATION-PROVEN, not merely mutation-provable (2026-08-22). Deleting the
// `datahelpers.PageWantedLivePredicateFor("p")` line from check_orphan_pages.go
// makes this test FAIL. **It took three attempts to get there, and the two failures
// are why this comment is long:**
//
//  1. The first version read the whole file, so the arm's own explanatory COMMENT
//     satisfied it — see stripGoComments.
//  2. The second version stripped comments but searched for an unqualified
//     `status = 'active'`, which the file's `site_nav_items sni` join supplies. It
//     passed on a tree with the arm deleted, which is the same false-clean reading
//     the original grep-count gave (WRONG_CALLS 2026-08-22) — the SAME mistake,
//     reproduced inside the guard written to prevent it.
//
// The lesson worth keeping: a source-scanning guard is a measurement, and it
// answers the question you ENCODED. Both failures were of a guard that looked
// right, passed its own suite, and asserted nothing. Only mutation told them apart.
func TestArmedChecksActuallyCarryALifecycleArm(t *testing.T) {
	for file, p := range pageLifecyclePostures {
		if p.Posture != PostureArmed {
			continue
		}
		src, err := os.ReadFile(filepath.Clean(file))
		if err != nil {
			t.Fatalf("cannot read %s: %v", file, err)
		}
		body := stripGoComments(string(src))

		if p.InheritsFrom != "" {
			if !strings.Contains(body, p.InheritsFrom) {
				t.Errorf("%s declares InheritsFrom %q but no longer references it. Either restore the shared fragment's use, or spell the arm locally and set Alias instead.", file, p.InheritsFrom)
			}
			continue
		}

		armed := false
		for _, re := range lifecycleArmFor(p.Alias) {
			if re.MatchString(body) {
				armed = true
				break
			}
		}
		if !armed {
			t.Errorf(`%s is declared PostureArmed but carries no lifecycle arm bound to the pages alias %q.

Three things this can mean, in order of likelihood:
  1. the arm was removed — restore it; the declaration says this check's remedy
     touches the page, so a retired page must not qualify;
  2. the pages alias changed — update Alias in pageLifecyclePostures to match the
     FROM clause;
  3. it is spelled a new way — extend lifecycleArmFor, and say in the Reason why a
     new spelling beat datahelpers.PageWantedLivePredicateFor.

A predicate on a JOINED table does not count and will not match: that is deliberate,
and it is the whole reason the needle is alias-bound.`, file, p.Alias)
		}
	}
}

// TestEveryPostureCarriesAReason. An entry with no reason is a rubber stamp, and
// a rubber stamp is how an allow-list turns live debt into a false all-clear.
func TestEveryPostureCarriesAReason(t *testing.T) {
	for file, p := range pageLifecyclePostures {
		if len(strings.TrimSpace(p.Reason)) < 20 {
			t.Errorf("%s declares a posture with no usable Reason. Say what the finding's remedy does — that is the discriminator, not the page's status.", file)
		}
		if p.Posture == PostureKnownGap && !strings.Contains(p.Reason, "bugs_open/356") {
			t.Errorf("%s is a PostureKnownGap whose Reason does not name bugs_open/356. A gap that cites no ticket is indistinguishable from a decision.", file)
		}
	}
}

// TestKnownLifecycleGapsAreReportedNotSilent is the BACKLOG half. It does not
// fail — the gaps are bugs_open/356's measured, unfixed debt and failing here
// would only tempt someone to reclassify them. It PRINTS, so the backlog cannot
// quietly read as clean, and so a shrinking number is visible progress.
//
// This is the deliberate answer to the allow-list landmine: the list does not
// silence the finding, it enumerates it.
func TestKnownLifecycleGapsAreReportedNotSilent(t *testing.T) {
	var gaps []string
	for file, p := range pageLifecyclePostures {
		if p.Posture == PostureKnownGap {
			gaps = append(gaps, file)
		}
	}
	sort.Strings(gaps)

	t.Logf("bugs_open/356: %d discovery checks still select retired pages without a lifecycle arm.", len(gaps))
	for _, g := range gaps {
		t.Logf("  GAP %s — %s", g, pageLifecyclePostures[g].Reason)
	}
	t.Logf("Fixing one: add `AND \" + datahelpers.PageWantedLivePredicateFor(\"<alias>\")` to the query, " +
		"move its entry to PostureArmed, and test 2 starts enforcing it. " +
		"Never move an entry to quieten anything — only because you read the FROM clause.")
}
