// FILE: internal/adapters/browserrunner/run_checks_action.go
//
// The Tier-4 check runner (P0 + P1 + P2). Drives the deployed page in headless
// Chromium (playwright-go) under one or more PROFILES and evaluates the tool's
// acceptance criteria against the live DOM:
//
//   page_status_ok         — the navigation response is HTTP 200
//   selector_exists        — the FULL selector matches in the live DOM (this is
//                            the tier that asserts "#tableWrap tr" for real —
//                            Tier 2 only confirms the anchor statically)
//   selector_count         — selector matches at least once (same as exists at v0)
//   no_console_errors      — no console.error and no uncaught page errors across
//                            the whole session, INCLUDING during interactions
//   no_horizontal_overflow — scrollWidth <= clientWidth (P1; typically mobile)
//   interaction            — run steps (fill/click/select) then assert an
//                            expect selector exists / its text matches (P2 —
//                            the tier that asserts a tool actually WORKS, not
//                            just that it boots)
//
// Profiles (P1): "desktop" 1366×900; "mobile" a 390×844 touch viewport. A check
// with no `profiles` runs on every requested profile; a check pinned to
// profiles runs only on those. Results carry their profile. A request with no
// profiles defaults to desktop (the P0 contract). Everything not applicable to
// a run is reported in `skipped`, never faked; `-EDIT` ids are skipped;
// navigation failure is a check FAIL, not an infra error.
//
// The browser is launched per (url, profile): a crashed Chromium poisons one
// run, not the pod. The `browserPage` interface is the test seam — real runs
// use Playwright (openChromium); tests inject a fake page.

package browserrunner

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	playwright "github.com/mxschmitt/playwright-go"
	"go.uber.org/zap"
)

// RunChecksRequest is the body.data payload (PLAN_tool_acceptance_runner rev 2).
type RunChecksRequest struct {
	RunID        string   `json:"run_id"`
	URLs         []string `json:"urls"`
	Profiles     []string `json:"profiles"`
	CriteriaJSON string   `json:"criteria_json"`
	Function     string   `json:"function"`
	SiteID       string   `json:"site_id"`
}

// CheckResult is one evaluated check on one URL under one profile.
//
// Scope/Culprit/Component are set by document-level checks (currently
// no_horizontal_overflow) so the judge can ATTRIBUTE the failure. The overflow
// is measured on the whole document, but the acceptance run is scoped to ONE
// tool: without attribution, a site footer that overflows raises an identical,
// unfixable tool ticket for every tool on the site (observed on vonc.com,
// 2026-07-14 — div.footer-legal, 506px at a 390px viewport, on every page).
type CheckResult struct {
	CheckID string `json:"check_id"`
	Profile string `json:"profile"`
	URL     string `json:"url"`
	Pass    bool   `json:"pass"`
	Detail  string `json:"detail"`

	Scope     string `json:"scope,omitempty"`     // tool | chrome | unknown
	Culprit   string `json:"culprit,omitempty"`   // human: "div.footer-legal (506px)"
	Component string `json:"component,omitempty"` // nearest structural ancestor, e.g. "site-footer"
	// Machine-usable repair handles: the culprit as a CSS selector, and which
	// site_components slot owns it. A fixer needs BOTH — a description cannot be
	// fed to a stylesheet, and without the slot the fixer must guess (the live
	// one defaulted to "header" and "fixed" the wrong thing).
	CulpritSelector string `json:"culprit_selector,omitempty"` // e.g. "div.footer-legal"
	Slot            string `json:"slot,omitempty"`             // header | footer | head | ""
}

// Result scopes. "unknown" means the tool's container could not be located, so
// the failure must NOT be blamed on site chrome — it falls back to the tool.
const (
	ScopeTool    = "tool"
	ScopeChrome  = "chrome"
	ScopeUnknown = "unknown"
)

