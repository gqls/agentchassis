// FILE: platform/orchestration/actions/refresh_evidence_fact_probe.go
//
// bugs_open/288 Phase 3a — DOES THE TOOL ACTUALLY CARRY THE REGISTERED FIGURE?
//
// CLM-022 answers "did the register's figure MOVE, and who declares they encode
// it". It has never looked at what a tool CONTAINS. On the one tool that has
// adopted it, the mechanism filed 13 `unreconciled_declaration` items on
// 2026-08-17 asking a human to confirm the tool encodes the registered values;
// seven days later all 13 were still `needs_human_review`, untouched, in the
// queue bugs_open/033 says has no working surface. This probe is the machine
// half of that question.
//
// ── ANNOTATION ONLY, AND THAT IS THE POINT ─────────────────────────────────
//
// Nothing here changes a route, a band or a status. It records what it saw on
// each emission so that ONE full fleet sweep produces the measured
// present/absent/markup-only distribution, and only then does a separate round
// decide whether presence may settle an item. PLAN_2026-08-09 §3 refused a
// static scan of tool JavaScript precisely because its false-positive rate was
// unmeasured; measuring before arming is the same discipline, not a way round it.
//
// ── SCRIPT TEXT, NEVER THE WHOLE PAGE. THIS IS THE LOAD-BEARING RULE ───────
//
// The evidence register's own `writer_line`/`writer_block` machinery puts the
// registered figure into the page's PROSE. Measured on the live stamp-duty page
// 2026-08-24 (15,111 bytes: 6,132 script / 8,962 non-script):
//
//	`500,000` (comma form) — present in the script AND in the prose
//	`500000`  (raw form)   — present in the script only
//	`625000`  (bug 225's expired cap) — absent from both
//
// So a whole-page probe matching the comma form finds the figure in the COPY
// whatever the JavaScript computes — and bugs_closed/225 was exactly "correct
// register, correct copy, stale code". A mechanism built to catch 225 would have
// certified 225's page every day for sixteen months. `extractScriptText` is what
// stops that, and the induced-red test for it is the whole of this file's claim
// to work.
//
// ── AND RAW DIGITS, NOT THE COMMA FORM, FOR THE SAME REASON ONE LEVEL DOWN ─
//
// Inside the script the comma form is almost always a comment or a string
// literal, i.e. copy again. Measured on the same page: `300,000`, `250,000`,
// `925,000`, `125,000` appear ONLY inside JS comments and quoted strings
// ("…no stamp duty up to £300,000"), while the raw forms are the band table the
// tool actually computes from — `{ upTo: 300000, rate: 0.00 }`. Matching raw
// digits found all seven real code constants and none of the prose.
//
// ── THE DISTINCTIVENESS FLOOR IS MEASURED, NOT CHOSEN ──────────────────────
//
// A short value matches anything. Measured 2026-08-24 over the script text of
// all 161 tool pages that have any, using INVENTED values so that every match is
// a false positive by construction — 40 probes per digit length, 6,440
// page-probe pairs each:
//
//	digits   1      2      3      4      5+
//	FP rate  32.75% 3.79%  0.06%  0.03%  0.00%
//
// Hence factProbeMinValue = 1000. Below it the probe REFUSES rather than
// guessing, which is what keeps the 110 of 294 current facts whose value is
// under 100 (percentage rates: 5, 2, 10, 12) from generating noise. A fact that
// needs a smaller figure proven has an escape hatch that is strictly better than
// a generated pattern: a human-authored `artifact_check` with real context, which
// takes precedence over this probe.
//
// ── WHAT ABSENCE PROVES, WHICH IS LESS THAN IT LOOKS ───────────────────────
//
// Absence is consistent with bug 225 AND with four benign causes: a value the
// tool DERIVES rather than states (250000*2), an exotic formatting, a figure
// carried in markup rather than code, and over-declaration (a fence naming a
// fact the tool does not encode). The guard also produces real false negatives.
// So absence is EVIDENCE FOR A HUMAN, never an instruction to a fixer, and this
// file never routes anything.

