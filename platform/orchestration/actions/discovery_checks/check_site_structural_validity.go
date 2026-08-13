// FILE: platform/orchestration/actions/discovery_checks/check_site_structural_validity.go
//
// Five discovery checks that ask, of a LIVE deployed page, "is this actually
// correct once served" — a standing, fleet-wide generalisation of the
// single-lane `docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/verify_site.py`,
// whose --live mode caught bugs_open/251 (every assembled homepage's canonical
// pointed at /index.html instead of /) that nothing fleet-wide was checking for.
//
//	dead_internal_link_live  — every internal <a href> on a shipped page resolves
//	                           live (200), confirmed, 404|410-only.
//	canonical_mismatch       — the served page's <link rel="canonical"> is
//	                           present, on-domain, resolves live, and names THIS
//	                           page's own preferred URL — not just "a" URL.
//	structured_data_invalid  — every <script type="application/ld+json"> block on
//	                           the served page is valid JSON. Zero blocks is zero
//	                           findings (verify_site.py's own rule: presence is a
//	                           separate, unbuilt concern here — see below).
//	head_essentials_missing  — the served page has a non-empty <title>, a
//	                           skip-link, and a <footer>.
//	sitemap_entry_dead_live  — every URL a site's OWN /sitemap.xml lists resolves
//	                           live, 404|410-only, confirmed. Deliberately NOT
//	                           "every page appears in the sitemap" — see WHAT IS
//	                           DELIBERATELY NOT GATED below for why that stays
//	                           out. Silent no-op on a site with no sitemap.xml.
//
// ── WHY THIS FETCHES LIVE RATHER THAN READING page_components.rendered_html ──
//
// The canonical/JSON-LD tags this file checks are injected by preferredPageURL's
// callers (injectCanonicalLink / injectPageJSONLD, rerender_single_page_action.go)
// at PAGE-ASSEMBLY time, on top of the stored section HTML — they are not
// guaranteed to be byte-present in page_components.rendered_html, and
// AssemblePageAction (multipage_actions.go, used by three other agent types) does
// not inject them at all (see docs026_concept_register/register/seo.md's standing
// landmine on that split). Only a live GET sees what a visitor's browser sees.
// This is also why a byte-identical live-vs-repo comparison (verify_site.py's
// check 6) is explicitly NOT implemented here: no stored artefact is byte-equal
// to what is served, by construction, so there is nothing to diff against.
//
// ── WHAT IS DELIBERATELY NOT GATED ────────────────────────────────────────────
//
//   - "every page appears in the sitemap": sitemap generation
//     (scripts/site-discovery-files.py, register entry SEO-002) is a manual,
//     rarely-run script here, not a standing mechanism — a site can be entirely
//     healthy and simply never have had the script run against it. Gating on
//     full completeness would be permanently red for almost every site and, per
//     this estate's own recorded lesson (016b §9), a check that is always red is
//     a check nobody reads. sitemap_entry_dead_live is the narrower, safe cousin
//     this note used to defer: "every URL the sitemap DOES list resolves live"
//     — presence in the sitemap is never asserted, only that whatever IS listed
//     still resolves. Skip entirely, silently, on a site with no sitemap.xml (a
//     404, a transport error confirmed on retry, or a body that does not parse
//     as sitemap XML all take this same silent branch — none of them is
//     evidence of a dead URL, only of "nothing this check can say").
//   - og:* social-metadata presence: verify_site.py exempts assembled pages from
//     its og:url check because the shared <head> cannot carry a per-page value
//     (PLAN_2026-08-05 §6, an accepted, stated loss). This file's
//     head_essentials_missing deliberately does NOT check og:url at all, so that
//     exemption has nothing to attach to here — see the check's own header for
//     why building one anyway would be inert from day one.
//
// ── THE PROBE RULES, carried over from check_asset_reference_404.go (read that
//
//	   file first; it is the closest existing template) ───────────────────────
//
//		(a) ONLY 404 AND 410 ARE FINDINGS. Every other status, and any transport
//		    failure, is a SKIP — logged, never filed. A prior sweep of exactly this
//		    kind of population (webdesign_tools_repair/NOTES:492,553) hit Cloudflare
//		    policy refusals under a non-browser client; treating a 403 as a finding
//		    would have been 63 false positives on one site alone.
//		(b) A candidate 404/410 is CONFIRMED with a second request before it is
//		    filed — see probeInternalLinkTargets. Flakiness, not caution for its
//		    own sake.
//		(c) An EMPTY (or bare "#") href is NEVER PROBED: per the HTML spec it
//		    resolves against the current document, so probing it would score a
//		    broken reference 200 — bugfix_128's landmine, true again here.
//
// The OUTER page fetch (the page this check is judging, as opposed to a link
// TARGET) gets the SAME confirm-before-conclude treatment for a NETWORK-
// dependent verdict (transport error / non-2xx) — see fetchAllPagesLive — but
// NOT for a verdict that is deterministic once the body is in hand (a malformed
// ld+json block parses the same way on a retry; retrying it would burn a probe
// and could not change the answer).
//
// ── ROUTING: FLAG-ONLY, THIS PASS ─────────────────────────────────────────────
//
// All five register with HandlerAgent "" — the flag-only idiom check_asset_
// reference_404 and check_site_unreachable also use: the finding surfaces as a
// visible 'detected' work item, and is deliberately not dispatchable. No
// auto-repair agent is wired in this pass. In particular, wiring a repair for
// canonical_mismatch is explicitly future work GATED on bugs_open/251's own fix
// (preferredPageURL, rerender_single_page_action.go) actually being live and
// reachable by every render path that owns a <head> — today it covers only the
// page-rerender path, not AssemblePageAction's three other callers, so an
// auto-fixer today could rewrite a canonical that the next AssemblePageAction
// render would immediately un-fix.
//
// ── SELF-CLEARING, NO COMPLETION-TIME VERIFIER ───────────────────────────────
//
// Same posture as asset_reference_404 and site_unreachable, for the same
// reason: a completion-time verifier would put an outbound HTTP call (a live
// re-fetch of the page) on the completion path, which is the thing those two
// checks' entries in itemTypesWithoutVerifiers already decline for an identical
// mechanism. Each check here retracts its OWN findings through
// CheckResult.Resolved on a positive re-observation each run — the same
// information a verifier would have to fetch, taken on the discovery path
// where the probe is already precedented. See verifier_coverage_test.go's five
// new itemTypesWithoutVerifiers entries for the classification-on-the-way-in.
//
// ── WHY preferredStructuralURL DUPLICATES actions.preferredPageURL ───────────
//
// The `actions` package already imports `discovery_checks` (discovery_checks.go
// and others), so the reverse import needed to call actions.preferredPageURL
// directly would cycle. Two copies of a two-line, test-pinned rule is a smaller
// risk than a package cycle; TestPreferredStructuralURL below pins the same
// three cases preferred_page_url_test.go pins on the actions side (root
// normalises, a non-root path — including one that merely CONTAINS
// "index.html" as a suffix trap — keeps its own full path). If a third
// consumer of this rule ever appears, that is the signal to hoist it into
// datahelpers, which neither package would need to change to depend on.
//
// ── KNOWN, STATED COST ────────────────────────────────────────────────────────
//
// The four PAGE-scoped checks (dead_internal_link_live, canonical_mismatch,
// structured_data_invalid, head_essentials_missing) each independently fetch
// every shipped page live — they do not share one fetch pass, because
// DiscoveryCheck.Run has no channel for one check to hand another its results,
// and check_news_feed.go's own five checks establish that independent,
// overlapping queries are this package's accepted idiom. Enabling all four on
// one agent therefore costs four GETs per page per discovery run, not one. A
// shared per-run page-fetch cache is a real future optimisation, left for the
// dispatch-wiring follow-up along with enabling these checks on a live
// discovery agent's `checks` array — NEITHER is done in this pass; see the
// file's originating task for the reasoning.
//
// sitemap_entry_dead_live has a DIFFERENT, smaller cost shape: one GET of
// /sitemap.xml plus one GET per distinct on-domain <loc> it lists (bounded by
// maxSitemapProbeURLs, capped and logged like the other per-site caps in this
// file) — not one GET per shipped page. It shares no fetch pass with the four
// above because its population is not "this site's pages", it is "whatever
// this site's own sitemap.xml currently claims", which can be a different set
// (a stale sitemap can list a page that has since been removed, or omit one
// that exists) — see WHAT IS DELIBERATELY NOT GATED above.
package discovery_checks

