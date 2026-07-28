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

// ClientIP extracts a per-client key THE CLIENT CANNOT CHOOSE.
//
// THE INCIDENT (idea.uk, proven against production 2026-07-26, bugs_open/090).
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
//  2. Within those headers, take what OUR proxy wrote: X-Real-IP (set with
//     proxy_set_header, so a client-supplied one is replaced), else the RIGHTMOST
//     X-Forwarded-For entry — the hop nginx appended — never the first.
//
// A service exposed directly to the internet (no reverse proxy) therefore keys on
// the real peer address and ignores headers entirely, which is correct.
func ClientIP(r *http.Request) string {
	host := r.RemoteAddr
	if h, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		host = h
	}
	peer := net.ParseIP(host)
	if peer == nil || !(peer.IsLoopback() || peer.IsPrivate()) {
		return host // not from our proxy: believe nothing it claims
	}
	if xr := strings.TrimSpace(r.Header.Get("X-Real-IP")); xr != "" {
		return xr
	}
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.LastIndexByte(xff, ','); i >= 0 {
			if last := strings.TrimSpace(xff[i+1:]); last != "" {
				return last
			}
		} else if only := strings.TrimSpace(xff); only != "" {
			return only
		}
	}
	return host
}
