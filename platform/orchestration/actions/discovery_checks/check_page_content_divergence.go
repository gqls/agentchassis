// FILE: platform/orchestration/actions/discovery_checks/check_page_content_divergence.go
//
// Discovery check: page_content_divergence — "is the origin serving the bytes we
// last sent for this page?"
//
// bugs_open/315 fix candidate 4, and PLAN_2026-08-19 decision D5.
//
// ── THE DEFECT ─────────────────────────────────────────────────────────────
//
// Delivery on this estate is asynchronous and batched: a handler commits a page's
// file to the site repo ("commit is deploy") and a GitHub Actions runner in the
// private gqls/sites repo `b2 sync`s the whole changed directory minutes later.
// Nothing on this side of that boundary observes the sync. So a page can commit
// successfully, stamp `deployed_at`, report a COMPLETED orchestration, and never
// reach the origin — which is exactly what bugs_open/315 recorded: a tool page
// carried on serving its OLD render for about six hours, through four completed
// rerenders, with every internal signal green.
//
// ── WHY THIS CHECK CAN EXIST TODAY AND COULD NOT ON 2026-08-19 ─────────────
//
// The deep finding of bugs_open/315 was that the platform could not tell "this
// page never needed republishing" from "this page failed to republish" — the two
// were identical in every signal it produced. `pages.content_hash` (RFC_038,
// register DGH-013) is what separated them: it is the sha256 of the bytes the
// deploy step actually committed, taken from the git-adapter's own reply. Until
// it existed there was nothing to compare the wire against.
//
// ── WHY NOT TIMESTAMPS. THIS IS THE HARD CONSTRAINT, AND IT IS MEASURED ────
//
// bugs_open/315's own candidate 4 proposed comparing `deployed_at` against the
// origin's `last-modified`. That does not work, and the refutation is a
// measurement rather than an argument: run across 40 live pages on 2026-08-19 it
// returned 40 of 40 "stale", every one of them HEALTHY, and the apparent
// staleness persisted 85 minutes — past any settle window one would configure.
// The cause is structural: a byte-identical rerender legitimately rewrites
// nothing, so the origin's mtime stays old for ever and says nothing about
// whether the page is current. Only a content hash separates the two cases.
//
// A settle window is therefore NOT the mechanism that makes this check sound —
// the hash is. The window only keeps the check off pages whose delivery is still
// in flight.
//
// ── WHAT IT FINDS TODAY: NOTHING, AND THAT IS RECORDED ON PURPOSE ──────────
//
// [MEASURED 2026-08-21 ~10:35Z] Every one of the 228 active pages then carrying a
// content_hash was fetched with a cache-buster and hashed: 228 of 228 MATCH,
// across 12 domains (webdesign.co.uk 124, ai-agent-orchestration.com 36,
// robot-hands.com 31, loanandmortgagecalculator.co.uk 21, dartsonline.com 6, and
// 7 sites with 1–2 each). So this check is a REGRESSION GUARD, not a repair of
// live damage — the same posture, and for the same reason, as
// check_asset_reference_404 (96 of 96 assets serving 200 when it was written).
//
// That measurement is also the end-to-end proof that the comparison is sound: the
// stored fingerprint and the served bytes agree on 228 independent pages, so the
// hash the stamp writes and the bytes the origin serves are the same thing —
// no encoding, no transform, no path-keying error anywhere between them.
//
// A guard with no live positive can rot unexercised, so every branch below is
// proven by an induced fault in check_page_content_divergence_test.go rather than
// by hope.
//
// ── THE FIVE THINGS THAT COULD MAKE THIS CHECK LIE, AND WHAT STOPS EACH ────
//
//  1. DELIVERY STILL IN FLIGHT. A page stamped seconds ago has not reached the
//     origin yet. → `divergenceSettleWindow`, and the finder will not look at a
//     page younger than it. See that const for what the window is measured
//     against and what is honestly unmeasured about it.
//
//  2. A CDN EDGE SERVING AN OLD COPY. → every request carries a unique
//     cache-buster query, and only a 200 is judged. The 228-page sweep above used
//     exactly this method fleet-wide, so it is measured to work rather than
//     assumed to.
//
//  3. THE PAGE WAS REDEPLOYED WHILE WE WERE PROBING IT. We read hash H1, a deploy
//     writes H2 and new bytes, we fetch the new bytes and convict H1. → after a
//     mismatch, `contentIntentUnchanged` re-reads content_hash and deployed_at and
//     DISCARDS the candidate if either moved. A race that files a work item
//     against a healthy page is worse than a missed pass; the population is
//     stable and the next pass will see it.
//
//  4. THE ORIGIN MID-WRITE. A sync in progress can serve one body then another.
//     → a mismatch is confirmed by a SECOND fetch, and the two served hashes must
//     AGREE before anything is filed. Two different bodies is not a divergence
//     finding, it is a moving target, and it is logged as a skip.
//
//  5. A STALE FINGERPRINT WRITTEN BY NOBODY. The stamp assigns content_hash only
//     when the deploy-evidence guard RAN (v3_site_actions.go, and the reasoning is
//     in the comment there). An UNARMED step that stamps `deployed` leaves the
//     column alone — so if an unarmed step ever deployed new bytes, the hash would
//     describe an older deploy and this check would convict a healthy page.
//
//     > **⚠ CORRECTED 2026-08-21, BEFORE THIS CHECK WAS EVER ENABLED — the
//     > paragraph that stood here was FALSE, and the council gate's `guardian`
//     > seat caught it (round 1, corr `be85a6d3`).** It said: *"That cannot
//     > happen today, and it is a query, not an argument: exactly THREE live
//     > steps set status='deployed' via update_page_status ... and ALL THREE
//     > declare deploy_result_field. Zero unarmed stampers."*
//     >
//     > **There are SIX, and THREE of them are UNARMED.** The census behind the
//     > old claim walked `default_config.<workflow>.steps.*` — one level — and
//     > the three it missed live at
//     > `workflow.steps.<loop>.config.sub_workflow.steps.update_page_status`,
//     > inside a loop's sub-workflow. The seat predicted exactly that
//     > ("almost certainly measured with a top-level workflow.steps census —
//     > documented elsewhere on this platform as blind to actions nested inside
//     > sub_workflow/substeps") without seeing the query. It was right.
//
//     [MEASURED 2026-08-21, with a RECURSIVE walk — `jsonb_path_query(...,
//     '$.**{0 to 25} ? (@.action == "update_page_status")')`]:
//
//	armed    page-rerender/update_status            deploy_result
//	armed    report-builder/update_status           deploy_result
//	armed    section-editor/update_page_status      git_result
//	UNARMED  page-rebuild            .steps.build_pages_loop.config.sub_workflow
//	UNARMED  pageflow-builder        .steps.build_pages_loop.config.sub_workflow
//	UNARMED  site-work-orchestrator  .steps.build_items_loop.config.sub_workflow
//
//     The three unarmed ones are the PAGE-BUILDING paths — precisely the paths
//     that emit new bytes — so this is not a remote possibility, it is the main
//     road. **Case 5 is LIVE.**
//
//     WHY NO PAGE IS ACTUALLY POISONED TODAY, and why that is luck rather than
//     safety: a stale fingerprint shows up as exactly the mismatch this check
//     looks for, and the 228-page sweep found 228 MATCH — so as of 2026-08-21
//     10:35Z no page was in that state. That is an observation about one moment,
//     not a property of the system.
//
//     ⚠⚠ THEREFORE THIS CHECK MUST NOT BE ENABLED WHILE ANY UNARMED
//     `deployed` STAMPER EXISTS. That is not left to whoever applies the
//     config: migration 526 REFUSES to apply while the recursive count above is
//     non-zero, so the precondition is enforced rather than remembered. The
//     preferred fix is to ARM the three (all three carry a `git_commit` step
//     `deploy_page` with `output_field: "page_deployed"`, so it is one migration
//     in 494's shape) — arming raises fingerprint coverage, whereas PLAN D6's
//     stamp-side NULLing lowers it and is the backstop for the NEXT unarmed
//     stamper rather than the answer to these three. See PLAN D6/D7.
//
// ── THE OTHER PUBLISH SEAM, AND WHY THIS CHECK STAYS OFF IT ────────────────
//
// Raised by the council gate's `reuse_agent` seat (corr be85a6d3): `sites.published_hash`
// and migration 422's site-publish reconciler occupy the same conceptual space —
// fingerprint-based publish verification — and the submission never mentioned
// them. Checked afterwards, and they are a DIFFERENT mechanism on a DIFFERENT
// seam rather than a duplicate of this one:
//
//   - 422 drives `publish_site`: a DIRECT B2 upload from a spawned site-publisher
//     pod holding B2 credentials. This check observes commit-is-deploy (git →
//     Actions → B2 sync). Different delivery path end to end.
//   - `sites.published_hash` is SITE-level and is not a page's bytes — the one
//     live value is `th1:05a06351`, a prefixed TREE digest. `pages.content_hash`
//     is a per-page sha256 of the file. The two are not comparable quantities.
//   - Populations barely touch: 422 fires only for sites with `publish_target`
//     set — [MEASURED 2026-08-21] **1 of 45** (noted.co.uk) — while this check's
//     population is 228 pages across 12 domains, none of them that site's.
//
// ⚠ BUT THE SEAT'S INSTINCT FOUND A REAL HAZARD, which is why the predicate now
// carries `s.publish_target IS NULL`. `publish_site_action.go` writes NEITHER
// `content_hash` NOR `deployed_at` (grep it). So a site that already had hashed
// pages and then opted into `publish_target` would keep fingerprints that the new
// seam never updates — a stale fingerprint reached by a different door, and this
// check would convict healthy pages. [MEASURED 2026-08-21] the one opted-in site
// has 12 active pages and 0 hashed, so there is no exposure today; the predicate
// makes it structurally impossible rather than currently absent. On a seam we
// cannot observe, silence is the correct output.
//
// ── WHY THE CANDIDATE PREDICATE LOOKS HAND-ROLLED AND IS NOT ───────────────
//
// Raised by the `debug_historian` seat in the same round: several landmines warn
// that a hand-rolled liveness filter is wrong for an AUDIT, and that the shared
// shipped-predicate should be reused. Half-right, and acted on:
//
//   - the "did this page ship" leg IS the shared one now —
//     `queryresolve.DeployedPageEligibilitySQL`, concatenated rather than
//     re-typed, honouring its `p`-for-pages alias contract;
//   - `status = 'active'` is NOT a liveness filter and is not in any shared
//     predicate: it excludes RETRACTED and ARCHIVED pages, which keep
//     `deployed_at` by design (D5) and are deliberately no longer served. Judging
//     them would report every retraction as a divergence.
//
// The enumeration that seat asked for, which was owed and had not been run
// [MEASURED 2026-08-21]:
//
//	status   | build_status  |   n | hashed
//	active   | deployed      | 651 |    232
//	active   | needs_rebuild |  56 |      0
//	active   | planned       |  42 |      0
//	archived | (all three)   |  69 |      0
//
// Every hashed page is `active` + `deployed`; no archived page carries a hash;
// and `status='deployed'` does not occur at all, which is its own landmine. So
// `status='active'` is not merely defensible here, it is currently a no-op that
// costs nothing and closes the retraction case the moment one appears.
//
// ── WHAT THIS CHECK DOES NOT OWN ───────────────────────────────────────────
//
// The landmine keyed to check_image_url_404.go warns that widening one asset/serving
// check silently competes with another. Each neighbour was read before this was
// written:
//
//   - "does the site answer at all" → check_site_unreachable. It probes the apex
//     for every deployed site and files on availability. It EXPLICITLY declines
//     the staleness class — its header records mortgagecalculator "serves a
//     divergent render today — a staleness defect, not an availability one" and
//     files nothing for it. That declined class is precisely this check's remit,
//     so the two do not overlap: a non-200 here is a skip, never a finding.
//
//     [MEASURED 2026-08-21] That skip is not theoretical tidiness either. In the
//     same 2h42m watch, two readings came back 404 — webdesign.co.uk/index.html at
//     age 782s and /tools/noise-generator/index.html at age 850s — both serving
//     the SAME body (an edge error page, sha256 e3ebaa16…), both surrounded by
//     MATCH readings before and after. The origin answers an intermittent 404.
//     Judged as content, those are two more false work items against healthy
//     pages; skipped as an unjudgeable status, they are correctly silent.
//   - "does a subresource this page references exist" → check_asset_reference_404
//     (a wire question about a URL this page names). This check asks about the
//     page's OWN bytes.
//   - "is the page's DB state self-consistent" → check_page_component_status_drift,
//     check_section_source_drift. Both are DB-only; neither can see the origin.
//   - "was an asset row never deployed" → check_undeployed_assets.
//
// ── SCOPE, AND THE POPULATION IT CANNOT SEE ────────────────────────────────
//
// `content_hash IS NOT NULL` makes the check structurally inert on any page whose
// last deploy predates the fingerprint. [MEASURED 2026-08-21] that is 588 of 816
// pages — the check can say nothing about them, and it says nothing rather than
// guessing. The population grows as pages redeploy, which is the only honest way
// in: a hash can only be written by a deploy that reports what it sent.
//
// `status = 'active'` excludes retracted and archived pages, which keep
// `deployed_at` by design (bugs_open/315 §candidate 4) and whose files are
// deliberately no longer served.
//
// A url carrying a fragment or query is refused by the shared
// datahelpers.PageFilePathFromURL — the SAME definition that decided what to
// hash — so this check judges only pages whose url maps 1:1 to one file. idea.uk
// has a live page row at "/tools.html#audience-check" while a DIFFERENT page owns
// "/tools.html"; fetching the first would compare one page against the other's
// bytes. Such a page never receives a hash in the first place, so the predicate
// above already excludes it; the explicit refusal is here so that the reason is
// visible at the point of use rather than inferred from another file.
//
// ── ROUTING ────────────────────────────────────────────────────────────────
//
// Flag-only: HandlerAgent is empty, per D5 ("no handler agent in v1"). The obvious
// repair — re-file a rerender — is the loop that already failed four times in
// bugs_open/315's own instance, so the honest first move is visibility, not an
// automated retry that would paper over the delivery boundary. The cost of
// flag-only is real and named in bugs_open/083: "a detector whose output nobody
// drains is not neutral — it is actively misleading". It is accepted here for one
// release, and the item carries the stored and served hashes so triage is a look
// rather than an investigation.
//
// SELF-CLEARING via CheckResult.Resolved: a page whose hashes now agree is a
// POSITIVE observation, which is what that field's contract requires. Nothing
// here infers resolution from absence.
//
// Per the owner ruling of 2026-08-02 §1, an item type with no automated consumer
// is normal council-gate scope rather than RFC scope. Producer and key shape are
// registered in docs026_concept_register/register/deployment-github.md.
//
// Registration: automatic via init() -> Register(&PageContentDivergenceCheck{}).
// Enable: add "page_content_divergence" to a discovery agent's
//   default_config {workflow,steps,run_checks,config,checks} array — and NOT
//   before the image carrying this file has rolled, because the runner hard-fails
//   on a check name the binary does not register (check_site_unreachable's
//   migration 368 learned that).

