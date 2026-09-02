// FILE: platform/orchestration/actions/resolve_internal_links_authored_destination_test.go
//
// Pins bugs_open/248, slug cta_recompute_clobbers_authored_contact_links: an
// AUTHORED CTA destination in a utility area (contact/about/privacy/terms/legal)
// must survive both writers, while a FRESHLY PICKED one must still never land
// there. Those are two different decisions that shared one set
// (areasExcludedFromCTA) and therefore one answer, which is how "Get in touch"
// -> /contact.html was replaced by an unrelated tool page on 13 live
// components, webdesign.uk's homepage hero among them.
//
// The asymmetry is the whole fix, so TestFreshPickRefusesUtilityWhileStoredUtilityIsKept
// exists to make re-merging the two uses impossible to do quietly.

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// contactSitePages is the shape every test below shares: a site with a real
// contact page, a real tool, and a real ordinary page.
func contactSitePages() datahelpers.PageURLSet {
	return datahelpers.NewPageURLSet([]string{
		"/index.html", "/contact.html", "/services.html",
		"/tools/tool-ai-data-risk-checker.html", "/tools/password-entropy.html",
	})
}

// TestStoredCTADestinationIsAuthored pins the predicate the whole fix rests on,
// including the arm that keeps bugs_open/203 repairable: an INVALID
// /contact.html is the phantom fallback, not an authored link, and must stay
// recomputable.
func TestStoredCTADestinationIsAuthored(t *testing.T) {
	valid := contactSitePages()

	cases := []struct {
		url  string
		want bool
		why  string
	}{
		{"/contact.html", true, "valid page in a utility area — no resolver path can produce this"},
		{"/services.html", false, "ordinary page: the resolver picks these, so provenance cannot be derived"},
		{"/tools/password-entropy.html", false, "a tool is exactly what the positional pick offers"},
		{"", false, "no stored value at all"},
		// The 203 arm. /about.html is a utility URL but this site has no such
		// page, so the stored value is a fabricated fallback pointing nowhere.
		{"/about.html", false, "utility area but NOT a real page — 203's phantom stays repairable"},
	}
	for _, c := range cases {
		stored := map[string]interface{}{"cta_url": c.url}
		if got := storedCTADestinationIsAuthored(stored, "cta_url", valid); got != c.want {
			t.Errorf("storedCTADestinationIsAuthored(%q) = %v, want %v — %s", c.url, got, c.want, c.why)
		}
	}

	// bugs_open/308 Phase A: the same shape test, now with the RECORD present.
	// A utility url THIS RESOLVER MINTED is its own output, so it is not
	// authored and must stay re-derivable — otherwise the keep freezes the
	// resolver's own answer for ever, which is the LNK-033 landmine and the
	// reason the widening could not ship before this stamp existed.
	minted := map[string]interface{}{
		"cta_url": "/contact.html",
		datahelpers.CTAMintedKey: map[string]interface{}{
			"cta_url": "/contact.html",
		},
	}
	if storedCTADestinationIsAuthored(minted, "cta_url", valid) {
		t.Error("a stored utility url covered by the mint record is the RESOLVER's output, not authored — " +
			"treating it as authored is exactly the freeze this record exists to prevent")
	}

	// Value-binding, stated as its own case because it is the property the
	// whole design rests on: a stamp naming a DIFFERENT url does not cover the
	// current value, so a human's edit under a stale stamp stays authored.
	stale := map[string]interface{}{
		"cta_url": "/contact.html",
		datahelpers.CTAMintedKey: map[string]interface{}{
			"cta_url": "/tools/password-entropy.html",
		},
	}
	if !storedCTADestinationIsAuthored(stale, "cta_url", valid) {
		t.Error("a stamp naming a different url must not cover this value — a presence-bound stamp would " +
			"licence the recompute to clobber a section-editor edit, which is bugs_open/248 again")
	}

	// The record is per-FIELD, not per-section: a stamp on the primary slot
	// says nothing about the secondary. Pinned because both live on one nested
	// map and a whole-map presence test would pass this by accident.
	crossField := map[string]interface{}{
		"secondary_cta_url": "/contact.html",
		datahelpers.CTAMintedKey: map[string]interface{}{
			"cta_url": "/contact.html",
		},
	}
	if !storedCTADestinationIsAuthored(crossField, "secondary_cta_url", valid) {
		t.Error("a mint stamp on cta_url must not vouch for secondary_cta_url")
	}
}

