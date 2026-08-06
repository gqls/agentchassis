// FILE: platform/orchestration/actions/discovery_checks/check_phantom_internal_links_fragments.go
//
// The dead_fragment_link arm of check_phantom_internal_links: a link whose
// #fragment names an element id that exists nowhere on the page it lands on.
// The visitor clicks "jump to pricing" and the page does not move.
//
// WHY IT IS AN ARM AND NOT A CHECK OF ITS OWN. The blind spot this closes
// (bugs_open/071 § "Related defect, same blind spot: nothing validates the
// FRAGMENT") is *inside* the link classifier every consumer already shares:
// datahelpers.ClassifyLinkScope sends "#x" to LinkScopeAnchor, which the gate
// and this audit both skip by name, and NormalizePagePath drops the "#x" tail
// off "/page.html#x" before either compares anything. So the two link surfaces
// this file judges are the same two rows the phantom scan already reads, on the
// same schedule, in the same run — and a separate check would need its own
// enablement in a discovery agent's checks array, which is precisely how
// bugs_open/093's fix reached production correct and never executed.
//
// WHAT COUNTS AS AN ID IS NOT DECIDED HERE. datahelpers.DocumentIDs is the
// shared presence test, extracted from OrphanElementRefs (check_orphan_element_refs)
// when this arm was written. That check answers the same question from the other
// end — a SCRIPT naming an id the page lacks, where this is a LINK naming one —
// and each of its conservatisms was bought with a false positive on a working
// tool (its header's 2026-07-29 correction). Two implementations would re-buy
// that lesson at the price of sending a fixer at a page that works.
//
// THREE RESOLUTION RULES, and the reason each is shaped the way it is:
//
//  1. A bare "#x" in a PAGE component resolves against that page's whole
//     document — every one of its components plus the site chrome — matching
//     OrphanElementRefs' whole-page rule. A component judged alone reports every
//     id living in its siblings.
//
//  2. A bare "#x" in a SITE component (header/footer skip-links, "back to top")
//     is satisfied if ANY page on the site satisfies it. Chrome renders on every
//     page, so per-page judgement would file N items for one template, and the
//     live instance is the shape that would suffer: loanandmortgagecalculator's
//     header skip-link targets id="content", which its pages carry and which a
//     decomposed page may not. Reporting a chrome link that works almost
//     everywhere is noise; reporting one that works NOWHERE is a template defect.
//
//  3. "/page.html#x" resolves against the TARGET page's document, not the
//     containing one. The path half is judged by the phantom scan already; this
//     arm speaks only when the path resolves to a real, built page — a phantom
//     path is reported once, as a phantom, and an unbuilt target is reported once,
//     as unbuilt_internal_link. One href never produces two findings.
//
// AND FOUR SILENCES, all deliberate:
//
//   - a target page with NO stored HTML is not judged. An empty document
//     satisfies nothing, so judging it would report every fragment pointing at a
//     page whose components have not rendered yet — absence of evidence.
//   - "#" and "#!" are IsNoopHref's, i.e. check_dead_controls' remit, not a
//     fragment that fails to resolve.
//   - hrefs inside a data-runtime-fill shell are hydrated client-side; their
//     targets need not exist at build time (same exemption the empty-href arm makes).
//   - a page that INTERPOLATES its ids (id="${...}") has a computed id set;
//     DocumentIDs.Satisfies already loosens there rather than guessing.
//
// THE INHERITED LOOSENESS FAILS TOWARD FALSE NEGATIVES, ON PURPOSE (council
// round 1, bug_historian seat, medium: "a shared predicate written for one INPUT
// SHAPE, reused on another"). DocumentIDs' presentIDRe harvests ids from the
// whole page text INCLUDING inside script string literals, which
// OrphanElementRefs does deliberately so that a tool building its own markup is
// never accused. Resolving a LINK against that same set inherits it: a
// "#pricing" whose id exists only inside a script string is called resolved even
// though the browser may never paint it. That is a known miss, and it is the
// direction this arm must fail in — the cost of a false negative is one
// unreported inert control; the cost of a false positive is a fixer sent at a
// working page. Tightening it is what would produce the second kind.
//
// SEVERITY IS LOW, AND THAT IS THE HONEST GRADE: a dead fragment is an inert
// control, not a 404. bugs_open/071 measured the difference on the same links —
// repairing the PATH of "/capabilities#approach" turns it "from 404 into inert,
// which is an improvement, not a fix".

package discovery_checks

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func init() { RegisterVerifier("dead_fragment_link", VerifyDeadFragmentLinkResolved) }

// fragmentIndex is one site's answer to "if a visitor lands on page P and the
// browser looks for #x, does it find it?". Built once per check run: the regex
// scans are per-document and a site's links vastly outnumber its documents.
type fragmentIndex struct {
	// byPageID holds each page's document ids (its own components + chrome).
	byPageID map[string]datahelpers.DocumentIDs
	// anywhere holds the union of chrome and EVERY page, for rule 2.
	anywhere datahelpers.DocumentIDs
	// pathToPageID maps a normalised page path to the page id whose document
	// answers for it, for rule 3. Only pages with stored HTML appear.
	pathToPageID map[string]string
}

