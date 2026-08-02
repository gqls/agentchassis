// FILE: platform/orchestration/actions/component_link_repair_test.go
//
// What this file pins is the SINGLE-COMPONENT call shape (bugs_open/136) — that
// the sibling writers get the same repair, the same fail-open and the same
// durable record as the section save. The rewrite/unlink SEMANTICS are owned by
// datahelpers/link_repair_test.go and the per-section application by
// save_sections_link_repair_test.go; retesting them here would be a second copy
// of the rules, which is the thing this change exists to avoid.

package actions

import (
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/pkg/models"
	"go.uber.org/zap"
)

func pageIndexRows(urls ...string) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"url"})
	for _, u := range urls {
		rows = rows.AddRow(u)
	}
	return rows
}

// A component whose links all resolve must come back byte-identical, and no
// agent_error_log row may be written: a clean write is not an event.
//
// HOW THE NEGATIVE IS ASSERTED, because the obvious way is vacuous. Registering
// no expectation and checking ExpectationsWereMet() proves nothing — it only
// reports on expectations that WERE registered, so a stray INSERT would sail
// past it (and this function swallows a failed log write by design, so the
// return value would not move either). The council's guidelines seat flagged
// exactly this shape in the sibling test below (corr 0275f9c2), and there is a
// standing landmine on it. So: register the INSERT we must NOT see, then require
// ExpectationsWereMet to FAIL naming it.
func TestRepairComponentHTML_CleanComponentIsUntouchedAndSilent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(linkablePageStatusPredicate)).
		WillReturnRows(pageIndexRows("/index.html", "/contact.html"))
	mock.ExpectExec("INSERT INTO agent_error_log").
		WillReturnResult(sqlmock.NewResult(1, 1)) // must stay UNMATCHED

	in := `<section><a href="/contact.html">Talk to us</a></section>`
	got := repairComponentHTMLBeforePersist(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, uuid.New(),
		"example.com", "about", "/about.html", "apply_section_edit", in, zap.NewNop())

	if got != in {
		t.Errorf("clean component was perturbed:\n got %q\nwant %q", got, in)
	}
	err = mock.ExpectationsWereMet()
	if err == nil {
		t.Fatal("a clean component wrote an agent_error_log row — a no-op repair is not an event")
	}
	if !strings.Contains(err.Error(), "ExpectedExec") {
		t.Errorf("the page index was never loaded, so this test proved nothing about silence: %v", err)
	}
}

// The two repair arms reach the single-component caller, and the change is
// recorded durably rather than only in a pod log line (071 gap 3).
func TestRepairComponentHTML_RewritesUnlinksAndRecords(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(linkablePageStatusPredicate)).
		WillReturnRows(pageIndexRows("/index.html", "/contact.html"))

	var gotAction, gotCode, gotContext string
	mock.ExpectExec("INSERT INTO agent_error_log").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			captureArg{&gotAction}, sqlmock.AnyArg(), captureArg{&gotCode},
			captureArg{&gotContext},
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	in := `<p><a href="/contact">Talk</a> or <a href="/pricing">see pricing</a></p>`
	got := repairComponentHTMLBeforePersist(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, uuid.New(),
		"example.com", "about", "/about.html", "apply_section_edit", in, zap.NewNop())

	if !strings.Contains(got, `href="/contact.html"`) {
		t.Errorf("the extensionless target was not rewritten to the stored url: %q", got)
	}
	if strings.Contains(got, `href="/pricing"`) {
		t.Errorf("the phantom link survived: %q", got)
	}
	if !strings.Contains(got, "see pricing") {
		t.Errorf("unlinking dropped the anchor text, which is content: %q", got)
	}
	if gotAction != "apply_section_edit" {
		t.Errorf("action = %q — the origin field is what discriminates this path (097)", gotAction)
	}
	if gotCode != linkRepairErrorCode {
		t.Errorf("error_code = %q, want %q — a new code breaks every query already written", gotCode, linkRepairErrorCode)
	}
	if !strings.Contains(gotContext, "/pricing") {
		t.Errorf("the durable record does not name the href it removed: %s", gotContext)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// FAIL-OPEN, the load-bearing branch. An untrustworthy page set means "we could
// not read the pages", not "this site has no pages" — repairing against it would
// unlink every real link on the component. The write must proceed unrepaired,
// and the skip must be findable a day later.
func TestRepairComponentHTML_UntrustworthyIndexShipsUnrepairedAndRecordsTheSkip(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(linkablePageStatusPredicate)).
		WillReturnError(context.DeadlineExceeded)

	// The skip code is a LITERAL in the statement, not a bound argument, so the
	// query TEXT is where it has to be asserted — matching the INSERT verb alone
	// would pass for a row carrying any code at all.
	var gotAction, gotMessage string
	mock.ExpectExec("CONTENT_LINK_REPAIR_SKIPPED").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			captureArg{&gotAction}, captureArg{&gotMessage}, sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	in := `<p><a href="/pricing">see pricing</a></p>`
	got := repairComponentHTMLBeforePersist(context.Background(),
		ActionParams{DB: db, Logger: zap.NewNop()}, uuid.New(),
		"example.com", "about", "/about.html", "create_report_page", in, zap.NewNop())

	if got != in {
		t.Errorf("a failed index read degraded the write:\n got %q\nwant %q", got, in)
	}
	if gotAction != "create_report_page" {
		t.Errorf("skip row action = %q — the skip must name the path that shipped unrepaired", gotAction)
	}
	if !strings.Contains(gotMessage, "page index unavailable") {
		t.Errorf("skip message = %q — it must say WHY the component shipped unrepaired", gotMessage)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectations: %v", err)
	}
}

