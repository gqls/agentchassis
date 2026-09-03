package datahelpers

import (
	"encoding/json"
	"strings"
	"testing"
)

// mechanismFlowStepsFieldDef is mechanism-flow's `steps` field declaration,
// verbatim from the live row (2026-09-03) and from
// sql_for_agents/247_mechanism_flow_component.sql:203-227 — INCLUDING
// `branches`, the nested array of objects the flat projection dropped and the
// prompt therefore declared a string (bugs_open/437).
//
// Deliberately a second fixture rather than an edit of
// actions.mechanismFlowLegacySchema(), which models the same component one level
// deep for the key-reconciler's tests: changing that one would rewrite every
// `want` list in a test file about a different defect.
func mechanismFlowStepsFieldDef() map[string]interface{} {
	var def map[string]interface{}
	if err := json.Unmarshal([]byte(`{
	  "type": "array", "minItems": 2, "source": "llm", "required": true,
	  "items": {
	    "type": "object",
	    "required": ["title"],
	    "properties": {
	      "marker": {"type": "string", "description": "optional override for the auto number"},
	      "title":  {"type": "string"},
	      "body":   {"type": "string"},
	      "note":   {"type": "string", "description": "an aside, rendered as a callout"},
	      "branches": {
	        "type": "array",
	        "description": "a decision point: two or more outcomes, rendered side by side",
	        "items": {
	          "type": "object",
	          "required": ["body"],
	          "properties": {
	            "label": {"type": "string"},
	            "body":  {"type": "string"}
	          }
	        }
	      }
	    }
	  }
	}`), &def); err != nil {
		panic(err)
	}
	return def
}

func mustUnmarshalDef(t *testing.T, raw string) map[string]interface{} {
	t.Helper()
	var def map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &def); err != nil {
		t.Fatalf("fixture is not valid JSON: %v", err)
	}
	return def
}

// The motivating case, asserted byte-exactly: this string IS the instruction the
// writer copies, so an approximate assertion would not be an assertion.
func TestStructuredItemShape_MechanismFlowStepsDeclaresBranchesAsArray(t *testing.T) {
	shape, notes := StructuredItemShape(mechanismFlowStepsFieldDef())

	const want = `[{ "body": "...", "branches": [{ "body": "...", "label": "..." }], ` +
		`"marker": "...", "note": "...", "title": "..." }]`
	if shape != want {
		t.Errorf("value shape:\n got %s\nwant %s", shape, want)
	}

	// The defect in one assertion: the old flat exemplar's spelling must not
	// survive anywhere in the new one.
	if strings.Contains(shape, `"branches": "..."`) {
		t.Errorf("value shape still declares branches a string: %s", shape)
	}

	if len(notes) != 1 {
		t.Fatalf("want exactly one note (branches is the only structured property), got %d: %v", len(notes), notes)
	}
	const wantNote = "`branches` must be an array of objects, each with `body`, `label`, " +
		"never a sentence of prose — a decision point: two or more outcomes, rendered side by side. " +
		"Omit it, or use [] where the element has none."
	if notes[0] != wantNote {
		t.Errorf("note:\n got %s\nwant %s", notes[0], wantNote)
	}
}

// Empty must stay reachable. A live page has served five steps with
// `branches: ""` since 2026-08-15 and IsEmptyContentValue keeps that legal at
// every depth; a prompt that only ever demonstrates a filled branches would
// push writers to invent decision points that are not in the source material.
func TestStructuredItemShape_OptionalPropertyKeepsTheOmissionAdvice(t *testing.T) {
	_, notes := StructuredItemShape(mechanismFlowStepsFieldDef())
	if len(notes) != 1 || !strings.Contains(notes[0], "Omit it, or use []") {
		t.Fatalf("optional structured property must carry omission advice, got %v", notes)
	}
}

// …and must NOT be offered when the schema forbids it, or the note would invite
// the one omission that fails the build.
func TestStructuredItemShape_RequiredPropertySuppressesTheOmissionAdvice(t *testing.T) {
	def := mechanismFlowStepsFieldDef()
	items := def["items"].(map[string]interface{})
	items["required"] = []interface{}{"title", "branches"}

	_, notes := StructuredItemShape(def)
	if len(notes) != 1 {
		t.Fatalf("want one note, got %v", notes)
	}
	if strings.Contains(notes[0], "Omit it") {
		t.Errorf("required property must not be told it may be omitted: %s", notes[0])
	}
	if !strings.Contains(notes[0], "must be an array of objects") {
		t.Errorf("required property still needs its shape stated: %s", notes[0])
	}
}

// The byte-identical-prompt guarantee, which is the whole deploy-safety story:
// a component whose flat exemplar was already correct emits no new spec keys, so
// its prompt cannot change. These are the two other live dialects.
func TestStructuredItemShape_FlatAndScalarDialectsEmitNothing(t *testing.T) {
	cases := map[string]string{
		"faq example-value items": `{"type":"array","source":"llm",
			"items":{"question":"string","answer":"string"}}`,
		"info-card-grid item_schema": `{"type":"array","source":"llm",
			"item_schema":{"title":{"type":"string"},"description":{"type":"string"}}}`,
		"json-schema items, scalars only": `{"type":"array","source":"llm",
			"items":{"type":"object","properties":{"title":{"type":"string"},"body":{"type":"string"}}}}`,
		"array of scalars":                     `{"type":"array","source":"llm","items":"string"}`,
		"not an array at all":                  `{"type":"text","source":"llm"}`,
		"array with no declared element shape": `{"type":"array","source":"llm"}`,
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			shape, notes := StructuredItemShape(mustUnmarshalDef(t, raw))
			if shape != "" || notes != nil {
				t.Errorf("want no shape and no notes, got %q / %v", shape, notes)
			}
		})
	}
}