// overflowInfo is what the page reports about a horizontal overflow.
type overflowInfo struct {
	Culprit   string // widest element crossing the viewport edge, with its width
	Selector  string // the same element as a CSS selector a fixer can act on
	Component string // its nearest structural ancestor (header/footer/section)
	Slot      string // the site_components slot that owns it: header | footer | head
	InTool    bool   // culprit lies inside the tool's container
	Located   bool   // the tool's container was found at all
}

// ScreenshotRef points at one captured full-page screenshot — the P3 evidence
// for a failing (url, profile) run. URI is the durable s3:// pointer (safe for
// travelling-doc notes); ViewURL is a presigned GET that expires (for the work
// item's spec). FailingChecks lists the id@profile instances it evidences.
type ScreenshotRef struct {
	URL           string   `json:"url"`
	Profile       string   `json:"profile"`
	URI           string   `json:"uri"`
	ViewURL       string   `json:"view_url,omitempty"`
	FailingChecks []string `json:"failing_checks"`
}

// RunChecksResult is the response body.data.
type RunChecksResult struct {
	RunID   string        `json:"run_id"`
	Results []CheckResult `json:"results"`
	Skipped []CheckResult `json:"skipped"` // not evaluated — never a fake pass
	// Screenshots is present only when a (url, profile) run had failures AND
	// object storage is configured — evidence, never load-bearing.
	Screenshots []ScreenshotRef `json:"screenshots,omitempty"`
	Summary     struct {
		Passed  int `json:"passed"`
		Failed  int `json:"failed"`
		Skipped int `json:"skipped"`
	} `json:"summary"`
}

// ── criteria document (schema v0, PLAN_tool_acceptance_runner) ──────────────

type criteriaDoc struct {
	Profiles []string        `json:"profiles"`
	Checks   []criteriaCheck `json:"checks"`
	// Container is the tool's root element in the page. Optional: PLANs written
	// before attribution existed omit it, and defaultToolContainer covers the
	// two conventions in the wild. A per-check container overrides it.
	Container string `json:"container"`
}

type criteriaCheck struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Selector  string         `json:"selector"`
	Path      string         `json:"path"`
	Profiles  []string       `json:"profiles"`
	Steps     []criteriaStep `json:"steps"`
	Expect    criteriaExpect `json:"expect"`
	Container string         `json:"container"`
}

// defaultToolContainer matches the tool root under BOTH delivery conventions:
// the generator's `.tool-container`, and the page-section path's
// `.tool-<function>-section` (vonc.com). Used when the PLAN names no container.
const defaultToolContainer = `.tool-container, [class*="tool-"][class*="-section"]`

// toolContainer resolves the container selector for a check: per-check, else
// document-level, else the convention.
func toolContainer(doc criteriaDoc, ch criteriaCheck) string {
	if s := strings.TrimSpace(ch.Container); s != "" {
		return s
	}
	if s := strings.TrimSpace(doc.Container); s != "" {
		return s
	}
	return defaultToolContainer
}

type criteriaStep struct {
	Action   string `json:"action"` // fill | click | select
	Selector string `json:"selector"`
	Value    string `json:"value"`
}

type criteriaExpect struct {
	Selector    string `json:"selector"`
	TextMatches string `json:"text_matches"`
}

const (
	desktopWidth  = 1366
	desktopHeight = 900
	mobileWidth   = 390
	mobileHeight  = 844
	mobileUA      = "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) " +
		"AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1"

	navigationTimeout = 30 * time.Second
	settleDelay       = 2 * time.Second // let JS-built DOM render before asserting
	stepDelay         = 300 * time.Millisecond
	runDeadline       = 120 * time.Second
)

