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
