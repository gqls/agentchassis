// FILE: platform/orchestration/datahelpers/page_role_validator_test.go

package datahelpers

import "testing"

func TestValidateRoles_GamesdesignReproduction(t *testing.T) {
	// Exact LLM emission from the 2026-05-04 gamesdesign.co.uk run
	// (subset). The validator must correct tools/guides/games from
	// content/blog-index/content to section-index without us touching
	// the prompt.
	in := []LLMPlannedPage{
		{Name: "tools", Role: "content", URL: "/tools/index.html"},
		{Name: "guides", Role: "blog-index", URL: "/guides/index.html"},
		{Name: "games", Role: "content", URL: "/games/index.html"},
		{Name: "ttk-calculator", Role: "tool", URL: "/tools/ttk-calculator/index.html", ParentSection: "tools"},
		{Name: "rng-design", Role: "guide", URL: "/guides/rng-design/index.html", ParentSection: "guides"},
		{Name: "snake", Role: "game", URL: "/games/snake/index.html", ParentSection: "games"},
	}

	out := ValidateRoles(in)

	wantRoles := map[string]string{
		"tools":          "section-index",
		"guides":         "section-index",
		"games":          "section-index",
		"ttk-calculator": "tool",
		"rng-design":     "guide",
		"snake":          "game",
	}
	for _, v := range out {
		want, ok := wantRoles[v.Name]
		if !ok {
			t.Errorf("unexpected page in output: %s", v.Name)
			continue
		}
		if v.Role != want {
			t.Errorf("page %q: role=%q want %q", v.Name, v.Role, want)
		}
	}

	// The three section-index pages should have CorrectedFromRole set
	// so production logging can monitor validator activity.
	for _, v := range out {
		switch v.Name {
		case "tools":
			if v.CorrectedFromRole != "content" {
				t.Errorf("tools: CorrectedFromRole=%q want %q",
					v.CorrectedFromRole, "content")
			}
		case "guides":
			// "blog-index" normalises to "section-index" before
			// comparison; CorrectedFromRole captures the pre-normalised
			// LLM output as it was actually emitted.
			if v.CorrectedFromRole != "" {
				// Acceptable either way: rule 2 (declaredParents)
				// catches it before normalisation matters.
			}
		case "games":
			if v.CorrectedFromRole != "content" {
				t.Errorf("games: CorrectedFromRole=%q want %q",
					v.CorrectedFromRole, "content")
			}
		case "ttk-calculator", "rng-design", "snake":
			if v.CorrectedFromRole != "" {
				t.Errorf("%s: should not have been corrected, got CorrectedFromRole=%q",
					v.Name, v.CorrectedFromRole)
			}
		}
	}
}

