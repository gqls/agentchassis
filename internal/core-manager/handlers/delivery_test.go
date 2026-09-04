// FILE: internal/core-manager/handlers/delivery_test.go
//
// HTTP-layer tests for the customer handover links. The SQL semantics are
// covered in platform/delivery against real Postgres; what is tested here is
// everything a customer or an attacker can observe: status, copy, headers, and
// what is NOT said.
//
// Every case is written so that REMOVING the behaviour it covers makes it fail.
// The fake confirms ANY token by default, which is what makes the malformed
// -token case a real assertion rather than a vacuous one: if the length guard
// were deleted, that request would reach the fake, succeed, and render the
// success page. Asserting "the dependency was not called" would have proven
// nothing by comparison (LANDMINES: assert the mechanism's EFFECT, never the
// absence of a call).
//
// ⚠ THAT RULE IS WHY THE GET TESTS LOOK INDIRECT. "The link click mutates
// nothing" is a NEGATIVE, and a fake's own bookkeeping cannot assert one. So it
// is tested by its effect instead: the fake confirms any token, so a GET that
// still reached it would render the success copy. Asserting that copy is ABSENT
// from the page fails the moment the method split is undone, which is exactly
// the regression worth catching.
package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gqls/agentchassis/platform/delivery"
)

type fakeDeliveryDeps struct {
	err       error
	gotTokens []string

	// zip arm: what ZipDownloadURL returns, and whether a stale hit was recorded.
	zipURL        string
	zipErr        error
	staleRecorded int
}

func (f *fakeDeliveryDeps) ConfirmTransfer(_ *gin.Context, token string) error {
	f.gotTokens = append(f.gotTokens, token)
	return f.err
}
func (f *fakeDeliveryDeps) ZipDownloadURL(_ *gin.Context, token string) (string, error) {
	f.gotTokens = append(f.gotTokens, token)
	if f.zipErr != nil {
		return "", f.zipErr
	}
	if f.zipURL == "" {
		return "", delivery.ErrTokenNotFound
	}
	return f.zipURL, nil
}
func (f *fakeDeliveryDeps) RecordStaleZipLink(_ *gin.Context) { f.staleRecorded++ }
func (f *fakeDeliveryDeps) Logger() *zap.Logger               { return zap.NewNop() }

// newRouter uses the SAME registration function api/server.go calls, so the
// route table has one definition and this file cannot drift from production.
// A hand-copied table here would let a test register the POST handler on GET
// and pass every assertion below while the live service did the opposite.
func newRouter(deps DeliveryDeps) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	NewDeliveryHandler(deps).RegisterRoutes(r)
	return r
}

// newRouterWithHEAD adds a route PRODUCTION DOES NOT HAVE. gin does not route
// HEAD to a GET handler, so the handlers' HEAD refusals are unreachable through
// the real router today; they are kept for the day someone reaches for r.Any(),
// and this helper is how they are exercised without widening the live surface.
// Anything using it is testing the handler, never the deployed routing.
func newRouterWithHEAD(deps DeliveryDeps, h gin.HandlerFunc) *gin.Engine {
	r := newRouter(deps)
	r.HEAD("/c/:token", h)
	return r
}

func serve(t *testing.T, deps DeliveryDeps, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	return serveWithHeader(t, deps, method, path, "", "")
}

func serveWithHeader(t *testing.T, deps DeliveryDeps, method, path, hdr, val string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	if hdr != "" {
		req.Header.Set(hdr, val)
	}
	newRouter(deps).ServeHTTP(w, req)
	return w
}

// ── The GET page: the mail-scanner mitigation itself ─────────────────────────

