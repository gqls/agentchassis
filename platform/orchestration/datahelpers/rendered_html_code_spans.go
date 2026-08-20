// FILE: platform/orchestration/datahelpers/rendered_html_code_spans.go
//
// The rendered_html-surface repair for literal markdown CODE SPANS: `x` in
// assertion text becomes <code>x</code>, spliced into the original bytes.
// Built 2026-08-20 for bugs_open/277 §5 (owner ruling, same day): a component
// whose content_data cannot reproduce its own rendered_html is unreachable by
// every regenerate-from-source route, so its finding can only be repaired by
// editing the finished HTML. Carried by apply_section_edit's
// rendered_html_transform edit type (section_editor_actions.go), opt-in there.
//
// WHY THIS LIVES IN datahelpers, NEXT TO THE PATTERNS. The repair contract for
// this surface is: after ConvertLiteralCodeSpansInHTML, the discovery scan
// (ExtractAssertionText + LiteralMarkdownPatterns) finds no code_span this
// transform could reach. Detection walks the parsed tree skipping
// nonAssertionElements (claims.go); this transform must skip EXACTLY the same
// subtrees or the two drift — a repairer that converts inside <script> corrupts
// JS template literals (the live fixture in testdata/ carries one), and one
// that skips MORE than the detector strands items in 'failed' for ever (the
// verifier-stricter-than-repair tale in verifier_coverage_test.go). Same
// package, same skip map, one source.
//
// WHY A TOKENIZER SPLICE AND NOT PARSE + RE-SERIALISE. html.Parse wraps a
// fragment in <html><head><body> and re-serialising normalises the whole
// document (attribute quoting, entities, tag case) — the diff would exceed the
// edit, on a component whose rendered_html is the ONLY copy of its content.
// html.NewTokenizer emits the source as tokens whose Raw() bytes we copy
// verbatim; the output differs from the input ONLY inside converted text
// nodes. converted==0 therefore implies byte-identical output, which the
// tests assert.
//
// WHY INSERTING MARKUP IS SAFE HERE AND BANNED IN THE STRIPPER. literal_markdown.go's
// header rules out markdown→HTML conversion because StripLiteralMarkdown runs on
// content_data values that FEED the unescaping render pipe (text/template, zero
// escaping) — a converter there is an injection surface for LLM-authored markup.
// This transform runs on the pipe's OUTPUT: the bytes between the backticks are
// already being served as text, are never LLM-authored at transform time, and the
// only bytes ADDED are the two fixed strings "<code>" and "</code>". The interior
// character class excludes `<`, `>` and backtick, so the wrapped region cannot
// contain a tag boundary and the inserted pair cannot misnest.
//
// WHY A THIRD HTML INSTRUMENT AND NOT A SHARED WALKER (council corr b72a4029
// r1, reuse_agent). This package now reads/rewrites HTML three ways, each fit
// to its job and none substitutable: link_repair.go rewrites ANCHOR ATTRIBUTES
// with a regex over <a> tags (repairAnchorRe — no tree, no text nodes);
// claims.go EXTRACTS TEXT with a full html.Parse tree walk (lossy by design —
// entities decoded, whitespace collapsed, inline tags fused); this file EDITS
// TEXT NODES byte-preservingly with the tokenizer, because the other two
// instruments each destroy exactly what this one must preserve (the parse
// walk cannot re-serialise without normalising; the attribute regex never
// sees text). What they share — the skip SET — is already shared:
// nonAssertionElements, one map, this package.
//
// CONVERSION ⊆ DETECTION, deliberately. codeSpanConvertRe is MDCodeSpanRe with
// `<` and `>` also excluded from the interior. Detection additionally sees
// entity-decoded, whitespace-collapsed, inline-fused text (ExtractAssertionText),
// so a span that crosses inline elements (`fetch<em>()</em>`), carries an
// entity-encoded backtick, or wraps across a newline is DETECTABLE but NOT
// convertible. Those stay unconverted; the caller refuses (converted==0 → error),
// the item fails verification and routes to a human — the safe direction. Do not
// widen the conversion class to close that gap: a conversion broader than
// detection writes markup the detector never objected to.

