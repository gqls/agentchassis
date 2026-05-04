// FILE: platform/orchestration/datahelpers/page_canonical_test.go

package datahelpers

import "testing"

func TestCanonicalisePage(t *testing.T) {
	cases := []struct {
		name     string
		in       PageDescriptor
		wantName string
		wantURL  string
		wantType string
	}{
		// ── Tools: planner shape vs adoption shape converge ────────────
		{
			name:     "tool planner shape",
			in:       PageDescriptor{Role: "tool", Slug: "ttk-calculator"},
			wantName: "tool-ttk-calculator",
			wantURL:  "/tools/ttk-calculator/index.html",
			wantType: "tool",
		},
		{
			name:     "tool adoption shape (already prefixed) is idempotent",
			in:       PageDescriptor{Role: "tool", Slug: "tool-ttk-calculator"},
			wantName: "tool-ttk-calculator",
			wantURL:  "/tools/ttk-calculator/index.html",
			wantType: "tool",
		},
		{
			name:     "tool with mixed case and space",
			in:       PageDescriptor{Role: "Tool", Slug: "TTK Calculator"},
			wantName: "tool-ttk-calculator",
			wantURL:  "/tools/ttk-calculator/index.html",
			wantType: "tool",
		},

		// ── Guides ─────────────────────────────────────────────────────
		{
			name:     "guide planner shape",
			in:       PageDescriptor{Role: "guide", Slug: "rng-design"},
			wantName: "guide-rng-design",
			wantURL:  "/guides/rng-design/index.html",
			wantType: "guide",
		},
		{
			name:     "guide adoption shape",
			in:       PageDescriptor{Role: "guide", Slug: "guide-rng-design"},
			wantName: "guide-rng-design",
			wantURL:  "/guides/rng-design/index.html",
			wantType: "guide",
		},

		// ── Games ──────────────────────────────────────────────────────
		{
			name:     "game planner shape",
			in:       PageDescriptor{Role: "game", Slug: "jelly-invaders"},
			wantName: "game-jelly-invaders",
			wantURL:  "/games/jelly-invaders/index.html",
			wantType: "game",
		},
		{
			name:     "game adoption shape",
			in:       PageDescriptor{Role: "game", Slug: "game-jelly-invaders"},
			wantName: "game-jelly-invaders",
			wantURL:  "/games/jelly-invaders/index.html",
			wantType: "game",
		},

		// ── Section indexes ────────────────────────────────────────────
		{
			name:     "section_index adoption shape (guides-index)",
			in:       PageDescriptor{Role: "section_index", Slug: "guides-index"},
			wantName: "guides-index",
			wantURL:  "/guides/index.html",
			wantType: "section_index",
		},
		{
			name:     "section_index planner shape (guides)",
			in:       PageDescriptor{Role: "section_index", Slug: "guides"},
			wantName: "guides-index",
			wantURL:  "/guides/index.html",
			wantType: "section_index",
		},
		{
			name:     "section_index with explicit Section overriding slug",
			in:       PageDescriptor{Role: "section_index", Slug: "irrelevant", Section: "tools"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "section_index",
		},
		{
			name:     "section_index for tools",
			in:       PageDescriptor{Role: "section_index", Slug: "tools-index"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "section_index",
		},

		// ── Content ────────────────────────────────────────────────────
		{
			name:     "content about",
			in:       PageDescriptor{Role: "content", Slug: "about"},
			wantName: "about",
			wantURL:  "/about.html",
			wantType: "content",
		},
		{
			name:     "content with empty role defaults to content",
			in:       PageDescriptor{Role: "", Slug: "contact"},
			wantName: "contact",
			wantURL:  "/contact.html",
			wantType: "content",
		},

		// ── Blog ───────────────────────────────────────────────────────
		{
			name:     "blog_post keeps adoption convention",
			in:       PageDescriptor{Role: "blog_post", Slug: "first-post"},
			wantName: "first-post",
			wantURL:  "/blog/first-post.html",
			wantType: "blog_post",
		},

		// ── Index / home aliasing ──────────────────────────────────────
		{
			name:     "explicit index role",
			in:       PageDescriptor{Role: "index", Slug: ""},
			wantName: "index",
			wantURL:  "/index.html",
			wantType: "index",
		},
		{
			name:     "content role with home slug collapses to index",
			in:       PageDescriptor{Role: "content", Slug: "home"},
			wantName: "index",
			wantURL:  "/index.html",
			wantType: "index",
		},
		{
			name:     "content role with index slug collapses to index",
			in:       PageDescriptor{Role: "content", Slug: "index"},
			wantName: "index",
			wantURL:  "/index.html",
			wantType: "index",
		},

		// ── Tolerance: URL fragments and trailing extensions ───────────
		{
			name:     "slug with leaked .html suffix",
			in:       PageDescriptor{Role: "content", Slug: "about.html"},
			wantName: "about",
			wantURL:  "/about.html",
			wantType: "content",
		},
		{
			name:     "slug with leaked path prefix",
			in:       PageDescriptor{Role: "tool", Slug: "/tools/ttk-calculator"},
			wantName: "tool-ttk-calculator",
			wantURL:  "/tools/ttk-calculator/index.html",
			wantType: "tool",
		},

		// ── Unknown role passes through but follows content URL shape ──
		{
			name:     "unknown role preserves role as page_type",
			in:       PageDescriptor{Role: "landing", Slug: "promo"},
			wantName: "promo",
			wantURL:  "/promo.html",
			wantType: "landing",
		},

		// ── Empty inputs ───────────────────────────────────────────────
		{
			name:     "empty slug under tool returns empty triple",
			in:       PageDescriptor{Role: "tool", Slug: ""},
			wantName: "",
			wantURL:  "",
			wantType: "",
		},
		{
			name:     "empty slug under section_index with no section",
			in:       PageDescriptor{Role: "section_index", Slug: ""},
			wantName: "",
			wantURL:  "",
			wantType: "",
		},

		// ── Phase 1: ParentSection for nested-URL synthesis ────────────
		{
			name:     "tool with custom parent_section nests under that dir",
			in:       PageDescriptor{Role: "tool", Slug: "calculator", ParentSection: "utilities"},
			wantName: "tool-calculator",
			wantURL:  "/utilities/calculator/index.html",
			wantType: "tool",
		},
		{
			name:     "tool with empty parent_section keeps default tools dir",
			in:       PageDescriptor{Role: "tool", Slug: "calculator", ParentSection: ""},
			wantName: "tool-calculator",
			wantURL:  "/tools/calculator/index.html",
			wantType: "tool",
		},
		{
			name:     "guide with parent_section having -index suffix has it stripped",
			in:       PageDescriptor{Role: "guide", Slug: "intro", ParentSection: "learn-index"},
			wantName: "guide-intro",
			wantURL:  "/learn/intro/index.html",
			wantType: "guide",
		},
		{
			name:     "game with custom parent",
			in:       PageDescriptor{Role: "game", Slug: "snake", ParentSection: "playables"},
			wantName: "game-snake",
			wantURL:  "/playables/snake/index.html",
			wantType: "game",
		},
		{
			name:     "blog_post with parent_section nests under that dir",
			in:       PageDescriptor{Role: "blog_post", Slug: "first-post", ParentSection: "guides"},
			wantName: "first-post",
			wantURL:  "/guides/first-post.html",
			wantType: "blog_post",
		},
		{
			name:     "section_index ignores parent_section (a section IS itself)",
			in:       PageDescriptor{Role: "section_index", Slug: "tools", ParentSection: "ignored"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "section_index",
		},
		{
			name:     "content ignores parent_section",
			in:       PageDescriptor{Role: "content", Slug: "team", ParentSection: "company"},
			wantName: "team",
			wantURL:  "/team.html",
			wantType: "content",
		},
		{
			name:     "tool with adoption-shape slug AND custom parent works",
			in:       PageDescriptor{Role: "tool", Slug: "tool-calculator", ParentSection: "utilities"},
			wantName: "tool-calculator",
			wantURL:  "/utilities/calculator/index.html",
			wantType: "tool",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotName, gotURL, gotType := CanonicalisePage(tc.in)
			if gotName != tc.wantName || gotURL != tc.wantURL || gotType != tc.wantType {
				t.Errorf("CanonicalisePage(%+v)\n  got  (%q, %q, %q)\n  want (%q, %q, %q)",
					tc.in, gotName, gotURL, gotType,
					tc.wantName, tc.wantURL, tc.wantType)
			}
		})
	}
}
