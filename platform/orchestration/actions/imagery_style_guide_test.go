package actions

import (
	"strings"
	"testing"
)

func TestStyleGuideDirectionForKind(t *testing.T) {
	g := &imageryStyleGuide{
		Palette: "deep charcoal, electric blue",
		Medium:  "industrial photography, atmospheric lighting",
		Mood:    "precise, engineered",
		Avoid:   "stock-photo people",
	}

	// Photographic kinds get the full voice — palette FIRST since 2026-07-20
	// (bugs_open/027 §4b: composed last, the palette was the truncation tail
	// and silently never reached the model on over-cap directions).
	for _, kind := range []string{"hero", "illustration", "infographic", ""} {
		d := g.directionForKind(kind)
		if d != "colour palette: deep charcoal, electric blue. industrial photography, atmospheric lighting. precise, engineered" {
			t.Errorf("directionForKind(%q) = %q", kind, d)
		}
	}

	// Flat-vector kinds get palette only — photographic medium contaminates
	// flat prompts (the 2026-05-20 lesson; sprite_sheet added Phase I2,
	// content_hero bugs_open/027 §5(a): without a per-kind override it must
	// NOT fall to the photographic base voice).
	for _, kind := range []string{"icon", "sprite_sheet", "content_hero"} {
		if d := g.directionForKind(kind); d != "Colour palette: deep charcoal, electric blue" {
			t.Errorf("directionForKind(%s) = %q", kind, d)
		}
	}

	// Logos get nothing: generated once, approved, locked.
	if d := g.directionForKind("logo"); d != "" {
		t.Errorf("directionForKind(logo) = %q, want empty", d)
	}

	// Nil guide is safe and yields nothing (sites without a guide).
	var nilGuide *imageryStyleGuide
	if d := nilGuide.directionForKind("hero"); d != "" {
		t.Errorf("nil guide directionForKind(hero) = %q, want empty", d)
	}

	// Partial guides skip empty parts without stray separators.
	partial := &imageryStyleGuide{Mood: "warm, friendly"}
	if d := partial.directionForKind("hero"); d != "warm, friendly" {
		t.Errorf("partial directionForKind(hero) = %q", d)
	}
	if d := partial.directionForKind("icon"); d != "" {
		t.Errorf("partial (no palette) directionForKind(icon) = %q, want empty", d)
	}
	// content_hero with no palette and no override yields nothing (unstyled
	// but uncontaminated) rather than the photographic base voice.
	if d := partial.directionForKind("content_hero"); d != "" {
		t.Errorf("partial (no palette) directionForKind(content_hero) = %q, want empty", d)
	}
}

// TestDirectionAppliesToKind pins the flat-vector exclusion set — the gate that
// keeps the site's PHOTOGRAPHIC free-text imagery_direction (and its
// photographic reference anchors) off flat kinds. content_hero (D14) is in the
// excluded set: bugs_open/027 §1 was exactly D14 adding content_hero to
// directionForKind but not here, so an override-less site prepended its
// photographic direction to a flat kind. §5(a) fixes BOTH functions in lockstep
// — this test and the directionForKind palette-only case are the two halves.
func TestDirectionAppliesToKind(t *testing.T) {
	// Flat-vector kinds: NO photographic free-text direction / reference anchors.
	for _, kind := range []string{"icon", "logo", "sprite_sheet", "content_hero"} {
		if directionAppliesToKind(kind) {
			t.Errorf("directionAppliesToKind(%q) = true, want false (flat-vector kind)", kind)
		}
	}
	// Photographic kinds and the empty/legacy default DO take the direction.
	for _, kind := range []string{"hero", "illustration", "infographic", ""} {
		if !directionAppliesToKind(kind) {
			t.Errorf("directionAppliesToKind(%q) = false, want true (photographic kind)", kind)
		}
	}
}

