// FILE: platform/orchestration/actions/discovery_checks/check_missing_prose_links.go
//
// Detects deployed prose pages whose WRITER emitted no internal links at all,
// and files one edit-in-place rewrite to add some. It also reads and resolves
// the `LINK_CONTEXT_UNAVAILABLE` rows that record one known cause of it.
//
// This is the near-INVERSE of check_orphan_pages.go, which asks "does anything
// link TO this page". This one asks "does this page link to ANYTHING".
//
// ── THE INSTRUMENT, because three of them disagree and only one is right ──
//
// A "prose link" is an <a href> whose target ClassifyLinkScope calls a page,
// found in the decoded STRING VALUES of page_components.content_data — the
// writer's own output. Deliberately NOT:
//
//   - page_components.rendered_html — the template injects nav, hero and CTA
//     links, so a page whose writer emitted nothing still reads 2-3. Measured
//     2026-08-25: that instrument finds 140 zero-link pages fleet-wide where
//     this one finds 411. The rendered instrument is counting the compositor's
//     work and calling it the writer's.
//   - structured link fields (cta_url, link_url, hero_url). Those are real
//     links and they are chosen by a DIFFERENT mechanism (CTA selection), not
//     by the writer's prose. They carry no `href=`, so ExtractHrefs excludes
//     them for free — but the exclusion is intended, not incidental, and a
//     later maintainer must not "fix" it.
//
// Link identification reuses datahelpers.ExtractHrefs + ClassifyLinkScope, the
// same pair the deploy gate and check_phantom_internal_links use, so "internal
// link" means one thing in all three.
//
// ── WHY THE REPAIR IS A REWRITE AND NOT A RERENDER ──
//
// The missing links live IN content_data, and a rerender regenerates the page
// FROM content_data — so page_rerender is structurally incapable of fixing
// this, and bugs_open/392's own fix candidate 1 (which names it) is wrong.
// Verified against live agent_definitions 2026-08-25: page-rerender neither
// spawns page-content-writer nor knows edit_live; page-build-handler does both;
// and page-content-writer is the ONLY agent fleet-wide whose workflow runs
// prepare_link_context. So the repair re-runs the writer, which means it gets a
// fresh link allow-list and picks its own targets — no upstream link-planning
// step is needed.
//
// spec.mode = "edit_live" is load-bearing: it makes the writer EDIT the current
// sections rather than regenerate them (migration 299, built for bugs_open/178).
// internal-linker's own rewrites do not set it, which is why they risk 178-style
// prose destruction. We do not inherit that gap.
//
// ── OWNED PAGES ARE FILED, NOT EXCLUDED (deliberate) ──
//
// page-build-handler declares refuse_owned_page, so writeWorkItem's door parks
// these at write time — before any dispatch or LLM spend — clearing the handler
// and marking the summary. That is better than excluding them here: an excluded
// page is recorded NOWHERE, while a parked one becomes standing demand evidence
// for the open question of what repairs owned-page content (register WII-028;
// bugs_open/277 is CLOSED, do not route there). Measured 2026-08-25: 48 of 48
// owned prose pages carry no writer links, which is the ownership guard working
// as designed, not a defect of this check's.
//
// ⚠ The door keys on the page_id COLUMN. An item without PageID walks straight
// past it and burns an LLM run before dying wont_fix at SavePageSectionsAction's
// OWNED_PAGE_GUARD. PageID below is therefore load-bearing twice (the verifier
// coverage test also asserts every needs_internal_links-family item carries one).
//
// Registration: automatic via init() → Register(&MissingProseLinksCheck{}).
// Inert until named in a live `run_checks.config.checks` array.
package discovery_checks

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func init() { Register(&MissingProseLinksCheck{}) }

type MissingProseLinksCheck struct{}

func (c *MissingProseLinksCheck) Name() string { return "missing_prose_links" }

