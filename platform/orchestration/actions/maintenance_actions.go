// FILE: platform/orchestration/actions/maintenance_actions.go
// Maintenance agent actions for page rebuilds, content refresh, triage scanning.
//
// Actions:
//   LoadSiteForRebuildAction       — loads site context for page-rebuild agent
//   ScanSitesForMaintenanceAction  — scans sites for issues, inserts into maintenance_queue
//   PrepareRebuildDispatchesAction — claims tasks from queue, flags pages, groups by site
//   MarkMaintenanceCompleteAction  — marks queue tasks as complete/failed after specialist

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpecs — standardised input contracts for all maintenance actions
// ============================================================================

// LoadSiteForRebuildInputSpec declares inputs for LoadSiteForRebuildAction.
// site_id comes from the prior ensure_site_record step via collectedData.
// task_id is optional (for future maintenance_queue integration).
var LoadSiteForRebuildInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"task_id"},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"site_id_field": "site_id",
		"task_id_field": "task_id",
	},
}

// ScanSitesForMaintenanceInputSpec declares inputs for ScanSitesForMaintenanceAction.
// domain is optional — omit to scan all deployed sites.
// stale_threshold_days can be passed via input_data to override the default.
// Note: "checks" ([]string) and "deduplicate" (bool) are pure config, not data inputs,
// so they stay as direct config reads in the action body.
var ScanSitesForMaintenanceInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"domain", "stale_threshold_days"},
	Defaults: map[string]interface{}{
		"stale_threshold_days": 30,
	},
	Deprecated: map[string]string{
		"domain_field": "domain",
	},
}

// PrepareRebuildDispatchesInputSpec declares inputs for PrepareRebuildDispatchesAction.
// This action reads from the maintenance_queue, not from collectedData.
// Config settings (task_type, max_tasks, flag_pages) are pure config, but we
// declare them as Optional with Defaults so callers CAN override via input_data.
var PrepareRebuildDispatchesInputSpec = datahelpers.ActionInputSpec{
	Required: []string{},
	Optional: []string{"task_type", "max_tasks", "flag_pages"},
	Defaults: map[string]interface{}{
		"task_type":  "page_rebuild",
		"max_tasks":  50,
		"flag_pages": true,
	},
	Deprecated: map[string]string{},
}

// MarkMaintenanceCompleteInputSpec declares inputs for MarkMaintenanceCompleteAction.
// task_ids comes from current_dispatch.task_ids (loop variable in triage workflow).
// result comes from the specialist agent's output (e.g. rebuild_result).
var MarkMaintenanceCompleteInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"task_ids"},
	Optional: []string{"result"},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"task_ids_field": "task_ids",
		"result_field":   "result",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("load_site_for_rebuild", LoadSiteForRebuildInputSpec)
	datahelpers.RegisterActionInputSpec("scan_sites_for_maintenance", ScanSitesForMaintenanceInputSpec)
	datahelpers.RegisterActionInputSpec("prepare_rebuild_dispatches", PrepareRebuildDispatchesInputSpec)
	datahelpers.RegisterActionInputSpec("mark_maintenance_complete", MarkMaintenanceCompleteInputSpec)
}

// ============================================================================
// ACTION: load_site_for_rebuild
// Used by: page-rebuild agent
// ============================================================================

