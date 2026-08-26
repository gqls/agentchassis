// FILE: platform/orchestration/actions/discovery_checks/check_archived_page_still_serving.go
//
// Discovery check: archived_page_still_serving — "we retired this page; is it
// still answering the public?"
//
// bugs_open/359. `pages.status = 'archived'` means the platform RETIRED a page.
// It does NOT mean the page stopped being served: retirement sets `status` and
// leaves the deployed artefact exactly where it is, so removing the file is a
// separate act (`retract_page_deployment_action.go`, bugs_closed/098). Nothing
// anywhere asked whether the two agree — no discovery check, no cron, no
// scheduled task — so the gap between *retired* and *retracted* was unobserved
// by construction, and a page could sit in it indefinitely. One did:
// robot-hands.com/gripper-catalog.html was first recorded archived-and-serving on
// 2026-08-14 and was still serving 30,997 bytes on 2026-08-26. Twelve days, no
// work item, no agent_error_log row, no doc note.
//
// [MEASURED 2026-08-26, scripts/audit-archived-still-serving.sh] 39 archived-and-
// shipped pages fleet-wide; **7 serving 200**; 32 correctly absent. And the
// population MOVES — both loancalculator.co.uk pages the bug filed on 2026-08-22
// now 404, and five of the seven were not in that sample. This is a flow, not a
// backlog, which is why it wants a meter and not a sweep.
//
// ── THE PROPERTY THAT SHAPES EVERYTHING: THIS CHECK'S FINDING IS A 200 ──────
//
// Every other live-probe check in this package files on a NEGATIVE observation —
// asset_reference_404 files a 404, dead_internal_link_live files a 404,
// site_unreachable files an outage. When one of those is blinded (origin down,
// DNS gone, everything timing out) it UNDER-REPORTS: it files less than the truth.
//
// This check is inverted, and so is its blindness. If the origin is down, every
// archived page answers 404, every page reads "correctly absent", and the check
// reports **ZERO** — which is precisely what a healthy estate reports. A silent
// decline is therefore NOT sufficient here, and neither is `stylesheet_gutted`'s
// otherwise-excellent "decline to judge and say nothing" posture: in the item
// table a blinded run and a clean estate are the same reading (016b §9, and
// bugs_open/359 §7 states it as an acceptance criterion in its own right).
//
// So both instrument controls are run INSIDE the check, before any page is
// judged, and **a failed control makes Run return an ERROR**. That is not a
// gesture — it buys three properties from machinery that already exists:
//
//  1. The runner writes a durable, dated `DISCOVERY_CHECK_ERROR` row to
//     agent_error_log (discovery_checks.go, severity `warning`, message naming
//     that the site was NOT checked for this class). A pod log line scrolls; a
//     row does not.
//  2. Retraction becomes STRUCTURALLY impossible on a blinded run. The runner's
//     `err != nil` branch `continue`s before the Resolved loop, and
//     registry.go's CheckResult.Resolved contract states it in terms: "The
//     runner additionally skips Resolved entirely when Run returned an error."
//     So a dead origin can never mass-close real findings.
//  3. The step output names the check in `checks_failed`.
//
// Each check errors independently inside the runner's per-check loop, so
// refusing here does not take `site_unreachable` or `page_content_divergence`
// down with it.
//
// ── THE TWO CONTROLS, AND WHY NEITHER IS OPTIONAL ──────────────────────────
//
//	INVENTED URL must be non-2xx.  Guards against FALSE POSITIVES. A parked or
//	  catch-all domain answers 200 on every path, so without this every archived
//	  page on it reads as damage. This is a recorded trap that has already
//	  reversed an architectural conclusion once ("a parked domain 200s EVERY
//	  path").
//	ACTIVE SIBLING must be 2xx.    Guards against the FALSE ALL-CLEAR above.
//
// They fail in opposite directions, which is why one cannot stand in for the
// other. Both are probed through the same seam as the pages, so a test can
// script all three URL classes through one stub and assert the ordering.
//
// ── WHAT IS A FINDING, AND WHAT IS MERELY NOT ONE ──────────────────────────
//
// ONLY a 2xx at the page's OWN final URL, confirmed by a second probe, is a
// finding. Everything else is a skip or a resolve, counted and logged, never
// filed — the discipline check_asset_reference_404 states for itself, applied to
// the mirrored question. In particular:
//
//   - 401/403/429/5xx and anything else: the check is BLINDED for that page, not
//     informed. Never a finding, never a resolve.
//   - a transport failure is NOT a status and is never compared against one.
//   - a 2xx at a DIFFERENT final URL is a redirected retirement, which is a
//     legitimate way to retire a page. It files nothing and RESOLVES: the retired
//     URL no longer serves the retired content, and that is a positive
//     observation. (The `redirects` table holds 0 rows fleet-wide as of
//     2026-08-26, so this is headroom rather than a live path — but the wire can
//     do it whether or not we model it.)
//
// Every request is CACHE-BUSTED through the package's existing `cacheBust`. An
// edge cache answering 200 for a file that has already been purged is exactly the
// wrong reading here, and the confirming probe gets a fresh buster so the two
// reads are independent.
//
// ── FLAG-ONLY, AND WHY AUTO-RETRACTION IS NOT IN THIS SHIP ─────────────────
//
// HandlerAgent is empty. A wrongly-archived page that is serving correctly is
// INDISTINGUISHABLE ON THE WIRE from a rightly-archived page that is serving
// wrongly — the difference is intent, which lives in a human. bugs_open/359 §6.4
// is explicit that un-publishing a good live page is the failure this estate
// calls "worse than the bug", and the owner ruling of 2026-08-02 §2 says new
// authority on a shared seam ships opt-in with the unsafe default OFF. So the
// item carries a `triage_hint` naming the remedy and the evidence query, and a
// person decides between retracting and un-archiving.
//
// The cost of flag-only is real and named in bugs_open/083: "a detector whose
// output nobody drains is not neutral — it is actively misleading". The item
// lands at `detected` where image_url_404, asset_reference_404 and
// stylesheet_gutted land, and is deliberately not dispatchable —
// detected-item-promoter's `handler_ok` door requires an agent_definitions row
// matching handler_agent, which `''` cannot satisfy.
//
// ── THE POPULATION IS THE CONJUNCTION OF TWO INDEPENDENT AXES ──────────────
//
//	LIFECYCLE  p.status = 'archived'                    the platform retired it
//	BUILD      datahelpers.PageHasShippedPredicateFor   an artefact exists to serve
//
// Spelled from the two existing helpers and NOT from a merged one:
// datahelpers/links.go states the contract — "Pair this with whichever
// build-axis arm YOUR question needs; do not expect one combined helper … A
// merged deployed-and-active helper would misdescribe two of the three." And not
// `deployed_at IS NOT NULL` either: the shipped helper also keeps the shape that
// carries `build_status='deployed'` with a never-stamped `deployed_at`.
//
// This check is declared **PostureObserves** in page_lifecycle_posture_test.go.
// Arming the lifecycle axis here would blind it to its own subject, which is the
// ⚠ case that registry's own header names.
//
// ── KNOWN, STATED GAPS — a reader should not have to discover these ─────────
//
//   - A page whose ROW is DELETED rather than archived is outside the population
//     entirely: there is nothing left to join to, and its artefact may still
//     serve. Same class as check_asset_reference_404's deleted-reference gap.
//   - A WAF that 404s the invented control but answers the real page with a 200
//     challenge body would file wrongly. [UNMEASURED] — not observed in the
//     2026-08-26 census, where all 7 serving pages were real pages corroborated
//     by byte size. The confirming probe and flag-only routing bound the damage
//     to one visible item a human reads.
//   - A site with archived-and-shipped pages and NO active page to serve as the
//     sibling control cannot be judged at all, and refuses loudly rather than
//     reporting clean. [UNMEASURED] whether any such site exists today.

