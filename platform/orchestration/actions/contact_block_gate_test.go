package actions

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

// contactBlockDetailsGated is the cb-details block of the `contact-block`
// component as migration 526 rewrites it: each cb-detail-item wrapped in an
// {{if}} on its OWN field.
//
// WHY THE GATE WRAPS THE WHOLE ITEM AND NOT THE VALUE. Each item is
// icon + label + value. Gating only the value leaves an icon and a heading
// standing over nothing — bugs_closed/111 exactly ("footer contact heading
// renders over an empty mailto"). The LANDMINE on on_missing/skip_field makes
// the same point generally: read what ENCLOSES the field and gate the smallest
// VALID unit, because a field in a fixed-arity row is either a no-op to gate or
// emits malformed HTML.
//
// MUTATION THAT MUST BREAK IT: drop the {{if .contact_phone}}/{{end}} pair —
// TestContactBlockGate_MissingPhoneRendersNothing then finds tel: back on the
// page, which is the live defect it was written for (6 rows, 3 sites,
// measured 2026-08-21).
const contactBlockDetailsGated = `      <div class="cb-details">
        {{if .contact_email}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22,6 12,13 2,6"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.email_label}}</div>
            <div class="cb-detail-value"><a href="mailto:{{.contact_email}}">{{.contact_email}}</a></div>
          </div>
        </div>{{end}}

        {{if .contact_phone}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M22 16.92v3a2 2 0 01-2.18 2 19.79 19.79 0 01-8.63-3.07A19.5 19.5 0 013.07 9.81 19.79 19.79 0 01.12 1.18 2 2 0 012.11 0h3a2 2 0 012 1.72 12.84 12.84 0 00.7 2.81 2 2 0 01-.45 2.11L6.09 7.91a16 16 0 006 6l1.27-1.27a2 2 0 012.11-.45 12.84 12.84 0 002.81.7A2 2 0 0122 14.92z"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.phone_label}}</div>
            <div class="cb-detail-value"><a href="tel:{{.contact_phone}}">{{.contact_phone}}</a></div>
          </div>
        </div>{{end}}

        {{if .contact_location}}<div class="cb-detail-item">
          <div class="cb-detail-icon" aria-hidden="true">
            <svg viewBox="0 0 24 24"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z"/><circle cx="12" cy="10" r="3"/></svg>
          </div>
          <div>
            <div class="cb-detail-label">{{.location_label}}</div>
            <div class="cb-detail-value">{{.contact_location}}</div>
          </div>
        </div>{{end}}
`

func TestContactBlockGate_AllPresentRendersAllThree(t *testing.T) {
	out, _, inURL, err := RenderTemplate(contactBlockDetailsGated, &RenderContext{
		ContentData: map[string]interface{}{
			"contact_email": "a@b.com", "email_label": "Email",
			"contact_phone": "+44 1234", "phone_label": "Phone",
			"contact_location": "London", "location_label": "Where",
		},
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("fixture must render: %v", err)
	}
	if got := strings.Count(out, "cb-detail-item"); got != 3 {
		t.Errorf("all three fields supplied: want 3 detail items, got %d", got)
	}
	if len(inURL) != 0 {
		t.Errorf("nothing should be reported dead when every field is supplied, got %v", inURL)
	}
}

func TestContactBlockGate_MissingPhoneRendersNothing(t *testing.T) {
	out, _, inURL, err := RenderTemplate(contactBlockDetailsGated, &RenderContext{
		ContentData: map[string]interface{}{
			"contact_email": "a@b.com", "email_label": "Email",
			"location_label": "Where", "contact_location": "London",
		},
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("fixture must render: %v", err)
	}

	// The live defect this gate removes.
	if strings.Contains(out, "tel:") {
		t.Errorf("a missing phone must render NO tel: control at all; got:\n%s", out)
	}
	// And it must not orphan the icon + label (bugs_closed/111).
	if got := strings.Count(out, "cb-detail-item"); got != 2 {
		t.Errorf("want exactly 2 detail items, got %d:\n%s", got, out)
	}
	// A gated field is invisible to the dead-URL report BY CONSTRUCTION. Asserted
	// so it is a known property of this component rather than a surprise when
	// dead_url_control goes quiet on it.
	if len(inURL) != 0 {
		t.Errorf("a gated field must not be reported as a dead URL, got %v", inURL)
	}
}

func TestContactBlockGate_AllMissingRendersEmptyDetails(t *testing.T) {
	out, _, _, err := RenderTemplate(contactBlockDetailsGated, &RenderContext{
		ContentData: map[string]interface{}{},
	}, zap.NewNop())
	if err != nil {
		t.Fatalf("fixture must render: %v", err)
	}
	if strings.Contains(out, "cb-detail-item") {
		t.Errorf("no fields supplied: want no detail items at all, got:\n%s", out)
	}
	if !strings.Contains(out, "cb-details") {
		t.Error("the surrounding cb-details container must survive — gating removes items, not the section")
	}
}
