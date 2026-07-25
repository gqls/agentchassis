package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// runTransform turns each source page into an owned-page fragment.
//
// What a fragment is: the page's own content and nothing else. The chassis
// supplies <head>, header, footer and the stylesheet at assembly time, so the
// sources' copy-pasted inline chrome must come OUT — otherwise every page would
// carry two headers, one of them pointing at the old domain.
//
// What a fragment keeps: the page-local <style> block (element-level <style> in
// body is valid and universally honoured, and it is where each tool's layout
// lives), and every <script>. Inline scripts are safe here — the assembly path
// concatenates rendered_html verbatim; the <script> refusal in
// create_report_page_action.go is local to that one action's prose inputs.

func runTransform(sitesDir, portDir, outDir, domain, only string) error {
	cat, err := loadCatalogue(filepath.Join(portDir, "catalogue.json"))
	if err != nil {
		return err
	}
	ov, err := loadOverrides(filepath.Join(portDir, "overrides.json"))
	if err != nil {
		return err
	}
	cm, err := loadColourMap(filepath.Join(portDir, "colour_map.json"))
	if err != nil {
		return err
	}
	engine, err := newColourEngine(cm)
	if err != nil {
		return err
	}

	urlMap := buildURLMap(cat)
	dropped := map[string]bool{}
	for _, d := range ov.Drop {
		dropped[d] = true
	}

	pagesDir := filepath.Join(outDir, "pages")
	if err := os.MkdirAll(pagesDir, 0o755); err != nil {
		return err
	}

	man := &Manifest{Domain: domain, Generator: "webdesignport/1"}
	assetSeen := map[string]bool{}

	entries := allEntries(cat)
	var failures []string

	for _, e := range entries {
		if dropped[e.Source] {
			continue
		}
		pageName := pageNameFor(e)
		if only != "" && !strings.Contains(pageName, only) {
			continue
		}

		p, rep, err := transformPage(sitesDir, e, ov, engine, urlMap, cat)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", e.Source, err))
			continue
		}

		frag := filepath.Join("pages", pageName+".html")
		if err := os.WriteFile(filepath.Join(outDir, frag), []byte(p.html), 0o644); err != nil {
			return err
		}

		sum := sha256.Sum256([]byte(p.html))
		mp := ManifestPage{
			Name:             pageName,
			URL:              e.targetURL(),
			PageType:         e.pageType(),
			Title:            p.title,
			MetaDescription:  p.metaDescription,
			Category:         e.Category,
			Label:            e.Label,
			Subtitle:         e.Subtitle,
			Description:      e.Description,
			Eyebrow:          e.Eyebrow,
			Keywords:         e.Keywords,
			QATier:           e.QATier,
			Source:           e.Source,
			Fragment:         frag,
			SHA256:           hex.EncodeToString(sum[:]),
			VisibleTextChars: visibleTextChars(p.html),
			Links:            p.links,
			SiblingAssets:    p.siblings,
			RuntimeFill:      ov.Pages[e.Source].RuntimeFill,
		}

		// The assembly's visible-text floor. Below it, the section is silently
		// dropped and the published page is an empty shell that still reports
		// success — so fail here instead, where it is cheap to see.
		if mp.VisibleTextChars <= 10 && !mp.RuntimeFill {
			failures = append(failures, fmt.Sprintf(
				"%s: only %d visible chars — assembly would drop this section. "+
					"Fix the content or set runtime_fill in overrides.json",
				e.Source, mp.VisibleTextChars))
			continue
		}

		man.Pages = append(man.Pages, mp)

		// Sibling assets (a tool's own engine JS) travel with the page.
		for _, s := range p.siblings {
			srcRel := path.Join(path.Dir(e.Source), s)
			dest := path.Join(path.Dir(strings.TrimPrefix(e.targetURL(), "/")), s)
			if assetSeen[dest] {
				continue
			}
			assetSeen[dest] = true

			// A sibling the source repo does not actually have must be one we
			// wrote ourselves — vibe-equalizer's state.js, the file whose
			// absence had that tool throwing on load since it shipped. Resolve
			// it against port/site_assets, and fail loudly if it is in neither
			// place, because the alternative is publishing a page whose script
			// 404s.
			if _, statErr := os.Stat(filepath.Join(sitesDir, filepath.FromSlash(srcRel))); statErr != nil {
				generated := filepath.Join(portDir, "site_assets", filepath.FromSlash(dest))
				if _, gerr := os.Stat(generated); gerr != nil {
					failures = append(failures, fmt.Sprintf(
						"%s: references %q, which exists neither in the sources nor in port/site_assets",
						e.Source, s))
					continue
				}
				man.Assets = append(man.Assets, ManifestAsset{Dest: dest, Generated: true})
				continue
			}
			man.Assets = append(man.Assets, ManifestAsset{Source: srcRel, Dest: dest})
		}

		for _, u := range rep.sortedUnmapped() {
			man.Warnings = append(man.Warnings,
				fmt.Sprintf("%s: unmapped colour %s (x%d)", pageName, u, rep.unmapped[u]))
		}
		for h, n := range rep.inScript {
			man.Warnings = append(man.Warnings, fmt.Sprintf(
				"%s: %s appears %dx inside a <script> — left alone (tool data, not skin). "+
					"If it IS skin, add a rewrite override.", pageName, h, n))
		}
	}

	if len(failures) > 0 {
		for _, f := range failures {
			fmt.Fprintln(os.Stderr, "FAIL "+f)
		}
		return fmt.Errorf("%d page(s) failed to transform", len(failures))
	}

	// Images ACTUALLY REFERENCED by the ported pages.
	collectReferencedImages(sitesDir, outDir, man, assetSeen)

	// Everything under port/site_assets mirrors the published site root: the
	// compat stylesheet, the vibe-equalizer fix, anything else we author.
	collectPortAssets(portDir, man, assetSeen)

	// The two section indexes, search.json, sitemap.xml, robots.txt and 404.html
	// are generated from the manifest itself — see generate.go for why.
	if only == "" {
		if err := generateAll(cat, man, outDir, domain); err != nil {
			return fmt.Errorf("generate: %w", err)
		}
	}

	sort.Slice(man.Pages, func(i, j int) bool { return man.Pages[i].Name < man.Pages[j].Name })
	if err := writeJSON(filepath.Join(outDir, "manifest.json"), man); err != nil {
		return err
	}

	fmt.Printf("transformed %d pages, %d assets, %d warnings -> %s\n",
		len(man.Pages), len(man.Assets), len(man.Warnings), outDir)
	for _, w := range man.Warnings {
		fmt.Println("  warn: " + w)
	}
	return nil
}

