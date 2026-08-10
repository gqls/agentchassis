// quotecheck — fetch a citation URL exactly as the evidence-freshness sweep
// does, run the REAL Go extractor over it, and dump the visible text so quotes
// can be lifted from it programmatically rather than retyped.
//
// Why a Go program and not python: the day-one gotcha on this lane is that a
// quote extracted by one extractor may not survive the other. Using
// datahelpers.VisibleTextFromHTML here removes that class entirely — anything
// this prints is, by construction, text the sweep will see.
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
)

func main() {
	url := os.Args[1]
	mode := "dump"
	if len(os.Args) > 2 {
		mode = os.Args[2]
	}

	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "agentchassis-citation-verifier/1 (+evidence re-verification)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9,*/*;q=0.1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "HTTP", resp.StatusCode)
		os.Exit(1)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	text := datahelpers.VisibleTextFromHTML(string(body))

	switch mode {
	case "dump":
		fmt.Println(text)
	case "check":
		// remaining args are candidate quotes, one per arg
		bad := 0
		for _, q := range os.Args[3:] {
			ok := datahelpers.QuoteFoundInText(q, text)
			status := "FOUND   "
			if !ok {
				status = "NOTFOUND"
				bad++
			}
			fmt.Printf("%s  %q\n", status, q)
		}
		if bad > 0 {
			os.Exit(2)
		}
	case "checkfile":
		// each line of the file at os.Args[3] is a candidate quote
		raw, err := os.ReadFile(os.Args[3])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		bad := 0
		for _, line := range strings.Split(string(raw), "\n") {
			q := strings.TrimSpace(line)
			if q == "" || strings.HasPrefix(q, "#") {
				continue
			}
			ok := datahelpers.QuoteFoundInText(q, text)
			status := "FOUND   "
			if !ok {
				status = "NOTFOUND"
				bad++
			}
			fmt.Printf("%s  %s\n", status, q)
		}
		if bad > 0 {
			os.Exit(2)
		}
	}
}
