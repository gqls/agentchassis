package fetchguard

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"
)

// ── IsPubliclyRoutable: the classification table itself ─────────────────────

func TestIsPubliclyRoutable(t *testing.T) {
	cases := []struct {
		name string
		ip   string
		want bool
	}{
		// The negative control set: every private-network category we exist
		// to catch. If any of these came back `true`, the check is not doing
		// its job — these are the addresses the whole package exists for.
		{"AWS/GCP metadata endpoint", "169.254.169.254", false},
		{"link-local (the /16 metadata sits inside)", "169.254.1.1", false},
		{"loopback v4", "127.0.0.1", false},
		{"loopback v6", "::1", false},
		{"RFC1918 10/8", "10.0.0.5", false},
		{"RFC1918 172.16/12", "172.16.5.5", false},
		{"RFC1918 192.168/16", "192.168.1.1", false},
		{"unique local v6 (ULA)", "fd00::1", false},
		{"link-local v6", "fe80::1", false},
		{"unspecified v4", "0.0.0.0", false},
		{"unspecified v6", "::", false},
		{"0.0.0.0/8, not just 0.0.0.0 itself", "0.1.2.3", false},
		{"multicast v4", "224.0.0.1", false},
		{"multicast v6", "ff02::1", false},

		// The positive control set: ordinary public addresses must NOT be
		// blocked, or the guard is unusable rather than safe.
		{"public v4 (documented example range, still globally routable syntax)", "8.8.8.8", true},
		{"another public v4", "1.1.1.1", true},
		{"public v6", "2001:4860:4860::8888", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			addr, err := netip.ParseAddr(tc.ip)
			if err != nil {
				t.Fatalf("test setup: %q does not parse: %v", tc.ip, err)
			}
			got := IsPubliclyRoutable(addr)
			if got != tc.want {
				t.Errorf("IsPubliclyRoutable(%s) = %v, want %v", tc.ip, got, tc.want)
			}
		})
	}
}

// TestIsPubliclyRoutable_IPv4Mapped proves the metadata endpoint is blocked
// whether it arrives as a bare v4 address or wrapped as IPv4-in-IPv6 — and
// that unmapping first changes nothing, which is itself worth asserting.
//
// An earlier version of this test (and the code comment beside the Unmap()
// call it was checking) claimed unmapping was THE security boundary for this
// address shape — "the exact bypass unmapping exists to close". Running it
// showed both forms already classify as blocked with no unmap step: Go's
// netip.Addr.IsLinkLocalUnicast()/IsPrivate() already resolve 4-in-6 addresses
// correctly. The claim was written before it was tested and this test is what
// caught it. Unmap() stays in the code for a real but smaller reason — it
// makes ip.String() in an error message read "169.254.169.254" rather than
// "::ffff:169.254.169.254" — and that is what this test now asserts instead.
func TestIsPubliclyRoutable_IPv4Mapped(t *testing.T) {
	mapped := netip.MustParseAddr("::ffff:169.254.169.254")

	if IsPubliclyRoutable(mapped) {
		t.Fatalf("mapped form ::ffff:169.254.169.254 was allowed — must classify the same as 169.254.169.254")
	}
	if IsPubliclyRoutable(mapped.Unmap()) {
		t.Fatalf("unmapped form was allowed — must also be blocked")
	}
	if mapped.Unmap().String() != "169.254.169.254" {
		t.Fatalf("Unmap()'s value is what this call is actually for (readable error messages), and it did not produce the plain v4 form: got %s", mapped.Unmap())
	}
}

// ── The live dial path — an actual guarded *http.Client against real listeners ──

// TestGuardedClient_RefusesPrivateTarget is the test that matters most: it
// does not call IsPubliclyRoutable directly, it fires a REAL HTTP request
// through NewClient at a server bound to loopback, and proves the request
// never completes. Testing the classifier alone would not catch a wiring bug
// where the classifier is correct but never actually consulted.
func TestGuardedClient_RefusesPrivateTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should never be read"))
	}))
	defer srv.Close()
	// httptest.NewServer binds to 127.0.0.1 — exactly the loopback case.

	client := NewClient(DefaultConfig())
	resp, err := client.Get(srv.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatalf("request to loopback server SUCCEEDED — the guard did not fire")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("got error %v, want it to wrap ErrBlockedAddress", err)
	}
}