// The one that matters. Undo the method split and this test goes red.
func TestGetRendersTheButtonAndConfirmsNothing(t *testing.T) {
	f := &fakeDeliveryDeps{} // confirms ANY token
	w := serve(t, f, http.MethodGet, "/c/abc123")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	body := w.Body.String()

	// THE assertion: a GET that still mutated would reach the fake, succeed and
	// render the success copy. This is what a mail scanner would have caused.
	if strings.Contains(body, "that is recorded") {
		t.Fatalf("a plain GET CONFIRMED the transfer; the second-click split is gone: %q", body)
	}
	if !strings.Contains(body, "Yes, I have moved everything") {
		t.Errorf("the page has no button: %q", body)
	}
	if !strings.Contains(body, `<form method="post"`) {
		t.Errorf("the button is not in a POST form, so pressing it cannot confirm: %q", body)
	}

	// ⚠ THE COPY ASSERTION ABOVE IS NOT ENOUGH, and finding that out is why this
	// block exists. A GET that called ConfirmTransfer and THEN rendered the
	// button page would mutate the database and leave the page identical — the
	// mutation was run (2026-08-25) and the suite stayed green until this went
	// in. This file's standing rule is to assert the effect and never the
	// absence of a call, and it does not apply here: "the link click reaches no
	// database" IS an absence, and nothing in the response can witness it. What
	// keeps this honest is that the mutation makes it fail.
	if len(f.gotTokens) != 0 {
		t.Fatalf("a plain GET reached ConfirmTransfer with %v; the link click must touch no database at all", f.gotTokens)
	}
}

// The form must submit to the current URL by having NO action attribute. With
// action="/c/<token>" the secret lands in the page body, and from there into
// anything that keeps a copy of the page.
func TestTheFormDoesNotCarryTheTokenIntoTheHTML(t *testing.T) {
	const token = "SUPERSECRETTOKENVALUE"
	f := &fakeDeliveryDeps{}
	body := serve(t, f, http.MethodGet, "/c/"+token).Body.String()

	if strings.Contains(body, token) {
		t.Errorf("the token is echoed into the button page, so it lands in history and any page cache")
	}
	if strings.Contains(body, "action=") {
		t.Errorf("the form names an action; without one it posts to the current URL and the token stays out of the HTML: %q", body)
	}
}

// The page is a fixed string. If it ever varies by token, it becomes a way to
// test a guess without pressing the button, which is the property the owner
// chose this shape for.
func TestTheButtonPageIsIdenticalForEveryToken(t *testing.T) {
	a := serve(t, &fakeDeliveryDeps{}, http.MethodGet, "/c/"+strings.Repeat("A", 43)).Body.String()
	b := serve(t, &fakeDeliveryDeps{}, http.MethodGet, "/c/"+strings.Repeat("B", 43)).Body.String()
	if a != b {
		t.Errorf("the page varies by token, which makes it a validity oracle")
	}
}

func TestGetRefusesHEADAndSpeculativeFetches(t *testing.T) {
	f := &fakeDeliveryDeps{}
	wh := httptest.NewRecorder()
	newRouterWithHEAD(f, NewDeliveryHandler(f).HandleConfirmPage).
		ServeHTTP(wh, httptest.NewRequest(http.MethodHead, "/c/abc123", nil))
	if wh.Code == http.StatusOK {
		t.Errorf("HEAD returned 200, want a refusal")
	}
	w := serveWithHeader(t, f, http.MethodGet, "/c/abc123", "Sec-Purpose", "prefetch")
	if w.Code == http.StatusOK {
		t.Errorf("a speculative GET returned 200; a 2xx is a usable prefetch result")
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}

// Both verbs must answer a hostile URL the same way, or the pair of them is the
// oracle neither is alone.
func TestGetRefusesAnOversizedTokenLikeThePostDoes(t *testing.T) {
	body := serve(t, &fakeDeliveryDeps{}, http.MethodGet, "/c/"+strings.Repeat("A", maxTokenLen+1)).Body.String()
	if strings.Contains(body, "Yes, I have moved everything") {
		t.Errorf("an oversized token was offered the button: %q", body)
	}
	if !strings.Contains(body, "no longer active") {
		t.Errorf("an oversized token got neither the button nor the failure page: %q", body)
	}
}

// ── The POST: the button press, and the state change ─────────────────────────

func TestConfirmTransferSucceedsAndSaysSo(t *testing.T) {
	f := &fakeDeliveryDeps{}
	w := serve(t, f, http.MethodPost, "/c/abc123")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), "that is recorded") {
		t.Errorf("success page missing its confirmation copy: %q", w.Body.String())
	}
	// The token must reach the dependency verbatim: a handler that trimmed,
	// lower-cased or otherwise "tidied" it would hash to a different value and
	// every real link would fail, which no amount of page-copy testing sees.
	if len(f.gotTokens) != 1 || f.gotTokens[0] != "abc123" {
		t.Errorf("dependency saw %v, want exactly [abc123]", f.gotTokens)
	}
}

