package main

// clientip.go — clientIP must return a key the caller cannot choose.
//
// This box is DIFFERENT from idea.uk's shape (see idea.uk/golang_files/
// client_ip_test.go, which trusts X-Real-IP/X-Forwarded-For because idea.uk's
// nginx sits in front with proxy_set_header overwriting them). webdesign.uk's
// ONLY ingress is a Cloudflare Tunnel: cloudflared dials OUT from this box to
// Cloudflare's edge, nginx binds 127.0.0.1:8080 only, and ufw denies all
// inbound. So every request this service ever sees arrived via Cloudflare's
// edge network, and Cloudflare sets CF-Connecting-IP from the real TCP peer
// at ITS edge — overwriting, never trusting, anything the original client
// tried to send in that header. X-Forwarded-For and X-Real-IP are NOT safe
// here: a visitor's own request to Cloudflare can carry whatever XFF/X-Real-IP
// it likes, and cloudflared/nginx have no reason to strip it before it reaches
// us, so trusting those two headers on this box would hand every visitor a
// free choice of rate-limit key. CF-Connecting-IP is the one header Cloudflare
// itself authors from the edge connection — see bugs_open/139 (the "per-IP
// limiter behind Caddy+Cloudflare is one global bucket" landmine) for the
// class of bug this file exists to not repeat.
//
// Proof this key is real, not a constant, is a LIVE check, not a unit test:
// two requests from two different networks must show two different values —
// see RUNBOOK "two-network CF-Connecting-IP proof".

import (
	"net"
	"net/http"
)

// clientIP returns the real visitor address, or "" if the request did not
// carry a CF-Connecting-IP at all — which on this box means it did not come
// through the tunnel, and should never be trusted as anyone's rate-limit key.
func clientIP(r *http.Request) string {
	v := r.Header.Get("CF-Connecting-IP")
	if v == "" {
		return ""
	}
	// CF-Connecting-IP is a bare address, no port — but validate rather than
	// trust the shape blindly, since a bare non-IP string would otherwise
	// silently become every such caller's shared rate-limit bucket.
	if net.ParseIP(v) == nil {
		return ""
	}
	return v
}
