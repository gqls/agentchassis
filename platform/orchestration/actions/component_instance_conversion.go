// FILE: platform/orchestration/actions/component_instance_conversion.go
//
// The DETERMINISTIC half of the bugs_open/283 conversion programme (RFC_034,
// owner ruling 2026-08-17: shape C hybrid, LMC first, THROUGH THE FRAMEWORK).
//
// ConvertTemplateToInstanceScope rewrites a component template so its element
// ids are namespaced per instance: every literal id the template DECLARES is
// prefixed with {{.InstanceID}}-, and every reference to one of those ids —
// getElementById, label for=, aria id-reference attributes, CSS #id selectors
// in <style>, querySelector('#id'), href="#id" — moves with it. Ids the
// template does not declare are never touched: a reference to something
// outside the component (a chrome id, another section) is not ours to rename.
//
// GateConvertedTemplate then decides whether the result may SHIP, by rendering
// two instances through the real render layer and running the real detector.
// The gate is what enforces RFC_034 §2.1: converting the ids ALONE produces a
// page that reads clean on the id check while every button runs the last
// instance's logic (the global function name survives), so a component whose
// script still declares into global scope is REFUSED to the judged pool rather
// than shipped half-converted. Measured 2026-08-17 with the corrected detector:
// 66 of the 91 live getElementById templates pass this gate after the
// deterministic pass alone; 25 (the 23 loans-*/mortgages-* calculators plus
// two tools) are refused to the judged pool, which is the intended split.
//
// FAILURE DIRECTION, deliberately: every ambiguity REFUSES rather than
// half-converts. An id whose spelling could be a hex colour refuses the whole
// component (the #id pass would corrupt `color:#abc`); a declared id that
// still appears in a binding position after the passes (built by string
// concatenation, an unrecognised quoting) refuses; a template already carrying
// {{.InstanceID}} is a no-op. A refused component is exactly as broken as it
// was yesterday; a half-converted one is broken in a new way with the warning
// light removed.
package actions

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"go.uber.org/zap"
)

const instancePrefix = "{{.InstanceID}}-"

// reHexish — an id that could be read as a hex colour by the #id pass.
// `#abc` in a template is a CSS colour or an id selector and nothing but the
// document's ids can say which; renaming colours breaks styling silently, so a
// component declaring such an id is refused outright.
var reHexish = regexp.MustCompile(`^[0-9a-fA-F]{3,8}$`)

// Attributes whose value is an id reference (or a space-separated list of id
// references, per the ARIA spec for labelledby/describedby).
var reIDRefAttr = regexp.MustCompile(
	`(aria-labelledby|aria-describedby|aria-controls|aria-activedescendant|for|list)=(["'])([^"'{}]+)(["'])`)

// data-* attributes whose value is EXACTLY one declared id. The live corpus
// does this: tool-css-unit-converter's copy buttons carry
// data-target="result-px" and the script binds getElementById(targetId) off
// that attribute — a runtime id reference no call-site pass can see. Only an
// exact-match value is rewritten; a value merely containing an id is the
// concatenation limit, documented in the tests, and stays untouched.
var reDataAttr = regexp.MustCompile(`(data-[a-z][\w-]*)=(["'])([^"'{}]+)(["'])`)

// InstanceConversionReport says what the transform did — counts a reviewer can
// check against the diff, not adjectives.
type InstanceConversionReport struct {
	IDsDeclared      []string
	IDAttrsRenamed   int
	GetElementByID   int
	IDRefAttrs       int
	DataAttrRefs     int // data-* attrs whose value is exactly a declared id
	HashRefs         int // CSS selectors, querySelector('#…'), href="#…"
	RefusedReason    string
	AlreadyConverted bool
}

