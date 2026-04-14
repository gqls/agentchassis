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
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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
	specsLoaded  bool
	pagesLoaded  bool
	assetsLoaded bool
}

func newSourceResolver(siteID uuid.UUID, db *sql.DB, logger *zap.Logger) *sourceResolver {
	return &sourceResolver{
		siteID: siteID,
		db:     db,
		logger: logger,
		specs:  make(map[string]map[string]interface{}),
		pages:  make(map[string]string),
		assets: make(map[string]string),
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

// loadAssets checks what site assets exist (once)
func (r *sourceResolver) ensureAssets(ctx context.Context) {
	if r.assetsLoaded {
		return
	}
	r.assetsLoaded = true

	// Check content_data for known asset URLs
	var contentDataJSON []byte
	err := r.db.QueryRowContext(ctx, `
		SELECT content_data FROM sites WHERE id = $1
	`, r.siteID).Scan(&contentDataJSON)
	if err != nil {
		return
	}

	var contentData map[string]interface{}
	if err := json.Unmarshal(contentDataJSON, &contentData); err != nil {
		return
	}

	// Map known asset keys
	if heroURL, ok := contentData["hero_url"].(string); ok && heroURL != "" {
		r.assets["hero"] = heroURL
	}
	if logoURL, ok := contentData["logo_url"].(string); ok && logoURL != "" {
		r.assets["logo"] = logoURL
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
		return nil, false

	case "pages":
		r.ensurePages(ctx)
		if url, ok := r.pages[path]; ok {
			return url, true
		}
		// Fallback: construct URL
		return "/" + path + ".html", true

	case "config":
		r.ensureSpecs(ctx)
		return r.resolveConfigPath(path)

	case "query":
		// Query sources are resolved at render time, not at planning time
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

type sectionPlanItem struct {
	Name         string                 `json:"name"`
	ComponentID  string                 `json:"component_id"`
	Function     string                 `json:"function"`
	Status       string                 `json:"status"` // "ready", "deferred", "skipped"
	ResolvedData map[string]interface{} `json:"resolved_data,omitempty"`
	LLMFields    []string               `json:"llm_fields,omitempty"`
	Missing      []missingField         `json:"missing,omitempty"`
	Reason       string                 `json:"reason,omitempty"`
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
	resolver := newSourceResolver(siteID, params.DB, logger)

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

	// Create work items for deferred sections
	if params.DB != nil && len(deferred) > 0 {
		createDeferredItems(ctx, params.DB, siteID, pageName, workItemID, deferred, logger)
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
}

func loadComponentSchemas(ctx context.Context, db *sql.DB, sectionNames []string, logger *zap.Logger) map[string]componentInfo {
	result := make(map[string]componentInfo)

	// Build query for all section names — match by name or function
	placeholders := make([]string, len(sectionNames))
	args := make([]interface{}, len(sectionNames))
	for i, name := range sectionNames {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = name
	}

	query := fmt.Sprintf(`
		SELECT id::text, name, function, COALESCE(input_schema::text, '{}'),
			CASE WHEN html_template LIKE '%%</section>%%' THEN true
			     WHEN html_template IS NULL THEN true
			     WHEN LENGTH(html_template) < 100 THEN true
			     ELSE false END as template_valid
		FROM content_components
		WHERE is_active = true
		  AND (name IN (%s) OR function IN (%s))
		ORDER BY name
	`, strings.Join(placeholders, ","), strings.Join(placeholders, ","))

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Warn("plan_sections: failed to load components", zap.Error(err))
		return result
	}
	defer rows.Close()

	for rows.Next() {
		var comp componentInfo
		var schemaJSON string
		var templateValid bool
		if err := rows.Scan(&comp.ID, &comp.Name, &comp.Function, &schemaJSON, &templateValid); err != nil {
			continue
		}

		if !templateValid {
			// Template is truncated (missing </section>) — skip this component
			// so it falls through to Path 3 and gets flagged for regeneration.
			logger.Warn("plan_sections: component template truncated, skipping",
				zap.String("function", comp.Function),
				zap.String("name", comp.Name))
			continue
		}

		json.Unmarshal([]byte(schemaJSON), &comp.InputSchema)

		// Index by both name and function for lookup
		result[comp.Name] = comp
		result[comp.Function] = comp
	}

	return result
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
// Same query pattern as loadComponentSchemas but for a single component.
// Rejects components with truncated templates (missing </section>).
func loadSingleComponentSchema(ctx context.Context, db *sql.DB, function string, logger *zap.Logger) *componentInfo {
	var comp componentInfo
	var schemaJSON string
	var templateValid bool

	err := db.QueryRowContext(ctx, `
		SELECT id::text, name, function, COALESCE(input_schema::text, '{}'),
			CASE WHEN html_template LIKE '%</section>%' THEN true
			     WHEN html_template IS NULL THEN true
			     WHEN LENGTH(html_template) < 100 THEN true
			     ELSE false END as template_valid
		FROM content_components
		WHERE function = $1
		  AND is_active = true
		LIMIT 1
	`, function).Scan(&comp.ID, &comp.Name, &comp.Function, &schemaJSON, &templateValid)

	if err != nil {
		logger.Warn("loadSingleComponentSchema: failed",
			zap.String("function", function),
			zap.Error(err))
		return nil
	}

	if !templateValid {
		logger.Warn("loadSingleComponentSchema: template truncated, rejecting",
			zap.String("function", function))
		return nil
	}

	json.Unmarshal([]byte(schemaJSON), &comp.InputSchema)
	return &comp
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
	}

	// Get fields from schema
	fieldsRaw, ok := comp.InputSchema["fields"].(map[string]interface{})
	if !ok || len(fieldsRaw) == 0 {
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
			continue
		}

		// Renderer/static/query fields — resolved at render time, not now
		if source == "renderer" || source == "static" ||
			strings.HasPrefix(source, "renderer.") ||
			strings.HasPrefix(source, "static.") ||
			strings.HasPrefix(source, "query.") {
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

	// Section is ready
	item.ResolvedData = resolvedData
	item.LLMFields = llmFields
	return item
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
