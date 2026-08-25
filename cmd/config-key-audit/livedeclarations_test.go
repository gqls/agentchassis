// FILE: cmd/config-key-audit/livedeclarations_test.go
//
// The comparison engine's branches, proven WITHOUT a database.
//
// The point of these tests is the FAILURE paths. An auditor that has only ever
// been seen passing cannot be distinguished from one that is not looking — which
// is the exact defect bugs_open/363 is about, and it would be an embarrassment to
// reproduce it inside the fix. So every "cannot tell" branch is asserted, and
// specifically asserted NOT to be a clean result.
//
// The exit-2 paths live in emitLiveDeclarationDrift, which calls os.Exit and so
// cannot be table-tested in-process; they are covered by the induced-drift run
// recorded in the lane's RUNBOOK, and the comparison logic they guard is covered
// here.
package main

import (
	"strings"
	"testing"

	"github.com/gqls/agentchassis/platform/livespec"
)

func TestDescribeScopeNamesWhatWasActuallyRead(t *testing.T) {
	got := describeScope(5, map[string]int{"scheduled_task": 2, "trigger_fn": 2, "trigger_bindings": 1})
	for _, want := range []string{"probed 5 live object(s)", "2 scheduled_task", "2 trigger_fn", "1 trigger_bindings"} {
		if !strings.Contains(got, want) {
			t.Errorf("scope %q omits %q — a finding count alone cannot distinguish 'looked at 5' from 'looked at 0'", got, want)
		}
	}
	// Kinds are sorted, so the report line is stable run to run and a diff between
	// two days' doc_notes rows means something changed, not that a map reordered.
	if a, b := strings.Index(got, "scheduled_task"), strings.Index(got, "trigger_bindings"); a > b {
		t.Errorf("scope is not in a stable order: %q", got)
	}
}

func TestDescribeScopeCannotDressUpAnEmptySweep(t *testing.T) {
	if got := describeScope(0, map[string]int{}); !strings.Contains(got, "0 live objects") {
		t.Errorf("an empty sweep must say so plainly, got %q", got)
	}
}

// The comparator must find each violation shape, one at a time.
func TestFragmentComparisonDetectsEachDriftShape(t *testing.T) {
	d := livespec.Declaration{
		Key:  "test.object",
		Mode: livespec.FragmentMatch,
		Fragments: []livespec.Fragment{
			{Text: "REQUIRED_CLAUSE", Min: 1, Max: 1},
			{Text: "STRICT_BOUNDARY", Forbidden: true},
		},
	}
	cases := []struct {
		name      string
		live      string
		wantCount int
	}{
		{"matches", "... REQUIRED_CLAUSE ...", 0},
		{"clause removed", "... nothing ...", 1},
		{"clause duplicated past Max", "REQUIRED_CLAUSE REQUIRED_CLAUSE", 1},
		{"forbidden text appeared", "REQUIRED_CLAUSE and STRICT_BOUNDARY", 1},
		{"both wrong at once", "STRICT_BOUNDARY only", 2},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := d.CompareFragments(c.live); len(got) != c.wantCount {
				t.Errorf("want %d finding(s), got %d: %v", c.wantCount, len(got), got)
			}
		})
	}
}

// A finding must NAME the object and the fragment, or an operator reading a
// doc_notes row at 07:20 cannot act on it.
func TestFindingsNameTheObjectAndTheFragment(t *testing.T) {
	d := livespec.Declaration{
		Key:       "scheduled_task.some-task.clause",
		Mode:      livespec.FragmentMatch,
		Fragments: []livespec.Fragment{{Text: "THE_CLAUSE", Min: 1}},
	}
	got := d.CompareFragments("drifted")
	if len(got) != 1 {
		t.Fatalf("want 1 finding, got %v", got)
	}
	for _, want := range []string{"scheduled_task.some-task.clause", "THE_CLAUSE"} {
		if !strings.Contains(got[0], want) {
			t.Errorf("finding %q does not name %q", got[0], want)
		}
	}
}

