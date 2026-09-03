package datahelpers

import "testing"

// ValidateRoles' rule 6: the directory a leaf page canonicalises into, recovered
// from the URL the write path is about to discard (bugs_open/463).
//
// The defect these lock down is not visible at ValidateRoles at all — it appears
// two calls later, at CanonicalisePage, whose blog-post arm defaults the
// directory to "blog" when ParentSection is empty. So most of these tests assert
// the FINAL URL rather than the intermediate field: a test that only checked
// ParentSection would pass while the page still landed in the wrong place.

// canonicalURLFor runs the real pair the two write surfaces run, in the order
// they run it, so these tests measure what actually reaches site_plan_pages.
func canonicalURLFor(t *testing.T, p LLMPlannedPage) (name, url, pageType string) {
	t.Helper()
	v := ValidateRoles([]LLMPlannedPage{p})[0]
	slug := v.Slug
	if slug == "" {
		slug = v.Name
	}
	return CanonicalisePage(PageDescriptor{
		Role:          v.Role,
		Slug:          slug,
		ParentSection: v.ParentSection,
	})
}

// The filed case, and the reason fixing Pass C alone would not have fixed the
// bug: a blog-post the planner placed under /articles/ was written to /blog/.
func TestValidateRoles_LeafPageKeepsThePlannedSection(t *testing.T) {
	cases := []struct {
		name       string
		page       LLMPlannedPage
		wantParent string
		wantURL    string
		why        string
	}{
		{
			name: "blog-post under a custom section keeps it",
			page: LLMPlannedPage{
				Name: "article-sign-off-problem", Role: "blog-post",
				URL: "/articles/the-sign-off-problem.html",
			},
			wantParent: "articles",
			wantURL:    "/articles/article-sign-off-problem.html",
			why:        "bugs_open/463: without the derivation this becomes /blog/... and the hub resolves zero children",
		},
		{
			name:       "blog-post under the default section is unchanged",
			page:       LLMPlannedPage{Name: "welcome", Role: "blog-post", URL: "/blog/welcome.html"},
			wantParent: "blog",
			wantURL:    "/blog/welcome.html",
			why:        "the overwhelming majority of live rows; derivation must be a no-op here",
		},
		{
			name:       "blog-post with a FLAT url gets no parent and takes the role default",
			page:       LLMPlannedPage{Name: "welcome", Role: "blog-post", URL: "/welcome.html"},
			wantParent: "",
			wantURL:    "/blog/welcome.html",
			why:        "a root-level url names no section, so CanonicalisePage's own default must still apply",
		},
		{
			name:       "blog-post with NO url is untouched",
			page:       LLMPlannedPage{Name: "welcome", Role: "blog-post"},
			wantParent: "",
			wantURL:    "/blog/welcome.html",
			why:        "nothing to derive from; behaviour identical to before the fix",
		},
		{
			name:       "entity-page keeps its directory instead of falling back to /entities/",
			page:       LLMPlannedPage{Name: "oakfield-clinic", Role: "entity-page", URL: "/clinics/oakfield-clinic.html"},
			wantParent: "clinics",
			wantURL:    "/clinics/oakfield-clinic.html",
			why:        "same defect, different role default",
		},
		{
			name:       "a hyphenated section survives normaliseSlug",
			page:       LLMPlannedPage{Name: "acme-rollout", Role: "blog-post", URL: "/case-studies/acme-rollout.html"},
			wantParent: "case-studies",
			wantURL:    "/case-studies/acme-rollout.html",
			why:        "normaliseSlug truncates at the LAST slash, so a multi-word section had to be checked, not assumed",
		},
		{
			name: "an explicit parent_section always wins over the url",
			page: LLMPlannedPage{
				Name: "acme-rollout", Role: "blog-post",
				URL: "/case-studies/acme-rollout.html", ParentSection: "stories",
			},
			wantParent: "stories",
			wantURL:    "/stories/acme-rollout.html",
			why:        "the derivation fills a gap; it must never overrule a stated value",
		},
		{
			name:       "a NON-leaf role is not derived",
			page:       LLMPlannedPage{Name: "about", Role: "content", URL: "/company/about.html"},
			wantParent: "",
			wantURL:    "/about.html",
			why:        "CanonicalisePage ignores ParentSection for content; deriving it would be a claim nothing reads",
		},
		{
			name:       "a section index is not derived — it IS its section",
			page:       LLMPlannedPage{Name: "articles-index", Role: "section-index", URL: "/articles/index.html"},
			wantParent: "",
			wantURL:    "/articles/index.html",
			why:        "CanonicalisePage documents that ParentSection is deliberately not honoured for a hub",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := ValidateRoles([]LLMPlannedPage{tc.page})[0]
			if v.ParentSection != tc.wantParent {
				t.Errorf("ParentSection = %q, want %q\n  why: %s", v.ParentSection, tc.wantParent, tc.why)
			}
			_, url, _ := canonicalURLFor(t, tc.page)
			if url != tc.wantURL {
				t.Errorf("canonical url = %q, want %q\n  why: %s", url, tc.wantURL, tc.why)
			}
		})
	}
}

