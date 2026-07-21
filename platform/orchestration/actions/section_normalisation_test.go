package actions

import (
	"reflect"
	"testing"
)

// These tests pin the fix for /bugs_open/041 (section-lookup-never-normalises):
// a snake_case/CamelCase section name must resolve to the kebab-case component
// the library actually stores, WITHOUT regressing the components whose own name
// is snake_case and whose function is a different kebab string.

func TestNormalizeComponentFunction_SectionNames(t *testing.T) {
	cases := map[string]string{
		"call_to_action": "call-to-action",
		"social_proof":   "social-proof",
		"SocialProof":    "social-proof",
		"call-to-action": "call-to-action", // already kebab — no-op
		"hero":           "hero",           // single word — no-op
		"":               "",               // empty — no-op
	}
	for in, want := range cases {
		if got := NormalizeComponentFunction(in); got != want {
			t.Errorf("NormalizeComponentFunction(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSectionLookupKeys(t *testing.T) {
	cases := []struct {
		name string
		want []string
	}{
		// snake_case: BOTH the raw and normalised forms are tried, so a plan
		// requesting "call_to_action" can match the stored "call-to-action".
		{"call_to_action", []string{"call_to_action", "call-to-action"}},
		// featured_article: the component's own NAME is snake_case (function is
		// "featured-content"), so it resolves ONLY by the raw name. The raw form
		// must stay in the key set or it would silently vanish.
		{"featured_article", []string{"featured_article", "featured-article"}},
		// already kebab: no second key, no wasted query arg.
		{"call-to-action", []string{"call-to-action"}},
		{"hero", []string{"hero"}},
	}
	for _, tc := range cases {
		if got := sectionLookupKeys(tc.name); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("sectionLookupKeys(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestSectionResolvedByFound(t *testing.T) {
	// The DB holds kebab "call-to-action" and a component literally named
	// "featured_article" (function "featured-content").
	found := map[string]bool{
		"call-to-action":   true, // name+function of the CTA component
		"featured_article": true, // the snake_case NAME
		"featured-content": true, // its kebab function
	}
	cases := []struct {
		name string
		want bool
	}{
		{"call_to_action", true},   // resolves via the normalised form
		{"call-to-action", true},   // resolves directly
		{"featured_article", true}, // resolves via its raw (snake) name — must NOT regress
		{"testimonials", false},    // genuinely absent — still unresolved
		{"category_section", false},
	}
	for _, tc := range cases {
		if got := sectionResolvedByFound(found, tc.name); got != tc.want {
			t.Errorf("sectionResolvedByFound(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestSectionLookupValueSet_Dedupes(t *testing.T) {
	// call_to_action expands to two forms; call-to-action (its normalised form
	// requested directly) must not add a duplicate.
	got := sectionLookupValueSet([]string{"call_to_action", "call-to-action", "hero"})
	want := []string{"call_to_action", "call-to-action", "hero"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("sectionLookupValueSet = %v, want %v", got, want)
	}
}

func TestAliasNormalisedSectionKeys(t *testing.T) {
	cta := componentInfo{ID: "1", Name: "call-to-action", Function: "call-to-action"}
	feat := componentInfo{ID: "2", Name: "featured_article", Function: "featured-content"}

	// loadSectionComponents resolves the CTA (stored kebab) and the featured
	// component (stored under its snake_case name). loadComponentSchemas keys the
	// map by the STORED identifiers.
	result := map[string]componentInfo{
		"call-to-action":   cta,
		"featured_article": feat,
		"featured-content": feat,
	}

	// The plan requested these RAW names.
	aliasNormalisedSectionKeys(result, []string{"call_to_action", "featured_article"})

	// The snake_case request now resolves to the kebab component...
	if got, ok := result["call_to_action"]; !ok || got.ID != "1" {
		t.Fatalf("expected result[call_to_action] aliased to the CTA component, got %+v (ok=%v)", got, ok)
	}
	// ...and the already-present snake_case NAME key is untouched (not rebound).
	if got := result["featured_article"]; got.ID != "2" {
		t.Fatalf("featured_article must keep its own component, got %+v", got)
	}
	// The kebab key remains as it was.
	if got := result["call-to-action"]; got.ID != "1" {
		t.Fatalf("call-to-action key must be preserved, got %+v", got)
	}
}

// A section name that resolves under NEITHER form must not be aliased into
// existence — it should still fall through to the needs_new_component path.
func TestAliasNormalisedSectionKeys_NoPhantom(t *testing.T) {
	result := map[string]componentInfo{
		"call-to-action": {ID: "1", Name: "call-to-action", Function: "call-to-action"},
	}
	aliasNormalisedSectionKeys(result, []string{"category_section"})
	if _, ok := result["category_section"]; ok {
		t.Fatal("category_section resolves to nothing and must not be aliased")
	}
}
