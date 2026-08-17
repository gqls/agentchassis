// FILE: platform/orchestration/collected_data_size_tripwire_test.go
//
// Tests for the collected_data size tripwire (bugs_open/289, residual 5).
//
// The bug this instrument exists for ran from 2026-07-29 to 2026-08-17 with
// tool-auditor's collected_data at 22 MB and nothing anywhere noticed, because
// nothing was looking. So the tests that matter are the ones that would fail if
// the tripwire went quiet again — each threshold test is paired with a control
// on the other side of the line, and the silence test is what a "log everything"
// implementation would fail.

package orchestration

import (
	"context"
	"database/sql/driver"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

// stateWithCollectedDataOfSize builds a state whose collected_data marshals to
// at least approxBytes, with the bulk in one named key so the "largest key"
// report has something real to find.
func stateWithCollectedDataOfSize(t *testing.T, bigKey string, approxBytes int) *OrchestrationState {
	t.Helper()
	return &OrchestrationState{
		OrchestrationID: "orch-under-test",
		OwnerAgentType:  "tool-auditor",
		CurrentStep:     "create_items_loop_complete",
		CollectedData: map[string]interface{}{
			"input_data": map[string]interface{}{"site_id": "abc"},
			"small_key":  strings.Repeat("s", 128),
			bigKey:       strings.Repeat("x", approxBytes),
		},
	}
}

func repoWithObservedLogs(level zapcore.Level) (*StateRepository, *observer.ObservedLogs) {
	core, logs := observer.New(level)
	return NewStateRepository(nil, zap.New(core)), logs
}

// The silence half. A normal orchestration must produce NOTHING — an instrument
// that fires on healthy traffic gets muted, and a muted tripwire is the state we
// were already in.
func TestCollectedDataTripwireSilentOnNormalState(t *testing.T) {
	repo, logs := repoWithObservedLogs(zapcore.WarnLevel)
	state := stateWithCollectedDataOfSize(t, "audit_result", 64*1024) // 64 KiB

	repo.reportOversizedCollectedData(state, 64*1024)

	if n := logs.Len(); n != 0 {
		t.Errorf("a 64 KiB collected_data must not trip the wire; got %d log(s): %v", n, logs.All())
	}
}

// Just under the line must also be silent. Without this, a tripwire wired to fire
// unconditionally would pass the threshold tests below and nothing would catch it.
func TestCollectedDataTripwireSilentJustUnderThreshold(t *testing.T) {
	repo, logs := repoWithObservedLogs(zapcore.WarnLevel)
	state := stateWithCollectedDataOfSize(t, "audit_result", 1024)

	repo.reportOversizedCollectedData(state, collectedDataWarnBytes-1)

	if n := logs.Len(); n != 0 {
		t.Errorf("one byte under the warn threshold must stay silent; got %d log(s)", n)
	}
}

func TestCollectedDataTripwireWarnsAndNamesTheLargestKey(t *testing.T) {
	repo, logs := repoWithObservedLogs(zapcore.WarnLevel)
	state := stateWithCollectedDataOfSize(t, "create_items_loop_iter_9_done", 9*1024*1024)

	repo.reportOversizedCollectedData(state, 9*1024*1024)

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log, got %d: %v", len(entries), entries)
	}
	if entries[0].Level != zapcore.WarnLevel {
		t.Errorf("level = %v, want WARN between the two thresholds", entries[0].Level)
	}

	// Naming the key is the whole value of the report: on 289 the total said
	// "something is wrong" and the KEY identified the mechanism.
	fields := entries[0].ContextMap()
	if got := fields["largest_key"]; got != "create_items_loop_iter_9_done" {
		t.Errorf("largest_key = %v, want create_items_loop_iter_9_done", got)
	}
	if got := fields["owner_agent_type"]; got != "tool-auditor" {
		t.Errorf("owner_agent_type = %v, want tool-auditor", got)
	}
	if got := fields["orchestration_id"]; got != "orch-under-test" {
		t.Errorf("orchestration_id = %v, want orch-under-test", got)
	}
}

// The alarm band must be distinguishable from the warn band, or an operator
// filtering on ERROR sees a dead run exactly as they see a large-but-live one.
func TestCollectedDataTripwireEscalatesToErrorAtAlarmSize(t *testing.T) {
	repo, logs := repoWithObservedLogs(zapcore.WarnLevel)
	state := stateWithCollectedDataOfSize(t, "create_items_loop_iter_9_done", 1024)

	repo.reportOversizedCollectedData(state, 29*1024*1024) // tool-auditor's measured max

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("expected exactly 1 log, got %d", len(entries))
	}
	if entries[0].Level != zapcore.ErrorLevel {
		t.Errorf("level = %v, want ERROR at/above the alarm threshold", entries[0].Level)
	}
}

