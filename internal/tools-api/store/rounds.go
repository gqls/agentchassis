package store

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Round is the in-memory representation of a gauntlet_rounds row.
type Round struct {
	ID           string
	SiteID       string
	Provocation  json.RawMessage
	Counter      json.RawMessage
	Verdict      json.RawMessage
	PositionText string
	DefenceText  string
}

// CreateRound inserts a new gauntlet_rounds row and returns it with the
// database-generated id.
func CreateRound(
	ctx context.Context,
	pool *pgxpool.Pool,
	siteID string,
	provocation json.RawMessage,
	ipHash string,
) (*Round, error) {
	const q = `
		INSERT INTO gauntlet_rounds (site_id, provocation, client_ip_hash)
		VALUES ($1, $2, $3)
		RETURNING id, site_id`

	row := pool.QueryRow(ctx, q, siteID, []byte(provocation), ipHash)

	var r Round
	r.Provocation = provocation
	if err := row.Scan(&r.ID, &r.SiteID); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetRound fetches a single gauntlet_rounds row by its primary key.
func GetRound(
	ctx context.Context,
	pool *pgxpool.Pool,
	roundID string,
) (*Round, error) {
	// COALESCE the text columns: they are NULL on every fresh round and pgx
	// cannot scan NULL into string — without this, GetRound errored on every
	// round that had not yet stored a position/defence, which the handlers
	// then reported as "round not found" (found at first island smoke,
	// 2026-07-25). The jsonb columns scan into []byte, where NULL is fine.
	const q = `
		SELECT id, site_id, provocation, counter, verdict,
		       COALESCE(position_text, ''), COALESCE(defence_text, '')
		FROM   gauntlet_rounds
		WHERE  id = $1`

	row := pool.QueryRow(ctx, q, roundID)

	var r Round
	var prov, counter, verdict []byte
	if err := row.Scan(
		&r.ID, &r.SiteID,
		&prov, &counter, &verdict,
		&r.PositionText, &r.DefenceText,
	); err != nil {
		return nil, err
	}
	r.Provocation = json.RawMessage(prov)
	r.Counter = json.RawMessage(counter)
	r.Verdict = json.RawMessage(verdict)
	return &r, nil
}

// UpdateRoundPosition persists the user's position_text and the AI-generated
// counter (counter_position + challenge) onto an existing round row.
func UpdateRoundPosition(
	ctx context.Context,
	pool *pgxpool.Pool,
	roundID string,
	positionText string,
	counter json.RawMessage,
) error {
	const q = `
		UPDATE gauntlet_rounds
		SET    position_text = $2,
		       counter      = $3,
		       updated_at   = now()
		WHERE  id = $1`

	_, err := pool.Exec(ctx, q, roundID, positionText, []byte(counter))
	return err
}

// UpdateRoundVerdict persists the user's defence_text and the AI-generated
// verdict (verdict + reasons) onto an existing round row.
func UpdateRoundVerdict(
	ctx context.Context,
	pool *pgxpool.Pool,
	roundID string,
	defenceText string,
	verdict json.RawMessage,
) error {
	const q = `
		UPDATE gauntlet_rounds
		SET    defence_text = $2,
		       verdict     = $3,
		       updated_at  = now()
		WHERE  id = $1`

	_, err := pool.Exec(ctx, q, roundID, defenceText, []byte(verdict))
	return err
}

// ── publication (migration 276) ──────────────────────────────────────────────
//
// A round becomes a public, linkable record only when the visitor who argued it
// presses share (owner ruling 2026-07-31). Nothing here backfills or defaults to
// published: every round already stored is harness traffic and stays private.

// ErrRoundNotPublishable is returned when the round exists but has no verdict,
// so there is no complete argument to publish. Distinguished from "not found" so
// the handler can say which it was — a 404 for a round the visitor is looking at
// on screen would be a lie.
var ErrRoundNotPublishable = errors.New("round has no verdict")

// SlugAlphabet has exactly 32 characters, which divides 256 evenly, so reducing
// a random byte with % introduces NO modulo bias. It also omits 0/O/1/l/I so a
// slug read off a share card cannot be mistyped into a different round.
//
// Exported, with ValidSlug below, so that the generator and the validator have
// ONE definition between them. They were briefly two — a regexp in the handler
// spelling out the same characters — which is the drift class that produces a
// silent total failure: widen the alphabet here, and every freshly published
// round 404s on read while both halves still look obviously correct.
const SlugAlphabet = "abcdefghjkmnpqrstuvwxyz23456789_"

// SlugLength is the token length. 32^10 ≈ 2^50, which is what keeps a published
// round's URL from being reachable by enumeration.
const SlugLength = 10

// ValidSlug reports whether s could have been produced by newSlug. Callers use it
// to reject a malformed slug before it costs a database round-trip.
func ValidSlug(s string) bool {
	if len(s) != SlugLength {
		return false
	}
	for _, r := range s {
		if !strings.ContainsRune(SlugAlphabet, r) {
			return false
		}
	}
	return true
}

// newSlug returns a SlugLength token from crypto/rand.
// It returns an error rather than falling back to a weak source: a predictable
// "fallback" token would look identical to a good one in the database.
func newSlug() (string, error) {
	b := make([]byte, SlugLength)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", fmt.Errorf("generate slug: %w", err)
	}
	for i := range b {
		b[i] = SlugAlphabet[int(b[i])%len(SlugAlphabet)]
	}
	return string(b), nil
}

