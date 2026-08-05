//go:build ignore

// prove_fence_can_fail_tool_fuel_cost_estimator.go — the S2 mutation discipline,
// applied to fence_tool_fuel_cost_estimator.json.
//
// SIBLING, NOT REUSE. prove_fence_can_fail.go's mutant list is hardcoded to
// tool-review-council-simulator and prove_fence_can_fail_teaser_reveal_panel.go's to that
// component — this lane has twice measured that a prover which reads as generic encodes its
// first subject. This file copies the proven architecture (baseline-first control, one
// string-replace per mutant, page-serves-200 must survive every body mutant, coverage
// table) and swaps in this tool's own mutants. All of this tool's behaviour is INLINE in
// the page (one <script> block, no external tool JS), so the teaser file's per-mutant
// asset-override machinery is not needed and was dropped.
//
// ONE DECLARED DEVIATION FROM THE TEASER FILE'S FAIRNESS RULE. The live origin 404s
// /assets/images/logo.png on every gaswholesalers.com page (asset row exists in `assets`,
// file absent at the served path — site chrome, not this tool; see NOTES 2026-08-05).
// Under the original rule ("redirect every non-page asset to the live origin") the
// BASELINE would therefore fail no-console-errors for a chrome reason and abort, proving
// nothing about the fence. So this harness serves a local 1x1 PNG at that ONE path in
// EVERY run, baseline and mutants alike — modelling the page as it stands once the chrome
// is repaired. This is declared here rather than silently done because a control that
// edits its own inputs is exactly the class of quiet self-deception this lane documents;
// the deviation is uniform across all runs, so no mutant's verdict depends on it, and the
// no-console-errors check is still proven able to fail by its own dedicated mutant.
//
// USAGE
//
//	go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail_tool_fuel_cost_estimator.go \
//	    <criteria.json> <live-url>
//
// EXIT: 0 only when the baseline is all-green AND every mutant was caught by the checks
// named for it AND every check in the fence has at least one mutant naming it.
package main

import (
	"context"
	"encoding/base64"
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

	// serveStatus, when non-zero, makes the throwaway server answer the page path with
	// that HTTP status instead of the page body. The one mutant for page_status_ok.
	serveStatus int

	// expectFail names the checks whose PASS must be contingent on the mutated
	// behaviour. Other checks may also go red (collateral); only these are asserted.
	expectFail []string
}

