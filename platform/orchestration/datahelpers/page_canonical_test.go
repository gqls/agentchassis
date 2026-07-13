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

		// ── Section indexes (kebab canonical) ──────────────────────────
		{
			name:     "section-index adoption shape (guides-index)",
			in:       PageDescriptor{Role: "section-index", Slug: "guides-index"},
			wantName: "guides-index",
			wantURL:  "/guides/index.html",
			wantType: "section-index",
		},
		{
			name:     "section-index planner shape (guides)",
			in:       PageDescriptor{Role: "section-index", Slug: "guides"},
			wantName: "guides-index",
			wantURL:  "/guides/index.html",
			wantType: "section-index",
		},
		{
			name:     "section-index with explicit Section overriding slug",
			in:       PageDescriptor{Role: "section-index", Slug: "irrelevant", Section: "tools"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "section-index",
		},
		{
			name:     "section-index for tools",
			in:       PageDescriptor{Role: "section-index", Slug: "tools-index"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "section-index",
		},

		// ── Section-index family convergence ───────────────────────────
		// Adoption emits role=blog-index for what the planner emits as
		// role=section-index. Same logical page; canonicaliser must
		// produce the same name and URL. page_type is preserved from
		// the input so downstream dispatch can distinguish flavours.
		{
			name:     "blog-index and section-index converge on name+url for slug=guides",
			in:       PageDescriptor{Role: "blog-index", Slug: "guides"},
			wantName: "guides-index",
			wantURL:  "/guides/index.html",
			wantType: "blog-index", // page_type retains the input role
		},
		{
			name:     "blog_index (legacy snake input) normalises to blog-index output",
			in:       PageDescriptor{Role: "blog_index", Slug: "guides"},
			wantName: "guides-index",
			wantURL:  "/guides/index.html",
			wantType: "blog-index",
		},
		{
			name:     "entity-directory: same shape as section-index, different page_type",
			in:       PageDescriptor{Role: "entity-directory", Slug: "tools"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "entity-directory",
		},
		{
			name:     "entity_directory (legacy snake input) normalises to entity-directory output",
			in:       PageDescriptor{Role: "entity_directory", Slug: "tools-index"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "entity-directory",
		},
		{
			name:     "section-index family: ParentSection ignored (a section IS itself)",
			in:       PageDescriptor{Role: "blog-index", Slug: "news", ParentSection: "ignored"},
			wantName: "news-index",
			wantURL:  "/news/index.html",
			wantType: "blog-index",
		},

		// ── Landing pages (flat, page_type preserved) ────────────────────
		{
			name:     "landing role: flat URL, page_type preserved",
			in:       PageDescriptor{Role: "landing", Slug: "free-trial"},
			wantName: "free-trial",
			wantURL:  "/free-trial.html",
			wantType: "landing",
		},
		{
			name:     "landing slug=home aliases to root index with landing page_type",
			in:       PageDescriptor{Role: "landing", Slug: "home"},
			wantName: "index",
			wantURL:  "/index.html",
			wantType: "landing", // Phase 1.5: name=index, type=landing
		},

		// ── Entity pages (nested under directory) ───────────────────────
		{
			name:     "entity-page with explicit ParentSection",
			in:       PageDescriptor{Role: "entity-page", Slug: "acme-corp", ParentSection: "clinics"},
			wantName: "acme-corp",
			wantURL:  "/clinics/acme-corp.html",
			wantType: "entity-page",
		},
		{
			name:     "entity_page (legacy snake input) normalises to entity-page output",
			in:       PageDescriptor{Role: "entity_page", Slug: "acme-corp"},
			wantName: "acme-corp",
			wantURL:  "/entities/acme-corp.html",
			wantType: "entity-page",
		},

		// ── Dash/underscore role normalisation (backward compat) ───────
		{
			name:     "blog-post (kebab canonical) round-trips unchanged",
			in:       PageDescriptor{Role: "blog-post", Slug: "first-post"},
			wantName: "first-post",
			wantURL:  "/blog/first-post.html",
			wantType: "blog-post",
		},
		{
			name:     "MIXED-CASE role with whitespace normalised",
			in:       PageDescriptor{Role: "  Blog-Post  ", Slug: "first-post"},
			wantName: "first-post",
			wantURL:  "/blog/first-post.html",
			wantType: "blog-post",
		},
		{
			name:     "blog_post (legacy snake input) normalises to blog-post output",
			in:       PageDescriptor{Role: "blog_post", Slug: "first-post"},
			wantName: "first-post",
			wantURL:  "/blog/first-post.html",
			wantType: "blog-post",
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

		// ── Blog (canonical kebab) ─────────────────────────────────────
		{
			name:     "blog-post with custom parent_section",
			in:       PageDescriptor{Role: "blog-post", Slug: "first-post", ParentSection: "guides"},
			wantName: "first-post",
			wantURL:  "/guides/first-post.html",
			wantType: "blog-post",
		},

		// ── Index / home aliasing (Phase 1.5: name=index, type=landing) ─
		{
			name:     "explicit index role: name=index, type=landing",
			in:       PageDescriptor{Role: "index", Slug: ""},
			wantName: "index",
			wantURL:  "/index.html",
			wantType: "landing",
		},
		{
			name:     "content role with home slug collapses to index name + landing type",
			in:       PageDescriptor{Role: "content", Slug: "home"},
			wantName: "index",
			wantURL:  "/index.html",
			wantType: "landing",
		},
		{
			name:     "content role with index slug collapses to index name + landing type",
			in:       PageDescriptor{Role: "content", Slug: "index"},
			wantName: "index",
			wantURL:  "/index.html",
			wantType: "landing",
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

		// ── Empty inputs ───────────────────────────────────────────────
		{
			name:     "empty slug under tool returns empty triple",
			in:       PageDescriptor{Role: "tool", Slug: ""},
			wantName: "",
			wantURL:  "",
			wantType: "",
		},
		{
			name:     "empty slug under section-index with no section",
			in:       PageDescriptor{Role: "section-index", Slug: ""},
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
			name:     "section-index ignores parent_section (a section IS itself)",
			in:       PageDescriptor{Role: "section-index", Slug: "tools", ParentSection: "ignored"},
			wantName: "tools-index",
			wantURL:  "/tools/index.html",
			wantType: "section-index",
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
