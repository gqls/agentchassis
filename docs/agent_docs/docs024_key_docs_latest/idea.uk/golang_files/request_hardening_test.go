package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Exercises the /request spam defences added 2026-07-16: honeypot, timing gate,
// email validation, and IP capture. Silent-drop paths (honeypot, too-fast) must
// return the ordinary success page while saving and emailing nothing.
func TestRequestHardening(t *testing.T) {
	post := func(app *App, form url.Values, xff string) *http.Response {
		srv := httptest.NewServer(app.routes())
		t.Cleanup(srv.Close)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/request",
			strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if xff != "" {
			req.Header.Set("X-Forwarded-For", xff)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("post /request: %v", err)
		}
		return resp
	}
	good := func() url.Values {
		return url.Values{
			"name": {"Ada"}, "email": {"ada@example.com"},
			"business": {"acme.co.uk"}, "audience": {"builders"},
			"notes": {""}, "_elapsed": {"6000"}, "accept_terms": {"yes"},
		}
	}

	t.Run("happy path saves order with IP and emails operator", func(t *testing.T) {
		app, sent := newTestApp()
		resp := post(app, good(), "203.0.113.7, 10.0.0.1")
		if resp.StatusCode != 200 {
			t.Fatalf("want 200, got %d", resp.StatusCode)
		}
		if len(app.store.Orders) != 1 {
			t.Fatalf("want 1 order saved, got %d", len(app.store.Orders))
		}
		var o *Order
		for _, v := range app.store.Orders {
			o = v
		}
		// CORRECTED 2026-07-26 (bugs_open/090). This assertion used to read
		// `want IP 203.0.113.7 (first XFF entry)` — it asserted the defect. The
		// first XFF entry is whatever the CALLER wrote; nginx appends the real
		// peer, so the trustworthy value is the LAST one. Under the old rule any
		// visitor could pick the address recorded against their order and the key
		// their rate limit was counted under. See client_ip_test.go.
		if o.IP != "10.0.0.1" {
			t.Errorf("want IP 10.0.0.1 (the peer nginx appended, not the caller's claim), got %q", o.IP)
		}
		if len(*sent) != 1 {
			t.Errorf("want 1 operator email, got %d", len(*sent))
		}
	})

	t.Run("honeypot: filled company_url is silently dropped", func(t *testing.T) {
		app, sent := newTestApp()
		f := good()
		f.Set("company_url", "http://spam.example")
		resp := post(app, f, "203.0.113.8")
		if resp.StatusCode != 200 {
			t.Fatalf("honeypot should still return 200 success, got %d", resp.StatusCode)
		}
		if len(app.store.Orders) != 0 {
			t.Errorf("honeypot submission must not be saved, got %d orders", len(app.store.Orders))
		}
		if len(*sent) != 0 {
			t.Errorf("honeypot submission must not email, got %d", len(*sent))
		}
	})

	t.Run("timing: too-fast submit is silently dropped", func(t *testing.T) {
		app, sent := newTestApp()
		f := good()
		f.Set("_elapsed", "100") // 100ms < 2500ms floor
		resp := post(app, f, "203.0.113.9")
		if resp.StatusCode != 200 {
			t.Fatalf("too-fast should still return 200 success, got %d", resp.StatusCode)
		}
		if len(app.store.Orders) != 0 || len(*sent) != 0 {
			t.Errorf("too-fast submission must not be saved/emailed; orders=%d sent=%d",
				len(app.store.Orders), len(*sent))
		}
	})

	t.Run("timing fails open when _elapsed absent (no-JS visitor)", func(t *testing.T) {
		app, _ := newTestApp()
		f := good()
		f.Del("_elapsed")
		resp := post(app, f, "203.0.113.10")
		if resp.StatusCode != 200 || len(app.store.Orders) != 1 {
			t.Errorf("no-JS visitor must be accepted; status=%d orders=%d",
				resp.StatusCode, len(app.store.Orders))
		}
	})

	t.Run("invalid email is rejected", func(t *testing.T) {
		app, _ := newTestApp()
		f := good()
		f.Set("email", "not-an-email")
		resp := post(app, f, "203.0.113.11")
		if resp.StatusCode != 400 {
			t.Errorf("want 400 for bad email, got %d", resp.StatusCode)
		}
		if len(app.store.Orders) != 0 {
			t.Errorf("bad email must not be saved")
		}
	})

	t.Run("GET is rejected", func(t *testing.T) {
		app, _ := newTestApp()
		srv := httptest.NewServer(app.routes())
		t.Cleanup(srv.Close)
		resp, _ := http.Get(srv.URL + "/request")
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("want 405 for GET, got %d", resp.StatusCode)
		}
	})
}

func TestSubjectSnippet(t *testing.T) {
	t.Run("short text unchanged", func(t *testing.T) {
		if got := subjectSnippet("acme widgets ltd"); got != "acme widgets ltd" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("long text cut at a word boundary with ellipsis", func(t *testing.T) {
		long := "agent framework that creates and maintains websites and that can create, test and fix clientside javascript tools on websites"
		got := subjectSnippet(long)
		if !strings.HasSuffix(got, "…") {
			t.Errorf("want … suffix, got %q", got)
		}
		if n := len([]rune(got)); n > 61 {
			t.Errorf("want ≤61 runes, got %d: %q", n, got)
		}
		if strings.Contains(got, "javascript") {
			t.Errorf("should have been cut well before the tail: %q", got)
		}
	})

	t.Run("whitespace and newlines collapsed", func(t *testing.T) {
		if got := subjectSnippet("two\n\nlines\t here"); got != "two lines here" {
			t.Errorf("got %q", got)
		}
	})
}
