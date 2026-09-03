// FILE: platform/orchestration/datahelpers/page_role_validator.go
//
// ValidateRoles corrects role mislabels in a planned page set before the
// pages are persisted to site_plan_pages. The motivating case is the
// 2026-05-04 gamesdesign.co.uk run, where the plan-builder LLM emitted:
//
//   {name: "tools",  role: "content",    url: "/tools/index.html"}
//   {name: "guides", role: "blog-index", url: "/guides/index.html"}
//   {name: "games",  role: "content",    url: "/games/index.html"}
//
// Every one of these is structurally a section-index (a directory landing
// page that lists the sub-pages beneath it), but the LLM picked
// "content" / "blog-index" / "content" respectively. Without correction,
// these write to pages with role=content and url=/<slug>.html, breaking
// every internal link the source site had to /<slug>/index.html.
//
// Doc 030 Q3: role assignment is a deterministic concern, not an LLM one.
// The plan-builder LLM is allowed to be sloppy about role; this validator
// catches and corrects before anything is persisted.
//
// Doc 051: canonical role output is now kebab-case (matches the standards
// doc and the data canonicaliser). Snake-form inputs are still accepted
// for backward compatibility with older planner prompts.
//
// The validator operates on the full set of pages because correction
// requires cross-page signals (e.g. "is any other page parented under
// this slug?"). It's pure Go, no DB, idempotent — running it twice on
// the same input produces the same output.

package datahelpers

import (
	"strings"
)

// LLMPlannedPage is the raw shape coming from the plan-builder LLM,
// before validation. Field names match what the LLM is prompted to
// emit; tolerance for variation is in the consumer (write_site_plan)
// not here.
type LLMPlannedPage struct {
	Name          string // raw, may be slug-shape ("tools") or name-shape ("tools-index")
	Role          string // raw, may be wrong
	Slug          string // bare stem; if empty, derived from Name
	URL           string // LLM-supplied URL, used as a corroborating signal
	ParentSection string // canonical parent name, or empty

	// RealisedIdentity marks an entry whose identity reconcilePlanWithRealised
	// has already settled against a realised page (see
	// datahelpers.PlanPageCarriesRealisedIdentity). The validator derives
	// nothing for such an entry — see the ParentSection derivation below.
	RealisedIdentity bool
}

// ValidatedPage is the corrected shape after running the validator. Role
// has been adjusted where structural signals contradicted the LLM. Slug
// is normalised. Other fields pass through unchanged.
type ValidatedPage struct {
	Name          string
	Role          string
	Slug          string
	URL           string // unchanged; URL synthesis happens later via BuildPageURL
	ParentSection string

	// CorrectedFromRole is set when the validator changed the role.
	// Empty means the LLM's role was accepted as-is. Caller should log
	// non-empty values so we can monitor validator activity in production.
	CorrectedFromRole string
}