type pageResult struct {
	html            string
	title           string
	metaDescription string
	links           []string
	siblings        []string
}

func transformPage(sitesDir string, e CatalogueEntry, ov *Overrides, engine *colourEngine,
	urlMap map[string]string, cat *Catalogue) (*pageResult, *sweepReport, error) {

	po := ov.Pages[e.Source]
	rep := newSweepReport()

	doc, err := parseFile(filepath.Join(sitesDir, e.Source))
	if err != nil {
		return nil, rep, err
	}

	// ---- head: title + description, then the page-local <style> ------------
	res := &pageResult{}
	if t := findFirst(doc, byAtom(atom.Title)); t != nil {
		res.title = cleanTitle(textOf(t), ov.BrandSuffix)
	}
	if m := findFirst(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.Meta &&
			strings.EqualFold(attr(n, "name"), "description")
	}); m != nil {
		res.metaDescription = strings.TrimSpace(attr(m, "content"))
	}
	if po.Title != "" {
		res.title = po.Title
	}
	if po.MetaDescription != "" {
		res.metaDescription = po.MetaDescription
	}

	head := findFirst(doc, byAtom(atom.Head))
	body := findFirst(doc, byAtom(atom.Body))
	if body == nil {
		return nil, rep, fmt.Errorf("no <body>")
	}

	var styles []string
	if head != nil {
		// Sibling stylesheets get INLINED, not published as assets.
		//
		// Three ported tools (blueprint-compiler, micro-cms, vibe-equalizer)
		// keep their CSS in a `style.css` next to their index.html rather than
		// in a <style> block. The first version of this transform kept only
		// <style> and dropped every <link>, so those three would have shipped
		// completely unstyled — and they would have LOOKED fine in the manifest,
		// because nothing else about them is wrong.
		//
		// Inlining rather than copying also puts them through the colour sweep,
		// which they need more than most: they carry the dark-terminal skin.
		// The site-wide sheets (global.css, ../css/main.css) are excluded — the
		// chassis supplies the site stylesheet.
		for _, l := range findAll(head, byAtom(atom.Link)) {
			if !strings.EqualFold(attr(l, "rel"), "stylesheet") {
				continue
			}
			href := attr(l, "href")
			if !isRelative(href) || isSiteWideStylesheet(href) {
				continue
			}
			cssPath := filepath.Join(sitesDir, filepath.Dir(e.Source), filepath.FromSlash(href))
			b, rerr := os.ReadFile(cssPath)
			if rerr != nil {
				return nil, rep, fmt.Errorf("sibling stylesheet %q: %w", href, rerr)
			}
			styles = append(styles, "/* inlined from "+href+" */\n"+string(b))
		}
		for _, s := range findAll(head, byAtom(atom.Style)) {
			styles = append(styles, textOf2(s))
		}
	}
	// A few pages carry a <style> in the body too.
	for _, s := range findAll(body, byAtom(atom.Style)) {
		styles = append(styles, textOf2(s))
		remove(s)
	}

	// ---- rescue anything the strip would take with it ----------------------
	var rescued []string
	for _, sel := range po.KeepSelectors {
		s := parseSelector(sel)
		for _, n := range findAll(body, s.matches) {
			rescued = append(rescued, renderNode(n))
			remove(n)
		}
	}

	// ---- strip the sources' inline chrome ----------------------------------
	strip := append(append([]string{}, ov.StripSelectors...), po.ExtraStrip...)
	for _, sel := range strip {
		s := parseSelector(sel)
		for _, n := range findAll(body, s.matches) {
			remove(n)
		}
	}

	// ---- scripts: harvested from the WHOLE BODY, before the content root ----
	//
	// Both sites put their <script> tags after </main>, at body level. The first
	// version of this took the content root first and only extracted scripts
	// from inside it, so every tool's engine was silently discarded — 60-odd
	// interactive pages reduced to static markup, with the transform reporting
	// 97 pages and 0 warnings. Nothing about the output looked wrong; the pages
	// were simply dead. Caught by grepping a fragment for <script> rather than
	// trusting the counts.
	//
	// Scripts are pulled out here for a second reason too: the colour sweep must
	// never see JavaScript. Half these tools generate colours for a living, and
	// rewriting a canvas fill or a palette seed would break the tool while
	// looking like a successful reskin.
	var scripts []scriptTag
	for _, n := range findAll(body, byAtom(atom.Script)) {
		st := scriptTag{src: attr(n, "src"), typ: attr(n, "type"), body: textOf2(n)}
		if attr(n, "defer") != "" {
			st.defer_ = true
		}
		remove(n)

		switch {
		case st.src == "":
			engine.noteScriptColours(st.body, rep)
		case isSupersededSiteScript(st.src):
			// The old global search engine. The chassis header supplies search
			// now; carrying this too would put two engines on every page, both
			// fighting over #globalSearch.
			continue
		case isRelative(st.src):
			st.src = rewriteSiblingSrc(st.src)
			res.siblings = append(res.siblings, st.src)
		}
		scripts = append(scripts, st)
	}

	// ---- the content root --------------------------------------------------
	// Site A pages put content in <main>. Site B TOOL pages have no <main> at
	// all — their layout div sits directly in <body> — so fall back to whatever
	// is left in the body once chrome is gone.
	contentRoot := findFirst(body, byAtom(atom.Main))
	var contentHTML string
	if contentRoot != nil {
		contentHTML = renderChildren(contentRoot)
	} else {
		contentHTML = renderChildren(body)
	}

	// ---- colour sweep: stylesheets, then inline style="" attributes --------
	if !po.SkipColourSweep {
		for i := range styles {
			styles[i] = engine.sweepStylesheet(styles[i], rep)
		}
		contentHTML = sweepInlineStyles(contentHTML, engine, rep)
		for i := range rescued {
			rescued[i] = sweepInlineStyles(rescued[i], engine, rep)
		}
	}

	// ---- link rewrite ------------------------------------------------------
	contentHTML = rewriteLinks(contentHTML, e, urlMap, ov)
	for i := range rescued {
		rescued[i] = rewriteLinks(rescued[i], e, urlMap, ov)
	}

	// ---- per-page exact-string rewrites ------------------------------------
	for _, rw := range po.Rewrites {
		if !strings.Contains(contentHTML, rw.From) {
			return nil, rep, fmt.Errorf(
				"override rewrite never matched: %q. The source prose has drifted; "+
					"re-read the page and update overrides.json", rw.From)
		}
		contentHTML = strings.ReplaceAll(contentHTML, rw.From, rw.To)
	}

	// ---- counted placeholders ----------------------------------------------
	// Any figure about the site's own size is substituted from the catalogue,
	// never typed. Written after a near-miss: the about page's replacement stat
	// was hand-typed as "64 Tools" when the real count is 63 — an invented
	// statistic introduced by the very edit that was removing invented
	// statistics. A number that cannot be typed cannot drift.
	//
	// This runs AFTER the rewrites, because a rewrite is what puts the
	// placeholder there. Running it before left a literal {{TOOL_COUNT}} on the
	// published page — caught by reading the output rather than the exit code,
	// which was 0 both times.
	contentHTML = strings.ReplaceAll(contentHTML, "{{TOOL_COUNT}}", fmt.Sprint(len(cat.Tools)))
	contentHTML = strings.ReplaceAll(contentHTML, "{{LEARN_COUNT}}", fmt.Sprint(len(cat.Learn)))

	// ---- assemble the fragment ---------------------------------------------
	var sb strings.Builder
	sb.WriteString(`<section class="ported-page" data-component="ported-page">` + "\n")
	for _, s := range styles {
		if strings.TrimSpace(s) == "" {
			continue
		}
		sb.WriteString("<style>\n" + strings.TrimSpace(s) + "\n</style>\n")
	}
	for _, r := range rescued {
		sb.WriteString(r + "\n")
	}
	sb.WriteString(strings.TrimSpace(contentHTML) + "\n")
	for _, s := range scripts {
		sb.WriteString(s.render() + "\n")
	}
	sb.WriteString("</section>\n")

	res.html = sb.String()

	// ---- old-domain mentions in CONTENT ------------------------------------
	// Link hrefs were made root-relative above. What is left is prose and sample
	// data that still names the source sites: regex-tester's demo string
	// (support@website-design.com), social-card's URL placeholder, and
	// cors-scraping's worked example. Those should read as the new site — a
	// tool that ships examples pointing at a different domain is just wrong —
	// so they are rewritten here rather than exempted. Applied to the assembled
	// fragment, scripts included, because regex-tester carries the same sample
	// string in both its markup and its JS initialiser.
	for _, old := range []string{"website-design.com", "WebSiteDesign.com", "websitedesign.com"} {
		res.html = strings.ReplaceAll(res.html, old, "webdesign.co.uk")
	}

	res.links = internalLinks(res.html)

	// Forbidden references: a surviving link to something we dropped would be a
	// live 404, and a surviving old-domain mention would advertise the wrong site.
	for _, f := range ov.ForbiddenRefs {
		if strings.Contains(res.html, f) {
			return nil, rep, fmt.Errorf(
				"fragment still references %q, which is not being ported. "+
					"Add a rewrite override for this page", f)
		}
	}
	return res, rep, nil
}

