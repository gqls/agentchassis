// FILE: platform/storage/brand_head_paths_test.go
//
// Pins the reason BrandHeadAssetPaths is DECLARED rather than derived
// (bugs_open/142), and — since bugs_open/168 — that the shared derivation
// consults it instead of competing with it.

package storage

import "testing"

// TestDeployedWebPathExpressesBrandHeadPaths is the INVERSION of
// TestDeployedWebPathCannotExpressBrandHeadPaths, performed 2026-08-02 at that
// test's own instruction.
//
// The 142 lane wrote it as a tripwire rather than a veto: it asserted that no
// argument pair yielded /assets/images/og-card.png, and its failure message said
// "Collapse BrandHeadAssetPaths into it and delete this test, or the platform has
// two answers to one question." bugs_open/168 did exactly that — the map is now
// an INPUT to storage.DeployedAssetPath rather than a rival answer — so the
// assertion flips from "cannot express" to "must express".
//
// The measurement the old test protected is not lost; it is preserved in
// BrandHeadAssetPaths' own comment, which records what the helper returned before
// the fix. This test now guards the other direction: if someone reverts the
// helper to purpose-derived spelling, the og card and favicon of every site in
// the fleet start resolving to a path nothing serves, and this fails.
func TestDeployedWebPathExpressesBrandHeadPaths(t *testing.T) {
	const wantOG = "/assets/images/og-card.png"

	if got := BrandHeadAssetPaths["og_card"]; got != wantOG {
		t.Fatalf("BrandHeadAssetPaths[og_card] = %q, want %q — this is the path "+
			"derive_brand_head_assets commits and injectBrandHeadTags emits", got, wantOG)
	}

	// og_card is the case that used to be inexpressible: the `_`→`-` swap in
	// AssetKeyFilename fires only when assetKey differs from purpose, and for
	// this artefact it does not. Every way of asking must now get it right,
	// including the empty asset_key a row written by the deriver carries.
	for _, c := range []struct{ assetKey, purpose string }{
		{"og_card", "og_card"},
		{"", "og_card"},
	} {
		if got := DeployedWebPath(c.assetKey, c.purpose); got != wantOG {
			t.Errorf("DeployedWebPath(%q, %q) = %q, want %q — the helper has stopped consulting "+
				"BrandHeadAssetPaths, so every site's og:image now points at a path nothing serves",
				c.assetKey, c.purpose, got, wantOG)
		}
	}

	// favicon agreed by coincidence before the fix (no underscore to disagree
	// about) and must agree by construction now. Saying so stops someone
	// concluding from favicon alone that the purpose-derived spelling is fine.
	if got := DeployedWebPath("favicon", "favicon"); got != BrandHeadAssetPaths["favicon"] {
		t.Errorf("DeployedWebPath(favicon, favicon) = %q but the map says %q", got, BrandHeadAssetPaths["favicon"])
	}

	// The purpose is what selects the brand-head spelling, not the asset_key —
	// a row whose purpose is empty is not a brand-head row however it is keyed.
	if got := DeployedWebPath("og_card", ""); got == wantOG {
		t.Errorf("DeployedWebPath(og_card, \"\") = %q — an empty purpose must NOT be treated as "+
			"brand-head; the purpose column is what identifies the writer", got)
	}
}

// TestDeployedAssetPathFormsAreConsistent pins the three path forms against each
// other. The deployer commits FilePath and the renderers reference RelativeURL;
// if those two ever disagree about the leading slash or the directory, a file is
// written where nothing looks for it — and both halves would still look right in
// isolation.
func TestDeployedAssetPathFormsAreConsistent(t *testing.T) {
	for _, c := range []struct{ assetKey, purpose string }{
		{"hero_home", "hero"},
		{"logo", "logo"},
		{"og_card", "og_card"},
		{"favicon", "favicon"},
		{"", "hero"},
		{"sprite_sheet_main", "sprite_sheet"},
		{"some_thing", "mystery"},
	} {
		p := DeployedAssetPath(c.assetKey, c.purpose)
		if p.RelativeURL != "/"+p.FilePath {
			t.Errorf("DeployedAssetPath(%q,%q): RelativeURL %q is not \"/\"+FilePath %q",
				c.assetKey, c.purpose, p.RelativeURL, p.FilePath)
		}
		if want := DefaultAssetBasePath + "/" + p.Filename; p.FilePath != want {
			t.Errorf("DeployedAssetPath(%q,%q): FilePath %q does not end in Filename %q under %q",
				c.assetKey, c.purpose, p.FilePath, p.Filename, DefaultAssetBasePath)
		}
		if DeployedWebPath(c.assetKey, c.purpose) != p.RelativeURL {
			t.Errorf("DeployedWebPath(%q,%q) disagrees with DeployedAssetPath().RelativeURL",
				c.assetKey, c.purpose)
		}
	}
}

// TestBrandHeadPurposesAreRealImagePurposes stops the map drifting away from the
// purpose vocabulary the rest of the platform resizes against.
func TestBrandHeadPurposesAreRealImagePurposes(t *testing.T) {
	for purpose := range BrandHeadAssetPaths {
		if _, ok := ImagePurposes[purpose]; !ok {
			t.Errorf("brand-head purpose %q is not in ImagePurposes — GetImageConfig would silently "+
				"fall back to the 1200x800 jpg default for it", purpose)
		}
		if !IsBrandHeadPurpose(purpose) {
			t.Errorf("IsBrandHeadPurpose(%q) is false for a key of its own map", purpose)
		}
	}
	if IsBrandHeadPurpose("hero") {
		t.Error("IsBrandHeadPurpose(hero) is true — a page-referenced purpose would be excluded " +
			"from the page-asset half of check_undeployed_assets and never checked at all")
	}
}
