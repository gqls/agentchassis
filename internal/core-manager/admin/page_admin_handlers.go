// FILE: internal/core-manager/admin/page_admin_handlers.go
//
// Admin endpoints for browsing and editing page structure:
//   - Pages: list with component counts and lock status
//   - Components: list with content preview, edit, lock, unlock, remove
//
// These endpoints power the page structure browser in the admin dashboard,
// enabling inline content editing with lock enforcement.

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

type PageAdminHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPageAdminHandlers(db *sql.DB, logger *zap.Logger) *PageAdminHandlers {
	return &PageAdminHandlers{db: db, logger: logger}
}

// ============================================================================
// GET /admin/sites/:site_id/pages
// ============================================================================

func (h *PageAdminHandlers) HandleListPages(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT p.id, p.name, COALESCE(p.title, ''), COALESCE(p.url, ''),
		       COALESCE(p.page_type, ''), COALESCE(p.status, 'active'),
		       COALESCE(p.build_status, 'pending'),
		       p.last_built_at, p.deployed_at,
		       COUNT(pc.id) AS component_count,
		       COUNT(pc.id) FILTER (WHERE pc.locked_at IS NOT NULL) AS locked_count,
		       COUNT(pc.id) FILTER (WHERE pc.rendered_html IS NULL OR TRIM(pc.rendered_html) = '' OR LENGTH(pc.rendered_html) < 50) AS empty_count
		FROM pages p
		LEFT JOIN page_components pc ON pc.page_id = p.id
		    AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')
		    AND pc.build_status != 'removed'
		WHERE p.site_id = $1
		  AND p.status IN ('active', 'deployed')
		GROUP BY p.id
		ORDER BY p.nav_order, p.name
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var pages []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, title, url, pageType, status, buildStatus string
		var lastBuiltAt, deployedAt sql.NullTime
		var componentCount, lockedCount, emptyCount int

		if err := rows.Scan(&id, &name, &title, &url, &pageType, &status, &buildStatus,
			&lastBuiltAt, &deployedAt,
			&componentCount, &lockedCount, &emptyCount); err != nil {
			h.logger.Warn("Failed to scan page", zap.Error(err))
			continue
		}

		page := map[string]interface{}{
			"id":              id.String(),
			"name":            name,
			"title":           title,
			"url":             url,
			"page_type":       pageType,
			"status":          status,
			"build_status":    buildStatus,
			"component_count": componentCount,
			"locked_count":    lockedCount,
			"empty_count":     emptyCount,
		}
		if lastBuiltAt.Valid {
			page["last_built_at"] = lastBuiltAt.Time
		}
		if deployedAt.Valid {
			page["deployed_at"] = deployedAt.Time
		}
		pages = append(pages, page)
	}

	c.JSON(http.StatusOK, gin.H{"pages": pages, "count": len(pages)})
}

// ============================================================================
// GET /admin/sites/:site_id/pages/:page_name/components
// ============================================================================

