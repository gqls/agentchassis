// Package httpguard is the platform's ONE set of inbound-abuse primitives for
// public HTTP endpoints: a trustworthy client key, a banded per-IP limiter, and
// the honeypot/timing gate for forms.
//
// WHY THIS EXISTS. Three per-IP limiters existed in the estate and the WEAKEST
// one guarded the only public endpoint. tools-api's
// (internal/tools-api/middleware/ratelimit.go) is a token bucket keyed on gin's
// c.ClientIP(), which trusts forwarding headers. idea.uk's is banded, returns a
// retry-after, and — crucially — derives the key from a peer it is willing to
// trust, after a live incident. The better implementation was unreachable, living
// in docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/ outside the
// build (features_open/024 A3).
//
// Deliberately framework-agnostic: net/http types only, no gin. tools-api is gin,
// idea.uk is net/http, and a shared package that picks one forces the other to
// wrap or fork. Callers write their own three-line middleware adapter.
//
// This package does NOT wire itself into any existing service. Adopting it in
// tools-api is a change to another workstream's code (gauntlet thread,
// bugs_open/083 open against it) and belongs in a separate, coordinated commit.
package httpguard

import (
	"net"
	"net/http"
	"strings"
)

// Entry says where in a header's value the address our proxy wrote can be found.
type Entry int

const (
	// EntryWhole: the header carries exactly one address, written by our proxy.
	EntryWhole Entry = iota
	// EntryRightmost: the header carries a comma-separated chain to which our
	// proxy APPENDED, so the last entry is the hop it observed. Never the first —
	// the first is whatever the caller sent (bugs_closed/090).
	EntryRightmost
)

// TrustedHeader names one header this deployment's proxy is known to WRITE (not
// merely forward), and where in it the address sits.
//
// "Writes" is the load-bearing word, and it is the whole reason this type exists
// rather than a hard-coded list. A header a proxy merely PASSES THROUGH is user
// input wearing a trusted name.
type TrustedHeader struct {
	Name  string
	Entry Entry
}

// FrontEnd is what the proxy chain in front of THIS deployment guarantees.
//
// WHY THIS IS A PARAMETER AND NOT A DEFAULT (corrected 2026-07-29, measured).
// This function used to hard-code nginx's behaviour: prefer X-Real-IP, "set with
// proxy_set_header, so a client-supplied one is replaced", else the rightmost
// X-Forwarded-For. That is true of nginx on idea.uk, where this code was written
// and proven. **It is false of Caddy**, which the island runs: Caddy does not set
// X-Real-IP at all and forwards a client-supplied one VERBATIM, and it OVERWRITES
// X-Forwarded-For with its own immediate peer rather than appending to the chain.
//
// So on the island every rule above resolves to either user input or a constant.
// Adopting the old default into tools-api would have keyed every visitor on the
// docker bridge gateway — measured: one distinct value across 83 of 83 rows —
// while reading like a fix. Evidence and the full per-hop measurement:
// bugs_open/139, 016b section 9 ("Who is the client" is decided by the PROXY CHAIN).
//
// The rule the estate should carry: a "trustworthy client key" helper is only as
// trustworthy as the specific proxy in front of the ADOPTING service. That cannot
// be a package-level constant, so it is an argument the caller must supply.
type FrontEnd struct {
	// Name is for error messages and logs; it has no effect on resolution.
	Name string
	// Trusted are consulted in order; the first that yields a parseable address
	// wins. Empty means believe no headers at all.
	Trusted []TrustedHeader
}

// Nginx is the front-end idea.uk runs: nginx with proxy_set_header X-Real-IP and
// $proxy_add_x_forwarded_for. Both rules are safe there because nginx WRITES both.
func Nginx() FrontEnd {
	return FrontEnd{
		Name: "nginx",
		Trusted: []TrustedHeader{
			{Name: "X-Real-IP", Entry: EntryWhole},
			{Name: "X-Forwarded-For", Entry: EntryRightmost},
		},
	}
}

// CloudflareTunnel is the front-end the island runs: Cloudflare's edge, a
// cloudflared tunnel, then Caddy.
//
// CF-Connecting-IP is the ONLY header here that carries the visitor and that the
// visitor cannot choose, and all three facts were measured on 2026-07-29:
//
//   - Cloudflare REFUSES a request that supplies its own CF-Connecting-IP —
//     403, "error code: 1000" — and sets a genuine one itself.
//   - Caddy forwards it to the upstream untouched.
//   - The other two candidates are useless here: Cloudflare STRIPS a
//     client-supplied X-Real-IP at the edge (so it never arrives), and Caddy
//     OVERWRITES X-Forwarded-For with its own peer (so it is a constant).
//
// Trusting this header is therefore trusting Cloudflare to be in front. The peer
// gate in ClientIP is what keeps that honest: if the origin is ever reachable
// without the tunnel, the peer is public, headers are ignored, and this reverts to
// the socket address rather than silently becoming spoofable.
func CloudflareTunnel() FrontEnd {
	return FrontEnd{
		Name:    "cloudflare-tunnel",
		Trusted: []TrustedHeader{{Name: "CF-Connecting-IP", Entry: EntryWhole}},
	}
}

// Direct is for a service exposed straight to the internet with no reverse proxy.
// It believes no headers, which is correct: there is no hop that could have
// written one.
func Direct() FrontEnd { return FrontEnd{Name: "direct"} }

// ClientIP extracts a per-client key THE CLIENT CANNOT CHOOSE, given what the
// caller's own proxy chain guarantees.
//
// THE INCIDENT (idea.uk, proven against production 2026-07-26, bugs_closed/090).
// The original trusted the FIRST X-Forwarded-For entry — which is exactly the
// part a caller writes for themselves. nginx forwards with
// $proxy_add_x_forwarded_for, which APPENDS the real peer to whatever arrived, so
// a request carrying "X-Forwarded-For: 203.0.113.77" arrived as
// "203.0.113.77, <real ip>" and was keyed on the invented address. The refusal log
// recorded 203.0.113.77 verbatim. That defeats every per-IP limiter downstream and
// poisons any IP captured for a future block list.
//
// Two rules, in order:
//
//  1. Forwarding headers are believed ONLY from a peer that is plausibly our own
//     proxy (loopback or RFC1918). A direct caller's headers are user input.
//  2. Within those, consult only the headers `front` declares its proxy WRITES,
//     in order, taking the entry that proxy controls. A header that is absent,
//     empty, or not a parseable IP is skipped rather than returned — a value that
//     is not an address cannot be a client key.
//
// A service exposed directly to the internet therefore passes Direct() and keys on
// the real peer address, which is correct.
//
// The FrontEnd argument is deliberately required. See its doc comment: the
// previous package-level default silently assumed nginx and would have resolved a
// constant on the estate's other front-end.
func ClientIP(r *http.Request, front FrontEnd) string {
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		host = h
	}
	peer := net.ParseIP(host)
	if peer == nil || !(peer.IsLoopback() || peer.IsPrivate()) {
		return host // not from our proxy: believe nothing it claims
	}
	for _, th := range front.Trusted {
		candidate := pickEntry(r.Header.Get(th.Name), th.Entry)
		if candidate == "" || net.ParseIP(candidate) == nil {
			continue
		}
		return candidate
	}
	return host
}

// pickEntry returns the address our proxy controls within one header value, or ""
// when there is nothing usable there.
func pickEntry(raw string, entry Entry) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if entry == EntryRightmost {
		if i := strings.LastIndexByte(raw, ','); i >= 0 {
			return strings.TrimSpace(raw[i+1:])
		}
	}
	return raw
}
