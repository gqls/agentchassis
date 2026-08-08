// FILE: platform/orchestration/actions/discovery_checks/check_phantom_internal_links.go
//
// Detects internal links in DEPLOYED rendered HTML whose target is not a real
// page, plus empty internal hrefs. This is the audit-time backstop for every
// link surface that bypasses build-time resolution:
//
//   - hero / call-to-action CTAs that resolved to a phantom (e.g. /services.html)
//   - header/footer literal links baked into site_components (e.g. /terms.html,
//     the hardcoded /contact.html "Get Started")
//   - "Browse All X" buttons left with href="" (unpopulated *_index_url specs)
//
// Why this exists: a phantom link is loud (present in deployed HTML, so it can be
// compared against the realised pages set), whereas a silently-dropped CTA leaves
// no fingerprint. The build-time link-resolution agent emits its own positive
// signal (unresolved_cta) BEFORE deploy; this check catches whatever reached
// deployed HTML wrong — regressions, literal links, and anything that slipped past.
//
// Substrate: page_components.rendered_html + site_components.rendered_html.
// link_registry is not used — it is currently unpopulated (extract_and_sync_links
// is wired into no live workflow).
//
// Extraction/classification/normalisation are the SHARED datahelpers definitions
// (ExtractHrefs / ClassifyLinkScope / PageURLSet), the same ones the deploy gate
// (validate_page_content) uses — so the gate and this audit agree, by one literal
// implementation, on what is an internal page link and what resolves to a real
// page. "Real page" = a pages row (status not deleted/archived).
//
// A page row that has NEVER BEEN DEPLOYED is reported separately, as
// `unbuilt_internal_link` (bugs_open/049 mechanism 2). This comment used to say
// such a page "has a row and is not flagged", which is the right answer for the
// build-time model and the wrong one here: this check runs against DEPLOYED HTML,
// where the only question a link raises is whether the target is fetchable.
// gaswholesalers.com/fuel-pricing-framework.html sat in 28 live footers as a 404
// for two months and was passed by this check by design.
//
// The two classes stay DISTINCT rather than merging into one "broken" verdict,
// because their remediations act on different rows: a phantom is fixed by
// rebuilding the page CONTAINING the link (build-time resolution re-runs), an
// unbuilt link by building the page it POINTS AT. Merging them would route half
// the findings at the wrong page.
//
// Routing — remediation by surface, mirroring existing improvement-loop pairs:
//   - site_component (header/footer/nav literals) -> nav-link-fixer (build):
//     its workflow force re-renders site components from real-page data.
//   - page_component (hero/CTA/body links)        -> page-build-handler (content):
//     a page rebuild re-runs page-content-writer, which runs build-time link
//     resolution (internal-link-resolver) — the resolver is a build-time
//     augmenter, not a rendered-HTML patcher, so findings trigger a rebuild
//     rather than routing to the resolver directly. Same handler pairing as
//     check_empty_sections.
//
// Registration: automatic via init() -> Register(&PhantomInternalLinksCheck{}).
// Enable by adding "phantom_internal_links" to a discovery agent's checks array.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&PhantomInternalLinksCheck{}) }

type PhantomInternalLinksCheck struct{}

func (c *PhantomInternalLinksCheck) Name() string { return "phantom_internal_links" }

