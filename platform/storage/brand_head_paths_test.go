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
		// The shared-directory join is asserted only for DERIVED purposes. A
		// brand-head entry's path comes from the map and is not required to
		// live under DefaultAssetBasePath — see
		// TestBrandHeadPathsAreTakenWholeNotReconstructed.
		if !IsBrandHeadPurpose(c.purpose) {
			if want := DefaultAssetBasePath + "/" + p.Filename; p.FilePath != want {
				t.Errorf("DeployedAssetPath(%q,%q): FilePath %q does not end in Filename %q under %q",
					c.assetKey, c.purpose, p.FilePath, p.Filename, DefaultAssetBasePath)
			}
		}
		if DeployedWebPath(c.assetKey, c.purpose) != p.RelativeURL {
			t.Errorf("DeployedWebPath(%q,%q) disagrees with DeployedAssetPath().RelativeURL",
				c.assetKey, c.purpose)
		}
	}
}

// TestDeployedAssetPathAgreesWithTheMapLiteral closes the gap the council's
// edit-quality seat named on round 1 (`abd9b119`): the forms test above only
// checks the three forms against EACH OTHER, so a derivation that mangled a
// brand-head path would still satisfy it as long as it mangled it consistently.
// This asserts agreement with the map's own literal, which is the thing the
// deriver actually commits and `injectBrandHeadTags` actually emits.
func TestDeployedAssetPathAgreesWithTheMapLiteral(t *testing.T) {
	for purpose, published := range BrandHeadAssetPaths {
		if got := DeployedWebPath(purpose, purpose); got != published {
			t.Errorf("DeployedWebPath(%q,%q) = %q but BrandHeadAssetPaths says %q — the derivation "+
				"has stopped agreeing with the one declaration of this path", purpose, purpose, got, published)
		}
		if got := DeployedWebPath("", purpose); got != published {
			t.Errorf("DeployedWebPath(\"\",%q) = %q but BrandHeadAssetPaths says %q — a row written "+
				"with no asset_key must resolve the same way", purpose, got, published)
		}
	}
}

// TestBrandHeadPathsAreTakenWholeNotReconstructed pins the fix for the
// edit-quality seat's round-1 objection.
//
// The first version of brandHeadAssetPathsFor lifted the FILENAME out of the
// map's value and re-joined it under DefaultAssetBasePath. That is correct for
// both entries that exist today and silently wrong for the first entry that is
// not served from the shared asset directory — and a favicon at the site root
// (`/favicon.ico`) is a common enough convention that this was a real trap
// rather than a hypothetical one. No test then in the change could have caught
// it, because both existing entries DO live under that directory.
//
// So this test adds one that does not. It mutates the package map, which is
// safe here only because nothing in this package calls t.Parallel(); restore is
// via t.Cleanup so a failure cannot leak the entry into a later test.
func TestBrandHeadPathsAreTakenWholeNotReconstructed(t *testing.T) {
	const purpose, rootPath = "test_root_favicon", "/favicon.ico"

	if _, exists := BrandHeadAssetPaths[purpose]; exists {
		t.Fatalf("%q is already a real brand-head purpose — pick another probe key", purpose)
	}
	BrandHeadAssetPaths[purpose] = rootPath
	t.Cleanup(func() { delete(BrandHeadAssetPaths, purpose) })

	if got := DeployedWebPath(purpose, purpose); got != rootPath {
		t.Errorf("DeployedWebPath(%q,%q) = %q, want %q — the derivation is reconstructing the path "+
			"under %q instead of taking the map's value whole, so any brand-head asset not served "+
			"from the shared asset directory resolves to a path nothing serves",
			purpose, purpose, got, rootPath, DefaultAssetBasePath)
	}

	p := DeployedAssetPath(purpose, purpose)
	if p.FilePath != "favicon.ico" || p.Filename != "favicon.ico" {
		t.Errorf("DeployedAssetPath(%q,%q) = {FilePath:%q Filename:%q}, want both \"favicon.ico\" — "+
			"the deployer commits FilePath, so a wrong split writes the file to the wrong place",
			purpose, purpose, p.FilePath, p.Filename)
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
