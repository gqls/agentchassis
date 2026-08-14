// FILE: cmd/config-key-audit/optionalbudget.go
//
// RFC 022, owner ruling 2026-08-11. Which SHARED action's optional-key set has
// accumulated past the budget?
//
// This is the counter that closes RFC 022. The owner's ruling was option (3) —
// the architecture seat's trigger moves from "any new reserved key on a shared
// action" to the ACCUMULATED optional-key count — with option (1) as the
// interim. The interim deliberately gave up the accumulation signal: ten
// opt-in fields, each individually inert and each individually exempt, are a
// shared action nobody understands, and the per-change trigger was the only
// thing that would have noticed the tenth. This mode notices the tenth.
//
// What is counted is exactly what the RFC names: the Optional list of each
// registered ActionInputSpec ("a sweep over RegisterActionInputSpec
// declarations per action"). ConfigKeys is deliberately NOT counted: those are
// settings rather than input references (see ActionInputSpec's own comment),
// and the harm the seat named — new authority arriving as opt-in fields — lands
// in Optional, where bugs_open/223's note_body_suffix_field (the motivating
// case) lives. Counting both would let a settings-heavy action trip a budget
// meant for accumulated authority.
//
// "Shared" is measured, not declared: distinct live agents carrying the action,
// counted from the same stdin export every other live-join mode reads, walked
// with validation.WalkSteps for the bugs_open/144 reason (a hand-written
// descent goes blind on `substeps`, and 25 real pairs live only inside loop
// sub-workflows). An action with fewer than two live carriers can carry any
// number of optional keys without tripping the budget — the seat's harm is a
// SHARED seam nobody understands, and a single-consumer action's surface is
// that consumer's own business. The census still lists it, so the reader sees
// the whole distribution and not just the findings.
//
// The budget itself is an argument, not a constant, because the threshold is
// the owner's to rule (RFC 022 §"The three options, costed": "an RFC when a
// shared action's optional-key set grows past N" — N is a governance choice,
// not a technical finding). Run without a budget, this is a report-only census
// and always exits 0.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/validation"
)

// optionalKeyCensusRow is one action's accumulated optional-key surface beside
// its live consumer count — the two numbers the RFC_022 trigger joins.
type optionalKeyCensusRow struct {
	Action       string   `json:"action"`
	OptionalKeys int      `json:"optional_keys"`
	Optional     []string `json:"optional"`
	Consumers    int      `json:"consumers"`
	Agents       []string `json:"agents"`
	// OverBudget is only ever true when a budget was given AND the action is
	// shared (Consumers >= 2). It is a field rather than a filtered list so a
	// consumer of the JSON sees the near-misses in the same shape as the
	// findings — the tenth field is the trigger, but the ninth is the warning.
	OverBudget bool `json:"over_budget"`
}

// censusOptionalKeys is the pure half (same split as findSingleOwnerViolations,
// for the same testability reason). budget < 0 means "no budget": census only,
// nothing can be over budget.
//
// Counts DISTINCT AGENTS per action, not steps — one agent calling an action
// from three steps is one consumer's design, not three parties to a contract.
// Actions with a spec but zero optional keys are omitted (they have no surface
// to accumulate); actions carried live but registering no spec are omitted too
// (they declare nothing to count — --unregistered-actions owns that class).
func censusOptionalKeys(agents []liveAgent, budget int) []optionalKeyCensusRow {
	carriers := make(map[string][]string) // action -> sorted distinct agent types
	seen := make(map[string]bool)         // action+"\x00"+agent
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.Action == "" {
				return
			}
			key := step.Action + "\x00" + agent.Type
			if seen[key] {
				return
			}
			seen[key] = true
			carriers[step.Action] = append(carriers[step.Action], agent.Type)
		})
	}

	var rows []optionalKeyCensusRow
	for _, name := range datahelpers.ListActionInputSpecNames() {
		spec, ok := datahelpers.GetActionInputSpec(name)
		if !ok || len(spec.Optional) == 0 {
			continue
		}
		agentsFor := append([]string(nil), carriers[name]...)
		sort.Strings(agentsFor)
		row := optionalKeyCensusRow{
			Action:       name,
			OptionalKeys: len(spec.Optional),
			Optional:     nonNil(spec.Optional),
			Consumers:    len(agentsFor),
			Agents:       agentsFor,
		}
		if budget >= 0 && row.Consumers >= 2 && row.OptionalKeys > budget {
			row.OverBudget = true
		}
		rows = append(rows, row)
	}

	// Findings first, then the widest surfaces, then name — so the top of the
	// report is always the answer to "what should an architecture round look
	// at next", with or without a budget.
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].OverBudget != rows[j].OverBudget {
			return rows[i].OverBudget
		}
		if rows[i].OptionalKeys != rows[j].OptionalKeys {
			return rows[i].OptionalKeys > rows[j].OptionalKeys
		}
		return rows[i].Action < rows[j].Action
	})
	return rows
}

// emitOptionalKeyBudget reads the same stdin shape as --single-owner-actions
// and prints {"budget": N|null, "actions": [...]}.
//
// Refuses over an empty registry and over an empty export, for the reasons on
// emitSingleOwnerViolations — a clean report that no input could ever fail is
// indistinguishable from a broken build. NOTE for wrapper authors: `go run`
// collapses every non-zero child status to 1 (LANDMINES.md), so a wrapper must
// discriminate the refusal by its EMPTY STDOUT, never by exit code.
func emitOptionalKeyBudget(args []string) {
	budget := -1
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 0 {
			fmt.Fprintf(os.Stderr,
				"config-key-audit --optional-key-budget: budget must be a non-negative integer, got %q\n", args[0])
			os.Exit(2)
		}
		budget = n
	}

	names := datahelpers.ListActionInputSpecNames()
	if len(names) == 0 {
		fmt.Fprintln(os.Stderr,
			"config-key-audit --optional-key-budget: no action registered an ActionInputSpec — "+
				"refusing to print a census no fleet could ever fail. Check the blank import.")
		os.Exit(2)
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --optional-key-budget: reading stdin: %v\n", err)
		os.Exit(2)
	}
	agents, failed, err := decodeLiveAgents(raw, "--optional-key-budget")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --optional-key-budget: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --optional-key-budget: 0 agents decoded (%d undecodable) — "+
				"refusing to print a census over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	rows := censusOptionalKeys(agents, budget)
	out := struct {
		Budget  *int                   `json:"budget"`
		Actions []optionalKeyCensusRow `json:"actions"`
	}{Actions: rows}
	if budget >= 0 {
		out.Budget = &budget
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --optional-key-budget: %v\n", err)
		os.Exit(1)
	}
	for _, r := range rows {
		if r.OverBudget {
			os.Exit(1)
		}
	}
}
