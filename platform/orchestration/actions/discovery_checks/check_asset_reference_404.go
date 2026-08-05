// FILE: platform/orchestration/actions/discovery_checks/check_asset_reference_404.go
//
// Discovery check: asset_reference_404 — "does the subresource this page asks the
// browser to load actually exist?"
//
// A deployed page can reference a `<script src>` or a `<link rel="stylesheet">`
// that 404s, and every other signal the platform has will say the page is fine.
// The HTML is well-formed, the render completed, `build_status` is 'deployed',
// the orchestration is COMPLETED. The page simply does nothing when a visitor
// clicks it. JavaScript is uniquely exposed to presence-only checking because
// its absence changes nothing visible until a human interacts.
//
// Filed as bugs_open/084 fix candidate 2, and its own words: "the single
// highest-value item here".
//
// ── WHY THIS IS NOT THE BROWSER TIER'S JOB, since TL-032 says it is ─────────
//
// The concept register (register/tool-lifecycle.md, TL-032) rules that the
// external-script class "belongs to the BROWSER tier". That is right about where
// the RICH answer lives and wrong about coverage, for the same reason
// check_orphan_element_refs states for itself: the browser tier is
// CRITERIA-GATED. Tier 4 runs only where a PLAN carries a ```criteria fence — 1
// of the 71 components TL-033 made newly eligible — and it never visits
// site_components chrome, ordinary sections or js_snippets bundles at all. This
// check needs no criteria, no PLAN and no browser, so it reaches the population
// Tier 4 structurally cannot: rebuild_policy='owned' pages, the chrome that
// appears on every page, and everything that is not a tool.
//
// ── WHAT THIS CHECK DOES NOT OWN ───────────────────────────────────────────
//
// Four checks already share the asset space and the landmine keyed to
// check_image_url_404.go warns that widening one silently competes with another.
// Each was read before this was written:
//
//   - <img src> — check_image_url_404, which answers it from the DATABASE by
//     inverting storage.DeployedWebPath, with no HTTP at all. Nothing here
//     touches <img>.
//   - favicon / og_card in the site head — check_undeployed_assets' brand-head
//     half, whose population is the purposes themselves so that absence is
//     representable.
//   - "an asset row exists but was never deployed" — check_undeployed_assets.
//   - "a style collection is installed but its css_theme is missing" —
//     check_missing_css. That is a DB question about a row; this is a wire
//     question about a URL, and a site can fail either independently.
//
// ── THE PROBE RULES, and the one that matters most ──────────────────────────
//
// ONLY 404 AND 410 ARE FINDINGS. 2xx, every 3xx after redirects, 401/403/429,
// every 5xx, timeouts, DNS and TLS failures all SKIP — counted, logged, never
// filed.
//
// That is not caution for its own sake, it is the difference between this check
// and a fleet-wide false alarm. A prior sweep of exactly this population fetched
// the tools' external scripts with a bare non-browser HTTP client and Cloudflare
// refused every one: 63 refusals on webdesign.co.uk, the very site this bug is
// about (webdesign_tools_repair/NOTES:492,553 — "would have been 63"). Under a
// "non-200 is a finding" rule that is 63 false 404s on a site whose scripts all
// serve 200. Under this rule a refusal is a 403 and files nothing. The check can
// be blinded by a policy refusal; it cannot be made to lie by one, and given the
// choice those are not equivalent.
//
// A candidate 404 is CONFIRMED by a second request before it is filed. Candidates
// are rare by construction, so the extra call costs nothing in the normal case.
//
// A transport failure is not a status. There is no equivalent here of curl's
// `000`, which has been mistaken for an HTTP result before: an error goes to the
// skip tally with its reason.
//
// ── WHY THIS PROBES AT ALL, given a promise that seems to forbid it ─────────
//
// check_image_url_404.go:21 states the platform keeps "no outbound HTTP on the
// discovery or completion path". Its SOURCE — verifier_coverage_test.go:173-174 —
// says only the COMPLETION path, and the ruling it records
// (work_item_completion_integrity/HANDOFF_2026-07-19) was about verifier
// candidates. Three discovery checks probe the live wire today:
// check_backend_unreachable, check_backend_entry_orphaned, and
// check_tool_acceptance — which already GETs these very domains from the chassis
// with a 12s timeout. So the promise is intact: this check is deliberately NOT a
// verifier candidate, and its classification in verifier_coverage_test.go says so
// in the same words as the other two.
//
// ── WHY THE DOM IS PARSED RATHER THAN THE HTML MATCHED ─────────────────────
//
// A regex over rendered_html cannot tell an ELEMENT from a MENTION of an element.
// The first measurement taken for this bug regexed `<script[^>]+src="([^"]+)"`
// over page_components and returned `<script src="...">` on a live tool page,
// which curl confirmed 404. It was a COMMENT inside that tool's own JavaScript,
// describing a regex:
//
//	// We want to keep anything that looks like <script src="..."> or <link rel="stylesheet">
//
// No browser ever requests it. The population most likely to produce that false
// positive is precisely this bug's population — tool pages, whose content is
// JavaScript that talks about HTML. goquery sees elements only, and the html
// parser treats <script> content as raw text, so a mention cannot be reached.
// TestAssetReference404_ScriptTagInsideJSCommentIsNotAReference pins it.
//
// ── AN EMPTY REFERENCE IS NEVER PROBED ─────────────────────────────────────
//
// Per the HTML spec an empty src resolves against the current document, so the
// page re-requests ITSELF and the probe scores a broken reference 200 —
// bugfix_128's recorded landmine, written for <img> and true verbatim here. An
// empty, whitespace-only or "#" reference is reported structurally as
// kind="empty_src" and never sent to the network.
//
// ── URLS ARE RESOLVED, NEVER CONSTRUCTED ───────────────────────────────────
//
// Every reference is resolved with net/url against https://<domain><pages.url>,
// the URL the page is actually served at. TL-032's own false positive came from a
// built URL — "a verdict from a URL you built is a verdict about a page you
// invented" — and 17 of webdesign.co.uk's references are page-relative
// (`bayes.js`, `js/app.js`), so this is load-bearing rather than tidy. Chrome has
// no page URL, so in site_components only root-relative and absolute references
// can be resolved; a page-relative reference there resolves differently on every
// page and is SKIPPED with a logged reason rather than guessed at.
//
// ── ROUTING ────────────────────────────────────────────────────────────────
//
// Flag-only: HandlerAgent is empty, following check_image_url_404 and
// check_orphan_element_refs. Repairing a 404ing reference means removing it,
// repointing it, or republishing the file it names — a judgement about intent, and
// no generator can make it. The cost of flag-only is real and named in
// bugs_open/083: "a detector whose output nobody drains is not neutral — it is
// actively misleading". The finding surfaces in the run's Findings and as a
// visible 'detected' work item, the same place image_url_404 and dead_control
// land, and it is deliberately not dispatchable: an item with no handler that
// reached a fixer would be marked blocked rather than repaired.
//
// Per the owner ruling of 2026-08-02 §1, a work item type with no automated
// consumer is not the kind of shared vocabulary whose guarantees change when a
// producer is added, so this is normal council-gate scope rather than RFC scope.
//
// ── WHAT IT FINDS TODAY: NOTHING, AND THAT IS RECORDED ON PURPOSE ──────────
//
// Measured 2026-08-05 across all 541 deployed pages: 854 <script src> elements,
// 96 distinct referenced assets, 96 of 96 returning 200. This check is a
// REGRESSION GUARD, not a repair of live damage. The class has bitten —
// bugs_closed/041 published chrome js_content to a path nothing loaded, and
// cmd/webdesignport carries checkScriptParity because roughly 60 of 63 tools
// nearly shipped as dead markup, caught (WRONG_CALLS.md) by luck. A guard with no
// live positive can rot unexercised, so every branch below is proven by an
// induced fault in the test file rather than by hope.
//
// KNOWN, STATED GAPS — a reader should not have to discover these:
//
//   - A reference that is DELETED rather than repaired leaves its work item open.
//     Retraction fires only on a positive observation (a URL still referenced that
//     now returns 200); a URL that has vanished from the HTML cannot be observed
//     at all, and inferring "resolved" from absence is exactly what
//     CheckResult.Resolved's contract forbids.
//   - A same-origin file served by no pipeline (webdesign.co.uk's legacy
//     hand-committed assets) is invisible here — it serves 200, so there is
//     nothing to report. That is the OPPOSITE error profile to
//     check_image_url_404's residual false positive, and it is the strongest
//     argument for probing rather than deriving.
//   - An external host that rate-limits or blocks datacentre IPs is a skip, not a
//     finding, so third-party breakage may go unreported. Findings on external
//     hosts carry external=true in their spec because the repair differs (pin a
//     version rather than republish a file).

