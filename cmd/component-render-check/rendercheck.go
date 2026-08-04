// component-render-check — the OUTPUT-level empty-element check.
//
// bugfix_140 plan item 1 (HANDOFF_2026-08-03_continue_here.md). The template-shape
// lint (check_placeholder_fallbacks.py) can only see fields declared
// on_missing:"skip_field", and can be satisfied by a gate that encloses nothing
// ({{if .v}}{{.v}}{{end}} inside a fixed <td>). This check is immune to both by
// construction, because it never reads declarations at all: for every ACTIVE
// component, for every field its template references — regardless of declaration
// flavour or schema presence — it renders the template through the PRODUCTION
// entry point (actions.RenderTemplate, not a replica of its config) twice: once
// with every referenced field supplied, once with the field under test absent.
// A finding is an empty-element shape (<h1..4></h1..4>, <a ...></a>,
// <img src="">, <td></td>, empty class-bearing block) that APPEARS when the
// field is absent.
//
// The POSITIVE CONTROL is the load-bearing half (the 20/20 harness rule): with
// every field supplied, each field's marker must actually reach the output —
// a gate that over-fires (drops the element even when the datum is there)
// passes any absence-only test, and an element empty even at full data is a
// hardcoded blank, reported separately.
//
// data-runtime-fill is honoured at COMPONENT granularity: a template carrying
// the attribute is deliberately empty at build time and browser-filled
// (check_empty_sections.go's documented first-catch exemption — vonc's
// provocation-card/lobby-grid). Its absence findings are demoted to info.
//
// Exit codes: 0 = ran (findings are the report — calibration mode; the
// fail-the-gate decision is plan item 3, deliberately AFTER this exists),
// 2 = could not load the library (a flake or refusal, NEVER a pass).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template/parse"
	"time"

	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Loading — same route and same flake as the Python lint: the ~2 MB stream
// through kubectl exec truncates intermittently, so three attempts, and the
// final failure exits 2.
// ---------------------------------------------------------------------------

const query = `SELECT COALESCE(jsonb_agg(jsonb_build_object(
  'name', name, 'function', function, 'html_template', html_template)), '[]'::jsonb)
FROM content_components WHERE is_active;`

var psqlArgv = []string{"kubectl", "exec", "-n", "ai-persona-system", "postgres-clients-0", "--",
	"psql", "-U", "clients_user", "-d", "clients_db", "-tAc"}

type component struct {
	Name     string `json:"name"`
	Function string `json:"function"`
	Template string `json:"html_template"`
}