package discovery_checks

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&PageContentDivergenceCheck{}) }

type PageContentDivergenceCheck struct{}

func (c *PageContentDivergenceCheck) Name() string { return "page_content_divergence" }

const (
	// divergenceSettleWindow is how long after `deployed_at` a page becomes
	// judgeable. It is NOT what makes the check sound — the hash is (see the
	// header). It only keeps the check off pages whose batched delivery is still
	// in flight.
	//
	// WHAT IT IS MEASURED AGAINST. Two measurements, and the SECOND ONE CORRECTS
	// THE FIRST BY TWO ORDERS OF MAGNITUDE. Both are kept, because the way the
	// first one misled is the useful part.
	//
	// [MEASURED 2026-08-21, 10:38Z–13:20Z] A watcher re-probed every page stamped
	// in the previous 45 minutes, every 2 minutes: 1,099 readings, 85 pages, 95
	// deploy events. The only 3 DIVERGED readings were at ages 1s, 13s and 14s, all
	// converged by 140–156s, and 0 of 995 readings at age >= 157s diverged. On that
	// evidence this comment claimed 30 minutes was "roughly 128x the largest lag
	// actually observed".
	//
	// > **⚠ CORRECTED 2026-08-21 19:36Z — THAT MARGIN WAS WRONG, and it was found
	// > by accident while re-running the proof after the check went live.** A random
	// > 40-page sample returned 2 DIVERGED, both on fundamentallyai.com, aged 15 and
	// > 21 MINUTES. Tracked to convergence:
	// >
	// >	/model-fine-tuning.html   MATCH @945s, DIVERGED @1012s, MATCH @1079s onward
	// >	/tools/automation-savings-estimator/index.html   MATCH @1293s onward
	// >
	// > **The largest observed divergence age is therefore ~1012s (~17 minutes), not
	// > 14 seconds.** So the window is about **1.8x** the worst observed case, not
	// > 128x. The first measurement was not wrong about what it saw; it was a
	// > 2h42m sample that happened to catch only fast deliveries, and quoting its
	// > maximum as "the tail" was the error — the same shape as reading a
	// > retention-bounded table as a lifetime.
	//
	// AND THE SHAPE IS NOT A SIMPLE LAG. `/model-fine-tuning.html` read MATCH, then
	// DIVERGED 67s LATER, then MATCH again — non-monotonic. Delivery lands
	// PROGRESSIVELY across edge nodes, so during the window a probe gets whichever
	// version the node that answered happens to hold. That is why the confirmation
	// fetch must AGREE with the first before anything is filed (guard 4 below):
	// two probes seconds apart can legitimately hit different nodes.
	//
	// THE WINDOW IS STILL LOAD-BEARING, and now doubly so: those 2 pages are 2 work
	// items this check would have filed against healthy pages in a single 40-page
	// sample, and 5 more were prevented in the earlier watch. But the margin is
	// THIN — 30 minutes against a 17-minute worst case.
	//
	// ⚠ RECOMMENDATION, NOT YET TAKEN (PLAN D8): widen this to 60 minutes at the
	// next build. The cost is that a real divergence stays invisible for its first
	// hour, which is still comfortable against the 6-hour case this check exists
	// for; the benefit is margin against a delivery batch slower than any yet
	// sampled. It is left at 30 for now because the failure mode is bounded and
	// self-clearing — a premature finding is FLAG-ONLY and is RETRACTED on the next
	// pass's positive re-observation — and because changing a const here costs a
	// rebuild and a fleet roll.
	divergenceSettleWindow = 30 * time.Minute

	// divergenceMaxPagesPerPass bounds the outbound fetches one site can cause in
	// one pass. webdesign.co.uk carries 124 hashed pages today, so this DOES bite
	// on the fleet's busiest site — and when it drops anything it LOGS what it
	// dropped, because a silent cap reads as "every page was checked".
	divergenceMaxPagesPerPass = 60

	// divergenceProbeWorkers keeps one slow origin from serialising a whole site.
	// Small on purpose: this runs inside a discovery sweep, not a load test.
	divergenceProbeWorkers = 4

	// divergenceProbeTimeout per request. A whole HTML document, so more generous
	// than check_asset_reference_404's 10s subresource probe and in line with
	// check_tool_acceptance's 12s page fetch.
	divergenceProbeTimeout = 20 * time.Second

	// divergenceMaxBodyBytes bounds what one probe will read. Unlike every other
	// body cap in this package, EXCEEDING IT MUST NOT PRODUCE A HASH: a sha256
	// over a truncated body is a different, confidently wrong answer, which would
	// convict a healthy page. So the reader takes cap+1 bytes and a body that
	// reaches that length is SKIPPED with a reason. The largest page measured on
	// this estate is well under 1MB.
	divergenceMaxBodyBytes = 8 << 20
)

