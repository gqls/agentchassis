package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gqls/agentchassis/internal/tools-api/clientip"
	"github.com/gqls/agentchassis/internal/tools-api/config"
	"github.com/gqls/agentchassis/internal/tools-api/gripper"
	"github.com/gqls/agentchassis/internal/tools-api/httperr"
	"github.com/gqls/agentchassis/internal/tools-api/store"
	"github.com/gqls/agentchassis/platform/aiservice"
	"github.com/gqls/agentchassis/platform/httpguard"
)

// The gripper-dossier intake: four handlers under /api/v1/tools/gripper.
//
//	POST /session   → {session_id, greeting}                 browser, CORS
//	POST /chat      → {reply, spec, missing_fields, complete} browser, CORS
//	POST /submit    → 201 {"accepted":true}                   browser, CORS
//	GET  /requests  → NDJSON feed for the cluster              X-Internal-Key, NO CORS
//
// Behaviour is DESIGN_2026-07-24_gripper_dossier_pilot.md §2 as corrected by
// PROPOSAL_2026-08-05_gripper_route_group_in_tools_api.md; the field vocabulary
// is the cluster's (see package gripper). The handlers are thin: every rule
// that can be a pure function lives in package gripper, every state change is
// a guarded write in package store, and this file only maps HTTP to those.

// GripperStore is what the handlers need from persistence; *store.Gripper is
// the production implementation and tests substitute an in-memory one.
type GripperStore interface {
	CreateSession(ctx context.Context, siteID, ipHash, userAgent string) (string, error)
	ClaimTurn(ctx context.Context, sessionID, siteID string) (*store.Session, error)
	ClaimDailyTurn(ctx context.Context, day time.Time, cap int) error
	RecordTurn(ctx context.Context, sessionID, visitor, reply string, turnSpec gripper.Spec, inTok, outTok int) (gripper.Spec, error)
	CreateRequestFromSession(ctx context.Context, siteID, sessionID, email, ipHash, userAgent string) (string, error)
	CreateRequestInline(ctx context.Context, siteID, email string, spec gripper.Spec, ipHash, userAgent string) (string, error)
	PendingSince(ctx context.Context, since *time.Time, limit int) ([]store.FeedRow, error)
	MarkPulled(ctx context.Context, ids []string, now time.Time) error
}

// TextGenerator is the one aiservice method the chat needs.
type TextGenerator interface {
	GenerateText(ctx context.Context, prompt string, options map[string]interface{}) (string, error)
}

// ChatGenerator constructs the LLM client per call (the gauntlet handlers do
// the same — NewAnthropicClient is cheap and reads the key by env name).
type ChatGenerator func(ctx context.Context) (TextGenerator, error)

// AnthropicChatGenerator is the production ChatGenerator for cfg.
func AnthropicChatGenerator(cfg *config.GripperConfig) ChatGenerator {
	return func(ctx context.Context) (TextGenerator, error) {
		return aiservice.NewAnthropicClient(ctx, map[string]interface{}{
			"model":           cfg.Model,
			"api_key_env_var": cfg.APIKeyEnvVar,
		})
	}
}

// chatCallTimeout bounds one model call from the visitor's request context.
// DESIGN §2 says 25s; aiservice's own client timeout is 600s and would leave a
// browser hanging on a stalled call.
const chatCallTimeout = 25 * time.Second

// feedLimit caps one /requests response. The cluster polls on a schedule and
// dedups by id, so a large backlog simply drains over several pulls.
const feedLimit = 500

// ── /session ─────────────────────────────────────────────────────────────────

// GripperSessionHandler handles POST /session: opens a session for the site
// the CORS layer resolved and returns the fixed greeting. No LLM call.
func GripperSessionHandler(st GripperStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		siteID := c.GetString("site_id")
		id, err := st.CreateSession(c.Request.Context(), siteID, hashIP(clientip.From(c)), c.GetHeader("User-Agent"))
		if err != nil {
			logGripperFailure("session", "create_session", "", err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not open a session")
			return
		}
		c.JSON(http.StatusOK, gin.H{"session_id": id, "greeting": gripper.Greeting})
	}
}

