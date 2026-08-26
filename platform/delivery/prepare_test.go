package delivery

// Tests for the delivery claim.
//
// ⚠ SAME LIMIT AS handover_test.go, and it bites harder here. sqlmock asserts the
// SQL this package SENDS and the ORDER it sends it in; it cannot tell you what
// Postgres does with any of it. What that buys here is real but narrow: the
// ordering properties below (review before stamp, stamp before mint, and nothing
// after a refusal) ARE Go-level properties, so a mock can hold them — and holding
// them is the whole safety argument for Claim.
//
// What a mock CANNOT hold is that an unmet expectation proves a call did not
// happen for the right reason. So every "nothing else happened" assertion below
// is backed by a mutation in the lane's NOTES: the guard is removed, and the test
// must fail. A test that passes because the mock was never asked is
// indistinguishable from one that passes because the code is correct.

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func testLinkConfig() LinkConfig {
	return LinkConfig{
		LinksHost:       "links.webdesign.uk",
		LiveSiteURL:     "https://example-site.co.uk",
		DomainRentURL:   "https://pay.example/rent",
		DomainBuyURL:    "https://pay.example/buy",
		StripePortalURL: "",
		ZipDownloadPath: false,
	}
}

// expectReview queues the gate read, answering with `approved` rows.
//
// The query it stands for reads BOTH site_work_items and its archive; the regex
// below deliberately names the archive so that dropping the UNION arm — which
// would silently make a one-time approval expire after ~7 days — fails here.
func expectReview(mock sqlmock.Sqlmock, siteID uuid.UUID, approvedRows int) {
	mock.ExpectQuery(`site_work_items_archive`).
		WithArgs(siteID, ReviewItemType).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(approvedRows))
}

// expectStamp mirrors StampHandover's post-2026-08-26 shape: the claim UPDATE
// carries `handed_over_at IS NULL` in its WHERE, so "already handed over" is the
// claim matching NO rows followed by a read of the existing stamp — not a
// boolean column. The regex pins the claim predicate so a regression to the old
// timestamp-comparison discriminator (which double-delivered on equal `now`
// values) fails here.
func expectStamp(mock sqlmock.Sqlmock, now time.Time, already bool) {
	claim := mock.ExpectQuery(`(?s)UPDATE sites.*handed_over_at IS NULL`)
	if already {
		// The claim matches nothing; StampHandover then reads the winner's stamp.
		claim.WillReturnRows(sqlmock.NewRows([]string{
			"handed_over_at", "live_link_expires_at", "transfer_confirmed",
		}))
		mock.ExpectQuery(`SELECT handed_over_at`).
			WillReturnRows(sqlmock.NewRows([]string{
				"handed_over_at", "live_link_expires_at", "transfer_confirmed",
			}).AddRow(now.UTC(), now.UTC().Add(LiveLinkWindow), false))
		return
	}
	claim.WillReturnRows(sqlmock.NewRows([]string{
		"handed_over_at", "live_link_expires_at", "transfer_confirmed",
	}).AddRow(now.UTC(), now.UTC().Add(LiveLinkWindow), false))
}

