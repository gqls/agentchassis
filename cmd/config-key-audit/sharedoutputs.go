// FILE: cmd/config-key-audit/sharedoutputs.go
//
// --shared-output-fields (RFC_012 (d), owner ruling 2026-08-06): find every
// pair of steps in one live workflow that write the SAME output_field, where
// the second is TRANSITIVELY REACHABLE from the first over the FULL routing
// graph, and their actions DIFFER.
//
// Every clause of that key is load-bearing, and both simplifications were
// tried and return 0 rows on a fleet that contained bugs_open/192 — a bug
// that had just taken every page build down (RFC_012 addendum 1):
//
//   - direct-edge only (a.next_step == b.name): 0 rows. Reachability is
//     transitive, never adjacent — 192's path was plan_sections -> ... ->
//     check_has_ready_sections -> load_current_section_content.
//
//   - transitive over next_step/error_step only: still 0 rows. The reaching
//     edge was config.then_step: a conditional routes through its CONFIG, and
//     routing lives in config for THIRTEEN distinct keys fleet-wide.
//
//   - and DIFFERENT action is the discriminator that turns 24 shared pairs
//     into 2 real hazards with 0 false negatives: a step that re-derives a
//     value with the SAME code cannot change its shape (the propose ->
//     repropose retry loops), so same-action pairs are structurally benign —
//     that is checkable, not taste (addendum 1's corrected table).
//
// Reachability across a nested sub-workflow is over-approximated: a nested
// step is treated as reachable from its containing step, and the container's
// successors from every nested step. Over-approximation errs toward a finding
// a human reads once; the two forbidden naive forms erred toward silence on a
// live outage, which is the worst possible failure for a check nobody
// re-reads.
//
// THE DESCENT IS validation.WalkSteps, NOT ITS OWN (changed 2026-08-08, and the
// reason is worth reading before you write another one). The council gate's
// reuse seat objected that this file walked agent_definitions routing config
// with a hand-written recursion while relaygaps.go — in THIS package — walks the
// same structure through validation.WalkSteps on purpose, citing bugs_open/144's
// rule that "a second hand-written descent goes blind in its own direction".
// The objection was right, and the justification this header used to give for
// the private descent ("the walker does not expose containment") was simply
// false: WalkSteps hands over a QUALIFIED PATH, and the container is the
// third-from-last segment of it, which is why containerOf below is four lines.
//
// It had gone blind, in the exact direction 144 predicts. A loop declares its
// body as EITHER `substeps` or `sub_workflow`, and `substeps` WINS at execution
// (loop_actions.go:91 reads it first and consults sub_workflow only when it is
// absent or empty — the precedence validation.subWorkflowsOf mirrors exactly).
// The private descent read `sub_workflow` ONLY. So it could not see the shape
// the runtime looks at first, and on a step carrying BOTH it walked the inert
// half — reporting a hazard in config that never executes.
//
// MEASURED before and after, because "0 findings" is the one result this class
// of bug produces for free: over live agent_definitions (is_active,
// non-snapshot, not deleted) with a recursive jsonpath
// `$.** ? (@.substeps != null)`, ZERO live definitions carry `substeps` at any
// depth on 2026-08-08 — the only two rows that do are soft-deleted
// multipage-website-builder definitions. The gap therefore cost nothing yet and
// no live run could ever have revealed it, which is why the proof is two tests
// (TestSharedOutputs_SubstepsShapeIsSeenToo and
// TestSharedOutputs_BothShapesResolveToTheExecutedOne) that both FAILED against
// the private descent and pass through the shared one.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/validation"
)

// routingConfigKeys is the config-level half of the routing graph, resolved
// against the LIVE fleet 2026-08-06 (13 keys; note error_step appears BOTH as
// a top-level Step field and as a config key — 158 config occurrences). If an
// action grows a new "*_step" config key, add it here; the count query is in
// the RUNBOOK of docs024_key_docs_latest/rfc012_await_findings/.
var routingConfigKeys = []string{
	"then_step", "else_step", "error_step",
	"repair_step", "ok_step", "emit_step",
	"alive_step", "unreachable_step", "lost_step",
	"failed_step", "gather_step", "complete_step", "probe_step",
}

