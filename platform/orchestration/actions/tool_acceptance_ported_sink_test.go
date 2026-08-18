package actions

import (
	"context"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/gqls/agentchassis/pkg/models"
	orchtypes "github.com/gqls/agentchassis/platform/orchestration/types"
)

// Covers bugs_open/146 (ported-tools slug) / bugs_closed/281 Finding B: a
// failing Tier-4 verdict on a PORTED instance used to write a note saying
// "route this manually" and file NOTHING — pasteboard and vibe-equalizer each
// failed twice (2026-08-05 and 08-14) into that sink. The judge now files
// ported_tool_fix, the vocabulary the Tier-1 and Tier-2 producers already
// emit, when the run item's own spec component resolves to a NON-tool level.
//
// The gate is opt-in-by-evidence: no spec.component_id, a component that
// resolves to a real fork, or any lookup error keeps the old behaviour
// byte-for-byte. The existing suite (tool_acceptance_convergence_test.go /
// tool_acceptance_no_auto_fix_test.go) never supplies spec.component_id, so
// its continued green IS the no-regression proof for the default path.

// portedRun is a judgeRun plus the judge's own log stream. The logs are the
// load-bearing negative assertion: sqlmock's recorder only sees statements it
// EXPECTED, so "no insert was recorded" is vacuously true when the arm fires
// against an unregistered expectation (the attempt errors and is swallowed
// into a Warn). Every utterance of the ported route — success Info, failure
// Warn, lookup-miss Info — contains "ported", so "the logs never say ported"
// is a real proof the arm stayed silent, where the mock's bookkeeping is not.
// Induced: with the level guard deleted, the fork control fails ONLY via this
// channel (the insert attempt logs "ported_tool_fix insert failed").
type portedRun struct {
	judgeRun
	logs *observer.ObservedLogs
}

func (r portedRun) saidPorted() bool {
	for _, e := range r.logs.All() {
		if strings.Contains(e.Message, "ported") {
			return true
		}
	}
	return false
}

// runJudgePortedPath drives the judge down the failure branch of a PORTED
// instance: the fork lookup by function returns nothing, and the run item's
// spec carries the shared wrapper's component id, which resolves to
// componentLevel. specComponentID "" models an older/bespoke run item whose
// spec never named a component — the arm must not fire at all.
//
// expectItem controls whether the driver registers an INSERT expectation:
// when false, an attempted work-item write errors at the mock and surfaces in
// the log stream (see portedRun.saidPorted).
func runJudgePortedPath(t *testing.T, specComponentID, componentLevel string, expectItem bool) portedRun {
	t.Helper()

	run := portedRun{}

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(
		sqlmock.QueryMatcherFunc(func(_, actual string) error {
			run.sql = append(run.sql, actual)
			return nil
		})))
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1. the convergence count
	mock.ExpectQuery("count").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	// 2. the fork lookup by FUNCTION — a ported subject key matches no row
	mock.ExpectQuery("content_components").
		WillReturnRows(sqlmock.NewRows([]string{"id", "page_id"}))
	// 3. the ported route's lookup by the spec's component ID (only reached
	//    when the spec names one)
	if specComponentID != "" {
		mock.ExpectQuery("content_components").
			WillReturnRows(sqlmock.NewRows([]string{"component_level"}).AddRow(componentLevel))
	}
	// 4. the ported_tool_fix insert — registered ONLY when this test expects
	//    the arm to fire; otherwise any work-item write fails the test
	if expectItem {
		mock.ExpectExec("site_work_items").
			WithArgs(
				sqlmock.AnyArg(),              // $1 site_id
				captureArg{got: &run.summary}, // $2 summary
				captureArg{got: &run.spec},    // $3 spec
				captureArg{got: &run.itemKey}, // $4 item_key
				sqlmock.AnyArg(),              // $5 batch_id
			).
			WillReturnResult(sqlmock.NewResult(1, 1))
	}
	// 5. the acceptance-fail doc_note — written LAST, from the outcome
	mock.ExpectQuery("doc_notes").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			captureArg{got: &run.note}, sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("note-1"))

	itemSpec := map[string]interface{}{"function": "vibe-equalizer"}
	if specComponentID != "" {
		itemSpec["component_id"] = specComponentID
		itemSpec["page_id"] = "3d53f70a-d46a-4821-a813-80de2ab7d757"
		itemSpec["page_name"] = "tool-vibe-equalizer"
	}
	collected := map[string]interface{}{
		"input_data":  map[string]interface{}{"spec": itemSpec},
		"site_record": map[string]interface{}{"site_id": benchSite},
		"doc_context": map[string]interface{}{"criteria_json": ""},
		"browser_run": map[string]interface{}{
			"results": []interface{}{
				map[string]interface{}{
					"check_id": "mobile-fit", "profile": "mobile", "passed": false,
					"detail": "page overflows at 390px, widest element div#preview-card.card (380px)",
				},
			},
		},
	}

	core, logs := observer.New(zap.InfoLevel)
	run.logs = logs

	params := ActionParams{
		Context:          context.Background(),
		DB:               db,
		Logger:           zap.New(core),
		Headers:          map[string]string{"agent_type": "tool-acceptance-agent"},
		ExecutionContext: &orchtypes.ExecutionContext{Action: "process"},
		CollectedData:    collected,
		StepConfig:       models.Step{Config: map[string]interface{}{}},
	}

	res, err := JudgeAcceptanceResultsAction(context.Background(), params)
	if err != nil {
		t.Fatalf("judge failed: %v", err)
	}
	run.out, _ = res.(map[string]interface{})

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("registered expectations not consumed: %v", err)
	}
	return run
}

