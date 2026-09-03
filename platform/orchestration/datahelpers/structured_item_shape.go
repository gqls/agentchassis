package datahelpers

import (
	"strconv"
	"strings"
)

// Nested item shapes, for the writer's prompt (bugs_open/437).
//
// THE DEFECT THIS EXISTS TO CLOSE. A component may declare an array field whose
// elements are objects that themselves carry an array — mechanism-flow's
// `steps[].branches` is the live one. The section planner projected that schema
// to a flat list of element field NAMES (actions.extractArrayItemFields), and
// page-content-writer's prompt built its JSON exemplar from those names, so the
// prompt told the writer:
//
//	"steps": [{ "body": "...", "branches": "...", "marker": "...", … }]
//
// i.e. it declared `branches` a STRING. The writer obeyed, ContentTypeViolations
// below refused the render (correctly), and the page never built — 119 failed
// builds across six sites in the fortnight to 2026-09-02, none of them the
// model's fault. A prompt exemplar is a demonstration, and a demonstration
// generated from a lossy projection of the schema teaches the wrong contract.
//
// StructuredItemShape is that projection made lossless for the nested case: it
// renders the shape the gate will enforce, from the same package as the gate, so
// the two cannot drift apart. It lives here rather than beside the planner for
// exactly that reason — itemsDeclareObject/nestedFieldDecls already read both
// live `items` dialects, and a third reader of that dialect split in another
// package is the drift class this estate keeps paying for.

// shapeMaxDepth bounds the recursion. Parsed JSON cannot cycle, so this is
// belt-and-braces plus prompt-size hygiene; the deepest live schema is 2
// (steps → branches) as of 2026-09-03.
const shapeMaxDepth = 3

// StructuredItemShape renders the prompt-facing shape of one array field whose
// elements carry STRUCTURE — a nested array or object — and returns:
//
//   - valueShape: a one-line JSON skeleton for the whole field value, e.g.
//     `[{ "body": "...", "branches": [{ "body": "...", "label": "..." }], … }]`.
//     Keys are sorted (Go map order is otherwise random and a prompt that
//     changes shape between runs is uncacheable and untestable) and the spacing
//     matches the flat exemplar the prompt already emits, so a component's
//     exemplar differs from today's only in its nested parts.
//   - itemNotes: one sentence per structured element property, naming the shape
//     and carrying the property's own schema description — which no prompt has
//     ever seen, because the flat projection dropped it.
//
// BOTH ARE EMPTY unless at least one element property is itself an array or an
// object. That is the whole safety story for the deploy: a field whose flat
// exemplar was already correct produces no new spec keys, so its prompt stays
// byte-identical. 1 of the live components qualifies today (mechanism-flow);
// the rest are untouched.
//
// Empty stays legal. The notes say to omit an optional structured property or
// send [], never to fill it for the sake of filling it: IsEmptyContentValue
// treats absent/nil/empty as no violation at every depth, templates gate on
// emptiness (`{{if $s.branches}}`), and a live page has served five steps with
// `branches: ""` since 2026-08-15. The omission advice is suppressed for a
// property the item schema marks required, or the note would invite the one
// omission the schema forbids.
func StructuredItemShape(fieldDef map[string]interface{}) (valueShape string, itemNotes []string) {
	if !declaresArray(fieldDef["type"]) {
		return "", nil
	}
	decls, required := elementDecls(fieldDef)
	if len(decls) == 0 {
		return "", nil
	}

	names := sortedKeys(decls)
	structured := make([]string, 0, len(names))
	for _, name := range names {
		if isStructuredDecl(decls[name]) {
			structured = append(structured, name)
		}
	}
	if len(structured) == 0 {
		return "", nil
	}

	valueShape = "[" + renderObjectSkeleton(decls, 1) + "]"
	for _, name := range structured {
		if note := structuredPropNote(name, declMap(decls[name]), required[name]); note != "" {
			itemNotes = append(itemNotes, note)
		}
	}
	return valueShape, itemNotes
}

// elementDecls projects whichever element dialect this field uses onto the
// {name: {type: …}} shape the renderer consumes, and returns the element-level
// required set alongside it. Three dialects are live and all three are read
// here, matching actions.extractArrayItemFields' entry conditions so the prompt
// and the planner never disagree about which fields have a declared element
// shape:
//
//   - JSON-Schema `items` ({"type":"object","required":[…],"properties":{…}})
//   - example-value `items` ({"question":"string","answer":"string"})
//   - `item_schema` ({"name":{"type":"string"},…}), the info-card-grid spelling
//
// Only the JSON-Schema dialect can express a required set; the other two return
// an empty one, which is the honest answer rather than a guess.
func elementDecls(fieldDef map[string]interface{}) (map[string]interface{}, map[string]bool) {
	if items, ok := fieldDef["items"].(map[string]interface{}); ok && itemsDeclareObject(items) {
		return nestedFieldDecls(items), requiredSet(items)
	}
	if itemSchema, ok := fieldDef["item_schema"].(map[string]interface{}); ok {
		out := make(map[string]interface{}, len(itemSchema))
		for name, v := range itemSchema {
			switch t := v.(type) {
			case map[string]interface{}:
				out[name] = t
			case string:
				out[name] = map[string]interface{}{"type": t}
			}
		}
		return out, requiredSet(itemSchema)
	}
	return nil, nil
}