import (
	"context"
	"encoding/json"
	"encoding/xml"
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

func init() {
	Register(&DeadInternalLinkLiveCheck{})
	Register(&CanonicalMismatchCheck{})
	Register(&StructuredDataInvalidCheck{})
	Register(&HeadEssentialsMissingCheck{})
	Register(&SitemapEntryDeadCheck{})
}

// ---------------------------------------------------------------------------
// shared population + live fetch, used by all four checks
// ---------------------------------------------------------------------------

const (
	// structuralProbeTimeout per page fetch. Matches check_site_unreachable's
	// per-site budget: fetching one page's full body is a comparable cost to
	// probing a whole site's apex.
	structuralProbeTimeout = 15 * time.Second

	// structuralProbeWorkers bounds concurrency per check, per site — small on
	// purpose, this runs inside a discovery sweep, not a load test. Shared by
	// the page-body fetch and the dead-link-target probe.
	structuralProbeWorkers = 4

	// structuralBodyCap bounds what one page fetch will read. Generous next to
	// check_site_unreachable's 256KiB apex cap: a page carrying several ld+json
	// blocks plus a full article body can legitimately be larger.
	structuralBodyCap = 1 << 20 // 1MiB

	// maxPagesProbedPerSite bounds outbound calls one site can cause per check,
	// per run. Headroom, not a limit expected to bite: measured 2026-08-05
	// (asset_reference_404's header), 541 deployed pages across ~23 sites is
	// ~23/site on average. Exceeding it is LOGGED, never silent.
	maxPagesProbedPerSite = 150

	// maxDeadLinkProbeURLs bounds distinct internal link TARGETS probed per
	// site — a different population from the pages above (a target need not be
	// one of this site's own pages at all). Logged, not silent, when it bites.
	maxDeadLinkProbeURLs = 80

	// maxSitemapProbeURLs bounds distinct on-domain <loc> entries probed per
	// site's sitemap.xml, same reasoning and same figure as maxDeadLinkProbeURLs
	// above (a different population — this site's OWN claimed URL list, not its
	// pages table — but the same "bound the outbound call count, log rather than
	// silently drop" discipline). Logged, not silent, when it bites.
	maxSitemapProbeURLs = 80
)

// structuralRetryWait before a confirming second attempt on a page fetch that
// initially transport-errored or returned a non-2xx. A var, not a const, so
// tests need not sleep through it — same reason check_site_unreachable's
// siteProbeRetryWait is a var.
var structuralRetryWait = 5 * time.Second

// structuralPage is one row of this check family's population.
type structuralPage struct {
	ID   uuid.UUID
	Name string
	URL  string // pages.url, root-relative, e.g. "/index.html"
}

// pageFetch is one page's live-fetch outcome.
type pageFetch struct {
	Page         structuralPage
	PreferredURL string // absolute, root-normalised — what a visitor's browser bar shows
	Status       int    // 0 when TransportErr is set
	Body         string
	TransportErr string
}

// loadStructuralDomain returns the site's domain, "" if it has none. Split out
// of loadStructuralPopulation below (rather than duplicated) because
// SitemapEntryDeadCheck needs the domain but not the per-page population —
// a sitemap is one fetch for the whole site, not one per page.
func loadStructuralDomain(dctx DiscoveryCheckContext) (string, error) {
	var domain string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, '') FROM sites WHERE id = $1`, dctx.SiteID,
	).Scan(&domain); err != nil {
		return "", fmt.Errorf("check_site_structural_validity: site lookup failed: %w", err)
	}
	return domain, nil
}

// loadStructuralPopulation returns the site's domain and every page a live
// probe should visit: shipped (PageHasShippedPredicateFor — NOT a hand-typed
// build_status filter, per bugs_open/185: 28 live pages ship under a
// non-'deployed' status) AND still wanted (PageWantedLivePredicateFor — the
// lifecycle axis, independent of the build axis per bugs_open/098). Both,
// exactly mirroring check_asset_reference_404.go's population query and cited
// reasoning.
func loadStructuralPopulation(dctx DiscoveryCheckContext) (domain string, pages []structuralPage, err error) {
	domain, err = loadStructuralDomain(dctx)
	if err != nil {
		return "", nil, err
	}
	if domain == "" {
		// No domain, no URL to probe against. Nothing this check can say.
		return "", nil, nil
	}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id, p.name, COALESCE(p.url, '')
		FROM pages p
		WHERE p.site_id = $1
		  AND COALESCE(p.url, '') <> ''
		  AND `+datahelpers.PageHasShippedPredicateFor("p")+`
		  AND `+datahelpers.PageWantedLivePredicateFor("p")+`
		ORDER BY p.name
	`, dctx.SiteID)
	if err != nil {
		return domain, nil, fmt.Errorf("check_site_structural_validity: pages query failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var p structuralPage
		if err := rows.Scan(&p.ID, &p.Name, &p.URL); err != nil {
			return domain, nil, fmt.Errorf("check_site_structural_validity: page scan failed: %w", err)
		}
		if !strings.HasPrefix(p.URL, "/") {
			continue // cannot resolve an absolute URL without a root-relative path
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		return domain, nil, err
	}
	return domain, pages, nil
}

// preferredStructuralURL builds the absolute URL a live probe fetches, and the
// value canonical_mismatch expects a page's own canonical to name. See the file
// header for why this duplicates actions.preferredPageURL rather than importing
// it, and keep the two in lockstep: the ONE normalisation is the site root
// ("/index.html" -> "/"), deliberately root-only — a section index like
// /guides/index.html keeps its full form, because directory URLs 404 on this
// hosting (measured for bugs_open/251, 2026-08-11: /guides/, /loans/, /blog/
// all 404 across three live domains; only the bare root serves).
func preferredStructuralURL(domain, pageURL string) string {
	if pageURL == "/index.html" {
		pageURL = "/"
	}
	return "https://" + domain + pageURL
}

// fetchStructuralPage GETs one absolute URL and returns its status and body.
// Swappable in tests — the same seam probeAssetURL/probeSiteOrigin use.
var fetchStructuralPage = func(ctx context.Context, absoluteURL string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, absoluteURL, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+site_structural_validity)")
	req.Header.Set("Accept", "text/html,*/*")

	client := &http.Client{Timeout: structuralProbeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, structuralBodyCap))
	return resp.StatusCode, string(body), nil
}

