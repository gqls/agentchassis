package main

import (
	"testing"
	"time"
)

// TestExpireStaleReleasesSlots locks the behaviour that was missing on
// 2026-07-26, when five cold orders held every capacity slot and the service
// had quietly closed itself to new work with no way out but hand-editing the
// order file under a running writer.
func TestExpireStaleReleasesSlots(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	old := now.Add(-30 * 24 * time.Hour)
	recent := now.Add(-1 * time.Hour)

	s, _ := NewStore("") // in-memory
	add := func(id, status string, updated time.Time) {
		s.Save(&Order{ID: id, Status: status, CreatedAt: updated, UpdatedAt: updated})
	}
	add("cold_review", "awaiting_review", old)
	add("cold_payment", "awaiting_payment", old)
	add("fresh_review", "awaiting_review", recent)
	add("running_long", "running", old)    // mid-run: never expire, however old
	add("delivered_old", "delivered", old) // terminal: nothing to release
	add("requested_old", "requested", old) // holds no slot; not ours to touch

	if got := s.ActiveCount(); got != 4 {
		t.Fatalf("before sweep: want 4 active, got %d", got)
	}

	released := s.ExpireStale(14*24*time.Hour, 7*24*time.Hour, now)
	if len(released) != 2 {
		t.Fatalf("want 2 released, got %d (%+v)", len(released), released)
	}
	byID := map[string]string{}
	for _, r := range released {
		byID[r.ID] = r.WasStatus
	}
	if byID["cold_review"] != "awaiting_review" || byID["cold_payment"] != "awaiting_payment" {
		t.Fatalf("wrong orders released: %+v", byID)
	}

	// Expired is its own terminal status — NOT declined. An abandoned order
	// must not be filed as a judgement the operator made.
	if o, _ := s.Get("cold_review"); o.Status != "expired" {
		t.Fatalf("want status expired, got %q", o.Status)
	}
	// Everything else untouched.
	for id, want := range map[string]string{
		"fresh_review": "awaiting_review", "running_long": "running",
		"delivered_old": "delivered", "requested_old": "requested",
	} {
		if o, _ := s.Get(id); o.Status != want {
			t.Fatalf("%s: want %q, got %q", id, want, o.Status)
		}
	}
	if got := s.ActiveCount(); got != 2 {
		t.Fatalf("after sweep: want 2 active (fresh_review + running_long), got %d", got)
	}

	// Idempotent — a ticker calls this hourly forever.
	if again := s.ExpireStale(14*24*time.Hour, 7*24*time.Hour, now); len(again) != 0 {
		t.Fatalf("second sweep released %d, want 0", len(again))
	}
}

// A row written before UpdatedAt was maintained must still age out, via CreatedAt.
func TestExpireStaleFallsBackToCreatedAt(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	s, _ := NewStore("")
	s.Save(&Order{ID: "legacy", Status: "awaiting_review", CreatedAt: now.Add(-40 * 24 * time.Hour)})
	if n := len(s.ExpireStale(14*24*time.Hour, 7*24*time.Hour, now)); n != 1 {
		t.Fatalf("legacy row not expired via CreatedAt (released %d)", n)
	}
}

// Zero/disabled thresholds must not expire anything (operator opt-out).
func TestSweepDisabledExpiresNothing(t *testing.T) {
	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	s, _ := NewStore("")
	s.Save(&Order{ID: "cold", Status: "awaiting_review", UpdatedAt: now.Add(-99 * 24 * time.Hour)})
	a := &App{cfg: Config{StaleReviewDays: 0, StalePaymentDays: 0}, store: s}
	a.sweepStale()
	if o, _ := s.Get("cold"); o.Status != "awaiting_review" {
		t.Fatalf("sweep ran with thresholds disabled: status now %q", o.Status)
	}
}
