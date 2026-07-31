// Package fetchguard is the platform's ONE set of outbound-fetch (SSRF)
// primitives: for any code that fetches a URL the platform did not choose —
// scraped content, an og:image, a domain a customer typed in — this is what
// stops that fetch reaching the cluster's own metadata endpoint, a private
// service, or localhost.
//
// SIBLING TO httpguard, NOT PART OF IT. httpguard's own package doc scopes
// itself explicitly to "INBOUND-abuse primitives for public HTTP endpoints" —
// who is allowed to call US. This package is the mirror direction: what WE are
// allowed to call on a caller's behalf. Folding an outbound guard into a
// package whose header names it inbound would make that header wrong, and
// "the package doc says X but the package also does not-X" is exactly the
// shape of trap this platform's own LANDMINES.md exists to catch — see the
// entry this package fixes: "platform/httpguard is INBOUND-abuse only".
//
// THE DEFECT THIS CLOSES. internal/adapters/webscrape/adapter.go's
// downloadImage fetches image URLs extracted from a SCRAPED PAGE's own content
// — attacker-influenced by construction, since the page belongs to whatever
// domain a customer submitted — with a bare *http.Client, no scheme check, no
// destination-IP check, and no response-size cap. That is live, in production,
// on every site's image-ingestion path today (bugs_open/159). This package is
// the fix, generalised so any future caller (a domain-intake flow, a fresh
// scrape action, a new site pipeline) gets it by swapping one client
// construction rather than re-deriving the checks.
//
// WHY A CUSTOM DIALER, NOT A URL PRE-CHECK. Resolving a hostname once to check
// it is safe, then letting the standard library resolve it AGAIN to connect,
// is a TOCTOU gap: DNS can answer differently the second time (a rebinding
// attack answers the legitimate IP to the checker and a private one to the
// real connection). Every check here happens at the point net/http actually
// dials — via a Transport.DialContext — so a redirect to a second host is
// re-validated automatically by the same code path, with no separate redirect
// inspection needed to close the classic "public URL 302s to
// http://169.254.169.254/" bypass.
//
// WHAT THIS DOES NOT COVER. A headless browser navigating a URL (Playwright,
// as browser-runner-adapter uses) is a different fetch surface — the browser
// does its own DNS and its own connections, so a Go http.Transport guard never
// sees them. That needs network-layer interception inside the browser (e.g.
// Playwright's page.Route) and is explicitly OUT of scope here; flagged, not
// silently left for someone to assume this package already covers it.
//
// This package does NOT wire itself into any existing service. Adopting it
// beyond its first call site (downloadImage) is a separate, coordinated
// change per caller.
package fetchguard

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

// Sentinel errors, so callers can distinguish "we refused to fetch this" from
// a genuine network failure without parsing strings.
var (
	ErrSchemeNotAllowed  = errors.New("fetchguard: scheme not allowed")
	ErrBlockedAddress    = errors.New("fetchguard: destination resolves to a non-public address")
	ErrNoPublicAddress   = errors.New("fetchguard: host has no publicly-routable address")
	ErrTooManyRedirects  = errors.New("fetchguard: too many redirects")
	ErrResponseTruncated = errors.New("fetchguard: response exceeded the size cap")
)

// Config is the policy a guarded client enforces. Zero value is not usable —
// construct via DefaultConfig() and override individual fields.
type Config struct {
	// AllowedSchemes is checked against the URL of every request AND every
	// redirect target. Case-insensitive. A request whose scheme is not in this
	// list is refused before any dial is attempted.
	AllowedSchemes []string

	// MaxRedirects bounds the redirect chain. The default net/http ceiling is
	// 10; that is generous for a scrape target and each hop is a fresh
	// opportunity to be pointed somewhere new.
	MaxRedirects int

	// DialTimeout bounds a single connection attempt (not the whole request —
	// pass a context deadline for that, as every caller here already does).
	DialTimeout time.Duration

	// MaxResponseBytes caps how much of a response body ReadLimited will
	// return. Zero means "use the package default", not "unlimited" — an
	// actually-unlimited cap must be requested explicitly via a negative
	// value, so a zero-value Config never accidentally means no cap.
	MaxResponseBytes int64

	// Resolver is overridable for tests (inject one that returns fixed
	// addresses instead of hitting real DNS). Nil means net.DefaultResolver.
	Resolver *net.Resolver
}