func (h *PageAdminHandlers) HandleListComponents(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	pageName := c.Param("page_name")
	if pageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_name required"})
		return
	}

	// Load the page record
	var pageID uuid.UUID
	var pageTitle, pageURL, pageType string
	var suppressedJSON, pageSpecJSON []byte
	err = h.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(title, name), COALESCE(url, ''), COALESCE(page_type, ''),
		       COALESCE(suppressed_sections, '[]'::jsonb),
		       COALESCE(page_spec, '{}'::jsonb)
		FROM pages WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&pageID, &pageTitle, &pageURL, &pageType, &suppressedJSON, &pageSpecJSON)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var suppressedSections []string
	json.Unmarshal(suppressedJSON, &suppressedSections)

	var pageSpec interface{}
	json.Unmarshal(pageSpecJSON, &pageSpec)

	// Load components in position order
	rows, err := h.db.QueryContext(ctx, `
		SELECT pc.id, pc.position, COALESCE(pc.slot_name, ''),
		       pc.content_data,
		       pc.rendered_html AS html_preview,
		       LENGTH(COALESCE(pc.rendered_html, '')) AS html_length,
		       COALESCE(pc.build_status, 'pending'),
		       pc.locked_at, COALESCE(pc.locked_by, ''),
		       pc.updated_at,
		       pc.content_brief
		FROM page_components pc
		WHERE pc.page_id = $1
		  AND COALESCE(pc.slot_name, '') NOT IN ('header', 'footer', 'head')
		  AND pc.build_status != 'removed'
		ORDER BY pc.position
	`, pageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var components []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var position int
		var slotName, buildStatus, lockedBy string
		var contentDataJSON, contentBriefJSON []byte
		var htmlPreview sql.NullString
		var htmlLength int
		var lockedAt sql.NullTime
		var updatedAt sql.NullTime

		if err := rows.Scan(&id, &position, &slotName,
			&contentDataJSON, &htmlPreview, &htmlLength,
			&buildStatus, &lockedAt, &lockedBy, &updatedAt,
			&contentBriefJSON); err != nil {
			h.logger.Warn("Failed to scan component", zap.Error(err))
			continue
		}

		var contentData interface{}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &contentData)
		}

		var contentBrief interface{}
		if len(contentBriefJSON) > 0 {
			json.Unmarshal(contentBriefJSON, &contentBrief)
		}

		comp := map[string]interface{}{
			"id":            id.String(),
			"position":      position,
			"slot_name":     slotName,
			"content_data":  contentData,
			"content_brief": contentBrief,
			"html_preview":  "",
			"html_length":   htmlLength,
			"build_status":  buildStatus,
			"locked":        lockedAt.Valid,
			"locked_by":     lockedBy,
			"is_empty":      htmlLength < 50,
		}
		if htmlPreview.Valid {
			comp["html_preview"] = htmlPreview.String
		}
		if lockedAt.Valid {
			comp["locked_at"] = lockedAt.Time
		}
		if updatedAt.Valid {
			comp["updated_at"] = updatedAt.Time
		}

		components = append(components, comp)
	}

	c.JSON(http.StatusOK, gin.H{
		"page": map[string]interface{}{
			"id":                  pageID.String(),
			"name":                pageName,
			"title":               pageTitle,
			"url":                 pageURL,
			"page_type":           pageType,
			"page_spec":           pageSpec,
			"suppressed_sections": suppressedSections,
		},
		"components": components,
		"count":      len(components),
	})
}

// ============================================================================
// PATCH /admin/sites/:site_id/pages/:page_name/components/:component_id
// ============================================================================
//
// Updates a component's content_data and/or rendered_html.
// Optionally locks the component and triggers a page rebuild.
//
// Request body:
//   {
//     "content_data": { ... },     // structured content fields
//     "rendered_html": "...",       // raw HTML override (optional)
//     "lock": true,                 // auto-lock on edit (default true)
//     "rebuild_page": true          // create page_rerender work item
//   }

