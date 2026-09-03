// FILE: platform/orchestration/datahelpers/template_context_contract_test.go
//
// These pin the DECLARATION to the BEHAVIOUR, which is the only thing that makes
// the declaration worth reading. Every test here drives ExtractFields for real
// and compares what it produced against what template_context_contract.go says
// it produces — so the contract cannot rot into a comment that used to be true.
//
// Deliberately NOT source scans. A test that greps unified_extractor.go for a
// literal passes when the literal moves into a helper, and passes vacuously when
// the needle matches the test's own comment.
package datahelpers

import (
	"sort"
	"testing"

	"go.uber.org/zap"
)

func testLogger() *zap.Logger { return zap.NewNop() }

// The three roots ExtractFields supplies whatever input_fields says. Driven with
// an EMPTY field list, so anything appearing in the result got there from
// ensureCoreFields and nowhere else.
func TestAlwaysEnsuredTemplateRootsMatchesExtractFields(t *testing.T) {
	collected := map[string]interface{}{
		"domain":    "example.co.uk",
		"objective": "build the thing",
		"model":     "claude-opus-5",
	}
	result := ExtractFields(collected, nil, testLogger())

	got := make([]string, 0, len(result))
	for k := range result {
		got = append(got, k)
	}
	want := AlwaysEnsuredTemplateRoots()
	sort.Strings(got)
	sort.Strings(want)

	if len(got) != len(want) {
		t.Fatalf("ExtractFields with no input_fields produced %v; AlwaysEnsuredTemplateRoots declares %v.\n"+
			"A root here that is not declared makes every offline check report a false finding on "+
			"templates that use it; a declared root that is not here does the opposite.", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("produced %v, declared %v", got, want)
		}
	}
}

// The single most misread line of the extractor: a dotted field is stored under
// its LAST segment. Asserted against the real behaviour, in both directions.
func TestDottedInputFieldIsStoredUnderItsLeaf(t *testing.T) {
	collected := map[string]interface{}{
		"sections_for_render": map[string]interface{}{
			"sections_ready": []interface{}{"one", "two"},
		},
	}
	result := ExtractFields(collected, []string{"sections_for_render.sections_ready"}, testLogger())

	leaf := TemplateRootForInputField("sections_for_render.sections_ready")
	if leaf != "sections_ready" {
		t.Fatalf("TemplateRootForInputField = %q, want sections_ready", leaf)
	}
	if _, ok := result[leaf]; !ok {
		t.Errorf("ExtractFields stored nothing under %q — the declared rule and the behaviour disagree", leaf)
	}
	if _, ok := result["sections_for_render"]; ok {
		t.Error("ExtractFields stored the HEAD segment too — then {{.sections_for_render.sections_ready}} " +
			"would work and the whole trap this rule exists for would not be a trap")
	}
}

// Each specially-handled field is stored under its own name, which is what lets
// TemplateRootsFor treat it like any other declared field.
func TestSpeciallyHandledFieldsAreStoredUnderTheirOwnName(t *testing.T) {
	collected := map[string]interface{}{
		"input_data":      map[string]interface{}{"domain": "example.co.uk"},
		"reviewed_brief":  map[string]interface{}{"services": "x"},
		"site_record":     map[string]interface{}{"domain": "example.co.uk"},
		"current_page":    map[string]interface{}{"title": "Home"},
		"current_section": map[string]interface{}{"name": "hero"},
	}
	for _, field := range SpeciallyHandledInputFields() {
		if !IsSpeciallyHandledInputField(field) {
			t.Fatalf("%q listed but not recognised — the two accessors disagree", field)
		}
		result := ExtractFields(collected, []string{field}, testLogger())
		if _, ok := result[field]; !ok {
			t.Errorf("ExtractFields(%q) stored nothing under that name; result keys %v", field, keysOfResult(result))
		}
	}
}

// The undecidable case, asserted rather than assumed: with input_data among the
// fields, keys of the runtime input_data map arrive at the ROOT. This is why the
// offline check must not convict a root on such a step.
func TestInputDataPromotesArbitraryKeysToTheRoot(t *testing.T) {
	collected := map[string]interface{}{
		"input_data": map[string]interface{}{
			"domain":        "example.co.uk",
			"business_name": "Acme",
		},
	}
	result := ExtractFields(collected, []string{"input_data"}, testLogger())
	if _, ok := result["business_name"]; !ok {
		t.Fatalf("business_name did not reach the root; keys %v.\n"+
			"If this ever stops being true, conditional_root findings become decidable "+
			"and should be convicted rather than reported.", keysOfResult(result))
	}

	roots, promoted := TemplateRootsFor([]string{"input_data"})
	if !promoted {
		t.Error("TemplateRootsFor must report input_data promotion — it is the whole basis of the conditional class")
	}
	if roots["business_name"] {
		t.Error("TemplateRootsFor must NOT claim to know a promoted key: it is a lower bound, not the truth")
	}
}

func TestTemplateRootsForIncludesTheAlwaysEnsuredSet(t *testing.T) {
	roots, promoted := TemplateRootsFor([]string{"current_page", "a.b.c"})
	if promoted {
		t.Error("no input_data declared, so nothing is promoted")
	}
	for _, r := range append(AlwaysEnsuredTemplateRoots(), "current_page", "c") {
		if !roots[r] {
			t.Errorf("missing root %q", r)
		}
	}
	if roots["a"] {
		t.Error("the HEAD of a dotted field is not a root")
	}
}

func keysOfResult(m map[string]interface{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
