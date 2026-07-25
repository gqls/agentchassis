package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// The generated pages and files.
//
// The two section indexes are BUILT FROM THE MANIFEST rather than ported or
// composed by an agent. Both source indexes were already out of date on the day
// we read them — the tools index linked 51 of 55 tool directories and the learn
// index 13 of 23 articles — so four finished tools and ten finished articles
// were live and unreachable. Porting those indexes would have carried the same
// omissions across; asking an LLM to write them would risk a different set. A
// listing generated from the same manifest the importer uses cannot disagree
// with what actually exists.
//
// search.json and sitemap.xml come from the same source for the same reason.

func generateAll(cat *Catalogue, man *Manifest, outDir, domain string) error {
	genDir := filepath.Join(outDir, "generated")
	if err := os.MkdirAll(filepath.Join(outDir, "pages"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(genDir, 0o755); err != nil {
		return err
	}

	toolsHTML := renderToolsIndex(cat)
	learnHTML := renderLearnIndex(cat)

	for _, g := range []struct {
		name, url, title, desc, body string
	}{
		{
			"tools-index", "/tools/index.html",
			fmt.Sprintf("All %d Tools | webdesign.co.uk", len(cat.Tools)),
			fmt.Sprintf("%d free browser tools for designers and developers. Nothing to install, no account, and nothing leaves your machine.", len(cat.Tools)),
			toolsHTML,
		},
		{
			"learn-index", "/learn/index.html",
			"Learn | webdesign.co.uk",
			fmt.Sprintf("%d plain-English articles on the ideas behind the tools — colour, type, layout, performance, accessibility and statistics.", len(cat.Learn)),
			learnHTML,
		},
	} {
		frag := filepath.Join("pages", g.name+".html")
		if err := os.WriteFile(filepath.Join(outDir, frag), []byte(g.body), 0o644); err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(g.body))
		man.Pages = append(man.Pages, ManifestPage{
			Name: g.name, URL: g.url, PageType: "section-index",
			Title: g.title, MetaDescription: g.desc,
			Category: "index", Label: g.title, QATier: 3,
			Source: "(generated from the manifest)", Fragment: frag,
			SHA256:           hex.EncodeToString(sum[:]),
			VisibleTextChars: visibleTextChars(g.body),
		})
	}

	// search.json, sitemap.xml, robots.txt, 404.html: static files published
	// alongside the pages. The chassis has no sitemap or robots machinery, so
	// they are ours to generate.
	files := map[string]string{
		"search.json": renderSearchIndex(cat),
		"sitemap.xml": renderSitemap(man, domain),
		"robots.txt":  "User-agent: *\nAllow: /\n\nSitemap: https://" + domain + "/sitemap.xml\n",
		"404.html":    render404(),
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(genDir, name), []byte(body), 0o644); err != nil {
			return err
		}
		sum := sha256.Sum256([]byte(body))
		man.Assets = append(man.Assets, ManifestAsset{
			Dest: name, Generated: true, SHA256: hex.EncodeToString(sum[:]),
		})
	}
	return nil
}

func renderToolsIndex(cat *Catalogue) string {
	var b strings.Builder
	b.WriteString(`<section class="ported-page index-page" data-component="ported-page">` + "\n")
	b.WriteString(`<div class="index-intro">` + "\n")
	b.WriteString(fmt.Sprintf("<h1>%d tools</h1>\n", len(cat.Tools)))
	b.WriteString(`<p class="index-lede">Every one runs entirely in your browser. Nothing uploads, nothing is stored on a server, and none of them ask you to sign in.</p>` + "\n")
	b.WriteString("</div>\n")

	byCat := map[string][]CatalogueEntry{}
	for _, t := range cat.Tools {
		byCat[t.Category] = append(byCat[t.Category], t)
	}
	for _, c := range cat.ToolCategories {
		items := byCat[c]
		if len(items) == 0 {
			continue
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
		b.WriteString(`<section class="index-section">` + "\n")
		b.WriteString(fmt.Sprintf(
			`<div class="index-heading"><h2>%s</h2><span class="index-count">%d</span></div>`+"\n",
			html.EscapeString(c), len(items)))
		b.WriteString(`<div class="grid-cards">` + "\n")
		for _, t := range items {
			b.WriteString(card(t))
		}
		b.WriteString("</div>\n</section>\n")
	}
	b.WriteString("</section>\n")
	return b.String()
}

func renderLearnIndex(cat *Catalogue) string {
	var b strings.Builder
	b.WriteString(`<section class="ported-page index-page" data-component="ported-page">` + "\n")
	b.WriteString(`<div class="index-intro">` + "\n")
	b.WriteString("<h1>Learn</h1>\n")
	b.WriteString(`<p class="index-lede">The reasoning behind the tools: why a five-star average is the wrong sort order, why removing the focus ring breaks a site, and what a fractional unit is actually solving.</p>` + "\n")
	b.WriteString("</div>\n")

	byCat := map[string][]CatalogueEntry{}
	for _, e := range cat.Learn {
		byCat[e.Category] = append(byCat[e.Category], e)
	}
	var cats []string
	for c := range byCat {
		cats = append(cats, c)
	}
	// Biggest sections first, then alphabetically — so the page opens with
	// substance rather than with whichever slug sorts first.
	sort.Slice(cats, func(i, j int) bool {
		if len(byCat[cats[i]]) != len(byCat[cats[j]]) {
			return len(byCat[cats[i]]) > len(byCat[cats[j]])
		}
		return cats[i] < cats[j]
	})

	for _, c := range cats {
		items := byCat[c]
		sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })
		name := cat.LearnCategories[c]
		if name == "" {
			name = titleCase(c)
		}
		b.WriteString(`<section class="index-section">` + "\n")
		b.WriteString(fmt.Sprintf(
			`<div class="index-heading"><h2>%s</h2><span class="index-count">%d</span></div>`+"\n",
			html.EscapeString(name), len(items)))
		b.WriteString(`<div class="grid-cards">` + "\n")
		for _, e := range items {
			b.WriteString(card(e))
		}
		b.WriteString("</div>\n</section>\n")
	}
	b.WriteString("</section>\n")
	return b.String()
}

