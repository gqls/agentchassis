package main

// service_test.go — end-to-end state-machine test. FakeProvider (no Stripe) +
// stubbed engine (no LLM spend). Run: go test ./...

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func newTestApp(tweaks ...func(*Config)) (*App, *[][3]string) {
	cfg := Config{
		PriceGBP: 199, AutoDeliver: true, ReviewBeforePay: false, PublicBaseURL: "http://test",
		InternalAPIKey: "testkey", OperatorEmail: "ops@test", MaxActive: 2,
		AllowedOrigins: []string{"*"},
	}
	for _, tw := range tweaks {
		tw(&cfg)
	}
	store, _ := NewStore("") // in-memory
	app := NewApp(cfg, store, &FakeProvider{publicBaseURL: "http://test"})
	app.engine = func(d, a, s string) (renderedReport, error) {
		return renderedReport{Text: "# Stub report for " + d, HTML: "<h1>Stub report for " + d + "</h1>"}, nil
	}
	sent := &[][3]string{}
	app.deliver = func(to, subj, body string) { *sent = append(*sent, [3]string{to, subj, body}) }
	app.deliverHTML = func(to, subj, text, htmlBody string) { *sent = append(*sent, [3]string{to, subj, text}) }
	app.dispatch = func(f func()) { f() } // run fulfilment inline for deterministic asserts
	return app, sent
}

func reqID(sent *[][3]string) string {
	for i := len(*sent) - 1; i >= 0; i-- {
		if s := (*sent)[i][1]; strings.Contains(s, "New report request ") {
			return strings.SplitN(strings.SplitN(s, "New report request ", 2)[1], " ", 2)[0]
		}
	}
	return ""
}

