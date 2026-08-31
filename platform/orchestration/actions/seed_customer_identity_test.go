// seed_customer_identity_test.go — the P5 arming (P4 plan §4): a build whose
// direction carries the intake's customer fields must land those details on
// the sites row and seed an evidence_base register BEFORE the first work item
// dispatches a build; every other direction shape must be untouched; and a
// re-seed must be structurally unable to clobber an existing register.
//
// MUTATIONS THAT MUST BREAK THESE: (1) make seedCustomerIdentity fire
// unconditionally — the no-intake test's sqlmock fails on the unexpected
// UPDATE; (2) swap the COALESCE argument order in the sites UPDATE (intake
// value overwrites an operator's correction) — the pinned regex fails;
// (3) drop the WHERE NOT EXISTS arm from the site_specs INSERT — the pinned
// regex fails.

package actions

import (
	"context"
	"database/sql/driver"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// The existing-value-wins COALESCE order and the no-clobber INSERT arm are
// pinned as regexes over the actual SQL — they are the two properties a
// hurried refactor would most plausibly lose without any test noticing.
var (
	sitesUpdatePin = regexp.MustCompile(
		`(?s)UPDATE sites.*email\s*=\s*COALESCE\(NULLIF\(email, ''\), NULLIF\(\$2, ''\)\).*company_name\s*=\s*COALESCE\(NULLIF\(company_name, ''\), NULLIF\(\$3, ''\)\)`)
	specsInsertPin = regexp.MustCompile(
		`(?s)INSERT INTO site_specs.*'evidence_base'.*WHERE NOT EXISTS.*aspect = 'evidence_base' AND is_current`)
)

// argContaining matches a string/[]byte argument that contains every needle —
// the seeded payload's exact prose may evolve; the arming facts must not.
type argContaining struct{ needles []string }

func (a argContaining) Match(v driver.Value) bool {
	var s string
	switch t := v.(type) {
	case string:
		s = t
	case []byte:
		s = string(t)
	default:
		return false
	}
	for _, n := range a.needles {
		if !strings.Contains(s, n) {
			return false
		}
	}
	return true
}

func TestSeedCustomerIdentityArmsIntakeBuilds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(sitesUpdatePin.String()).
		WithArgs(siteID.String(), "aaa@example.com", "Boxing Online").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(specsInsertPin.String()).
		WithArgs(siteID.String(),
			argContaining{needles: []string{`"business_name"`, `"contact"`, "BR-TEST01", "aaa@example.com", "Boxing Online", `"customer_attested"`}},
			"order-intake P5 seeding (BR-TEST01)").
		WillReturnResult(sqlmock.NewResult(0, 1))

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	direction := map[string]interface{}{
		"objective":       "a boxing news site",
		"source":          "site-chat-intake",
		"order_reference": "BR-TEST01",
		"customer_email":  "aaa@example.com",
		"customer_name":   "Boxing Online",
	}
	if err := seedCustomerIdentity(context.Background(), tx, siteID, direction, zap.NewNop()); err != nil {
		t.Fatalf("seedCustomerIdentity: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected writes did not happen as pinned: %v", err)
	}
}

func TestSeedCustomerIdentityIgnoresNonIntakeDirections(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	// No Exec expectations: if the helper touches the database for any of
	// these shapes, sqlmock fails the test with "call was not expected".
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	for _, direction := range []map[string]interface{}{
		nil,
		{"objective": "an ordinary estate build"},
		{"adopt_from": "somewhere.uk"},
		{"customer_email": "   ", "customer_name": ""},
	} {
		if err := seedCustomerIdentity(context.Background(), tx, uuid.New(), direction, zap.NewNop()); err != nil {
			t.Fatalf("direction %v: unexpected error %v", direction, err)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("%v", err)
	}
}

func TestSeedCustomerIdentityExistingRegisterSurvives(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectExec(sitesUpdatePin.String()).
		WithArgs(siteID.String(), "aaa@example.com", "").
		WillReturnResult(sqlmock.NewResult(0, 1))
	// The guarded INSERT finds a current register and writes 0 rows — the
	// helper must treat that as success, not retry or error.
	mock.ExpectExec(specsInsertPin.String()).
		WithArgs(siteID.String(), sqlmock.AnyArg(), "order-intake P5 seeding").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	direction := map[string]interface{}{"customer_email": "aaa@example.com"}
	if err := seedCustomerIdentity(context.Background(), tx, siteID, direction, zap.NewNop()); err != nil {
		t.Fatalf("existing register must not be an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("%v", err)
	}
}
