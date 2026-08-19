// FILE: platform/orchestration/actions/component_instance_judged.go
//
// The JUDGED half of bugs_open/283's conversion programme (RFC_034 shape C,
// owner ruling 2026-08-17; design PLAN_2026-08-18_judged_pipeline.md).
//
// 25 components convert cleanly on ids and still declare into global scope —
// `function runCalc()` at top level, inline `onclick="runCalc()"`, one
// `window.onload` slot. Namespacing their ids alone produces a page that reads
// clean and still cross-talks (RFC_034 §2.1), and the IIFE route is forced
// because {{.InstanceID}} is not a valid JS identifier (§2.2). So for these the
// script is rewritten by an LLM, and everything in this file exists to make
// that rewrite CHECKABLE rather than trusted:
//
//   - the LLM receives the IDS-CONVERTED template (the deterministic pass has
//     already moved every id, getElementById, label-for, CSS #id and data-*
//     reference — the surfaces it is proven on) and a brief narrowed to the
//     script: wrap, rewire the inventoried on*= handlers, replace window.onload,
//     change nothing else;
//   - JudgedConversionIssues then refuses anything outside that brief: the two-
//     instance gate must be FULLY clean, the markup outside <script> bodies must
//     equal the baseline's with only on*= attributes removed, and the declared
//     id set must be identical. A cut generation loses ids and unbalances
//     markup; an "improving" LLM changes markup — both refuse.
//
// Every helper is pure so the thresholds stay testable against live row bytes.
package actions

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"go.uber.org/zap"
)

// reInlineHandlerAttr matches one inline event-handler attribute, with the
// whitespace that precedes it so removal leaves no double space behind.
// Either quote style; the value may span lines.
var reInlineHandlerAttr = regexp.MustCompile(`\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*')`)

// reScriptBody matches the BODY of an inline <script> element, leaving the
// tags in place: `<script src=…></script>` is markup and must survive the
// comparison, while the inline program is exactly what the LLM is licensed to
// change. (?s) so the body may span lines; non-greedy so sibling scripts do
// not merge.
var reScriptBody = regexp.MustCompile(`(?s)(<script\b[^>]*>).*?(</script>)`)

// reMarkdownFence — an LLM told "no fences" still sometimes opens one. The
// fence is not part of the template; strip a leading ```html / ``` line and
// a trailing ``` so the structural checks judge the template, not the wrapper.
var reMarkdownFence = regexp.MustCompile("(?s)^\\s*```[a-zA-Z]*\\s*\\n(.*?)\\n?```\\s*$")

// InlineHandlerInventory lists every inline on*= handler attribute in the
// template, verbatim, in document order. It is handed to the LLM so the brief
// names the exact handlers to rewire — and carried on the result so a reviewer
// can count them against the rewrite.
func InlineHandlerInventory(tpl string) []string {
	var out []string
	for _, m := range reInlineHandlerAttr.FindAllString(tpl, -1) {
		out = append(out, strings.TrimSpace(m))
	}
	return out
}

// WindowOnloadCount reports how many `window.onload =` assignments the template
// makes. One slot per page; the second instance's assignment silently replaces
// the first's (RFC_034 §3).
func WindowOnloadCount(tpl string) int {
	return len(reWindowOnload.FindAllString(tpl, -1))
}

// StripMarkdownFence removes a surrounding ```…``` wrapper if, and only if, the
// whole text is one fenced block. Anything else is returned unchanged.
func StripMarkdownFence(s string) string {
	if m := reMarkdownFence.FindStringSubmatch(s); m != nil {
		return m[1]
	}
	return s
}

// markupOutsideScripts blanks every inline <script> BODY, keeping the tags, so
// two templates can be compared on the part the judged rewrite must not touch.
func markupOutsideScripts(tpl string) string {
	return reScriptBody.ReplaceAllString(tpl, "$1$2")
}

// removeInlineHandlers drops every on*= attribute. Applied to the BASELINE so
// that the expected markup after the rewrite is derived, not asserted.
func removeInlineHandlers(markup string) string {
	return reInlineHandlerAttr.ReplaceAllString(markup, "")
}

// normaliseMarkup collapses whitespace so cosmetic differences — an attribute
// removed leaving one space, a re-wrapped line — do not read as edits, while
// any change to a tag, an attribute value or text still does.
func normaliseMarkup(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, " >", ">")
	s = strings.ReplaceAll(s, "< ", "<")
	s = strings.ReplaceAll(s, "> <", "><")
	return strings.TrimSpace(s)
}

