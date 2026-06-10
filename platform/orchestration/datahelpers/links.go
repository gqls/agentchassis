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
