package main

// service_test.go — end-to-end state-machine test. FakeProvider (no Stripe) +
// stubbed engine (no LLM spend). Run: go test ./...

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestApp() (*App, *[][3]string) {
	cfg := Config{
		PriceGBP: 199, AutoDeliver: true, PublicBaseURL: "http://test",
		InternalAPIKey: "testkey", OperatorEmail: "ops@test", MaxActive: 2,
		AllowedOrigins: []string{"*"},
	}
	store, _ := NewStore("") // in-memory
	app := NewApp(cfg, store, &FakeProvider{publicBaseURL: "http://test"})
	app.engine = func(d, a, s string) (string, error) { return "# Stub report for " + d, nil }
	sent := &[][3]string{}
	app.deliver = func(to, subj, body string) { *sent = append(*sent, [3]string{to, subj, body}) }
	app.dispatch = func(f func()) { f() } // run fulfilment inline for deterministic asserts
	return app, sent
}

func reqID(sent *[][3]string) string {
	for i := len(*sent) - 1; i >= 0; i-- {
		if s := (*sent)[i][1]; strings.Contains(s, "NEW REQUEST ") {
			return strings.SplitN(strings.SplitN(s, "NEW REQUEST ", 2)[1], " ", 2)[0]
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
