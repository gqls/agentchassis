package actions

import (
	"reflect"
	"testing"
)

// fleetSeedRows mirrors 689_theme_kits.sql's fleet-scope INSERT exactly —
// if that migration's seed data changes, this fixture must change with it.
// Kept as a literal (not read from the .sql file) so this test fails loudly
// on drift rather than silently agreeing with whatever the migration says.
func fleetSeedRows() []archetypeRow {
	return []archetypeRow{
		{MatchKind: "page_type", MatchValue: "news-index", Sections: []string{"hero", "news-listing", "call-to-action"}},
		{MatchKind: "page_type", MatchValue: "entity-directory", Sections: []string{"hero", "directory-listing"}},
		{MatchKind: "page_type", MatchValue: "section-index", Sections: []string{"hero", "content-listing"}},
		{MatchKind: "page_name", MatchValue: "faq", Sections: []string{"hero", "faq", "call-to-action"}},
		{MatchKind: "page_name_contains", MatchValue: "faq", Sections: []string{"hero", "faq", "call-to-action"}},
		{MatchKind: "page_name", MatchValue: "contact", Sections: []string{"contact-hero", "contact-form", "contact-info"}},
		{MatchKind: "page_name", MatchValue: "pricing", Sections: []string{"hero", "pricing", "faq", "call-to-action"}},
		{MatchKind: "page_name_contains", MatchValue: "pricing", Sections: []string{"hero", "pricing", "faq", "call-to-action"}},
		{MatchKind: "page_name", MatchValue: "about", Sections: []string{"hero-about", "about-content", "call-to-action"}},
		{MatchKind: "page_name_suffix", MatchValue: "guides-index", Sections: []string{"hero", "guide-list"}},
		{MatchKind: "page_name", MatchValue: "guide-index", Sections: []string{"hero", "guide-list"}},
		{MatchKind: "page_name_suffix", MatchValue: "tools-index", Sections: []string{"hero", "tool-list"}},
		{MatchKind: "page_name", MatchValue: "tool-index", Sections: []string{"hero", "tool-list"}},
		{MatchKind: "default", MatchValue: "", Sections: []string{"hero", "generic-text-block", "call-to-action"}},
	}
}

// TestMatchArchetypeRows_ParityWithDefaultSectionsForPage asserts the ported
// fleet rows reproduce defaultSectionsForPage's own output EXACTLY for every
// (name, type) case the switch documents — including the section-index case
// added here, which defaultSectionsForPage never had (see 689_theme_kits.sql
// header note; credit: calendar session, 2026-09-02).
func TestMatchArchetypeRows_ParityWithDefaultSectionsForPage(t *testing.T) {
	rows := fleetSeedRows()

	cases := []struct {
		name         string
		pageName     string
		pageType     string
		wantFromGo   bool // false only for the new section-index case
		wantSections []string
	}{
		{"news-index by type", "noticias", "news-index", true, nil},
		{"entity-directory by type", "practices", "entity-directory", true, nil},
		{"faq by equality", "faq", "", true, nil},
		{"faq by contains", "some-faq-page", "", true, nil},
		{"contact", "contact", "", true, nil},
		{"pricing by equality", "pricing", "", true, nil},
		{"pricing by contains", "our-pricing-plans", "", true, nil},
		{"about", "about", "", true, nil},
		{"guides-index by suffix", "buying-guides-index", "", true, nil},
		{"guide-index by equality", "guide-index", "", true, nil},
		{"tools-index by suffix", "founder-tools-index", "", true, nil},
		{"tool-index by equality", "tool-index", "", true, nil},
		{"unmatched falls to default", "something-else-entirely", "", true, nil},
		{"section-index — NEW, not in the Go switch", "july", "section-index", false, []string{"hero", "content-listing"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, matched := matchArchetypeRows(rows, tc.pageName, tc.pageType)
			if !matched {
				t.Fatalf("matchArchetypeRows(%q, %q): no match, want one", tc.pageName, tc.pageType)
			}
			want := tc.wantSections
			if tc.wantFromGo {
				want = defaultSectionsForPage(tc.pageName, tc.pageType)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("matchArchetypeRows(%q, %q) = %v, want %v", tc.pageName, tc.pageType, got, want)
			}
		})
	}
}

// TestMatchArchetypeRows_TypeOutranksName pins bugs_open/015's rule (a
// page_type match must win over any page_name match, since type is not
// localised and name can be) at the resolver level, not just by inheriting
// the Go switch's structure.
func TestMatchArchetypeRows_TypeOutranksName(t *testing.T) {
	rows := []archetypeRow{
		{MatchKind: "page_name", MatchValue: "noticias", Sections: []string{"WRONG — name-matched"}},
		{MatchKind: "page_type", MatchValue: "news-index", Sections: []string{"hero", "news-listing", "call-to-action"}},
	}
	got, matched := matchArchetypeRows(rows, "noticias", "news-index")
	if !matched {
		t.Fatal("expected a match")
	}
	want := []string{"hero", "news-listing", "call-to-action"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("type should outrank name: got %v, want %v", got, want)
	}
}

// TestMatchArchetypeRows_NoMatchReturnsFalse — a scope with no default row
// and nothing matching must report matched=false so the caller can fall
// through to the next scope, not silently return a nil/empty section list.
func TestMatchArchetypeRows_NoMatchReturnsFalse(t *testing.T) {
	rows := []archetypeRow{
		{MatchKind: "page_type", MatchValue: "news-index", Sections: []string{"hero"}},
	}
	_, matched := matchArchetypeRows(rows, "anything", "unrelated-type")
	if matched {
		t.Error("expected no match for an unrelated page with no default row in scope")
	}
}
