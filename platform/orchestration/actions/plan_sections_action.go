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
	Optional:   []string{"sections", "page_name", "domain", "work_item_id"},
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
	Field     string `json:"field"`
	Source    string `json:"source"`
	OnMissing string `json:"on_missing"`
	Reason    string `json:"reason"`
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
	domain := inputs.Get("domain")
	workItemID := inputs.Get("work_item_id")

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

	// Plan each section
	var ready []sectionPlanItem
	var deferred []sectionPlanItem
	var skipped []sectionPlanItem

	for _, sectionName := range sectionNames {
		comp, ok := components[sectionName]
		if !ok {
			logger.Warn("plan_sections: component not found for section",
				zap.String("section", sectionName))
			// Unknown section — let content writer handle it (backward compat)
			ready = append(ready, sectionPlanItem{
				Name:   sectionName,
				Status: "ready",
				Reason: "no component found — passing to content writer as-is",
			})
			continue
		}

		item := planSection(ctx, sectionName, comp, resolver, logger)

		switch item.Status {
		case "ready":
			ready = append(ready, item)
		case "deferred":
			deferred = append(deferred, item)
		case "skipped":
			skipped = append(skipped, item)
		}
	}

	// Create work items for deferred sections
	if params.DB != nil && len(deferred) > 0 {
		createDeferredItems(ctx, params.DB, siteID, domain, pageName, workItemID, deferred, logger)
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
		SELECT id::text, name, function, COALESCE(input_schema::text, '{}')
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
		if err := rows.Scan(&comp.ID, &comp.Name, &comp.Function, &schemaJSON); err != nil {
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
		// No v2 schema — treat as fully LLM-generated (backward compat)
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

		// LLM-generated fields — always available
		if source == "llm" {
			llmFields = append(llmFields, fieldName)
			continue
		}

		// Renderer/static/query fields — resolved at render time, not now
		if source == "renderer" || source == "static" || strings.HasPrefix(source, "query.") {
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
				})
			case "block":
				// Block entire page build — this is handled upstream
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: "block",
					Reason:    missingReason,
				})
			default:
				// Unknown on_missing — default to defer for safety
				shouldDefer = true
				missingFields = append(missingFields, missingField{
					Field:     fieldName,
					Source:    source,
					OnMissing: onMissing,
					Reason:    missingReason,
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

// ============================================================================
// Create work items for deferred sections
// ============================================================================

func createDeferredItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, domain, pageName, parentWorkItemID string, deferred []sectionPlanItem, logger *zap.Logger) {
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
				site_id, source, domain, item_type, severity, summary,
				spec, priority, status, created_by,
				item_key, parent_item_id
			) VALUES ($1, 'section-planner', $2, 'needs_section_data', 'medium', $3,
			          $4::jsonb, 50, 'needs_human_review',
			          'plan_sections', $5, $6)
			ON CONFLICT DO NOTHING
		`, siteID, domain, summary, string(specJSON), itemKey, parentID)

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
