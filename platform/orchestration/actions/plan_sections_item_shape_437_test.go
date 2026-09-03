package actions

import (
	"encoding/json"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

// The section plan is the wire between the planner and the writer's prompt, and
// `omitempty` on the two new keys is not tidiness — it IS the deploy-safety
// mechanism (bugs_open/437). A flat component must emit a spec byte-identical to
// the one it emitted before this change, because that is what guarantees its
// prompt cannot move when the migration lands ahead of the image.
func TestLLMFieldSpecOmitsShapeKeysForFlatComponents(t *testing.T) {
	flat := llmFieldSpec{
		Name: "questions", Type: "array", Required: true,
		ItemFields: []string{"answer", "question"},
	}
	got, err := json.Marshal(flat)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"name":"questions","type":"array","required":true,"item_fields":["answer","question"]}`
	if string(got) != want {
		t.Errorf("flat spec changed shape on the wire:\n got %s\nwant %s", got, want)
	}
}

// …and the nesting component must carry both, spelled in snake_case: the prompt
// template reads a JSON-round-tripped map, so the TAGS are what it addresses,
// not the Go field names.
func TestLLMFieldSpecCarriesShapeKeysForNestedComponents(t *testing.T) {
	shape, notes := datahelpers.StructuredItemShape(mechanismFlowStepsFieldDefForSpec(t))
	if shape == "" || len(notes) == 0 {
		t.Fatal("fixture produced no nested shape")
	}

	var round map[string]interface{}
	raw, err := json.Marshal(llmFieldSpec{
		Name: "steps", Type: "array", Required: true,
		ItemFields: []string{"body", "branches", "marker", "note", "title"},
		ValueShape: shape, ItemNotes: notes,
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := json.Unmarshal(raw, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"value_shape", "item_notes"} {
		if _, ok := round[key]; !ok {
			t.Errorf("spec is missing %q — the prompt addresses these tags, not the Go field names: %s", key, raw)
		}
	}
	if round["value_shape"] != shape {
		t.Errorf("value_shape did not survive the round trip: %v", round["value_shape"])
	}
}

func mechanismFlowStepsFieldDefForSpec(t *testing.T) map[string]interface{} {
	t.Helper()
	var def map[string]interface{}
	if err := json.Unmarshal([]byte(`{
	  "type":"array","source":"llm","required":true,
	  "items":{"type":"object","required":["title"],"properties":{
	    "title":{"type":"string"},
	    "branches":{"type":"array","items":{"type":"object","required":["body"],
	      "properties":{"label":{"type":"string"},"body":{"type":"string"}}}}}}}`), &def); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	return def
}