// divergenceProbeResult is what one fetch observed. A transport failure is NOT a
// status and must never be compared against 200.
type divergenceProbeResult struct {
	hash   string
	status int
	bytes  int64
	err    error
	// oversize is set when the body reached divergenceMaxBodyBytes, in which case
	// hash is deliberately empty — see that const.
	oversize bool
}

// fetchServedPage is swappable in tests — the same seam probeAssetURL uses in
// check_asset_reference_404.go, and the reason the test file can prove every
// branch without a network.
//
// It returns the sha256 of the WHOLE body, which is the only thing that can be
// compared against pages.content_hash.
var fetchServedPage = func(ctx context.Context, target string) divergenceProbeResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return divergenceProbeResult{err: err}
	}
	// GET, not HEAD: the body IS the question. An explicit, honest User-Agent —
	// if an origin refuses it that is a 403, and a 403 files nothing.
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+page_content_divergence)")
	req.Header.Set("Accept", "text/html,*/*")
	// DELIBERATELY NO Accept-Encoding. Go's transport adds gzip itself and then
	// transparently decompresses, so the body we hash is the file's own bytes.
	// Setting the header by hand DISABLES that transparent decode and we would
	// hash the compressed stream instead — a wrong answer on every gzip-serving
	// origin, which is all of them.

	client := &http.Client{Timeout: divergenceProbeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return divergenceProbeResult{err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Not judgeable. Drain a token amount so the connection can be reused.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		return divergenceProbeResult{status: resp.StatusCode}
	}

	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(resp.Body, divergenceMaxBodyBytes+1))
	if err != nil {
		return divergenceProbeResult{status: resp.StatusCode, err: err}
	}
	if n > divergenceMaxBodyBytes {
		return divergenceProbeResult{status: resp.StatusCode, bytes: n, oversize: true}
	}
	return divergenceProbeResult{
		hash:   hex.EncodeToString(h.Sum(nil)),
		status: resp.StatusCode,
		bytes:  n,
	}
}

