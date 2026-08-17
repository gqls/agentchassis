package actions

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Coverage test for the section-list lock merge (bugs_closed/285, the
// section_list_assembly_is_lock_blind case; register LOCK-008; RFC_033 option 2,
// ruled by the owner 2026-08-17).
//
// THE RULE. If you read `site_plan_sections` to work out what sections a page
// has, you must merge the page's human-LOCKED live rows into that answer
// (datahelpers.LoadLockedPageSlots + MergeLockedPageSlots) — or be listed here
// with the reason you must not.
//
// WHY THAT TABLE AND NOT `pages.sections`. The plan is the store that CANNOT
// know about a lock: a human pins a section to a live page, and no plan tier
// ever hears about it — that is the whole of bugs_closed/285. The cache is a
// different case: since the loader started writing the MERGED list to
// `pages.sections` (2026-08-16, live v1.0.1305, proven on natural rebuilds of
// loancalculator.co.uk/index and webdesign.uk/contact), a reader of the cache
// gets the locked rows for free. So cache readers need no declaration, and the
// detector deliberately does not chase them — a check that fires on ~18 files
// teaches authors to add a line to a list, and this one has to make them think.
//
// WHY A TEST AND NOT A COMMENT. This is the same shape as the defect: one
// reader did not know a rule everyone else assumed, and nothing failed when it
// was written. The council's architecture seat raised exactly that on the fix
// (needs_rfc, corr 79f70435: "a future section-list consumer can still be
// written lock-blind exactly as the original bug arose … worth an RFC to decide
// whether this becomes a single mandatory entrypoint rather than a convention").
// The owner ruled for this, the cheap half of RFC_033: not one mandatory
// entrypoint — the readers legitimately want different answers, and one of them
// must NOT merge — but a build failure the next author has to answer.
//
// KNOWN WEAKNESS, stated so nobody mistakes its reach: it reads SOURCE, so it
// proves the CALL EXISTS, not that it executes on the path that matters. Same
// as component_template_writer_coverage_test.go, whose shape this copies.
// Comments are stripped before matching, because a source-scan test otherwise
// makes doc comments load-bearing (LANDMINES: "a source-scan test makes your
// COMMENTS load-bearing — first occurrence wins") — this very file names both
// the table and the helpers in prose above.

// readsThePlansSections matches a query against the authoritative tier. Written
// as a FROM/JOIN match rather than a bare table-name match so that naming the
// table in a string, an error message or a work-item spec is not a "read".
var readsThePlansSections = regexp.MustCompile(`(?is)\b(FROM|JOIN)\s+site_plan_sections\b`)

// mergesLockedRows: either half of the shared mechanism. A caller that loads the
// locked rows and does its own thing with them still had to think about locks,
// which is what this test is for.
var mergesLockedRows = regexp.MustCompile(`\b(MergeLockedPageSlots|LoadLockedPageSlots(ForSite)?)\s*\(`)

// lockBlindPlanReaders — readers of site_plan_sections that must NOT merge
// locked rows, each with the reason. A file here is a DECISION, and the reason
// is the deliverable: "it doesn't need it" is not one. Census measured
// 2026-08-17 by running this detector over actions/, actions/*/ and
// datahelpers/: **7 readers — 2 merge, 5 here**. (First written as "6: 2 merge,
// 4 here", which was the count BEFORE this test's own first run found
// check_sectionless_pages. Adding the entry without re-deriving the total left a
// stale figure that reached the register, the RFC and a council submission —
// council 02cb2134 caught it; WRONG_CALLS 2026-08-17. A census that gains a
// member has a new total, and the total is the part everything else quotes.)
var lockBlindPlanReaders = map[string]string{
	// Fills in a default layout ONLY for a page that has no sections from ANY
	// source, and no-ops if it has any. That is precisely the one case the
	// loader also declines to merge (a locked-only list is neither plan nor
	// page, and a rebuild on it would delete the unlocked siblings). Merging
	// here would make this action disagree with the pipeline it feeds: it
	// would see "this page has a section" and decline to scaffold, leaving the
	// page with a locked row and still no plan. Convergence is already correct
	// without it — once this writes a layout, a tier serves, and the next
	// build's loader merges the locked row back in.
	"ensure_page_section_layout_action.go": "acts only when EVERY source is empty — the same case the loader declines to merge; merging here would suppress the scaffold and strand the page",

	// The shared read-only answer to "would page-build-handler find anything to
	// build for this page?" (bugs_closed/177 → bugs_open/187). Its whole
	// contract is to PREDICT the loader, so it must mirror the loader's rule
	// including the loader's refusal: with no tier serving, the loader merges
	// nothing and the build does nothing, so a satisfiability check that
	// counted locked rows would answer "buildable" for a page the pipeline
	// will no-op on — the exact false promise 177 was filed for.
	"page_section_satisfiability.go": "predicts the loader, so it must mirror the loader's refusal to merge when no tier serves; counting locked rows would promise a build that no-ops",

	// Compares the CURRENT plan's sections against the sections of the plan
	// version a page was built from, to tell "the plan re-composed this page"
	// from "the plan left it alone" (the restamp path). Both sides are plan
	// versions — immutable, per-plan — and its own header says the comparison
	// is deliberately plan-to-plan and NOT plan-to-pages.sections. A locked
	// live row belongs to neither version; adding it to one side would make
	// every locked page read as re-composed, and rebuild it forever.
	"reconcile_site_plan_action.go": "compares plan version to plan version (restamp); a live locked row is in neither version and would make every locked page look re-composed",

	// Reads site_plan_sections for the SIBLING, not for the subject: it asks
	// "does a same-role sibling have a layout worth borrowing?", because the
	// remedy it files depends on the loader's sibling-synthesis fallback
	// firing. The subject page's emptiness is read from `pages.sections`, which
	// now carries the merged locked rows — so a page whose only content is a
	// locked section is only flagged when NO tier serves it, and then the
	// sequence is correct: scaffold a layout, and the next build's loader
	// merges the locked row into it. (This entry exists because the test found
	// the file and the author's own census had not: the JOIN form is a read.)
	"discovery_checks/check_sectionless_pages.go": "reads the plan for the SIBLING's layout, not the subject's membership; the subject's emptiness comes from pages.sections, which now carries merged locked rows",

	// "How many instances of each component does the PLAN say this page has?" —
	// the allowance denominator for the duplicate-section guards
	// (remove_duplicate_page_sections, save_sections_dedup). A locked row is
	// never a dedup victim in the first place (those guards carry the
	// agent-writable predicate), so adding locked rows here would raise the
	// allowance without adding a candidate, and a genuine duplicate would
	// survive. This is the pin-predicate-is-not-the-pool-predicate rule
	// (LANDMINES) applied to a count.
	"../datahelpers/plan_section_counts.go": "the PLAN's own instance count, used as a dedup allowance; locked rows are already excluded from the victim pool, so merging them would let a real duplicate survive",
}

// mustMergeByName pins the two readers that DO merge. Without this, deleting the
// merge from the loader would leave a test that passes by seeing no offenders —
// the "a quiet test passes when the RULE is gone" failure.
var mustMergeByName = []string{
	"load_page_sections_from_spec_action.go",
	"discovery_checks/check_section_source_drift.go",
}

func sectionListReaderScanRoots() []string {
	return []string{"*.go", "*/*.go", "../datahelpers/*.go"}
}

func TestEverySitePlanSectionsReaderDecidesAboutLocks(t *testing.T) {
	var offenders, merged []string
	sawAnyReader := false

	for _, pattern := range sectionListReaderScanRoots() {
		files, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("glob %s: %v", pattern, err)
		}
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			src, readErr := os.ReadFile(f)
			if readErr != nil {
				t.Fatalf("read %s: %v", f, readErr)
			}
			body := withoutLineComments(string(src))
			if !readsThePlansSections.MatchString(body) {
				continue
			}
			sawAnyReader = true

			key := f
			if strings.HasPrefix(f, "../datahelpers/") {
				key = f // datahelpers entries are keyed with their prefix
			}
			if _, declared := lockBlindPlanReaders[key]; declared {
				continue
			}
			if mergesLockedRows.MatchString(body) {
				merged = append(merged, f)
				continue
			}
			offenders = append(offenders, f)
		}
	}

	if !sawAnyReader {
		t.Fatal("found NO readers of site_plan_sections — the detector is broken, not the codebase; " +
			"this test must never pass vacuously")
	}

	for _, want := range mustMergeByName {
		found := false
		for _, m := range merged {
			if m == want || strings.HasSuffix(m, "/"+want) || m == filepath.Base(want) {
				found = true
				break
			}
		}
		// The drift check is globbed as "discovery_checks/check_...go"; compare on suffix too.
		if !found {
			for _, m := range merged {
				if strings.HasSuffix(m, filepath.Base(want)) {
					found = true
					break
				}
			}
		}
		if !found {
			t.Errorf("%s no longer reads site_plan_sections AND merges locked rows — either the merge "+
				"was removed (bugs_closed/285: the assembler was lock-blind, and this is the mechanism "+
				"that fixed it) or the detector stopped seeing it. Both are failures: without this pin, "+
				"deleting the merge would make this test pass by finding no offenders.", want)
		}
	}

	if len(offenders) > 0 {
		t.Fatalf("these files read site_plan_sections to work out a page's sections, but neither merge "+
			"the page's LOCKED live rows nor declare why they must not: %s\n"+
			"A human can pin a section to a live page (page_components.lock_type); no plan tier ever "+
			"hears about it, so a list built from the plan alone proposes REMOVING it on every rebuild "+
			"— that is bugs_closed/285, and the write guard was the only thing between it and deletion. "+
			"Either call datahelpers.LoadLockedPageSlots + MergeLockedPageSlots (the 058 guard's own "+
			"predicate and matchLockedRow's own pairing arms), or add the file to lockBlindPlanReaders "+
			"WITH THE REASON it must not. See RFC_033 and register LOCK-008.",
			strings.Join(offenders, ", "))
	}
}

// The exemption list must not rot into a car park: a stale entry would silently
// cover a future lock-blind read that lands in the same file.
func TestLockBlindPlanReadersAreAllStillPlanSectionReaders(t *testing.T) {
	for f, reason := range lockBlindPlanReaders {
		if strings.TrimSpace(reason) == "" {
			t.Errorf("%s is declared lock-blind with no reason — the reason is the deliverable", f)
		}
		src, err := os.ReadFile(f)
		if err != nil {
			t.Errorf("%s is declared lock-blind but cannot be read: %v — remove the stale entry or fix the path", f, err)
			continue
		}
		if !readsThePlansSections.MatchString(withoutLineComments(string(src))) {
			t.Errorf("%s is declared lock-blind but no longer reads site_plan_sections — remove the stale entry", f)
		}
	}
}
