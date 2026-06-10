// Command analyser walks a Go source tree and emits the structural summary as
// JSON. It is a thin CLI wrapper over analysis.Analyse — the same parsing
// primitive the in-cluster analyser adapter imports, so harness and production
// produce identical output.
//
// Usage:
//
//	analyser <root-dir>            # JSON to stdout
//	analyser <root-dir> -pretty    # indented JSON
//
// Skips: vendor/, testdata/, hidden dirs (starting with "."), and *_test.go.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"contextkit/internal/analysis"
)

func main() {
	args := os.Args[1:]
	pretty := false
	var root string
	for _, a := range args {
		switch a {
		case "-pretty", "--pretty":
			pretty = true
		default:
			if root == "" {
				root = a
			}
		}
	}
	if root == "" {
		fmt.Fprintln(os.Stderr, "usage: analyser <root-dir> [-pretty]")
		os.Exit(2)
	}

	out, err := analysis.Analyse(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "analyse error: %v\n", err)
		os.Exit(1)
	}

	enc := json.NewEncoder(os.Stdout)
	if pretty {
		enc.SetIndent("", "  ")
	}
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "encode error: %v\n", err)
		os.Exit(1)
	}
}
