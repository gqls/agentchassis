// FILE: internal/core-manager/admin/asset_admin_handlers.go
//
// Admin endpoints for browsing and managing site assets (images, logos, etc.):
//   - List assets with deployment status and component references
//   - Update asset metadata (purpose, name)
//   - Check which components reference an asset
//   - Delete an asset

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

type AssetAdminHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewAssetAdminHandlers(db *sql.DB, logger *zap.Logger) *AssetAdminHandlers {
	return &AssetAdminHandlers{db: db, logger: logger}
}

// ============================================================================
// GET /admin/sites/:site_id/assets
// ============================================================================
//
// Lists all assets for a site with deployment status (whether any deployed
// page_component references the asset URL).

func (h *AssetAdminHandlers) HandleListAssets(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT a.id, a.asset_type, COALESCE(a.purpose, ''),
		       COALESCE(a.name, ''), a.url,
		       COALESCE(a.filename, ''), COALESCE(a.mime_type, ''),
		       a.file_size, a.dimensions,
		       COALESCE(a.origin_type, 'uploaded'), COALESCE(a.origin_prompt, ''),
		       COALESCE(a.status, 'active'),
		       a.created_at,
		       EXISTS (
		           SELECT 1 FROM page_components pc
		           JOIN pages p ON pc.page_id = p.id
		           WHERE p.site_id = a.site_id
		             AND pc.rendered_html LIKE '%' || SUBSTRING(a.url FROM '[^/]+$') || '%'
		       ) AS is_deployed,
		       (
		           SELECT COUNT(*) FROM page_components pc
		           JOIN pages p ON pc.page_id = p.id
		           WHERE p.site_id = a.site_id
		             AND pc.rendered_html LIKE '%' || SUBSTRING(a.url FROM '[^/]+$') || '%'
		       ) AS reference_count
		FROM assets a
		WHERE a.site_id = $1
		ORDER BY a.purpose, a.created_at
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var assets []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var assetType, purpose, name, url, filename, mimeType string
		var originType, originPrompt, status string
		var fileSize sql.NullInt64
		var dimensionsJSON []byte
		var createdAt sql.NullTime
		var isDeployed bool
		var refCount int

		if err := rows.Scan(&id, &assetType, &purpose, &name, &url,
			&filename, &mimeType, &fileSize, &dimensionsJSON,
			&originType, &originPrompt, &status, &createdAt,
			&isDeployed, &refCount); err != nil {
			h.logger.Warn("Failed to scan asset", zap.Error(err))
			continue
		}

		var dimensions interface{}
		if len(dimensionsJSON) > 0 {
			json.Unmarshal(dimensionsJSON, &dimensions)
		}

		asset := map[string]interface{}{
			"id":              id.String(),
			"asset_type":      assetType,
			"purpose":         purpose,
			"name":            name,
			"url":             url,
			"filename":        filename,
			"mime_type":       mimeType,
			"origin_type":     originType,
			"origin_prompt":   originPrompt,
			"status":          status,
			"is_deployed":     isDeployed,
			"reference_count": refCount,
			"dimensions":      dimensions,
		}
		if fileSize.Valid {
			asset["file_size"] = fileSize.Int64
		}
		if createdAt.Valid {
			asset["created_at"] = createdAt.Time
		}
		assets = append(assets, asset)
	}

	c.JSON(http.StatusOK, gin.H{"assets": assets, "count": len(assets)})
}

// ============================================================================
// GET /admin/sites/:site_id/assets/:asset_id/references
// ============================================================================
//
// Lists which page_components reference this asset.

func (h *AssetAdminHandlers) HandleAssetReferences(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	assetID, err := uuid.Parse(c.Param("asset_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	// Get the asset URL filename for matching
	var assetURL string
	err = h.db.QueryRowContext(ctx, `
		SELECT url FROM assets WHERE id = $1 AND site_id = $2
	`, assetID, siteID).Scan(&assetURL)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	// Extract filename from URL for matching
	// e.g. "https://cdn.example.com/assets/images/hero-bg.webp" → "hero-bg.webp"
	filename := assetURL
	for i := len(assetURL) - 1; i >= 0; i-- {
		if assetURL[i] == '/' {
			filename = assetURL[i+1:]
			break
		}
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT p.name, pc.slot_name, pc.id::text, pc.locked_at IS NOT NULL as locked
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.rendered_html LIKE '%' || $2 || '%'
		ORDER BY p.name, pc.position
	`, siteID, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var refs []map[string]interface{}
	for rows.Next() {
		var pageName, slotName, compID string
		var locked bool
		if err := rows.Scan(&pageName, &slotName, &compID, &locked); err != nil {
			continue
		}
		refs = append(refs, map[string]interface{}{
			"page_name":    pageName,
			"slot_name":    slotName,
			"component_id": compID,
			"locked":       locked,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"asset_id":   assetID.String(),
		"url":        assetURL,
		"references": refs,
		"count":      len(refs),
	})
}

// ============================================================================
// PATCH /admin/sites/:site_id/assets/:asset_id
// ============================================================================
//
// Update asset metadata (purpose, name, status).

func (h *AssetAdminHandlers) HandleUpdateAsset(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	assetID, err := uuid.Parse(c.Param("asset_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	var body struct {
		Purpose *string `json:"purpose"`
		Name    *string `json:"name"`
		Status  *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setClauses := []string{"updated_at = NOW()"}
	args := []interface{}{}
	argIdx := 1

	if body.Purpose != nil {
		setClauses = append(setClauses, fmt.Sprintf("purpose = $%d", argIdx))
		args = append(args, *body.Purpose)
		argIdx++
	}
	if body.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *body.Name)
		argIdx++
	}
	if body.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *body.Status)
		argIdx++
	}

	if len(args) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no fields to update"})
		return
	}

	query := fmt.Sprintf("UPDATE assets SET %s WHERE id = $%d AND site_id = $%d",
		joinStringSlice(setClauses, ", "), argIdx, argIdx+1)
	args = append(args, assetID, siteID)

	result, err := h.db.ExecContext(ctx, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"updated": true, "asset_id": assetID.String()})
}

// ============================================================================
// DELETE /admin/sites/:site_id/assets/:asset_id
// ============================================================================

func (h *AssetAdminHandlers) HandleDeleteAsset(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	assetID, err := uuid.Parse(c.Param("asset_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid asset_id"})
		return
	}

	// Soft delete — set status to 'deleted'
	result, err := h.db.ExecContext(ctx, `
		UPDATE assets SET status = 'deleted', updated_at = NOW()
		WHERE id = $1 AND site_id = $2
	`, assetID, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "asset not found"})
		return
	}

	h.logger.Info("Asset deleted via admin",
		zap.String("asset_id", assetID.String()))

	c.JSON(http.StatusOK, gin.H{"deleted": true, "asset_id": assetID.String()})
}

func joinStringSlice(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
