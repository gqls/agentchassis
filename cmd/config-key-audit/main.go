// FILE: cmd/config-key-audit/main.go
//
// Dumps the config-key contract every action has DECLARED, as JSON.
//
// This exists because of bugs_open/101: an unknown step-config key is silently
// ignored at execution, so a stale or aspirational key is indistinguishable by
// inspection from a live one. The runtime validator reports unknown keys for
// actions that have opted in (ActionInputSpec.ConfigKeys), but that only covers
// steps that actually RUN, and only for actions someone has already declared.
//
// The declarations are compiled into the binary — they are Go, registered by
// init() — so the only honest way to compare them against what live
// agent_definitions carry is to ask the binary. That is all this does; the join
// against the database is scripts/audit-config-keys.sh.
//
// Usage:
//
//	go run ./cmd/config-key-audit
//	  {"declared":    {"<action>": ["<key>", ...], ...},
//	   "conditional": {"<action>": {"<key>": "<the condition>", ...}, ...}}
//
//	go run ./cmd/config-key-audit --specs
//	  {"<action>": {"required": [...], "optional": [...], "config_keys": [...],
//	                "deprecated": [...], "opted_in": bool}, ...}
//
// --specs answers the ADOPTION question rather than the coverage one: not "who has
// opted in?" but "who is one line away?". An action whose spec already enumerates
// every key its live steps carry can opt in with `CheckConfig: true` and assert
// nothing new about behaviour; one whose spec is missing keys, or which has no spec
// at all, needs reading first. Declaring a key an action does NOT read is worse than
// declaring nothing — it silences the detector for a dead key (WRONG_CALLS.md
// 2026-07-28) — so that distinction is the whole cost model of the ratchet.
//
// It is a FLAG rather than a second binary because the council gate's editquality
// and reuse_agent seats both objected to the second binary independently
// (2026-07-28, corr 07cf67c6): same registry API, same struct, same init-side-effect
// import, overlapping question. They were right — "two code paths independently
// solving overlapping problems that nobody unifies once both exist" is a named
// failure pattern in this estate. The DEFAULT output is unchanged byte-for-byte, so
// scripts/audit-config-keys.sh is unaffected.
//
// TWO maps, not one merged list. "declared" is every key an action recognises;
// "conditional" is the subset honoured only under a stated condition. Merging
// them would recreate exactly the blindness the second map was added to fix —
// a key can be perfectly well declared and still describe behaviour a given step
// never performs (bugs_closed/101; WRONG_CALLS.md 2026-07-28).
//
// The only consumer of this output is scripts/audit-config-keys.sh — checked, not
// assumed, when the shape changed: `grep -rn "config-key-audit"` over the tree
// returns that script and this file's own comments, nothing else.
//
// Deliberately imports the actions package for its registration side effects and
// nothing else. If that import is ever dropped the output becomes an empty
// object, which would read as "nothing is declared" rather than failing — so the
// tool refuses to print an empty result instead.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/gqls/agentchassis/pkg/models"

	// Imported for its init() side effects: every action registers its
	// ActionInputSpec here, AND registers the actioncheck.IsLocalAction checker
	// that --unregistered-actions depends on (registry.go's init calls
	// actioncheck.RegisterLocalActionChecker). Drop this import and both modes
	// go quiet in different ways: ConfigKeys reports empty, IsLocalAction reports
	// "not local" for every action — which is why both refuse to print a
	// clean-looking result over zero registrations.
	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	_ "github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/validation"
)

// liveAgent is one row of the export both --live-pairs and --unregistered-actions
// read on stdin: [{"type": "<agent>", "workflow": {...}}, ...]. Shared so the two
// questions ("what (action,key) pairs exist" and "what action is unrunnable") are
// asked of the exact same decode, not two copies that could drift.
type liveAgent struct {
	Type     string              `json:"type"`
	Workflow models.WorkflowPlan `json:"workflow"`
}

// decodeLiveAgents parses the stdin export. One undecodable definition costs that
// definition, not the whole report — see emitLivePairs' original comment on why
// failures are counted and printed rather than swallowed.
func decodeLiveAgents(raw []byte, caller string) (agents []liveAgent, failed int, err error) {
	var rows []json.RawMessage
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, 0, fmt.Errorf("stdin is not a JSON array of agents: %w\n"+
			"A truncated kubectl exec exits 0, so a short read arrives here looking like a small fleet.", err)
	}
	for _, row := range rows {
		var agent liveAgent
		if jsonErr := json.Unmarshal(row, &agent); jsonErr != nil {
			failed++
			fmt.Fprintf(os.Stderr, "config-key-audit %s: undecodable agent row: %v\n", caller, jsonErr)
			continue
		}
		agents = append(agents, agent)
	}
	return agents, failed, nil
}

