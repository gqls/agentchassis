// FILE: check_phantom_internal_links_fragments_test.go
//
// Tests for the dead_fragment_link arm. Every case here is either a live shape
// measured on the fleet on 2026-08-06 or one of the four deliberate silences —
// because the risk this arm carries is not "it misses one", it is "it files a
// finding against a page that works", and the silences are what prevent that.
//
// The no-op cases are asserted as hard as the damage case on purpose: a
// detector is only evidence if it could have come out the other way, and the
// same day this was written the estate's 66 fragment links ALL resolved. A test
// suite that only proves the arm fires would pass identically if the arm fired
// on everything.

package discovery_checks

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// oneSitePageTargets builds the target set the phantom scan would build for a
// site whose pages are all real and all deployed.
func oneSitePageTargets(urls ...string) sitePageTargets {
	return sitePageTargets{
		valid:   datahelpers.NewPageURLSet(append(urls, "/", "/index.html")),
		unbuilt: map[string]string{},
	}
}

func countsFor(t *testing.T, counts map[plKey]int, issue string) []plKey {
	t.Helper()
	var out []plKey
	for k := range counts {
		if k.issue == issue {
			out = append(out, k)
		}
	}
	return out
}

func TestDeadFragmentLinkFiresOnAMissingID(t *testing.T) {
	// The 071 shape: /capabilities.html#approach where no section emits an id.
	pageHTML := map[string]string{
		"page-1": `<a href="/capabilities.html#approach">Our approach</a>`,
		"page-2": `<section><h2>Capabilities</h2></section>`,
	}
	idx := newFragmentIndex(pageHTML, "<header></header>", map[string]string{
		"/capabilities.html": "page-2",
	})
	targets := oneSitePageTargets("/index.html", "/capabilities.html")

	counts := map[plKey]int{}
	accumulateFragmentIssues(counts, "page_component", "index", "page-1", "hero", pageHTML["page-1"], targets, idx)

	found := countsFor(t, counts, "dead_fragment_link")
	if len(found) != 1 {
		t.Fatalf("expected 1 dead_fragment_link, got %d (%v)", len(found), counts)
	}
	if found[0].href != "/capabilities.html#approach" {
		t.Errorf("finding names the wrong href: %q", found[0].href)
	}
	if found[0].pageID != "page-1" {
		t.Errorf("finding must be filed against the page CONTAINING the link, got %q", found[0].pageID)
	}
}

func TestDeadFragmentLinkSilentWhenTheIDIsThere(t *testing.T) {
	// The live idea.uk shape, which must stay silent: /tools.html#audience-check
	// where tools.html really does carry id="audience-check".
	pageHTML := map[string]string{
		"page-1": `<a href="/tools.html#audience-check">Try it</a>`,
		"page-2": `<section id="audience-check">…</section>`,
	}
	idx := newFragmentIndex(pageHTML, "", map[string]string{"/tools.html": "page-2"})
	targets := oneSitePageTargets("/index.html", "/tools.html")

	counts := map[plKey]int{}
	accumulateFragmentIssues(counts, "page_component", "index", "page-1", "hero", pageHTML["page-1"], targets, idx)

	if n := len(countsFor(t, counts, "dead_fragment_link")); n != 0 {
		t.Fatalf("a resolving fragment must file nothing, got %d findings: %v", n, counts)
	}
}

func TestBareFragmentResolvesAgainstItsOwnPageAndChrome(t *testing.T) {
	// The loancash/loanandmortgage shape: a "#content" skip-link whose target id
	// lives in a SIBLING component of the same page (or in the chrome), which is
	// why the whole-document rule exists.
	pageHTML := map[string]string{
		"page-1": `<a href="#content">Skip to content</a>` + "\n" + `<main id="content">…</main>`,
		"page-2": `<a href="#nowhere">Jump</a>`,
	}
	idx := newFragmentIndex(pageHTML, `<footer id="site-footer"></footer>`, nil)
	targets := oneSitePageTargets("/index.html")

	counts := map[plKey]int{}
	accumulateFragmentIssues(counts, "page_component", "index", "page-1", "nav", pageHTML["page-1"], targets, idx)
	if n := len(countsFor(t, counts, "dead_fragment_link")); n != 0 {
		t.Fatalf("#content resolves on its own page; got %d findings", n)
	}

	// Chrome ids count too — resolving against the page alone would flag this.
	counts = map[plKey]int{}
	accumulateFragmentIssues(counts, "page_component", "index", "page-1", "nav",
		`<a href="#site-footer">Footer</a>`, targets, idx)
	if n := len(countsFor(t, counts, "dead_fragment_link")); n != 0 {
		t.Fatalf("a fragment satisfied by the chrome must not be flagged; got %d", n)
	}

	// And the genuinely missing one on page-2 still fires.
	counts = map[plKey]int{}
	accumulateFragmentIssues(counts, "page_component", "about", "page-2", "body", pageHTML["page-2"], targets, idx)
	if n := len(countsFor(t, counts, "dead_fragment_link")); n != 1 {
		t.Fatalf("expected the missing #nowhere to fire, got %d", n)
	}
}

