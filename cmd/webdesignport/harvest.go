package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// harvest bootstraps the catalogue from the source sites' own index pages.
//
// website-design.com/tools/index.html already groups every listed tool under a
// `.category-title` heading and gives each a card with an <h3> label, a <p>
// one-liner and a "Launch" link — i.e. the entire curation we would otherwise
// hand-type for 51 tools. Same for learn/index.html. So we read it rather than
// retype it, then report what the indexes DO NOT link (the orphans) so those get
// hand-written entries and nothing is silently left out.
//
// The output is a starting point that is then hand-edited: subtitles for the
// near-duplicate pairs, categories for the additions, QA tiers.

func runHarvest(sitesDir, portDir string) error {
	cat := &Catalogue{
		LearnCategories: map[string]string{},
	}

	// ---- Site A tools index -------------------------------------------------
	toolsIdx := filepath.Join(sitesDir, "website-design.com/tools/index.html")
	if err := harvestToolCards(toolsIdx, cat); err != nil {
		return fmt.Errorf("harvest tools index: %w", err)
	}

	// ---- Site A learn index -------------------------------------------------
	learnIdx := filepath.Join(sitesDir, "website-design.com/learn/index.html")
	if err := harvestLearnCards(learnIdx, cat); err != nil {
		return fmt.Errorf("harvest learn index: %w", err)
	}

	// ---- Report what exists on disk but is not in either index --------------
	var orphans []string
	linked := map[string]bool{}
	for _, e := range cat.Tools {
		linked[e.Source] = true
	}
	for _, e := range cat.Learn {
		linked[e.Source] = true
	}

	// Site A tool dirs
	toolDirs, _ := filepath.Glob(filepath.Join(sitesDir, "website-design.com/tools/*/index.html"))
	for _, p := range toolDirs {
		rel := relTo(sitesDir, p)
		if !linked[rel] {
			orphans = append(orphans, "TOOL (unlinked by source index): "+rel)
		}
	}
	// Site A articles
	_ = filepath.Walk(filepath.Join(sitesDir, "website-design.com/learn"), func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() || !strings.HasSuffix(p, ".html") {
			return nil
		}
		rel := relTo(sitesDir, p)
		if strings.HasSuffix(rel, "learn/index.html") || linked[rel] {
			return nil
		}
		orphans = append(orphans, "ARTICLE (unlinked by source index): "+rel)
		return nil
	})
	// Site B: no usable index (its homepage is a dashboard), so everything is
	// "unlisted" by definition — enumerate it for the hand-written entries.
	bTools, _ := filepath.Glob(filepath.Join(sitesDir, "websitedesign.com/tools/*/index.html"))
	for _, p := range bTools {
		orphans = append(orphans, "SITE-B TOOL: "+relTo(sitesDir, p))
	}
	bGuides, _ := filepath.Glob(filepath.Join(sitesDir, "websitedesign.com/guides/*.html"))
	for _, p := range bGuides {
		orphans = append(orphans, "SITE-B GUIDE: "+relTo(sitesDir, p))
	}

	sort.Strings(orphans)

	// ---- Carry keywords from the old search index ---------------------------
	kw := loadOldSearchKeywords(filepath.Join(sitesDir, "website-design.com/search.json"))
	applyKeywords := func(entries []CatalogueEntry) {
		for i := range entries {
			if k, ok := kw[entries[i].Slug]; ok {
				entries[i].Keywords = k
			}
		}
	}
	applyKeywords(cat.Tools)
	applyKeywords(cat.Learn)

	if err := writeJSON(filepath.Join(portDir, "catalogue.harvested.json"), cat); err != nil {
		return err
	}
	fmt.Printf("harvested %d tools, %d learn pages, %d categories from the source indexes\n",
		len(cat.Tools), len(cat.Learn), len(cat.ToolCategories))

	// ---- merge the hand-written additions -----------------------------------
	// catalogue.json is generated, not hand-maintained: re-running harvest after
	// a source edit must reproduce it exactly. The hand-authored half lives in
	// catalogue_additions.json and is the only file anyone edits.
	adds, err := loadAdditions(filepath.Join(portDir, "catalogue_additions.json"))
	if err != nil {
		return fmt.Errorf("additions: %w", err)
	}
	before := len(cat.Tools) + len(cat.Learn) + len(cat.Pages)
	cat.Tools = append(cat.Tools, adds.Tools...)
	cat.Learn = append(cat.Learn, adds.Learn...)
	cat.Pages = append(cat.Pages, adds.Pages...)
	for k, v := range adds.LearnCategories {
		cat.LearnCategories[k] = v
	}

	// Overlays refine harvested entries in place (subtitles, QA tiers).
	applied := map[string]bool{}
	overlay := func(entries []CatalogueEntry) {
		for i := range entries {
			o, ok := adds.Overlays[entries[i].Slug]
			if !ok {
				continue
			}
			applied[entries[i].Slug] = true
			if o.Subtitle != "" {
				entries[i].Subtitle = o.Subtitle
			}
			if o.QATier != 0 {
				entries[i].QATier = o.QATier
			}
		}
	}
	overlay(cat.Tools)
	overlay(cat.Learn)
	overlay(cat.Pages)
	for slug := range adds.Overlays {
		if !applied[slug] {
			return fmt.Errorf("overlay for %q matched no catalogue entry — "+
				"the slug is wrong or the page was dropped", slug)
		}
	}

	// A duplicate slug within tools (or within learn) would collide on URL and
	// silently shadow a page. The whole point of D7 was that there are none.
	if err := assertUniqueSlugs("tools", cat.Tools); err != nil {
		return err
	}
	if err := assertUniqueSlugs("learn", cat.Learn); err != nil {
		return err
	}
	for _, e := range allEntries(cat) {
		if e.Label == "" || e.Category == "" || e.Slug == "" || e.Source == "" {
			return fmt.Errorf("incomplete catalogue entry: %+v", e)
		}
	}

	out := filepath.Join(portDir, "catalogue.json")
	if err := writeJSON(out, cat); err != nil {
		return err
	}
	fmt.Printf("merged %d hand-written additions and %d overlays -> %d pages total\n",
		len(allEntries(cat))-before, len(adds.Overlays), len(allEntries(cat)))
	fmt.Printf("wrote %s\n", out)

	// Anything on disk that is neither harvested nor added nor deliberately
	// dropped is a page we would silently lose. Report it every run.
	known := map[string]bool{}
	for _, e := range allEntries(cat) {
		known[e.Source] = true
	}
	var unaccounted []string
	for _, o := range orphans {
		src := o[strings.LastIndex(o, " ")+1:]
		if _, isDropped := adds.Dropped[src]; !known[src] && !isDropped {
			unaccounted = append(unaccounted, src)
		}
	}
	if len(unaccounted) > 0 {
		fmt.Printf("\n%d source page(s) are neither catalogued nor explicitly dropped:\n", len(unaccounted))
		for _, u := range unaccounted {
			fmt.Println("  UNACCOUNTED " + u)
		}
		return fmt.Errorf("%d unaccounted source page(s) — add to catalogue_additions.json "+
			"or list in its \"dropped\" map with a reason", len(unaccounted))
	}
	fmt.Println("\nevery source page is either catalogued or explicitly dropped.")
	return nil
}

