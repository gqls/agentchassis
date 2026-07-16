// FILE: platform/orchestration/actions/queryresolve/queryresolve.go
//
// Package queryresolve runs named queries against the database for `query.*`
// source resolution in component input_schemas. Components that show lists
// (tool-list, blog-listing, directory-listing, etc.) declare their items
// field with `source: "query.<name>"`, and this package executes that
// resolution.
//
// Initial scope (v1, focus doc FOCUS_directory_builder_and_list_components.md):
//   - Concrete queries: pages_where_type:<type>, pages_under_section:<area>
//   - Called inline from plan_sections_action.go's source resolver
//   - Returns []map[string]interface{} suitable for the "items" field of an
//     array-typed schema field
//
// The resolver is a separate package so that a future directory-builder
// agent (per doc 002) can call the same Resolve() entry point without
// changing the resolver logic.

package queryresolve

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/storage"
	"go.uber.org/zap"
)

// QueryRequest describes a single resolution call.
//
// Name is the suffix after "query." in the source declaration. For
// `source: "query.pages_where_type:tool"`, Name is `pages_where_type:tool`.
//
// SiteID scopes the query to a particular site's data. Every query in this
// package is site-scoped — there are no cross-site queries.
//
// Limit (when > 0) caps the number of rows returned. The component schema
// can specify a limit; if zero, the query's default applies.
type QueryRequest struct {
	Name   string
	SiteID uuid.UUID
	Limit  int
}

// Resolve dispatches the request to a registered query handler. Returns
// nil + error on unknown query name or DB failure. Returns the resolved
// data on success — currently always []map[string]interface{} but the
// signature is interface{} to allow future query types that return other
// shapes (e.g. a single object, a count).
func Resolve(ctx context.Context, db *sql.DB, req QueryRequest, logger *zap.Logger) (interface{}, error) {
	if db == nil {
		return nil, fmt.Errorf("queryresolve.Resolve: nil db")
	}
	if req.SiteID == uuid.Nil {
		return nil, fmt.Errorf("queryresolve.Resolve: empty site_id")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("queryresolve.Resolve: empty query name")
	}

	// Normalise: lowercase, trim whitespace. Query names are case-insensitive
	// in source declarations to be forgiving (component-creator's prompt may
	// vary capitalisation).
	name := strings.ToLower(strings.TrimSpace(req.Name))

	// Parse the name into base + parameter. The vocabulary uses one of two
	// shapes:
	//   pages_where_type:tool        → base "pages_where_type", arg "tool"
	//   pages_under_section:guides   → base "pages_under_section", arg "guides"
	//   site_specs.case_studies      → not a query (handled by site_specs source)
	//
	// Names without `:` are treated as the whole name being the base, with
	// no argument.
	base, arg := parseQueryName(name)

	switch base {
	case "pages_where_type":
		return resolvePagesWhereType(ctx, db, req.SiteID, arg, req.Limit, logger)

	case "pages_under_section":
		return resolvePagesUnderSection(ctx, db, req.SiteID, arg, req.Limit, logger)

	case "section_index_for":
		return resolveSectionIndexForType(ctx, db, req.SiteID, arg, logger)

	case "products":
		return resolveProducts(ctx, db, req.SiteID, arg, req.Limit, logger)

	case "blog_posts":
		// Article listings (content-listing, blog-listing components declare
		// `source: "query.blog_posts"`). Fleet convention: articles are pages
		// with page_type 'blog-post', so this is pages_where_type with the
		// type fixed — one vocabulary entry, zero new query machinery.
		return resolvePagesWhereType(ctx, db, req.SiteID, "blog-post", req.Limit, logger)

	default:
		return nil, fmt.Errorf("queryresolve.Resolve: unknown query name %q (base %q)", req.Name, base)
	}
}

// parseQueryName splits a query name into base and argument. The first
// colon is the separator; subsequent colons are part of the argument
// (rarely needed but harmless).
func parseQueryName(name string) (base, arg string) {
	idx := strings.Index(name, ":")
	if idx < 0 {
		return name, ""
	}
	return name[:idx], name[idx+1:]
}

// pageImageProjection / pageImageJoins are the shared SQL fragments that give
// every page-listing query its item image (Phase I3, Lane B). Two candidates,
// in preference order:
//   - ca: the page's entity-linked CARD asset (assets.entity_type='page' +
//     entity_id, purpose 'card') — the purpose-built listing crop.
//   - ha: the page's own Lane A plan hero (current plan, page scope) — the
//     always-present fallback until a card is derived.
//
// Alias `p` for pages is required in the enclosing query. The lateral hero
// lookup runs per returned row; listings are capped at 24 rows so this stays
// cheap.
const pageImageProjection = `
		    COALESCE(ca.asset_key, '') AS card_key,
		    COALESCE(ha.asset_key, '') AS hero_key,
		    COALESCE(ha.purpose, '')   AS hero_purpose`

