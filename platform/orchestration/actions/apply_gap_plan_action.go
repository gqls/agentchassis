// FILE: platform/orchestration/actions/apply_gap_plan_action.go
//
// Executes the content-gap-planner's LLM decision. Takes the plan output
// and creates the appropriate database records and work items.
//
// Handles four approaches:
//   - add_to_page:    creates content_rewrite work item for existing page
//   - new_page:       creates page record + needs_content_page work item
//   - update_spec:    writes to site_specs via inline query
//   - not_actionable: marks the original work item as wont_fix
//
// Registration:
//   "apply_gap_plan": {
//       Handler:     ApplyGapPlanAction,
//       Category:    "site",
//       Description: "Execute content gap plan — create pages, work items, or spec updates",
//       IsLocal:     true,
//   },

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

var ApplyGapPlanInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"plan", "work_item_id", "domain"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("apply_gap_plan", ApplyGapPlanInputSpec)
}

func ApplyGapPlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "apply_gap_plan"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		ApplyGapPlanInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	domain := inputs.Get("domain")

	// Parse the original work_item_id so we can mark it complete/wont_fix
	var originalItemID *uuid.UUID
	if itemIDStr := inputs.Get("work_item_id"); itemIDStr != "" {
		if parsed, err := uuid.Parse(itemIDStr); err == nil {
			originalItemID = &parsed
		}
	}

	// Parse the plan — could be string or map
	planField := "gap_plan.result"
	if p, ok := params.StepConfig.Config["plan"].(string); ok && p != "" {
		planField = p
	}

	planRaw := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planRaw == nil {
		// Try fallbacks
		for _, alt := range []string{"gap_plan.result", "gap_plan", "plan"} {
			planRaw = datahelpers.ExtractNestedField(params.CollectedData, alt)
			if planRaw != nil {
				break
			}
		}
	}

	if planRaw == nil {
		return map[string]interface{}{
			"applied":  false,
			"approach": "none",
			"reason":   "no plan found",
		}, nil
	}

	// Parse plan into a map
	var plan map[string]interface{}

	switch v := planRaw.(type) {
	case string:
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
		}
	case map[string]interface{}:
		plan = v
	default:
		return nil, fmt.Errorf("unexpected plan type: %T", planRaw)
	}

	approach, _ := plan["approach"].(string)
	reasoning, _ := plan["reasoning"].(string)

	logger.Info("ApplyGapPlanAction: executing plan",
		zap.String("approach", approach),
		zap.String("reasoning", reasoning),
		zap.String("site_id", siteIDStr))

	switch approach {

	case "add_to_page":
		return applyAddToPage(ctx, params.DB, plan, siteID, originalItemID, logger)

	case "new_page":
		return applyNewPage(ctx, params.DB, plan, siteID, domain, originalItemID, logger)

	case "update_spec":
		return applyUpdateSpec(ctx, params.DB, plan, siteID, originalItemID, logger)

	case "not_actionable":
		return applyNotActionable(ctx, params.DB, plan, originalItemID, logger)

	default:
		logger.Warn("Unknown plan approach",
			zap.String("approach", approach))
		return map[string]interface{}{
			"applied":  false,
			"approach": approach,
			"reason":   "unknown approach",
		}, nil
	}
}

// ============================================================================
// add_to_page: create a content_rewrite work item for an existing page
// ============================================================================

