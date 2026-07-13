package actions

import (
	"image/color"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestParseHexColour(t *testing.T) {
	cases := []struct {
		in   string
		want color.RGBA
	}{
		{"#0080FF", color.RGBA{0x00, 0x80, 0xFF, 0xff}},
		{"1a1a2e", color.RGBA{0x1a, 0x1a, 0x2e, 0xff}},
		{"#abc", color.RGBA{0xaa, 0xbb, 0xcc, 0xff}},
		// Gradients / junk fall back to the dark neutral — OG cards need a solid.
		{"linear-gradient(135deg, #3b82f6 0%, #2563eb 100%)", color.RGBA{0x1a, 0x1a, 0x2e, 0xff}},
		{"", color.RGBA{0x1a, 0x1a, 0x2e, 0xff}},
	}
	for _, c := range cases {
		got := parseHexColour(c.in)
		if got != color.Color(c.want) {
			t.Errorf("parseHexColour(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestInjectBrandHeadTags(t *testing.T) {
	log := zap.NewNop()
	ctx := &RenderContext{Domain: "robot-hands.com", CompanyName: "Robot-Hands", Tagline: "Grip intelligence", LogoURL: "/assets/images/logo.jpg"}

	head := "<head>\n  <title>x</title>\n</head>"
	out := injectBrandHeadTags(head, ctx, log)

	for _, want := range []string{
		`rel="icon" href="/assets/images/favicon.png"`,
		`rel="icon" href="/assets/images/logo.jpg"`,
		`property="og:image" content="https://robot-hands.com/assets/images/og-card.png"`,
		`property="og:title" content="Robot-Hands"`,
		`name="twitter:card" content="summary_large_image"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("injected head missing %q\n---\n%s", want, out)
		}
	}
	if !strings.HasSuffix(strings.TrimSpace(out), "</head>") {
		t.Errorf("tags must be injected BEFORE </head>; got tail: %q", out[len(out)-40:])
	}

	// Idempotent: a head that already has icon markup is left untouched.
	already := `<head><link rel="icon" href="/x.png"></head>`
	if got := injectBrandHeadTags(already, ctx, log); got != already {
		t.Errorf("expected no-op on head with existing favicon, got: %s", got)
	}

	// No </head> → returned unchanged (defensive).
	if got := injectBrandHeadTags("no head here", ctx, log); got != "no head here" {
		t.Errorf("expected unchanged on malformed head, got: %s", got)
	}

	// Escaping: a company name with quotes/ampersand must not break the attr.
	ctx2 := &RenderContext{Domain: "x.com", CompanyName: `A&B "Co"`, Tagline: ""}
	out2 := injectBrandHeadTags("<head></head>", ctx2, log)
	if !strings.Contains(out2, `content="A&amp;B &quot;Co&quot;"`) {
		t.Errorf("attr not escaped: %s", out2)
	}
}
