// Analyse walks a Go source tree and returns the structural summary defined by
// Output: per file the package, imports, function/method signatures (with the
// callee names for a name-matched call graph), and struct/interface/type
// declarations with line ranges.
//
// This is the layer-1 parsing primitive. The analyser CLI is a thin wrapper
// over it, and the in-cluster analyser adapter imports the same function — so
// the harness and production parse identically. It is Go-only by design; a
// non-Go producer fills the same Output contract behind the analyser adapter.
//
// Skips: vendor/, testdata/, hidden dirs (starting with "."), and *_test.go.
package analysis

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Analyse walks root and returns the structural summary. A walk error (e.g. a
// missing root) is returned; per-file parse errors are recorded on the file's
// ParseError and do not abort the walk.
func Analyse(root string) (Output, error) {
	out := Output{
		Root:        root,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := info.Name()
			if base == "vendor" || base == "testdata" || (strings.HasPrefix(base, ".") && base != ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, rerr := filepath.Rel(root, path)
		if rerr != nil {
			rel = path
		}
		out.Files = append(out.Files, analyseFile(path, rel))
		return nil
	})
	if err != nil {
		return out, err
	}

	sort.Slice(out.Files, func(i, j int) bool { return out.Files[i].Path < out.Files[j].Path })
	out.FileCount = len(out.Files)
	return out, nil
}

func analyseFile(path, rel string) FileInfo {
	fi := FileInfo{Path: rel}
	fset := token.NewFileSet()
	src, err := os.ReadFile(path)
	if err != nil {
		fi.ParseError = err.Error()
		return fi
	}
	f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		// Record the error but keep whatever the parser recovered.
		fi.ParseError = err.Error()
		if f == nil {
			return fi
		}
	}
	fi.Package = f.Name.Name

	for _, imp := range f.Imports {
		i := Import{Path: strings.Trim(imp.Path.Value, `"`)}
		if imp.Name != nil {
			i.Alias = imp.Name.Name
		}
		fi.Imports = append(fi.Imports, i)
	}

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			fi.Functions = append(fi.Functions, funcDef(fset, d))
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					fi.Types = append(fi.Types, typeDef(fset, d, ts))
				}
			}
		}
	}
	return fi
}

func funcDef(fset *token.FileSet, d *ast.FuncDecl) FuncDef {
	fn := FuncDef{
		Name:      d.Name.Name,
		Exported:  d.Name.IsExported(),
		StartLine: fset.Position(d.Pos()).Line,
		EndLine:   fset.Position(d.End()).Line,
		Doc:       firstDocLine(d.Doc),
	}
	if d.Recv != nil && len(d.Recv.List) > 0 {
		r := d.Recv.List[0]
		fn.Receiver = &Param{Name: fieldName(r), Type: exprString(fset, r.Type)}
	}
	fn.Params = fieldList(fset, d.Type.Params)
	fn.Results = fieldList(fset, d.Type.Results)
	fn.Signature = renderFuncSig(fn)
	fn.Calls = callsIn(d)
	return fn
}

