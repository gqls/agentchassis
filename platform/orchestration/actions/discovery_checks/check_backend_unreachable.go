// FILE: platform/orchestration/actions/discovery_checks/check_backend_unreachable.go
//
// Discovery check: backend_unreachable
//
// Per-site health probe for VM-hosted backend sites (deploy_config.target='vm').
// NOOPs for static (B2/Worker) sites. Probes the public https://<domain>/health
// (which exercises BOTH nginx and the site-engine behind it). Self-clearing:
// when the box recovers, it resolves its own open item via the runner's tx.
//
// This is an ALERT, not an auto-fix: a down VM isn't a chassis-fixable content
// problem, so the work item carries NO handler_agent and simply stays visible
// at 'detected' until the box is back (or, later, until the P5 vmhost adapter
// becomes its handler_agent and can remediate). Modeled on check_tool_health.go.
//
// Enablement: add "backend_unreachable" to the discovery agent's
// default_config checks array (the agent the improvement sweep targets).

package discovery_checks

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

func init() { Register(&BackendUnreachableCheck{}) }

type BackendUnreachableCheck struct{}

func (c *BackendUnreachableCheck) Name() string { return "backend_unreachable" }

func (c *BackendUnreachableCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// Only VM-hosted backend sites have a backend to probe.
	var domain, target string
	err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT domain, COALESCE(deploy_config->>'target', '')
		   FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain, &target)
	if err != nil {
		return nil, fmt.Errorf("backend_unreachable: load site: %w", err)
	}
	if target != "vm" || domain == "" {
		return result, nil // not a backend site — nothing to check
	}

	reachable, detail := probeBackendHealth(dctx.Ctx, domain)

	if reachable {
		// Self-clear: resolve any open backend_unreachable item for this site.
		res, uerr := dctx.TX.ExecContext(dctx.Ctx, `
			UPDATE site_work_items
			SET status = 'complete', completed_at = now(), updated_at = now(),
			    result = COALESCE(result, '{}'::jsonb)
			             || '{"resolved_by":"backend_unreachable","reason":"health recovered"}'::jsonb
			WHERE site_id = $1
			  AND item_type = 'backend_unreachable'
			  AND status <> ALL (ARRAY['complete','verified','rejected','wont_fix','failed','unresolved'])`,
			dctx.SiteID)
		if uerr != nil {
			return nil, fmt.Errorf("backend_unreachable: self-clear: %w", uerr)
		}
		if n, _ := res.RowsAffected(); n > 0 {
			dctx.Logger.Info("backend_unreachable: cleared recovered alert",
				zap.String("domain", domain), zap.Int64("items", n))
		}
		return result, nil
	}

	// Unreachable → emit one alert. insertWorkItem (in the runner) dedups on
	// (site_id, item_key); idx_swi_dedup keeps a single open item per site.
	dctx.Logger.Warn("backend_unreachable: probe failed",
		zap.String("domain", domain), zap.String("detail", detail))

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":  "backend_unreachable",
		"domain": domain,
		"detail": detail,
	})
	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: "build",
		ItemType: "backend_unreachable",
		Severity: "high",
		Summary:  fmt.Sprintf("Backend unreachable for %s (%s) — engine/nginx health probe failed", domain, detail),
		SpecJSON: fmt.Sprintf(
			`{"check": "backend_unreachable", "domain": "%s", "probe": "https://%s/health", "detail": "%s"}`,
			escapeJSON(domain), escapeJSON(domain), escapeJSON(detail),
		),
		Priority:     30, // high — same band as tool_health blockers
		HandlerAgent: "", // alert only; no auto-fixer (P5 vmhost adapter later)
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("backend_unreachable:%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})

	return result, nil
}

// probeBackendHealth GETs https://<domain>/health and checks for {"ok":true}.
func probeBackendHealth(ctx context.Context, domain string) (ok bool, detail string) {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(cctx, http.MethodGet, "https://"+domain+"/health", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, "request error: " + err.Error()
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256))
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("status %d", resp.StatusCode)
	}
	if !strings.Contains(string(body), `"ok":true`) {
		return false, "unexpected /health body"
	}
	return true, ""
}
