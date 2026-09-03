package store

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FormInbox is the island-side persistence for the static-site form receiver:
// one table, form_submissions_inbox, created by migration
// 757_form_submissions_inbox_ISLAND.sql on the ISLAND Postgres.
//
// # IT RESOLVES NOTHING, AND THAT IS THE DESIGN
//
// This process cannot see clients_db. It therefore cannot tell which site a
// token belongs to, cannot read a recipient, and must not send anything. So it
// records the token EXACTLY AS PRESENTED and stops. The cluster pulls, resolves
// the token against site_form_routes (migration 756, clients_db) and only then
// notifies anyone.
//
// That split is the security property, not a workaround for it. A forged token
// — or a forged Origin, which is easier — produces an inbox row that fails to
// resolve at ingest and is discarded. It can never reach a mailbox, because the
// machine that accepts submissions is not the machine that decides whose they
// are. Compare middleware.CORSMiddleware, which resolves a site from the Origin
// header: correct for choosing a rate-limit bucket, wrong for choosing a
// recipient, and the reason site_id below is a cross-check rather than identity.
//
// Shape borrowed from Gripper (same island, same table conventions): every state
// change is a guarded UPDATE that reports whether it moved a row, because two
// pullers may overlap.
type FormInbox struct {
	Pool *pgxpool.Pool
}

// InboxRow is one submission, in and out. The json tags are the wire format of
// GET /requests, so the cluster's collector unmarshals this struct directly.
type InboxRow struct {
	ID         string          `json:"id"`
	Token      string          `json:"token"`
	Intent     string          `json:"intent"`
	Payload    json.RawMessage `json:"payload"`
	SiteID     *string         `json:"site_id,omitempty"`
	SiteDomain string          `json:"site_domain,omitempty"`
	IPHash     string          `json:"ip_hash,omitempty"`
	UserAgent  string          `json:"user_agent,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// Insert records a submission. payload is stored verbatim; nothing in it is
// interpreted here.
func (f *FormInbox) Insert(ctx context.Context, r InboxRow) (string, error) {
	const q = `
		INSERT INTO form_submissions_inbox
		       (token, intent, payload, site_id, site_domain, ip_hash, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	var id string
	err := f.Pool.QueryRow(ctx, q,
		r.Token, r.Intent, r.Payload, r.SiteID,
		truncate(r.SiteDomain, 255), r.IPHash, truncate(r.UserAgent, 300),
	).Scan(&id)
	return id, err
}

// ClaimPending atomically moves up to limit pending rows to 'pulled' and
// returns them, oldest first.
//
// # WHY CLAIM-AND-RETURN IN ONE STATEMENT
//
// Selecting and then updating would let two overlapping pulls hand the same
// submission to the collector twice. FOR UPDATE SKIP LOCKED makes concurrent
// pulls take disjoint sets instead of blocking on each other.
//
// # THE FAILURE THIS CHOOSES, STATED PLAINLY
//
// The rows are marked pulled before the HTTP response is written, so a response
// lost in flight leaves them 'pulled' and un-ingested — visible, but not
// re-served. That is deliberate: the alternative (mark after acknowledgement)
// needs a two-phase protocol, and the alternative to THAT (never mark, dedupe
// downstream) delivers every row repeatedly for ever if the collector's dedupe
// is ever wrong. A stranded row is recoverable by hand and is queryable —
// `WHERE status='pulled' AND pulled_at < now() - interval '1 hour'` cross-checked
// against the cluster — whereas a silently duplicated lead emails a customer
// twice. Prefer the failure that is visible and reversible.
func (f *FormInbox) ClaimPending(ctx context.Context, limit int, now time.Time) ([]InboxRow, error) {
	const q = `
		WITH claimed AS (
		    SELECT id
		      FROM form_submissions_inbox
		     WHERE status = 'pending'
		     ORDER BY created_at
		     LIMIT $1
		       FOR UPDATE SKIP LOCKED
		)
		UPDATE form_submissions_inbox f
		   SET status = 'pulled', pulled_at = $2
		  FROM claimed c
		 WHERE f.id = c.id
		RETURNING f.id, f.token, f.intent, f.payload, f.site_id,
		          COALESCE(f.site_domain, ''), COALESCE(f.ip_hash, ''),
		          COALESCE(f.user_agent, ''), f.created_at`

	rows, err := f.Pool.Query(ctx, q, limit, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []InboxRow
	for rows.Next() {
		var r InboxRow
		if err := rows.Scan(&r.ID, &r.Token, &r.Intent, &r.Payload, &r.SiteID,
			&r.SiteDomain, &r.IPHash, &r.UserAgent, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