// declaredIDs returns the sorted set of id= attribute values — the same
// predicate the converter and the detector use.
func declaredIDs(tpl string) []string {
	seen := map[string]bool{}
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`\sid="([^"]+)"`),
		regexp.MustCompile(`\sid='([^']+)'`),
	} {
		for _, m := range re.FindAllStringSubmatch(tpl, -1) {
			seen[m[1]] = true
		}
	}
	out := make([]string, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// JudgedConversionIssues is the acceptance gate for an LLM-rewritten template.
// `baseline` is the IDS-CONVERTED template the LLM was given (re-derived by the
// caller from the live row, never trusted from the workflow); `candidate` is
// the rewrite. An empty result means the rewrite may be written. Every issue
// names the check that failed, because a refusal here goes to a human.
//
// Order matters only for the message a human reads first: the two-instance
// gate is the property the whole programme exists for, so it leads.
func JudgedConversionIssues(function, baseline, candidate string, logger *zap.Logger) []string {
	var issues []string

	if strings.TrimSpace(candidate) == "" {
		return []string{"rewrite is empty"}
	}

	// 1. The programme's own gate, and it must be FULLY clean: zero duplicate
	//    ids across two rendered instances, zero unrendered or empty tokens,
	//    zero unscoped inline scripts, at most one window.onload. A judged
	//    rewrite that still needs the judged pool has not done its one job.
	needsJudged, err := GateConvertedTemplate(function, candidate, logger)
	if err != nil {
		issues = append(issues, fmt.Sprintf("two-instance gate: %v", err))
	} else if needsJudged {
		issues = append(issues, "two-instance gate: script still declares into global scope (or assigns window.onload more than once) — the rewrite did not scope the script")
	}

	// 2. Markup parity outside the script bodies. The expected markup is the
	//    baseline's with inline handlers removed — derived, so a rewrite that
	//    removed them correctly matches and one that edited anything else does
	//    not. Cosmetic reformatting can refuse here; that is the chosen failure
	//    direction (refuse to a human, never write), relaxed only on evidence.
	wantMarkup := normaliseMarkup(removeInlineHandlers(markupOutsideScripts(baseline)))
	gotMarkup := normaliseMarkup(markupOutsideScripts(candidate))
	if wantMarkup != gotMarkup {
		issues = append(issues, fmt.Sprintf(
			"markup parity: the rewrite changed markup outside the <script> bodies (expected %d normalised chars, got %d; first divergence at offset %d) — the brief was script-only",
			len(wantMarkup), len(gotMarkup), firstDivergence(wantMarkup, gotMarkup)))
	}

	// 3. Id-set parity. A cut generation loses ids; an LLM that renames one
	//    breaks the binding the deterministic pass already moved. Either way
	//    the set must be identical.
	wantIDs, gotIDs := declaredIDs(baseline), declaredIDs(candidate)
	if strings.Join(wantIDs, "\x00") != strings.Join(gotIDs, "\x00") {
		issues = append(issues, fmt.Sprintf(
			"id-set parity: baseline declares %d ids, rewrite declares %d (missing: %v; added: %v)",
			len(wantIDs), len(gotIDs), setDiff(wantIDs, gotIDs), setDiff(gotIDs, wantIDs)))
	}

	// 4. Every binding must carry the prefix (component_instance_bindings.go):
	//    the LLM was told to prefix dynamic lookups with a `ns` const; if any
	//    bare literal or concatenated prefix survives, two instances read
	//    clean and both dangle at runtime.
	if ub := UnprefixedBindings(candidate); len(ub) > 0 {
		issues = append(issues, "unprefixed bindings: "+strings.Join(ub, "; "))
	}

	// 5. No inline handler may survive — the gate above catches a surviving
	//    handler only if it resolves to a global; a handler left pointing at a
	//    now-IIFE-private function throws at click time and the gate is blind
	//    to it. Count them directly.
	if left := InlineHandlerInventory(candidate); len(left) > 0 {
		issues = append(issues, fmt.Sprintf(
			"inline handlers: %d on*= attribute(s) survive the rewrite (%s) — inside an IIFE they would resolve to nothing at click time",
			len(left), strings.Join(left, "; ")))
	}

	return issues
}

func firstDivergence(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return n
}

func setDiff(a, b []string) []string {
	in := map[string]bool{}
	for _, s := range b {
		in[s] = true
	}
	var out []string
	for _, s := range a {
		if !in[s] {
			out = append(out, s)
		}
	}
	return out
}
