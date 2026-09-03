// FILE: platform/orchestration/actions/discovery_checks/check_stylesheet_gutted.go
//
// Discovery check: stylesheet_gutted — "does the stylesheet this site serves
// still DEFINE the custom properties its deployed pages consume?"
//
// A site can serve a stylesheet that returns 200, is well-formed CSS, and
// defines nothing. Every other signal says the site is fine: the pages render,
// the orchestrations completed, build_status is 'deployed', and the file is
// right there at the URL the pages ask for. What the visitor sees is unstyled
// markup, or worse — invisible text, because an undefined var() with no
// fallback makes the whole declaration invalid at computed-value time, so
// `color: var(--color-text)` falls back to INHERIT (usually black) while the
// background it was chosen against falls back to transparent.
//
// ── WHY THIS EXISTS: it has happened three times, and nothing noticed ───────
//
// bugs_open/198. css-patch-agent loads a base stylesheet from
// css_themes.css_content, appends its patch to it server-side, and then deploys
// THE WHOLE DB ROW over assets/css/styles.css. On a site composed by the normal
// path that column is the empty string from birth — install_site_composition
// writes an empty string there deliberately, because the renderer reads its
// composition via FKs and never consults that column — so
// the append produces a file containing only the patches, and the deploy
// replaces a 13–25KB stylesheet with a few hundred bytes.
//
//	relojistas.com   2026-08-04   26,152 B row → 149 B served
//	six sites        2026-08-17/18 idea.uk 23,650 B → 428 B; oufe.com nine
//	                              successive clobber commits in one day
//	remortgagecalculator.uk 2026-08-21  17,403 B → 136 B, and loanzy.uk the
//	                              same hour, 17,160 B → 1,577 B and climbing
//
// Every one of those was found by a human looking at a page. The last two were
// found while this check was being written.
//
// ── WHY NO EXISTING CHECK FIRES, walked one by one ─────────────────────────
//
// This was measured before the check was written, not assumed:
//
//   - check_asset_reference_404 is the ONLY check that fetches this URL, and it
//     scores HTTP STATUS ONLY. A 136-byte 200 is a positive observation to it,
//     so it does not merely stay silent — it RETRACTS any open item keyed to
//     that URL. That is why the item key here is deliberately NOT URL-shaped
//     (see stylesheetGuttedItemKey): sharing a key namespace with a check that
//     closes on 2xx would have this check's findings closed by its sibling on
//     the next sweep.
//   - check_missing_css asks whether the css_themes ROW exists. During a clobber
//     it does — it is the row that did the damage.
//   - check_generic_theme reads the head site_component's rendered_html, which
//     is intact and still carries the <link> to the gutted file.
//   - check_palette_contrast reads palettes.colours — the palette row is
//     untouched; this is exactly the intent-versus-artefact gap its own header
//     names.
//   - THE RENDER AUDIT IS ACTIVELY BLIND HERE, which is the worst of it. With
//     :root gone, text computes to black and background to transparent/white,
//     so the contrast probe measures ~21:1 and files NOTHING. A gutted
//     stylesheet reads as a HIGH-CONTRAST CLEAN audit. Worse, when it does
//     eventually file (1.00:1 invisible text on a dark-fallback section) it
//     routes the finding to css-patch-agent — the agent that caused the damage,
//     which then appends another rule to the ruined file. bugs_open/198 records
//     loancash.co.uk taking 11 items in 8 minutes that way, every one 'complete'.
//
// ── THE PREDICATE, and why it is coverage rather than size ─────────────────
//
// The obvious predicate is a byte floor: "styles.css is suspiciously small".
// It is the wrong one. bugs_open/211 is the same defect one notch subtler —
// ai-agent-orchestration.com served a ~26KB stylesheet that was missing the
// renderer's final alias :root block, with --color-heading defined ZERO times
// and 30 firm contrast failures downstream. A floor cannot see that. So:
//
//	FIRE when a custom property is REFERENCED by deployed markup and DEFINED
//	by none of the sources that could define it.
//
// That catches both shapes with one rule, and it is stated in terms of the
// thing that actually breaks — a declaration the browser throws away.
//
// A REFERENCE MEANS var(--x) WITH NO FALLBACK. `var(--x, #fff)` is deliberately
// NOT counted. With a fallback the browser silently uses the fallback and the
// page is styled; the declaration survives. It is also the defensive form this
// platform's own renderer emits on purpose — buildTokenAliases writes
// `var(--color-primary-ink, var(--color-primary))` precisely "so a stylesheet
// rendered before this block renders byte-identically". Counting the fallback
// form would fire on the code the house tells authors to write, and a signal
// that fires on the recommended remedy stops discriminating. bugs_open/211 is
// still caught, because its root reference — var(--hero-ink) — carries none.
//
// ── WHAT COUNTS AS A DEFINITION: over-approximate, on purpose ──────────────
//
// The definition set is the UNION of three sources, and where there is doubt it
// is widened rather than narrowed. Over-approximating definitions can only
// produce a FALSE NEGATIVE (a real gap missed); under-approximating produces a
// FALSE POSITIVE on a healthy site, which is what makes a new detector get
// switched off. For a first-generation check those are not equivalent costs.
//
//  1. The served stylesheets themselves — every same-host <link rel=stylesheet>
//     the corpus references, fetched live. This is where :root and the
//     renderer's alias block live, and it is the surface that gets clobbered.
//     The DB row is deliberately NOT consulted: in 198 the row was clobbered
//     too, and after a repair the row can be healthy while the FILE is not.
//     The wire is what a browser gets.
//
//  2. Inline definitions in the same corpus — a component that styles itself
//     (`.hero-section { --section-heading: … }`) defines what it then uses.
//
//  3. css_snippets.css_content, ALL rows, unscoped. injectComponentCSS adds
//     these to the served head at ASSEMBLY time, so they are invisible in
//     rendered_html; a per-function scoping of this source would have to
//     re-implement collectComponentCSS's selection logic and would drift from
//     it. Unscoped is the safe direction (see above).
//
//     ⚠ BE PRECISE ABOUT WHICH WAY THIS ERRS — the council's edit-quality seat
//     was right to press on it. `css_snippets` has no site_id (schema: id, name,
//     description, css_content, semantic_tags, applies_to, created_at), so this
//     source is GLOBAL by construction, not by choice. A property defined by a
//     snippet that never reaches site B can therefore mask a genuine gap on site
//     B — a cross-site FALSE NEGATIVE. That is the same direction as every other
//     widening here, and it is the direction chosen deliberately; but it is a
//     concrete masking mechanism rather than a theoretical one, so do not read
//     "over-approximated" as "harmless". Narrowing it means scoping snippets to
//     the site's own composition, which needs collectComponentCSS's selection
//     logic to be shared rather than duplicated.
//
// ── BLINDED IS NOT HEALTHY ─────────────────────────────────────────────────
//
// If any same-host stylesheet fails to fetch — transport error, timeout, or any
// non-2xx including a 403 from a policy layer — this check files NOTHING and
// RETRACTS NOTHING, and says so in the skip tally. It cannot then tell a
// missing definition from a definition it failed to read. asset_reference_404's
// header records what the alternative costs: a bare non-browser client drew 63
// Cloudflare refusals on one site, which under a "non-200 is a finding" rule
// would have been 63 false findings. A stylesheet that 404s is that check's
// finding, not this one's.
//
// Retraction fires only on the positive observation — every same-host
// stylesheet fetched 2xx AND every referenced property is defined.
//
// ── ROUTING: flag-only, and that is a decision rather than an omission ──────
//
// HandlerAgent is empty. The repair for a gutted stylesheet is to restore the
// real file — from the site repo's history, or by re-running webdesign-agent —
// and both are judgements. Routing this at webdesign-agent automatically would
// be actively dangerous: its analyze_design step re-rolls the palette on every
// run (check_generic_theme.go records four CSS rewrites in one day, one of
// which put a light background on a dark site), so an automatic "repair" of a
// clobbered stylesheet could hand back a site that is styled but wrong, and
// silently overwrite a hand-restored one.
//
// The cost of flag-only is real and named in bugs_open/083: "a detector whose
// output nobody drains is not neutral — it is actively misleading". The finding
// surfaces in the run's Findings and as a 'detected' work item, the same place
// image_url_404 and asset_reference_404's own items land. Per the owner ruling
// of 2026-08-02 §1, a work item type with no automated consumer is not the kind
// of shared vocabulary whose guarantees change when a producer is added, so
// this is normal council-gate scope rather than RFC scope.
//
// ── KNOWN, STATED GAPS ─────────────────────────────────────────────────────
//
//   - A property defined only by runtime JavaScript (element.style.setProperty)
//     or by a stylesheet on a DIFFERENT host is invisible here and would read as
//     undefined. Mitigated by what this estate actually does: the token
//     vocabulary is renderer-owned and emitted into a same-host styles.css.
//     External stylesheets are not fetched — probing third-party hosts to decide
//     a local site is broken is a dependency this check should not take.
//   - The "only :root survived" shape is NOT caught. Three sites currently carry
//     a 1,649-byte legacy row holding a bare :root palette block and no layout
//     rules; if one of those were deployed over a real stylesheet the properties
//     would all still be defined and every layout rule would be gone. That is a
//     size-drift question, not a coverage one, and it wants the DB↔file drift
//     check bugs_open/198 lists as its own candidate.
//   - Detection latency is the design rotation's, roughly weekly per site. For a
//     whole-site visual outage that is slow, and it is the owner's call whether
//     this also belongs on a faster clock.
package discovery_checks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&StylesheetGuttedCheck{}) }

