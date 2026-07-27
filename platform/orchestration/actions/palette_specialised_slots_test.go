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
	"fmt"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
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

// --- council f17b0a77 round 1, bug_historian (gating) ------------------------
//
// "a silent no-op fallback path sits right next to the loud one this plan is
// adding, in the same function, and nothing distinguishes 'derived' from 'fell
// through to the light literal' in the output CSS". Both tests below assert the
// SIGNAL, not the colour: the whole defect class is things that fail by looking
// like nothing happened.

// TestUndeliverableDerivationIsLoud — when the core slot a derivation reads is
// itself missing, the layout's literal ships after all. That is the same
// outcome as doing nothing and must not be silent.
func TestUndeliverableDerivationIsLoud(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)

	// A dark palette with no `surface`: card_bg, cta_bg, background_alt and
	// surface_alt all derive from it and none can be filled.
	p := map[string]string{
		"background": "#080E1C", "text": "#E4EAF2", "text_muted": "#7E91A8",
		"primary": "#86ADDE", "secondary": "#4A6C99", "accent": "#C8902A",
	}
	fillDarkSchemeSpecialisedSlots(p, zap.New(core))

	if _, ok := p["card_bg"]; ok {
		t.Fatal("card_bg was derived from a slot that does not exist")
	}
	found := false
	for _, e := range logs.All() {
		if strings.Contains(e.Message, "could NOT derive") {
			found = true
			for _, f := range e.Context {
				if f.Key == "undeliverable" {
					if !strings.Contains(fmt.Sprint(f.Interface), "card_bg<-surface") {
						t.Errorf("warning does not name the undeliverable slot: %v", f.Interface)
					}
				}
			}
		}
	}
	if !found {
		t.Error("a derivation was skipped and nothing warned — indistinguishable from success")
	}
}

// TestUncoveredLightLiteralIsNamed — the derivation table covers the nine slots
// all 18 layouts share, but `{{palette "X" "<literal>"}}` is generic and the
// fleet declares 60+ further slot names. Those cannot be derived safely (nobody
// can say what `badge_deal` should be on a dark site), so they must at least be
// visible.
func TestUncoveredLightLiteralIsNamed(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	layout := `:root {
	  --color-card-bg:  {{palette "card_bg"  "#ffffff"}};
	  --color-badge-bg: {{palette "badge_bg" "#f7fafc"}};
	  --color-code-bg:  {{palette "code_bg"  "#0b1020"}};
	  --color-accent:   {{palette "accent"   "#3182ce"}};
	}`
	p := darkCoreOnlyPalette()
	fillDarkSchemeSpecialisedSlots(p, zap.NewNop()) // card_bg now covered
	warnLightLiteralsOnDarkSite(layout, p, zap.New(core))

	var named string
	for _, e := range logs.All() {
		for _, f := range e.Context {
			if f.Key == "light_literals" {
				named = fmt.Sprint(f.Interface)
			}
		}
	}
	if !strings.Contains(named, "badge_bg=#f7fafc") {
		t.Errorf("an uncovered LIGHT literal was not named: %q", named)
	}
	if strings.Contains(named, "code_bg") {
		t.Errorf("a DARK literal is not a problem on a dark site and must not be reported: %q", named)
	}
	if strings.Contains(named, "accent") {
		t.Errorf("a slot the palette supplies has an unreachable literal and must not be reported: %q", named)
	}
	if strings.Contains(named, "card_bg") {
		t.Errorf("a slot the derivation covers must not also be reported: %q", named)
	}
}

// TestLightSiteIsNotWarnedAboutItsOwnLiterals — the whole point of the light
// literals is that they are correct on a light site.
func TestLightSiteIsNotWarnedAboutItsOwnLiterals(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	layout := `--color-badge-bg: {{palette "badge_bg" "#f7fafc"}};`
	warnLightLiteralsOnDarkSite(layout, map[string]string{"background": "#ffffff"}, zap.New(core))
	if logs.Len() != 0 {
		t.Errorf("a light site was warned about literals chosen for it: %v", logs.All())
	}
}

// TestUnparsablePaletteCallIsReported — council f17b0a77 round 2, raised by
// bug_historian AND guardian independently: the scanner knows one spelling of
// the helper call, so a layout written differently would match nothing and
// report nothing, which is the same silent fall-through the round is closing.
// Verified separately that all 18 active layouts parse cleanly today; this is
// the guard for when that stops being true.
func TestUnparsablePaletteCallIsReported(t *testing.T) {
	core, logs := observer.New(zap.WarnLevel)
	// Second call uses single quotes — a shape the strict pattern rejects.
	layout := `--a: {{palette "card_bg" "#ffffff"}}; --b: {{palette 'badge_bg' '#f7fafc'}};`
	warnLightLiteralsOnDarkSite(layout, map[string]string{"background": "#080E1C"}, zap.New(core))

	var reported bool
	for _, e := range logs.All() {
		if strings.Contains(e.Message, "cannot parse") {
			reported = true
			for _, f := range e.Context {
				if f.Key == "unparsed" && fmt.Sprint(f.Integer) != "1" {
					t.Errorf("expected 1 unparsed call, got %v", f.Integer)
				}
			}
		}
	}
	if !reported {
		t.Error("a palette declaration the scanner could not parse was passed over in silence")
	}
}
