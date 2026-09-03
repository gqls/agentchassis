package actions

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// The page-content-writer prompt, at the two sites migration 724 edits
// (bugs_open/437).
//
// WHY THESE CONSTS EXIST RATHER THAN A READ OF THE MIGRATION. A migration is
// append-only history — asserting its text is an assertion that cannot fail —
// and the live row keeps moving. So the contract is declared in platform/livespec
// and these consts embed the declared fragments in the enclosing template
// constructs they cannot be rendered without. TestWriterPromptSitesCarryTheDeclaredFragments
// pins the two together, so a change to the live contract is made once, in
// livespec, and everything else fails until it is carried through.
//
// Kept byte-identical to the live template's own spelling (read from the live
// row 2026-09-03), including the leading two spaces of the exemplar line.
const (
	writerPromptFieldListSite = "{{range .current_section.llm_field_specs}}\n" +
		"- `{{.name}}` ({{.type}}{{if .required}}, required{{else}}, optional — return \"\" if you have nothing true to put here{{end}})" +
		"{{if .description}}: {{.description}}{{end}}" +
		"{{if .item_fields}}. Each item is an object with exactly these fields: " +
		"{{range $i, $f := .item_fields}}{{if $i}}, {{end}}`{{$f}}`{{end}}{{end}}" +
		livespec.WriterPromptItemNotesTail + "\n{{end}}"

	writerPromptExemplarSite = "{\n" +
		"{{range $i, $f := .current_section.llm_field_specs}}{{if $i}},\n{{end}}  " +
		livespec.WriterPromptNestedExemplar +
		"[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}\"{{$k}}\": \"...\"{{end}} }]" +
		"{{else}}\"...\"{{end}}{{end}}\n}"

	// The pre-437 exemplar, frozen. It is the mutation control: the assertions
	// below must fail against it, or they are testing nothing.
	writerPromptExemplarSitePre437 = "{\n" +
		"{{range $i, $f := .current_section.llm_field_specs}}{{if $i}},\n{{end}}  " +
		livespec.WriterPromptFlatExemplarPre437 +
		"[{ {{range $j, $k := $f.item_fields}}{{if $j}}, {{end}}\"{{$k}}\": \"...\"{{end}} }]" +
		"{{else}}\"...\"{{end}}{{end}}\n}"
)

// mechanismFlowSpecNewGo is the field spec a chassis carrying this fix emits:
// the flat names AND the nested shape. Built as map[string]interface{} because
// that is what the prompt renderer actually receives — the spec travels
// Go struct → JSON → jsonb → generic map, which is why the template addresses
// `.item_fields` rather than `.ItemFields`.
func mechanismFlowSpecNewGo(t *testing.T) map[string]interface{} {
	t.Helper()
	fieldDef := map[string]interface{}{}
	if err := json.Unmarshal([]byte(`{
	  "type":"array","source":"llm","required":true,"minItems":2,
	  "items":{"type":"object","required":["title"],"properties":{
	    "marker":{"type":"string"},"title":{"type":"string"},"body":{"type":"string"},
	    "note":{"type":"string"},
	    "branches":{"type":"array","description":"a decision point: two or more outcomes, rendered side by side",
	      "items":{"type":"object","required":["body"],"properties":{
	        "label":{"type":"string"},"body":{"type":"string"}}}}}}}`), &fieldDef); err != nil {
		t.Fatalf("fixture: %v", err)
	}
	shape, notes := datahelpers.StructuredItemShape(fieldDef)
	if shape == "" {
		t.Fatal("fixture produced no shape — the helper and this test have drifted apart")
	}
	spec := map[string]interface{}{
		"name":        "steps",
		"type":        "array",
		"required":    true,
		"item_fields": []interface{}{"body", "branches", "marker", "note", "title"},
		"value_shape": shape,
	}
	asAny := make([]interface{}, 0, len(notes))
	for _, n := range notes {
		asAny = append(asAny, n)
	}
	spec["item_notes"] = asAny
	return spec
}

// mechanismFlowSpecOldGo is the SAME field as a chassis WITHOUT this fix emits
// it: the two new keys are absent entirely, not empty. That is the mixed deploy
// state the migration may sit in for as long as it takes the fleet to roll.
func mechanismFlowSpecOldGo() map[string]interface{} {
	return map[string]interface{}{
		"name":        "steps",
		"type":        "array",
		"required":    true,
		"item_fields": []interface{}{"body", "branches", "marker", "note", "title"},
	}
}

func renderWriterSite(t *testing.T, site string, specs ...map[string]interface{}) string {
	t.Helper()
	asAny := make([]interface{}, 0, len(specs))
	for _, s := range specs {
		asAny = append(asAny, s)
	}
	out, err := datahelpers.RenderPromptTemplate(site, map[string]interface{}{
		"current_section": map[string]interface{}{"llm_field_specs": asAny},
	}, *zap.NewNop())
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return out
}

