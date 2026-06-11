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
	Path       string    `json:"path"` // relative to root
	Package    string    `json:"package"`
	Imports    []Import  `json:"imports"`
	Functions  []FuncDef `json:"functions"`
	Types      []TypeDef `json:"types"`
	ParseError string    `json:"parse_error,omitempty"`
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
