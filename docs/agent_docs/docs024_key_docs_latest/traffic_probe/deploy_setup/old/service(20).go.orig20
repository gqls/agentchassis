package main

// service.go — site-engine: the capture backend for VM-hosted backend sites (API only).
//
// First feature: visitor-intent capture for traffic-probe sites; the engine is
// the class-level backend and grows by feature (e.g. chat, boards) later.
//
// Division of labour after the chassis reconciliation:
//   - The per-domain PAGE (tagline, the one invited action, the privacy line) is
//     a normal chassis-built static site, deployed to the VM's web root.
//   - nginx serves those static files from disk and proxies ONLY the capture API
//     paths below to this engine.
//   - This engine does the one thing nginx/static cannot: record a structured
//     intent event server-side, keyed by Host, into the JSON store.
//
// There is no page rendering and no per-domain content registry here anymore
// (the chassis owns both). Kept from the idea.uk fork: the App/routes/cors shape,
// writeJSON, the store pattern.
//
// Endpoints (all proxied by nginx; the engine is not exposed directly):
//   POST /intent    capture one intent event, then 303 to a thanks path
//   GET  /api/hit   1x1 visit beacon (no cookie, no JS) -> visit denominator
//   GET  /stats     key-gated per-host summary
//   GET  /health    liveness

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"time"
)

// ProbeKind values — the kind of invited action recorded on an event. The page
// (built by the chassis) sets these in its form's hidden "kind" field.
const (
	ProbeSearch   = "search"
	ProbeCategory = "categories"
	ProbeFreeText = "freetext"
)

type Config struct {
	InternalKey  string          // INTERNAL_API_KEY — gates /stats
	GeoHeader    string          // header carrying a coarse country code (e.g. CF-IPCountry), or ""
	MaxValueLen  int             // cap on a captured value's length
	ThanksPath   string          // where /intent redirects on success (a static page on the site)
	AcceptHosts  map[string]bool // optional allow-list of canonical hosts this VM serves; empty = accept any
	AllowOrigins []string        // CORS allow-list (kept for parity)
}

type App struct {
	cfg   Config
	store *Store
}

func NewApp(cfg Config, store *Store) *App { return &App{cfg: cfg, store: store} }

func newEventID() string { return "ev_" + fmt.Sprintf("%d", time.Now().UnixNano()) }

// canonicalHost normalises a request Host into the store key: lower-case, port
// stripped, leading "www." dropped. Single source of truth for the key shape.
func canonicalHost(h string) string {
	h = strings.ToLower(strings.TrimSpace(h))
	if i := strings.IndexByte(h, ':'); i >= 0 {
		h = h[:i]
	}
	return strings.TrimPrefix(h, "www.")
}

// accepted reports whether this VM should record events for a host. With an empty
// allow-list we accept any (the manual-phase default); set ACCEPT_HOSTS once the
// domain->VM registry is authoritative so stray hosts do not create junk rows.
func (a *App) accepted(host string) bool {
	if len(a.cfg.AcceptHosts) == 0 {
		return true
	}
	return a.cfg.AcceptHosts[host]
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.health)
	mux.HandleFunc("/stats", a.stats)
	mux.HandleFunc("/events", a.events)
	mux.HandleFunc("/intent", a.intent)
	mux.HandleFunc("/api/hit", a.hit)
	return a.cors(mux)
}

// intent records one event from a probe submission then 303-redirects to the
// thanks path on the same site. Empty submissions and unaccepted hosts are not
// stored. Pure form POST — no JS, no cookies.
func (a *App) intent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	_ = r.ParseForm()
	host := canonicalHost(r.Host)

	kind := r.FormValue("kind")
	switch kind {
	case ProbeSearch, ProbeCategory, ProbeFreeText:
	default:
		kind = ProbeFreeText
	}
	value := sanitizeValue(r.FormValue("value"), a.cfg.MaxValueLen)
	if value == "" || !a.accepted(host) {
		http.Redirect(w, r, a.cfg.ThanksPath, http.StatusSeeOther)
		return
	}

	a.store.AddEvent(&IntentEvent{
		ID:        newEventID(),
		Host:      host,
		Kind:      kind,
		Value:     value,
		RefHost:   refHost(r.Header.Get("Referer"), host),
		Country:   a.country(r),
		CreatedAt: time.Now().UTC(),
	})
	http.Redirect(w, r, a.cfg.ThanksPath, http.StatusSeeOther)
}

