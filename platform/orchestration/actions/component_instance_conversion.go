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

// reComponentIDInIDAttr — an id= attribute whose whole value is the OLD
// per-instance placeholder. Deliberately tolerant of Go template whitespace and
// trim markers ({{- .ComponentID -}}), because the corpus is LLM-written and a
// spelling this misses would be silently left behind rather than loudly
// refused. Group 1 is the leading whitespace, group 2 the quote character, so
// the rewrite preserves both.
var reComponentIDInIDAttr = regexp.MustCompile(`(\s)id=(["'])\{\{-?\s*\.ComponentID\s*-?\}\}(["'])`)

// reComponentIDAnywhere — the same placeholder in ANY position, for the
// completeness refusal. A reference that is not an id attribute (a data-*
// value, a #selector, a script literal) is not something pass 0 can safely
// rewrite, and leaving it while converting the rest is the half-state this
// file's failure direction forbids.
var reComponentIDAnywhere = regexp.MustCompile(`\{\{-?\s*\.ComponentID\b`)

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
	// Binding passes (2026-08-19, component_instance_bindings.go): literals
	// and concatenation sites that carry an id to a lookup without containing
	// the lookup — the class the original passes could not see.
	Bindings BindingPassReport
	// TemplatedIDSwaps counts id= attributes whose value was the OLD
	// per-instance placeholder {{.ComponentID}} and is now {{.InstanceID}}
	// (pass 0). Reported separately from IDAttrsRenamed because it is a
	// different act: those are literal ids gaining a namespace, this is one
	// convention being replaced by the one that works (RFC_032 §8).
	TemplatedIDSwaps int
	// UnprefixedBindings is the completeness detector's report over the
	// OUTPUT. Non-empty does not refuse here — the gate turns it into a
	// judged-pool route, and the caller carries it on the result.
	UnprefixedBindings []string
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

	// PASS 0 — an id already written as the OLD per-instance placeholder.
	//
	// {{.ComponentID}} is the estate's previous convention (RFC_032, RULED
	// 2026-08-22: converge on {{.InstanceID}} and retire it). It resolves to
	// the component ROW id on the two live render paths, so every instance of
	// the component takes the SAME value — the collision this seam exists to
	// remove, wearing a templated id's clothes.
	//
	// It runs before the declared-id harvest below, and it has to, because
	// that harvest CANNOT SEE IT: `[^"{}]` excludes braces, so a template
	// whose only id is `id="{{.ComponentID}}"` produced an empty `seen` and
	// refused with "declares no literal element ids" — the exact opposite of
	// the truth, on a template that plainly declares one. Five live templates
	// were in that state as of 2026-08-22 and would have drained through the
	// conversion queue as five polite no-ops (LANDMINES, "A converter that
	// harvests ids with id="([^"{}]+)" cannot SEE a templated id").
	//
	// The wrapper takes the BARE token, not the {{.InstanceID}}- prefix: this
	// id IS the instance's identity rather than an id scoped WITHIN it, and
	// the prefix form would render `c-faq-` with nothing after the hyphen.
	swapped := reComponentIDInIDAttr.ReplaceAllStringFunc(tpl, func(m string) string {
		rep.TemplatedIDSwaps++
		g := reComponentIDInIDAttr.FindStringSubmatch(m)
		return g[1] + "id=" + g[2] + "{{.InstanceID}}" + g[2]
	})

	// The declared-id set: only ids carried by an id= attribute, the same
	// predicate the detector uses. Both quote styles — JS that builds markup
	// inside a single-quoted string writes id="x", but templates vary.
	//
	// Harvested from `swapped`, not `tpl`: pass 0's output carries
	// {{.InstanceID}}, which this predicate also cannot see, so the swapped
	// wrapper is neither double-prefixed here nor renamed by the passes below.
	seen := map[string]bool{}
	for _, re := range []*regexp.Regexp{
		regexp.MustCompile(`\sid="([^"{}]+)"`),
		regexp.MustCompile(`\sid='([^'{}]+)'`),
	} {
		for _, m := range re.FindAllStringSubmatch(swapped, -1) {
			seen[m[1]] = true
		}
	}
	if len(seen) == 0 && rep.TemplatedIDSwaps == 0 {
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

	out := swapped

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

	// Pass 5 — bindings that do not CONTAIN the lookup (2026-08-19, after the
	// mechanical batch shipped 32 templates with dangling bindings): bare
	// literals equal to a declared id (fed to a helper or through an array),
	// and concatenated declarations/lookups on a declared id's `-`/`_` prefix.
	// A dynamically-declared id with no static prefix cannot be carried and
	// refuses to the judged pool.
	baseIDs, fragments, rejects := BindingIDSets(out)
	if len(rejects) > 0 {
		rep.RefusedReason = fmt.Sprintf("dynamic id declaration %q has no static `-`/`_` prefix to carry the token — convert through the judged pool", rejects[0])
		return tpl, rep, false
	}
	out, rep.Bindings = ApplyBindingPasses(out, baseIDs, fragments)

	// COMPLETENESS — the check that turns "the passes I thought of" into "the
	// bindings that exist". Any declared id still reachable in a binding
	// position after the passes was built in a way the passes do not
	// recognise (concatenation, a template literal, an unquoted attribute);
	// shipping would leave that one binding pointing at nothing, silently.
	//
	// ⚠ This literal-form check was the ONLY completeness check until
	// 2026-08-19, and its premise — that a surviving binding contains the id
	// literal at the binding site — is false for the three constructions pass
	// 5 now handles. UnprefixedBindings below is the check that actually
	// closes the class; this loop stays as the cheap first line.
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

	// COMPLETENESS, the {{.ComponentID}} half. Pass 0 rewrites the placeholder
	// only where it IS an id attribute's whole value. A reference anywhere
	// else — a data-target, a #selector, a script literal, an id built by
	// concatenation around it — survives, and it survives INVISIBLY: the gate
	// renders with InstanceID bound and nothing else, so a leftover
	// {{.ComponentID}} resolves through missingkey=zero to "" and is stripped,
	// leaving id="" on every instance. DetectInstanceCollisions cannot report
	// that, because reElementID requires at least one non-brace character. So
	// two instances would read CLEAN at the gate while serving colliding empty
	// ids — a half-conversion with the warning light removed, which is exactly
	// what this file's stated failure direction exists to prevent.
	//
	// Measured 2026-08-22: 0 active templates mix a literal id with
	// {{.ComponentID}} (control: 87 active templates carry a literal id at
	// all), so this refuses nothing today. It is here for the arrival that
	// half-learns the convention, not for the current corpus.
	if loc := reComponentIDAnywhere.FindString(out); loc != "" {
		rep.RefusedReason = fmt.Sprintf(
			"%s reference survived pass 0 — it is not an id attribute's whole value, so it cannot be swapped mechanically and would render EMPTY on every instance; convert this component through the judged pool", loc)
		return tpl, rep, false
	}

	rep.UnprefixedBindings = UnprefixedBindings(out)
	return out, rep, true
}

