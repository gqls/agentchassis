// FILE: platform/orchestration/actions/lock_policy.go
//
// PLAN_lock_coherence.md Step 1 — the single source of truth for
// "what lock_type and expiry does a given locked_by get".
//
// Today this mapping is split between the check_component_lock.go hard/soft
// switch and the prose policy table in the docs. This function centralises it.
// Lock-writers should call LockPolicyFor at lock-set time and stamp the
// returned lock_type / lock_expires_at onto the row, instead of each writer
// hardcoding its own values.
//
// Approved policy (031_LOCKS_should_locks_expire.md, 2026-05-08, + adoption
// 2026-05-19):
//
//   locked_by                                   lock_type   expiry
//   admin / admin-removed / checkpoint / manual permanent   —
//   deploy (auto-lock-on-deploy)                timed       +30d
//   visual-design-auditor / imagery-quality-auditor
//                                               timed       +90d
//   content-quality-auditor                     timed       +90d
//   adoption (faithful first pass)              timed       +90d
//   (unrecognised source)                       permanent   —   (conservative:
//                                                                don't silently
//                                                                expire what we
//                                                                can't classify)
//
// Invariant this preserves: permanent <=> human-set. Every human source maps to
// permanent, and only human sources (plus the conservative unknown default) do.
// That invariant is what lets readers use lock_type='permanent' as the hard/soft
// test (PLAN_lock_coherence.md Step 3), retiring the locked_by string-switch.

package actions

import "time"

// Lock type values (lock_type column).
const (
	LockTypePermanent = "permanent"
	LockTypeTimed     = "timed"
	LockTypeReview    = "review"
)

// Default expiry windows.
const (
	deployLockWindow   = 30 * 24 * time.Hour // auto-lock-on-deploy
	auditorLockWindow  = 90 * 24 * time.Hour // auditor approvals
	adoptionLockWindow = 90 * 24 * time.Hour // faithful first pass
)

// Human (permanent) lock sources. permanent <=> human-set.
var humanLockSources = map[string]bool{
	"admin":         true,
	"admin-removed": true,
	"checkpoint":    true,
	"manual":        true,
}

// Timed lock sources and their windows.
var timedLockSources = map[string]time.Duration{
	"deploy":                  deployLockWindow,
	"visual-design-auditor":   auditorLockWindow,
	"imagery-quality-auditor": auditorLockWindow,
	"content-quality-auditor": auditorLockWindow,
	"adoption":                adoptionLockWindow,
}

// LockPolicyFor returns the lock_type and lock_expires_at a new lock from the
// given source should carry. expiresAt is nil for permanent locks. `now` is
// passed in (rather than calling time.Now internally) so callers can use the
// same clock as the surrounding transaction and so the function is trivially
// testable.
//
// Unrecognised sources default to permanent (never silently expire something we
// can't classify). Callers that want to surface unregistered sources can check
// IsKnownLockSource and log a warning.
func LockPolicyFor(lockedBy string, now time.Time) (lockType string, expiresAt *time.Time) {
	if humanLockSources[lockedBy] {
		return LockTypePermanent, nil
	}
	if window, ok := timedLockSources[lockedBy]; ok {
		exp := now.Add(window)
		return LockTypeTimed, &exp
	}
	// Conservative default — see policy note above.
	return LockTypePermanent, nil
}

// IsKnownLockSource reports whether the source is explicitly covered by the
// policy. Lets writers log a warning when they encounter an unregistered source
// (which would otherwise silently get the conservative permanent default).
func IsKnownLockSource(lockedBy string) bool {
	if humanLockSources[lockedBy] {
		return true
	}
	_, ok := timedLockSources[lockedBy]
	return ok
}

// IsHardLockType reports whether a lock_type denotes a hard (human, never
// auto-expiring) lock. This is the replacement for the locked_by string-switch
// in check_component_lock.go (PLAN_lock_coherence.md Step 3): once 053's
// backfill has established permanent<=>human, IsHardLockType(lockType) equals
// the old IsHard derivation.
func IsHardLockType(lockType string) bool {
	return lockType == LockTypePermanent
}