// The token is single-use at the database. The second press must land on the
// same page as a stranger's guess, and say nothing about why.
func TestASecondPressSaysTheLinkIsNoLongerActive(t *testing.T) {
	f := &fakeDeliveryDeps{}
	if first := serve(t, f, http.MethodPost, "/c/abc123"); !strings.Contains(first.Body.String(), "that is recorded") {
		t.Fatalf("the first press did not confirm: %q", first.Body.String())
	}
	f.err = delivery.ErrTokenNotFound // the database has now spent it
	second := serve(t, f, http.MethodPost, "/c/abc123")
	if !strings.Contains(second.Body.String(), "no longer active") {
		t.Errorf("a second press did not get the spent-link page: %q", second.Body.String())
	}
}

// The whole point of ErrTokenNotFound being ONE error is that the page cannot
// become an oracle. This pins that at the layer a stranger can actually see.
func TestConfirmTransferNeverSaysWHICHFailure(t *testing.T) {
	f := &fakeDeliveryDeps{err: delivery.ErrTokenNotFound}
	w := serve(t, f, http.MethodPost, "/c/abc123")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (a 404 from a link we emailed reads as a lost site)", w.Code)
	}
	body := strings.ToLower(w.Body.String())
	if !strings.Contains(body, "no longer active") {
		t.Errorf("failure page missing its copy: %q", body)
	}
	for _, leak := range []string{"expired", "revoked", "already been used by", "unknown", "not found", "purpose", "site"} {
		if strings.Contains(body, leak) {
			t.Errorf("failure page names a specific failure mode (%q), which makes it an oracle: %q", leak, body)
		}
	}
}

// A database fault is ours. Telling the customer their link is invalid would
// send them chasing a problem they cannot fix, at a service that does not
// answer questions.
func TestConfirmTransferDistinguishesOurFaultFromTheirLink(t *testing.T) {
	f := &fakeDeliveryDeps{err: errors.New("connection refused")}
	w := serve(t, f, http.MethodPost, "/c/abc123")

	body := w.Body.String()
	if !strings.Contains(body, "our end") {
		t.Errorf("a database fault must not be reported as a bad link: %q", body)
	}
	if strings.Contains(body, "no longer active") {
		t.Errorf("a database fault rendered the invalid-link page: %q", body)
	}
	if strings.Contains(body, "connection refused") {
		t.Errorf("the internal error leaked to the customer: %q", body)
	}
}

// The fake succeeds on ANY token, so this fails if the length guard is removed.
func TestConfirmTransferRefusesAnOversizedTokenBeforeTheDatabase(t *testing.T) {
	f := &fakeDeliveryDeps{}
	w := serve(t, f, http.MethodPost, "/c/"+strings.Repeat("A", maxTokenLen+1))

	if strings.Contains(w.Body.String(), "that is recorded") {
		t.Fatalf("an oversized token was confirmed; the length guard is not doing anything")
	}
	if len(f.gotTokens) != 0 {
		t.Errorf("oversized token reached the dependency: %v", f.gotTokens[0][:20])
	}
	// A token exactly AT the limit must still work, or the guard is off by one
	// and would reject real tokens the day the format changes.
	f2 := &fakeDeliveryDeps{}
	w2 := serve(t, f2, http.MethodPost, "/c/"+strings.Repeat("A", maxTokenLen))
	if !strings.Contains(w2.Body.String(), "that is recorded") {
		t.Errorf("a token at exactly maxTokenLen was refused; the guard is off by one")
	}
}

