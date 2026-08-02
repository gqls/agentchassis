package actions

import "testing"

// The motivating case and its boundaries (bugs_open/178). The robot-hands
// numbers are the real ones: generic-text-block went 4020→1650 stripped chars
// (57% loss) while its siblings grew by one anchor each, and the page-total
// guard passed the save.
func TestEvaluateSectionShrink(t *testing.T) {
	cases := []struct {
		name       string
		floor      float64
		existing   map[string]int
		incoming   map[string]int
		wantSlots  []string
	}{
		{
			name:  "the 178 case: one prose slot gutted, siblings grew — refused",
			floor: 0.5,
			existing: map[string]int{
				"generic-text-block": 4020, "hero": 2900, "call-to-action": 2400,
			},
			incoming: map[string]int{
				"generic-text-block": 1650, "hero": 2930, "call-to-action": 2460,
			},
			wantSlots: []string{"generic-text-block"},
		},
		{
			name:      "growth is always allowed",
			floor:     0.5,
			existing:  map[string]int{"generic-text-block": 4020},
			incoming:  map[string]int{"generic-text-block": 4110},
			wantSlots: nil,
		},
		{
			name:      "exactly at the floor passes (floor is a strict less-than)",
			floor:     0.5,
			existing:  map[string]int{"generic-text-block": 4000},
			incoming:  map[string]int{"generic-text-block": 2000},
			wantSlots: nil,
		},
		{
			name:      "a dropped slot is not a shrink — the completeness floor's case",
			floor:     0.5,
			existing:  map[string]int{"generic-text-block": 4020},
			incoming:  map[string]int{},
			wantSlots: nil,
		},
		{
			name:      "a NEW slot is out of scope",
			floor:     0.5,
			existing:  map[string]int{},
			incoming:  map[string]int{"generic-text-block": 100},
			wantSlots: nil,
		},
		{
			name:  "param-sized slots shrink legitimately — the gamesdesign hero rewrite",
			floor: 0.5,
			// 499 stripped chars is under minShrinkGuardChars: a -70% shrink
			// here was a measured IMPROVEMENT, not damage.
			existing:  map[string]int{"hero-tool": 499},
			incoming:  map[string]int{"hero-tool": 150},
			wantSlots: nil,
		},
		{
			name:      "floor 0 disables the guard",
			floor:     0,
			existing:  map[string]int{"generic-text-block": 4020},
			incoming:  map[string]int{"generic-text-block": 10},
			wantSlots: nil,
		},
		{
			name:  "an absurd floor is clamped to 0.95, not treated as 'refuse everything'",
			floor: 7.5,
			// 96% kept must PASS under the clamp; an unclamped floor of 7.5
			// would refuse every save including pure growth expressed as ratio.
			existing:  map[string]int{"generic-text-block": 1000},
			incoming:  map[string]int{"generic-text-block": 960},
			wantSlots: nil,
		},
		{
			name:  "multiple violations all reported",
			floor: 0.5,
			existing: map[string]int{
				"faq": 7000, "about-content": 3000, "hero-about": 460,
			},
			incoming: map[string]int{
				"faq": 2100, "about-content": 1400, "hero-about": 100,
			},
			// hero-about is under the size threshold; the two prose slots are
			// the vetcomparison 08-01 numbers.
			wantSlots: []string{"about-content", "faq"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := evaluateSectionShrink(tc.floor, tc.existing, tc.incoming)
			gotSlots := make(map[string]bool, len(got))
			for _, v := range got {
				gotSlots[v.Slot] = true
			}
			if len(got) != len(tc.wantSlots) {
				t.Fatalf("got %d violations (%v), want %d (%v)", len(got), gotSlots, len(tc.wantSlots), tc.wantSlots)
			}
			for _, want := range tc.wantSlots {
				if !gotSlots[want] {
					t.Fatalf("expected violation on slot %q, got %v", want, gotSlots)
				}
			}
		})
	}
}

// A violation's ratio is reported against the EXISTING length; zero existing
// must not divide by zero.
func TestSlotShrinkRatio(t *testing.T) {
	if r := (slotShrink{Existing: 4020, Incoming: 1650}).ratio(); r < 0.40 || r > 0.42 {
		t.Fatalf("ratio = %v, want ~0.41", r)
	}
	if r := (slotShrink{Existing: 0, Incoming: 10}).ratio(); r != 1 {
		t.Fatalf("zero-existing ratio = %v, want 1", r)
	}
}
