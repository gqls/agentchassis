// FILE: cmd/voicescan/main.go
//
// Operator CLI for the voice-tells layer (SPEC_voice_tells_check). Runs the
// SAME shared scan engine as the voice_tells discovery check over exported
// component HTML — so an operator can see what the platform will flag,
// against live data, without deploying anything. Mirrors cmd/claimscan.
//
// Inputs:
//   -voice <path>      voice spec JSON (the site_specs 'voice' data, which
//                      must contain an enabled voice_gate block)
//   -components <path> TSV export: page_name <TAB> slot_name <TAB> base64(html)
//                      one row per component; '-' reads stdin
//   -longform <csv>    comma-separated page_name prefixes treated as long-form
//                      (default "blog/,guides/,hierarchical")
//
// Export the TSV with the same psql one-liner documented in cmd/claimscan.
// Output: one line per finding and a summary; exit 1 when findings exist.

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
	voicePath := flag.String("voice", "", "path to voice spec JSON (with voice_gate)")
	componentsPath := flag.String("components", "-", "path to TSV export (page, slot, base64 html); '-' for stdin")
	longformPrefixes := flag.String("longform", "blog/,guides/,hierarchical", "page_name prefixes scanned with long-form thresholds")
	flag.Parse()

	if *voicePath == "" {
		fmt.Fprintln(os.Stderr, "voicescan: -voice is required")
		os.Exit(2)
	}
	voiceJSON, err := os.ReadFile(*voicePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "voicescan: read voice spec: %v\n", err)
		os.Exit(2)
	}
	gate, err := datahelpers.ParseVoiceGate(voiceJSON)
	if err != nil {
		fmt.Fprintf(os.Stderr, "voicescan: parse voice_gate: %v\n", err)
		os.Exit(2)
	}
	if gate == nil {
		fmt.Fprintln(os.Stderr, "voicescan: voice spec has no enabled voice_gate — site not opted in")
		os.Exit(2)
	}
	var prefixes []string
	for _, p := range strings.Split(*longformPrefixes, ",") {
		if p = strings.TrimSpace(p); p != "" {
			prefixes = append(prefixes, p)
		}
	}

	in := os.Stdin
	if *componentsPath != "-" {
		f, err := os.Open(*componentsPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "voicescan: open components: %v\n", err)
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
			fmt.Fprintf(os.Stderr, "voicescan: skipping malformed line (%d fields)\n", len(parts))
			continue
		}
		page, slot := parts[0], parts[1]
		htmlBytes, err := base64.StdEncoding.DecodeString(parts[2])
		if err != nil {
			fmt.Fprintf(os.Stderr, "voicescan: %s/%s: base64 decode failed: %v\n", page, slot, err)
			continue
		}
		componentsScanned++

		longForm := false
		for _, p := range prefixes {
			if strings.HasPrefix(page, p) {
				longForm = true
				break
			}
		}
		for _, f := range gate.ScanVoice(datahelpers.ExtractAssertionText(string(htmlBytes)), longForm) {
			total++
			if f.Threshold > 0 {
				fmt.Printf("VOICE    %-45s %-22s %-18s %.2f>%.2f ×%d  %s\n",
					page, slot, f.Check, f.Value, f.Threshold, f.Occurrences, f.Reason)
			} else {
				fmt.Printf("VOICE    %-45s %-22s %-18s %-25q ×%d  %s\n    …%s…\n",
					page, slot, f.Check, f.Matched, f.Occurrences, f.Reason, f.Snippet)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "voicescan: read components: %v\n", err)
		os.Exit(2)
	}

	fmt.Printf("\nvoicescan: %d finding(s) across %d component(s)\n", total, componentsScanned)
	if total > 0 {
		os.Exit(1)
	}
}
