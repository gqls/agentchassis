// FILE: internal/adapters/imagegenerator/banana/provider_test.go
//
// bugs_open/028 — the negative-prompt fold.
//
// The defect these cover was not a crash and not a wrong value: the provider
// ACCEPTED a constraint and then dropped it, so every assertion anyone could
// make from outside ("the avoid list was set", "generation succeeded") stayed
// true while the constraint reached nothing. The tests therefore assert on the
// composed prompt text, which is the only place the honouring is observable.

package banana

import (
	"strings"
	"testing"
)

func TestFoldNegativeIntoPrompt_AppendsProhibition(t *testing.T) {
	got := foldNegativeIntoPrompt(
		"Header image for a web-based tool representing: XP Curve Designer.",
		"text, lettering, numerals, white background",
	)
	for _, want := range []string{"text", "lettering", "numerals", "white background"} {
		if !strings.Contains(got, want) {
			t.Errorf("folded prompt lost avoid term %q\ngot: %s", want, got)
		}
	}
	if !strings.Contains(got, "must not contain or use") {
		t.Errorf("folded prompt has no prohibition clause\ngot: %s", got)
	}
	if !strings.HasPrefix(got, "Header image") {
		t.Errorf("fold must preserve the positive prompt as the prefix\ngot: %s", got)
	}
}

// The positive prompt carries the brand palette (bugs_open/027 §4b). The fold
// appends, so nothing it adds may evict what is already there.
func TestFoldNegativeIntoPrompt_PreservesWholePositivePrompt(t *testing.T) {
	prompt := "flat duotone editorial illustration. colour palette: cyan #00bcd4 on near-black #121212."
	got := foldNegativeIntoPrompt(prompt, "photorealism, gradients")
	if !strings.HasPrefix(got, prompt) {
		t.Errorf("positive prompt was altered by the fold\nwant prefix: %s\ngot: %s", prompt, got)
	}
	if !strings.Contains(got, "#00bcd4") || !strings.Contains(got, "#121212") {
		t.Errorf("fold dropped palette hex codes\ngot: %s", got)
	}
}

func TestFoldNegativeIntoPrompt_EmptyNegativeIsNoop(t *testing.T) {
	prompt := "A flat illustration."
	for _, neg := range []string{"", "   ", ",", " , ; . "} {
		if got := foldNegativeIntoPrompt(prompt, neg); got != prompt {
			t.Errorf("negative %q should be a no-op, got: %s", neg, got)
		}
	}
}

// An empty positive prompt must never yield a bare constraint clause — that
// would send the model nothing but a list of things not to draw.
func TestFoldNegativeIntoPrompt_EmptyPromptStaysEmpty(t *testing.T) {
	if got := foldNegativeIntoPrompt("", "text, people"); got != "" {
		t.Errorf("empty prompt must stay empty, got: %s", got)
	}
}

func TestFoldNegativeIntoPrompt_SeparatorMatchesBoundary(t *testing.T) {
	// Already terminated → single space, no doubled full stop.
	got := foldNegativeIntoPrompt("A flat illustration.", "text")
	if strings.Contains(got, "..") {
		t.Errorf("doubled sentence boundary: %s", got)
	}
	if !strings.Contains(got, "illustration. Do not depict") {
		t.Errorf("want single-space join after boundary, got: %s", got)
	}

	// Not terminated → the fold supplies the boundary.
	got = foldNegativeIntoPrompt("A flat illustration", "text")
	if !strings.Contains(got, "illustration. Do not depict") {
		t.Errorf("want supplied sentence boundary, got: %s", got)
	}
}

// Assembled avoid lists arrive with trailing separators (the action layer
// joins the guide's `avoid` onto the kind default with ", ").
func TestFoldNegativeIntoPrompt_TrimsTrailingSeparators(t *testing.T) {
	got := foldNegativeIntoPrompt("A flat illustration.", "  text, watermark,  ")
	if !strings.HasSuffix(got, "text, watermark.") {
		t.Errorf("trailing separators not normalised, got: %s", got)
	}
}

// The real gamesdesign.co.uk content_hero avoid list, verbatim from
// site_specs on 2026-07-19 — the config whose violation exposed the bug.
func TestFoldNegativeIntoPrompt_RealWorldAvoidList(t *testing.T) {
	avoid := "photorealism, photographic texture, gradients, 3D rendering, drop shadows, " +
		"text, lettering, words, numerals, labels, captions, logos, watermarks, busy detail, " +
		"colour outside the palette, white background, pale background, bright full-bleed colour field"

	got := foldNegativeIntoPrompt("Header image for a tool.", avoid)

	// The two constraints the live images were observed violating.
	if !strings.Contains(got, "numerals") {
		t.Error("lost 'numerals' — the drop_rate_simulator violation")
	}
	if !strings.Contains(got, "white background") {
		t.Error("lost 'white background' — the xp_curve_designer violation")
	}
	if !strings.Contains(got, "bright full-bleed colour field") {
		t.Error("lost the final term — a truncating fold would cut here first")
	}
}