// The token is in the URL. Everything that could copy it onward is a leak.
func TestConfirmTransferDoesNotEchoTheTokenOrLeakItOnward(t *testing.T) {
	const token = "SUPERSECRETTOKENVALUE"
	f := &fakeDeliveryDeps{}
	w := serve(t, f, http.MethodPost, "/c/"+token)

	if strings.Contains(w.Body.String(), token) {
		t.Errorf("the token is echoed into the page, so it lands in history and any page cache")
	}
	if got := w.Header().Get("Referrer-Policy"); got != "no-referrer" {
		t.Errorf("Referrer-Policy = %q, want no-referrer: the token is in the path", got)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store: a cached success page would be shown to the next clicker", got)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html: a human is reading this", ct)
	}
}

// The page is read on a phone, from an email, possibly offline-ish, with no CDN
// in front of it. Anything external is a way to look broken on someone else's
// bad day, and the register bans em dashes everywhere.
func TestConfirmPageIsSelfContainedAndObeysTheVoiceRules(t *testing.T) {
	type page struct {
		method string
		deps   DeliveryDeps
	}
	// The button page is in this loop deliberately: it is the one a customer is
	// most likely to see, and it was written last.
	for _, p := range []page{
		{http.MethodGet, &fakeDeliveryDeps{}},
		{http.MethodPost, &fakeDeliveryDeps{}},
		{http.MethodPost, &fakeDeliveryDeps{err: delivery.ErrTokenNotFound}},
		{http.MethodPost, &fakeDeliveryDeps{err: errors.New("boom")}},
	} {
		body := serve(t, p.deps, p.method, "/c/abc123").Body.String()
		for _, external := range []string{"http://", "https://", "<script", "<img", "<link"} {
			if strings.Contains(strings.ToLower(body), external) {
				t.Errorf("%s page references something external (%q): %q", p.method, external, body)
			}
		}
		if strings.Contains(body, "—") {
			t.Errorf("%s page contains an em dash, which is banned on every surface of this site: %q", p.method, body)
		}
	}
}

// TestNoConfirmPagePromisesRemindersStopWhileNothingSendsThem is a TRIPWIRE for
// bugs_open/477, and it is meant to be deleted one day.
//
// Both confirm pages used to end on "You will not get any more reminders about
// it." Nothing in the estate sends a reminder, so that sentence asked a customer
// to take an action and told them what the action prevented, when the thing it
// prevented did not exist. Measured 2026-09-04: one agent can send mail
// (delivery-email-sender), ZERO scheduled tasks target it, and
// sites.transfer_confirmed_at had no reader outside platform/delivery/handover.go.
//
// WHAT THIS TEST IS FOR: making the restoration deliberate. When the follow-up
// sender exists and this stamp suppresses it, the wording SHOULD come back, and
// this test should be deleted in the same commit as part of doing so.
//
// WHAT IT IS NOT: a check that no sender exists, and not a check that the page
// tells no other lie. It matches this promise's vocabulary only, so a rewording
// ("we will not chase you") would pass it. That limit is stated rather than
// papered over: the class of defect — customer-facing copy promising a mechanism
// the code does not have — is the one bugs_open/475, 476 and 477 all share, and
// nothing automated in the estate catches it.
func TestNoConfirmPagePromisesRemindersStopWhileNothingSendsThem(t *testing.T) {
	for _, p := range []struct {
		what   string
		method string
	}{
		{"the button page a customer lands on", http.MethodGet},
		{"the success page after pressing", http.MethodPost},
	} {
		// The fake confirms any token, so the POST really does render the
		// success copy: this reads the page a customer sees, not a failure page.
		body := serve(t, &fakeDeliveryDeps{}, p.method, "/c/abc123").Body.String()
		if strings.Contains(strings.ToLower(body), "reminder") {
			t.Errorf("%s promises something about reminders, and nothing in the estate sends one (bugs_open/477). "+
				"If the follow-up sender now exists, delete this test in the commit that restores the wording: %q", p.what, body)
		}
	}
}