// linkContextUnavailableCode must be declared in THIS file, not imported.
// The finding-code registry checker (cmd/config-key-audit/findingcodes.go)
// verifies that a `consumed` code's declared reader FILE literally contains the
// code string; a const in another package does not satisfy it. The writer's
// copy is prepare_link_context_action.go:33 — the two are deliberate duplicates
// and the registry entry names this file as the reader.
const linkContextUnavailableCode = "LINK_CONTEXT_UNAVAILABLE"

// proseLinkPageTypes are the page types where an internal link in the body is
// the NORM, so its absence is anomalous. Chosen from a census, not from taste
// (2026-08-25, zero-prose-link / total deployed): blog-post 31/164, guide
// 59/98, content 97/158 — against tool 159/200 and entity-page 15/17, where
// zero is the ordinary state and convicting would be the wrong instrument.
//
// ⚠ This list does NOT exclude tool pages by name, and that is not the safety
// property it looks like: fleet-wide 103 deployed `tool-%` pages are typed
// blog-post. It is acceptable today only because the type rule and a structural
// rule ("does the page carry a tool-level component") select the same set —
// checked 2026-08-25. If a guide ever gains an embedded tool slot they diverge
// SILENTLY, and the structural rule is the one to reach for. Re-run that
// comparison before trusting this list a second time.
var proseLinkPageTypes = map[string]bool{
	"blog-post": true,
	"guide":     true,
	"content":   true,
}

const (
	// minProseCharsForLinks — a stub page with a hero and nothing else is not
	// "prose that should link". Measured against content_data text length.
	minProseCharsForLinks = 1500

	// minSiteTargetsForLinks — on a site with almost no other pages, "no
	// internal links" may be the truthful state rather than damage.
	minSiteTargetsForLinks = 3

	// maxMissingProseLinkItemsPerPass bounds the blast radius. 187 pages match
	// on the prose types today; the quality rotation visits roughly one site
	// per fire, so this caps the fleet-wide rate to a few items a day against
	// a content_rewrite baseline of ~30/day. The remainder is reported in the
	// census finding and re-detected on the next pass.
	maxMissingProseLinkItemsPerPass = 2
)

type proseLinkPage struct {
	ID          string
	Name        string
	URL         string
	PageType    string
	ContentData string
}

// pageMissingProseLinks is the whole rule, as a pure function, so the in-run
// canary and the unit tests grade the same code the site query does.
func pageMissingProseLinks(pageType string, internalPageLinks, proseChars, siteTargets int) bool {
	if !proseLinkPageTypes[pageType] {
		return false
	}
	if internalPageLinks > 0 {
		return false
	}
	if proseChars < minProseCharsForLinks {
		return false
	}
	if siteTargets < minSiteTargetsForLinks {
		return false
	}
	return true
}

// countProseLinks returns the number of internal PAGE links in the decoded
// string values of a content_data JSON document.
//
// Walking the decoded values rather than the raw text matters: in
// content_data::text an href is JSON-escaped (`href=\"`), so a naive scan of
// the raw string with an unescaped needle matches nothing at all and returns a
// confident, internally consistent zero (LANDMINES; WRONG_CALLS 2026-08-25).
// Unmarshalling first makes the escaping somebody else's problem.
func countProseLinks(contentDataJSON string) int {
	if strings.TrimSpace(contentDataJSON) == "" {
		return 0
	}
	var doc interface{}
	if err := json.Unmarshal([]byte(contentDataJSON), &doc); err != nil {
		// Unparseable content_data is NOT evidence of zero links. Report a
		// sentinel the caller treats as "cannot tell", never as "none".
		return -1
	}
	n := 0
	walkJSONStrings(doc, func(s string) {
		if !strings.Contains(s, "href") {
			return
		}
		for _, href := range datahelpers.ExtractHrefs(s) {
			if datahelpers.ClassifyLinkScope(href) == datahelpers.LinkScopePage {
				n++
			}
		}
	})
	return n
}

func walkJSONStrings(v interface{}, fn func(string)) {
	switch t := v.(type) {
	case string:
		fn(t)
	case []interface{}:
		for _, e := range t {
			walkJSONStrings(e, fn)
		}
	case map[string]interface{}:
		for _, e := range t {
			walkJSONStrings(e, fn)
		}
	}
}

