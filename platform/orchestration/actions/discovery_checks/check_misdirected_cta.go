// FILE: platform/orchestration/actions/discovery_checks/check_misdirected_cta.go
//
// Detects CTAs/links in DEPLOYED rendered HTML whose visible text names one
// real page while the href points at a different one — the failure mode the
// phantom check cannot see, because the wrong destination is still a real
// page ("Enter the Gauntlet" → /contact.html passes phantom_internal_links).
//
// Two deterministic findings, both from page_components.rendered_html:
//
//   - misdirected_cta: anchor text token-matches a real page's name/title/
//     nav_label (tool/game pages preferred) but the href normalises to a
//     different page. Emits ONE page_rerender work item per affected page
//     (not per anchor — multiple CTAs on a page need one rerender) with
//     spec.reason = "cta_links_stale", which page-rerender's section
//     re-render path uses to recompute CTA targets (rerender_page_sections'
//     applyCTARecompute). Note the recompute only rewrites the CTA url
//     fields of components in the actions package's ctaFieldNames set (hero,
//     call-to-action, archetype-grid, archetype-combinations, gauntlet-cta,
//     content-block-about); a misdirected link inside any other component
//     (e.g. prose) — or one the recompute deliberately keeps because it is
//     an authored link to a real, non-excluded page — is re-detected on the
//     next discovery pass and escalates via the two-strike rule to human
//     review — loud, not silent.
//
//   - cta_names_unknown_destination: anchor text names NO real page AND the
//     href is empty, phantom, self-referential, or lands in an excluded area
//     (contact/legal/about). Copy promising a destination that does not exist
//     is a product decision (build the page? rewrite the copy?), so this goes
//     to needs_human_review with no handler.
//
// Guard against false positives: anchor text is reduced to distinctive tokens
// (stopwords + generic CTA words removed); generic texts ("Learn More",
// "Get Started") match nothing and are skipped entirely. Index/home pages are
// excluded as match candidates — link text rarely "names" the homepage.
//
// Extraction/normalisation are the SHARED datahelpers definitions
// (ExtractAnchors / ClassifyLinkScope / PageURLSet) so this check, the deploy
// gate and the phantom audit agree on what an internal page link is.
//
// Registration: automatic via init() -> Register(&MisdirectedCTACheck{}).
// Enable by adding "misdirected_cta" to a discovery agent's checks array.

package discovery_checks

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { Register(&MisdirectedCTACheck{}) }

type MisdirectedCTACheck struct{}

func (c *MisdirectedCTACheck) Name() string { return "misdirected_cta" }

// ctaStopwords: words that carry no destination meaning in link text. One
// combined set — grammatical stopwords plus generic CTA verbs/nouns. A text
// whose every token is in here is "generic" and is never matched.
var ctaStopwords = map[string]bool{
	"a": true, "an": true, "and": true, "as": true, "at": true, "be": true,
	"by": true, "for": true, "from": true, "in": true, "into": true,
	"is": true, "it": true, "its": true, "of": true, "on": true, "or": true,
	"our": true, "s": true, "the": true, "their": true, "this": true,
	"to": true, "with": true, "your": true, "you": true, "yours": true,
	"we": true, "us": true,
	// generic CTA vocabulary
	"all": true, "click": true, "discover": true, "enter": true,
	"explore": true, "get": true, "go": true, "here": true, "learn": true,
	"meet": true, "more": true, "now": true, "read": true, "see": true,
	"start": true, "started": true, "take": true, "today": true,
	"todays": true, "try": true, "view": true, "visit": true, "join": true,
}

// ctaExcludedAreas mirrors resolve_internal_links' areasExcludedFromCTA —
// destinations that link copy never "names" and CTAs should not target.
// Duplicated (not imported) because actions imports this package.
var ctaExcludedAreas = map[string]bool{
	"about": true, "contact": true, "privacy": true, "terms": true, "legal": true,
}

// ctaTokens reduces text to its distinctive lowercase tokens.
func ctaTokens(text string) []string {
	var tokens []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			t := cur.String()
			if !ctaStopwords[t] {
				tokens = append(tokens, t)
			}
			cur.Reset()
		}
	}
	for _, r := range strings.ToLower(text) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			cur.WriteRune(r)
		default:
			flush()
		}
	}
	flush()
	return tokens
}

// ctaMatchPage is one real page as a match candidate.
type ctaMatchPage struct {
	ID          string
	Name        string
	Title       string
	URL         string
	Interactive bool // page_type tool/game — preferred CTA destinations
	tokens      map[string]bool
}

