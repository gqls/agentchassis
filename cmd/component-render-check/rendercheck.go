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
	"crypto/md5"
	"database/sql"
	_ "embed"
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
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Loading — same route and same flake as the Python lint: the ~2 MB stream
// through kubectl exec truncates intermittently, so three attempts, and the
// final failure exits 2.
// ---------------------------------------------------------------------------

const query = `SELECT COALESCE(jsonb_agg(jsonb_build_object(
  'name', name, 'function', function, 'html_template', html_template,
  'created_at', created_at)), '[]'::jsonb)
FROM content_components WHERE is_active;`

var psqlArgv = []string{"kubectl", "exec", "-n", "ai-persona-system", "postgres-clients-0", "--",
	"psql", "-U", "clients_user", "-d", "clients_db", "-tAc"}

type component struct {
	Name      string `json:"name"`
	Function  string `json:"function"`
	Template  string `json:"html_template"`
	CreatedAt string `json:"created_at"` // only for picking a stable clone representative
}

// dbConn returns a direct Postgres handle when PG_CLIENTS_HOST is set, and nil
// otherwise. Two routes exist for the same reason the Python lint has two: a
// session at a terminal reaches the DB through `kubectl exec`, but the CronJob
// CANNOT — the ai-persona-app service account has no pods/exec RBAC in this
// namespace, and a kubectl-only tool fails there in a way that looks like a
// clean run if you are not watching exit codes. PG_CLIENTS_HOST being set is how
// the CronJob declares itself, exactly as check_placeholder_fallbacks.py does.
func dbConn() (*sql.DB, error) {
	host := os.Getenv("PG_CLIENTS_HOST")
	if host == "" {
		return nil, nil
	}
	pw := os.Getenv("CLIENTS_DB_PASSWORD")
	if pw == "" {
		return nil, fmt.Errorf("PG_CLIENTS_HOST is set but CLIENTS_DB_PASSWORD is not — " +
			"refusing to guess a connection")
	}
	port := os.Getenv("PG_CLIENTS_PORT")
	if port == "" {
		port = "5432"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=clients_user password=%s dbname=clients_db sslmode=disable",
		host, port, pw)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func loadComponents(path string) ([]component, error) {
	var body []byte
	var err error
	if path != "" {
		body, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}
	} else if db, derr := dbConn(); derr != nil || db != nil {
		if derr != nil {
			return nil, derr
		}
		defer db.Close()
		// No retry loop here: this is a single query over a local socket, not the
		// ~2 MB kubectl exec stream whose truncation the retry below exists for.
		var raw string
		if err := db.QueryRow(query).Scan(&raw); err != nil {
			return nil, fmt.Errorf("component query failed: %w", err)
		}
		body = []byte(raw)
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
//
// IT ALSO EXCLUDES A CLONE'S OWN NAME, for exactly the same reason and added for
// exactly the same failure (owner ruling, 2026-08-05). The first UNATTENDED run
// went red with 13 NEW findings that were all one component:
// `tool-ab-test-calculator_pre_037-idea-uk`, created 01:19 that morning with an
// html_template BYTE-IDENTICAL (md5 8673be08…) to `tool-ab-test-calculator_pre_037`
// from February — whose same 13 field/shape findings were already accepted in the
// baseline. An identical template renders identically, so those were not 13 new
// defects; they were 13 known ones arriving under a new name. Clones are routine
// (5 families / 10 components live, 3 created in the last 7 days), so left alone
// this makes the daily job red about weekly with findings nobody can clear —
// precisely the "red nobody clears" outcome the baseline exists to prevent.
//
// So findings are keyed by the REPRESENTATIVE of their template-identity group:
// the OLDEST active component sharing that exact html_template. Oldest, not
// alphabetically-first, so a newly-arrived clone can never displace the incumbent
// and shift every key under it. Consequences worth knowing:
//   - a clone that is later EDITED no longer matches the hash, gets its own
//     identity, and reports its findings as NEW — which is correct, because its
//     template genuinely differs now;
//   - if the representative is deleted or deactivated while a clone survives, the
//     clone becomes the oldest and the keys move once: its findings read NEW and
//     the old ones read fixed, in the same run. That is loud and self-explaining
//     rather than silent, which is the side to fail on;
//   - offline `--json` fixtures carry no created_at, so the tie-break falls back
//     to the lexicographically smallest name — deterministic, which is all a
//     fixture needs.
// ---------------------------------------------------------------------------

// canonicalComponent maps every active component name to the representative of
// its template-identity group. Package-level because findingKey is called from
// both the compare path and writeBaseline, and threading a map through both for
// a single-file command buys nothing. Empty until buildCanonicalComponents runs,
// and an absent entry means "the component is its own representative", so the
// zero value behaves exactly like the pre-2026-08-05 tool.
var canonicalComponent = map[string]string{}

func buildCanonicalComponents(comps []component) {
	canonicalComponent = map[string]string{}
	type rep struct{ name, created string }
	byTemplate := map[string]rep{}
	for _, c := range comps {
		h := fmt.Sprintf("%x", md5.Sum([]byte(c.Template)))
		cur, seen := byTemplate[h]
		// Oldest wins; with no created_at (offline fixtures) fall back to the
		// smallest name so the choice is still deterministic.
		if !seen ||
			(c.CreatedAt != "" && cur.created != "" && c.CreatedAt < cur.created) ||
			((c.CreatedAt == "" || cur.created == "") && c.Name < cur.name) {
			byTemplate[h] = rep{name: c.Name, created: c.CreatedAt}
		}
	}
	for _, c := range comps {
		h := fmt.Sprintf("%x", md5.Sum([]byte(c.Template)))
		if r, ok := byTemplate[h]; ok && r.name != c.Name {
			canonicalComponent[c.Name] = r.name
		}
	}
}

// canonicalName maps a component onto the representative of its template-identity
// group. ONE implementation, because the baseline's KEY set and its COVERED-component
// set must agree about identity: if findingKey folds a clone onto its representative
// while the covered set records the clone's own name, that clone reads as unbaselined
// for ever and the ratchet quietly stops watching it.
func canonicalName(name string) string {
	if rep, ok := canonicalComponent[name]; ok {
		return rep
	}
	return name
}

func findingKey(f finding) string {
	return canonicalName(f.Component) + "\x00" + f.Field + "\x00" + f.Shape
}

// classifyAgainstBaseline is THE ratchet decision, in one place so it can be proved
// by test rather than only by running the whole tool (bugs_open/361 §6 requires BOTH
// arms be mutation-proved, and a fix verified only on the "does not fail" arm is a
// fix that turned the check off).
//
//   - regression — the baseline COVERED this component and does not hold this key:
//     something that used to be fine got worse. Fails the run.
//   - unbaselined — the baseline never analysed this component: growth, not
//     regression. Reported with its own count, does not fail the run.
//
// Both arms key on the CANONICAL name, so a clone and its representative agree.
func classifyAgainstBaseline(findings []finding, base, covered map[string]bool) (regressions, unbaselined []finding, unbaselinedComps map[string]bool) {
	unbaselinedComps = map[string]bool{}
	for _, f := range findings {
		if base[findingKey(f)] {
			continue
		}
		if covers(covered, f.Component) {
			regressions = append(regressions, f)
			continue
		}
		unbaselined = append(unbaselined, f)
		unbaselinedComps[canonicalName(f.Component)] = true
	}
	return regressions, unbaselined, unbaselinedComps
}

// baselineComponent is the canonical component a baseline key belongs to — the
// first of its three NUL-separated segments.
func baselineComponent(key string) string {
	return strings.SplitN(key, "\x00", 2)[0]
}

// writeDocNote records the run in doc_notes on EVERY run, clean or not — the
// convention check_placeholder_fallbacks.py established for the same reason: a
// check that only speaks when it fails is indistinguishable from one that has
// stopped running, which is the exact ambiguity bugs_open/140 hit. Best-effort:
// a failure to record must never become a failure to report.
func writeDocNote(body string) {
	db, err := dbConn()
	if err != nil || db == nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "(could not record run: %v)\n", err)
		}
		return
	}
	defer db.Close()
	_, err = db.Exec(`INSERT INTO doc_notes (subject_type, subject_key, body, categories, source)
	                  VALUES ('pipeline', 'component-library', $1, '["component-integrity"]'::jsonb,
	                          'component_render_check')`, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "(could not write doc_notes: %v)\n", err)
	}
}

