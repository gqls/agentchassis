// FILE: platform/orchestration/actions/inject_robots_noindex_test.go
//
// bugs_open/232: pages.noindex is an opt-in field (default false) read by
// assemblePage. injectRobotsNoindex itself is unconditional — the gate lives
// at the call site (`if page.Noindex`) so a reviewer of assemblePage sees the
// decision. These tests exercise the pure injection function only.
package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

const noindexTestHead = `<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>A round of The Gauntlet | Vonc</title>
    <meta name="description" content="One published round of The Gauntlet.">
    <link rel="canonical" href="https://vonc.com/tools/gauntlet/round.html">
</head>`

func TestInjectRobotsNoindexInsertsBeforeHeadClose(t *testing.T) {
	out := injectRobotsNoindex(noindexTestHead, zap.NewNop())
	if !strings.Contains(out, robotsNoindexTag) {
		t.Fatalf("robots noindex tag missing.\ngot:\n%s", out)
	}
	if strings.Index(out, robotsNoindexTag) > strings.Index(out, "</head>") {
		t.Fatalf("tag emitted outside <head>:\n%s", out)
	}
	// Comes after the canonical link already present, matching real
	// injection order (canonical is injected earlier in assemblePage).
	if strings.Index(out, `rel="canonical"`) > strings.Index(out, robotsNoindexTag) {
		t.Fatalf("expected the noindex tag after the canonical link:\n%s", out)
	}
}

func TestInjectRobotsNoindexIsIdempotent(t *testing.T) {
	logger := zap.NewNop()
	once := injectRobotsNoindex(noindexTestHead, logger)
	twice := injectRobotsNoindex(once, logger)
	if once != twice {
		t.Fatalf("second injection changed the head:\n%s", twice)
	}
	if n := strings.Count(twice, robotsNoindexTag); n != 1 {
		t.Fatalf("expected exactly one tag after two calls, got %d:\n%s", n, twice)
	}
}

func TestInjectRobotsNoindexIsCaseInsensitiveOnHeadClose(t *testing.T) {
	upper := strings.Replace(noindexTestHead, "</head>", "</HEAD>", 1)
	out := injectRobotsNoindex(upper, zap.NewNop())
	if !strings.Contains(out, robotsNoindexTag) {
		t.Fatalf("tag missing when close tag is uppercase:\n%s", out)
	}
	if strings.Index(out, robotsNoindexTag) > strings.Index(strings.ToLower(out), "</head>") {
		t.Fatalf("tag emitted after the (uppercase) head close:\n%s", out)
	}
}

func TestInjectRobotsNoindexFallsBackWhenNoHeadClose(t *testing.T) {
	fragment := `<meta charset="UTF-8"><title>No head close here</title>`
	out := injectRobotsNoindex(fragment, zap.NewNop())
	if !strings.HasPrefix(out, robotsNoindexTag) {
		t.Fatalf("expected the tag prepended when there is no </head>:\n%s", out)
	}
	if !strings.Contains(out, fragment) {
		t.Fatalf("fallback must not otherwise alter the fragment:\n%s", out)
	}
}

func TestInjectRobotsNoindexCoexistsWithAForeignPermissiveRobotsMeta(t *testing.T) {
	// A pre-existing, more permissive robots meta must not suppress ours —
	// crawlers combine multiple robots directives and take the most
	// restrictive, so coexistence still noindexes. Silently deferring to the
	// foreign tag would be the wrong result looking exactly like the right
	// one (the class this repo keeps a standing eye on).
	withForeign := strings.Replace(noindexTestHead, "<title>",
		`<meta name="robots" content="index, follow">`+"\n    <title>", 1)
	out := injectRobotsNoindex(withForeign, zap.NewNop())
	if !strings.Contains(out, `content="index, follow"`) {
		t.Fatalf("foreign robots meta was removed, not just coexisted with:\n%s", out)
	}
	if !strings.Contains(out, robotsNoindexTag) {
		t.Fatalf("our noindex tag was not added alongside the foreign one:\n%s", out)
	}
}
