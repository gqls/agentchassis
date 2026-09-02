// FILE: platform/orchestration/datahelpers/page_canonical.go
//
// CanonicalisePage maps any logical page descriptor — from adoption's
// analyze_site LLM, the site-planner's LLM, the reconciler, or any
// future role-validation pipeline — to the single canonical
// (name, url, page_type) triple that the pages table, the
// site_plan_pages table, and `needs_page:<name>` work-item keys all use.
//
// History:
//
//	Phase 0 (doc 029): introduced to converge adoption-shape names like
//	"tool-ttk-calculator" / "guides-index" with planner-shape names like
//	"ttk-calculator" / "guides". Without this convergence, both surfaces
//	upserted into pages with disagreeing names, the
//	ON CONFLICT (site_id, name) DO UPDATE never fired, and we got
//	duplicate rows plus parallel work-item streams that idx_swi_dedup
//	could not collapse.
//
//	Phase 1 (doc 030): extended with ParentSection so the same helper
//	covers nested-URL synthesis for sites whose section directories
//	don't match the role defaults (e.g. tools nested under "utilities"
//	instead of "tools"). Adoption's existing call sites pass empty
//	ParentSection and get identical Phase 0 behaviour.
//
//	Phase 1.5 + kebab (doc 051): canonical page_type output is now
//	kebab-case (matches the standards doc and the bulk of existing
//	pages.page_type data). Snake inputs (blog_post, blog_index, etc.)
//	are still accepted for backward compatibility with older planner
//	prompts, but the function always returns kebab. The homepage's
//	page_type is now "landing" rather than "index" (the page's name
//	is still "index" — the rename separates the storage-convention
//	name from the page-type description).
//
// The helper is idempotent: feeding it the adoption-shape produces the
// same output as feeding it the planner-shape. It does not query the
// database — naming and URL synthesis are decided purely from the
// descriptor.

package datahelpers

import (
	"strings"
)

// PageDescriptor is the input to CanonicalisePage.
//
// Role is the LLM's `page_type` field. The canonical vocabulary is
// kebab-case; underscore inputs are still accepted for backward
// compatibility:
//
//	"index" | "content" | "landing" |
//	"tool" | "guide" | "game" |
//	"section-index" | "blog-index" | "news-index" |
//	"entity-directory" |
//	"blog-post" |
//	"entity-page"
//
// Underscore forms (blog_post, blog_index, section_index, news_index,
// entity_directory, entity_page) are normalised to their kebab
// equivalents on input.
//
// The section-index family (`section-index`, `blog-index`,
// `news-index`, `entity-directory`) all share the same name and URL shape — they
// represent "a page that lists other pages of a category" — but the
// helper preserves the original role as `page_type` so downstream
// consumers (page-build-handler in particular) can dispatch differently
// for blog vs entity vs generic section.
//
// Unknown roles fall back to the "content" rule (name=slug,
// url=/<slug>.html), preserving the unknown role string as page_type
// so callers can see in logs that an unknown role slipped through.
//
// Slug is the page-name stem from the upstream surface. May be the
// adoption-shape ("tool-ttk-calculator", "guides-index") or the
// planner-shape ("ttk-calculator", "guides"). The helper strips the
// redundant role prefix when present so both shapes converge.
//
// Section is used by the section-index family when the slug doesn't
// already encode it. Optional — if empty, derived from Slug by
// stripping a trailing "-index". This is a Phase 0 field, kept for
// backward compatibility with adoption's call sites; see ParentSection
// for the Phase 1 nested-URL field.
//
// ParentSection (Phase 1, doc 030) is the canonical name of the section
// directory this page sits under. When non-empty, it overrides the
// role-default section directory in URL synthesis:
//
//	role=tool, slug=foo, ParentSection=""           → /tools/foo/index.html
//	role=tool, slug=foo, ParentSection="utilities"  → /utilities/foo/index.html
//	role=tool, slug=foo, ParentSection="tools-index"→ /tools/foo/index.html (suffix stripped)
//
// The nested roles (tool, guide, game, blog-post, entity-page) honour
// ParentSection. The section-index family ignores it because a section
// index IS its section. content/landing/index ignore it because the
// concept doesn't apply.
// FlatURLs (2026-08-10) selects the FLAT url shape for the nested roles
// (tool, guide, game): /<dir>/<slug>.html instead of
// /<dir>/<slug>/index.html. Name and page_type are unaffected — this
// changes the URL only.
//
// It is an OPT-IN FIELD with the new behaviour OFF by default, because
// the helper has twelve call sites and every one of them must keep its
// existing output byte-for-byte (owner ruling 2026-08-02: new authority
// on a shared seam ships as a field with the unsafe default off, not as
// a documented contract — a comment is not a control on a tree this
// many sessions share). A caller that does not set it cannot be
// affected by it.
//
// Why it exists: the nested shape is unrepresentable for a site that
// already serves flat tool/guide URLs, so running the planner over one
// silently rewrote every such URL — pages.url is overwritten
// unconditionally by upsertPage and the deployer takes the file path
// from it. Measured on loancalculator.co.uk 2026-08-10: 24 of 26 live
// URLs would have moved (11 tools + 13 guides). blog-post and
// entity-page already emit the flat shape, so this makes tool/guide/game
// expressible in a vocabulary the helper already had, rather than
// inventing a new one.
type PageDescriptor struct {
	Role          string
	Slug          string
	Section       string
	ParentSection string
	FlatURLs      bool
}