// callsIn returns the distinct callee names invoked in a function body.
// Name-based, not type-resolved: `Bar()` -> "Bar", `x.Bar()` / `pkg.Bar()` -> "Bar".
// This is enough for a name-matched call-graph slice (cheap; some false matches).
func callsIn(d *ast.FuncDecl) []string {
	if d.Body == nil {
		return nil
	}
	set := map[string]bool{}
	ast.Inspect(d.Body, func(n ast.Node) bool {
		ce, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		switch fun := ce.Fun.(type) {
		case *ast.Ident:
			set[fun.Name] = true
		case *ast.SelectorExpr:
			set[fun.Sel.Name] = true
		}
		return true
	})
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func typeDef(fset *token.FileSet, d *ast.GenDecl, ts *ast.TypeSpec) TypeDef {
	td := TypeDef{
		Name:      ts.Name.Name,
		Exported:  ts.Name.IsExported(),
		StartLine: fset.Position(ts.Pos()).Line,
		EndLine:   fset.Position(ts.End()).Line,
	}
	// Doc may sit on the GenDecl (for a single-spec block) or the spec.
	if ts.Doc != nil {
		td.Doc = firstDocLine(ts.Doc)
	} else if d.Doc != nil && len(d.Specs) == 1 {
		td.Doc = firstDocLine(d.Doc)
	}

	switch t := ts.Type.(type) {
	case *ast.StructType:
		td.Kind = "struct"
		if t.Fields != nil {
			for _, fld := range t.Fields.List {
				typ := exprString(fset, fld.Type)
				tag := ""
				if fld.Tag != nil {
					tag = fld.Tag.Value
				}
				if len(fld.Names) == 0 { // embedded
					td.Fields = append(td.Fields, Field{Type: typ, Tag: tag})
				}
				for _, n := range fld.Names {
					td.Fields = append(td.Fields, Field{Name: n.Name, Type: typ, Tag: tag})
				}
			}
		}
	case *ast.InterfaceType:
		td.Kind = "interface"
		if t.Methods != nil {
			for _, m := range t.Methods.List {
				if len(m.Names) == 0 { // embedded interface
					td.Methods = append(td.Methods, Param{Type: exprString(fset, m.Type)})
					continue
				}
				sig := exprString(fset, m.Type) // func(...) (...)
				for _, n := range m.Names {
					td.Methods = append(td.Methods, Param{Name: n.Name, Type: n.Name + strings.TrimPrefix(sig, "func")})
				}
			}
		}
	default:
		td.Kind = "alias"
		td.Underlying = exprString(fset, ts.Type)
	}
	return td
}

func fieldList(fset *token.FileSet, fl *ast.FieldList) []Param {
	var out []Param
	if fl == nil {
		return out
	}
	for _, f := range fl.List {
		typ := exprString(fset, f.Type)
		if len(f.Names) == 0 {
			out = append(out, Param{Type: typ})
			continue
		}
		for _, n := range f.Names {
			out = append(out, Param{Name: n.Name, Type: typ})
		}
	}
	return out
}

func fieldName(f *ast.Field) string {
	if len(f.Names) > 0 {
		return f.Names[0].Name
	}
	return ""
}

// exprString renders an AST expression (a type) back to its source form.
func exprString(fset *token.FileSet, e ast.Expr) string {
	if e == nil {
		return ""
	}
	var b bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces, Tabwidth: 1}
	if err := cfg.Fprint(&b, fset, e); err != nil {
		return "?"
	}
	return b.String()
}

func renderFuncSig(fn FuncDef) string {
	var b strings.Builder
	b.WriteString("func ")
	if fn.Receiver != nil {
		b.WriteString("(")
		if fn.Receiver.Name != "" {
			b.WriteString(fn.Receiver.Name + " ")
		}
		b.WriteString(fn.Receiver.Type)
		b.WriteString(") ")
	}
	b.WriteString(fn.Name)
	b.WriteString("(")
	b.WriteString(joinParams(fn.Params))
	b.WriteString(")")
	switch len(fn.Results) {
	case 0:
	case 1:
		if fn.Results[0].Name == "" {
			b.WriteString(" " + fn.Results[0].Type)
		} else {
			b.WriteString(" (" + joinParams(fn.Results) + ")")
		}
	default:
		b.WriteString(" (" + joinParams(fn.Results) + ")")
	}
	return b.String()
}

func joinParams(ps []Param) string {
	parts := make([]string, 0, len(ps))
	for _, p := range ps {
		if p.Name != "" {
			parts = append(parts, p.Name+" "+p.Type)
		} else {
			parts = append(parts, p.Type)
		}
	}
	return strings.Join(parts, ", ")
}

func firstDocLine(cg *ast.CommentGroup) string {
	if cg == nil {
		return ""
	}
	text := strings.TrimSpace(cg.Text())
	if text == "" {
		return ""
	}
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		return strings.TrimSpace(text[:i])
	}
	return text
}