// divergencePage is one judgeable page: it has shipped, it reported what it sent,
// and its delivery window has passed.
type divergencePage struct {
	PageID      string
	Name        string
	URL         string
	StoredHash  string
	BuildStatus string
	DeployedAt  time.Time
	AgeSeconds  int64
}

// divergenceSkips records why a page was not judged. A check that reports only
// findings cannot be told apart from one that was blinded.
type divergenceSkips struct {
	unfetchableURL    int // url does not map 1:1 to a file (fragment/query/absolute)
	overCap           int
	transportError    int
	notOK             int // any status that is not 200 — availability is not our remit
	oversizeBody      int
	movingTarget      int // two fetches, two different bodies
	redeployedMidPass int // the page's own intent changed while we looked
}

func (c *PageContentDivergenceCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, '') FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("page_content_divergence: site lookup failed: %w", err)
	}
	if domain == "" {
		// No domain, no URL to fetch. Nothing this check can say.
		return result, nil
	}

	pages, err := findDivergenceCandidates(dctx)
	if err != nil {
		return nil, err
	}
	if len(pages) == 0 {
		return result, nil
	}

	var skips divergenceSkips

	// Resolve each page to the one URL it is served at, dropping any whose url
	// does not name a single file. Done BEFORE the cap so the cap counts pages
	// that were actually going to be fetched.
	type target struct {
		page divergencePage
		url  string
	}
	targets := make([]target, 0, len(pages))
	for _, pg := range pages {
		if _, ok := datahelpers.PageFilePathFromURL(pg.URL); !ok {
			skips.unfetchableURL++
			dctx.Logger.Info("page_content_divergence: page url does not name a single file; not judged",
				zap.String("page_id", pg.PageID), zap.String("url", pg.URL))
			continue
		}
		targets = append(targets, target{page: pg, url: "https://" + domain + pg.URL})
	}

	if len(targets) > divergenceMaxPagesPerPass {
		dropped := targets[divergenceMaxPagesPerPass:]
		skips.overCap = len(dropped)
		names := make([]string, 0, len(dropped))
		for _, t := range dropped {
			names = append(names, t.page.URL)
		}
		dctx.Logger.Warn("page_content_divergence: per-pass cap reached — these pages were NOT checked",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", divergenceMaxPagesPerPass),
			zap.Int("dropped", len(dropped)),
			zap.Strings("dropped_urls", names))
		targets = targets[:divergenceMaxPagesPerPass]
	}

	// Probe concurrently; judge in the finder's deterministic order so two runs
	// over the same data produce the same sequence of items.
	type judged struct {
		first  divergenceProbeResult
		second divergenceProbeResult // only populated for a candidate mismatch
	}
	results := make(map[string]judged, len(targets))
	var mu sync.Mutex
	var wg sync.WaitGroup
	work := make(chan target)

	workers := divergenceProbeWorkers
	if len(targets) < workers {
		workers = len(targets)
	}
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range work {
				j := judged{first: fetchServedPage(dctx.Ctx, cacheBust(t.url))}
				// Confirm a candidate mismatch with a second fetch. Candidates are
				// rare by construction, so the extra call costs nothing normally.
				if j.first.err == nil && j.first.status == http.StatusOK &&
					!j.first.oversize && j.first.hash != t.page.StoredHash {
					j.second = fetchServedPage(dctx.Ctx, cacheBust(t.url))
				}
				mu.Lock()
				results[t.page.PageID] = j
				mu.Unlock()
			}
		}()
	}
	for _, t := range targets {
		work <- t
	}
	close(work)
	wg.Wait()

	for _, t := range targets {
		pg := t.page
		j, ok := results[pg.PageID]
		if !ok {
			continue
		}

		switch {
		case j.first.err != nil:
			skips.transportError++
			dctx.Logger.Info("page_content_divergence: fetch failed, not a finding",
				zap.String("url", t.url), zap.Error(j.first.err))

		case j.first.status != http.StatusOK:
			// Availability belongs to check_site_unreachable, which files on it.
			// Judging it here would double-file one fault as two defects.
			skips.notOK++
			dctx.Logger.Info("page_content_divergence: non-200, not judged",
				zap.String("url", t.url), zap.Int("status", j.first.status))

		case j.first.oversize:
			skips.oversizeBody++
			dctx.Logger.Warn("page_content_divergence: body over cap, cannot hash without truncating; not judged",
				zap.String("url", t.url), zap.Int64("bytes", j.first.bytes),
				zap.Int("cap", divergenceMaxBodyBytes))

		case j.first.hash == pg.StoredHash:
			// A POSITIVE observation, and the only thing that may retract.
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "page_content_divergence",
				ItemKey:  divergenceItemKey(pg.PageID),
				Reason: fmt.Sprintf("origin now serves the bytes we sent for %s (sha256 %s)",
					pg.URL, shortHash(pg.StoredHash)),
			})

		default:
			// A candidate divergence. Three ways it still does not get filed.
			if j.second.err != nil || j.second.status != http.StatusOK || j.second.oversize {
				skips.transportError++
				dctx.Logger.Info("page_content_divergence: confirmation fetch inconclusive, discarding candidate",
					zap.String("url", t.url), zap.Int("status", j.second.status), zap.Error(j.second.err))
				continue
			}
			if j.second.hash != j.first.hash {
				// The origin is mid-write. Two different bodies is a moving
				// target, not a divergence.
				skips.movingTarget++
				dctx.Logger.Info("page_content_divergence: origin served two different bodies, discarding candidate",
					zap.String("url", t.url),
					zap.String("first", shortHash(j.first.hash)),
					zap.String("second", shortHash(j.second.hash)))
				continue
			}
			current, err := contentIntentUnchanged(dctx, pg)
			if err != nil {
				dctx.Logger.Warn("page_content_divergence: could not re-read page intent, discarding candidate",
					zap.String("page_id", pg.PageID), zap.Error(err))
				continue
			}
			if !current {
				// It was redeployed while we were probing: we would be convicting
				// the page against a superseded intent.
				skips.redeployedMidPass++
				dctx.Logger.Info("page_content_divergence: page redeployed during the pass, discarding candidate",
					zap.String("page_id", pg.PageID), zap.String("url", pg.URL))
				continue
			}

			result.Findings = append(result.Findings, map[string]interface{}{
				"check":       "page_content_divergence",
				"page_url":    pg.URL,
				"stored_hash": pg.StoredHash,
				"served_hash": j.first.hash,
				"age_seconds": pg.AgeSeconds,
			})
			result.WorkItems = append(result.WorkItems, buildDivergenceWorkItem(dctx, pg, j.first))
		}
	}

	logDivergenceSkips(dctx, skips, len(pages))
	return result, nil
}