func applyAddToPage(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	addPlan, ok := plan["add_to_page"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("add_to_page plan missing or invalid")
	}

	pageNameRaw, _ := addPlan["page_name"].(string)
	if pageNameRaw == "" {
		return nil, fmt.Errorf("add_to_page.page_name is required")
	}

	contentGuidance, _ := addPlan["content_guidance"].(string)
	reasoning, _ := plan["reasoning"].(string)

	// LLM sometimes produces comma-separated page names like "index, services".
	// Split and process each one individually.
	pageNames := splitPageNames(pageNameRaw)

	var created int
	var skipped []string

	for _, pageName := range pageNames {
		// Look up the page
		var pageID uuid.UUID
		err := db.QueryRowContext(ctx, `
			SELECT id FROM pages WHERE site_id = $1 AND name = $2 LIMIT 1
		`, siteID, pageName).Scan(&pageID)

		if err == sql.ErrNoRows {
			logger.Warn("Page not found, skipping",
				zap.String("page_name", pageName),
				zap.String("raw_input", pageNameRaw))
			skipped = append(skipped, pageName)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("query page %q: %w", pageName, err)
		}

		// Build spec for the content rewrite
		spec := map[string]interface{}{
			"page_name":        pageName,
			"content_guidance": contentGuidance,
			"source":           "content-gap-planner",
		}
		if addSections, ok := addPlan["add_sections"].([]interface{}); ok {
			spec["add_sections"] = addSections
		}
		specJSON, _ := json.Marshal(spec)

		summary := fmt.Sprintf("Add content to %s: %s", pageName, truncate(reasoning, 80))
		itemKey := fmt.Sprintf("gap_plan_add_%s_%s", pageName, siteID)

		_, err = db.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, domain, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by,
				item_key, parent_item_id
			) VALUES ($1, 'content-gap-planner', 'build', 'content_rewrite', 'medium', $2,
			          $3::jsonb, $4, 35, 'page-build-handler', 'triaged', 'content-gap-planner',
			          $5, $6)
			ON CONFLICT DO NOTHING
		`, siteID, summary, string(specJSON), pageID, itemKey, originalItemID)

		if err != nil {
			logger.Warn("Failed to create work item for page",
				zap.String("page_name", pageName), zap.Error(err))
			continue
		}
		created++
	}

	if created == 0 && len(skipped) > 0 {
		return nil, fmt.Errorf("no valid pages found from %q — pages not found: %s", pageNameRaw, strings.Join(skipped, ", "))
	}

	// Mark original as complete
	markOriginalComplete(ctx, db, originalItemID)

	logger.Info("ApplyGapPlanAction: add_to_page applied",
		zap.String("raw_page_names", pageNameRaw),
		zap.Int("items_created", created),
		zap.Strings("skipped", skipped))

	return map[string]interface{}{
		"applied":       true,
		"approach":      "add_to_page",
		"pages":         pageNames,
		"items_created": created,
		"skipped":       skipped,
	}, nil
}

// ============================================================================
// new_page: create page record + needs_content_page work item
// ============================================================================

func applyNewPage(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, domain string, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	newPlan, ok := plan["new_page"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("new_page plan missing or invalid")
	}

	pageName, _ := newPlan["name"].(string)
	if pageName == "" {
		return nil, fmt.Errorf("new_page.name is required")
	}
	pageName = strings.ToLower(strings.ReplaceAll(pageName, " ", "-"))

	title, _ := newPlan["title"].(string)
	if title == "" {
		title = strings.Title(strings.ReplaceAll(pageName, "-", " "))
		if domain != "" {
			title = title + " | " + domain
		}
	}

	pageType, _ := newPlan["page_type"].(string)
	if pageType == "" {
		pageType = "content"
	}

	purpose, _ := newPlan["purpose"].(string)

	// Parse sections
	sections := []string{"hero", "generic-text-block", "call-to-action"}
	if sectionsRaw, ok := newPlan["sections"].([]interface{}); ok && len(sectionsRaw) > 0 {
		sections = nil
		for _, s := range sectionsRaw {
			if ss, ok := s.(string); ok {
				sections = append(sections, ss)
			}
		}
	}
	sectionsJSON, _ := json.Marshal(sections)

	navLabel, _ := newPlan["nav_label"].(string)
	inHeader := true
	if ih, ok := newPlan["in_header"].(bool); ok {
		inHeader = ih
	}
	inFooter := true
	if inf, ok := newPlan["in_footer"].(bool); ok {
		inFooter = inf
	}

	url := "/" + pageName + ".html"

	// Create the page record
	var pageID uuid.UUID
	err := db.QueryRowContext(ctx, `
		INSERT INTO pages (site_id, name, url, title, page_type, build_status,
		                   sections, nav_label, in_header, in_footer)
		VALUES ($1, $2, $3, $4, $5, 'planned', $6::jsonb, $7, $8, $9)
		ON CONFLICT (site_id, name) DO UPDATE SET
			title = EXCLUDED.title,
			sections = EXCLUDED.sections,
			updated_at = NOW()
		RETURNING id
	`, siteID, pageName, url, title, pageType,
		string(sectionsJSON), navLabel, inHeader, inFooter,
	).Scan(&pageID)

	if err != nil {
		return nil, fmt.Errorf("create page: %w", err)
	}

	// Create work item for page-build-handler
	spec := map[string]interface{}{
		"page_name": pageName,
		"page_type": pageType,
		"title":     title,
		"purpose":   purpose,
		"sections":  sections,
		"source":    "content-gap-planner",
	}
	specJSON, _ := json.Marshal(spec)

	reasoning, _ := plan["reasoning"].(string)
	summary := fmt.Sprintf("Build new page: %s — %s", title, truncate(reasoning, 60))

	itemKey := fmt.Sprintf("gap_plan_new_%s_%s", pageName, siteID)

	_, err = db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, domain, item_type, severity, summary,
			spec, page_id, priority, handler_agent, status, created_by,
			item_key, parent_item_id
		) VALUES ($1, 'content-gap-planner', 'build', 'needs_content_page', 'medium', $2,
		          $3::jsonb, $4, 40, 'page-build-handler', 'triaged', 'content-gap-planner',
		          $5, $6)
		ON CONFLICT DO NOTHING
	`, siteID, summary, string(specJSON), pageID, itemKey, originalItemID)

	if err != nil {
		return nil, fmt.Errorf("create work item: %w", err)
	}

	// Mark original as complete
	markOriginalComplete(ctx, db, originalItemID)

	logger.Info("ApplyGapPlanAction: new_page created",
		zap.String("page_name", pageName),
		zap.String("page_id", pageID.String()))

	return map[string]interface{}{
		"applied":      true,
		"approach":     "new_page",
		"page_name":    pageName,
		"page_id":      pageID.String(),
		"item_created": true,
	}, nil
}

