//go:build ignore

// prove_fence_can_fail_tool_loan_vs_savings.go — the S2 mutation discipline,
// applied to fence_tool_loan_vs_savings.json.
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
// No fairness deviations: loancalculator.co.uk serves every referenced asset 200
// (probed 2026-08-05), so all non-page requests are redirected to the live origin,
// exactly the teaser file's original rule.
//
// USAGE
//
//	go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail_tool_loan_vs_savings.go \
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

// Every `from` below was verified against the served page on 2026-08-05: each
// occurs exactly once (grep counts recorded in NOTES).
var mutants = []mutant{
	{
		name:        "the page stops being served (404)",
		serveStatus: http.StatusNotFound,
		expectFail:  []string{"page-serves-200"},
	},
	{
		// This also nulls the script's own root lookup (`querySelector('.tool-loan-
		// vs-savings-section')` bails), so most behavioural checks go red as
		// collateral — reported, not asserted.
		name:       "the tool section loses the class acceptance resolves it by",
		from:       `<section class="tool-loan-vs-savings-section">`,
		to:         `<section class="tool-loan-vs-savings-section-renamed">`,
		expectFail: []string{"tool-renders"},
	},
	{
		// bugs_open/157's class: present in the DOM, rendered too small to use.
		// 10x10 with overflow:hidden rather than display:none so the inputs keep
		// non-empty bounding boxes and interaction steps stay actionable.
		name:       "the tool collapses below usable size (10x10, overflow hidden)",
		from:       "</body>",
		to:         `<style>.tool-loan-vs-savings-section{width:10px!important;height:10px!important;overflow:hidden!important;padding:0!important}</style></body>`,
		expectFail: []string{"tool-has-visible-area"},
	},
	{
		name:       "the heading stops posing the tool's question",
		from:       `<h1>Pay Off Loan or Save the Cash?</h1>`,
		to:         `<h1>Loan Tool</h1>`,
		expectFail: []string{"heading-poses-the-question"},
	},
	{
		// The loan side's arithmetic drifts by 10%: £75.00 reads £67.50. Loan still
		// wins the default vector (67.50 > 50), so the verdict checks stay green —
		// this isolates the on-load golden.
		name:       "the loan-side arithmetic drifts (x0.9)",
		from:       `var loanSave = amount * loanRate;`,
		to:         `var loanSave = amount * loanRate * 0.9;`,
		expectFail: []string{"computes-default-vector-on-load"},
	},
	{
		name:       "no panel is ever highlighted as the winner",
		from:       `(loanWins ? loanPanel : savePanel).classList.add('winner');`,
		to:         `(loanWins ? loanPanel : savePanel).classList.add('winner-disabled');`,
		expectFail: []string{"default-winner-highlighted"},
	},
	{
		// The copy bank key is renamed, so copy('winner-badge') returns "" and the
		// badge is textless — the accessible verdict silently vanishes.
		name:       "the winner badge copy goes missing from the copy bank",
		from:       `<span data-copy="winner-badge">Better option</span>`,
		to:         `<span data-copy="winner-badge-renamed">Better option</span>`,
		expectFail: []string{"default-winner-badge-reads"},
	},
	{
		// The exact defect the component's own copy-bank comment warns about:
		// BOTH panels claim "Better option" at once.
		name:       "the losing badge carries the winner text too",
		from:       `saveBadge.innerText = loanWins ? '' : badge;`,
		to:         `saveBadge.innerText = badge;`,
		expectFail: []string{"losing-badge-stays-empty"},
	},
	{
		// Typing no longer recomputes (the input listener is unbound; change still
		// fires, so the select-driven tax check keeps passing — the fill sets the
		// value silently and the later select recomputes from it).
		name:       "live recompute on typing is unbound",
		from:       `el.addEventListener('input', compare);`,
		to:         `el.addEventListener('input-disabled', compare);`,
		expectFail: []string{"recomputes-live-when-savings-win"},
	},
	{
		name:       "the loan panel always wins regardless of the numbers",
		from:       `var loanWins = loanSave > saveEarn;`,
		to:         `var loanWins = true;`,
		expectFail: []string{"savings-win-flips-verdict"},
	},
	{
		// The tax fraction is ignored: after selecting 40%, net savings still read
		// £80.00 instead of £48.00. Every other vector has tax=0 where (1-tax)=1,
		// so this isolates the tax check exactly.
		name:       "the taxpayer's marginal rate is ignored",
		from:       `var netSaveRate = saveRate * (1 - tax);`,
		to:         `var netSaveRate = saveRate;`,
		expectFail: []string{"tax-select-cuts-net-savings"},
	},
	{
		// The recorded tie rule (strict >, else-branch takes the tie -> savings)
		// is flipped to >=. Only an exact tie can see this.
		name:       "the tie is handed to the loan panel",
		from:       `var loanWins = loanSave > saveEarn;`,
		to:         `var loanWins = loanSave >= saveEarn;`,
		expectFail: []string{"tie-goes-to-savings"},
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
		fmt.Fprintln(os.Stderr, "usage: prove_fence_can_fail_tool_loan_vs_savings.go <criteria.json> <live-url>")
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

	fmt.Printf("=== proving the fence can fail (tool-loan-vs-savings) ===\n")
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
			RunID:        "prove-fence-lvs",
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
