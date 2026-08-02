// FILE: platform/orchestration/actions/contact_info_no_fabrication_test.go
//
// bugs_open/140 — the shared `contact-info` component fabricated a phone number,
// office hours and an email address when the site's datum was absent. 8 of 8 live
// uses served invented hours; vetcomparison.uk also served `tel:+1234567890`,
// confirmed on the wire 2026-08-02.
//
// The template now obeys the contract its own input_schema already published
// (`"on_missing": "skip_field"` for phone/hours/address). This test proves that
// end to end, through the REAL render path — executeGoTemplate, the same
// text/template configuration with `missingkey=zero` and the same funcMap that
// RenderTemplateReportingMissing uses in production.
//
// WHY IT READS THE MIGRATION RATHER THAN A COPY. The template lives in the
// database, and the file that put it there is
// docs/agent_docs/sql_for_agents/287_contact_info_obeys_its_own_schema.sql. A
// const in this file would be a second hand-maintained copy of one contract —
// exactly the drift class this repo reviews for — and it would keep passing after
// someone edited the migration. So the test parses the $mig$…$mig$ body out of the
// migration and renders THAT. Same approach as write_experience_pattern_test.go
// and doc_subjects_common_test.go, which read their own migrations.
//
// Limitation, stated rather than implied: this proves the template we SHIPPED is
// correct. It cannot see a later hand-edit of the live row. The standing lint
// scripts/check_placeholder_fallbacks.py is what watches the live library.

package actions

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

// The fabricated literals. Each one is a business fact the platform asserted on a
// live commercial site that nobody had ever stated.
var contactInfoFabrications = []string{
	"+1234567890",
	"+1 (234) 567-890",
	"Monday", // the invented "Monday – Friday, 9am – 6pm" hours string
	"9am",
	"info@example.com",
}

// contactInfoShippedTemplate returns the html_template as shipped by migration 287.
func contactInfoShippedTemplate(t *testing.T) string {
	t.Helper()

	glob := filepath.Join("..", "..", "..", "docs", "agent_docs", "sql_for_agents",
		"*_contact_info_obeys_its_own_schema.sql")
	matches, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("globbing for the contact-info migration: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected exactly 1 contact-info migration matching %s, found %d: %v",
			glob, len(matches), matches)
	}

	body, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatalf("reading %s: %v", matches[0], err)
	}

	const marker = "$mig$"
	first := strings.Index(string(body), marker)
	if first < 0 {
		t.Fatalf("%s has no %s-quoted template body", matches[0], marker)
	}
	rest := string(body)[first+len(marker):]
	last := strings.Index(rest, marker)
	if last < 0 {
		t.Fatalf("%s has an unterminated %s block", matches[0], marker)
	}
	tpl := rest[:last]

	if !strings.Contains(tpl, `data-component="contact-info"`) {
		t.Fatalf("extracted body does not look like the contact-info template:\n%.200s", tpl)
	}
	return tpl
}

func renderContactInfo(t *testing.T, tpl string, data map[string]interface{}) string {
	t.Helper()
	out, err := executeGoTemplate(tpl, data, zap.NewNop())
	if err != nil {
		t.Fatalf("executeGoTemplate failed — the template would fall through to the regex\n"+
			"fallback, which handles only handlebars {{#if}} and would ship this markup raw: %v", err)
	}
	return out
}

func countContactCards(html string) int {
	return strings.Count(html, `class="contact-card"`)
}

// assertNoFabrication is the load-bearing assertion. It is proved to detect its own
// defect by TestContactInfoFabricationAssertionsDetectThePreFixTemplate below.
func assertNoFabrication(t *testing.T, what, html string) {
	t.Helper()
	for _, lit := range contactInfoFabrications {
		if strings.Contains(html, lit) {
			t.Errorf("%s: rendered a fabricated contact fact %q — this is bugs_open/140", what, lit)
		}
	}
}

