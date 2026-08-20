// Package delivery holds the handover state for a finished customer site: when it
// was handed over, when the address WE host it at stops serving, and the hashed
// tokens behind the customer-facing links.
//
// WHY A PACKAGE AND NOT A HANDFUL OF QUERIES IN A HANDLER. Three customer links
// are coming (the ZIP download and the confirm-transfer click now; Phase 5's
// editor session next) and every one of them is the same four decisions: mint a
// secret we never store, expire it, spend it at most once where that matters, and
// resolve it back to exactly one site. Getting any of those subtly different per
// link is how a tenant-scoping hole appears. They are decided once, here.
//
// # The six weeks, and why the ZIP link cannot simply be a longer presign
//
// Owner ruling 2026-08-19: the address we host the finished site at expires six
// weeks after handover. Owner ruling 2026-08-20: the ZIP download link should last
// "the longest time we have", i.e. the same six weeks.
//
// That is NOT deliverable by lengthening the presigned URL, and the failure mode
// is silent. A presigned URL is capped by the SigV4 signing protocol at 604,800
// seconds, the cap is enforced by the object store and NOT by the AWS SDK
// (checked: aws-sdk-go-v2 v1.25.1's aws/signer/v4 and service/s3 v1.51.0 both
// sign whatever duration you ask for and return a well-formed URL). Measured
// 2026-08-20 against the live bucket, with a deliberately absent key so the HTTP
// status is the whole answer:
//
//	X-Amz-Expires=604800  -> HTTP 404 NoSuchKey   (the control: signature accepted)
//	X-Amz-Expires=604801  -> HTTP 403 SignatureDoesNotMatch
//	X-Amz-Expires=3628800 -> HTTP 403 SignatureDoesNotMatch   (six weeks)
//
// Exact to the second, and the error is SignatureDoesNotMatch — which reads as
// broken credentials, not a long expiry, so a six-week link would have sent the
// next reader to debug the one healthy thing. See LANDMINES.md.
//
// So the customer never receives a presigned URL. They receive a token of ours
// with a six-week life; redeeming it is what mints a short presign, server-side,
// per click. The six weeks then lives in ONE place — LiveLinkWindow — instead of
// being a number two systems have to agree on.
//
// # What this package does NOT do, stated so the columns are not mistaken for a
// working mechanism
//
//   - Nothing enforces LiveLinkExpiresAt. Delivered sites serve from a git repo
//     synced to object storage and nothing takes them down: no scheduled
//     retraction, no retention job, no TTL (checked 2026-08-19). The stamp records
//     the intent; the retraction job is a separate build.
//   - Nothing here mints a presign or sends an email. RedeemToken resolves a
//     token to a site and spends it; the caller decides what that permits.
//   - There is no HTTP surface yet. As of this commit the only callers are tests,
//     which is deliberate: the package's whole dependency is the database, and the
//     tests exercise that. A capability whose dependency is its ENVIRONMENT cannot
//     be trusted until something real calls it — this one's is not.
package delivery

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// LiveLinkWindow is how long we keep serving a delivered site, and therefore how
// long its download link lives. Owner ruling 2026-08-19 (six weeks), restated
// 2026-08-20 as the window the ZIP link should share. ONE definition on purpose:
// two systems agreeing on "six weeks" is two places to get it wrong.
const LiveLinkWindow = 6 * 7 * 24 * time.Hour

// MaxPresignWindow is the hard ceiling on a presigned object URL: the SigV4
// signing protocol's 604,800 seconds. It is here as a named constant because the
// SDK does not enforce it and the object store's refusal is indistinguishable
// from a credentials fault — see the package comment for the measurement. Any
// code minting a presign for a customer must clamp to this, and must NOT read
// LiveLinkWindow for that purpose.
const MaxPresignWindow = 7 * 24 * time.Hour

// Token purposes. This is a CLOSED vocabulary, enforced by a CHECK constraint in
// migration 511: adding one costs a migration, which is the point — a fourth
// customer-facing link should be visible in the ledger, not appear in a constant.
const (
	PurposeZipDownload     = "zip_download"
	PurposeConfirmTransfer = "confirm_transfer"
)