// The firing arm: a ported instance (spec component resolves to a non-tool
// level) whose Tier-4 run failed must file ported_tool_fix — handler-less, at
// needs_human_review, under its own check-segmented dedup key, with the
// verdict's substance in the spec.
func TestPortedAcceptanceFailure_FilesPortedToolFix(t *testing.T) {
	run := runJudgePortedPath(t, "a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef", "section", true)

	ins := run.itemInsert()
	if !strings.Contains(ins, "'ported_tool_fix'") {
		t.Fatalf("expected a ported_tool_fix insert, got:\n%s", ins)
	}
	if !strings.Contains(ins, "'needs_human_review'") {
		t.Errorf("ported item must be born at needs_human_review:\n%s", ins)
	}
	if strings.Contains(ins, "tool-improver") {
		t.Errorf("a ported instance must NEVER be handed to tool-improver (shared-wrapper clobber, bugs_closed/281):\n%s", ins)
	}
	// The sibling producers' contract: no handler agent at all.
	if !strings.Contains(ins, "60, '', 'needs_human_review'") {
		t.Errorf("ported item must carry an EMPTY handler_agent like its sibling producers:\n%s", ins)
	}
	if run.itemKey != "ported_tool_fix:tool_acceptance_tier4:vibe-equalizer:"+benchSite {
		t.Errorf("dedup key must carry the tier-4 check segment, got %q", run.itemKey)
	}
	// Re-verdict semantics: refresh the standing decision, never duplicate it,
	// never clobber a triaging human's own spec keys.
	if !strings.Contains(ins, "DO UPDATE") ||
		!strings.Contains(ins, "spec = site_work_items.spec || EXCLUDED.spec") {
		t.Errorf("ported insert must use the refresh-and-merge idiom:\n%s", ins)
	}
	if !strings.Contains(ins, "ON CONFLICT (site_id, item_key)") ||
		!strings.Contains(ins, "status NOT IN ('complete','failed'") {
		t.Errorf("arbiter predicate must be the canonical dedup one (idx_swi_dedup):\n%s", ins)
	}
	// The spec is what a human routes from.
	for _, want := range []string{
		`"check":"tool_acceptance_tier4"`,
		`"subject_key":"vibe-equalizer"`,
		`"component_id":"a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef"`,
		`"page_id":"3d53f70a-d46a-4821-a813-80de2ab7d757"`,
		`"page_name":"tool-vibe-equalizer"`,
		`"failing_checks":["mobile-fit"]`,
	} {
		if !strings.Contains(run.spec, want) {
			t.Errorf("spec lacks %s: %s", want, run.spec)
		}
	}
	if !strings.Contains(run.summary, "ported tool vibe-equalizer") {
		t.Errorf("summary must name the ported tool, got %q", run.summary)
	}
	// The note is the loop's durable record: it must record the filing, not
	// the old dead-end.
	if strings.Contains(run.note, "route this manually") {
		t.Errorf("note still claims manual routing after an item WAS filed:\n%s", run.note)
	}
	if !strings.Contains(run.note, "ported_tool_fix") {
		t.Errorf("note must record the ported_tool_fix filing:\n%s", run.note)
	}
	// Result-map contract: the new key present only when the arm fired; the
	// old flags stay false.
	if filed, _ := run.out["ported_tool_fix_filed"].(bool); !filed {
		t.Errorf("result must report ported_tool_fix_filed=true, got %v", run.out)
	}
	if created, _ := run.out["improve_tool_created"].(bool); created {
		t.Errorf("no improve_tool may be raised for a ported instance")
	}
	if escalated, _ := run.out["escalated"].(bool); escalated {
		t.Errorf("acceptance_stuck must not fire on the ported route")
	}
	if !run.saidPorted() {
		t.Errorf("the firing arm must log its filing — the negative controls' silence proof depends on this channel speaking when the arm runs")
	}
}

