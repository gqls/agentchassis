// FILE: cmd/config-key-audit/migration_lint_predicate_parity_test.go
//
// bugs_closed/314, last residual. `scripts/pattern-check.py`'s migration idempotency
// lint (bugs_open/007 Class C) carries a COPY of the migration runner's appliable-name
// pattern and a DELIBERATELY DIFFERENT exclusion. Same shape as the three siblings in
// this package — optional_budget_cron_parity_test.go, cron_parity_test.go,
// retention_parity_test.go — which each extract a literal from a non-Go file and assert
// it against its source of truth.
//
// WHY A PARITY TEST AND NOT A "HAS THE RUNNER CHANGED?" WATCHER. The obvious guard is
// the one scripts/council-scope.sh uses for its own copy: grep the runner for its line
// and warn if it moved. That guard could not have caught this defect, and the reason is
// worth stating because it decides the shape of every future guard here. The lint's
// predicate NEVER MATCHED the runner: the runner gained [A-Za-z] on 2026-07-20
// (a51333fd7) and the lint was written on 2026-07-25 (9d95e1c31) already lowercase-only.
// It was WRONG AT BIRTH, not drifted. A watcher keyed on the runner changing would have
// sat green for six weeks over a lint that silently skipped every migration with a
// capital in its name. So this compares THE TWO LITERALS, and pins THE DECISIONS.
//
// THE THREE DRIFT RISKS IT PINS:
//
//  1. the runner's name pattern moves and the Python copy does not, or the reverse;
//  2. the lint's exclusion grows to name a suffix the runner would actually APPLY —
//     every excluded suffix must be a runner sidecar, or the lint exempts a file that
//     is genuinely replayable, which is the 007 hazard with the alarm switched off;
//  3. a session "reconciles" the lint to the runner's SIDECAR_RE and silently drops
//     _HOLD. That is bugs_closed/314's own defect one level down — the council's
//     editquality seat caught exactly it inside 314's fix — and ONLY a
//     must-lint/must-not-lint fixture table catches it. A literal-only check passes
//     happily while the RULE is wrong.
//
// WHY _HOLD MUST STAY LINTED (test 3's load-bearing row). A _HOLD.sql is a migration
// held back from the runner FOR ORDERING and applied BY HAND — out of band by
// definition. run-migrations.sh:245-250 REFUSES `--record-only` on any SIDECAR_RE
// match, so it CANNOT be ledger-recorded while it carries the suffix; the house
// sequence is forced: hand-apply, RENAME to drop the suffix, then record. Between the
// rename and the record the runner sees a pending, unrecorded, appliable file and
// REPLAYS it. [MEASURED 2026-09-02] that promotion is routine, not theoretical: 26
// distinct files renamed _HOLD -> plain between 2026-08-01 and 09-02, all 26 present on
// disk today under the plain name (e.g. 6e8fa6a3c, df0d718dd, 465679270). So _HOLD is
// the one category GUARANTEED to be applied out of band before the ledger can know.
//
// If python3 is unavailable the function-level test SKIPS rather than passing quietly —
// a silent pass here would be the exact failure mode this file exists to prevent.
package main

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

const (
	runnerPath       = "../../scripts/migration/run-migrations.sh"
	patternCheckPath = "../../scripts/pattern-check.py"
)