// LoadSiteForRebuildAction loads the supplementary context needed by
// page-content-writer when rebuilding pages on an existing site.
//
// Expects in collectedData:
//   - site_id (from ensure_site_record → site_record.site_id)
//   - site_record.content_data (from ensure_site_record)
//
// Returns:
//   - reviewed_brief, site_plan, db_sync, logo_url, hero_url, site_id
func LoadSiteForRebuildAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("LoadSiteForRebuildAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required for rebuild context loading")
	}

	// --- Standardised input extraction ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadSiteForRebuildInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", siteIDStr, err)
	}
	taskID := inputs.Get("task_id")

	// --- Extract content_data from collectedData ---
	// ensure_site_record returns content_data as site_record.content_data.
	// This needs custom handling (DB fallback) beyond what the spec covers.
	contentData, _ := datahelpers.ExtractNestedField(params.CollectedData, "site_record.content_data").(map[string]interface{})
	if contentData == nil {
		logger.Warn("LoadSiteForRebuildAction: content_data not in collectedData, loading from DB")
		contentData, err = loadContentDataFromDB(ctx, params.DB, siteID)
		if err != nil {
			return nil, fmt.Errorf("failed to load content_data: %w", err)
		}
	}

	// --- Build reviewed_brief from content_data ---
	// During original build, store_reviewed_brief merged brief fields into content_data
	// at the top level. The content writer expects reviewed_brief as a map with these fields.
	reviewedBrief := contentData

	// --- Ensure contact fields from sites table columns ---
	// content_data may not have email/phone if they were only stored in the
	// sites columns. Load them and inject if not already present.
	var siteEmail, sitePhone, siteCompany, siteTagline sql.NullString
	_ = params.DB.QueryRowContext(ctx,
		`SELECT COALESCE(email, ''), COALESCE(phone, ''), COALESCE(company_name, ''), COALESCE(tagline, '')
		 FROM sites WHERE id = $1`, siteID,
	).Scan(&siteEmail, &sitePhone, &siteCompany, &siteTagline)

	if siteEmail.String != "" {
		if existing, _ := reviewedBrief["contact_email"].(string); existing == "" {
			reviewedBrief["contact_email"] = siteEmail.String
		}
		if existing, _ := reviewedBrief["email"].(string); existing == "" {
			reviewedBrief["email"] = siteEmail.String
		}
	}
	if sitePhone.String != "" {
		if existing, _ := reviewedBrief["contact_phone"].(string); existing == "" {
			reviewedBrief["contact_phone"] = sitePhone.String
		}
		if existing, _ := reviewedBrief["phone"].(string); existing == "" {
			reviewedBrief["phone"] = sitePhone.String
		}
	}
	if siteCompany.String != "" {
		if existing, _ := reviewedBrief["company_name"].(string); existing == "" {
			reviewedBrief["company_name"] = siteCompany.String
		}
	}
	if siteTagline.String != "" {
		if existing, _ := reviewedBrief["tagline"].(string); existing == "" {
			reviewedBrief["tagline"] = siteTagline.String
		}
	}

	// --- Build site_plan from content_data ---
	// store_site_plan also merged plan data into content_data.
	// The content writer uses site_plan for page structure context.
	sitePlan := contentData

	// --- Extract brand asset URLs ---
	logoURL, _ := contentData["logo_url"].(string)
	heroURL, _ := contentData["hero_url"].(string)

	// --- Load all active pages for link context (db_sync.pages) ---
	allPages, err := loadActivePagesForLinkContext(ctx, params.DB, siteID, logger)
	if err != nil {
		logger.Warn("LoadSiteForRebuildAction: Failed to load pages for link context", zap.Error(err))
		allPages = []interface{}{}
	}

	// --- Load navigation structure ---
	var navResult interface{}
	nav, err := GetNavigationStructure(ctx, params.DB, siteID, "header", logger)
	if err != nil {
		logger.Warn("LoadSiteForRebuildAction: Failed to load navigation", zap.Error(err))
		navResult = map[string]interface{}{"items": []interface{}{}}
	} else {
		// Convert to map for JSON compatibility in collected_data
		navJSON, _ := json.Marshal(nav)
		var navMap map[string]interface{}
		json.Unmarshal(navJSON, &navMap)
		navResult = navMap
	}

	// --- Build db_sync equivalent ---
	// This mimics what sync_pages_to_db returns, providing the navigation
	// and pages list that prepare_link_context expects
	dbSync := map[string]interface{}{
		"pages_synced": len(allPages),
		"navigation":   navResult,
		"pages":        allPages,
		"site_id":      siteIDStr,
		"db_available": true,
	}

	logger.Info("LoadSiteForRebuildAction: Context loaded",
		zap.String("site_id", siteIDStr),
		zap.Int("pages_for_linking", len(allPages)),
		zap.Bool("has_logo", logoURL != ""),
		zap.Bool("has_hero", heroURL != ""),
		zap.String("task_id", taskID),
	)

	result := map[string]interface{}{
		"reviewed_brief": reviewedBrief,
		"site_plan":      sitePlan,
		"db_sync":        dbSync,
		"logo_url":       logoURL,
		"hero_url":       heroURL,
		"site_id":        siteIDStr,
	}
	if taskID != "" {
		result["task_id"] = taskID
	}
	return result, nil
}

