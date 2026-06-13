// FILE: platform/orchestration/actions/intent_collector_actions.go
// P4 — pull new intent events from EVERY VM-hosted backend site's /events
// endpoint into intent_events. One action, self-querying, loops all backend
// sites (complexity in Go; the workflow stays a single step). Invoked by the
// 'intent-collection' scheduled task via the intent-collector agent.
//
// Verified against the live chassis source (production context dump):
//   - ActionParams has DB *sql.DB and Logger *zap.Logger (used here).
//   - Registry is GlobalActionRegistry map[string]ActionDefinition; the entry
//     for this action is in registry_insertion below.
//   - The scheduler fires ONE message per tick (no per-row fan-out), so the
//     action loops sites itself rather than relying on pre_query fan-out — the
//     build-pipeline-trigger gate + thunder-monitor in-agent-loop pattern.
// No datahelpers dependency: inputs come from the DB query, not CollectedData.

package actions

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// collectedEvent matches one NDJSON line from the engine's /events stream.
// The final line is a {"_meta":{...}} object and is skipped.
type collectedEvent struct {
	ID           string    `json:"id"`
	Host         string    `json:"host"`
	Kind         string    `json:"kind"`
	Value        string    `json:"value"`
	RefHost      string    `json:"ref_host"`
	LandingQuery string    `json:"landing_query"`
	Country      string    `json:"country"`
	CreatedAt    time.Time `json:"created_at"`
	Meta         *struct {
		Count     int  `json:"count"`
		Truncated bool `json:"truncated"`
	} `json:"_meta,omitempty"`
}

type backendSite struct {
	siteID   sql.NullString
	domain   string
	baseURL  string
	statsKey string
}

// CollectIntentEventsAction collects new events from all VM-hosted backend
// sites. Per-site failures are logged and skipped; the run continues.
func CollectIntentEventsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	db := params.DB
	log := params.Logger

	sites, err := loadBackendSites(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("intent-collector: list backend sites: %w", err)
	}

	totalInserted, totalScanned, sitesOK, sitesErr := 0, 0, 0, 0
	perSite := make([]map[string]interface{}, 0, len(sites))

	for _, s := range sites {
		ins, scn, err := collectOneSite(ctx, db, log, s)
		entry := map[string]interface{}{"domain": s.domain, "inserted": ins, "scanned": scn}
		if err != nil {
			sitesErr++
			entry["error"] = err.Error()
			log.Warn("intent-collector: site failed", zap.String("domain", s.domain), zap.Error(err))
		} else {
			sitesOK++
		}
		totalInserted += ins
		totalScanned += scn
		perSite = append(perSite, entry)
	}

	log.Info("intent-collector: run complete",
		zap.Int("sites_ok", sitesOK), zap.Int("sites_err", sitesErr),
		zap.Int("scanned", totalScanned), zap.Int("inserted", totalInserted))

	return map[string]interface{}{
		"sites_ok": sitesOK, "sites_err": sitesErr,
		"scanned": totalScanned, "inserted": totalInserted,
		"per_site": perSite,
	}, nil
}

func loadBackendSites(ctx context.Context, db *sql.DB) ([]backendSite, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, domain,
		       COALESCE(deploy_config->'engine'->>'base_url',  ''),
		       COALESCE(deploy_config->'engine'->>'stats_key', '')
		FROM sites
		WHERE status IN ('active', 'deployed')
		  AND deploy_config->>'target' = 'vm'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []backendSite
	for rows.Next() {
		var s backendSite
		if err := rows.Scan(&s.siteID, &s.domain, &s.baseURL, &s.statsKey); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func collectOneSite(ctx context.Context, db *sql.DB, log *zap.Logger, s backendSite) (inserted, scanned int, err error) {
	if s.baseURL == "" || s.statsKey == "" {
		return 0, 0, fmt.Errorf("%s has no deploy_config.engine.base_url/stats_key", s.domain)
	}

	// Checkpoint: newest event already held for this host, minus a safety
	// margin. The unique engine_event_id makes the overlap a no-op.
	var since sql.NullTime
	if err := db.QueryRowContext(ctx,
		`SELECT max(event_created_at) FROM intent_events WHERE host = $1`, s.domain).Scan(&since); err != nil {
		return 0, 0, fmt.Errorf("checkpoint: %w", err)
	}

	endpoint := s.baseURL + "/events?host=" + url.QueryEscape(s.domain)
	if since.Valid {
		endpoint += "&since=" + url.QueryEscape(since.Time.Add(-2*time.Minute).UTC().Format(time.RFC3339))
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	req.Header.Set("X-Internal-Key", s.statsKey)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("%s returned %d", s.domain, resp.StatusCode)
	}

	const upsert = `
		INSERT INTO intent_events
		    (engine_event_id, site_id, host, kind, value, ref_host, landing_query, country, event_created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (engine_event_id) DO NOTHING`

	var sid interface{}
	if s.siteID.Valid {
		sid = s.siteID.String
	}

	sc := bufio.NewScanner(resp.Body)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		var ev collectedEvent
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		if ev.Meta != nil || ev.ID == "" || ev.Host == "" {
			continue
		}
		scanned++
		res, execErr := db.ExecContext(ctx, upsert,
			ev.ID, sid, ev.Host, ev.Kind, ev.Value, ev.RefHost, ev.LandingQuery, ev.Country, ev.CreatedAt)
		if execErr != nil {
			log.Warn("intent-collector: upsert failed", zap.String("domain", s.domain), zap.Error(execErr))
			continue
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	}
	if scErr := sc.Err(); scErr != nil {
		// Partial success is fine — next run resumes from the new checkpoint.
		log.Warn("intent-collector: stream ended early", zap.String("domain", s.domain), zap.Error(scErr))
	}
	return inserted, scanned, nil
}
