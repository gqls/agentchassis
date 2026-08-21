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
}

func (f *fakeDeliveryDeps) ConfirmTransfer(_ *gin.Context, token string) error {
	f.gotTokens = append(f.gotTokens, token)
	return f.err
}
func (f *fakeDeliveryDeps) Logger() *zap.Logger { return zap.NewNop() }

func serve(t *testing.T, deps DeliveryDeps, method, path string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewDeliveryHandler(deps)
	r.GET("/c/:token", h.HandleConfirmTransfer)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(method, path, nil))
	return w
}

func TestConfirmTransferSucceedsAndSaysSo(t *testing.T) {
	f := &fakeDeliveryDeps{}
	w := serve(t, f, http.MethodGet, "/c/abc123")

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

// The whole point of ErrTokenNotFound being ONE error is that the page cannot
// become an oracle. This pins that at the layer a stranger can actually see.
func TestConfirmTransferNeverSaysWHICHFailure(t *testing.T) {
	f := &fakeDeliveryDeps{err: delivery.ErrTokenNotFound}
	w := serve(t, f, http.MethodGet, "/c/abc123")

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
	w := serve(t, f, http.MethodGet, "/c/abc123")

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
	w := serve(t, f, http.MethodGet, "/c/"+strings.Repeat("A", maxTokenLen+1))

	if strings.Contains(w.Body.String(), "that is recorded") {
		t.Fatalf("an oversized token was confirmed; the length guard is not doing anything")
	}
	if len(f.gotTokens) != 0 {
		t.Errorf("oversized token reached the dependency: %v", f.gotTokens[0][:20])
	}
	// A token exactly AT the limit must still work, or the guard is off by one
	// and would reject real tokens the day the format changes.
	f2 := &fakeDeliveryDeps{}
	w2 := serve(t, f2, http.MethodGet, "/c/"+strings.Repeat("A", maxTokenLen))
	if !strings.Contains(w2.Body.String(), "that is recorded") {
		t.Errorf("a token at exactly maxTokenLen was refused; the guard is off by one")
	}
}

// The token is in the URL. Everything that could copy it onward is a leak.
func TestConfirmTransferDoesNotEchoTheTokenOrLeakItOnward(t *testing.T) {
	const token = "SUPERSECRETTOKENVALUE"
	f := &fakeDeliveryDeps{}
	w := serve(t, f, http.MethodGet, "/c/"+token)

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
	for _, deps := range []DeliveryDeps{
		&fakeDeliveryDeps{},
		&fakeDeliveryDeps{err: delivery.ErrTokenNotFound},
		&fakeDeliveryDeps{err: errors.New("boom")},
	} {
		body := serve(t, deps, http.MethodGet, "/c/abc123").Body.String()
		for _, external := range []string{"http://", "https://", "<script", "<img", "<link"} {
			if strings.Contains(strings.ToLower(body), external) {
				t.Errorf("page references something external (%q): %q", external, body)
			}
		}
		if strings.Contains(body, "—") {
			t.Errorf("page contains an em dash, which is banned on every surface of this site: %q", body)
		}
	}
}
