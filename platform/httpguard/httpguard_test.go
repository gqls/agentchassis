package httpguard

import (
	"net"
	"net/http"
	"net/http/httptest"
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

// THE regression test for bugs_closed/090. nginx appends the real peer with
// $proxy_add_x_forwarded_for, so a client-supplied header arrives on the LEFT and
// the truth on the RIGHT. Trusting the first entry let a caller pick their own
// rate-limit bucket, proven in production.
func TestClientIPIgnoresAForgedForwardedForPrefix(t *testing.T) {
	got := ClientIP(req("127.0.0.1:5000", map[string]string{
		"X-Forwarded-For": "203.0.113.77, 198.51.100.9",
	}), Nginx())
	if got == "203.0.113.77" {
		t.Fatal("trusted the client-supplied FIRST X-Forwarded-For entry — this is bugs_closed/090")
	}
	if got != "198.51.100.9" {
		t.Fatalf("want the rightmost (proxy-appended) entry 198.51.100.9, got %q", got)
	}
}

func TestClientIPPrefersXRealIPFromOurProxy(t *testing.T) {
	got := ClientIP(req("10.0.0.5:41000", map[string]string{
		"X-Real-IP":       "198.51.100.9",
		"X-Forwarded-For": "203.0.113.77",
	}), Nginx())
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
	}), Nginx())
	if got != "198.51.100.9" {
		t.Fatalf("headers from a public peer must be ignored, got %q", got)
	}
}

func TestClientIPFallsBackToPeerWithNoHeaders(t *testing.T) {
	if got := ClientIP(req("127.0.0.1:5000", nil), Nginx()); got != "127.0.0.1" {
		t.Fatalf("want 127.0.0.1, got %q", got)
	}
	// A single-entry XFF from our proxy is the whole chain.
	if got := ClientIP(req("127.0.0.1:5000", map[string]string{"X-Forwarded-For": "198.51.100.9"}), Nginx()); got != "198.51.100.9" {
		t.Fatalf("want 198.51.100.9, got %q", got)
	}
}

// islandShape is what tools-api actually receives behind Cloudflare + cloudflared
// + Caddy, measured 2026-07-29 (bugs_open/139):
//
//   - X-Forwarded-For is whatever CADDY wrote, and Caddy OVERWRITES it with its own
//     immediate peer — the docker bridge gateway — rather than appending. So it is
//     a constant, identical for every visitor on earth.
//   - X-Real-IP is passed through by Caddy VERBATIM, so if one ever arrives it is
//     the caller's own value. (Cloudflare happens to strip it today; that is the
//     edge's behaviour, not the origin's, and it is not a guarantee this package
//     may lean on.)
//   - CF-Connecting-IP is written by Cloudflare and cannot be supplied by the
//     caller — the edge 403s a request that tries.
func islandShape(clientSuppliedRealIP string) *http.Request {
	h := map[string]string{
		"X-Forwarded-For":  "172.18.0.1",
		"CF-Connecting-IP": "2a02:c7c:f61f:ac00:37e4:bb5f:7ea9:389a",
	}
	if clientSuppliedRealIP != "" {
		h["X-Real-IP"] = clientSuppliedRealIP
	}
	return req("172.18.0.1:44000", h)
}

func TestClientIPOnTheIslandResolvesTheRealVisitor(t *testing.T) {
	got := ClientIP(islandShape("203.0.113.77"), CloudflareTunnel())
	if got == "203.0.113.77" {
		t.Fatal("trusted a client-supplied X-Real-IP: Caddy forwards that header verbatim")
	}
	if got == "172.18.0.1" {
		t.Fatal("resolved the docker bridge gateway — the constant every visitor shares (bugs_open/139)")
	}
	if got != "2a02:c7c:f61f:ac00:37e4:bb5f:7ea9:389a" {
		t.Fatalf("want the CF-Connecting-IP the edge wrote, got %q", got)
	}
}

// The failing branch, made executable: this is what adopting the OLD hard-coded
// default into tools-api would have done. It is why FrontEnd is a required
// argument rather than a package-level constant. Both outcomes below are wrong,
// and which one you get depends only on whether the edge happened to strip a
// header that day — neither is a client identity.
func TestNginxRulesOnTheIslandShapeAreBothWrong(t *testing.T) {
	// With a client-supplied X-Real-IP present, nginx's first rule hands the
	// caller their own chosen key.
	if got := ClientIP(islandShape("203.0.113.77"), Nginx()); got != "203.0.113.77" {
		t.Fatalf("expected the nginx rules to return the caller's forged value here, got %q", got)
	}
	// With it absent, nginx's second rule falls to Caddy's overwritten
	// X-Forwarded-For, which is the same constant for everybody.
	if got := ClientIP(islandShape(""), Nginx()); got != "172.18.0.1" {
		t.Fatalf("expected the nginx rules to return the gateway constant here, got %q", got)
	}
}

