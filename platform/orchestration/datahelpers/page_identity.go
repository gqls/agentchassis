package datahelpers

import "strings"

// Page IDENTITY KEYS — the deterministic ways two page spellings can be
// recognised as denoting one page.
//
// Distinct from CanonicalisePage, which answers "what SHOULD this page be
// called": these answer "could these two names/URLs be the same page". They
// live beside the canonicaliser because every one of them is derived from a
// rule CanonicalisePage itself applies, and a copy that drifts from it is
// silently wrong rather than loudly broken.
//
// WHY THEY ARE HERE (bugs_open/215, the quiet mode). Each of these keys
// already existed in the estate, hand-copied into the package that needed it:
// pageURLPathKey in discovery_checks/check_page_canonical_collision.go and
// itemStemOf in actions/v3_site_actions.go, the latter carrying the comment
// "keep this prefix list in sync with them" — a sync obligation discharged by
// hand, across two packages, against a canonicaliser in a third. The 215 fix
// needs all three in a THIRD place (the plan/realised reconciler), which is
// the trigger the imagery lane's round-3 reuse note named for exactly this
// pattern: the third instance builds the shared utility rather than another
// copy (register IMG-070). The original call sites now delegate here, so the
// prefix list and the suffix rules have one definition.
//
// All three are pure and DB-free, like CanonicalisePage. They are keys, not
// decisions: two rows sharing a key are CANDIDATES to be one page. What the
// caller may then do about it is the caller's own judgement, and in the
// reconciler's case is deliberately gated (see reconcilePlanWithRealised).

// PagePathKey maps a stored url to the path it claims: /news/index.html and
// /news.html both claim /news. "" (from /index.html) is the homepage, not
// no-signal — it normalises to "/" so a `home` page beside `index` still
// collides. A url that ends in neither suffix (a fragment url, a directory)
// passes through unchanged, so it can only collide with an exact twin.
//
// This is the ONLY key that sees the flat-vs-nested divergence, which is why
// the reconciler runs it: the legacy tool-deploy arm wrote /tools/<bare>.html
// while the canonicaliser writes /tools/<bare>/index.html, so a plan entry and
// the live page it denotes can differ in BOTH name and url at once and match
// on neither exactly (bugs_open/215; the shape measured on fundamentallyai
// 2026-08-11).
func PagePathKey(url string) string {
	k := url
	if strings.HasSuffix(k, "/index.html") {
		k = strings.TrimSuffix(k, "/index.html")
	} else {
		k = strings.TrimSuffix(k, ".html")
	}
	if k == "" {
		k = "/"
	}
	return k
}

// PageItemStem returns the topic stem of an item page name by stripping the
// role prefixes CanonicalisePage adds (tool-, guide-, game-): e.g.
// "guide-economy-basics" -> "economy-basics", "economy-basics" ->
// "economy-basics". Returns the input lowercased/trimmed when no role prefix
// is present, so a bare page and its prefixed twin share a stem.
//
// The prefix list mirrors the TrimPrefix calls in CanonicalisePage's
// tool/guide/game cases; keeping them in one file is the whole point of this
// file, and TestPageItemStem_InvertsCanonicalisePagePrefixes fails if they
// diverge.
//
// This is the WEAKEST of the three keys and the only one that can pair two
// genuinely different pages: a new "tool-pricing" beside a built
// "guide-pricing" shares the stem "pricing" while denoting different pages.
// That is a real risk recorded on the reconciler's Pass C2 since 2026-07-20,
// and it is why the stem key is never sufficient on its own — see the
// bare-vs-prefixed guard at its reconciler call site, which makes precisely
// that pair unmatchable.
func PageItemStem(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, p := range []string{"tool-", "guide-", "game-"} {
		if strings.HasPrefix(n, p) {
			return strings.TrimPrefix(n, p)
		}
	}
	return n
}

// PageCanonicalNameForRow is the name signal for a STORED row: what the shared
// canonicaliser says this row's name ought to be. Empty means the row
// contributes no name signal — an uncanonicalisable identity is skipped rather
// than grouped, else every empty triple would "collide" with every other.
//
// A legacy page_type of "index" is skipped for the same reason: CanonicalisePage
// collapses that role to the homepage regardless of slug, so grouping on it
// would false-collide any such row with the real homepage. That skip is load
// bearing, not defensive — it is the homepage-collapse family, the one collapse
// CanonicalisePage performs that is a deliberate MERGE of two writer conventions
// rather than a spelling correction, and so the one that must never be read as
// evidence two rows are the same page.
func PageCanonicalNameForRow(name, pageType string) string {
	if pageType == "index" {
		return ""
	}
	canonName, _, _ := CanonicalisePage(PageDescriptor{
		Role: pageType,
		Slug: name,
	})
	return canonName
}

// ParentSectionFromURL recovers the directory a page lives in, so a page
// canonicalises back to where it is SERVING (or, for a proposed page, to where
// it was PLANNED) rather than to its role's default hub. "/guides/x.html" and
// "/guides/x/index.html" both give "guides"; a root-level "/x.html" gives ""
// (no parent), which is also what a URL we cannot read gives — in both cases
// CanonicalisePage's own default applies, which is the behaviour that existed
// before any caller carried this.
//
// One definition, because two callers now need it and they must not disagree:
// the reconciler carries it off a REALISED row (normaliseRealisedToPlanPage,
// against the bugs_open/241 URL-move hazard), and ValidateRoles derives it for
// a PROPOSED page whose section is expressed only in the URL the write path is
// about to discard (bugs_open/463).
func ParentSectionFromURL(url string) string {
	trimmed := strings.Trim(url, "/")
	i := strings.Index(trimmed, "/")
	if i <= 0 {
		return ""
	}
	return trimmed[:i]
}

// PlanPageCarriesRealisedIdentity reports whether a plan entry was derived from,
// or paired with, a realised page by reconcilePlanWithRealised — i.e. whether
// its identity is the stored one rather than the planner's proposal.
//
// Read by both write surfaces so neither invents an identity the reconciler has
// already settled. The two markers are not redundant: a snapped or unioned entry
// carries BOTH (normaliseRealisedToPlanPage), while a same-name stamp carries
// only identity_authority — it keeps the planner's title, meta and nav, so it is
// not wholly realised-derived. Either marker is enough to mean "the reconciler
// decided this entry's identity; do not re-derive parts of it here".
func PlanPageCarriesRealisedIdentity(p map[string]interface{}) bool {
	if s, _ := p["identity_authority"].(string); s != "" {
		return true
	}
	b, _ := p["from_realised"].(bool)
	return b
}
