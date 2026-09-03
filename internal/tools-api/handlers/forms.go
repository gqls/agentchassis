package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gqls/agentchassis/internal/tools-api/clientip"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/gqls/agentchassis/platform/httpguard"
)

// FormInboxStore is the seam the forms handlers depend on, so tests substitute
// a recorder rather than reaching for a database.
type FormInboxStore interface {
	Insert(ctx context.Context, r store.InboxRow) (string, error)
	ClaimPending(ctx context.Context, limit int, now time.Time) ([]store.InboxRow, error)
}

// Reserved field names. Everything a form posts that is NOT one of these
// becomes the payload, so a site can ask for whatever its brief needs without
// this service knowing anything about it.
//
// company_website and _elapsed are spelled exactly as the gripper spells them
// (handlers/gripper.go), because they feed the same httpguard.CheckIntake and a
// second spelling for one concept is how two paths diverge quietly.
const (
	fieldToken    = "_token"
	fieldIntent   = "_intent"
	fieldNext     = "_next"
	fieldElapsed  = "_elapsed"
	fieldHoneypot = "company_website"
)

// Bounds on what becomes a payload. The body cap (InputCapMiddleware) already
// bounds the total; these bound its SHAPE, so 32 KB cannot arrive as 4,000
// one-character fields that then have to be stored, pulled, and rendered into
// an email.
const (
	maxPayloadFields   = 64
	maxPayloadValueLen = 8192
	minTokenLen        = 32
	maxTokenLen        = 256
	maxIntentLen       = 40
)

// formAcceptedBody is THE success body for the JSON path, byte-identical for a
// human and a bot. A distinguishable rejection tells a spammer which gate to
// tune — httpguard.CheckIntake's own doc, and the gripper's production
// experience it was ported from.
var formAcceptedBody = []byte(`{"accepted":true}` + "\n")

// FormSubmitHandler handles POST /submit: the bot gates, then a shape check on
// the two values the cluster will need, then storage. It resolves nothing and
// sends nothing — see store.FormInbox's package comment for why that is the
// design and not an omission.
func FormSubmitHandler(st FormInboxStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		fields, wantsRedirect, err := readSubmission(c)
		if err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "invalid request body")
			return
		}

		// Bot gates FIRST, before any validation that could produce a
		// distinguishable 400. A bot verdict gets the success response and
		// nothing is stored — only the shape is logged.
		verdict := httpguard.CheckIntake(fields[fieldHoneypot], fields[fieldElapsed], httpguard.DefaultMinFill)
		if verdict.Bot {
			log.Printf("forms/submit: DROPPED reason=%s", verdict.Reason)
			writeFormAccepted(c, wantsRedirect, fields[fieldNext])
			return
		}

		token := strings.TrimSpace(fields[fieldToken])
		if len(token) < minTokenLen || len(token) > maxTokenLen {
			// Shape only. This process cannot tell a real token from a
			// well-formed invented one — it has nothing to check against — so
			// this refusal leaks nothing a viewer of the page's source does not
			// already know. Authenticity is the cluster's decision at ingest.
			httperr.JSONError(c, http.StatusBadRequest, "this form is not configured correctly")
			return
		}

		intent := strings.ToLower(strings.TrimSpace(fields[fieldIntent]))
		if intent == "" {
			intent = "enquiry"
		}
		if !validIntent(intent) {
			httperr.JSONError(c, http.StatusBadRequest, "this form is not configured correctly")
			return
		}

		payload, err := json.Marshal(payloadFrom(fields))
		if err != nil {
			logGripperFailure("forms/submit", "marshal_payload", "", err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not accept the submission")
			return
		}

		row := store.InboxRow{
			Token:      token,
			Intent:     intent,
			Payload:    payload,
			SiteDomain: c.GetString("site_domain"),
			IPHash:     hashIP(clientip.From(c)),
			UserAgent:  c.GetHeader("User-Agent"),
		}
		// Origin-derived, from the island's own minimal sites mirror. Recorded
		// as a cross-check the collector can compare against the token's site,
		// never as identity.
		if sid := c.GetString("site_id"); sid != "" {
			row.SiteID = &sid
		}

		id, err := st.Insert(c.Request.Context(), row)
		if err != nil {
			logGripperFailure("forms/submit", "insert", "", err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not accept the submission")
			return
		}

		// Structural log only: the id, the intent and the field count. Never a
		// field name, never a value — the payload is a stranger's personal data
		// and this is a shared log.
		log.Printf("forms/submit: STORED id=%s intent=%s fields=%d", id, intent, len(payloadFrom(fields)))
		writeFormAccepted(c, wantsRedirect, fields[fieldNext])
	}
}

