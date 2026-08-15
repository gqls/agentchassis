// FILE: platform/orchestration/input_contracts/input_mapping_strict_test.go
//
// RFC_029 §9 D3 (owner-delegated ruling, 2026-08-15) — the `!` STRICT marker on
// the input_mapping surface. An UNMARKED field here already hard-fails when its
// source path does not resolve, so on this surface the marker's value is the
// TRANSITION: flipping "field?" to "field!" converts a silent skip — after which
// the CHILD's own resolver may fall back to the whole-tree search — into a loud
// failure at the caller, with the happy path unchanged. First live adopter:
// migration 417 (HOLD) flips asset_id? to asset_id! on image-build-handler's
// call_asset_deployer.
package input_contracts

import (
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestResolveInputMappingStrictMarker(t *testing.T) {
	logger := zap.NewNop()
	collected := map[string]interface{}{
		"asset_stored": map[string]interface{}{"asset_id": "a-1"},
	}

	t.Run("strict field resolves and is delivered under its bare name", func(t *testing.T) {
		result, err := ResolveInputMapping(collected,
			InputMapping{"asset_id!": "asset_stored.asset_id"}, logger)
		if err != nil {
			t.Fatalf("ResolveInputMapping: %v", err)
		}
		if result["asset_id"] != "a-1" {
			t.Fatalf("asset_id = %v, want \"a-1\" — the marker must be stripped from the "+
				"delivered key, or the child sees a field named \"asset_id!\" and its own "+
				"resolver falls back to the whole-tree search (the exact regression the "+
				"marker exists to prevent)", result["asset_id"])
		}
		if _, leaked := result["asset_id!"]; leaked {
			t.Error("the suffixed key leaked into the resolved map")
		}
	})

	t.Run("strict field that cannot resolve fails loudly, naming the marker", func(t *testing.T) {
		_, err := ResolveInputMapping(collected,
			InputMapping{"asset_id!": "asset_stored.missing_key"}, logger)
		if err == nil {
			t.Fatal("strict field with an unresolvable path returned nil error — the loud " +
				"failure is the marker's entire contract on this surface")
		}
		if !strings.Contains(err.Error(), "STRICT") || !strings.Contains(err.Error(), "asset_id") {
			t.Errorf("error %q must name the marker and the field", err.Error())
		}
	})

	t.Run("optional field is still silently skipped — the pre-flip behaviour", func(t *testing.T) {
		result, err := ResolveInputMapping(collected,
			InputMapping{"asset_id?": "asset_stored.missing_key"}, logger)
		if err != nil {
			t.Fatalf("ResolveInputMapping: %v", err)
		}
		if _, present := result["asset_id"]; present {
			t.Error("optional unresolved field appeared in the result")
		}
	})
}

func TestResolveInputMappingWithItemStrictMarker(t *testing.T) {
	logger := zap.NewNop()
	collected := map[string]interface{}{
		"asset_stored": map[string]interface{}{"asset_id": "a-1"},
	}

	result, err := ResolveInputMappingWithItem(collected,
		InputMapping{"asset_id!": "asset_stored.asset_id", "item!": "$item"},
		map[string]interface{}{"id": "wi-9"}, logger)
	if err != nil {
		t.Fatalf("ResolveInputMappingWithItem: %v", err)
	}
	if result["asset_id"] != "a-1" {
		t.Errorf("asset_id = %v, want \"a-1\"", result["asset_id"])
	}
	if item, ok := result["item"].(map[string]interface{}); !ok || item["id"] != "wi-9" {
		t.Errorf("item = %#v, want the $item passthrough under the bare name", result["item"])
	}

	if _, err := ResolveInputMappingWithItem(collected,
		InputMapping{"asset_id!": "asset_stored.missing_key"}, nil, logger); err == nil {
		t.Fatal("strict field with an unresolvable path returned nil error (with_item)")
	}
}