// ---------------------------------------------------------------------------
// scripts
// ---------------------------------------------------------------------------

type scriptTag struct {
	src    string
	typ    string
	body   string
	defer_ bool
}

func (s scriptTag) render() string {
	var b strings.Builder
	b.WriteString("<script")
	if s.typ != "" {
		b.WriteString(` type="` + s.typ + `"`)
	}
	if s.src != "" {
		b.WriteString(` src="` + rewriteSiblingSrc(s.src) + `"`)
	}
	if s.defer_ {
		b.WriteString(" defer")
	}
	b.WriteString(">")
	b.WriteString(s.body)
	b.WriteString("</script>")
	return b.String()
}

// isSupersededSiteScript identifies scripts the chassis now provides itself.
// Carrying the old global search engine alongside the header's would put two
// engines on the page, both bound to #globalSearch.
func isSupersededSiteScript(src string) bool {
	s := strings.ToLower(src)
	return strings.HasSuffix(s, "/assets/js/search.js") ||
		strings.HasSuffix(s, "/js/main.js")
}

// rewriteSiblingSrc fixes the one broken relative path in the sources:
// vibe-equalizer loads ../../js/state.js, a file that does not exist. On the
// new site state.js ships beside the tool.
func rewriteSiblingSrc(src string) string {
	if strings.HasSuffix(src, "/js/state.js") {
		return "state.js"
	}
	return src
}

