// test/e2e/scenarios/diagnostic_test.go
package scenarios

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDiagnostic(t *testing.T) {
	t.Log("=== E2E DIAGNOSTIC TEST ===")

	// Log runtime info
	t.Logf("Go version: %s", runtime.Version())
	t.Logf("GOOS: %s, GOARCH: %s", runtime.GOOS, runtime.GOARCH)
	t.Logf("Working directory: %s", mustGetwd())

	// List all test files in this package
	t.Log("Test files in e2e/scenarios:")
	err := filepath.Walk("test/e2e/scenarios", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			t.Logf("  - %s", filepath.Base(path))
		}
		return nil
	})

	if err != nil {
		t.Logf("Error walking directory: %v", err)
	}

	// Check database connectivity
	t.Log("Checking database connectivity...")
	db := setupTestDB(t)
	if db != nil {
		defer db.Close()
		err := db.Ping()
		if err != nil {
			t.Logf("Database ping failed: %v", err)
		} else {
			t.Log("Database connection successful")
		}
	} else {
		t.Log("Failed to setup test database")
	}

	t.Log("=== DIAGNOSTIC COMPLETE ===")
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}

// This test runs last alphabetically
func TestZZZFinalCheck(t *testing.T) {
	t.Log(">>> FINAL CHECK: Reached end of e2e test suite")
	t.Log(">>> If you see this, all tests were attempted")
}