// FormRequestsHandler handles GET /requests: the cluster's collector claims a
// batch. Gated by X-Internal-Key (middleware.InternalKey), NOT by CORS — it
// arrives with no Origin header, exactly like the gripper's /requests.
func FormRequestsHandler(st FormInboxStore, maxBatch int) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := st.ClaimPending(c.Request.Context(), maxBatch, time.Now().UTC())
		if err != nil {
			logGripperFailure("forms/requests", "claim_pending", "", err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not read the inbox")
			return
		}
		if rows == nil {
			rows = []store.InboxRow{}
		}
		log.Printf("forms/requests: CLAIMED n=%d", len(rows))
		c.JSON(http.StatusOK, gin.H{"submissions": rows, "count": len(rows)})
	}
}

// readSubmission accepts both encodings a static page can produce, and reports
// which one, because that decides the success response.
//
//   - application/x-www-form-urlencoded (or multipart): a PLAIN form post, so
//     the browser NAVIGATES and expects a document. It needs no JavaScript,
//     which is the whole reason to support it — a form that only works with JS
//     is a form that silently fails for some visitors.
//   - anything else: parsed as JSON, for the progressive-enhancement path.
//
// Values are flattened to strings: this service does not interpret the payload,
// and a string map is what keeps that true.
func readSubmission(c *gin.Context) (map[string]string, bool, error) {
	ct := c.ContentType()
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data") {
		if err := c.Request.ParseForm(); err != nil {
			return nil, true, err
		}
		out := make(map[string]string, len(c.Request.PostForm))
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				out[k] = v[0]
			}
		}
		return out, true, nil
	}

	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		return nil, false, err
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		switch t := v.(type) {
		case string:
			out[k] = t
		case nil:
			out[k] = ""
		default:
			// A number, bool or nested object arrives as its JSON text. The
			// cluster renders these into an email; keeping them as text means
			// there is exactly one representation to reason about.
			if b, err := json.Marshal(t); err == nil {
				out[k] = string(b)
			}
		}
	}
	return out, false, nil
}

// payloadFrom is everything that is not a reserved field, bounded in shape.
func payloadFrom(fields map[string]string) map[string]string {
	out := make(map[string]string, len(fields))
	for k, v := range fields {
		switch k {
		case fieldToken, fieldIntent, fieldNext, fieldElapsed, fieldHoneypot:
			continue
		}
		if len(out) >= maxPayloadFields {
			break
		}
		if len(v) > maxPayloadValueLen {
			v = v[:maxPayloadValueLen]
		}
		out[k] = v
	}
	return out
}

// validIntent mirrors the CHECK constraint on both tables
// (^[a-z][a-z0-9_]{1,39}$). Checked here as well as there so a bad value is a
// 400 the site's author can act on, rather than a 500 from a constraint.
func validIntent(s string) bool {
	if len(s) < 2 || len(s) > maxIntentLen {
		return false
	}
	if s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		ch := s[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' {
			continue
		}
		return false
	}
	return true
}

// writeFormAccepted answers a plain form post with a redirect the browser can
// follow, and a JSON post with the flat success body.
func writeFormAccepted(c *gin.Context, redirect bool, next string) {
	if redirect {
		if target, ok := safeRedirect(c.GetString("site_domain"), next); ok {
			c.Redirect(http.StatusSeeOther, target)
			return
		}
	}
	c.Data(http.StatusCreated, "application/json; charset=utf-8", formAcceptedBody)
}

// safeRedirect builds the thank-you URL from the domain the CORS layer
// RESOLVED, never from anything in the request body.
//
// # WHY NOT JUST HONOUR _next
//
// Because "_next" is attacker-controlled, and a redirector that follows it is
// an open redirect: post to our endpoint with _next=https://evil.example and we
// hand the visitor a link that starts on a domain they trust. So _next may only
// ever be a PATH, and the host comes from site_domain, which was resolved
// against the island's own sites table and cannot be chosen by the caller.
//
// A path is rejected unless it starts with a single "/" and contains no
// backslash or control character.
//
// # WHAT THE "//" GUARD IS AND IS NOT
//
// Stated precisely, because an overstated security rationale gets a guard
// trusted for the wrong reason. Since the host is ALWAYS prefixed here,
// "//evil.example" yields "https://<resolved-domain>//evil.example" — host
// unchanged, evil.example demoted to a path segment. So it is NOT an open
// redirect today, and this guard is not what prevents one; prefixing the
// resolved host is.
//
// It is kept for two smaller reasons. It refuses a value whose AUTHOR plainly
// meant another origin, rather than silently sending the visitor to a 404 on a
// doubled path. And it fails safe under the refactor that would make it matter:
// the moment anyone uses `next` as a whole URL instead of a path — a one-line
// change that looks harmless — a protocol-relative value becomes a real
// cross-origin redirect. The test names both halves.
func safeRedirect(siteDomain, next string) (string, bool) {
	siteDomain = strings.TrimSpace(siteDomain)
	if siteDomain == "" || strings.ContainsAny(siteDomain, "/\\ ") {
		return "", false
	}

	path := strings.TrimSpace(next)
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "", false
	}
	if strings.ContainsAny(path, "\\") {
		return "", false
	}
	for _, r := range path {
		if r < 0x20 || r == 0x7f {
			return "", false
		}
	}
	return "https://" + siteDomain + path, true
}