// isSiteWideStylesheet identifies the two sources' own global sheets, which the
// chassis stylesheet plus port-compat.css replace.
func isSiteWideStylesheet(href string) bool {
	h := strings.ToLower(href)
	return strings.HasSuffix(h, "assets/css/global.css") || strings.HasSuffix(h, "css/main.css")
}

func isRelative(u string) bool {
	return u != "" && !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") &&
		!strings.HasPrefix(u, "//") && !strings.HasPrefix(u, "/")
}

// ---------------------------------------------------------------------------
// inline style="" sweep and link rewriting
// ---------------------------------------------------------------------------

func sweepInlineStyles(fragment string, engine *colourEngine, rep *sweepReport) string {
	root, err := parseFragment(fragment)
	if err != nil {
		return fragment
	}
	for _, n := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && attr(n, "style") != ""
	}) {
		setAttr(n, "style", engine.sweepDeclarations(attr(n, "style"), rep))
	}
	return renderChildren(root)
}

func rewriteLinks(fragment string, e CatalogueEntry, urlMap map[string]string, ov *Overrides) string {
	root, err := parseFragment(fragment)
	if err != nil {
		return fragment
	}
	for _, n := range findAll(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.A && attr(n, "href") != ""
	}) {
		setAttr(n, "href", rewriteHref(attr(n, "href"), e, urlMap, ov))
	}
	return renderChildren(root)
}

