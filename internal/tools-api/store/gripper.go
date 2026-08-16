package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gqls/agentchassis/internal/tools-api/gripper"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Gripper is the persistence for the gripper-dossier intake: three tables
// created by migration 436 (gripper_chat_sessions, gripper_report_requests,
// gripper_daily_turns) on the ISLAND Postgres, beside gauntlet_rounds.
//
// Every state change is a guarded UPDATE — `WHERE status = <expected>` (or a
// counter bound) — and reports whether it moved a row. Two writers touch these
// rows concurrently by design (the request handlers and the poller, plus
// overlapping poller ticks), so "did my update apply" is the return value, not
// an assumption. Nothing here overwrites blind.
type Gripper struct {
	Pool *pgxpool.Pool
}

// Session errors, distinguished so a handler can answer 404 / 409 honestly.
var (
	ErrSessionNotFound = errors.New("gripper: session not found")
	ErrSessionClosed   = errors.New("gripper: session is not active")
	ErrSessionCapped   = errors.New("gripper: session turn or token cap reached")
	ErrDailyCapReached = errors.New("gripper: global daily turn cap reached")
)

// Session is the handler's view of a gripper_chat_sessions row.
type Session struct {
	ID           string
	SiteID       string
	Spec         gripper.Spec
	Transcript   []gripper.Turn
	Turns        int
	InputTokens  int
	OutputTokens int
	Status       string
}

// ── sessions ─────────────────────────────────────────────────────────────────

// CreateSession inserts an active session for the site the CORS layer resolved.
func (g *Gripper) CreateSession(ctx context.Context, siteID, ipHash, userAgent string) (string, error) {
	const q = `
		INSERT INTO gripper_chat_sessions (site_id, client_ip_hash, user_agent)
		VALUES ($1, $2, $3)
		RETURNING id`
	var id string
	if err := g.Pool.QueryRow(ctx, q, siteID, ipHash, truncate(userAgent, 300)).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

// ClaimTurn atomically takes one turn on an active session under the caps and
// returns the session as it stood. The WHERE clause is the enforcement of
// MaxTurns and MaxSessionTokens: two concurrent /chat calls cannot both take
// the 30th turn. When nothing moved, a second read says why.
func (g *Gripper) ClaimTurn(ctx context.Context, sessionID, siteID string) (*Session, error) {
	const claim = `
		UPDATE gripper_chat_sessions
		SET    turns = turns + 1, last_activity_at = now()
		WHERE  id = $1 AND site_id = $2 AND status = 'active'
		  AND  turns < $3
		  AND  input_tokens + output_tokens < $4
		RETURNING id, site_id, spec, transcript, turns, input_tokens, output_tokens, status`
	s, err := scanSession(g.Pool.QueryRow(ctx, claim, sessionID, siteID, gripper.MaxTurns, gripper.MaxSessionTokens))
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	// Explain the refusal without leaking whether a session exists on
	// ANOTHER site: the read is site-scoped exactly as the claim was.
	const why = `
		SELECT status, turns, input_tokens + output_tokens
		FROM   gripper_chat_sessions WHERE id = $1 AND site_id = $2`
	var status string
	var turns, tokens int
	if err := g.Pool.QueryRow(ctx, why, sessionID, siteID).Scan(&status, &turns, &tokens); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	if status != "active" {
		return nil, ErrSessionClosed
	}
	return nil, ErrSessionCapped
}

// ClaimDailyTurn takes one unit of the global daily turn budget. One row per
// UTC day; the conditional upsert is the cap. Returns ErrDailyCapReached when
// the budget is spent.
func (g *Gripper) ClaimDailyTurn(ctx context.Context, day time.Time, cap int) error {
	const q = `
		INSERT INTO gripper_daily_turns (day, turns) VALUES ($1::date, 1)
		ON CONFLICT (day) DO UPDATE
		   SET turns = gripper_daily_turns.turns + 1
		 WHERE gripper_daily_turns.turns < $2
		RETURNING turns`
	var n int
	err := g.Pool.QueryRow(ctx, q, day.UTC(), cap).Scan(&n)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrDailyCapReached
	}
	return err
}

