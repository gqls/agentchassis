package actions

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gqls/agentchassis/internal/analysis"
)

// The agent_error_log line format is the contract: `agent/step (action)`.
// Benchmark run 4d43d002: `page-build-handler/complete_error (complete_workflow)`
// was the causal step's name, present in every bundle, unreachable as evidence.
func TestWorkflowRefsFromRuntime(t *testing.T) {
	runtime := `### agent_error_log (most recent)
- [2026-07-06T12:51:15Z] page-build-handler/complete_error (complete_workflow) fatal: failed to send response
- [2026-07-06T16:39:51Z] image-build-handler/call_asset_deployer (call_agent) fatal: timed out
- [2026-07-06T12:51:15Z] page-build-handler/complete_error (complete_workflow) fatal: repeated line
- [2026-07-06T10:51:54Z] generic/persist_roadmap (write_site_spec) error: missing fields`

	refs, _ := workflowRefsFromRuntime(runtime, 4)
	want := [][2]string{
		{"page-build-handler", "complete_error"},
		{"image-build-handler", "call_asset_deployer"},
		{"generic", "persist_roadmap"},
	}
	if len(refs) != len(want) {
		t.Fatalf("want %d distinct refs, got %d: %v", len(want), len(refs), refs)
	}
	for i := range want {
		if refs[i] != want[i] {
			t.Fatalf("ref %d: want %v, got %v", i, want[i], refs[i])
		}
	}

	// The cap is honoured.
	if got, _ := workflowRefsFromRuntime(runtime, 2); len(got) != 2 {
		t.Fatalf("cap 2 not honoured: %v", got)
	}
	// No refs in plain prose.
	if got, _ := workflowRefsFromRuntime("no error lines here", 4); got != nil {
		t.Fatalf("expected nil on no matches, got %v", got)
	}
}

func TestSiblingSignatures(t *testing.T) {
	out := analysis.Output{Files: []analysis.FileInfo{
		{
			Path: "platform/orchestration/actions/populate_nav_tables_action.go",
			Functions: []analysis.FuncDef{
				{Name: "isLegalPage", Signature: "func isLegalPage(p pageNavInfo) bool"},
				{Name: "loadPagesForNav", Signature: "func loadPagesForNav(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]pageNavInfo, error)"},
			},
		},
		{
			Path:      "unrelated.go",
			Functions: []analysis.FuncDef{{Name: "X", Signature: "func X()"}},
		},
	}}

	// The benchmark gap: isLegalPage in scope, loadPagesForNav its unseen sibling.
	got := siblingSignatures(out, []string{"platform/orchestration/actions/populate_nav_tables_action.go:isLegalPage"}, 6000)
	if !strings.Contains(got, "loadPagesForNav") {
		t.Fatalf("sibling loadPagesForNav must be listed, got:\n%s", got)
	}
	if strings.Contains(got, "isLegalPage") {
		t.Fatalf("the in-scope symbol itself must not be listed as a sibling:\n%s", got)
	}
	if strings.Contains(got, "unrelated.go") {
		t.Fatalf("files with no in-scope symbol must not appear:\n%s", got)
	}

	// A whole-file scope entry has no siblings to add.
	if got := siblingSignatures(out, []string{"platform/orchestration/actions/populate_nav_tables_action.go"}, 6000); got != "" {
		t.Fatalf("whole-file scope should yield no sibling section, got:\n%s", got)
	}

	// The cap degrades to a truncation marker, not an oversized section.
	if got := siblingSignatures(out, []string{"platform/orchestration/actions/populate_nav_tables_action.go:isLegalPage"}, 10); !strings.Contains(got, "omitted") {
		t.Fatalf("cap should leave a truncation marker, got:\n%s", got)
	}
}

// Fair-share regression (benchmark forensics 2026-07-10): an alphabetically
// early file with many functions must not starve a later scoped file — that is
// exactly how loadPagesForNav stayed invisible across four benchmark runs.
func TestSiblingSignatures_FairShare(t *testing.T) {
	big := analysis.FileInfo{Path: "a_giant_file.go"}
	for i := 0; i < 80; i++ {
		big.Functions = append(big.Functions, analysis.FuncDef{
			Name:      fmt.Sprintf("helperNumber%02d", i),
			Signature: fmt.Sprintf("func helperNumber%02d(ctx context.Context, db *sql.DB, id uuid.UUID) error", i),
		})
	}
	nav := analysis.FileInfo{
		Path: "populate_nav_tables_action.go",
		Functions: []analysis.FuncDef{
			{Name: "isLegalPage", Signature: "func isLegalPage(nameLower string) bool"},
			{Name: "loadPagesForNav", Signature: "func loadPagesForNav(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) ([]pageNavInfo, error)"},
		},
	}
	out := analysis.Output{Files: []analysis.FileInfo{big, nav}}
	got := siblingSignatures(out,
		[]string{"a_giant_file.go:helperNumber00", "populate_nav_tables_action.go:isLegalPage"}, 3000)
	if !strings.Contains(got, "loadPagesForNav") {
		t.Fatalf("fair share must reach the later file's siblings:\n%s", got)
	}
	if !strings.Contains(got, "more in this file") {
		t.Fatalf("the truncated giant must carry a +N-more marker:\n%s", got)
	}
}

// The cap must report what it excluded (council-gate eba040a9 round 5): a bundle
// that inlines 2 of 5 named steps, with nothing saying so, lets the verdicter
// read "not inlined" as "not involved".
func TestWorkflowRefsFromRuntimeReportsExcluded(t *testing.T) {
	runtime := "a/x (act) b/y (act) c/z (act) d/w (act) e/v (act)"
	refs, excluded := workflowRefsFromRuntime(runtime, 2)
	if len(refs) != 2 {
		t.Fatalf("cap not honoured: %v", refs)
	}
	if excluded != 3 {
		t.Fatalf("want 3 excluded, got %d", excluded)
	}
	if _, none := workflowRefsFromRuntime("a/x (act)", 4); none != 0 {
		t.Fatalf("nothing excluded under the cap; got %d", none)
	}
	// Repeats are deduped BEFORE the cap, so they must not inflate the count.
	if _, dup := workflowRefsFromRuntime("a/x (act) a/x (act) a/x (act)", 1); dup != 0 {
		t.Fatalf("duplicate refs must not count as excluded; got %d", dup)
	}
}
