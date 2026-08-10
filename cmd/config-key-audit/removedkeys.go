// FILE: cmd/config-key-audit/removedkeys.go
//
// bugs_open/234 / RFC_021 Q1 (owner ruling 2026-08-10). Which live step carries
// a config key its action declares REMOVED?
//
// Every finding is an agent one roll away from refusing to run: the runtime
// validator (checkStepConfigKeys) hard-fails a workflow carrying a removed key
// on every message once a binary with the declaration rolls. The offline
// question therefore has to be answerable BEFORE a roll, against live
// agent_definitions — which is this mode, and the CronJob
// (deployments/kustomize/services/removed-config-keys-check) that runs its
// Python mirror daily. The parity between the two is pinned by
// removed_keys_cron_parity_test.go.
//
// In its own file for singleowner.go's reason: modes are separable questions
// over a shared decode, and a file per mode is what lets two concurrent
// sessions add one each without colliding.
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

// removedKeyCarrier is one step still carrying a key its action retired. Agent,
// path and key together make the finding actionable without a second query —
// and unlike singleOwnerViolation this reports per STEP, because every carrying
// step is independently fatal to its workflow, not a property of the set.
type removedKeyCarrier struct {
	Agent  string `json:"agent"`
	Path   string `json:"path"`
	Action string `json:"action"`
	Key    string `json:"key"`
}

// findRemovedKeyCarriers is the pure check. Same validation.WalkSteps traversal
// as every other mode, for the bugs_open/144 reason.
func findRemovedKeyCarriers(agents []liveAgent, declared map[string]map[string]string) []removedKeyCarrier {
	findings := []removedKeyCarrier{}
	for _, agent := range agents {
		validation.WalkSteps(agent.Workflow, func(path string, step models.Step, nested bool) {
			if step.Action == "" {
				return
			}
			removed, ok := declared[step.Action]
			if !ok || len(step.Config) == 0 {
				return
			}
			for key := range removed {
				if _, present := step.Config[key]; present {
					findings = append(findings, removedKeyCarrier{
						Agent:  agent.Type,
						Path:   path,
						Action: step.Action,
						Key:    key,
					})
				}
			}
		})
	}
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Agent != b.Agent {
			return a.Agent < b.Agent
		}
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		return a.Key < b.Key
	})
	return findings
}

// emitRemovedKeyCarriers reads the same stdin shape as the sibling modes.
// Refuses over an empty agent set AND an empty declaration set, for
// emitSingleOwnerViolations' reason: a clean report no export could ever fail
// is indistinguishable from a broken build.
func emitRemovedKeyCarriers() {
	declared := datahelpers.ListRemovedConfigKeys()
	if len(declared) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --removed-keys-in-use: 0 actions declare RemovedConfigKeys — "+
				"refusing to print a clean report that no export could ever fail.\n")
		os.Exit(2)
	}

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --removed-keys-in-use: reading stdin: %v\n", err)
		os.Exit(2)
	}

	agents, failed, err := decodeLiveAgents(raw, "--removed-keys-in-use")
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --removed-keys-in-use: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --removed-keys-in-use: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken export.\n", failed)
		os.Exit(2)
	}

	findings := findRemovedKeyCarriers(agents, declared)

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(findings); err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --removed-keys-in-use: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr,
		"config-key-audit --removed-keys-in-use: %d agents decoded (%d undecodable), %d declaring action(s), %d carrier(s)\n",
		len(agents), failed, len(declared), len(findings))
	if len(findings) > 0 {
		os.Exit(1)
	}
}
