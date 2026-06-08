// Command assembler builds one paste-ready bundle for a single task.
//
// It consumes the analyser's JSON (the repo's structural summary), the repo
// itself (to pull full bodies by line range), a flat constitution file, and a
// task + scope spec. It renders: constitution, task, the in-scope code in full,
// the surrounding package(s) as signatures, the schema (hand-supplied for now),
// and a pointers note saying what was left out.
//
// Usage:
//
//	assembler -analysis analysis.json -root /path/to/repo \
//	          -constitution constitution.md \
//	          -task "Fix the silent-completion bug in the page build handler" \
//	          -step implementation \
//	          -scope internal/foo/handler.go \
//	          -scope internal/foo/bar.go:ResolveQuery \
//	          [-schema schema.txt] [-max-neighbour 60] > bundle.md
//
// step is framing | implementation | debug.
//   - implementation/debug: in-scope shown in FULL + neighbourhood signatures.
//   - framing:              in-scope shown as SIGNATURES (intent over code).
//   - debug:                adds a runtime-evidence placeholder (no run trace in the thin slice).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---- subset of the analyser's JSON we need ----

type analysis struct {
	Root      string     `json:"root"`
	FileCount int        `json:"file_count"`
	Files     []fileInfo `json:"files"`
}

type fileInfo struct {
	Path      string    `json:"path"`
	Package   string    `json:"package"`
	Functions []funcDef `json:"functions"`
	Types     []typeDef `json:"types"`
}

