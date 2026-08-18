// FILE: platform/orchestration/actions/queryresolve/vocabulary_test.go
//
// IsKnownQueryName answers from the SAME map Resolve dispatches through, so
// there is no roster to drift — these tests pin the normalisation contract
// (case, whitespace, :arg) and the census's real offenders (bugs_open/309).

package queryresolve

import "testing"

func TestIsKnownQueryNameAcceptsTheRegisteredVocabulary(t *testing.T) {
	for _, name := range []string{
		"blog_posts",
		"pages_where_type:tool",
		"pages_under_section:guides",
		"section_index_for:game",
		"products:gripper",
		"latest_news",
		"news_archive",
		"business_directory",
		"model_directory_full",
		"  Blog_Posts  ", // Resolve normalises case + whitespace; so must this
	} {
		if !IsKnownQueryName(name) {
			t.Errorf("IsKnownQueryName(%q) = false, want true", name)
		}
	}
}

func TestIsKnownQueryNameRefusesTheCensusOffenders(t *testing.T) {
	// The 7 names declared in active component schemas that the resolver
	// has never registered (census 2026-08-18, bugs_open/309).
	for _, name := range []string{
		"affiliate_products",
		"category",
		"category_posts",
		"comparison_filter_types",
		"comparison_results",
		"featured_post",
		"pages",
		"",
	} {
		if IsKnownQueryName(name) {
			t.Errorf("IsKnownQueryName(%q) = true, want false", name)
		}
	}
}

func TestKnownQueryBasesIsSortedAndComplete(t *testing.T) {
	bases := KnownQueryBases()
	if len(bases) != len(queryHandlers) {
		t.Fatalf("KnownQueryBases returned %d entries, map has %d", len(bases), len(queryHandlers))
	}
	seen := map[string]bool{}
	for i, b := range bases {
		if i > 0 && bases[i-1] >= b {
			t.Fatalf("bases not strictly sorted at %d: %q >= %q", i, bases[i-1], b)
		}
		if queryHandlers[b] == nil {
			t.Fatalf("base %q has nil handler", b)
		}
		seen[b] = true
	}
	for b := range queryHandlers {
		if !seen[b] {
			t.Fatalf("map base %q missing from KnownQueryBases", b)
		}
	}
}