// browserPage is what per-check evaluation drives. The real implementation
// wraps a live Playwright page (chromiumPage); tests inject a fake.
type browserPage interface {
	Status() int           // navigation response status (0 = nav failed)
	NavError() string      // non-empty when navigation itself failed
	ConsoleErrors() []string
	Count(selector string) int  // matches in the live DOM
	// HorizontalOverflow reports scrollWidth > clientWidth, names the widest
	// offender, and says whether it sits inside the given tool container.
	HorizontalOverflow(container string) (bool, overflowInfo, error)
	Do(step criteriaStep) error // one interaction step
	Text(selector string) (string, error)
	// Screenshot captures the page as PNG bytes (P3 failure evidence).
	Screenshot(fullPage bool) ([]byte, error)
	Close()
}

// openFunc launches a browser for one (url, profile) and returns a driven page.
// Swappable in tests. A navigation failure is reported via the page's NavError,
// not an error return; an error return means infra failure (browser/driver).
type openFunc func(ctx context.Context, url, profile string, logger *zap.Logger) (browserPage, error)

type RunChecksAction struct {
	logger *zap.Logger
	open   openFunc
	store  screenshotStore // nil = screenshots disabled (P0 behaviour)
}

func NewRunChecksAction(logger *zap.Logger, store screenshotStore) *RunChecksAction {
	return &RunChecksAction{logger: logger.Named("run_checks"), open: openChromium, store: store}
}

// Execute runs the criteria for each requested URL under each requested profile.
func (a *RunChecksAction) Execute(ctx context.Context, req RunChecksRequest) (*RunChecksResult, error) {
	if len(req.URLs) == 0 {
		return nil, fmt.Errorf("run_checks: no urls in request")
	}
	if strings.TrimSpace(req.CriteriaJSON) == "" {
		return nil, fmt.Errorf("run_checks: empty criteria_json — an undocumented tool gets a needs_criteria note upstream, not a browser run")
	}
	var crit criteriaDoc
	if err := json.Unmarshal([]byte(req.CriteriaJSON), &crit); err != nil {
		return nil, fmt.Errorf("run_checks: criteria_json does not parse: %w", err)
	}

	runCtx, cancel := context.WithTimeout(ctx, runDeadline)
	defer cancel()

	profiles := resolveProfiles(req.Profiles)
	out := &RunChecksResult{RunID: req.RunID}

	for urlIdx, url := range req.URLs {
		for _, profile := range profiles {
			applicable, skipped := splitByProfile(crit, profile, url)
			out.Skipped = append(out.Skipped, skipped...)
			if len(applicable) == 0 {
				continue
			}

			page, err := a.open(runCtx, url, profile, a.logger)
			if err != nil {
				// Infra failure (browser/driver would not launch) — the whole
				// run fails; a navigation failure is NOT this path.
				return nil, fmt.Errorf("run_checks: browser open failed for %s [%s]: %w", url, profile, err)
			}
			res := evaluateOnPage(page, crit, applicable, profile, url)
			// P3: evidence while the page is still open — a failing verdict
			// carries what the page actually looked like.
			if ref, ok := a.captureFailureEvidence(runCtx, page, req, res, profile, url, urlIdx); ok {
				out.Screenshots = append(out.Screenshots, ref)
			}
			page.Close()
			out.Results = append(out.Results, res...)
		}
	}

	for _, r := range out.Results {
		if r.Pass {
			out.Summary.Passed++
		} else {
			out.Summary.Failed++
		}
	}
	out.Summary.Skipped = len(out.Skipped)

	a.logger.Info("run_checks complete",
		zap.String("run_id", req.RunID), zap.String("function", req.Function),
		zap.Strings("profiles", profiles),
		zap.Int("passed", out.Summary.Passed), zap.Int("failed", out.Summary.Failed),
		zap.Int("skipped", out.Summary.Skipped))
	return out, nil
}

