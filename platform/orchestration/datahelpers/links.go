// FILE: platform/orchestration/datahelpers/links.go
//
// Canonical internal-link extraction, classification and page-path
// normalisation. Single source of truth shared by:
//   - validate_page_content.go        (validateInternalLinks)   [deploy gate]
//   - discovery_checks/check_phantom_internal_links.go          [post-deploy audit]
//   - site_db_actions.go              (ExtractAndSyncLinksAction)[link inventory]
//
// Previously each of these decided "is this an internal page link, and does it
// point to a real page" with its own regex/SQL and its own normalisation, and
// they disagreed — the validator lowercased and appended ".html", the audit
// stripped "index.html" and trailing slashes, the inventory had no asset
// handling at all — so e.g. a "/tools/" directory link could read as a phantom
// under one path and valid under another. Consolidating means the rule changes
// here once and every consumer agrees.

package datahelpers

import (
	"fmt"
	"regexp"
	"strings"
)

// Link scope values. Plain strings (not a typed enum) to stay byte-compatible
// with the existing link_registry.scope column written by ExtractAndSyncLinksAction.
const (
	LinkScopeEmpty    = "empty"    // href=""
	LinkScopeAnchor   = "internal" // #fragment within the same page
	LinkScopeExternal = "external" // http(s):// or protocol-relative //
	LinkScopeMailto   = "mailto"   // mailto:/tel:/javascript:
	LinkScopeAsset    = "asset"    // static file (css/js/img/...), not a page
	LinkScopePage     = "page"     // internal navigable page link
)

var hrefAttrRe = regexp.MustCompile(`href=["']([^"']*)["']`)

// ExtractHrefs returns every href attribute value in the HTML, in document
// order, including empty ones (href=""). Duplicates are preserved; callers
// dedupe if they need to. Regex-based on purpose: the consumers here need only
// the raw href, and this keeps the helper dependency-free (no goquery).
func ExtractHrefs(html string) []string {
	refs := HrefOffsets(html)
	hrefs := make([]string, 0, len(refs))
	for _, r := range refs {
		hrefs = append(hrefs, r.Value)
	}
	return hrefs
}

// HrefRef is one href attribute plus the byte offset of the tag it was found in,
// so a caller can ask whether THIS link lies inside a runtime-fill shell rather
// than whether the document contains one anywhere (bugs_open/137). ExtractHrefs
// is this function with the offsets dropped, so the two can never disagree about
// which hrefs exist.
type HrefRef struct {
	Value  string
	Offset int
}

// HrefOffsets returns every href value in document order with its byte offset.
func HrefOffsets(html string) []HrefRef {
	matches := hrefAttrRe.FindAllStringSubmatchIndex(html, -1)
	refs := make([]HrefRef, 0, len(matches))
	for _, m := range matches {
		refs = append(refs, HrefRef{Value: html[m[2]:m[3]], Offset: m[0]})
	}
	return refs
}

// Anchor is one <a> element: its href and its visible text.
type Anchor struct {
	Href string
	Text string
}

var (
	anchorRe   = regexp.MustCompile(`(?is)<a\s[^>]*href=["']([^"']*)["'][^>]*>(.*?)</a>`)
	innerTagRe = regexp.MustCompile(`(?s)<[^>]*>`)
	multiWSRe  = regexp.MustCompile(`\s+`)
)

// ExtractAnchors returns every <a href=...>text</a> pair in document order,
// with inner markup stripped and whitespace collapsed from the text.
// ExtractHrefs drops the anchor text; checks that compare what a link SAYS
// against where it GOES (misdirected CTAs) need both halves. Same
// dependency-free regex approach as ExtractHrefs — nested <a> elements are
// invalid HTML and not handled.
func ExtractAnchors(html string) []Anchor {
	matches := anchorRe.FindAllStringSubmatch(html, -1)
	anchors := make([]Anchor, 0, len(matches))
	for _, m := range matches {
		anchors = append(anchors, Anchor{Href: m[1], Text: normaliseAnchorText(m[2])})
	}
	return anchors
}

