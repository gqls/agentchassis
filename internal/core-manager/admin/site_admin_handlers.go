// FILE: internal/core-manager/admin/site_admin_handlers.go
//
// Admin endpoints for the site-building domain:
//   - Work items: list, get, update, retry, resolve
//   - Site specs: get, update (versioned)
//   - Sites: list with dashboard stats, detail with specs

package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SiteAdminHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSiteAdminHandlers(db *sql.DB, logger *zap.Logger) *SiteAdminHandlers {
	return &SiteAdminHandlers{db: db, logger: logger}
}

// ============================================================================
// GET /admin/sites
// ============================================================================

func (h *SiteAdminHandlers) HandleListSites(c *gin.Context) {
	rows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT s.id, s.domain, s.company_name, s.status, s.email, s.phone,
		       s.build_status, s.last_deployed_at,
		       s.locked_at, COALESCE(s.locked_by, ''),
		       COUNT(*) FILTER (WHERE wi.status = 'complete') as items_done,
		       COUNT(*) FILTER (WHERE wi.status = 'claimed') as items_active,
		       COUNT(*) FILTER (WHERE wi.status = 'triaged') as items_ready,
		       COUNT(*) FILTER (WHERE wi.status = 'needs_human_review') as items_review,
		       COUNT(*) FILTER (WHERE wi.status = 'failed') as items_failed,
		       COUNT(*) FILTER (WHERE wi.status = 'blocked') as items_blocked
		FROM sites s
		LEFT JOIN site_work_items wi ON wi.site_id = s.id AND wi.pipeline = 'build'
		WHERE s.status IN ('active', 'deployed')
		GROUP BY s.id
		ORDER BY s.domain
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var sites []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var domain, status, buildStatus, lockedBy string
		var companyName, email, phone sql.NullString
		var lastDeployed, lockedAt sql.NullTime
		var done, active, ready, review, failed, blocked int

		if err := rows.Scan(&id, &domain, &companyName, &status, &email, &phone,
			&buildStatus, &lastDeployed, &lockedAt, &lockedBy,
			&done, &active, &ready, &review, &failed, &blocked); err != nil {
			h.logger.Warn("Failed to scan site", zap.Error(err))
			continue
		}

		site := map[string]interface{}{
			"id":            id.String(),
			"domain":        domain,
			"company_name":  companyName.String,
			"status":        status,
			"email":         email.String,
			"phone":         phone.String,
			"build_status":  buildStatus,
			"last_deployed": nil,
			"locked":        lockedAt.Valid,
			"locked_by":     lockedBy,
			"work_items": map[string]int{
				"done":    done,
				"active":  active,
				"ready":   ready,
				"review":  review,
				"failed":  failed,
				"blocked": blocked,
			},
		}
		if lastDeployed.Valid {
			site["last_deployed"] = lastDeployed.Time
		}
		if lockedAt.Valid {
			site["locked_at"] = lockedAt.Time
		}
		sites = append(sites, site)
	}

	c.JSON(http.StatusOK, gin.H{"sites": sites})
}

// ============================================================================
// GET /admin/sites/:site_id
// ============================================================================

