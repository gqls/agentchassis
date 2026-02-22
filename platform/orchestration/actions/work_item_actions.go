// FILE: platform/orchestration/actions/work_item_actions.go
// Work item actions for the unified build/maintenance queue (site_work_items table).
//
// Actions:
//   WriteBuildItemsAction    — queries pages from DB, creates work items in site_work_items
//   LoadWorkItemsAction      — loads pending items for a site, respecting dependencies
//   CompleteWorkItemAction   — marks a work item as complete with result data
//   FailWorkItemAction       — marks a work item as failed with error info
//
// These actions run alongside existing maintenance_queue actions. No migration needed.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpecs
// ============================================================================

var WriteBuildItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"site_plan", "batch_id"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

// LoadWorkItemsInputSpec — only site_id needs resolution from collectedData.
// Filter fields (item_domain, handler_agent, max_items) are pure config literals
// and are read directly from params.StepConfig.Config in the action body.
// NOTE: "domain" is NOT listed here because site_record.domain would be picked
// up by nested lookup (field name collision — see checklist section on this).
var LoadWorkItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

var CompleteWorkItemInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"work_item_id"},
	Optional: []string{"result", "commit_sha"},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"work_item_id_field": "work_item_id",
		"commit_sha_field":   "commit_sha",
		"result_field":       "result",
	},
}

// FailWorkItemInputSpec — work_item_id needs path resolution.
// error_message is a config literal (e.g. "Content review not approved")
// and is read directly from params.StepConfig.Config.
var FailWorkItemInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"work_item_id"},
	Optional: []string{},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"work_item_id_field": "work_item_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_build_items", WriteBuildItemsInputSpec)
	datahelpers.RegisterActionInputSpec("load_work_items", LoadWorkItemsInputSpec)
	datahelpers.RegisterActionInputSpec("complete_work_item", CompleteWorkItemInputSpec)
	datahelpers.RegisterActionInputSpec("fail_work_item", FailWorkItemInputSpec)
}