// newFragmentIndex builds the index from raw HTML already loaded by the caller:
// per-page component HTML (keyed by page id), the site chrome, and the
// normalised-path → page-id map for link targets.
func newFragmentIndex(pageHTML map[string]string, chrome string, pathToPageID map[string]string) fragmentIndex {
	idx := fragmentIndex{
		byPageID:     make(map[string]datahelpers.DocumentIDs, len(pageHTML)),
		pathToPageID: pathToPageID,
	}
	var all strings.Builder
	all.WriteString(chrome)
	for pageID, html := range pageHTML {
		idx.byPageID[pageID] = datahelpers.NewDocumentIDs(html + "\n" + chrome)
		all.WriteString("\n")
		all.WriteString(html)
	}
	idx.anywhere = datahelpers.NewDocumentIDs(all.String())
	return idx
}

// resolvesOnPage reports whether the fragment resolves in the given page's
// document, and whether a judgement was possible at all. A page with no stored
// HTML yields (false, false) — no document, no verdict.
func (f fragmentIndex) resolvesOnPage(pageID, fragment string) (resolved, judged bool) {
	ids, ok := f.byPageID[pageID]
	if !ok {
		return false, false
	}
	return ids.Satisfies(fragment), true
}

// resolvesAnywhere reports whether any page on the site satisfies the fragment
// (rule 2 — chrome surfaces). judged is false for a site with no stored HTML at
// all, where every verdict would be vacuous.
func (f fragmentIndex) resolvesAnywhere(fragment string) (resolved, judged bool) {
	if len(f.byPageID) == 0 {
		return false, false
	}
	return f.anywhere.Satisfies(fragment), true
}

// accumulateFragmentIssues records dead fragments for ONE component's HTML.
//
// It deliberately does not duplicate any judgement accumulateLinkIssues makes:
// it speaks only where that function is silent (anchor scope) or where that
// function has already accepted the href as pointing at a real, built page.
//
// targets is the same sitePageTargets the phantom scan uses, so "is this a real
// page" and "has it shipped" are answered by one definition on both arms.
func accumulateFragmentIssues(
	counts map[plKey]int,
	surface, pageName, pageID, slotName, html string,
	targets sitePageTargets,
	idx fragmentIndex,
) {
	shells := datahelpers.RuntimeFillSpans(html)
	for _, m := range datahelpers.HrefOffsets(html) {
		href := m.Value
		if datahelpers.IsNoopHref(href) {
			continue // check_dead_controls' remit, not a fragment that misses
		}
		if shells.Contains(m.Offset) {
			continue // hydrated client-side; its target need not exist at build time
		}

		path, fragment := datahelpers.SplitFragment(href)
		if fragment == "" {
			continue // no fragment, or "#" — nothing to resolve
		}

		var resolved, judged bool
		switch {
		case path == "":
			// Rule 1 / rule 2 — a bare "#x", resolved where it renders.
			if surface == "site_component" {
				resolved, judged = idx.resolvesAnywhere(fragment)
			} else {
				resolved, judged = idx.resolvesOnPage(pageID, fragment)
			}
		case datahelpers.ClassifyLinkScope(path) != datahelpers.LinkScopePage:
			continue // external, mailto or asset — not ours to resolve
		default:
			// Rule 3 — "/page.html#x". Silent unless the path itself is sound:
			// a phantom path is already reported as a phantom, and a
			// never-deployed target as unbuilt_internal_link.
			if !targets.valid.Contains(path) {
				continue
			}
			if _, unbuilt := targets.unbuilt[datahelpers.NormalizePagePath(path)]; unbuilt {
				continue
			}
			targetPageID, known := idx.pathToPageID[datahelpers.NormalizePagePath(path)]
			if !known {
				continue // target has no stored HTML — no document, no verdict
			}
			resolved, judged = idx.resolvesOnPage(targetPageID, fragment)
		}

		if judged && !resolved {
			counts[plKey{surface, pageName, pageID, slotName, href, "dead_fragment_link"}]++
		}
	}
}

