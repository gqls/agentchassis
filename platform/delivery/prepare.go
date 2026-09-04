package delivery

// Preparing a delivery: the gate, the claim, and the links.
//
// This is everything the delivery email needs BEFORE a word of it is written,
// and it is deliberately separate from the writing. The copy is in flux — the
// owner's brief landed in the register on 2026-08-25 and two product questions
// are still open — while the mechanics below are settled by rulings and will not
// move with the wording.
//
// It does NOT send anything, and it does not compose anything. It answers three
// questions in the one order that is safe: may we deliver, have we already, and
// what are the links.

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ReviewItemType is the work-item type carrying the owner's pre-delivery review.
//
// The gate is a work item rather than a flag or a column BECAUSE a flag is
// settable by hand and a queue row is not: "the email step reads the item; it
// does not read a flag somebody could set by hand"
// (DECISION_2026-08-21e §2, owner ruling 2026-08-21).
//
// ⚠ PRODUCERS MUST FILE THIS THROUGH actions.writeWorkItem, NOT with their own
// INSERT. That helper is the shared door carrying the policy-routability guard
// (bugs_open/333); a producer that inserts directly bypasses it, which is the
// exact defect closed on 2026-08-25 when write_audit_findings was the last
// bypassing producer. This package deliberately only READS.
const ReviewItemType = "needs_delivery_review"

// ReviewItemKeyPrefix is the item_key prefix the producer files under, stated
// here because the 2026-08-02 owner ruling requires the producer set and the key
// shape visible in one place rather than rederived per call site.
//
// ⚠ THE FULL KEY IS `delivery_review_<domain>`, NOT a site-id form — corrected
// 2026-08-26 when the producer was built: it files through the create_work_item
// action, whose key convention is `<item_key_prefix>_<domain>`, and the register
// must state the REAL shape the rows carry, not the shape a helper wished for.
// One review per site either way (a domain names one site), which is what
// idx_swi_dedup needs.
const ReviewItemKeyPrefix = "delivery_review"

var (
	// ErrNotReviewed means the owner has not approved this site for delivery.
	// The email must not be composed, let alone sent.
	ErrNotReviewed = errors.New("delivery: site has not passed pre-delivery review")

	// ErrAlreadyDelivered means this site was already handed over. It is a
	// refusal, not a warning: see Claim.
	ErrAlreadyDelivered = errors.New("delivery: site has already been handed over")
)