// ValidateRoles applies role corrections to a set of planned pages.
// Returns a new slice with corrections applied; does not mutate the
// input. Order is preserved.
//
// Correction rules, applied in order (first match wins):
//
//  1. Explicit role "index" or page name "index" → role="landing", regardless
//     of LLM input. "index" is the page name convention for the homepage;
//     "landing" is the canonical page type for it.
//
//     1b. Explicit flavoured index role (isTypedIndexRole: "news-index") →
//     accepted as-is. These are section hubs, so rules 2-4 would all fire
//     and flatten the flavour to "section-index" — which orphans the page
//     from every gate keying on the flavour (bugs_open/015).
//
//  2. Page NAME ends in "-index" AND the LLM role is not an explicit leaf
//     role → role="section-index". This is the NAME signal, and it is the
//     reliable one: the plan-builder names section hubs "<section>-index"
//     even when it omits the URL and parent_section that rules 3-5 depend
//     on. On the 2026-05-23 gamesdesign.co.uk run, games-index/tools-index
//     arrived as role="content" with NO url and NO declared children,
//     starving rules 3-5; only the name carried the truth. Guarded by
//     isLeafRole so an explicit tool/guide/game/blog-post/entity-page is
//     not overridden by an unusual "-index" name.
//
//  3. Slug appears as ParentSection of any other page in the set →
//     role="section-index". This is the structural signal: a page that
//     other pages claim as their parent IS a section index, no matter
//     what the LLM said. This catches the gamesdesign case directly.
//
//  4. LLM-supplied URL ends in "/index.html" AND slug appears as the
//     directory immediately above (e.g. url="/tools/index.html" and
//     slug="tools") → role="section-index". This is a fallback for when
//     the LLM correctly emits the URL pattern but mislabels the role.
//
//  5. LLM-supplied URL fits a nested role pattern AND role disagrees:
//     url="/tools/<x>/index.html" with role!="tool" → role="tool".
//     url="/guides/<x>/index.html" with role!="guide" → role="guide".
//     url="/games/<x>/index.html" with role!="game" → role="game".
//     This catches the symmetric case where the LLM gets the URL right
//     but the role wrong for nested pages.
//
//  6. Otherwise: accept LLM's role as-is, normalising underscore variants
//     ("blog_post", "blog_index", "section_index", etc.) to their kebab
//     canonical forms.
//
// Rule precedence matters. Rule 2 (name suffix) and rule 3 (declared
// parent) are both strong, reliable signals and agree wherever both fire
// (both → section-index), so their relative order is immaterial; rule 2 is
// placed first only because it depends on no other page. Both are
// structurally stronger than rules 4-5, which are URL-pattern heuristics.
func ValidateRoles(pages []LLMPlannedPage) []ValidatedPage {
	out := make([]ValidatedPage, len(pages))

	// Build the parent-section index first. Maps a slug/name to true
	// if any other page in the set claims it as ParentSection.
	declaredParents := make(map[string]bool, len(pages))
	for _, p := range pages {
		ps := normaliseSlug(p.ParentSection)
		if ps != "" {
			declaredParents[ps] = true
			// Also index by the canonical "<slug>-index" form so
			// pages using either shape match.
			declaredParents[ps+"-index"] = true
		}
	}

	for i, p := range pages {
		v := ValidatedPage{
			URL:           p.URL,
			ParentSection: p.ParentSection,
		}

		// Derive slug from name if not given.
		slug := normaliseSlug(p.Slug)
		if slug == "" {
			slug = normaliseSlug(p.Name)
			// Strip role prefix if the name carried one.
			slug = strings.TrimPrefix(slug, "tool-")
			slug = strings.TrimPrefix(slug, "guide-")
			slug = strings.TrimPrefix(slug, "game-")
			slug = strings.TrimSuffix(slug, "-index")
		}
		v.Slug = slug
		v.Name = normaliseSlug(p.Name)

		rawRole := normaliseRole(p.Role)
		correctedRole := rawRole

		// Rule 1: index identity. The page name "index" is the homepage
		// storage convention; its canonical page type is "landing".
		if rawRole == "index" || v.Name == "index" || slug == "index" {
			correctedRole = "landing"
		} else if isTypedIndexRole(rawRole) {
			// Rule 1b: an explicitly flavoured index role is trusted as-is.
			// These ARE section hubs, so every section-index correction
			// below (rules 2-4: "-index" name, declared parent, index URL)
			// would fire on them — and each would flatten the flavour to
			// generic "section-index". The flavour is a routing key, not a
			// label: render_news_section, MissingNewsPageCheck and
			// page-build-handler all select on page_type='news-index', so
			// flattening it orphans the page from all of them at once
			// (bugs_open/015, relojistas.com /noticias).
			correctedRole = rawRole
		} else if strings.HasSuffix(v.Name, "-index") && !isLeafRole(rawRole) {
			// Rule 2: name signal — a page NAMED "<section>-index" is a
			// section hub regardless of URL/parent. This is the reliable
			// signal the plan-builder emits even when it omits everything
			// rules 3-5 read (gamesdesign.co.uk 2026-05-23). Guarded so an
			// explicit leaf role is not clobbered by an unusual name.
			correctedRole = "section-index"
		} else if declaredParents[slug] || declaredParents[slug+"-index"] {
			// Rule 3: structural — this slug is someone's parent.
			correctedRole = "section-index"
		} else if isSectionIndexURL(p.URL, slug) {
			// Rule 4: URL pattern says section-index.
			correctedRole = "section-index"
		} else if r, ok := nestedRoleFromURL(p.URL); ok && r != rawRole {
			// Rule 5: URL pattern says nested role.
			correctedRole = r
		}

		v.Role = correctedRole
		if correctedRole != rawRole && rawRole != "" {
			v.CorrectedFromRole = rawRole
		}

		// Rule 6: the DIRECTORY a leaf page will canonicalise into, recovered
		// from the URL when the planner expressed its section only there
		// (bugs_open/463).
		//
		// WHY THIS IS NEEDED AT ALL. Both write surfaces discard the proposed
		// URL and re-derive it from CanonicalisePage, which reads ParentSection
		// for exactly the five leaf roles and otherwise defaults to the role's
		// own hub — blog-post to /blog/, tool to /tools/, and so on. So a page
		// the planner placed under /articles/ is written to /blog/ instead, and
		// the section index it was planned as a child of still resolves zero
		// children (listing_item_sources.go countSectionChildren, which matches
		// on the URL prefix). The planner cannot avoid this: the live plan_site
		// prompt has no parent_section field at all, and [MEASURED 2026-09-03]
		// 109 of 109 blog-post rows in site_plan_pages carry none.
		//
		// This is the same carry normaliseRealisedToPlanPage already performs in
		// the REALISED direction, for the same stated reason, so it adds no
		// mechanism — it removes an asymmetry. Rule 5 above is the accidental
		// half of it: nestedRoleFromURL rescues a nested page under
		// /tools//guides//games/ by RETYPING it, which is why those three
		// directories work today and /articles/, /news/, /case-studies/,
		// /resources/ and /insights/ do not.
		//
		// THREE GATES, EACH LOAD-BEARING.
		//   isLeafRole      — exactly the roles CanonicalisePage reads parent
		//                     for. Every other role ignores it, so the gate
		//                     costs nothing and bounds the change to where it
		//                     can act at all.
		//   !RealisedIdentity — an entry the reconciler paired with a realised
		//                     page is left alone. A snapped or unioned entry
		//                     already carries parent_section; a SAME-NAME
		//                     STAMPED entry carries the realised URL but
		//                     deliberately no parent_section, and deriving from
		//                     that URL would honour the realised identity with
		//                     the site's honour_realised_identity flag OFF —
		//                     precisely the behaviour change
		//                     stampSameNameRealisedIdentity refuses to make
		//                     (bugs_open/215). This gate is what keeps flag-off
		//                     behaviour for those entries byte-identical.
		//   ParentSection == "" — an explicit value always wins.
		//
		// Deliberately AFTER the role ladder and deliberately NOT fed into
		// declaredParents (built before the loop from the INPUT field only), so
		// no role decision can change: rule 3 sees exactly what it saw before.
		if v.ParentSection == "" && !p.RealisedIdentity && isLeafRole(v.Role) {
			v.ParentSection = ParentSectionFromURL(p.URL)
		}

		out[i] = v
	}

	return out
}

