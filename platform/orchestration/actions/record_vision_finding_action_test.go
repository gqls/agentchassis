// FILE: platform/orchestration/actions/record_vision_finding_action_test.go
//
// The contract under test is the one that closes bugs_open/243 candidate 3:
// a critique that reports defects FILES, a critique that says none STAYS
// QUIET, and — the load-bearing direction — a critique with no machine line
// files anyway, because this mechanism's failure mode must never again be
// silence. The SQL-shape test pins the arbiter predicate to the
// acceptance_stuck one (idx_swi_dedup lockstep) and the spec MERGE.

package actions

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

func TestParseVisionVerdictLine(t *testing.T) {
	cases := []struct {
		name, critique, want string
	}{
		{"explicit none", "The page looks clean.\n\nFINDINGS: none", "none"},
		{"explicit none, case-insensitive", "fine.\nfindings: NONE", "none"},
		{"reported", "Low-contrast chips on desktop and mobile.\nFINDINGS: reported", "reported"},
		{"reported with trailing blank lines", "Broken overlap.\nFINDINGS: reported\n\n  \n", "reported"},
		{"no marker at all", "Several options render near-invisible against their backgrounds.", "unparsed"},
		{"marker not on last line", "FINDINGS: none\nActually, one more thing: the CTA is washed out.", "unparsed"},
		{"empty critique", "", "unparsed"},
		{"marker with unexpected value", "ok\nFINDINGS: possibly", "reported"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := parseVisionVerdictLine(c.critique); got != c.want {
				t.Errorf("parseVisionVerdictLine(%q) = %q, want %q", c.critique, got, c.want)
			}
		})
	}
}

// visionParams builds the minimum ActionParams for RecordVisionFindingAction.
// DB nil by default: the no-file paths must not need one.
func visionParams(critique string) ActionParams {
	return ActionParams{
		Logger:     zap.NewNop(),
		StepConfig: models.Step{Config: map[string]interface{}{}},
		CollectedData: map[string]interface{}{
			"vision_look": map[string]interface{}{"result": critique},
			"input_data": map[string]interface{}{
				"spec": map[string]interface{}{"function": "tool-setup-builder"},
			},
			"site_record": map[string]interface{}{
				"site_id": "5fe8785b-223d-41a3-88ee-c07187622381",
				"domain":  "dartsonline.com",
			},
			"browser_run": map[string]interface{}{
				"response": map[string]interface{}{
					"renders": []interface{}{
						map[string]interface{}{
							"uri": "s3://bucket/a_desktop.png", "profile": "desktop",
							"url": "https://dartsonline.com/tools/setup-builder/index.html",
						},
					},
				},
			},
		},
		ExecutionContext: &types.ExecutionContext{
			CorrelationID:   "corr-vision-1",
			OrchestrationID: "orch-vision-1",
			Sender:          types.AgentIdentity{AgentType: "tool-acceptance-agent"},
		},
	}
}

func TestVisionNoneFilesNothing(t *testing.T) {
	params := visionParams("All clean.\nFINDINGS: none")
	out, err := RecordVisionFindingAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]interface{})
	if m["filed"] != false || m["verdict_line"] != "none" {
		t.Fatalf("FINDINGS: none must not file: %+v", m)
	}
}

