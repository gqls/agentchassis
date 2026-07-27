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
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap/zaptest"
)

// knownUnserialisedContextFields are the keys contextToInterfaceMap advertises
// to component templates that renderCtxToMap does NOT write into
// collected_data. Each one is a field a template can read and, on the
// page-build path, always find empty — the same shape as bugs_open/085.
//
// This list is a RECORD OF THE GAP, not an approval of it.
//
// It is DERIVED from renderContextUnserialised (bugs_open/109) rather than
// written out again here. It used to be a second hand-maintained copy of the
// same set, which is precisely the drift class this whole area keeps producing:
// two lists that must agree, with nothing checking that they do. The only entry
// that is not a struct field is contact_email, so that one is named here.
//
//   - contact_email — benign: an alias of email, derived at render time from
//     the same struct field, which IS serialised. It has no field of its own on
//     RenderContext, so it cannot come from renderContextUnserialised.
//
// Every other entry, and its reason, lives with the mechanism it constrains.
func knownUnserialisedContextFieldsFn() []string {
	out := []string{"contact_email"}
	for k := range renderContextUnserialised {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// TestRenderContextUnserialisedEntriesGiveReasons: an entry with no reason is
// indistinguishable from a field somebody forgot, which is the exact failure
// renderContextUnserialised exists to make impossible. Empty reasons defeat it.
func TestRenderContextUnserialisedEntriesGiveReasons(t *testing.T) {
	for key, reason := range renderContextUnserialised {
		if len(strings.TrimSpace(reason)) < 40 {
			t.Errorf("renderContextUnserialised[%q] has no real reason (%q) — "+
				"an unexplained omission is what this map exists to prevent", key, reason)
		}
	}
}

// TestRenderCtxToMapDerivationIsBehaviourPreserving pins the refactor that
// replaced renderCtxToMap's hand-written scalar literal with a derivation from
// the struct's json tags (bugs_open/109).
//
// The keys below are the EXACT set the literal produced before the change,
// transcribed from it. If derivation produces a different set, the refactor
// changed what crosses the step boundary — which is a behaviour change, not a
// tidy-up, and must be justified rather than discovered later.
func TestRenderCtxToMapDerivationIsBehaviourPreserving(t *testing.T) {
	wasLiteral := []string{
		"domain", "logo_text", "company_name", "tagline", "current_page",
		"email", "phone", "primary_color", "secondary_color", "accent_color",
		"text_color", "background_color", "year", "cta_text", "cta_url",
		"industry", "tone", "target_audience",
	}

	ctx := &RenderContext{
		Domain: "example.com", LogoText: "Example", CompanyName: "Example Ltd",
		Tagline: "t", CurrentPage: "index", Email: "e@example.com", Phone: "0",
		PrimaryColor: "#1", SecondaryColor: "#2", AccentColor: "#3",
		TextColor: "#4", BackgroundColor: "#5", Year: "2026",
		CTAText: "Go", CTAUrl: "/contact.html", Industry: "i", Tone: "plain",
		TargetAudience: "a",
		// The deliberately-unserialised ones, all populated: if any of them
		// starts crossing the boundary, this test says so.
		ThemeCSS: "css", Title: "ti", Description: "de", LogoURL: "/l.png",
		SchemaMode: "flexible",
	}

	derived := renderContextScalarFields(ctx)
	got := map[string]bool{}
	for k := range derived {
		if _, skip := renderContextUnserialised[k]; skip {
			continue
		}
		got[k] = true
	}

	for _, k := range wasLiteral {
		if !got[k] {
			t.Errorf("key %q was serialised by the old literal and is not derived now — "+
				"the refactor DROPPED a field from collected_data", k)
		}
		delete(got, k)
	}
	var added []string
	for k := range got {
		added = append(added, k)
	}
	sort.Strings(added)
	if len(added) > 0 {
		t.Errorf("key(s) %v are newly serialised that the old literal did not write. "+
			"That may well be correct, but it is a behaviour change: decide it "+
			"deliberately and update this test's transcription.", added)
	}
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

	known := make(map[string]bool)
	for _, k := range knownUnserialisedContextFieldsFn() {
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

// TestRenderContextJSONTagsAreUnique closes the failure mode the derivation
// introduces, raised by the council's bug_historian seat (bugs_open/109).
//
// Deriving keys from struct tags trades one silent failure for a narrower one:
// a duplicate tag makes two fields collide in the derived map, and map
// insertion silently overwrites, so one field's value would vanish under the
// other's with no compile-time signal. The old hand-written literal could not
// express that, so the guard has to come with the mechanism.
//
// A typo'd (rather than duplicated) tag is not detectable here — that shows up
// as a key nothing reads, which TestRenderContextSerialisationCoversTemplate
// Contract catches from the other side.
func TestRenderContextJSONTagsAreUnique(t *testing.T) {
	seen := map[string]string{}
	rt := reflect.TypeOf(RenderContext{})
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		key := strings.Split(f.Tag.Get("json"), ",")[0]
		if key == "" || key == "-" {
			continue
		}
		if prev, dup := seen[key]; dup {
			t.Errorf("json tag %q is on both %s and %s — the derived map would silently "+
				"drop one of them (map insertion overwrites)", key, prev, f.Name)
		}
		seen[key] = f.Name
	}
}

// TestRenderContextUnserialisedReasonsAreDistinct: the council also noted that a
// length-only check on the reasons would pass long filler. Copy-pasted reasons
// are the realistic form of that, so require them to differ from one another.
func TestRenderContextUnserialisedReasonsAreDistinct(t *testing.T) {
	byReason := map[string]string{}
	for key, reason := range renderContextUnserialised {
		r := strings.ToLower(strings.Join(strings.Fields(reason), " "))
		if prev, dup := byReason[r]; dup {
			t.Errorf("%q and %q carry an identical reason — at least one of them was not "+
				"thought about", prev, key)
		}
		byReason[r] = key
	}
}
