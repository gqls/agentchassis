// FILE: platform/orchestration/actions/tool_backend_provision_test.go
//
// The deploy-time requires-backend gate and its provision item. The
// load-bearing cases are the two refusals (a widget shipped against no
// backend, and a provision item that would be born unsatisfiable —
// bugs_open/177's class), proven on the extracted decision function rather
// than asserted by comment, and the recurrence flag, proven by EFFECT the same
// way tool_content_item_test.go proves it: the two-strike probe's branding
// would change the insert's arguments, so a pinned summary and status catch a
// dropped flag where a probe-count assertion cannot (see the LANDMINES entry
// on vacuous not-issued assertions against insertWorkItem).
package actions

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestToolRequiresBackend(t *testing.T) {
	cases := []struct {
		name string
		tags string
		want bool
	}{
		{"tagged", `["tool", "chat", "requires-backend"]`, true},
		{"untagged", `["tool", "calculator"]`, false},
		{"empty array", `[]`, false},
		{"empty string (SQL null scans to this)", ``, false},
		{"json null", `null`, false},
		// A malformed column must not disarm the gate: the unsafe side is
		// "deployed against no backend", so parse failure falls back to the
		// containment the `?` operator would have matched.
		{"malformed but tagged", `["requires-backend", `, true},
		{"malformed and untagged", `["tool", `, false},
		// The quoted-element check cannot be fooled by a superstring tag.
		{"superstring tag is not the tag", `["not-requires-backend-really"]`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := toolRequiresBackend(c.tags); got != c.want {
				t.Errorf("toolRequiresBackend(%q) = %v, want %v", c.tags, got, c.want)
			}
		})
	}
}

func TestBackendEligibilityRefusal(t *testing.T) {
	// Not capable: refused whatever the facts say — 406's predicate is the
	// authority and the error must point the reader at it.
	err := backendEligibilityRefusal(backendEligibility{Capable: false, FactsCount: 15}, "chat-input-box", "static.example")
	if err == nil {
		t.Fatal("no-backend-capability site was not refused")
	}
	if !regexp.MustCompile(`406`).MatchString(err.Error()) {
		t.Errorf("capability refusal does not name the suggester gate it mirrors: %v", err)
	}

	// Capable but zero facts: refused — the relay 404s and the backend
	// refuses to start, so the provision item would be unsatisfiable at birth.
	err = backendEligibilityRefusal(backendEligibility{Capable: true, FactsCount: 0}, "chat-input-box", "noted.co.uk")
	if err == nil {
		t.Fatal("zero-facts site was not refused")
	}
	if !regexp.MustCompile(`evidence_base`).MatchString(err.Error()) {
		t.Errorf("facts refusal does not name the missing half: %v", err)
	}

	// Capable with facts: proceeds.
	if err := backendEligibilityRefusal(backendEligibility{Capable: true, FactsCount: 13}, "chat-input-box", "relojistas.com"); err != nil {
		t.Errorf("eligible site refused: %v", err)
	}
}

func TestLoadBackendEligibility_ScansTheThreeFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectQuery("SELECT").
		WithArgs(siteID).
		WillReturnRows(sqlmock.NewRows([]string{"capable", "facts", "contact"}).
			AddRow(true, 13, `{"id": "contact", "claim": "Email us at x@example.com"}`))

	e, err := loadBackendEligibility(context.Background(), db, siteID)
	if err != nil {
		t.Fatalf("loadBackendEligibility: %v", err)
	}
	if !e.Capable || e.FactsCount != 13 || e.ContactFact == "" {
		t.Errorf("got %+v, want capable, 13 facts, contact fact present", e)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func backendProvisionTestRequest() backendProvisionRequest {
	return backendProvisionRequest{
		siteID:       uuid.New(),
		domain:       "noted.co.uk",
		pageID:       uuid.New(),
		pageURL:      "/tools/chat-input-box/index.html",
		toolFunction: "chat-input-box",
		displayName:  "Site Chat",
		forkID:       uuid.New(),
		eligibility:  backendEligibility{Capable: true, FactsCount: 13},
	}
}

// expectBackendProvisionInsert pins the fields that make the row routable to a
// person and dedupable: the new item_type, no handler agent, and the
// needs_human_review status — the live idiom for items whose consumer is an
// operator. Sixteen arguments; parent_item_id's $17 is only appended by
// callers that set one, and this caller does not.
func expectBackendProvisionInsert(mock sqlmock.Sqlmock, rowsAffected int64) {
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"backend_provision", // $4 item_type
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			"",                   // $11 handler_agent — no automated consumer exists
			"needs_human_review", // $12 status
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, rowsAffected))
	mock.ExpectCommit()
}

func TestRaiseBackendProvisionItem_Raised(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectBackendProvisionInsert(mock, 1)

	got := raiseBackendProvisionItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), backendProvisionTestRequest())

	if got != "raised" {
		t.Errorf("disposition = %q, want %q", got, "raised")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// An OPEN item already holds the key: idx_swi_dedup takes the insert to
// nothing. "Already asked for" and "asked just now" are different facts and
// the caller's output map must carry which one happened.
func TestRaiseBackendProvisionItem_OpenItemHoldsKey_Deduped(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	expectBackendProvisionInsert(mock, 0)

	got := raiseBackendProvisionItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), backendProvisionTestRequest())

	if got != "deduped_open_item" {
		t.Errorf("disposition = %q, want %q", got, "deduped_open_item")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The recurrence flag, by effect. A completed provision is a SUCCESS; a later
// re-ask (box rebuilt, fork re-created) must arrive undamaged. If the flag
// were dropped, the two-strike probe below would be issued, answered with two
// terminal rows, and the item would arrive branded `unresolved` with a
// rewritten summary — so the pinned summary and status here would not match,
// the insert would fail, and the disposition would be insert_failed. The probe
// expectation is meant to go unfulfilled; that is the assertion, which is why
// ExpectationsWereMet is deliberately not called.
func TestRaiseBackendProvisionItem_RecurrenceExpected_SurvivesTerminalPredecessors(t *testing.T) {
	db, mock := newInsertMock(t)
	defer db.Close()

	mock.MatchExpectationsInOrder(false)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*),")).
		WillReturnRows(sqlmock.NewRows([]string{"count", "age"}).AddRow(2, 5.0))
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO site_work_items").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"Provision the Site Chat backend for noted.co.uk — the deployed widget is fail-closed until this is done", // $6 summary, unbranded
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"needs_human_review", // $12 status, not `unresolved`
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	got := raiseBackendProvisionItem(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, zap.NewNop(), backendProvisionTestRequest())

	if got != "raised" {
		t.Errorf("disposition = %q, want %q — a re-provision request must not be suppressed or branded", got, "raised")
	}
}
