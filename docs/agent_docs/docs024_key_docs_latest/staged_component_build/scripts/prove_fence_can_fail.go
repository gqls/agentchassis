//go:build ignore

// prove_fence_can_fail.go — the S2 DISCIPLINE, applied to a criteria fence.
//
// D8 cut the mutation HARNESS and kept the mutation RULE: "every check proven
// able to fail", plus the sub-rule "a mutant counts only if the artefact
// provably changed". This is that rule discharged for one fence, in one file,
// without the forty minutes of scaffolding the forward-run lane spent for
// nothing.
//
// HOW IT WORKS, and why it is a fair test:
//
//  1. It serves a LOCAL COPY of the real page from a throwaway HTTP server, and
//     302-redirects every other request (CSS, snippets.js, images) to the live
//     origin. Without that redirect the copy 404s its own stylesheet, which
//     Chromium logs as a console error — the no-console-errors baseline would go
//     red for a reason that has nothing to do with the mutation.
//  2. THE BASELINE RUNS FIRST AND MUST BE ALL-GREEN. If the unmutated local copy
//     does not pass exactly as the live URL does, the control is broken and no
//     mutant result below it means anything, so the program aborts rather than
//     reporting.
//  3. Each mutant is ONE string replacement whose target was verified to occur
//     exactly once. The replacement is asserted to have changed the bytes — a
//     no-op replacement would otherwise "prove" a check can fail while changing
//     nothing, which is the conservation-proof failure already in WRONG_CALLS.
//  4. `page-serves-200` must still pass under every mutant. That is the
//     targeted-not-demolished control: a mutant that merely broke the page would
//     turn everything red and appear to validate the whole fence at once.
//
// MUTANTS RUN ON DESKTOP ONLY, deliberately, and this is stated rather than
// quietly done: profile coverage is established by try_fence.go against the live
// URL (36/36 across desktop and mobile). What a mutant establishes is that a
// check's PASS is contingent on the behaviour it names, and one profile settles
// that. Running both would double a ~4-minute job to prove nothing further.
//
// USAGE
//
//	go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail.go \
//	    <criteria.json> <live-url>
//
// EXIT: 0 only when the baseline is all-green AND every mutant was caught by the
//
//	checks named for it. 1 otherwise.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"

	"github.com/gqls/agentchassis/internal/adapters/browserrunner"
)

type mutant struct {
	name string
	from string
	to   string
	// serveStatus, when non-zero, makes the throwaway server answer the tool
	// path with that HTTP status instead of the page. Used for the one check
	// (page_status_ok) that no string edit inside the page can ever falsify.
	serveStatus int
	// expectFail names the checks whose PASS must be contingent on the mutated
	// behaviour. Other checks may also go red (a broken init breaks most of the
	// tool); only these are asserted, because only these are the claim.
	expectFail []string
}

