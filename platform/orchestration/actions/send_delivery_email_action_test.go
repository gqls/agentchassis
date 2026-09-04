package actions

// Tests for the delivery email's send action. The mailer is a seam
// (newDeliverySender) and the DB is sqlmock; delivery.Claim's own semantics are
// proven in platform/delivery — here the properties are ORDER and REFUSAL:
//
//  1. the sender is constructed BEFORE the claim, so a missing SMTP secret
//     cannot strand a site stamped-but-unemailed;
//  2. a gate refusal sends nothing;
//  3. a surviving {{placeholder}} refuses the send;
//  4. the happy path sends the filled body to the customer.

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/mailer"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

type recordingSender struct{ sent []mailer.Message }

func (r *recordingSender) Send(_ context.Context, m mailer.Message) error {
	r.sent = append(r.sent, m)
	return nil
}

// withSender swaps the seam for one test and restores it.
func withSender(t *testing.T, s mailer.Sender, err error) *recordingSender {
	t.Helper()
	rec, _ := s.(*recordingSender)
	orig := newDeliverySender
	newDeliverySender = func() (mailer.Sender, error) { return s, err }
	t.Cleanup(func() { newDeliverySender = orig })
	return rec
}

func baseEmailConfig() map[string]interface{} {
	return map[string]interface{}{
		"site_id":        "input_data.site_id",
		"customer_email": "input_data.customer_email",
		"live_site_url":  "input_data.live_site_url",
		"links_host":     "links.webdesign.uk",
		"subject":        "Your website is ready",
		"body_template":  "Your site is live at {{live_site}} for {{days}} days. Tell us when you have moved: {{confirm_link}}",
	}
}

func emailCollected(siteID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"site_id":        siteID.String(),
			"customer_email": "customer@example.co.uk",
			"live_site_url":  "https://their-site.co.uk",
		},
	}
}

// The sender must be constructed BEFORE the claim: a missing SMTP secret found
// after the stamp strands the site handed-over-but-unemailed. Asserted by what
// does NOT happen — sqlmock is given NO expectations, so any DB touch fails.
func TestSendDeliveryEmailConstructsTheSenderBeforeClaiming(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	withSender(t, nil, fmt.Errorf("no DELIVERY_SMTP_HOST"))

	_, err = SendDeliveryEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: baseEmailConfig()},
		CollectedData:    emailCollected(uuid.New()),
	})
	if err == nil || !strings.Contains(err.Error(), "sender unavailable") {
		t.Fatalf("expected the sender-unavailable refusal, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the database was touched before the sender existed: %v", err)
	}
}

// A gate refusal (unreviewed site) must send nothing.
func TestSendDeliveryEmailSendsNothingWhenTheGateRefuses(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec := withSender(t, &recordingSender{}, nil)

	siteID := uuid.New()
	mock.ExpectQuery(`site_work_items_archive`).
		WithArgs(siteID, "needs_delivery_review").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	_, err = SendDeliveryEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: baseEmailConfig()},
		CollectedData:    emailCollected(siteID),
	})
	if err == nil || !strings.Contains(err.Error(), "not passed pre-delivery review") {
		t.Fatalf("expected the review-gate refusal, got %v", err)
	}
	if len(rec.sent) != 0 {
		t.Fatalf("an unreviewed site was EMAILED: %+v", rec.sent)
	}
}

