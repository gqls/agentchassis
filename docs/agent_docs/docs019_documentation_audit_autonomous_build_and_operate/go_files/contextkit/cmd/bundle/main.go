// Command bundle is a thin ORCHESTRATION WRAPPER around dbcontext + assembler.
// It gathers read-only DB context (schema, capabilities, runtime evidence) via
// dbcontext, writes each to a temp file, then invokes the assembler with those
// files wired in alongside the task/scope/docs you pass through. The result is
// one command that produces a complete bundle "including the SQL", WITHOUT the
// assembler ever running SQL itself.
//
// Why a wrapper and not flags on the assembler: the assembler is a pure,
// read-only, offline composer — it reads source and pastes authored evidence,
// never touching the DB or cluster. Putting query-execution inside it would
// break that property and couple bundle-assembly to a live connection. So SQL
// runs ONLY in dbcontext (itself bounded + read-only: \d, capped SELECTs,
// existing-log reads), and this wrapper sequences the two. It triggers NOTHING
// — no builds, no spawns, no writes.
//
// Usage:
//
//	bundle \
//	  -analysis analysis.json -root /repo -constitution constitution.md \
//	  -task "one sentence" -step debug \
//	  -scope path/a.go:Sym -scope path/b.go \
//	  -include path/registry.go \                 # forwarded to assembler as -scope (file)
//	  -doc docs/016_debugging_guide.md \           # extra authored docs, passed through
//	  -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
//	  -schema-tables site_work_items,pages,page_components \
//	  -runtime-site gamesdesign.co.uk -runtime-page index \
//	  -capabilities \
//	  -out bundle.md
//
// What it runs (each only if its inputs are given; all read-only):
//   - dbcontext -schema <tables>                 -> a schema file        -> assembler -schema
//   - dbcontext -capabilities [-df-filter F]     -> a dbfacts file       -> assembler -dbfacts
//   - dbcontext -runtime-site S [-runtime-page P]-> a runtime file       -> assembler -doc
//
// Without -psql, the DB steps are skipped and it behaves as a plain assembler
// front-end (still useful for the one-command ergonomics). -dry-run prints the
// dbcontext + assembler commands it WOULD run and exits.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func main() {
	// Hand-rolled arg parse (consistent with the other contextkit CLIs, and so
	// repeated -scope/-doc/-include accumulate and order is free).
	var (
		analysis, root, constitution, task string
		step                               = "debug"
		neighbour                          string
		maxNeighbour                       string
		psql                               string
		schemaTables                       string
		runtimeSite, runtimePage           string
		dfFilter                           string
		capabilities                       bool
		out                                string
		dryRun                             bool
		scopes, includes, docs             multiFlag
		assemblerBin, dbcontextBin         string
	)
	assemblerBin = "./cmd/assembler"
	dbcontextBin = "./cmd/dbcontext"

	args := os.Args[1:]
	need := func(i *int) string {
		if *i+1 >= len(args) {
			fmt.Fprintf(os.Stderr, "flag %s needs a value\n", args[*i])
			os.Exit(2)
		}
		*i++
		return args[*i]
	}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-analysis":
			analysis = need(&i)
		case "-root":
			root = need(&i)
		case "-constitution":
			constitution = need(&i)
		case "-task":
			task = need(&i)
		case "-step":
			step = need(&i)
		case "-neighbour":
			neighbour = need(&i)
		case "-max-neighbour":
			maxNeighbour = need(&i)
		case "-scope":
			scopes = append(scopes, need(&i))
		case "-include":
			includes = append(includes, need(&i))
		case "-doc":
			docs = append(docs, need(&i))
		case "-psql":
			psql = need(&i)
		case "-schema-tables":
			schemaTables = need(&i)
		case "-runtime-site":
			runtimeSite = need(&i)
		case "-runtime-page":
			runtimePage = need(&i)
		case "-df-filter":
			dfFilter = need(&i)
		case "-capabilities":
			capabilities = true
		case "-out":
			out = need(&i)
		case "-dry-run":
			dryRun = true
		case "-assembler-bin":
			assemblerBin = need(&i)
		case "-dbcontext-bin":
			dbcontextBin = need(&i)
		case "-h", "-help", "--help":
			usage()
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[i])
			usage()
			os.Exit(2)
		}
	}
	if analysis == "" || root == "" || constitution == "" || task == "" {
		fmt.Fprintln(os.Stderr, "required: -analysis -root -constitution -task")
		usage()
		os.Exit(2)
	}
	if len(scopes) == 0 && len(includes) == 0 {
		fmt.Fprintln(os.Stderr, "at least one -scope (or -include) is required")
		os.Exit(2)
	}

	// Temp dir for the gathered evidence files.
	tmp, err := os.MkdirTemp("", "bundle-")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tempdir: %v\n", err)
		os.Exit(1)
	}
	// Leave the temp files on disk (they're useful to inspect); print the path.

	// Helper to run dbcontext with a set of args, capture stdout to a file.
	runDbToFile := func(label string, dbArgs []string, dest string) (string, bool) {
		full := append([]string{}, dbArgs...)
		if dryRun {
			fmt.Printf("[dry-run] %s %s  > %s\n", dbcontextBin, strings.Join(full, " "), dest)
			return dest, true
		}
		cmd := exec.Command("go", append([]string{"run", dbcontextBin}, full...)...)
		f, ferr := os.Create(dest)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "%s: create %s: %v\n", label, dest, ferr)
			return "", false
		}
		defer f.Close()
		cmd.Stdout = f
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "%s: dbcontext failed: %v (continuing without it)\n", label, err)
			return "", false
		}
		fmt.Fprintf(os.Stderr, "gathered %s -> %s\n", label, dest)
		return dest, true
	}

	var schemaFile, dbfactsFile string
	var docFiles []string
	docFiles = append(docFiles, docs...) // user-passed docs first

	if psql != "" {
		if schemaTables != "" {
			dest := filepath.Join(tmp, "schema.md")
			if p, ok := runDbToFile("schema", []string{"-psql", psql, "-schema", schemaTables}, dest); ok {
				schemaFile = p
			}
		}
		if capabilities {
			dest := filepath.Join(tmp, "dbfacts.md")
			a := []string{"-psql", psql, "-capabilities"}
			if dfFilter != "" {
				a = append(a, "-df-filter", dfFilter)
			}
			if p, ok := runDbToFile("capabilities", a, dest); ok {
				dbfactsFile = p
			}
		}
		if runtimeSite != "" {
			dest := filepath.Join(tmp, "runtime.md")
			a := []string{"-psql", psql, "-runtime-site", runtimeSite}
			if runtimePage != "" {
				a = append(a, "-runtime-page", runtimePage)
			}
			if p, ok := runDbToFile("runtime", a, dest); ok {
				docFiles = append(docFiles, p) // runtime evidence rides in as a -doc
			}
		}
	} else {
		fmt.Fprintln(os.Stderr, "no -psql given: skipping DB gather (schema/runtime/capabilities); composing from -scope/-doc only")
	}

	// Build the assembler invocation.
	aArgs := []string{"run", assemblerBin,
		"-analysis", analysis, "-root", root, "-constitution", constitution,
		"-task", task, "-step", step,
	}
	for _, s := range scopes {
		aArgs = append(aArgs, "-scope", s)
	}
	for _, inc := range includes {
		aArgs = append(aArgs, "-scope", inc) // assembler force-includes files via -scope
	}
	for _, d := range docFiles {
		aArgs = append(aArgs, "-doc", d)
	}
	if schemaFile != "" {
		aArgs = append(aArgs, "-schema", schemaFile)
	}
	if dbfactsFile != "" {
		aArgs = append(aArgs, "-dbfacts", dbfactsFile)
	}
	if neighbour != "" {
		aArgs = append(aArgs, "-neighbour", neighbour)
	}
	if maxNeighbour != "" {
		aArgs = append(aArgs, "-max-neighbour", maxNeighbour)
	}

	if dryRun {
		fmt.Printf("[dry-run] go %s\n", strings.Join(aArgs, " "))
		fmt.Printf("[dry-run] (evidence temp dir: %s)\n", tmp)
		return
	}

	cmd := exec.Command("go", aArgs...)
	cmd.Stderr = os.Stderr
	var sink *os.File
	if out != "" {
		sink, err = os.Create(out)
		if err != nil {
			fmt.Fprintf(os.Stderr, "create %s: %v\n", out, err)
			os.Exit(1)
		}
		defer sink.Close()
		cmd.Stdout = sink
	} else {
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "assembler failed: %v\n", err)
		os.Exit(1)
	}
	if out != "" {
		fmt.Fprintf(os.Stderr, "wrote %s (evidence files kept in %s)\n", out, tmp)
	} else {
		fmt.Fprintf(os.Stderr, "(evidence files kept in %s)\n", tmp)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `bundle — one-command read-only context bundle (dbcontext + assembler)

REQUIRED:
  -analysis FILE       analyser JSON
  -root DIR            repo root (assembler reads bodies)
  -constitution FILE   flat constitution
  -task STRING         one-sentence task
  at least one -scope or -include

SCOPE / DOCS (repeatable):
  -scope path[:Sym]    in-scope file or symbol
  -include path        wiring/registration file (forwarded as -scope)
  -doc FILE            authored doc to include verbatim

DB GATHER (all read-only; only run if -psql given):
  -psql STRING         psql invocation (e.g. 'kubectl exec -n … -- psql -U … -d …')
  -schema-tables CSV   tables to \d            -> assembler -schema
  -runtime-site DOMAIN runtime evidence        -> assembler -doc
  -runtime-page NAME   narrow runtime to a page
  -capabilities        \dx + \df               -> assembler -dbfacts
  -df-filter PATTERN   restrict \df

OTHER:
  -step framing|implementation|debug   (default debug)
  -neighbour callgraph|package
  -max-neighbour N
  -out FILE            write bundle here (default stdout)
  -dry-run             print the dbcontext + assembler commands, run nothing
  -assembler-bin / -dbcontext-bin   override the cmd paths (default ./cmd/…)
`)
}