// Phase I3.1 (D14): per-kind overrides replace the guide-level voice
// wholesale — direction, avoid, and reference anchors alike. Driving case:
// content_hero moved to flat duotone illustration while the site's base
// voice stays photographic; partial merging would let the photographic
// medium, the anti-illustration avoid terms, or the photographic reference
// anchors contaminate the override's visual language.
func TestStyleGuideKindOverrides(t *testing.T) {
	g := &imageryStyleGuide{
		Palette:            "deep charcoal, electric blue",
		Medium:             "industrial photography",
		Mood:               "precise, engineered",
		Avoid:              "cartoonish rendering, decorative colour",
		ReferenceAssetKeys: []string{"hero_canonical"},
		Kinds: map[string]imageryStyleGuideKindOverride{
			"content_hero": {
				Palette: "charcoal ground, electric blue shapes",
				Medium:  "flat duotone editorial illustration",
				Avoid:   "photorealism, gradients, text",
				// ReferenceAssetKeys deliberately absent: no photographic
				// anchors for a flat-illustration kind.
			},
		},
	}

	// Override composes its own direction (no Mood set → skipped cleanly);
	// palette first (bugs_open/027 §4b).
	want := "colour palette: charcoal ground, electric blue shapes. flat duotone editorial illustration"
	if d := g.directionForKind("content_hero"); d != want {
		t.Errorf("directionForKind(content_hero) = %q, want %q", d, want)
	}
	// Non-overridden kinds keep the guide-level voice.
	if d := g.directionForKind("hero"); d != "colour palette: deep charcoal, electric blue. industrial photography. precise, engineered" {
		t.Errorf("directionForKind(hero) = %q", d)
	}

	// Override avoid replaces the guide-level avoid.
	if a := g.avoidForKind("content_hero"); a != "photorealism, gradients, text" {
		t.Errorf("avoidForKind(content_hero) = %q", a)
	}
	if a := g.avoidForKind("hero"); a != "cartoonish rendering, decorative colour" {
		t.Errorf("avoidForKind(hero) = %q", a)
	}
	// An override with empty avoid still replaces (empty means none, not
	// "fall back to the contradictory base terms").
	g.Kinds["illustration"] = imageryStyleGuideKindOverride{Medium: "woodcut"}
	if a := g.avoidForKind("illustration"); a != "" {
		t.Errorf("avoidForKind(illustration, empty override) = %q, want empty", a)
	}
	// Logos get nothing even if someone writes an override for them.
	g.Kinds["logo"] = imageryStyleGuideKindOverride{Avoid: "anything"}
	if a := g.avoidForKind("logo"); a != "" {
		t.Errorf("avoidForKind(logo) = %q, want empty", a)
	}
	if d := g.directionForKind("logo"); d != "" {
		t.Errorf("directionForKind(logo with override) = %q, want empty", d)
	}

	// Reference keys: override replaces (here: with nothing); guide-level
	// keys flow for photographic kinds and stay gated off flat-vector kinds.
	if keys := g.referenceKeysForKind("content_hero"); len(keys) != 0 {
		t.Errorf("referenceKeysForKind(content_hero) = %v, want none", keys)
	}
	if keys := g.referenceKeysForKind("hero"); len(keys) != 1 || keys[0] != "hero_canonical" {
		t.Errorf("referenceKeysForKind(hero) = %v", keys)
	}
	if keys := g.referenceKeysForKind("icon"); len(keys) != 0 {
		t.Errorf("referenceKeysForKind(icon) = %v, want none (flat-vector gate)", keys)
	}
	// An override's explicit keys flow ungated, even on a flat-vector kind.
	g.Kinds["sprite_sheet"] = imageryStyleGuideKindOverride{ReferenceAssetKeys: []string{"sprite_master"}}
	if keys := g.referenceKeysForKind("sprite_sheet"); len(keys) != 1 || keys[0] != "sprite_master" {
		t.Errorf("referenceKeysForKind(sprite_sheet override) = %v", keys)
	}
	// ...but logo stays locked in ALL THREE accessors, override or not.
	g.Kinds["logo"] = imageryStyleGuideKindOverride{
		Avoid:              "anything",
		ReferenceAssetKeys: []string{"hero_canonical"},
	}
	if keys := g.referenceKeysForKind("logo"); keys != nil {
		t.Errorf("referenceKeysForKind(logo with override) = %v, want nil — logos are locked", keys)
	}

	// Nil guide is safe across all accessors.
	var nilGuide *imageryStyleGuide
	if a := nilGuide.avoidForKind("hero"); a != "" {
		t.Errorf("nil avoidForKind = %q", a)
	}
	if keys := nilGuide.referenceKeysForKind("hero"); keys != nil {
		t.Errorf("nil referenceKeysForKind = %v", keys)
	}
}