// card renders one index card. The whole card is the link (the warm system's
// own pattern), and the subtitle — present only on the near-duplicate pairs —
// is what stops a visitor having to open both to tell them apart.
func card(e CatalogueEntry) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`<a class="index-card" href="%s">`, html.EscapeString(e.targetURL())))
	if e.Eyebrow != "" {
		b.WriteString(fmt.Sprintf(`<span class="index-eyebrow">%s</span>`, html.EscapeString(e.Eyebrow)))
	}
	b.WriteString(fmt.Sprintf(`<h3>%s</h3>`, html.EscapeString(e.Label)))
	if e.Subtitle != "" {
		b.WriteString(fmt.Sprintf(`<p class="index-subtitle">%s</p>`, html.EscapeString(e.Subtitle)))
	}
	b.WriteString(fmt.Sprintf(`<p class="index-desc">%s</p>`, html.EscapeString(e.Description)))
	b.WriteString("</a>\n")
	return b.String()
}

// renderSearchIndex regenerates search.json from the catalogue. The field names
// match what the carried search.js filters on; changing them breaks the widget.
func renderSearchIndex(cat *Catalogue) string {
	var rows []SearchEntry
	for _, e := range allEntries(cat) {
		kw := append([]string{}, e.Keywords...)
		kw = append(kw, strings.Fields(strings.ToLower(e.Label))...)
		if e.Subtitle != "" {
			kw = append(kw, strings.Fields(strings.ToLower(e.Subtitle))...)
		}
		rows = append(rows, SearchEntry{
			Title:    e.Label,
			URL:      e.targetURL(),
			Category: e.Category,
			Keywords: strings.Join(dedupe(kw), " "),
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Title < rows[j].Title })
	b, _ := json.MarshalIndent(rows, "", "  ")
	return string(b) + "\n"
}

func renderSitemap(man *Manifest, domain string) string {
	var urls []string
	urls = append(urls, "/index.html")
	for _, p := range man.Pages {
		urls = append(urls, p.URL)
	}
	sort.Strings(urls)

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	seen := map[string]bool{}
	for _, u := range urls {
		if seen[u] {
			continue
		}
		seen[u] = true
		b.WriteString("  <url><loc>https://" + domain + u + "</loc></url>\n")
	}
	b.WriteString("</urlset>\n")
	return b.String()
}

// render404 is a complete standalone page, not a fragment: the object store
// serves it directly, outside the chassis assembly, so it carries its own head
// and its own minimal styling.
func render404() string {
	return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Page not found | webdesign.co.uk</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800&family=Fira+Code:wght@400;600&display=swap">
<style>
  :root { --bg:#f9f8f6; --panel:#fff; --primary:#5c6b5d; --text:#2b2b2b; --dim:#717171; --border:#edece9; }
  body { margin:0; background:var(--bg); color:var(--text);
         font-family:'Inter',system-ui,sans-serif; line-height:1.6;
         display:grid; place-items:center; min-height:100vh; padding:2rem; }
  .box { background:var(--panel); border:1px solid var(--border); border-radius:12px;
         box-shadow:0 4px 16px rgba(43,43,43,0.06); padding:3rem; max-width:34rem; }
  .code { font-family:'Fira Code',monospace; color:var(--primary); font-size:.85rem;
          letter-spacing:.08em; text-transform:uppercase; }
  h1 { font-size:2rem; font-weight:800; letter-spacing:-.02em; margin:.5rem 0 1rem; }
  p { color:var(--dim); margin:0 0 1.5rem; }
  a { color:var(--primary); font-weight:600; }
  ul { list-style:none; padding:0; margin:0; display:flex; gap:1.25rem; flex-wrap:wrap; }
</style>
</head>
<body>
  <div class="box">
    <span class="code">404</span>
    <h1>That page isn't here</h1>
    <p>It may have moved when this site absorbed its two predecessors, or the link may simply be wrong. The search box in the header covers every tool and article.</p>
    <ul>
      <li><a href="/tools/index.html">All tools</a></li>
      <li><a href="/learn/index.html">Learn</a></li>
      <li><a href="/index.html">Home</a></li>
    </ul>
  </div>
</body>
</html>
`
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		s = strings.Trim(strings.ToLower(s), ".,:;()\"'")
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}
