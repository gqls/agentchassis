package main

// store_uploads.go — the database half of chunked uploads
// (PLAN_2026-08-26_large_uploads.md). A pending_uploads row is a quota
// RESERVATION: it counts toward the account's usage from `begin`, so
// concurrent begins cannot promise the same headroom twice, and it converts
// into a media row (finish) or disappears (abort/reap) — never both, because
// finalize does the insert and the delete in one transaction.

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"
	"time"
)

var ErrNoUpload = errors.New("no such pending upload")

type PendingPart struct {
	Size int64  `json:"size"`
	Sha1 string `json:"sha1"`
}

type PendingUpload struct {
	ID            int64
	AccountID     int64
	NoteID        int64
	Kind          string
	Mime          string
	DeclaredBytes int64
	PartSizeBytes int64
	StorageKey    string
	B2FileID      string
	Parts         map[string]PendingPart // key = part number as decimal string
}

// PartsTotal is the arithmetic both sides agree on: ceil(declared/partSize).
func (p *PendingUpload) PartsTotal() int {
	return int((p.DeclaredBytes + p.PartSizeBytes - 1) / p.PartSizeBytes)
}

// ExpectedPartSize: every part is exactly part_size_bytes except the last,
// which is the remainder. Strict equality is what keeps finish's size-sum
// check meaningful.
func (p *PendingUpload) ExpectedPartSize(n int) int64 {
	total := p.PartsTotal()
	if n < total {
		return p.PartSizeBytes
	}
	last := p.DeclaredBytes - p.PartSizeBytes*int64(total-1)
	return last
}

// EffectiveLimits resolves an account's quota and per-file cap: the override
// column when set, the process-wide default otherwise. The defaults are
// PARAMETERS, not fields, so this file stays free of env reads and the caller
// (who owns the env) states what the defaults are.
func (s *Store) EffectiveLimits(ctx context.Context, accountID, defQuota, defMaxUpload int64) (quota, maxUpload int64, err error) {
	err = s.DB.QueryRowContext(ctx,
		`SELECT COALESCE(media_quota_override_bytes, $2),
		        COALESCE(max_upload_override_bytes, $3)
		 FROM accounts WHERE id=$1`,
		accountID, defQuota, defMaxUpload).Scan(&quota, &maxUpload)
	return
}