// RecordTurn appends the exchange to the transcript, merges the turn's spec
// and adds the token usage, returning the merged spec.
//
// jsonb `||` is the merge: it is commutative across two overlapping turns and
// implements "non-null never regresses" for free, because turnSpec has already
// been through gripper.Normalise and so holds no nulls to overwrite with.
func (g *Gripper) RecordTurn(ctx context.Context, sessionID, visitor, reply string, turnSpec gripper.Spec, inTok, outTok int) (gripper.Spec, error) {
	turns, err := json.Marshal([]gripper.Turn{{Role: "visitor", Text: visitor}, {Role: "assistant", Text: reply}})
	if err != nil {
		return nil, err
	}
	specJSON, err := json.Marshal(turnSpec)
	if err != nil {
		return nil, err
	}
	const q = `
		UPDATE gripper_chat_sessions
		SET    transcript    = transcript || $2::jsonb,
		       spec          = spec || $3::jsonb,
		       input_tokens  = input_tokens + $4,
		       output_tokens = output_tokens + $5,
		       last_activity_at = now()
		WHERE  id = $1
		RETURNING spec`
	var raw []byte
	if err := g.Pool.QueryRow(ctx, q, sessionID, turns, specJSON, inTok, outTok).Scan(&raw); err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return gripper.Normalise(m), nil
}

func scanSession(row pgx.Row) (*Session, error) {
	var s Session
	var spec, tr []byte
	if err := row.Scan(&s.ID, &s.SiteID, &spec, &tr, &s.Turns, &s.InputTokens, &s.OutputTokens, &s.Status); err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if len(spec) > 0 {
		if err := json.Unmarshal(spec, &m); err != nil {
			return nil, fmt.Errorf("session %s: spec column unreadable: %w", s.ID, err)
		}
	}
	s.Spec = gripper.Normalise(m)
	if len(tr) > 0 {
		if err := json.Unmarshal(tr, &s.Transcript); err != nil {
			return nil, fmt.Errorf("session %s: transcript column unreadable: %w", s.ID, err)
		}
	}
	return &s, nil
}

// ── submit ───────────────────────────────────────────────────────────────────

// CreateRequestFromSession closes the session and files a report request
// carrying the SESSION's spec — the server's copy, never the client's. Both
// writes are one transaction: a request cannot exist without its session
// having moved to submitted, and pressing submit twice cannot file twice
// (the second UPDATE matches no active row).
func (g *Gripper) CreateRequestFromSession(ctx context.Context, siteID, sessionID, email, ipHash, userAgent string) (string, error) {
	tx, err := g.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	const close = `
		UPDATE gripper_chat_sessions
		SET    status = 'submitted', last_activity_at = now()
		WHERE  id = $1 AND site_id = $2 AND status = 'active'
		RETURNING spec`
	var raw []byte
	if err := tx.QueryRow(ctx, close, sessionID, siteID).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var status string
			if e := tx.QueryRow(ctx, `SELECT status FROM gripper_chat_sessions WHERE id=$1 AND site_id=$2`, sessionID, siteID).Scan(&status); e != nil {
				return "", ErrSessionNotFound
			}
			return "", ErrSessionClosed
		}
		return "", err
	}
	var m map[string]interface{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &m); err != nil {
			return "", err
		}
	}
	spec := gripper.Normalise(m)
	if !gripper.Complete(spec) {
		return "", ErrSpecIncomplete
	}
	id, err := insertRequest(ctx, tx, siteID, &sessionID, email, spec, ipHash, userAgent)
	if err != nil {
		return "", err
	}
	return id, tx.Commit(ctx)
}

// ErrSpecIncomplete is returned when a submit arrives for a spec that does not
// hold every required field (the plain-form fallback, or a session submitted
// before the assistant said complete).
var ErrSpecIncomplete = errors.New("gripper: spec is missing required fields")

// CreateRequestInline files a request from a spec supplied directly (the
// widget's degraded plain-form mode). The spec must already be Normalised and
// Complete — the handler checks and answers 400 otherwise.
func (g *Gripper) CreateRequestInline(ctx context.Context, siteID, email string, spec gripper.Spec, ipHash, userAgent string) (string, error) {
	if !gripper.Complete(spec) {
		return "", ErrSpecIncomplete
	}
	return insertRequest(ctx, g.Pool, siteID, nil, email, spec, ipHash, userAgent)
}

type execQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func insertRequest(ctx context.Context, db execQuerier, siteID string, sessionID *string, email string, spec gripper.Spec, ipHash, userAgent string) (string, error) {
	specJSON, err := json.Marshal(gripper.ForCluster(spec))
	if err != nil {
		return "", err
	}
	const q = `
		INSERT INTO gripper_report_requests
		    (site_id, session_id, email, spec, status, expires_at, next_check_at, client_ip_hash, user_agent)
		VALUES ($1, $2, $3, $4::jsonb, 'pending', now() + $5::interval, now() + $6::interval, $7, $8)
		RETURNING id`
	var id string
	err = db.QueryRow(ctx, q, siteID, sessionID, email, specJSON,
		gripper.RequestTTL.String(), gripper.FirstCheckAfter.String(),
		ipHash, truncate(userAgent, 300)).Scan(&id)
	return id, err
}

