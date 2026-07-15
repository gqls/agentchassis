package discovery_checks

import "testing"

func TestOrderedListsEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b []string
		want bool
	}{
		{"identical", []string{"hero", "cta"}, []string{"hero", "cta"}, true},
		{"different length", []string{"hero"}, []string{"hero", "cta"}, false},
		{"different order (layout matters)", []string{"hero", "cta"}, []string{"cta", "hero"}, false},
		{"different content", []string{"hero", "specs"}, []string{"hero", "cta"}, false},
		{"both empty", []string{}, []string{}, true},
		// The exact robot-hands product-detail drift that motivated this check:
		// authoritative table still held the old e-commerce layout while
		// pages.sections had been swapped to the spec sheet.
		{
			"product-detail drift",
			[]string{"product-hero", "product-specs", "call-to-action"},
			[]string{"gripper-spec-sheet", "call-to-action"},
			false,
		},
	}
	for _, tc := range cases {
		if got := orderedListsEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: orderedListsEqual(%v,%v)=%v want %v", tc.name, tc.a, tc.b, got, tc.want)
		}
	}
}
