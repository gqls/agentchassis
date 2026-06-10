// FILE: platform/orchestration/actions/resolve_internal_links_action.go
//
// ResolveInternalLinksAction is the core action of the internal-link-resolver
// agent. It augments a page's CTA-bearing sections (hero, call-to-action) with
// intent-appropriate internal link destinations resolved from the REAL pages —
// never a hardcoded or fabricated target — writing them into each section's
// resolved_data so the existing render path (render_component's merge_with:
// current_section.resolved_data) picks them up with no render-loop change.
//
// Contract:
//   in : site_id (required), page_type, page_name, and a `sections` config PATH
//        pointing at section_plan.sections_ready
//   out: { "sections_ready": [...augmented...],
//          "unresolved":     [ {section, component, field, slot} ... ] }
// The caller (page-content-writer) iterates the returned sections_ready; an
// unresolved entry feeds the build-time unresolved_cta signal.
//
// v1 rule (deterministic, generic): primary/secondary CTA point at the site's
// top content hubs (page_type='section-index', by nav_order, excluding
// about/contact/legal, skipping the page's own hub). Real, validated, never a
// phantom; absent hub -> field left unset (gated template renders no button) and
// reported unresolved. The agent boundary lets this be upgraded (LLM intent-
// matching: a guide -> its related tool) without changing callers — page_type is
// carried for that future use.
//
// Field names differ by component:
//   hero            -> cta_url (primary), secondary_cta_url (secondary)
//   call-to-action  -> primary_cta_url (primary), secondary_cta_url (secondary)
//
// Input handling follows 003/001 contracts: site_id/page_type/page_name are
// scalars via ExtractActionInputs; `sections` is a complex (array) value whose
// config holds a PATH, so it is resolved directly from collected_data. `sections`
// is deliberately NOT an InputSpec field — that name collides with the
// pages.sections column reachable through the current_page nested-source loop.
//
// Registration (action_registry.go):
//   "resolve_internal_links": { Handler: ResolveInternalLinksAction,
//       Category: "content", IsLocal: true }

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
	Optional:   []string{"page_type", "page_name"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("resolve_internal_links", ResolveInternalLinksInputSpec)
}

type contentHub struct {
	Name     string
	URL      string
	Area     string
	NavOrder int
}

var areasExcludedFromCTA = map[string]bool{
	"about": true, "contact": true, "privacy": true, "terms": true, "legal": true,
}

// ctaFieldNames maps a CTA component to its primary/secondary url field names.
var ctaFieldNames = map[string][2]string{
	"hero":           {"cta_url", "secondary_cta_url"},
	"call-to-action": {"primary_cta_url", "secondary_cta_url"},
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
	pageType := inputs.Get("page_type")
	pageName := inputs.Get("page_name")

	// `sections` config value is a PATH (e.g. "input_data.section_plan.sections_ready").
	// Resolve it directly from collected_data: it is a complex array value (not a
	// scalar ExtractActionInputs could return) and keeping it out of the InputSpec
	// avoids the current_page.sections field-name collision.
	var sections []interface{}
	if sectionsPath, ok := params.StepConfig.Config["sections"].(string); ok && sectionsPath != "" {
		if raw := datahelpers.ExtractNestedField(params.CollectedData, sectionsPath); raw != nil {
			sections, _ = raw.([]interface{})
		}
	}

	hubs, err := loadContentHubs(ctx, params, siteID, logger)
	if err != nil {
		return nil, err
	}
	validPages := loadResolverPageSet(ctx, params, siteID, logger)
	primary, secondary := chooseCTATargets(pageType, pageName, hubs)

	var unresolved []map[string]interface{}
	for _, raw := range sections {
		section, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		function := sectionComponentFunction(section)
		fields, isCTA := ctaFieldNames[function]
		if !isCTA {
			continue
		}
		sectionName := stringOrEmpty(section["name"])
		resolved := sectionResolvedData(section)
		setCTAField(resolved, fields[0], primary, validPages, function, sectionName, "primary", &unresolved)
		setCTAField(resolved, fields[1], secondary, validPages, function, sectionName, "secondary", &unresolved)
		section["resolved_data"] = resolved
	}

	logger.Info("resolve_internal_links: augmented CTA sections",
		zap.String("page_type", pageType),
		zap.String("page_name", pageName),
		zap.Int("section_count", len(sections)),
		zap.Int("hub_count", len(hubs)),
		zap.Int("unresolved", len(unresolved)))

	return map[string]interface{}{
		"sections_ready": sections,
		"unresolved":     unresolved,
	}, nil
}

// setCTAField writes a validated url into resolved_data, or records it unresolved.
func setCTAField(resolved map[string]interface{}, field, url string, validPages datahelpers.PageURLSet,
	function, sectionName, slot string, unresolved *[]map[string]interface{}) {
	if url != "" && validPages.Contains(url) {
		resolved[field] = url
		return
	}
	*unresolved = append(*unresolved, map[string]interface{}{
		"section":   sectionName,
		"component": function,
		"field":     field,
		"slot":      slot,
	})
}

// chooseCTATargets — v1: top two content hubs by nav_order, excluding the page's
// own hub (by name) and utility/legal areas. Empty string => no sensible target.
// pageType is carried for a future intent-aware (LLM) upgrade; v1 does not branch
// on it.
func chooseCTATargets(pageType, pageName string, hubs []contentHub) (string, string) {
	ordered := make([]contentHub, 0, len(hubs))
	for _, h := range hubs {
		if areasExcludedFromCTA[h.Area] {
			continue
		}
		if pageName != "" && h.Name == pageName { // don't point a page's hero at itself
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

func sectionComponentFunction(section map[string]interface{}) string {
	if comp, ok := section["component"].(map[string]interface{}); ok {
		if fn := stringOrEmpty(comp["function"]); fn != "" {
			return fn
		}
		if nm := stringOrEmpty(comp["name"]); nm != "" {
			return nm
		}
	}
	// Fall back to the section name (often equals the component function).
	return stringOrEmpty(section["name"])
}

func sectionResolvedData(section map[string]interface{}) map[string]interface{} {
	if rd, ok := section["resolved_data"].(map[string]interface{}); ok && rd != nil {
		return rd
	}
	return map[string]interface{}{}
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