package datahelpers

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// codeSpanConvertRe is the conversion counterpart of MDCodeSpanRe
// (literal_markdown.go): same opening discipline (first interior char
// alphanumeric — `${x}` interpolations and `/api` paths never fire), same
// length bound, interior additionally excludes '<' and '>' so a converted
// region can never contain a tag boundary. The ⊆ relationship to MDCodeSpanRe
// is enforced property-wise in rendered_html_code_spans_test.go — edit the two
// together or that test fails.
var codeSpanConvertRe = regexp.MustCompile("`([A-Za-z0-9][^`<>\n]{0,80})`")

// ConvertLiteralCodeSpansInHTML rewrites literal markdown code spans that sit
// in assertion text of an HTML fragment into <code> elements: `x` →
// <code>x</code>. Returns the rewritten fragment and how many spans were
// converted.
//
// Guarantees:
//   - every byte outside a converted text node is copied verbatim (Raw splice);
//   - text inside nonAssertionElements subtrees (script, style, code, pre,
//     textarea, svg, iframe, noscript, template, select, option, head) is never
//     touched — matching the detector's skip set exactly;
//   - attribute values are never touched (they live inside tag tokens, which
//     are copied whole);
//   - idempotent: a converted span's text now lives inside <code>, a skipped
//     subtree, so a second pass converts nothing;
//   - converted == 0 ⇒ output == input, byte for byte.
//
// On a tokenizer error other than EOF the ORIGINAL fragment is returned with
// the error: the caller must treat that as "refused, live section unchanged".
// An unclosed skip element suppresses conversion for the remainder of the
// fragment (conservative: when nesting is unclear, do not edit).
func ConvertLiteralCodeSpansInHTML(fragment string) (string, int, error) {
	z := html.NewTokenizer(strings.NewReader(fragment))
	var out strings.Builder
	out.Grow(len(fragment) + 64)
	converted := 0
	skipDepth := 0 // >0 ⇒ inside a nonAssertionElements subtree

	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			if err := z.Err(); err != io.EOF {
				return fragment, 0, fmt.Errorf("code-span transform: tokenizer error, fragment left unchanged: %w", err)
			}
			return out.String(), converted, nil

		// ⚠ ORDER IS LOAD-BEARING in the two tag cases: TagName() lower-cases
		// the tag-name bytes IN PLACE in the tokenizer's buffer (x/net/html
		// escape.go `lower`), and Raw() aliases that same buffer — so Raw()
		// after TagName() returns `<Script>` as `<script>`, a silent byte
		// mutation outside any converted text node. Write (which copies) must
		// run first. The mixed-case fixture in the tests exists to catch a
		// reorder here.
		case html.StartTagToken:
			out.Write(z.Raw())
			if name, _ := z.TagName(); nonAssertionElements[string(name)] {
				skipDepth++
			}

		case html.EndTagToken:
			out.Write(z.Raw())
			if name, _ := z.TagName(); nonAssertionElements[string(name)] && skipDepth > 0 {
				skipDepth--
			}

		case html.TextToken:
			raw := z.Raw()
			if skipDepth > 0 || !strings.Contains(string(raw), "`") {
				out.Write(raw)
				break
			}
			text := string(raw)
			converted += len(codeSpanConvertRe.FindAllStringIndex(text, -1))
			out.WriteString(codeSpanConvertRe.ReplaceAllString(text, "<code>$1</code>"))

		default:
			// SelfClosingTagToken, CommentToken, DoctypeToken: copied verbatim.
			// None of the skip elements is void, so a self-closing tag never
			// opens a skip subtree.
			out.Write(z.Raw())
		}
	}
}