// cacheBust appends a unique query so no edge or proxy can answer from cache.
// The stored url never carries a query (PageFilePathFromURL refuses one), so
// this cannot collide with a real parameter.
//
// crypto/rand rather than math/rand: two checks running in the same second in
// the same process must not produce the same buster, and a seeded PRNG in a
// long-lived pod is exactly how that happens.
func cacheBust(target string) string {
	n, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		// A failure here is not a reason to skip the page; it is a reason to use
		// a less unique buster. Time is still monotonic enough to defeat a cache.
		return fmt.Sprintf("%s?cb=%d", target, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s?cb=%d", target, n.Int64())
}

func shortHash(h string) string {
	if len(h) <= 12 {
		return h
	}
	return h[:12]
}

// divergenceItemKey is the dedup key: one open item per page, so a second pass
// re-files nothing while the first is still open. The page id rather than the
// url — a url can be edited while the defect persists, and idx_swi_dedup would
// then hold two rows for one page.
func divergenceItemKey(pageID string) string {
	return "page_content_divergence:" + pageID
}

func buildDivergenceWorkItem(dctx DiscoveryCheckContext, pg divergencePage, probe divergenceProbeResult) WorkItemSpec {
	spec := map[string]interface{}{
		"check": "page_content_divergence",
		// Both hashes IN FULL: this item is triaged by comparing them, and a
		// truncated hash makes a reader go and re-derive what we already knew.
		"stored_hash":           pg.StoredHash,
		"served_hash":           probe.hash,
		"page_url":              pg.URL,
		"page_name":             pg.Name,
		"build_status":          pg.BuildStatus,
		"deployed_at":           pg.DeployedAt.UTC().Format(time.RFC3339),
		"age_seconds":           pg.AgeSeconds,
		"served_bytes":          probe.bytes,
		"settle_window_seconds": int(divergenceSettleWindow.Seconds()),
		"reason": "the origin is serving bytes that are not the ones this page's " +
			"last deploy committed; the commit succeeded and the delivery did not",
		// Named so a reader of the item knows where the boundary is without
		// finding the bug file first.
		"delivery_boundary": "commit is deploy: a gqls/sites runner b2-syncs the changed directory; nothing on this side observes that sync",
	}
	specJSON, _ := json.Marshal(spec)

	var pageIDPtr *uuid.UUID
	if parsed, perr := uuid.Parse(pg.PageID); perr == nil {
		pageIDPtr = &parsed
	}

	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   pageIDPtr,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "page_content_divergence",
		// A visitor is being served content this platform believes it replaced.
		Severity: "high",
		Summary: fmt.Sprintf(
			"Page %s is serving bytes that are not the ones we deployed %s ago (stored %s, served %s)",
			pg.URL, time.Duration(pg.AgeSeconds)*time.Second, shortHash(pg.StoredHash), shortHash(probe.hash)),
		SpecJSON:  string(specJSON),
		Priority:  40,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   divergenceItemKey(pg.PageID),
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the header.
	}
}

