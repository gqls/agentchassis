// FILE: platform/orchestration/datahelpers/declared_bool_test.go
//
// bugs_open/193 — a step config key declared with the wrong TYPE was read as no
// declaration at all, silently.
//
// EVERY TEST HERE NAMES THE MUTATION IT KILLS. A test that passes against both
// the fixed and the broken spelling asserts nothing, and this bug's own file says
// so: "Mutation-check it: with the fix reverted, the string case must fail the
// test." The mutations were run — see the commit message for the results.
//
// The load-bearing one is TestGetBoolFieldLoud_NonBoolWarns: the VALUE returned
// for a malformed declaration is the same before and after the fix (the fallback,
// either way). Only the warning distinguishes them. So a test that checked the
// value alone would be green against the defect — which is exactly how this
// defect survived at the loop level while its substep twin was fixed.
package datahelpers

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func observedLogger() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.WarnLevel)
	return zap.New(core), logs
}

// TestGetBoolFieldLoud_AbsentKeyIsSilent — saying nothing is not a mistake, so an
// absent key must return the fallback with NO warning. Run against both fallbacks
// so the test cannot pass by the helper simply always returning false.
//
// MUTATION: delete the `if !present { return fallback }` early return — an absent
// key then reads as a nil interface, takes the non-bool path, and warns. RED.
func TestGetBoolFieldLoud_AbsentKeyIsSilent(t *testing.T) {
	for _, fallback := range []bool{true, false} {
		logger, logs := observedLogger()
		got := GetBoolFieldLoud(map[string]interface{}{"other": 1}, "missing", fallback, logger)
		if got != fallback {
			t.Errorf("absent key with fallback %v returned %v", fallback, got)
		}
		if n := logs.Len(); n != 0 {
			t.Errorf("absent key produced %d warning(s); silence is correct here", n)
		}
	}
}

// TestGetBoolFieldLoud_DeclaredFalseSurvives is the "presence is not truth" case
// and the reason the type assertion is tested on its own.
//
// MUTATION: fold the assertion into the idiom `if v, ok := m[key].(bool); ok && v`
// — a declared FALSE then reads as no declaration and returns the fallback TRUE,
// silently converting an explicit opt-out into an inherited opt-in. RED.
func TestGetBoolFieldLoud_DeclaredFalseSurvives(t *testing.T) {
	logger, logs := observedLogger()
	got := GetBoolFieldLoud(map[string]interface{}{"k": false}, "k", true, logger)
	if got != false {
		t.Error("a declared false was swallowed by the fallback — an explicit opt-out must survive")
	}
	if logs.Len() != 0 {
		t.Error("a well-formed declaration warned")
	}
}

// TestGetBoolFieldLoud_DeclaredTrueWins — the ordinary path.
//
// MUTATION: return `fallback` from the bool branch instead of `value`. RED.
func TestGetBoolFieldLoud_DeclaredTrueWins(t *testing.T) {
	logger, _ := observedLogger()
	if !GetBoolFieldLoud(map[string]interface{}{"k": true}, "k", false, logger) {
		t.Error("a declared true did not win over the fallback")
	}
}

// TestGetBoolFieldLoud_NonBoolWarns is THE test for this bug. The returned value
// is the fallback both before and after the fix; the warning is the entire
// behavioural delta, so this asserts on the warning's CONTENT, not just its
// existence — a reader has to be able to find the offending declaration.
//
// MUTATION: delete the logger.Warn call (keeping the fallback return, i.e. exactly
// the pre-fix behaviour). The value assertions still pass; this goes RED.
func TestGetBoolFieldLoud_NonBoolWarns(t *testing.T) {
	for _, declared := range []interface{}{"true", "false", 1, 0, nil, []interface{}{true}} {
		logger, logs := observedLogger()
		got := GetBoolFieldLoud(map[string]interface{}{"continue_on_error": declared}, "continue_on_error", false, logger)
		if got != false {
			t.Errorf("declared %#v returned %v, want the fallback false", declared, got)
		}
		if logs.Len() != 1 {
			t.Fatalf("declared %#v produced %d warnings, want exactly 1", declared, logs.Len())
		}
		fields := logs.All()[0].ContextMap()
		if fields["config_key"] != "continue_on_error" {
			t.Errorf("warning does not name the key: %v", fields)
		}
		if _, ok := fields["declared_type"]; !ok {
			t.Errorf("warning does not report the declared type, so a reader cannot tell what they typed: %v", fields)
		}
	}
}

// TestGetBoolFieldLoud_CallerFieldsAreCarried — the substep caller identifies its
// declaration with an extra field; if the helper dropped those, the substep
// warning would name a key but not WHICH substep, which is the difference between
// a usable warning and noise.
//
// MUTATION: ignore the variadic `fields` when building the warning. RED.
func TestGetBoolFieldLoud_CallerFieldsAreCarried(t *testing.T) {
	logger, logs := observedLogger()
	GetBoolFieldLoud(map[string]interface{}{"k": "yes"}, "k", false, logger, zap.String("substep", "write_page"))
	if logs.Len() != 1 {
		t.Fatalf("want 1 warning, got %d", logs.Len())
	}
	if logs.All()[0].ContextMap()["substep"] != "write_page" {
		t.Error("the caller's identifying field was dropped from the warning")
	}
}

// TestGetBoolFieldLoud_NilLoggerIsTolerated — a caller without a logger must still
// get the right VALUE rather than being tempted to hand-roll the silent read.
//
// MUTATION: remove the `logger != nil` guard — this panics. RED.
func TestGetBoolFieldLoud_NilLoggerIsTolerated(t *testing.T) {
	if GetBoolFieldLoud(map[string]interface{}{"k": "true"}, "k", false, nil) != false {
		t.Error("nil-logger path returned the wrong value")
	}
	if !GetBoolFieldLoud(map[string]interface{}{"k": true}, "k", false, nil) {
		t.Error("nil-logger path dropped a valid declaration")
	}
}