type StylesheetGuttedCheck struct{}

func (c *StylesheetGuttedCheck) Name() string { return "stylesheet_gutted" }

const (
	// stylesheetFetchTimeout per request, matching asset_reference_404's
	// probeTimeout — the same origins, the same expectations.
	stylesheetFetchTimeout = 10 * time.Second

	// maxStylesheetFetches bounds outbound calls. A site serves one or two
	// stylesheets; this is headroom. When it bites it LOGS what it dropped and
	// the run goes BLIND rather than reporting on a partial definition set —
	// a dropped stylesheet is a dropped set of definitions, which would read as
	// missing properties.
	maxStylesheetFetches = 6

	// stylesheetBodyCap bounds what is read from one response. The largest
	// stylesheet in the fleet is ~26KB; 1MB is far beyond any real sheet and
	// stops a misrouted URL from pulling a large body into memory.
	stylesheetBodyCap = 1 << 20

	// maxMissingSamples bounds the example list carried in the spec and the
	// summary. The COUNT is always exact; only the examples are capped.
	maxMissingSamples = 20
)

// fetchSiteStylesheet is swappable in tests — the same seam probeAssetURL uses
// in check_asset_reference_404.go, and the reason the tests can prove every
// branch without a network.
//
// Returns (status, body, nil) on an HTTP response, or (0, "", err) for a
// transport failure. A transport failure is NOT a status.
var fetchSiteStylesheet = func(ctx context.Context, target string) (int, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, "", err
	}
	// An explicit, honest User-Agent. If an origin refuses it that is a non-2xx,
	// and a non-2xx blinds this check rather than informing it.
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+stylesheet_gutted)")
	req.Header.Set("Accept", "text/css,*/*")

	client := &http.Client{Timeout: stylesheetFetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, stylesheetBodyCap))
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(body), nil
}