// TestApplyCTARecomputeKeepsAuthoredContactLink is bug 248's own A/B
// reproduction, restated as the CORRECT assertion. The bug file ran the
// package's existing generic-label test with one variable changed — the stored
// url — and got:
//
//	CONTROL  stored=/tools/password-entropy.html label="Get in Touch" -> resolved=map[]
//	CASE     stored=/contact.html                label="Get in Touch" -> resolved=map[
//	             cta_url:/tools/tool-ai-data-risk-checker.html ...]
//
// Both halves run here, so the control proves the case is comparing like with
// like rather than passing for an unrelated reason.
func TestApplyCTARecomputeKeepsAuthoredContactLink(t *testing.T) {
	valid := contactSitePages()
	positionalTarget := contentHub{
		Name: "tool-risk-checker", Title: "AI Data Risk Checker",
		URL: "/tools/tool-ai-data-risk-checker.html",
	}
	candidates := riskCheckerCandidates(t)

	// "Get in Touch" reduces to [touch] — get/in are stopwords — so it names no
	// candidate. That is precisely why genuine contact copy could not save
	// itself via the label-match branch.
	const genericContactLabel = "Get in Touch"

	// CONTROL: an ordinary stored destination, same label. Untouched.
	control := map[string]interface{}{}
	applyCTARecompute(control, map[string]interface{}{"cta_url": "/tools/password-entropy.html"},
		"cta_url", positionalTarget, valid, "/index.html", genericContactLabel, candidates, "")
	if len(control) != 0 {
		t.Errorf("CONTROL: ordinary stored destination should be left untouched, got %v", control)
	}

	// CASE: the same call with an authored contact destination. Kept.
	got := map[string]interface{}{}
	applyCTARecompute(got, map[string]interface{}{"cta_url": "/contact.html"},
		"cta_url", positionalTarget, valid, "/index.html", genericContactLabel, candidates, "")
	if got["cta_url"] != "/contact.html" {
		t.Errorf("CASE: authored contact link clobbered — cta_url = %v, want /contact.html", got["cta_url"])
	}
	if _, wrote := got["cta_target_title"]; wrote {
		t.Errorf("CASE: wrote a target title the stored data never had: %v", got)
	}
}

// TestApplyCTARecomputeCarriesAuthoredTargetTitle — when the stored slot does
// carry a companion title, keeping the url must keep the pair together, or a
// title-comparing checker reads the section as inconsistent.
func TestApplyCTARecomputeCarriesAuthoredTargetTitle(t *testing.T) {
	got := map[string]interface{}{}
	applyCTARecompute(got,
		map[string]interface{}{"cta_url": "/contact.html", "cta_target_title": "Contact us"},
		"cta_url", contentHub{Name: "t", Title: "T", URL: "/tools/password-entropy.html"},
		contactSitePages(), "/index.html", "Get in Touch", nil, "")
	if got["cta_url"] != "/contact.html" || got["cta_target_title"] != "Contact us" {
		t.Errorf("authored url/title pair not kept together: %v", got)
	}
}

// TestApplyCTARecomputeStillRepairsFabricatedContactWithPageNamingLabel is bug
// 248's verification bar #2, and the reason the label match stays AHEAD of the
// authored keep: a fabricated contact fallback whose label names a real page is
// bugs_open/203's defect and must still be recomputed. Over-correcting here
// would trade one silent failure for another.
func TestApplyCTARecomputeStillRepairsFabricatedContactWithPageNamingLabel(t *testing.T) {
	got := map[string]interface{}{}
	applyCTARecompute(got, map[string]interface{}{"cta_url": "/contact.html"},
		"cta_url", contentHub{Name: "tool-password-entropy", Title: "Password Strength Physics", URL: "/tools/password-entropy.html"},
		contactSitePages(), "/index.html", "Run the Risk Checker", riskCheckerCandidates(t), "")
	if got["cta_url"] != "/tools/tool-ai-data-risk-checker.html" {
		t.Errorf("label naming a real page must still win over a stored contact url: %v", got)
	}
}

