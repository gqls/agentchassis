// FILE: platform/orchestration/actions/queryresolve/directory_items.go
//
// Directory resolution for `query.model_directory` /
// `query.model_directory_full` (model_directory_pipeline Phase C) and
// `query.adoption_tracker` / `query.adoption_tracker_full` (Phase E).
//
// ONE registry, MANY kinds. directory_entities.kind partitions the register:
// 'model' holds AI models, 'company' holds organisations adopting agents,
// 'protocol' holds agentic-communication protocols (MCP and friends). All
// three are read by the same query with a different kind — the Phase E
// adoption tracker needed no schema change and no second query path, which
// was the point of designing the register kind-first in Phase A.
//
// Renamed from model_directory_items.go 2026-07-25 when the second kind
// arrived: a file called model_* that also serves companies is the kind of
// small lie that costs the next reader ten minutes.
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
	"strings"

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
// with their current, found claims. Retained as a named entry point because
// the model directory is the registry's oldest reader and its callers read
// better for it; QueryDirectoryEntries is the general form.
func QueryModelDirectoryEntries(ctx context.Context, db *sql.DB, limit int, logger *zap.Logger) ([]ModelDirectoryEntry, error) {
	return QueryDirectoryEntries(ctx, db, "model", limit, logger)
}

// QueryDirectoryEntries selects up to `limit` active entities of one kind
// with their current, found claims.
//
// An empty kind is REFUSED rather than treated as "every kind". The register
// mixes AI models with companies and protocols; a caller that forgot to say
// which one it wanted would silently render a page listing all three, and
// that failure would look like bad content rather than a bug. The one place
// a kind-wide sweep is legitimate — the freshness re-verification in
// directory_claims.go — asks for it explicitly.
//
// The LIMIT is applied inside the entity subquery, before the claims join,
// so it bounds the number of ENTITIES returned, not the number of
// entity-claim rows (a naive LIMIT after the join would truncate mid-entity).
func QueryDirectoryEntries(ctx context.Context, db *sql.DB, kind string, limit int, logger *zap.Logger) ([]ModelDirectoryEntry, error) {
	if strings.TrimSpace(kind) == "" {
		return nil, fmt.Errorf("QueryDirectoryEntries: kind is required")
	}
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
			WHERE kind = $1 AND status = 'active'
			ORDER BY updated_at DESC
			LIMIT $2
		) de
		LEFT JOIN directory_claims dc
			ON dc.entity_id = de.id AND dc.is_current AND dc.status = 'found'
		ORDER BY de.updated_at DESC, dc.field
	`, strings.TrimSpace(kind), limit)
	if err != nil {
		return nil, fmt.Errorf("QueryDirectoryEntries(%s): %w", kind, err)
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
			logger.Warn("QueryDirectoryEntries: scan failed", zap.String("kind", kind), zap.Error(err))
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
		return nil, fmt.Errorf("QueryDirectoryEntries(%s) rows: %w", kind, err)
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

// resolveDirectoryKind backs the Phase E kinds — `query.adoption_tracker`
// (companies adopting agents) and `query.protocol_tracker` (agentic
// communication protocols), plus their `_full` listing variants. Same
// register, same projection, same escaping; only `kind` and the default
// depth differ, which is exactly as much difference as there should be.
func resolveDirectoryKind(ctx context.Context, db *sql.DB, kind string, limit, fallback int, logger *zap.Logger) (interface{}, error) {
	if limit <= 0 {
		limit = fallback
	}
	entries, err := QueryDirectoryEntries(ctx, db, kind, limit, logger)
	if err != nil {
		return nil, err
	}
	logger.Info("queryresolve: resolved directory kind",
		zap.String("kind", kind), zap.Int("entries", len(entries)))
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
