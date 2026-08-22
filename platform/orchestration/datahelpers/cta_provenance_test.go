// FILE: platform/orchestration/datahelpers/cta_provenance_test.go
//
// Pins the properties bugs_open/308's fix rests on. Each test names the
// mutation that must kill it, because on this estate a guard is not established
// by reasoning about it — LNK-034's own seam was mutation-proven and that is the
// bar here.

package datahelpers

import "testing"

// TestAbsentStampCoversNothing is the NO-BACKFILL guarantee: every row in the
// fleet is unstamped on the day this ships, and each one must behave exactly as
// it did before the mechanism existed.
//
// MUTATION: make CTAMintedCovers return true for a nil/missing record. This
// fails — and so would the whole fleet, by freezing every CTA at once.
func TestAbsentStampCoversNothing(t *testing.T) {
	cases := []struct {
		name string
		cd   map[string]interface{}
	}{
		{"nil content_data", nil},
		{"no stamp key", map[string]interface{}{"cta_url": "/contact.html"}},
		{"stamp present but empty", map[string]interface{}{
			CTAMintedKey: map[string]interface{}{},
		}},
		{"stamp key holds a non-map", map[string]interface{}{
			CTAMintedKey: "not-a-map",
		}},
		{"stamp entry holds a non-string", map[string]interface{}{
			CTAMintedKey: map[string]interface{}{"cta_url": 42},
		}},
	}
	for _, c := range cases {
		if CTAMintedCovers(c.cd, "cta_url", "/contact.html") {
			t.Errorf("%s: an absent or unusable record must cover nothing", c.name)
		}
	}
}

// TestStaleStampDoesNotCoverEditedValue is THE value-binding test — the reason
// the stamp records a url rather than a boolean.
//
// The scenario is the section editor's field update (a MERGE writer): a person
// changes the destination, the old stamp survives the merge. Value-bound, it
// names a url the field no longer carries, so it does not cover and the edit
// reads as authored. Presence-bound, the surviving stamp would licence the
// recompute to overwrite the person's choice — which is bugs_open/248, the bug
// the keep branches exist to prevent.
//
// MUTATION: drop the NormalizePagePath comparison in CTAMintedCovers and return
// true whenever an entry exists. This fails.
func TestStaleStampDoesNotCoverEditedValue(t *testing.T) {
	cd := map[string]interface{}{
		"cta_url": "/contact.html", // the human's new value
		CTAMintedKey: map[string]interface{}{
			"cta_url": "/tools/password-entropy.html", // what the resolver had minted
		},
	}
	if CTAMintedCovers(cd, "cta_url", "/contact.html") {
		t.Fatal("a stamp naming a different url must not cover the current value — " +
			"a presence-bound stamp here is bugs_open/248's clobber")
	}
}

// TestStampCoversItsOwnValueAcrossPathForms pins that the record cannot be
// defeated by the spellings of ONE page that NormalizePagePath equates —
// directory form, explicit index.html, trailing slash, and a query/fragment
// suffix. All four are live in the fleet.
//
// ⚠ AND IT PINS THE BOUNDARY, which I got wrong first time and the test caught:
// NormalizePagePath does NOT equate "/contact.html" with "/contact/". It trims
// a trailing "index.html" and trailing slashes, so "/contact/index.html" and
// "/contact/" both become "/contact" — while "/contact.html" stays
// "/contact.html". Those are two DIFFERENT pages to this codebase (only
// ctaExcludedDestination collapses them, and only to decide the AREA). A stamp
// must therefore not vouch across that boundary: if it did, minting
// "/contact/" would silently mark an authored "/contact.html" as the resolver's
// own and make it recomputable.
//
// MUTATION: compare the raw strings instead of NormalizePagePath. The
// equivalent-forms half fails, and the consequence would be a resolver-minted
// link reading as authored purely because the stored spelling differs — a
// silent freeze.
func TestStampCoversItsOwnValueAcrossPathForms(t *testing.T) {
	cd := map[string]interface{}{
		CTAMintedKey: map[string]interface{}{"cta_url": "/contact/index.html"},
	}
	for _, form := range []string{"/contact/index.html", "/contact/", "/contact", "/contact/?utm=x#top"} {
		if !CTAMintedCovers(cd, "cta_url", form) {
			t.Errorf("stamp /contact/index.html must cover the equivalent form %q", form)
		}
	}
	// The boundary: a different page, not a different spelling.
	if CTAMintedCovers(cd, "cta_url", "/contact.html") {
		t.Error("/contact.html is a DIFFERENT page from /contact/index.html under NormalizePagePath — " +
			"a stamp must not vouch across that boundary")
	}
}

