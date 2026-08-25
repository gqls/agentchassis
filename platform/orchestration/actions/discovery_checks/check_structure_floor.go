// FILE: platform/orchestration/actions/discovery_checks/check_structure_floor.go
//
// Discovery check: structure_floor — the STRUCTURE seat of the site acceptance
// council (RFC_056; design in docs/agent_docs/docs024_key_docs_latest/
// loanzy_uk_example_site/REFERENCE_2026-08-25_site_acceptance_council.md §4c).
//
// OWNER RULING, 2026-08-25, verbatim in substance: "implement at least 6
// different structures unless the site doesn't warrant it or refuses it."
// N = 6 is the owner's number (structureFloorDefault), overridable per site at
// sites.settings->'maintenance_profile'->'structure_floor'->>'n'.
//
// ── WHY A COUNT AND NOT A CHECKLIST ─────────────────────────────────────────
//
// A checklist ("has a directory", "has a calendar") collides with the vertical
// classifier's taxonomy and is wrong for every site it was not written for. A
// count is vertical-agnostic: it does not care WHICH structures a site delivers,
// only that a reader gets more than one shape of thing. It is ONE number — the
// RFC_022 governance shape this estate has already accepted (an optional-key
// budget of N = 10, ruled 2026-08-14) — so the ruling is a threshold, not a
// prompt clause. And it forces the planner to EXPLORE a 47-component vocabulary
// that most sites use three of: below the floor the planner cannot pick the
// same three again and pass.
//
// THE MOTIVATING MEASUREMENT [MEASURED 2026-08-25]: homegarden.uk, 21 pages,
// 17 of them identical — hero + generic-text-block + content-listing. One
// structure delivered, seventeen times over. Every internal signal read green.
//
// ── "STRUCTURE", AND "DELIVERED" ─────────────────────────────────────────────
//
// A structure is defined by WHAT THE READER GETS, not by component name: a
// section called `seasonal-calendar` that renders three paragraphs is not a
// calendar. Hero, nav, footer and CTA do not count — every site has them and
// they carry nothing. Each structure counts AT MOST ONCE per site, and only if
// it is delivered on at least one SERVED page: a live GET of the recorded
// pages.url resolved against https://<domain>, never rendered_html. The reader
// gets the wire, not the row.
//
// Content-scoped: <style>, <script>, <nav>, <header> and <footer> are stripped
// before anything is counted. "Non-anchor text" is text not inside an <a>, so
// twelve month names that are all nav links are twelve links, not a calendar
// (TestStructureFloor_MonthNamesInNavLinksAreNotACalendar). Class matching is
// on WHOLE class tokens split on whitespace: `card__title` is not `card`, so a
// BEM element cannot count as the block it belongs to.
//
// ── V1 RUBRIC — a FIRST rubric; the RFC asks the architecture seat to rule on the shapes ──
//
//	list        a <ul>/<ol> with ≥5 <li> whose non-anchor text is non-empty
//	table       a <table> with ≥3 <tr>
//	calendar    class token `period-cal` or `calendar` present, OR ≥6 distinct
//	            month names (January…December, whole words) in one page's
//	            non-anchor text
//	checklist   ≥3 input[type=checkbox], OR class token `checklist` holding ≥3 <li>
//	comparison  a <table> with ≥3 columns (max <td>/<th> per row) AND an h1–h3 on
//	            that page containing compare|comparison|vs|versus
//	tool        a <form> (or class token `tool`) holding ≥2 input/select/textarea
//	            AND a <button>
//	directory   ≥6 sibling elements with class token card|item|entry|listing, each
//	            holding an <a href> to an absolute http(s) URL on another host
//	feed        class token `news`|`feed` holding ≥3 children that each carry a
//	            <time> or an external <a href>
//	guide       a page with ≥4 <h2> in content AND ≥600 words of non-anchor text
//	faq         ≥3 <details>, OR class token `faq` holding ≥3 <li> or ≥3 <h3>
//
// One reading the rubric leaves open, taken here and stated so the architecture
// seat can overrule it: `tool` counts only fields a reader can OPERATE —
// input[type=hidden|submit|button|reset|image] are not fields — so a newsletter
// box carrying a CSRF token and a button is not a tool. This only removes false
// positives; a real tool still has two operable fields
// (TestStructureFloor_ToolNeedsOperableFieldsAndAButton). Month names are
// matched Title-case or ALL-CAPS, so "you may" is not May.
//
// ── A REFUSAL IS A RECORDED VERDICT, NOT A SILENT OPTION ───────────────────
//
// "Unless the site doesn't warrant it or refuses it" is the second half of the
// ruling, and it is a RECORD, not an exemption a planner takes quietly. A
// non-empty sites.settings->'maintenance_profile'->'structure_floor'->>'refusal'
// makes this seat file nothing, put the refusal text in its Findings, and
// retract any open item (Resolved, reason "refusal recorded: …"). Refusals are
// then auditable: rare and reasoned means the floor works; common and thin means
// the planner prompt is wrong. The count is still measured and recorded beside
// the refusal, so a refusal never hides the number it refuses.
//
// ── WHAT IT DOES NOT DO ─────────────────────────────────────────────────────
//
//   - It does not measure DEPTH, copy or imagery. Breadth only; the depth seat
//     sits beside it, and neither touches prose or pictures.
//   - It does not DISPATCH. Below the floor it RECORDS: one flag-only item per
//     site (empty handler_agent, status 'detected', never promoted — the
//     promoter and triage both require a handler). The refusal is a
//     planner/human verdict, not something a fixer produces, so there is
//     nothing to route and the spec says so in remit.go's words.
//   - It does not JUDGE ON A PARKED DOMAIN. A registrar parking page answers
//     200 to EVERY path (LANDMINES.md: "a parked domain 200s every path"), so
//     every page would "serve" and the count would be a verdict about the
//     parking page. Before any page is fetched the seat GETs an invented path,
//     /__acceptance-control-<8 hex>.html; a 2xx there means the seat is
//     blinded and it returns an ERROR — no item, no retraction. The same holds
//     when fewer than one page fetches 2xx: no verdict, an error.
//
// Self-clearing per RFC_010 (CheckResult.Resolved): at or above N, or on a
// recorded refusal, the seat retracts its own item by key on a POSITIVE
// observation only. An errored run retracts nothing — the runner skips Resolved
// on error, and this file never populates it on a blind run.
//
// Enablement: `structure_floor` in a discovery agent's checks array, held until
// this file's image has rolled — the runner hard-fails on a name the binary does
// not register.