// PublishRound marks a completed round public and returns its slug.
//
// Idempotent by design: pressing share twice must not mint a second slug or move
// published_at, because the card the visitor already has carries the first URL.
// The UPDATE therefore only fires when published_at IS NULL, and the RETURNING
// value is read back either way.
//
// siteID scopes the write: the round id arrives from the browser, so without it
// one site's page could publish another site's round.
func PublishRound(
	ctx context.Context,
	pool *pgxpool.Pool,
	siteID string,
	roundID string,
) (string, error) {
	// Distinguish "no such round for this site" from "not finished yet" before
	// attempting the write, so the caller can explain which happened.
	const check = `
		SELECT verdict IS NOT NULL, COALESCE(public_slug, '')
		FROM   gauntlet_rounds
		WHERE  id = $1 AND site_id = $2`

	var hasVerdict bool
	var existing string
	if err := pool.QueryRow(ctx, check, roundID, siteID).Scan(&hasVerdict, &existing); err != nil {
		return "", err // pgx.ErrNoRows -> handler answers 404
	}
	if existing != "" {
		return existing, nil // already published; same URL as before
	}
	if !hasVerdict {
		return "", ErrRoundNotPublishable
	}

	// Retry only on a slug collision. At 32^10 that is vanishingly unlikely, but
	// a UNIQUE violation must not surface to a visitor as a failed share.
	for attempt := 0; attempt < 5; attempt++ {
		slug, err := newSlug()
		if err != nil {
			return "", err
		}

		const q = `
			UPDATE gauntlet_rounds
			SET    public_slug  = $3,
			       published_at = now(),
			       updated_at   = now()
			WHERE  id = $1 AND site_id = $2
			  AND  verdict IS NOT NULL
			  AND  published_at IS NULL
			RETURNING public_slug`

		var out string
		err = pool.QueryRow(ctx, q, roundID, siteID, slug).Scan(&out)
		if err == nil {
			return out, nil
		}
		// Another request published it between the check and here: return the
		// slug that won rather than reporting an error the visitor cannot act on.
		if errors.Is(err, pgx.ErrNoRows) {
			var raced string
			if e := pool.QueryRow(ctx,
				`SELECT COALESCE(public_slug,'') FROM gauntlet_rounds WHERE id=$1 AND site_id=$2`,
				roundID, siteID).Scan(&raced); e == nil && raced != "" {
				return raced, nil
			}
			return "", ErrRoundNotPublishable
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			continue
		}
		return "", err
	}
	return "", errors.New("could not allocate a unique slug")
}

// PublicRound is the publishable projection of a round — deliberately a separate
// type from Round, so adding a column to the table cannot silently widen what the
// public endpoint serves. client_ip_hash and site_id are absent BY DESIGN.
type PublicRound struct {
	Slug        string          `json:"slug"`
	PublishedAt time.Time       `json:"published_at"`
	ArguedAt    time.Time       `json:"argued_at"`
	Provocation json.RawMessage `json:"provocation"`
	Position    string          `json:"position"`
	Counter     json.RawMessage `json:"counter"`
	Defence     string          `json:"defence"`
	Verdict     json.RawMessage `json:"verdict"`
}

// GetPublishedRound fetches a round by its public slug, and ONLY if it has been
// published. An unpublished or unknown slug is pgx.ErrNoRows, which the handler
// answers 404 — the two are indistinguishable to a caller on purpose, so the
// endpoint cannot be used to discover that an unpublished round exists.
func GetPublishedRound(
	ctx context.Context,
	pool *pgxpool.Pool,
	siteID string,
	slug string,
) (*PublicRound, error) {
	const q = `
		SELECT public_slug, published_at, created_at,
		       provocation, COALESCE(position_text,''), counter,
		       COALESCE(defence_text,''), verdict
		FROM   gauntlet_rounds
		WHERE  public_slug = $1
		  AND  site_id     = $2
		  AND  published_at IS NOT NULL
		  AND  verdict      IS NOT NULL`

	var r PublicRound
	var prov, counter, verdict []byte
	if err := pool.QueryRow(ctx, q, slug, siteID).Scan(
		&r.Slug, &r.PublishedAt, &r.ArguedAt,
		&prov, &r.Position, &counter, &r.Defence, &verdict,
	); err != nil {
		return nil, err
	}
	r.Provocation = json.RawMessage(prov)
	r.Counter = json.RawMessage(counter)
	r.Verdict = json.RawMessage(verdict)
	return &r, nil
}