func TestFlow(t *testing.T) {
	app, sent := newTestApp()
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	cl := srv.Client()

	postForm := func(path string, v url.Values) *http.Response {
		r, err := cl.PostForm(srv.URL+path, v)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	postJSON := func(path, body, key string) *http.Response {
		req, _ := http.NewRequest("POST", srv.URL+path, strings.NewReader(body))
		req.Header.Set("content-type", "application/json")
		if key != "" {
			req.Header.Set("X-Internal-Key", key)
		}
		r, err := cl.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	jsonField := func(r *http.Response, field string) string {
		b, _ := io.ReadAll(r.Body)
		m := map[string]any{}
		json.Unmarshal(b, &m)
		if v, ok := m[field]; ok {
			return strings.TrimSuffix(strings.TrimPrefix(toStr(v), ""), "")
		}
		return ""
	}
	makeRequest := func(email string) string {
		postForm("/request", url.Values{"name": {"Sam"}, "email": {email},
			"business": {"acme.co.uk"}, "audience": {"aud"}, "notes": {"n"}})
		return reqID(sent)
	}
	pass := 0
	ok := func(label string, cond bool) {
		if !cond {
			t.Fatalf("FAIL: %s", label)
		}
		pass++
		t.Logf("ok  %s", label)
	}

	// health & capacity
	ok("health ok", jsonField(get(t, cl, srv.URL+"/health"), "ok") == "true")
	capR := get(t, cl, srv.URL+"/capacity")
	ok("capacity open initially", jsonField(capR, "open") == "true")

	// happy path
	*sent = (*sent)[:0]
	id := makeRequest("buyer@x.com")
	o, _ := app.store.Get(id)
	ok("order is 'requested'", o.Status == "requested")
	ok("operator notified", reqID(sent) == id)

	ok("confirm without key 401", postJSON("/confirm", `{"order_id":"`+id+`"}`, "").StatusCode == 401)
	r := postJSON("/confirm", `{"order_id":"`+id+`"}`, "testkey")
	ok("confirm -> awaiting_payment", r.StatusCode == 200 && jsonField2(r) == "awaiting_payment")

	*sent = (*sent)[:0]
	evt := `{"event_id":"evt_1","type":"checkout.session.completed","order_id":"` + id + `","paid":true}`
	r = postJSON("/stripe/webhook", evt, "")
	ok("webhook accepted", jsonField(r, "status") == "accepted")
	o, _ = app.store.Get(id)
	ok("order delivered (AUTO_DELIVER)", o.Status == "delivered")
	ok("report emailed to buyer", anySent(sent, "buyer@x.com", "Stub report"))

	// idempotency
	r = postJSON("/stripe/webhook", evt, "")
	ok("duplicate webhook ignored", jsonField(r, "status") == "duplicate_ignored")

	// decline
	*sent = (*sent)[:0]
	id2 := makeRequest("decline@x.com")
	r = postJSON("/decline", `{"order_id":"`+id2+`","reason":"no edge"}`, "testkey")
	o2, _ := app.store.Get(id2)
	ok("declined", r.StatusCode == 200 && o2.Status == "declined")
	ok("customer emailed decline", anySent(sent, "decline@x.com", ""))

	// capacity throttle (MaxActive=2)
	a := makeRequest("a@x.com")
	b := makeRequest("b@x.com")
	d := makeRequest("d@x.com")
	ok("confirm #1 ok", postJSON("/confirm", `{"order_id":"`+a+`"}`, "testkey").StatusCode == 200)
	ok("confirm #2 ok", postJSON("/confirm", `{"order_id":"`+b+`"}`, "testkey").StatusCode == 200)
	ok("confirm #3 blocked", postJSON("/confirm", `{"order_id":"`+d+`"}`, "testkey").StatusCode == 409)
	ok("capacity now closed", jsonField(get(t, cl, srv.URL+"/capacity"), "open") == "false")

	// internal path
	r = postJSON("/internal/run", `{"domain":"ours.com","audience":"x","assets":"y"}`, "testkey")
	rep := jsonField(r, "report")
	ok("internal run returns report", r.StatusCode == 200 && strings.Contains(rep, "Stub report"))
	ok("internal run needs key", postJSON("/internal/run", `{"domain":"d","audience":"a","assets":"s"}`, "").StatusCode == 401)

	// subscribe
	ok("subscribe ok", postForm("/subscribe", url.Values{"email": {"sub@x.com"}}).StatusCode == 200)

	t.Logf("ALL %d CHECKS PASSED", pass)
}

// ── tiny test helpers ────────────────────────────────────────────────────────
func get(t *testing.T, cl *http.Client, u string) *http.Response {
	r, err := cl.Get(u)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func anySent(sent *[][3]string, to, bodyContains string) bool {
	for _, s := range *sent {
		if s[0] == to && (bodyContains == "" || strings.Contains(s[2], bodyContains)) {
			return true
		}
	}
	return false
}

func toStr(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

// jsonField2 reads the "status" field (separate reader since body is consumed once).
func jsonField2(r *http.Response) string {
	b, _ := io.ReadAll(r.Body)
	m := map[string]any{}
	json.Unmarshal(b, &m)
	return toStr(m["status"])
}

// ── /audience-check (free taster) tests ──────────────────────────────────────

func newTasterApp() *App {
	app, _ := newTestApp()
	app.audience = func(business, aud, assets string) (audienceResult, error) {
		return audienceResult{
			CarriedAudience:  "stubbed reframed audience",
			WillingnessToPay: "stubbed wtp explanation",
			Alternatives: []struct {
				Audience string `json:"audience"`
				Why      string `json:"why"`
			}{
				{Audience: "alt-1", Why: "reason-1"},
				{Audience: "alt-2", Why: "reason-2"},
			},
		}, nil
	}
	return app
}

func TestAudienceCheckRejectsGET(t *testing.T) {
	app := newTasterApp()
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	r, err := srv.Client().Get(srv.URL + "/audience-check")
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", r.StatusCode)
	}
}

func TestAudienceCheckRequiresBothFields(t *testing.T) {
	app := newTasterApp()
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	cases := []url.Values{
		{"business": {"acme.uk"}}, // missing audience
		{"audience": {"farmers"}}, // missing business
		{},                        // missing both
	}
	for i, v := range cases {
		r, _ := srv.Client().PostForm(srv.URL+"/audience-check", v)
		if r.StatusCode != http.StatusBadRequest {
			t.Fatalf("case %d: expected 400, got %d", i, r.StatusCode)
		}
	}
}

func TestAudienceCheckHappyPath(t *testing.T) {
	app := newTasterApp()
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	r, err := srv.Client().PostForm(srv.URL+"/audience-check", url.Values{
		"business": {"acme.uk"},
		"audience": {"UK farmers"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", r.StatusCode)
	}
	b, _ := io.ReadAll(r.Body)
	body := string(b)
	for _, want := range []string{
		"stubbed reframed audience",
		"stubbed wtp explanation",
		"alt-1", "alt-2", "reason-1",
		"Request the full report — £29",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q\nbody:\n%s", want, body)
		}
	}
}

func TestAudienceCheckEscapesUserInput(t *testing.T) {
	app := newTasterApp()
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	r, _ := srv.Client().PostForm(srv.URL+"/audience-check", url.Values{
		"business": {`<script>alert(1)</script>`},
		"audience": {`<b>bold</b>`},
	})
	b, _ := io.ReadAll(r.Body)
	body := string(b)
	if strings.Contains(body, "<script>") || strings.Contains(body, "<b>bold</b>") {
		t.Fatalf("user input not escaped — XSS risk\nbody:\n%s", body)
	}
}

func TestAudienceCheckRateLimit(t *testing.T) {
	app := newTasterApp()
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	v := url.Values{"business": {"acme.uk"}, "audience": {"farmers"}}
	// First 3 should pass (per-hour band cap = 3).
	for i := 0; i < 3; i++ {
		r, _ := srv.Client().PostForm(srv.URL+"/audience-check", v)
		if r.StatusCode != 200 {
			t.Fatalf("call %d: expected 200, got %d", i+1, r.StatusCode)
		}
	}
	// 4th should be rate-limited.
	r, _ := srv.Client().PostForm(srv.URL+"/audience-check", v)
	if r.StatusCode == 200 {
		b, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(b), "Have another go") {
			t.Fatalf("4th call wasn't rate-limited (200 OK with normal body)")
		}
	}
}

// TestReviewBeforePayFlow exercises REVIEW_BEFORE_PAY=true: confirm runs the
// engine and holds the draft (no pay link yet); approve sends the pay link;
// payment releases the already-generated report. No money before approval.
func TestReviewBeforePayFlow(t *testing.T) {
	app, sent := newTestApp(func(c *Config) { c.ReviewBeforePay = true })
	srv := httptest.NewServer(app.routes())
	defer srv.Close()
	cl := srv.Client()

	op := func(path, body string) *http.Response {
		req, _ := http.NewRequest("POST", srv.URL+path, strings.NewReader(body))
		req.Header.Set("content-type", "application/json")
		req.Header.Set("X-Internal-Key", "testkey")
		r, err := cl.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	if _, err := cl.PostForm(srv.URL+"/request", url.Values{"name": {"Sam"}, "email": {"buyer@x.com"},
		"business": {"acme.co.uk"}, "audience": {"aud"}, "notes": {"n"}}); err != nil {
		t.Fatal(err)
	}
	id := reqID(sent)
	if id == "" {
		t.Fatal("no order id from request")
	}

	// confirm → engine runs inline → awaiting_review, draft to operator, NO pay link to buyer
	*sent = (*sent)[:0]
	r := op("/confirm", `{"order_id":"`+id+`"}`)
	if r.StatusCode != 200 || jsonField2(r) != "running" {
		t.Fatalf("confirm: want 200/running, got %d", r.StatusCode)
	}
	if o, _ := app.store.Get(id); o.Status != "awaiting_review" {
		t.Fatalf("after confirm want awaiting_review, got %s", o.Status)
	}
	if !anySent(sent, "ops@test", "DRAFT REPORT") {
		t.Fatal("operator did not receive the draft to review")
	}
	if anySent(sent, "buyer@x.com", "") {
		t.Fatal("buyer was emailed before approval — must not happen in review-before-pay")
	}

	// approve → pay link to buyer → awaiting_payment
	*sent = (*sent)[:0]
	r = op("/approve", `{"order_id":"`+id+`"}`)
	if r.StatusCode != 200 || jsonField2(r) != "awaiting_payment" {
		t.Fatalf("approve: want 200/awaiting_payment, got %d", r.StatusCode)
	}
	if o, _ := app.store.Get(id); o.Status != "awaiting_payment" {
		t.Fatalf("after approve want awaiting_payment, got %s", o.Status)
	}
	// The email wording is "pay £<price> here:" (payLinkText) — the old bare
	// "pay here" assertion went stale when the price was put into the sentence,
	// and this test has been failing on wording, not behaviour, ever since.
	if !anySent(sent, "buyer@x.com", "To go ahead, pay") {
		t.Fatal("buyer did not get the pay link after approval")
	}

	// approving an order that isn't awaiting_review is rejected
	if rr := op("/approve", `{"order_id":"`+id+`"}`); rr.StatusCode != 409 {
		t.Fatalf("approve on non-review order: want 409, got %d", rr.StatusCode)
	}

	// pay → stored report delivered to buyer (no second engine run)
	*sent = (*sent)[:0]
	evt := `{"event_id":"evt_b1","type":"checkout.session.completed","order_id":"` + id + `","paid":true}`
	if rr := op("/stripe/webhook", evt); jsonField2(rr) != "accepted" {
		t.Fatal("webhook not accepted")
	}
	if o, _ := app.store.Get(id); o.Status != "delivered" {
		t.Fatalf("after pay want delivered, got %s", o.Status)
	}
	if !anySent(sent, "buyer@x.com", "Stub report") {
		t.Fatal("buyer did not receive the report after payment")
	}
}

// TestRenderReadable prints the plain-text report so the layout can be eyeballed,
// and guards against the old markdown clutter creeping back in.
func TestRenderReadable(t *testing.T) {
	adv := []scored{
		{
			Title: "Read scanned forms and drawings into clean data", Flag: "consider",
			Defensibility: 4, Willingness: 4, Buildability: 3, Reuse: 3, Durability: 3, Risk: 4, Sum: 17,
			candidate: candidate{
				Asset:            "the scanned floor plans, wiring diagrams and survey sheets your customers email you",
				Capability:       "an AI that reads those images and pulls the measurements and details into a tidy table",
				BeatsFreeBecause: "It reads a customer's scanned drawings and pulls out the measurements and details as clean, checked data, instead of someone retyping them by hand.",
				Findings:         "We searched and found that tools like this already exist and have got noticeably better over the past year. The part that is hard to copy is teaching it to read your own customers' drawing styles, which the off-the-shelf tools get wrong.",
			},
			CheapestTest: "Ask one customer for twenty of their old drawings, run them through a basic reader and a version set up for their style, and show them the difference in accuracy.",
		},
		{
			Title: "Draft letters in the firm's own style", Flag: "consider", ShortLived: true, NeedsLiabilityWork: true,
			Defensibility: 4, Willingness: 5, Buildability: 3, Reuse: 3, Durability: 2, Risk: 2, Sum: 17,
			candidate: candidate{
				Asset:            "the firm's past letters and contracts",
				Capability:       "an AI trained on those documents so it writes the way the firm writes",
				BeatsFreeBecause: "It drafts letters and contract clauses in the firm's own style and wording, rather than the generic phrasing a public chatbot produces.",
				Findings:         "We checked and there is a real, paying market for legal drafting tools, several of them charging a monthly fee per person. The part that is genuinely hard to copy is training it on a firm's own past work.",
			},
			CheapestTest: "Check there is demand before building anything: ask three firms whether they would pay a monthly fee per person for a tool trained on their own letters. Do not build it until you have insurance for handling personal data and a solicitor has checked the terms.",
		},
	}
	dropped := []scored{{
		Title: "Flag when a saved document is out of date", Defensibility: 2, Willingness: 4,
		candidate: candidate{BeatsFreeBecause: "It watches a firm's stored documents and warns them when one is probably out of date and needs reviewing."},
	}}
	riskDropped := []scored{{
		Title: "Give the public legal advice on their own case", Defensibility: 4, Willingness: 5, Buildability: 3, Risk: 1,
		candidate: candidate{BeatsFreeBecause: "It gives a member of the public direct advice on their own legal situation, the way a solicitor would, instead of them paying for an appointment."},
	}}

	domain := "a small firm that builds AI tools for solicitors"
	aud := "Small and mid-size law firms that have plenty of their own documents but no in-house software or AI specialists."
	wtp := "They would pay because their real problem is turning their own messy, confidential files into something useful, and they cannot safely paste client documents into a public chatbot like ChatGPT."

	// The step-0 assessment of the submitted idea — the report's headline
	// section, with the checkable sources the report page promises.
	assess := assessment{
		IsAssessable:     true,
		Reading:          "A tool that reads solicitors' scanned post and files each letter against the right case.",
		Problem:          "Firms spend paralegal hours filing scanned post by hand, and misfiled letters cause missed deadlines — several practice-management vendors describe exactly this pain on their own sites.",
		DemandEvidence:   "Multiple paid products exist for adjacent document-handling problems and law-firm forums discuss the filing burden regularly, which is evidence people pay to reduce it.",
		WhoElse:          "The big practice-management suites offer generic document upload but nothing that reads and files scanned post automatically.",
		SubstitutesToday: "A paralegal doing it by hand, or generic OCR software plus manual filing.",
		Defensible:       "Training on the firm's own filing conventions is the hard-to-copy part.",
		Exposed:          "The practice-management incumbents could add this as a feature; the firm's data lives in their systems already.",
		NextStep:         "Ask two firms to let you file one week of their post with a rough prototype and count the errors against their paralegal's.",
		Sources: []source{
			{Title: "Example practice-management vendor page", URL: "https://example.com/pm-vendor"},
			{Title: "Law-firm operations forum thread", URL: "https://example.com/forum-thread"},
		},
	}
	adv[0].Sources = []source{{Title: "Competitor comparison page", URL: "https://example.com/competitors"}}

	out := render(domain, aud, wtp, assess, adv, dropped, riskDropped, "")
	t.Logf("\n%s", out)

	// Write a standalone, viewable HTML sample of the email (dev convenience;
	// silently skipped if the outputs dir isn't present, e.g. on the build box).
	outHTML := renderHTML(domain, aud, wtp, assess, adv, dropped, riskDropped, "")
	_ = os.WriteFile("/mnt/user-data/outputs/sample_report_email.html",
		[]byte(`<!doctype html><html><head><meta charset="utf-8"><title>idea.uk sample report email</title></head><body style="margin:0">`+outHTML+`</body></html>`), 0o644)
	if !strings.Contains(outHTML, "Your idea report") || !strings.Contains(outHTML, "Further ideas worth pursuing") {
		t.Fatal("HTML report missing expected structure")
	}
	for _, want := range []string{"Your idea, assessed", "Check it yourself:", "https://example.com/pm-vendor", "https://example.com/competitors"} {
		if !strings.Contains(outHTML, want) {
			t.Fatalf("HTML report missing %q", want)
		}
	}

	for _, want := range []string{"IDEA REPORT — ", "This report is from idea.uk", "YOUR IDEA, ASSESSED",
		"A considered next step:", "Check it yourself:", "https://example.com/pm-vendor",
		"WHO IT'S FOR", "WHY THEY'D PAY", "FURTHER IDEAS WORTH PURSUING", "A cheap first test:",
		"DIDN'T MAKE THE CUT", "SET ASIDE ON RISK", "We use AI to research and draft this report"} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered report missing %q", want)
		}
	}

	// The too-early outcome must render as an honest refusal, not a padded verdict.
	early := render(domain, aud, wtp, assessment{IsAssessable: false,
		Reading: "An interest in doing something with AI for solicitors, with no specific proposition yet."},
		nil, nil, nil, "No candidate survived the cut.")
	if !strings.Contains(early, "too early to assess honestly") {
		t.Fatal("unassessable submission did not render the honest too-early outcome")
	}
	for _, bad := range []string{"### ", "**", "- **", " · "} {
		if strings.Contains(out, bad) {
			t.Fatalf("rendered report still has markdown clutter %q", bad)
		}
	}
}

func TestOperatorLink(t *testing.T) {
	app, _ := newTestApp() // ReviewBeforePay=false → confirm sends a pay link (FakeProvider)
	id := "ord_link1"
	app.store.Save(&Order{ID: id, Name: "Bo", Email: "b@x.com", Domain: "a farm in devon",
		Audience: "shops", Status: "requested"})
	tok := app.orderToken(id)

	// GET /op with a valid token shows the confirm button and changes nothing.
	rec := httptest.NewRecorder()
	app.opPage(rec, httptest.NewRequest("GET", "/op?o="+id+"&t="+tok, nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Confirm and run the report") {
		t.Fatalf("op page: code=%d, confirm button present=%v", rec.Code, strings.Contains(rec.Body.String(), "Confirm and run the report"))
	}
	if o, _ := app.store.Get(id); o.Status != "requested" {
		t.Fatalf("op GET was not a no-op: status=%s", o.Status)
	}

	// GET /op with a bad token is rejected.
	rec = httptest.NewRecorder()
	app.opPage(rec, httptest.NewRequest("GET", "/op?o="+id+"&t=wrong", nil))
	if rec.Code != 404 {
		t.Fatalf("bad-token op page: code=%d, want 404", rec.Code)
	}

	// POST /confirm with the token (no header) authorises and advances the order.
	rec = httptest.NewRecorder()
	pr := httptest.NewRequest("POST", "/confirm", strings.NewReader("order_id="+id+"&t="+tok))
	pr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	pr.Header.Set("Accept", "text/html")
	app.confirm(rec, pr)
	if rec.Code != 200 {
		t.Fatalf("token confirm: code=%d", rec.Code)
	}
	if o, _ := app.store.Get(id); o.Status != "awaiting_payment" {
		t.Fatalf("token confirm: status=%s, want awaiting_payment", o.Status)
	}

	// POST /confirm with a wrong token is rejected.
	app.store.Save(&Order{ID: "ord_link2", Email: "b@x.com", Domain: "x", Audience: "y", Status: "requested"})
	rec = httptest.NewRecorder()
	pr = httptest.NewRequest("POST", "/confirm", strings.NewReader("order_id=ord_link2&t=wrong"))
	pr.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	app.confirm(rec, pr)
	if rec.Code != 401 {
		t.Fatalf("wrong-token confirm: code=%d, want 401", rec.Code)
	}
}
