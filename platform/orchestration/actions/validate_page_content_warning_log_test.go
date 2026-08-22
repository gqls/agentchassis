// FILE: platform/orchestration/actions/validate_page_content_warning_log_test.go
//
// bugs_open/071 gap 3, the residual half: a VALID build whose warnings were not
// repaired must persist them durably, and a warning the repair pass DID act on
// must not become a second row (it is already in the repair row's
// context.repairs).
//
// The write/no-write decision is pinned on the pure filter, not on the mocked
// writer: agenterrors.Write is best-effort and swallows an unexpected exec as a
// warn, so sqlmock bookkeeping cannot prove the NEGATIVE (a suite with no
// expectations stays green whether or not a write was attempted). The filter's
// empty return can, and the writer refuses on an empty filter result.
package actions

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

func warningLogFixtureIssues() []ValidationIssue {
	return []ValidationIssue{
		{ // repaired by the pass below — must be EXCLUDED (already in the repair row)
			Type: "phantom_link", Category: "link", Severity: "warning",
			Value: "/dead", Description: "href \"/dead\" has no matching page",
		},
		{ // never repairable — must SURVIVE
			Type: "short_content", Category: "content", Severity: "warning",
			Location: "hero", Value: "38 chars", Description: "section under minimum length",
		},
		{ // not a warning — the failure recorder's business, never this one's
			Type: "phantom_link", Category: "link", Severity: "error",
			Value: "/other", Description: "error-severity link finding",
		},
	}
}

func warningLogFixtureRepairs() []datahelpers.LinkRepair {
	return []datahelpers.LinkRepair{
		{Href: "/dead", NewHref: "/dead.html", Action: datahelpers.LinkRepairRewrite},
	}
}

func TestSurvivingWarnings_RepairedLinkExcludedUnrepairedIncluded(t *testing.T) {
	got := survivingWarnings(warningLogFixtureIssues(), warningLogFixtureRepairs())
	if len(got) != 1 {
		t.Fatalf("surviving = %d issues, want exactly 1 (the unrepaired short_content): %v", len(got), got)
	}
	if got[0]["type"] != "short_content" {
		t.Errorf("survivor is %q, want short_content", got[0]["type"])
	}
	for _, m := range got {
		if m["value"] == "/dead" {
			t.Errorf("repaired href /dead survived the filter — it is already recorded in the repair row and must not become two rows")
		}
	}
}

func TestSurvivingWarnings_EveryWarningRepairedMeansNothingSurvives(t *testing.T) {
	issues := []ValidationIssue{
		{Type: "phantom_link", Category: "link", Severity: "warning", Value: "/dead"},
	}
	if got := survivingWarnings(issues, warningLogFixtureRepairs()); len(got) != 0 {
		t.Fatalf("all warnings were repaired, filter returned %v — the writer would file a duplicate row", got)
	}
}

func TestSurvivingWarnings_ARepairDoesNotShieldANonLinkWarningWithTheSameValue(t *testing.T) {
	// The exclusion is (category == link) AND (href repaired) — a non-link
	// warning that coincidentally carries a repaired href's string must survive.
	issues := []ValidationIssue{
		{Type: "short_content", Category: "content", Severity: "warning", Value: "/dead"},
	}
	if got := survivingWarnings(issues, warningLogFixtureRepairs()); len(got) != 1 {
		t.Fatalf("non-link warning was shielded by a link repair with the same value: %v", got)
	}
}

// warningContextPayload pins the row's context: the count, and that the issues
// array carries the survivor and NOT the repaired href. jsonHas-style presence
// checks cannot see inside the array, and this filter's whole job is what is
// inside the array.
type warningContextPayload struct{}

func (warningContextPayload) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	var payload struct {
		WarningCount int                 `json:"warning_count"`
		Issues       []map[string]string `json:"issues"`
	}
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		return false
	}
	if payload.WarningCount != 1 || len(payload.Issues) != 1 {
		return false
	}
	if payload.Issues[0]["type"] != "short_content" {
		return false
	}
	for _, m := range payload.Issues {
		if m["value"] == "/dead" {
			return false
		}
	}
	return true
}

// failureContextPayload pins the OTHER path (council round 1's missing item):
// an INVALID build's failure row must now carry its warnings under a separate
// key, while context.issues keeps its blocker/error-only shape for existing
// consumers.
type failureContextPayload struct{}

func (failureContextPayload) Match(v driver.Value) bool {
	s, ok := v.(string)
	if !ok {
		return false
	}
	var payload struct {
		WarningCount int                 `json:"warning_count"`
		Issues       []map[string]string `json:"issues"`
		Warnings     []map[string]string `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(s), &payload); err != nil {
		return false
	}
	// issues: blocker/error only — no warning may leak in.
	for _, m := range payload.Issues {
		if m["severity"] == "warning" {
			return false
		}
	}
	if len(payload.Issues) != 1 || payload.Issues[0]["value"] != "/other" {
		return false
	}
	// warnings: both warning-severity fixtures (no repair pass has run on the
	// failure path, so nothing is excluded).
	if payload.WarningCount != 2 || len(payload.Warnings) != 2 {
		return false
	}
	return true
}

func TestWriteValidationFailureLog_WarningsRideTheFailureRow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_error_log`)).
		WithArgs(
			sqlmock.AnyArg(),           // 1 site_id
			"example.com",              // 2 domain
			sqlmock.AnyArg(),           // 3 work_item_id
			sqlmock.AnyArg(),           // 4 orchestration_id
			"page-build-handler",       // 5 agent_type
			sqlmock.AnyArg(),           // 6 agent_id
			sqlmock.AnyArg(),           // 7 pod_name
			"validate_content",         // 8 step_name
			"validate_page_content",    // 9 action
			sqlmock.AnyArg(),           // 10 error_message
			validationDetailErrorCode,  // 11 error_code — the FAILURE row, not the warning row
			"warning",                  // 12 severity (the chassis row is the canonical error)
			failureContextPayload{},    // 13 context
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	writeValidationFailureLog(context.Background(), runningStepParams(db), "", "example.com",
		warningLogFixtureIssues(), 0, 1, zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("failure row does not carry the warnings as specified: %v", err)
	}
}

func TestWriteValidationWarningLog_RowShape(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 13-argument INSERT (positions per log_action_error_test.go's header).
	// agent_type (5), step_name (8), action (9) and error_code (11) are pinned —
	// the provenance-blind AnyArg suite was this package's own recorded defect.
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO agent_error_log`)).
		WithArgs(
			sqlmock.AnyArg(),        // 1 site_id
			"example.com",           // 2 domain
			sqlmock.AnyArg(),        // 3 work_item_id
			sqlmock.AnyArg(),        // 4 orchestration_id
			"page-build-handler",    // 5 agent_type — the gate's own provenance
			sqlmock.AnyArg(),        // 6 agent_id
			sqlmock.AnyArg(),        // 7 pod_name
			"validate_content",      // 8 step_name
			"validate_page_content", // 9 action
			sqlmock.AnyArg(),        // 10 error_message
			validationWarningErrorCode, // 11 error_code
			"warning",               // 12 severity
			warningContextPayload{}, // 13 context
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	params := runningStepParams(db)
	writeValidationWarningLog(context.Background(), params, "", "example.com",
		warningLogFixtureIssues(), warningLogFixtureRepairs(), zap.NewNop())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("surviving-warning row not written as specified: %v", err)
	}
}
