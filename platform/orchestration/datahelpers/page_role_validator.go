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
// Every one of these is structurally a section_index (a directory landing
// page that lists the sub-pages beneath it), but the LLM picked
// "content" / "blog-index" / "content" respectively. Without correction,
// these write to pages with role=content and url=/<slug>.html, breaking
// every internal link the source site had to /<slug>/index.html.
//
// Doc 030 Q3: role assignment is a deterministic concern, not an LLM one.
// The plan-builder LLM is allowed to be sloppy about role; this validator
// catches and corrects before anything is persisted.
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
//  1. Explicit role "index" or page name "index" → role="index", regardless
//     of LLM input.
//
//  2. Slug appears as ParentSection of any other page in the set →
//     role="section_index". This is the structural signal: a page that
//     other pages claim as their parent IS a section index, no matter
//     what the LLM said. This catches the gamesdesign case directly.
//
//  3. LLM-supplied URL ends in "/index.html" AND slug appears as the
//     directory immediately above (e.g. url="/tools/index.html" and
//     slug="tools") → role="section_index". This is a fallback for when
//     the LLM correctly emits the URL pattern but mislabels the role.
//
//  4. LLM-supplied URL fits a nested role pattern AND role disagrees:
//     url="/tools/<x>/index.html" with role!="tool" → role="tool".
//     url="/guides/<x>/index.html" with role!="guide" → role="guide".
//     url="/games/<x>/index.html" with role!="game" → role="game".
//     This catches the symmetric case where the LLM gets the URL right
//     but the role wrong for nested pages.
//
//  5. Otherwise: accept LLM's role as-is, normalising "blog-post" /
//     "blog-index" hyphenated variants to underscored canonical forms.
//
// Rule precedence matters because rule 2 is structurally stronger than
// rule 3 (a page being someone's declared parent is decisive; URL
// pattern matching is heuristic).
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

		// Rule 1: index identity.
		if rawRole == "index" || v.Name == "index" || slug == "index" {
			correctedRole = "index"
		} else if declaredParents[slug] || declaredParents[slug+"-index"] {
			// Rule 2: structural — this slug is someone's parent.
			correctedRole = "section_index"
		} else if isSectionIndexURL(p.URL, slug) {
			// Rule 3: URL pattern says section_index.
			correctedRole = "section_index"
		} else if r, ok := nestedRoleFromURL(p.URL); ok && r != rawRole {
			// Rule 4: URL pattern says nested role.
			correctedRole = r
		}

		v.Role = correctedRole
		if correctedRole != rawRole && rawRole != "" {
			v.CorrectedFromRole = rawRole
		}
		out[i] = v
	}

	return out
}

// normaliseRole converts hyphenated LLM variants to canonical underscore
// form and lowercases everything.
func normaliseRole(r string) string {
	r = strings.ToLower(strings.TrimSpace(r))
	switch r {
	case "blog-post":
		return "blog_post"
	case "blog-index", "section-index":
		return "section_index"
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
