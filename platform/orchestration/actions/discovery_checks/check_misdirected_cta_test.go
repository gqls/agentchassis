package discovery_checks

import (
	"reflect"
	"testing"
)

func TestCTATokens(t *testing.T) {
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
		got := ctaTokens(text)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("ctaTokens(%q) = %v, want %v", text, got, want)
		}
	}
}

func TestBestPageMatch(t *testing.T) {
	pages := []ctaMatchPage{
		{Name: "tool-gauntlet", Title: "The Gauntlet", URL: "/tools/gauntlet/index.html",
			Interactive: true, tokens: map[string]bool{"tool": true, "gauntlet": true}},
		{Name: "tool-quiz", Title: "Archetype Quiz", URL: "/tools/quiz/index.html",
			Interactive: true, tokens: map[string]bool{"tool": true, "quiz": true, "archetype": true}},
		{Name: "archetypes", Title: "The Archetypes", URL: "/archetypes.html",
			tokens: map[string]bool{"archetypes": true}},
	}

	if got := bestPageMatch([]string{"gauntlet"}, pages); got == nil || got.Name != "tool-gauntlet" {
		t.Errorf("gauntlet: got %+v, want tool-gauntlet", got)
	}
	// Interactive pages beat content pages on equal-strength matches:
	// "archetype" hits the quiz (interactive), "archetypes" hits the hub.
	if got := bestPageMatch([]string{"archetype"}, pages); got == nil || got.Name != "tool-quiz" {
		t.Errorf("archetype: got %+v, want tool-quiz", got)
	}
	if got := bestPageMatch([]string{"archetypes"}, pages); got == nil || got.Name != "archetypes" {
		t.Errorf("archetypes: got %+v, want archetypes hub", got)
	}
	if got := bestPageMatch([]string{"arena"}, pages); got != nil {
		t.Errorf("arena: got %+v, want nil (no page names it)", got)
	}
}

func TestCTAAreaExcluded(t *testing.T) {
	cases := map[string]bool{
		"/contact.html":       true, // top-level page — the firstPathSegment blind spot
		"/contact/index.html": true,
		"/legal/privacy.html": true,
		"/about.html":         true,
		"/tools/gauntlet/index.html": false,
		"/archetypes.html":           false,
		"/":                          false,
	}
	for url, want := range cases {
		if got := ctaAreaExcluded(url); got != want {
			t.Errorf("ctaAreaExcluded(%q) = %v, want %v", url, got, want)
		}
	}
}