// fetchAllPagesLive fetches every page's preferred URL, confirming a failed
// fetch (transport error or non-2xx) with one retry before treating the page as
// unreadable — a single blip must not blind every check for a whole page.
// Bounded by maxPagesProbedPerSite; a drop is logged, never silent.
func fetchAllPagesLive(dctx DiscoveryCheckContext, domain string, pages []structuralPage) ([]pageFetch, int) {
	capped := pages
	dropped := 0
	if len(capped) > maxPagesProbedPerSite {
		dropped = len(capped) - maxPagesProbedPerSite
		capped = capped[:maxPagesProbedPerSite]
	}

	out := make([]pageFetch, len(capped))
	var wg sync.WaitGroup
	work := make(chan int)
	workers := structuralProbeWorkers
	if len(capped) < workers {
		workers = len(capped)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range work {
				p := capped[idx]
				abs := preferredStructuralURL(domain, p.URL)
				status, body, err := fetchStructuralPage(dctx.Ctx, abs)
				if err != nil || status < 200 || status > 299 {
					time.Sleep(structuralRetryWait)
					status, body, err = fetchStructuralPage(dctx.Ctx, abs)
				}
				fr := pageFetch{Page: p, PreferredURL: abs, Status: status, Body: body}
				if err != nil {
					fr.TransportErr = err.Error()
				}
				out[idx] = fr
			}
		}()
	}
	for i := range capped {
		work <- i
	}
	close(work)
	wg.Wait()

	if dropped > 0 {
		dctx.Logger.Warn("check_site_structural_validity: page probe cap reached — these pages were NOT checked",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", maxPagesProbedPerSite),
			zap.Int("dropped", dropped))
	}
	return out, dropped
}

// structuralItemKey is the dedup key shape shared by the three page-scoped
// checks below (canonical_mismatch, structured_data_invalid,
// head_essentials_missing): one item per page, keyed on the page's own id so
// two pages can never collide.
func structuralItemKey(itemType string, pageID uuid.UUID) string {
	return itemType + ":" + pageID.String()
}

// ---------------------------------------------------------------------------
// structuralProbeOutcome + the link-target probe, mirroring check_asset_
// reference_404.go's probeAll shape (see that file's header for why each rule
// is shaped the way it is). Written fresh rather than calling into that file's
// private probeAll/probeAssetURL: those carry that check's own name in their
// User-Agent and log lines, and reaching into them from here would be a
// coupling a reader of check_asset_reference_404.go alone would not expect.
// resolveAssetURL and sameHost, below, ARE reused directly — they are pure,
// generic URL helpers with no check-specific behaviour or logging baked in.
// ---------------------------------------------------------------------------

type structuralProbeOutcome struct {
	code int
	err  error
}

// probeInternalLinkTarget GETs one URL and returns its status. Swappable in
// tests. Deliberately status-only (no body retained) — the dead-link-target
// population can be much larger than the page population above, and nothing
// here needs the bytes.
var probeInternalLinkTarget = func(ctx context.Context, target string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+dead_internal_link_live)")
	req.Header.Set("Accept", "text/html,*/*")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
	return resp.StatusCode, nil
}