// cssCommentRe strips /* … */ before any scan. A commented-out rule defines
// nothing and references nothing, and the alias block this check most wants to
// see is surrounded by prose that mentions the very token names it defines.
var cssCommentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)

// cssVarHardRefRe matches var(--name) with NO fallback. The captured group 2 is
// the character that ends the reference: ')' means no fallback was supplied,
// ',' means one was. See the header for why only the former counts.
var cssVarHardRefRe = regexp.MustCompile(`var\(\s*(--[A-Za-z0-9_-]+)\s*([,)])`)

// cssVarDefRe matches a custom property DEFINITION. The colon is what makes it
// a definition rather than a use, and the name is captured whole so that
// `--shadow-md:` can never be read as defining `--shadow`.
var cssVarDefRe = regexp.MustCompile(`(--[A-Za-z0-9_-]+)\s*:`)

// rendererGuaranteedTokens is the vocabulary the RENDERER promises every
// stylesheet will define — the theme's own palette/typography/structure slots
// plus the compatibility aliases buildTokenAliases appends. Only a gap in THIS
// set is a finding.
//
// ── WHY THE GATE EXISTS, and it is a correction to this check's first cut ───
//
// The first version fired on any referenced-but-undefined property. Calibrated
// across all 25 deployed/active sites on 2026-08-21 before enabling, that would
// have filed on NINETEEN of them — and seventeen were the SAME four
// component-invented names (--color-hero-title, --color-hero-subtitle,
// --color-secondary-text, --color-secondary-hover) which NO site's stylesheet
// has ever defined, including in the pre-clobber originals. That is a real
// defect, but it is a different one: a component vocabulary the renderer does
// not emit. Shipping it here would have buried the one signal this check exists
// for under seventeen copies of another, which is precisely bugs_open/083's
// "a detector whose output nobody drains is actively misleading".
//
// The honest question this check asks is therefore narrower and better posed:
// *has the stylesheet stopped honouring its own contract?* — not *is every
// token anyone ever wrote defined?*
//
// It still catches both incident shapes, which is the test of the narrowing:
// a clobbered stylesheet (bugs_open/198) defines NOTHING, so every one of these
// is missing; and bugs_open/211's missing alias block takes --color-heading and
// --hero-ink, both of which are in this set.
//
// KEPT IN SYNC BY A TEST, not by discipline: canonicalCSSTokens in
// platform/orchestration/actions/component_validation.go is the same
// vocabulary, declared for the same reason on the authoring side. This package
// cannot import it — actions imports discovery_checks, so the dependency would
// be a cycle — so the parity test parses that file's source and fails if the
// two drift. Do not edit this list without running it.
var rendererGuaranteedTokens = map[string]bool{
	// palette / theme-defined
	"--color-primary": true, "--color-primary-hover": true, "--color-primary-text": true,
	"--color-secondary": true, "--color-accent": true, "--color-background": true,
	"--color-surface": true, "--color-card-bg": true, "--color-text": true,
	"--color-text-muted": true, "--color-border": true, "--color-cta-bg": true,
	"--color-cta-text": true, "--color-header-bg": true, "--color-header-text": true,
	"--color-footer-bg": true, "--color-footer-text": true,
	"--section-text": true, "--section-text-muted": true, "--section-surface": true,
	"--section-border": true, "--section-heading": true, "--section-pad-y": true,
	"--section-pad-y-sm": true,
	"--radius":           true, "--radius-sm": true, "--radius-lg": true,
	"--shadow-sm": true, "--shadow-md": true, "--shadow-lg": true,
	"--container-max": true, "--container-pad-x": true,
	"--font-body": true, "--font-heading": true, "--font-size-base": true,
	"--line-height-base": true, "--grid-gap": true, "--card-pad": true, "--transition": true,
	// compatibility aliases guaranteed by buildTokenAliases (render_css_from_spec)
	"--border-radius": true, "--shadow": true, "--spacing-section": true,
	"--container-max-width": true, "--primary-color": true, "--secondary-color": true,
	"--accent-color": true, "--color-heading": true, "--color-white": true,
	"--color-error": true, "--hero-ink": true,
	// renderer-owned legible-ink companions (buildLegibleInkDefaults). These
	// three ARE unconditionally emitted: primary/accent always have a source
	// colour and a ground to measure against. Present in 7 of 7 served
	// stylesheets sampled `[MEASURED 2026-09-03]`.
	// --color-cta-bg-ink is DELIBERATELY ABSENT — see inkCompanionsNotGuaranteed
	// in the parity test for why policing it would be an over-reach.
	"--color-primary-ink": true, "--color-accent-ink": true, "--color-accent-text": true,
}