// captureFailureEvidence takes and stores a full-page screenshot when this
// (url, profile) run has at least one failing check. Best-effort by design:
// no store, a capture error, or an upload error all degrade to a log line —
// evidence must never fail, slow-fail, or alter the verdict. Navigation
// failures are excluded: there is no page state worth photographing.
func (a *RunChecksAction) captureFailureEvidence(ctx context.Context, page browserPage,
	req RunChecksRequest, results []CheckResult, profile, url string, urlIdx int) (ScreenshotRef, bool) {

	if a.store == nil || page.NavError() != "" {
		return ScreenshotRef{}, false
	}
	var failing []string
	for _, r := range results {
		if !r.Pass {
			failing = append(failing, r.CheckID+"@"+r.Profile)
		}
	}
	if len(failing) == 0 {
		return ScreenshotRef{}, false
	}

	png, err := page.Screenshot(true)
	if err != nil {
		a.logger.Warn("failure screenshot capture failed — evidence skipped, verdict unaffected",
			zap.String("url", url), zap.String("profile", profile), zap.Error(err))
		return ScreenshotRef{}, false
	}
	key := screenshotKey(req.SiteID, req.Function, req.RunID, profile, urlIdx)
	uri, viewURL, err := a.store.Save(ctx, key, png)
	if err != nil {
		a.logger.Warn("failure screenshot upload failed — evidence skipped, verdict unaffected",
			zap.String("key", key), zap.Error(err))
		return ScreenshotRef{}, false
	}
	a.logger.Info("failure screenshot stored",
		zap.String("uri", uri), zap.String("profile", profile),
		zap.Strings("failing_checks", failing), zap.Int("bytes", len(png)))
	return ScreenshotRef{URL: url, Profile: profile, URI: uri, ViewURL: viewURL, FailingChecks: failing}, true
}

// resolveProfiles: desktop always runs unless only mobile is asked for; mobile
// runs when requested. Empty request → desktop (the P0 default).
func resolveProfiles(requested []string) []string {
	if len(requested) == 0 {
		return []string{"desktop"}
	}
	var out []string
	for _, p := range []string{"desktop", "mobile"} { // stable order
		if contains(requested, p) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		// Unknown profiles only — nothing we can run.
		return nil
	}
	return out
}

// splitByProfile returns the checks applicable to `profile` and the skips (with
// reasons) for this (profile, url). A check's own `profiles` list, if present,
// gates it to those profiles.
func splitByProfile(crit criteriaDoc, profile, url string) ([]criteriaCheck, []CheckResult) {
	var applicable []criteriaCheck
	var skipped []CheckResult
	skip := func(id, why string) {
		skipped = append(skipped, CheckResult{CheckID: id, Profile: profile, URL: url, Pass: false, Detail: "SKIPPED: " + why})
	}
	for _, ch := range crit.Checks {
		if strings.HasSuffix(ch.ID, "-EDIT") {
			skip(ch.ID, "placeholder selector (-EDIT)")
			continue
		}
		if len(ch.Profiles) > 0 && !contains(ch.Profiles, profile) {
			skip(ch.ID, "not run on profile "+profile)
			continue
		}
		switch ch.Type {
		case "page_status_ok", "selector_exists", "selector_count",
			"no_console_errors", "no_horizontal_overflow", "interaction":
			applicable = append(applicable, ch)
		default:
			skip(ch.ID, ch.Type+" not implemented")
		}
	}
	return applicable, skipped
}

