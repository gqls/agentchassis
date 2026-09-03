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
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
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
		       COUNT(*) FILTER (WHERE wi.status = 'blocked') as items_blocked,
		       COUNT(*) FILTER (WHERE wi.status = 'unresolved') as items_unresolved
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
		var done, active, ready, review, failed, blocked, unresolved int

		if err := rows.Scan(&id, &domain, &companyName, &status, &email, &phone,
			&buildStatus, &lastDeployed, &lockedAt, &lockedBy,
			&done, &active, &ready, &review, &failed, &blocked, &unresolved); err != nil {
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
				"done":       done,
				"active":     active,
				"ready":      ready,
				"review":     review,
				"failed":     failed,
				"blocked":    blocked,
				"unresolved": unresolved,
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
		// ConfirmEmpty must be true to supersede an evidence_base that parses
		// non-nil with one that parses to nothing scannable — see the guard below.
		ConfirmEmpty bool `json:"confirm_empty"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// evidence_base is the one aspect that is a CONTROL rather than prompt text:
	// its banned_claims/facts are what the claims lanes enforce, read live via
	// datahelpers.ParseEvidenceBase — which returns nil with NO error for
	// well-formed JSON of the wrong shape (a misspelt key, a fragment, one level
	// of extra nesting). Without this guard such a save supersedes the good
	// register, returns 200, and claims checking for the site silently no-ops
	// from then on, with the old register already is_current=false.
	var eb *datahelpers.EvidenceBase
	if aspect == "evidence_base" {
		eb, err = datahelpers.ParseEvidenceBase(body.Data)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "evidence_base does not parse: " + err.Error()})
			return
		}
		if eb == nil && !body.ConfirmEmpty {
			var curData []byte
			scanErr := h.db.QueryRowContext(ctx, `
				SELECT data FROM site_specs
				WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
			`, siteID).Scan(&curData)
			if scanErr != nil && scanErr != sql.ErrNoRows {
				c.JSON(http.StatusInternalServerError, gin.H{"error": scanErr.Error()})
				return
			}
			if scanErr == nil {
				if curEB, _ := datahelpers.ParseEvidenceBase(curData); curEB != nil {
					c.JSON(http.StatusConflict, gin.H{
						"error": fmt.Sprintf(
							"this save would replace an evidence_base holding %d facts and %d banned claims with one that parses to nothing scannable — claims checking for this site would silently stop; resend with confirm_empty=true if that is intended",
							len(curEB.Facts), len(curEB.BannedClaims)),
						"code":                        "EMPTY_EVIDENCE_BASE",
						"current_facts_count":         len(curEB.Facts),
						"current_banned_claims_count": len(curEB.BannedClaims),
					})
					return
				}
			}
		}
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE site_specs
		SET is_current = false, superseded_at = NOW()
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	superseded, err := res.RowsAffected()
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

	resp := gin.H{
		"id":      newID.String(),
		"aspect":  aspect,
		"updated": true,
		// superseded=false means this save CREATED the aspect. The :aspect param
		// has no allow-list, so that is also what a typo produces — an aspect
		// nothing reads. The caller should surface it, not treat it as routine.
		"superseded": superseded > 0,
	}
	if aspect == "evidence_base" {
		// The counts are the honest signal: a wrong-shape save cannot fake them,
		// and a zero here is the whole warning.
		var facts, bans, allowed int
		var regulated bool
		if eb != nil {
			facts, bans, allowed = len(eb.Facts), len(eb.BannedClaims), len(eb.AllowedEntities)
			regulated = eb.Regulated != nil
		}
		resp["facts_count"] = facts
		resp["banned_claims_count"] = bans
		resp["allowed_entities_count"] = allowed
		resp["regulated_attestation"] = regulated
	}
	c.JSON(http.StatusOK, resp)
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

const (
	// Page size for the work-item list. The default is generous because the
	// dashboard renders a filterable table rather than a paged one; the ceiling
	// exists so a caller cannot ask for the whole table at once.
	defaultWorkItemPageSize = 200
	maxWorkItemPageSize     = 1000
)

// parseBoundedQueryInt reads an integer query param, falling back to def when it
// is absent or unparseable, and clamping into [min, max].
func parseBoundedQueryInt(c *gin.Context, name string, def, min, max int) int {
	raw := c.Query(name)
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func (h *SiteAdminHandlers) HandleListWorkItems(c *gin.Context) {
	ctx := c.Request.Context()
	status := c.Query("status")
	pipeline := c.DefaultQuery("pipeline", "build")
	siteIDStr := c.Query("site_id")
	itemType := c.Query("item_type")

	// Paging. The limit used to be a hardcoded 50 with no total returned, and
	// the dashboard filtered the resulting window client-side — so with 687 open
	// build items it showed the newest 50, of which none were needs_human_review,
	// and reported that status's count as 0. A 208-item backlog read as empty for
	// four months (bugs_open/033). Hence: caller-settable, and `total` is always
	// returned so a truncated window can never look like the whole set again.
	limit := parseBoundedQueryInt(c, "limit", defaultWorkItemPageSize, 1, maxWorkItemPageSize)
	offset := parseBoundedQueryInt(c, "offset", 0, 0, math.MaxInt32)

	// Scope predicate: pipeline + site only. Both count queries below use exactly
	// this, deliberately excluding the status AND item_type predicates — a count
	// scoped by the filter it populates would collapse its own dropdown to the
	// one option already selected, leaving no way back to the others.
	baseWhere := ""
	baseArgs := []interface{}{}
	argIdx := 1

	// pipeline=all deliberately disables the filter: it defaults to "build", and
	// the dashboard hardcoded that, making the 94 content-pipeline items
	// unreachable by any route.
	if pipeline != "all" {
		baseWhere += fmt.Sprintf(" AND wi.pipeline = $%d", argIdx)
		baseArgs = append(baseArgs, pipeline)
		argIdx++
	}

	if siteIDStr != "" {
		if siteID, err := uuid.Parse(siteIDStr); err == nil {
			baseWhere += fmt.Sprintf(" AND wi.site_id = $%d", argIdx)
			baseArgs = append(baseArgs, siteID)
			argIdx++
		}
	}

	// Status counts across every status in the filtered set, so "Needs Review (N)"
	// is a fact about the table and not about the page.
	statusCounts := map[string]int{}
	countRows, err := h.db.QueryContext(ctx, `
		SELECT wi.status, count(*)
		FROM site_work_items wi
		JOIN sites s ON s.id = wi.site_id
		WHERE 1=1`+baseWhere+`
		GROUP BY wi.status
	`, baseArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for countRows.Next() {
		var st string
		var n int
		if err := countRows.Scan(&st, &n); err != nil {
			h.logger.Warn("Failed to scan work item status count", zap.Error(err))
			continue
		}
		statusCounts[st] = n
	}
	countRows.Close()

	// Same treatment for the item-type filter: its option list was also built
	// from the returned window, so types absent from the newest N were unlistable
	// and therefore unfilterable.
	typeCounts := map[string]int{}
	typeRows, err := h.db.QueryContext(ctx, `
		SELECT wi.item_type, count(*)
		FROM site_work_items wi
		JOIN sites s ON s.id = wi.site_id
		WHERE 1=1`+baseWhere+`
		GROUP BY wi.item_type
	`, baseArgs...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for typeRows.Next() {
		var it string
		var n int
		if err := typeRows.Scan(&it, &n); err != nil {
			h.logger.Warn("Failed to scan work item type count", zap.Error(err))
			continue
		}
		typeCounts[it] = n
	}
	typeRows.Close()

	// Now the predicates that apply to the page but not to the counts.
	pageWhere := baseWhere
	pageArgs := append([]interface{}{}, baseArgs...)

	if itemType != "" {
		pageWhere += fmt.Sprintf(" AND wi.item_type = $%d", argIdx)
		pageArgs = append(pageArgs, itemType)
		argIdx++
	}

	if status != "" && status != "all" {
		pageWhere += fmt.Sprintf(" AND wi.status = $%d", argIdx)
		pageArgs = append(pageArgs, status)
		argIdx++
	} else if status == "" {
		pageWhere += " AND wi.status != 'complete'"
	}

	// filing_mode: finds RFC_056 "record" verdicts (see HandleReleaseRecordVerdict)
	// among an otherwise undifferentiated `status='deferred'` backlog — 1,284
	// deferred needs_content_page/needs_content_planning rows fleet-wide as of
	// 2026-09-02, of which the filing_mode='record' subset is the reviewable
	// one; without this filter a human has no way to find them short of a raw
	// SQL query, which is the surface bugs_open/428 asked for.
	if filingMode := c.Query("filing_mode"); filingMode != "" {
		pageWhere += fmt.Sprintf(" AND wi.spec->>'filing_mode' = $%d", argIdx)
		pageArgs = append(pageArgs, filingMode)
		argIdx++
	}

	total := 0
	if err := h.db.QueryRowContext(ctx, `
		SELECT count(*)
		FROM site_work_items wi
		JOIN sites s ON s.id = wi.site_id
		WHERE 1=1`+pageWhere, pageArgs...).Scan(&total); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	query := `
		SELECT wi.id, wi.site_id, s.domain, wi.item_type, wi.status,
		       wi.severity, wi.summary, wi.spec,
				wi.handler_agent, wi.attempt_count, wi.max_attempts,
				wi.error, wi.created_at, wi.completed_at
		FROM site_work_items wi
		JOIN sites s ON s.id = wi.site_id
		WHERE 1=1` + pageWhere

	query += fmt.Sprintf(" ORDER BY wi.created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args := append(pageArgs, limit, offset)

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
			"handler_agent": handlerAgentNullable.String,
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

	// `count` stays the window size for backwards compatibility; `total` is the
	// size of the filtered set and `truncated` says plainly whether this response
	// is the whole story. A consumer that reads only `count` can no longer mistake
	// a page for a queue.
	c.JSON(http.StatusOK, gin.H{
		"items":         items,
		"count":         len(items),
		"total":         total,
		"limit":         limit,
		"offset":        offset,
		"truncated":     offset+len(items) < total,
		"status_counts": statusCounts,
		"type_counts":   typeCounts,
	})
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
	var handlerAgentNullable sql.NullString
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
		&handlerAgentNullable, &attemptCount, &maxAttempts,
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
		"handler_agent": handlerAgentNullable.String,
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
		  AND dup.status NOT IN ('complete', 'verified', 'rejected', 'wont_fix', 'failed', 'unresolved', 'cancelled')
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
		  AND status IN ('needs_human_review', 'failed', 'blocked', 'unresolved')
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
		  AND status IN ('needs_human_review', 'failed', 'blocked', 'triaged', 'unresolved')
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
// POST /admin/work-items/:item_id/release
// ============================================================================
//
// Releases ONE `filing_mode='record'` verdict row (RFC_056,
// write_audit_findings_action.go) — an LLM-audit-seat finding that was
// deliberately filed undispatchable (`status='deferred'`, `handler_agent=”`,
// the route it would have taken kept only in `spec.routed_handler`/
// `spec.routed_status`) rather than automatically triggering a page rewrite.
// RFC_056 exists because an earlier auto-dispatch of exactly this finding
// class destroyed live content (bugs_closed/238) — so this endpoint is the
// human half of that design, not a way around it: it releases the ONE row a
// person has actually looked at, never a class of them (bugs_open/428's own
// candidate-2 mistake was proposing a standing promoter for this shape).
//
// Deliberately does NOT execute `spec.release_recipe` as a stored SQL string
// — that field is a human-readable statement of intent for someone reading
// the row by hand, not something this endpoint should ever exec() verbatim.
// The WHERE clause below performs the identical, parameterised operation.
func (h *SiteAdminHandlers) HandleReleaseRecordVerdict(c *gin.Context) {
	ctx := c.Request.Context()
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	var body struct {
		ReleasedBy string `json:"released_by"`
		Notes      string `json:"notes"`
	}
	c.ShouldBindJSON(&body)
	if strings.TrimSpace(body.ReleasedBy) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "released_by is required — this is a human decision on a per-row basis, and the row records who made it"})
		return
	}

	var newStatus, newHandler string
	err = h.db.QueryRowContext(ctx, `
		UPDATE site_work_items
		SET status = spec->>'routed_status',
		    handler_agent = spec->>'routed_handler',
		    spec = spec || jsonb_build_object(
		        'filing_mode', 'released',
		        'released_by', $2::text,
		        'released_at', now(),
		        'released_notes', $3::text
		    ),
		    updated_at = NOW()
		WHERE id = $1
		  AND status = 'deferred'
		  AND spec->>'filing_mode' = 'record'
		  AND COALESCE(spec->>'routed_handler', '') <> ''
		  AND COALESCE(spec->>'routed_status', '') <> ''
		RETURNING status, handler_agent
	`, itemID, body.ReleasedBy, body.Notes).Scan(&newStatus, &newHandler)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found, not a filing_mode='record' verdict, already released, or missing routed_handler/routed_status"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("admin: released a filing_mode=record verdict (RFC_056)",
		zap.String("item_id", itemID.String()),
		zap.String("released_by", body.ReleasedBy),
		zap.String("new_status", newStatus),
		zap.String("new_handler", newHandler))

	c.JSON(http.StatusOK, gin.H{
		"released":      true,
		"id":            itemID.String(),
		"status":        newStatus,
		"handler_agent": newHandler,
	})
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
	// created_at is the moment the proposal was WRITTEN, and it is the only
	// reference point a staleness check can use: a parked `field_updates`
	// payload is a full-field replacement frozen then, so anything that touched
	// the target since is silently reverted by applying it (LANDMINES,
	// copy_quality_two_stage 2026-08-17).
	var reviewCreatedAt time.Time

	err = h.db.QueryRowContext(ctx, `
		SELECT site_id, spec, status, created_at
		FROM site_work_items
		WHERE id = $1
	`, itemID).Scan(&siteID, &specJSON, &currentStatus, &reviewCreatedAt)
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

	// ── Step 2: Create follow-on work item(s) if on_approve is set ───────
	//
	// bugs_open/466 — two independent defects, met here on the FIRST real
	// approval (2026-09-03, boxingonline, the first paid site):
	//
	//  1. `include_fields` named fields the review item's spec CANNOT hold.
	//     checkpoint_for_review — the only producer of these items — writes a
	//     fixed key set (review_data, checkpoint, source_agent, correlation_id,
	//     domain?, spec_aspect?, on_approve), so a lookup in `spec` can only
	//     ever yield null. [MEASURED 2026-09-03] 42 field mentions across 21
	//     items, all history since 2026-08-24: ZERO present at spec top level.
	//     The approved content lives in the body the admin submitted, so that
	//     is now the FALLBACK source. `spec` is still consulted FIRST, so no
	//     resolution that works today can change.
	//
	//  2. The shapes did not match even once the fields were plumbed:
	//     copy-editor proposes N edits, each with its own page_component_id and
	//     field_updates; section-editor applies ONE, and a two-edit proposal has
	//     no single target. `fan_out_from` names the array in the approved data
	//     and files one follow-on per element — the shape two lanes had already
	//     hand-built by the time this was found (this site 2026-09-03;
	//     copy_quality_two_stage on review be23d897, 2026-09-02).
	//
	// Both new keys default ABSENT: an on_approve naming neither behaves
	// exactly as it did before (owner ruling 2026-08-02 §2 — new authority on a
	// shared seam ships opt-in, with the unsafe side OFF).
	//
	// FOUR THINGS THE COUNCIL ROUND ON THIS CHANGE ASKED TO BE MEASURED RATHER
	// THAN ASSERTED (correlation d04c1bc1, REVISE round 1). All four are
	// recorded here because each is the kind of claim that goes stale by
	// ADDITION and reads as current for ever:
	//
	//  1. BLAST RADIUS. [MEASURED 2026-09-03] Across EVERY live, non-snapshot,
	//     undeleted agent_definition, exactly ONE workflow step configures
	//     on_approve: copy-editor's `request_review`. Not "one other consumer" —
	//     zero others. (fork_theme_from_site_action.go writes an on_approve from
	//     Go, but with no item_type, no include_fields and no checkpoint:true,
	//     so HandleApproveWorkItem 400s it before this block.) Re-measure with:
	//       SELECT d.type, s.key FROM agent_definitions d,
	//              jsonb_each(d.default_config->'workflow'->'steps') s
	//        WHERE d.is_active AND COALESCE(d.is_snapshot,false)=false
	//          AND d.deleted_at IS NULL AND s.value->'config' ? 'on_approve';
	//
	//  2. RFC_022 REOPENS ON THE SECOND CONSUMER. The architecture seat's point,
	//     recorded so the next author does not have to re-derive it: this ships
	//     WITHOUT the RFC_022 exemption because migration 750 wires a live
	//     consumer in the same commit, and it is defensible as a point fix for
	//     ONE wired consumer. A SECOND checkpoint producer setting fan_out_from
	//     is a different question and should come back as needs_rfc, not as
	//     another same-commit wiring.
	//
	//  3. N CHILDREN CANNOT COLLIDE ON THE DEDUP INDEX. idx_swi_dedup is
	//     UNIQUE(site_id, item_key) WHERE item_key IS NOT NULL AND status NOT IN
	//     (the terminal set). This handler has never set item_key, so every child
	//     is NULL and the partial index excludes it. Deliberately left that way:
	//     minting a per-element key would convert a cannot-happen into a
	//     can-fail-at-insert, and the endpoint already refuses a second press
	//     (the review row is no longer needs_human_review). The two hand-built
	//     precedents DID carry per-element keys (`section_edit:be23d897:hero`,
	//     `…:call-to-action`) — that is uniqueness WITHIN one approval, which
	//     NULL gives for free.
	//
	//  4. NO EXISTING PRIMITIVE WAS PASSED OVER. There is no generic
	//     one-item-fans-into-N helper on the admin side; `create_work_item` is a
	//     workflow ACTION, reachable by an agent step and not by an HTTP handler.
	//     `createFollowOn` below is the SINGLE creation path — the fan-out calls
	//     the same closure the single-item branch calls, N times, rather than
	//     forking a second inserter.
	var followOnID *string
	followOnIDs := []string{}
	skippedEdits := []map[string]interface{}{}
	var fanOutNote string

	if onApprove, ok := spec["on_approve"].(map[string]interface{}); ok {
		followItemType, _ := onApprove["item_type"].(string)
		followHandler, _ := onApprove["handler_agent"].(string)

		if followItemType == "" {
			followItemType = "approved_for_processing"
		}
		if followHandler == "" {
			followHandler = "page-build-handler"
		}

		// The admin's submitted body as an object, for include_fields and
		// fan_out_from lookups. A non-object body simply yields no lookups.
		var approvedMap map[string]interface{}
		_ = json.Unmarshal(body.ReviewData, &approvedMap)

		// include_fields lookup: the review item's own spec first (documented
		// semantics, preserved), then the approved body (the repair, note 1).
		resolveIncluded := func(dst map[string]interface{}) {
			includeFields, ok := onApprove["include_fields"].([]interface{})
			if !ok {
				return
			}
			for _, f := range includeFields {
				fieldName, ok := f.(string)
				if !ok {
					continue
				}
				if v, present := spec[fieldName]; present && v != nil {
					dst[fieldName] = v
					continue
				}
				if v, present := approvedMap[fieldName]; present {
					dst[fieldName] = v
				}
			}
		}

		// Applied LAST, so a fanned-out element cannot overwrite its own
		// provenance with a field of the same name.
		applyIdentity := func(dst map[string]interface{}) {
			dst["approved_by"] = "admin"
			dst["source_item_id"] = itemID.String()
			dst["original_pipeline"] = "build"
			if specAspect != "" {
				dst["spec_aspect"] = specAspect
			}
		}

		createFollowOn := func(followSpec map[string]interface{}, label string) {
			followSpecJSON, mErr := json.Marshal(followSpec)
			if mErr != nil {
				h.logger.Warn("Failed to marshal follow-on work item spec",
					zap.String("item_id", itemID.String()),
					zap.String("label", label),
					zap.Error(mErr))
				return
			}

			var newID uuid.UUID
			iErr := h.db.QueryRowContext(ctx, `
				INSERT INTO site_work_items (
					site_id, source, pipeline, item_type, severity, summary,
					spec, priority, handler_agent, status, created_by
				) VALUES ($1, 'admin-approved', 'build', $2, 'high', $3, $4::jsonb, 10, $5, 'triaged', 'admin')
				RETURNING id
			`, siteID, followItemType,
				fmt.Sprintf("Approved: %s%s", followItemType, label),
				string(followSpecJSON), followHandler).Scan(&newID)

			if iErr != nil {
				h.logger.Warn("Failed to create follow-on work item",
					zap.String("item_id", itemID.String()),
					zap.String("label", label),
					zap.Error(iErr))
				// Non-fatal — the spec is already saved
				return
			}
			id := newID.String()
			followOnIDs = append(followOnIDs, id)
			h.logger.Info("Follow-on work item created",
				zap.String("follow_on_id", id),
				zap.String("item_type", followItemType),
				zap.String("handler", followHandler),
				zap.String("label", label))
		}

		// ── fan_out_from: one follow-on per element of the named array ──
		fanOutFrom, _ := onApprove["fan_out_from"].(string)
		fanOutDefaults, _ := onApprove["defaults"].(map[string]interface{})
		var elements []interface{}
		fannedOut := false
		if fanOutFrom != "" {
			raw, present := approvedMap[fanOutFrom]
			switch {
			case !present:
				fanOutNote = fmt.Sprintf("fan_out_from %q is not present in the approved data; filed one follow-on carrying the whole payload instead", fanOutFrom)
			default:
				if arr, isArr := raw.([]interface{}); isArr {
					elements = arr
					fannedOut = true
					if len(arr) == 0 {
						fanOutNote = fmt.Sprintf("fan_out_from %q is an EMPTY array — nothing was filed", fanOutFrom)
					}
				} else {
					fanOutNote = fmt.Sprintf("fan_out_from %q is not an array; filed one follow-on carrying the whole payload instead", fanOutFrom)
				}
			}
		}

		if fannedOut {
			for i, raw := range elements {
				el, isObj := raw.(map[string]interface{})
				if !isObj {
					skippedEdits = append(skippedEdits, map[string]interface{}{
						"index":  i,
						"reason": "element is not an object",
					})
					continue
				}

				// A parked proposal's address rots TWO ways, and the council
				// round on this change was right that checking only the first is
				// a weaker guard than it reads as:
				//
				//   DELETED — a rerender REPLACES the component row with a new
				//   id, so the proposal's address stops existing (LANDMINES,
				//   copy_quality_two_stage 2026-08-18). [MEASURED 2026-09-03]
				//   3 of the 31 edits parked in needs_human_review already point
				//   at a row that is gone.
				//
				//   MOVED — the row survives but its content changed after the
				//   proposal was written. `field_updates` is a FULL-FIELD
				//   replacement frozen at proposal time, so applying it reverts
				//   every change to that field since, and on a single-field
				//   component that is the whole component (LANDMINES 2026-08-17).
				//   [MEASURED 2026-09-03] 0 of 31 today — so this arm costs
				//   nothing now and closes the class before it costs something.
				//
				// Either way: refuse and REPORT. Filing an edit the approver
				// cannot see is the shape 466 is about, and the fan-out must not
				// manufacture more of it while fixing it.
				//
				// ⚠ Known and accepted: the check and the INSERT are not one
				// transaction, so a rerender landing between them is not caught.
				// That window is milliseconds and its outcome is today's
				// behaviour (the item dies at load_edit_context), not a new
				// failure — so it is stated rather than engineered around.
				if pcID, _ := el["page_component_id"].(string); pcID != "" {
					var movedSince bool
					qErr := h.db.QueryRowContext(ctx, `
						SELECT pc.updated_at > $3
						FROM page_components pc
						JOIN pages p ON p.id = pc.page_id
						WHERE pc.id = $1::uuid AND p.site_id = $2
					`, pcID, siteID, reviewCreatedAt).Scan(&movedSince)

					reason := ""
					switch {
					case qErr == sql.ErrNoRows:
						reason = "page_component_id no longer exists on this site — a rerender replaces the row with a new id, so this proposal's address is gone. Re-run the proposer"
					case qErr != nil:
						reason = "could not verify page_component_id: " + qErr.Error()
					case movedSince:
						reason = "the target section changed after this proposal was written — field_updates is a full-field replacement frozen at proposal time, so applying it now would revert whatever changed since. Re-run the proposer"
					}
					if reason != "" {
						skippedEdits = append(skippedEdits, map[string]interface{}{
							"index":             i,
							"page_component_id": pcID,
							"slot_name":         el["slot_name"],
							"reason":            reason,
						})
						continue
					}
				}

				followSpec := map[string]interface{}{}
				for k, v := range fanOutDefaults {
					followSpec[k] = v
				}
				resolveIncluded(followSpec)
				for k, v := range el {
					followSpec[k] = v
				}
				applyIdentity(followSpec)
				followSpec["fan_out_index"] = i
				createFollowOn(followSpec, fmt.Sprintf(" (%d of %d)", i+1, len(elements)))
			}
		} else {
			followSpec := map[string]interface{}{
				"approved_data": json.RawMessage(body.ReviewData),
			}
			for k, v := range fanOutDefaults {
				followSpec[k] = v
			}
			resolveIncluded(followSpec)
			applyIdentity(followSpec)
			createFollowOn(followSpec, "")
		}

		if len(followOnIDs) > 0 {
			followOnID = &followOnIDs[0]
		}
	}

	// ── Step 3: Mark review item complete ────────────────────────────────
	resolution := "Approved via admin dashboard"
	if body.Notes != "" {
		resolution = body.Notes
	}

	// The review row's result must record what the approval actually PRODUCED —
	// bugs_open/466 candidate 3. A row reading `complete` while its spawned work
	// died is the estate's "a complete work item is not a repaired artefact"
	// shape aimed at a person: the approver has no other way to learn that his
	// decision did not land. `follow_on_item` is kept for existing readers and
	// now names the first of possibly several.
	followOnDetail := map[string]interface{}{
		"follow_on_items": followOnIDs,
	}
	if len(skippedEdits) > 0 {
		followOnDetail["skipped_edits"] = skippedEdits
	}
	if fanOutNote != "" {
		followOnDetail["fan_out_note"] = fanOutNote
	}
	followOnDetailJSON, _ := json.Marshal(followOnDetail)

	_, err = h.db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete',
		    completed_at = NOW(),
		    result = jsonb_build_object(
		        'resolution', $2,
		        'approved_by', 'admin',
		        'follow_on_item', $3
		    ) || $4::jsonb,
		    updated_at = NOW()
		WHERE id = $1
	`, itemID, resolution, followOnID, string(followOnDetailJSON))
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
	// Always present, so "nothing was filed" is a visible [] rather than an
	// absent key that reads as success.
	result["follow_on_item_ids"] = followOnIDs
	if len(skippedEdits) > 0 {
		result["skipped_edits"] = skippedEdits
	}
	if fanOutNote != "" {
		result["fan_out_note"] = fanOutNote
	}

	c.JSON(http.StatusOK, result)
}