const pageImageJoins = `
		LEFT JOIN assets ca
		  ON ca.site_id = p.site_id AND ca.entity_type = 'page'
		 AND ca.entity_id = p.id AND ca.purpose = 'card' AND ca.status = 'active'
		LEFT JOIN LATERAL (
		    SELECT a.asset_key, a.purpose
		      FROM site_plan_imagery spi
		      JOIN site_plans sp ON sp.id = spi.plan_id AND sp.is_current = true
		      JOIN assets a ON a.site_id = p.site_id AND a.asset_key = spi.key AND a.status = 'active'
		     WHERE sp.site_id = p.site_id AND spi.kind = 'hero'
		       AND spi.scope = 'page' AND spi.scope_ref = p.name
		     LIMIT 1
		) ha ON true`

// pageImageCols carries the scanned image candidates for one listing row.
type pageImageCols struct {
	CardKey     string
	HeroKey     string
	HeroPurpose string
}

// webPath resolves the item's image to a deployed git path: card first, plan
// hero second, empty when the page has neither. Never assets.url — that holds
// an expiring presigned S3 URL.
func (c pageImageCols) webPath() string {
	if c.CardKey != "" {
		return storage.DeployedWebPath(c.CardKey, "card")
	}
	if c.HeroKey != "" {
		return storage.DeployedWebPath(c.HeroKey, c.HeroPurpose)
	}
	return ""
}

