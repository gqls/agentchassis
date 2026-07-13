package diagnose

import (
	"os"
	"strings"
	"testing"

	"contextkit/internal/analysis"
)

func TestCallGraph_ResolvesAndDropsUbiquitous(t *testing.T) {
	// Hand-built analysis: foo calls bar (resolvable) + Run (ubiquitous, dropped).
	out := analysis.Output{Files: []analysis.FileInfo{
		{Path: "a.go", Functions: []analysis.FuncDef{
			{Name: "foo", Calls: []string{"bar", "Run", "baz"}},
			{Name: "bar", Calls: []string{"qux"}},
		}},
		{Path: "b.go", Functions: []analysis.FuncDef{
			{Name: "baz", Calls: nil},
			{Name: "qux", Calls: nil},
		}},
	}}
	g := NewCallGraph(out)

	nb := g.Neighbourhood("a.go:foo")
	// foo -> bar (a.go:bar), baz (b.go:baz); Run dropped (ubiquitous)
	joined := strings.Join(nb, ",")
	if !strings.Contains(joined, "a.go:bar") || !strings.Contains(joined, "b.go:baz") {
		t.Fatalf("expected bar+baz in neighbourhood, got %v", nb)
	}
	for _, s := range nb {
		if strings.HasSuffix(s, ":Run") {
			t.Fatalf("ubiquitous Run should have been dropped, got %v", nb)
		}
	}
}

func TestCallGraph_RealAnalysis(t *testing.T) {
	// Uses the analysis built in the shell step (/tmp/diag_analysis.json).
	if _, err := os.Stat("/tmp/diag_analysis.json"); err != nil {
		t.Skip("no /tmp/diag_analysis.json (run the analyser step first)")
	}
	g, err := NewCallGraphFromFile("/tmp/diag_analysis.json")
	if err != nil {
		t.Fatal(err)
	}
	// NewCallGraphFromFile calls NewCallGraph (resolvable in-package) + ReadFile,
	// Unmarshal (external, no def in this analysis → absent from neighbourhood).
	nb := g.Neighbourhood("callgraph.go:NewCallGraphFromFile")
	found := false
	for _, s := range nb {
		if strings.HasSuffix(s, ":NewCallGraph") {
			found = true
		}
		if strings.HasSuffix(s, ":ReadFile") || strings.HasSuffix(s, ":Unmarshal") {
			t.Fatalf("external callee should not resolve to a def: %v", nb)
		}
	}
	if !found {
		t.Fatalf("expected NewCallGraph in neighbourhood, got %v", nb)
	}
}

func TestGatherer_ScopeToFlags_DryRun(t *testing.T) {
	dir := t.TempDir()
	g := &BundleGatherer{
		BundleBin:    "./cmd/bundle",
		UseGoRun:     true,
		AnalysisPath: "/tmp/chassis_clean.json",
		Root:         "/repo",
		Constitution: "const.md",
		Docs:         []string{"docs/016_debugging_guide.md", "docs/dev_guide_s3.md"},
		SchemaTables: []string{"site_specs", "pages"}, // constant domain schema; "pages" also in scope below → dedup
		DocRules: []DocRule{
			{Doc: "matched_by_keyword.md", Keywords: []string{"h"}}, // hypothesis is "h"
			{Doc: "matched_by_path.md", PathGlobs: []string{"a.go"}}, // scope has a.go:Foo
			{Doc: "not_matched.md", Keywords: []string{"zzz"}, PathGlobs: []string{"qqq"}},
		},
		Psql:         "kubectl exec -n ai-persona-system pg -- psql -U u -d d",
		OutDir:       dir,
		DryRun:       true,
	}
	scope := Scope{
		Symbols:      []string{"a.go:Foo", "b.go"},
		Tables:       []string{"pages", "page_components"},
		RuntimeSite:  "gamesdesign.co.uk",
		RuntimePage:  "index",
		Capabilities: true,
	}
	path, err := g.Gather("h", scope)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	cmd := string(raw)
	// the translated command must carry each scope element as the right bundle flag
	for _, want := range []string{
		"-analysis /tmp/chassis_clean.json",
		"-root /repo",
		"-task h", // the hypothesis MUST be forwarded as bundle's required -task
		"-scope a.go:Foo",
		"-scope b.go",
		"-doc docs/016_debugging_guide.md", // authored context forwarded to the bundle
		"-doc docs/dev_guide_s3.md",
		"-doc matched_by_keyword.md", // per-hypothesis: keyword matched "h"
		"-doc matched_by_path.md",    // per-hypothesis: path glob matched a.go
		"-schema-tables site_specs,pages,page_components", // constant SchemaTables first, then scope; "pages" deduped
		"-runtime-site gamesdesign.co.uk",
		"-runtime-page index",
		"-capabilities",
		"-step debug",
	} {
		if !strings.Contains(cmd, want) {
			t.Fatalf("gathered command missing %q\ncmd: %s", want, cmd)
		}
	}
	// the non-matching catalogue doc must NOT be included — proves SelectDocs filters
	if strings.Contains(cmd, "not_matched.md") {
		t.Fatalf("non-matching doc leaked into the bundle command\ncmd: %s", cmd)
	}
}

func TestGatherer_NoPsql_SkipsDBFlags(t *testing.T) {
	dir := t.TempDir()
	g := &BundleGatherer{
		BundleBin: "./cmd/bundle", UseGoRun: true,
		AnalysisPath: "a.json", Root: "/r", Constitution: "c.md",
		OutDir: dir, DryRun: true, // no Psql
	}
	path, err := g.Gather("h", Scope{Symbols: []string{"x.go"}, Tables: []string{"pages"}, RuntimeSite: "s"})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(path)
	cmd := string(raw)
	// without -psql, bundle skips the DB gather → these flags must NOT appear
	for _, absent := range []string{"-schema-tables", "-runtime-site", "-psql"} {
		if strings.Contains(cmd, absent) {
			t.Fatalf("no-psql gather should omit %q, got: %s", absent, cmd)
		}
	}
	if !strings.Contains(cmd, "-scope x.go") {
		t.Fatalf("scope should still be present: %s", cmd)
	}
}