// CanonicalisePage returns the canonical (name, url, pageType) for the
// descriptor. Empty inputs return empty triples; the caller decides
// whether to skip or default.
func CanonicalisePage(d PageDescriptor) (name, url, pageType string) {
	role := normalisePageType(d.Role)
	slug := normaliseSlug(d.Slug)
	section := normaliseSlug(d.Section)
	parent := strings.TrimSuffix(normaliseSlug(d.ParentSection), "-index")

	// "index" role is treated as a writer convention for "the homepage".
	// We normalise to name="index", url="/index.html", page_type="landing"
	// — the page_type captures the semantic (this IS a landing page),
	// distinct from the name which is the convention for storing the
	// homepage. Slug is irrelevant here.
	if role == "index" {
		return "index", "/index.html", "landing"
	}

	// "home" / "index" slug under content/landing/empty role collapses
	// to the homepage convention. Without this, an LLM emitting
	// {role:"content", name:"home"} would produce a separate /home.html
	// page that doesn't match the actual site root.
	if (slug == "home" || slug == "index") && (role == "" || role == "content" || role == "landing") {
		return "index", "/index.html", "landing"
	}

	// Section-index family: section-index, blog-index, news-index,
	// entity-directory.
	// Same name and URL shape; page_type preserved as the input role so
	// downstream dispatch can distinguish flavours. Adoption's
	// (role=blog-index, slug=guides) and planner's
	// (role=section-index, slug=guides) both produce
	// (name=guides-index, url=/guides/index.html) — convergent.
	if isSectionIndexRole(role) {
		// Adoption shape: slug="guides-index" (with -index suffix).
		// Planner shape: slug="guides" (bare).
		// Optional explicit section overrides slug-derivation.
		// ParentSection is intentionally NOT honoured — a section
		// index page IS its section.
		sec := section
		if sec == "" {
			sec = strings.TrimSuffix(slug, "-index")
		}
		if sec == "" {
			return "", "", ""
		}
		return sec + "-index", "/" + sec + "/index.html", role
	}

	switch role {
	case "tool":
		bare := strings.TrimPrefix(slug, "tool-")
		if bare == "" {
			return "", "", ""
		}
		dir := parent
		if dir == "" {
			dir = "tools"
		}
		return "tool-" + bare, nestedOrFlatURL(dir, bare, d.FlatURLs), "tool"

	case "guide":
		bare := strings.TrimPrefix(slug, "guide-")
		if bare == "" {
			return "", "", ""
		}
		dir := parent
		if dir == "" {
			dir = "guides"
		}
		return "guide-" + bare, nestedOrFlatURL(dir, bare, d.FlatURLs), "guide"

	case "game":
		bare := strings.TrimPrefix(slug, "game-")
		if bare == "" {
			return "", "", ""
		}
		dir := parent
		if dir == "" {
			dir = "games"
		}
		return "game-" + bare, nestedOrFlatURL(dir, bare, d.FlatURLs), "game"

	case "blog-post":
		// Adoption convention: name=<slug>, url=/blog/<slug>.html.
		// ParentSection allows nesting under a custom directory
		// (e.g. /guides/<slug>.html for sites that put blog-shaped
		// content under a guides hierarchy).
		if slug == "" {
			return "", "", ""
		}
		dir := parent
		if dir == "" {
			dir = "blog"
		}
		return slug, "/" + dir + "/" + slug + ".html", "blog-post"

	case "entity-page":
		// Individual entity within an entity directory. Name is the
		// slug (no prefix); URL nests under the parent section if
		// given, otherwise under "entities" as a sensible default.
		// ParentSection here usually points at the corresponding
		// entity-directory (e.g. parent=clinics for an entity-page
		// representing one clinic).
		if slug == "" {
			return "", "", ""
		}
		dir := parent
		if dir == "" {
			dir = "entities"
		}
		return slug, "/" + dir + "/" + slug + ".html", "entity-page"

	case "landing":
		// Conversion-focused flat page. Same URL shape as content;
		// page_type retained so downstream can render landing-specific
		// CTA logic. (Homepage-shaped landings — slug=home/index —
		// are caught by the earlier collapse rule before reaching here.)
		if slug == "" {
			return "", "", ""
		}
		return slug, "/" + slug + ".html", "landing"

	default:
		// "content" and anything unrecognised: name=slug, url=/<slug>.html.
		// pageType passes through the original role string when set, so
		// unknown roles surface in logs rather than being silently
		// rewritten to "content". ParentSection is intentionally NOT
		// honoured for content — the LLM should give the page a more
		// specific role if it belongs under a section.
		if slug == "" {
			return "", "", ""
		}
		pt := role
		if pt == "" {
			pt = "content"
		}
		return slug, "/" + slug + ".html", pt
	}
}