package discovery_checks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&AssetReference404Check{}) }

type AssetReference404Check struct{}

func (c *AssetReference404Check) Name() string { return "asset_reference_404" }

const (
	// maxProbeURLs bounds the outbound calls one site can cause. The fleet's
	// busiest site references 6 distinct assets today, so this is headroom
	// rather than a limit that bites — but it is a limit, and when it drops
	// anything it LOGS what it dropped. A silent cap reads as "everything was
	// checked" when it was not.
	maxProbeURLs = 40

	// probeWorkers keeps a slow origin from serialising the whole site. Small on
	// purpose: this runs inside a discovery sweep, not a load test.
	probeWorkers = 4

	// probeTimeout per request. check_tool_acceptance uses 12s for a whole page;
	// a subresource status needs less.
	probeTimeout = 10 * time.Second

	// maxEmptyRefSamples bounds the sample list carried in the spec. The count is
	// always exact; only the examples are capped.
	maxEmptyRefSamples = 5
)

// probeAssetURL is swappable in tests — the same seam fetchDeployedPage uses in
// check_tool_acceptance.go, and the reason the test file can prove every branch
// without a network.
//
// Returns the HTTP status, or (0, err) for a transport failure. A transport
// failure is NOT a status and must never be compared against 404.
var probeAssetURL = func(ctx context.Context, target string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, err
	}
	// GET rather than HEAD: a static origin need not implement HEAD, and its 405
	// would be indistinguishable from a policy refusal. An explicit, honest
	// User-Agent — if an origin refuses it, that is a 403 and files nothing.
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+asset_reference_404)")
	req.Header.Set("Accept", "*/*")

	client := &http.Client{Timeout: probeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	// The status is the whole answer; drain a token amount so the connection can
	// be reused and discard the rest.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	return resp.StatusCode, nil
}

