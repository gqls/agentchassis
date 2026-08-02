//go:build ignore

// prove_fence_can_fail_teaser_reveal_panel.go — the S2 mutation discipline,
// applied to fence_teaser_reveal_panel.json.
//
// WHY THIS FILE EXISTS RATHER THAN REUSING prove_fence_can_fail.go DIRECTLY.
// That script's name and RUNBOOK §8's own phrasing ("go run .../prove_fence_can_fail.go
// <fence.json> <url>") read as if it is a generic harness driven by the fence file. It is
// not: its `mutants` slice is hardcoded to tool-review-council-simulator's own source
// strings (`el.passN.textContent`, `rcs-bar-row`, `applyPreset('typical')`, ...), so every
// mutant in it is stale against any other page — confirmed empirically: run against this
// component's live page, all 14 string-edit mutants report "target string not present" and
// only the 3 generic ones (404, wide-div, console-error) fire. That is precisely the class
// of mistake this lane's own HANDOFF_2026-07-31c §7 names: a step assumed to be reusable
// "wiring" that turns out to embed an assumption specific to the first thing it was built
// for. This file is the sibling the discipline actually requires, built by copying the
// proven architecture (baseline-first, one string-replace per mutant, targeted-not-
// demolished, coverage table) and swapping in this component's own mutants.
//
// ONE EXTENSION OVER THE ORIGINAL: an optional per-mutant ASSET OVERRIDE. Nine of this
// fence's twelve checks are provable by editing the page body, exactly like the original.
// But "opening a card closes its siblings" is implemented in `/assets/js/snippets.js` —
// a fleet-shared file loaded by <script src>, not inlined in this page's own markup — and
// the original harness 302-redirects every non-page request to the live origin, so it has
// no way to serve a mutated copy of an external asset. This file adds that: a mutant may
// name one extra path or the throwaway server serves a locally mutated version of it while
// everything else still redirects live, same fairness rule as the page body (baseline must
// run the real unmutated asset via a normal redirect; only the one mutant that needs it
// gets an override).
//
// USAGE
//
//	go run docs/agent_docs/docs024_key_docs_latest/staged_component_build/scripts/prove_fence_can_fail_teaser_reveal_panel.go \
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

	// assetPath/assetFrom/assetTo, when assetPath is non-empty, make the throwaway
	// server serve a locally mutated copy of that ONE extra path (fetched once at
	// startup, one string-replace applied) instead of redirecting it to the live
	// origin. Every other mutant leaves assetPath empty and gets the normal redirect.
	assetPath string
	assetFrom string
	assetTo   string

	// expectFail names the checks whose PASS must be contingent on the mutated
	// behaviour. Other checks may also go red (collateral); only these are asserted.
	expectFail []string
}

