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

// TestIconGroundAndTheTileItLandsInAgree is the pairing test, and what it pins
// changed once I measured.
//
// The defect is not "the tile is light" on its own — it is that TWO independent
// things decide what sits behind an icon. The generator is told, by
// flatKindPaletteClause, "flat solid <surface> background". The page paints the
// tile from a CSS variable. If those disagree the icon reads as a sticker
// pasted onto the card: the same visual failure as the pale icons, inverted,
// and invisible to any check that only asks "is the tile dark?".
//
// I first pinned this by deriving a new `icon_chip_bg` palette slot. That was
// wrong and the measurement said so: the palette reaches the stylesheet only via
// `{{palette "X"}}` calls in a layout template, and 0 of 18 layouts name that
// slot — so the derivation would have been dead config. What image cards
// actually use is `image-hover-card-grid`, whose tile reads
// `var(--color-surface-alt, var(--color-surface))`, and surface_alt IS derived.
//
// So the assertion is on the real pair: the ground the generator is told to use
// and the slot the tile is painted from must both resolve to `surface`.
func TestIconGroundAndTheTileItLandsInAgree(t *testing.T) {
	core := darkCoreOnlyPalette()
	p := fillDarkSchemeSpecialisedSlots(darkCoreOnlyPalette(), zap.NewNop())

	tile, ok := p["surface_alt"]
	if !ok || tile == "" {
		t.Fatalf("surface_alt was not derived; image cards paint their icon tile from it")
	}
	clause := flatKindPaletteClause(core)
	if clause == "" {
		t.Fatalf("flatKindPaletteClause returned nothing for a complete core palette")
	}
	if !strings.Contains(clause, tile) {
		t.Errorf("the generator is told %q but the tile is painted %s — the icon will read as a sticker on the card", clause, tile)
	}
}

// The tile is behind an ICON, so the icon's own linework has to be legible on
// it. Same bar as every other derived pair.
func TestIconLineworkClearsAAOnTheTile(t *testing.T) {
	core := darkCoreOnlyPalette()
	p := fillDarkSchemeSpecialisedSlots(darkCoreOnlyPalette(), zap.NewNop())

	ink, _ := pickInkOn(core["surface"], core)
	ratio, err := wcagContrastRatio(ink, p["surface_alt"])
	if err != nil {
		t.Fatalf("contrast: %v", err)
	}
	if ratio < 4.5 {
		t.Errorf("icon linework %s on tile %s = %.2f:1, want >= 4.5", ink, p["surface_alt"], ratio)
	}
}

