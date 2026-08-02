// FILE: cmd/config-key-audit/singleowner.go
//
// RFC 006, owner ruling 2026-08-02. Which action that DECLARES single-ownership
// is carried by more than one live agent?
//
// In its own file rather than main.go for the reason main.go's header gives
// about --specs: modes here are separable questions over a shared decode, and
// keeping each mode's logic beside its own tests is what stops main.go becoming
// the place every mode collides. (It is also, prosaically, what lets two
// concurrent sessions add a mode each without one taking the other's
// half-finished work into its commit.)
//
// See datahelpers.ActionInputSpec.SingleOwner for why the check has to be
// offline: an action cannot tell at execution time whether a SIBLING agent also
// carries it, so nothing inside a run can ever detect this class.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/gqls/agentchassis/pkg/models"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/validation"
)

// singleOwnerViolation is one action declared SingleOwner that more than one
// live agent carries. Owners lists EVERY carrier, not just the extras, because
// the interesting question when this fires is "which of these is the owner?" —
// and answering it needs the whole set, not a diff against an arbitrary first.
type singleOwnerViolation struct {
	Action string   `json:"action"`
	Owners []string `json:"owners"`
	Paths  []string `json:"paths"`
}

// findSingleOwnerViolations is the pure check (see findUnregisteredActions for
// why the I/O is split off). Same validation.WalkSteps traversal as every other
// mode here, top-level and nested, for the bugs_open/144 reason: two
// hand-written descents go blind in different directions and then agree.
//
// Counts DISTINCT AGENTS, not steps. One agent carrying the step twice is a
// different defect (a duplicated step within one workflow) and not this one —
// the ownership question is about which agent owns the effect, and reporting a
// within-agent repeat here would put a finding of a different class under this
// heading. Paths are still collected in full so the report can point at every
// site, which is what makes a finding actionable without a second query.
//
// A declared action that NO live agent carries is not a finding: an action can
// be registered and unused, and reporting that as a violation would make the
// detector fire on the healthiest possible state.
func findSingleOwnerViolations(agents []liveAgent, declared []string) []singleOwnerViolation {
	isSingleOwner := make(map[string]bool, len(declared))
	for _, a := range declared {
		isSingleOwner[a] = true
	}

	owners := make(map[string][]string)
	paths := make(map[string][]string)
	seen := make(map[string]bool) // action+"\x00"+agent — dedupes within one agent

	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.Action == "" || !isSingleOwner[step.Action] {
				return
			}
			paths[step.Action] = append(paths[step.Action], agent.Type+"."+path)
			key := step.Action + "\x00" + agent.Type
			if seen[key] {
				return
			}
			seen[key] = true
			owners[step.Action] = append(owners[step.Action], agent.Type)
		})
	}

	findings := []singleOwnerViolation{}
	for action, carriers := range owners {
		if len(carriers) < 2 {
			continue
		}
		sort.Strings(carriers)
		p := append([]string(nil), paths[action]...)
		sort.Strings(p)
		findings = append(findings, singleOwnerViolation{Action: action, Owners: carriers, Paths: p})
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Action < findings[j].Action })
	return findings
}

// emitSingleOwnerViolations reads the same stdin shape as --live-pairs and
// --unregistered-actions.
//
// Refuses to print over an empty agent set for the reason on
// emitUnregisteredActions, and ALSO over an empty declaration set: with nothing
// declared SingleOwner, every possible export produces zero findings, so a
// clean report would be indistinguishable from a broken build in which the
// actions package failed to register. Zero findings against a non-empty
// declaration set is a real, meaningful pass.
func emitSingleOwnerViolations() {
	declared := datahelpers.ListSingleOwnerActions()
	if len(declared) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --single-owner-actions: 0 actions declare SingleOwner — "+
				"refusing to print a clean report that no export could ever fail.\n")
		os.Exit(2)
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --single-owner-actions: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--single-owner-actions")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --single-owner-actions: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --single-owner-actions: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	findings := findSingleOwnerViolations(agents, declared)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --single-owner-actions: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr,
		"config-key-audit --single-owner-actions: %d agents decoded (%d undecodable), %d declared single-owner action(s), %d finding(s)\n",
		len(agents), failed, len(declared), len(findings))
	if len(findings) > 0 {
		os.Exit(1)
	}
}
