// FILE: platform/orchestration/actions/tool_meta_description_339_test.go
//
// bugs_open/339. bugs_closed/103's guard (datahelpers.PublicMetaDescription) was
// re-measured on 2026-08-20 and found blind to the CURRENT brief population: its
// length threshold (320) was derived from a July census of 449-1,206-char briefs,
// and today's briefs run 200-320; its briefMarkers regex, the designed backstop
// for short briefs, matched ZERO of the nine live tool pages publishing their
// build brief as the public meta description.
//
// The fix is not a cleverer guard — a guard re-fitted to today's brief style goes
// stale the same way the July one did. It is candidate 3 of the 339 file: the two
// tool-page call sites stop offering the brief as a candidate AT ALL (the empty
// string goes in its place), because for a component_level='tool' row the
// description IS the brief by definition. These tests pin both arms of that.
package actions

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// liveBriefBand296 is VERBATIM from pages.meta_description for
// webdesign.co.uk/tool-css-unit-converter as served on 2026-08-20 — a 296-char
// build brief the guard passed to the public. It is the measured premise of the
// call-site closure.
const liveBriefBand296 = "Converts between px, rem, em, vw, and vh units given a base font size and viewport dimensions. A daily friction point for developers implementing fluid layouts or translating design specs — surprisingly absent from the existing toolset despite the fluid typography and grid tools already present."

// TestComposedToolMetaDescriptionSurvivesItsOwnVetting is the ACCEPTING arm: the
// only text the tool call sites now publish is the composed line, and
// PublicMetaDescription vets the composed side too (a brief-shaped composed line
// yields ""), so the composed helper's output must pass that vetting or every
// tool page ships with no description at all.
func TestComposedToolMetaDescriptionSurvivesItsOwnVetting(t *testing.T) {
	for _, name := range []string{
		"CSS Unit Converter",
		"UK CSS Unit Converter | Tools", // the suffix/prefix-stripping path
		"Aria Builder",
	} {
		composed := composedToolMetaDescription(name)
		if strings.TrimSpace(composed) == "" {
			t.Fatalf("composedToolMetaDescription(%q) is empty — every tool page would publish no description", name)
		}
		if datahelpers.MetaDescriptionLooksInternal(composed) {
			t.Errorf("composedToolMetaDescription(%q) = %q reads as internal to the guard — the composed fallback would be refused and tool pages would publish nothing", name, composed)
		}
		got, replaced := datahelpers.PublicMetaDescription("", composed)
		if got != composed {
			t.Errorf("PublicMetaDescription(\"\", composed) = %q, want the composed line %q — the empty-candidate convention the two tool call sites rely on has changed", got, composed)
		}
		if replaced {
			t.Errorf("PublicMetaDescription(\"\", composed) reported replaced=true — an empty candidate is 'nothing offered', not 'something rejected', and callers may log on this flag")
		}
	}
}

// TestGuardIsStillBlindToTheBriefBand pins the MEASURED PREMISE of the call-site
// closure: the live 296-char brief passes the guard, which is why the guard alone
// is not protection and the candidate slot is closed at the call sites.
//
// If this test ever FAILS, that is good news — the guard has learned to catch the
// brief band — but it does NOT reopen the candidate slot: the closure exists
// because brief style tracks brief authors and any re-fitted signal can go blind
// again (bugs_open/339 §5). Delete this test with thanks; do not revert the call
// sites.
func TestGuardIsStillBlindToTheBriefBand(t *testing.T) {
	if datahelpers.MetaDescriptionLooksInternal(liveBriefBand296) {
		t.Errorf("MetaDescriptionLooksInternal now catches the 296-char live brief — the 339 premise has changed. " +
			"This is an improvement to the guard, not a reason to reopen the tool call sites' candidate slot; " +
			"read the comment above this test before touching create_tool_component_action.go or deploy_tool_action.go")
	}
	got, _ := datahelpers.PublicMetaDescription(liveBriefBand296, composedToolMetaDescription("CSS Unit Converter"))
	if got != liveBriefBand296 {
		t.Errorf("PublicMetaDescription would no longer publish the live brief when offered it — see the message above; the premise moved")
	}
}

// TestToolCallSitesNeverOfferTheBriefAsACandidate scans the two files that create
// tool pages and requires every PublicMetaDescription call to pass the EMPTY
// STRING as its candidate. This is the check whose absence let bugs_closed/103
// regress invisibly: both sites were "guarded", and the guard was blind.
//
// Source-scan caveats handled per update_page_status_config_contract_test.go:
// comment lines are skipped, and finding ZERO calls is a loud failure (a broken
// scan must not pass silently).
func TestToolCallSitesNeverOfferTheBriefAsACandidate(t *testing.T) {
	callRe := regexp.MustCompile(`datahelpers\.PublicMetaDescription\(`)
	for _, file := range []string{
		"create_tool_component_action.go",
		"deploy_tool_action.go",
	} {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("cannot read %s: %v", file, err)
		}
		lines := strings.Split(string(src), "\n")
		calls := 0
		for i, l := range lines {
			if strings.HasPrefix(strings.TrimSpace(l), "//") {
				continue
			}
			if !callRe.MatchString(l) {
				continue
			}
			calls++
			// The candidate is either on the same line after the paren, or on
			// the next non-comment line (the current formatting).
			rest := l[strings.Index(l, "PublicMetaDescription(")+len("PublicMetaDescription("):]
			candidateLine := strings.TrimSpace(rest)
			for j := i + 1; candidateLine == "" && j < len(lines); j++ {
				candidateLine = strings.TrimSpace(lines[j])
				if strings.HasPrefix(candidateLine, "//") {
					candidateLine = ""
				}
			}
			if !strings.HasPrefix(candidateLine, `"",`) && !strings.HasPrefix(candidateLine, `""`) {
				t.Errorf("%s line %d: PublicMetaDescription's candidate is %q, not the empty string — a component_level='tool' description is the BUILD BRIEF by definition and must never be offered as public copy (bugs_open/339)", file, i+1, candidateLine)
			}
		}
		if calls == 0 {
			t.Errorf("%s: found no PublicMetaDescription calls — the scan is broken or the call moved; this test is asserting nothing", file)
		}
	}
}
