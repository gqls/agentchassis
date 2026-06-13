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
	if !anySent(sent, "buyer@x.com", "pay here") {
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
			Title: "Drawing/Form Reader to Structured Data", Flag: "consider",
			Defensibility: 4, Willingness: 4, Buildability: 3, Reuse: 3, Durability: 3, Risk: 4, Sum: 18,
			candidate: candidate{
				Asset:            "the customer's own engineering drawings",
				Capability:       "tuned vision reading with schema-checked output",
				BeatsFreeBecause: "reads a customer's drawings and returns validated, structured fields",
				Findings:         "Checked that tools like this exist and have matured in the last year; the edge is tuning on the customer's own drawing conventions, which generic tools miss.",
			},
			CheapestTest: "Ask one engineering prospect for 20 anonymised drawings, run a plain pass against a 5-example tuned pass, and show them the accuracy gap.",
		},
		{
			Title: "Firm-Voice Drafting Model", Flag: "consider", ShortLived: true, NeedsLiabilityWork: true,
			Defensibility: 4, Willingness: 5, Buildability: 3, Reuse: 3, Durability: 2, Risk: 2, Sum: 17,
			candidate: candidate{
				Asset:            "the firm's past letters and contracts",
				Capability:       "a private model tuned on that house style",
				BeatsFreeBecause: "drafts in the firm's actual style and clause preferences, not a generic template",
				Findings:         "A paying market for legal drafting tools is real (several priced per user per month). A true tune on a firm's own past work is the defensible part.",
			},
			CheapestTest: "Validate demand first; do not build until PII insurance is in force and T&Cs are reviewed by a UK solicitor. Then ask three firms whether they'd pay a monthly per-seat fee for a model trained on their own drafts.",
		},
	}
	dropped := []scored{{Title: "Stale-Doc Drift Detector", Defensibility: 2, Willingness: 4}}
	riskDropped := []scored{{Title: "Symptom Triage Assistant", Defensibility: 4, Willingness: 5, Buildability: 3}}

	domain := "a small legal-AI services firm"
	aud := "Regional law firms with their own documents but no ML staff"
	wtp := "They pay because the bottleneck is turning their own messy, sensitive documents into something usable, which they can't safely paste into a public chatbot."

	out := render(domain, aud, wtp, adv, dropped, riskDropped, "")
	t.Logf("\n%s", out)

	// Write a standalone, viewable HTML sample of the email (dev convenience;
	// silently skipped if the outputs dir isn't present, e.g. on the build box).
	outHTML := renderHTML(domain, aud, wtp, adv, dropped, riskDropped, "")
	_ = os.WriteFile("/mnt/user-data/outputs/sample_report_email.html",
		[]byte(`<!doctype html><html><head><meta charset="utf-8"><title>idea.uk sample report email</title></head><body style="margin:0;background:#EFE7D6">`+outHTML+`</body></html>`), 0o644)
	if !strings.Contains(outHTML, "Idea report") || !strings.Contains(outHTML, "Advancing ideas") {
		t.Fatal("HTML report missing expected structure")
	}

	for _, want := range []string{"IDEA REPORT — ", "WHO IT'S FOR", "ADVANCING IDEAS", "First test:", "DIDN'T MAKE THE CUT", "SET ASIDE ON RISK"} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered report missing %q", want)
		}
	}
	for _, bad := range []string{"### ", "**", "- **", " · "} {
		if strings.Contains(out, bad) {
			t.Fatalf("rendered report still has markdown clutter %q", bad)
		}
	}
}