// probeInternalLinkTargets fetches each URL once, confirming a candidate
// 404/410 with a second request before it can become a finding — rule (b) from
// the file header, carried over verbatim from check_asset_reference_404.go's
// probeAll.
func probeInternalLinkTargets(dctx DiscoveryCheckContext, urls []string) map[string]structuralProbeOutcome {
	out := make(map[string]structuralProbeOutcome, len(urls))
	if len(urls) == 0 {
		return out
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	work := make(chan string)

	workers := structuralProbeWorkers
	if len(urls) < workers {
		workers = len(urls)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range work {
				code, err := probeInternalLinkTarget(dctx.Ctx, u)
				if err == nil && (code == http.StatusNotFound || code == http.StatusGone) {
					code2, err2 := probeInternalLinkTarget(dctx.Ctx, u)
					if err2 != nil {
						code, err = 0, err2
					} else if code2 != code {
						dctx.Logger.Info("dead_internal_link_live: candidate 404 not reproduced, discarding",
							zap.String("url", u), zap.Int("first", code), zap.Int("second", code2))
						code = code2
					}
				}
				mu.Lock()
				out[u] = structuralProbeOutcome{code: code, err: err}
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

// ===========================================================================
// 1. dead_internal_link_live
// ===========================================================================
//
// WHAT THIS CHECK DOES NOT OWN, mirroring asset_reference_404's own such
// section: <script src>/<link rel=stylesheet> belong to asset_reference_404;
// <img src> belongs to check_image_url_404; a link whose target is not a real
// `pages` row at all (a "phantom") or whose target IS a real row that has
// simply never been deployed ("unbuilt") are check_phantom_internal_links'
// DB-derived verdicts, over STORED rendered_html.
//
// This check deliberately OVERLAPS both of those in one respect: it probes
// EVERY internal <a href> target LIVE, including ones that would already
// classify as phantom or unbuilt, rather than excluding them. That is the same
// choice asset_reference_404 makes relative to check_image_url_404 (probe
// rather than derive), and for the same reason: a same-origin path served by
// no pipeline the DB knows about is invisible to a DB-only check, and a page
// the DB believes is fine can still 404 live (deploy drift, the single-page
// shape of bugs_open/236). The two failure profiles are opposite, which is the
// argument for keeping both checks rather than merging them.
//
// NOR does this duplicate the platform's other BUILD-TIME and RENDER-TIME
// internal-link machinery — same shape of distinction as above, extended to
// four more mechanisms, every one of which answers "what does the DATABASE
// believe about this link", never "what does a live GET of it actually
// return":
//
//   - loadResolverPageSet (resolve_internal_links_action.go) builds the set of
//     valid page URLs from `pages.url` for internal-link-resolver to resolve
//     authored link tokens against WHILE A PAGE IS BEING BUILT. Runs once per
//     build, reads the DB, issues no outbound request.
//   - markStaleChromeLinkSlot (chrome_link_policy.go) tests a site_component's
//     STORED rendered_html against a ChromeLinkPolicy and marks the slot
//     build_status='pending' for a forced rerender when a stored href violates
//     policy — a judgement about what is already IN the database, not about
//     what is currently being served.
//   - datahelpers.RepairPageLinks (link_repair.go) rewrites hrefs in HTML text
//     in memory, against a DB-derived PageURLIndex, as content is saved. No
//     network call.
//   - datahelpers' content_data_links.go census (bugs_open/097) goes a step
//     earlier still, over stored content_data before it is even rendered to
//     HTML.
//
// And check_phantom_internal_links.go's unbuilt_internal_link /
// phantom_internal_link findings — in THIS SAME PACKAGE — classify a link
// purely by matching its target against the `pages` table (does a row exist;
// has it ever been deployed), reading page_components.rendered_html /
// site_components.rendered_html. Also never fetches the live URL.
//
// dead_internal_link_live is the only one of these six mechanisms that asks
// the question none of them can: as this link is actually served RIGHT NOW,
// does a real GET of its target return 200? That can diverge from every
// DB-derived answer above in both directions — a page the DB believes is built
// and deployed can still 404 live (deploy/CDN drift), and a link every
// DB-only check would wave through as "resolves to a real, deployed page" is
// exactly the class this check exists to catch when that belief turns out to
// be stale.
type DeadInternalLinkLiveCheck struct{}

func (c *DeadInternalLinkLiveCheck) Name() string { return "dead_internal_link_live" }

// deadLinkTarget is one distinct internal link target, resolved to an absolute
// URL, with the first page that referenced it kept for reporting — the same
// "first wins" rule asset_reference_404 uses when several pages share a target.
type deadLinkTarget struct {
	URL      string
	PageID   uuid.UUID
	PageURL  string
	External bool
}

func (c *DeadInternalLinkLiveCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	domain, pages, err := loadStructuralPopulation(dctx)
	if err != nil {
		return nil, err
	}
	if domain == "" || len(pages) == 0 {
		return result, nil
	}

	fetches, _ := fetchAllPagesLive(dctx, domain, pages)

	targets := map[string]deadLinkTarget{}
	var emptyCount int
	for _, f := range fetches {
		if f.Status < 200 || f.Status > 299 {
			continue // this page could not be read live; out of THIS check's remit
		}
		for _, a := range datahelpers.ExtractAnchors(f.Body) {
			href := strings.TrimSpace(a.Href)
			scope := datahelpers.ClassifyLinkScope(href)
			if scope == datahelpers.LinkScopeEmpty {
				// NEVER PROBED — rule (c). An empty href resolves against the
				// current document and would score a broken reference 200.
				emptyCount++
				continue
			}
			if scope != datahelpers.LinkScopePage {
				continue // anchor/external/mailto/asset: another check's remit
			}
			path, _ := datahelpers.SplitFragment(href)
			if path == "" {
				// A bare "#fragment" — resolves against this document, same as
				// rule (c) above, just arrived at via a different href shape.
				emptyCount++
				continue
			}
			resolved, external, ok := resolveAssetURL(domain, f.Page.URL, path)
			if !ok {
				continue
			}
			if _, seen := targets[resolved]; seen {
				continue
			}
			targets[resolved] = deadLinkTarget{URL: resolved, PageID: f.Page.ID, PageURL: f.Page.URL, External: external}
		}
	}

	urls := make([]string, 0, len(targets))
	for u := range targets {
		urls = append(urls, u)
	}
	sort.Strings(urls)

	if len(urls) > maxDeadLinkProbeURLs {
		dropped := urls[maxDeadLinkProbeURLs:]
		dctx.Logger.Warn("dead_internal_link_live: probe cap reached — these link targets were NOT checked",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", maxDeadLinkProbeURLs),
			zap.Int("dropped", len(dropped)))
		urls = urls[:maxDeadLinkProbeURLs]
	}

	outcomes := probeInternalLinkTargets(dctx, urls)
	for _, u := range urls {
		t := targets[u]
		o := outcomes[u]
		switch {
		case o.err != nil:
			// A transport failure is not a status. Skip, never a finding.
			continue

		case o.code == http.StatusNotFound || o.code == http.StatusGone:
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":                "dead_internal_link_live",
				"url":                  u,
				"http_status":          o.code,
				"linked_from_page_id":  t.PageID.String(),
				"linked_from_page_url": t.PageURL,
				"external":             t.External,
			})
			result.WorkItems = append(result.WorkItems, buildDeadLinkWorkItem(dctx, t, o.code))

		case o.code >= 200 && o.code < 400:
			// A positive observation, and the only thing that may retract.
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "dead_internal_link_live",
				ItemKey:  "dead_internal_link_live:" + u,
				Reason:   fmt.Sprintf("re-probed %s: now returns HTTP %d", u, o.code),
			})

		default:
			// 401/403/429/5xx and anything else. Rule (a): never a finding.
			continue
		}
	}

	if emptyCount > 0 {
		dctx.Logger.Info("dead_internal_link_live: empty/self-referencing hrefs skipped",
			zap.String("site_id", dctx.SiteID.String()), zap.Int("count", emptyCount))
	}
	return result, nil
}

