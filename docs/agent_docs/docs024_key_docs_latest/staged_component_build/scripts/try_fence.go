//go:build ignore

// try_fence.go — evaluate a candidate ```criteria fence against a live URL using
// the PLATFORM'S OWN evaluator, locally, BEFORE the fence is written into a PLAN.
//
// WHY THIS EXISTS. This lane's one surviving rule is "watch every branch of a
// check fail before quoting anything it says". A criteria fence is the one
// artefact where that was impossible to do cheaply: the only way to exercise one
// was to write it into `doc_plans` and dispatch an acceptance run through the
// cluster — so the first time anybody saw the fence run was AFTER it had been
// published as the tool's contract. Authoring a selector and then discovering it
// matched nothing is exactly how a run comes back "clean" having asserted
// nothing (`needs_criteria`, or a check that passes vacuously).
//
// It imports `internal/adapters/browserrunner` and calls RunChecksAction.Execute,
// so it is NOT a lookalike — it is the same switch, the same Playwright driver,
// the same settle delay and the same pass/fail semantics the fleet will use. A
// second implementation would be a third thing to keep in step, and this lane's
// own PLAN D5′ exists because a contract with two enforcement points drifted.
//
// USAGE
//
//	go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/try_fence.go \
//	    <criteria.json> <url> [more urls...]
//
// EXIT: 0 only when every check was EVALUATED and PASSED on every profile.
//
//	1 on any failure, any unimplemented-type skip, or any arithmetic
//	  mismatch. A profile-gated skip is reported and does NOT fail.
//
// THE PROPERTY THAT MATTERS: it can never print a verdict about a check it did
// not evaluate. Every check id in the fence must appear in exactly one of the
// three buckets below, and the reconciliation is asserted, not assumed — because
// the failure this lane keeps hitting is a report that reads healthy about
// something it never measured.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/adapters/browserrunner"
)

// minimal mirror of the fence's own shape — used ONLY to enumerate the ids the
// fence declares, so the reconciliation can be checked. The evaluation itself
// uses the real parser inside the adapter.
type fenceShape struct {
	Profiles []string `json:"profiles"`
	Checks   []struct {
		ID       string   `json:"id"`
		Type     string   `json:"type"`
		Profiles []string `json:"profiles"`
	} `json:"checks"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: try_fence.go <criteria.json> <url> [url...]")
		os.Exit(2)
	}
	path, urls := os.Args[1], os.Args[2:]

	raw, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", path, err)
		os.Exit(2)
	}

	// Parse-check FIRST and separately. run_checks returns a bare error for a
	// malformed fence, which upstream would surface as an infra failure rather
	// than as "your JSON is wrong" — the distinction a fence author needs.
	var shape fenceShape
	if err := json.Unmarshal(raw, &shape); err != nil {
		fmt.Printf("FAIL: the fence does not parse as JSON: %v\n", err)
		os.Exit(1)
	}
	if len(shape.Checks) == 0 {
		fmt.Println("FAIL: the fence declares zero checks — an empty fence is the silent class this")
		fmt.Println("      lane exists to catch: the run starts, asserts nothing, and reads clean.")
		os.Exit(1)
	}

	profiles := shape.Profiles
	if len(profiles) == 0 {
		profiles = []string{"desktop"} // resolveProfiles' default, mirrored for the arithmetic
	}

	fmt.Printf("=== fence trial: %s ===\n", path)
	fmt.Printf("checks declared: %d   profiles: %s   urls: %d\n\n",
		len(shape.Checks), strings.Join(profiles, ","), len(urls))

	logger, _ := zap.NewDevelopment()
	action := browserrunner.NewRunChecksAction(logger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel)), nil)

	res, err := action.Execute(context.Background(), browserrunner.RunChecksRequest{
		RunID:        "try-fence-local",
		URLs:         urls,
		Profiles:     profiles,
		CriteriaJSON: string(raw),
	})
	if err != nil {
		// Infra, or an unparseable fence the adapter rejected.
		fmt.Printf("FAIL: the evaluator refused the run: %v\n", err)
		os.Exit(1)
	}

	// ── results, grouped so a failure is readable ───────────────────────────
	type row struct{ profile, url, id, detail string }
	var passed, failed, gated, unimpl []row

	for _, r := range res.Results {
		rw := row{r.Profile, r.URL, r.CheckID, r.Detail}
		if r.Pass {
			passed = append(passed, rw)
		} else {
			failed = append(failed, rw)
		}
	}
	for _, s := range res.Skipped {
		rw := row{s.Profile, s.URL, s.CheckID, s.Detail}
		// A profile-gated skip is the fence author's own instruction and is fine.
		// "<type> not implemented" is NOT: it means the running binary has no
		// evaluator for that type, and an all-skipped set reads as PASS upstream.
		if strings.Contains(s.Detail, "not run on profile") {
			gated = append(gated, rw)
		} else {
			unimpl = append(unimpl, rw)
		}
	}

	show := func(title string, rows []row) {
		if len(rows) == 0 {
			return
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].profile != rows[j].profile {
				return rows[i].profile < rows[j].profile
			}
			return rows[i].id < rows[j].id
		})
		fmt.Printf("--- %s (%d) ---\n", title, len(rows))
		for _, r := range rows {
			fmt.Printf("  [%-7s] %-34s %s\n", r.profile, r.id, r.detail)
		}
		fmt.Println()
	}

	show("FAILED", failed)
	show("SKIPPED — TYPE NOT IMPLEMENTED IN THIS BINARY", unimpl)
	show("skipped — profile-gated by the fence itself", gated)
	show("passed", passed)

	// ── the reconciliation, asserted ────────────────────────────────────────
	// Every declared check must land in exactly one bucket per (profile, url).
	// If this does not add up, the report above is INCOMPLETE and no verdict it
	// prints can be trusted — which is the whole failure class, so it is fatal
	// rather than a warning.
	want := len(shape.Checks) * len(profiles) * len(urls)
	got := len(passed) + len(failed) + len(gated) + len(unimpl)
	fmt.Printf("arithmetic: %d passed + %d failed + %d gated + %d unimplemented = %d; expected %d checks x %d profiles x %d urls = %d\n",
		len(passed), len(failed), len(gated), len(unimpl), got,
		len(shape.Checks), len(profiles), len(urls), want)

	bad := false
	if got != want {
		fmt.Printf("FAIL: %d check-evaluations are unaccounted for — this report is incomplete.\n", want-got)
		bad = true
	}
	if len(unimpl) > 0 {
		fmt.Println("FAIL: at least one check type is not implemented in the binary this harness")
		fmt.Println("      built against. Upstream, an unknown type is SKIPPED, not failed, and an")
		fmt.Println("      all-skipped set reads as PASS plus a 7-day cooldown. Do not ship this fence.")
		bad = true
	}
	if len(failed) > 0 {
		bad = true
	}
	if bad {
		fmt.Println("\nRESULT: FAIL")
		os.Exit(1)
	}
	fmt.Printf("\nRESULT: PASS — %d check-evaluations, all evaluated, all passed.\n", len(passed))
}
