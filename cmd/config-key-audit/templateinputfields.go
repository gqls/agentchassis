// FILE: cmd/config-key-audit/templateinputfields.go
//
// --template-input-fields (bugs_open/453): which live prompt template names a
// variable its step's `input_fields` can never supply?
//
// THE SEAM. A step that renders a prompt renders it against
// datahelpers.ExtractFields(CollectedData, input_fields) — a SUBSET of the run's
// collected data, holding only what input_fields asked for. A template variable
// whose ROOT is outside that subset is simply absent when the template executes,
// and Go's text/template has no opinion about that: a {{range}} or {{if}} over an
// absent key renders NOTHING and an unguarded {{.x}} renders "<no value>". The
// step succeeds. The output is a plausible prompt with a silently missing
// section. bugs_open/453 records four separate catches of this, every one found
// by a fixture somebody happened to write rather than by any guard.
//
// WHICH OF THE SEAM'S THREE SHAPES THIS CLOSES — stated because the 437 lane's
// contribution to 453 correctly objected that the original fix candidate did not
// say:
//
//  1. NO input_fields at all -> inputs resolve by randomised recursive search
//     (the LANDMINES sibling entry). NOT closed here; reported as context on the
//     finding (`no_input_fields`) and never convicted, because the search may
//     well find the value.
//  2. Root MISSING from input_fields -> the variable can never resolve on any
//     row. **THIS IS WHAT THE MODE CLOSES.**
//  3. Root PRESENT, sub-field absent in the row's data -> "<no value>".
//     NOT closed here, and no check over config can close it: the config is
//     correct and the outcome depends on a row. Measured by that lane at ~65% of
//     page-content-writer prompts. Its remedy is at the other end — promoting
//     RenderPromptTemplate's existing "<no value>" scan from a Warn to a durable
//     row — and belongs to whoever takes that.
//
// THE ROOT SET IS READ FROM THE CODE THAT OWNS IT, NEVER COPIED. 453's own fix
// candidate names the trap: "the extractor's speciallyHandled set must be read
// from ONE place or the lint inherits the classifier-gap problem". So the model
// of what a template can see is assembled by
// actions.TemplateRootsAvailableTo(action, input_fields), which composes
// datahelpers' half (always-ensured roots + the input_fields rule) with the
// action's half (what each action injects after extraction). This mode holds no
// list of its own. That is not fastidiousness: the first sizing pass for this
// check DID hold one global list, and reported both live execute_vision_prompt
// steps as broken because it did not know that action injects
// vision_image_manifest and does not inject the platform voice blocks — two
// false positives out of twelve findings.
//
// THE TEMPLATE IS PARSED, NOT REGEXED, and with production's own func map
// (datahelpers.PromptTemplateFuncs). text/template rebinds the dot inside
// {{range}} and {{with}}, so in {{range .items}}{{.name}}{{end}} the variable
// `name` is a field of the ITEM and says nothing about the root set. A regex
// over {{\.(\w+)}} convicts it. On this fleet's largest template that is dozens
// of false positives — enough noise to get the check ignored in its first week —
// and the parser gives the correct answer for free, along with $variables,
// parenthesised pipelines and function arguments.
//
// THREE FINDING KINDS, ONE OF WHICH FAILS THE RUN:
//
//   - unreachable_root  input_data is NOT among input_fields, so the root set is
//     fully determined by config and the variable resolves on no
//     row, ever. **Exit 1.**
//   - conditional_root  input_data IS among input_fields. ExtractFields promotes
//     every key of the runtime input_data map to the root, so
//     whether this resolves depends on a row this check cannot
//     see. Reported, never convicted. (Same treatment, and the
//     same reason, as defaultshadow.go's dotted_conditional.)
//   - declared_unread   the cheap reverse direction 453 also asked for: an
//     input_fields entry no template variable reads. Costs a
//     whole-tree extraction per run and usually means a template
//     lost a reference. Reported, never convicted.
//
// SCOPE IS THE ACTION, NOT THE PRESENCE OF A prompt_template KEY.
// actions.RendersPromptTemplate is the membership test, so a prompt_template
// sitting in the config of some action that never renders one is inert config
// rather than a broken template, and is not reported.
//
// TWO STATED LIMITATIONS, both invisible to config:
//
//   - Tier 1 of getPromptWithPriority lets a PARENT's call_agent pass a `prompt`
//     that outranks the step's own prompt_template. Where that happens the
//     template analysed here is not the one rendered. Nothing in a definition
//     records it.
//   - A root reported as reachable is only reachable to its FIRST segment. Shape
//     3 above is the rest of that sentence.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/template"
	"text/template/parse"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/validation"
)

