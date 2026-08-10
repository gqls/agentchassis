// FILE: platform/orchestration/actions/discovery_checks/check_site_unreachable.go
//
// Discovery check: site_unreachable
//
// bugs_open/236 (522 half): a fully built site served HTTP 522 to every visitor
// indefinitely while every internal signal — sites.status, work items, pages,
// the sync workflow — read green, because nothing anywhere asks "if a stranger
// requests https://<domain>/, do they get the site?" This check asks exactly
// that, from the serving side, so it catches every cause at once: a missing
// worker route (the lendzy case), a dead origin, a dangling delegation, an
// expired certificate, an apex 404.
//
// Complement to check_backend_unreachable.go, which probes /health and NOOPs
// unless deploy_config.target='vm'. This check probes the public apex for EVERY
// active/deployed site regardless of serving mode — the static/worker-routed
// class, where the motivating outage happened, is covered by nothing else.
//
// ALERT, not auto-fix (same posture as backend_unreachable): no agent can
// repair Cloudflare routing today, so the item carries no handler_agent and
// stays visible at 'detected'. SELF-CLEARING via CheckResult.Resolved with
// AllOfType — the exact case that field's contract names as its example.
//
// WHAT DOES NOT FILE, deliberately (measured 2026-08-10 across all 21 deployed
// sites — see bugfix_236_site_availability/PLAN, decision D4):
//   - an off-domain redirect (webdesign.uk deliberately 302s to webdesign.co.uk);
//   - a 2xx HTML body that lacks the stored index title (mortgagecalculator
//     serves a divergent render today — a staleness defect, not an availability
//     one; and filing on it would have been 1/21 false-positive on day one).
//
// Both are recorded as findings with a named reason, so they are visible and a
// later policy tightening is a one-line change. Known limitation, stated: a
// registrar-parked domain answering 200 lands in the title_absent finding, not
// in a work item.
//
// Enablement: `site_unreachable` in the checks array of
// availability-discovery-agent (migration 368, held until this file's image
// rolls — the runner hard-fails on a name the binary does not register).

package discovery_checks

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&SiteUnreachableCheck{}) }

type SiteUnreachableCheck struct{}

func (c *SiteUnreachableCheck) Name() string { return "site_unreachable" }

const (
	// siteProbeTimeout per attempt. Generous next to the 10s subresource probes:
	// the motivating failure mode (proxy → dead origin) takes ~19s to answer 522
	// only because Cloudflare waits; a healthy site answers in well under 15s.
	siteProbeTimeout = 15 * time.Second

	// siteProbeBodyCap bounds what one probe will read. The title sits in the
	// first few KB; the cap only exists so a pathological origin cannot make
	// the check buffer without limit.
	siteProbeBodyCap = 256 * 1024
)

// siteProbeRetryWait before the confirming second attempt. A single failed
// probe files a high-severity item about a whole site, so a transient blip
// must not be enough — same confirm-before-filing shape as the 404 check's
// second probe. A var, not a const, only so tests need not sleep through it.
var siteProbeRetryWait = 5 * time.Second

// siteProbeResult is one attempt's observation, network-free for tests.
type siteProbeResult struct {
	TransportErr string // non-empty means no HTTP conversation happened
	Status       int    // final status, after redirects
	FinalHost    string // host that answered, after redirects
	Body         []byte // capped at siteProbeBodyCap
}

// probeSiteOrigin fetches https://<domain>/ the way a visitor would — GET,
// redirects followed. Swappable in tests (the probeAssetURL seam, same reason).
var probeSiteOrigin = func(ctx context.Context, domain string) siteProbeResult {
	cctx, cancel := context.WithTimeout(ctx, siteProbeTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, "https://"+domain+"/", nil)
	if err != nil {
		return siteProbeResult{TransportErr: "build request: " + err.Error()}
	}
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+site_unreachable)")
	req.Header.Set("Accept", "text/html,*/*")

	client := &http.Client{Timeout: siteProbeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return siteProbeResult{TransportErr: err.Error()}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, siteProbeBodyCap))
	return siteProbeResult{
		Status:    resp.StatusCode,
		FinalHost: resp.Request.URL.Host,
		Body:      body,
	}
}

// siteVerdict is the judged outcome of a probe.
type siteVerdict struct {
	Unreachable bool
	// Reason is machine-readable: transport_error | http_<status> | empty_body |
	// not_html | delegated | title_absent | healthy.
	Reason string
	Detail string
}

