package diagnose

import "testing"

func queriesHave(qs []QueryRule, name string) bool {
	for _, q := range qs {
		if q.Name == name {
			return true
		}
	}
	return false
}

func TestSelectQueries_KeywordAndAlways(t *testing.T) {
	cat := []QueryRule{
		{Name: "recent_errors", Always: true, SQL: "SELECT 1"},
		{Name: "page_components_for_page", Keywords: []string{"component", "placeholder", "render"}, SQL: "SELECT 1", Params: []string{"site_id", "page_name"}},
		{Name: "work_items_recent", Keywords: []string{"work item", "claim", "timeout"}, SQL: "SELECT 1", Params: []string{"site_id"}},
		{Name: "never", Keywords: []string{"zzz"}, SQL: "SELECT 1"},
	}
	got := SelectQueries("the index page components show placeholder text", Scope{Symbols: []string{"x.go:Foo"}}, cat)
	for _, want := range []string{"recent_errors", "page_components_for_page"} {
		if !queriesHave(got, want) {
			t.Errorf("expected %q selected, got %v", want, names(got))
		}
	}
	for _, absent := range []string{"work_items_recent", "never"} {
		if queriesHave(got, absent) {
			t.Errorf("did not expect %q, got %v", absent, names(got))
		}
	}
}

func TestSelectQueries_PathGlob(t *testing.T) {
	cat := []QueryRule{
		{Name: "work_items_recent", PathGlobs: []string{"dispatch", "work_item"}, SQL: "SELECT 1"},
		{Name: "page_components_for_page", Keywords: []string{"component"}, SQL: "SELECT 1"},
	}
	got := SelectQueries("unrelated text", Scope{Symbols: []string{"platform/orchestration/actions/build_dispatch_loop.go:Run"}}, cat)
	if !queriesHave(got, "work_items_recent") {
		t.Errorf("expected work_items_recent via path glob, got %v", names(got))
	}
	if queriesHave(got, "page_components_for_page") {
		t.Errorf("did not expect page_components_for_page, got %v", names(got))
	}
}

func TestSelectQueries_DedupByName(t *testing.T) {
	cat := []QueryRule{
		{Name: "dup", Keywords: []string{"alpha"}, SQL: "SELECT 1"},
		{Name: "dup", PathGlobs: []string{"x.go"}, SQL: "SELECT 2"}, // same Name, would match too
	}
	got := SelectQueries("alpha", Scope{Symbols: []string{"x.go"}}, cat)
	if len(got) != 1 {
		t.Fatalf("expected 1 query after dedup-by-name, got %v", names(got))
	}
}

func TestSelectQueries_NilCatalogue(t *testing.T) {
	if got := SelectQueries("anything", Scope{Symbols: []string{"x.go"}}, nil); len(got) != 0 {
		t.Fatalf("expected no queries from nil catalogue, got %v", names(got))
	}
}

func names(qs []QueryRule) []string {
	out := make([]string, 0, len(qs))
	for _, q := range qs {
		out = append(out, q.Name)
	}
	return out
}
