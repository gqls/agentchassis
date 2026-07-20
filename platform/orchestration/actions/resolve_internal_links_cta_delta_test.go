// FILE: platform/orchestration/actions/resolve_internal_links_cta_delta_test.go
//
// Reachability proof for ctaDerivationDelta — written in direct answer to the
// council round-6 editquality objection that this comparison might be dead
// code by the same fresh-map failure that killed the plan_sections observe
// log (trail 2525f980). It is not: the two sides are INDEPENDENT sources —
// the static package-level ctaFieldNames map vs the schema-derived set — so
// the delta is non-empty exactly when their coverage differs, which is the
// observable the stage exists to measure. These tests exercise all four
// coverage relations, including the live case that motivates the stage
// (a component the map has never heard of).

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func TestCTADerivationDelta_IdenticalCoverageIsSilent(t *testing.T) {
	// hero per ctaFieldNames: {"cta_url", "secondary_cta_url"} — derived agrees.
	derived := []datahelpers.CTAField{
		{URLField: "cta_url", LabelField: "cta_text", Source: "renderer"},
		{URLField: "secondary_cta_url", LabelField: "secondary_cta", Source: "renderer"},
	}
	if d := ctaDerivationDelta(ctaFieldNames["hero"], true, derived); d != "" {
		t.Fatalf("identical coverage should be silent, got %q", d)
	}
}

func TestCTADerivationDelta_UnmappedComponentFires(t *testing.T) {
	// The motivating live case: a component outside the map entirely
	// (e.g. archetype-result-card) derives fields the map cannot repair.
	derived := []datahelpers.CTAField{
		{URLField: "cta_primary_url", LabelField: "cta_primary_label", Source: "site_specs.cta.primary_url"},
	}
	d := ctaDerivationDelta([2]string{}, false, derived)
	if d != "derived-not-mapped: cta_primary_url(site_specs.cta.primary_url)" {
		t.Fatalf("unmapped component must fire, got %q", d)
	}
}

func TestCTADerivationDelta_MappedNotDerivedFires(t *testing.T) {
	// The inverse gap: the map claims a field the schema no longer pairs —
	// exactly what the bare-stem correction fixed; keep it observable.
	d := ctaDerivationDelta(ctaFieldNames["hero"], true, []datahelpers.CTAField{
		{URLField: "cta_url", LabelField: "cta_text", Source: "renderer"},
	})
	if d != "mapped-not-derived: secondary_cta_url" {
		t.Fatalf("map-only field must fire, got %q", d)
	}
}

func TestCTADerivationDelta_BothDirectionsCompose(t *testing.T) {
	d := ctaDerivationDelta([2]string{"old_url", ""}, true, []datahelpers.CTAField{
		{URLField: "new_url", LabelField: "new_label", Source: "renderer"},
	})
	want := "derived-not-mapped: new_url(renderer); mapped-not-derived: old_url"
	if d != want {
		t.Fatalf("got %q want %q", d, want)
	}
}
