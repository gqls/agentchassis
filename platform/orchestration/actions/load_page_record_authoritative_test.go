// FILE: platform/orchestration/actions/load_page_record_authoritative_test.go
//
// Tests for load_page_record's authoritative_page_id input (bugs_open/220).
//
// The load-bearing case is the PRIORITY INVERSION: the action resolves
// page_name before page_id (load_page_record_action.go's documented order), so
// for a cross-page work item — where spec.page_name is the page CONTAINING a
// defect and the page_id COLUMN is the page the item is actually about — a
// correctly-forwarded id used to lose to the container's name every time, and
// the handler rebuilt the wrong page while reporting success. These tests pin
// that an authoritative id, when supplied, wins over a present-and-resolving
// page_name — and that WITHOUT it, the name-first behaviour is exactly what it
// was, which is what makes the field a safe opt-in for every other config.

package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

func loadPageRecordParams(t *testing.T, config map[string]interface{}, collected map[string]interface{}) (ActionParams, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return ActionParams{
		Context:          context.Background(),
		Logger:           zap.NewNop(),
		DB:               db,
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: config},
	}, mock
}

func pageRecordRow(id, name string) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "title", "page_type", "sections", "url", "build_status", "nav_label", "nav_order", "content_direction"}).
		AddRow(id, name, "Title", "content", "[]", "/"+name+".html", "planned", nil, nil, nil)
}

// THE LOAD-BEARING CASE: page_name resolves ("index", the container) AND an
// authoritative id is supplied (the target) — the lookup must be BY THE ID.
// Asserted at the query arguments, which is where the wrong page is chosen.
func TestLoadPageRecord_AuthoritativeIDWinsOverPresentPageName(t *testing.T) {
	siteID := uuid.New()
	targetID := uuid.New()
	containerID := uuid.New()

	params, mock := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":               "site_record.site_id",
			"page_name":             "input_data.spec.page_name",
			"page_id":               "input_data.spec.page_id",
			"authoritative_page_id": "input_data.page_id",
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"page_id": targetID.String(), // the work item's column — the TARGET
				"spec": map[string]interface{}{
					"page_name": "index", // the CONTAINER — must NOT win
					"page_id":   containerID.String(),
				},
			},
		})

	mock.ExpectQuery(`WHERE site_id = \$1 AND id = \$2::uuid`).
		WithArgs(siteID.String(), targetID.String()).
		WillReturnRows(pageRecordRow(targetID.String(), "directory-index"))

	out, err := LoadPageRecordAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Fatalf("the lookup did not go by the authoritative id — with a resolving page_name present "+
			"this is bugs_open/220's container-wins defect: %v", merr)
	}
	m := out.(map[string]interface{})
	if m["found"] != true || m["id"] != targetID.String() {
		t.Fatalf("expected the TARGET page record, got %v", m)
	}
}

// Without authoritative_page_id in the config, the name-first behaviour is
// byte-for-byte today's — that unchanged default is what licenses shipping the
// field without an RFC (it is opt-in, reachable by nothing until a config names
// it).
func TestLoadPageRecord_WithoutAuthoritativeIDNameStillWins(t *testing.T) {
	siteID := uuid.New()
	containerID := uuid.New()

	params, mock := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":   "site_record.site_id",
			"page_name": "input_data.spec.page_name",
			"page_id":   "input_data.spec.page_id",
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"spec": map[string]interface{}{
					"page_name": "index",
					"page_id":   containerID.String(),
				},
			},
		})

	mock.ExpectQuery(`WHERE site_id = \$1 AND name = \$2`).
		WithArgs(siteID.String(), "index").
		WillReturnRows(pageRecordRow(containerID.String(), "index"))

	if _, err := LoadPageRecordAction(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Fatalf("name-first lookup changed without the opt-in field — that widens the fix beyond its "+
			"licence: %v", merr)
	}
}

// A VALID uuid that matches no row must be FATAL, not the {found:false} soft
// miss the name flow returns (council round 1, bug_historian, HIGH). The id
// came from the work item's page_id column, so no-row means the target was
// deleted or is foreign to the site — a soft miss would route the saga through
// the success-labelled complete_error path, the exact silent no-op shape this
// input exists to close. This test pins that it neither errors-as-not-found
// NOR silently falls back to the (resolving) container name.
func TestLoadPageRecord_AuthoritativeIDMatchingNoRowIsFatal(t *testing.T) {
	siteID := uuid.New()
	targetID := uuid.New()

	params, mock := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":               "site_record.site_id",
			"page_name":             "input_data.spec.page_name",
			"authoritative_page_id": "input_data.page_id",
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"page_id": targetID.String(),
				"spec":    map[string]interface{}{"page_name": "index"},
			},
		})

	mock.ExpectQuery(`WHERE site_id = \$1 AND id = \$2::uuid`).
		WithArgs(siteID.String(), targetID.String()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "title", "page_type", "sections", "url", "build_status", "nav_label", "nav_order", "content_direction"}))

	_, err := LoadPageRecordAction(context.Background(), params)
	if err == nil {
		t.Fatal("expected a fatal error for an authoritative id matching no row — a soft {found:false} " +
			"routes the saga through the success-labelled complete_error path, and a silent name fallback " +
			"loads the container: both are the defect")
	}
	if !strings.Contains(err.Error(), "authoritative_page_id") {
		t.Fatalf("error should name the input; got: %v", err)
	}
	if merr := mock.ExpectationsWereMet(); merr != nil {
		t.Fatalf("the zero-row id lookup must be the one and only query — a second (name) query would be "+
			"the silent fallback this test forbids: %v", merr)
	}
}

// A malformed authoritative value is a misconfigured path, not a data state:
// falling through to the name would silently load a different page, which is
// the exact defect the field exists to close. Loud error, matching site_id's
// handling.
func TestLoadPageRecord_MalformedAuthoritativeIDErrors(t *testing.T) {
	siteID := uuid.New()

	params, _ := loadPageRecordParams(t,
		map[string]interface{}{
			"site_id":               "site_record.site_id",
			"page_name":             "input_data.spec.page_name",
			"authoritative_page_id": "input_data.page_id",
		},
		map[string]interface{}{
			"site_record": map[string]interface{}{"site_id": siteID.String()},
			"input_data": map[string]interface{}{
				"page_id": "not-a-uuid",
				"spec":    map[string]interface{}{"page_name": "index"},
			},
		})

	_, err := LoadPageRecordAction(context.Background(), params)
	if err == nil {
		t.Fatal("expected an error for a malformed authoritative_page_id — silent fallback to the name is the bug")
	}
	if !strings.Contains(err.Error(), "authoritative_page_id") {
		t.Fatalf("error should name the misconfigured input; got: %v", err)
	}
}
