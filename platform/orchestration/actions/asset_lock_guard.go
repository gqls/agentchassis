// FILE: platform/orchestration/actions/asset_lock_guard.go
//
// ONE definition of "may automation overwrite this assets row?" for every
// writer that replaces an asset — the DB row, the object in storage, or the
// file the row points at in a site repo (bugs_open/143).
//
// Why this file exists. `assets.locked_at` is an owner approval (Phase I1 D5):
// a locked artefact must never be silently replaced by machinery. Before this
// helper the rule was written out by hand at four sites — StoreAssetAction's
// upsert, ingest_staged_asset's pre-check and in-tx re-check, recordDerivedAsset's
// upsert, and derive_brand_head_assets' pre-commit check — and a fifth writer
// (derive_card_asset) had the predicate ONLY on its provenance upsert, which
// runs AFTER the git commit has already replaced the file. That is the whole of
// bugs_open/143: a duplicated classifier where one copy sat at the wrong point
// in the sequence, so the approved row survived and its artefact did not. 016b's
// rule for that shape is one shared predicate plus a lockstep test, not five
// careful edits.
//
// Two things must be true of every action that regenerates an asset's bytes:
//
//  1. The lock is read BEFORE the artefact is written. A provenance upsert
//     guarded with `WHERE assets.locked_at IS NULL` protects only the DB row.
//     lockedAssetKeys is that pre-check, and it exists to stop the expensive,
//     UNGUARDABLE half — the S3 read, the image work, the git commit — from
//     happening at all.
//  2. The write itself still carries the predicate, so the decision is race-free
//     (TOCTOU): a lock taken between the pre-check and the write still
//     suppresses the write. That is assetAgentWritableSQL. A suppressed write
//     must then be NOTICED — an `ExecContext` whose RowsAffected is discarded
//     reports a lock-blocked write as a clean success, which is the second half
//     of what 143 was about.
//
// TWO PROPERTIES ARE DELIBERATE, and both fail CLOSED:
//
//  1. NO status filter. `assets.status` is unconstrained text (no CHECK; the
//     live vocabulary already holds active/superseded/retired), so conditioning
//     a safety guard on it fails OPEN the day a locked row carries a status
//     nobody enumerated. A lock on ANY row of the key locks the key. This was
//     argued and approved at council on the brand-head fix (corr bfd73f71,
//     round 1 HIGH — the original `status='active'` filter was removed as a
//     result); it is carried here unchanged. The cost is stated rather than
//     hidden: a locked *superseded* row blocks re-derivation of the active one
//     until a human clears it, which is the conservative direction.
//  2. NOT expiry-aware — `locked_at IS NULL` is the whole test, unlike
//     pageComponentAgentWritableSQL, which releases an EXPIRED timed lock.
//     `assets` is one of migration 115's Pattern-A lock tables and does carry
//     lock_type / lock_expires_at, so the canonical predicate is
//     schema-applicable here — but unifying them is register LOCK-004's named,
//     still-outstanding "Go predicate sweep", and adopting it would WEAKEN this
//     guard (an expired timed lock would stop protecting the artefact). That is
//     a change to what the mechanism GUARANTEES, not a bug fix, so it is not
//     done as a side effect of one. Measured 2026-07-31: 5 locked assets rows
//     fleet-wide, lock_expires_at NULL on all 5, so the two readings are
//     indistinguishable today — which is why the choice is made on the guarantee
//     and not on the data. When LOCK-004 lands, this file is the ONE line it
//     edits.
//
// The lock READ is derived from the write predicate rather than spelled out a
// second time, so the two can never drift apart — which is the failure this
// file was created to end.
//
// Design credit: a concurrent session reached bugs_open/143 at the same time and
// its (withdrawn) asset_lock_helpers.go is where the derived-negation predicate,
// the lock-detail set below and the LOCK-004 framing came from. Kept here so the
// better half of two collided attempts is the half that survives.

package actions

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