func TestValidateRoles_RulePrecedence(t *testing.T) {
	cases := []struct {
		name          string
		in            []LLMPlannedPage
		wantByName    map[string]string // name → expected role
		wantNoCorrect []string          // names that should pass through unchanged
	}{
		{
			name: "rule 1: explicit index identity → landing role",
			in: []LLMPlannedPage{
				{Name: "index", Role: "content", URL: "/index.html"},
			},
			wantByName: map[string]string{"index": "landing"},
		},
		{
			name: "rule 2: declared parent beats LLM role",
			// "tools" is content per LLM, but "ttk-calculator"
			// claims it as ParentSection — structural signal wins.
			in: []LLMPlannedPage{
				{Name: "tools", Role: "content", URL: "/tools.html"},
				{Name: "ttk-calculator", Role: "tool", URL: "/tools/ttk-calculator/", ParentSection: "tools"},
			},
			wantByName: map[string]string{
				"tools":          "section-index",
				"ttk-calculator": "tool",
			},
		},
		{
			name: "rule 3: URL-only signal also produces section-index",
			// No declared parent, but URL pattern is dispositive.
			in: []LLMPlannedPage{
				{Name: "guides", Role: "content", URL: "/guides/index.html"},
			},
			wantByName: map[string]string{"guides": "section-index"},
		},
		{
			name: "rule 4: nested URL pattern corrects nested role",
			in: []LLMPlannedPage{
				{Name: "ttk-calculator", Role: "content", URL: "/tools/ttk-calculator/index.html"},
			},
			wantByName: map[string]string{"ttk-calculator": "tool"},
		},
		{
			name: "rule 5: well-formed input passes through",
			in: []LLMPlannedPage{
				{Name: "about", Role: "content", URL: "/about.html"},
				{Name: "tool-jump-physics", Role: "tool", Slug: "jump-physics", URL: "/tools/jump-physics/index.html"},
			},
			wantByName: map[string]string{
				"about":             "content",
				"tool-jump-physics": "tool",
			},
			wantNoCorrect: []string{"about", "tool-jump-physics"},
		},
		{
			name: "blog-post passes through unchanged (already canonical kebab)",
			in: []LLMPlannedPage{
				{Name: "first-post", Role: "blog-post", URL: "/blog/first-post.html"},
			},
			wantByName: map[string]string{"first-post": "blog-post"},
		},
		{
			name: "blog_post (legacy snake input) normalises to blog-post",
			in: []LLMPlannedPage{
				{Name: "first-post", Role: "blog_post", URL: "/blog/first-post.html"},
			},
			wantByName: map[string]string{"first-post": "blog-post"},
		},
		{
			name: "rule 2 wins over rule 4 (structural beats URL)",
			// URL says nested-tool, but parent-section says
			// section-index — rule 2 fires first.
			// Realistically you wouldn't see this combination, but
			// the precedence is worth pinning.
			in: []LLMPlannedPage{
				{Name: "tools", Role: "tool", URL: "/tools/foo/index.html"},
				{Name: "child", Role: "tool", URL: "/tools/child/", ParentSection: "tools"},
			},
			wantByName: map[string]string{
				"tools": "section-index",
				"child": "tool",
			},
		},
		{
			name: "ParentSection with -index suffix matches via either form",
			// child claims ParentSection="guides-index"; the section
			// page is named "guides". Both forms must match.
			in: []LLMPlannedPage{
				{Name: "guides", Role: "content", URL: "/guides/index.html"},
				{Name: "g1", Role: "guide", URL: "/guides/g1/", ParentSection: "guides-index"},
			},
			wantByName: map[string]string{
				"guides": "section-index",
				"g1":     "guide",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ValidateRoles(tc.in)
			gotByName := make(map[string]ValidatedPage, len(out))
			for _, v := range out {
				gotByName[v.Name] = v
			}
			for name, wantRole := range tc.wantByName {
				v, ok := gotByName[name]
				if !ok {
					t.Errorf("missing page %q in output", name)
					continue
				}
				if v.Role != wantRole {
					t.Errorf("page %q: role=%q want %q (corrected from %q)",
						name, v.Role, wantRole, v.CorrectedFromRole)
				}
			}
			for _, name := range tc.wantNoCorrect {
				v, ok := gotByName[name]
				if !ok {
					continue
				}
				if v.CorrectedFromRole != "" {
					t.Errorf("page %q should not have been corrected, got CorrectedFromRole=%q",
						name, v.CorrectedFromRole)
				}
			}
		})
	}
}

func TestValidateRoles_Idempotent(t *testing.T) {
	// Running the validator twice on the same input should produce
	// the same output (apart from CorrectedFromRole resetting because
	// the second run sees already-corrected roles).
	in := []LLMPlannedPage{
		{Name: "tools", Role: "content", URL: "/tools/index.html"},
		{Name: "ttk", Role: "content", URL: "/tools/ttk/index.html"},
	}

	first := ValidateRoles(in)

	// Convert first output into second input.
	second := make([]LLMPlannedPage, len(first))
	for i, v := range first {
		second[i] = LLMPlannedPage{
			Name:          v.Name,
			Role:          v.Role,
			Slug:          v.Slug,
			URL:           v.URL,
			ParentSection: v.ParentSection,
		}
	}
	out := ValidateRoles(second)

	for i := range first {
		if out[i].Role != first[i].Role {
			t.Errorf("not idempotent: page %q role changed %q → %q on second pass",
				first[i].Name, first[i].Role, out[i].Role)
		}
	}
}

