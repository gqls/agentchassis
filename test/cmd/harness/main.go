// test/cmd/harness/main.go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("=== GO TEST HARNESS ===")

	testSuite := os.Getenv("TEST_SUITE")
	if testSuite == "" {
		testSuite = "e2e"
	}

	fmt.Printf("TEST_SUITE = %s\n", testSuite)
	fmt.Printf("Working directory: %s\n", mustGetwd())

	var args []string
	switch testSuite {
	case "e2e":
		args = []string{"test", "-v", "-count=1", "-timeout", "60s", "./test/e2e/..."}
	case "integration":
		args = []string{"test", "-v", "-count=1", "-timeout", "30m", "./test/integration/..."}
	case "unit":
		args = []string{"test", "-v", "-short", "./test/unit/..."}
	default:
		fmt.Printf("ERROR: Unknown test suite: %s\n", testSuite)
		os.Exit(1)
	}

	fmt.Printf("\nRunning: go %s\n", strings.Join(args, " "))

	cmd := exec.Command("go", args...)
	cmd.Dir = "/workspace"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Println("Starting test execution...")
	err := cmd.Run()
	fmt.Println("Test execution completed")

	if err != nil {
		fmt.Printf("Tests failed with error: %v\n", err)
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Exit code: %d\n", exitErr.ExitCode())
			os.Exit(exitErr.ExitCode())
		}
		os.Exit(1)
	}

	fmt.Println("=== TESTS PASSED ===")
	os.Exit(0)
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