// TestStampIsPerField — the record is one nested map holding both slots of a
// two-CTA component, so a whole-map presence test would vouch for a field
// nobody stamped.
//
// MUTATION: have CTAMintedCovers ignore `field` and test only that the record
// is non-empty. This fails.
func TestStampIsPerField(t *testing.T) {
	cd := map[string]interface{}{
		CTAMintedKey: map[string]interface{}{"cta_url": "/tools/a.html"},
	}
	if CTAMintedCovers(cd, "secondary_cta_url", "/tools/a.html") {
		t.Fatal("a stamp on cta_url must not vouch for secondary_cta_url")
	}
}

// TestSeedCTAMintedPreservesTheSiblingSlot is the defect I introduced and then
// caught: both persist paths merge SHALLOWLY, and the record is a nested map,
// so a resolved_data stamping only the primary REPLACES the stored record and
// drops the secondary's stamp — after which the secondary reads as authored and
// freezes. Every ctaFieldNames component but two has both slots.
//
// MUTATION: make this function a no-op. This fails. The CALL SITES are pinned
// separately, in the actions package — both live INSIDE setCTAField and
// applyCTARecompute rather than in the callers' loops, because a loop-level
// call proved deletable with every test in the tree still green.
func TestSeedCTAMintedPreservesTheSiblingSlot(t *testing.T) {
	stored := map[string]interface{}{
		CTAMintedKey: map[string]interface{}{
			"cta_url":           "/tools/a.html",
			"secondary_cta_url": "/tools/b.html",
		},
	}
	resolved := map[string]interface{}{}

	SeedCTAMinted(resolved, stored)
	SetCTAMinted(resolved, "cta_url", "/tools/c.html") // primary re-minted this pass

	if !CTAMintedCovers(resolved, "secondary_cta_url", "/tools/b.html") {
		t.Error("the untouched sibling slot lost its stamp — it will read as authored and freeze")
	}
	if !CTAMintedCovers(resolved, "cta_url", "/tools/c.html") {
		t.Error("the re-minted slot must carry its NEW value, not the seeded one")
	}
	if CTAMintedCovers(resolved, "cta_url", "/tools/a.html") {
		t.Error("the seeded entry must be overwritten by this pass's mint, not merged alongside it")
	}
}

// TestSeedCTAMintedDoesNotMutateStored — both writers keep reading the stored
// content_data after the seed (the keeps consult it per field), so seeding must
// not write through into the caller's map.
//
// MUTATION: assign resolved[CTAMintedKey] = storedMinted directly instead of
// copying. This fails.
func TestSeedCTAMintedDoesNotMutateStored(t *testing.T) {
	stored := map[string]interface{}{
		CTAMintedKey: map[string]interface{}{"cta_url": "/tools/a.html"},
	}
	resolved := map[string]interface{}{}

	SeedCTAMinted(resolved, stored)
	SetCTAMinted(resolved, "cta_url", "/tools/c.html")

	if !CTAMintedCovers(stored, "cta_url", "/tools/a.html") {
		t.Fatal("seeding mutated the caller's stored content_data — the keeps read that map afterwards")
	}
}

// TestSetCTAMintedRefusesEmpties — an empty entry would be indistinguishable
// from a real mint of "" and would make CTAMintedCovers' own empty-check the
// only thing standing between the fleet and a bad record.
func TestSetCTAMintedRefusesEmpties(t *testing.T) {
	resolved := map[string]interface{}{}
	SetCTAMinted(resolved, "", "/contact.html")
	SetCTAMinted(resolved, "cta_url", "")
	if _, present := resolved[CTAMintedKey]; present {
		t.Fatal("an empty field or url must not create a record")
	}
	SetCTAMinted(nil, "cta_url", "/contact.html") // must not panic
}
