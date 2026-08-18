// FILE: platform/orchestration/actions/discovery_checks/check_cta_nonpage_test.go
//
// Pins classifyNonPageAnchor — the whole per-anchor decision of the
// cta_nonpage_destination check (bugs_open/299). The first case is the
// motivating live defect verbatim: webdesign.uk's home-page CTA whose copy
// named the Website Brief Starter (and, after the 08-18 rewrite, how-it-works)
// while the href dialled the phone — invisible to check_misdirected_cta by
// construction, because its loop skips non-page scopes before classification.
//
// Also pins the SQL lockstep: both CTA checks must scan through
// ctaComponentScanQuery, and that query must exclude deleted/archived pages
// (the index-rejected-v1 doomed-item leak) — asserted on the constant, which
// no behaviour change can satisfy by accident.

package discovery_checks

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// nonPageTestPages: the webdesign.uk shape — a tool page and a content page,
// both real. Names/titles follow the live rows.
func nonPageTestPages(t *testing.T) []datahelpers.LabelMatchCandidate {
	t.Helper()
	briefStarter, ok := datahelpers.NewLabelMatchCandidate(
		"1", "tool-website-brief-starter", "Website Brief Starter | Tools",
		"/tools/website-brief-starter/index.html", true, "")
	if !ok {
		t.Fatal("brief-starter fixture produced no tokens")
	}
	howItWorks, ok := datahelpers.NewLabelMatchCandidate(
		"2", "how-it-works", "How It Works", "/how-it-works.html", false, "")
	if !ok {
		t.Fatal("how-it-works fixture produced no tokens")
	}
	return []datahelpers.LabelMatchCandidate{briefStarter, howItWorks}
}

func TestClassifyNonPageAnchor(t *testing.T) {
	pages := func(t *testing.T) []datahelpers.LabelMatchCandidate { return nonPageTestPages(t) }

	tests := []struct {
		name       string
		text, href string
		wantKind   string // "" = no finding
		wantTarget string
	}{
		{
			// bugs_open/299 as filed: the copy names the tool, the href dials.
			name:     "copy names the tool, href dials the phone",
			text:     "Or answer a couple of quick questions first with the Website Brief Starter, a tool that helps you set out what you need before we talk.",
			href:     "tel:+44 (0) 7934 524 911",
			wantKind: "cta_names_nonpage_destination", wantTarget: "/tools/website-brief-starter/index.html",
		},
		{
			// bugs_open/299 after the 08-18 rewrite: one distinctive token
			// ("works") is enough — a two-token minimum would miss it.
			name: "single-token copy names a page, href dials",
			text: "See how it works", href: "tel:+44 (0) 7934 524 911",
			wantKind: "cta_names_nonpage_destination", wantTarget: "/how-it-works.html",
		},
		{
			// External is round 1's STATED residue: the calibration measured
			// ~211/226 false positives with it in scope (news headlines,
			// regulator links). Detected = nothing, deliberately.
			name: "external href is out of round-1 scope",
			text: "Read the Website Brief Starter guide", href: "https://example.org/elsewhere",
			wantKind: "",
		},
		{
			// Self-agreement: text that IS the address can never be a
			// misdirect, whatever its tokens score against page titles —
			// the ai-agent-orchestration privacy-page class from the
			// calibration corpus.
			name: "mailto whose text contains its own address",
			text: "Email us at hello@webdesign.uk today", href: "mailto:hello@webdesign.uk",
			wantKind: "",
		},
		{
			// Self-agreement, tel form: display digits and URI digits differ
			// legitimately (trunk "(0)", separators) — matched on the digit
			// tail, so a phone button whose copy states its number is clean
			// even on a site with a "book-a-call" page.
			name: "tel whose text states the dialled number",
			text: "Call us on +44 (0) 7934 524 911 to see how it works", href: "tel:+447934524911",
			wantKind: "",
		},
		{
			// A GENUINE phone button: generic phone copy names no page.
			// This is the false-positive guard — faq/how-it-works carry
			// exactly this shape live, well-formed after normalisation.
			name: "genuine phone button, well-formed",
			text: "Call us on +44 (0) 7934 524 911", href: "tel:+447934524911",
			wantKind: "",
		},
		{
			name: "genuine phone button, malformed separators",
			text: "Call us on +44 (0) 7934 524 911", href: "tel:+44 (0) 7934 524 911",
			wantKind: "cta_tel_malformed",
		},
		{
			// The live undialable form: normaliser refuses, a human owns it.
			name: "collapsed trunk prefix",
			text: "Call us", href: "tel:+4407934524911",
			wantKind: "cta_tel_malformed",
		},
		{
			name: "genuine email button",
			text: "Get in Touch", href: "mailto:hello@webdesign.uk",
			wantKind: "",
		},
		{
			// Not this check's remit — page scopes belong to the sibling…
			name: "page href is the sibling's remit",
			text: "See how it works", href: "/contact.html",
			wantKind: "",
		},
		{
			// …and dead controls belong to check_dead_controls.
			name: "javascript href is dead_controls' remit",
			text: "See how it works", href: "javascript:void(0)",
			wantKind: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ok := classifyNonPageAnchor(
				datahelpers.Anchor{Text: tt.text, Href: tt.href}, "call-to-action", pages(t))
			if tt.wantKind == "" {
				if ok {
					t.Fatalf("want no finding, got %+v", f)
				}
				return
			}
			if !ok {
				t.Fatalf("want a %s finding, got none", tt.wantKind)
			}
			if f.Kind != tt.wantKind {
				t.Errorf("kind = %s, want %s", f.Kind, tt.wantKind)
			}
			if tt.wantTarget != "" && f.SuggestedTarget != tt.wantTarget {
				t.Errorf("suggested_target = %s, want %s", f.SuggestedTarget, tt.wantTarget)
			}
		})
	}
}

// TestCTAScanQueryExcludesRetiredPages pins the lifecycle filter on the shared
// scan constant. Asserted on the SQL text: a predicate can only disappear by
// an edit to the constant itself, which this test then names. The spelling
// must match loadCTAMatchIndex's (the actions-package shared constant is
// unreachable from here — import cycle).
func TestCTAScanQueryExcludesRetiredPages(t *testing.T) {
	const predicate = `p.status NOT IN ('deleted', 'archived')`
	if !strings.Contains(ctaComponentScanQuery, predicate) {
		t.Fatalf("ctaComponentScanQuery lost its page-lifecycle filter %q — archived pages "+
			"mint doomed work items (bugs_open/299: index-rejected-v1 filed two failed "+
			"cta_links_stale rerenders)", predicate)
	}
}
