package delivery

// The ZIP download link's server side: mint a token that CARRIES the presigned
// URL, and redeem a token for a URL that is known to still work.
//
// Why the URL rides the token row (DECISION_2026-08-21b, owner-resolved
// 2026-08-21): a presign is capped at 7 days by SigV4 and the customer's window
// is AdvertisedLiveWindowDays; the only process allowed to MINT presigns is the
// spawned zip-deliverer (bugs_open/245 — no standing service holds object-store
// keys). So the deliverer presigns at ZIP-cut time, the URL is stored here, and
// /d/ is pure DB -> 302: no credentials on any listener, ever.
//
// The freshness split matters more than it looks: a stale stored URL must NEVER
// be redirected to, because the object store answers an expired presign with
// 403 SignatureDoesNotMatch — which reads as broken credentials, not an old
// link, and would send the customer (and the next debugger) at the one healthy
// component. Stale is therefore its own answer, distinct from "no such token".

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ErrZipURLStale means the TOKEN is valid but the stored presign has aged out.
//
// Deliberately distinct from ErrTokenNotFound, and deliberately NOT folded into
// it: the no-oracle rule exists so an attacker cannot learn which failure they
// hit, but reaching this error requires HOLDING a valid token — there is nothing
// left to protect from its holder, and the customer-facing behaviour must
// differ (an honest "being refreshed" page, never a redirect to a 403).
var ErrZipURLStale = errors.New("delivery: zip link needs refreshing")

// MintZipToken mints a zip_download token carrying the presigned URL it redeems
// to. presignExpiresAt is when the STORED URL dies (<=7 days out, the SigV4
// ceiling); tokenExpiresAt is when the TOKEN dies (the handover's own expiry,
// LiveLinkWindow out). The token outliving the URL is the designed state — the
// refresher re-stamps stored_url; the token is the customer's durable handle.
func MintZipToken(ctx context.Context, db *sql.DB, siteID uuid.UUID, tokenExpiresAt time.Time,
	presignedURL string, presignExpiresAt time.Time, createdBy string) (string, error) {

	if presignedURL == "" {
		// An empty stored URL would mint a token that can only ever render the
		// refresh page — a link that is born broken and looks retriable.
		return "", fmt.Errorf("delivery: mint zip token: presigned URL is required")
	}

	plaintext, err := mintTokenPlaintext()
	if err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO customer_access_tokens
		       (site_id, purpose, token_hash, expires_at, single_use, created_by,
		        stored_url, stored_url_expires_at)
		VALUES ($1, $2, $3, $4, false, $5, $6, $7)
	`, siteID, PurposeZipDownload, HashToken(plaintext), tokenExpiresAt.UTC(), createdBy,
		presignedURL, presignExpiresAt.UTC())
	if err != nil {
		return "", fmt.Errorf("delivery: insert zip token: %w", err)
	}
	return plaintext, nil
}

// LookupZipURL redeems a zip_download token for its stored URL.
//
// One statement, and the WHERE carries every validity rule so the row cannot
// change between a check and an update: purpose, unexpired, unrevoked. A miss
// for ANY of those reasons is the single undifferentiated ErrTokenNotFound.
// A HIT whose stored URL has aged out is ErrZipURLStale — see the var's comment
// for why that distinction leaks nothing.
//
// use_count increments on every successful lookup (fresh or stale): the column
// is the only record of whether a customer ever used their link, and a stale
// visit is a use — it is what arms the refresh.
func LookupZipURL(ctx context.Context, db *sql.DB, plaintext string, now time.Time) (string, error) {
	var url sql.NullString
	var urlExpires sql.NullTime
	err := db.QueryRowContext(ctx, `
		UPDATE customer_access_tokens
		   SET use_count = use_count + 1,
		       used_at   = COALESCE(used_at, $3)
		 WHERE token_hash = $1
		   AND purpose    = $2
		   AND expires_at > $3
		   AND revoked_at IS NULL
		RETURNING stored_url, stored_url_expires_at
	`, HashToken(plaintext), PurposeZipDownload, now.UTC()).Scan(&url, &urlExpires)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrTokenNotFound
	}
	if err != nil {
		return "", fmt.Errorf("delivery: look up zip url: %w", err)
	}
	if !url.Valid || url.String == "" || !urlExpires.Valid || !urlExpires.Time.After(now) {
		return "", ErrZipURLStale
	}
	return url.String, nil
}
