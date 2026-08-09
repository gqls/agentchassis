//go:build ignore

// prove_contact_delivery.go — drive the rewritten contact components through
// EVERY destination shape in a real browser and assert what the visitor is told.
//
// WHY. `bugs_open/228` exists because a form printed "Your message has been
// sent" with no destination. The fix is only worth anything if the success
// message is now unreachable without a destination accepting the message — and
// that is a claim about branches, not about code reading well. So every branch
// is driven here, including the two that must NOT produce a success.
//
// The page under test is the LIVE served markup with the component's <script
// src> swapped for the candidate js_content, so the DOM is production's and the
// script is the one about to be persisted.
//
// USAGE
//
//	go run docs/.../scripts/prove_contact_delivery.go <live-url> <candidate.js> <cb|cf>
//
// EXIT 0 only if every case behaves as named.
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/mxschmitt/playwright-go"
)

type sel struct {
	form, status, submit string
	fields               map[string]string
}

var profiles = map[string]sel{
	"cb": {
		form: "#cb-contact-form", status: "#cb-status", submit: "#cb-submit-btn",
		fields: map[string]string{
			"#cb-first-name": "Ada", "#cb-last-name": "Lovelace",
			"#cb-email": "ada@example.com", "#cb-message": "Please quote for 40 grippers, IP65.",
		},
	},
	"cf": {
		form: "#cf-contact-form", status: "#cf-status", submit: ".form-submit",
		fields: map[string]string{
			"#name": "Ada Lovelace", "#email": "ada@example.com",
			"#message": "Please quote for 40 grippers, IP65.",
		},
	},
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: prove_contact_delivery.go <live-url> <candidate.js> <cb|cf>")
		os.Exit(2)
	}
	liveURL, jsPath, which := os.Args[1], os.Args[2], os.Args[3]
	p, ok := profiles[which]
	if !ok {
		fmt.Fprintln(os.Stderr, "third arg must be cb or cf")
		os.Exit(2)
	}

	candidate, err := os.ReadFile(jsPath)
	must(err, "read candidate js")

	origin := liveURL[:strings.Index(liveURL[8:], "/")+8]
	resp, err := http.Get(liveURL)
	must(err, "fetch live page")
	pageHTML, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	must(err, "read live page")

	// Server state the test drives.
	var mu sync.Mutex
	var posted []string
	endpointStatus := 200

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	must(err, "listen")
	mux := http.NewServeMux()
	mux.HandleFunc("/endpoint", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		posted = append(posted, string(b))
		st := endpointStatus
		mu.Unlock()
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(st)
		_, _ = w.Write([]byte(`{}`))
	})

	// The page is served with its own <script src> replaced by the candidate,
	// and the form's action rewritten per case.
	var served struct {
		sync.RWMutex
		body []byte
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/endpoint") {
			return
		}
		served.RLock()
		b := served.body
		served.RUnlock()
		if b == nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(b)
	})
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(ln) }()
	defer func() { _ = srv.Close() }()
	base := "http://" + ln.Addr().String()

	scriptRe := regexp.MustCompile(`<script src="/tools/assets/[^"]+\.js"></script>`)
	formTagRe := regexp.MustCompile(`(?s)<form[^>]*` + regexp.QuoteMeta(strings.TrimPrefix(p.form, "#")) + `[^>]*>`)

	build := func(action string) []byte {
		h := string(pageHTML)
		if which == "cf" && !strings.Contains(h, `id="cf-contact-form"`) {
			// The live page predates the template change this JS ships with, so
			// apply exactly the same three edits apply_contact_form_delivery.py
			// makes — no more — and say so. Proving the script against the DOM
			// it will actually meet is the point; inventing a friendlier DOM
			// would be a test of nothing.
			h = strings.Replace(h, `<form class="contact-form"`,
				`<form class="contact-form" id="cf-contact-form"`, 1)
			h = strings.Replace(h, `<button type="submit" class="form-submit"`,
				`<div class="contact-form-status" id="cf-status" role="alert" aria-live="polite"></div>`+
					`<button type="submit" class="form-submit"`, 1)
			// After the COMPONENT's own </section>, which is where
			// store_generated_component_action.go puts it — NOT after the first
			// </section> on the page. The first version of this harness used
			// strings.Replace(h, "</section>", ..., 1) and landed the script in
			// the hero, i.e. BEFORE the form existed; the script then hit its
			// own `if (!form) return` guard, the browser did a native submit,
			// and all five cases failed in a way that looked like a defect in
			// the component. Placement relative to the markup is the whole
			// contract of a `<script src>` at the end of a section.
			if i := strings.Index(h, `data-component="contact-form"`); i >= 0 {
				if j := strings.Index(h[i:], `</section>`); j >= 0 {
					at := i + j + len(`</section>`)
					h = h[:at] + `<script src="/tools/assets/contact-form.js"></script>` + h[at:]
				}
			}
		}
		// Inline the candidate so no asset fetch is involved.
		h = scriptRe.ReplaceAllString(h, "<script>"+strings.ReplaceAll(string(candidate), "</script>", "<\\/script>")+"</script>")
		// Set (or replace) the action on the component's own form.
		h = formTagRe.ReplaceAllStringFunc(h, func(tag string) string {
			t := regexp.MustCompile(`\s+action="[^"]*"`).ReplaceAllString(tag, "")
			if action != "" {
				t = strings.Replace(t, ">", ` action="`+action+`">`, 1)
			}
			return t
		})
		// Absolute asset URLs so CSS/images do not 404 into the console.
		h = strings.ReplaceAll(h, `href="/`, `href="`+origin+`/`)
		h = strings.ReplaceAll(h, `src="/`, `src="`+origin+`/`)
		return []byte(h)
	}

	pw, err := playwright.Run()
	must(err, "playwright run")
	defer func() { _ = pw.Stop() }()
	browser, err := pw.Chromium.Launch()
	must(err, "launch")
	defer func() { _ = browser.Close() }()

	type outcome struct {
		status         string
		fieldsKept     bool
		browserRefused bool
		navTargets     []string
	}

	run := func(action string, fill map[string]string) outcome {
		served.Lock()
		served.body = build(action)
		served.Unlock()

		page, err := browser.NewPage()
		must(err, "new page")
		defer func() { _ = page.Close() }()

		var nav []string
		var navMu sync.Mutex
		page.OnRequest(func(r playwright.Request) {
			if r.IsNavigationRequest() {
				navMu.Lock()
				nav = append(nav, r.URL())
				navMu.Unlock()
			}
		})

		_, err = page.Goto(base+"/page", playwright.PageGotoOptions{
			WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		})
		must(err, "goto")

		for s, v := range fill {
			if v == "" {
				continue
			}
			_ = page.Locator(s).Fill(v)
		}
		if which == "cb" {
			// The select is part of validation; choose a real option.
			if _, err := page.Locator("#cb-subject").SelectOption(playwright.SelectOptionValues{
				Values: &[]string{"sales"},
			}); err != nil {
				_ = err // an empty-field case may not reach it
			}
		}
		_ = page.Locator(p.submit).Click()
		page.WaitForTimeout(1500)

		st, _ := page.Locator(p.status).TextContent()
		// contact-block carries `novalidate`, so ITS script owns validation and
		// writes the reason into the status element. contact-form does NOT, so
		// the browser's own constraint validation refuses first and the submit
		// event never fires — the reason is shown by Chromium, not by us. Both
		// are correct refusals; asserting only our own status text would have
		// scored the second as a defect (it did, on the first run).
		formValid := true
		if v, err := page.Evaluate(`(sel) => { const f = document.querySelector(sel); return f ? f.checkValidity() : true; }`, p.form); err == nil {
			if b, ok := v.(bool); ok {
				formValid = b
			}
		}
		kept := true
		for s, v := range fill {
			if v == "" {
				continue
			}
			cur, _ := page.Locator(s).InputValue()
			if cur != v {
				kept = false
			}
		}
		navMu.Lock()
		out := outcome{status: strings.TrimSpace(st), fieldsKept: kept, browserRefused: !formValid, navTargets: append([]string{}, nav...)}
		navMu.Unlock()
		return out
	}

	pass := true
	check := func(name string, cond bool, detail string) {
		if cond {
			fmt.Printf("  PASS  %-46s %s\n", name, detail)
			return
		}
		fmt.Printf("  FAIL  %-46s %s\n", name, detail)
		pass = false
	}

	fmt.Printf("=== proving contact delivery: %s against %s ===\n", jsPath, liveURL)

	fmt.Println("\n--- case 1: no destination (action absent) — MUST refuse, MUST NOT say sent ---")
	o := run("", p.fields)
	check("says the form has no destination", strings.Contains(o.status, "no destination configured"), trunc(o.status))
	check("does NOT claim the message was sent", !strings.Contains(strings.ToLower(o.status), "has been sent"), trunc(o.status))
	check("keeps the visitor's text", o.fieldsKept, fmt.Sprintf("fields preserved=%v", o.fieldsKept))

	fmt.Println("\n--- case 2: mailto: destination — hands off, MUST NOT claim receipt ---")
	o = run("mailto:enquiries@example.com?subject=example.com%20enquiry", p.fields)
	check("names the address it is opening", strings.Contains(o.status, "enquiries@example.com"), trunc(o.status))
	check("says it is OPENING a mail app, not that it sent", strings.Contains(o.status, "Opening your email app"), trunc(o.status))
	check("does NOT claim the message was sent", !strings.Contains(strings.ToLower(o.status), "has been sent"), trunc(o.status))
	check("keeps the visitor's text (mail app may not open)", o.fieldsKept, fmt.Sprintf("fields preserved=%v", o.fieldsKept))
	mailtoNav := ""
	for _, u := range o.navTargets {
		if strings.HasPrefix(strings.ToLower(u), "mailto:") {
			mailtoNav = u
		}
	}
	if mailtoNav != "" {
		check("navigates to a mailto carrying the typed text",
			strings.Contains(mailtoNav, "Lovelace") || strings.Contains(mailtoNav, "Lovelace"), trunc(mailtoNav))
	} else {
		// Reached only if a future Chromium stops emitting the navigation
		// request for an unhandled external scheme. It DOES emit it today
		// (measured 2026-08-09), which is why this is a hard FAIL rather than a
		// note: silently degrading to "the visitor sees the right words" would
		// let a broken URL through, and the URL is the whole delivery.
		check("navigates to a mailto carrying the typed text", false,
			"no mailto navigation was observed at all")
	}

	fmt.Println("\n--- case 3: real endpoint returning 200 — MAY say sent, MUST have posted ---")
	mu.Lock()
	posted = nil
	endpointStatus = 200
	mu.Unlock()
	o = run(base+"/endpoint", p.fields)
	mu.Lock()
	gotPosts := len(posted)
	body := strings.Join(posted, "\n")
	mu.Unlock()
	check("a request actually left the browser", gotPosts == 1, fmt.Sprintf("posts=%d", gotPosts))
	check("the post carries the visitor's text", strings.Contains(body, "Lovelace"), trunc(firstLine(body)))
	check("says the message has been sent", strings.Contains(strings.ToLower(o.status), "has been sent"), trunc(o.status))
	check("clears the form only after acceptance", !o.fieldsKept, fmt.Sprintf("fields cleared=%v", !o.fieldsKept))

	fmt.Println("\n--- case 4: real endpoint returning 500 — MUST NOT say sent, MUST keep text ---")
	mu.Lock()
	posted = nil
	endpointStatus = 500
	mu.Unlock()
	o = run(base+"/endpoint", p.fields)
	check("does NOT claim the message was sent", !strings.Contains(strings.ToLower(o.status), "has been sent"), trunc(o.status))
	check("reports the failure with its status code", strings.Contains(o.status, "500"), trunc(o.status))
	check("keeps the visitor's text", o.fieldsKept, fmt.Sprintf("fields preserved=%v", o.fieldsKept))

	fmt.Println("\n--- case 5: invalid input — MUST NOT reach any destination ---")
	mu.Lock()
	posted = nil
	endpointStatus = 200
	mu.Unlock()
	bad := map[string]string{}
	for k := range p.fields {
		bad[k] = ""
	}
	if which == "cb" {
		bad["#cb-first-name"] = "Ada"
	} else {
		bad["#name"] = "Ada Lovelace"
	}
	o = run(base+"/endpoint", bad)
	mu.Lock()
	gotPosts = len(posted)
	mu.Unlock()
	check("nothing was posted", gotPosts == 0, fmt.Sprintf("posts=%d", gotPosts))
	reason := o.status
	if reason == "" && o.browserRefused {
		reason = "(refused by the browser's own constraint validation — this form has no novalidate)"
	}
	check("the visitor is given a reason, by us or by the browser",
		o.status != "" || o.browserRefused, trunc(reason))
	check("does NOT claim the message was sent", !strings.Contains(strings.ToLower(o.status), "has been sent"), trunc(o.status))

	fmt.Println()
	if !pass {
		fmt.Println("RESULT: FAIL")
		os.Exit(1)
	}
	fmt.Println("RESULT: PASS — every destination shape behaves as named, and the success")
	fmt.Println("message was reachable ONLY from a destination that accepted the message.")
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func trunc(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= 110 {
		return s
	}
	return s[:110] + "…"
}

func must(err error, what string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", what, err)
		os.Exit(2)
	}
}