// evaluateOnPage evaluates the applicable checks against a live page. Console
// errors are checked LAST so they capture anything an interaction triggered.
func evaluateOnPage(page browserPage, doc criteriaDoc, checks []criteriaCheck, profile, url string) []CheckResult {
	var results []CheckResult
	add := func(id string, pass bool, detail string) {
		results = append(results, CheckResult{CheckID: id, Profile: profile, URL: url, Pass: pass, Detail: detail})
	}
	addScoped := func(id string, pass bool, detail string, r CheckResult) {
		r.CheckID, r.Profile, r.URL, r.Pass, r.Detail = id, profile, url, pass, detail
		results = append(results, r)
	}
	navigated := page.NavError() == ""

	var consoleChecks []criteriaCheck
	for _, ch := range checks {
		if ch.Type == "no_console_errors" {
			consoleChecks = append(consoleChecks, ch)
			continue
		}
		if !navigated && ch.Type != "page_status_ok" {
			add(ch.ID, false, "not evaluated: navigation failed ("+page.NavError()+")")
			continue
		}
		switch ch.Type {
		case "page_status_ok":
			if page.NavError() != "" {
				add(ch.ID, false, "navigation failed: "+page.NavError())
			} else if page.Status() == 200 {
				add(ch.ID, true, "HTTP 200")
			} else {
				add(ch.ID, false, fmt.Sprintf("HTTP %d", page.Status()))
			}
		case "selector_exists", "selector_count":
			if n := page.Count(ch.Selector); n > 0 {
				add(ch.ID, true, fmt.Sprintf("%d element(s) match %s in the live DOM", n, ch.Selector))
			} else {
				add(ch.ID, false, "no element matches "+ch.Selector+" in the live DOM after settle")
			}
		case "no_horizontal_overflow":
			over, info, err := page.HorizontalOverflow(toolContainer(doc, ch))
			if err != nil {
				add(ch.ID, false, "could not measure overflow: "+err.Error())
			} else if over {
				// ATTRIBUTE the offender: this decides whether the judge raises a
				// tool ticket or a site-chrome one (see HorizontalOverflow).
				detail := "page overflows horizontally (scrollWidth > clientWidth) on " + profile
				if info.Culprit != "" {
					detail += "; widest offending element: " + info.Culprit
				}
				scope := ScopeUnknown
				switch {
				case !info.Located || info.Culprit == "":
					// Tool container not found: do NOT blame site chrome on a guess.
					detail += " (tool container not found — attribution unknown)"
				case info.InTool:
					scope = ScopeTool
					detail += " — inside the tool"
				default:
					scope = ScopeChrome
					detail += " — OUTSIDE the tool container: site chrome"
					if info.Component != "" {
						detail += " (" + info.Component + ")"
					}
				}
				addScoped(ch.ID, false, detail, CheckResult{
					Scope: scope, Culprit: info.Culprit, Component: info.Component,
					CulpritSelector: info.Selector, Slot: info.Slot,
				})
			} else {
				add(ch.ID, true, "no horizontal overflow on "+profile)
			}
		case "interaction":
			pass, detail := runInteraction(page, ch)
			add(ch.ID, pass, detail)
		}
	}

	// Console last — captures interaction-triggered errors too.
	for _, ch := range consoleChecks {
		if !navigated {
			add(ch.ID, false, "not evaluated: navigation failed")
			continue
		}
		errs := page.ConsoleErrors()
		if len(errs) == 0 {
			add(ch.ID, true, "no console errors")
		} else {
			sample := errs
			if len(sample) > 3 {
				sample = sample[:3]
			}
			add(ch.ID, false, fmt.Sprintf("%d console error(s): %s", len(errs), strings.Join(sample, " | ")))
		}
	}
	return results
}

// runInteraction executes the check's steps then asserts its expect. A step
// that can't run (missing selector) fails the check with that detail — the
// tool didn't behave.
func runInteraction(page browserPage, ch criteriaCheck) (bool, string) {
	for i, st := range ch.Steps {
		if err := page.Do(st); err != nil {
			return false, fmt.Sprintf("step %d (%s %s) failed: %s", i+1, st.Action, st.Selector, err.Error())
		}
	}
	if ch.Expect.Selector == "" {
		// No assertion beyond the steps completing — steps ran cleanly.
		return true, "interaction steps completed"
	}
	if page.Count(ch.Expect.Selector) == 0 {
		return false, "expected element " + ch.Expect.Selector + " absent after interaction"
	}
	if ch.Expect.TextMatches != "" {
		txt, err := page.Text(ch.Expect.Selector)
		if err != nil {
			return false, "could not read " + ch.Expect.Selector + " text: " + err.Error()
		}
		re, err := regexp.Compile(ch.Expect.TextMatches)
		if err != nil {
			return false, "invalid text_matches pattern " + ch.Expect.TextMatches
		}
		if !re.MatchString(txt) {
			return false, fmt.Sprintf("%s text %q does not match /%s/ after interaction", ch.Expect.Selector, strings.TrimSpace(txt), ch.Expect.TextMatches)
		}
	}
	return true, "interaction produced the expected result (" + ch.Expect.Selector + ")"
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// ── the real Playwright page ────────────────────────────────────────────────

type chromiumPage struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	page    playwright.Page
	status  int
	navErr  string
	console []string
}

