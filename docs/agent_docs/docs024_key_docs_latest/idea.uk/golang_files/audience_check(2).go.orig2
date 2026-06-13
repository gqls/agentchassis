package main

// audience_check.go — the free public taster endpoint.
//
// The page's taster widget POSTs {business, audience} to /audience-check and
// drops the HTML response into a result div. Cost ~£0.02 per call (one Opus
// call, ~1.5k input + 500 output tokens). Rate-limited per IP to bound abuse;
// disable entirely with TASTER_ENABLED=false if it ever needs to be turned off.
//
// This is the only endpoint on the service that's open to the world without
// either operator auth or a paid order. Everything else (request → confirm →
// pay → fulfil) still funnels through the existing gates.

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ── per-IP rate limiter ──────────────────────────────────────────────────────
//
// Simple sliding-window counter. Two bands: short (per-hour) and long (per-day),
// because three-per-hour alone lets an attacker do 72/day from one IP. Both
// bands must clear for the request to proceed. In-memory only; on restart we
// lose state, which is fine for this risk profile.

type rateBand struct {
	window time.Duration
	max    int
}

type rateLimiter struct {
	mu    sync.Mutex
	hits  map[string][]time.Time
	bands []rateBand
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		hits: map[string][]time.Time{},
		bands: []rateBand{
			{window: time.Hour, max: 3},
			{window: 24 * time.Hour, max: 20},
		},
	}
}

// allow returns (ok, retryAfter). retryAfter is the earliest moment the IP
// could try again if it's currently throttled; zero when allowed.
func (rl *rateLimiter) allow(ip string) (bool, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	// Drop anything older than the longest band.
	longest := rl.bands[0].window
	for _, b := range rl.bands {
		if b.window > longest {
			longest = b.window
		}
	}
	cutoff := now.Add(-longest)
	hits := rl.hits[ip]
	pruned := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}
	// Check each band.
	for _, b := range rl.bands {
		bandCutoff := now.Add(-b.window)
		count := 0
		var oldestInBand time.Time
		for _, t := range pruned {
			if t.After(bandCutoff) {
				if count == 0 || t.Before(oldestInBand) {
					oldestInBand = t
				}
				count++
			}
		}
		if count >= b.max {
			retry := oldestInBand.Add(b.window).Sub(now)
			rl.hits[ip] = pruned
			return false, retry
		}
	}
	pruned = append(pruned, now)
	rl.hits[ip] = pruned
	return true, 0
}

// clientIP extracts a stable per-client key. Trusts X-Forwarded-For when
// present (the service is expected to sit behind a TLS-terminating proxy in
// production). Falls back to RemoteAddr's host part.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// XFF is comma-separated; first entry is the original client.
		if i := strings.IndexByte(xff, ','); i > 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	addr := r.RemoteAddr
	if i := strings.LastIndexByte(addr, ':'); i > 0 {
		return addr[:i]
	}
	return addr
}

// ── handler ──────────────────────────────────────────────────────────────────

const tasterEnabledEnv = "TASTER_ENABLED"

func (a *App) audienceCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	// Quick kill switch via env var. If set to "false" or "0", returns 503.
	if v := envFallback(tasterEnabledEnv, "true"); v == "false" || v == "0" {
		http.Error(w, "Taster is temporarily unavailable. Please use the request form below.", http.StatusServiceUnavailable)
		return
	}
	// Rate limit.
	if ok, retry := a.taster.allow(clientIP(r)); !ok {
		mins := int((retry + 59*time.Second) / time.Minute)
		w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retry.Seconds())))
		writeHTML(w, fmt.Sprintf(
			`<h4>Have another go in a bit</h4><p>You've used your free tasters for now — try again in about %d minute(s), or jump straight to the £29 full report below.</p>`,
			mins))
		return
	}
	// Parse form.
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form", http.StatusBadRequest)
		return
	}
	business := strings.TrimSpace(r.FormValue("business"))
	audience := strings.TrimSpace(r.FormValue("audience"))
	if business == "" || audience == "" {
		http.Error(w, "Both fields are required", http.StatusBadRequest)
		return
	}
	if len(business) > 200 || len(audience) > 200 {
		http.Error(w, "Please keep each field under 200 characters", http.StatusBadRequest)
		return
	}
	// Run the audience step. Empty assets — the public taster doesn't ask for
	// them; the audience challenge alone is the value.
	aud, err := a.audience(business, audience, "")
	if err != nil {
		writeHTML(w, `<h4>We couldn't run the check just now</h4><p>Sorry — try again in a minute, or skip straight to the full report below.</p>`)
		return
	}
	log.Printf("free taster: business=%q audience=%q", business, audience)
	writeHTML(w, renderAudienceHTML(business, audience, aud))
}

