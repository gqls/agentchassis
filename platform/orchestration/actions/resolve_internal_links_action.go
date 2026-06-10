// FILE: platform/orchestration/actions/resolve_internal_links_action.go
//
// ResolveInternalLinksAction resolves intent-appropriate internal link
// destinations for a page's CTA-bearing sections, from the real pages — never a
// hardcoded or fabricated target. It is the core action of the
// internal-link-resolver agent.
//
// v1 rule (deterministic, generic): a page's primary/secondary CTA point at the
// site's top content hubs (page_type='section-index', ordered by nav_order,
// excluding about/contact/legal areas, skipping the page itself). Every target
// is a real page; an absent hub yields no target (the gated templates render no
// button) and an unresolved CTA is reported so the absence is not silent. The
// agent bubble lets this be upgraded later (e.g. LLM intent-matching: a guide ->
// its related tool) without changing callers.
//
// Returned shape:
//   {
//     "cta_targets": { "primary_cta_url": "/tools/index.html",
//                      "secondary_cta_url": "/guides/index.html" },
//     "unresolved":  [ "primary_cta_url" ],   // requested but no real target
//     "page_type":   "landing"
//   }
// The caller (page-content-writer) merges cta_targets into the section
// resolved_data before render; "unresolved" feeds an unresolved_cta signal.
//
// Registration:
//   "resolve_internal_links": {
//       Handler:     ResolveInternalLinksAction,
//       Category:    "content",
//       Description: "Resolve intent-appropriate internal CTA destinations from real pages",
//       IsLocal:     true,
//   }

package actions

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ResolveInternalLinksInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"page_id", "page_name", "page_type"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("resolve_internal_links", ResolveInternalLinksInputSpec)
}

// contentHub is a section-index page eligible as a CTA destination.
type contentHub struct {
	Name     string
	URL      string
	Area     string // first path segment, e.g. "tools"
	NavOrder int
}

// areasExcludedFromCTA are section areas that are not content destinations for a
// hero/CTA (utility/legal pages).
var areasExcludedFromCTA = map[string]bool{
	"about": true, "contact": true, "privacy": true, "terms": true, "legal": true,
}

func ResolveInternalLinksAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "resolve_internal_links"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(params.CollectedData, params.StepConfig.Config, ResolveInternalLinksInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	pageName := inputs.Get("page_name")
	pageType := inputs.Get("page_type")
	pageID := inputs.Get("page_id")

	// Resolve page_type if only an id/name was given (rules key off type).
	if pageType == "" && (pageID != "" || pageName != "") {
		pageType = loadPageType(ctx, params, siteID, pageID, pageName, logger)
	}

	hubs, err := loadContentHubs(ctx, params, siteID, logger)
	if err != nil {
		return nil, err
	}

	primary, secondary := chooseCTATargets(pageType, pageName, hubs)

	// Validate against real pages (belt-and-suspenders — targets come from the
	// pages table already, but this guarantees no phantom can ever leave here).
	validPages := loadResolverPageSet(ctx, params, siteID, logger)
	ctaTargets := map[string]interface{}{}
	var unresolved []string

	if primary != "" && validPages.Contains(primary) {
		ctaTargets["primary_cta_url"] = primary
	} else {
		unresolved = append(unresolved, "primary_cta_url")
	}
	if secondary != "" && validPages.Contains(secondary) {
		ctaTargets["secondary_cta_url"] = secondary
	} else {
		unresolved = append(unresolved, "secondary_cta_url")
	}

	logger.Info("resolve_internal_links: resolved CTA targets",
		zap.String("page_type", pageType),
		zap.String("page_name", pageName),
		zap.Int("hub_count", len(hubs)),
		zap.Int("resolved", len(ctaTargets)),
		zap.Strings("unresolved", unresolved))

	return map[string]interface{}{
		"cta_targets": ctaTargets,
		"unresolved":  unresolved,
		"page_type":   pageType,
	}, nil
}

// chooseCTATargets returns (primaryURL, secondaryURL) for the page. v1: the top
// two content hubs by nav_order, excluding the page's own hub. Empty string
// means "no sensible target" (the caller drops the button + reports unresolved).
func chooseCTATargets(pageType, pageName string, hubs []contentHub) (string, string) {
	ordered := make([]contentHub, 0, len(hubs))
	for _, h := range hubs {
		if areasExcludedFromCTA[h.Area] {
			continue
		}
		if h.Name == pageName { // don't point a page's hero at itself
			continue
		}
		ordered = append(ordered, h)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].NavOrder != ordered[j].NavOrder {
			return ordered[i].NavOrder < ordered[j].NavOrder
		}
		return ordered[i].Name < ordered[j].Name
	})

	var primary, secondary string
	if len(ordered) > 0 {
		primary = ordered[0].URL
	}
	if len(ordered) > 1 {
		secondary = ordered[1].URL
	}
	return primary, secondary
}

func loadContentHubs(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) ([]contentHub, error) {
	rows, err := params.DB.QueryContext(ctx, `
		SELECT name, url, COALESCE(nav_order, 100)
		FROM pages
		WHERE site_id = $1
		  AND page_type = 'section-index'
		  AND status IN ('active', 'deployed')
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("loadContentHubs query failed: %w", err)
	}
	defer rows.Close()

	var hubs []contentHub
	for rows.Next() {
		var h contentHub
		if err := rows.Scan(&h.Name, &h.URL, &h.NavOrder); err != nil {
			logger.Warn("loadContentHubs: scan error", zap.Error(err))
			continue
		}
		h.Area = firstPathSegment(h.URL)
		hubs = append(hubs, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("loadContentHubs iteration failed: %w", err)
	}
	return hubs, nil
}

func loadPageType(ctx context.Context, params ActionParams, siteID uuid.UUID, pageID, pageName string, logger *zap.Logger) string {
	var pt string
	var err error
	if pageID != "" {
		if id, perr := uuid.Parse(pageID); perr == nil {
			err = params.DB.QueryRowContext(ctx,
				`SELECT COALESCE(page_type, '') FROM pages WHERE id = $1`, id).Scan(&pt)
		}
	} else if pageName != "" {
		err = params.DB.QueryRowContext(ctx,
			`SELECT COALESCE(page_type, '') FROM pages WHERE site_id = $1 AND name = $2`, siteID, pageName).Scan(&pt)
	}
	if err != nil {
		logger.Warn("resolve_internal_links: could not load page_type", zap.Error(err))
		return ""
	}
	return pt
}

func loadResolverPageSet(ctx context.Context, params ActionParams, siteID uuid.UUID, logger *zap.Logger) datahelpers.PageURLSet {
	rows, err := params.DB.QueryContext(ctx, `
		SELECT url FROM pages WHERE site_id = $1 AND status NOT IN ('deleted', 'archived')
	`, siteID)
	if err != nil {
		logger.Warn("resolve_internal_links: page set load failed", zap.Error(err))
		return datahelpers.NewPageURLSet(nil)
	}
	defer rows.Close()
	var urls []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			continue
		}
		urls = append(urls, u)
	}
	return datahelpers.NewPageURLSet(urls)
}

// firstPathSegment("/tools/index.html") -> "tools"; "/index.html" -> "".
func firstPathSegment(url string) string {
	trimmed := strings.TrimPrefix(url, "/")
	if i := strings.IndexByte(trimmed, '/'); i >= 0 {
		return trimmed[:i]
	}
	return ""
}