func openChromium(ctx context.Context, url, profile string, logger *zap.Logger) (browserPage, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("start playwright driver: %w", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		_ = pw.Stop()
		return nil, fmt.Errorf("launch chromium: %w", err)
	}

	ctxOpts := playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{Width: desktopWidth, Height: desktopHeight},
	}
	if profile == "mobile" {
		ctxOpts.Viewport = &playwright.Size{Width: mobileWidth, Height: mobileHeight}
		ctxOpts.UserAgent = playwright.String(mobileUA)
		ctxOpts.DeviceScaleFactor = playwright.Float(3)
		ctxOpts.IsMobile = playwright.Bool(true)
		ctxOpts.HasTouch = playwright.Bool(true)
	}
	bctx, err := browser.NewContext(ctxOpts)
	if err != nil {
		_ = browser.Close()
		_ = pw.Stop()
		return nil, fmt.Errorf("new context: %w", err)
	}
	page, err := bctx.NewPage()
	if err != nil {
		_ = browser.Close()
		_ = pw.Stop()
		return nil, fmt.Errorf("new page: %w", err)
	}

	cp := &chromiumPage{pw: pw, browser: browser, page: page}
	page.OnConsole(func(m playwright.ConsoleMessage) {
		if m.Type() == "error" {
			cp.console = append(cp.console, m.Text())
		}
	})
	page.OnPageError(func(e error) { cp.console = append(cp.console, "pageerror: "+e.Error()) })

	resp, err := page.Goto(url, playwright.PageGotoOptions{
		Timeout:   playwright.Float(float64(navigationTimeout.Milliseconds())),
		WaitUntil: playwright.WaitUntilStateLoad,
	})
	if err != nil {
		cp.navErr = err.Error()
		return cp, nil // navigation failure is a CHECK failure, not infra
	}
	if resp != nil {
		cp.status = resp.Status()
	}

	select {
	case <-time.After(settleDelay):
	case <-ctx.Done():
		cp.Close()
		return nil, ctx.Err()
	}
	return cp, nil
}

func (c *chromiumPage) Status() int           { return c.status }
func (c *chromiumPage) NavError() string       { return c.navErr }
func (c *chromiumPage) ConsoleErrors() []string { return c.console }

func (c *chromiumPage) Count(selector string) int {
	n, err := c.page.Locator(selector).Count()
	if err != nil {
		return 0 // an invalid/unmatched selector counts as absent
	}
	return n
}

