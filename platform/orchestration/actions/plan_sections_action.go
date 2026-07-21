// FILE: platform/orchestration/actions/plan_sections_action.go
//
// PlanSectionsAction reads a page's section list, loads each component's
// input_schema (v2 format), resolves data sources, and determines which
// sections can be generated vs which need human input vs which should skip.
//
// This sits between load_page_record and spawn_content_writer in the
// page-build-handler workflow. The content writer only receives sections
// that have all required data available.
//
// Registration:
//   "plan_sections": {
//       Handler:     PlanSectionsAction,
//       Category:    "site",
//       Description: "Resolve section data requirements and triage readiness",
//       IsLocal:     true,
//   },
//
// Workflow config:
//   "plan_sections": {
//       "action": "plan_sections",
//       "config": {
//           "site_id": "site_record.site_id",
//           "sections": "page_record.sections",
//           "page_name": "page_record.name"
//       },
//       "next_step": "check_has_ready_sections",
//       "output_field": "section_plan"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/orchestration/imageryplan"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

var PlanSectionsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"sections", "page_name", "pipeline", "work_item_id", "site_type", "page_type"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("plan_sections", PlanSectionsInputSpec)
}

// ============================================================================
// Source resolution
// ============================================================================

// sourceResolver holds cached lookups for a single invocation
type sourceResolver struct {
	siteID       uuid.UUID
	db           *sql.DB
	logger       *zap.Logger
	specs        map[string]map[string]interface{} // aspect → data
	pages        map[string]string                 // page name → url
	assets       map[string]string                 // asset type → url
	pageName     string                            // page being planned; scopes per-page asset resolution (hero)
	specsLoaded  bool
	pagesLoaded  bool
	assetsLoaded bool
}

// NOTE (signature change): newSourceResolver now takes pageName so site_assets
// resolution can be page-aware (this page's hero rather than a single
// site-wide hero_url). There is one caller (PlanSectionsAction); it passes the
// page_name it already has. An empty pageName degrades safely — the per-page
// hero lookup is skipped and resolution falls back to content_data.
func newSourceResolver(siteID uuid.UUID, db *sql.DB, logger *zap.Logger, pageName string) *sourceResolver {
	return &sourceResolver{
		siteID:   siteID,
		db:       db,
		logger:   logger,
		pageName: pageName,
		specs:    make(map[string]map[string]interface{}),
		pages:    make(map[string]string),
		assets:   make(map[string]string),
	}
}

// loadSpecs loads all current site_specs for this site (once)
func (r *sourceResolver) ensureSpecs(ctx context.Context) {
	if r.specsLoaded {
		return
	}
	r.specsLoaded = true

	rows, err := r.db.QueryContext(ctx, `
		SELECT aspect, data FROM site_specs
		WHERE site_id = $1 AND is_current = true
	`, r.siteID)
	if err != nil {
		r.logger.Warn("plan_sections: failed to load site_specs", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var aspect string
		var dataJSON []byte
		if err := rows.Scan(&aspect, &dataJSON); err != nil {
			continue
		}
		var data map[string]interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			continue
		}
		r.specs[aspect] = data
	}

	r.logger.Info("plan_sections: loaded site_specs",
		zap.Int("aspect_count", len(r.specs)))
}