// divergenceCandidatesQuery is a package-level const rather than an inline string
// so check_page_content_divergence_test.go can assert the guards are still in it.
// Each is a one-line deletion that leaves every behavioural test green — sqlmock
// returns the rows a test hands it whatever the WHERE clause says — so the query
// text is the only place they can be pinned.
var divergenceCandidatesQuery = `
		SELECT p.id::text, p.name, p.url, p.content_hash,
		       COALESCE(p.build_status, ''), p.deployed_at,
		       round(extract(epoch FROM (now() - p.deployed_at)))::bigint
		  FROM pages p
		  JOIN sites s ON s.id = p.site_id
		 WHERE p.site_id = $1
		   -- retracted and archived pages keep deployed_at by design and are
		   -- deliberately no longer served
		   AND p.status = 'active'
		   -- the page reported what it sent; without this there is nothing to
		   -- compare the wire against and the check is structurally inert
		   AND p.content_hash IS NOT NULL
		   -- a site publishing through the publish_site seam has no fingerprint
		   -- authority here — see the header. Silent beats wrong.
		   AND s.publish_target IS NULL` +
	// "did this page ship" is the platform's line, not this check's to redraw.
	// The shared constant's alias contract is `p` for pages, which this query honours.
	queryresolve.DeployedPageEligibilitySQL + `
		   -- batched delivery may still be in flight inside the settle window
		   AND p.deployed_at < now() - make_interval(secs => $2)
		 ORDER BY p.deployed_at DESC, p.url
	`

