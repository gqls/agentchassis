package actions

import (
	"testing"

	"go.uber.org/zap"
)

// mustRenderReporting is the test-side companion to bugs_open/260's seam
// change. Every existing test that renders through this seam passes a template
// it expects to work, so the error return has exactly one correct handling in a
// test: fail, loudly, naming the template. Discarding it with `_` would restore
// in the tests the very silence the change removes from production — a render
// that stopped executing would show up as a mysterious empty-output assertion
// failure three lines later, or as nothing at all.
// It returns the rendered output only: no existing test consumes the missing /
// dead-URL reports through this path (they are exercised directly by the
// dead-control tests), and returning three values where callers wanted one made
// every call site carry two blank identifiers.
func mustRenderReporting(t *testing.T, tmpl string, ctx *RenderContext, logger *zap.Logger) string {
	t.Helper()
	out, _, _, err := RenderTemplateReportingMissing(tmpl, ctx, logger)
	if err != nil {
		t.Fatalf("template failed to render (bugs_open/260 seam): %v", err)
	}
	return out
}

// mustRender is the same for the single-value wrapper.
func mustRender(t *testing.T, tmpl string, ctx *RenderContext, logger *zap.Logger) string {
	t.Helper()
	out, err := RenderTemplate(tmpl, ctx, logger)
	if err != nil {
		t.Fatalf("template failed to render (bugs_open/260 seam): %v", err)
	}
	return out
}
