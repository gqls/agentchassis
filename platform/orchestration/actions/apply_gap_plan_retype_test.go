// FILE: platform/orchestration/actions/apply_gap_plan_retype_test.go
//
// bugs_open/015: MissingNewsPageCheck (candidate 3, live v1.0.1144) tells
// content-gap-planner to RE-TYPE a stranded page rather than mint a
// duplicate — but until applyRetypeExisting existed, no approach in this
// executor could change pages.page_type, so the advice dead-ended in the
// unknown-approach branch with applied=false.
//
// The branch is fail-closed: the LLM plan only NAMES the page; the
// authorising facts (retype_candidates, target page_type) come from the
// original work item's spec, written deterministically by the discovery
// check. These tests pin both the happy path and the refusals.
package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const retypeCandidateID = "11111111-1111-1111-1111-111111111111"

// itemSpecJSON is the shape check_news_feed.go writes: page_type is the
// routing key, retype_candidates the structural stranded-set.
const retypeItemSpecJSON = `{
	"check": "missing_news_page",
	"page_name": "news",
	"page_type": "news-index",
	"approach": "retype_existing",
	"retype_candidates": [
		{"id": "` + retypeCandidateID + `", "name": "noticias-index", "page_type": "section-index", "build_status": "planned"}
	]
}`

func retypePlan(pageName string) map[string]interface{} {
	return map[string]interface{}{
		"approach":  "retype_existing",
		"reasoning": "the stranded Spanish listing IS the news page",
		"retype_existing": map[string]interface{}{
			"page_name": pageName,
		},
	}
}

func TestApplyRetypeExisting_HappyPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	itemID := uuid.New()

	mock.ExpectQuery("SELECT spec FROM site_work_items").
		WithArgs(itemID, siteID).
		WillReturnRows(sqlmock.NewRows([]string{"spec"}).AddRow(retypeItemSpecJSON))

	// The UPDATE must key on the candidate's id from the item spec — not a
	// name the LLM chose — and must set the spec's page_type.
	mock.ExpectExec("UPDATE pages").
		WithArgs(uuid.MustParse(retypeCandidateID), "news-index", sqlmock.AnyArg(), siteID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// markOriginalComplete
	mock.ExpectExec("UPDATE site_work_items").
		WillReturnResult(sqlmock.NewResult(0, 1))

	got, err := applyRetypeExisting(context.Background(), db, retypePlan("noticias-index"), siteID, &itemID, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := got.(map[string]interface{})
	if res["applied"] != true {
		t.Fatalf("applied=%v want true (reason=%v)", res["applied"], res["reason"])
	}
	if res["from_type"] != "section-index" || res["to_type"] != "news-index" {
		t.Errorf("re-type recorded %v -> %v, want section-index -> news-index",
			res["from_type"], res["to_type"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// A page outside retype_candidates must be refused — the LLM must not be
// able to re-type a live, working page it happens to know the name of.
func TestApplyRetypeExisting_RefusesPageOutsideCandidates(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	itemID := uuid.New()

	mock.ExpectQuery("SELECT spec FROM site_work_items").
		WithArgs(itemID, siteID).
		WillReturnRows(sqlmock.NewRows([]string{"spec"}).AddRow(retypeItemSpecJSON))

	got, err := applyRetypeExisting(context.Background(), db, retypePlan("about"), siteID, &itemID, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := got.(map[string]interface{})
	if res["applied"] != false {
		t.Fatalf("applied=%v want false", res["applied"])
	}
	reason, _ := res["reason"].(string)
	if !strings.Contains(reason, "not in retype_candidates") {
		t.Errorf("reason %q should name the candidate-set refusal", reason)
	}
	// No UPDATE pages, no work item, no completion — ExpectationsWereMet
	// fails if anything beyond the spec read was executed.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("refusal must not touch the database further: %v", err)
	}
}

// An item spec without page_type carries nothing deterministic to re-type
// to; the LLM's own choice of type must never be used instead.
func TestApplyRetypeExisting_RefusesWithoutSpecPageType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	itemID := uuid.New()

	mock.ExpectQuery("SELECT spec FROM site_work_items").
		WithArgs(itemID, siteID).
		WillReturnRows(sqlmock.NewRows([]string{"spec"}).AddRow(
			`{"retype_candidates": [{"id": "` + retypeCandidateID + `", "name": "noticias-index"}]}`))

	got, err := applyRetypeExisting(context.Background(), db, retypePlan("noticias-index"), siteID, &itemID, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := got.(map[string]interface{})
	if res["applied"] != false {
		t.Fatalf("applied=%v want false", res["applied"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("refusal must not touch the database further: %v", err)
	}
}

// Without an original work item there is nothing that authorises any
// re-type at all.
func TestApplyRetypeExisting_RefusesWithoutOriginalItem(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	got, err := applyRetypeExisting(context.Background(), db, retypePlan("noticias-index"), uuid.New(), nil, zap.NewNop())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	res := got.(map[string]interface{})
	if res["applied"] != false {
		t.Fatalf("applied=%v want false", res["applied"])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("refusal must not touch the database: %v", err)
	}
}

// defaultSectionsForPage must key the news archetype on page_type, not the
// page name — the name is localised ("noticias"), the type is not. Without
// this a news-index page whose plan omits sections got a generic-text-block
// instead of the news-listing component, orphaning it from the news
// pipeline all over again.
func TestDefaultSectionsForPage_NewsIndexByType(t *testing.T) {
	got := defaultSectionsForPage("noticias", "news-index")
	want := []string{"hero", "news-listing", "call-to-action"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v want %v", got, want)
		}
	}

	// Name-based archetypes are unaffected for untyped pages.
	faq := defaultSectionsForPage("faq", "content")
	if faq[1] != "faq" {
		t.Errorf("faq archetype regressed: %v", faq)
	}
}