// The reversal lever. DB config is live-immediately, so this is what withdraws
// the behaviour fleet-wide without waiting for an image roll — and it must cost
// nothing when off: no index query, no log row.
//
// CORRECTED after the council's guidelines seat read the plan (corr 0275f9c2):
// its first version registered NO expectations and asserted
// ExpectationsWereMet() == nil, which is the vacuous-negative pattern this
// platform has a landmine for — "a mock's own bookkeeping cannot assert a
// NEGATIVE". With nothing registered that call returns nil unconditionally, and
// the html assert could not catch it either, because the fail-open path returns
// the input unchanged too. It would have passed with the lever deleted. Now the
// index query is REGISTERED and required to go UNMATCHED, which is a positive
// assertion that it never ran. Mutation-checked: deleting the lever fails it.
func TestRepairComponentHTML_ConfigLeverOffDoesNothingAtAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(regexp.QuoteMeta(linkablePageStatusPredicate)).
		WillReturnRows(pageIndexRows("/index.html")) // must stay UNMATCHED

	in := `<p><a href="/pricing">see pricing</a></p>`
	got := repairComponentHTMLBeforePersist(context.Background(),
		ActionParams{
			DB:         db,
			Logger:     zap.NewNop(),
			StepConfig: models.Step{Config: map[string]interface{}{"repair_internal_links": false}},
		},
		uuid.New(), "example.com", "about", "/about.html", "apply_section_edit", in, zap.NewNop())

	if got != in {
		t.Errorf("the disabled lever still changed the html: %q", got)
	}
	if err := mock.ExpectationsWereMet(); err == nil {
		t.Fatal("the disabled lever still loaded the page index — 'off' must cost nothing")
	}
}

// The DEFAULT is ON, and asserting that is not pedantry: an off-by-default
// repair is inert, and an inert fix reads exactly like a live one from the
// commit. Checked against a nil config, which is what a caller with no step
// config supplies.
func TestRepairComponentHTML_DefaultsOnWithNoStepConfig(t *testing.T) {
	if !configBoolOrDefault(nil, "repair_internal_links", true) {
		t.Fatal("repair must default ON — the same default as the gate's lever")
	}
}

// The origin fields must survive the trip to the durable record even when the
// action runs outside an execution context (unit tests, odd adoptions): a row
// with an empty agent_type or step_name splits one path across two buckets in
// every later query.
func TestComponentRepairOriginDegradesToStableLabels(t *testing.T) {
	if got := componentRepairAgentType(ActionParams{}); got != "unknown" {
		t.Errorf("agent_type fallback = %q, want %q", got, "unknown")
	}
	if got := componentRepairStepName(ActionParams{}, "create_report_page"); got != "create_report_page" {
		t.Errorf("step_name fallback = %q, want the caller's action name", got)
	}
	if got := componentRepairStepName(ActionParams{CurrentStep: "edit_it"}, "x"); got != "edit_it" {
		t.Errorf("step_name = %q, want the current step", got)
	}
	// saveSectionsStepName delegates here and must keep its own fallback label,
	// because existing agent_error_log queries match on it.
	if got := saveSectionsStepName(ActionParams{}); got != "save_sections" {
		t.Errorf("saveSectionsStepName fallback = %q, want %q", got, "save_sections")
	}
}