// fleetExportQueryWithPrompts is --template-input-fields' export. It is
// fleetExportQuery plus one projection: the agent-level prompt_template, which
// is tier 2 of getPromptWithPriority and is the template 6 live steps actually
// render (measured 2026-09-03).
//
// Deliberately NOT a widening of the shared const. That one is pinned
// character-for-character to the query in several audit scripts, and every other
// mode ignores this key; changing it would churn files belonging to other lanes
// for no behaviour. The cost of the split is that an export produced for another
// mode lacks the key — which emitTemplateInputFields REFUSES on rather than
// silently going blind to tier 2. See requireAgentPromptProjection.
const fleetExportQueryWithPrompts = `
SELECT jsonb_agg(jsonb_build_object('type', type, 'workflow', default_config->'workflow',
                                    'agent_prompt_template', default_config->'prompt_template'))
FROM agent_definitions
WHERE deleted_at IS NULL
  AND COALESCE(is_snapshot,false) = false
  AND is_active
  AND default_config ? 'workflow';`

// templateInputFieldFinding is one (step, template) pair with something to say.
// Roots carries the offending identifiers: the unreachable/conditional template
// roots for those kinds, the unread input_fields entries for declared_unread.
type templateInputFieldFinding struct {
	Agent  string `json:"agent"`
	Path   string `json:"path"`
	Action string `json:"action"`
	// Tier is "step" (config.prompt_template) or "agent"
	// (default_config.prompt_template) — which template was analysed.
	Tier string `json:"template_tier"`
	Kind string `json:"kind"`
	// Roots is sorted, so a diff between two runs is a real change.
	Roots []string `json:"roots"`
	// InputFields is the step's declared list, echoed so a reader can judge the
	// finding without opening the definition. Nil when the step declares none.
	InputFields []string `json:"input_fields"`
	// NoInputFields marks shape 1 of the seam: the step declares none at all, so
	// extractDataForAiAgent defaults it to ["input_data"] and resolution is by
	// randomised recursive search. Context on the finding, never the finding.
	NoInputFields bool `json:"no_input_fields,omitempty"`
}

// templateInputFieldReport is the mode's whole output. ParseFailures is a
// first-class field rather than a stderr line because a template this binary
// cannot parse is a template it did not CHECK, and a run that silently checked
// fewer templates than it walked reads exactly like a clean one.
type templateInputFieldReport struct {
	AgentsScanned       int                         `json:"agents_scanned"`
	AgentsUndecoded     int                         `json:"agents_undecoded"`
	StepsWalked         int                         `json:"steps_walked"`
	TemplatesChecked    int                         `json:"templates_checked"`
	TemplatesAgentTier  int                         `json:"templates_agent_tier"`
	ParseFailures       []templateParseFailure      `json:"parse_failures"`
	Findings            []templateInputFieldFinding `json:"findings"`
	UnreachableFindings int                         `json:"unreachable_findings"`
}

type templateParseFailure struct {
	Agent string `json:"agent"`
	Path  string `json:"path"`
	Error string `json:"error"`
}

const (
	kindUnreachable    = "unreachable_root"
	kindConditional    = "conditional_root"
	kindDeclaredUnread = "declared_unread"
)

