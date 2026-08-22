package main

import (
	"strings"
	"testing"
)

// The real `hero` template's shape: an if/else on the background, a bare
// headline, and three optional blocks. This is the 5-row case.
const heroTpl = `<section class="hero" style="{{if or .hero_url .background_image}}background-image: url('{{or .hero_url .background_image}}');{{else}}background: var(--color-primary);{{end}}">
        <div class="hero-content">
            <h1>{{.headline}}</h1>
            {{if .subheadline}}<p class="hero-subheadline">{{.subheadline}}</p>{{end}}
            {{if and .cta_text .cta_url}}<a href="{{.cta_url}}" class="btn btn-primary">{{.cta_text}}</a>{{end}}
        </div>
    </section>`

func TestRecoversHeroWithBranchesTaken(t *testing.T) {
	html := `<section class="hero" style="background: var(--color-primary);">
        <div class="hero-content">
            <h1>Time-To-Kill Calculator</h1>
            <p class="hero-subheadline">Work out your combat loop's heartbeat.</p>
            <a href="/tools/ttk.html" class="btn btn-primary">Open the tool</a>
        </div>
    </section>`
	data, reason := recoverRow(Row{Component: "hero", Template: heroTpl, RenderedHTML: html})
	if reason != "" {
		t.Fatalf("expected recovery, got refusal: %s", reason)
	}
	for k, want := range map[string]string{
		"headline":    "Time-To-Kill Calculator",
		"subheadline": "Work out your combat loop's heartbeat.",
		"cta_text":    "Open the tool",
		"cta_url":     "/tools/ttk.html",
	} {
		if got, _ := data[k].(string); got != want {
			t.Errorf("%s = %q, want %q", k, got, want)
		}
	}
	// The untaken branch must bind NOTHING — inventing a background_image would
	// be exactly the over-fill this tool exists to avoid.
	if _, ok := data["hero_url"]; ok {
		t.Error("hero_url bound although the else-branch rendered")
	}
	if _, ok := data["background_image"]; ok {
		t.Error("background_image bound although the else-branch rendered")
	}
}

func TestRecoversHeroWithOptionalBlocksAbsent(t *testing.T) {
	html := `<section class="hero" style="background: var(--color-primary);">
        <div class="hero-content">
            <h1>Just a headline</h1>
            
            
        </div>
    </section>`
	data, reason := recoverRow(Row{Component: "hero", Template: heroTpl, RenderedHTML: html})
	if reason != "" {
		t.Fatalf("expected recovery, got refusal: %s", reason)
	}
	if got, _ := data["headline"].(string); got != "Just a headline" {
		t.Errorf("headline = %q", got)
	}
	if len(data) != 1 {
		t.Errorf("expected ONLY headline, got %v", data)
	}
}

func TestRecoversRangeBody(t *testing.T) {
	tpl := `<ul>{{range .features}}<li><b>{{.title}}</b>: {{.body}}</li>{{end}}</ul>`
	html := `<ul><li><b>Fast</b>: renders in milliseconds</li><li><b>Safe</b>: proven by round-trip</li></ul>`
	data, reason := recoverRow(Row{Component: "features", Template: tpl, RenderedHTML: html})
	if reason != "" {
		t.Fatalf("expected recovery, got refusal: %s", reason)
	}
	items, ok := data["features"].([]interface{})
	if !ok || len(items) != 2 {
		t.Fatalf("features = %#v, want 2 items", data["features"])
	}
	first, _ := items[0].(map[string]interface{})
	if first["title"] != "Fast" || first["body"] != "renders in milliseconds" {
		t.Errorf("first item = %#v", first)
	}
}

// ── CONTROLS: each of these MUST refuse. A tool that cannot fail is not a gate.

func TestRefusesHTMLThatIsNotTheTemplatesOutput(t *testing.T) {
	// The 9-row case: a whole tool page stored in a slot pointing at `hero`.
	html := `<div class="tool-page"><div class="tool-header"><h1>Time-To-Kill</h1></div></div>`
	if data, reason := recoverRow(Row{Component: "hero", Template: heroTpl, RenderedHTML: html}); reason == "" {
		t.Fatalf("MUST refuse tool-page HTML for a hero template, but recovered %v", data)
	}
}

