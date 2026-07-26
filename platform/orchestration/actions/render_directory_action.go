// FILE: platform/orchestration/actions/render_directory_action.go
//
// RenderDirectoryAction produces JSON files ready for git commit from
// the global directory_entities/directory_claims registry (model_directory_pipeline
// Phase C for kind='model', Phase E for 'company'/'protocol') — the publish
// leg, a cousin of render_news_section_action.go.
//
// ONE action, one profile per kind (directoryPublishProfiles below). The
// only things that vary between an AI-model directory and a company
// adoption tracker are which kind to query, what to call the file, and
// which components mark a page as carrying it. Everything else — the
// conditional full listing, the git handoff, the scoped rerender queue — is
// identical, so it is written once.
//
// Registered under BOTH "render_directory" (general) and
// "render_model_directory" (the original name, kept because the live
// model-directory-publisher workflow seed references it; renaming a
// registered action would break a scheduled task in production for no gain).
//
// Unlike the news feed, the registry is NOT per-site: every opted-in site
// renders the SAME entities/claims. This action still takes a site_id
// (needed to resolve the domain for the git commit, and to scope which
// pages get a rerender queued) but the query itself is shared, unscoped —
// queryresolve.QueryDirectoryEntries, the same function the server-rendered
// query.model_directory / query.adoption_tracker resolvers use, so the JSON
// file and the baked HTML can never disagree about which entities exist.
//
// Output: {files: {"<profile.SnippetFile>": "<json>"[, "<profile.FullFile>": "<json>"]},
//          domain: "...", entity_count: N}
// The git_commit step in the workflow reads files_field and commits to the repo.
//
// The full listing file is only produced if the site has a deployed page
// carrying this kind's listing component — same conditional-archive pattern
// render_news_section_action.go uses for news-archive.json.
//
// Workflow config (kind omitted ⇒ 'model', so pre-Phase-E seeds are unchanged):
//   "render_directory_json": {
//       "action": "render_directory",
//       "config": {"site_id": "input_data.site_id", "kind": "company"},
//       "output_field": "directory_render_result",
//       "next_step": "commit_directory"
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
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var RenderModelDirectoryInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"max_items", "full_max_items", "kind"},
}

func init() {
	datahelpers.RegisterActionInputSpec("render_model_directory", RenderModelDirectoryInputSpec)
	datahelpers.RegisterActionInputSpec("render_directory", RenderModelDirectoryInputSpec)
}

// directoryPublishProfile is everything that differs between one register
// kind and another at publish time.
type directoryPublishProfile struct {
	Kind             string
	Headline         string
	SnippetFile      string
	FullFile         string
	SnippetComponent string
	ListingComponent string
	BrowseLabel      string
}