// proseTextLength approximates how much prose a page carries, from the same
// decoded values the link count reads.
func proseTextLength(contentDataJSON string) int {
	var doc interface{}
	if err := json.Unmarshal([]byte(contentDataJSON), &doc); err != nil {
		return 0
	}
	n := 0
	walkJSONStrings(doc, func(s string) { n += len(s) })
	return n
}

// canaryMissingProseLinks grades fabricated cases through the pure predicate and
// the instrument on every run. A check whose predicate has drifted, or whose
// link extraction has stopped seeing links, must REFUSE rather than report a
// clean site — an errored check can neither file nor retract (the runner skips
// Resolved entirely when Run returned an error), which is the safe direction.
//
// A DB-pinned demand control would be self-blinding here: the known-positive
// population is exactly what this mechanism exists to drain, so it would go red
// on the day the check started working.
func canaryMissingProseLinks() error {
	cases := []struct {
		name                           string
		pageType                       string
		links, proseChars, siteTargets int
		want                           bool
	}{
		{"linkless guide is convicted", "guide", 0, 5000, 10, true},
		{"guide with a link is clean", "guide", 2, 5000, 10, false},
		{"linkless tool page is out of remit", "tool", 0, 5000, 10, false},
		{"stub is too thin to judge", "content", 0, 200, 10, false},
		{"tiny site has nothing to link to", "content", 0, 5000, 1, false},
	}
	for _, tc := range cases {
		if got := pageMissingProseLinks(tc.pageType, tc.links, tc.proseChars, tc.siteTargets); got != tc.want {
			return fmt.Errorf("predicate drift: %q returned %v, want %v", tc.name, got, tc.want)
		}
	}

	// The instrument itself must be able to answer both ways, and must NOT
	// count a structured link field — that exclusion is the check's stated
	// scope, so it is asserted, not assumed.
	if n := countProseLinks(`{"content":"<p>see <a href=\"/about.html\">about</a></p>"}`); n != 1 {
		return fmt.Errorf("instrument drift: prose anchor counted %d, want 1", n)
	}
	if n := countProseLinks(`{"content":"<p>no links here</p>"}`); n != 0 {
		return fmt.Errorf("instrument drift: link-free prose counted %d, want 0", n)
	}
	if n := countProseLinks(`{"cta_url":"/contact.html","headline":"Talk to us"}`); n != 0 {
		return fmt.Errorf("instrument drift: structured cta_url counted %d, want 0 (it is not a prose anchor)", n)
	}
	return nil
}

// missingProseLinksPagesSQL reads every page on the site that is BOTH still
// wanted live and has actually shipped, with its writer-authored content.
//
// Both axes, per bugs_open/356: PageWantedLivePredicateFor is the LIFECYCLE arm
// ("does the platform still want this page") and PageHasShippedPredicateFor is
// the BUILD arm ("has it ever been served"). Taking the build arm alone
// enumerates an ARCHIVED page that shipped before it was retired — and this
// check's remedy REWRITES the page, so acting on a retired one would republish
// content the platform has decided to withdraw. The correct action on a retired
// page that is still serving is retraction (bugs_closed/098), never a rewrite.
var missingProseLinksPagesSQL = fmt.Sprintf(`
	SELECT p.id::text, p.name, COALESCE(p.url, ''), COALESCE(p.page_type, ''),
	       COALESCE(string_agg(pc.content_data::text, ' '), '')
	  FROM pages p
	  JOIN page_components pc ON pc.page_id = p.id
	 WHERE p.site_id = $1
	   AND %s
	   AND %s
	 GROUP BY p.id, p.name, p.url, p.page_type
	 ORDER BY p.name`,
	datahelpers.PageWantedLivePredicateFor("p"),
	datahelpers.PageHasShippedPredicateFor("p"))