// A light site keeps every literal it has today. Four live palettes rely on
// them and deriving would repaint live pages to no benefit — the boundary every
// slot here observes, and the one composedPaletteDirection deliberately copies.
func TestLightSiteGetsNoDerivationAtAll(t *testing.T) {
	light := map[string]string{
		"primary": "#1A1816", "secondary": "#4A4640", "accent": "#A8391A",
		"background": "#EFE7D6", "surface": "#F7F2E6",
		"text": "#1A1816", "text_muted": "#6B655C", "border": "#D8CEB8",
	}
	p := fillDarkSchemeSpecialisedSlots(light, zap.NewNop())
	for _, slot := range []string{"card_bg", "surface_alt", "header_bg", "cta_bg"} {
		if _, ok := p[slot]; ok {
			t.Errorf("%s was derived for a LIGHT palette; the layout literal is correct there", slot)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Legible-ink companions (bugs_open/122). Each case is chosen to fail a
// plausible WRONG implementation rather than to cover a line. Case 3 is the one
// a single-ground implementation passes everything else and fails here.
// ─────────────────────────────────────────────────────────────────────────────

// dartsonlinePalette is the real palette behind the measured failures:
// .image-hover-card-grid__eyebrow at 1.04:1 on the page and .tl-card-link at
// 1.07:1 on the card. primary sits in the background's own family.
func dartsonlinePalette() map[string]string {
	return map[string]string{
		"primary": "#111520", "secondary": "#2A3348", "accent": "#E4B95B",
		"background": "#0E1019", "surface": "#1A1F2E",
		"text": "#E8EAF0", "text_muted": "#9BA3B8", "border": "#252B3D",
	}
}

// The substitution has to be CORRECT, not merely present. A value that differs
// from the source but still fails is the same defect wearing a fix.
func TestLegibleInkFor_SubstitutesAndTheSubstituteActuallyWorks(t *testing.T) {
	p := dartsonlinePalette()
	grounds := []string{p["background"], p["surface"]}

	hex, src := legibleInkFor(p["primary"], grounds, p, inkMinContrast)
	if hex == p["primary"] {
		t.Fatalf("primary %s was left as the ink, but it scores 1.04:1 on this background", hex)
	}
	if src == "source:unchanged" {
		t.Errorf("source label says unchanged while the value changed: %s", hex)
	}
	for _, g := range grounds {
		ratio, err := wcagContrastRatio(hex, g)
		if err != nil {
			t.Fatalf("contrast(%s,%s): %v", hex, g, err)
		}
		if ratio < inkMinContrast {
			t.Errorf("substitute %s on ground %s = %.2f:1, want >= %.1f", hex, g, ratio, inkMinContrast)
		}
	}
}

// THE NO-OP CASE. A fix that repaints sites whose colours are already legible
// is a regression dressed as a win, and a gate that fires on everything is as
// useless as one that fires on nothing.
func TestLegibleInkFor_AlreadyLegibleIsLeftExactlyAlone(t *testing.T) {
	p := dartsonlinePalette()
	p["primary"] = "#E8EAF0" // now a light ink on a dark page: already fine

	hex, src := legibleInkFor(p["primary"], []string{p["background"], p["surface"]}, p, inkMinContrast)
	if hex != "#E8EAF0" {
		t.Errorf("a primary that already clears AA was changed to %s", hex)
	}
	if src != "source:unchanged" {
		t.Errorf("source = %q, want source:unchanged so the log can distinguish a no-op", src)
	}
}

// THE TWO-GROUNDS CONSTRAINT. dartsonline places the same ink on the page AND
// on a card. A single-ground implementation passes every other test in this
// file and fails here: it picks a colour that clears `background` while still
// failing `surface`, which would have read as a working fix on whichever page
// happened to be opened.
// THE FIXTURE MUST BE SATISFIABLE, and my first one was not. It used grounds
// #101010 and #E9E9E9, for which AA against BOTH is arithmetically impossible:
// the darker ground demands relative luminance >= 0.200 and the lighter one
// demands <= 0.140. Every candidate correctly fell through to the achromatic
// fallback, so the test failed while the code was right. A trap that no value
// can escape does not test preference — it tests the fallback. The grounds
// below are both dark, like dartsonline's real #0E1019 / #1A1F2E, so a
// satisfying colour exists and the CHOICE is what is under test.
func TestLegibleInkFor_TwoGroundsDisagree(t *testing.T) {
	p := map[string]string{
		"primary": "#3A3A3A", // the failing source: below AA on both
		// `text` is tried FIRST and is the trap: 4.61:1 on the background,
		// 2.06:1 on the surface. A single-ground implementation stops here.
		"text": "#7A7A7A",
		// `accent` is tried second and clears both (18.0:1 and 8.1:1).
		"accent":     "#F5F5F5",
		"text_muted": "#6A6A6A",
		"background": "#0A0A0A",
		"surface":    "#4A4A4A",
	}
	grounds := []string{p["background"], p["surface"]}

	hex, src := legibleInkFor(p["primary"], grounds, p, inkMinContrast)
	if hex == p["text"] {
		t.Fatalf("chose text %s, which clears background but scores below AA on surface %s — "+
			"this is the single-ground bug: it fixes the page and leaves the card broken",
			hex, p["surface"])
	}
	if hex != p["accent"] {
		t.Errorf("chose %s (%s); wanted the palette's own %s, which clears both grounds — "+
			"falling through to an achromatic fallback loses the site's character "+
			"when a palette colour would have done", hex, src, p["accent"])
	}
	for _, g := range grounds {
		ratio, err := wcagContrastRatio(hex, g)
		if err != nil {
			t.Fatalf("contrast(%s,%s): %v", hex, g, err)
		}
		if ratio < inkMinContrast {
			t.Errorf("chose %s (%s) but it scores %.2f:1 on ground %s", hex, src, ratio, g)
		}
	}
}

// LIGHT SCHEME. gaswholesalers.com's real values. fillDarkSchemeSpecialisedSlots
// is dark-only by deliberate design; this block is NOT, because "is this ink
// legible on this ground" does not depend on scheme and two of the three
// accent-direction sites are light. If someone later tidies this behind an
// isDarkHex guard, this test is what says why they must not.
func TestBuildLegibleInkDefaults_LightSchemeStillEmits(t *testing.T) {
	p := map[string]string{
		"primary": "#1A1A2E", "secondary": "#C8880A", "accent": "#E8A020",
		"background": "#F4F1EB", "surface": "#FFFFFF",
		"text": "#1A1A2E", "text_muted": "#5A5A6E",
	}
	css := buildLegibleInkDefaults("", p, defaultInkPolicy(), zap.NewNop())
	if css == "" {
		t.Fatal("nothing emitted for a LIGHT palette; gaswholesalers' accent-on-white " +
			"link ink (2.22:1, six on the homepage) is exactly what this must reach")
	}
	if !strings.Contains(css, "--color-accent-ink:") {
		t.Errorf("no --color-accent-ink in:\n%s", css)
	}
	if strings.Contains(css, "--color-accent-ink: "+p["accent"]+";") {
		t.Errorf("accent %s was kept as its own ink, but it scores 2.22:1 on white", p["accent"])
	}
}

// Every emitted value must be a concrete colour. An EMPTY fallback is what makes
// a CSS declaration drop entirely rather than degrade — the cascade failure mode
// the council raised on round 1 — so nothing here may emit "" or a bare var().
func TestBuildLegibleInkDefaults_NeverEmitsAnEmptyOrIndirectValue(t *testing.T) {
	css := buildLegibleInkDefaults("", dartsonlinePalette(), defaultInkPolicy(), zap.NewNop())
	if css == "" {
		t.Fatal("expected an emission for dartsonline's palette")
	}
	for _, line := range strings.Split(css, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "--color-") {
			continue
		}
		value := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(line, ":", 2)[1]), ";")
		if value == "" || strings.Contains(value, "var(") {
			t.Errorf("emitted %q — an empty or indirect value makes the consuming "+
				"declaration DROP rather than degrade", line)
		}
		if !strings.HasPrefix(value, "#") {
			t.Errorf("emitted %q, want a literal hex", line)
		}
	}
}

