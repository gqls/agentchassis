package actions

// Tests for the follow-up email's send action (bugs_open/477). The mailer is the
// same seam the delivery send uses (newDeliverySender) and the DB is sqlmock;
// delivery.ClaimFollowup's SQL semantics are proven in platform/delivery. What
// is proven HERE is the behaviour a customer or an operator can observe:
//
//  1. the sender is constructed BEFORE the claim, so a missing SMTP secret
//     cannot consume a site's one follow-up and send nothing;
//  2. every refusal from the claim sends nothing AND is not an error — a
//     scheduled run refusing correctly must not fill the queue with red rows;
//  3. the interval is REQUIRED: no default, because it is the owner's decision;
//  4. a template naming a link this dispatch cannot produce refuses BEFORE the
//     claim;
//  5. the happy path sends the filled body, with no placeholder surviving.

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
)

func baseFollowupConfig() map[string]interface{} {
	return map[string]interface{}{
		"site_id":             "input_data.site_id",
		"customer_email":      "input_data.customer_email",
		"live_site_url":       "input_data.live_site_url",
		"links_host":          "links.webdesign.uk",
		"instructions_url":    "https://webdesign.uk/your-site",
		"followup_after_days": float64(7), // float64: this is how JSON config arrives from the DB
		"subject":             "Your website files, and where the instructions live",
		"body_template": "Your site is still live at {{live_site}}. The instructions are kept up to date here: " +
			"{{instructions_url}} . When you have moved, tell us: {{confirm_link}}",
	}
}

func followupCollected(siteID uuid.UUID) map[string]interface{} {
	return map[string]interface{}{
		"input_data": map[string]interface{}{
			"site_id":        siteID.String(),
			"customer_email": "customer@example.co.uk",
			"live_site_url":  "https://their-site.co.uk",
		},
	}
}