// All anchored with (?m)^ so a COMMENT quoting the pattern cannot satisfy the match.
// (LANDMINES: "a source-scan test makes your COMMENTS load-bearing — first occurrence
// wins, and the needle matching your OWN comment passes VACUOUSLY".)
var (
	pyNameRe = regexp.MustCompile(`(?m)^MIGRATION_NAME_RE = re\.compile\(r"([^"]+)"\)`)
	pyExclRe = regexp.MustCompile(`(?m)^MIGRATION_NEVER_REPLAYED_RE = re\.compile\(r"([^"]+)"\)`)
	// The runner's appliable-name test. ANCHORED STRUCTURALLY, on the fact that the
	// appliable rule is the one immediately PAIRED WITH A SIDECAR_RE FILTER:
	//     | grep -E '<name pattern>' \
	//     | grep -vE "$SIDECAR_RE" \        (:283-284, the scan)
	//     | grep -E  "$SIDECAR_RE" \        (:293-294, the sidecar report)
	// A looser "any grep -E starting ^[0-9]" matched a THIRD site — :312's near-miss
	// warner, `^[0-9]{3}.*\.sql$`, which is deliberately WIDER and is not this rule.
	// The first cut of this test did exactly that and failed on its first run; the
	// pairing is what distinguishes the rule from the two greps that merely resemble it.
	shNameRe    = regexp.MustCompile(`(?m)^\s*\| grep -E '([^']+)' \\\n\s*\| grep -v?E "\$SIDECAR_RE"`)
	shSidecarRe = regexp.MustCompile(`(?m)^SIDECAR_RE='([^']+)'`)
)

func readFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	return string(raw)
}

func extract(t *testing.T, re *regexp.Regexp, body, what string) string {
	t.Helper()
	m := re.FindStringSubmatch(body)
	if m == nil {
		t.Fatalf("%s: literal not found. If it was reformatted, update this test and "+
			"re-read the header before changing the RULE (bugs_closed/314).", what)
	}
	return m[1]
}

// Risk (1): the copy is still the runner's, byte for byte. The Python literal is
// written [0-9]{3} rather than \d{3} precisely so this can be string equality
// rather than a translation between regex dialects.
func TestMigrationLintNamePatternIsTheRunnersVerbatim(t *testing.T) {
	runner := readFile(t, runnerPath)
	py := extract(t, pyNameRe, readFile(t, patternCheckPath), "pattern-check.py MIGRATION_NAME_RE")

	all := shNameRe.FindAllStringSubmatch(runner, -1)
	// Two sites are expected: the scan (:283) and the sidecar report (:293). A count
	// that is not 2 means a site moved, was added, or was removed — any of which
	// changes what "the runner's rule" means and must be looked at, not assumed.
	if len(all) != 2 {
		t.Fatalf("run-migrations.sh: expected 2 appliable-name greps paired with SIDECAR_RE, found %d — "+
			"reconcile this test with the runner before trusting either", len(all))
	}
	for _, m := range all {
		if m[1] != py {
			t.Errorf("migration-name pattern DRIFTED.\n  run-migrations.sh: %s\n  pattern-check.py : %s\n"+
				"The idempotency lint silently stops covering whatever the two disagree about.", m[1], py)
		}
	}
}

// Risk (2): every suffix the lint excludes must be one the runner would refuse to
// apply. If the lint ever excludes something --apply WOULD run, it has switched off
// the alarm for a genuinely replayable file.
func TestMigrationLintExclusionsAreAllRunnerSidecars(t *testing.T) {
	sidecar := regexp.MustCompile(extract(t, shSidecarRe, readFile(t, runnerPath), "run-migrations.sh SIDECAR_RE"))
	excl := regexp.MustCompile(extract(t, pyExclRe, readFile(t, patternCheckPath), "pattern-check.py MIGRATION_NEVER_REPLAYED_RE"))

	for _, name := range []string{
		"600_x_ROLLBACK.sql", "600_x_VERIFY.sql", "600_x_SUPERSEDED.sql",
		"600_x_HOLD_ROLLBACK.sql", "600_x_ROLLBACK_SUPERSEDED.sql",
	} {
		if !excl.MatchString(name) {
			continue // not excluded by the lint; nothing to prove
		}
		if !sidecar.MatchString(name) {
			t.Errorf("%s is EXCLUDED by the lint but the runner would APPLY it — "+
				"the lint has exempted a replayable file (bugs_open/007 Class C)", name)
		}
	}
}