// loadPages loads all active pages for URL resolution (once)
func (r *sourceResolver) ensurePages(ctx context.Context) {
	if r.pagesLoaded {
		return
	}
	r.pagesLoaded = true

	rows, err := r.db.QueryContext(ctx, `
		SELECT name, url FROM pages
		WHERE site_id = $1 AND status = 'active'
	`, r.siteID)
	if err != nil {
		r.logger.Warn("plan_sections: failed to load pages", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var name, url string
		if err := rows.Scan(&name, &url); err != nil {
			continue
		}
		r.pages[name] = url
	}
}

// loadAssets resolves the asset URLs this page's sections may reference, once.
//
// CHANGE: previously this mapped a single site-wide content_data["hero_url"]
// to assets["hero"], so every page shared one hero (and StoreAssetAction
// overwrites hero_url per generation — last-write-wins). It now resolves the
// PAGE'S hero from the current plan's imagery rows joined to the deployed
// asset: site_plan_imagery.key is the asset_key, assets.url is the web path.
// So site_assets.hero on the index page resolves to hero-home.jpg, on
// games-index to hero-games.jpg, etc. The site logo (scope='site',
// kind='logo') resolves the same way. content_data remains a fallback for
// legacy/adopted sites with no plan imagery rows, or assets not yet active.
func (r *sourceResolver) ensureAssets(ctx context.Context) {
	if r.assetsLoaded {
		return
	}
	r.assetsLoaded = true

	// Per-page hero: this page's hero asset from the current plan, joined to
	// the deployed asset row. Skipped when pageName is empty (degrades to the
	// content_data fallback below).
	if r.pageName != "" {
		var assetKey, purpose string
		err := r.db.QueryRowContext(ctx, `
			SELECT a.asset_key, a.purpose
			  FROM site_plan_imagery spi
			  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
			  JOIN assets a ON a.site_id = sp.site_id
			               AND a.asset_key = spi.key
			               AND a.status = 'active'
			 WHERE sp.site_id = $1
			   AND spi.scope = 'page'
			   AND spi.scope_ref = $2
			   AND spi.kind = 'hero'
			 ORDER BY spi.ordering
			 LIMIT 1
		`, r.siteID, r.pageName).Scan(&assetKey, &purpose)
		switch {
		case err == nil && assetKey != "":
			// Resolve to the deployed git path, NOT assets.url (a presigned S3
			// URL that expires and is per-generation).
			r.assets["hero"] = storage.DeployedWebPath(assetKey, purpose)
		case err != nil && err != sql.ErrNoRows:
			r.logger.Warn("plan_sections: per-page hero lookup failed",
				zap.String("page", r.pageName), zap.Error(err))
		}
	}

	// Lane B content hero (Phase I3, D13): a per-article image generated from
	// the article's own content, stored under the literal ContentHeroKey
	// convention with no plan row. The planner's page hero (above) always
	// wins; the site brand hero (below) stays the last resort. This is what
	// makes the article page show the same image family as its listing card.
	if _, ok := r.assets["hero"]; !ok && r.pageName != "" {
		var assetKey, purpose string
		err := r.db.QueryRowContext(ctx, `
			SELECT a.asset_key, a.purpose
			  FROM assets a
			 WHERE a.site_id = $1
			   AND a.asset_key = $2
			   AND a.status = 'active'
			 LIMIT 1
		`, r.siteID, imageryplan.ContentHeroKey(r.pageName)).Scan(&assetKey, &purpose)
		switch {
		case err == nil && assetKey != "":
			r.assets["hero"] = storage.DeployedWebPath(assetKey, purpose)
		case err != nil && err != sql.ErrNoRows:
			r.logger.Warn("plan_sections: content hero lookup failed",
				zap.String("page", r.pageName), zap.Error(err))
		}
	}

	// Site-scope brand hero: fallback when the page has no hero of its own,
	// so image-role-aliased fields still resolve to something brand-consistent
	// rather than nothing. Page-scope (above) always wins.
	if _, ok := r.assets["hero"]; !ok {
		var assetKey, purpose string
		err := r.db.QueryRowContext(ctx, `
			SELECT a.asset_key, a.purpose
			  FROM site_plan_imagery spi
			  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
			  JOIN assets a ON a.site_id = sp.site_id
			               AND a.asset_key = spi.key
			               AND a.status = 'active'
			 WHERE sp.site_id = $1
			   AND spi.scope = 'site'
			   AND spi.kind = 'hero'
			 ORDER BY spi.ordering
			 LIMIT 1
		`, r.siteID).Scan(&assetKey, &purpose)
		switch {
		case err == nil && assetKey != "":
			r.assets["hero"] = storage.DeployedWebPath(assetKey, purpose)
		case err != nil && err != sql.ErrNoRows:
			r.logger.Warn("plan_sections: site-scope hero lookup failed", zap.Error(err))
		}
	}

	// Per-page section imagery: illustrations / icons / infographics requested at
	// section scope for this page (scope_ref = "<page>:<ordinal>"), joined to the
	// deployed asset row. Mapped by KEY (per-key schema paths, e.g. icon sets) and
	// aliased by KIND first-wins (generic paths like site_assets.illustration),
	// mirroring the hero mapping above. Skipped when pageName is empty.
	if r.pageName != "" {
		rows, err := r.db.QueryContext(ctx, `
			SELECT spi.kind, a.asset_key, a.purpose
			  FROM site_plan_imagery spi
			  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
			  JOIN assets a ON a.site_id = sp.site_id
			               AND a.asset_key = spi.key
			               AND a.status = 'active'
			 WHERE sp.site_id = $1
			   AND spi.scope = 'section'
			   AND spi.scope_ref LIKE $2 || ':%'
			   AND spi.kind IN ('illustration', 'icon', 'infographic')
			 ORDER BY spi.kind, spi.ordering
		`, r.siteID, r.pageName)
		if err != nil {
			r.logger.Warn("plan_sections: section imagery lookup failed",
				zap.String("page", r.pageName), zap.Error(err))
		} else {
			defer rows.Close()
			for rows.Next() {
				var kind, assetKey, purpose string
				if err := rows.Scan(&kind, &assetKey, &purpose); err != nil {
					continue
				}
				if assetKey == "" {
					continue
				}
				url := storage.DeployedWebPath(assetKey, purpose)
				r.assets[assetKey] = url
				if _, exists := r.assets[kind]; !exists {
					r.assets[kind] = url
				}
			}
			if err := rows.Err(); err != nil {
				r.logger.Warn("plan_sections: section imagery rows error",
					zap.String("page", r.pageName), zap.Error(err))
			}
		}
	}

	// Site logo.
	var logoKey, logoPurpose string
	err := r.db.QueryRowContext(ctx, `
		SELECT a.asset_key, a.purpose
		  FROM site_plan_imagery spi
		  JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		  JOIN assets a ON a.site_id = sp.site_id
		               AND a.asset_key = spi.key
		               AND a.status = 'active'
		 WHERE sp.site_id = $1
		   AND spi.scope = 'site'
		   AND spi.kind = 'logo'
		 ORDER BY spi.ordering
		 LIMIT 1
	`, r.siteID).Scan(&logoKey, &logoPurpose)
	switch {
	case err == nil && logoKey != "":
		r.assets["logo"] = storage.DeployedWebPath(logoKey, logoPurpose)
	case err != nil && err != sql.ErrNoRows:
		r.logger.Warn("plan_sections: logo lookup failed", zap.Error(err))
	}

	// Fallback: content_data for anything not resolved above (legacy/adopted
	// sites without plan imagery, or assets not yet active). Gap-fill only —
	// the per-plan values above take precedence.
	var contentDataJSON []byte
	if err := r.db.QueryRowContext(ctx, `
		SELECT content_data FROM sites WHERE id = $1
	`, r.siteID).Scan(&contentDataJSON); err != nil {
		return
	}
	var contentData map[string]interface{}
	if err := json.Unmarshal(contentDataJSON, &contentData); err != nil {
		return
	}
	if _, ok := r.assets["hero"]; !ok {
		if heroURL, ok := contentData["hero_url"].(string); ok && heroURL != "" {
			r.assets["hero"] = heroURL
		}
	}
	if _, ok := r.assets["logo"]; !ok {
		if logoURL, ok := contentData["logo_url"].(string); ok && logoURL != "" {
			r.assets["logo"] = logoURL
		}
	}
}

// sectionDescription returns the purpose/description for a section from the
// site_plan spec. Falls back to page purpose if no section-level description exists.
// Uses already-loaded specs — no extra DB query.
func (r *sourceResolver) sectionDescription(pageName, sectionType string) string {
	plan, ok := r.specs["site_plan"]
	if !ok {
		return ""
	}

	pages, ok := plan["pages"].([]interface{})
	if !ok {
		return ""
	}

	for _, pageRaw := range pages {
		page, ok := pageRaw.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := page["name"].(string)
		if name != pageName {
			continue
		}

		// Check section_descriptions map (if planner provides it)
		if descs, ok := page["section_descriptions"].(map[string]interface{}); ok {
			if desc, ok := descs[sectionType].(string); ok && desc != "" {
				return desc
			}
		}

		// Check section_types array for objects with description
		if sectionTypes, ok := page["section_types"].([]interface{}); ok {
			for _, stRaw := range sectionTypes {
				if st, ok := stRaw.(map[string]interface{}); ok {
					if stName, _ := st["name"].(string); stName == sectionType {
						if desc, _ := st["description"].(string); desc != "" {
							return desc
						}
					}
				}
			}
		}

		// Fall back to page purpose
		if purpose, ok := page["purpose"].(string); ok && purpose != "" {
			return fmt.Sprintf("Section '%s' on page '%s' (purpose: %s)", sectionType, pageName, purpose)
		}
	}

	return ""
}

// resolve checks if a data source has a value available
// Returns: value (if found), found (bool)
func (r *sourceResolver) resolve(ctx context.Context, source string) (interface{}, bool) {
	if source == "" || source == "llm" || source == "renderer" || source == "static" {
		// These sources don't need resolution — they're generated at render time
		return nil, true
	}

	parts := strings.SplitN(source, ".", 2)
	if len(parts) < 2 {
		return nil, false
	}

	prefix := parts[0]
	path := parts[1]

	switch prefix {
	case "renderer", "static":
		// These are injected at render time — always considered available
		return nil, true

	case "site_specs":
		r.ensureSpecs(ctx)
		return r.resolveSpecPath(path)

	case "site_assets":
		r.ensureAssets(ctx)
		if url, ok := r.assets[path]; ok {
			return url, true
		}
		// Literal key missed — try the image-role alias. Preset/imported
		// components name their image fields freely (site_assets.background,
		// site_assets.product_screenshot, ...) but the pipeline generates
		// per-page heroes; without the alias those fields resolve to nothing
		// and templates render src="". Exact keys above always win, so a
		// future dedicated asset under the literal key takes precedence.
		if role, ok := imageryplan.ImageRoleForPath(path); ok {
			if url, ok := r.assets[role]; ok {
				r.logger.Info("plan_sections: site_assets path resolved via image-role alias",
					zap.String("path", path),
					zap.String("role", role),
					zap.String("page", r.pageName))
				return url, true
			}
		}
		return nil, false

	case "pages":
		r.ensurePages(ctx)
		if url, ok := r.pages[path]; ok {
			return url, true
		}
		// No such page — do NOT fabricate a URL. Returning (nil, false) lets
		// the field's on_missing govern (skip_field drops the field; gated
		// templates then render no button). Fabricating "/<path>.html" here
		// was the phantom-link generator (/contact.html, /services.html on
		// every hero/CTA site-wide).
		r.logger.Info("plan_sections: pages source not found; deferring to on_missing",
			zap.String("page_ref", path),
			zap.String("site_id", r.siteID.String()))
		return nil, false

	case "config":
		r.ensureSpecs(ctx)
		return r.resolveConfigPath(path)

	case "query":
		// Query sources are resolved by the field-loop in planSection via
		// the queryresolve package, BEFORE this method is called. This case
		// is defensive — if a future caller invokes resolveSource directly
		// on a query.* source (instead of going through the field loop),
		// returning (nil, true) keeps the system stable rather than treating
		// it as an unknown source. Real callers should not hit this branch.
		return nil, true

	default:
		return nil, false
	}
}

// resolveSpecPath navigates site_specs: "identity.team" → specs["identity"]["team"]
func (r *sourceResolver) resolveSpecPath(path string) (interface{}, bool) {
	parts := strings.SplitN(path, ".", 2)
	aspect := parts[0]

	specData, ok := r.specs[aspect]
	if !ok {
		return nil, false
	}

	if len(parts) == 1 {
		return specData, true
	}

	// Navigate deeper: "identity.team" → specs["identity"]["team"]
	return navigateMap(specData, parts[1])
}

// resolveConfigPath navigates site content_data config
func (r *sourceResolver) resolveConfigPath(path string) (interface{}, bool) {
	// Config values live in site_specs under various aspects
	// Search across relevant aspects
	for _, aspect := range []string{"site_config", "identity", "design_intent"} {
		if specData, ok := r.specs[aspect]; ok {
			if val, found := navigateMap(specData, path); found {
				return val, true
			}
		}
	}
	return nil, false
}

func navigateMap(data map[string]interface{}, dotPath string) (interface{}, bool) {
	parts := strings.Split(dotPath, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, exists := v[part]
			if !exists {
				return nil, false
			}
			current = val
		default:
			return nil, false
		}
	}

	// Check if the value is actually populated (not empty string, not empty array)
	switch v := current.(type) {
	case string:
		if v == "" {
			return nil, false
		}
	case []interface{}:
		if len(v) == 0 {
			return nil, false
		}
	case nil:
		return nil, false
	}

	return current, true
}

