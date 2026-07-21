// FILE: platform/orchestration/actions/discovery_checks/check_backend_entry_orphaned.go
//
// Discovery check: backend_entry_orphaned  (Finding A — method_mismatch_link)
//
// Detects a browser-clickable <a href="/route"> that lands on a backend handler
// which does NOT accept GET — i.e. a plain GET link to a POST-only endpoint. The
// visitor clicks, the handler answers "405 Method Not Allowed" ("POST only"), and
// the funnel is a dead end even though every static check passes: the href DOES
// resolve (nginx proxies it), the page DOES exist, nothing errors on build.
//
// Filed as bugs_open/017 (static-cutover orphans backend entry forms). It shipped
// on idea.uk the moment the VM cutover gave "/" to the static site: the tool had
// served its own landing page carrying the audience-check and report-request
// forms (POST /audience-check, POST /request). The cutover kept the forms' TARGETS
// (nginx proxies all 16 tool routes) but lost the forms themselves, leaving two
// `href="/audience-check"` links — GET requests to a POST-only handler. The paid
// funnel was unreachable on a live earning site and NO existing check caught it,
// because none models the backend: they ask "does this href resolve?" and
// /audience-check resolves (200-family from nginx to a browser HEAD/probe... no —
// to a GET it is 405). This check asks the question they don't: does the link's
// TARGET accept the METHOD the link will use?
//
// Why a live probe, and why un-gated by deploy_config.target:
//   - The chassis cannot introspect a foreign binary's route table, so the only
//     reliable source of "which method does /route accept" is to ask the route.
//     Modeled on check_backend_unreachable.go, which already probes the live site.
//   - It deliberately does NOT gate on deploy_config.target='vm'. idea.uk — the
//     site this was written for — carries an EMPTY deploy_config {} despite being
//     VM-hosted with a backend (verified 2026-07-21), so a target='vm' gate would
//     NOOP on the very site the bug is about. The backend is not modelled in the
//     data; a probe sidesteps that entirely. A static-only site simply produces no
//     405s and no findings — it costs only the probes, which are bounded below.
//
// Deliberate boundaries (each one bounds cost or false positives):
//   - **405 only.** A GET that returns 405 means the route EXISTS and rejects the
//     method — precisely the "POST only" class. 404 is a broken link (owned by
//     phantom_internal_links); 5xx is a server fault (backend_unreachable); 200/3xx
//     are fine. Flagging only 405 keeps this a low-false-positive, high-severity
//     signal and off every other check's turf.
//   - **Extensionless internal page links only.** Funnel handler routes are clean
//     paths (/audience-check, /request, /subscribe); a browser <a href> to one
//     never carries a file extension. Skipping paths whose last segment has a "."
//     (.html pages, assets) turns ~30 probes/site into ~1-5 without losing the
//     class — a .html can't 405, and if it 404s that is phantom_internal_links'.
//     This is a COST bound, not the decision: 405 is still the only thing flagged.
//   - **page_components only, NOT site_components.** Matches dead_controls: chrome
//     nav has its own fixers, and a broken funnel link lives in content.
//   - **Deduped per destination path and capped.** A path linked from three pages
//     is one probe and one work item (the destination is what's orphaned). The
//     probe count is capped (maxProbePaths) and a hit on the cap is logged, never
//     silent (CLAUDE.md: no silent caps).
//   - **runtime-fill shells exempt** (data-runtime-fill): their hrefs hydrate
//     client-side, same carve-out as dead_controls.
//
// Routing: needs_human_review with NO handler. The fix is a decision the chassis
// cannot make safely — repoint the link at a GET page, or author the entry form
// that posts to the handler (bugs_open/017 site-fix options 1 and 2) — so picking
// one automatically would guess. Same reasoning as dead_controls / contact_form_
// undeliverable, the two closest siblings.
//
// Registration: automatic via init(). Enable by adding "backend_entry_orphaned"
// to a discovery agent's checks array AFTER the image carrying this file is live.