// ============================================================================
// ACTION: scan_sites_for_maintenance
// Used by: maintenance-triage agent
// ============================================================================

// ScanSitesForMaintenanceAction scans deployed sites for maintenance issues
// and inserts findings into the maintenance_queue.
//
// Data inputs (via ActionInputSpec):
//   - domain (optional) — scope to one site
//   - stale_threshold_days (optional, default 30)
//
// Config-only settings (not data inputs):
//   - checks: []string — which checks to run: "stale_pages", "missing_content", "orphan_nav"
//   - deduplicate: bool (default true) — skip if pending task already exists
func ScanSitesForMaintenanceAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("ScanSitesForMaintenanceAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Standardised input extraction ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		ScanSitesForMaintenanceInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	domain := inputs.Get("domain")
	thresholdDays := inputs.GetInt("stale_threshold_days", 30)

	// --- Config-only settings (not data inputs, no ActionInputSpec) ---
	config := params.StepConfig.Config
	deduplicate := datahelpers.GetBoolField(config, "deduplicate", true)

	checks := []string{"stale_pages", "missing_content", "orphan_nav"}
	if configChecks, ok := config["checks"].([]interface{}); ok && len(configChecks) > 0 {
		checks = make([]string, len(configChecks))
		for i, c := range configChecks {
			checks[i] = fmt.Sprintf("%v", c)
		}
	}

	// --- Load sites to scan ---
	sites, err := loadSitesToScan(ctx, params.DB, domain, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to load sites: %w", err)
	}

	logger.Info("ScanSitesForMaintenanceAction: Sites loaded",
		zap.Int("site_count", len(sites)),
		zap.String("domain_filter", domain),
		zap.Strings("checks", checks),
		zap.Int("stale_threshold_days", thresholdDays),
	)

	// --- Run checks per site ---
	var findings []interface{}
	totalIssues := 0
	totalInserted := 0
	totalSkipped := 0

	for _, site := range sites {
		siteID := site.id
		siteDomain := site.domain

		var siteIssues []interface{}

		if containsCheck(checks, "stale_pages") {
			stalePages, err := findStalePages(ctx, params.DB, siteID, thresholdDays)
			if err != nil {
				logger.Warn("Stale pages check failed", zap.Error(err), zap.String("domain", siteDomain))
			} else if len(stalePages) > 0 {
				siteIssues = append(siteIssues, map[string]interface{}{
					"check":   "stale_pages",
					"pages":   stalePages,
					"details": fmt.Sprintf("not updated in %d+ days", thresholdDays),
				})
			}
		}

		if containsCheck(checks, "missing_content") {
			emptyPages, err := findPagesWithNoContent(ctx, params.DB, siteID)
			if err != nil {
				logger.Warn("Missing content check failed", zap.Error(err), zap.String("domain", siteDomain))
			} else if len(emptyPages) > 0 {
				siteIssues = append(siteIssues, map[string]interface{}{
					"check":   "missing_content",
					"pages":   emptyPages,
					"details": "0 page_components stored",
				})
			}
		}

		if containsCheck(checks, "orphan_nav") {
			orphanPages, err := findOrphanNavPages(ctx, params.DB, siteID)
			if err != nil {
				logger.Warn("Orphan nav check failed", zap.Error(err), zap.String("domain", siteDomain))
			} else if len(orphanPages) > 0 {
				siteIssues = append(siteIssues, map[string]interface{}{
					"check":   "orphan_nav",
					"pages":   orphanPages,
					"details": "in nav but build_status != deployed",
				})
			}
		}

		allAffectedPages := collectUniquePages(siteIssues)
		taskInserted := false

		if len(allAffectedPages) > 0 {
			totalIssues += len(allAffectedPages)

			inserted, err := insertMaintenanceTask(ctx, params.DB, siteID, "page_rebuild", allAffectedPages, deduplicate, logger)
			if err != nil {
				logger.Warn("Failed to insert maintenance task",
					zap.Error(err), zap.String("domain", siteDomain))
			} else if inserted {
				totalInserted++
				taskInserted = true
			} else {
				totalSkipped++
			}
		}

		findings = append(findings, map[string]interface{}{
			"site_id":        siteID.String(),
			"domain":         siteDomain,
			"issues":         siteIssues,
			"pages_affected": len(allAffectedPages),
			"task_inserted":  taskInserted,
		})
	}

	logger.Info("ScanSitesForMaintenanceAction: Complete",
		zap.Int("sites_scanned", len(sites)),
		zap.Int("issues_found", totalIssues),
		zap.Int("tasks_inserted", totalInserted),
		zap.Int("tasks_skipped", totalSkipped),
	)

	return map[string]interface{}{
		"sites_scanned":  len(sites),
		"issues_found":   totalIssues,
		"tasks_inserted": totalInserted,
		"tasks_skipped":  totalSkipped,
		"findings":       findings,
	}, nil
}

