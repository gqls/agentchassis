package store

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/gqls/agentchassis/internal/tools-api/gripper"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Integration test for the gripper store's SQL. `go build` cannot parse SQL,
// and pgxpool has no mock in this repo, so this runs against a real Postgres
// that has had migration 436 applied — set TOOLS_API_TEST_DATABASE_URL to run
// it (see the migration header for a throwaway docker recipe). Skipped otherwise.
//
// It walks the whole lifecycle once, asserting the GUARDS rather than the
// happy path: a second submit is refused, a capped session is refused, a
// fulfilled row cannot be re-fulfilled, an email claim is bounded.
func newTestGripper(t *testing.T) (*Gripper, context.Context) {
	t.Helper()
	dsn := os.Getenv("TOOLS_API_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TOOLS_API_TEST_DATABASE_URL not set")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	// Start clean, in FK order.
	for _, q := range []string{
		`DELETE FROM gripper_report_requests`, `DELETE FROM gripper_chat_sessions`, `DELETE FROM gripper_daily_turns`,
	} {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	return &Gripper{Pool: pool}, ctx
}

const itSite = "00ff3af5-dad8-4770-9f70-3edc267a3c92" // robot-hands.com, seeded by 436

func TestGripperStoreLifecycle(t *testing.T) {
	g, ctx := newTestGripper(t)

	// ── session + turns ──
	sid, err := g.CreateSession(ctx, itSite, "iphash", "UA/1.0")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.ClaimTurn(ctx, sid, "9ec3b9ee-5b08-461b-b4f8-9e1e03579c74"); !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("cross-site claim: err=%v, want not found", err)
	}
	s, err := g.ClaimTurn(ctx, sid, itSite)
	if err != nil || s.Turns != 1 || len(s.Spec) != 0 {
		t.Fatalf("first claim: %+v err=%v", s, err)
	}
	merged, err := g.RecordTurn(ctx, sid, "about 2.5 kg", "Noted. Shape?", gripper.Normalise(map[string]interface{}{"mass_kg": 2.5}), 100, 20)
	if err != nil || merged["mass_kg"] != 2.5 {
		t.Fatalf("record turn 1: %#v err=%v", merged, err)
	}
	// Turn 2: model nulls mass (omitted from the normalised turn spec) — must survive.
	if _, err := g.ClaimTurn(ctx, sid, itSite); err != nil {
		t.Fatal(err)
	}
	merged, err = g.RecordTurn(ctx, sid, "cylinder 60mm", "Material?", gripper.Normalise(map[string]interface{}{"travel_mm": 60, "part_geometry": "cylinder 60mm"}), 100, 20)
	if err != nil || merged["mass_kg"] != 2.5 || merged["travel_mm"] != 60.0 {
		t.Fatalf("record turn 2 (merge): %#v err=%v", merged, err)
	}
	s, _ = g.ClaimTurn(ctx, sid, itSite)
	if s.Turns != 3 || len(s.Transcript) != 4 || s.InputTokens != 200 || s.Spec["mass_kg"] != 2.5 {
		t.Fatalf("session state after 2 turns: %+v", s)
	}

	// Token cap: push tokens over the line and the claim is refused as capped.
	if _, err := g.Pool.Exec(ctx, `UPDATE gripper_chat_sessions SET input_tokens = $2 WHERE id = $1`, sid, gripper.MaxSessionTokens); err != nil {
		t.Fatal(err)
	}
	if _, err := g.ClaimTurn(ctx, sid, itSite); !errors.Is(err, ErrSessionCapped) {
		t.Fatalf("token-capped claim: err=%v", err)
	}
	if _, err := g.Pool.Exec(ctx, `UPDATE gripper_chat_sessions SET input_tokens = 0 WHERE id = $1`, sid); err != nil {
		t.Fatal(err)
	}

	// ── daily cap ──
	day := time.Date(2026, 8, 16, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 2; i++ {
		if err := g.ClaimDailyTurn(ctx, day, 2); err != nil {
			t.Fatalf("daily claim %d: %v", i, err)
		}
	}
	if err := g.ClaimDailyTurn(ctx, day, 2); !errors.Is(err, ErrDailyCapReached) {
		t.Fatalf("daily cap: err=%v", err)
	}
	if err := g.ClaimDailyTurn(ctx, day.AddDate(0, 0, 1), 2); err != nil {
		t.Fatalf("next day: %v", err)
	}

	// ── submit ──
	// Incomplete spec: refused, session stays active (tx rolled back).
	if _, err := g.CreateRequestFromSession(ctx, itSite, sid, "v@example.org", "iphash", "UA"); !errors.Is(err, ErrSpecIncomplete) {
		t.Fatalf("incomplete submit: err=%v", err)
	}
	var status string
	_ = g.Pool.QueryRow(ctx, `SELECT status FROM gripper_chat_sessions WHERE id=$1`, sid).Scan(&status)
	if status != "active" {
		t.Fatalf("session moved to %s on a refused submit", status)
	}
	// Complete it and submit.
	if _, err := g.ClaimTurn(ctx, sid, itSite); err != nil {
		t.Fatal(err)
	}
	if _, err := g.RecordTurn(ctx, sid, "steel, 12/min, UR5e", "All recorded.", gripper.Normalise(map[string]interface{}{
		"surface_material": "steel", "cycle_rate": 12, "mounting": "UR5e"}), 50, 10); err != nil {
		t.Fatal(err)
	}
	rid, err := g.CreateRequestFromSession(ctx, itSite, sid, "v@example.org", "iphash", "UA")
	if err != nil {
		t.Fatalf("submit: %v", err)
	}
	if _, err := g.CreateRequestFromSession(ctx, itSite, sid, "v@example.org", "iphash", "UA"); !errors.Is(err, ErrSessionClosed) {
		t.Fatalf("second submit: err=%v, want closed", err)
	}
	if _, err := g.ClaimTurn(ctx, sid, itSite); !errors.Is(err, ErrSessionClosed) {
		t.Fatalf("chat after submit: err=%v, want closed", err)
	}
	// Inline (plain-form) request too.
	inlineSpec := gripper.Normalise(map[string]interface{}{"mass_kg": 1, "travel_mm": 30, "part_geometry": "box", "surface_material": "plastic", "cycle_rate": 5, "mounting": "arm"})
	rid2, err := g.CreateRequestInline(ctx, itSite, "w@example.org", inlineSpec, "iphash2", "UA")
	if err != nil {
		t.Fatalf("inline: %v", err)
	}

	// ── feed ──
	rows, err := g.PendingSince(ctx, nil, 100)
	if err != nil || len(rows) != 2 {
		t.Fatalf("PendingSince nil: %d rows err=%v", len(rows), err)
	}
	if rows[0].ID != rid || rows[0].Host != "robot-hands.com" || rows[0].Spec["mass_kg"] != 2.5 || rows[0].Spec["surface_material"] != "steel" {
		t.Fatalf("feed row: %+v", rows[0])
	}
	if _, has := rows[0].Spec["email"]; has {
		t.Fatal("feed spec carries email")
	}
	if err := g.MarkPulled(ctx, []string{rid, rid2}, time.Now()); err != nil {
		t.Fatal(err)
	}
	var st1 string
	var fp, lp *time.Time
	_ = g.Pool.QueryRow(ctx, `SELECT status, first_pulled_at, last_pulled_at FROM gripper_report_requests WHERE id=$1`, rid).Scan(&st1, &fp, &lp)
	if st1 != "pulled" || fp == nil || lp == nil {
		t.Fatalf("after pull: status=%s first=%v last=%v", st1, fp, lp)
	}
	// Still served while pulled (cluster dedups); since filter works.
	rows, _ = g.PendingSince(ctx, nil, 100)
	if len(rows) != 2 {
		t.Fatalf("pulled rows dropped from feed: %d", len(rows))
	}
	future := time.Now().Add(time.Hour)
	if rows, _ = g.PendingSince(ctx, &future, 100); len(rows) != 0 {
		t.Fatalf("since filter: %d rows", len(rows))
	}

	// ── poller transitions ──
	now := time.Now()
	due, err := g.DueChecks(ctx, now.Add(gripper.FirstCheckAfter+time.Second), 10)
	if err != nil || len(due) != 2 {
		t.Fatalf("DueChecks: %d err=%v", len(due), err)
	}
	if len(g.mustDue(t, ctx, now)) != 0 {
		t.Fatal("rows due before FirstCheckAfter")
	}
	ok, err := g.MarkFulfilled(ctx, rid, "https://robot-hands.com/reports/"+rid+".html", now)
	if err != nil || !ok {
		t.Fatalf("MarkFulfilled: ok=%v err=%v", ok, err)
	}
	if ok, _ := g.MarkFulfilled(ctx, rid, "x", now); ok {
		t.Fatal("re-fulfilled a fulfilled row")
	}
	if ok, _ := g.MarkExpired(ctx, rid, now); ok {
		t.Fatal("expired a fulfilled row")
	}
	// Second request: sidecar failed.
	if ok, _ := g.MarkFailed(ctx, rid2, now); !ok {
		t.Fatal("MarkFailed")
	}
	// The pull feed no longer serves either.
	if rows, _ = g.PendingSince(ctx, nil, 100); len(rows) != 0 {
		t.Fatalf("terminal rows still in feed: %d", len(rows))
	}

	// Link email lane: claim is bounded and guarded.
	links, _ := g.DueLinkEmails(ctx, now, gripper.MaxEmailAttempts, 10)
	if len(links) != 1 || links[0].ID != rid || links[0].Email != "v@example.org" || links[0].ReportURL == "" {
		t.Fatalf("DueLinkEmails: %+v", links)
	}
	for i := 0; i < gripper.MaxEmailAttempts; i++ {
		if ok, err := g.ClaimEmailAttempt(ctx, rid, []string{gripper.StatusFulfilled}, now); !ok || err != nil {
			t.Fatalf("claim %d: ok=%v err=%v", i, ok, err)
		}
	}
	if links, _ = g.DueLinkEmails(ctx, now, gripper.MaxEmailAttempts, 10); len(links) != 0 {
		t.Fatalf("row still due after %d attempts", gripper.MaxEmailAttempts)
	}
	if ok, _ := g.ClaimEmailAttempt(ctx, rid, []string{gripper.StatusFailed}, now); ok {
		t.Fatal("claim succeeded against the wrong expected status")
	}
	if ok, _ := g.MarkEmailed(ctx, rid, now); !ok {
		t.Fatal("MarkEmailed")
	}
	if ok, _ := g.MarkEmailFailed(ctx, rid); ok {
		t.Fatal("email_failed applied to an emailed row")
	}
	// Apology lane for the failed one, once.
	apols, _ := g.DueApologies(ctx, now, gripper.MaxEmailAttempts, 10)
	if len(apols) != 1 || apols[0].ID != rid2 {
		t.Fatalf("DueApologies: %+v", apols)
	}
	if ok, _ := g.ClaimEmailAttempt(ctx, rid2, []string{gripper.StatusFailed, gripper.StatusExpired}, now); !ok {
		t.Fatal("apology claim")
	}
	if ok, _ := g.MarkFailureNotified(ctx, rid2, now); !ok {
		t.Fatal("MarkFailureNotified")
	}
	if ok, _ := g.MarkFailureNotified(ctx, rid2, now); ok {
		t.Fatal("notified twice")
	}
	if apols, _ = g.DueApologies(ctx, now.Add(time.Hour), gripper.MaxEmailAttempts, 10); len(apols) != 0 {
		t.Fatalf("apology still due: %+v", apols)
	}

	// ── retention ──
	if _, err := g.Pool.Exec(ctx, `UPDATE gripper_report_requests SET emailed_at = now() - interval '91 days', failure_notified_at = now() - interval '91 days'`); err != nil {
		t.Fatal(err)
	}
	if _, err := g.Pool.Exec(ctx, `UPDATE gripper_chat_sessions SET last_activity_at = now() - interval '91 days'`); err != nil {
		t.Fatal(err)
	}
	n, err := g.ScrubTerminalPII(ctx, time.Now().Add(-gripper.PIIRetention))
	if err != nil || n != 3 { // 2 requests + 1 session
		t.Fatalf("scrub: n=%d err=%v", n, err)
	}
	var email *string
	_ = g.Pool.QueryRow(ctx, `SELECT email FROM gripper_report_requests WHERE id=$1`, rid).Scan(&email)
	if email != nil {
		t.Fatalf("email survived scrub: %v", *email)
	}
	// Idle active session expiry.
	sid2, _ := g.CreateSession(ctx, itSite, "h", "ua")
	if _, err := g.Pool.Exec(ctx, `UPDATE gripper_chat_sessions SET last_activity_at = now() - interval '25 hours' WHERE id=$1`, sid2); err != nil {
		t.Fatal(err)
	}
	if n, _ := g.ExpireIdleSessions(ctx, time.Now().Add(-gripper.SessionIdleTTL)); n != 1 {
		t.Fatalf("ExpireIdleSessions: %d", n)
	}
}

func (g *Gripper) mustDue(t *testing.T, ctx context.Context, now time.Time) []gripper.Request {
	t.Helper()
	due, err := g.DueChecks(ctx, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	return due
}
