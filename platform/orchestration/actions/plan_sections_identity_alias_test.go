// FILE: platform/orchestration/actions/plan_sections_identity_alias_test.go
//
// Tests for the site_specs identity fallback chain (bugs_open/072 fix).
//
// The load-bearing property: a component field declaring the FLAT spec path
// (site_specs.identity.email — which is what every schema in the fleet declares)
// must resolve from the writer's nested shape (identity.contact.email) or from
// the canonical sites row (sites.email) when the flat key is absent, WITHOUT
// changing the value of any path that already resolves literally.
//
// DB-free: resolveSpecAlias's two steps are exercised by pre-seeding the
// resolver's caches and setting their loaded flags, so ensureSpecs/ensureSiteRow
// never run. The DB read itself (ensureSiteRow) is a single COALESCEd SELECT
// verified against the live schema.
package actions

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// newTestResolver builds a resolver with both caches pre-loaded, so no query runs.
func newTestResolver(identity map[string]interface{}, siteRow map[string]string) *sourceResolver {
	r := newSourceResolver(uuid.New(), nil, zap.NewNop(), "contact")
	if identity != nil {
		r.specs["identity"] = identity
	}
	r.specsLoaded = true
	r.siteRow = siteRow
	if r.siteRow == nil {
		r.siteRow = map[string]string{}
	}
	r.siteRowLoaded = true
	return r
}

func TestIdentityFlatPathResolvesFromNestedContactShape(t *testing.T) {
	// The 072 contract mismatch: domain-research-classifier writes contact
	// details nested; every component schema asks for them flat.
	r := newTestResolver(map[string]interface{}{
		"company_name": "Vonc",
		"contact": map[string]interface{}{
			"email":   "hello@vonc.com",
			"phone":   "07934 524 911",
			"address": "12 Example Street",
		},
	}, nil)

	for _, tc := range []struct{ path, want string }{
		{"identity.email", "hello@vonc.com"},
		{"identity.phone", "07934 524 911"},
		{"identity.address", "12 Example Street"},
	} {
		got, ok := r.resolve(context.Background(), "site_specs."+tc.path)
		if !ok {
			t.Fatalf("%s: expected to resolve via the nested contact shape, got not-found", tc.path)
		}
		if got != tc.want {
			t.Errorf("%s: got %v, want %q", tc.path, got, tc.want)
		}
	}
}

func TestIdentityFlatPathResolvesFromCanonicalSitesRow(t *testing.T) {
	// The real gap: on the eight sites that render no contact-info, the nested
	// keys exist but are EMPTY, while the sites row carries the address. This is
	// the store loadSiteDataFull treats as canonical and plan_sections could not
	// see at all.
	r := newTestResolver(map[string]interface{}{
		"contact": map[string]interface{}{"email": nil, "phone": nil},
	}, map[string]string{
		"email":           "vonc@contactforsales.com",
		"phone":           "+44 (0) 7934 524 911",
		"contact_address": "12 Example Street",
		"company_name":    "Vonc Ltd",
		"tagline":         "A tagline",
		"logo_text":       "VONC",
		"logo_url":        "/assets/logo.png",
	})

	for _, tc := range []struct{ path, want string }{
		{"identity.email", "vonc@contactforsales.com"},
		{"identity.phone", "+44 (0) 7934 524 911"},
		{"identity.address", "12 Example Street"},
		{"identity.contact_address", "12 Example Street"},
		{"identity.company_name", "Vonc Ltd"},
		{"identity.tagline", "A tagline"},
		{"identity.logo_text", "VONC"},
		{"identity.logo_url", "/assets/logo.png"},
	} {
		got, ok := r.resolve(context.Background(), "site_specs."+tc.path)
		if !ok {
			t.Fatalf("%s: expected to resolve from the sites row, got not-found", tc.path)
		}
		if got != tc.want {
			t.Errorf("%s: got %v, want %q", tc.path, got, tc.want)
		}
	}
}

// The safety property the whole change rests on: an alias may only ADD
// resolution. Anything resolving literally today must keep its literal value.
func TestLiteralSpecPathAlwaysWinsOverBothAliases(t *testing.T) {
	r := newTestResolver(map[string]interface{}{
		"email": "flat-wins@example.com",
		"contact": map[string]interface{}{
			"email": "nested-loses@example.com",
		},
	}, map[string]string{"email": "siterow-loses@example.com"})

	got, ok := r.resolve(context.Background(), "site_specs.identity.email")
	if !ok {
		t.Fatal("expected the literal flat path to resolve")
	}
	if got != "flat-wins@example.com" {
		t.Errorf("literal path must win: got %v", got)
	}
}

// Nested beats the sites row: spec data is the richer, site-specific store.
func TestNestedShapeBeatsSitesRow(t *testing.T) {
	r := newTestResolver(map[string]interface{}{
		"contact": map[string]interface{}{"email": "nested-wins@example.com"},
	}, map[string]string{"email": "siterow-loses@example.com"})

	got, _ := r.resolve(context.Background(), "site_specs.identity.email")
	if got != "nested-wins@example.com" {
		t.Errorf("nested shape must beat the sites row: got %v", got)
	}
}

