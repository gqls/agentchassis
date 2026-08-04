// FILE: platform/orchestration/actions/create_blog_posts_canonical_test.go
//
// bugs_open/080, surface: create_blog_posts. This action was the last creation
// surface still hand-rolling page identity (lowercase the LLM's name, sprintf
// "/blog/<name>.html") — the exact shape 080 was filed about on the gap-planner.
// These tests pin the convergence through datahelpers.CanonicalisePage.
//
// The important one is CleanSlugUnchanged: for role=blog-post with a clean slug
// the helper's output is byte-identical to the old hand-rolled result (measured
// fleet-wide 2026-08-03: zero existing names change), so the fix is not a
// behaviour change for the population that exists — only mistyped roles move,
// and those are the defect.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// runCreateBlogPosts drives the whole action with a one-post plan and asserts
// the (name, url, page_type) triple written to `pages` via WithArgs — if the
// action writes a different identity the INSERT no longer matches and the test
// fails with the mismatch printed.
func runCreateBlogPosts(t *testing.T, postName, postType, wantName, wantURL, wantType string) map[string]interface{} {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	pageID := uuid.New()

	// SELECT domain — return no rows; the action ignores the Scan error.
	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}))

	// Growth-budget probes are deliberately not expected: CheckPageGrowthBudget
	// ignores its own Scan errors, counts stay zero, the post is allowed, and
	// sqlmock does not consume an expectation on a non-matching call.

	// args: site_id, name, url, title, page_type, nav_order, sections, page_spec
	mock.ExpectQuery("INSERT INTO pages").
		WithArgs(
			sqlmock.AnyArg(), // site_id
			wantName,
			wantURL,
			sqlmock.AnyArg(), // title
			wantType,
			sqlmock.AnyArg(), // nav_order
			sqlmock.AnyArg(), // sections
			sqlmock.AnyArg(), // page_spec
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pageID))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Blog-index lookup for the rerender item — no rows, the arm is skipped.
	mock.ExpectQuery("SELECT id FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	out, err := CreateBlogPostsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id": siteID.String(),
			"site_plan": map[string]interface{}{
				"result": map[string]interface{}{
					"posts": []interface{}{
						map[string]interface{}{
							"name":      postName,
							"title":     "A Title",
							"purpose":   "test",
							"page_type": postType,
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateBlogPostsAction: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
	m, ok := out.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T, want map", out)
	}
	return m
}

// TestCreateBlogPosts_CleanSlugUnchanged pins the no-op property: a clean
// blog-post slug produces exactly what the hand-rolled code produced.
func TestCreateBlogPosts_CleanSlugUnchanged(t *testing.T) {
	m := runCreateBlogPosts(t,
		"why-grippers-fail", "blog-post",
		"why-grippers-fail", "/blog/why-grippers-fail.html", "blog-post")
	if got, _ := m["pages_created"].(int); got != 1 {
		t.Errorf("pages_created = %v, want 1", m["pages_created"])
	}
}

// TestCreateBlogPosts_SpacesAndCaseNormalised: the sanitisation the old code
// did by hand still happens, inside the helper.
func TestCreateBlogPosts_SpacesAndCaseNormalised(t *testing.T) {
	runCreateBlogPosts(t,
		"Why Grippers Fail", "blog-post",
		"why-grippers-fail", "/blog/why-grippers-fail.html", "blog-post")
}

// TestCreateBlogPosts_MistypedRoleLandsCanonical is the 080 regression: a post
// the LLM types as news-index must land on the canonical section-index identity
// every other surface produces, not on /blog/news.html.
func TestCreateBlogPosts_MistypedRoleLandsCanonical(t *testing.T) {
	runCreateBlogPosts(t,
		"news", "news-index",
		"news-index", "/news/index.html", "news-index")
}

// TestCreateBlogPosts_SnakeCaseTypeKebabbed: a snake_case page_type is
// normalised instead of violating chk_page_type_kebab_case at INSERT time.
func TestCreateBlogPosts_SnakeCaseTypeKebabbed(t *testing.T) {
	runCreateBlogPosts(t,
		"a-post", "blog_post",
		"a-post", "/blog/a-post.html", "blog-post")
}

// TestCreateBlogPosts_UncanonicalisableNameRefused: an identity the helper
// cannot canonicalise is skipped and counted, never hand-rolled — the silent
// fallback is how divergent rows are minted.
func TestCreateBlogPosts_UncanonicalisableNameRefused(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT domain FROM sites").
		WillReturnRows(sqlmock.NewRows([]string{"domain"}))
	// No INSERT INTO pages expected: the post must be refused before it.
	mock.ExpectQuery("SELECT id FROM pages").
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	out, err := CreateBlogPostsAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"site_id": uuid.New().String(),
			"site_plan": map[string]interface{}{
				"result": map[string]interface{}{
					"posts": []interface{}{
						// normaliseSlug strips the extension and directories;
						// ".html" reduces to an empty slug.
						map[string]interface{}{"name": ".html", "title": "x", "page_type": "blog-post"},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("CreateBlogPostsAction: %v", err)
	}
	m := out.(map[string]interface{})
	if got, _ := m["pages_created"].(int); got != 0 {
		t.Errorf("pages_created = %v, want 0", m["pages_created"])
	}
	if got, _ := m["canonicalisation_failed"].(int); got != 1 {
		t.Errorf("canonicalisation_failed = %v, want 1", m["canonicalisation_failed"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
