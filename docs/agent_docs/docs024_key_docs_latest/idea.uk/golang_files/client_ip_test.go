package main

// client_ip_test.go — clientIP must return a key the caller cannot choose.
//
// The per-IP limiters are the only thing bounding LLM spend on the free taster
// (3/hour, 20/day) and spam on the intake form (5/hour, 15/day), and the IP
// stored on an order exists to seed a future block list. All three are worth
// exactly as much as clientIP's resistance to a forged header.
//
// The header shapes below are the ones our nginx actually produces:
//   proxy_set_header X-Real-IP       $remote_addr;               // replaced, never client-supplied
//   proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for; // "<whatever arrived>, <real peer>"
// See bugs_open/090.

import (
	"net/http/httptest"
	"testing"
)

const (
	realPeer = "198.51.100.9" // the address nginx saw
	forged   = "203.0.113.77" // TEST-NET-3: can never be a real client
)

func TestClientIPIgnoresForgedForwardedFor(t *testing.T) {
	cases := []struct {
		name    string
		xff     string
		xrealip string
	}{
		{
			// The live reproduction: caller sends its own XFF, nginx appends the peer.
			name: "forged entry prepended by the caller, real peer appended by nginx",
			xff:  forged + ", " + realPeer,
		},
		{
			name: "several forged hops, real peer still last",
			xff:  forged + ", 192.0.2.5, 198.18.0.1, " + realPeer,
		},
		{
			// X-Real-IP is set with proxy_set_header, so nginx replaces any
			// client-supplied value: it must win over a forged XFF.
			name:    "forged XFF with X-Real-IP present",
			xff:     forged + ", " + realPeer,
			xrealip: realPeer,
		},
		{
			name: "no spoofing at all — plain single-entry XFF from nginx",
			xff:  realPeer,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/audience-check", nil)
			r.RemoteAddr = "127.0.0.1:41234" // nginx proxy_pass http://127.0.0.1:8080
			if c.xff != "" {
				r.Header.Set("X-Forwarded-For", c.xff)
			}
			if c.xrealip != "" {
				r.Header.Set("X-Real-IP", c.xrealip)
			}
			if got := clientIP(r); got != realPeer {
				t.Fatalf("clientIP = %q, want %q — a caller can choose its own rate-limit key", got, realPeer)
			}
		})
	}
}

// A request that did NOT come from our proxy has no business claiming to speak
// for anyone else, whatever headers it carries.
func TestClientIPIgnoresHeadersFromAnUntrustedPeer(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "203.0.113.200:5555" // direct hit, not via nginx
	r.Header.Set("X-Forwarded-For", forged)
	r.Header.Set("X-Real-IP", forged)
	if got := clientIP(r); got != "203.0.113.200" {
		t.Fatalf("clientIP = %q, want the actual peer 203.0.113.200", got)
	}
}

// IPv6 peers must not lose their address to naive host:port splitting — the
// order placed on 2026-07-26 came from an IPv6 client, so this is the live shape.
func TestClientIPHandlesIPv6Peer(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "[2a02:c7c:f61f:ac00::1]:60123"
	if got := clientIP(r); got != "2a02:c7c:f61f:ac00::1" {
		t.Fatalf("clientIP = %q, want 2a02:c7c:f61f:ac00::1", got)
	}
}

// The limiter must actually be driven by the un-spoofable key: two requests
// wearing different forged addresses but arriving from the same peer are the
// same client.
func TestRateLimiterNotBypassableByRotatingForgedHeaders(t *testing.T) {
	rl := newRateLimiter() // 3/hour
	key := func(forgedIP string) string {
		r := httptest.NewRequest("POST", "/audience-check", nil)
		r.RemoteAddr = "127.0.0.1:41234"
		r.Header.Set("X-Forwarded-For", forgedIP+", "+realPeer)
		return clientIP(r)
	}
	for i, forgedIP := range []string{"203.0.113.1", "203.0.113.2", "203.0.113.3"} {
		if ok, _ := rl.allow(key(forgedIP)); !ok {
			t.Fatalf("request %d refused too early", i+1)
		}
	}
	if ok, _ := rl.allow(key("203.0.113.4")); ok {
		t.Fatal("4th request allowed: rotating a forged X-Forwarded-For still buys unlimited free LLM calls")
	}
}