// assetRef is one referenced subresource, already resolved to an absolute URL.
type assetRef struct {
	URL      string // resolved, absolute
	Raw      string // as authored, for the summary
	Element  string // "script" | "stylesheet"
	Surface  string // "page" | "chrome"
	PageURL  string // "" for chrome
	PageID   *uuid.UUID
	External bool // host is not this site's domain
}

// emptyRefTally records references that cannot resolve and must not be probed.
type emptyRefTally struct {
	count   int
	samples []string
}

func (t *emptyRefTally) add(sample string) {
	t.count++
	if len(t.samples) < maxEmptyRefSamples {
		t.samples = append(t.samples, sample)
	}
}

// skipTally records why a reference was not probed or not judged. It exists so
// the run can say what it did NOT look at — a check that reports only findings
// cannot be told apart from one that was blinded.
type skipTally struct {
	unresolvable   int // chrome page-relative, or a src net/url cannot parse
	nonSchemeHTTP  int // mailto:, data:, javascript:
	overCap        int
	transportError int
	refusedOrError int // 401/403/429/5xx and anything else that is not 404/410
}

func (c *AssetReference404Check) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, '') FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("asset_reference_404: site lookup failed: %w", err)
	}
	if domain == "" {
		// No domain, no URL to resolve against. Nothing this check can say.
		return result, nil
	}

	refs, empty, skips, err := collectAssetReferences(dctx, domain)
	if err != nil {
		return nil, err
	}

	// Deduped, sorted: one probe per distinct URL, and a deterministic order so
	// two runs over the same data produce the same sequence of items.
	urls := make([]string, 0, len(refs))
	for u := range refs {
		urls = append(urls, u)
	}
	sort.Strings(urls)

	if len(urls) > maxProbeURLs {
		dropped := urls[maxProbeURLs:]
		skips.overCap = len(dropped)
		dctx.Logger.Warn("asset_reference_404: probe cap reached — these references were NOT checked",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", maxProbeURLs),
			zap.Int("dropped", len(dropped)),
			zap.Strings("dropped_urls", dropped))
		urls = urls[:maxProbeURLs]
	}

	statuses := probeAll(dctx, urls)

	for _, u := range urls {
		ref := refs[u]
		st, ok := statuses[u]
		if !ok {
			continue
		}

		switch {
		case st.err != nil:
			skips.transportError++
			dctx.Logger.Info("asset_reference_404: probe failed, not a finding",
				zap.String("url", u), zap.Error(st.err))
			continue

		case st.code == http.StatusNotFound || st.code == http.StatusGone:
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":       "asset_reference_404",
				"kind":        "unresolvable_reference",
				"url":         u,
				"element":     ref.Element,
				"surface":     ref.Surface,
				"http_status": st.code,
			})
			result.WorkItems = append(result.WorkItems, buildAssetRefWorkItem(dctx, ref, st.code))

		case st.code >= 200 && st.code < 400:
			// A positive observation, and the only thing that may retract.
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "asset_reference_404",
				ItemKey:  assetRefItemKey("unresolvable_reference", u),
				Reason: fmt.Sprintf("%s reference %s now returns HTTP %d",
					ref.Element, u, st.code),
			})

		default:
			// 401/403/429/5xx and anything else. The check is blinded here, not
			// informed — see the header. Never a finding.
			skips.refusedOrError++
			dctx.Logger.Info("asset_reference_404: inconclusive status, not a finding",
				zap.String("url", u), zap.Int("status", st.code))
		}
	}

	if empty.count > 0 {
		spec := map[string]interface{}{
			"check":   "asset_reference_404",
			"kind":    "empty_src",
			"count":   empty.count,
			"samples": empty.samples,
		}
		specJSON, _ := json.Marshal(spec)

		result.Findings = append(result.Findings, map[string]interface{}{
			"check": "asset_reference_404",
			"kind":  "empty_src",
			"count": empty.count,
		})
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:    dctx.SiteID,
			Source:    "discovery",
			Pipeline:  dctx.Pipeline,
			ItemType:  "asset_reference_404",
			Severity:  "medium",
			Summary:   fmt.Sprintf("%d script/stylesheet reference(s) have no usable URL (empty or \"#\")", empty.count),
			SpecJSON:  string(specJSON),
			Priority:  40,
			Status:    "detected",
			CreatedBy: dctx.AgentType,
			ItemKey:   assetRefItemKey("empty_src", ""),
			BatchID:   dctx.BatchID,
			// HandlerAgent intentionally empty — flag-only, see the header.
		})
	}

	logSkips(dctx, skips, len(refs))
	return result, nil
}