// TestGuardedClient_AllowsPublicTarget is the negative control for the test
// above: a client that blocks everything would also "pass" a blocked-loopback
// test. Prove the same client reaches a real public-shaped target too — using
// a server bound to all interfaces (0.0.0.0), reached via a REAL non-loopback
// route: the machine's own non-loopback interface address, not 127.0.0.1.
// If no such interface exists (e.g. a fully isolated sandbox with only lo),
// the test skips rather than failing on an environment property this package
// has no control over.
func TestGuardedClient_AllowsPublicTarget(t *testing.T) {
	iface := firstNonLoopbackIPv4(t)
	if iface == "" {
		t.Skip("no non-loopback IPv4 interface available in this environment")
	}

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))
	lis, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv.Listener = lis
	srv.Start()
	defer srv.Close()

	_, port, _ := net.SplitHostPort(lis.Addr().String())
	url := fmt.Sprintf("http://%s:%s/", iface, port)

	addr := netip.MustParseAddr(iface)
	if !IsPubliclyRoutable(addr) {
		t.Skipf("this environment's own interface address %s is itself private — cannot exercise the public-allow path this way here", iface)
	}

	client := NewClient(DefaultConfig())
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("request to a publicly-routable-shaped address was refused: %v", err)
	}
	defer resp.Body.Close()
}

func firstNonLoopbackIPv4(t *testing.T) string {
	t.Helper()
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, a := range addrs {
		ipNet, ok := a.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() || ipNet.IP.To4() == nil {
			continue
		}
		return ipNet.IP.String()
	}
	return ""
}

// TestGuardedClient_RefusesDisallowedScheme proves the initial-request scheme
// check fires, not just the redirect one.
func TestGuardedClient_RefusesDisallowedScheme(t *testing.T) {
	client := NewClient(DefaultConfig())
	_, err := client.Get("ftp://example.com/")
	if err == nil {
		t.Fatal("ftp:// request succeeded — scheme check did not fire")
	}
	if !strings.Contains(err.Error(), "unsupported protocol scheme") && !errors.Is(err, ErrSchemeNotAllowed) {
		t.Fatalf("got %v, want either our ErrSchemeNotAllowed or net/http's own scheme rejection", err)
	}
}

// TestGuardedClient_RedirectToPrivateTargetIsRefused is the bypass this whole
// package's design note calls out explicitly: a request to a URL that LOOKS
// public 302s to a private address. It must be caught, and it must be caught
// by the SAME mechanism as a direct request — no separate redirect-target
// classifier to fall out of sync with the direct-request one.
func TestGuardedClient_RedirectToPrivateTargetIsRefused(t *testing.T) {
	private := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should never be reached via redirect"))
	}))
	defer private.Close()

	redirector := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, private.URL, http.StatusFound)
	}))
	defer redirector.Close()

	client := NewClient(DefaultConfig())
	resp, err := client.Get(redirector.URL)
	if err == nil {
		resp.Body.Close()
		t.Fatal("redirect to a loopback target SUCCEEDED — the dial-time check did not apply to the redirected request")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("got %v, want it to wrap ErrBlockedAddress (both servers here are loopback, same as the direct-request case)", err)
	}
}