// VerifyDeadFragmentLinkResolved re-checks one dead_fragment_link item at
// completion time, by the same rule that raised it: the defect is gone when the
// href is no longer present on the surface, OR its fragment now resolves.
//
// BOTH DISJUNCTS ARE THE FIX, and that is why this is not simply "does the
// fragment resolve now". The item's own suggested remedies are "point it at a
// real id" and "drop the fragment" — the second removes the href entirely, and
// a verifier that only asked about the id would refuse to complete an item that
// had been fixed the way the item itself recommended.
//
// It re-reads the LIVE rows rather than trusting the spec's occurrence count:
// the point of a verifier is that a handler saga can report success without
// touching anything (verifiers.go).
func VerifyDeadFragmentLinkResolved(ctx context.Context, db *sql.DB, target VerifyTarget, logger *zap.Logger) (VerifyResult, error) {
	href, _ := target.Spec["href"].(string)
	if href == "" {
		return VerifyResult{}, fmt.Errorf("dead_fragment_link verifier: item carries no href")
	}
	surface, _ := target.Spec["surface"].(string)
	_, fragment := datahelpers.SplitFragment(href)
	if fragment == "" {
		return VerifyResult{}, fmt.Errorf("dead_fragment_link verifier: href %q carries no fragment", href)
	}

	// 1. Is the offending href still on the surface it was found on?
	var stillPresent bool
	var err error
	if surface == "site_component" {
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM site_components
				WHERE site_id = $1 AND COALESCE(rendered_html, '') LIKE '%' || $2 || '%'
			)`, target.SiteID, `href="`+href+`"`).Scan(&stillPresent)
	} else {
		if target.PageID == nil {
			return VerifyResult{}, fmt.Errorf("dead_fragment_link verifier: page-surface item carries no page_id")
		}
		err = db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM page_components
				WHERE page_id = $1 AND COALESCE(rendered_html, '') LIKE '%' || $2 || '%'
			)`, *target.PageID, `href="`+href+`"`).Scan(&stillPresent)
	}
	if err != nil {
		return VerifyResult{}, fmt.Errorf("dead_fragment_link verifier: surface load failed: %w", err)
	}
	if !stillPresent {
		return VerifyResult{Resolved: true, Detail: fmt.Sprintf("href %q is no longer rendered on this %s", href, surface)}, nil
	}

	// 2. The href survives, so the fragment must now resolve. Rebuild the same
	//    document the check judged: the landing page's components plus chrome
	//    (or, for a chrome surface, any page on the site — rule 2).
	chrome, cErr := concatSiteComponentHTML(ctx, db, target.SiteID)
	if cErr != nil {
		return VerifyResult{}, fmt.Errorf("dead_fragment_link verifier: chrome load failed: %w", cErr)
	}

	path, _ := datahelpers.SplitFragment(href)
	var doc string
	switch {
	case surface == "site_component" && path == "":
		// Resolves anywhere on the site, matching the raising rule.
		doc, err = concatAllPageHTML(ctx, db, target.SiteID)
	case path == "":
		doc, err = concatPageHTML(ctx, db, *target.PageID)
	default:
		// A path#fragment item is filed against the page CONTAINING the link,
		// so the target page is resolved from the path, exactly as the check did.
		doc, err = concatPageHTMLByPath(ctx, db, target.SiteID, datahelpers.NormalizePagePath(path))
	}
	if err != nil {
		return VerifyResult{}, fmt.Errorf("dead_fragment_link verifier: document load failed: %w", err)
	}
	if strings.TrimSpace(doc) == "" {
		// No document, no verdict — the same silence the check keeps. Failing
		// open here would stamp complete on an unexamined page.
		return VerifyResult{}, fmt.Errorf("dead_fragment_link verifier: target document for %q has no stored HTML", href)
	}

	if datahelpers.NewDocumentIDs(doc + "\n" + chrome).Satisfies(fragment) {
		return VerifyResult{Resolved: true, Detail: fmt.Sprintf("fragment #%s now resolves on the target page", fragment)}, nil
	}
	logger.Info("dead_fragment_link verifier: fragment still unresolved",
		zap.String("item_id", target.ItemID.String()), zap.String("href", href))
	return VerifyResult{Resolved: false, Detail: fmt.Sprintf("href %q is still rendered and #%s still resolves to no element", href, fragment)}, nil
}

func concatSiteComponentHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID) (string, error) {
	var html sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT string_agg(COALESCE(rendered_html, ''), E'\n')
		FROM site_components WHERE site_id = $1`, siteID).Scan(&html)
	return html.String, err
}

func concatAllPageHTML(ctx context.Context, db *sql.DB, siteID uuid.UUID) (string, error) {
	var html sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT string_agg(COALESCE(pc.rendered_html, ''), E'\n')
		FROM page_components pc JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1`, siteID).Scan(&html)
	return html.String, err
}

func concatPageHTML(ctx context.Context, db *sql.DB, pageID uuid.UUID) (string, error) {
	var html sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT string_agg(COALESCE(rendered_html, ''), E'\n' ORDER BY position)
		FROM page_components WHERE page_id = $1`, pageID).Scan(&html)
	return html.String, err
}

// concatPageHTMLByPath resolves a normalised page path to its page and returns
// that page's document. The normalisation is datahelpers' own, so "/news.html"
// and "/news/index.html" are one target here exactly as they are in the check.
func concatPageHTMLByPath(ctx context.Context, db *sql.DB, siteID uuid.UUID, normPath string) (string, error) {
	var html sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT string_agg(COALESCE(pc.rendered_html, ''), E'\n' ORDER BY pc.position)
		FROM page_components pc
		JOIN pages p ON p.id = pc.page_id
		WHERE p.site_id = $1
		  AND p.status NOT IN ('deleted', 'archived')
		  AND lower(rtrim(regexp_replace(split_part(p.url, '#', 1), 'index\.html$', ''), '/')) = $2`,
		siteID, normPath).Scan(&html)
	return html.String, err
}