func (c *PhantomInternalLinksCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	findings, err := findPhantomInternalLinks(dctx)
	if err != nil {
		return nil, err
	}
	if len(findings) == 0 {
		return &CheckResult{}, nil
	}

	result := &CheckResult{
		Findings: []map[string]interface{}{{
			"check":    "phantom_internal_links",
			"count":    len(findings),
			"findings": findings,
		}},
	}

	for _, f := range findings {
		handlerAgent, pipeline, priority := routeBySurface(f.Surface)

		severity := "high"
		switch f.IssueType {
		case "empty_internal_href":
			severity = "medium"
		case "dead_fragment_link":
			// An inert control, not a 404: the page loads, it simply does not
			// move. Graded below every arm that reaches a missing page.
			severity = "low"
		}

		spec := map[string]interface{}{
			"check":       "phantom_internal_links",
			"issue_type":  f.IssueType, // phantom_internal_link | empty_internal_href | unbuilt_internal_link
			"surface":     f.Surface,   // page_component | site_component
			"page_name":   f.PageName,
			"page_id":     f.PageID,
			"slot_name":   f.SlotName,
			"href":        f.Href,
			"occurrences": f.Occurrences,
			"fix": "Internal href has no matching pages.url row (or is empty). " +
				"Page surfaces: rebuild the page (build-time link resolution re-runs). " +
				"Site surfaces: re-render site components from real-page data.",
		}

		// Locate within the offending surface for the dedup key.
		locator := f.SlotName
		if f.PageName != "" {
			locator = f.PageName + ":" + f.SlotName
		}

		// Which page the work item is filed against. For a phantom or an empty
		// href that is the page CONTAINING the link, whose rebuild re-runs link
		// resolution. For an unbuilt link the target exists and is simply not
		// built, so the item is filed against the TARGET — rebuilding the
		// container would re-emit the same correct href and resolve nothing.
		actionablePageID := f.PageID
		summary := fmt.Sprintf("%s in %s (%s): href %q has no matching page",
			f.IssueType, f.Surface, locator, f.Href)

		if f.IssueType == "dead_fragment_link" {
			priority -= 10 // behind every arm that reaches a missing page
			spec["fix"] = "The href's #fragment names an element id that exists nowhere on the page it lands on, " +
				"so the control loads the page and does not move. Either point it at an id the target page " +
				"actually carries, or drop the fragment and link to the page. Page surfaces: rebuild the page. " +
				"Site surfaces: the chrome link resolves on NO page of this site — re-render site components."
			summary = fmt.Sprintf("dead_fragment_link in %s (%s): href %q names an element id no page carries",
				f.Surface, locator, f.Href)
		}

		if f.IssueType == "unbuilt_internal_link" {
			actionablePageID = f.TargetPageID
			handlerAgent, pipeline, priority = "page-build-handler", "content", 45
			spec["target_page_id"] = f.TargetPageID
			spec["fix"] = "Target page exists but has never been deployed (deployed_at IS NULL), " +
				"so this link is a live 404. Build and deploy the target page, or remove the link " +
				"and clear the page's in_header/in_footer flags if it is not wanted. " +
				"Do NOT rebuild the linking page — its href is correct."
			summary = fmt.Sprintf("unbuilt_internal_link in %s (%s): href %q points at a page that has never been deployed",
				f.Surface, locator, f.Href)
		}

		specJSON, _ := json.Marshal(spec)

		var pageIDPtr *uuid.UUID
		if actionablePageID != "" {
			if parsed, perr := uuid.Parse(actionablePageID); perr == nil {
				pageIDPtr = &parsed
			}
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     pipeline,
			ItemType:     f.IssueType,
			Severity:     severity,
			Summary:      summary,
			SpecJSON:     string(specJSON),
			Priority:     priority,
			HandlerAgent: handlerAgent,
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("%s:%s:%s:%s", f.IssueType, f.Surface, locator, f.Href),
			BatchID:      dctx.BatchID,
		})
	}

	dctx.Logger.Warn("phantom_internal_links: found broken/empty internal links",
		zap.Int("count", len(findings)),
		zap.String("site_id", dctx.SiteID.String()))

	return result, nil
}

// routeBySurface pairs each link surface with its remediation handler and
// pipeline, mirroring the existing improvement-loop pairs (broken_nav_links ->
// nav-link-fixer/build; empty_sections -> page-build-handler/content). A
// page_component phantom is fixed by REBUILDING the page (the rebuild re-runs
// build-time link resolution); it is not routed to internal-link-resolver
// directly, which is a build-time augmenter with no rendered-HTML surface.
func routeBySurface(surface string) (handlerAgent, pipeline string, priority int) {
	switch surface {
	case "site_component":
		return "nav-link-fixer", "build", 40
	default: // page_component
		return "page-build-handler", "content", 35
	}
}

type phantomLinkFinding struct {
	Surface  string `json:"surface"`
	PageName string `json:"page_name"`
	PageID   string `json:"page_id"`
	SlotName string `json:"slot_name"`
	Href     string `json:"href"`
	// TargetPageID is set only for unbuilt_internal_link: the never-deployed page
	// the href points AT. The work item is filed against this page, not PageID —
	// the fix is to build the target.
	TargetPageID string `json:"target_page_id,omitempty"`
	IssueType    string `json:"issue_type"`
	Occurrences  int    `json:"occurrences"`
}

// plKey identifies a distinct offending link occurrence for aggregation.
type plKey struct {
	surface, pageName, pageID, slotName, href, issue string
}