// ============================================================================
// ACTION: prepare_rebuild_dispatches
// Used by: maintenance-triage agent
// ============================================================================

// PrepareRebuildDispatchesAction claims page_rebuild tasks from the
// maintenance_queue, flags the affected pages as needs_rebuild, and
// groups dispatches by site so each page-rebuild call covers one site.
//
// Data inputs (via ActionInputSpec, all optional with defaults):
//   - task_type (default "page_rebuild")
//   - max_tasks (default 50)
//   - flag_pages (default true)
func PrepareRebuildDispatchesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("PrepareRebuildDispatchesAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Standardised input extraction ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		PrepareRebuildDispatchesInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	taskType := inputs.Get("task_type")
	if taskType == "" {
		taskType = "page_rebuild" // safety fallback beyond spec default
	}
	maxTasks := inputs.GetInt("max_tasks", 50)
	flagPages := inputs.GetBool("flag_pages", true)

	agentID := params.ExecutionContext.FromAgentID
	if agentID == "" {
		agentID = "maintenance-triage"
	}

	// --- Claim pending tasks ---
	tasks, err := claimPendingTasks(ctx, params.DB, taskType, agentID, maxTasks, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to claim tasks: %w", err)
	}

	if len(tasks) == 0 {
		logger.Info("PrepareRebuildDispatchesAction: No pending tasks")
		return map[string]interface{}{
			"dispatch_count":      0,
			"total_pages_flagged": 0,
			"dispatches":          []interface{}{},
		}, nil
	}

	// --- Group tasks by site_id ---
	type siteDispatch struct {
		domain  string
		siteID  uuid.UUID
		taskIDs []string
		pages   []string
	}
	dispatchMap := make(map[uuid.UUID]*siteDispatch)

	for _, task := range tasks {
		sd, exists := dispatchMap[task.siteID]
		if !exists {
			sd = &siteDispatch{
				siteID: task.siteID,
				domain: task.domain,
			}
			dispatchMap[task.siteID] = sd
		}
		sd.taskIDs = append(sd.taskIDs, task.id.String())
		sd.pages = append(sd.pages, task.pages...)
	}

	// --- Deduplicate page lists and flag pages ---
	var dispatches []interface{}
	totalFlagged := 0

	for _, sd := range dispatchMap {
		uniquePages := deduplicateStrings(sd.pages)

		if flagPages && len(uniquePages) > 0 {
			flagged, err := flagPagesForRebuild(ctx, params.DB, sd.siteID, uniquePages, logger)
			if err != nil {
				logger.Warn("Failed to flag pages",
					zap.Error(err), zap.String("domain", sd.domain))
			}
			totalFlagged += flagged
		}

		dispatches = append(dispatches, map[string]interface{}{
			"domain":        sd.domain,
			"site_id":       sd.siteID.String(),
			"task_ids":      sd.taskIDs,
			"pages_flagged": uniquePages,
		})
	}

	logger.Info("PrepareRebuildDispatchesAction: Complete",
		zap.Int("dispatch_count", len(dispatches)),
		zap.Int("total_pages_flagged", totalFlagged),
		zap.Int("tasks_claimed", len(tasks)),
	)

	return map[string]interface{}{
		"dispatch_count":      len(dispatches),
		"total_pages_flagged": totalFlagged,
		"dispatches":          dispatches,
	}, nil
}

