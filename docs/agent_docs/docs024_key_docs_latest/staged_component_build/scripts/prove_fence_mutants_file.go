//go:build ignore

// prove_fence_mutants_file.go — the S2 mutation discipline with the mutant list
// supplied as a JSON FILE instead of a hardcoded Go slice.
//
// WHY THIS EXISTS. The lane's rule stands: a prover that READS as generic while
// hardcoding its first subject's strings is the recurring trap (twice measured).
// This file is the third form: the ARCHITECTURE (baseline-first control, one
// string-replace per mutant, page-serves-200 must survive every body mutant,
// coverage table, exit non-zero unless every check is watched red) is shared and
// subject-free; the SUBJECT lives entirely in a per-subject mutants.json that the
// author writes by hand after reading the served page — so nothing here can encode
// a first subject, and D10's exhaustive clearance stops minting a sibling .go file
// per component. Validated before first use by reproducing the call-to-action
// prover's exact 6/6 result from a JSON transcription of its mutants (2026-08-05).
//
// mutants.json shape (verify every `from` occurs exactly ONCE in the served page
// before writing it; record the grep counts in NOTES):
//
//	[
//	  {"name": "the page stops being served (404)", "serve_status": 404,
//	   "expect_fail": ["page-serves-200"]},
//	  {"name": "...", "from": "<exact string>", "to": "<replacement>",
//	   "expect_fail": ["check-id"]}
//	]
//
// USAGE
//
//	go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_mutants_file.go \
//	    <criteria.json> <mutants.json> <live-url>
//
// TWO DECLARED HARNESS ACCOMMODATIONS, both optional, both uniform across the
// baseline and every mutant (so no mutant's verdict can depend on them), both
// declared in the mutants file where a reviewer reads the mutants themselves:
//
//   - "serve_local": ["/data/x.json", ...] — paths the page's own JS fetches
//     same-origin. Under the redirect harness those become cross-origin and CORS
//     kills them, failing no-console-errors on a page that is CLEAN live (measured
//     twice: robot-hands /data/latest-news.json, webdesign /search.json). Each
//     listed path is fetched once from the live origin at startup and served
//     locally with its content type in EVERY run.
//   - "strip": ["<script ...beacon...>", ...] — exact substrings removed from the
//     page in EVERY run. For third-party telemetry (Cloudflare RUM) that posts to
//     an EXTERNAL origin and can never pass CORS from a localhost harness. Each
//     listed string must occur exactly once, like a mutant's `from`.
//
// A page whose origin 404s one of its own assets still cannot get a green baseline
// here — fix the asset or write a dedicated sibling that declares that deviation
// (the fuel-cost-estimator prover is the precedent).
//
// The mutants file is either a bare ARRAY of mutants (the original form) or an
// object: {"serve_local": [...], "strip": [...], "mutants": [...]}.
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
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`

	// ServeStatus, when non-zero, makes the throwaway server answer the page path
	// with that HTTP status instead of the page body. The one mutant for
	// page_status_ok.
	ServeStatus int `json:"serve_status"`

	// ExpectFail names the checks whose PASS must be contingent on the mutated
	// behaviour. Other checks may also go red (collateral); only these are asserted.
	ExpectFail []string `json:"expect_fail"`
}

var mutants []mutant // loaded from the mutants.json argument

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
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: prove_fence_mutants_file.go <criteria.json> <mutants.json> <live-url>")
		os.Exit(2)
	}
	critPath, mutantsPath, liveURL := os.Args[1], os.Args[2], os.Args[3]

	rawMut, err := os.ReadFile(mutantsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", mutantsPath, err)
		os.Exit(2)
	}
	var serveLocal, strip []string
	trimmed := strings.TrimSpace(string(rawMut))
	if strings.HasPrefix(trimmed, "{") {
		var cfg struct {
			ServeLocal []string `json:"serve_local"`
			Strip      []string `json:"strip"`
			Mutants    []mutant `json:"mutants"`
		}
		if err := json.Unmarshal(rawMut, &cfg); err != nil {
			fmt.Fprintf(os.Stderr, "mutants file does not parse: %v\n", err)
			os.Exit(2)
		}
		serveLocal, strip, mutants = cfg.ServeLocal, cfg.Strip, cfg.Mutants
	} else if err := json.Unmarshal(rawMut, &mutants); err != nil {
		fmt.Fprintf(os.Stderr, "mutants file does not parse: %v\n", err)
		os.Exit(2)
	}
	if len(mutants) == 0 {
		fmt.Fprintln(os.Stderr, "mutants file is EMPTY — an empty mutant list would vacuously pass nothing and prove nothing")
		os.Exit(2)
	}

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

	for _, s := range strip {
		n := strings.Count(string(original), s)
		if n != 1 {
			fmt.Fprintf(os.Stderr, "strip target occurs %d times (must be exactly 1): %q\n", n, s)
			os.Exit(2)
		}
		original = []byte(strings.Replace(string(original), s, "", 1))
	}
	localAssets := map[string][]byte{}
	localTypes := map[string]string{}
	for _, ap := range serveLocal {
		aresp, aerr := http.Get(origin + ap)
		if aerr != nil {
			fmt.Fprintf(os.Stderr, "cannot fetch serve_local %s: %v\n", ap, aerr)
			os.Exit(2)
		}
		ab, rerr := io.ReadAll(aresp.Body)
		_ = aresp.Body.Close()
		if rerr != nil || aresp.StatusCode != 200 {
			fmt.Fprintf(os.Stderr, "fetch serve_local %s: HTTP %d, err %v\n", ap, aresp.StatusCode, rerr)
			os.Exit(2)
		}
		localAssets[ap] = ab
		localTypes[ap] = aresp.Header.Get("Content-Type")
	}

	fmt.Printf("=== proving the fence can fail (mutants: %s) ===\n", mutantsPath)
	if len(serveLocal) > 0 {
		fmt.Printf("serve_local (uniform across ALL runs): %s\n", strings.Join(serveLocal, ", "))
	}
	if len(strip) > 0 {
		fmt.Printf("strip (uniform across ALL runs): %d substring(s) removed from the page\n", len(strip))
	}
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
		if ab, ok := localAssets[r.URL.Path]; ok {
			if ct := localTypes[r.URL.Path]; ct != "" {
				w.Header().Set("Content-Type", ct)
			}
			_, _ = w.Write(ab)
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
			RunID:        "prove-fence-mutfile",
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
		fmt.Printf("--- mutant %d/%d: %s ---\n", i+1, len(mutants), m.Name)

		mutatedBody := original

		switch {
		case m.ServeStatus != 0:
			if m.ServeStatus == http.StatusOK {
				fmt.Println("  FAIL: a 200 'status mutant' changes nothing.")
				verdicts = append(verdicts, verdict{m.Name, false, "status mutant is a no-op"})
				allCaught = false
				continue
			}
			fmt.Printf("  response changed: HTTP 200 -> HTTP %d (body untouched)\n", m.ServeStatus)

		default:
			if !strings.Contains(string(original), m.From) {
				fmt.Printf("  FAIL: target string not present — this mutant is stale: %q\n", m.From)
				verdicts = append(verdicts, verdict{m.Name, false, "target string absent (stale mutant)"})
				allCaught = false
				continue
			}
			mutatedBody = []byte(strings.Replace(string(original), m.From, m.To, 1))
			if string(mutatedBody) == string(original) {
				fmt.Println("  FAIL: replacement was a no-op — the artefact did not change.")
				verdicts = append(verdicts, verdict{m.Name, false, "no-op replacement"})
				allCaught = false
				continue
			}
			fmt.Printf("  artefact changed: %d -> %d bytes\n", len(original), len(mutatedBody))
		}

		failed, skipped, err := run(mutatedBody, m.ServeStatus)
		if err != nil {
			fmt.Printf("  FAIL: run errored: %v\n", err)
			verdicts = append(verdicts, verdict{m.Name, false, "run errored"})
			allCaught = false
			continue
		}
		if len(skipped) > 0 {
			fmt.Printf("  FAIL: %d check(s) were SKIPPED: %s\n", len(skipped), strings.Join(skipped, ", "))
			verdicts = append(verdicts, verdict{m.Name, false, "checks skipped"})
			allCaught = false
			continue
		}
		if d, ok := failed["page-serves-200"]; ok && !namesCheck(m.ExpectFail, "page-serves-200") {
			fmt.Printf("  FAIL: the mutant broke the page itself (page-serves-200: %s).\n", d)
			verdicts = append(verdicts, verdict{m.Name, false, "mutant broke the page, not the behaviour"})
			allCaught = false
			continue
		}

		var missed []string
		for _, want := range m.ExpectFail {
			if d, ok := failed[want]; ok {
				fmt.Printf("  caught by %-42s %s\n", want, d)
			} else {
				missed = append(missed, want)
			}
		}
		var collateral []string
		for id := range failed {
			named := false
			for _, want := range m.ExpectFail {
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
			verdicts = append(verdicts, verdict{m.Name, false, "missed by " + strings.Join(missed, ", ")})
			allCaught = false
		} else {
			verdicts = append(verdicts, verdict{m.Name, true, fmt.Sprintf("%d check(s) went red as named", len(m.ExpectFail))})
		}
		fmt.Println()
	}

	watched := map[string]bool{}
	for _, m := range mutants {
		for _, id := range m.ExpectFail {
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
