package gripper

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gqls/agentchassis/platform/mailer"
)

// fakeStore is an in-memory RequestStore that applies the SAME guards the SQL
// does (status = expected), so the tests exercise the poller's use of them.
type fakeStore struct {
	mu       sync.Mutex
	rows     map[string]*fakeRow
	expired  int64
	scrubbed int64
}

type fakeRow struct {
	Request
	NextCheck        time.Time
	FailureNotified  bool
	FulfilledAt      *time.Time
	EmailedAt        *time.Time
	rescheduledCount int
}

func newFakeStore(rows ...*fakeRow) *fakeStore {
	f := &fakeStore{rows: map[string]*fakeRow{}}
	for _, r := range rows {
		f.rows[r.ID] = r
	}
	return f
}

func (f *fakeStore) list(pred func(*fakeRow) bool) []Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []Request
	for _, r := range f.rows {
		if pred(r) {
			out = append(out, r.Request)
		}
	}
	return out
}

func in(s string, set ...string) bool {
	for _, x := range set {
		if s == x {
			return true
		}
	}
	return false
}

func (f *fakeStore) DueChecks(_ context.Context, now time.Time, _ int) ([]Request, error) {
	return f.list(func(r *fakeRow) bool {
		return in(r.Status, StatusPending, StatusPulled) && !r.NextCheck.After(now)
	}), nil
}
func (f *fakeStore) move(id string, from []string, apply func(*fakeRow)) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	r, ok := f.rows[id]
	if !ok || !in(r.Status, from...) {
		return false, nil
	}
	apply(r)
	return true, nil
}
func (f *fakeStore) MarkFulfilled(_ context.Context, id, u string, now time.Time) (bool, error) {
	return f.move(id, []string{StatusPending, StatusPulled}, func(r *fakeRow) {
		r.Status, r.ReportURL, r.FulfilledAt, r.NextCheck = StatusFulfilled, u, &now, now
	})
}
func (f *fakeStore) MarkFailed(_ context.Context, id string, now time.Time) (bool, error) {
	return f.move(id, []string{StatusPending, StatusPulled}, func(r *fakeRow) { r.Status, r.NextCheck = StatusFailed, now })
}
func (f *fakeStore) MarkExpired(_ context.Context, id string, now time.Time) (bool, error) {
	return f.move(id, []string{StatusPending, StatusPulled}, func(r *fakeRow) { r.Status, r.NextCheck = StatusExpired, now })
}
func (f *fakeStore) RescheduleCheck(_ context.Context, id string, next time.Time) error {
	_, err := f.move(id, []string{StatusPending, StatusPulled}, func(r *fakeRow) { r.NextCheck = next; r.rescheduledCount++ })
	return err
}
func (f *fakeStore) DueLinkEmails(_ context.Context, now time.Time, max, _ int) ([]Request, error) {
	return f.list(func(r *fakeRow) bool {
		return r.Status == StatusFulfilled && r.EmailAttempts < max && !r.NextCheck.After(now)
	}), nil
}
func (f *fakeStore) DueApologies(_ context.Context, now time.Time, max, _ int) ([]Request, error) {
	return f.list(func(r *fakeRow) bool {
		return in(r.Status, StatusFailed, StatusExpired) && !r.FailureNotified && r.EmailAttempts < max && !r.NextCheck.After(now)
	}), nil
}
func (f *fakeStore) ClaimEmailAttempt(_ context.Context, id string, expect []string, retryAt time.Time) (bool, error) {
	return f.move(id, expect, func(r *fakeRow) { r.EmailAttempts++; r.NextCheck = retryAt })
}
func (f *fakeStore) MarkEmailed(_ context.Context, id string, now time.Time) (bool, error) {
	return f.move(id, []string{StatusFulfilled}, func(r *fakeRow) { r.Status, r.EmailedAt = StatusEmailed, &now })
}
func (f *fakeStore) MarkEmailFailed(_ context.Context, id string) (bool, error) {
	return f.move(id, []string{StatusFulfilled}, func(r *fakeRow) { r.Status = StatusEmailFailed })
}
func (f *fakeStore) MarkFailureNotified(_ context.Context, id string, _ time.Time) (bool, error) {
	return f.move(id, []string{StatusFailed, StatusExpired}, func(r *fakeRow) { r.FailureNotified = true })
}
func (f *fakeStore) ExpireIdleSessions(context.Context, time.Time) (int64, error) {
	f.expired++
	return 0, nil
}
func (f *fakeStore) ScrubTerminalPII(context.Context, time.Time) (int64, error) {
	f.scrubbed++
	return 0, nil
}

