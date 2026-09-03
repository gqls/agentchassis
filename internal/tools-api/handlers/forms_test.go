package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/store"
)

// recordingInbox is the seam substitute: it records what it was asked to store
// so a test can assert that nothing was stored, which is the assertion most of
// this file turns on.
type recordingInbox struct {
	inserted  []store.InboxRow
	insertErr error
}

func (r *recordingInbox) Insert(_ context.Context, row store.InboxRow) (string, error) {
	if r.insertErr != nil {
		return "", r.insertErr
	}
	r.inserted = append(r.inserted, row)
	return "11111111-1111-1111-1111-111111111111", nil
}

func (r *recordingInbox) ClaimPending(_ context.Context, limit int, _ time.Time) ([]store.InboxRow, error) {
	return nil, nil
}

const (
	testSiteID     = "00ff3af5-dad8-4770-9f70-3edc267a3c92"
	testSiteDomain = "robot-hands.com"
	goodToken      = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
)

// newFormsRouter mounts POST /submit with the site already resolved, standing in
// for CORSMiddleware. The forms group's own logic must not depend on how those
// two values got there — which is precisely what makes them unsafe as identity.
func newFormsRouter(st FormInboxStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("site_id", testSiteID)
		c.Set("site_domain", testSiteDomain)
	})
	r.POST("/submit", FormSubmitHandler(st))
	return r
}

func postForm(t *testing.T, r *gin.Engine, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// A filled honeypot must produce the SAME outcome a person gets and store
// nothing. The "store nothing" half is the one worth asserting: a gate that
// answers correctly and files the row anyway has not gated anything.
func TestFormSubmitHoneypotStoresNothing(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)

	w := postForm(t, r, "_token="+goodToken+"&_intent=enquiry&name=Bot&"+fieldHoneypot+"=http://spam.example")

	if len(st.inserted) != 0 {
		t.Fatalf("honeypot submission was STORED (%d rows) — the bot gate did not gate", len(st.inserted))
	}
	if w.Code != http.StatusSeeOther && w.Code != http.StatusCreated {
		t.Fatalf("honeypot got status %d; a bot must get the same answer a person gets", w.Code)
	}
}

// The positive control for the test above: the identical request WITHOUT the
// honeypot must be stored. Without this, TestFormSubmitHoneypotStoresNothing
// passes just as well against a handler that stores nothing at all, ever.
func TestFormSubmitStoresAGenuineSubmission(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)

	w := postForm(t, r, "_token="+goodToken+"&_intent=enquiry&name=Ada&brief=needs+a+landing+page")

	if len(st.inserted) != 1 {
		t.Fatalf("genuine submission stored %d rows, want 1 (status %d)", len(st.inserted), w.Code)
	}
	got := st.inserted[0]
	if got.Token != goodToken {
		t.Errorf("token stored as %q, want it VERBATIM as presented: %q", got.Token, goodToken)
	}
	if got.Intent != "enquiry" {
		t.Errorf("intent = %q, want enquiry", got.Intent)
	}
	var payload map[string]string
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("payload is not a JSON object: %v", err)
	}
	if payload["name"] != "Ada" || payload["brief"] != "needs a landing page" {
		t.Errorf("payload lost the visitor's answers: %v", payload)
	}
	for _, reserved := range []string{fieldToken, fieldIntent, fieldNext, fieldElapsed, fieldHoneypot} {
		if _, present := payload[reserved]; present {
			t.Errorf("reserved field %q leaked into the payload", reserved)
		}
	}
}

// A token that cannot be one of ours is refused and nothing is stored. This is
// shape only — the island has nothing to check authenticity against, which is
// the design.
func TestFormSubmitRefusesAMalformedToken(t *testing.T) {
	for _, tc := range []struct{ name, token string }{
		{"absent", ""},
		{"too short", "abc"},
		{"31 chars, one under the floor", strings.Repeat("a", 31)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st := &recordingInbox{}
			r := newFormsRouter(st)
			w := postForm(t, r, "_token="+tc.token+"&_intent=enquiry&name=Ada")
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", w.Code)
			}
			if len(st.inserted) != 0 {
				t.Errorf("a malformed token was STORED")
			}
		})
	}
}

// The intent must match the CHECK constraint both tables carry, so a bad value
// is a 400 the site's author can act on rather than a 500 from Postgres.
func TestFormSubmitRefusesAMalformedIntent(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)
	w := postForm(t, r, "_token="+goodToken+"&_intent=Removal-Request&name=Ada")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 for an intent that violates the table CHECK", w.Code)
	}
	if len(st.inserted) != 0 {
		t.Errorf("a malformed intent was STORED and would have failed at the constraint")
	}
}

// An absent intent defaults, so the commonest form (one purpose, no _intent
// field) works without the author knowing the concept exists.
func TestFormSubmitDefaultsTheIntent(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)
	postForm(t, r, "_token="+goodToken+"&name=Ada")
	if len(st.inserted) != 1 || st.inserted[0].Intent != "enquiry" {
		t.Fatalf("absent intent did not default to enquiry: %+v", st.inserted)
	}
}

