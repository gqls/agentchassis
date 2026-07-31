// FILE: platform/orchestration/datahelpers/runtime_fill.go
//
// ONE definition of the runtime-fill exemption, at ELEMENT scope.
//
// THE DEFECT THIS FIXES (bugs_open/137). "Is this control exempt because it
// hydrates client-side?" was answered in eight places by the same line —
// `strings.Contains(html, "data-runtime-fill")` — and that line's correctness
// depends entirely on how finely the CALLER happened to chunk its input:
//
//	caller passes ONE SECTION  → the marker is on that section's root, so the
//	                             answer is right by accident of framing;
//	caller passes a WHOLE PAGE → one hydrating section anywhere exempts every
//	                             unrelated section on the page.
//
// Nobody wrote the second behaviour down and no test pinned it, because at each
// individual call site the line reads as obviously correct. save_sections_link_repair.go
// already worked around it by chunking its input per section and recorded why
// ("its statically-linked neighbours no longer ride on its exemption") — the
// right conclusion, applied by call-site convention rather than by fixing the
// predicate, which leaves every other caller to rediscover it.
//
// MEASURED, live, 2026-07-31: vonc.com/index carries the marker in `lobby-grid`
// and `provocation-card`, so the assembled-page repair paths skip the WHOLE page
// — including `gauntlet-cta`'s two empty hrefs ("Enter the Gauntlet", "Find Your
// Archetype"). Those are the very controls check_dead_controls.go names in its
// own header as the case it was built for.
//
// THE BOUNDARY, and it is deliberate:
//
//	"is this CONTROL alive?"    → element scope. A control is exempt only when it
//	                              is itself inside a hydrating subtree. That is
//	                              this file.
//	"is this SECTION a shell?"  → whole-input scope, unchanged. check_empty_sections,
//	                              check_component_standards, check_component_template_corrupted
//	                              and sectionHasVisibleContent ask a genuine
//	                              whole-section question; HasRuntimeFillMarker
//	                              names it so it is a choice rather than a
//	                              coincidence.
//
// WHY BYTE SPANS AND NOT A DOM. RepairPageLinks guarantees byte-identical output
// when it changes nothing, and re-serialising through goquery would break that
// guarantee on every page it touches. So the string world gets byte spans and the
// DOM world gets InRuntimeFillShell; both derive from RuntimeFillMarker, so there
// is one marker and one meaning.

package datahelpers

import (
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// RuntimeFillMarker is the attribute that marks a section whose contents are
// populated client-side. Its placeholder hrefs are the mechanism, not a defect.
const RuntimeFillMarker = "data-runtime-fill"

// RuntimeFillSelector is the CSS form of the same marker, so the DOM half of
// this file cannot drift from the string half.
const RuntimeFillSelector = "[" + RuntimeFillMarker + "]"

// ByteSpan is a half-open byte range [Start, End) of the input HTML.
type ByteSpan struct{ Start, End int }

// Contains reports whether an offset falls inside the span.
func (s ByteSpan) Contains(off int) bool { return off >= s.Start && off < s.End }

// ByteSpanSet is a sorted, non-overlapping set of spans.
type ByteSpanSet []ByteSpan

// Contains reports whether an offset falls inside any span in the set.
func (set ByteSpanSet) Contains(off int) bool {
	for _, s := range set {
		if off < s.Start {
			return false // sorted: no later span can contain it either
		}
		if s.Contains(off) {
			return true
		}
	}
	return false
}

// HasRuntimeFillMarker is the WHOLE-INPUT test — the historical predicate, now
// named. Correct for "is this section a shell?", and wrong for any question
// about an individual control on a multi-section page. Callers asking about a
// control want RuntimeFillSpans or InRuntimeFillShell instead.
func HasRuntimeFillMarker(html string) bool {
	return strings.Contains(html, RuntimeFillMarker)
}

// InRuntimeFillShell reports whether a parsed element is, or lies inside, a
// runtime-fill shell. The DOM counterpart of RuntimeFillSpans — goquery's
// Closest tests the element itself before its ancestors, which is the semantics
// wanted here (a shell root is inside itself).
func InRuntimeFillShell(s *goquery.Selection) bool {
	return s.Closest(RuntimeFillSelector).Length() > 0
}

// voidElements never have a closing tag, so they span only their own tag.
var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true,
	"hr": true, "img": true, "input": true, "link": true, "meta": true,
	"param": true, "source": true, "track": true, "wbr": true,
}

