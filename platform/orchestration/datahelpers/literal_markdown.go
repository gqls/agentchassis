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

// WHY TWO TIERS (bugs_open/332, 2026-09-03)
// ------------------------------------------
// Everything above is TIER 1: detection and strip, single-sourced, applied
// everywhere a text-typed value is written or scanned. Below StripLiteralMarkdown
// there is a TIER 2 — StripFeedDisplayMarkdown — which strips MORE and detects
// NOTHING. That asymmetry is deliberate and this block is the argument for it.
//
// The contract is Scan(Strip(x)) == nothing, which requires strip ⊇ scan. It says
// nothing about the other direction, so a strip broader than the scan is legal.
// rendered_html_code_spans.go already establishes the mirror of this
// (conversion ⊊ detection) and its warning is about INSERTION: "a conversion
// broader than detection writes markup the detector never objected to". Tier 2
// only DELETES, so it can neither insert nor strand an item in 'failed'.
//
// What tier 2 holds, and why each one is disqualified from detection:
//
//	list markers   A (?m)^ anchor means two DIFFERENT things on the two surfaces
//	               the check scans. ExtractAssertionText yields one block per
//	               <li>, so on rendered_html ^ is BLOCK start; on content_data it
//	               is line start. A <li> legitimately opening with a hyphen would
//	               file a finding no repair can clear; the item terminates
//	               wont_fix/failed, both excluded from idx_swi_dedup, so it
//	               re-files weekly — and wont_fix is excluded from both sides of
//	               the promoter ratio, so the loop is INVISIBLE in the metric you
//	               would check it with (migration 499 says exactly this).
//	bracket tails  A lone unclosed `[` is one character of harm and the weakest
//	               evidence in the set. It is only sound GIVEN that the value was
//	               truncated, and the feed projection is the only place that
//	               knowledge exists.
//	bold tails     `**kwargs`, `**args` and footnote glue are live shapes on the
//	               developer-facing sites in this estate. The guards below exclude
//	               the known cases, but confining the rule to feed snippets — text
//	               we never author — keeps the residual away from all seven
//	               content_data strip seams.
//
// So: a feed snippet is a fragment of a scraped markdown DOCUMENT, cut mid-token
// by our own 197-byte truncation. Our own content_data is prose an LLM was told
// not to put markdown in. Two populations, two appetites for risk.

// Detection patterns. Exported so the discovery check, its verifier and any
// future consumer match EXACTLY what the stripper removes.
var (
	MDBoldRe     = regexp.MustCompile(`\*\*[A-Za-z][^*\n]{0,80}\*\*`)
	MDCodeSpanRe = regexp.MustCompile("`[A-Za-z0-9][^`\n]{0,80}`")
	MDHeadingRe  = regexp.MustCompile(`(?m)^#{1,6} \S`)
	MDLinkRe     = regexp.MustCompile(`\[[A-Za-z][^\]\n]{0,80}\]\((?:https?://|/)[^)\s]{0,200}\)`)
	// MDLinkTruncatedRe is MDLinkRe with the closing `)` replaced by end-of-text,
	// buying a LEFT WORD BOUNDARY to pay for the delimiter it has lost. That trade
	// is the whole design: `)` was carrying the discrimination, `$` plus a
	// boundary carries it now. Without the boundary, `config[Debug](/api/v2/logs`
	// would fire; with it, only a bracket opening a word does.
	//
	// It exists because the half-patterns are OURS. websearch/providers/
	// firecrawl.go cuts every snippet at 197 bytes, so 288 of 5,863 feed rows in
	// 30 days [MEASURED 2026-09-03] carry a link severed mid-URL — a shape
	// MDLinkRe cannot match by construction, which is why the news pages measured
	// clean in August 2026 and served dirty in September (bugs_open/332).
	// ⚠ `!?` ADDED 2026-09-04, and it is the whole of bugs_open/332's residual.
	// The first cut's left boundary was `(?:^|[\s(])`, which EXCLUDES a preceding
	// `!` — and `!` is exactly the markdown image marker. So `![alt](url…`, an
	// image whose alt text closed but whose URL was severed, fell through every
	// rule in this file: mdImageStripRe needs the closing `)`, this pattern's
	// boundary rejected the `!`, and mdFeedImageTailRe requires no `]` at all.
	// Found live on idea.uk's news-listing, re-rendered 2026-09-04 10:49Z — HOURS
	// after the roll that shipped the rest of this fix, so it was a genuine gap
	// and not a stale binary (proven separately: dartsonline's column was dirty,
	// its page re-rendered post-roll, and it came out clean).
	MDLinkTruncatedRe = regexp.MustCompile(`(?:^|[\s(])!?\[[A-Za-z][^\]\n]{0,80}\]\((?:https?://|/)[^)\s]{0,200}$`)
	// A value carrying markup or script is not a text-typed field — the
	// code-span and md-link patterns are suppressed there.
	HTMLMarkupRe = regexp.MustCompile(`<[A-Za-z/!]`)
)