// normaliseAnchorText strips inner markup and collapses whitespace. Extracted so
// the span-aware extractor in runtime_fill.go produces anchor text identical to
// ExtractAnchors' — two copies of this would be two definitions of "what the
// link says", and the misdirected-CTA check compares that text against a URL.
func normaliseAnchorText(inner string) string {
	text := innerTagRe.ReplaceAllString(inner, " ")
	return strings.TrimSpace(multiWSRe.ReplaceAllString(text, " "))
}

// ClassifyLinkScope is the single definition of what an href "is". Order
// matters: empty, then anchor/external/mailto, then asset, then page.
func ClassifyLinkScope(href string) string {
	switch {
	case href == "":
		return LinkScopeEmpty
	case strings.HasPrefix(href, "#"):
		return LinkScopeAnchor
	case strings.HasPrefix(href, "http://"),
		strings.HasPrefix(href, "https://"),
		strings.HasPrefix(href, "//"):
		return LinkScopeExternal
	case strings.HasPrefix(href, "mailto:"),
		strings.HasPrefix(href, "tel:"),
		strings.HasPrefix(href, "javascript:"):
		return LinkScopeMailto
	case IsAssetPath(href):
		return LinkScopeAsset
	default:
		return LinkScopePage
	}
}

// IsNoopHref reports whether an href is a no-op: the control renders as a
// link/button but clicking it goes nowhere. Bare "#" (no fragment name),
// "#!", and the javascript: void idioms. NOTE: href="" is NOT included — the
// empty class belongs to phantom_internal_links (empty_internal_href), and a
// named fragment ("#section") is a real in-page anchor. A <button> with no
// handler cannot be judged statically (JS may bind by class at runtime);
// that is Tier 4's claim, post-hydration.
func IsNoopHref(href string) bool {
	h := strings.TrimSpace(strings.ToLower(href))
	switch h {
	case "#", "#!", "javascript:void(0)", "javascript:void(0);", "javascript:;", "javascript:":
		return true
	}
	return false
}

// DeadControlAnchors returns the anchors in the HTML whose href is a no-op
// (IsNoopHref). Callers exempt runtime-fill shells (data-runtime-fill) —
// their placeholder hrefs are hydrated client-side.
func DeadControlAnchors(html string) []Anchor {
	var dead []Anchor
	for _, a := range ExtractAnchors(html) {
		if IsNoopHref(a.Href) {
			dead = append(dead, a)
		}
	}
	return dead
}

