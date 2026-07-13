package storage

import "testing"

// Verified against the files actually committed to the robot-hands.com repo
// (assets/images/hero-home.jpg, hero-gripper-catalog.jpg, icon-cycle-time.jpg,
// logo.png), so the resolver references what the deployer wrote.
func TestDeployedWebPath(t *testing.T) {
	cases := []struct {
		assetKey string
		purpose  string
		want     string
	}{
		{"hero_home", "hero", "/assets/images/hero-home.jpg"},
		{"hero_gripper_catalog", "hero", "/assets/images/hero-gripper-catalog.jpg"},
		{"hero_product_detail", "hero", "/assets/images/hero-product-detail.jpg"},
		{"icon_cycle_time", "icon", "/assets/images/icon-cycle-time.jpg"},
		// asset_key == purpose → the canonical purpose file, purpose extension.
		{"logo", "logo", "/assets/images/logo.png"},
		{"hero", "hero", "/assets/images/hero.jpg"},
		// Unknown purpose falls back to the default image config (jpg).
		{"some_thing", "mystery", "/assets/images/some-thing.jpg"},
	}
	for _, c := range cases {
		if got := DeployedWebPath(c.assetKey, c.purpose); got != c.want {
			t.Errorf("DeployedWebPath(%q, %q) = %q, want %q", c.assetKey, c.purpose, got, c.want)
		}
	}
}

func TestAssetKeyFilename(t *testing.T) {
	cases := []struct {
		assetKey string
		ext      string
		want     string
	}{
		{"hero_home", ".jpg", "hero-home.jpg"},
		{"icon_cycle_time", "jpg", "icon-cycle-time.jpg"}, // leading dot added
		{"logo", ".png", "logo.png"},
		{"hero_about", "", "hero-about.jpg"}, // empty ext defaults to .jpg
	}
	for _, c := range cases {
		if got := AssetKeyFilename(c.assetKey, c.ext); got != c.want {
			t.Errorf("AssetKeyFilename(%q, %q) = %q, want %q", c.assetKey, c.ext, got, c.want)
		}
	}
}
