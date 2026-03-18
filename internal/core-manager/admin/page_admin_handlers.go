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
	err = h.db.QueryRowContext(ctx, `
		SELECT id, COALESCE(title, name), COALESCE(url, ''), COALESCE(page_type, '')
		FROM pages WHERE site_id = $1 AND name = $2
	`, siteID, pageName).Scan(&pageID, &pageTitle, &pageURL, &pageType)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "page not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Load components in position order
	rows, err := h.db.QueryContext(ctx, `
		SELECT pc.id, pc.position, COALESCE(pc.slot_name, ''),
		       pc.content_data,
		       LEFT(pc.rendered_html, 500) AS html_preview,
		       LENGTH(COALESCE(pc.rendered_html, '')) AS html_length,
		       COALESCE(pc.build_status, 'pending'),
		       pc.locked_at, COALESCE(pc.locked_by, ''),
		       pc.updated_at
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
		var contentDataJSON []byte
		var htmlPreview sql.NullString
		var htmlLength int
		var lockedAt sql.NullTime
		var updatedAt sql.NullTime

		if err := rows.Scan(&id, &position, &slotName,
			&contentDataJSON, &htmlPreview, &htmlLength,
			&buildStatus, &lockedAt, &lockedBy, &updatedAt); err != nil {
			h.logger.Warn("Failed to scan component", zap.Error(err))
			continue
		}

		var contentData interface{}
		if len(contentDataJSON) > 0 {
			json.Unmarshal(contentDataJSON, &contentData)
		}

		comp := map[string]interface{}{
			"id":           id.String(),
			"position":     position,
			"slot_name":    slotName,
			"content_data": contentData,
			"html_preview": "",
			"html_length":  htmlLength,
			"build_status": buildStatus,
			"locked":       lockedAt.Valid,
			"locked_by":    lockedBy,
			"is_empty":     htmlLength < 50,
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
			"id":        pageID.String(),
			"name":      pageName,
			"title":     pageTitle,
			"url":       pageURL,
			"page_type": pageType,
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
				site_id, source, domain, item_type, severity, summary,
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
			site_id, source, domain, item_type, severity, summary,
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
