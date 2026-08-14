// FILE: platform/messaging/processor_dispatch_resolution_test.go
package messaging

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/gqls/agentchassis/platform/errors"
	"github.com/gqls/agentchassis/platform/kafka"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/types"
	"go.uber.org/zap"
)

// These tests pin what selectWorkflow does when it CANNOT resolve the agent a
// message asked for (bugs_open/239).
//
// Before the fix, every one of these cases fell through to the consuming pod's
// own default workflow. On the shared chassis that pod is `generic`, whose
// entire workflow is one no-op `complete` step, so a dispatch naming a real
// agent ran nothing, filed owner_agent_type='generic' with an empty
// execution_path, and was stamped COMPLETED. Three of the four fall-throughs
// logged nothing at all, which is why eleven black-box bisections could not
// find it: the failure had no signal of any kind, and its DB row was
// indistinguishable from a fast success.
//
// The tests assert the REFUSAL, not the log line: an error out of
// selectWorkflow with a typed code and a reason, and no plan for the caller to
// run.

// dispatchTestProcessor builds a processor standing in for a shared chassis
// pod: its own agent type is `generic`, so any fall-through to Priority 3 hands
// back that pod's own workflow — exactly the substitution these tests forbid.
func dispatchTestProcessor(db *sql.DB) *MessageProcessor {
	return &MessageProcessor{
		agentType: "generic",
		agentRole: "worker",
		agentID:   "agent-239",
		podName:   "chassis-test-pod",
		db:        db,
		logger:    zap.NewNop(),
	}
}

// genericAgentDef is the real shape of the `generic` definition in production:
// a single no-op step whose description says so.
func genericAgentDef() *actions.AgentDefinition {
	return &actions.AgentDefinition{
		Type: "generic",
		DefaultConfig: map[string]interface{}{
			"workflow": map[string]interface{}{
				"start_step": "complete",
				"steps": map[string]interface{}{
					"complete": map[string]interface{}{
						"action":      "complete_workflow",
						"description": "No-op — scheduled task pre_query already did the work",
					},
				},
			},
		},
	}
}

func dispatchMsgCtx(body []byte, action string) *MessageContext {
	return &MessageContext{
		Message: kafka.Message{
			Topic: "system.agent.generic.requests",
			Value: body,
		},
		Headers: map[string]string{"action": action},
		ExecutionContext: &types.ExecutionContext{
			CorrelationID:   "corr-239",
			OrchestrationID: "orch-239",
			ClientID:        "demo_client",
			RequestID:       "req-239",
			Action:          action,
			Timestamp:       time.Now(),
		},
		Logger:        zap.NewNop(),
		CollectedData: map[string]interface{}{},
		StartTime:     time.Now(),
	}
}

// agentDefinitionRow builds the single row FindByType scans, in its column
// order. workflowJSON of "" stands for SQL NULL — a definition that exists but
// carries nothing to run.
func agentDefinitionRow(agentType, workflowJSON string) *sqlmock.Rows {
	cols := []string{
		"id", "type", "display_name", "description", "category",
		"default_config", "workflow", "capabilities", "briefing_questionnaire",
		"usage_count", "version", "is_snapshot",
	}
	var workflow interface{}
	if workflowJSON != "" {
		workflow = []byte(workflowJSON)
	}
	return sqlmock.NewRows(cols).AddRow(
		"def-1", agentType, agentType, "", "",
		[]byte(`{}`), workflow, []byte(`[]`), nil,
		0, 1, false,
	)
}

