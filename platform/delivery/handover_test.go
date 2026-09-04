// Tests for the delivery handover state.
//
// ⚠ READ THIS BEFORE TRUSTING THESE TESTS. sqlmock asserts the SQL this package
// SENDS; it cannot tell you what Postgres DOES with it. Every property that
// actually protects a customer here is a SQL-level property — single-use spending,
// purpose isolation, idempotent stamping, the CHECK on the purpose vocabulary —
// and a mock's own bookkeeping cannot assert any of them. So the mock tests below
// cover the Go-level behaviour only, and the SQL semantics are verified separately
// against a real Postgres inside a rolled-back transaction; that run and its
// results are recorded in the lane's NOTES for 2026-08-20.
//
// The division is deliberate rather than lazy: a test that passes because a mock
// was told to return one row is indistinguishable from a test that passes because
// the query is correct.
package delivery

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

func TestHashTokenIsStableAndNotThePlaintext(t *testing.T) {
	const plain = "abc123"
	h1, h2 := HashToken(plain), HashToken(plain)
	if h1 != h2 {
		t.Fatalf("hash is not stable: %q vs %q", h1, h2)
	}
	if len(h1) != 64 {
		t.Fatalf("expected 64 hex chars of sha256, got %d (%q)", len(h1), h1)
	}
	if h1 == plain {
		t.Fatal("hash equals the plaintext")
	}
	if HashToken("abc124") == h1 {
		t.Fatal("two different plaintexts hashed the same")
	}
}