// Every `from` below was verified to occur EXACTLY ONCE in the served page
// before this file was written. A target that matches twice would mutate
// something the mutant does not name.
var mutants = []mutant{
	{
		name: "the script never initialises (the teaser-reveal-panel class: JS that never ran client-side)",
		from: "function init() {",
		to:   "function init() { return;",
		expectFail: []string{
			"roster-is-built-client-side",
			"headline-is-computed-not-a-placeholder",
			"cleared-panel-refuses-to-invent-a-number",
			"default-preset-is-our-typical-eight",
		},
	},
	// CORRECTED after the first run of this file REFUTED the check it names.
	// The original mutant killed only the `input` listener and the check still
	// PASSED, because the tool also binds `change` and Playwright's fill()
	// dispatches BOTH. So the check never asserted the PLAN's "live-updating on
	// the `input` event" claim — it asserted only "the lever is wired to
	// something". Two things changed rather than one: the mutant now kills both
	// listeners (the real dead-slider defect), and the check was RENAMED from
	// `threshold-lever-updates-live` to `threshold-lever-updates-the-readout` so
	// its name no longer claims more than it proves. The residual gap is
	// recorded in the PLAN: the criteria vocabulary has no way to dispatch one
	// DOM event without the other, so `input`-vs-`change` wiring is not
	// distinguishable by any fence, and saying so is the honest position.
	{
		name: "dead slider: the lever is wired to nothing (BOTH listeners killed — see the note above)",
		from: "input.addEventListener('input', update);\n        input.addEventListener('change', update);",
		to:   "input.addEventListener('input', function () {});\n        input.addEventListener('change', function () {});",
		expectFail: []string{"threshold-lever-updates-the-readout"},
	},
	{
		name:       "the within-rounds figure reverts to its placeholder",
		from:       "el.passN.textContent = pct(r.pPassN * 100);",
		to:         "el.passN.textContent = '--';",
		expectFail: []string{"within-rounds-is-computed"},
	},
	{
		name:       "the mean-rounds figure reverts to its placeholder",
		from:       "? r.meanRounds.toFixed(1)",
		to:         "? '--'",
		expectFail: []string{"mean-rounds-is-computed"},
	},
	{
		name:       "the seats-firing figure reverts to its placeholder",
		from:       "el.firing.textContent = r.firing.toFixed(1);",
		to:         "el.firing.textContent = '--';",
		expectFail: []string{"seats-firing-is-computed"},
	},
	{
		// Not a string edit: the server answers the tool path with a status
		// instead of the page. Without this, page_status_ok would be the one
		// check in the fence that had never been seen red — and a status check
		// nobody has watched fail is exactly as worthless as any other.
		name:        "the page stops being served (404)",
		serveStatus: http.StatusNotFound,
		expectFail:  []string{"page-serves-200"},
	},
	{
		name:       "an empty roster invents a 100% pass rate instead of refusing",
		from:       "el.pass1.textContent = 'n/a';",
		to:         "el.pass1.textContent = '100.0%';",
		expectFail: []string{"cleared-panel-refuses-to-invent-a-number"},
	},
	{
		name:       "the page opens on the wrong preset",
		from:       "applyPreset('typical');",
		to:         "applyPreset('all');",
		expectFail: []string{"default-preset-is-our-typical-eight"},
	},
	{
		name: "the preset buttons stop responding",
		from: "btn.addEventListener('click', function () {",
		to:   "btn.addEventListener('nope', function () {",
		expectFail: []string{
			"every-seat-preset-counts-twenty-six",
			"minimal-preset-is-the-measured-pair",
			"cleared-panel-refuses-to-invent-a-number",
		},
	},
	{
		name:       "the tool container loses the identity acceptance resolves it by",
		from:       `data-component="tool-review-council-simulator"`,
		to:         `data-component="tool-something-else-entirely"`,
		expectFail: []string{"tool-container-renders"},
	},
	{
		name:       "the denominator stops saying what it counts",
		from:       "count ROUNDS",
		to:         "count things",
		expectFail: []string{"denominator-states-what-it-counts"},
	},
	{
		name:       "the ranked blocker chart renders no rows",
		from:       `<div class="rcs-bar-row">`,
		to:         `<div class="rcs-bar-row-suppressed">`,
		expectFail: []string{"blocker-chart-is-ranked"},
	},
	{
		name:       "the reality band renders no legend",
		from:       `'<li><span class="rcs-legend-pct">'`,
		to:         `'<li><span class="rcs-legend-pct-suppressed">'`,
		expectFail: []string{"reality-band-is-drawn"},
	},
	{
		name:       "the threshold slider starts somewhere other than where we run it",
		from:       `id="rcs-threshold" min="0" max="2" step="1" value="2"`,
		to:         `id="rcs-threshold" min="0" max="2" step="1" value="1"`,
		expectFail: []string{"threshold-starts-where-we-run-it"},
	},
	{
		name:       "the computed headline reverts to its placeholder",
		from:       "el.pass1.textContent = pct(r.pPass1 * 100);",
		to:         "el.pass1.textContent = '--';",
		expectFail: []string{"headline-is-computed-not-a-placeholder"},
	},
	{
		name:       "a wide element pushes the page past the viewport",
		from:       "</body>",
		to:         `<div style="width:3000px;height:8px"></div></body>`,
		expectFail: []string{"no-horizontal-overflow"},
	},
	{
		name:       "a script error reaches the console",
		from:       "</body>",
		to:         `<script>window.__deliberatelyUndefined.boom = 1;</script></body>`,
		expectFail: []string{"no-console-errors"},
	},
}

// verdict is one mutant's outcome.
type verdict struct {
	name   string
	caught bool
	note   string
}