// ============================================================================
// Section planning result
// ============================================================================

// llmFieldSpec carries per-field metadata for the Step 3 targeted-prompt
// path on page-content-writer. Each entry corresponds to one field whose
// `source` is "llm" in the component's input_schema. The page-content-writer
// prompt template iterates this list instead of dumping the full schema —
// the LLM is asked for exactly the fields it should write, with their
// types and intent, and never given the opportunity to fabricate
// query-resolved data (items, urls, page lists) that the system handles
// elsewhere.
type llmFieldSpec struct {
	Name        string      `json:"name"`
	Type        string      `json:"type,omitempty"` // text | url | image | rich_text | …
	Required    bool        `json:"required,omitempty"`
	Description string      `json:"description,omitempty"` // sourced from input_schema field's `llm_guidance` key
	OnMissing   string      `json:"on_missing,omitempty"`  // skip_field | use_fallback | error
	Fallback    interface{} `json:"fallback,omitempty"`    // value used when on_missing=use_fallback
	// ItemFields lists the field names each element of an array-typed field
	// must contain, from the schema field's `items` (or `item_schema`) map.
	// Empty for non-array fields. Surfaced to the LLM (via the prompt) and to
	// the render-time reconciler so the model emits the exact keys the
	// component template reads, instead of guessing item field names (e.g.
	// title/body) that render empty against a template reading name/description.
	ItemFields []string `json:"item_fields,omitempty"`
}

type sectionPlanItem struct {
	Name         string                 `json:"name"`
	ComponentID  string                 `json:"component_id"`
	Function     string                 `json:"function"`
	Status       string                 `json:"status"` // "ready", "deferred", "skipped"
	ResolvedData map[string]interface{} `json:"resolved_data,omitempty"`
	LLMFields    []string               `json:"llm_fields,omitempty"`
	// LLMFieldSpecs is the richer counterpart to LLMFields: each spec carries
	// the field's name plus the metadata the targeted-prompt template needs
	// (type, required flag, description, on_missing handling, fallback value).
	// LLMFields stays as a fast lookup of "which fields are LLM-written";
	// LLMFieldSpecs is what page-content-writer's prompt iterates.
	LLMFieldSpecs []llmFieldSpec `json:"llm_field_specs,omitempty"`
	Missing       []missingField `json:"missing,omitempty"`
	Reason        string         `json:"reason,omitempty"`
	// Component carries the full per-section component data as returned by
	// the shared loadSectionComponents helper. Populated when a component
	// was found (Paths 1 and 2). Nil for paths where no component was
	// resolved (Path 3: not_found / selector_unavailable). Downstream
	// consumers — page-content-writer in Step 3 — read input_schema,
	// html_template, render_mode, description, category, content_brief
	// etc. from here instead of re-loading via load_page_section_components.
	Component map[string]interface{} `json:"component,omitempty"`
}

type missingField struct {
	Field     string                 `json:"field"`
	Source    string                 `json:"source"`
	OnMissing string                 `json:"on_missing"`
	Reason    string                 `json:"reason"`
	Type      string                 `json:"type,omitempty"`      // from input_schema field type
	Items     map[string]interface{} `json:"items,omitempty"`     // from input_schema items (array element schema)
	MinItems  int                    `json:"min_items,omitempty"` // from input_schema min_items
}

// ============================================================================
// Main action
// ============================================================================

func PlanSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "plan_sections"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		PlanSectionsInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	pageName := inputs.Get("page_name")
	workItemID := inputs.Get("work_item_id")

	// site_type and page_type for the component selector fallback path.
	// If not provided (existing workflows), the selector still works —
	// it just scores without site/page type relevance bonuses.
	siteType := inputs.Get("site_type")
	pageType := inputs.Get("page_type")
	if pageType == "" {
		pageType = pageName // fall back to page name as page type
	}

	// Parse sections list
	sectionsRaw := inputs.GetRaw("sections")
	var sectionNames []string

	switch v := sectionsRaw.(type) {
	case []interface{}:
		for _, s := range v {
			if name, ok := s.(string); ok {
				sectionNames = append(sectionNames, name)
			}
		}
	case []string:
		sectionNames = v
	case string:
		// Try JSON parse
		if err := json.Unmarshal([]byte(v), &sectionNames); err != nil {
			return nil, fmt.Errorf("failed to parse sections: %w", err)
		}
	}

	// ── Filter out site-level component names ────────────────────────
	// The planner or adoption flow may include header/footer names in the
	// page sections list. These are site-level components handled by
	// InjectHeader/InjectFooter — if we process them here they end up as
	// page_components rows, causing duplicate headers/footers on assembly.
	sectionNames = filterSiteLevelSections(sectionNames, logger)

	if len(sectionNames) == 0 {
		return map[string]interface{}{
			"sections_ready":    []interface{}{},
			"sections_deferred": []interface{}{},
			"sections_skipped":  []interface{}{},
			"ready_count":       0,
			"reason":            "no sections to plan",
		}, nil
	}

	// Load component schemas for these sections
	components := loadComponentSchemas(ctx, params.DB, sectionNames, logger)

	// Create resolver
	resolver := newSourceResolver(siteID, params.DB, logger, pageName)

	// Pre-load specs so we can extract design_direction for needs_new_component items.
	// ensureSpecs is idempotent — later calls in planSection() won't re-query.
	resolver.ensureSpecs(ctx)

	// Extract design_direction from design_intent spec (if present).
	// Passed to needs_new_component items so the component-creator knows the visual style.
	designDirection := ""
	if di, ok := resolver.specs["design_intent"]; ok {
		if sd, ok := di["style_direction"].(string); ok && sd != "" {
			designDirection = sd
		}
	}

	// ── Load open data requests for reconciliation after planning ────
	// Used after the planning loop to:
	//   1. Close stale requests for sections that are now ready (component created, data arrived)
	//   2. Skip creating duplicate requests for sections that are still deferred
	openDataRequests := loadOpenSectionDataRequests(ctx, params.DB, siteID, pageName, logger)

	// Plan each section
	var ready []sectionPlanItem
	var deferred []sectionPlanItem
	var skipped []sectionPlanItem

	// Build selector context for the fallback path.
	// This is only used when a section name doesn't match a component function directly.
	selCtx := SelectorContext{
		SiteType: siteType,
		PageType: pageType,
		PageName: pageName,
	}

	for _, sectionName := range sectionNames {
		// Path 1: Direct function/name lookup (existing behaviour).
		// All current sites hit this path — their planners output function names.
		comp, ok := components[sectionName]
		if ok {
			item := planSection(ctx, sectionName, comp, resolver, logger)

			switch item.Status {
			case "ready":
				ready = append(ready, item)
			case "deferred":
				deferred = append(deferred, item)
			case "skipped":
				skipped = append(skipped, item)
			}
			continue
		}

		// Path 2: Section type selector.
		// The planner output a section_type (e.g. "provocation-card") rather than
		// a specific function name. The selector queries content_components by
		// section_type, scores candidates, and returns the best match.
		resolved, resolution := resolveSectionComponent(ctx, params.DB, sectionName, selCtx, logger)
		if resolved != nil {
			// Selector found a matching component. Its function flows through the
			// rest of the pipeline exactly as if the planner had specified it directly.
			item := planSection(ctx, resolved.Function, *resolved, resolver, logger)
			// Preserve the original section_type name — downstream logging and
			// the content writer use item.Name as the section identifier.
			item.Name = sectionName
			switch item.Status {
			case "ready":
				ready = append(ready, item)
			case "deferred":
				deferred = append(deferred, item)
			case "skipped":
				skipped = append(skipped, item)
			}
			continue
		}

		// Path 3: No component found anywhere.
		if resolution == "not_found" {
			logger.Info("plan_sections: no component for section_type, creating work item",
				zap.String("section_type", sectionName),
				zap.String("page", pageName))

			// Try to get a meaningful description from the site_plan spec
			// (resolver already has specs loaded — no extra DB query needed)
			description := resolver.sectionDescription(pageName, sectionName)
			if description == "" {
				description = fmt.Sprintf("Component for section type %q on page %q (%s site)", sectionName, pageName, siteType)
			}

			err := CreateNeedsNewComponentItem(
				ctx, params.DB, siteIDStr,
				sectionName, pageName, description,
				designDirection, // extracted from resolver.specs before the loop
				siteType, logger,
			)
			if err != nil {
				logger.Warn("plan_sections: failed to create needs_new_component work item",
					zap.String("section_type", sectionName),
					zap.Error(err))
			}

			deferred = append(deferred, sectionPlanItem{
				Name:   sectionName,
				Status: "deferred",
				Reason: fmt.Sprintf("no component for section_type %q — needs_new_component work item created", sectionName),
			})
		} else {
			// Selector error or unavailable — fall through to content writer (backward compat).
			// This keeps the same behaviour as before for edge cases where the DB query fails.
			logger.Warn("plan_sections: selector unavailable, passing section to content writer as-is",
				zap.String("section", sectionName),
				zap.String("resolution", resolution))
			ready = append(ready, sectionPlanItem{
				Name:   sectionName,
				Status: "ready",
				Reason: "selector unavailable — passing to content writer as-is",
			})
		}
	}

	// ── Reconcile open data requests with planning results ───────────
	// Close stale requests for sections that are now ready (component
	// created since the request was filed, data now available, etc.)
	if params.DB != nil && len(openDataRequests) > 0 {
		for _, section := range ready {
			if _, wasOpen := openDataRequests[section.Name]; wasOpen {
				closeResolvedDataRequest(ctx, params.DB, siteID, pageName, section.Name, logger)
			}
		}
	}

	// Create work items for deferred sections (skips those that already have open requests)
	if params.DB != nil && len(deferred) > 0 {
		// Filter out sections that already have open data requests — no duplicate items
		var newDeferred []sectionPlanItem
		for _, section := range deferred {
			if _, alreadyOpen := openDataRequests[section.Name]; !alreadyOpen {
				newDeferred = append(newDeferred, section)
			}
		}
		if len(newDeferred) > 0 {
			createDeferredItems(ctx, params.DB, siteID, pageName, workItemID, newDeferred, logger)
		}
	}

	// Build section names lists for the content writer
	readyNames := make([]string, len(ready))
	for i, s := range ready {
		readyNames[i] = s.Name
	}

	logger.Info("plan_sections: planning complete",
		zap.Int("ready", len(ready)),
		zap.Int("deferred", len(deferred)),
		zap.Int("skipped", len(skipped)),
		zap.String("page", pageName))

	return map[string]interface{}{
		"sections_ready":    ready,
		"sections_deferred": deferred,
		"sections_skipped":  skipped,
		"ready_names":       readyNames,
		"ready_count":       len(ready),
		"deferred_count":    len(deferred),
		"skipped_count":     len(skipped),
		"total_sections":    len(sectionNames),
	}, nil
}

// ============================================================================
// Load component schemas by section name
// ============================================================================

type componentInfo struct {
	ID          string
	Name        string
	Function    string
	InputSchema map[string]interface{}
	// Raw carries the full per-section component map produced by the
	// shared loadSectionComponents loader. Plan_sections attaches this
	// onto sectionPlanItem.Component so downstream consumers can read
	// html_template, render_mode, description, category, etc. without
	// re-loading from content_components. Step 3 swaps the page-content-
	// writer over to consume this directly.
	Raw map[string]interface{}
}