func TestChromeFragmentJudgedAcrossTheWholeSite(t *testing.T) {
	// Rule 2. A header skip-link renders on every page; it is a template defect
	// only when it resolves on NO page. Flagging per-page would file one item
	// per page for a link that works almost everywhere.
	chrome := `<a href="#content">Skip</a><a href="#never">Dead</a>`
	pageHTML := map[string]string{
		"page-1": `<main id="content">…</main>`,
		"page-2": `<article>no content id here</article>`,
	}
	idx := newFragmentIndex(pageHTML, chrome, nil)
	targets := oneSitePageTargets("/index.html")

	counts := map[plKey]int{}
	accumulateFragmentIssues(counts, "site_component", "", "", "header", chrome, targets, idx)

	found := countsFor(t, counts, "dead_fragment_link")
	if len(found) != 1 {
		t.Fatalf("expected exactly the site-wide-dead fragment to fire, got %d: %v", len(found), counts)
	}
	if found[0].href != "#never" {
		t.Errorf("wrong href flagged: %q — #content resolves on page-1 and must be silent", found[0].href)
	}
}

func TestDeadFragmentLinkKeepsItsFourSilences(t *testing.T) {
	pageHTML := map[string]string{"page-1": `<section id="real">…</section>`}
	idx := newFragmentIndex(pageHTML, "", map[string]string{"/index.html": "page-1"})
	targets := oneSitePageTargets("/index.html", "/built.html")
	// A page that exists but has never been deployed, and one with no stored HTML.
	targets.unbuilt["/unbuilt.html"] = "page-unbuilt"
	targets.valid = datahelpers.NewPageURLSet([]string{"/", "/index.html", "/built.html", "/unbuilt.html", "/nohtml.html"})

	cases := []struct {
		name string
		html string
	}{
		{"noop hrefs are dead_controls' remit", `<a href="#">x</a><a href="#!">y</a>`},
		{"runtime-fill shells are hydrated client-side",
			`<div data-runtime-fill="cards"><a href="#not-yet">card</a></div>`},
		{"a phantom path is reported once, as a phantom", `<a href="/ghost.html#any">x</a>`},
		{"an unbuilt target is reported once, as unbuilt", `<a href="/unbuilt.html#any">x</a>`},
		{"a target with no stored HTML gets no verdict", `<a href="/nohtml.html#any">x</a>`},
		{"external and asset hrefs are not ours", `<a href="https://x.com/p#f">x</a><a href="/assets/a.pdf#page=2">y</a>`},
		{"an href with no fragment is not this arm's business", `<a href="/built.html">x</a>`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			counts := map[plKey]int{}
			accumulateFragmentIssues(counts, "page_component", "index", "page-1", "slot", tc.html, targets, idx)
			if n := len(countsFor(t, counts, "dead_fragment_link")); n != 0 {
				t.Fatalf("expected silence, got %d findings: %v", n, counts)
			}
		})
	}
}

func TestFragmentPresenceUsesTheSharedIDDefinition(t *testing.T) {
	// The lockstep with check_orphan_element_refs: an id created at runtime, or
	// interpolated from data, counts as present. If these two ever disagree
	// about what an id is, one of them files findings against working pages —
	// which is the false positive OrphanElementRefs' header records paying for.
	dynamic := `<a href="#built-later">go</a><script>el.id = 'built-later';</script>`
	interpolated := `<a href="#slider-blur">go</a>` +
		"<script>group.innerHTML = `<input id=\"${f.name}\">`; const fields=['slider-blur'];</script>"

	for name, html := range map[string]string{"runtime-assigned id": dynamic, "interpolated id": interpolated} {
		t.Run(name, func(t *testing.T) {
			idx := newFragmentIndex(map[string]string{"p": html}, "", nil)
			counts := map[plKey]int{}
			accumulateFragmentIssues(counts, "page_component", "index", "p", "slot", html, oneSitePageTargets("/index.html"), idx)
			if n := len(countsFor(t, counts, "dead_fragment_link")); n != 0 {
				t.Fatalf("expected silence (shared id definition), got %d", n)
			}
			if !datahelpers.NewDocumentIDs(html).Satisfies(map[string]string{
				"runtime-assigned id": "built-later", "interpolated id": "slider-blur"}[name]) {
				t.Error("DocumentIDs disagrees with the arm about the same id")
			}
		})
	}
}

func TestSplitFragmentKeepsTheQueryWithThePath(t *testing.T) {
	// Per URL syntax the fragment follows the query, so a naive IndexAny("#?")
	// split — which is what NormalizePagePath correctly does for ITS question —
	// would hand this arm "q=1#frag" as a fragment and flag a working link.
	path, frag := datahelpers.SplitFragment("/search.html?q=1#results")
	if path != "/search.html?q=1" || frag != "results" {
		t.Fatalf("SplitFragment(%q) = (%q, %q)", "/search.html?q=1#results", path, frag)
	}
	if p, f := datahelpers.SplitFragment("/a.html"); p != "/a.html" || f != "" {
		t.Fatalf("no-fragment href mis-split: (%q, %q)", p, f)
	}
	if p, f := datahelpers.SplitFragment("#"); p != "" || f != "" {
		t.Fatalf(`"#" must yield an empty fragment, got (%q, %q)`, p, f)
	}
}