func (h *PageAdminHandlers) HandleUpdateComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	componentID, err := uuid.Parse(c.Param("component_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component_id"})
		return
	}
	pageName := c.Param("page_name")

	var body struct {
		ContentData  *json.RawMessage `json:"content_data"`
		RenderedHTML *string          `json:"rendered_html"`
		Lock         *bool            `json:"lock"`
		RebuildPage  *bool            `json:"rebuild_page"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify the component belongs to the right site/page
	var pageID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT pc.page_id FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.id = $1 AND p.site_id = $2 AND p.name = $3
	`, componentID, siteID, pageName).Scan(&pageID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "component not found on this page"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// Step 1: Save current state to history
	_, err = tx.ExecContext(ctx, `
		INSERT INTO page_component_history (component_id, page_id, site_id, content_data, source)
		SELECT pc.id, pc.page_id, p.site_id,
		       COALESCE(pc.content_data, jsonb_build_object(
		           'rendered_html', pc.rendered_html,
		           'slot_name', pc.slot_name,
		           'build_status', pc.build_status
		       )),
		       'admin-edit'
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.id = $1
	`, componentID)
	if err != nil {
		h.logger.Warn("Failed to save component history", zap.Error(err))
		// Non-fatal — continue with the edit
	}

	// Step 2: Build the UPDATE
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if body.ContentData != nil {
		setClauses = append(setClauses, fmt.Sprintf("content_data = $%d::jsonb", argIdx))
		args = append(args, string(*body.ContentData))
		argIdx++
	}
	if body.RenderedHTML != nil {
		setClauses = append(setClauses, fmt.Sprintf("rendered_html = $%d", argIdx))
		args = append(args, *body.RenderedHTML)
		argIdx++
	}

	// Default: lock on edit
	shouldLock := true
	if body.Lock != nil {
		shouldLock = *body.Lock
	}
	if shouldLock {
		setClauses = append(setClauses, "locked_at = NOW()")
		setClauses = append(setClauses, "locked_by = 'admin'")
	}

	query := fmt.Sprintf("UPDATE page_components SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, componentID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Component updated via admin",
		zap.String("component_id", componentID.String()),
		zap.String("page", pageName),
		zap.Bool("locked", shouldLock))

	// Step 3: Optionally create rebuild work item
	shouldRebuild := true
	if body.RebuildPage != nil {
		shouldRebuild = *body.RebuildPage
	}

	var rebuildItemID *string
	if shouldRebuild {
		var newID uuid.UUID
		err = h.db.QueryRowContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by
			) VALUES ($1, 'admin-edit', 'build', 'page_rerender', 'high',
			          $2, $3::jsonb, $4, 5, 'rerender-pages', 'triaged', 'admin')
			RETURNING id
		`, siteID,
			fmt.Sprintf("Rerender page: %s (admin edit)", pageName),
			fmt.Sprintf(`{"page_name":"%s","reason":"admin_component_edit","component_id":"%s"}`, pageName, componentID),
			pageID).Scan(&newID)
		if err != nil {
			h.logger.Warn("Failed to create rerender work item", zap.Error(err))
		} else {
			id := newID.String()
			rebuildItemID = &id
		}
	}

	result := gin.H{
		"updated":      true,
		"component_id": componentID.String(),
		"locked":       shouldLock,
	}
	if rebuildItemID != nil {
		result["rebuild_item_id"] = *rebuildItemID
	}
	c.JSON(http.StatusOK, result)
}

// ============================================================================
// POST /admin/sites/:site_id/pages/:page_name/components/:component_id/unlock
// ============================================================================

func (h *PageAdminHandlers) HandleUnlockComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	componentID, err := uuid.Parse(c.Param("component_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component_id"})
		return
	}

	// Verify ownership
	result, err := h.db.ExecContext(ctx, `
		UPDATE page_components pc
		SET locked_at = NULL, locked_by = NULL, updated_at = NOW()
		FROM pages p
		WHERE pc.id = $1 AND pc.page_id = p.id AND p.site_id = $2
	`, componentID, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "component not found"})
		return
	}

	h.logger.Info("Component unlocked via admin",
		zap.String("component_id", componentID.String()))

	c.JSON(http.StatusOK, gin.H{"unlocked": true, "component_id": componentID.String()})
}

// ============================================================================
// POST /admin/sites/:site_id/pages/:page_name/components/:component_id/lock
// ============================================================================

func (h *PageAdminHandlers) HandleLockComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	componentID, err := uuid.Parse(c.Param("component_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component_id"})
		return
	}

	result, err := h.db.ExecContext(ctx, `
		UPDATE page_components pc
		SET locked_at = NOW(), locked_by = 'admin', updated_at = NOW()
		FROM pages p
		WHERE pc.id = $1 AND pc.page_id = p.id AND p.site_id = $2
	`, componentID, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "component not found"})
		return
	}

	h.logger.Info("Component locked via admin",
		zap.String("component_id", componentID.String()))

	c.JSON(http.StatusOK, gin.H{"locked": true, "component_id": componentID.String()})
}

// ============================================================================
// DELETE /admin/sites/:site_id/pages/:page_name/components/:component_id
// ============================================================================
//
// Soft-deletes a component: sets build_status='removed', locks it as
// 'admin-removed', and triggers a page rebuild without the section.

func (h *PageAdminHandlers) HandleRemoveComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	componentID, err := uuid.Parse(c.Param("component_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component_id"})
		return
	}
	pageName := c.Param("page_name")

	// Get the component's slot_name and page_id for suppression
	var slotName string
	var pageID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT pc.slot_name, pc.page_id
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.id = $1 AND p.site_id = $2 AND p.name = $3
	`, componentID, siteID, pageName).Scan(&slotName, &pageID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "component not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// Save to history
	_, _ = tx.ExecContext(ctx, `
		INSERT INTO page_component_history (component_id, page_id, site_id, content_data, source)
		SELECT pc.id, pc.page_id, p.site_id,
		       COALESCE(pc.content_data, jsonb_build_object(
		           'rendered_html', pc.rendered_html,
		           'slot_name', pc.slot_name
		       )),
		       'admin-removed'
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.id = $1
	`, componentID)

	// Soft delete: mark removed and lock
	_, err = tx.ExecContext(ctx, `
		UPDATE page_components
		SET build_status = 'removed',
		    locked_at = NOW(),
		    locked_by = 'admin-removed',
		    updated_at = NOW()
		WHERE id = $1
	`, componentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Add to page suppressed_sections (Phase 5 prep — column may not exist yet)
	_, _ = tx.ExecContext(ctx, `
		UPDATE pages
		SET suppressed_sections = COALESCE(suppressed_sections, '[]'::jsonb) || to_jsonb($2::text)
		WHERE id = $1
		  AND NOT (COALESCE(suppressed_sections, '[]'::jsonb) ? $2)
	`, pageID, slotName)

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create rerender work item
	h.db.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, page_id, priority, handler_agent, status, created_by
		) VALUES ($1, 'admin-edit', 'build', 'page_rerender', 'high',
		          $2, $3::jsonb, $4, 5, 'rerender-pages', 'triaged', 'admin')
	`, siteID,
		fmt.Sprintf("Rerender page: %s (section removed)", pageName),
		fmt.Sprintf(`{"page_name":"%s","reason":"section_removed","removed_slot":"%s"}`, pageName, slotName),
		pageID)

	h.logger.Info("Component removed via admin",
		zap.String("component_id", componentID.String()),
		zap.String("slot_name", slotName),
		zap.String("page", pageName))

	c.JSON(http.StatusOK, gin.H{
		"removed":      true,
		"component_id": componentID.String(),
		"slot_name":    slotName,
		"suppressed":   true,
	})
}

// ============================================================================
// POST /admin/sites/:site_id/pages/:page_name/restore-section
// ============================================================================
//
// Restores a suppressed section: removes it from suppressed_sections and
// optionally creates a work item to populate the section.

func (h *PageAdminHandlers) HandleRestoreSection(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	pageName := c.Param("page_name")

	var body struct {
		SlotName   string `json:"slot_name" binding:"required"`
		CreateItem *bool  `json:"create_item"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get page ID
	var pageID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT id FROM pages WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&pageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}

	// Remove from suppressed_sections
	_, err = h.db.ExecContext(ctx, `
		UPDATE pages
		SET suppressed_sections = COALESCE(suppressed_sections, '[]'::jsonb) - $2
		WHERE id = $1
	`, pageID, body.SlotName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Restore the component if it exists with build_status='removed'
	_, _ = h.db.ExecContext(ctx, `
		UPDATE page_components
		SET build_status = 'pending',
		    locked_at = NULL,
		    locked_by = NULL,
		    updated_at = NOW()
		WHERE page_id = $1 AND slot_name = $2 AND build_status = 'removed'
	`, pageID, body.SlotName)

	h.logger.Info("Section restored via admin",
		zap.String("page", pageName),
		zap.String("slot_name", body.SlotName))

	// Optionally create a work item to populate the section
	shouldCreate := true
	if body.CreateItem != nil {
		shouldCreate = *body.CreateItem
	}

	var itemID *string
	if shouldCreate {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"page_name":    pageName,
			"slot_name":    body.SlotName,
			"reason":       "section_restored",
			"check":        "empty_sections",
			"component_id": "",
		})

		var newID uuid.UUID
		err = h.db.QueryRowContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by
			) VALUES ($1, 'admin-restore', 'build', 'empty_section', 'medium',
			          $2, $3::jsonb, $4, 50, 'page-build-handler', 'triaged', 'admin')
			RETURNING id
		`, siteID,
			fmt.Sprintf("Populate restored section '%s' on page %s", body.SlotName, pageName),
			string(specJSON), pageID).Scan(&newID)

		if err != nil {
			h.logger.Warn("Failed to create restore work item", zap.Error(err))
		} else {
			id := newID.String()
			itemID = &id
		}
	}

	result := gin.H{
		"restored":  true,
		"slot_name": body.SlotName,
		"page":      pageName,
	}
	if itemID != nil {
		result["item_id"] = *itemID
	}
	c.JSON(http.StatusOK, result)
}

// ============================================================================
// POST /admin/sites/:site_id/pages/:page_name/components/:component_id/regenerate
// ============================================================================
//
// Creates a content_rewrite work item with an updated content_brief.
// The content writer will use the brief as its instructions when rewriting.

func (h *PageAdminHandlers) HandleRegenerateComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	componentID, err := uuid.Parse(c.Param("component_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid component_id"})
		return
	}
	pageName := c.Param("page_name")

	var body struct {
		Brief map[string]interface{} `json:"brief"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify component exists and get slot_name and page_id
	var slotName string
	var pageID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT pc.slot_name, pc.page_id
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE pc.id = $1 AND p.site_id = $2 AND p.name = $3
	`, componentID, siteID, pageName).Scan(&slotName, &pageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "component not found"})
		return
	}

	// Save the updated brief to the component
	if body.Brief != nil {
		briefJSON, _ := json.Marshal(body.Brief)
		h.db.ExecContext(ctx, `
			UPDATE page_components SET content_brief = $1::jsonb, updated_at = NOW()
			WHERE id = $2
		`, string(briefJSON), componentID)
	}

	// Create content_rewrite work item
	spec := map[string]interface{}{
		"page_name":     pageName,
		"component_id":  componentID.String(),
		"slot_name":     slotName,
		"reason":        "admin_regenerate",
		"content_brief": body.Brief,
	}
	specJSON, _ := json.Marshal(spec)

	var newID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, page_id, priority, handler_agent, status, created_by
		) VALUES ($1, 'admin-regenerate', 'build', 'content_rewrite', 'medium',
		          $2, $3::jsonb, $4, 20, 'page-build-handler', 'triaged', 'admin')
		RETURNING id
	`, siteID,
		fmt.Sprintf("Regenerate section '%s' on %s (brief updated by admin)", slotName, pageName),
		string(specJSON), pageID).Scan(&newID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Component regeneration queued",
		zap.String("component_id", componentID.String()),
		zap.String("slot_name", slotName),
		zap.String("page", pageName))

	c.JSON(http.StatusOK, gin.H{
		"regenerating": true,
		"component_id": componentID.String(),
		"item_id":      newID.String(),
	})
}

