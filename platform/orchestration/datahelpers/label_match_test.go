// FILE: platform/orchestration/datahelpers/label_match_test.go
//
// Moved from discovery_checks/check_misdirected_cta_test.go unchanged in
// substance — LabelTokens/BestLabelMatch are ctaTokens/bestPageMatch,
// extracted here so the audit-time detector and the write-time resolver
// share one definition (bugs_open/203 follow-on).

package datahelpers

import (
	"reflect"
	"testing"
)

func TestLabelTokens(t *testing.T) {
	cases := map[string][]string{
		"Enter the Gauntlet":  {"gauntlet"},
		"Take the Quiz":       {"quiz"},
		"Enter today's Arena": {"arena"},
		"Learn More":          nil, // fully generic — names nothing
		"Get Started":         nil,
		"Meet the Archetypes": {"archetypes"},
		"  ":                  nil,
		"Read our FAQ now":    {"faq"},
	}
	for text, want := range cases {
		got := LabelTokens(text)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("LabelTokens(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestBestLabelMatch(t *testing.T) {
	mk := func(name, title string, interactive bool) LabelMatchCandidate {
		c, ok := NewLabelMatchCandidate(name, name, title, "/"+name+".html", interactive, "")
		if !ok {
			t.Fatalf("NewLabelMatchCandidate(%q) unexpectedly produced no tokens", name)
		}
		return c
	}
	pages := []LabelMatchCandidate{
		mk("tool-gauntlet", "The Gauntlet", true),
		mk("tool-quiz", "Archetype Quiz", true),
		mk("archetypes", "The Archetypes", false),
	}

	if got, ok := BestLabelMatch("gauntlet", pages); !ok || got.Name != "tool-gauntlet" {
		t.Errorf("gauntlet: got %+v ok=%v, want tool-gauntlet", got, ok)
	}
	// Interactive pages beat content pages on equal-strength matches:
	// "archetype" hits the quiz (interactive), "archetypes" hits the hub.
	if got, ok := BestLabelMatch("archetype", pages); !ok || got.Name != "tool-quiz" {
		t.Errorf("archetype: got %+v ok=%v, want tool-quiz", got, ok)
	}
	if got, ok := BestLabelMatch("archetypes", pages); !ok || got.Name != "archetypes" {
		t.Errorf("archetypes: got %+v ok=%v, want archetypes hub", got, ok)
	}
	if _, ok := BestLabelMatch("arena", pages); ok {
		t.Errorf("arena: got a match, want none (no page names it)")
	}
	if _, ok := BestLabelMatch("Learn More", pages); ok {
		t.Errorf("generic text: got a match, want none")
	}
}

// TestBestLabelMatchOverlapBeatsCategory reproduces the live defect found on
// robot-hands.com (bugs_open/203 follow-on, NOTES 2026-08-09): a hub page with
// a STRONGER token match must win over a tool/game page with a WEAKER one —
// interactive-preference is a tie-break, not an override. Before the fix, this
// failed: interactive beat non-interactive regardless of overlap count.
func TestBestLabelMatchOverlapBeatsCategory(t *testing.T) {
	mk := func(name, title string, interactive bool) LabelMatchCandidate {
		c, ok := NewLabelMatchCandidate(name, name, title, "/"+name+".html", interactive, "")
		if !ok {
			t.Fatalf("NewLabelMatchCandidate(%q) unexpectedly produced no tokens", name)
		}
		return c
	}
	pages := []LabelMatchCandidate{
		// 1-token overlap with "Browse the Gripper Catalog": "gripper" only.
		mk("tool-gripper-cycle-time-estimator", "Gripper Cycle Time Estimator", true),
		// 2-token overlap: "gripper" and "catalog" — the better match, non-interactive.
		mk("gripper-catalog-index", "Gripper Catalog", false),
	}
	got, ok := BestLabelMatch("Browse the Gripper Catalog", pages)
	if !ok || got.Name != "gripper-catalog-index" {
		t.Errorf("Browse the Gripper Catalog: got %+v ok=%v, want the stronger hub match gripper-catalog-index, not the weaker tool match", got, ok)
	}
}

// TestBestLabelMatchInteractiveTiesBreakToInteractive covers the genuine
// equal-strength case BestLabelMatch's own doc comment describes: same
// overlap count, category decides.
func TestBestLabelMatchInteractiveTiesBreakToInteractive(t *testing.T) {
	mk := func(name, title string, interactive bool) LabelMatchCandidate {
		c, ok := NewLabelMatchCandidate(name, name, title, "/"+name+".html", interactive, "")
		if !ok {
			t.Fatalf("NewLabelMatchCandidate(%q) unexpectedly produced no tokens", name)
		}
		return c
	}
	pages := []LabelMatchCandidate{
		mk("tool-widget", "Widget Tool", true),
		mk("widget-guide", "Widget Guide", false),
	}
	got, ok := BestLabelMatch("Widget", pages)
	if !ok || got.Name != "tool-widget" {
		t.Errorf("Widget (equal overlap both candidates): got %+v ok=%v, want the interactive tool-widget", got, ok)
	}
}

func TestNewLabelMatchCandidateRejectsAllGenericSources(t *testing.T) {
	// name/title/nav all reduce to zero non-stopword tokens: "learn" and
	// "more" are both stopwords.
	if _, ok := NewLabelMatchCandidate("x", "learn-more", "Learn More", "/x.html", false, ""); ok {
		t.Errorf("expected ok=false when every token source is generic/empty")
	}
}

// TestBestLabelMatchIdentityBeatsDescription is the bugs_open/253 live trio,
// verbatim rows from robot-hands.com (2026-08-11). Before this fix, the label
// "Gripper Safety Factor Calculator" resolved to
// tool-gripper-payload-calculator: its nav_label ("…Validate Capacity with
// Safety Factor…") happened to contain "safety" and "factor", tying total
// overlap with the page the label actually names, and the alphabetical
// tie-break broke that tie the wrong way. Identity (name/title) tokens must
// be compared first: only tool-gripper-safety-factor-calculator's own
// name/title contain "safety" and "factor", so it wins outright on
// identityOverlap before totalOverlap (which ties all three candidates at 4)
// is ever consulted.
func TestBestLabelMatchIdentityBeatsDescription(t *testing.T) {
	payloadOverview, ok := NewLabelMatchCandidate(
		"1", "gripper-payload-calculator",
		"Gripper Payload Calculator — Overview | Robot-Hands.com",
		"/gripper-payload-calculator.html", false,
		"Gripper Payload Calculator — Calculate Required Grip Force with Safety Factor | Robot-Hands.com")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	payloadTool, ok := NewLabelMatchCandidate(
		"2", "tool-gripper-payload-calculator",
		"Gripper Payload Calculator | Robot-Hands.com",
		"/tools/gripper-payload-calculator/index.html", true,
		"Gripper Payload Calculator — Validate Capacity with Safety Factor | Robot-Hands.com")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	safetyFactorTool, ok := NewLabelMatchCandidate(
		"3", "tool-gripper-safety-factor-calculator",
		"Gripper Safety Factor Calculator | Tools",
		"/tools/gripper-safety-factor-calculator/index.html", true, "")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	pages := []LabelMatchCandidate{payloadOverview, payloadTool, safetyFactorTool}

	got, ok := BestLabelMatch("Gripper Safety Factor Calculator", pages)
	if !ok || got.URL != "/tools/gripper-safety-factor-calculator/index.html" {
		t.Errorf("Gripper Safety Factor Calculator: got %+v ok=%v, want the page it names, not a payload-calculator page whose nav_label incidentally contains its words", got, ok)
	}
}

// TestBestLabelMatchRichNameBeatsShortGenericPage guards against a future
// precision-style scoring that advantages tiny token sets outright: a page
// with MORE identity tokens matching the label must still win over a page
// with fewer, even though the smaller page's single token is a larger
// fraction of its own token set.
func TestBestLabelMatchRichNameBeatsShortGenericPage(t *testing.T) {
	tools, ok := NewLabelMatchCandidate("1", "tools", "Tools", "/tools.html", false, "")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	toolsGuide, ok := NewLabelMatchCandidate("2", "tools-guide", "Tools Guide", "/tools-guide.html", false, "")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	pages := []LabelMatchCandidate{tools, toolsGuide}

	got, ok := BestLabelMatch("Tools Guide", pages)
	if !ok || got.Name != "tools-guide" {
		t.Errorf("Tools Guide: got %+v ok=%v, want tools-guide (identity overlap 2 beats 1)", got, ok)
	}
}

// TestBestLabelMatchTiesStillBreakByName pins the FINAL tie-break at Name —
// unchanged from before the bugs_open/253 fix. A candidate-token-set-size
// tie-break (smaller wins) was tried here and DROPPED before shipping:
// 2026-08-11 calibration against the live fleet showed it was decided almost
// entirely by tokenisation artefacts and site-wide generic words with no real
// signal (a stray hyphen in one candidate's own copy flipped 9 already-
// correct, live gaswholesalers.com CTAs onto the wrong tool purely because
// the loser's title happened to be one token longer — see
// CALIBRATION_2026-08-11_label_match_identity_report.txt). This test exists
// so a future re-introduction of a size-based tie-break has to consciously
// break this assertion, not slip in unnoticed.
func TestBestLabelMatchTiesStillBreakByName(t *testing.T) {
	big, ok := NewLabelMatchCandidate(
		"1", "aaa-gadget-long-name-thing", "Aaa Gadget Long Name Thing",
		"/aaa-gadget-long-name-thing.html", false, "")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	small, ok := NewLabelMatchCandidate(
		"2", "zz-gadget", "Zz Gadget", "/zz-gadget.html", false, "")
	if !ok {
		t.Fatal("fixture candidate produced no tokens")
	}
	pages := []LabelMatchCandidate{big, small}

	got, ok := BestLabelMatch("gadget", pages)
	if !ok || got.Name != "aaa-gadget-long-name-thing" {
		t.Errorf("gadget: got %+v ok=%v, want aaa-gadget-long-name-thing (both tie on identity/total overlap and interactivity, so Name decides — alphabetically first, regardless of either candidate's token-set size)", got, ok)
	}
}
