package store

import (
	"context"
	"errors"
	"net/url"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Site holds the fields returned by ActiveSiteByOrigin.
type Site struct {
	ID     string
	Domain string
}

// ActiveSiteByOrigin resolves an HTTP Origin header value to a deployed site.
// It strips the scheme and port from origin to get a bare hostname, then queries
// sites WHERE status = 'deployed' AND domain = $1 ORDER BY id LIMIT 1.
// Returns nil, nil when no matching site exists (caller treats as CORS-reject).
// Returns a non-nil error only on unexpected DB failure.
func ActiveSiteByOrigin(ctx context.Context, pool *pgxpool.Pool, origin string) (*Site, error) {
	u, err := url.Parse(origin)
	if err != nil || u.Hostname() == "" {
		// Unparseable or empty origin — treat as no match.
		return nil, nil
	}
	host := u.Hostname()

	var s Site
	err = pool.QueryRow(
		ctx,
		`SELECT id, domain FROM sites WHERE status = 'deployed' AND domain = $1 ORDER BY id LIMIT 1`,
		host,
	).Scan(&s.ID, &s.Domain)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}