// Every `from` below was verified against the served page on 2026-08-05:
// each occurs exactly once, EXCEPT the empty-state paragraph which occurs twice
// (static markup first, then the JS template string that restores it). Replace n=1
// edits the FIRST — the static markup — which is precisely the default state the
// check asserts, so the two-occurrence case is correct here, not a hazard.
var mutants = []mutant{
	{
		name:        "the page stops being served (404)",
		serveStatus: http.StatusNotFound,
		expectFail:  []string{"page-serves-200"},
	},
	{
		name:       "the tool container loses the class acceptance resolves it by",
		from:       `<div class="tool-container">`,
		to:         `<div class="tool-container-renamed">`,
		expectFail: []string{"tool-renders"},
	},
	{
		// bugs_open/157's class: present in the DOM, rendered too small to use.
		// 10x10 with overflow:hidden rather than display:none so the inputs keep
		// non-empty bounding boxes and the interaction checks fail fast as
		// collateral instead of each waiting out an actionability timeout.
		name:       "the tool collapses below usable size (10x10, overflow hidden)",
		from:       "</body>",
		to:         `<style>.tool-container{width:10px!important;height:10px!important;overflow:hidden!important;padding:0!important}</style></body>`,
		expectFail: []string{"tool-has-visible-area"},
	},
	{
		name:       "the heading stops naming the tool",
		from:       `<h2>Fuel Cost Estimator</h2>`,
		to:         `<h2>Fuel Price Helper</h2>`,
		expectFail: []string{"heading-names-the-tool"},
	},
	{
		name:       "the pre-input guidance is replaced with something unhelpful",
		from:       `<p class="empty-state">Fill in the fields above to see your potential savings.</p>`,
		to:         `<p class="empty-state-renamed">Loading...</p>`,
		expectFail: []string{"empty-state-before-input"},
	},
	{
		name:       "the results region stops announcing updates to assistive tech",
		from:       `aria-live="polite"`,
		to:         `aria-live-removed="polite"`,
		expectFail: []string{"results-are-live-region"},
	},
	{
		// The arithmetic moves: 52 weeks -> 50. Annual reads $100,000.00 instead of
		// the captured $104,000.00; the weekly and per-gallon values are untouched,
		// so this proves the computed_values check and only it.
		name:       "the annualisation constant drifts (52 -> 50 weeks)",
		from:       `period === 'weekly' ? 52 : 12`,
		to:         `period === 'weekly' ? 50 : 12`,
		expectFail: []string{"computes-known-savings"},
	},
	{
		// Same line, other arm: 12 months -> 10. Only the monthly-toggle assertion
		// ($24,000.00) can see this; the weekly-path golden is untouched.
		name:       "the monthly annualisation drifts (12 -> 10 months)",
		from:       `period === 'weekly' ? 52 : 12`,
		to:         `period === 'weekly' ? 52 : 10`,
		expectFail: []string{"monthly-toggle-recomputes"},
	},
	{
		name:       "switching to litres no longer relabels the unit suffix",
		from:       `const unitShort = unit === 'gallons' ? 'gal' : 'L';`,
		to:         `const unitShort = 'gal';`,
		expectFail: []string{"unit-toggle-relabels"},
	},
	{
		name:       "the wholesale-above-retail guard is disabled",
		from:       `if (savingPerUnit <= 0) {`,
		to:         `if (savingPerUnit < -999999) {`,
		expectFail: []string{"rejects-wholesale-above-retail"},
	},
	{
		name:       "invalid input no longer marks the field",
		from:       `if (group) group.classList.add('input-error');`,
		to:         `if (group) group.classList.add('input-error-disabled');`,
		expectFail: []string{"invalid-input-shows-error"},
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

// A valid 1x1 transparent PNG, served at /assets/images/logo.png in every run —
// see the header comment for why the baseline needs it.
var logoPNG, _ = base64.StdEncoding.DecodeString(
	"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==")

type verdict struct {
	name   string
	caught bool
	note   string
}

var served struct {
	sync.RWMutex
	body   []byte
	status int // 0 = serve the body with 200
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: prove_fence_can_fail_tool_fuel_cost_estimator.go <criteria.json> <live-url>")
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

	fmt.Printf("=== proving the fence can fail (tool-fuel-cost-estimator) ===\n")
	fmt.Printf("fence: %s (%d checks, desktop only)\n", critPath, len(checks))
	fmt.Printf("page:  %s (%d bytes)\n", liveURL, len(original))
	fmt.Printf("assets are 302'd to %s, EXCEPT /assets/images/logo.png which is served\n", origin)
	fmt.Printf("locally in every run (the origin 404s it — site chrome defect, see header)\n\n")

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
		if r.URL.Path == "/assets/images/logo.png" {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(logoPNG)
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

	run := func(body []byte, status int) (failed map[string]string, skipped []string, err error) {
		served.Lock()
		served.body, served.status = body, status
		served.Unlock()

		res, err := action.Execute(context.Background(), browserrunner.RunChecksRequest{
			RunID:        "prove-fence-fce",
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

	fmt.Println("--- baseline: the unmutated local copy (logo served locally, see header) ---")
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
		os.Exit(1)
	}
	fmt.Printf("  green: all %d checks pass on the local copy.\n\n", len(checks))

	var verdicts []verdict
	allCaught := true

	for i, m := range mutants {
		fmt.Printf("--- mutant %d/%d: %s ---\n", i+1, len(mutants), m.name)

		mutatedBody := original

		switch {
		case m.serveStatus != 0:
			if m.serveStatus == http.StatusOK {
				fmt.Println("  FAIL: a 200 'status mutant' changes nothing.")
				verdicts = append(verdicts, verdict{m.name, false, "status mutant is a no-op"})
				allCaught = false
				continue
			}
			fmt.Printf("  response changed: HTTP 200 -> HTTP %d (body untouched)\n", m.serveStatus)

		default:
			if !strings.Contains(string(original), m.from) {
				fmt.Printf("  FAIL: target string not present — this mutant is stale: %q\n", m.from)
				verdicts = append(verdicts, verdict{m.name, false, "target string absent (stale mutant)"})
				allCaught = false
				continue
			}
			mutatedBody = []byte(strings.Replace(string(original), m.from, m.to, 1))
			if string(mutatedBody) == string(original) {
				fmt.Println("  FAIL: replacement was a no-op — the artefact did not change.")
				verdicts = append(verdicts, verdict{m.name, false, "no-op replacement"})
				allCaught = false
				continue
			}
			fmt.Printf("  artefact changed: %d -> %d bytes\n", len(original), len(mutatedBody))
		}

		failed, skipped, err := run(mutatedBody, m.serveStatus)
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
		fmt.Println("no mutant at all.")
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
