package actions

import (
	"strings"
	"testing"
)

func TestBuildSpriteCSS_geometry(t *testing.T) {
	names := []string{"check", "gauge", "gripper", "cog", "chart", "download", "arrow", "info", "warning"}
	css := buildSpriteCSS("/assets/images/sprite-sheet-main.jpg", 3, 3, names, "robot-hands.com")

	// T=20, cols=3 → sheet drawn at 60×60; cells at reading-order positions.
	must := []string{
		"background-size:60px 60px",
		".sprite-check{background-position:0px 0px}",     // cell 0 (0,0)
		".sprite-gauge{background-position:-20px 0px}",   // cell 1 (0,1)
		".sprite-gripper{background-position:-40px 0px}", // cell 2 (0,2)
		".sprite-cog{background-position:0px -20px}",     // cell 3 (1,0)
		".sprite-chart{background-position:-20px -20px}", // cell 4 (1,1)
		".sprite-warning{background-position:-40px -40px}", // cell 8 (2,2)
		"url(/assets/images/sprite-sheet-main.jpg)",
		"ul.sprite-list>li::before",
		"ul.sprite-list>li.sprite-b-download::before,ol.sprite-list>li.sprite-b-download::before{background-position:-40px -20px}", // cell 5 (1,2)
	}
	for _, m := range must {
		if !strings.Contains(css, m) {
			t.Errorf("sprite CSS missing %q\n---\n%s", m, css)
		}
	}
	// Default bullet = first cell (check @ 0 0).
	if !strings.Contains(css, "background-position:0px 0px}\nli.sprite-b-check") &&
		!strings.Contains(css, "top:.15em;width:20px;height:20px;background-image:url(/assets/images/sprite-sheet-main.jpg);background-repeat:no-repeat;background-size:60px 60px;background-position:0px 0px}") {
		t.Errorf("default bullet should be cell 0 (0 0):\n%s", css)
	}
}

// A per-item override must out-specify the default bullet rule, or every
// bullet silently renders the default glyph (caught on the live gate: all four
// wired items showed `check`). Default `ul.sprite-list>li::before` is (0,1,3);
// a bare `li.sprite-b-x::before` is only (0,1,2) and always loses, regardless
// of source order. Overrides must therefore stay scoped under the list class.
func TestBuildSpriteCSS_overridesOutspecifyDefault(t *testing.T) {
	names := []string{"check", "gauge", "gripper", "cog", "chart", "download", "arrow", "info", "warning"}
	css := buildSpriteCSS("/assets/images/sprite-sheet-main.jpg", 3, 3, names, "robot-hands.com")

	for _, n := range names {
		scoped := "ul.sprite-list>li.sprite-b-" + n + "::before"
		if !strings.Contains(css, scoped) {
			t.Errorf("override for %q must be scoped under the list class (%s)\n---\n%s", n, scoped, css)
		}
		// The bare, losing form must not be emitted as its own rule.
		if strings.Contains(css, "\nli.sprite-b-"+n+"::before") {
			t.Errorf("override for %q emitted unscoped — specificity (0,1,2) loses to the default (0,1,3)", n)
		}
	}
}

func TestSanitiseSpriteName(t *testing.T) {
	cases := map[string]string{
		"Check":          "check",
		"arrow_right":    "arrow-right",
		"bar chart":      "bar-chart",
		"  Info Circle ": "info-circle",
		"warning!":       "warning",
		"__x__":          "x",
	}
	for in, want := range cases {
		if got := sanitiseSpriteName(in); got != want {
			t.Errorf("sanitiseSpriteName(%q) = %q, want %q", in, got, want)
		}
	}
}
