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
	matches := hrefAttrRe.FindAllStringSubmatch(html, -1)
	hrefs := make([]string, 0, len(matches))
	for _, m := range matches {
		hrefs = append(hrefs, m[1])
	}
	return hrefs
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
		text := innerTagRe.ReplaceAllString(m[2], " ")
		text = strings.TrimSpace(multiWSRe.ReplaceAllString(text, " "))
		anchors = append(anchors, Anchor{Href: m[1], Text: text})
	}
	return anchors
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
const NeverDeployedPagePredicate = `deployed_at IS NULL AND COALESCE(build_status, '') <> 'deployed'`

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