// ── the cluster's pull feed ──────────────────────────────────────────────────

// FeedRow is one NDJSON line of GET /requests.
type FeedRow struct {
	ID        string
	Host      string
	CreatedAt time.Time
	Spec      map[string]interface{}
}

// PendingSince lists requests the cluster should have, oldest first: those
// still awaiting a report (pending or pulled) created at or after since (all
// of them when since is nil). Fulfilled/emailed rows are ones the cluster
// already built; expired/failed ones have had their apology and must not
// spawn a report nobody will be told about.
func (g *Gripper) PendingSince(ctx context.Context, since *time.Time, limit int) ([]FeedRow, error) {
	const q = `
		SELECT r.id, s.domain, r.created_at, r.spec
		FROM   gripper_report_requests r
		JOIN   sites s ON s.id = r.site_id
		WHERE  r.status IN ('pending', 'pulled')
		  AND  ($1::timestamptz IS NULL OR r.created_at >= $1)
		ORDER  BY r.created_at, r.id
		LIMIT  $2`
	rows, err := g.Pool.Query(ctx, q, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FeedRow
	for rows.Next() {
		var fr FeedRow
		var raw []byte
		if err := rows.Scan(&fr.ID, &fr.Host, &fr.CreatedAt, &raw); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(raw, &fr.Spec); err != nil {
			return nil, fmt.Errorf("request %s: spec column unreadable: %w", fr.ID, err)
		}
		out = append(out, fr)
	}
	return out, rows.Err()
}

// MarkPulled records that the cluster has been handed these ids. pending →
// pulled on the first pull; a later status is never regressed.
func (g *Gripper) MarkPulled(ctx context.Context, ids []string, now time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	const q = `
		UPDATE gripper_report_requests
		SET    first_pulled_at = COALESCE(first_pulled_at, $2),
		       last_pulled_at  = $2,
		       status          = CASE WHEN status = 'pending' THEN 'pulled' ELSE status END
		WHERE  id = ANY($1::uuid[])`
	_, err := g.Pool.Exec(ctx, q, ids, now)
	return err
}

// ── poller (gripper.RequestStore) ────────────────────────────────────────────

const requestCols = `r.id, s.domain, COALESCE(r.email,''), r.status, r.created_at, r.expires_at,
                     r.email_attempts, COALESCE(r.report_url,'')`

func scanRequests(rows pgx.Rows) ([]gripper.Request, error) {
	defer rows.Close()
	var out []gripper.Request
	for rows.Next() {
		var r gripper.Request
		if err := rows.Scan(&r.ID, &r.SiteDomain, &r.Email, &r.Status, &r.CreatedAt, &r.ExpiresAt,
			&r.EmailAttempts, &r.ReportURL); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (g *Gripper) DueChecks(ctx context.Context, now time.Time, limit int) ([]gripper.Request, error) {
	rows, err := g.Pool.Query(ctx, `
		SELECT `+requestCols+`
		FROM   gripper_report_requests r JOIN sites s ON s.id = r.site_id
		WHERE  r.status IN ('pending','pulled') AND r.next_check_at <= $1
		ORDER  BY r.next_check_at LIMIT $2`, now, limit)
	if err != nil {
		return nil, err
	}
	return scanRequests(rows)
}

func (g *Gripper) MarkFulfilled(ctx context.Context, id, reportURL string, now time.Time) (bool, error) {
	return g.moved(ctx, `
		UPDATE gripper_report_requests
		SET    status = 'fulfilled', report_url = $2, fulfilled_at = $3, next_check_at = $3
		WHERE  id = $1 AND status IN ('pending','pulled')`, id, reportURL, now)
}

func (g *Gripper) MarkFailed(ctx context.Context, id string, now time.Time) (bool, error) {
	return g.moved(ctx, `
		UPDATE gripper_report_requests
		SET    status = 'failed', next_check_at = $2
		WHERE  id = $1 AND status IN ('pending','pulled')`, id, now)
}

func (g *Gripper) MarkExpired(ctx context.Context, id string, now time.Time) (bool, error) {
	return g.moved(ctx, `
		UPDATE gripper_report_requests
		SET    status = 'expired', next_check_at = $2
		WHERE  id = $1 AND status IN ('pending','pulled')`, id, now)
}

func (g *Gripper) RescheduleCheck(ctx context.Context, id string, next time.Time) error {
	_, err := g.Pool.Exec(ctx, `
		UPDATE gripper_report_requests SET next_check_at = $2
		WHERE  id = $1 AND status IN ('pending','pulled')`, id, next)
	return err
}

func (g *Gripper) DueLinkEmails(ctx context.Context, now time.Time, maxAttempts, limit int) ([]gripper.Request, error) {
	rows, err := g.Pool.Query(ctx, `
		SELECT `+requestCols+`
		FROM   gripper_report_requests r JOIN sites s ON s.id = r.site_id
		WHERE  r.status = 'fulfilled' AND r.email IS NOT NULL
		  AND  r.email_attempts < $2 AND r.next_check_at <= $1
		ORDER  BY r.next_check_at LIMIT $3`, now, maxAttempts, limit)
	if err != nil {
		return nil, err
	}
	return scanRequests(rows)
}

func (g *Gripper) DueApologies(ctx context.Context, now time.Time, maxAttempts, limit int) ([]gripper.Request, error) {
	rows, err := g.Pool.Query(ctx, `
		SELECT `+requestCols+`
		FROM   gripper_report_requests r JOIN sites s ON s.id = r.site_id
		WHERE  r.status IN ('failed','expired') AND r.failure_notified_at IS NULL
		  AND  r.email IS NOT NULL
		  AND  r.email_attempts < $2 AND r.next_check_at <= $1
		ORDER  BY r.next_check_at LIMIT $3`, now, maxAttempts, limit)
	if err != nil {
		return nil, err
	}
	return scanRequests(rows)
}

func (g *Gripper) ClaimEmailAttempt(ctx context.Context, id string, expectStatus []string, retryAt time.Time) (bool, error) {
	return g.moved(ctx, `
		UPDATE gripper_report_requests
		SET    email_attempts = email_attempts + 1, next_check_at = $3
		WHERE  id = $1 AND status = ANY($2::text[])`, id, expectStatus, retryAt)
}

func (g *Gripper) MarkEmailed(ctx context.Context, id string, now time.Time) (bool, error) {
	return g.moved(ctx, `
		UPDATE gripper_report_requests
		SET    status = 'emailed', emailed_at = $2
		WHERE  id = $1 AND status = 'fulfilled'`, id, now)
}

func (g *Gripper) MarkEmailFailed(ctx context.Context, id string) (bool, error) {
	return g.moved(ctx, `
		UPDATE gripper_report_requests
		SET    status = 'email_failed'
		WHERE  id = $1 AND status = 'fulfilled'`, id)
}

func (g *Gripper) MarkFailureNotified(ctx context.Context, id string, now time.Time) (bool, error) {
	return g.moved(ctx, `
		UPDATE gripper_report_requests
		SET    failure_notified_at = $2
		WHERE  id = $1 AND status IN ('failed','expired') AND failure_notified_at IS NULL`, id, now)
}

// ExpireIdleSessions closes sessions idle past the retention window and drops
// their transcripts (Q7: transcript GC at 24h idle).
func (g *Gripper) ExpireIdleSessions(ctx context.Context, idleBefore time.Time) (int64, error) {
	tag, err := g.Pool.Exec(ctx, `
		UPDATE gripper_chat_sessions
		SET    status = 'expired', transcript = '[]'::jsonb
		WHERE  status = 'active' AND last_activity_at < $1`, idleBefore)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// ScrubTerminalPII nulls email, ip hash and user agent on rows that reached a
// terminal state before the retention window (Q7: 90 days), and the same on
// closed sessions. Both tables, one count.
func (g *Gripper) ScrubTerminalPII(ctx context.Context, terminalBefore time.Time) (int64, error) {
	t1, err := g.Pool.Exec(ctx, `
		UPDATE gripper_report_requests
		SET    email = NULL, client_ip_hash = NULL, user_agent = NULL
		WHERE  status IN ('emailed','email_failed','failed','expired')
		  AND  COALESCE(emailed_at, failure_notified_at, fulfilled_at, created_at) < $1
		  AND  (email IS NOT NULL OR client_ip_hash IS NOT NULL OR user_agent IS NOT NULL)`, terminalBefore)
	if err != nil {
		return 0, err
	}
	t2, err := g.Pool.Exec(ctx, `
		UPDATE gripper_chat_sessions
		SET    client_ip_hash = NULL, user_agent = NULL, transcript = '[]'::jsonb
		WHERE  status <> 'active' AND last_activity_at < $1
		  AND  (client_ip_hash IS NOT NULL OR user_agent IS NOT NULL OR transcript <> '[]'::jsonb)`, terminalBefore)
	if err != nil {
		return t1.RowsAffected(), err
	}
	return t1.RowsAffected() + t2.RowsAffected(), nil
}

func (g *Gripper) moved(ctx context.Context, sql string, args ...any) (bool, error) {
	tag, err := g.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
