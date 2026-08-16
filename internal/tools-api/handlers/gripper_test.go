package handlers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/gripper"
	"github.com/gqls/agentchassis/internal/tools-api/store"
)

// fakeGripperStore is an in-memory GripperStore. It mirrors the guards the
// SQL applies (active/turn caps/one submit per session) so the handler tests
// see the same refusals production would.
type fakeGripperStore struct {
	sessions   map[string]*store.Session
	requests   []fakeReq
	daily      int
	dailyErr   error
	pulled     []string
	nextID     int
	failCreate error
}

type fakeReq struct {
	ID, SiteID, Email string
	SessionID         *string
	Spec              gripper.Spec
	CreatedAt         time.Time
}

func newFakeGripperStore() *fakeGripperStore {
	return &fakeGripperStore{sessions: map[string]*store.Session{}}
}

const testSite = "00ff3af5-dad8-4770-9f70-3edc267a3c92"
const testSession = "11111111-2222-3333-4444-555555555555"

func (f *fakeGripperStore) CreateSession(_ context.Context, siteID, ipHash, ua string) (string, error) {
	if f.failCreate != nil {
		return "", f.failCreate
	}
	f.sessions[testSession] = &store.Session{ID: testSession, SiteID: siteID, Spec: gripper.Spec{}, Status: "active"}
	return testSession, nil
}

func (f *fakeGripperStore) ClaimTurn(_ context.Context, id, siteID string) (*store.Session, error) {
	s, ok := f.sessions[id]
	if !ok || s.SiteID != siteID {
		return nil, store.ErrSessionNotFound
	}
	if s.Status != "active" {
		return nil, store.ErrSessionClosed
	}
	if s.Turns >= gripper.MaxTurns || s.InputTokens+s.OutputTokens >= gripper.MaxSessionTokens {
		return nil, store.ErrSessionCapped
	}
	s.Turns++
	cp := *s
	return &cp, nil
}

func (f *fakeGripperStore) ClaimDailyTurn(_ context.Context, _ time.Time, cap int) error {
	if f.dailyErr != nil {
		return f.dailyErr
	}
	if f.daily >= cap {
		return store.ErrDailyCapReached
	}
	f.daily++
	return nil
}

func (f *fakeGripperStore) RecordTurn(_ context.Context, id, visitor, reply string, turnSpec gripper.Spec, in, out int) (gripper.Spec, error) {
	s := f.sessions[id]
	s.Transcript = append(s.Transcript, gripper.Turn{Role: "visitor", Text: visitor}, gripper.Turn{Role: "assistant", Text: reply})
	s.Spec = gripper.Merge(s.Spec, turnSpec)
	s.InputTokens += in
	s.OutputTokens += out
	return s.Spec, nil
}

func (f *fakeGripperStore) CreateRequestFromSession(_ context.Context, siteID, sessionID, email, ipHash, ua string) (string, error) {
	s, ok := f.sessions[sessionID]
	if !ok || s.SiteID != siteID {
		return "", store.ErrSessionNotFound
	}
	if s.Status != "active" {
		return "", store.ErrSessionClosed
	}
	if !gripper.Complete(s.Spec) {
		return "", store.ErrSpecIncomplete
	}
	s.Status = "submitted"
	return f.add(siteID, &sessionID, email, s.Spec), nil
}

func (f *fakeGripperStore) CreateRequestInline(_ context.Context, siteID, email string, spec gripper.Spec, ipHash, ua string) (string, error) {
	if !gripper.Complete(spec) {
		return "", store.ErrSpecIncomplete
	}
	return f.add(siteID, nil, email, spec), nil
}

func (f *fakeGripperStore) add(siteID string, sess *string, email string, spec gripper.Spec) string {
	f.nextID++
	id := "req-" + string(rune('0'+f.nextID))
	f.requests = append(f.requests, fakeReq{ID: id, SiteID: siteID, SessionID: sess, Email: email, Spec: spec,
		CreatedAt: time.Date(2026, 8, 16, 12, 0, f.nextID, 0, time.UTC)})
	return id
}

