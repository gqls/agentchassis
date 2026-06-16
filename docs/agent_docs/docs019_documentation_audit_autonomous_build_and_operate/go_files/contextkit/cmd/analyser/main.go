// Command analyser walks a Go source tree and emits the structural summary as
// JSON. It is a thin CLI wrapper over analysis.AnalyseWithExclude — the same
// parsing primitive the in-cluster analyser adapter imports (via the no-exclude
// analysis.Analyse), so harness and production produce identical output for the
// same inputs.
//
// Usage:
//
//	analyser <root-dir>                          # JSON to stdout
//	analyser <root-dir> -pretty                  # indented JSON
//	analyser <root-dir> -exclude go_files_old/,docubundle/   # skip archived copies
//
// Skips ALWAYS: vendor/, testdata/, hidden dirs, *_test.go, and *(N).go
// download-duplicates. Skips ALSO: any path containing an -exclude substring.
//
// When analysing a repo that stores archived copies of its OWN code under docs/
// (the chassis does), exclude them or the index double-counts symbols. The set
// that matched the chassis tree on 2026-06-13:
//
//	-exclude go_files_old/,go_files/,docubundle/,thin_slice_run/,scripts/documentation_project/,docs024_key_docs_latest/
//
// (Tune per tree; verify with the post-analysis duplicate check in the runbook.)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"contextkit/internal/analysis"
)

func main() {
	args := os.Args[1:]
	pretty := false
	var root string
	var exclude []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "-pretty", "--pretty":
			pretty = true
		case "-exclude", "--exclude":
			if i+1 < len(args) {
				i++
				for _, e := range strings.Split(args[i], ",") {
					if e = strings.TrimSpace(e); e != "" {
						exclude = append(exclude, e)
					}
				}
			}
		default:
			if root == "" {
				root = a
			}
		}
	}
	if root == "" {
		fmt.Fprintln(os.Stderr, "usage: analyser <root-dir> [-pretty] [-exclude sub1,sub2,...]")
		os.Exit(2)
	}

	out, err := analysis.AnalyseWithExclude(root, exclude)
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
