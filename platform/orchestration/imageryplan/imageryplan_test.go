package imageryplan

import "testing"

func TestImageRoleForPath(t *testing.T) {
	cases := []struct {
		path     string
		wantRole string
		wantOK   bool
	}{
		{"background", "hero", true},
		{"background_image", "hero", true},
		{"image", "hero", true},
		{"hero_image", "hero", true},
		{"product_screenshot", "hero", true},
		{"product_image", "hero", true},
		{"banner", "hero", true},
		// Literal keys and roles themselves are NOT aliased — exact lookups
		// handle them before the alias is consulted.
		{"hero", "", false},
		{"logo", "", false},
		{"hero_product_detail", "", false},
		{"icon_precision", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		role, ok := ImageRoleForPath(c.path)
		if ok != c.wantOK || role != c.wantRole {
			t.Errorf("ImageRoleForPath(%q) = (%q, %v), want (%q, %v)",
				c.path, role, ok, c.wantRole, c.wantOK)
		}
	}
}

// ContentHeroKey (Phase I3, D13) is shared by three consumers AND mirrored
// inline in SQL as 'content_hero_' || replace(name, '-', '_') — pin the
// transform so a Go-side change can't silently diverge from the SQL.
func TestContentHeroKey(t *testing.T) {
	cases := map[string]string{
		"gripper-payload-calculator-guide": "content_hero_gripper_payload_calculator_guide",
		"news-post":                        "content_hero_news_post",
		"index":                            "content_hero_index",
	}
	for in, want := range cases {
		if got := ContentHeroKey(in); got != want {
			t.Errorf("ContentHeroKey(%q) = %q, want %q", in, got, want)
		}
	}
}