// IsAssetPath reports whether the path is a static asset rather than a page.
// Moved verbatim from validate_page_content.go so there is one definition.
func IsAssetPath(path string) bool {
	lower := strings.ToLower(path)

	if strings.HasPrefix(lower, "/assets/") ||
		strings.HasPrefix(lower, "/images/") ||
		strings.HasPrefix(lower, "/static/") ||
		strings.HasPrefix(lower, "/fonts/") ||
		strings.HasPrefix(lower, "/js/") ||
		strings.HasPrefix(lower, "/css/") {
		return true
	}

	assetExts := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".webp", ".svg",
		".ico", ".woff", ".woff2", ".ttf", ".eot", ".pdf", ".xml", ".json",
		".map", ".txt", ".mp4", ".webm",
	}
	for _, ext := range assetExts {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// SplitFragment splits an href into its path half and its #fragment, the one
// definition of "the fragment of an href" for every consumer that reasons about
// in-page anchors (the dead_fragment_link arm of check_phantom_internal_links,
// and anything after it). The fragment excludes the '#'; an href with no
// fragment returns ("", false) in the second value's place — callers get
// (path, "") — so "no fragment" and "#" (empty fragment, a no-op control that
// belongs to IsNoopHref/dead_controls, not to fragment resolution) are
// distinguishable from a named fragment. A ?query stays with the path half:
// per URL syntax the fragment follows the query, so the split is on the FIRST
// '#', not on IndexAny("#?") — NormalizePagePath's combined strip is correct
// for its question (path identity) and wrong for this one.
func SplitFragment(href string) (path, fragment string) {
	if i := strings.IndexByte(href, '#'); i >= 0 {
		return href[:i], href[i+1:]
	}
	return href, ""
}

// NormalizePagePath maps an internal page href OR a stored pages.url to one
// canonical form, so the two can be compared for equality. This replaces the
// two previously-divergent normalisers. Rules: drop #fragment and ?query,
// lowercase, strip a trailing "index.html", strip trailing slashes; the site
// root in any of its forms normalises to "/".
//
//	/contact.html        -> /contact.html
//	/contact/index.html  -> /contact
//	/contact/            -> /contact
//	/tools/index.html    -> /tools
//	/index.html  /  ""   -> /
func NormalizePagePath(href string) string {
	p := href
	if i := strings.IndexAny(p, "#?"); i >= 0 {
		p = p[:i]
	}
	p = strings.ToLower(strings.TrimSpace(p))
	p = strings.TrimSuffix(p, "index.html")
	p = strings.TrimRight(p, "/")
	if p == "" {
		return "/"
	}
	return p
}

// NeverDeployedPagePredicate is the SQL for "this page has never shipped, so a
// link to it is a live 404". It answers in SQL the question PageURLSet answers
// in Go, which is why it lives beside it: the renderer that WRITES links and the
// audit that FLAGS them must agree, or the platform reports links it authored.
//
// It is `deployed_at IS NULL`, NOT `build_status <> 'deployed'`. Measured across
// the fleet on 2026-07-20 against live HTTP:
//
//	needs_rebuild AND deployed_at IS NOT NULL   34 pages — 34/34 return 200
//	deployed_at IS NULL                         22 pages — every one tested 404s
//
// A page deployed once and later flagged needs_rebuild keeps serving its old
// artefact. Flagging on build_status alone would have been wrong about 34 of the
// 56 rows it selects. The discriminating pair, same build_status, opposite
// outcome:
//
//	gaswholesalers /fuel-pricing-framework.html  needs_rebuild  deployed_at NULL  -> 404
//	aao            /tools.html                   needs_rebuild  deployed_at set   -> 200
//
// `build_status <> 'deployed'` is retained as a second conjunct only to exclude
// the one fleet row that is 'deployed' yet never stamped (idea.uk — a
// bugs_open/040 shape, and it serves 200). See bugs_open/049 Correction 2 and
// bugs_open/052's addendum, which needs this same predicate.
//
// The column names are UNQUALIFIED, so this is valid only in a query whose FROM
// is `pages` with no ambiguous alias. A joined query needs a qualified variant;
// nobody has needed one yet.
// RELATED, and named here so the pair cannot drift unnoticed (bugs_open/185,
// raised by the council gate on the bugs_open/175 round):
// `queryresolve.FetchablePageEligibilitySQL` is the SAME judgement inverted —
// `AND (p.deployed_at IS NOT NULL OR p.build_status = 'deployed')`, i.e. exactly
// NOT(this) with an alias prefix. That is deliberate rather than drift (it is one
// of a documented family of three eligibility fragments), but two constants in two
// packages expressing one rule with no mutual reference is how drift begins.
// Change one, read the other. Whether they should be merged is bugs_open/185's
// fix candidate 2, open on purpose.
var NeverDeployedPagePredicate = NeverDeployedPagePredicateFor("")

// NeverDeployedPagePredicateFor is the same predicate with a table alias applied
// ("p" → `p.deployed_at IS NULL AND …`); pass "" for the bare form.
//
// It exists because the alias is the ONLY reason this judgement kept being
// re-typed by hand (bugs_open/185). Every consumer that aliases its `pages` table
// — which is most of them — could not use the bare constant, so it wrote the
// expression out again, and each hand-written copy is a chance to spell it
// `build_status = 'deployed'` and lose the 28 pages that shipped under another
// status. NeverDeployedPagePredicate itself is now DERIVED from this function, so
// the bare and aliased forms cannot drift from each other.
func NeverDeployedPagePredicateFor(alias string) string {
	q := ""
	if alias != "" {
		q = alias + "."
	}
	return fmt.Sprintf("%[1]sdeployed_at IS NULL AND COALESCE(%[1]sbuild_status, '') <> 'deployed'", q)
}

// PageHasShippedPredicateFor is the negation: "this page HAS been served, and is
// still serving whatever it last deployed". Use it for any question of the form
// *may I mutate this*, or *does this live page owe me a check* — NOT
// `build_status = 'deployed'`, which misses every page flagged `needs_rebuild`
// after a real deploy (bugs_closed/037; 35 of 46 such rows carry a deployed_at).
//
// Parenthesised, so it drops into a WHERE clause beside other conjuncts without
// the caller having to remember that the predicate is itself an AND.
func PageHasShippedPredicateFor(alias string) string {
	return "NOT (" + NeverDeployedPagePredicateFor(alias) + ")"
}

// PageWantedLivePredicateFor is the LIFECYCLE axis: "the platform still wants
// this page served". It is deliberately a separate axis from the build/shipped
// predicates above, because the two answer different questions and nothing
// keeps them in step — archiving a page sets `status='archived'` and leaves
// `build_status`/`deployed_at` untouched, so a selector keyed on a build
// column alone keeps choosing a retired page for ever (bugs_open/098: an
// archived page was re-rendered and re-published twice a day, which also made
// its retraction self-undoing).
//
// Pair this with whichever build-axis arm YOUR question needs; do not expect
// one combined helper. The estate legitimately runs three combinations
// (measured 2026-08-04): lifecycle alone (load_site_pages, plan_sections),
// lifecycle + shipped (request_render_audit — bugs_open/185 tranche 2 chose
// PageHasShippedPredicateFor precisely because `build_status = 'deployed'`
// was the wrong shipped-test for its question), and lifecycle + the exact
// 'deployed' STATE where that value is a state-machine arm
// (markPagesForRebuild transitions deployed→needs_rebuild). A merged
// "deployed-and-active" helper would misdescribe two of the three.
//
// Related, NOT interchangeable: `linkablePageStatusPredicate` in the actions
// package is `status NOT IN ('deleted','archived')` — "may be linked TO",
// which deliberately admits statuses this predicate rejects. Only two page
// statuses exist today (active 557 / archived 25, 2026-08-03), so the two
// currently select the same rows, but their fail-open directions differ the
// day a third status appears.
func PageWantedLivePredicateFor(alias string) string {
	q := ""
	if alias != "" {
		q = alias + "."
	}
	return q + "status = 'active'"
}

// InboundLinkSurfaces is the declared list of every table an inbound-link
// census over a site must read. It is the LOCKSTEP CONTRACT between the two
// censuses that ask "what links to this page":
//
//   - discovery_checks/check_orphan_pages.go (what IS unreachable — substring
//     match, over-matching on purpose: its safe failure is under-reporting
//     orphans);
//   - actions/retract_page_graph.go (what WOULD a retraction strand — quoted
//     match + content_data scan, precise on purpose: its unsafe failure is
//     refusing legitimate retractions).
//
// The QUERIES are deliberately different — same sources, different questions,
// opposite safe-failure directions; forcing them into one parameterised query
// was considered and rejected (council round 5a965452, 098 debt 3). What must
// not drift is THIS LIST: a link surface added to the platform (a new table
// holding hrefs or page references) that only one census learns about makes
// the two disagree about the same site. Each census hoists its SQL to a
// package-level var, and a lockstep test on each side asserts every entry
// here appears in that SQL — so extending this list breaks both tests until
// both censuses have answered for the new surface.
var InboundLinkSurfaces = []string{
	"site_nav_items",
	"site_components",
	"page_components",
}

// PageURLSet is a normalised set of real page URLs for membership tests.
type PageURLSet map[string]bool

// NewPageURLSet builds a set from raw pages.url values, normalising each.
func NewPageURLSet(urls []string) PageURLSet {
	set := make(PageURLSet, len(urls)+1)
	for _, u := range urls {
		set[NormalizePagePath(u)] = true
	}
	return set
}

// Contains reports whether an href resolves to a real page in the set.
// The href is normalised the same way the set members were.
func (s PageURLSet) Contains(href string) bool {
	return s[NormalizePagePath(href)]
}
