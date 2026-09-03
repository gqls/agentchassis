// FILE: platform/orchestration/actions/listing_item_sources.go
//
// bugs_open/444 — the listing-page item-source registry: ONE answer to "is this
// planned page a listing page, and does its item source resolve to anything for
// THIS site?", shared by plan-time validation (ValidateSitePlanAction) so a
// listing page whose producer does not exist becomes a capability_gap work item
// instead of a built page of meta-prose.
//
// Why this is judged at PLAN time and not at render time: the render layer
// either cannot close the class (news-listing.items is required:false BY DESIGN
// — the client JSON refresh is the freshness path between rerenders, so an
// empty list is legal on a site WITH sources) or must not close it alone
// (directory-listing's schema contract did not fire on the 2026-09-02 remakes
// because resolveBusinessDirectory deliberately ERRORS on missing config,
// bugs_open/206, and plan_sections' error branch deliberately bypasses
// on_missing, bugs_open/054 — two correct guards composing into the silent
// hollow section both exist to prevent).
//
// POLICY: fail OPEN on ambiguity. A resolver only reports "not producible" for
// a shape it positively understands; anything it cannot classify is left alone
// with a log line. A false drop silently shrinks a site; a false keep ships one
// more empty page of a class we now detect downstream — the asymmetry decides.
//
// ⚠ SIBLING GATE, AND THE ORDER BETWEEN THEM IS LOAD-BEARING: tool_item_sources.go
// (bugs_open/450, BLD-029, key enforce_tool_sources) runs BEFORE this one in
// ValidateSitePlanAction. Held tool children make a /tools/ hub resolve zero
// children here, so this gate then holds the hub too — which is intended, and
// means that with both keys armed THIS gate's behaviour changes. Reversing the
// order ships an empty hub, a 444-class page. Pinned by
// TestToolGateRunsBeforeListingGate + TestListingGateFirstWouldKeepTheEmptyHub.
// This pointer exists so a future editor changing gate order finds both halves
// (council corr 4e7497ed, architecture seat). Comment only — no behaviour of
// this file is changed by the 450 lane.
package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	discovery_checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/agenterrors"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// ListingSourceResolution reports whether a planned page is listing-family and
// whether its item source resolves for the site being planned.
type ListingSourceResolution struct {
	ListingFamily  bool   // page promises a list of items by role or by section component
	Producible     bool   // meaningful only when ListingFamily: the item source resolves
	ProducerNeeded string // when !Producible: the slug capability_gap consumers group on (spec.builder_needed)
	Evidence       string // one line saying what was checked and what it found
}

// isListingRole is datahelpers' listing-family vocabulary — ONE definition,
// not a mirror (the per-kind directory page types are covered via
// directoryListingComponents below: their pages always carry their kind's
// listing component, per directoryCheckProfiles).
func isListingRole(role string) bool { return datahelpers.IsSectionIndexRole(role) }

// directoryListingComponents maps a per-kind directory listing/snippet
// component name to the content_features spec key that opts a site into that
// kind (DIR-001). DERIVED from discovery_checks' directoryCheckProfiles at
// package init — a kind added there is automatically known here; no second
// hand-maintained roster (council corr c0990eb3 round 2).
var directoryListingComponents = discovery_checks.ListingComponentSpecKeys()

// planPageView is the subset of a plan page the resolver needs, extracted once
// by the caller so the resolver never touches the raw map shape.
type planPageView struct {
	Name     string
	Role     string // canonicalised page_type/role, kebab form
	URL      string // may be empty at plan time
	Sections []string
}

