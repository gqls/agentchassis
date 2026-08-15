// FILE: platform/orchestration/park_carries_collected_data_test.go
//
// Tests for carryCollectedDataOntoFreshState — the park path's additive carry
// (bugs_open/236, RFC_012 (a), owner ruling 2026-08-15).
//
// The shapes here are the live ones, not invented: the logo case reproduces
// orchestration 3e46be5b-8788-447b-9643-e32ae33f601b's persisted record from
// bugs_open/236 §1, and the action-side keys are the three
// deploy_image_asset_action.go assigns onto its own result before dispatching.
package orchestration

import (
	"testing"

	"go.uber.org/zap"
)

// logoDeployedInMemory is what the action returns and storeActionResult writes
// into CollectedData in memory, moments before the park.
func logoDeployedInMemory() map[string]interface{} {
	return map[string]interface{}{
		"deployed":       true,
		"await_response": true,
		"request_id":     "req-logo-1",
		"purpose":        "logo",
		"domain":         "cookly.uk",
		"image_url":      "/assets/images/logo.png",
		"output_path":    "assets/images/logo.png",
		"size_bytes":     40213,
	}
}

func parkedState(stepName string) (*OrchestrationState, *OrchestrationState) {
	inMemory := &OrchestrationState{
		OrchestrationID: "3e46be5b-8788-447b-9643-e32ae33f601b",
		CollectedData: map[string]interface{}{
			stepName:     logoDeployedInMemory(),
			"input_data": map[string]interface{}{"domain": "cookly.uk"},
		},
		AwaitedRequests: map[string]*AwaitedRequest{
			"req-logo-1": {RequestID: "req-logo-1", StepName: stepName},
		},
	}
	// The DB copy the park reloads: it has none of the step's in-flight work.
	fresh := &OrchestrationState{
		OrchestrationID: inMemory.OrchestrationID,
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"domain": "cookly.uk"},
		},
	}
	return inMemory, fresh
}

// The defect itself: before the fix the parked row held nothing the action
// computed, so the reader's image_url was gone before the reply ever arrived.
func TestParkCarriesTheActionsOwnKeys(t *testing.T) {
	inMemory, fresh := parkedState("logo_deployed")

	carried := carryCollectedDataOntoFreshState(fresh, inMemory, zap.NewNop())

	container, ok := fresh.CollectedData["logo_deployed"].(map[string]interface{})
	if !ok {
		t.Fatalf("logo_deployed absent from the parked state; carried=%v", carried)
	}
	if got := container["image_url"]; got != "/assets/images/logo.png" {
		t.Errorf("image_url = %v, want /assets/images/logo.png — this is the key the readers need", got)
	}
	if got := container["output_path"]; got != "assets/images/logo.png" {
		t.Errorf("output_path = %v, want it carried too", got)
	}
	if got := container["size_bytes"]; got != 40213 {
		t.Errorf("size_bytes = %v, want it carried too", got)
	}
	if len(carried) != 1 || carried[0] != "logo_deployed" {
		t.Errorf("carried = %v, want exactly [logo_deployed] (input_data was already on the fresh copy)", carried)
	}
}

// The safety direction. The fresh copy is the DB authority: if a key is already
// there, the in-memory value must not replace it. This is what makes the change
// additive rather than a second overwrite bug.
func TestParkNeverOverwritesAKeyTheFreshStateAlreadyHas(t *testing.T) {
	inMemory, fresh := parkedState("logo_deployed")
	inMemory.CollectedData["shared_key"] = "stale in-memory value"
	fresh.CollectedData["shared_key"] = "value written by someone else"

	carried := carryCollectedDataOntoFreshState(fresh, inMemory, zap.NewNop())

	if got := fresh.CollectedData["shared_key"]; got != "value written by someone else" {
		t.Errorf("shared_key = %v, want the fresh state's value kept", got)
	}
	for _, k := range carried {
		if k == "shared_key" {
			t.Error("shared_key reported as carried; it was already present and must be skipped")
		}
	}
}

