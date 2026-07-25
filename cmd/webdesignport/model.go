package main

// Shared types for the webdesign.co.uk port: the catalogue (what we are porting
// and how it is labelled), the overrides (per-page exceptions), and the manifest
// (what the transform produced, and the contract the importer consumes).

// Catalogue is the curated inventory. It is bootstrapped by `harvest` from the
// two source sites' own index pages — which already carry every tool's label,
// one-line description and category — and then hand-extended for the pages the
// source indexes never linked.
type Catalogue struct {
	// ToolCategories in display order. Every tool's Category must be one of these.
	ToolCategories []string `json:"tool_categories"`
	// LearnCategories maps a learn category slug to its display name.
	LearnCategories map[string]string `json:"learn_categories"`
	Tools           []CatalogueEntry  `json:"tools"`
	Learn           []CatalogueEntry  `json:"learn"`
}

// CatalogueEntry is one ported page as the indexes and search will present it.
type CatalogueEntry struct {
	// Source is the path relative to the sites repo root,
	// e.g. "website-design.com/tools/aspect-ratio/index.html".
	Source string `json:"source"`
	// Slug is the tool dir name or the article file stem.
	Slug string `json:"slug"`
	// Label is the human name shown on cards ("Aspect Ratio Calculator").
	Label string `json:"label"`
	// Description is the card's one-liner, harvested from the source index.
	Description string `json:"description"`
	// Subtitle disambiguates near-duplicate tools in the index. Empty for most.
	// See PLAN D7 — five pairs across the two sites do similar-sounding jobs.
	Subtitle string `json:"subtitle,omitempty"`
	// Category is a tool category name, or a learn category slug.
	Category string `json:"category"`
	// Eyebrow is the small label above a card title, e.g. "AI UTILITY".
	Eyebrow string `json:"eyebrow,omitempty"`
	// Keywords feed search.json. Harvested from the old search index where present.
	Keywords []string `json:"keywords,omitempty"`
	// QATier: 1 = hands-on browser session, 2 = click-through, 3 = scroll check.
	QATier int `json:"qa_tier"`
}

// Overrides carries every per-page exception the transform needs. Data, not code,
// so a QA fix is an edit here plus a re-run rather than a code change.
type Overrides struct {
	// Drop lists source paths (relative to the sites repo) never to port.
	Drop []string `json:"drop"`
	// BrandSuffix replaces the old "| website-design.com" title tails.
	BrandSuffix string `json:"brand_suffix"`
	// StripSelectors are removed from every page: the sources' inline chrome.
	// Matched by the tiny selector engine in dom.go (tag, .class, #id only).
	StripSelectors []string `json:"strip_selectors"`
	// Pages keyed by source path.
	Pages map[string]PageOverride `json:"pages"`
	// LinkMap rewrites specific hrefs site-wide, applied after the generic rules.
	LinkMap map[string]string `json:"link_map"`
	// ForbiddenRefs fail the build if any survives in a fragment. These are the
	// pages we dropped: a live link to one would be a 404 on the new site.
	ForbiddenRefs []string `json:"forbidden_refs"`
}

// PageOverride is one page's exceptions.
type PageOverride struct {
	// KeepSelectors are rescued from inside a stripped region before it is
	// removed — e.g. vibe-equalizer's functional Share button lives in the
	// nav bar that everything else about is chrome.
	KeepSelectors []string `json:"keep_selectors,omitempty"`
	// ExtraStrip is stripped in addition to the global list.
	ExtraStrip []string `json:"extra_strip,omitempty"`
	// Rewrites are exact-string replacements applied to the fragment HTML.
	// Each MUST match at least once or the build fails: they encode assumptions
	// about source prose (e.g. rewording a mention of the dropped LLM builder),
	// and a silent no-op would ship the old wording.
	Rewrites []Rewrite `json:"rewrites,omitempty"`
	// RuntimeFill marks a page whose fragment is legitimately almost all widget
	// and would otherwise trip the assembly's visible-text floor.
	RuntimeFill bool `json:"runtime_fill,omitempty"`
	// Title and MetaDescription override what was extracted from the source head.
	Title           string `json:"title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	// SkipColourSweep disables the literal-colour rewrite for this page. For
	// tools whose subject IS colour, where a literal in CSS may be sample data.
	SkipColourSweep bool `json:"skip_colour_sweep,omitempty"`
}

// Rewrite is one exact-string replacement.
type Rewrite struct {
	From string `json:"from"`
	To   string `json:"to"`
	Why  string `json:"why,omitempty"`
}

// Manifest is the transform's output and the importer's input.
type Manifest struct {
	Domain    string          `json:"domain"`
	Generator string          `json:"generator"`
	Pages     []ManifestPage  `json:"pages"`
	Assets    []ManifestAsset `json:"assets"`
	Warnings  []string        `json:"warnings"`
}

// ManifestPage is one ported page.
type ManifestPage struct {
	// Name is the pages.name key, e.g. "tool-aspect-ratio".
	Name string `json:"name"`
	// URL is the pages.url — it MUST carry an explicit filename, because
	// getPageInfo derives the published filename as TrimPrefix(url,"/").
	URL             string   `json:"url"`
	PageType        string   `json:"page_type"`
	Title           string   `json:"title"`
	MetaDescription string   `json:"meta_description"`
	Category        string   `json:"category"`
	Label           string   `json:"label"`
	Subtitle        string   `json:"subtitle,omitempty"`
	Description     string   `json:"description,omitempty"`
	Eyebrow         string   `json:"eyebrow,omitempty"`
	Keywords        []string `json:"keywords,omitempty"`
	QATier          int      `json:"qa_tier"`
	// Source is the originating file, for the structural diff in `verify`.
	Source string `json:"source"`
	// Fragment is the emitted file, relative to the output dir.
	Fragment string `json:"fragment"`
	// SHA256 of the fragment — the importer's idempotency key.
	SHA256 string `json:"sha256"`
	// VisibleTextChars reimplements the assembly's sectionHasVisibleContent
	// metric. At or below 10, assembly silently drops the section.
	VisibleTextChars int `json:"visible_text_chars"`
	// Links are the internal hrefs found in the fragment, for closure checking.
	Links []string `json:"links,omitempty"`
	// SiblingAssets are files the page loads relatively (its own engine JS).
	SiblingAssets []string `json:"sibling_assets,omitempty"`
	RuntimeFill   bool     `json:"runtime_fill,omitempty"`
}

// ManifestAsset is a static file to publish into the deploy repo.
type ManifestAsset struct {
	// Source path relative to the sites repo (empty when generated).
	Source string `json:"source,omitempty"`
	// Dest path relative to the site root, e.g. "assets/img/hero.jpg".
	Dest      string `json:"dest"`
	Generated bool   `json:"generated,omitempty"`
	SHA256    string `json:"sha256,omitempty"`
}

// SearchEntry is one row of the regenerated search.json. Field names match what
// the carried search.js filters on — do not rename without changing the engine.
type SearchEntry struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Category string `json:"category"`
	Keywords string `json:"keywords"`
}