// The example-value dialect can say "array" as a type NAME. No live component
// does today (0 as of 2026-09-03), but the flat exemplar would be just as wrong
// for one, so it is handled rather than left to be rediscovered.
func TestStructuredItemShape_ExampleValueDialectArrayTypeName(t *testing.T) {
	shape, notes := StructuredItemShape(mustUnmarshalDef(t, `{"type":"array","source":"llm",
		"items":{"title":"string","tags":"array"}}`))

	const want = `[{ "tags": ["..."], "title": "..." }]`
	if shape != want {
		t.Errorf("value shape:\n got %s\nwant %s", shape, want)
	}
	if len(notes) != 1 || !strings.Contains(notes[0], "`tags` must be an array of values") {
		t.Errorf("want a keys-free array note, got %v", notes)
	}
}

func TestStructuredItemShape_NestedObjectProperty(t *testing.T) {
	shape, notes := StructuredItemShape(mustUnmarshalDef(t, `{"type":"array","source":"llm",
		"items":{"type":"object","required":["name","author"],"properties":{
			"name":{"type":"string"},
			"author":{"type":"object","description":"who said it","properties":{
				"name":{"type":"string"},"role":{"type":"string"}}}}}}`))

	const want = `[{ "author": { "name": "...", "role": "..." }, "name": "..." }]`
	if shape != want {
		t.Errorf("value shape:\n got %s\nwant %s", shape, want)
	}
	if len(notes) != 1 {
		t.Fatalf("want one note, got %v", notes)
	}
	const wantNote = "`author` must be an object with `name`, `role`, never a sentence of prose — who said it."
	if notes[0] != wantNote {
		t.Errorf("note:\n got %s\nwant %s", notes[0], wantNote)
	}
}

// The skeleton is pasted into a prompt that asks for JSON and is copied by the
// model, so it must itself parse. Structural keys are quoted by strconv.Quote
// rather than by hand for this reason.
func TestStructuredItemShape_SkeletonIsAlwaysValidJSON(t *testing.T) {
	defs := []map[string]interface{}{
		mechanismFlowStepsFieldDef(),
		mustUnmarshalDef(t, `{"type":"array","source":"llm","items":{"title":"string","tags":"array"}}`),
		mustUnmarshalDef(t, `{"type":"array","source":"llm","items":{"type":"object","properties":{
			"a":{"type":"object","properties":{"b":{"type":"array","items":{"type":"object",
			"properties":{"c":{"type":"array","items":{"type":"object","properties":{"d":{"type":"string"}}}}}}}}}}}}`),
		mustUnmarshalDef(t, `{"type":"array","source":"llm","items":{"type":"object","properties":{
			"od\"d":{"type":"array","items":{"type":"object","properties":{"x":{"type":"string"}}}}}}}`),
		mustUnmarshalDef(t, `{"type":"array","source":"llm","items":{"type":"object","properties":{
			"empty":{"type":"object"}}}}`),
	}
	for i, def := range defs {
		shape, _ := StructuredItemShape(def)
		if shape == "" {
			continue
		}
		if !json.Valid([]byte(shape)) {
			t.Errorf("case %d: skeleton is not valid JSON: %s", i, shape)
		}
	}
}

// Depth is capped, and the cap must not lie about a type: at the cap a
// collection renders as its own empty form, never as `"..."` — mis-stating a
// type at depth N is the defect this file exists to remove.
func TestStructuredItemShape_DepthCapKeepsCollectionsCollections(t *testing.T) {
	shape, _ := StructuredItemShape(mustUnmarshalDef(t, `{"type":"array","source":"llm","items":{"type":"object",
		"properties":{"l1":{"type":"array","items":{"type":"object",
		"properties":{"l2":{"type":"array","items":{"type":"object",
		"properties":{"l3":{"type":"array","items":{"type":"object",
		"properties":{"l4":{"type":"string"}}}}}}}}}}}}}`))

	const want = `[{ "l1": [{ "l2": [{ "l3": [] }] }] }]`
	if shape != want {
		t.Errorf("value shape:\n got %s\nwant %s", shape, want)
	}
	if strings.Contains(shape, `"l3": "..."`) {
		t.Errorf("depth cap mis-stated an array as a string: %s", shape)
	}
}

// The renderer must agree with the gate that will judge the writer's output: a
// value shaped exactly as the skeleton demonstrates has to pass
// ContentTypeViolations. Same package, so this is cheap and it is the assertion
// that keeps the prompt and the gate from drifting apart.
func TestStructuredItemShape_DemonstratedShapePassesTheTypeGate(t *testing.T) {
	schema := map[string]interface{}{
		"fields": map[string]interface{}{"steps": mechanismFlowStepsFieldDef()},
	}
	shape, _ := StructuredItemShape(mechanismFlowStepsFieldDef())

	var steps interface{}
	if err := json.Unmarshal([]byte(strings.ReplaceAll(shape, `"..."`, `"real prose"`)), &steps); err != nil {
		t.Fatalf("skeleton did not unmarshal: %v", err)
	}
	if v := ContentTypeViolations(schema, map[string]interface{}{"steps": steps}); len(v) != 0 {
		t.Errorf("the shape the prompt demonstrates is refused by the gate: %s", DescribeTypeViolations(v))
	}

	// Control: the shape the OLD flat exemplar demonstrated must still be
	// refused, or this test would pass for a fix that changed nothing.
	bad := map[string]interface{}{"steps": []interface{}{
		map[string]interface{}{"title": "t", "branches": "a sentence of prose"},
	}}
	if v := ContentTypeViolations(schema, bad); len(v) == 0 {
		t.Error("control failed: the old string-branches shape is no longer refused, so this test proves nothing")
	}
}
