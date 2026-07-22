// FILE: platform/orchestration/actions/queryresolve/model_directory_items.go
//
// Model directory resolution for `query.model_directory` /
// `query.model_directory_full` (model_directory_pipeline Phase C).
//
// Unlike every other resolver in this package, this one is deliberately NOT
// site-scoped: directory_entities/directory_claims (migration 192) is one
// global registry, read identically by every opted-in site — a model's price
// is not a fact about any one site. Resolve()'s SiteID parameter is still
// required (package invariant, "there are no cross-site queries" — this is
// the one query whose answer just happens not to depend on which site asked).
//
// ONE QUERY, TWO PROJECTIONS, same discipline as news_items.go:
// QueryModelDirectoryEntries is shared between this resolver (server-rendered
// HTML) and render_model_directory_action.go's JSON path, so the two can
// never disagree about which entities/claims exist.
//
// Only status='found' claims are surfaced. A claim currently citation_lost
// or fetch_error is not fit to publish as a cited fact — directory_claims.go's
// freshness sweep can flip it back to found later, at which point it
// reappears automatically; there is no separate "republish" step.

package queryresolve

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ModelDirectoryClaim is one cited fact about an entity, as selected.
type ModelDirectoryClaim struct {
	Field     string
	Value     string
	Unit      string
	URL       string
	Publisher string
}

// ModelDirectoryEntry is one directory_entities row with its current, found
// claims.
type ModelDirectoryEntry struct {
	ID      uuid.UUID
	Slug    string
	Name    string
	Owner   string
	Summary string
	Links   map[string]interface{}
	Claims  []ModelDirectoryClaim
}

// QueryModelDirectoryEntries selects up to `limit` active 'model' entities
// with their current, found claims. kind is fixed to 'model' for Phase C —
// Phase E adds 'company'/'protocol' with their own resolver entry points,
// same tables, no schema change.
//
// The LIMIT is applied inside the entity subquery, before the claims join,
// so it bounds the number of MODELS returned, not the number of
// entity-claim rows (a naive LIMIT after the join would truncate mid-model).
func QueryModelDirectoryEntries(ctx context.Context, db *sql.DB, limit int, logger *zap.Logger) ([]ModelDirectoryEntry, error) {
	if limit <= 0 {
		limit = 24
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := db.QueryContext(ctx, `
		SELECT de.id, de.slug, de.name, COALESCE(de.owner, ''), COALESCE(de.summary, ''), de.links,
		       dc.field, dc.value, dc.unit, dc.citation->>'url', dc.citation->>'publisher'
		FROM (
			SELECT id, slug, name, owner, summary, links, updated_at
			FROM directory_entities
			WHERE kind = 'model' AND status = 'active'
			ORDER BY updated_at DESC
			LIMIT $1
		) de
		LEFT JOIN directory_claims dc
			ON dc.entity_id = de.id AND dc.is_current AND dc.status = 'found'
		ORDER BY de.updated_at DESC, dc.field
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("QueryModelDirectoryEntries: %w", err)
	}
	defer rows.Close()

	order := make([]uuid.UUID, 0)
	byID := make(map[uuid.UUID]*ModelDirectoryEntry)
	for rows.Next() {
		var id uuid.UUID
		var slug, name, owner, summary, linksJSON string
		var field, value, unit, url, publisher sql.NullString
		if err := rows.Scan(&id, &slug, &name, &owner, &summary, &linksJSON,
			&field, &value, &unit, &url, &publisher); err != nil {
			logger.Warn("QueryModelDirectoryEntries: scan failed", zap.Error(err))
			continue
		}
		e, ok := byID[id]
		if !ok {
			e = &ModelDirectoryEntry{ID: id, Slug: slug, Name: name, Owner: owner, Summary: summary}
			_ = json.Unmarshal([]byte(linksJSON), &e.Links)
			byID[id] = e
			order = append(order, id)
		}
		if field.Valid && field.String != "" {
			e.Claims = append(e.Claims, ModelDirectoryClaim{
				Field: field.String, Value: value.String, Unit: unit.String,
				URL: url.String, Publisher: publisher.String,
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("QueryModelDirectoryEntries rows: %w", err)
	}

	entries := make([]ModelDirectoryEntry, 0, len(order))
	for _, id := range order {
		entries = append(entries, *byID[id])
	}
	return entries, nil
}

// resolveModelDirectory backs `source: "query.model_directory"` — the
// homepage snippet card grid. Default 12 entries, cap 24 (QueryModelDirectoryEntries's cap).
func resolveModelDirectory(ctx context.Context, db *sql.DB, limit int, logger *zap.Logger) (interface{}, error) {
	if limit <= 0 {
		limit = 12
	}
	entries, err := QueryModelDirectoryEntries(ctx, db, limit, logger)
	if err != nil {
		return nil, err
	}
	logger.Info("queryresolve: resolved model_directory", zap.Int("entries", len(entries)))
	return projectModelDirectoryEntries(entries), nil
}

// resolveModelDirectoryFull backs `source: "query.model_directory_full"` —
// the dedicated model-directory listing page. Default 50 entries.
func resolveModelDirectoryFull(ctx context.Context, db *sql.DB, limit int, logger *zap.Logger) (interface{}, error) {
	if limit <= 0 {
		limit = 50
	}
	entries, err := QueryModelDirectoryEntries(ctx, db, limit, logger)
	if err != nil {
		return nil, err
	}
	logger.Info("queryresolve: resolved model_directory_full", zap.Int("entries", len(entries)))
	return projectModelDirectoryEntries(entries), nil
}

// projectModelDirectoryEntries shapes raw entries for template rendering:
// HTML-escaped text, only the link kinds a template actually needs pulled to
// the top level. Component templates render through text/template
// (component_library.go RenderTemplate), which does NOT auto-escape, and
// researched model names/summaries are third-party content — same escaping
// discipline as news_items.go's projectNewsItems.
func projectModelDirectoryEntries(entries []ModelDirectoryEntry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		claims := make([]map[string]interface{}, 0, len(e.Claims))
		for _, c := range e.Claims {
			claims = append(claims, map[string]interface{}{
				"field":     html.EscapeString(c.Field),
				"value":     html.EscapeString(c.Value),
				"unit":      html.EscapeString(c.Unit),
				"url":       html.EscapeString(c.URL),
				"publisher": html.EscapeString(c.Publisher),
			})
		}
		m := map[string]interface{}{
			"slug":    html.EscapeString(e.Slug),
			"name":    html.EscapeString(e.Name),
			"owner":   html.EscapeString(e.Owner),
			"summary": html.EscapeString(e.Summary),
			"claims":  claims,
		}
		if docs, ok := e.Links["docs"].(string); ok && docs != "" {
			m["docs_url"] = html.EscapeString(docs)
		}
		if weights, ok := e.Links["weights"].(string); ok && weights != "" {
			m["weights_url"] = html.EscapeString(weights)
		}
		if wrapper, ok := e.Links["wrapper_url"].(string); ok && wrapper != "" {
			m["wrapper_url"] = html.EscapeString(wrapper)
		}
		if videosRaw, ok := e.Links["video_urls"].([]interface{}); ok {
			videos := make([]string, 0, len(videosRaw))
			for _, v := range videosRaw {
				if s, ok2 := v.(string); ok2 && s != "" {
					videos = append(videos, html.EscapeString(s))
				}
			}
			if len(videos) > 0 {
				m["video_urls"] = videos
			}
		}
		out = append(out, m)
	}
	return out
}
