package actions

import (
	"strings"
	"testing"
)

// The selector crosses a process boundary — a browser measured it, and it is
// interpolated into an HTML <style> block. Only a plain tag/id/class chain may
// pass; anything that could close the block or open an at-rule must not.
func TestSafeCSSSelector(t *testing.T) {
	ok := []string{
		"div.footer-legal",
		".footer-legal",
		"#legal",
		"div#legal.footer-legal.wide",
		"footer",
		".a.b.c",
	}
	for _, s := range ok {
		if !safeCSSSelector.MatchString(s) {
			t.Errorf("legitimate selector rejected: %q", s)
		}
	}

	bad := []string{
		"",                              // nothing to aim at
		"div.footer-legal { } </style>", // closes the style block
		"div > .child",                  // combinators not supported by the fix
		".a, .b",                        // selector lists
		"@media print",                  // at-rule
		"div[data-x='y']",               // attribute selectors
		"*",                             // everything
		".a; color:red",                 // declaration injection
		"div.footer-legal\n</style>",    // newline escape
	}
	for _, s := range bad {
		if safeCSSSelector.MatchString(s) {
			t.Errorf("unsafe selector ACCEPTED: %q — this reaches a <style> block", s)
		}
	}
}

// The injected CSS must actually address the observed cause (a nowrap flex row)
// and must carry a per-selector marker so a second run is a no-op rather than a
// duplicate patch.
func TestBuildOverflowCSS(t *testing.T) {
	sel := "div.footer-legal"
	marker := overflowMarker(sel)
	css := buildOverflowCSS(sel, marker)

	if !strings.Contains(css, "flex-wrap: wrap") {
		t.Error("the live cause was display:flex with the default flex-wrap:nowrap — the fix must let it wrap")
	}
	if !strings.Contains(css, "@media (max-width: 768px)") {
		t.Error("the fix must be scoped to mobile widths, not applied to desktop")
	}
	if !strings.Contains(css, sel+" {") {
		t.Errorf("the fix must target the offender itself; got:\n%s", css)
	}
	if !strings.Contains(css, marker) {
		t.Error("the marker makes the patch idempotent — without it a re-run duplicates the block")
	}
	if strings.Contains(css, "!important") {
		t.Error("no !important needed: the block is appended after the slot's own <style>, so it wins on order")
	}
	if strings.Contains(css, ".site-header") || strings.Contains(css, ".main-nav") {
		t.Error("this must NOT be the legacy canned header CSS — that is the bug being fixed")
	}
}

// Two different offenders in one slot must not collide on the marker.
func TestOverflowMarkerIsPerSelector(t *testing.T) {
	if overflowMarker(".footer-legal") == overflowMarker(".footer-nav") {
		t.Error("markers must differ per selector, or the second fix is skipped as 'already patched'")
	}
}
