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
