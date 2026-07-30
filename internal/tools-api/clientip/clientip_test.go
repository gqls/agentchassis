package clientip

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// ctx builds a gin context with a chosen peer address and headers.
func ctx(peer string, hdr map[string]string) *gin.Context {
	r := httptest.NewRequest(http.MethodGet, "/round", nil)
	r.RemoteAddr = peer
	for k, v := range hdr {
		r.Header.Set(k, v)
	}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = r
	return c
}

// The point of the fix: behind the island's chain the visitor's real address
// arrives ONLY in CF-Connecting-IP, and From must return it.
//
// This is deliberately a POSITIVE assertion. A test shaped as "the forged address
// no longer appears" passes against the unfixed code, because the unfixed code
// returned a constant that also is not the forged address.
func TestFrom_ReturnsCloudflareAddress(t *testing.T) {
	got := From(ctx("172.18.0.1:41000", map[string]string{
		"CF-Connecting-IP": "203.0.113.7",
	}))
	if got != "203.0.113.7" {
		t.Fatalf("want the CF-Connecting-IP address 203.0.113.7, got %q", got)
	}
}

// The regression this package exists to prevent, stated as a test.
//
// Caddy OVERWRITES X-Forwarded-For with its own peer, so on the island that
// header is the docker bridge gateway for every visitor. If someone "simplifies"
// From to httpguard.Nginx() — which trusts X-Forwarded-For and reads like the
// obvious choice — this fails. Without this test that swap is invisible: it
// compiles, it returns a plausible IP string, and it silently restores one global
// bucket for the whole site.
func TestFrom_IgnoresXForwardedFor(t *testing.T) {
	got := From(ctx("172.18.0.1:41000", map[string]string{
		"X-Forwarded-For": "172.18.0.1",
		"X-Real-IP":       "198.51.100.9",
	}))
	if got == "198.51.100.9" {
		t.Fatalf("X-Real-IP must not be trusted here — Cloudflare strips a client-supplied one, so its presence means a client wrote it")
	}
	if got != "172.18.0.1" {
		t.Fatalf("with no CF-Connecting-IP the peer is the only honest answer, want 172.18.0.1, got %q", got)
	}
}

// CF-Connecting-IP is believed only from a peer that could be our own proxy.
// A public peer means the origin was reached without the tunnel, and then the
// header is user input. Failing to a COARSE key is correct; failing to a
// SPOOFABLE one is the bug in bugs_closed/090.
func TestFrom_PublicPeerIgnoresHeaders(t *testing.T) {
	got := From(ctx("203.0.113.200:52000", map[string]string{
		"CF-Connecting-IP": "198.51.100.1",
	}))
	if got != "203.0.113.200" {
		t.Fatalf("headers from a public peer are user input; want the socket address 203.0.113.200, got %q", got)
	}
}

// An unparseable header must be skipped rather than returned — a value that is
// not an address cannot be a client key. This is also the inert-failure path if
// CF-Connecting-IP never reaches the process: we get the peer, i.e. exactly the
// pre-fix constant, and never a garbage key.
func TestFrom_UnparseableHeaderFallsBackToPeer(t *testing.T) {
	got := From(ctx("172.18.0.1:41000", map[string]string{
		"CF-Connecting-IP": "not-an-ip",
	}))
	if got != "172.18.0.1" {
		t.Fatalf("want fallback to peer 172.18.0.1, got %q", got)
	}
}

// Two visitors must produce two keys. The whole defect was that they did not:
// one distinct value across 83 of 83 rows. A single-request test cannot see that,
// which is why the acceptance check is count(DISTINCT …) > 1 and not a presence
// check — this is its unit-level shadow.
func TestFrom_DistinguishesTwoVisitors(t *testing.T) {
	a := From(ctx("172.18.0.1:41000", map[string]string{"CF-Connecting-IP": "203.0.113.7"}))
	b := From(ctx("172.18.0.1:41002", map[string]string{"CF-Connecting-IP": "203.0.113.8"}))
	if a == b {
		t.Fatalf("two visitors collapsed to one key %q — this is the 83-of-83 defect", a)
	}
}