func rewriteHref(href string, e CatalogueEntry, urlMap map[string]string, ov *Overrides) string {
	h := strings.TrimSpace(href)
	if h == "" || strings.HasPrefix(h, "#") || strings.HasPrefix(h, "mailto:") {
		return href
	}

	// Absolute references to either source domain become root-relative.
	for _, host := range []string{
		"https://www.website-design.com", "https://website-design.com",
		"http://www.website-design.com", "http://website-design.com",
		"https://www.websitedesign.com", "https://websitedesign.com",
		"http://www.websitedesign.com", "http://websitedesign.com",
	} {
		if strings.HasPrefix(h, host) {
			h = strings.TrimPrefix(h, host)
			if h == "" {
				h = "/"
			}
		}
	}
	if strings.HasPrefix(h, "http://") || strings.HasPrefix(h, "https://") {
		return h // genuinely external, leave it
	}

	// Resolve relative links against the page's OWN source location, so that
	// Site B's ../../index.html and ../index.html land somewhere meaningful.
	if !strings.HasPrefix(h, "/") {
		h = "/" + path.Clean(path.Join(path.Dir(sourceSitePath(e.Source)), h))
	}

	if mapped, ok := ov.LinkMap[h]; ok {
		return mapped
	}
	if mapped, ok := urlMap[h]; ok {
		return mapped
	}
	return h
}