// findPhantomInternalLinks scans deployed page_components and site_components
// HTML, extracts hrefs with the shared datahelpers helpers, and returns the
// internal page links that match no real page (phantoms) plus empty hrefs,
// aggregated with occurrence counts.
func findPhantomInternalLinks(dctx DiscoveryCheckContext) ([]phantomLinkFinding, error) {
	targets, err := loadSitePageTargets(dctx)
	if err != nil {
		return nil, err
	}

	counts := make(map[plKey]int)
	targetIDs := make(map[plKey]string) // unbuilt_internal_link -> target page id

	// The fragment arm (dead_fragment_link, this file's sibling) resolves an
	// href's #fragment against the ids of the page it LANDS on, which is not
	// the component the href was found in — so the components are collected as
	// they stream past and judged in a second pass, once every document is
	// known. The phantom/empty arms below are unchanged and still judge inline:
	// they need nothing but the href and the pages table.
	type componentHTML struct{ surface, pageName, pageID, slotName, html string }
	var components []componentHTML
	pageHTML := make(map[string]string)     // page id -> its components concatenated
	pagePathByID := make(map[string]string) // normalised page path -> page id
	var chrome strings.Builder

	// page_components — body sections (hero/CTA/lists/prose).
	pageRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, pc.page_id::text, COALESCE(pc.slot_name, ''), pc.rendered_html,
		       COALESCE(p.url, '')
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("phantom_internal_links page query failed: %w", err)
	}
	defer pageRows.Close()
	for pageRows.Next() {
		var pageName, pageID, slotName, html, pageURL string
		if err := pageRows.Scan(&pageName, &pageID, &slotName, &html, &pageURL); err != nil {
			dctx.Logger.Warn("phantom_internal_links: page scan error", zap.Error(err))
			continue
		}
		accumulateLinkIssues(counts, targetIDs, "page_component", pageName, pageID, slotName, html, targets)
		components = append(components, componentHTML{"page_component", pageName, pageID, slotName, html})
		pageHTML[pageID] = pageHTML[pageID] + "\n" + html
		if pageURL != "" {
			pagePathByID[datahelpers.NormalizePagePath(pageURL)] = pageID
		}
	}
	if err := pageRows.Err(); err != nil {
		return nil, fmt.Errorf("phantom_internal_links page iteration failed: %w", err)
	}

	// site_components — header/footer/head (literal nav/legal links).
	siteRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT COALESCE(sc.slot_name, ''), sc.rendered_html
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.rendered_html IS NOT NULL AND sc.rendered_html <> ''
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("phantom_internal_links site query failed: %w", err)
	}
	defer siteRows.Close()
	for siteRows.Next() {
		var slotName, html string
		if err := siteRows.Scan(&slotName, &html); err != nil {
			dctx.Logger.Warn("phantom_internal_links: site scan error", zap.Error(err))
			continue
		}
		accumulateLinkIssues(counts, targetIDs, "site_component", "", "", slotName, html, targets)
		components = append(components, componentHTML{"site_component", "", "", slotName, html})
		chrome.WriteString("\n")
		chrome.WriteString(html)
	}
	if err := siteRows.Err(); err != nil {
		return nil, fmt.Errorf("phantom_internal_links site iteration failed: %w", err)
	}

	// Second pass: the fragment arm, now that every document is known.
	fragIdx := newFragmentIndex(pageHTML, chrome.String(), pagePathByID)
	for _, c := range components {
		accumulateFragmentIssues(counts, c.surface, c.pageName, c.pageID, c.slotName, c.html, targets, fragIdx)
	}

	findings := make([]phantomLinkFinding, 0, len(counts))
	for k, n := range counts {
		findings = append(findings, phantomLinkFinding{
			Surface:      k.surface,
			PageName:     k.pageName,
			PageID:       k.pageID,
			SlotName:     k.slotName,
			Href:         k.href,
			TargetPageID: targetIDs[k], // empty for every type but unbuilt_internal_link
			IssueType:    k.issue,
			Occurrences:  n,
		})
	}
	// Deterministic order (map iteration is random) — stable findings/work items.
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Surface != findings[j].Surface {
			return findings[i].Surface < findings[j].Surface
		}
		if findings[i].PageName != findings[j].PageName {
			return findings[i].PageName < findings[j].PageName
		}
		if findings[i].Href != findings[j].Href {
			return findings[i].Href < findings[j].Href
		}
		return findings[i].IssueType < findings[j].IssueType
	})
	return findings, nil
}