func loadComponents(path string) ([]component, error) {
	var body []byte
	var err error
	if path != "" {
		body, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else {
		var last error
		for attempt := 0; attempt < 3; attempt++ {
			if attempt > 0 {
				fmt.Fprintf(os.Stderr, "fetch attempt %d/3 after: %v\n", attempt+1, last)
				time.Sleep(time.Duration(3*attempt) * time.Second)
			}
			out, cerr := exec.Command(psqlArgv[0], append(psqlArgv[1:], query)...).Output()
			if cerr != nil {
				last = cerr
				continue
			}
			trimmed := []byte(strings.TrimSpace(string(out)))
			var probe interface{}
			if jerr := json.Unmarshal(trimmed, &probe); jerr != nil {
				last = fmt.Errorf("invalid JSON (truncated stream?): %w", jerr)
				continue
			}
			body = trimmed
			break
		}
		if body == nil {
			return nil, fmt.Errorf("could not fetch the component library after 3 attempts: %w", last)
		}
	}
	var comps []component
	if err := json.Unmarshal(body, &comps); err != nil {
		return nil, err
	}
	return comps, nil
}

// ---------------------------------------------------------------------------
// Field analysis — walk the real parse tree with a scope stack, so a field
// referenced inside {{range .items}} is recorded as items[].field, not as a
// top-level "field". parse.SkipFuncCheck lets analysis parse templates whose
// FuncMap belongs to the production renderer.
// ---------------------------------------------------------------------------

type shape struct {
	kind     string // "scalar", "list", "map"
	children map[string]*shape
}

func newShape() *shape { return &shape{kind: "scalar", children: map[string]*shape{}} }

func (s *shape) at(path []string) *shape {
	cur := s
	for _, seg := range path {
		if cur.children[seg] == nil {
			cur.children[seg] = newShape()
		}
		if cur.kind == "scalar" {
			cur.kind = "map"
		}
		cur = cur.children[seg]
	}
	return cur
}

type analysis struct {
	root *shape
	err  error
}

func analyse(tpl string) analysis {
	tr := parse.New("t")
	tr.Mode = parse.SkipFuncCheck
	treeSet := map[string]*parse.Tree{}
	tree, err := tr.Parse(tpl, "{{", "}}", treeSet)
	if err != nil {
		return analysis{err: err}
	}
	a := analysis{root: newShape()}
	vars := map[string][]string{} // $var -> path it aliases
	a.walkList(tree.Root, nil, vars)
	return a
}

func (a *analysis) walkList(l *parse.ListNode, scope []string, vars map[string][]string) {
	if l == nil {
		return
	}
	for _, n := range l.Nodes {
		a.walkNode(n, scope, vars)
	}
}

func (a *analysis) walkNode(n parse.Node, scope []string, vars map[string][]string) {
	switch t := n.(type) {
	case *parse.ActionNode:
		a.walkPipe(t.Pipe, scope, vars)
	case *parse.IfNode:
		a.walkPipe(t.Pipe, scope, vars)
		a.walkList(t.List, scope, vars) // {{if}} does not rebind dot
		a.walkList(t.ElseList, scope, vars)
	case *parse.RangeNode:
		target := a.pipeFieldPath(t.Pipe, scope, vars)
		a.walkPipe(t.Pipe, scope, vars)
		inner := scope
		if target != nil {
			a.root.at(target).kind = "list"
			inner = append(append([]string{}, target...), "[]")
		}
		innerVars := cloneVars(vars)
		// {{range $x := .xs}} -> $x is the element; {{range $i, $x := .xs}} too.
		if decls := t.Pipe.Decl; len(decls) > 0 && target != nil {
			last := decls[len(decls)-1]
			if len(last.Ident) > 0 {
				innerVars[last.Ident[0]] = inner
			}
		}
		a.walkList(t.List, inner, innerVars)
		a.walkList(t.ElseList, scope, vars)
	case *parse.WithNode:
		target := a.pipeFieldPath(t.Pipe, scope, vars)
		a.walkPipe(t.Pipe, scope, vars)
		inner := scope
		if target != nil {
			inner = target
		}
		innerVars := cloneVars(vars)
		if decls := t.Pipe.Decl; len(decls) > 0 && target != nil {
			last := decls[len(decls)-1]
			if len(last.Ident) > 0 {
				innerVars[last.Ident[0]] = target
			}
		}
		a.walkList(t.List, inner, innerVars)
		a.walkList(t.ElseList, scope, vars)
	case *parse.TemplateNode:
		if t.Pipe != nil {
			a.walkPipe(t.Pipe, scope, vars)
		}
	}
}

func (a *analysis) walkPipe(p *parse.PipeNode, scope []string, vars map[string][]string) {
	if p == nil {
		return
	}
	for _, cmd := range p.Cmds {
		for _, arg := range cmd.Args {
			switch t := arg.(type) {
			case *parse.FieldNode:
				a.root.at(append(append([]string{}, scope...), t.Ident...))
			case *parse.VariableNode:
				if base, ok := vars[t.Ident[0]]; ok {
					a.root.at(append(append([]string{}, base...), t.Ident[1:]...))
				}
			case *parse.PipeNode:
				a.walkPipe(t, scope, vars)
			case *parse.ChainNode:
				if f, ok := t.Node.(*parse.FieldNode); ok {
					a.root.at(append(append(append([]string{}, scope...), f.Ident...), t.Field...))
				}
			}
		}
	}
}

// pipeFieldPath returns the path of the first field the pipe reads, resolved
// against the current scope — the range/with target.
func (a *analysis) pipeFieldPath(p *parse.PipeNode, scope []string, vars map[string][]string) []string {
	if p == nil {
		return nil
	}
	for _, cmd := range p.Cmds {
		for _, arg := range cmd.Args {
			switch t := arg.(type) {
			case *parse.FieldNode:
				return append(append([]string{}, scope...), t.Ident...)
			case *parse.VariableNode:
				if base, ok := vars[t.Ident[0]]; ok {
					return append(append([]string{}, base...), t.Ident[1:]...)
				}
			}
		}
	}
	return nil
}

func cloneVars(v map[string][]string) map[string][]string {
	out := make(map[string][]string, len(v))
	for k, p := range v {
		out[k] = p
	}
	return out
}

// ---------------------------------------------------------------------------
// Data synthesis — a full ContentData in which every referenced field carries a
// unique marker, so the positive control can ask "did field X reach the output".
// ---------------------------------------------------------------------------

func marker(path []string) string {
	clean := make([]string, 0, len(path))
	for _, p := range path {
		if p != "[]" {
			clean = append(clean, p)
		}
	}
	return "RCKMARK_" + strings.Join(clean, "_")
}

func synthesize(s *shape, path []string) interface{} {
	switch s.kind {
	case "list":
		item := s.children["[]"]
		if item == nil || (item.kind == "scalar" && len(item.children) == 0) {
			return []interface{}{marker(path) + "_1", marker(path) + "_2"}
		}
		return []interface{}{
			synthesize(item, append(append([]string{}, path...), "1")),
			synthesize(item, append(append([]string{}, path...), "2")),
		}
	case "map":
		m := map[string]interface{}{}
		for name, child := range s.children {
			if name == "[]" {
				continue
			}
			m[name] = synthesize(child, append(append([]string{}, path...), name))
		}
		return m
	default:
		return marker(path)
	}
}

// contextKeys returns the template keys RenderContext supplies from its own
// json-tagged fields — a field named after one cannot be made absent by
// removing it from ContentData (the about-commercial-block.domain collision),
// so its absence test is skipped and reported as such.
func contextKeys() map[string]bool {
	keys := map[string]bool{}
	t := reflect.TypeOf(actions.RenderContext{})
	for i := 0; i < t.NumField(); i++ {
		tag := t.Field(i).Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}
		keys[strings.Split(tag, ",")[0]] = true
	}
	delete(keys, "content_data")
	return keys
}