// The fix, at the artefact the writer actually reads.
func TestWriterPromptExemplarDeclaresNestedBranchesAsAnArray(t *testing.T) {
	got := renderWriterSite(t, writerPromptExemplarSite, mechanismFlowSpecNewGo(t))

	if strings.Contains(got, `"branches": "..."`) {
		t.Errorf("the prompt still declares branches a string:\n%s", got)
	}
	if !strings.Contains(got, `"branches": [{ "body": "...", "label": "..." }]`) {
		t.Errorf("the prompt does not demonstrate the nested array:\n%s", got)
	}

	// The block is handed to a model told to reply with JSON, and the model
	// copies it. If the exemplar itself does not parse, the instruction is
	// incoherent however well-typed it is.
	var probe map[string]interface{}
	if err := json.Unmarshal([]byte(strings.ReplaceAll(got, `"..."`, `"x"`)), &probe); err != nil {
		t.Errorf("rendered exemplar is not valid JSON: %v\n%s", err, got)
	}
}

// THE MUTATION CONTROL. The same assertions, run against the frozen pre-437
// exemplar, must fail — otherwise the test above passes for reasons unrelated to
// the fix.
func TestWriterPromptPre437ExemplarWouldFailTheseAssertions(t *testing.T) {
	got := renderWriterSite(t, writerPromptExemplarSitePre437, mechanismFlowSpecNewGo(t))

	if !strings.Contains(got, `"branches": "..."`) {
		t.Error("the frozen pre-437 exemplar no longer reproduces the defect, so the assertions above prove nothing")
	}
}

// The deploy-safety property, proven rather than reasoned about: an OLD chassis
// emitting no new keys against the NEW template must render exactly what it
// rendered before the migration.
func TestWriterPromptOldChassisRendersUnchanged(t *testing.T) {
	newTemplate := renderWriterSite(t, writerPromptExemplarSite, mechanismFlowSpecOldGo())
	oldTemplate := renderWriterSite(t, writerPromptExemplarSitePre437, mechanismFlowSpecOldGo())

	if newTemplate != oldTemplate {
		t.Errorf("migration changed an un-upgraded chassis's prompt:\n new: %s\n old: %s", newTemplate, oldTemplate)
	}

	// ⚠ RenderPromptTemplate runs under text/template's DEFAULT missingkey
	// (invalid), NOT missingkey=zero: an absent key is falsy in {{if}} — which is
	// what makes the two halves order-free — but printing one bare would emit a
	// literal <no value> into the prompt. This is the assertion that would catch
	// a future edit moving either key outside its guard.
	if strings.Contains(newTemplate, "<no value>") {
		t.Errorf("absent spec keys leaked <no value> into the prompt:\n%s", newTemplate)
	}
}

// A component whose elements are scalars must be untouched in BOTH deploy
// states — that is what bounds the blast radius of this change to the one
// component that needs it (1 live, measured 2026-09-03).
func TestWriterPromptFlatComponentIsByteIdenticalInBothStates(t *testing.T) {
	flat := map[string]interface{}{
		"name": "questions", "type": "array", "required": true,
		"item_fields": []interface{}{"answer", "question"},
	}
	newTemplate := renderWriterSite(t, writerPromptExemplarSite, flat)
	oldTemplate := renderWriterSite(t, writerPromptExemplarSitePre437, flat)

	if newTemplate != oldTemplate {
		t.Errorf("a flat component's exemplar changed:\n new: %s\n old: %s", newTemplate, oldTemplate)
	}
	if !strings.Contains(newTemplate, `"questions": [{ "answer": "...", "question": "..." }]`) {
		t.Errorf("flat exemplar lost its shape:\n%s", newTemplate)
	}
}

// The notes reach the writer, carrying the nested description that the flat
// projection dropped — and nothing at all for a flat component.
func TestWriterPromptFieldListCarriesTheShapeNotes(t *testing.T) {
	got := renderWriterSite(t, writerPromptFieldListSite, mechanismFlowSpecNewGo(t))
	if !strings.Contains(got, "`branches` must be an array of objects, each with `body`, `label`") {
		t.Errorf("field list is missing the shape note:\n%s", got)
	}
	if !strings.Contains(got, "a decision point: two or more outcomes, rendered side by side") {
		t.Errorf("field list is missing the nested description:\n%s", got)
	}

	old := renderWriterSite(t, writerPromptFieldListSite, mechanismFlowSpecOldGo())
	if strings.Contains(old, "must be an array of objects") || strings.Contains(old, "<no value>") {
		t.Errorf("un-upgraded chassis rendered notes it never emitted:\n%s", old)
	}
}

// The chain that keeps three spellings in step: what livespec DECLARES the live
// row contains is what these test templates render, so a contract change made in
// livespec cannot silently leave the Go side asserting the old one. The
// migration is tied to the same fragments by its own verify block.
func TestWriterPromptSitesCarryTheDeclaredFragments(t *testing.T) {
	d := livespec.MustGet("workflow.page-content-writer.prompt_item_shape")
	both := writerPromptExemplarSite + writerPromptFieldListSite

	for _, f := range d.Fragments {
		switch {
		case f.Forbidden:
			if strings.Contains(both, f.Text) {
				t.Errorf("a site still carries the forbidden fragment %q", f.Text)
			}
		default:
			if strings.Count(both, f.Text) != 1 {
				t.Errorf("declared fragment %q appears %d times in the test sites, want 1",
					f.Text, strings.Count(both, f.Text))
			}
		}
	}
}
