//go:build ignore

// NOT PART OF THIS MODULE'S BUILD — standalone tool; 01simple-debug.go in the same directory also declares func main.
// Two programs in one directory cannot both be compiled as a package, so this
// tag excludes it from `go build ./...`. It is still runnable directly:
//   go run test/cmd/harness/unused/simple/debug.go
// Tag added 2026-07-31; no behaviour changed.
// test/cmd/harness/debug.go
package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	fmt.Println("=== DEBUG OUTPUT TEST ===")
	fmt.Fprintf(os.Stdout, "1. Direct stdout write\n")
	fmt.Fprintf(os.Stderr, "2. Direct stderr write\n")

	os.Stdout.WriteString("3. os.Stdout.WriteString\n")
	os.Stderr.WriteString("4. os.Stderr.WriteString\n")

	fmt.Printf("5. fmt.Printf output\n")

	// Test if environment is set correctly
	fmt.Printf("TEST_SUITE env: %s\n", os.Getenv("TEST_SUITE"))
	fmt.Printf("Working dir: %s\n", mustGetwd())

	// List what's in test directory
	entries, err := os.ReadDir("/workspace/test")
	if err != nil {
		fmt.Printf("ERROR reading /workspace/test: %v\n", err)
	} else {
		fmt.Println("Contents of /workspace/test:")
		for _, e := range entries {
			fmt.Printf("  - %s (dir: %v)\n", e.Name(), e.IsDir())
		}
	}

	fmt.Println("=== END DEBUG OUTPUT ===")
	time.Sleep(2 * time.Second) // Give time for output to flush
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return wd
}