// assetRefItemKey is the dedup key. The FULL resolved URL is in it, not the
// basename: `app.js` under two tool directories are two files with two HTTP
// results, and a key that cannot tell them apart lets idx_swi_dedup silently drop
// the second finding — the failure mode of bugs_open/091.
func assetRefItemKey(kind, resolvedURL string) string {
	if resolvedURL == "" {
		return "asset_reference_404:" + kind
	}
	return "asset_reference_404:" + kind + ":" + resolvedURL
}

func buildAssetRefWorkItem(dctx DiscoveryCheckContext, ref assetRef, status int) WorkItemSpec {
	spec := map[string]interface{}{
		"check":       "asset_reference_404",
		"kind":        "unresolvable_reference",
		"url":         ref.URL,
		"reference":   ref.Raw,
		"element":     ref.Element,
		"surface":     ref.Surface,
		"page_url":    ref.PageURL,
		"http_status": status,
		"external":    ref.External,
	}
	specJSON, _ := json.Marshal(spec)

	what := "Page"
	if ref.Surface == "chrome" {
		what = "Site chrome (every page)"
	}
	kindWord := "script"
	if ref.Element == "stylesheet" {
		kindWord = "stylesheet"
	}

	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   ref.PageID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "asset_reference_404",
		// Severity deliberately does not vary by surface: a dead script breaks
		// the page it is on as completely as one in chrome breaks every page.
		// The surface is in the spec, where a reader can act on it.
		Severity: "high",
		Summary: fmt.Sprintf("%s loads %s %s which returns HTTP %d",
			what, kindWord, ref.URL, status),
		SpecJSON:  string(specJSON),
		Priority:  40,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   assetRefItemKey("unresolvable_reference", ref.URL),
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the header.
	}
}

type probeOutcome struct {
	code int
	err  error
}

// probeAll fetches each URL once, confirming a candidate 404 with a second
// request before it can become a finding. An edge that answers 404 once and 200
// on retry has told us the first answer was not about the file.
func probeAll(dctx DiscoveryCheckContext, urls []string) map[string]probeOutcome {
	out := make(map[string]probeOutcome, len(urls))
	if len(urls) == 0 {
		return out
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	work := make(chan string)

	workers := probeWorkers
	if len(urls) < workers {
		workers = len(urls)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range work {
				code, err := probeAssetURL(dctx.Ctx, u)
				if err == nil && (code == http.StatusNotFound || code == http.StatusGone) {
					// Confirm before filing. Only a repeated negative is a finding.
					code2, err2 := probeAssetURL(dctx.Ctx, u)
					if err2 != nil {
						code, err = 0, err2
					} else if code2 != code {
						dctx.Logger.Info("asset_reference_404: candidate 404 not reproduced, discarding",
							zap.String("url", u), zap.Int("first", code), zap.Int("second", code2))
						code = code2
					}
				}
				mu.Lock()
				out[u] = probeOutcome{code: code, err: err}
				mu.Unlock()
			}
		}()
	}
	for _, u := range urls {
		work <- u
	}
	close(work)
	wg.Wait()
	return out
}