// Reviewed reports whether the owner has approved this site for delivery.
//
// # The predicate, and why it is not the obvious one
//
// ⚠ IT IS NOT `status = 'approved'`. That was this function's first version and
// it was wrong in a way that would never have surfaced as an error: the gate
// simply could never open, and delivery would have stalled for ever, silently.
// `[MEASURED 2026-08-25]` across 11,983 live rows the status vocabulary is
// complete / needs_human_review / detected / unresolved / cancelled / deferred /
// wont_fix / failed / rejected / verified / triaged — and `approved` is **not
// among them**. (It survives in one archive INDEX predicate, which is exactly
// what makes the mistake attractive: the word is in the schema, and nothing
// writes it.)
//
// What `HandleApproveWorkItem` actually does, read from the handler rather than
// inferred from its name: it REQUIRES the item to be at `needs_human_review`
// (anything else is a 400), and on approval it sets `status = 'complete'` with
// `result = {'resolution':…, 'approved_by':'admin', 'follow_on_item':…}`.
//
// So `complete` is necessary and NOT sufficient: `HandleResolveWorkItem` also
// writes `complete`, with `resolved_by` instead. The distinguishing key is
// `approved_by`, and it is the whole predicate. `[MEASURED 2026-08-25]` of 8,003
// live `complete` rows, 155 carry `resolved_by` and **0 carry `approved_by`**.
//
// > ⚠ CORRECTED 2026-08-26: this comment first concluded from that census that
// > "the approve path has never been exercised in production" — a census over
// > the LIVE TABLE ONLY, i.e. the exact mistake this function exists to correct,
// > caught by adversarial review the next day. `site_work_items_archive` holds
// > exactly ONE `approved_by` row (`needs_brief_review`, completed 2026-03-17,
// > `created_by='test'`). Honest statement: exercised ONCE, five months ago, on
// > a test-created item — which also proves `approved_by` survives archival in
// > the wild. The advice stands: drive the first real approval deliberately
// > rather than assume the path works.
//
// # Why it reads the archive too
//
// `site_work_items` is a ~7-day rolling window: terminal rows are moved to
// `site_work_items_archive` (24,990 `complete` rows are there already,
// `[MEASURED 2026-08-25]`). An approval IS terminal, so reading only the live
// table would convert a one-time review into one that EXPIRES after about a
// week — the site would silently read as unreviewed, and the failure would look
// exactly like "the owner has not looked at it yet".
//
// Reading both is the established house pattern for this table, not an
// invention here (migration 465, "the promoter reads archived history", and the
// 471 remedy's own census both UNION the pair).
//
// A site with no matching row in EITHER table is not approved. The absence of a
// row is a "no", never a "nothing to check": a producer that forgets to file the
// review cannot thereby skip it.
func Reviewed(ctx context.Context, db *sql.DB, siteID uuid.UUID) (bool, error) {
	var n int
	err := db.QueryRowContext(ctx, `
		SELECT count(*) FROM (
			SELECT status, result
			  FROM site_work_items
			 WHERE site_id = $1 AND item_type = $2
			UNION ALL
			SELECT status, result
			  FROM site_work_items_archive
			 WHERE site_id = $1 AND item_type = $2
		) z
		WHERE z.status = 'complete'
		  AND z.result ? 'approved_by'
	`, siteID, ReviewItemType).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("delivery: read review item: %w", err)
	}
	return n > 0, nil
}

// ReviewItemFiledStatus is the status a producer MUST file the review item at.
//
// It is a constant rather than a comment because getting it wrong is silent at
// filing time and only fails later, at the human: HandleApproveWorkItem refuses
// with a 400 ("item status is %s, expected needs_human_review") for an item in
// any other state, so a review filed at the table's default `detected` cannot be
// approved at all — the owner would press the button and be told no, with the
// delivery blocked and nothing explaining why.
const ReviewItemFiledStatus = "needs_human_review"

// ReviewItemRequiredSpec is the spec fragment a producer MUST include when
// filing the review item, merged into whatever else the item carries.
//
// ⚠ THE STATUS IS NOT THE WHOLE FILING CONTRACT, and the first version of this
// file thought it was — adversarial review (2026-08-26) read the approve handler
// one guard further than I had: after the status check, HandleApproveWorkItem
// ALSO refuses any item whose spec lacks `checkpoint: true` — 400, "this item is
// not a checkpoint — use retry or resolve instead". A producer honouring only
// ReviewItemFiledStatus files an item the owner CANNOT approve; worse, the
// error's own advice steers the owner to press RESOLVE, which writes
// `resolved_by` — the key Reviewed() deliberately ignores — so the delivery
// stalls for ever while every button press appears to succeed.
//
// (The approve endpoint additionally requires a non-empty `review_data` JSON
// body from the CALLER of the API — the admin screen's concern, not the
// producer's, so it is noted here and not encoded.)
func ReviewItemRequiredSpec() map[string]any {
	// A fresh map per call: a shared map handed to producers is a shared
	// mutable, and one producer annotating it would silently rewrite the
	// contract for the next.
	return map[string]any{"checkpoint": true}
}