package discovery_checks

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/net/html"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&StructureFloorCheck{}) }

type StructureFloorCheck struct{}

func (c *StructureFloorCheck) Name() string { return "structure_floor" }

const (
	// structureFloorDefault is the owner's N (ruling 2026-08-25). Per-site
	// override: sites.settings->'maintenance_profile'->'structure_floor'->>'n'.
	structureFloorDefault = 6

	// structureFloorPageCap bounds the outbound GETs one site can cause. When it
	// drops anything it LOGS what it dropped — a silent cap reads as "every page
	// was looked at" when it was not.
	structureFloorPageCap = 40

	// structureFloorFetchTimeout per GET; the same 15s the apex probe uses.
	structureFloorFetchTimeout = 15 * time.Second

	// structureFloorBodyCap bounds one page read (2 MB, as fetchDeployedPage).
	structureFloorBodyCap = 2 << 20

	// structureFloorRefusalPath is where a planner or a human records "this site
	// does not warrant / refuses the floor". Carried in the spec so the reader
	// of an item knows where the verdict goes.
	structureFloorRefusalPath = "sites.settings->'maintenance_profile'->'structure_floor'->>'refusal'"
)

// structureFloorHTTPClient is the one seam a test swaps (for an httptest TLS
// server's client). URL building, status handling and the control probe stay
// real code, exercised by the tests rather than stubbed around.
var structureFloorHTTPClient = &http.Client{Timeout: structureFloorFetchTimeout}

// structureMonthRe matches a month name as a whole word, Title-case or
// ALL-CAPS. Deliberately not case-insensitive: "you may" and "march on" are
// English, not a calendar.
var structureMonthRe = regexp.MustCompile(`\b(?:January|February|March|April|May|June|July|August|September|October|November|December|` +
	`JANUARY|FEBRUARY|MARCH|APRIL|MAY|JUNE|JULY|AUGUST|SEPTEMBER|OCTOBER|NOVEMBER|DECEMBER)\b`)

