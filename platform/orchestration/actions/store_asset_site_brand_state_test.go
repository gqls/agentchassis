// FILE: platform/orchestration/actions/store_asset_site_brand_state_test.go
//
// bugs_open/114. store_asset wrote sites.content_data.<purpose>_url on EVERY
// store, with a value derived from the PURPOSE alone. Two consequences, both
// measured on the live fleet 2026-08-22 before the fix:
//
//   - every page-scoped generation overwrote the site-wide default, so a hero
//     generated for one page re-pointed every page that falls back to it. 18
//     sites carried an identical hero_url; one site's hand-repair had been
//     silently undone by the next generation.
//   - the value named a file the deployer never commits. The deployer derives
//     from the ASSET KEY (storage.DeployedAssetPath), so an asset stored as
//     content_hero_tool_repayment deploys to content-hero-tool-repayment.jpg
//     while content_data was told content_hero.jpg — 404 on all six sites
//     carrying it, and a filename the deployer cannot produce for any input.
//
// Both arms of the gate AND the derivation are proved here rather than asserted,
// because a guard nobody has watched fail is not a guard: delete either
// condition in storeAssetContentDataUpdate, or revert its URL to the old
// purpose-derived form, and a case below fails.
//
// The gate's shape came from MEASURING the live callers rather than assuming
// them. A first cut made asset_key == purpose non-negotiable, which would have
// silenced the brand-update branch's legitimate purpose=hero/asset_key=hero_home
// stores; the enumeration of all nine live store_asset steps is what caught it.

package actions

import (
	"testing"

	"github.com/gqls/agentchassis/platform/storage"
)

// TestStoreAssetContentDataUpdate pins both conditions, the default, and the
// URL the action records. Each case isolates one arm — a single combined case
// would still pass with one arm deleted, which is exactly the mutation this
// test exists to catch.
//
// The URL is asserted HERE, through the same helper the action calls, and not
// against storage.DeployedWebPath in a test of its own. The first cut did the
// latter and a mutation reverting the action to the old purpose-derived path
// passed unnoticed: the test was exercising the storage package while claiming
// to cover store_asset.
func TestStoreAssetContentDataUpdate(t *testing.T) {
	cases := []struct {
		name     string
		assetKey string
		purpose  string
		config   map[string]interface{}
		want     bool
		wantURL  string
		why      string
	}{
		{
			name:     "canonical asset, no config",
			assetKey: "hero",
			purpose:  "hero",
			config:   map[string]interface{}{},
			want:     true,
			wantURL:  "/assets/images/hero.jpg",
			why:      "the site-wide hero IS site-wide brand state; absent config must keep today's behaviour for pageflow-builder's brand stores",
		},
		{
			name:     "page-scoped hero variant",
			assetKey: "hero_about",
			purpose:  "hero",
			config:   map[string]interface{}{},
			want:     false,
			wantURL:  "/assets/images/hero-about.jpg",
			why:      "one page's hero is not the site's hero; this is the arm that stopped the 18-site spread",
		},
		{
			name:     "content hero for one article",
			assetKey: "content_hero_tool_repayment",
			purpose:  "content_hero",
			config:   map[string]interface{}{},
			want:     false,
			wantURL:  "/assets/images/content-hero-tool-repayment.jpg",
			why:      "content_hero is per-page by construction, so a site-wide content_hero_url can never be right",
		},
		{
			name:     "canonical asset, caller opts out",
			assetKey: "hero",
			purpose:  "hero",
			config:   map[string]interface{}{"update_site_brand_assets": false},
			want:     false,
			wantURL:  "/assets/images/hero.jpg",
			why:      "the key has been in image-build-handler's config since it was written; before this fix no Go code read it",
		},
		{
			name:     "canonical asset, caller opts in explicitly",
			assetKey: "logo",
			purpose:  "logo",
			config:   map[string]interface{}{"update_site_brand_assets": true},
			want:     true,
			wantURL:  "/assets/images/logo.png",
			why:      "an explicit true must not be read as an opt-out",
		},
		{
			name:     "brand update names a page-scoped key deliberately",
			assetKey: "hero_home",
			purpose:  "hero",
			config:   map[string]interface{}{"update_site_brand_assets": true},
			want:     true,
			wantURL:  "/assets/images/hero-home.jpg",
			why:      "the live brand-update branch files purpose=hero asset_key=hero_home (10 items measured 2026-08-22); an explicit true is an assertion that this asset IS the site hero, and the URL it records must be the file that exists",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotURL, got := storeAssetContentDataUpdate(tc.assetKey, tc.purpose, tc.config)
			if got != tc.want {
				t.Fatalf("storeAssetContentDataUpdate(%q, %q, %v) writeSiteWide = %v, want %v — %s",
					tc.assetKey, tc.purpose, tc.config, got, tc.want, tc.why)
			}
			if gotURL != tc.wantURL {
				t.Fatalf("storeAssetContentDataUpdate(%q, %q, %v) url = %q, want %q — the recorded pointer must name the file the deployer commits",
					tc.assetKey, tc.purpose, tc.config, gotURL, tc.wantURL)
			}
		})
	}
}

