package main

// error_page_test.go — the public /request form's error pages.
//
// Reported by the owner 2026-07-28: "they typed too much into the text box and it
// showed an error page but that error page wasn't designed." Reproduced live —
// http.Error returns text/plain, so the visitor landed on a bare unstyled line.
//
// The form is a NATIVE POST (its JS only stamps the timing field), so a rejection
// navigates the browser away from the form. That is what makes an unstyled error
// expensive rather than merely ugly: it reads as "your work is gone", and it
// happens to whoever wrote the most.

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func postForm(t *testing.T, app *App, form url.Values) (*http.Response, string) {
	t.Helper()
	srv := httptest.NewServer(app.routes())
	t.Cleanup(srv.Close)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/request", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("post /request: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func validForm() url.Values {
	return url.Values{
		"name": {"Ada"}, "email": {"ada@example.com"},
		"business": {"a tool for surveyors"}, "audience": {"UK surveyors"},
		"notes": {""}, "_elapsed": {"6000"},
	}
}

// The defect itself: an error must be a styled HTML page, never text/plain.
func TestRequestErrorsAreStyledHTMLNotPlainText(t *testing.T) {
	cases := map[string]func(url.Values){
		"over-length notes": func(f url.Values) { f.Set("notes", strings.Repeat("a", 4200)) },
		"missing field":     func(f url.Values) { f.Set("business", "") },
		"bad email":         func(f url.Values) { f.Set("email", "not-an-address") },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			app, _ := newTestApp()
			f := validForm()
			mutate(f)
			resp, body := postForm(t, app, f)

			if resp.StatusCode != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", resp.StatusCode)
			}
			if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
				t.Errorf("content-type = %q, want text/html — this IS the reported bug", ct)
			}
			if !strings.Contains(body, "<!doctype html>") {
				t.Errorf("no doctype: the page furniture is missing, so it renders unstyled:\n%.200s", body)
			}
			// A route back that keeps what they wrote.
			if !strings.Contains(body, "history.back()") || !strings.Contains(body, "#request-a-report") {
				t.Errorf("no back-to-the-filled-in-form control:\n%.400s", body)
			}
		})
	}
}

// "One of the fields is too long" made the visitor guess. Name the field, both
// numbers, and the shortfall.
func TestOverLengthErrorNamesTheFieldAndTheNumbers(t *testing.T) {
	app, _ := newTestApp()
	f := validForm()
	f.Set("notes", strings.Repeat("a", 5200))
	_, body := postForm(t, app, f)

	for _, want := range []string{"extra notes", "5,200", "4,000", "1,200"} {
		if !strings.Contains(body, want) {
			t.Errorf("error page does not mention %q:\n%.500s", want, body)
		}
	}
}

// Counting by rune, not byte. A byte count silently gives a shorter allowance to
// anyone writing with accents or non-Latin script, and this direction only ever
// accepts MORE, so it cannot newly reject a real person.
func TestLengthLimitsCountRunesNotBytes(t *testing.T) {
	// 2,000 three-byte runes = 6,000 bytes: within the 2,000-rune audience limit,
	// but three times over it if counted as bytes.
	audience := strings.Repeat("測", 2000)
	if got := overLongField("Ada", "an idea", audience, ""); got != "" {
		t.Errorf("2,000 runes rejected against a 2,000 limit — counting bytes, not runes: %q", got)
	}
	if got := overLongField("Ada", "an idea", audience+"測", ""); got == "" {
		t.Error("2,001 runes accepted against a 2,000 limit — the cap is not being applied")
	}
}

// Nothing over the limit must produce no message at all, or every valid
// submission would be rejected.
func TestOverLongFieldSilentWhenWithinLimits(t *testing.T) {
	if got := overLongField("Ada", "an idea", "surveyors", strings.Repeat("a", 4000)); got != "" {
		t.Errorf("a submission exactly at the limit was rejected: %q", got)
	}
}

// The rejected submission must not create an order — the styling fix must not
// have quietly moved the validation.
func TestRejectedSubmissionCreatesNoOrder(t *testing.T) {
	app, _ := newTestApp()
	f := validForm()
	f.Set("notes", strings.Repeat("a", 4200))
	postForm(t, app, f)
	if n := len(app.store.Orders); n != 0 {
		t.Errorf("a rejected submission created %d order(s)", n)
	}
}

// The address is echoed back so the visitor can see the typo — escaped, and
// bounded so a pasted wall of text cannot make the error page itself unreadable.
func TestBadEmailIsEchoedSafely(t *testing.T) {
	app, _ := newTestApp()
	f := validForm()
	f.Set("email", "<script>alert(1)</script>@nope")
	_, body := postForm(t, app, f)
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Error("echoed address was not escaped")
	}
	if !strings.Contains(body, "&lt;script&gt;") {
		t.Errorf("address not echoed back at all — the visitor cannot see the typo:\n%.300s", body)
	}
}