// bugs_open/011 R1 — providerForKind is the site-owned escape hatch from the
// adapter's kind routing. It mirrors avoidForKind's override-wins-even-when-
// empty contract; that subtlety is the half a later edit would break silently.
func TestProviderForKind(t *testing.T) {
	// Nil guide and unset field both mean "no opinion" — the adapter default
	// stands. This is the fleet's state today (one site has a guide at all),
	// so it is the case that must not regress.
	var nilGuide *imageryStyleGuide
	if p := nilGuide.providerForKind("hero"); p != "" {
		t.Errorf("nil providerForKind = %q, want empty", p)
	}
	plain := &imageryStyleGuide{Medium: "industrial photography"}
	if p := plain.providerForKind("hero"); p != "" {
		t.Errorf("providerForKind with no preference = %q, want empty", p)
	}

	g := &imageryStyleGuide{
		Provider: "banana",
		Kinds: map[string]imageryStyleGuideKindOverride{
			"hero":         {Provider: "stability"},
			"content_hero": {Medium: "flat duotone"}, // Provider deliberately absent
		},
	}
	// Guide-level preference applies to kinds with no override.
	if p := g.providerForKind("illustration"); p != "banana" {
		t.Errorf("providerForKind(illustration) = %q, want banana", p)
	}
	// A per-kind override replaces the guide-level value.
	if p := g.providerForKind("hero"); p != "stability" {
		t.Errorf("providerForKind(hero) = %q, want stability", p)
	}
	// An override without a Provider is a deliberate "no opinion", NOT a
	// fallthrough to the guide-level value — same rule as avoidForKind.
	if p := g.providerForKind("content_hero"); p != "" {
		t.Errorf("providerForKind(content_hero, override w/o provider) = %q, want empty", p)
	}
	// Surrounding whitespace in a hand-written spec must not become an
	// unrecognised hint at the adapter.
	spaced := &imageryStyleGuide{Provider: "  stability  "}
	if p := spaced.providerForKind("hero"); p != "stability" {
		t.Errorf("providerForKind trimmed = %q, want stability", p)
	}

	// A guide carrying ONLY a provider preference must survive the loader's
	// all-empty discard, or the field is unusable on its own.
	onlyProvider := &imageryStyleGuide{Provider: "banana"}
	if p := onlyProvider.providerForKind("hero"); p != "banana" {
		t.Errorf("provider-only guide = %q, want banana", p)
	}
}

// TestComposeImagePromptWithDirectionKeepsPaletteAndMedium drives the REAL
// composition path with robot-hands.com's verbatim content_hero override — the
// fixture that distinguishes the old first-sentence backoff from the new
// last-sentence one: its palette clause (122 chars) exceeds minKeep, so under
// strings.Index the cut landed at the palette's end and discarded the medium
// and mood that fit. bugs_open/027 §4b; council correlation 0a07f5ed.
func TestComposeImagePromptWithDirectionKeepsPaletteAndMedium(t *testing.T) {
	g := &imageryStyleGuide{Kinds: map[string]imageryStyleGuideKindOverride{
		"content_hero": {
			Medium:  "flat duotone editorial illustration",
			Mood:    "bold simple silhouette of the subject, minimal detail, precise, technical",
			Palette: "deep charcoal ground, electric blue (#0080FF) flat shapes and linework, light grey secondary accents only",
		},
	}}
	direction := g.directionForKind("content_hero")
	if len(direction) <= maxImageryDirectionInPrompt {
		t.Fatalf("fixture must exceed the cap to exercise truncation; got %d", len(direction))
	}
	got, truncated := composeImagePromptWithDirection("Article header image for a robotics guide", direction)
	if !truncated {
		t.Fatal("expected truncated=true for an over-cap direction")
	}
	for _, want := range []string{
		"#0080FF",                             // the accent that used to be invented
		"accents only",                        // the palette's TAIL — whole clause survived
		"flat duotone editorial illustration", // the medium the first-sentence backoff discarded
		"Article header image for a robotics guide", // the subject is never truncated
	} {
		if !strings.Contains(got, want) {
			t.Errorf("lost %q from the composed prompt.\ngot: %q", want, got)
		}
	}
}

// TestComposeDirectionPaletteFirst pins the survival ORDER property: the
// palette is composed first so no subsequent prefix truncation ≥ its length
// can remove it. Asserting position rather than mere presence means a future
// reordering cannot silently reintroduce bugs_open/027.
func TestComposeDirectionPaletteFirst(t *testing.T) {
	got := composeDirection("medium prose", "mood prose", "gold on charcoal")
	if !strings.HasPrefix(got, "colour palette: gold on charcoal") {
		t.Errorf("palette must lead the composed direction; got %q", got)
	}
	// Omitted fields must not leave separators behind.
	if got := composeDirection("", "", "gold on charcoal"); got != "colour palette: gold on charcoal" {
		t.Errorf("palette-only composition wrong: %q", got)
	}
}
