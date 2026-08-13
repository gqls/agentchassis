// FILE: platform/orchestration/actions/palette_ink_derivation_test.go
//
// Pins the 2026-08-13 repair to legibleInkFor: it must try to rescue the SOURCE
// colour by lightness before substituting a different palette colour.
//
// Why this test exists rather than a comment: for six days the substitution
// branch was reachable only at `text`, because `text` is first in the walk and
// `text` is by construction the slot that clears the grounds whenever anything
// does. The mechanism was council-approved, register-documented and measured as
// working — on elements that had been invisible, so "it improved" was true and
// said nothing. Only a test that names the RETURNED COLOUR can tell the two
// apart. bugs_open/122, contribution 2026-08-13.
package actions

import "testing"

// dartsonline.com's real palette, from its served stylesheet 2026-08-13.
func dartsonlineInkPalette() map[string]string {
	return map[string]string{
		"primary":    "#1A1F2E",
		"accent":     "#E8311A",
		"background": "#111520",
		"surface":    "#1E2436",
		"text":       "#F0F2F7",
		"text_muted": "#A8B0C0",
		"secondary":  "#2A3142",
		"border":     "#2A3142",
	}
}

// MUTATION TARGET: remove the colour.LegibleVariant call from legibleInkFor, or
// move it after the palette walk, and this goes RED — the walk returns `text`.
func TestLegibleInkFor_PrefersATintOfTheSourceOverSubstitution(t *testing.T) {
	p := dartsonlineInkPalette()
	grounds := []string{p["background"], p["surface"]}

	hex, source := legibleInkFor(p["primary"], grounds, p, inkMinContrast)

	if hex == p["text"] {
		t.Fatalf("legibleInkFor returned --color-text (%s) for primary %s — this is the pre-repair "+
			"behaviour the whole of bugs_open/122's 08-13 contribution is about; the ink companion "+
			"must be a tint of the brand colour, not the body text colour", hex, p["primary"])
	}
	if source != "source:lightness" {
		t.Errorf("source = %q, want %q — a substitution happened where a tint was available", source, "source:lightness")
	}
	// It must still be legible on BOTH grounds, or the repair traded one defect
	// for another.
	for _, g := range grounds {
		ratio, err := wcagContrastRatio(hex, g)
		if err != nil {
			t.Fatalf("wcagContrastRatio(%s,%s): %v", hex, g, err)
		}
		if ratio < inkMinContrast {
			t.Errorf("%s on %s = %.2f:1, below %.1f", hex, g, ratio, inkMinContrast)
		}
	}
	t.Logf("primary %s -> %s (%s); was %s before the repair", p["primary"], hex, source, p["text"])
}

// The already-legible branch must be untouched by the repair. This is the
// property that makes every existing consumer's
// var(--color-X-ink, var(--color-X)) a genuine no-op on a healthy site, and it
// is what let the mechanism ship safely in the first place.
func TestLegibleInkFor_AlreadyLegibleSourceStillReturnsUnchanged(t *testing.T) {
	p := dartsonlineInkPalette()
	grounds := []string{p["background"], p["surface"]}

	hex, source := legibleInkFor(p["text"], grounds, p, inkMinContrast)
	if hex != p["text"] || source != "source:unchanged" {
		t.Errorf("legibleInkFor(text) = (%s, %s), want (%s, source:unchanged) — the no-op branch "+
			"must survive the repair or every shipped consumer changes colour on a healthy site",
			hex, source, p["text"])
	}
}

// The palette walk is still the fallback for the case it genuinely owns: an
// achromatic source, where there is no hue to preserve.
func TestLegibleInkFor_AchromaticSourceStillFallsThroughToTheWalk(t *testing.T) {
	p := dartsonlineInkPalette()
	p["primary"] = "#1a1a1a" // achromatic, and illegible on this dark background
	grounds := []string{p["background"], p["surface"]}

	hex, source := legibleInkFor(p["primary"], grounds, p, inkMinContrast)
	if source == "source:lightness" {
		t.Errorf("an achromatic source was rescued by lightness (%s) — it has no hue to preserve, "+
			"so it belongs to the walk or the achromatic fallback", hex)
	}
	if hex == p["primary"] {
		t.Errorf("returned the illegible source unchanged (%s)", hex)
	}
}