// TestReproducesBug239_FragmentBodyIsRefusedNotSubstituted is the reproduction.
//
// kcat publishes ONE MESSAGE PER LINE of stdin, so a multi-line JSON envelope
// arrives as several fragments, each invalid JSON, each carrying the full
// header set — including action=orchestrate. This is the first line of the
// envelope the repo's own drive-loop recipe teaches, exactly as it reached the
// chassis on 2026-08-09.
//
// Pre-fix, this test fails: selectWorkflow returned the generic no-op plan and
// a nil error.
func TestReproducesBug239_FragmentBodyIsRefusedNotSubstituted(t *testing.T) {
	fragment := []byte(`{"action":"orchestrate","config":{"agent_type":"page-build-handler"},`)

	p := dispatchTestProcessor(nil)
	msgCtx := dispatchMsgCtx(fragment, "orchestrate")

	selection, err := p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err == nil {
		t.Fatalf("fragmented body was accepted and resolved to %q (source %q) — this is the bug: "+
			"a mangled dispatch ran the pod's own no-op workflow and reported success",
			selection.RunAgentType, selection.Source)
	}
	if !errors.IsDispatchUnresolvable(err) {
		t.Fatalf("error is not a terminal dispatch refusal: %v (code %q)", err, errors.CodeOf(err))
	}
	assertReason(t, err, dispatchReasonParseFailure)
	if len(selection.Plan.Steps) != 0 {
		t.Fatalf("a refused dispatch handed back %d steps to run; want none", len(selection.Plan.Steps))
	}
	// The disposition the intake pool reads: terminal, so the event is marked
	// failed rather than retried for ever.
	if recoverable, _ := RetryDisposition(err); recoverable {
		t.Fatal("parse failure classified recoverable; retrying cannot make a fragment parse")
	}
}

// TestSelectWorkflow_UnknownAgentType_FailsClosedNamingTheType covers the case
// an operator hits by typo, and a stale seed hits by rename.
func TestSelectWorkflow_UnknownAgentType_FailsClosedNamingTheType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM agent_definitions").WillReturnError(sql.ErrNoRows)

	p := dispatchTestProcessor(db)
	body := []byte(`{"action":"orchestrate","config":{"agent_type":"no-such-agent-239"},"input_data":{}}`)
	msgCtx := dispatchMsgCtx(body, "orchestrate")

	selection, err := p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err == nil {
		t.Fatalf("unknown agent type resolved to %q instead of being refused", selection.RunAgentType)
	}
	if !errors.IsDispatchUnresolvable(err) {
		t.Fatalf("want terminal refusal, got %v (code %q)", err, errors.CodeOf(err))
	}
	assertReason(t, err, dispatchReasonTypeUnresolved)
	assertRequestedType(t, err, "no-such-agent-239")
	if recoverable, _ := RetryDisposition(err); recoverable {
		t.Fatal("unknown agent type classified recoverable; the definition will not appear by retrying")
	}
}

// TestSelectWorkflow_TransientLookupError_IsRetryable is the other half of the
// same branch, and the half that used to be indistinguishable from it: any
// error from the lookup — a dial timeout, a recycled connection, pool
// exhaustion — took the identical silent fall-through as a genuine miss.
func TestSelectWorkflow_TransientLookupError_IsRetryable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM agent_definitions").WillReturnError(fmt.Errorf("driver: bad connection"))

	p := dispatchTestProcessor(db)
	body := []byte(`{"action":"orchestrate","config":{"agent_type":"page-build-handler"},"input_data":{}}`)
	msgCtx := dispatchMsgCtx(body, "orchestrate")

	_, err = p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err == nil {
		t.Fatal("a failed lookup produced a workflow to run")
	}
	if !errors.IsDispatchLookupUnavailable(err) {
		t.Fatalf("want transient refusal, got %v (code %q)", err, errors.CodeOf(err))
	}
	if errors.IsDispatchUnresolvable(err) {
		t.Fatal("a transient fault must not read as terminal — the message would be dropped for good")
	}
	recoverable, _ := RetryDisposition(err)
	if !recoverable {
		t.Fatal("transient lookup failure classified permanent; the intake pool would refuse to re-attempt it")
	}
}

