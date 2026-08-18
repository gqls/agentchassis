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
		if got := storedCTADestinationIsAuthored(c.url, valid); got != c.want {
			t.Errorf("storedCTADestinationIsAuthored(%q) = %v, want %v — %s", c.url, got, c.want, c.why)
		}
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
		"cta_url", positionalTarget, valid, "/index.html", genericContactLabel, candidates)
	if len(control) != 0 {
		t.Errorf("CONTROL: ordinary stored destination should be left untouched, got %v", control)
	}

	// CASE: the same call with an authored contact destination. Kept.
	got := map[string]interface{}{}
	applyCTARecompute(got, map[string]interface{}{"cta_url": "/contact.html"},
		"cta_url", positionalTarget, valid, "/index.html", genericContactLabel, candidates)
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
		contactSitePages(), "/index.html", "Get in Touch", nil)
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
		contactSitePages(), "/index.html", "Run the Risk Checker", riskCheckerCandidates(t))
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
		"hero", "hero", "primary", &unresolved, "Get in touch", riskCheckerCandidates(t))

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
		"hero", "hero", "primary", &unresolved, "Learn More", riskCheckerCandidates(t))

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
		"Run the Risk Checker", riskCheckerCandidates(t))

	if resolved["cta_url"] != "/tools/tool-ai-data-risk-checker.html" {
		t.Errorf("label naming a real page must still win on the build path: %v", resolved)
	}
}

// TestFreshPickRefusesUtilityWhileStoredUtilityIsKept is the ASYMMETRY PIN, and
// the point of the whole change. Three assertions that must hold together:
//
//	(a) a fresh POSITIONAL pick still refuses a utility destination;
//	(b) the label-match CANDIDATE SET refuses one too, which is what makes
//	    "the resolver cannot mint a utility url" exact rather than approximate;
//	(c) an already-STORED one is kept.
//
// If someone later re-unifies these two uses of areasExcludedFromCTA — the
// natural-looking tidy-up, and the original 2026-07-14 design note explicitly
// asked for it ("reuse it in chooseCTATargets' area filter too so the two
// agree") — this test goes red rather than 13 live contact buttons going quiet.
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

	// (b) the label-match candidate set refuses it too.
	for _, c := range candidatesFromHubs(nil, []contentHub{contactHub, servicesHub}) {
		if c.URL == contactHub.URL {
			t.Errorf("label-match candidates offered a utility page (%v) — the resolver can now mint "+
				"the very urls storedCTADestinationIsAuthored treats as authored", c.URL)
		}
	}

	// (c) an already-stored one is kept.
	got := map[string]interface{}{}
	applyCTARecompute(got, map[string]interface{}{"cta_url": "/contact/index.html"},
		"cta_url", servicesHub, valid, "/index.html", "Talk to us", nil)
	if got["cta_url"] != "/contact/index.html" {
		t.Errorf("stored utility destination not kept: %v", got)
	}
}
