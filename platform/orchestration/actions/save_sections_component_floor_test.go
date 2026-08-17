package actions

import "testing"

// The decisive test is the REAL one: the bug-253 case, at the counts measured on
// the live rows (2026-08-12). Both arms matter and a fixture with only the bad
// arm would license any floor at all — it is the GOOD rewrite passing that says
// the floor is not simply "refuse everything".
//
//	prose-0 class attributes   before 43 · flattened 1 (0.02) · good rewrite 31 (0.72)
func TestComponentFloorSeparatesTheRealFlatteningFromTheRealGoodRewrite(t *testing.T) {
	existing := map[string]int{"prose-0": 43}

	if v := evaluateComponentLoss(defaultSectionComponentFloor, existing, map[string]int{"prose-0": 1}); len(v) != 1 {
		t.Fatalf("the flattening that shipped (43→1) must be refused, got %+v", v)
	}
	if v := evaluateComponentLoss(defaultSectionComponentFloor, existing, map[string]int{"prose-0": 31}); len(v) != 0 {
		t.Fatalf("the good rewrite (43→31) must be allowed, got %+v", v)
	}
	// The pair is 35x apart, so the verdict must not depend on the exact floor.
	for _, floor := range []float64{0.25, 0.34, 0.5} {
		if len(evaluateComponentLoss(floor, existing, map[string]int{"prose-0": 1})) != 1 {
			t.Fatalf("floor %.2f failed to refuse the flattening", floor)
		}
		if len(evaluateComponentLoss(floor, existing, map[string]int{"prose-0": 31})) != 0 {
			t.Fatalf("floor %.2f falsely refused the good rewrite", floor)
		}
	}
}

func TestComponentFloorScopeAndEscapeHatches(t *testing.T) {
	// Below the scope threshold: the fleet median is 5 class attributes, so a
	// plain-prose slot going to zero must NOT be refusable.
	if v := evaluateComponentLoss(0.5, map[string]int{"p": minComponentGuardClasses - 1}, map[string]int{"p": 0}); len(v) != 0 {
		t.Fatalf("a slot below the scope threshold must be out of scope, got %+v", v)
	}
	// At the threshold it IS in scope — the boundary is asserted in both
	// directions, or "out of scope" could be silently swallowing everything.
	if v := evaluateComponentLoss(0.5, map[string]int{"p": minComponentGuardClasses}, map[string]int{"p": 0}); len(v) != 1 {
		t.Fatalf("a slot at the scope threshold must be in scope, got %+v", v)
	}
	// floor 0 disables.
	if v := evaluateComponentLoss(0, map[string]int{"p": 100}, map[string]int{"p": 0}); len(v) != 0 {
		t.Fatalf("floor 0 must disable the guard, got %+v", v)
	}
	// A slot absent from the incoming set is a DROP — the completeness floor's
	// case, not this one. Refusing it here would make two guards fight over the
	// same event and report different remedies.
	if v := evaluateComponentLoss(0.5, map[string]int{"p": 100}, map[string]int{}); len(v) != 0 {
		t.Fatalf("an absent slot is a drop, not a flattening, got %+v", v)
	}
	// Growth is always allowed.
	if v := evaluateComponentLoss(0.5, map[string]int{"p": 20}, map[string]int{"p": 60}); len(v) != 0 {
		t.Fatalf("growth must be allowed, got %+v", v)
	}
}

// The counter is the load-bearing measurement: an undercount on the EXISTING
// side silently shrinks the floor, so its tolerance of real-world spellings is
// asserted rather than assumed.
func TestCountComponentClassesTolerance(t *testing.T) {
	cases := []struct {
		name string
		html string
		want int
	}{
		{"double quotes", `<div class="card"><p class="lede">x</p></div>`, 2},
		{"single quotes", `<div class='card'></div>`, 1},
		{"spaces around =", `<div class = "card"></div>`, 1},
		{"upper case attr", `<div CLASS="card"></div>`, 1},
		{"multi-token counts once", `<div class="card card--wide mt-40"></div>`, 1},
		{"no classes", `<div><p>plain prose</p></div>`, 0},
		// The trap this guards against: a word "class=" inside TEXT would be a
		// phantom element. It needs the leading whitespace of an attribute, so
		// prose discussing a CSS class does not inflate the existing count and
		// thereby raise the bar on its own replacement.
		{"the word in prose", `<p>the class="card" convention</p>`, 1},
	}
	for _, c := range cases {
		if got := countComponentClasses(c.html); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

// Independence from the text guard is the whole point of this file existing, so
// it is asserted directly: the case that shipped kept 84% of its text.
func TestComponentFloorCatchesWhatTheTextFloorAllows(t *testing.T) {
	// Text: 3776 -> 3172 chars is 84% kept, above the 0.5 text floor.
	if v := evaluateSectionShrink(defaultSectionShrinkFloor, minShrinkGuardVisibleChars,
		map[string]int{"prose-0": 3776}, map[string]int{"prose-0": 3172}); len(v) != 0 {
		t.Fatalf("precondition: the text floor allowed this save, got %+v", v)
	}
	// Components: 43 -> 1 on the same save.
	if v := evaluateComponentLoss(defaultSectionComponentFloor,
		map[string]int{"prose-0": 43}, map[string]int{"prose-0": 1}); len(v) != 1 {
		t.Fatal("the component floor must refuse the save the text floor allowed")
	}
}