// TestSetCTAFieldKeepsAuthoredContactLink covers the BUILD-path half, which had
// no keep branch of any kind — so before this fix an authored contact link died
// on the next full regeneration even once the rerender path stopped clobbering
// it. The url must be WRITTEN, not merely skipped: a regeneration replaces the
// row, and a gated template renders no button at all for an absent url.
func TestSetCTAFieldKeepsAuthoredContactLink(t *testing.T) {
	stored := map[string]interface{}{
		"cta_url":          "/contact.html",
		"cta_target_title": "Contact us",
		"cta_text":         "Get in touch",
	}
	positionalTarget := contentHub{
		Name: "tool-risk-checker", Title: "AI Data Risk Checker",
		URL: "/tools/tool-ai-data-risk-checker.html",
	}

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, stored, "cta_url", positionalTarget, contactSitePages(),
		"hero", "hero", "primary", &unresolved, "Get in touch", riskCheckerCandidates(t), "")

	if resolved["cta_url"] != "/contact.html" {
		t.Errorf("authored contact link not kept on the build path: %v", resolved["cta_url"])
	}
	if resolved["cta_target_title"] != "Contact us" {
		t.Errorf("companion target title not carried: %v", resolved)
	}
	if len(unresolved) != 0 {
		t.Errorf("keeping an authored link must not report the field unresolved: %v", unresolved)
	}
}

// TestSetCTAFieldRederivesOrdinaryStoredDestination is the NEGATIVE CONTROL for
// the branch above. Without it, a fix that simply froze every stored CTA would
// pass every other test in this file. An ordinary stored destination must still
// be re-derived exactly as before.
func TestSetCTAFieldRederivesOrdinaryStoredDestination(t *testing.T) {
	stored := map[string]interface{}{"cta_url": "/tools/password-entropy.html"}
	positionalTarget := contentHub{
		Name: "tool-risk-checker", Title: "AI Data Risk Checker",
		URL: "/tools/tool-ai-data-risk-checker.html",
	}

	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, stored, "cta_url", positionalTarget, contactSitePages(),
		"hero", "hero", "primary", &unresolved, "Learn More", riskCheckerCandidates(t), "")

	if resolved["cta_url"] != "/tools/tool-ai-data-risk-checker.html" {
		t.Errorf("ordinary stored destination was frozen instead of re-derived: %v", resolved["cta_url"])
	}
}

// TestSetCTAFieldStillRepairsFabricatedContactWithPageNamingLabel mirrors the
// recompute-side ordering pin on the build path.
func TestSetCTAFieldStillRepairsFabricatedContactWithPageNamingLabel(t *testing.T) {
	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(resolved, map[string]interface{}{"cta_url": "/contact.html"}, "cta_url",
		contentHub{Name: "tool-password-entropy", Title: "Password Strength Physics", URL: "/tools/password-entropy.html"},
		contactSitePages(), "hero", "hero", "primary", &unresolved,
		"Run the Risk Checker", riskCheckerCandidates(t), "")

	if resolved["cta_url"] != "/tools/tool-ai-data-risk-checker.html" {
		t.Errorf("label naming a real page must still win on the build path: %v", resolved)
	}
}