// TestSelectWorkflow_DefinitionWithoutWorkflow_FailsClosed covers the quietest
// fall-through of the four: the definition exists, so the lookup "succeeded",
// but default_config->'workflow' is SQL NULL. Pre-fix this logged NOTHING.
func TestSelectWorkflow_DefinitionWithoutWorkflow_FailsClosed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectQuery("FROM agent_definitions").WillReturnRows(agentDefinitionRow("page-build-handler", ""))

	p := dispatchTestProcessor(db)
	body := []byte(`{"action":"orchestrate","config":{"agent_type":"page-build-handler"},"input_data":{}}`)
	msgCtx := dispatchMsgCtx(body, "orchestrate")

	_, err = p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err == nil {
		t.Fatal("a definition with no workflow was accepted")
	}
	assertReason(t, err, dispatchReasonWorkflowMissing)
	assertRequestedType(t, err, "page-build-handler")
}

// TestSelectWorkflow_ExplicitGenericStillRunsItsNoop is the regression that
// keeps the fleet running: ~9,400 scheduled ticks in an eight-day window name
// an agent_type explicitly, and several of them name `generic` — whose real
// workflow IS the no-op, because a pre_query already did the work. Failing
// closed must not touch that path: it resolved through the lookup, so it is a
// resolution, not a fallback.
func TestSelectWorkflow_ExplicitGenericStillRunsItsNoop(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	noop := `{"start_step":"complete","steps":{"complete":{"action":"complete_workflow","description":"No-op — scheduled task pre_query already did the work"}}}`
	mock.ExpectQuery("FROM agent_definitions").WillReturnRows(agentDefinitionRow("generic", noop))

	p := dispatchTestProcessor(db)
	body := []byte(`{"action":"orchestrate","config":{"agent_type":"generic"},"input_data":{}}`)
	msgCtx := dispatchMsgCtx(body, "orchestrate")

	selection, err := p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err != nil {
		t.Fatalf("an explicit generic dispatch was refused: %v", err)
	}
	if selection.RunAgentType != "generic" {
		t.Fatalf("run agent type = %q, want generic", selection.RunAgentType)
	}
	if selection.Source != "agent_definition" {
		t.Fatalf("source = %q, want agent_definition (it resolved; it did not fall back)", selection.Source)
	}
	if len(selection.Plan.Steps) == 0 {
		t.Fatal("the generic no-op workflow came back empty")
	}
}

// TestSelectWorkflow_InlineWorkflow_Priority1Unchanged pins the biggest
// legitimate population in the census: 711 of the orchestrate messages in an
// eight-day window name NO agent type at all and carry their workflow inline
// (call_agent and spawn envelopes). Fail-closed must never reach them.
func TestSelectWorkflow_InlineWorkflow_Priority1Unchanged(t *testing.T) {
	// No DB at all: reaching the lookup would panic the nil dereference or
	// refuse, and either way this test would fail — which is the assertion.
	p := dispatchTestProcessor(nil)
	body := []byte(`{"action":"orchestrate","config":{"workflow":{"start_step":"spawn_ingester","steps":{"spawn_ingester":{"action":"spawn_agent"}}}},"input_data":{}}`)
	msgCtx := dispatchMsgCtx(body, "orchestrate")

	selection, err := p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err != nil {
		t.Fatalf("an inline-workflow envelope was refused: %v", err)
	}
	if selection.Source != "inline_config" {
		t.Fatalf("source = %q, want inline_config", selection.Source)
	}
	if selection.Plan.StartStep != "spawn_ingester" {
		t.Fatalf("start step = %q, want the inline workflow's own", selection.Plan.StartStep)
	}
}

// TestSelectWorkflow_NoTypeNamed_RunsOwnDefault keeps a dedicated or spawned
// pod working: its own type IS the target, so no message names one. This is the
// ONE surviving path to Priority 3 for an orchestration action, and it is
// reached by a message that asked for nothing — never by a failed resolution.
func TestSelectWorkflow_NoTypeNamed_RunsOwnDefault(t *testing.T) {
	p := dispatchTestProcessor(nil)
	body := []byte(`{"action":"orchestrate","input_data":{"domain":"example.com"}}`)
	msgCtx := dispatchMsgCtx(body, "orchestrate")

	selection, err := p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err != nil {
		t.Fatalf("a no-type orchestrate was refused: %v", err)
	}
	if selection.Source != "own_default" {
		t.Fatalf("source = %q, want own_default", selection.Source)
	}
	if selection.RunAgentType != "generic" {
		t.Fatalf("run agent type = %q, want the pod's own type", selection.RunAgentType)
	}
}

