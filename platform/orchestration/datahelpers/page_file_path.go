// FILE: platform/orchestration/datahelpers/page_file_path.go
//
// ONE definition of "which file in the site repo does this page deploy to?"
// (bugs_open/125).
//
// Before this, five places derived a deploy path from a page and they had
// drifted: determineFilename (this package), rerender_single_page_action.go,
// get_pages_for_rerender_action.go and rerender_pages_actions.go all consulted
// pages.url first and got it broadly right, while determinePageFilename
// (git_deployer_actions.go) — the one the three BUILD pipelines actually reach
// — never consulted url at all and synthesised a path from the page's name.
// Measured 2026-07-31: 316 of 472 pages carrying a url (67%) disagree with
// '/'||name||'.html', so every /guides/…, /blog/… and /tools/… page would have
// been published to the wrong path, as a live, fetchable duplicate of itself.
//
// pages.url IS the authoritative path — it is what the nav, the sitemap, the
// link checker and every internal link resolve against. A deployer that writes
// somewhere else has not deployed the page; it has published a second one.
//
// The four "correct" copies were not identical either, and the differences are
// what the rules below encode:
//   - none of them guarded a FRAGMENT url, and one live row has one;
//   - none handled a directory-style url ("/guides/" → "/guides/index.html");
//   - three would rewrite "/foo.php" to "foo.php.html".

package datahelpers

import (
	"path"
	"strings"
)

// PageFilePathFromURL converts a page's canonical url (pages.url — site
// absolute, e.g. "/tools/password-entropy.html") into the repo-relative file
// path a deployer must write.
//
// ok == false means the url does NOT designate a file of its own. That is a
// legitimate state, not an error: the caller must fall back to its own
// name/slug chain rather than invent a path from a url that never named one.
//
// The returned path is always repo-relative with no leading slash, because the
// git adapter prefixes the site's domain itself — CommitToRepo does
// `data.Domain + "/" + path` (internal/adapters/git/github_client.go), so a
// leading slash produces "example.com//tools/x.html" and an empty segment in
// the GitHub tree. Every other path in this pipeline is already unprefixed
// (e.g. "assets/css/styles.css").
func PageFilePathFromURL(rawURL string) (string, bool) {
	u := strings.TrimSpace(rawURL)
	if u == "" {
		return "", false
	}

	// A FRAGMENT points INTO a page, it does not name one — and sanitising it
	// is worse than declining it. The live instance: idea.uk's
	// "tool-audience-check" has url "/tools.html#audience-check" while a
	// DIFFERENT page ("tools") owns "/tools.html", so stripping the fragment
	// would make a rebuild of one page overwrite the other's file. A query
	// string is the same class — a parameterised view of a page, not a file.
	if strings.ContainsAny(u, "#?") {
		return "", false
	}

	// An absolute or protocol-relative url addresses another origin; its path
	// is not ours to write. Backslashes are rejected outright: GitHub tree
	// paths take them literally, so they can only ever produce a junk filename.
	if strings.Contains(u, "://") || strings.HasPrefix(u, "//") || strings.Contains(u, `\`) {
		return "", false
	}

	u = strings.TrimPrefix(u, "/")

	// The site root, however it was spelled ("/" or "").
	if u == "" {
		return "index.html", true
	}

	// A directory-style url serves its index document.
	if strings.HasSuffix(u, "/") {
		u += "index.html"
	}

	// Anything path.Clean would rewrite is malformed for our purposes: "..",
	// "//", "./". Rejecting rather than cleaning keeps traversal out of a path
	// that is about to be written to a shared repo, and none exist live
	// (measured 2026-07-31: 0 of 472).
	cleaned := path.Clean(u)
	if cleaned != u || cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", false
	}

	// Append .html ONLY when the final segment has no extension at all. The url
	// is authoritative: if it names "/legacy/report.php" then the file must be
	// report.php or the canonical url 404s. This is the one point where this
	// helper deliberately differs from ensureHTMLExtension (git_deployer_
	// actions.go), which REPLACES an existing extension — correct for a page
	// name, wrong for a url.
	final := cleaned
	if i := strings.LastIndexByte(cleaned, '/'); i >= 0 {
		final = cleaned[i+1:]
	}
	if final == "" {
		return "", false
	}
	if !strings.Contains(final, ".") {
		cleaned += ".html"
	}

	return cleaned, true
}

// PageDeployFilename resolves a page's deploy path from its url, falling back
// to a name-derived path when the url cannot be used as one (see
// PageFilePathFromURL). This is the url-then-name form; callers with a richer
// fallback chain (slug/page_name/filename/id — determinePageFilename) call
// PageFilePathFromURL directly and keep their own chain.
//
// Both arguments may be empty; the result is never empty.
func PageDeployFilename(rawURL, name string) string {
	if p, ok := PageFilePathFromURL(rawURL); ok {
		return p
	}
	n := strings.TrimSpace(name)
	if n == "" || n == "index" || n == "home" {
		return "index.html"
	}
	if strings.HasSuffix(n, ".html") {
		return n
	}
	return n + ".html"
}
