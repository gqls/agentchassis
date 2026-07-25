// FILE: platform/orchestration/actions/report_request_pull_action.go
//
// Two actions for the gripper dossier pilot's cluster⇄island boundary
// (DESIGN_2026-07-24_gripper_dossier_pilot.md §3, §5):
//
//   pull_report_requests — ONE-WAY PULL from the island (the cluster is
//     never called by the public; precedent: intent_collector_actions.go).
//     For each site with deploy_config.report_island, GETs
//     {base_url}/requests?since=<checkpoint> with X-Internal-Key, parses the
//     NDJSON stream ({id, host, submitted_at, spec} lines + a trailing
//     {"_meta":...}), and inserts one report_request work item per line.
//     The island's payload carries NO visitor email — PII stays on the
//     island by design.
//
//   emit_report_status_files — tiny deterministic helper producing the
//     status sidecar file map for git_commit ({"reports/<id>.json": ...}).
//     Exists because the failure branch must publish a files map even when
//     create_report_page never ran, and no existing action synthesizes a
//     dynamic-key files map. The island polls exactly this sidecar; the
//     ready sidecar is committed AFTER the page commit so "ready" can never
//     precede the artefact (gauntlet misstep-4 lesson: status ≠ artefact).

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpecs
// ============================================================================

var PullReportRequestsInputSpec = datahelpers.ActionInputSpec{
	Required: []string{}, Optional: []string{},
	Defaults: map[string]interface{}{}, Deprecated: map[string]string{},
}

var EmitReportStatusFilesInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"request_id"},
	Optional:   []string{"page_url"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("pull_report_requests", PullReportRequestsInputSpec)
	datahelpers.RegisterActionInputSpec("emit_report_status_files", EmitReportStatusFilesInputSpec)
}

// ============================================================================
// ACTION: pull_report_requests
// ============================================================================

type reportIslandSite struct {
	siteID  string
	domain  string
	baseURL string
	pullKey string
}

type pulledReportRequest struct {
	ID          string                 `json:"id"`
	Host        string                 `json:"host"`
	SubmittedAt string                 `json:"submitted_at"`
	Spec        map[string]interface{} `json:"spec"`
	Meta        map[string]interface{} `json:"_meta,omitempty"`
}

func PullReportRequestsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "pull_report_requests"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	sites, err := loadReportIslandSites(ctx, params.DB)
	if err != nil {
		return nil, fmt.Errorf("loading report-island sites: %w", err)
	}

	sitesOK, sitesErr, totalScanned, totalInserted := 0, 0, 0, 0
	perSite := map[string]interface{}{}
	for _, s := range sites {
		inserted, scanned, pullErr := pullOneIslandSite(ctx, params.DB, logger, s)
		if pullErr != nil {
			sitesErr++
			perSite[s.domain] = map[string]interface{}{"error": pullErr.Error()}
			logger.Warn("pull_report_requests: site failed",
				zap.String("domain", s.domain), zap.Error(pullErr))
			continue
		}
		sitesOK++
		totalScanned += scanned
		totalInserted += inserted
		perSite[s.domain] = map[string]interface{}{"scanned": scanned, "inserted": inserted}
	}

	logger.Info("pull_report_requests: run complete",
		zap.Int("sites_ok", sitesOK), zap.Int("sites_err", sitesErr),
		zap.Int("scanned", totalScanned), zap.Int("inserted", totalInserted))
	return map[string]interface{}{
		"sites_ok": sitesOK, "sites_err": sitesErr,
		"scanned": totalScanned, "inserted": totalInserted, "per_site": perSite,
	}, nil
}