// Mirrors buildTokenAliases' contract: a stylesheet that already has an opinion
// keeps it, and the skip reads the ASSEMBLED css.
func TestBuildLegibleInkDefaults_SkipsNamesTheCSSAlreadyDefines(t *testing.T) {
	css := buildLegibleInkDefaults(":root{--color-primary-ink: #abcdef;}", dartsonlinePalette(), defaultInkPolicy(), zap.NewNop())
	if strings.Count(css, "--color-primary-ink:") != 0 {
		t.Errorf("redefined a name the CSS already declares:\n%s", css)
	}
	if !strings.Contains(css, "--color-accent-ink:") {
		t.Errorf("skipped the others too; only the defined name should be skipped:\n%s", css)
	}
}

// The log has to distinguish a substitution from a no-op, because the output CSS
// cannot — that indistinguishability is the property that let the defect hide.
func TestBuildLegibleInkDefaults_LogNamesSubstitutedAndUnchangedSeparately(t *testing.T) {
	core, logs := observer.New(zap.InfoLevel)
	buildLegibleInkDefaults("", dartsonlinePalette(), defaultInkPolicy(), zap.New(core))

	entries := logs.FilterMessageSnippet("legible-ink companions").All()
	if len(entries) != 1 {
		t.Fatalf("want exactly one summary line, got %d", len(entries))
	}
	fields := entries[0].ContextMap()
	if _, ok := fields["substituted"]; !ok {
		t.Error("no `substituted` field: a reader cannot tell which values were replaced")
	}
	if _, ok := fields["already_legible_left_as_is"]; !ok {
		t.Error("no `already_legible_left_as_is` field: a no-op emission would look like a repair")
	}
}

// No measurable ground must not produce a vacuous pass. A value "checked"
// against nothing reaches the stylesheet looking checked.
func TestLegibleInkFor_UnmeasurableGroundsAreNotAPass(t *testing.T) {
	p := dartsonlinePalette()
	if hex, src := legibleInkFor(p["primary"], []string{"", ""}, p, inkMinContrast); src == "source:unchanged" {
		t.Errorf("returned %s as unchanged after measuring zero grounds", hex)
	}
}