func (f *fakeGripperStore) PendingSince(_ context.Context, since *time.Time, limit int) ([]store.FeedRow, error) {
	var out []store.FeedRow
	for _, r := range f.requests {
		if since != nil && r.CreatedAt.Before(*since) {
			continue
		}
		out = append(out, store.FeedRow{ID: r.ID, Host: "robot-hands.com", CreatedAt: r.CreatedAt, Spec: gripper.ForCluster(r.Spec)})
	}
	return out, nil
}

func (f *fakeGripperStore) MarkPulled(_ context.Context, ids []string, _ time.Time) error {
	f.pulled = append(f.pulled, ids...)
	return nil
}

// fakeGen returns a canned model reply (or error).
type fakeGen struct {
	text string
	err  error
	last string // last prompt seen
}

func (g *fakeGen) GenerateText(_ context.Context, prompt string, opts map[string]interface{}) (string, error) {
	g.last = prompt
	if g.err != nil {
		return "", g.err
	}
	opts["__usage_input_tokens"] = 100
	opts["__usage_output_tokens"] = 20
	return g.text, nil
}

func gripperRouter(st GripperStore, gen TextGenerator) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Stand in for CORSMiddleware: set the site the way it would.
	r.Use(func(c *gin.Context) { c.Set("site_id", testSite); c.Set("site_domain", "robot-hands.com") })
	cfg := &config.GripperConfig{DailyTurnCap: 3, MaxBodyBytes: 16384}
	r.POST("/session", GripperSessionHandler(st))
	r.POST("/chat", GripperChatHandler(st, cfg, func(context.Context) (TextGenerator, error) { return gen, nil }))
	r.POST("/submit", GripperSubmitHandler(st))
	r.GET("/requests", GripperRequestsHandler(st))
	return r
}

func post(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(rec, req)
	return rec
}

func completeSpecJSON() string {
	b, _ := json.Marshal(map[string]interface{}{
		"mass_kg": 2.5, "part_geometry": "cylinder 60mm", "travel_mm": 60, "surface_material": "steel",
		"cycle_rate": 12, "mounting": "UR5e",
	})
	return string(b)
}

// ── /session ─────────────────────────────────────────────────────────────────