// ── Speculative-fetch refusal ────────────────────────────────────────────────
//
// These follow this file's rule: assert the EFFECT, never that the dependency
// went uncalled. The fake confirms ANY token, so if the guard were deleted each
// of these requests would reach it, succeed, and render the success page. The
// assertion "the success copy is absent" therefore fails the moment the
// behaviour is removed, which is the whole point.
//
// Browsers do not speculatively POST, so on this verb the guard is defence
// against a future routing change rather than against today's browsers. It is
// tested because an untested guard is one a refactor deletes without noticing.

func TestConfirmTransferRefusesSpeculativeFetches(t *testing.T) {
	// Every vendor's spelling of the same idea. A browser sends whichever its
	// generation used, so missing one is missing that whole browser.
	cases := []struct{ name, hdr, val string }{
		{"sec-purpose prefetch", "Sec-Purpose", "prefetch"},
		{"sec-purpose prefetch;prerender", "Sec-Purpose", "prefetch;prerender"},
		{"sec-purpose prerender", "Sec-Purpose", "prerender"},
		{"sec-purpose mixed case", "Sec-Purpose", "PreFetch"},
		{"purpose prefetch", "Purpose", "prefetch"},
		{"x-purpose preview", "X-Purpose", "preview"},
		{"x-purpose prefetch", "X-Purpose", "prefetch"},
		{"x-moz prefetch", "X-Moz", "prefetch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &fakeDeliveryDeps{}
			w := serveWithHeader(t, f, http.MethodPost, "/c/abc123", tc.hdr, tc.val)

			// THE assertion that makes this test real: with the guard gone, the
			// fake would confirm and this copy would be present.
			if strings.Contains(w.Body.String(), "that is recorded") {
				t.Fatalf("a speculative fetch was CONFIRMED: %q", w.Body.String())
			}
			if w.Code == http.StatusOK {
				t.Errorf("status = 200; a 2xx is a usable prefetch result and may be "+
					"replayed to the customer instead of their real click (got body %q)", w.Body.String())
			}
			if got := w.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
		})
	}
}

// The POST handler's own HEAD arm is unreachable through gin's router today
// (HEAD does not fall through to a POST route), so it is driven directly. It
// stays because r.Any() is one refactor away and this is the file that must not
// depend on the router's current shape.
func TestTheConfirmHandlerItselfRefusesHEAD(t *testing.T) {
	f := &fakeDeliveryDeps{}
	w := httptest.NewRecorder()
	newRouterWithHEAD(f, NewDeliveryHandler(f).HandleConfirmTransfer).
		ServeHTTP(w, httptest.NewRequest(http.MethodHead, "/c/abc123", nil))

	if w.Code == http.StatusOK {
		t.Errorf("HEAD returned 200, want a refusal")
	}
	if strings.Contains(w.Body.String(), "that is recorded") {
		t.Errorf("HEAD confirmed a transfer")
	}
}

// The must-pass arm. Without it, a guard that refused EVERYTHING would pass
// every test above, and the whole confirm mechanism would be dead with the
// suite green.
func TestConfirmTransferStillSucceedsForAnOrdinaryButtonPress(t *testing.T) {
	f := &fakeDeliveryDeps{}
	w := serveWithHeader(t, f, http.MethodPost, "/c/abc123", "User-Agent", "Mozilla/5.0")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 for a normal press", w.Code)
	}
	if !strings.Contains(w.Body.String(), "that is recorded") {
		t.Fatalf("a normal press was not confirmed: %q", w.Body.String())
	}
	if len(f.gotTokens) != 1 || f.gotTokens[0] != "abc123" {
		t.Errorf("token not passed through: %v", f.gotTokens)
	}
}