var (
	// ErrTokenNotFound covers every reason a token will not redeem — unknown,
	// expired, revoked, already spent, or the wrong purpose. ONE error on
	// purpose: distinguishing them to a caller who will put the result in an
	// HTTP response tells an attacker which guess was closer.
	ErrTokenNotFound = errors.New("delivery: token not valid")

	// ErrNotHandedOver is for callers gating on handover.
	ErrNotHandedOver = errors.New("delivery: site has not been handed over")
)

// Handover is the state of one site's delivery.
type Handover struct {
	SiteID            uuid.UUID
	HandedOverAt      time.Time
	LiveLinkExpiresAt time.Time
	TransferConfirmed bool
	AlreadyHandedOver bool // true when StampHandover found an existing stamp and left it alone
}

// StampHandover records that a site has been handed to the customer, and is
// IDEMPOTENT: a second call returns the existing stamp with AlreadyHandedOver set
// and changes nothing. That matters because handover is the sort of button an
// operator double-clicks, and re-stamping would silently extend the six-week
// window every time.
//
// It does not mint tokens. Minting is the caller's step, so a re-issued link
// never depends on re-running handover.
func StampHandover(ctx context.Context, db *sql.DB, siteID uuid.UUID, now time.Time) (Handover, error) {
	var h Handover
	h.SiteID = siteID

	// One statement. The COALESCE keeps an existing stamp, and RETURNING tells us
	// which branch happened without a second read that another writer could win.
	err := db.QueryRowContext(ctx, `
		UPDATE sites
		   SET handed_over_at       = COALESCE(handed_over_at, $2),
		       live_link_expires_at = COALESCE(live_link_expires_at, $3)
		 WHERE id = $1
		RETURNING handed_over_at, live_link_expires_at,
		          transfer_confirmed_at IS NOT NULL,
		          handed_over_at <> $2
	`, siteID, now.UTC(), now.UTC().Add(LiveLinkWindow)).
		Scan(&h.HandedOverAt, &h.LiveLinkExpiresAt, &h.TransferConfirmed, &h.AlreadyHandedOver)
	if errors.Is(err, sql.ErrNoRows) {
		return Handover{}, fmt.Errorf("delivery: no site %s", siteID)
	}
	if err != nil {
		return Handover{}, fmt.Errorf("delivery: stamp handover: %w", err)
	}
	return h, nil
}

// Get reads a site's handover state without changing it.
func Get(ctx context.Context, db *sql.DB, siteID uuid.UUID) (Handover, error) {
	h := Handover{SiteID: siteID}
	var handed, expires sql.NullTime
	err := db.QueryRowContext(ctx, `
		SELECT handed_over_at, live_link_expires_at, transfer_confirmed_at IS NOT NULL
		  FROM sites WHERE id = $1
	`, siteID).Scan(&handed, &expires, &h.TransferConfirmed)
	if errors.Is(err, sql.ErrNoRows) {
		return Handover{}, fmt.Errorf("delivery: no site %s", siteID)
	}
	if err != nil {
		return Handover{}, fmt.Errorf("delivery: read handover: %w", err)
	}
	h.HandedOverAt, h.LiveLinkExpiresAt = handed.Time, expires.Time
	return h, nil
}

// IsHandedOver is the gate Phase 5's editor session exchange uses. It is the ONLY
// thing handover gates: not deploys, not rewrites, not locks, not reconciliation.
func IsHandedOver(ctx context.Context, db *sql.DB, siteID uuid.UUID) (bool, error) {
	var ok bool
	err := db.QueryRowContext(ctx,
		`SELECT handed_over_at IS NOT NULL FROM sites WHERE id = $1`, siteID).Scan(&ok)
	if errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("delivery: no site %s", siteID)
	}
	if err != nil {
		return false, fmt.Errorf("delivery: read handover: %w", err)
	}
	return ok, nil
}