func TestContactInfoRendersOnlyWhatTheSiteSupplied(t *testing.T) {
	tpl := contactInfoShippedTemplate(t)

	t.Run("every datum supplied renders every card", func(t *testing.T) {
		html := renderContactInfo(t, tpl, map[string]interface{}{
			"section_title": "Get in Touch",
			"intro_text":    "We reply within one working day.",
			"email":         "hello@example-site.co.uk",
			"phone":         "+44 (0) 7934 524 911",
			"address":       "12 Example Street, Leeds",
			"hours":         "Mon-Thu 8-4",
		})
		if got := countContactCards(html); got != 4 {
			t.Errorf("expected 4 cards with every datum supplied, got %d\n%s", got, html)
		}
		for _, want := range []string{
			"Get in Touch", "We reply within one working day.",
			"hello@example-site.co.uk", "+44 (0) 7934 524 911",
			"12 Example Street, Leeds", "Mon-Thu 8-4",
		} {
			if !strings.Contains(html, want) {
				t.Errorf("supplied datum %q did not reach the page", want)
			}
		}
		assertNoFabrication(t, "full data", html)
	})

	// The bug file's own acceptance criterion: "check a site supplying ONLY email
	// renders exactly one card". This is the shape 6 of 8 live sites are in for
	// hours, and vetcomparison.uk/idea.uk for phone.
	t.Run("email only renders exactly one card and invents nothing", func(t *testing.T) {
		html := renderContactInfo(t, tpl, map[string]interface{}{
			"section_title": "Contact VetComparison.uk",
			"email":         "hello@vetcomparison.uk",
		})
		if got := countContactCards(html); got != 1 {
			t.Errorf("a site supplying only email must render exactly 1 card, got %d\n%s", got, html)
		}
		if strings.Contains(html, "tel:") {
			t.Error("rendered a tel: link for a site that supplied no phone — bugs_open/140")
		}
		if strings.Contains(html, "<h3>Hours</h3>") {
			t.Error("rendered an Hours card for a site that supplied no hours — bugs_open/140")
		}
		assertNoFabrication(t, "email only", html)
	})

	// hours is supplied by 0 of 1,089 page_components fleet-wide, so this is the
	// shape EVERY live use is in.
	t.Run("email and phone but no hours renders no Hours card", func(t *testing.T) {
		html := renderContactInfo(t, tpl, map[string]interface{}{
			"email": "hello@finetuning.uk",
			"phone": "+44 (0) 7934 524 911",
		})
		if got := countContactCards(html); got != 2 {
			t.Errorf("expected exactly 2 cards (email, phone), got %d\n%s", got, html)
		}
		if strings.Contains(html, "<h3>Hours</h3>") {
			t.Error("rendered an Hours card with no hours datum — bugs_open/140")
		}
		assertNoFabrication(t, "no hours", html)
	})

	// bugs_open/111's container rule, applied to the section component: furniture
	// renders only when it has contents.
	t.Run("no contact datum at all renders no grid", func(t *testing.T) {
		html := renderContactInfo(t, tpl, map[string]interface{}{
			"section_title": "Contact",
		})
		if countContactCards(html) != 0 {
			t.Errorf("expected no cards when nothing is supplied\n%s", html)
		}
		if strings.Contains(html, `class="contact-grid"`) {
			t.Error("rendered an empty .contact-grid shell — 111's container rule says gate it")
		}
		assertNoFabrication(t, "nothing supplied", html)
	})

	// The keys absent ENTIRELY, not merely empty — this is what the live rows look
	// like (`content_data ? 'hours'` is false, the key does not exist) and it is the
	// path that exercises missingkey=zero.
	t.Run("absent keys behave like absent, not like a default", func(t *testing.T) {
		html := renderContactInfo(t, tpl, map[string]interface{}{
			"email": "hello@idea.uk",
		})
		assertNoFabrication(t, "keys absent entirely", html)
		if got := countContactCards(html); got != 1 {
			t.Errorf("expected 1 card, got %d", got)
		}
	})

	// The desync half: all 8 live sites supply section_title and intro_text, and the
	// pre-fix template read .title/.intro, so it discarded every one of them.
	t.Run("the schema-declared heading and intro reach the page", func(t *testing.T) {
		html := renderContactInfo(t, tpl, map[string]interface{}{
			"section_title": "Reach us directly",
			"intro_text":    "Leopardess Consulting, by email.",
			"email":         "hello@leopardessconsulting.co.uk",
		})
		if !strings.Contains(html, "Reach us directly") {
			t.Error("section_title did not reach the page — the pre-fix template read .title and discarded it")
		}
		if !strings.Contains(html, "Leopardess Consulting, by email.") {
			t.Error("intro_text did not reach the page — the pre-fix template read .intro and discarded it")
		}
		if strings.Contains(html, "Contact Information") {
			t.Error("fell back to the hardcoded heading despite section_title being supplied")
		}
	})
}

// contactInfoPreFixTemplate is the template as it stood before migration 287, kept
// verbatim and FROZEN. It is not a maintained second copy — it is the defect, kept
// so the assertions above can be proved to detect it. Without this, every
// assertNoFabrication call could be passing vacuously.
const contactInfoPreFixTemplate = `<section class="contact-info-section" data-component="contact-info">
    <div class="contact-grid">
        <div class="contact-card">
            <h3>Email</h3>
            <a href="mailto:{{if .email}}{{.email}}{{else}}info@example.com{{end}}">{{if .email}}{{.email}}{{else}}info@example.com{{end}}</a>
        </div>
        <div class="contact-card">
            <h3>Phone</h3>
            <a href="tel:{{if .phone}}{{.phone}}{{else}}+1234567890{{end}}">{{if .phone_display}}{{.phone_display}}{{else if .phone}}{{.phone}}{{else}}+1 (234) 567-890{{end}}</a>
        </div>
        <div class="contact-card">
            <h3>Hours</h3>
            <p>{{if .hours}}{{.hours}}{{else}}Monday – Friday, 9am – 6pm{{end}}</p>
        </div>
    </div>
</section>`

// Mutation proof. Render the PRE-FIX template through the same path with the same
// email-only data, and require the assertions to fire. A guard that cannot be shown
// to fail on the real defect is not a guard.
func TestContactInfoFabricationAssertionsDetectThePreFixTemplate(t *testing.T) {
	html := renderContactInfo(t, contactInfoPreFixTemplate, map[string]interface{}{
		"email": "hello@vetcomparison.uk",
	})

	// The defect, reproduced: three cards where one datum was supplied.
	if got := countContactCards(html); got != 3 {
		t.Fatalf("pre-fix template should render all 3 cards unconditionally, got %d", got)
	}

	var fired []string
	for _, lit := range contactInfoFabrications {
		if strings.Contains(html, lit) {
			fired = append(fired, lit)
		}
	}
	// The email IS supplied here, so info@example.com correctly does not appear;
	// the phone and hours fabrications must.
	for _, want := range []string{"+1234567890", "+1 (234) 567-890", "Monday", "9am"} {
		if !strings.Contains(html, want) {
			t.Errorf("pre-fix template did not reproduce the fabrication %q — this test can no "+
				"longer prove the assertions above are load-bearing", want)
		}
	}
	if len(fired) == 0 {
		t.Fatal("assertNoFabrication's literal set matched nothing in the known-bad template — " +
			"every other assertion in this file may be passing vacuously")
	}
	t.Logf("pre-fix template reproduced %d fabricated literals: %v", len(fired), fired)
}