// templateRootsReferenced returns the roots a template reads off the TOP-LEVEL
// dot, using production's parser and production's func map.
//
// The dot-scope tracking is the whole point. Inside a {{range}} or {{with}}
// body the dot is a different value, so a FieldNode there names a field of that
// value and not a root; the ranged expression ITSELF is evaluated in the outer
// scope and is collected. {{if}} does not rebind, so its body stays in scope.
// $variables are VariableNodes and are correctly never roots.
func templateRootsReferenced(text string) (map[string]bool, error) {
	t, err := template.New("prompt").Funcs(datahelpers.PromptTemplateFuncs()).Parse(text)
	if err != nil {
		return nil, err
	}
	roots := map[string]bool{}
	if t.Tree == nil {
		return roots, nil
	}

	var walkNode func(n parse.Node, dotIsRoot bool)
	var walkPipe func(p *parse.PipeNode, dotIsRoot bool)
	var walkArg func(n parse.Node, dotIsRoot bool)

	walkArg = func(n parse.Node, dotIsRoot bool) {
		switch a := n.(type) {
		case *parse.FieldNode:
			if dotIsRoot && len(a.Ident) > 0 {
				roots[a.Ident[0]] = true
			}
		case *parse.ChainNode:
			// (expr).Field — the root, if any, is inside expr.
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
			// if does NOT rebind the dot.
			walkPipe(v.Pipe, dotIsRoot)
			walkNode(v.List, dotIsRoot)
			walkNode(v.ElseList, dotIsRoot)
		case *parse.RangeNode:
			walkPipe(v.Pipe, dotIsRoot) // ranged expression: outer scope
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
	return roots, nil
}

// stepInputFields reads the step's declared list under the action's own read: a
// []interface{} of strings, anything else ignored (extractDataForAiAgent's
// `fields, ok := ...([]interface{})` then `field, ok := ...(string)`).
// declared=false means the key is absent, which the action turns into
// ["input_data"] with a Warn.
func stepInputFields(cfg map[string]interface{}) (fields []string, declared bool) {
	raw, ok := cfg["input_fields"].([]interface{})
	if !ok {
		return nil, false
	}
	for _, f := range raw {
		if s, ok := f.(string); ok && s != "" {
			fields = append(fields, s)
		}
	}
	return fields, true
}

// auditTemplateInputFields is the pure half: agents in, report out, no I/O.
func auditTemplateInputFields(agents []liveAgent, undecoded int) templateInputFieldReport {
	rep := templateInputFieldReport{
		AgentsScanned:   len(agents),
		AgentsUndecoded: undecoded,
		ParseFailures:   []templateParseFailure{},
		Findings:        []templateInputFieldFinding{},
	}

	for _, agent := range agents {
		agentTier := agent.agentPromptTemplate()

		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, _ bool) {
			rep.StepsWalked++
			if !actions.RendersPromptTemplate(step.Action) {
				return
			}
			cfg := step.Config
			if cfg == nil {
				cfg = map[string]interface{}{}
			}

			// Tier 3 (the step's own) then tier 2 (the agent's), which is the
			// order getPromptWithPriority resolves them in once tier 1 is absent.
			tmplText, _ := cfg["prompt_template"].(string)
			tier := "step"
			if tmplText == "" {
				tmplText, tier = agentTier, "agent"
			}
			if tmplText == "" {
				return
			}
			rep.TemplatesChecked++
			if tier == "agent" {
				rep.TemplatesAgentTier++
			}

			referenced, err := templateRootsReferenced(tmplText)
			if err != nil {
				rep.ParseFailures = append(rep.ParseFailures, templateParseFailure{
					Agent: agent.Type, Path: path, Error: err.Error(),
				})
				return
			}

			fields, declared := stepInputFields(cfg)
			effective := fields
			if !declared {
				effective = []string{"input_data"} // extractDataForAiAgent's default
			}
			available, inputDataPromoted := actions.TemplateRootsAvailableTo(step.Action, effective)

			var missing []string
			for r := range referenced {
				if !available[r] {
					missing = append(missing, r)
				}
			}
			if len(missing) > 0 {
				sort.Strings(missing)
				kind := kindUnreachable
				if inputDataPromoted {
					kind = kindConditional
				} else {
					rep.UnreachableFindings++
				}
				rep.Findings = append(rep.Findings, templateInputFieldFinding{
					Agent: agent.Type, Path: path, Action: step.Action, Tier: tier,
					Kind: kind, Roots: missing, InputFields: fields, NoInputFields: !declared,
				})
			}

			// Reverse direction. input_data is skipped: it is the promotion
			// switch, not a variable — a template naming none of its keys by name
			// still relies on it for domain/objective/model.
			var unread []string
			for _, f := range fields {
				if f == "input_data" {
					continue
				}
				if !referenced[datahelpers.TemplateRootForInputField(f)] {
					unread = append(unread, f)
				}
			}
			if len(unread) > 0 {
				sort.Strings(unread)
				rep.Findings = append(rep.Findings, templateInputFieldFinding{
					Agent: agent.Type, Path: path, Action: step.Action, Tier: tier,
					Kind: kindDeclaredUnread, Roots: unread, InputFields: fields,
				})
			}
		})
	}

	sort.Slice(rep.Findings, func(i, j int) bool {
		a, b := rep.Findings[i], rep.Findings[j]
		if a.Kind != b.Kind {
			// unreachable_root first: it is the only kind that fails the run.
			return kindRank(a.Kind) < kindRank(b.Kind)
		}
		if a.Agent != b.Agent {
			return a.Agent < b.Agent
		}
		return a.Path < b.Path
	})
	return rep
}

func kindRank(kind string) int {
	switch kind {
	case kindUnreachable:
		return 0
	case kindConditional:
		return 1
	default:
		return 2
	}
}

// requireAgentPromptProjection refuses an export that does not carry the
// agent_prompt_template key at all.
//
// jsonb_build_object emits the key with a JSON null when the agent has no
// prompt_template, so a row carrying it is the normal case and "no row carries
// it" can only mean the export came from a query that does not project it —
// another mode's script, most likely. Without this the mode would run happily
// and be blind to tier 2, reporting fewer findings and looking clean. That is
// the exact shape this binary exists to refuse.
func requireAgentPromptProjection(agents []liveAgent) error {
	for _, a := range agents {
		if len(a.AgentPromptTemplate) > 0 {
			return nil
		}
	}
	return fmt.Errorf("config-key-audit --template-input-fields: no agent row carried an "+
		"'agent_prompt_template' key across %d agents — this export cannot see the agent-level "+
		"prompt tier, and a run over it would look clean while blind to it.\n"+
		"Use scripts/audit-template-input-fields.sh, or an export whose jsonb_build_object "+
		"projects 'agent_prompt_template', default_config->'prompt_template'.", len(agents))
}