// TestFreshPickRefusesUtilityWhileStoredUtilityIsKept is the ASYMMETRY PIN.
// Its name is unchanged so `git log -S` still finds the whole history, but
// assertion (b) is INVERTED by bugs_open/308 Phase B (2026-08-23) and the test
// has grown a fourth arm.
//
// WHAT IT PINNED, 2026-07-14 → 2026-08-22:
//
//	(a) a fresh POSITIONAL pick refuses a utility destination;
//	(b) the label-match CANDIDATE SET refuses one too, which is what made
//	    "the resolver cannot mint a utility url" exact rather than approximate;
//	(c) an already-STORED one is kept.
//
// (b) IS NOW RETIRED, deliberately and by owner ruling (2026-08-18), because it
// is the invariant that made bug 308 unfixable: the detector could name
// /contact.html as the repair for "Contact our supply team" and the repairer's
// candidate list could not contain one, so the repair ran, reported success and
// changed nothing — 188 findings, 99 of them on work items marked `complete`
// (measured 2026-08-23). The bug file's own bar #2 says editing this test is the
// signal that the invariant was consciously retired rather than quietly broken,
// and that it is only legitimate once provenance has landed. It has:
// __cta_minted shipped 2026-08-22 (LNK-035) and was verified at the artefact.
//
// (d) IS NEW, and is the half that never existed: the positional pick must not
// DISPLACE a utility destination either. Before Phase B it could not — nothing
// could store one except a person, and (c) covered that. Now the resolver mints
// them, and a minted one whose copy later goes generic reaches the keeps with
// its record saying "the machine wrote this", which is exactly the state that
// used to mean "recompute it away". Doing so would replace a working contact
// button with a tool page: bugs_open/248's damage, arriving through 308's fix.
func TestFreshPickRefusesUtilityWhileStoredUtilityIsKept(t *testing.T) {
	// A contact page that IS a section-index hub, so it reaches the loaders.
	// Four such pages exist fleet-wide, which is why the filters are URL-shape
	// tests and not page_type tests.
	contactHub := contentHub{Name: "contact-index", Title: "Contact us", URL: "/contact/index.html", Area: "contact", NavOrder: 1}
	servicesHub := contentHub{Name: "services", Title: "Services", URL: "/services.html", NavOrder: 5}
	valid := datahelpers.NewPageURLSet([]string{"/index.html", "/contact/index.html", "/services.html"})

	// (a) the positional pick refuses it, even ranked first by nav order.
	primary, _ := chooseCTATargets("", "index", nil, []contentHub{contactHub, servicesHub})
	if primary.URL == contactHub.URL {
		t.Errorf("fresh positional pick landed on a utility page: %v", primary.URL)
	}

	// (b) INVERTED: the shared label-match universe now OFFERS a utility page.
	// Asserted at the membership rule rather than the query, because that rule
	// is the only place a page can be excluded from candidacy — there is no
	// area predicate left in CTALabelUniverseSQL to assert the absence of.
	if _, ok := datahelpers.CTALabelCandidateRow("1", "contact", "Contact us", "", "/contact.html", "content", true); !ok {
		t.Error("the CTA label universe refused a contact page — bug 308's repair " +
			"target is unreachable again and the widening has been undone")
	}
	// …and the homepage is still not a candidate, which is the one exclusion
	// that survived the widening.
	if _, ok := datahelpers.CTALabelCandidateRow("2", "index", "Home", "", "/index.html", "content", true); ok {
		t.Error("the CTA label universe offered the homepage as a candidate")
	}

	// (c) an already-stored, UNSTAMPED one is kept — a person's link.
	got := map[string]interface{}{}
	applyCTARecompute(got, map[string]interface{}{"cta_url": "/contact/index.html"},
		"cta_url", servicesHub, valid, "/index.html", "Talk to us", nil, "")
	if got["cta_url"] != "/contact/index.html" {
		t.Errorf("stored utility destination not kept: %v", got)
	}

	// (d) a MINTED one is kept too, on BOTH writers. The mint record is what
	// makes storedCTADestinationIsAuthored decline, so this row reaches the
	// later keeps — and before Phase B's keep changes it reached the positional
	// pick instead and was overwritten with /services.html.
	stored := map[string]interface{}{
		"cta_url":                "/contact/index.html",
		datahelpers.CTAMintedKey: map[string]interface{}{"cta_url": "/contact/index.html"},
	}
	if storedCTADestinationIsAuthored(stored, "cta_url", valid) {
		t.Fatal("fixture is wrong: a minted destination must NOT read as authored, " +
			"or this case is testing arm (c) again")
	}

	gotRecompute := map[string]interface{}{}
	applyCTARecompute(gotRecompute, stored, "cta_url", servicesHub, valid, "/index.html", "Get Started", nil, "")
	if u, _ := gotRecompute["cta_url"].(string); u != "" && u != "/contact/index.html" {
		t.Errorf("repair path DISPLACED a minted utility destination with the positional pick: %v", gotRecompute)
	}

	gotBuild := map[string]interface{}{}
	var unresolved []map[string]interface{}
	setCTAField(gotBuild, stored, "cta_url", servicesHub, valid, "hero", "hero", "primary", &unresolved,
		"Get Started", nil, "")
	if gotBuild["cta_url"] != "/contact/index.html" {
		t.Errorf("build path DISPLACED a minted utility destination with the positional pick: %v", gotBuild)
	}
}