package actions

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// factProbeMinValue is the measured distinctiveness floor — see the file header.
// A registered value below this is not probed at all.
const factProbeMinValue = 1000

// Probe outcomes. Deliberately five, not two: "I could not look" and "it is in
// the copy but not the code" are different findings from "it is not there", and
// collapsing them is how a check starts lying about its own coverage.
const (
	factProbePresentInScript = "present_in_script"      // the registered figure is in the tool's code
	factProbeMarkupOnly      = "present_in_markup_only" // in the page, but NOT in its script — 225's shape
	factProbeAbsent          = "absent"                 // nowhere on the page
	factProbeNoSurface       = "no_surface"             // no stored HTML to read
	factProbeNotProbed       = "not_probed"             // refused: no value, or below the measured floor
)

// factProbeResult is what one (fact, tool) probe saw. Detail is written into the
// work item body so a human reads what was looked for and where, rather than a
// bare verdict.
type factProbeResult struct {
	Outcome string
	Form    string // the literal form that matched, when one did
	Detail  string
}

// factProbeSurface is a page's stored bytes, kept in the two shapes the probe
// needs — and built PER COMPONENT, which is the whole point.
//
// ⚠ NEVER TOKENIZE A string_agg OF page_components.rendered_html. Each row is a
// PARTIAL HTML FRAGMENT, not a standalone document, so an unbalanced or oddly
// self-closed <script> in one component leaves the tokenizer's inScript flag SET
// when it crosses into the next component — and that component's genuine PROSE
// is then collected as "script text". The probe would report
// present_in_script for a figure that is only in the copy, which is bug 225's
// false certification exactly, one level down from the page-vs-script mistake
// this file was written to avoid. It is the estate's documented multi-component
// string_agg landmine ("stripping style/script AFTER string_agg lets one
// component's block eat the NEXT component's prose").
//
// So: tokenize each fragment on its own, and concatenate the EXTRACTED SCRIPT
// TEXT rather than the raw HTML. Tokenizer state cannot cross a boundary that
// the tokenizer never sees. Caught by the council's debug_historian seat at
// severity high (corr 041b3026) — every fixture in the first cut was a single
// synthetic page, so the failure mode was untested.
//
// RawHTML stays a plain concatenation on purpose: the markup arm is a literal
// search with no state to leak, and it must see the whole page.
type factProbeSurface struct {
	ScriptText string
	RawHTML    string
}

// buildFactProbeSurface extracts per fragment and joins the results.
func buildFactProbeSurface(fragments []string) factProbeSurface {
	var scripts, raw strings.Builder
	for _, f := range fragments {
		if strings.TrimSpace(f) == "" {
			continue
		}
		if s := extractScriptText(f); strings.TrimSpace(s) != "" {
			scripts.WriteString(s)
			scripts.WriteByte('\n')
		}
		raw.WriteString(f)
		raw.WriteByte('\n')
	}
	return factProbeSurface{ScriptText: scripts.String(), RawHTML: raw.String()}
}

// extractScriptText returns the concatenated text of every <script> element.
//
// IN THIS PACKAGE, NOT IN datahelpers, deliberately. datahelpers/claims.go is the
// shared claims surface and it has an active concurrent editor (four commits
// 2026-08-19 → 2026-08-22 adding a regulated-claims family); this change must not
// put a same-file passenger in their way. It is also the exact INVERSE of what
// that file does — claims.go drops script subtrees via nonAssertionElements
// because a figure in a script is not an assertion about the business, and it is
// right. This wants only the scripts, because the question is not "what does the
// page claim" but "what does the code compute from".
// ⚠ The x/net/html `TagName()` lower-cases tag bytes IN PLACE in the buffer
// `Raw()` aliases, so a byte-preserving rewriter must Write(Raw()) BEFORE
// calling TagName() (LANDMINES; rendered_html_code_spans.go is the estate's
// worked case). It does NOT bite here, and the reason is worth stating rather
// than leaving to luck: this function never calls Raw() and promises nothing
// about byte-verbatim output — it reads Text() only, into a Builder that copies
// at call time. The in-place lower-casing is in fact HELPFUL here, because it
// makes `<SCRIPT>` compare equal to "script"; there is a mixed-case test for
// exactly that, so the behaviour is pinned rather than assumed.
func extractScriptText(pageHTML string) string {
	if strings.TrimSpace(pageHTML) == "" {
		return ""
	}
	var b strings.Builder
	z := html.NewTokenizer(strings.NewReader(pageHTML))
	inScript := false
	for {
		switch z.Next() {
		case html.ErrorToken:
			return b.String()
		case html.StartTagToken:
			if name, _ := z.TagName(); string(name) == "script" {
				inScript = true
			}
		case html.EndTagToken:
			if name, _ := z.TagName(); string(name) == "script" {
				inScript = false
			}
		case html.TextToken:
			if inScript {
				b.Write(z.Text())
				b.WriteByte('\n')
			}
		}
	}
}

