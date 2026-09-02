// FILE: platform/orchestration/actions/dead_url_guard_test.go
//
// bugs_open/238's guard. Two properties are worth more than the rest:
//
//   - With the config key unset the decision must be byte-identical to the old
//     behaviour. That is the whole basis on which this ships to a shared fleet
//     (owner ruling 2026-08-02: new authority on a shared seam is an opt-in
//     field with the unsafe default OFF), so it is asserted directly rather
//     than argued in a comment.
//   - The item key must carry page AND slot. Its sibling image_url_404's
//     empty-src key is site-wide, and finetuning.uk has had one `blocked` row
//     holding that single fleet slot since 2026-08-03 — so new damage on any
//     other page of that site could not mint an item at all. The detector was
//     live and correct and structurally unable to report. TestDeadURLControlItemKey
//     is the pin against reproducing it.
package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestShouldRefuseDeadURLControls(t *testing.T) {
	armed := map[string]interface{}{deadURLGuardConfigKey: true}
	dead := []string{"card1_image_url"}

	cases := []struct {
		name     string
		config   map[string]interface{}
		dead     []string
		template string
		want     bool
		why      string
	}{
		{
			name: "armed and a dead control present", config: armed, dead: dead,
			template: `<img src="" />`, want: true,
			why: "the motivating case: an ungated {{.card1_image_url}} inside src= rendered empty",
		},
		{
			name: "config absent entirely", config: map[string]interface{}{}, dead: dead,
			template: `<img src="" />`, want: false,
			why: "an un-armed caller must behave exactly as it did before this guard existed",
		},
		{
			name: "config explicitly false", config: map[string]interface{}{deadURLGuardConfigKey: false}, dead: dead,
			template: `<img src="" />`, want: false,
			why: "the unsafe default, stated explicitly, must still be the unsafe default",
		},
		{
			name: "config present but not a bool", config: map[string]interface{}{deadURLGuardConfigKey: "true"}, dead: dead,
			template: `<img src="" />`, want: false,
			why: "a string \"true\" is a mis-typed config, and a mis-typed config must fail OPEN — arming a fleet-wide refusal by accident is the worse error",
		},
		{
			name: "armed, nothing dead", config: armed, dead: nil,
			template: `<img src="/a.jpg" />`, want: false,
			why: "a clean render must never be touched",
		},
		{
			name: "armed, dead, but a runtime-fill shell", config: armed, dead: dead,
			template: `<div data-runtime-fill="tool"><a href=""></a></div>`, want: false,
			why: "runtime-fill shells hydrate their own hrefs client-side — an empty URL attribute there is intentional. Mirrors the chrome renderer's exemption exactly (render_guardian, 2026-07-22); the two must not drift",
		},
		{
			// THE DISCRIMINATING CASE, added 2026-09-02 with the template swap.
			// Before it, this function tested the RENDERED output, so a marker
			// appearing only in the render — a data-borne one, or under
			// features_open/035 a marker inside an EMBEDDED CHILD — exempted the
			// whole section. It now reads the TEMPLATE, the same artefact
			// deadURLFields is a fact about.
			//
			// If this ever returns false, the exemption has drifted back to reading
			// rendered bytes, and bugs_closed/137's upward leak is available again
			// one grain down: one child's marker covering its parent's own dead
			// controls and its siblings'. Every other case here passes under BOTH
			// spellings — this is the only one that can tell them apart.
			name: "marker in the RENDER but not the TEMPLATE must NOT exempt", config: armed, dead: dead,
			template: `<img src="{{.card1_image_url}}" />`, want: true,
			why: "the template carries no marker, so the dead control is real however the output looked",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldRefuseDeadURLControls(tc.config, tc.dead, tc.template); got != tc.want {
				t.Errorf("got %v, want %v — %s", got, tc.want, tc.why)
			}
		})
	}
}

// TestDeadURLControlItemKey pins the granularity, not the format. A key that
// omits page or slot lets ONE unresolved item hold the dedup slot for an entire
// site, which is not hypothetical: `image_url_404:empty-src` does exactly that,
// and finetuning.uk's has been `blocked` (a non-terminal status, so still inside
// idx_swi_dedup's predicate) since 2026-08-03 — count stuck at 16 while five
// fresh empty srcs on the homepage could mint nothing.
func TestDeadURLControlItemKey(t *testing.T) {
	fields := []string{"card1_image_url", "card2_image_url"}

	samePageDifferentSlot := []string{
		deadURLControlItemKey("index", "case-studies-grid", fields),
		deadURLControlItemKey("index", "featured-work", fields),
	}
	if samePageDifferentSlot[0] == samePageDifferentSlot[1] {
		t.Error("two slots on one page share a key — one section's open item would suppress the other's")
	}

	sameSlotDifferentPage := []string{
		deadURLControlItemKey("index", "case-studies-grid", fields),
		deadURLControlItemKey("about", "case-studies-grid", fields),
	}
	if sameSlotDifferentPage[0] == sameSlotDifferentPage[1] {
		t.Error("the same slot on two pages shares a key — this is the image_url_404:empty-src defect exactly")
	}

	key := deadURLControlItemKey("index", "case-studies-grid", fields)
	if !strings.Contains(key, "index") || !strings.Contains(key, "case-studies-grid") {
		t.Errorf("the key must name both page and slot so a human can act on it without a lookup, got %q", key)
	}
	// Stable across runs of the same defect: missingBareFields returns sorted
	// fields, so the same damage must not mint a second item on the next build.
	if key != deadURLControlItemKey("index", "case-studies-grid", fields) {
		t.Error("the key is not stable for identical input")
	}
}