func stripCSSComments(css string) string {
	return cssCommentRe.ReplaceAllString(css, " ")
}

// cssVarHardRefs returns the custom properties referenced with no fallback.
func cssVarHardRefs(css string) map[string]bool {
	out := map[string]bool{}
	for _, m := range cssVarHardRefRe.FindAllStringSubmatch(stripCSSComments(css), -1) {
		if m[2] == ")" {
			out[m[1]] = true
		}
	}
	return out
}

// cssVarDefs returns the custom properties defined by this CSS text.
func cssVarDefs(css string) map[string]bool {
	out := map[string]bool{}
	for _, m := range cssVarDefRe.FindAllStringSubmatch(stripCSSComments(css), -1) {
		out[m[1]] = true
	}
	return out
}

// stylesheetGuttedItemKey is the dedup key, and it is deliberately CONSTANT
// per site rather than URL-shaped.
//
// Two reasons, both load-bearing. (1) asset_reference_404 retracts on any 2xx
// for a URL-shaped key in its own namespace; a key that collided with that
// pattern invites the next sweep to close this check's findings while the site
// is still broken. (2) The finding is about the SITE's token coverage, not about
// one file — the same missing property is reachable through every stylesheet the
// pages link, and one item per site is the honest cardinality. Precedent for a
// constant key: asset_reference_404's own "empty_src".
func stylesheetGuttedItemKey() string {
	return "stylesheet_gutted:definition_coverage"
}