// HandleRequestChangesWorkItem files the owner's critique of a site as a new
// owner_critique work item, WITHOUT resolving or approving the item it was typed
// against — a "request changes" verb alongside approve, so a pre-delivery review
// can send the site back for work while the delivery gate stays closed.
//
// The owner_critique item is deliberately NOT cluster-routed: handler_agent is
// empty and no dispatch loop claims the type. Its consumer is the workstation
// dispatcher thread, which polls for open owner_critique rows and routes the text
// to the owning session threads (docs024_key_docs_latest/dispatcher_thread/).
// Minting a cluster-routed type here would repeat bugs_open/279 (items filed
// under an unrouteable type were hand-cancelled); minting an explicitly
// thread-consumed type, stated in the row itself, is the honest version.
func (h *SiteAdminHandlers) HandleRequestChangesWorkItem(c *gin.Context) {
	ctx := c.Request.Context()

	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	var body struct {
		Critique string `json:"critique" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	critique := strings.TrimSpace(body.Critique)
	if critique == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "critique must not be empty"})
		return
	}

	// The origin item anchors the critique to a site; it is read, never written —
	// a checkpoint review item stays at needs_human_review and keeps gating.
	var siteID uuid.UUID
	var originType, originStatus string
	var domain sql.NullString
	err = h.db.QueryRowContext(ctx, `
		SELECT w.site_id, w.item_type, w.status, s.domain
		FROM site_work_items w
		JOIN sites s ON s.id = w.site_id
		WHERE w.id = $1
	`, itemID).Scan(&siteID, &originType, &originStatus, &domain)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}
	if err != nil {
		h.logger.Error("request_changes: origin lookup failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		return
	}

	summary := "Owner critique: " + critique
	if len(summary) > 180 {
		summary = summary[:177] + "..."
	}

	spec := map[string]interface{}{
		"critique":         critique,
		"origin_item_id":   itemID.String(),
		"origin_item_type": originType,
		"domain":           domain.String,
		"filed_via":        "admin_request_changes",
		"consumer":         "thread-dispatcher",
		"consumer_note":    "not cluster-routed by design; the workstation dispatcher thread polls open owner_critique items and routes them to session threads",
	}
	specJSON, err := json.Marshal(spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spec marshal failed"})
		return
	}

	// status='needs_human_review' is the load-bearing half of "not cluster-routed",
	// and the first cut got it wrong twice over — council 9f1cb042 round 1
	// (editquality, gating) objected that the empty-handler_agent-at-triaged claim
	// was asserted, not verified, and verification proved the objection right at
	// two artefacts: (1) LoadWorkItemsAction's WHERE has no pipeline default
	// (item_pipeline is optional config) and build-dispatch-loop's load_items
	// config sets no item_pipeline, so a bare triaged row here WOULD have been
	// loaded and claimed on the site's next dispatch pick; (2) the schema itself
	// refuses the shipped shape — CHECK swi_no_handlerless_promotable forbids an
	// empty handler_agent at triaged/approved/claimed, so the first cut's INSERT
	// could never have succeeded at runtime. needs_human_review is outside both
	// the loader's status set (triaged/approved) and the constraint's forbidden
	// set, and it says what is true: the item awaits human-side routing (the
	// dispatcher thread), never a cluster handler. approval_mode='manual' stays
	// as defence in depth against any future status transition.
	var newID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, approval_mode, created_by
		) VALUES ($1, 'owner', 'delivery', 'owner_critique', 'high', $2, $3::jsonb, 10, '', 'needs_human_review', 'manual', 'admin-request-changes')
		RETURNING id
	`, siteID, summary, specJSON).Scan(&newID)
	if err != nil {
		h.logger.Error("request_changes: insert failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "insert failed"})
		return
	}

	h.logger.Info("owner critique filed",
		zap.String("critique_item_id", newID.String()),
		zap.String("origin_item_id", itemID.String()),
		zap.String("site_id", siteID.String()))

	c.JSON(http.StatusOK, gin.H{
		"critique_item_id": newID.String(),
		"origin_item_id":   itemID.String(),
		"origin_status":    originStatus,
		"site_id":          siteID.String(),
		"note":             "critique filed as owner_critique; the origin item is unchanged and delivery stays gated",
	})
}
