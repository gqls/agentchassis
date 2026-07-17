package actions

import "testing"

func TestStyleGuideDirectionForKind(t *testing.T) {
	g := &imageryStyleGuide{
		Palette: "deep charcoal, electric blue",
		Medium:  "industrial photography, atmospheric lighting",
		Mood:    "precise, engineered",
		Avoid:   "stock-photo people",
	}

	// Photographic kinds get the full voice.
	for _, kind := range []string{"hero", "illustration", "infographic", ""} {
		d := g.directionForKind(kind)
		if d != "industrial photography, atmospheric lighting. precise, engineered. colour palette: deep charcoal, electric blue" {
			t.Errorf("directionForKind(%q) = %q", kind, d)
		}
	}

	// Flat-vector kinds get palette only — photographic medium contaminates
	// glyph prompts (the 2026-05-20 lesson; sprite_sheet added Phase I2).
	for _, kind := range []string{"icon", "sprite_sheet"} {
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

	// Override composes its own direction (no Mood set → skipped cleanly).
	want := "flat duotone editorial illustration. colour palette: charcoal ground, electric blue shapes"
	if d := g.directionForKind("content_hero"); d != want {
		t.Errorf("directionForKind(content_hero) = %q, want %q", d, want)
	}
	// Non-overridden kinds keep the guide-level voice.
	if d := g.directionForKind("hero"); d != "industrial photography. precise, engineered. colour palette: deep charcoal, electric blue" {
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

	// Nil guide is safe across all accessors.
	var nilGuide *imageryStyleGuide
	if a := nilGuide.avoidForKind("hero"); a != "" {
		t.Errorf("nil avoidForKind = %q", a)
	}
	if keys := nilGuide.referenceKeysForKind("hero"); keys != nil {
		t.Errorf("nil referenceKeysForKind = %v", keys)
	}
}
