// Package clientip resolves the per-visitor key for tools-api.
//
// It exists so that "who is this visitor" is answered in exactly ONE place for
// this service, the way httperr centralises the error shape. Every caller that
// wants a per-client key — a rate limiter, a stored hash, a future block list —
// must go through From, because the correct answer depends on the proxy chain in
// front of THIS service and nothing about a gin.Context reveals that.
//
// WHY NOT c.ClientIP(), WHICH IS WHAT WE USED BEFORE
//
// gin's ClientIP trusts X-Forwarded-For. On the island that header is a constant:
// Caddy OVERWRITES it with its own immediate peer rather than appending to the
// chain, and that peer is the docker bridge gateway. Measured by the
// consolidation lane on 2026-07-29 (bugs_open/139): client_ip_hash was
// sha256("172.18.0.1") in 83 of 83 rows since 2026-07-25 — ONE distinct value
// across the whole table. So the per-IP limiter was a single global bucket shared
// by every visitor, and the stored column had never distinguished anybody.
//
// WHY CloudflareTunnel() AND NOT Nginx() — THE TRAP
//
// httpguard.ClientIP requires a FrontEnd precisely because there is no safe
// default across the estate. Nginx() is the wrong one here and is dangerous
// BECAUSE IT LOOKS RIGHT: it reproduces the old hard-coded behaviour, resolves to
// the same 172.18.0.1 constant behind Caddy, and would therefore read as a fix
// while changing nothing measurable.
//
// The island's chain is Cloudflare edge -> cloudflared tunnel -> Caddy -> here.
// CF-Connecting-IP is the only header that carries the visitor AND that the
// visitor cannot choose: Cloudflare 403s a request supplying its own
// (error code 1000) and writes a genuine one itself, Caddy forwards it untouched,
// while X-Real-IP is stripped at the edge and X-Forwarded-For is the constant
// above. That is what CloudflareTunnel() declares.
//
// WHAT KEEPS IT HONEST, AND HOW IT FAILS
//
// httpguard.ClientIP believes forwarding headers only from a loopback or RFC1918
// peer. Our peer is 172.18.0.1, which is RFC1918, so the gate passes. If tools-api
// ever moves off docker networking to a public peer, the gate stops trusting the
// header and From returns the socket address — a COARSE key, not a spoofable one.
// That is the correct direction to fail in, and it is the reason not to hand-roll
// this.
//
// The same property covers the one thing consolidation could not prove: that
// CF-Connecting-IP actually reaches this process (they measured it arriving at
// Caddy and being forwarded verbatim, but would not add a header-echo endpoint to
// our service). If it does not arrive, ClientIP skips the unparseable header and
// returns the peer — i.e. exactly today's constant. So this change cannot make
// things worse than they already are; it either improves the key or is inert.
// Settle it with the acceptance check, which is the only test that can:
//
//	count(DISTINCT client_ip_hash) > 1 across visits from two different networks
//
// A presence check ("the forged address no longer appears") passes against the
// UNFIXED code, and one test machine cannot tell a constant from a working key.
package clientip

import (
	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/platform/httpguard"
)

// From returns the per-visitor key for this request: an address the visitor
// cannot choose, given the island's Cloudflare-tunnel front end.
//
// Use this for anything keyed per client. Do not call c.ClientIP() directly and
// do not read forwarding headers by hand — see the package comment for what that
// cost us.
func From(c *gin.Context) string {
	return httpguard.ClientIP(c.Request, httpguard.CloudflareTunnel())
}