type funcDef struct {
	Name      string `json:"name"`
	Signature string `json:"signature"`
	Doc       string `json:"doc"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type typeDef struct {
	Name       string  `json:"name"`
	Kind       string  `json:"kind"`
	Doc        string  `json:"doc"`
	Fields     []field `json:"fields"`
	Methods    []field `json:"methods"`
	Underlying string  `json:"underlying"`
	StartLine  int     `json:"start_line"`
	EndLine    int     `json:"end_line"`
}

// ---- repeatable -scope flag ----

type scopeList []string

func (s *scopeList) String() string { return strings.Join(*s, ",") }
func (s *scopeList) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	var (
		analysisPath = flag.String("analysis", "", "path to the analyser JSON")
		root         = flag.String("root", "", "repo root (to read source for full bodies)")
		constitution = flag.String("constitution", "", "path to the flat constitution file")
		task         = flag.String("task", "", "the task description")
		step         = flag.String("step", "implementation", "framing | implementation | debug")
		schemaPath   = flag.String("schema", "", "optional schema text to include")
		maxNeighbour = flag.Int("max-neighbour", 60, "cap on neighbourhood signatures per package")
	)
	var scopes scopeList
	flag.Var(&scopes, "scope", "in-scope file (path) or symbol (path:Name); repeatable")
	var docs scopeList
	flag.Var(&docs, "doc", "path to an authored doc to include verbatim (debug guide, a 003 section); repeatable")
	flag.Parse()

	must := func(name, val string) {
		if val == "" {
			fmt.Fprintf(os.Stderr, "missing required -%s\n", name)
			os.Exit(2)
		}
	}
	must("analysis", *analysisPath)
	must("root", *root)
	must("constitution", *constitution)
	must("task", *task)
	if len(scopes) == 0 {
		fmt.Fprintln(os.Stderr, "at least one -scope is required")
		os.Exit(2)
	}
	switch *step {
	case "framing", "implementation", "debug":
	default:
		fmt.Fprintf(os.Stderr, "step must be framing|implementation|debug, got %q\n", *step)
		os.Exit(2)
	}

	an, err := loadAnalysis(*analysisPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load analysis: %v\n", err)
		os.Exit(1)
	}
	byPath := map[string]fileInfo{}
	byPackage := map[string][]fileInfo{}
	for _, f := range an.Files {
		byPath[f.Path] = f
		byPackage[f.Package] = append(byPackage[f.Package], f)
	}

	constText, err := os.ReadFile(*constitution)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read constitution: %v\n", err)
		os.Exit(1)
	}
	schemaText := ""
	if *schemaPath != "" {
		b, err := os.ReadFile(*schemaPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read schema: %v\n", err)
			os.Exit(1)
		}
		schemaText = string(b)
	}

	full := *step != "framing" // framing shows signatures only

	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format, a...) }

	// ---- header ----
	w("# Task bundle\n\n")
	w("- **Task:** %s\n", *task)
	w("- **Step:** %s\n", *step)
	w("- **In scope:** %s\n", strings.Join(scopes, ", "))
	w("- **Repo:** %s (%d files analysed)\n", an.Root, an.FileCount)
	w("- **Note:** in-scope items shown %s; the surrounding package(s) are shown as signatures; everything else is omitted — ask for a path to add it.\n\n",
		map[bool]string{true: "in full", false: "as signatures"}[full])

	// ---- constitution ----
	w("---\n\n## Constitution (always-on rules)\n\n")
	w("%s\n\n", strings.TrimSpace(string(constText)))

	// ---- task ----
	w("---\n\n## Task\n\n%s\n\n", strings.TrimSpace(*task))

	// ---- reference documents (pasted-in authored docs: debug guides, 003 sections) ----
	if len(docs) > 0 {
		w("---\n\n## Reference documents (standards / guides for this task)\n\n")
		for _, d := range docs {
			b2, err := os.ReadFile(d)
			if err != nil {
				w("> doc %q: %v\n\n", d, err)
				continue
			}
			w("### %s\n\n%s\n\n", filepath.Base(d), strings.TrimSpace(string(b2)))
		}
	}

	// ---- in-scope ----
	w("---\n\n## In-scope code\n\n")
	inScopePkgs := map[string]bool{}
	type seen struct{ path, sym string }
	inScopeSyms := map[seen]bool{}
	for _, sc := range scopes {
		path, sym := splitScope(sc)
		fi, ok := byPath[path]
		if !ok {
			w("> scope %q: path not found in analysis — skipped\n\n", sc)
			continue
		}
		inScopePkgs[fi.Package] = true

		if sym == "" {
			// whole file
			src, err := readLines(*root, path, 1, -1)
			if err != nil {
				w("> scope %q: %v\n\n", sc, err)
				continue
			}
			w("### %s (package `%s`) — whole file\n\n", path, fi.Package)
			if full {
				w("```go\n%s\n```\n\n", src)
			} else {
				w("%s\n\n", fileSignatures(fi))
			}
			continue
		}

		// a named symbol
		start, end, kind, found := locateSymbol(fi, sym)
		if !found {
			w("> scope %q: symbol not found in %s — skipped\n\n", sc, path)
			continue
		}
		inScopeSyms[seen{path, sym}] = true
		w("### %s — `%s` (%s, lines %d–%d)\n\n", path, sym, kind, start, end)
		if full {
			src, err := readLines(*root, path, start, end)
			if err != nil {
				w("> %v\n\n", err)
				continue
			}
			w("```go\n%s\n```\n\n", src)
		} else {
			w("`%s`\n\n", symbolSignature(fi, sym))
		}
	}

	// ---- neighbourhood ----
	w("---\n\n## Neighbourhood (signatures)\n\n")
	if len(inScopePkgs) == 0 {
		w("_none — no in-scope package resolved_\n\n")
	}
	pkgs := sortedKeys(inScopePkgs)
	for _, pkg := range pkgs {
		w("### package `%s`\n\n", pkg)
		var lines []string
		for _, fi := range byPackage[pkg] {
			for _, fn := range fi.Functions {
				if inScopeSyms[seen{fi.Path, fn.Name}] {
					continue // already shown in full above
				}
				lines = append(lines, fn.Signature)
			}
			for _, td := range fi.Types {
				if inScopeSyms[seen{fi.Path, td.Name}] {
					continue
				}
				lines = append(lines, typeSignature(td))
			}
		}
		sort.Strings(lines)
		if len(lines) == 0 {
			w("_no other declarations in this package_\n\n")
			continue
		}
		shown := lines
		omitted := 0
		if len(lines) > *maxNeighbour {
			shown = lines[:*maxNeighbour]
			omitted = len(lines) - *maxNeighbour
		}
		w("```go\n%s\n```\n", strings.Join(shown, "\n"))
		if omitted > 0 {
			w("_…and %d more in this package (ask to include)_\n", omitted)
		}
		w("\n")
	}

	// ---- schema ----
	w("---\n\n## Schema\n\n")
	if schemaText == "" {
		w("_none provided — supply with -schema if this task touches the database_\n\n")
	} else {
		w("```\n%s\n```\n\n", strings.TrimSpace(schemaText))
	}

	// ---- runtime evidence (debug only) ----
	if *step == "debug" {
		w("---\n\n## Runtime evidence\n\n")
		w("_not available in the thin slice. For a real debug bundle this is the run trace, step sequence, errors, and logs correlated by `orchestration_id`._\n\n")
	}

	// ---- pointers ----
	w("---\n\n## Pointers\n\n")
	w("- This bundle is a selection from %d analysed files. In-scope: %s.\n", an.FileCount, strings.Join(scopes, ", "))
	w("- Neighbourhood limited to the in-scope package(s): %s.\n", strings.Join(pkgs, ", "))
	w("- Everything else is omitted. To pull more in, re-run with another `-scope path[:Symbol]`.\n")

	fmt.Print(b.String())
}

// ---- helpers ----

func loadAnalysis(path string) (*analysis, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var a analysis
	if err := json.Unmarshal(b, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func splitScope(s string) (path, sym string) {
	// Split on the LAST colon so Windows-y paths (unlikely here) still work;
	// a symbol never contains a colon.
	i := strings.LastIndex(s, ":")
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i+1:]
}

func locateSymbol(fi fileInfo, name string) (start, end int, kind string, found bool) {
	for _, fn := range fi.Functions {
		if fn.Name == name {
			return fn.StartLine, fn.EndLine, "func", true
		}
	}
	for _, td := range fi.Types {
		if td.Name == name {
			return td.StartLine, td.EndLine, td.Kind, true
		}
	}
	return 0, 0, "", false
}

func symbolSignature(fi fileInfo, name string) string {
	for _, fn := range fi.Functions {
		if fn.Name == name {
			return fn.Signature
		}
	}
	for _, td := range fi.Types {
		if td.Name == name {
			return typeSignature(td)
		}
	}
	return name
}

func fileSignatures(fi fileInfo) string {
	var lines []string
	for _, fn := range fi.Functions {
		lines = append(lines, fn.Signature)
	}
	for _, td := range fi.Types {
		lines = append(lines, typeSignature(td))
	}
	sort.Strings(lines)
	return "```go\n" + strings.Join(lines, "\n") + "\n```"
}

func typeSignature(td typeDef) string {
	switch td.Kind {
	case "struct":
		parts := make([]string, 0, len(td.Fields))
		for _, f := range td.Fields {
			if f.Name == "" {
				parts = append(parts, f.Type) // embedded
			} else {
				parts = append(parts, f.Name+" "+f.Type)
			}
		}
		return fmt.Sprintf("type %s struct { %s }", td.Name, strings.Join(parts, "; "))
	case "interface":
		parts := make([]string, 0, len(td.Methods))
		for _, m := range td.Methods {
			parts = append(parts, m.Type) // already rendered as MethodSig
		}
		return fmt.Sprintf("type %s interface { %s }", td.Name, strings.Join(parts, "; "))
	default:
		if td.Underlying != "" {
			return fmt.Sprintf("type %s %s", td.Name, td.Underlying)
		}
		return "type " + td.Name
	}
}

// readLines returns lines [start,end] (1-indexed, inclusive) of root/relPath.
// end == -1 means to end of file.
func readLines(root, relPath string, start, end int) (string, error) {
	b, err := os.ReadFile(filepath.Join(root, relPath))
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(b), "\n")
	if start < 1 {
		start = 1
	}
	if end == -1 || end > len(lines) {
		end = len(lines)
	}
	if start > len(lines) {
		return "", fmt.Errorf("start line %d beyond file length %d", start, len(lines))
	}
	return strings.Join(lines[start-1:end], "\n"), nil
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