// ============================================================================
// ACTION: mark_maintenance_complete
// Used by: maintenance-triage agent (after specialist finishes)
// ============================================================================

// MarkMaintenanceCompleteAction marks maintenance_queue tasks as complete
// or failed based on the specialist agent's result.
//
// Data inputs (via ActionInputSpec):
//   - task_ids (required) — []string of queue task UUIDs from current_dispatch
//   - result (optional) — specialist agent output for recording
func MarkMaintenanceCompleteAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("MarkMaintenanceCompleteAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Standardised input extraction ---
	// Note: task_ids is Required. If ExtractActionInputs can't find it,
	// the deprecated task_ids_field path will be tried automatically.
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		MarkMaintenanceCompleteInputSpec,
		logger,
	)
	if err != nil {
		// task_ids is required — but in a loop context it might be nested.
		// Log but don't hard-fail if we can recover below.
		logger.Warn("MarkMaintenanceCompleteAction: input extraction issue", zap.Error(err))
	}

	// --- Extract task_ids (handle both []string and []interface{}) ---
	var taskIDs []string
	if inputs != nil {
		taskIDsRaw, _ := inputs.Values["task_ids"]
		switch v := taskIDsRaw.(type) {
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok {
					taskIDs = append(taskIDs, s)
				}
			}
		case []string:
			taskIDs = v
		}
	}

	if len(taskIDs) == 0 {
		logger.Warn("MarkMaintenanceCompleteAction: No task IDs found")
		return map[string]interface{}{
			"tasks_completed": 0,
			"tasks_failed":    0,
		}, nil
	}

	// --- Check if specialist reported success or failure ---
	var resultData map[string]interface{}
	if inputs != nil {
		resultData = inputs.GetMap("result")
	}

	isSuccess := true
	var resultJSON []byte
	if resultData != nil {
		if errMsg, ok := resultData["error"].(string); ok && errMsg != "" {
			isSuccess = false
		}
		if success, ok := resultData["success"].(bool); ok && !success {
			isSuccess = false
		}
		resultJSON, _ = json.Marshal(resultData)
	} else {
		resultJSON = []byte("{}")
	}

	// --- Mark tasks ---
	completed := 0
	failed := 0

	for _, taskIDStr := range taskIDs {
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			logger.Warn("Invalid task ID", zap.String("task_id", taskIDStr))
			continue
		}

		if isSuccess {
			_, err = params.DB.ExecContext(ctx,
				`SELECT complete_maintenance_task($1, $2)`,
				taskID, resultJSON,
			)
			if err != nil {
				logger.Warn("Failed to complete task",
					zap.Error(err), zap.String("task_id", taskIDStr))
				failed++
			} else {
				completed++
			}
		} else {
			errMsg := "specialist reported failure"
			if resultData != nil {
				if e, ok := resultData["error"].(string); ok {
					errMsg = e
				}
			}
			_, err = params.DB.ExecContext(ctx,
				`SELECT fail_maintenance_task($1, $2)`,
				taskID, errMsg,
			)
			if err != nil {
				logger.Warn("Failed to mark task as failed",
					zap.Error(err), zap.String("task_id", taskIDStr))
			}
			failed++
		}
	}

	logger.Info("MarkMaintenanceCompleteAction: Complete",
		zap.Int("tasks_completed", completed),
		zap.Int("tasks_failed", failed),
	)

	return map[string]interface{}{
		"tasks_completed": completed,
		"tasks_failed":    failed,
	}, nil
}