// nestedOrFlatURL synthesises the URL for the nested-directory roles
// (tool, guide, game). Nested (/<dir>/<slug>/index.html) is the
// long-standing default; flat (/<dir>/<slug>.html) is opt-in via
// PageDescriptor.FlatURLs — see that field's comment for why the
// default must stay nested.
func nestedOrFlatURL(dir, bare string, flat bool) string {
	if flat {
		return "/" + dir + "/" + bare + ".html"
	}
	return "/" + dir + "/" + bare + "/index.html"
}

// normalisePageType lowercases, trims whitespace, and converts known
// underscore-form role variants to their canonical kebab form. This is
// distinct from the validator's normaliseRole (in
// page_role_validator.go), which additionally collapses the
// section-index family into a single canonical "section-index" form
// for the validator's routing logic.
//
// Here we deliberately do NOT collapse the section-index family —
// CanonicalisePage retains them as distinct page_type values
// (blog-index, entity-directory, section-index) so downstream
// consumers can dispatch differently for each.
//
// The underscore-to-kebab conversion is a backward-compatibility
// affordance for older planner prompts that emit snake_case. New code
// should emit kebab directly; the safety net should not be relied
// upon (see standards doc, "Lookup safety net" note).
func normalisePageType(r string) string {
	r = strings.ToLower(strings.TrimSpace(r))
	switch r {
	case "blog_post":
		return "blog-post"
	case "blog_index":
		return "blog-index"
	case "section_index":
		return "section-index"
	case "news_index":
		return "news-index"
	case "entity_directory":
		return "entity-directory"
	case "entity_page":
		return "entity-page"
	default:
		return r
	}
}

// IsSectionIndexRole is the exported face of isSectionIndexRole, so the
// listing-family vocabulary has ONE definition: the bugs_open/444 plan-time
// gate classifies by this instead of carrying a mirror (council corr c0990eb3
// round 2, architecture seat — a third hand-maintained copy of this set was
// the objection).
func IsSectionIndexRole(role string) bool { return isSectionIndexRole(role) }

// isSectionIndexRole returns true for any role that represents "a page
// listing other pages of a category". These all share name/URL shape;
// only page_type differs. Input is expected to already be canonicalised
// by normalisePageType (kebab form).
func isSectionIndexRole(role string) bool {
	switch role {
	case "section-index", "blog-index", "entity-directory", "news-index":
		return true
	}
	return false
}

// normaliseSlug applies the cleanup that adoption used to do inline
// before this helper existed (apply_adoption_plan_action.go pre-029) so
// callers can pass raw LLM output. Also tolerates a leaked URL fragment
// where someone passes "/tools/ttk-calculator.html" as the slug.
func normaliseSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.TrimPrefix(s, "/")
	if i := strings.LastIndex(s, "/"); i >= 0 {
		s = s[i+1:]
	}
	s = strings.TrimSuffix(s, ".html")
	s = strings.TrimSuffix(s, ".htm")
	return s
}
