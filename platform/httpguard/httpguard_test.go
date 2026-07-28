package httpguard

import (
	"net/http"
	"testing"
	"time"
)

func req(remoteAddr string, headers map[string]string) *http.Request {
	r, _ := http.NewRequest(http.MethodPost, "/submit", nil)
	r.RemoteAddr = remoteAddr
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

// THE regression test for bugs_open/090. nginx appends the real peer with
// $proxy_add_x_forwarded_for, so a client-supplied header arrives on the LEFT and
// the truth on the RIGHT. Trusting the first entry let a caller pick their own
// rate-limit bucket, proven in production.
func TestClientIPIgnoresAForgedForwardedForPrefix(t *testing.T) {
	got := ClientIP(req("127.0.0.1:5000", map[string]string{
		"X-Forwarded-For": "203.0.113.77, 198.51.100.9",
	}))
	if got == "203.0.113.77" {
		t.Fatal("trusted the client-supplied FIRST X-Forwarded-For entry — this is bugs_open/090")
	}
	if got != "198.51.100.9" {
		t.Fatalf("want the rightmost (proxy-appended) entry 198.51.100.9, got %q", got)
	}
}

func TestClientIPPrefersXRealIPFromOurProxy(t *testing.T) {
	got := ClientIP(req("10.0.0.5:41000", map[string]string{
		"X-Real-IP":       "198.51.100.9",
		"X-Forwarded-For": "203.0.113.77",
	}))
	if got != "198.51.100.9" {
		t.Fatalf("want X-Real-IP 198.51.100.9, got %q", got)
	}
}

// A direct caller's headers are user input. If the peer is not plausibly our own
// proxy we must key on the real socket address and ignore what it claims.
func TestClientIPIgnoresHeadersFromAPublicPeer(t *testing.T) {
	got := ClientIP(req("198.51.100.9:33000", map[string]string{
		"X-Real-IP":       "203.0.113.77",
		"X-Forwarded-For": "203.0.113.77",
	}))
	if got != "198.51.100.9" {
		t.Fatalf("headers from a public peer must be ignored, got %q", got)
	}
}

func TestClientIPFallsBackToPeerWithNoHeaders(t *testing.T) {
	if got := ClientIP(req("127.0.0.1:5000", nil)); got != "127.0.0.1" {
		t.Fatalf("want 127.0.0.1, got %q", got)
	}
	// A single-entry XFF from our proxy is the whole chain.
	if got := ClientIP(req("127.0.0.1:5000", map[string]string{"X-Forwarded-For": "198.51.100.9"})); got != "198.51.100.9" {
		t.Fatalf("want 198.51.100.9, got %q", got)
	}
}

func TestLimiterEnforcesEachBandAndReportsRetryAfter(t *testing.T) {
	base := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	now := base
	l := NewLimiter(Band{Window: time.Hour, Max: 3}, Band{Window: 24 * time.Hour, Max: 5})
	l.now = func() time.Time { return now }

	for i := 0; i < 3; i++ {
		if ok, _ := l.Allow("ip"); !ok {
			t.Fatalf("event %d should be allowed inside the hour band", i+1)
		}
	}
	ok, retry := l.Allow("ip")
	if ok {
		t.Fatal("fourth event in the hour must be refused")
	}
	if retry <= 0 || retry > time.Hour {
		t.Fatalf("retryAfter %v is not inside the hour band", retry)
	}

	// Past the hour window the hour band clears, but the DAY band still binds:
	// 3 spent, ceiling 5, so exactly 2 more get through and the next does not.
	now = base.Add(61 * time.Minute)
	for i := 0; i < 2; i++ {
		if ok, _ := l.Allow("ip"); !ok {
			t.Fatalf("event %d after the hour window should be allowed", i+1)
		}
	}
	if ok, _ := l.Allow("ip"); ok {
		t.Fatal("daily band must bind once its ceiling is reached")
	}
}

// A refused event must not be recorded, or being throttled would extend the
// throttle every time the caller retried.
func TestLimiterDoesNotRecordRefusedEvents(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	l := NewLimiter(Band{Window: time.Hour, Max: 1})
	l.now = func() time.Time { return now }

	l.Allow("ip")
	_, first := l.Allow("ip")
	now = now.Add(10 * time.Minute)
	_, second := l.Allow("ip")
	if second >= first {
		t.Fatalf("retryAfter should shrink as the window ages (%v then %v)", first, second)
	}
}

func TestLimiterKeysAreIndependent(t *testing.T) {
	l := NewLimiter(Band{Window: time.Hour, Max: 1})
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("first event for a")
	}
	if ok, _ := l.Allow("b"); !ok {
		t.Fatal("key b must not inherit key a's usage")
	}
	if ok, _ := l.Allow("a"); ok {
		t.Fatal("key a is spent")
	}
}

func TestLimiterSweepDropsOnlyStaleKeys(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	l := NewLimiter(Band{Window: time.Hour, Max: 10})
	l.now = func() time.Time { return now }
	l.Allow("old")
	now = now.Add(90 * time.Minute)
	l.Allow("fresh")
	if dropped := l.Sweep(); dropped != 1 {
		t.Fatalf("want 1 stale key dropped, got %d", dropped)
	}
	if ok, _ := l.Allow("fresh"); !ok {
		t.Fatal("the fresh key must survive a sweep")
	}
}

func TestNewLimiterRefusesNoBands(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("a limiter with no bands permits everything and must panic at construction")
		}
	}()
	NewLimiter()
}

func TestCheckIntakeCatchesHoneypotAndSpeed(t *testing.T) {
	if v := CheckIntake("i-am-a-bot", "9000", DefaultMinFill); !v.Bot || v.Reason != "honeypot" {
		t.Fatalf("a filled honeypot is a bot, got %+v", v)
	}
	if v := CheckIntake("", "120", DefaultMinFill); !v.Bot || v.Reason != "too-fast" {
		t.Fatalf("120ms is faster than any human, got %+v", v)
	}
	if v := CheckIntake("", "9000", DefaultMinFill); v.Bot {
		t.Fatalf("a 9s fill is a human, got %+v", v)
	}
}

// Fail open is the deliberate asymmetry: turning away a real customer costs more
// than letting one bot reach the limiter behind this.
func TestCheckIntakeFailsOpenWithoutUsableTiming(t *testing.T) {
	for _, elapsed := range []string{"", "   ", "not-a-number", "-5"} {
		if v := CheckIntake("", elapsed, DefaultMinFill); v.Bot {
			t.Fatalf("elapsed %q must fail open, got %+v", elapsed, v)
		}
	}
}
