// internal/analysis/symbolbody.go  (package analysis)
//
// ReadSymbolBody is the single place a "path:Symbol" (or "path") scope entry is
// turned into source text: by SLICING the analyser's already-recorded line span
// out of the file on disk — never re-parsing. It is the slicer the bundle
// assembler (cmd/bundle) and the chassis diagnose_assemble_bundle action both
// use, so a symbol's body renders identically wherever it is gathered. The
// chassis action's old readSymbolBody stub is replaced by a call to this (it
// already has the analyser Output under repo_analysis; the stub did not).
//
// Convention — VERIFIED byte-for-byte against cmd/bundle's own output: the body
// is the file lines [StartLine, EndLine] INCLUSIVE, 1-indexed, exactly as the
// analyser records them. StartLine is the `func`/`type` keyword line (d.Pos()),
// so the preceding doc comment is NOT included; EndLine is the closing brace
// (d.End()). A scope entry with no ":Symbol" returns the whole file (matching
// how cmd/bundle renders a whole-file scope, leading comments included).
//
// Reuse note: this is intended as the ONE slicer. cmd/bundle currently has an
// equivalent inline; collapse it onto this function (and confirm cmd/bundle's
// symbol resolution matches spanOf for any method-name-collision edge cases)
// once both are in the same tree.
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

	abs := filepath.Join(root, filepath.FromSlash(pathPart))
	src, err := os.ReadFile(abs)
	if err != nil {
		return "", fmt.Errorf("ReadSymbolBody: read %s: %w", abs, err)
	}

	// Whole-file scope entry.
	if namePart == "" {
		return string(src), nil
	}

	fi := findFile(out, pathPart)
	if fi == nil {
		return "", fmt.Errorf("ReadSymbolBody: %q not in analysis (no FileInfo for that path)", pathPart)
	}

	start, end, ok := spanOf(fi, namePart)
	if !ok {
		return "", fmt.Errorf("ReadSymbolBody: symbol %q not found in %s", namePart, pathPart)
	}
	return sliceLines(src, start, end)
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

// sliceLines returns src lines [start,end] inclusive, 1-indexed — the verified
// cmd/bundle convention. Bounds are clamped so a stale/incorrect span reports an
// error or trims rather than panicking.
func sliceLines(src []byte, start, end int) (string, error) {
	if start <= 0 || end < start {
		return "", fmt.Errorf("ReadSymbolBody: bad span [%d,%d]", start, end)
	}
	lines := strings.Split(string(src), "\n")
	if start > len(lines) {
		return "", fmt.Errorf("ReadSymbolBody: span start %d past end of file (%d lines)", start, len(lines))
	}
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start-1:end], "\n"), nil
}
