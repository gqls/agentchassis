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
		"li.sprite-b-download::before{background-position:-40px -20px}", // cell 5 (1,2)
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