// ---------------------------------------------------------------------------
// --color-cta-bg-ink — the CTA band's own colour, re-tinted until it is legible
// as an INK on the inverted button that sits inside the band. bugs_open/398.
//
// Both palettes below are live fleet values as served 2026-08-25, not synthetic
// cases, and they were chosen because they DISAGREE: finetuning's #1e40af
// already clears the target on white and must survive untouched, while the
// default/robot-hands #3b82f6 does not and must be re-tinted. A test suite that
// only carried one of them would pass against an implementation that always
// substituted, or one that never did.
// ---------------------------------------------------------------------------

// finetuningGradientPalette is finetuning.uk's live palette. cta_bg is a
// GRADIENT, which is what produced the bug: substituted into `color:` it is
// invalid at computed-value time, so the button label inherited the band's own
// white — measured at 1.00:1 by scripts/render_audit.py on the served page.
func finetuningGradientPalette() map[string]string {
	return map[string]string{
		"primary": "#1A1A2E", "primary_text": "#ffffff",
		"background": "#F5F3EF", "surface": "#FFFFFF",
		"text": "#1A1A2E", "text_muted": "#6B6860", "accent": "#C8873A",
		"cta_bg":   "linear-gradient(135deg, #1e40af 0%, #1e3a8a 100%)",
		"cta_text": "#ffffff",
	}
}

// emittedToken pulls one companion's value out of the rendered block. Returns
// "" when the token was not emitted at all, which every caller distinguishes
// from an emitted empty value.
func emittedToken(css, name string) string {
	for _, line := range strings.Split(css, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, name+":") {
			continue
		}
		return strings.TrimSuffix(strings.TrimSpace(strings.SplitN(line, ":", 2)[1]), ";")
	}
	return ""
}

func TestBuildLegibleInkDefaults_EmitsACTAInkForAGradientCTABg(t *testing.T) {
	css := buildLegibleInkDefaults("", finetuningGradientPalette(), defaultInkPolicy(), zap.NewNop())
	got := emittedToken(css, "--color-cta-bg-ink")
	if got == "" {
		t.Fatalf("no --color-cta-bg-ink emitted for a gradient cta_bg; this is the "+
			"whole point of the slot — without it a component has no colour-typed "+
			"token to reach for and writes `color: var(--color-cta-bg)`, which is "+
			"invalid at computed-value time. Emitted:\n%s", css)
	}
	if !strings.HasPrefix(got, "#") {
		t.Errorf("--color-cta-bg-ink = %q, want a literal hex — a gradient here just "+
			"moves the original defect behind a new name", got)
	}
}

// THE LOAD-BEARING ASSERTION. The first gradient stop is a FILL colour, and
// using it directly as an ink is the obvious wrong fix: it looks right on
// finetuning (#1e40af scores 8.8:1 on white) and silently fails on the five
// other fleet themes whose stop does not clear AA. Replace the legibleInkFor
// call with the raw stop and this goes RED while every other test here stays
// green.
func TestBuildLegibleInkDefaults_CTAInkRetintsAnIllegibleFirstStop(t *testing.T) {
	p := finetuningGradientPalette()
	// default / theme-robot-hands-com / theme-loanandmortgagecalculator-co-uk.
	p["cta_bg"] = "linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)"
	css := buildLegibleInkDefaults("", p, defaultInkPolicy(), zap.NewNop())
	got := emittedToken(css, "--color-cta-bg-ink")
	if got == "" {
		t.Fatal("nothing emitted for the #3b82f6 gradient")
	}
	if strings.EqualFold(got, "#3b82f6") {
		t.Fatalf("--color-cta-bg-ink = %s, the raw first stop, which scores 3.68:1 "+
			"against the button's white face — below the 4.5 AA floor. The stop is "+
			"the SOURCE for the derivation, never the answer", got)
	}
	ratio, err := wcagContrastRatio(got, p["cta_text"])
	if err != nil {
		t.Fatalf("emitted %q is not measurable against cta_text: %v", got, err)
	}
	if ratio < inkFloorContrast {
		t.Errorf("--color-cta-bg-ink %s scores %.2f on the button face %s, want >= %.1f",
			got, ratio, p["cta_text"], inkFloorContrast)
	}
}

