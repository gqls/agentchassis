// FILE: cmd/config-key-audit/removed_keys_cron_parity_test.go
//
// bugs_open/234 / RFC_021 Q1. The CronJob that makes the removed-config-keys
// check automatic (deployments/kustomize/services/removed-config-keys-check)
// carries a Python mirror of this binary's --removed-keys-in-use mode, for
// cron_parity_test.go's reason (no Go toolchain in the job container). Same two
// drift risks, pinned the same way:
//
//  1. check.py's DECLARED_REMOVED is a literal map. Out of step with
//     datahelpers.ListRemovedConfigKeys() it either stops asking about a
//     declared key (reports clean about a question it never asked — while the
//     runtime validator kills any carrier) or fires on a key the platform does
//     not reject.
//  2. check.py's walk_steps() is a copy of the mirrored traversal. Fixtures
//     below exercise the nested-substeps case and the substeps-beats-
//     sub_workflow case, where naive walks go blind.
//
// If python3 is unavailable the parity test SKIPS rather than passes quietly.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const removedCheckPyPath = "../../deployments/kustomize/services/removed-config-keys-check/base/check.py"

// TestRemovedKeysCronDeclaredListMatchesTheRegistry pins drift risk (1).
// Compares the action→key-set SHAPE only — the replacement messages live in the
// Go declaration and the register, and a third synced copy of the prose would
// itself be a drift surface.
func TestRemovedKeysCronDeclaredListMatchesTheRegistry(t *testing.T) {
	raw, err := os.ReadFile(removedCheckPyPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v — if the CronJob moved, move this test with it, "+
			"do not delete it: the literal it guards goes stale silently", removedCheckPyPath, err)
	}

	block := regexp.MustCompile(`(?s)DECLARED_REMOVED\s*=\s*\{(.*?)\n\}`).FindSubmatch(raw)
	if block == nil {
		t.Fatal("DECLARED_REMOVED map not found in check.py — the CronJob's declared set " +
			"is unreadable, so nothing can confirm it matches the registry")
	}

	fromPython := map[string][]string{}
	entryRe := regexp.MustCompile(`"([^"]+)"\s*:\s*\[([^\]]*)\]`)
	keyRe := regexp.MustCompile(`"([^"]+)"`)
	for _, entry := range entryRe.FindAllSubmatch(block[1], -1) {
		action := string(entry[1])
		var keys []string
		for _, k := range keyRe.FindAllSubmatch(entry[2], -1) {
			keys = append(keys, string(k[1]))
		}
		sort.Strings(keys)
		fromPython[action] = keys
	}

	registry := datahelpers.ListRemovedConfigKeys()
	if len(registry) == 0 {
		t.Fatal("registry declares no RemovedConfigKeys — both the Go detector and the " +
			"CronJob would report every fleet clean")
	}
	fromRegistry := map[string][]string{}
	for action, removed := range registry {
		keys := make([]string, 0, len(removed))
		for k := range removed {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fromRegistry[action] = keys
	}

	if !reflect.DeepEqual(fromPython, fromRegistry) {
		t.Errorf("check.py DECLARED_REMOVED = %v, registry = %v.\n"+
			"The CronJob is the only thing that asks this question automatically, and the "+
			"runtime validator hard-fails any carrier it misses.", fromPython, fromRegistry)
	}
}

// TestRemovedKeysCronAgreesWithTheGoDetector pins drift risk (2).
func TestRemovedKeysCronAgreesWithTheGoDetector(t *testing.T) {
	python, err := exec.LookPath("python3")
	if err != nil {
		t.Skip("python3 not available — skipping rather than passing, because a quiet pass " +
			"here is exactly the failure this test exists to catch")
	}
	if _, statErr := os.Stat(removedCheckPyPath); statErr != nil {
		t.Fatalf("check.py not found at %s: %v", removedCheckPyPath, statErr)
	}

	cases := []struct {
		name  string
		input string
	}{
		{
			name: "clean fleet — the post-migration-364 state, and the negative control",
			input: `[
				{"type": "improvement-loop", "workflow": {"steps": {
					"insert_rerender_item": {"action": "create_work_item", "config": {
						"site_id": "site_record.site_id",
						"spec_literal": {"refresh_site_components": true}}}
				}}}
			]`,
		},
		{
			name: "top-level carrier — the pre-364 state this check exists to catch",
			input: `[
				{"type": "improvement-loop", "workflow": {"steps": {
					"insert_rerender_item": {"action": "create_work_item", "config": {
						"site_id": "site_record.site_id",
						"spec": {"refresh_site_components": true}}}
				}}}
			]`,
		},
		{
			name: "nested — a carrier inside a loop's substeps, invisible to a top-level walk",
			input: `[
				{"type": "loop-agent", "workflow": {"steps": {
					"per_site": {"action": "loop", "config": {"substeps": {
						"file_item": {"action": "create_work_item", "config": {
							"site_id": "x", "spec": {"k": true}}}
					}}}
				}}}
			]`,
		},
		{
			name: "both shapes on one step — the runtime takes substeps and ignores sub_workflow",
			input: `[
				{"type": "loop-agent", "workflow": {"steps": {
					"per_site": {"action": "loop", "config": {
						"substeps": {"real": {"action": "create_work_item", "config": {"spec": {}}}},
						"sub_workflow": {"steps": {"ignored": {"action": "create_work_item", "config": {"spec": {}}}}}
					}}
				}}}
			]`,
		},
		{
			name: "the key on a NON-declaring action is not a finding, in both",
			input: `[
				{"type": "other-agent", "workflow": {"steps": {
					"some_step": {"action": "some_other_action", "config": {"spec": {"k": 1}}}
				}}}
			]`,
		},
	}

	declared := datahelpers.ListRemovedConfigKeys()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			agents, _, decErr := decodeLiveAgents([]byte(tc.input), "removed-parity")
			if decErr != nil {
				t.Fatalf("decodeLiveAgents: %v", decErr)
			}
			goFindings := findRemovedKeyCarriers(agents, declared)

			cmd := exec.Command(python, filepath.Clean(removedCheckPyPath), "--stdin")
			cmd.Stdin = bytes.NewBufferString(tc.input)
			var out, errBuf bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &errBuf
			// Exit 1 means "findings", not a crash — only a non-(0,1) code is a failure.
			if runErr := cmd.Run(); runErr != nil {
				if ee, ok := runErr.(*exec.ExitError); !ok || ee.ExitCode() > 1 {
					t.Fatalf("check.py failed: %v\nstderr: %s", runErr, errBuf.String())
				}
			}

			var pyFindings []removedKeyCarrier
			if jsonErr := json.Unmarshal(out.Bytes(), &pyFindings); jsonErr != nil {
				t.Fatalf("check.py output is not the findings shape: %v\ngot: %s", jsonErr, out.String())
			}
			if pyFindings == nil {
				pyFindings = []removedKeyCarrier{}
			}

			if !reflect.DeepEqual(goFindings, pyFindings) {
				t.Errorf("the CronJob and the Go detector disagree.\n  go:     %+v\n  python: %+v\n"+
					"Two traversals that disagree is the good case — you can see it (bugs_open/144).",
					goFindings, pyFindings)
			}
		})
	}
}