package discovery_checks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&ArchivedPageStillServingCheck{}) }

type ArchivedPageStillServingCheck struct{}

func (c *ArchivedPageStillServingCheck) Name() string { return "archived_page_still_serving" }

const (
	// archivedItemType is the work item type. One producer — this file — by
	// design: adopting the retraction seam on a co-dedup'd type closes the OTHER
	// producer's finding, which is a recorded landmine, and a first adoption
	// should not take that risk.
	//
	// ⚠ THE STRUCT FIELDS BELOW SPELL THE LITERAL, NOT THIS CONST, AND THAT IS
	// DELIBERATE. verifier_coverage_test.go's SENSOR reads the work
	// item's type field from SOURCE; a const there makes it a "computed site" that must be declared in
	// computedItemTypeSites — which would buy tidiness by putting a hole in the
	// guard exactly where this file is, and an allow-list that silences the
	// detector written to catch you is a recorded landmine. The const stays for
	// the SQL and the key prefix, and TestArchivedItemTypeConstMatchesTheLiteral
	// pins the two together so they cannot drift.
	archivedItemType = "archived_page_still_serving"

	// archivedProbeTimeout per request. Status-only, no body needed, so the
	// subresource figure rather than site_unreachable's 15s.
	archivedProbeTimeout = 10 * time.Second

	// maxArchivedProbeURLs bounds the outbound calls one site can cause. The
	// WHOLE FLEET held 39 archived-and-shipped pages on 2026-08-26, so this is
	// headroom — but it is a limit, and when it drops anything it LOGS what it
	// dropped. A silent cap reads as "everything was checked" when it was not.
	maxArchivedProbeURLs = 40
)