// ResolveListingItemSource classifies one planned page and, when it is
// listing-family, resolves its item source for siteID. allPages and
// realisedPages supply the plan-internal and already-built context for
// child-page counting (section-index / blog-index), so post-plan builders
// (tool-deployer) cannot false-positive: their pages are IN the plan even
// though they are built later.
func ResolveListingItemSource(ctx context.Context, db *sql.DB, siteID uuid.UUID, page planPageView, allPages []planPageView, logger *zap.Logger) ListingSourceResolution {
	// Section-component half of the family test: a listing component on a page
	// of ANY role makes the page listing-family (advertise's glossary-shaped
	// trap is the inverse — a listing PAGE with no listing component — caught
	// by the role half below).
	var listingComponents []string
	for _, s := range page.Sections {
		if isListingComponent(s) {
			listingComponents = append(listingComponents, s)
		}
	}

	family := isListingRole(page.Role) || len(listingComponents) > 0
	if !family {
		return ListingSourceResolution{}
	}
	res := ListingSourceResolution{ListingFamily: true}

	// Any resolvable section component satisfies the page. Collect the reasons
	// so an unproducible page names everything it would need. A component of a
	// shape this registry does not understand contributes NOTHING either way
	// (unknownShapes) — only positively-understood "not producible" verdicts
	// may hold a page.
	var unmet []string
	var unmetSlugs []string
	unknownShapes := 0
	triedNews := false
	for _, comp := range listingComponents {
		ok, needed, evidence := resolveListingComponent(ctx, db, siteID, comp, allPages, logger)
		if comp == "news-listing" {
			triedNews = true
		}
		if ok {
			res.Producible = true
			res.Evidence = evidence
			return res
		}
		if needed == "" {
			unknownShapes++
			continue
		}
		unmet = append(unmet, evidence)
		unmetSlugs = append(unmetSlugs, needed)
	}

	// Role-level resolution for pages whose sections carry no (resolvable)
	// listing component.
	switch page.Role {
	case "news-index":
		if !triedNews {
			ok, evidence := newsSourceResolves(ctx, db, siteID)
			if ok {
				res.Producible = true
				res.Evidence = evidence
				return res
			}
			unmet = append(unmet, evidence)
			unmetSlugs = append(unmetSlugs, "news_source_enablement")
		}
	case "blog-index":
		if n := countPagesOfRole(allPages, "blog-post"); n > 0 {
			res.Producible = true
			res.Evidence = fmt.Sprintf("%d blog-post pages in plan/realised set", n)
			return res
		}
		unmet = append(unmet, "no blog-post pages in the plan or realised set")
		unmetSlugs = append(unmetSlugs, "blog_posts")
	case "section-index":
		if n := countSectionChildren(page, allPages); n > 0 {
			res.Producible = true
			res.Evidence = fmt.Sprintf("%d child pages under %s", n, sectionPrefixOf(page))
			return res
		}
		unmet = append(unmet, fmt.Sprintf("no child pages under %s in the plan or realised set", sectionPrefixOf(page)))
		unmetSlugs = append(unmetSlugs, "section_children:"+page.Name)
	case "entity-directory":
		if len(listingComponents) == 0 {
			// An entity-directory page planned with an empty sections array
			// (the prompt's rule 3 shape) receives hero+directory-listing from
			// defaultSectionsForPage at build time — so resolve the same
			// source that injected component will use before holding the page.
			ok, needed, evidence := resolveListingComponent(ctx, db, siteID, "directory-listing", allPages, logger)
			if ok {
				res.Producible = true
				res.Evidence = evidence + " (empty-sections entity-directory; default layout injects directory-listing)"
				return res
			}
			if needed != "" {
				unmet = append(unmet, evidence)
				unmetSlugs = append(unmetSlugs, needed)
			} else {
				unmet = append(unmet, "entity-directory page plans no directory listing component and no source resolves")
				unmetSlugs = append(unmetSlugs, "directory_kind:"+page.Name)
			}
		}
	}

	if len(unmetSlugs) == 0 {
		// Listing-family by component only, and every component was of a shape
		// this registry does not understand (e.g. category-listing) — fail
		// OPEN, loudly.
		logger.Info("listing_item_sources: listing components of unknown source shape — keeping page (fail-open)",
			zap.String("page", page.Name),
			zap.Strings("components", listingComponents))
		res.Producible = true
		res.Evidence = "components of unknown source shape; fail-open"
		return res
	}

	res.Producible = false
	res.ProducerNeeded = unmetSlugs[0] // the first unmet need is the headline; evidence carries the rest
	res.Evidence = strings.Join(unmet, "; ")
	return res
}