type baselineFile struct {
	Note string   `json:"note"`
	Keys []string `json:"keys"`
	// Components is the COVERED set at write time: every component whose template
	// PARSED — raw names, clones and static templates included, never canonicalised.
	//
	// WHY IT EXISTS (bugs_open/361). Without it the ratchet can only ask "does this
	// finding's key appear in the baseline?", which cannot distinguish a component
	// that REGRESSED from one that did not exist when the baseline was cut — so a
	// growing library manufactures "NEW" findings and the job goes permanently red.
	// Scoping by "does this component own any baseline KEY?" is not enough either,
	// and the old artefact proves why: its own note says "1023 findings across 139
	// analysed components" while its keys span only 115 components, so 24 components
	// were analysed and CLEAN. Those 24 are invisible to a keys-derived covered set,
	// and a clean component that later regresses is exactly what a ratchet is for.
	// Recording coverage separately from findings is what makes that state
	// unrepresentable.
	//
	// TWO REASONS IT IS RAW AND NOT CANONICAL, both found by review of the first cut:
	//   - a clone that is later EDITED stops matching its representative's hash and
	//     gets its own identity (see the note above findingKey, which says its findings
	//     then reporting as NEW "is correct"). Only its OWN name can vouch for it; a
	//     canonical-only set silently exempts every clone that later diverges.
	//   - a STATIC template (no template actions) is skipped before `checked++`, so an
	//     analysed-only set cannot see it either. A static component later rewritten to
	//     reference fields and render a hole is exactly this check's stated signal —
	//     "a component started being able to render a hole" — and it must fail.
	// `covers` below reads raw first, then the representative, so a clone born after
	// the baseline still inherits its representative's coverage.
	Components []string `json:"components,omitempty"`
}