// Links are the customer-facing URLs the delivery email carries.
//
// Two are deliberately empty for now and that is visible in the type rather than
// discovered at send time: ZipDownload needs /d/, which is not built, and
// StripePortal needs Stripe keys the owner has not issued. An empty field means
// "this link does not exist yet"; the composer decides what to say instead, and
// nothing here invents a URL that would 404 in a customer's inbox.
type Links struct {
	LiveSite        string // the site, live at the address we provide
	ConfirmTransfer string // /c/<token> — the customer confirms they have moved
	ZipDownload     string // /d/<token> — EMPTY until the download route exists
	DomainRent      string // £10/mo subscription payment link
	DomainBuy       string // £200 one-off payment link
	StripePortal    string // EMPTY until Stripe keys exist
}

// LinkConfig is the deployment-supplied half of the links: hosts and payment
// URLs that differ per environment and must never be compiled in.
type LinkConfig struct {
	// LinksHost is the emailed-links origin, e.g. "links.webdesign.uk". It is
	// the ONLY host customer links are built on: one hostname for everything a
	// customer is sent, nothing else on it (box/links.webdesign.uk.nginx).
	LinksHost string

	// LiveSiteBase is where a delivered site is served, e.g. "https://%s" with
	// the site's own domain substituted by the caller before it gets here.
	LiveSiteURL string

	DomainRentURL   string
	DomainBuyURL    string
	StripePortalURL string

	// ZipPresignedURL + ZipPresignExpiresAt: when set, Claim mints the ZIP token
	// CARRYING this presign (delivery.MintZipToken), so /d/<token> can 302 to
	// it. Both empty leaves Links.ZipDownload empty — the composer's problem to
	// word, never a placeholder URL. The presign comes from zip-deliverer's own
	// output (it is the only process allowed to mint one, bugs_open/245); this
	// replaces the earlier ZipDownloadPath bool the day /d/ became real
	// (2026-08-26): a bool could mint a token with NO stored URL, which /d/
	// correctly reads as stale — a link born broken that looks retriable.
	ZipPresignedURL     string
	ZipPresignExpiresAt time.Time
}

// Prepared is the result of a successful claim: the handover state, the links,
// and the plaintext confirm token (which exists nowhere else, ever again).
type Prepared struct {
	Handover Handover
	Links    Links

	// AdvertisedWindowDays is what the customer is TOLD, which is not how long
	// we actually serve. See LiveLinkWindow: we promise 30 and serve 42 on
	// purpose. A composer must use this and never derive days from
	// Handover.LiveLinkExpiresAt, or the email will promise the slack away.
	AdvertisedWindowDays int
}

// Claim is the one entry point, and it is the at-most-once gate for delivery.
//
// Order, and why it is this order:
//
//  1. REVIEW FIRST. Nothing else happens for a site the owner has not approved,
//     so a bug later in this function cannot deliver an unreviewed site.
//
//  2. STAMP SECOND, AND THE STAMP IS THE CLAIM. StampHandover is idempotent and
//     reports AlreadyHandedOver; we treat that as a REFUSAL. So the stamp — not
//     the send — is what makes delivery once-only, and a second call cannot
//     produce a second email however it is retried.
//
//     This means a stamp can exist for a site whose email then failed to send.
//     That is the deliberate direction. A failed send returns an error, so the
//     caller's work item fails and a human sees it in the queue; the opposite
//     ordering (send, then stamp) turns a stamp failure into a customer
//     receiving two delivery emails, silently, with no row anywhere recording
//     that it happened. Loud-and-recoverable beats silent-and-duplicated.
//
//  3. MINT LAST, against the handover's own expiry, so the link cannot outlive
//     the window it belongs to.
//
// A caller that wants to re-send after a failure must do it deliberately, with a
// fresh token, not by calling Claim again and hoping.
func Claim(ctx context.Context, db *sql.DB, siteID uuid.UUID, cfg LinkConfig, now time.Time) (Prepared, error) {
	if err := cfg.validate(); err != nil {
		return Prepared{}, err
	}

	reviewed, err := Reviewed(ctx, db, siteID)
	if err != nil {
		return Prepared{}, err
	}
	if !reviewed {
		return Prepared{}, ErrNotReviewed
	}

	h, err := StampHandover(ctx, db, siteID, now)
	if err != nil {
		return Prepared{}, err
	}
	if h.AlreadyHandedOver {
		return Prepared{}, ErrAlreadyDelivered
	}

	confirmToken, err := MintToken(ctx, db, siteID, PurposeConfirmTransfer,
		h.LiveLinkExpiresAt, false /* singleUse */, "delivery-email")
	if err != nil {
		return Prepared{}, err
	}

	links := Links{
		LiveSite:        cfg.LiveSiteURL,
		ConfirmTransfer: tokenURL(cfg.LinksHost, "c", confirmToken),
		DomainRent:      cfg.DomainRentURL,
		DomainBuy:       cfg.DomainBuyURL,
		StripePortal:    cfg.StripePortalURL,
	}
	if cfg.ZipPresignedURL != "" {
		zipToken, err := MintZipToken(ctx, db, siteID, h.LiveLinkExpiresAt,
			cfg.ZipPresignedURL, cfg.ZipPresignExpiresAt, "delivery-email")
		if err != nil {
			return Prepared{}, err
		}
		links.ZipDownload = tokenURL(cfg.LinksHost, "d", zipToken)
	}

	return Prepared{
		Handover:             h,
		Links:                links,
		AdvertisedWindowDays: AdvertisedLiveWindowDays,
	}, nil
}

