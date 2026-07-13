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
// page. "Real page" = a pages row (status not deleted/archived); a planned-but-
// unbuilt page has a row and is not flagged.
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
	"encoding/json"
	"fmt"
	"sort"

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
		if f.IssueType == "empty_internal_href" {
			severity = "medium"
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":       "phantom_internal_links",
			"issue_type":  f.IssueType, // phantom_internal_link | empty_internal_href
			"surface":     f.Surface,   // page_component | site_component
			"page_name":   f.PageName,
			"page_id":     f.PageID,
			"slot_name":   f.SlotName,
			"href":        f.Href,
			"occurrences": f.Occurrences,
			"fix": "Internal href has no matching pages.url row (or is empty). " +
				"Page surfaces: rebuild the page (build-time link resolution re-runs). " +
				"Site surfaces: re-render site components from real-page data.",
		})

		// Locate within the offending surface for the dedup key.
		locator := f.SlotName
		if f.PageName != "" {
			locator = f.PageName + ":" + f.SlotName
		}

		var pageIDPtr *uuid.UUID
		if f.PageID != "" {
			if parsed, perr := uuid.Parse(f.PageID); perr == nil {
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
			Summary:      fmt.Sprintf("%s in %s (%s): href %q has no matching page", f.IssueType, f.Surface, locator, f.Href),
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
	Surface     string `json:"surface"`
	PageName    string `json:"page_name"`
	PageID      string `json:"page_id"`
	SlotName    string `json:"slot_name"`
	Href        string `json:"href"`
	IssueType   string `json:"issue_type"`
	Occurrences int    `json:"occurrences"`
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
	validPages, err := loadSitePageURLSet(dctx)
	if err != nil {
		return nil, err
	}

	counts := make(map[plKey]int)

	// page_components — body sections (hero/CTA/lists/prose).
	pageRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, pc.page_id::text, COALESCE(pc.slot_name, ''), pc.rendered_html
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
		var pageName, pageID, slotName, html string
		if err := pageRows.Scan(&pageName, &pageID, &slotName, &html); err != nil {
			dctx.Logger.Warn("phantom_internal_links: page scan error", zap.Error(err))
			continue
		}
		accumulateLinkIssues(counts, "page_component", pageName, pageID, slotName, html, validPages)
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
		accumulateLinkIssues(counts, "site_component", "", "", slotName, html, validPages)
	}
	if err := siteRows.Err(); err != nil {
		return nil, fmt.Errorf("phantom_internal_links site iteration failed: %w", err)
	}

	findings := make([]phantomLinkFinding, 0, len(counts))
	for k, n := range counts {
		findings = append(findings, phantomLinkFinding{
			Surface:     k.surface,
			PageName:    k.pageName,
			PageID:      k.pageID,
			SlotName:    k.slotName,
			Href:        k.href,
			IssueType:   k.issue,
			Occurrences: n,
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
func accumulateLinkIssues(counts map[plKey]int, surface, pageName, pageID, slotName, html string, validPages datahelpers.PageURLSet) {
	for _, href := range datahelpers.ExtractHrefs(html) {
		switch datahelpers.ClassifyLinkScope(href) {
		case datahelpers.LinkScopeEmpty:
			counts[plKey{surface, pageName, pageID, slotName, href, "empty_internal_href"}]++
		case datahelpers.LinkScopePage:
			if !validPages.Contains(href) {
				counts[plKey{surface, pageName, pageID, slotName, href, "phantom_internal_link"}]++
			}
		}
		// external / anchor / mailto / asset are not internal page links — skip.
	}
}

// loadSitePageURLSet builds the normalised set of real page targets for the
// site, using the same page selection and normalisation as the deploy gate.
func loadSitePageURLSet(dctx DiscoveryCheckContext) (datahelpers.PageURLSet, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT url FROM pages
		WHERE site_id = $1 AND status NOT IN ('deleted', 'archived')
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("phantom_internal_links pages query failed: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("phantom_internal_links pages iteration failed: %w", err)
	}
	urls = append(urls, "/", "/index.html") // site root is always valid

	return datahelpers.NewPageURLSet(urls), nil
}