func findDivergenceCandidates(dctx DiscoveryCheckContext) ([]divergencePage, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, divergenceCandidatesQuery,
		dctx.SiteID, divergenceSettleWindow.Seconds())
	if err != nil {
		return nil, fmt.Errorf("page_content_divergence query failed: %w", err)
	}
	defer rows.Close()

	var out []divergencePage
	for rows.Next() {
		var p divergencePage
		if err := rows.Scan(&p.PageID, &p.Name, &p.URL, &p.StoredHash,
			&p.BuildStatus, &p.DeployedAt, &p.AgeSeconds); err != nil {
			dctx.Logger.Warn("page_content_divergence: scan failed", zap.Error(err))
			continue
		}
		if strings.TrimSpace(p.StoredHash) == "" {
			// Defensive: an empty string is not a fingerprint. The column is
			// written as NULL when unknown, but a caller could yet write "".
			continue
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("page_content_divergence rows iter failed: %w", err)
	}
	return out, nil
}

// contentIntentUnchanged re-reads the page's fingerprint and stamp and reports
// whether BOTH are still what the finder saw.
//
// This is the race guard, and it is the difference between a check that can be
// trusted and one that occasionally libels a healthy page. A deploy landing
// between the finder's SELECT and the probe's response replaces both the bytes on
// the origin and the hash in the row; comparing the new bytes against the old
// hash is guaranteed to mismatch.
func contentIntentUnchanged(dctx DiscoveryCheckContext, pg divergencePage) (bool, error) {
	var hash string
	var deployedAt time.Time
	err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(content_hash, ''), COALESCE(deployed_at, to_timestamp(0)) FROM pages WHERE id = $1`,
		pg.PageID).Scan(&hash, &deployedAt)
	if err != nil {
		return false, err
	}
	return hash == pg.StoredHash && deployedAt.Equal(pg.DeployedAt), nil
}

func logDivergenceSkips(dctx DiscoveryCheckContext, s divergenceSkips, candidates int) {
	if s.unfetchableURL == 0 && s.overCap == 0 && s.transportError == 0 &&
		s.notOK == 0 && s.oversizeBody == 0 && s.movingTarget == 0 && s.redeployedMidPass == 0 {
		return
	}
	dctx.Logger.Info("page_content_divergence: pages not judged",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("candidates", candidates),
		zap.Int("url_not_a_file", s.unfetchableURL),
		zap.Int("over_cap", s.overCap),
		zap.Int("transport_error", s.transportError),
		zap.Int("non_200", s.notOK),
		zap.Int("oversize_body", s.oversizeBody),
		zap.Int("moving_target", s.movingTarget),
		zap.Int("redeployed_mid_pass", s.redeployedMidPass))
}
