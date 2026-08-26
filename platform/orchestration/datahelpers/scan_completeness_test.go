// FILE: platform/orchestration/datahelpers/scan_completeness_test.go
//
// MUTATION CHECK for whoever changes ScanShortfall: replace its body with
// `return nil` and TestScanShortfall_ShortfallIsAnError must fail. If it still
// passes, the guard is not what is producing the refusal and the coverage hole
// bugs_open/410 was filed for is back.
//
// The zero-offered case is not padding: a guard that treats an empty result set
// as a failure fires constantly on healthy input, and a guard that fires
// constantly on healthy input gets loosened within a week. That is the failure
// mode bugs_open/410 pins as the reason the intuitive version of this check
// dies in production.

package datahelpers

import (
	"strings"
	"testing"
)

func TestScanShortfall_ShortfallIsAnError(t *testing.T) {
	err := ScanShortfall(5, 3, "test_reader: rows")
	if err == nil {
		t.Fatal("a scan that kept 3 of 5 offered rows must be refused, not returned as a short result — " +
			"that silent thinning is bugs_open/410 instance 3")
	}

	// The three numbers must all survive into the message: a caller reading the
	// error in a log needs to know it lost rows, how many, and out of how many.
	// An error saying only "scan failed" is the pre-fix Warn with a worse name.
	msg := err.Error()
	for _, want := range []string{"kept 3 of 5", "2 lost", "test_reader: rows", "bugs_open/410"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error message must contain %q so the loss is diagnosable from the log alone; got: %s", want, msg)
		}
	}
}

func TestScanShortfall_CompleteScanIsNotAnError(t *testing.T) {
	if err := ScanShortfall(4, 4, "test_reader: rows"); err != nil {
		t.Fatalf("a complete scan must not be refused: %v", err)
	}
}

func TestScanShortfall_EmptyResultSetIsNotAnError(t *testing.T) {
	// A genuinely empty result set is not a failure. scanBlogArticles makes this
	// exact point in its own doc comment ("a genuinely empty result set is NOT an
	// error and never reaches this branch — attempted stays 0"), and it is what
	// keeps the guard invariant to legitimate SQL-side filtering: a query whose
	// WHERE excludes every row for a key yields nothing and loses nothing.
	if err := ScanShortfall(0, 0, "test_reader: rows"); err != nil {
		t.Fatalf("zero rows offered and zero kept is an empty table, not a loss: %v", err)
	}
}

func TestScanShortfall_KeptExceedingOfferedIsNotAnError(t *testing.T) {
	// Defensive, and it documents a real caller shape rather than an imagined
	// one: a loop may append more than one value per row (a row fanning out into
	// several results). Such a caller has lost nothing, and the guard must not
	// invent a failure for it — kept >= offered is the passing condition, not
	// kept == offered.
	if err := ScanShortfall(2, 3, "test_reader: rows"); err != nil {
		t.Fatalf("kept exceeding offered is a fan-out, not a loss: %v", err)
	}
}
