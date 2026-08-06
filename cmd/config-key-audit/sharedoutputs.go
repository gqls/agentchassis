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
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
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
// returns the step graph. Its own recursion rather than validation.WalkSteps
// because the EDGES of a nested step matter here, not just its presence, and
// the walker does not expose containment.
func flattenWorkflowGraph(wf models.WorkflowPlan) []graphStep {
	var out []graphStep
	var walk func(steps map[string]models.Step, container string)
	walk = func(steps map[string]models.Step, container string) {
		for name, step := range steps {
			gs := graphStep{
				name:      name,
				action:    step.Action,
				output:    step.OutputField,
				container: container,
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
			if sub := subWorkflowOf(step); sub != nil {
				walk(sub, name)
			}
		}
	}
	walk(wf.Steps, "")
	return out
}

// subWorkflowOf extracts a nested sub-workflow's steps from a step's config,
// tolerating both the typed and the raw-map shapes.
func subWorkflowOf(step models.Step) map[string]models.Step {
	raw, ok := step.Config["sub_workflow"]
	if !ok {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	var sub struct {
		Steps map[string]models.Step `json:"steps"`
	}
	if err := json.Unmarshal(b, &sub); err != nil || len(sub.Steps) == 0 {
		return nil
	}
	return sub.Steps
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