// ---------------------------------------------------------------------------
// Output shapes — 295's classes. Counted, and a finding is a shape whose count
// INCREASES when the field is removed, so pre-existing (hardcoded) empties are
// reported once as their own class instead of being blamed on every field.
// ---------------------------------------------------------------------------

// componentIsRuntimeFillShell reports whether this COMPONENT declares itself a
// build-time shell filled by a browser-side loader. The input here is a single
// component's template — never a page — so a bare containment test cannot
// exempt an unrelated section the way bugs_open/137's page-shaped inputs did;
// the scope of the exemption is exactly the scope of the finding it demotes.
func componentIsRuntimeFillShell(tpl string) bool {
	return strings.Contains(tpl, "data-runtime-fill")
}

var shapeRes = []struct {
	name string
	re   *regexp.Regexp
}{
	{"empty_heading", regexp.MustCompile(`<h[1-4][^>]*>\s*</h[1-4]>`)},
	{"dead_anchor", regexp.MustCompile(`<a\s[^>]*>\s*</a>`)},
	{"broken_img", regexp.MustCompile(`<img[^>]*\ssrc=""`)},
	{"empty_cell", regexp.MustCompile(`<td[^>]*>\s*</td>`)},
	{"empty_block", regexp.MustCompile(`<(div|p|span|li)\s+[^>]*class="[^"]*"[^>]*>\s*</(div|p|span|li)>`)},
}