// served holds the bytes the throwaway server currently returns for the tool path.
var served struct {
	sync.RWMutex
	body   []byte
	status int // 0 = serve the body with 200
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: prove_fence_can_fail.go <criteria.json> <live-url>")
		os.Exit(2)
	}
	critPath, liveURL := os.Args[1], os.Args[2]

	u, err := url.Parse(liveURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bad url: %v\n", err)
		os.Exit(2)
	}
	origin := u.Scheme + "://" + u.Host
	toolPath := u.Path

	// ── the criteria, forced to desktop only (stated in the header, not hidden) ─
	rawCrit, err := os.ReadFile(critPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", critPath, err)
		os.Exit(2)
	}
	var critMap map[string]interface{}
	if err := json.Unmarshal(rawCrit, &critMap); err != nil {
		fmt.Fprintf(os.Stderr, "criteria does not parse: %v\n", err)
		os.Exit(2)
	}
	critMap["profiles"] = []string{"desktop"}
	desktopCrit, _ := json.Marshal(critMap)
	checks, _ := critMap["checks"].([]interface{})

	// ── the original page, straight from the live origin ───────────────────────
	resp, err := http.Get(liveURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot fetch %s: %v\n", liveURL, err)
		os.Exit(2)
	}
	original, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil || resp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "fetch %s: HTTP %d, err %v\n", liveURL, resp.StatusCode, err)
		os.Exit(2)
	}
	fmt.Printf("=== proving the fence can fail ===\n")
	fmt.Printf("fence: %s (%d checks, desktop only — see the file header for why)\n", critPath, len(checks))
	fmt.Printf("page:  %s (%d bytes)\n", liveURL, len(original))
	fmt.Printf("assets other than the page itself are 302'd to %s so the local copy is a fair control\n\n", origin)

	// ── throwaway server: the tool path from memory, everything else redirected ─
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(2)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == toolPath {
			served.RLock()
			b, st := served.body, served.status
			served.RUnlock()
			if st != 0 {
				http.Error(w, http.StatusText(st), st)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(b)
			return
		}
		http.Redirect(w, r, origin+r.URL.Path, http.StatusFound)
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer func() { _ = srv.Close() }()
	localURL := "http://" + ln.Addr().String() + toolPath

	logger, _ := zap.NewDevelopment()
	quiet := logger.WithOptions(zap.IncreaseLevel(zap.ErrorLevel))
	action := browserrunner.NewRunChecksAction(quiet, nil)

	// run serves `body` and returns the set of check ids that FAILED, plus the
	// ids that were skipped (which must always be empty here).
	run := func(body []byte, status int) (failed map[string]string, skipped []string, err error) {
		served.Lock()
		served.body, served.status = body, status
		served.Unlock()
		res, err := action.Execute(context.Background(), browserrunner.RunChecksRequest{
			RunID:        "prove-fence",
			URLs:         []string{localURL},
			Profiles:     []string{"desktop"},
			CriteriaJSON: string(desktopCrit),
		})
		if err != nil {
			return nil, nil, err
		}
		failed = map[string]string{}
		for _, r := range res.Results {
			if !r.Pass {
				failed[r.CheckID] = r.Detail
			}
		}
		for _, s := range res.Skipped {
			skipped = append(skipped, s.CheckID+" ("+s.Detail+")")
		}
		return failed, skipped, nil
	}

	// ── 1. THE BASELINE. Nothing below it means anything if this is not green ──
	fmt.Println("--- baseline: the unmutated local copy ---")
	baseFail, baseSkip, err := run(original, 0)
	if err != nil {
		fmt.Printf("ABORT: the baseline run itself errored: %v\n", err)
		os.Exit(1)
	}
	if len(baseSkip) > 0 {
		fmt.Printf("ABORT: the baseline skipped %d check(s): %s\n", len(baseSkip), strings.Join(baseSkip, ", "))
		os.Exit(1)
	}
	if len(baseFail) > 0 {
		fmt.Printf("ABORT: the baseline is NOT green — %d check(s) fail on the unmutated copy:\n", len(baseFail))
		for id, d := range baseFail {
			fmt.Printf("    %-40s %s\n", id, d)
		}
		fmt.Println("  The control is broken, so no mutant result below would mean anything.")
		fmt.Println("  Usual cause: an asset the local copy cannot reach. Fix the control first.")
		os.Exit(1)
	}
	fmt.Printf("  green: all %d checks pass on the local copy, matching the live URL.\n\n", len(checks))

	// ── 2. THE MUTANTS ─────────────────────────────────────────────────────────
	var verdicts []verdict
	allCaught := true

	for i, m := range mutants {
		fmt.Printf("--- mutant %d/%d: %s ---\n", i+1, len(mutants), m.name)

		// The artefact must PROVABLY change, or the "proof" proves nothing.
		// For a status mutant the artefact is the RESPONSE, not the body, so the
		// change is asserted on the status rather than skipped.
		mutated := original
		if m.serveStatus != 0 {
			if m.serveStatus == http.StatusOK {
				fmt.Println("  FAIL: a 200 'status mutant' changes nothing — the response is unaltered.")
				verdicts = append(verdicts, verdict{m.name, false, "status mutant is a no-op"})
				allCaught = false
				continue
			}
			fmt.Printf("  response changed: HTTP 200 -> HTTP %d (body untouched)\n", m.serveStatus)
		} else {
			if !strings.Contains(string(original), m.from) {
				fmt.Printf("  FAIL: target string not present — this mutant is stale: %q\n", m.from)
				verdicts = append(verdicts, verdict{m.name, false, "target string absent (stale mutant)"})
				allCaught = false
				continue
			}
			mutated = []byte(strings.Replace(string(original), m.from, m.to, 1))
			if string(mutated) == string(original) {
				fmt.Println("  FAIL: replacement was a no-op — the artefact did not change.")
				verdicts = append(verdicts, verdict{m.name, false, "no-op replacement"})
				allCaught = false
				continue
			}
			fmt.Printf("  artefact changed: %d -> %d bytes\n", len(original), len(mutated))
		}

		failed, skipped, err := run(mutated, m.serveStatus)
		if err != nil {
			fmt.Printf("  FAIL: run errored: %v\n", err)
			verdicts = append(verdicts, verdict{m.name, false, "run errored"})
			allCaught = false
			continue
		}
		if len(skipped) > 0 {
			fmt.Printf("  FAIL: %d check(s) were SKIPPED: %s\n", len(skipped), strings.Join(skipped, ", "))
			verdicts = append(verdicts, verdict{m.name, false, "checks skipped"})
			allCaught = false
			continue
		}
		// Targeted, not demolished: the page must still serve — unless breaking
		// exactly that is the mutant's whole point.
		if d, ok := failed["page-serves-200"]; ok && !namesCheck(m.expectFail, "page-serves-200") {
			fmt.Printf("  FAIL: the mutant broke the page itself (page-serves-200: %s).\n", d)
			verdicts = append(verdicts, verdict{m.name, false, "mutant broke the page, not the behaviour"})
			allCaught = false
			continue
		}

		var missed []string
		for _, want := range m.expectFail {
			if d, ok := failed[want]; ok {
				fmt.Printf("  caught by %-42s %s\n", want, d)
			} else {
				missed = append(missed, want)
			}
		}
		// Report collateral so the mutant's blast radius is visible, not implied.
		var collateral []string
		for id := range failed {
			named := false
			for _, want := range m.expectFail {
				if id == want {
					named = true
				}
			}
			if !named {
				collateral = append(collateral, id)
			}
		}
		if len(collateral) > 0 {
			fmt.Printf("  also went red (not asserted, reported so the blast radius is visible): %s\n",
				strings.Join(collateral, ", "))
		}
		if len(missed) > 0 {
			fmt.Printf("  FAIL: NOT caught by %s — that check's pass does not depend on this behaviour.\n",
				strings.Join(missed, ", "))
			verdicts = append(verdicts, verdict{m.name, false, "missed by " + strings.Join(missed, ", ")})
			allCaught = false
		} else {
			verdicts = append(verdicts, verdict{m.name, true, fmt.Sprintf("%d check(s) went red as named", len(m.expectFail))})
		}
		fmt.Println()
	}

	// ── 3. WHICH CHECKS HAVE BEEN WATCHED TO FAIL, AND WHICH HAVE NOT ─────────
	// This table is the point of the file: a check with no mutant has never been
	// seen red, and this lane's rule is that such a check is not evidence.
	watched := map[string]bool{}
	for _, m := range mutants {
		for _, id := range m.expectFail {
			watched[id] = true
		}
	}
	fmt.Println("--- coverage: which checks have been WATCHED to fail ---")
	var unwatched []string
	for _, c := range checks {
		cm, _ := c.(map[string]interface{})
		id, _ := cm["id"].(string)
		if watched[id] {
			fmt.Printf("  watched red   %s\n", id)
		} else {
			fmt.Printf("  NEVER RED     %s\n", id)
			unwatched = append(unwatched, id)
		}
	}
	fmt.Println()

	fmt.Println("--- summary ---")
	for _, v := range verdicts {
		mark := "caught"
		if !v.caught {
			mark = "MISSED"
		}
		fmt.Printf("  %-6s  %s — %s\n", mark, v.name, v.note)
	}
	fmt.Printf("\nmutants: %d, caught: %d\n", len(mutants), countCaught(verdicts))
	fmt.Printf("checks watched red: %d of %d", len(checks)-len(unwatched), len(checks))
	if len(unwatched) > 0 {
		fmt.Printf("  — NEVER RED: %s", strings.Join(unwatched, ", "))
	}
	fmt.Println()

	if !allCaught {
		fmt.Println("\nRESULT: FAIL — at least one mutant was not caught by the check that names it.")
		os.Exit(1)
	}
	if len(unwatched) > 0 {
		fmt.Println("\nRESULT: FAIL — every mutant was caught, but the checks listed as NEVER RED have")
		fmt.Println("no mutant at all. A check nobody has seen go red is not evidence, so this exits 1")
		fmt.Println("rather than reporting a green with a hole in it.")
		os.Exit(1)
	}
	fmt.Println("\nRESULT: PASS — baseline green, every mutant caught, every check watched red.")
}

func namesCheck(ids []string, want string) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func countCaught(vs []verdict) int {
	n := 0
	for _, v := range vs {
		if v.caught {
			n++
		}
	}
	return n
}
