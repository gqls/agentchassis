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

	if got, ok, _ := BestLabelMatch("gauntlet", pages); !ok || got.Name != "tool-gauntlet" {
		t.Errorf("gauntlet: got %+v ok=%v, want tool-gauntlet", got, ok)
	}
	// Interactive pages beat content pages on equal-strength matches:
	// "archetype" hits the quiz (interactive), "archetypes" hits the hub.
	if got, ok, _ := BestLabelMatch("archetype", pages); !ok || got.Name != "tool-quiz" {
		t.Errorf("archetype: got %+v ok=%v, want tool-quiz", got, ok)
	}
	if got, ok, _ := BestLabelMatch("archetypes", pages); !ok || got.Name != "archetypes" {
		t.Errorf("archetypes: got %+v ok=%v, want archetypes hub", got, ok)
	}
	if _, ok, _ := BestLabelMatch("arena", pages); ok {
		t.Errorf("arena: got a match, want none (no page names it)")
	}
	if _, ok, _ := BestLabelMatch("Learn More", pages); ok {
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
	got, ok, _ := BestLabelMatch("Browse the Gripper Catalog", pages)
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
	got, ok, _ := BestLabelMatch("Widget", pages)
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

	got, ok, _ := BestLabelMatch("Gripper Safety Factor Calculator", pages)
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

	got, ok, _ := BestLabelMatch("Tools Guide", pages)
	if !ok || got.Name != "tools-guide" {
		t.Errorf("Tools Guide: got %+v ok=%v, want tools-guide (identity overlap 2 beats 1)", got, ok)
	}
}

// TestBestLabelMatchRefusesAnAlphabeticalTie — this test REPLACES
// TestBestLabelMatchTiesStillBreakByName, whose fixture is kept verbatim below
// and whose assertion is inverted (bugs_open/308 Phase B, 2026-08-23).
//
// WHAT THE OLD TEST PINNED, and why it was right at the time: the final
// tie-break is Name, and a candidate-token-set-size key must never replace it.
// That key was tried and DROPPED on 2026-08-11 after fleet calibration showed
// it was decided by tokenisation artefacts (a stray hyphen flipped 9 already-
// correct live gaswholesalers.com CTAs). That finding STANDS and is now
// reinforced: this lane measured two more substitute keys (name-tier,
// path-depth) on 2026-08-23 and dropped both for the same reason —
// CALIBRATION_2026-08-23_phase_b_widening_report.md §3.
//
// WHAT CHANGED: the conclusion those three rejections were pointing at. If no
// key can break the tie meaningfully, the tie itself is the finding, and the
// honest answer is "this label does not name one page". Measured on the fleet
// dump of 2026-08-23: 263 of 1,146 matches against the widened CTA universe
// were decided by nothing but alphabetical order, and 137 of those would have
// OVERWRITTEN a live CTA destination — including finetuning.uk's "how we work",
// which would have moved OFF the /how-we-work.html its copy names.
//
// So an alphabetical-only win now reports ok=false with ambiguous=true. The
// tie-break itself is NOT removed: it still decides which candidate is compared
// against the rest (and two rows for the SAME page are not a tie at all), so a
// size-based key re-introduced here would still change behaviour and still has
// to break an assertion — that is what the second half of this test pins.
func TestBestLabelMatchRefusesAnAlphabeticalTie(t *testing.T) {
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

	got, ok, ambiguous := BestLabelMatch("gadget", []LabelMatchCandidate{big, small})
	if ok || !ambiguous {
		t.Errorf("gadget: got %+v ok=%v ambiguous=%v — two DIFFERENT pages tie on identity, "+
			"total overlap and interactivity, so only alphabetical order separated them and the "+
			"label names neither", got, ok, ambiguous)
	}

	// The tie must be about the PAGE, not the row: two candidates naming one
	// destination are not a disagreement, and refusing there would silently
	// disable matching for every site whose page appears under both /x/ and
	// /x/index.html.
	sameA, _ := NewLabelMatchCandidate("3", "gadget-hub", "Gadget Hub", "/gadget/index.html", false, "")
	sameB, _ := NewLabelMatchCandidate("4", "gadget-hub-alias", "Gadget Hub", "/gadget/", false, "")
	got, ok, ambiguous = BestLabelMatch("gadget", []LabelMatchCandidate{sameA, sameB})
	if !ok || ambiguous || NormalizePagePath(got.URL) != "/gadget" {
		t.Errorf("two rows for ONE page must still resolve: got %+v ok=%v ambiguous=%v", got, ok, ambiguous)
	}

	// And a clear winner is still returned: the ambiguity rule must not fire
	// whenever two candidates merely both overlap.
	clear, _ := NewLabelMatchCandidate("5", "gadget-calibrator", "Gadget Calibrator", "/gadget-calibrator.html", false, "")
	other, _ := NewLabelMatchCandidate("6", "widget-press", "Widget Press", "/widget-press.html", false, "")
	got, ok, ambiguous = BestLabelMatch("gadget calibrator", []LabelMatchCandidate{clear, other})
	if !ok || ambiguous || got.Name != "gadget-calibrator" {
		t.Errorf("unambiguous winner refused: got %+v ok=%v ambiguous=%v", got, ok, ambiguous)
	}
}
