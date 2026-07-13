// test/cmd/harness/step2.go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("=== STEP 2: RUN ACTUAL TESTS ===")
	fmt.Printf("TEST_SUITE = %s\n", os.Getenv("TEST_SUITE"))

	// Run tests with direct output
	fmt.Println("\nRunning: go test -v -count=1 ./test/e2e/scenarios/...")

	cmd := exec.Command("go", "test", "-v", "-count=1", "./test/e2e/scenarios/...")
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
}