type sharedOutputFinding struct {
	Agent       string `json:"agent"`
	OutputField string `json:"output_field"`
	Producer    string `json:"producer"` // step name that writes first
	ProducerAct string `json:"producer_action"`
	Refiner     string `json:"refiner"` // reachable later step that overwrites
	RefinerAct  string `json:"refiner_action"`
}

// graphStep is one step flattened out of the (possibly nested) workflow.
type graphStep struct {
	name      string
	action    string
	output    string
	succs     []string // routing edges by target step name
	container string   // non-empty for nested steps: the containing step's name
}

// flattenWorkflowGraph walks the workflow (nested sub-workflows included) and
// returns the step graph, through validation.WalkSteps — the SAME step set the
// runtime validator enforces against and relaygaps.go audits. See the header for
// what the private descent it replaced could not see.
//
// Two properties inherited from the shared walker, both of which the private
// descent had to get right by hand and got wrong: nested bodies are found in
// BOTH shapes with the executor's precedence, and each nested step is decoded by
// models.DecodeSubWorkflowStep — the function the loop action executes with —
// rather than by a JSON round-trip into models.Step. The round-trip populated
// fields the executor drops, so a detector reading one would have vouched for
// behaviour that does not happen. Every field this check reads (action, config,
// next_step, error_step, output_field) is in the honoured seven, so the switch
// costs nothing here and closes that door for whatever this file reads next.
func flattenWorkflowGraph(wf models.WorkflowPlan) []graphStep {
	var out []graphStep
	validation.WalkSteps(wf, func(path string, step models.Step, nested bool) {
		gs := graphStep{
			// Authored name, not the qualified path: routing edges name steps as
			// an author writes them. Two nested steps sharing an authored name at
			// different depths therefore collapse into one node — pre-existing
			// behaviour, preserved deliberately, and it over-approximates
			// reachability rather than under-approximating it.
			name:      stepName(path),
			action:    step.Action,
			output:    step.OutputField,
			container: containerOf(path),
		}
		if step.NextStep != "" {
			gs.succs = append(gs.succs, step.NextStep)
		}
		if step.ErrorStep != "" {
			gs.succs = append(gs.succs, step.ErrorStep)
		}
		for _, k := range routingConfigKeys {
			if v, ok := step.Config[k].(string); ok && v != "" {
				gs.succs = append(gs.succs, v)
			}
		}
		out = append(out, gs)
	})
	return out
}

// containerOf recovers the containing step's authored name from a WalkSteps
// path, which is what the old header wrongly claimed the walker could not give.
// A path is "steps.<name>" at the top level and
// "steps.<container>.<shape>.<name>" nested, where <shape> is substeps or
// sub_workflow, appending a further pair per level — so the container is always
// the third segment from the end, and its absence means top-level.
func containerOf(path string) string {
	parts := strings.Split(path, ".")
	if len(parts) < 4 {
		return ""
	}
	return parts[len(parts)-3]
}