// TestDeployedWebPathShapes pins the storage-package derivation for the shapes
// that occur on the fleet. It does NOT cover store_asset — that is
// TestStoreAssetContentDataUpdate's job, and conflating the two is the mistake
// this file's first cut made. Kept because it names the expected filenames in
// one place, including the og_card case BuildAssetPaths got wrong even for a
// canonical asset (og_card.png vs the committed og-card.png).
func TestDeployedWebPathShapes(t *testing.T) {
	cases := []struct {
		assetKey string
		purpose  string
		want     string
	}{
		{"hero", "hero", "/assets/images/hero.jpg"},
		{"hero_about", "hero", "/assets/images/hero-about.jpg"},
		{"content_hero_tool_repayment", "content_hero", "/assets/images/content-hero-tool-repayment.jpg"},
		{"logo", "logo", "/assets/images/logo.png"},
		// og_card is the case BuildAssetPaths got wrong even for a canonical
		// asset: it would have said og_card.png while the deployer commits
		// og-card.png, because brand-head purposes declare their published path
		// rather than deriving it. Switching to DeployedWebPath fixes that too.
		{"og_card", "og_card", "/assets/images/og-card.png"},
	}

	for _, tc := range cases {
		t.Run(tc.assetKey, func(t *testing.T) {
			got := storage.DeployedWebPath(tc.assetKey, tc.purpose)
			if got != tc.want {
				t.Fatalf("DeployedWebPath(%q, %q) = %q, want %q", tc.assetKey, tc.purpose, got, tc.want)
			}
		})
	}
}

// TestBuildAssetPathsDivergesFromDeployedPath is the negative control. It is not
// asserting that BuildAssetPaths is broken — it is a correct function for its
// own question — but that the two answers genuinely DIFFER for the shapes this
// fix is about. Without it, the URL assertions above would pass just as happily
// against the old derivation for the canonical cases and prove nothing.
func TestBuildAssetPathsDivergesFromDeployedPath(t *testing.T) {
	diverging := []struct {
		assetKey string
		purpose  string
	}{
		{"hero_about", "hero"},
		{"content_hero_tool_repayment", "content_hero"},
		{"og_card", "og_card"},
	}

	for _, tc := range diverging {
		_, _, _, ext := storage.GetImageConfig(tc.purpose)
		old := storage.BuildAssetPaths(tc.purpose, ext).RelativeURL
		now := storage.DeployedWebPath(tc.assetKey, tc.purpose)
		if old == now {
			t.Fatalf("purpose-derived and key-derived paths agree for asset_key=%q purpose=%q (both %q); "+
				"this case can no longer demonstrate the defect and the test above proves nothing for it",
				tc.assetKey, tc.purpose, old)
		}
	}
}
