package main

// store.go — every database access lives here, so the SQL is reviewable in one
// place and no handler can quietly invent its own query.
//
// THE RULE THIS FILE EXISTS TO ENFORCE: every read and every write is scoped by
// account_id, in the SQL, always. Not "the handler checks first" — in the WHERE
// clause. A notes service's whole job is that one person's writing is not
// another's, and the only version of that guarantee which survives someone
// adding a handler in a hurry is the one the query itself carries.

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	_ "embed"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//go:embed schema.sql
var schemaSQL string

// PBKDF2-HMAC-SHA256 at OWASP's recommended iteration count for that KDF.
// Deliberately stdlib (Go 1.24's crypto/pbkdf2) rather than argon2 from
// x/crypto: it keeps this binary's dependency list to the database driver
// alone, which matters on a box that also serves a live commercial site.
const (
	pbkdf2Iters  = 600_000
	pbkdf2KeyLen = 32
	saltLen      = 16
)

var (
	ErrNoAccount     = errors.New("no such account")
	ErrEmailTaken    = errors.New("email already registered")
	ErrBadPassword   = errors.New("incorrect password")
	ErrQuotaExceeded = errors.New("media quota exceeded")
)

type Store struct {
	DB         *sql.DB
	QuotaBytes int64
}

type Account struct {
	ID         int64
	Email      string
	MediaBytes int64
}

