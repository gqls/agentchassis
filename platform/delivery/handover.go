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
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/gqls/agentchassis/platform/storage"
)

// LiveLinkWindow is how long we keep serving a delivered site, and therefore how
// long its download link lives. Owner ruling 2026-08-19 (six weeks), restated
// 2026-08-20 as the window the ZIP link should share. ONE definition on purpose:
// two systems agreeing on "six weeks" is two places to get it wrong.
//
// ⚠ THIS IS 42 DAYS AND THE CUSTOMER IS TOLD 30. THAT GAP IS DELIBERATE — DO NOT
// "FIX" IT (owner ruling 2026-08-25).
//
// The attested customer-facing fact says 30 days, and it is the newer statement:
// `delivery_live_link_and_zip` on webdesign.uk's evidence_base reads "It stays
// live for 30 days", attested "owner, 2026-08-25, copy brief ... the month is
// fixed at 30 days" (register row written 2026-08-25 14:51:36Z). So a reader who
// compares the promise with this constant finds a 12-day discrepancy, dated, with
// the register — which IS the wire — on the other side. Everything about that
// reading says "stale constant, correct it".
//
// It is not stale. We PROMISE 30 and SERVE 42, on purpose: a customer must never
// be cut off at the exact moment they were told, so the operational window is
// deliberately longer than the advertised one. The slack absorbs a customer who
// comes back on day 30, a clock skew, or a retraction job that runs late.
//
// Consequences to keep straight, because they are what makes this safe:
//   - The delivery email and every piece of customer-facing copy state THIRTY
//     days, sourced from the attested fact, never from this constant.
//   - Tokens minted against this window (zip_download, confirm_transfer) outlive
//     the promise, which is the safe direction: a link that still works after the
//     promised date costs nothing; one that dies early is a support incident.
//   - The retraction job (unbuilt) must retract on THIS constant, not on the
//     promise, or it removes the slack it exists to provide.
//
// If the owner ever wants the two numbers equal, that is a ruling to record here
// and in the register together — not a drift to quietly close.
const LiveLinkWindow = 6 * 7 * 24 * time.Hour

// AdvertisedLiveWindowDays is what the CUSTOMER is told, and it exists so the gap
// above is a named thing rather than an accident nobody can see.
//
// It is deliberately NOT the source of the customer-facing copy — the attested
// fact in the register is, because that is the wire and it is what the claims gate
// checks. This constant exists so that code which needs to reason about the
// promise (a test asserting the gap is intentional, a future retraction job
// choosing which window to honour) can name it instead of hardcoding 30.
const AdvertisedLiveWindowDays = 30

