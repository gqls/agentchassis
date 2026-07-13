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