// A header that merely MENTIONS a signal word must not trip the guard: the
// check is on the four purpose headers, not on any header containing
// "prefetch". Without this, tightening the guard into a substring search over
// all headers would look correct.
func TestConfirmTransferIgnoresUnrelatedHeaders(t *testing.T) {
	for _, hdr := range []string{"User-Agent", "Referer", "Accept"} {
		f := &fakeDeliveryDeps{}
		w := serveWithHeader(t, f, http.MethodPost, "/c/abc123", hdr, "prefetch")
		if !strings.Contains(w.Body.String(), "that is recorded") {
			t.Errorf("%s: an ordinary press was refused because an unrelated header said %q", hdr, "prefetch")
		}
	}
}

// validToken is a 43-char token-shaped path segment, the shape the box regex
// admits and mintTokenPlaintext produces.
var validToken = strings.Repeat("a", 43)

// --- /d/<token>: the ZIP download ---------------------------------------------

// A fresh stored URL 302s; the Location is the stored presign and nothing else.
func TestZipDownloadRedirectsToTheStoredURL(t *testing.T) {
	f := &fakeDeliveryDeps{zipURL: "https://bucket.example/zip?sig=abc"}
	rec := serve(t, f, http.MethodGet, "/d/"+validToken)
	if rec.Code != http.StatusFound {
		t.Fatalf("fresh zip link returned %d, want 302", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "https://bucket.example/zip?sig=abc" {
		t.Errorf("Location = %q, want the stored presign", got)
	}
	if f.staleRecorded != 0 {
		t.Errorf("a fresh hit recorded staleness %d times", f.staleRecorded)
	}
}

// A STALE stored URL must never be redirected to: an expired presign answers
// 403 SignatureDoesNotMatch, which reads as broken credentials. The customer
// gets the honest refresh page, and the staleness is RECORDED so it becomes a
// row somebody sees rather than the customer's private dead-end.
func TestZipDownloadStaleRendersHonestPageAndRecords(t *testing.T) {
	f := &fakeDeliveryDeps{zipErr: delivery.ErrZipURLStale}
	rec := serve(t, f, http.MethodGet, "/d/"+validToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("stale zip link returned %d, want 200 (the refresh page)", rec.Code)
	}
	if rec.Header().Get("Location") != "" {
		t.Fatal("stale zip link REDIRECTED: the target would 403 as SignatureDoesNotMatch")
	}
	body := rec.Body.String()
	if !strings.Contains(body, "being refreshed") {
		t.Errorf("stale page does not say the link is being refreshed: %q", body[:min(len(body), 200)])
	}
	if f.staleRecorded != 1 {
		t.Errorf("staleness recorded %d times, want exactly 1", f.staleRecorded)
	}
}

// Unknown token: the uniform failure page, one message for every cause, and no
// staleness row (nothing real happened).
func TestZipDownloadUnknownTokenGetsTheUniformFailurePage(t *testing.T) {
	f := &fakeDeliveryDeps{zipErr: delivery.ErrTokenNotFound}
	rec := serve(t, f, http.MethodGet, "/d/"+validToken)
	if rec.Code != http.StatusOK {
		t.Fatalf("unknown zip token returned %d, want 200 (the failure page)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "no longer active") {
		t.Error("unknown-token page does not carry the uniform failure copy")
	}
	if f.staleRecorded != 0 {
		t.Errorf("an unknown token recorded staleness %d times", f.staleRecorded)
	}
}

// POST /d/ must not exist: the route registers GET only, so gin 404s it before
// any handler runs. This is the shape the box's limit_except GET mirrors.
func TestZipDownloadRefusesPost(t *testing.T) {
	f := &fakeDeliveryDeps{zipURL: "https://bucket.example/zip"}
	rec := serve(t, f, http.MethodPost, "/d/"+validToken)
	if rec.Code == http.StatusFound || rec.Code == http.StatusOK {
		t.Fatalf("POST /d/ returned %d; the route must be GET-only", rec.Code)
	}
	if len(f.gotTokens) != 0 {
		t.Errorf("POST /d/ reached the deps: %v", f.gotTokens)
	}
}
