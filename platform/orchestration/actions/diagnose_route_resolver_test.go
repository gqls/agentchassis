// diagnose_route_resolver_test.go — §7D resolver core (pure parts).
// The search function is injected, so no DB/HTTP is needed here; the wired
// chain (embedding client + vector/trigram search) is the SAME one
// lookup_code_symbols exercises in production.
package actions

import (
	"errors"
	"testing"

	"go.uber.org/zap"
)

func fakeSearch(rows []map[string]interface{}, err error) (func(string) ([]map[string]interface{}, error), *int) {
	calls := 0
	return func(string) ([]map[string]interface{}, error) {
		calls++
		return rows, err
	}, &calls
}

func TestResolveScopeEntries_ExactPassthroughNoSearch(t *testing.T) {
	files := map[string]bool{"platform/orchestration/result_spec.go": true}
	syms := map[string]bool{"platform/orchestration/result_spec.go:resolveResultSpec": true}
	search, calls := fakeSearch(nil, nil)
	in := []string{
		"platform/orchestration/result_spec.go:resolveResultSpec", // known symbol
		"platform/orchestration/result_spec.go",                   // known bare file
	}
	out := resolveScopeEntries(in, files, syms, search, 0.55, zap.NewNop())
	if len(out) != 2 || out[0] != in[0] || out[1] != in[1] {
		t.Fatalf("exact entries must pass through untouched, got %v", out)
	}
	if *calls != 0 {
		t.Fatalf("exact entries must not trigger a search, got %d calls", *calls)
	}
}

func TestResolveScopeEntries_FuzzySubstitutionWithFloorAndDedupe(t *testing.T) {
	files, syms := map[string]bool{}, map[string]bool{
		"a.go:Keep": true,
	}
	rows := []map[string]interface{}{
		{"path": "actions/plan_sections_action.go", "symbol": "(*sourceResolver).resolveSpecPath", "similarity": 0.81},
		{"path": "a.go", "symbol": "Keep", "similarity": 0.80},    // dedupes against the exact entry
		{"path": "x.go", "symbol": "TooWeak", "similarity": 0.31}, // below floor — dropped
	}
	search, calls := fakeSearch(rows, nil)
	in := []string{"a.go:Keep", "the code that resolves site_specs references at build time"}
	out := resolveScopeEntries(in, files, syms, search, 0.55, zap.NewNop())
	want := []string{"a.go:Keep", "actions/plan_sections_action.go:(*sourceResolver).resolveSpecPath"}
	if len(out) != len(want) || out[0] != want[0] || out[1] != want[1] {
		t.Fatalf("fuzzy substitution wrong: got %v want %v", out, want)
	}
	if *calls != 1 {
		t.Fatalf("expected exactly one search (one fuzzy entry), got %d", *calls)
	}
}

func TestResolveScopeEntries_UnresolvableStaysLabel(t *testing.T) {
	search, _ := fakeSearch([]map[string]interface{}{
		{"path": "x.go", "symbol": "Weak", "similarity": 0.20},
	}, nil)
	in := []string{"some vague description with no good match"}
	out := resolveScopeEntries(in, map[string]bool{}, map[string]bool{}, search, 0.55, zap.NewNop())
	if len(out) != 1 || out[0] != in[0] {
		t.Fatalf("all-below-floor must keep the prose label, got %v", out)
	}
}

func TestResolveScopeEntries_SearchErrorKeepsLabel(t *testing.T) {
	search, _ := fakeSearch(nil, errors.New("boom"))
	in := []string{"describe something"}
	out := resolveScopeEntries(in, map[string]bool{}, map[string]bool{}, search, 0.55, zap.NewNop())
	if len(out) != 1 || out[0] != in[0] {
		t.Fatalf("search error must keep the prose label (fail-open), got %v", out)
	}
}

func TestResolveScopeEntries_TrigramRowsWithoutSimilarityAccepted(t *testing.T) {
	search, _ := fakeSearch([]map[string]interface{}{
		{"path": "actions/save_page_sections_action.go", "symbol": "SavePageSectionsAction"},
	}, nil)
	out := resolveScopeEntries([]string{"save page sections"}, map[string]bool{}, map[string]bool{}, search, 0.55, zap.NewNop())
	if len(out) != 1 || out[0] != "actions/save_page_sections_action.go:SavePageSectionsAction" {
		t.Fatalf("similarity-less (trigram) rows must be accepted, got %v", out)
	}
}

func TestResolveScopeEntries_EmptyAndBlankEntries(t *testing.T) {
	search, calls := fakeSearch(nil, nil)
	out := resolveScopeEntries([]string{"", "   "}, map[string]bool{}, map[string]bool{}, search, 0.55, zap.NewNop())
	if len(out) != 0 || *calls != 0 {
		t.Fatalf("blank entries must be dropped without searching, got %v (%d calls)", out, *calls)
	}
}

func TestKnownScopeIdentities_ParsesFilesFunctionsTypes(t *testing.T) {
	raw := map[string]interface{}{
		"files": []interface{}{
			map[string]interface{}{
				"path": "platform/orchestration/result_spec.go",
				"functions": []interface{}{
					map[string]interface{}{"name": "resolveResultSpec"},
					map[string]interface{}{"name": "(*SagaCoordinator).fallbackDumpInto"},
				},
				"types": []interface{}{
					map[string]interface{}{"name": "ResultSpec"},
				},
			},
			map[string]interface{}{
				"path":      "cmd/auth-service/swagger_endpoints.go",
				"functions": nil, // symbol-less file — path still known
			},
		},
	}
	files, syms := knownScopeIdentities(raw)
	for _, wantFile := range []string{"platform/orchestration/result_spec.go", "cmd/auth-service/swagger_endpoints.go"} {
		if !files[wantFile] {
			t.Fatalf("missing file %s", wantFile)
		}
	}
	for _, wantSym := range []string{
		"platform/orchestration/result_spec.go:resolveResultSpec",
		"platform/orchestration/result_spec.go:(*SagaCoordinator).fallbackDumpInto",
		"platform/orchestration/result_spec.go:ResultSpec",
	} {
		if !syms[wantSym] {
			t.Fatalf("missing symbol %s", wantSym)
		}
	}
}
