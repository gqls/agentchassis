// FILE: platform/orchestration/actions/render_context_derivation_test.go
//
// Guards the bugs_open/109 completion: all four render-context maps (build,
// serialise, restore, render×2) now derive their scalar cores from the ONE
// declaration — RenderContext's json tags — instead of five hand-written
// allowlists that had already drifted apart. The transcription tests pin the
// refactor as behaviour-preserving where it must be, and pin the two
// deliberate divergence closures where it must not be.

package actions

import (
	"sort"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// fullyPopulatedRenderContext sets every scalar so that a dropped field cannot
// hide behind being a zero value.
func fullyPopulatedRenderContext() *RenderContext {
	return &RenderContext{
		Domain: "example.com", LogoText: "Example", LogoURL: "/l.png",
		CompanyName: "Example Ltd", Tagline: "t", CurrentPage: "index",
		Email: "e@example.com", Phone: "01", PrimaryColor: "#1",
		SecondaryColor: "#2", AccentColor: "#3", TextColor: "#4",
		BackgroundColor: "#5", ThemeCSS: "css", Title: "ti", Description: "de",
		CTAText: "Go", CTAUrl: "/go.html", Year: "2026", Industry: "i",
		Tone: "plain", TargetAudience: "a", SchemaMode: "flexible",
	}
}

// TestContextToInterfaceMapDerivationIsBehaviourPreserving transcribes the
// exact key set the hand-written literal produced before the derivation. The
// html/template path's contract must not change shape in a refactor.
func TestContextToInterfaceMapDerivationIsBehaviourPreserving(t *testing.T) {
	wasLiteral := []string{
		"domain", "site_id", "logo_text", "logo_url", "company_name", "tagline",
		"nav_items", "current_page", "primary_color", "secondary_color",
		"accent_color", "text_color", "background_color", "theme_css", "title",
		"description", "email", "contact_email", "phone", "cta_text", "cta_url",
		"year", "industry", "tone", "target_audience", "services",
	}

	got := contextToInterfaceMap(fullyPopulatedRenderContext())

	for _, k := range wasLiteral {
		if _, ok := got[k]; !ok {
			t.Errorf("key %q was advertised by the old literal and is missing now — "+
				"the refactor narrowed the template contract", k)
		}
		delete(got, k)
	}
	var added []string
	for k := range got {
		added = append(added, k)
	}
	sort.Strings(added)
	if len(added) > 0 {
		t.Errorf("key(s) %v are newly advertised to templates. That may be correct, "+
			"but it is a contract change: decide it deliberately and update this "+
			"transcription.", added)
	}
}

// ── TestContextToMapDerivationClosesTheLogoURLGap DELETED 2026-08-19 with its
// subject (bugs_open/260). It transcribed the REGEX-FALLBACK map's key set —
// contextToMap's 22 literals plus the deliberate logo_url closure — and existed
// because the two hand-written key lists had already drifted apart, which is
// the drift class bugs_open/109 exists to end.
//
// The drift is now unrepresentable rather than merely tested: there is ONE map
// builder left. Its key set is still transcribed, by the sibling test above,
// and that test is what a contract change now has to argue with.

// TestSchemaModeNeverReachesTemplatesOrStruct: control fields are the one
// class the derivation must NOT liberate. schema_mode steers validation; if
// source data could set it, content could turn validation off.
func TestSchemaModeNeverReachesTemplatesOrStruct(t *testing.T) {
	if _, ok := contextToInterfaceMap(fullyPopulatedRenderContext())["schema_mode"]; ok {
		t.Error("schema_mode is advertised to templates — control fields must stay out of the contract")
	}
	ctx := &RenderContext{}
	setRenderContextScalarsFromData(ctx, map[string]interface{}{"schema_mode": "strict"})
	if ctx.SchemaMode != "" {
		t.Error("schema_mode was set from data — arbitrary content could flip validation strictness")
	}
	mergeIntoRenderContext(ctx, map[string]interface{}{"schema_mode": "strict"})
	if ctx.SchemaMode != "" {
		t.Error("restore set schema_mode from data — control fields must not be data-settable")
	}
}

// TestBuildAcceptsEveryStepContractScalar: the build map's old hand-list
// silently dropped any field it did not name (bugs_open/085's first drop
// point). Now every step-contract scalar must be accepted from a source.
func TestBuildAcceptsEveryStepContractScalar(t *testing.T) {
	logger := zap.NewNop()
	for key := range renderContextScalarFields(&RenderContext{}) {
		if renderContextStepContractExcluded(key) {
			continue
		}
		ctx := &RenderContext{}
		// A source supplies the scalar under its STEP name (the same as the
		// template name for all but renderContextStepContractRenames).
		mergeIntoRenderContextEnhanced(ctx, map[string]interface{}{renderContextStepContractKey(key): "value-for-" + key}, "test", logger)
		if got := renderContextScalarFields(ctx)[key]; got != "value-for-"+key {
			t.Errorf("build map dropped step-contract scalar %q (got %q) — "+
				"a source supplying it is silently ignored", key, got)
		}
	}
}

// TestBuildStillExcludesPerPageFields: title/description/theme_css stay out of
// the step contract until their producers are decided (they are per-page
// values; a site-level build adopting them would bleed one page's values onto
// every page — the reasons live in renderContextUnserialised).
func TestBuildStillExcludesPerPageFields(t *testing.T) {
	logger := zap.NewNop()
	ctx := &RenderContext{}
	mergeIntoRenderContextEnhanced(ctx, map[string]interface{}{
		"title": "T", "description": "D", "theme_css": "C",
	}, "test", logger)
	if ctx.Title != "" || ctx.Description != "" || ctx.ThemeCSS != "" {
		t.Errorf("build map adopted per-page fields at site level (title=%q description=%q theme_css=%q) — "+
			"these are excluded until their producer is decided", ctx.Title, ctx.Description, ctx.ThemeCSS)
	}
}

// TestStepContractScalarsSurviveTheRoundTrip generalises what
// TestCurrentPageSurvivesBothRenderPaths proved for one field: every scalar
// the serialiser emits must come back into the STRUCT on restore, because the
// regex-fallback renderer reads the struct, not ContentData. Before this, a
// dozen fields (colours, cta, industry, year …) restored into ContentData
// only, so fallback renders showed default colours and empty metadata where
// the main path showed real values.
func TestStepContractScalarsSurviveTheRoundTrip(t *testing.T) {
	built := fullyPopulatedRenderContext()
	stored := renderCtxToMap(built)

	revived := &RenderContext{}
	mergeIntoRenderContext(revived, stored)

	builtScalars := renderContextScalarFields(built)
	revivedScalars := renderContextScalarFields(revived)
	for key, want := range builtScalars {
		if renderContextStepContractExcluded(key) {
			continue
		}
		if revivedScalars[key] != want {
			t.Errorf("step-contract scalar %q did not survive serialise→restore into the "+
				"struct: got %q, want %q — the regex fallback path renders without it", key, revivedScalars[key], want)
		}
	}

	// And the concrete symptom that motivated it: real colours must reach the
	// RENDER, not the hard-coded defaults. This used to be asserted against
	// contextToMap, the regex fallback's map — deleted with the fallback
	// (bugs_open/260) — so it is asserted through the seam, which is both the
	// only remaining consumer and the place the symptom was actually seen.
	out := mustRender(t, `<div data-primary="{{.primary_color}}" data-industry="{{.industry}}" data-year="{{.year}}">`,
		revived, zap.NewNop())
	for field, want := range map[string]string{"primary": "#1", "industry": "i", "year": "2026"} {
		if !strings.Contains(out, `data-`+field+`="`+want+`"`) {
			t.Errorf("restored %s did not reach the render (want %q):\n%s", field, want, out)
		}
	}
}

// TestRestoreKeepsExcludedFieldsOutOfTheStruct: a data map carrying title (a
// per-page field outside the step contract) must not set the struct field on
// restore — it reaches templates via the ContentData catch-all exactly as
// before, so the html/template path is unchanged while the struct contract
// stays honest.
func TestRestoreKeepsExcludedFieldsOutOfTheStruct(t *testing.T) {
	ctx := &RenderContext{}
	mergeIntoRenderContext(ctx, map[string]interface{}{"title": "Page One"})
	if ctx.Title != "" {
		t.Errorf("restore set ctx.Title = %q from data — title is outside the step contract", ctx.Title)
	}
	if got, _ := ctx.ContentData["title"].(string); got != "Page One" {
		t.Errorf("title did not reach ContentData via the catch-all: got %q", got)
	}
}
