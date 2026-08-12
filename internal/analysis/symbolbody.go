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
	"sort"
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
// matching FuncDef. A receiver-qualified form disambiguates when two types share
// a method name in one file, and BOTH spellings are accepted: the code index's
// canonical "(*AnalysisCallGraph).Neighbourhood" (which is what actually reaches
// this function — see splitReceiver and bugs_open/261) and the bare
// "AnalysisCallGraph.Neighbourhood". Package-level var/const resolve by name.
func ReadSymbolBody(root string, out Output, symbol string) (string, error) {
	pathPart, namePart := SplitSymbol(symbol)
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
	fi := FindFile(out, pathPart)
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

// SplitSymbol splits on the LAST colon: "path:Name" -> ("path","Name");
// "path" -> ("path",""). Last-colon so a path that itself contains a colon is
// not mangled (paths here are slash-relative, so this is belt-and-braces).
//
// EXPORTED 2026-08-03 for diagnose_code_lookup's symbol arm (bugs_open/163),
// which was tokenising a whole "path:Symbol" query and AND-ing every token —
// path fragments included — against the code_symbols.symbol column, a column
// that never holds a path. It is exported rather than re-implemented for the
// same reason SliceLines was below: one function keeps owning the convention,
// and this handle format now has three producers (scopeFromCodeResults,
// resolveScopeEntries, the landmine-verifier's derive_checks prompt) whose
// output must all parse the same way.
//
// COLLAPSED 2026-08-04 (bugs_closed/189, slug
// siblingsignatures_hand_rolls_the_path_symbol_split): the third copy — inline
// in siblingSignatures, splitting on `i > 0` rather than `i >= 0`, so a leading
// colon read as no-symbol — now calls this function, and there are ZERO
// in-build duplicates of the grammar left. Its one divergent input (":Foo")
// was proven unobservable there before the collapse, so the edge was decided on
// its merits rather than preserved by mimicry: an entry with no path names no
// file, so that caller skips it, which is the judgement ReadSymbolBody above
// already makes on the same scope slice. The archived contextkit copy keeps its
// unexported splitSymbol — it is in no build, and that drift is recorded in
// CTXK-002 rather than fixed here.
func SplitSymbol(symbol string) (path, name string) {
	i := strings.LastIndex(symbol, ":")
	if i < 0 {
		return symbol, ""
	}
	return symbol[:i], symbol[i+1:]
}

// SymbolSize is one addressable symbol and the length, in the SAME units the
// consumer's caps count (Go's len over the body text), of what ReadSymbolBody
// would return for it.
type SymbolSize struct {
	Symbol string // the "path:Name" handle, in the spelling next_scope resolves
	Chars  int
}

// SymbolSizes ranks every symbol the analyser recorded for fi by body size,
// LARGEST FIRST, slicing them out of `src` — the file text the caller already
// holds. Ties break on the handle so the order is deterministic.
//
// bugs_open/267. A bundle that has just refused an over-budget whole file needs
// to tell the model what it COULD ask for instead; without sizes that advice is
// another guess, and a guess is what cost the loop an iteration in the first
// place.
//
// Sizes are produced by calling SliceLines rather than by re-deriving its span
// arithmetic, so an advertised size is BY CONSTRUCTION the number the consumer's
// cap will later compare against and the two cannot drift. That costs one split
// of the file per symbol; this runs once per over-budget whole-file marker, which
// is rare by definition, and correctness-by-construction is worth more here than
// the allocations.
//
// A symbol whose span does not slice (stale or zero span) is OMITTED rather than
// reported as zero: a confident wrong size is worse than a missing row when the
// number is the entire point of the row.
//
// Methods are keyed in the CANONICAL "(*Recv).Name" spelling that
// code_symbols_actions.go writes and splitReceiver accepts, so a handle printed
// from here still resolves when it comes back as next_scope — a bare method name
// would be ambiguous in a file where two types share one.
func SymbolSizes(fi *FileInfo, src string) []SymbolSize {
	if fi == nil {
		return nil
	}
	b := []byte(src)
	var out []SymbolSize
	add := func(name string, start, end int) {
		body, err := SliceLines(b, start, end)
		if err != nil {
			return
		}
		out = append(out, SymbolSize{Symbol: fi.Path + ":" + name, Chars: len(body)})
	}
	for _, fn := range fi.Functions {
		name := fn.Name
		if fn.Receiver != nil {
			name = "(" + fn.Receiver.Type + ")." + fn.Name
		}
		add(name, fn.StartLine, fn.EndLine)
	}
	for _, td := range fi.Types {
		add(td.Name, td.StartLine, td.EndLine)
	}
	// Package-level var/const, addressable since bugs_open/223 phase 2 and
	// resolvable by spanOf since bugs_open/261 — so they belong in a list whose
	// purpose is "what you may name in next_scope".
	for _, vd := range fi.Values {
		add(vd.Name, vd.StartLine, vd.EndLine)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Chars != out[j].Chars {
			return out[i].Chars > out[j].Chars
		}
		return out[i].Symbol < out[j].Symbol
	})
	return out
}

