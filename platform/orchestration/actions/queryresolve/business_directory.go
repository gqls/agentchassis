// FILE: platform/orchestration/actions/queryresolve/business_directory.go
//
// query.business_directory — resolves a site's own verified business_intel
// directory for build-time rendering by the `directory-listing` component.
// Written for bugs_open/206 (vetcomparison.uk's directory-index page had a
// plan-scaffolded `entity-directory` page and a live, already-exporting data
// pipeline, but nothing turned the two into an actual page).
//
// Deliberately NOT parametrised by a static vertical arg (contrast
// `query.pages_where_type:tool`, whose arg is genuinely site-agnostic
// vocabulary). A business vertical is a per-site fact, and hardcoding one
// into this SHARED component's schema would make the component unusable for
// any other vertical. Instead this resolver looks up the SAME config
// `directory_export_action.go` already reads (its `scheduled_tasks` row, by
// domain), and runs the IDENTICAL filter `loadDirectoryEntries` uses — so the
// server-rendered listing and the client-fetchable exported JSON can never
// name a different set of businesses.
package queryresolve

import (
	"context"
	"database/sql"
	"fmt"
	"html"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// businessDirectoryConfig is the subset of a site's directory-export-json
// scheduled_tasks.input_data this resolver needs. Absence (no such task for
// this site) is not an error — it means the site has no business-directory
// data source configured yet, and the resolver returns an empty list so the
// component's `on_missing` handling applies exactly as it would for a
// genuinely empty query result.
type businessDirectoryConfig struct {
	Vertical          string
	BusinessTypeILike string
}

// HasBusinessDirectoryConfig reports whether siteID has a
// directory-json-exporter config row — the exact precondition
// resolveBusinessDirectory requires before it will return entries (it ERRORS
// without one, deliberately: bugs_open/206). Exported for plan-time validation
// (bugs_open/444) so the gate and the renderer share ONE predicate and cannot
// drift.
func HasBusinessDirectoryConfig(ctx context.Context, db *sql.DB, siteID uuid.UUID) (bool, error) {
	_, ok, err := lookupBusinessDirectoryConfig(ctx, db, siteID)
	return ok, err
}

// lookupBusinessDirectoryConfig finds the exporter config for siteID's own
// domain. Matches by domain (scheduled_tasks carries no site_id column) and
// by target_agent_type (there is exactly one such task fleet-wide today,
// vetcomparison.uk's `directory-export-json`, but the lookup is written to
// support a second site adding its own row under a different `name`).
func lookupBusinessDirectoryConfig(ctx context.Context, db *sql.DB, siteID uuid.UUID) (businessDirectoryConfig, bool, error) {
	var cfg businessDirectoryConfig
	var vertical, businessTypeILike sql.NullString
	err := db.QueryRowContext(ctx, `
		SELECT st.input_data->>'vertical', st.input_data->>'business_type_ilike'
		FROM scheduled_tasks st
		JOIN sites s ON s.domain = st.input_data->>'domain'
		WHERE s.id = $1
		  AND st.target_agent_type = 'directory-json-exporter'
		LIMIT 1
	`, siteID).Scan(&vertical, &businessTypeILike)
	if err == sql.ErrNoRows {
		return cfg, false, nil
	}
	if err != nil {
		return cfg, false, err
	}
	if !vertical.Valid || vertical.String == "" {
		return cfg, false, nil
	}
	cfg.Vertical = vertical.String
	cfg.BusinessTypeILike = businessTypeILike.String
	return cfg, true, nil
}

// resolveBusinessDirectory returns verified businesses for siteID's own
// configured vertical, projected and escaped for a text/template render
// (this package's templates do not auto-escape — see news_items.go).
//
// Filter mirrors directory_export_action.go's loadDirectoryEntries EXACTLY
// (verified, website_url + postcode present, matching vertical [+ optional
// business_type_ilike]) so a business appearing in the rendered page and one
// appearing in the exported JSON are always the same set.
//
// Capped for build-time SSR — this is a listing SECTION on one page, not the
// full client-searchable archive (that already exists as the exported JSON
// file; a client-side search UI against it is separate, unbuilt, out of
// scope here). Default 60, hard cap 100 — larger than a card grid's 12/24
// because a directory page's whole purpose is the list, but still bounded.
func resolveBusinessDirectory(ctx context.Context, db *sql.DB, siteID uuid.UUID, limit int, logger *zap.Logger) (interface{}, error) {
	if siteID == uuid.Nil {
		return nil, fmt.Errorf("resolveBusinessDirectory: empty site_id")
	}

	const hardCap = 100
	if limit <= 0 {
		limit = 60
	}
	if limit > hardCap {
		limit = hardCap
	}

	cfg, ok, err := lookupBusinessDirectoryConfig(ctx, db, siteID)
	if err != nil {
		return nil, fmt.Errorf("resolveBusinessDirectory: config lookup failed: %w", err)
	}
	if !ok {
		// council review (bugs_open/206, round 1, bug_historian — gating):
		// this must NOT look like "zero eligible businesses", which is a
		// legitimate empty result plan_sections already handles via
		// min_items/on_missing. A missing exporter config is a
		// MISCONFIGURATION (domain drift, vertical rename, a config edit
		// that broke the site<->vertical link) and must surface as a loud
		// failure, not a silent hollow section that looks like success.
		return nil, fmt.Errorf("resolveBusinessDirectory: no directory-export-json config found for site %s — cannot distinguish this from a real zero-business result, refusing rather than rendering an unexplained empty directory", siteID)
	}

	query := `
		SELECT COALESCE(NULLIF(b.trading_name, ''), b.name) AS name,
		       b.postcode,
		       NULLIF(TRIM(BOTH ', ' FROM CONCAT_WS(', ', b.town, b.county)), ''),
		       b.website_url,
		       COALESCE(b.is_claimed, false)
		FROM business_intel.businesses b
		JOIN business_intel.business_verticals v ON v.id = b.vertical_id
		WHERE v.slug = $1
		  AND b.verification_status = 'verified'
		  AND b.website_url IS NOT NULL
		  AND b.postcode IS NOT NULL`
	args := []interface{}{cfg.Vertical}
	// Claimed listings sort ahead of unclaimed within the SSR cap. The cap
	// means ordering IS visibility: alphabetical order silently excluded any
	// claimed listing sorting past the cut, which inverted the product's own
	// promise (claiming is the route to being seen). Ties stay alphabetical.
	if cfg.BusinessTypeILike != "" {
		query += ` AND b.business_type ILIKE $2 ORDER BY COALESCE(b.is_claimed, false) DESC, name LIMIT $3`
		args = append(args, cfg.BusinessTypeILike, limit)
	} else {
		query += ` ORDER BY COALESCE(b.is_claimed, false) DESC, name LIMIT $2`
		args = append(args, limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("resolveBusinessDirectory query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, postcode, website string
		var location sql.NullString
		var isClaimed bool
		if err := rows.Scan(&name, &postcode, &location, &website, &isClaimed); err != nil {
			logger.Warn("resolveBusinessDirectory: scan failed", zap.Error(err))
			continue
		}
		items = append(items, projectBusinessDirectoryRow(name, postcode, location.String, website, isClaimed))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolveBusinessDirectory rows iter failed: %w", err)
	}

	logger.Info("queryresolve: resolved business_directory",
		zap.String("vertical", cfg.Vertical),
		zap.String("site_id", siteID.String()),
		zap.Int("count", len(items)),
		zap.Int("limit", limit),
	)

	return items, nil
}

// projectBusinessDirectoryRow is the pure, DB-free projection step — same
// split as news_items.go's projectNewsItems, so the escaping contract is
// unit-testable without a database.
func projectBusinessDirectoryRow(name, postcode, location, website string, isClaimed bool) map[string]interface{} {
	return map[string]interface{}{
		"name":       html.EscapeString(name),
		"postcode":   html.EscapeString(postcode),
		"location":   html.EscapeString(location),
		"website":    html.EscapeString(website),
		"is_claimed": isClaimed,
	}
}
