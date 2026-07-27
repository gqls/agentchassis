// FILE: platform/orchestration/actions/render_context_current_page_test.go
//
// Guards bugs_open/085 — the render data advertised `current_page` and the
// build path always supplied it empty, so no section component could know
// which page it was on and nothing could vary per page.
//
// The defect was never a missing assignment in one place. It was a value
// dropped at THREE points on one journey, each of which looked complete on its
// own: the page source was merged through an allowlist that did not include
// it, the context-to-map serialiser did not emit it, and the render side did
// not restore it into the struct. Fixing any one of them leaves the field
// still empty at the template, which is why the tests below follow the value
// end to end rather than asserting each function in isolation.
//
// TestCurrentPageSurvivesBothRenderPaths is the one that actually fails if
// someone "tidies" the restore in mergeIntoRenderContext away: the html/template
// path would keep working (ContentData wins there) while the regex fallback
// silently went back to empty — a half-fix that reads as a whole one.

package actions

import (
	"sort"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"
)

// knownUnserialisedContextFields are the keys contextToInterfaceMap advertises
// to component templates that renderCtxToMap does NOT write into
// collected_data. Each one is a field a template can read and, on the
// page-build path, always find empty — the same shape as bugs_open/085.
//
// This list is a RECORD OF THE GAP, not an approval of it. It is pinned so the
// gap cannot grow silently: adding a field to RenderContext and wiring it into
// the template map without the serialiser fails TestRenderContextSerialisation
// CoversTemplateContract, which is precisely the mistake nobody caught for
// current_page. Shrinking the list is the goal; growing it needs a reason.
//
//   - contact_email — benign: an alias of email, derived at render time from
//     the same struct field, which IS serialised.
//   - logo_url      — reaches the map in practice via ContentData (both
//     BuildRenderContextAction and mergeIntoRenderContextEnhanced write the
//     ContentData key alongside the struct field), so the struct field alone is
//     latent rather than live.
//   - theme_css / title / description — genuinely dropped at the step boundary,
//     exactly as current_page was. Tracked in bugs_open/109; not fixed here
//     because each needs its own producer decided, and bundling three unrelated
//     behaviour changes into a bug fix is what the reviewers object to.
var knownUnserialisedContextFields = []string{
	"contact_email",
	"description",
	"logo_url",
	"theme_css",
	"title",
}

// TestRenderContextSerialisationCoversTemplateContract is the guard the
// bugs_open/085 fix would have needed to be unnecessary. The defect was a field
// the template contract advertises and the serialiser omits; that is a property
// of two maps in one package, and it is checkable without running anything.
//
// It deliberately checks the CONTRACT rather than the values: a runtime warning
// on an empty field would be noise, because most of these are legitimately
// empty most of the time. What is never legitimate is a field that CANNOT
// arrive.
func TestRenderContextSerialisationCoversTemplateContract(t *testing.T) {
	// Populate every scalar so nothing is omitted for being a zero value — the
	// question is whether the serialiser has a slot, not whether it is filled.
	ctx := &RenderContext{
		Domain: "example.com", LogoText: "Example", CompanyName: "Example Ltd",
		Tagline: "t", CurrentPage: "index", Email: "e@example.com", Phone: "0",
		PrimaryColor: "#1", SecondaryColor: "#2", AccentColor: "#3",
		TextColor: "#4", BackgroundColor: "#5", Year: "2026",
		CTAText: "Go", CTAUrl: "/contact.html", Industry: "i", Tone: "plain",
		TargetAudience: "a", SiteID: uuid.New(),
		ThemeCSS: "css", Title: "ti", Description: "de", LogoURL: "/l.png",
	}

	serialised := renderCtxToMap(ctx)        // what crosses the step boundary
	advertised := contextToInterfaceMap(ctx) // what templates are told they may read

	known := make(map[string]bool, len(knownUnserialisedContextFields))
	for _, k := range knownUnserialisedContextFields {
		known[k] = true
	}

	var unexpected, staleAllowance []string
	for key := range advertised {
		if _, ok := serialised[key]; ok {
			if known[key] {
				staleAllowance = append(staleAllowance, key)
			}
			continue
		}
		if !known[key] {
			unexpected = append(unexpected, key)
		}
	}
	sort.Strings(unexpected)
	sort.Strings(staleAllowance)

	if len(unexpected) > 0 {
		t.Errorf("field(s) %v are advertised to component templates but never written by renderCtxToMap, "+
			"so on the page-build path a template reading them always gets empty — this is bugs_open/085's "+
			"exact shape. Either serialise them in renderCtxToMap, or add them to "+
			"knownUnserialisedContextFields with a reason.", unexpected)
	}
	if len(staleAllowance) > 0 {
		t.Errorf("field(s) %v are listed in knownUnserialisedContextFields but ARE now serialised — "+
			"delete them from the list so it keeps naming real gaps.", staleAllowance)
	}
}