// FindFile returns the analyser's FileInfo for a slash-relative path, or nil if
// the analyser never parsed that file — which is also the READ BOUNDARY above,
// so "nil" and "not readable here" are the same statement.
//
// EXPORTED 2026-08-12 for bugs_open/267, whose fix must reach a file's recorded
// symbols in order to tell a model what it could ask for instead of an
// over-budget whole file. Exported rather than re-walked out.Files at the call
// site for the same reason SplitSymbol and SliceLines were: this package's own
// header records that the path→FileInfo lookup has already been hand-rolled once
// per consumer and drifted (bugs_closed/189), and a third inline copy is exactly
// that shape again.
func FindFile(out Output, slashRelPath string) *FileInfo {
	for i := range out.Files {
		if out.Files[i].Path == slashRelPath {
			return &out.Files[i]
		}
	}
	return nil
}

// spanOf returns the StartLine/EndLine for a func/method/type/package-level
// value named `name` in fi. Functions are searched first, then types, then
// values. `name` may be receiver-qualified — in EITHER the index's canonical
// "(*Type).Method" spelling or the bare "Type.Method" one, see splitReceiver —
// to disambiguate a method-name collision; a bare name matches a func/method by
// name (first wins) or, failing that, a type or a package-level var/const.
func spanOf(fi *FileInfo, name string) (start, end int, ok bool) {
	wantRecv, wantName := splitReceiver(name)
	for _, fn := range fi.Functions {
		if fn.Name != wantName {
			continue
		}
		if wantRecv != "" && receiverType(fn) != wantRecv {
			continue
		}
		return fn.StartLine, fn.EndLine, true
	}
	if wantRecv == "" { // a bare name can also be a type or a package-level value
		for _, td := range fi.Types {
			if td.Name == wantName {
				return td.StartLine, td.EndLine, true
			}
		}
		// bugs_open/261, second half. The analyser has recorded package-level
		// var/const since bugs_open/223 phase 2 and the indexer writes them as
		// rows, but this reader was never taught the kind, so every one of them
		// answered "symbol not found" — 20 such failures fleet-wide, every one a
		// var or const. Searched LAST so no existing func/type resolution moves.
		for _, vd := range fi.Values {
			if vd.Name == wantName {
				return vd.StartLine, vd.EndLine, true
			}
		}
	}
	return 0, 0, false
}

// splitReceiver splits a possibly receiver-qualified symbol into the receiver's
// BASE TYPE NAME and the method name, accepting every spelling this estate's
// producers actually emit:
//
//	"(*SagaCoordinator).applyResponseToState" -> ("SagaCoordinator", "apply…")
//	"(Greeter).Greet"                         -> ("Greeter", "Greet")
//	"*Helper.Greet" / "Helper.Greet"          -> ("Helper", "Greet")
//	"Hello"                                   -> ("", "Hello")
//
// bugs_open/261. The parenthesised form is not an exotic input — it is the
// CANONICAL one: code_symbols_actions.go:598 writes every method as
// "(" + Receiver.Type + ")." + Name, diagnose_assemble_bundle's
// scopeFromCodeResults concatenates that column straight into a scope entry, and
// the bundle then RENDERS index rows in that spelling and invites the model to
// name them in next_scope. So the estate's own tooling and its own LANDMINES
// entry ("name it `(*Receiver).Method`") both teach the one spelling this
// function used to reject: the raw prefix "(*SagaCoordinator)" was compared
// against receiverType()'s "SagaCoordinator" and could never match, which cost
// 301 unreadable function bodies across 44 diagnosis runs.
//
// Normalising HERE rather than at the producers is deliberate: there are two
// independent producers of a scope entry (the code-search fallback, and an LLM's
// next_scope copied from whatever the bundle showed it), and only one of them is
// code we can fix. This is the same "one function keeps owning the convention"
// judgement that exported SplitSymbol and SliceLines rather than re-implementing
// them — the grammar has one owner, and this is it.
//
// The receiver is REQUIRED to match once given (see receiverType's caller above):
// widening the spelling must not widen the MATCH, or "(*Nope).Greet" would
// silently return some other type's Greet.
func splitReceiver(name string) (recv, sym string) {
	i := strings.LastIndex(name, ".")
	if i < 0 {
		return "", name
	}
	recv, sym = name[:i], name[i+1:]
	// Order matters: parentheses outside, pointer star inside — "(*T)" -> "T".
	recv = strings.TrimSuffix(strings.TrimPrefix(recv, "("), ")")
	recv = strings.TrimPrefix(recv, "*")
	return recv, sym
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