// structureComparisonRe is the heading vocabulary that turns a wide table into
// a comparison. Whole words, so "compared" and "comparing" do not match — the
// rubric names four forms and the architecture seat can widen it.
var structureComparisonRe = regexp.MustCompile(`(?i)\b(?:compare|comparison|vs|versus)\b`)

// structureFloorItemKey is the dedup key: one verdict row per site.
func structureFloorItemKey(siteID uuid.UUID) string {
	return fmt.Sprintf("structure_floor_unmet:%s", siteID)
}

type structurePage struct {
	ID  uuid.UUID
	URL string
}

func (c *StructureFloorCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	var domain string
	var settingsRaw []byte
	if err := dctx.DB.QueryRowContext(dctx.Ctx,
		`SELECT COALESCE(domain, ''), settings FROM sites WHERE id = $1`, dctx.SiteID).
		Scan(&domain, &settingsRaw); err != nil {
		return nil, fmt.Errorf("structure_floor: load site: %w", err)
	}
	if domain == "" {
		// No domain, nothing served, nothing to judge. An error rather than an
		// empty result: an empty result reads as "looked and found nothing".
		return nil, fmt.Errorf("structure_floor: site %s has no domain, so nothing is served and nothing can be judged", dctx.SiteID)
	}

	n, nSource, refusal, err := structureFloorSettings(settingsRaw)
	if err != nil {
		return nil, fmt.Errorf("structure_floor: sites.settings for %s: %w", dctx.SiteID, err)
	}
	if nSource == "invalid" {
		dctx.Logger.Warn("structure_floor: settings carry an unusable structure_floor.n — using the default",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("default_n", structureFloorDefault))
	}

	// PARKED-DOMAIN CONTROL FIRST. A parking page 200s every path, so a 2xx
	// from a page proves nothing until an invented path has been seen to 404.
	controlURL, err := structureControlURL(domain)
	if err != nil {
		return nil, fmt.Errorf("structure_floor: build control url: %w", err)
	}
	controlStatus, _, err := structureFetch(dctx.Ctx, controlURL)
	if err != nil {
		return nil, fmt.Errorf("structure_floor: parked-domain control %s could not be fetched, so the seat cannot tell the site from a parking page: %w", controlURL, err)
	}
	if controlStatus >= 200 && controlStatus <= 299 {
		dctx.Logger.Warn("structure_floor: BLINDED — an invented path answered 2xx; the domain is parked or rewrites every path, no verdict",
			zap.String("site_id", dctx.SiteID.String()),
			zap.String("domain", domain),
			zap.String("control_url", controlURL),
			zap.Int("status", controlStatus))
		return nil, fmt.Errorf("structure_floor: blinded — invented path %s answered HTTP %d, so a 2xx from %s proves nothing about the site; no verdict", controlURL, controlStatus, domain)
	}

	pages, err := structureLoadPages(dctx)
	if err != nil {
		return nil, err
	}
	overCap := 0
	if len(pages) > structureFloorPageCap {
		dropped := make([]string, 0, len(pages)-structureFloorPageCap)
		for _, p := range pages[structureFloorPageCap:] {
			dropped = append(dropped, p.URL)
		}
		overCap = len(dropped)
		dctx.Logger.Warn("structure_floor: page cap reached — these pages were NOT fetched and cannot contribute a structure",
			zap.String("site_id", dctx.SiteID.String()),
			zap.Int("cap", structureFloorPageCap),
			zap.Int("dropped", overCap),
			zap.Strings("dropped_urls", dropped))
		pages = pages[:structureFloorPageCap]
	}

	base, err := url.Parse("https://" + domain + "/")
	if err != nil {
		return nil, fmt.Errorf("structure_floor: domain %q does not form a URL: %w", domain, err)
	}

	// delivered maps each structure to the FIRST served URL it was seen on —
	// the evidence a reader of the item can open.
	delivered := map[string]string{}
	fetched, skipped := 0, 0
	for _, p := range pages {
		ref, err := url.Parse(p.URL)
		if err != nil {
			skipped++
			dctx.Logger.Info("structure_floor: pages.url does not parse, skipped",
				zap.String("page_id", p.ID.String()), zap.String("url", p.URL), zap.Error(err))
			continue
		}
		target := base.ResolveReference(ref).String()
		status, body, err := structureFetch(dctx.Ctx, target)
		if err != nil {
			skipped++
			dctx.Logger.Info("structure_floor: page fetch failed, skipped",
				zap.String("page_id", p.ID.String()), zap.String("url", target), zap.Error(err))
			continue
		}
		if status < 200 || status > 299 {
			skipped++
			dctx.Logger.Info("structure_floor: page did not serve 2xx, skipped",
				zap.String("page_id", p.ID.String()), zap.String("url", target), zap.Int("status", status))
			continue
		}
		found, err := structureDetect(body, domain)
		if err != nil {
			skipped++
			dctx.Logger.Warn("structure_floor: served body could not be parsed, skipped",
				zap.String("page_id", p.ID.String()), zap.String("url", target), zap.Error(err))
			continue
		}
		fetched++
		for _, s := range found {
			if _, seen := delivered[s]; !seen {
				delivered[s] = target
			}
		}
	}
	if fetched == 0 {
		// Not a verdict. A site with no served page is not "zero structures";
		// it is unjudgeable, and an item here would be filed against nothing.
		return nil, fmt.Errorf("structure_floor: none of %d page(s) on %s fetched 2xx (%d skipped); cannot judge", len(pages), domain, skipped)
	}

	deliveredList := make([]string, 0, len(delivered))
	for s := range delivered {
		deliveredList = append(deliveredList, s)
	}
	sort.Strings(deliveredList)
	count := len(deliveredList)
	key := structureFloorItemKey(dctx.SiteID)

	finding := map[string]interface{}{
		"check":          "structure_floor",
		"seat":           "structure",
		"domain":         domain,
		"n":              n,
		"n_source":       nSource,
		"count":          count,
		"delivered":      deliveredList,
		"evidence":       delivered,
		"pages_fetched":  fetched,
		"pages_skipped":  skipped,
		"pages_over_cap": overCap,
	}
	result := &CheckResult{}

	if refusal != "" {
		// A recorded refusal is the ruling's own escape hatch, exercised on the
		// record. The count stays in the finding beside it.
		finding["refused"] = refusal
		result.Findings = append(result.Findings, finding)
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType: "structure_floor_unmet",
			ItemKey:  key,
			Reason:   "refusal recorded: " + refusal,
		})
		dctx.Logger.Info("structure_floor: refusal recorded for this site; nothing filed",
			zap.String("site_id", dctx.SiteID.String()),
			zap.String("domain", domain),
			zap.Int("count", count), zap.Int("n", n),
			zap.String("refusal", refusal))
		return result, nil
	}
	result.Findings = append(result.Findings, finding)

	if count >= n {
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType: "structure_floor_unmet",
			ItemKey:  key,
			Reason:   fmt.Sprintf("%d of %d delivered", count, n),
		})
		return result, nil
	}

	listText := "none"
	if count > 0 {
		listText = strings.Join(deliveredList, ", ")
	}
	spec := map[string]interface{}{
		"check":         "structure_floor",
		"seat":          "structure",
		"rfc":           "RFC_056",
		"n":             n,
		"n_source":      nSource,
		"count":         count,
		"delivered":     deliveredList,
		"evidence":      delivered,
		"pages_fetched": fetched,
		"pages_skipped": skipped,
		"refusal_path":  structureFloorRefusalPath,
		// Read by whoever picks this up, and by nobody automatically: this row
		// is a verdict, not a dispatch — remit.go's wording, adapted for a row
		// that is 'detected' rather than 'deferred'.
		"not_dispatchable": "empty handler_agent — deliberate; this row is a verdict, not a dispatch: " +
			"promoting it dispatches work no handler can do (bugs_open/077). Below the floor the planner " +
			"adds structures or a human records a refusal at refusal_path",
	}
	if overCap > 0 {
		spec["pages_over_cap"] = overCap
	}
	// An encoder, not json.Marshal: Marshal HTML-escapes ">" to \u003e, and
	// refusal_path is a jsonb path a human copies into SQL — it must read "->>".
	var specBuf bytes.Buffer
	specEnc := json.NewEncoder(&specBuf)
	specEnc.SetEscapeHTML(false)
	if err := specEnc.Encode(spec); err != nil {
		return nil, fmt.Errorf("structure_floor: encode spec: %w", err)
	}
	specJSON := strings.TrimSpace(specBuf.String())

	dctx.Logger.Warn("structure_floor: below the floor",
		zap.String("site_id", dctx.SiteID.String()),
		zap.String("domain", domain),
		zap.Int("count", count), zap.Int("n", n),
		zap.Strings("delivered", deliveredList),
		zap.Int("pages_fetched", fetched), zap.Int("pages_skipped", skipped))

	result.WorkItems = append(result.WorkItems, WorkItemSpec{
		SiteID:   dctx.SiteID,
		Source:   "discovery",
		Pipeline: dctx.Pipeline,
		ItemType: "structure_floor_unmet",
		Severity: "medium",
		Summary: fmt.Sprintf("%d of %d reader-facing structures delivered across %d pages: %s",
			count, n, fetched, listText),
		SpecJSON:     specJSON,
		Priority:     115,
		HandlerAgent: "", // flag-only: the seat RECORDS; the refusal is a planner/human verdict
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      key,
		BatchID:      dctx.BatchID,
	})
	return result, nil
}

