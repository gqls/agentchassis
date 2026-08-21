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
// ONE HELPER, because there is now ONE SPELLING (2026-08-21). `mustRenderReporting`
// was the companion to the deleted `RenderTemplateReportingMissing` and is kept as
// a thin alias only so the ~20 existing call sites do not churn in the same commit
// that changes the seam — a rename and a semantic change in one diff is two things
// to review as one. It returns the rendered output only: no test consumes the
// missing / dead-URL reports through this path (the dead-control tests call the
// seam directly).
func mustRenderReporting(t *testing.T, tmpl string, ctx *RenderContext, logger *zap.Logger) string {
	t.Helper()
	return mustRender(t, tmpl, ctx, logger)
}

// mustRender is the strict wrapper every test should use.
func mustRender(t *testing.T, tmpl string, ctx *RenderContext, logger *zap.Logger) string {
	t.Helper()
	out, _, _, err := RenderTemplate(tmpl, ctx, logger)
	if err != nil {
		t.Fatalf("template failed to render (bugs_open/260 seam): %v", err)
	}
	return out
}