func (h *SiteAdminHandlers) HandleGetSite(c *gin.Context) {
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	var domain, status, buildStatus string
	var companyName, tagline, email, phone, logoText sql.NullString
	var contentDataJSON []byte
	var lastDeployed sql.NullTime

	err = h.db.QueryRowContext(c.Request.Context(), `
		SELECT domain, status, build_status, company_name, tagline,
		       email, phone, logo_text, COALESCE(content_data, '{}'::jsonb),
		       last_deployed_at
		FROM sites WHERE id = $1
	`, siteID).Scan(&domain, &status, &buildStatus, &companyName, &tagline,
		&email, &phone, &logoText, &contentDataJSON, &lastDeployed)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		return
	}

	specRows, err := h.db.QueryContext(c.Request.Context(), `
		SELECT aspect, data, source, created_at
		FROM site_specs
		WHERE site_id = $1 AND is_current = true
		ORDER BY aspect
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer specRows.Close()

	specs := map[string]interface{}{}
	for specRows.Next() {
		var aspect, source string
		var dataJSON []byte
		var createdAt sql.NullTime
		if err := specRows.Scan(&aspect, &dataJSON, &source, &createdAt); err != nil {
			continue
		}
		var data interface{}
		json.Unmarshal(dataJSON, &data)
		specs[aspect] = map[string]interface{}{
			"data":       data,
			"source":     source,
			"created_at": createdAt.Time,
		}
	}

	site := map[string]interface{}{
		"id":            siteID.String(),
		"domain":        domain,
		"status":        status,
		"build_status":  buildStatus,
		"company_name":  companyName.String,
		"tagline":       tagline.String,
		"email":         email.String,
		"phone":         phone.String,
		"logo_text":     logoText.String,
		"last_deployed": nil,
		"specs":         specs,
	}
	if lastDeployed.Valid {
		site["last_deployed"] = lastDeployed.Time
	}

	c.JSON(http.StatusOK, site)
}

// ============================================================================
// PATCH /admin/sites/:site_id/specs/:aspect
// ============================================================================

func (h *SiteAdminHandlers) HandleUpdateSiteSpec(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	aspect := c.Param("aspect")
	if aspect == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "aspect required"})
		return
	}

	var body struct {
		Data json.RawMessage `json:"data" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		UPDATE site_specs
		SET is_current = false, superseded_at = NOW()
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var newID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current)
		VALUES ($1, $2, $3, 'admin-api', 'admin', true)
		RETURNING id
	`, siteID, aspect, body.Data).Scan(&newID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Site spec updated via admin API",
		zap.String("site_id", siteID.String()),
		zap.String("aspect", aspect),
		zap.String("new_spec_id", newID.String()))

	c.JSON(http.StatusOK, gin.H{"id": newID.String(), "aspect": aspect, "updated": true})
}

// ============================================================================
// PATCH /admin/sites/:site_id
// ============================================================================

func (h *SiteAdminHandlers) HandleUpdateSite(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	var body struct {
		CompanyName    *string `json:"company_name"`
		Tagline        *string `json:"tagline"`
		Email          *string `json:"email"`
		Phone          *string `json:"phone"`
		ContactAddress *string `json:"contact_address"`
		LogoText       *string `json:"logo_text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if body.CompanyName != nil {
		setClauses = append(setClauses, fmt.Sprintf("company_name = $%d", argIdx))
		args = append(args, *body.CompanyName)
		argIdx++
	}
	if body.Tagline != nil {
		setClauses = append(setClauses, fmt.Sprintf("tagline = $%d", argIdx))
		args = append(args, *body.Tagline)
		argIdx++
	}
	if body.Email != nil {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *body.Email)
		argIdx++
	}
	if body.Phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *body.Phone)
		argIdx++
	}
	if body.ContactAddress != nil {
		setClauses = append(setClauses, fmt.Sprintf("contact_address = $%d", argIdx))
		args = append(args, *body.ContactAddress)
		argIdx++
	}
	if body.LogoText != nil {
		setClauses = append(setClauses, fmt.Sprintf("logo_text = $%d", argIdx))
		args = append(args, *body.LogoText)
		argIdx++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	query := fmt.Sprintf("UPDATE sites SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, siteID)

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		return
	}

	h.logger.Info("Site updated via admin API",
		zap.String("site_id", siteID.String()))

	c.JSON(http.StatusOK, gin.H{"updated": true, "id": siteID.String()})
}

// ============================================================================
// POST /admin/sites/:site_id/lock
// ============================================================================
// Locks a site — discovery checks and dispatch loop skip it entirely.
// Admin can still manually create and process work items.

