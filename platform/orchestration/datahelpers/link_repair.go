// FILE: platform/orchestration/datahelpers/link_repair.go
//
// In-body internal-link REPAIR, as opposed to detection (links.go) — bugs_open/079.
//
// links.go answers "does this href resolve to a real page". This file answers
// "so what do we do about it before the page ships". The two are deliberately
// separate files: links.go is the shared vocabulary read by three consumers, and
// a repair pass is a behaviour change that should not arrive as a passenger on
// an edit to it (the same reasoning that put DropDeadURLControls in its own
// file).
//
// WHY THIS EXISTS. The deploy gate (validate_page_content) has always detected
// dead in-body links accurately and then deployed the page anyway: a miss is
// filed at Severity "warning", and `valid` counts only blockers and errors. The
// written justification was that "the improvement loop resolves it" — that loop
// has been off since 2026-05 (bugs_open/083: phantom_internal_link has been
// detected 22 times and fixed ZERO times, ever). So the finding had nowhere to
// go and the 404 shipped.
//
// WHY REPAIR RATHER THAN BLOCK. Measured over the retained window (2026-07-13
// to 07-26) before choosing: 3 of 16 validated builds carried phantom links, 17
// instances across 15 unique targets, and all 3 deployed. Both affected pages
// were HOMEPAGES (oufe.com, webdesign.co.uk). Promoting the severity would have
// stopped two homepages from shipping at all, because a non-valid verdict
// returns an error and routes the build to mark_needs_review — no page, rather
// than a page with one bad link. Filing a work item instead was ruled out by
// bugs_open/083 (nothing drains 'detected') and bugs_open/077 (do not file items
// whose handler has no remit).
//
// THE RULE. Repair if the target really exists; otherwise remove the LINK and
// keep the TEXT. The page ships, the prose survives, the 404 dies.
//
//	<a href="/contact">Talk to us</a>       -> <a href="/contact.html">Talk to us</a>
//	                                           (only when /contact.html is a real
//	                                            pages.url — the stored value, never
//	                                            one this code assembled)
//	<a href="/invented">our pricing</a>     -> our pricing
//
// The rewrite arm exists because bugs_open/049 measured 8 fleet targets that
// EXIST and return 200 at their .html form while the writer emitted them bare.
// Those must not be unlinked — they are a href rewrite, and doing it here is the
// placement 049 asked for. It explicitly must NOT be done by making
// NormalizePagePath tolerant of a missing .html: "/contact" genuinely 404s on
// these sites, so a tolerant matcher would silence the detector on a live defect.

package datahelpers

import (
	"regexp"
	"strings"
)

// LinkRepair actions.
const (
	LinkRepairRewrite = "rewrite" // href replaced with the real stored pages.url
	LinkRepairUnlink  = "unlink"  // <a> dropped, inner content kept
)

// runtimeFillMarker exempts a shell whose hrefs are hydrated client-side. Its
// placeholder/empty hrefs are the mechanism, not a defect — the same exemption
// every other link check applies (check_phantom_internal_links, check_dead_controls,
// render_site_components_action). Whole-document, matching those callers.
const runtimeFillMarker = "data-runtime-fill"

// LinkRepair records one change made to the markup, so the caller can persist a
// durable account of what the gate did rather than only what it saw.
type LinkRepair struct {
	Href    string `json:"href"`              // the href as the writer emitted it
	NewHref string `json:"new_href,omitempty"` // rewrite only
	Action  string `json:"action"`
}

// repairAnchorRe matches one anchor with a quoted href, through to its closing
// tag, capturing the href VALUE (group 1) and the anchor's inner markup
// (group 2). Same shape as links.go's anchorRe and drop_dead_url_controls.go's
// deadAnchorRe: \b after <a so <area>/<abbr> are not matched; [^>]* confined to
// this tag's attributes so a sibling anchor is never swept in; anchors cannot
// nest, so the non-greedy .*? closes at this anchor's own </a>. (?is): dot spans
// newlines, names case-insensitive.
var repairAnchorRe = regexp.MustCompile(`(?is)<a\b[^>]*\shref\s*=\s*["']([^"']*)["'][^>]*>(.*?)</a>`)

