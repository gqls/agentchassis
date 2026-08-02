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
		// Self-clear, through the shared retraction seam rather than an inline
		// UPDATE (RFC_010, owner ruling 2026-08-02 "Decision 1: option 1").
		//
		// This check was the ONE of fifty that could retract a finding, and it
		// did so with its own query and its own copy of the status vocabulary.
		// That copy is what generalised: CheckResult.Resolved now carries the
		// claim and the runner owns the SQL, so the vocabulary cannot drift
		// fifty ways as other checks adopt it.
		//
		// TWO BEHAVIOUR CHANGES, both deliberate and both improvements:
		//   * the inline query excluded 'unresolved' and 'failed', so an alert
		//     that had been given up on stayed open for ever even once the host
		//     recovered. The shared predicate reaches them — the owner's
		//     Decision 2 ruling that `unresolved` is OPEN, delivered here
		//     without touching the dedup index it also implies.
		//   * the retraction is now skipped if this check returns an error,
		//     which the inline version could not guarantee for itself.
		//
		// AllOfType is correct here and is stated rather than implied: a
		// successful health probe answers the whole item type for this site at
		// once, which is exactly the breadth the flag exists to make visible at
		// the call site.
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType:  "backend_unreachable",
			AllOfType: true,
			Reason:    "health recovered: " + domain + " responded to the backend probe",
		})
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
