package actions

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// seedAnalyticsDefaultSQL writes the network's default analytics container into a
// brand-new site's site_config spec, as {analytics: {gtm_container_id, mode:"default"}}.
//
// Why this exists: the head/header templates emit the GTM snippet from
// site_specs.site_config.analytics.gtm_container_id ({{if .gtm_container_id}} — STY-050),
// and nothing seeded that key at site creation, so every new site was born untagged and
// silently fell out of analytics (bugs_open/397 §6.2: 12+4 sites on 2026-08-26, 8 more by
// 2026-09-02 — ~one per day). The owner's standing instruction (2026-08-24) is that the
// estate tag is standard for new builds.
//
// The three guards, all load-bearing:
//   - the network must carry settings.analytics.gtm_container_id — no network default,
//     no seed. Tagging stays OPT-IN at the network level (the 2026-08-02 §2 ruling's
//     unsafe-default-OFF shape); a network for third-party/customer sites simply carries
//     no value (customer hosted copies use the separate GTM-TH5XGNQ4 mechanism —
//     webdesign_uk_build_service/DECISION_2026-08-26).
//   - only when the site has NO current site_config row at all. An existing row —
//     including {analytics:{mode:"none"}}, the explicit opt-out — is never touched and
//     never superseded, so this can run on every ensure-site pass (upsertSite is an
//     upsert) without ever creating a second current row or resurrecting a retracted tag.
//   - mode is stamped "default" so readers (and the ZIP export path) can tell a seeded
//     value from an operator-set one ("custom") or an opt-out ("none").
//   - the system/test pseudo-sites (system.internal carries platform-wide work items,
//     never pages) are excluded: an analytics row there is pure census pollution.
//     Measured 2026-09-03: the no-current-row population was 17 pool + 1 system + 1 test,
//     zero deployed/active — so the re-ensure path backfills only pool sites, at the
//     moment they are actually built.
const seedAnalyticsDefaultSQL = `
	INSERT INTO site_specs (site_id, aspect, data, source, created_by, notes, is_current)
	SELECT $1, 'site_config',
	       jsonb_build_object('analytics', jsonb_build_object(
	           'gtm_container_id', n.settings->'analytics'->>'gtm_container_id',
	           'mode', 'default')),
	       'system', 'seed-analytics-default',
	       'Network default analytics container, seeded at site creation (bugs_open/397 §6.2; owner 2026-08-24: standard for new builds). mode=default; operators set mode=custom with a site-specific id, or mode=none to opt out — backfills and this seeder honour an existing row.',
	       true
	  FROM networks n
	 WHERE n.id = $2
	   AND COALESCE(n.settings->'analytics'->>'gtm_container_id', '') <> ''
	   AND NOT EXISTS (
	         SELECT 1 FROM site_specs ss
	          WHERE ss.site_id = $1 AND ss.aspect = 'site_config' AND ss.is_current)
	   AND EXISTS (
	         SELECT 1 FROM sites st
	          WHERE st.id = $1 AND st.status NOT IN ('system', 'test'))
`

// seedAnalyticsDefault is best-effort by contract: a failure leaves the site exactly as
// unseeded as it was before this mechanism existed, and the census/backfill catches it.
// Callers must log, never fail site creation on it.
func seedAnalyticsDefault(ctx context.Context, db interface{}, siteID, networkID uuid.UUID, logger *zap.Logger) (bool, error) {
	switch d := db.(type) {
	case *sql.DB:
		res, err := d.ExecContext(ctx, seedAnalyticsDefaultSQL, siteID, networkID)
		if err != nil {
			return false, fmt.Errorf("seedAnalyticsDefault: %w", err)
		}
		n, _ := res.RowsAffected()
		return n > 0, nil
	case *pgxpool.Pool:
		ct, err := d.Exec(ctx, seedAnalyticsDefaultSQL, siteID, networkID)
		if err != nil {
			return false, fmt.Errorf("seedAnalyticsDefault: %w", err)
		}
		return ct.RowsAffected() > 0, nil
	default:
		return false, fmt.Errorf("seedAnalyticsDefault: unsupported database type %T", db)
	}
}
