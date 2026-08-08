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

func TestNewLabelMatchCandidateRejectsAllGenericSources(t *testing.T) {
	if _, ok := NewLabelMatchCandidate("x", "x", "", "/x.html", false, "Learn More", ""); ok {
		t.Errorf("expected ok=false when every token source is generic/empty")
	}
}