func (h *SiteAdminHandlers) HandleLockSite(c *gin.Context) {
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	result, err := h.db.ExecContext(c.Request.Context(), `
		UPDATE sites SET locked_at = NOW(), locked_by = 'admin'
		WHERE id = $1
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		return
	}

	h.logger.Info("Site locked via admin", zap.String("site_id", siteID.String()))
	c.JSON(http.StatusOK, gin.H{"locked": true, "site_id": siteID.String()})
}

// ============================================================================
// POST /admin/sites/:site_id/unlock
// ============================================================================

func (h *SiteAdminHandlers) HandleUnlockSite(c *gin.Context) {
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	result, err := h.db.ExecContext(c.Request.Context(), `
		UPDATE sites SET locked_at = NULL, locked_by = NULL
		WHERE id = $1
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "site not found"})
		return
	}

	h.logger.Info("Site unlocked via admin", zap.String("site_id", siteID.String()))
	c.JSON(http.StatusOK, gin.H{"unlocked": true, "site_id": siteID.String()})
}

// ============================================================================
// POST /admin/work-items
// ============================================================================

func (h *SiteAdminHandlers) HandleCreateWorkItem(c *gin.Context) {
	ctx := c.Request.Context()

	var body struct {
		SiteID       string           `json:"site_id" binding:"required"`
		ItemType     string           `json:"item_type" binding:"required"`
		Summary      string           `json:"summary" binding:"required"`
		Severity     string           `json:"severity"`
		HandlerAgent string           `json:"handler_agent"`
		PageID       *string          `json:"page_id"`
		Priority     *int             `json:"priority"`
		Spec         *json.RawMessage `json:"spec"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	siteID, err := uuid.Parse(body.SiteID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	severity := "medium"
	if body.Severity != "" {
		severity = body.Severity
	}
	handlerAgent := "page-build-handler"
	if body.HandlerAgent != "" {
		handlerAgent = body.HandlerAgent
	}
	priority := 30
	if body.Priority != nil {
		priority = *body.Priority
	}
	spec := json.RawMessage("{}")
	if body.Spec != nil {
		spec = *body.Spec
	}

	var pageID *uuid.UUID
	if body.PageID != nil && *body.PageID != "" {
		parsed, err := uuid.Parse(*body.PageID)
		if err == nil {
			pageID = &parsed
		}
	}

	var newID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, page_id, priority, handler_agent, status, created_by
		) VALUES ($1, 'admin', 'build', $2, $3, $4, $5::jsonb, $6, $7, $8, 'triaged', 'admin')
		RETURNING id
	`, siteID, body.ItemType, severity, body.Summary,
		string(spec), pageID, priority, handlerAgent).Scan(&newID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Work item created via admin API",
		zap.String("item_id", newID.String()),
		zap.String("site_id", siteID.String()),
		zap.String("item_type", body.ItemType))

	c.JSON(http.StatusCreated, gin.H{"id": newID.String(), "status": "triaged"})
}

// ============================================================================
// GET /admin/work-items
// ============================================================================

func (h *SiteAdminHandlers) HandleListWorkItems(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Query("status")
	pipeline := c.DefaultQuery("pipeline", "build")
	siteIDStr := c.Query("site_id")
	itemType := c.Query("item_type")
	limit := 50

	query := `
		SELECT wi.id, wi.site_id, s.domain, wi.item_type, wi.status,
		       wi.severity, wi.summary, wi.spec,
				wi.handler_agent, wi.attempt_count, wi.max_attempts,
				wi.error, wi.created_at, wi.completed_at
		FROM site_work_items wi
		JOIN sites s ON s.id = wi.site_id
		WHERE wi.pipeline = $1
	`
	args := []interface{}{pipeline}
	argIdx := 2

	if status != "" {
		query += fmt.Sprintf(" AND wi.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	} else {
		query += " AND wi.status != 'complete'"
	}

	if siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			query += fmt.Sprintf(" AND wi.site_id = $%d", argIdx)
			args = append(args, siteID)
			argIdx++
		}
	}

	if itemType != "" {
		query += fmt.Sprintf(" AND wi.item_type = $%d", argIdx)
		args = append(args, itemType)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY wi.created_at DESC LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := h.db.QueryContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, siteID uuid.UUID
		var siteDomain, itemTypeVal, statusVal, severity, summary string
		var handlerAgentNullable sql.NullString
		var attemptCount, maxAttempts int
		var specJSON []byte
		var errorMsg sql.NullString
		var createdAt, completedAt sql.NullTime

		if err := rows.Scan(&id, &siteID, &siteDomain, &itemTypeVal, &statusVal,
			&severity, &summary, &specJSON,
			&handlerAgentNullable, &attemptCount, &maxAttempts,
			&errorMsg, &createdAt, &completedAt); err != nil {
			h.logger.Warn("Failed to scan work item", zap.Error(err))
			continue
		}

		var spec interface{}
		json.Unmarshal(specJSON, &spec)

		item := map[string]interface{}{
			"id":            id.String(),
			"site_id":       siteID.String(),
			"domain":        siteDomain,
			"item_type":     itemTypeVal,
			"status":        statusVal,
			"severity":      severity,
			"summary":       summary,
			"spec":          spec,
			"handler_agent": handlerAgentNullable,
			"attempts":      fmt.Sprintf("%d/%d", attemptCount, maxAttempts),
			"error":         errorMsg.String,
			"created_at":    createdAt.Time,
			"completed_at":  nil,
		}
		if completedAt.Valid {
			item["completed_at"] = completedAt.Time
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "count": len(items)})
}