// structureLoadPages lists the pages the platform still WANTS served and that
// carry a recorded URL to resolve. The lifecycle arm is the shared builder, not
// a fresh spelling (page_lifecycle_posture_test.go). No build-axis arm: whether
// a page has shipped is established by the GET itself — a never-built page 404s
// and is skipped, and a URL this check composed would be a verdict about a page
// it invented.
func structureLoadPages(dctx DiscoveryCheckContext) ([]structurePage, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id, p.url
		FROM pages p
		WHERE p.site_id = $1
		  AND `+datahelpers.PageWantedLivePredicateFor("p")+`
		  AND COALESCE(p.url, '') <> ''
		ORDER BY p.url
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("structure_floor: load pages: %w", err)
	}
	defer rows.Close()

	var pages []structurePage
	for rows.Next() {
		var p structurePage
		if err := rows.Scan(&p.ID, &p.URL); err != nil {
			return nil, fmt.Errorf("structure_floor: scan page: %w", err)
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("structure_floor: load pages: %w", err)
	}
	return pages, nil
}

// structureFloorSettings reads the per-site N override and the refusal. nSource
// is "default", "settings", or "invalid" — present but unusable, in which case
// the default is used and the CALLER logs it, because an override that is
// silently ignored is a landmine for whoever set it.
func structureFloorSettings(raw []byte) (n int, nSource string, refusal string, err error) {
	n, nSource = structureFloorDefault, "default"
	if len(strings.TrimSpace(string(raw))) == 0 {
		return n, nSource, "", nil
	}
	var settings map[string]interface{}
	if err := json.Unmarshal(raw, &settings); err != nil {
		return 0, "", "", fmt.Errorf("not a JSON object: %w", err)
	}
	profile, _ := settings["maintenance_profile"].(map[string]interface{})
	floor, _ := profile["structure_floor"].(map[string]interface{})
	if floor == nil {
		return n, nSource, "", nil
	}
	if r, ok := floor["refusal"].(string); ok {
		refusal = strings.TrimSpace(r)
	}
	switch v := floor["n"].(type) {
	case nil:
	case float64:
		if v >= 1 && v == float64(int(v)) {
			n, nSource = int(v), "settings"
		} else {
			nSource = "invalid"
		}
	case string:
		if i, perr := strconv.Atoi(strings.TrimSpace(v)); perr == nil && i >= 1 {
			n, nSource = i, "settings"
		} else {
			nSource = "invalid"
		}
	default:
		nSource = "invalid"
	}
	return n, nSource, refusal, nil
}

