package actions

import "testing"

// These expectations mirror 006_unify_prices_schema.sql's slug computation:
//   regexp_replace(lower(a || '-' || b), '[^a-z0-9]+', '-', 'g')
// Go and SQL must produce identical slugs or migrated and live-written rows
// split into duplicate products. Do not "fix" trailing hyphens here — the
// SQL keeps them, so Go must too.
func TestOfferingSlugMatchesMigrationSemantics(t *testing.T) {
	cases := []struct {
		parts []string
		want  string
	}{
		// Example from HANDOFF_2026-05-18 §2 (service slug body)
		{[]string{"consultation", "15 Minute Consultation"}, "consultation-15-minute-consultation"},
		// Example from HANDOFF_2026-05-18 §2 (medicine slug)
		{[]string{"Apoquel", "3.6mg", "20 tablets"}, "apoquel-3-6mg-20-tablets"},
		// Real production values
		{[]string{"vaccination", "Booster Vac (Dog)"}, "vaccination-booster-vac-dog-"},
		{[]string{"diagnostic", "Six-month healthy pet check"}, "diagnostic-six-month-healthy-pet-check"},
		// Punctuation runs collapse to one hyphen; trailing runs keep theirs
		{[]string{"Emergency", "Out of Hours!!"}, "emergency-out-of-hours-"},
		// Empty middle part drops out entirely (no doubled separator)
		{[]string{"Metacam", "", "100ml"}, "metacam-100ml"},
	}
	for _, c := range cases {
		if got := offeringSlug(c.parts...); got != c.want {
			t.Errorf("offeringSlug(%q) = %q, want %q", c.parts, got, c.want)
		}
	}
}

func TestServiceSlugPrefix(t *testing.T) {
	got := "service-" + offeringSlug("consultation", "Consultation")
	if got != "service-consultation-consultation" {
		t.Fatalf("service slug = %q", got)
	}
}