// sourceSitePath strips the site dir: "websitedesign.com/tools/x/index.html"
// becomes "/tools/x/index.html" — the path as that site served it.
func sourceSitePath(src string) string {
	parts := strings.SplitN(src, "/", 2)
	if len(parts) != 2 {
		return "/" + src
	}
	return "/" + parts[1]
}

// buildURLMap maps each source site's own URL to the new site's URL. This is
// what moves websitedesign.com's /guides/x.html into /learn/<cat>/x.html.
func buildURLMap(cat *Catalogue) map[string]string {
	m := map[string]string{}
	for _, e := range allEntries(cat) {
		m[sourceSitePath(e.Source)] = e.targetURL()
		// Also accept the directory form without index.html.
		if strings.HasSuffix(e.Source, "/index.html") {
			m[strings.TrimSuffix(sourceSitePath(e.Source), "index.html")] = e.targetURL()
		}
	}
	// The section indexes and home, which every source page links to.
	m["/"] = "/index.html"
	m["/index.html"] = "/index.html"
	m["/tools/"] = "/tools/index.html"
	m["/tools/index.html"] = "/tools/index.html"
	m["/learn/"] = "/learn/index.html"
	m["/learn/index.html"] = "/learn/index.html"
	m["/guides/"] = "/learn/index.html"
	m["/about/"] = "/about/index.html"
	m["/about/index.html"] = "/about/index.html"
	return m
}

func internalLinks(fragment string) []string {
	doc, err := html.Parse(strings.NewReader(fragment))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, n := range findAll(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.A
	}) {
		h := attr(n, "href")
		if !strings.HasPrefix(h, "/") {
			continue
		}
		if i := strings.Index(h, "#"); i > 0 {
			h = h[:i]
		}
		if !seen[h] {
			seen[h] = true
			out = append(out, h)
		}
	}
	sort.Strings(out)
	return out
}

// ---------------------------------------------------------------------------
// catalogue helpers
// ---------------------------------------------------------------------------

// allEntries is every catalogued page in one slice: tools, learn, standalone.
func allEntries(cat *Catalogue) []CatalogueEntry {
	out := append([]CatalogueEntry{}, cat.Tools...)
	out = append(out, cat.Learn...)
	return append(out, cat.Pages...)
}

func (e CatalogueEntry) isTool() bool {
	return strings.Contains(e.Source, "/tools/")
}

func (e CatalogueEntry) targetURL() string {
	if e.URLOverride != "" {
		return e.URLOverride
	}
	if e.isTool() {
		return "/tools/" + e.Slug + "/index.html"
	}
	return "/learn/" + e.Category + "/" + e.Slug + ".html"
}

func (e CatalogueEntry) pageType() string {
	if e.PageTypeOverride != "" {
		return e.PageTypeOverride
	}
	if e.isTool() {
		return "tool"
	}
	return "guide"
}

func pageNameFor(e CatalogueEntry) string {
	if e.URLOverride != "" {
		return e.Slug
	}
	if e.isTool() {
		return "tool-" + e.Slug
	}
	return "learn-" + e.Category + "-" + e.Slug
}