// enforceListingItemSources is ValidateSitePlanAction's gate (bugs_open/444,
// opt-in via step config `enforce_listing_sources`): every listing-family plan
// page whose item source resolves to nothing for this site is REMOVED from the
// plan and filed as a capability_gap work item naming the missing producer —
// never built as meta-prose. Returns the filtered pages slice.
//
// Two deliberate boundaries:
//   - A page that is already REALISED on the site is kept even when
//     unproducible (the bugs_open/001 preserve guard owns built pages; the
//     gap item is still filed, so the enablement debt is on record) — this
//     gate stops NEW empty listing pages, it does not tear down old ones.
//   - Every ambiguity fails OPEN (page kept, logged): a false drop silently
//     shrinks a site, a false keep ships one more page of a class the
//     downstream checks now detect.
func enforceListingItemSources(ctx context.Context, params ActionParams, pages []interface{}, existingPages []interface{}) []interface{} {
	if params.DB == nil {
		params.Logger.Warn("listing_item_sources: enforce_listing_sources set but no DB — gate skipped (fail-open)")
		return pages
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		params.Logger.Warn("listing_item_sources: no parseable site id — gate skipped (fail-open)",
			zap.String("site_id", siteIDStr))
		return pages
	}

	planViews := make([]planPageView, 0, len(pages))
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			planViews = append(planViews, pageViewFromMap(pm))
		}
	}
	realisedViews := make([]planPageView, 0, len(existingPages))
	realisedKeys := make(map[string]bool, len(existingPages))
	for _, p := range existingPages {
		if pm, ok := p.(map[string]interface{}); ok {
			v := pageViewFromMap(pm)
			realisedViews = append(realisedViews, v)
			if v.Name != "" {
				realisedKeys[v.Name] = true
			}
			if v.URL != "" {
				realisedKeys[v.URL] = true
			}
		}
	}
	allViews := append(append([]planPageView{}, planViews...), realisedViews...)

	kept := make([]interface{}, 0, len(pages))
	var drops []droppedSectionName // reuse the drop-record shape: Page = page name, Name = producer needed
	viewIdx := 0
	for _, p := range pages {
		if _, ok := p.(map[string]interface{}); !ok {
			kept = append(kept, p)
			continue
		}
		view := planViews[viewIdx]
		viewIdx++
		res := ResolveListingItemSource(ctx, params.DB, siteID, view, allViews, params.Logger)
		if !res.ListingFamily || res.Producible {
			kept = append(kept, p)
			continue
		}
		if realisedKeys[view.Name] || (view.URL != "" && realisedKeys[view.URL]) {
			params.Logger.Warn("listing_item_sources: realised listing page has no item source — kept (preserve guard); capability_gap filed",
				zap.String("page", view.Name),
				zap.String("producer_needed", res.ProducerNeeded),
				zap.String("evidence", res.Evidence))
			fileListingCapabilityGap(ctx, params, siteID, view, res)
			kept = append(kept, p)
			continue
		}
		params.Logger.Warn("listing_item_sources: dropped listing page with no resolvable item source",
			zap.String("page", view.Name),
			zap.String("role", view.Role),
			zap.String("producer_needed", res.ProducerNeeded),
			zap.String("evidence", res.Evidence))
		fileListingCapabilityGap(ctx, params, siteID, view, res)
		drops = append(drops, droppedSectionName{Page: view.Name, Name: res.ProducerNeeded})
	}
	if len(drops) > 0 {
		// Durable record through the same findings door the section-name drops
		// use, so "the gate fired" and "the planner proposed nothing
		// unfillable" never produce identical evidence.
		attempted, recorded := LogActionFindings(ctx, params, siteIDStr, "", "validate_plan",
			listingDropFindings(drops), params.Logger)
		warnUnrecordedDrops(attempted, recorded, params.Logger)
	}
	return kept
}

// fileListingCapabilityGap files ONE deferred capability_gap row for an
// unproducible listing page, through the SHARED work-item writer
// (insertWorkItem — the same door WriteBuildItemsAction's capability_gap arm
// uses, council corr c0990eb3 round 2, reuse_agent: a fourth hand-rolled
// INSERT of this shape was the objection, and the shared struct's own header
// records what walking round the shared door has cost before). Shape matches
// the sibling producers exactly: item_key `capability_gap:<page_type>:<page_name>`
// co-dedups on the partial unique index; handler_agent EMPTY (the
// bugs_closed/078/291 rule); spec.builder_needed is the field diagnose_triage's
// roadmap grouping reads. Best-effort by design: a failed receipt must not
// change the disposition (the drop is already durable in the findings row) —
// but it must not be silent either.
func fileListingCapabilityGap(ctx context.Context, params ActionParams, siteID uuid.UUID, view planPageView, res ListingSourceResolution) {
	gapKey := fmt.Sprintf("capability_gap:%s:%s", view.Role, view.Name)
	gapSpec, _ := json.Marshal(map[string]interface{}{
		"gap_kind":       "producer_missing",
		"page_name":      view.Name,
		"page_type":      view.Role,
		"builder_needed": res.ProducerNeeded,
		"reason":         res.Evidence,
	})
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		params.Logger.Error("listing_item_sources: capability_gap tx open failed",
			zap.String("page", view.Name), zap.Error(err))
		return
	}
	inserted, err := insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "validate_site_plan",
		pipeline:     "build",
		itemType:     "capability_gap",
		severity:     "low",
		summary:      fmt.Sprintf("Listing page '%s' has no item source (%s) — held from the plan", view.Name, res.ProducerNeeded),
		spec:         string(gapSpec),
		priority:     200,
		handlerAgent: "",
		status:       "deferred",
		createdBy:    "validate_site_plan",
		itemKey:      gapKey,
	}, params.Logger)
	if err != nil {
		_ = tx.Rollback()
		params.Logger.Error("listing_item_sources: capability_gap insert failed",
			zap.String("page", view.Name), zap.Error(err))
		return
	}
	if err := tx.Commit(); err != nil {
		params.Logger.Error("listing_item_sources: capability_gap commit failed",
			zap.String("page", view.Name), zap.Error(err))
		return
	}
	if !inserted {
		params.Logger.Info("listing_item_sources: capability_gap already on file (dedup)",
			zap.String("item_key", gapKey))
	}
}

