// FILE: platform/orchestration/actions/discovery_checks/check_unverified_claims_archivedskip_test.go
//
// Pins the page-status exclusion added to ScanDeployedClaims on 2026-08-15
// (owner instruction: "scan level exclusion"): archived AND never deployed.
//
// WHAT IS DEMONSTRATED: that the emitted page query carries the CONJUNCTION, and
// that neither half survives on its own. That is a narrow claim, and it is the
// honest one — see below.
//
// WHAT IS NOT DEMONSTRATED, and cannot be from here: that Postgres selects the
// rows we think it does. sqlmock does not execute predicates; it returns
// whatever rows the test hands it regardless of the WHERE clause. So this is a
// CONTRACT test on query text, exactly like
// TestLockedFilteringHappensWhereEachSideClaimsItDoes in the sibling file, and
// for the same reason. The behavioural half was established against the live
// database instead, and is recorded here because a future reader will otherwise
// re-derive it:
//
//	measured 2026-08-15, fleet-wide, pages carrying scannable components
//	----------------------------------------------------------------------
//	archived AND deployed_at IS NULL   1 page  / 7 components  <- excluded
//	archived AND deployed_at NOT NULL  11 pages                <- kept
//	  ...of which serving HTTP 200     5                       <- why it is kept
//
// THE FIVE ARE THE POINT. fundamentallyai.com/blog/ai-readiness-checker-guide.html
// (25,861 B), fundamentallyai.com/tools/llm-cost-calculator/index.html (35,331 B),
// leopardessconsulting.co.uk/our-approach.html (28,948 B),
// robot-hands.com/gripper-catalog.html (30,997 B) and robot-hands.com/news.html
// (48,047 B) all return 200 while `pages.status = 'archived'`, each against a
// fabricated-URL control on its own domain returning 404 at a different byte
// count. Weakening this predicate to `p.status = 'archived'` — which reads as
// the obvious simplification, and which matches the liveness predicate five
// other call sites in this estate use — would silently stop claim-checking five
// live pages. That is what the second and third tests below exist to stop.
//
// PROVEN BY MUTATION, not by a green run (2026-08-15, against a clean
// `git archive HEAD` tree so no other session's WIP could colour it — another
// session had this package's check_missing_structure.go dirty at the time).
// An unmutated control ran first and produced no failures, so an empty result
// below would have meant a broken harness rather than a passing guard — which
// is exactly what the first attempt did produce, from a grep anchored to
// `^\s+--- FAIL:` that cannot match a top-level failure line:
//
//	mutation applied to check_unverified_claims.go     caught by
//	------------------------------------------------   ----------------------
//	(control: no mutation)                             nothing — as required
//	drop the exclusion entirely                        test 1
//	weaken to `NOT (p.status = 'archived')`            tests 1 and 2
//	weaken to `NOT (p.deployed_at IS NOT NULL)`        tests 1 and 3
//	drop it here, add a p.status filter to site query  tests 1 and 4

package discovery_checks

import (
	"strings"
	"testing"
)

// theExclusion is the predicate as it must appear in the page query. Written out
// once, and compared as a whole: a partial match is precisely the failure this
// file is about, so there is deliberately no "contains p.status" shortcut.
const theExclusion = "NOT (p.status = 'archived' AND p.deployed_at IS NULL)"

// capturePageAndSiteSQL reuses the sibling file's sqlmock harness rather than
// standing up a second one. Its name says "lockedSkip" because that is what it
// was built for; what it actually does is drive ScanDeployedClaims over given
// rows and hand back the SQL issued, which is what both files need.
func capturePageAndSiteSQL(t *testing.T) (pageSQL, siteSQL string) {
	t.Helper()
	_, _, seen := runLockedSkipScanCapturingSQL(t, []lockedSkipComponent{unlockedRow()})
	pageSQL, siteSQL = seen[0], seen[1]

	// Premise check, borrowed from the sibling test for the same reason: without
	// it the two assertions could silently swap and still pass.
	if !strings.Contains(pageSQL, "FROM page_components pc") || !strings.Contains(siteSQL, "FROM site_components sc") {
		t.Fatalf("captured queries are not (page, site) as assumed:\n [0] %s\n [1] %s", pageSQL, siteSQL)
	}
	return pageSQL, siteSQL
}