package discovery_checks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&BackendEntryOrphanedCheck{}) }

type BackendEntryOrphanedCheck struct{}

func (c *BackendEntryOrphanedCheck) Name() string { return "backend_entry_orphaned" }

// maxProbePaths bounds the live probes per site per run. Handler-like routes are
// few (idea.uk has ~16, of which a handful are ever linked); this is a runaway
// backstop, and hitting it is logged.
const maxProbePaths = 40

// probePerRequestTimeout is the per-probe deadline. A 405/200 returns fast; this
// is the cap for a hung backend so one bad route can't stall the whole run.
const probePerRequestTimeout = 6 * time.Second

type methodMismatchFinding struct {
	Path       string   `json:"path"`        // request path probed, e.g. "/audience-check"
	Status     int      `json:"status"`      // 405
	LinkedFrom []string `json:"linked_from"` // page names carrying an <a href> to Path
}

func (c *BackendEntryOrphanedCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, '') FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("backend_entry_orphaned: load site: %w", err)
	}
	if domain == "" {
		return result, nil // nothing to probe against
	}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, pc.rendered_html
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND p.status = 'active'
		  AND p.build_status = 'deployed'
		  AND pc.rendered_html IS NOT NULL
		  AND pc.rendered_html <> ''
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("backend_entry_orphaned: page_components query failed: %w", err)
	}
	defer rows.Close()

	// destination path (normalised) -> set of page names that link to it, plus
	// the raw request path to probe (first-seen wins; they normalise equal).
	type dest struct {
		probePath string
		pages     map[string]bool
	}
	dests := map[string]*dest{}
	for rows.Next() {
		var pageName, html string
		if err := rows.Scan(&pageName, &html); err != nil {
			continue
		}
		if strings.Contains(html, "data-runtime-fill") {
			continue // client-hydrated shell: placeholder hrefs are by design
		}
		for _, a := range datahelpers.ExtractAnchors(html) {
			probePath, ok := handlerRouteCandidate(a.Href)
			if !ok {
				continue
			}
			key := datahelpers.NormalizePagePath(a.Href)
			d := dests[key]
			if d == nil {
				d = &dest{probePath: probePath, pages: map[string]bool{}}
				dests[key] = d
			}
			d.pages[pageName] = true
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("backend_entry_orphaned: row scan failed: %w", err)
	}
	if len(dests) == 0 {
		return result, nil
	}

	// Deterministic probe order + cap.
	keys := make([]string, 0, len(dests))
	for k := range dests {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) > maxProbePaths {
		dctx.Logger.Warn("backend_entry_orphaned: probe cap hit — some paths not checked this run",
			zap.String("domain", domain),
			zap.Int("distinct_paths", len(keys)),
			zap.Int("cap", maxProbePaths))
		keys = keys[:maxProbePaths]
	}

	var findings []methodMismatchFinding
	var probeErrors int
	for _, k := range keys {
		d := dests[k]
		status, perr := probeGETStatus(dctx.Ctx, "https://"+domain+d.probePath)
		if perr != nil {
			// A probe that could not RUN is not evidence the route is clean — it
			// means this route went UNCHECKED. Count it and surface it below
			// (Warn), distinctly from a clean 200/404, so a check whose whole job
			// is catching a silent failure does not itself go silently blind on
			// exactly the fragile VM backends it exists to watch.
			probeErrors++
			dctx.Logger.Debug("backend_entry_orphaned: probe could not run — route UNCHECKED this run",
				zap.String("domain", domain), zap.String("path", d.probePath), zap.Error(perr))
			continue
		}
		if status != http.StatusMethodNotAllowed {
			continue
		}
		pages := make([]string, 0, len(d.pages))
		for p := range d.pages {
			pages = append(pages, p)
		}
		sort.Strings(pages)
		findings = append(findings, methodMismatchFinding{
			Path: d.probePath, Status: status, LinkedFrom: pages,
		})
	}
	// Distinguish "probed clean" from "could not probe". A run where routes went
	// unchecked is NOT a clean bill of health — say so, so the check's own blind
	// spots are visible in its output, not hidden behind a zero-findings result.
	if probeErrors > 0 {
		dctx.Logger.Warn("backend_entry_orphaned: some routes could not be probed — UNCHECKED this run, NOT proven clean",
			zap.String("domain", domain),
			zap.Int("unchecked", probeErrors),
			zap.Int("total_paths", len(keys)))
	}

	if len(findings) == 0 {
		return result, nil
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Path < findings[j].Path })

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":    "backend_entry_orphaned",
		"finding":  "method_mismatch_link",
		"count":    len(findings),
		"findings": findings,
	})

	for _, f := range findings {
		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":       "backend_entry_orphaned",
			"finding":     "method_mismatch_link",
			"path":        f.Path,
			"status":      f.Status,
			"linked_from": f.LinkedFrom,
			"probe":       fmt.Sprintf("GET https://%s%s -> %d", domain, f.Path, f.Status),
			"fix": "A GET link points at a POST-only backend handler — clicking it shows " +
				"'405 Method Not Allowed'. Decide: repoint the link at a GET page that " +
				"hosts the entry form, or author the form (a <form method=POST action> " +
				"same-origin) so the funnel has an entry. See bugs_open/017.",
		})

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			Source:   "discovery",
			Pipeline: "build",
			ItemType: "backend_entry_orphaned",
			Severity: "high",
			Summary: fmt.Sprintf("GET link to a POST-only handler on %s: %s returns 405 (linked from %s)",
				domain, f.Path, strings.Join(f.LinkedFrom, ", ")),
			SpecJSON:     string(specJSON),
			Priority:     30, // high band — an unreachable funnel entry
			HandlerAgent: "", // alert only; the fix is a business/design decision
			Status:       "needs_human_review",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("backend_entry_orphaned:%s", datahelpers.NormalizePagePath(f.Path)),
			BatchID:      dctx.BatchID,
		})
	}

	dctx.Logger.Warn("backend_entry_orphaned: GET links to POST-only handlers on deployed pages",
		zap.String("domain", domain),
		zap.Int("count", len(findings)),
		zap.String("site_id", dctx.SiteID.String()))

	return result, nil
}

