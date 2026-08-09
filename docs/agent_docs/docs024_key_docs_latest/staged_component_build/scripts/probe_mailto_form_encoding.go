//go:build ignore

// probe_mailto_form_encoding.go — settle, at a real browser, what a <form> whose
// action is a mailto: actually does with the visitor's typed text.
//
// WHY THIS EXISTS. `bugs_open/228` recorded, explicitly marked [UNVERIFIED], that
// the sibling `contact-form` component submits to `action="mailto:…"
// method="POST"` on 13 live pages and that a current browser may not deliver the
// body. Reasoning from the HTML spec is not this lane's standard — the rule is to
// watch the branch. So this drives the three combinations a form can use against
// a mailto target and reports, for each, the URL the browser actually navigates
// to and whether the typed text survives in it.
//
// HOW IT CAPTURES A mailto: NAVIGATION. Chromium treats mailto: as an external
// protocol and headless Chromium neither navigates nor reports it as a request,
// so page.on("request") sees nothing. The observable that DOES exist is the
// navigation attempt itself: overriding window.open and hooking the submit event
// is not enough (a native form submit does not route through either). So instead
// the harness rewrites the form's action from `mailto:` to `http://127.0.0.1:PORT/
// capture?original=<the mailto>` immediately before submitting, keeping method and
// enctype EXACTLY as authored. The local server then records the method, the
// Content-Type and the raw body the browser produced. That answers the question
// that actually decides the fix — does the browser put the visitor's text in the
// BODY (which a mailto cannot carry as a body, so it must become ?body=) or in
// the QUERY (which would destroy the ?subject= already in the action) — without
// depending on Chromium's external-protocol behaviour at all.
//
// USAGE
//
//	go run docs/.../scripts/probe_mailto_form_encoding.go
//
// It prints one block per combination and makes no claim it did not measure.
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/mxschmitt/playwright-go"
)

type capture struct {
	Method      string
	ContentType string
	Body        string
	RawQuery    string
}

func main() {
	captured := make(chan capture, 8)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		os.Exit(2)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/capture", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		captured <- capture{
			Method:      r.Method,
			ContentType: r.Header.Get("Content-Type"),
			Body:        string(b),
			RawQuery:    r.URL.RawQuery,
		}
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body>captured</body></html>"))
	})

	page := func(method, enctype string) string {
		enc := ""
		if enctype != "" {
			enc = fmt.Sprintf(` enctype=%q`, enctype)
		}
		// The field names are the live component's own, so the answer is about
		// THAT form, not a synthetic one.
		return fmt.Sprintf(`<html><body>
<form id="f" action="mailto:enquiries@example.com?subject=example.com enquiry" method=%q%s>
  <input name="first_name" value="Ada">
  <input name="last_name" value="Lovelace">
  <input name="email" value="ada@example.com">
  <textarea name="message">Please quote for 40 grippers.</textarea>
  <button type="submit" id="go">Send</button>
</form></body></html>`, method, enc)
	}

	var served struct{ body string }
	mux.HandleFunc("/form", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(served.body))
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer func() { _ = srv.Close() }()
	base := "http://" + ln.Addr().String()

	pw, err := playwright.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "playwright: %v\n", err)
		os.Exit(2)
	}
	defer func() { _ = pw.Stop() }()
	browser, err := pw.Chromium.Launch()
	if err != nil {
		fmt.Fprintf(os.Stderr, "launch: %v\n", err)
		os.Exit(2)
	}
	defer func() { _ = browser.Close() }()

	type combo struct{ method, enctype, label string }
	combos := []combo{
		{"POST", "", "POST + default enctype (application/x-www-form-urlencoded) — WHAT THE LIVE contact-form USES"},
		{"POST", "text/plain", "POST + text/plain — the classic mailto form"},
		{"GET", "", "GET — form data replaces the URL query"},
	}

	fmt.Println("=== what a mailto: <form> hands the transport, measured at Chromium ===")
	fmt.Println("The action is rewritten to a local capture URL immediately before submit;")
	fmt.Println("method and enctype are exactly as authored. The mailto: string is in ?original=.")
	fmt.Println()

	for _, c := range combos {
		served.body = page(c.method, c.enctype)
		p, err := browser.NewPage()
		if err != nil {
			fmt.Fprintf(os.Stderr, "new page: %v\n", err)
			os.Exit(2)
		}
		if _, err := p.Goto(base + "/form"); err != nil {
			fmt.Fprintf(os.Stderr, "goto: %v\n", err)
			os.Exit(2)
		}
		// Swap the target, preserving everything the browser uses to encode.
		if _, err := p.Evaluate(`(cap) => {
			const f = document.getElementById('f');
			const original = f.getAttribute('action');
			f.setAttribute('action', cap + '?original=' + encodeURIComponent(original));
		}`, base+"/capture"); err != nil {
			fmt.Fprintf(os.Stderr, "evaluate: %v\n", err)
			os.Exit(2)
		}
		if err := p.Locator("#go").Click(); err != nil {
			fmt.Fprintf(os.Stderr, "click: %v\n", err)
			os.Exit(2)
		}
		got := <-captured
		_ = p.Close()

		fmt.Printf("--- %s ---\n", c.label)
		fmt.Printf("  transport method : %s\n", got.Method)
		fmt.Printf("  content-type     : %s\n", emptyAs(got.ContentType, "(none)"))
		fmt.Printf("  query            : %s\n", truncate(got.RawQuery, 200))
		fmt.Printf("  body             : %s\n", emptyAs(truncate(got.Body, 300), "(empty)"))
		inBody := strings.Contains(got.Body, "Lovelace")
		inQuery := strings.Contains(got.RawQuery, "Lovelace")
		switch {
		case inBody:
			fmt.Println("  VERDICT          : the typed text is in the BODY.")
			fmt.Println("                     A mailto: URL has no body, so the browser must fold it")
			fmt.Println("                     into ?body= or drop it — see the navigation note below.")
		case inQuery:
			fmt.Println("  VERDICT          : the typed text is in the QUERY, which REPLACES the")
			fmt.Println("                     action's existing query — so ?subject= is destroyed and")
			fmt.Println("                     each field becomes a mailto header (ignored except a few).")
		default:
			fmt.Println("  VERDICT          : the typed text reached NEITHER body nor query — lost.")
		}
		fmt.Println()
	}

	fmt.Println("Read this together with the component fix: the design that removes the")
	fmt.Println("ambiguity entirely is to build the mailto: URL in JS with explicit subject=")
	fmt.Println("and body= parameters and navigate to it, rather than submitting a form to it.")
}

func emptyAs(s, alt string) string {
	if strings.TrimSpace(s) == "" {
		return alt
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