// RepairConvertedTemplateBindings applies pass 5 to a template that has
// ALREADY been converted — the repair path for the 32 rows the mechanical
// batch wrote before pass 5 existed. Idempotent. Returns ok=false with a reason
// when a dynamic declaration cannot be carried (route to the judged pool).
func RepairConvertedTemplateBindings(converted string) (string, BindingPassReport, []string, bool) {
	baseIDs, fragments, rejects := BindingIDSets(converted)
	if len(rejects) > 0 {
		return converted, BindingPassReport{}, rejects, false
	}
	out, rep := ApplyBindingPasses(converted, baseIDs, fragments)
	return out, rep, nil, true
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
		// bugs_open/260: the gate must not green-light a template the REAL
		// renderer cannot execute. Before the seam had an error channel this
		// call returned regex-substituted output, so a converted template that
		// fails to execute could pass every check below and ship.
		rendered, _, _, err := RenderTemplate(converted, rc, logger)
		if err != nil {
			return false, fmt.Errorf("gate: converted template failed to execute in the real renderer: %w", err)
		}
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
	if ub := UnprefixedBindings(converted); len(ub) > 0 {
		// The 2026-08-19 refusal: ids and script scope are clean, but a
		// binding still carries the OLD id to a lookup (through a variable,
		// a helper, or a concatenation). Two instances would both read clean
		// here and both be broken at runtime — getElementById returns null.
		// Judged pool.
		return true, nil
	}
	return false, nil
}
