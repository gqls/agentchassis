// FILE: platform/orchestration/actions/load_page_record_refuse_owned_test.go
//
// Tests for load_page_record's refuse_owned_page opt-in (bugs_open/301).
//
// The load-bearing property is the ORDERING contract: with the key set, an
// owned page is refused by THIS action — before any writer is spawned — with an
// error that LEADS with ownedPageSkipReasonPrefix, because that exact marker is
// what update_work_item_status' owned_page_refusal_status (migration 480)
// matches in __step_error.message to stamp the item wont_fix rather than
// failed. And with the key ABSENT — every caller except page-build-handler,
// including tool-recreation-handler whose whole job is owned pages — the action
// must not so much as read rebuild_policy, which is what makes the field a safe
// opt-in (owner ruling 2026-08-02 §2).

package actions

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

// THE LOAD-BEARING CASE: key set, page owned — refused at load, review row
// filed, error carries the Tier 1 marker. Asserted at the mock's expectation
// set: the policy read and the review INSERT must BOTH happen, and nothing else.
func TestLoadPageRecord_RefuseOwnedPageRefusesBeforeTheWriter(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()

	params, mock := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":           "site_record.site_id",
			"page_name":         "input_data.spec.page_name",
			"refuse_owned_page": true,
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"spec": map[string]interface{}{"page_name": "tool-quiz"},
			},
		})

	mock.ExpectQuery(`WHERE site_id = \$1 AND name = \$2`).
		WithArgs(siteID.String(), "tool-quiz").
		WillReturnRows(pageRecordRow(pageID.String(), "tool-quiz"))
	mock.ExpectQuery(`SELECT COALESCE\(pages.rebuild_policy`).
		WithArgs(pageID.String()).
		WillReturnRows(policyRows("owned", false))
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := LoadPageRecordAction(context.Background(), params)
	if err == nil {
		t.Fatal("expected a refusal error for an owned page with refuse_owned_page set — " +
			"without it the workflow proceeds to spawn the content writer, which is bugs_open/301")
	}
	if !strings.HasPrefix(err.Error(), ownedPageSkipReasonPrefix) {
		t.Fatalf("the refusal must LEAD with %q — update_work_item_status' owned_page_refusal_status "+
			"matches that marker in __step_error.message to stamp wont_fix instead of failed "+
			"(migration 480); got: %v", ownedPageSkipReasonPrefix, err)
	}
	if !strings.Contains(err.Error(), "tool-quiz") {
		t.Fatalf("the refusal should name the page; got: %v", err)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Fatalf("the refusal must both read the policy and file the owned_page_review row: %v", merr)
	}
}

// Key set, page generic: the build proceeds exactly as before. The policy read
// happens; no review row is filed; no error.
func TestLoadPageRecord_RefuseOwnedPageGenericProceeds(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()

	params, mock := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":           "site_record.site_id",
			"page_name":         "input_data.spec.page_name",
			"refuse_owned_page": true,
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"spec": map[string]interface{}{"page_name": "about"},
			},
		})

	mock.ExpectQuery(`WHERE site_id = \$1 AND name = \$2`).
		WithArgs(siteID.String(), "about").
		WillReturnRows(pageRecordRow(pageID.String(), "about"))
	mock.ExpectQuery(`SELECT COALESCE\(pages.rebuild_policy`).
		WithArgs(pageID.String()).
		WillReturnRows(policyRows("generic", false))

	out, err := LoadPageRecordAction(context.Background(), params)
	if err != nil {
		t.Fatalf("a generic page must load normally with the key set: %v", err)
	}
	m := out.(map[string]interface{})
	if m["found"] != true || m["name"] != "about" {
		t.Fatalf("expected the loaded page record, got %v", m)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Fatalf("unexpected DB traffic on the generic path (an INSERT here would be a review "+
			"row filed for a page that was not refused): %v", merr)
	}
}

// THE UNCHANGED DEFAULT, which is what licenses shipping this without an RFC:
// with the key absent, the action must not even READ rebuild_policy. The mock
// expects only the page lookup — a policy query or an INSERT would fail
// ExpectationsWereMet. tool-recreation-handler (the tool pipeline, which
// legitimately builds owned pages) is this case in production.
func TestLoadPageRecord_WithoutKeyOwnedPageIsUntouched(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()

	params, mock := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":   "site_record.site_id",
			"page_name": "input_data.spec.page_name",
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"spec": map[string]interface{}{"page_name": "tool-quiz"},
			},
		})

	mock.ExpectQuery(`WHERE site_id = \$1 AND name = \$2`).
		WithArgs(siteID.String(), "tool-quiz").
		WillReturnRows(pageRecordRow(pageID.String(), "tool-quiz"))

	out, err := LoadPageRecordAction(context.Background(), params)
	if err != nil {
		t.Fatalf("without the key the load must behave byte-for-byte as before: %v", err)
	}
	if m := out.(map[string]interface{}); m["found"] != true {
		t.Fatalf("expected the page record, got %v", m)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Fatalf("with refuse_owned_page absent the action read more than the page row — "+
			"the opt-in leaked into the default path: %v", merr)
	}
}

// Fail-open: the policy read erroring must NOT block the load (the save-path
// guard is still downstream as the backstop; here fail-open costs at most one
// wasted chain, never a clobber). No review row, no error, record returned.
func TestLoadPageRecord_RefuseOwnedPageFailsOpenOnUnreadablePolicy(t *testing.T) {
	siteID := uuid.New()
	pageID := uuid.New()

	params, mock := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":           "site_record.site_id",
			"page_name":         "input_data.spec.page_name",
			"refuse_owned_page": true,
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"spec": map[string]interface{}{"page_name": "tool-quiz"},
			},
		})

	mock.ExpectQuery(`WHERE site_id = \$1 AND name = \$2`).
		WithArgs(siteID.String(), "tool-quiz").
		WillReturnRows(pageRecordRow(pageID.String(), "tool-quiz"))
	mock.ExpectQuery(`SELECT COALESCE\(pages.rebuild_policy`).
		WithArgs(pageID.String()).
		WillReturnError(fmt.Errorf("relation scan failed"))

	out, err := LoadPageRecordAction(context.Background(), params)
	if err != nil {
		t.Fatalf("an unreadable policy must stand the early guard down, not fail the load: %v", err)
	}
	if m := out.(map[string]interface{}); m["found"] != true {
		t.Fatalf("expected the page record on fail-open, got %v", m)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Fatalf("fail-open must not file a review row: %v", merr)
	}
}