// ============================================================================
// HELPERS — shared by the actions above
// ============================================================================

// maintenanceSiteRecord holds minimal site info for scanning
type maintenanceSiteRecord struct {
	id     uuid.UUID
	domain string
}

// maintenanceTask holds a claimed task from the queue
type maintenanceTask struct {
	id     uuid.UUID
	siteID uuid.UUID
	domain string
	pages  []string
}

// loadSitesToScan queries sites to check. If domain is provided,
// returns just that site. Otherwise returns all deployed sites.
func loadSitesToScan(ctx context.Context, db *sql.DB, domain string, logger *zap.Logger) ([]maintenanceSiteRecord, error) {
	var rows *sql.Rows
	var err error

	if domain != "" {
		rows, err = db.QueryContext(ctx,
			`SELECT id, domain FROM sites WHERE domain = $1 AND status = 'deployed'`, domain)
	} else {
		rows, err = db.QueryContext(ctx,
			`SELECT id, domain FROM sites WHERE status = 'deployed' ORDER BY domain`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sites []maintenanceSiteRecord
	for rows.Next() {
		var s maintenanceSiteRecord
		if err := rows.Scan(&s.id, &s.domain); err != nil {
			logger.Warn("Failed to scan site row", zap.Error(err))
			continue
		}
		sites = append(sites, s)
	}
	return sites, nil
}

// findStalePages returns page names where deployed_at is older than thresholdDays.
func findStalePages(ctx context.Context, db *sql.DB, siteID uuid.UUID, thresholdDays int) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name FROM pages
		WHERE site_id = $1
		  AND build_status = 'deployed'
		  AND status = 'active'
		  AND deployed_at < NOW() - make_interval(days => $2)
		ORDER BY deployed_at ASC
	`, siteID, thresholdDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			pages = append(pages, name)
		}
	}
	return pages, nil
}

// findPagesWithNoContent returns deployed page names that have zero page_components.
func findPagesWithNoContent(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT p.name
		FROM pages p
		LEFT JOIN page_components pc ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND p.build_status = 'deployed'
		  AND p.status = 'active'
		GROUP BY p.id, p.name
		HAVING COUNT(pc.id) = 0
		ORDER BY p.name
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			pages = append(pages, name)
		}
	}
	return pages, nil
}

// findOrphanNavPages returns page names that appear in nav but aren't deployed.
func findOrphanNavPages(ctx context.Context, db *sql.DB, siteID uuid.UUID) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name FROM pages
		WHERE site_id = $1
		  AND (in_header = true OR in_footer = true)
		  AND build_status NOT IN ('deployed')
		  AND status = 'active'
		ORDER BY name
	`, siteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pages []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			pages = append(pages, name)
		}
	}
	return pages, nil
}

// collectUniquePages extracts all unique page names from issue findings.
func collectUniquePages(issues []interface{}) []string {
	seen := make(map[string]bool)
	var result []string
	for _, issue := range issues {
		issueMap, ok := issue.(map[string]interface{})
		if !ok {
			continue
		}
		pages, ok := issueMap["pages"].([]string)
		if !ok {
			continue
		}
		for _, p := range pages {
			if !seen[p] {
				seen[p] = true
				result = append(result, p)
			}
		}
	}
	return result
}