func expectHappyClaim(mock sqlmock.Sqlmock, siteID uuid.UUID) {
	mock.ExpectQuery(`site_work_items_archive`).
		WithArgs(siteID, "needs_delivery_review").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	now := time.Now().UTC()
	mock.ExpectQuery(`(?s)UPDATE sites.*handed_over_at IS NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"h", "l", "t"}).
			AddRow(now, now.Add(42*24*time.Hour), false))
	mock.ExpectExec(`INSERT INTO customer_access_tokens`).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// The happy path: the customer gets the filled body, no placeholder survives.
func TestSendDeliveryEmailSendsTheFilledBody(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec := withSender(t, &recordingSender{}, nil)

	siteID := uuid.New()
	expectHappyClaim(mock, siteID)

	out, err := SendDeliveryEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: baseEmailConfig()},
		CollectedData:    emailCollected(siteID),
	})
	if err != nil {
		t.Fatalf("SendDeliveryEmailAction: %v", err)
	}
	if len(rec.sent) != 1 {
		t.Fatalf("sent %d emails, want 1", len(rec.sent))
	}
	m := rec.sent[0]
	if m.To[0] != "customer@example.co.uk" {
		t.Errorf("sent to %q", m.To[0])
	}
	if strings.Contains(m.Text, "{{") {
		t.Errorf("a placeholder survived into a customer email: %q", m.Text)
	}
	if !strings.Contains(m.Text, "https://their-site.co.uk") ||
		!strings.Contains(m.Text, "for 30 days") ||
		!strings.Contains(m.Text, "https://links.webdesign.uk/c/") {
		t.Errorf("body missing a required element: %q", m.Text)
	}
	if res, ok := out.(map[string]interface{}); !ok || res["sent"] != true {
		t.Errorf("output = %#v", out)
	}
}

// A template naming a link the config cannot produce must refuse BEFORE the
// stamp: {{zip_link}} with no presign would otherwise be replaced by an EMPTY
// string, mailing a customer "Your files: " with nothing after it — invisible
// to the post-fill scan, because the fill succeeded.
func TestSendDeliveryEmailRefusesAnEmptyLinkBeforeStamping(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec := withSender(t, &recordingSender{}, nil)
	// NO expectations queued: the refusal must happen before ANY database call,
	// so the stamp cannot strand the site over a config mistake.

	cfg := baseEmailConfig()
	cfg["body_template"] = "Your files: {{zip_link}}" // no zip_presigned_url supplied

	_, err = SendDeliveryEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: cfg},
		CollectedData:    emailCollected(uuid.New()),
	})
	if err == nil || !strings.Contains(err.Error(), "would carry a blank") {
		t.Fatalf("expected the empty-link refusal, got %v", err)
	}
	if !strings.Contains(err.Error(), "Nothing was stamped") {
		t.Errorf("the refusal should say the stamp did NOT happen: %v", err)
	}
	if len(rec.sent) != 0 {
		t.Fatalf("emailed despite the empty link: %+v", rec.sent)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the database was touched before the template was validated: %v", err)
	}
}

// A typo the vocabulary cannot fill refuses the send BEFORE the claim.
//
// ⚠ THIS TEST ASSERTED THE OPPOSITE UNTIL 2026-09-04, AND THE CHANGE IS THE
// POINT (bugs_open/475; council c8ed56d2). It used to expect the POST-fill
// "{{" scan to catch `{{zip_lnik}}` — which meant a one-character typo in an
// operator-editable template BURNED THE HANDOVER STAMP before anyone found
// out, recoverable only by the operator re-mint recipe. delivery.Fill.Check now
// refuses any token the vocabulary does not know, pre-claim, so the same typo
// costs nothing.
//
// The old contract is not merely re-worded here, it is superseded: the
// assertions below are strictly stronger (nothing stamped, database untouched).
// The post-fill scan still exists and still has a reachable case of its own —
// see TestSendDeliveryEmailStillScansForASurvivingBraceAfterFilling.
func TestSendDeliveryEmailRefusesAnUnknownPlaceholderBeforeStamping(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec := withSender(t, &recordingSender{}, nil)
	// NO claim expectations: the typo must be caught before any database call.

	siteID := uuid.New()
	cfg := baseEmailConfig()
	cfg["body_template"] = "Files: {{zip_lnik}}" // one transposed character

	_, err = SendDeliveryEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: cfg},
		CollectedData:    emailCollected(siteID),
	})
	if err == nil || !strings.Contains(err.Error(), "{{zip_lnik}}") {
		t.Fatalf("expected the unknown-token refusal naming the typo, got %v", err)
	}
	if !strings.Contains(err.Error(), "Nothing was stamped") {
		t.Errorf("the refusal must say the stamp did NOT happen: %v", err)
	}
	if err2 := mock.ExpectationsWereMet(); err2 != nil {
		t.Errorf("the database was touched before the template was validated: %v", err2)
	}
	if len(rec.sent) != 0 {
		t.Fatalf("a broken template was EMAILED: %+v", rec.sent)
	}
}

// This sender declares every token in the shared vocabulary.
//
// THIS IS THE ASSERTION THE WHOLE bugs_open/475 CHANGE EXISTS FOR. Adding a
// token to delivery.Vocabulary without teaching this sender turns it red here,
// in CI, instead of shipping a blank to a paying customer — which is what the
// previous arrangement (a hand-kept guard slice in each sender, kept in step by
// a comment) could not do. Council c8ed56d2.
func TestDeliveryEmailFillCoversTheVocabulary(t *testing.T) {
	fill := deliveryEmailFill(baseEmailConfig(), "https://example.co.uk", "https://presign")
	for _, problem := range fill.CoverageErrors() {
		t.Errorf("the delivery sender %s", problem)
	}
}

// The post-fill "{{" scan still has a reachable case of its own, so it is not
// redundant now that Check refuses unknown tokens pre-claim.
//
// Check's unknown-token scan looks for a COMPLETE {{...}}; an unterminated "{{"
// is not a token and it passes. Apply then substitutes nothing, and the
// post-fill scan is the thing that stops a customer being emailed a stray
// brace. The claim HAS happened by then — which is exactly why the pre-claim
// checks were widened, and exactly why this last one is kept.
func TestSendDeliveryEmailStillScansForASurvivingBraceAfterFilling(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec := withSender(t, &recordingSender{}, nil)

	siteID := uuid.New()
	expectHappyClaim(mock, siteID)

	cfg := baseEmailConfig()
	cfg["body_template"] = "Your site: {{live_site}} and a stray {{ brace"

	_, err = SendDeliveryEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: cfg},
		CollectedData:    emailCollected(siteID),
	})
	if err == nil || !strings.Contains(err.Error(), "still contains") {
		t.Fatalf("expected the surviving-brace refusal, got %v", err)
	}
	if !strings.Contains(err.Error(), "handover is now stamped") {
		t.Errorf("this one IS post-claim and must say so: %v", err)
	}
	if len(rec.sent) != 0 {
		t.Fatalf("a broken template was EMAILED: %+v", rec.sent)
	}
}