// archivedProbeRetryWait is the pause before a control's confirming second
// attempt. A var, not a const, only so tests need not sleep through it — the
// siteProbeRetryWait idiom.
var archivedProbeRetryWait = 5 * time.Second

// archivedProbeResult is one attempt's observation. FinalHost/FinalPath are
// where the request ENDED, after redirects — the difference between "this page
// serves" and "this page redirects somewhere that serves" is the whole
// distinction between a finding and a clean retirement.
//
// TransportErr non-empty means no HTTP conversation happened at all. It is not a
// status and must never be compared against one.
type archivedProbeResult struct {
	Status       int
	FinalHost    string
	FinalPath    string
	TransportErr string
}

// probeArchivedPageURL GETs one absolute URL the way a visitor would, following
// redirects. Swappable in tests — the same seam probeAssetURL, probeSiteOrigin
// and fetchStructuralPage use, and the reason the test file can prove every
// branch without a network.
var probeArchivedPageURL = func(ctx context.Context, target string) archivedProbeResult {
	cctx, cancel := context.WithTimeout(ctx, archivedProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cctx, http.MethodGet, target, nil)
	if err != nil {
		return archivedProbeResult{TransportErr: "build request: " + err.Error()}
	}
	// GET rather than HEAD: a static origin need not implement HEAD, and its 405
	// would be indistinguishable from a policy refusal. An explicit, honest
	// User-Agent — if an origin refuses it, that is a 403 and files nothing.
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+archived_page_still_serving)")
	req.Header.Set("Accept", "text/html,*/*")

	client := &http.Client{Timeout: archivedProbeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return archivedProbeResult{TransportErr: err.Error()}
	}
	defer resp.Body.Close()
	// The status and the final URL are the whole answer; drain a token amount so
	// the connection can be reused and discard the rest.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	out := archivedProbeResult{Status: resp.StatusCode}
	if resp.Request != nil && resp.Request.URL != nil {
		out.FinalHost = resp.Request.URL.Host
		out.FinalPath = resp.Request.URL.Path
	}
	return out
}

// archivedVerdictKind is the judged outcome for ONE archived page.
type archivedVerdictKind string