// loadComponentSchemas is a thin wrapper over the shared loadSectionComponents
// helper. It converts the helper's per-component maps into componentInfo
// records keyed by both name and function (the lookup pattern planSection
// expects), parses input_schema JSON for the field-resolution walk, and
// applies the template-truncation guard that the previous in-line SQL did.
//
// Note: plan_sections doesn't have a pageID at this point in the workflow,
// so brief enrichment is skipped here. Briefs apply at content-write time
// (page-content-writer's load path) where pageID is known. Step 3 may move
// brief loading into the section_plan so plan_sections becomes the single
// source for all per-section content.
func loadComponentSchemas(ctx context.Context, db *sql.DB, sectionNames []string, logger *zap.Logger) map[string]componentInfo {
	result := make(map[string]componentInfo)

	// activeOnly=true preserves the historical is_active=true filter that the
	// inline SQL had. Inactive components stay out of plan_sections so they
	// flow to Path 2 (selector) and may be replaced by a current alternative.
	components := loadSectionComponents(ctx, db, sectionNames, "", true, logger)

	for _, comp := range components {
		// Stubs from the helper have no component_id — drop them so plan_sections
		// falls through to Path 2 (selector) for those names.
		if _, hasID := comp["component_id"]; !hasID {
			continue
		}

		// Template truncation guard: components with HTML content but no
		// closing </section> tag are treated as broken and skipped so they
		// flow to Path 3 (needs_new_component work item) instead of
		// rendering broken markup. Empty/very-short templates are NOT
		// dropped here — they may be stubs that legitimately have no body.
		//
		// component_level='tool' templates get their own check: a tool is
		// self-contained HTML, not a <section> wrapper, so the '</section>'
		// marker is the wrong truncation signal in BOTH directions there —
		// it dropped healthy tools that end '</script>' (which is how a
		// durable tool fix could never re-render, bugs_open/024) and passed
		// truncated ones that happen to contain '</section>' upstream of
		// the cut.
		htmlTpl, _ := comp["html_template"].(string)
		level, _ := comp["component_level"].(string)
		if !componentTemplateValid(htmlTpl, level) {
			name, _ := comp["name"].(string)
			function, _ := comp["function"].(string)
			logger.Warn("plan_sections: component template truncated, skipping",
				zap.String("function", function),
				zap.String("name", name),
				zap.String("component_level", level))
			continue
		}

		var ci componentInfo
		if id, ok := comp["component_id"].(string); ok {
			ci.ID = id
		}
		if name, ok := comp["name"].(string); ok {
			ci.Name = name
		}
		if fn, ok := comp["function"].(string); ok {
			ci.Function = fn
		}
		if schemaStr, ok := comp["input_schema"].(string); ok && schemaStr != "" {
			if err := json.Unmarshal([]byte(schemaStr), &ci.InputSchema); err != nil {
				logger.Warn("plan_sections: failed to parse input_schema",
					zap.String("component", ci.Name),
					zap.Error(err))
			}
		}
		if ci.InputSchema == nil {
			ci.InputSchema = make(map[string]interface{})
		}
		ci.Raw = comp

		// Index by both name and function for fast lookup in the section loop.
		if ci.Name != "" {
			result[ci.Name] = ci
		}
		if ci.Function != "" && ci.Function != ci.Name {
			result[ci.Function] = ci
		}
	}

	// Alias each raw requested name to its resolved component when the plan asked
	// in snake_case/CamelCase but the component is stored kebab-case
	// (bugs_open/041). loadSectionComponents now resolves such a section, but this
	// map is keyed by the STORED name/function; callers look it up by the
	// REQUESTED name (plan_sections' section loop, rerender's slot_name). Without
	// the alias a "call_to_action" request misses the "call-to-action" entry and
	// falls through to a spurious needs_new_component.
	aliasNormalisedSectionKeys(result, sectionNames)

	return result
}

// aliasNormalisedSectionKeys adds, for each requested section name that is not
// already a key, an alias to the entry stored under its kebab-normalised form.
// Strict superset — it only ADDS keys and never rebinds an existing one, so a
// component whose own name is snake_case (keyed by its raw name) is untouched.
// See /bugs_open/041 (section-lookup-never-normalises).
func aliasNormalisedSectionKeys(result map[string]componentInfo, sectionNames []string) {
	for _, name := range sectionNames {
		if _, ok := result[name]; ok {
			continue
		}
		norm := NormalizeComponentFunction(name)
		if norm == name {
			continue
		}
		if ci, ok := result[norm]; ok {
			result[name] = ci
		}
	}
}

// sectionTemplateValid mirrors the original SQL CASE used by loadComponentSchemas:
//
//	WHEN html_template LIKE '%</section>%' THEN true
//	WHEN html_template IS NULL            THEN true
//	WHEN LENGTH(html_template) < 100      THEN true
//	ELSE false
//
// The only "invalid" case is a long template with no closing </section> tag —
// the signature of a truncated LLM generation. Empty/short templates are
// allowed through because they may be intentional stubs.
func sectionTemplateValid(htmlTemplate string) bool {
	if htmlTemplate == "" {
		return true
	}
	if len(htmlTemplate) < 100 {
		return true
	}
	return strings.Contains(htmlTemplate, "</section>")
}

// componentTemplateValid is THE truncation gate for a loaded component, and the
// only one either loader should call.
//
// It exists because there are TWO call sites making this identical judgement —
// loadComponentSchemas (the bulk loader) and loadSingleComponentSchema (the
// by-function loader) — and the first fix for bugs_open/024 patched only the
// bulk one. The council's bug_historian seat predicted the second call site from
// this platform's documented history of the same filter existing twice, and it
// was right: `loadSingleComponentSchema` was still rejecting self-contained tool
// templates on the '</section>' marker and returning nil, silently.
//
// Both loaders now share this predicate so the two cannot drift again. A
// component that is dropped here is invisible downstream — no error, no work
// item — which is what made the original defect cost three fix cycles.
func componentTemplateValid(htmlTemplate, componentLevel string) bool {
	if componentLevel == "tool" {
		return toolTemplateValid(htmlTemplate)
	}
	return sectionTemplateValid(htmlTemplate)
}

// toolTemplateValid is the truncation guard for component_level='tool'
// templates. A tool is self-contained HTML, not a <section> wrapper, so
// sectionTemplateValid's '</section>' marker misclassifies tools in both
// directions: healthy tools ending '</script>' read as truncated (and were
// silently dropped from the schemas map, so a durable tool fix could never
// re-render — bugs_open/024), while genuinely cut templates that contain
// '</section>' upstream of the cut read as whole.
//
// Reuses the component write guard's absolute structural signals instead:
// every paired tag balanced, and the template ends on a closed tag.
//
// Calibrated against all 27 active tool components on 2026-07-20: the 19
// structurally whole templates pass; the 8 truncated rows all fail (each cut
// mid-JavaScript by a pre-guard truncation write, bugs_open/012's class —
// four of which contain '</section>' and so pass sectionTemplateValid today).
// Rejecting those here is load-bearing: it keeps the re-render loop on the
// carry-stored-HTML path for a damaged template instead of deploying broken
// markup from it.
func toolTemplateValid(htmlTemplate string) bool {
	if htmlTemplate == "" {
		return true
	}
	if len(htmlTemplate) < 100 {
		return true
	}
	folded := strings.ToLower(htmlTemplate)
	for _, pair := range balancedPairs {
		if strings.Count(folded, pair.open) > strings.Count(folded, pair.close) {
			return false
		}
	}
	return endsCleanly(htmlTemplate)
}