// ── /chat ────────────────────────────────────────────────────────────────────

type chatRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
}

// GripperChatHandler handles POST /chat: one visitor turn in, one assistant
// turn out, spec merged server-side.
func GripperChatHandler(st GripperStore, cfg *config.GripperConfig, gen ChatGenerator) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req chatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		if _, err := uuid.Parse(req.SessionID); err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "session_id is not a valid id")
			return
		}
		if !gripper.ValidMessage(req.Message) {
			httperr.JSONError(c, http.StatusBadRequest,
				fmt.Sprintf("message is required and must be at most %d characters", gripper.MaxMessageRunes))
			return
		}
		message := strings.TrimSpace(req.Message)
		siteID := c.GetString("site_id")

		ctx, cancel := context.WithTimeout(c.Request.Context(), chatCallTimeout)
		defer cancel()

		sess, err := st.ClaimTurn(ctx, req.SessionID, siteID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrSessionNotFound):
				httperr.JSONError(c, http.StatusNotFound, "session not found")
			case errors.Is(err, store.ErrSessionClosed):
				httperr.JSONError(c, http.StatusConflict, "this session has already been submitted or has expired")
			case errors.Is(err, store.ErrSessionCapped):
				httperr.JSONError(c, http.StatusConflict, "this session has reached its limit — please submit what you have or start again")
			default:
				logGripperFailure("chat", "claim_turn", req.SessionID, err)
				httperr.JSONError(c, http.StatusInternalServerError, "session lookup failed")
			}
			return
		}
		if err := st.ClaimDailyTurn(ctx, time.Now().UTC(), cfg.DailyTurnCap); err != nil {
			if errors.Is(err, store.ErrDailyCapReached) {
				logGripperFailure("chat", "daily_cap", req.SessionID, err)
				httperr.JSONError(c, http.StatusConflict, "the assistant has reached today's limit — please try again tomorrow")
				return
			}
			logGripperFailure("chat", "daily_cap", req.SessionID, err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not record the turn")
			return
		}

		client, err := gen(ctx)
		if err != nil {
			logGripperAIFailure("chat", "client_init", req.SessionID, err)
			httperr.JSONError(c, http.StatusServiceUnavailable, "intake assistant unavailable")
			return
		}
		prompt := gripper.BuildPrompt(sess.Spec, sess.Transcript, message)
		opts := map[string]interface{}{"max_tokens": gripper.MaxTokensPerReply}
		text, err := client.GenerateText(ctx, prompt, opts)
		if err != nil {
			// 503, not 502: Cloudflare replaces an origin 502's body with its
			// own page and the JSON error shape never reaches the widget
			// (commit b498df16b, the gauntlet's lesson).
			logGripperAIFailure("chat", "generate", req.SessionID, err)
			httperr.JSONError(c, http.StatusServiceUnavailable, "intake assistant unavailable")
			return
		}
		rep, err := gripper.ParseReply(text)
		if err != nil {
			logGripperBadReply("chat", err.Error(), req.SessionID, text)
			httperr.JSONError(c, http.StatusServiceUnavailable, "intake assistant response was invalid")
			return
		}

		merged, err := st.RecordTurn(ctx, req.SessionID, message, *rep.Reply,
			gripper.Normalise(*rep.Spec), usageInt(opts, "__usage_input_tokens"), usageInt(opts, "__usage_output_tokens"))
		if err != nil {
			logGripperFailure("chat", "record_turn", req.SessionID, err)
			httperr.JSONError(c, http.StatusInternalServerError, "could not record the turn")
			return
		}
		missing := gripper.Missing(merged)
		c.JSON(http.StatusOK, gin.H{
			"reply":          *rep.Reply,
			"spec":           merged,
			"missing_fields": missing,
			"complete":       len(missing) == 0,
		})
	}
}

