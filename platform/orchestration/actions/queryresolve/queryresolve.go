// FILE: platform/orchestration/actions/queryresolve/queryresolve.go
//
// Package queryresolve runs named queries against the database for `query.*`
// source resolution in component input_schemas. Components that show lists
// (tool-list, blog-listing, directory-listing, etc.) declare their items
// field with `source: "query.<name>"`, and this package executes that
// resolution.
//
// Initial scope (v1, focus doc FOCUS_directory_builder_and_list_components.md):
//   - One concrete query: pages_where_type:tool
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
	"fmt"
	"strings"

	"github.com/google/uuid"
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
//	  "nav_label":        "Jump Physics Architect"
//	}
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
		    name,
		    COALESCE(title, name)               AS title,
		    url,
		    COALESCE(meta_description, '')      AS meta_description,
		    COALESCE(nav_label, title, name)    AS nav_label
		FROM pages
		WHERE site_id   = $1
		  AND page_type = $2
		  AND status   IN ('active', 'deployed')
		ORDER BY COALESCE(nav_order, 100), name
		LIMIT $3
	`, siteID, pageType, limit)
	if err != nil {
		return nil, fmt.Errorf("resolvePagesWhereType query failed: %w", err)
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var name, title, url, metaDesc, navLabel string
		if err := rows.Scan(&name, &title, &url, &metaDesc, &navLabel); err != nil {
			logger.Warn("resolvePagesWhereType: scan failed", zap.Error(err))
			continue
		}
		items = append(items, map[string]interface{}{
			"name":             name,
			"title":            title,
			"url":              url,
			"meta_description": metaDesc,
			"nav_label":        navLabel,
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