// specDump is the --specs shape: an action's FULL declared contract, not just the
// key set the coverage report joins on.
type specDump struct {
	Required   []string `json:"required"`
	Optional   []string `json:"optional"`
	ConfigKeys []string `json:"config_keys"`
	Deprecated []string `json:"deprecated"`
	OptedIn    bool     `json:"opted_in"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--specs" {
		emitSpecs()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--live-pairs" {
		emitLivePairs()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--unregistered-actions" {
		emitUnregisteredActions()
		return
	}
	declared := datahelpers.ListDeclaredConfigKeys()
	conditional := datahelpers.ListConditionalConfigKeys()

	if len(declared) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit: no action declares ConfigKeys.\n"+
				"That is either true (nothing has opted in yet) or the actions package "+
				"is no longer linked in, in which case this tool is silently reporting "+
				"nothing rather than failing. Check the blank import above before "+
				"believing an empty result.")
		os.Exit(2)
	}

	// Two maps, not one: "every key this action recognises" and "of those, the
	// ones honoured only under a condition". Merging them would recreate exactly
	// the blindness this second map was added to fix.
	out := map[string]interface{}{
		"declared":    declared,
		"conditional": conditional,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit: %v\n", err)
		os.Exit(1)
	}
}

// emitSpecs dumps every registered action's full ActionInputSpec.
//
// Same refusal-on-empty as the default path, and for the same reason: a dropped
// blank import would make this print "{}" — read as "no action declares anything",
// which is a coverage gap of exactly the wrong sign — rather than failing.
func emitSpecs() {
	names := datahelpers.ListActionInputSpecNames()
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --specs: no action registered an ActionInputSpec.\n"+
				"That is either true or the actions package is no longer linked in, in "+
				"which case this would report an empty gap rather than failing. Check "+
				"the blank import before believing it.")
		os.Exit(2)
	}

	// opted_in is taken from datahelpers' OWN gate rather than recomputed here.
	// ListDeclaredConfigKeys already returns exactly the opted-in actions, so
	// membership of it IS the answer. Re-deriving `CheckConfig || len(ConfigKeys)>0`
	// in this file would create a fourth copy of a rule that until today had three,
	// and those three silently disagreeing is the defect class this whole tool
	// exists to surface — writing a fresh one here would be the tool committing it.
	optedIn := datahelpers.ListDeclaredConfigKeys()

	out := make(map[string]specDump, len(names))
	for _, name := range names {
		spec, ok := datahelpers.GetActionInputSpec(name)
		if !ok {
			continue
		}
		dep := make([]string, 0, len(spec.Deprecated))
		for k := range spec.Deprecated {
			dep = append(dep, k)
		}
		sort.Strings(dep)
		_, isOptedIn := optedIn[name]
		out[name] = specDump{
			Required:   nonNil(spec.Required),
			Optional:   nonNil(spec.Optional),
			ConfigKeys: nonNil(spec.ConfigKeys),
			Deprecated: dep,
			OptedIn:    isOptedIn,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --specs: %v\n", err)
		os.Exit(1)
	}
}

// emitLivePairs reads an export of live agent workflows on stdin and prints every
// (action, config-key) pair they carry, as TSV: action, key, "nested"|"top".
//
// WHY THIS IS A GO MODE RATHER THAN THE SQL IT REPLACES (bugs_open/144). The audit's
// query walked `default_config->'workflow'->'steps'` — the top level only, exactly
// like the runtime validator did. Two halves blind in the same direction agree with
// each other, and consistent blindness reads exactly like correctness: 25 (action,
// key) pairs existed ONLY inside loop sub-workflows and were invisible to both
// (measured 2026-07-29). Rewriting the SQL to descend would have fixed today's
// numbers and left two hand-written traversals to drift apart again. This walks the
// definition with validation.WalkSteps — the SAME traversal the validator enforces
// against — so "what the audit sees" and "what is validated" cannot diverge without
// a compile error.
//
// A flag rather than a second binary, on the precedent recorded in this file's
// header: the council gate objected to a second binary over the same registry and
// the same question, and was right.
//
// Input shape (see scripts/audit-config-keys.sh):
//
//	[{"type": "<agent>", "workflow": {"start_step": ..., "steps": {...}}}, ...]
func emitLivePairs() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --live-pairs: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--live-pairs")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --live-pairs: %v\n", err)
		os.Exit(2)
	}

	// Walked per agent: one undecodable definition (counted above, in failed) must
	// cost that definition, not the whole report. A report that quietly covers
	// less than it claims is the defect this tool exists to expose.
	pairs := map[string]bool{}
	decoded := len(agents)
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(_ string, step models.Step, nested bool) {
			if step.Action == "" {
				return
			}
			where := "top"
			if nested {
				where = "nested"
			}
			for key := range step.Config {
				pairs[step.Action+"\t"+key+"\t"+where] = true
			}
		})
	}

	if len(pairs) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --live-pairs: no (action, key) pairs found in %d decoded agents (%d undecodable).\n"+
				"That is either an empty fleet or a broken export — refusing rather than printing "+
				"an empty report that reads as 'nothing to fix'.\n", decoded, failed)
		os.Exit(2)
	}

	out := make([]string, 0, len(pairs))
	for p := range pairs {
		out = append(out, p)
	}
	sort.Strings(out)
	for _, p := range out {
		fmt.Println(p)
	}
	fmt.Fprintf(os.Stderr, "config-key-audit --live-pairs: %d agents decoded (%d undecodable), %d distinct pairs\n",
		decoded, failed, len(pairs))
}

// unregisteredActionFinding names one step whose action neither the local
// registry nor a topic can dispatch — the exact condition
// platform/validation/workflow.go's validateStep and subworkflow.go's
// validateSubWorkflowStep reject a MESSAGE on ("requires a topic"). This makes the
// same check offline (bugs_open/148): the runtime validator only speaks once a
// message arrives at a structurally unrunnable step, so a definition can carry
// the defect for weeks with nothing reporting it.
type unregisteredActionFinding struct {
	Agent  string `json:"agent"`
	Path   string `json:"path"`
	Action string `json:"action"`
	Nested bool   `json:"nested"`
}

// findUnregisteredActions is the pure check, separated from I/O so it can be
// exercised by a fixture test without stdin/stdout plumbing. It walks with
// validation.WalkSteps — the SAME traversal the runtime validator enforces
// against, top-level and nested — for the reason recorded on that function:
// bugs_open/144 was two hand-written traversals blind in the same direction,
// agreeing with each other. A step with a topic is legitimately remote and is
// NOT a finding; only the unregistered-and-untopic'd pair is (bugs_open/148 §6).
func findUnregisteredActions(agents []liveAgent) []unregisteredActionFinding {
	findings := []unregisteredActionFinding{}
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.Action == "" {
				return
			}
			// Mirrors platform/validation/workflow.go's validateStep and
			// subworkflow.go's validateSubWorkflowStep EXACTLY — including their
			// choice not to exempt "fan_out": if a live fan_out step is neither
			// locally registered nor topic-routed, the real validator rejects it
			// too, and this tool exists to say so before a message finds out.
			if actioncheck.IsLocalAction(step.Action) || step.Topic != "" {
				return
			}
			findings = append(findings, unregisteredActionFinding{
				Agent: agent.Type, Path: path, Action: step.Action, Nested: nested,
			})
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Agent != findings[j].Agent {
			return findings[i].Agent < findings[j].Agent
		}
		return findings[i].Path < findings[j].Path
	})
	return findings
}

// emitUnregisteredActions reads the same stdin shape as --live-pairs and prints
// every step that findUnregisteredActions flags, as JSON. Zero findings is a
// legitimate, meaningful result (unlike an empty decoded-agent set, which is
// refused below) — it says the fleet currently has none, not that the check
// didn't run.
func emitUnregisteredActions() {
	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--unregistered-actions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --unregistered-actions: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	findings := findUnregisteredActions(agents)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "config-key-audit --unregistered-actions: %d agents decoded (%d undecodable), %d finding(s)\n",
		len(agents), failed, len(findings))
	if len(findings) > 0 {
		os.Exit(1)
	}
}

// nonNil keeps the JSON shape stable: a missing list encodes as [] rather than
// null, so a consumer can iterate without a nil check and cannot mistake
// "declares nothing" for "key absent from the dump".
func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