func usageInt(opts map[string]interface{}, key string) int {
	if v, ok := opts[key]; ok {
		if n, ok := v.(int); ok {
			return n
		}
	}
	return 0
}

// ── /submit ──────────────────────────────────────────────────────────────────

type submitRequest struct {
	SessionID      *string                `json:"session_id"`
	Email          string                 `json:"email"`
	CompanyWebsite string                 `json:"company_website"` // honeypot: humans never see it
	Elapsed        json.RawMessage        `json:"_elapsed"`        // ms on page, client-side delta
	Spec           map[string]interface{} `json:"spec"`            // plain-form (degraded) mode only
}

// acceptedBody is THE /submit success response. One byte slice, written for a
// human and for a bot alike, because a distinguishable rejection tells a
// spammer which gate to tune (httpguard.CheckIntake's own doc, and idea.uk's
// production experience it was ported from). No request id in it: the email
// carries the link, so the browser needs nothing back but "accepted".
var acceptedBody = []byte(`{"accepted":true}` + "\n")

func writeAccepted(c *gin.Context) {
	c.Data(http.StatusCreated, "application/json; charset=utf-8", acceptedBody)
}

// GripperSubmitHandler handles POST /submit: the honeypot/timing gate, then
// the email, then a report request filed from the session's spec (or from an
// inline spec in the widget's degraded plain-form mode).
func GripperSubmitHandler(st GripperStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req submitRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			httperr.JSONError(c, http.StatusBadRequest, "invalid request body")
			return
		}

		// Bot gates FIRST, before any validation that could produce a
		// distinguishable 400. A bot verdict gets the success body and
		// nothing is stored — only the shape is logged.
		verdict := httpguard.CheckIntake(req.CompanyWebsite, elapsedString(req.Elapsed), httpguard.DefaultMinFill)
		if verdict.Bot {
			log.Printf("gripper/submit: DROPPED reason=%s", verdict.Reason)
			writeAccepted(c)
			return
		}

		email, ok := gripper.ValidEmail(req.Email)
		if !ok {
			httperr.JSONError(c, http.StatusBadRequest, "a valid email address is required")
			return
		}
		siteID := c.GetString("site_id")
		ipHash := hashIP(clientip.From(c))
		ua := c.GetHeader("User-Agent")
		ctx := c.Request.Context()

		var id string
		var err error
		if req.SessionID != nil && strings.TrimSpace(*req.SessionID) != "" {
			if _, perr := uuid.Parse(*req.SessionID); perr != nil {
				httperr.JSONError(c, http.StatusBadRequest, "session_id is not a valid id")
				return
			}
			id, err = st.CreateRequestFromSession(ctx, siteID, *req.SessionID, email, ipHash, ua)
		} else {
			spec := gripper.Normalise(req.Spec)
			if missing := gripper.Missing(spec); len(missing) > 0 {
				httperr.JSONError(c, http.StatusBadRequest, "spec is missing required fields: "+strings.Join(missing, ", "))
				return
			}
			id, err = st.CreateRequestInline(ctx, siteID, email, spec, ipHash, ua)
		}
		if err != nil {
			switch {
			case errors.Is(err, store.ErrSessionNotFound):
				httperr.JSONError(c, http.StatusNotFound, "session not found")
			case errors.Is(err, store.ErrSessionClosed):
				httperr.JSONError(c, http.StatusConflict, "this session has already been submitted")
			case errors.Is(err, store.ErrSpecIncomplete):
				httperr.JSONError(c, http.StatusBadRequest, "the spec is not complete yet — answer the remaining questions first")
			default:
				logGripperFailure("submit", "create_request", "", err)
				httperr.JSONError(c, http.StatusInternalServerError, "could not file the request")
			}
			return
		}
		// Structural log only: the id and which fields are set, never the
		// email or any spec text.
		log.Printf("gripper/submit: FILED request_id=%s", id)
		writeAccepted(c)
	}
}