// accumulateLinkIssues extracts hrefs from one rendered component and records
// empty hrefs and phantom page links into counts.
//
// Runtime-fill guard: a data-runtime-fill section is DELIBERATELY empty at build
// time — its hrefs are placeholders filled client-side (vonc's provocation-card /
// lobby-grid shells), so an empty href there is by-design, not a defect. Empty
// hrefs are exempted for such shells (matching check_empty_sections' exemption).
// A NON-empty href that resolves to no real page is still flagged even in a
// shell: a phantom target baked into the template is a genuine bug, not a
// placeholder.
func accumulateLinkIssues(counts map[plKey]int, targetIDs map[plKey]string, surface, pageName, pageID, slotName, html string, targets sitePageTargets) {
	shells := datahelpers.RuntimeFillSpans(html)
	for _, m := range datahelpers.HrefOffsets(html) {
		href := m.Value
		switch datahelpers.ClassifyLinkScope(href) {
		case datahelpers.LinkScopeEmpty:
			if shells.Contains(m.Offset) {
				continue // empty href is intended in a client-hydrated shell
			}
			counts[plKey{surface, pageName, pageID, slotName, href, "empty_internal_href"}]++
		case datahelpers.LinkScopePage:
			if !targets.valid.Contains(href) {
				counts[plKey{surface, pageName, pageID, slotName, href, "phantom_internal_link"}]++
				continue
			}
			// Resolves to a row, but that row has never been deployed — the link
			// is live and 404s. Carry the TARGET page id: remediation builds the
			// page pointed at, not the page containing the link.
			if targetID, ok := targets.unbuilt[datahelpers.NormalizePagePath(href)]; ok {
				k := plKey{surface, pageName, pageID, slotName, href, "unbuilt_internal_link"}
				counts[k]++
				targetIDs[k] = targetID
			}
		}
		// external / anchor / mailto / asset are not internal page links — skip.
	}
}

// sitePageTargets is what one site's pages table says about link targets: the
// set that resolves to a row at all, and the subset of those that has never
// been deployed and therefore 404s despite having a row.
type sitePageTargets struct {
	valid datahelpers.PageURLSet
	// unbuilt maps a NORMALISED page path to the id of the never-deployed page
	// at that path. Normalised so lookups can reuse datahelpers.NormalizePagePath
	// and agree with PageURLSet on what "same target" means.
	unbuilt map[string]string
}

// The "never shipped, so a link to it is a live 404" predicate moved to
// datahelpers.NeverDeployedPagePredicate on 2026-07-26, with its measurement
// rationale, so the chrome renderer can decide by the same rule this audit
// judges by. Chrome was writing nav links to never-deployed pages while this
// check flagged them — bugs_open/049 mechanism 2.