func TestSessionReturnsIdAndGreeting(t *testing.T) {
	st := newFakeGripperStore()
	rec := post(gripperRouter(st, nil), "/session", "")
	if rec.Code != 200 {
		t.Fatalf("code %d body %s", rec.Code, rec.Body.String())
	}
	var out map[string]string
	_ = json.Unmarshal(rec.Body.Bytes(), &out)
	if out["session_id"] != testSession || out["greeting"] != gripper.Greeting {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

// ── /chat ────────────────────────────────────────────────────────────────────

func TestChatHappyPathMergesSpecServerSide(t *testing.T) {
	st := newFakeGripperStore()
	st.sessions[testSession] = &store.Session{ID: testSession, SiteID: testSite, Status: "active",
		Spec: gripper.Spec{"mass_kg": 2.5}}
	gen := &fakeGen{text: `{"reply":"Got it. What is the part made of?","spec":{"mass_kg":null,"travel_mm":60,"part_geometry":"cylinder","surface_material":null,"ip_min":null,"cycle_rate":null,"mounting":null,"application":null,"budget":null},"complete":true}`}
	rec := post(gripperRouter(st, gen), "/chat", `{"session_id":"`+testSession+`","message":"about 60 mm across, a cylinder"}`)
	if rec.Code != 200 {
		t.Fatalf("code %d body %s", rec.Code, rec.Body.String())
	}
	var out struct {
		Reply    string                 `json:"reply"`
		Spec     map[string]interface{} `json:"spec"`
		Missing  []string               `json:"missing_fields"`
		Complete bool                   `json:"complete"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	// mass_kg was nulled by the model this turn but must survive (merge).
	if out.Spec["mass_kg"] != 2.5 || out.Spec["travel_mm"] != 60.0 {
		t.Errorf("spec = %#v", out.Spec)
	}
	// The model said complete:true; the server disagrees and its answer wins.
	if out.Complete || len(out.Missing) == 0 {
		t.Errorf("complete=%v missing=%v — the model's flag must not be trusted", out.Complete, out.Missing)
	}
	if !strings.Contains(gen.last, `"mass_kg":2.5`) {
		t.Errorf("prompt did not carry the stored spec")
	}
	if st.sessions[testSession].InputTokens != 100 || st.sessions[testSession].OutputTokens != 20 {
		t.Errorf("usage not recorded: %+v", st.sessions[testSession])
	}
	if len(st.sessions[testSession].Transcript) != 2 {
		t.Errorf("transcript = %+v", st.sessions[testSession].Transcript)
	}
}

func TestChatRefusalsAreHonest(t *testing.T) {
	st := newFakeGripperStore()
	st.sessions[testSession] = &store.Session{ID: testSession, SiteID: testSite, Status: "active"}
	gen := &fakeGen{text: `{"reply":"x","spec":{},"complete":false}`}
	r := gripperRouter(st, gen)

	cases := []struct {
		name string
		body string
		prep func()
		want int
	}{
		{"bad json", `{`, nil, 400},
		{"bad id", `{"session_id":"nope","message":"hi"}`, nil, 400},
		{"blank message", `{"session_id":"` + testSession + `","message":"   "}`, nil, 400},
		{"unknown session", `{"session_id":"99999999-2222-3333-4444-555555555555","message":"hi"}`, nil, 404},
		{"closed session", `{"session_id":"` + testSession + `","message":"hi"}`,
			func() { st.sessions[testSession].Status = "submitted" }, 409},
		{"capped session", `{"session_id":"` + testSession + `","message":"hi"}`,
			func() { st.sessions[testSession].Status = "active"; st.sessions[testSession].Turns = gripper.MaxTurns }, 409},
		{"daily cap", `{"session_id":"` + testSession + `","message":"hi"}`,
			func() { st.sessions[testSession].Turns = 0; st.daily = 3 }, 409},
	}
	for _, tc := range cases {
		if tc.prep != nil {
			tc.prep()
		}
		if rec := post(r, "/chat", tc.body); rec.Code != tc.want {
			t.Errorf("%s: code %d want %d body %s", tc.name, rec.Code, tc.want, rec.Body.String())
		}
	}
}

func TestChatModelFailureAndBadReplyAre503NotPersisted(t *testing.T) {
	st := newFakeGripperStore()
	st.sessions[testSession] = &store.Session{ID: testSession, SiteID: testSite, Status: "active"}
	body := `{"session_id":"` + testSession + `","message":"hi"}`

	if rec := post(gripperRouter(st, &fakeGen{err: errors.New("429")}), "/chat", body); rec.Code != 503 {
		t.Errorf("model error: code %d", rec.Code)
	}
	if rec := post(gripperRouter(st, &fakeGen{text: `Sure! Here you go: {"reply":"x"}`}), "/chat", body); rec.Code != 503 {
		t.Errorf("bad reply: code %d", rec.Code)
	}
	if len(st.sessions[testSession].Transcript) != 0 {
		t.Errorf("a failed turn was persisted: %+v", st.sessions[testSession].Transcript)
	}
}

// ── /submit ──────────────────────────────────────────────────────────────────

func TestSubmitBotAndHumanGetByteIdenticalResponses(t *testing.T) {
	st := newFakeGripperStore()
	st.sessions[testSession] = &store.Session{ID: testSession, SiteID: testSite, Status: "active",
		Spec: gripper.Normalise(mustMap(completeSpecJSON()))}
	r := gripperRouter(st, nil)

	human := post(r, "/submit", `{"session_id":"`+testSession+`","email":"v@example.org","company_website":"","_elapsed":9000}`)
	if human.Code != 201 {
		t.Fatalf("human: code %d body %s", human.Code, human.Body.String())
	}
	if len(st.requests) != 1 || st.requests[0].Email != "v@example.org" {
		t.Fatalf("human request not filed: %+v", st.requests)
	}

	// Honeypot filled → dropped, same bytes back, nothing stored.
	bot := post(r, "/submit", `{"session_id":"`+testSession+`","email":"b@example.org","company_website":"http://spam","_elapsed":9000}`)
	if bot.Code != human.Code || !bytes.Equal(bot.Body.Bytes(), human.Body.Bytes()) {
		t.Fatalf("bot response differs: %d %q vs %d %q", bot.Code, bot.Body.String(), human.Code, human.Body.String())
	}
	// Too fast → dropped too, even with an INVALID email (the gate runs first
	// so a bot cannot learn which validation it tripped).
	fast := post(r, "/submit", `{"email":"not-an-email","company_website":"","_elapsed":"300"}`)
	if fast.Code != human.Code || !bytes.Equal(fast.Body.Bytes(), human.Body.Bytes()) {
		t.Fatalf("too-fast response differs: %d %q", fast.Code, fast.Body.String())
	}
	if len(st.requests) != 1 {
		t.Fatalf("a bot submission was stored: %+v", st.requests)
	}
	// Body carries no request id.
	if strings.Contains(human.Body.String(), "req-") {
		t.Errorf("response leaks the request id: %s", human.Body.String())
	}
}

func TestSubmitValidationAndSessionRules(t *testing.T) {
	st := newFakeGripperStore()
	st.sessions[testSession] = &store.Session{ID: testSession, SiteID: testSite, Status: "active",
		Spec: gripper.Spec{"mass_kg": 1.0}} // incomplete
	r := gripperRouter(st, nil)

	if rec := post(r, "/submit", `{"session_id":"`+testSession+`","email":"nope","_elapsed":9000}`); rec.Code != 400 {
		t.Errorf("bad email: %d", rec.Code)
	}
	if rec := post(r, "/submit", `{"session_id":"`+testSession+`","email":"v@example.org","_elapsed":9000}`); rec.Code != 400 {
		t.Errorf("incomplete session spec: %d body %s", rec.Code, rec.Body.String())
	}
	if rec := post(r, "/submit", `{"session_id":"99999999-2222-3333-4444-555555555555","email":"v@example.org","_elapsed":9000}`); rec.Code != 404 {
		t.Errorf("unknown session: %d", rec.Code)
	}
	// Plain-form mode with an incomplete inline spec names what is missing.
	rec := post(r, "/submit", `{"email":"v@example.org","_elapsed":9000,"spec":{"mass_kg":2}}`)
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), "travel_mm") {
		t.Errorf("inline incomplete: %d %s", rec.Code, rec.Body.String())
	}
	// Plain-form mode with a complete inline spec files a request.
	rec = post(r, "/submit", `{"email":"v@example.org","_elapsed":9000,"spec":`+completeSpecJSON()+`}`)
	if rec.Code != 201 || len(st.requests) != 1 || st.requests[0].SessionID != nil {
		t.Errorf("inline complete: %d %s reqs=%+v", rec.Code, rec.Body.String(), st.requests)
	}
	// Submitting a session twice: second is 409.
	st.sessions[testSession].Spec = gripper.Normalise(mustMap(completeSpecJSON()))
	if rec := post(r, "/submit", `{"session_id":"`+testSession+`","email":"v@example.org","_elapsed":9000}`); rec.Code != 201 {
		t.Fatalf("first submit: %d", rec.Code)
	}
	if rec := post(r, "/submit", `{"session_id":"`+testSession+`","email":"v@example.org","_elapsed":9000}`); rec.Code != 409 {
		t.Errorf("second submit: %d", rec.Code)
	}
}

// ── /requests ────────────────────────────────────────────────────────────────

func TestRequestsFeedMatchesTheClustersParser(t *testing.T) {
	st := newFakeGripperStore()
	spec := gripper.Normalise(mustMap(completeSpecJSON()))
	st.add(testSite, nil, "a@example.org", spec)
	st.add(testSite, nil, "b@example.org", spec)
	r := gripperRouter(st, nil)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/requests", nil))
	if rec.Code != 200 || !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/x-ndjson") {
		t.Fatalf("code %d ct %s", rec.Code, rec.Header().Get("Content-Type"))
	}

	// Parse exactly as report_request_pull_action.go does.
	type pulled struct {
		ID          string                 `json:"id"`
		Host        string                 `json:"host"`
		SubmittedAt string                 `json:"submitted_at"`
		Spec        map[string]interface{} `json:"spec"`
		Meta        map[string]interface{} `json:"_meta,omitempty"`
	}
	var data, meta int
	sc := bufio.NewScanner(bytes.NewReader(rec.Body.Bytes()))
	for sc.Scan() {
		var p pulled
		if err := json.Unmarshal(sc.Bytes(), &p); err != nil {
			t.Fatalf("line unparseable: %s", sc.Text())
		}
		if p.Meta != nil {
			meta++
			continue
		}
		if p.ID == "" || p.Spec == nil {
			t.Fatalf("cluster would skip line: %s", sc.Text())
		}
		data++
		if p.Host != "robot-hands.com" {
			t.Errorf("host = %q", p.Host)
		}
		if _, err := time.Parse(time.RFC3339, p.SubmittedAt); err != nil || !strings.HasSuffix(p.SubmittedAt, "Z") {
			t.Errorf("submitted_at %q is not RFC3339 UTC", p.SubmittedAt)
		}
		if p.Spec["mass_kg"] != 2.5 || p.Spec["surface_material"] != "steel" {
			t.Errorf("spec = %#v", p.Spec)
		}
		for _, k := range []string{"email", "request_id", "submitted_at"} {
			if _, ok := p.Spec[k]; ok {
				t.Errorf("spec carries %s", k)
			}
		}
	}
	if data != 2 || meta != 1 {
		t.Fatalf("data=%d meta=%d body=%s", data, meta, rec.Body.String())
	}
	if len(st.pulled) != 2 {
		t.Errorf("MarkPulled ids = %v", st.pulled)
	}
	// Body must not contain an email anywhere.
	if strings.Contains(rec.Body.String(), "example.org") {
		t.Errorf("feed leaks email: %s", rec.Body.String())
	}
}

func TestRequestsSinceFilterAndValidation(t *testing.T) {
	st := newFakeGripperStore()
	spec := gripper.Normalise(mustMap(completeSpecJSON()))
	st.add(testSite, nil, "a@example.org", spec) // 12:00:01
	st.add(testSite, nil, "b@example.org", spec) // 12:00:02
	r := gripperRouter(st, nil)

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/requests?since=2026-08-16T12:00:02Z", nil))
	if got := strings.Count(rec.Body.String(), `"host"`); got != 1 {
		t.Errorf("since filter served %d rows, want 1: %s", got, rec.Body.String())
	}
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/requests?since=yesterday", nil))
	if rec.Code != 400 {
		t.Errorf("bad since: %d", rec.Code)
	}
	// Empty feed still terminates with _meta.
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest("GET", "/requests?since=2030-01-01T00:00:00Z", nil))
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"_meta"`) || strings.Contains(rec.Body.String(), `"host"`) {
		t.Errorf("empty feed: %d %s", rec.Code, rec.Body.String())
	}
}

func mustMap(s string) map[string]interface{} {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		panic(err)
	}
	return m
}

var _ http.Handler = (*gin.Engine)(nil)