// handlerRouteCandidate reports whether an href is a browser GET link to a
// handler-like backend route worth probing, and returns the request path to
// probe (fragment and query stripped). It keeps only internal navigable page
// links (LinkScopePage — excludes empty, #anchors, external, mailto, assets)
// whose last path segment has no file extension. See the boundaries note: this
// is a COST filter; the 405 status is the actual decision.
func handlerRouteCandidate(href string) (path string, ok bool) {
	if datahelpers.ClassifyLinkScope(href) != datahelpers.LinkScopePage {
		return "", false
	}
	// Request path = href minus #fragment and ?query. Do NOT lowercase: server
	// paths can be case-sensitive and we must probe exactly what a click sends.
	p := href
	if i := strings.IndexAny(p, "#?"); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return "", false // the homepage is not a handler route
	}
	last := p
	if i := strings.LastIndex(p, "/"); i >= 0 {
		last = p[i+1:]
	}
	if strings.Contains(last, ".") {
		return "", false // .html page or asset — cannot be a POST-only handler
	}
	return p, true
}

// probeGETStatus issues a GET to the given URL and returns the HTTP status.
// GET (not HEAD) reproduces exactly what a click on the <a href> sends, since a
// handler may treat the two methods differently. The body is discarded. Takes a
// full URL (rather than domain+path) so the 405-discrimination can be exercised
// against an httptest server in the unit test.
func probeGETStatus(ctx context.Context, url string) (int, error) {
	cctx, cancel := context.WithTimeout(ctx, probePerRequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512))
	return resp.StatusCode, nil
}
