// Package analysis defines the contract the analyser emits and the assembler,
// embed, and resolve_targets consume. It lives here so the shape has ONE source
// of truth, instead of being re-declared (and drifting) in each tool. Consumers
// alias these names locally (e.g. `type funcDef = analysis.FuncDef`) so their
// bodies are unchanged; the structs themselves are defined only here.
package analysis

type Output struct {
	Root        string     `json:"root"`
	GeneratedAt string     `json:"generated_at"`
	FileCount   int        `json:"file_count"`
	Files       []FileInfo `json:"files"`
}

type FileInfo struct {
	Path      string    `json:"path"` // relative to root
	Package   string    `json:"package"`
	Imports   []Import  `json:"imports"`
	Functions []FuncDef `json:"functions"`
	Types     []TypeDef `json:"types"`
	// Values are package-level var and const declarations (bugs_open/223 phase 2).
	// Additive and omitempty, so an older consumer of this JSON is unaffected and
	// an older analyser's output still unmarshals here as an empty slice.
	Values     []ValueDef `json:"values,omitempty"`
	ParseError string     `json:"parse_error,omitempty"`
}

type Import struct {
	Alias string `json:"alias,omitempty"`
	Path  string `json:"path"`
}

type Param struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
}

type FuncDef struct {
	Name      string   `json:"name"`
	Receiver  *Param   `json:"receiver,omitempty"` // nil for plain functions
	Params    []Param  `json:"params"`
	Results   []Param  `json:"results"`
	Exported  bool     `json:"exported"`
	Doc       string   `json:"doc,omitempty"`   // first line of doc comment
	Signature string   `json:"signature"`       // rendered, paste-ready
	Calls     []string `json:"calls,omitempty"` // distinct callee names in the body (name-based)
	StartLine int      `json:"start_line"`
	EndLine   int      `json:"end_line"`
}

type Field struct {
	Name string `json:"name,omitempty"`
	Type string `json:"type"`
	Tag  string `json:"tag,omitempty"`
}

// ValueDef is ONE package-level var or const name.
//
// bugs_open/223 phase 2. Until this existed the analyser dropped `token.VAR` and
// `token.CONST` at the AST walk, so `code_symbols` held no row of either kind —
// although its own CHECK constraint permits both and diagnose_code_lookup's
// codeKindList already treats them as code. The cost was not theoretical: asked
// about `metaCommentaryPatterns` (a live `var` at validate_page_content.go:1229)
// the landmine-verifier answered "no longer resolves as a standalone symbol
// (possibly inlined or renamed)", and the diagnosis loop stopped at UNVERIFIABLE
// on `DeployImageAssetInputSpec` naming the very declaration it could not see
// (bugs_open/231). Both are the same gap: every USE of a package-level value was
// findable and the DECLARATION was not.
//
// ONE ENTRY PER NAME, not per spec: `var a, b = f()` is two addressable symbols,
// and a lookup for `b` must find it. They share a line span, which is correct —
// the declaration really is one region of source.
type ValueDef struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"` // "var" | "const" — both already in the code_symbols CHECK set
	Exported bool   `json:"exported"`
	Doc      string `json:"doc,omitempty"`
	// Type is the DECLARED type when the source states one ("var x []byte"), and
	// empty when it is inferred ("var x = f()"). Deliberately not inferred here:
	// this package parses, it does not type-check, and a guessed type in a
	// signature field would be a confident wrong answer of exactly the kind
	// bugs_open/223 is about.
	Type      string `json:"type,omitempty"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type TypeDef struct {
	Name       string  `json:"name"`
	Kind       string  `json:"kind"` // "struct" | "interface" | "alias"
	Exported   bool    `json:"exported"`
	Doc        string  `json:"doc,omitempty"`
	Fields     []Field `json:"fields,omitempty"`     // struct
	Methods    []Param `json:"methods,omitempty"`    // interface: Name + rendered signature in Type
	Underlying string  `json:"underlying,omitempty"` // alias/defined type
	StartLine  int     `json:"start_line"`
	EndLine    int     `json:"end_line"`
}