// ============================================================================
// GET /admin/work-items/:item_id
// ============================================================================

func (h *SiteAdminHandlers) HandleGetWorkItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	var siteID uuid.UUID
	var siteDomain, itemType, status, severity, summary string
	var handlerAgent sql.NullString
	var attemptCount, maxAttempts int
	var specJSON []byte
	var errorMsg sql.NullString
	var createdAt, completedAt sql.NullTime

	err = h.db.QueryRowContext(c.Request.Context(), `
		SELECT wi.site_id, s.domain, wi.item_type, wi.status,
		       wi.severity, wi.summary, wi.spec,
		       wi.handler_agent, wi.attempt_count, wi.max_attempts,
		       wi.error, wi.created_at, wi.completed_at
		FROM site_work_items wi
		JOIN sites s ON s.id = wi.site_id
		WHERE wi.id = $1
	`, itemID).Scan(&siteID, &siteDomain, &itemType, &status,
		&severity, &summary, &specJSON,
		&handlerAgent, &attemptCount, &maxAttempts,
		&errorMsg, &createdAt, &completedAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}

	var spec interface{}
	json.Unmarshal(specJSON, &spec)

	result := gin.H{
		"id":            itemID.String(),
		"site_id":       siteID.String(),
		"domain":        siteDomain,
		"item_type":     itemType,
		"status":        status,
		"severity":      severity,
		"summary":       summary,
		"spec":          spec,
		"handler_agent": handlerAgent,
		"attempts":      fmt.Sprintf("%d/%d", attemptCount, maxAttempts),
		"error":         errorMsg.String,
		"created_at":    createdAt.Time,
		"completed_at":  nil,
	}
	if completedAt.Valid {
		result["completed_at"] = completedAt.Time
	}
	c.JSON(http.StatusOK, result)
}

// ============================================================================
// PATCH /admin/work-items/:item_id
// ============================================================================