func cleanTitle(t, brandSuffix string) string {
	t = strings.TrimSpace(t)
	for _, tail := range []string{
		"| website-design.com", "| WebSiteDesign.com", "| websitedesign.com",
		"- website-design.com", "- WebSiteDesign.com",
	} {
		if i := strings.Index(t, tail); i >= 0 {
			t = strings.TrimSpace(t[:i])
		}
	}
	if brandSuffix != "" {
		t += " " + brandSuffix
	}
	return t
}

// textOf2 returns raw (uncollapsed) text — for <style> and <script> bodies,
// where collapsing whitespace would corrupt the content.
func textOf2(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	}
	return sb.String()
}

// collectReferencedImages publishes only the images the fragments actually use.
//
// Walking the whole assets/img tree instead pulled in 22 files, of which 6 are
// referenced. The other 16 are the ~1MB PNG originals of those same 6 (the
// pages all use the smaller content/*.jpg versions) and a set of pasteboard
// sample photos that nothing links — the pasteboard tool works on images you
// paste into it, and ships none. Publishing them cost several megabytes and a
// string of API failures for files nobody would ever request.
func collectReferencedImages(sitesDir, outDir string, man *Manifest, seen map[string]bool) {
	ref := map[string]bool{}
	for _, p := range man.Pages {
		b, err := os.ReadFile(filepath.Join(outDir, p.Fragment))
		if err != nil {
			continue
		}
		for _, m := range imgRefRe.FindAllString(string(b), -1) {
			ref[m] = true
		}
	}

	var dests []string
	for d := range ref {
		dests = append(dests, d)
	}
	sort.Strings(dests)

	for _, dest := range dests {
		if seen[dest] {
			continue
		}
		src := path.Join("website-design.com", dest)
		if _, err := os.Stat(filepath.Join(sitesDir, filepath.FromSlash(src))); err != nil {
			man.Warnings = append(man.Warnings,
				fmt.Sprintf("referenced image %s not found in the sources", dest))
			continue
		}
		seen[dest] = true
		man.Assets = append(man.Assets, ManifestAsset{Source: src, Dest: dest})
	}
}

// collectPortAssets walks port/site_assets, whose layout mirrors the published
// site root, and records each file as a generated asset.
func collectPortAssets(portDir string, man *Manifest, seen map[string]bool) {
	base := filepath.Join(portDir, "site_assets")
	_ = filepath.Walk(base, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(base, p)
		dest := filepath.ToSlash(rel)
		if seen[dest] {
			return nil
		}
		seen[dest] = true
		man.Assets = append(man.Assets, ManifestAsset{Dest: dest, Generated: true})
		return nil
	})
}

// imgRefRe finds image paths as the fragments actually write them.
var imgRefRe = regexp.MustCompile(`assets/img/[A-Za-z0-9/_.-]+\.(?:jpg|jpeg|png|gif|svg|webp|avif|ico)`)

// ---------------------------------------------------------------------------
// loaders
// ---------------------------------------------------------------------------

func loadCatalogue(p string) (*Catalogue, error) {
	var c Catalogue
	if err := readJSON(p, &c); err != nil {
		return nil, err
	}
	if len(c.Tools) == 0 {
		return nil, fmt.Errorf("%s: no tools — run `webdesignport harvest` first", p)
	}
	return &c, nil
}

func loadOverrides(p string) (*Overrides, error) {
	var o Overrides
	if err := readJSON(p, &o); err != nil {
		return nil, err
	}
	if o.Pages == nil {
		o.Pages = map[string]PageOverride{}
	}
	if o.LinkMap == nil {
		o.LinkMap = map[string]string{}
	}
	return &o, nil
}

func loadColourMap(p string) (*ColourMap, error) {
	var c ColourMap
	if err := readJSON(p, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func readJSON(p string, v any) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("%s: %w", p, err)
	}
	return nil
}