// THE GUARD, PROVEN BY MUTATION RATHER THAN BY AN ABSENT EXPECTATION.
//
// "An entry the reconciler paired with a realised page is never derived" is a
// rule about NOT acting, and a test that passes because the branch was never
// reached looks exactly like one that passes because the guard held. So this
// test asserts BOTH sides of the same input: with the marker the URL must stay
// at the role default, and with the marker removed — the mutation — the
// derivation must fire. If the guard were deleted, the first half fails; if the
// derivation were deleted, the second half fails.
//
// Why the guard exists: a same-name stamped entry (Pass B2) carries
// identity_authority and the REALISED url but deliberately no parent_section.
// Deriving from that url would honour the realised identity while the site's
// honour_realised_identity flag is OFF — the behaviour change
// stampSameNameRealisedIdentity refuses to make (bugs_open/215).
func TestValidateRoles_RealisedEntryIsNotDerived(t *testing.T) {
	// The url must be one rule 5 does NOT already rescue. /guides/, /tools/ and
	// /games/ are retyped by nestedRoleFromURL before rule 6 is reached, so a
	// fixture under those directories would test rule 5, not this guard — I used
	// one at first and it silently measured the wrong thing.
	realised := LLMPlannedPage{
		Name: "welcome", Role: "blog-post",
		URL:              "/articles/welcome.html", // the url the reconciler stamped
		RealisedIdentity: true,
	}

	if v := ValidateRoles([]LLMPlannedPage{realised})[0]; v.ParentSection != "" {
		t.Errorf("ParentSection = %q for a realised-identity entry, want empty — "+
			"deriving here honours the stored identity with honour_realised_identity OFF", v.ParentSection)
	}
	if _, url, _ := canonicalURLFor(t, realised); url != "/blog/welcome.html" {
		t.Errorf("canonical url = %q, want /blog/welcome.html (unchanged from before the fix)", url)
	}

	// The mutation: same page, marker cleared. The derivation MUST fire, or the
	// assertions above were vacuous.
	mutated := realised
	mutated.RealisedIdentity = false
	if v := ValidateRoles([]LLMPlannedPage{mutated})[0]; v.ParentSection != "articles" {
		t.Fatalf("with the marker cleared ParentSection = %q, want \"articles\" — "+
			"the guard cannot be shown to hold if the branch never runs", v.ParentSection)
	}
	if _, url, _ := canonicalURLFor(t, mutated); url != "/articles/welcome.html" {
		t.Fatalf("with the marker cleared url = %q, want /articles/welcome.html", url)
	}
}

