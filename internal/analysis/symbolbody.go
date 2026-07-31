// internal/analysis/symbolbody.go  (package analysis)
//
// ReadSymbolBody is the single place a "path:Symbol" (or "path") scope entry is
// turned into source text: by SLICING the analyser's already-recorded line span
// out of the file on disk — never re-parsing.
//
// CITATIONS CORRECTED 2026-07-31 (bugs_open/145): this header used to name
// `cmd/bundle` three times as the other user and the byte-for-byte reference.
// **There is no `cmd/bundle` in this repo, and never was.** The tool is
// contextkit's `cmd/assembler`, which lives in ARCHIVED reference code under its
// own go.mod (docs019_…/go_files/contextkit/) and is therefore not in this
// module's build — `go list ./...` yields 0 contextkit packages. So the ONLY live
// caller is the chassis's diagnose_assemble_bundle action (:201), whose old
// inline readSymbolBody stub this replaced (the stub had no analyser Output; the
// action has one under repo_analysis). The "collapse cmd/bundle's equivalent
// inline onto this function" note that used to close this header was stale in the
// other direction: that collapse HAPPENED and was proven byte-identical — see
// concept register CTXK-002.
//
// Convention — VERIFIED byte-for-byte against cmd/assembler's own output: the
// body is the file lines [StartLine, EndLine] INCLUSIVE, 1-indexed, exactly as
// the analyser records them. StartLine is the `func`/`type` keyword line
// (d.Pos()), so the preceding doc comment is NOT included; EndLine is the closing
// brace (d.End()). A scope entry with no ":Symbol" returns the whole file
// (matching how cmd/assembler renders a whole-file scope, leading comments
// included) — but ONLY for a file the analyser parsed; see the boundary comment
// on ReadSymbolBody itself.
package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadSymbolBody returns the source text of `symbol` within `root`.
//
//   - "path/to/file.go:Name" -> the func/method/type named Name in that file,
//     sliced from the analyser's StartLine..EndLine.
//   - "path/to/file.go"      -> the whole file.
//
// `out` is the analyser Output that ALREADY carries the line spans — read it
// from repo_analysis rather than re-parsing. `root` is the checked-out repo the
// Output was produced from (out.Root); bodies are read from there.
//
// A method given as just its name ("Neighbourhood") resolves to the first
// matching FuncDef. A receiver-qualified form ("AnalysisCallGraph.Neighbourhood")
// disambiguates when two types share a method name in one file.
func ReadSymbolBody(root string, out Output, symbol string) (string, error) {
	pathPart, namePart := splitSymbol(symbol)
	if pathPart == "" {
		return "", fmt.Errorf("ReadSymbolBody: empty path in symbol %q", symbol)
	}

	// THE BOUNDARY (bugs_open/145). The analyser Output — not the disk — decides
	// what this function will read, for BOTH branches, and it decides BEFORE the
	// read. Until 2026-07-31 this check sat after os.ReadFile and after the
	// whole-file return, so a scope entry naming any committed file (`.env`-shaped
	// files, bugs_open/*.md, fixtures) got that whole file back, and the consumer
	// rendered it to the verdicter inside a ```go fence labelled "In-scope code".
	//
	// Why Output membership and not an ".go" extension test: Output IS the
	// analyser's inclusion rule (analyse.go:80-99 — Go, non-test, minus vendor/,
	// testdata/, download-duplicates and excluded() paths), so an extension test
	// would be a second, drifting copy of it that still admitted a vendored or
	// excluded file. Nothing here to keep in step, and if the analyser ever learns
	// another language this boundary widens with it, correctly.
	//
	// This is not a new invariant — it is the one the ORIGINAL caller had and the
	// chassis port dropped: contextkit's cmd/assembler resolves `byPath[path]` and
	// skips with "path not found in analysis" before it ever calls this function
	// (docs019_…/go_files/contextkit/cmd/assembler/main.go:178-200). Enforcing it
	// HERE is what stops the next producer losing it again — the whole-file branch
	// is reachable by an LLM (the verdict's next_scope), and the bundle text at
	// diagnose_assemble_bundle_action.go:597 explicitly invites bare file paths.
	//
	// Side effect worth knowing: path traversal is closed too. `../../etc/passwd`
	// cannot be in Output, so filepath.Join can no longer be walked out of `root`.
	fi := findFile(out, pathPart)
	if fi == nil {
		return "", fmt.Errorf("ReadSymbolBody: %q not in analysis (no FileInfo for that path) — bodies are read only for files the analyser parsed", pathPart)
	}

	abs := filepath.Join(root, filepath.FromSlash(pathPart))
	src, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("ReadSymbolBody: read %s: %w", abs, err)
	}

	// Whole-file scope entry — an ANALYSED source file, guaranteed by the check
	// above. Kept deliberately: the bundle advertises it (see above), so removing
	// the branch would break a documented next_scope capability.
	if namePart == "" {
		return string(src), nil
	}

	start, end, ok := spanOf(fi, namePart)
	if !ok {
		return "", fmt.Errorf("ReadSymbolBody: symbol %q not found in %s", namePart, pathPart)
	}
	return SliceLines(src, start, end)
}

