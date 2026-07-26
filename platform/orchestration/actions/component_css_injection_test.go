// FILE: platform/orchestration/actions/component_css_injection_test.go
//
// Guards the placement contract of injectComponentCSS (bugs_open/072).
//
// The ordering is the load-bearing part, not the presence. A site's
// assets/css/styles.css is frozen at its last webdesign-agent run and may
// hold an OLDER copy of the same css_snippet; the injected block only wins
// that tie if it sits AFTER the stylesheet <link> in document order. Placing
// it before </head> is what guarantees that, so a test that only asserted
// "the block is somewhere in the head" would pass while the fix silently
// stopped working.

package actions

import (
	"strings"
	"testing"
)

const testCSSBlock = `<style ` + componentCSSMarker + `>
.news-card { color: red; }
</style>
`

func TestInjectComponentCSS_LandsAfterTheStylesheetLink(t *testing.T) {
	head := `<head>
    <title>x</title>
    <link rel="stylesheet" href="/assets/css/styles.css">
</head>`

	got := injectComponentCSS(head, testCSSBlock)

	linkAt := strings.Index(got, `href="/assets/css/styles.css"`)
	blockAt := strings.Index(got, componentCSSMarker)
	closeAt := strings.Index(got, "</head>")

	if linkAt < 0 || blockAt < 0 || closeAt < 0 {
		t.Fatalf("expected link, block and </head> all present, got:\n%s", got)
	}
	if blockAt < linkAt {
		t.Errorf("component CSS must follow the stylesheet link, or a stale styles.css wins the cascade tie")
	}
	if blockAt > closeAt {
		t.Errorf("component CSS escaped the head")
	}
}

func TestInjectComponentCSS_IsIdempotent(t *testing.T) {
	head := "<head><link rel=\"stylesheet\" href=\"/assets/css/styles.css\"></head>"

	once := injectComponentCSS(head, testCSSBlock)
	twice := injectComponentCSS(once, testCSSBlock)

	if once != twice {
		t.Errorf("second injection changed the head; blocks would stack on a re-run:\n%s", twice)
	}
	if n := strings.Count(twice, componentCSSMarker); n != 1 {
		t.Errorf("expected exactly 1 injected block, got %d", n)
	}
}

func TestInjectComponentCSS_UppercaseHeadTag(t *testing.T) {
	// Stored head components are hand-authored; their casing is not guaranteed.
	head := "<HEAD><TITLE>x</TITLE></HEAD>"

	got := injectComponentCSS(head, testCSSBlock)

	if !strings.Contains(got, componentCSSMarker) {
		t.Fatalf("block not injected into an uppercase head: %s", got)
	}
	if strings.Index(got, componentCSSMarker) > strings.Index(strings.ToUpper(got), "</HEAD>") {
		t.Errorf("block landed outside the head")
	}
}

func TestInjectComponentCSS_NoHeadCloseStillShips(t *testing.T) {
	// A truncated or hand-written head must not silently swallow the CSS.
	head := "<head><title>x</title>"

	got := injectComponentCSS(head, testCSSBlock)

	if !strings.Contains(got, componentCSSMarker) {
		t.Errorf("component CSS was dropped when </head> was absent")
	}
}

func TestInjectComponentCSS_EmptyBlockIsANoOp(t *testing.T) {
	head := "<head><title>x</title></head>"

	if got := injectComponentCSS(head, ""); got != head {
		t.Errorf("empty block must leave the head untouched, got: %s", got)
	}
}
