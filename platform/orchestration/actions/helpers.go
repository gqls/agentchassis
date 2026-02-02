package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

// PageInfo holds page details for rendering
type PageInfo struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	MetaDesc    string    `json:"meta_desc"`
	Description string    `json:"description"`
}

func getDomainForSite(ctx context.Context, db *sql.DB, siteID uuid.UUID) (string, error) {
	var domain string
	err := db.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain)
	return domain, err
}