// ============================================================================
// ACTION: write_build_items
// Used by: site-work-orchestrator (after sync_pages_to_db)
// ============================================================================
//
// Queries pages from the pages table (build_status = 'planned') and creates
// one work item per page. The spec field contains the full page record from DB,
// matching exactly what get_pages_to_build returns — so current_item.spec has
// the same fields as current_page in pageflow-builder (id, name, url, title,
// sections, page_type, etc).
//
// Also creates tracking items for site-wide tasks (logo, hero, design) based
// on flags in the site_plan.
//
// Data inputs:
//   - site_id (required) — UUID of the site (pages must already be synced)
//   - site_plan (optional) — planner output for asset/design flags
//   - batch_id (optional) — UUID to group items; generated if not provided
func WriteBuildItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("WriteBuildItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		WriteBuildItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	batchID := uuid.New()
	if b := inputs.Get("batch_id"); b != "" {
		if parsed, err := uuid.Parse(b); err == nil {
			batchID = parsed
		}
	}

	// ---- Query pages from DB (same data as get_pages_to_build) ----
	pages, err := queryPagesForBuild(ctx, params.DB, siteID, []string{"planned"}, false, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}

	if len(pages) == 0 {
		logger.Warn("WriteBuildItemsAction: No planned pages found")
		return map[string]interface{}{
			"items_inserted": 0,
			"items_skipped":  0,
			"batch_id":       batchID.String(),
			"total_items":    0,
		}, nil
	}

	// ---- Extract plan flags for asset/design items ----
	planData := inputs.GetMap("site_plan")
	if planData == nil {
		planData = make(map[string]interface{})
	}
	if resp, ok := planData["response"].(map[string]interface{}); ok {
		if pd, ok := resp["plan_data"].(map[string]interface{}); ok {
			planData = pd
		} else {
			planData = resp
		}
	}

	needsLogo, _ := planData["needs_logo"].(bool)
	needsImages, _ := planData["needs_images"].(bool)

	// ---- Insert work items in a transaction ----
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	inserted := 0
	skipped := 0

	// Content page items
	for i, page := range pages {
		pageName, _ := page["name"].(string)
		if pageName == "" {
			continue
		}

		pageIDStr, _ := page["id"].(string)
		var pageIDPtr *uuid.UUID
		if parsed, err := uuid.Parse(pageIDStr); err == nil {
			pageIDPtr = &parsed
		}

		handlerAgent := "page-content-writer"
		itemType := "needs_content_page"
		if pt, _ := page["page_type"].(string); pt != "" {
			switch pt {
			case "tool":
				handlerAgent = "tool-builder"
				itemType = "needs_tool_page"
			case "entity-directory":
				handlerAgent = "directory-builder"
				itemType = "needs_directory"
			case "entity-page":
				handlerAgent = "entity-page-builder"
				itemType = "needs_entity_page"
			}
		}

		specJSON, err := json.Marshal(page)
		if err != nil {
			logger.Warn("WriteBuildItemsAction: marshal error",
				zap.String("page", pageName), zap.Error(err))
			continue
		}

		ok, err := insertWorkItem(ctx, tx, workItem{
			siteID:       siteID,
			source:       "planner",
			domain:       "build",
			itemType:     itemType,
			severity:     "high",
			summary:      fmt.Sprintf("Build page: %s", pageName),
			spec:         string(specJSON),
			pageID:       pageIDPtr,
			priority:     10 + i,
			handlerAgent: handlerAgent,
			status:       "triaged",
			createdBy:    "site-planner",
			itemKey:      fmt.Sprintf("needs_page:%s", pageName),
			batchID:      batchID,
		}, logger)
		if err != nil {
			logger.Warn("WriteBuildItemsAction: insert error",
				zap.String("page", pageName), zap.Error(err))
			continue
		}
		if ok {
			inserted++
		} else {
			skipped++
		}
	}

	// Asset tracking items
	if needsLogo {
		logoSpec, _ := json.Marshal(map[string]interface{}{
			"purpose":       "logo",
			"image_prompts": planData["image_prompts"],
		})
		ok, _ := insertWorkItem(ctx, tx, workItem{
			siteID: siteID, source: "planner", domain: "build",
			itemType: "needs_logo", severity: "high",
			summary: "Generate site logo", spec: string(logoSpec),
			priority: 5, handlerAgent: "image-generator",
			status: "triaged", createdBy: "site-planner",
			itemKey: "needs_logo", batchID: batchID,
		}, logger)
		if ok {
			inserted++
		}
	}
	if needsImages {
		heroSpec, _ := json.Marshal(map[string]interface{}{
			"purpose":       "hero",
			"image_prompts": planData["image_prompts"],
		})
		ok, _ := insertWorkItem(ctx, tx, workItem{
			siteID: siteID, source: "planner", domain: "build",
			itemType: "needs_hero_image", severity: "medium",
			summary: "Generate hero image", spec: string(heroSpec),
			priority: 5, handlerAgent: "image-generator",
			status: "triaged", createdBy: "site-planner",
			itemKey: "needs_hero:home", batchID: batchID,
		}, logger)
		if ok {
			inserted++
		}
	}

	// Design tracking item
	ok, _ := insertWorkItem(ctx, tx, workItem{
		siteID: siteID, source: "planner", domain: "design",
		itemType: "needs_design", severity: "high",
		summary: "Generate site stylesheet", spec: "{}",
		priority: 8, handlerAgent: "webdesign-agent",
		status: "triaged", createdBy: "site-planner",
		itemKey: "needs_design", batchID: batchID,
	}, logger)
	if ok {
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit work items: %w", err)
	}

	logger.Info("WriteBuildItemsAction: Complete",
		zap.Int("items_inserted", inserted),
		zap.Int("items_skipped", skipped),
		zap.Int("pages_found", len(pages)),
		zap.String("batch_id", batchID.String()),
	)

	return map[string]interface{}{
		"items_inserted": inserted,
		"items_skipped":  skipped,
		"batch_id":       batchID.String(),
		"total_items":    inserted + skipped,
		"page_count":     len(pages),
	}, nil
}