// The other half of the pair: a stop that ALREADY clears the target must be
// left alone. Together with the test above this pins the derivation as a real
// measurement rather than an unconditional substitution — an implementation
// that always re-tints fails here, one that never re-tints fails above.
func TestBuildLegibleInkDefaults_CTAInkKeepsALegibleFirstStop(t *testing.T) {
	p := finetuningGradientPalette()
	css := buildLegibleInkDefaults("", p, defaultInkPolicy(), zap.NewNop())
	got := emittedToken(css, "--color-cta-bg-ink")
	if !strings.EqualFold(got, "#1e40af") {
		t.Errorf("--color-cta-bg-ink = %q, want the untouched first stop #1e40af — it "+
			"already scores 8.8:1 on white, and substituting a brand colour that "+
			"needs no help is the de-branding this builder exists to avoid", got)
	}
}

// THE VACUOUS-WHITE TRAP. legibleInkFor's terminal branch compares
// worstRatioAgainst("#ffffff") with worstRatioAgainst("#000000"); with no
// measurable ground both return 0, `0 >= 0` holds, and it hands back WHITE —
// which on a white button face is the exact 1.00:1 defect being repaired.
// Remove the cta_text guard in buildLegibleInkDefaults and this goes RED.
func TestBuildLegibleInkDefaults_NoCTAInkWithoutAGround(t *testing.T) {
	p := finetuningGradientPalette()
	delete(p, "cta_text")
	css := buildLegibleInkDefaults("", p, defaultInkPolicy(), zap.NewNop())
	if got := emittedToken(css, "--color-cta-bg-ink"); got != "" {
		t.Errorf("emitted --color-cta-bg-ink = %s with no cta_text to measure against; "+
			"an ink derived from nothing is worse than an absent one, because the "+
			"consumer's var() fallback is today's working colour", got)
	}
	// The unrelated companions must still be emitted — the guard is meant to
	// skip ONE slot, not to silence the builder.
	if emittedToken(css, "--color-primary-ink") == "" {
		t.Error("the missing cta_text suppressed --color-primary-ink too; the guard is scoped wrong")
	}
}

// A stylesheet that already has an opinion keeps it — same contract as every
// other companion, and the reason it is asserted separately is that the cta
// entry is appended AFTER the wanted slice is built and could easily miss the
// shared skip.
func TestBuildLegibleInkDefaults_CTAInkRespectsAnExistingDefinition(t *testing.T) {
	pre := ":root{--color-cta-bg-ink: #123456;}"
	css := buildLegibleInkDefaults(pre, finetuningGradientPalette(), defaultInkPolicy(), zap.NewNop())
	if got := emittedToken(css, "--color-cta-bg-ink"); got != "" {
		t.Errorf("re-emitted --color-cta-bg-ink = %s over a stylesheet that already defines it", got)
	}
}

func TestFirstGradientStop(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"linear-gradient(135deg, #1e40af 0%, #1e3a8a 100%)", "#1e40af"},
		{"linear-gradient(135deg, #059669 0%, #10b981 100%)", "#059669"},
		{"linear-gradient(to right, #abc 0%, #def 100%)", "#abc"},
		// Not hex stops: the narrow reading returns "" so the caller emits
		// nothing and behaviour is exactly what it is today.
		{"linear-gradient(135deg, rgb(30,64,175) 0%, rgb(30,58,138) 100%)", ""},
		{"linear-gradient(135deg, rebeccapurple 0%, black 100%)", ""},
		{"#1e40af", "#1e40af"},
		{"", ""},
		// A malformed stop length must not be salvaged into a wrong colour.
		{"linear-gradient(135deg, #12345 0%, #fff 100%)", ""},
	} {
		if got := firstGradientStop(tc.in); got != tc.want {
			t.Errorf("firstGradientStop(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSolidCTAFill(t *testing.T) {
	if got := solidCTAFill(map[string]string{"cta_bg": "#c2410c"}); got != "#c2410c" {
		t.Errorf("a plain hex cta_bg must pass straight through, got %q", got)
	}
	if got := solidCTAFill(finetuningGradientPalette()); got != "#1e40af" {
		t.Errorf("a gradient cta_bg must reduce to its first stop, got %q", got)
	}
	if got := solidCTAFill(map[string]string{}); got != "" {
		t.Errorf("an absent cta_bg must reduce to the empty string, got %q", got)
	}
}