// judgeSiteProbe applies the verdict table from the PLAN. Pure, so every row is
// a table test.
func judgeSiteProbe(domain, storedTitle string, r siteProbeResult) siteVerdict {
	if r.TransportErr != "" {
		return siteVerdict{Unreachable: true, Reason: "transport_error", Detail: r.TransportErr}
	}
	if r.Status < 200 || r.Status > 299 {
		return siteVerdict{Unreachable: true,
			Reason: fmt.Sprintf("http_%d", r.Status),
			Detail: fmt.Sprintf("final status %d at https://%s/", r.Status, domain)}
	}
	body := strings.TrimSpace(string(r.Body))
	if body == "" {
		return siteVerdict{Unreachable: true, Reason: "empty_body",
			Detail: "2xx with an empty body at the apex"}
	}
	if !strings.Contains(strings.ToLower(body), "<html") {
		return siteVerdict{Unreachable: true, Reason: "not_html",
			Detail: "2xx but the apex body contains no <html"}
	}
	if !sameSiteHost(domain, r.FinalHost) {
		// A deliberate delegation (302 to a sibling domain) serves the visitor;
		// the delegate's own row gets its own probe. Reachable, noted.
		return siteVerdict{Reason: "delegated",
			Detail: "redirects off-domain to " + r.FinalHost}
	}
	if storedTitle != "" &&
		!strings.Contains(body, storedTitle) &&
		!strings.Contains(body, html.EscapeString(storedTitle)) {
		// Something HTML answers on our own host but it is not carrying the
		// shipped index title — stale render, rebrand, or a parked page.
		// Visible, unfiled: see the file header for the measured why.
		return siteVerdict{Reason: "title_absent",
			Detail: "2xx HTML on-host, but the shipped index title is not in the body"}
	}
	return siteVerdict{Reason: "healthy"}
}

// sameSiteHost treats www.<domain> and <domain> as one site; anything else that
// answers after redirects is a different host.
func sameSiteHost(domain, finalHost string) bool {
	if finalHost == "" {
		return true // no redirect happened; the origin that answered is ours
	}
	norm := func(h string) string {
		h = strings.ToLower(h)
		if i := strings.IndexByte(h, ':'); i >= 0 {
			h = h[:i]
		}
		return strings.TrimPrefix(h, "www.")
	}
	return norm(domain) == norm(finalHost)
}

func (c *SiteUnreachableCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	// Only sites that are supposed to serve are probed. A pool site's domain is
	// unrouted BY DESIGN (the 071 lane's fixtures rely on exactly that), so
	// probing one would file a fabricated outage.
	var domain, siteStatus string
	err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, ''), COALESCE(status, '') FROM sites WHERE id = $1`,
		dctx.SiteID).Scan(&domain, &siteStatus)
	if err != nil {
		return nil, fmt.Errorf("site_unreachable: load site: %w", err)
	}
	if domain == "" || (siteStatus != "active" && siteStatus != "deployed") {
		return result, nil
	}

	// The shipped index title is the body property that tells our site apart
	// from an arbitrary 200. PageHasShippedPredicateFor, not a hand-typed
	// build_status filter — pages.status cannot answer "is this live"
	// (bugs_open/185) and needs_rebuild pages are still serving.
	var storedTitle string
	titleQuery := `SELECT COALESCE(p.title, '') FROM pages p
		 WHERE p.site_id = $1 AND p.name = 'index' AND ` +
		datahelpers.PageHasShippedPredicateFor("p") +
		` ORDER BY p.updated_at DESC LIMIT 1`
	if err := dctx.DB.QueryRowContext(dctx.Ctx, titleQuery, dctx.SiteID).Scan(&storedTitle); err != nil {
		// No shipped index page is not an availability question — the probe
		// still runs, with the weaker html-only assertion.
		storedTitle = ""
	}
	storedTitle = strings.TrimSpace(storedTitle)

	verdict := judgeSiteProbe(domain, storedTitle, probeSiteOrigin(dctx.Ctx, domain))
	if verdict.Unreachable {
		// Confirm before filing: one blip must not raise a site-wide alarm.
		select {
		case <-dctx.Ctx.Done():
			return nil, dctx.Ctx.Err()
		case <-time.After(siteProbeRetryWait):
		}
		verdict = judgeSiteProbe(domain, storedTitle, probeSiteOrigin(dctx.Ctx, domain))
	}

	if !verdict.Unreachable {
		if verdict.Reason != "healthy" {
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":  "site_unreachable",
				"domain": domain,
				"reason": verdict.Reason,
				"detail": verdict.Detail,
			})
		}
		// The site serves. Every open unreachable item for it is answered at
		// once — the breadth AllOfType exists to state at the call site.
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType:  "site_unreachable",
			AllOfType: true,
			Reason:    fmt.Sprintf("probe recovered: https://%s/ serves (%s)", domain, verdict.Reason),
		})
		return result, nil
	}

	dctx.Logger.Warn("site_unreachable: apex probe failed twice",
		zap.String("domain", domain),
		zap.String("reason", verdict.Reason),
		zap.String("detail", verdict.Detail))

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":  "site_unreachable",
		"domain": domain,
		"reason": verdict.Reason,
		"detail": verdict.Detail,
	})
	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: "build",
		ItemType: "site_unreachable",
		Severity: "high",
		Summary: fmt.Sprintf("Site does not serve: https://%s/ — %s (%s)",
			domain, verdict.Reason, verdict.Detail),
		SpecJSON: fmt.Sprintf(
			`{"check": "site_unreachable", "domain": "%s", "probe": "https://%s/", "reason": "%s", "detail": "%s"}`,
			escapeJSON(domain), escapeJSON(domain), escapeJSON(verdict.Reason), escapeJSON(verdict.Detail),
		),
		Priority:     30, // high — same band as backend_unreachable / tool_health blockers
		HandlerAgent: "", // alert only; nothing can repair routing today
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      fmt.Sprintf("site_unreachable:%s", dctx.SiteID),
		BatchID:      dctx.BatchID,
	})
	return result, nil
}