// ============================================================================
// Section type resolution via component selector
// ============================================================================

// resolveSectionComponent attempts to find a component for a section name
// that didn't match any function directly. It queries the component selector
// by section_type, which scores candidates against the site/page context.
//
// Returns the resolved componentInfo and a resolution status string:
//   - "selected": selector found a match
//   - "not_found": no components with this section_type exist
//   - "selector_error": DB query failed
func resolveSectionComponent(
	ctx context.Context,
	db *sql.DB,
	sectionName string,
	selCtx SelectorContext,
	logger *zap.Logger,
) (*componentInfo, string) {

	candidate, err := SelectComponentByType(ctx, db, sectionName, selCtx, logger)
	if err != nil {
		logger.Warn("plan_sections: selector query failed",
			zap.String("section", sectionName),
			zap.Error(err))
		return nil, "selector_error"
	}

	if candidate == nil {
		return nil, "not_found"
	}

	// Selector found a match — load the full component info including input_schema.
	// The selector only returns metadata; we need the schema for field resolution.
	comp := loadSingleComponentSchema(ctx, db, candidate.Function, logger)
	if comp == nil {
		logger.Warn("plan_sections: selector matched but component load failed",
			zap.String("section_type", sectionName),
			zap.String("function", candidate.Function))
		return nil, "selector_error"
	}

	// Increment usage count for the selected component
	IncrementUsageCount(ctx, db, candidate.ID, logger)

	logger.Info("plan_sections: resolved via section_type selector",
		zap.String("section_type", sectionName),
		zap.String("resolved_function", candidate.Function),
		zap.Float64("score", candidate.Score))

	return comp, "selected"
}

// loadSingleComponentSchema loads one component's schema by function name.
// Thin wrapper over the shared loadSectionComponents helper. Rejects components
// with truncated templates (missing </section>) so they don't render broken HTML.
// Always uses activeOnly=true to match the original is_active filter.
func loadSingleComponentSchema(ctx context.Context, db *sql.DB, function string, logger *zap.Logger) *componentInfo {
	components := loadSectionComponents(ctx, db, []string{function}, "", true, logger)

	for _, raw := range components {
		// Stubs from the helper (no component_id) mean the function wasn't found.
		if _, hasID := raw["component_id"]; !hasID {
			continue
		}

		htmlTpl, _ := raw["html_template"].(string)
		level, _ := raw["component_level"].(string)
		if !componentTemplateValid(htmlTpl, level) {
			logger.Warn("loadSingleComponentSchema: template truncated, rejecting",
				zap.String("function", function),
				zap.String("component_level", level))
			return nil
		}

		var ci componentInfo
		if id, ok := raw["component_id"].(string); ok {
			ci.ID = id
		}
		if name, ok := raw["name"].(string); ok {
			ci.Name = name
		}
		if fn, ok := raw["function"].(string); ok {
			ci.Function = fn
		}
		if schemaStr, ok := raw["input_schema"].(string); ok && schemaStr != "" {
			if err := json.Unmarshal([]byte(schemaStr), &ci.InputSchema); err != nil {
				logger.Warn("loadSingleComponentSchema: failed to parse input_schema",
					zap.String("function", function),
					zap.Error(err))
			}
		}
		if ci.InputSchema == nil {
			ci.InputSchema = make(map[string]interface{})
		}
		ci.Raw = raw
		return &ci
	}

	logger.Warn("loadSingleComponentSchema: function not found",
		zap.String("function", function))
	return nil
}

// ============================================================================
// Plan a single section
// ============================================================================

