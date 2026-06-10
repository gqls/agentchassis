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
// Authority for "is this a real page": a row in `pages`. A valid internal link
// normalises to some pages.url; a phantom matches none. We compare on a forgiving
// normal form (strip trailing index.html and slashes, drop ?query/#fragment) so
// /tools/, /tools/index.html and /tools all compare equal and do not false-flag.
//
// Routing — distinct responsibilities by surface:
//   - site_component (header/footer/nav literals) -> nav-link-fixer
//   - page_component (hero/CTA/body links)        -> internal-link-resolver
//
// Registration: automatic via init() -> Register(&PhantomInternalLinksCheck{}).
// Enable by adding "phantom_internal_links" to a discovery agent's checks array.

package discovery_checks

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
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
		handlerAgent, priority := routeBySurface(f.Surface)

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
				"Resolve the destination against real pages, or remove the link.",
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
			Pipeline:     "build",
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

// routeBySurface keeps the two link surfaces under distinct, single-responsibility
// handlers. Header/footer literals are a site-component concern; page-body CTAs are
// a link-resolution concern. Both agents must exist before enabling this check.
func routeBySurface(surface string) (handlerAgent string, priority int) {
	switch surface {
	case "site_component":
		return "nav-link-fixer", 40
	default: // page_component
		return "internal-link-resolver", 35
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

// findPhantomInternalLinks extracts every href from deployed page_components and
// site_components HTML, keeps the internal page links (and empty hrefs), and
// returns those whose normalised target matches no row in `pages`.
//
// "Real page" = existence of a pages row for the site, regardless of build_status:
// a planned-but-not-yet-deployed page is a valid target and must not be flagged.
const phantomInternalLinksSQL = `
WITH real_pages AS (
    SELECT DISTINCT
        regexp_replace(regexp_replace(url, 'index\.html$', ''), '/+$', '') AS norm_url
    FROM pages
    WHERE site_id = $1
),
raw_links AS (
    SELECT 'page_component'::text         AS surface,
           p.name                         AS page_name,
           pc.page_id::text               AS page_id,
           COALESCE(pc.slot_name, '')     AS slot_name,
           (regexp_matches(pc.rendered_html, 'href="([^"]*)"', 'g'))[1] AS href
    FROM page_components pc
    JOIN pages p ON p.id = pc.page_id
    WHERE p.site_id = $1
      AND pc.rendered_html IS NOT NULL
      AND pc.rendered_html <> ''
    UNION ALL
    SELECT 'site_component'::text         AS surface,
           ''                             AS page_name,
           ''                             AS page_id,
           COALESCE(sc.slot_name, '')     AS slot_name,
           (regexp_matches(sc.rendered_html, 'href="([^"]*)"', 'g'))[1] AS href
    FROM site_components sc
    WHERE sc.site_id = $1
      AND sc.rendered_html IS NOT NULL
      AND sc.rendered_html <> ''
),
classified AS (
    SELECT surface, page_name, page_id, slot_name, href,
           CASE
               WHEN href = '' THEN 'empty'
               WHEN href LIKE '/%' AND href NOT LIKE '//%'
                    AND split_part(split_part(href, '#', 1), '?', 1) ~ '\.html$'
                   THEN 'page_link'
               ELSE 'skip'
           END AS kind,
           regexp_replace(
               regexp_replace(
                   split_part(split_part(href, '#', 1), '?', 1),
                   'index\.html$', ''),
               '/+$', '') AS norm_href
    FROM raw_links
)
SELECT c.surface, c.page_name, c.page_id, c.slot_name, c.href,
       CASE WHEN c.kind = 'empty' THEN 'empty_internal_href'
            ELSE 'phantom_internal_link' END AS issue_type,
       COUNT(*)::int AS occurrences
FROM classified c
LEFT JOIN real_pages rp ON rp.norm_url = c.norm_href
WHERE c.kind = 'empty'
   OR (c.kind = 'page_link' AND rp.norm_url IS NULL)
GROUP BY c.surface, c.page_name, c.page_id, c.slot_name, c.href, issue_type
ORDER BY c.surface, c.page_name, c.href
`

func findPhantomInternalLinks(dctx DiscoveryCheckContext) ([]phantomLinkFinding, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, phantomInternalLinksSQL, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("phantom_internal_links query failed: %w", err)
	}
	defer rows.Close()

	var findings []phantomLinkFinding
	for rows.Next() {
		var f phantomLinkFinding
		var pageName, pageID, slotName, href sql.NullString
		if err := rows.Scan(&f.Surface, &pageName, &pageID, &slotName, &href, &f.IssueType, &f.Occurrences); err != nil {
			dctx.Logger.Warn("phantom_internal_links: scan error", zap.Error(err))
			continue
		}
		f.PageName = pageName.String
		f.PageID = pageID.String
		f.SlotName = slotName.String
		f.Href = href.String
		findings = append(findings, f)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("phantom_internal_links row iteration failed: %w", err)
	}
	return findings, nil
}
