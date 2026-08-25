// FILE: platform/orchestration/actions/builder_routing.go
//
// The ONE authority for "which handler builds a page of this page_type".
//
// bugs_open/206 was fixed at WriteBuildItemsAction on 2026-08-08 and RE-FIRED
// through reconcile_site_plan between 08-18 and 08-23 — five parked items on
// three sites, every one with the 206 signature ("page-build-handler no-op:
// no sections ready to build"), one of them an entity-directory page the
// WriteBuildItemsAction map would have routed correctly. The cause was not a
// missing entry but a SECOND COPY of the routing decision: reconcile's emit
// hardcoded 'page-build-handler' for every page. Two hand-maintained copies of
// one decision is the drift class this estate keeps paying for. Do not add a
// third copy; call this function.
//
// BOTH DOORS NOW CALL THIS (2026-08-25). The transient duplicate is gone:
// WriteBuildItemsAction's inline maps were removed the day its file came clean,
// and `section-index` was added here in the same commit — which is the order
// council round 4 required, because the two producers mint into the SAME
// item_key namespace (`needs_page:<name>`, both of them) under idx_swi_dedup,
// so a disagreement between them is not untidiness: whichever fires first
// silently wins and the loser's routing never happens.
//
// grep before adding a third caller: `builderForPageType` should be the only
// place a page_type is turned into a handler name.

package actions

// builderRouteInfo describes where a page build routes.
type builderRouteInfo struct {
	handler  string
	itemType string
}

// builderForPageType decides which handler builds a page of the given
// (already-normalized — see normalizePageType) page_type. Three outcomes:
//
//   - a route (known=true, neededBuilder==""): mint a dispatch item at
//     info.handler with info.itemType;
//   - a recorded capability gap (neededBuilder!=""): the caller files a
//     deferred capability_gap item naming neededBuilder and does NOT mint a
//     dispatch item — dispatching one burns an attempt on a handler that
//     cannot build it and parks the row in needs_human_review (the fate of
//     dartsonline's brand-detail for 35 days, bugs_open/206 closure census);
//   - the generic default for an unknown type (known=false, caller may log).
func builderForPageType(pageType string) (info builderRouteInfo, neededBuilder string, known bool) {
	availableBuilders := map[string]builderRouteInfo{
		"content":    {handler: "page-build-handler", itemType: "needs_content_page"},
		"index":      {handler: "page-build-handler", itemType: "needs_content_page"},
		"landing":    {handler: "page-build-handler", itemType: "needs_content_page"},
		"blog-index": {handler: "page-build-handler", itemType: "needs_content_page"},
		"blog-post":  {handler: "page-build-handler", itemType: "needs_content_page"},
		// directory-build-handler ensures the page's plan layout
		// (ensure_page_section_layout → defaultSectionsForPage) then
		// delegates the actual build to page-build-handler — see
		// agent_definitions seed 001_directory_build_handler.sql and
		// migrations 336/337 (the seed alone does NOT match the live row).
		"entity-directory": {handler: "directory-build-handler", itemType: "needs_directory"},

		// section-index joins entity-directory on directory-build-handler for
		// the SAME reason, and it is the reason that matters: that handler is
		// the only one whose workflow carries `ensure_page_section_layout`, so
		// it is the only one that can build a page whose layout is missing
		// from every source. page-build-handler no-ops on exactly that case
		// ("no sections ready to build") — which is the whole of bugs_open/206.
		//
		// The route is proven at the artefact, twice, by hand re-routes that
		// predate any map knowing it: vetcomparison guides-index
		// (page_type='section-index', sections ["hero","guide-list"], serving
		// since 2026-08-08) and vetcomparison practice (2026-08-24).
		//
		// It was held out of this map from 2026-08-24 to 2026-08-25 ONLY
		// because WriteBuildItemsAction still had its own copy of these
		// entries and the two producers would have disagreed on this one
		// page_type. That copy is gone as of the commit that adds this line;
		// the hold-out has served its purpose and the reason it existed no
		// longer applies.
		"section-index": {handler: "directory-build-handler", itemType: "needs_directory"},
	}
	// Known page types whose builders don't exist yet. entity-page is a
	// DECISION, not an oversight: the two parked entity-page builds as of
	// 2026-08-24 (dartsonline brand-detail, garden-tools brand-profile) are
	// data-backed profile classes where auto-filled sections are fabrication
	// risk; the capability_gap row this branch causes is the visibility
	// mechanism, and the 2026-08-24 practice build (vetcomparison) proves a
	// human-decided page reaches production via re-route + content_direction
	// rails without a dedicated builder. A dedicated builder is its own
	// council round, when a lane wants entity profiles AND has entity data.
	unavailableBuilders := map[string]string{
		"tool":        "tool-builder",
		"entity-page": "entity-page-builder",
	}
	if info, ok := availableBuilders[pageType]; ok {
		return info, "", true
	}
	if nb, ok := unavailableBuilders[pageType]; ok {
		return builderRouteInfo{}, nb, true
	}
	return builderRouteInfo{handler: "page-build-handler", itemType: "needs_content_page"}, "", false
}
