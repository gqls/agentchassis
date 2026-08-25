// FILE: platform/orchestration/actions/discovery_checks/check_heading_promise.go
//
// Discovery check: heading_promise — "does each served page CONTAIN what its own
// headings say it does?"
//
// The "promise" seat of the site acceptance council (RFC_056). A page's own
// <h1>/<h2>/<h3> is a promise to the reader. "Month by month" promises a
// calendar; "checklist" promises a list; "comparison" promises a table; "Top 10"
// promises ten things. Every other signal the platform has — 200, 60KB, leak-free,
// fully linked, build_status deployed — is satisfied by a page that makes the
// promise and does not keep it. Byte count is not a completeness check, and
// neither is "it deployed".
//
// WHY IT EXISTS — the owner's review, 2026-08-25. Reading homegarden.uk by hand,
// the owner found a page headed "Garden maintenance for UK gardens, month by
// month" that served THREE month names in its content, and found it first — before
// any harness, before any check. The v2 acceptance harness
// (loanzy_uk_example_site/after_test.sh) gained a PROMISE vs DELIVERY section the
// same day; this file ports that section's rules, including the two corrections it
// had to make to itself before it stopped being wrong.
//
// ── CORRECTION 1: COUNTS ARE CHROME-STRIPPED AND NON-ANCHOR ─────────────────
//
// Measured 2026-08-25 (the harness's own note, re-measured here by curl the same
// day): EVERY page on homegarden.uk carries all twelve month names as MENU LINKS,
// so a raw month count read 12 on /contact.html — a page with no months in it at
// all — and would have passed the calendar page for the wrong reason. Anything
// that is link text is chrome or cross-reference; the promise has to be kept in
// NON-ANCHOR content. So every body measure below is taken after removing
// <style>, <script>, <nav>, the site's landmark <header>/<footer>, and every <a>.
// Re-measured on the live site under that rule: /garden/index.html 3 (raw 12),
// /contact.html 0 (raw 12), /january/index.html 4 (raw 12), / 12 (raw 12).
//
// "Landmark" header/footer, not every header/footer — a refinement of the port,
// measured rather than assumed. The motivating site uses <header> INSIDE content:
// on / the promise-bearing h2 "The garden and home year, month by month" sits in
// <header class="period-cal__header"> within <main>, alongside the calendar's own
// introductory text. Stripping every <header> would strip content the promise is
// kept in. The HTML spec's own rule is applied instead: a <header>/<footer> is the
// page's banner/contentinfo only when it is NOT a descendant of main, article,
// section or aside; scoped to one of those it is that section's header and stays.
//
// Headings, by contrast, are read from the whole document minus <nav>, <style> and
// <script>: a nav heading is a menu label, not a page promise, and would flag every
// page at once; a hero h1 inside a landmark <header> is a real page promise.
//
// ── CORRECTION 2: A KEYWORD IS NOT A PROMISE ────────────────────────────────
//
// This check fires on heading WORDS. A section index whose headings describe what
// its articles will do trips the rule without promising that structure on THAT
// page. Live example, measured 2026-08-25: /comparisons/index.html is headed "What
// each comparison covers" and has no table — correctly, it is the index of the
// comparisons, not one of them. The check therefore NOMINATES candidates; it does
// not adjudicate them. The work item says so in its spec, and the summary carries
// the heading verbatim so a reader can read it before acting.
//
// ── WHAT THIS CHECK DOES NOT DO ─────────────────────────────────────────────
//
//   - It does not ADJUDICATE. An item is a nomination for a planner or writer to
//     read; the rule cannot tell a promise from a description of one.
//   - It does not DISPATCH. HandlerAgent is empty and Status is 'detected': the
//     repair — write the calendar, build the table, cut the heading — is a
//     judgement about what the page is for, and no generator can make it.
//   - It does not judge on a PARKED DOMAIN. LANDMINES.md: "a parked domain 200s
//     every path", and a stub that 200s every path would print PROMISE UNMET for
//     every page — a screenful of confident findings about a 128-byte redirect.
//     Before any page is fetched, an INVENTED path is requested; if it serves 200
//     the run returns an error and files NOTHING, resolves NOTHING. Measured
//     2026-08-25: https://homegarden.uk/__acceptance-control-deadbeef.html → 404.
//   - It does not infer from ABSENCE. A page that could not be fetched (transport
//     error, non-2xx) is skipped with a logged reason: neither filed nor resolved.
//     Retraction fires ONLY on a positive observation — the page was fetched 2xx
//     and every promise-bearing heading is met, or it no longer carries one —
//     exactly the contract CheckResult.Resolved states.
//
// ── THE RULES (ported verbatim from after_test.sh:181-234, plus the top-N rule) ──
//
// A heading matches at most one rule; first match wins, in this order:
//
//	"month by month" | "month-by-month" | "calendar"           → ≥6 distinct month names
//	"step by step" | "step-by-step" | "checklist" | "steps to"  → ≥3 non-anchor <li>
//	"compare" | "comparison" | "side by side" | " vs " | " versus " → ≥1 <table>
//	(top|best|the) N <things>  or  N (tips|ways|steps|ideas|reasons), 3 ≤ N ≤ 50
//	                                                            → ≥N <li> OR ≥N card-token elements
//
// "Card-token" means an element whose class attribute, split on whitespace,
// contains the WHOLE token card, item or entry. BEM substrings (card__title,
// card-grid) do not count — ten titles are not ten cards.
//
// Item shape: ONE per unmet page, keyed heading_promise_unmet:<page_id>, flag-only.
// Registered under "heading_promise"; the item type is the literal string
// "heading_promise_unmet", classified in verifier_coverage_test.go.