func shapeCounts(html string) map[string]int {
	out := map[string]int{}
	for _, s := range shapeRes {
		if n := len(s.re.FindAllStringIndex(html, -1)); n > 0 {
			out[s.name] = n
		}
	}
	return out
}

type finding struct {
	Component   string `json:"component"`
	Field       string `json:"field"`
	Shape       string `json:"shape"`
	Count       int    `json:"count"` // how many NEW instances of the shape
	RuntimeFill bool   `json:"runtime_fill,omitempty"`
}

// ---------------------------------------------------------------------------
// Baseline — what makes this runnable unattended without being a red nobody
// clears. The census is ~1,000 findings; the SIGNAL is a finding that was not
// there yesterday. A baseline file is a committed, reviewable artefact, so
// growth arrives as a diff a human can read, and no new storage is needed.
//
// The key deliberately EXCLUDES Count: a shape going from x1 to x2 in the same
// component+field is the same defect rendered twice, not a new one, and letting
// counts into the key makes an unrelated markup edit look like regression.
// ---------------------------------------------------------------------------

func findingKey(f finding) string {
	return f.Component + "\x00" + f.Field + "\x00" + f.Shape
}

type baselineFile struct {
	Note string   `json:"note"`
	Keys []string `json:"keys"`
}

func loadBaseline(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bf baselineFile
	if err := json.Unmarshal(b, &bf); err != nil {
		return nil, err
	}
	set := make(map[string]bool, len(bf.Keys))
	for _, k := range bf.Keys {
		set[k] = true
	}
	if len(set) == 0 {
		// A baseline that accidentally parses to zero keys would silently
		// report every finding as NEW — loud refusal instead.
		return nil, fmt.Errorf("baseline %s holds no keys", path)
	}
	return set, nil
}