// Additions is the hand-authored half of the catalogue.
type Additions struct {
	LearnCategories map[string]string  `json:"learn_categories"`
	Tools           []CatalogueEntry   `json:"tools"`
	Learn           []CatalogueEntry   `json:"learn"`
	Pages           []CatalogueEntry   `json:"pages"`
	Overlays        map[string]Overlay `json:"overlays"`
	// Dropped maps a source path to the reason it is not ported. Present so a
	// deliberate omission is visibly different from an accidental one.
	Dropped map[string]string `json:"dropped"`
}

// Overlay refines a harvested entry.
type Overlay struct {
	Subtitle string `json:"subtitle,omitempty"`
	QATier   int    `json:"qa_tier,omitempty"`
}

func loadAdditions(p string) (*Additions, error) {
	var a Additions
	if err := readJSON(p, &a); err != nil {
		return nil, err
	}
	if a.Overlays == nil {
		a.Overlays = map[string]Overlay{}
	}
	if a.Dropped == nil {
		a.Dropped = map[string]string{}
	}
	if a.LearnCategories == nil {
		a.LearnCategories = map[string]string{}
	}
	return &a, nil
}

func assertUniqueSlugs(kind string, entries []CatalogueEntry) error {
	seen := map[string]string{}
	for _, e := range entries {
		key := e.Slug
		if kind == "learn" {
			key = e.Category + "/" + e.Slug
		}
		if prev, dup := seen[key]; dup {
			return fmt.Errorf("%s: duplicate slug %q (%s and %s)", kind, key, prev, e.Source)
		}
		seen[key] = e.Source
	}
	return nil
}