// renderAudienceHTML produces the HTML fragment the taster widget drops into
// its result div. Uses the existing .taster-result styles on the page.
// All user-supplied strings are HTML-escaped; model-supplied strings are
// trusted as plain text (no model is supposed to emit HTML, but escaping is
// defensive — a future model that hallucinates a <script> tag should not run
// in the visitor's browser).
func renderAudienceHTML(business, statedAudience string, aud audienceResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "<h4>Your stated audience</h4><p>%s</p>", html.EscapeString(statedAudience))
	if aud.CarriedAudience != "" {
		fmt.Fprintf(&b, "<h4>Who we think you should actually be selling to</h4><p>%s</p>",
			html.EscapeString(aud.CarriedAudience))
	}
	if aud.WillingnessToPay != "" {
		fmt.Fprintf(&b, "<h4>Why</h4><p>%s</p>", html.EscapeString(aud.WillingnessToPay))
	}
	if len(aud.Alternatives) > 0 {
		b.WriteString("<h4>Three other audiences worth considering</h4><ol>")
		for _, a := range aud.Alternatives {
			fmt.Fprintf(&b, "<li><strong>%s</strong> — %s</li>",
				html.EscapeString(a.Audience), html.EscapeString(a.Why))
		}
		b.WriteString("</ol>")
	}
	b.WriteString(`<div class="taster-upsell">` +
		`<p><strong>That's the free part — the audience check.</strong> It tells you who you'd be best off selling to.</p>` +
		`<p>The <strong>£29 report</strong> gives you the ideas themselves: three to six specific things you could build for that audience, ranked best first. For each one it spells out, in plain terms, what it would do, what it would use from your own business, why it's better than just typing the question into a free chatbot, and the cheapest way to find out whether people will actually pay for it.</p>` +
		`<p>We check the facts behind every idea — who your real competitors are, what they charge, the rules you'd have to follow — and throw out any that don't stand up. And if there's nothing worth building in your area yet, we tell you that plainly, with the reason.</p>` +
		`<p>It's a written report, sent by email, usually within a few days. You only pay once we've confirmed we can do a useful job for you — and if the report doesn't turn up anything worth acting on, you get your money back.</p>` +
		`<p><a href="#request" class="taster-cta">Request the full report — £29 →</a></p>` +
		`</div>`)
	return b.String()
}

// envFallback is a tiny helper used here and (eventually) elsewhere — pulls
// env with a default. Mirrors the existing env() helper in engine.go but takes
// the default as the second arg the conventional way.
func envFallback(key, def string) string {
	v := envOrEmpty(key)
	if v == "" {
		return def
	}
	return v
}

// envOrEmpty: small indirection so tests can shim it if ever needed.
var envOrEmpty = func(key string) string {
	// Calling os.Getenv directly would import "os" again; instead piggyback on
	// the existing env() helper in engine.go which already uses os.Getenv.
	// env() takes (key, default) and returns default when unset, so use a
	// sentinel value that's improbable in env vars.
	const sentinel = "__UNSET__"
	v := env(key, sentinel)
	if v == sentinel {
		return ""
	}
	return v
}