func (h *SiteAdminHandlers) HandleUpdateWorkItem(c *gin.Context) {
	ctx := c.Request.Context()
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	var body struct {
		Status       *string          `json:"status"`
		Spec         *json.RawMessage `json:"spec"`
		HandlerAgent *string          `json:"handler_agent"`
		ItemType     *string          `json:"item_type"`
		Error        *string          `json:"error"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if body.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *body.Status)
		argIdx++
		if *body.Status == "triaged" {
			setClauses = append(setClauses, "attempt_count = 0", "claimed_by = NULL", "claimed_at = NULL")
		}
	}
	if body.Spec != nil {
		setClauses = append(setClauses, fmt.Sprintf("spec = $%d", argIdx))
		args = append(args, *body.Spec)
		argIdx++
	}
	if body.HandlerAgent != nil {
		setClauses = append(setClauses, fmt.Sprintf("handler_agent = $%d", argIdx))
		args = append(args, *body.HandlerAgent)
		argIdx++
	}
	if body.ItemType != nil {
		setClauses = append(setClauses, fmt.Sprintf("item_type = $%d", argIdx))
		args = append(args, *body.ItemType)
		argIdx++
	}
	if body.Error != nil {
		setClauses = append(setClauses, fmt.Sprintf("error = $%d", argIdx))
		args = append(args, *body.Error)
		argIdx++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	query := fmt.Sprintf("UPDATE site_work_items SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, itemID)

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}

	h.logger.Info("Work item updated via admin API",
		zap.String("item_id", itemID.String()))

	c.JSON(http.StatusOK, gin.H{"updated": true, "id": itemID.String()})
}

// ============================================================================
// POST /admin/work-items/:item_id/retry
// ============================================================================

func (h *SiteAdminHandlers) HandleRetryWorkItem(c *gin.Context) {
	ctx := c.Request.Context()
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	// Before retrying, resolve any duplicate item_key that would conflict
	// with the dedup index (which excludes failed/complete/etc but includes triaged)
	_, _ = h.db.ExecContext(ctx, `
		UPDATE site_work_items dup
		SET status = 'complete',
		    error = 'Superseded by admin retry of ' || $1::text,
		    completed_at = NOW(),
		    updated_at = NOW()
		FROM site_work_items src
		WHERE src.id = $1
		  AND src.item_key IS NOT NULL
		  AND dup.site_id = src.site_id
		  AND dup.item_key = src.item_key
		  AND dup.id != src.id
		  AND dup.status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed')
	`, itemID)

	result, err := h.db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'triaged',
		    attempt_count = 0,
		    claimed_by = NULL,
		    claimed_at = NULL,
		    error = NULL,
		    updated_at = NOW()
		WHERE id = $1
		  AND status IN ('needs_human_review', 'failed', 'blocked')
	`, itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found or not in retryable status"})
		return
	}

	h.logger.Info("Work item retried via admin API", zap.String("item_id", itemID.String()))
	c.JSON(http.StatusOK, gin.H{"retried": true, "id": itemID.String()})
}

// ============================================================================
// POST /admin/work-items/:item_id/resolve
// ============================================================================

func (h *SiteAdminHandlers) HandleResolveWorkItem(c *gin.Context) {
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	var body struct {
		Resolution string `json:"resolution"`
	}
	c.ShouldBindJSON(&body)
	if body.Resolution == "" {
		body.Resolution = "Resolved via admin API"
	}

	result, err := h.db.ExecContext(c.Request.Context(), `
		UPDATE site_work_items
		SET status = 'complete',
		    completed_at = NOW(),
		    error = NULL,
		    result = jsonb_build_object('resolution', $2, 'resolved_by', 'admin'),
		    updated_at = NOW()
		WHERE id = $1
		  AND status IN ('needs_human_review', 'failed', 'blocked', 'triaged')
	`, itemID, body.Resolution)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found or not in resolvable status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"resolved": true, "id": itemID.String()})
}

// ============================================================================
// POST /admin/work-items/:item_id/approve
// ============================================================================
//
// Handles checkpoint work items created by checkpoint_for_review.
// The admin submits corrected review_data. This endpoint:
//   1. Updates the site_spec with the corrected data (if spec_aspect is set)
//   2. Creates the follow-on work item from on_approve config
//   3. Marks the review item complete

