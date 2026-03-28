// FILE: internal/core-manager/admin/tool_admin_handlers.go
//
// Admin endpoints for managing deployed tools on sites:
//   - List deployed tools with page and component context
//   - Remove a tool (deactivate fork, delete page_component and page)
//   - Deploy a library tool to a site (shortcut to what tool-deployer does)
//   - List library tools (catalogue)

package admin

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ToolAdminHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewToolAdminHandlers(db *sql.DB, logger *zap.Logger) *ToolAdminHandlers {
	return &ToolAdminHandlers{db: db, logger: logger}
}

// ============================================================================
// GET /admin/sites/:site_id/tools
// ============================================================================
//
// Lists all deployed tools for a site: the fork, its library source,
// the tool page, and deployment status.

func (h *ToolAdminHandlers) HandleListTools(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT
			cc.id::text          AS fork_id,
			cc.function,
			cc.display_name,
			cc.category,
			COALESCE(cc.description, '')  AS description,
			cc.forked_from::text AS library_id,
			COALESCE(lib.display_name, '') AS library_display_name,
			p.id::text           AS page_id,
			p.name               AS page_name,
			p.url                AS page_url,
			p.build_status,
			pc.id::text          AS page_component_id,
			cc.created_at
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		LEFT JOIN content_components lib ON lib.id = cc.forked_from
		WHERE p.site_id = $1
		  AND cc.component_level = 'tool'
		  AND cc.is_active = true
		ORDER BY cc.display_name
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tools []map[string]interface{}
	for rows.Next() {
		var forkID, function, displayName, category, description string
		var libraryID, libraryDisplayName sql.NullString
		var pageID, pageName, pageURL, buildStatus string
		var pageComponentID string
		var createdAt sql.NullTime

		if err := rows.Scan(
			&forkID, &function, &displayName, &category, &description,
			&libraryID, &libraryDisplayName,
			&pageID, &pageName, &pageURL, &buildStatus,
			&pageComponentID, &createdAt,
		); err != nil {
			h.logger.Warn("Failed to scan tool", zap.Error(err))
			continue
		}

		tool := map[string]interface{}{
			"fork_id":           forkID,
			"function":          function,
			"display_name":      displayName,
			"category":          category,
			"description":       description,
			"page_id":           pageID,
			"page_name":         pageName,
			"page_url":          pageURL,
			"build_status":      buildStatus,
			"page_component_id": pageComponentID,
		}
		if libraryID.Valid {
			tool["library_id"] = libraryID.String
			tool["library_display_name"] = libraryDisplayName.String
			tool["is_fork"] = true
		} else {
			tool["is_fork"] = false
		}
		if createdAt.Valid {
			tool["deployed_at"] = createdAt.Time
		}

		tools = append(tools, tool)
	}

	if tools == nil {
		tools = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"tools": tools, "count": len(tools)})
}

// ============================================================================
// DELETE /admin/sites/:site_id/tools/:function
// ============================================================================
//
// Removes a deployed tool from a site:
//   1. Deletes page_components linking the fork to the page
//   2. Deletes the tool page (cascade removes nav items)
//   3. Deactivates the fork component (keeps for audit)
//   4. Creates a needs_rerender item so the nav updates