// Every `from` / `assetFrom` below was verified to occur EXACTLY ONCE in the served
// page / asset before this file was written (see NOTES for the grep counts).
var mutants = []mutant{
	{
		name:        "the page stops being served (404)",
		serveStatus: http.StatusNotFound,
		expectFail:  []string{"page-serves-200"},
	},
	{
		name:       "the panel loses the identity acceptance would resolve it by",
		from:       `data-component="teaser-reveal-panel"`,
		to:         `data-component="something-else-entirely"`,
		expectFail: []string{"panel-renders"},
	},
	{
		name:       "the first card's key is renamed out from under its selector",
		from:       `data-trp-key="orchestration"`,
		to:         `data-trp-key="orchestration-renamed"`,
		expectFail: []string{"first-card-present"},
	},
	{
		name:       "the last card's key is renamed out from under its selector",
		from:       `data-trp-key="our-own-sites"`,
		to:         `data-trp-key="our-own-sites-renamed"`,
		expectFail: []string{"last-card-present"},
	},
	{
		// PLAN decision 2/6: the full paragraph stays in the DOM at all times, only
		// CSS decides what paints, so the claims gate and crawlers can always read
		// it. This mutant renames the class that carries that paragraph's lead
		// strong tag on the ONE card the check names, leaving the other five alone.
		name:       "the closed-state body text is dropped from the DOM instead of merely hidden",
		from:       `<strong class="trp__body-lead">Workflows live in the database, not in the code.</strong>`,
		to:         `<strong class="trp__body-lead-renamed">Workflows live in the database, not in the code.</strong>`,
		expectFail: []string{"closed-card-full-text-stays-in-dom"},
	},
	{
		// bugs_open/157's exact class of defect: present in the DOM, collapsed to
		// zero on screen. `has_visible_area` exists because `selector_exists` alone
		// cannot see this.
		name:       "the whole panel collapses to zero height",
		from:       "</body>",
		to:         `<style>.trp{display:none!important}</style></body>`,
		expectFail: []string{"panel-has-visible-area"},
	},
	{
		name:       "the first card collapses to zero height (same class, scoped to the .trp__card rule)",
		from:       "</body>",
		to:         `<style>.trp__card{display:none!important}</style></body>`,
		expectFail: []string{"card-has-visible-area"},
	},
	{
		// PLAN's own verification contract: "forcing .open = true is not a click...
		// dispatch the visitor's real gesture." This mutant leaves the native
		// <details>/<summary> disclosure fully intact and only prevents the CLICK
		// from ever reaching the first card's summary -- if this check still passed,
		// it would mean the check was never actually driving a real gesture.
		name:       "the first card's summary stops accepting clicks (pointer-events:none, scoped to that one card)",
		from:       "</body>",
		to:         `<style>details.trp__card[data-trp-key="orchestration"] summary{pointer-events:none!important}</style></body>`,
		expectFail: []string{"real-click-opens-first-card"},
	},
	{
		name:       "the second card's summary stops accepting clicks (pointer-events:none, scoped to that one card)",
		from:       "</body>",
		to:         `<style>details.trp__card[data-trp-key="data-verification"] summary{pointer-events:none!important}</style></body>`,
		expectFail: []string{"real-click-opens-second-card"},
	},
	{
		// The sibling-close behaviour is NOT in this page's markup at all -- it is
		// the 'toggle' listener in the fleet-shared /assets/js/snippets.js. This is
		// the one mutant that needs the asset override: everything else about the
		// page is untouched, but this one path is served locally with the sibling
		// -close line disabled.
		name:       "the sibling-close listener is disabled in snippets.js (opening one card no longer closes the others)",
		assetPath:  "/assets/js/snippets.js",
		assetFrom:  "if (other !== card) { other.open = false; }",
		assetTo:    "if (other !== card) { /* sibling-close disabled by mutant */ }",
		expectFail: []string{"opening-a-sibling-closes-the-first"},
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

var servedAsset struct {
	sync.RWMutex
	path string // "" means no override active -- redirect this path like any other
	body []byte
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: prove_fence_can_fail_teaser_reveal_panel.go <criteria.json> <live-url>")
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

	// Fetch snippets.js once, straight from the live origin, so the one asset-override
	// mutant has an unmutated base to edit -- fetched fresh here, not copied from any
	// earlier session's read, exactly like `original` above.
	snippetsURL := origin + "/assets/js/snippets.js"
	snipResp, err := http.Get(snippetsURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot fetch %s: %v\n", snippetsURL, err)
		os.Exit(2)
	}
	originalSnippets, err := io.ReadAll(snipResp.Body)
	_ = snipResp.Body.Close()
	if err != nil || snipResp.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "fetch %s: HTTP %d, err %v\n", snippetsURL, snipResp.StatusCode, err)
		os.Exit(2)
	}

	fmt.Printf("=== proving the fence can fail (teaser-reveal-panel) ===\n")
	fmt.Printf("fence: %s (%d checks, desktop only)\n", critPath, len(checks))
	fmt.Printf("page:  %s (%d bytes)\n", liveURL, len(original))
	fmt.Printf("snippets.js: %s (%d bytes) -- redirected live except for its one dedicated mutant\n", snippetsURL, len(originalSnippets))
	fmt.Printf("assets other than the page and (for one mutant) snippets.js are 302'd to %s\n\n", origin)

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
		servedAsset.RLock()
		aPath, aBody := servedAsset.path, servedAsset.body
		servedAsset.RUnlock()
		if aPath != "" && r.URL.Path == aPath {
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
			_, _ = w.Write(aBody)
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

	run := func(body []byte, status int, assetPath string, assetBody []byte) (failed map[string]string, skipped []string, err error) {
		served.Lock()
		served.body, served.status = body, status
		served.Unlock()
		servedAsset.Lock()
		servedAsset.path, servedAsset.body = assetPath, assetBody
		servedAsset.Unlock()

		res, err := action.Execute(context.Background(), browserrunner.RunChecksRequest{
			RunID:        "prove-fence-trp",
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
	baseFail, baseSkip, err := run(original, 0, "", nil)
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
	fmt.Printf("  green: all %d checks pass on the local copy, matching the live URL.\n\n", len(checks))

	var verdicts []verdict
	allCaught := true

	for i, m := range mutants {
		fmt.Printf("--- mutant %d/%d: %s ---\n", i+1, len(mutants), m.name)

		mutatedBody := original
		mutatedAssetPath := ""
		var mutatedAssetBody []byte

		switch {
		case m.serveStatus != 0:
			if m.serveStatus == http.StatusOK {
				fmt.Println("  FAIL: a 200 'status mutant' changes nothing.")
				verdicts = append(verdicts, verdict{m.name, false, "status mutant is a no-op"})
				allCaught = false
				continue
			}
			fmt.Printf("  response changed: HTTP 200 -> HTTP %d (body untouched)\n", m.serveStatus)

		case m.assetPath != "":
			if !strings.Contains(string(originalSnippets), m.assetFrom) {
				fmt.Printf("  FAIL: target string not present in %s — this mutant is stale: %q\n", m.assetPath, m.assetFrom)
				verdicts = append(verdicts, verdict{m.name, false, "asset target string absent (stale mutant)"})
				allCaught = false
				continue
			}
			mutatedAssetPath = m.assetPath
			mutatedAssetBody = []byte(strings.Replace(string(originalSnippets), m.assetFrom, m.assetTo, 1))
			if string(mutatedAssetBody) == string(originalSnippets) {
				fmt.Println("  FAIL: asset replacement was a no-op — the asset did not change.")
				verdicts = append(verdicts, verdict{m.name, false, "no-op asset replacement"})
				allCaught = false
				continue
			}
			fmt.Printf("  asset %s changed: %d -> %d bytes\n", m.assetPath, len(originalSnippets), len(mutatedAssetBody))

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

		failed, skipped, err := run(mutatedBody, m.serveStatus, mutatedAssetPath, mutatedAssetBody)
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
