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
		// Geometry only — the bullet override for cell 5 (1,2) must carry that
		// cell's offset. Which SCOPES the rule is emitted under is the container
		// test's concern; pinning the whole selector list here just made this test
		// fail whenever a scope was added.
		"sprite-b-download::before{background-position:-40px -20px}", // cell 5 (1,2)
	}
	for _, m := range must {
		if !strings.Contains(css, m) {
			t.Errorf("sprite CSS missing %q\n---\n%s", m, css)
		}
	}
	// Default list bullet = arrow (cell 6 = 0,-40px), NOT check — user decision
	// 2026-07-15: the container opt-in themes every content list, so the fallback
	// glyph is a neutral marker; check is explicit-only. The default lives on the
	// `>li::before` rule (no sprite-b- class); the per-item check override keeps 0 0.
	if !strings.Contains(css, ">li::before,.sprite-bullets ul>li::before,.sprite-bullets ol>li::before{content:\"\";position:absolute;left:0;top:.15em;width:20px;height:20px;background-image:url(/assets/images/sprite-sheet-main.jpg);background-repeat:no-repeat;background-size:60px 60px;background-position:0px -40px}") {
		t.Errorf("default list bullet should be arrow (0px -40px):\n%s", css)
	}
	// check must still be reachable as an explicit override at cell 0.
	if !strings.Contains(css, "sprite-b-check::before,ol.sprite-list>li.sprite-b-check::before,.sprite-bullets ul>li.sprite-b-check::before,.sprite-bullets ol>li.sprite-b-check::before{background-position:0px 0px}") {
		t.Errorf("explicit sprite-b-check override (cell 0) missing:\n%s", css)
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

// I2.5: the container opt-in. Article bodies are LLM-generated HTML dropped into
// a component's {{.content}}, so their <ul>s carry no classes and never will —
// the only way to theme them is a class on a WRAPPER the template owns. These
// assertions pin the contract the article-body template depends on.
func TestBuildSpriteCSS_containerOptIn(t *testing.T) {
	names := []string{"check", "gauge", "gripper", "cog", "chart", "download", "arrow", "info", "warning"}
	css := buildSpriteCSS("/assets/images/sprite-sheet-main.jpg", 3, 3, names, "robot-hands.com")

	must := []string{
		// A plain <ul> inside the container gets bullets with no class of its own.
		".sprite-bullets ul>li::before",
		".sprite-bullets ol>li::before",
		// list-style reset must cover the container scope too, or the disc marker
		// survives alongside the glyph.
		".sprite-bullets ul",
		// Per-item overrides work inside the container as well as on ul.sprite-list.
		".sprite-bullets ul>li.sprite-b-warning::before",
	}
	for _, m := range must {
		if !strings.Contains(css, m) {
			t.Errorf("sprite CSS missing container-scope selector %q\n---\n%s", m, css)
		}
	}

	// Same specificity trap as the per-list scope: a container override
	// (.sprite-bullets ul>li.sprite-b-x::before = 0,2,3) must out-specify the
	// container default (.sprite-bullets ul>li::before = 0,1,3).
	for _, n := range names {
		if strings.Contains(css, "\nli.sprite-b-"+n+"::before") {
			t.Errorf("override %q emitted unscoped — it would lose to every default rule", n)
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
