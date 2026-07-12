package actions

import "testing"

// The awaited-reply shapes vary across the codebase; extractRunResults must
// find results through every fallback and recompute the verdict itself.
func TestExtractRunResultsFallbackPaths(t *testing.T) {
	results := []interface{}{
		map[string]interface{}{"check_id": "boots", "pass": true, "detail": "ok"},
		map[string]interface{}{"check_id": "rows", "pass": false, "detail": "no match"},
	}
	skipped := []interface{}{
		map[string]interface{}{"check_id": "mobile-fit", "detail": "P0"},
	}

	shapes := map[string]map[string]interface{}{
		"response.data": {"browser_run": map[string]interface{}{
			"response": map[string]interface{}{
				"data": map[string]interface{}{"results": results, "skipped": skipped}}}},
		"response": {"browser_run": map[string]interface{}{
			"response": map[string]interface{}{"results": results, "skipped": skipped}}},
		"flattened": {"browser_run": map[string]interface{}{
			"results": results, "skipped": skipped}},
	}

	for name, collected := range shapes {
		v := extractRunResults(collected, "browser_run")
		if len(v.Results) != 2 {
			t.Errorf("%s: expected 2 results, got %d", name, len(v.Results))
			continue
		}
		if len(v.Passed) != 1 || v.Passed[0] != "boots" {
			t.Errorf("%s: passed wrong: %v", name, v.Passed)
		}
		if len(v.Failed) != 1 || v.Failed[0] != "rows" {
			t.Errorf("%s: failed wrong: %v", name, v.Failed)
		}
		if len(v.Details) != 1 || v.Details[0] != "rows: no match" {
			t.Errorf("%s: details wrong: %v", name, v.Details)
		}
		if len(v.SkipList) != 1 || v.SkipList[0] != "mobile-fit" {
			t.Errorf("%s: skip list wrong: %v", name, v.SkipList)
		}
	}
}

func TestExtractRunResultsSkippedRun(t *testing.T) {
	// request_browser_run's no-criteria no-op: {skipped: true, reason: ...}
	collected := map[string]interface{}{
		"browser_run": map[string]interface{}{"skipped": true, "reason": "needs_criteria"},
	}
	v := extractRunResults(collected, "browser_run")
	if !v.Skipped {
		t.Fatal("a skipped run must be recognised, never judged")
	}
	if len(v.Results) != 0 {
		t.Fatalf("skipped run must carry no results, got %d", len(v.Results))
	}
}

func TestExtractRunResultsEmpty(t *testing.T) {
	v := extractRunResults(map[string]interface{}{}, "browser_run")
	if v.Skipped || len(v.Results) != 0 {
		t.Fatalf("missing field must yield empty verdict, got %+v", v)
	}
}
