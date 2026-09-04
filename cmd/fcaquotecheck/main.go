// FILE: cmd/fcaquotecheck/main.go
//
// Write-time probe for a candidate citation: does this quote actually survive
// the SAME fetch AND the same extraction the nightly citation refresher uses?
// A quote that does not match reads as `citation_lost` (drift) every day for
// ever, which is a false alarm indistinguishable from a real one.
//
// ⚠ CORRECTED 2026-09-04 — this tool used to do its own bare http.Get and run
// datahelpers.VisibleTextFromHTML over whatever bytes came back. That made it a
// LOOKALIKE of production rather than production, and it broke
// RUNBOOK_lendzy §8g's own rule ("never validate a citation host with an
// instrument other than the one that will re-check it daily") while §8 step 4
// claimed it "calls the production fetch + extraction". It called the
// extraction only.
//
// What that cost: on a PDF this tool returned `false` for a quote that IS in
// the document AND `false` for the deliberately-absent control — measured
// 2026-09-03 against the CMA draft Order (HTTP 200, raw 392,144 bytes, 296,699
// chars of extracted binary noise). Two falses means the instrument is blind,
// but it reads exactly like "you mistyped the quote". A 414-lane author then
// inferred a PRODUCTION drift danger from it and wrote that into a landmine, a
// migration header and four documents. Production was never at risk:
// actions.fetchCitationDocument refuses a non-text content type and routes it
// to fetch_error, which refreshCitationFact reports as `error` and never as
// drift.
//
// It now calls actions.FetchCitationDocumentForProbe — the production fetch —
// so an unsupported content type, a bot-challenge interstitial, a non-200 and a
// UA-differential blank all surface HERE, at write time, in the words
// production will use.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gqls/agentchassis/platform/orchestration/actions"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: fcaquotecheck <url> <quote> [more quotes...] [\"zzz absent control\"]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "ALWAYS pass a deliberately-absent control as the last quote.")
		fmt.Fprintln(os.Stderr, "Expect true for a real quote and false for the control. If EVERY quote")
		fmt.Fprintln(os.Stderr, "including the control returns false, the instrument is blind, not your quote wrong.")
		os.Exit(2)
	}
	url := os.Args[1]

	text, err := actions.FetchCitationDocumentForProbe(context.Background(), url)
	if err != nil {
		// This is the branch that used to be silently absent. Production
		// classifies it fetch_error -> outcome "error", NOT drift.
		fmt.Printf("NOT VERIFIABLE UNATTENDED: %v\n", err)
		fmt.Println()
		fmt.Println("This is what the nightly refresher would see. It reports this as an ERROR")
		fmt.Println("(unknown), never as citation_lost drift — an unreadable source is not evidence")
		fmt.Println("the fact is wrong.")
		fmt.Println()
		fmt.Println("DO NOT register a source.citation against this URL. Either cite an HTML source")
		fmt.Println("carrying the same fact, or keep the citation and set \"reverifiable\": false with a")
		fmt.Println("human attestation of having read it (evidence_citations.go's header states this is")
		fmt.Println("the intended path for a PDF), plus staleness_days so it still ages by policy.")
		os.Exit(1)
	}

	fmt.Printf("FETCHED OK  visible=%d chars\n", len(text))
	trues, falses := 0, 0
	for _, q := range os.Args[2:] {
		found := datahelpers.QuoteFoundInText(q, text)
		if found {
			trues++
		} else {
			falses++
		}
		fmt.Printf("%-6v  %.90s...\n", found, q)
	}
	if trues == 0 && falses > 1 {
		fmt.Println()
		fmt.Println("⚠ EVERY quote returned false, including what should be your control.")
		fmt.Println("  That pattern means the extraction found nothing usable — suspect the")
		fmt.Println("  instrument or the page, not your transcription. Check what was fetched.")
	}
}
