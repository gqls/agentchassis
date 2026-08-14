// FILE: platform/colour/legible_variant_test.go
//
// Tests for the hue-preserving lightness search added 2026-08-13 to repair the
// defect recorded in bugs_open/122: --color-<x>-ink resolved to --color-text on
// every site in the fleet, so repointing a consumer onto it de-branded the
// element rather than making its brand colour legible.
//
// The fixtures are REAL fleet palettes, read from the served stylesheets on
// 2026-08-13, because the defect was invisible to synthetic fixtures for six
// days. A fixture that cannot reproduce the original failure cannot prove the
// repair.
package colour

import (
	"math"
	"testing"
)

// dartsonline.com, the site the owner reported: primary is nearly its own
// surface, so as an ink on the page it measured 1.11:1 live.
const (
	dartsPrimary    = "#1A1F2E"
	dartsBackground = "#111520"
	dartsSurface    = "#1E2436"
	dartsText       = "#F0F2F7"
)

func dartsGrounds() []string { return []string{dartsBackground, dartsSurface} }

func hueOf(t *testing.T, hex string) float64 {
	t.Helper()
	r, g, b, err := ParseHex(hex)
	if err != nil {
		t.Fatalf("ParseHex(%q): %v", hex, err)
	}
	h, _, _ := rgbToHSL(r, g, b)
	return h
}

func satOf(t *testing.T, hex string) float64 {
	t.Helper()
	r, g, b, err := ParseHex(hex)
	if err != nil {
		t.Fatalf("ParseHex(%q): %v", hex, err)
	}
	_, s, _ := rgbToHSL(r, g, b)
	return s
}

// The headline property: the returned colour is still the brand colour. This is
// the whole point of the repair — the previous behaviour returned --color-text,
// which is legible and unrecognisable.
func TestLegibleVariant_PreservesHueAndSaturationExactly(t *testing.T) {
	got, ok := LegibleVariant(dartsPrimary, dartsGrounds(), AANormal)
	if !ok {
		t.Fatal("no legible variant found for dartsonline's primary; the repair cannot work at all")
	}
	// Hue and saturation must survive the move. Allow a small tolerance only for
	// 8-bit rounding on the way back out of HSL — not for a design change.
	if d := math.Abs(hueOf(t, got) - hueOf(t, dartsPrimary)); d > 2.0 {
		t.Errorf("hue moved by %.2f deg (%s -> %s): the point of this function is that it does not", d, dartsPrimary, got)
	}
	if d := math.Abs(satOf(t, got) - satOf(t, dartsPrimary)); d > 0.06 {
		t.Errorf("saturation moved by %.3f (%s -> %s): a desaturated rescue is the blend behaviour this deliberately avoids", d, dartsPrimary, got)
	}
	if got == dartsText {
		t.Errorf("returned %s, which IS --color-text — this is precisely the defect being repaired", got)
	}
}

// It must actually be legible, on EVERY ground, not just the one a reviewer
// happens to open.
func TestLegibleVariant_ClearsAAOnEveryGround(t *testing.T) {
	got, ok := LegibleVariant(dartsPrimary, dartsGrounds(), AANormal)
	if !ok {
		t.Fatal("expected a variant")
	}
	for _, ground := range dartsGrounds() {
		ratio, err := ContrastRatio(got, ground)
		if err != nil {
			t.Fatalf("ContrastRatio(%s, %s): %v", got, ground, err)
		}
		if ratio < AANormal {
			t.Errorf("%s on ground %s = %.2f:1, below the %.1f floor", got, ground, ratio, AANormal)
		}
	}
}

// MUTATION TARGET. If the outward walk is replaced by "try one direction to
// exhaustion first", this fails: the source is dark, so a downward walk reaches
// #000000-ish before ever testing the lightening that actually wins.
func TestLegibleVariant_ReturnsTheSmallestSufficientChange(t *testing.T) {
	got, ok := LegibleVariant(dartsPrimary, dartsGrounds(), AANormal)
	if !ok {
		t.Fatal("expected a variant")
	}
	r, g, b, _ := ParseHex(got)
	_, _, gotL := rgbToHSL(r, g, b)
	r0, g0, b0, _ := ParseHex(dartsPrimary)
	_, _, srcL := rgbToHSL(r0, g0, b0)

	// Walking one step less far must NOT be legible — otherwise the search
	// overshot and returned a colour further from the brand than necessary.
	dir := 1.0
	if gotL < srcL {
		dir = -1.0
	}
	h, s, _ := rgbToHSL(r0, g0, b0)
	oneStepLess := ToHex(hslToRGB(h, s, gotL-dir*lightnessSearchStep))
	for _, ground := range dartsGrounds() {
		if ratio, err := ContrastRatio(oneStepLess, ground); err == nil && ratio < AANormal {
			return // correct: one step less is not legible, so `got` is minimal
		}
	}
	t.Errorf("%s is legible one step short of the returned %s — the search overshot the brand colour", oneStepLess, got)
}

