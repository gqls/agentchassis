// FILE: platform/orchestration/actions/adopt_verbatim_identity_convergence_test.go
//
// bugs_open/080, surface: adopt_verbatim — deliberately EXEMPT from routing
// through CanonicalisePage, because preserving the crawled URL is the entire
// feature (LANDMINES.md records that pushing adopted pages through the
// canonicaliser, which synthesises fresh URLs, is precisely the trap that
// restyles an adopted site).
//
// What is NOT exempt is the NAME half: verbatimPageIdentity's own doc comment
// says "Names mirror CanonicalisePage's prefixes … so an adopted page is named
// the way a planner-built page of the same kind would be". That mirror is a
// hand-maintained copy of the canonical convention, and a hand-maintained copy
// is the drift class this estate keeps getting bitten by. This test pins it:
// every name verbatimPageIdentity produces must be a FIXED POINT of
// CanonicalisePage under its own page_type — feeding the (name, type) back
// through the canonical helper must return the name unchanged. If either side
// moves its convention, this fails and the drift is caught at build time.
//
// URLs are deliberately not asserted: "Only the URL differs, which is the
// whole point of this path."
package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestVerbatimPageIdentityNamesAreCanonicalFixedPoints(t *testing.T) {
	cases := []struct {
		deployPath string
		wantName   string
		wantType   string
	}{
		{"/index.html", "index", "landing"},
		{"/about.html", "about", "content"},
		{"/tools/repayment.html", "tool-repayment", "tool"},
		{"/tools/repayment/index.html", "tool-repayment", "tool"},
		{"/tools/index.html", "tools-index", "section-index"},
		{"/guides/overpaying.html", "guide-overpaying", "guide"},
		{"/guides/index.html", "guides-index", "section-index"},
		{"/services/consulting.html", "services-consulting", "content"},
	}

	for _, tc := range cases {
		name, pageType, ok := verbatimPageIdentity(tc.deployPath)
		if !ok {
			t.Errorf("verbatimPageIdentity(%q): not ok", tc.deployPath)
			continue
		}
		if name != tc.wantName || pageType != tc.wantType {
			t.Errorf("verbatimPageIdentity(%q) = (%q, %q), want (%q, %q)",
				tc.deployPath, name, pageType, tc.wantName, tc.wantType)
			continue
		}

		// The fixed-point assertion — the actual convergence pin.
		canonName, _, canonType := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
			Role: pageType,
			Slug: name,
		})
		if canonName != name {
			t.Errorf("drift: verbatim name %q (from %q) canonicalises to %q — the mirror in verbatimPageIdentity no longer matches CanonicalisePage",
				name, tc.deployPath, canonName)
		}
		if canonType != pageType {
			t.Errorf("drift: verbatim page_type %q (from %q) canonicalises to %q",
				pageType, tc.deployPath, canonType)
		}
	}
}

// TestCompanionGuideIdentityMatchesLegacyShape pins surface C's enforcement
// change: for every existing guide the helper output is byte-identical to the
// old hand-rolled sprintf, so routing through it moves nothing.
func TestCompanionGuideIdentityMatchesLegacyShape(t *testing.T) {
	for _, toolPageName := range []string{
		"tool-ab-test-calculator", // canonical tool page name
		"password-entropy",        // legacy tool page name
	} {
		name, url, err := companionGuideIdentity(toolPageName)
		if err != nil {
			t.Fatalf("companionGuideIdentity(%q): %v", toolPageName, err)
		}
		wantName := toolPageName + "-guide"
		wantURL := "/guides/" + wantName + ".html"
		if name != wantName || url != wantURL {
			t.Errorf("companionGuideIdentity(%q) = (%q, %q), want (%q, %q)",
				toolPageName, name, url, wantName, wantURL)
		}
	}
}