func writeBaseline(path string, findings []finding, checked int) error {
	keys := make([]string, 0, len(findings))
	for _, f := range findings {
		keys = append(keys, findingKey(f))
	}
	sort.Strings(keys)
	bf := baselineFile{
		Note: fmt.Sprintf("component-render-check baseline: %d findings across %d analysed components. "+
			"Keys are component\\0field\\0shape — counts deliberately excluded. "+
			"Regenerate with --write-baseline; a NEW key is the signal.", len(keys), checked),
		Keys: keys,
	}
	b, err := json.MarshalIndent(bf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func main() {
	jsonPath := flag.String("json", "", "read components from a file instead of the cluster")
	only := flag.String("component", "", "check a single component by name")
	emitJSON := flag.Bool("emit-json", false, "machine-readable findings on stdout")
	baseline := flag.String("baseline", "", "compare against a baseline file; exit 1 if any finding is NEW")
	writeBase := flag.String("write-baseline", "", "write the current findings as a baseline file and exit 0")
	flag.Parse()

	if *baseline != "" && *only != "" {
		fmt.Fprintln(os.Stderr, "--baseline with --component would report every other component's "+
			"findings as FIXED — refusing (compare whole runs, or use --write-baseline on a full run)")
		os.Exit(2)
	}

	comps, err := loadComponents(*jsonPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	logger := zap.NewNop()
	ctxKeys := contextKeys()

	var findings []finding
	var hardcoded []finding  // empty even with every field present
	var overfiring []finding // positive control failed: marker never reached output
	var skippedCtx []string  // component.field skipped for context collision
	var unanalysed []string  // components whose template failed to parse
	unanalysedNames := map[string]bool{}
	checked, runtimeFillComps := 0, 0

	for _, c := range comps {
		if *only != "" && c.Name != *only {
			continue
		}
		an := analyse(c.Template)
		if an.err != nil {
			unanalysed = append(unanalysed, fmt.Sprintf("%s (%v)", c.Name, an.err))
			unanalysedNames[c.Name] = true
			continue
		}
		if len(an.root.children) == 0 {
			continue // static template, nothing to probe
		}
		checked++
		runtimeFill := componentIsRuntimeFillShell(c.Template)
		if runtimeFill {
			runtimeFillComps++
		}

		full := map[string]interface{}{}
		for name, child := range an.root.children {
			full[name] = synthesize(child, []string{name})
		}
		baseRender := actions.RenderTemplate(c.Template, &actions.RenderContext{ContentData: full}, logger)
		baseCounts := shapeCounts(baseRender)
		for shapeName, n := range baseCounts {
			hardcoded = append(hardcoded, finding{Component: c.Name, Shape: shapeName, Count: n, RuntimeFill: runtimeFill})
		}

		fields := make([]string, 0, len(an.root.children))
		for name := range an.root.children {
			fields = append(fields, name)
		}
		sort.Strings(fields)

		for _, f := range fields {
			if ctxKeys[f] {
				skippedCtx = append(skippedCtx, c.Name+"."+f)
				continue
			}
			// Positive control: the field's markers must reach the baseline.
			if !strings.Contains(baseRender, marker([]string{f})) && !strings.Contains(baseRender, "RCKMARK_"+f+"_") {
				overfiring = append(overfiring, finding{Component: c.Name, Field: f, Shape: "marker_never_rendered", RuntimeFill: runtimeFill})
			}
			probe := map[string]interface{}{}
			for k, v := range full {
				if k != f {
					probe[k] = v
				}
			}
			out := actions.RenderTemplate(c.Template, &actions.RenderContext{ContentData: probe}, logger)
			for shapeName, n := range shapeCounts(out) {
				if n > baseCounts[shapeName] {
					findings = append(findings, finding{
						Component: c.Name, Field: f, Shape: shapeName,
						Count: n - baseCounts[shapeName], RuntimeFill: runtimeFill,
					})
				}
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Component != findings[j].Component {
			return findings[i].Component < findings[j].Component
		}
		if findings[i].Field != findings[j].Field {
			return findings[i].Field < findings[j].Field
		}
		return findings[i].Shape < findings[j].Shape
	})

	if *writeBase != "" {
		if len(unanalysed) > 0 {
			// Baselining a run that silently dropped components would bake the
			// drop in as "clean" — the blind-check-outlives-the-blindness shape.
			fmt.Fprintf(os.Stderr, "refusing to baseline: %d component(s) failed to parse and are "+
				"therefore uncovered\n", len(unanalysed))
			os.Exit(2)
		}
		if err := writeBaseline(*writeBase, findings, checked); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		fmt.Printf("wrote baseline %s: %d findings across %d analysed components\n",
			*writeBase, len(findings), checked)
		return
	}

	// Baseline comparison. New findings are the signal; disappeared ones are
	// reported too, because a baseline nobody regenerates is a baseline that
	// slowly stops describing the tree.
	var newFindings []finding
	var fixedKeys, uncoveredKeys []string
	if *baseline != "" {
		base, err := loadBaseline(*baseline)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		seen := map[string]bool{}
		for _, f := range findings {
			k := findingKey(f)
			seen[k] = true
			if !base[k] {
				newFindings = append(newFindings, f)
			}
		}
		// A component that failed to parse produced NO findings, so every
		// baseline key it owns would otherwise read as "fixed" — a silent pass
		// bought by going blind, which is how a broken template would clear its
		// own findings. Proven by mutation 2026-08-04: breaking one template
		// reported its 2 findings as fixed and still exited 0. Those keys are
		// UNCOVERED, and an uncovered key fails the run.
		for k := range base {
			if seen[k] {
				continue
			}
			pretty := strings.ReplaceAll(k, "\x00", " .")
			if unanalysedNames[strings.SplitN(k, "\x00", 2)[0]] {
				uncoveredKeys = append(uncoveredKeys, pretty)
				continue
			}
			fixedKeys = append(fixedKeys, pretty)
		}
		sort.Strings(fixedKeys)
		sort.Strings(uncoveredKeys)
	}

	if *emitJSON {
		out := map[string]interface{}{
			"findings": findings, "hardcoded_empties": hardcoded, "positive_control_failures": overfiring,
			"skipped_context_collisions": skippedCtx, "unanalysed": unanalysed,
			"components_checked": checked,
		}
		if *baseline != "" {
			out["new_findings"] = newFindings
			out["fixed_since_baseline"] = fixedKeys
			out["uncovered_since_baseline"] = uncoveredKeys
		}
		json.NewEncoder(os.Stdout).Encode(out)
		if len(newFindings) > 0 || len(uncoveredKeys) > 0 {
			os.Exit(1)
		}
		return
	}

	if *baseline != "" {
		fmt.Printf("component-render-check vs baseline %s: %d findings now, %d NEW, %d fixed, %d UNCOVERED\n\n",
			*baseline, len(findings), len(newFindings), len(fixedKeys), len(uncoveredKeys))
		if len(newFindings) > 0 {
			fmt.Println("NEW since the baseline — a component started being able to render a hole:")
			for _, f := range newFindings {
				fmt.Printf("  %-38s .%-28s %s x%d\n", f.Component, f.Field, f.Shape, f.Count)
			}
			fmt.Println()
		}
		if len(uncoveredKeys) > 0 {
			fmt.Println("UNCOVERED — the baseline knows these findings but the component no longer PARSES, " +
				"so their absence is blindness, not a fix:")
			for _, k := range uncoveredKeys {
				fmt.Printf("  %s\n", k)
			}
			fmt.Println()
		}
		if len(fixedKeys) > 0 {
			fmt.Println("gone since the baseline (regenerate the baseline to bank these):")
			for _, k := range fixedKeys {
				fmt.Printf("  %s\n", k)
			}
			fmt.Println()
		}
		if len(unanalysed) > 0 {
			fmt.Printf("unanalysed (NOT cleared): %d — %s\n", len(unanalysed), strings.Join(unanalysed, "; "))
		}
		if len(newFindings) > 0 || len(uncoveredKeys) > 0 {
			os.Exit(1)
		}
		return
	}
	real, info := 0, 0
	fmt.Printf("component-render-check: %d active components analysed (%d with runtime-fill)\n\n", checked, runtimeFillComps)
	if len(findings) > 0 {
		fmt.Println("ABSENT FIELD => EMPTY ELEMENT (the class the template lint cannot see):")
		for _, f := range findings {
			tag := ""
			if f.RuntimeFill {
				tag = "  [runtime-fill: info only]"
				info++
			} else {
				real++
			}
			fmt.Printf("  %-38s .%-28s %s x%d%s\n", f.Component, f.Field, f.Shape, f.Count, tag)
		}
		fmt.Printf("  => %d findings (%d demoted to info by data-runtime-fill)\n\n", real+info, info)
	}
	if len(hardcoded) > 0 {
		fmt.Println("EMPTY EVEN WITH EVERY FIELD PRESENT (hardcoded blank — not a gating bug):")
		for _, f := range hardcoded {
			fmt.Printf("  %-38s %s x%d\n", f.Component, f.Shape, f.Count)
		}
		fmt.Println()
	}
	if len(overfiring) > 0 {
		fmt.Println("POSITIVE CONTROL FAILED (field supplied, marker never rendered — possible over-firing gate, or a field only feeding an attribute/condition):")
		for _, f := range overfiring {
			fmt.Printf("  %-38s .%s\n", f.Component, f.Field)
		}
		fmt.Println()
	}
	if len(skippedCtx) > 0 {
		fmt.Printf("skipped (RenderContext supplies the key — absence not testable via ContentData): %s\n", strings.Join(skippedCtx, ", "))
	}
	if len(unanalysed) > 0 {
		fmt.Printf("unanalysed (template failed to parse — NOT cleared, listed so nothing is silently dropped): %d\n", len(unanalysed))
		for _, u := range unanalysed {
			fmt.Println("   ", u)
		}
	}
}
