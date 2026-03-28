// FILE: internal/core-manager/admin/confirm_work_item_handler.go
//
// POST /admin/work-items/:item_id/confirm
//
// Converts a needs_human_review item into a follow-up work item.
// The admin reviews the finding and confirms it should be actioned.
//
// For tool-auditor findings: creates improve_tool → tool-improver
// For other review sources: creates the item type specified in the request
//
// The original review item is marked complete with resolution details.
//
// Request body:
//   {
//     "follow_up_type": "improve_tool",        // optional — defaults from spec
//     "follow_up_handler": "tool-improver",     // optional — defaults from spec
//     "issue_override": "edited description",   // optional — override the issue text
//     "severity_override": "high",              // optional — override severity
//     "notes": "Admin notes"                    // optional
//   }
//
// The endpoint reads component_id, issue, fix_suggestion, page_id, page_name
// from the review item's spec and passes them through to the follow-up item.

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

// ============================================================================
// POST /admin/work-items/:item_id/confirm
// ============================================================================

func (h *SiteAdminHandlers) HandleConfirmWorkItem(c *gin.Context) {
	ctx := c.Request.Context()
	itemID, err := uuid.Parse(c.Param("item_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
		return
	}

	var body struct {
		FollowUpType     string `json:"follow_up_type"`
		FollowUpHandler  string `json:"follow_up_handler"`
		IssueOverride    string `json:"issue_override"`
		SeverityOverride string `json:"severity_override"`
		Notes            string `json:"notes"`
	}
	c.ShouldBindJSON(&body)

	// Load the review item
	var siteID uuid.UUID
	var specJSON []byte
	var currentStatus, summary, source string
	var pageID sql.NullString
	var severity string

	err = h.db.QueryRowContext(ctx, `
		SELECT site_id, spec, status, summary, COALESCE(source, ''),
		       page_id::text, COALESCE(severity, 'medium')
		FROM site_work_items
		WHERE id = $1
	`, itemID).Scan(&siteID, &specJSON, &currentStatus, &summary, &source, &pageID, &severity)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "work item not found"})
		return
	}

	if currentStatus != "needs_human_review" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("item status is '%s', expected 'needs_human_review'", currentStatus),
		})
		return
	}

	// Parse the spec
	var spec map[string]interface{}
	if err := json.Unmarshal(specJSON, &spec); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse work item spec"})
		return
	}

	// Determine follow-up item type
	followUpType := body.FollowUpType
	if followUpType == "" {
		// Infer from the review source
		if check, ok := spec["check"].(string); ok {
			switch check {
			case "tool_auditor", "tool_health":
				followUpType = "improve_tool"
			default:
				followUpType = "content_rewrite"
			}
		} else {
			followUpType = "content_rewrite"
		}
	}

	followUpHandler := body.FollowUpHandler
	if followUpHandler == "" {
		switch followUpType {
		case "improve_tool":
			followUpHandler = "tool-improver"
		case "content_rewrite", "needs_content_page":
			followUpHandler = "page-build-handler"
		default:
			followUpHandler = "page-build-handler"
		}
	}

	// Build follow-up spec from the review item's data
	followUpSpec := map[string]interface{}{
		"confirmed_from": itemID.String(),
		"source":         "admin-confirm",
	}

	// Pass through relevant fields from the review spec
	for _, key := range []string{
		"component_id", "issue", "fix_suggestion", "category",
		"page_id", "page_name", "tool_function",
		"page_name", "section_name",
	} {
		if val, ok := spec[key]; ok {
			followUpSpec[key] = val
		}
	}

	// Apply overrides
	if body.IssueOverride != "" {
		followUpSpec["issue"] = body.IssueOverride
	}

	followUpSeverity := severity
	if body.SeverityOverride != "" {
		followUpSeverity = body.SeverityOverride
	}

	followUpSummary := summary
	if body.IssueOverride != "" {
		followUpSummary = body.IssueOverride
	}

	// Priority based on severity
	priority := 60
	switch followUpSeverity {
	case "high":
		priority = 30
	case "low":
		priority = 80
	}

	followUpSpecJSON, _ := json.Marshal(followUpSpec)

	itemKey := fmt.Sprintf("confirmed_%s_%s", followUpType, itemID.String()[:8])

	// Begin transaction
	tx, err := h.db.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// 1. Create the follow-up work item
	var followUpID uuid.UUID
	followUpPageID := sql.NullString{}
	if pid, ok := followUpSpec["page_id"].(string); ok && pid != "" {
		followUpPageID = sql.NullString{String: pid, Valid: true}
	}

	err = tx.QueryRowContext(ctx, `
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary,
			spec, page_id, priority, handler_agent, status, created_by,
			item_key, parent_item_id
		) VALUES (
			$1, 'admin-confirm', 'build', $2, $3, $4,
			$5::jsonb, $6::uuid, $7, $8, 'triaged', 'admin',
			$9, $10
		)
		ON CONFLICT DO NOTHING
		RETURNING id
	`, siteID, followUpType, followUpSeverity, followUpSummary,
		string(followUpSpecJSON), followUpPageID, priority, followUpHandler,
		itemKey, itemID).Scan(&followUpID)

	if err == sql.ErrNoRows {
		// Duplicate — item_key conflict, item already confirmed
		c.JSON(http.StatusConflict, gin.H{"error": "follow-up item already exists for this finding"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create follow-up: " + err.Error()})
		return
	}

	// 2. Mark the review item complete
	resolution := fmt.Sprintf("Confirmed → %s item %s", followUpType, followUpID.String()[:8])
	if body.Notes != "" {
		resolution += " | " + body.Notes
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete',
		    completed_at = NOW(),
		    result = jsonb_build_object(
		        'resolution', $2,
		        'resolved_by', 'admin',
		        'follow_up_id', $3::text,
		        'follow_up_type', $4
		    ),
		    updated_at = NOW()
		WHERE id = $1
	`, itemID, resolution, followUpID.String(), followUpType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update review item: " + err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Review item confirmed via admin",
		zap.String("review_item_id", itemID.String()),
		zap.String("follow_up_id", followUpID.String()),
		zap.String("follow_up_type", followUpType),
		zap.String("handler", followUpHandler))

	c.JSON(http.StatusCreated, gin.H{
		"confirmed":         true,
		"review_item_id":    itemID.String(),
		"follow_up_id":      followUpID.String(),
		"follow_up_type":    followUpType,
		"follow_up_handler": followUpHandler,
		"severity":          followUpSeverity,
	})
}
