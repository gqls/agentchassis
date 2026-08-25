// FILE: cmd/config-key-audit/unarmedcompleters.go
//
// The LIVE half of bugs_open/375 candidate 4: does livespec's declaration of
// unarmed `complete` arms still match production?
//
// WHY IT CANNOT BE A GO TEST. `go test` has no cluster. The set of steps that
// stamp `complete` through update_work_item_status lives entirely in
// agent_definitions.default_config — DB config, editable by a seed with no code
// change and no review, which is the property bugs_open/213 relies on when it
// refuses to enumerate producers in code. So the build-time lockstep
// (platform/orchestration/actions/unarmed_completer_lockstep_test.go) can check
// that every DECLARED arm is honest, and nothing more: a new agent completing
// through this writer is invisible to it. This mode is what makes the declaration
// go stale LOUDLY instead of silently.
//
// That is not a theoretical worry. It is the same criticism a council seat made of
// the first cut of verifier_coverage_test.go — a guard against "the mechanism
// relies on someone remembering" that itself relies on someone remembering — and
// the reason RFC_006 put the SingleOwner check here rather than in a hook.
//
// ⚠ TWO WAYS A NAIVE VERSION OF THIS GETS THE WRONG ANSWER, both measured
// 2026-08-24 and both guarded below:
//
//  1. `status` DEFAULTS TO `complete`. UpdateWorkItemStatusAction sets
//     newStatus := "complete" and only then reads config["status"]. A step that
//     omits the key is a `complete` arm, and a filter written as
//     config["status"] == "complete" cannot see it. All 22 live steps name it
//     today; this mode must be able to tell you the day one does not.
//  2. A TOP-LEVEL WALK MISSES NESTED STEPS. Steps sit inside sub_workflow configs
//     too — a $.workflow.steps scan finds only 2 of the 4 live complete_work_item
//     callers, because the dispatch loops carry theirs nested. validation.WalkSteps
//     is the shared traversal that sees both (bugs_open/144: two hand-written
//     descents go blind in different directions and then agree).
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/livespec"
	"github.com/gqls/agentchassis/platform/validation"
)

// unarmedCompleterFinding is one disagreement between the declaration and live
// config. Kind is "undeclared" (live, not declared — the dangerous direction) or
// "stale" (declared, not live — the tidy-up direction).
type unarmedCompleterFinding struct {
	Kind     string `json:"kind"`
	Agent    string `json:"agent"`
	Step     string `json:"step"`
	Path     string `json:"path,omitempty"`
	ItemType string `json:"item_type,omitempty"`
	Detail   string `json:"detail"`
}

// liveUnarmedCompleter is one live arm found in config.
type liveUnarmedCompleter struct {
	Agent string
	Step  string
	Path  string
}

// stepNameFromPath takes the last dot-segment of a WalkSteps path.
//
// The declaration names the STEP, not its full path, because that is what a human
// reads in a seed and what an operator arms. The full path is carried alongside so
// a finding can point at a nested arm precisely.
func stepNameFromPath(path string) string {
	if i := strings.LastIndex(path, "."); i >= 0 && i+1 < len(path) {
		return path[i+1:]
	}
	return path
}

// completesUnarmed reports whether a step stamps `complete` through
// update_work_item_status without consulting the verifier.
//
// The `status` default is the whole subtlety — see trap 1 in the file header. An
// absent key means `complete`, so the test is "not explicitly some OTHER status",
// never "explicitly complete".
func completesUnarmed(step models.Step) bool {
	if step.Action != "update_work_item_status" {
		return false
	}
	status := "complete"
	if s, ok := step.Config["status"].(string); ok && s != "" {
		status = s
	}
	if status != "complete" {
		return false
	}
	armed, _ := step.Config["verify_before_complete"].(bool)
	return !armed
}

// findUnarmedCompleterDrift is the pure check (see findSingleOwnerViolations for
// why the I/O is split off).
func findUnarmedCompleterDrift(agents []liveAgent, declared []livespec.UnarmedCompleter) []unarmedCompleterFinding {
	declaredBy := make(map[string]livespec.UnarmedCompleter, len(declared))
	for _, d := range declared {
		declaredBy[d.Agent+"."+d.Step] = d
	}

	live := map[string]liveUnarmedCompleter{}
	for _, a := range agents {
		validation.WalkSteps(a.Workflow, func(path string, step models.Step, nested bool) {
			if !completesUnarmed(step) {
				return
			}
			name := stepNameFromPath(path)
			live[a.Type+"."+name] = liveUnarmedCompleter{Agent: a.Type, Step: name, Path: path}
		})
	}

	var out []unarmedCompleterFinding

	// The dangerous direction FIRST: an arm production carries that nobody wrote
	// down. This is the one that fires when a new agent is seeded, and it is the
	// whole reason this mode exists.
	for key, l := range live {
		if _, ok := declaredBy[key]; ok {
			continue
		}
		out = append(out, unarmedCompleterFinding{
			Kind: "undeclared", Agent: l.Agent, Step: l.Step, Path: l.Path,
			Detail: "this step stamps `complete` through update_work_item_status without setting " +
				"verify_before_complete, and is NOT in livespec.UnarmedVerifiedCompleters. Add it, with the " +
				"item_type it completes and why it is unarmed — then the build-time lockstep can tell you " +
				"the day that item_type gains a verifier (bugs_open/375, WII-030)",
		})
	}

	// The tidy-up direction. Not merely cosmetic: a declared arm that no longer
	// exists still carries an Acknowledged reason, and that reason reads to the next
	// author as a live decision about a live arm.
	for key, d := range declaredBy {
		if _, ok := live[key]; ok {
			continue
		}
		out = append(out, unarmedCompleterFinding{
			Kind: "stale", Agent: d.Agent, Step: d.Step, ItemType: d.ItemType,
			Detail: "declared in livespec.UnarmedVerifiedCompleters but no live agent carries an unarmed " +
				"`complete` arm by that name — the step was renamed, removed, or has been ARMED. If it was " +
				"armed, delete the entry; leaving it asserts a bypass that no longer happens",
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind // "stale" < "undeclared" alphabetically; both printed
		}
		if out[i].Agent != out[j].Agent {
			return out[i].Agent < out[j].Agent
		}
		return out[i].Step < out[j].Step
	})
	return out
}

func emitUnarmedCompleterDrift() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unarmed-verified-completers: read stdin: %v\n", err)
		os.Exit(2)
	}
	agents, failed, err := decodeLiveAgents(raw, "--unarmed-verified-completers")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unarmed-verified-completers: %v\n", err)
		os.Exit(2)
	}
	// Exit 2, never a clean report, when the input could not be read: a mode whose
	// blind failure looks identical to "nothing is wrong" is the class this whole
	// directory exists to detect (--live-declaration-drift makes the same choice).
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr, "config-key-audit --unarmed-verified-completers: zero agents decoded (%d failed) — "+
			"refusing to report a clean result from an empty input\n", failed)
		os.Exit(2)
	}

	findings := findUnarmedCompleterDrift(agents, livespec.UnarmedVerifiedCompleters)
	if findings == nil {
		findings = []unarmedCompleterFinding{}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unarmed-verified-completers: encode: %v\n", err)
		os.Exit(2)
	}
	if failed > 0 {
		fmt.Fprintf(os.Stderr, "config-key-audit --unarmed-verified-completers: %d agent definition(s) failed to decode\n", failed)
	}
	if len(findings) > 0 {
		os.Exit(1)
	}
}
