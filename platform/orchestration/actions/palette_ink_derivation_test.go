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

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/colour"
	"go.uber.org/zap"
)

func zapNop() *zap.Logger { return zap.NewNop() }

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

// The overlay alpha is now read by TWO places: buildSectionDefaults emits it as a
// literal inside a format string, and buildLegibleInkDefaults composites it onto
// the page grounds. Two readers of one number, with nothing checking they agree,
// is the drift class platform/colour's own header was written about.
//
// The literal stays in the format string deliberately — changing a format string
// risks changing emitted CSS bytes for no gain — so this test is the joint.
func TestSectionSurfaceOverlayAlphaMatchesTheEmittedCSS(t *testing.T) {
	palette := map[string]string{
		"background": "#111520", "surface": "#1E2436",
		"text": "#F0F2F7", "text_muted": "#A8B0C0", "primary": "#1A1F2E",
	}
	css := buildSectionDefaults("#111520", "#1E2436", palette, true, true, zapNop())
	want := fmt.Sprintf("--section-surface: rgba(255,255,255,%g);", sectionSurfaceOverlayAlpha)
	if !strings.Contains(css, want) {
		t.Errorf("buildSectionDefaults does not emit %q.\nsectionSurfaceOverlayAlpha (%.2f) and the "+
			"emitted literal have drifted. buildLegibleInkDefaults composites the CONSTANT onto the "+
			"page grounds, so if the emitted overlay is a different alpha, every ink is certified "+
			"against a ground no visitor sees.\nemitted CSS:\n%s", want, sectionSurfaceOverlayAlpha, css)
	}
}

// WIRING, not arithmetic — and this test exists because its absence was caught by
// a mutation that PASSED.
//
// platform/colour has a test proving LegibleVariant honours composited grounds
// when it is GIVEN them. That says nothing about whether buildLegibleInkDefaults
// passes them. Deleting the compositing loop from pageGrounds left every test in
// this package green (mutation M5, 2026-08-14), which is the textbook shape: the
// unit is pinned, the wiring is not, and the mutation walks straight through the
// gap.
//
// So this reads the EMITTED :root block and measures the value it actually
// contains against the ground a visitor sees.
func TestBuildLegibleInkDefaults_EmittedInkClearsTheCompositedGround(t *testing.T) {
	// robot-hands.com, served stylesheet 2026-08-14. Its primary is the case that
	// measured 3.93:1 on the composited surface before the fix.
	palette := map[string]string{
		"primary": "#1A1F2E", "accent": "#E8500A",
		"background": "#0F1218", "surface": "#1E2535",
		"text": "#E2E8F0", "text_muted": "#A8B0C0", "secondary": "#2A3142",
	}
	css := buildLegibleInkDefaults("", palette, zapNop())
	if css == "" {
		t.Fatal("no ink block emitted")
	}

	re := regexp.MustCompile(`--color-primary-ink:\s*(#[0-9a-fA-F]{3,8})\s*;`)
	m := re.FindStringSubmatch(css)
	if m == nil {
		t.Fatalf("no --color-primary-ink in the emitted block:\n%s", css)
	}
	emitted := m[1]

	if emitted == palette["text"] {
		t.Fatalf("emitted --color-primary-ink is %s, which IS --color-text — the pre-repair behaviour", emitted)
	}

	// The ground a component painting with --section-surface actually shows.
	composited := colour.CompositeOverGround("#ffffff", sectionSurfaceOverlayAlpha, palette["surface"])
	if composited == palette["surface"] {
		t.Fatal("the composited ground equals the declared surface, so this test cannot discriminate")
	}
	ratio, err := wcagContrastRatio(emitted, composited)
	if err != nil {
		t.Fatalf("wcagContrastRatio(%s,%s): %v", emitted, composited, err)
	}
	if ratio < inkMinContrast {
		t.Errorf("emitted --color-primary-ink %s measures %.2f:1 on the COMPOSITED surface %s "+
			"(declared %s), below the %.1f floor. buildLegibleInkDefaults is certifying inks against "+
			"a ground no visitor sees — the compositing loop in pageGrounds has been removed or "+
			"broken. Measured pre-fix value was 3.93:1.",
			emitted, ratio, composited, palette["surface"], inkMinContrast)
	}
}