// PageURLIndex maps a normalised page path to the REAL stored pages.url it came
// from. PageURLSet answers membership, which is all a detector needs; a repair
// has to emit a URL, and it must be one the database actually holds.
// bugs_closed/029 was a whole bug about an emitter that assembled plausible URLs
// instead of citing real ones, so the rewrite arm here never constructs a target
// — it can only hand back a string it was given.
type PageURLIndex map[string]string

// NewPageURLIndex builds the index from raw pages.url values. Where two rows
// normalise to the same path the later one wins; both are real, so either is a
// legitimate target.
func NewPageURLIndex(urls []string) PageURLIndex {
	ix := make(PageURLIndex, len(urls))
	for _, u := range urls {
		ix[NormalizePagePath(u)] = u
	}
	return ix
}

// Lookup resolves an href to the stored pages.url for that target.
func (ix PageURLIndex) Lookup(href string) (string, bool) {
	u, ok := ix[NormalizePagePath(href)]
	return u, ok
}

// Contains reports whether the href resolves to a real page, matching
// PageURLSet.Contains so a detector can use either type.
func (ix PageURLIndex) Contains(href string) bool {
	_, ok := ix[NormalizePagePath(href)]
	return ok
}

// hrefSuffix returns the #fragment / ?query tail of an href, or "". A rewrite
// swaps the PATH and must carry the tail across: /capabilities#approach becomes
// /capabilities.html#approach, not /capabilities.html. (Whether that fragment
// resolves to a real id is a separate, known gap — bugs_open/071 measured 24 of
// 25 anchored links fleet-wide pointing at an id that does not exist. Repairing
// the path turns those from 404 into inert, which is an improvement, not a fix.)
func hrefSuffix(href string) string {
	if i := strings.IndexAny(href, "#?"); i >= 0 {
		return href[i:]
	}
	return ""
}

// RepairPageLinks rewrites or removes the internal page links in html that do
// not resolve against index, returning the repaired markup and an account of
// every change. Output is byte-identical to the input when nothing needed
// repairing, so a clean page is never perturbed.
//
// Only LinkScopePage and LinkScopeEmpty are touched. External, mailto:/tel:,
// #fragment and asset hrefs are returned untouched — they are not page links and
// this index cannot judge them.
//
// The index MUST be one the caller loaded successfully. An empty index means
// "no page set", not "no pages exist", and is treated as a no-op: a failed
// query must never cause every link on the page to be stripped.
func RepairPageLinks(html string, index PageURLIndex) (string, []LinkRepair) {
	if html == "" || len(index) == 0 {
		return html, nil
	}
	if strings.Contains(html, runtimeFillMarker) {
		return html, nil
	}

	matches := repairAnchorRe.FindAllStringSubmatchIndex(html, -1)
	if len(matches) == 0 {
		return html, nil
	}

	var out strings.Builder
	var repairs []LinkRepair
	last := 0

	for _, m := range matches {
		fullStart, fullEnd := m[0], m[1]
		hrefStart, hrefEnd := m[2], m[3]
		innerStart, innerEnd := m[4], m[5]
		href := html[hrefStart:hrefEnd]

		unlink := func(recorded string) {
			out.WriteString(html[last:fullStart])
			out.WriteString(html[innerStart:innerEnd])
			last = fullEnd
			repairs = append(repairs, LinkRepair{Href: recorded, Action: LinkRepairUnlink})
		}

		switch ClassifyLinkScope(href) {
		case LinkScopePage:
			if index.Contains(href) {
				continue // real target — leave this anchor byte-identical
			}
			// The one mechanical repair: the target exists, the writer just
			// omitted the extension. Emit the stored URL, keeping any tail.
			if stored, ok := index.Lookup(NormalizePagePath(href) + ".html"); ok {
				newHref := stored + hrefSuffix(href)
				out.WriteString(html[last:hrefStart])
				out.WriteString(newHref)
				last = hrefEnd
				repairs = append(repairs, LinkRepair{Href: href, NewHref: newHref, Action: LinkRepairRewrite})
				continue
			}
			unlink(href)
		case LinkScopeEmpty:
			// An empty href renders as a control that goes nowhere. Same class,
			// same treatment — but the text stays, because in body prose the
			// anchor text is content. (Site chrome drops the whole control
			// instead — DropDeadURLControls — because a nav label without its
			// link is meaningless.)
			unlink("")
		}
	}

	if len(repairs) == 0 {
		return html, nil
	}
	out.WriteString(html[last:])
	return out.String(), repairs
}