func buildDeadLinkWorkItem(dctx DiscoveryCheckContext, t deadLinkTarget, status int) WorkItemSpec {
	spec := map[string]interface{}{
		"check":            "dead_internal_link_live",
		"url":              t.URL,
		"http_status":      status,
		"linked_from_page": t.PageURL,
		"external":         t.External,
	}
	specJSON, _ := json.Marshal(spec)

	pageID := t.PageID
	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   &pageID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "dead_internal_link_live",
		// Matches asset_reference_404's choice for the same shape of defect: a
		// browser fetch that 404s is a broken user-facing experience regardless
		// of which page linked to it.
		Severity: "high",
		Summary: fmt.Sprintf("Internal link to %s (from %s) returns HTTP %d",
			t.URL, t.PageURL, status),
		SpecJSON:  string(specJSON),
		Priority:  40,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   "dead_internal_link_live:" + t.URL,
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the file header.
	}
}

// ===========================================================================
// 2. canonical_mismatch
// ===========================================================================
//
// This is the check that generalises bugs_open/251 into a standing regression
// guard. It does NOT fix 251 and does NOT hand-roll "https://"+domain+url the
// way the bug did — canonicalVerdict compares the served tag against
// preferredStructuralURL's output, the same root-only normalisation
// preferredPageURL now applies at render time.
type CanonicalMismatchCheck struct{}

func (c *CanonicalMismatchCheck) Name() string { return "canonical_mismatch" }

// canonicalVerdict is judgeCanonical's pure verdict — no network inside the
// judge itself, matching check_site_unreachable's judgeSiteProbe discipline.
type canonicalVerdict struct {
	OK     bool
	Reason string // "missing" | "not_absolute" | "off_domain" | "not_live" | "wrong_target"; "" when OK
	Detail string
	Actual string // the href found in the page, "" if none
}

// judgeCanonical asserts, in order: a canonical tag is present; it is an
// absolute http(s) URL on this site's own domain; it names THIS page's
// preferred URL exactly; and (only once the first three hold, since probing an
// already-wrong target teaches nothing new) it resolves live. resolvesLive is
// injected so this function stays network-free and table-testable.
func judgeCanonical(domain, expectedAbs, body string, resolvesLive func(href string) (bool, string)) canonicalVerdict {
	href, ok := extractCanonicalHref(body)
	if !ok {
		return canonicalVerdict{Reason: "missing", Detail: "no <link rel=\"canonical\"> in the served head"}
	}
	u, err := url.Parse(href)
	if err != nil || u.Scheme == "" || u.Host == "" {
		// Distinct from off_domain: this canonical names no host at all (a
		// relative href, or something url.Parse cannot make sense of), so
		// "different domain" would misdescribe it. injectCanonicalLink always
		// emits an absolute https:// URL, so this shape never comes from the
		// correct code path.
		return canonicalVerdict{Reason: "not_absolute", Actual: href,
			Detail: fmt.Sprintf("canonical %q is not an absolute http(s) URL", href)}
	}
	if !sameHost(u.Host, domain) {
		return canonicalVerdict{Reason: "off_domain", Actual: href,
			Detail: fmt.Sprintf("canonical names a different host: %s", u.Host)}
	}
	if href != expectedAbs {
		return canonicalVerdict{Reason: "wrong_target", Actual: href,
			Detail: fmt.Sprintf("expected %s", expectedAbs)}
	}
	if resolvesLive != nil {
		if okLive, detail := resolvesLive(href); !okLive {
			return canonicalVerdict{Reason: "not_live", Actual: href, Detail: detail}
		}
	}
	return canonicalVerdict{OK: true, Actual: href}
}

// extractCanonicalHref is goquery-based rather than a regex over the body, for
// the same reason check_asset_reference_404.go parses the DOM for <script src>:
// a regex cannot tell an ELEMENT from a MENTION of one (that file's own
// TestAssetReference404_ScriptTagInsideJSCommentIsNotAReference is the proof).
func extractCanonicalHref(body string) (string, bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return "", false
	}
	sel := doc.Find(`link[rel="canonical"]`).First()
	if sel.Length() == 0 {
		return "", false
	}
	href, exists := sel.Attr("href")
	href = strings.TrimSpace(href)
	if !exists || href == "" {
		return "", false
	}
	return href, true
}

func (c *CanonicalMismatchCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	domain, pages, err := loadStructuralPopulation(dctx)
	if err != nil {
		return nil, err
	}
	if domain == "" || len(pages) == 0 {
		return result, nil
	}

	fetches, _ := fetchAllPagesLive(dctx, domain, pages)

	resolvesLive := func(href string) (bool, string) {
		code, err := probeInternalLinkTarget(dctx.Ctx, href)
		if err != nil || code == http.StatusNotFound || code == http.StatusGone {
			// Confirm before filing — rule (b), applied to the canonical TARGET
			// rather than a link target, same reasoning.
			code2, err2 := probeInternalLinkTarget(dctx.Ctx, href)
			if err2 == nil {
				code, err = code2, nil
			}
		}
		if err != nil {
			return false, "transport error probing the canonical target: " + err.Error()
		}
		if code < 200 || code >= 300 {
			return false, fmt.Sprintf("canonical target returns HTTP %d", code)
		}
		return true, ""
	}

	for _, f := range fetches {
		if f.Status < 200 || f.Status > 299 {
			continue // page itself unreadable live; out of this check's remit
		}
		expected := preferredStructuralURL(domain, f.Page.URL)
		v := judgeCanonical(domain, expected, f.Body, resolvesLive)

		if v.OK {
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "canonical_mismatch",
				ItemKey:  structuralItemKey("canonical_mismatch", f.Page.ID),
				Reason:   fmt.Sprintf("re-probed %s: canonical correctly names %s", f.PreferredURL, expected),
			})
			continue
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":    "canonical_mismatch",
			"page_id":  f.Page.ID.String(),
			"page_url": f.Page.URL,
			"reason":   v.Reason,
			"detail":   v.Detail,
			"actual":   v.Actual,
			"expected": expected,
		})
		result.WorkItems = append(result.WorkItems, buildCanonicalWorkItem(dctx, f.Page, v, expected))
	}
	return result, nil
}

