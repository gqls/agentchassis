// FILE: platform/content/markup_balance.go
//
// Markup-context structural tag balance — the truncation-signature counter
// shared by every guard that asks "was this generation cut mid-stream?".
// Callers: toolTemplateValid + componentTemplateValid (plan_sections, the tool
// birth/load gates), hasUnbalancedStructuralTags + componentRegressionIssues
// (component_write_guard, the section birth and rewrite gates), and
// discovery_checks/check_truncated_component (the fleet sweep, its verifier,
// and its intact-version probe). It lives in this leaf package so the
// discovery_checks package can import it too — before this file, that package
// kept a hand-maintained MIRROR of the pair list because `actions` imports
// `discovery_checks` and the reverse import is a cycle.
//
// bugs_open/303: the previous counters were raw substring counts —
// strings.Count(folded, "<script") vs strings.Count(folded, "</script>") over
// the whole file — with no awareness of string literals, regular expressions,
// code comments or prose. A tool that MANIPULATES markup necessarily mentions
// tags in its own JavaScript; one unpaired mention (a comment
// `// protect <style> blocks`, a regex /<script[^>]*>/) tipped opens over
// closes and the tool was refused at birth, permanently, as "truncated" —
// while the guard's own recorded context said ends_cleanly: true. Both live
// HTML-manipulating tools passed by EXACTLY ZERO margin, surviving only on a
// lucky regex phrasing (an alternation group in which the substring "<script"
// never occurs).
//
// This scanner counts a token ONLY when it is
//   (a) TAG-SHAPED — the name is followed by whitespace, '>', '/', or (for an
//       open tag) end-of-input, so "<div" does not match "<divider"; and
//   (b) in MARKUP context:
//       - <!-- --> comment bodies are skipped (an unterminated comment
//         consumes the rest of the input, as in a browser);
//       - the raw-text bodies of <script> and <style> are skipped up to the
//         first case-insensitive "</script"/"</style" — exactly where a
//         browser ends the element, so a literal "</script>" inside a JS
//         string counts as the close BECAUSE the browser treats it as one
//         (working tools escape it as "<\/script>", which does not match);
//       - '>' inside a quoted attribute value does not end a tag.
//
// Every true positive of the substring counter is kept. A generation cut
// mid-JavaScript leaves "<script" with no "</script" anywhere after it, so
// the scanner reaches end-of-input still inside the body and the open stays
// unmatched — the bugs_open/012 / 046 signature. A cut mid-markup leaves a
// counted <div/<section open. A cut inside an HTML comment is invisible to
// tag balance under BOTH countings (the tool path's ends-cleanly check still
// catches it). A close tag cut before its '>' ("...</scrip" or "...</script")
// is not counted as a close, matching the old token which included the '>'.
//
// One REPORTING difference from the substring counter, deliberate: in a
// template cut inside a raw-text body (e.g. an unterminated <style>), tags
// appearing AFTER the cut are style text to a browser, so they are neither
// opened nor closed here — the old counter reported them all as unbalanced.
// The verdict on such a template is identical (truncated); only the token
// list in the message is shorter, and more honest.
//
// Calibration 2026-08-18 (per component_write_guard.go's standing
// instruction), old vs new over the full live population:
//   - component_versions (264 rows ≥100 chars): 26 flagged by BOTH, 0 verdict
//     flips — every recorded casualty is kept.
//   - comparative check-2 over the 121 consecutive version pairs: 1 block
//     under both (the bugs_open/012 write), 0 disagreements.
//   - content_components (300 rows): old flagged 11, new flags 8 — a strict
//     subset, so zero NEW positives. The 3 un-flagged rows were each
//     hand-read: all are mentions inside CSS comments ("do not move this
//     block above the base <style>", "Inline <script> means no …"), i.e. the
//     false-positive class this file exists to remove — and TWO of them
//     carried open truncated_component work items filed on the strength of
//     the substring count (bugs_open/303 §calibration). All 8 still-flagged
//     rows are inactive historical casualties.
//   - the bugs_open/303 recipe on the real stored html-minifier /
//     svg-optimizer templates: as stored, both predicates pass; with the
//     mention comment injected, old refuses (the bug) and new passes; cut at
//     60%, BOTH refuse, new naming the unterminated <script.

package content

import "strings"

// StructuralTagPair is one open/close token pair whose imbalance (more opens
// than closes) signals a generation cut mid-stream.
type StructuralTagPair struct {
	Open  string // e.g. "<script"
	Close string // e.g. "</script>"
}

// StructuralTagPairs is the CANONICAL pair list, shared by every truncation
// guard. History: <script>/<style>/<section> were the original three;
// <div> and <fieldset> were added after the council gate's edit-quality seat
// noted the bugs_open/012 wreck was missing those too — simulated before
// adding, zero additional blocks across the recorded component_versions
// transitions. structural_tag_pairs_test.go pins the list; if you change it,
// re-run the calibration in component_write_guard.go's header first.
var StructuralTagPairs = []StructuralTagPair{
	{"<script", "</script>"},
	{"<style", "</style>"},
	{"<section", "</section>"},
	{"<div", "</div>"},
	{"<fieldset", "</fieldset>"},
}

