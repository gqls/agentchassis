// FILE: platform/orchestration/actions/component_library_form_action_test.go
//
// Guards the fleet-wide dead-contact-form fix (bugs_open/006 §B).
//
// The load-bearing property is ORDERING, not the substitution itself:
// ContentData is merged over the base map, so the broken values — which the
// content LLM actively wrote — would survive any default set alongside
// cta_url. TestFormActionSurvivesContentDataMerge is the regression that
// actually fails if someone "simplifies" this into a defaultString() call.

package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestSanitiseFormAction(t *testing.T) {
	cases := []struct {
		name    string
		action  interface{}
		present bool
		email   string
		domain  string
		want    string
		wantKey bool
	}{
		{
			name:   "hash anchor with a real address becomes a mailto",
			action: "#contact", present: true,
			email: "gas@contactforsales.com", domain: "gaswholesalers.com",
			want:    "mailto:gas@contactforsales.com?subject=gaswholesalers.com enquiry",
			wantKey: true,
		},
		{
			name:   "empty action with a real address becomes a mailto",
			action: "", present: true,
			email: "finetuning@contactforsales.com", domain: "finetuning.uk",
			want:    "mailto:finetuning@contactforsales.com?subject=finetuning.uk enquiry",
			wantKey: true,
		},
		{
			name:   "the historic /contact endpoint that never existed is also repaired",
			action: "/contact", present: true,
			email: "darts@contactforsales.com", domain: "dartsonline.com",
			want:    "mailto:darts@contactforsales.com?subject=dartsonline.com enquiry",
			wantKey: true,
		},
		{
			// The honesty case. robot-hands, relojistas, vetcomparison and vonc
			// had no contact address at all on 2026-07-20. Synthesising
			// info@<domain> would make the form look repaired while still
			// dropping the message — and would remove the only outward sign
			// that anything is wrong.
			name:   "no resolvable address leaves the action alone rather than fabricating one",
			action: "#contact", present: true,
			email: "", domain: "robot-hands.com",
			want:    "#contact",
			wantKey: true,
		},
		{
			name:   "a malformed address is not treated as an address",
			action: "#contact", present: true,
			email: "not-an-address", domain: "vonc.com",
			want:    "#contact",
			wantKey: true,
		},
		{
			// The gap the council REVISE (guardian + editquality) flagged.
			// section_editor_actions.go:452 / multipage_actions.go:333 set
			// ctx.Email = "info@"+Domain as a display fallback BEFORE rendering,
			// so on those paths an address-less site would otherwise get a
			// fabricated mailto:info@<domain>. The synthesised value must be
			// refused, same as an empty one — the form is left for the check.
			name:   "the synthesised info@<domain> fallback is refused, not turned into a mailto",
			action: "#contact", present: true,
			email: "info@robot-hands.com", domain: "robot-hands.com",
			want:    "#contact",
			wantKey: true,
		},
		{
			// The guard is precise: it refuses ONLY info@<the site's own domain>
			// (the synthesised fallback). A genuinely configured info@ on a
			// different domain — a shared CRM inbox, say — is a real address and
			// must still be honoured.
			name:   "a real info@ on a different domain is still honoured",
			action: "#contact", present: true,
			email: "info@leopardess.uk", domain: "robot-hands.com",
			want:    "mailto:info@leopardess.uk?subject=robot-hands.com enquiry",
			wantKey: true,
		},
		{
			name:   "an existing mailto is left untouched",
			action: "mailto:idea.uk@contactforsales.com?subject=idea.uk enquiry", present: true,
			email: "someone@else.com", domain: "idea.uk",
			want:    "mailto:idea.uk@contactforsales.com?subject=idea.uk enquiry",
			wantKey: true,
		},
		{
			// idea.uk's tool genuinely accepts POST at /request. Rewriting a
			// live backend route to a mailto would break a working funnel.
			name:   "a real backend handler is left untouched",
			action: "/request", present: true,
			email: "idea.uk@contactforsales.com", domain: "idea.uk",
			want:    "/request",
			wantKey: true,
		},
		{
			name:    "a component with no form does not acquire a form_action",
			present: false,
			email:   "gas@contactforsales.com", domain: "gaswholesalers.com",
			wantKey: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := map[string]interface{}{}
			if tc.present {
				data["form_action"] = tc.action
			}
			ctx := &RenderContext{Email: tc.email, Domain: tc.domain}

			sanitiseFormAction(data, ctx)

			got, ok := data["form_action"]
			if ok != tc.wantKey {
				t.Fatalf("form_action present = %v, want %v", ok, tc.wantKey)
			}
			if !tc.wantKey {
				return
			}
			if gotStr, _ := got.(string); gotStr != tc.want {
				t.Errorf("form_action = %q, want %q", gotStr, tc.want)
			}
		})
	}
}