// structureControlURL is the invented path no real site serves. Random per run
// so an edge cannot have cached an answer for it.
func structureControlURL(domain string) (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "https://" + domain + "/__acceptance-control-" + hex.EncodeToString(b[:]) + ".html", nil
}

// structureFetch GETs the way a visitor would: redirects followed, body capped.
// (0, "", err) is a transport failure and is NOT a status.
func structureFetch(ctx context.Context, target string) (int, string, error) {
	cctx, cancel := context.WithTimeout(ctx, structureFloorFetchTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(cctx, http.MethodGet, target, nil)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("User-Agent", "agentchassis-discovery/1.0 (+structure_floor)")
	req.Header.Set("Accept", "text/html,*/*")
	resp, err := structureFloorHTTPClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, structureFloorBodyCap))
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(body), nil
}

// structureDetect applies the v1 rubric to ONE served page and returns the
// structures it delivers, sorted. Pure over the body, so every rule is a table
// test. goquery, never a regex over the HTML: an element and a mention of an
// element are different things (check_asset_reference_404's lesson).
func structureDetect(body, domain string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	content := doc.Find("body")
	// Chrome and code carry nothing a reader gets. Stripped in place; the
	// document is ours.
	content.Find("style, script, nav, header, footer").Remove()

	text := structureNonAnchorText(content)
	found := map[string]bool{}

	// list: ≥5 <li> that say something outside a link. A list of links is
	// navigation wearing a <ul>.
	content.Find("ul, ol").EachWithBreak(func(_ int, l *goquery.Selection) bool {
		n := 0
		l.ChildrenFiltered("li").Each(func(_ int, li *goquery.Selection) {
			if strings.TrimSpace(structureNonAnchorText(li)) != "" {
				n++
			}
		})
		if n >= 5 {
			found["list"] = true
			return false
		}
		return true
	})

	// table, and the width half of comparison.
	wide := false
	content.Find("table").Each(func(_ int, t *goquery.Selection) {
		if t.Find("tr").Length() >= 3 {
			found["table"] = true
		}
		if structureMaxColumns(t) >= 3 {
			wide = true
		}
	})
	if wide {
		vocab := content.Find("h1, h2, h3").FilterFunction(func(_ int, h *goquery.Selection) bool {
			return structureComparisonRe.MatchString(h.Text())
		})
		if vocab.Length() > 0 {
			found["comparison"] = true
		}
	}

	// calendar: named as one, or six distinct months said in the reader's text.
	if structureWithClassToken(content, "period-cal", "calendar").Length() > 0 ||
		structureDistinctMonths(text) >= 6 {
		found["calendar"] = true
	}

	// checklist
	if content.Find("input").FilterFunction(structureInputTypeIs("checkbox")).Length() >= 3 {
		found["checklist"] = true
	} else {
		structureWithClassToken(content, "checklist").EachWithBreak(func(_ int, e *goquery.Selection) bool {
			if e.Find("li").Length() >= 3 {
				found["checklist"] = true
				return false
			}
			return true
		})
	}

	// tool: two operable fields and a button, inside a form or a `tool` block.
	content.Find("form").AddSelection(structureWithClassToken(content, "tool")).
		EachWithBreak(func(_ int, f *goquery.Selection) bool {
			fields := f.Find("input, select, textarea").FilterFunction(structureOperableField).Length()
			if fields >= 2 && f.Find("button").Length() >= 1 {
				found["tool"] = true
				return false
			}
			return true
		})

	// directory: six siblings, each a card pointing off-site. Grouped by
	// PARENT NODE so six cards scattered across six sections are not a directory.
	byParent := map[*html.Node]int{}
	structureWithClassToken(content, "card", "item", "entry", "listing").Each(func(_ int, e *goquery.Selection) {
		if len(e.Nodes) == 0 || e.Nodes[0].Parent == nil {
			return
		}
		if structureHasExternalLink(e, domain) {
			byParent[e.Nodes[0].Parent]++
		}
	})
	for _, siblings := range byParent {
		if siblings >= 6 {
			found["directory"] = true
			break
		}
	}

	// feed: a news/feed block whose children are dated or point off-site.
	structureWithClassToken(content, "news", "feed").EachWithBreak(func(_ int, e *goquery.Selection) bool {
		n := 0
		e.Children().Each(func(_ int, ch *goquery.Selection) {
			if ch.Is("time") || ch.Find("time").Length() > 0 || structureHasExternalLink(ch, domain) {
				n++
			}
		})
		if n >= 3 {
			found["feed"] = true
			return false
		}
		return true
	})

	// guide: sectioned and long enough to be read, not skimmed.
	if content.Find("h2").Length() >= 4 && len(strings.Fields(text)) >= 600 {
		found["guide"] = true
	}

	// faq
	if content.Find("details").Length() >= 3 {
		found["faq"] = true
	} else {
		structureWithClassToken(content, "faq").EachWithBreak(func(_ int, e *goquery.Selection) bool {
			if e.Find("li").Length() >= 3 || e.Find("h3").Length() >= 3 {
				found["faq"] = true
				return false
			}
			return true
		})
	}

	out := make([]string, 0, len(found))
	for s := range found {
		out = append(out, s)
	}
	sort.Strings(out)
	return out, nil
}