// An UNREVIEWED site must not be stamped, must not mint a token, and must not
// produce links. This is the gate the owner asked for on 2026-08-21, and its
// value is entirely in what does NOT happen after it.
func TestClaimRefusesAnUnreviewedSiteAndDoesNothingElse(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	siteID := uuid.New()
	expectReview(mock, siteID, 0)
	// Deliberately NOTHING else is queued. An UPDATE sites or an INSERT INTO
	// customer_access_tokens now fails as an unexpected call.

	_, err = Claim(context.Background(), db, siteID, testLinkConfig(), time.Now())
	if !errors.Is(err, ErrNotReviewed) {
		t.Fatalf("Claim on an unreviewed site returned %v, want ErrNotReviewed", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The stamp IS the claim, so a site already handed over is a refusal and mints
// nothing. Without this, a retry sends a second delivery email to a customer.
func TestClaimRefusesASiteAlreadyHandedOverAndMintsNoToken(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	siteID := uuid.New()
	now := time.Now()
	expectReview(mock, siteID, 1)
	expectStamp(mock, now, true /* already handed over */)
	// No INSERT is queued: minting after a refused claim would hand out a live
	// customer link for a delivery that is not happening.

	_, err = Claim(context.Background(), db, siteID, testLinkConfig(), now)
	if !errors.Is(err, ErrAlreadyDelivered) {
		t.Fatalf("Claim on a delivered site returned %v, want ErrAlreadyDelivered", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

func TestClaimBuildsCustomerLinksOnTheLinksHost(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	siteID := uuid.New()
	now := time.Now()
	expectReview(mock, siteID, 1)
	expectStamp(mock, now, false)
	mock.ExpectExec(`INSERT INTO customer_access_tokens`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	got, err := Claim(context.Background(), db, siteID, testLinkConfig(), now)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}

	if !strings.HasPrefix(got.Links.ConfirmTransfer, "https://links.webdesign.uk/c/") {
		t.Errorf("confirm link is %q; it must sit on the links host, which is the "+
			"only hostname the box serves customer links from", got.Links.ConfirmTransfer)
	}
	// The token in the URL is the plaintext, which exists nowhere else — not in
	// the row (hashed), not in a log. If this is ever empty the email ships a
	// link to nothing.
	token := strings.TrimPrefix(got.Links.ConfirmTransfer, "https://links.webdesign.uk/c/")
	if len(token) < 20 {
		t.Errorf("confirm token %q is shorter than the box's own regex admits (20-128 chars)", token)
	}

	// The ZIP route does not exist yet. An invented URL here would 404 in a
	// customer's inbox, which is worse than an absent link the copy can explain.
	if got.Links.ZipDownload != "" {
		t.Errorf("ZipDownload = %q, want empty while /d/ is unbuilt", got.Links.ZipDownload)
	}
	if got.Links.StripePortal != "" {
		t.Errorf("StripePortal = %q, want empty while no Stripe keys exist", got.Links.StripePortal)
	}
	if got.Links.LiveSite == "" || got.Links.DomainRent == "" || got.Links.DomainBuy == "" {
		t.Errorf("a configured link came back empty: %+v", got.Links)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// The email must promise what we ADVERTISE, never what we serve. Deriving the
// window from the handover's expiry would tell the customer 42 days and give
// away the slack that exists so nobody is cut off on the day they were told.
func TestClaimReportsTheAdvertisedWindowNotTheServedOne(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	siteID := uuid.New()
	now := time.Now()
	expectReview(mock, siteID, 1)
	expectStamp(mock, now, false)
	mock.ExpectExec(`INSERT INTO customer_access_tokens`).WillReturnResult(sqlmock.NewResult(1, 1))

	got, err := Claim(context.Background(), db, siteID, testLinkConfig(), now)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if got.AdvertisedWindowDays != 30 {
		t.Errorf("AdvertisedWindowDays = %d, want 30 (the attested fact)", got.AdvertisedWindowDays)
	}

	servedDays := int(got.Handover.LiveLinkExpiresAt.Sub(got.Handover.HandedOverAt).Hours() / 24)
	if servedDays != 42 {
		t.Errorf("served window = %d days, want 42", servedDays)
	}
	if got.AdvertisedWindowDays >= servedDays {
		t.Errorf("advertised (%d) is not less than served (%d): the gap is deliberate "+
			"and the email must quote the smaller number", got.AdvertisedWindowDays, servedDays)
	}
}

// A misconfigured host builds "https:///c/<token>" — a well-formed string, a
// broken link, and invisible until a customer clicks it. It must fail here.
func TestClaimRefusesAConfigThatWouldBuildABrokenLink(t *testing.T) {
	for _, tc := range []struct {
		name string
		cfg  LinkConfig
	}{
		{"no links host", LinkConfig{LiveSiteURL: "https://x.co.uk"}},
		{"links host is a URL", LinkConfig{LinksHost: "https://links.webdesign.uk/", LiveSiteURL: "https://x.co.uk"}},
		{"no live site", LinkConfig{LinksHost: "links.webdesign.uk"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, _, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			// No queries queued at all: config is checked before the database is
			// touched, so a broken config cannot stamp a handover.

			_, err = Claim(context.Background(), db, uuid.New(), tc.cfg, time.Now())
			if err == nil {
				t.Fatal("a config that cannot build a working link was accepted")
			}

			// ⚠ "err != nil" IS NOT ENOUGH HERE, and proving that cost a mutation.
			// With validate() deleted, Claim runs on to the gate query, sqlmock
			// refuses a call it was never told to expect, and Claim returns THAT
			// error — so a bare non-nil check passes on a build with no validation
			// at all. mock.ExpectationsWereMet() does not catch it either: it
			// reports unmet expectations, and there were none to meet.
			//
			// The error must therefore be identifiable AS the config error. These
			// are the exact phrases validate() produces; if they are reworded,
			// reword them here, because the alternative is this test silently
			// going back to asserting nothing.
			msg := err.Error()
			isConfigErr := strings.Contains(msg, "LinksHost is required") ||
				strings.Contains(msg, "must be a bare host") ||
				strings.Contains(msg, "LiveSiteURL is required")
			if !isConfigErr {
				t.Fatalf("Claim failed with %q, which is not the config refusal: the "+
					"config must be rejected BEFORE the database is touched, or a "+
					"broken link ships with a handover already stamped", msg)
			}
			if errors.Is(err, ErrNotReviewed) || errors.Is(err, ErrAlreadyDelivered) {
				t.Fatalf("config error reported as a gate refusal: %v", err)
			}
		})
	}
}

// Reviewed must treat an absent row as "no". A producer that forgets to file the
// review item must not thereby skip the gate.
func TestReviewedTreatsNoRowAsNotApproved(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	siteID := uuid.New()
	expectReview(mock, siteID, 0)

	ok, err := Reviewed(context.Background(), db, siteID)
	if err != nil {
		t.Fatalf("Reviewed: %v", err)
	}
	if ok {
		t.Error("a site with no review item read as approved")
	}
}

// The gate's predicate is the single load-bearing thing in this package, and its
// first version was wrong in a way no error would ever have shown: `status =
// 'approved'` is a value nothing writes, so the gate could never open and
// delivery would have stalled for ever, looking exactly like "not reviewed yet".
//
// These assertions pin what the approve handler ACTUALLY writes. If the handler
// changes, this must fail rather than the delivery pipeline going quiet.
func TestReviewedAsksForWhatTheApproveHandlerActuallyWrites(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	siteID := uuid.New()
	// sqlmock matches the query text against this regex, so it IS the assertion:
	// the gate must ask for complete + approved_by, across both tables.
	mock.ExpectQuery(`(?s)site_work_items.*UNION ALL.*site_work_items_archive.*complete.*approved_by`).
		WithArgs(siteID, ReviewItemType).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	ok, err := Reviewed(context.Background(), db, siteID)
	if err != nil {
		t.Fatalf("Reviewed: %v", err)
	}
	if !ok {
		t.Error("an approved review read as not approved")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("the gate query is not the one the approve handler feeds: %v", err)
	}
}

// A producer filing at the table's default `detected` cannot ever be approved:
// HandleApproveWorkItem 400s for any status other than needs_human_review, so
// the owner would press approve and be refused, with delivery blocked and
// nothing saying why.
func TestReviewItemMustBeFiledAtNeedsHumanReview(t *testing.T) {
	if ReviewItemFiledStatus != "needs_human_review" {
		t.Errorf("ReviewItemFiledStatus = %q; HandleApproveWorkItem requires "+
			"needs_human_review and refuses anything else with a 400",
			ReviewItemFiledStatus)
	}
}

func TestReviewItemKeyIsOnePerSite(t *testing.T) {
	a, b := uuid.New(), uuid.New()
	if ReviewItemKey(a) == ReviewItemKey(b) {
		t.Error("two sites share a review item key: the dedup index would collapse them")
	}
	if !strings.HasPrefix(ReviewItemKey(a), "delivery_review:") {
		t.Errorf("item key %q does not carry its type prefix", ReviewItemKey(a))
	}
}

var _ = sql.ErrNoRows // keep the import honest if assertions above are edited

// The filing contract is status AND spec, and the second half was missed once:
// HandleApproveWorkItem refuses any item whose spec lacks checkpoint:true, with
// an error whose own advice ("use retry or resolve instead") steers the owner to
// the button that writes resolved_by — the key the gate deliberately ignores. A
// producer honouring only the status files a review nobody can ever approve.
func TestReviewItemRequiredSpecCarriesTheCheckpointFlag(t *testing.T) {
	spec := ReviewItemRequiredSpec()
	if v, ok := spec["checkpoint"].(bool); !ok || !v {
		t.Fatalf("ReviewItemRequiredSpec() = %v: HandleApproveWorkItem 400s any item "+
			"whose spec lacks checkpoint:true, so a producer using this contract would "+
			"file an unapprovable review", spec)
	}
	// A fresh map per call: a producer annotating a shared map would silently
	// rewrite the contract for the next producer.
	a, b := ReviewItemRequiredSpec(), ReviewItemRequiredSpec()
	a["checkpoint"] = false
	if v := b["checkpoint"].(bool); !v {
		t.Error("ReviewItemRequiredSpec returns a shared map: one caller's mutation reached another")
	}
}