// rawTextElements hold TEXT, not markup. Their contents must not be scanned for
// tags, or a "<" in a script string opens a phantom element and every span after
// it is wrong. This is also what stops the marker being recognised when it
// appears inside JavaScript rather than as an attribute.
var rawTextElements = map[string]bool{
	"script": true, "style": true, "textarea": true, "title": true,
}

func isTagNameByte(b byte) bool {
	return b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' ||
		b >= '0' && b <= '9' || b == '-' || b == ':' || b == '_'
}

func isASCIISpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '\f'
}

// scanTagEnd finds the '>' that closes a tag, starting from the byte after the
// tag name. Quoted attribute values are skipped whole, so `title="a > b"` does
// not end the tag early — the failure a plain IndexByte('>') would produce.
func scanTagEnd(html string, from int) (end int, selfClosing bool, ok bool) {
	i := from
	for i < len(html) {
		switch html[i] {
		case '"', '\'':
			q := html[i]
			i++
			for i < len(html) && html[i] != q {
				i++
			}
			if i >= len(html) {
				return 0, false, false // unterminated quote: refuse to guess
			}
			i++
		case '>':
			j := i - 1
			for j >= from && isASCIISpace(html[j]) {
				j--
			}
			return i + 1, j >= from && html[j] == '/', true
		default:
			i++
		}
	}
	return 0, false, false
}

// attrsCarryRuntimeFill reports whether the attribute text of a start tag
// carries the marker as an ATTRIBUTE NAME. Token-delimited on purpose: a
// substring test would match `data-runtime-filler` and `x-data-runtime-fill`,
// which are different attributes.
func attrsCarryRuntimeFill(attrs string) bool {
	for i := 0; i < len(attrs); {
		k := strings.Index(attrs[i:], RuntimeFillMarker)
		if k < 0 {
			return false
		}
		at := i + k
		after := at + len(RuntimeFillMarker)
		startsToken := at > 0 && isASCIISpace(attrs[at-1])
		endsToken := after >= len(attrs) || attrs[after] == '=' ||
			isASCIISpace(attrs[after]) || attrs[after] == '/' || attrs[after] == '>'
		if startsToken && endsToken {
			return true
		}
		i = after
	}
	return false
}

// indexCloseTag returns the offset just past `</name ...>` at or after from, or
// -1 when the element is never closed.
func indexCloseTag(html string, from int, name string) int {
	for i := from; i < len(html); {
		k := strings.Index(strings.ToLower(html[i:]), "</"+name)
		if k < 0 {
			return -1
		}
		at := i + k
		j := at + 2 + len(name)
		// `</scriptx>` is not a close tag for `script`.
		if j < len(html) && isTagNameByte(html[j]) {
			i = j
			continue
		}
		if g := strings.IndexByte(html[at:], '>'); g >= 0 {
			return at + g + 1
		}
		return -1
	}
	return -1
}

