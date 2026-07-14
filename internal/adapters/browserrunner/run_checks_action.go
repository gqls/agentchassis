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
type CheckResult struct {
	CheckID string `json:"check_id"`
	Profile string `json:"profile"`
	URL     string `json:"url"`
	Pass    bool   `json:"pass"`
	Detail  string `json:"detail"`
}

// RunChecksResult is the response body.data.
type RunChecksResult struct {
	RunID   string        `json:"run_id"`
	Results []CheckResult `json:"results"`
	Skipped []CheckResult `json:"skipped"` // not evaluated — never a fake pass
	Summary struct {
		Passed  int `json:"passed"`
		Failed  int `json:"failed"`
		Skipped int `json:"skipped"`
	} `json:"summary"`
}

// ── criteria document (schema v0, PLAN_tool_acceptance_runner) ──────────────

type criteriaDoc struct {
	Profiles []string        `json:"profiles"`
	Checks   []criteriaCheck `json:"checks"`
}

type criteriaCheck struct {
	ID       string         `json:"id"`
	Type     string         `json:"type"`
	Selector string         `json:"selector"`
	Path     string         `json:"path"`
	Profiles []string       `json:"profiles"`
	Steps    []criteriaStep `json:"steps"`
	Expect   criteriaExpect `json:"expect"`
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
	Count(selector string) int         // matches in the live DOM
	HorizontalOverflow() (bool, string, error) // scrollWidth > clientWidth; names the widest offender
	Do(step criteriaStep) error        // one interaction step
	Text(selector string) (string, error)
	Close()
}

// openFunc launches a browser for one (url, profile) and returns a driven page.
// Swappable in tests. A navigation failure is reported via the page's NavError,
// not an error return; an error return means infra failure (browser/driver).
type openFunc func(ctx context.Context, url, profile string, logger *zap.Logger) (browserPage, error)

type RunChecksAction struct {
	logger *zap.Logger
	open   openFunc
}

func NewRunChecksAction(logger *zap.Logger) *RunChecksAction {
	return &RunChecksAction{logger: logger.Named("run_checks"), open: openChromium}
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

	for _, url := range req.URLs {
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
			res := evaluateOnPage(page, applicable, profile, url)
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
func evaluateOnPage(page browserPage, checks []criteriaCheck, profile, url string) []CheckResult {
	var results []CheckResult
	add := func(id string, pass bool, detail string) {
		results = append(results, CheckResult{CheckID: id, Profile: profile, URL: url, Pass: pass, Detail: detail})
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
			over, culprit, err := page.HorizontalOverflow()
			if err != nil {
				add(ch.ID, false, "could not measure overflow: "+err.Error())
			} else if over {
				// Name the offender: it decides whether this is the tool's bug or
				// the site template's (see HorizontalOverflow).
				detail := "page overflows horizontally (scrollWidth > clientWidth) on " + profile
				if culprit != "" {
					detail += "; widest offending element: " + culprit
				}
				add(ch.ID, false, detail)
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
// widest offending element: without that, one overflowing site footer raises an
// identical, unactionable improve_tool ticket for EVERY tool on the site — the
// fixer edits a tool that cannot possibly fix a template footer. Observed
// 2026-07-14: vonc.com's div.footer-legal (506px at a 390px viewport) failed
// the quiz's mobile-fit check on every page of the site.
func (c *chromiumPage) HorizontalOverflow() (bool, string, error) {
	v, err := c.page.Evaluate(`() => {
		const vw = document.documentElement.clientWidth;
		const over = document.documentElement.scrollWidth - vw;
		if (over <= 2) return {over: over, culprit: ""};
		// Widest element crossing the viewport edge; deepest wins ties, since an
		// ancestor is usually just inheriting an overflowing child's width.
		let best = null;
		for (const el of document.querySelectorAll('*')) {
			const r = el.getBoundingClientRect();
			if (r.width === 0 && r.height === 0) continue;
			if (r.right > vw + 1 || r.left < -1) {
				const depth = (function(n){ let d = 0; while (n.parentElement) { d++; n = n.parentElement; } return d; })(el);
				if (!best || r.width > best.width || (r.width === best.width && depth > best.depth)) {
					best = {
						width: Math.round(r.width),
						depth: depth,
						desc: el.tagName.toLowerCase()
						      + (el.id ? '#' + el.id : '')
						      + (el.className && el.className.toString().trim()
						         ? '.' + el.className.toString().trim().split(/\s+/).join('.')
						         : ''),
					};
				}
			}
		}
		return {over: over, culprit: best ? best.desc + ' (' + best.width + 'px)' : ""};
	}`)
	if err != nil {
		return false, "", err
	}
	m, ok := v.(map[string]interface{})
	if !ok {
		return false, "", nil
	}
	culprit, _ := m["culprit"].(string)
	// JS numbers come back as float64/int; tolerate 2px of rounding.
	switch n := m["over"].(type) {
	case float64:
		return n > 2, culprit, nil
	case int:
		return n > 2, culprit, nil
	default:
		return false, "", nil
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

func (c *chromiumPage) Close() {
	if c.browser != nil {
		_ = c.browser.Close()
	}
	if c.pw != nil {
		_ = c.pw.Stop()
	}
}