// The thresholds are a claim about real traffic, so they are asserted against the
// census they were set from. If someone later lowers the warn threshold under
// page-content-writer's normal ceiling, this fails rather than the fleet becoming
// noisy in production.
func TestCollectedDataThresholdsSitAboveMeasuredHealthyTraffic(t *testing.T) {
	const (
		pageContentWriterMax = 1854 * 1024 // measured 2026-08-17
		toolSuggesterAvg     = 447 * 1024
		toolAuditorAvg       = 22 * 1024 * 1024
	)
	if collectedDataWarnBytes <= pageContentWriterMax {
		t.Errorf("warn threshold %d is at or below page-content-writer's measured max %d — healthy traffic would trip it",
			collectedDataWarnBytes, pageContentWriterMax)
	}
	if collectedDataWarnBytes <= toolSuggesterAvg {
		t.Errorf("warn threshold %d is at or below tool-suggester's measured average %d", collectedDataWarnBytes, toolSuggesterAvg)
	}
	if collectedDataWarnBytes >= toolAuditorAvg {
		t.Errorf("warn threshold %d is at or above tool-auditor's pathological average %d — the case this exists for would not fire",
			collectedDataWarnBytes, toolAuditorAvg)
	}
	if collectedDataAlarmBytes <= collectedDataWarnBytes {
		t.Errorf("alarm threshold %d must exceed warn threshold %d", collectedDataAlarmBytes, collectedDataWarnBytes)
	}
}

func TestLargestCollectedDataKey(t *testing.T) {
	t.Run("picks the biggest entry, not the first or last", func(t *testing.T) {
		key, size := largestCollectedDataKey(map[string]interface{}{
			"aaa_first": strings.Repeat("a", 10),
			"the_big":   strings.Repeat("b", 5000),
			"zzz_last":  strings.Repeat("c", 20),
		})
		if key != "the_big" {
			t.Errorf("key = %q, want the_big", key)
		}
		if size < 5000 {
			t.Errorf("size = %d, want >= 5000", size)
		}
	})

	t.Run("empty map returns empty, not a panic", func(t *testing.T) {
		key, size := largestCollectedDataKey(map[string]interface{}{})
		if key != "" || size != 0 {
			t.Errorf("got (%q, %d), want (\"\", 0)", key, size)
		}
	})

	t.Run("nil map returns empty, not a panic", func(t *testing.T) {
		key, size := largestCollectedDataKey(nil)
		if key != "" || size != 0 {
			t.Errorf("got (%q, %d), want (\"\", 0)", key, size)
		}
	})

	// An unmarshalable value must be skipped rather than abort the scan — the
	// tripwire runs on states that are already abnormal, which is exactly where
	// odd values live.
	t.Run("skips an unmarshalable value and still finds the biggest", func(t *testing.T) {
		key, _ := largestCollectedDataKey(map[string]interface{}{
			"bad":     make(chan int),
			"the_big": strings.Repeat("b", 300),
		})
		if key != "the_big" {
			t.Errorf("key = %q, want the_big", key)
		}
	})
}

// THE WIRING TEST — the one that actually matters.
//
// Every test above calls reportOversizedCollectedData directly, so all of them
// would pass just as happily against a tripwire that is never called from
// anywhere. That exact failure ("tests that passed with the guard unwired") is
// already on this estate's WRONG_CALLS list, so this drives the REAL
// UpdateStateWithVersion through a mocked DB and asserts the wire fires from the
// production persist path.
func TestTripwireIsWiredIntoTheRealPersistPath(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	args := make([]driver.Value, updateArgCount)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	core, logs := observer.New(zapcore.WarnLevel)
	repo := NewStateRepository(db, zap.New(core))

	// 9 MiB of payload in one key: over the warn threshold, under the alarm one.
	state := stateWithCollectedDataOfSize(t, "create_items_loop_iter_9_done", 9*1024*1024)
	if err := repo.UpdateStateWithVersion(context.Background(), state, state.Version); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	entries := logs.FilterMessageSnippet("collected_data").All()
	if len(entries) == 0 {
		t.Fatal("the real persist path wrote a 9 MiB collected_data and the tripwire never fired — the helper is not wired in")
	}
	if got := entries[0].ContextMap()["largest_key"]; got != "create_items_loop_iter_9_done" {
		t.Errorf("largest_key = %v, want create_items_loop_iter_9_done", got)
	}
}

// Its control: the same real path with a normal payload must persist in silence,
// so the test above cannot be satisfied by a path that simply logs on every write.
func TestRealPersistPathIsSilentForNormalCollectedData(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	args := make([]driver.Value, updateArgCount)
	for i := range args {
		args[i] = sqlmock.AnyArg()
	}
	mock.ExpectExec("UPDATE orchestration_states").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1))

	core, logs := observer.New(zapcore.WarnLevel)
	repo := NewStateRepository(db, zap.New(core))

	state := stateWithCollectedDataOfSize(t, "audit_result", 64*1024)
	if err := repo.UpdateStateWithVersion(context.Background(), state, state.Version); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if n := len(logs.FilterMessageSnippet("collected_data").All()); n != 0 {
		t.Errorf("a normal 64 KiB persist must be silent; got %d tripwire log(s)", n)
	}
}
