package store

import (
	"context"
	"encoding/json"

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
	const q = `
		SELECT id, site_id, provocation, counter, verdict,
		       position_text, defence_text
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