const (
	// archivedServing is the defect: the public gets this retired page.
	archivedServing archivedVerdictKind = "still_serving"
	// archivedAbsent is the healthy outcome and a positive observation.
	archivedAbsent archivedVerdictKind = "absent"
	// archivedRedirected is a legitimate retirement: the URL answers, but with
	// something else. Files nothing; resolves.
	archivedRedirected archivedVerdictKind = "redirected"
	// archivedInconclusive means the check was BLINDED for this page — it learned
	// nothing, which is not the same as learning the page is fine.
	archivedInconclusive archivedVerdictKind = "inconclusive"
)

type archivedVerdict struct {
	Kind   archivedVerdictKind
	Reason string
	Detail string
}

// judgeArchivedProbe is pure, so the verdict table is testable row by row without
// a network or a database — judgeSiteProbe's shape, for judgeSiteProbe's reason.
//
// `pageURL` is the url the platform RECORDS for this page. The comparison is
// against the FINAL location after redirects, normalised through the shared
// NormalizePagePath so that /foo/index.html, /foo/ and /foo are one path and the
// cache-buster query is stripped.
func judgeArchivedProbe(domain, pageURL string, r archivedProbeResult) archivedVerdict {
	if r.TransportErr != "" {
		return archivedVerdict{
			Kind:   archivedInconclusive,
			Reason: "transport_error",
			Detail: r.TransportErr,
		}
	}
	if r.Status < 200 || r.Status >= 300 {
		// 3xx never reaches here: the client follows redirects, so a redirect
		// presents as a 2xx at a different final URL. Everything else — 404 and
		// 410 included — is judged below.
		if r.Status == http.StatusNotFound || r.Status == http.StatusGone {
			return archivedVerdict{
				Kind:   archivedAbsent,
				Reason: "gone",
				Detail: fmt.Sprintf("HTTP %d", r.Status),
			}
		}
		return archivedVerdict{
			Kind:   archivedInconclusive,
			Reason: "inconclusive_status",
			Detail: fmt.Sprintf("HTTP %d", r.Status),
		}
	}

	// A 2xx. The question is now WHOSE 2xx it is.
	if !sameSiteHost(domain, r.FinalHost) {
		return archivedVerdict{
			Kind:   archivedRedirected,
			Reason: "redirected_off_site",
			Detail: fmt.Sprintf("HTTP %d at %s%s", r.Status, r.FinalHost, r.FinalPath),
		}
	}
	if datahelpers.NormalizePagePath(r.FinalPath) != datahelpers.NormalizePagePath(pageURL) {
		return archivedVerdict{
			Kind:   archivedRedirected,
			Reason: "redirected",
			Detail: fmt.Sprintf("HTTP %d at %s — the retired URL no longer serves the retired content", r.Status, r.FinalPath),
		}
	}
	return archivedVerdict{
		Kind:   archivedServing,
		Reason: "still_serving",
		Detail: fmt.Sprintf("HTTP %d at its own URL", r.Status),
	}
}

// archivedPage is one row of the population.
type archivedPage struct {
	ID   uuid.UUID
	Name string
	URL  string
}

// archivedSkipTally records what was NOT judged, and why. A check that reports
// only findings cannot be told apart from one that was blinded.
type archivedSkipTally struct {
	underivableURL    int // fragment, query, off-origin — no file of its own
	collisionSkipped  int // an active page owns the same derived file path
	inconclusive      int // 401/403/429/5xx and anything else
	transportError    int
	overCap           int
	unconfirmedServe  int // read 200 once, not twice
	unconfirmedAbsent int // read 404 once, not twice
}

func archivedItemKey(pageID uuid.UUID) string {
	return structuralItemKey(archivedItemType, pageID)
}

