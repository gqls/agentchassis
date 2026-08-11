package actions

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// siteUsesFlatURLs reports whether the site's structure spec pins the flat
// URL shape for the nested page roles (data->>'url_shape' = 'flat'), i.e.
// whether CanonicalisePage should be called with PageDescriptor.FlatURLs set.
//
// ONE fact, ONE reader: both canonicalisation surfaces — WriteSitePlanAction
// (site_plan_pages) and SyncPagesToDBAction (pages) — must take the flag from
// this helper and nowhere else. The two surfaces diverged once before over a
// rule one of them applied and the other did not, and the divergence shipped
// a regression (see the ValidateRoles comment in SyncPagesToDBAction); a
// site-level fact read through two hand-rolled queries is the same hole.
// TestFlatURLFlagReachesBothCanonicalisationSurfaces pins this.
//
// Why the flag exists (bugs_open/241): running the planner over an adopted
// site that already serves flat tool/guide URLs silently rewrote them to the
// nested shape — upsertPage overwrites pages.url unconditionally and the
// deployer takes the file path from it. Measured on loancalculator.co.uk
// 2026-08-10: 24 of 26 live URLs would have moved.
//
// Absent spec, absent key, any value other than "flat", nil db, or a read
// error all mean false — the long-standing nested default. Nothing writes
// the key automatically; it is seeded per site into the structure aspect
// (supersede-then-insert, see any SEED_*.sql).
func siteUsesFlatURLs(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) bool {
	if db == nil {
		return false
	}
	var shape string
	err := db.QueryRowContext(ctx, `
		SELECT COALESCE(data->>'url_shape', '')
		FROM site_specs
		WHERE site_id = $1 AND aspect = 'structure' AND is_current = true
	`, siteID).Scan(&shape)
	switch {
	case err == sql.ErrNoRows:
		return false
	case err != nil:
		logger.Warn("siteUsesFlatURLs: structure spec read failed; defaulting to nested URLs",
			zap.String("site_id", siteID.String()),
			zap.Error(err))
		return false
	}
	return shape == "flat"
}
