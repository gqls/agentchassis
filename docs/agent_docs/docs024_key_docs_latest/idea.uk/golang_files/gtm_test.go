package main

// gtm_test.go — Google Tag Manager on the pages this BINARY serves.
//
// Why this file exists at all: idea.uk is two applications behind one domain.
// The chassis-built static site carries GTM in its head/header chrome slots, but
// this service serves eleven pages the static build never sees — including
// "Request received" and "Payment received". Those two are the only pages that
// can evidence a conversion, so a tag that stops at the static site measures
// traffic and cannot measure a sale.
//
// The tests below assert the MECHANISM FIRED, not merely that nothing broke: a
// test that only checks the off-state passes just as happily when the feature has
// been deleted.

import (
	"strings"
	"testing"
)

func gtmApp(t *testing.T, id string) *App {
	t.Helper()
	return &App{cfg: Config{GTMContainerID: id, OperatorEmail: "ops@idea.uk"}}
}

// The container must land in <head> as high as it can go — after <meta charset>,
// before everything else — and the noscript must be the FIRST thing after <body>.
// Position is the whole requirement here; "the string appears somewhere" is not.
func TestGTMPlacement(t *testing.T) {
	out := gtmApp(t, "GTM-PQ3WCTBD").page("Payment received", "<p>thanks</p>")

	charset := strings.Index(out, `<meta charset="utf-8">`)
	script := strings.Index(out, "googletagmanager.com/gtm.js")
	viewport := strings.Index(out, `name="viewport"`)
	title := strings.Index(out, "<title>")
	if charset < 0 || script < 0 || viewport < 0 || title < 0 {
		t.Fatalf("missing landmark: charset=%d script=%d viewport=%d title=%d",
			charset, script, viewport, title)
	}
	if !(charset < script && script < viewport && script < title) {
		t.Errorf("GTM script not as high as possible in <head>: charset=%d script=%d viewport=%d title=%d",
			charset, script, viewport, title)
	}

	// "Immediately after" means immediately — not merely somewhere in the body.
	body := strings.Index(out, "<body>")
	if body < 0 {
		t.Fatal("no <body> in page output")
	}
	after := strings.TrimLeft(out[body+len("<body>"):], "\r\n")
	if !strings.HasPrefix(after, "<!-- Google Tag Manager (noscript) -->") {
		t.Errorf("noscript is not the first thing after <body>; got: %.90q", after)
	}
	if !strings.Contains(out, `googletagmanager.com/ns.html?id=GTM-PQ3WCTBD`) {
		t.Error("noscript iframe missing or wrong container")
	}
	if strings.Count(out, "googletagmanager.com/gtm.js") != 1 {
		t.Error("GTM script should appear exactly once")
	}
}

// Unset must be genuinely inert — not "renders an empty comment", not "renders a
// container id of the empty string", which would load GTM's script with no id.
func TestGTMUnsetEmitsNothing(t *testing.T) {
	out := gtmApp(t, "").page("Terms", "<p>x</p>")
	if strings.Contains(out, "googletagmanager") || strings.Contains(out, "Google Tag Manager") {
		t.Error("GTM markup emitted with no container configured")
	}
}

// The id reaches a JavaScript string literal AND a URL. Validate, don't escape:
// anything that is not a container id is dropped whole.
func TestGTMRejectsMalformedContainer(t *testing.T) {
	for _, bad := range []string{
		`GTM-X');alert(1);('`, // breaks out of the JS string literal
		`GTM-X" onload="evil`, // breaks out of the iframe src attribute
		`GTM X`,               // space
		strings.Repeat("G", 33),
	} {
		if got := gtmSanitiseID(bad); got != "" {
			t.Errorf("gtmSanitiseID(%q) = %q, want \"\"", bad, got)
		}
		if out := gtmApp(t, bad).page("t", "b"); strings.Contains(out, "googletagmanager") {
			t.Errorf("malformed container %q still rendered GTM", bad)
		}
	}
	for _, good := range []string{"GTM-PQ3WCTBD", "GTM-ABC1234", "G-XYZ_9"} {
		if gtmSanitiseID(good) != good {
			t.Errorf("gtmSanitiseID(%q) rejected a valid container", good)
		}
	}
}

// Every HTML route goes through page(), so proving the wrapper covers the two
// conversion pages is what makes the "every page" claim true rather than hopeful.
func TestGTMOnConversionPages(t *testing.T) {
	app := gtmApp(t, "GTM-PQ3WCTBD")
	for _, title := range []string{"Request received", "Payment received", "Nothing was charged", "Privacy"} {
		out := app.page(title, "<p>body</p>")
		if !strings.Contains(out, "googletagmanager.com/gtm.js") ||
			!strings.Contains(out, "googletagmanager.com/ns.html") {
			t.Errorf("page %q missing GTM", title)
		}
	}
}

// The embedded landing page is shadowed by the static site today, but if the
// route ever falls back it must not be the one untagged page on the domain.
func TestGTMLandingPagePlaceholdersAreWired(t *testing.T) {
	if !strings.Contains(string(pageHTML), "<!--GTM_HEAD-->") ||
		!strings.Contains(string(pageHTML), "<!--GTM_BODY-->") {
		t.Fatal("page.html has lost its GTM placeholders — NewApp's Replacer would be a silent no-op")
	}
	app := NewApp(Config{GTMContainerID: "GTM-PQ3WCTBD", OperatorEmail: "ops@idea.uk"}, nil, nil)
	got := string(app.landingHTML)
	if strings.Contains(got, "<!--GTM_HEAD-->") || strings.Contains(got, "<!--GTM_BODY-->") {
		t.Error("placeholders survived rendering — they were not substituted")
	}
	if !strings.Contains(got, "googletagmanager.com/gtm.js") ||
		!strings.Contains(got, "googletagmanager.com/ns.html") {
		t.Error("landing page did not receive GTM")
	}
}
