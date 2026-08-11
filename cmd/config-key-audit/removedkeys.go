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
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

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

// removedKeysRunSummary is the doc_notes body: what was looked at, what was
// found, and — on a clean run — a sentence saying so in terms, because a row
// that only appears on failure cannot be told apart from a job that stopped
// running (writeDocNote's own comment; bugs_open/140 is what that costs).
func removedKeysRunSummary(scanned, undecoded int, declared map[string]map[string]string, findings []removedKeyCarrier) string {
	var keys []string
	for action, removed := range declared {
		for k := range removed {
			keys = append(keys, action+"."+k)
		}
	}
	sort.Strings(keys)

	var b strings.Builder
	fmt.Fprintf(&b, "REMOVED CONFIG KEYS IN USE CHECK (bugs_open/234, RFC_021 Q1 ruling)\n\n")
	fmt.Fprintf(&b, "live agent definitions walked: %d", scanned)
	if undecoded > 0 {
		fmt.Fprintf(&b, " (%d undecodable, counted not swallowed)", undecoded)
	}
	fmt.Fprintf(&b, "\nkeys declared removed:         %s\ncarriers found:                %d\n\n",
		strings.Join(keys, ", "), len(findings))

	if len(findings) == 0 {
		b.WriteString("No live definition carries a removed config key.\n\n" +
			"This row exists on a clean run ON PURPOSE: a MISSING row means the job did\n" +
			"not run, which is not the same as 'nothing is wrong', and the two must not\n" +
			"look alike.\n")
		return b.String()
	}

	b.WriteString("CARRIERS — each of these definitions is REJECTED at validation on every\n" +
		"message by any binary carrying the declaration (the agent stops working\n" +
		"until the definition is fixed):\n\n")
	for _, f := range findings {
		fmt.Fprintf(&b, "  %s.%s: %s carries removed key %q\n", f.Agent, f.Path, f.Action, f.Key)
	}
	b.WriteString("\nFix the DEFINITION (and its seed, or a reseed replays the key): the\n" +
		"replacement spelling is in the action's RemovedConfigKeys message and in\n" +
		"register SCR-007. Do NOT un-declare the key to silence this.\n")
	return b.String()
}

// emitRemovedKeyCarriers reads live agents from stdin, or — with --report —
// straight from Postgres, and in that mode records the run in doc_notes.
//
// THE --report MODE IS WHY THIS FILE EXISTS RATHER THAN A PYTHON MIRROR.
// The first version of the RFC_021 Q1 CronJob re-implemented the step traversal
// in Python because the sibling single-owner check does, and the council's
// `reuse_agent` seat gated on it (round 2, corr 3eb0d1f1): a THIRD hand-written
// walk of "every step at all depths" is precisely the shape bugs_open/144
// burned this platform with — two walks go blind in the same direction and then
// agree with each other. The seat was right, and the answer was already on the
// shelf: shared-output-fields-check builds THIS binary into a tiny image and
// runs it. So the CronJob now runs the same validation.WalkSteps the validator
// enforces with, the Python mirror is deleted, and the drift class it created —
// along with the parity test that policed it — no longer exists.
//
// Refuses over an empty agent set AND an empty declaration set, for
// emitSingleOwnerViolations' reason: a clean report no fleet could ever fail is
// indistinguishable from a broken build.
func emitRemovedKeyCarriers() {
	report := false
	for _, a := range os.Args[2:] {
		if a == "--report" {
			report = true
			continue
		}
		fmt.Fprintf(os.Stderr, "config-key-audit --removed-keys-in-use: unrecognised argument %q (want: [--report])\n", a)
		os.Exit(2)
	}

	declared := datahelpers.ListRemovedConfigKeys()
	if len(declared) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --removed-keys-in-use: 0 actions declare RemovedConfigKeys — "+
				"refusing to print a clean report that no fleet could ever fail.\n")
		os.Exit(2)
	}

	var (
		agents []liveAgent
		failed int
		err    error
	)
	if report {
		// Straight from Postgres: this image contains no kubectl and the service
		// account has no pods/exec RBAC (see fleetdb.go).
		var db *sql.DB
		db, err = dbConn()
		if err != nil {
			fmt.Fprintf(os.Stderr, "config-key-audit --removed-keys-in-use: %v\n", err)
			os.Exit(2)
		}
		defer db.Close()
		agents, failed, err = loadLiveAgentsFromDB(db, "--removed-keys-in-use")
	} else {
		var raw []byte
		raw, err = io.ReadAll(os.Stdin)
		if err == nil {
			agents, failed, err = decodeLiveAgents(raw, "--removed-keys-in-use")
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "config-key-audit --removed-keys-in-use: %v\n", err)
		os.Exit(2)
	}
	if len(agents) == 0 {
		fmt.Fprintf(os.Stderr,
			"config-key-audit --removed-keys-in-use: 0 agents decoded (%d undecodable) — "+
				"refusing to print a clean report over an empty or broken fleet.\n", failed)
		os.Exit(2)
	}

	findings := findRemovedKeyCarriers(agents, declared)

	if report {
		summary := removedKeysRunSummary(len(agents), failed, declared, findings)
		fmt.Print(summary)
		writeDocNote("removed-config-keys", summary, "removed-config-keys", "removed-config-keys-check")
		if len(findings) > 0 {
			os.Exit(1)
		}
		return
	}

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