// listingDropFindings converts gate drops to durable findings rows.
func listingDropFindings(drops []droppedSectionName) []agenterrors.Finding {
	findings := make([]agenterrors.Finding, 0, len(drops))
	for _, d := range drops {
		findings = append(findings, agenterrors.Finding{
			ErrorCode: "LISTING_PAGE_HELD_NO_ITEM_SOURCE",
			Severity:  "warning",
			Message:   fmt.Sprintf("listing page %q held from plan: producer needed %q", d.Page, d.Name),
			Context:   map[string]interface{}{"page": d.Page, "producer_needed": d.Name},
		})
	}
	return findings
}

// pageViewFromMap tolerantly extracts the resolver's view from either a plan
// page map or a realised-page row (page_type|type|role, url|slug, sections as
// strings or object entries).
func pageViewFromMap(pm map[string]interface{}) planPageView {
	v := planPageView{}
	if s, ok := pm["name"].(string); ok {
		v.Name = s
	}
	for _, k := range []string{"page_type", "type", "role"} {
		if s, ok := pm[k].(string); ok && s != "" {
			v.Role = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), "_", "-")
			break
		}
	}
	if s, ok := pm["url"].(string); ok && s != "" {
		v.URL = s
	} else if s, ok := pm["slug"].(string); ok && s != "" {
		v.URL = "/" + strings.TrimPrefix(s, "/")
	}
	if secs, ok := pm["sections"].([]interface{}); ok {
		for _, s := range secs {
			switch x := s.(type) {
			case string:
				v.Sections = append(v.Sections, x)
			case map[string]interface{}:
				if n, ok := x["name"].(string); ok {
					v.Sections = append(v.Sections, n)
				}
			}
		}
	}
	return v
}

// isListingComponent reports whether a section/component name is one this
// registry treats as promising a list of items.
func isListingComponent(name string) bool {
	if _, ok := directoryListingComponents[name]; ok {
		return true
	}
	switch name {
	case "directory-listing", "news-listing", "content-listing", "blog-listing", "category-listing":
		return true
	}
	return false
}

// resolveListingComponent answers for ONE component on one site. ok=false with
// an empty `needed` slug means "shape not understood" (fail open); a non-empty
// slug is a positive "not producible" with the enablement named.
func resolveListingComponent(ctx context.Context, db *sql.DB, siteID uuid.UUID, comp string, allPages []planPageView, logger *zap.Logger) (ok bool, needed string, evidence string) {
	if specKey, kind := directoryListingComponents[comp]; kind {
		recommended, err := contentFeatureRecommended(ctx, db, siteID, specKey)
		if err != nil {
			logger.Warn("listing_item_sources: spec-key read failed — fail-open", zap.String("component", comp), zap.Error(err))
			return true, "", "spec-key read failed; fail-open"
		}
		if recommended {
			return true, "", fmt.Sprintf("content_features.%s.recommended=true", specKey)
		}
		return false, "directory_kind_opt_in:" + specKey, fmt.Sprintf("site has not opted into content_features.%s (DIR-001)", specKey)
	}

	switch comp {
	case "directory-listing":
		// The SAME predicate resolveBusinessDirectory applies at render time —
		// shared so the gate and the renderer cannot disagree (the renderer
		// ERRORS without this config; bugs_open/206/444).
		has, err := queryresolve.HasBusinessDirectoryConfig(ctx, db, siteID)
		if err != nil {
			logger.Warn("listing_item_sources: business-directory config lookup failed — fail-open", zap.Error(err))
			return true, "", "business-directory config lookup failed; fail-open"
		}
		if has {
			return true, "", "directory-json-exporter config present (query.business_directory resolvable)"
		}
		return false, "business_directory_config", "no directory-json-exporter config for this site (query.business_directory errors by design, bugs_open/206)"
	case "news-listing":
		okNews, evidence := newsSourceResolves(ctx, db, siteID)
		if okNews {
			return true, "", evidence
		}
		return false, "news_source_enablement", evidence
	case "content-listing", "blog-listing":
		// Both declare source query.blog_posts (pages of role blog-post).
		if n := countPagesOfRole(allPages, "blog-post"); n > 0 {
			return true, "", fmt.Sprintf("%d blog-post pages feed query.blog_posts", n)
		}
		return false, "blog_posts", "query.blog_posts resolves to zero (no blog-post pages planned or realised)"
	}
	// Unknown shape (e.g. category-listing) — not understood, fail open.
	return false, "", ""
}

