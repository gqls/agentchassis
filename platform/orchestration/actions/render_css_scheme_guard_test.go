package actions

// FILE: platform/orchestration/actions/render_css_scheme_guard_test.go
//
// Unit tests for enforceLayoutScheme (bugs_open/022): a layout's declared
// colour scheme must overrule a design_spec background that contradicts
// it. The reproduction case is the real incident — a light #F4F5F7
// background merged onto robot-hands.com's scheme=dark layout.

import (
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// Observer captures Info and above: the guard's unjudgeable paths emit
// Info, its restore path emits Warn, and its refusal path returns an
// error without logging.
func observedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func TestEnforceLayoutScheme_DarkLayoutRejectsLightBackground(t *testing.T) {
	// The shipped incident: tool-portal-dark (scheme=dark) + spec
	// background #F4F5F7. Theme background AND text must both come back.
	theme := map[string]string{
		"background": "#0f172a",
		"text":       "#e2e8f0",
	}
	merged := map[string]string{
		"background": "#F4F5F7",
		"text":       "#1e293b",
	}
	logger, logs := observedLogger()

	if err := enforceLayoutScheme("dark", theme, merged, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged["background"] != "#0f172a" {
		t.Errorf("background: got %q, want theme's %q", merged["background"], "#0f172a")
	}
	if merged["text"] != "#e2e8f0" {
		t.Errorf("text: got %q, want theme's %q (never half-swap)", merged["text"], "#e2e8f0")
	}
	if logs.Len() != 1 {
		t.Fatalf("expected exactly one Warn, got %d", logs.Len())
	}
	entry := logs.All()[0]
	if entry.Level != zap.WarnLevel {
		t.Errorf("restore log must be Warn, got %v", entry.Level)
	}
	fields := entry.ContextMap()
	if fields["rejected_background"] != "#F4F5F7" {
		t.Errorf("Warn must name the rejected value, got %v", fields["rejected_background"])
	}
	if fields["kept_background"] != "#0f172a" {
		t.Errorf("Warn must name the kept value, got %v", fields["kept_background"])
	}
}

func TestEnforceLayoutScheme_LightLayoutRejectsDarkBackground(t *testing.T) {
	theme := map[string]string{
		"background": "#fdfcf9",
		"text":       "#2d2a26",
	}
	merged := map[string]string{
		"background": "#111827",
		"text":       "#f9fafb",
	}
	logger, logs := observedLogger()

	if err := enforceLayoutScheme("light", theme, merged, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged["background"] != "#fdfcf9" {
		t.Errorf("background: got %q, want theme's %q", merged["background"], "#fdfcf9")
	}
	if merged["text"] != "#2d2a26" {
		t.Errorf("text: got %q, want theme's %q", merged["text"], "#2d2a26")
	}
	if logs.Len() != 1 {
		t.Fatalf("expected exactly one Warn, got %d", logs.Len())
	}
}

func TestEnforceLayoutScheme_ConformingBackgroundUntouched(t *testing.T) {
	theme := map[string]string{"background": "#0f172a", "text": "#e2e8f0"}
	merged := map[string]string{"background": "#1a2332", "text": "#f0f4f8"}
	logger, logs := observedLogger()

	if err := enforceLayoutScheme("dark", theme, merged, logger); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged["background"] != "#1a2332" || merged["text"] != "#f0f4f8" {
		t.Errorf("conforming spec override must survive, got %v", merged)
	}
	if logs.Len() != 0 {
		t.Errorf("no log expected, got %d", logs.Len())
	}
}

func TestEnforceLayoutScheme_InertWithoutDeclaredScheme(t *testing.T) {
	// 15 of 18 seeded layouts have scheme NULL; "neutral" is also a
	// legal column value. Both must disable the guard entirely, with
	// no per-render log noise.
	for _, scheme := range []string{"", "neutral"} {
		theme := map[string]string{"background": "#0f172a", "text": "#e2e8f0"}
		merged := map[string]string{"background": "#F4F5F7", "text": "#1e293b"}
		logger, logs := observedLogger()

		if err := enforceLayoutScheme(scheme, theme, merged, logger); err != nil {
			t.Fatalf("scheme %q: unexpected error: %v", scheme, err)
		}

		if merged["background"] != "#F4F5F7" {
			t.Errorf("scheme %q: merge must be untouched, got %q", scheme, merged["background"])
		}
		if logs.Len() != 0 {
			t.Errorf("scheme %q: no log expected, got %d", scheme, logs.Len())
		}
	}
}

func TestEnforceLayoutScheme_UnjudgeableBackgroundPassesWithInfo(t *testing.T) {
	// A gradient/var() or an absent background cannot be judged — the
	// merge passes through, but never silently: a scheme-declaring site
	// gets an Info naming what was skipped (council round 2,
	// bug_historian's visibility-parity point).
	cases := []struct {
		name   string
		merged map[string]string
	}{
		{"non-hex background", map[string]string{
			"background": "linear-gradient(#fff, #eee)",
			"text":       "#1e293b",
		}},
		{"no background at all", map[string]string{
			"text": "#1e293b",
		}},
	}
	for _, tc := range cases {
		theme := map[string]string{"background": "#0f172a", "text": "#e2e8f0"}
		before := tc.merged["background"]
		logger, logs := observedLogger()

		if err := enforceLayoutScheme("dark", theme, tc.merged, logger); err != nil {
			t.Fatalf("%s: unexpected error: %v", tc.name, err)
		}

		if tc.merged["background"] != before {
			t.Errorf("%s: unjudgeable background must survive, got %q", tc.name, tc.merged["background"])
		}
		if logs.Len() != 1 {
			t.Fatalf("%s: expected one Info, got %d", tc.name, logs.Len())
		}
		if logs.All()[0].Level != zap.InfoLevel {
			t.Errorf("%s: skip log must be Info, got %v", tc.name, logs.All()[0].Level)
		}
	}
}

func TestEnforceLayoutScheme_IncompleteThemeFailsTheRender(t *testing.T) {
	// A theme that cannot supply BOTH background and text gives the
	// guard nothing safe to swap in: restoring one slot would be the
	// forbidden half-swap (council round 1), and shipping the violating
	// merge with only a Warn is the unwatched-signal shape behind this
	// incident class (council round 2). A complete theme palette is a
	// Phase 3 data invariant, so this is the loader's migration-gap
	// contract: hard error, nothing ships, the failure sweep sees it.
	cases := []struct {
		name  string
		theme map[string]string
	}{
		{"no background", map[string]string{"text": "#e2e8f0"}},
		{"no text", map[string]string{"background": "#0f172a"}},
		{"neither", map[string]string{}},
	}
	for _, tc := range cases {
		merged := map[string]string{"background": "#F4F5F7", "text": "#1e293b"}
		logger, logs := observedLogger()

		err := enforceLayoutScheme("dark", tc.theme, merged, logger)

		if err == nil {
			t.Fatalf("%s: expected an error — violating CSS must not render", tc.name)
		}
		if !strings.Contains(err.Error(), "refusing to render scheme-violating CSS") {
			t.Errorf("%s: error must state the refusal, got %q", tc.name, err.Error())
		}
		if merged["background"] != "#F4F5F7" || merged["text"] != "#1e293b" {
			t.Errorf("%s: merge must be left whole for the error report, got %v", tc.name, merged)
		}
		if logs.Len() != 0 {
			t.Errorf("%s: the error is the signal — no extra log expected, got %d", tc.name, logs.Len())
		}
	}
}