// directoryPublishProfiles is the closed set of publishable kinds. A kind
// absent from this map is REFUSED rather than defaulted: publishing under a
// guessed filename would commit a file nothing reads and report success —
// the silent-success shape this pipeline has already been bitten by twice
// (bugs_closed/062, and the never-seeded publish leg of Phase D).
var directoryPublishProfiles = map[string]directoryPublishProfile{
	"model": {
		Kind:             "model",
		Headline:         "AI Model Directory",
		SnippetFile:      "data/model-directory.json",
		FullFile:         "data/model-directory-full.json",
		SnippetComponent: "model-directory",
		ListingComponent: "model-directory-listing",
		BrowseLabel:      "Browse the full directory →",
	},
	"company": {
		Kind:             "company",
		Headline:         "Who is actually deploying AI agents",
		SnippetFile:      "data/adoption-tracker.json",
		FullFile:         "data/adoption-tracker-full.json",
		SnippetComponent: "adoption-tracker",
		ListingComponent: "adoption-tracker-listing",
		BrowseLabel:      "See every tracked deployment →",
	},
	"protocol": {
		Kind:             "protocol",
		Headline:         "Agent communication protocols",
		SnippetFile:      "data/protocol-tracker.json",
		FullFile:         "data/protocol-tracker-full.json",
		SnippetComponent: "protocol-tracker",
		ListingComponent: "protocol-tracker-listing",
		BrowseLabel:      "See every tracked protocol →",
	},
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

func RenderDirectoryAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "render_directory"))

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

	// Which register kind to publish.
	//
	// THE TRAP THIS CODE EXISTS TO AVOID (paid for in production 2026-07-26):
	// ExtractActionInputs treats every STRING config value as a REFERENCE to
	// resolve against collected_data, never as a literal — deliberately, so a
	// broken reference cannot masquerade as a working literal
	// (action_inputs.go Strategy 5, and the bugs_open/042 case it cites). So a
	// step configured `"kind": "company"` resolves nothing, the field is
	// dropped, and the Go default silently wins. The first live run of the
	// three-kind publish chain therefore published the MODEL register three
	// times, under commit messages that said "Update adoption tracker" and
	// "Update protocol tracker".
	//
	// The fix is not to make the generic resolver take string literals — that
	// would reintroduce exactly the masking it was written to prevent. It is
	// for this action to read its OWN step config for a value from a CLOSED
	// set. A profile name cannot be confused with a reference: references are
	// dotted paths or collected_data keys, and no profile name is either. An
	// unrecognised value falls through to the reference path and then to the
	// unknown-kind refusal below, so a typo still fails loudly.
	kind := strings.TrimSpace(inputs.Get("kind"))
	if raw, ok := params.StepConfig.Config["kind"].(string); ok {
		if _, known := directoryPublishProfiles[strings.TrimSpace(raw)]; known {
			kind = strings.TrimSpace(raw)
		}
	}
	// Absent kind means 'model' — the workflows seeded before Phase E existed
	// pass no kind at all, and they must keep publishing exactly what they
	// published yesterday.
	if kind == "" {
		kind = "model"
	}
	profile, ok := directoryPublishProfiles[kind]
	if !ok {
		return nil, fmt.Errorf("render_directory: unknown register kind %q (known: model, company, protocol)", kind)
	}
	logger = logger.With(zap.String("kind", kind))

	var domain string
	if err := params.DB.QueryRowContext(ctx, `SELECT domain FROM sites WHERE id = $1`, siteID).Scan(&domain); err != nil {
		return nil, fmt.Errorf("query site domain: %w", err)
	}

	snippetEntries, err := queryresolve.QueryDirectoryEntries(ctx, params.DB, profile.Kind, maxItems, logger)
	if err != nil {
		return nil, fmt.Errorf("query %s directory: %w", profile.Kind, err)
	}

	var listingURL sql.NullString
	_ = params.DB.QueryRowContext(ctx, `
		SELECT p.url FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1 AND cc.function = $2 AND p.status = 'active'
		LIMIT 1
	`, siteID, profile.ListingComponent).Scan(&listingURL)
	hasListingPage := listingURL.Valid && listingURL.String != ""

	snippetOutput := modelDirectoryJSONOutput{
		Headline:  profile.Headline,
		Entries:   projectModelDirectoryJSON(snippetEntries),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		Count:     len(snippetEntries),
	}
	if hasListingPage {
		snippetOutput.DirectoryURL = listingURL.String
		snippetOutput.DirectoryLabel = profile.BrowseLabel
	}
	snippetJSON, err := json.MarshalIndent(snippetOutput, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal snippet JSON: %w", err)
	}

	filesMap := map[string]interface{}{
		profile.SnippetFile: string(snippetJSON),
	}
	totalCount := len(snippetEntries)

	if hasListingPage {
		fullEntries, err := queryresolve.QueryDirectoryEntries(ctx, params.DB, profile.Kind, fullMaxItems, logger)
		if err != nil {
			logger.Warn("RenderDirectoryAction: full listing query failed, skipping", zap.Error(err))
		} else {
			fullOutput := modelDirectoryJSONOutput{
				Headline:  profile.Headline,
				Entries:   projectModelDirectoryJSON(fullEntries),
				UpdatedAt: time.Now().UTC().Format(time.RFC3339),
				Count:     len(fullEntries),
			}
			fullJSON, err := json.MarshalIndent(fullOutput, "", "  ")
			if err != nil {
				logger.Warn("RenderDirectoryAction: marshal full JSON failed", zap.Error(err))
			} else {
				filesMap[profile.FullFile] = string(fullJSON)
				totalCount += len(fullEntries)
			}
		}
	}

	rerenderQueued := 0
	if totalCount > 0 {
		rerenderQueued = queueDirectoryPageRerenders(ctx, params.DB, siteID, profile, logger)
	}

	logger.Info("RenderDirectoryAction: JSON produced",
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

// queueDirectoryPageRerenders emits one scoped re-render work item per page
// carrying either of this kind's components, so freshly published/refreshed
// claims reach the deployed HTML — cousin of queueNewsPageRerenders
// (render_news_section_html.go). recurrenceExpected: true for the same
// reason — a re-request every publish cycle is the design, not a failed fix.
func queueDirectoryPageRerenders(ctx context.Context, db *sql.DB, siteID uuid.UUID, profile directoryPublishProfile, logger *zap.Logger) int {
	if db == nil {
		return 0
	}

	rows, err := db.QueryContext(ctx, `
		SELECT DISTINCT p.name
		FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND cc.function IN ($2, $3)
		  AND p.build_status = 'deployed'
	`, siteID, profile.SnippetComponent, profile.ListingComponent)
	if err != nil {
		logger.Warn("queueDirectoryPageRerenders: page lookup failed", zap.Error(err))
		return 0
	}
	defer rows.Close()

	var pages []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			logger.Warn("queueDirectoryPageRerenders: scan failed", zap.Error(err))
			continue
		}
		pages = append(pages, name)
	}
	if len(pages) == 0 {
		return 0
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		logger.Warn("queueDirectoryPageRerenders: begin tx failed", zap.Error(err))
		return 0
	}
	defer tx.Rollback()

	batchID := uuid.New()
	queued := 0
	for _, page := range pages {
		item := workItem{
			siteID:       siteID,
			source:       "render_directory",
			pipeline:     "build",
			itemType:     "needs_page",
			severity:     "low",
			summary:      fmt.Sprintf("Re-render %s — %s data refreshed", page, profile.Kind),
			spec:         fmt.Sprintf(`{"reason":"section_data_resolved","page_name":%q,"directory_kind":%q}`, page, profile.Kind),
			priority:     99,
			handlerAgent: "page-build-handler",
			status:       "triaged",
			createdBy:    "render_directory",
			// itemKey deliberately keys on the PAGE only, not the kind: a page
			// carrying both a model directory and an adoption tracker needs one
			// re-render, not two. Dedup is by item_key (idx_swi_dedup).
			itemKey: fmt.Sprintf("page_rerender:%s", page),
			batchID: batchID,

			recurrenceExpected: true,
		}
		inserted, err := insertWorkItem(ctx, tx, item, logger)
		if err != nil {
			logger.Warn("queueDirectoryPageRerenders: insert failed",
				zap.String("page", page), zap.Error(err))
			continue
		}
		if inserted {
			queued++
		}
	}
	if err := tx.Commit(); err != nil {
		logger.Warn("queueDirectoryPageRerenders: commit failed", zap.Error(err))
		return 0
	}

	if queued > 0 {
		logger.Info("queueDirectoryPageRerenders: queued scoped re-renders",
			zap.Int("queued", queued), zap.Int("pages", len(pages)))
	}
	return queued
}
