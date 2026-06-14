// FILE: platform/orchestration/actions/discovery_checks/check_backend_unreachable.go
// Discovery check: backend_unreachable — per-site health probe for VM-hosted
// backend sites (deploy_config.target = 'vm'). Runs on the improvement sweep
// like the other checks; NOOPS for static (B2/Worker) sites. Self-clearing:
// when the box recovers it resolves its own open item.
//
// Pattern source: check_tool_health.go / check_missing_tools.go (per-site
// structural checks that emit site_work_items at 'detected'). This is an
// ALERT, not an auto-fix: a down VM isn't a chassis-fixable content problem, so
// the item carries NO handler_agent and simply stays visible until the box is
// back (or, later, until the P5 vmhost adapter becomes its handler_agent).
//
// VERIFY against discovery_checks/check_tool_health.go before wiring (the dump
// didn't contain the literal interface):
//   [I1] the DiscoveryCheck interface method set + exact signature — this file
//        assumes per-site `Name() string` and
//        `Run(ctx, db *sql.DB, siteID string, log *zap.Logger) error`, matching
//        the per-site sweep. Adjust names/return to the real interface.
//   [I2] self-registration mechanism — assumes an init() that appends to the
//        package registry (as check_sectionless_pages.go "self-registers").
//   [I3] the emit helper — checks insert via the shared work-item path with
//        7-day item_key dedup. This file inserts directly to show the columns;
//        swap to the package's insertWorkItem/emit helper if one is exposed to
//        checks (keeps the 3h/7d terminal-dedup behaviour).
//
// Enablement (not code): add "backend_unreachable" to the discovery agent's
// default_config checks array (the same agent the improvement sweep targets),
// per the check_sectionless_pages enablement note.

package discovery_checks

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

const backendUnreachableItemKey = "backend_unreachable"

func init() {
	Register(backendUnreachableCheck{}) // [I2] confirm registration fn name
}

type backendUnreachableCheck struct{}

func (backendUnreachableCheck) Name() string { return "backend_unreachable" } // [I1]

// Run probes one site if it is a VM backend; emits or clears the alert.
func (backendUnreachableCheck) Run(ctx context.Context, db *sql.DB, siteID string, log *zap.Logger) error {
	var domain, target string
	err := db.QueryRowContext(ctx,
		`SELECT domain, COALESCE(deploy_config->>'target', '') FROM sites WHERE id = $1`, siteID).
		Scan(&domain, &target)
	if err != nil {
		return fmt.Errorf("backend_unreachable: load site %s: %w", siteID, err)
	}
	if target != "vm" {
		return nil // not a backend site — nothing to probe
	}

	reachable, detail := probeHealth(ctx, domain)
	if reachable {
		return clearOpenAlert(ctx, db, siteID, log)
	}

	// Unreachable → emit one open alert (the partial unique index on
	// (site_id, item_key) keeps it single while non-terminal).
	log.Warn("backend_unreachable: probe failed", zap.String("domain", domain), zap.String("detail", detail))
	return emitAlert(ctx, db, siteID, domain, detail, log)
}

func probeHealth(ctx context.Context, domain string) (ok bool, detail string) {
	url := "https://" + domain + "/health"
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(cctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "request error: " + err.Error()
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("status %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	if !strings.Contains(string(body), "\"ok\":true") {
		return false, "unexpected /health body"
	}
	return true, ""
}

// emitAlert inserts a detected backend_unreachable item if no open one exists.
// [I3] Prefer the package's shared insert helper (3h/7d terminal dedup) if it's
// callable from checks; this direct insert relies on the partial unique index
// idx_swi_dedup to prevent duplicate non-terminal items.
func emitAlert(ctx context.Context, db *sql.DB, siteID, domain, detail string, log *zap.Logger) error {
	spec, _ := json.Marshal(map[string]interface{}{
		"domain": domain, "probe": "https://" + domain + "/health", "detail": detail,
	})
	_, err := db.ExecContext(ctx, `
		INSERT INTO site_work_items
		    (site_id, source, item_type, severity, summary, spec, affected_url,
		     handler_agent, status, created_by, item_key, pipeline)
		VALUES ($1, 'discovery', 'backend_unreachable', 'high', $2, $3, $4,
		        NULL, 'detected', 'backend-unreachable-check', $5, 'build')
		ON CONFLICT (site_id, item_key)
		    WHERE item_key IS NOT NULL
		      AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved'])
		DO NOTHING`,
		siteID,
		fmt.Sprintf("Backend unreachable for %s (%s) — engine/nginx health probe failed", domain, detail),
		spec,
		"https://"+domain+"/health",
		backendUnreachableItemKey,
	)
	if err != nil {
		return fmt.Errorf("backend_unreachable: emit for %s: %w", domain, err)
	}
	return nil
}

// clearOpenAlert resolves any open backend_unreachable item for the site once
// the box is healthy again (self-clearing alert).
func clearOpenAlert(ctx context.Context, db *sql.DB, siteID string, log *zap.Logger) error {
	res, err := db.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete', completed_at = now(), updated_at = now(),
		    result = COALESCE(result, '{}'::jsonb) || '{"resolved_by":"backend-unreachable-check","reason":"health recovered"}'::jsonb
		WHERE site_id = $1 AND item_key = $2
		  AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved'])`,
		siteID, backendUnreachableItemKey)
	if err != nil {
		return fmt.Errorf("backend_unreachable: clear for %s: %w", siteID, err)
	}
	if n, _ := res.RowsAffected(); n > 0 {
		log.Info("backend_unreachable: cleared recovered alert", zap.String("site_id", siteID), zap.Int64("items", n))
	}
	return nil
}