// ============================================================================
// POST /admin/sites/:site_id/pages/:page_name/regenerate
// ============================================================================
//
// Creates content_rewrite work items for all unlocked sections on a page.
// Optionally updates the page_spec first.

func (h *PageAdminHandlers) HandleRegeneratePage(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	pageName := c.Param("page_name")

	var body struct {
		PageSpec *map[string]interface{} `json:"page_spec"`
	}
	c.ShouldBindJSON(&body)

	// Get page
	var pageID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT id FROM pages WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&pageID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}

	// Update page_spec if provided
	if body.PageSpec != nil {
		specJSON, _ := json.Marshal(*body.PageSpec)
		h.db.ExecContext(ctx, `
			UPDATE pages SET page_spec = $1::jsonb, updated_at = NOW()
			WHERE id = $2
		`, string(specJSON), pageID)
	}

	// Find unlocked sections
	rows, err := h.db.QueryContext(ctx, `
		SELECT id, slot_name FROM page_components
		WHERE page_id = $1
		  AND build_status != 'removed'
		  AND COALESCE(slot_name, '') NOT IN ('header', 'footer', 'head')
		  AND (locked_at IS NULL OR locked_by NOT IN ('admin', 'admin-removed', 'checkpoint'))
		ORDER BY position
	`, pageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var created []map[string]interface{}
	var skipped int
	for rows.Next() {
		var compID uuid.UUID
		var slotName string
		if err := rows.Scan(&compID, &slotName); err != nil {
			continue
		}

		spec := map[string]interface{}{
			"page_name":    pageName,
			"component_id": compID.String(),
			"slot_name":    slotName,
			"reason":       "admin_regenerate_page",
		}
		specJSON, _ := json.Marshal(spec)

		var newID uuid.UUID
		err = h.db.QueryRowContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by
			) VALUES ($1, 'admin-regenerate', 'build', 'content_rewrite', 'medium',
			          $2, $3::jsonb, $4, 20, 'page-build-handler', 'triaged', 'admin')
			RETURNING id
		`, siteID,
			fmt.Sprintf("Regenerate '%s' on %s (page regeneration)", slotName, pageName),
			string(specJSON), pageID).Scan(&newID)

		if err != nil {
			h.logger.Warn("Failed to create regeneration item",
				zap.String("slot", slotName), zap.Error(err))
			continue
		}
		created = append(created, map[string]interface{}{
			"slot_name": slotName,
			"item_id":   newID.String(),
		})
	}

	h.logger.Info("Page regeneration queued",
		zap.String("page", pageName),
		zap.Int("sections", len(created)),
		zap.Int("skipped_locked", skipped))

	c.JSON(http.StatusOK, gin.H{
		"regenerating":  true,
		"page":          pageName,
		"items_created": len(created),
		"items_skipped": skipped,
		"items":         created,
	})
}

// ============================================================================
// PATCH /admin/sites/:site_id/pages/:page_name/spec
// ============================================================================
//
// Updates the page_spec for a page (purpose, direction).

func (h *PageAdminHandlers) HandleUpdatePageSpec(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	pageName := c.Param("page_name")

	var body struct {
		PageSpec map[string]interface{} `json:"page_spec" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	specJSON, _ := json.Marshal(body.PageSpec)
	result, err := h.db.ExecContext(ctx, `
		UPDATE pages SET page_spec = $1::jsonb, updated_at = NOW()
		WHERE site_id = $2 AND name = $3
	`, string(specJSON), siteID, pageName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true, "page": pageName})
}

