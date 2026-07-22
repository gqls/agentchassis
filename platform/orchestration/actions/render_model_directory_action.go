// FILE: platform/orchestration/actions/render_model_directory_action.go
//
// RenderModelDirectoryAction produces JSON files ready for git commit from
// the global directory_entities/directory_claims registry (model_directory_pipeline
// Phase C) — the publish leg, a cousin of render_news_section_action.go.
//
// Unlike the news feed, the registry is NOT per-site: every opted-in site
// renders the SAME entities/claims. This action still takes a site_id
// (needed to resolve the domain for the git commit, and to scope which
// pages get a rerender queued) but the query itself is shared, unscoped —
// queryresolve.QueryModelDirectoryEntries, the same function the
// server-rendered query.model_directory/query.model_directory_full
// resolvers use, so the JSON file and the baked HTML can never disagree
// about which entities exist.
//
// Output: {files: {"data/model-directory.json": "<json>"[, "data/model-directory-full.json": "<json>"]},
//          domain: "...", entity_count: N}
// The git_commit step in the workflow reads files_field and commits to the repo.
//
// The full listing file is only produced if the site has a deployed page
// carrying the model-directory-listing component — same conditional-archive
// pattern render_news_section_action.go uses for news-archive.json.
//
// Registration:
//   "render_model_directory": {
//       Handler:     RenderModelDirectoryAction,
//       Category:    "site",
//       Description: "Produce model-directory JSON for git commit from the global directory registry",
//       IsLocal:     true,
//   }
//
// Workflow config:
//   "render_model_directory_json": {
//       "action": "render_model_directory",
//       "config": {"site_id": "input_data.site_id"},
//       "output_field": "directory_render_result",
//       "next_step": "commit_model_directory"
//   },
//   "commit_model_directory": {
//       "action": "git_commit",
//       "config": {
//           "files_field": "directory_render_result.files",
//           "domain_field": "directory_render_result.domain",
//           "commit_message": "Update model directory"
//       },
//       "output_field": "directory_commit_result",
//       "next_step": "complete"
//   }

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenderModelDirectoryInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"max_items", "full_max_items"},
}

func init() {
	datahelpers.RegisterActionInputSpec("render_model_directory", RenderModelDirectoryInputSpec)
}

// modelDirectoryJSONOutput is the shape of /data/model-directory.json and
// /data/model-directory-full.json.
type modelDirectoryJSONOutput struct {
	Headline       string                   `json:"headline"`
	Entries        []map[string]interface{} `json:"entries"`
	DirectoryURL   string                   `json:"directory_url,omitempty"`
	DirectoryLabel string                   `json:"directory_label,omitempty"`
	UpdatedAt      string                   `json:"updated_at"`
	Count          int                      `json:"count"`
}

func RenderModelDirectoryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "render_model_directory"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config, RenderModelDirectoryInputSpec, logger)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}
	maxItems := inputs.GetInt("max_items", 12)
	fullMaxItems := inputs.GetInt("full_max_items", 50)

	var domain string
	if err := params.DB.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("query site domain: %w", err)
	}

	snippetEntries, err := queryresolve.QueryModelDirectoryEntries(ctx, params.DB, maxItems, logger)
	if err != nil {
		return nil, fmt.Errorf("query model directory: %w", err)
	}

	var listingURL sql.NullString
	_ = params.DB.QueryRowContext(ctx, `
		SELECT p.url FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1 AND cc.function = 'model-directory-listing' AND p.status = 'active'
		LIMIT 1
	`, siteID).Scan(&listingURL)
	hasListingPage := listingURL.Valid && listingURL.String != ""

	snippetOutput := modelDirectoryJSONOutput{
		Headline:  "AI Model Directory",
		Entries:   projectModelDirectoryJSON(snippetEntries),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Count:     len(snippetEntries),
	}
	if hasListingPage {
		snippetOutput.DirectoryURL = listingURL.String
		snippetOutput.DirectoryLabel = "Browse the full directory →"
	}
	snippetJSON, err := json.MarshalIndent(snippetOutput, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal snippet JSON: %w", err)
	}

	filesMap := map[string]interface{}{
		"data/model-directory.json": string(snippetJSON),
	}
	totalCount := len(snippetEntries)

	if hasListingPage {
		fullEntries, err := queryresolve.QueryModelDirectoryEntries(ctx, params.DB, fullMaxItems, logger)
		if err != nil {
			logger.Warn("RenderModelDirectoryAction: full listing query failed, skipping", zap.Error(err))
		} else {
			fullOutput := modelDirectoryJSONOutput{
				Headline:  "AI Model Directory",
				Entries:   projectModelDirectoryJSON(fullEntries),
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
				Count:     len(fullEntries),
			}
			fullJSON, err := json.MarshalIndent(fullOutput, "", "  ")
			if err != nil {
				logger.Warn("RenderModelDirectoryAction: marshal full JSON failed", zap.Error(err))
			} else {
				filesMap["data/model-directory-full.json"] = string(fullJSON)
				totalCount += len(fullEntries)
			}
		}
	}

	rerenderQueued := 0
	if totalCount > 0 {
		rerenderQueued = queueModelDirectoryPageRerenders(ctx, params.DB, siteID, logger)
	}

	logger.Info("RenderModelDirectoryAction: JSON produced",
		zap.Int("snippet_entries", len(snippetEntries)),
		zap.Int("files_count", len(filesMap)),
		zap.String("domain", domain),
		zap.Bool("has_listing", hasListingPage))

	return map[string]interface{}{
		"files":           filesMap,
		"domain":          domain,
		"entity_count":    totalCount,
		"has_listing":     hasListingPage,
		"rendered":        true,
		"rerender_queued": rerenderQueued,
	}, nil
}

