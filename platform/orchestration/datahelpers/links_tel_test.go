// FILE: platform/orchestration/datahelpers/links_tel_test.go
//
// Pins the non-page CTA destination vocabulary (bugs_open/299). The tel: rows
// include every value live in the fleet on 2026-08-18 — four malformed
// spaces-and-parens forms and the one collapsed-trunk undialable — so the
// normaliser is calibrated against the population it exists for, not invented
// examples. The "+440 refused" case is the load-bearing one: the obvious
// clean-up (strip separators) MANUFACTURES that undialable form, so a
// normaliser that accepted it would certify its own failure mode.

package datahelpers

import "testing"

func TestIsAuthoredNonPageCTADestination(t *testing.T) {
	tests := []struct {
		href string
		want bool
	}{
		// the motivating class
		{"tel:+44 (0) 7934 524 911", true},
		{"tel:+447934524911", true},
		{"mailto:hello@webdesign.uk", true},
		{"mailto:hello@x.uk?subject=Hi", true},
		{"https://example.org/booking", true},
		{"http://example.org", true},
		{"//cdn.example.org/page", true},
		{"#pricing", true}, // named fragment
		// excluded: dead controls are check_dead_controls' remit
		{"#", false},
		{"#!", false},
		{"javascript:void(0)", false},
		{"javascript:run()", false},
		// excluded: pages/assets/empty are the page machinery's remit
		{"", false},
		{"/contact.html", false}, // BOUNDARY PIN — the page-scheme keep is
		// bugs_open/248's storedCTADestinationIsAuthored; no url may satisfy both
		{"/tools/website-brief-starter/index.html", false},
		{"/assets/x.pdf", false},
	}
	for _, tt := range tests {
		if got := IsAuthoredNonPageCTADestination(tt.href); got != tt.want {
			t.Errorf("IsAuthoredNonPageCTADestination(%q) = %v, want %v", tt.href, got, tt.want)
		}
	}
}

func TestNormalizeTelHref(t *testing.T) {
	tests := []struct {
		name string
		href string
		want string
		ok   bool
	}{
		// the four live malformed forms (one spelling, four rows fleet-wide)
		{"live malformed: spaces+parens trunk", "tel:+44 (0) 7934 524 911", "tel:+447934524911", true},
		// the live undialable form: separators were stripped by hand and the
		// trunk zero collapsed in — REFUSED, never "repaired" by guessing
		{"live undialable: collapsed trunk", "tel:+4407934524911", "", false},
		// national form: no "+" means no trunk rule — the zero is real
		{"national with spaces", "tel:020 7946 0000", "tel:02079460000", true},
		{"separators only", "tel:+44-7934-524-911", "tel:+447934524911", true},
		{"dots", "tel:+1.234.567.8901", "tel:+12345678901", true},
		{"already normal", "tel:+447934524911", "tel:+447934524911", true},
		{"case-insensitive scheme", "TEL:+447934524911", "tel:+447934524911", true},
		{"no-space parens trunk", "tel:+44(0)7934524911", "tel:+447934524911", true},
		// refusals: not tel, empty, junk, extensions, out of range
		{"mailto is not tel", "mailto:x@y", "", false},
		{"empty number", "tel:", "", false},
		{"letters", "tel:+44 CALL-ME", "", false},
		{"extension param", "tel:+447934524911;ext=2", "", false},
		{"plus not at front", "tel:44+7934", "", false},
		{"too short", "tel:123", "", false},
		{"too long", "tel:+1234567890123456", "", false},
	}
	for _, tt := range tests {
		got, ok := NormalizeTelHref(tt.href)
		if ok != tt.ok || got != tt.want {
			t.Errorf("%s: NormalizeTelHref(%q) = (%q, %v), want (%q, %v)",
				tt.name, tt.href, got, ok, tt.want, tt.ok)
		}
	}
}

func TestDescribeCTADestination(t *testing.T) {
	tests := []struct {
		href string
		want string
	}{
		// the authored display form survives — a visitor reads "+44 (0) …"
		{"tel:+44 (0) 7934 524 911", "a phone call to +44 (0) 7934 524 911"},
		{"mailto:hello@webdesign.uk", "an email to hello@webdesign.uk"},
		{"mailto:hello@x.uk?subject=Hi", "an email to hello@x.uk"},
		{"https://example.org/booking?x=1", "an external site: example.org"},
		{"//cdn.example.org/page", "an external site: cdn.example.org"},
		{"#pricing", "a section on this page (#pricing)"},
		// not ours: callers gate on IsAuthoredNonPageCTADestination first
		{"/contact.html", ""},
		{"javascript:void(0)", ""},
		{"#", ""},
	}
	for _, tt := range tests {
		if got := DescribeCTADestination(tt.href); got != tt.want {
			t.Errorf("DescribeCTADestination(%q) = %q, want %q", tt.href, got, tt.want)
		}
	}
}
