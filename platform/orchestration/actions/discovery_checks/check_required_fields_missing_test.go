package discovery_checks

import (
	"reflect"
	"testing"
)

func TestMissingRequiredValueFields(t *testing.T) {
	// Modelled on gripper-detail's product-details_pre_037 schema: value
	// fields required from the LLM, chrome fields optional, one image field
	// and one query-sourced field that must be ignored.
	fields := map[string]interface{}{
		"product_name":  map[string]interface{}{"type": "text", "source": "llm", "required": true},
		"product_price": map[string]interface{}{"type": "text", "source": "llm", "required": true},
		"feature_1":     map[string]interface{}{"type": "text", "source": "llm", "required": "true"}, // string encoding
		"product_sku":   map[string]interface{}{"type": "text", "required": true},                    // no source → llm-ish, checked
		"sku_label":     map[string]interface{}{"type": "text", "source": "llm", "required": false},
		"main_image":    map[string]interface{}{"type": "image", "source": "site_assets.hero", "required": true},
		"products":      map[string]interface{}{"type": "array", "source": "query.affiliate_products", "required": true},
	}

	t.Run("chrome-only content_data misses every value field", func(t *testing.T) {
		content := map[string]interface{}{
			"sku_label":         "SKU:",
			"add_to_cart_label": "Add to Cart",
			"product_price":     "  ", // whitespace-only counts as missing
		}
		got := missingRequiredValueFields(fields, content)
		want := []string{"feature_1", "product_name", "product_price", "product_sku"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("filled content_data passes", func(t *testing.T) {
		content := map[string]interface{}{
			"product_name":  "PG-90 Parallel Gripper",
			"product_price": "£1,240",
			"feature_1":     "90mm stroke",
			"product_sku":   "PG-90",
		}
		if got := missingRequiredValueFields(fields, content); len(got) != 0 {
			t.Errorf("expected no missing fields, got %v", got)
		}
	})

	t.Run("zero and false are values, empty collections are not", func(t *testing.T) {
		fields := map[string]interface{}{
			"count":  map[string]interface{}{"type": "number", "source": "llm", "required": true},
			"active": map[string]interface{}{"type": "bool", "source": "llm", "required": true},
			"items":  map[string]interface{}{"type": "array", "source": "llm", "required": true},
		}
		content := map[string]interface{}{
			"count":  float64(0),
			"active": false,
			"items":  []interface{}{},
		}
		got := missingRequiredValueFields(fields, content)
		want := []string{"items"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}

// TestMissingRequiredValueFields_SiteAssetsSourcedFieldsAreChecked pins the
// 2026-08-11 widening (bugs_open/238).
//
// The old predicate skipped every non-llm source on the stated premise that
// "render-time sources … are not baked into content_data". For `site_assets.*`
// that is false and the platform says so: plan_sections resolves those into
// resolvedData, and RenderComponentAction's merge_with overlay persists
// resolvedData INTO content_data — PBP-014, the property that makes no-LLM
// re-rendering possible at all. So this check was blind to exactly the class
// that shipped five empty <img src=""> to a live homepage.
//
// The test also pins what did NOT widen, because an over-eager widening here is
// a fleet-wide flag-volume change nobody measured.
func TestMissingRequiredValueFields_SiteAssetsSourcedFieldsAreChecked(t *testing.T) {
	// The bugs_open/238 shape: type "url", not "image" — which is why the
	// sibling image check was silent on it too until the same commit.
	fields := map[string]interface{}{
		"card1_image_url": map[string]interface{}{"type": "url", "source": "site_assets.image", "required": true},
		"card1_link_url":  map[string]interface{}{"type": "url", "source": "site_specs.case_studies.card1_url", "required": true},
		"cta_link_url":    map[string]interface{}{"type": "url", "source": "site_specs.pages.contact_url", "required": true},
		"hero_image":      map[string]interface{}{"type": "image", "source": "site_assets.hero", "required": true},
		"card1_title":     map[string]interface{}{"type": "text", "source": "llm", "required": true},
		"products":        map[string]interface{}{"type": "array", "source": "query.affiliate_products", "required": true},
	}

	t.Run("a site_assets-sourced url absent from a deployed row is now reported", func(t *testing.T) {
		content := map[string]interface{}{"card1_title": "Cutting the Manual Work Out of Quote Requests"}
		got := missingRequiredValueFields(fields, content)
		want := []string{"card1_image_url"}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v.\n"+
				"card1_image_url must be reported: it IS baked into content_data via resolved_data (PBP-014), and its absence is what renders src=\"\".\n"+
				"card1_link_url / cta_link_url must NOT be: site_specs.* stays skipped in this slice (unmeasured volume, and those fields are {{if}}-gated).\n"+
				"hero_image must NOT be: image-typed fields stay owned by image_source_unsatisfiable.\n"+
				"products must NOT be: query.* stays skipped.", got, want)
		}
	})

	t.Run("a populated site_assets field passes", func(t *testing.T) {
		content := map[string]interface{}{
			"card1_title":     "Cutting the Manual Work Out of Quote Requests",
			"card1_image_url": "/assets/images/case-study-facilities.jpg",
		}
		if got := missingRequiredValueFields(fields, content); len(got) != 0 {
			t.Errorf("expected no missing fields once the value is present, got %v", got)
		}
	})

	t.Run("an empty string is missing, not present", func(t *testing.T) {
		// The distinction matters: `content_data ? 'card1_image_url'` is TRUE for
		// an empty string, and it renders src="" exactly as an absent key does.
		content := map[string]interface{}{
			"card1_title":     "Cutting the Manual Work Out of Quote Requests",
			"card1_image_url": "",
		}
		got := missingRequiredValueFields(fields, content)
		if len(got) != 1 || got[0] != "card1_image_url" {
			t.Errorf("an empty site_assets value must count as missing, got %v", got)
		}
	})
}