// factValueLiterals returns the literal renderings of a registered value that a
// tool's CODE plausibly contains, or ok=false when the value must not be probed.
//
// Raw digits and the underscore form only — never the comma form, which inside a
// script is a comment or a string (see the file header's measurement). An
// integral value is rendered without a decimal point because that is how a JS
// constant is written; a non-integral one keeps its shortest exact form.
func factValueLiterals(v float64) (literals []string, ok bool) {
	if v < factProbeMinValue {
		return nil, false
	}
	if v == float64(int64(v)) {
		n := int64(v)
		raw := strconv.FormatInt(n, 10)
		lits := []string{raw}
		if u := underscoreGroups(raw); u != raw {
			lits = append(lits, u)
		}
		return lits, true
	}
	return []string{strconv.FormatFloat(v, 'f', -1, 64)}, true
}

// factValueDisplayLiterals returns the HUMAN-READABLE renderings — the comma
// form, plus the code forms — used ONLY for the markup-only arm.
//
// Two form sets, one per surface, and the asymmetry is the design rather than an
// oversight. In CODE the comma form is not a number at all (`500,000` is two
// arguments), and where it does appear inside a script it is a comment or a
// string — measured on the live stamp-duty page, `300,000` / `250,000` /
// `925,000` / `125,000` occur ONLY in JS comments and quoted strings, while the
// band table the tool computes from is raw. In COPY the opposite holds: prose
// writes `£500,000`, which is exactly what the register's own writer_line
// produces. Probing the script for code forms and the page for display forms is
// what lets "the copy is current, the code is not" — bugs_closed/225's shape —
// be stated as a finding instead of vanishing into a bare absence.
func factValueDisplayLiterals(v float64) []string {
	lits, ok := factValueLiterals(v)
	if !ok {
		return nil
	}
	if v == float64(int64(v)) {
		n := int64(v)
		if c := commaGroups(strconv.FormatInt(n, 10)); c != "" {
			lits = append(lits, c)
		}
	}
	return lits
}

// commaGroups renders 500000 as 500,000 — the way the register's writer_line
// and every piece of published copy writes a currency figure.
func commaGroups(digits string) string {
	if len(digits) <= 3 {
		return ""
	}
	var out []byte
	for i, c := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, c)
	}
	return string(out)
}

// underscoreGroups renders 1500000 as 1_500_000 — legal JS, and used in hand
// written constants often enough to be worth one extra literal.
func underscoreGroups(digits string) string {
	if len(digits) <= 3 {
		return digits
	}
	var out []byte
	for i, c := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, '_')
		}
		out = append(out, c)
	}
	return string(out)
}

// valueOccursGuarded reports whether literal occurs in text as a NUMBER IN ITS
// OWN RIGHT, not as part of a longer number or an identifier.
//
// Hand-rolled rather than a regexp because Go's RE2 has NO LOOKAROUND, and the
// rule needs it in both directions: `1500000` must match inside
// `{ upTo: 1500000, rate: 0.10 }` (a trailing comma is a list separator) but must
// NOT match inside `21500000` or `1500000.5` or `x1500000` or `1,500,000` (where
// a comma sits BETWEEN digits). A regexp without lookaround either loses the
// list-separator case or admits the thousands-separator one — the first version of
// this rule excluded every trailing comma and so failed to see the real band
// table on the very page it was written for.
func valueOccursGuarded(text, literal string) bool {
	if literal == "" {
		return false
	}
	from := 0
	for {
		i := strings.Index(text[from:], literal)
		if i < 0 {
			return false
		}
		i += from
		end := i + len(literal)
		if boundaryOK(text, i, end) {
			return true
		}
		from = i + 1
	}
}