// elapsedString accepts _elapsed as a JSON number or string; CheckIntake
// fails open on anything it cannot parse, so a missing value is "no JS".
func elapsedString(raw json.RawMessage) string {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return fmt.Sprintf("%g", f)
	}
	return string(raw)
}

// ── /requests ────────────────────────────────────────────────────────────────

// GripperRequestsHandler handles GET /requests?since=<RFC3339> for the
// cluster's pull_report_requests action. It writes exactly what that action's
// parser reads (report_request_pull_action.go: pulledReportRequest): one
// object per line {id, host, submitted_at, spec}, then one {"_meta":{...}}
// line, Content-Type application/x-ndjson, status 200 or nothing.
//
// submitted_at is created_at rendered RFC3339 in UTC with a Z: the cluster
// casts it ::timestamptz to derive its checkpoint, and one malformed value in
// any historical row would break that query for the site permanently.
func GripperRequestsHandler(st GripperStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var since *time.Time
		if raw := c.Query("since"); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				httperr.JSONError(c, http.StatusBadRequest, "since must be RFC3339")
				return
			}
			since = &t
		}
		ctx := c.Request.Context()
		rows, err := st.PendingSince(ctx, since, feedLimit)
		if err != nil {
			logGripperFailure("requests", "pending_since", "", err)
			httperr.JSONError(c, http.StatusInternalServerError, "feed unavailable")
			return
		}
		now := time.Now().UTC()
		var buf bytes.Buffer
		ids := make([]string, 0, len(rows))
		for _, r := range rows {
			line, err := json.Marshal(map[string]interface{}{
				"id":           r.ID,
				"host":         r.Host,
				"submitted_at": r.CreatedAt.UTC().Format(time.RFC3339),
				"spec":         r.Spec,
			})
			if err != nil {
				logGripperFailure("requests", "marshal_line", r.ID, err)
				httperr.JSONError(c, http.StatusInternalServerError, "feed unavailable")
				return
			}
			buf.Write(line)
			buf.WriteByte('\n')
			ids = append(ids, r.ID)
		}
		meta, _ := json.Marshal(map[string]interface{}{
			"_meta": map[string]interface{}{
				"count":       len(ids),
				"truncated":   len(rows) == feedLimit,
				"server_time": now.Format(time.RFC3339),
			},
		})
		buf.Write(meta)
		buf.WriteByte('\n')
		c.Data(http.StatusOK, "application/x-ndjson", buf.Bytes())

		// After the write, so a request that failed to serialise is not marked
		// pulled. A pull the cluster then failed to store is harmless: pulled
		// rows are served again next time.
		if err := st.MarkPulled(ctx, ids, now); err != nil {
			logGripperFailure("requests", "mark_pulled", "", err)
		}
		log.Printf("gripper/requests: served=%d since=%v", len(ids), since != nil)
	}
}

// ── logging ──────────────────────────────────────────────────────────────────
//
// Same three shapes as ailog.go's gauntlet helpers and for the same reasons
// (bugs_open/083: an error that is not logged cannot be diagnosed; a model
// response is logged as a FINGERPRINT, never as text — owner ruling
// 2026-07-27). Prefixed gripper/ so the two tools' lines sort apart in the
// island's one shared log stream.

func logGripperFailure(endpoint, stage, id string, err error) {
	log.Printf("gripper/%s: %s FAILED id=%s err=%v", endpoint, stage, id, err)
}

func logGripperAIFailure(endpoint, stage, id string, err error) {
	if te, ok := aiservice.IsTruncated(err); ok {
		log.Printf("gripper/%s: %s TRUNCATED id=%s provider=%s reason=%s output_tokens=%d partial_chars=%d",
			endpoint, stage, id, te.Provider, te.Reason, te.OutputTokens, len(te.Partial))
		return
	}
	log.Printf("gripper/%s: %s FAILED id=%s err=%v", endpoint, stage, id, err)
}

func logGripperBadReply(endpoint, reason, id, body string) {
	log.Printf("gripper/%s: response UNUSABLE id=%s reason=%s %s", endpoint, id, reason, aiservice.Fingerprint(body))
}
