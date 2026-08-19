// FILE: platform/orchestration/actions/component_instance_bindings.go
//
// The binding passes the deterministic converter was MISSING, and the detector
// that would have caught their absence — added 2026-08-19 after the mechanical
// batch (bugs_open/283 §13) was found to have written 32 of 69 templates with
// at least one DANGLING binding, 14 of them serving live.
//
// What went wrong, precisely. The converter renamed every `id="x"` and every
// literal `getElementById('x')`, then checked completeness by looking for those
// same literal forms surviving. The premise — that a binding which survives will
// CONTAIN the id literal at the binding site — is false for three constructions
// the live corpus uses constantly:
//
//	A. the id literal lives somewhere else and travels through a variable:
//	     var ids = ['amount', 'interest'];  ids.forEach(id => getElementById(id))
//	     var fields = [{ id: 'gsfc-accel', … }];  getElementById(field.id)
//	     function el(id) { return document.getElementById(id) }  el('rw-ev')
//	   The declaration `id="amount"` was renamed; the literal 'amount' was not;
//	   getElementById returns null; the first `.addEventListener` on it throws
//	   and the whole IIFE aborts. tool-loan-repayment, tool-gripper-safety-
//	   factor-calculator and eighteen more.
//	B. the id is DECLARED by concatenation inside a JS string that builds
//	   markup — `'<input id="name-' + index + '">'` — so the declared-id regex
//	   captured `name-' + index + ` as an "id", pass 1 prefixed the declaration,
//	   and nothing could prefix the lookup or the label's `for`.
//	C. the id is LOOKED UP by concatenation — `getElementById('step-' + n)` for
//	   declared `step-1 … step-5` — so the literal lookup pass never matched.
//
// The three detect and repair mechanically, which is why this file exists
// rather than the 32 going to the judged pool: the scripts are already scoped;
// only the binding text moved short of the declaration.
//
// Two rules carried over from the converter: every rename is anchored text with
// a boundary, and anything the passes cannot place is REPORTED by
// UnprefixedBindings so the gate routes the component to the judged pool
// instead of writing it. The detector is also the completeness check, applied
// to every converted and every repaired template before any write — so the
// class this file fixes cannot recur silently even if a future construction
// defeats the passes.
package actions

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// reInstanceIDDecl matches an id= attribute whose value carries the instance
// prefix — how a CONVERTED template declares its ids. Used to recover the base
// id set when repairing a template that has already been converted.
var (
	reInstanceIDDeclDQ = regexp.MustCompile(`\sid="\{\{\.InstanceID\}\}-([^"]+)"`)
	reInstanceIDDeclSQ = regexp.MustCompile(`\sid='\{\{\.InstanceID\}\}-([^']+)'`)
	reBareIDDeclDQ     = regexp.MustCompile(`\sid="([^"{}]+)"`)
	reBareIDDeclSQ     = regexp.MustCompile(`\sid='([^'{}]+)'`)
	// reDynamicFragment says a captured "id" is really a concatenation
	// fragment: it contains a quote, a plus, a template-literal dollar or
	// whitespace. Class B.
	reDynamicFragment = regexp.MustCompile(`['"+$\s]`)
	reBlockComment    = regexp.MustCompile(`(?s)/\*.*?\*/`)
	reLineComment     = regexp.MustCompile(`(?m)^\s*//.*$`)
	// reScriptElement captures open tag, body, close tag of each inline script.
	reScriptElement = regexp.MustCompile(`(?s)(<script\b[^>]*>)(.*?)(</script>)`)
)

