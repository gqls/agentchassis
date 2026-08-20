// FILE: platform/orchestration/datahelpers/content_type_violations.go
//
// The ACTIONABLE half of bugs_open/260. The seam fix (component_library.go)
// makes a failed component render fail loudly instead of emitting output no
// template engine produced; what it cannot do on its own is say WHY, because
// text/template's error names the expression, not the contract:
//
//	template: component:41:12: executing "component" at <range .steps>:
//	range can't iterate over "We assess, then we advise, then we file."
//
// That is a true statement about a template. This file turns it into a
// statement about a CONTRACT — which field, declared as what, arrived as what:
//
//	steps[2].branches: declared array (items: object), got string
//
// Every one of the 26 recorded occurrences (7 domains, 2026-08-11 → 08-18) was
// an EXECUTE error, never a parse error: the template was fine and one content
// field was the wrong SHAPE. So the diagnosis is always available from the
// schema plus the content, with no render and no database.
//
// TWO MODES, TWO DIFFERENT AUTHORITIES, and the split is the point:
//
//  1. ENRICHER — unconditional. Called on a render that has ALREADY failed, to
//     name the field in the error. It fires only on failure, so it can refuse
//     nothing that works today and needs no opt-in key.
//
//  2. PRE-RENDER REFUSAL — opt-in, key refuse_mistyped_llm_fields, unsafe
//     default OFF. This checker keys on the SCHEMA, not on the template, so a
//     mistyped field that the template never references renders perfectly well
//     today; refusing it unconditionally would be NEW AUTHORITY OVER CONTENT
//     THAT CURRENTLY RENDERS. Owner ruling 2026-08-02 §2, with the dead-URL
//     guard (dead_url_guard.go:104) as the local precedent — TWO council seats
//     (guardian, architecture) independently demanded default-OFF for this
//     shape on council 98852baa. ⚠ Corrected 2026-08-20: I wrote "three" here,
//     inheriting it from dead_url_guard.go's own header; the corpus says two,
//     and render_guardian's objection that round was a different one (that the
//     rerender path records without refusing). Checkable in one query —
//     SELECT body FROM diagnosis_artifacts WHERE correlation_id LIKE
//     '98852baa%' AND kind='council_report' — which is the point: a comment's
//     account of a council round is a claim about an artefact, so query it.
//
// The hard error at the seam remains the COMPLETE detector. This gate is the
// early, named one, and it is silent for the 75-of-253 active components that
// declare no schema at all (measured 2026-08-19). A green gate is not fleet
// coverage and must never be reported as one.
//
// DELIBERATELY CONSERVATIVE. Absent, nil and empty are NEVER violations — that
// is the presence gate's job (missingRequiredLLMFields, json_envelope.go:451),
// and the two must not disagree.
//
// ⚠ THE ABSENT-FIELD GAP IS REAL AND IS NOT CLOSED HERE (council a44d9eb8
// round 1, bug_historian, gating). Go's missingkey=zero renders an absent field
// as empty with NO error — the mechanism behind the fleet-wide silent blanking
// of article bodies (bugs_closed/004/005) — and this checker does not touch it.
// What covers it today is the presence gate, and it runs at exactly TWO of the
// fifteen render call sites: v3_site_actions.go (the build path) and
// rerender_page_sections_action.go (the rerender pre-check), and only for
// fields the schema marks BOTH source:"llm" AND required. Everywhere else, and
// for every optional field, an absent value still renders empty and silently.
// Stated as a known-open gap rather than implied safe: the seam change makes
// the MISTYPED shape loud; it leaves the ABSENT shape exactly as it was. Filed
// 2026-08-20 as bugs_open/342, UNOWNED, with three fix candidates costed.
//
// An undeclared or unrecognised type is skipped, never guessed. Only
// source:"llm" fields are examined: a resolver fill already passes through
// resolvedValueSatisfiesDeclaredType (render_site_components_action.go), the
// live precedent on this estate for refusing a declared-type violation rather
// than coercing it — and since 2026-08-19 that check IS this file's
// DeclaredTypeSatisfied below, called by both, rather than two copies of one
// rule (council a44d9eb8 round 1, reuse_agent seat).
//
// The deleted validateContentAgainstSchema is NOT prior art for this, despite
// its name: it tested REQUIRED-FIELD PRESENCE over a map[string]string in the
// legacy dialect, and missingRequiredLLMFields already superseded it. There was
// no type check on content to extend.
//
// Coercion is NOT this file's business and is not in bugs
// 260's renderer half at all: repairing at render time would permanently
// silence the only measure of the writer's violation rate.

package datahelpers

import (
	"fmt"
	"sort"
	"strings"
)

// TypeViolation is one content field whose value contradicts its declared type.
// Path is dotted with array indices ("steps[2].branches") so it names the exact
// element, which for the motivating case is the whole diagnosis: the violation
// is NESTED, and a top-level-only check would have reported nothing at all.
type TypeViolation struct {
	Path     string
	Declared string
	Actual   string
}

