// FILE: platform/orchestration/actions/apply_gap_plan_new_page_test.go
//
// bugs_open/080: the platform has four surfaces that create `pages` rows and
// only three of them canonicalised. `applyNewPage` lowercased the LLM's name
// and synthesised "/<name>.html" by hand, so for the SAME logical page the
// gap-planner produced a different (name, url, page_type) triple from the
// planner, adoption and the tool path.
//
// That is not a cosmetic disagreement. `pages` upserts
// ON CONFLICT (site_id, name); a name that disagrees does not conflict, so the
// second surface INSERTS A SECOND ROW instead of updating the first.
// robot-hands.com carries both `news` (/news.html) and `news-index`
// (/news/index.html), both live, both serving 200, one orphaned from the nav.
//
// These tests pin the convergence at the surface that lacked it. The important
// one is ContentPageUnchanged: it proves the fix is not a behaviour change for
// the ordinary content page, so the blast radius is confined to the typed roles.
package actions

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// newPagePlan builds the plan shape applyNewPage consumes. Sections are
// deliberately omitted so the component-name resolver is never loaded — the
// defaults come from defaultSectionsForPage.
func newPagePlan(name, pageType string) map[string]interface{} {
	np := map[string]interface{}{"name": name}
	if pageType != "" {
		np["page_type"] = pageType
	}
	return map[string]interface{}{
		"approach":  "new_page",
		"reasoning": "the site has no news listing",
		"new_page":  np,
	}
}

// assertNewPageInsert drives applyNewPage against sqlmock and asserts the
// (name, url, page_type) triple written to `pages`.
//
// The growth-budget probes that run first are deliberately not expected:
// CheckPageGrowthBudget ignores its own Scan errors, so unmatched SELECTs leave
// the counts at zero, which is under the initial target and therefore allowed.
// sqlmock does not consume an expectation on a non-matching call, so the INSERT
// expectations below still line up.
//
// The assertion is carried by WithArgs — if applyNewPage writes a different
// name, url or page_type the INSERT no longer matches, applyNewPage returns
// "create page" and the test fails with the mismatch printed.
func assertNewPageInsert(t *testing.T, plan map[string]interface{}, wantName, wantURL, wantType string) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	pageID := uuid.New()

	// args: site_id, name, url, title, page_type, sections, nav_label,
	//       in_header, in_footer
	mock.ExpectQuery("INSERT INTO pages").
		WithArgs(
			sqlmock.AnyArg(), // site_id
			wantName,
			wantURL,
			sqlmock.AnyArg(), // title
			wantType,
			sqlmock.AnyArg(), // sections
			sqlmock.AnyArg(), // nav_label
			sqlmock.AnyArg(), // in_header
			sqlmock.AnyArg(), // in_footer
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(pageID))
	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(1, 1))

	res, err := applyNewPage(context.Background(), db, plan, uuid.New(), "example.com", nil, zap.NewNop())
	if err != nil {
		t.Fatalf("applyNewPage: %v", err)
	}

	m, ok := res.(map[string]interface{})
	if !ok {
		t.Fatalf("result is %T, want map", res)
	}
	if got := m["page_name"]; got != wantName {
		t.Errorf("reported page_name = %v, want %q", got, wantName)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestApplyNewPage_SectionIndexFamilyCanonicalises is the regression bugs_open/080
// describes: a gap-planned news listing must land on the canonical identity the
// other three creation surfaces produce, not on /news.html.
func TestApplyNewPage_SectionIndexFamilyCanonicalises(t *testing.T) {
	assertNewPageInsert(t,
		newPagePlan("news", "news-index"),
		"news-index", "/news/index.html", "news-index")
}

// TestApplyNewPage_ContentPageUnchanged is the blast-radius test. For the
// ordinary content page CanonicalisePage returns exactly what the hand-rolled
// code produced, so routing through it changes nothing for the common case.
func TestApplyNewPage_ContentPageUnchanged(t *testing.T) {
	assertNewPageInsert(t,
		newPagePlan("Restructuring Plan", ""),
		"restructuring-plan", "/restructuring-plan.html", "content")
}

// TestApplyNewPage_HomeCollapsesToIndex pins the other divergence the hand-rolled
// path had: an LLM emitting {name:"home"} used to mint a separate /home.html
// page that did not match the site root.
func TestApplyNewPage_HomeCollapsesToIndex(t *testing.T) {
	assertNewPageInsert(t,
		newPagePlan("home", "content"),
		"index", "/index.html", "landing")
}

// TestApplyNewPage_LeakedURLFragmentIsTolerated: normaliseSlug strips a path and
// extension, so an LLM that answers with a URL where a name was asked for no
// longer creates a page called "/guides/pricing.html".
func TestApplyNewPage_LeakedURLFragmentIsTolerated(t *testing.T) {
	assertNewPageInsert(t,
		newPagePlan("/guides/pricing.html", ""),
		"pricing", "/pricing.html", "content")
}

// TestApplyNewPage_UncanonicalisableNameIsRefused: an empty slug after
// normalisation used to produce a page named "" at "/.html". It is now a
// refusal, which surfaces as a failed work item rather than a poisoned row.
func TestApplyNewPage_UncanonicalisableNameIsRefused(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	_, err = applyNewPage(context.Background(), db,
		newPagePlan("/", ""), uuid.New(), "example.com", nil, zap.NewNop())
	if err == nil {
		t.Fatal("expected a refusal for a name that cannot be canonicalised, got nil")
	}
}