func buildCanonicalWorkItem(dctx DiscoveryCheckContext, p structuralPage, v canonicalVerdict, expected string) WorkItemSpec {
	spec := map[string]interface{}{
		"check":              "canonical_mismatch",
		"page_url":           p.URL,
		"reason":             v.Reason,
		"detail":             v.Detail,
		"actual_canonical":   v.Actual,
		"expected_canonical": expected,
	}
	specJSON, _ := json.Marshal(spec)

	pageID := p.ID
	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   &pageID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "canonical_mismatch",
		// bugs_open/251's own framing: "an SEO correctness defect, not an
		// outage" — both URLs still serve the page. Medium, not high.
		Severity: "medium",
		Summary: fmt.Sprintf("Page %s canonical is wrong (%s): %s",
			p.Name, v.Reason, v.Detail),
		SpecJSON:  string(specJSON),
		Priority:  60,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   structuralItemKey("canonical_mismatch", p.ID),
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the file header.
	}
}

// ===========================================================================
// 3. structured_data_invalid
// ===========================================================================
type StructuredDataInvalidCheck struct{}

func (c *StructuredDataInvalidCheck) Name() string { return "structured_data_invalid" }

// ldJSONBlock is one <script type="application/ld+json"> element's content and
// its parse verdict.
type ldJSONBlock struct {
	Index int
	Raw   string
	Err   string // "" means it parsed
}

// extractAndValidateLDJSON returns every ld+json block on the page, parsed or
// not. Zero blocks returns a nil/empty slice — that is a legitimate, common
// state (see the check's own header) and callers must not treat it as a
// finding.
//
// goquery, not a regex: verify_site.py's own regex
// (`<script type="application/ld\+json">(.*?)</script>`) is fine for one known
// site; check_asset_reference_404.go's header records why a regex over
// rendered_html mistook a JS comment describing a <script> tag for a real one.
// The same class of false positive is possible here if a page ever prints
// example ld+json inside a code sample.
func extractAndValidateLDJSON(body string) []ldJSONBlock {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil
	}
	var out []ldJSONBlock
	doc.Find(`script[type="application/ld+json"]`).Each(func(i int, s *goquery.Selection) {
		raw := s.Text()
		var v interface{}
		errStr := ""
		if uerr := json.Unmarshal([]byte(raw), &v); uerr != nil {
			errStr = uerr.Error()
		}
		out = append(out, ldJSONBlock{Index: i, Raw: raw, Err: errStr})
	})
	return out
}

func (c *StructuredDataInvalidCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	domain, pages, err := loadStructuralPopulation(dctx)
	if err != nil {
		return nil, err
	}
	if domain == "" || len(pages) == 0 {
		return result, nil
	}

	fetches, _ := fetchAllPagesLive(dctx, domain, pages)

	for _, f := range fetches {
		if f.Status < 200 || f.Status > 299 {
			continue
		}
		blocks := extractAndValidateLDJSON(f.Body)

		var bad []map[string]interface{}
		for _, b := range blocks {
			if b.Err == "" {
				continue
			}
			bad = append(bad, map[string]interface{}{
				"block_index": b.Index,
				"error":       b.Err,
				"snippet":     datahelpers.TruncateString(strings.TrimSpace(b.Raw), 200),
			})
		}

		// Zero blocks, or every block parses: a positive observation. Cheap and
		// safe to emit unconditionally — resolveWorkItems is a plain UPDATE that
		// no-ops when nothing matches the key (work_items_common.go), so this
		// never errors on a page that never had a finding filed against it.
		if len(bad) == 0 {
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "structured_data_invalid",
				ItemKey:  structuralItemKey("structured_data_invalid", f.Page.ID),
				Reason: fmt.Sprintf("re-probed %s: every ld+json block parses (%d checked)",
					f.PreferredURL, len(blocks)),
			})
			continue
		}

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":          "structured_data_invalid",
			"page_id":        f.Page.ID.String(),
			"page_url":       f.Page.URL,
			"invalid_blocks": len(bad),
			"total_blocks":   len(blocks),
		})
		result.WorkItems = append(result.WorkItems, buildStructuredDataWorkItem(dctx, f.Page, bad, len(blocks)))
	}
	return result, nil
}

func buildStructuredDataWorkItem(dctx DiscoveryCheckContext, p structuralPage, bad []map[string]interface{}, total int) WorkItemSpec {
	spec := map[string]interface{}{
		"check":          "structured_data_invalid",
		"page_url":       p.URL,
		"invalid_blocks": bad,
		"total_blocks":   total,
	}
	specJSON, _ := json.Marshal(spec)

	pageID := p.ID
	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   &pageID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "structured_data_invalid",
		Severity: "medium",
		Summary: fmt.Sprintf("Page %s: %d of %d ld+json block(s) do not parse as JSON",
			p.Name, len(bad), total),
		SpecJSON:  string(specJSON),
		Priority:  60,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   structuralItemKey("structured_data_invalid", p.ID),
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the file header.
	}
}