// structureNonAnchorText is the text a reader gets that is not a link: every
// text node under sel whose ancestry contains no <a>. Text nodes are joined with
// a space so adjacent inline elements do not fuse into one word.
func structureNonAnchorText(sel *goquery.Selection) string {
	var b strings.Builder
	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			b.WriteString(n.Data)
			b.WriteByte(' ')
			return
		case html.ElementNode:
			if n.Data == "a" {
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	for _, n := range sel.Nodes {
		walk(n)
	}
	return b.String()
}

// structureWithClassToken returns the descendants of content carrying ANY of
// the given classes as a WHOLE token — split on whitespace, compared
// case-insensitively. `card__title` does not carry `card`; `faq-list` does not
// carry `faq`. Substring matching is how a BEM element would count as its block.
func structureWithClassToken(content *goquery.Selection, tokens ...string) *goquery.Selection {
	want := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		want[strings.ToLower(t)] = true
	}
	return content.Find("[class]").FilterFunction(func(_ int, s *goquery.Selection) bool {
		cls, _ := s.Attr("class")
		for _, tok := range strings.Fields(cls) {
			if want[strings.ToLower(tok)] {
				return true
			}
		}
		return false
	})
}

// structureInputTypeIs filters <input> by type, case-insensitively.
func structureInputTypeIs(want string) func(int, *goquery.Selection) bool {
	return func(_ int, s *goquery.Selection) bool {
		t, _ := s.Attr("type")
		return strings.EqualFold(strings.TrimSpace(t), want)
	}
}