// assetAgentWritableSQL returns the canonical "may automation overwrite this
// assets row?" predicate. alias is the table alias including the trailing dot
// ("a." / "assets.") or "" when the statement has no alias.
//
// Use it in the WHERE clause of the UPDATE / ON CONFLICT DO UPDATE itself, so
// the decision is race-free, and pair it with lockedAssetKeys for the pre-check.
func assetAgentWritableSQL(alias string) string {
	return alias + "locked_at IS NULL"
}

// assetLockedSQL is the exact negation of assetAgentWritableSQL. Derived, never
// re-typed: a future change to the writable predicate (LOCK-004) moves both.
func assetLockedSQL(alias string) string {
	return "NOT (" + assetAgentWritableSQL(alias) + ")"
}

// assetLock is one locked assets row, carrying what an operator needs in order
// to act on a refusal. LockedBy is free text on live rows — real values range
// from "admin" to a whole sentence citing the bug that set it — so it is
// reported verbatim rather than classified.
type assetLock struct {
	AssetKey string
	LockedBy string
	LockType string
	LockedAt time.Time
}

// assetLockSet maps asset_key -> the lock holding it. A key absent from the set
// is writable.
type assetLockSet map[string]assetLock

// Locked reports whether the key carries a lock.
func (s assetLockSet) Locked(key string) bool {
	_, ok := s[key]
	return ok
}

// Keys lists the locked keys, sorted, so a result payload is stable.
func (s assetLockSet) Keys() []string {
	out := make([]string, 0, len(s))
	for k := range s {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Describe renders one lock for a refusal reason or log line. Empty string when
// the key is not locked, so callers can build a message unconditionally.
func (s assetLockSet) Describe(key string) string {
	l, ok := s[key]
	if !ok {
		return ""
	}
	lockType := l.LockType
	if lockType == "" {
		lockType = "unclassified"
	}
	by := l.LockedBy
	if by == "" {
		by = "unknown"
	}
	return fmt.Sprintf("%s locked %s by %s (%s)",
		key, l.LockedAt.UTC().Format("2006-01-02"), by, lockType)
}

// lockedAssetKeys reports which of the named asset_keys on a site are locked.
// Absent keys are simply not in the result; an empty keys slice returns an empty
// set without querying (`IN ()` would be a syntax error, and a caller with
// nothing to check finds nothing to skip either way).
//
// Callers MUST treat an error as LOCKED — the whole point of the guard is that
// "I could not tell" and "it is free to overwrite" are different answers and
// only one of them is safe. Every current caller returns the error rather than
// proceeding.
func lockedAssetKeys(ctx context.Context, db *sql.DB, siteID uuid.UUID, keys ...string) (assetLockSet, error) {
	locked := make(assetLockSet, len(keys))
	if db == nil || len(keys) == 0 {
		return locked, nil
	}

	// DISTINCT ON: a key can have several rows (superseded/retired copies) and,
	// with no status filter, any of them can hold the lock. Report the most
	// recent — the guard only needs "is it locked", the row is for the message.
	//
	// toPGTextArrayLiteral (+ the explicit ::text[] cast it documents) is the
	// package's existing way to pass a []string to Postgres — reused rather than
	// reaching for pq.Array, which nothing else here uses.
	q := `
		SELECT DISTINCT ON (asset_key)
		       asset_key, COALESCE(locked_by, ''), COALESCE(lock_type, ''), locked_at
		  FROM assets
		 WHERE site_id = $1
		   AND asset_key = ANY($2::text[])
		   AND ` + assetLockedSQL("") + `
		 ORDER BY asset_key, locked_at DESC`

	rows, err := db.QueryContext(ctx, q, siteID, toPGTextArrayLiteral(keys))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var l assetLock
		if err := rows.Scan(&l.AssetKey, &l.LockedBy, &l.LockType, &l.LockedAt); err != nil {
			return nil, err
		}
		locked[l.AssetKey] = l
	}
	return locked, rows.Err()
}
