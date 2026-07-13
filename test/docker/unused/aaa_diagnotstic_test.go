// test/e2e/scenarios/aaa_diagnostic_test.go
package unused

import (
	"fmt"
	"os"
	"testing"
)

// This test runs first alphabetically and FORCES output
func TestAAADiagnostic(t *testing.T) {
	// Force output to stderr which is less likely to be buffered
	fmt.Fprintf(os.Stderr, "\n=== DIAGNOSTIC TEST STARTING ===\n")
	fmt.Fprintf(os.Stderr, "Test name: %s\n", t.Name())
	fmt.Fprintf(os.Stderr, "Can we see this output?\n")

	// Also use t.Log
	t.Log("=== DIAGNOSTIC via t.Log ===")
	t.Log("If you see this, t.Log is working")

	// And regular fmt.Println
	fmt.Println("=== DIAGNOSTIC via fmt.Println ===")
	fmt.Println("If you see this, fmt.Println is working")

	// Force a failure to see if failure messages appear
	t.Error("FORCED ERROR: This is a diagnostic error to test output")

	fmt.Fprintf(os.Stderr, "=== DIAGNOSTIC TEST ENDING ===\n")
}

// This ensures we see at least something
func TestAAASimple(t *testing.T) {
	fmt.Fprintf(os.Stderr, ">>> TestAAASimple running\n")
	t.Log("Simple test log message")
	// This test passes
}
