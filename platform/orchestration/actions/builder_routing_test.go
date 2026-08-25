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
// cannot silently change an existing route.
//
// 2026-08-25: the swap landed — WriteBuildItemsAction's inline copy is gone and
// section-index is now IN the map, so this table pins the route rather than the
// hold-out. Both producers read one function, so there is no longer a second
// copy for this table to be compared against by hand.

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
		// section-index routes to the layout-ensuring builder as of 2026-08-25
		// (the WriteBuildItemsAction swap; see builder_routing.go). Before that
		// it was held at the generic default so the two producers could not
		// disagree. If this row ever goes back to page-build-handler, a
		// section-index page with no layout in any source silently no-ops
		// again — that is bugs_open/206's entire signature.
		{"section-index", "directory-build-handler", "needs_directory", "", true},

		// Pre-existing routes.
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

// Every handler this function can return as a ROUTE must be an agent that
// actually exists. A route naming an unregistered agent mints items that
// dispatch to nothing — bugs_closed/078's shape (an unroutable handler_agent
// livelocking the dispatcher), and the exact hazard two council seats raised
// against this change on 2026-08-24.
//
// The allow-list is deliberately a LITERAL rather than a DB lookup: this is a
// unit test with no cluster, and the point is that adding a route to a new
// handler should FAIL here until someone has confirmed that handler is seeded
// and active. Both entries were verified live on 2026-08-24 —
// `SELECT type, is_active, is_snapshot, deleted_at FROM agent_definitions`
// returns `directory-build-handler|t|f|<null>` and `page-build-handler|t|f|<null>`.
// If you add a route, run that query for your handler and add it here in the
// same commit; if you cannot, your page_type belongs in unavailableBuilders
// instead, which is what that map is for.
func TestEveryRoutedHandlerIsAKnownRegisteredAgent(t *testing.T) {
	registered := map[string]bool{
		"page-build-handler":      true,
		"directory-build-handler": true,
	}
	// Every key the function routes, plus the unknown-type default.
	for _, pageType := range []string{
		"content", "index", "landing", "blog-index", "blog-post",
		"entity-directory", "section-index",
		"entity-page", "tool", // gap arms — must yield NO handler at all
		"never-heard-of-it",
	} {
		info, needed, _ := builderForPageType(pageType)
		if needed != "" {
			// A capability gap must not also name a dispatch handler: that is
			// the combination that would put an unregistered name on a row.
			if info.handler != "" {
				t.Errorf("builderForPageType(%q) reports a capability gap (%q) AND a handler (%q) — "+
					"a gap must yield no handler, or an unregistered name reaches a work item",
					pageType, needed, info.handler)
			}
			continue
		}
		if !registered[info.handler] {
			t.Errorf("builderForPageType(%q) routes to %q, which is not in this test's list of "+
				"handlers verified to exist as active agent_definitions rows. Confirm it is seeded "+
				"and active, then add it here in the same commit — or put the page_type in "+
				"unavailableBuilders so it files a capability_gap instead of an undispatchable item",
				pageType, info.handler)
		}
	}
}

// The routing is only reachable through normalizePageType at every producer,
// so pin the composition: spelling variants of entity-directory must reach the
// directory-build-handler route, not the unknown-type default.
func TestBuilderForPageTypeNormalizeComposition(t *testing.T) {
	for _, raw := range []string{"Entity Directory", "entity_directory", "ENTITY-DIRECTORY", "Entity_Directory"} {
		info, needed, known := builderForPageType(normalizePageType(raw))
		if !known || needed != "" || info.handler != "directory-build-handler" {
			t.Errorf("builderForPageType(normalizePageType(%q)) = ({%q %q}, %q, %v), want the entity-directory route",
				raw, info.handler, info.itemType, needed, known)
		}
	}
}
