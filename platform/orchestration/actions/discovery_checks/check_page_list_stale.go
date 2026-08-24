// FILE: platform/orchestration/actions/discovery_checks/check_page_list_stale.go
//
// page_list_stale — the SWEEP half of bugs_open/384. The event half
// (actions.requestPageListReresolve, register PBP-048) tells the listings when
// a card or page image lands. This check is the backstop for every producer
// that does not: it compares each page-list array a page has STORED in
// page_components.content_data against what its `query.*` source resolves NOW,
// and files a section_data_resolved re-render when an entry's image differs.
//
// WHAT IT COMPARES, AND WHAT IT DELIBERATELY DOES NOT
//   Only entries present in BOTH the stored array and the fresh resolve, matched
//   by url, and only their `image`. Membership drift (a page newly listed, or
//   no longer listed) is another check's business (orphan_pages,
//   cta_links_stale); title/meta drift is not what a landed image breaks. Being
//   narrow is what keeps this from churning: two arrays that agree on every
//   image are "current" whatever else differs about them.
//
// "COULD NOT TELL" IS NOT "CURRENT" (352 lane, 2026-08-24; RFC_010's rule).
//   A source that fails to resolve, or resolves to nothing, files nothing and
//   proves nothing — the page is logged as unknown, never counted as healthy.
//
// NO Resolved ARM, ON PURPOSE. A page_rerender is an ACTION REQUEST that runs
//   and completes by itself; it does not need retracting. And its key is SHARED
//   with every other section_data_resolved producer for that page (PBP-048
//   names them), so a retraction here — "the images match" — could close a
//   request another producer filed for a different reason (fresh news items,
//   resolved section data). A positive observation about one field is not an
//   observation about the whole key.
//
// SHAPE: page_rerender → page-rerender, status detected (the promoter dispatches
//   it — the (page_rerender, page-rerender) pair has thousands of completes), key
//   PageRerenderItemKey(page, site, "section_data_resolved") so it collapses
//   onto the event emitter's item when both fire for one page.
//
// ENABLEMENT: this file registers the name; the check runs only when
//   "page_list_stale" is in completeness-discovery-agent's run_checks.config.checks
//   (migration 603_enable_page_list_stale_HOLD.sql — held until the binary that
//   registers it has rolled; an unregistered name is warn-and-skip, so the
//   ordering is about not fooling the reader, not about breaking the agent).

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"go.uber.org/zap"
)

// PageListStaleCheck compares stored page-list arrays against a fresh resolve.
type PageListStaleCheck struct{}

func init() {
	Register(&PageListStaleCheck{})
}

func (c *PageListStaleCheck) Name() string { return "page_list_stale" }

// pageListStaleReason is the ONE reason value that makes page-rerender re-run
// query.* sources for a stored array (check_rerender_mode; STY-048).
const pageListStaleReason = "section_data_resolved"

// pageListStaleResolveLimit is queryresolve's hard cap: resolving at the cap
// means every entry a schema could have asked for is present to compare against.
const pageListStaleResolveLimit = 24

// storedPageListEntry is one element of a stored array, as far as this check
// reads it.
type storedPageListEntry struct {
	URL   string
	Image string
}

// staleEntry is one mismatch, carried in the item spec and the finding.
type staleEntry struct {
	Component    string `json:"component"`
	Field        string `json:"field"`
	Source       string `json:"source"`
	URL          string `json:"url"`
	StoredImage  string `json:"stored_image"`
	CurrentImage string `json:"current_image"`
}

