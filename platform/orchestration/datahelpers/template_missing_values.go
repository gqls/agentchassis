// FILE: platform/orchestration/datahelpers/template_missing_values.go
//
// What to do when a prompt template renders a hole.
//
// THE DEFECT. text/template renders a reference it cannot resolve as the literal
// string "<no value>". RenderPromptTemplate then sent that string to the model,
// so a prompt could say
//
//	## Official Contact Information (USE ONLY THESE - DO NOT INVENT)
//	Location: <no value>
//
// — a stand-in token presented to the model as authoritative business data,
// inside the one block the prompt tells it not to invent around. Measured by the
// bugs_open/437 lane on 2026-09-03: roughly 65% of page-content-writer prompts
// carried at least one. Nothing had ever asserted on it; the only signal was a
// Warn that counted occurrences and named no field.
//
// This is bugs_open/453's shape 3 — the one no check over CONFIG can reach,
// because the config is correct and the outcome depends on a row. At RENDER time
// it is decidable, because the data is in hand. That asymmetry is the whole
// reason this file exists on this side of the seam and the WFA-024 lint on the
// other.
//
// WHY IT WAS WORSE THAN AN EMPTY STRING. bugs_open/387 measured the cost of a
// human-authored stand-in reaching instructed copy: 14 of 137 instructed calls
// shipped it. This token is not authored by anyone — the renderer manufactures
// it, silently, on two thirds of calls. The models have been declining to copy
// it: 0 of 3,228 stored page components carry the string (437 lane, 2026-09-03).
// That is a behaviour we rely on and have never asked for, which is not a
// control.
//
// THE SHAPE OF THE FIX IS NOT NEW, AND DELIBERATELY SO. The component render
// seam solved the same problem for page templates and its answer is the
// precedent this mirrors (component_library.go, around missingBareFields):
// strip the artefact, name the fields rather than counting them, and escalate
// from Warn to Error when the blank landed somewhere that makes it dangerous
// rather than merely absent. There, dangerous means an href=/src= — a dead
// control. Here it means inside a block instructing the model not to invent.
//
// THREE THINGS IT DOES NOT DO, each for a stated reason:
//
//   - It does not REFUSE. Refusing would be new authority over prompts that
//     render successfully today across the whole fleet (owner ruling
//     2026-08-02 §2), and the damage is a missing sentence, not a corrupt one.
//     What this closes is the SILENCE.
//   - It does not SUBSTITUTE a value. There is nothing this layer could
//     honestly put there, and inventing one is the failure being prevented.
//   - It does not attribute perfectly. Occurrence counting is exact (it reads
//     the output); field attribution is best-effort (it reads the template and
//     the data). Where the two disagree the count is the truth, and the report
//     says so rather than quietly reconciling them.
package datahelpers

import (
	"regexp"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"
)

// MissingValueToken is what text/template emits for an unresolvable reference.
// Named once because three things key on it — the strip, the count, and the
// tests — and a literal in three places is a rename away from a silent no-op.
const MissingValueToken = "<no value>"

// antiInventionRe marks a block whose data the prompt has declared authoritative.
//
// The marker set is MEASURED, not invented: across the 139 live prompt templates
// on 2026-09-03, these phrases occur 64 times, while the looser candidates I
// first considered — "exact" (161) and "verified" (73) — appear in 87 of 139
// templates and would have made the escalation fire on almost everything. A
// severity that fires on two thirds of the fleet is not a severity.
//
// Deliberately anti-INVENTION directives only. "Use the exact figures below" is
// an instruction about form; "do not invent" is an instruction about provenance,
// and provenance is what a manufactured stand-in token corrupts.
var antiInventionRe = regexp.MustCompile(`(?i)\b(do not invent|never invent|do not make up|never make up|use only these|only these|use only the)\b`)

// MissingValueReport is what one render produced.
//
// Occurrences and Authoritative are EXACT: both are counted in the rendered
// output, so neither can be wrong about whether it happened. Fields is
// BEST-EFFORT and may be short — a reference inside a {{range}} body resolves
// against the loop item, which this cannot see. Kept separate rather than
// merged so a reader is never invited to treat a short field list as a small
// problem.
type MissingValueReport struct {
	Occurrences   int
	Authoritative int
	// Fields are template paths that do not resolve against the render data,
	// dotted and sorted — "reviewed_brief.headquarters", not "reviewed_brief".
	Fields []string
	// Contexts are short excerpts around the authoritative occurrences, so a
	// log reader can see WHICH instruction block was corrupted without
	// re-rendering anything.
	Contexts []string
}

// Empty reports whether the render was clean.
func (r MissingValueReport) Empty() bool { return r.Occurrences == 0 }

// StripMissingValues removes the artefact and says how many it removed.
//
// Stripping is the same choice the component seam made, and for prompts the
// argument is stronger rather than weaker. On a PAGE an empty string is worse
// than a visible break, because a human reader cannot see what is absent — that
// is the LANDMINES warning on this class. In a PROMPT the reader is a model, and
// the two outcomes are not symmetrical: "Location: " says there is no location,
// which is true, while "Location: <no value>" asserts that the location IS the
// string "<no value>", which is false and sits inside an instruction block
// telling the model to trust it.
func StripMissingValues(rendered string) (string, int) {
	n := strings.Count(rendered, MissingValueToken)
	if n == 0 {
		return rendered, 0
	}
	return strings.ReplaceAll(rendered, MissingValueToken, ""), n
}