// MaxPresignWindow is the SigV4 ceiling, expressed as a Duration for this
// package's convenience. It is DERIVED from platform/storage's constant and is
// deliberately not a second definition: a council round (DGH-014 round 2) caught
// the first version of this file owning its own copy, which would have diverged
// from the shared helper the day an object store raised its cap. The enforcement
// lives in storage.GetPresignedURL, so every presign caller in the estate gets it
// — not just this package.
const MaxPresignWindow = time.Duration(storage.MaxPresignExpiryMinutes) * time.Minute

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

	// The claim is the WHERE clause, not a timestamp comparison.
	//
	// ⚠ The first version of this function returned `handed_over_at <> $2` as
	// AlreadyHandedOver, which conflates "I stamped just now" with "someone else
	// stamped with my exact timestamp". Two callers whose clocks collide at
	// microsecond precision (pgx encodes timestamptz at microseconds), or any
	// caller that retries with the SAME `now` — a value hoisted out of a retry
	// loop, or derived from a work item's timestamp — would BOTH read
	// AlreadyHandedOver=false, and both would proceed to mint tokens and send a
	// delivery email. Found by adversarial review 2026-08-26; the operator
	// double-click this function's own comment names as its reason to exist is
	// exactly the shape that produces near-simultaneous calls.
	//
	// Now the row can be claimed by AT MOST ONE statement, by construction:
	// `WHERE handed_over_at IS NULL` matches for exactly one winner. A concurrent
	// second caller blocks on the row lock, re-evaluates the WHERE after commit
	// (READ COMMITTED), matches nothing, and lands in the ErrNoRows arm below —
	// which reads the stamp the winner wrote. No timestamp takes part in the
	// decision, so equal or reused `now` values cannot confuse it.
	err := db.QueryRowContext(ctx, `
		UPDATE sites
		   SET handed_over_at       = $2,
		       live_link_expires_at = COALESCE(live_link_expires_at, $3)
		 WHERE id = $1
		   AND handed_over_at IS NULL
		RETURNING handed_over_at, live_link_expires_at,
		          transfer_confirmed_at IS NOT NULL
	`, siteID, now.UTC(), now.UTC().Add(LiveLinkWindow)).
		Scan(&h.HandedOverAt, &h.LiveLinkExpiresAt, &h.TransferConfirmed)
	if errors.Is(err, sql.ErrNoRows) {
		// Either the site does not exist, or it is already stamped. Get
		// distinguishes the two: it errors on a missing site, and an existing
		// site that failed the claim necessarily has a non-null stamp committed
		// by whoever won.
		existing, getErr := Get(ctx, db, siteID)
		if getErr != nil {
			return Handover{}, getErr
		}
		existing.AlreadyHandedOver = true
		return existing, nil
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
// singleUse=true is for a token whose whole meaning is the first click. As of
// 2026-08-25 NO customer-facing token is minted single-use: the download link is
// not (a customer who clicks twice wants the file twice), and the CONFIRM link
// is not either — that changed with the second-click ruling (GET renders, POST
// confirms, so scanners cannot spend anything) and because the stamp is COALESCEd
// a re-click cannot move the recorded date; spending the token would only turn a
// customer's second press into "no longer active" for no protective gain
// (prepare.go's minting comment carries the full reasoning). The parameter stays
// because the property is real and a future token may want it.
func MintToken(ctx context.Context, db *sql.DB, siteID uuid.UUID, purpose string,
	expiresAt time.Time, singleUse bool, createdBy string) (string, error) {

	plaintext, err := mintTokenPlaintext()
	if err != nil {
		return "", err
	}

	_, err = db.ExecContext(ctx, `
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

// mintTokenPlaintext is the ONE generation rule: 32 random bytes, base64url,
// 43 chars of [A-Za-z0-9_-] — the shape the box's anchored regex admits. Every
// minter shares it so a second token kind cannot drift to a shape the box 404s.
func mintTokenPlaintext() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("delivery: mint token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
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
// confirming does not move the recorded date. ⚠ CORRECTED 2026-08-26: this
// comment used to add "the token itself is single-use, so the second click fails
// at RedeemToken" — that described the retired pre-second-click design and was
// FALSE of the running code: prepare.go mints confirm tokens with
// singleUse=false, deliberately, so a second press re-redeems and shows the same
// success page. The COALESCE is the only belt, and it is sufficient: nothing is
// protected by spending the token, so nothing is lost by not.
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

// ── The follow-up claim ──────────────────────────────────────────────────────
//
// ClaimFollowup is the follow-up email's once-only gate, and it exists because
// the delivery email's gate CANNOT be reused for it. SendDeliveryEmailAction
// goes through Claim, which calls StampHandover and refuses anything already
// stamped (ErrAlreadyDelivered). Every site a follow-up targets is by definition
// already handed over, so the existing path refuses, by design, exactly the
// population the follow-up exists for. (bugs_open/477 recorded the opposite —
// that a follow-up would be "mostly seeding rather than Go" — and that is
// corrected in this lane's PLAN.)
//
// WHAT MAKES IT AT-MOST-ONCE is `followup_sent_at IS NULL` in the WHERE clause,
// exactly as StampHandover is claimed by `handed_over_at IS NULL`: the row can
// be claimed by at most ONE statement, a concurrent second caller blocks on the
// row lock, re-evaluates after commit (READ COMMITTED), matches nothing, and is
// refused. Read StampHandover's own comment before changing this: its first
// version decided the winner by comparing timestamps, and two callers with the
// same `now` both read themselves as the winner.
//
// ⚠ THE TWO TIMESTAMP PREDICATES ARE NOT PART OF THE CLAIM. `handed_over_at <=
// $3` is DUE-ness and `live_link_expires_at > $2` is the far end of the same
// window; neither decides ownership. Deleting the first sends follow-ups too
// early; deleting the second emails a customer a confirm link that is already
// dead; deleting `followup_sent_at IS NULL` sends one every time the scheduler
// ticks. They fail differently and only the last is unbounded.
//
// The window check is IN THE CLAIM rather than after it on purpose: a site past
// its window is then never stamped, so it cannot be silently consumed by a run
// that was never going to send anything.
//
// AND THE SUPPRESSION IS IN THIS STATEMENT, NOT IN THE CALLER'S SELECT. That is
// the whole point of the bug this closes: `transfer_confirmed_at IS NULL` here
// gives that column its first reader, in the only position that is race-free. A
// scheduler selects candidates and then dispatches, and a customer who presses
// the confirm button in the gap between the two must not be emailed. A pre-query
// filter cannot promise that; this can.
var (
	// ErrFollowupSuppressed is the SUCCESS case of the confirm button: the
	// customer told us they have moved, so we do not chase them.
	ErrFollowupSuppressed = errors.New("delivery: transfer already confirmed, follow-up suppressed")

	// ErrFollowupAlreadySent is the at-most-once gate refusing a second send.
	ErrFollowupAlreadySent = errors.New("delivery: follow-up already sent for this site")

	// ErrFollowupNotDue covers both "never handed over" and "handed over too
	// recently". One error: no caller can act differently on the two, and both
	// mean the same thing to an operator reading a log.
	ErrFollowupNotDue = errors.New("delivery: site is not due a follow-up")

	// ErrFollowupWindowClosed is the OTHER end of the interval, and it is a
	// separate error because it is permanent where ErrFollowupNotDue is
	// temporary. Past live_link_expires_at every customer link this email would
	// carry is dead, so chasing a confirmation is worse than silence: it invites
	// a click on a link that will refuse.
	ErrFollowupWindowClosed = errors.New("delivery: live-link window has closed, follow-up would carry a dead link")
)

// ClaimFollowup claims the right to send ONE follow-up email for one site, and
// stamps followup_sent_at as it does so. It sends nothing: the caller sends only
// if this returns nil, and the stamp standing after a failed send is deliberate
// — for a chase email, not sending beats sending twice.
//
// handedOverBefore is the DUE cutoff, supplied by the caller because the
// interval is config (the owner has not yet ruled on it) and must not be
// compiled in.
func ClaimFollowup(ctx context.Context, db *sql.DB, siteID uuid.UUID,
	handedOverBefore, now time.Time) (Handover, error) {

	h := Handover{SiteID: siteID}
	err := db.QueryRowContext(ctx, `
		UPDATE sites
		   SET followup_sent_at = $2
		 WHERE id = $1
		   AND handed_over_at IS NOT NULL
		   AND handed_over_at <= $3
		   AND live_link_expires_at > $2
		   AND transfer_confirmed_at IS NULL
		   AND followup_sent_at IS NULL
		RETURNING handed_over_at, live_link_expires_at
	`, siteID, now.UTC(), handedOverBefore.UTC()).
		Scan(&h.HandedOverAt, &h.LiveLinkExpiresAt)
	if err == nil {
		return h, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Handover{}, fmt.Errorf("delivery: claim follow-up: %w", err)
	}

	// Nothing was claimed. Say WHY, because unlike the customer-facing token
	// errors there is no oracle to protect here: the reader is an operator or a
	// log, and "refused" without a reason is what makes a silent scheduler
	// impossible to diagnose.
	var handed, expires, confirmed, sent sql.NullTime
	if qErr := db.QueryRowContext(ctx, `
		SELECT handed_over_at, live_link_expires_at, transfer_confirmed_at, followup_sent_at
		  FROM sites WHERE id = $1
	`, siteID).Scan(&handed, &expires, &confirmed, &sent); qErr != nil {
		if errors.Is(qErr, sql.ErrNoRows) {
			return Handover{}, fmt.Errorf("delivery: no site %s", siteID)
		}
		return Handover{}, fmt.Errorf("delivery: diagnose follow-up refusal: %w", qErr)
	}
	// ORDER MATTERS, and it is "why we are not sending", most meaningful first.
	// A confirmed site that is also past its window is reported as confirmed,
	// because that is the outcome the button exists to produce and it is what an
	// operator wants to see.
	switch {
	case confirmed.Valid:
		return Handover{}, ErrFollowupSuppressed
	case sent.Valid:
		return Handover{}, ErrFollowupAlreadySent
	case handed.Valid && (!expires.Valid || !expires.Time.After(now.UTC())):
		return Handover{}, ErrFollowupWindowClosed
	default:
		return Handover{}, ErrFollowupNotDue
	}
}

// ConfirmTokenURL builds the customer-facing confirm link for a token, and
// VALIDATES the host rather than producing a plausible-looking broken URL: an
// empty host yields "https:///c/<token>", which is a well-formed string, a dead
// link, and invisible until a customer clicks it.
//
// ⚠ DUPLICATION, DELIBERATE AND TEMPORARY. prepare.go has an unexported
// tokenURL doing the same construction for the delivery email, and its host
// validation lives in LinkConfig.validate. They should be ONE function. They are
// not yet only because the bugs_open/475 lane is editing prepare.go's Links and
// LinkConfig right now, and a same-file edit from two lanes is how one lane's
// work rides into the other's commit. Collapse tokenURL into this once that
// lane is done — the two must not be allowed to drift, because the failure mode
// is a customer link that goes somewhere else.
func ConfirmTokenURL(host, token string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", fmt.Errorf("delivery: links host is required (the confirm link would have nowhere to point)")
	}
	if strings.Contains(host, "/") {
		return "", fmt.Errorf("delivery: links host %q must be a bare host, not a URL", host)
	}
	u := url.URL{Scheme: "https", Host: host, Path: "/c/" + token}
	return u.String(), nil
}

// PresignWindowFor clamps a requested download window to what the signing
// protocol will actually honour. Callers pass the customer-facing window (six
// weeks) and get back something a presign can be minted with; the difference is
// exactly why the customer holds our token instead of the presigned URL.
func PresignWindowFor(requested time.Duration) time.Duration {
	mins := int(requested / time.Minute)
	return time.Duration(storage.ClampPresignExpiryMinutes(mins)) * time.Minute
}