// BindingIDSets recovers, from a template in either state (bare or already
// converted), the two sets the binding passes work on:
//
//   - ids: the base ids (without prefix) that are declared by a complete,
//     literal id= attribute;
//   - fragments: the static, `-`/`_`-terminated prefixes of ids that are
//     declared DYNAMICALLY (class B: `id="name-' + index + '"`). A dynamic
//     declaration whose static part is empty or does not end on a separator
//     cannot be prefixed mechanically and is returned in rejects instead.
//
// A template may mix both, so both sets are computed from all declarations.
func BindingIDSets(tpl string) (ids []string, fragments []string, rejects []string) {
	seen := map[string]bool{}
	for _, re := range []*regexp.Regexp{reInstanceIDDeclDQ, reInstanceIDDeclSQ, reBareIDDeclDQ, reBareIDDeclSQ} {
		for _, m := range re.FindAllStringSubmatch(tpl, -1) {
			seen[m[1]] = true
		}
	}
	fragSeen := map[string]bool{}
	for raw := range seen {
		if !reDynamicFragment.MatchString(raw) {
			ids = append(ids, raw)
			continue
		}
		// Class B. The static part is everything before the first quote.
		cut := len(raw)
		for i, ch := range raw {
			if ch == '\'' || ch == '"' || ch == '`' || ch == '$' {
				cut = i
				break
			}
		}
		frag := strings.TrimSpace(raw[:cut])
		if frag == "" || !(strings.HasSuffix(frag, "-") || strings.HasSuffix(frag, "_")) {
			rejects = append(rejects, raw)
			continue
		}
		fragSeen[frag] = true
	}
	for f := range fragSeen {
		fragments = append(fragments, f)
	}
	sort.Strings(ids)
	sort.Strings(fragments)
	sort.Strings(rejects)
	return ids, fragments, rejects
}

// quoted returns a pattern matching s inside matching quotes of any of the
// three JS kinds. RE2 has no backreferences, so the pairs are spelled out.
func quoted(s string) string {
	q := regexp.QuoteMeta(s)
	return "(?:'" + q + "'|\"" + q + "\"|`" + q + "`)"
}

// concatSitePatterns returns the patterns that match one unprefixed
// concatenation site on prefix p, paired with the replacement that prefixes
// it. Shared by the rename pass and the detector so they cannot disagree:
// `'p' +`, `"p" +`, `'#p' +`, `"#p" +`, “ `p${ “, “ `#p${ “, and the
// markup-building forms `for="p' +` / `id="p' +` (either outer quote).
func concatSitePatterns(p string) []struct {
	re   *regexp.Regexp
	repl string
} {
	q := regexp.QuoteMeta(p)
	return []struct {
		re   *regexp.Regexp
		repl string
	}{
		{regexp.MustCompile(`'` + q + `'(\s*\+)`), "'" + instancePrefix + p + "'${1}"},
		{regexp.MustCompile(`"` + q + `"(\s*\+)`), `"` + instancePrefix + p + `"${1}`},
		{regexp.MustCompile(`'#` + q + `'(\s*\+)`), "'#" + instancePrefix + p + "'${1}"},
		{regexp.MustCompile(`"#` + q + `"(\s*\+)`), `"#` + instancePrefix + p + `"${1}`},
		{regexp.MustCompile("`" + q + `\$\{`), "`" + instancePrefix + p + "${"},
		{regexp.MustCompile("`#" + q + `\$\{`), "`#" + instancePrefix + p + "${"},
		{regexp.MustCompile(`\b(for|id)="` + q + `'(\s*\+)`), `${1}="` + instancePrefix + p + `'${2}`},
		{regexp.MustCompile(`\b(for|id)='` + q + `"(\s*\+)`), `${1}='` + instancePrefix + p + `"${2}`},
	}
}

// separatorPrefixes returns every proper prefix of id that ends on `-` or `_`
// and is at least two characters long — the candidate static halves of a
// class-C concatenated lookup (`'step-' + n` for id `step-1`).
func separatorPrefixes(id string) []string {
	var out []string
	for i := 2; i < len(id); i++ {
		if id[i-1] == '-' || id[i-1] == '_' {
			out = append(out, id[:i])
		}
	}
	return out
}

