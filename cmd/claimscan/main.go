// FILE: cmd/claimscan/main.go
//
// Operator CLI for the claims-verification layer (SPEC_claims_verification).
// Runs the SAME shared scan engine as the deploy gate (validate_page_content
// check 8) and the post-deploy audit (check_unverified_claims) over exported
// component HTML — so an operator can verify what the platform will flag,
// against live data, without deploying anything.
//
// Inputs:
//   -evidence <path>   evidence_base JSON (the site_specs 'evidence_base' data).
//                      OPTIONAL since bugs_open/104: omit it to scan exactly as
//                      an unarmed site is scanned — fleet-wide patterns only.
//   -no-global         exclude the fleet-wide banned-claim set and scan ONLY
//                      -evidence. For testing a CANDIDATE pattern set in
//                      isolation; the default includes the fleet-wide set
//                      because that is what the platform enforces, and a dry run
//                      that quietly omits half the rules under-reports.
//   -components <path> TSV export: page_name <TAB> slot_name <TAB> base64(html)
//                      [<TAB> page_type]; one row per component; '-' reads stdin
//
// The 4th column is OPTIONAL and is pages.page_type. Supply it: without it every
// page reads as UNKNOWN, which scans prose numbers on every page — so a scan run
// without it will report the editorial-page false positives that the platform
// itself no longer raises (bugs_open/102), and the tool will disagree with the
// gate it exists to predict. Warned about on stderr when absent.
//
// Export the TSV from the live DB with:
//
//	kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
//	  "SELECT p.name || E'\t' || COALESCE(pc.slot_name,'') || E'\t' ||
//	          replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n', '') ||
//	          E'\t' || COALESCE(p.page_type,'')
//	   FROM page_components pc JOIN pages p ON p.id = pc.page_id
//	   WHERE p.site_id = '<site>' AND pc.rendered_html IS NOT NULL
//	     AND pc.rendered_html <> '' AND pc.locked_at IS NULL"
//
// Output: one line per finding (page, slot, check, matched, snippet) and a
// summary count. Exit code 1 when findings exist, 0 when clean — usable as a
// scripted acceptance gate.

