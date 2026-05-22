// FILE: platform/orchestration/actions/lock_policy_test.go
package actions

import (
	"testing"
	"time"
)

func TestLockPolicyFor(t *testing.T) {
	now := time.Date(2026, 5, 19, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		lockedBy   string
		wantType   string
		wantExpiry *time.Time // nil = permanent
		wantKnown  bool
	}{
		// Human / permanent
		{"admin", LockTypePermanent, nil, true},
		{"admin-removed", LockTypePermanent, nil, true},
		{"checkpoint", LockTypePermanent, nil, true},
		{"manual", LockTypePermanent, nil, true},
		// Auto / timed 30d
		{"deploy", LockTypeTimed, ptr(now.Add(30 * 24 * time.Hour)), true},
		// Audit / timed 90d
		{"visual-design-auditor", LockTypeTimed, ptr(now.Add(90 * 24 * time.Hour)), true},
		{"imagery-quality-auditor", LockTypeTimed, ptr(now.Add(90 * 24 * time.Hour)), true},
		{"content-quality-auditor", LockTypeTimed, ptr(now.Add(90 * 24 * time.Hour)), true},
		// Adoption / timed 90d
		{"adoption", LockTypeTimed, ptr(now.Add(90 * 24 * time.Hour)), true},
		// Unknown -> conservative permanent, not known
		{"some-future-agent", LockTypePermanent, nil, false},
		{"", LockTypePermanent, nil, false},
	}

	for _, c := range cases {
		gotType, gotExpiry := LockPolicyFor(c.lockedBy, now)
		if gotType != c.wantType {
			t.Errorf("LockPolicyFor(%q) type = %q, want %q", c.lockedBy, gotType, c.wantType)
		}
		switch {
		case c.wantExpiry == nil && gotExpiry != nil:
			t.Errorf("LockPolicyFor(%q) expiry = %v, want nil", c.lockedBy, *gotExpiry)
		case c.wantExpiry != nil && gotExpiry == nil:
			t.Errorf("LockPolicyFor(%q) expiry = nil, want %v", c.lockedBy, *c.wantExpiry)
		case c.wantExpiry != nil && gotExpiry != nil && !gotExpiry.Equal(*c.wantExpiry):
			t.Errorf("LockPolicyFor(%q) expiry = %v, want %v", c.lockedBy, *gotExpiry, *c.wantExpiry)
		}
		if got := IsKnownLockSource(c.lockedBy); got != c.wantKnown {
			t.Errorf("IsKnownLockSource(%q) = %v, want %v", c.lockedBy, got, c.wantKnown)
		}
	}
}

func TestIsHardLockType(t *testing.T) {
	// permanent <=> hard; timed and review are not hard.
	if !IsHardLockType(LockTypePermanent) {
		t.Error("permanent should be hard")
	}
	if IsHardLockType(LockTypeTimed) {
		t.Error("timed should not be hard")
	}
	if IsHardLockType(LockTypeReview) {
		t.Error("review should not be hard")
	}
}

// permanent<=>human invariant: every human source is permanent, and no timed
// source is in the human set. Guards the assumption Step 3 relies on.
func TestPermanentHumanInvariant(t *testing.T) {
	now := time.Now()
	for src := range humanLockSources {
		if lt, _ := LockPolicyFor(src, now); lt != LockTypePermanent {
			t.Errorf("human source %q is not permanent (%q)", src, lt)
		}
	}
	for src := range timedLockSources {
		if humanLockSources[src] {
			t.Errorf("source %q is in both human and timed sets", src)
		}
		if lt, _ := LockPolicyFor(src, now); lt != LockTypeTimed {
			t.Errorf("timed source %q is not timed (%q)", src, lt)
		}
	}
}

func ptr(t time.Time) *time.Time { return &t }