// bestPageMatch returns the candidate page whose tokens best overlap the
// anchor's, or nil. Ranking: interactive pages beat content pages, then
// higher overlap, then name (deterministic). Requires >= 1 token overlap.
func bestPageMatch(anchorTokens []string, pages []ctaMatchPage) *ctaMatchPage {
	var best *ctaMatchPage
	bestOverlap := 0
	for i := range pages {
		p := &pages[i]
		overlap := 0
		for _, t := range anchorTokens {
			if p.tokens[t] {
				overlap++
			}
		}
		if overlap == 0 {
			continue
		}
		if best == nil ||
			(p.Interactive && !best.Interactive) ||
			(p.Interactive == best.Interactive && overlap > bestOverlap) ||
			(p.Interactive == best.Interactive && overlap == bestOverlap && p.Name < best.Name) {
			best = p
			bestOverlap = overlap
		}
	}
	return best
}

// misdirectedAnchor is one flagged anchor on a page.
type misdirectedAnchor struct {
	SlotName             string `json:"slot_name"`
	Text                 string `json:"text"`
	Href                 string `json:"href"`
	SuggestedTarget      string `json:"suggested_target"`
	SuggestedTargetTitle string `json:"suggested_target_title"`
}

func (c *MisdirectedCTACheck) Run(dctx DiscoveryCheckContext) (*CheckResult, error) {
	pages, validPages, err := loadCTAMatchIndex(dctx)
	if err != nil {
		return nil, err
	}

	type pageAgg struct {
		pageID     string
		misdirects []misdirectedAnchor
	}
	misdirectedByPage := map[string]*pageAgg{}
	var misdirectedOrder []string

	type unknownDest struct {
		pageName, pageID, slotName, text, href, why string
	}
	var unknowns []unknownDest

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.name, p.url, pc.page_id::text, COALESCE(pc.slot_name, ''), pc.rendered_html
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
		ORDER BY p.name, pc.position
	`, dctx.SiteID)
	if err != nil {
		return nil, fmt.Errorf("misdirected_cta page query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var pageName, pageURL, pageID, slotName, html string
		if err := rows.Scan(&pageName, &pageURL, &pageID, &slotName, &html); err != nil {
			dctx.Logger.Warn("misdirected_cta: scan error", zap.Error(err))
			continue
		}

		for _, a := range datahelpers.ExtractAnchors(html) {
			scope := datahelpers.ClassifyLinkScope(a.Href)
			if scope != datahelpers.LinkScopePage && scope != datahelpers.LinkScopeEmpty {
				continue // external/mailto/asset/anchor — not internal page links
			}
			tokens := ctaTokens(a.Text)
			if len(tokens) == 0 {
				continue // generic text ("Learn More") names nothing — skip
			}

			best := bestPageMatch(tokens, pages)
			hrefNorm := datahelpers.NormalizePagePath(a.Href)

			if best != nil {
				if datahelpers.NormalizePagePath(best.URL) == hrefNorm {
					continue // copy and destination agree
				}
				agg, seen := misdirectedByPage[pageName]
				if !seen {
					agg = &pageAgg{pageID: pageID}
					misdirectedByPage[pageName] = agg
					misdirectedOrder = append(misdirectedOrder, pageName)
				}
				agg.misdirects = append(agg.misdirects, misdirectedAnchor{
					SlotName:             slotName,
					Text:                 a.Text,
					Href:                 a.Href,
					SuggestedTarget:      best.URL,
					SuggestedTargetTitle: best.Title,
				})
				continue
			}

			// Distinctive text matching NO page: only a problem when the href
			// is itself wrong — empty, phantom, circular, an excluded area, or
			// the homepage (copy naming something specific that dumps the
			// visitor back at the start: the "Enter the Arena" -> /index.html
			// case). All review-only; none of these auto-fix.
			why := ""
			switch {
			case scope == datahelpers.LinkScopeEmpty:
				why = "empty href"
			case !validPages.Contains(a.Href):
				why = "phantom destination"
			case hrefNorm == datahelpers.NormalizePagePath(pageURL):
				why = "links back to its own page"
			case ctaAreaExcluded(a.Href):
				why = "lands in an excluded area (contact/legal/about)"
			case hrefNorm == "/":
				why = "names a specific destination but points at the homepage"
			}
			if why != "" {
				unknowns = append(unknowns, unknownDest{
					pageName: pageName, pageID: pageID, slotName: slotName,
					text: a.Text, href: a.Href, why: why,
				})
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("misdirected_cta iteration failed: %w", err)
	}

	result := &CheckResult{}
	sort.Strings(misdirectedOrder)

	for _, pageName := range misdirectedOrder {
		agg := misdirectedByPage[pageName]

		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "misdirected_cta",
			"page_name": pageName,
			"count":     len(agg.misdirects),
			"findings":  agg.misdirects,
		})

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "misdirected_cta",
			"reason":    "cta_links_stale", // page-rerender gates its CTA recompute on this
			"page_name": pageName,
			"page_id":   agg.pageID,
			"findings":  agg.misdirects,
			"fix": "Link copy names a real page but the href points elsewhere. " +
				"A cta_links_stale rerender recomputes CTA targets from real pages " +
				"(interactive pages first) for components in the ctaFieldNames set.",
		})

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(agg.pageID); perr == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:       dctx.SiteID,
			PageID:       pageIDPtr,
			Source:       "discovery",
			Pipeline:     "build",
			ItemType:     "page_rerender",
			Severity:     "high",
			Summary:      fmt.Sprintf("%d misdirected CTA(s) on %s — copy names a different page than the link target", len(agg.misdirects), pageName),
			SpecJSON:     string(specJSON),
			Priority:     35,
			HandlerAgent: "page-rerender",
			Status:       "detected",
			CreatedBy:    dctx.AgentType,
			ItemKey:      fmt.Sprintf("misdirected_cta:%s:%s", pageName, dctx.SiteID),
			BatchID:      dctx.BatchID,
		})
	}

	for _, u := range unknowns {
		result.Findings = append(result.Findings, map[string]interface{}{
			"check":     "cta_names_unknown_destination",
			"page_name": u.pageName,
			"slot_name": u.slotName,
			"text":      u.text,
			"href":      u.href,
			"why":       u.why,
		})

		specJSON, _ := json.Marshal(map[string]interface{}{
			"check":     "cta_names_unknown_destination",
			"page_name": u.pageName,
			"page_id":   u.pageID,
			"slot_name": u.slotName,
			"text":      u.text,
			"href":      u.href,
			"why":       u.why,
			"fix": "Link copy names a destination no real page provides (" + u.why + "). " +
				"Product decision: build the promised page, or rewrite the copy/link.",
		})

		var pageIDPtr *uuid.UUID
		if parsed, perr := uuid.Parse(u.pageID); perr == nil {
			pageIDPtr = &parsed
		}

		result.WorkItems = append(result.WorkItems, WorkItemSpec{
			SiteID:    dctx.SiteID,
			PageID:    pageIDPtr,
			Source:    "discovery",
			Pipeline:  "content",
			ItemType:  "cta_names_unknown_destination",
			Severity:  "medium",
			Summary:   fmt.Sprintf("CTA %q on %s (%s): %s — no real page matches the copy", u.text, u.pageName, u.slotName, u.why),
			SpecJSON:  string(specJSON),
			Priority:  40,
			Status:    "needs_human_review",
			CreatedBy: dctx.AgentType,
			ItemKey:   fmt.Sprintf("cta_unknown_dest:%s:%s:%s", u.pageName, u.href, u.text),
			BatchID:   dctx.BatchID,
		})
	}

	if len(result.WorkItems) > 0 {
		dctx.Logger.Warn("misdirected_cta: found copy/destination mismatches",
			zap.Int("misdirected_pages", len(misdirectedOrder)),
			zap.Int("unknown_destinations", len(unknowns)),
			zap.String("site_id", dctx.SiteID.String()))
	}

	return result, nil
}

// ctaAreaExcluded — first meaningful path segment against the excluded set,
// handling top-level pages ("/contact.html" -> "contact").
func ctaAreaExcluded(href string) bool {
	p := strings.TrimPrefix(datahelpers.NormalizePagePath(href), "/")
	if i := strings.IndexByte(p, '/'); i >= 0 {
		p = p[:i]
	}
	p = strings.TrimSuffix(p, ".html")
	return ctaExcludedAreas[p]
}

// loadCTAMatchIndex loads the real pages as match candidates (index/home
// excluded — link text rarely "names" the homepage) plus the full validity
// set (which DOES include the homepage) for phantom tests.
func loadCTAMatchIndex(dctx DiscoveryCheckContext) ([]ctaMatchPage, datahelpers.PageURLSet, error) {
	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.id::text, p.name, COALESCE(p.title, p.name),
		       COALESCE(p.nav_label, ''), p.url,
		       COALESCE(p.page_type, 'content')
		FROM pages p
		WHERE p.site_id = $1
		  AND p.status NOT IN ('deleted', 'archived')
		  AND p.url IS NOT NULL AND p.url <> ''
	`, dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("misdirected_cta pages query failed: %w", err)
	}
	defer rows.Close()

	var pages []ctaMatchPage
	var urls []string
	for rows.Next() {
		var id, name, title, navLabel, url, pageType string
		if err := rows.Scan(&id, &name, &title, &navLabel, &url, &pageType); err != nil {
			continue
		}
		urls = append(urls, url)
		if name == "index" || name == "home" {
			continue
		}
		tokens := map[string]bool{}
		for _, src := range []string{name, title, navLabel} {
			for _, t := range ctaTokens(src) {
				tokens[t] = true
			}
		}
		if len(tokens) == 0 {
			continue
		}
		pages = append(pages, ctaMatchPage{
			ID:          id,
			Name:        name,
			Title:       title,
			URL:         url,
			Interactive: pageType == "tool" || pageType == "game",
			tokens:      tokens,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("misdirected_cta pages iteration failed: %w", err)
	}
	urls = append(urls, "/", "/index.html")
	return pages, datahelpers.NewPageURLSet(urls), nil
}
