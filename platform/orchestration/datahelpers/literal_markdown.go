// FILE: platform/orchestration/datahelpers/literal_markdown.go
//
// One source for the literal-markdown pattern set and BOTH operations over it:
// detection (the scan the discovery check and its completion verifier run) and
// removal (the strip the write seams and the mechanical repair run). See
// bugs_open/184.
//
// WHY ONE FILE
// ------------
// The repair contract is: after StripLiteralMarkdown, the scan finds nothing.
// If the detector and the stripper each kept their own regexes, the first
// pattern edit that touched only one of them would either strand items in
// 'failed' for ever (verifier stricter than repair — the page_rerender
// cautionary tale in verifier_coverage_test.go) or stamp 'complete' over a
// still-served defect (repair stricter than verifier). Both halves live here,
// and literal_markdown_test.go enforces the contract property-wise:
// Scan(Strip(x)) == nothing, for the live corpus and generated composites.
//
// WHY STRIP-ONLY (never markdown→HTML)
// ------------------------------------
// The render pipe is text/template with ZERO escaping anywhere (the
// LANDMINES entry on RenderTemplateReportingMissing: "text/template escapes
// NOTHING, by design"), so a converter that INSERTS markup would be writing
// into an unescaping pipe — the injection surface CQ-019's deferral names.
// Stripping only ever deletes marker characters; the pipe receives a strict
// subset of the characters it receives today.
//
// FALSE-POSITIVE DISCIPLINE (the check's letter-guards, single-sourced here):
//
//	bold      \*\*[A-Za-z]...  — `3 * 4 = 12`, `2**10`, `a ** b` never fire
//	code span `[A-Za-z0-9]...  — `${x}` interpolations and `/api` paths never
//	          fire; applied only to values carrying no HTML markup (a
//	          markup-bearing value is not a text-typed field; backticks there
//	          are code, not prose)
//	heading   ^#{1,6} \S       — #fff, "#1 rated", href="#", "issue #12"
//	          never fire
//	md link   \[[A-Za-z]...\]((https?://|/)...) — [1] citations, array[0](x)
//	          indexing never fire (link text must start with a letter, target
//	          must be an http(s) or root-relative URL). Same markup-free
//	          suppression as code spans: inside markup-bearing values a
//	          bracket-paren sequence is likelier code than prose.
//
// The scan reports ONE finding per pattern per value (the repair is a
// value-level rewrite, so occurrence counts change nothing about what to do);
// the strip removes EVERY occurrence, to a bounded fixpoint, because a single
// pass can uncover a nested form (e.g. "## [Title](url)" needs heading AND
// link handling, "**`x`**" needs bold then code span).

package datahelpers

import (
	"regexp"
	"strconv"
	"strings"
)

// Detection patterns. Exported so the discovery check, its verifier and any
// future consumer match EXACTLY what the stripper removes.
var (
	MDBoldRe     = regexp.MustCompile(`\*\*[A-Za-z][^*\n]{0,80}\*\*`)
	MDCodeSpanRe = regexp.MustCompile("`[A-Za-z0-9][^`\n]{0,80}`")
	MDHeadingRe  = regexp.MustCompile(`(?m)^#{1,6} \S`)
	MDLinkRe     = regexp.MustCompile(`\[[A-Za-z][^\]\n]{0,80}\]\((?:https?://|/)[^)\s]{0,200}\)`)
	// A value carrying markup or script is not a text-typed field — the
	// code-span and md-link patterns are suppressed there.
	HTMLMarkupRe = regexp.MustCompile(`<[A-Za-z/!]`)
)

// Strip counterparts. The capture groups keep the human-visible text and drop
// only the marker characters. These MUST stay derivable from the detection
// patterns above — literal_markdown_test.go's fixpoint property is the guard.
var (
	mdBoldStripRe    = regexp.MustCompile(`\*\*([A-Za-z][^*\n]{0,80})\*\*`)
	mdCodeSpanStripRe = regexp.MustCompile("`([A-Za-z0-9][^`\n]{0,80})`")
	mdHeadingStripRe = regexp.MustCompile(`(?m)^#{1,6} `)
	mdLinkStripRe    = regexp.MustCompile(`\[([A-Za-z][^\]\n]{0,80})\]\((?:https?://|/)[^)\s]{0,200}\)`)
)

// LiteralMarkdownPatterns returns the (pattern name, matched text) pairs the
// scan finds in one plain-text value. includeCodeSpan gates the two patterns
// that are suppressed on markup-bearing values (code_span, md_link) — pass
// !HTMLMarkupRe.MatchString(value) for content_data values, true for text
// already extracted from rendered HTML (script/style/code subtrees are gone).
func LiteralMarkdownPatterns(text string, includeCodeSpan bool) [][2]string {
	var out [][2]string
	if m := MDBoldRe.FindString(text); m != "" {
		out = append(out, [2]string{"bold", m})
	}
	if includeCodeSpan {
		if m := MDCodeSpanRe.FindString(text); m != "" {
			out = append(out, [2]string{"code_span", m})
		}
		if m := MDLinkRe.FindString(text); m != "" {
			out = append(out, [2]string{"md_link", m})
		}
	}
	if m := MDHeadingRe.FindString(text); m != "" {
		out = append(out, [2]string{"heading", m})
	}
	return out
}

// StripLiteralMarkdown removes markdown marker characters from one plain-text
// value, keeping the human-visible text: **x** → x, `x` → x, "# H" → "H",
// [text](url) → text. Strip-only — no output character that was not in the
// input. Runs to a bounded fixpoint (nested forms uncover in later passes).
// Returns the cleaned value and whether anything changed.
func StripLiteralMarkdown(s string, includeCodeSpan bool) (string, bool) {
	out := s
	for i := 0; i < 5; i++ {
		prev := out
		// Links first: heading/bold strips would otherwise leave "[Title](url)"
		// intact only where it was the whole heading text, and md links can
		// carry bold/code inside their text.
		if includeCodeSpan {
			out = mdLinkStripRe.ReplaceAllString(out, "$1")
		}
		out = mdBoldStripRe.ReplaceAllString(out, "$1")
		out = mdHeadingStripRe.ReplaceAllString(out, "")
		if includeCodeSpan {
			out = mdCodeSpanStripRe.ReplaceAllString(out, "$1")
		}
		if out == prev {
			break
		}
	}
	return out, out != s
}

// StripLiteralMarkdownFromContentData walks every string leaf of a decoded
// content_data map, strips literal markdown in place, and returns the dotted
// paths of the fields it changed (empty slice = untouched). Keys beginning
// "_" are platform metadata (_built_at, ...), never writer output — skipped,
// mirroring the discovery check's walk. The markup-free suppression is applied
// per VALUE, exactly as the check applies it.
func StripLiteralMarkdownFromContentData(cd map[string]interface{}) []string {
	var changed []string
	var walk func(prefix string, v interface{}) interface{}
	walk = func(prefix string, v interface{}) interface{} {
		switch t := v.(type) {
		case map[string]interface{}:
			for k, val := range t {
				if strings.HasPrefix(k, "_") {
					continue
				}
				p := k
				if prefix != "" {
					p = prefix + "." + k
				}
				t[k] = walk(p, val)
			}
			return t
		case []interface{}:
			for i, val := range t {
				t[i] = walk(prefix+"["+strconv.Itoa(i)+"]", val)
			}
			return t
		case string:
			cleaned, did := StripLiteralMarkdown(t, !HTMLMarkupRe.MatchString(t))
			if did {
				changed = append(changed, prefix)
			}
			return cleaned
		default:
			return v
		}
	}
	walk("", cd)
	return changed
}
