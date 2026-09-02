// FILE: cmd/fcaquotecheck/main.go
//
// Throwaway probe for the lendzy_co_uk lane: does a candidate FCA Handbook
// quote actually survive the SAME extraction and matching the daily citation
// refresher uses? A quote that does not match reads as `citation_lost` (drift)
// every day for ever, which is a false alarm indistinguishable from a real one.
//
// It deliberately calls datahelpers.VisibleTextFromHTML / QuoteFoundInText —
// the production functions — rather than re-implementing the extraction, since
// a mirror passes happily while production disagrees.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func main() {
	url := os.Args[1]
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("FETCH ERROR:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	text := datahelpers.VisibleTextFromHTML(string(body))
	fmt.Printf("HTTP %d  raw=%d  visible=%d\n", resp.StatusCode, len(body), len(text))
	for _, q := range os.Args[2:] {
		fmt.Printf("%-6v  %.90s...\n", datahelpers.QuoteFoundInText(q, text), q)
	}
}