// A reply can land between the fresh load and the save. Its record must survive
// untouched — carrying the action's pre-dispatch map over it would revert the
// arrival and strand the orchestration.
func TestParkDoesNotClobberAReplyThatAlreadyArrived(t *testing.T) {
	inMemory, fresh := parkedState("logo_deployed")
	arrived := map[string]interface{}{
		"response":        map[string]interface{}{"data": map[string]interface{}{"success": true}},
		"response_status": "complete",
	}
	fresh.CollectedData["logo_deployed"] = arrived

	carryCollectedDataOntoFreshState(fresh, inMemory, zap.NewNop())

	container, ok := fresh.CollectedData["logo_deployed"].(map[string]interface{})
	if !ok {
		t.Fatal("logo_deployed is no longer a map")
	}
	if _, stillThere := container["response"]; !stillThere {
		t.Error("the arrived reply was overwritten by the park's carry")
	}
	if container["response_status"] != "complete" {
		t.Errorf("response_status = %v, want complete", container["response_status"])
	}
}

// A carried key spelled "response" under an awaited step would be read as an
// arrived reply by the arrival check. It must not survive the carry.
func TestParkStripsAForgedResponseMarkerUnderAnAwaitedStep(t *testing.T) {
	inMemory, fresh := parkedState("logo_deployed")
	result := logoDeployedInMemory()
	result[awaitedResponseMarker] = map[string]interface{}{"data": "not a real reply"}
	inMemory.CollectedData["logo_deployed"] = result

	carryCollectedDataOntoFreshState(fresh, inMemory, zap.NewNop())

	container := fresh.CollectedData["logo_deployed"].(map[string]interface{})
	if _, present := container[awaitedResponseMarker]; present {
		t.Error("a 'response' key was carried under an awaited step - it forges an arrived reply")
	}
	if container["image_url"] != "/assets/images/logo.png" {
		t.Error("stripping the marker must not cost the action's real keys")
	}
	// The live in-memory map must not be mutated by the strip.
	if _, gone := inMemory.CollectedData["logo_deployed"].(map[string]interface{})[awaitedResponseMarker]; !gone {
		t.Error("the strip mutated the caller's in-memory map instead of copying")
	}
}

// The guard is scoped to awaited steps. An ordinary key that happens to hold a
// "response" sub-key is data, not a protocol signal, and must cross intact.
func TestParkLeavesAResponseKeyAloneOnANonAwaitedKey(t *testing.T) {
	inMemory, fresh := parkedState("logo_deployed")
	inMemory.CollectedData["earlier_call"] = map[string]interface{}{
		"response": map[string]interface{}{"data": "a genuine earlier reply"},
	}

	carryCollectedDataOntoFreshState(fresh, inMemory, zap.NewNop())

	earlier, ok := fresh.CollectedData["earlier_call"].(map[string]interface{})
	if !ok {
		t.Fatal("earlier_call was not carried")
	}
	if _, present := earlier["response"]; !present {
		t.Error("an earlier step's genuine response was stripped; the guard must apply only to awaited steps")
	}
}

func TestParkCarryHandlesEmptyAndNilStates(t *testing.T) {
	if got := carryCollectedDataOntoFreshState(nil, nil, zap.NewNop()); got != nil {
		t.Errorf("nil states should carry nothing, got %v", got)
	}

	// A fresh row with no collected_data at all must be initialised, not panic.
	inMemory, fresh := parkedState("logo_deployed")
	fresh.CollectedData = nil

	carried := carryCollectedDataOntoFreshState(fresh, inMemory, zap.NewNop())

	if len(carried) != 2 {
		t.Errorf("carried = %v, want both keys onto a nil map", carried)
	}
	if _, ok := fresh.CollectedData["logo_deployed"]; !ok {
		t.Error("logo_deployed missing after carrying onto a nil CollectedData")
	}
}