func planSection(ctx context.Context, sectionName string, comp componentInfo, resolver *sourceResolver, logger *zap.Logger) sectionPlanItem {
	item := sectionPlanItem{
		Name:        sectionName,
		ComponentID: comp.ID,
		Function:    comp.Function,
		Status:      "ready",
		// Attach the full component data so downstream consumers (Step 3:
		// page-content-writer) can read input_schema, html_template,
		// render_mode, description, category, content_brief etc. without
		// re-loading via load_page_section_components. Nil for sections
		// where no component was resolved.
		Component: comp.Raw,
	}

	// Get fields from schema. Read via schemaContentFields so a component in the
	// legacy JSON-Schema dialect (`properties`+`required[]`, no `fields`) has its
	// fields planned for — before bugs_open/026 a missed dialect fell through to
	// "all fields from LLM" with no field specs, so a required field the writer
	// was never told about (the news-listing headline) was never generated.
	fieldsRaw, ok := schemaContentFields(comp.InputSchema)
	if !ok || len(fieldsRaw) == 0 {
		// A self-contained TOOL component legitimately has an empty input_schema:
		// its HTML renders entirely from its own template, with no LLM-authored
		// content fields to supply. Exempt it by the SAME explicit
		// component_level='tool' marker the rerender escalation guard uses
		// (isSelfContainedSection), NEVER by the name heuristic below. Without
		// this, a future tool whose Function name happens to contain
		// "content"/"body"/"article"/… would be marked `deferred` here — carried
		// unchanged, so a durable template fix is computed and silently discarded
		// — the identical end-state as bugs_open/024, reached one function away by
		// a different route (bugs_open/044). Two call sites of the "is this
		// emptiness legitimate?" judgement, now one shared predicate so they
		// cannot drift apart again.
		if isSelfContainedSection(comp) {
			item.Reason = "self-contained tool component — renders from its own template, no content fields"
			return item
		}

		// No v2 schema — component has no declared content fields.
		// Check if the template has actual HTML structure. If it's CSS-only
		// or truncated, it was likely created by a broken component-creator
		// run and will produce empty/broken output.
		if comp.Function != "" {
			funcLower := strings.ToLower(comp.Function)
			// Components that typically need LLM content should not have empty schemas
			needsContent := strings.Contains(funcLower, "article") ||
				strings.Contains(funcLower, "content") ||
				strings.Contains(funcLower, "body") ||
				strings.Contains(funcLower, "text") ||
				strings.Contains(funcLower, "blog")
			if needsContent {
				item.Status = "deferred"
				item.Reason = "component has empty input_schema — needs regeneration with content fields"
				logger.Warn("plan_sections: content component has empty schema, deferring",
					zap.String("function", comp.Function),
					zap.String("section", sectionName))
				return item
			}
		}
		// Non-content components with empty schema (e.g. decorative sections,
		// separators) — treat as fully LLM-generated for backward compat
		item.Reason = "no field schema — all fields from LLM"
		return item
	}

	resolvedData := make(map[string]interface{})
	var llmFields []string
	var llmFieldSpecs []llmFieldSpec
	var missingFields []missingField
	shouldSkip := false
	shouldDefer := false

	for fieldName, fieldDefRaw := range fieldsRaw {
		fieldDef, ok := fieldDefRaw.(map[string]interface{})
		if !ok {
			continue
		}

		source, _ := fieldDef["source"].(string)
		required, _ := fieldDef["required"].(bool)
		onMissing, _ := fieldDef["on_missing"].(string)
		if onMissing == "" {
			onMissing = "skip_field"
		}
		fallback := fieldDef["fallback"]
		missingReason, _ := fieldDef["missing_reason"].(string)

		// Extract type info for enriched HITL work items
		fieldType, _ := fieldDef["type"].(string)
		fieldItems, _ := fieldDef["items"].(map[string]interface{})
		fieldMinItems := 0
		if mi, ok := fieldDef["min_items"].(float64); ok {
			fieldMinItems = int(mi)
		}

		// LLM-generated fields — always available
		if source == "llm" {
			llmFields = append(llmFields, fieldName)
			llmFieldSpecs = append(llmFieldSpecs, llmFieldSpec{
				Name:        fieldName,
				Type:        fieldType,
				Required:    required,
				Description: stringOrEmpty(fieldDef["llm_guidance"]),
				OnMissing:   onMissing,
				Fallback:    fallback,
				ItemFields:  extractArrayItemFields(fieldDef),
			})
			continue
		}

		// Query.* fields — resolve via the queryresolve package.
		// The query resolver runs SQL against the database and returns
		// data shaped for the field. For array-typed fields (lists, grids,
		// directories) the result is []map[string]interface{} that the
		// downstream content writer / template renderer iterates over.
		//
		// Failure handling:
		//   - Unknown query name → log warning, fall through to fallback/skip
		//   - DB error → log warning, fall through to fallback/skip
		//   - Empty result → put empty slice in resolvedData (the component's
		//     html_template should handle empty lists; on_missing applies if
		//     the field is required and the schema treats empty as missing)
		if strings.HasPrefix(source, "query.") {
			queryName := strings.TrimPrefix(source, "query.")

			// Optional limit from the field schema (max items for the list).
			itemLimit := 0
			if l, ok := fieldDef["limit"].(float64); ok {
				itemLimit = int(l)
			}

			req := queryresolve.QueryRequest{
				Name:   queryName,
				SiteID: resolver.siteID,
				Limit:  itemLimit,
			}
			value, qerr := queryresolve.Resolve(ctx, resolver.db, req, resolver.logger)
			if qerr != nil {
				resolver.logger.Warn("plan_sections: query resolution failed",
					zap.String("field", fieldName),
					zap.String("source", source),
					zap.Error(qerr))
				// Fall through to fallback handling below
			} else if value != nil {
				resolvedData[fieldName] = value
				continue
			}

			// Resolution failed or returned nil — apply fallback if any,
			// otherwise leave the field unresolved (the template will see
			// nothing for this field, which the html_template should handle).
			if fallback != nil {
				resolvedData[fieldName] = fallback
			}
			continue
		}

		// Renderer/static fields — resolved at render time, not now
		if source == "renderer" || source == "static" ||
			strings.HasPrefix(source, "renderer.") ||
			strings.HasPrefix(source, "static.") {
			if fallback != nil {
				resolvedData[fieldName] = fallback
			}
			continue
		}

		// Resolve data source
		value, found := resolver.resolve(ctx, source)

		if found && value != nil {
			resolvedData[fieldName] = value
			continue
		}

		// Data not found — apply on_missing rule
		if !found || value == nil {
			if !required {
				// Optional field missing — apply on_missing
				switch onMissing {
				case "use_fallback":
					if fallback != nil {
						resolvedData[fieldName] = fallback
					}
				case "skip_field":
					// Just omit it
				case "skip_section":
					shouldSkip = true
				case "needs_human_review":
					shouldDefer = true
					missingFields = append(missingFields, missingField{
						Field:     fieldName,
						Source:    source,
						OnMissing: onMissing,
						Reason:    missingReason,
						Type:      fieldType,
						Items:     fieldItems,
						MinItems:  fieldMinItems,
					})
				}
				continue
			}

			// Required field missing — apply on_missing
			switch onMissing {
			case "skip_field":
				// Required-but-skippable: honour the schema's declared intent and
				// omit the field instead of deferring the section (mirrors the
				// optional branch; templates gate on the field).
				logger.Info("plan_sections: required field missing with on_missing=skip_field — omitting field",
					zap.String("field", fieldName),
					zap.String("source", source))
			case "use_fallback":
				if fallback != nil {
					resolvedData[fieldName] = fallback
				} else {
					// Required with no fallback — defer
					shouldDefer = true
					missingFields = append(missingFields, missingField{
						Field:     fieldName,
						Source:    source,
						OnMissing: onMissing,
						Reason:    missingReason,
						Type:      fieldType,
						Items:     fieldItems,
						MinItems:  fieldMinItems,
					})
				}
			case "skip_section":
				shouldSkip = true
			case "needs_human_review":
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: onMissing,
					Reason:    missingReason,
					Type:      fieldType,
					Items:     fieldItems,
					MinItems:  fieldMinItems,
				})
			case "block":
				// Block entire page build — this is handled upstream
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: "block",
					Reason:    missingReason,
					Type:      fieldType,
					Items:     fieldItems,
					MinItems:  fieldMinItems,
				})
			default:
				// Unknown on_missing — default to defer for safety
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: onMissing,
					Reason:    missingReason,
					Type:      fieldType,
					Items:     fieldItems,
					MinItems:  fieldMinItems,
				})
			}
		}
	}

	// Skip takes priority over defer
	if shouldSkip {
		item.Status = "skipped"
		if len(missingFields) > 0 {
			item.Reason = fmt.Sprintf("missing data: %s", missingFields[0].Reason)
		} else {
			item.Reason = "on_missing=skip_section triggered"
		}
		item.Missing = missingFields
		return item
	}

	if shouldDefer {
		item.Status = "deferred"
		item.Missing = missingFields
		if len(missingFields) > 0 {
			reasons := make([]string, len(missingFields))
			for i, m := range missingFields {
				if m.Reason != "" {
					reasons[i] = m.Reason
				} else {
					reasons[i] = fmt.Sprintf("%s (from %s)", m.Field, m.Source)
				}
			}
			item.Reason = strings.Join(reasons, "; ")
		}
		return item
	}

	// Authoritative hero aliasing: when this section declares an image-typed
	// field, also write the resolved page hero under the legacy alias keys
	// (hero_url, background_image) unless the schema declares them itself.
	// resolved_data is merged LAST at render time (RenderComponentAction's
	// merge_with overlay — "resolved data wins on conflicts, by design"), so
	// this is what lets the per-page hero defeat the site-wide hero_url that
	// BuildRenderContext still injects for legacy templates: without it,
	// {{or .hero_url .background_image}} picks the site-wide value and every
	// page shows the same image.
	if sectionHasImageField(fieldsRaw) {
		resolver.ensureAssets(ctx)
		if heroURL, ok := resolver.assets["hero"]; ok && heroURL != "" {
			for _, alias := range []string{"hero_url", "background_image"} {
				if _, declared := fieldsRaw[alias]; declared {
					continue // the field's own resolution governs
				}
				if _, already := resolvedData[alias]; !already {
					resolvedData[alias] = heroURL
				}
			}
		}
	}

	// Section is ready
	item.ResolvedData = resolvedData
	item.LLMFields = llmFields
	item.LLMFieldSpecs = llmFieldSpecs
	return item
}