func (c *ArchivedPageStillServingCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}
	var skips archivedSkipTally

	// ── 1. Site gate. A pool site's domain is unrouted BY DESIGN, so probing one
	// fabricates a reading. Same gate, same reason, as check_site_unreachable.
	var domain, siteStatus string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, ''), COALESCE(status, '') FROM sites WHERE id = $1`,
		dctx.SiteID).Scan(&domain, &siteStatus); err != nil {
		return nil, fmt.Errorf("archived_page_still_serving: load site: %w", err)
	}
	if domain == "" || (siteStatus != "active" && siteStatus != "deployed") {
		return result, nil
	}

	// ── 2. ONE query, both arms. Split in Go on the scanned status so the two
	// axes stay visibly independent rather than being fused into a predicate
	// nobody can read back.
	active, archived, err := loadArchivedCheckPages(dctx)
	if err != nil {
		return nil, err
	}

	// ── 3. RESOLVE-ON-ACTIVE, BEFORE ANY EARLY RETURN.
	//
	// Clearance 3: the page is no longer retired, so this finding's premise is
	// gone. That is a POSITIVE read of the row — a SELECT returning
	// status='active' is an observation, not an inference from an empty findings
	// slice, which is what CheckResult.Resolved's contract forbids.
	//
	// ⚠ THIS PASS MUST STAY ABOVE STEP 4. The recorded landmine: "a monotonic
	// check's `if len(findings) == 0 { return }` early return makes its new
	// retraction INERT on exactly the sites that need it" — the zero-findings
	// site is the ONLY site the early return fires on, and it is precisely the
	// site whose stale items need closing. TestUnarchivedPageStillResolvesWithNoProbe
	// is the mutation proof and was written before this code.
	//
	// Emitted only for keys that actually have an OPEN item (one cheap SELECT),
	// so the common case costs no UPDATEs at all rather than one per active page.
	openKeys, err := loadOpenArchivedItemKeys(dctx)
	if err != nil {
		return nil, err
	}
	for _, p := range active {
		key := archivedItemKey(p.ID)
		if !openKeys[key] {
			continue
		}
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType: "archived_page_still_serving", // literal, not the const — see archivedItemType
			ItemKey:  key,
			Reason: fmt.Sprintf("pages.status for %q reads 'active' — the page is no longer retired, "+
				"so the finding's premise is gone", p.Name),
		})
	}

	// ── 4. Nothing retired on this site: no outbound calls at all. Legal ONLY
	// because step 3 has already run.
	if len(archived) == 0 {
		return result, nil
	}

	// ── 5. Both instrument controls, or refuse. A blinded run must change
	// NOTHING — it must not file, and it must not retract.
	if err := runArchivedInstrumentControls(dctx, domain, active); err != nil {
		return nil, err
	}

	// ── 6. Guards, probe, judge.
	activePaths, err := datahelpers.ActivePageFilePaths(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("archived_page_still_serving: %w", err)
	}

	sort.Slice(archived, func(i, j int) bool { return archived[i].URL < archived[j].URL })
	if len(archived) > maxArchivedProbeURLs {
		dropped := archived[maxArchivedProbeURLs:]
		skips.overCap = len(dropped)
		urls := make([]string, 0, len(dropped))
		for _, p := range dropped {
			urls = append(urls, p.URL)
		}
		dctx.Logger.Warn("archived_page_still_serving: probe cap reached — these retired pages were NOT checked",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", maxArchivedProbeURLs),
			zap.Int("dropped", len(dropped)),
			zap.Strings("dropped_urls", urls))
		archived = archived[:maxArchivedProbeURLs]
	}

	for _, p := range archived {
		// A url that designates no file of its own (a fragment, a query, an
		// off-origin url) would have us probe ANOTHER page's artefact and convict
		// this row for it. The live shape is idea.uk's "/tools.html#audience-check",
		// where "/tools.html" is a different page's canonical url. Declined by the
		// same shared function the publish and retraction sides use — never
		// sanitised.
		filePath, ok := datahelpers.PageFilePathFromURL(p.URL)
		if !ok {
			skips.underivableURL++
			dctx.Logger.Info("archived_page_still_serving: url designates no file of its own, not probed",
				zap.String("page", p.Name), zap.String("url", p.URL))
			continue
		}

		// An ACTIVE page derives the same file — the artefact there is the live
		// page's, so a 200 is the live page answering. Flagging it would raise an
		// item whose only remedy (retract_page_deployment) refuses this very page
		// on its own guard 3. The skip-set and the refuse-set are ONE function.
		if owner, clash := activePaths[filePath]; clash {
			skips.collisionSkipped++
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":     archivedItemType,
				"kind":      "path_owned_by_active_page",
				"page":      p.Name,
				"page_url":  p.URL,
				"file_path": filePath,
				"owner":     owner,
			})
			continue
		}

		target := preferredStructuralURL(domain, p.URL)
		first := probeArchivedPageURL(dctx.Ctx, cacheBust(target))
		v := judgeArchivedProbe(domain, p.URL, first)

		switch v.Kind {
		case archivedServing:
			// Confirm before filing. Here the FINDING is the 200, so the 200 is
			// what gets confirmed — the mirror of asset_reference_404, where the
			// finding is a 404. A fresh cache-buster, so the two reads are
			// independent.
			second := judgeArchivedProbe(domain, p.URL, probeArchivedPageURL(dctx.Ctx, cacheBust(target)))
			if second.Kind != archivedServing {
				skips.unconfirmedServe++
				dctx.Logger.Info("archived_page_still_serving: candidate 200 not reproduced, discarding",
					zap.String("url", target), zap.String("second", string(second.Kind)))
				continue
			}
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":       archivedItemType,
				"kind":        "still_serving",
				"page":        p.Name,
				"page_url":    p.URL,
				"probe_url":   target,
				"http_status": first.Status,
			})
			result.WorkItems = append(result.WorkItems, buildArchivedServingItem(dctx, domain, p, filePath, first))

		case archivedAbsent:
			// Confirm before RESOLVING too. Closing an item is an authority act,
			// and at this population size the second request is free.
			second := judgeArchivedProbe(domain, p.URL, probeArchivedPageURL(dctx.Ctx, cacheBust(target)))
			if second.Kind != archivedAbsent {
				skips.unconfirmedAbsent++
				dctx.Logger.Info("archived_page_still_serving: candidate absence not reproduced, not resolving",
					zap.String("url", target), zap.String("second", string(second.Kind)))
				continue
			}
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "archived_page_still_serving", // literal, not the const — see archivedItemType
				ItemKey:  archivedItemKey(p.ID),
				Reason: fmt.Sprintf("probed %s twice: %s — the retirement is real on the wire",
					target, v.Detail),
			})

		case archivedRedirected:
			// A legitimate retirement. Visible as a finding with its reason named,
			// so a later policy tightening is a one-line change rather than an
			// archaeology exercise — the title_absent/delegated idiom.
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":    archivedItemType,
				"kind":     v.Reason,
				"page":     p.Name,
				"page_url": p.URL,
				"detail":   v.Detail,
			})
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "archived_page_still_serving", // literal, not the const — see archivedItemType
				ItemKey:  archivedItemKey(p.ID),
				Reason:   fmt.Sprintf("probed %s: %s", target, v.Detail),
			})

		default:
			// Blinded for this page, not informed. Never a finding, and — the
			// half that matters — never a resolve either.
			if v.Reason == "transport_error" {
				skips.transportError++
			} else {
				skips.inconclusive++
			}
			dctx.Logger.Info("archived_page_still_serving: not judged",
				zap.String("url", target),
				zap.String("reason", v.Reason),
				zap.String("detail", v.Detail))
		}
	}

	logArchivedSkips(dctx, skips, len(archived))
	return result, nil
}

// loadArchivedCheckPages reads BOTH arms in one query. The conjunction is spelled
// from the two existing helpers, never from a merged one: datahelpers/links.go
// states that contract, and the two axes really are independent — archiving sets
// `status` and leaves the build columns untouched.
func loadArchivedCheckPages(dctx DiscoveryCheckContext) (active, archived []archivedPage, err error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id, COALESCE(p.name,''), COALESCE(p.url,''), COALESCE(p.status,'')
		  FROM pages p
		 WHERE p.site_id = $1
		   AND COALESCE(p.url,'') <> ''
		   AND (p.status = 'active'
		        OR (p.status = 'archived' AND `+datahelpers.PageHasShippedPredicateFor("p")+`))`,
		dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("archived_page_still_serving: load pages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var p archivedPage
		var status string
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &status); err != nil {
			return nil, nil, fmt.Errorf("archived_page_still_serving: scan page: %w", err)
		}
		if status == "archived" {
			archived = append(archived, p)
		} else {
			active = append(active, p)
		}
	}
	return active, archived, rows.Err()
}