// The clamp is the whole reason the customer holds our token instead of a
// presigned URL, so it is asserted at the boundary and one second past it — the
// same boundary the live measurement found.
func TestPresignWindowForClampsAtTheSigV4Ceiling(t *testing.T) {
	cases := []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{"six weeks clamps down", LiveLinkWindow, MaxPresignWindow},
		{"the ceiling itself survives", MaxPresignWindow, MaxPresignWindow},
		{"one second past the ceiling clamps", MaxPresignWindow + time.Second, MaxPresignWindow},
		{"one minute inside survives", MaxPresignWindow - time.Minute, MaxPresignWindow - time.Minute},
		// Sub-minute precision is NOT expressible: the storage API is
		// GetPresignedURL(ctx, key, expiryMinutes int), so a duration is truncated
		// to whole minutes. This case previously asserted that a second inside the
		// ceiling survived to the second, which was asserting precision the system
		// has never had — the test was wrong, not the clamp. Truncation is also the
		// SAFE direction: it can only move a window further below the ceiling.
		{"sub-minute precision truncates DOWN, never up", MaxPresignWindow - time.Second, MaxPresignWindow - time.Minute},
		{"a few seconds truncates to zero and therefore clamps to the ceiling", 5 * time.Second, MaxPresignWindow},
		{"an hour survives", time.Hour, time.Hour},
		{"zero and negative fall back to the ceiling", 0, MaxPresignWindow},
		{"negative", -time.Hour, MaxPresignWindow},
	}
	for _, c := range cases {
		if got := PresignWindowFor(c.in); got != c.want {
			t.Errorf("%s: PresignWindowFor(%v) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

// LiveLinkWindow must exceed what a presign can do, or this package's whole
// reason for existing is gone and nobody would notice from the code.
func TestTheSixWeekWindowIsLongerThanAPresignCanBe(t *testing.T) {
	if LiveLinkWindow <= MaxPresignWindow {
		t.Fatalf("LiveLinkWindow (%v) is not longer than MaxPresignWindow (%v) — "+
			"if these ever converge, delete the token indirection instead of leaving it unexplained",
			LiveLinkWindow, MaxPresignWindow)
	}
	if want := 42 * 24 * time.Hour; LiveLinkWindow != want {
		t.Fatalf("LiveLinkWindow = %v, want %v (owner ruling 2026-08-19: six weeks)", LiveLinkWindow, want)
	}
	// The gap between what we serve and what we promise is DELIBERATE (owner
	// ruling 2026-08-25), and it is asserted here so that closing it fails a test
	// rather than looking like tidying up. A session comparing this constant with
	// the attested "30 days" fact will read a 12-day drift with the newer date on
	// the register's side; without this assertion, correcting it is the obvious
	// and wrong move. See LiveLinkWindow's comment for why we serve longer.
	if AdvertisedLiveWindowDays != 30 {
		t.Fatalf("AdvertisedLiveWindowDays = %d, want 30 (attested fact "+
			"delivery_live_link_and_zip, owner copy brief 2026-08-25)", AdvertisedLiveWindowDays)
	}
	if promised := time.Duration(AdvertisedLiveWindowDays) * 24 * time.Hour; LiveLinkWindow <= promised {
		t.Fatalf("LiveLinkWindow (%v) must be LONGER than the %d days we promise (%v): "+
			"we promise 30 and serve 42 on purpose, so a customer is never cut off at the "+
			"exact moment they were told. Equalising these removes the slack deliberately "+
			"put there (owner ruling 2026-08-25) — change both numbers and the register "+
			"together, or neither", LiveLinkWindow, AdvertisedLiveWindowDays, promised)
	}
	if want := 604800 * time.Second; MaxPresignWindow != want {
		t.Fatalf("MaxPresignWindow = %v, want %v (the SigV4 ceiling, measured 2026-08-20)", MaxPresignWindow, want)
	}
}

// Every failing redemption must be indistinguishable to the caller: an HTTP
// handler that says "expired" rather than "not found" tells an attacker their
// guess was closer.
func TestRedeemMapsEveryMissToOneError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("UPDATE customer_access_tokens").
		WillReturnError(sql.ErrNoRows)

	_, err = RedeemToken(context.Background(), db, "whatever", PurposeZipDownload, time.Now())
	if !errors.Is(err, ErrTokenNotFound) {
		t.Fatalf("expected ErrTokenNotFound, got %v", err)
	}
}

// A real database error must NOT be flattened into "token not valid" — that would
// make an outage look like a bad link and hide it from whoever is on call.
func TestRedeemDoesNotFlattenARealDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	boom := errors.New("connection reset")
	mock.ExpectQuery("UPDATE customer_access_tokens").WillReturnError(boom)

	_, err = RedeemToken(context.Background(), db, "whatever", PurposeZipDownload, time.Now())
	if errors.Is(err, ErrTokenNotFound) {
		t.Fatal("a connection error was reported as an invalid token")
	}
	if !errors.Is(err, boom) {
		t.Fatalf("the underlying error was lost: %v", err)
	}
}

// The purpose travels into the query. Asserted because the table is shared by
// every customer link, so a download token must not open the confirm door.
func TestRedeemSendsThePurposeAsAPredicate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	site := uuid.New()
	now := time.Now().UTC()
	mock.ExpectQuery("UPDATE customer_access_tokens").
		WithArgs(HashToken("tok"), PurposeConfirmTransfer, now).
		WillReturnRows(sqlmock.NewRows([]string{"site_id"}).AddRow(site))

	got, err := RedeemToken(context.Background(), db, "tok", PurposeConfirmTransfer, now)
	if err != nil {
		t.Fatal(err)
	}
	if got != site {
		t.Fatalf("site mismatch: %s vs %s", got, site)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// A missing site must not read as "not handed over" — that is the difference
// between a closed door and a door that was never there.
func TestIsHandedOverDistinguishesMissingSiteFromNotHandedOver(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT handed_over_at IS NOT NULL").WillReturnError(sql.ErrNoRows)
	if _, err := IsHandedOver(context.Background(), db, uuid.New()); err == nil {
		t.Fatal("a missing site returned no error")
	}

	mock.ExpectQuery("SELECT handed_over_at IS NOT NULL").
		WillReturnRows(sqlmock.NewRows([]string{"ok"}).AddRow(false))
	ok, err := IsHandedOver(context.Background(), db, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("an un-handed-over site reported true")
	}
}

// ── ClaimFollowup (bugs_open/477) ─────────────────────────────────────────────
//
// ⚠ WHAT THESE CAN AND CANNOT PROVE, per this file's opening rule. sqlmock
// asserts the SQL this package SENDS; it cannot tell you what Postgres DOES with
// it. The properties that actually protect a customer here are SQL-level — that
// `followup_sent_at IS NULL` makes the claim at-most-once, and that
// `transfer_confirmed_at IS NULL` suppresses a customer who has already
// confirmed — and they were verified against a real Postgres inside a
// rolled-back transaction, WITH a negative control proving the suppression
// predicate is what refuses (drop it and the confirmed site claims). That run
// and its output are recorded in the lane's NOTES for 2026-09-04.
//
// What is left for a mock is the Go half: which error a caller gets, which is
// what decides whether a scheduled run logs "nothing to do" or fails a work item.

func TestClaimFollowupClassifiesEveryRefusal(t *testing.T) {
	now := time.Now().UTC()
	for _, tc := range []struct {
		name                    string
		confirmed, sent, handed interface{}
		expires                 interface{}
		want                    error
	}{
		// BOTH stamps set, and the window closed too: this is the case that
		// actually tests the switch's ORDER. With only transfer_confirmed_at set
		// it would pass whatever order the arms were in, and the name would be a
		// claim the fixture never made.
		{"confirmed wins over everything", now, now, now.Add(-60 * 24 * time.Hour), now.Add(-2 * 24 * time.Hour), ErrFollowupSuppressed},
		{"already sent", nil, now, now.Add(-8 * 24 * time.Hour), now.Add(30 * 24 * time.Hour), ErrFollowupAlreadySent},
		{"window closed", nil, nil, now.Add(-60 * 24 * time.Hour), now.Add(-2 * 24 * time.Hour), ErrFollowupWindowClosed},
		{"never handed over", nil, nil, nil, nil, ErrFollowupNotDue},
		{"handed over too recently", nil, nil, now.Add(-1 * time.Hour), now.Add(30 * 24 * time.Hour), ErrFollowupNotDue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()
			site := uuid.New()

			mock.ExpectQuery(`(?s)UPDATE sites.*followup_sent_at IS NULL`).
				WillReturnError(sql.ErrNoRows)
			mock.ExpectQuery(`(?s)SELECT handed_over_at, live_link_expires_at`).
				WillReturnRows(sqlmock.NewRows([]string{"handed", "expires", "confirmed", "sent"}).
					AddRow(tc.handed, tc.expires, tc.confirmed, tc.sent))

			_, err = ClaimFollowup(context.Background(), db, site, now.Add(-7*24*time.Hour), now)
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v — a caller cannot tell why nothing was sent", err, tc.want)
			}
		})
	}
}