// Negative control 1: a run item whose spec never named a component (older
// item shapes, bespoke dispatches) keeps today's behaviour exactly — no
// lookup, no item of any kind, the note still says routing is manual, and the
// result map does NOT grow the new key.
func TestPortedRoute_NoSpecComponent_IsExactlyTodaysBehaviour(t *testing.T) {
	run := runJudgePortedPath(t, "", "", false)

	if got := run.itemInsert(); got != "" {
		t.Fatalf("no work item of any kind may be written, got:\n%s", got)
	}
	if run.saidPorted() {
		t.Errorf("the ported arm must stay completely silent without a spec component — it spoke:\n%v", run.logs.All())
	}
	if !strings.Contains(run.note, "route this manually") {
		t.Errorf("note must keep the manual-routing record:\n%s", run.note)
	}
	if _, present := run.out["ported_tool_fix_filed"]; present {
		t.Errorf("result map must not grow ported_tool_fix_filed on the old path: %v", run.out)
	}
}

// Negative control 2: the spec component resolves to a real FORK
// (component_level='tool'). The function-keyed lookup missing it is a
// different defect (a renamed function) — filing a ported item would mislabel
// a tool that HAS an automated fixer, so the arm must refuse.
func TestPortedRoute_ForkComponent_DoesNotFile(t *testing.T) {
	run := runJudgePortedPath(t, "a7daa5c5-8cfd-4f2c-8e09-de6abcb637ef", "tool", false)

	if got := run.itemInsert(); got != "" {
		t.Fatalf("a fork-level component must not take the ported route, got:\n%s", got)
	}
	// The load-bearing negative: an arm that fires against the mock's
	// unregistered insert errors SILENTLY at the recorder — only the judge's
	// own logs betray the attempt. Induced: deleting the level guard fails
	// HERE and nowhere else.
	if run.saidPorted() {
		t.Errorf("the ported arm ran against a fork-level component — its logs betray the attempt:\n%v", run.logs.All())
	}
	if !strings.Contains(run.note, "route this manually") {
		t.Errorf("note must keep the manual-routing record:\n%s", run.note)
	}
	if _, present := run.out["ported_tool_fix_filed"]; present {
		t.Errorf("result map must not claim a filing that did not happen: %v", run.out)
	}
}
