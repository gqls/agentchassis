// FILE: platform/orchestration/actions/builder_routing_test.go
//
// Pins builderForPageType — the single routing authority for "which handler
// builds a page of this page_type" (bugs_open/206 closure hardening). Until
// 2026-08-24 this decision had ZERO test coverage anywhere: it lived as
// closure-local maps inside WriteBuildItemsAction plus a hardcoded literal in
// reconcile_site_plan, and the two disagreed (the literal minted five doomed
// items the map would have routed or gap-filed).
//
// The table covers every map entry AND both non-entries, so the extraction
// cannot silently change an existing route, and removing the section-index
// entry (the 2026-08-24 addition) fails the test.

package actions

import "testing"

func TestBuilderForPageType(t *testing.T) {
	cases := []struct {
		pageType     string
		wantHandler  string
		wantItemType string
		wantNeeded   string
		wantKnown    bool
	}{
		// The 2026-08-24 addition: proven live twice by hand re-routes
		// (guides-index 08-08, practice 08-24) before the map learned it.
		{"section-index", "directory-build-handler", "needs_content_page", "", true},

		// Pre-existing routes — must be byte-identical to the inline map.
		{"entity-directory", "directory-build-handler", "needs_directory", "", true},
		{"content", "page-build-handler", "needs_content_page", "", true},
		{"index", "page-build-handler", "needs_content_page", "", true},
		{"landing", "page-build-handler", "needs_content_page", "", true},
		{"blog-index", "page-build-handler", "needs_content_page", "", true},
		{"blog-post", "page-build-handler", "needs_content_page", "", true},

		// Known types with NO builder: the caller must file a deferred
		// capability_gap naming the builder, and must NOT mint a dispatch
		// item (dispatching one burns an attempt and parks the row —
		// dartsonline brand-detail sat that way for 35 days).
		{"entity-page", "", "", "entity-page-builder", true},
		{"tool", "", "", "tool-builder", true},

		// Unknown type: generic default, known=false so the caller can log.
		{"never-heard-of-it", "page-build-handler", "needs_content_page", "", false},
		{"", "page-build-handler", "needs_content_page", "", false},
	}

	for _, c := range cases {
		info, needed, known := builderForPageType(c.pageType)
		if info.handler != c.wantHandler || info.itemType != c.wantItemType ||
			needed != c.wantNeeded || known != c.wantKnown {
			t.Errorf("builderForPageType(%q) = ({%q %q}, %q, %v), want ({%q %q}, %q, %v)",
				c.pageType, info.handler, info.itemType, needed, known,
				c.wantHandler, c.wantItemType, c.wantNeeded, c.wantKnown)
		}
	}
}

// The routing is only reachable through normalizePageType at every producer,
// so pin the composition: spelling variants of section-index must reach the
// directory-build-handler route, not the unknown-type default.
func TestBuilderForPageTypeNormalizeComposition(t *testing.T) {
	for _, raw := range []string{"Section Index", "section_index", "SECTION-INDEX", "Section_Index"} {
		info, needed, known := builderForPageType(normalizePageType(raw))
		if !known || needed != "" || info.handler != "directory-build-handler" {
			t.Errorf("builderForPageType(normalizePageType(%q)) = ({%q %q}, %q, %v), want the section-index route",
				raw, info.handler, info.itemType, needed, known)
		}
	}
}
