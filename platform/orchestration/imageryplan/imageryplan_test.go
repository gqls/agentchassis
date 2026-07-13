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
