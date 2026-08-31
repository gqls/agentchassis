// seed_customer_identity_test.go — the P5 arming (P4 plan §4): a build whose
// direction carries the intake's customer fields must land the CONSENTED
// details on the sites row and seed an evidence_base register BEFORE the first
// work item dispatches a build; every other direction shape must be untouched;
// and a re-seed must be structurally unable to clobber an existing register.
//
// REWRITTEN 2026-08-31 for bugs_open/420 (slug:
// order_intake_publishes_the_billing_email_as_the_sites_public_contact…).
// Until today these tests PINNED THE DEFECT: they asserted that the payer's
// email lands in sites.email and is minted as an "Enquiries reach …" evidence
// fact. That is exactly what published the owner's own address on 13 pages of
// the first paid build. The contract now separates the two: the address the
// customer PAID with is delivery-only, and the address the SITE PUBLISHES comes
// only from direction.published_contact — the customer's explicit answer.
// Absent that answer the site publishes nothing.
//
// MUTATIONS THAT MUST BREAK THESE — apply each ALONE and watch the named test
// fail (a mutation that passes has hit a guard in series; investigate it):
//  1. rebind the sites UPDATE's $2 from publishEmail back to email →
//     TestSeedCustomerIdentityNeverPublishesTheBillingEmail fails on the arg.
//  2. re-mint the contact fact from `email` instead of `publishEmail` →
//     the same test fails on the payload's absent-needle assertion.
//  3. swap publishEmail/email in the published-contact path →
//     TestSeedCustomerIdentityPublishesOnlyTheAskedForContact fails both ways.
//  4. make the helper write unconditionally →
//     TestSeedCustomerIdentityDeliveryOnlyIntakeWritesNothing and
//     TestSeedCustomerIdentityIgnoresNonIntakeDirections fail on sqlmock's
//     "call was not expected".
//  5. swap the COALESCE argument order in the sites UPDATE (intake value
//     overwrites an operator's correction) → the pinned regex fails.
//  6. drop the WHERE NOT EXISTS arm from the site_specs INSERT → pinned regex.

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

// argContaining matches a string/[]byte argument that contains every needle and
// NONE of the absent strings — the seeded payload's exact prose may evolve; the
// arming facts, and the address that must never appear in it, must not.
//
// The absent half is the load-bearing half for bugs_open/420: a test that only
// checks what IS in the payload cannot see a leak, and "the payer's address is
// not here" is the whole claim.
type argContaining struct {
	needles []string
	absent  []string
}

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
	for _, n := range a.absent {
		if strings.Contains(s, n) {
			return false
		}
	}
	return true
}

// TestSeedCustomerIdentityNeverPublishesTheBillingEmail is the regression test
// for the incident itself: an ordinary paid order, no published_contact answer.
// The business NAME is consented (it is constitutive — no page can exist
// without naming the business); the payer's ADDRESS is not, and must reach
// neither the sites column nor the register.
func TestSeedCustomerIdentityNeverPublishesTheBillingEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectBegin()
	// $2 is the PUBLISHED contact and must be empty — not the payer address.
	mock.ExpectQuery(sitesUpdatePin.String()).
		WithArgs(siteID.String(), "", "Boxing Online").
		WillReturnRows(sqlmock.NewRows([]string{"email", "company_name"}).
			AddRow("", "Boxing Online"))
	mock.ExpectExec(specsInsertPin.String()).
		WithArgs(siteID.String(),
			argContaining{
				needles: []string{`"business_name"`, "BR-TEST01", "Boxing Online", `"customer_attested"`},
				absent:  []string{"payer@billing.example", `"contact"`, "Enquiries reach"},
			},
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
		"customer_email":  "payer@billing.example",
		"customer_name":   "Boxing Online",
	}
	if err := seedCustomerIdentity(context.Background(), tx, siteID, direction, zap.NewNop()); err != nil {
		t.Fatalf("seedCustomerIdentity: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expected writes did not happen as pinned: %v", err)
	}
}

// TestSeedCustomerIdentityPublishesOnlyTheAskedForContact is the opt-in arm:
// the customer answered "what contact should the site show?", and THAT address
// — never the billing one — becomes the published contact and the register's
// contact fact.
func TestSeedCustomerIdentityPublishesOnlyTheAskedForContact(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	siteID := uuid.New()
	mock.ExpectBegin()
	mock.ExpectQuery(sitesUpdatePin.String()).
		WithArgs(siteID.String(), "hello@boxingonline.com", "Boxing Online").
		WillReturnRows(sqlmock.NewRows([]string{"email", "company_name"}).
			AddRow("hello@boxingonline.com", "Boxing Online"))
	mock.ExpectExec(specsInsertPin.String()).
		WithArgs(siteID.String(),
			argContaining{
				needles: []string{`"contact"`, "Enquiries reach hello@boxingonline.com.", `"business_name"`},
				absent:  []string{"payer@billing.example"},
			},
			sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	direction := map[string]interface{}{
		"order_reference": "BR-TEST02",
		"customer_email":  "payer@billing.example",
		"customer_name":   "Boxing Online",
		"published_contact": map[string]interface{}{
			"email": "hello@boxingonline.com",
		},
	}
	if err := seedCustomerIdentity(context.Background(), tx, siteID, direction, zap.NewNop()); err != nil {
		t.Fatalf("seedCustomerIdentity: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("%v", err)
	}
}

// TestSeedCustomerIdentityDeliveryOnlyIntakeWritesNothing — an order carrying
// only a payer address (no business name, no published contact) has consented
// to nothing, so there is no row to write at all. sqlmock fails the test if the
// helper touches the database.
func TestSeedCustomerIdentityDeliveryOnlyIntakeWritesNothing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	direction := map[string]interface{}{
		"order_reference": "BR-TEST03",
		"customer_email":  "payer@billing.example",
	}
	if err := seedCustomerIdentity(context.Background(), tx, uuid.New(), direction, zap.NewNop()); err != nil {
		t.Fatalf("delivery-only intake must not be an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("%v", err)
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
		// An empty published_contact object is an absence, not an answer.
		{"customer_email": "payer@billing.example", "published_contact": map[string]interface{}{}},
		{"customer_email": "payer@billing.example", "published_contact": map[string]interface{}{"email": "   "}},
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
	mock.ExpectQuery(sitesUpdatePin.String()).
		WithArgs(siteID.String(), "", "Boxing Online").
		WillReturnRows(sqlmock.NewRows([]string{"email", "company_name"}).
			AddRow("", "Boxing Online"))
	// The guarded INSERT finds a current register and writes 0 rows — the
	// helper must treat that as success, not retry or error.
	mock.ExpectExec(specsInsertPin.String()).
		WithArgs(siteID.String(), sqlmock.AnyArg(), "order-intake P5 seeding").
		WillReturnResult(sqlmock.NewResult(0, 0))

	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	direction := map[string]interface{}{"customer_name": "Boxing Online"}
	if err := seedCustomerIdentity(context.Background(), tx, siteID, direction, zap.NewNop()); err != nil {
		t.Fatalf("existing register must not be an error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("%v", err)
	}
}