func (c *PageListStaleCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	logger := dctx.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	result := &CheckResult{}

	consumers, err := queryresolve.PageListConsumerPages(dctx.Ctx, dctx.DB, dctx.SiteID, logger)
	if err != nil {
		return nil, fmt.Errorf("page_list_stale: consumer lookup failed: %w", err)
	}
	if len(consumers) == 0 {
		return result, nil
	}

	// One fresh resolve per distinct source per site. A nil map for a source
	// means "could not tell" for every page that consumes it.
	freshBySource := map[string]map[string]string{}
	freshImages := func(source string) (map[string]string, bool) {
		if m, seen := freshBySource[source]; seen {
			return m, m != nil
		}
		name := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(source)), "query.")
		val, rerr := queryresolve.Resolve(dctx.Ctx, dctx.DB, queryresolve.QueryRequest{
			Name: name, SiteID: dctx.SiteID, Limit: pageListStaleResolveLimit,
		}, logger)
		items, _ := val.([]map[string]interface{})
		if rerr != nil || len(items) == 0 {
			logger.Warn("page_list_stale: source did not resolve — its consumers are UNKNOWN this run, not current",
				zap.String("source", source), zap.String("site_id", dctx.SiteID.String()), zap.Error(rerr))
			freshBySource[source] = nil
			return nil, false
		}
		m := make(map[string]string, len(items))
		for _, it := range items {
			u, _ := it["url"].(string)
			img, _ := it["image"].(string)
			if u != "" {
				m[u] = img
			}
		}
		freshBySource[source] = m
		return m, true
	}

	pagesStale, pagesCurrent, pagesUnknown := 0, 0, 0
	for _, page := range consumers {
		stored, err := loadStoredPageLists(dctx, page.ID)
		if err != nil {
			logger.Warn("page_list_stale: could not read stored arrays — page UNKNOWN this run",
				zap.String("page", page.Name), zap.Error(err))
			pagesUnknown++
			continue
		}

		var stale []staleEntry
		compared, unknown := 0, false
		for _, f := range page.Fields {
			images, ok := freshImages(f.Source)
			if !ok {
				unknown = true
				continue
			}
			entries := stored[f.Component][f.Field]
			if len(entries) == 0 {
				continue // nothing stored → nothing to compare; the plan path fills it
			}
			compared++
			for _, e := range entries {
				want, listed := images[e.URL]
				if !listed {
					continue // membership is not this check's question
				}
				if want != e.Image {
					stale = append(stale, staleEntry{
						Component: f.Component, Field: f.Field, Source: f.Source,
						URL: e.URL, StoredImage: e.Image, CurrentImage: want,
					})
				}
			}
		}

		switch {
		case len(stale) > 0:
			pagesStale++
			pageID := page.ID
			specJSON, _ := json.Marshal(map[string]interface{}{
				"check":     c.Name(),
				"reason":    pageListStaleReason,
				"page_id":   page.ID.String(),
				"page_name": page.Name,
				"domain":    page.Domain,
				"cause":     "page_list_stale",
				"consumes":  page.Sources(),
				"stale":     stale,
				"fix": "A section_data_resolved re-render re-runs the page's query.* sources and " +
					"rewrites the stored array with the current card/hero image for every entry.",
			})
			result.WorkItems = append(result.WorkItems, WorkItemSpec{
				SiteID:       dctx.SiteID,
				PageID:       &pageID,
				Source:       "discovery",
				Pipeline:     dctx.Pipeline,
				ItemType:     "page_rerender",
				Severity:     "medium",
				Summary:      fmt.Sprintf("Page-list on %s shows %d stale image(s) — the stored array predates a landed card/hero; a section re-resolve refreshes it", page.Name, len(stale)),
				SpecJSON:     string(specJSON),
				Priority:     60,
				HandlerAgent: "page-rerender",
				Status:       "detected",
				CreatedBy:    dctx.AgentType,
				ItemKey:      PageRerenderItemKey(page.Name, dctx.SiteID, pageListStaleReason),
				BatchID:      dctx.BatchID,
			})
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":     c.Name(),
				"page_id":   page.ID.String(),
				"page_name": page.Name,
				"url":       page.URL,
				"stale":     stale,
			})
		case unknown || compared == 0:
			pagesUnknown++
		default:
			pagesCurrent++
		}
	}

	// One summary finding per run, so "unknown" is visible in the report and
	// testable — without it, unknown and current are indistinguishable from
	// outside (both file nothing), which is exactly the collapse RFC_010 warns
	// against.
	result.Findings = append(result.Findings, map[string]interface{}{
		"check":          c.Name(),
		"summary":        true,
		"consumer_pages": len(consumers),
		"stale":          pagesStale,
		"current":        pagesCurrent,
		"unknown":        pagesUnknown,
	})
	logger.Info("page_list_stale: swept",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("consumer_pages", len(consumers)),
		zap.Int("stale", pagesStale), zap.Int("current", pagesCurrent), zap.Int("unknown", pagesUnknown))
	return result, nil
}

// loadStoredPageLists returns, for one page, component name → field name →
// the stored array's (url, image) pairs, reading only array-typed values whose
// elements are objects. A field the page has no stored value for is absent.
func loadStoredPageLists(dctx DiscoveryCheckContext, pageID uuid.UUID) (map[string]map[string][]storedPageListEntry, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT cc.name, COALESCE(pc.content_data::text, '{}')
		  FROM page_components pc
		  JOIN content_components cc ON cc.id = pc.component_id
		 WHERE pc.page_id = $1
		   AND pc.build_status <> 'removed'
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string]map[string][]storedPageListEntry{}
	for rows.Next() {
		var component, raw string
		if err := rows.Scan(&component, &raw); err != nil {
			return nil, err
		}
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &data); err != nil {
			continue
		}
		for field, v := range data {
			arr, ok := v.([]interface{})
			if !ok {
				continue
			}
			var entries []storedPageListEntry
			for _, el := range arr {
				obj, ok := el.(map[string]interface{})
				if !ok {
					continue
				}
				u, _ := obj["url"].(string)
				if u == "" {
					continue
				}
				img, _ := obj["image"].(string)
				entries = append(entries, storedPageListEntry{URL: u, Image: img})
			}
			if len(entries) == 0 {
				continue
			}
			if out[component] == nil {
				out[component] = map[string][]storedPageListEntry{}
			}
			out[component][field] = entries
		}
	}
	return out, rows.Err()
}