// TestFormActionSurvivesContentDataMerge is the real regression guard.
//
// contextToInterfaceMap merges ContentData OVER the base map. "#contact" is a
// value the content LLM wrote into content_data, so a default declared next to
// cta_url would be silently overwritten — the fix would appear to work (the 3
// sites with an empty action would be repaired) while the 8 sites carrying
// "#contact" stayed broken. That is precisely the shape this test exists to
// catch, because it fails loudly rather than half-passing.
func TestFormActionSurvivesContentDataMerge(t *testing.T) {
	ctx := &RenderContext{
		Domain: "vonc.com",
		Email:  "hello@vonc.com",
		ContentData: map[string]interface{}{
			"form_action": "#contact",
			"heading":     "Say Something Worth Saying",
		},
	}

	data := contextToInterfaceMap(ctx)

	got, _ := data["form_action"].(string)
	if got == "#contact" {
		t.Fatal("form_action is still #contact after the merge — the sanitiser " +
			"ran before ContentData was merged, or was replaced by a base-map " +
			"default. Live submissions would still be lost. See bugs_open/006 §B.")
	}
	if !strings.HasPrefix(got, "mailto:hello@vonc.com") {
		t.Errorf("form_action = %q, want a mailto to the site address", got)
	}

	// The merge itself must still work — sanitising must not clobber content.
	if h, _ := data["heading"].(string); h != "Say Something Worth Saying" {
		t.Errorf("heading = %q, want the ContentData value preserved", h)
	}
}

// TestFormActionSanitisedOnRegexFallbackPath covers the SECOND render branch.
//
// RenderTemplate falls back to regex substitution (contextToMap) whenever Go
// template execution errors. That path merges ContentData too, so it can carry
// the same "#contact". Fixing only contextToInterfaceMap would leave a branch
// that still renders a dead form while the change reads as complete — the
// failure mode 016b §9 calls out and which the pre-commit pattern check caught
// here as an untouched twin.
func TestFormActionSanitisedOnRegexFallbackPath(t *testing.T) {
	ctx := &RenderContext{
		Domain: "gaswholesalers.com",
		Email:  "gas@contactforsales.com",
		ContentData: map[string]interface{}{
			"form_action": "#contact",
		},
	}

	data := contextToMap(ctx)

	if data["form_action"] == "#contact" {
		t.Fatal("form_action is still #contact on the regex fallback path — " +
			"the Go-template path was fixed and its twin was not, so a template " +
			"execution error silently restores the original defect.")
	}
	if want := "mailto:gas@contactforsales.com?subject=gaswholesalers.com enquiry"; data["form_action"] != want {
		t.Errorf("form_action = %q, want %q", data["form_action"], want)
	}
}

// The honesty rule must hold identically on both paths — a site with no address
// must not acquire a fabricated one just because rendering took the fallback.
func TestFormActionNotFabricatedOnRegexFallbackPath(t *testing.T) {
	ctx := &RenderContext{
		Domain:      "robot-hands.com",
		ContentData: map[string]interface{}{"form_action": "#contact"},
	}

	if got := contextToMap(ctx)["form_action"]; got != "#contact" {
		t.Errorf("form_action = %q, want it left as #contact — no address is "+
			"resolvable and a synthesised one would hide the breakage", got)
	}
}

