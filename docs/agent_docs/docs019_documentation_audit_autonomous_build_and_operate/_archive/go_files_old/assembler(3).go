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
	"regexp"
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
	Name      string   `json:"name"`
	Signature string   `json:"signature"`
	Doc       string   `json:"doc"`
	Calls     []string `json:"calls"`
	StartLine int      `json:"start_line"`
	EndLine   int      `json:"end_line"`
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
		maxNeighbour = flag.Int("max-neighbour", 60, "cap on neighbourhood signatures per group")
		neighbour    = flag.String("neighbour", "callgraph", "neighbourhood mode: callgraph | package")
	)
	var scopes scopeList
	flag.Var(&scopes, "scope", "in-scope file (path) or symbol (path:Name); repeatable")
	var docs scopeList
	flag.Var(&docs, "doc", "path to an authored doc to include verbatim (debug guide, a 003 section); repeatable")
	var includes scopeList
	flag.Var(&includes, "include", "path to a wiring/shared file to force-include as signatures regardless of the call graph (e.g. registry.go); repeatable")
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
	index := map[string][]symRef{} // name -> declarations (funcs and types)
	var allFuncs []symRef          // for the caller scan
	for _, f := range an.Files {
		byPath[f.Path] = f
		byPackage[f.Package] = append(byPackage[f.Package], f)
		for _, fn := range f.Functions {
			r := symRef{Name: fn.Name, Path: f.Path, Pkg: f.Package, Kind: "func", Signature: fn.Signature, Calls: fn.Calls, IsFunc: true}
			index[fn.Name] = append(index[fn.Name], r)
			allFuncs = append(allFuncs, r)
		}
		for _, td := range f.Types {
			r := symRef{Name: td.Name, Path: f.Path, Pkg: f.Package, Kind: td.Kind, Signature: typeSignature(td)}
			index[td.Name] = append(index[td.Name], r)
		}
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
	var inScopeRefs []symRef
	for _, sc := range scopes {
		path, sym := splitScope(sc)
		fi, ok := byPath[path]
		if !ok {
			w("> scope %q: path not found in analysis — skipped\n\n", sc)
			continue
		}
		inScopePkgs[fi.Package] = true

		if sym == "" {
			// whole file: every decl in it is in scope
			for _, fn := range fi.Functions {
				inScopeSyms[seen{path, fn.Name}] = true
				inScopeRefs = append(inScopeRefs, symRef{Name: fn.Name, Path: path, Pkg: fi.Package, Kind: "func", Signature: fn.Signature, Calls: fn.Calls, IsFunc: true})
			}
			for _, td := range fi.Types {
				inScopeSyms[seen{path, td.Name}] = true
				inScopeRefs = append(inScopeRefs, symRef{Name: td.Name, Path: path, Pkg: fi.Package, Kind: td.Kind, Signature: typeSignature(td)})
			}
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
		inScopeRefs = append(inScopeRefs, refFor(fi, sym))
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
	if *neighbour == "package" {
		// whole in-scope package(s) as signatures
		if len(inScopePkgs) == 0 {
			w("_none — no in-scope package resolved_\n\n")
		}
		for _, pkg := range sortedKeys(inScopePkgs) {
			w("### package `%s`\n\n", pkg)
			var lines []string
			for _, fi := range byPackage[pkg] {
				for _, fn := range fi.Functions {
					if inScopeSyms[seen{fi.Path, fn.Name}] {
						continue
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
			shown, omitted := capLines(lines, *maxNeighbour)
			w("```go\n%s\n```\n", strings.Join(shown, "\n"))
			if omitted > 0 {
				w("_…and %d more in this package (ask to include)_\n", omitted)
			}
			w("\n")
		}
	} else {
		// call-graph slice: callees, callers, types used by the in-scope symbols
		inName := map[string]bool{}
		for _, r := range inScopeRefs {
			inName[r.Name] = true
		}
		callees := map[string]symRef{}
		for _, r := range inScopeRefs {
			if !r.IsFunc {
				continue
			}
			for _, cn := range r.Calls {
				if inName[cn] {
					continue
				}
				for _, d := range index[cn] {
					callees[d.Path+"|"+d.Name] = d
				}
			}
		}
		callers := map[string]symRef{}
		for _, f := range allFuncs {
			if inName[f.Name] {
				continue
			}
			for _, cn := range f.Calls {
				if inName[cn] {
					callers[f.Path+"|"+f.Name] = f
					break
				}
			}
		}
		typesUsed := map[string]symRef{}
		for _, r := range inScopeRefs {
			for _, tn := range extractIdents(r.Signature) {
				if inName[tn] {
					continue
				}
				for _, d := range index[tn] {
					if d.Kind == "struct" || d.Kind == "interface" || d.Kind == "alias" {
						typesUsed[d.Path+"|"+d.Name] = d
					}
				}
			}
		}
		renderGroup(w, "Calls (callees)", callees, *maxNeighbour)
		renderGroup(w, "Called by (callers)", callers, *maxNeighbour)
		renderGroup(w, "Types used", typesUsed, *maxNeighbour)
		if len(callees)+len(callers)+len(typesUsed) == 0 {
			w("_call-graph found no internal neighbours (the in-scope symbols may call only stdlib/external, or nothing). Try `-neighbour package` for the surrounding package._\n\n")
		}
		w("_Note: name-matched call graph — a name shared across packages can show extra candidates, and calls through interfaces aren't resolved. Coarse but tight; widen with `-scope` or `-neighbour package` if something's missing._\n\n")
	}

	// ---- forced includes (wiring/shared files the call graph can't reach) ----
	if len(includes) > 0 {
		forced := map[string]symRef{}
		for _, inc := range includes {
			fi, ok := byPath[inc]
			if !ok {
				w("> include %q: path not found in analysis — skipped\n\n", inc)
				continue
			}
			for _, fn := range fi.Functions {
				if inScopeSyms[seen{inc, fn.Name}] {
					continue
				}
				forced[fi.Path+"|"+fn.Name] = symRef{Name: fn.Name, Path: fi.Path, Pkg: fi.Package, Kind: "func", Signature: fn.Signature}
			}
			for _, td := range fi.Types {
				if inScopeSyms[seen{inc, td.Name}] {
					continue
				}
				forced[fi.Path+"|"+td.Name] = symRef{Name: td.Name, Path: fi.Path, Pkg: fi.Package, Kind: td.Kind, Signature: typeSignature(td)}
			}
		}
		renderGroup(w, "Forced includes (wiring/shared — files named with -include that the call graph can't reach)", forced, 1000)
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
	if *neighbour == "package" {
		w("- Neighbourhood mode: whole in-scope package(s): %s.\n", strings.Join(sortedKeys(inScopePkgs), ", "))
	} else {
		w("- Neighbourhood mode: call-graph slice (callees, callers, types used) of the in-scope symbols.\n")
	}
	w("- Everything else is omitted. To pull more in, re-run with another `-scope path[:Symbol]`, or `-neighbour package` for the whole package.\n")
	if len(includes) > 0 {
		w("- Force-included wiring/shared files: %s.\n", strings.Join(includes, ", "))
	}

	fmt.Print(b.String())
}

// ---- helpers ----

// symRef is one declaration (function or type) located in the repo.
type symRef struct {
	Name      string
	Path      string
	Pkg       string
	Kind      string // func | struct | interface | alias
	Signature string
	Calls     []string
	IsFunc    bool
}

// refFor builds a symRef for a named decl in a file (used for in-scope symbols).
func refFor(fi fileInfo, name string) symRef {
	for _, fn := range fi.Functions {
		if fn.Name == name {
			return symRef{Name: name, Path: fi.Path, Pkg: fi.Package, Kind: "func", Signature: fn.Signature, Calls: fn.Calls, IsFunc: true}
		}
	}
	for _, td := range fi.Types {
		if td.Name == name {
			return symRef{Name: name, Path: fi.Path, Pkg: fi.Package, Kind: td.Kind, Signature: typeSignature(td)}
		}
	}
	return symRef{Name: name, Path: fi.Path, Pkg: fi.Package}
}

var identRE = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)

// extractIdents pulls bare identifiers out of a rendered signature, so type
// names used in params/results/fields can be matched against the type index.
func extractIdents(s string) []string {
	seen := map[string]bool{}
	var out []string
	for _, m := range identRE.FindAllString(s, -1) {
		if seen[m] {
			continue
		}
		seen[m] = true
		out = append(out, m)
	}
	return out
}

// renderGroup prints one labelled group of neighbour signatures, each annotated
// with its file path, sorted by name, capped.
func renderGroup(w func(string, ...any), label string, set map[string]symRef, cap int) {
	if len(set) == 0 {
		return
	}
	refs := make([]symRef, 0, len(set))
	for _, r := range set {
		refs = append(refs, r)
	}
	sort.Slice(refs, func(i, j int) bool {
		if refs[i].Name != refs[j].Name {
			return refs[i].Name < refs[j].Name
		}
		return refs[i].Path < refs[j].Path
	})
	omitted := 0
	if len(refs) > cap {
		omitted = len(refs) - cap
		refs = refs[:cap]
	}
	w("**%s**\n\n```go\n", label)
	for _, r := range refs {
		w("%s  // %s\n", r.Signature, r.Path)
	}
	w("```\n")
	if omitted > 0 {
		w("_…and %d more (widen with -scope or -neighbour package)_\n", omitted)
	}
	w("\n")
}

func capLines(lines []string, cap int) (shown []string, omitted int) {
	if len(lines) > cap {
		return lines[:cap], len(lines) - cap
	}
	return lines, 0
}

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
