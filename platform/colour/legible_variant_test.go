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
