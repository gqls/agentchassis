//go:build ignore

// prove_fence_can_fail_tool_llm_cost_calculator.go — the S2 mutation discipline,
// applied to fence_tool_llm_cost_calculator.json.
//
// SIBLING, NOT REUSE — same reason prove_fence_can_fail_tool_loan_vs_savings.go gives:
// prove_fence_can_fail.go's mutant list is hardcoded to a different tool, and a
// "generic" prover that reads as generic encodes its first subject. This file copies
// the proven architecture (baseline-first control, one string-replace per mutant,
// page-serves-200 must survive every body mutant, coverage table) and swaps in this
// tool's own mutants. All of this tool's behaviour is INLINE in the page (one
// <script> block, no external tool JS), so no asset-override machinery is needed
// beyond the plain 302-to-origin redirect.
//
// FORK AWARENESS NOTE (why the fence itself avoids exact-cent goldens tied to one
// formatter): this subject has FOUR forks besides canonical (fundamentallyai.com,
// webdesign.co.uk, finetuning.uk, leopardessconsulting.co.uk) sharing this PLAN via
// `function='tool-llm-cost-calculator'`. Read finetuning.uk's live template and it
// uses `fmtMoney()` (`toLocaleString` with no `minimumFractionDigits`, drops a
// trailing-zero cent) where canonical/webdesign/fundamentallyai/leopardess use
// `formatCurrencyFull()` (always 2dp). The fence's `computes-known-monthly-costs`
// check therefore asserts figures captured with a nonzero hundredths digit AND a
// safety margin (>=0.05 of a cent) from any x.xx5 rounding boundary — verified
// empirically against the live canonical AND webdesign pages (NOT hand-derived only:
// a hand-derived $45.67 for one figure at the first, tighter margin turned out wrong
// against the live page, which rendered $45.68 — `toLocaleString`'s rounding runs on
// the shortest round-tripping decimal of the double, not the exact binary value, so
// Python's own rounding of the same float is not a safe stand-in for JS's). This
// prover mutates the CANONICAL page only, same as every other sibling in this
// directory; fork parity for the fence overall was established separately (see NOTES,
// staged_component_build workstream, 2026-08-12) by running try_fence.go against both
// currently-resolvable placements (canonical and the webdesign.co.uk fork).
//
// USAGE
//
//	go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail_tool_llm_cost_calculator.go \
//	    <criteria.json> <live-url>
//
// EXIT: 0 only when the baseline is all-green AND every mutant was caught by the checks
// named for it AND every check in the fence has at least one mutant naming it.
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

	// serveStatus, when non-zero, makes the throwaway server answer the page path with
	// that HTTP status instead of the page body. The one mutant for page_status_ok.
	serveStatus int

	// expectFail names the checks whose PASS must be contingent on the mutated
	// behaviour. Other checks may also go red (collateral); only these are asserted.
	expectFail []string
}

// Every `from` below was verified against the served canonical page on 2026-08-12:
// each occurs exactly once (grep -o -F counts recorded in NOTES).
var mutants = []mutant{
	{
		name:        "the page stops being served (404)",
		serveStatus: http.StatusNotFound,
		expectFail:  []string{"page-serves-200"},
	},
	{
		name:       "the tool loses the class acceptance resolves it by",
		from:       `class="tool-container"`,
		to:         `class="tool-container-renamed"`,
		expectFail: []string{"tool-renders"},
	},
	{
		// 10x10 with overflow:hidden rather than display:none so the element keeps a
		// non-empty box for the interaction steps that come later in fence order.
		name:       "the tool collapses below usable size (10x10, overflow hidden)",
		from:       "</body>",
		to:         `<style>.tool-container{width:10px!important;height:10px!important;overflow:hidden!important;padding:0!important}</style></body>`,
		expectFail: []string{"tool-has-visible-area"},
	},
	{
		name:       "the calculate button loses the id acceptance clicks",
		from:       `id="calc-btn"`,
		to:         `id="calc-btn-renamed"`,
		expectFail: []string{"calc-button-present"},
	},
	{
		name:       "the results table loses the id the fence reads",
		from:       `id="results-tbody"`,
		to:         `id="results-tbody-renamed"`,
		expectFail: []string{"results-table-present"},
	},
	{
		name:       "the self-host panel loses the id the fence reads",
		from:       `id="self-host-params"`,
		to:         `id="self-host-params-renamed"`,
		expectFail: []string{"self-host-panel-present"},
	},
	{
		// The per-provider monthly-cost formula drifts by x0.5 (e.g. a dropped
		// per-1M-token divisor). The default-checked providers' costs all halve, so
		// the captured golden figures (5.48/45.68/91.35/120.96-shaped) no longer
		// appear anywhere in the table text.
		name:       "the monthly-cost arithmetic drifts (x0.5)",
		from:       `const monthlyCost = (inputTokensPerMonth / 1e6) * provider.inputPer1M`,
		to:         `const monthlyCost = 0.5 * (inputTokensPerMonth / 1e6) * provider.inputPer1M`,
		expectFail: []string{"computes-known-monthly-costs"},
	},
	{
		name:       "the annual toggle button stops marking itself active",
		from:       `document.getElementById('btn-annual').classList.toggle('active', period === 'annual');`,
		to:         `document.getElementById('btn-annual').classList.toggle('active-renamed', period === 'annual');`,
		expectFail: []string{"annual-toggle-switches-active-state"},
	},
	{
		// The self-host checkbox no longer reveals its own parameter fields, so the
		// later `fill` steps on #gpu-count/#gpu-cost-month/#ops-overhead cannot find
		// an actionable (visible) element and the interaction step itself fails.
		name:       "the self-host checkbox stops revealing its parameter fields",
		from:       `document.getElementById('self-host-params').classList.toggle('visible', e.target.checked);`,
		to:         `document.getElementById('self-host-params').classList.toggle('visible-renamed', e.target.checked);`,
		expectFail: []string{"self-host-toggle-produces-breakeven-rows"},
	},
	{
		// Only the inputTokens<=0 arm of the validation OR-chain is disabled; the
		// other fields are left at the valid values the prior check in fence order
		// set them to, so this isolates the one arm.
		name:       "zero input tokens no longer counts as invalid",
		from:       `inputs.inputTokens <= 0`,
		to:         `inputs.inputTokens <= -1`,
		expectFail: []string{"zero-tokens-shows-error"},
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
		fmt.Fprintln(os.Stderr, "usage: prove_fence_can_fail_tool_llm_cost_calculator.go <criteria.json> <live-url>")
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

	fmt.Printf("=== proving the fence can fail (tool-llm-cost-calculator) ===\n")
	fmt.Printf("fence: %s (%d checks, desktop only)\n", critPath, len(checks))
	fmt.Printf("page:  %s (%d bytes)\n", liveURL, len(original))
	fmt.Printf("assets other than the page are 302'd to %s\n", origin)

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

	run := func(body []byte, status int) (failed map[string]string, skipped []string, err error) {
		served.Lock()
		served.body, served.status = body, status
		served.Unlock()

		res, err := action.Execute(context.Background(), browserrunner.RunChecksRequest{
			RunID:        "prove-fence-llm-cost-calc",
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
		fmt.Println("  The control is broken, so no mutant result below it would mean anything.")
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