func requiredSet(decl map[string]interface{}) map[string]bool {
	raw, ok := decl["required"].([]interface{})
	if !ok {
		return nil
	}
	out := make(map[string]bool, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out[s] = true
		}
	}
	return out
}

// isStructuredDecl reports whether a property declaration says "this value is
// itself a collection" — the only case a flat `"…"` exemplar gets WRONG. A
// scalar property is described perfectly well by the existing flat rendering.
func isStructuredDecl(raw interface{}) bool {
	def := declMap(raw)
	if def == nil {
		return false
	}
	if declaresArray(def["type"]) {
		return true
	}
	t, _ := def["type"].(string)
	return t == "object"
}

// renderSkeleton renders one declaration as its exemplar value. Depth-capped:
// at the cap a collection renders as its own empty form rather than as `"..."`,
// because a skeleton that mis-states a type at depth N is the very defect this
// file exists to remove, and an empty collection is both honest and valid JSON.
func renderSkeleton(raw interface{}, depth int) string {
	def := declMap(raw)
	if def == nil {
		return `"..."`
	}
	switch {
	case declaresArray(def["type"]):
		if depth >= shapeMaxDepth {
			return "[]"
		}
		items, ok := def["items"].(map[string]interface{})
		if !ok || !itemsDeclareObject(items) {
			// A declared array of scalars, or one whose element shape is not
			// declared: say "array of values" and stop guessing — the same
			// restraint checkArrayValue shows when it declines to check
			// undeclared element shapes.
			return `["..."]`
		}
		return "[" + renderObjectSkeleton(nestedFieldDecls(items), depth+1) + "]"
	case declaredTypeOf(def) == "object":
		if depth >= shapeMaxDepth {
			return "{}"
		}
		props, ok := def["properties"].(map[string]interface{})
		if !ok || len(props) == 0 {
			return "{}"
		}
		return renderObjectSkeleton(props, depth+1)
	default:
		return `"..."`
	}
}

// renderObjectSkeleton renders `{ "a": …, "b": … }` with sorted keys and the
// spacing the prompt's existing flat exemplar uses, so only the nested parts of
// a structured component's exemplar differ from what it emitted before.
// Keys are quoted by strconv.Quote, not by hand: a schema is data, and a key
// carrying a quote or a backslash must not be able to produce a skeleton that
// no longer parses as JSON.
func renderObjectSkeleton(decls map[string]interface{}, depth int) string {
	names := sortedKeys(decls)
	parts := make([]string, 0, len(names))
	for _, name := range names {
		parts = append(parts, strconv.Quote(name)+": "+renderSkeleton(decls[name], depth))
	}
	if len(parts) == 0 {
		return "{}"
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// structuredPropNote writes the one sentence the prompt shows for a structured
// property. It carries the property's own description because the flat
// projection dropped it — mechanism-flow's `branches` has said "a decision
// point: two or more outcomes, rendered side by side" in its schema since it was
// created, and no writer has ever been shown it.
func structuredPropNote(name string, def map[string]interface{}, isRequired bool) string {
	if def == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("`" + name + "` must be ")

	switch {
	case declaresArray(def["type"]):
		items, ok := def["items"].(map[string]interface{})
		if ok && itemsDeclareObject(items) {
			b.WriteString("an array of objects, each with " + backquotedList(sortedKeys(nestedFieldDecls(items))))
		} else {
			b.WriteString("an array of values")
		}
	case declaredTypeOf(def) == "object":
		props, _ := def["properties"].(map[string]interface{})
		if len(props) > 0 {
			b.WriteString("an object with " + backquotedList(sortedKeys(props)))
		} else {
			b.WriteString("an object")
		}
	default:
		return ""
	}

	b.WriteString(", never a sentence of prose")
	if d := propDescription(def); d != "" {
		b.WriteString(" — " + strings.TrimRight(d, ". "))
	}
	b.WriteString(".")
	if !isRequired {
		// Only for an optional property: for a required one this advice would
		// invite the omission the schema forbids.
		b.WriteString(" Omit it, or use [] where the element has none.")
	}
	return b.String()
}

// propDescription prefers llm_guidance (the house key the top-level field specs
// already read) and falls back to JSON Schema's description, which is what the
// nested declarations actually carry today.
func propDescription(def map[string]interface{}) string {
	if g, ok := def["llm_guidance"].(string); ok && strings.TrimSpace(g) != "" {
		return strings.TrimSpace(g)
	}
	if d, ok := def["description"].(string); ok && strings.TrimSpace(d) != "" {
		return strings.TrimSpace(d)
	}
	return ""
}

func backquotedList(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, n := range names {
		quoted = append(quoted, "`"+n+"`")
	}
	return strings.Join(quoted, ", ")
}

func declMap(raw interface{}) map[string]interface{} {
	def, _ := raw.(map[string]interface{})
	return def
}