// expectHappyFollowupClaim: the claim UPDATE wins, then the confirm token is
// minted. The regex pins BOTH predicates that carry the safety properties —
// `transfer_confirmed_at IS NULL` (the suppression this whole bug is about, and
// the first reader that column has ever had) and `followup_sent_at IS NULL` (the
// at-most-once claim). A claim that lost either would still pass a looser
// pattern: without the first it emails a customer who has already confirmed,
// without the second it emails every time the scheduler ticks.
//
// ⚠ A REGEX OVER THE SQL A MOCK RECEIVED IS A WEAK WITNESS — it proves the
// string was sent, not that Postgres honours it. That is why the same two
// properties are ALSO exercised against a real Postgres in a rolled-back
// transaction; that run is recorded in the lane's NOTES for 2026-09-04.
func expectHappyFollowupClaim(mock sqlmock.Sqlmock, siteID uuid.UUID) {
	now := time.Now().UTC()
	mock.ExpectQuery(`(?s)UPDATE sites.*transfer_confirmed_at IS NULL.*followup_sent_at IS NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"handed", "expires"}).
			AddRow(now.Add(-8*24*time.Hour), now.Add(30*24*time.Hour)))
	mock.ExpectExec(`INSERT INTO customer_access_tokens`).
		WillReturnResult(sqlmock.NewResult(1, 1))
}

// The sender is constructed BEFORE the claim. Asserted by what does NOT happen:
// sqlmock is given NO expectations, so any database touch fails the test. If the
// order were reversed, this site's one follow-up would be consumed by a run that
// then discovered it could not send.
func TestSendFollowupEmailConstructsTheSenderBeforeClaiming(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	withSender(t, nil, fmt.Errorf("no SMTP secret"))

	siteID := uuid.New()
	_, err = SendFollowupEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: baseFollowupConfig()},
		CollectedData:    followupCollected(siteID),
	})
	if err == nil || !strings.Contains(err.Error(), "sender unavailable") {
		t.Fatalf("expected the sender-unavailable refusal, got %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the database was touched before the sender was built: %v", err)
	}
}

// Every refusal the claim can return: nothing is sent, and it is NOT an error.
// A scheduled sender refuses far more often than it sends — a confirmed
// customer, one already followed up, one not yet due — and turning correct
// behaviour into failed work items is how a queue becomes unreadable.
func TestSendFollowupEmailRefusalsSendNothingAndAreNotErrors(t *testing.T) {
	for _, tc := range []struct {
		name   string
		refuse string // the value the diagnosis SELECT returns
		want   string
	}{
		{"the customer already confirmed", "confirmed", "suppressed"},
		{"a follow-up already went", "sent", "already sent"},
		{"handed over too recently", "notdue", "not due"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rec := withSender(t, &recordingSender{}, nil)

			siteID := uuid.New()
			// The claim matches nothing…
			mock.ExpectQuery(`(?s)UPDATE sites.*followup_sent_at IS NULL`).
				WillReturnError(sql.ErrNoRows)
			// …so the diagnosis read runs and says why.
			now := time.Now().UTC()
			row := sqlmock.NewRows([]string{"handed", "expires", "confirmed", "sent"})
			switch tc.refuse {
			case "confirmed":
				row.AddRow(now.Add(-8*24*time.Hour), now.Add(30*24*time.Hour), now, nil)
			case "sent":
				row.AddRow(now.Add(-8*24*time.Hour), now.Add(30*24*time.Hour), nil, now)
			default:
				row.AddRow(nil, nil, nil, nil)
			}
			mock.ExpectQuery(`(?s)SELECT handed_over_at, live_link_expires_at`).WillReturnRows(row)

			out, err := SendFollowupEmailAction(context.Background(), ActionParams{
				DB:               db,
				Logger:           zap.NewNop(),
				ExecutionContext: &types.ExecutionContext{},
				StepConfig:       models.Step{Config: baseFollowupConfig()},
				CollectedData:    followupCollected(siteID),
			})
			if err != nil {
				t.Fatalf("a correct refusal was reported as an error: %v", err)
			}
			if len(rec.sent) != 0 {
				t.Fatalf("a refused site was EMAILED: %+v", rec.sent)
			}
			res, ok := out.(map[string]interface{})
			if !ok || res["sent"] != false {
				t.Fatalf("output = %#v, want sent=false", out)
			}
			if reason, _ := res["reason"].(string); !strings.Contains(reason, tc.want) {
				t.Errorf("reason %q does not say %q — an operator cannot tell why nothing was sent", reason, tc.want)
			}
		})
	}
}

// The interval has no default, and that is the point: the owner said "a week or
// so", which is not a number, and a value invented in Go would become the answer
// without anyone deciding it. Missing, zero and negative all refuse, BEFORE the
// database is touched.
func TestSendFollowupEmailRefusesWithoutAnInterval(t *testing.T) {
	for name, value := range map[string]interface{}{
		"missing":  nil,
		"zero":     float64(0),
		"negative": float64(-3),
		"nonsense": "soon",
	} {
		t.Run(name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rec := withSender(t, &recordingSender{}, nil)

			cfg := baseFollowupConfig()
			if value == nil {
				delete(cfg, "followup_after_days")
			} else {
				cfg["followup_after_days"] = value
			}

			_, err = SendFollowupEmailAction(context.Background(), ActionParams{
				DB:               db,
				Logger:           zap.NewNop(),
				ExecutionContext: &types.ExecutionContext{},
				StepConfig:       models.Step{Config: cfg},
				CollectedData:    followupCollected(uuid.New()),
			})
			if err == nil || !strings.Contains(err.Error(), "followup_after_days") {
				t.Fatalf("expected a followup_after_days refusal, got %v", err)
			}
			if len(rec.sent) != 0 {
				t.Fatalf("an email went out with no interval configured: %+v", rec.sent)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("the database was touched before the interval was validated: %v", err)
			}
		})
	}
}

// A template naming a link this dispatch cannot produce refuses BEFORE the
// claim. {{zip_link}} is the sharp case: a scheduled follow-up has no zip step,
// so it can NEVER produce one, and without this guard the placeholder would be
// replaced by an empty string and the customer would read "Your files: " with
// nothing after it — invisible to the post-fill scan, because the fill succeeded.
func TestSendFollowupEmailRefusesAnUnfillableLinkBeforeClaiming(t *testing.T) {
	for _, placeholder := range []string{"{{zip_link}}", "{{instructions_url}}", "{{stripe_portal_link}}"} {
		t.Run(placeholder, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			rec := withSender(t, &recordingSender{}, nil)

			cfg := baseFollowupConfig()
			cfg["body_template"] = "Your files: " + placeholder
			// instructions_url is the one that IS configurable, so empty it to
			// make its case real rather than vacuous.
			cfg["instructions_url"] = ""

			_, err = SendFollowupEmailAction(context.Background(), ActionParams{
				DB:               db,
				Logger:           zap.NewNop(),
				ExecutionContext: &types.ExecutionContext{},
				StepConfig:       models.Step{Config: cfg},
				CollectedData:    followupCollected(uuid.New()),
			})
			if err == nil || !strings.Contains(err.Error(), placeholder) {
				t.Fatalf("expected a refusal naming %s, got %v", placeholder, err)
			}
			if !strings.Contains(err.Error(), "Nothing was claimed") {
				t.Errorf("the refusal does not say the site is still claimable: %v", err)
			}
			if len(rec.sent) != 0 {
				t.Fatalf("an email with a blank link was SENT: %+v", rec.sent)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("the site was claimed before the template was checked: %v", err)
			}
		})
	}
}

// The happy path. The customer gets the filled body, and it carries the
// instructions URL rather than the instructions themselves (bugs_open/475: a
// copy the customer already holds cannot be corrected).
func TestSendFollowupEmailSendsTheFilledBody(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	rec := withSender(t, &recordingSender{}, nil)

	siteID := uuid.New()
	expectHappyFollowupClaim(mock, siteID)

	out, err := SendFollowupEmailAction(context.Background(), ActionParams{
		DB:               db,
		Logger:           zap.NewNop(),
		ExecutionContext: &types.ExecutionContext{},
		StepConfig:       models.Step{Config: baseFollowupConfig()},
		CollectedData:    followupCollected(siteID),
	})
	if err != nil {
		t.Fatalf("SendFollowupEmailAction: %v", err)
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
		!strings.Contains(m.Text, "https://webdesign.uk/your-site") ||
		!strings.Contains(m.Text, "https://links.webdesign.uk/c/") {
		t.Errorf("body missing a required element: %q", m.Text)
	}
	if res, ok := out.(map[string]interface{}); !ok || res["sent"] != true {
		t.Errorf("output = %#v", out)
	}
}
