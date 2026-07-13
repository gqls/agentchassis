// test/cmd/harness/main.go
package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("=== GO TEST HARNESS STARTING ===")

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("=== GO TEST HARNESS STARTING ===")

	// Show all environment variables for debugging
	log.Println("Environment variables:")
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "TEST_") || strings.HasPrefix(env, "DB_") || strings.HasPrefix(env, "KAFKA_") {
			log.Printf("  %s", env)
		}
	}

	// List test directories
	log.Println("Test directories:")
	entries, err := os.ReadDir("/workspace/test")
	if err == nil {
		for _, e := range entries {
			if e.IsDir() {
				log.Printf("  /workspace/test/%s", e.Name())
			}
		}
	}

	// Debug environment
	log.Printf("Working directory: %s", mustGetwd())
	log.Printf("TEST_SUITE: %s", os.Getenv("TEST_SUITE"))

	testSuite := os.Getenv("TEST_SUITE")
	if testSuite == "" {
		testSuite = "e2e"
	}

	log.Printf("Running %s tests...", testSuite)

	// Build command based on test suite
	var cmd *exec.Cmd
	switch testSuite {
	case "e2e":
		log.Println("Command: go test -v -count=1 -timeout 60s ./test/e2e/...")
		cmd = exec.Command("go", "test",
			"-v",              // verbose
			"-count=1",        // disable caching
			"-timeout", "60s", // timeout
			"./test/e2e/...") // all e2e tests
	case "integration":
		log.Println("Command: go test -v -count=1 -timeout 30m ./test/integration/...")
		cmd = exec.Command("go", "test",
			"-v",
			"-count=1",
			"-timeout", "30m",
			"./test/integration/...")
	case "unit":
		log.Println("Command: go test -v -short ./test/unit/...")
		cmd = exec.Command("go", "test",
			"-v",
			"-short",
			"./test/unit/...")
	default:
		log.Fatalf("Unknown test suite: %s", testSuite)
	}

	// Set working directory
	cmd.Dir = "/workspace"

	// Get both stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("Failed to get stderr pipe: %v", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start command: %v", err)
	}

	// Read output in real-time with wait group
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		streamOutput("STDOUT", stdout)
	}()

	go func() {
		defer wg.Done()
		streamOutput("STDERR", stderr)
	}()

	// Wait for output to finish
	wg.Wait()

	// Wait for command to complete
	err = cmd.Wait()

	// Print final status
	log.Println("=== TEST EXECUTION COMPLETED ===")

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Tests FAILED with exit code: %d", exitErr.ExitCode())
			os.Exit(exitErr.ExitCode())
		}
		log.Printf("Tests FAILED: %v", err)
		os.Exit(1)
	}

	log.Println("✓ All tests passed")
	os.Exit(0)
}

func streamOutput(prefix string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		// Always print the line immediately to ensure visibility
		fmt.Printf("[%s] %s\n", prefix, line)

		// Flush output to ensure it's visible
		os.Stdout.Sync()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading %s: %v", prefix, err)
	}
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