// insertMaintenanceTask inserts a page_rebuild task into the maintenance_queue.
// If deduplicate is true, skips insert when a pending/claimed task already exists
// for the same site_id and task_type.
// Returns true if a row was inserted.
func insertMaintenanceTask(ctx context.Context, db *sql.DB, siteID uuid.UUID, taskType string, pages []string, deduplicate bool, logger *zap.Logger) (bool, error) {
	payload := map[string]interface{}{
		"pages":       pages,
		"detected_at": time.Now().UTC().Format(time.RFC3339),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	if deduplicate {
		var exists bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS(
				SELECT 1 FROM maintenance_queue
				WHERE site_id = $1 AND task_type = $2 AND status IN ('pending', 'claimed')
			)
		`, siteID, taskType).Scan(&exists)
		if err != nil {
			return false, fmt.Errorf("dedup check failed: %w", err)
		}
		if exists {
			logger.Info("Task already pending, skipping",
				zap.String("site_id", siteID.String()),
				zap.String("task_type", taskType))
			return false, nil
		}
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO maintenance_queue (site_id, task_type, priority, reason, payload, requested_by)
		VALUES ($1, $2, 5, 'automated_scan', $3, 'maintenance-triage')
	`, siteID, taskType, payloadJSON)
	if err != nil {
		return false, fmt.Errorf("insert failed: %w", err)
	}

	logger.Info("Maintenance task inserted",
		zap.String("site_id", siteID.String()),
		zap.String("task_type", taskType),
		zap.Int("pages", len(pages)),
	)
	return true, nil
}

// claimPendingTasks claims up to maxTasks pending tasks of the given type.
// Uses FOR UPDATE SKIP LOCKED for concurrency-safe claiming.
func claimPendingTasks(ctx context.Context, db *sql.DB, taskType string, agentID string, maxTasks int, logger *zap.Logger) ([]maintenanceTask, error) {
	rows, err := db.QueryContext(ctx, `
		UPDATE maintenance_queue
		SET status = 'claimed',
		    claimed_by = $1,
		    claimed_at = NOW(),
		    updated_at = NOW()
		WHERE id IN (
			SELECT id FROM maintenance_queue
			WHERE status = 'pending'
			  AND task_type = $2
			  AND retry_count < max_retries
			ORDER BY priority ASC, created_at ASC
			LIMIT $3
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, site_id, payload
	`, agentID, taskType, maxTasks)
	if err != nil {
		return nil, fmt.Errorf("claim query failed: %w", err)
	}
	defer rows.Close()

	type rawTask struct {
		id      uuid.UUID
		siteID  uuid.UUID
		payload []byte
	}
	var rawTasks []rawTask

	for rows.Next() {
		var t rawTask
		if err := rows.Scan(&t.id, &t.siteID, &t.payload); err != nil {
			logger.Warn("Failed to scan claimed task", zap.Error(err))
			continue
		}
		rawTasks = append(rawTasks, t)
	}

	if len(rawTasks) == 0 {
		return nil, nil
	}

	// Look up domains for all site_ids
	domainCache := make(map[uuid.UUID]string)
	for _, t := range rawTasks {
		if _, ok := domainCache[t.siteID]; !ok {
			var domain string
			err := db.QueryRowContext(ctx,
				`SELECT domain FROM sites WHERE id = $1`, t.siteID,
			).Scan(&domain)
			if err != nil {
				logger.Warn("Failed to look up domain for site",
					zap.Error(err), zap.String("site_id", t.siteID.String()))
				domain = "unknown"
			}
			domainCache[t.siteID] = domain
		}
	}

	// Parse payloads and build result
	var tasks []maintenanceTask
	for _, rt := range rawTasks {
		mt := maintenanceTask{
			id:     rt.id,
			siteID: rt.siteID,
			domain: domainCache[rt.siteID],
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(rt.payload, &payload); err == nil {
			if pagesRaw, ok := payload["pages"].([]interface{}); ok {
				for _, p := range pagesRaw {
					if s, ok := p.(string); ok {
						mt.pages = append(mt.pages, s)
					}
				}
			}
		}

		tasks = append(tasks, mt)
		logger.Info("Claimed maintenance task",
			zap.String("task_id", mt.id.String()),
			zap.String("domain", mt.domain),
			zap.Int("pages", len(mt.pages)),
		)
	}

	return tasks, nil
}

// flagPagesForRebuild sets build_status = 'needs_rebuild' on the specified pages.
func flagPagesForRebuild(ctx context.Context, db *sql.DB, siteID uuid.UUID, pageNames []string, logger *zap.Logger) (int, error) {
	if len(pageNames) == 0 {
		return 0, nil
	}

	args := make([]interface{}, 0, len(pageNames)+1)
	args = append(args, siteID)
	placeholders := make([]string, len(pageNames))
	for i, name := range pageNames {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, name)
	}

	query := fmt.Sprintf(`
		UPDATE pages
		SET build_status = 'needs_rebuild', updated_at = NOW()
		WHERE site_id = $1
		  AND name IN (%s)
		  AND status = 'active'
	`, strings.Join(placeholders, ", "))

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to flag pages: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Info("Flagged pages for rebuild",
		zap.String("site_id", siteID.String()),
		zap.Int64("pages_flagged", rowsAffected),
		zap.Strings("page_names", pageNames),
	)
	return int(rowsAffected), nil
}

// deduplicateStrings returns unique strings preserving order.
func deduplicateStrings(input []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range input {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// containsCheck returns true if the checks slice contains the given check name.
func containsCheck(checks []string, name string) bool {
	for _, c := range checks {
		if c == name {
			return true
		}
	}
	return false
}

// loadContentDataFromDB loads content_data directly from sites table.
// Fallback when ensure_site_record didn't provide it in collectedData.
func loadContentDataFromDB(ctx context.Context, db *sql.DB, siteID uuid.UUID) (map[string]interface{}, error) {
	var contentJSON []byte
	err := db.QueryRowContext(ctx,
		`SELECT COALESCE(content_data, '{}'::jsonb) FROM sites WHERE id = $1`,
		siteID,
	).Scan(&contentJSON)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(contentJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal content_data: %w", err)
	}
	return result, nil
}

// loadActivePagesForLinkContext queries all active pages for a site.
// Returns them in the format that extractPagesForLinking expects:
// []interface{} of map[string]interface{} with name, url, title, description.
func loadActivePagesForLinkContext(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]interface{}, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT name, url, COALESCE(title, name) as title,
		       COALESCE(meta_description, '') as description,
		       COALESCE(nav_order, 100) as nav_order
		FROM pages
		WHERE site_id = $1 AND status = 'active'
		ORDER BY nav_order ASC, name ASC
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}
	defer rows.Close()

	var pages []interface{}
	for rows.Next() {
		var name, url, title, description string
		var navOrder int
		if err := rows.Scan(&name, &url, &title, &description, &navOrder); err != nil {
			logger.Warn("loadActivePagesForLinkContext: Failed to scan page", zap.Error(err))
			continue
		}
		// Build URL from name if missing (same logic as extractPagesForLinking)
		if url == "" {
			if name == "index" || name == "home" {
				url = "/index.html"
			} else {
				url = "/" + name + ".html"
			}
		}
		// Title fallback
		if title == name {
			title = strings.Title(strings.ReplaceAll(name, "-", " "))
		}
		pages = append(pages, map[string]interface{}{
			"name":        name,
			"url":         url,
			"title":       title,
			"description": description,
			"nav_order":   navOrder,
		})
	}

	logger.Info("loadActivePagesForLinkContext: Loaded pages",
		zap.Int("count", len(pages)),
		zap.String("site_id", siteID.String()),
	)
	return pages, nil
}