// TestSelectWorkflow_NonOrchestrationAction_UnaffectedByFailClosed: the refusal
// is scoped to orchestration actions. A `process` message with an unparseable
// body keeps the old behaviour — it is not a dispatch, and nothing about it
// claims an agent that never ran.
func TestSelectWorkflow_NonOrchestrationAction_UnaffectedByFailClosed(t *testing.T) {
	p := dispatchTestProcessor(nil)
	msgCtx := dispatchMsgCtx([]byte(`not json at all`), "process")

	selection, err := p.selectWorkflow(context.Background(), genericAgentDef(), msgCtx)
	if err != nil {
		t.Fatalf("a non-orchestration action was refused by the dispatch gate: %v", err)
	}
	if selection.Source != "own_default" {
		t.Fatalf("source = %q, want own_default", selection.Source)
	}
}

// TestBody_ParsesOnceAndKeepsItsAnswer pins the single-parse property. The two
// readers used to be independent json.Unmarshal calls that could disagree about
// the same bytes — one choosing the workflow, the other choosing the
// owner_agent_type recorded against it.
func TestBody_ParsesOnceAndKeepsItsAnswer(t *testing.T) {
	msgCtx := dispatchMsgCtx([]byte(`{"config":{"agent_type":"page-rerender"}}`), "orchestrate")

	first, err := msgCtx.Body()
	if err != nil {
		t.Fatalf("first parse failed: %v", err)
	}
	// Mutating the raw bytes afterwards must not change the answer: one parse.
	msgCtx.Message.Value = []byte(`{"config":{"agent_type":"something-else"}}`)
	second, err := msgCtx.Body()
	if err != nil {
		t.Fatalf("second parse failed: %v", err)
	}
	firstType := first["config"].(map[string]interface{})["agent_type"]
	secondType := second["config"].(map[string]interface{})["agent_type"]
	if firstType != secondType {
		t.Fatalf("two reads of one body disagreed: %v then %v", firstType, secondType)
	}
}

// TestBody_ReportsTheParseErrorRatherThanSwallowingIt — the error is the whole
// point of the memoised accessor. Callers that must refuse cannot refuse on an
// error nobody returned.
func TestBody_ReportsTheParseErrorRatherThanSwallowingIt(t *testing.T) {
	msgCtx := dispatchMsgCtx([]byte(` "input_data":{"domain":"webdesign.uk",`), "orchestrate")

	body, err := msgCtx.Body()
	if err == nil {
		t.Fatal("a kcat fragment parsed cleanly")
	}
	if body != nil {
		t.Fatalf("a failed parse returned a body: %v", body)
	}
}

// TestDispatchFailure_SendsUnrecoverableEnvelope asserts what an awaiting parent
// receives when a dispatch is refused: a failure, not a silent success. The old
// path sent nothing at all — the caller saw COMPLETED and nothing else.
func TestDispatchFailure_SendsUnrecoverableEnvelope(t *testing.T) {
	producer := &recordingResponseProducer{}
	p := dispatchTestProcessor(nil)
	p.producer = producer

	msgCtx := responseTestContext()
	derr := p.dispatchUnresolvable(dispatchReasonParseFailure, "", "orchestrate",
		fmt.Errorf("unexpected end of JSON input"))

	if err := p.sendErrorResponse(context.Background(), msgCtx, derr); err != nil {
		t.Fatalf("sendErrorResponse: %v", err)
	}
	sent := onlyResponse(t, producer)
	if got := sent.message.Headers.Status; got != "error_unrecoverable" {
		t.Fatalf("status = %q, want error_unrecoverable", got)
	}
	if coordinatorArm(sent.message.Headers.Status) != "unrecoverable" {
		t.Fatalf("status %q does not route to the unrecoverable arm", sent.message.Headers.Status)
	}
	if sent.message.Body.Error == nil || sent.message.Body.Error.Code != string(errors.ErrDispatchUnresolvable) {
		t.Fatalf("ErrorInfo.Code = %+v, want %s", sent.message.Body.Error, errors.ErrDispatchUnresolvable)
	}
	if sent.message.Body.Error.Recoverable {
		t.Fatal("a terminal dispatch refusal was advertised as recoverable")
	}
}