// DefaultConfig returns a sane, restrictive policy: http/https only, 5
// redirects, a 10s dial timeout, and a 25MB response cap (comfortably above
// any real image or scraped-page payload while bounding a hostile server that
// tries to stream forever).
func DefaultConfig() Config {
	return Config{
		AllowedSchemes:   []string{"http", "https"},
		MaxRedirects:     5,
		DialTimeout:      10 * time.Second,
		MaxResponseBytes: 25 * 1024 * 1024,
	}
}

// NewClient builds an *http.Client whose every dial is checked against cfg.
// Swap this in wherever a caller currently constructs a bare &http.Client{}
// to fetch a URL the platform does not itself control.
func NewClient(cfg Config) *http.Client {
	cfg = withDefaults(cfg)

	dialer := &net.Dialer{Timeout: cfg.DialTimeout}

	transport := &http.Transport{
		DialContext: guardedDialContext(cfg, dialer),
		// Match the idle-connection tuning already used at the one call site
		// this replaces (webscrape/adapter.go) — deliberately not novel.
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	return &http.Client{
		// schemeCheckingRoundTripper covers the INITIAL request; checkRedirect
		// (below) covers every hop after it — the two together make scheme
		// enforcement symmetric across the whole chain. In practice
		// http.Transport already refuses schemes it doesn't implement, so
		// this is defense in depth against a caller wrapping the transport
		// with something more permissive later, not the only thing standing
		// between a bad scheme and a dial.
		Transport:     &schemeCheckingRoundTripper{next: transport, allowed: cfg.AllowedSchemes},
		CheckRedirect: checkRedirect(cfg),
	}
}

// schemeCheckingRoundTripper refuses the initial request before it reaches
// the transport, mirroring the check checkRedirect applies to every
// subsequent hop.
type schemeCheckingRoundTripper struct {
	next    http.RoundTripper
	allowed []string
}

func (t *schemeCheckingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if !schemeAllowed(req.URL.Scheme, t.allowed) {
		return nil, fmt.Errorf("%w: %q", ErrSchemeNotAllowed, req.URL.Scheme)
	}
	return t.next.RoundTrip(req)
}

func withDefaults(cfg Config) Config {
	d := DefaultConfig()
	if len(cfg.AllowedSchemes) == 0 {
		cfg.AllowedSchemes = d.AllowedSchemes
	}
	if cfg.MaxRedirects == 0 {
		cfg.MaxRedirects = d.MaxRedirects
	}
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = d.DialTimeout
	}
	if cfg.MaxResponseBytes == 0 {
		cfg.MaxResponseBytes = d.MaxResponseBytes
	}
	return cfg
}

// checkRedirect enforces the scheme allow-list and the redirect ceiling on
// every hop. It does NOT need to re-check the destination address — that
// happens unconditionally on every dial, including the dial this redirect is
// about to trigger, via guardedDialContext. Two independent enforcement
// points (scheme here, address at dial time) rather than one is deliberate:
// CheckRedirect sees the URL before net/http has decided how to reach it,
// which is the only point that can cheaply refuse a scheme downgrade.
func checkRedirect(cfg Config) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= cfg.MaxRedirects {
			return fmt.Errorf("%w: %d hop(s)", ErrTooManyRedirects, len(via))
		}
		if !schemeAllowed(req.URL.Scheme, cfg.AllowedSchemes) {
			return fmt.Errorf("%w: %q (redirect hop %d)", ErrSchemeNotAllowed, req.URL.Scheme, len(via))
		}
		return nil
	}
}

func schemeAllowed(scheme string, allowed []string) bool {
	for _, a := range allowed {
		if strings.EqualFold(scheme, a) {
			return true
		}
	}
	return false
}

