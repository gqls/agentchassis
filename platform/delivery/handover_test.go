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
		{"one second inside survives", MaxPresignWindow - time.Second, MaxPresignWindow - time.Second},
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