// recorder is a mailer.Sender that records or refuses.
type recorder struct {
	mu   sync.Mutex
	sent []mailer.Message
	fail error
}

func (r *recorder) Send(_ context.Context, m mailer.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.fail != nil {
		return r.fail
	}
	r.sent = append(r.sent, m)
	return nil
}

// sidecarHost serves https://<domain>/reports/<id>.json from a map, and the
// poller's HTTP client is pointed at it for every host.
func sidecarHost(t *testing.T, files map[string]string) (*http.Client, func()) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if body, ok := files[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(body))
			return
		}
		http.NotFound(w, r)
	}))
	u, _ := url.Parse(srv.URL)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test server
		Proxy:           nil,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, u.Host)
		},
	}
	return &http.Client{Transport: tr, Timeout: 5 * time.Second}, srv.Close
}

func newPoller(st RequestStore, s mailer.Sender, hc *http.Client, now time.Time) *Poller {
	return &Poller{Store: st, Sender: s, HTTP: hc, Now: func() time.Time { return now }, Log: nil}
}

func baseRow(id string, created time.Time, status string) *fakeRow {
	return &fakeRow{Request: Request{
		ID: id, SiteDomain: "robot-hands.com", Email: "v@example.org", Status: status,
		CreatedAt: created, ExpiresAt: created.Add(RequestTTL),
	}, NextCheck: created.Add(FirstCheckAfter)}
}

func TestPollerReadySidecarLeadsToOneEmailAndEmailedStatus(t *testing.T) {
	created := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	now := created.Add(3 * time.Minute)
	hc, closeSrv := sidecarHost(t, map[string]string{
		"/reports/req-1.json": `{"status":"ready","generated_at":"2026-08-16T10:02:00Z","url":"/reports/req-1.html"}`,
	})
	defer closeSrv()
	st := newFakeStore(baseRow("req-1", created, StatusPulled))
	rec := &recorder{}
	p := newPoller(st, rec, hc, now)

	p.RunOnce(context.Background())

	r := st.rows["req-1"]
	if r.Status != StatusEmailed {
		t.Fatalf("status = %s, want emailed", r.Status)
	}
	if r.ReportURL != "https://robot-hands.com/reports/req-1.html" {
		t.Errorf("report url = %q", r.ReportURL)
	}
	if len(rec.sent) != 1 || rec.sent[0].To[0] != "v@example.org" {
		t.Fatalf("sent = %+v", rec.sent)
	}
	if !strings.Contains(rec.sent[0].Text, r.ReportURL) {
		t.Errorf("email does not carry the link: %q", rec.sent[0].Text)
	}
	if r.EmailAttempts != 1 {
		t.Errorf("email_attempts = %d, want 1 (claimed before send)", r.EmailAttempts)
	}
	// A second tick must not send again.
	p.RunOnce(context.Background())
	if len(rec.sent) != 1 {
		t.Fatalf("second tick re-sent: %d", len(rec.sent))
	}
}

func TestPollerNotYetReadyReschedulesOnCadence(t *testing.T) {
	created := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	hc, closeSrv := sidecarHost(t, map[string]string{})
	defer closeSrv()
	st := newFakeStore(baseRow("req-2", created, StatusPending))

	early := created.Add(3 * time.Minute)
	newPoller(st, &recorder{}, hc, early).RunOnce(context.Background())
	if got := st.rows["req-2"].NextCheck; !got.Equal(early.Add(EarlyCheckEvery)) {
		t.Errorf("early reschedule = %s, want +5m", got)
	}
	late := created.Add(2 * time.Hour)
	st.rows["req-2"].NextCheck = late
	newPoller(st, &recorder{}, hc, late).RunOnce(context.Background())
	if got := st.rows["req-2"].NextCheck; !got.Equal(late.Add(LateCheckEvery)) {
		t.Errorf("late reschedule = %s, want +15m", got)
	}
	if st.rows["req-2"].Status != StatusPending {
		t.Errorf("status moved to %s on a 404", st.rows["req-2"].Status)
	}
}

func TestPollerFailedSidecarSendsApologyOnce(t *testing.T) {
	created := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	now := created.Add(10 * time.Minute)
	hc, closeSrv := sidecarHost(t, map[string]string{
		"/reports/req-3.json": `{"status":"failed","generated_at":"2026-08-16T10:05:00Z"}`,
	})
	defer closeSrv()
	st := newFakeStore(baseRow("req-3", created, StatusPulled))
	rec := &recorder{}
	p := newPoller(st, rec, hc, now)
	p.RunOnce(context.Background())
	p.RunOnce(context.Background())
	r := st.rows["req-3"]
	if r.Status != StatusFailed || !r.FailureNotified {
		t.Fatalf("status=%s notified=%v", r.Status, r.FailureNotified)
	}
	if len(rec.sent) != 1 || rec.sent[0].Subject != ApologyMessage("x").Subject {
		t.Fatalf("sent = %+v", rec.sent)
	}
}

