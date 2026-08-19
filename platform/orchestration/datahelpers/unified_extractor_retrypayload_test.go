package datahelpers

// bugs_open/306 candidate 3 — the whole-tree search does not descend into
// retry_payload subtrees.
//
// A retry_payload is the verbatim produced message an awaited-request action
// records for timeout replay (bugs_open/129, types.RetryPayloadKey). It carries
// a frozen copy of the sender's inputs, and those echoes were the losing
// candidates in every genuinely-different-page conflict measured on
// page-build-handler (13/139, 2026-08-18) — and the WINNING candidates in the
// build-dispatch-loop current_page class (4,370 rows to 2026-08-19, a slot the
// lane's decision-3 sweep found no reader of). Its only real consumer reads it
// by direct key from the action result before that result is merged into
// CollectedData (coordinator.go extractRetryPayload), then from
// awaited_requests.request_payload at retry time — never through this search.
//
// Mutation proof (run before committing): remove the types.RetryPayloadKey case
// from isInfrastructureKey → TestSearchSkipsRetryPayloadSubtree FAILS (the
// conflict WARN fires again) and TestSearchSkipsNestedRetryPayloadEcho FAILS
// (the stale echo wins on depth). The rank tie-break tests in
// unified_extractor_tiebreak_test.go use non-retry_payload sibling keys and are
// untouched by the skip — the tie-break stays load-bearing for every conflict
// the skip does not remove (e.g. the page-content-writer shape class).

import (
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/types"
)

// The production shape from the 306 measurement, reduced: the dispatched page
// at input_data.current_page, and a stale page frozen inside the retry echo a
// call step left under its output field. Before the skip this was a genuine
// two-page conflict resolved only by the rank tie-break; with the skip the
// candidate set is unique and no conflict WARN fires at all.
func retryEchoFixture() map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"current_page": map[string]interface{}{"name": "disclaimer"},
		},
		"page_content": map[string]interface{}{
			types.RetryPayloadKey: map[string]interface{}{
				"topic": "core.requests.content-writer",
				"key":   "site-1",
				"message": map[string]interface{}{
					"payload": map[string]interface{}{
						"current_page": map[string]interface{}{"name": "contact-index"},
					},
				},
			},
			"response": map[string]interface{}{"status": "complete"},
		},
	}
}

func TestSearchSkipsRetryPayloadSubtree(t *testing.T) {
	value, logs := observedSearch(t, retryEchoFixture(), "current_page")
	page, _ := value.(map[string]interface{})
	if page["name"] != "disclaimer" {
		t.Fatalf("current_page = %v, want the dispatched page \"disclaimer\" — the retry echo must not supply candidates", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 0 {
		t.Fatalf("conflict WARN fired %d times, want 0 — with the echo skipped the candidate set is unique, not a conflict the tie-break has to resolve", n)
	}
}

func TestSearchSkipsNestedRetryPayloadEcho(t *testing.T) {
	// The echo can hold a SHALLOWER copy than the live value (the
	// build-dispatch-loop class, where the echoed copy WON on depth before the
	// skip). Skipping must remove it wherever it sits, not only at depth 1.
	data := map[string]interface{}{
		"handler": map[string]interface{}{
			types.RetryPayloadKey: map[string]interface{}{
				"message": map[string]interface{}{"current_page": map[string]interface{}{"name": "stale-echo"}},
			},
		},
		"deeper": map[string]interface{}{
			"wrapper": map[string]interface{}{
				"current_page": map[string]interface{}{"name": "live-page"},
			},
		},
	}
	value, logs := observedSearch(t, data, "current_page")
	page, _ := value.(map[string]interface{})
	if page["name"] != "live-page" {
		t.Fatalf("current_page = %v, want \"live-page\" — a shallower candidate inside the retry echo must not outrank the live one", value)
	}
	if n := logs.FilterMessage(conflictWarnMsg).Len(); n != 0 {
		t.Fatalf("conflict WARN fired %d times, want 0", n)
	}
}

func TestRetryPayloadItselfStillResolvesByDirectMatch(t *testing.T) {
	// Same semantics as agent_config: the skip guards RECURSION only. A caller
	// that explicitly asks for the retry_payload field still gets it.
	data := map[string]interface{}{
		"page_content": map[string]interface{}{
			types.RetryPayloadKey: map[string]interface{}{"topic": "core.requests.x"},
		},
	}
	value, _ := observedSearch(t, data, types.RetryPayloadKey)
	m, _ := value.(map[string]interface{})
	if m["topic"] != "core.requests.x" {
		t.Fatalf("retry_payload = %v, want the direct match — the skip must not hide the key from a caller that names it", value)
	}
}