// missingProseLinksTargetsSQL counts the pages on the site a link could point
// AT, using the linkability floor rather than the stricter shipped predicate:
// the question is "would a visitor get something", not "has this been served".
var missingProseLinksTargetsSQL = fmt.Sprintf(`
	SELECT count(*) FROM pages p
	 WHERE p.site_id = $1 AND COALESCE(p.url,'') <> '' AND %s`,
	datahelpers.PageWantedLivePredicateFor("p"))

// missingProseLinksOpenItemSQL asks whether some other producer already has a
// rebuild-shaped item open on this page, so we never stack a second repair on a
// pending one. The status list mirrors workItemTerminalStatuses
// (actions/work_items_common.go:42) — it lives in package `actions` and cannot
// be imported here, so it is duplicated deliberately; keep them in lockstep.
const missingProseLinksOpenItemSQL = `
	SELECT count(*) FROM site_work_items
	 WHERE site_id = $1 AND page_id = $2
	   AND item_type IN ('content_rewrite','needs_page','page_rerender','needs_internal_links')
	   AND status NOT IN ('complete','failed','verified','rejected','wont_fix','unresolved','cancelled')`

func (c *MissingProseLinksCheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	if err := canaryMissingProseLinks(); err != nil {
		return nil, fmt.Errorf("missing_prose_links canary failed, refusing to report: %w", err)
	}

	var siteTargets int
	if err := dctx.DB.QueryRowContext(dctx.Ctx, missingProseLinksTargetsSQL, dctx.SiteID).Scan(&siteTargets); err != nil {
		return nil, fmt.Errorf("missing_prose_links target count failed: %w", err)
	}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, missingProseLinksPagesSQL, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("missing_prose_links page query failed: %w", err)
	}
	defer rows.Close()

	var pages []proseLinkPage
	for rows.Next() {
		var p proseLinkPage
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.PageType, &p.ContentData); err != nil {
			dctx.Logger.Warn("missing_prose_links: failed to scan page", zap.Error(err))
			continue
		}
		pages = append(pages, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("missing_prose_links page scan failed: %w", err)
	}

	result := &CheckResult{}

	// Census by page type, emitted on EVERY run including a clean one — an
	// instrument that shifts is then visible in the run record rather than
	// silent, and out-of-remit types are counted rather than quietly dropped.
	census := map[string]int{}
	censusZero := map[string]int{}

	var missing []proseLinkPage
	var healthy []proseLinkPage
	unparseable := 0

	for _, p := range pages {
		census[p.PageType]++
		links := countProseLinks(p.ContentData)
		if links < 0 {
			unparseable++
			continue
		}
		if links == 0 {
			censusZero[p.PageType]++
		}
		if pageMissingProseLinks(p.PageType, links, proseTextLength(p.ContentData), siteTargets) {
			missing = append(missing, p)
			continue
		}
		if proseLinkPageTypes[p.PageType] && links > 0 {
			healthy = append(healthy, p)
		}
	}

	result.Findings = append(result.Findings, map[string]interface{}{
		"check":                    "missing_prose_links",
		"pages_by_type":            census,
		"zero_links_by_type":       censusZero,
		"site_link_targets":        siteTargets,
		"unparseable_content_data": unparseable,
		"in_remit_missing":         len(missing),
		"filed_this_pass":          minInt(len(missing), maxMissingProseLinkItemsPerPass),
	})

	// RETRACTION — positive observation only. A page that now carries prose
	// links closes its own item, by the detector-scoped key so it can never
	// close check_orphan_pages' inbound `needs_links:` items.
	for _, p := range healthy {
		result.Resolved = append(result.Resolved, ResolvedFinding{
			ItemType: "content_rewrite",
			ItemKey:  missingProseLinksItemKey(p.Name, dctx.SiteID),
			Reason:   "page now carries writer-authored internal links",
		})
	}

	filed := 0
	for _, p := range missing {
		if filed >= maxMissingProseLinkItemsPerPass {
			break
		}
		open, err := c.hasOpenRebuildItem(dctx, p.ID)
		if err != nil {
			dctx.Logger.Warn("missing_prose_links: open-item probe failed, skipping page",
				zap.String("page", p.Name), zap.Error(err))
			continue
		}
		if open {
			continue
		}
		item, ok := c.buildRewriteItem(dctx, p, "artefact")
		if !ok {
			continue
		}
		result.WorkItems = append(result.WorkItems, item)
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "missing_prose_links",
			"page_name": p.Name,
			"page_url":  p.URL,
			"page_type": p.PageType,
			"page_id":   p.ID,
			"trigger":   "artefact",
		})
		filed++
	}

	// ── The bugs_open/392 arm: read and resolve LINK_CONTEXT_UNAVAILABLE ──
	c.readLinkContextRows(dctx, result)

	dctx.Logger.Info("missing_prose_links: complete",
		zap.Int("pages_examined", len(pages)),
		zap.Int("in_remit_missing", len(missing)),
		zap.Int("items_filed", filed),
		zap.Int("retractions", len(result.Resolved)),
	)
	return result, nil
}