// MintToken issues one customer link and returns its PLAINTEXT, which is the only
// place the plaintext ever exists — the row stores sha256 hex. A leaked database
// therefore yields no working links, and there is no "resend the same link" path:
// re-issuing means minting a new one, which is the correct behaviour anyway.
//
// singleUse=true is for a token whose whole meaning is the first click (the
// confirm-transfer link). A download link is not single-use: a customer who
// clicks twice wants the file twice.
func MintToken(ctx context.Context, db *sql.DB, siteID uuid.UUID, purpose string,
	expiresAt time.Time, singleUse bool, createdBy string) (string, error) {

	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("delivery: mint token: %w", err)
	}
	plaintext := base64.RawURLEncoding.EncodeToString(raw)

	_, err := db.ExecContext(ctx, `
		INSERT INTO customer_access_tokens
		       (site_id, purpose, token_hash, expires_at, single_use, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, siteID, purpose, HashToken(plaintext), expiresAt.UTC(), singleUse, createdBy)
	if err != nil {
		// A bad purpose lands here as a CHECK violation rather than passing
		// silently, which is why the vocabulary is a constraint and not a comment.
		return "", fmt.Errorf("delivery: insert token (purpose %q): %w", purpose, err)
	}
	return plaintext, nil
}

// HashToken is the one hashing rule. Exported so a test — or an operator holding a
// link from an email — can find the row without guessing the encoding.
func HashToken(plaintext string) string {
	sum := sha256.Sum256([]byte(plaintext))
	return hex.EncodeToString(sum[:])
}

// RedeemToken resolves a plaintext token to its site and spends it. It returns
// ErrTokenNotFound for every failing reason, deliberately (see the var block).
//
// The whole check is ONE statement with the predicates in the WHERE clause, so
// two simultaneous clicks on a single-use link cannot both succeed: exactly one
// UPDATE matches `used_at IS NULL`. Reading first and updating second would be a
// race, and this is precisely the sort of link a customer double-clicks.
//
// purpose is required and matched: a download token must never open a
// confirm-transfer door just because both are tokens on the same table.
func RedeemToken(ctx context.Context, db *sql.DB, plaintext, purpose string, now time.Time) (uuid.UUID, error) {
	var siteID uuid.UUID
	err := db.QueryRowContext(ctx, `
		UPDATE customer_access_tokens
		   SET use_count = use_count + 1,
		       used_at   = COALESCE(used_at, $3)
		 WHERE token_hash = $1
		   AND purpose    = $2
		   AND revoked_at IS NULL
		   AND expires_at > $3
		   AND (NOT single_use OR used_at IS NULL)
		RETURNING site_id
	`, HashToken(plaintext), purpose, now.UTC()).Scan(&siteID)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, ErrTokenNotFound
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("delivery: redeem token: %w", err)
	}
	return siteID, nil
}

// ConfirmTransfer spends a confirm-transfer token and stamps the site. The click
// IS the state (owner ruling 2026-08-19): no form, no reply parsing. It is what
// stops the weekly chase email.
//
// The stamp is COALESCEd, so a customer who clicks an old link after already
// confirming does not move the recorded date — but note the token itself is
// single-use, so in practice the second click fails at RedeemToken. Both belts
// are here because the one that matters is whichever survives the next refactor.
func ConfirmTransfer(ctx context.Context, db *sql.DB, plaintext string, now time.Time) (uuid.UUID, error) {
	siteID, err := RedeemToken(ctx, db, plaintext, PurposeConfirmTransfer, now)
	if err != nil {
		return uuid.Nil, err
	}
	if _, err := db.ExecContext(ctx,
		`UPDATE sites SET transfer_confirmed_at = COALESCE(transfer_confirmed_at, $2) WHERE id = $1`,
		siteID, now.UTC()); err != nil {
		return uuid.Nil, fmt.Errorf("delivery: stamp transfer confirmation: %w", err)
	}
	return siteID, nil
}

// PresignWindowFor clamps a requested download window to what the signing
// protocol will actually honour. Callers pass the customer-facing window (six
// weeks) and get back something a presign can be minted with; the difference is
// exactly why the customer holds our token instead of the presigned URL.
func PresignWindowFor(requested time.Duration) time.Duration {
	if requested <= 0 || requested > MaxPresignWindow {
		return MaxPresignWindow
	}
	return requested
}