type Note struct {
	ID        int64     `json:"id"`
	ClientID  string    `json:"client_id,omitempty"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// Board arrangement (stage 2). nil = never arranged. In a save payload:
	// absent keeps whatever the row holds (an old client cannot erase an
	// arrangement it does not know about); an explicit JSON null clears it.
	Layout json.RawMessage `json:"layout,omitempty"`
	Audio  []Media         `json:"audio,omitempty"`
	Images []Media         `json:"images,omitempty"`
	// Every kind, in upload order — the shape the editor (and the coming
	// pasteboard) consumes. Audio/Images above remain for anything that
	// already reads the grouped form.
	Media []Media `json:"media,omitempty"`
}

type Media struct {
	ID      int64  `json:"id"`
	Kind    string `json:"kind"`
	Mime    string `json:"mime"`
	ByteLen int64  `json:"byte_len"`
	Caption string `json:"caption,omitempty"`
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.DB.ExecContext(ctx, schemaSQL)
	return err
}

// --- accounts -------------------------------------------------------------

func canonicalEmail(e string) string { return strings.ToLower(strings.TrimSpace(e)) }

func hashPassword(pw string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, pw, salt, pbkdf2Iters, pbkdf2KeyLen)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%s$%s", pbkdf2Iters,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key)), nil
}

func verifyPassword(encoded, pw string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	iters, err := strconv.Atoi(parts[1])
	if err != nil || iters < 1 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[2])
	want, err2 := base64.RawStdEncoding.DecodeString(parts[3])
	if err1 != nil || err2 != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pw, salt, iters, len(want))
	if err != nil {
		return false
	}
	// Constant time: a timing difference here leaks whether a guess was close.
	return subtle.ConstantTimeCompare(got, want) == 1
}

func (s *Store) CreateAccount(ctx context.Context, email, password string) (*Account, error) {
	hash, err := hashPassword(password)
	if err != nil {
		return nil, err
	}
	var a Account
	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO accounts (email, email_canonical, password_hash)
		 VALUES ($1,$2,$3) RETURNING id, email, media_bytes`,
		strings.TrimSpace(email), canonicalEmail(email), hash).
		Scan(&a.ID, &a.Email, &a.MediaBytes)
	if err != nil {
		if strings.Contains(err.Error(), "email_canonical") || strings.Contains(err.Error(), "23505") {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return &a, nil
}

func (s *Store) Authenticate(ctx context.Context, email, password string) (*Account, error) {
	var a Account
	var hash string
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, email, media_bytes, password_hash FROM accounts WHERE email_canonical=$1`,
		canonicalEmail(email)).Scan(&a.ID, &a.Email, &a.MediaBytes, &hash)
	if errors.Is(err, sql.ErrNoRows) {
		// Spend the same work as a real verification would. Returning early
		// here makes "no such account" measurably faster than "wrong password",
		// which turns this endpoint into an account-existence oracle.
		_, _ = pbkdf2.Key(sha256.New, password, []byte("absent-account-timing-equaliser"), pbkdf2Iters, pbkdf2KeyLen)
		return nil, ErrNoAccount
	}
	if err != nil {
		return nil, err
	}
	if !verifyPassword(hash, password) {
		return nil, ErrBadPassword
	}
	_, _ = s.DB.ExecContext(ctx, `UPDATE accounts SET last_login_at=now() WHERE id=$1`, a.ID)
	return &a, nil
}

// --- sessions -------------------------------------------------------------

// Returns the raw token for the cookie; only its hash is stored, so a database
// leak does not hand over live sessions.
func (s *Store) CreateSession(ctx context.Context, accountID int64, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, account_id, expires_at) VALUES ($1,$2,$3)`,
		hex.EncodeToString(sum[:]), accountID, time.Now().Add(ttl))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Store) AccountForSession(ctx context.Context, token string) (*Account, error) {
	sum := sha256.Sum256([]byte(token))
	var a Account
	err := s.DB.QueryRowContext(ctx,
		`SELECT a.id, a.email, a.media_bytes
		   FROM sessions s JOIN accounts a ON a.id = s.account_id
		  WHERE s.token_hash=$1 AND s.expires_at > now()`,
		hex.EncodeToString(sum[:])).Scan(&a.ID, &a.Email, &a.MediaBytes)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoAccount
	}
	return &a, err
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	sum := sha256.Sum256([]byte(token))
	_, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash=$1`, hex.EncodeToString(sum[:]))
	return err
}

func (s *Store) PurgeExpiredSessions(ctx context.Context) (int64, error) {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at < now()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// --- notes ----------------------------------------------------------------

func (s *Store) ListNotes(ctx context.Context, accountID int64) ([]Note, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT id, COALESCE(client_id,''), title, content, created_at, updated_at,
		        COALESCE(layout::text,'')
		   FROM notes WHERE account_id=$1 AND deleted_at IS NULL
		  ORDER BY updated_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notes := []Note{}
	byID := map[int64]int{}
	for rows.Next() {
		var n Note
		var layout string
		if err := rows.Scan(&n.ID, &n.ClientID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt, &layout); err != nil {
			return nil, err
		}
		if layout != "" {
			n.Layout = json.RawMessage(layout)
		}
		byID[n.ID] = len(notes)
		notes = append(notes, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Media metadata in one query rather than N — the bytes themselves are
	// fetched separately, per item, so a note list never carries megabytes.
	// Ordered by id, which is insert order both within a kind (ordering and
	// id advance together) and across kinds — the unified array wants the
	// order things were added, not audio-then-images.
	mrows, err := s.DB.QueryContext(ctx,
		`SELECT id, note_id, kind, mime, byte_len, COALESCE(caption,'') FROM media
		  WHERE account_id=$1 ORDER BY note_id, id`, accountID)
	if err != nil {
		return nil, err
	}
	defer mrows.Close()
	for mrows.Next() {
		var m Media
		var noteID int64
		if err := mrows.Scan(&m.ID, &noteID, &m.Kind, &m.Mime, &m.ByteLen, &m.Caption); err != nil {
			return nil, err
		}
		if idx, ok := byID[noteID]; ok {
			notes[idx].Media = append(notes[idx].Media, m)
			switch m.Kind {
			case "audio":
				notes[idx].Audio = append(notes[idx].Audio, m)
			case "image":
				notes[idx].Images = append(notes[idx].Images, m)
			}
		}
	}
	return notes, mrows.Err()
}

func (s *Store) SaveNote(ctx context.Context, accountID int64, n Note) (*Note, error) {
	// Layout semantics: nil (absent from the payload) preserves the stored
	// value; the literal JSON null clears it; anything else replaces it.
	var layoutArg any
	if n.Layout != nil {
		layoutArg = string(n.Layout)
	}
	var out Note
	var layoutOut string
	if n.ID > 0 {
		err := s.DB.QueryRowContext(ctx,
			`UPDATE notes SET title=$1, content=$2, updated_at=now(),
			        layout = CASE WHEN $5::text IS NULL THEN layout
			                      WHEN $5::text = 'null' THEN NULL
			                      ELSE $5::jsonb END
			  WHERE id=$3 AND account_id=$4 AND deleted_at IS NULL
			  RETURNING id, COALESCE(client_id,''), title, content, created_at, updated_at, COALESCE(layout::text,'')`,
			n.Title, n.Content, n.ID, accountID, layoutArg).
			Scan(&out.ID, &out.ClientID, &out.Title, &out.Content, &out.CreatedAt, &out.UpdatedAt, &layoutOut)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoAccount // not yours, or not there — same answer either way
		}
		if layoutOut != "" {
			out.Layout = json.RawMessage(layoutOut)
		}
		return &out, err
	}

	var clientID any
	if n.ClientID != "" {
		clientID = n.ClientID
	}
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO notes (account_id, client_id, title, content, layout)
		 VALUES ($1,$2,$3,$4, CASE WHEN $5::text IS NULL OR $5::text='null' THEN NULL ELSE $5::jsonb END)
		 ON CONFLICT (account_id, client_id) WHERE client_id IS NOT NULL
		 DO UPDATE SET title=EXCLUDED.title, content=EXCLUDED.content, updated_at=now(),
		               layout = CASE WHEN $5::text IS NULL THEN notes.layout
		                             WHEN $5::text = 'null' THEN NULL
		                             ELSE $5::jsonb END
		 RETURNING id, COALESCE(client_id,''), title, content, created_at, updated_at, COALESCE(layout::text,'')`,
		accountID, clientID, n.Title, n.Content, layoutArg).
		Scan(&out.ID, &out.ClientID, &out.Title, &out.Content, &out.CreatedAt, &out.UpdatedAt, &layoutOut)
	if layoutOut != "" {
		out.Layout = json.RawMessage(layoutOut)
	}
	return &out, err
}