func missingProseLinksItemKey(pageName string, siteID uuid.UUID) string {
	// Prefix agreed with the webdesign_tool_rebuilds lane 2026-08-25 and
	// deliberately distinct from internal-linker's `internal_link:` and the
	// tool lane's `tool_crosslink:`. idx_swi_dedup is UNIQUE on
	// (site_id, item_key) with NO item_type column, so a shared prefix would
	// let two different defects on one page silently absorb each other.
	return fmt.Sprintf("no_outbound_links:%s:%s", pageName, siteID)
}

func (c *MissingProseLinksCheck) hasOpenRebuildItem(dctx DiscoveryCheckContext, pageID string) (bool, error) {
	pid, err := uuid.Parse(pageID)
	if err != nil {
		return false, err
	}
	var n int
	if err := dctx.DB.QueryRowContext(dctx.Ctx, missingProseLinksOpenItemSQL, dctx.SiteID, pid).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

func (c *MissingProseLinksCheck) buildRewriteItem(dctx DiscoveryCheckContext, p proseLinkPage, trigger string) (WorkItemSpec, bool) {
	pid, err := uuid.Parse(p.ID)
	if err != nil {
		dctx.Logger.Warn("missing_prose_links: unparseable page id, refusing to file",
			zap.String("page", p.Name), zap.Error(err))
		return WorkItemSpec{}, false
	}
	specJSON, err := json.Marshal(map[string]interface{}{
		"check":     "missing_prose_links",
		"mode":      "edit_live",
		"reason":    "missing_prose_links",
		"page_name": p.Name,
		"page_id":   p.ID,
		"page_url":  p.URL,
		"page_type": p.PageType,
		"trigger":   trigger,
		"suggestion": "This page's prose contains no internal links to other pages on this site. " +
			"Weave 2-4 contextual internal links into the EXISTING prose, choosing only pages from " +
			"the internal-linking allow-list provided to you. Do not add, remove or reword anything " +
			"else: keep every factual claim, figure and heading exactly as it is.",
	})
	if err != nil {
		return WorkItemSpec{}, false
	}
	return WorkItemSpec{
		SiteID:   dctx.SiteID,
		PageID:   &pid, // load-bearing: the owned-page door keys on this COLUMN
		Source:   "discovery",
		Pipeline: "build",
		ItemType: "content_rewrite",
		Severity: "low",
		Summary: fmt.Sprintf("Page %s (%s) has no internal links in its prose — add 2-4 in place",
			p.Name, p.URL),
		SpecJSON:     string(specJSON),
		Priority:     90,
		HandlerAgent: "page-build-handler",
		Status:       "detected",
		CreatedBy:    dctx.AgentType,
		ItemKey:      missingProseLinksItemKey(p.Name, dctx.SiteID),
		BatchID:      dctx.BatchID,
	}, true
}

// readLinkContextRows is the reader half that flips LINK_CONTEXT_UNAVAILABLE
// from `human-evidence` to `consumed` in the finding-code registry.
//
// The writer degrades deliberately when it cannot establish what pages exist
// and records one row saying so; until this function existed nothing ever read
// one. Rows are resolved ONLY on a healed verdict — the page now carries prose
// links, or it is gone. A row whose page is still link-less stays open and is
// re-graded next pass; that costs nothing, because migrations 567/580 removed
// the arm that used to shorten a resolved row's life and these rows live 365
// days either way.
//
// Best-effort throughout: a failure here must never lose the artefact findings
// already gathered.
func (c *MissingProseLinksCheck) readLinkContextRows(dctx DiscoveryCheckContext, result *CheckResult) {
	const selectSQL = `
		SELECT id::text, COALESCE(context->>'page_name',''), COALESCE(orchestration_id,'')
		  FROM agent_error_log
		 WHERE error_code = '` + linkContextUnavailableCode + `'
		   AND resolved = false AND site_id = $1`

	rows, err := dctx.DB.QueryContext(dctx.Ctx, selectSQL, dctx.SiteID)
	if err != nil {
		dctx.Logger.Warn("missing_prose_links: link-context row read failed", zap.Error(err))
		return
	}
	type lcRow struct{ ID, PageName, OrchestrationID string }
	var pending []lcRow
	for rows.Next() {
		var r lcRow
		if err := rows.Scan(&r.ID, &r.PageName, &r.OrchestrationID); err != nil {
			continue
		}
		pending = append(pending, r)
	}
	rows.Close()
	if len(pending) == 0 {
		return
	}

	for _, r := range pending {
		pageName := r.PageName
		if pageName == "" && r.OrchestrationID != "" {
			// Fallback while the writer's context enrichment has not rolled.
			// ⚠ This join has an expiry the row does not: orchestration rows
			// are reaped, the log row lives 365 days.
			_ = dctx.DB.QueryRowContext(dctx.Ctx, `
				SELECT COALESCE(collected_data->'input_data'->'current_page'->>'name','')
				  FROM orchestration_states WHERE orchestration_id = $1::uuid`,
				r.OrchestrationID).Scan(&pageName)
		}
		if pageName == "" {
			// Unattributable, but the artefact sweep above has already covered
			// this whole site, so nothing is owed on this page beyond saying so.
			c.resolveLinkContextRow(dctx, r.ID, "unattributable_site_scanned")
			continue
		}

		var links int
		var found bool
		var contentData string
		if scanErr := dctx.DB.QueryRowContext(dctx.Ctx, `
			SELECT COALESCE(string_agg(pc.content_data::text,' '),'')
			  FROM pages p JOIN page_components pc ON pc.page_id = p.id
			 WHERE p.site_id = $1 AND p.name = $2
			 GROUP BY p.id`, dctx.SiteID, pageName).Scan(&contentData); scanErr == nil {
			found = true
			links = countProseLinks(contentData)
		}

		switch {
		case !found:
			// The page is gone, or never got components. Nothing to repair.
			c.resolveLinkContextRow(dctx, r.ID, "page_gone")
		case links > 0:
			c.resolveLinkContextRow(dctx, r.ID, "healed")
		default:
			result.Findings = append(result.Findings, map[string]interface{}{
				"check":     "missing_prose_links",
				"page_name": pageName,
				"trigger":   linkContextUnavailableCode,
				"row_id":    r.ID,
				"note":      "row left open: page still carries no prose links",
			})
		}
	}
}

func (c *MissingProseLinksCheck) resolveLinkContextRow(dctx DiscoveryCheckContext, rowID, verdict string) {
	// `AND resolved = false` makes the update idempotent under re-runs, the
	// same guard cmd/content-loss-check/main.go:354 uses — the estate's only
	// other reader that resolves rows in agent_error_log.
	if _, err := dctx.DB.ExecContext(dctx.Ctx, `
		UPDATE agent_error_log SET resolved = true, resolved_at = now(), resolved_by = $2
		 WHERE id = $1::uuid AND resolved = false`,
		rowID, "missing_prose_links:"+verdict); err != nil {
		dctx.Logger.Warn("missing_prose_links: failed to resolve link-context row",
			zap.String("row", rowID), zap.Error(err))
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