// TagBalance carries the markup-context counts for one structural pair.
type TagBalance struct {
	StructuralTagPair
	Opens  int
	Closes int
}

// StructuralTagCounts scans html once and returns per-pair open/close counts
// in StructuralTagPairs order, counting only tag-shaped tokens in markup
// context (see the file header for the exact semantics).
func StructuralTagCounts(html string) []TagBalance {
	counts := make([]TagBalance, len(StructuralTagPairs))
	for i, p := range StructuralTagPairs {
		counts[i] = TagBalance{StructuralTagPair: p}
	}
	// Case-folded copy used for ALL matching; counts need no original offsets,
	// so a fold that changes byte lengths (non-ASCII) cannot misalign anything.
	s := strings.ToLower(html)
	i := 0
	for i < len(s) {
		lt := strings.IndexByte(s[i:], '<')
		if lt < 0 {
			break
		}
		i += lt
		rest := s[i:]

		// HTML comment: skip the body entirely. Unterminated ⇒ the rest of
		// the input is comment text, as in a browser.
		if strings.HasPrefix(rest, "<!--") {
			end := strings.Index(rest[len("<!--"):], "-->")
			if end < 0 {
				break
			}
			i += len("<!--") + end + len("-->")
			continue
		}

		// Close tag of a tracked pair. Counted only if its '>' is present —
		// a close cut before '>' did not count under the old token either.
		if k, ok := matchStructuralToken(rest, "</"); ok {
			next, closed := skipPastTagEnd(s, i)
			if closed {
				counts[k].Closes++
			}
			i = next
			continue
		}

		// Open tag of a tracked pair.
		if k, ok := matchStructuralToken(rest, "<"); ok {
			counts[k].Opens++
			next, closed := skipPastTagEnd(s, i)
			i = next
			if !closed {
				break // input ends inside the open tag — a cut
			}
			name := counts[k].Open[1:]
			if name == "script" || name == "style" {
				// Raw-text body: markup rules are suspended until the first
				// matching close token, exactly as a browser reads it.
				closeAt := rawTextCloseIndex(s, i, name)
				if closeAt < 0 {
					break // body never closed — the open stays unmatched
				}
				next, closed = skipPastTagEnd(s, closeAt)
				if closed {
					counts[k].Closes++
				}
				i = next
				if !closed {
					break
				}
			}
			continue
		}

		i++ // a '<' that is not a comment or a tracked tag
	}
	return counts
}

// UnbalancedStructuralTags returns the open tokens (e.g. "<script") whose
// markup-context opens exceed closes, in StructuralTagPairs order. Empty means
// every pair balances — no truncation signature.
func UnbalancedStructuralTags(html string) []string {
	var bad []string
	for _, tb := range StructuralTagCounts(html) {
		if tb.Opens > tb.Closes {
			bad = append(bad, tb.Open)
		}
	}
	return bad
}

// matchStructuralToken reports which StructuralTagPairs entry, if any, begins
// at rest with the given prefix ("<" open, "</" close), requiring the name to
// be tag-shaped: followed by whitespace, '>', '/', or end-of-input. No tracked
// name is a prefix of another, so first match is the only match.
func matchStructuralToken(rest, prefix string) (int, bool) {
	if !strings.HasPrefix(rest, prefix) {
		return 0, false
	}
	after := rest[len(prefix):]
	for k, p := range StructuralTagPairs {
		name := p.Open[1:]
		if strings.HasPrefix(after, name) && nameDelimited(after, len(name)) {
			return k, true
		}
	}
	return 0, false
}

// nameDelimited reports whether s[idx] terminates a tag name. End-of-input
// counts: "...<script" cut at the very end is still the open it looks like.
func nameDelimited(s string, idx int) bool {
	if idx >= len(s) {
		return true
	}
	switch s[idx] {
	case ' ', '\t', '\n', '\r', '\f', '>', '/':
		return true
	}
	return false
}

// skipPastTagEnd advances from the '<' at i past the tag's terminating '>',
// ignoring '>' inside quoted attribute values. Returns the index after '>'
// and whether one was found; absent, the input ended inside the tag.
func skipPastTagEnd(s string, i int) (int, bool) {
	var quote byte
	for j := i; j < len(s); j++ {
		c := s[j]
		if quote != 0 {
			if c == quote {
				quote = 0
			}
			continue
		}
		switch c {
		case '"', '\'':
			quote = c
		case '>':
			return j + 1, true
		}
	}
	return len(s), false
}

// rawTextCloseIndex returns the index of the first tag-shaped close token
// ("</"+name) at or after from, or -1. Tag-shaped matters here too: a JS
// string "</scriptFoo" must not end the element it would not end in a browser.
func rawTextCloseIndex(s string, from int, name string) int {
	tok := "</" + name
	for j := from; j < len(s); {
		idx := strings.Index(s[j:], tok)
		if idx < 0 {
			return -1
		}
		j += idx
		if nameDelimited(s, j+len(tok)) {
			return j
		}
		j += len(tok)
	}
	return -1
}
