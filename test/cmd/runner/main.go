// test/cmd/runner/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	testSuite = flag.String("test-suite", "all", "Test suite to run (all|unit|integration|e2e|spawning)")
	testPath  = flag.String("test-path", "", "Base path for tests (auto-detected if not set)")
	verbose   = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	// Auto-detect test path
	if *testPath == "" {
		// Check common locations
		possiblePaths := []string{
			"./test",          // Running from project root
			".",               // Running from test directory
			"/workspace/test", // Running in container
		}

		for _, p := range possiblePaths {
			if _, err := os.Stat(filepath.Join(p, "unit")); err == nil {
				*testPath = p
				break
			}
		}

		if *testPath == "" {
			log.Fatal("Could not find test directory. Use -test-path to specify.")
		}
	}

	fmt.Printf("Test path: %s\n", *testPath)
	fmt.Printf("Running test suite: %s\n", *testSuite)

	// Change to test directory
	if err := os.Chdir(*testPath); err != nil {
		log.Fatalf("Failed to change to test directory: %v", err)
	}

	// Run appropriate test suite
	var exitCode int
	switch *testSuite {
	case "all":
		exitCode = runAllTests()
	case "unit":
		exitCode = runUnitTests()
	case "integration":
		exitCode = runIntegrationTests()
	case "e2e":
		exitCode = runE2ETests()
	case "spawning":
		exitCode = runSpawningTests()
	default:
		log.Fatalf("Unknown test suite: %s", *testSuite)
	}

	os.Exit(exitCode)
}

func runGoTest(suite, path string) int {
	fmt.Printf("--- Running %s tests ---\n", suite)

	// Build the full path from the module root
	fullPath := filepath.Join("./test", strings.TrimPrefix(path, "./"))
	fmt.Printf("Test path: %s\n", fullPath)

	args := []string{"test", "-v"}
	if suite == "unit" {
		args = append(args, "-short")
	}
	args = append(args, fullPath)

	// Run from the module root directory
	cmd := exec.Command("go", args...)
	cmd.Dir = "/workspace" // Set working directory to module root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1
	}
	return 0
}

func runAllTests() int {
	fmt.Println("Running all tests...")
	if code := runUnitTests(); code != 0 {
		return code
	}
	if code := runIntegrationTests(); code != 0 {
		return code
	}
	if code := runE2ETests(); code != 0 {
		return code
	}
	return 0
}

func runUnitTests() int {
	return runGoTest("unit", "./unit/...")
}

func runIntegrationTests() int {
	return runGoTest("integration", "./integration/...")
}

func runE2ETests() int {
	return runGoTest("e2e", "./e2e/...")
}

func runSpawningTests() int {
	return runGoTest("spawning", "./integration/agents/agent_spawning_test.go")
}