// concatPrefixesInUse returns the separator-prefixes of the declared ids that
// the script actually uses in a concatenated or template-literal position —
// the class-C set. Computed from the script bodies outside comments so a
// tool-doc mention cannot manufacture one.
func concatPrefixesInUse(tpl string, ids []string) []string {
	body := scriptBodiesOutsideComments(tpl)
	seen := map[string]bool{}
	for _, id := range ids {
		for _, p := range separatorPrefixes(id) {
			if seen[p] {
				continue
			}
			// Either form counts as "in use": the bare site a repair will
			// prefix, and the already-prefixed site a repair produced — the
			// composition hazard is checked on OUTPUT too.
			for _, form := range []string{p, instancePrefix + p} {
				for _, pat := range concatSitePatterns(form) {
					if pat.re.MatchString(body) {
						seen[p] = true
						break
					}
				}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

// scriptBodiesOutsideComments concatenates the inline <script> bodies with
// block and line comments removed. Detection reads THIS; renames are applied
// to the real body (a rename inside a comment is harmless, a false detection
// from one is not).
func scriptBodiesOutsideComments(tpl string) string {
	var b strings.Builder
	for _, m := range reScriptElement.FindAllStringSubmatch(tpl, -1) {
		body := reBlockComment.ReplaceAllString(m[2], "")
		body = reLineComment.ReplaceAllString(body, "")
		b.WriteString(body)
		b.WriteString("\n")
	}
	return b.String()
}

// BindingPassReport counts what the passes did, for the result and the log.
type BindingPassReport struct {
	LiteralIDsRenamed  int      // class A: bare literals equal to a declared id
	ConcatPrefixes     []string // class B+C: the static prefixes that were prefixed
	ConcatSitesRenamed int      // number of concatenation sites rewritten
	SkippedContexts    []string // class-A literals left alone because their context reads as not-a-binding
}

// reLiteralContextRefuse — a bare literal equal to an id is NOT renamed when it
// sits where a string that merely happens to equal an id would sit: the
// right-hand side of a comparison, a `case` label, an object KEY, or a
// computed property access on an identifier. Those are reported via
// UnprefixedBindings and the gate routes the component to the judged pool,
// where a reader can tell. Everything else — array elements, object VALUES,
// call arguments, assignments — is a binding in every live template read.
var reLiteralContextRefuse = regexp.MustCompile(`(===|!==|==|!=|\bcase|[A-Za-z0-9_\)\]]\[)\s*$`)

// reObjectKeyBefore / reObjectKeyAfter together recognise an object KEY:
// `{ 'x': 1 }` or `, 'x': 1` — preceded by `{` or `,` AND followed by `:`.
// Both halves are required so a ternary `big ? 'x' : 'y'` (a binding) still
// renames: its `'x'` is preceded by `?`, its `'y'` by `:`.
var (
	reObjectKeyBefore = regexp.MustCompile(`[{,]\s*$`)
	reObjectKeyAfter  = regexp.MustCompile(`^\s*:`)
)

// ApplyBindingPasses rewrites the script bodies of tpl so that every
// recognised binding construction carries the instance prefix. `ids` and
// `fragments` come from BindingIDSets. It is idempotent: a literal already
// prefixed does not match the bare-literal pattern, and a prefixed
// concatenation site does not match the bare one.
func ApplyBindingPasses(tpl string, ids, fragments []string) (string, BindingPassReport) {
	var rep BindingPassReport

	// Class B fragments and class C prefixes share one rename.
	prefixes := append([]string{}, fragments...)
	prefixes = append(prefixes, concatPrefixesInUse(tpl, ids)...)
	sort.Slice(prefixes, func(i, j int) bool { return len(prefixes[i]) > len(prefixes[j]) })
	seenP := map[string]bool{}
	for _, p := range prefixes {
		if !seenP[p] {
			seenP[p] = true
			rep.ConcatPrefixes = append(rep.ConcatPrefixes, p)
		}
	}

	// Longest id first, boundary-anchored by the quotes, so an id that is a
	// prefix of another cannot be renamed through it.
	sortedIDs := append([]string{}, ids...)
	sort.Slice(sortedIDs, func(i, j int) bool { return len(sortedIDs[i]) > len(sortedIDs[j]) })

	out := reScriptElement.ReplaceAllStringFunc(tpl, func(m string) string {
		g := reScriptElement.FindStringSubmatch(m)
		open, body, close := g[1], g[2], g[3]

		// Pass A — bare literals equal to a declared id.
		for _, id := range sortedIDs {
			re := regexp.MustCompile(quoted(id))
			// Walk matches by index: the rename needs the context BEFORE each
			// literal, which ReplaceAllStringFunc cannot see.
			var b strings.Builder
			last := 0
			for _, loc := range re.FindAllStringIndex(body, -1) {
				before := body[:loc[0]]
				// Already prefixed: the char before the opening quote is not a
				// concern — the PREFIX would be inside the quotes. A literal
				// `'{{.InstanceID}}-x'` never matches quoted("x"), so any match
				// here is bare.
				tail := before
				if len(tail) > 40 {
					tail = tail[len(tail)-40:]
				}
				after := body[loc[1]:]
				if len(after) > 8 {
					after = after[:8]
				}
				if reLiteralContextRefuse.MatchString(tail) ||
					(reObjectKeyBefore.MatchString(tail) && reObjectKeyAfter.MatchString(after)) {
					rep.SkippedContexts = append(rep.SkippedContexts, id)
					continue
				}
				b.WriteString(body[last:loc[0]])
				quote := body[loc[0] : loc[0]+1]
				b.WriteString(quote + instancePrefix + id + quote)
				last = loc[1]
				rep.LiteralIDsRenamed++
			}
			b.WriteString(body[last:])
			body = b.String()
		}

		// Pass B/C — concatenated sites: `'P' +`, `'#P' +`, `` `P${ ``,
		// `` `#P${ ``, and the markup-building forms `for="P' +` / `id="P' +`
		// (the latter only when pass 1 did not already catch it).
		for _, p := range rep.ConcatPrefixes {
			for _, pat := range concatSitePatterns(p) {
				n := len(pat.re.FindAllStringIndex(body, -1))
				if n > 0 {
					body = pat.re.ReplaceAllString(body, pat.repl)
					rep.ConcatSitesRenamed += n
				}
			}
		}
		return open + body + close
	})
	sort.Strings(rep.SkippedContexts)
	return out, rep
}

// compositionHazards reports the shape the passes cannot repair and the other
// checks cannot see: a concatenated lookup `'P' + v` where v may itself be a
// declared id that ALSO travels through a literal — `validateField('fleetSize')`
// then `getElementById('fg-' + id)` with both `fleetSize` and `fg-fleetSize`
// declared. Prefixing both halves yields `{{.InstanceID}}-fg-{{.InstanceID}}-
// fleetSize`, which dangles, and the output reads clean on every other check.
// The conjunction that makes it a hazard: P is concatenated in the script,
// some X with P+X declared is ALSO declared, and X appears as a string literal
// (bare or already prefixed — either way it travels through a variable).
// Reported on input and on output alike; the remedy is the judged pool.
func compositionHazards(tpl string, ids []string, prefixes []string) []string {
	declared := map[string]bool{}
	for _, id := range ids {
		declared[id] = true
	}
	body := scriptBodiesOutsideComments(tpl)
	var out []string
	for _, p := range prefixes {
		for _, x := range ids {
			if !declared[p+x] {
				continue
			}
			if regexp.MustCompile(quoted(x)).MatchString(body) || regexp.MustCompile(quoted(instancePrefix+x)).MatchString(body) {
				out = append(out, fmt.Sprintf("composition: prefix %q is concatenated and id %q (with %q also declared) travels through a literal — the composed lookup cannot be shown single-prefixed", p, x, p+x))
			}
		}
	}
	return out
}

// UnprefixedBindings is the completeness detector: every place in the script
// bodies (outside comments) where a declared id — or a concatenation prefix of
// one — still appears UNPREFIXED in a binding-shaped position. An empty result
// is the precondition for any write; a non-empty one routes the component to
// the judged pool. Reports are human-readable because that is who reads them.
func UnprefixedBindings(tpl string) []string {
	ids, fragments, rejects := BindingIDSets(tpl)
	body := scriptBodiesOutsideComments(tpl)
	var out []string

	for _, r := range rejects {
		out = append(out, fmt.Sprintf("dynamic id declaration %q has no static `-`/`_` prefix to carry the token", r))
	}
	// Class A.
	for _, id := range ids {
		re := regexp.MustCompile(quoted(id))
		if n := len(re.FindAllStringIndex(body, -1)); n > 0 {
			out = append(out, fmt.Sprintf("literal %q appears bare %d time(s) in script — a binding through a variable or helper", id, n))
		}
	}
	// Class B+C.
	prefixes := append([]string{}, fragments...)
	prefixes = append(prefixes, concatPrefixesInUse(tpl, ids)...)
	seen := map[string]bool{}
	for _, p := range prefixes {
		if seen[p] {
			continue
		}
		seen[p] = true
		n := 0
		for _, pat := range concatSitePatterns(p) {
			n += len(pat.re.FindAllStringIndex(body, -1))
		}
		if n > 0 {
			out = append(out, fmt.Sprintf("concatenated lookup/declaration on prefix %q is unprefixed at %d site(s)", p, n))
		}
	}
	allPrefixes := make([]string, 0, len(seen))
	for p := range seen {
		allPrefixes = append(allPrefixes, p)
	}
	sort.Strings(allPrefixes)
	out = append(out, compositionHazards(tpl, ids, allPrefixes)...)
	sort.Strings(out)
	return out
}