package discovery_checks

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&HeadingPromiseCheck{}) }

type HeadingPromiseCheck struct{}

func (c *HeadingPromiseCheck) Name() string { return "heading_promise" }

const (
	// promiseMaxPages bounds the outbound fetches one site can cause per run. The
	// motivating site has 20 active pages; when the cap drops anything it LOGS
	// what it dropped, because a silent cap reads as "everything was checked".
	promiseMaxPages = 40

	// promiseFetchTimeout per page, the same 15s the apex probe uses: a healthy
	// site answers in well under it, and a proxy waiting on a dead origin is the
	// case a longer wait would only confirm more slowly.
	promiseFetchTimeout = 15 * time.Second

	// promiseBodyCap bounds what one fetch will read. The motivating pages are
	// 60–78KB; the cap exists so a pathological origin cannot make the check
	// buffer without limit.
	promiseBodyCap = 2 * 1024 * 1024

	// promiseHeadingMinChars / MaxChars — the harness's own bounds on a heading
	// worth reading (`[^<]{3,90}`): shorter is a glyph, longer is a paragraph.
	promiseHeadingMinChars = 3
	promiseHeadingMaxChars = 90
)

// promiseSiteBaseURL is where a site's pages are served from. Swappable so the
// tests can point the real fetch path at an httptest server; in production it is
// the https origin of the recorded domain, and every page URL is RESOLVED against
// it with net/url — never composed by string concatenation, because a verdict
// from a URL you built is a verdict about a page you invented.
var promiseSiteBaseURL = func(domain string) string { return "https://" + domain }

// promiseMeasures is what the non-anchor, chrome-stripped body of one page
// actually contains. Pure data so the rule table can be tested without HTML.
type promiseMeasures struct {
	Months    int // distinct month names, whole words, outside anchors
	ListItems int // <li> whose non-anchor text is non-empty
	Tables    int // <table>
	Cards     int // elements carrying a whole class token card|item|entry
}

// promiseRule is what a heading promises, and how much of it the body must show.
type promiseRule struct {
	Name   string // calendar | checklist | comparison | top_n
	Needed int
	Text   string // for the summary: "a month-by-month calendar"
}

// promiseMonthRe: whole words, capitalised as month names are written. The
// harness's own pattern; "may" the verb does not count, "May" does — the same
// limitation, stated rather than fixed, because fixing it needs a parser of
// English rather than of HTML.
var promiseMonthRe = regexp.MustCompile(`\b(January|February|March|April|May|June|July|August|September|October|November|December)\b`)