func TestRefusesTemplateReadingRenderContextKeys(t *testing.T) {
	tpl := `<footer>{{.company_name}} &copy; {{.year}} — {{.tagline}}</footer>`
	html := `<footer>Acme &copy; 2026 — we do things</footer>`
	_, reason := recoverRow(Row{Component: "footer", Template: tpl, RenderedHTML: html})
	if !strings.Contains(reason, "RenderContext key") {
		t.Fatalf("MUST refuse a RenderContext-reading template; got %q", reason)
	}
}

func TestRefusesUnattributablePipeline(t *testing.T) {
	// {{or .a .b}} renders identically whichever field held the value, so the
	// round-trip could NOT catch a wrong attribution. The matcher must refuse.
	tpl := `<img src="{{or .hero_url .background_image}}">`
	html := `<img src="/img/x.png">`
	if _, reason := recoverRow(Row{Component: "hero", Template: tpl, RenderedHTML: html}); reason == "" {
		t.Fatal("MUST refuse an {{or}} pipeline it cannot attribute to one field")
	}
}

func TestRefusesPartialMatchThatLeavesTrailingBytes(t *testing.T) {
	html := `<section class="hero" style="background: var(--color-primary);">
        <div class="hero-content">
            <h1>Headline</h1>
            
            
        </div>
    </section><!-- something else appended -->`
	if _, reason := recoverRow(Row{Component: "hero", Template: heroTpl, RenderedHTML: html}); reason == "" {
		t.Fatal("MUST refuse when the template does not account for every stored byte")
	}
}

// The gate itself: prove that a recovery which does NOT round-trip is rejected,
// by mutating the stored HTML so no binding can reproduce it.
func TestRoundTripGateRejectsMutatedHTML(t *testing.T) {
	html := `<section class="hero" style="background: var(--color-primary);">
        <div class="hero-content">
            <h2>Wrong tag</h2>
            
            
        </div>
    </section>`
	if _, reason := recoverRow(Row{Component: "hero", Template: heroTpl, RenderedHTML: html}); reason == "" {
		t.Fatal("MUST refuse: <h2> cannot come from a template that writes <h1>")
	}
}

func TestEmittedSQLGuardsOnNullAndDigest(t *testing.T) {
	stmt, err := updateStatement(
		Row{PageComponentID: "11111111-2222-3333-4444-555555555555", RenderedHTML: "<p>x</p>"},
		map[string]interface{}{"headline": "it's fine"},
	)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"content_data IS NULL",
		"md5(rendered_html) = '" + "9e3f4b0a4dbb1b2b7b6b1e2a3c4d5e6f"[:0], // presence of the clause, not the value
		"md5(rendered_html) = '",
		"WHERE id = '11111111-2222-3333-4444-555555555555'",
	} {
		if !strings.Contains(stmt, want) {
			t.Errorf("emitted SQL missing %q:\n%s", want, stmt)
		}
	}
	// An apostrophe in page copy must not break the literal.
	if !strings.Contains(stmt, "it's fine") {
		t.Errorf("dollar-quoting lost the value:\n%s", stmt)
	}
}

// ── The test that makes the ROUND-TRIP GATE load-bearing. ─────────────────────
//
// Every control above is caught by the MATCHER (the bytes cannot be matched at
// all), so disabling the round-trip gate left the suite green — a guard in
// series, and the reason this test exists. Here the matcher SUCCEEDS and only
// the gate can reject: it binds subheadline="" and takes the {{if}} branch, but
// re-rendering with "" makes {{if}} false, the <p></p> vanishes, and the output
// no longer matches the stored bytes.
//
// The wider point this protects: an EMPTY value must never be recovered as a
// present field. Writing subheadline:"" would satisfy ContentDataCanFillTemplate
// and make the component regenerable into something it does not currently serve.
func TestRoundTripGateCatchesEmptyValueInTakenBranch(t *testing.T) {
	tpl := `<div>{{if .subheadline}}<p>{{.subheadline}}</p>{{end}}</div>`
	html := `<div><p></p></div>`
	data, reason := recoverRow(Row{Component: "x", Template: tpl, RenderedHTML: html})
	if reason == "" {
		t.Fatalf("MUST refuse: an empty value in a taken branch cannot re-render to the stored bytes (got %v)", data)
	}
	if !strings.Contains(reason, "round-trip") {
		t.Fatalf("expected the ROUND-TRIP gate to reject this, not an earlier guard; got %q", reason)
	}
}
