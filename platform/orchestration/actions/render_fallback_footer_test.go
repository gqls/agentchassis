// FILE: platform/orchestration/actions/render_fallback_footer_test.go
//
// Guards bugs_open/111's Go half: the fallback footer's Contact container is
// gated on its contents, like the DB footer components' {{if}} gates. A site
// with no contact route must not render a bare English "Contact" heading over
// empty space (the empty <a href="mailto:"></a> Cloudflare then rewrites into
// an email-protection stub was live on relojistas for a day).
package actions

import (
	"strings"
	"testing"
)

func TestRenderFallbackFooterGatesContactOnEmail(t *testing.T) {
	withEmail := RenderFallbackFooter(&RenderContext{
		LogoText: "Example", Email: "hello@example.com",
	})
	if !strings.Contains(withEmail, "footer-contact") || !strings.Contains(withEmail, "hello@example.com") {
		t.Errorf("footer with an email must render the contact block, got: %.200s", withEmail)
	}

	withoutEmail := RenderFallbackFooter(&RenderContext{LogoText: "Example"})
	if strings.Contains(withoutEmail, `class="footer-contact"`) {
		t.Error("footer without an email renders the contact container — a bare heading over nothing")
	}
	if strings.Contains(withoutEmail, "<h4>Contact</h4>") {
		t.Error("footer without an email renders the Contact heading")
	}
	// The gate must not take neighbouring columns with it.
	if !strings.Contains(withoutEmail, "footer-links") || !strings.Contains(withoutEmail, "footer-brand") {
		t.Error("gating the contact block removed a neighbouring footer column")
	}
}
