// FILE: cmd/config-key-audit/cron_parity_test.go
//
// RFC 006. The CronJob that makes the single-owner check automatic
// (deployments/kustomize/services/single-owner-carriers-check) cannot run this Go
// binary — it would need a git clone of a 262M repo plus a module download plus a
// compile, in a job with uncertain egress. So it carries a Python re-implementation,
// and that buys two drift risks which this file exists to make impossible to ship:
//
//  1. check.py's DECLARED_SINGLE_OWNER is a literal list, because the container has
//     no Go toolchain. If it falls behind the registry the CronJob silently stops
//     checking a newly-declared action — a gate that reports "clean" about a question
//     it never asked.
//
//  2. check.py's walk_steps() is a SECOND implementation of validation.WalkSteps.
//     bugs_open/144 is precisely this: two hand-written traversals go blind in the
//     same direction and then agree with each other, so agreement alone proves
//     nothing. What makes it evidence here is that BOTH are compared on fixtures
//     chosen to exercise the nesting the naive walk gets wrong.
//
// If python3 is unavailable the parity test SKIPS rather than passes quietly — a
// silent pass here would be the same failure mode it is written to prevent.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const checkPyPath = "../../deployments/kustomize/services/single-owner-carriers-check/base/check.py"

// TestCronCheckDeclaredListMatchesTheRegistry pins drift risk (1).
func TestCronCheckDeclaredListMatchesTheRegistry(t *testing.T) {
	raw, err := os.ReadFile(checkPyPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v — if the CronJob moved, move this test with it, "+
			"do not delete it: the literal it guards goes stale silently", checkPyPath, err)
	}

	block := regexp.MustCompile(`(?s)DECLARED_SINGLE_OWNER\s*=\s*\[(.*?)\]`).FindSubmatch(raw)
	if block == nil {
		t.Fatal("DECLARED_SINGLE_OWNER list not found in check.py — the CronJob's declared " +
			"set is unreadable, so nothing can confirm it matches the registry")
	}
	found := regexp.MustCompile(`"([^"]+)"`).FindAllSubmatch(block[1], -1)

	var fromPython []string
	for _, m := range found {
		fromPython = append(fromPython, string(m[1]))
	}

	fromRegistry := datahelpers.ListSingleOwnerActions()
	if len(fromRegistry) == 0 {
		t.Fatal("registry declares no single-owner actions — both the Go detector and the " +
			"CronJob would report every fleet clean")
	}

	if !reflect.DeepEqual(fromPython, fromRegistry) {
		t.Errorf("check.py DECLARED_SINGLE_OWNER = %v, registry = %v.\n"+
			"The CronJob is the only thing that runs this check automatically. Out of step, "+
			"it either misses a declared action (reports clean about a question it never asked) "+
			"or fires on one the platform does not consider wrong.", fromPython, fromRegistry)
	}
}

// TestCronCheckAgreesWithTheGoDetector pins drift risk (2). Feeds identical fixtures
// to both implementations and requires identical findings.
//
// The fixtures are not decorative: `nested` puts a carrier inside a loop's substeps
// (the case a top-level-only walk misses entirely), and `bothShapes` carries BOTH
// `substeps` and `sub_workflow` on one step — where the runtime takes substeps and
// ignores the other, so a traversal that reads sub_workflow instead would find a
// carrier that cannot execute.
func TestCronCheckAgreesWithTheGoDetector(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available — skipping rather than passing, because a quiet pass " +
			"here is exactly the failure this test exists to catch")
	}
	if _, statErr := os.Stat(checkPyPath); statErr != nil {
		t.Fatalf("check.py not found at %s: %v", checkPyPath, statErr)
	}

	cases := []struct {
		name  string
		input string
	}{
		{
			name:  "the three-carrier fleet that was live until 2026-08-02",
			input: threeCarrierFleet,
		},
		{
			name: "one owner — the post-migration fleet, and the negative control",
			input: `[
				{"type": "improvement-loop", "workflow": {"steps": {
					"triage_findings": {"action": "triage_detected_items", "next_step": "complete"}
				}}},
				{"type": "design-audit-agent", "workflow": {"steps": {
					"complete": {"action": "complete_workflow"}
				}}}
			]`,
		},
		{
			name: "nested — a carrier inside a loop's substeps, invisible to a top-level walk",
			input: `[
				{"type": "improvement-loop", "workflow": {"steps": {
					"triage_findings": {"action": "triage_detected_items", "next_step": "complete"}
				}}},
				{"type": "loop-agent", "workflow": {"steps": {
					"per_site": {"action": "loop", "config": {"substeps": {
						"triage": {"action": "triage_detected_items", "next_step": ""}
					}}}
				}}}
			]`,
		},
		{
			name: "both shapes on one step — the runtime takes substeps and ignores sub_workflow",
			input: `[
				{"type": "improvement-loop", "workflow": {"steps": {
					"triage_findings": {"action": "triage_detected_items", "next_step": "complete"}
				}}},
				{"type": "loop-agent", "workflow": {"steps": {
					"per_site": {"action": "loop", "config": {
						"substeps": {"real": {"action": "triage_detected_items"}},
						"sub_workflow": {"steps": {"ignored": {"action": "triage_detected_items"}}}
					}}
				}}}
			]`,
		},
		{
			name: "one agent carrying it twice is not an ownership violation, in both",
			input: `[
				{"type": "improvement-loop", "workflow": {"steps": {
					"triage_findings": {"action": "triage_detected_items", "next_step": "again"},
					"again":           {"action": "triage_detected_items", "next_step": "complete"}
				}}}
			]`,
		},
	}

	declared := datahelpers.ListSingleOwnerActions()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			agents, _, decErr := decodeLiveAgents([]byte(tc.input), "parity")
			if decErr != nil {
				t.Fatalf("decodeLiveAgents: %v", decErr)
			}
			goFindings := findSingleOwnerViolations(agents, declared)

			cmd := exec.Command(python, filepath.Clean(checkPyPath), "--stdin")
			cmd.Stdin = bytes.NewBufferString(tc.input)
			var out, errBuf bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errBuf
			// Exit 1 means "findings", not a crash — only a non-(0,1) code is a failure.
			if runErr := cmd.Run(); runErr != nil {
				if ee, ok := runErr.(*exec.ExitError); !ok || ee.ExitCode() > 1 {
					t.Fatalf("check.py failed: %v\nstderr: %s", runErr, errBuf.String())
				}
			}

			var pyFindings []singleOwnerViolation
			if jsonErr := json.Unmarshal(out.Bytes(), &pyFindings); jsonErr != nil {
				t.Fatalf("check.py output is not the findings shape: %v\ngot: %s", jsonErr, out.String())
			}
			if pyFindings == nil {
				pyFindings = []singleOwnerViolation{}
			}

			if !reflect.DeepEqual(goFindings, pyFindings) {
				t.Errorf("the CronJob and the Go detector disagree.\n  go:     %+v\n  python: %+v\n"+
					"Two traversals that disagree is the good case — you can see it. The one to fear "+
					"is two that agree because both are blind the same way (bugs_open/144).",
					goFindings, pyFindings)
			}
		})
	}
}