func logSkips(dctx DiscoveryCheckContext, s skipTally, probed int) {
	if s.unresolvable == 0 && s.nonSchemeHTTP == 0 && s.overCap == 0 &&
		s.transportError == 0 && s.refusedOrError == 0 {
		return
	}
	dctx.Logger.Info("asset_reference_404: references not judged",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("distinct_referenced", probed),
		zap.Int("unresolvable", s.unresolvable),
		zap.Int("non_http_scheme", s.nonSchemeHTTP),
		zap.Int("over_cap", s.overCap),
		zap.Int("transport_error", s.transportError),
		zap.Int("inconclusive_status", s.refusedOrError))
}

// collectAssetReferences walks both surfaces and returns the distinct resolved
// URLs, keyed by URL so each is probed once. When two pages reference the same
// URL the first wins for reporting purposes; the URL is what the finding is
// about, and the surface is recorded on it.
func collectAssetReferences(dctx DiscoveryCheckContext, domain string) (
	map[string]assetRef, emptyRefTally, skipTally, error) {

	refs := make(map[string]assetRef)
	var empty emptyRefTally
	var skips skipTally

	add := func(html, pageURL, surface string, pageID *uuid.UUID) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		if err != nil {
			dctx.Logger.Warn("asset_reference_404: could not parse rendered html",
				zap.String("surface", surface), zap.Error(err))
			return
		}

		consider := func(raw, element string) {
			trimmed := strings.TrimSpace(raw)
			// Never probed: an empty src resolves against the current document,
			// so the probe would score a broken reference 200.
			if trimmed == "" || trimmed == "#" {
				empty.add(fmt.Sprintf("<%s> on %s", element, surfaceLabel(surface, pageURL)))
				return
			}
			// A scheme we cannot GET is not a broken reference.
			if lower := strings.ToLower(trimmed); strings.HasPrefix(lower, "data:") ||
				strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "mailto:") ||
				strings.HasPrefix(lower, "blob:") {
				skips.nonSchemeHTTP++
				return
			}

			resolved, external, ok := resolveAssetURL(domain, pageURL, trimmed)
			if !ok {
				skips.unresolvable++
				dctx.Logger.Info("asset_reference_404: reference could not be resolved to a URL",
					zap.String("reference", trimmed),
					zap.String("surface", surface),
					zap.String("page_url", pageURL))
				return
			}
			if _, seen := refs[resolved]; seen {
				return
			}
			refs[resolved] = assetRef{
				URL: resolved, Raw: trimmed, Element: element,
				Surface: surface, PageURL: pageURL, PageID: pageID, External: external,
			}
		}

		doc.Find("script[src]").Each(func(_ int, s *goquery.Selection) {
			v, _ := s.Attr("src")
			consider(v, "script")
		})
		// rel is matched by token rather than by equality: "stylesheet preload"
		// and "alternate stylesheet" are both real, and an == comparison misses
		// them. rel="icon" is NOT ours — check_undeployed_assets owns brand-head.
		doc.Find("link[href]").Each(func(_ int, s *goquery.Selection) {
			rel, _ := s.Attr("rel")
			for _, tok := range strings.Fields(strings.ToLower(rel)) {
				if tok == "stylesheet" {
					v, _ := s.Attr("href")
					consider(v, "stylesheet")
					return
				}
			}
		})
	}

	// THE TWO PAGE-LEVEL AXES ARE THE SHARED BUILDERS, NOT A FRESH SPELLING.
	// Added after the council's edit-quality seat objected [medium] that this
	// query hand-rolled its liveness test, and it was right — the first version
	// asked `p.deployed_at IS NOT NULL`, which is the exact shape the landmine on
	// `pages.build_status` warns about from the other side: 28 pages shipped
	// under another status (bugs_open/185), and 35 of 46 needs_rebuild rows carry
	// a deployed_at and are still being served. A probe list built from a fresh
	// spelling silently omits live pages, and a check that omits pages reports
	// clean for the wrong reason.
	//
	//	PageHasShippedPredicateFor  — BUILD axis: a visitor can see this page
	//	PageWantedLivePredicateFor  — LIFECYCLE axis: the platform still wants it
	//
	// Both, because they are independent: archiving sets status='archived' and
	// leaves the build columns untouched (bugs_open/098), so the build axis alone
	// keeps probing the assets of retired pages for ever.
	//
	// `pc.build_status = 'deployed'` below is a DIFFERENT column on a different
	// table — the COMPONENT's render state, as check_image_url_404 uses it — and
	// is not what that landmine is about. Spelled out so the next reviewer does
	// not have to re-derive it.
	pageRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.rendered_html, COALESCE(p.url, ''), p.id
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html IS NOT NULL
		  AND `+datahelpers.PageHasShippedPredicateFor("p")+`
		  AND `+datahelpers.PageWantedLivePredicateFor("p")+`
	`, dctx.SiteID)
	if err != nil {
		return nil, empty, skips, fmt.Errorf("asset_reference_404: page_components scan failed: %w", err)
	}
	for pageRows.Next() {
		var html, pageURL string
		var pageID uuid.UUID
		if err := pageRows.Scan(&html, &pageURL, &pageID); err != nil {
			dctx.Logger.Warn("asset_reference_404: scan page component failed", zap.Error(err))
			continue
		}
		if pageURL == "" {
			// Without the served URL a relative reference cannot be resolved, and
			// a guessed URL is a verdict about a page that does not exist.
			skips.unresolvable++
			continue
		}
		id := pageID
		add(html, pageURL, "page", &id)
	}
	if err := pageRows.Err(); err != nil {
		pageRows.Close()
		return nil, empty, skips, fmt.Errorf("asset_reference_404: page_components scan failed: %w", err)
	}
	pageRows.Close()

	// The chrome surface: one bad reference here is on every page of the site.
	// site_components carries no build_status='deployed' contract comparable to
	// page_components — it is the stored artefact the whole site renders
	// (bugs_open/117) — so every unlocked row is in scope.
	chromeRows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.rendered_html
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.locked_at IS NULL
		  AND sc.rendered_html IS NOT NULL
	`, dctx.SiteID)
	if err != nil {
		return nil, empty, skips, fmt.Errorf("asset_reference_404: site_components scan failed: %w", err)
	}
	defer chromeRows.Close()
	for chromeRows.Next() {
		var html string
		if err := chromeRows.Scan(&html); err != nil {
			dctx.Logger.Warn("asset_reference_404: scan site component failed", zap.Error(err))
			continue
		}
		add(html, "", "chrome", nil)
	}
	return refs, empty, skips, chromeRows.Err()
}