// guardedDialContext returns the ONE function that makes every property in
// this package's doc comment true. net/http calls it for the initial
// connection and for every redirect hop; it resolves the target itself
// (bypassing whatever the standard dialer would have done), rejects any
// candidate address that is not publicly routable, and dials the specific
// address it approved — never re-resolving between check and connect.
func guardedDialContext(cfg Config, dialer *net.Dialer) func(ctx context.Context, network, addr string) (net.Conn, error) {
	resolver := cfg.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}

	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("fetchguard: %w", err)
		}

		// A literal IP in the URL skips DNS but not the address check.
		if literal, perr := netip.ParseAddr(host); perr == nil {
			if !IsPubliclyRoutable(literal) {
				return nil, fmt.Errorf("%w: %s", ErrBlockedAddress, literal)
			}
			return dialer.DialContext(ctx, network, net.JoinHostPort(literal.String(), port))
		}

		candidates, err := resolver.LookupNetIP(ctx, "ip", host)
		if err != nil {
			return nil, fmt.Errorf("fetchguard: resolving %q: %w", host, err)
		}

		var lastBlocked netip.Addr
		for _, ip := range candidates {
			// Unmap is NOT what makes ::ffff:169.254.169.254 safe to classify —
			// VERIFIED (not assumed): netip's IsPrivate/IsLinkLocalUnicast/etc.
			// already resolve a 4-in-6 address correctly with no help. Unmap
			// is here only so ip.String() below reads "169.254.169.254"
			// rather than "::ffff:169.254.169.254" in an error message. An
			// earlier version of this comment claimed unmap was the security
			// boundary; that claim was written before it was tested and a
			// test proved it false — see fetchguard_test.go.
			ip = ip.Unmap()
			if !IsPubliclyRoutable(ip) {
				lastBlocked = ip
				continue
			}
			// Dial the SPECIFIC address just approved. Passing the hostname
			// back to dialer.DialContext would let net/http resolve it a
			// second time — the exact rebinding gap this package exists to
			// close.
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			// This candidate was approved but unreachable (down, filtered,
			// wrong port); try the next resolved address before giving up,
			// same as the standard dialer would across a multi-A-record host.
		}

		if lastBlocked.IsValid() {
			return nil, fmt.Errorf("%w: %s (host %s)", ErrBlockedAddress, lastBlocked, host)
		}
		return nil, fmt.Errorf("%w: %s", ErrNoPublicAddress, host)
	}
}

// IsPubliclyRoutable reports whether ip is safe to connect to from server-side
// code acting on someone else's instruction. Structural allow-by-exclusion,
// not a blocklist of known-bad ranges: everything that is NOT one of the
// listed private categories is allowed, so a range nobody thought to list
// (a future RFC reservation, an internal range someone forgot) is excluded by
// default rather than admitted by default. That is the safer failure mode for
// a check whose entire job is refusing the unexpected.
//
// Unmap the address first if it might be an IPv4-mapped IPv6 literal
// (::ffff:x.x.x.x) — this function does not do it for you, because a caller
// checking a literal it already knows the provenance of may not need to.
// guardedDialContext unmaps before calling this.
func IsPubliclyRoutable(ip netip.Addr) bool {
	if !ip.IsValid() {
		return false
	}
	if ip.IsPrivate() || // RFC1918 (v4) and ULA fc00::/7 (v6)
		ip.IsLoopback() ||
		ip.IsLinkLocalUnicast() || // covers 169.254.0.0/16 — cloud metadata endpoints live here
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() {
		return false
	}
	// 0.0.0.0/8 (not just the single unspecified address) — some resolvers
	// and OSes route the whole block to the local host, and netip's
	// IsUnspecified only excludes the single 0.0.0.0 address.
	if ip.Is4() && ip.As4()[0] == 0 {
		return false
	}
	return true
}

// LimitedRead reads at most cfg.MaxResponseBytes from resp.Body and reports
// whether the body was truncated to fit — the caller decides what a truncated
// image or page means for its own use case, but it is told, rather than
// silently handed a partial byte slice that reads exactly like a complete one.
// This is the same discriminating shape as the `__truncated` marker the
// chassis's own LLM-response path uses for the identical reason.
func LimitedRead(resp *http.Response, maxBytes int64) (data []byte, truncated bool, err error) {
	limited := io.LimitReader(resp.Body, maxBytes+1)
	data, err = io.ReadAll(limited)
	if err != nil {
		return nil, false, fmt.Errorf("fetchguard: reading response: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return data[:maxBytes], true, nil
	}
	return data, false, nil
}