// harvestToolCards reads the category headings and cards from a tools index.
func harvestToolCards(path string, cat *Catalogue) error {
	doc, err := parseFile(path)
	if err != nil {
		return err
	}
	// Walk in document order; a `.category-title` opens a new category, and every
	// `.card` after it belongs to that category until the next heading.
	current := ""
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if hasClass(n, "category-title") {
				current = textOf(n)
				cat.ToolCategories = append(cat.ToolCategories, current)
			} else if hasClass(n, "card") {
				if e, ok := cardEntry(n); ok {
					e.Category = current
					e.QATier = 2 // default; tier 1 set by hand for canvas/file tools
					cat.Tools = append(cat.Tools, e)
				}
				return // do not descend into a card
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return nil
}

// harvestLearnCards reads the learn index. It is shaped differently from the
// tools index in two ways that both had to be handled: its cards are
// `.article-card` with `.article-title`/`.article-desc`/`.article-tag` inside,
// and the link is the card's PARENT <a> rather than a child of it. Its sections
// are plain <h2> headings, so the category comes from the article's own path
// (/learn/<cat>/<slug>.html), which is the more reliable signal anyway.
func harvestLearnCards(path string, cat *Catalogue) error {
	doc, err := parseFile(path)
	if err != nil {
		return err
	}
	for _, card := range findAll(doc, func(n *html.Node) bool {
		return n.Type == html.ElementNode && hasClass(n, "article-card")
	}) {
		e, ok := cardEntry(card)
		if !ok {
			continue
		}
		// /learn/<category>/<slug>.html
		parts := strings.Split(strings.Trim(e.Source, "/"), "/")
		if len(parts) >= 3 {
			e.Category = parts[len(parts)-2]
		}
		if e.Category != "" {
			if _, seen := cat.LearnCategories[e.Category]; !seen {
				cat.LearnCategories[e.Category] = titleCase(e.Category)
			}
		}
		e.QATier = 3
		cat.Learn = append(cat.Learn, e)
	}
	return nil
}

// cardEntry pulls one card's label, description, eyebrow and target. It handles
// both card shapes the two indexes use: the tools index nests the link inside
// the card, the learn index wraps the card in the link.
func cardEntry(card *html.Node) (CatalogueEntry, bool) {
	var e CatalogueEntry
	href := ""
	if link := findFirst(card, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.A && attr(n, "href") != ""
	}); link != nil {
		href = attr(link, "href")
	}
	if href == "" {
		for p := card.Parent; p != nil; p = p.Parent {
			if p.Type == html.ElementNode && p.DataAtom == atom.A && attr(p, "href") != "" {
				href = attr(p, "href")
				break
			}
		}
	}
	if href == "" {
		return e, false
	}

	h := findFirst(card, func(n *html.Node) bool {
		return n.Type == html.ElementNode && (n.DataAtom == atom.H3 || n.DataAtom == atom.H2)
	})
	if h != nil {
		e.Label = textOf(h)
	}
	// The first <p> is the card blurb. The eyebrow is a small span above the
	// title — `.text-mono` on the tools index, `.article-tag` on the learn index.
	if p := findFirst(card, byAtom(atom.P)); p != nil {
		e.Description = textOf(p)
	}
	if sp := findFirst(card, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.DataAtom == atom.Span &&
			(hasClass(n, "text-mono") || hasClass(n, "article-tag"))
	}); sp != nil {
		e.Eyebrow = textOf(sp)
	}

	// Source path relative to the sites repo. Index hrefs are root-relative to
	// the OLD site, so prefix the site dir back on.
	e.Source = "website-design.com" + href
	e.Slug = slugFromURL(href)
	if e.Label == "" || e.Slug == "" {
		return e, false
	}
	return e, true
}

// slugFromURL: /tools/aspect-ratio/index.html -> aspect-ratio
//
//	/learn/design/oklch.html       -> oklch
func slugFromURL(u string) string {
	u = strings.TrimSuffix(u, "/")
	parts := strings.Split(strings.Trim(u, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if last == "index.html" {
		if len(parts) >= 2 {
			return parts[len(parts)-2]
		}
		return ""
	}
	return strings.TrimSuffix(last, ".html")
}

// loadOldSearchKeywords carries the curated keywords from the source search
// index, keyed by slug so a renamed URL still finds them.
func loadOldSearchKeywords(path string) map[string][]string {
	out := map[string][]string{}
	b, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	var rows []map[string]any
	if json.Unmarshal(b, &rows) != nil {
		return out
	}
	for _, r := range rows {
		u, _ := r["url"].(string)
		k, _ := r["keywords"].(string)
		slug := slugFromURL(u)
		if slug != "" && k != "" {
			out[slug] = strings.Fields(strings.ReplaceAll(k, ",", " "))
		}
	}
	return out
}

func titleCase(s string) string {
	s = strings.ReplaceAll(s, "-", " ")
	parts := strings.Fields(s)
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func relTo(base, p string) string {
	r, err := filepath.Rel(base, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(r)
}

func parseFile(path string) (*html.Node, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return html.Parse(f)
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