// resolvePagesWhereType returns all pages on the site with the given
// page_type, projected to the standard list-item shape.
//
// Standard list-item shape (consumed by tool-list, blog-listing, etc.):
//
//	{
//	  "name":             "tool-jump-physics",
//	  "title":            "Jump Physics",
//	  "url":              "/tools/jump-physics/index.html",
//	  "meta_description": "...",
//	  "nav_label":        "Jump Physics Architect",
//	  "image":            "/assets/images/card-tool-jump-physics.jpg"
//	}
//
// image (Phase I3, Lane B) is the page's entity-linked CARD asset when one
// exists, else the page's own plan hero (heavier, but present — the card
// supersedes it as derivations land), else "" (components treat a missing
// image as no-thumbnail).
//
// The shape is intentionally generic — components can pick any subset of
// these keys via their `items.items.X.source: "field.X"` declarations
// (handled at template-render time, not here).
//
// Filter: status IN ('active', 'deployed') so unbuilt pages don't appear.
//
// Note on in_header: we deliberately do NOT filter on in_header here.
// Tool/blog/entity detail pages typically have in_header=false (they are
// individual items, not nav destinations) but they SHOULD appear in their
// parent index's list. The page_type filter is the correct gate; in_header
// is for navigation, not for listings.
func resolvePagesWhereType(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	pageType string,
	limit int,
	logger *zap.Logger,
) (interface{}, error) {
	if pageType == "" {
		return nil, fmt.Errorf("resolvePagesWhereType: empty page_type argument")
	}

	// Hard cap. Components rarely want more than 12 items in a list; cap at
	// 24 to stop runaway queries. If the schema asks for more, we honour up
	// to 24.
	const hardCap = 24
	if limit <= 0 {
		limit = 12
	}
	if limit > hardCap {
		limit = hardCap
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
		    p.name,
		    COALESCE(p.title, p.name)               AS title,
		    p.url,
		    COALESCE(p.meta_description, '')        AS meta_description,
		    COALESCE(p.nav_label, p.title, p.name)  AS nav_label,
		    `+pageImageProjection+`
		FROM pages p
		`+pageImageJoins+`
		WHERE p.site_id   = $1
		  AND p.page_type = $2
		  AND p.status   IN ('active', 'deployed')
		ORDER BY COALESCE(p.nav_order, 100), p.name
		LIMIT $3
	`, siteID, pageType, limit)
	if err != nil {
		return nil, fmt.Errorf("resolvePagesWhereType query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, title, url, metaDesc, navLabel string
		var img pageImageCols
		if err := rows.Scan(&name, &title, &url, &metaDesc, &navLabel, &img.CardKey, &img.HeroKey, &img.HeroPurpose); err != nil {
			logger.Warn("resolvePagesWhereType: scan failed", zap.Error(err))
			continue
		}
		items = append(items, map[string]interface{}{
			"name":             name,
			"title":            title,
			"url":              url,
			"meta_description": metaDesc,
			"nav_label":        navLabel,
			"image":            img.webPath(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolvePagesWhereType rows iter failed: %w", err)
	}

	logger.Info("queryresolve: resolved pages_where_type",
		zap.String("page_type", pageType),
		zap.String("site_id", siteID.String()),
		zap.Int("count", len(items)),
		zap.Int("limit", limit),
	)

	return items, nil
}

// resolvePagesUnderSection returns all pages that belong to the named site
// area (section), projected to the same standard list-item shape as
// resolvePagesWhereType. The <area> argument matches site_areas.name (the
// per-site unique key) via pages.site_area_id; url_prefix is a forgiving
// fallback so `query.pages_under_section:guides` resolves whether the area was
// keyed "guides" or "/guides".
//
// Filter: status IN ('active', 'deployed') so unbuilt pages don't appear.
// in_header is deliberately NOT filtered — item pages under a section often
// have in_header=false but still belong in the section's listing (same
// reasoning as resolvePagesWhereType).
func resolvePagesUnderSection(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	area string,
	limit int,
	logger *zap.Logger,
) (interface{}, error) {
	if area == "" {
		return nil, fmt.Errorf("resolvePagesUnderSection: empty section argument")
	}

	const hardCap = 24
	if limit <= 0 {
		limit = 12
	}
	if limit > hardCap {
		limit = hardCap
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
		    p.name,
		    COALESCE(p.title, p.name)              AS title,
		    p.url,
		    COALESCE(p.meta_description, '')       AS meta_description,
		    COALESCE(p.nav_label, p.title, p.name) AS nav_label,
		    `+pageImageProjection+`
		FROM pages p
		JOIN site_areas sa
		  ON sa.id = p.site_area_id
		 AND sa.site_id = p.site_id
		`+pageImageJoins+`
		WHERE p.site_id = $1
		  AND (lower(sa.name) = lower($2)
		       OR sa.url_prefix = $2
		       OR sa.url_prefix = '/' || $2)
		  AND p.status IN ('active', 'deployed')
		ORDER BY COALESCE(p.nav_order, 100), p.name
		LIMIT $3
	`, siteID, area, limit)
	if err != nil {
		return nil, fmt.Errorf("resolvePagesUnderSection query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, title, url, metaDesc, navLabel string
		var img pageImageCols
		if err := rows.Scan(&name, &title, &url, &metaDesc, &navLabel, &img.CardKey, &img.HeroKey, &img.HeroPurpose); err != nil {
			logger.Warn("resolvePagesUnderSection: scan failed", zap.Error(err))
			continue
		}
		items = append(items, map[string]interface{}{
			"name":             name,
			"title":            title,
			"url":              url,
			"meta_description": metaDesc,
			"nav_label":        navLabel,
			"image":            img.webPath(),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolvePagesUnderSection rows iter failed: %w", err)
	}

	logger.Info("queryresolve: resolved pages_under_section",
		zap.String("section", area),
		zap.String("site_id", siteID.String()),
		zap.Int("count", len(items)),
		zap.Int("limit", limit),
	)

	return items, nil
}

// resolveProducts returns active products for the site, optionally filtered
// by category (`query.products:gripper` → category "gripper"). Only
// `status = 'active'` rows are returned — a row can exist (e.g. pending
// verification) without being render-eligible yet.
//
// The `specifications` jsonb column is flattened into the returned map
// alongside the fixed fields, so a component's html_template can reference
// any spec key directly (e.g. {{.stroke}}, {{.gripping_force}}) without a
// nested lookup — the schema's `items` block documents which keys a given
// category's rows are expected to carry, but this resolver doesn't enforce
// that shape; it returns whatever specifications each row has.
//
// Provenance (source_url, verified_date) lives in content_data rather than
// dedicated columns — no migration needed, and it travels with the row
// exactly like every other per-product fact a template might show.
func resolveProducts(
	ctx context.Context,
	db *sql.DB,
	siteID uuid.UUID,
	category string,
	limit int,
	logger *zap.Logger,
) (interface{}, error) {
	const hardCap = 24
	if limit <= 0 {
		limit = 12
	}
	if limit > hardCap {
		limit = hardCap
	}

	query := `
		SELECT
		    id::text, name, COALESCE(sku, ''), COALESCE(subcategory, ''),
		    COALESCE(specifications, '{}'::jsonb),
		    COALESCE(price_display, ''),
		    COALESCE(content_data->>'source_url', ''),
		    COALESCE(content_data->>'verified_date', '')
		FROM products
		WHERE site_id = $1
		  AND status = 'active'
	`
	args := []interface{}{siteID}
	if category != "" {
		query += ` AND category = $2 ORDER BY name LIMIT $3`
		args = append(args, category, limit)
	} else {
		query += ` ORDER BY name LIMIT $2`
		args = append(args, limit)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("resolveProducts query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name, sku, subcategory, priceDisplay, sourceURL, verifiedDate string
		var specsJSON []byte
		if err := rows.Scan(&id, &name, &sku, &subcategory, &specsJSON, &priceDisplay, &sourceURL, &verifiedDate); err != nil {
			logger.Warn("resolveProducts: scan failed", zap.Error(err))
			continue
		}

		item := map[string]interface{}{
			"id":            id,
			"name":          name,
			"sku":           sku,
			"subcategory":   subcategory,
			"price_display": priceDisplay,
			"source_url":    sourceURL,
			"verified_date": verifiedDate,
		}

		var specs map[string]interface{}
		if len(specsJSON) > 0 {
			if err := json.Unmarshal(specsJSON, &specs); err == nil {
				for k, v := range specs {
					item[k] = v
				}
			}
		}

		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("resolveProducts rows iter failed: %w", err)
	}

	logger.Info("queryresolve: resolved products",
		zap.String("category", category),
		zap.String("site_id", siteID.String()),
		zap.Int("count", len(items)),
		zap.Int("limit", limit),
	)

	return items, nil
}
