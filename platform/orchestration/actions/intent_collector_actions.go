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
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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

// statsEntry is one row of the engine's /stats response.
type statsEntry struct {
	Host   string `json:"host"`
	Visits int64  `json:"visits"`
	Events int64  `json:"events"`
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
		// Pull the visit/event counters (/stats) for the events-per-1k
		// denominator. Non-fatal: a stats failure doesn't fail the site.
		if serr := collectSiteStats(ctx, db, s); serr != nil {
			entry["stats_error"] = serr.Error()
			log.Warn("intent-collector: stats pull failed", zap.String("domain", s.domain), zap.Error(serr))
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

	const upsert = `
		INSERT INTO intent_events
		    (engine_event_id, site_id, host, kind, value, ref_host, landing_query, country, event_created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (engine_event_id) DO NOTHING`

	var sid interface{}
	if s.siteID.Valid {
		sid = s.siteID.String
	}

	feedErr := scanInternalNDJSONFeed(ctx, endpoint, s.statsKey, 30*time.Second, func(line []byte) {
		var ev collectedEvent
		if json.Unmarshal(line, &ev) != nil {
			return
		}
		if ev.Meta != nil || ev.ID == "" || ev.Host == "" {
			return
		}
		scanned++
		res, execErr := db.ExecContext(ctx, upsert,
			ev.ID, sid, ev.Host, ev.Kind, ev.Value, ev.RefHost, ev.LandingQuery, ev.Country, ev.CreatedAt)
		if execErr != nil {
			log.Warn("intent-collector: upsert failed", zap.String("domain", s.domain), zap.Error(execErr))
			return
		}
		if n, _ := res.RowsAffected(); n > 0 {
			inserted++
		}
	})
	if feedErr != nil {
		var partial *partialStreamError
		if !errors.As(feedErr, &partial) {
			return 0, 0, fmt.Errorf("%s: %w", s.domain, feedErr)
		}
		// Partial success is fine — next run resumes from the new checkpoint.
		log.Warn("intent-collector: stream ended early", zap.String("domain", s.domain), zap.Error(feedErr))
	}
	return inserted, scanned, nil
}

// collectSiteStats pulls the engine's /stats and upserts the latest cumulative
// visit/event counts for this site's host into intent_site_stats. /stats
// returns all hosts on the box; we take the row matching this site's domain.
func collectSiteStats(ctx context.Context, db *sql.DB, s backendSite) error {
	if s.baseURL == "" || s.statsKey == "" {
		return fmt.Errorf("%s has no deploy_config.engine.base_url/stats_key", s.domain)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/stats", nil)
	req.Header.Set("X-Internal-Key", s.statsKey)
	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s /stats returned %d", s.domain, resp.StatusCode)
	}
	var rows []statsEntry
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return fmt.Errorf("%s /stats decode: %w", s.domain, err)
	}
	var sid interface{}
	if s.siteID.Valid {
		sid = s.siteID.String
	}
	for _, r := range rows {
		if r.Host != s.domain {
			continue
		}
		_, err := db.ExecContext(ctx, `
			INSERT INTO intent_site_stats (host, site_id, visits, events, observed_at)
			VALUES ($1, $2, $3, $4, now())
			ON CONFLICT (host) DO UPDATE
			SET visits = EXCLUDED.visits, events = EXCLUDED.events,
			    site_id = EXCLUDED.site_id, observed_at = now()`,
			r.Host, sid, r.Visits, r.Events)
		return err
	}
	return nil // host not present in /stats yet (no traffic) — fine
}
