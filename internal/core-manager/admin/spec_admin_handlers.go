// FILE: internal/core-manager/admin/spec_admin_handlers.go
//
// Admin endpoints for viewing and managing site specs (direction control):
//   - List all current specs with pinned status
//   - Pin/unpin individual spec aspects
//   - Propagate spec changes into targeted work items

package admin

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SpecAdminHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSpecAdminHandlers(db *sql.DB, logger *zap.Logger) *SpecAdminHandlers {
	return &SpecAdminHandlers{db: db, logger: logger}
}

// ============================================================================
// GET /admin/sites/:site_id/specs
// ============================================================================
//
// Returns all current specs for a site, grouped by aspect, with pinned status.

func (h *SpecAdminHandlers) HandleListSpecs(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT id, aspect, data, COALESCE(source, ''), COALESCE(source_agent, ''),
		       COALESCE(created_by, ''), COALESCE(pinned, false),
		       created_at
		FROM site_specs
		WHERE site_id = $1 AND is_current = true
		ORDER BY aspect
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var specs []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var aspect, source, sourceAgent, createdBy string
		var pinned bool
		var dataJSON []byte
		var createdAt sql.NullTime

		if err := rows.Scan(&id, &aspect, &dataJSON, &source, &sourceAgent,
			&createdBy, &pinned, &createdAt); err != nil {
			h.logger.Warn("Failed to scan spec", zap.Error(err))
			continue
		}

		var data interface{}
		json.Unmarshal(dataJSON, &data)

		spec := map[string]interface{}{
			"id":           id.String(),
			"aspect":       aspect,
			"data":         data,
			"source":       source,
			"source_agent": sourceAgent,
			"created_by":   createdBy,
			"pinned":       pinned,
		}
		if createdAt.Valid {
			spec["created_at"] = createdAt.Time
		}
		specs = append(specs, spec)
	}

	c.JSON(http.StatusOK, gin.H{"specs": specs, "count": len(specs)})
}

// ============================================================================
// POST /admin/sites/:site_id/specs/:aspect/pin
// ============================================================================