func loadReportIslandSites(ctx context.Context, db *sql.DB) ([]reportIslandSite, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, domain,
		       COALESCE(deploy_config->'report_island'->>'base_url', ''),
		       COALESCE(deploy_config->'report_island'->>'pull_key', '')
		FROM sites
		WHERE status IN ('active', 'deployed')
		  AND deploy_config ? 'report_island'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []reportIslandSite
	for rows.Next() {
		var s reportIslandSite
		if err := rows.Scan(&s.siteID, &s.domain, &s.baseURL, &s.pullKey); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func pullOneIslandSite(ctx context.Context, db *sql.DB, log *zap.Logger, s reportIslandSite) (inserted, scanned int, err error) {
	if s.baseURL == "" || s.pullKey == "" {
		return 0, 0, fmt.Errorf("%s has no deploy_config.report_island.base_url/pull_key", s.domain)
	}

	// Checkpoint: newest submitted_at already held for this site, minus a
	// 2-minute overlap. item_key dedup makes the overlap a no-op, and the
	// NOT-EXISTS guard below keeps completed/failed requests from being
	// re-created even if the checkpoint regresses.
	var since sql.NullTime
	if err := db.QueryRowContext(ctx, `
		SELECT max((spec->>'submitted_at')::timestamptz)
		FROM site_work_items
		WHERE site_id = $1 AND item_type = 'report_request'`, s.siteID).Scan(&since); err != nil {
		return 0, 0, fmt.Errorf("checkpoint: %w", err)
	}

	endpoint := s.baseURL + "/requests"
	if since.Valid {
		endpoint += "?since=" + url.QueryEscape(since.Time.Add(-2*time.Minute).UTC().Format(time.RFC3339))
	}

	// item_key dedup rides idx_swi_dedup (partial unique, non-terminal
	// statuses); the NOT EXISTS spans ALL statuses so a terminal request is
	// never resurrected (DESIGN §3 A1).
	const insert = `
		INSERT INTO site_work_items
		    (site_id, source, pipeline, item_type, severity, summary, spec,
		     priority, handler_agent, status, created_by, item_key, max_attempts)
		SELECT $1, 'report-island', 'reports', 'report_request', 'medium', $2, $3::jsonb,
		       50, 'report-builder', 'awaiting_report', 'pull_report_requests', $4, 1
		WHERE NOT EXISTS (
		    SELECT 1 FROM site_work_items WHERE site_id = $1 AND item_key = $4)
		ON CONFLICT DO NOTHING`

	feedErr := scanInternalNDJSONFeed(ctx, endpoint, s.pullKey, 30*time.Second, func(line []byte) {
		var r pulledReportRequest
		if json.Unmarshal(line, &r) != nil {
			return
		}
		if r.Meta != nil || r.ID == "" || r.Spec == nil {
			return
		}
		scanned++

		// The work item spec = island spec + identifiers the workflow needs.
		spec := make(map[string]interface{}, len(r.Spec)+2)
		for k, v := range r.Spec {
			spec[k] = v
		}
		spec["request_id"] = r.ID
		spec["submitted_at"] = r.SubmittedAt
		specJSON, mErr := json.Marshal(spec)
		if mErr != nil {
			return
		}

		res, execErr := db.ExecContext(ctx, insert,
			s.siteID,
			fmt.Sprintf("Gripper dossier request %s", r.ID),
			string(specJSON),
			"report_request:"+r.ID)
		if execErr != nil {
			log.Warn("pull_report_requests: insert failed",
				zap.String("domain", s.domain), zap.String("request_id", r.ID), zap.Error(execErr))
			return
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	})
	if feedErr != nil {
		var partial *partialStreamError
		if !errors.As(feedErr, &partial) {
			// Transport or status failure: nothing was delivered.
			return 0, 0, fmt.Errorf("%s: %w", s.domain, feedErr)
		}
		// Partial success is fine — the next run resumes from the checkpoint.
		log.Warn("pull_report_requests: stream ended early",
			zap.String("domain", s.domain), zap.Error(feedErr))
	}
	return inserted, scanned, nil
}

// ============================================================================
// ACTION: emit_report_status_files
// ============================================================================

func EmitReportStatusFilesAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "emit_report_status_files"),
		zap.String("step_name", params.ExecutionContext.StepName),
	)
	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	status, _ := params.StepConfig.Config["status"].(string)
	if status != "ready" && status != "failed" {
		return nil, fmt.Errorf(`emit_report_status_files: config "status" must be "ready" or "failed", got %q`, status)
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, EmitReportStatusFilesInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	requestID := inputs.Get("request_id")
	if requestID == "" {
		return nil, fmt.Errorf("emit_report_status_files: request_id is empty")
	}

	payload := map[string]interface{}{
		"status":       status,
		"generated_at": time.Now().UTC().Format(time.RFC3339),
	}
	if status == "ready" {
		pageURL := inputs.Get("page_url")
		if pageURL == "" {
			return nil, fmt.Errorf(`emit_report_status_files: status "ready" requires page_url`)
		}
		payload["url"] = pageURL
	}
	body, _ := json.Marshal(payload)

	logger.Info("emit_report_status_files: sidecar built",
		zap.String("request_id", requestID), zap.String("status", status))
	return map[string]interface{}{
		"files": map[string]interface{}{
			"reports/" + requestID + ".json": string(body),
		},
	}, nil
}