// The info@<domain> honesty guard must also hold on the regex fallback path.
// Here ctx.Email is the synthesised "info@"+Domain value some render paths set
// before rendering; if template execution errors and drops to contextToMap, a
// fabricated mailto:info@<domain> must NOT appear on the page.
func TestFormActionInfoFallbackNotFabricatedOnRegexFallbackPath(t *testing.T) {
	ctx := &RenderContext{
		Domain:      "robot-hands.com",
		Email:       "info@robot-hands.com", // the section_editor/multipage synth fallback
		ContentData: map[string]interface{}{"form_action": "#contact"},
	}

	if got := contextToMap(ctx)["form_action"]; got != "#contact" {
		t.Errorf("form_action = %q, want it left as #contact — info@<own domain> "+
			"is the synthesised display fallback, not a real inbox", got)
	}
}

// TestRenderTemplateReportingMissingSeedsFormActionWhenTemplateReferencesIt is
// the regression guard for bugs_open/228: a component whose html_template
// references {{.form_action}} but whose ContentData never authored the field
// (contact-block's actual shape — the content-generation schema never asked
// for one) must still get the sanitiser's protection, not a silently
// empty/absent field.
func TestRenderTemplateReportingMissingSeedsFormActionWhenTemplateReferencesIt(t *testing.T) {
	ctx := &RenderContext{
		Domain: "example.com",
		Email:  "hello@example.com",
		ContentData: map[string]interface{}{
			"heading": "Contact us",
		},
	}

	out, _, _ := RenderTemplateReportingMissing(
		`<form action="{{.form_action}}" method="POST"><h2>{{.heading}}</h2></form>`,
		ctx, zap.NewNop(),
	)

	want := `mailto:hello@example.com?subject=example.com enquiry`
	if !strings.Contains(out, `action="`+want+`"`) {
		t.Fatalf("rendered form action = %q, want it to contain action=%q (bugs_open/228: "+
			"a template referencing form_action must get the sanitiser even when "+
			"ContentData never authored the field)", out, want)
	}
	if strings.Contains(out, "<no value>") {
		t.Fatalf("rendered output still contains the raw <no value> artefact: %q", out)
	}
}

// TestRenderTemplateReportingMissingLeavesFormActionHonestWhenNoAddress mirrors
// the sanitiser's own refusal-to-fabricate rule through the seeding path: a
// site with no resolvable email must not get a fabricated mailto just because
// the seed made the field present.
func TestRenderTemplateReportingMissingLeavesFormActionHonestWhenNoAddress(t *testing.T) {
	ctx := &RenderContext{
		Domain:      "robot-hands.com",
		ContentData: map[string]interface{}{},
	}

	out, _, _ := RenderTemplateReportingMissing(
		`<form action="{{.form_action}}" method="POST"></form>`,
		ctx, zap.NewNop(),
	)

	if !strings.Contains(out, `action=""`) {
		t.Errorf("rendered output = %q, want action=\"\" (seeded, sanitised, "+
			"left empty — not fabricated, not the stripped <no value> shape)", out)
	}
	if strings.Contains(out, "mailto:") {
		t.Errorf("rendered output = %q, fabricated a mailto with no resolvable address", out)
	}
}

// TestRenderTemplateReportingMissingDoesNotSeedFormActionForUnrelatedTemplate
// guards the OTHER direction: a template that never mentions form_action must
// not have the key injected into its shared ContentData map — mirroring
// TestSanitiseFormAction's "a component with no form does not acquire a
// form_action" at the RenderTemplateReportingMissing entry point, since that
// is where the seeding actually happens.
func TestRenderTemplateReportingMissingDoesNotSeedFormActionForUnrelatedTemplate(t *testing.T) {
	ctx := &RenderContext{
		Domain:      "example.com",
		Email:       "hello@example.com",
		ContentData: map[string]interface{}{"heading": "No form here"},
	}

	RenderTemplateReportingMissing(`<h2>{{.heading}}</h2>`, ctx, zap.NewNop())

	if _, present := ctx.ContentData["form_action"]; present {
		t.Fatalf("form_action was seeded into ContentData for a template that never "+
			"references it — got %v", ctx.ContentData["form_action"])
	}
}