// isLeafRole reports whether a normalised role denotes a leaf/detail page
// type — one that is never itself a section index. The name-suffix rule
// (rule 2) consults this so that an explicit tool/guide/game/blog-post/
// entity-page is trusted even if the page name unusually ends in "-index"
// (e.g. a genuine blog post titled "weekly-index"). Roles that are absent,
// generic ("content"), or already section-index-family fall through and
// are eligible for the name-suffix correction.
func isLeafRole(role string) bool {
	switch role {
	case "tool", "guide", "game", "blog-post", "entity-page":
		return true
	}
	return false
}

// isTypedIndexRole reports whether a normalised role is an index-family
// role that carries a routing FLAVOUR downstream gates key on. Rule 1b
// consults this so the section-index corrections (rules 2-4) cannot
// flatten an explicit flavoured index to generic "section-index".
//
// Only "news-index" today. blog-index cannot join it here because
// normaliseRole deliberately collapses it before the rules run, and
// entity-directory has the same exposure but no observed incident —
// widen this only with evidence (bugs_open/015 records the observation).
func isTypedIndexRole(role string) bool {
	return role == "news-index"
}

// normaliseRole converts LLM role variants to canonical kebab form and
// lowercases everything. Accepts both kebab and snake inputs for
// backward compatibility with older planner prompts; output is always
// kebab.
//
// Unlike CanonicalisePage's normalisePageType, this function additionally
// collapses the section-index family (blog-index, section-index, and
// their snake variants) into a single "section-index" form because
// the validator only needs the routing distinction, not the flavour.
func normaliseRole(r string) string {
	r = strings.ToLower(strings.TrimSpace(r))
	switch r {
	case "blog-post", "blog_post":
		return "blog-post"
	case "blog-index", "blog_index", "section-index", "section_index":
		return "section-index"
	case "news-index", "news_index":
		// NOT collapsed into section-index: the news flavour is a routing
		// key several gates select on (bugs_open/015). See isTypedIndexRole.
		return "news-index"
	case "entity-directory", "entity_directory":
		return "entity-directory"
	case "entity-page", "entity_page":
		return "entity-page"
	default:
		return r
	}
}

// isSectionIndexURL returns true when the URL has the shape
// /<slug>/index.html or /<slug>/ (trailing slash) — the canonical
// section-index URL pattern.
func isSectionIndexURL(rawURL, slug string) bool {
	if rawURL == "" || slug == "" {
		return false
	}
	url := strings.ToLower(strings.TrimSpace(rawURL))
	if url == "/"+slug+"/index.html" {
		return true
	}
	if url == "/"+slug+"/" {
		return true
	}
	if url == "/"+slug+"/index" {
		return true
	}
	return false
}

// nestedRoleFromURL infers the role from a URL of the shape
// /<dir>/<x>/index.html where <dir> is one of the recognised section
// directory names. Returns the role and true on match.
//
// Only well-known directory names trigger inference — we don't want to
// guess a role from an arbitrary directory name like /resources/<x>/
// because the LLM may legitimately have intended "content" with a
// custom path.
func nestedRoleFromURL(rawURL string) (string, bool) {
	url := strings.ToLower(strings.TrimSpace(rawURL))
	if url == "" {
		return "", false
	}
	// Strip trailing /index.html if present.
	url = strings.TrimSuffix(url, "/index.html")
	url = strings.TrimSuffix(url, "/")

	parts := strings.Split(strings.TrimPrefix(url, "/"), "/")
	if len(parts) < 2 {
		return "", false
	}

	switch parts[0] {
	case "tools":
		return "tool", true
	case "guides":
		return "guide", true
	case "games":
		return "game", true
	}
	return "", false
}
