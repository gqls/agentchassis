// test/integration/zzz_diagnostic_test.go
package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestZZZDiagnostic(t *testing.T) {
	t.Log("=== DIAGNOSTIC TEST ===")

	// Log runtime info
	t.Logf("Go version: %s", runtime.Version())
	t.Logf("GOOS: %s", runtime.GOOS)
	t.Logf("GOARCH: %s", runtime.GOARCH)
	t.Logf("NumCPU: %d", runtime.NumCPU())
	t.Logf("NumGoroutine: %d", runtime.NumGoroutine())

	// Check if all expected test directories exist
	testDirs := []string{
		"test/integration/agents",
		"test/integration/database",
		"test/integration/kafka",
	}

	for _, dir := range testDirs {
		if _, err := os.Stat(dir); err != nil {
			t.Logf("WARNING: Directory %s does not exist or is not accessible: %v", dir, err)
		} else {
			// Count test files
			count := 0
			filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err == nil && filepath.Ext(path) == ".go" && filepath.Base(path) != "doc.go" {
					count++
				}
				return nil
			})
			t.Logf("Directory %s: %d .go files", dir, count)
		}
	}

	// This test always passes but provides diagnostic info
	t.Log("=== DIAGNOSTIC COMPLETE ===")
}

func TestZZZFinalCheck(t *testing.T) {
	// This ensures we reach the end of the test suite
	fmt.Println("\n>>> FINAL CHECK: This is the last test in the suite")
	fmt.Println(">>> If you see this, all tests were at least attempted")

	// Check for any panic recovery
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Recovered from panic in final check: %v", r)
		}
	}()

	t.Log("Final check completed successfully")
}