// The three directories rule 5 already rescues must be BYTE-IDENTICAL before and
// after this change: nestedRoleFromURL retypes the page first, and the retyped
// role's default directory is the same one rule 6 would derive. Asserted rather
// than assumed — /guides/, /tools/ and /games/ carry 280+ live blog-post pages
// between them, so a shift here would move real URLs.
func TestValidateRoles_Rule5DirectoriesAreUnchanged(t *testing.T) {
	cases := []struct{ url, wantRole, wantURL string }{
		{"/guides/welcome.html", "guide", "/guides/welcome/index.html"},
		{"/tools/widget.html", "tool", "/tools/widget/index.html"},
		{"/games/puzzle.html", "game", "/games/puzzle/index.html"},
	}
	for _, c := range cases {
		p := LLMPlannedPage{Name: "welcome", Role: "blog-post", URL: c.url}
		if c.url != "/guides/welcome.html" {
			p.Name = c.wantRole + "-thing"
		}
		v := ValidateRoles([]LLMPlannedPage{p})[0]
		if v.Role != c.wantRole {
			t.Errorf("%s: role = %q, want %q (rule 5 owns this, not rule 6)", c.url, v.Role, c.wantRole)
		}
	}
	// The load-bearing half: the derived parent equals the role default, so the
	// URL cannot move.
	p := LLMPlannedPage{Name: "welcome", Role: "blog-post", URL: "/guides/welcome.html"}
	if _, url, _ := canonicalURLFor(t, p); url != "/guides/welcome/index.html" {
		t.Errorf("url = %q, want /guides/welcome/index.html — rule 5's directories must not move", url)
	}
}

// The derivation must change no ROLE. declaredParents is built before the loop
// from the INPUT ParentSection only, so a derived value must not be able to
// reach rule 3 and retype a page to section-index.
func TestValidateRoles_DerivationChangesNoRole(t *testing.T) {
	pages := []LLMPlannedPage{
		{Name: "articles-index", Role: "section-index", URL: "/articles/index.html"},
		{Name: "article-one", Role: "blog-post", URL: "/articles/one.html"},
		{Name: "guide-two", Role: "guide", URL: "/guides/two/index.html"},
		{Name: "about", Role: "content", URL: "/about.html"},
		{Name: "pricing", Role: "landing", URL: "/pricing.html"},
	}
	wantRoles := []string{"section-index", "blog-post", "guide", "content", "landing"}

	for i, v := range ValidateRoles(pages) {
		if v.Role != wantRoles[i] {
			t.Errorf("page %q role = %q, want %q — the ParentSection derivation must not move a role",
				v.Name, v.Role, wantRoles[i])
		}
		if v.CorrectedFromRole != "" {
			t.Errorf("page %q was retyped from %q; no role correction is expected here", v.Name, v.CorrectedFromRole)
		}
	}
}

func TestParentSectionFromURL(t *testing.T) {
	cases := []struct{ url, want string }{
		{"/articles/x.html", "articles"},
		{"/articles/x/index.html", "articles"},
		{"/case-studies/acme.html", "case-studies"},
		{"/x.html", ""},
		{"/", ""},
		{"", ""},
		{"/articles/index.html", "articles"},
	}
	for _, c := range cases {
		if got := ParentSectionFromURL(c.url); got != c.want {
			t.Errorf("ParentSectionFromURL(%q) = %q, want %q", c.url, got, c.want)
		}
	}
}

func TestPlanPageCarriesRealisedIdentity(t *testing.T) {
	cases := []struct {
		name string
		page map[string]interface{}
		want bool
	}{
		{"a snapped or unioned entry carries both markers",
			map[string]interface{}{"identity_authority": "realised", "from_realised": true}, true},
		{"a same-name stamp carries only identity_authority",
			map[string]interface{}{"identity_authority": "realised"}, true},
		{"from_realised alone is enough",
			map[string]interface{}{"from_realised": true}, true},
		{"a plain LLM page carries neither",
			map[string]interface{}{"name": "x"}, false},
		{"an empty authority string is not a marker",
			map[string]interface{}{"identity_authority": ""}, false},
		{"from_realised false is not a marker",
			map[string]interface{}{"from_realised": false}, false},
	}
	for _, c := range cases {
		if got := PlanPageCarriesRealisedIdentity(c.page); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
