package actions

// bugs_open/345, owner ruling 2026-08-25 (decision 2b): an orphan-only
// schema/template mismatch DROPS the unrendered field(s) and stores, instead of
// refusing the whole component. These tests pin the decision at the level the
// gate decides it — classifySyncIssues (block vs drop) and dropSchemaFields (the
// drop itself) — without standing up a DB, because the gate's choice is a pure
// function of the score and the schema.
//
// Mutation map (each guard fails its own test, and only that one):
//   - block the orphan-only branch again (append to blockingIssues)     -> TestOrphanOnlyIsDroppable* semantics captured here via classify
//   - drop unknown-vars instead of blocking                             -> TestUnknownVarStillBlocks
//   - dropSchemaFields deletes a non-named field / keeps a named one     -> TestDropSchemaFieldsRemovesExactlyTheNamed
//   - dropSchemaFields discards a sibling top-level key                  -> TestDropSchemaFieldsPreservesOtherSchemaKeys

import (
	"encoding/json"
	"testing"
)

// A template that renders {{.gamma}} against a schema declaring alpha/beta/gamma:
// alpha and beta are ORPHANS (declared, not rendered); no unknown vars.
const orphanOnlySchema = `{"fields":{"alpha":{"source":"site_specs.x"},"beta":{"source":"site_specs.y"},"gamma":{"source":"site_specs.z"}}}`
const orphanOnlyTemplate = `<section data-component="x"><div>{{.gamma}}</div></section>`

// A template that renders an UNDECLARED {{.delta}} — a real defect (empty slot).
const unknownVarSchema = `{"fields":{"gamma":{"source":"site_specs.z"}}}`
const unknownVarTemplate = `<section data-component="x"><div>{{.gamma}} {{.delta}}</div></section>`

func TestUnknownVarStillBlocks(t *testing.T) {
	score := scoreComponent("", "x", unknownVarTemplate, unknownVarSchema, "section")
	if score.SchemaTemplateSynced {
		t.Fatal("fixture defect: this pair must be out of sync")
	}
	_, unknownVars, _ := classifySyncIssues(score.QualityIssues)
	if len(unknownVars) == 0 {
		t.Fatal("an undeclared {{.delta}} must classify as an unknown template var — the gate BLOCKS this class; misclassifying it would silently store a component with an empty slot")
	}
	// delta named
	found := false
	for _, v := range unknownVars {
		if v == "delta" {
			found = true
		}
	}
	if !found {
		t.Fatalf("unknown var not named: %v", unknownVars)
	}
}

func TestOrphanOnlyClassifiesAsDroppable(t *testing.T) {
	score := scoreComponent("", "x", orphanOnlyTemplate, orphanOnlySchema, "section")
	if score.SchemaTemplateSynced {
		t.Fatal("fixture defect: this pair must be out of sync")
	}
	orphans, unknownVars, _ := classifySyncIssues(score.QualityIssues)
	if len(unknownVars) != 0 {
		t.Fatalf("orphan-only fixture produced unknown vars %v — the gate would BLOCK instead of drop", unknownVars)
	}
	// Both alpha and beta orphaned (exhaustive, not first-hit).
	if len(orphans) != 2 {
		t.Fatalf("want 2 orphans (alpha, beta), got %v — a first-hit classifier would drop only one and leave the other to re-refuse", orphans)
	}
}

func TestDropSchemaFieldsRemovesExactlyTheNamed(t *testing.T) {
	reduced, err := dropSchemaFields(orphanOnlySchema, []string{"alpha", "beta"})
	if err != nil {
		t.Fatalf("dropSchemaFields: %v", err)
	}
	set := schemaFieldSet(reduced)
	if set["alpha"] || set["beta"] {
		t.Fatalf("named orphans survived the drop: %v", set)
	}
	if !set["gamma"] {
		t.Fatalf("the rendered field gamma was dropped — the drop must remove ONLY the named orphans, else it strands the content the template actually uses")
	}
}

func TestDropSchemaFieldsPreservesOtherSchemaKeys(t *testing.T) {
	// A schema with a sibling top-level key beside .fields must keep it.
	in := `{"fields":{"a":{"source":"s"},"b":{"source":"s"}},"render_mode":"resolved","version":2}`
	reduced, err := dropSchemaFields(in, []string{"b"})
	if err != nil {
		t.Fatalf("dropSchemaFields: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(reduced), &m); err != nil {
		t.Fatalf("reduced schema does not parse: %v", err)
	}
	if m["render_mode"] != "resolved" {
		t.Fatalf("sibling key render_mode lost: %v", m)
	}
	if _, ok := m["version"]; !ok {
		t.Fatalf("sibling key version lost: %v", m)
	}
	if set := schemaFieldSet(reduced); set["b"] || !set["a"] {
		t.Fatalf("wrong field set after drop: %v", set)
	}
}

func TestDropSchemaFieldsRejectsUnparseable(t *testing.T) {
	// The gate falls back to BLOCKING if the drop can't be done, so the helper
	// must error (not silently return the input) on bad JSON.
	if _, err := dropSchemaFields(`{not json`, []string{"a"}); err == nil {
		t.Fatal("dropSchemaFields must error on unparseable JSON so the gate can fall back to blocking rather than store an unreduced schema")
	}
}

// The incumbent-orphan rule (regeneration): an orphan that is ALSO an existing
// field is NOT dropped — dropping it would remove a name a live page's
// content_data may be keyed on. This pins the set arithmetic the gate does.
func TestIncumbentOrphanIsNotDroppable(t *testing.T) {
	score := scoreComponent("", "x", orphanOnlyTemplate, orphanOnlySchema, "section")
	orphans, _, _ := classifySyncIssues(score.QualityIssues)          // {alpha, beta}
	incumbent := schemaFieldSet(`{"fields":{"beta":{"source":"s"}}}`) // beta already exists on live pages
	droppable := make([]string, 0, len(orphans))
	for _, f := range orphans {
		if !incumbent[f] {
			droppable = append(droppable, f)
		}
	}
	// alpha is new -> droppable; beta is incumbent -> kept.
	if len(droppable) != 1 || droppable[0] != "alpha" {
		t.Fatalf("want to drop only the NEW orphan alpha, keeping incumbent beta; got %v — dropping beta would strand its content and collide with the stranding guard", droppable)
	}
}