// singleUse is FALSE for the confirm token, and that is a decision with a dead
// predecessor behind it.
//
// DECISION_2026-08-21b §4 required not-single-use because /c/ was then a GET that
// mutated: a mail scanner's prefetch would SPEND the token and lock the customer
// out of their own link. The 2026-08-24 second-click ruling removed that vector —
// GET now renders and POST confirms, and scanners do not POST.
//
// It stays false anyway, for a plainer reason that outlives the original: the
// stamp is COALESCEd, so a second confirmation cannot move the recorded date, and
// a customer who clicks twice (or forwards the mail to themselves, or has a
// flaky connection on the first press) should get the same answer both times
// rather than "that link is no longer active". Nothing is protected by spending
// it, so nothing is lost by not.

// tokenURL builds a customer link. Paths are two characters because they are
// read aloud and retyped out of an email.
// ⚠ THERE IS A SECOND BUILDER FOR THE /c/ CONFIRM LINK: delivery.ConfirmTokenURL
// in handover.go (bugs_open/477). Two builders for one customer-facing link is a
// drift risk whose failure mode is a link that quietly goes somewhere else, and
// three council seats flagged it — this cross-reference is here because a comment
// only in the OTHER file protects nobody reading this one. They exist apart only
// because two lanes were editing these two files in the same week. COLLAPSE THEM:
// export this one (taking ConfirmTokenURL's host validation with it, which this
// one delegates to LinkConfig.validate) and delete the duplicate. If you change
// the scheme, host handling or path here, change it there in the same commit.
func tokenURL(host, prefix, token string) string {
	u := url.URL{Scheme: "https", Host: host, Path: "/" + prefix + "/" + token}
	return u.String()
}

func (c LinkConfig) validate() error {
	// An empty host would build "https:///c/<token>", which is a well-formed
	// string, a broken link, and invisible until a customer clicks it. Every
	// field here fails loudly rather than producing a plausible URL.
	if strings.TrimSpace(c.LinksHost) == "" {
		return fmt.Errorf("delivery: LinksHost is required (customer links have nowhere to point)")
	}
	if strings.Contains(c.LinksHost, "/") {
		return fmt.Errorf("delivery: LinksHost %q must be a bare host, not a URL", c.LinksHost)
	}
	if strings.TrimSpace(c.LiveSiteURL) == "" {
		return fmt.Errorf("delivery: LiveSiteURL is required (the email's whole subject is that the site is live)")
	}
	return nil
}