// ============================================================================
// Site-Wide Components (Phase 7)
// Header, footer, and head (CSS) — shared across all pages.
// ============================================================================

// GET /admin/sites/:site_id/site-components

func (h *PageAdminHandlers) HandleListSiteComponents(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT sc.id, sc.slot_name, sc.rendered_html, sc.content_data,
		       LENGTH(COALESCE(sc.rendered_html, '')) AS html_length,
		       COALESCE(sc.build_status, 'pending'),
		       sc.locked_at, COALESCE(sc.locked_by, ''),
		       sc.updated_at
		FROM site_components sc
		WHERE sc.site_id = $1
		ORDER BY CASE sc.slot_name
		    WHEN 'header' THEN 1
		    WHEN 'footer' THEN 2
		    WHEN 'head' THEN 3
		    ELSE 4
		END
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var components []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var slotName, buildStatus, lockedBy string
		var renderedHTML sql.NullString
		var contentDataJSON []byte
		var htmlLength int
		var lockedAt, updatedAt sql.NullTime

		if err := rows.Scan(&id, &slotName, &renderedHTML, &contentDataJSON,
			&htmlLength, &buildStatus, &lockedAt, &lockedBy, &updatedAt); err != nil {
			h.logger.Warn("Failed to scan site component", zap.Error(err))
			continue
		}

		var contentData interface{}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &contentData)
		}

		comp := map[string]interface{}{
			"id":            id.String(),
			"slot_name":     slotName,
			"rendered_html": "",
			"content_data":  contentData,
			"html_length":   htmlLength,
			"build_status":  buildStatus,
			"locked":        lockedAt.Valid,
			"locked_by":     lockedBy,
		}
		if renderedHTML.Valid {
			comp["rendered_html"] = renderedHTML.String
		}
		if lockedAt.Valid {
			comp["locked_at"] = lockedAt.Time
		}
		if updatedAt.Valid {
			comp["updated_at"] = updatedAt.Time
		}
		components = append(components, comp)
	}

	c.JSON(http.StatusOK, gin.H{"components": components, "count": len(components)})
}