// loadSitePageTargets builds the normalised set of real page targets for the
// site, using the same page selection and normalisation as the deploy gate, and
// alongside it the never-deployed subset.
func loadSitePageTargets(dctx DiscoveryCheckContext) (sitePageTargets, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT url, id::text, (`+datahelpers.NeverDeployedPagePredicate+`) AS never_deployed
		FROM pages
		WHERE site_id = $1 AND status NOT IN ('deleted', 'archived')
	`, dctx.SiteID)
	if err != nil {
		return sitePageTargets{}, fmt.Errorf("phantom_internal_links pages query failed: %w", err)
	}
	defer rows.Close()

	var urls []string
	unbuilt := make(map[string]string)
	built := make(map[string]bool)

	for rows.Next() {
		var url, pageID string
		var neverDeployed bool
		if err := rows.Scan(&url, &pageID, &neverDeployed); err != nil {
			continue
		}
		urls = append(urls, url)

		// Two rows can normalise to one target (/news.html vs /news/index.html).
		// If ANY row at a path is built, the path is fetchable and must not be
		// reported — so track built paths and let them win regardless of the
		// order rows arrive in.
		norm := datahelpers.NormalizePagePath(url)
		if neverDeployed {
			if _, seen := unbuilt[norm]; !seen {
				unbuilt[norm] = pageID
			}
		} else {
			built[norm] = true
		}
	}
	if err := rows.Err(); err != nil {
		return sitePageTargets{}, fmt.Errorf("phantom_internal_links pages iteration failed: %w", err)
	}

	for norm := range built {
		delete(unbuilt, norm)
	}

	urls = append(urls, "/", "/index.html") // site root is always valid
	// The root is served by the site itself; never report it as unbuilt even if
	// some row claims that path.
	delete(unbuilt, datahelpers.NormalizePagePath("/"))

	return sitePageTargets{valid: datahelpers.NewPageURLSet(urls), unbuilt: unbuilt}, nil
}

func init() { RegisterVerifier("unbuilt_internal_link", VerifyUnbuiltInternalLinkResolved) }

// VerifyUnbuiltInternalLinkResolved re-checks one unbuilt_internal_link item at
// completion time (bugs_open/220): the handler saga can succeed having rebuilt
// the page CONTAINING the link — the dispatch defect this bug is about — and
// until this verifier existed that stamped the item 'complete' with the target
// still a live 404, terminal for the dedup index, so the next discovery pass
// re-minted the same item and the loop converged on nothing, green, forever.
//
// BOTH DISJUNCTS ARE THE FIX, exactly as for VerifyDeadFragmentLinkResolved:
// the item's own spec.fix names two remedies — build and deploy the target, or
// remove the link — so the defect is gone when the container no longer renders
// the href, OR when the target page has shipped. A verifier asking only about
// the target would refuse an item fixed the way the item itself recommended.
//
// The shipped/unbuilt judgement is datahelpers.NeverDeployedPagePredicate — the
// SAME SQL predicate the detector minted the finding by, which is verifiers.go's
// stated contract, and the same reason the predicate is a shared constant at all
// (each hand-written copy is a chance to spell it `build_status = 'deployed'`
// and disagree with the check).
//
// Remit check (the page_rerender trap in verifier_coverage_test.go): the handler
// for this item type is pointed at the TARGET page and its remit is to build and
// deploy it, so "target shipped" is not stricter than what the handler is
// responsible for; link-removal is the documented alternative remedy, accepted
// not required.
func VerifyUnbuiltInternalLinkResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	href, _ := target.Spec["href"].(string)
	if href == "" {
		return VerifyResult{}, fmt.Errorf("unbuilt_internal_link verifier: item carries no href")
	}
	surface, _ := target.Spec["surface"].(string)

	// 1. Is the offending href still rendered on the surface it was found on?
	//    NOTE the container/target split: spec.page_id is the page CONTAINING
	//    the link; target.PageID (the work item's page_id column) is the TARGET
	//    the href points at. Confusing the two is this bug's whole mechanism.
	var stillPresent bool
	var err error
	if surface == "site_component" {
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM site_components
				WHERE site_id = $1 AND COALESCE(rendered_html, '') LIKE '%' || $2 || '%'
			)`, target.SiteID, `href="`+href+`"`).Scan(&stillPresent)
	} else {
		containerIDStr, _ := target.Spec["page_id"].(string)
		containerID, perr := uuid.Parse(containerIDStr)
		if perr != nil {
			return VerifyResult{}, fmt.Errorf("unbuilt_internal_link verifier: page-surface item carries no container page_id in spec (%q): %w", containerIDStr, perr)
		}
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM page_components
				WHERE page_id = $1 AND COALESCE(rendered_html, '') LIKE '%' || $2 || '%'
			)`, containerID, `href="`+href+`"`).Scan(&stillPresent)
	}
	if err != nil {
		return VerifyResult{}, fmt.Errorf("unbuilt_internal_link verifier: surface load failed: %w", err)
	}
	if !stillPresent {
		return VerifyResult{Resolved: true, Detail: fmt.Sprintf("href %q is no longer rendered on this %s — the link was removed, which the item's own fix text accepts", href, surface)}, nil
	}

	// 2. The link is still live, so the target must now have shipped.
	if target.PageID == nil {
		return VerifyResult{}, fmt.Errorf("unbuilt_internal_link verifier: item carries no target page_id")
	}
	var stillUnbuilt bool
	err = db.QueryRowContext(ctx, `
		SELECT `+datahelpers.NeverDeployedPagePredicate+`
		FROM pages
		WHERE id = $1 AND site_id = $2 AND status NOT IN ('deleted', 'archived')
	`, *target.PageID, target.SiteID).Scan(&stillUnbuilt)
	if err == sql.ErrNoRows {
		// The target row is gone (deleted or archived) while the link is still
		// rendered: the href 404s with no page left to build. Resolved:false,
		// not an error — this is a judged defect, and the phantom arm of the
		// same check is what will re-classify the finding on the next pass.
		return VerifyResult{Resolved: false, Detail: fmt.Sprintf("href %q is still rendered and its target page row %s is deleted/archived — the link now dangles with nothing to build", href, target.PageID)}, nil
	}
	if err != nil {
		return VerifyResult{}, fmt.Errorf("unbuilt_internal_link verifier: target page load failed: %w", err)
	}
	if stillUnbuilt {
		return VerifyResult{Resolved: false, Detail: fmt.Sprintf("target page %s has still never been deployed and href %q is still rendered — a handler that rebuilt the CONTAINER instead of the target lands here (bugs_open/220)", target.PageID, href)}, nil
	}
	return VerifyResult{Resolved: true, Detail: fmt.Sprintf("target page %s has shipped; href %q now resolves", target.PageID, href)}, nil
}