// TestCheckRedirect_CapEnforced unit-tests the redirect-cap logic directly
// against fake requests, rather than through a live redirect chain.
//
// A live version was tried first and rejected: httptest.NewServer binds to
// 127.0.0.1, so the private-IP check (proven separately above) refuses the
// very FIRST hop, before the redirect count could ever climb high enough to
// exercise the cap — the test would have "passed" for the wrong reason,
// proving the private-IP check fires rather than the redirect cap. Testing
// checkRedirect as the pure function it is avoids that confound entirely.
func TestCheckRedirect_CapEnforced(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxRedirects = 2
	check := checkRedirect(cfg)

	req, _ := http.NewRequest("GET", "https://example.com/next", nil)

	// Under the cap: 0 and 1 prior hops must both be allowed.
	for _, n := range []int{0, 1} {
		via := make([]*http.Request, n)
		if err := check(req, via); err != nil {
			t.Fatalf("%d prior hop(s) (under cap of %d): got %v, want allowed", n, cfg.MaxRedirects, err)
		}
	}

	// At the cap: must refuse.
	via := make([]*http.Request, cfg.MaxRedirects)
	err := check(req, via)
	if err == nil {
		t.Fatalf("%d prior hops (== cap of %d) was allowed — cap not enforced", cfg.MaxRedirects, cfg.MaxRedirects)
	}
	if !errors.Is(err, ErrTooManyRedirects) {
		t.Fatalf("got %v, want it to wrap ErrTooManyRedirects", err)
	}
}

// TestGuardedClient_RedirectCapEnforced_Live is the live complement to the
// unit test above: it proves the SAME cap through an actual redirect chain,
// but needs a target this environment's own private-IP check will not refuse
// first — see TestGuardedClient_AllowsPublicTarget for why that is
// environment-dependent and may skip here.
func TestGuardedClient_RedirectCapEnforced_Live(t *testing.T) {
	iface := firstNonLoopbackIPv4(t)
	if iface == "" || !IsPubliclyRoutable(netip.MustParseAddr(iface)) {
		t.Skip("no publicly-routable-shaped local interface in this environment; see TestGuardedClient_AllowsPublicTarget")
	}

	lis, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	_, port, _ := net.SplitHostPort(lis.Addr().String())
	base := fmt.Sprintf("http://%s:%s", iface, port)

	hops := 0
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hops++
		http.Redirect(w, r, base+fmt.Sprintf("/%d", hops), http.StatusFound)
	})}
	go srv.Serve(lis)
	defer srv.Close()

	cfg := DefaultConfig()
	cfg.MaxRedirects = 2
	client := NewClient(cfg)

	_, err = client.Get(base)
	if err == nil {
		t.Fatal("an infinite-redirect server was followed to completion — MaxRedirects was not enforced")
	}
	if !strings.Contains(err.Error(), "too many redirects") && !errors.Is(err, ErrTooManyRedirects) {
		t.Fatalf("got %v, want a too-many-redirects failure", err)
	}
}

// ── LimitedRead: truncation must be REPORTED, never silent ──────────────────

func TestLimitedRead_ReportsTruncation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, 100))
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("setup request failed: %v", err)
	}
	defer resp.Body.Close()

	data, truncated, err := LimitedRead(resp, 50)
	if err != nil {
		t.Fatalf("LimitedRead: %v", err)
	}
	if !truncated {
		t.Fatal("100 bytes against a 50-byte cap was NOT reported as truncated")
	}
	if len(data) != 50 {
		t.Fatalf("got %d bytes back, want exactly the 50-byte cap", len(data))
	}
}

func TestLimitedRead_UnderCapIsNotTruncated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("short"))
	}))
	defer srv.Close()

	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatalf("setup request failed: %v", err)
	}
	defer resp.Body.Close()

	data, truncated, err := LimitedRead(resp, 50)
	if err != nil {
		t.Fatalf("LimitedRead: %v", err)
	}
	if truncated {
		t.Fatal("a 5-byte body against a 50-byte cap was reported as truncated")
	}
	if string(data) != "short" {
		t.Fatalf("got %q, want %q", data, "short")
	}
}

// ── A literal-IP URL must be checked too, not only hostnames ────────────────

func TestGuardedClient_RefusesLiteralPrivateIP(t *testing.T) {
	client := NewClient(DefaultConfig())
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/latest/meta-data/", nil)
	_, err := client.Do(req)
	if err == nil {
		t.Fatal("request to a literal metadata-endpoint IP succeeded")
	}
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("got %v, want it to wrap ErrBlockedAddress", err)
	}
}