// findSharedOutputFields is the pure check (I/O split off, per this binary's
// convention). One workflow at a time: build the adjacency, close it
// transitively, and report ordered (producer, refiner) pairs sharing an
// output_field with differing actions.
func findSharedOutputFields(agents []liveAgent) []sharedOutputFinding {
	var findings []sharedOutputFinding

	for _, agent := range agents {
		steps := flattenWorkflowGraph(agent.Workflow)
		if len(steps) == 0 {
			continue
		}
		idx := make(map[string]int, len(steps))
		for i, s := range steps {
			idx[s.name] = i
		}

		// Adjacency, including the nested over-approximation: container ->
		// nested step, and nested step -> container's successors.
		adj := make([][]int, len(steps))
		addEdge := func(from int, toName string) {
			if to, ok := idx[toName]; ok && to != from {
				adj[from] = append(adj[from], to)
			}
		}
		for i, s := range steps {
			for _, succ := range s.succs {
				addEdge(i, succ)
			}
		}
		for i, s := range steps {
			if s.container == "" {
				continue
			}
			if c, ok := idx[s.container]; ok {
				adj[c] = append(adj[c], i)
				for _, succ := range steps[c].succs {
					addEdge(i, succ)
				}
			}
		}

		// Transitive closure by BFS from each writer (fleet workflows are
		// tens of steps, not thousands — clarity over asymptotics).
		reach := func(from int) map[int]bool {
			seen := map[int]bool{}
			queue := append([]int(nil), adj[from]...)
			for len(queue) > 0 {
				n := queue[0]
				queue = queue[1:]
				if seen[n] {
					continue
				}
				seen[n] = true
				queue = append(queue, adj[n]...)
			}
			return seen
		}

		for i, a := range steps {
			if a.output == "" {
				continue
			}
			reachable := reach(i)
			for j, b := range steps {
				if i == j || b.output != a.output || b.action == a.action {
					continue
				}
				if reachable[j] {
					findings = append(findings, sharedOutputFinding{
						Agent:       agent.Type,
						OutputField: a.output,
						Producer:    a.name,
						ProducerAct: a.action,
						Refiner:     b.name,
						RefinerAct:  b.action,
					})
				}
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		if findings[i].OutputField != findings[j].OutputField {
			return findings[i].OutputField < findings[j].OutputField
		}
		return findings[i].Producer < findings[j].Producer
	})
	return findings
}

// findingKey is the acknowledgement identity: one line in the ack file names
// one (agent, output_field, producer, refiner) pair.
func (f sharedOutputFinding) findingKey() string {
	return f.Agent + " " + f.OutputField + " " + f.Producer + " " + f.Refiner
}

// loadAckList reads acknowledged findings, one findingKey per line, '#'
// comments and blank lines ignored. The file lives in-repo
// (scripts/shared_output_fields_ack.txt) so acknowledging a hazard is a
// reviewed change, and the standing job is a RATCHET: green until a NEW pair
// appears, never red forever over the two pairs the fleet already knows.
func loadAckList(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	acked := map[string]bool{}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		acked[line] = true
	}
	return acked, nil
}

// emitSharedOutputFields reads the same stdin export as the other modes; an
// optional --ack <file> names acknowledged findings. Exit codes follow
// --single-owner-actions: 2 on a refused (unusable) input, 1 on NEW findings
// (so a scheduled job shows failed), 0 on a meaningful pass. An acked entry
// that no longer reproduces is reported as stale — an ack list that outlives
// its findings is how a ratchet quietly loosens.
func emitSharedOutputFields() {
	acked := map[string]bool{}
	if len(os.Args) > 3 && os.Args[2] == "--ack" {
		var err error
		if acked, err = loadAckList(os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --shared-output-fields: reading ack file: %v\n", err)
			os.Exit(2)
		}
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --shared-output-fields: reading stdin: %v\n", err)
		os.Exit(2)
	}
	agents, failed, err := decodeLiveAgents(raw, "--shared-output-fields")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --shared-output-fields: 0 live agents decoded — refusing to print a clean report over an empty fleet.\n")
		os.Exit(2)
	}

	findings := findSharedOutputFields(agents)
	// Initialised, not nil: an empty result must marshal as [] — a consumer
	// len()-ing json null is a crash wearing a clean report's clothes.
	fresh := []sharedOutputFinding{}
	known := []sharedOutputFinding{}
	staleAcks := []string{}
	seen := map[string]bool{}
	for _, f := range findings {
		seen[f.findingKey()] = true
		if acked[f.findingKey()] {
			known = append(known, f)
		} else {
			fresh = append(fresh, f)
		}
	}
	for k := range acked {
		if !seen[k] {
			staleAcks = append(staleAcks, k)
		}
	}
	sort.Strings(staleAcks)

	out := map[string]interface{}{
		"agents_scanned":    len(agents),
		"agents_undecoded":  failed,
		"routing_keys_read": append([]string{"next_step", "error_step (top-level)"}, routingConfigKeys...),
		"findings_new":      fresh,
		"findings_acked":    known,
		"acks_stale":        staleAcks,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)

	if len(fresh) > 0 {
		fmt.Fprintf(os.Stderr, "%d NEW shared-output-field hazard(s): same output_field, transitively reachable, DIFFERENT action.\n", len(fresh))
		os.Exit(1)
	}
}
