// FILE: internal/core-manager/admin/snapshot_admin_handlers.go
//
// Admin endpoints for site snapshot management (point-in-time revert):
//   - Take a snapshot of current site state
//   - List available snapshots for a site
//   - Get full snapshot detail
//   - Revert a site to a previous snapshot
//
// Routes (added to siteGroup in server.go):
//   POST   /admin/sites/:site_id/snapshots                      → HandleTakeSnapshot
//   GET    /admin/sites/:site_id/snapshots                      → HandleListSnapshots
//   GET    /admin/sites/:site_id/snapshots/:snapshot_id         → HandleGetSnapshot
//   POST   /admin/sites/:site_id/snapshots/:snapshot_id/revert  → HandleRevertSnapshot

package admin

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SnapshotAdminHandlers struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSnapshotAdminHandlers(db *sql.DB, logger *zap.Logger) *SnapshotAdminHandlers {
	return &SnapshotAdminHandlers{db: db, logger: logger}
}

// ============================================================================
// POST /admin/sites/:site_id/snapshots
// ============================================================================

func (h *SnapshotAdminHandlers) HandleTakeSnapshot(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	var body struct {
		Trigger string `json:"trigger"`
		Label   string `json:"label"`
		GitSHA  string `json:"git_commit_sha"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		// Allow empty body — defaults to manual trigger
		body.Trigger = "manual"
	}
	if body.Trigger == "" {
		body.Trigger = "manual"
	}

	var gitSHA *string
	if body.GitSHA != "" {
		gitSHA = &body.GitSHA
	}
	var label *string
	if body.Label != "" {
		label = &body.Label
	}

	var snapshotID uuid.UUID
	err = h.db.QueryRowContext(ctx,
		`SELECT take_site_snapshot($1, $2, $3, $4, $5)`,
		siteID, body.Trigger, gitSHA, label, "admin",
	).Scan(&snapshotID)
	if err != nil {
		h.logger.Error("Failed to take snapshot",
			zap.String("site_id", siteID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Snapshot taken via admin API",
		zap.String("site_id", siteID.String()),
		zap.String("snapshot_id", snapshotID.String()),
		zap.String("trigger", body.Trigger))

	c.JSON(http.StatusCreated, gin.H{
		"id":      snapshotID.String(),
		"site_id": siteID.String(),
		"trigger": body.Trigger,
	})
}

// ============================================================================
// GET /admin/sites/:site_id/snapshots
// ============================================================================

func (h *SnapshotAdminHandlers) HandleListSnapshots(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}

	rows, err := h.db.QueryContext(ctx, `
		SELECT
			id, trigger, label, git_commit_sha,
			jsonb_array_length(spec_snapshot) as spec_count,
			jsonb_array_length(pages_snapshot) as page_count,
			created_at, created_by
		FROM site_snapshots
		WHERE site_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var snapshots []gin.H
	for rows.Next() {
		var (
			id        uuid.UUID
			trigger   string
			label     sql.NullString
			gitSHA    sql.NullString
			specCount int
			pageCount int
			createdAt sql.NullTime
			createdBy string
		)
		if err := rows.Scan(&id, &trigger, &label, &gitSHA,
			&specCount, &pageCount, &createdAt, &createdBy); err != nil {
			continue
		}
		snap := gin.H{
			"id":         id.String(),
			"trigger":    trigger,
			"spec_count": specCount,
			"page_count": pageCount,
			"created_by": createdBy,
		}
		if label.Valid {
			snap["label"] = label.String
		}
		if gitSHA.Valid {
			snap["git_commit_sha"] = gitSHA.String
		}
		if createdAt.Valid {
			snap["created_at"] = createdAt.Time
		}
		snapshots = append(snapshots, snap)
	}

	if snapshots == nil {
		snapshots = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{
		"site_id":   siteID.String(),
		"snapshots": snapshots,
		"count":     len(snapshots),
	})
}

// ============================================================================
// GET /admin/sites/:site_id/snapshots/:snapshot_id
// ============================================================================

func (h *SnapshotAdminHandlers) HandleGetSnapshot(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	snapshotID, err := uuid.Parse(c.Param("snapshot_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid snapshot_id"})
		return
	}

	var (
		trigger   string
		label     sql.NullString
		gitSHA    sql.NullString
		siteRec   []byte
		specSnap  []byte
		pagesSnap []byte
		navSnap   []byte
		compsSnap []byte
		createdAt sql.NullTime
		createdBy string
	)

	err = h.db.QueryRowContext(ctx, `
		SELECT trigger, label, git_commit_sha,
			site_record, spec_snapshot, pages_snapshot, nav_snapshot,
			components_snapshot, created_at, created_by
		FROM site_snapshots
		WHERE id = $1 AND site_id = $2
	`, snapshotID, siteID).Scan(
		&trigger, &label, &gitSHA,
		&siteRec, &specSnap, &pagesSnap, &navSnap,
		&compsSnap, &createdAt, &createdBy,
	)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result := gin.H{
		"id":                  snapshotID.String(),
		"site_id":             siteID.String(),
		"trigger":             trigger,
		"site_record":         json.RawMessage(siteRec),
		"spec_snapshot":       json.RawMessage(specSnap),
		"pages_snapshot":      json.RawMessage(pagesSnap),
		"nav_snapshot":        json.RawMessage(navSnap),
		"components_snapshot": json.RawMessage(compsSnap),
		"created_by":          createdBy,
	}
	if label.Valid {
		result["label"] = label.String
	}
	if gitSHA.Valid {
		result["git_commit_sha"] = gitSHA.String
	}
	if createdAt.Valid {
		result["created_at"] = createdAt.Time
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================================
// POST /admin/sites/:site_id/snapshots/:snapshot_id/revert
// ============================================================================

func (h *SnapshotAdminHandlers) HandleRevertSnapshot(c *gin.Context) {
	ctx := c.Request.Context()
	siteID, err := uuid.Parse(c.Param("site_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid site_id"})
		return
	}
	snapshotID, err := uuid.Parse(c.Param("snapshot_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid snapshot_id"})
		return
	}

	// Verify snapshot belongs to site
	var snapSiteID uuid.UUID
	err = h.db.QueryRowContext(ctx,
		`SELECT site_id FROM site_snapshots WHERE id = $1`, snapshotID,
	).Scan(&snapSiteID)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if snapSiteID != siteID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "snapshot does not belong to this site"})
		return
	}

	// Execute revert
	var resultJSON []byte
	err = h.db.QueryRowContext(ctx,
		`SELECT revert_site_to_snapshot($1, $2)`,
		snapshotID, "admin",
	).Scan(&resultJSON)
	if err != nil {
		h.logger.Error("Revert failed",
			zap.String("snapshot_id", snapshotID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Site reverted via admin API",
		zap.String("site_id", siteID.String()),
		zap.String("snapshot_id", snapshotID.String()))

	c.Data(http.StatusOK, "application/json", resultJSON)
}
