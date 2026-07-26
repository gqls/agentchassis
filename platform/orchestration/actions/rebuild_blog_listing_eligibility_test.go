// FILE: platform/orchestration/actions/rebuild_blog_listing_eligibility_test.go
//
// Tests for the build-state floor on the blog-listing derivation
// (bugs_open/052). Sibling of queryresolve/page_list_eligibility_test.go, and
// deliberately the same shape: DB-free assertions on the SQL the action
// splices together, so the fix cannot silently regress. The SQL semantics are
// verified against live data in the bug file; what is pinned here is the
// fragment's SHAPE.
//
// Why this file exists at all: queryresolve gained the shared floor in
// v1.0.1146 and this action kept its own hand-written predicate, so the fleet
// carried a fixed listing path and a broken one for five days. A shared
// constant does not propagate itself.

package actions

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
)

// The regression that WAS bugs_open/052's residual: the blog listing selected
// posts on `build_status IN ('deployed','needs_rebuild')` alone. That admits
// pages which are needs_rebuild but were never deployed — 4 fleet pages, all
// live 404s — because build_status cannot answer "will this URL serve".
func TestBlogPostsQueryHasFetchabilityFloor(t *testing.T) {
	if !strings.Contains(blogPostsQuery, "deployed_at IS NOT NULL") {
		t.Errorf("blogPostsQuery lacks a deployed_at floor — a never-deployed post would be advertised (bugs_open/052):\n%s", blogPostsQuery)
	}
	if strings.Contains(blogPostsQuery, "build_status IN ('deployed', 'needs_rebuild')") {
		t.Errorf("blogPostsQuery still carries the naive build_status predicate; needs_rebuild does not mean 'still serves':\n%s", blogPostsQuery)
	}
}

// The archived-containment defeat, which was the worse half: the query had NO
// status filter, so a page deliberately archived to take it out of listings —
// the fleet's containment route for a dead page — was listed anyway. Every
// other page-set-derived listing filters status; this one did not.
func TestBlogPostsQueryRespectsArchivedPages(t *testing.T) {
	if !strings.Contains(blogPostsQuery, "p.status IN ('active', 'deployed')") {
		t.Errorf("blogPostsQuery must filter pages.status, else archiving a dead post does not delist it (bugs_open/052):\n%s", blogPostsQuery)
	}
}

// The floor must be the SHARED constant, not a copy of its text. This action
// and queryresolve's `blog_posts` source both derive the article set for the
// same blog page; if they carry separate literals they drift, which is the
// class bugs_closed/023 documents. Substring containment is the strongest
// DB-free proof available that the constant itself was spliced in.
func TestBlogPostsQueryUsesTheSharedArticleContract(t *testing.T) {
	if !strings.Contains(blogPostsQuery, queryresolve.ListedPageEligibilitySQL) {
		t.Errorf("blogPostsQuery must splice in queryresolve.ListedPageEligibilitySQL verbatim so the two derivations cannot drift.\nwant fragment:\n%s\ngot query:\n%s",
			queryresolve.ListedPageEligibilitySQL, blogPostsQuery)
	}
}

// The shared constants carry a `p` alias contract. A query that aliased pages
// differently would splice in fragments referencing an alias it does not
// define and fail at runtime, not compile time — so pin the alias here.
func TestBlogPostsQueryHonoursTheAliasContract(t *testing.T) {
	if !strings.Contains(blogPostsQuery, "FROM pages p") {
		t.Errorf("blogPostsQuery must alias pages as `p` — the shared eligibility fragments reference p.*:\n%s", blogPostsQuery)
	}
}
