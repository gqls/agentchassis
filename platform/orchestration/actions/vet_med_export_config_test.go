package actions

import "testing"

// The export must never fall back to a default domain: a misconfigured task
// must fail closed rather than publish to a site we did not intend.
func TestParseMedExportConfigHasNoDefaultDomain(t *testing.T) {
	ec := parseMedExportConfig(map[string]interface{}{})
	if ec.Domain != "" {
		t.Fatalf("empty config produced default domain %q; exports must require an explicit domain", ec.Domain)
	}
}

func TestParseMedExportConfigUsesExplicitDomain(t *testing.T) {
	ec := parseMedExportConfig(map[string]interface{}{"domain": "example.uk"})
	if ec.Domain != "example.uk" {
		t.Fatalf("explicit domain not honoured: got %q", ec.Domain)
	}
}