// Strip counterparts. The capture groups keep the human-visible text and drop
// only the marker characters. These MUST stay derivable from the detection
// patterns above — literal_markdown_test.go's fixpoint property is the guard.
var (
	mdBoldStripRe     = regexp.MustCompile(`\*\*([A-Za-z][^*\n]{0,80})\*\*`)
	mdCodeSpanStripRe = regexp.MustCompile("`([A-Za-z0-9][^`\n]{0,80})`")
	mdHeadingStripRe  = regexp.MustCompile(`(?m)^#{1,6} `)
	mdLinkStripRe     = regexp.MustCompile(`\[([A-Za-z][^\]\n]{0,80})\]\((?:https?://|/)[^)\s]{0,200}\)`)
	// mdImageStripRe has NO detection counterpart, and that is correct rather than
	// an omission: MDLinkRe already fires on the inner `[alt](url)` of every
	// letter-alt image, so a separate `md_image` pattern NAME would add zero
	// detection while perturbing the two things keyed on pattern names —
	// transformRouteSlot's routing and the check's exact-pattern-set test.
	//
	// It must run BEFORE mdLinkStripRe. Otherwise the link strip eats `[alt](url)`
	// out of `![alt](url)` and leaves a stray `!` — which is what this file did
	// until 2026-09-03, and the only reason strip("![alt](url)") was "!alt".
	//
	// The alt must start with a letter, so `![](url)` (30 of 123 image-bearing
	// feed rows [MEASURED 2026-09-03]) is deliberately left alone: an image token
	// with no alt text has no visible text to keep, and deleting it wholesale is
	// the one shape that could manufacture a blank.
	mdImageStripRe = regexp.MustCompile(`!\[([A-Za-z][^\]\n]{0,80})\]\((?:https?://|/)[^)\s]{0,200}\)`)
	// mdLinkTruncatedStripRe is MDLinkTruncatedRe with the visible text captured.
	// Applied through stripTruncatedLink, NOT ReplaceAllString — see there.
	mdLinkTruncatedStripRe = regexp.MustCompile(`(^|[\s(])\[([A-Za-z][^\]\n]{0,80})\]\((?:https?://|/)[^)\s]{0,200}$`)
	// mdImageTruncatedUrlStripRe is the image form of the rule above: alt text
	// CLOSED, URL severed. It must run BEFORE mdLinkTruncatedStripRe for the same
	// reason mdImageStripRe runs before mdLinkStripRe — otherwise the link rule
	// would keep the `!` as part of its boundary capture. Keeps the alt text,
	// matching what the complete-image rule already does (`![alt](url)` -> `alt`,
	// council 060bcc0a's property: a bare image token keeps its visible text).
	mdImageTruncatedUrlStripRe = regexp.MustCompile(`(^|[\s(\[])!\[([A-Za-z][^\]\n]{0,80})\]\((?:https?://|/)[^)\s]{0,200}$`)
)

// TIER 2 — feed display only. In no detector; see the WHY TWO TIERS block above.
var (
	mdFeedListMarkerRe  = regexp.MustCompile(`(?m)^[ \t]{0,3}[-*+][ \t]+`)
	mdFeedBracketTailRe = regexp.MustCompile(`(^|[\s(])\[([A-Za-z][^\]\n]{0,80})$`)
	// mdFeedImageTailRe handles an image opener severed before its `]`. Anchored
	// at end so it can only match the truncated case — a COMPLETE `![](url)` is
	// untouched, which matters because turning it into `](url)` would be worse
	// than leaving it. The worked case is `[![Results: … titles in...`, where this
	// rule strips `![` and the bracket-tail rule then takes the remaining `[` on
	// the next pass of the fixpoint loop.
	mdFeedImageTailRe = regexp.MustCompile(`(^|[\s(\[])!\[([A-Za-z][^\]\n]{0,80})$`)
	// mdFeedBoldTailRe carries FOUR guards, and the last two are why it is safe
	// enough for feed text and not safe enough for ours:
	//   digit guard          `2**10` never fires (the [A-Za-z] after `**`)
	//   letter-after guard   `a ** b`, `3 * 4` never fire (same)
	//   LEFT WORD BOUNDARY   `O(n**k`, `Free delivery**Terms apply` never fire —
	//                        the char before `**` must be start, space, `(` or `[`
	//   PHRASE guard         the surviving run must hold a space with 6+ more
	//                        characters after it, so `**args here` (a real Python
	//                        idiom on this estate's own AI sites) never fires
	// Residual, accepted and stated: `**kwargs, args in the loop` still fires.
	// At that length it is markdown more often than not, and 15 rows in 30 days
	// [MEASURED 2026-09-03] does not buy a fifth guard.
	mdFeedBoldTailRe = regexp.MustCompile(`(^|[\s(\[])\*\*([A-Za-z][^*\n]{0,60} [^*\n]{6,60})$`)
)