// ConvertTemplateToInstanceScope performs the deterministic rename. It returns
// the converted template, the report, and ok=false when the component REFUSES
// deterministic conversion (report.RefusedReason says why; the template is
// returned unmodified in that case).
func ConvertTemplateToInstanceScope(tpl string) (string, InstanceConversionReport, bool) {
	var rep InstanceConversionReport

	if strings.Contains(tpl, "{{.InstanceID}}") {
		rep.AlreadyConverted = true
		rep.RefusedReason = "template already references {{.InstanceID}} — conversion is not re-runnable over itself by design; nothing to do"
		return tpl, rep, false
	}

	// The declared-id set: only ids carried by an id= attribute, the same
	// predicate the detector uses. Both quote styles — JS that builds markup
	// inside a single-quoted string writes id="x", but templates vary.
	seen := map[string]bool{}
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`\sid="([^"{}]+)"`),
		regexp.MustCompile(`\sid='([^'{}]+)'`),
	} {
		for _, m := range re.FindAllStringSubmatch(tpl, -1) {
			seen[m[1]] = true
		}
	}
	if len(seen) == 0 {
		rep.RefusedReason = "template declares no literal element ids — nothing to namespace"
		return tpl, rep, false
	}
	for id := range seen {
		rep.IDsDeclared = append(rep.IDsDeclared, id)
		if reHexish.MatchString(id) {
			rep.RefusedReason = fmt.Sprintf(
				"id %q is indistinguishable from a hex colour in a #-selector — the CSS pass would corrupt colours; convert this component through the judged pool", id)
			return tpl, rep, false
		}
	}
	sort.Strings(rep.IDsDeclared)

	out := tpl

	// Pass 1 — the declarations themselves, both quote styles. Plain-text
	// anchored replacement, so ids inside JS strings that BUILD markup (e.g.
	// parts.push('<div id="x">')) are renamed consistently with the markup.
	for id := range seen {
		n := strings.Count(out, `id="`+id+`"`) + strings.Count(out, `id='`+id+`'`)
		out = strings.ReplaceAll(out, `id="`+id+`"`, `id="`+instancePrefix+id+`"`)
		out = strings.ReplaceAll(out, `id='`+id+`'`, `id='`+instancePrefix+id+`'`)
		rep.IDAttrsRenamed += n
	}

	// Pass 2 — getElementById with either quote style.
	for id := range seen {
		for _, q := range []string{`'`, `"`} {
			from := "getElementById(" + q + id + q + ")"
			to := "getElementById(" + q + instancePrefix + id + q + ")"
			rep.GetElementByID += strings.Count(out, from)
			out = strings.ReplaceAll(out, from, to)
		}
	}

	// Pass 3 — id-reference attributes, including ARIA's space-separated lists.
	out = reIDRefAttr.ReplaceAllStringFunc(out, func(m string) string {
		g := reIDRefAttr.FindStringSubmatch(m)
		toks := strings.Fields(g[3])
		changed := false
		for i, t := range toks {
			if seen[t] {
				toks[i] = instancePrefix + t
				changed = true
			}
		}
		if !changed {
			return m
		}
		rep.IDRefAttrs++
		return g[1] + "=" + g[2] + strings.Join(toks, " ") + g[4]
	})

	// Pass 3b — data-* attributes whose value is exactly a declared id. The
	// script side of this reference (getElementById(someVar)) is invisible to
	// any textual pass, so the ATTRIBUTE side must move or the runtime lookup
	// dangles — found by the tool-css-unit-converter fixture, whose five copy
	// buttons bind exactly this way.
	out = reDataAttr.ReplaceAllStringFunc(out, func(m string) string {
		g := reDataAttr.FindStringSubmatch(m)
		if !seen[g[3]] {
			return m
		}
		rep.DataAttrRefs++
		return g[1] + "=" + g[2] + instancePrefix + g[3] + g[4]
	})

	// Pass 4 — #id references: CSS selectors inside <style>, querySelector
	// strings, href="#…" anchors. Longest id first so an id that prefixes
	// another (rate, rateYears) cannot be clobbered through it; the boundary
	// lookahead does the same job, belt and braces.
	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return len(ids[i]) > len(ids[j]) })
	for _, id := range ids {
		re := regexp.MustCompile(`#` + regexp.QuoteMeta(id) + `([^A-Za-z0-9_-]|$)`)
		out = re.ReplaceAllStringFunc(out, func(m string) string {
			rep.HashRefs++
			return "#" + instancePrefix + id + m[len("#"+id):]
		})
	}

	// COMPLETENESS — the check that turns "the passes I thought of" into "the
	// bindings that exist". Any declared id still reachable in a binding
	// position after the passes was built in a way the passes do not
	// recognise (concatenation, a template literal, an unquoted attribute);
	// shipping would leave that one binding pointing at nothing, silently.
	for id := range seen {
		leftovers := []string{
			`id="` + id + `"`, `id='` + id + `'`,
			`getElementById('` + id + `')`, `getElementById("` + id + `")`,
		}
		for _, l := range leftovers {
			if strings.Contains(out, l) {
				rep.RefusedReason = fmt.Sprintf("binding %q survived every pass — unrecognised construction; convert through the judged pool", l)
				return tpl, rep, false
			}
		}
		if re := regexp.MustCompile(`#` + regexp.QuoteMeta(id) + `([^A-Za-z0-9_-]|$)`); re.MatchString(out) {
			rep.RefusedReason = fmt.Sprintf("#%s reference survived every pass — unrecognised construction; convert through the judged pool", id)
			return tpl, rep, false
		}
	}

	return out, rep, true
}

// GateConvertedTemplate decides whether a converted template may SHIP. It is
// the acceptance gate the owner's ruling names: render two instances through
// the REAL render layer with the REAL token derivation, then ask the REAL
// detector. Returning needsJudgedPool=true means the deterministic pass is not
// enough for this component (its script genuinely declares into global scope,
// or assigns window.onload) — RFC_034 §2.1 forbids shipping that state, so the
// caller must route it to the judged pipeline, not write it.
func GateConvertedTemplate(function, converted string, logger *zap.Logger) (needsJudgedPool bool, err error) {
	if !strings.Contains(converted, "{{.InstanceID}}") {
		return false, fmt.Errorf("gate: converted template contains no {{.InstanceID}} — the transform did nothing")
	}

	toks := InstanceTokensForPage([]string{function, function})
	var page strings.Builder
	for _, tok := range toks {
		rc := &RenderContext{}
		BindInstanceToken(rc, tok)
		rendered := RenderTemplate(converted, rc, logger)
		if strings.Contains(rendered, "{{.InstanceID}}") {
			return false, fmt.Errorf("gate: {{.InstanceID}} survived rendering — mangled placeholder")
		}
		if strings.Contains(rendered, `id="-`) || strings.Contains(rendered, `id='-`) {
			return false, fmt.Errorf("gate: an id rendered with an EMPTY token — the missingkey failure this seam exists to remove")
		}
		page.WriteString(rendered)
	}

	report := DetectInstanceCollisions(page.String())
	if n := len(report.DuplicateElementIDs); n > 0 {
		// Ids that still collide across two DIFFERENT tokens were not actually
		// namespaced — a transform defect, never a judged-pool case.
		return false, fmt.Errorf("gate: %d id(s) still duplicated across two instances (%s) — transform incomplete",
			n, strings.Join(report.DuplicateElementIDs, ", "))
	}
	if report.UnscopedInlineScripts > 0 || report.WindowOnloadAssignments > 1 {
		// The §2.1 refusal: ids are clean, the script half is not. Shipping
		// now would remove the only visible signal while both buttons still
		// run the last instance's logic.
		return true, nil
	}
	return false, nil
}