// An unmeasurable grounds slice must NOT produce a confident answer. A pass
// derived from measuring nothing is how an untested value reaches a stylesheet
// looking tested — the same rule legibleInkFor's clearsAll already follows.
func TestLegibleVariant_UnmeasurableGroundsAreNotAPass(t *testing.T) {
	for _, grounds := range [][]string{nil, {}, {""}, {"not-a-colour"}} {
		if got, ok := LegibleVariant(dartsPrimary, grounds, AANormal); ok {
			t.Errorf("grounds %v returned %s, ok=true — nothing was measurable, so nothing was checked", grounds, got)
		}
	}
}

// An achromatic source has no hue to preserve, so this function must decline and
// leave the case to the caller's achromatic fallback. Two mechanisms both
// claiming one case is how they drift.
func TestLegibleVariant_DeclinesAnAchromaticSource(t *testing.T) {
	for _, grey := range []string{"#000000", "#ffffff", "#808080", "#1a1a1a"} {
		if got, ok := LegibleVariant(grey, dartsGrounds(), AANormal); ok {
			t.Errorf("LegibleVariant(%s) = %s, ok=true — an achromatic source should fall through", grey, got)
		}
	}
}

// Real light-site palettes, because the original defect's worst consequence was
// on LIGHT sites: a brand accent became near-black there. Each of these three
// resolved to --color-text before the repair.
func TestLegibleVariant_LightSitePalettesKeepTheirBrandColour(t *testing.T) {
	cases := []struct {
		site, src, text string
		grounds         []string
	}{
		{"cookly.uk", "#C8502A", "#2C2C27", []string{"#FAF7F2", "#FFFFFF"}},
		{"webdesign.co.uk", "#d4a373", "#2b2b2b", []string{"#FDFCFA", "#FFFFFF"}},
		{"lendzy.co.uk", "#E8700A", "#1A1A1A", []string{"#FFFFFF", "#F7F9FC"}},
	}
	for _, c := range cases {
		got, ok := LegibleVariant(c.src, c.grounds, AANormal)
		if !ok {
			t.Errorf("%s: no variant for %s", c.site, c.src)
			continue
		}
		if got == c.text {
			t.Errorf("%s: returned --color-text (%s) — the pre-repair behaviour", c.site, got)
		}
		if d := math.Abs(hueOf(t, got) - hueOf(t, c.src)); d > 2.0 {
			t.Errorf("%s: hue moved %.2f deg (%s -> %s)", c.site, d, c.src, got)
		}
		t.Logf("%-18s %s -> %s (was %s)", c.site, c.src, got, c.text)
	}
}

// HSL round-trip integrity. If this drifts, every figure above is measuring the
// wrong colour and would still look plausible.
func TestHSLRoundTripsWithinRounding(t *testing.T) {
	for _, hex := range []string{"#1A1F2E", "#E8311A", "#d4a373", "#00bcd4", "#7c3cff", "#F0F2F7", "#010203"} {
		r, g, b, err := ParseHex(hex)
		if err != nil {
			t.Fatalf("ParseHex(%q): %v", hex, err)
		}
		h, s, l := rgbToHSL(r, g, b)
		r2, g2, b2 := hslToRGB(h, s, l)
		if absDiff(r, r2) > 1 || absDiff(g, g2) > 1 || absDiff(b, b2) > 1 {
			t.Errorf("%s round-tripped to %s (hsl %.1f %.3f %.3f)", hex, ToHex(r2, g2, b2), h, s, l)
		}
	}
}

func absDiff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

// ============================================================================
// Composited grounds — added 2026-08-14 after the bugfix_122_contrast_ink_slots
// lane raised the objection this section exists to answer.
// ============================================================================

// sectionOverlay composites the renderer's own --section-surface overlay, the way
// buildLegibleInkDefaults now does.
func sectionOverlay(ground string) string {
	return CompositeOverGround("#ffffff", 0.05, ground)
}

