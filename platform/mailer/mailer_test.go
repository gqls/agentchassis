package mailer

import (
	"encoding/base64"
	"strings"
	"testing"
)

func baseCfg() Config {
	return Config{Host: "smtp.example.com", Port: "587", From: "robot-hands@contactforsales.com"}
}

// The 998-octet SMTP line limit is the reason bodies are base64-wrapped at all.
// A long unbroken HTML line was corrupted by server-side folding before the
// original added this, so it is asserted rather than assumed.
func TestBuildWrapsEveryLineWellUnderTheSMTPLimit(t *testing.T) {
	long := strings.Repeat("<div>a very long unbroken line of html</div>", 400)
	raw, err := Build(baseCfg(), Message{To: []string{"a@b.com"}, Subject: "s", Text: "t", HTML: long})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	for i, line := range strings.Split(string(raw), "\r\n") {
		if len(line) > 998 {
			t.Fatalf("line %d is %d octets, over the SMTP 998 limit", i, len(line))
		}
	}
}

func TestBuildRoundTripsBothBodies(t *testing.T) {
	raw, err := Build(baseCfg(), Message{
		To: []string{"a@b.com"}, Subject: "Your dossier", Text: "plain here", HTML: "<p>rich here</p>",
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	s := string(raw)
	if !strings.Contains(s, "multipart/alternative") {
		t.Fatal("two bodies must produce multipart/alternative")
	}
	// The plain part must precede the HTML part or a text-only client shows the
	// wrong one — clients take the LAST part they can render.
	if strings.Index(s, "text/plain") > strings.Index(s, "text/html") {
		t.Fatal("text/plain part must come first in multipart/alternative")
	}
	for _, want := range []string{"plain here", "<p>rich here</p>"} {
		if !decodedContains(t, s, want) {
			t.Fatalf("body %q did not survive encoding", want)
		}
	}
}

func TestBuildTextOnlyIsNotMultipart(t *testing.T) {
	raw, err := Build(baseCfg(), Message{To: []string{"a@b.com"}, Subject: "s", Text: "only text"})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if strings.Contains(string(raw), "multipart") {
		t.Fatal("a text-only message must not be multipart")
	}
	if !decodedContains(t, string(raw), "only text") {
		t.Fatal("text body did not survive encoding")
	}
}

// Header injection: the property this package adds over the ported original,
// which interpolated the recipient directly.
func TestBuildRefusesHeaderInjection(t *testing.T) {
	cases := []struct {
		name string
		m    Message
	}{
		{"recipient", Message{To: []string{"a@b.com\r\nBcc: evil@x.com"}, Subject: "s", Text: "t"}},
		{"subject", Message{To: []string{"a@b.com"}, Subject: "s\r\nBcc: evil@x.com", Text: "t"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Build(baseCfg(), tc.m); err == nil {
				t.Fatal("expected refusal, got a message")
			}
		})
	}
	// Positive control: the same shapes WITHOUT a line break must still build,
	// or the check above would pass by rejecting everything.
	if _, err := Build(baseCfg(), Message{To: []string{"a@b.com"}, Subject: "s Bcc: x", Text: "t"}); err != nil {
		t.Fatalf("positive control failed: %v", err)
	}
}

func TestBuildRejectsEmptyRecipientsAndBody(t *testing.T) {
	if _, err := Build(baseCfg(), Message{Subject: "s", Text: "t"}); err == nil {
		t.Fatal("expected refusal with no recipients")
	}
	if _, err := Build(baseCfg(), Message{To: []string{"a@b.com"}, Subject: "s"}); err == nil {
		t.Fatal("expected refusal with no text body — it is the fallback part")
	}
}

// 465 is implicit TLS and everything else is STARTTLS. Getting this the wrong
// way round hangs rather than erroring, so it gets its own assertion.
func TestUsesImplicitTLSOnlyFor465(t *testing.T) {
	if !UsesImplicitTLS("465") {
		t.Fatal("465 must take the implicit-TLS path")
	}
	for _, p := range []string{"587", "25", "2525"} {
		if UsesImplicitTLS(p) {
			t.Fatalf("port %s must take the STARTTLS path", p)
		}
	}
}

func TestNewValidatesConfig(t *testing.T) {
	bad := []Config{
		{Port: "587", From: "a@b.com"},                                  // no host
		{Host: "h", From: "a@b.com"},                                    // no port
		{Host: "h", Port: "587"},                                        // no from
		{Host: "h", Port: "587", From: "not-an-address"},                // malformed
		{Host: "h", Port: "587", From: "a@b.com", FromName: "x\r\ny"},   // injection
		{Host: "h", Port: "587", From: "a@b.com", ReplyTo: "bad\r\nhi"}, // injection
	}
	for i, cfg := range bad {
		if _, err := New(cfg); err == nil {
			t.Fatalf("case %d: expected rejection", i)
		}
	}
	if _, err := New(baseCfg()); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
}

func TestFromEnvDefaultsPortAndRequiresHostAndFrom(t *testing.T) {
	t.Setenv("TESTMAIL_HOST", "smtp.example.com")
	t.Setenv("TESTMAIL_FROM", "a@b.com")
	cfg, err := FromEnv("TESTMAIL")
	if err != nil {
		t.Fatalf("FromEnv: %v", err)
	}
	if cfg.Port != "587" {
		t.Fatalf("expected default port 587, got %q", cfg.Port)
	}
	t.Setenv("TESTMAIL_HOST", "")
	if _, err := FromEnv("TESTMAIL"); err == nil {
		t.Fatal("expected an error when HOST is unset")
	}
}

// decodedContains base64-decodes every encoded part and looks for want, so the
// assertion tests what a mail client would actually render rather than the
// wire bytes.
func decodedContains(t *testing.T, msg, want string) bool {
	t.Helper()
	for _, chunk := range strings.Split(msg, "\r\n\r\n")[1:] {
		clean := strings.NewReplacer("\r\n", "", "\n", "").Replace(chunk)
		if i := strings.Index(clean, "--"); i >= 0 {
			clean = clean[:i]
		}
		if dec, err := base64.StdEncoding.DecodeString(clean); err == nil {
			if strings.Contains(string(dec), want) {
				return true
			}
		}
	}
	return false
}