// promiseTopNRe / promiseCountNounRe: the two spellings of "N things" a heading
// uses. Digits are captured unbounded and range-checked afterwards so that "the
// 2024 guide" parses as 2024 and is rejected, rather than as "202" and accepted.
var (
	promiseTopNRe      = regexp.MustCompile(`\b(?:top|best|the)\s+(\d+)\s+\pL`)
	promiseCountNounRe = regexp.MustCompile(`\b(\d+)\s+(?:tips|ways|steps|ideas|reasons)\b`)
)

// promiseRuleFor maps a LOWER-CASED heading to the rule it trips, if any. A
// heading matches at most one rule; first match wins, in the harness's order.
// This is the nomination step — see CORRECTION 2 in the header.
func promiseRuleFor(lower string) (promiseRule, bool) {
	has := func(needles ...string) bool {
		for _, n := range needles {
			if strings.Contains(lower, n) {
				return true
			}
		}
		return false
	}
	switch {
	case has("month by month", "month-by-month", "calendar"):
		return promiseRule{Name: "calendar", Needed: 6, Text: "a month-by-month calendar (>=6 distinct month names)"}, true
	case has("step by step", "step-by-step", "checklist", "steps to"):
		return promiseRule{Name: "checklist", Needed: 3, Text: "a checklist or steps (>=3 list items)"}, true
	case has("compare", "comparison", "side by side", " vs ", " versus "):
		return promiseRule{Name: "comparison", Needed: 1, Text: "a comparison (>=1 table)"}, true
	}
	for _, re := range []*regexp.Regexp{promiseTopNRe, promiseCountNounRe} {
		m := re.FindStringSubmatch(lower)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil || n < 3 || n > 50 {
			continue
		}
		return promiseRule{Name: "top_n", Needed: n,
			Text: fmt.Sprintf("%d things (>=%d list items or >=%d card elements)", n, n, n)}, true
	}
	return promiseRule{}, false
}

// met says whether the body keeps the promise. The default arm is deliberately
// false with a name in it: a rule this function does not know is not "met".
func (m promiseMeasures) met(r promiseRule) bool {
	switch r.Name {
	case "calendar":
		return m.Months >= r.Needed
	case "checklist":
		return m.ListItems >= r.Needed
	case "comparison":
		return m.Tables >= r.Needed
	case "top_n":
		return m.ListItems >= r.Needed || m.Cards >= r.Needed
	}
	return false
}

// describe is the measured half of the summary, phrased for the rule that failed.
func (m promiseMeasures) describe(r promiseRule) string {
	switch r.Name {
	case "calendar":
		return fmt.Sprintf("%d distinct month name(s) in non-anchor content", m.Months)
	case "checklist":
		return fmt.Sprintf("%d non-anchor list item(s)", m.ListItems)
	case "comparison":
		return fmt.Sprintf("%d table(s)", m.Tables)
	case "top_n":
		return fmt.Sprintf("%d non-anchor list item(s) and %d card element(s)", m.ListItems, m.Cards)
	}
	return fmt.Sprintf("months=%d list_items=%d tables=%d cards=%d", m.Months, m.ListItems, m.Tables, m.Cards)
}

// promiseText is the text of a selection with a SPACE between text nodes.
// goquery's own Text() concatenates adjacent text nodes with nothing between
// them, so "<h2>month by month</h2><p>January" reads "month by monthJanuary" and
// a whole-word match fails at every element boundary — found 2026-08-25 when the
// section-header test measured 4 months out of 6. The harness never met this:
// grep ran over the raw HTML, where the tags themselves were the separators.
func promiseText(sel *goquery.Selection) string {
	var b strings.Builder
	var walk func(*goquery.Selection)
	walk = func(s *goquery.Selection) {
		s.Contents().Each(func(_ int, c *goquery.Selection) {
			if goquery.NodeName(c) == "#text" {
				b.WriteString(c.Text())
				b.WriteByte(' ')
				return
			}
			walk(c)
		})
	}
	walk(sel)
	return b.String()
}