func TestPollerExpiresAfterTTLAndApologises(t *testing.T) {
	created := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	hc, closeSrv := sidecarHost(t, map[string]string{})
	defer closeSrv()
	row := baseRow("req-4", created, StatusPulled)
	st := newFakeStore(row)
	rec := &recorder{}
	// Just before expiry: still pending, rescheduled.
	before := created.Add(RequestTTL - time.Minute)
	row.NextCheck = before
	newPoller(st, rec, hc, before).RunOnce(context.Background())
	if row.Status != StatusPulled {
		t.Fatalf("expired early: %s", row.Status)
	}
	// At expiry: expired + apology.
	at := created.Add(RequestTTL)
	row.NextCheck = at
	newPoller(st, rec, hc, at).RunOnce(context.Background())
	if row.Status != StatusExpired || !row.FailureNotified || len(rec.sent) != 1 {
		t.Fatalf("status=%s notified=%v sent=%d", row.Status, row.FailureNotified, len(rec.sent))
	}
}

func TestPollerLinkEmailRetriesThenGivesUp(t *testing.T) {
	created := time.Date(2026, 8, 16, 10, 0, 0, 0, time.UTC)
	hc, closeSrv := sidecarHost(t, map[string]string{})
	defer closeSrv()
	row := baseRow("req-5", created, StatusFulfilled)
	row.ReportURL = DefaultReportLink("robot-hands.com", "req-5")
	st := newFakeStore(row)
	rec := &recorder{fail: errors.New("smtp down")}

	now := created.Add(5 * time.Minute)
	row.NextCheck = now
	for i := 1; i <= MaxEmailAttempts; i++ {
		newPoller(st, rec, hc, now).RunOnce(context.Background())
		if row.EmailAttempts != i {
			t.Fatalf("after tick %d attempts=%d", i, row.EmailAttempts)
		}
		// Not due again until the retry delay has passed.
		newPoller(st, rec, hc, now.Add(time.Minute)).RunOnce(context.Background())
		if row.EmailAttempts != i {
			t.Fatalf("retried before EmailRetryAfter: attempts=%d", row.EmailAttempts)
		}
		now = now.Add(EmailRetryAfter)
	}
	if row.Status != StatusEmailFailed {
		t.Fatalf("status = %s, want email_failed after %d attempts", row.Status, MaxEmailAttempts)
	}
	// SMTP recovers: nothing more is sent for a given-up row.
	rec.fail = nil
	newPoller(st, rec, hc, now.Add(time.Hour)).RunOnce(context.Background())
	if len(rec.sent) != 0 {
		t.Fatalf("email_failed row was emailed after recovery")
	}
}

func TestPollerRunOnceSkipsOverlap(t *testing.T) {
	st := newFakeStore()
	p := newPoller(st, &recorder{}, &http.Client{}, time.Now())
	p.running.Store(true) // simulate an in-flight tick
	p.RunOnce(context.Background())
	if st.expired != 0 {
		t.Fatal("overlapping tick ran maintenance")
	}
	p.running.Store(false)
	p.RunOnce(context.Background())
	if st.expired != 1 || st.scrubbed != 1 {
		t.Fatalf("first tick did not run maintenance: expired=%d scrubbed=%d", st.expired, st.scrubbed)
	}
	// Second tick within the hour: no maintenance again.
	p.RunOnce(context.Background())
	if st.expired != 1 {
		t.Fatalf("maintenance ran twice within an hour")
	}
}

func TestReportLinkHonoursSameHostSidecarURLOnly(t *testing.T) {
	p := &Poller{}
	r := Request{ID: "abc", SiteDomain: "robot-hands.com"}
	cases := map[string]string{
		"":                  "https://robot-hands.com/reports/abc.html",
		"/reports/abc.html": "https://robot-hands.com/reports/abc.html",
		"https://robot-hands.com/reports/other.html": "https://robot-hands.com/reports/other.html",
		"https://evil.example/reports/abc.html":      "https://robot-hands.com/reports/abc.html",
		"http://robot-hands.com/reports/abc.html":    "https://robot-hands.com/reports/abc.html",
		"javascript:alert(1)":                        "https://robot-hands.com/reports/abc.html",
	}
	for in, want := range cases {
		if got := p.reportLink(r, in); got != want {
			t.Errorf("reportLink(%q) = %q, want %q", in, got, want)
		}
	}
}