// stripTruncatedLink removes a link severed mid-URL, keeping the link TEXT and
// RE-EMITTING THE TRUNCATION MARKER.
//
// The marker is the point. firecrawl.go appends a literal "..." at its cut, and
// that ellipsis lives inside the URL the pattern deletes. A plain
// ReplaceAllString would therefore turn
//
//	"…punched himself out and [lost in the ninth round](https://sports.yahoo.com/…-..."
//
// into "…punched himself out and lost in the ninth round" — a grammatical,
// complete-looking sentence that IS NOT WHAT THE SOURCE SAID, with nothing left
// to tell the reader anything was cut. Today's output is ugly and honest; that
// one would be pretty and dishonest, on a paid customer's page, and no test or
// served-artefact grep could ever see it: TestStripNeverInserts asserts only
// length, and the result scans clean by construction.
//
// Still strip-only. The re-emitted "..." was in the input — it is the same three
// bytes, moved left, and the output is always shorter than the input because the
// bracket, the "](" and the scheme go with it.
// stripTruncatedImageURL removes an image whose URL was severed, keeping the alt
// text and RE-EMITTING THE TRUNCATION MARKER — same contract and same reason as
// stripTruncatedLink below, which see for why the marker matters.
func stripTruncatedImageURL(s string) string {
	return mdImageTruncatedUrlStripRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := mdImageTruncatedUrlStripRe.FindStringSubmatch(m)
		if sub == nil {
			return m
		}
		out := sub[1] + sub[2]
		switch {
		case strings.HasSuffix(m, "..."):
			out += "..."
		case strings.HasSuffix(m, "\u2026"):
			out += "\u2026"
		}
		return out
	})
}

func stripTruncatedLink(s string) string {
	return mdLinkTruncatedStripRe.ReplaceAllStringFunc(s, func(m string) string {
		sub := mdLinkTruncatedStripRe.FindStringSubmatch(m)
		if sub == nil {
			return m
		}
		out := sub[1] + sub[2]
		switch {
		case strings.HasSuffix(m, "..."):
			out += "..."
		case strings.HasSuffix(m, "…"):
			out += "…"
		}
		return out
	})
}

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
		// ⚠ WIRED 2026-09-04, and it should have been wired on 2026-09-03.
		// MDLinkTruncatedRe was declared, exported, documented as a detection
		// pattern and NEVER CALLED HERE. The strip removed truncated links; the
		// scan never looked for them. Two consequences worth stating because both
		// are the shapes this file exists to prevent:
		//   - the detector was blind to the exact defect bugs_open/332 is about,
		//     so a page serving one scanned CLEAN;
		//   - and TestStripThenScanFindsNothing passed VACUOUSLY for it — the
		//     fixpoint holds trivially when Scan cannot see the pattern at all.
		// Caught only by a test that asserted detection directly, on a live
		// production string. A declared-but-uncalled pattern is invisible to every
		// property test in this package, because they all go through this function.
		if m := MDLinkTruncatedRe.FindString(text); m != "" {
			out = append(out, [2]string{"md_link_truncated", m})
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
	out := stripTier1(s, includeCodeSpan)
	return out, out != s
}

// stripTier1 is StripLiteralMarkdown's body, factored out so the feed-display
// tier can interleave with it inside one fixpoint rather than running after it —
// tier 2's rules UNCOVER tier 1 shapes (a stripped `![` leaves a `[` for the
// bracket-tail rule) and vice versa.
func stripTier1(s string, includeCodeSpan bool) string {
	out := s
	for i := 0; i < 8; i++ {
		prev := out
		// Images before links: mdLinkStripRe would otherwise eat "[alt](url)"
		// out of "![alt](url)" and leave the stray "!".
		if includeCodeSpan {
			out = mdImageStripRe.ReplaceAllString(out, "$1")
			// Links before headings/bold: those strips would otherwise leave
			// "[Title](url)" intact only where it was the whole heading text, and
			// md links can carry bold/code inside their text.
			out = mdLinkStripRe.ReplaceAllString(out, "$1")
			out = stripTruncatedImageURL(out)
			out = stripTruncatedLink(out)
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
	return out
}

// StripFeedDisplayMarkdown is StripLiteralMarkdown plus the TIER 2 rules — the
// ones that are sound for a snippet cut out of a scraped markdown document and
// are NOT sound for prose we authored. Read the WHY TWO TIERS block above before
// moving anything between the two; the reasons are specific, not stylistic.
//
// Use it wherever `content_feed_items.source_summary` (or a title) is on its way
// to a visitor. Everything else keeps StripLiteralMarkdown.
//
// Superset by construction, and both directions are pinned by property tests:
// the output is a character-subset of StripLiteralMarkdown's output, and it still
// satisfies the repair contract (the scan finds nothing in it).
func StripFeedDisplayMarkdown(s string, includeCodeSpan bool) (string, bool) {
	out := s
	for i := 0; i < 8; i++ {
		prev := out
		out = stripTier1(out, includeCodeSpan)
		if includeCodeSpan {
			out = mdFeedImageTailRe.ReplaceAllString(out, "$1$2")
			out = mdFeedBracketTailRe.ReplaceAllString(out, "$1$2")
		}
		out = mdFeedBoldTailRe.ReplaceAllString(out, "$1$2")
		out = mdFeedListMarkerRe.ReplaceAllString(out, "")
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