// stylesheetSkipTally records why the check declined to judge. A check that
// reports only findings cannot be told apart from one that was blinded.
type stylesheetSkipTally struct {
	unresolvable  int // a href net/url cannot resolve without guessing
	external      int // not same-host; deliberately not fetched
	overCap       int
	fetchFailed   int // transport error
	nonSuccess    int // any non-2xx, including policy refusals
	noStylesheets int // corpus referenced none at all
}

func (c *StylesheetGuttedCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, '') FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("stylesheet_gutted: site lookup failed: %w", err)
	}
	if domain == "" {
		// No domain, no URL to resolve against. Nothing this check can say.
		return result, nil
	}

	refs, inPageDefs, sheetURLs, skips, err := collectStylesheetCoverage(dctx, domain)
	if err != nil {
		return nil, err
	}

	if len(refs) == 0 {
		// Nothing consumes a custom property. Not a finding, and not a
		// retraction either: a corpus that scanned as empty is the same shape a
		// blinded scan produces.
		logStylesheetSkips(dctx, skips, 0, 0)
		return result, nil
	}

	if len(sheetURLs) == 0 {
		skips.noStylesheets++
		logStylesheetSkips(dctx, skips, len(refs), 0)
		return result, nil
	}

	sort.Strings(sheetURLs)
	if len(sheetURLs) > maxStylesheetFetches {
		dropped := sheetURLs[maxStylesheetFetches:]
		skips.overCap = len(dropped)
		dctx.Logger.Warn("stylesheet_gutted: fetch cap reached — declining to judge, definitions would be incomplete",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", maxStylesheetFetches),
			zap.Strings("dropped_urls", dropped))
		logStylesheetSkips(dctx, skips, len(refs), 0)
		return result, nil
	}

	// Fetch every same-host stylesheet. ANY failure blinds the check: a
	// stylesheet not read is a set of definitions not seen, and reporting the
	// difference would be reporting our own failure as the site's.
	wireDefs := map[string]bool{}
	sheetInfo := make([]map[string]interface{}, 0, len(sheetURLs))
	for _, u := range sheetURLs {
		status, body, ferr := fetchSiteStylesheet(dctx.Ctx, u)
		if ferr != nil {
			skips.fetchFailed++
			dctx.Logger.Info("stylesheet_gutted: stylesheet fetch failed — declining to judge",
				zap.String("url", u), zap.Error(ferr))
			logStylesheetSkips(dctx, skips, len(refs), 0)
			return result, nil
		}
		if status < 200 || status >= 300 {
			skips.nonSuccess++
			dctx.Logger.Info("stylesheet_gutted: stylesheet did not return 2xx — declining to judge",
				zap.String("url", u), zap.Int("status", status))
			logStylesheetSkips(dctx, skips, len(refs), 0)
			return result, nil
		}
		for name := range cssVarDefs(body) {
			wireDefs[name] = true
		}
		sheetInfo = append(sheetInfo, map[string]interface{}{
			"url": u, "http_status": status, "bytes": len(body),
		})
	}

	snippetDefs, err := collectSnippetDefs(dctx)
	if err != nil {
		return nil, err
	}

	// Only a gap in the RENDERER'S GUARANTEED vocabulary is a finding — see
	// rendererGuaranteedTokens for the calibration that forced this gate, and
	// for why an "any undefined property" predicate filed on 19 of 25 sites.
	// Non-canonical gaps are counted and logged so the other defect stays
	// visible, but they are not this check's to file.
	missing := make([]string, 0)
	nonCanonicalGaps := 0
	for name := range refs {
		if wireDefs[name] || inPageDefs[name] || snippetDefs[name] {
			continue
		}
		if !rendererGuaranteedTokens[name] {
			nonCanonicalGaps++
			continue
		}
		missing = append(missing, name)
	}
	sort.Strings(missing)
	if nonCanonicalGaps > 0 {
		dctx.Logger.Info("stylesheet_gutted: undefined NON-canonical properties seen and deliberately not filed",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("count", nonCanonicalGaps),
			zap.String("why", "component-invented vocabulary the renderer never emits — a different defect; see rendererGuaranteedTokens"))
	}

	logStylesheetSkips(dctx, skips, len(refs), len(wireDefs))

	if len(missing) == 0 {
		// The positive observation, and the only thing that may retract.
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType: "stylesheet_gutted",
			ItemKey:  stylesheetGuttedItemKey(),
			Reason: fmt.Sprintf("every renderer-guaranteed custom property referenced without a fallback is defined (%d referenced, %d served stylesheet(s), %d definitions)",
				len(refs), len(sheetURLs), len(wireDefs)),
		})
		return result, nil
	}

	samples := missing
	if len(samples) > maxMissingSamples {
		samples = samples[:maxMissingSamples]
	}

	spec := map[string]interface{}{
		"check":            "stylesheet_gutted",
		"kind":             "undefined_custom_properties",
		"missing_count":    len(missing),
		"missing":          samples,
		"referenced_count": len(refs),
		"stylesheets":      sheetInfo,
	}
	specJSON, _ := json.Marshal(spec)

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":         "stylesheet_gutted",
		"kind":          "undefined_custom_properties",
		"missing_count": len(missing),
		"missing":       samples,
		"stylesheets":   sheetInfo,
	})

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "stylesheet_gutted",
		Severity: "high",
		Summary: fmt.Sprintf("Served stylesheet no longer defines %d renderer-guaranteed custom propert%s the pages use (e.g. %s) — deployed pages will render unstyled or with invisible text",
			len(missing), pluralY(len(missing)), strings.Join(firstN(samples, 3), ", ")),
		SpecJSON:  string(specJSON),
		Priority:  70,
		Status:    "detected",
		CreatedBy: dctx.AgentType,
		ItemKey:   stylesheetGuttedItemKey(),
		BatchID:   dctx.BatchID,
		// HandlerAgent intentionally empty — flag-only, see the header.
	})

	return result, nil
}