// loadOpenArchivedItemKeys is what keeps the resolve-on-active pass cheap: only
// keys with an item still open are worth an UPDATE. It is a read, so it cannot
// be the thing that makes the pass inert — the pass runs whatever this returns.
func loadOpenArchivedItemKeys(dctx DiscoveryCheckContext) (map[string]bool, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT item_key FROM site_work_items
		 WHERE site_id = $1 AND item_type = $2
		   AND status NOT IN ('complete','verified','cancelled','rejected','wont_fix')`,
		dctx.SiteID, archivedItemType)
	if err != nil {
		return nil, fmt.Errorf("archived_page_still_serving: load open items: %w", err)
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var k string
		if err := rows.Scan(&k); err != nil {
			return nil, fmt.Errorf("archived_page_still_serving: scan open item: %w", err)
		}
		out[k] = true
	}
	return out, rows.Err()
}

// runArchivedInstrumentControls proves the instrument BEFORE any page is judged,
// in both directions. Returns an error — never a silent decline — when either
// control fails, because for this check a silent decline is indistinguishable
// from a clean estate.
func runArchivedInstrumentControls(dctx DiscoveryCheckContext, domain string, active []archivedPage) error {
	// CONTROL A — an invented URL must NOT answer 2xx. A catch-all domain makes
	// every archived page read as damage.
	invented := "https://" + domain + "/never-published-" + uuid.NewString() + "-359-control.html"
	inv := probeArchivedPageURL(dctx.Ctx, invented)
	if inv.TransportErr != "" {
		if err := archivedControlPause(dctx); err != nil {
			return err
		}
		inv = probeArchivedPageURL(dctx.Ctx, invented)
	}
	if inv.TransportErr != "" {
		return fmt.Errorf("archived_page_still_serving: BLINDED — the invented-URL control %s could not be "+
			"fetched (%s) twice, so nothing on this site can be judged and nothing may be retracted",
			invented, inv.TransportErr)
	}
	if inv.Status >= 200 && inv.Status < 300 {
		return fmt.Errorf("archived_page_still_serving: BLINDED — the invented-URL control %s returned HTTP %d; "+
			"a permissive router answers every path, so every archived page here would read as serving. "+
			"Refusing to judge, and refusing to retract", invented, inv.Status)
	}

	// CONTROL B — a known-good ACTIVE page must answer 2xx. Without it, a dead
	// origin makes every archived page read "correctly absent" and this check
	// reports zero, which is what a healthy estate reports.
	sibling := pickArchivedSiblingControl(active)
	if sibling == nil {
		return fmt.Errorf("archived_page_still_serving: BLINDED — this site has retired pages but no active page "+
			"to serve as the known-good control, so a 404 on a retired page cannot be told from a dead origin "+
			"(site %s)", domain)
	}
	sibURL := preferredStructuralURL(domain, sibling.URL)
	sib := probeArchivedPageURL(dctx.Ctx, cacheBust(sibURL))
	if !archivedControlServed(sib) {
		if err := archivedControlPause(dctx); err != nil {
			return err
		}
		sib = probeArchivedPageURL(dctx.Ctx, cacheBust(sibURL))
	}
	if !archivedControlServed(sib) {
		detail := fmt.Sprintf("HTTP %d", sib.Status)
		if sib.TransportErr != "" {
			detail = sib.TransportErr
		}
		return fmt.Errorf("archived_page_still_serving: BLINDED — the known-good control %s returned %s twice, "+
			"so every retired page on this site would read as correctly absent whether it is or not. "+
			"Refusing to judge, and refusing to retract", sibURL, detail)
	}
	return nil
}

// archivedControlServed accepts any 2xx after redirects, INCLUDING one that ends
// on another host. A deliberate off-domain delegation still proves the instrument
// can reach a serving origin, which is all this control claims — the same
// judgement check_site_unreachable makes about a delegated site.
func archivedControlServed(r archivedProbeResult) bool {
	return r.TransportErr == "" && r.Status >= 200 && r.Status < 300
}

// pickArchivedSiblingControl prefers the site's index — the page most likely to
// be genuinely live — and otherwise takes a deterministic first by name, so two
// runs over unchanged data probe the same control.
func pickArchivedSiblingControl(active []archivedPage) *archivedPage {
	var best *archivedPage
	for i := range active {
		p := active[i]
		if _, ok := datahelpers.PageFilePathFromURL(p.URL); !ok {
			continue
		}
		if p.Name == "index" {
			return &active[i]
		}
		if best == nil || p.Name < best.Name {
			best = &active[i]
		}
	}
	return best
}

func archivedControlPause(dctx DiscoveryCheckContext) error {
	select {
	case <-dctx.Ctx.Done():
		return dctx.Ctx.Err()
	case <-time.After(archivedProbeRetryWait):
		return nil
	}
}

func buildArchivedServingItem(dctx DiscoveryCheckContext, domain string, p archivedPage,
	filePath string, probe archivedProbeResult) WorkItemSpec {

	spec := map[string]interface{}{
		"check":             archivedItemType,
		"kind":              "still_serving",
		"page":              p.Name,
		"page_url":          p.URL,
		"probe_url":         preferredStructuralURL(domain, p.URL),
		"http_status":       probe.Status,
		"derived_file_path": filePath,
		"confirmed":         true,
		"triage_hint": "The remedy exists: retract_page_deployment (bugs_closed/098) removes the file, and " +
			"bugs_open/266's ARCHIVED_PAGE_GUARD now stops it being republished. Before choosing, read the " +
			"agent_error_log rows with error_code IN ('RETRACTION_AUDIT','RETRACTION_REFUSED') for this page — " +
			"a retraction may have been requested and failed. And decide the OTHER direction too: a page " +
			"archived by accident that is serving correctly should be un-archived, not deleted. Do NOT " +
			"auto-delete (bugs_open/359 §6.4).",
	}
	specJSON, _ := json.Marshal(spec)

	pageID := p.ID
	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   &pageID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "archived_page_still_serving", // literal, not the const — see archivedItemType
		// medium, not high: the page WORKS — nothing is broken for a visitor, and
		// the site is up (an outage is site_unreachable's finding). What is wrong
		// is governance: we are publishing something we decided to stop
		// publishing. Measured 2026-08-26, none of the seven live cases was
		// linked from its own site or listed in its sitemap, so this is a failure
		// to withdraw rather than an active invitation to index.
		Severity:  "medium",
		Summary:   fmt.Sprintf("Retired page still serving: %s returns HTTP %d at its own URL (confirmed twice; the site's invented-URL control does not answer)", preferredStructuralURL(domain, p.URL), probe.Status),
		SpecJSON:  string(specJSON),
		Priority:  50,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   archivedItemKey(p.ID),
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the header.
	}
}

func logArchivedSkips(dctx DiscoveryCheckContext, s archivedSkipTally, considered int) {
	if s.underivableURL == 0 && s.collisionSkipped == 0 && s.inconclusive == 0 &&
		s.transportError == 0 && s.overCap == 0 && s.unconfirmedServe == 0 && s.unconfirmedAbsent == 0 {
		return
	}
	dctx.Logger.Info("archived_page_still_serving: retired pages not judged",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("considered", considered),
		zap.Int("underivable_url", s.underivableURL),
		zap.Int("collision_skipped", s.collisionSkipped),
		zap.Int("inconclusive_status", s.inconclusive),
		zap.Int("transport_error", s.transportError),
		zap.Int("over_cap", s.overCap),
		zap.Int("unconfirmed_serving", s.unconfirmedServe),
		zap.Int("unconfirmed_absence", s.unconfirmedAbsent))
}