// The baseline is EMBEDDED, not mounted, and that is the whole design.
//
// It is the check's own definition of "already known", so it must be inseparable
// from the code that interprets it. A baseline in a ConfigMap (or COPYed into an
// image from elsewhere) could be edited to silence a finding with no diff anybody
// reviews — which is precisely the class of quiet, unreviewed clearing this check
// exists to catch. Embedded, banking a real fix is: --write-baseline, commit the
// diff, rebuild. Visible, reviewable, revertable, and impossible to do by accident.
//
// `--baseline <path>` still overrides, for a session comparing against an older
// baseline by hand. It cannot be overridden in the CronJob, which passes no flag.
//
//go:embed baseline.json
var embeddedBaseline []byte

// loadBaseline returns the key set, the COVERED-component set (canonical names),
// and whether the artefact predates component scoping.
//
// A legacy baseline (no "components" field) does NOT refuse. Refusing would turn a
// ratchet that is merely mis-scoped into a job that cannot run at all, and
// regenerating the baseline is a debt decision — it banks every outstanding finding
// as "already known" — which is not this tool's to take on its own. It falls back to
// deriving coverage from the keys, which is strictly better than today's behaviour,
// and it says so LOUDLY on stdout and in the doc_notes row, naming the blind spot,
// because a fallback nobody is told about is how a stale artefact becomes folklore.
func loadBaseline(path string) (keys map[string]bool, covered map[string]bool, legacy bool, err error) {
	b := embeddedBaseline
	if path != "" {
		b, err = os.ReadFile(path)
		if err != nil {
			return nil, nil, false, err
		}
	}
	var bf baselineFile
	if err := json.Unmarshal(b, &bf); err != nil {
		return nil, nil, false, err
	}
	keys = make(map[string]bool, len(bf.Keys))
	for _, k := range bf.Keys {
		keys[k] = true
	}
	if len(keys) == 0 {
		// A baseline that accidentally parses to zero keys would silently
		// report every finding as NEW — loud refusal instead.
		return nil, nil, false, fmt.Errorf("baseline %s holds no keys", path)
	}
	covered = make(map[string]bool, len(bf.Components))
	switch {
	case bf.Components == nil:
		// Legacy artefact: derive coverage from the keys. Strictly better than the
		// key-level diff it replaces, and it announces its own blind spot.
		legacy = true
		for k := range keys {
			covered[baselineComponent(k)] = true
		}
	case len(bf.Components) == 0:
		// Present but EMPTY is a ratchet switched off by a hand edit: every finding
		// would be unbaselined and nothing could ever fail. The tool never writes it.
		// Distinguished from absent, which Go's decoder gives us for free.
		return nil, nil, false, fmt.Errorf("baseline %s: \"components\" is present but EMPTY — every "+
			"finding would read as unbaselined and nothing could fail; the tool never writes this", path)
	default:
		for _, c := range bf.Components {
			covered[c] = true
		}
	}
	return keys, covered, legacy, nil
}