// HorizontalOverflow measures the DOCUMENT, so a failure may be caused by site
// chrome (header/footer) rather than the tool. It therefore also names the
// widest offending element AND says whether that element lies inside the tool's
// container — without which one overflowing site footer raises an identical,
// unactionable improve_tool ticket for EVERY tool on the site, and the fixer
// edits a tool that cannot possibly fix a template footer. Observed 2026-07-14:
// vonc.com's div.footer-legal (506px at a 390px viewport) failed the quiz's
// mobile-fit check, on every page of the site.
func (c *chromiumPage) HorizontalOverflow(container string) (bool, overflowInfo, error) {
	v, err := c.page.Evaluate(`(containerSel) => {
		const vw = document.documentElement.clientWidth;
		const over = document.documentElement.scrollWidth - vw;
		if (over <= 2) return {over: over};

		const describe = (el) => el.tagName.toLowerCase()
			+ (el.id ? '#' + el.id : '')
			+ (el.className && typeof el.className === 'string' && el.className.trim()
			   ? '.' + el.className.trim().split(/\s+/).join('.')
			   : '');

		// Widest element crossing the viewport edge; deepest wins ties, since an
		// ancestor is usually just inheriting an overflowing child's width.
		let best = null, bestEl = null;
		for (const el of document.querySelectorAll('*')) {
			const r = el.getBoundingClientRect();
			if (r.width === 0 && r.height === 0) continue;
			if (r.right > vw + 1 || r.left < -1) {
				let depth = 0;
				for (let n = el; n.parentElement; n = n.parentElement) depth++;
				if (!best || r.width > best.width || (r.width === best.width && depth > best.depth)) {
					best = {width: Math.round(r.width), depth: depth};
					bestEl = el;
				}
			}
		}
		if (!bestEl) return {over: over};

		// Attribution: is the offender inside the tool, or is it site chrome?
		let tool = null;
		try { tool = containerSel ? document.querySelector(containerSel) : null; } catch (e) { tool = null; }

		// Nearest structural ancestor names the component a fixer would edit.
		const structural = bestEl.closest('header, footer, nav, main, section, [class*="-section"]');
		let component = '';
		if (structural) {
			component = (structural.className && typeof structural.className === 'string'
				? structural.className.trim().split(/\s+/)[0] : '') || structural.tagName.toLowerCase();
		}

		// Which site_components slot owns the offender? The fixer edits ONE slot's
		// rendered_html, so this must be derived, never defaulted.
		let slot = '';
		if (bestEl.closest('footer')) slot = 'footer';
		else if (bestEl.closest('header')) slot = 'header';

		return {
			over: over,
			culprit: describe(bestEl) + ' (' + best.width + 'px)',
			selector: describe(bestEl),
			component: component,
			slot: slot,
			located: !!tool,
			inTool: !!(tool && tool.contains(bestEl)),
		};
	}`, container)
	if err != nil {
		return false, overflowInfo{}, err
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return false, overflowInfo{}, nil
	}
	info := overflowInfo{}
	info.Culprit, _ = m["culprit"].(string)
	info.Selector, _ = m["selector"].(string)
	info.Component, _ = m["component"].(string)
	info.Slot, _ = m["slot"].(string)
	info.InTool, _ = m["inTool"].(bool)
	info.Located, _ = m["located"].(bool)
	// JS numbers come back as float64/int; tolerate 2px of rounding.
	switch n := m["over"].(type) {
	case float64:
		return n > 2, info, nil
	case int:
		return n > 2, info, nil
	default:
		return false, info, nil
	}
}

func (c *chromiumPage) Do(step criteriaStep) error {
	loc := c.page.Locator(step.Selector).First()
	var err error
	switch step.Action {
	case "fill":
		err = loc.Fill(step.Value)
	case "click":
		err = loc.Click()
	case "select":
		_, err = loc.SelectOption(playwright.SelectOptionValues{Values: playwright.StringSlice(step.Value)})
	default:
		return fmt.Errorf("unknown step action %q", step.Action)
	}
	if err != nil {
		return err
	}
	time.Sleep(stepDelay) // let handlers/render settle before the next step or assertion
	return nil
}

func (c *chromiumPage) Text(selector string) (string, error) {
	return c.page.Locator(selector).First().InnerText()
}

func (c *chromiumPage) Screenshot(fullPage bool) ([]byte, error) {
	return c.page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(fullPage),
		Type:     playwright.ScreenshotTypePng,
	})
}

func (c *chromiumPage) Close() {
	if c.browser != nil {
		_ = c.browser.Close()
	}
	if c.pw != nil {
		_ = c.pw.Stop()
	}
}