// ── bugs_open/308 Phase A: the WIRING tests ────────────────────────────────
//
// These exist because the unit tests in datahelpers/cta_provenance_test.go
// exercise the helpers directly, and mutation showed that is not enough: with
// the seed and the carry called from the CALLER'S LOOP, deleting either call
// site left every test in the tree green. That is the "a helper with no callers
// looks like a finished refactor" shape. Both calls were moved INSIDE
// setCTAField/applyCTARecompute so they sit on the path of these tests, and
// these tests are what make removing them fail.

// TestSetCTAFieldPreservesSiblingSlotStamp — the freeze this fix would otherwise
// have introduced. Both persist paths merge SHALLOWLY and the mint record is a
// nested map, so writing a record for the primary slot alone REPLACES the stored
// record and drops the secondary's stamp. The secondary then reads as authored
// on the next pass and is frozen at whatever it currently points to.
//
// MUTATION: delete the SeedCTAMinted call inside setCTAField. This fails.
func TestSetCTAFieldPreservesSiblingSlotStamp(t *testing.T) {
	valid := contactSitePages()
	stored := map[string]interface{}{
		"cta_url":           "/tools/password-entropy.html",
		"secondary_cta_url": "/services.html",
		datahelpers.CTAMintedKey: map[string]interface{}{
			"cta_url":           "/tools/password-entropy.html",
			"secondary_cta_url": "/services.html",
		},
	}
	resolved := map[string]interface{}{}
	var unresolved []map[string]interface{}
	target := contentHub{Name: "risk-checker", Title: "Risk Checker",
		URL: "/tools/tool-ai-data-risk-checker.html"}

	// Only the PRIMARY slot is processed this pass.
	setCTAField(resolved, stored, "cta_url", target, valid,
		"hero", "hero", "primary", &unresolved, "", nil, "")

	if !datahelpers.CTAMintedCovers(resolved, "secondary_cta_url", "/services.html") {
		t.Error("the untouched secondary slot lost its mint record — it will read as authored " +
			"next pass and freeze, which is the bug this record exists to end")
	}
	if !datahelpers.CTAMintedCovers(resolved, "cta_url", "/tools/tool-ai-data-risk-checker.html") {
		t.Error("the primary slot must carry a record for the value written THIS pass")
	}
}