func TestResolveCurrentPageName(t *testing.T) {
	// Envelope shapes taken from live orchestration_states on 2026-07-27, not
	// from the struct: this value is assembled by workflow config, so the
	// config is the contract.
	cases := []struct {
		name      string
		collected map[string]interface{}
		config    map[string]interface{}
		want      string
	}{
		{
			name: "page-content-writer: name, via the configured source path",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"current_page": map[string]interface{}{
						"name": "capabilities",
						"url":  "/capabilities.html",
					},
				},
			},
			config: map[string]interface{}{
				"sources": map[string]interface{}{
					"page": "input_data.current_page",
					"site": "input_data.site_record",
				},
			},
			want: "capabilities",
		},
		{
			name: "rerender envelope: page_name",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"current_page": map[string]interface{}{
						"domain":    "vonc.com",
						"page_name": "about",
						"reason":    "section_data_resolved",
					},
				},
			},
			config: map[string]interface{}{
				"sources": map[string]interface{}{"page": "input_data.current_page"},
			},
			want: "about",
		},
		{
			name: "the .html suffix is stripped — CurrentPage is compared bare",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"current_page": map[string]interface{}{"name": "index.html"},
				},
			},
			config: map[string]interface{}{
				"sources": map[string]interface{}{"page": "input_data.current_page"},
			},
			want: "index",
		},
		{
			name: "both keys present: name wins, and they agree in the live data",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"current_page": map[string]interface{}{
						"name":      "about",
						"page_name": "about",
					},
				},
			},
			config: map[string]interface{}{
				"sources": map[string]interface{}{"page": "input_data.current_page"},
			},
			want: "about",
		},
		{
			name: "no sources map (array form): falls back to the conventional path",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"current_page": map[string]interface{}{"name": "pricing"},
				},
			},
			config: map[string]interface{}{
				"sources": []interface{}{"input_data", "site_record"},
			},
			want: "pricing",
		},
		{
			name: "a non-page source is not mined for a name",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"site_record": map[string]interface{}{"name": "Fundamentally AI"},
				},
			},
			config: map[string]interface{}{
				"sources": map[string]interface{}{"page": "input_data.current_page"},
			},
			want: "",
		},
		{
			name:      "nothing at all: empty, not a panic",
			collected: map[string]interface{}{},
			config:    map[string]interface{}{},
			want:      "",
		},
		{
			name: "a page object with no name key yields empty rather than a guess",
			collected: map[string]interface{}{
				"input_data": map[string]interface{}{
					"current_page": map[string]interface{}{"page_id": "a28abcd7"},
				},
			},
			config: map[string]interface{}{
				"sources": map[string]interface{}{"page": "input_data.current_page"},
			},
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveCurrentPageName(tc.collected, tc.config, zaptest.NewLogger(t)); got != tc.want {
				t.Errorf("resolveCurrentPageName() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestCurrentPageSurvivesBothRenderPaths follows the value the whole way: the
// context built by build_render_context is serialised into collected_data,
// reconstituted by render_component, and turned into template data by each of
// the two renderers. A break anywhere on that chain fails here.
func TestCurrentPageSurvivesBothRenderPaths(t *testing.T) {
	built := &RenderContext{Domain: "fundamentallyai.com", CurrentPage: "capabilities"}

	// build_render_context stores this map in collected_data.
	stored := renderCtxToMap(built)
	if got, _ := stored["current_page"].(string); got != "capabilities" {
		t.Fatalf("renderCtxToMap dropped current_page: got %q", got)
	}

	// render_component reads it back out of collected_data.
	revived := &RenderContext{}
	mergeIntoRenderContext(revived, stored)
	if revived.CurrentPage != "capabilities" {
		t.Fatalf("mergeIntoRenderContext dropped current_page: got %q", revived.CurrentPage)
	}

	// Path 1: Go html/template. ContentData is merged over the base map here,
	// so this one would pass on the catch-all alone.
	if got, _ := contextToInterfaceMap(revived)["current_page"].(string); got != "capabilities" {
		t.Errorf("html/template data has current_page = %q, want %q", got, "capabilities")
	}

	// Path 2: the regex fallback, used whenever template execution errors.
	// contextToMap skips any ContentData key the base map already holds, and
	// the base map holds ctx.CurrentPage — so this passes ONLY because the
	// struct field was restored above. This is the half of the fix that is
	// easy to delete without noticing.
	if got := contextToMap(revived)["current_page"]; got != "capabilities" {
		t.Errorf("regex fallback data has current_page = %q, want %q", got, "capabilities")
	}
}

// TestCurrentPageEmptyStaysEmpty is the negative control: an unset page must
// render as empty on both paths, never as "<no value>" or a stale neighbour.
// A component that branches on current_page degrades to "show everything", and
// that degradation has to stay reachable and quiet.
func TestCurrentPageEmptyStaysEmpty(t *testing.T) {
	revived := &RenderContext{}
	mergeIntoRenderContext(revived, renderCtxToMap(&RenderContext{Domain: "example.com"}))

	if got, _ := contextToInterfaceMap(revived)["current_page"].(string); got != "" {
		t.Errorf("html/template data has current_page = %q, want empty", got)
	}
	if got := contextToMap(revived)["current_page"]; got != "" {
		t.Errorf("regex fallback data has current_page = %q, want empty", got)
	}
}