// structureOperableField keeps the fields a reader can operate: every <select>
// and <textarea>, and an <input> that is not hidden and not a button in
// disguise. See the header for why this narrowing is stated rather than silent.
func structureOperableField(_ int, s *goquery.Selection) bool {
	if !s.Is("input") {
		return true
	}
	t, _ := s.Attr("type")
	switch strings.ToLower(strings.TrimSpace(t)) {
	case "hidden", "submit", "button", "reset", "image":
		return false
	}
	return true
}

// structureMaxColumns is the widest row of a table, counting <td>/<th> cells.
// colspan is not honoured — a v1 simplification the architecture seat can lift.
func structureMaxColumns(table *goquery.Selection) int {
	widest := 0
	table.Find("tr").Each(func(_ int, row *goquery.Selection) {
		if n := row.ChildrenFiltered("td, th").Length(); n > widest {
			widest = n
		}
	})
	return widest
}

// structureDistinctMonths counts distinct month names in text (see
// structureMonthRe for the case rule).
func structureDistinctMonths(text string) int {
	seen := map[string]bool{}
	for _, m := range structureMonthRe.FindAllString(text, -1) {
		seen[strings.ToLower(m)] = true
	}
	return len(seen)
}

// structureHasExternalLink reports whether sel contains an <a href> to an
// absolute http(s) URL on a host other than the site's own.
func structureHasExternalLink(sel *goquery.Selection, domain string) bool {
	external := false
	sel.Find("a[href]").EachWithBreak(func(_ int, a *goquery.Selection) bool {
		href, _ := a.Attr("href")
		if structureIsExternalHref(href, domain) {
			external = true
			return false
		}
		return true
	})
	return external
}

// structureIsExternalHref: absolute, http(s), and not this site — where the
// site's www. form and its bare form are one host, and a port is not a host.
func structureIsExternalHref(href, domain string) bool {
	u, err := url.Parse(strings.TrimSpace(href))
	if err != nil || !u.IsAbs() {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	return structureHostKey(u.Host) != structureHostKey(domain)
}

func structureHostKey(hostport string) string {
	h := strings.ToLower(strings.TrimSpace(hostport))
	if host, _, err := net.SplitHostPort(h); err == nil {
		h = host
	}
	h = strings.TrimSuffix(h, ".")
	return strings.TrimPrefix(h, "www.")
}