// Test 1 — the exclusion is present, whole.
func TestPageQueryExcludesArchivedAndNeverDeployed(t *testing.T) {
	pageSQL, _ := capturePageAndSiteSQL(t)

	if strings.Contains(pageSQL, theExclusion) {
		return
	}

	// Not a bare "missing" — say WHICH way it broke, because the two weakenings
	// have opposite consequences and a reader fixing this needs to know which one
	// they caused.
	switch {
	case !strings.Contains(pageSQL, "p.status") && !strings.Contains(pageSQL, "p.deployed_at"):
		t.Errorf("the page-status exclusion is GONE — every archived, never-deployed page is "+
			"being scanned again, so a withdrawn draft will file findings that can never close.\n%s", pageSQL)
	case !strings.Contains(pageSQL, "p.deployed_at"):
		t.Errorf("the exclusion has lost its deployed_at half. It now drops archived pages "+
			"REGARDLESS of publication, and five archived pages were measured serving HTTP 200 on "+
			"2026-08-15 — this stops claim-checking live copy.\n%s", pageSQL)
	case !strings.Contains(pageSQL, "p.status"):
		t.Errorf("the exclusion has lost its status half. It now drops every never-deployed page, "+
			"including active ones still being built, which is a much wider blind spot than intended.\n%s", pageSQL)
	default:
		t.Errorf("the exclusion is present but no longer matches %q verbatim — if the rewrite is "+
			"deliberate, confirm it is still a CONJUNCTION before updating this constant.\n%s",
			theExclusion, pageSQL)
	}
}

// Test 2 — the status half must never stand alone. This is the mutation that
// looks like a tidy-up and costs five live pages.
func TestExclusionIsNotStatusAlone(t *testing.T) {
	pageSQL, _ := capturePageAndSiteSQL(t)

	// Demand control: if the query no longer mentions archived at all, test 1 owns
	// that failure and this test would pass vacuously. Refuse to be the one that
	// certifies it.
	if !strings.Contains(pageSQL, "'archived'") {
		t.Skip("no archived predicate at all — TestPageQueryExcludesArchivedAndNeverDeployed reports this")
	}

	if strings.Contains(pageSQL, "NOT (p.status = 'archived')") ||
		strings.Contains(pageSQL, "p.status <> 'archived'") ||
		strings.Contains(pageSQL, "p.status != 'archived'") {
		t.Errorf("archived pages are being excluded WITHOUT the never-deployed half. Measured "+
			"2026-08-15: 5 of 11 archived pages served HTTP 200 (fundamentallyai.com ×2, "+
			"leopardessconsulting.co.uk/our-approach.html, robot-hands.com/gripper-catalog.html, "+
			"robot-hands.com/news.html). Archiving a page sets a column; it does not take the "+
			"artefact out of the deploy repo.\n%s", pageSQL)
	}
}

// Test 3 — the deployed_at half must never stand alone either. Its failure mode
// is quieter: it would drop pages that are merely still being built.
func TestExclusionIsNotNeverDeployedAlone(t *testing.T) {
	pageSQL, _ := capturePageAndSiteSQL(t)

	if !strings.Contains(pageSQL, "p.deployed_at") {
		t.Skip("no deployed_at predicate at all — TestPageQueryExcludesArchivedAndNeverDeployed reports this")
	}

	if strings.Contains(pageSQL, "NOT (p.deployed_at IS NULL)") ||
		strings.Contains(pageSQL, "p.deployed_at IS NOT NULL") {
		t.Errorf("pages are being excluded on publication alone. That drops every ACTIVE page "+
			"that has not shipped yet — copy which is about to be published unchecked, which is "+
			"the case bugs_open/093 put the content_data surface in this scan for.\n%s", pageSQL)
	}
}

// Test 4 — the exclusion belongs to the page query only. site_components rows
// hang off the site, not a page: there is no p.status to filter on, and adding
// one there would drop site chrome for reasons that have nothing to do with it.
func TestSiteChromeQueryCarriesNoPageStatusFilter(t *testing.T) {
	_, siteSQL := capturePageAndSiteSQL(t)

	if strings.Contains(siteSQL, "p.status") || strings.Contains(siteSQL, "deployed_at") {
		t.Errorf("the site_components query has acquired a page-status filter. Chrome is not "+
			"owned by a page, so this cannot mean what it appears to mean.\n%s", siteSQL)
	}
}
