// FILE: platform/orchestration/actions/site_snapshot_actions.go
//
// Actions for site snapshot management:
//   - take_site_snapshot: capture current site state
//   - revert_site_snapshot: restore site to a previous snapshot
//   - list_site_snapshots: list available snapshots for a site
//
// Workflow usage (take snapshot after deploy):
//
//   "snapshot_site": {
//       "action": "take_site_snapshot",
//       "config": {
//           "site_id_field": "site_record.site_id",
//           "trigger": "deploy",
//           "git_sha_field": "page_deployed.commit_sha",
//           "label": "Post-deploy snapshot"
//       },
//       "output_field": "snapshot_result",
//       "next_step": "complete"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ACTION: take_site_snapshot
// ============================================================================

// TakeSiteSnapshotAction captures the current state of a site.
// Calls the take_site_snapshot SQL function which captures site_specs,
// pages, page_components, navigation, and site_components into a single
// self-contained JSONB snapshot row.
//
// Config:
//   - site_id_field: path to site_id in collected_data (default: "site_record.site_id")
//   - trigger: snapshot trigger type — "deploy", "manual", "pre_edit", "scheduled"
//   - git_sha_field: path to git commit SHA in collected_data (optional)
//   - git_sha: literal git SHA string (optional, git_sha_field takes precedence)
//   - label: optional human-readable label
func TakeSiteSnapshotAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("TakeSiteSnapshotAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Extract site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Extract trigger
	trigger := "manual"
	if t, ok := config["trigger"].(string); ok && t != "" {
		trigger = t
	}

	// Extract git SHA (from field or literal)
	var gitSHA *string
	if shaField, ok := config["git_sha_field"].(string); ok && shaField != "" {
		if sha := datahelpers.ExtractNestedFieldString(params.CollectedData, shaField); sha != "" {
			gitSHA = &sha
		}
	}
	if gitSHA == nil {
		if sha, ok := config["git_sha"].(string); ok && sha != "" {
			gitSHA = &sha
		}
	}

	// Extract label
	var label *string
	if l, ok := config["label"].(string); ok && l != "" {
		label = &l
	}

	// Created by
	createdBy := params.AgentType
	if createdBy == "" {
		createdBy = "system"
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	// Call the SQL function
	var snapshotID uuid.UUID
	err = params.DB.QueryRowContext(ctx,
		`SELECT take_site_snapshot($1, $2, $3, $4, $5)`,
		siteID, trigger, gitSHA, label, createdBy,
	).Scan(&snapshotID)
	if err != nil {
		return nil, fmt.Errorf("take_site_snapshot failed: %w", err)
	}

	params.Logger.Info("TakeSiteSnapshotAction: Snapshot created",
		zap.String("snapshot_id", snapshotID.String()),
		zap.String("site_id", siteIDStr),
		zap.String("trigger", trigger),
	)

	return map[string]interface{}{
		"snapshot_id": snapshotID.String(),
		"site_id":     siteIDStr,
		"trigger":     trigger,
		"created_by":  createdBy,
		"success":     true,
	}, nil
}

// ============================================================================
// ACTION: revert_site_snapshot
// ============================================================================

// RevertSiteSnapshotAction restores a site to a previous snapshot.
// Calls the revert_site_to_snapshot SQL function which:
//   - Takes a safety snapshot first (trigger='pre_revert')
//   - Replaces site_specs, pages, page_components, navigation, site_components
//   - Updates site record fields
//
// Config:
//   - snapshot_id_field: path to snapshot_id in collected_data
//   - snapshot_id: literal snapshot UUID (snapshot_id_field takes precedence)
func RevertSiteSnapshotAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RevertSiteSnapshotAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Extract snapshot_id
	var snapshotIDStr string
	if field, ok := config["snapshot_id_field"].(string); ok && field != "" {
		snapshotIDStr = datahelpers.ExtractNestedFieldString(params.CollectedData, field)
	}
	if snapshotIDStr == "" {
		if id, ok := config["snapshot_id"].(string); ok && id != "" {
			snapshotIDStr = id
		}
	}
	if snapshotIDStr == "" {
		return nil, fmt.Errorf("snapshot_id not provided (set snapshot_id_field or snapshot_id)")
	}

	snapshotID, err := uuid.Parse(snapshotIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid snapshot_id: %w", err)
	}

	revertedBy := params.AgentType
	if revertedBy == "" {
		revertedBy = "admin"
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	// Call the SQL function — returns JSONB with revert details
	var resultJSON []byte
	err = params.DB.QueryRowContext(ctx,
		`SELECT revert_site_to_snapshot($1, $2)`,
		snapshotID, revertedBy,
	).Scan(&resultJSON)
	if err != nil {
		return nil, fmt.Errorf("revert_site_to_snapshot failed: %w", err)
	}

	// Parse the result
	var result map[string]interface{}
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, fmt.Errorf("failed to parse revert result: %w", err)
	}

	params.Logger.Info("RevertSiteSnapshotAction: Revert complete",
		zap.String("snapshot_id", snapshotIDStr),
		zap.Any("result", result),
	)

	return result, nil
}

// ============================================================================
// ACTION: list_site_snapshots
// ============================================================================

// ListSiteSnapshotsAction returns available snapshots for a site.
//
// Config:
//   - site_id_field: path to site_id in collected_data (default: "site_record.site_id")
//   - limit: max rows to return (default: 20)
func ListSiteSnapshotsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ListSiteSnapshotsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	limit := 20
	if l, ok := config["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	if params.DB == nil {
		return nil, fmt.Errorf("no database connection available")
	}

	rows, err := params.DB.QueryContext(ctx, `
		SELECT
			id, trigger, label, git_commit_sha,
			jsonb_array_length(spec_snapshot) as spec_count,
			jsonb_array_length(pages_snapshot) as page_count,
			created_at, created_by
		FROM site_snapshots
		WHERE site_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, siteID, limit)
	if err != nil {
		return nil, fmt.Errorf("query snapshots: %w", err)
	}
	defer rows.Close()

	var snapshots []map[string]interface{}
	for rows.Next() {
		var (
			id        uuid.UUID
			trigger   string
			label     sql.NullString
			gitSHA    sql.NullString
			specCount int
			pageCount int
			createdAt string
			createdBy string
		)
		if err := rows.Scan(&id, &trigger, &label, &gitSHA, &specCount, &pageCount, &createdAt, &createdBy); err != nil {
			params.Logger.Warn("ListSiteSnapshotsAction: scan error", zap.Error(err))
			continue
		}
		snap := map[string]interface{}{
			"id":         id.String(),
			"trigger":    trigger,
			"spec_count": specCount,
			"page_count": pageCount,
			"created_at": createdAt,
			"created_by": createdBy,
		}
		if label.Valid {
			snap["label"] = label.String
		}
		if gitSHA.Valid {
			snap["git_commit_sha"] = gitSHA.String
		}
		snapshots = append(snapshots, snap)
	}

	if snapshots == nil {
		snapshots = []map[string]interface{}{}
	}

	return map[string]interface{}{
		"site_id":   siteIDStr,
		"snapshots": snapshots,
		"count":     len(snapshots),
	}, nil
}