// ============================================================================
// update_spec: write a value to site_specs
// ============================================================================

func applyUpdateSpec(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	specPlan, ok := plan["update_spec"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("update_spec plan missing or invalid")
	}

	aspect, _ := specPlan["aspect"].(string)
	if aspect == "" {
		aspect = "identity"
	}

	field, _ := specPlan["field"].(string)
	suggestedValue := specPlan["suggested_value"]

	if field == "" || suggestedValue == nil {
		return nil, fmt.Errorf("update_spec needs field and suggested_value")
	}

	// Read current spec
	var currentDataJSON []byte
	var oldID *uuid.UUID
	err := db.QueryRowContext(ctx, `
		SELECT id, data FROM site_specs
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect).Scan(&oldID, &currentDataJSON)

	var currentData map[string]interface{}
	if err == sql.ErrNoRows {
		currentData = make(map[string]interface{})
	} else if err != nil {
		return nil, fmt.Errorf("read current spec: %w", err)
	} else {
		json.Unmarshal(currentDataJSON, &currentData)
	}

	// Merge
	currentData[field] = suggestedValue
	mergedJSON, _ := json.Marshal(currentData)

	// Supersede old
	if oldID != nil {
		db.ExecContext(ctx, `
			UPDATE site_specs SET is_current = false, superseded_at = now() WHERE id = $1
		`, *oldID)
	}

	// Insert new
	_, err = db.ExecContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, source_agent, is_current, created_by)
		VALUES ($1, $2, $3::jsonb, 'content-gap-planner', 'content-gap-planner', true, 'content-gap-planner')
	`, siteID, aspect, string(mergedJSON))

	if err != nil {
		return nil, fmt.Errorf("write spec: %w", err)
	}

	// Mark original as complete
	markOriginalComplete(ctx, db, originalItemID)

	logger.Info("ApplyGapPlanAction: update_spec applied",
		zap.String("aspect", aspect),
		zap.String("field", field),
		zap.String("value_preview", truncate(fmt.Sprintf("%v", suggestedValue), 60)))

	return map[string]interface{}{
		"applied":  true,
		"approach": "update_spec",
		"aspect":   aspect,
		"field":    field,
	}, nil
}

// ============================================================================
// not_actionable: mark original item as wont_fix
// ============================================================================

func applyNotActionable(ctx context.Context, db *sql.DB, plan map[string]interface{}, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error) {
	reason := "Not actionable"
	if naPlan, ok := plan["not_actionable"].(map[string]interface{}); ok {
		if r, ok := naPlan["reason"].(string); ok {
			reason = r
		}
	}

	if originalItemID != nil {
		db.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'wont_fix', error = $2, completed_at = NOW()
			WHERE id = $1
		`, *originalItemID, reason)
	}

	logger.Info("ApplyGapPlanAction: not_actionable",
		zap.String("reason", reason))

	return map[string]interface{}{
		"applied":  true,
		"approach": "not_actionable",
		"reason":   reason,
	}, nil
}

// ============================================================================
// Helpers
// ============================================================================

func markOriginalComplete(ctx context.Context, db *sql.DB, itemID *uuid.UUID) {
	if itemID != nil {
		db.ExecContext(ctx, `
			UPDATE site_work_items
			SET status = 'complete', completed_at = NOW(), handled_by = 'content-gap-planner'
			WHERE id = $1 AND status IN ('triaged', 'claimed')
		`, *itemID)
	}
}

// splitPageNames handles LLM output that sometimes produces comma-separated,
// "and"-separated, or slash-separated page lists like "index, services",
// "index and services", or "index/services".
// Returns cleaned, trimmed individual page names.
func splitPageNames(raw string) []string {
	// Normalise separators
	normalized := strings.ReplaceAll(raw, " and ", ",")
	normalized = strings.ReplaceAll(normalized, " & ", ",")
	normalized = strings.ReplaceAll(normalized, "/", ",")

	parts := strings.Split(normalized, ",")
	var result []string
	seen := map[string]bool{}
	for _, p := range parts {
		name := strings.TrimSpace(p)
		name = strings.ToLower(name)
		if name != "" && !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	return result
}
