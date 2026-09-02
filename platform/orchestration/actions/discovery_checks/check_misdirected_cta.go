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
//     an authored link to a real page (bugs_open/248: including one in an
//     excluded area, which it used to overwrite) — is re-detected on the
//     next discovery pass and escalates via the two-strike rule to human
//     review — loud, not silent.
//
//   - cta_names_unknown_destination: anchor text names NO real page AND the
//     href is empty, phantom, self-referential, or points at the homepage.
//     Copy promising a destination that does not exist is a product decision
//     (build the page? rewrite the copy?), so this goes to needs_human_review
//     with no handler. An href that lands in an excluded area is RECORDED as a
//     finding but files no work item (bugs_open/248) — a real contact page is
//     a legitimate CTA destination, and judging one by its area filed 103
//     items of which a sampled 18 of 18 were correct buttons.
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

// ctaExcludedAreas is resolve_internal_links' areasExcludedFromCTA —
// destinations that link copy never "names" and CTAs should not target.
// Was a hand-mirrored copy ("duplicated (not imported) because actions
// imports this package") until bugs_open/436 moved the set itself to
// datahelpers, where this package CAN read it; the two can no longer drift.
var ctaExcludedAreas = datahelpers.CTAExcludedAreas

// ctaComponentScanQuery is the deployed-HTML scan shared by this check and
// check_cta_nonpage — one spelling, so the two CTA checks cannot disagree on
// which components they examine (the lockstep discipline).
//
// The page-status filter was ABSENT until 2026-08-18 (bugs_open/299 lane):
// archived pages were scanned, and webdesign.uk's index-rejected-v1-20260806
// minted two cta_links_stale rerender items — both failed, because nothing
// downstream will touch an archived page (archived_page_guard refuses the
// deploy). The filter is spelled exactly as loadCTAMatchIndex below spells it;
// the shared constant in the actions package (prepare_link_context_action.go)
// is unreachable from here — actions imports this package.
const ctaComponentScanQuery = `
		SELECT p.name, p.url, pc.page_id::text, COALESCE(pc.slot_name, ''), pc.rendered_html
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND p.status NOT IN ('deleted', 'archived')
		  AND pc.rendered_html IS NOT NULL AND pc.rendered_html <> ''
		ORDER BY p.name, pc.position
	`

