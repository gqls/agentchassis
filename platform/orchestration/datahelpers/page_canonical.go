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
// Role is normally the LLM's `page_type` field. Recognised values:
//
//	"tool" | "guide" | "game" | "section_index" |
//	"blog_post" | "content" | "index"
//
// Unknown roles fall back to the "content" rule (name=slug,
// url=/<slug>.html), preserving the unknown role string as page_type so
// callers can see in logs that an unknown role slipped through.
//
// Slug is the page-name stem from the upstream surface. May be the
// adoption-shape ("tool-ttk-calculator", "guides-index") or the
// planner-shape ("ttk-calculator", "guides"). The helper strips the
// redundant role prefix when present so both shapes converge.
//
// Section is used by section_index when the slug doesn't already encode
// it. Optional — if empty, derived from Slug by stripping a trailing
// "-index". This is a Phase 0 field, kept for backward compatibility
// with adoption's call sites; see ParentSection for the Phase 1 nested-
// URL field.
//
// ParentSection (Phase 1, doc 030) is the canonical name of the section
// directory this page sits under. When non-empty, it overrides the
// role-default section directory in URL synthesis:
//
//	role=tool, slug=foo, ParentSection=""           → /tools/foo/index.html
//	role=tool, slug=foo, ParentSection="utilities"  → /utilities/foo/index.html
//	role=tool, slug=foo, ParentSection="tools-index"→ /tools/foo/index.html (suffix stripped)
//
// Only the nested roles (tool, guide, game, blog_post) honour
// ParentSection. section_index, content, and index ignore it because
// the concept doesn't apply to those page kinds. Adoption code that
// pre-dates Phase 1 leaves ParentSection empty and gets identical
// behaviour to before — the field is purely additive.
type PageDescriptor struct {
	Role          string
	Slug          string
	Section       string
	ParentSection string
}

// CanonicalisePage returns the canonical (name, url, pageType) for the
// descriptor. Empty inputs return empty triples; the caller decides
// whether to skip or default.
func CanonicalisePage(d PageDescriptor) (name, url, pageType string) {
	role := strings.ToLower(strings.TrimSpace(d.Role))
	slug := normaliseSlug(d.Slug)
	section := normaliseSlug(d.Section)
	parent := strings.TrimSuffix(normaliseSlug(d.ParentSection), "-index")

	// "index" role is unconditional — Slug is irrelevant.
	if role == "index" {
		return "index", "/index.html", "index"
	}

	// "home" / "index" slug under content/empty role collapses to index.
	// Without this, an LLM emitting {role:"content", name:"home"} would
	// produce a separate /home.html page that doesn't match the actual
	// site root.
	if (slug == "home" || slug == "index") && (role == "" || role == "content") {
		return "index", "/index.html", "index"
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
		return "tool-" + bare, "/" + dir + "/" + bare + "/index.html", "tool"

	case "guide":
		bare := strings.TrimPrefix(slug, "guide-")
		if bare == "" {
			return "", "", ""
		}
		dir := parent
		if dir == "" {
			dir = "guides"
		}
		return "guide-" + bare, "/" + dir + "/" + bare + "/index.html", "guide"

	case "game":
		bare := strings.TrimPrefix(slug, "game-")
		if bare == "" {
			return "", "", ""
		}
		dir := parent
		if dir == "" {
			dir = "games"
		}
		return "game-" + bare, "/" + dir + "/" + bare + "/index.html", "game"

	case "section_index":
		// Adoption shape: slug="guides-index" (with -index suffix).
		// Planner shape: slug="guides" (bare).
		// Optional explicit section overrides slug-derivation.
		// ParentSection is intentionally NOT honoured here — a section
		// index page IS its section, so nesting it under another
		// section makes no sense.
		sec := section
		if sec == "" {
			sec = strings.TrimSuffix(slug, "-index")
		}
		if sec == "" {
			return "", "", ""
		}
		return sec + "-index", "/" + sec + "/index.html", "section_index"

	case "blog_post":
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
		return slug, "/" + dir + "/" + slug + ".html", "blog_post"

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