func (s *Store) DeleteNote(ctx context.Context, accountID, noteID int64) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE notes SET deleted_at=now() WHERE id=$1 AND account_id=$2 AND deleted_at IS NULL`,
		noteID, accountID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNoAccount
	}
	return nil
}

// --- media ----------------------------------------------------------------

// AddMediaB2 records a media row whose bytes already sit in B2 under
// storageKey/fileID. Same transactional quota as AddMedia; no bytes column.
// If the quota refuses, the CALLER deletes the just-uploaded B2 object — this
// function has no B2 client on purpose (the store stays database-only).
func (s *Store) AddMediaB2(ctx context.Context, accountID, noteID int64, kind, mime string, byteLen int, storageKey, fileID string) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var used int64
	if err := tx.QueryRowContext(ctx,
		`SELECT media_bytes FROM accounts WHERE id=$1 FOR UPDATE`, accountID).Scan(&used); err != nil {
		return 0, err
	}
	if used+int64(byteLen) > s.QuotaBytes {
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
		`INSERT INTO media (note_id, account_id, kind, mime, bytes, byte_len, storage_key, b2_file_id, ordering)
		 VALUES ($1,$2,$3,$4,NULL,$5,$6,$7,
		         COALESCE((SELECT MAX(ordering)+1 FROM media WHERE note_id=$1 AND kind=$3),0))
		 RETURNING id`,
		noteID, accountID, kind, mime, byteLen, storageKey, fileID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// AddMedia enforces the per-account quota inside the same transaction as the
// insert. Checking the quota in the handler and inserting afterwards would let
// two concurrent uploads each see room and both proceed.
func (s *Store) AddMedia(ctx context.Context, accountID, noteID int64, kind, mime string, data []byte) (int64, error) {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Lock the account row so the quota read cannot be stale by the time we insert.
	var used int64
	if err := tx.QueryRowContext(ctx,
		`SELECT media_bytes FROM accounts WHERE id=$1 FOR UPDATE`, accountID).Scan(&used); err != nil {
		return 0, err
	}
	if used+int64(len(data)) > s.QuotaBytes {
		return 0, ErrQuotaExceeded
	}

	// Scoped by account_id as well as id: a note id from another account must
	// not be attachable to, even with a valid session.
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
		`INSERT INTO media (note_id, account_id, kind, mime, bytes, byte_len, ordering)
		 VALUES ($1,$2,$3,$4,$5,$6,
		         COALESCE((SELECT MAX(ordering)+1 FROM media WHERE note_id=$1 AND kind=$3),0))
		 RETURNING id`,
		noteID, accountID, kind, mime, data, len(data)).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// DeleteMedia is account-scoped in the SQL like every other query here; the
// media_bytes trigger hands the freed bytes back to the quota.
func (s *Store) DeleteMedia(ctx context.Context, accountID, mediaID int64) error {
	res, err := s.DB.ExecContext(ctx,
		`DELETE FROM media WHERE id=$1 AND account_id=$2`, mediaID, accountID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNoAccount
	}
	return nil
}

// GetMedia returns the row's mime plus EITHER inline bytes (storageKey empty)
// OR the B2 location. Account-scoped in the SQL like everything here.
func (s *Store) GetMedia(ctx context.Context, accountID, mediaID int64) (mime string, data []byte, storageKey string, err error) {
	err = s.DB.QueryRowContext(ctx,
		`SELECT mime, bytes, COALESCE(storage_key,'') FROM media WHERE id=$1 AND account_id=$2`,
		mediaID, accountID).Scan(&mime, &data, &storageKey)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil, "", ErrNoAccount
	}
	return
}

// SetMediaCaption is account-scoped in the SQL like every other write here.
func (s *Store) SetMediaCaption(ctx context.Context, accountID, mediaID int64, caption string) error {
	res, err := s.DB.ExecContext(ctx,
		`UPDATE media SET caption=$1 WHERE id=$2 AND account_id=$3`, caption, mediaID, accountID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNoAccount
	}
	return nil
}

// MediaStorageRef names one B2 object an account owns.
type MediaStorageRef struct{ Key, FileID string }

// ListAccountMediaStorage returns every B2-backed object the account owns —
// the delete-account path removes these from B2 BEFORE the row cascade, so an
// object can never outlive its account invisibly (it is paid storage).
func (s *Store) ListAccountMediaStorage(ctx context.Context, accountID int64) ([]MediaStorageRef, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT storage_key, COALESCE(b2_file_id,'') FROM media
		  WHERE account_id=$1 AND storage_key IS NOT NULL AND storage_key <> ''`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MediaStorageRef
	for rows.Next() {
		var r MediaStorageRef
		if err := rows.Scan(&r.Key, &r.FileID); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// DeleteAccount removes the account row; sessions, notes and media rows go
// with it (ON DELETE CASCADE — rows cannot half-survive). B2 objects are the
// CALLER's responsibility, before this is called.
func (s *Store) DeleteAccount(ctx context.Context, accountID int64) error {
	res, err := s.DB.ExecContext(ctx, `DELETE FROM accounts WHERE id=$1`, accountID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNoAccount
	}
	return nil
}

// MediaStorage returns where a row's bytes live, for the delete path.
func (s *Store) MediaStorage(ctx context.Context, accountID, mediaID int64) (storageKey, fileID string, err error) {
	err = s.DB.QueryRowContext(ctx,
		`SELECT COALESCE(storage_key,''), COALESCE(b2_file_id,'') FROM media WHERE id=$1 AND account_id=$2`,
		mediaID, accountID).Scan(&storageKey, &fileID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNoAccount
	}
	return
}