// ctaClassifyAnchor is THE definition of "this CTA is misdirected", shared by
// the discovery check and the completion verifier so the two cannot drift. If
// they drifted, a verifier could resolve a defect the next discovery pass
// immediately re-detects — churn that reads as progress, the failure bugs_open/017
// was filed about.
//
// Three outcomes, which the caller must distinguish:
//   - named=false            → the anchor's copy names no real page. Not a
//     misdirect; the check applies its own unknown-destination rules to it.
//   - named=true, nil        → copy names a page and the href agrees. Healthy.
//   - named=true, non-nil    → copy names one page, href points elsewhere.
//
// Generic link text ("Learn More") reduces to zero distinctive tokens and is
// reported as named=false, so it is never a misdirect.
func ctaClassifyAnchor(a datahelpers.Anchor, slotName string, pages []datahelpers.LabelMatchCandidate,
	pageName, pageURL string) (*misdirectedAnchor, bool) {
	// AMBIGUOUS copy is named=false (bugs_open/308 Phase B). BestLabelMatch
	// reports ambiguity when the winner was separated from a DIFFERENT page by
	// nothing but alphabetical order, and this check must not file a repair on
	// that: measured 2026-08-23, 19 live findings in two families
	// (finetuning.uk "how we work" → /about.html over the /how-we-work.html the
	// copy names; dartsonline.com "Read the guides" → /about.html over
	// /guides/index.html) are exactly this, and Phase B's widening would have
	// had the repairer EXECUTE them. The third return is discarded rather than
	// recorded on purpose: a "copy names two pages equally" finding is a real
	// signal but a new work-item type, and this change is already a widening.
	// The page identity is a parameter because "does this copy name a page?"
	// cannot be answered without knowing which page the copy is ON: a label
	// naming its own page names nothing, and this check already files the
	// mirror-image defect ("links back to its own page") on the href side.
	// Suggesting one as the repair target was that defect arriving from the
	// other direction, and Phase B's widening made it common — 35 of 291 writes
	// on the writers' side, measured 2026-08-23.
	// ADAPTOR (bugs_open/399). The three lines this replaced are now
	// datahelpers.JudgeCTALabel, so the write-time audit that watches a CTA
	// being persisted and this detector ask ONE question with one answer. The
	// proof that the extraction changed nothing is that this file's tests and
	// cta_classify_anchor_test.go pass UNCHANGED — the same bar bugs_open/203's
	// extraction of BestLabelMatch met. The third return of the match (the
	// RFC_047 ambiguity flag) is still discarded HERE for the reason recorded
	// above; JudgeCTALabel now carries it so the audit can record it without
	// this check having to file a new work-item type.
	j := datahelpers.JudgeCTALabel(a.Text, a.Href, pages, pageName, pageURL)
	switch j.Verdict {
	case datahelpers.CTALabelNoOpinion:
		return nil, false
	case datahelpers.CTALabelAgrees:
		return nil, true
	}
	return &misdirectedAnchor{
		SlotName:             slotName,
		Text:                 a.Text,
		Href:                 a.Href,
		SuggestedTarget:      j.Named.URL,
		SuggestedTargetTitle: j.Named.Title,
	}, true
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
		// filesWorkItem distinguishes "report this to a human" from "record it
		// and move on". Every arm but one files; the excluded-area arm does
		// not — see the switch below and bugs_open/248.
		filesWorkItem bool
	}
	var unknowns []unknownDest

	rows, err := dctx.DB.QueryContext(dctx.Ctx, ctaComponentScanQuery, dctx.SiteID)
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
			misdirect, named := ctaClassifyAnchor(a, slotName, pages, pageName, pageURL)
			if named {
				if misdirect == nil {
					continue // copy and destination agree
				}
				agg, seen := misdirectedByPage[pageName]
				if !seen {
					agg = &pageAgg{pageID: pageID}
					misdirectedByPage[pageName] = agg
					misdirectedOrder = append(misdirectedOrder, pageName)
				}
				agg.misdirects = append(agg.misdirects, *misdirect)
				continue
			}

			hrefNorm := datahelpers.NormalizePagePath(a.Href)

			// Distinctive text matching NO page: only a problem when the href
			// is itself wrong — empty, phantom, circular, an excluded area, or
			// the homepage (copy naming something specific that dumps the
			// visitor back at the start: the "Enter the Arena" -> /index.html
			// case). All review-only; none of these auto-fix.
			why := ""
			filesWorkItem := true
			switch {
			case scope == datahelpers.LinkScopeEmpty:
				why = "empty href"
			case !validPages.Contains(a.Href):
				why = "phantom destination"
			case hrefNorm == datahelpers.NormalizePagePath(pageURL):
				why = "links back to its own page"
			case ctaAreaExcluded(a.Href):
				// RECORDED, NOT FILED (bugs_open/248, slug
				// cta_recompute_clobbers_authored_contact_links). This arm
				// judged a destination that IS a real page, purely on the area
				// it lives in — the same conflation that let a CTA repair
				// overwrite authored contact buttons, and the premise that fix
				// overturns for stored links. Its measured precision on live
				// data was ~0: 103 items filed, 103 still open, and an
				// independent 2026-08-07 audit found 18 of 18 sampled were
				// CORRECT contact buttons ("Get in Touch", "Talk to Us").
				//
				// It is DEMOTED rather than deleted because it is the only arm
				// that can see a FABRICATED but valid contact link: the phantom
				// arm is blind once the contact page exists, and the misdirect
				// arm is blind when the copy names no page. The finding is still
				// emitted, so the signal survives in discovery output; what
				// stops is the needs_human_review work item, which is what made
				// a queue nobody could drain.
				why = "lands in an excluded area (contact/legal/about)"
				filesWorkItem = false
			case hrefNorm == "/":
				why = "names a specific destination but points at the homepage"
			}
			if why != "" {
				unknowns = append(unknowns, unknownDest{
					pageName: pageName, pageID: pageID, slotName: slotName,
					text: a.Text, href: a.Href, why: why,
					filesWorkItem: filesWorkItem,
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
			"filed":     u.filesWorkItem,
		})

		if !u.filesWorkItem {
			continue // recorded as a finding above; no human queue item
		}

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

// loadCTAMatchIndex returns this check's two page sets: the MATCH CANDIDATES —
// now datahelpers.LoadCTALabelUniverse, the definition shared with the two CTA
// writers (bugs_open/308 Phase B) — and the VALIDITY set used by the phantom
// arm.
//
// THE TWO SETS ARE DIFFERENT ON PURPOSE and are not merged:
//   - the homepage is in the validity set (a link to "/" is not a phantom) and
//     out of the candidate set (link copy rarely names the homepage);
//   - a page that is planned and never deployed is OUT of the candidate set
//     (the writers' validPages gate refuses it, so suggesting one files a
//     repair no writer can perform — 10 live findings did exactly that,
//     measured 2026-08-23) but stays IN the validity set, because narrowing
//     validity would newly classify those links as phantom and file work items
//     this change has not measured.
//
// So it costs a second query, and that is the honest price of two axes.
func loadCTAMatchIndex(dctx DiscoveryCheckContext) ([]datahelpers.LabelMatchCandidate, datahelpers.PageURLSet, error) {
	pages, err := datahelpers.LoadCTALabelUniverse(dctx.Ctx, dctx.DB, dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("misdirected_cta candidate universe failed: %w", err)
	}

	rows, err := dctx.DB.QueryContext(dctx.Ctx, `
		SELECT p.url
		FROM pages p
		WHERE p.site_id = $1
		  AND p.status NOT IN ('deleted', 'archived')
		  AND p.url IS NOT NULL AND p.url <> ''
	`, dctx.SiteID)
	if err != nil {
		return nil, nil, fmt.Errorf("misdirected_cta pages query failed: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			continue
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("misdirected_cta pages iteration failed: %w", err)
	}
	urls = append(urls, "/", "/index.html")
	return pages, datahelpers.NewPageURLSet(urls), nil
}
