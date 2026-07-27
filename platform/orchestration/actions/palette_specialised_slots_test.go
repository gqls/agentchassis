// FILE: platform/orchestration/actions/palette_specialised_slots_test.go
//
// Guards the white-card-on-a-dark-site defect: a layout declares ~17 palette
// slots with light literal fallbacks, a generated palette supplies only the 8
// core ones, and the remaining 9 therefore paint a light UI onto a dark site
// while carrying that site's light text colour.
//
// The tests assert the RESULT (a legible pair), not the mechanism, because the
// mechanism is exactly what was wrong: every individual piece — the Template
// Helper Fallback contract, the core-vs-specialised merge rule, the layout's
// literals — is correct in isolation and the composition is not. So the bar
// each test holds is a contrast ratio, computed the same way a browser does.

package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func darkCoreOnlyPalette() map[string]string {
	// The exact shape a generated per-site palette has: the 8 core slots and
	// nothing else. Values are fundamentallyai.com's, as measured.
	return map[string]string{
		"primary":    "#86ADDE",
		"secondary":  "#4A6C99",
		"accent":     "#C8902A",
		"background": "#080E1C",
		"surface":    "#111E33",
		"text":       "#E4EAF2",
		"text_muted": "#7E91A8",
		"border":     "#1B2D47",
	}
}

// TestDarkPaletteNeverInheritsLightCardLiteral is the whole bug in one
// assertion: before the fix, card_bg was absent from the merged palette, the
// layout supplied #ffffff, and the site's own --color-text (#E4EAF2) landed on
// it at 1.21:1.
func TestDarkPaletteNeverInheritsLightCardLiteral(t *testing.T) {
	p := fillDarkSchemeSpecialisedSlots(darkCoreOnlyPalette(), zap.NewNop())

	cardBG, ok := p["card_bg"]
	if !ok || cardBG == "" {
		t.Fatalf("card_bg was not derived; the layout's #ffffff literal would ship onto a dark site")
	}
	ratio, err := wcagContrastRatio(p["text"], cardBG)
	if err != nil {
		t.Fatalf("contrast: %v", err)
	}
	if ratio < 4.5 {
		t.Errorf("text %s on derived card_bg %s = %.2f:1, want >= 4.5 (AA)", p["text"], cardBG, ratio)
	}
}

// TestEveryDerivedInkClearsAAOnItsOwnFill walks the ink slots rather than
// spot-checking one, because the failure mode is per-pair: cta_text can be
// right while footer_text is wrong and the page still looks half-correct.
func TestEveryDerivedInkClearsAAOnItsOwnFill(t *testing.T) {
	p := fillDarkSchemeSpecialisedSlots(darkCoreOnlyPalette(), zap.NewNop())

	pairs := []struct{ ink, fill string }{
		{"primary_text", "primary"},
		{"secondary_text", "secondary"},
		{"header_text", "header_bg"},
		{"cta_text", "cta_bg"},
		{"footer_text", "footer_bg"},
		{"accent_text", "accent"},
	}
	for _, pair := range pairs {
		ink, okI := p[pair.ink]
		fill, okF := p[pair.fill]
		if !okI || !okF {
			t.Errorf("%s/%s: not derived (ink=%q fill=%q)", pair.ink, pair.fill, ink, fill)
			continue
		}
		ratio, err := wcagContrastRatio(ink, fill)
		if err != nil {
			t.Errorf("%s on %s: %v", pair.ink, pair.fill, err)
			continue
		}
		if ratio < 4.5 {
			t.Errorf("%s (%s) on %s (%s) = %.2f:1, want >= 4.5", pair.ink, ink, pair.fill, fill, ratio)
		}
	}
}

// TestLightPaletteIsLeftAlone pins the scope. Four live sites depend on the
// layout literals today; widening the fix to them would repaint them for no
// benefit, and this is the test that fails if someone drops the dark check.
func TestLightPaletteIsLeftAlone(t *testing.T) {
	light := map[string]string{
		"primary": "#1a365d", "secondary": "#2c5282", "accent": "#3182ce",
		"background": "#ffffff", "surface": "#f7fafc",
		"text": "#2d3748", "text_muted": "#718096", "border": "#e2e8f0",
	}
	before := len(light)
	after := fillDarkSchemeSpecialisedSlots(light, zap.NewNop())
	if len(after) != before {
		t.Errorf("light palette gained %d derived slots; it must render exactly as before", len(after)-before)
	}
}

// TestCuratedSlotSurvivesDerivation — the 15 library palettes that DO define
// these slots must render byte-identically.
func TestCuratedSlotSurvivesDerivation(t *testing.T) {
	p := darkCoreOnlyPalette()
	p["card_bg"] = "#2A1A3E" // a curated choice unrelated to `surface`
	p["cta_text"] = "#FFEEDD"

	out := fillDarkSchemeSpecialisedSlots(p, zap.NewNop())
	if out["card_bg"] != "#2A1A3E" {
		t.Errorf("curated card_bg overwritten: got %s", out["card_bg"])
	}
	if out["cta_text"] != "#FFEEDD" {
		t.Errorf("curated cta_text overwritten: got %s", out["cta_text"])
	}
}

// TestSectionDefaultsUseTheReadingColourNotTheMutedOne guards the one-line
// preference-order defect. Before the fix this emitted #7E91A8 (text_muted)
// for BOTH body and heading on every dark site, leaving #E4EAF2 unused.
func TestSectionDefaultsUseTheReadingColourNotTheMutedOne(t *testing.T) {
	p := darkCoreOnlyPalette()
	css := buildSectionDefaults(p["background"], p["surface"], p, true, true, zap.NewNop())

	if css == "" {
		t.Fatal("no section defaults emitted for a dark palette")
	}
	if strings.Contains(css, "--section-text: "+p["text_muted"]) {
		t.Errorf("body text fell back to text_muted (%s); the palette's reading colour %s clears AA",
			p["text_muted"], p["text"])
	}
	if !strings.Contains(css, "--section-text: "+p["text"]) {
		t.Errorf("expected --section-text: %s in:\n%s", p["text"], css)
	}
}

// TestSectionHeadingPrefersTheCuratedHeadingSlot — `heading` is defined by
// about half the palettes in the library and was consulted by nothing.
func TestSectionHeadingPrefersTheCuratedHeadingSlot(t *testing.T) {
	p := darkCoreOnlyPalette()
	p["heading"] = "#F2F6FA"

	css := buildSectionDefaults(p["background"], p["surface"], p, true, true, zap.NewNop())
	if !strings.Contains(css, "--section-heading: #F2F6FA") {
		t.Errorf("curated heading slot ignored; got:\n%s", css)
	}
}