func boundaryOK(text string, start, end int) bool {
	if start > 0 {
		p := text[start-1]
		if isNumIdentByte(p) {
			return false
		}
		// A separator only if a digit precedes it: "1,500000" / "0.500000".
		if (p == ',' || p == '.') && start-2 >= 0 && text[start-2] >= '0' && text[start-2] <= '9' {
			return false
		}
	}
	if end < len(text) {
		n := text[end]
		if isNumIdentByte(n) {
			return false
		}
		// A separator only if a digit follows it: "500000,000" / "500000.5".
		if (n == ',' || n == '.') && end+1 < len(text) && text[end+1] >= '0' && text[end+1] <= '9' {
			return false
		}
	}
	return true
}

func isNumIdentByte(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

// probeFactValueOnSurface is the whole probe: given a page's stored HTML and a
// registered value, did the tool's CODE carry it?
//
// The markup-only outcome is not a nicety. It is the single most useful thing
// this probe can say, because it is bug 225's exact shape: the copy states the
// current figure (the register put it there) while the code does not.
func probeFactValueOnSurface(surface factProbeSurface, value float64, hasValue bool) factProbeResult {
	if strings.TrimSpace(surface.RawHTML) == "" {
		return factProbeResult{Outcome: factProbeNoSurface,
			Detail: "no stored component HTML for this tool's page — nothing was read, so nothing is claimed"}
	}
	if !hasValue {
		return factProbeResult{Outcome: factProbeNotProbed,
			Detail: "the register carries no numeric value for this fact, so there is no literal to look for"}
	}
	lits, ok := factValueLiterals(value)
	if !ok {
		return factProbeResult{Outcome: factProbeNotProbed,
			Detail: fmt.Sprintf(
				"registered value %s is below the probe's measured distinctiveness floor of %d — a figure this "+
					"short matches unrelated code (measured 2026-08-24 over 161 tool pages: %.2f%% false positives at "+
					"two digits, %.2f%% at one). Attach a human-authored artifact_check with real surrounding context "+
					"to prove a figure this small; it takes precedence over this probe.",
				formatEvidenceNumber(value), factProbeMinValue, 3.79, 32.75)}
	}

	scripts := surface.ScriptText
	for _, lit := range lits {
		if valueOccursGuarded(scripts, lit) {
			return factProbeResult{Outcome: factProbePresentInScript, Form: lit,
				Detail: fmt.Sprintf("the registered figure is present in the tool's script as %q", lit)}
		}
	}
	// Not in the code. Is it in the page at all? If it is, the copy is current and
	// the code is not — which is the motivating bug, and is worth saying out loud
	// rather than reporting a bare absence.
	for _, lit := range factValueDisplayLiterals(value) {
		if valueOccursGuarded(surface.RawHTML, lit) {
			return factProbeResult{Outcome: factProbeMarkupOnly, Form: lit,
				Detail: fmt.Sprintf(
					"the registered figure appears in the page as %q but NOT in its script — the COPY carries the "+
						"current figure while the code does not. This is bugs_closed/225's exact shape (correct "+
						"register, correct copy, stale code) and is the case a whole-page check cannot see, because "+
						"the register's own writer_line puts the figure in the prose.", lit)}
		}
	}
	return factProbeResult{Outcome: factProbeAbsent,
		Detail: fmt.Sprintf(
			"none of the registered figure's literal forms (%s) occurs in this tool's page as a number in its own "+
				"right. That is consistent with a stale calculator (bugs_closed/225) AND with four benign causes: a "+
				"value the tool DERIVES rather than states, an unusual formatting, a figure carried in markup this "+
				"probe cannot attribute, or a fence declaring a fact the tool does not encode. A human decides which; "+
				"nothing is routed at a fixer on this evidence.",
			strings.Join(lits, ", "))}
}