func assertReason(t *testing.T, err error, want string) {
	t.Helper()
	de, ok := errors.AsDomainError(err)
	if !ok {
		t.Fatalf("error carries no DomainError: %v", err)
	}
	got, _ := de.Details["reason"].(string)
	if got != want {
		t.Fatalf("reason = %q, want %q (err: %v)", got, want, err)
	}
}

func assertRequestedType(t *testing.T, err error, want string) {
	t.Helper()
	de, ok := errors.AsDomainError(err)
	if !ok {
		t.Fatalf("error carries no DomainError: %v", err)
	}
	got, _ := de.Details["requested_agent_type"].(string)
	if got != want {
		t.Fatalf("requested_agent_type = %q, want %q — the FAILED row's owner column is built from this", got, want)
	}
}

// keep the json import honest: the fixtures above are hand-written strings, but
// this asserts they are the shape the production parser accepts.
func TestFixturesAreTheShapeProductionSends(t *testing.T) {
	var body map[string]interface{}
	if err := json.Unmarshal([]byte(`{"action":"orchestrate","config":{"agent_type":"page-build-handler"},"input_data":{"domain":"webdesign.uk"}}`), &body); err != nil {
		t.Fatalf("the canonical drive-loop envelope does not parse: %v", err)
	}
	p := dispatchTestProcessor(nil)
	if got, _ := p.extractGroupInfo(body); got != "page-build-handler" {
		t.Fatalf("extractGroupInfo = %q on the canonical envelope", got)
	}
}

// TestRecordDispatchFailureState_UsesTheDbTheChassisActuallyHas pins the gap the
// post-roll verification found on v1.0.1284 (bugs_open/239).
//
// recordDispatchFailureState originally guarded on p.sqlDB, a second handle
// populated ONLY when DATABASE_URL is set — and it is not set on the chassis pods.
// So in production the guard returned early every time and the FAILED
// orchestration row, the whole point of the function, was never written. The
// refusal still worked and the intake row still recorded it, which is exactly why
// this was invisible: two of the three traces were present.
//
// The test asserts the INSERT is attempted on the handle the chassis actually has.
// On the pre-fix code sqlmock reports the expectation unmet, because nothing was
// executed at all.
//
// It used to also assert the fixture had `sqlDB` nil — "the production shape" —
// because a fixture with both handles set would have passed against the defect.
// bugs_open/259 deleted the second handle, so there is only one shape left to
// drive and that assertion is now vacuous rather than protective.
func TestRecordDispatchFailureState_UsesTheDbTheChassisActuallyHas(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	mock.ExpectExec("INSERT INTO orchestration_states").
		WillReturnResult(sqlmock.NewResult(0, 1))

	p := dispatchTestProcessor(db) // the one handle the chassis has
	msgCtx := dispatchMsgCtx([]byte(`{"action":"orchestrate","config":{"agent_type":"no-such-agent-239"}}`), "orchestrate")
	derr := p.dispatchUnresolvable(dispatchReasonTypeUnresolved, "no-such-agent-239", "orchestrate", nil)

	p.recordDispatchFailureState(context.Background(), msgCtx, derr)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("no FAILED orchestration row was written — the trace bugs_open/239 promises is a no-op in production: %v", err)
	}
}