// THE OBJECTION, PINNED. Certifying an ink against the DECLARED surface while the
// visitor sees the COMPOSITED one certifies it against the wrong colour. Because
// the search returns the smallest sufficient change, every output sits within
// ~0.1 of the floor by construction, so a 0.62 shift is ~10x the margin and flips
// the element back into a filed failure.
//
// MEASURED 2026-08-14, robot-hands' real palette: against {bg, surface} the search
// emitted #7d8bb6, which measures 3.93:1 on the composited surface — a FAIL, on an
// element migration 368 had repaired. Against all four grounds it emits #8a97bd at
// 4.56:1 worst-case. This test fails if the composited grounds are dropped.
func TestLegibleVariant_ClearsTheCOMPOSITEDGroundNotJustTheDeclaredOne(t *testing.T) {
	// robot-hands.com, read from its served stylesheet 2026-08-14.
	const bg, surface = "#0F1218", "#1E2535"
	grounds := []string{bg, surface, sectionOverlay(bg), sectionOverlay(surface)}

	got, ok := LegibleVariant("#1A1F2E", grounds, AANormal)
	if !ok {
		t.Fatal("no variant against the composited grounds")
	}
	for _, g := range grounds {
		ratio, err := ContrastRatio(got, g)
		if err != nil {
			t.Fatalf("ContrastRatio(%s,%s): %v", got, g, err)
		}
		if ratio < AANormal {
			t.Errorf("%s on ground %s = %.2f:1 — below the floor on a ground the renderer itself paints", got, g, ratio)
		}
	}

	// And the discriminator: the two-ground answer must NOT satisfy four grounds.
	// If this ever stops holding, the composited grounds have become a no-op and
	// this test would pass while asserting nothing.
	twoGround, ok := LegibleVariant("#1A1F2E", []string{bg, surface}, AANormal)
	if !ok {
		t.Fatal("no two-ground variant")
	}
	if twoGround == got {
		t.Fatalf("two-ground and four-ground searches both returned %s — the composited grounds "+
			"are not changing the answer, so this test is vacuous and the objection is unpinned", got)
	}
	if r, err := ContrastRatio(twoGround, sectionOverlay(surface)); err == nil && r >= AANormal {
		t.Errorf("the two-ground answer %s already clears the composited surface at %.2f:1 — "+
			"the fixture no longer reproduces the reported failure", twoGround, r)
	}
}

// The emitted hex, pinned for real palettes. This is the test the reviewing lane
// asked for, and it exists because I quoted #7785b2 for robot-hands from a probe
// whose GROUNDS I had invented (#0F1319/#1A1F2E instead of the served
// #0F1218/#1E2535). The code was right and my figure was wrong, and no test could
// contradict me because none named an output.
//
// A fixture whose inputs are transcribed from the artefact, asserting an exact
// output, is the only shape that catches that.
func TestLegibleVariant_EmittedHexIsPinnedForRealPalettes(t *testing.T) {
	// EVERY ground here is transcribed from the served stylesheet. That sentence is
	// load-bearing: the first version of this table invented background/surface for
	// three sites, and TWO of the resulting figures were wrong and were published —
	// in a commit message, a bug file, a handoff and a council submission — before
	// a reviewer's independent replication disagreed. See the note below the table.
	cases := []struct {
		site, src, bg, surface, want string
	}{
		{"robot-hands.com primary", "#1A1F2E", "#0F1218", "#1E2535", "#8a97bd"},
		{"dartsonline.com primary", "#1A1F2E", "#111520", "#1E2436", "#8a97bd"},
		{"webdesign.co.uk accent", "#d4a373", "#f9f8f6", "#ffffff", "#9d6630"},
		{"vonc.com primary", "#7c3cff", "#0a0a0f", "#13121f", "#9b6aff"},

		// ⚠ THE THINNEST EMISSION IN THE FLEET, and the reason it is pinned rather
		// than padded. oufe.com sets primary == surface (both #1B2A3B), so the ink
		// is being made legible against its own colour and lands at 4.51 — +0.01.
		// A reviewer proposed walking to 5.0 for absorption; that was declined
		// because it buys a cushion without fixing a wrong ground and would imply
		// cover for grounds nobody models. What +0.01 argues for is a PIN, not a
		// cushion: at that margin this is the case most likely to flip silently
		// under a refactor, and two independent implementations of this algorithm
		// have already disagreed twice in two days.
		{"oufe.com primary (fleet-thinnest, +0.01)", "#1B2A3B", "#0F1820", "#1B2A3B", "#7d9ec4"},

		// The two that a reviewer's replication caught. Both were quoted wrong from
		// invented grounds: cookly as #c04d28 (real surface is the cream #F0E8D5,
		// not the white I assumed) and lendzy as #b75808 with bg/surface swapped.
		{"cookly.uk accent", "#C8502A", "#FDFAF4", "#F0E8D5", "#af4625"},
		{"lendzy.co.uk accent", "#E8700A", "#F8F7F4", "#FFFFFF", "#b25608"},
	}
	for _, c := range cases {
		grounds := []string{c.bg, c.surface}
		for _, g := range []string{c.bg, c.surface} {
			if o := sectionOverlay(g); o != g {
				grounds = append(grounds, o)
			}
		}
		got, ok := LegibleVariant(c.src, grounds, AANormal)
		if !ok {
			t.Errorf("%s: no variant", c.site)
			continue
		}
		if got != c.want {
			t.Errorf("%s: LegibleVariant(%s) = %s, want %s — if this is a deliberate change, "+
				"re-measure every figure quoted in bugs_open/122, the lane NOTES, the handoff and "+
				"the council submission, because they all name these hexes", c.site, c.src, got, c.want)
		}
	}
}
