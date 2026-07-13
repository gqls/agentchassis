// workflow_actions_size_test.go — the response truncation stub (Option A).
package actions

import "testing"

func TestTruncatedResponseStub(t *testing.T) {
	stub := truncatedResponseStub(1500000, 900000, map[string]interface{}{"a": 1, "b": 2})
	if stub["__truncated__"] != true || stub["original_size_bytes"] != 1500000 || stub["max_response_bytes"] != 900000 {
		t.Fatalf("stub fields wrong: %#v", stub)
	}
	keys, ok := stub["result_top_level_keys"].([]string)
	if !ok || len(keys) != 2 {
		t.Fatalf("stub must carry the result's top-level keys: %#v", stub["result_top_level_keys"])
	}
}