package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func main() {
	evidencePath := flag.String("evidence", "", "path to evidence_base JSON ('' = fleet-wide set only, as an unarmed site is scanned)")
	componentsPath := flag.String("components", "-", "path to TSV export (page, slot, base64 html); '-' for stdin")
	noGlobal := flag.Bool("no-global", false,
		"exclude the fleet-wide banned-claim set, scanning ONLY -evidence. Use when testing a candidate pattern set in isolation; the default matches what the platform actually enforces")
	showSuppressed := flag.Bool("show-suppressed", false,
		"print each match the negation guard removed ('negated' lines). They are never findings; use this to see what the guard is doing to real copy")
	flag.Parse()

	// eb may stay nil: since bugs_open/104 the banned-claim scan is fleet-wide,
	// so "no register" is a real thing to reproduce, not an error. -evidence is
	// therefore only required when -no-global leaves nothing else to scan with.
	var eb *datahelpers.EvidenceBase
	if *evidencePath != "" {
		evidenceJSON, err := os.ReadFile(*evidencePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "claimscan: read evidence: %v\n", err)
			os.Exit(2)
		}
		eb, err = datahelpers.ParseEvidenceBase(evidenceJSON)
		if err != nil {
			fmt.Fprintf(os.Stderr, "claimscan: parse evidence: %v\n", err)
			os.Exit(2)
		}
		if eb == nil && *noGlobal {
			fmt.Fprintln(os.Stderr, "claimscan: evidence base is empty and -no-global set — nothing to scan against")
			os.Exit(2)
		}
	} else if *noGlobal {
		fmt.Fprintln(os.Stderr, "claimscan: -no-global needs -evidence — that combination scans against nothing")
		os.Exit(2)
	}
	if !*noGlobal {
		fmt.Fprintf(os.Stderr, "claimscan: including %d fleet-wide banned-claim pattern(s); -no-global to exclude\n",
			datahelpers.GlobalBannedClaimCount())
	}

	in := os.Stdin
	if *componentsPath != "-" {
		f, err := os.Open(*componentsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "claimscan: open components: %v\n", err)
			os.Exit(2)
		}
		defer f.Close()
		in = f
	}

	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)

	var total, suppressedTotal, componentsScanned, withoutPageType int
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) < 3 {
			fmt.Fprintf(os.Stderr, "claimscan: skipping malformed line (%d fields)\n", len(parts))
			continue
		}
		page, slot := parts[0], parts[1]
		// 4th field optional: pages.page_type. Absent means UNKNOWN, which
		// scans everything — the same fail-noisy direction the platform takes.
		pageType := ""
		if len(parts) == 4 {
			pageType = strings.TrimSpace(parts[3])
		}
		if pageType == "" {
			withoutPageType++
		}
		htmlBytes, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "claimscan: %s/%s: base64 decode failed: %v\n", page, slot, err)
			continue
		}
		componentsScanned++

		blocks := datahelpers.ExtractAssertionText(string(htmlBytes))
		banned, suppressed := datahelpers.ScanAllBannedClaimsWithSuppressed(blocks, eb)
		if *noGlobal {
			banned = eb.ScanBannedClaims(blocks)
			suppressed = subtractByPattern(eb.ScanBannedClaimsIgnoringNegation(blocks), banned)
		}
		for _, f := range banned {
			total++
			fmt.Printf("BANNED   %-45s %-22s %-30q ×%d  %s\n    …%s…\n",
				page, slot, f.Matched, f.Occurrences, f.Reason, f.Snippet)
		}
		// NOT findings, and not counted: matches the negation guard removed
		// because the sentence denies the claim rather than making it. Printed
		// so the guard cannot suppress silently — a dry run that shows nothing
		// must be distinguishable from a guard that ate everything.
		for _, f := range suppressed {
			suppressedTotal++
			if *showSuppressed {
				fmt.Printf("negated  %-45s %-22s %-30q      %s\n    …%s…\n",
					page, slot, f.Matched, f.Reason, f.Snippet)
			}
		}
		for _, f := range eb.ScanUnregisteredNumbers(blocks,
			datahelpers.ClaimSurface{PageType: pageType}) {
			total++
			fmt.Printf("NUMBER   %-45s %-22s %-30q ×%d  %s\n    …%s…\n",
				page, slot, f.Matched, f.Occurrences, f.Reason, f.Snippet)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "claimscan: read components: %v\n", err)
		os.Exit(2)
	}

	if withoutPageType > 0 {
		fmt.Fprintf(os.Stderr,
			"claimscan: WARNING — %d of %d component(s) had no page_type (4th TSV field). "+
				"Those scanned as UNKNOWN, i.e. prose numbers checked on every page type. "+
				"Findings on guide/blog-post/tool pages here are ones the platform itself "+
				"no longer raises (bugs_open/102) — add the column to compare like with like.\n",
			withoutPageType, componentsScanned)
	}

	fmt.Printf("\nclaimscan: %d finding(s) across %d component(s)\n", total, componentsScanned)
	if suppressedTotal > 0 {
		fmt.Printf("claimscan: %d match(es) suppressed by the negation guard (not findings)%s\n",
			suppressedTotal, map[bool]string{true: "", false: " — -show-suppressed to list them"}[*showSuppressed])
	}
	if total > 0 {
		os.Exit(1)
	}
}

// subtractByPattern returns the members of `all` whose pattern is not in `kept`.
// Only needed for the -no-global path, where the two scans are run directly
// rather than through ScanAllBannedClaimsWithSuppressed.
func subtractByPattern(all, kept []datahelpers.ClaimFinding) []datahelpers.ClaimFinding {
	keptPatterns := make(map[string]bool, len(kept))
	for _, f := range kept {
		keptPatterns[f.Pattern] = true
	}
	var out []datahelpers.ClaimFinding
	for _, f := range all {
		if !keptPatterns[f.Pattern] {
			out = append(out, f)
		}
	}
	return out
}