// hit is a 1x1 visit beacon. The chassis-built page includes
// <img src="/api/hit" width="1" height="1" alt=""> so visits are counted
// server-side with no cookie and no JS (the denominator for events-per-1k).
func (a *App) hit(w http.ResponseWriter, r *http.Request) {
	host := canonicalHost(r.Host)
	if a.accepted(host) {
		a.store.AddVisit(host)
	}
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte{ // 1x1 transparent GIF
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00,
		0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00,
		0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00,
		0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
	})
}

func (a *App) country(r *http.Request) string {
	if a.cfg.GeoHeader == "" {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(r.Header.Get(a.cfg.GeoHeader)))
}

// sanitizeValue cleans a captured value at the source. v2 (the v1 control-char
// strip was insufficient for terms that get stored, aggregated, and displayed):
//   - strips Cc (controls incl. \n,\t) AND Cf (format chars: zero-widths
//     U+200B…, bidi overrides U+202A–E/U+2066–9 — display-spoofing in any
//     terminal or dashboard — BOM, soft hyphen). Trade-off: stripping Cf
//     breaks ZWJ emoji sequences into components; acceptable for term analytics.
//   - collapses internal whitespace runs to a single space (and trims), so
//     "rolex   gmt" and "rolex gmt" aggregate as one term.
//   - caps by RUNES (multibyte-safe).
// Deliberately NOT here: NFC normalisation (needs golang.org/x/text — engine is
// stdlib-only) and lowercasing — both belong to the collector/analysis side,
// which normalises before aggregation. Raw-but-safe is stored; HTML escaping
// happens at display, never at capture.
func sanitizeValue(s string, maxRunes int) string {
	out := make([]rune, 0, len(s))
	pendingSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) { // FIRST: \t and \n are also Cc — they must
			pendingSpace = true //        separate words, not vanish
			continue
		}
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			continue
		}
		if pendingSpace && len(out) > 0 {
			out = append(out, ' ')
		}
		pendingSpace = false
		out = append(out, r)
		if maxRunes > 0 && len(out) >= maxRunes {
			break
		}
	}
	return string(out)
}

// refHost reduces a Referer to its bare host for privacy; blanks same-site.
func refHost(referer, selfHost string) string {
	if referer == "" {
		return ""
	}
	u, err := url.Parse(referer)
	if err != nil {
		return ""
	}
	h := canonicalHost(u.Host)
	if h == "" || h == selfHost {
		return ""
	}
	return h
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (a *App) stats(w http.ResponseWriter, r *http.Request) {
	if a.cfg.InternalKey == "" || r.Header.Get("X-Internal-Key") != a.cfg.InternalKey {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, a.store.Stats())
}

// events streams stored intent events as NDJSON for the off-box collector
// (P4). Key-gated like /stats. Params: since (RFC3339, strictly-after), host,
// limit (default 5000). The FINAL line is a _meta object — count, truncated,
// server_time — the collector's checkpoint aid.
func (a *App) events(w http.ResponseWriter, r *http.Request) {
	if a.cfg.InternalKey == "" || r.Header.Get("X-Internal-Key") != a.cfg.InternalKey {
		http.Error(w, "unauthorised", http.StatusUnauthorized)
		return
	}
	var since time.Time
	if v := r.URL.Query().Get("since"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			http.Error(w, "since must be RFC3339", http.StatusBadRequest)
			return
		}
		since = t
	}
	limit := 5000
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	n, truncated, err := a.store.StreamEvents(w, since, r.URL.Query().Get("host"), limit)
	if err != nil && n == 0 {
		http.Error(w, "read failed", http.StatusInternalServerError)
		return
	}
	meta, _ := json.Marshal(map[string]any{"_meta": map[string]any{
		"count": n, "truncated": truncated,
		"server_time": time.Now().UTC().Format(time.RFC3339Nano),
	}})
	w.Write(meta)
	w.Write([]byte("\n"))
}

func (a *App) cors(next http.Handler) http.Handler {
	allowed := map[string]bool{}
	for _, o := range a.cfg.AllowOrigins {
		allowed[o] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && (allowed["*"] || allowed[origin]) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
			w.Header().Set("Access-Control-Allow-Headers", "*")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