// The count comparator parses TEXT, because every probe returns one text column.
// A non-integer must be a PROBLEM, never a silent pass — that is the nil-result
// shape this estate keeps re-learning.
func TestCountComparisonRefusesUnparseableResults(t *testing.T) {
	d := livespec.Declaration{Key: "trigger_bindings.x", Mode: livespec.CountEqual, ExpectCount: 3}
	cases := []struct {
		name      string
		live      string
		wantCount int
	}{
		{"matches", "3", 0},
		{"whitespace padded", " 3\n", 0},
		{"drifted down", "2", 1},
		{"drifted up", "4", 1},
		{"empty", "", 1},
		{"not a number", "three", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := d.CompareCount(c.live); len(got) != c.wantCount {
				t.Errorf("want %d finding(s), got %d: %v", c.wantCount, len(got), got)
			}
		})
	}
}

// THE REGISTRY MUST BE NON-EMPTY AND EVERY ENTRY PROBEABLE. If livespec ever
// ships with no declarations, or one without a probe, this binary would sweep
// nothing and print a clean line; emitLiveDeclarationDrift exits 2 on that, and
// this test makes sure the condition is not reachable in the first place.
func TestEveryLivespecDeclarationIsProbeableByThisBinary(t *testing.T) {
	if len(livespec.Declarations) == 0 {
		t.Fatal("livespec holds zero declarations — this auditor would report clean over an empty set")
	}
	for _, d := range livespec.Declarations {
		if strings.TrimSpace(d.ProbeSQL) == "" {
			t.Errorf("%s has no ProbeSQL — this binary cannot read it, and would silently never check it", d.Key)
		}
		if d.Mode == livespec.CountEqual && d.ExpectCount <= 0 {
			t.Errorf("%s is CountEqual with ExpectCount %d — a zero expectation matches an empty probe", d.Key, d.ExpectCount)
		}
		if d.Mode == livespec.FragmentMatch && len(d.Fragments) == 0 {
			t.Errorf("%s is FragmentMatch with no fragments — it can never produce a finding", d.Key)
		}
	}
}

// Phase 2 exists to exercise the declarations phase 1 could not.
//
// > **CORRECTED 2026-08-25.** This comment used to say "once this binary ships, the
// > inert count should be able to fall to zero". The binary shipped (CronJob
// > live-declaration-drift-check, daily 07:00 UTC, since 2026-08-23) and the count
// > went 1 → 6, not to zero. The premise was wrong: shipping the auditor does not
// > convert these into Go-checked declarations, because a Go test still has no
// > database. "Live-audit-only" is a permanent, legitimate category, not a backlog.
// > Left uncorrected it was the very defect this lane exists to close — a written
// > statement outliving its truth.
//
// What this test is actually for: the count of auditor-only declarations is
// asserted in BOTH packages, so livespec cannot quietly grow one that this binary
// never probes.
func TestLiveAuditOnlyDeclarationsAreTheOnesThisBinaryUnblocks(t *testing.T) {
	var auditorOnly []string
	for _, d := range livespec.Declarations {
		if d.Phase == livespec.PhaseLiveAudit {
			auditorOnly = append(auditorOnly, d.Key)
		}
	}
	// ⚠ THE CROSS-PACKAGE TRAP THIS LINE KEEPS WALKING INTO. This names a constant
	// defined in platform/livespec, so the two must move in ONE commit. On
	// 2026-08-25 a session pointed it at LiveAuditOnlyDeclarations while that
	// rename existed only in another session's UNCOMMITTED livespec.go (6d3e0027e),
	// breaking this package at HEAD for everyone while compiling fine in its own
	// tree; it was reverted forward-only (8b9128131) and landed properly here.
	// `go build ./...` CANNOT see this — it does not build test files. `go vet` can.
	if len(auditorOnly) != livespec.LiveAuditOnlyDeclarations {
		t.Fatalf("livespec says %d live-audit-only, found %d (%v) — the two must agree or an unchecked "+
			"declaration reads as guarded", livespec.LiveAuditOnlyDeclarations, len(auditorOnly), auditorOnly)
	}
	t.Logf("declarations that only this binary can check: %v", auditorOnly)
}