// A missing fact must stay missing so on_missing governs. This is the guard
// against the fallback satisfying a needs_human_review field with a value
// nobody supplied — the reason ensureSiteRow does not COALESCE across columns
// the way loadSiteDataFull does.
func TestMissingIdentityFactStaysMissing(t *testing.T) {
	r := newTestResolver(map[string]interface{}{
		"company_name": "Games Design",
		"contact":      map[string]interface{}{"email": nil, "phone": nil, "address": nil},
	}, map[string]string{}) // sites row empty too — the three genuinely dataless sites

	for _, path := range []string{"identity.email", "identity.phone", "identity.address", "identity.hours"} {
		if val, ok := r.resolve(context.Background(), "site_specs."+path); ok {
			t.Errorf("%s: must NOT resolve when the fact is absent everywhere, got %v", path, val)
		}
	}
}

// The alias set is bounded on both axes: only enumerated leaves, only the
// identity aspect, and only two-segment paths.
func TestAliasChainIsBounded(t *testing.T) {
	r := newTestResolver(map[string]interface{}{
		"contact": map[string]interface{}{"secret": "should-not-surface"},
		"team":    map[string]interface{}{"members": []interface{}{"a"}},
	}, map[string]string{"email": "e@example.com", "company_name": "C"})

	// A leaf with no sites-row mapping and no nested twin does not resolve.
	if _, ok := r.resolve(context.Background(), "site_specs.identity.hours"); ok {
		t.Error("identity.hours has no nested twin and no column — must not resolve")
	}
	// A non-identity aspect never reaches the sites row.
	if _, ok := r.resolve(context.Background(), "site_specs.commercial.email"); ok {
		t.Error("a non-identity aspect must not resolve from the sites row")
	}
	// A deeper path is left entirely alone (no alias applied).
	if _, ok := r.resolve(context.Background(), "site_specs.identity.team.email"); ok {
		t.Error("a three-segment path must not be aliased")
	}
	// A nested key with no flat counterpart in the schema vocabulary still
	// resolves via the container — the container alias is by shape, not by an
	// allow-list of leaves. Documents the deliberate asymmetry with step 2.
	if _, ok := r.resolve(context.Background(), "site_specs.identity.secret"); !ok {
		t.Error("the nested container alias applies to any leaf inside it")
	}
}

// Every aliased resolution must land in aliasesUsed, which the action surfaces as
// source_aliases_used. Asked for by the council gate's bug_historian seat (corr
// dd03a73b): a zap line is not a queryable record of a section whose data
// provenance changed. The map must stay EMPTY when everything resolved literally,
// so its presence in a build record is itself the signal.
func TestAliasedResolutionsAreRecordedStructurally(t *testing.T) {
	r := newTestResolver(map[string]interface{}{
		"company_name": "Literal Co",
		"contact":      map[string]interface{}{"phone": "07934 524 911"},
	}, map[string]string{"email": "vonc@contactforsales.com"})

	r.resolve(context.Background(), "site_specs.identity.company_name") // literal
	r.resolve(context.Background(), "site_specs.identity.phone")        // nested alias
	r.resolve(context.Background(), "site_specs.identity.email")        // sites-row alias
	r.resolve(context.Background(), "site_specs.identity.hours")        // resolves nowhere

	want := map[string]string{
		"site_specs.identity.phone": "site_specs.identity.contact.phone",
		"site_specs.identity.email": "sites.email",
	}
	if len(r.aliasesUsed) != len(want) {
		t.Fatalf("got %v, want %v", r.aliasesUsed, want)
	}
	for k, v := range want {
		if r.aliasesUsed[k] != v {
			t.Errorf("%s: got %q, want %q", k, r.aliasesUsed[k], v)
		}
	}
	// A literal resolve must not be recorded — otherwise the signal is noise.
	if _, recorded := r.aliasesUsed["site_specs.identity.company_name"]; recorded {
		t.Error("a literal resolution must NOT be recorded as an alias")
	}
}

func TestNothingRecordedWhenEverythingResolvesLiterally(t *testing.T) {
	r := newTestResolver(map[string]interface{}{
		"email": "flat@example.com", "phone": "123", "company_name": "C",
	}, map[string]string{"email": "unused@example.com"})

	for _, p := range []string{"identity.email", "identity.phone", "identity.company_name"} {
		r.resolve(context.Background(), "site_specs."+p)
	}
	if len(r.aliasesUsed) != 0 {
		t.Errorf("expected no aliases recorded, got %v", r.aliasesUsed)
	}
}

// A site with no identity aspect at all (loancalculator.co.uk) must still reach
// the sites row — the aspect being absent is not a reason to skip the canonical
// store.
func TestSitesRowReachableWithNoIdentityAspect(t *testing.T) {
	r := newTestResolver(nil, map[string]string{"email": "owner@example.com"})

	got, ok := r.resolve(context.Background(), "site_specs.identity.email")
	if !ok {
		t.Fatal("expected the sites row to resolve with no identity aspect present")
	}
	if got != "owner@example.com" {
		t.Errorf("got %v", got)
	}
}