// A site row that does not exist is NOT one of the four refusals: it is a
// caller error, and flattening it into "not due" would hide a bad site_id in a
// scheduled run for ever.
func TestClaimFollowupDistinguishesAMissingSiteFromARefusal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	site := uuid.New()

	mock.ExpectQuery(`(?s)UPDATE sites.*followup_sent_at IS NULL`).WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`(?s)SELECT handed_over_at, live_link_expires_at`).WillReturnError(sql.ErrNoRows)

	_, err = ClaimFollowup(context.Background(), db, site, time.Now().Add(-7*24*time.Hour), time.Now())
	if err == nil || !strings.Contains(err.Error(), "no site") {
		t.Fatalf("got %v, want a missing-site error", err)
	}
	for _, refusal := range []error{ErrFollowupSuppressed, ErrFollowupAlreadySent, ErrFollowupNotDue, ErrFollowupWindowClosed} {
		if errors.Is(err, refusal) {
			t.Fatalf("a missing site was reported as %v, which reads as correct behaviour", refusal)
		}
	}
}

// The claim's own success path returns the window the caller needs to mint a
// token that cannot outlive the site.
func TestClaimFollowupReturnsTheLiveLinkWindow(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	site := uuid.New()
	now := time.Now().UTC()
	expires := now.Add(30 * 24 * time.Hour)

	mock.ExpectQuery(`(?s)UPDATE sites.*transfer_confirmed_at IS NULL.*followup_sent_at IS NULL`).
		WillReturnRows(sqlmock.NewRows([]string{"handed", "expires"}).
			AddRow(now.Add(-8*24*time.Hour), expires))

	h, err := ClaimFollowup(context.Background(), db, site, now.Add(-7*24*time.Hour), now)
	if err != nil {
		t.Fatalf("ClaimFollowup: %v", err)
	}
	if !h.LiveLinkExpiresAt.Equal(expires) {
		t.Errorf("LiveLinkExpiresAt = %v, want %v — the confirm token would be minted against the wrong window", h.LiveLinkExpiresAt, expires)
	}
	if h.SiteID != site {
		t.Errorf("SiteID = %v, want %v", h.SiteID, site)
	}
}

// ConfirmTokenURL must refuse a host that would build a plausible dead link
// rather than producing one. "https:///c/<token>" is well formed, resolves
// nowhere, and is invisible until a customer clicks it.
func TestConfirmTokenURLRefusesAHostThatWouldBuildADeadLink(t *testing.T) {
	if _, err := ConfirmTokenURL("", "tok"); err == nil {
		t.Error("an empty host built a URL; it must refuse")
	}
	if _, err := ConfirmTokenURL("links.webdesign.uk/c", "tok"); err == nil {
		t.Error("a host containing a path built a URL; it must refuse")
	}
	got, err := ConfirmTokenURL("links.webdesign.uk", "tok")
	if err != nil {
		t.Fatalf("a good host was refused: %v", err)
	}
	if got != "https://links.webdesign.uk/c/tok" {
		t.Errorf("built %q", got)
	}
}
