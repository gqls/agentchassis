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
	mk := func(name, title string, interactive bool, tokenSources ...string) LabelMatchCandidate {
		c, ok := NewLabelMatchCandidate(name, name, title, "/"+name+".html", interactive, tokenSources...)
		if !ok {
			t.Fatalf("NewLabelMatchCandidate(%q) unexpectedly produced no tokens", name)
		}
		return c
	}
	pages := []LabelMatchCandidate{
		mk("tool-gauntlet", "The Gauntlet", true, "tool", "gauntlet"),
		mk("tool-quiz", "Archetype Quiz", true, "tool", "quiz", "archetype"),
		mk("archetypes", "The Archetypes", false, "archetypes"),
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
	mk := func(name, title string, interactive bool, tokenSources ...string) LabelMatchCandidate {
		c, ok := NewLabelMatchCandidate(name, name, title, "/"+name+".html", interactive, tokenSources...)
		if !ok {
			t.Fatalf("NewLabelMatchCandidate(%q) unexpectedly produced no tokens", name)
		}
		return c
	}
	pages := []LabelMatchCandidate{
		// 1-token overlap with "Browse the Gripper Catalog": "gripper" only.
		mk("tool-gripper-cycle-time-estimator", "Gripper Cycle Time Estimator", true,
			"tool", "gripper", "cycle", "time", "estimator"),
		// 2-token overlap: "gripper" and "catalog" — the better match, non-interactive.
		mk("gripper-catalog-index", "Gripper Catalog", false, "gripper", "catalog", "index"),
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
	mk := func(name, title string, interactive bool, tokenSources ...string) LabelMatchCandidate {
		c, ok := NewLabelMatchCandidate(name, name, title, "/"+name+".html", interactive, tokenSources...)
		if !ok {
			t.Fatalf("NewLabelMatchCandidate(%q) unexpectedly produced no tokens", name)
		}
		return c
	}
	pages := []LabelMatchCandidate{
		mk("tool-widget", "Widget Tool", true, "widget"),
		mk("widget-guide", "Widget Guide", false, "widget"),
	}
	got, ok := BestLabelMatch("Widget", pages)
	if !ok || got.Name != "tool-widget" {
		t.Errorf("Widget (equal overlap both candidates): got %+v ok=%v, want the interactive tool-widget", got, ok)
	}
}

func TestNewLabelMatchCandidateRejectsAllGenericSources(t *testing.T) {
	if _, ok := NewLabelMatchCandidate("x", "x", "", "/x.html", false, "Learn More", ""); ok {
		t.Errorf("expected ok=false when every token source is generic/empty")
	}
}