// ============================================================================
// ACTION: load_work_items
// Used by: site-work-orchestrator (to get items to process)
// ============================================================================
//
// Returns items ordered by priority. Filters out items whose dependencies
// haven't completed. Each item's spec field is a parsed map (not JSON string)
// so dot-notation access works: current_item.spec.name, current_item.spec.id.
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — resolved from collectedData via path
//
// Config literals (read directly, NOT through ExtractActionInputs):
//   - item_domain (optional) — filter by work item domain (e.g. "build")
//     Named item_domain to avoid collision with site_record.domain
//   - handler_agent (optional) — filter by handler agent type
//   - max_items (optional, default 50)
func LoadWorkItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("LoadWorkItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// site_id needs path resolution from collectedData
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadWorkItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Filter params are config literals — read directly to avoid
	// nested lookup collisions (e.g. "domain" → site_record.domain)
	config := params.StepConfig.Config
	domainFilter, _ := config["item_domain"].(string)
	handlerFilter, _ := config["handler_agent"].(string)
	maxItems := datahelpers.GetIntField(config, "max_items", 50)

	query := `
		SELECT 
			wi.id, wi.site_id, wi.source, wi.domain, wi.item_type,
			wi.severity, wi.summary, wi.spec, wi.page_id, 
			wi.priority, wi.handler_agent, wi.status, wi.item_key,
			wi.batch_id, wi.attempt_count
		FROM site_work_items wi
		WHERE wi.site_id = $1
		  AND wi.status IN ('triaged', 'approved')
		  AND wi.attempt_count < wi.max_attempts
		  AND (
		    wi.depends_on IS NULL 
		    OR NOT EXISTS (
		      SELECT 1 FROM unnest(wi.depends_on) dep_id
		      WHERE dep_id NOT IN (
		        SELECT id FROM site_work_items
		        WHERE site_id = $1 AND status IN ('complete', 'verified')
		      )
		    )
		  )
	`

	args := []interface{}{siteID}
	argIdx := 2

	if domainFilter != "" {
		query += fmt.Sprintf(" AND wi.domain = $%d", argIdx)
		args = append(args, domainFilter)
		argIdx++
	}

	if handlerFilter != "" {
		query += fmt.Sprintf(" AND wi.handler_agent = $%d", argIdx)
		args = append(args, handlerFilter)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY wi.priority ASC, wi.created_at ASC LIMIT $%d", argIdx)
	args = append(args, maxItems)

	rows, err := params.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var workItems []interface{}
	handlerSet := make(map[string]bool)

	for rows.Next() {
		var (
			id, wiSiteID             uuid.UUID
			source, domain, itemType string
			severity, summary        string
			specJSON                 []byte
			pageID                   *uuid.UUID
			priority                 int
			handlerAgent, status     string
			itemKey                  *string
			batchID                  *uuid.UUID
			attemptCount             int
		)

		err := rows.Scan(
			&id, &wiSiteID, &source, &domain, &itemType,
			&severity, &summary, &specJSON, &pageID,
			&priority, &handlerAgent, &status, &itemKey,
			&batchID, &attemptCount,
		)
		if err != nil {
			logger.Warn("LoadWorkItemsAction: scan error", zap.Error(err))
			continue
		}

		// Parse spec into map so current_item.spec.name works in dot-notation
		var specData interface{}
		if len(specJSON) > 0 {
			var specMap map[string]interface{}
			if err := json.Unmarshal(specJSON, &specMap); err == nil {
				specData = specMap
			} else {
				specData = string(specJSON)
			}
		}

		item := map[string]interface{}{
			"id":            id.String(),
			"site_id":       wiSiteID.String(),
			"source":        source,
			"domain":        domain,
			"item_type":     itemType,
			"severity":      severity,
			"summary":       summary,
			"spec":          specData,
			"priority":      priority,
			"handler_agent": handlerAgent,
			"status":        status,
			"attempt_count": attemptCount,
		}

		if pageID != nil {
			item["page_id"] = pageID.String()
		}
		if itemKey != nil {
			item["item_key"] = *itemKey
		}
		if batchID != nil {
			item["batch_id"] = batchID.String()
		}

		workItems = append(workItems, item)
		handlerSet[handlerAgent] = true
	}

	var agents []string
	for agent := range handlerSet {
		agents = append(agents, agent)
	}

	logger.Info("LoadWorkItemsAction: Complete",
		zap.Int("items_loaded", len(workItems)),
		zap.Strings("handler_agents", agents),
	)

	return map[string]interface{}{
		"items":          workItems,
		"item_count":     len(workItems),
		"has_items":      len(workItems) > 0,
		"handler_agents": agents,
	}, nil
}

// ============================================================================
// ACTION: complete_work_item
// Used by: site-work-orchestrator loop (after page is built and committed)
// ============================================================================
func CompleteWorkItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("CompleteWorkItemAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		CompleteWorkItemInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	itemIDStr := inputs.Get("work_item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id: %w", err)
	}

	resultData := inputs.GetMap("result")
	if resultData == nil {
		resultData = make(map[string]interface{})
	}
	if sha := inputs.Get("commit_sha"); sha != "" {
		resultData["commit_sha"] = sha
	}

	resultJSON, err := json.Marshal(resultData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	agentType := "unknown"
	if params.ExecutionContext.Sender.AgentType != "" {
		agentType = params.ExecutionContext.Sender.AgentType
	}

	_, err = params.DB.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete',
		    result = $2::jsonb,
		    completed_at = NOW(),
		    handled_by = $3
		WHERE id = $1
	`, itemID, string(resultJSON), agentType)
	if err != nil {
		return nil, fmt.Errorf("failed to complete work item: %w", err)
	}

	logger.Info("CompleteWorkItemAction: Done", zap.String("item_id", itemIDStr))

	return map[string]interface{}{
		"completed": true,
		"item_id":   itemIDStr,
	}, nil
}

// ============================================================================
// ACTION: fail_work_item
// ============================================================================
func FailWorkItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("FailWorkItemAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		FailWorkItemInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	itemIDStr := inputs.Get("work_item_id")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id: %w", err)
	}

	// error_message is a config literal (e.g. "Content review not approved"),
	// not a path — read directly from config to avoid path resolution
	errorMsg, _ := params.StepConfig.Config["error_message"].(string)
	agentType := "unknown"
	if params.ExecutionContext.Sender.AgentType != "" {
		agentType = params.ExecutionContext.Sender.AgentType
	}

	var newStatus string
	var attemptsLeft int

	err = params.DB.QueryRowContext(ctx, `
		UPDATE site_work_items
		SET attempt_count = attempt_count + 1,
		    error = $2,
		    status = CASE 
		        WHEN attempt_count + 1 >= max_attempts THEN 'failed'
		        ELSE 'triaged'
		    END,
		    handled_by = $3
		WHERE id = $1
		RETURNING status, max_attempts - (attempt_count)
	`, itemID, errorMsg, agentType).Scan(&newStatus, &attemptsLeft)
	if err != nil {
		return nil, fmt.Errorf("failed to update work item: %w", err)
	}

	logger.Info("FailWorkItemAction: Done",
		zap.String("item_id", itemIDStr),
		zap.String("new_status", newStatus),
		zap.Int("attempts_left", attemptsLeft),
	)

	return map[string]interface{}{
		"item_id":       itemIDStr,
		"new_status":    newStatus,
		"attempts_left": attemptsLeft,
		"will_retry":    newStatus == "triaged",
	}, nil
}

// ============================================================================
// Internal helpers
// ============================================================================

type workItem struct {
	siteID       uuid.UUID
	source       string
	domain       string
	itemType     string
	severity     string
	summary      string
	spec         string
	pageID       *uuid.UUID
	priority     int
	handlerAgent string
	status       string
	createdBy    string
	itemKey      string
	batchID      uuid.UUID
	dependsOn    []uuid.UUID
}

func insertWorkItem(ctx context.Context, tx *sql.Tx, item workItem, logger *zap.Logger) (bool, error) {
	var batchIDPtr *uuid.UUID
	if item.batchID != uuid.Nil {
		batchIDPtr = &item.batchID
	}

	var itemKeyPtr *string
	if item.itemKey != "" {
		itemKeyPtr = &item.itemKey
	}

	var dependsOnStr *string
	if len(item.dependsOn) > 0 {
		ids := "{"
		for i, id := range item.dependsOn {
			if i > 0 {
				ids += ","
			}
			ids += id.String()
		}
		ids += "}"
		dependsOnStr = &ids
	}

	result, err := tx.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, domain, item_type, severity, summary, spec,
			page_id, priority, handler_agent, status, created_by,
			item_key, batch_id, depends_on
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7::jsonb,
			$8, $9, $10, $11, $12,
			$13, $14, $15::uuid[]
		)
		ON CONFLICT (site_id, item_key) 
			WHERE item_key IS NOT NULL 
			  AND status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed')
		DO NOTHING
	`, item.siteID, item.source, item.domain, item.itemType, item.severity,
		item.summary, item.spec,
		item.pageID, item.priority, item.handlerAgent, item.status, item.createdBy,
		itemKeyPtr, batchIDPtr, dependsOnStr,
	)
	if err != nil {
		return false, fmt.Errorf("insert failed for %s: %w", item.itemKey, err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		logger.Debug("Work item inserted",
			zap.String("item_key", item.itemKey),
			zap.String("handler", item.handlerAgent),
			zap.Int("priority", item.priority),
		)
	}
	return rows > 0, nil
}