func pluralY(n int) string {
	if n == 1 {
		return "y"
	}
	return "ies"
}

func firstN(s []string, n int) []string {
	if len(s) < n {
		return s
	}
	return s[:n]
}

// collectStylesheetCoverage scans the deployed corpus for (a) custom properties
// referenced without a fallback, (b) custom properties defined inline in that
// same corpus, and (c) the same-host stylesheet URLs the corpus links.
//
// The population and the liveness predicates are the shared builders used by
// check_asset_reference_404 — not a fresh spelling. That check's council round
// established why: a hand-rolled `deployed_at IS NOT NULL` silently omits live
// pages (28 pages ship under another status, bugs_open/185), and a check that
// omits pages reports clean for the wrong reason.
func collectStylesheetCoverage(dctx DiscoveryCheckContext, domain string) (
	refs map[string]bool, inPageDefs map[string]bool, sheetURLs []string,
	skips stylesheetSkipTally, err error) {

	refs = map[string]bool{}
	inPageDefs = map[string]bool{}
	seenSheet := map[string]bool{}

	// scan pulls CSS out of one rendered artefact. goquery rather than a regex
	// over the HTML: a regex cannot tell an ELEMENT from a MENTION of one, and
	// this corpus is full of tool pages whose JavaScript talks about CSS.
	// The html parser treats <script> content as raw text, so a var(--x) written
	// inside a tool's own JS is never reached.
	scan := func(html, pageURL string) {
		doc, derr := goquery.NewDocumentFromReader(strings.NewReader(html))
		if derr != nil {
			dctx.Logger.Warn("stylesheet_gutted: parse failed", zap.Error(derr))
			return
		}

		collect := func(css string) {
			for name := range cssVarHardRefs(css) {
				refs[name] = true
			}
			for name := range cssVarDefs(css) {
				inPageDefs[name] = true
			}
		}

		doc.Find("style").Each(func(_ int, s *goquery.Selection) {
			collect(s.Text())
		})
		doc.Find("[style]").Each(func(_ int, s *goquery.Selection) {
			v, _ := s.Attr("style")
			collect(v)
		})

		doc.Find("link[href]").Each(func(_ int, s *goquery.Selection) {
			rel, _ := s.Attr("rel")
			isSheet := false
			// rel matched by token, not equality: "stylesheet preload" and
			// "alternate stylesheet" are both real.
			for _, tok := range strings.Fields(strings.ToLower(rel)) {
				if tok == "stylesheet" {
					isSheet = true
					break
				}
			}
			if !isSheet {
				return
			}
			href, _ := s.Attr("href")
			trimmed := strings.TrimSpace(href)
			if trimmed == "" || trimmed == "#" {
				// Resolves to the document itself; never fetched. That is
				// asset_reference_404's empty_src finding, not ours.
				skips.unresolvable++
				return
			}
			resolved, external, ok := resolveAssetURL(domain, pageURL, trimmed)
			if !ok {
				skips.unresolvable++
				return
			}
			if external {
				// Not fetched on purpose: deciding a local site is broken on the
				// basis of a third-party host's availability is a dependency this
				// check should not take. Its definitions are therefore unseen,
				// which is a stated false-negative source in the header.
				skips.external++
				return
			}
			if !seenSheet[resolved] {
				seenSheet[resolved] = true
				sheetURLs = append(sheetURLs, resolved)
			}
		})
	}

	pageRows, qerr := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT pc.rendered_html, COALESCE(p.url, '')
		FROM page_components pc
		JOIN pages p ON pc.page_id = p.id
		WHERE p.site_id = $1
		  AND pc.build_status = 'deployed'
		  AND pc.locked_at IS NULL
		  AND pc.rendered_html IS NOT NULL
		  AND `+datahelpers.PageHasShippedPredicateFor("p")+`
		  AND `+datahelpers.PageWantedLivePredicateFor("p")+`
	`, dctx.SiteID)
	if qerr != nil {
		return nil, nil, nil, skips, fmt.Errorf("stylesheet_gutted: page_components scan failed: %w", qerr)
	}
	for pageRows.Next() {
		var html, pageURL string
		if serr := pageRows.Scan(&html, &pageURL); serr != nil {
			dctx.Logger.Warn("stylesheet_gutted: scan page component failed", zap.Error(serr))
			continue
		}
		scan(html, pageURL)
	}
	if rerr := pageRows.Err(); rerr != nil {
		pageRows.Close()
		return nil, nil, nil, skips, fmt.Errorf("stylesheet_gutted: page_components scan failed: %w", rerr)
	}
	pageRows.Close()

	// Chrome: one stylesheet link here is on every page of the site.
	chromeRows, cerr := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT sc.rendered_html
		FROM site_components sc
		WHERE sc.site_id = $1
		  AND sc.locked_at IS NULL
		  AND sc.rendered_html IS NOT NULL
	`, dctx.SiteID)
	if cerr != nil {
		return nil, nil, nil, skips, fmt.Errorf("stylesheet_gutted: site_components scan failed: %w", cerr)
	}
	defer chromeRows.Close()
	for chromeRows.Next() {
		var html string
		if serr := chromeRows.Scan(&html); serr != nil {
			dctx.Logger.Warn("stylesheet_gutted: scan site component failed", zap.Error(serr))
			continue
		}
		scan(html, "")
	}
	return refs, inPageDefs, sheetURLs, skips, chromeRows.Err()
}