// ===========================================================================
// 4. head_essentials_missing
// ===========================================================================
//
// Checks exactly three things: <title> present and non-empty, a skip-link, a
// <footer>. Deliberately NOT og:url or any other og:* tag.
//
// ── WHY THIS DOES NOT DUPLICATE check_missing_structure.go ───────────────────
//
// CORRECTED 2026-08-12 (council round 2): this section originally claimed
// check_missing_structure.go's `pages.rendered_header/rendered_footer IS NULL`
// predicate was a meaningful "was a header/footer ever rendered" signal. It is
// not — per a standing landmine (LANDMINES.md, "pages.rendered_header /
// rendered_footer / rendered_head are VESTIGIAL") those three columns are
// EMPTY ON EVERY PAGE FLEET-WIDE (re-verified live 2026-08-13: 0 of 683 pages
// have a non-empty rendered_header or rendered_footer). Chrome is written to
// `site_components`, not these `pages` columns, and has been for some time;
// `findPagesWithMissingStructure`'s only other filter is `p.status IN
// ('active','deployed')`, which — per the OTHER landmine on this same
// file, `pages.status` never actually taking the value 'deployed' —
// matches nearly every non-archived page. So MissingStructureCheck's
// predicate is not a discriminating "build-time completeness" question at
// all; it is, on the evidence, universally true, which is a DIFFERENT and
// more fundamental reason this is not a duplicate: a check that (on the
// available evidence) cannot discriminate provides no real signal to overlap
// WITH. head_essentials_missing checks the actual thing a visitor receives —
// an HTTP GET of the live served page, non-empty <title>, a skip-link, a
// <footer> genuinely present in the response body — which is real, current
// information regardless of what MissingStructureCheck's own columns say.
// MissingStructureCheck is NOT dormant, and this was measured rather than
// hedged (2026-08-13): it is named in completeness-discovery-agent's live
// checks config and has filed 43 needs_rerender items since 2026-04-24
// (25 completed — i.e. ~25 full-site rerenders dispatched fleet-wide on a
// predicate that cannot be false). That is its own defect, filed as
// bugs_open/270 with the full census — repairing or retiring it is that
// bug's question, not this file's.
//
// WHY THIS CHECK BUILDS NO "assembled page" EXEMPTION, though verify_site.py's
// equivalent check has one. verify_site.py's exemption (ASSEMBLED_MARKER /
// OG_PER_PAGE) fires ONLY for og:url — its own source shows title, skip-link
// and footer are checked with the identical assembled-or-not code path and
// NEVER exempted, because none of the three is a shared-<head>-only field the
// way og:url is: title is per-page (pages.title, injected on assembled pages
// too), and skip-link/footer are chrome elements a visitor needs regardless of
// how the page was built. Since this check's scope excludes og:url entirely
// (by this task's own design), there is no case in which an exemption here
// would ever fire. Building one anyway would be exactly the failure this
// codebase's own convention warns against: a mechanism that is silently inert
// from the day it ships is indistinguishable from one that is broken (016b
// §9's "a gate's 0 findings has two causes"). So: no suppression. Each finding
// DOES carry an "assembled" flag (pageIsAssembled) as context for whoever
// triages it — informational, never gating — because whether the affected page
// went through the section-based build pipeline is a genuinely useful thing
// for a human to know even though it never changes the verdict.
type HeadEssentialsMissingCheck struct{}

func (c *HeadEssentialsMissingCheck) Name() string { return "head_essentials_missing" }

// headEssentials reports the three assertions. goquery throughout: a regex for
// "is there a <footer> anywhere" is fine, but the skip-link test needs an
// attribute match (href="#content") that a regex would either over- or
// under-match depending on quoting and attribute order.
func headEssentials(body string) (hasTitle, hasSkipLink, hasFooter bool) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return false, false, false
	}
	hasTitle = strings.TrimSpace(doc.Find("title").First().Text()) != ""
	hasFooter = doc.Find("footer").Length() > 0
	// Two independent structural signals, deliberately not a copy of
	// verify_site.py's literal class string as the ONLY one: `class="skip-link"`
	// is one site's styling convention, and `href="#content"` is the fleet
	// convention named in check_phantom_internal_links_fragments.go ("header
	// skip-link targets id=\"content\", which its pages carry"). Either
	// satisfies — a site may use one, the other, or both.
	hasSkipLink = doc.Find(`a[href="#content"]`).Length() > 0 || doc.Find(`.skip-link`).Length() > 0
	return
}