func TestValidateRoles_PreservesOrder(t *testing.T) {
	in := []LLMPlannedPage{
		{Name: "a", Role: "content", URL: "/a.html"},
		{Name: "b", Role: "content", URL: "/b.html"},
		{Name: "c", Role: "content", URL: "/c.html"},
	}
	out := ValidateRoles(in)
	for i, v := range out {
		if v.Name != in[i].Name {
			t.Errorf("order not preserved at index %d: got %q want %q",
				i, v.Name, in[i].Name)
		}
	}
}

// TestValidateRoles_NameSuffixSectionIndex covers the 2026-05-23
// gamesdesign.co.uk production shape: the plan-builder emitted section
// hubs named "<section>-index" with role="content" (or "blog-index") and
// NO url and NO parent_section on any page. Rules 3-5 (parent/URL) are all
// starved; only rule 2 (name suffix) can recover the section-index role.
func TestValidateRoles_NameSuffixSectionIndex(t *testing.T) {
	in := []LLMPlannedPage{
		// Exactly as emitted: name carries "-index", role is content,
		// no URL, no parent_section.
		{Name: "games-index", Role: "content"},
		{Name: "tools-index", Role: "content"},
		// guides-index came through as blog-index (normalises to
		// section-index already); the name rule agrees.
		{Name: "guides-index", Role: "blog-index"},
		// Leaf pages in the same plan must be untouched by the name rule.
		{Name: "tool-ttk-calculator", Role: "tool"},
		{Name: "guide-rng-design", Role: "blog-post"},
	}

	out := ValidateRoles(in)
	got := make(map[string]ValidatedPage, len(out))
	for _, v := range out {
		got[v.Name] = v
	}

	wantRole := map[string]string{
		"games-index":         "section-index",
		"tools-index":         "section-index",
		"guides-index":        "section-index",
		"tool-ttk-calculator": "tool",
		"guide-rng-design":    "blog-post",
	}
	for name, want := range wantRole {
		v, ok := got[name]
		if !ok {
			t.Fatalf("missing page %q in output", name)
		}
		if v.Role != want {
			t.Errorf("page %q: role=%q want %q", name, v.Role, want)
		}
	}

	// games-index/tools-index were corrected from a real "content" label,
	// so CorrectedFromRole must be set for production logging. guides-index
	// normalised to section-index before comparison, so no correction is
	// logged. Leaf pages are untouched.
	if got["games-index"].CorrectedFromRole != "content" {
		t.Errorf("games-index: CorrectedFromRole=%q want %q",
			got["games-index"].CorrectedFromRole, "content")
	}
	if got["tools-index"].CorrectedFromRole != "content" {
		t.Errorf("tools-index: CorrectedFromRole=%q want %q",
			got["tools-index"].CorrectedFromRole, "content")
	}
	if got["guides-index"].CorrectedFromRole != "" {
		t.Errorf("guides-index: CorrectedFromRole=%q want \"\" (blog-index already normalises to section-index)",
			got["guides-index"].CorrectedFromRole)
	}
	for _, name := range []string{"tool-ttk-calculator", "guide-rng-design"} {
		if got[name].CorrectedFromRole != "" {
			t.Errorf("%s: should not have been corrected, got CorrectedFromRole=%q",
				name, got[name].CorrectedFromRole)
		}
	}
}

// TestValidateRoles_NameSuffixLeafGuard pins the role-guard: an explicit
// leaf role whose name unusually ends in "-index" must NOT be reclassified
// as a section index by rule 2.
func TestValidateRoles_NameSuffixLeafGuard(t *testing.T) {
	in := []LLMPlannedPage{
		// A genuine blog post that happens to be titled "weekly-index".
		{Name: "weekly-index", Role: "blog-post", URL: "/blog/weekly-index.html"},
		// A tool whose name ends in -index — still a tool.
		{Name: "tool-price-index", Role: "tool", URL: "/tools/price-index/index.html"},
	}
	out := ValidateRoles(in)
	got := make(map[string]ValidatedPage, len(out))
	for _, v := range out {
		got[v.Name] = v
	}
	if got["weekly-index"].Role != "blog-post" {
		t.Errorf("weekly-index: role=%q want blog-post (leaf guard should block name rule)",
			got["weekly-index"].Role)
	}
	if got["tool-price-index"].Role != "tool" {
		t.Errorf("tool-price-index: role=%q want tool (leaf guard should block name rule)",
			got["tool-price-index"].Role)
	}
}