func TestVisionReportedWithoutDBDoesNotError(t *testing.T) {
	// A finding with nowhere to put it must degrade to a warning, not fail the
	// run — TL-035's best-effort line.
	params := visionParams("The CTA is illegible.\nFINDINGS: reported")
	out, err := RecordVisionFindingAction(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := out.(map[string]interface{})
	if m["filed"] != false || m["verdict_line"] != "reported" {
		t.Fatalf("want filed=false verdict=reported without a DB: %+v", m)
	}
}

func TestVisionReportedFilesDedupedItem(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// page lookup (best-effort)
	mock.ExpectQuery(`SELECT COALESCE\(p\.id::text, ''\)`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("e0325e16-6df6-41de-903c-61f5778708e2"))

	// The filing insert: item_type, handler, status, the DEDUP arbiter with the
	// partial-index predicate, and the spec MERGE — each is load-bearing.
	insertRe := regexp.MustCompile(`(?s)INSERT INTO site_work_items.*'vision_finding'.*'human-review', 'needs_human_review'.*ON CONFLICT \(site_id, item_key\).*WHERE item_key IS NOT NULL AND status NOT IN.*DO UPDATE SET spec = site_work_items\.spec \|\| EXCLUDED\.spec`)
	mock.ExpectExec(insertRe.String()).
		WithArgs(
			"5fe8785b-223d-41a3-88ee-c07187622381",
			sqlmock.AnyArg(), // summary
			sqlmock.AnyArg(), // spec json
			"vision_finding:tool-setup-builder:5fe8785b-223d-41a3-88ee-c07187622381",
			sqlmock.AnyArg(), // batch uuid
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := visionParams("Several options render near-invisible.\nFINDINGS: reported")
	params.DB = db
	out, aErr := RecordVisionFindingAction(context.Background(), params)
	if aErr != nil {
		t.Fatalf("unexpected error: %v", aErr)
	}
	m := out.(map[string]interface{})
	if m["filed"] != true {
		t.Fatalf("want filed=true: %+v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestVisionInsertFailureLeavesDurableTrace(t *testing.T) {
	// Council 310dee45 round 1 (bug_historian, medium): a failed filing must
	// leave a trace that outlives orchestration retention. Round 3
	// (reuse_agent, medium) redirected the fix from a bespoke render-critique
	// doc_note onto agent_error_log — the platform's one durable-failure-trace
	// mechanism (RFC_012) — so the expectation here is that write, not a
	// doc_notes insert.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT COALESCE\(p\.id::text, ''\)`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(""))
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnError(context.DeadlineExceeded)
	mock.ExpectExec(`INSERT INTO agent_error_log`).
		WithArgs(
			"5fe8785b-223d-41a3-88ee-c07187622381", // site_id
			sqlmock.AnyArg(),                        // domain
			sqlmock.AnyArg(),                        // work_item_id
			"orch-vision-1",                         // orchestration_id (JOIN half, inherited)
			sqlmock.AnyArg(),                        // agent_type (provenance half, inherited)
			sqlmock.AnyArg(),                        // agent_id
			sqlmock.AnyArg(),                        // pod_name
			sqlmock.AnyArg(),                        // step_name (provenance half, inherited)
			sqlmock.AnyArg(),                        // action
			sqlmock.AnyArg(),                        // error_message
			"VISION_FINDING_INSERT_FAILED",
			"error",
			sqlmock.AnyArg(), // context jsonb
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := visionParams("The CTA is illegible.\nFINDINGS: reported")
	params.DB = db
	out, aErr := RecordVisionFindingAction(context.Background(), params)
	if aErr != nil {
		t.Fatalf("a filing failure must not fail the run: %v", aErr)
	}
	m := out.(map[string]interface{})
	if m["filed"] != false {
		t.Fatalf("want filed=false after insert failure: %+v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("the durable agent_error_log row was not written: %v", err)
	}
}

func TestVisionUnparsedFilesRatherThanStaysSilent(t *testing.T) {
	// The whole point: an ambiguous critique errs toward a human seeing it.
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery(`SELECT COALESCE\(p\.id::text, ''\)`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(""))
	mock.ExpectExec(`INSERT INTO site_work_items`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	params := visionParams("A defect described with no machine line at the end.")
	params.DB = db
	out, aErr := RecordVisionFindingAction(context.Background(), params)
	if aErr != nil {
		t.Fatalf("unexpected error: %v", aErr)
	}
	m := out.(map[string]interface{})
	if m["filed"] != true || m["verdict_line"] != "unparsed" {
		t.Fatalf("unparsed must file, visibly marked: %+v", m)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