// Risk (3): the DECISIONS, which is the only test that can catch a deliberate
// divergence being un-deliberated. Evaluated with the literals as extracted, so it
// tests what ships rather than a restatement of it.
func TestMigrationLintPredicateFixtures(t *testing.T) {
	name := regexp.MustCompile(extract(t, pyNameRe, readFile(t, patternCheckPath), "MIGRATION_NAME_RE"))
	excl := regexp.MustCompile(extract(t, pyExclRe, readFile(t, patternCheckPath), "MIGRATION_NEVER_REPLAYED_RE"))
	lintable := func(n string) bool { return name.MatchString(n) && !excl.MatchString(n) }

	for _, f := range migrationLintFixtures {
		if got := lintable(f.name); got != f.lint {
			t.Errorf("%s: lintable=%v, want %v — %s", f.name, got, f.lint, f.why)
		}
	}
}

var migrationLintFixtures = []struct {
	name string
	lint bool
	why  string
}{
	{"482_ROLLBACK_claim_timeout_exclusion.sql", true,
		"the motivating case: ROLLBACK MID-name is not a suffix, and the runner applies this file"},
	{"582_dispatch_sibling_A_task_name_on_trigger_row.sql", true, "capital mid-name, runner-appliable"},
	{"584_dispatch_sibling_C_insert_trigger_2.sql", true, "capital mid-name, INSERTs into scheduled_tasks"},
	{"600_plain_lowercase.sql", true, "the ordinary case — must not regress"},
	{"600_some_config_HOLD.sql", true,
		"THE deliberate divergence from the runner: hand-applied, unrecordable, then renamed into the appliable set"},
	{"668_terms_publish_RELOCK.sql", true,
		"an UNCLASSIFIED suffix defaults to linted — over-inclusion costs one advisory line, under-inclusion costs a 23505"},
	{"600_some_config_ROLLBACK.sql", false, "the undo, hand-run against an already-decided state"},
	{"600_some_config_VERIFY.sql", false, "assertions only, never writes"},
	{"600_some_config_SUPERSEDED.sql", false, "retired"},
	{"600_some_config_HOLD_ROLLBACK.sql", false, "stacked: a HOLD's rollback is still the undo"},
	{"676_build_standard_optins_HOLD_SUPERSEDED.sql", false, "stacked, real file"},
	{"600b_odd.sql", false, "fails the runner's name rule"},
	{"600_odd-name.sql", false, "hyphen — fails the runner's name rule"},
	{"README.md", false, "not a migration at all"},
}

// The three tests above read the literals. This one runs the REAL function, so a
// change to migration_is_lintable() that does not touch the constants is still caught.
func TestMigrationLintPythonFunctionAgreesWithTheFixtures(t *testing.T) {
	if _, err := exec.LookPath("python3"); err != nil {
		t.Skip("python3 unavailable — parity NOT checked here; do not read this as a pass")
	}
	const prog = `import importlib.util,sys
s=importlib.util.spec_from_file_location("pc",sys.argv[1])
m=importlib.util.module_from_spec(s); s.loader.exec_module(m)
print("\n".join(str(m.migration_is_lintable(n)) for n in sys.argv[2:]))`

	args := []string{"-c", prog, patternCheckPath}
	for _, f := range migrationLintFixtures {
		args = append(args, f.name)
	}
	out, err := exec.Command("python3", args...).Output()
	if err != nil {
		t.Fatalf("could not evaluate migration_is_lintable(): %v", err)
	}
	got := strings.Fields(string(out))
	if len(got) != len(migrationLintFixtures) {
		t.Fatalf("expected %d verdicts, got %d", len(migrationLintFixtures), len(got))
	}
	for i, f := range migrationLintFixtures {
		if (got[i] == "True") != f.lint {
			t.Errorf("migration_is_lintable(%q) = %s, want %v — %s", f.name, got[i], f.lint, f.why)
		}
	}
}