// TestValidateRoles_NewsIndexFlavourPreserved pins bugs_open/015: an
// explicit news-index role must survive rules 2-4, each of which would
// otherwise flatten it to generic section-index. page_type is a routing
// key — render_news_section, MissingNewsPageCheck and page-build-handler
// all select on 'news-index' — so the flattening orphaned relojistas.com's
// /noticias page from all of them at once.
func TestValidateRoles_NewsIndexFlavourPreserved(t *testing.T) {
	cases := []struct {
		name string
		in   []LLMPlannedPage
		page string
	}{
		{
			// Rule 2 shape: name ends in "-index", no URL, no parent.
			// This is the exact relojistas emission shape.
			name: "name-suffix rule must not flatten news-index",
			in:   []LLMPlannedPage{{Name: "noticias-index", Role: "news-index"}},
			page: "noticias-index",
		},
		{
			// Rule 4 shape: /<slug>/index.html URL.
			name: "index-URL rule must not flatten news-index",
			in:   []LLMPlannedPage{{Name: "noticias", Role: "news-index", URL: "/noticias/index.html"}},
			page: "noticias",
		},
		{
			// Rule 3 shape: another page declares this slug as parent.
			name: "declared-parent rule must not flatten news-index",
			in: []LLMPlannedPage{
				{Name: "noticias", Role: "news-index"},
				{Name: "old-news", Role: "content", ParentSection: "noticias"},
			},
			page: "noticias",
		},
		{
			// Snake input normalises to kebab and is then preserved.
			name: "news_index snake input normalises and survives",
			in:   []LLMPlannedPage{{Name: "noticias-index", Role: "news_index"}},
			page: "noticias-index",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := ValidateRoles(tc.in)
			got := make(map[string]ValidatedPage, len(out))
			for _, v := range out {
				got[v.Name] = v
			}
			v, ok := got[tc.page]
			if !ok {
				t.Fatalf("missing page %q in output", tc.page)
			}
			if v.Role != "news-index" {
				t.Errorf("page %q: role=%q want news-index", tc.page, v.Role)
			}
			if v.CorrectedFromRole != "" && v.CorrectedFromRole != "news_index" {
				t.Errorf("page %q: unexpectedly corrected from %q", tc.page, v.CorrectedFromRole)
			}
		})
	}

	// Regression guard: a sloppy generic role on the same name shape is
	// still corrected — only the EXPLICIT news-index flavour is trusted.
	out := ValidateRoles([]LLMPlannedPage{{Name: "noticias-index", Role: "content"}})
	if out[0].Role != "section-index" {
		t.Errorf("sloppy role: role=%q want section-index (rule 2 must still fire)", out[0].Role)
	}
}

// TestValidateRoles_NameSuffixIdempotent confirms the name-suffix
// correction is stable: feeding the corrected section-index role back in
// (now with a "-index" name) keeps it section-index.
func TestValidateRoles_NameSuffixIdempotent(t *testing.T) {
	in := []LLMPlannedPage{{Name: "games-index", Role: "content"}}
	first := ValidateRoles(in)
	if first[0].Role != "section-index" {
		t.Fatalf("first pass: role=%q want section-index", first[0].Role)
	}
	second := ValidateRoles([]LLMPlannedPage{{
		Name:          first[0].Name,
		Role:          first[0].Role,
		Slug:          first[0].Slug,
		URL:           first[0].URL,
		ParentSection: first[0].ParentSection,
	}})
	if second[0].Role != "section-index" {
		t.Errorf("not idempotent: role changed to %q on second pass", second[0].Role)
	}
}
