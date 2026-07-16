// FILE: platform/orchestration/actions/discovery_checks/content_image_helpers_test.go

package discovery_checks

import (
	"encoding/json"
	"testing"
)

func TestContentImageItemKey(t *testing.T) {
	if got := contentImageItemKey("gripper-payload-calculator-guide"); got != "content_image:gripper-payload-calculator-guide" {
		t.Fatalf("item key = %q", got)
	}
}

// The spec is the contract with asset-deployer's content_card routing
// (spec.mode) and derive_card_asset's inputs (entity_type/entity_id/
// page_name) — pin every field the other side reads.
func TestContentImageSpecJSON(t *testing.T) {
	s, err := contentImageSpecJSON("content_image_missing", "3fbb0f5e-0000-0000-0000-000000000000", "my-guide")
	if err != nil {
		t.Fatal(err)
	}
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(s), &spec); err != nil {
		t.Fatalf("spec is not valid JSON: %v", err)
	}
	want := map[string]string{
		"mode":        "content_card",
		"check":       "content_image_missing",
		"entity_type": "page",
		"entity_id":   "3fbb0f5e-0000-0000-0000-000000000000",
		"page_name":   "my-guide",
		"purpose":     "card",
	}
	for k, v := range want {
		if spec[k] != v {
			t.Errorf("spec[%q] = %v, want %q", k, spec[k], v)
		}
	}
}

func TestContentImageMissingCheckRegistered(t *testing.T) {
	c := &ContentImageMissingCheck{}
	if c.Name() != "content_image_missing" {
		t.Fatalf("check name = %q", c.Name())
	}
}