// collectSnippetDefs reads css_snippets, which injectComponentCSS adds to the
// served head at assembly time. These definitions are real and are invisible in
// rendered_html, so omitting them would false-positive on any component that
// defines its own tokens through a snippet. Unscoped by design — see the header.
func collectSnippetDefs(dctx DiscoveryCheckContext) (map[string]bool, error) {
	defs := map[string]bool{}
	rows, err := dctx.DB.QueryContext(dctx.Ctx,
		`SELECT css_content FROM css_snippets WHERE css_content IS NOT NULL`)
	if err != nil {
		return nil, fmt.Errorf("stylesheet_gutted: css_snippets scan failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var css string
		if serr := rows.Scan(&css); serr != nil {
			dctx.Logger.Warn("stylesheet_gutted: scan css snippet failed", zap.Error(serr))
			continue
		}
		for name := range cssVarDefs(css) {
			defs[name] = true
		}
	}
	return defs, rows.Err()
}

func logStylesheetSkips(dctx DiscoveryCheckContext, s stylesheetSkipTally, refCount, defCount int) {
	dctx.Logger.Info("stylesheet_gutted: scan complete",
		zap.String("site_id", dctx.SiteID.String()),
		zap.Int("properties_referenced", refCount),
		zap.Int("definitions_found", defCount),
		zap.Int("skip_unresolvable", s.unresolvable),
		zap.Int("skip_external", s.external),
		zap.Int("skip_over_cap", s.overCap),
		zap.Int("skip_fetch_failed", s.fetchFailed),
		zap.Int("skip_non_success", s.nonSuccess),
		zap.Int("skip_no_stylesheets", s.noStylesheets))
}