// TestSetCTAFieldCarriesMintAtUnresolvedFallthrough — setCTAField's last branch
// writes no url at all, so a value left in resolved_data by plan_sections'
// PBP-039 carry is what gets persisted; and the plan->save funnel REPLACES
// content_data, so the previous generation's record does not survive on its own.
// Without the carry the value arrives unstamped, reads as authored, and freezes.
//
// MUTATION: no-op the SeedCTAMinted call at the top of setCTAField. This fails.
// (A second, value-guarded re-stamp AT this branch was written first and then
// deleted: mutation showed removing it changed no test, because the seed had
// already carried the record — two guards in series, one of them dead.)
func TestSetCTAFieldCarriesMintAtUnresolvedFallthrough(t *testing.T) {
	valid := contactSitePages()
	stored := map[string]interface{}{
		"cta_url": "/tools/password-entropy.html",
		datahelpers.CTAMintedKey: map[string]interface{}{
			"cta_url": "/tools/password-entropy.html",
		},
	}
	// The carry has already put the previously-minted url here; this pass has no
	// valid positional target, so setCTAField writes nothing.
	resolved := map[string]interface{}{"cta_url": "/tools/password-entropy.html"}
	var unresolved []map[string]interface{}

	setCTAField(resolved, stored, "cta_url", contentHub{}, valid,
		"hero", "hero", "primary", &unresolved, "", nil, "")

	if len(unresolved) != 1 {
		t.Fatalf("expected the unresolved fallthrough to be the branch under test, got %d entries", len(unresolved))
	}
	if !datahelpers.CTAMintedCovers(resolved, "cta_url", "/tools/password-entropy.html") {
		t.Error("a carried, previously-minted value lost its record at the unresolved fallthrough — " +
			"it becomes indistinguishable from an authored link and freezes")
	}
}

// TestSetCTAFieldInventsNoProvenanceForAnAuthoredValue is the control for the
// test above, and it is the one that matters if the carry is ever loosened: an
// AUTHORED value sitting in resolved_data at the fallthrough must NOT acquire a
// record, or the resolver would license itself to recompute a person's link.
//
// It pins an ABSENCE — that no provenance is invented — so name the mutation
// that actually kills it, because the obvious guess does not. Making
// CTAMintedCovers presence-bound leaves this test GREEN (checked, not assumed:
// with no record on the row at all, Covers returns false at its nil-map guard
// long before the comparison). The mutation that DOES kill it is
// SeedCTAMinted inventing a record from whatever `resolved` already holds when
// the stored row has none — verified by running it. That is the realistic
// mistake here: "the field has a value, so stamp it" is one plausible reading
// of what seeding means, and it would quietly relabel every authored link as
// the resolver's own and make it recomputable.
func TestSetCTAFieldInventsNoProvenanceForAnAuthoredValue(t *testing.T) {
	valid := contactSitePages()
	stored := map[string]interface{}{"cta_url": "/contact.html"} // authored: no record
	resolved := map[string]interface{}{"cta_url": "/contact.html"}
	var unresolved []map[string]interface{}

	setCTAField(resolved, stored, "cta_url", contentHub{}, valid,
		"hero", "hero", "primary", &unresolved, "", nil, "")

	if datahelpers.CTAMintedCovers(resolved, "cta_url", "/contact.html") {
		t.Error("provenance was manufactured for an authored value — the resolver would then be " +
			"free to recompute a link a person chose")
	}
}

// TestApplyCTARecomputePreservesSiblingSlotStamp — the repair-path twin of the
// first test. The two writers are two halves of one seam and bugs_open/248 was
// caused by only one of them being fixed, so the sibling-stamp property is
// pinned on both.
//
// MUTATION: delete the SeedCTAMinted call inside applyCTARecompute. This fails.
func TestApplyCTARecomputePreservesSiblingSlotStamp(t *testing.T) {
	valid := contactSitePages()
	stored := map[string]interface{}{
		"cta_url":           "/tools/password-entropy.html",
		"secondary_cta_url": "/services.html",
		datahelpers.CTAMintedKey: map[string]interface{}{
			"cta_url":           "/tools/password-entropy.html",
			"secondary_cta_url": "/services.html",
		},
	}
	resolved := map[string]interface{}{}
	target := contentHub{Name: "risk-checker", Title: "Risk Checker",
		URL: "/tools/tool-ai-data-risk-checker.html"}

	applyCTARecompute(resolved, stored, "cta_url", target, valid, "/index.html", "", nil, "")

	if !datahelpers.CTAMintedCovers(resolved, "secondary_cta_url", "/services.html") {
		t.Error("the untouched secondary slot lost its mint record on the repair path")
	}
}
