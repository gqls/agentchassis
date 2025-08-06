// test/cmd/runner/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
)

var (
	testSuite = flag.String("test-suite", "all", "Test suite to run (all|unit|integration|e2e|spawning)")
	testAgent = flag.String("test-agent", "", "Specific agent to test")
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
	// Ensure database is ready
	// Ensure Kafka is ready
	// Load test data
	return nil
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
	fmt.Println("Running unit tests...")
	return testing.MainStart(
		&testDeps{},
		[]testing.InternalTest{
			{Name: "TestSpawnAgentAction", F: TestSpawnAgentAction},
			{Name: "TestWorkflowExecution", F: TestWorkflowExecution},
		},
		nil,
		nil,
		nil,
	).Run()
}

// ... implement other test runners
