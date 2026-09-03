package actions

import (
	"regexp"
	"testing"
)

// TestCanonicalTokensCoverEveryEmittedInkCompanion is a LOCKSTEP test between
// two lists that must agree and live in different files:
//
//	buildLegibleInkDefaults (palette_specialised_slots.go) EMITS the legible-ink
//	companions, and canonicalCSSTokens (component_validation.go) decides which
//	custom properties AuditTemplateTokens treats as known.
//
// They drifted, and the drift pointed the signal backwards: the renderer's own
// comment tells a component to opt into var(--color-primary-ink, …) as the
// repair for an illegible palette colour, while the audit reported that exact
// opt-in as unknown vocabulary (bugs_open/458 §6). Adoption was 15 of 412
// active unforked components `[MEASURED 2026-09-03]`.
//
// The expectation is DERIVED FROM THE EMITTER rather than written out here, so
// a future slot added to buildLegibleInkDefaults fails this test instead of
// silently becoming unknown drift.
func TestCanonicalTokensCoverEveryEmittedInkCompanion(t *testing.T) {
	// A palette that reaches every branch: primary and accent both illegible on
	// their grounds, plus the cta_bg/cta_text pair --color-cta-bg-ink needs.
	palette := map[string]string{
		"primary": "#1A1F2E", "accent": "#C49A3C",
		"background": "#0F1820", "surface": "#1B2A3B", "text": "#E8E2D9",
		"cta_bg": "#1A1F2E", "cta_text": "#FFFFFF", "primary_text": "#FFFFFF",
	}

	css := buildLegibleInkDefaults("", palette, defaultInkPolicy(), zapNop())
	if css == "" {
		t.Fatal("buildLegibleInkDefaults emitted nothing for a palette chosen to " +
			"trigger every slot; this test can no longer see what it guards")
	}

	decl := regexp.MustCompile(`(--[a-z0-9-]+)\s*:`)
	found := decl.FindAllStringSubmatch(css, -1)
	if len(found) == 0 {
		t.Fatalf("no custom-property declarations parsed out of the emitted block:\n%s", css)
	}

	seen := map[string]bool{}
	for _, m := range found {
		name := m[1]
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, ok := canonicalCSSTokens[name]; !ok {
			t.Errorf("buildLegibleInkDefaults emits %s but canonicalCSSTokens does not "+
				"contain it, so AuditTemplateTokens reports a component opting into the "+
				"renderer's own documented repair as unknown drift. Add %q to "+
				"canonicalCSSTokens in component_validation.go.", name, name)
		}
	}
	t.Logf("checked %d emitted ink companions against canonicalCSSTokens", len(seen))
}