func (h *SiteAdminHandlers) HandleApproveWorkItem(c *gin.Context) {
	ctx := c.Request.Context()
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	// Parse the admin's submission
	var body struct {
		ReviewData json.RawMessage `json:"review_data" binding:"required"`
		Notes      string          `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Load the work item to get its spec (which contains on_approve, spec_aspect, etc.)
	var siteID uuid.UUID
	var specJSON []byte
	var currentStatus string

	err = h.db.QueryRowContext(ctx, `
		SELECT site_id, spec, status
		FROM site_work_items
		WHERE id = $1
	`, itemID).Scan(&siteID, &specJSON, &currentStatus)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}

	if currentStatus != "needs_human_review" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("item status is %s, expected needs_human_review", currentStatus)})
		return
	}

	// Parse the spec
	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse work item spec"})
		return
	}

	// Verify this is a checkpoint item
	isCheckpoint, _ := spec["checkpoint"].(bool)
	if !isCheckpoint {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this item is not a checkpoint — use retry or resolve instead"})
		return
	}

	// ── Step 1: Update site_specs if spec_aspect is set ──────────────────
	specAspect, _ := spec["spec_aspect"].(string)
	if specAspect != "" {
		tx, err := h.db.BeginTx(ctx, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback()

		// Supersede current
		_, err = tx.ExecContext(ctx, `
			UPDATE site_specs
			SET is_current = false, superseded_at = NOW()
			WHERE site_id = $1 AND aspect = $2 AND is_current = true
		`, siteID, specAspect)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to supersede spec: " + err.Error()})
			return
		}

		// Insert approved version
		_, err = tx.ExecContext(ctx, `
			INSERT INTO site_specs (site_id, aspect, data, source, created_by, is_current)
			VALUES ($1, $2, $3, 'admin-approved', 'admin', true)
		`, siteID, specAspect, body.ReviewData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save approved spec: " + err.Error()})
			return
		}

		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		h.logger.Info("Approved spec saved",
			zap.String("site_id", siteID.String()),
			zap.String("aspect", specAspect))
	}

	// ── Step 2: Create follow-on work item if on_approve is set ──────────
	var followOnID *string
	if onApprove, ok := spec["on_approve"].(map[string]interface{}); ok {
		followItemType, _ := onApprove["item_type"].(string)
		followHandler, _ := onApprove["handler_agent"].(string)

		if followItemType == "" {
			followItemType = "approved_for_processing"
		}
		if followHandler == "" {
			followHandler = "page-build-handler"
		}

		// Build follow-on spec with the approved data and any include_fields
		followSpec := map[string]interface{}{
			"approved_data":     json.RawMessage(body.ReviewData),
			"approved_by":       "admin",
			"source_item_id":    itemID.String(),
			"original_pipeline": "build",
		}
		if specAspect != "" {
			followSpec["spec_aspect"] = specAspect
		}

		// Copy include_fields from on_approve config
		if includeFields, ok := onApprove["include_fields"].([]interface{}); ok {
			for _, f := range includeFields {
				if fieldName, ok := f.(string); ok {
					followSpec[fieldName] = spec[fieldName]
				}
			}
		}

		followSpecJSON, _ := json.Marshal(followSpec)

		var newID uuid.UUID
		err = h.db.QueryRowContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, priority, handler_agent, status, created_by
			) VALUES ($1, 'admin-approved', 'build', $2, 'high', $3, $4::jsonb, 10, $5, 'triaged', 'admin')
			RETURNING id
		`, siteID, followItemType,
			fmt.Sprintf("Approved: %s", followItemType),
			string(followSpecJSON), followHandler).Scan(&newID)

		if err != nil {
			h.logger.Warn("Failed to create follow-on work item",
				zap.String("item_id", itemID.String()),
				zap.Error(err))
			// Non-fatal — the spec is already saved
		} else {
			id := newID.String()
			followOnID = &id
			h.logger.Info("Follow-on work item created",
				zap.String("follow_on_id", id),
				zap.String("item_type", followItemType),
				zap.String("handler", followHandler))
		}
	}

	// ── Step 3: Mark review item complete ────────────────────────────────
	resolution := "Approved via admin dashboard"
	if body.Notes != "" {
		resolution = body.Notes
	}

	_, err = h.db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete',
		    completed_at = NOW(),
		    result = jsonb_build_object(
		        'resolution', $2,
		        'approved_by', 'admin',
		        'follow_on_item', $3
		    ),
		    updated_at = NOW()
		WHERE id = $1
	`, itemID, resolution, followOnID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := gin.H{
		"approved": true,
		"id":       itemID.String(),
		"site_id":  siteID.String(),
	}
	if specAspect != "" {
		result["spec_updated"] = specAspect
	}
	if followOnID != nil {
		result["follow_on_item_id"] = *followOnID
	}

	c.JSON(http.StatusOK, result)
}
