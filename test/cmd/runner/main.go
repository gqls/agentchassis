// test/cmd/runner/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

var (
	testSuite = flag.String("test-suite", "all", "Test suite to run (all|unit|integration|e2e|spawning)")
	// These flags are kept for compatibility but the runner will execute suites, not individual tests.
	testAgent = flag.String("test-agent", "", "Specific agent to test (not used by this runner)")
	verbose   = flag.Bool("v", false, "Verbose output")
)

func main() {
	flag.Parse()

	// Set up test environment
	if err := setupTestEnvironment(); err != nil {
		log.Fatal("Failed to setup test environment:", err)
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

func setupTestEnvironment() error {
	// In a real scenario, this could check for DB/Kafka connections.
	// For now, we assume the environment is ready.
	fmt.Println("Assuming test environment is ready...")
	return nil
}

// runGoTest is a helper to execute 'go test' for a given path.
func runGoTest(suite, path string) int {
	fmt.Printf("--- Running %s tests ---\n", suite)
	cmd := exec.Command("go", "test", "-v", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode()
		}
		return 1 // General error
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
	// Spawning tests might have a more specific path if they are separated
	return runGoTest("spawning", "./integration/agents/agent_spawning_test.go")
}
