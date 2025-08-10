// test/cmd/harness/step1.go
package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	fmt.Println("=== STEP 1: TEST GO COMMAND ===")
	fmt.Printf("TEST_SUITE = %s\n", os.Getenv("TEST_SUITE"))
	fmt.Printf("Working directory: %s\n", mustGetwd())

	// First, test if go command works
	fmt.Println("\n1. Testing 'go version':")
	cmd := exec.Command("go", "version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}

	// List what's in the test directory
	fmt.Println("\n2. Checking /workspace/test directory:")
	entries, err := os.ReadDir("/workspace/test")
	if err != nil {
		fmt.Printf("ERROR reading directory: %v\n", err)
	} else {
		for _, e := range entries {
			fmt.Printf("  - %s (dir: %v)\n", e.Name(), e.IsDir())
		}
	}

	// List test packages
	fmt.Println("\n3. Listing test packages:")
	cmd = exec.Command("go", "list", "./test/e2e/...")
	cmd.Dir = "/workspace"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("ERROR listing packages: %v\n", err)
	}

	fmt.Println("\n=== STEP 1 COMPLETED ===")
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
