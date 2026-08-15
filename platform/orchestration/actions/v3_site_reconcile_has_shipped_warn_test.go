// FILE: platform/orchestration/actions/v3_site_reconcile_has_shipped_warn_test.go
//
// bugs_open/185 tranche 3's fallback, made audible. realisedPageHasShipped reads
// the has_shipped column migration 302 surfaces in load_existing_pages and falls
// back to the narrow build_status test when the column is absent — the
// either-order deployment contract. The council's bug_historian seat noted the
// fallback was SILENT: a reverted migration, or a caller wired to a loader that
// does not select the column, would quietly restore the predicate this bug exists
// to retire. These tests pin the surface: one Warn per reconcile run when the
// column is missing, none when it is present — and never one per page.
package actions

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func reconcileObserved() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zapcore.WarnLevel)
	return zap.New(core), logs
}

func hasShippedWarnings(logs *observer.ObservedLogs) int {
	return len(logs.FilterMessageSnippet("has_shipped").All())
}

// Rows without the column: exactly ONE warning, however many pages there are.
func TestReconcile_WarnsOncePerRunWhenHasShippedColumnAbsent(t *testing.T) {
	logger, logs := reconcileObserved()
	existing := []interface{}{
		realised("index", "/index.html", "deployed", `["hero"]`, false),
		realised("about", "/about.html", "deployed", `["hero"]`, false),
		realised("learning-center", "/learning-center.html", "needs_rebuild", `[]`, false),
	}
	llm := []interface{}{
		llmPage("index", "/index.html", "hero"),
		llmPage("about", "/about.html", "hero"),
		llmPage("learning-center", "/learning-center.html", "content-block"),
	}

	reconcilePlanWithRealised(llm, existing, reconcileOptions{}, logger)

	if n := hasShippedWarnings(logs); n != 1 {
		t.Errorf("has_shipped absent on %d rows: want exactly 1 warning per run, got %d", len(existing), n)
	}
}

// Rows carrying the column — even when its value is FALSE — draw no warning:
// the fallback is not taken, so there is nothing to surface.
func TestReconcile_NoWarningWhenHasShippedColumnPresent(t *testing.T) {
	logger, logs := reconcileObserved()
	withCol := func(name, url, status, sections string, shipped bool) map[string]interface{} {
		rm := realised(name, url, status, sections, false)
		rm["has_shipped"] = shipped
		return rm
	}
	existing := []interface{}{
		withCol("index", "/index.html", "deployed", `["hero"]`, true),
		withCol("brands-index", "/brands-index.html", "needs_rebuild", `[]`, false),
	}
	llm := []interface{}{
		llmPage("index", "/index.html", "hero"),
		llmPage("brands-index", "/brands-index.html", "content-block"),
	}

	reconcilePlanWithRealised(llm, existing, reconcileOptions{}, logger)

	if n := hasShippedWarnings(logs); n != 0 {
		t.Errorf("has_shipped present on every row: want 0 warnings, got %d: %v", n, logs.All())
	}
}
