package main

import (
	"net/http/httptest"
	"testing"
)

const (
	realPeer = "198.51.100.9" // what Cloudflare's edge actually saw
	forged   = "203.0.113.77" // TEST-NET-3: can never be a real client
)

// The header shape a request through the tunnel actually carries: cloudflared
// sets CF-Connecting-IP from the edge connection. X-Forwarded-For and
// X-Real-IP, even if present, must NOT influence the result on this box.
func TestClientIPTrustsOnlyCFConnectingIP(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/chat", nil)
	r.Header.Set("CF-Connecting-IP", realPeer)
	r.Header.Set("X-Forwarded-For", forged)
	r.Header.Set("X-Real-IP", forged)
	if got := clientIP(r); got != realPeer {
		t.Fatalf("clientIP = %q, want %q — X-Forwarded-For/X-Real-IP leaked into the key", got, realPeer)
	}
}

// A request with no CF-Connecting-IP at all did not come through the tunnel
// (or the header was stripped somewhere) — it must not fall back to a spoofable
// header and must not resolve to a shared/empty key that different callers share.
func TestClientIPRejectsRequestsWithoutCFConnectingIP(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/chat", nil)
	r.Header.Set("X-Forwarded-For", forged)
	r.Header.Set("X-Real-IP", forged)
	if got := clientIP(r); got != "" {
		t.Fatalf("clientIP = %q, want empty — a request with no CF-Connecting-IP must not resolve to a usable key", got)
	}
}

// A garbage CF-Connecting-IP (should never happen from real Cloudflare traffic,
// but the header is still attacker-adjacent input) must not be trusted as-is.
func TestClientIPRejectsMalformedCFConnectingIP(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/chat", nil)
	r.Header.Set("CF-Connecting-IP", "not-an-ip; drop table conversations")
	if got := clientIP(r); got != "" {
		t.Fatalf("clientIP = %q, want empty for a malformed value", got)
	}
}

// Two different forged X-Forwarded-For values riding the SAME real
// CF-Connecting-IP must resolve to the SAME key — rotating the spoofable
// headers must not buy extra rate-limit budget.
func TestClientIPNotBypassableByRotatingForgedHeaders(t *testing.T) {
	key := func(forgedXFF string) string {
		r := httptest.NewRequest("POST", "/api/chat", nil)
		r.Header.Set("CF-Connecting-IP", realPeer)
		r.Header.Set("X-Forwarded-For", forgedXFF)
		return clientIP(r)
	}
	first := key("203.0.113.1")
	second := key("203.0.113.2")
	if first != second {
		t.Fatalf("clientIP varied with a spoofable header: %q vs %q", first, second)
	}
}

func TestClientIPHandlesIPv6(t *testing.T) {
	r := httptest.NewRequest("POST", "/api/chat", nil)
	r.Header.Set("CF-Connecting-IP", "2a02:c7c:f61f:ac00::1")
	if got := clientIP(r); got != "2a02:c7c:f61f:ac00::1" {
		t.Fatalf("clientIP = %q, want the IPv6 address unmangled", got)
	}
}