func (v TypeViolation) String() string {
	return fmt.Sprintf("%s: declared %s, got %s", v.Path, v.Declared, v.Actual)
}

// ContentTypeViolations reports content fields whose value contradicts the type
// their component's input_schema declares. Pure: no database, no render, no
// logger — so it is testable against live row bytes.
//
// SCOPE, stated so nobody reads a clean result as "the content is valid": it
// checks DECLARED-ARRAY fields only, recursing through declared object items
// into nested declared arrays. A declared `text` field holding a map is not
// reported, because text/template renders that (badly, but without erroring)
// and this file exists to explain render FAILURES, not to grade content.
//
// Returns nil for a nil schema, a schema with no readable field set, or clean
// content. Results are sorted by Path so an error message is deterministic.
func ContentTypeViolations(inputSchema map[string]interface{}, content map[string]interface{}) []TypeViolation {
	fields, ok, _ := SchemaContentFields(inputSchema)
	if !ok || len(fields) == 0 || len(content) == 0 {
		return nil
	}

	var out []TypeViolation
	for name, defRaw := range fields {
		def, ok := defRaw.(map[string]interface{})
		if !ok {
			continue
		}
		// Resolver/query fills have their own type guard one layer up; this one
		// is about what the WRITER produced.
		if source, _ := def["source"].(string); source != "llm" {
			continue
		}
		if !declaresArray(def["type"]) {
			continue
		}
		checkArrayValue(name, def, content[name], &out)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

// DescribeTypeViolations renders violations for a log field or an error
// suffix. Empty input gives an empty string, so a caller can append it
// unconditionally without producing a dangling separator.
func DescribeTypeViolations(v []TypeViolation) string {
	if len(v) == 0 {
		return ""
	}
	parts := make([]string, 0, len(v))
	for _, one := range v {
		parts = append(parts, one.String())
	}
	return strings.Join(parts, "; ")
}

// checkArrayValue tests one declared-array value and recurses into its elements.
// path already names the value ("steps", or "steps[2].branches").
func checkArrayValue(path string, def map[string]interface{}, val interface{}, out *[]TypeViolation) {
	// Absent / nil / empty are never violations — see IsEmptyContentValue's own
	// header for the live row that proves this is load-bearing rather than
	// tidy-minded. An empty slice and an empty STRING are both "nothing to
	// show", and templates gate on both.
	if IsEmptyContentValue(val) {
		return
	}
	if !DeclaredTypeSatisfied(declaredTypeOf(def), val) {
		*out = append(*out, TypeViolation{
			Path:     path,
			Declared: describeArrayDecl(def),
			Actual:   actualTypeName(val),
		})
		return
	}
	arr, _ := val.([]interface{})
	if len(arr) == 0 {
		return
	}

	items, hasItems := def["items"].(map[string]interface{})
	if !hasItems || !itemsDeclareObject(items) {
		// A declared array of scalars, or an array whose element shape is not
		// declared: the outer type held, and guessing the rest is how a checker
		// starts refusing valid content.
		return
	}

	nested := nestedFieldDecls(items)
	for i, el := range arr {
		elPath := fmt.Sprintf("%s[%d]", path, i)
		elMap, ok := el.(map[string]interface{})
		if !ok {
			*out = append(*out, TypeViolation{
				Path:     elPath,
				Declared: "object",
				Actual:   actualTypeName(el),
			})
			continue
		}
		for subName, subDefRaw := range nested {
			subDef, ok := subDefRaw.(map[string]interface{})
			if !ok {
				continue
			}
			if !declaresArray(subDef["type"]) {
				continue
			}
			checkArrayValue(elPath+"."+subName, subDef, elMap[subName], out)
		}
	}
}

// IsEmptyContentValue reports whether a content field carries no usable value —
// nil, a blank/whitespace string, or an empty collection. Moved here from
// actions/json_envelope.go on 2026-08-19 so the presence gate
// (missingRequiredLLMFields) and the declared-type checker below share ONE
// definition of empty.
//
// ⚠ THIS IS WHAT STOPS THE TYPE GATE REFUSING A HEALTHY PAGE, and it was caught
// by measurement rather than by review. fundamentallyai.com's
// production-backend-engineering page stores a `mechanism-flow` section whose
// five `steps[].branches` are the EMPTY STRING — declared `array`, holding
// `""`. The template gates them (`{{if $s.branches}}` before
// `{{range $s.branches}}`), so the page renders clean and has been deployed and
// serving since 2026-08-15. A type check that read "" as "a string where an
// array was declared" would have refused a rebuild of a live, correct page: the
// first version of ContentTypeViolations did exactly that, and this is the only
// such row on the estate, so nothing but the census would have found it.
//
// "Not filled in" is the presence gate's business at every type. Only a value
// that is actually THERE can contradict its declared type.
func IsEmptyContentValue(v interface{}) bool {
	switch t := v.(type) {
	case nil:
		return true
	case string:
		return strings.TrimSpace(t) == ""
	case []interface{}:
		return len(t) == 0
	case map[string]interface{}:
		return len(t) == 0
	}
	return false
}

// DeclaredTypeSatisfied reports whether a value may be handed to a schema field
// declaring this type. MOVED here from render_site_components_action.go on
// 2026-08-19 so the estate has ONE answer to that question; its history, kept
// verbatim because it is the reason for the narrowness:
//
//	Deliberately narrow: only the shapes whose mismatch DESTROYS the render are
//	enforced — a non-array under {{range}} errors the whole template into the
//	silent regex-fallback renderer (the bug_historian objection on council corr
//	56ab6e23). Every other declared type ("text", "url", "number", unknown,
//	absent) is allowed through untouched: enforcing those would change behaviour
//	on ~2,200 live fields on unmeasured ground for no render-safety gain, and a
//	wrong scalar in a bare output slot renders wrong text, not a destroyed slot.
//
// The regex fallback that sentence names is now deleted (bugs_open/260), so the
// consequence is sharper, not milder: a non-array under {{range}} now fails the
// render outright. The narrowness stands for the same reason it always did.
func DeclaredTypeSatisfied(declared string, v interface{}) bool {
	switch declared {
	case "array", "list":
		_, ok := v.([]interface{})
		return ok
	default:
		return true
	}
}

// declaresArray reads the two array spellings this estate actually uses. `list`
// is not a typo to be normalised away: 5 live fields declare it.
func declaresArray(t interface{}) bool {
	s, _ := t.(string)
	return s == "array" || s == "list"
}

func declaredTypeOf(def map[string]interface{}) string {
	s, _ := def["type"].(string)
	return s
}

func describeArrayDecl(def map[string]interface{}) string {
	t, _ := def["type"].(string)
	if items, ok := def["items"].(map[string]interface{}); ok && itemsDeclareObject(items) {
		return t + " (items: object)"
	}
	return t
}

// itemsDeclareObject reports whether an `items` block says "each element is an
// object".
//
// WHY THIS IS NOT SchemaContentFields' JOB (council a44d9eb8 round 2,
// reuse_agent — a fair question, and the answer is a boundary rather than a
// refusal). SchemaContentFields normalises the FIELD-SET dialect: v2 `fields`
// versus legacy top-level `properties`, and ContentTypeViolations calls it for
// exactly that. It does NOT normalise the `items` sub-dialect — it copies
// `items` through verbatim, along with `source`, `on_missing`, `fallback`,
// `missing_reason` and `min_items` — so there was nothing there to extend.
// Adding element-shape normalisation to it would change what every one of its
// callers reads (the generation planner, the render gate, the post-deploy
// audit) for the benefit of the one caller that walks INTO items, which is the
// widening this estate's own reuse discipline warns about in the other
// direction.
//
// BOTH live dialects mean "each element is an object" and neither is going
// away, so both are read here rather than one being declared canonical:
//
//   - JSON-Schema-ish (2 components, incl. mechanism-flow, the motivating case):
//     {"type":"object","required":[...],"properties":{...}}
//   - example-value (12 components, e.g. faq, features):
//     {"question":"string","answer":"string"}
//
// The example-value dialect is recognised by its VALUES being type names rather
// than declaration objects. A bare string items ("string") declares scalars and
// is correctly not an object shape.
func itemsDeclareObject(items map[string]interface{}) bool {
	if t, ok := items["type"].(string); ok {
		return t == "object"
	}
	if _, ok := items["properties"].(map[string]interface{}); ok {
		return true
	}
	if len(items) == 0 {
		return false
	}
	for _, v := range items {
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
}

// nestedFieldDecls projects either items dialect onto the {name: {type: ...}}
// shape the recursion consumes. For the example-value dialect a value of
// "array"/"list" is the only nested declaration available, and that is enough:
// it is the outer shape a range can fail on.
func nestedFieldDecls(items map[string]interface{}) map[string]interface{} {
	if props, ok := items["properties"].(map[string]interface{}); ok {
		return props
	}
	out := make(map[string]interface{}, len(items))
	for name, v := range items {
		if s, ok := v.(string); ok {
			out[name] = map[string]interface{}{"type": s}
		}
	}
	return out
}

// actualTypeName names what arrived in the vocabulary the SCHEMA uses, not Go's
// — "string", not "string"; "object", not "map[string]interface {}" — because
// the reader of this message is comparing it against a declared type, and %T
// makes them do the translation.
func actualTypeName(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		return "string"
	case bool:
		return "boolean"
	case float64, float32, int, int64, int32:
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return fmt.Sprintf("%T", t)
	}
}