// promiseHeading is one <h1>-<h3> worth reading, with the rule it trips (if any).
type promiseHeading struct {
	Text string // as served, whitespace-collapsed
	Rule promiseRule
	Has  bool
}

// promiseReadPage parses one served page into its headings and its body
// measures. It mutates the document it parses, in this order, because the
// headings must be read before the body is stripped:
//
//  1. remove <style>, <script>, <nav>  — never content, on any surface;
//  2. read h1/h2/h3 from what remains;
//  3. remove LANDMARK <header>/<footer> (not scoped to main/article/section/aside);
//  4. remove every <a>;
//  5. measure what is left.
func promiseReadPage(body []byte) ([]promiseHeading, promiseMeasures, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, promiseMeasures{}, err
	}
	doc.Find("style, script, nav").Remove()

	var headings []promiseHeading
	doc.Find("h1, h2, h3").Each(func(_ int, s *goquery.Selection) {
		text := strings.Join(strings.Fields(promiseText(s)), " ")
		n := utf8.RuneCountInString(text)
		if n < promiseHeadingMinChars || n > promiseHeadingMaxChars {
			return
		}
		h := promiseHeading{Text: text}
		h.Rule, h.Has = promiseRuleFor(strings.ToLower(text))
		headings = append(headings, h)
	})

	scope := doc.Find("body")
	scope.Find("header, footer").Each(func(_ int, s *goquery.Selection) {
		if s.ParentsFiltered("main, article, section, aside").Length() == 0 {
			s.Remove()
		}
	})
	scope.Find("a").Remove()

	var m promiseMeasures
	seen := map[string]bool{}
	for _, match := range promiseMonthRe.FindAllString(promiseText(scope), -1) {
		seen[match] = true
	}
	m.Months = len(seen)
	scope.Find("li").Each(func(_ int, s *goquery.Selection) {
		if strings.TrimSpace(s.Text()) != "" {
			m.ListItems++
		}
	})
	m.Tables = scope.Find("table").Length()
	scope.Find("[class]").Each(func(_ int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		for _, tok := range strings.Fields(class) {
			switch strings.ToLower(tok) {
			case "card", "item", "entry":
				m.Cards++
				return
			}
		}
	})
	return headings, m, nil
}

// promiseFetchResult is one GET's observation. Err set means no HTTP
// conversation happened; it is not a status and is never compared to one.
type promiseFetchResult struct {
	Status int
	Body   []byte
	Err    error
}

// promiseGET fetches a page the way a visitor would — GET, redirects followed,
// bounded by promiseFetchTimeout on both the context and the client.
func promiseGET(ctx context.Context, target string) promiseFetchResult {
	cctx, cancel := context.WithTimeout(ctx, promiseFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, target, nil)
	if err != nil {
		return promiseFetchResult{Err: fmt.Errorf("build request: %w", err)}
	}
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+heading_promise)")
	req.Header.Set("Accept", "text/html,*/*")
	client := &http.Client{Timeout: promiseFetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return promiseFetchResult{Err: err}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, promiseBodyCap))
	return promiseFetchResult{Status: resp.StatusCode, Body: body}
}

// promiseControlPath is the invented path of the parked-domain control. Random
// per run so a catch-all cannot have learned it, and so two concurrent runs
// cannot be told apart in an access log by accident.
func promiseControlPath() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("random control path: %w", err)
	}
	return "__acceptance-control-" + hex.EncodeToString(b) + ".html", nil
}

// promisePage is one active page as recorded — the URL it is served at, never one
// this check composed.
type promisePage struct {
	ID  uuid.UUID
	URL string
}