func TestDirectFrontEndBelievesNoHeaders(t *testing.T) {
	got := ClientIP(islandShape("203.0.113.77"), Direct())
	if got != "172.18.0.1" {
		t.Fatalf("Direct() must ignore every header and key on the peer, got %q", got)
	}
}

// A trusted header carrying something that is not an address is skipped, not
// returned: a value that cannot be an IP cannot be a client key. Without this the
// next header in the list would never be consulted.
func TestClientIPSkipsATrustedHeaderThatIsNotAnAddress(t *testing.T) {
	got := ClientIP(req("10.0.0.5:41000", map[string]string{
		"X-Real-IP":       "not-an-ip",
		"X-Forwarded-For": "198.51.100.9",
	}), Nginx())
	if got != "198.51.100.9" {
		t.Fatalf("want the fallback to the next trusted header, got %q", got)
	}
}

// The peer gate is the single thing keeping CloudflareTunnel() honest: trusting
// CF-Connecting-IP means trusting Cloudflare to be in front, and only this gate
// makes the header revert to being ignored if the origin is ever reachable
// without the tunnel. The architecture seat approved the FrontEnd change with one
// standing objection — that the reversion was unit-tested but its BOUNDARY was
// never stated — and asked that the next thing landing here close it.
//
// This pins where the boundary actually falls, measured rather than assumed.
// Note the two that may surprise: Go's net.IP.IsPrivate covers RFC1918 and
// RFC4193 ONLY, so a proxy behind CGNAT (100.64.0.0/10, RFC6598) or on a
// link-local address is NOT trusted and its headers are ignored. That is the safe
// direction to be wrong in — such a deployment keys on the peer instead of the
// real visitor, which degrades to a coarse key rather than a spoofable one — but
// it is a real constraint on where this package can be adopted, so it is stated
// here rather than discovered later.
func TestPeerGateBoundaryIsStatedNotAssumed(t *testing.T) {
	const forged = "203.0.113.77"
	for _, tc := range []struct {
		remoteAddr string
		peer       string
		trusted    bool
		why        string
	}{
		{"127.0.0.1:5000", "127.0.0.1", true, "IPv4 loopback"},
		{"[::1]:5000", "::1", true, "IPv6 loopback, bracketed as a real socket presents it"},
		{"10.0.0.5:41000", "10.0.0.5", true, "RFC1918"},
		{"172.18.0.1:44000", "172.18.0.1", true, "RFC1918, the island's docker gateway"},
		{"192.168.1.1:41000", "192.168.1.1", true, "RFC1918"},
		{"[fc00::1]:41000", "fc00::1", true, "RFC4193 unique-local"},
		{"100.64.0.1:41000", "100.64.0.1", false, "CGNAT — NOT covered by IsPrivate"},
		{"169.254.1.1:41000", "169.254.1.1", false, "link-local — NOT covered by IsPrivate"},
		{"198.51.100.9:33000", "198.51.100.9", false, "public IPv4: a direct caller"},
		{"[2a02:c7c:f61f:ac00::1]:443", "2a02:c7c:f61f:ac00::1", false, "public IPv6: a direct caller"},
	} {
		got := ClientIP(req(tc.remoteAddr, map[string]string{"CF-Connecting-IP": forged}), CloudflareTunnel())
		want := tc.peer
		if tc.trusted {
			want = forged
		}
		if got != want {
			t.Errorf("%s (%s): want %q, got %q", tc.remoteAddr, tc.why, want, got)
		}
	}
}

// What could NOT be closed locally, stated so the gap is not mistaken for
// coverage: every case above supplies RemoteAddr as a string. A genuine
// direct-exposure path — a real connection from a real public peer — is not
// constructible on a dev machine, because every address a local listener can bind
// is loopback or RFC1918 and therefore lands on the trusted side of the gate.
//
// This test closes the half that IS constructible, and it is the half where a bug
// would actually hide: that the peer is parsed correctly from what the RUNTIME
// puts in RemoteAddr, rather than from what a test author types. A real IPv6 peer
// arrives bracketed and with a port; leaking either into the key would silently
// fragment every limiter bucket per-connection.
func TestClientIPParsesTheRemoteAddrARealSocketProduces(t *testing.T) {
	var resolved, actualPeer string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualPeer = r.RemoteAddr
		resolved = ClientIP(r, Direct())
	}))
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL)
	if err != nil {
		t.Fatalf("probe request: %v", err)
	}
	resp.Body.Close()

	if resolved == "" {
		t.Fatal("resolved nothing from a real request")
	}
	if net.ParseIP(resolved) == nil {
		t.Fatalf("resolved %q, which is not a parseable IP — a port or brackets leaked into the key", resolved)
	}
	wantHost, _, err := net.SplitHostPort(actualPeer)
	if err != nil {
		t.Fatalf("real RemoteAddr %q did not split: %v", actualPeer, err)
	}
	if resolved != wantHost {
		t.Fatalf("real RemoteAddr %q should resolve to %q, got %q", actualPeer, wantHost, resolved)
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