func surfaceLabel(surface, pageURL string) string {
	if surface == "chrome" {
		return "site chrome"
	}
	if pageURL == "" {
		return "a page"
	}
	return pageURL
}

// resolveAssetURL turns an authored reference into the absolute URL a browser on
// that page would request. Returns ok=false when the reference cannot be resolved
// without guessing — which is the case for a PAGE-RELATIVE reference in shared
// chrome, because it resolves to a different URL on every page that renders it.
func resolveAssetURL(domain, pageURL, ref string) (resolved string, external bool, ok bool) {
	refURL, err := url.Parse(ref)
	if err != nil {
		return "", false, false
	}

	if refURL.IsAbs() {
		if refURL.Scheme != "http" && refURL.Scheme != "https" {
			return "", false, false
		}
		return refURL.String(), !sameHost(refURL.Host, domain), true
	}
	// Protocol-relative (//cdn.example/x.js): scheme comes from the page.
	if refURL.Host != "" {
		refURL.Scheme = "https"
		return refURL.String(), !sameHost(refURL.Host, domain), true
	}

	if pageURL == "" {
		// Chrome. Only a rooted reference has one answer for the whole site.
		if !strings.HasPrefix(ref, "/") {
			return "", false, false
		}
		base, err := url.Parse("https://" + domain + "/")
		if err != nil {
			return "", false, false
		}
		return base.ResolveReference(refURL).String(), false, true
	}

	base, err := url.Parse("https://" + domain + pageURL)
	if err != nil {
		return "", false, false
	}
	return base.ResolveReference(refURL).String(), false, true
}

func sameHost(host, domain string) bool {
	h := strings.ToLower(strings.TrimSuffix(host, "."))
	d := strings.ToLower(domain)
	if i := strings.Index(h, ":"); i >= 0 {
		h = h[:i]
	}
	return h == d || h == "www."+d
}