// newsSourceResolves is the NEWS-001 driver test: the classification spec
// recommends a news feed (the 6-hourly trigger then seeds content_sources from
// the spec) OR active source rows already exist.
func newsSourceResolves(ctx context.Context, db *sql.DB, siteID uuid.UUID) (bool, string) {
	var hasSources, recommended bool
	err := db.QueryRowContext(ctx, `
		SELECT
		  EXISTS (SELECT 1 FROM content_sources WHERE site_id = $1 AND is_active),
		  COALESCE((SELECT (data->'content_features'->'news_feed'->>'recommended')::boolean
		            FROM site_specs
		            WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
		            LIMIT 1), false)
	`, siteID).Scan(&hasSources, &recommended)
	if err != nil {
		// Fail open on read failure; the caller logs.
		return true, fmt.Sprintf("news source check failed (%v); fail-open", err)
	}
	if hasSources {
		return true, "active content_sources rows exist"
	}
	if recommended {
		return true, "content_features.news_feed.recommended=true (trigger will seed sources)"
	}
	return false, "no active content_sources and classification spec does not recommend news_feed (NEWS-001 never drives this site)"
}

// contentFeatureRecommended mirrors discovery_checks.directorySpecFlags'
// predicate: content_features.<specKey>.recommended in the current
// classification spec; absent anything = not opted in, not an error.
func contentFeatureRecommended(ctx context.Context, db *sql.DB, siteID uuid.UUID, specKey string) (bool, error) {
	var recommended sql.NullBool
	err := db.QueryRowContext(ctx, `
		SELECT (data->'content_features'->$2->>'recommended')::boolean
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'classification' AND is_current = true
		LIMIT 1
	`, siteID, specKey).Scan(&recommended)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return recommended.Valid && recommended.Bool, nil
}

// countPagesOfRole counts pages of one canonical role across the combined
// plan + realised view the caller assembled.
func countPagesOfRole(pages []planPageView, role string) int {
	n := 0
	for _, p := range pages {
		if p.Role == role {
			n++
		}
	}
	return n
}

// sectionPrefixOf derives the URL directory a section-index page indexes:
// "/guides/index.html" → "/guides/". Falls back to the page name stem
// ("guides-index" → "/guides/") when the plan carries no URL yet.
func sectionPrefixOf(page planPageView) string {
	if page.URL != "" {
		if i := strings.LastIndex(page.URL, "/"); i >= 0 {
			return page.URL[:i+1]
		}
	}
	stem := strings.TrimSuffix(page.Name, "-index")
	if stem != "" && stem != page.Name && stem != "index" {
		return "/" + stem + "/"
	}
	return ""
}

// countSectionChildren counts pages living under the section-index's directory
// prefix (excluding the index itself). A prefix that cannot be derived counts
// as unresolvable-by-shape and the CALLER's fail-open policy does not apply
// here — an underivable prefix returns 0 only when the page ALSO matched no
// component resolver, and the evidence string says why.
func countSectionChildren(page planPageView, allPages []planPageView) int {
	prefix := sectionPrefixOf(page)
	if prefix == "" || prefix == "/" {
		return 0
	}
	n := 0
	for _, p := range allPages {
		if p.Name == page.Name {
			continue
		}
		if p.URL != "" && strings.HasPrefix(p.URL, prefix) {
			n++
		}
	}
	return n
}