// BeginPendingUpload reserves quota for a declared size and records the B2
// large-file handle. Same locking discipline as AddMedia: the account row is
// locked so the usage read (media_bytes PLUS open reservations) cannot be
// stale by the time the reservation lands. quota arrives resolved (the
// caller's EffectiveLimits) so this transaction makes no policy decisions.
func (s *Store) BeginPendingUpload(ctx context.Context, accountID, noteID int64, kind, mime string, declared, partSize, quota int64, storageKey, b2FileID string) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var used int64
	if err := tx.QueryRowContext(ctx,
		`SELECT a.media_bytes + COALESCE((SELECT SUM(p.declared_bytes) FROM pending_uploads p WHERE p.account_id=a.id),0)
		 FROM accounts a WHERE a.id=$1 FOR UPDATE OF a`, accountID).Scan(&used); err != nil {
		return 0, err
	}
	if used+declared > quota {
		return 0, ErrQuotaExceeded
	}
	var owned int64
	if err := tx.QueryRowContext(ctx,
		`SELECT id FROM notes WHERE id=$1 AND account_id=$2 AND deleted_at IS NULL`,
		noteID, accountID).Scan(&owned); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNoAccount
		}
		return 0, err
	}
	var id int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO pending_uploads (account_id, note_id, kind, mime, declared_bytes, part_size_bytes, storage_key, b2_file_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		accountID, noteID, kind, mime, declared, partSize, storageKey, b2FileID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// GetPendingUpload is account-scoped like every other read: an upload id from
// another account is indistinguishable from one that never existed.
func (s *Store) GetPendingUpload(ctx context.Context, id, accountID int64) (*PendingUpload, error) {
	p := &PendingUpload{ID: id, AccountID: accountID}
	var partsRaw []byte
	err := s.DB.QueryRowContext(ctx,
		`SELECT note_id, kind, mime, declared_bytes, part_size_bytes, storage_key, b2_file_id, parts
		 FROM pending_uploads WHERE id=$1 AND account_id=$2`,
		id, accountID).Scan(&p.NoteID, &p.Kind, &p.Mime, &p.DeclaredBytes, &p.PartSizeBytes, &p.StorageKey, &p.B2FileID, &partsRaw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoUpload
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(partsRaw, &p.Parts); err != nil {
		return nil, err
	}
	return p, nil
}

// RecordUploadPart notes that B2 confirmed part n. Re-recording a part number
// REPLACES it — the idempotence that makes the editor's per-part retry safe
// (B2's own re-upload semantics are the same, so the two stores agree).
func (s *Store) RecordUploadPart(ctx context.Context, id, accountID int64, n int, size int64, sha1hex string) error {
	// The part number travels as TEXT because it is a jsonb object key — and
	// because pgx refuses to encode a Go int into a text-typed parameter.
	res, err := s.DB.ExecContext(ctx,
		`UPDATE pending_uploads
		 SET parts = jsonb_set(parts, ARRAY[$3::text],
		                       jsonb_build_object('size', $4::bigint, 'sha1', $5::text), true)
		 WHERE id=$1 AND account_id=$2`,
		id, accountID, strconv.Itoa(n), size, sha1hex)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNoUpload
	}
	return nil
}

// FinalizePendingUpload converts the reservation into a media row — insert and
// delete in ONE transaction, so the quota arithmetic is conserved atomically:
// the media insert's trigger adds declared_bytes to accounts.media_bytes in
// the same moment the reservation stops counting. The caller has already run
// b2_finish_large_file; a finalize without a finished B2 file would record a
// row whose download 404s, which is why the order is B2 first, exactly as the
// small path's B2-first-then-insert.
func (s *Store) FinalizePendingUpload(ctx context.Context, id, accountID int64) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var p PendingUpload
	err = tx.QueryRowContext(ctx,
		`SELECT note_id, kind, mime, declared_bytes, storage_key, b2_file_id
		 FROM pending_uploads WHERE id=$1 AND account_id=$2 FOR UPDATE`,
		id, accountID).Scan(&p.NoteID, &p.Kind, &p.Mime, &p.DeclaredBytes, &p.StorageKey, &p.B2FileID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoUpload
	}
	if err != nil {
		return 0, err
	}
	var mediaID int64
	err = tx.QueryRowContext(ctx,
		`INSERT INTO media (note_id, account_id, kind, mime, bytes, byte_len, storage_key, b2_file_id, ordering)
		 VALUES ($1,$2,$3,$4,NULL,$5,$6,$7,
		         COALESCE((SELECT MAX(ordering)+1 FROM media WHERE note_id=$1 AND kind=$3),0))
		 RETURNING id`,
		p.NoteID, accountID, p.Kind, p.Mime, p.DeclaredBytes, p.StorageKey, p.B2FileID).Scan(&mediaID)
	if err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM pending_uploads WHERE id=$1`, id); err != nil {
		return 0, err
	}
	return mediaID, tx.Commit()
}

// DeletePendingUpload releases the reservation and hands back the B2 handle
// for the caller to cancel. Row first, cancel second: if the cancel then
// fails, the orphan is an unfinished B2 file with no row — precisely the shape
// the reaper's B2-side listing exists to find.
func (s *Store) DeletePendingUpload(ctx context.Context, id, accountID int64) (storageKey, b2FileID string, err error) {
	err = s.DB.QueryRowContext(ctx,
		`DELETE FROM pending_uploads WHERE id=$1 AND account_id=$2
		 RETURNING storage_key, b2_file_id`,
		id, accountID).Scan(&storageKey, &b2FileID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNoUpload
	}
	return
}

type stalePendingUpload struct {
	ID        int64
	AccountID int64
	B2FileID  string
}

// StalePendingUploads lists reservations older than the cutoff — abandoned
// browsers, closed laptops. The reaper cancels their B2 half and then deletes
// the rows, releasing the reserved quota.
func (s *Store) StalePendingUploads(ctx context.Context, olderThan time.Duration) ([]stalePendingUpload, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, account_id, b2_file_id FROM pending_uploads
		 WHERE created_at < now() - $1::interval`,
		olderThan.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []stalePendingUpload
	for rows.Next() {
		var s stalePendingUpload
		if err := rows.Scan(&s.ID, &s.AccountID, &s.B2FileID); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// PendingB2FileIDs is the reaper's join key: which unfinished B2 files have a
// live reservation governing them. Everything else B2 lists is an orphan
// candidate (age-guarded at the caller).
func (s *Store) PendingB2FileIDs(ctx context.Context) (map[string]bool, error) {
	rows, err := s.DB.QueryContext(ctx, `SELECT b2_file_id FROM pending_uploads`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}

// AccountPendingUploads lists an account's open reservations — account
// deletion cancels their B2 half before the row cascade, because "goodbye only
// after everything is gone" includes bytes B2 is holding for an upload that
// never finished.
func (s *Store) AccountPendingUploads(ctx context.Context, accountID int64) ([]stalePendingUpload, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, account_id, b2_file_id FROM pending_uploads WHERE account_id=$1`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []stalePendingUpload
	for rows.Next() {
		var s stalePendingUpload
		if err := rows.Scan(&s.ID, &s.AccountID, &s.B2FileID); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
