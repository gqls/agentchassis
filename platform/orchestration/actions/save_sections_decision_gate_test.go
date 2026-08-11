// FILE: platform/orchestration/actions/save_sections_decision_gate_test.go
//
// Guards for the rebuild-door decision gate (RFC_015 §5b, owner ruling
// 2026-08-10). The first test is the important one and it is not about the
// feature at all — it is about not breaking every rebuild on every site.
package actions

import (
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TestDecisionProtectedIDArrayLiteral_EmptyIsAnEmptyArrayNotNull is the
// fleet-safety test.
//
// The rebuild's DELETE reads `AND NOT (id = ANY($2::uuid[]))`. Postgres three-
// valued logic makes the empty case dangerous in a way the populated case is not:
// if the parameter arrives NULL, `id = ANY(NULL)` is NULL, `NOT NULL` is NULL, the
// WHERE matches NOTHING, and save_page_sections stops clearing old rows — on every
// page of every site, silently, because the save still reports success and simply
// inserts the new composition alongside the old. That is a fleet-wide page
// duplication caused by a feature that is inert on 13 of 14 sites.
//
// Almost every save fleet-wide takes this path (no decision records ⇒ no protected
// rows), so the empty case is not an edge case here — it is the normal one.
func TestDecisionProtectedIDArrayLiteral_EmptyIsAnEmptyArrayNotNull(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []*decisionProtectedRow
	}{
		{"nil slice", nil},
		{"empty slice", []*decisionProtectedRow{}},
	} {
		got := decisionProtectedIDArrayLiteral(tc.in)
		if got != "{}" {
			t.Fatalf("%s: got %q, want %q — anything else risks a NULL or malformed "+
				"array parameter, and a NULL one makes the rebuild's DELETE match no rows at all",
				tc.name, got, "{}")
		}
	}
}

