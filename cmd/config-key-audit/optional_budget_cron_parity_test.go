// FILE: cmd/config-key-audit/optional_budget_cron_parity_test.go
//
// RFC 022. The CronJob that makes the optional-key budget automatic
// (deployments/kustomize/services/optional-key-budget-check) carries a Python
// re-implementation — the owner's explicit choice ("we can keep the python",
// 2026-08-14) — and that buys FOUR drift risks, each pinned here in the same
// style as cron_parity_test.go (the RFC 006 sibling):
//
//  1. OPTIONAL_KEY_COUNTS is a literal; if it falls behind the registry the
//     job silently stops watching a newly-grown action.
//  2. ACKED_LEVELS is a literal mirroring optional_key_budget_acks.json; if
//     they drift, the daily job and the repo disagree about what was reviewed.
//  3. BUDGET is a literal that must equal the wrapper script's default, or the
//     daily job and a hand run disagree about N.
//  4. walk_steps() is a THIRD hand-written traversal (bugs_open/144's shape);
//     it is compared against the Go detector on fixtures that include the
//     substeps-wins nesting.
//
// If python3 is unavailable the traversal test SKIPS rather than passes
// quietly — a silent pass here would be the failure mode it prevents.
package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"testing"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

const budgetCheckPyPath = "../../deployments/kustomize/services/optional-key-budget-check/base/check.py"
const budgetAcksPath = "../../docs/agent_docs/docs024_key_docs_latest/architecture_review/optional_key_budget_acks.json"

// pyDictInts extracts a `NAME = { "k": 1, ... }` literal from check.py.
func pyDictInts(t *testing.T, raw []byte, name string) map[string]int {
	t.Helper()
	block := regexp.MustCompile(`(?s)` + name + `\s*=\s*\{(.*?)\}`).FindSubmatch(raw)
	if block == nil {
		t.Fatalf("%s literal not found in check.py", name)
	}
	out := map[string]int{}
	for _, m := range regexp.MustCompile(`"([^"]+)"\s*:\s*(\d+)`).FindAllSubmatch(block[1], -1) {
		n, _ := strconv.Atoi(string(m[2]))
		out[string(m[1])] = n
	}
	return out
}

// Drift risk (1): the counts literal against the live registry.
func TestBudgetCronCountsLiteralMatchesTheRegistry(t *testing.T) {
	raw, err := os.ReadFile(budgetCheckPyPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v — if the CronJob moved, move this test with it, "+
			"do not delete it: the literal it guards goes stale silently", budgetCheckPyPath, err)
	}
	got := pyDictInts(t, raw, "OPTIONAL_KEY_COUNTS")

	want := map[string]int{}
	for _, name := range datahelpers.ListActionInputSpecNames() {
		if spec, ok := datahelpers.GetActionInputSpec(name); ok && len(spec.Optional) > 0 {
			want[name] = len(spec.Optional)
		}
	}
	if len(want) == 0 {
		t.Fatal("registry returned no optional-key declarations — the actions import has come unlinked")
	}

	for name, n := range want {
		if got[name] != n {
			t.Errorf("check.py OPTIONAL_KEY_COUNTS[%q] = %d, registry declares %d — "+
				"regenerate the literal (command in check.py's comment)", name, got[name], n)
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			t.Errorf("check.py OPTIONAL_KEY_COUNTS carries %q which the registry no longer "+
				"declares optional keys for — remove it", name)
		}
	}
}

// Drift risk (2): the acks literal against the repo's source of truth.
func TestBudgetCronAckedLevelsMatchTheAcksFile(t *testing.T) {
	raw, err := os.ReadFile(budgetCheckPyPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v", budgetCheckPyPath, err)
	}
	got := pyDictInts(t, raw, "ACKED_LEVELS")

	want, err := loadAckedLevels(budgetAcksPath)
	if err != nil {
		t.Fatalf("cannot read acks file %s: %v — the file is the source of truth the "+
			"cron mirrors; if it moved, move this test's path with it", budgetAcksPath, err)
	}
	// Compare both directions so neither side can carry a baseline the other lacks.
	for name, n := range want {
		if got[name] != n {
			t.Errorf("acks file has %q at %d, check.py ACKED_LEVELS has %d — update the mirror", name, n, got[name])
		}
	}
	for name, n := range got {
		if want[name] != n {
			t.Errorf("check.py ACKED_LEVELS has %q at %d, acks file has %d — the file is "+
				"the source of truth; never edit the literal alone", name, n, want[name])
		}
	}
}