// TestMissingBareFields_ScopeAwarenessOnTheCaseStudiesGridShape is the
// false-positive defence, on the real template shape that produced
// bugs_open/238. The guard is only safe to arm because the report it consumes is
// SCOPE-aware: it must name the ungated <img src="{{.card1_image_url}}"> and
// must NOT name the {{if}}-gated link, whose emptiness is authored degradation
// rather than a defect.
//
// It also demonstrates the composition with PBP-039's carry, which is the whole
// argument for having both: with the key present (the carry succeeded) there is
// nothing to refuse.
func TestMissingBareFields_ScopeAwarenessOnTheCaseStudiesGridShape(t *testing.T) {
	tpl := `<article class="csg-card">
  <img class="csg-card-image" src="{{.card1_image_url}}" alt="{{.card1_image_alt}}" loading="lazy" />
  {{if .card1_link_url}}<a class="csg-card-link" href="{{.card1_link_url}}">{{.read_case_study_label}}</a>{{end}}
</article>`

	t.Run("carry failed — the image field is named, the gated link is not", func(t *testing.T) {
		data := map[string]interface{}{
			"card1_image_alt":       "Abstract network diagram",
			"read_case_study_label": "Read case study",
			// card1_image_url and card1_link_url absent: the 238 state.
		}
		_, inURLAttr := missingBareFields(tpl, data)

		var namesImage, namesLink bool
		for _, f := range inURLAttr {
			switch f {
			case "card1_image_url":
				namesImage = true
			case "card1_link_url":
				namesLink = true
			}
		}
		if !namesImage {
			t.Error("card1_image_url must be reported: it sits at root scope inside src=, so its absence ships src=\"\" — the live defect")
		}
		if namesLink {
			t.Error("card1_link_url must NOT be reported: the template gates it with {{if}}, so its absence renders no anchor at all — authored degradation, and reporting it would make the guard refuse pages that are working as designed")
		}
	})

	t.Run("carry succeeded — nothing to refuse", func(t *testing.T) {
		data := map[string]interface{}{
			"card1_image_url": "/assets/images/case-study-facilities.jpg",
			"card1_image_alt": "Abstract network diagram",
		}
		_, inURLAttr := missingBareFields(tpl, data)
		if len(inURLAttr) != 0 {
			t.Errorf("a value supplied by the plan-time carry must leave nothing for the guard to refuse, got %v", inURLAttr)
		}
		if shouldRefuseDeadURLControls(map[string]interface{}{deadURLGuardConfigKey: true}, inURLAttr, "") {
			t.Error("the guard fired on a section the carry had already repaired — the two mechanisms must compose, not collide")
		}
	})
}

// TestDeadURLGuardConfigKeysAreDeclared pins that BOTH dead-URL flags are
// declared on the spec of the action that reads them.
//
// Why this is a test and not a comment. Neither flag reaches its action through
// the input extractor — the code calls recordDeadURLControls(config) /
// shouldRefuseDeadURLControls(config, ...), so a grep over the function body
// finds no literal config["..."] and the 2026-08-18 declaration census missed
// both. The cost of the omission is not a broken render: it is
// platform/validation/workflow.go reporting a LIVE, WORKING setting as a key
// "this action does not read — silently ignored at execution", whose stated fix
// is to delete it. A report that tells the next reader to disarm a working
// guard is worse than no report.
//
// MUTATION THAT MUST BREAK IT: remove "record_dead_url_controls" from
// RerenderPageSectionsInputSpec.ConfigKeys (or "refuse_dead_url_controls" from
// RenderComponentInputSpec.ConfigKeys) — the arm for that action then reports
// its own armed key as unknown.
func TestDeadURLGuardConfigKeysAreDeclared(t *testing.T) {
	for _, tc := range []struct {
		action string
		key    string
	}{
		{"rerender_page_sections", deadURLRecordConfigKey},
		{"render_component", deadURLGuardConfigKey},
	} {
		t.Run(tc.action, func(t *testing.T) {
			unknown, checked := datahelpers.UnknownConfigKeys(tc.action, map[string]interface{}{tc.key: true})
			if !checked {
				t.Fatalf("%s is not opted into unknown-config-key detection, so this test proves nothing — a green result here would be vacuous", tc.action)
			}
			if len(unknown) != 0 {
				t.Errorf("%s arms %q, but the action reports it as unrecognised %v — the migration that sets it would read as a no-op in the config report", tc.action, tc.key, unknown)
			}
		})
	}
}