// TestDecisionProtectedIDArrayLiteral_RendersEveryID pins the populated shape:
// brace-wrapped, comma-separated, no spaces, every id present.
func TestDecisionProtectedIDArrayLiteral_RendersEveryID(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	got := decisionProtectedIDArrayLiteral([]*decisionProtectedRow{
		{id: a, slot: "hero"},
		{id: b, slot: "tool-list"},
	})

	if !strings.HasPrefix(got, "{") || !strings.HasSuffix(got, "}") {
		t.Fatalf("not a Postgres array literal: %q", got)
	}
	if strings.Contains(got, " ") {
		t.Fatalf("array literal contains a space, which uuid[] parsing will reject: %q", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if !strings.Contains(got, id.String()) {
			t.Fatalf("id %s missing from %q — an id that does not reach the exclusion list "+
				"gets DELETEd, which is the protected content being destroyed", id, got)
		}
	}
	if want := "{" + a.String() + "," + b.String() + "}"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestMatchDecisionProtectedRow_ExactThenNormalised mirrors matchLockedRow's
// contract: exact slot match first, then kebab-normalised (the 041 naming
// landmine — the library stores kebab-case while older rows and plans may carry
// snake_case or CamelCase spellings of one slot).
func TestMatchDecisionProtectedRow_ExactThenNormalised(t *testing.T) {
	rows := []*decisionProtectedRow{
		{id: uuid.New(), slot: "brief-explanation"},
		{id: uuid.New(), slot: "tool-list"},
	}

	if got := matchDecisionProtectedRow(rows, "tool-list", ""); got == nil || got.slot != "tool-list" {
		t.Fatalf("exact match failed: %+v", got)
	}
	// A different spelling of the same slot must still be caught, or a rebuild
	// naming it snake_case walks straight past the protection.
	if got := matchDecisionProtectedRow(rows, "tool_list", ""); got == nil || got.slot != "tool-list" {
		t.Fatalf("normalised match failed for tool_list: %+v", got)
	}
	if got := matchDecisionProtectedRow(rows, "", ""); got != nil {
		t.Fatalf("empty section name must never match, got %+v", got)
	}
	if got := matchDecisionProtectedRow(rows, "call-to-action", ""); got != nil {
		t.Fatalf("uncovered slot must not match, got %+v", got)
	}
}

// TestMatchDecisionProtectedRow_ConsumedRowMatchesOnce: each protected row blocks
// at most one incoming section, so a page carrying a duplicated slot name cannot
// have one decision swallow several sections (same property as matchLockedRow).
func TestMatchDecisionProtectedRow_ConsumedRowMatchesOnce(t *testing.T) {
	rows := []*decisionProtectedRow{{id: uuid.New(), slot: "hero"}}

	first := matchDecisionProtectedRow(rows, "hero", "")
	if first == nil {
		t.Fatal("first match should succeed")
	}
	first.consumed = true

	if second := matchDecisionProtectedRow(rows, "hero", ""); second != nil {
		t.Fatalf("a consumed row matched a second section: %+v", second)
	}
}

// TestReadDecisionCitation_FindsNestedAndPoolsBothVerbs: the rebuild envelope is
// not the edit seam's, so the citation has to be found where a work item's spec
// actually lands. Both verbs pool, because either one NAMES the decision and that
// is the whole test.
func TestReadDecisionCitation_FindsNestedAndPoolsBothVerbs(t *testing.T) {
	collected := map[string]interface{}{
		"input_data": map[string]interface{}{
			"spec": map[string]interface{}{
				"acknowledges_decision": "D-001-free-beside-paid",
			},
		},
	}
	got := readDecisionCitation(collected, nil)
	if !strings.Contains(got, "D-001-free-beside-paid") {
		t.Fatalf("citation not found at input_data.spec.acknowledges_decision, got %q", got)
	}

	// supersedes reads the same way
	collected2 := map[string]interface{}{
		"spec": map[string]interface{}{"supersedes_decision": "D-004-guide-copy-hand-authored"},
	}
	if got := readDecisionCitation(collected2, nil); !strings.Contains(got, "D-004-guide-copy-hand-authored") {
		t.Fatalf("supersedes citation not found, got %q", got)
	}

	// An envelope with no citation must produce the empty string, NOT something
	// CitationSatisfies could mistake for a name.
	if got := readDecisionCitation(map[string]interface{}{"input_data": map[string]interface{}{}}, nil); got != "" {
		t.Fatalf("expected no citation, got %q", got)
	}
}

// TestReadDecisionCitation_ConfigOverride: a caller whose envelope puts the
// citation somewhere else can say so, the same way this action's other field
// lookups are config-overridable.
func TestReadDecisionCitation_ConfigOverride(t *testing.T) {
	collected := map[string]interface{}{
		"work_item": map[string]interface{}{"cites": "D-002-no-tools-directory-on-index"},
	}
	config := map[string]interface{}{"decision_citation_field": "work_item.cites"}

	if got := readDecisionCitation(collected, config); !strings.Contains(got, "D-002-no-tools-directory-on-index") {
		t.Fatalf("config-specified path not honoured, got %q", got)
	}
}

// TestCoveredKeySliceAndCoveredKeysAgree: the two renderings of "the covering
// decision keys" must not drift — the rebuild gate uses the slice as data, the
// refusal messages use the joined string as prose, and they are one definition.
func TestCoveredKeySliceAndCoveredKeysAgree(t *testing.T) {
	covered := []DecisionCoverage{{Key: "D-001"}, {Key: "D-002"}}

	slice := CoveredKeySlice(covered)
	if len(slice) != 2 || slice[0] != "D-001" || slice[1] != "D-002" {
		t.Fatalf("CoveredKeySlice wrong: %#v", slice)
	}
	if joined := CoveredKeys(covered); joined != strings.Join(slice, ", ") {
		t.Fatalf("CoveredKeys %q does not match the slice %#v", joined, slice)
	}
}

// TestMatchDecisionProtectedRow_ComponentIDBeatsARenamedSlot is the guard against
// the gate DUPLICATING what it means to protect.
//
// bugs_open/189: extractSectionsFromMetadata prefers component_function over
// component_name once a component resolves, so a positionally-named stored slot
// ("tool-2") never matches the incoming resolved name
// ("tool-loan-vs-savings"). On a name-only match the fresh copy is INSERTED while
// the protected row — excluded from the DELETE by this very gate — survives
// beside it: same component_id twice on one page, every step reporting success.
//
// Not armed today (the 14 positionally-named sections are on loancalculator.co.uk
// and oufe.com, neither of which has decision records), which is precisely why it
// is worth pinning now rather than after it fires.
func TestMatchDecisionProtectedRow_ComponentIDBeatsARenamedSlot(t *testing.T) {
	compID := uuid.New().String()
	rows := []*decisionProtectedRow{
		{id: uuid.New(), componentID: compID, slot: "tool-2"},
	}

	// The incoming section resolved to a different NAME but is the same COMPONENT.
	got := matchDecisionProtectedRow(rows, "tool-loan-vs-savings", compID)
	if got == nil {
		t.Fatal("a protected row was not matched by component_id when the slot name " +
			"had been resolved to a different spelling — the fresh copy would be INSERTED " +
			"alongside the protected row, duplicating the section this gate exists to protect")
	}
	if got.slot != "tool-2" {
		t.Fatalf("matched the wrong row: %+v", got)
	}
}

// TestMatchDecisionProtectedRow_EmptyComponentIDDoesNotPairEverything: the
// metadata path often arrives with no component_id, so an empty id must not act
// as a wildcard that pairs an unresolved section with the first idless protected
// row.
func TestMatchDecisionProtectedRow_EmptyComponentIDDoesNotPairEverything(t *testing.T) {
	rows := []*decisionProtectedRow{
		{id: uuid.New(), componentID: "", slot: "brief-explanation"},
	}

	if got := matchDecisionProtectedRow(rows, "call-to-action", ""); got != nil {
		t.Fatalf("an empty component_id matched an unrelated slot: %+v", got)
	}
	// The name path still works when ids are absent on both sides.
	if got := matchDecisionProtectedRow(rows, "brief-explanation", ""); got == nil {
		t.Fatal("name matching regressed when component_id is empty on both sides")
	}
}