// projectModelDirectoryJSON re-shapes the resolver's escaped
// []map[string]interface{} for the JSON file: the client fetch renders this
// with its own script, so it wants raw text, not HTML entities — same
// unescape-for-JSON step render_news_section_action.go's loadNewsItems
// performs implicitly by projecting from the raw NewsItem struct rather than
// the escaped template map. Here the resolver already escaped once (for the
// template projection), so this re-derives the JSON shape from the same
// entries structurally rather than double-escaping.
func projectModelDirectoryJSON(entries []queryresolve.ModelDirectoryEntry) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		claims := make([]map[string]interface{}, 0, len(e.Claims))
		for _, c := range e.Claims {
			claims = append(claims, map[string]interface{}{
				"field": c.Field, "value": c.Value, "unit": c.Unit,
				"url": c.URL, "publisher": c.Publisher,
			})
		}
		m := map[string]interface{}{
			"slug": e.Slug, "name": e.Name, "owner": e.Owner, "summary": e.Summary,
			"claims": claims,
		}
		if docs, ok := e.Links["docs"].(string); ok && docs != "" {
			m["docs_url"] = docs
		}
		if weights, ok := e.Links["weights"].(string); ok && weights != "" {
			m["weights_url"] = weights
		}
		if wrapper, ok := e.Links["wrapper_url"].(string); ok && wrapper != "" {
			m["wrapper_url"] = wrapper
		}
		if videosRaw, ok := e.Links["video_urls"].([]interface{}); ok {
			videos := make([]string, 0, len(videosRaw))
			for _, v := range videosRaw {
				if s, ok2 := v.(string); ok2 && s != "" {
					videos = append(videos, s)
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

// queueModelDirectoryPageRerenders emits one scoped re-render work item per
// page carrying a model-directory component, so freshly published/refreshed
// claims reach the deployed HTML — cousin of queueNewsPageRerenders
// (render_news_section_html.go). recurrenceExpected: true for the same
// reason — a re-request every publish cycle is the design, not a failed fix.
func queueModelDirectoryPageRerenders(ctx context.Context, db *sql.DB, siteID uuid.UUID, logger *zap.Logger) int {
	if db == nil {
		return 0
	}

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.name
		FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND cc.function IN ('model-directory', 'model-directory-listing')
		  AND p.build_status = 'deployed'
	`, siteID)
	if err != nil {
		logger.Warn("queueModelDirectoryPageRerenders: page lookup failed", zap.Error(err))
		return 0
	}
	defer rows.Close()

	var pages []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			logger.Warn("queueModelDirectoryPageRerenders: scan failed", zap.Error(err))
			continue
		}
		pages = append(pages, name)
	}
	if len(pages) == 0 {
		return 0
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("queueModelDirectoryPageRerenders: begin tx failed", zap.Error(err))
		return 0
	}
	defer tx.Rollback()

	batchID := uuid.New()
	queued := 0
	for _, page := range pages {
		item := workItem{
			siteID:       siteID,
			source:       "render_model_directory",
			pipeline:     "build",
			itemType:     "needs_page",
			severity:     "low",
			summary:      fmt.Sprintf("Re-render %s — model directory data refreshed", page),
			spec:         fmt.Sprintf(`{"reason":"section_data_resolved","page_name":%q}`, page),
			priority:     99,
			handlerAgent: "page-build-handler",
			status:       "triaged",
			createdBy:    "render_model_directory",
			itemKey:      fmt.Sprintf("page_rerender:%s", page),
			batchID:      batchID,

			recurrenceExpected: true,
		}
		inserted, err := insertWorkItem(ctx, tx, item, logger)
		if err != nil {
			logger.Warn("queueModelDirectoryPageRerenders: insert failed",
				zap.String("page", page), zap.Error(err))
			continue
		}
		if inserted {
			queued++
		}
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("queueModelDirectoryPageRerenders: commit failed", zap.Error(err))
		return 0
	}

	if queued > 0 {
		logger.Info("queueModelDirectoryPageRerenders: queued scoped re-renders",
			zap.Int("queued", queued), zap.Int("pages", len(pages)))
	}
	return queued
}
