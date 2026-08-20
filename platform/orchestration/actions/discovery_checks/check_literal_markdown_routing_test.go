// Tests for transformRouteSlot (bugs_open/277 §5, 2026-08-20) — migration
// 499's routing test automated. Every refusal direction must land on the
// page-rerender default; the section-editor transform route fires ONLY on the
// exact population it was built for.

package discovery_checks

import "testing"

func codeSpanFinding(slot string) literalMarkdownFinding {
	return literalMarkdownFinding{
		SlotName: slot, Pattern: "code_span", Matched: "`fetch()`", Source: "rendered_html",
	}
}

// portedPageSlots is the worked case: one un-regenerable slot (bugs_open/277
// §5.2 — template reads {{.body}}, content_data holds provenance only).
func portedPageSlots() map[string]*slotRepro {
	return map[string]*slotRepro{
		"ported-page": {occurrences: 1, canRegenerate: false},
	}
}

func TestTransformRouteSlot_FiresOnTheWorkedCase(t *testing.T) {
	slot, ok := transformRouteSlot(
		[]literalMarkdownFinding{codeSpanFinding("ported-page")}, portedPageSlots())
	if !ok || slot != "ported-page" {
		t.Fatalf("the 277 §5 population must route to the transform: got (%q, %v)", slot, ok)
	}
	// Several code_span findings on the SAME slot still route.
	slot, ok = transformRouteSlot(
		[]literalMarkdownFinding{codeSpanFinding("ported-page"), codeSpanFinding("ported-page")},
		portedPageSlots())
	if !ok || slot != "ported-page" {
		t.Fatalf("multiple code_span findings on one slot must route: got (%q, %v)", slot, ok)
	}
}

func TestTransformRouteSlot_RefusalDirections(t *testing.T) {
	base := codeSpanFinding("ported-page")

	contentDataFinding := base
	contentDataFinding.Source = "content_data"
	contentDataFinding.Field = "body"

	boldFinding := base
	boldFinding.Pattern = "bold"

	otherSlot := codeSpanFinding("hero")

	unnamed := base
	unnamed.SlotName = ""

	cases := []struct {
		name     string
		findings []literalMarkdownFinding
		slots    map[string]*slotRepro
	}{
		{"no findings", nil, portedPageSlots()},
		{"content_data finding needs the regenerate route", []literalMarkdownFinding{contentDataFinding}, portedPageSlots()},
		{"mixed surfaces", []literalMarkdownFinding{base, contentDataFinding}, portedPageSlots()},
		{"non-code_span pattern has no transform", []literalMarkdownFinding{boldFinding}, portedPageSlots()},
		{"mixed patterns", []literalMarkdownFinding{base, boldFinding}, portedPageSlots()},
		{"two slots is not one target", []literalMarkdownFinding{base, otherSlot},
			map[string]*slotRepro{
				"ported-page": {occurrences: 1, canRegenerate: false},
				"hero":        {occurrences: 1, canRegenerate: false},
			}},
		{"empty slot name cannot be resolved by the editor", []literalMarkdownFinding{unnamed},
			map[string]*slotRepro{"": {occurrences: 1, canRegenerate: false}}},
		{"duplicated slot is an ambiguous target", []literalMarkdownFinding{base},
			map[string]*slotRepro{"ported-page": {occurrences: 2, canRegenerate: false}}},
		{"regenerable component keeps the regenerate route", []literalMarkdownFinding{base},
			map[string]*slotRepro{"ported-page": {occurrences: 1, canRegenerate: true}}},
		{"slot missing from the scan metadata", []literalMarkdownFinding{base},
			map[string]*slotRepro{}},
		{"nil slot metadata", []literalMarkdownFinding{base}, nil},
	}
	for _, c := range cases {
		if slot, ok := transformRouteSlot(c.findings, c.slots); ok {
			t.Errorf("%s: must refuse (keep page-rerender), routed to %q instead", c.name, slot)
		}
	}
}