func (h *SpecAdminHandlers) HandlePinSpec(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	aspect := c.Param("aspect")

	result, err := h.db.ExecContext(ctx, `
		UPDATE site_specs SET pinned = true
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "spec not found"})
		return
	}

	h.logger.Info("Spec pinned via admin",
		zap.String("site_id", siteID.String()),
		zap.String("aspect", aspect))

	c.JSON(http.StatusOK, gin.H{"pinned": true, "aspect": aspect})
}

// ============================================================================
// POST /admin/sites/:site_id/specs/:aspect/unpin
// ============================================================================

func (h *SpecAdminHandlers) HandleUnpinSpec(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	aspect := c.Param("aspect")

	result, err := h.db.ExecContext(ctx, `
		UPDATE site_specs SET pinned = false
		WHERE site_id = $1 AND aspect = $2 AND is_current = true
	`, siteID, aspect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "spec not found"})
		return
	}

	h.logger.Info("Spec unpinned via admin",
		zap.String("site_id", siteID.String()),
		zap.String("aspect", aspect))

	c.JSON(http.StatusOK, gin.H{"pinned": false, "aspect": aspect})
}

// ============================================================================
// POST /admin/sites/:site_id/specs/:aspect/propagate
// ============================================================================
//
// Creates targeted work items to propagate a spec change across the site.
// The admin chooses scope (all pages, or specific pages) and the item type
// to create. Returns the list of created items for review.

func (h *SpecAdminHandlers) HandlePropagateSpec(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	aspect := c.Param("aspect")

	var body struct {
		ItemType     string   `json:"item_type"`
		HandlerAgent string   `json:"handler_agent"`
		Severity     string   `json:"severity"`
		Priority     int      `json:"priority"`
		Pages        []string `json:"pages"` // empty = all content pages
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Defaults
	if body.ItemType == "" {
		body.ItemType = "content_rewrite"
	}
	if body.HandlerAgent == "" {
		body.HandlerAgent = "page-build-handler"
	}
	if body.Severity == "" {
		body.Severity = "medium"
	}
	if body.Priority == 0 {
		body.Priority = 30
	}

	// Determine scope based on aspect type
	scopeQuery := `
		SELECT p.id, p.name FROM pages p
		WHERE p.site_id = $1
		  AND p.status IN ('active', 'deployed')
		  AND COALESCE(p.page_type, '') NOT IN ('blog-index')
	`
	args := []interface{}{siteID}

	// If specific pages requested, filter
	if len(body.Pages) > 0 {
		placeholders := make([]string, len(body.Pages))
		for i, name := range body.Pages {
			args = append(args, name)
			placeholders[i] = fmt.Sprintf("$%d", i+2)
		}
		scopeQuery += fmt.Sprintf(" AND p.name IN (%s)", joinStrings(placeholders, ","))
	}

	// For design aspects, include all pages. For identity/tone, skip blog posts.
	if aspect == "identity" || aspect == "tone" || aspect == "audience" {
		scopeQuery += " AND COALESCE(p.page_type, '') NOT IN ('blog-post')"
	}

	scopeQuery += " ORDER BY p.name"

	rows, err := h.db.QueryContext(ctx, scopeQuery, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var created []map[string]interface{}
	var skipped int

	for rows.Next() {
		var pageID uuid.UUID
		var pageName string
		if err := rows.Scan(&pageID, &pageName); err != nil {
			continue
		}

		// Check if all components on this page are locked
		var lockedCount, totalCount int
		h.db.QueryRowContext(ctx, `
			SELECT COUNT(*),
			       COUNT(*) FILTER (WHERE locked_at IS NOT NULL AND locked_by IN ('admin', 'admin-removed', 'checkpoint'))
			FROM page_components
			WHERE page_id = $1 AND build_status != 'removed'
			  AND COALESCE(slot_name, '') NOT IN ('header', 'footer', 'head')
		`, pageID).Scan(&totalCount, &lockedCount)

		if totalCount > 0 && lockedCount == totalCount {
			skipped++
			continue // All components locked, skip this page
		}

		specJSON, _ := json.Marshal(map[string]interface{}{
			"page_name":      pageName,
			"spec_aspect":    aspect,
			"reason":         "spec_propagation",
			"source":         "admin-propagate",
			"locked_count":   lockedCount,
			"unlocked_count": totalCount - lockedCount,
		})

		summary := fmt.Sprintf("Update %s to reflect new %s direction", pageName, aspect)
		itemKey := fmt.Sprintf("propagate_%s_%s_%s", aspect, pageName, siteID)

		var newID uuid.UUID
		err = h.db.QueryRowContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by,
				item_key
			) VALUES ($1, 'admin-propagate', 'build', $2, $3, $4,
			          $5::jsonb, $6, $7, $8, 'triaged', 'admin', $9)
			ON CONFLICT DO NOTHING
			RETURNING id
		`, siteID, body.ItemType, body.Severity, summary,
			string(specJSON), pageID, body.Priority, body.HandlerAgent,
			itemKey).Scan(&newID)

		if err == sql.ErrNoRows {
			// ON CONFLICT — duplicate, skip
			continue
		}
		if err != nil {
			h.logger.Warn("Failed to create propagation item",
				zap.String("page", pageName), zap.Error(err))
			continue
		}

		created = append(created, map[string]interface{}{
			"id":        newID.String(),
			"page_name": pageName,
			"item_type": body.ItemType,
		})
	}

	h.logger.Info("Spec propagation complete",
		zap.String("aspect", aspect),
		zap.Int("created", len(created)),
		zap.Int("skipped_locked", skipped))

	c.JSON(http.StatusOK, gin.H{
		"propagated":    true,
		"aspect":        aspect,
		"items_created": len(created),
		"items_skipped": skipped,
		"items":         created,
	})
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