// Drift risk (3): the daily job's N against the hand-run default.
func TestBudgetCronBudgetMatchesTheWrapperDefault(t *testing.T) {
	py, err := os.ReadFile(budgetCheckPyPath)
	if err != nil {
		t.Fatalf("cannot read %s: %v", budgetCheckPyPath, err)
	}
	pyBudget := regexp.MustCompile(`(?m)^BUDGET\s*=\s*(\d+)`).FindSubmatch(py)
	if pyBudget == nil {
		t.Fatal("BUDGET literal not found in check.py")
	}

	sh, err := os.ReadFile("../../scripts/audit-optional-key-budget.sh")
	if err != nil {
		t.Fatalf("cannot read the wrapper script: %v", err)
	}
	shBudget := regexp.MustCompile(`(?m)^\s*BUDGET=(\d+)`).FindSubmatch(sh)
	if shBudget == nil {
		t.Fatal("ruled default (BUDGET=<n>) not found in audit-optional-key-budget.sh")
	}

	if string(pyBudget[1]) != string(shBudget[1]) {
		t.Errorf("check.py BUDGET = %s but the wrapper defaults to %s — the daily job and "+
			"a hand run would disagree about the ruled N; change both together",
			pyBudget[1], shBudget[1])
	}
}

// budgetParityFleet exercises both nesting shapes the runtime honours: a
// sub_workflow carrier AND a substeps carrier (substeps WINS when both are
// present — the half a naive walk misses, bugs_open/144). append_doc_note's
// registry count (11 as of the ruling) exceeds the budget, so with three
// distinct carriers it must be the one finding from BOTH implementations.
const budgetParityFleet = `[
	{"type": "landmine-verifier", "workflow": {"start_step": "load_entry", "steps": {
		"persist_verdict": {"action": "append_doc_note"}
	}}},
	{"type": "council-gate", "workflow": {"start_step": "loop_seats", "steps": {
		"loop_seats": {"action": "loop_over_items", "config": {"sub_workflow": {"steps": {
			"persist": {"action": "append_doc_note"}
		}}}}
	}}},
	{"type": "tool-improver", "workflow": {"start_step": "loop_fixes", "steps": {
		"loop_fixes": {"action": "loop_over_items", "config": {
			"substeps": {"note": {"action": "append_doc_note"}},
			"sub_workflow": {"steps": {"decoy": {"action": "append_doc_note"}}}
		}}
	}}}
]`

// Drift risk (4): the third traversal against the Go detector, same fixtures.
func TestBudgetCronAgreesWithTheGoDetector(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 unavailable — parity cannot be checked here; do not read this as a pass")
	}

	agents, failed, err := decodeLiveAgents([]byte(budgetParityFleet), "parity")
	if err != nil || failed != 0 {
		t.Fatalf("fixture decode: err=%v failed=%d", err, failed)
	}
	var goFindings []map[string]interface{}
	for _, r := range censusOptionalKeys(agents, 10, nil) {
		if !r.OverBudget {
			continue
		}
		goFindings = append(goFindings, map[string]interface{}{
			"action": r.Action, "optional_keys": float64(r.OptionalKeys),
			"consumers": float64(r.Consumers),
		})
	}
	if len(goFindings) != 1 || goFindings[0]["action"] != "append_doc_note" ||
		goFindings[0]["consumers"] != float64(3) {
		t.Fatalf("fixture premise broken: Go detector should find append_doc_note with 3 "+
			"consumers (the substeps carrier included), got %v", goFindings)
	}

	cmd := exec.Command("python3", budgetCheckPyPath, "--stdin")
	cmd.Stdin = bytes.NewReader([]byte(budgetParityFleet))
	out, _ := cmd.Output() // exit 1 on findings is expected; the output is the comparison
	var pyFindings []map[string]interface{}
	if err := json.Unmarshal(out, &pyFindings); err != nil {
		t.Fatalf("check.py --stdin output is not JSON: %v\n%s", err, out)
	}

	if len(pyFindings) != len(goFindings) {
		t.Fatalf("python found %d finding(s), Go found %d — the traversals disagree:\npy: %v\ngo: %v",
			len(pyFindings), len(goFindings), pyFindings, goFindings)
	}
	for i := range goFindings {
		for _, k := range []string{"action", "optional_keys", "consumers"} {
			if pyFindings[i][k] != goFindings[i][k] {
				t.Errorf("finding %d field %q: python %v, Go %v", i, k, pyFindings[i][k], goFindings[i][k])
			}
		}
	}
}