// safeRedirect must build the host from the RESOLVED site domain and never from
// the request. Every value below must be refused.
//
// ⚠ ACCURACY NOTE, because the obvious framing of this test is wrong and was
// written that way first. These are NOT all open redirects today: because the
// host is always prefixed, "//evil.example" produces
// "https://robot-hands.com//evil.example" — the real host, with evil.example as
// a path segment. The mutation run that removed the "//" guard proved exactly
// that, and the claim was corrected rather than left standing.
//
// What this test actually pins is narrower and still worth having: a value that
// names another origin is refused outright instead of being silently mangled
// into a doubled path, and the function keeps failing safe if someone later
// uses `next` as a whole URL rather than as a path — the one-line change that
// WOULD make these open redirects.
func TestSafeRedirectRefusesAnythingOffSite(t *testing.T) {
	for _, next := range []string{
		"//evil.example",
		"//evil.example/thanks",
		"https://evil.example",
		"http://evil.example/thanks",
		"\\\\evil.example",
		"/thanks\\@evil.example",
		"javascript:alert(1)",
		"/thanks\nLocation: https://evil.example",
		"thanks",
	} {
		if got, ok := safeRedirect(testSiteDomain, next); ok {
			t.Errorf("safeRedirect(%q) = %q, accepted — a value naming another origin must be refused", next, got)
		}
	}
}

// The positive control: legitimate paths must still work, or the test above is
// satisfied by a function that refuses everything.
func TestSafeRedirectAcceptsAnOnSitePath(t *testing.T) {
	for _, tc := range []struct{ next, want string }{
		{"", "https://" + testSiteDomain + "/"},
		{"/", "https://" + testSiteDomain + "/"},
		{"/thank-you.html", "https://" + testSiteDomain + "/thank-you.html"},
		{"/thanks/?ref=form", "https://" + testSiteDomain + "/thanks/?ref=form"},
	} {
		got, ok := safeRedirect(testSiteDomain, tc.next)
		if !ok || got != tc.want {
			t.Errorf("safeRedirect(%q) = %q,%v; want %q,true", tc.next, got, ok, tc.want)
		}
	}
}

// A plain form post gets a 303 to the site's own domain, so the visitor sees a
// page rather than a JSON body — and it works with JavaScript off.
func TestFormSubmitRedirectsAPlainFormPost(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)
	w := postForm(t, r, "_token="+goodToken+"&_next=%2Fthank-you.html&name=Ada")

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303 for a form-encoded post", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "https://"+testSiteDomain+"/thank-you.html" {
		t.Errorf("Location = %q, want the resolved site's own domain", loc)
	}
}

// A hostile _next must not stop the submission being STORED. Losing a real
// enquiry because someone tampered with the redirect would be the wrong
// trade — the fallback is the JSON body, not a refusal.
func TestFormSubmitStoresEvenWhenTheRedirectIsRefused(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)
	w := postForm(t, r, "_token="+goodToken+"&_next=https%3A%2F%2Fevil.example&name=Ada")

	if len(st.inserted) != 1 {
		t.Fatalf("submission was dropped because its _next was hostile; stored %d", len(st.inserted))
	}
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201 (the safe fallback), and certainly not a redirect", w.Code)
	}
	if loc := w.Header().Get("Location"); loc != "" {
		t.Errorf("Location = %q, want none", loc)
	}
}

// A JSON post gets the flat success body, byte-identical to the one a bot gets.
func TestFormSubmitJSONPathReturnsTheFlatBody(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)
	req := httptest.NewRequest(http.MethodPost, "/submit",
		strings.NewReader(`{"_token":"`+goodToken+`","_intent":"removal_request","name":"Ada"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if w.Body.String() != string(formAcceptedBody) {
		t.Errorf("body = %q, want the flat accepted body", w.Body.String())
	}
	if len(st.inserted) != 1 || st.inserted[0].Intent != "removal_request" {
		t.Fatalf("JSON submission not stored with its intent: %+v", st.inserted)
	}
}

// site_id is recorded, and recorded as a CROSS-CHECK. The assertion that matters
// is that it comes from the context and the token comes from the body: they are
// independent, so the collector can compare them.
func TestFormSubmitRecordsOriginSiteSeparatelyFromTheToken(t *testing.T) {
	st := &recordingInbox{}
	r := newFormsRouter(st)
	postForm(t, r, "_token="+goodToken+"&name=Ada")

	if len(st.inserted) != 1 {
		t.Fatal("nothing stored")
	}
	got := st.inserted[0]
	if got.SiteID == nil || *got.SiteID != testSiteID {
		t.Errorf("site_id = %v, want the Origin-resolved id recorded for cross-checking", got.SiteID)
	}
	if got.SiteDomain != testSiteDomain {
		t.Errorf("site_domain = %q, want %q", got.SiteDomain, testSiteDomain)
	}
	if got.Token != goodToken {
		t.Errorf("token = %q; it must be the body's, not derived from the resolved site", got.Token)
	}
}

// The payload's shape is bounded independently of the body cap, so 32 KB cannot
// arrive as thousands of tiny fields.
func TestPayloadFromBoundsFieldCountAndValueLength(t *testing.T) {
	fields := map[string]string{"_token": goodToken}
	for i := 0; i < maxPayloadFields*3; i++ {
		fields["f"+string(rune('a'+i%26))+string(rune('a'+i/26))] = "x"
	}
	fields["long"] = strings.Repeat("y", maxPayloadValueLen+500)

	got := payloadFrom(fields)
	if len(got) > maxPayloadFields {
		t.Errorf("payload has %d fields, cap is %d", len(got), maxPayloadFields)
	}
	if v, ok := got["long"]; ok && len(v) > maxPayloadValueLen {
		t.Errorf("value length %d exceeds cap %d", len(v), maxPayloadValueLen)
	}
}
