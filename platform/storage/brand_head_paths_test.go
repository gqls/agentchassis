// FILE: platform/storage/brand_head_paths_test.go
//
// Pins the reason BrandHeadAssetPaths is DECLARED rather than derived
// (bugs_open/142).

package storage

import "testing"

// TestDeployedWebPathCannotExpressBrandHeadPaths answers the reuse question
// permanently, with a measurement rather than an assertion.
//
// DeployedWebPath is documented as mirroring the deployer's path derivation and
// AssetKeyFilename calls itself "the single source of truth for the variant
// filename convention". Both are true for every purpose EXCEPT the brand-head
// pair, because the underscore→hyphen swap in AssetKeyFilename fires only when
// assetKey differs from purpose — and for favicon/og_card it does not. So
// `og_card` comes out as `og_card.png` while the file the deriver actually
// commits is `og-card.png`.
//
// If someone later teaches DeployedWebPath about this case, this test fails and
// tells them to collapse the map into it. Until then it documents why the
// duplication is real and not laziness.
func TestDeployedWebPathCannotExpressBrandHeadPaths(t *testing.T) {
	const wantOG = "/assets/images/og-card.png"

	if got := BrandHeadAssetPaths["og_card"]; got != wantOG {
		t.Fatalf("BrandHeadAssetPaths[og_card] = %q, want %q — this is the path "+
			"derive_brand_head_assets commits and injectBrandHeadTags emits", got, wantOG)
	}

	// Every way of asking the generic helper for the og card gets it wrong.
	for _, c := range []struct{ assetKey, purpose string }{
		{"og_card", "og_card"},
		{"", "og_card"},
		{"og_card", ""},
	} {
		if got := DeployedWebPath(c.assetKey, c.purpose); got == wantOG {
			t.Errorf("DeployedWebPath(%q, %q) now returns %q — the generic helper can express the "+
				"brand-head path after all.\nCollapse BrandHeadAssetPaths into it and delete this test, "+
				"or the platform has two answers to one question.", c.assetKey, c.purpose, got)
		}
	}

	// favicon is the coincidence, and saying so stops someone concluding from it
	// that the generic helper is fine for both.
	if got := DeployedWebPath("favicon", "favicon"); got != BrandHeadAssetPaths["favicon"] {
		t.Errorf("DeployedWebPath(favicon, favicon) = %q but the map says %q — favicon used to agree "+
			"by coincidence (no underscore to disagree about); if that changed, re-read the map's comment",
			got, BrandHeadAssetPaths["favicon"])
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