// pageIsAssembled reports whether the page has been built through the
// section-based pipeline (any page_components row), as opposed to being a
// hand-built document with no such rows. Best-effort: an error here never
// blocks a finding, it only means the finding's context is incomplete.
func pageIsAssembled(dctx DiscoveryCheckContext, pageID uuid.UUID) bool {
	var assembled bool
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT EXISTS (SELECT 1 FROM page_components WHERE page_id = $1)`, pageID,
	).Scan(&assembled); err != nil {
		dctx.Logger.Info("head_essentials_missing: could not determine assembly status",
			zap.String("page_id", pageID.String()), zap.Error(err))
		return false
	}
	return assembled
}

func (c *HeadEssentialsMissingCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	domain, pages, err := loadStructuralPopulation(dctx)
	if err != nil {
		return nil, err
	}
	if domain == "" || len(pages) == 0 {
		return result, nil
	}

	fetches, _ := fetchAllPagesLive(dctx, domain, pages)

	for _, f := range fetches {
		if f.Status < 200 || f.Status > 299 {
			continue
		}
		hasTitle, hasSkipLink, hasFooter := headEssentials(f.Body)

		var missing []string
		if !hasTitle {
			missing = append(missing, "title")
		}
		if !hasSkipLink {
			missing = append(missing, "skip_link")
		}
		if !hasFooter {
			missing = append(missing, "footer")
		}

		if len(missing) == 0 {
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "head_essentials_missing",
				ItemKey:  structuralItemKey("head_essentials_missing", f.Page.ID),
				Reason:   fmt.Sprintf("re-probed %s: title, skip-link and footer all present", f.PreferredURL),
			})
			continue
		}

		assembled := pageIsAssembled(dctx, f.Page.ID)
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "head_essentials_missing",
			"page_id":   f.Page.ID.String(),
			"page_url":  f.Page.URL,
			"missing":   missing,
			"assembled": assembled,
		})
		result.WorkItems = append(result.WorkItems, buildHeadEssentialsWorkItem(dctx, f.Page, missing, assembled))
	}
	return result, nil
}

func buildHeadEssentialsWorkItem(dctx DiscoveryCheckContext, p structuralPage, missing []string, assembled bool) WorkItemSpec {
	spec := map[string]interface{}{
		"check":     "head_essentials_missing",
		"page_url":  p.URL,
		"missing":   missing,
		"assembled": assembled,
	}
	specJSON, _ := json.Marshal(spec)

	pageID := p.ID
	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   &pageID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "head_essentials_missing",
		Severity: "medium",
		Summary: fmt.Sprintf("Page %s is missing: %s",
			p.Name, strings.Join(missing, ", ")),
		SpecJSON:  string(specJSON),
		Priority:  60,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   structuralItemKey("head_essentials_missing", p.ID),
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the file header.
	}
}

// ===========================================================================
// 5. sitemap_entry_dead_live
// ===========================================================================
//
// The narrow, safe cousin of "every page appears in the sitemap" that the file
// header's WHAT IS DELIBERATELY NOT GATED section names and explicitly defers.
// The direction is the OPPOSITE of completeness: this never asks whether a
// page is missing from the sitemap, only whether a URL the sitemap DOES list
// still resolves. A site with no sitemap.xml (most of them — generation is
// scripts/site-discovery-files.py, a manual script, not a standing mechanism)
// produces zero findings, silently — see loadStructuralDomain/parseSitemapEntries
// below for the exact silent-skip conditions.
//
// UNLIKE the four page-scoped checks above, this one has no per-page
// population of its own: its population is "whatever this site's own
// /sitemap.xml currently claims", fetched once per site, not once per page —
// see the file header's KNOWN, STATED COST section for the cost-shape
// distinction. It reuses loadStructuralDomain (not loadStructuralPopulation:
// no pages query is needed), fetchStructuralPage for the sitemap.xml GET
// itself, and probeInternalLinkTargets — UNCHANGED — for probing every entry,
// so the confirm-before-file / 404|410-only discipline (rules (a) and (b) in
// the file header) is the same code, not a re-implementation of it.
//
// Carries no page_id: a sitemap entry is the site's OWN stated URL, not a link
// found ON a particular page the way dead_internal_link_live's targets are,
// and mapping a <loc> back to a `pages` row would need a second DB query this
// check does not otherwise require — left out on purpose, matching
// undeployed_asset's "site-scoped, no per-item target id" shape in
// verifier_coverage_test.go's itemTypesWithoutVerifiers map.
type SitemapEntryDeadCheck struct{}

func (c *SitemapEntryDeadCheck) Name() string { return "sitemap_entry_dead_live" }

// sitemapURLSetXML is the sitemaps.org protocol shape — the exact shape
// scripts/site-discovery-files.py's sitemap_xml() emits
// (`<urlset xmlns="...">   <url><loc>...</loc><lastmod>...</lastmod></url>`).
// Only <loc> is read; <lastmod> and any other sitemap extension element are
// not this check's concern.
type sitemapURLSetXML struct {
	XMLName xml.Name `xml:"urlset"`
	URLs    []struct {
		Loc string `xml:"loc"`
	} `xml:"url"`
}

// parseSitemapEntries parses a sitemap.xml body and returns every non-empty
// <loc> value. ok is false when the body does not parse as XML at all — the
// caller treats that identically to "no sitemap.xml found": a malformed file
// is not evidence any URL is dead, it is evidence this check has nothing to
// assert (WHAT IS DELIBERATELY NOT GATED, above — validating the sitemap's
// own well-formedness is a different, unbuilt check).
func parseSitemapEntries(body string) (locs []string, ok bool) {
	var doc sitemapURLSetXML
	if err := xml.Unmarshal([]byte(body), &doc); err != nil {
		return nil, false
	}
	for _, u := range doc.URLs {
		loc := strings.TrimSpace(u.Loc)
		if loc != "" {
			locs = append(locs, loc)
		}
	}
	return locs, true
}

func (c *SitemapEntryDeadCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	domain, err := loadStructuralDomain(dctx)
	if err != nil {
		return nil, err
	}
	if domain == "" {
		return result, nil
	}

	// Fetch sitemap.xml itself with the same confirm-before-conclude treatment
	// fetchAllPagesLive gives the OUTER page fetch (file header, THE PROBE
	// RULES): a single transient blip on the sitemap fetch must not be read as
	// "this site has no sitemap".
	sitemapURL := "https://" + domain + "/sitemap.xml"
	status, body, ferr := fetchStructuralPage(dctx.Ctx, sitemapURL)
	if ferr != nil || status < 200 || status > 299 {
		time.Sleep(structuralRetryWait)
		status, body, ferr = fetchStructuralPage(dctx.Ctx, sitemapURL)
	}
	if ferr != nil || status < 200 || status > 299 {
		// No sitemap.xml (confirmed), or unreachable. Per the file header's own
		// rule: this is "not listed", not "broken". Skip entirely, silently.
		return result, nil
	}

	entries, ok := parseSitemapEntries(body)
	if !ok || len(entries) == 0 {
		// Present but not a valid/populated sitemap. Same silent skip —
		// asserting anything about sitemap.xml's OWN well-formedness is a
		// different, unbuilt check (see parseSitemapEntries's own doc comment).
		return result, nil
	}

	seen := map[string]bool{}
	var urls []string
	for _, loc := range entries {
		u, perr := url.Parse(loc)
		if perr != nil || !u.IsAbs() {
			continue // not a usable absolute URL; nothing to probe
		}
		if !sameHost(u.Host, domain) {
			// A sitemap listing another site's URL is not this site's remit —
			// same reasoning canonical_mismatch's off_domain branch uses.
			continue
		}
		if seen[loc] {
			continue
		}
		seen[loc] = true
		urls = append(urls, loc)
	}
	sort.Strings(urls)

	if len(urls) > maxSitemapProbeURLs {
		dropped := urls[maxSitemapProbeURLs:]
		dctx.Logger.Warn("sitemap_entry_dead_live: probe cap reached — these sitemap entries were NOT checked",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", maxSitemapProbeURLs),
			zap.Int("dropped", len(dropped)))
		urls = urls[:maxSitemapProbeURLs]
	}

	outcomes := probeInternalLinkTargets(dctx, urls)
	for _, u := range urls {
		o := outcomes[u]
		switch {
		case o.err != nil:
			// A transport failure is not a status. Skip, never a finding.
			continue

		case o.code == http.StatusNotFound || o.code == http.StatusGone:
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":       "sitemap_entry_dead_live",
				"url":         u,
				"http_status": o.code,
			})
			result.WorkItems = append(result.WorkItems, buildSitemapEntryWorkItem(dctx, u, o.code))

		case o.code >= 200 && o.code < 400:
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "sitemap_entry_dead_live",
				ItemKey:  "sitemap_entry_dead_live:" + u,
				Reason:   fmt.Sprintf("re-probed %s: now returns HTTP %d", u, o.code),
			})

		default:
			// 401/403/429/5xx and anything else. Rule (a): never a finding.
			continue
		}
	}
	return result, nil
}

func buildSitemapEntryWorkItem(dctx DiscoveryCheckContext, u string, status int) WorkItemSpec {
	spec := map[string]interface{}{
		"check":       "sitemap_entry_dead_live",
		"url":         u,
		"http_status": status,
	}
	specJSON, _ := json.Marshal(spec)

	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "sitemap_entry_dead_live",
		// An SEO correctness defect surfaced to CRAWLERS via a sitemap entry,
		// not a user-facing link a visitor would actually click — the same
		// grading canonical_mismatch uses for the same reason (bugs_open/251's
		// own framing: "both URLs still serve the page" doesn't apply here
		// verbatim, but the page is very often still reachable some other way;
		// this misdirects the crawl rather than breaking a visit outright).
		Severity:  "medium",
		Summary:   fmt.Sprintf("Sitemap lists %s, which returns HTTP %d", u, status),
		SpecJSON:  string(specJSON),
		Priority:  55,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   "sitemap_entry_dead_live:" + u,
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the file header.
	}
}
