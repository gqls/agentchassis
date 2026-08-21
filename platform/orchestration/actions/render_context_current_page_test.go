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
//
// UPDATED 2026-08-19 (bugs_open/260): there is no longer a regex fallback, and
// contextToMap — the "context-to-map serialiser" named above — is deleted. The
// assertions that named it now render through the seam instead, which is the
// same journey with one fewer place to drop the value.

package actions

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
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
	}

	derived := renderContextScalarFields(ctx)
	got := map[string]bool{}
	for k := range derived {
		// The same predicate renderCtxToMap itself uses: unserialised fields
		// and control fields both stay out of the step contract. (The control
		// set is empty since 2026-08-21 — schema_mode was deleted as dead and
		// its successor, InputSchema, is a map and so is excluded structurally.)
		if renderContextStepContractExcluded(k) {
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
		// A template key crosses the boundary under its STEP name — the same
		// key for every field but the renderContextStepContractRenames entries
		// (current_page → current_page_name). Looking the slot up under the
		// step name is what keeps this a check on "can it arrive", not on
		// "is it spelt the same".
		if _, ok := serialised[renderContextStepContractKey(key)]; ok {
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

	// build_render_context stores this map in collected_data — under the STEP
	// name, current_page_name (RFC_029 §10.13 step 4: the page's name string
	// must not share a key with the page RECORD, input_data.current_page, in
	// one collected_data tree; see TestStepOutputNeverCarriesCurrentPage).
	stored := renderCtxToMap(built)
	if got, _ := stored["current_page_name"].(string); got != "capabilities" {
		t.Fatalf("renderCtxToMap dropped current_page (expected under current_page_name): got %q", got)
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

	// Path 2 WAS the regex fallback, used whenever template execution errored.
	// It is deleted (bugs_open/260) along with contextToMap, so there is no
	// second map builder that can disagree with the first. What the assertion
	// protected — that a revived context reaches the RENDER with its
	// current_page intact — is now asserted where it actually matters, at the
	// seam, since that is the only remaining consumer.
	if out := mustRender(t, `<body class="page-{{.current_page}}">`, revived, zap.NewNop()); !strings.Contains(out, `page-capabilities`) {
		t.Errorf("rendered output lost current_page: %s", out)
	}
}

// TestStepOutputNeverCarriesCurrentPage is the door RFC_029 §10.13 step 4
// closes. The resolver's whole-tree search, asked for `current_page` on behalf
// of page-content-writer's generate_content step, used to collect the page
// RECORD (input_data.current_page) AND build_render_context's page-NAME string
// (render_context.current_page) and record a conflict on every run — 23 rows in
// the four hours after v1.0.1315, the whole surviving class. The string now
// crosses the boundary as current_page_name, and nothing in the step output may
// reintroduce the old key — not the scalar projection and not the ContentData
// merge that follows it.
//
// MUTATION THAT MUST BREAK IT: emit `result[key]` instead of
// `result[renderContextStepContractKey(key)]` in renderCtxToMap (first
// subtest), or drop the renamed-key skip from its ContentData loop (second).
func TestStepOutputNeverCarriesCurrentPage(t *testing.T) {
	t.Run("the scalar crosses under the step name only", func(t *testing.T) {
		stored := renderCtxToMap(&RenderContext{Domain: "example.com", CurrentPage: "about"})
		if _, present := stored["current_page"]; present {
			t.Fatalf("step output carries `current_page` = %v — the page-name string shares a key "+
				"with the page record again and the resolver conflict is back", stored["current_page"])
		}
		if got, _ := stored["current_page_name"].(string); got != "about" {
			t.Fatalf("step output current_page_name = %q, want %q", got, "about")
		}
	})

	t.Run("ContentData cannot smuggle the old key across", func(t *testing.T) {
		ctx := &RenderContext{Domain: "example.com", CurrentPage: "about",
			ContentData: map[string]interface{}{"current_page": "smuggled", "hero_url": "/h.png"}}
		stored := renderCtxToMap(ctx)
		if _, present := stored["current_page"]; present {
			t.Errorf("ContentData[\"current_page\"] reached the step output — the merge loop must skip renamed template names")
		}
		if stored["hero_url"] != "/h.png" {
			t.Errorf("the ContentData merge stopped carrying ordinary keys (hero_url = %v)", stored["hero_url"])
		}
	})
}

// TestRestoreReadsOnlyTheStepBoundaryName: the read-side tolerance that also
// accepted the TEMPLATE name as a string fallback (orchestrations in flight
// across the step-4 roll) was RETIRED on 2026-08-21 — zero non-terminal
// pre-roll orchestrations remained, and every live stored row carrying the old
// key as a string agreed with the fresh base value (18/18; grounds in the
// setRenderContextScalarsFromData comment). The old spelling is not part of
// the read contract: a string under it is IGNORED, not adopted. The "old key
// alone" case below is the pin that fails if the tolerance quietly returns —
// asserting want="" there is only meaningful because the same map with the
// step name present resolves (case 1), so an empty result cannot be the loop
// skipping the field altogether.
func TestRestoreReadsOnlyTheStepBoundaryName(t *testing.T) {
	cases := []struct {
		name string
		data map[string]interface{}
		want string
	}{
		{"step name: current_page_name", map[string]interface{}{"current_page_name": "about"}, "about"},
		{"old key alone is NOT adopted (tolerance retired)", map[string]interface{}{"current_page": "about"}, ""},
		{"both present: only the step name is read", map[string]interface{}{"current_page_name": "new", "current_page": "old"}, "new"},
		{"the page RECORD under the old key stays unread too", map[string]interface{}{"current_page": map[string]interface{}{"name": "about"}}, ""},
		{"neither: stays empty", map[string]interface{}{"domain": "example.com"}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &RenderContext{}
			mergeIntoRenderContext(ctx, tc.data)
			if ctx.CurrentPage != tc.want {
				t.Errorf("CurrentPage = %q, want %q", ctx.CurrentPage, tc.want)
			}
		})
	}
}

// TestStepContractRenamesAreWellFormed keeps the rename map honest: every
// entry must name a real template tag, its target must not collide with any
// other tag (or the rename would overwrite a sibling in the step output — the
// same silent-overwrite TestRenderContextJSONTagsAreUnique guards), and the
// template contract must still advertise the ORIGINAL name — a rename is a
// step-boundary split, never a template rename.
func TestStepContractRenamesAreWellFormed(t *testing.T) {
	tags := renderContextScalarFields(&RenderContext{})
	for from, to := range renderContextStepContractRenames {
		if _, ok := tags[from]; !ok {
			t.Errorf("renderContextStepContractRenames[%q] names no RenderContext json tag", from)
		}
		if _, collides := tags[to]; collides {
			t.Errorf("rename %q → %q collides with an existing tag %q — the step output would overwrite it", from, to, to)
		}
		if from == to {
			t.Errorf("rename %q → %q is a no-op entry", from, to)
		}
		if _, advertised := contextToInterfaceMap(&RenderContext{})[from]; !advertised {
			t.Errorf("template data no longer advertises %q — the rename leaked into the template contract", from)
		}
		// The regex-fallback half of this check went with contextToMap
		// (bugs_open/260): with one map builder left, contextToInterfaceMap
		// above is the whole template contract.
		if _, advertisedNew := contextToInterfaceMap(&RenderContext{})[to]; advertisedNew {
			t.Errorf("template data advertises the step name %q — the rename leaked into the template contract", to)
		}
	}
}

// TestStepContractRenamesStayRare is the growth guard the council's architecture
// seat asked for (corr f3716ebe, round 2, advisory): renderContextStepContractRenames
// is a hand-authored collision table, and nothing else stops it quietly
// accumulating entries the way a shared action accumulates optional keys
// (RFC_022's budget exists for exactly that shape). So the size is PINNED. Adding
// an entry is allowed — but it costs reading the new producer, writing its reason
// into the map comment, and raising this number on purpose in the same commit.
// If you are here because this failed and you did not add an entry, someone
// parked a collision in the map without tracing it: read the candidate-set query
// in the staged_component_build RUNBOOK first.
func TestStepContractRenamesStayRare(t *testing.T) {
	const declared = 1 // current_page → current_page_name (RFC_029 §10.13 step 4)
	if got := len(renderContextStepContractRenames); got != declared {
		t.Fatalf("renderContextStepContractRenames has %d entries, %d declared — a new step-boundary "+
			"rename needs its producer read and its reason written beside the map; then raise `declared` "+
			"deliberately in the same commit", got, declared)
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
	// The former regex-fallback assertion is now made at the seam, which is
	// where an unset field could still surface as Go's "<no value>" artefact
	// (bugs_open/260 deleted the fallback, not the missingkey handling).
	if out := mustRender(t, `<body class="page-{{.current_page}}">`, revived, zap.NewNop()); !strings.Contains(out, `class="page-"`) {
		t.Errorf("an unset current_page must render as empty, got: %s", out)
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
