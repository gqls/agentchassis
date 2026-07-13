package diagnose

import "testing"

func docsContain(docs []string, want string) bool {
	for _, d := range docs {
		if d == want {
			return true
		}
	}
	return false
}

func TestSelectDocs_KeywordAndAlways(t *testing.T) {
	rules := []DocRule{
		{Doc: "constitution_adjacent.md", Always: true},
		{Doc: "component_linkage.md", Keywords: []string{"component", "linkage", "unlinked"}},
		{Doc: "css_inheritance.md", PathGlobs: []string{"theme", "css"}},
		{Doc: "never.md", Keywords: []string{"zzz"}},
		{Doc: "empty_rule.md"}, // no conditions, not Always -> never matches
	}

	got := SelectDocs("the component linkage is wrong on the hero", Scope{Symbols: []string{"x.go:Foo"}}, rules)
	for _, want := range []string{"component_linkage.md", "constitution_adjacent.md"} {
		if !docsContain(got, want) {
			t.Errorf("expected %q in %v", want, got)
		}
	}
	for _, absent := range []string{"css_inheritance.md", "never.md", "empty_rule.md"} {
		if docsContain(got, absent) {
			t.Errorf("did not expect %q in %v", absent, got)
		}
	}
}

func TestSelectDocs_PathGlob(t *testing.T) {
	rules := []DocRule{
		{Doc: "css_inheritance.md", PathGlobs: []string{"theme", "css"}},
		{Doc: "component_linkage.md", Keywords: []string{"component"}},
	}
	// hypothesis has no matching keyword; the scope PATH carries the signal
	got := SelectDocs("something went wrong somewhere", Scope{Symbols: []string{"platform/orchestration/actions/theme_actions.go:ApplyTheme"}}, rules)
	if !docsContain(got, "css_inheritance.md") {
		t.Errorf("expected css_inheritance.md via path glob, got %v", got)
	}
	if docsContain(got, "component_linkage.md") {
		t.Errorf("did not expect component_linkage.md, got %v", got)
	}
}

func TestSelectDocs_DedupAndStable(t *testing.T) {
	// the same doc matched by two rules must appear once
	rules := []DocRule{
		{Doc: "d.md", Keywords: []string{"alpha"}},
		{Doc: "d.md", PathGlobs: []string{"x.go"}},
		{Doc: "a.md", Always: true},
	}
	got := SelectDocs("alpha beta", Scope{Symbols: []string{"x.go"}}, rules)
	if len(got) != 2 {
		t.Fatalf("expected 2 unique docs, got %v", got)
	}
	// sorted: a.md before d.md
	if got[0] != "a.md" || got[1] != "d.md" {
		t.Fatalf("expected stable sorted [a.md d.md], got %v", got)
	}
}

func TestSelectDocs_NoRulesNoDocs(t *testing.T) {
	if got := SelectDocs("anything", Scope{Symbols: []string{"x.go"}}, nil); len(got) != 0 {
		t.Fatalf("expected no docs from nil rules, got %v", got)
	}
}