// RuntimeFillSpans returns the byte ranges of every element carrying the marker,
// including that element's whole subtree. Merged and sorted, so a caller can ask
// "is this offset inside a shell?" with one pass.
//
// UNCLOSED MARKED ELEMENTS SPAN TO END OF INPUT. That matches what a browser
// does with the same markup, and it means malformed HTML degrades to exactly
// today's whole-input exemption — never to a narrower one. A fix for
// over-exemption must not become a source of under-exemption on markup it cannot
// parse.
func RuntimeFillSpans(html string) ByteSpanSet {
	type frame struct {
		name   string
		start  int
		marked bool
	}
	var stack []frame
	var spans ByteSpanSet
	n := len(html)

	for i := 0; i < n; {
		lt := strings.IndexByte(html[i:], '<')
		if lt < 0 {
			break
		}
		i += lt
		rest := html[i:]

		if strings.HasPrefix(rest, "<!--") {
			e := strings.Index(rest, "-->")
			if e < 0 {
				break
			}
			i += e + 3
			continue
		}
		if strings.HasPrefix(rest, "<!") || strings.HasPrefix(rest, "<?") {
			g := strings.IndexByte(rest, '>')
			if g < 0 {
				break
			}
			i += g + 1
			continue
		}

		j := i + 1
		closing := false
		if j < n && html[j] == '/' {
			closing = true
			j++
		}
		ns := j
		for j < n && isTagNameByte(html[j]) {
			j++
		}
		if j == ns { // a bare "<" in text, not a tag
			i++
			continue
		}
		name := strings.ToLower(html[ns:j])

		tagEnd, selfClosing, ok := scanTagEnd(html, j)
		if !ok {
			break
		}
		attrs := html[j : tagEnd-1]

		if closing {
			for x := len(stack) - 1; x >= 0; x-- {
				if stack[x].name == name {
					if stack[x].marked {
						spans = append(spans, ByteSpan{stack[x].start, tagEnd})
					}
					stack = stack[:x]
					break
				}
			}
			i = tagEnd
			continue
		}

		marked := attrsCarryRuntimeFill(attrs)

		if selfClosing || voidElements[name] {
			if marked {
				spans = append(spans, ByteSpan{i, tagEnd})
			}
			i = tagEnd
			continue
		}

		if rawTextElements[name] {
			end := indexCloseTag(html, tagEnd, name)
			if end < 0 {
				end = n
			}
			if marked {
				spans = append(spans, ByteSpan{i, end})
			}
			i = end
			continue
		}

		stack = append(stack, frame{name: name, start: i, marked: marked})
		i = tagEnd
	}

	for _, f := range stack {
		if f.marked {
			spans = append(spans, ByteSpan{f.start, n})
		}
	}
	return mergeSpans(spans)
}

// mergeSpans sorts and coalesces overlapping ranges so Contains can stop early.
func mergeSpans(in ByteSpanSet) ByteSpanSet {
	if len(in) < 2 {
		return in
	}
	sort.Slice(in, func(a, b int) bool {
		if in[a].Start != in[b].Start {
			return in[a].Start < in[b].Start
		}
		return in[a].End < in[b].End
	})
	out := in[:1]
	for _, s := range in[1:] {
		last := &out[len(out)-1]
		if s.Start <= last.End {
			if s.End > last.End {
				last.End = s.End
			}
			continue
		}
		out = append(out, s)
	}
	return out
}

// DeadControlAnchorsOutsideRuntimeFill is DeadControlAnchors with the exemption
// applied where it belongs: an anchor is exempt only when the anchor ITSELF sits
// inside a hydrating subtree, not because something else on the page does.
//
// On a single section whose root carries the marker this returns exactly what
// the old page-wide skip returned — nothing — so the callers that were already
// chunking their input see no change.
func DeadControlAnchorsOutsideRuntimeFill(html string) []Anchor {
	spans := RuntimeFillSpans(html)
	if len(spans) == 0 {
		return DeadControlAnchors(html)
	}
	var dead []Anchor
	for _, m := range anchorRe.FindAllStringSubmatchIndex(html, -1) {
		if spans.Contains(m[0]) {
			continue
		}
		href := html[m[2]:m[3]]
		if !IsNoopHref(href) {
			continue
		}
		dead = append(dead, Anchor{Href: href, Text: normaliseAnchorText(html[m[4]:m[5]])})
	}
	return dead
}
