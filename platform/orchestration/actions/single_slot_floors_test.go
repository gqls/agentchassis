package actions

import (
	"strings"
	"testing"
)

// buildSlotHTML makes a section carrying `classes` class attributes and `words`
// words of text, so the two quantities can be varied INDEPENDENTLY — which is
// the whole point of there being two floors.
func buildSlotHTML(classes, words int) string {
	var b strings.Builder
	b.WriteString(`<section>`)
	for i := 0; i < classes; i++ {
		b.WriteString(`<div class="card">x</div>`)
	}
	for i := 0; i < words; i++ {
		b.WriteString("word ")
	}
	b.WriteString(`</section>`)
	return b.String()
}

// The single-row path must reach the SAME verdicts as the whole-page path on the
// same content. These assert the composed decision, not the plumbing: the
// plumbing is one call and the risk is that it composes the two floors wrongly
// (e.g. checks only one of them, which is the defect that prompted the file).
func TestSingleSlotFloors_CatchesFlatteningThatKeepsTheText(t *testing.T) {
	// The bug-253 shape: text survives, layout does not.
	existing := buildSlotHTML(43, 800)
	incoming := buildSlotHTML(1, 700) // 87% of words, 2% of class attributes

	if got := countComponentClasses(existing); got != 43 {
		t.Fatalf("fixture: existing should carry 43 class attributes, got %d", got)
	}
	// The TEXT floor alone would allow this — that is the precondition.
	et := visibleTextLength(existing)
	it := visibleTextLength(incoming)
	if v := evaluateSectionShrink(defaultSectionShrinkFloor, minShrinkGuardVisibleChars,
		map[string]int{"prose-0": et}, map[string]int{"prose-0": it}); len(v) != 0 {
		t.Fatalf("precondition: the text floor must ALLOW this save, got %+v", v)
	}
	// The COMPONENT floor must refuse it.
	if v := evaluateComponentLoss(defaultSectionComponentFloor,
		map[string]int{"prose-0": countComponentClasses(existing)},
		map[string]int{"prose-0": countComponentClasses(incoming)}); len(v) != 1 {
		t.Fatal("the component floor must refuse a flattening that keeps the text")
	}
}

func TestSingleSlotFloors_CatchesTextLossThatKeepsTheLayout(t *testing.T) {
	// The mirror case, which proves the composition is not just the component
	// floor wearing a new name: layout survives, text does not.
	existing := buildSlotHTML(20, 900)
	incoming := buildSlotHTML(20, 100)

	et := visibleTextLength(existing)
	it := visibleTextLength(incoming)
	if v := evaluateSectionShrink(defaultSectionShrinkFloor, minShrinkGuardVisibleChars,
		map[string]int{"prose-0": et}, map[string]int{"prose-0": it}); len(v) != 1 {
		t.Fatal("the text floor must refuse a gutting that keeps the layout")
	}
	if v := evaluateComponentLoss(defaultSectionComponentFloor,
		map[string]int{"prose-0": countComponentClasses(existing)},
		map[string]int{"prose-0": countComponentClasses(incoming)}); len(v) != 0 {
		t.Fatal("the component floor must not fire when layout is intact")
	}
}

func TestSingleSlotFloors_AllowsAGenuineRewrite(t *testing.T) {
	// The observed GOOD rewrite: 43 -> 31 class attributes, text broadly kept.
	// Without this arm the tests would license a floor that refuses everything.
	existing := buildSlotHTML(43, 800)
	incoming := buildSlotHTML(31, 760)

	et := visibleTextLength(existing)
	it := visibleTextLength(incoming)
	if v := evaluateSectionShrink(defaultSectionShrinkFloor, minShrinkGuardVisibleChars,
		map[string]int{"prose-0": et}, map[string]int{"prose-0": it}); len(v) != 0 {
		t.Fatalf("a genuine rewrite must pass the text floor, got %+v", v)
	}
	if v := evaluateComponentLoss(defaultSectionComponentFloor,
		map[string]int{"prose-0": countComponentClasses(existing)},
		map[string]int{"prose-0": countComponentClasses(incoming)}); len(v) != 0 {
		t.Fatalf("a genuine rewrite must pass the component floor, got %+v", v)
	}
}

// A first write has no prior to compare against. Guarding it would refuse every
// initial population, and "existing is empty" must not be read as "lost 100%".
func TestSingleSlotFloors_FirstWriteHasNoPrior(t *testing.T) {
	if v := evaluateComponentLoss(defaultSectionComponentFloor,
		map[string]int{"prose-0": 0}, map[string]int{"prose-0": 0}); len(v) != 0 {
		t.Fatal("an empty existing row must not be treated as a loss")
	}
}
