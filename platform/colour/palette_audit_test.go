// FILE: platform/colour/palette_audit_test.go
//
// The fixtures are not invented. They are fundamentallyai.com's real palette
// before and after the 2026-07-27 repair, so this test is a regression corpus
// for a defect that actually shipped rather than a demonstration that the
// arithmetic works.
//
// The pair that matters most is in TestAuditPaletteCatchesTheRepairRegression:
// repairing the palette FIXED five pairings and BROKE a sixth, because
// --color-primary flipped from near-black to light blue and every ink over it
// inverted. A checker that only ran after the repair would have called that a
// success.

package colour

import (
	"math"
	"testing"
)

// The live palette as it shipped, and as the owner saw it on his phone.
var brokenPalette = PaletteSlots{
	"background": "#080E1C", "text": "#E4EAF2", "text_muted": "#7E91A8",
	"card_bg": "#ffffff", "surface": "#111E33", "border": "#1B2D47",
	"primary": "#0E1B2E", "secondary": "#1A2E48", "accent": "#C8902A",
	"primary_text": "#ffffff", "cta_bg": "#1a365d", "cta_text": "#ffffff",
}

// The same site after the stylesheet was regenerated with the pinned palette
// and derived specialised slots.
var repairedPalette = PaletteSlots{
	"background": "#080E1C", "text": "#E4EAF2", "text_muted": "#7E91A8",
	"card_bg": "#132239", "surface": "#111E33", "border": "#1B2D47",
	"primary": "#86ADDE", "secondary": "#4A6C99", "accent": "#C8902A",
	"primary_text": "#071019", "cta_bg": "#101E33", "cta_text": "#E8EDF3",
}

func failuresBySlots(f []Failure) map[string]float64 {
	m := make(map[string]float64, len(f))
	for _, x := range f {
		m[x.Pair.ForegroundSlot+"/"+x.Pair.BackgroundSlot] = x.Ratio
	}
	return m
}

func TestAuditPaletteCatchesTheShippedDefect(t *testing.T) {
	got := failuresBySlots(AuditPalette(brokenPalette))

	// The three the owner could see, and their measured ratios.
	want := map[string]float64{
		"text/card_bg":       1.21, // "Every decision leaves a record"
		"text_muted/card_bg": 3.23,
		"primary/background": 1.11, // section eyebrows
	}
	for k, w := range want {
		g, ok := got[k]
		if !ok {
			t.Errorf("%s should have FAILED and did not — this is the defect that shipped", k)
			continue
		}
		if math.Abs(g-w) > 0.02 {
			t.Errorf("%s ratio = %.2f, want ≈%.2f", k, g, w)
		}
	}

	// The control: the pair that was FINE must not be reported. A checker that
	// flags everything on a broken palette tells you nothing about which part
	// is broken, and this pair is why the failure was confusing to diagnose —
	// section text was perfectly readable while card text was invisible.
	if r, bad := got["text_muted/background"]; bad {
		t.Errorf("text_muted/background reported as failing at %.2f:1, but it measures 5.97:1 and passed", r)
	}
	if r, bad := got["text/background"]; bad {
		t.Errorf("text/background reported as failing at %.2f:1, but it passed throughout", r)
	}
}

func TestAuditPaletteCatchesTheRepairRegression(t *testing.T) {
	before := failuresBySlots(AuditPalette(brokenPalette))
	after := failuresBySlots(AuditPalette(repairedPalette))

	// What the repair fixed.
	for _, k := range []string{"text/card_bg", "text_muted/card_bg", "primary/background"} {
		if r, still := after[k]; still {
			t.Errorf("%s still fails after the repair at %.2f:1", k, r)
		}
		if _, was := before[k]; !was {
			t.Errorf("fixture error: %s was supposed to fail BEFORE the repair", k)
		}
	}

	// What the repair BROKE, and the whole reason this pair is on the list:
	// white ink over a primary that used to be near-black. Measured 2.32:1 on
	// the live calculator page.
	//
	// NOTE the repaired palette derives primary_text = #071019, so it PASSES
	// here — the palette was right and two component templates hard-coded
	// #fff instead of reading the slot. That is family 3 and this checker
	// cannot see it; the assertion below pins what this checker DOES prove,
	// which is that the palette itself offers a legible ink.
	if r, bad := after["primary_text/primary"]; bad {
		t.Errorf("the repaired palette offers no legible ink on its own primary (%.2f:1) — "+
			"that is a palette defect, not a component one", r)
	}

	// And the inverse, which is the regression as it would look if the derived
	// ink had NOT been fixed: white on the new light primary.
	unfixed := PaletteSlots{}
	for k, v := range repairedPalette {
		unfixed[k] = v
	}
	unfixed["primary_text"] = "#ffffff"
	regressed := failuresBySlots(AuditPalette(unfixed))
	r, caught := regressed["primary_text/primary"]
	if !caught {
		t.Fatal("white ink on the light primary must be caught — this is the regression a palette repair introduces")
	}
	if math.Abs(r-2.32) > 0.02 {
		t.Errorf("primary_text/primary ratio = %.2f, want ≈2.32 (the measured live value)", r)
	}
}

func TestAuditPaletteRefusesToPassAnUnusablePalette(t *testing.T) {
	// A missing REQUIRED slot must fail, not skip. "No failures" on a palette
	// with no background is the vacuous green this package exists to prevent.
	if f := AuditPalette(PaletteSlots{"text": "#000000"}); len(f) == 0 {
		t.Error("a palette with no background reported zero failures")
	}
	// A missing OPTIONAL slot is skipped: a site with no cards is not broken.
	full := PaletteSlots{"background": "#000000", "text": "#ffffff", "text_muted": "#cccccc",
		"primary": "#88bbff", "accent": "#ffcc00"}
	if f := AuditPalette(full); len(f) != 0 {
		t.Errorf("a sound palette with no card_bg reported %d failure(s): %+v", len(f), f)
	}
	// An unparseable value is a failure, not a skip — in practice a var()
	// reference or a gradient reached a slot that must hold a literal.
	bad := PaletteSlots{"background": "var(--something)", "text": "#ffffff", "text_muted": "#cccccc",
		"primary": "#88bbff", "accent": "#ffcc00"}
	if f := AuditPalette(bad); len(f) == 0 {
		t.Error("an unparseable background reported zero failures")
	}
}

func TestContrastRatioAnchors(t *testing.T) {
	// Two values that can be checked by hand against the WCAG spec, so a
	// refactor of the maths cannot drift unnoticed.
	for _, tc := range []struct {
		a, b string
		want float64
	}{
		{"#000000", "#ffffff", 21.0},
		{"#ffffff", "#ffffff", 1.0},
	} {
		got, err := ContrastRatio(tc.a, tc.b)
		if err != nil {
			t.Fatalf("ContrastRatio(%s,%s): %v", tc.a, tc.b, err)
		}
		if math.Abs(got-tc.want) > 0.001 {
			t.Errorf("ContrastRatio(%s,%s) = %.3f, want %.1f", tc.a, tc.b, got, tc.want)
		}
	}
	if _, err := ContrastRatio("#nothex", "#ffffff"); err == nil {
		t.Error("an unparseable colour must error rather than return a confident ratio")
	}
}