// emitTemplateInputFields is the I/O half: DB route when PG_CLIENTS_HOST is set,
// stdin otherwise. Same refusals as every mode here.
func emitTemplateInputFields(args []string) {
	report := false
	for _, a := range args {
		switch a {
		case "--report":
			report = true
		default:
			fmt.Fprintf(os.Stderr, "config-key-audit --template-input-fields: unknown argument %q\n", a)
			os.Exit(2)
		}
	}

	var agents []liveAgent
	var failed int
	var err error
	if db, derr := dbConn(); derr != nil || db != nil {
		if derr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --template-input-fields: %v\n", derr)
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDBWithQuery(db, "--template-input-fields", fleetExportQueryWithPrompts)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	} else {
		raw, rerr := io.ReadAll(os.Stdin)
		if rerr != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --template-input-fields: reading stdin: %v\n", rerr)
			os.Exit(2)
		}
		agents, failed, err = decodeLiveAgents(raw, "--template-input-fields")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --template-input-fields: 0 live agents decoded — refusing to print a clean report over an empty fleet.\n")
		os.Exit(2)
	}
	if err := requireAgentPromptProjection(agents); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	rep := auditTemplateInputFields(agents, failed)
	if rep.TemplatesChecked == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --template-input-fields: 0 of %d steps across %d agents rendered a prompt_template — "+
				"refusing to report a clean fleet over zero templates (bugs_open/453 §How to verify).\n",
			rep.StepsWalked, len(agents))
		os.Exit(2)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(rep)

	if report {
		writeDocNote("template-input-fields", templateInputFieldRunSummary(rep),
			"config-integrity", "template_input_field_check")
	}

	if len(rep.ParseFailures) > 0 {
		fmt.Fprintf(os.Stderr, "%d template(s) could not be parsed and were NOT checked.\n", len(rep.ParseFailures))
	}
	if rep.UnreachableFindings > 0 {
		fmt.Fprintf(os.Stderr,
			"%d template variable set(s) can never resolve: the root is absent from the step's input_fields "+
				"and input_data is not promoted, so the block renders empty on every row (bugs_open/453).\n",
			rep.UnreachableFindings)
		os.Exit(1)
	}
}

// templateInputFieldRunSummary is the doc_notes body. It states the SCOPE as
// well as the result, because "0 findings over 4 templates" and "0 findings over
// 139" have opposite meanings (sharedOutputRunSummary's convention).
func templateInputFieldRunSummary(rep templateInputFieldReport) string {
	var b strings.Builder
	var unreachable, conditional, unread []templateInputFieldFinding
	for _, f := range rep.Findings {
		switch f.Kind {
		case kindUnreachable:
			unreachable = append(unreachable, f)
		case kindConditional:
			conditional = append(conditional, f)
		default:
			unread = append(unread, f)
		}
	}

	if len(unreachable) == 0 {
		fmt.Fprintf(&b, "template-input-fields check CLEAN: every template variable across %d live prompt templates "+
			"(%d of them agent-tier), on %d agents, has a root its step's input_fields can supply.",
			rep.TemplatesChecked, rep.TemplatesAgentTier, rep.AgentsScanned)
	} else {
		fmt.Fprintf(&b, "template-input-fields: %d of %d live prompt templates name a variable that can NEVER resolve "+
			"(bugs_open/453): ", len(unreachable), rep.TemplatesChecked)
		for i, f := range unreachable {
			if i > 0 {
				b.WriteString("; ")
			}
			fmt.Fprintf(&b, "%s %s wants {{.%s}}", f.Agent, f.Path, strings.Join(f.Roots, "}}, {{."))
		}
		b.WriteString(".")
	}
	fmt.Fprintf(&b, " Advisory: %d conditional (input_data promoted, undecidable from config), %d declared-but-unread input_fields entries.",
		len(conditional), len(unread))
	if len(rep.ParseFailures) > 0 {
		fmt.Fprintf(&b, " %d template(s) FAILED TO PARSE and were not checked.", len(rep.ParseFailures))
	}
	if rep.AgentsUndecoded > 0 {
		fmt.Fprintf(&b, " %d agent row(s) failed to decode and were not scanned.", rep.AgentsUndecoded)
	}
	return b.String()
}