// covers answers the ratchet's scoping question: did this baseline vouch for the
// component? Raw name first — an edited clone must stay accountable under its own
// name — then its template-identity representative, so a clone born after the
// baseline inherits the coverage of the component it is a copy of.
func covers(covered map[string]bool, component string) bool {
	if covered[component] {
		return true
	}
	if rep, ok := canonicalComponent[component]; ok {
		return covered[rep]
	}
	return false
}

func writeBaseline(path string, findings []finding, analysed map[string]bool, checkedCount int) error {
	// DEDUPED, which only started mattering on 2026-08-05: once a clone's findings
	// key by their template's representative, two identical components contribute
	// the SAME key and an append-per-finding loop would write it twice. A duplicate
	// is harmless to the lookup (a set either way) but not to the artefact — the
	// baseline is meant to be a diff a human reads, and its own count would
	// overstate what it covers.
	seen := make(map[string]bool, len(findings))
	keys := make([]string, 0, len(findings))
	for _, f := range findings {
		k := findingKey(f)
		if seen[k] {
			continue
		}
		seen[k] = true
		keys = append(keys, k)
	}
	sort.Strings(keys)
	// RAW names, one per component that parsed — NOT canonicalised and NOT deduped by
	// template. See the Components field comment: canonicalising here would exempt
	// every clone that is later edited, and dropping the static templates would exempt
	// every component that gains its first field after the baseline was cut.
	comps := make([]string, 0, len(analysed))
	for name := range analysed {
		comps = append(comps, name)
	}
	sort.Strings(comps)
	bf := baselineFile{
		Note: fmt.Sprintf("component-render-check baseline: %d findings across %d analysed components; "+
			"%d COVERED (analysed + static templates, which are vouched for but not probed). "+
			"Keys are component\\0field\\0shape — counts deliberately excluded. "+
			"\"components\" is the ANALYSED set: a finding in a component listed there but not in "+
			"\"keys\" is a REGRESSION and fails the run; a finding in a component absent from it is "+
			"unbaselined growth and does not. Regenerate with --write-baseline.",
			len(keys), checkedCount, len(comps)),
		Keys:       keys,
		Components: comps,
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
	compare := flag.Bool("compare", false, "compare against the EMBEDDED baseline; exit 1 on a NEW or UNCOVERED finding")
	baseline := flag.String("baseline", "", "compare against a baseline FILE instead of the embedded one")
	writeBase := flag.String("write-baseline", "", "write the current findings as a baseline file and exit 0")
	report := flag.Bool("report", false, "also write a doc_notes row (used by the CronJob)")
	flag.Parse()

	comparing := *compare || *baseline != ""
	if comparing && *only != "" {
		fmt.Fprintln(os.Stderr, "--baseline with --component would report every other component's "+
			"findings as FIXED — refusing (compare whole runs, or use --write-baseline on a full run)")
		os.Exit(2)
	}
	// Since the baseline records COVERAGE (bugs_open/361), a single-component write
	// would claim the run analysed one component — and a later compare would then read
	// every OTHER component as unbaselined and pass. That is the "fix that turned the
	// check off" shape, bought silently, so it is refused at the flag rather than
	// discovered in a green run.
	if *writeBase != "" && *only != "" {
		fmt.Fprintln(os.Stderr, "--write-baseline with --component would record a covered set of ONE "+
			"component, so every other component would read as unbaselined and pass — refusing "+
			"(write a baseline from a full run)")
		os.Exit(2)
	}

	comps, err := loadComponents(*jsonPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	// Built from the FULL set, before --component filters anything: a group's
	// representative must not depend on which component you asked about. (--compare
	// with --component is already refused above, so the compare path always sees
	// every component anyway.)
	buildCanonicalComponents(comps)

	logger := zap.NewNop()
	ctxKeys := contextKeys()

	var findings []finding
	var hardcoded []finding  // empty even with every field present
	var overfiring []finding // positive control failed: marker never reached output
	var skippedCtx []string  // component.field skipped for context collision
	var unanalysed []string  // components whose template failed to parse
	unanalysedNames := map[string]bool{}
	// The components this run actually probed. Kept in lockstep with `checked` — it
	// IS `checked`, by name rather than by count — because the baseline must be able
	// to say which components it covered, not merely how many.
	analysedNames := map[string]bool{}
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
		// COVERED, and recorded BEFORE the static skip below. A template with no
		// actions has no field whose absence could be probed, so it is not "analysed"
		// and never reaches `checked` — but the baseline DID look at it and found it
		// incapable of rendering a hole. That is a vouch, and if the template is later
		// rewritten to reference a field and render one, that is a regression. Placing
		// this after the `continue` was a real defect in the first cut of this fix.
		analysedNames[c.Name] = true
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
		// bugs_open/260: the render seam now returns an error instead of
		// silently regex-substituting. A baseline that cannot execute is an
		// UNCOVERED component, not a clean one — the same bucket a parse
		// failure goes to, which already refuses --write-baseline and counts
		// the component's baseline keys as uncovered rather than fixed.
		baseRender, _, _, baseErr := actions.RenderTemplate(c.Template, &actions.RenderContext{ContentData: full}, logger)
		if baseErr != nil {
			unanalysed = append(unanalysed, fmt.Sprintf("%s (baseline render: %v)", c.Name, baseErr))
			unanalysedNames[c.Name] = true
			delete(analysedNames, c.Name)
			checked--
			continue
		}
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
			out, _, _, probeErr := actions.RenderTemplate(c.Template, &actions.RenderContext{ContentData: probe}, logger)
			if probeErr != nil {
				// ⚠ THE DETECTOR CAN BE BLINDED BY ITS OWN FIX. This probe
				// deliberately REMOVES one field and re-renders; since
				// bugs_open/260 that render can fail outright, and a failed
				// render returns "" — whose shapeCounts are all zero, i.e.
				// indistinguishable from "this field is safely gated". Left
				// unhandled, the strictest possible seam would have made this
				// audit quietly report fewer findings. Recorded as uncovered
				// instead, per field, so it can never read as a pass.
				unanalysed = append(unanalysed, fmt.Sprintf("%s.%s (probe render: %v)", c.Name, f, probeErr))
				unanalysedNames[c.Name] = true
				// It is uncovered, so it must leave the covered set too — otherwise a
				// baseline would vouch for a component this run could not fully probe.
				delete(analysedNames, c.Name)
				continue
			}
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
		if err := writeBaseline(*writeBase, findings, analysedNames, checked); err != nil {
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
	// THE RATCHET IS COMPONENT-SCOPED (bugs_open/361). A finding is only NEW — i.e.
	// only fails the run — when the baseline COVERED its component and did not record
	// the key: that is a component which used to be fine and got worse, which is the
	// one thing a ratchet is for. A finding in a component the baseline never analysed
	// is unbaselined GROWTH: real debt, reported with its own count every day, but not
	// a regression, and failing on it makes the job permanently red as the library
	// grows (25 consecutive red days, 2026-08-09 → 2026-09-03).
	//
	// THE COST, STATED RATHER THAN HIDDEN: a brand-new component that renders a hole
	// will not fail this job. That debt belongs to birth-time gating (CGV-029 and the
	// component birth path), not to a regression ratchet.
	var newFindings []finding // regressions — these fail the run
	var unbaselined []finding // growth in components the baseline never covered
	var fixedKeys, uncoveredKeys []string
	unbaselinedComps := map[string]bool{}
	legacyBaseline := false
	inherited := 0
	inheritedFrom := map[string]bool{}
	if comparing {
		base, covered, legacy, err := loadBaseline(*baseline)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		legacyBaseline = legacy
		newFindings, unbaselined, unbaselinedComps = classifyAgainstBaseline(findings, base, covered)
		seen := map[string]bool{}
		for _, f := range findings {
			seen[findingKey(f)] = true
			// Counted and REPORTED, never silently dropped. A finding that only
			// stayed quiet because an identical template already owns its key is
			// exactly the thing a reader would want to know was suppressed —
			// otherwise this is a filter that hides its own effect, which is the
			// blindness shape this tool exists to close.
			if rep, ok := canonicalComponent[f.Component]; ok {
				inherited++
				inheritedFrom[rep] = true
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
		if comparing {
			// "new_findings" keeps its name and now holds exactly the set that FAILS
			// the run — regressions in covered components. The growth it used to be
			// diluted by is reported beside it, never dropped.
			out["new_findings"] = newFindings
			out["unbaselined_findings"] = unbaselined
			out["unbaselined_components"] = len(unbaselinedComps)
			out["baseline_is_legacy"] = legacyBaseline
			out["fixed_since_baseline"] = fixedKeys
			out["uncovered_since_baseline"] = uncoveredKeys
		}
		json.NewEncoder(os.Stdout).Encode(out)
		if len(newFindings) > 0 || len(uncoveredKeys) > 0 {
			os.Exit(1)
		}
		return
	}

	if comparing {
		// "%d analysed" WITH ITS DENOMINATOR (2026-08-06). It read "139 analysed" on
		// 2026-08-04 and again on 08-06, while the library grew 176 -> 184: a reader
		// tracking the number would have seen a constant and concluded coverage was
		// steady, when what stayed constant was the count of components carrying any
		// template action at all. Both facts are needed to read either. The skipped
		// remainder is not a gap — a template with no actions has no field whose
		// absence could be probed — but an invisible denominator is how a filtered
		// count gets quoted as a census.
		// The unbaselined count is IN THE SUMMARY LINE, not only in the detail, because
		// it is the number this change stops failing on — if it were reported anywhere
		// less prominent, scoping the ratchet would read as making the debt disappear.
		summary := fmt.Sprintf("component-render-check vs baseline: %d findings across %d of %d active "+
			"components (%d have no template actions to probe), %d REGRESSION, "+
			"%d unbaselined across %d new component(s), %d fixed, %d UNCOVERED",
			len(findings), checked, len(comps), len(comps)-checked-len(unanalysed),
			len(newFindings), len(unbaselined), len(unbaselinedComps),
			len(fixedKeys), len(uncoveredKeys))
		if inherited > 0 {
			reps := make([]string, 0, len(inheritedFrom))
			for r := range inheritedFrom {
				reps = append(reps, r)
			}
			sort.Strings(reps)
			summary += fmt.Sprintf(", %d inherited from an identical template (%s)",
				inherited, strings.Join(reps, ", "))
		}
		// AFTER the inherited clause, deliberately: this starts a NEW line, so anything
		// appended to `summary` below it lands on the warning rather than the summary.
		// Shipped the other way round on 2026-09-03 and the live row read
		// "...close that blind spot., 3 inherited from an identical template (...)" —
		// which put the clone-suppression count on the warning line, where the daily
		// series query (split_part(body, E'\n', 1)) cannot see it.
		if legacyBaseline {
			summary += "\n⚠ LEGACY BASELINE: it records no \"components\" list, so coverage was derived " +
				"from its keys. A component that was analysed and CLEAN when the baseline was cut is " +
				"therefore read as unbaselined, and a regression in one would NOT fail this run. " +
				"Regenerate with --write-baseline to close that blind spot."
		}
		if *report {
			detail := summary
			for _, f := range newFindings {
				detail += fmt.Sprintf("\nREGRESSION  %s .%s %s x%d", f.Component, f.Field, f.Shape, f.Count)
			}
			// Listed, not merely counted: the row is the only durable record of this
			// debt, and a count with no names cannot be acted on by whoever reads it.
			for _, f := range unbaselined {
				detail += fmt.Sprintf("\nunbaselined  %s .%s %s x%d", f.Component, f.Field, f.Shape, f.Count)
			}
			for _, k := range uncoveredKeys {
				detail += "\nUNCOVERED  " + k
			}
			for _, k := range fixedKeys {
				detail += "\nfixed  " + k
			}
			writeDocNote(detail)
		}
		src := *baseline
		if src == "" {
			src = "(embedded)"
		}
		fmt.Printf("component-render-check vs baseline %s: %d findings across %d of %d active components, "+
			"%d REGRESSION, %d unbaselined across %d new component(s), %d fixed, %d UNCOVERED\n",
			src, len(findings), checked, len(comps), len(newFindings),
			len(unbaselined), len(unbaselinedComps), len(fixedKeys), len(uncoveredKeys))
		if legacyBaseline {
			fmt.Println("  ⚠ LEGACY BASELINE (no \"components\" list): coverage derived from the keys, so a " +
				"component analysed and CLEAN at baseline time reads as unbaselined and a regression in " +
				"one would NOT fail. Regenerate with --write-baseline to close it.")
		}
		// On stdout as well as in the doc_notes row: a session running --compare by
		// hand must be able to see what the clone rule suppressed, or the rule is a
		// filter that hides its own effect.
		if inherited > 0 {
			reps := make([]string, 0, len(inheritedFrom))
			for r := range inheritedFrom {
				reps = append(reps, r)
			}
			sort.Strings(reps)
			fmt.Printf("  (%d finding(s) inherited from an identical template and therefore NOT new — "+
				"clone(s) of: %s)\n", inherited, strings.Join(reps, ", "))
		}
		fmt.Println()
		if len(newFindings) > 0 {
			fmt.Println("REGRESSION — a component the baseline COVERED started being able to render a hole:")
			for _, f := range newFindings {
				fmt.Printf("  %-38s .%-28s %s x%d\n", f.Component, f.Field, f.Shape, f.Count)
			}
			fmt.Println()
		}
		if len(unbaselined) > 0 {
			fmt.Printf("unbaselined — %d finding(s) in %d component(s) the baseline never analysed. "+
				"Real debt, NOT a regression, and deliberately not failing this run: birth-time gating "+
				"(CGV-029) owns it, a regression ratchet cannot.\n", len(unbaselined), len(unbaselinedComps))
			for _, f := range unbaselined {
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
	fmt.Printf("component-render-check: %d of %d active components analysed — %d have no template "+
		"actions to probe, %d failed to parse (%d with runtime-fill)\n\n",
		checked, len(comps), len(comps)-checked-len(unanalysed), len(unanalysed), runtimeFillComps)
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
