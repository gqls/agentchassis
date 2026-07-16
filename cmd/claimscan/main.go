// FILE: cmd/claimscan/main.go
//
// Operator CLI for the claims-verification layer (SPEC_claims_verification).
// Runs the SAME shared scan engine as the deploy gate (validate_page_content
// check 8) and the post-deploy audit (check_unverified_claims) over exported
// component HTML — so an operator can verify what the platform will flag,
// against live data, without deploying anything.
//
// Inputs:
//   -evidence <path>   evidence_base JSON (the site_specs 'evidence_base' data)
//   -components <path> TSV export: page_name <TAB> slot_name <TAB> base64(html)
//                      one row per component; '-' reads stdin
//
// Export the TSV from the live DB with:
//
//	kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
//	  "SELECT p.name || E'\t' || COALESCE(pc.slot_name,'') || E'\t' ||
//	          replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n', '')
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
	evidencePath := flag.String("evidence", "", "path to evidence_base JSON")
	componentsPath := flag.String("components", "-", "path to TSV export (page, slot, base64 html); '-' for stdin")
	flag.Parse()

	if *evidencePath == "" {
		fmt.Fprintln(os.Stderr, "claimscan: -evidence is required")
		os.Exit(2)
	}
	evidenceJSON, err := os.ReadFile(*evidencePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "claimscan: read evidence: %v\n", err)
		os.Exit(2)
	}
	eb, err := datahelpers.ParseEvidenceBase(evidenceJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "claimscan: parse evidence: %v\n", err)
		os.Exit(2)
	}
	if eb == nil {
		fmt.Fprintln(os.Stderr, "claimscan: evidence base is empty — nothing to scan against")
		os.Exit(2)
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

	var total, componentsScanned int
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			fmt.Fprintf(os.Stderr, "claimscan: skipping malformed line (%d fields)\n", len(parts))
			continue
		}
		page, slot := parts[0], parts[1]
		htmlBytes, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "claimscan: %s/%s: base64 decode failed: %v\n", page, slot, err)
			continue
		}
		componentsScanned++

		blocks := datahelpers.ExtractAssertionText(string(htmlBytes))
		for _, f := range eb.ScanBannedClaims(blocks) {
			total++
			fmt.Printf("BANNED   %-45s %-22s %-30q ×%d  %s\n    …%s…\n",
				page, slot, f.Matched, f.Occurrences, f.Reason, f.Snippet)
		}
		for _, f := range eb.ScanUnregisteredNumbers(blocks) {
			total++
			fmt.Printf("NUMBER   %-45s %-22s %-30q ×%d  %s\n    …%s…\n",
				page, slot, f.Matched, f.Occurrences, f.Reason, f.Snippet)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "claimscan: read components: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("\nclaimscan: %d finding(s) across %d component(s)\n", total, componentsScanned)
	if total > 0 {
		os.Exit(1)
	}
}