func (c *HeadingPromiseCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	result := &CheckResult{}

	var domain string
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, '') FROM sites WHERE id = $1`, dctx.SiteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("heading_promise: load site: %w", err)
	}
	if domain == "" {
		dctx.Logger.Info("heading_promise: site has no domain; nothing is served, nothing to read",
			zap.String("site_id", dctx.SiteID.String()))
		return result, nil
	}

	// The page set uses the estate's shared lifecycle predicate, not a hand-rolled
	// status filter: `pages.status` has more than one live spelling (LANDMINES:
	// "a pages query that filters on status may be filtering on NOTHING"), and a
	// flag-only, self-clearing check that silently selects zero pages reads as
	// "nothing to report" — the wrong direction. Council round d1342f2a
	// (debug_historian) caught the raw spelling here; check_structure_floor had
	// the predicate from birth.
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id, p.url FROM pages p
		WHERE p.site_id = $1 AND `+datahelpers.PageWantedLivePredicateFor("p")+` AND COALESCE(p.url, '') <> ''
		ORDER BY p.url, p.id`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("heading_promise: load pages: %w", err)
	}
	var pages []promisePage
	for rows.Next() {
		var p promisePage
		if err := rows.Scan(&p.ID, &p.URL); err != nil {
			rows.Close()
			return nil, fmt.Errorf("heading_promise: scan page: %w", err)
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("heading_promise: load pages: %w", err)
	}
	rows.Close()
	if len(pages) == 0 {
		return result, nil
	}

	if len(pages) > promiseMaxPages {
		dropped := make([]string, 0, len(pages)-promiseMaxPages)
		for _, p := range pages[promiseMaxPages:] {
			dropped = append(dropped, p.URL)
		}
		dctx.Logger.Warn("heading_promise: page cap reached — these pages were NOT read",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", promiseMaxPages),
			zap.Int("dropped", len(dropped)),
			zap.Strings("dropped_urls", dropped))
		pages = pages[:promiseMaxPages]
	}

	base, err := url.Parse(strings.TrimSuffix(promiseSiteBaseURL(domain), "/") + "/")
	if err != nil {
		return nil, fmt.Errorf("heading_promise: site base url for %q: %w", domain, err)
	}

	// PARKED-DOMAIN CONTROL, before any page is read. An origin that serves an
	// invented path 200 would serve every page 200 too, and every one of them
	// would read UNMET. That is not a finding about the site; it is blindness,
	// and it is returned as an error so the runner files nothing and retracts
	// nothing (registry.go: Resolved is skipped entirely when Run errors).
	controlPath, err := promiseControlPath()
	if err != nil {
		return nil, fmt.Errorf("heading_promise: %w", err)
	}
	controlURL := base.ResolveReference(&url.URL{Path: controlPath}).String()
	control := promiseGET(dctx.Ctx, controlURL)
	switch {
	case control.Err != nil:
		return nil, fmt.Errorf("heading_promise: blinded: control fetch %s failed (%v); cannot judge, filing nothing",
			controlURL, control.Err)
	case control.Status >= 200 && control.Status <= 299:
		return nil, fmt.Errorf("heading_promise: blinded: invented path %s served %d — parked domain or catch-all; filing nothing",
			controlURL, control.Status)
	}

	for _, p := range pages {
		ref, err := url.Parse(p.URL)
		if err != nil {
			dctx.Logger.Info("heading_promise: recorded page url does not parse, skipped",
				zap.String("page_id", p.ID.String()), zap.String("url", p.URL), zap.Error(err))
			result.Findings = append(result.Findings, promiseSkipFinding(p, "unparseable_url: "+err.Error()))
			continue
		}
		target := base.ResolveReference(ref).String()

		fetched := promiseGET(dctx.Ctx, target)
		if fetched.Err != nil {
			dctx.Logger.Info("heading_promise: page fetch failed, skipped (neither filed nor resolved)",
				zap.String("page_id", p.ID.String()), zap.String("url", target), zap.Error(fetched.Err))
			result.Findings = append(result.Findings, promiseSkipFinding(p, "transport_error: "+fetched.Err.Error()))
			continue
		}
		if fetched.Status < 200 || fetched.Status > 299 {
			dctx.Logger.Info("heading_promise: page not 2xx, skipped (neither filed nor resolved)",
				zap.String("page_id", p.ID.String()), zap.String("url", target), zap.Int("status", fetched.Status))
			result.Findings = append(result.Findings, promiseSkipFinding(p, fmt.Sprintf("http_%d", fetched.Status)))
			continue
		}

		headings, measures, err := promiseReadPage(fetched.Body)
		if err != nil {
			dctx.Logger.Warn("heading_promise: served html did not parse, skipped",
				zap.String("page_id", p.ID.String()), zap.String("url", target), zap.Error(err))
			result.Findings = append(result.Findings, promiseSkipFinding(p, "unparseable_html: "+err.Error()))
			continue
		}

		var promises []map[string]interface{}
		var unmet []promiseHeading
		for _, h := range headings {
			if !h.Has {
				continue
			}
			ok := measures.met(h.Rule)
			promises = append(promises, map[string]interface{}{
				"heading": h.Text, "rule": h.Rule.Name, "needed": h.Rule.Needed, "met": ok,
			})
			if !ok {
				unmet = append(unmet, h)
			}
		}
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":       "heading_promise",
			"page_id":     p.ID.String(),
			"url":         target,
			"http_status": fetched.Status,
			"months":      measures.Months,
			"list_items":  measures.ListItems,
			"tables":      measures.Tables,
			"cards":       measures.Cards,
			"promises":    promises,
			"unmet":       len(unmet) > 0,
		})

		itemKey := fmt.Sprintf("heading_promise_unmet:%s", p.ID)
		if len(unmet) == 0 {
			// A POSITIVE observation: the page was served 2xx and either keeps
			// every promise its headings make or no longer makes one. Narrow —
			// this page's key only, never AllOfType.
			reason := "re-read: no promise-bearing heading on the served page"
			if len(promises) > 0 {
				reason = fmt.Sprintf("re-read: all %d promise-bearing heading(s) met (%s)",
					len(promises), measures.describe(promiseRule{}))
			}
			result.Resolved = append(result.Resolved, ResolvedFinding{
				ItemType: "heading_promise_unmet",
				ItemKey:  itemKey,
				Reason:   reason,
			})
			continue
		}

		first := unmet[0]
		unmetSpec := make([]map[string]interface{}, 0, len(unmet))
		for _, h := range unmet {
			unmetSpec = append(unmetSpec, map[string]interface{}{
				"heading": h.Text, "rule": h.Rule.Name, "needed": h.Rule.Needed, "rule_text": h.Rule.Text,
			})
		}
		spec := map[string]interface{}{
			"check":     "heading_promise",
			"seat":      "promise",
			"rfc":       "RFC_056",
			"page_id":   p.ID.String(),
			"url":       target,
			"heading":   first.Text,
			"rule":      first.Rule.Name,
			"rule_text": first.Rule.Text,
			"needed":    first.Rule.Needed,
			"measured": map[string]int{
				"months": measures.Months, "list_items": measures.ListItems,
				"tables": measures.Tables, "cards": measures.Cards,
			},
			"unmet_headings": unmetSpec,
			// CORRECTION 2, carried on the row so it travels with the finding.
			"nominates_not_adjudicates": "a keyword is not a promise — read the heading before acting",
			// Same wording as remit.go's capability_gap: this row is a verdict to
			// read, not a dispatch, and promoting it hands a fixer a judgement.
			"not_dispatchable": "status 'detected' + empty handler_agent — deliberate; " +
				"promoting this row dispatches work no handler can do (bugs_open/077)",
		}
		specJSON, _ := json.Marshal(spec)

		pageID := p.ID
		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:   dctx.SiteID,
			PageID:   &pageID,
			Source:   "discovery",
			Pipeline: dctx.Pipeline,
			ItemType: "heading_promise_unmet",
			Severity: "medium",
			Summary: fmt.Sprintf("%s: heading '%s' promises %s but the body has %s",
				target, first.Text, first.Rule.Text, measures.describe(first.Rule)),
			SpecJSON:     string(specJSON),
			Priority:     110,
			HandlerAgent: "", // flag-only: the repair is a planner/writer judgement
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      itemKey,
			BatchID:      dctx.BatchID,
		})
	}

	return result, nil
}

// promiseSkipFinding records a page this run did NOT judge, so a run that was
// blinded on some pages cannot be read as a clean bill for them.
func promiseSkipFinding(p promisePage, reason string) map[string]interface{} {
	return map[string]interface{}{
		"check":   "heading_promise",
		"page_id": p.ID.String(),
		"url":     p.URL,
		"skipped": reason,
	}
}