// ScanMissingValues attributes a render's holes.
//
// tpl and data are used only for the best-effort field names; the counts come
// from rendered, so passing a template that no longer matches the output can
// make Fields wrong but can never make Occurrences wrong.
func ScanMissingValues(tpl, rendered string, data map[string]interface{}) MissingValueReport {
	rep := MissingValueReport{Occurrences: strings.Count(rendered, MissingValueToken)}
	if rep.Occurrences == 0 {
		return rep
	}
	rep.Authoritative, rep.Contexts = authoritativeOccurrences(rendered)
	rep.Fields = unresolvedRootPaths(tpl, data)
	return rep
}

// authoritativeOccurrences counts the occurrences whose enclosing BLOCK carries
// an anti-invention directive, and returns an excerpt of each.
//
// The block is bounded by the nearest blank line before the occurrence, which is
// what makes this a proximity test rather than a document-level one: 87 of 139
// live templates carry such a directive SOMEWHERE, so "the template mentions it"
// would escalate almost every render. A paragraph is the unit the instruction
// actually governs — the worked case is a four-line contact block whose heading
// carries the directive and whose fourth line is the hole.
func authoritativeOccurrences(rendered string) (int, []string) {
	var (
		count    int
		contexts []string
		offset   int
	)
	for {
		i := strings.Index(rendered[offset:], MissingValueToken)
		if i < 0 {
			break
		}
		at := offset + i
		block := rendered[blockStart(rendered, at):at]
		if antiInventionRe.MatchString(block) {
			count++
			if len(contexts) < 3 { // enough to identify it; never the whole prompt
				contexts = append(contexts, TruncateString(strings.TrimSpace(collapseWhitespace(block)), 160))
			}
		}
		offset = at + len(MissingValueToken)
	}
	return count, contexts
}

// blockStart returns the index just after the blank line preceding at, or 0.
func blockStart(s string, at int) int {
	if j := strings.LastIndex(s[:at], "\n\n"); j >= 0 {
		return j + 2
	}
	return 0
}

func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// unresolvedRootPaths names the ROOT-SCOPE template paths that do not resolve
// against the render data.
//
// Root-scope only, and that is a correctness limit rather than laziness: inside
// {{range}} and {{with}} the dot is rebound to a value this function cannot see,
// so a reference there says nothing about the data map. Reporting those was the
// false-positive noise the council flagged on the sibling seam (a noisy channel
// is one nobody escalates on), and here it would name a field that is present
// and fine.
//
// The lookup deliberately mirrors text/template's own map indexing rather than
// calling ExtractNestedField: that helper is a RESOLVER with fallbacks — it
// retries under input_data, unwraps `response` envelopes and indexes arrays — so
// it answers "can this be found?" while the template asked "is this here?". Using
// it would report a path as resolvable that the template had already rendered as
// a hole, which is the one wrong answer this function must not give.
func unresolvedRootPaths(tpl string, data map[string]interface{}) []string {
	t, err := template.New("scan").Funcs(PromptTemplateFuncs()).Parse(tpl)
	if err != nil || t.Tree == nil || t.Tree.Root == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string

	consider := func(idents []string) {
		if len(idents) == 0 {
			return
		}
		path := strings.Join(idents, ".")
		if seen[path] {
			return
		}
		seen[path] = true
		if !pathResolves(data, idents) {
			out = append(out, path)
		}
	}

	var walkNode func(n parse.Node, dotIsRoot bool)
	var walkPipe func(p *parse.PipeNode, dotIsRoot bool)
	var walkArg func(n parse.Node, dotIsRoot bool)

	walkArg = func(n parse.Node, dotIsRoot bool) {
		switch a := n.(type) {
		case *parse.FieldNode:
			if dotIsRoot {
				consider(a.Ident)
			}
		case *parse.ChainNode:
			walkArg(a.Node, dotIsRoot)
		case *parse.PipeNode:
			walkPipe(a, dotIsRoot)
		}
	}
	walkPipe = func(p *parse.PipeNode, dotIsRoot bool) {
		if p == nil {
			return
		}
		for _, cmd := range p.Cmds {
			for _, arg := range cmd.Args {
				walkArg(arg, dotIsRoot)
			}
		}
	}
	walkNode = func(n parse.Node, dotIsRoot bool) {
		switch v := n.(type) {
		case *parse.ListNode:
			if v == nil {
				return
			}
			for _, c := range v.Nodes {
				walkNode(c, dotIsRoot)
			}
		case *parse.ActionNode:
			walkPipe(v.Pipe, dotIsRoot)
		case *parse.IfNode:
			walkPipe(v.Pipe, dotIsRoot)
			walkNode(v.List, dotIsRoot)
			walkNode(v.ElseList, dotIsRoot)
		case *parse.RangeNode:
			walkPipe(v.Pipe, dotIsRoot) // the ranged expression is outer scope
			walkNode(v.List, false)     // body: dot rebound to the element
			walkNode(v.ElseList, dotIsRoot)
		case *parse.WithNode:
			walkPipe(v.Pipe, dotIsRoot)
			walkNode(v.List, false)
			walkNode(v.ElseList, dotIsRoot)
		case *parse.TemplateNode:
			walkPipe(v.Pipe, dotIsRoot)
		}
	}
	walkNode(t.Tree.Root, true)
	sort.Strings(out)
	return out
}

// pathResolves walks idents through nested maps exactly as text/template indexes
// a map[string]interface{}. A present key holding nil does NOT resolve — that is
// what renders the token.
func pathResolves(data map[string]interface{}, idents []string) bool {
	var cur interface{} = data
	for _, id := range idents {
		m, ok := cur.(map[string]interface{})
		if !ok {
			// Not a map: text/template would try a struct field or method. This
			// package's render data is always a map[string]interface{}, so
			// anything else is out of scope and reported as resolvable rather
			// than guessed at — a false NEGATIVE, never a false accusation.
			return true
		}
		v, present := m[id]
		if !present || v == nil {
			return false
		}
		cur = v
	}
	return true
}