// sectionHasImageField reports whether any declared field in a component's
// input_schema fields map is image-typed (type "image" or "image_url").
func sectionHasImageField(fieldsRaw map[string]interface{}) bool {
	for _, defRaw := range fieldsRaw {
		def, ok := defRaw.(map[string]interface{})
		if !ok {
			continue
		}
		if t, _ := def["type"].(string); t == "image" || t == "image_url" {
			return true
		}
	}
	return false
}

// filterSiteLevelSections removes section names that correspond to site-level
// components (headers, footers). These are managed by site_components and injected
// during page assembly — they should never enter the page_components pipeline.
func filterSiteLevelSections(sections []string, logger *zap.Logger) []string {
	filtered := make([]string, 0, len(sections))
	for _, s := range sections {
		lower := strings.ToLower(s)
		if strings.Contains(lower, "header") ||
			strings.Contains(lower, "footer") ||
			lower == "site-header" ||
			lower == "site-footer" ||
			lower == "head" ||
			strings.HasPrefix(lower, "head-") {
			logger.Info("plan_sections: filtered site-level section",
				zap.String("section", s))
			continue
		}
		filtered = append(filtered, s)
	}
	return filtered
}

// ============================================================================
// Check for open data requests (prevents repeated LLM waste)
// ============================================================================

// loadOpenSectionDataRequests returns a map of section_name → reason for all
// sections on this page that have an open needs_section_data work item.
// "Open" means status not in a terminal state (complete, wont_fix, rejected, failed).
func loadOpenSectionDataRequests(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName string, logger *zap.Logger) map[string]string {
	result := make(map[string]string)

	rows, err := db.QueryContext(ctx, `
		SELECT spec->>'section_name', LEFT(summary, 120)
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'needs_section_data'
		  AND spec->>'page_name' = $2
		  AND status NOT IN ('complete', 'wont_fix', 'rejected', 'failed')
	`, siteID, pageName)
	if err != nil {
		logger.Warn("loadOpenSectionDataRequests: query failed", zap.Error(err))
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var sectionName, summary string
		if err := rows.Scan(&sectionName, &summary); err != nil {
			continue
		}
		if sectionName != "" {
			result[sectionName] = summary
		}
	}

	if len(result) > 0 {
		logger.Info("loadOpenSectionDataRequests: found open data requests",
			zap.Int("count", len(result)),
			zap.String("page", pageName))
	}

	return result
}

// closeResolvedDataRequest marks a needs_section_data item as complete when
// plan_sections determines the section is now ready (component created, data
// arrived, etc.). This closes the feedback loop — data requests don't block
// sections forever once the underlying issue is resolved.
func closeResolvedDataRequest(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, sectionName string, logger *zap.Logger) {
	result, err := db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete',
		    completed_at = NOW(),
		    handled_by = 'plan_sections',
		    result = jsonb_build_object('auto_resolved', true, 'reason', 'section now ready — component or data available'),
		    updated_at = NOW()
		WHERE site_id = $1
		  AND item_type = 'needs_section_data'
		  AND spec->>'page_name' = $2
		  AND spec->>'section_name' = $3
		  AND status NOT IN ('complete', 'wont_fix', 'rejected', 'failed')
	`, siteID, pageName, sectionName)
	if err != nil {
		logger.Warn("closeResolvedDataRequest: update failed",
			zap.String("section", sectionName),
			zap.String("page", pageName),
			zap.Error(err))
		return
	}
	if rows, _ := result.RowsAffected(); rows > 0 {
		logger.Info("closeResolvedDataRequest: stale data request auto-closed",
			zap.String("section", sectionName),
			zap.String("page", pageName))
	}
}

// ============================================================================
// Create work items for deferred sections
// ============================================================================

func createDeferredItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageName, parentWorkItemID string, deferred []sectionPlanItem, logger *zap.Logger) {
	var parentID *uuid.UUID
	if parentWorkItemID != "" {
		if parsed, err := uuid.Parse(parentWorkItemID); err == nil {
			parentID = &parsed
		}
	}

	for _, section := range deferred {
		// Build missing fields summary
		missingDescs := make([]string, len(section.Missing))
		for i, m := range section.Missing {
			if m.Reason != "" {
				missingDescs[i] = m.Reason
			} else {
				missingDescs[i] = fmt.Sprintf("field '%s' from %s", m.Field, m.Source)
			}
		}

		spec := map[string]interface{}{
			"page_name":    pageName,
			"section_name": section.Name,
			"component_id": section.ComponentID,
			"function":     section.Function,
			"missing":      section.Missing,
			"source":       "plan_sections",
		}
		specJSON, _ := json.Marshal(spec)

		summary := fmt.Sprintf("Section '%s' on %s needs: %s",
			section.Name, pageName, strings.Join(missingDescs, "; "))
		if len(summary) > 250 {
			summary = summary[:247] + "..."
		}

		itemKey := fmt.Sprintf("section_data_%s_%s_%s",
			pageName, sanitiseSectionKey(section.Name), siteID)

		_, err := db.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, priority, status, created_by,
				item_key, parent_item_id
			) VALUES ($1, 'section-planner', 'build', 'needs_section_data', 'medium', $2,
					  $3::jsonb, 50, 'needs_human_review',
					  'plan_sections', $4, $5)
			ON CONFLICT DO NOTHING
		`, siteID, summary, string(specJSON), itemKey, parentID)

		if err != nil {
			logger.Warn("createDeferredItems: failed to insert",
				zap.String("section", section.Name), zap.Error(err))
		} else {
			logger.Info("createDeferredItems: HITL item created",
				zap.String("section", section.Name),
				zap.String("page", pageName))
		}
	}
}

func sanitiseSectionKey(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			return r
		}
		if r == ' ' {
			return '_'
		}
		return -1
	}, s)
	if len(s) > 40 {
		s = s[:40]
	}
	return s
}

// stringOrEmpty extracts a string from an interface{}, returning "" when the
// value isn't a string or is nil. Used when reading optional fields off the
// parsed input_schema map (llm_guidance, on_missing) where the schema author
// may simply omit the key.
// extractArrayItemFields returns the sorted field names each element of an
// array-typed input_schema field must contain. Supports both conventions in
// use: `items` (flat name->type map: faq, differentiators, services-grid) and
// `item_schema` (name->{type,...} map: info-card-grid). Returns nil for
// non-array fields or fields with no declared element shape. Sorted because Go
// map iteration is otherwise random and we want stable prompts and specs.
func extractArrayItemFields(fieldDef map[string]interface{}) []string {
	var fields []string
	if items, ok := fieldDef["items"].(map[string]interface{}); ok {
		for k := range items {
			fields = append(fields, k)
		}
	}
	if itemSchema, ok := fieldDef["item_schema"].(map[string]interface{}); ok {
		for k := range itemSchema {
			fields = append(fields, k)
		}
	}
	sort.Strings(fields)
	return fields
}

func stringOrEmpty(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
