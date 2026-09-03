// FILE: platform/orchestration/datahelpers/claims_plaintext.go
//
// The PLAIN-TEXT sibling of ExtractAssertionText (bugs_open/123).
//
// WHY THIS EXISTS, and it is not a convenience. `ExtractAssertionText` is an
// HTML parser: `extractAssertions` (claims.go:327) flushes a block boundary on
// entering and leaving each element in `assertionBlockElements`. Plain text and
// markdown contain no such elements, so `html.Parse` wraps the whole document
// in one `<body>` text node and `wsCollapseRe` joins every line with a single
// space. MEASURED 2026-08-03 with a probe test, not reasoned:
//
//	input:  "## Why recovery matters\n\nIndustry data shows … 3% and 10% …\n\n
//	         A second paragraph, with 66% of people agreeing.\n"
//	output: 1 block —
//	        "## Why recovery matters Industry data shows … A second paragraph…"
//
// The heading fuses to the following sentence with NO FULL STOP between them.
// Every pattern in this layer is windowed by `[^.]{0,N}`, and `negatedClaimMatch`
// scans backwards to the nearest clause boundary — so a fused block lets a match
// window, and a negation cue, cross a paragraph boundary the prose does not have.
// That is the identical hazard the claims floor recorded when it chose to scan
// sections individually rather than as one joined document:
//
//	"joining raw section HTML can fuse a trailing fragment of one section to the
//	 opening of the next into a sentence neither section contains"
//	— save_sections_claims_guard.go:116-121
//
// A scan that reports on sentences the document does not contain is worse than
// no scan, because its findings look specific.
//
// content-creator's request declares `format` as one of markdown / html /
// plain_text (contentcreator/agent.go:48), so two of its three output formats
// need this. It lives here rather than in that service because the next
// site-less producer needs the same thing, and because a private copy of an
// assertion splitter is `bugs_open/093`'s shape ("one call site of a shared
// judgement gets the rigorous fix; the sibling stays heuristic").

package datahelpers

import (
	"regexp"
	"strings"
)

var (
	// Code is not an assertion — the same ruling `nonAssertionElements` makes
	// for <code>/<pre>. A fenced block collapses to a newline so the lines
	// either side of it stay separate blocks rather than fusing.
	fencedCodeRe = regexp.MustCompile("(?s)```.*?```|~~~.*?~~~")
	inlineCodeRe = regexp.MustCompile("`[^`\n]*`")

	// Structural markers at the head of a line: ATX headings, blockquote
	// markers, bullet and ordered list markers. Stripped so the assertion text
	// reads as prose; the LINE BREAK is what carries the block boundary, so
	// removing the marker cannot fuse anything.
	// ⚠ THIS PACKAGE HOLDS TWO MARKDOWN VOCABULARIES, DELIBERATELY. This one is
	// looser than literal_markdown.go's (it covers blockquotes and ordered-list
	// markers, which that file has no pattern for, and mdEmphasisRe below unwraps
	// `_italic_` and `*single*` which MDBoldRe does not). Do NOT unify them
	// without reading the WHY TWO TIERS block in literal_markdown.go first.
	//
	// The asymmetry is one of CONSEQUENCE, not taste. This splitter's output is
	// never served — it feeds a scanner, so a wrong strip here costs a missed or
	// an extra finding. literal_markdown.go's output IS served, through a
	// text/template pipe that escapes nothing, so a wrong strip there is a visible
	// mutation of a customer's page. The looser vocabulary is right for the first
	// job and would be reckless for the second.
	//
	// And where they genuinely overlap they disagree ON PURPOSE: this file KEEPS
	// markdown link URLs (see below — they are the citation evidence
	// ScanAttributedUncitedStats reads), where literal_markdown.go deletes them.
	// That is the strongest single argument against merging the two.
	mdLinePrefixRe = regexp.MustCompile(`^\s{0,3}(?:#{1,6}\s+|>\s?|[-*+]\s+|\d{1,3}[.)]\s+)`)

	// Emphasis runs, unwrapped to their content. Deliberately conservative: it
	// only unwraps a run with no internal marker character, so an asterisk used
	// as a footnote glyph is left alone rather than eating the line after it.
	mdEmphasisRe = regexp.MustCompile(`(\*{1,3}|_{1,3})([^*_\n]+)(\*{1,3}|_{1,3})`)
)

// SplitPlainAssertionText splits plain text or markdown into assertion blocks,
// one per non-empty line.
//
// Line-per-block is deliberate and is the safe direction: it can only make a
// window SMALLER than the prose the author wrote, never larger, so it cannot
// invent a sentence. A claim genuinely spanning two lines of one paragraph will
// be split and may go unmatched — a missed finding, which is the failure this
// layer already tolerates everywhere (the numeric scan's whole design), rather
// than a fabricated one, which it does not.
//
// Markdown LINK URLS ARE DELIBERATELY KEPT in the block. They are not prose, but
// they are the citation evidence `ScanAttributedUncitedStats` reads — stripping
// them here would make every cited figure look uncited, which is the exact
// false-positive class the corpus measurement caught.
func SplitPlainAssertionText(text string) []string {
	if text == "" {
		return nil
	}

	text = fencedCodeRe.ReplaceAllString(text, "\n")
	text = inlineCodeRe.ReplaceAllString(text, " ")

	var blocks []string
	for _, line := range strings.Split(text, "\n") {
		line = mdLinePrefixRe.ReplaceAllString(line, "")
		line = mdEmphasisRe.ReplaceAllString(line, "$2")
		line = strings.TrimSpace(wsCollapseRe.ReplaceAllString(line, " "))
		if line != "" {
			blocks = append(blocks, line)
		}
	}
	return blocks
}

// AssertionBlocks routes text to the right splitter for its declared format.
//
// The format string is the producer's own declaration (content-creator's
// `data.format`), and an UNKNOWN or EMPTY format is treated as plain text, not
// as HTML. That direction is the safe one: running the plain splitter over HTML
// yields one block per source line, which under-matches; running the HTML
// parser over markdown yields ONE block for the whole document, which
// over-matches across boundaries that do not exist. Under-match is a missed
// finding; over-match is a fabricated one.
func AssertionBlocks(text, format string) []string {
	if strings.EqualFold(strings.TrimSpace(format), "html") {
		return ExtractAssertionText(text)
	}
	return SplitPlainAssertionText(text)
}