// PATCH /admin/sites/:site_id/site-components/:slot_name

func (h *PageAdminHandlers) HandleUpdateSiteComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	slotName := c.Param("slot_name")

	var body struct {
		RenderedHTML *string          `json:"rendered_html"`
		ContentData  *json.RawMessage `json:"content_data"`
		Lock         *bool            `json:"lock"`
		RebuildSite  *bool            `json:"rebuild_site"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify exists
	var scID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		SELECT id FROM site_components WHERE site_id = $1 AND slot_name = $2
	`, siteID, slotName).Scan(&scID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "site component not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Build UPDATE
	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if body.RenderedHTML != nil {
		setClauses = append(setClauses, fmt.Sprintf("rendered_html = $%d", argIdx))
		args = append(args, *body.RenderedHTML)
		argIdx++
	}
	if body.ContentData != nil {
		setClauses = append(setClauses, fmt.Sprintf("content_data = $%d::jsonb", argIdx))
		args = append(args, string(*body.ContentData))
		argIdx++
	}

	shouldLock := true
	if body.Lock != nil {
		shouldLock = *body.Lock
	}
	if shouldLock {
		setClauses = append(setClauses, "locked_at = NOW()")
		setClauses = append(setClauses, "locked_by = 'admin'")
	}

	query := fmt.Sprintf("UPDATE site_components SET %s WHERE id = $%d",
		strings.Join(setClauses, ", "), argIdx)
	args = append(args, scID)

	_, err = h.db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Site component updated via admin",
		zap.String("slot_name", slotName),
		zap.Bool("locked", shouldLock))

	// Create full-site rerender work item
	shouldRebuild := true
	if body.RebuildSite != nil {
		shouldRebuild = *body.RebuildSite
	}

	var rebuildItemID *string
	if shouldRebuild {
		pageRows, _ := h.db.QueryContext(ctx, `
			SELECT name FROM pages
			WHERE site_id = $1 AND status IN ('active', 'deployed')
			ORDER BY name
		`, siteID)
		var pageNames []string
		if pageRows != nil {
			defer pageRows.Close()
			for pageRows.Next() {
				var name string
				if pageRows.Scan(&name) == nil {
					pageNames = append(pageNames, name)
				}
			}
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"reason":                  "admin_site_component_edit",
			"slot_name":               slotName,
			"refresh_site_components": true,
			"affected_pages":          pageNames,
			"affected_page_count":     len(pageNames),
		})

		var newID uuid.UUID
		err = h.db.QueryRowContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, priority, handler_agent, status, created_by
			) VALUES ($1, 'admin-edit', 'build', 'needs_rerender', 'high',
			          $2, $3::jsonb, 5, 'rerender-pages', 'triaged', 'admin')
			RETURNING id
		`, siteID,
			fmt.Sprintf("Full site rerender: %s edited by admin", slotName),
			string(specJSON)).Scan(&newID)
		if err != nil {
			h.logger.Warn("Failed to create rerender work item", zap.Error(err))
		} else {
			id := newID.String()
			rebuildItemID = &id
		}
	}

	scResult := gin.H{
		"updated":   true,
		"slot_name": slotName,
		"locked":    shouldLock,
	}
	if rebuildItemID != nil {
		scResult["rebuild_item_id"] = *rebuildItemID
	}
	c.JSON(http.StatusOK, scResult)
}

// POST /admin/sites/:site_id/site-components/:slot_name/lock

func (h *PageAdminHandlers) HandleLockSiteComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	slotName := c.Param("slot_name")

	res, err := h.db.ExecContext(ctx, `
		UPDATE site_components SET locked_at = NOW(), locked_by = 'admin', updated_at = NOW()
		WHERE site_id = $1 AND slot_name = $2
	`, siteID, slotName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "site component not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"locked": true, "slot_name": slotName})
}

// POST /admin/sites/:site_id/site-components/:slot_name/unlock

func (h *PageAdminHandlers) HandleUnlockSiteComponent(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	slotName := c.Param("slot_name")

	res, err := h.db.ExecContext(ctx, `
		UPDATE site_components SET locked_at = NULL, locked_by = NULL, updated_at = NOW()
		WHERE site_id = $1 AND slot_name = $2
	`, siteID, slotName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "site component not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"unlocked": true, "slot_name": slotName})
}
