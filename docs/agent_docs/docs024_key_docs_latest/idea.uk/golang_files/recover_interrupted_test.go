package main

import (
	"strings"
	"testing"
	"time"
)

// TestRunningOrderIsUnreachableByTheSweep is the defect itself, kept as a test
// so nobody "simplifies" the fix away by teaching ExpireStale about `running`.
//
// ExpireStale must NEVER expire a running order: called on a ticker in a live
// process it cannot tell a genuine 20-minute run from an abandoned one, and
// killing a live run would lose a report the customer is owed. That restraint
// is correct, and it is exactly what leaves the slot stranded after a restart —
// so the release has to come from somewhere that knows no run can be in
// flight, which is startup, and nowhere else.
func TestRunningOrderIsUnreachableByTheSweep(t *testing.T) {
	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	s, _ := NewStore("")
	s.Save(&Order{ID: "stranded", Status: "running", UpdatedAt: now.Add(-99 * 24 * time.Hour)})

	// Ninety-nine days old, both thresholds long past — and still held.
	if n := len(s.ExpireStale(24*time.Hour, 24*time.Hour, now)); n != 0 {
		t.Fatalf("ExpireStale released a running order (%d) — it must not; recovery is startup's job", n)
	}
	if got := s.ActiveCount(); got != 1 {
		t.Fatalf("want the slot still held by the sweep, got active=%d", got)
	}

	// Startup recovery is the only thing that frees it.
	if n := len(s.RecoverInterrupted(now)); n != 1 {
		t.Fatalf("want 1 recovered, got %d", n)
	}
	if got := s.ActiveCount(); got != 0 {
		t.Fatalf("slot still leaked after recovery: active=%d", got)
	}
}

// An unbilled run (review-before-pay, killed before the draft) goes back to
// `requested`: the slot frees and the operator's existing /confirm link works.
func TestRecoverInterruptedUnbilledGoesBackToRequested(t *testing.T) {
	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	s, _ := NewStore("")
	s.Save(&Order{ID: "unbilled", Status: "running", UpdatedAt: now.Add(-time.Hour)})

	rec := s.RecoverInterrupted(now)
	if len(rec) != 1 || rec[0].Status != "requested" || rec[0].Resume {
		t.Fatalf("want one order back to requested with no resume, got %+v", rec)
	}
	o, _ := s.Get("unbilled")
	if o.Status != "requested" {
		t.Fatalf("want status requested, got %q", o.Status)
	}
	if s.ActiveCount() != 0 {
		t.Fatalf("requested must hold no slot, got active=%d", s.ActiveCount())
	}
}

// A paid run (charge-first, killed after the webhook) must NOT go back to
// `requested` — /confirm would issue a second pay link and charge them twice.
// It goes back to `paid`, keeps its slot, and is flagged for re-run.
func TestRecoverInterruptedPaidIsNeverSentBackToRequested(t *testing.T) {
	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	s, _ := NewStore("")
	s.Save(&Order{ID: "paid_run", Status: "running", ProviderSessionID: "cs_live_abc", UpdatedAt: now})

	rec := s.RecoverInterrupted(now)
	if len(rec) != 1 || rec[0].Status != "paid" || !rec[0].Resume {
		t.Fatalf("want one paid order flagged for resume, got %+v", rec)
	}
	if o, _ := s.Get("paid_run"); o.Status != "paid" {
		t.Fatalf("want status paid, got %q", o.Status)
	}
	if s.ActiveCount() != 1 {
		t.Fatalf("a paid order is genuinely owed a slot, got active=%d", s.ActiveCount())
	}
}

// Nothing but `running` is touched, and a second call is a no-op — this runs on
// every single startup, including the ones with nothing to do.
func TestRecoverInterruptedTouchesNothingElse(t *testing.T) {
	now := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	s, _ := NewStore("")
	untouched := map[string]string{
		"a": "requested", "b": "awaiting_payment", "c": "awaiting_review",
		"d": "paid", "e": "delivered", "f": "declined", "g": "expired", "h": "failed",
	}
	for id, st := range untouched {
		s.Save(&Order{ID: id, Status: st, UpdatedAt: now})
	}
	if n := len(s.RecoverInterrupted(now)); n != 0 {
		t.Fatalf("recovery touched %d non-running orders", n)
	}
	for id, want := range untouched {
		if o, _ := s.Get(id); o.Status != want {
			t.Fatalf("%s: want %q, got %q", id, want, o.Status)
		}
	}

	s.Save(&Order{ID: "run", Status: "running", UpdatedAt: now})
	if n := len(s.RecoverInterrupted(now)); n != 1 {
		t.Fatalf("want 1 recovered, got %d", n)
	}
	if n := len(s.RecoverInterrupted(now)); n != 0 {
		t.Fatalf("second call recovered %d, want 0 (must be idempotent)", n)
	}
}

// The wiring, end to end through the real startup entry point: a paid run is
// actually re-executed, an unbilled one is freed, and the operator is told once.
//
// Thresholds are left at 0 (expiry disabled) deliberately: recovery is not the
// staleness sweep and must not inherit its opt-out. An operator who turns
// expiry off has not asked us to leak slots on restart.
func TestStartSweeperRecoversInterruptedRuns(t *testing.T) {
	app, sent := newTestApp() // AutoDeliver, charge-first, dispatch inline
	app.store.Save(&Order{ID: "paid_run", Status: "running", Email: "buyer@example.com",
		Domain: "example.com", ProviderSessionID: "cs_live_abc"})
	app.store.Save(&Order{ID: "unbilled", Status: "running", Email: "other@example.com",
		Domain: "other.example"})

	app.StartSweeper()

	// The paid one was re-run to completion and delivered.
	o, _ := app.store.Get("paid_run")
	if o.Status != "delivered" {
		t.Fatalf("paid order not re-run on startup: status %q", o.Status)
	}
	if !strings.Contains(o.Report, "example.com") {
		t.Fatalf("paid order has no regenerated report: %q", o.Report)
	}
	// The unbilled one was freed, and NOT re-run at our own expense.
	u, _ := app.store.Get("unbilled")
	if u.Status != "requested" {
		t.Fatalf("unbilled order: want requested, got %q", u.Status)
	}
	if u.Report != "" {
		t.Fatalf("unbilled order was re-run without an operator asking: %q", u.Report)
	}
	if got := app.store.ActiveCount(); got != 0 {
		t.Fatalf("want no slots held after recovery, got %d", got)
	}

	// Exactly one operator email, naming both orders.
	var notices [][3]string
	for _, m := range *sent {
		if strings.Contains(m[1], "recovered after a restart") {
			notices = append(notices, m)
		}
	}
	if len(notices) != 1 {
		t.Fatalf("want 1 recovery notice to the operator, got %d", len(notices))
	}
	if to := notices[0][0]; to != app.cfg.OperatorEmail {
		t.Fatalf("recovery notice went to %q, want the operator", to)
	}
	if body := notices[0][2]; !strings.Contains(body, "paid_run") || !strings.Contains(body, "unbilled") {
		t.Fatalf("recovery notice does not name both orders: %q", body)
	}
}

// A startup with nothing to recover must stay silent — no email, no noise.
func TestStartSweeperSilentWhenNothingInterrupted(t *testing.T) {
	app, sent := newTestApp()
	app.store.Save(&Order{ID: "quiet", Status: "delivered"})

	app.StartSweeper()

	for _, m := range *sent {
		if strings.Contains(m[1], "recovered after a restart") {
			t.Fatalf("emailed the operator with nothing to recover: %q", m[1])
		}
	}
}
