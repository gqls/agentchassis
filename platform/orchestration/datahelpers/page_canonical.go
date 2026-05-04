// FILE: platform/orchestration/datahelpers/page_canonical.go
//
// CanonicalisePage maps any logical page descriptor — from adoption's
// analyze_site LLM, the site-planner's LLM, or a future reconciler — to
// the single canonical (name, url, page_type) triple that the pages
// table and `needs_page:<name>` work-item keys both use.
//
// Without this helper, adoption's LLM emits names like
// "tool-ttk-calculator" / "guides-index" while the planner's LLM emits
// "ttk-calculator" / "guides". Both surfaces upsert to pages with
// `ON CONFLICT (site_id, name) DO UPDATE`, but the conflict never fires
// because the names diverge — the result is duplicate rows and parallel
// `adoption_page_*` vs `needs_page:*` work-item streams that the
// `idx_swi_dedup` index cannot collapse. See doc 029 Phase 0.
//
// The helper is idempotent: feeding it the adoption-shape produces the
// same output as feeding it the planner-shape. It does not query the
// database — naming is decided purely from the descriptor.

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
// "-index".
type PageDescriptor struct {
	Role    string
	Slug    string
	Section string
}

// CanonicalisePage returns the canonical (name, url, pageType) for the
// descriptor. Empty inputs return empty triples; the caller decides
// whether to skip or default.
func CanonicalisePage(d PageDescriptor) (name, url, pageType string) {
	role := strings.ToLower(strings.TrimSpace(d.Role))
	slug := normaliseSlug(d.Slug)
	section := normaliseSlug(d.Section)

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
		return "tool-" + bare, "/tools/" + bare + "/index.html", "tool"

	case "guide":
		bare := strings.TrimPrefix(slug, "guide-")
		if bare == "" {
			return "", "", ""
		}
		return "guide-" + bare, "/guides/" + bare + "/index.html", "guide"

	case "game":
		bare := strings.TrimPrefix(slug, "game-")
		if bare == "" {
			return "", "", ""
		}
		return "game-" + bare, "/games/" + bare + "/index.html", "game"

	case "section_index":
		// Adoption shape: slug="guides-index" (with -index suffix).
		// Planner shape: slug="guides" (bare).
		// Optional explicit section overrides slug-derivation.
		sec := section
		if sec == "" {
			sec = strings.TrimSuffix(slug, "-index")
		}
		if sec == "" {
			return "", "", ""
		}
		return sec + "-index", "/" + sec + "/index.html", "section_index"

	case "blog_post":
		// Phase 0 keeps adoption's existing convention so already-deployed
		// blog URLs don't break: name=<slug>, url=/blog/<slug>.html. Phase 1
		// may revisit if blog gets dedicated section_index treatment.
		if slug == "" {
			return "", "", ""
		}
		return slug, "/blog/" + slug + ".html", "blog_post"

	default:
		// "content" and anything unrecognised: name=slug, url=/<slug>.html.
		// pageType passes through the original role string when set, so
		// unknown roles surface in logs rather than being silently
		// rewritten to "content".
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
