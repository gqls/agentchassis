// FILE: platform/orchestration/actions/validate_page_content_email_test.go
//
// Contract for check 5 (hallucinated emails), bugs_open/063. A site WITH a
// registered contact address is protected by the mismatch branch; a site with
// NO registered address previously had no protection at all — a plausible
// fabrication fell through both branches and deployed (relojistas homepage,
// 2026-07-24: mailto:relojistas@contactforsales.com served live ~4h). The
// contract now: no registered contact address means NO email may be asserted.

package actions

import "testing"

// A page asserting a plausible, non-placeholder contact email in both
// assertion contexts (text node and mailto: href).
const emailTestHTML = `
<section>
  <h3>Contacto</h3>
  <p>Escríbenos a <a href="mailto:relojistas@contactforsales.com">relojistas@contactforsales.com</a>.</p>
</section>`

func TestEmails_NoOfficialEmail_FlagsAnyAssertedEmail(t *testing.T) {
	issues := checkEmails(emailTestHTML, "")

	var saw bool
	for _, is := range issues {
		if is.Value == "relojistas@contactforsales.com" {
			saw = true
			if is.Type != "invalid_email" || is.Severity != "error" || is.Category != "email" {
				t.Errorf("expected invalid_email error, got %+v", is)
			}
		}
	}
	if !saw {
		t.Fatalf("with no registered contact address, an asserted email must be flagged; issues=%+v", issues)
	}
}

func TestEmails_OfficialEmail_OwnAddressAccepted(t *testing.T) {
	issues := checkEmails(emailTestHTML, "Relojistas@ContactForSales.com")

	if len(issues) != 0 {
		t.Fatalf("the site's own contact address (case-insensitive) must pass; issues=%+v", issues)
	}
}

func TestEmails_OfficialEmail_MismatchStillFlagged(t *testing.T) {
	issues := checkEmails(emailTestHTML, "hola@relojistas.com")

	var saw bool
	for _, is := range issues {
		if is.Value == "relojistas@contactforsales.com" {
			saw = true
			if is.Type != "invalid_email" || is.Severity != "error" || is.Expected != "hola@relojistas.com" {
				t.Errorf("expected mismatch invalid_email error carrying Expected, got %+v", is)
			}
		}
	}
	if !saw {
		t.Fatalf("a non-official asserted email must still be flagged as a mismatch; issues=%+v", issues)
	}
}

func TestEmails_NoOfficialEmail_PlaceholderStaysBlocker(t *testing.T) {
	issues := checkEmails(`<p>Mail us: <a href="mailto:info@example.com">info@example.com</a></p>`, "")

	var saw bool
	for _, is := range issues {
		if is.Value == "info@example.com" {
			saw = true
			if is.Type != "placeholder_email" || is.Severity != "blocker" {
				t.Errorf("placeholder classification must take precedence over the no-official-email branch, got %+v", is)
			}
		}
	}
	if !saw {
		t.Fatalf("placeholder email must still be flagged; issues=%+v", issues)
	}
}

func TestEmails_NoOfficialEmail_NonAssertionContextsStillIgnored(t *testing.T) {
	issues := checkEmails(`<input type="email" placeholder="you@yourcompany.com"><script>var e="x@tracker.io";</script>`, "")

	if len(issues) != 0 {
		t.Fatalf("emails in placeholder attributes / script bodies are examples, not contact claims — must not be flagged; issues=%+v", issues)
	}
}