// splitSymbol splits on the LAST colon: "path:Name" -> ("path","Name");
// "path" -> ("path",""). Last-colon so a path that itself contains a colon is
// not mangled (paths here are slash-relative, so this is belt-and-braces).
func splitSymbol(symbol string) (path, name string) {
	i := strings.LastIndex(symbol, ":")
	if i < 0 {
		return symbol, ""
	}
	return symbol[:i], symbol[i+1:]
}

func findFile(out Output, slashRelPath string) *FileInfo {
	for i := range out.Files {
		if out.Files[i].Path == slashRelPath {
			return &out.Files[i]
		}
	}
	return nil
}

// spanOf returns the StartLine/EndLine for a func/method/type named `name` in
// fi. Functions are searched first, then types. `name` may be receiver-qualified
// ("Type.Method") to disambiguate a method-name collision; a bare name matches
// a func/method by name (first wins) or, failing that, a type.
func spanOf(fi *FileInfo, name string) (start, end int, ok bool) {
	wantRecv, wantName := "", name
	if i := strings.LastIndex(name, "."); i >= 0 {
		wantRecv, wantName = name[:i], name[i+1:]
	}
	for _, fn := range fi.Functions {
		if fn.Name != wantName {
			continue
		}
		if wantRecv != "" && receiverType(fn) != wantRecv {
			continue
		}
		return fn.StartLine, fn.EndLine, true
	}
	if wantRecv == "" { // a bare name can also be a type
		for _, td := range fi.Types {
			if td.Name == wantName {
				return td.StartLine, td.EndLine, true
			}
		}
	}
	return 0, 0, false
}

// receiverType is the receiver's base type name: "*AnalysisCallGraph" ->
// "AnalysisCallGraph", so a receiver-qualified query matches pointer or value.
func receiverType(fn FuncDef) string {
	if fn.Receiver == nil {
		return ""
	}
	return strings.TrimPrefix(fn.Receiver.Type, "*")
}

// SliceLines returns src lines [start,end] inclusive, 1-indexed — the verified
// cmd/bundle convention. Bounds are clamped so a stale/incorrect span reports an
// error or trims rather than panicking.
//
// EXPORTED 2026-07-27 for index_code_symbols, which now stores each symbol's
// body at index time (D11 layer 1, council 18fe4035). It is exported rather than
// re-implemented so ONE function keeps owning the [start,end] inclusive 1-indexed
// convention: the indexer that WRITES a body and ReadSymbolBody, which reads one
// on demand, must agree byte-for-byte or a stored body and a freshly sliced one
// silently differ. Round 2 of that council proposed EXTRACTING a shared slicer;
// prior_art_librarian pointed out this one already existed, which is why the edit
// is an export and not a new function.
func SliceLines(src []byte, start, end int) (string, error) {
	if start <= 0 || end < start {
		return "", fmt.Errorf("SliceLines: bad span [%d,%d]", start, end)
	}
	lines := strings.Split(string(src), "\n")
	if start > len(lines) {
		return "", fmt.Errorf("SliceLines: span start %d past end of file (%d lines)", start, len(lines))
	}
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start-1:end], "\n"), nil
}