func (h *ToolAdminHandlers) HandleRemoveTool(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	function := c.Param("function")
	if function == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "function is required"})
		return
	}

	// Find the fork, page_component, and page
	var forkID, pageID, pageComponentID uuid.UUID
	var pageName string
	err = h.db.QueryRowContext(ctx, `
		SELECT cc.id, p.id, pc.id, p.name
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND cc.function = $2
		  AND cc.component_level = 'tool'
		  AND cc.is_active = true
		LIMIT 1
	`, siteID, function).Scan(&forkID, &pageID, &pageComponentID, &pageName)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("tool %s not found on this site", function)})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Begin transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// 1. Delete page_component
	_, err = tx.ExecContext(ctx,
		`DELETE FROM page_components WHERE id = $1`, pageComponentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove page_component: " + err.Error()})
		return
	}

	// 2. Delete the tool page (CASCADE handles nav items, other page_components)
	_, err = tx.ExecContext(ctx,
		`DELETE FROM pages WHERE id = $1 AND site_id = $2`, pageID, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove tool page: " + err.Error()})
		return
	}

	// 3. Deactivate the fork (keep for audit trail)
	_, err = tx.ExecContext(ctx, `
		UPDATE content_components
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`, forkID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deactivate fork: " + err.Error()})
		return
	}

	// 4. Create rerender item so nav gets updated
	_, err = tx.ExecContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by, item_key
		) VALUES (
			$1, 'admin', 'build', 'needs_rerender', 'medium',
			$2, '{"reason": "tool_removed", "refresh_site_components": true}'::jsonb,
			20, 'rerender-pages', 'triaged', 'admin', $3
		) ON CONFLICT DO NOTHING
	`, siteID,
		fmt.Sprintf("Re-render site after removing tool: %s", function),
		fmt.Sprintf("rerender_tool_removed_%s_%s", function, siteID))
	if err != nil {
		h.logger.Warn("Failed to create rerender item (non-fatal)", zap.Error(err))
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Tool removed via admin",
		zap.String("site_id", siteID.String()),
		zap.String("function", function),
		zap.String("fork_id", forkID.String()),
		zap.String("page_name", pageName))

	c.JSON(http.StatusOK, gin.H{
		"removed":   true,
		"function":  function,
		"fork_id":   forkID.String(),
		"page_id":   pageID.String(),
		"page_name": pageName,
	})
}

// ============================================================================
// POST /admin/sites/:site_id/tools
// ============================================================================
//
// Deploys a library tool to a site by creating a work item for tool-deployer.
// This is the admin equivalent of what tool-suggester does automatically.

func (h *ToolAdminHandlers) HandleDeployTool(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	var body struct {
		ToolComponentID string `json:"tool_component_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	toolID, err := uuid.Parse(body.ToolComponentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tool_component_id"})
		return
	}

	// Verify the library tool exists and has a template
	var toolFunction, toolDisplayName string
	var templateLen int
	err = h.db.QueryRowContext(ctx, `
		SELECT function, display_name, LENGTH(COALESCE(html_template, ''))
		FROM content_components
		WHERE id = $1 AND component_level = 'tool'
		  AND forked_from IS NULL AND is_active = true
	`, toolID).Scan(&toolFunction, &toolDisplayName, &templateLen)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "library tool not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if templateLen == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("tool %s has no HTML template", toolFunction)})
		return
	}

	// Check if already deployed
	var existingFork string
	err = h.db.QueryRowContext(ctx, `
		SELECT cc.id::text
		FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE cc.forked_from = $1 AND p.site_id = $2
		  AND cc.is_active = true
		LIMIT 1
	`, toolID, siteID).Scan(&existingFork)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":   fmt.Sprintf("tool %s is already deployed to this site", toolFunction),
			"fork_id": existingFork,
		})
		return
	}

	// Create work item for tool-deployer
	spec := fmt.Sprintf(`{"tool_component_id": "%s", "name": "%s", "function": "%s"}`,
		toolID.String(), toolDisplayName, toolFunction)
	itemKey := fmt.Sprintf("add_tool:%s:%s", toolFunction, siteID)
	summary := fmt.Sprintf("Deploy tool: %s", toolDisplayName)

	var itemID uuid.UUID
	err = h.db.QueryRowContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, priority, handler_agent, status, created_by, item_key
		) VALUES (
			$1, 'admin', 'build', 'add_tool', 'low', $2,
			$3::jsonb, 80, 'tool-deployer', 'triaged', 'admin', $4
		) ON CONFLICT DO NOTHING
		RETURNING id
	`, siteID, summary, spec, itemKey).Scan(&itemID)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusConflict, gin.H{"error": "deploy work item already exists"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Tool deploy requested via admin",
		zap.String("site_id", siteID.String()),
		zap.String("tool", toolFunction),
		zap.String("item_id", itemID.String()))

	c.JSON(http.StatusCreated, gin.H{
		"queued":        true,
		"work_item_id":  itemID.String(),
		"tool_function": toolFunction,
		"tool_name":     toolDisplayName,
	})
}

// ============================================================================
// GET /admin/tools/library
// ============================================================================
//
// Lists all library tools (unforked originals) with template status.

func (h *ToolAdminHandlers) HandleListLibraryTools(c *gin.Context) {
	ctx := c.Request.Context()

	rows, err := h.db.QueryContext(ctx, `
		SELECT
			cc.id::text, cc.function, cc.display_name, cc.category,
			COALESCE(cc.description, ''),
			LENGTH(COALESCE(cc.html_template, '')) AS template_len,
			cc.semantic_tags::text,
			(SELECT COUNT(*) FROM content_components fork
			 WHERE fork.forked_from = cc.id AND fork.is_active = true
			) AS deploy_count
		FROM content_components cc
		WHERE cc.component_level = 'tool'
		  AND cc.forked_from IS NULL
		  AND cc.is_active = true
		ORDER BY cc.category, cc.display_name
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var tools []map[string]interface{}
	for rows.Next() {
		var id, function, displayName, category, description string
		var templateLen, deployCount int
		var tagsJSON sql.NullString

		if err := rows.Scan(&id, &function, &displayName, &category, &description,
			&templateLen, &tagsJSON, &deployCount); err != nil {
			continue
		}

		tool := map[string]interface{}{
			"id":           id,
			"function":     function,
			"display_name": displayName,
			"category":     category,
			"description":  description,
			"has_template": templateLen > 0,
			"template_len": templateLen,
			"deploy_count": deployCount,
			"tags":         tagsJSON.String,
		}
		tools = append(tools, tool)
	}

	if tools == nil {
		tools = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, gin.H{"tools": tools, "count": len(tools)})
}

// ============================================================================
// domainSlug helper (matches deploy_tool_action.go)
// ============================================================================

func toolDomainSlug(domain string) string {
	if domain == "" {
		return "site"
	}
	return strings.ReplaceAll(domain, ".", "-")
}
