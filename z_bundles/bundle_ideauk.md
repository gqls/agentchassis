# Task bundle

- **Task:** On idea.uk's freshly built index page the differentiators section renders its heading but seven empty cards — every item title and description blank — while the same page's writer-generated method narrative and 13-item FAQ populated correctly; since reconcile_section_data (wired 9 June) only re-triggers pages whose deferred section data is query-resolvable, we need to establish where a differentiators component's items are meant to come from — query-resolved section data, a human-entered spec field, or page-content-writer prose — and fix whichever link is leaving them empty.
- **Step:** debug
- **In scope:** platform/orchestration/actions/v3_site_actions.go, platform/orchestration/actions/reconcile_section_data_action.go, platform/orchestration/actions/registry.go
- **Repo:** /home/ant/projects/agentchassis/ (533 files analysed)
- **Note:** in-scope items shown in full; the surrounding package(s) are shown as signatures; everything else is omitted — ask for a path to add it.

---

## Constitution (always-on rules)

# Constitution (thin slice)

The always-on rules for any task on this codebase. Included in full in every bundle. These are the things true regardless of task; task-specific standards (the detailed contracts in 003) are listed at the end and included only when a task touches them.

This is the flat-file version for the thin slice. Later it becomes the `standards` rows with `scope = constitution`; the content is the same.

---

## Reuse and structure

- **Reuse before recreate.** Before writing a new function, struct, or component, look for an existing one that does the same or similar and improve/alter it. Recreating something the system already has is a defect, not a shortcut.
- **Fix structural problems, not symptoms.** Prefer the fix that addresses the cause over the quick patch, even when the patch is faster. Note when you are knowingly deferring a structural fix.

## Agents and workflows

- **Every agent is an orchestrator.**
- **Distinct responsibilities, minimal overlap.** Each agent owns its area; agents overlap as little as possible.
- **Reply to the caller's responses topic.** An agent always responds on its parent's (caller's) responses topic, never on its own responses topic.
- **Workflows stay simple; complexity lives in Go action code.** Keep the workflow declarations thin and put the real logic in the actions.
- **No subworkflows in SQL — spawn sub-agents instead.** When work needs to branch, spawn a sub-agent with its own workflow rather than nesting subworkflows in SQL. This keeps logs clear, maintenance easier, and responsibilities separate.
- **Keep workflow variable names in sync with what the actions expect.** A workflow variable name must match the name the action reads. Do not let them drift.

## Code and data conventions

- **Check the database schema before writing SQL.** Always inspect the real schema first; never write SQL against an assumed shape.
- **Parameterised queries only.** Pass values as query parameters. Never build SQL by interpolating values into a template string.
- **Don't change variable names silently.** Keep names stable; if a rename is deliberate, say so explicitly and note it.
- **String-value naming — the two-and-a-half conventions (003):**
  - `snake_case` for **identifier-shaped** values used as keys in code: `switch case` constants, `map` keys, action-registry names, work-item `item_type`, dispatch routing, Kafka topic segments, k8s labels. (e.g. `needs_blog_posts`, `create_blog_posts`.)
  - `kebab-case` for **data-shaped** values that describe what a thing is and never act as code identifiers — they end up in CSS, URLs, HTML, prompts. (e.g. `social-proof`, `blog-post`.)
  - lowercase single word where no separator is needed (e.g. `planned`, `triaged`).
  - The test: does any Go `case`/`map` key/route/label use the value? Yes → snake. No → kebab.
- **Storage conventions (chassis):** enum-like columns are `text` + `CHECK`, not native enums; versioned entities use `version` + `previous_version_id`; soft-delete via `deleted_at`, not a `status = archived`.

## Logging

- **Don't use `logger.Debug`** — it does not show in the logs. Use a level that is actually emitted.
- **Put the run id in log lines.** Log the `orchestration_id` (and `correlation_id`) so a run can be traced across agents and tables. (Coverage of this is not yet verified everywhere — treat its absence in a line as unknown, not as "didn't happen.")
- **Log agent creation and the messages between agents** (headers and body), so the spawn tree and message flow are reconstructable.

## Deployment and infrastructure

- **Deployment path:** write to GitHub (or a future adapter) → GitHub Actions triggers a write to Backblaze S3.
- **Kubernetes namespaces:** main is `ai-persona-system` (e.g. `kubectl -n ai-persona-system get pods`); Kafka is in `kafka` (e.g. `kubectl -n kafka get pods`).
- **Kafka cluster:** `personae-kafka-cluster-*` (combined-pool and entity-operator pods). Use the real cluster names.

## Tone of generated text

- Plain, pragmatic, concrete. Avoid hype words and filler. This governs generated content and commit messages, not just chat.

---

## Task-specific standards (the 003 contracts) — included only when a task touches them

These are not always-on; the bundle pulls the relevant one in when the task's area matches. Listed here so the index is visible:

- Component Naming Contract (kebab `function`; one function, one active component; `data-component` flow).
- JS Content Separation Contract (HTML/JS split; asset path convention; `js_snippets`).
- Component Creation & Regeneration Contract (return statuses; version-history preservation).
- Site Component Linkage Contract (slot ↔ function mapping; `unlinked_site_components` check).
- CSS Colour Inheritance Model and Section Context (dark sections) Contract.
- CSS Theme Template Contract (responsibility split; theme storage/lineage columns; review gate; forking rules).
- Query Database Parameterisation Contract (the parameterised-query rule above, with examples).

For the thin slice these stay in 003; when a task touches one, paste that section alongside this constitution.

---

## Task

On idea.uk's freshly built index page the differentiators section renders its heading but seven empty cards — every item title and description blank — while the same page's writer-generated method narrative and 13-item FAQ populated correctly; since reconcile_section_data (wired 9 June) only re-triggers pages whose deferred section data is query-resolvable, we need to establish where a differentiators component's items are meant to come from — query-resolved section data, a human-entered spec field, or page-content-writer prose — and fix whichever link is leaving them empty.

---

## Reference documents (standards / guides for this task)

> doc "docs/agent_docs/docs024_key_docs_latest/026_component_regeneration_flow.md": open docs/agent_docs/docs024_key_docs_latest/026_component_regeneration_flow.md: no such file or directory

### runtime.md

# Runtime evidence

Correlated by site `idea.uk`, page `index`. Most recent first.

## Recent errors (agent_error_log)

_5 row(s)._

| occurred_at | agent_type | step_name | action | error | pod_name | 
|---|---|---|---|---|---|
| 2026-06-16 19:43:04.08032+00 | image-build-handler | call_asset_deployer | call_agent | Request ac1da3a0-d76b-4d1b-a61e-aa6062e35fb1 timed out after 3 retries | agent-image-build-handler-a49d61b0-lvw5h | 
| 2026-06-16 15:34:17.411937+00 | generic | persist_roadmap_brief | write_site_spec | step persist_roadmap_brief failed: failed to execute action write_site_spec: input extraction failed: missing required fields: [spec_data] | agent-chassis-6548d4847f-8jbdk | 
| 2026-06-16 15:34:17.356365+00 | generic | persist_roadmap | write_site_spec | step persist_roadmap failed: failed to execute action write_site_spec: input extraction failed: missing required fields: [spec_data] | agent-chassis-6548d4847f-8jbdk | 
| 2026-06-16 15:34:15.972491+00 | generic | persist_mission_brief | write_site_spec | step persist_mission_brief failed: failed to execute action write_site_spec: input extraction failed: missing required fields: [spec_data] | agent-chassis-6548d4847f-8jbdk | 
| 2026-06-16 15:34:15.917725+00 | generic | persist_mission | write_site_spec | step persist_mission failed: failed to execute action write_site_spec: input extraction failed: missing required fields: [spec_data] | agent-chassis-6548d4847f-8jbdk | 

## Work-item lifecycle (site_work_items)

_6 row(s)._

| created_at | item_type | status | claimed_by | attempts | error | item_key | 
|---|---|---|---|---|---|---|
| 2026-06-16 21:05:06.218715+00 | page_rerender | complete | build-dispatch-loop | 0/3 |  | page_rerender_index_97ed2f64-65ca-4b67-8a98-dfd8195a0d3a | 
| 2026-06-16 19:29:07.105383+00 | needs_page | complete | build-dispatch-loop | 2/3 | Claim timed out — handler pod likely died | page_rerender:index | 
| 2026-06-16 15:49:50.075639+00 | unresolved_cta | needs_human_review |  | 0/3 |  | unresolved_cta_index_call-to-action_97ed2f64-65ca-4b67-8a98-dfd8195a0d3a | 
| 2026-06-16 15:49:50.071754+00 | unresolved_cta | needs_human_review |  | 0/3 |  | unresolved_cta_index_hero_97ed2f64-65ca-4b67-8a98-dfd8195a0d3a | 
| 2026-06-16 15:48:35.219262+00 | needs_section_data | needs_human_review |  | 0/3 |  | section_data_index_pricing_97ed2f64-65ca-4b67-8a98-dfd8195a0d3a | 
| 2026-06-16 15:46:14.183917+00 | needs_page | complete | build-dispatch-loop | 0/3 |  | needs_page:index |

---

## In-scope code

### platform/orchestration/actions/v3_site_actions.go (package `actions`) — whole file

```go
// FILE: platform/orchestration/actions/v3_site_actions.go
// Additional actions needed for the v3 multipage website builder component-based architecture.
// These complement existing actions in site_db_actions.go and component_library.go.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"github.com/gqls/agentchassis/platform/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// SelectStyleCollectionAction selects a style collection for a site
// Either by explicit ID, by site_id lookup, or by domain/industry matching
// Config:
//   - style_collection_id: explicit UUID (optional)
//   - site_id_field: path to site_id in collected_data (optional)
//   - domain_field: path to domain for industry matching (optional)
func SelectStyleCollectionAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("SelectStyleCollectionAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	if params.DB == nil {
		params.Logger.Warn("SelectStyleCollectionAction: No database, returning default style")
		return getDefaultStyleCollection(), nil
	}

	// Resolve site_id early — needed for persist step at the end
	var siteID uuid.UUID
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = parsed
			}
		}
	}

	// Helper: persist and return
	persistAndReturn := func(coll *StyleCollection, source string) (interface{}, error) {
		params.Logger.Info("SelectStyleCollectionAction: Resolved style",
			zap.String("name", coll.Name),
			zap.String("source", source),
			zap.String("id", coll.ID.String()))

		// Persist style_collection_id to sites table so downstream agents
		// (webdesign-agent, maintenance agents) can find it via DB lookup
		if siteID != uuid.Nil {
			query := `UPDATE sites SET style_collection_id = $2, updated_at = NOW() WHERE id = $1`
			if err := execDB(ctx, params.DB, query, siteID, coll.ID); err != nil {
				params.Logger.Warn("SelectStyleCollectionAction: Failed to persist style_collection_id",
					zap.Error(err))
				// Non-fatal — continue with the result
			} else {
				params.Logger.Info("SelectStyleCollectionAction: Persisted style_collection_id to sites table",
					zap.String("site_id", siteID.String()),
					zap.String("style_collection_id", coll.ID.String()))
			}
		}

		return styleCollectionToResult(coll), nil
	}

	// Priority 1: Explicit style_collection_id in config
	if scID, ok := config["style_collection_id"].(string); ok && scID != "" {
		scUUID, err := uuid.Parse(scID)
		if err == nil {
			coll, err := getStyleCollectionByID(ctx, params.DB, scUUID, params.Logger)
			if err == nil {
				return persistAndReturn(coll, "explicit_id")
			}
		}
	}

	// Priority 2 (NEW): Planner's style choice via style_from config
	// Config example: "style_from": "site_plan.style_collection"
	// The planner writes a style name (e.g. "professional-dark") to that path
	if styleFromField, ok := config["style_from"].(string); ok && styleFromField != "" {
		styleName := datahelpers.ExtractNestedFieldString(params.CollectedData, styleFromField)
		if styleName != "" {
			params.Logger.Info("SelectStyleCollectionAction: Trying planner style choice",
				zap.String("style_from", styleFromField),
				zap.String("style_name", styleName))
			coll, err := GetStyleCollectionByName(ctx, params.DB, styleName, params.Logger)
			if err == nil && coll != nil {
				return persistAndReturn(coll, "planner_style_from")
			}
			params.Logger.Warn("SelectStyleCollectionAction: Planner style not found in DB",
				zap.String("style_name", styleName))
		}
	}

	// Priority 3: Look up by site_id (existing style_collection_id on sites table)
	if siteID != uuid.Nil {
		coll, err := GetStyleCollectionForSite(ctx, params.DB, siteID, params.Logger)
		if err == nil && coll != nil {
			// Already persisted — just return
			return styleCollectionToResult(coll), nil
		}
	}

	// Priority 4: Match by domain keywords
	domainField := "input_data.domain"
	if df, ok := config["domain_field"].(string); ok && df != "" {
		domainField = df
	}
	domain := datahelpers.ExtractNestedFieldString(params.CollectedData, domainField)
	if domain != "" {
		coll, err := SelectStyleCollectionByDomain(ctx, params.DB, domain, params.Logger)
		if err == nil && coll != nil {
			return persistAndReturn(coll, "domain_keywords")
		}
	}

	// Fallback: Return default (not persisted — it's a synthetic default)
	params.Logger.Info("SelectStyleCollectionAction: No matching style found, using default")
	return getDefaultStyleCollection(), nil
}

func styleCollectionToResult(coll *StyleCollection) map[string]interface{} {
	result := map[string]interface{}{
		"style_collection_id": coll.ID.String(),
		"name":                coll.Name,
		"display_name":        coll.DisplayName,
		"category":            coll.Category,
		"color_palette":       coll.ColorPalette,
		"typography":          coll.Typography,
	}
	if coll.HeaderComponentID != nil {
		result["header_component_id"] = coll.HeaderComponentID.String()
	}
	if coll.FooterComponentID != nil {
		result["footer_component_id"] = coll.FooterComponentID.String()
	}
	if coll.CSSThemeID != nil {
		result["css_theme_id"] = coll.CSSThemeID.String()
	}
	return result
}

func getDefaultStyleCollection() map[string]interface{} {
	return map[string]interface{}{
		"style_collection_id": "",
		"name":                "default",
		"display_name":        "Default Style",
		"category":            "general",
		"color_palette": map[string]string{
			"primary":    "#1a1a2e",
			"secondary":  "#16213e",
			"accent":     "#0f3460",
			"text":       "#333333",
			"background": "#ffffff",
		},
		"typography": map[string]string{
			"heading": "system-ui, sans-serif",
			"body":    "system-ui, sans-serif",
		},
	}
}

// ============================================================================
// ACTION: update_site_content
// ============================================================================

// UpdateSiteContentAction updates the sites.content_data JSONB column
// Used to store the site plan, brand DNA, or other structured content
// Config:
//   - site_id_field: path to site_id in collected_data
//   - content_field: path to content data to store
//   - merge: boolean - if true, merges with existing content_data
func UpdateSiteContentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteContentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get content to store
	contentField := "page_plan"
	if f, ok := config["content_field"].(string); ok && f != "" {
		contentField = f
	}
	contentValue := datahelpers.ExtractNestedField(params.CollectedData, contentField)
	if contentValue == nil {
		return nil, fmt.Errorf("content not found at %s", contentField)
	}

	contentJSON, err := json.Marshal(contentValue)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteContentAction: No database, skipping update")
		return map[string]interface{}{
			"updated":     false,
			"site_id":     siteIDStr,
			"reason":      "no database connection",
			"content_key": contentField,
		}, nil
	}

	// Determine if merge or replace
	merge, _ := config["merge"].(bool)

	var query string
	if merge {
		query = `
			UPDATE sites 
			SET content_data = COALESCE(content_data, '{}'::jsonb) || $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	} else {
		query = `
			UPDATE sites 
			SET content_data = $2::jsonb,
			    updated_at = NOW()
			WHERE id = $1
		`
	}

	if err := execDB(ctx, params.DB, query, siteID, string(contentJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site content: %w", err)
	}

	// --- Sync key columns to sites table ---
	// When storing brief/identity data, also populate the sites columns
	// so loadSiteDataFull and RenderSiteComponentsAction can find them.
	syncColumns, _ := config["sync_columns"].(bool)
	if syncColumns {
		if contentMap, ok := contentValue.(map[string]interface{}); ok {
			companyName := getFirstNonEmpty(contentMap, "company_name")
			tagline := getFirstNonEmpty(contentMap, "tagline")
			email := getFirstNonEmpty(contentMap, "contact_email", "email")
			phone := getFirstNonEmpty(contentMap, "contact_phone", "phone")

			if companyName != "" || tagline != "" || email != "" || phone != "" {
				syncErr := execDB(ctx, params.DB, `
					UPDATE sites SET
						company_name = CASE WHEN COALESCE(company_name, '') IN ('', domain) AND $2 != '' THEN $2 ELSE company_name END,
						tagline      = CASE WHEN COALESCE(tagline, '')      = '' AND $3 != '' THEN $3 ELSE tagline END,
						email        = CASE WHEN COALESCE(email, '')        = '' AND $4 != '' THEN $4 ELSE email END,
						phone        = CASE WHEN COALESCE(phone, '')        = '' AND $5 != '' THEN $5 ELSE phone END,
						updated_at   = now()
					WHERE id = $1
				`, siteID, companyName, tagline, email, phone)
				if syncErr != nil {
					params.Logger.Warn("UpdateSiteContentAction: column sync failed", zap.Error(syncErr))
				} else {
					params.Logger.Info("UpdateSiteContentAction: synced columns",
						zap.String("company_name", companyName),
						zap.String("email", email),
					)
				}
			}
		}
	}

	params.Logger.Info("UpdateSiteContentAction: Site content updated",
		zap.String("site_id", siteIDStr),
		zap.String("content_field", contentField),
		zap.Bool("merged", merge),
	)

	return map[string]interface{}{
		"updated":      true,
		"site_id":      siteIDStr,
		"content_key":  contentField,
		"content_size": len(contentJSON),
		"merged":       merge,
	}, nil
}

// getFirstNonEmpty returns the first non-empty string value found at any of
// the given keys in the map.
func getFirstNonEmpty(m map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key].(string); ok && v != "" {
			return v
		}
	}
	return ""
}

// ============================================================================
// ACTION: update_site_status
// ============================================================================

// UpdateSiteStatusAction updates the sites.status column
// Config:
//   - site_id_field: path to site_id
//   - status: new status value (draft, building, review, published, archived)
func UpdateSiteStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Get new status
	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required in config")
	}

	// Validate status
	validStatuses := map[string]bool{
		"draft": true, "building": true, "review": true,
		"published": true, "deployed": true, "archived": true, "error": true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid status: %s (valid: draft, building, review, published, deployed, archived, error)", newStatus)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	// Check if deployed_at should be set
	var query string
	deployedAt, hasDeployedAt := config["deployed_at"].(string)
	if hasDeployedAt && (deployedAt == "now" || deployedAt == "NOW()") && newStatus == "deployed" {
		query = `UPDATE sites SET status = $2, last_deployed_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else {
		query = `UPDATE sites SET status = $2, updated_at = NOW() WHERE id = $1`
	}

	if err := execDB(ctx, params.DB, query, siteID, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update site status: %w", err)
	}

	params.Logger.Info("UpdateSiteStatusAction: Status updated",
		zap.String("site_id", siteIDStr),
		zap.String("status", newStatus),
	)

	return map[string]interface{}{
		"updated":   true,
		"site_id":   siteIDStr,
		"status":    newStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// ============================================================================
// ACTION: update_site_defaults
// ============================================================================

// UpdateSiteDefaultsAction updates the sites.default_components JSONB column
// Used to store default header/footer component IDs
// Config:
//   - site_id_field: path to site_id in collected_data
//   - header_component_id: UUID of header component (optional)
//   - footer_component_id: UUID of footer component (optional)
//   - css_theme_id: UUID of CSS theme (optional)
//   - defaults_field: path to a map containing all defaults (alternative to individual fields)
func UpdateSiteDefaultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateSiteDefaultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	if siteIDStr == "" {
		return nil, fmt.Errorf("site_id not found at %s", siteIDField)
	}

	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Build defaults map
	defaults := make(map[string]interface{})

	// Try getting from defaults_field first
	if defaultsField, ok := config["defaults_field"].(string); ok && defaultsField != "" {
		if defaultsData := datahelpers.ExtractNestedField(params.CollectedData, defaultsField); defaultsData != nil {
			if m, ok := defaultsData.(map[string]interface{}); ok {
				defaults = m
			}
		}
	}

	// Override/add from explicit config fields
	if headerID, ok := config["header_component_id"].(string); ok && headerID != "" {
		defaults["header_component_id"] = headerID
	}
	if footerID, ok := config["footer_component_id"].(string); ok && footerID != "" {
		defaults["footer_component_id"] = footerID
	}
	if cssThemeID, ok := config["css_theme_id"].(string); ok && cssThemeID != "" {
		defaults["css_theme_id"] = cssThemeID
	}

	// Also check collected data for style_collection results
	if sc := datahelpers.ExtractNestedField(params.CollectedData, "style_collection"); sc != nil {
		if scMap, ok := sc.(map[string]interface{}); ok {
			if hID, ok := scMap["header_component_id"].(string); ok && hID != "" {
				if defaults["header_component_id"] == nil {
					defaults["header_component_id"] = hID
				}
			}
			if fID, ok := scMap["footer_component_id"].(string); ok && fID != "" {
				if defaults["footer_component_id"] == nil {
					defaults["footer_component_id"] = fID
				}
			}
			if cID, ok := scMap["css_theme_id"].(string); ok && cID != "" {
				if defaults["css_theme_id"] == nil {
					defaults["css_theme_id"] = cID
				}
			}
		}
	}

	if len(defaults) == 0 {
		params.Logger.Warn("UpdateSiteDefaultsAction: No defaults to set")
		return map[string]interface{}{
			"updated": false,
			"site_id": siteIDStr,
			"reason":  "no defaults provided",
		}, nil
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateSiteDefaultsAction: No database")
		return map[string]interface{}{
			"updated":  false,
			"site_id":  siteIDStr,
			"defaults": defaults,
			"reason":   "no database connection",
		}, nil
	}

	defaultsJSON, err := json.Marshal(defaults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal defaults: %w", err)
	}

	query := `
		UPDATE sites 
		SET default_components = COALESCE(default_components, '{}'::jsonb) || $2::jsonb,
		    updated_at = NOW()
		WHERE id = $1
	`

	if err := execDB(ctx, params.DB, query, siteID, string(defaultsJSON)); err != nil {
		return nil, fmt.Errorf("failed to update site defaults: %w", err)
	}

	params.Logger.Info("UpdateSiteDefaultsAction: Defaults updated",
		zap.String("site_id", siteIDStr),
		zap.Any("defaults", defaults),
	)

	return map[string]interface{}{
		"updated":  true,
		"site_id":  siteIDStr,
		"defaults": defaults,
	}, nil
}

// ============================================================================
// ACTION: update_page_status
// ============================================================================

// UpdatePageStatusAction updates a single page's build_status
// Config:
//   - page_id_field: path to page_id OR
//   - site_id_field + page_name_field: to look up page
//   - status: new build_status value (e.g., "deployed", "failed", "building")
func UpdatePageStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageStatusAction: Starting",
		zap.String("step_name", params.ExecutionContext.StepName),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("UpdatePageStatusAction: No database connection available")
		return map[string]interface{}{"updated": false, "reason": "no database"}, nil
	}

	newStatus, ok := config["status"].(string)
	if !ok || newStatus == "" {
		return nil, fmt.Errorf("status is required")
	}

	var pageID uuid.UUID

	// Try direct page_id from config field
	if pageIDField, ok := config["page_id_field"].(string); ok && pageIDField != "" {
		pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, pageIDField)
		if pageIDStr != "" {
			var err error
			pageID, err = uuid.Parse(pageIDStr)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Invalid page_id format",
					zap.String("page_id_field", pageIDField),
					zap.String("value", pageIDStr),
					zap.Error(err))
				return nil, fmt.Errorf("invalid page_id: %w", err)
			}
		}
	}

	// Alternative: try current_page.id (common in loop iterations)
	if pageID == uuid.Nil {
		if pageIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.id"); pageIDStr != "" {
			if parsed, err := uuid.Parse(pageIDStr); err == nil {
				pageID = parsed
			}
		}
	}

	// Alternative: look up by site_id + page_name
	if pageID == uuid.Nil {
		siteIDField, _ := config["site_id_field"].(string)
		pageNameField, _ := config["page_name_field"].(string)

		if siteIDField != "" && pageNameField != "" {
			siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
			pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, pageNameField)

			if siteIDStr != "" && pageName != "" {
				siteUUID, _ := uuid.Parse(siteIDStr)
				var err error
				pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
				if err != nil {
					return nil, fmt.Errorf("page lookup failed: %w", err)
				}
			}
		}
	}

	// Last resort: try current_page.name with site_record.site_id
	if pageID == uuid.Nil {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		pageName := datahelpers.ExtractNestedFieldString(params.CollectedData, "current_page.name")

		if siteIDStr != "" && pageName != "" {
			siteUUID, _ := uuid.Parse(siteIDStr)
			var err error
			pageID, err = lookupPageID(ctx, params.DB, siteUUID, pageName, params.Logger)
			if err != nil {
				params.Logger.Warn("UpdatePageStatusAction: Page lookup by name failed",
					zap.String("site_id", siteIDStr),
					zap.String("page_name", pageName),
					zap.Error(err))
			}
		}
	}

	if pageID == uuid.Nil {
		params.Logger.Error("UpdatePageStatusAction: Could not determine page_id",
			zap.Any("config", config))
		return nil, fmt.Errorf("could not determine page_id")
	}

	// Option B guard: never mark a page "deployed" without rendered components.
	// build_status='deployed' is what the reconciler trusts to skip a page, so a
	// 0-component page marked deployed becomes permanently fileless (gamesdesign
	// homepage, 2026-06-04: its needs_content_page was auto-completed on a lost
	// response with zero components, then a deploy path stamped it deployed, so
	// the reconciler never rebuilt it). Refuse the deploy mark and flip to
	// needs_rebuild (clearing the stamp) so the reconciler re-emits a build.
	// Fail-open on a check error: a transient check failure must not halt
	// legitimate deploys; Option A (the claimed-item-timeout evidence check) is
	// the other layer of protection.
	if newStatus == "deployed" {
		hasComponents, checkErr := pageHasComponents(ctx, params.DB, pageID)
		if checkErr != nil {
			params.Logger.Warn("UpdatePageStatusAction: component check failed; proceeding with deploy",
				zap.String("page_id", pageID.String()),
				zap.Error(checkErr))
		} else if !hasComponents {
			params.Logger.Warn("UpdatePageStatusAction: refusing to mark page deployed with no rendered components; setting needs_rebuild",
				zap.String("page_id", pageID.String()))
			const rebuildQuery = `UPDATE pages SET build_status = 'needs_rebuild', built_from_plan_version = NULL, updated_at = NOW() WHERE id = $1`
			if rbErr := execDB(ctx, params.DB, rebuildQuery, pageID); rbErr != nil {
				params.Logger.Error("UpdatePageStatusAction: failed to set needs_rebuild after refusing deploy",
					zap.String("page_id", pageID.String()),
					zap.Error(rbErr))
				return nil, fmt.Errorf("failed to set needs_rebuild for 0-component page: %w", rbErr)
			}
			return map[string]interface{}{
				"updated":      false,
				"page_id":      pageID.String(),
				"build_status": "needs_rebuild",
				"reason":       "refused deploy: page has no rendered components",
			}, nil
		}
	}

	// Build the query - use build_status column (not status)
	var query string
	if newStatus == "deployed" {
		// Also set deployed_at, and stamp built_from_plan_version with the site's
		// current plan id. This is the build-time drift stamp the reconciler
		// compares against (029/030 design; the deferred item in HANDOFF_2026-05-07
		// #5). COALESCE keeps any existing value when no current plan exists yet —
		// e.g. tool-recreation deploys before build-site-planner has written the
		// plan — and SyncPagesToDBAction then fills it on its first pass. With this
		// stamp in place the reconciler detects genuine drift (built_from_plan_version
		// != current) rather than relying on the blunt deployed->needs_rebuild flip.
		query = `UPDATE pages
		         SET build_status = $2,
		             deployed_at = NOW(),
		             built_from_plan_version = COALESCE(
		                 (SELECT sp.id FROM site_plans sp
		                   WHERE sp.site_id = pages.site_id AND sp.is_current = true),
		                 built_from_plan_version
		             ),
		             updated_at = NOW()
		         WHERE id = $1`
	} else {
		query = `UPDATE pages SET build_status = $2, updated_at = NOW() WHERE id = $1`
	}

	result, err := params.DB.ExecContext(ctx, query, pageID, newStatus)
	if err != nil {
		params.Logger.Error("UpdatePageStatusAction: Failed to update page",
			zap.String("page_id", pageID.String()),
			zap.String("build_status", newStatus),
			zap.Error(err))
		return nil, fmt.Errorf("failed to update page build_status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()

	params.Logger.Info("UpdatePageStatusAction: Updated page",
		zap.String("page_id", pageID.String()),
		zap.String("build_status", newStatus),
		zap.Int64("rows_affected", rowsAffected))

	return map[string]interface{}{
		"updated":       true,
		"page_id":       pageID.String(),
		"build_status":  newStatus,
		"rows_affected": rowsAffected,
	}, nil
}

// pageHasComponents reports whether a page has at least one real rendered
// component (non-null component_id, non-empty rendered_html). This is the
// "positive evidence" check from FOCUS_page_build_handler_silent_completion.md,
// used by Option B to stop a 0-component page being marked deployed. Mirrors the
// db type switch used by lookupPageID/execDB so it works with both *sql.DB and
// *pgxpool.Pool.
func pageHasComponents(ctx context.Context, db interface{}, pageID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM page_components
			WHERE page_id = $1
			  AND component_id IS NOT NULL
			  AND rendered_html IS NOT NULL
			  AND rendered_html <> ''
		)`
	var exists bool
	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, pageID).Scan(&exists)
		return exists, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, pageID).Scan(&exists)
		return exists, err
	default:
		return false, fmt.Errorf("unsupported database type: %T", db)
	}
}

// BuildRenderContextAction assembles a RenderContext from multiple sources
// Used before rendering components with templates
func BuildRenderContextAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildRenderContextAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Start with empty context
	renderCtx := &RenderContext{
		Year: fmt.Sprintf("%d", time.Now().Year()),
	}

	sourcesMerged := 0

	// Handle sources config - can be array or map format
	// Array format: ["input_data", "site_record", ...]
	// Map format: {"page": "input_data.current_page", "site": "input_data.site_record", ...}

	if sourcesMap, ok := config["sources"].(map[string]interface{}); ok {
		// Map format: keys are logical names, values are paths to data
		params.Logger.Info("Using map-format sources config",
			zap.Int("source_count", len(sourcesMap)))

		for logicalName, pathVal := range sourcesMap {
			path, ok := pathVal.(string)
			if !ok {
				continue
			}

			sourceData := datahelpers.ExtractNestedField(params.CollectedData, path)
			if sourceData == nil {
				params.Logger.Debug("Source not found",
					zap.String("name", logicalName),
					zap.String("path", path))
				continue
			}

			if m, ok := sourceData.(map[string]interface{}); ok {
				params.Logger.Debug("Merging source",
					zap.String("name", logicalName),
					zap.String("path", path))
				mergeIntoRenderContextEnhanced(renderCtx, m, logicalName, params.Logger)
				sourcesMerged++
			}
		}
	} else if sourcesArray, ok := config["sources"].([]interface{}); ok {
		// Array format: direct paths
		for _, src := range sourcesArray {
			if srcStr, ok := src.(string); ok {
				sourceData := datahelpers.ExtractNestedField(params.CollectedData, srcStr)
				if sourceData == nil {
					continue
				}
				if m, ok := sourceData.(map[string]interface{}); ok {
					mergeIntoRenderContextEnhanced(renderCtx, m, srcStr, params.Logger)
					sourcesMerged++
				}
			}
		}
	} else {
		// Default sources
		defaultSources := []string{"input_data", "site_record", "style_collection", "reviewed_brief", "page_plan"}
		for _, source := range defaultSources {
			sourceData := datahelpers.ExtractNestedField(params.CollectedData, source)
			if sourceData == nil {
				// Try with input_data prefix
				sourceData = datahelpers.ExtractNestedField(params.CollectedData, "input_data."+source)
			}
			if sourceData == nil {
				continue
			}
			if m, ok := sourceData.(map[string]interface{}); ok {
				mergeIntoRenderContextEnhanced(renderCtx, m, source, params.Logger)
				sourcesMerged++
			}
		}
	}

	// Try to load navigation from DB if we have site_id
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if siteUUID, err := uuid.Parse(siteIDStr); err == nil && params.DB != nil {
				renderCtx.SiteID = siteUUID
				/*nav, _ := getNavigationFromDB(ctx, params.DB, siteUUID, "header", params.Logger)*/
				headerNav := GetNavItems(ctx, params.DB, siteUUID, []string{NavGroupPrimary}, false, 0, params.Logger)
				if len(headerNav) > 0 {
					renderCtx.NavItems = headerNav
				}
			}
		}
	}

	// Extract image URLs from deploy_image_asset output
	// (adds logo_deployed block between hero_deployed and fallback)
	// =========================================================================
	if heroDeployed, ok := params.CollectedData["hero_deployed"].(map[string]interface{}); ok {
		if imageURL, ok := heroDeployed["image_url"].(string); ok && imageURL != "" {
			if renderCtx.ContentData == nil {
				renderCtx.ContentData = make(map[string]interface{})
			}
			renderCtx.ContentData["hero_url"] = imageURL
			params.Logger.Info("Set hero_url from hero_deployed.image_url",
				zap.String("url", imageURL))
		}
	}

	if logoDeployed, ok := params.CollectedData["logo_deployed"].(map[string]interface{}); ok {
		if imageURL, ok := logoDeployed["image_url"].(string); ok && imageURL != "" {
			if renderCtx.ContentData == nil {
				renderCtx.ContentData = make(map[string]interface{})
			}
			renderCtx.ContentData["logo_url"] = imageURL
			renderCtx.LogoURL = imageURL
			params.Logger.Info("Set logo_url from logo_deployed.image_url",
				zap.String("url", imageURL))
		}
	}

	// Also check for direct hero_url in collected_data (fallback)
	for _, field := range []string{"hero_url", "hero_home_url", "logo_url"} {
		if url := datahelpers.ExtractNestedFieldString(params.CollectedData, field); url != "" {
			if renderCtx.ContentData == nil {
				renderCtx.ContentData = make(map[string]interface{})
			}
			renderCtx.ContentData[field] = url
			if field == "logo_url" {
				renderCtx.LogoURL = url
			}
		}
	}

	params.Logger.Info("BuildRenderContextAction: Context built",
		zap.String("domain", renderCtx.Domain),
		zap.String("company_name", renderCtx.CompanyName),
		zap.String("logo_text", renderCtx.LogoText),
		zap.Int("nav_items", len(renderCtx.NavItems)),
		zap.Int("sources_merged", sourcesMerged),
	)

	// Return the context directly, not wrapped in "render_context" key
	// The workflow output_field already specifies where to store it
	// Adding metadata fields at same level
	result := renderCtxToMap(renderCtx)
	result["_sources_merged"] = sourcesMerged
	result["_built_at"] = time.Now().Format(time.RFC3339)

	return result, nil
}

// mergeIntoRenderContextEnhanced extracts data from various source formats
func mergeIntoRenderContextEnhanced(ctx *RenderContext, data map[string]interface{}, sourceName string, logger *zap.Logger) {
	// =========================================================================
	// STEP 1: Unwrap .response wrapper if present (common in agent responses)
	// reviewed_brief has: {"response": {...actual data...}, "response_status": "complete"}
	// This recursively processes the unwrapped data first
	// =========================================================================
	if response, ok := data["response"].(map[string]interface{}); ok {
		// Check if this looks like a wrapped response (has response_status sibling)
		if _, hasStatus := data["response_status"]; hasStatus {
			logger.Debug("Unwrapping .response wrapper for source",
				zap.String("source", sourceName))
			// Recursively merge the unwrapped response data FIRST
			mergeIntoRenderContextEnhanced(ctx, response, sourceName+".response", logger)
			// Then continue to process the outer data for any additional fields
		}
	}

	// =========================================================================
	// STEP 2: Direct field extraction from current data level
	// =========================================================================

	// Domain
	if v, ok := data["domain"].(string); ok && v != "" {
		ctx.Domain = v
	}

	// Company name (sets logo_text as fallback)
	if v, ok := data["company_name"].(string); ok && v != "" {
		ctx.CompanyName = v
		if ctx.LogoText == "" {
			ctx.LogoText = v
		}
	}

	// Logo text (explicit override)
	if v, ok := data["logo_text"].(string); ok && v != "" {
		ctx.LogoText = v
	}

	// Tagline
	if v, ok := data["tagline"].(string); ok && v != "" {
		ctx.Tagline = v
	}

	// Email - check both "email" and "contact_email" (reviewed_brief uses contact_email)
	if v, ok := data["email"].(string); ok && v != "" {
		ctx.Email = v
	}
	if v, ok := data["contact_email"].(string); ok && v != "" {
		ctx.Email = v
	}

	// Phone - check both "phone" and "contact_phone"
	if v, ok := data["phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["contact_phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["contact_email"].(string); ok && v != "" {
		ctx.Email = v
	}

	// Extract image URLs from content_data or collected_data
	imageURLFields := []string{
		"hero_url",      // pageflow-builder uses this
		"hero_home_url", // multipage-website-builder uses this
		"hero_about_url",
		"hero_services_url",
		"logo_url",
	}

	for _, field := range imageURLFields {
		if v, ok := data[field].(string); ok && v != "" {
			if ctx.ContentData == nil {
				ctx.ContentData = make(map[string]interface{})
			}
			ctx.ContentData[field] = v
			logger.Debug("Extracted image URL",
				zap.String("field", field),
				zap.String("url", v))
		}
	}

	// Colors - direct fields
	if v, ok := data["primary_color"].(string); ok && v != "" {
		ctx.PrimaryColor = v
	}
	if v, ok := data["secondary_color"].(string); ok && v != "" {
		ctx.SecondaryColor = v
	}
	if v, ok := data["accent_color"].(string); ok && v != "" {
		ctx.AccentColor = v
	}
	if v, ok := data["text_color"].(string); ok && v != "" {
		ctx.TextColor = v
	}
	if v, ok := data["background_color"].(string); ok && v != "" {
		ctx.BackgroundColor = v
	}

	// =========================================================================
	// STEP 3: Check nested color_palette (from style_collection)
	// =========================================================================
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok && v != "" {
			ctx.AccentColor = v
		}
		if v, ok := palette["background"].(string); ok && v != "" {
			ctx.BackgroundColor = v
		}
		if v, ok := palette["text"].(string); ok && v != "" {
			ctx.TextColor = v
		}
	}

	// =========================================================================
	// STEP 4: Extract from nested structures (business_context, contact_info, brand)
	// =========================================================================

	// business_context (alternative structure)
	if brief, ok := data["business_context"].(map[string]interface{}); ok {
		if v, ok := brief["company_name"].(string); ok && v != "" {
			ctx.CompanyName = v
			if ctx.LogoText == "" {
				ctx.LogoText = v
			}
		}
		if v, ok := brief["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
		if v, ok := brief["industry"].(string); ok && v != "" {
			ctx.Industry = v
		}
	}

	// contact_info (nested contact structure)
	if contact, ok := data["contact_info"].(map[string]interface{}); ok {
		if v, ok := contact["email"].(string); ok && v != "" {
			ctx.Email = v
		}
		if v, ok := contact["phone"].(string); ok && v != "" {
			ctx.Phone = v
		}
	}

	// brand (nested brand/visual settings)
	if brand, ok := data["brand"].(map[string]interface{}); ok {
		if v, ok := brand["primary_color"].(string); ok && v != "" {
			ctx.PrimaryColor = v
		}
		if v, ok := brand["secondary_color"].(string); ok && v != "" {
			ctx.SecondaryColor = v
		}
		if v, ok := brand["tagline"].(string); ok && v != "" {
			ctx.Tagline = v
		}
	}

	// =========================================================================
	// STEP 5: Content generation context (tone, target_audience, industry)
	// =========================================================================
	if v, ok := data["tone"].(string); ok && v != "" {
		ctx.Tone = v
	}
	if v, ok := data["target_audience"].(string); ok && v != "" {
		ctx.TargetAudience = v
	}
	if v, ok := data["industry"].(string); ok && v != "" {
		ctx.Industry = v
	}

	// =========================================================================
	// STEP 6: Site/page identifiers
	// =========================================================================
	if v, ok := data["site_id"].(string); ok && v != "" {
		if siteUUID, err := uuid.Parse(v); err == nil {
			ctx.SiteID = siteUUID
		}
	}

	// =========================================================================
	// STEP 7: CTA settings
	// =========================================================================
	if v, ok := data["cta_text"].(string); ok && v != "" {
		ctx.CTAText = v
	}
	if v, ok := data["cta_url"].(string); ok && v != "" {
		ctx.CTAUrl = v
	}

	// =========================================================================
	// STEP 8: Extract services array (for footer and services sections)
	// Services appear in reviewed_brief.response.services as []interface{}
	// Each service is {"name": "...", "description": "..."}
	//
	// ctx.Services is []string (just names)
	// ctx.ContentData["services"] is []interface{} (full objects for {{range .services}})
	// =========================================================================
	if services, ok := data["services"].([]interface{}); ok && len(services) > 0 && len(ctx.Services) == 0 {
		// Store full services in ContentData for template access via {{range .services}}
		// Only if not already populated (avoid duplicates from brief + brief.response)
		if ctx.ContentData == nil {
			ctx.ContentData = make(map[string]interface{})
		}
		// ctx.ContentData["services"] = services
		ctx.ContentData["services"] = normaliseToNameDescArray(services)

		// Also extract just the names to ctx.Services ([]string)
		for _, svc := range services {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				if name, ok := svcMap["name"].(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				}
			}
		}

		logger.Info("Extracted services array",
			zap.String("source", sourceName),
			zap.Int("full_count", len(services)),
			zap.Int("names_count", len(ctx.Services)))
	}

	// =========================================================================
	// STEP 9: Extract navigation from db_sync source
	// db_sync contains: {"navigation": {"items": [{"label": "Home", "url": "/index.html"}, ...]}}
	// =========================================================================
	if sourceName == "db_sync" || sourceName == "db_sync.response" {
		if navigation, ok := data["navigation"].(map[string]interface{}); ok {
			if items, ok := navigation["items"].([]interface{}); ok {
				for _, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						label, _ := itemMap["label"].(string)
						url, _ := itemMap["url"].(string)
						if label != "" && url != "" {
							ctx.NavItems = append(ctx.NavItems, NavItem{
								Label: label,
								URL:   url,
							})
						}
					}
				}
				if len(ctx.NavItems) > 0 {
					logger.Info("Extracted navigation items from db_sync",
						zap.Int("count", len(ctx.NavItems)))
				}
			}
		}
	}

	// Handle site_record.content_data - recursively merge nested content_data
	if sourceName == "site_record" || sourceName == "site" {
		if contentData, ok := data["content_data"].(map[string]interface{}); ok {
			logger.Debug("Processing site_record.content_data",
				zap.Int("fields", len(contentData)))
			mergeIntoRenderContextEnhanced(ctx, contentData, "site_record.content_data", logger)
		}
	}

	// Handle services array (for services_html generation)
	if len(ctx.Services) == 0 {
		if services, ok := data["services"].([]interface{}); ok && len(services) > 0 {
			for _, svc := range services {
				if name, ok := svc.(string); ok && name != "" {
					ctx.Services = append(ctx.Services, name)
				} else if svcMap, ok := svc.(map[string]interface{}); ok {
					if name, ok := svcMap["name"].(string); ok && name != "" {
						ctx.Services = append(ctx.Services, name)
					}
				}
			}
		}
	}

	// =========================================================================
	// STEP 10: Log final state for debugging
	// =========================================================================
	logger.Info("Merged source into render context",
		zap.String("source", sourceName),
		zap.String("company_name", ctx.CompanyName),
		zap.String("domain", ctx.Domain),
		zap.String("tagline", ctx.Tagline),
		zap.String("email", ctx.Email),
		zap.Int("nav_items", len(ctx.NavItems)),
		zap.Int("services", len(ctx.Services)))
}

// renderCtxToMap converts RenderContext to map for template substitution
func renderCtxToMap(ctx *RenderContext) map[string]interface{} {
	result := map[string]interface{}{
		"domain":           ctx.Domain,
		"logo_text":        ctx.LogoText,
		"company_name":     ctx.CompanyName,
		"tagline":          ctx.Tagline,
		"email":            ctx.Email,
		"phone":            ctx.Phone,
		"primary_color":    ctx.PrimaryColor,
		"secondary_color":  ctx.SecondaryColor,
		"accent_color":     ctx.AccentColor,
		"text_color":       ctx.TextColor,
		"background_color": ctx.BackgroundColor,
		"year":             ctx.Year,
		"cta_text":         ctx.CTAText,
		"cta_url":          ctx.CTAUrl,
		"industry":         ctx.Industry,
		"tone":             ctx.Tone,
		"target_audience":  ctx.TargetAudience,
	}

	if ctx.SiteID != uuid.Nil {
		result["site_id"] = ctx.SiteID.String()
	}

	// =========================================================================
	// Generate nav_items and nav_items_html from NavItems
	// Templates may use either:
	//   {{range .nav_items}} for iteration
	//   {{.nav_items_html}} for pre-rendered HTML
	// =========================================================================
	if len(ctx.NavItems) > 0 {
		// nav_items as array (for {{range .nav_items}})
		navItems := make([]map[string]interface{}, len(ctx.NavItems))
		for i, item := range ctx.NavItems {
			navItems[i] = map[string]interface{}{
				"label":     item.Label,
				"url":       item.URL,
				"is_active": item.IsActive,
			}
		}
		result["nav_items"] = navItems

		// nav_items_html as pre-rendered string (for {{.nav_items_html}})
		// Format: <li><a href="/page.html">Label</a></li>
		var htmlParts []string
		for _, item := range ctx.NavItems {
			activeClass := ""
			if item.IsActive {
				activeClass = ` class="active"`
			}
			htmlParts = append(htmlParts, fmt.Sprintf(
				`<li><a href="%s"%s>%s</a></li>`,
				item.URL, activeClass, item.Label,
			))
		}
		result["nav_items_html"] = strings.Join(htmlParts, "\n                ")
	} else {
		// Ensure empty values don't render as "<no value>"
		result["nav_items"] = []map[string]interface{}{}
		result["nav_items_html"] = ""
	}

	// =========================================================================
	// Generate services list HTML for footer
	// Templates use {{.services_html}} for footer services list
	// ctx.Services is []string (service names extracted from reviewed_brief)
	// =========================================================================
	if len(ctx.Services) > 0 {
		result["services"] = ctx.Services

		// services_html as pre-rendered string
		var servicesParts []string
		for _, serviceName := range ctx.Services {
			if serviceName != "" {
				servicesParts = append(servicesParts, fmt.Sprintf(
					`<li><a href="/services.html">%s</a></li>`,
					serviceName,
				))
			}
		}
		result["services_html"] = strings.Join(servicesParts, "\n                ")
	} else {
		result["services"] = []string{}
		result["services_html"] = ""
	}

	// Add image URLs if present in ContentData
	if ctx.ContentData != nil {
		for _, field := range []string{"hero_url", "hero_home_url", "hero_about_url", "logo_url"} {
			if url, ok := ctx.ContentData[field].(string); ok && url != "" {
				result[field] = url
			}
		}
	}

	// =========================================================================
	// Merge ContentData fields for additional template access
	// This includes full service objects for {{range .services}} iteration
	// =========================================================================
	if ctx.ContentData != nil {
		for key, value := range ctx.ContentData {
			// Don't overwrite explicit fields
			if _, exists := result[key]; !exists {
				result[key] = value
			}
		}
	}

	return result
}

func mergeIntoRenderContext(ctx *RenderContext, data map[string]interface{}) {
	if v, ok := data["domain"].(string); ok && v != "" {
		ctx.Domain = v
	}
	if v, ok := data["company_name"].(string); ok && v != "" {
		ctx.CompanyName = v
		if ctx.LogoText == "" {
			ctx.LogoText = v
		}
	}
	if v, ok := data["logo_text"].(string); ok && v != "" {
		ctx.LogoText = v
	}
	if v, ok := data["tagline"].(string); ok && v != "" {
		ctx.Tagline = v
	}
	if v, ok := data["email"].(string); ok && v != "" {
		ctx.Email = v
	}
	if v, ok := data["phone"].(string); ok && v != "" {
		ctx.Phone = v
	}
	if v, ok := data["primary_color"].(string); ok && v != "" {
		ctx.PrimaryColor = v
	}
	if v, ok := data["secondary_color"].(string); ok && v != "" {
		ctx.SecondaryColor = v
	}
	if v, ok := data["accent_color"].(string); ok && v != "" {
		ctx.AccentColor = v
	}

	// Check nested color_palette
	if palette, ok := data["color_palette"].(map[string]interface{}); ok {
		if v, ok := palette["primary"].(string); ok {
			ctx.PrimaryColor = v
		}
		if v, ok := palette["secondary"].(string); ok {
			ctx.SecondaryColor = v
		}
		if v, ok := palette["accent"].(string); ok {
			ctx.AccentColor = v
		}
	}

	// Capture ALL fields into ContentData for template access
	if ctx.ContentData == nil {
		ctx.ContentData = make(map[string]interface{})
	}
	for key, value := range data {
		ctx.ContentData[key] = value
	}
}

// DEPRECATED
func convertNavigationItems(items []NavigationItem) []NavItem {
	result := make([]NavItem, len(items))
	for i, item := range items {
		result[i] = NavItem{
			Label: item.Label,
			URL:   item.URL,
		}
	}
	return result
}

// RenderComponentAction renders a single component template with context
// Config:
//   - component_function: function name to look up (e.g., "hero-banner")
//   - component_id: explicit component UUID (alternative to function)
//   - component_from: path to object containing function/id (e.g., "current_section")
//   - context_field: path to render context in collected_data
//   - content_field: path to additional content data
//   - content_from: alias for content_field (for consistency)
func RenderComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("RenderComponentAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		return nil, fmt.Errorf("database connection required for component rendering")
	}

	// Get component - now with support for component_from indirection
	var comp *Component
	var err error
	var componentFunction string
	var componentID string

	// Priority 1: Direct component_id in config
	if compID, ok := config["component_id"].(string); ok && compID != "" {
		componentID = compID
	}

	// Priority 2: Direct component_function in config
	if compFunc, ok := config["component_function"].(string); ok && compFunc != "" {
		componentFunction = compFunc
	}

	// Priority 3: Extract from component_from field (indirection)
	if componentID == "" && componentFunction == "" {
		if componentFrom, ok := config["component_from"].(string); ok && componentFrom != "" {
			componentData := datahelpers.ExtractNestedField(params.CollectedData, componentFrom)
			if componentData != nil {
				// Case 1: component_from points to a map (e.g., "current_section")
				if compMap, ok := componentData.(map[string]interface{}); ok {
					// Try to get function first (most common)
					if fn, ok := compMap["function"].(string); ok && fn != "" {
						componentFunction = fn
						params.Logger.Debug("RenderComponentAction: Extracted function from component_from",
							zap.String("component_from", componentFrom),
							zap.String("function", componentFunction),
						)
					}
					// Also check for component_function key
					if fn, ok := compMap["component_function"].(string); ok && fn != "" {
						componentFunction = fn
					}
					// Check for id/component_id
					if id, ok := compMap["id"].(string); ok && id != "" {
						componentID = id
					}
					if id, ok := compMap["component_id"].(string); ok && id != "" {
						componentID = id
					}
					// Check for name as fallback (some components use name as function)
					if componentFunction == "" {
						if name, ok := compMap["name"].(string); ok && name != "" {
							componentFunction = name
							params.Logger.Debug("RenderComponentAction: Using name as function fallback",
								zap.String("name", name),
							)
						}
					}
				} else if compStr, ok := componentData.(string); ok && compStr != "" {
					// Case 2: component_from points directly to a string value
					// e.g., "current_section.function" resolves to "hero"
					componentFunction = compStr
					params.Logger.Debug("RenderComponentAction: Extracted function string directly from component_from",
						zap.String("component_from", componentFrom),
						zap.String("function", componentFunction),
					)
				}
			} else {
				params.Logger.Warn("RenderComponentAction: component_from field not found",
					zap.String("component_from", componentFrom),
				)
			}
		}
	}

	// Enforce naming contract: normalize to kebab-case before lookup
	if componentFunction != "" {
		normalized := NormalizeComponentFunction(componentFunction)
		if normalized != componentFunction {
			params.Logger.Info("RenderComponentAction: Normalized component function to kebab-case",
				zap.String("original", componentFunction),
				zap.String("normalized", normalized),
			)
			componentFunction = normalized
		}
	}

	// Now resolve the component
	if componentID != "" {
		compUUID, parseErr := uuid.Parse(componentID)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid component_id: %w", parseErr)
		}
		comp, err = GetComponentByID(ctx, params.DB, compUUID, params.Logger)
	} else if componentFunction != "" {
		comp, err = GetComponentWithFallback(ctx, params.DB, componentFunction, params.Logger)
	} else {
		// Log available info for debugging
		params.Logger.Error("RenderComponentAction: No component identifier found",
			zap.Any("config_keys", datahelpers.GetMapKeys(config)),
			zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
		)
		return nil, fmt.Errorf("component_function, component_id, or component_from required")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get component '%s': %w", componentFunction, err)
	}

	// Get render context
	contextField := "render_context"
	if cf, ok := config["context_field"].(string); ok && cf != "" {
		contextField = cf
	}
	// Also support context_from as an alias
	if cf, ok := config["context_from"].(string); ok && cf != "" {
		contextField = cf
	}

	renderCtxData := datahelpers.ExtractNestedField(params.CollectedData, contextField)
	renderCtx := &RenderContext{}
	if m, ok := renderCtxData.(map[string]interface{}); ok {
		mergeIntoRenderContext(renderCtx, m)
	}

	// Merge additional content if specified (support both content_field and content_from)
	contentField := ""
	if cf, ok := config["content_field"].(string); ok && cf != "" {
		contentField = cf
	}
	if cf, ok := config["content_from"].(string); ok && cf != "" {
		contentField = cf
	}

	var sectionContentData map[string]interface{}

	if contentField != "" {
		// Try to extract content with fallback paths
		// LLM responses sometimes have .result wrapper, sometimes not
		contentData := extractContentWithFallbacks(params.CollectedData, contentField, params.Logger)
		if contentData != nil {
			sectionContentData = contentData // ← capture before merge
			params.Logger.Info("RenderComponentAction: Merging content data",
				zap.String("content_field", contentField),
				zap.Int("field_count", len(contentData)),
				zap.Any("keys", datahelpers.GetMapKeys(contentData)))
			mergeIntoRenderContext(renderCtx, contentData)
		} else {
			params.Logger.Warn("RenderComponentAction: No content data found at any path",
				zap.String("content_field", contentField),
				zap.Any("available_top_keys", datahelpers.GetMapKeys(params.CollectedData)))
		}
	}

	// Step 3: optional merge_with — overlay pre-resolved data on top of the
	// LLM/content output. Used by page-content-writer's loop with
	// `merge_with: current_section.resolved_data` so query-resolved items,
	// static fallback values, and other authoritative data land in both the
	// rendered HTML AND the persisted content_data. The merge happens AFTER
	// the content_from block so resolved_data wins on conflicts — by design,
	// because it's database-derived and authoritative; the LLM should never
	// be writing items/urls/labels that the resolver already produced.
	if mw, ok := config["merge_with"].(string); ok && mw != "" {
		mergeData := datahelpers.ExtractNestedField(params.CollectedData, mw)
		if mergeMap, ok := mergeData.(map[string]interface{}); ok && len(mergeMap) > 0 {
			params.Logger.Info("RenderComponentAction: Merging resolved data",
				zap.String("merge_with", mw),
				zap.Int("merge_field_count", len(mergeMap)),
				zap.Any("merge_keys", datahelpers.GetMapKeys(mergeMap)))
			if sectionContentData == nil {
				sectionContentData = make(map[string]interface{})
			}
			// Overlay merge data onto section content data so it lands in both
			// the render context AND the persisted content_data output.
			// Last write wins → resolved_data overrides LLM duplicates.
			for k, v := range mergeMap {
				sectionContentData[k] = v
			}
			mergeIntoRenderContext(renderCtx, mergeMap)
		} else if mergeData != nil {
			params.Logger.Warn("RenderComponentAction: merge_with did not resolve to a map",
				zap.String("merge_with", mw),
				zap.String("type", fmt.Sprintf("%T", mergeData)))
		}
		// If mergeData is nil, that's fine — the path simply wasn't populated
		// for this section (e.g. an all-LLM section with no resolved_data).
	}

	// Before calling RenderTemplate, set ComponentID in context
	renderCtx.ContentData["ComponentID"] = comp.ID

	// Render template
	rendered := RenderTemplate(comp.HTMLTemplate, renderCtx, params.Logger)

	// Dark section contract validation (warning only, non-blocking)
	// Uses is_dark_section from DB when available, falls back to CSS auto-detection.
	// See validate_dark_section.go and 014_section_context_contract.md.
	if missing := ValidateDarkSectionContract(rendered, comp.IsDarkSection, params.Logger); len(missing) > 0 {
		params.Logger.Warn("RenderComponentAction: Dark section missing --section-* variables",
			zap.String("component_function", comp.Function),
			zap.Bool("is_dark_section", comp.IsDarkSection),
			zap.Strings("missing_vars", missing),
		)
	}

	params.Logger.Info("RenderComponentAction: Component rendered",
		zap.String("component", comp.Name),
		zap.String("function", comp.Function),
		zap.Int("output_length", len(rendered)),
	)

	result := map[string]interface{}{
		"rendered_html":      rendered,
		"component_id":       comp.ID,
		"component_name":     comp.Name,
		"component_function": comp.Function,
	}
	if sectionContentData != nil {
		result["content_data"] = sectionContentData
	}
	return result, nil
}

// ============================================================================
// ACTION: compile_page_sections
// ============================================================================

// CompilePageSectionsAction combines multiple rendered sections into a page
// Config:
//   - sections_from OR sections_field: path to array of section results
//   - page_from: path to page data for name/title
//   - page_name: explicit name for the page (fallback)
//   - inject_header: boolean
//   - inject_footer: boolean
func CompilePageSectionsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("CompilePageSectionsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Accept both "sections_from" and "sections_field" for compatibility
	sectionsField := "rendered_sections"
	if sf, ok := config["sections_from"].(string); ok && sf != "" {
		sectionsField = sf
	} else if sf, ok := config["sections_field"].(string); ok && sf != "" {
		sectionsField = sf
	}

	params.Logger.Info("CompilePageSectionsAction: Looking for sections",
		zap.String("sections_field", sectionsField))

	sectionsData := datahelpers.ExtractNestedField(params.CollectedData, sectionsField)
	if sectionsData == nil {
		// Try with .results suffix (loop action output format)
		sectionsData = datahelpers.ExtractNestedField(params.CollectedData, sectionsField+".results")
		if sectionsData != nil {
			params.Logger.Info("CompilePageSectionsAction: Found sections at .results path")
		}
	}

	if sectionsData == nil {
		params.Logger.Error("CompilePageSectionsAction: Sections not found",
			zap.String("tried_path", sectionsField),
			zap.String("also_tried", sectionsField+".results"),
			zap.Strings("available_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return nil, fmt.Errorf("sections not found at %s", sectionsField)
	}

	var sections []string
	var sectionsMetadata []map[string]interface{}

	switch v := sectionsData.(type) {
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				sections = append(sections, s)
				// String-only item: no metadata available
				sectionsMetadata = append(sectionsMetadata, map[string]interface{}{
					"rendered_html": s,
				})
			} else if m, ok := item.(map[string]interface{}); ok {
				html, meta := extractSectionFromMap(m, params.Logger)
				if html != "" {
					sections = append(sections, html)
					sectionsMetadata = append(sectionsMetadata, meta)
				}
			}
		}
	case map[string]interface{}:
		// Check if this is a loop output with "results" array
		if results, ok := v["results"].([]interface{}); ok {
			for _, item := range results {
				if m, ok := item.(map[string]interface{}); ok {
					html, meta := extractSectionFromMap(m, params.Logger)
					if html != "" {
						sections = append(sections, html)
						sectionsMetadata = append(sectionsMetadata, meta)
					}
				}
			}
		} else {
			// Ordered by keys (section_0, section_1, etc.)
			for i := 0; i < len(v); i++ {
				key := fmt.Sprintf("section_%d", i)
				if section, ok := v[key]; ok {
					if s, ok := section.(string); ok {
						sections = append(sections, s)
						sectionsMetadata = append(sectionsMetadata, map[string]interface{}{
							"rendered_html": s,
						})
					} else if m, ok := section.(map[string]interface{}); ok {
						html, meta := extractSectionFromMap(m, params.Logger)
						if html != "" {
							sections = append(sections, html)
							sectionsMetadata = append(sectionsMetadata, meta)
						}
					}
				}
			}
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Extracted sections",
		zap.Int("count", len(sections)))

	if len(sections) == 0 {
		params.Logger.Warn("CompilePageSectionsAction: No sections to compile, returning placeholder")

		// Get page info for context
		pageName := "page"
		if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
			if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
				if pm, ok := pageData.(map[string]interface{}); ok {
					if name, ok := pm["name"].(string); ok && name != "" {
						pageName = name
					}
				}
			}
		}

		return map[string]interface{}{
			"page_body":     "",
			"page_name":     pageName,
			"section_count": 0,
			"skipped":       true,
			"reason":        "no sections defined for page",
		}, nil
	}

	// Build page body
	pageBody := strings.Join(sections, "\n\n")

	// Get page name - try page_from first, then page_name config
	pageName := "index"
	if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
		if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
			if pm, ok := pageData.(map[string]interface{}); ok {
				if name, ok := pm["name"].(string); ok && name != "" {
					pageName = name
				}
			}
		}
	}
	if pn, ok := config["page_name"].(string); ok && pn != "" {
		pageName = pn
	}

	// Build full HTML page
	pageHTML := buildPageHTML(pageName, pageBody)

	// Optionally inject head/header/footer from component library.
	// Head is injected here (same time as header/footer) rather than deferred
	// to a later assemble_page step — deferring caused <head> to end up inside
	// <body> when cleanHTMLStructure's dedup logic picked the wrong block.
	injectHead, _ := config["inject_head"].(bool)
	injectHeader, _ := config["inject_header"].(bool)
	injectFooter, _ := config["inject_footer"].(bool)

	if params.DB != nil && (injectHead || injectHeader || injectFooter) {
		// Get site_id for component lookup
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, "site_record.site_id")
		siteUUID := uuid.Nil
		if siteIDStr != "" {
			siteUUID, _ = uuid.Parse(siteIDStr)
		}

		renderCtx := &RenderContext{}
		if rc := datahelpers.ExtractNestedField(params.CollectedData, "render_context"); rc != nil {
			if m, ok := rc.(map[string]interface{}); ok {
				mergeIntoRenderContext(renderCtx, m)
			}
		}

		// Ensure page-specific title/description are set for head component.
		// The head template uses {{.title}} and {{.description}} — these must
		// reflect the current page, not just the site-level defaults.
		if pageFrom, ok := config["page_from"].(string); ok && pageFrom != "" {
			if pageData := datahelpers.ExtractNestedField(params.CollectedData, pageFrom); pageData != nil {
				if pm, ok := pageData.(map[string]interface{}); ok {
					if t, ok := pm["title"].(string); ok && t != "" {
						renderCtx.Title = t
					} else if n, ok := pm["name"].(string); ok && n != "" && renderCtx.Title == "" {
						renderCtx.Title = strings.Title(strings.ReplaceAll(n, "-", " "))
					}
					if d, ok := pm["meta_description"].(string); ok && d != "" {
						renderCtx.Description = d
					}
				}
			}
		}

		if injectHead {
			pageHTML = InjectHead(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
		if injectHeader {
			pageHTML = InjectHeader(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
		if injectFooter {
			pageHTML = InjectFooter(ctx, params.DB, pageHTML, siteUUID, renderCtx, params.Logger)
		}
	}

	params.Logger.Info("CompilePageSectionsAction: Page compiled",
		zap.String("page_name", pageName),
		zap.Int("section_count", len(sections)),
		zap.Int("html_length", len(pageHTML)),
	)

	return map[string]interface{}{
		"page_html":         pageHTML,
		"page_name":         pageName,
		"section_count":     len(sections),
		"sections_metadata": sectionsMetadata,
	}, nil
}

// extractSectionFromMap reads HTML and component metadata from a section result map.
// It handles two shapes:
//
//  1. Direct RenderComponentAction output: metadata lives at top level alongside
//     rendered_html. This is what CompilePageSectionsAction used to assume.
//
//  2. Loop-wrapped output: LoopAction's completion step (Strategy 1,
//     substep_output_fields) promotes HTML to top-level rendered_html/page_html,
//     but component_id/component_name/component_function stay nested inside the
//     substep output key (section_output, render_section, render_from_template).
//     This is what content-writer's process_sections_loop produces in practice.
//
// Lookup order: top-level first, then nested substep keys (earliest-wins). This
// preserves the historical behaviour for callers with the flat shape while
// recovering metadata from the nested shape — without which save_page_sections
// ends up with ComponentName = "section" (the default from extractSectionsFromMetadata)
// and enrichSectionsWithComponentIDs skips every section, leaving page_components.component_id NULL.
//
// Returned meta always contains rendered_html. When available, also:
// component_id, component_name, component_function, content_data.
// Returns ("", nil) if no HTML could be found in any known position.
func extractSectionFromMap(m map[string]interface{}, logger *zap.Logger) (string, map[string]interface{}) {
	// Extract HTML — check top-level first, then common substep keys.
	html := extractHTMLFromSectionMap(m)
	if html == "" {
		return "", nil
	}

	meta := map[string]interface{}{
		"rendered_html": html,
	}

	// Collect component metadata from top level first.
	if id, ok := m["component_id"]; ok && id != nil {
		meta["component_id"] = fmt.Sprintf("%v", id)
	}
	if name, ok := m["component_name"].(string); ok && name != "" {
		meta["component_name"] = name
	}
	if fn, ok := m["component_function"].(string); ok && fn != "" {
		meta["component_function"] = fn
	}
	if cd, ok := m["content_data"]; ok && cd != nil {
		meta["content_data"] = cd
	}

	// Remember whether top-level already had the name, so we only log recovery
	// when the nested fallback actually contributed it.
	_, hadTopName := m["component_name"].(string)

	if meta["component_id"] == nil || meta["component_name"] == nil ||
		meta["component_function"] == nil || meta["content_data"] == nil {

		for _, subKey := range []string{"section_output", "render_section", "render_from_template"} {
			nested, ok := m[subKey].(map[string]interface{})
			if !ok {
				continue
			}
			if meta["component_id"] == nil {
				if id, ok := nested["component_id"]; ok && id != nil {
					meta["component_id"] = fmt.Sprintf("%v", id)
				}
			}
			if meta["component_name"] == nil {
				if name, ok := nested["component_name"].(string); ok && name != "" {
					meta["component_name"] = name
				}
			}
			if meta["component_function"] == nil {
				if fn, ok := nested["component_function"].(string); ok && fn != "" {
					meta["component_function"] = fn
				}
			}
			if meta["content_data"] == nil {
				if cd, ok := nested["content_data"]; ok && cd != nil {
					meta["content_data"] = cd
				}
			}
			if meta["component_id"] != nil && meta["component_name"] != nil &&
				meta["component_function"] != nil && meta["content_data"] != nil {
				break
			}
		}

		// Signal that the nested-lookup fallback fired — the primary diagnostic
		// signal that this fix is taking effect in production logs.
		if !hadTopName {
			if n, ok := meta["component_name"].(string); ok && n != "" {
				logger.Info("CompilePageSectionsAction: recovered component_name from nested substep output",
					zap.String("component_name", n))
			}
		}
	}

	return html, meta
}

// extractHTMLFromSectionMap pulls the rendered HTML string from a section-result
// map, checking top-level keys first, then common substep output keys where
// LoopAction may have nested the RenderComponentAction result.
func extractHTMLFromSectionMap(m map[string]interface{}) string {
	if h, ok := m["rendered_html"].(string); ok && h != "" {
		return h
	}
	if h, ok := m["page_html"].(string); ok && h != "" {
		return h
	}
	if h, ok := m["html"].(string); ok && h != "" {
		return h
	}
	// Also try nested substep keys (loop-wrapped shape where top-level HTML
	// promotion didn't happen for some reason).
	for _, subKey := range []string{"section_output", "render_section", "render_from_template"} {
		if nested, ok := m[subKey].(map[string]interface{}); ok {
			if h, ok := nested["rendered_html"].(string); ok && h != "" {
				return h
			}
			if h, ok := nested["page_html"].(string); ok && h != "" {
				return h
			}
		}
	}
	return ""
}

func buildPageHTML(pageName, body string) string {
	title := strings.Title(strings.ReplaceAll(pageName, "-", " "))
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body>
%s
</body>
</html>`, title, body)
}

// ============================================================================
// ACTION: insert_research_result
// ============================================================================

// InsertResearchResultAction stores research findings in the database
// Config:
//   - table: target table name (default: "research_results")
//   - fields: map of column_name -> data_path for dynamic field mapping
//   - site_id_field: path to site_id (fallback if not in fields)
//   - result_type: type of research (fallback if not in fields)
func InsertResearchResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("InsertResearchResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("InsertResearchResultAction: No database")
		return map[string]interface{}{"inserted": false, "reason": "no database"}, nil
	}

	// Get table name (default to research_results)
	tableName := "research_results"
	if tn, ok := config["table"].(string); ok && tn != "" {
		tableName = tn
	}

	// Get field mappings from config
	fieldMappings, hasFields := config["fields"].(map[string]interface{})

	// Generate new ID
	resultID := uuid.New()

	// Build dynamic INSERT based on field mappings
	if hasFields && len(fieldMappings) > 0 {
		// Dynamic field-based insert
		columns := []string{"id"}
		placeholders := []string{"$1"}
		values := []interface{}{resultID}
		paramIdx := 2

		for column, dataPath := range fieldMappings {
			dataPathStr, ok := dataPath.(string)
			if !ok {
				continue
			}

			// Extract the value from collected data
			value := datahelpers.ExtractNestedField(params.CollectedData, dataPathStr)

			// Special handling for site_id (needs UUID conversion)
			if column == "site_id" {
				if siteIDStr, ok := value.(string); ok && siteIDStr != "" {
					if siteUUID, err := uuid.Parse(siteIDStr); err == nil {
						value = siteUUID
					} else {
						value = nil
					}
				} else {
					value = nil
				}
			}

			// Skip nil values for optional fields, but include empty strings
			if value == nil {
				continue
			}

			// For complex types, marshal to JSON
			/*switch v := value.(type) {
			case map[string]interface{}, []interface{}:
				jsonBytes, err := json.Marshal(v)
				if err != nil {
					params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
						zap.String("column", column),
						zap.Error(err))
					continue
				}
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
				values = append(values, string(jsonBytes))
			default:
				columns = append(columns, column)
				placeholders = append(placeholders, fmt.Sprintf("$%d", paramIdx))
				values = append(values, value)
			}*/

			// make it all json
			jsonBytes, err := json.Marshal(value)
			if err != nil {
				params.Logger.Warn("InsertResearchResultAction: Failed to marshal field",
					zap.String("column", column),
					zap.String("value_type", fmt.Sprintf("%T", value)),
					zap.Error(err))
				continue
			}
			columns = append(columns, column)
			placeholders = append(placeholders, fmt.Sprintf("$%d::jsonb", paramIdx))
			values = append(values, string(jsonBytes))
			paramIdx++
		}

		// Add created_at
		columns = append(columns, "created_at")
		placeholders = append(placeholders, "NOW()")

		query := fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			tableName,
			strings.Join(columns, ", "),
			strings.Join(placeholders, ", "),
		)

		params.Logger.Info("InsertResearchResultAction: Executing dynamic insert",
			zap.String("table", tableName),
			zap.Strings("columns", columns),
			zap.Int("value_count", len(values)))

		if err := execDB(ctx, params.DB, query, values...); err != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed",
				zap.String("query", query),
				zap.Error(err))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": "general",
				"error":       err.Error(),
			}, nil
		}

		return map[string]interface{}{
			"inserted":  true,
			"id":        resultID.String(),
			"result_id": resultID.String(),
			"table":     tableName,
			"columns":   columns,
		}, nil
	}

	// Fallback: Legacy mode with hardcoded columns
	// Get site_id
	siteIDField := "site_record.site_id"
	if f, ok := config["site_id_field"].(string); ok && f != "" {
		siteIDField = f
	}
	siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
	siteID := uuid.Nil
	if siteIDStr != "" {
		siteID, _ = uuid.Parse(siteIDStr)
	}

	// Get result type
	resultType := "general"
	if rt, ok := config["result_type"].(string); ok && rt != "" {
		resultType = rt
	}

	// Get research data - try multiple field paths
	dataField := "research_result"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	researchData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// If no data found, try to build from synthesis
	if researchData == nil {
		researchData = map[string]interface{}{
			"summary":  datahelpers.ExtractNestedField(params.CollectedData, "synthesis.summary"),
			"findings": datahelpers.ExtractNestedField(params.CollectedData, "synthesis"),
			"query":    datahelpers.ExtractNestedField(params.CollectedData, "search_query.result"),
			"topic":    datahelpers.ExtractNestedField(params.CollectedData, "extracted.topic"),
		}
	}

	dataJSON, err := json.Marshal(researchData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal research data: %w", err)
	}

	// Try insert with 'findings' column first (new schema), fall back to 'data' (old schema)
	var siteIDArg interface{} = nil
	if siteID != uuid.Nil {
		siteIDArg = siteID
	}

	// Try new schema first
	query := `
		INSERT INTO research_results (id, site_id, result_type, findings, created_at)
		VALUES ($1, $2, $3, $4::jsonb, NOW())
	`

	if err := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err != nil {
		// Try with 'data' column (old schema)
		query = `
			INSERT INTO research_results (id, site_id, result_type, data, created_at)
			VALUES ($1, $2, $3, $4::jsonb, NOW())
		`
		if err2 := execDB(ctx, params.DB, query, resultID, siteIDArg, resultType, string(dataJSON)); err2 != nil {
			params.Logger.Warn("InsertResearchResultAction: Insert failed with both schemas",
				zap.Error(err),
				zap.Error(err2))
			return map[string]interface{}{
				"inserted":    false,
				"result_type": resultType,
				"error":       err2.Error(),
			}, nil
		}
	}

	return map[string]interface{}{
		"inserted":    true,
		"result_id":   resultID.String(),
		"result_type": resultType,
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// ACTION: store_asset
// ============================================================================

// StoreAssetAction stores an asset (image, font, file) in the assets table
// This is a LOCAL action - it does NOT require a topic
// Config:
//   - asset_type: type of asset (image, font, css, js, etc.)
//   - site_id_field: path to site_id in collected_data
//   - data_field: path to asset data (URL, base64, or content)
//   - name_field: path to asset name
//   - metadata_field: optional path to additional metadata
func StoreAssetAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("StoreAssetAction: Starting",
		zap.Any("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)),
	)

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get asset type
	assetType := "image"
	if at, ok := config["asset_type"].(string); ok && at != "" {
		assetType = at
	}

	// Get purpose from config — supports literal OR field lookup.
	// Phase 2H: purpose_field added so the new needs_imagery branch can pass
	// spec.purpose through dynamically (logo, hero, illustration, icon,
	// infographic) without hardcoding in workflow step config. Mirrors the
	// asset_key / asset_key_field pattern below.
	//
	// Resolution priority:
	//   1. config["purpose"]        — literal string (e.g. "hero", "logo")
	//   2. config["purpose_field"]  — JSONPath into collected_data
	//   3. ""                       — empty; downstream asset_key resolution
	//                                 may still backfill via asset_key_field
	purpose := ""
	if p, ok := config["purpose"].(string); ok && p != "" {
		purpose = p
	}
	if purpose == "" {
		if pf, ok := config["purpose_field"].(string); ok && pf != "" {
			purpose = datahelpers.ExtractNestedFieldString(params.CollectedData, pf)
		}
	}

	// Phase 2C: extract asset_key from config, defaulting to purpose.
	// Phase 2E: also support asset_key_field for JSONPath lookup so
	// per-item variants can be passed through the workflow without
	// hardcoded literals.
	//
	// Resolution priority:
	//   1. config["asset_key"]        — literal string (e.g. "logo")
	//   2. config["asset_key_field"]  — JSONPath into collected_data
	//   3. purpose (default — Phase 2C backward-compat)
	assetKey := ""
	if k, ok := config["asset_key"].(string); ok && k != "" {
		assetKey = k
	}
	if assetKey == "" {
		if kf, ok := config["asset_key_field"].(string); ok && kf != "" {
			assetKey = datahelpers.ExtractNestedFieldString(params.CollectedData, kf)
		}
	}
	if assetKey == "" {
		assetKey = purpose
	}

	// Get site_id (optional - assets can be global)
	var siteID *uuid.UUID
	if siteIDField, ok := config["site_id_field"].(string); ok && siteIDField != "" {
		siteIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, siteIDField)
		if siteIDStr != "" {
			if parsed, err := uuid.Parse(siteIDStr); err == nil {
				siteID = &parsed
			}
		}
	}

	// Get asset data
	dataField := "asset_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	assetData := datahelpers.ExtractNestedField(params.CollectedData, dataField)

	// Get asset name
	nameField := "asset_name"
	if nf, ok := config["name_field"].(string); ok && nf != "" {
		nameField = nf
	}
	assetName := datahelpers.ExtractNestedFieldString(params.CollectedData, nameField)
	if assetName == "" {
		assetName = fmt.Sprintf("%s_%s", assetType, uuid.New().String()[:8])
	}

	// Extract URL or content from asset data
	var assetURL string
	switch v := assetData.(type) {
	case string:
		if strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "s3://") {
			assetURL = v
		}
	case map[string]interface{}:
		if url, ok := v["url"].(string); ok {
			assetURL = url
		}
		if url, ok := v["image_url"].(string); ok {
			assetURL = url
		}
	}

	if assetURL == "" {
		params.Logger.Warn("StoreAssetAction: No asset URL found")
		return map[string]interface{}{
			"stored":     false,
			"asset_name": assetName,
			"asset_type": assetType,
			"reason":     "no asset URL found",
		}, nil
	}

	// If no DB, return success without persisting
	if params.DB == nil {
		params.Logger.Warn("StoreAssetAction: No database, returning without persistence")
		return map[string]interface{}{
			"stored":     true,
			"persisted":  false,
			"asset_id":   uuid.New().String(),
			"asset_name": assetName,
			"asset_type": assetType,
			"asset_url":  assetURL,
		}, nil
	}

	// Determine origin_type based on URL
	originType := "uploaded"
	if strings.HasPrefix(assetURL, "s3://") || strings.Contains(assetURL, "backblazeb2.com") {
		originType = "generated"
	}

	// Phase 0.2: extract origin_prompt and origin_model from step config.
	// Bug fix — origin_prompt_field has been silently passed by workflows
	// without this action ever reading it. After this lands, new generations
	// populate the column. New — origin_model accepts a literal (used today,
	// e.g. "sdxl") or a path (origin_model_field) for future provider routing.
	// Literal wins if both are set.
	originPrompt := ""
	if pf, ok := config["origin_prompt_field"].(string); ok && pf != "" {
		originPrompt = datahelpers.ExtractNestedFieldString(params.CollectedData, pf)
	}
	originModel := ""
	if m, ok := config["origin_model"].(string); ok && m != "" {
		originModel = m
	} else if mf, ok := config["origin_model_field"].(string); ok && mf != "" {
		originModel = datahelpers.ExtractNestedFieldString(params.CollectedData, mf)
	}

	// Insert into assets table - matches actual schema
	assetID := uuid.New()

	query := `
		INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url, origin_type,
		                    origin_prompt, origin_model, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		ON CONFLICT (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active' DO UPDATE SET
			purpose = EXCLUDED.purpose,
			url = EXCLUDED.url,
			name = EXCLUDED.name,
			origin_type = EXCLUDED.origin_type,
			origin_prompt = COALESCE(EXCLUDED.origin_prompt, assets.origin_prompt),
			origin_model = COALESCE(EXCLUDED.origin_model, assets.origin_model),
			updated_at = NOW()
		RETURNING id
	`

	var returnedID uuid.UUID
	err := queryRowScanUUID(ctx, params.DB, query, &returnedID,
		assetID, siteID, assetName, assetType, nullString(purpose),
		nullString(assetKey), assetURL, originType,
		nullString(originPrompt), nullString(originModel))

	if err != nil {
		// Try simpler insert without upsert if constraint doesn't exist
		params.Logger.Warn("StoreAssetAction: Upsert failed, trying simple insert",
			zap.Error(err))

		simpleQuery := `
			INSERT INTO assets (id, site_id, name, asset_type, purpose, asset_key, url, origin_type,
			                    origin_prompt, origin_model, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
			RETURNING id
		`
		err = queryRowScanUUID(ctx, params.DB, simpleQuery, &returnedID,
			assetID, siteID, assetName, assetType, nullString(purpose),
			nullString(assetKey), assetURL, originType,
			nullString(originPrompt), nullString(originModel))

		if err != nil {
			params.Logger.Warn("StoreAssetAction: Insert failed",
				zap.Error(err))
			return map[string]interface{}{
				"stored":     true,
				"persisted":  false,
				"asset_id":   assetID.String(),
				"asset_name": assetName,
				"asset_type": assetType,
				"asset_url":  assetURL,
				"error":      err.Error(),
			}, nil
		}
	}

	// If purpose is set and we have a site_id, update sites.content_data
	// This stores the storage URI (for download) and relative URL (for templates)
	storageURI := ""
	if purpose != "" && siteID != nil {
		// Find storage URI from asset data
		if assetDataMap, ok := assetData.(map[string]interface{}); ok {
			if uri, ok := assetDataMap["image_uri"].(string); ok {
				storageURI = uri
			}
		}
		// Also check collected_data for {purpose}_result.image_uri pattern
		if storageURI == "" {
			uriField := strings.TrimSuffix(dataField, ".image_url") + ".image_uri"
			if uri := datahelpers.ExtractNestedFieldString(params.CollectedData, uriField); uri != "" {
				storageURI = uri
			}
		}
		// Use assetURL if it's a storage URI
		if storageURI == "" && storage.IsS3URI(assetURL) {
			storageURI = assetURL
		}

		// Generate paths using storage package helper (use correct extension for purpose)
		_, _, _, purposeExt := storage.GetImageConfig(purpose)
		paths := storage.BuildAssetPaths(purpose, purposeExt)

		// Update sites.content_data
		if storageURI != "" {
			// Store URI for deploy_image_asset to download from
			updateContentDataField(ctx, params.DB, *siteID, purpose+"_uri", storageURI, params.Logger)
			params.CollectedData[purpose+"_uri"] = storageURI
		}

		// Store relative URL for templates
		updateContentDataField(ctx, params.DB, *siteID, purpose+"_url", paths.RelativeURL, params.Logger)
		params.CollectedData[purpose+"_url"] = paths.RelativeURL

		params.Logger.Info("StoreAssetAction: Updated content_data for purpose",
			zap.String("purpose", purpose),
			zap.String("storage_uri", storageURI),
			zap.String("relative_url", paths.RelativeURL))
	}

	params.Logger.Info("StoreAssetAction: Asset stored",
		zap.String("asset_id", returnedID.String()),
		zap.String("asset_name", assetName),
		zap.String("asset_type", assetType),
	)

	result := map[string]interface{}{
		"stored":     true,
		"persisted":  true,
		"asset_id":   returnedID.String(),
		"asset_name": assetName,
		"asset_type": assetType,
		"asset_url":  assetURL,
	}

	// Add purpose-specific fields if set
	if purpose != "" {
		result["purpose"] = purpose
		_, _, _, purposeExt := storage.GetImageConfig(purpose)
		paths := storage.BuildAssetPaths(purpose, purposeExt)
		result[purpose+"_url"] = paths.RelativeURL
	}

	// Add storage URI to result for downstream deploy step
	if storageURI != "" {
		result["image_uri"] = storageURI
		result["s3_uri"] = storageURI
	}

	return result, nil
}

// updateContentDataField updates a single field in sites.content_data
func updateContentDataField(ctx context.Context, db interface{}, siteID uuid.UUID, field, value string, logger *zap.Logger) {
	query := `
        UPDATE sites 
        SET content_data = jsonb_set(
            COALESCE(content_data, '{}'::jsonb),
            $2::text[],
            to_jsonb($3::text),
            true
        ),
        updated_at = NOW()
        WHERE id = $1
    `
	jsonPath := fmt.Sprintf("{%s}", field)

	if err := execDB(ctx, db, query, siteID, jsonPath, value); err != nil {
		logger.Warn("Failed to update content_data field",
			zap.String("field", field),
			zap.Error(err))
	} else {
		logger.Debug("Updated content_data field",
			zap.String("field", field),
			zap.String("value", value))
	}
}

func ValidateSitePlanAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ValidateSitePlanAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	planField := "llm_plan"
	if pf, ok := config["plan_field"].(string); ok && pf != "" {
		planField = pf
	}

	// Extract the plan data
	planData := datahelpers.ExtractNestedField(params.CollectedData, planField)
	if planData == nil {
		return nil, fmt.Errorf("plan not found at '%s'", planField)
	}

	planData = datahelpers.UnwrapDeep(planData, params.Logger)

	params.Logger.Info("ValidateSitePlanAction: After UnwrapDeep",
		zap.String("planData_type", fmt.Sprintf("%T", planData)))

	var plan map[string]interface{}
	switch v := planData.(type) {
	case map[string]interface{}:
		plan = v
	case string:
		if v == "" {
			return nil, fmt.Errorf("plan is empty string - LLM may have returned no content. Check template rendering logs for <no value> placeholders")
		}
		cleaned := strings.TrimSpace(v)
		cleaned = strings.TrimPrefix(cleaned, "```json")
		cleaned = strings.TrimPrefix(cleaned, "```")
		cleaned = strings.TrimSuffix(cleaned, "```")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			return nil, fmt.Errorf("plan is empty after cleaning markdown")
		}
		if err := json.Unmarshal([]byte(cleaned), &plan); err != nil {
			// Include content preview in error for debugging
			preview := cleaned
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return nil, fmt.Errorf("failed to parse plan JSON: %w (preview: %s)", err, preview)
		}
	default:
		return nil, fmt.Errorf("plan must be object or JSON string, got %T", planData)
	}

	pagesRaw, ok := plan["pages"]
	if !ok {
		// Log available keys for debugging
		keys := make([]string, 0, len(plan))
		for k := range plan {
			keys = append(keys, k)
		}
		params.Logger.Error("ValidateSitePlanAction: plan missing 'pages'",
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("plan must have 'pages' array (available keys: %v)", keys)
	}

	pages, ok := pagesRaw.([]interface{})
	if !ok || len(pages) == 0 {
		return nil, fmt.Errorf("pages must be non-empty array")
	}

	// ── Deterministic convergence with adoption-locked pages ────────────────
	// existing_pages is loaded by the load_existing_pages workflow step and
	// carries an adoption_locked flag per page. reconcilePlanWithRealised
	// force-preserves only the locked subset; it is a no-op once the adoption
	// lock has expired (or for from-scratch builds). See
	// FOCUS_adoption_faithfulness_via_locks.md.
	existingField := "existing_pages"
	if ef, ok := config["existing_pages_field"].(string); ok && ef != "" {
		existingField = ef
	}
	var existingPages []interface{}
	if ev := datahelpers.ExtractNestedField(params.CollectedData, existingField); ev != nil {
		switch vv := ev.(type) {
		case []interface{}:
			existingPages = vv
		case []map[string]interface{}:
			// query_database (output_format=array) returns []map[string]interface{},
			// which does NOT satisfy a []interface{} assertion in Go. Convert it so
			// the convergence actually sees the realised pages. Without this the
			// assertion silently fails, existingPages stays empty, and
			// reconcilePlanWithRealised no-ops for every site (adopted pages never
			// preserved, planner siblings never dropped).
			existingPages = make([]interface{}, len(vv))
			for i := range vv {
				existingPages[i] = vv[i]
			}
		}
	}
	// Surface the convergence input size so an empty set is never silent again.
	params.Logger.Info("ValidateSitePlanAction: existing pages loaded for convergence",
		zap.Int("existing_pages", len(existingPages)),
		zap.String("existing_pages_field", existingField))
	var unionedIn, droppedCollision, snappedRename int
	pages, unionedIn, droppedCollision, snappedRename =
		reconcilePlanWithRealised(pages, existingPages, params.Logger)
	plan["pages"] = pages
	params.Logger.Info("ValidateSitePlanAction: reconciled with adoption-locked pages",
		zap.Int("unioned_in", unionedIn),
		zap.Int("dropped_collision", droppedCollision),
		zap.Int("snapped_rename", snappedRename),
		zap.Int("pages_after", len(pages)))

	// ── Truncate, preserving adoption-locked pages ──────────────────────────
	maxPages := 20
	if mp, ok := config["max_pages"].(float64); ok {
		maxPages = int(mp)
	}
	if len(pages) > maxPages {
		// Build the must-keep set: only the adoption-locked existing pages.
		var lockedOnly []interface{}
		for _, rp := range existingPages {
			if rm, ok := rp.(map[string]interface{}); ok {
				if locked, _ := rm["adoption_locked"].(bool); locked {
					lockedOnly = append(lockedOnly, rp)
				}
			}
		}
		pages = truncatePreservingRealised(pages, lockedOnly, maxPages, params.Logger)
		plan["pages"] = pages
	}

	pageNames := make(map[string]bool)
	for _, p := range pages {
		if pm, ok := p.(map[string]interface{}); ok {
			if name, ok := pm["name"].(string); ok {
				pageNames[name] = true
			}
		}
	}

	if ensurePages, ok := config["ensure_pages"].([]interface{}); ok {
		for _, req := range ensurePages {
			if reqName, ok := req.(string); ok && !pageNames[reqName] {
				pages = append(pages, map[string]interface{}{
					"name": reqName, "title": strings.Title(reqName),
					"nav_label": strings.Title(reqName), "nav_order": len(pages) + 1,
					"in_header": true, "in_footer": true, "sections": []interface{}{},
				})
			}
		}
		plan["pages"] = pages
	}

	if _, ok := plan["style_collection"].(string); !ok {
		if ds, ok := config["default_style"].(string); ok {
			plan["style_collection"] = ds
		}
	}

	if plan["needs_logo"] == nil {
		plan["needs_logo"] = false
	}
	if plan["needs_images"] == nil {
		plan["needs_images"] = false
	}
	if plan["image_prompts"] == nil {
		plan["image_prompts"] = map[string]interface{}{}
	}

	// ── Strip site-chrome components from page sections ──────────────────
	// The LLM sometimes includes header/footer components in page sections
	// arrays (e.g. "header-bold-gradient", "footer-standard"). These are
	// site-level components injected during assembly — not page content.
	// If left in, plan_sections creates bogus HITL items for them.
	if params.DB != nil {
		siteChrome := loadSiteChromeNames(ctx, params.DB, params.Logger)
		if len(siteChrome) > 0 {
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				var filtered []interface{}
				for _, s := range sectionsRaw {
					name, ok := s.(string)
					if !ok {
						filtered = append(filtered, s)
						continue
					}
					if siteChrome[name] {
						params.Logger.Info("ValidateSitePlanAction: stripped site-chrome component from page sections",
							zap.Any("page", pm["name"]),
							zap.String("component", name))
					} else {
						filtered = append(filtered, s)
					}
				}
				pm["sections"] = filtered
			}
		}
	}

	// ── Resolve section names to canonical component functions ───────────
	// Implements config flag `validate_components`. Each section name must
	// map to a real content_components.function. Display names ("FAQ
	// Section"), wrong case, and underscore variants are resolved;
	// unresolvable names are dropped + logged. This does NOT deduplicate or
	// make content-intent decisions — it only guarantees every surviving
	// section name is a valid component function.
	validateComponents := false
	if vc, ok := config["validate_components"].(bool); ok {
		validateComponents = vc
	}
	if validateComponents && params.DB != nil {
		resolver := loadComponentNameResolver(ctx, params.DB, params.Logger)
		if len(resolver.validFunctions) > 0 { // only act if components actually loaded
			for _, p := range pages {
				pm, ok := p.(map[string]interface{})
				if !ok {
					continue
				}
				sectionsRaw, ok := pm["sections"].([]interface{})
				if !ok {
					continue
				}
				resolved := make([]interface{}, 0, len(sectionsRaw))
				for _, s := range sectionsRaw {
					name, ok := s.(string)
					if !ok {
						resolved = append(resolved, s) // brief objects pass through
						continue
					}
					fn, ok := resolver.resolve(name)
					if !ok {
						params.Logger.Warn("ValidateSitePlanAction: dropped unresolvable section name",
							zap.Any("page", pm["name"]),
							zap.String("section", name))
						continue
					}
					if fn != name {
						params.Logger.Info("ValidateSitePlanAction: resolved section name to function",
							zap.Any("page", pm["name"]),
							zap.String("from", name),
							zap.String("to", fn))
					}
					resolved = append(resolved, fn)
				}
				pm["sections"] = resolved
			}
		} else {
			params.Logger.Warn("ValidateSitePlanAction: validate_components set but no components loaded — skipping name resolution")
		}
	}

	params.Logger.Info("ValidateSitePlanAction: Complete", zap.Int("pages", len(pages)))
	return plan, nil
}

// nullString returns nil for empty strings, otherwise the string pointer
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// loadSiteChromeNames returns a set of component names that are site-level
// chrome (headers, footers, head) — not page content sections.
func loadSiteChromeNames(ctx context.Context, db *sql.DB, logger *zap.Logger) map[string]bool {
	result := make(map[string]bool)
	rows, err := db.QueryContext(ctx,
		`SELECT name FROM content_components WHERE component_level = 'site' AND is_active = true`)
	if err != nil {
		logger.Warn("loadSiteChromeNames: query failed", zap.Error(err))
		return result
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if rows.Scan(&name) == nil {
			result[name] = true
		}
	}
	return result
}

// componentNameResolver resolves plan section names to canonical
// content_components.function values. Used to implement the
// validate_components flag and to normalise gap-planner section names.
type componentNameResolver struct {
	validFunctions map[string]bool   // function -> true
	displayToFunc  map[string]string // lower(display_name) -> function
	nameToFunc     map[string]string // lower(name) -> function
}

// loadComponentNameResolver loads section/element component identity so
// plan section names can be resolved to a canonical function. Returns an
// empty (non-nil) resolver on error so callers can no-op safely.
func loadComponentNameResolver(ctx context.Context, db *sql.DB, logger *zap.Logger) *componentNameResolver {
	r := &componentNameResolver{
		validFunctions: make(map[string]bool),
		displayToFunc:  make(map[string]string),
		nameToFunc:     make(map[string]string),
	}
	if db == nil {
		return r
	}
	rows, err := db.QueryContext(ctx,
		`SELECT "function", name, COALESCE(display_name, '')
		   FROM content_components
		  WHERE component_level IN ('section','element')
		    AND is_active = true
		    AND "function" <> ''`)
	if err != nil {
		logger.Warn("loadComponentNameResolver: query failed", zap.Error(err))
		return r
	}
	defer rows.Close()
	for rows.Next() {
		var fn, name, display string
		if err := rows.Scan(&fn, &name, &display); err != nil {
			continue
		}
		r.validFunctions[fn] = true
		if name != "" {
			r.nameToFunc[strings.ToLower(name)] = fn
		}
		if display != "" {
			r.displayToFunc[strings.ToLower(display)] = fn
		}
	}
	return r
}

// resolve maps a raw section name to a canonical component function.
// Returns (function, true) if resolved, ("", false) if not. It does NOT
// deduplicate or make content-intent decisions — only name resolution.
func (r *componentNameResolver) resolve(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	// 1. Already a valid function.
	if r.validFunctions[raw] {
		return raw, true
	}
	// 2. Normalise (underscore->hyphen, camelCase->kebab) and re-check.
	norm := NormalizeComponentFunction(raw)
	if norm != raw && r.validFunctions[norm] {
		return norm, true
	}
	// 3. Display-name lookup (handles "FAQ Section" -> "faq").
	if fn, ok := r.displayToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 4. Component name lookup (row name differing from function).
	if fn, ok := r.nameToFunc[strings.ToLower(raw)]; ok {
		return fn, true
	}
	// 5. Display lookup on the normalised form.
	if fn, ok := r.displayToFunc[strings.ToLower(norm)]; ok {
		return fn, true
	}
	return "", false
}

// ============================================================================
// ACTION: db_sync
// ============================================================================

// DBSyncAction is a general-purpose database sync action
// Can be used to insert/update records in any table
// Config:
//   - operation: insert, update, upsert
//   - table: table name
//   - data_field: path to data to sync
//   - key_fields: array of fields that form the primary/unique key
func DBSyncAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("DBSyncAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	if params.DB == nil {
		params.Logger.Warn("DBSyncAction: No database connection")
		return map[string]interface{}{"synced": false, "reason": "no database"}, nil
	}

	operation, _ := config["operation"].(string)
	if operation == "" {
		operation = "upsert"
	}

	table, ok := config["table"].(string)
	if !ok || table == "" {
		return nil, fmt.Errorf("table is required")
	}

	// Get data to sync
	dataField := "sync_data"
	if df, ok := config["data_field"].(string); ok && df != "" {
		dataField = df
	}
	syncData := datahelpers.ExtractNestedField(params.CollectedData, dataField)
	if syncData == nil {
		return nil, fmt.Errorf("no data found at %s", dataField)
	}

	dataMap, ok := syncData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("sync data must be a map")
	}

	// Build and execute query based on operation
	// This is a simplified implementation - in production you'd want more robust SQL building

	params.Logger.Info("DBSyncAction: Syncing data",
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Int("field_count", len(dataMap)),
	)

	// For now, just marshal and log - actual implementation would build SQL
	dataJSON, _ := json.Marshal(dataMap)

	return map[string]interface{}{
		"synced":      true,
		"operation":   operation,
		"table":       table,
		"field_count": len(dataMap),
		"data_size":   len(dataJSON),
	}, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

func execDB(ctx context.Context, db interface{}, query string, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		_, err := d.ExecContext(ctx, query, args...)
		return err
	case *pgxpool.Pool:
		_, err := d.Exec(ctx, query, args...)
		return err
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

func lookupPageID(ctx context.Context, db interface{}, siteID uuid.UUID, pageName string, logger *zap.Logger) (uuid.UUID, error) {
	query := `SELECT id FROM pages WHERE site_id = $1 AND name = $2`
	var pageID uuid.UUID

	switch d := db.(type) {
	case *sql.DB:
		err := d.QueryRowContext(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	case *pgxpool.Pool:
		err := d.QueryRow(ctx, query, siteID, pageName).Scan(&pageID)
		return pageID, err
	default:
		return uuid.Nil, fmt.Errorf("unsupported database type: %T", db)
	}
}

// queryRowScanUUID executes a query and scans the result into a UUID
func queryRowScanUUID(ctx context.Context, db interface{}, query string, dest *uuid.UUID, args ...interface{}) error {
	switch d := db.(type) {
	case *sql.DB:
		return d.QueryRowContext(ctx, query, args...).Scan(dest)
	case *pgxpool.Pool:
		return d.QueryRow(ctx, query, args...).Scan(dest)
	default:
		return fmt.Errorf("unsupported database type: %T", db)
	}
}

// getStyleCollectionByID looks up a style collection by UUID
// This is a local helper since component_library.go doesn't have this function yet
func getStyleCollectionByID(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*StyleCollection, error) {
	query := `
		SELECT id, name, display_name, category, color_palette, typography,
		       header_component_id, footer_component_id, css_theme_id
		FROM style_collections
		WHERE id = $1 AND is_active = true
	`

	var coll StyleCollection
	var colorPaletteJSON, typographyJSON []byte
	var headerID, footerID, cssThemeID *uuid.UUID

	var err error
	switch d := db.(type) {
	case *sql.DB:
		err = d.QueryRowContext(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	case *pgxpool.Pool:
		err = d.QueryRow(ctx, query, id).Scan(
			&coll.ID, &coll.Name, &coll.DisplayName, &coll.Category,
			&colorPaletteJSON, &typographyJSON,
			&headerID, &footerID, &cssThemeID,
		)
	default:
		return nil, fmt.Errorf("unsupported database type: %T", db)
	}

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if colorPaletteJSON != nil {
		json.Unmarshal(colorPaletteJSON, &coll.ColorPalette)
	}
	if typographyJSON != nil {
		json.Unmarshal(typographyJSON, &coll.Typography)
	}
	coll.HeaderComponentID = headerID
	coll.FooterComponentID = footerID
	coll.CSSThemeID = cssThemeID

	return &coll, nil
}

func BuildReviewResultAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("BuildReviewResultAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := map[string]interface{}{
		"reviewed_at": time.Now().UTC().Format(time.RFC3339),
		"review_mode": "unknown",
		"approved":    false,
		"issues":      []interface{}{},
		"edits":       map[string]interface{}{},
	}

	if approved, ok := config["approved"].(bool); ok {
		result["approved"] = approved
	} else if field, ok := config["approved_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			if b, ok := val.(bool); ok {
				result["approved"] = b
			}
		}
	}

	if reviewer, ok := config["reviewer"].(string); ok && reviewer != "" {
		result["reviewed_by"] = reviewer
	} else if field, ok := config["reviewer_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			result["reviewed_by"] = val
		}
	}
	if result["reviewed_by"] == nil {
		result["reviewed_by"] = "system"
	}

	if mode, ok := config["review_mode"].(string); ok {
		result["review_mode"] = mode
	}

	if field, ok := config["eval_score"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["eval_score"] = val
		}
	}

	if field, ok := config["edits_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["edits"] = val
		}
	}

	if field, ok := config["auto_eval_issues"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedField(params.CollectedData, field); val != nil {
			result["auto_eval_issues"] = val
			if issues, ok := val.([]interface{}); ok {
				result["issues"] = issues
			}
		}
	}

	if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
		result["content"] = content
	}

	params.Logger.Info("BuildReviewResultAction: Complete",
		zap.Bool("approved", result["approved"].(bool)),
		zap.String("mode", result["review_mode"].(string)))
	return result, nil
}

// ============================================================================
// ACTION: prepare_review_data
// ============================================================================

func PrepareReviewDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("PrepareReviewDataAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	reviewData := map[string]interface{}{
		"prepared_at": time.Now().UTC().Format(time.RFC3339),
		"fields":      map[string]interface{}{},
	}

	if includeFields, ok := config["include_fields"].([]interface{}); ok {
		fieldsMap := reviewData["fields"].(map[string]interface{})
		for _, field := range includeFields {
			if fieldName, ok := field.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, fieldName); value != nil {
					fieldsMap[fieldName] = value
				}
			}
		}
	}

	if formatForDisplay, ok := config["format_for_display"].(bool); ok && formatForDisplay {
		if page := datahelpers.ExtractNestedField(params.CollectedData, "current_page"); page != nil {
			if pm, ok := page.(map[string]interface{}); ok {
				reviewData["page_name"] = pm["name"]
				reviewData["page_title"] = pm["title"]
			}
		}
		if content := datahelpers.ExtractNestedField(params.CollectedData, "page_content"); content != nil {
			reviewData["content"] = content
		}
		if brief := datahelpers.ExtractNestedField(params.CollectedData, "reviewed_brief"); brief != nil {
			if bm, ok := brief.(map[string]interface{}); ok {
				reviewData["company_name"] = bm["company_name"]
				reviewData["tone"] = bm["tone"]
			}
		}
	}

	params.Logger.Info("PrepareReviewDataAction: Complete")
	return reviewData, nil
}

// ============================================================================
// ACTION: update_page_components_status
// ============================================================================

func UpdatePageComponentsStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdatePageComponentsStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	newStatus := "approved"
	if status, ok := config["status"].(string); ok && status != "" {
		newStatus = status
	}

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	pageData := datahelpers.ExtractNestedField(params.CollectedData, pageField)
	if pageData == nil {
		return map[string]interface{}{"updated": false, "reason": "no page data", "status_set": newStatus}, nil
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	var pageID uuid.UUID
	if idStr, ok := page["id"].(string); ok && idStr != "" {
		pageID, _ = uuid.Parse(idStr)
	}

	reviewedAt := time.Now().UTC()
	if field, ok := config["reviewed_at_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			if t, err := time.Parse(time.RFC3339, val); err == nil {
				reviewedAt = t
			}
		}
	}

	reviewedBy := "system"
	if field, ok := config["reviewed_by_field"].(string); ok && field != "" {
		if val := datahelpers.ExtractNestedFieldString(params.CollectedData, field); val != "" {
			reviewedBy = val
		}
	}

	if params.DB != nil && pageID != uuid.Nil {
		query := `UPDATE page_components SET build_status = $1, reviewed_at = $2, reviewed_by = $3, updated_at = NOW() WHERE page_id = $4`
		result, err := params.DB.ExecContext(ctx, query, newStatus, reviewedAt, reviewedBy, pageID)
		if err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}
		rows, _ := result.RowsAffected()
		return map[string]interface{}{
			"updated": true, "rows_affected": rows, "page_id": pageID.String(),
			"status_set": newStatus, "reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
		}, nil
	}

	return map[string]interface{}{
		"updated": false, "reason": "no db or page_id", "status_set": newStatus,
		"reviewed_at": reviewedAt.Format(time.RFC3339), "reviewed_by": reviewedBy,
	}, nil
}

// ============================================================================
// SHARED SECTION-COMPONENT LOADER
// ============================================================================
//
// loadSectionComponents is the canonical loader for component rows used by
// section-level callers. Extracted from LoadPageSectionComponentsAction so
// plan_sections can reuse the same logic without a second SQL path. Both
// callers get the same component-row shape; differences in behaviour are
// expressed by what each caller does with the returned data, not by what
// each caller queries.
//
// Behaviour matches the previous in-action implementation:
//   - Match by name first, fall back to function (DISTINCT ON, newest first)
//   - Stubs for sections with no matching component
//   - Order preserved relative to sectionNames input
//   - When pageID != "", content_brief is attached per slot
//   - When activeOnly is true, only is_active=true rows are returned
//
// activeOnly preserves the historical behaviour difference between the two
// callers: plan_sections used to filter `is_active = true` inline;
// LoadPageSectionComponentsAction did not. Passing the flag explicitly keeps
// both callers' behaviour intact while sharing one query path.
//
// The returned per-component map carries: component_id (when from DB),
// name, function, display_name, category, semantic_tags (when set),
// description (when set), html_template (when set), input_schema (when
// set, as raw JSON string), render_mode, agent_type (when set),
// component_level, needs_llm, and content_brief (when found).
func loadSectionComponents(
	ctx context.Context,
	db *sql.DB,
	sectionNames []string,
	pageID string,
	activeOnly bool,
	logger *zap.Logger,
) []map[string]interface{} {
	if len(sectionNames) == 0 {
		return []map[string]interface{}{}
	}
	if db == nil {
		// No DB available — return name-stubs so callers can still proceed.
		return buildStubSectionComponents(sectionNames)
	}

	placeholders := make([]string, len(sectionNames))
	args := make([]interface{}, len(sectionNames))
	for i, name := range sectionNames {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = name
	}

	var components []map[string]interface{}

	// Pass 1: lookup by name
	activeFilter := ""
	if activeOnly {
		activeFilter = " AND is_active = true"
	}
	nameQuery := fmt.Sprintf(`
		SELECT
			id,
			name,
			COALESCE(display_name, name) AS display_name,
			function,
			COALESCE(category, '') AS category,
			semantic_tags,
			description,
			html_template,
			input_schema,
			COALESCE(render_mode, 'template') AS render_mode,
			agent_type,
			COALESCE(component_level, 'section') AS component_level
		FROM content_components
		WHERE name IN (%s)%s`, strings.Join(placeholders, ", "), activeFilter)

	rows, err := db.QueryContext(ctx, nameQuery, args...)
	if err != nil {
		logger.Error("loadSectionComponents: name query failed",
			zap.Error(err),
			zap.Strings("sections", sectionNames))
	} else {
		for rows.Next() {
			comp, scanErr := scanSectionComponentRow(rows)
			if scanErr != nil {
				logger.Error("loadSectionComponents: name row scan failed",
					zap.Error(scanErr))
				continue
			}
			components = append(components, comp)
		}
		rows.Close()
	}

	// Track which inputs are already satisfied by name or function
	foundNames := make(map[string]bool)
	for _, comp := range components {
		if n, ok := comp["name"].(string); ok {
			foundNames[n] = true
		}
		if fn, ok := comp["function"].(string); ok {
			foundNames[fn] = true
		}
	}

	var missing []string
	for _, name := range sectionNames {
		if !foundNames[name] {
			missing = append(missing, name)
		}
	}

	// Pass 2: lookup by function for anything still missing
	if len(missing) > 0 {
		logger.Info("loadSectionComponents: trying function lookup for missing",
			zap.Strings("missing", missing))

		funcPlaceholders := make([]string, len(missing))
		funcArgs := make([]interface{}, len(missing))
		for i, name := range missing {
			funcPlaceholders[i] = fmt.Sprintf("$%d", i+1)
			funcArgs[i] = name
		}

		funcQuery := fmt.Sprintf(`
			SELECT DISTINCT ON (function)
				id,
				name,
				COALESCE(display_name, name) AS display_name,
				function,
				COALESCE(category, '') AS category,
				semantic_tags,
				description,
				html_template,
				input_schema,
				COALESCE(render_mode, 'template') AS render_mode,
				agent_type,
				COALESCE(component_level, 'section') AS component_level
			FROM content_components
			WHERE function IN (%s)%s
			ORDER BY function, created_at DESC
		`, strings.Join(funcPlaceholders, ", "), activeFilter)

		funcRows, ferr := db.QueryContext(ctx, funcQuery, funcArgs...)
		if ferr != nil {
			logger.Warn("loadSectionComponents: function lookup failed",
				zap.Error(ferr))
		} else {
			for funcRows.Next() {
				comp, scanErr := scanSectionComponentRow(funcRows)
				if scanErr != nil {
					continue
				}
				function, _ := comp["function"].(string)
				if !containsString(missing, function) {
					continue
				}
				components = append(components, comp)
				foundNames[function] = true
				logger.Info("loadSectionComponents: found component by function",
					zap.String("function", function),
					zap.String("name", comp["name"].(string)))
			}
			funcRows.Close()
		}
	}

	// Stubs for anything still not found
	var stillMissing []string
	for _, name := range sectionNames {
		if !foundNames[name] {
			stillMissing = append(stillMissing, name)
		}
	}
	if len(stillMissing) > 0 {
		logger.Warn("loadSectionComponents: stubs for unresolved sections",
			zap.Strings("missing", stillMissing))
		for _, name := range stillMissing {
			components = append(components, map[string]interface{}{
				"name":         name,
				"display_name": name,
				"function":     name,
				"category":     "",
				"needs_llm":    true,
				"description":  "",
			})
		}
	}

	// Reorder to match sectionNames input order
	ordered := make([]map[string]interface{}, 0, len(components))
	for _, sectionName := range sectionNames {
		for _, comp := range components {
			name, _ := comp["name"].(string)
			function, _ := comp["function"].(string)
			if name == sectionName || function == sectionName {
				ordered = append(ordered, comp)
				break
			}
		}
	}

	// Optional: content_brief enrichment from page_components
	if pageID != "" {
		enrichSectionComponentsWithBriefs(ctx, db, pageID, ordered, logger)
	}

	return ordered
}

// scanSectionComponentRow turns one SQL row into the per-component map shape.
// Centralised so the by-name and by-function passes produce identical shapes.
func scanSectionComponentRow(rows *sql.Rows) (map[string]interface{}, error) {
	var id, name, function string
	var displayName, category sql.NullString
	var semanticTags, description, htmlTemplate, inputSchema sql.NullString
	var renderMode, agentType, componentLevel sql.NullString

	if err := rows.Scan(
		&id, &name, &displayName, &function, &category,
		&semanticTags, &description, &htmlTemplate, &inputSchema,
		&renderMode, &agentType, &componentLevel,
	); err != nil {
		return nil, err
	}

	comp := map[string]interface{}{
		"component_id": id,
		"name":         name,
		"function":     function,
	}
	if displayName.Valid {
		comp["display_name"] = displayName.String
	} else {
		comp["display_name"] = name
	}
	if category.Valid {
		comp["category"] = category.String
	} else {
		comp["category"] = ""
	}
	if semanticTags.Valid {
		comp["semantic_tags"] = semanticTags.String
	}
	if description.Valid && description.String != "" {
		comp["description"] = description.String
	}
	if htmlTemplate.Valid && htmlTemplate.String != "" {
		comp["html_template"] = htmlTemplate.String
	}
	if inputSchema.Valid && inputSchema.String != "" {
		comp["input_schema"] = inputSchema.String
	}
	if renderMode.Valid && renderMode.String != "" {
		comp["render_mode"] = renderMode.String
	} else {
		comp["render_mode"] = "template"
	}
	if agentType.Valid && agentType.String != "" {
		comp["agent_type"] = agentType.String
	}
	if componentLevel.Valid && componentLevel.String != "" {
		comp["component_level"] = componentLevel.String
	} else {
		comp["component_level"] = "section"
	}
	comp["needs_llm"] = detectNeedsLLMContent(htmlTemplate.String, inputSchema.String)
	return comp, nil
}

// buildStubSectionComponents returns minimal stubs for the no-DB code path.
func buildStubSectionComponents(sectionNames []string) []map[string]interface{} {
	stubs := make([]map[string]interface{}, len(sectionNames))
	for i, name := range sectionNames {
		stubs[i] = map[string]interface{}{
			"name":         name,
			"function":     name,
			"display_name": name,
			"description":  "",
			"needs_llm":    true,
		}
	}
	return stubs
}

// enrichSectionComponentsWithBriefs attaches per-section admin content briefs
// (from page_components.content_brief) onto the components in-place.
func enrichSectionComponentsWithBriefs(
	ctx context.Context,
	db *sql.DB,
	pageID string,
	components []map[string]interface{},
	logger *zap.Logger,
) {
	briefRows, briefErr := db.QueryContext(ctx, `
		SELECT COALESCE(slot_name, ''), content_brief
		FROM page_components
		WHERE page_id = $1
		  AND content_brief IS NOT NULL
		  AND build_status != 'removed'
	`, pageID)
	if briefErr != nil {
		logger.Warn("enrichSectionComponentsWithBriefs: query failed",
			zap.Error(briefErr))
		return
	}
	defer briefRows.Close()

	briefMap := make(map[string]interface{})
	for briefRows.Next() {
		var slotName string
		var briefJSON []byte
		if err := briefRows.Scan(&slotName, &briefJSON); err != nil {
			continue
		}
		if len(briefJSON) > 0 && slotName != "" {
			var brief interface{}
			if err := json.Unmarshal(briefJSON, &brief); err == nil {
				briefMap[slotName] = brief
			}
		}
	}
	if len(briefMap) == 0 {
		return
	}

	for _, comp := range components {
		name, _ := comp["name"].(string)
		function, _ := comp["function"].(string)
		if brief, ok := briefMap[name]; ok {
			comp["content_brief"] = brief
		} else if brief, ok := briefMap[function]; ok {
			comp["content_brief"] = brief
		}
	}
	logger.Info("enrichSectionComponentsWithBriefs: attached briefs",
		zap.Int("briefs_found", len(briefMap)))
}

// ============================================================================
// ACTION: load_page_section_components
// ============================================================================
//
// LoadPageSectionComponentsAction loads component definitions for a page's
// sections. Thin workflow wrapper around loadSectionComponents.
//
// Config:
//   - page_from: collected_data path to the page record (default "current_page")
//   - include_templates, include_input_schema: kept for back-compat; the shared
//     loader always returns both, so these flags are not consulted. See
//     doc 019_tool_library.md for the rationale.
//
// When call_agent passes input_fields, they arrive under input_data.*
// extractWithInputDataFallback handles both root and input_data locations.
func LoadPageSectionComponentsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("LoadPageSectionComponentsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	pageField := "current_page"
	if pf, ok := config["page_from"].(string); ok && pf != "" {
		pageField = pf
	}

	pageData := extractWithInputDataFallback(params.CollectedData, pageField, params.Logger)
	if pageData == nil {
		keys := make([]string, 0, len(params.CollectedData))
		for k := range params.CollectedData {
			keys = append(keys, k)
		}
		params.Logger.Error("Page not found",
			zap.String("page_field", pageField),
			zap.Strings("available_keys", keys))
		return nil, fmt.Errorf("page not found at '%s'", pageField)
	}

	page, ok := pageData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("page must be object")
	}

	params.Logger.Info("Found page data",
		zap.String("page_field", pageField),
		zap.Any("page_name", page["name"]),
		zap.Any("page_title", page["title"]))

	sectionsRaw := page["sections"]
	if sectionsRaw == nil {
		sectionsRaw = extractWithInputDataFallback(params.CollectedData, "sections", params.Logger)
	}

	sectionNames := datahelpers.ExtractSectionNamesHelper(sectionsRaw)
	if len(sectionNames) == 0 {
		params.Logger.Warn("No sections found for page")
		return map[string]interface{}{
			"components":    []interface{}{},
			"count":         0,
			"from_database": false,
		}, nil
	}

	// Enforce naming contract: LLM site plans may output "social_proof" or "SocialProof"
	NormalizeSectionNames(sectionNames, params.Logger)

	params.Logger.Info("Loading components for sections",
		zap.Strings("sections", sectionNames))

	pageID, _ := page["id"].(string)
	// activeOnly=false: this action historically returned components regardless
	// of is_active. Preserved here so existing callers (page-content-writer's
	// load_page_components step, audit flows, admin tools) behave identically.
	components := loadSectionComponents(ctx, params.DB, sectionNames, pageID, false, params.Logger)

	// Detect whether any row carries a real component_id (i.e. came from DB)
	// to preserve the from_database signal callers rely on.
	fromDB := false
	for _, c := range components {
		if _, has := c["component_id"]; has {
			fromDB = true
			break
		}
	}

	return map[string]interface{}{
		"components":    components,
		"count":         len(components),
		"from_database": fromDB,
		"requested":     sectionNames,
	}, nil
}

// Order: direct path -> input_data.{path} -> FindByPath helper
// extractWithInputDataFallback tries to extract a field, falling back to input_data prefix
// This handles the common case where workflows specify paths like "current_section.name"
// but the data is actually at "input_data.current_section.name"
func extractWithInputDataFallback(data map[string]interface{}, path string, logger *zap.Logger) interface{} {
	// Try direct path first
	if value := datahelpers.ExtractNestedField(data, path); value != nil {
		logger.Debug("extractWithInputDataFallback: Found at direct path",
			zap.String("path", path),
		)
		return value
	}

	// If path doesn't already start with input_data, try with prefix
	if !strings.HasPrefix(path, "input_data.") {
		prefixedPath := "input_data." + path
		if value := datahelpers.ExtractNestedField(data, prefixedPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via input_data prefix",
				zap.String("original_path", path),
				zap.String("actual_path", prefixedPath),
			)
			return value
		}
	}

	// Try in __raw_message__.body.input_data (deeply nested case from child agents)
	if !strings.HasPrefix(path, "__raw_message__") {
		rawMsgPath := "__raw_message__.body.input_data." + path
		if value := datahelpers.ExtractNestedField(data, rawMsgPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via __raw_message__.body.input_data",
				zap.String("original_path", path),
			)
			return value
		}
	}

	// Also try agent_config location (for workflow config data)
	if !strings.HasPrefix(path, "agent_config") {
		agentConfigPath := "agent_config." + path
		if value := datahelpers.ExtractNestedField(data, agentConfigPath); value != nil {
			logger.Debug("extractWithInputDataFallback: Found via agent_config",
				zap.String("original_path", path),
			)
			return value
		}
	}

	logger.Debug("extractWithInputDataFallback: Not found anywhere",
		zap.String("path", path),
	)
	return nil
}

// ============================================================================
// ACTION: filter_search_results
// ============================================================================

// FilterSearchResultsAction filters search results based on criteria
// Handles various response formats: direct array, {results: []}, {data: {results: []}}, etc.
func FilterSearchResultsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("FilterSearchResultsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	resultsField := "search_results"
	if rf, ok := config["results_field"].(string); ok && rf != "" {
		resultsField = rf
	}

	// Try to find results array - handles various response formats
	results := datahelpers.FindResultsArray(params.CollectedData, resultsField, params.Logger)
	if results == nil {
		params.Logger.Warn("FilterSearchResultsAction: No results found",
			zap.String("results_field", resultsField),
			zap.Strings("collected_data_keys", datahelpers.GetMapKeys(params.CollectedData)))
		return map[string]interface{}{"filtered_results": []interface{}{}, "count": 0, "original_count": 0}, nil
	}

	params.Logger.Info("FilterSearchResultsAction: Found results",
		zap.Int("count", len(results)))

	// Support both max_results and max_sources config keys
	maxResults := 10
	if mr, ok := config["max_results"].(float64); ok {
		maxResults = int(mr)
	} else if ms, ok := config["max_sources"].(float64); ok {
		maxResults = int(ms)
	}

	excludePatterns := datahelpers.ExtractStringListHelper(config["exclude_patterns"])
	requiredKeywords := datahelpers.ExtractStringListHelper(config["required_keywords"])
	preferDomains := datahelpers.ExtractStringListHelper(config["prefer_domains"])

	var filtered []interface{}
	var preferred []interface{}

	for _, r := range results {
		result, ok := r.(map[string]interface{})
		if !ok {
			continue
		}

		title, _ := result["title"].(string)
		content, _ := result["content"].(string)
		snippet, _ := result["snippet"].(string)
		url, _ := result["url"].(string)
		searchText := strings.ToLower(title + " " + content + " " + snippet + " " + url)

		// Check exclusions
		excluded := false
		for _, pattern := range excludePatterns {
			if strings.Contains(searchText, strings.ToLower(pattern)) {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		// Check required keywords
		if len(requiredKeywords) > 0 {
			hasKeyword := false
			for _, kw := range requiredKeywords {
				if strings.Contains(searchText, strings.ToLower(kw)) {
					hasKeyword = true
					break
				}
			}
			if !hasKeyword {
				continue
			}
		}

		// Check if from preferred domain
		isPreferred := false
		for _, domain := range preferDomains {
			if strings.Contains(strings.ToLower(url), strings.ToLower(domain)) {
				isPreferred = true
				break
			}
		}

		if isPreferred {
			preferred = append(preferred, result)
		} else {
			filtered = append(filtered, result)
		}
	}

	// Combine: preferred first, then others, up to maxResults
	combined := append(preferred, filtered...)
	if len(combined) > maxResults {
		combined = combined[:maxResults]
	}

	params.Logger.Info("FilterSearchResultsAction: Complete",
		zap.Int("original", len(results)),
		zap.Int("preferred", len(preferred)),
		zap.Int("filtered", len(combined)))

	return map[string]interface{}{
		"filtered_results": combined,
		"count":            len(combined),
		"original_count":   len(results),
		"preferred_count":  len(preferred),
	}, nil
}

// ============================================================================
// ACTION: extract_fields
// ============================================================================

func ExtractFieldsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("ExtractFieldsAction: Starting",
		zap.Any("config_keys", getConfigKeys(params.StepConfig.Config)))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config
	result := make(map[string]interface{})

	// Handle fields array format: ["field1", "field2"]
	if fields, ok := config["fields"].([]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as array")
		for _, f := range fields {
			switch field := f.(type) {
			case string:
				if value := datahelpers.ExtractNestedField(params.CollectedData, field); value != nil {
					parts := strings.Split(field, ".")
					result[parts[len(parts)-1]] = value
				}
			case map[string]interface{}:
				source, _ := field["source"].(string)
				target, _ := field["target"].(string)
				if source == "" {
					continue
				}
				if target == "" {
					parts := strings.Split(source, ".")
					target = parts[len(parts)-1]
				}
				if value := datahelpers.ExtractNestedField(params.CollectedData, source); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Handle fields as map-of-arrays format (fallback paths)
	// Example: {"topic": ["path1", "path2"], "company": ["path3"]}
	if fieldsMap, ok := config["fields"].(map[string]interface{}); ok {
		params.Logger.Info("ExtractFieldsAction: Processing fields as map-of-arrays",
			zap.Int("field_count", len(fieldsMap)))

		for targetField, pathsRaw := range fieldsMap {
			var found bool

			// Handle array of paths
			if paths, ok := pathsRaw.([]interface{}); ok {
				for _, pathRaw := range paths {
					if path, ok := pathRaw.(string); ok {
						// Try direct path first
						if value := datahelpers.ExtractNestedField(params.CollectedData, path); value != nil {
							result[targetField] = value
							found = true
							params.Logger.Info("ExtractFieldsAction: Found via direct path",
								zap.String("target", targetField),
								zap.String("path", path))
							break
						}

						// Try with input_data prefix
						if !strings.HasPrefix(path, "input_data.") {
							prefixedPath := "input_data." + path
							if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
								result[targetField] = value
								found = true
								params.Logger.Info("ExtractFieldsAction: Found via input_data prefix",
									zap.String("target", targetField),
									zap.String("original_path", path),
									zap.String("prefixed_path", prefixedPath))
								break
							}
						}
					}
				}
			}

			// Handle single string path (not array)
			if singlePath, ok := pathsRaw.(string); ok && !found {
				if value := datahelpers.ExtractNestedField(params.CollectedData, singlePath); value != nil {
					result[targetField] = value
					found = true
				} else if !strings.HasPrefix(singlePath, "input_data.") {
					prefixedPath := "input_data." + singlePath
					if value := datahelpers.ExtractNestedField(params.CollectedData, prefixedPath); value != nil {
						result[targetField] = value
						found = true
					}
				}
			}

			if !found {
				params.Logger.Warn("ExtractFieldsAction: Field not found in any path",
					zap.String("target", targetField),
					zap.Any("tried_paths", pathsRaw))
			}
		}
	}

	// Handle field_map format: {"target": "source"}
	if fieldMap, ok := config["field_map"].(map[string]interface{}); ok {
		for target, source := range fieldMap {
			if sourceStr, ok := source.(string); ok {
				if value := datahelpers.ExtractNestedField(params.CollectedData, sourceStr); value != nil {
					result[target] = value
				}
			}
		}
	}

	// Apply defaults
	if defaults, ok := config["defaults"].(map[string]interface{}); ok {
		for key, val := range defaults {
			if result[key] == nil {
				result[key] = val
			}
		}
	}

	params.Logger.Info("ExtractFieldsAction: Complete",
		zap.Int("fields_extracted", len(result)),
		zap.Strings("result_keys", datahelpers.GetMapKeys(result)))
	return result, nil
}

// Priority 5: Try to build query from section context (for research-agent)
// Check both root level and inside input_data
func getSearchQueryFromSectionContext(params ActionParams) string {
	var currentSection map[string]interface{}

	// Try root level first
	if cs, ok := params.CollectedData["current_section"].(map[string]interface{}); ok {
		currentSection = cs
	}
	// Try input_data.current_section
	if currentSection == nil {
		if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
			if cs, ok := inputData["current_section"].(map[string]interface{}); ok {
				currentSection = cs
			}
		}
	}

	if currentSection == nil {
		return ""
	}

	// Get section name/function
	sectionName := ""
	if fn, ok := currentSection["function"].(string); ok && fn != "" {
		sectionName = fn
	} else if name, ok := currentSection["name"].(string); ok && name != "" {
		sectionName = name
	}

	if sectionName == "" {
		return ""
	}

	// Get domain
	domain := ""
	if d, ok := params.CollectedData["domain"].(string); ok {
		domain = d
	} else if siteRecord, ok := params.CollectedData["site_record"].(map[string]interface{}); ok {
		if d, ok := siteRecord["domain"].(string); ok {
			domain = d
		}
	} else if inputData, ok := params.CollectedData["input_data"].(map[string]interface{}); ok {
		if siteRecord, ok := inputData["site_record"].(map[string]interface{}); ok {
			if d, ok := siteRecord["domain"].(string); ok {
				domain = d
			}
		}
	}

	// Build query
	query := sectionName
	if domain != "" {
		query = fmt.Sprintf("%s %s", sectionName, domain)
	}

	return query
}

// containsString checks if a string slice contains a specific string
// extractContentWithFallbacks tries multiple paths to find content data
// This handles different output formats from execute_llm_prompt (with/without .result wrapper)
func extractContentWithFallbacks(data map[string]interface{}, contentField string, logger *zap.Logger) map[string]interface{} {
	// Build list of paths to try
	pathsToTry := []string{contentField}

	// If path ends with ".result", also try without it
	if strings.HasSuffix(contentField, ".result") {
		base := strings.TrimSuffix(contentField, ".result")
		pathsToTry = append(pathsToTry, base)
		pathsToTry = append(pathsToTry, base+".response")
		pathsToTry = append(pathsToTry, base+".content")
	} else {
		// If path doesn't end with .result, also try with it
		pathsToTry = append(pathsToTry, contentField+".result")
		pathsToTry = append(pathsToTry, contentField+".response")
		pathsToTry = append(pathsToTry, contentField+".content")
	}

	// Try each path
	for _, path := range pathsToTry {
		if extracted := datahelpers.ExtractNestedField(data, path); extracted != nil {
			if m, ok := extracted.(map[string]interface{}); ok && len(m) > 0 {
				logger.Debug("extractContentWithFallbacks: Found content",
					zap.String("path", path),
					zap.Int("field_count", len(m)))
				return m
			}
		}
	}

	// Last resort: check if the base field exists and contains the content directly
	// Sometimes LLM output is stored as the field value itself
	parts := strings.Split(contentField, ".")
	if len(parts) > 0 {
		baseField := parts[0]
		if baseData := datahelpers.ExtractNestedField(data, baseField); baseData != nil {
			if m, ok := baseData.(map[string]interface{}); ok && len(m) > 0 {
				// Check if this map contains content-like fields
				if hasContentFields(m) {
					logger.Debug("extractContentWithFallbacks: Found content at base field",
						zap.String("base_field", baseField),
						zap.Int("field_count", len(m)))
					return m
				}
			}
		}
	}

	logger.Debug("extractContentWithFallbacks: No content found",
		zap.String("original_path", contentField),
		zap.Strings("tried_paths", pathsToTry))

	return nil
}

// hasContentFields checks if a map contains typical content field names
func hasContentFields(m map[string]interface{}) bool {
	contentFieldNames := []string{
		"headline", "subheadline", "body", "content", "heading",
		"title", "description", "text", "features", "items",
		"primary_cta", "cta_text", "button_text",
	}
	for _, fieldName := range contentFieldNames {
		if _, exists := m[fieldName]; exists {
			return true
		}
	}
	return false
}

// detectNeedsLLMContent determines if a component template needs LLM-generated content
// Returns true if template has content placeholders that need dynamic content
// Returns false if template only has structural placeholders (domain, logo_text, etc.)
func detectNeedsLLMContent(htmlTemplate, inputSchema string) bool {
	// No template = needs LLM to generate something
	if htmlTemplate == "" {
		return true
	}

	// If has input_schema, it defines what content is needed
	if inputSchema != "" && inputSchema != "{}" && inputSchema != "null" {
		return true
	}

	// Content placeholders that indicate LLM content is needed
	contentPlaceholders := []string{
		"{{.headline}}", "{{.subheadline}}", "{{.body}}", "{{.content}}",
		"{{.description}}", "{{.text}}", "{{.paragraph}}",
		"{{.primary_cta}}", "{{.secondary_cta}}", "{{.cta_text}}",
		"{{.features}}", "{{.services}}", "{{.benefits}}", "{{.items}}",
		"{{.testimonial}}", "{{.quote}}", "{{.author}}",
		"{{.heading}}", "{{.subtitle}}",
		"{{ .headline}}", "{{ .body}}", "{{ .content}}",
		"{{range .features}}", "{{range .services}}", "{{range .items}}",
		"{{range .benefits}}", "{{range .testimonials}}",
	}

	// Check for content placeholders
	for _, placeholder := range contentPlaceholders {
		if strings.Contains(htmlTemplate, placeholder) {
			return true
		}
	}

	// Structural fields that DON'T require LLM (filled from render_context)
	structuralFields := map[string]bool{
		"domain": true, "logo_text": true, "company_name": true,
		"tagline": true, "current_page": true, "year": true,
		"primary_color": true, "secondary_color": true, "accent_color": true,
		"background_color": true, "text_color": true,
		"email": true, "phone": true, "address": true,
		"nav_items_html": true, "theme_css": true, "title": true,
		"cta_url": true, "industry": true, "tone": true,
	}

	// Check for any {{.X}} where X is not a known structural field
	re := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(htmlTemplate, -1)
	for _, match := range matches {
		if len(match) > 1 {
			fieldName := match[1]
			if !structuralFields[fieldName] {
				return true
			}
		}
	}

	// Only structural placeholders - doesn't need LLM
	return false
}

// UpdateWorkItemStatusAction updates a site_work_items row's status.
// Config:
//   - work_item_id_field: path to work_item_id in collected_data
//     (default: "input_data.work_item_id")
//   - status:             new status — default "complete"
//     (valid: complete, failed, claimed, executing,
//     detected, wont_fix)
//   - skip_if_missing:    bool — when true (default), gracefully no-op if
//     work_item_id absent. When false, error.
//   - result_fields:      optional map of extra fields to merge into the
//     row's result JSONB. Values are literals; the
//     action always adds orchestration_id and step
//     metadata automatically.
func UpdateWorkItemStatusAction(ctx context.Context, params ActionParams) (interface{}, error) {
	params.Logger.Info("UpdateWorkItemStatusAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}

	config := params.StepConfig.Config

	// Get work_item_id
	workItemIDField := "input_data.work_item_id"
	if f, ok := config["work_item_id_field"].(string); ok && f != "" {
		workItemIDField = f
	}
	workItemIDStr := datahelpers.ExtractNestedFieldString(params.CollectedData, workItemIDField)

	// Skip gracefully if missing — supports manual triggers without work_item_id
	skipIfMissing := true
	if v, ok := config["skip_if_missing"].(bool); ok {
		skipIfMissing = v
	}

	if workItemIDStr == "" {
		if skipIfMissing {
			params.Logger.Info("UpdateWorkItemStatusAction: no work_item_id in collected_data; skipping",
				zap.String("looked_at", workItemIDField))
			return map[string]interface{}{
				"updated": false,
				"skipped": true,
				"reason":  "work_item_id not present",
			}, nil
		}
		return nil, fmt.Errorf("work_item_id not found at %s and skip_if_missing=false", workItemIDField)
	}

	workItemID, err := uuid.Parse(workItemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id %q: %w", workItemIDStr, err)
	}

	// Status (default complete)
	newStatus := "complete"
	if s, ok := config["status"].(string); ok && s != "" {
		newStatus = s
	}
	validStatuses := map[string]bool{
		"complete":  true,
		"failed":    true,
		"claimed":   true,
		"executing": true,
		"detected":  true,
		"wont_fix":  true,
	}
	if !validStatuses[newStatus] {
		return nil, fmt.Errorf("invalid work item status: %s (valid: complete, failed, claimed, executing, detected, wont_fix)", newStatus)
	}

	if params.DB == nil {
		params.Logger.Warn("UpdateWorkItemStatusAction: No database")
		return map[string]interface{}{"updated": false, "status": newStatus}, nil
	}

	// Build result payload — always includes orchestration tracking; merges
	// any caller-supplied extras under result_fields.
	resultPayload := map[string]interface{}{
		"completed_by_orchestration_id": params.ExecutionContext.OrchestrationID,
		"completed_by_step":             params.ExecutionContext.StepName,
		"completed_at_iso":              time.Now().UTC().Format(time.RFC3339),
	}
	if extras, ok := config["result_fields"].(map[string]interface{}); ok {
		for k, v := range extras {
			resultPayload[k] = v
		}
	}
	resultJSON, err := json.Marshal(resultPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result payload: %w", err)
	}

	// Build query. completed_at only set when transitioning to complete; for
	// other statuses (failed, etc.) leave it alone and just update status.
	var query string
	if newStatus == "complete" {
		query = `UPDATE site_work_items
		            SET status = $2,
		                completed_at = NOW(),
		                updated_at = NOW(),
		                attempt_count = attempt_count + 1,
		                result = COALESCE(result, '{}'::jsonb) || $3::jsonb
		          WHERE id = $1`
	} else {
		query = `UPDATE site_work_items
		            SET status = $2,
		                updated_at = NOW(),
		                attempt_count = attempt_count + 1,
		                result = COALESCE(result, '{}'::jsonb) || $3::jsonb
		          WHERE id = $1`
	}

	if err := execDB(ctx, params.DB, query, workItemID, newStatus, resultJSON); err != nil {
		return nil, fmt.Errorf("failed to update work item status: %w", err)
	}

	params.Logger.Info("UpdateWorkItemStatusAction: status updated",
		zap.String("work_item_id", workItemIDStr),
		zap.String("status", newStatus),
	)

	return map[string]interface{}{
		"updated":      true,
		"work_item_id": workItemIDStr,
		"status":       newStatus,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ============================================================================
// Adoption-faithfulness convergence (doc 029 Phase 1 / FOCUS_adoption_faithfulness_via_locks.md)
//
// These helpers make ValidateSitePlanAction deterministically preserve the
// pages that are currently under a live adoption lock, so the planner LLM
// cannot drop, rename, or duplicate adopted pages during the faithful-first-
// pass window. They become no-ops once the adoption lock has expired (or for
// from-scratch builds), letting the site develop normally thereafter.
// ============================================================================

// isSectionIndexType reports whether a page_type is a directory/section index.
func isSectionIndexType(pageType string) bool {
	switch pageType {
	case "blog-index", "entity-directory", "section-index":
		return true
	}
	return false
}

// sectionStemOf returns the section stem for a realised section-index page, or
// "" if it isn't a section index. e.g. ("games-index", "/games/index.html",
// "entity-directory") -> "games". Prefers the URL path segment; falls back to
// the name minus the "-index" suffix.
func sectionStemOf(name, url, pageType string) string {
	isIndex := isSectionIndexType(pageType)
	if !isIndex {
		// Also treat any non-root URL ending in /index.html as a section index.
		if strings.HasSuffix(url, "/index.html") && url != "/index.html" {
			isIndex = true
		}
	}
	if !isIndex {
		return ""
	}
	trimmed := strings.Trim(url, "/")
	if i := strings.Index(trimmed, "/"); i > 0 {
		return trimmed[:i]
	}
	return strings.TrimSuffix(name, "-index")
}

// slugOf derives a comparison slug from an LLM page's name/url. A flat page URL
// like /games.html yields "games"; falls back to the name.
func slugOf(name, url string) string {
	if url != "" {
		t := strings.Trim(url, "/")
		t = strings.TrimSuffix(t, ".html")
		t = strings.TrimSuffix(t, "/index")
		if t != "" {
			if i := strings.Index(t, "/"); i > 0 {
				return t[:i]
			}
			return t
		}
	}
	return name
}

// itemStemOf returns the topic stem of an item page name by stripping the
// role prefixes that CanonicalisePage adds (tool-, guide-, game-): e.g.
// "guide-economy-basics" -> "economy-basics", "economy-basics" ->
// "economy-basics". Mirrors the TrimPrefix calls in CanonicalisePage's
// tool/guide/game cases - keep this prefix list in sync with them. Returns
// the input unchanged when no role prefix is present, so two adopted pages on
// the same topic share a stem and a re-proposed bare sibling collides with
// them. Unlike sectionStemOf, this is name-based rather than URL/hub-based.
func itemStemOf(name string) string {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, p := range []string{"tool-", "guide-", "game-"} {
		if strings.HasPrefix(n, p) {
			return strings.TrimPrefix(n, p)
		}
	}
	return n
}

// normaliseRealisedToPlanPage converts a realised pages-table row (as returned
// by the load_existing_pages query) into the plan-page shape the downstream
// write_site_plan / ValidateRoles expects. Carries a from_realised marker so
// logging/debugging can distinguish these from LLM-proposed pages.
func normaliseRealisedToPlanPage(rm map[string]interface{}) map[string]interface{} {
	name, _ := rm["name"].(string)
	pageType, _ := rm["page_type"].(string)
	// Carry the realised page's sections so a unioned adopted page keeps them.
	// load_existing_pages runs via query_database, which stringifies jsonb
	// columns, so rm["sections"] arrives as a JSON string (e.g. ["hero",
	// "guide-list"]); tolerate a native []interface{} too. Empty/missing -> [].
	// Without carrying these, the union emits empty values and the page sync's
	// "<col> = EXCLUDED.<col>" clobbers the adopted page's real sections,
	// meta_description, and nav_order. (nav_label is COALESCE-preserved by the
	// upsert, so it is safe without carrying.)
	sections := []interface{}{}
	switch v := rm["sections"].(type) {
	case []interface{}:
		sections = v
	case string:
		if v != "" {
			var parsed []interface{}
			if err := json.Unmarshal([]byte(v), &parsed); err == nil {
				sections = parsed
			}
		}
	}
	return map[string]interface{}{
		"name":             name,
		"page_type":        pageType,
		"url":              rm["url"],
		"title":            rm["title"],
		"nav_label":        rm["nav_label"],
		"in_header":        rm["in_header"],
		"in_footer":        rm["in_footer"],
		"meta_description": rm["meta_description"],
		"nav_order":        rm["nav_order"],
		"sections":         sections,
		"from_realised":    true,
	}
}

// reconcilePlanWithRealised enforces preservation of and convergence on the
// realised pages that are CURRENTLY under an active adoption lock.
//
// The load_existing_pages query carries an "adoption_locked" boolean per page:
// true while the 90-day adoption lock is live, false once expired (or for any
// page never adoption-locked). This function force-preserves ONLY the
// adoption_locked pages:
//   - During the window: every adopted page is locked -> preserved faithfully.
//   - After the window: nothing locked -> no-op -> site develops normally.
//   - From-scratch builds: never locked -> always a no-op.
//
// Three passes over the locked subset:
//
//	Pass C — section-collision dedup: drop an LLM page whose slug equals the
//	         stem of a realised section index ("games" vs "games-index").
//	Pass B — rename snap-back: same URL as a realised page, different name ->
//	         replace with the realised identity.
//	Pass A — union: append every locked realised page not already present.
//
// Returns the reconciled page slice plus counts for logging.
func reconcilePlanWithRealised(
	llmPages []interface{},
	existingPages []interface{},
	logger *zap.Logger,
) ([]interface{}, int, int, int) {
	// Force-preserve only the pages under a live adoption lock.
	var lockedPages []interface{}
	for _, rp := range existingPages {
		if rm, ok := rp.(map[string]interface{}); ok {
			if locked, _ := rm["adoption_locked"].(bool); locked {
				lockedPages = append(lockedPages, rp)
			}
		}
	}
	if len(lockedPages) == 0 {
		// No active adoption locks: post-window, from-scratch, or a normally-
		// developing site. Leave the LLM plan untouched.
		return llmPages, 0, 0, 0
	}
	existingPages = lockedPages

	realisedByURL := make(map[string]map[string]interface{})
	sectionStems := make(map[string]string) // stem -> realised index name
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		url, _ := rm["url"].(string)
		pageType, _ := rm["page_type"].(string)
		if url != "" {
			realisedByURL[url] = rm
		}
		if stem := sectionStemOf(name, url, pageType); stem != "" {
			sectionStems[stem] = name
		}
	}

	// Item-topic stems: the role-prefix-stripped name stem of each realised
	// page (guide-economy-basics -> economy-basics). Keyed to a SET of realised
	// names so a topic legitimately covered by two adopted pages (e.g. a tool
	// and a guide) does not false-positive on either of them. Lets Pass C2 drop
	// an LLM page that re-proposes an adopted item under a different
	// prefix/role/URL.
	itemStemSets := make(map[string]map[string]bool)
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		if name == "" {
			continue
		}
		stem := itemStemOf(name)
		if stem == "" {
			continue
		}
		if itemStemSets[stem] == nil {
			itemStemSets[stem] = make(map[string]bool)
		}
		itemStemSets[stem][name] = true
	}

	// Pass C (collision) + Pass B (rename) over the LLM pages.
	var kept []interface{}
	droppedCollision, snappedRename := 0, 0
	for _, lp := range llmPages {
		lm, ok := lp.(map[string]interface{})
		if !ok {
			kept = append(kept, lp)
			continue
		}
		lname, _ := lm["name"].(string)
		lurl, _ := lm["url"].(string)
		ltype, _ := lm["page_type"].(string)
		lslug := slugOf(lname, lurl)

		// Pass C: flat page colliding with a realised section index.
		if idxName, isStem := sectionStems[lslug]; isStem &&
			!isSectionIndexType(ltype) && lname != idxName {
			logger.Info("validate: dropped flat page colliding with realised section index",
				zap.String("dropped", lname), zap.String("kept_index", idxName))
			droppedCollision++
			continue
		}

		// Pass C2: item-topic collision - the LLM re-proposes an adopted item
		// under a different name/prefix/role (e.g. "economy-basics" beside the
		// adopted "guide-economy-basics"; different URL, so Pass B misses it).
		// Drop it; the adopted page already covers the topic. Skips when the LLM
		// name IS one of the realised names for that stem (a preserved page).
		if names, isStem := itemStemSets[itemStemOf(lname)]; isStem && !names[lname] {
			logger.Info("validate: dropped page duplicating an adopted item topic",
				zap.String("dropped", lname),
				zap.String("stem", itemStemOf(lname)))
			droppedCollision++
			continue
		}

		// Pass B: same URL as a realised page, different name -> snap back.
		if rp, ok := realisedByURL[lurl]; ok {
			if rname, _ := rp["name"].(string); rname != "" && rname != lname {
				logger.Info("validate: snapped renamed page back to realised identity",
					zap.String("llm_name", lname), zap.String("realised_name", rname))
				kept = append(kept, normaliseRealisedToPlanPage(rp))
				snappedRename++
				continue
			}
		}
		kept = append(kept, lm)
	}

	// Pass A: union — add locked realised pages not present by name.
	presentName := make(map[string]bool)
	for _, p := range kept {
		if pm, ok := p.(map[string]interface{}); ok {
			if n, _ := pm["name"].(string); n != "" {
				presentName[n] = true
			}
		}
	}
	unioned := 0
	for _, rp := range existingPages {
		rm, ok := rp.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := rm["name"].(string)
		if name == "" || presentName[name] {
			continue
		}
		kept = append(kept, normaliseRealisedToPlanPage(rm))
		presentName[name] = true
		unioned++
	}

	return kept, unioned, droppedCollision, snappedRename
}

// truncatePreservingRealised caps the plan at maxPages but never drops a
// must-keep (adoption-locked) page. Locked pages are kept first; net-new
// proposed pages fill the remaining budget in order.
func truncatePreservingRealised(
	pages, mustKeep []interface{},
	maxPages int,
	logger *zap.Logger,
) []interface{} {
	keepNames := make(map[string]bool)
	for _, rp := range mustKeep {
		if rm, ok := rp.(map[string]interface{}); ok {
			if n, _ := rm["name"].(string); n != "" {
				keepNames[n] = true
			}
		}
	}
	var locked, netNew []interface{}
	for _, p := range pages {
		name := ""
		if pm, ok := p.(map[string]interface{}); ok {
			name, _ = pm["name"].(string)
		}
		if keepNames[name] {
			locked = append(locked, p)
		} else {
			netNew = append(netNew, p)
		}
	}
	if len(locked) >= maxPages {
		logger.Warn("validate: adoption-locked pages exceed max_pages; keeping all locked, dropping all net-new",
			zap.Int("locked", len(locked)), zap.Int("max_pages", maxPages))
		return locked
	}
	budget := maxPages - len(locked)
	if budget > len(netNew) {
		budget = len(netNew)
	}
	return append(locked, netNew[:budget]...)
}

```

### platform/orchestration/actions/reconcile_section_data_action.go (package `actions`) — whole file

```go
// FILE: platform/orchestration/actions/reconcile_section_data_action.go
//
// ReconcileSectionDataAction re-attempts open needs_section_data items whose
// missing data is query-resolvable, and re-renders the affected page once that
// data exists. It closes the loop the original needs_section_data design left
// open, WITHOUT a new LLM agent.
//
// CONTEXT
//   plan_sections defers a section to needs_section_data (status
//   needs_human_review) when a required field can't resolve. For list/grid
//   sections the missing field is query-sourced (query.pages_where_type:...,
//   query.pages_under_section:...) — derived data, not human data. It defers
//   when the listed pages don't exist yet (or, until recently, when the query
//   wasn't implemented). Those items then sit forever because nothing re-plans
//   the page after the data lands; loadOpenSectionDataRequests +
//   closeResolvedDataRequest only close an item on a LATER plan_sections run.
//
// WHAT IT DOES
//   For each open needs_section_data item whose missing fields are ALL
//   query-sourced, it re-runs the query now (queryresolve.Resolve). If the data
//   now exists, it emits a needs_page so page-build-handler re-renders the page
//   through plan_sections, which re-resolves the field and auto-closes the item
//   via closeResolvedDataRequest. Items with any non-query (human) missing
//   field — team, pricing, case studies, contact — are left at
//   needs_human_review untouched (distinguished by spec.missing[].source).
//
//   It does NOT close items itself: re-rendering through plan_sections is the
//   single source of truth for resolution + close, so the page actually gets
//   the data, not just a cleared flag.
//
// WIRING (recommended): run periodically in the improvement loop, or as a step
//   in the finalize/rerender path AFTER pages are built (running it at
//   plan time is too early — the listed pages don't exist yet). It is
//   invocation-agnostic: given site_id it scans and re-triggers. Not wired
//   here — pick the host agent and add a step:
//     "reconcile_section_data": {
//        "action": "reconcile_section_data",
//        "config": { "site_id": "<path to site_id>" },
//        "next_step": "..." }
//
// REGISTRATION (registry.go):
//   "reconcile_section_data": {
//       Handler:     ReconcileSectionDataAction,
//       Category:    "site",
//       Description: "Re-trigger pages whose deferred section data is now query-resolvable",
//       IsLocal:     true,
//   }

package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/actions/queryresolve"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

var ReconcileSectionDataInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id"},
	Optional: []string{"page_name"},
	Defaults: map[string]interface{}{},
}

func init() {
	datahelpers.RegisterActionInputSpec("reconcile_section_data", ReconcileSectionDataInputSpec)
}

// reconcileMissingField mirrors the fields of plan_sections' missingField that
// the reconciler reads from needs_section_data spec.missing[].
type reconcileMissingField struct {
	Field  string `json:"field"`
	Source string `json:"source"`
}

type reconcileSpec struct {
	PageName    string                  `json:"page_name"`
	SectionName string                  `json:"section_name"`
	Missing     []reconcileMissingField `json:"missing"`
}

func ReconcileSectionDataAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(zap.String("action", "reconcile_section_data"))

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData, params.StepConfig.Config,
		ReconcileSectionDataInputSpec, logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}
	siteID, err := uuid.Parse(inputs.Get("site_id"))
	if err != nil {
		return nil, fmt.Errorf("invalid site_id %q: %w", inputs.Get("site_id"), err)
	}
	pageFilter := inputs.Get("page_name") // optional scope

	// Load open needs_section_data items for this site.
	rows, err := params.DB.QueryContext(ctx, `
		SELECT id, spec
		FROM site_work_items
		WHERE site_id = $1
		  AND item_type = 'needs_section_data'
		  AND status IN ('needs_human_review', 'triaged')
	`, siteID)
	if err != nil {
		return nil, fmt.Errorf("load open section-data items: %w", err)
	}
	defer rows.Close()

	type openItem struct {
		id   uuid.UUID
		spec reconcileSpec
	}
	var open []openItem
	for rows.Next() {
		var id uuid.UUID
		var specJSON []byte
		if err := rows.Scan(&id, &specJSON); err != nil {
			logger.Warn("reconcile_section_data: scan failed", zap.Error(err))
			continue
		}
		var s reconcileSpec
		if err := json.Unmarshal(specJSON, &s); err != nil {
			logger.Warn("reconcile_section_data: spec unmarshal failed",
				zap.String("item_id", id.String()), zap.Error(err))
			continue
		}
		if pageFilter != "" && s.PageName != pageFilter {
			continue
		}
		open = append(open, openItem{id: id, spec: s})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate section-data items: %w", err)
	}

	scanned := len(open)
	resolvablePages := make(map[string]bool) // page_name → needs re-render
	skippedHuman := 0

	for _, item := range open {
		if len(item.spec.Missing) == 0 {
			continue
		}

		allQuery := true
		allResolved := true
		for _, m := range item.spec.Missing {
			if !strings.HasPrefix(m.Source, "query.") {
				// Non-query (human/spec) source — needs human input. Leave it.
				allQuery = false
				break
			}
			queryName := strings.TrimPrefix(m.Source, "query.")
			val, qerr := queryresolve.Resolve(ctx, params.DB, queryresolve.QueryRequest{
				Name:   queryName,
				SiteID: siteID,
			}, logger)
			if qerr != nil || !hasItems(val) {
				allResolved = false
			}
		}

		if !allQuery {
			skippedHuman++
			continue
		}
		if allResolved && item.spec.PageName != "" {
			resolvablePages[item.spec.PageName] = true
		}
	}

	// One re-render per resolvable page. Shared dedup key with
	// flag_page_image_rebuild (page_rerender:<page>) so concurrent triggers
	// collapse into a single re-render that re-resolves both image and section
	// data. plan_sections closes the needs_section_data item on re-render.
	var retriggered []string
	if len(resolvablePages) > 0 {
		tx, err := params.DB.BeginTx(ctx, nil)
		if err != nil {
			return nil, fmt.Errorf("begin tx: %w", err)
		}
		defer tx.Rollback()

		batchID := uuid.New()
		for page := range resolvablePages {
			spec := fmt.Sprintf(`{"reason":"section_data_resolved","page_name":%q}`, page)
			if _, err := insertWorkItem(ctx, tx, workItem{
				siteID:       siteID,
				source:       "section-data-reconciler",
				pipeline:     "build",
				itemType:     "needs_page",
				severity:     "medium",
				summary:      fmt.Sprintf("Re-render %s — deferred section data now resolvable", page),
				spec:         spec,
				priority:     99,
				handlerAgent: "page-build-handler",
				status:       "triaged",
				createdBy:    "section-data-reconciler",
				itemKey:      fmt.Sprintf("page_rerender:%s", page),
				batchID:      batchID,
			}, logger); err != nil {
				return nil, fmt.Errorf("emit needs_page for %s: %w", page, err)
			}
			retriggered = append(retriggered, page)
		}
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit: %w", err)
		}
	}

	logger.Info("reconcile_section_data: done",
		zap.Int("scanned", scanned),
		zap.Int("retriggered_pages", len(retriggered)),
		zap.Int("skipped_human", skippedHuman))

	return map[string]interface{}{
		"scanned":           scanned,
		"retriggered_pages": retriggered,
		"skipped_human":     skippedHuman,
	}, nil
}

// hasItems reports whether a queryresolve result is a non-empty list.
func hasItems(v interface{}) bool {
	switch t := v.(type) {
	case []map[string]interface{}:
		return len(t) > 0
	case []interface{}:
		return len(t) > 0
	default:
		return false
	}
}

```

### platform/orchestration/actions/registry.go (package `actions`) — whole file

```go
// platform/orchestration/actions/registry.go
package actions

import (
	"github.com/gqls/agentchassis/platform/orchestration/actioncheck"
	"go.uber.org/zap"
)

// GlobalActionRegistry is the single source of truth for all available actions.
// Each entry carries metadata (category, description, deprecation status) alongside the handler.
//
// Categories map to future subdirectory structure:
//
//	core     — workflow control primitives (complete, loop, conditional, await)
//	agent    — agent lifecycle (spawn, call, discover)
//	web      — web search, scraping, research
//	llm      — LLM prompt execution
//	site     — site building, assembly, rendering, pages, styles, git, navigation
//	data     — transform, validate, extract, aggregate, database queries
//	hitl     — human-in-the-loop approval and input
//	storage  — S3, assets, entity state, memory
//	image    — image generation and deployment
//	external — notifications, HTTP requests
var GlobalActionRegistry = map[string]ActionDefinition{

	// =========================================================================
	// CORE — workflow control primitives
	// =========================================================================
	"complete_workflow": {
		Handler:     CompleteWorkflowAction,
		Category:    "core",
		Description: "Signal workflow completion and notify parent",
		IsLocal:     true,
	},
	"await_response": {
		Handler:     AwaitResponseAction,
		Category:    "core",
		Description: "Pause workflow execution until an async response arrives",
		IsLocal:     true,
	},
	"evaluate_condition": {
		Handler:     EvaluateConditionAction,
		Category:    "core",
		Description: "Evaluate a condition expression against collected data",
		IsLocal:     true,
	},
	"loop": {
		Handler:     LoopAction,
		Category:    "core",
		Description: "Iterate over a collection, executing sub-steps for each item",
		IsLocal:     true,
	},
	"loop_complete": {
		Handler:     LoopCompleteAction,
		Category:    "core",
		Description: "Signal completion of the current loop iteration",
		IsLocal:     true,
	},
	"conditional_branch": {
		Handler:     ConditionalBranchAction,
		Category:    "core",
		Description: "Branch workflow based on a condition",
		IsLocal:     true,
	},
	"conditional": {
		Handler:      ConditionalBranchAction,
		Category:     "core",
		Description:  "Alias for conditional_branch",
		IsLocal:      true,
		Deprecated:   true,
		DeprecatedBy: "conditional_branch",
	},
	"conditional_route": {
		Handler:     ConditionalRouteAction,
		Category:    "core",
		Description: "Route to different next steps based on multiple conditions",
		IsLocal:     true,
	},

	// =========================================================================
	// AGENT — agent lifecycle and discovery
	// =========================================================================
	"spawn_agent": {
		Handler:     SpawnAgentAction,
		Category:    "agent",
		Description: "Spawn a new agent as a Kubernetes job",
		IsLocal:     true,
	},
	"spawn_group": {
		Handler:     SpawnGroupAction,
		Category:    "agent",
		Description: "Spawn a group of agents in parallel",
		IsLocal:     true,
	},
	"dispatch_agent": {
		Handler:     DispatchAgentAction,
		Category:    "agent",
		Description: "Dispatch agent creation to a remote cluster via Kafka - remote job spawner",
		IsLocal:     true,
	},
	"call_agent": {
		Handler:     CallAgentAction,
		Category:    "agent",
		Description: "Send a message to a spawned agent and await its response",
		IsLocal:     true,
	},
	"discover_agents": {
		Handler:     DiscoverAgentsAction,
		Category:    "agent",
		Description: "Query available agents by capability",
		IsLocal:     true,
	},
	"start_orchestration": {
		Handler:     StartOrchestrationAction,
		Category:    "agent",
		Description: "Start a new orchestration workflow",
		IsLocal:     true,
	},
	"discover_best_agents": {
		Handler:     DiscoverBestAgentsAction,
		Category:    "agent",
		Description: "Find best-performing agents for a given task type",
		IsLocal:     true,
	},
	"review_performance": {
		Handler:     ReviewPerformanceAction,
		Category:    "agent",
		Description: "Review an agent's performance metrics",
		IsLocal:     true,
	},
	"approve_improvement": {
		Handler:     ApproveImprovementAction,
		Category:    "agent",
		Description: "Approve a proposed agent improvement",
		IsLocal:     true,
	},
	"query_agent_definitions": {
		Handler:     QueryAgentDefinitionsAction,
		Category:    "agent",
		Description: "Query the agent_definitions table for agent configs",
		IsLocal:     true,
	},
	"evaluate_task": {
		Handler:     EvaluateTaskAction,
		Category:    "agent",
		Description: "Evaluate a task and determine required agents or actions",
		IsLocal:     true,
	},
	"spawn_agent_k8s": {
		Handler:      SpawnAgentAction,
		Category:     "agent",
		Description:  "Legacy alias for spawn_agent",
		IsLocal:      true,
		Deprecated:   true,
		DeprecatedBy: "spawn_agent",
	},

	// =========================================================================
	// IMAGE — generation and deployment
	// =========================================================================
	"generate_image": {
		Handler:     GenerateImageAction,
		Category:    "image",
		Description: "Generate an image via the image generation adapter",
		IsLocal:     true,
	},
	"store_generated_image": {
		Handler:     StoreGeneratedImageAction,
		Category:    "image",
		Description: "Store a generated image to object storage",
		IsLocal:     true,
	},
	"deploy_image_asset": {
		Handler:     DeployImageAssetAction,
		Category:    "image",
		Description: "Download image from storage and commit to git as a site asset",
		IsLocal:     true,
	},

	// =========================================================================
	// WEB — search, scraping, research
	// =========================================================================
	"web_search": {
		Handler:     WebSearchAction,
		Category:    "web",
		Description: "Perform a web search via the search adapter",
		IsLocal:     true,
	},
	"scrape_web": {
		Handler:     WebscrapeAction,
		Category:    "web",
		Description: "Scrape a web page for content",
		IsLocal:     true,
	},
	"firecrawl_scrape": {
		Handler:     FirecrawlScrapeAction,
		Category:    "web",
		Description: "Scrape a single page via Firecrawl",
		IsLocal:     true,
	},
	"firecrawl_crawl": {
		Handler:     FirecrawlCrawlAction,
		Category:    "web",
		Description: "Multi-page crawl via Firecrawl",
		IsLocal:     true,
	},
	"firecrawl_extract": {
		Handler:     FirecrawlExtractAction,
		Category:    "web",
		Description: "Structured data extraction via Firecrawl",
		IsLocal:     true,
	},
	"validate_url": {
		Handler:     ValidateURLAction,
		Category:    "web",
		Description: "Validate and normalise a URL",
		IsLocal:     true,
	},
	"aggregate_scraped_data": {
		Handler:     AggregateScrapedDataAction,
		Category:    "web",
		Description: "Combine results from multiple scrape operations",
		IsLocal:     true,
	},
	"split_urls": {
		Handler:     SplitURLsAction,
		Category:    "web",
		Description: "Split a list of URLs for parallel processing",
		IsLocal:     true,
	},
	"batch_webscrape": {
		Handler:     BatchWebscrapeAction,
		Category:    "web",
		Description: "Scrape multiple pages from a site sequentially",
		IsLocal:     true,
	},
	"scan_discovery_candidates": {
		Handler:     ScanDiscoveryCandidatesAction,
		Category:    "web",
		Description: "Extract new potential vet practices from scraped data",
		IsLocal:     true,
	},
	"prepare_urls": {
		Handler:     PrepareUrlsAction,
		Category:    "web",
		Description: "Prepare and normalise URLs for research scraping",
		IsLocal:     true,
	},
	"format_research_content": {
		Handler:     FormatResearchContentAction,
		Category:    "web",
		Description: "Format scraped content for use in research context",
		IsLocal:     true,
	},
	"prepare_extraction_context": {
		Handler:     PrepareExtractionContextAction,
		Category:    "web",
		Description: "Format scraped content for use in vet verifier",
		IsLocal:     true,
	},
	"process_area_sweep": {
		Handler:     ProcessAreaSweepAction,
		Category:    "web",
		Description: "Format scraped content for use in vet verifier",
		IsLocal:     true,
	},
	"load_unswept_areas": {
		Handler:     LoadUnsweptAreasAction,
		Category:    "web",
		Description: "Collect areas in UK for collecting vet practice details that havent yet been searched",
		IsLocal:     true,
	},
	"dispatch_area_discoverers": {
		Handler:     DispatchAreaDiscoverersAction,
		Category:    "web",
		Description: "Reads unswept areas from collected_data (output of load_unswept_areas). Produces one Kafka message per district to trigger area-sweep-discoverer agents.",
		IsLocal:     true,
	},

	// =========================================================================
	// LLM — prompt execution
	// =========================================================================
	"execute_llm_prompt": {
		Handler:     ExecuteLLMPromptAction,
		Category:    "llm",
		Description: "Execute an LLM prompt with templated input",
		IsLocal:     true,
	},
	"check_endpoint_health": {
		Handler:     CheckEndpointHealthAction,
		Category:    "system",
		Description: "Ping AI endpoints and update health table",
		IsLocal:     true,
	},

	// =========================================================================
	// DATA — transform, validate, extract, aggregate, database
	// =========================================================================
	"validate_input": {
		Handler:     ValidateInputAction,
		Category:    "data",
		Description: "Validate input data against expected schema",
		IsLocal:     true,
	},
	"transform_data": {
		Handler:     TransformDataAction,
		Category:    "data",
		Description: "Transform data between formats",
		IsLocal:     true,
	},
	"prepare_training_data": {
		Handler:     PrepareTrainingDataAction,
		Category:    "data",
		Description: "Stream training_exports.rows to NDJSON in S3 and INSERT a model_lifecycle.training_runs row in pending state",
		IsLocal:     true,
	},
	"validate_schema": {
		Handler:     ValidateSchemaAction,
		Category:    "data",
		Description: "Validate data against a JSON schema",
		IsLocal:     true,
	},
	"parse_json_field": {
		Handler:     ParseJSONFieldAction,
		Category:    "data",
		Description: "Parse a JSON string field into structured data",
		IsLocal:     true,
	},
	"extract_field": {
		Handler:     ExtractFieldAction,
		Category:    "data",
		Description: "Extract a single field from collected data by path",
		IsLocal:     true,
	},
	"extract_fields": {
		Handler:     ExtractFieldsAction,
		Category:    "data",
		Description: "Extract multiple fields from collected data",
		IsLocal:     true,
	},
	"collect_intent_events": {
		Handler:     CollectIntentEventsAction,
		Category:    "data",
		Description: "Pull new intent events from every VM-hosted backend site's /events endpoint into intent_events.",
		IsLocal:     true,
	},
	"calculate": {
		Handler:     CalculateAction,
		Category:    "data",
		Description: "Perform mathematical calculations",
		IsLocal:     true,
	},
	"aggregate_data": {
		Handler:     AggregateDataAction,
		Category:    "data",
		Description: "Aggregate data from multiple sources into a single result",
		IsLocal:     true,
	},
	"aggregate_webpage": {
		Handler:     AggregateWebpageAction,
		Category:    "data",
		Description: "Aggregate webpage content from multiple scrape results",
		IsLocal:     true,
	},
	"filter_search_results": {
		Handler:     FilterSearchResultsAction,
		Category:    "data",
		Description: "Filter and rank search results by relevance",
		IsLocal:     true,
	},
	"query_database": {
		Handler:     QueryDatabaseAction,
		Category:    "data",
		Description: "Execute a database query and return results",
		IsLocal:     true,
	},

	// =========================================================================
	// DATA — Business Intelligence (business_intel schema) (vet)
	// =========================================================================
	"load_business_record": {
		Handler:     LoadBusinessRecordAction,
		Category:    "data",
		Description: "Load a business record with optional vet details and prices",
		IsLocal:     true,
	},
	"store_business_verification": {
		Handler:     StoreBusinessVerificationAction,
		Category:    "data",
		Description: "Store verification results from agent run to business_intel tables",
		IsLocal:     true,
	},
	"load_business_batch": {
		Handler:     LoadBusinessBatchAction,
		Category:    "data",
		Description: "Claim and load next batch of pending collection tasks",
		IsLocal:     true,
	},
	"promote_candidates": {
		Handler:     PromoteCandidatesAction,
		Category:    "data",
		Description: "Move pending discovery candidates into businesses table with dedup",
		IsLocal:     true,
	},
	"ensure_collection_tasks": {
		Handler:     EnsureCollectionTasksAction,
		Category:    "data",
		Description: "Backfill collection_tasks for pending businesses missing them",
		IsLocal:     true,
	},
	"dispatch_verifiers": {
		Handler:     DispatchVerifiersAction,
		Category:    "web",
		Description: "Load pending-verification businesses and dispatch vet-practice-verifier agents via Kafka",
		IsLocal:     true,
	},

	// =========================================================================
	// COMPANIES HOUSE — enrichment, financials, ownership
	// =========================================================================
	"load_ch_enrichment_batch": {
		Handler:     LoadCHEnrichmentBatchAction,
		Category:    "data",
		Description: "Load verified businesses not yet enriched with Companies House data",
		IsLocal:     true,
	},
	"companies_house_search": {
		Handler:     CompaniesHouseSearchAction,
		Category:    "data",
		Description: "Search Companies House API by business name, match by postcode and SIC code",
		IsLocal:     true,
	},
	"companies_house_fetch": {
		Handler:     CompaniesHouseFetchAction,
		Category:    "data",
		Description: "Fetch company profile, officers, and PSC from Companies House",
		IsLocal:     true,
	},
	"store_ch_enrichment": {
		Handler:     StoreCHEnrichmentAction,
		Category:    "data",
		Description: "Store Companies House enrichment data and derive succession signals",
		IsLocal:     true,
	},
	"ch_bulk_collect": {
		Handler:     CHBulkCollectAction,
		Category:    "data",
		Description: "Bulk collect all CH companies with a given SIC code into local mirror table",
		IsLocal:     true,
	},
	"ch_local_match": {
		Handler:     CHLocalMatchAction,
		Category:    "data",
		Description: "Match verified businesses against local CH vet companies table by postcode + name similarity",
		IsLocal:     true,
	},
	"ch_llm_review": {
		Handler:     CHLLMReviewAction,
		Category:    "data",
		Description: "Review ambiguous CH matches using LLM judgment",
		IsLocal:     true,
	},
	"ch_fetch_accounts": {
		Handler:     CHFetchAccountsAction,
		Category:    "data",
		Description: "Fetch and parse filed accounts (iXBRL) for financial data",
		IsLocal:     true,
	},
	"ch_scrape_company_number": {
		Handler:     CHScrapeCompanyNumberAction,
		Category:    "data",
		Description: "Scrape business websites for company registration numbers",
		IsLocal:     true,
	},

	// =========================================================================
	// SITE — building, assembly, rendering, pages, styles, git, navigation
	// =========================================================================
	"git_commit": {
		Handler:     GitCommitAction,
		Category:    "site",
		Description: "Commit files to a git repository via GitHub API",
		IsLocal:     true,
	},
	"git_commit_action": {
		Handler:      GitCommitAction,
		Category:     "site",
		Description:  "Alias for git_commit",
		IsLocal:      true,
		Deprecated:   true,
		DeprecatedBy: "git_commit",
	},
	"assemble_from_library": {
		Handler:     AssembleFromLibraryAction,
		Category:    "site",
		Description: "Assemble a page from component library templates",
		IsLocal:     true,
	},
	"new_site_architect": {
		Handler:      AssembleFromLibraryAction,
		Category:     "site",
		Description:  "Alias for assemble_from_library",
		IsLocal:      true,
		Deprecated:   true,
		DeprecatedBy: "assemble_from_library",
	},
	"assemble_page": {
		Handler:     AssemblePageAction,
		Category:    "site",
		Description: "Assemble a single page from content and components",
		IsLocal:     true,
	},
	"assemble_multipage_site": {
		Handler:     AssembleMultipageSiteAction,
		Category:    "site",
		Description: "Assemble a complete multi-page site structure",
		IsLocal:     true,
	},
	"load_component_library": {
		Handler:     LoadComponentLibraryAction,
		Category:    "site",
		Description: "Load the HTML component library from the database",
		IsLocal:     true,
	},
	"plan_sections": {
		Handler:     PlanSectionsAction,
		Category:    "site",
		Description: "Resolve section data requirements and triage readiness",
		IsLocal:     true,
	},
	"fetch_agent_questionnaire": {
		Handler:     FetchAgentQuestionnaireAction,
		Category:    "site",
		Description: "Fetch the briefing questionnaire for an agent type",
		IsLocal:     true,
	},
	"load_site_for_design": {
		Handler:     LoadSiteForDesignAction,
		Category:    "site",
		Description: "Load site record and associated data for design operations",
		IsLocal:     true,
	},
	"load_site_for_rebuild": {
		Handler:     LoadSiteForRebuildAction,
		Category:    "site",
		Description: "Load site context for page rebuilds: brief, navigation, pages, brand assets from DB",
		IsLocal:     true,
	},
	"scan_sites_for_maintenance": {
		Handler:     ScanSitesForMaintenanceAction,
		Category:    "site",
		Description: "Scan deployed sites for maintenance issues and insert tasks into queue",
		IsLocal:     true,
	},
	"prepare_rebuild_dispatches": {
		Handler:     PrepareRebuildDispatchesAction,
		Category:    "site",
		Description: "Claim page_rebuild tasks from queue, flag pages, group dispatches by site",
		IsLocal:     true,
	},
	"mark_maintenance_complete": {
		Handler:     MarkMaintenanceCompleteAction,
		Category:    "site",
		Description: "Mark maintenance queue tasks as complete or failed after specialist finishes",
		IsLocal:     true,
	},
	// site specs (site_specs table — versioned spec storage)
	"write_site_spec": {
		Handler:     WriteSiteSpecAction,
		Category:    "site",
		Description: "Write or merge a versioned spec aspect to site_specs, superseding previous",
		IsLocal:     true,
	},
	"read_site_spec": {
		Handler:     ReadSiteSpecAction,
		Category:    "site",
		Description: "Read current site_specs for one aspect or all aspects",
		IsLocal:     true,
	},
	// component history (page_component_history table)
	"save_component_history": {
		Handler:     SaveComponentHistoryAction,
		Category:    "site",
		Description: "Snapshot current page_component content_data before overwrite",
		IsLocal:     true,
	},
	// build queue seeding
	"seed_build_queue": {
		Handler:     SeedBuildQueueAction,
		Category:    "site",
		Description: "Process queued build_queue entries into site records and first work items",
		IsLocal:     true,
	},
	// work item chaining
	"create_work_item": {
		Handler:     CreateWorkItemAction,
		Category:    "site",
		Description: "Insert a single work item for pipeline chaining between handler agents",
		IsLocal:     true,
	},
	"claim_work_item": {
		Handler:     ClaimWorkItemAction,
		Category:    "site",
		Description: "Atomically claim a work item for processing, preventing double-dispatch",
		IsLocal:     true,
	},
	"create_blog_posts": {
		Handler:     CreateBlogPostsAction,
		Category:    "site",
		Description: "Create page records and work items for LLM-planned blog posts",
		IsLocal:     true,
	},
	"rebuild_blog_listing": {
		Handler:     RebuildBlogListingAction,
		Category:    "site",
		Description: "Rebuild blog listing page_component from published posts",
		IsLocal:     true,
	},
	"ch_detail_fetch": {
		Handler:     CHDetailFetchAction,
		Category:    "data",
		Description: "Fetch profile, officers, PSC from CH API for confirmed matches",
		IsLocal:     true,
	},
	// work items (site_work_items table — unified build/maintenance queue)
	"write_build_items": {
		Handler:     WriteBuildItemsAction,
		Category:    "site",
		Description: "Convert planned pages into work items in site_work_items table",
		IsLocal:     true,
	},
	"apply_gap_plan": {
		Handler:     ApplyGapPlanAction,
		Category:    "site",
		Description: "Execute content gap plan — create pages, work items, or spec updates",
		IsLocal:     true,
	},
	"update_site_spec_from_item": {
		Handler:     UpdateSiteSpecFromItemAction,
		Category:    "site",
		Description: "Apply a spec update from a work item (needs_spec_update handler)",
		IsLocal:     true,
	},
	"load_work_items": {
		Handler:     LoadWorkItemsAction,
		Category:    "site",
		Description: "Load pending work items for a site, respecting dependencies",
		IsLocal:     true,
	},
	"complete_work_item": {
		Handler:     CompleteWorkItemAction,
		Category:    "site",
		Description: "Mark a work item as complete with result data and commit SHA",
		IsLocal:     true,
	},
	"fail_work_item": {
		Handler:     FailWorkItemAction,
		Category:    "site",
		Description: "Mark a work item as failed, increment attempt count for retry",
		IsLocal:     true,
	},
	"run_discovery_checks": {
		Handler:     RunDiscoveryChecksAction,
		Category:    "maintenance",
		Description: "Run discovery checks, write findings to site_work_items",
		IsLocal:     true,
	},
	"fix_nav_link_templates": {
		Handler:     FixNavLinkTemplatesAction,
		Category:    "maintenance",
		Description: "Fix broken nav links in header/footer templates (# → /page.html)",
		IsLocal:     true,
	},
	"fix_hardcoded_colors": {
		Handler:     FixHardcodedColorsAction,
		Category:    "maintenance",
		Description: "Replace hardcoded hex colors with CSS variables in component styles",
		IsLocal:     true,
	},
	"fork_theme_from_site": {
		Handler:     ForkThemeFromSiteAction,
		Category:    "site",
		Description: "Fork an adopted site's generated theme into the reusable theme library with HITL review",
		IsLocal:     true,
	},
	"link_site_components": {
		Handler:     LinkSiteComponentsAction,
		Category:    "site",
		Description: "Link site_components to content_components from style collection",
		IsLocal:     true,
	},
	"compute_component_quality": {
		Handler:     ComputeComponentQualityAction,
		Category:    "site",
		Description: "Score and store component quality metrics",
		IsLocal:     true,
	},
	"fix_component_template": {
		Handler:     FixComponentTemplateAction,
		Category:    "site",
		Description: "Apply targeted fixes to component HTML/CSS",
		IsLocal:     true,
	},
	"write_audit_findings": {
		Handler:     WriteAuditFindingsAction,
		Category:    "site",
		Description: "Convert LLM audit findings into site_work_items",
		IsLocal:     true,
	},
	"render_css_from_spec": {
		Handler:     RenderCSSFromSpecAction,
		Category:    "site",
		Description: "Render CSS from Go template + design spec (deterministic, no LLM)",
		IsLocal:     true,
	},
	"render_js_snippets_for_site": {
		Handler:     RenderJSSnippetsForSiteAction,
		Category:    "site",
		Description: "Render concatenated JS bundle from active js_snippets for a site",
		IsLocal:     true,
	},
	"sync_site_identity": {
		Handler:     SyncSiteIdentityAction,
		Category:    "data",
		Description: "Populate sites table columns from site_specs identity/briefing",
		IsLocal:     true,
	},
	"triage_detected_items": {
		Handler:     TriageDetectedItemsAction,
		Category:    "maintenance",
		Description: "Promote detected discovery items to triaged with target domain",
		IsLocal:     true,
	},
	"prepare_link_context": {
		Handler:     PrepareLinkContextAction,
		Category:    "site",
		Description: "Prepare internal link context for page rendering",
		IsLocal:     true,
	},
	"resolve_internal_links": {
		Handler:     ResolveInternalLinksAction,
		Category:    "content",
		Description: "Resolve intent-appropriate internal CTA destinations from real pages",
		IsLocal:     true,
	},
	"validate_page_content": {
		Handler:     ValidatePageContentAction,
		Category:    "site",
		Description: "Validate page content structure and completeness",
		IsLocal:     true,
	},
	"render_site_components": {
		Handler:     RenderSiteComponentsAction,
		Category:    "site",
		Description: "Render all components for a site using templates",
		IsLocal:     true,
	},
	"get_pages_for_rerender": {
		Handler:     GetPagesForRerenderAction,
		Category:    "site",
		Description: "Get list of pages that need re-rendering",
		IsLocal:     true,
	},
	"rerender_single_page": {
		Handler:     RerenderSinglePageAction,
		Category:    "site",
		Description: "Re-render a single page with updated content or styles",
		IsLocal:     true,
	},
	"create_rerender_items": {
		Handler:     CreateRerenderItemsAction,
		Category:    "site",
		Description: "Create per-page rerender work items for dispatch loop",
		IsLocal:     true,
	},
	"save_page_sections": {
		Handler:     SavePageSectionsAction,
		Category:    "site",
		Description: "Save page section data for subsequent rendering",
		IsLocal:     true,
	},
	"insert_research_result": {
		Handler:     InsertResearchResultAction,
		Category:    "site",
		Description: "Insert research results into site content data",
		IsLocal:     true,
	},
	"select_style_collection": {
		Handler:     SelectStyleCollectionAction,
		Category:    "site",
		Description: "Select a style collection for the site based on industry and preferences",
		IsLocal:     true,
	},
	"update_site_content": {
		Handler:     UpdateSiteContentAction,
		Category:    "site",
		Description: "Update the content_data field of a site record",
		IsLocal:     true,
	},
	"update_site_status": {
		Handler:     UpdateSiteStatusAction,
		Category:    "site",
		Description: "Update the build status of a site",
		IsLocal:     true,
	},
	"update_site_defaults": {
		Handler:     UpdateSiteDefaultsAction,
		Category:    "site",
		Description: "Update default settings for a site",
		IsLocal:     true,
	},
	"update_page_status": {
		Handler:     UpdatePageStatusAction,
		Category:    "site",
		Description: "Update the build status of a page",
		IsLocal:     true,
	},
	"update_work_item_status": {
		Handler:     UpdateWorkItemStatusAction,
		Category:    "site",
		Description: "Update status of a site_work_items row (complete, failed, etc) from inside a workflow",
		IsLocal:     true,
	},
	"build_render_context": {
		Handler:     BuildRenderContextAction,
		Category:    "site",
		Description: "Build the template rendering context for a page",
		IsLocal:     true,
	},
	"render_component": {
		Handler:     RenderComponentAction,
		Category:    "site",
		Description: "Render a single component with its template and data",
		IsLocal:     true,
	},
	"compile_page_sections": {
		Handler:     CompilePageSectionsAction,
		Category:    "site",
		Description: "Compile all sections of a page into final HTML",
		IsLocal:     true,
	},
	"db_sync": {
		Handler:     DBSyncAction,
		Category:    "site",
		Description: "Synchronise in-memory state with the database",
		IsLocal:     true,
	},
	"store_asset": {
		Handler:     StoreAssetAction,
		Category:    "site",
		Description: "Store a site asset (image, file) to the assets table",
		IsLocal:     true,
	},
	"ensure_site_record": {
		Handler:     EnsureSiteRecordAction,
		Category:    "site",
		Description: "Create or update the site record in the database",
		IsLocal:     true,
	},
	"sync_pages_to_db": {
		Handler:     SyncPagesToDBAction,
		Category:    "site",
		Description: "Synchronise planned pages to the pages table",
		IsLocal:     true,
	},
	"get_pages_to_build": {
		Handler:     GetPagesToBuildAction,
		Category:    "site",
		Description: "Query pages that still need building",
		IsLocal:     true,
	},
	"extract_and_sync_links": {
		Handler:     ExtractAndSyncLinksAction,
		Category:    "site",
		Description: "Extract internal links from content and sync to navigation",
		IsLocal:     true,
	},
	"update_site_timestamps": {
		Handler:     UpdateSiteTimestampsAction,
		Category:    "site",
		Description: "Update last_built_at and related timestamps on the site record",
		IsLocal:     true,
	},
	"get_navigation_from_db": {
		Handler:     GetNavigationFromDBAction,
		Category:    "site",
		Description: "Load navigation structure from the database",
		IsLocal:     true,
	},
	"populate_nav_tables": {
		Handler:     PopulateNavTablesAction,
		Category:    "site",
		Description: "Populate site_nav_groups and site_nav_items from page records",
		IsLocal:     true,
	},
	"validate_site_plan": {
		Handler:     ValidateSitePlanAction,
		Category:    "site",
		Description: "Validate a site plan structure before building",
		IsLocal:     true,
	},
	"write_site_plan": {
		Handler:     WriteSitePlanAction,
		Category:    "site",
		Description: "Write the validated plan to site_plans + site_plan_pages + site_plan_partials in one tx (doc 030)",
		IsLocal:     true,
	},
	"reconcile_site_plan": {
		Handler:     ReconcileSitePlanAction,
		Category:    "site",
		Description: "Diff site_plan_pages vs realised pages; emit needs_page work items for the delta; emit terminal needs_rerender if any (doc 030)",
		IsLocal:     true,
	},
	"reconcile_section_data": {
		Handler:     ReconcileSectionDataAction,
		Category:    "site",
		Description: "Re-trigger pages whose deferred section data is now query-resolvable",
		IsLocal:     true,
	},
	"emit_design_items": {
		Handler:     EmitDesignItemsAction,
		Category:    "site",
		Description: "Plan-time: queue needs_composition + needs_design for a build (guarded on style_collection_id IS NULL)",
		IsLocal:     true,
	},
	"emit_imagery_items": {
		Handler:     EmitImageryItemsAction,
		Category:    "site",
		Description: "Plan-time: queue needs_imagery from the current plan's site_plan_imagery rows",
		IsLocal:     true,
	},
	"flag_page_image_rebuild": {
		Handler:     FlagPageImageRebuildAction,
		Category:    "site",
		Description: "Re-render a page after its image asset lands so the hero resolves",
		IsLocal:     true,
	},
	"generate_html": {
		Handler:     GenerateHTMLAction,
		Category:    "site",
		Description: "Generate HTML from structured content",
		IsLocal:     true,
	},
	"process_html": {
		Handler:     ProcessHTMLAction,
		Category:    "site",
		Description: "Post-process HTML (minify, clean, fix)",
		IsLocal:     true,
	},
	"validate_html": {
		Handler:     ValidateHTMLAction,
		Category:    "site",
		Description: "Validate HTML structure and correctness",
		IsLocal:     true,
	},
	"load_edit_context": {
		Handler:     LoadEditContextAction,
		Category:    "site",
		Description: "Load context for editing a page section",
		IsLocal:     true,
	},
	"apply_section_edit": {
		Handler:     ApplySectionEditAction,
		Category:    "site",
		Description: "Apply an edit to a page section and reassemble page",
		IsLocal:     true,
	},
	"load_page_sections_from_spec": {
		Handler:     LoadPageSectionsFromSpecAction,
		Category:    "site",
		Description: "Load page sections from site_specs with fallback to pages table",
		IsLocal:     true,
	},

	// Site Composition
	"validate_composition_inputs": {
		Handler:     ValidateCompositionInputsAction,
		Category:    "site",
		Description: "Check identity + classification specs exist, queue classifier if not",
		IsLocal:     true,
	},
	"resolve_composition_layout": {
		Handler:     ResolveCompositionLayoutAction,
		Category:    "site",
		Description: "Pick a library layout by tag-overlap against classification",
		IsLocal:     true,
	},
	"resolve_composition_typography": {
		Handler:     ResolveCompositionTypographyAction,
		Category:    "site",
		Description: "Pick a typography_set by font-family match with spec cascade",
		IsLocal:     true,
	},
	"install_site_composition": {
		Handler:     InstallSiteCompositionAction,
		Category:    "site",
		Description: "Install composition into css_themes + style_collections + resolved_composition spec",
		IsLocal:     true,
	},
	// --- Site composition (site-design-planner) ---
	"resolve_composition_palette": {
		Handler:     ResolveCompositionPaletteAction,
		Category:    "site",
		Description: "Extract palette colours via priority cascade and create palette row",
		IsLocal:     true,
	},

	// layout taxonomy — expose current categories + industry_tags to prompts
	"read_layout_taxonomy": {
		Handler:     ReadLayoutTaxonomyAction,
		Category:    "site",
		Description: "Read current categories and industry_tags from the layouts library for prompt templates",
		IsLocal:     true,
	},

	// RAG — retrieval-augmented generation
	"rag_lookup": {
		Handler:     RAGLookupAction,
		Category:    "storage",
		Description: "Search the knowledge base for relevant content using vector similarity",
		IsLocal:     true,
	},
	"rag_index": {
		Handler:     RAGIndexAction,
		Category:    "storage",
		Description: "Chunk, embed, and store content in the knowledge base",
		IsLocal:     true,
	},
	"training_data_export": {
		Handler:     TrainingDataExportAction,
		Category:    "storage",
		Description: "Export successful LLM calls from llm_call_log as NDJSON training data (ChatML format) for fine-tuning.",
		IsLocal:     true,
	},

	// ========================================================================
	// DOCUMENT AND CODE ANALYSIS — code-context retrieval (analyser adapter + code_symbols)
	// ========================================================================
	"lookup_code_symbols": {
		Handler:     LookupCodeSymbolsAction,
		Category:    "storage",
		Description: "Retrieve relevant code symbols from code_symbols (vector, trigram fallback)",
		IsLocal:     true,
	},
	"index_code_symbols": {
		Handler:     IndexCodeSymbolsAction,
		Category:    "storage",
		Description: "Upsert analysed symbols into code_symbols; embed changed, prune by commit",
		IsLocal:     true,
	},
	"request_repo_analysis": {
		Handler:     RequestRepoAnalysisAction,
		Category:    "code",
		Description: "Ask the analyser adapter to parse a repo at ref; awaits the symbol output",
		IsLocal:     true,
	},

	// =========================================================================
	// FEED — content feed ingestion pipeline
	// =========================================================================
	"fetch_rss": {
		Handler:     FetchRSSAction,
		Category:    "feed",
		Description: "Fetch and parse RSS/Atom feed, return normalised items",
		IsLocal:     true,
	},
	"fetch_llm_news": {
		Handler:     FetchLLMNewsAction,
		Category:    "feed",
		Description: "Fetch news via LLM API (xAI/Grok, OpenAI, Perplexity)",
		IsLocal:     true,
	},
	"write_feed_items": {
		Handler:     WriteFeedItemsAction,
		Category:    "feed",
		Description: "Normalise and write feed items to content_feed_items with dedup",
		IsLocal:     true,
	},
	"load_due_sources": {
		Handler:     LoadDueSourcesAction,
		Category:    "feed",
		Description: "Query content_sources for sources due to be fetched",
		IsLocal:     true,
	},
	"update_source_timestamps": {
		Handler:     UpdateSourceTimestampsAction,
		Category:    "feed",
		Description: "Update last_fetched_at/next_fetch_at after ingestion",
		IsLocal:     true,
	},
	"normalize_to_feed_items": {
		Handler:     NormalizeToFeedItemsAction,
		Category:    "feed",
		Description: "Transform web_search or firecrawl results into normalised feed items",
		IsLocal:     true,
	},
	"dispatch_feed_sources": {
		Handler:     DispatchFeedSourcesAction,
		Category:    "feed",
		Description: "Query due content_sources and dispatch feed-ingester per source",
		IsLocal:     true,
	},
	"fetch_scrape": {
		Handler:     FetchScrapeAction,
		Category:    "feed",
		Description: "Read URL from source_config, delegate to WebscrapeAction (async via adapter)",
		IsLocal:     true,
	},
	"fetch_news_search": {
		Handler:     FetchNewsSearchAction,
		Category:    "feed",
		Description: "Read query from source_config, delegate to WebSearchAction with search_type=news (async via adapter)",
		IsLocal:     true,
	},

	// tool lifecycle (deploy, update)
	"deploy_tool_to_site": {
		Handler:     DeployToolToSiteAction,
		Category:    "site",
		Description: "Fork a library tool into a site-owned copy, create tool page and page_component",
		IsLocal:     true,
	},
	"update_component_html": {
		Handler:     UpdateComponentHTMLAction,
		Category:    "site",
		Description: "Update html_template of a content_component with optional version snapshot",
		IsLocal:     true,
	},
	"create_tool_component": {
		Handler:     CreateToolComponentAction,
		Category:    "site",
		Description: "Create a new tool component from generated HTML and set up its page",
		IsLocal:     true,
	},
	"check_tool_completeness": {
		Handler:     CheckToolCompletenessAction,
		Category:    "validation",
		Description: "Verify LLM tool output is complete and not truncated",
		IsLocal:     true,
	},
	"create_tool_cross_link_items": {
		Handler:     CreateToolCrossLinkItemsAction,
		Category:    "site",
		Description: "Create content_rewrite items to cross-link tools from related pages",
		IsLocal:     true,
	},
	"apply_feed_scores": {
		Handler:     ApplyFeedScoresAction,
		Category:    "feed",
		Description: "Update content_feed_items with LLM relevance scores and status",
		IsLocal:     true,
	},
	"load_feed_items_for_triage": {
		Handler:     LoadFeedItemsForTriageAction,
		Category:    "feed",
		Description: "Load unscored ingested items with source metadata for triage",
		IsLocal:     true,
	},
	"render_news_section": {
		Handler:     RenderNewsSectionAction,
		Category:    "feed",
		Description: "Render latest-news component from content_feed_items",
		IsLocal:     true,
	},
	"evaluate_news_feed": {
		Handler:     EvaluateNewsFeedAction,
		Category:    "feed",
		Description: "Post-classification enrichment: determine if site should have news feed",
		IsLocal:     true,
	},
	"seed_content_sources": {
		Handler:     SeedContentSourcesAction,
		Category:    "feed",
		Description: "Create content_sources from classification news_feed recommendation",
		IsLocal:     true,
	},
	// training
	"save_tool_training_data": {
		Handler:     SaveToolTrainingDataAction,
		Category:    "training",
		Description: "Save tool recreation triple for model fine-tuning",
		IsLocal:     true,
	},
	"dispatch_thunder_decommission": {
		Handler:     DispatchThunderDecommissionAction,
		Category:    "training",
		Description: "Publish a decommission_instance request to thunder-adapter and await the response. Used by thunder-reaper and any other workflow terminating a Thunder Compute instance.",
		IsLocal:     true,
	},
	"dispatch_thunder_provision": {
		Handler:     DispatchThunderProvisionAction,
		Category:    "training",
		Description: "Publish a provision_instance request to thunder-adapter and await the response with the running instance details (instance_ip, ssh_user, ssh_key_secret_name, provisioning_id, thunder_identifier, provisioned_at). Used by gpu-provisioner.",
		IsLocal:     true,
	},
	"dispatch_thunder_ssh_exec": {
		Handler:     DispatchThunderSSHExecAction,
		Category:    "training",
		Description: "Publish an ssh_exec request to thunder-adapter (run a command on a provisioned instance by provisioning_id) and await the response with exit_code/stdout/stderr/reachable. Used by training-launcher to fetch data and launch the backgrounded training process.",
		IsLocal:     true,
	},

	"dispatch_thunder_prepare_object_url": {
		Handler:     DispatchThunderPrepareObjectURLAction,
		Category:    "training",
		Description: "Publish a prepare_object_url request to thunder-adapter to presign a B2 object by explicit key (GET or PUT) and await the presigned URL. Used by training-launcher to presign the dataset and the training scripts; the adapter remains the B2 credential boundary.",
		IsLocal:     true,
	},

	"dispatch_thunder_prepare_object_urls": {
		Handler:     DispatchThunderPrepareObjectURLsAction,
		Category:    "training",
		Description: "Publish a BATCH prepare_object_urls request to thunder-adapter to presign many B2 keys (same method/expiry) in ONE awaited round-trip; returns ordered presigned_urls[] aligned 1:1 with the input keys. Replaces training-launcher's per-checkpoint presign loop + flatten (the loop re-persisted the expanded workflow each iteration, O(K^2)); the adapter remains the B2 credential boundary.",
		IsLocal:     true,
	},

	"dispatch_thunder_prepare_resume_url": {
		Handler:     DispatchThunderPrepareResumeURLAction,
		Category:    "training",
		Description: "Publish a prepare_resume_url request to thunder-adapter: list the run's checkpoints in B2, presign a GET for the latest, and reply {found, presigned_url, key, index}. found=false means no checkpoints yet (fresh start). Lets the launcher resume an interrupted run from its latest checkpoint; the adapter is the B2 credential boundary.",
		IsLocal:     true,
	},

	"mark_training_run_running": {
		Handler:     MarkTrainingRunRunningAction,
		Category:    "training",
		Description: "Transition a model_lifecycle.training_runs row from pending to running and stamp started_at (sibling of the failed-marker in prepare_training_data). Used by training-launcher after the training process is launched.",
		IsLocal:     true,
	},
	"dispatch_thunder_ssh_get_status": {
		Handler:     DispatchThunderSSHGetStatusAction,
		Category:    "training",
		Description: "Publish an ssh_get_status request to thunder-adapter (probe reachability + run an optional status_command on a provisioned instance by provisioning_id) and await reachable/exit_code/stdout/stderr. Used by thunder-training-monitor to poll whether a detached training run is alive or finished.",
		IsLocal:     true,
	},
	"classify_training_probe": {
		Handler:     ClassifyTrainingProbeAction,
		Category:    "training",
		Description: "Classify a prior ssh_get_status probe (training-monitor) into a verdict (alive/done_ok/done_fail/gone_unknown/unreachable/no_status) and route the workflow via a next_step override. Pure logic; no DB.",
		IsLocal:     true,
	},
	"record_probe_streak": {
		Handler:     RecordProbeStreakAction,
		Category:    "training",
		Description: "Maintain the per-instance consecutive-unreachable counter on thunder_instances (mode=bump|reset) for thunder-training-monitor; route to lost_step once unreachable_threshold is reached. Requires migration 106.",
		IsLocal:     true,
	},
	"mark_training_run_terminal": {
		Handler:     MarkTrainingRunTerminalAction,
		Category:    "training",
		Description: "Transition a model_lifecycle.training_runs row running→complete|failed (config status); stamps completed_at and (on failed) error_message. Idempotent: only transitions rows still 'running'. Used by thunder-training-monitor's mark_complete/mark_failed steps.",
		IsLocal:     true,
	},
	"find_active_training_instances": {
		Handler:     FindActiveTrainingInstancesAction,
		Category:    "training",
		Description: "Query clients_db for running Thunder instances with a training_run_id and no decommission requested; returns {instances:[{provisioning_id,training_run_id,thunder_instance_id,instance_ip}], count} for thunder-training-monitor's loop fan-out.",
		IsLocal:     true,
	},
	"assemble_upload_manifest": {
		Handler:     AssembleUploadManifestAction,
		Category:    "training",
		Description: "Pure: build /workspace/upload_manifest.json content (and a base64 form for an ssh_exec base64 -d write) from the checkpoint keys + presigned PUT/GET URLs. Pairs checkpoint_keys[i] with checkpoint_urls[i] (same order) into the shape 02_train's --upload-manifest consumes; errors on a key/url length mismatch.",
		IsLocal:     true,
	},

	"compute_checkpoint_keys": {
		Handler:     ComputeCheckpointKeysAction,
		Category:    "training",
		Description: "Pure: produce the list of B2 checkpoint keys (finetuning/checkpoints/<run>/ckpt-<i>.tar.gz) plus the final-adapter key for the Phase 5 launcher to presign. K = ceil(max_steps/save_steps)+buffer when both are known, else a config fallback (default 64). Clamped to [1,512].",
		IsLocal:     true,
	},
	"flatten_presign_results": {
		Handler:     FlattenPresignResultsAction,
		Category:    "training",
		Description: "Pure connector: reshape the checkpoint presign loop's loop_complete results array into flat, ordered, same-length checkpoint_urls[] and checkpoint_keys[] for assemble_upload_manifest. Keeps assemble loop-agnostic; errors on any element missing url/key.",
		IsLocal:     true,
	},

	// =========================================================================
	// Site Snapshots
	// =========================================================================
	"take_site_snapshot": {
		Handler:     TakeSiteSnapshotAction,
		Category:    "site",
		Description: "Capture full site state (specs, pages, components, nav) into a snapshot",
		IsLocal:     true,
	},
	"revert_site_snapshot": {
		Handler:     RevertSiteSnapshotAction,
		Category:    "site",
		Description: "Restore a site to a previous snapshot state",
		IsLocal:     true,
	},
	"list_site_snapshots": {
		Handler:     ListSiteSnapshotsAction,
		Category:    "site",
		Description: "List available snapshots for a site",
		IsLocal:     true,
	},

	// =========================================================================
	// Adoption
	// =========================================================================
	"apply_adoption_plan": {
		Handler:     ApplyAdoptionPlanAction,
		Category:    "site",
		Description: "Create specs, pages, and work items from site adoption analysis",
		IsLocal:     true,
	},
	"extract_design_fingerprint": {
		Handler:     ExtractDesignFingerprintAction,
		Category:    "analysis",
		Description: "Extract concrete design data (colours, fonts, layout) from crawled HTML",
		IsLocal:     true,
	},
	"extract_interactive_fingerprint": {
		Handler:     ExtractInteractiveFingerprintAction,
		Category:    "analysis",
		Description: "Extract interactive element signals (scripts, canvas, forms) from crawled HTML",
		IsLocal:     true,
	},
	"format_crawl_for_analysis": {
		Handler:     FormatCrawlForAnalysisAction,
		Category:    "web",
		Description: "Format Firecrawl crawl output into readable text for LLM analysis",
		IsLocal:     true,
	},
	"load_existing_content": {
		Handler:     LoadExistingContentAction,
		Category:    "site",
		Description: "Load existing page content from research_results for recreate mode",
		IsLocal:     true,
	},
	"select_representative_content": {
		Handler:     SelectRepresentativeContentAction,
		Category:    "site",
		Description: "To select representative prose-heavy pages from crawl for style analysis",
		IsLocal:     true,
	},
	"enrich_fingerprint_with_css": {
		Handler:     EnrichFingerprintWithCSSAction,
		Category:    "analysis",
		Description: "Parse fetched CSS and merge into design fingerprint",
		IsLocal:     true,
	},
	"firecrawl_map": {
		Handler:     FirecrawlMapAction,
		Category:    "webscrape",
		Description: "Discover site URLs via firecrawl /map endpoint (no content fetched)",
		IsLocal:     false,
	},
	"prepare_scrape_batches": {
		Handler:     PrepareScrapeBatchesAction,
		Category:    "webscrape",
		Description: "Split discovered URLs into batches for paginated scraping",
		IsLocal:     true,
	},
	"store_crawl_batch": {
		Handler:     StoreCrawlBatchAction,
		Category:    "site",
		Description: "Store batch of scraped pages to research_results for paginated crawl",
		IsLocal:     true,
	},

	// =========================================================================
	// New Components
	// =========================================================================
	"store_generated_component": {
		Handler:     StoreGeneratedComponentAction,
		Category:    "site",
		Description: "Store a generated component template in the component library",
		IsLocal:     true,
	},

	// =========================================================================
	// VET MED PRICING — veterinary medicine price collection
	// =========================================================================
	"med_scrape_prices": {
		Handler:     MedScrapePricesAction,
		Category:    "med_pricing",
		Description: "Scrape veterinary medicine prices from retailer product pages via Firecrawl",
		IsLocal:     true,
	},
	"med_discover_urls": {
		Handler:     MedDiscoverURLsAction,
		Category:    "med_pricing",
		Description: "Discover product URLs from retailer category pages",
		IsLocal:     true,
	},
	"med_map_urls": {
		Handler:     MedMapURLsAction,
		Category:    "med_pricing",
		Description: "Discover product URLs via Firecrawl /map endpoint (site-wide)",
		IsLocal:     true,
	},

	// =========================================================================
	// HITL — human-in-the-loop approval and input
	// =========================================================================
	"await_approval": {
		Handler:     AwaitApprovalAction,
		Category:    "hitl",
		Description: "Pause workflow and wait for human approval",
		IsLocal:     true,
	},
	"process_approval_decision": {
		Handler:     ProcessApprovalDecisionAction,
		Category:    "hitl",
		Description: "Process the result of an approval decision",
		IsLocal:     true,
	},
	"process_data": {
		Handler:      ProcessApprovalDecisionAction,
		Category:     "hitl",
		Description:  "Legacy alias for process_approval_decision",
		IsLocal:      true,
		Deprecated:   true,
		DeprecatedBy: "process_approval_decision",
	},
	"create_approval_request": {
		Handler:     CreateApprovalRequestAction,
		Category:    "hitl",
		Description: "Create an approval request for human review",
		IsLocal:     true,
	},
	"wait_for_approval_response": {
		Handler:     WaitForApprovalResponseAction,
		Category:    "hitl",
		Description: "Wait for a previously created approval request to be resolved",
		IsLocal:     true,
	},
	"request_human_input": {
		Handler:     RequestHumanInputAction,
		Category:    "hitl",
		Description: "Request free-form input from a human operator",
		IsLocal:     true,
	},
	"process_human_input_response": {
		Handler:     ProcessHumanInputResponseAction,
		Category:    "hitl",
		Description: "Process the response from a human input request",
		IsLocal:     true,
	},
	"build_review_result": {
		Handler:     BuildReviewResultAction,
		Category:    "hitl",
		Description: "Build a structured review result from review data",
		IsLocal:     true,
	},
	"prepare_review_data": {
		Handler:     PrepareReviewDataAction,
		Category:    "hitl",
		Description: "Prepare content for human or automated review",
		IsLocal:     true,
	},
	"update_page_components_status": {
		Handler:     UpdatePageComponentsStatusAction,
		Category:    "hitl",
		Description: "Update component approval status after review",
		IsLocal:     true,
	},
	"load_page_section_components": {
		Handler:     LoadPageSectionComponentsAction,
		Category:    "hitl",
		Description: "Load page section components for review",
		IsLocal:     true,
	},
	"load_page_record": {
		Handler:     LoadPageRecordAction,
		Category:    "site",
		Description: "Load a page record by site_id and page name",
		IsLocal:     true,
	},
	"load_site_pages": {
		Handler:     LoadSitePagesAction,
		Category:    "site",
		Description: "Load all pages for a site",
		IsLocal:     true,
	},

	// =========================================================================
	// STORAGE — S3, assets, entity state, memory
	// =========================================================================
	"upload_to_s3": {
		Handler:     UploadToS3Action,
		Category:    "storage",
		Description: "Upload a file to S3-compatible object storage",
		IsLocal:     true,
	},
	"s3_upload": {
		Handler:      UploadToS3Action,
		Category:     "storage",
		Description:  "Alias for upload_to_s3",
		IsLocal:      true,
		Deprecated:   true,
		DeprecatedBy: "upload_to_s3",
	},
	"validate_assets": {
		Handler:     ValidateAssetsAction,
		Category:    "storage",
		Description: "Validate that required assets exist and are accessible",
		IsLocal:     true,
	},
	"deploy_to_hosting": {
		Handler:     DeployToHostingAction,
		Category:    "storage",
		Description: "Deploy files to hosting provider",
		IsLocal:     true,
	},
	"store_result": {
		Handler:     StoreResultAction,
		Category:    "storage",
		Description: "Store an action result to persistent storage",
		IsLocal:     true,
	},
	"route_storage": {
		Handler:     RouteStorageAction,
		Category:    "storage",
		Description: "Route data to the appropriate storage backend",
		IsLocal:     true,
	},
	"append_entity_state": {
		Handler:     AppendEntityStateAction,
		Category:    "storage",
		Description: "Append a new state entry to an entity's history",
		IsLocal:     true,
	},
	"read_latest_entity_state": {
		Handler:     ReadLatestEntityStateAction,
		Category:    "storage",
		Description: "Read the most recent state of an entity",
		IsLocal:     true,
	},
	"read_entity_history": {
		Handler:     ReadEntityHistoryAction,
		Category:    "storage",
		Description: "Read the full state history of an entity",
		IsLocal:     true,
	},
	"read_my_state": {
		Handler:     ReadMyStateAction,
		Category:    "storage",
		Description: "Read the current agent's persisted state",
		IsLocal:     true,
	},
	"write_my_state": {
		Handler:     WriteMyStateAction,
		Category:    "storage",
		Description: "Write the current agent's state to persistent storage",
		IsLocal:     true,
	},
	"retrieve_memory": {
		Handler:     RetrieveMemoryAction,
		Category:    "storage",
		Description: "Retrieve a stored memory by key",
		IsLocal:     true,
	},
	"store_memory": {
		Handler:     StoreMemoryAction,
		Category:    "storage",
		Description: "Store a value to the memory system",
		IsLocal:     true,
	},
	"cache_lookup": {
		Handler:     CacheLookupAction,
		Category:    "storage",
		Description: "Look up a value in the cache",
		IsLocal:     true,
	},

	// =========================================================================
	// EXTERNAL — notifications, HTTP
	// =========================================================================
	"send_notification": {
		Handler:     SendNotificationAction,
		Category:    "external",
		Description: "Send a notification via configured channel",
		IsLocal:     true,
	},
	"http_request": {
		Handler:     HTTPRequestAction,
		Category:    "external",
		Description: "Make an HTTP request to an external endpoint",
		IsLocal:     true,
	},
}

// deprecationLogger is set once at startup to avoid repeated logger creation.
// If nil, deprecation warnings are silently skipped.
var deprecationLogger *zap.Logger

// SetDeprecationLogger allows the application to provide a logger for deprecation warnings.
// Call this once during startup (e.g. in main.go or agent init).
func SetDeprecationLogger(logger *zap.Logger) {
	deprecationLogger = logger
}

// GetAction returns the action handler function if it exists.
// Signature is unchanged from the previous registry — coordinator does not need updating.
func GetAction(action string) (ActionFunc, bool) {
	def, exists := GlobalActionRegistry[action]
	if !exists {
		return nil, false
	}
	if def.Deprecated && deprecationLogger != nil {
		deprecationLogger.Warn("Deprecated action used",
			zap.String("action", action),
			zap.String("replacement", def.DeprecatedBy),
		)
	}
	return def.Handler, true
}

// GetActionDefinition returns the full definition including metadata.
// Useful for documentation, tooling, and introspection.
func GetActionDefinition(action string) (ActionDefinition, bool) {
	def, exists := GlobalActionRegistry[action]
	return def, exists
}

// IsLocalAction checks if an action is available for local execution.
// Replaces the previous delegation to actions_list package.
func IsLocalAction(action string) bool {
	def, exists := GlobalActionRegistry[action]
	return exists && def.IsLocal
}

// RegisterLocalActionChecker sets the function used to check if an action is local.
func init() {
	// Register the local action checker so other packages can check without importing actions
	actioncheck.RegisterLocalActionChecker(IsLocalAction)
}

// ListActions returns all available non-deprecated action names.
func ListActions() []string {
	actions := make([]string, 0, len(GlobalActionRegistry))
	for name, def := range GlobalActionRegistry {
		if !def.Deprecated {
			actions = append(actions, name)
		}
	}
	return actions
}

// ListAllActions returns all action names including deprecated ones.
func ListAllActions() []string {
	actions := make([]string, 0, len(GlobalActionRegistry))
	for name := range GlobalActionRegistry {
		actions = append(actions, name)
	}
	return actions
}

// ListActionsByCategory returns action names grouped by category, excluding deprecated.
func ListActionsByCategory() map[string][]string {
	result := make(map[string][]string)
	for name, def := range GlobalActionRegistry {
		if !def.Deprecated {
			result[def.Category] = append(result[def.Category], name)
		}
	}
	return result
}

// ListDeprecatedActions returns all deprecated actions with their replacements.
func ListDeprecatedActions() map[string]string {
	result := make(map[string]string)
	for name, def := range GlobalActionRegistry {
		if def.Deprecated {
			result[name] = def.DeprecatedBy
		}
	}
	return result
}

```

---

## Neighbourhood (signatures)

**Calls (callees)**

```go
func BuildAssetPaths(purpose string, extension string) AssetPaths  // platform/storage/url_helpers.go
func (m *Manager) Close() error  // platform/infrastructure/connections.go
func (c *Consumer) Close() error  // platform/kafka/consumer.go
func (m *MockProducer) Close() error  // platform/kafka/mock_producer.go
func (p *KafkaProducer) Close() error  // platform/kafka/producer.go
func (p *pgxRowScanner) Close() error  // platform/orchestration/actions/query_agent_definitions_actions.go
func (m *MockProducer) Close() error  // test/unit/helpers/test_helpers.go
func (s PageURLSet) Contains(href string) bool  // platform/orchestration/datahelpers/links.go
func (p *pgxRowScanner) Err() error  // platform/orchestration/actions/query_agent_definitions_actions.go
func (e *APIError) Error() string  // internal/adapters/imagegenerator/banana/api/client.go
func (e *APIError) Error() string  // internal/adapters/imagegenerator/stability/api/client.go
func (e *APIError) Error() string  // internal/adapters/thunder/api/client.go
func (e *AgentError) Error() string  // platform/errors/errors.go
func (e *AIUnavailableError) Error() string  // platform/orchestration/actions/ai_errors.go
func (s *SSHExecAction) Exec(ctx context.Context, req SSHExecRequest) (*SSHExecResult, error)  // internal/adapters/thunder/ssh_exec_actions.go
func ExtractActionInputs(collectedData map[string]interface{}, config map[string]interface{}, spec ActionInputSpec, logger *zap.Logger) (*ActionInputs, error)  // platform/orchestration/datahelpers/action_inputs.go
func ExtractNestedField(data map[string]interface{}, fieldPath string) interface{}  // platform/orchestration/datahelpers/data_helpers.go
func ExtractNestedFieldString(data map[string]interface{}, fieldPath string) string  // platform/orchestration/datahelpers/data_helpers.go
func ExtractSectionNamesHelper(sectionsRaw interface{}) []string  // platform/orchestration/datahelpers/data_helpers.go
func ExtractStringListHelper(val interface{}) []string  // platform/orchestration/datahelpers/data_helpers.go
func FindResultsArray(collectedData map[string]interface{}, basePath string, logger *zap.Logger) []interface{}  // platform/orchestration/datahelpers/data_helpers.go
func (s *Store) Get(id string) (*Order, bool)  // docs/agent_docs/docs024_key_docs_latest/idea.uk/golang_files/store.go
func Get(name string) DiscoveryCheck  // platform/orchestration/actions/discovery_checks/registry.go
func (ai *ActionInputs) Get(key string) string  // platform/orchestration/datahelpers/action_inputs.go
func GetComponentByID(ctx context.Context, db interface{}, id uuid.UUID, logger *zap.Logger) (*Component, error)  // platform/orchestration/actions/component_library.go
func GetComponentWithFallback(ctx context.Context, db interface{}, function string, logger *zap.Logger) (*Component, error)  // platform/orchestration/actions/component_library.go
func GetImageConfig(purpose string) (width uint, height uint, quality int, extension string)  // platform/storage/url_helpers.go
func GetMapKeys(m map[string]interface{}) []string  // platform/orchestration/datahelpers/data_helpers.go
func GetNavItems(ctx context.Context, db *sql.DB, siteID uuid.UUID, groupTypes []string, deployedOnly bool, maxItems int, logger *zap.Logger) []NavItem  // platform/orchestration/actions/nav_tables.go
func GetStyleCollectionByName(ctx context.Context, db interface{}, name string, logger *zap.Logger) (*StyleCollection, error)  // platform/orchestration/actions/component_library.go
func GetStyleCollectionForSite(ctx context.Context, db interface{}, siteID uuid.UUID, logger *zap.Logger) (*StyleCollection, error)  // platform/orchestration/actions/component_library.go
func InjectFooter(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string  // platform/orchestration/actions/component_library.go
func InjectHead(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string  // platform/orchestration/actions/component_library.go
func InjectHeader(ctx context.Context, db interface{}, html string, siteID uuid.UUID, renderCtx *RenderContext, logger *zap.Logger) string  // platform/orchestration/actions/component_library.go
func IsS3URI(uri string) bool  // platform/storage/url_helpers.go
func New(cfg Config, fetcher provider.ReferenceFetcher, logger *zap.Logger) (*Provider, error)  // internal/adapters/imagegenerator/banana/provider.go
func New(cfg Config, logger *zap.Logger) (*Provider, error)  // internal/adapters/imagegenerator/stability/provider.go
func New(logger *zap.Logger) *Throttle  // internal/adapters/shared/throttle/throttle.go
func New(ctx context.Context, cfg *config.ServiceConfig, logger *zap.Logger) (*Agent, error)  // platform/agentbase/agent.go
func New(code ErrorCode, message string) *ErrorBuilder  // platform/errors/errors.go
func New(logLevel string) (*zap.Logger, error)  // platform/logger/logger.go
func (p *pgxRowScanner) Next() bool  // platform/orchestration/actions/query_agent_definitions_actions.go
func NormalizeComponentFunction(function string) string  // platform/orchestration/actions/component_validation.go
func NormalizeSectionNames(names []string, logger *zap.Logger)  // platform/orchestration/actions/component_validation.go
func RegisterActionInputSpec(actionName string, spec ActionInputSpec)  // platform/orchestration/datahelpers/action_inputs.go
func RegisterLocalActionChecker(checker IsLocalActionFunc)  // platform/orchestration/actioncheck/actioncheck.go
func RenderTemplate(templateStr string, ctx *RenderContext, logger *zap.Logger) string  // platform/orchestration/actions/component_library.go
func Resolve(ctx context.Context, db *sql.DB, req QueryRequest, logger *zap.Logger) (interface{}, error)  // platform/orchestration/actions/queryresolve/queryresolve.go
func Resolve(ctx context.Context, db *sql.DB, req QueryRequest, logger *zap.Logger) (interface{}, error)  // z_context/queryresolve.go
func (j *JSONB) Scan(value interface{}) error  // internal/auth-service/user/models.go
func (p *pgxRowScanner) Scan(dest ...interface{}) error  // platform/orchestration/actions/query_agent_definitions_actions.go
func SelectStyleCollectionByDomain(ctx context.Context, db interface{}, domain string, logger *zap.Logger) (*StyleCollection, error)  // platform/orchestration/actions/component_library.go
func (i *inputs) String() string  // docs/_archive/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/fuse.go
func (i *inputs) String() string  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/fuse.go
func (s *scopeList) String() string  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/assembler.go
func (s *scopeList) String() string  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/assembler.go
func (s *scopeList) String() string  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/cmd/assembler/main.go
func (m *multiFlag) String() string  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/cmd/bundle/main.go
func (i *inputs) String() string  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/cmd/fuse/main.go
func (s *scopeList) String() string  // scripts/documentation_project/02/assembler.go
```
_…and 6 more (widen with -scope or -neighbour package)_

**Called by (callers)**

```go
func (v *WorkflowValidator) RequiresTopic(action string) bool  // platform/validation/workflow.go
func applyAddToPage(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error)  // platform/orchestration/actions/apply_gap_plan_action.go
func applyNewPage(ctx context.Context, db *sql.DB, plan map[string]interface{}, siteID uuid.UUID, domain string, originalItemID *uuid.UUID, logger *zap.Logger) (interface{}, error)  // platform/orchestration/actions/apply_gap_plan_action.go
func classifyPagesForNav(pages []pageNavInfo, logger *zap.Logger) (primary []pageNavInfo, legal []pageNavInfo, utility []pageNavInfo)  // platform/orchestration/actions/populate_nav_tables_action.go
func getActionHandler(action string) (actions.ActionFunc, error)  // platform/orchestration/coordinator.go
func isLocalAction(action string) bool  // platform/orchestration/coordinator.go
func loadComponentSchemas(ctx context.Context, db *sql.DB, sectionNames []string, logger *zap.Logger) map[string]componentInfo  // platform/orchestration/actions/plan_sections_action.go
func loadSingleComponentSchema(ctx context.Context, db *sql.DB, function string, logger *zap.Logger) *componentInfo  // platform/orchestration/actions/plan_sections_action.go
func navPriorityTier(nameLower string, pageType string) int  // platform/orchestration/actions/populate_nav_tables_action.go
func planSection(ctx context.Context, sectionName string, comp componentInfo, resolver *sourceResolver, logger *zap.Logger) sectionPlanItem  // platform/orchestration/actions/plan_sections_action.go
func (v *WorkflowValidator) validateStep(name string, step models.Step, allSteps map[string]models.Step) error  // platform/validation/workflow.go
```

**Types used**

```go
type ActionDefinition struct { Handler ActionFunc; Category string; Description string; IsLocal bool; Deprecated bool; DeprecatedBy string }  // platform/orchestration/actions/types.go
type ActionFunc func(context.Context, ActionParams) (interface{}, error)  // platform/orchestration/actions/types.go
type ActionParams struct { Context context.Context; ExecutionContext *types.ExecutionContext; Headers map[string]string; StepConfig models.Step; InputData []byte; CollectedData map[string]interface{}; SagaCoordinator interface{}; Producer kafka.Producer; DB *sql.DB; StorageClient storage.Client; Logger *zap.Logger; Tracer types.MessageTracer; AgentType string; CurrentStep string }  // platform/orchestration/actions/types.go
type Field struct { Name string; Type string; Tag string }  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/analyser.go
type Field struct { Name string; Type string; Tag string }  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/analyser.go
type Field struct { Name string; Type string; Tag string }  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/internal/analysis/types.go
type Field struct { Name string; Type string; Tag string }  // internal/analysis/types.go
type Field struct { Name string; Type string; Tag string }  // scripts/documentation_project/01/analyser.go
type Field struct { Name string; Type string; Tag string }  // scripts/documentation_project/02/analyser.go
type NavItem struct { Label string; URL string; IsActive bool }  // platform/orchestration/actions/component_library.go
type NavigationItem struct { PageID string; Label string; URL string; Children []NavigationItem }  // platform/orchestration/actions/site_db_actions.go
type RenderContext struct { Domain string; SiteID uuid.UUID; LogoText string; LogoURL string; CompanyName string; Tagline string; NavItems []NavItem; FooterNavItems []NavItem; CurrentPage string; PrimaryColor string; SecondaryColor string; AccentColor string; TextColor string; BackgroundColor string; ThemeCSS string; Title string; Description string; Email string; Phone string; CTAText string; CTAUrl string; Year string; Industry string; Tone string; TargetAudience string; Services []string; ContentData map[string]interface{}; SchemaMode string; SchemaSnapshot map[string]interface{} }  // platform/orchestration/actions/component_library.go
type StyleCollection struct { ID uuid.UUID; Name string; DisplayName string; HeaderComponentID *uuid.UUID; FooterComponentID *uuid.UUID; CSSThemeID *uuid.UUID; ColorPalette map[string]string; Typography map[string]string; Category string; IndustryTags []string }  // platform/orchestration/actions/component_library.go
type field struct { Name string; Type string }  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/go_files_old/go_files/assembler.go
type field struct { Name string; Type string }  // docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/_archive/thin_slice_run/assembler.go
type field struct { Name string; Type string }  // scripts/documentation_project/02/assembler.go
```

_Note: name-matched call graph — a name shared across packages can show extra candidates, and calls through interfaces aren't resolved. Coarse but tight; widen with `-scope` or `-neighbour package` if something's missing._

---

## Schema

```
# Database schema

### `content_components`

```
                                   Table "public.content_components"
         Column          |           Type           | Collation | Nullable |          Default           
-------------------------+--------------------------+-----------+----------+----------------------------
 id                      | uuid                     |           | not null | gen_random_uuid()
 name                    | text                     |           | not null | 
 description             | text                     |           |          | 
 html_template           | text                     |           | not null | 
 input_schema            | jsonb                    |           |          | 
 function                | text                     |           | not null | 'generic-text-block'::text
 created_at              | timestamp with time zone |           |          | now()
 updated_at              | timestamp with time zone |           |          | now()
 display_name            | text                     |           |          | ''::text
 category                | character varying(255)   |           |          | ''::character varying
 semantic_tags           | jsonb                    |           |          | 
 sort_order              | integer                  |           |          | 
 render_mode             | text                     |           |          | 'template'::text
 agent_type              | text                     |           |          | 
 agent_workflow          | text                     |           |          | 
 data_sources            | text[]                   |           |          | 
 child_components        | jsonb                    |           |          | 
 component_level         | text                     |           |          | 'section'::text
 is_active               | boolean                  |           |          | true
 is_dark_section         | boolean                  |           |          | false
 forked_from             | uuid                     |           |          | 
 section_type            | text                     |           |          | 
 suitable_site_types     | jsonb                    |           |          | '[]'::jsonb
 suitable_page_types     | jsonb                    |           |          | '[]'::jsonb
 content_shape           | text                     |           |          | 
 visual_density          | text                     |           |          | 
 usage_count             | integer                  |           |          | 0
 avg_quality_score       | double precision         |           |          | 
 created_from            | text                     |           |          | 'manual'::text
 js_content              | text                     |           |          | 
 template_variable_count | integer                  |           |          | 
 schema_field_count      | integer                  |           |          | 
 template_closed         | boolean                  |           |          | 
 schema_template_synced  | boolean                  |           |          | 
 has_data_component      | boolean                  |           |          | 
 quality_checked_at      | timestamp with time zone |           |          | 
 quality_score           | smallint                 |           |          | 
 quality_issues          | jsonb                    |           |          | '[]'::jsonb
Indexes:
    "content_components_pkey" PRIMARY KEY, btree (id)
    "content_components_name_key" UNIQUE CONSTRAINT, btree (name)
    "idx_cc_forked_from" btree (forked_from) WHERE forked_from IS NOT NULL
    "idx_cc_section_type" btree (section_type) WHERE section_type IS NOT NULL AND is_active = true AND forked_from IS NULL
    "idx_cc_selector" btree (section_type, component_level) WHERE is_active = true AND forked_from IS NULL AND section_type IS NOT NULL
    "idx_cc_suitable_site_types" gin (suitable_site_types) WHERE is_active = true AND forked_from IS NULL
    "idx_cc_tool_function_unique" UNIQUE, btree (function) WHERE component_level = 'tool'::text AND forked_from IS NULL AND is_active = true
    "idx_cc_tool_library" btree (component_level) WHERE component_level = 'tool'::text AND forked_from IS NULL AND is_active = true
    "idx_components_agent_type" btree (agent_type) WHERE agent_type IS NOT NULL
    "idx_components_level" btree (component_level)
    "idx_components_render_mode" btree (render_mode)
    "idx_content_components_category" btree (category)
    "idx_content_components_function" btree (function)
    "idx_content_components_function_quality" btree (function, quality_score DESC NULLS LAST) WHERE is_active = true
    "idx_content_components_quality" btree (quality_score, quality_checked_at) WHERE is_active = true
Check constraints:
    "chk_created_from_valid" CHECK (created_from IS NULL OR (created_from = ANY (ARRAY['manual'::text, 'generated'::text, 'adopted'::text, 'tool'::text, 'forked'::text])))
    "chk_function_kebab_case" CHECK (function = ''::text OR function ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$'::text)
    "chk_quality_score_range" CHECK (quality_score IS NULL OR quality_score >= 0 AND quality_score <= 100)
    "chk_section_type_kebab_case" CHECK (section_type IS NULL OR section_type ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$'::text)
    "chk_usage_count_non_negative" CHECK (usage_count IS NULL OR usage_count >= 0)
    "chk_visual_density_valid" CHECK (visual_density IS NULL OR (visual_density = ANY (ARRAY['low'::text, 'medium'::text, 'high'::text])))
Foreign-key constraints:
    "content_components_forked_from_fkey" FOREIGN KEY (forked_from) REFERENCES content_components(id)
Referenced by:
    TABLE "area_components" CONSTRAINT "area_components_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id)
    TABLE "component_versions" CONSTRAINT "component_versions_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id) ON DELETE CASCADE
    TABLE "content_components" CONSTRAINT "content_components_forked_from_fkey" FOREIGN KEY (forked_from) REFERENCES content_components(id)
    TABLE "layouts" CONSTRAINT "layouts_default_footer_component_id_fkey" FOREIGN KEY (default_footer_component_id) REFERENCES content_components(id) ON DELETE SET NULL
    TABLE "layouts" CONSTRAINT "layouts_default_header_component_id_fkey" FOREIGN KEY (default_header_component_id) REFERENCES content_components(id) ON DELETE SET NULL
    TABLE "page_components" CONSTRAINT "page_components_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id)
    TABLE "site_components" CONSTRAINT "site_components_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_footer_component_id_fkey" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_footer_fk" FOREIGN KEY (footer_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_component_id_fkey" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_fk" FOREIGN KEY (header_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_home_component_id_fkey" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)
    TABLE "style_collections" CONSTRAINT "style_collections_header_home_fk" FOREIGN KEY (header_home_component_id) REFERENCES content_components(id)
```

### `site_specs`

```
                              Table "public.site_specs"
     Column     |           Type           | Collation | Nullable |      Default      
----------------+--------------------------+-----------+----------+-------------------
 id             | uuid                     |           | not null | gen_random_uuid()
 site_id        | uuid                     |           | not null | 
 aspect         | text                     |           | not null | 
 data           | jsonb                    |           | not null | 
 source         | text                     |           | not null | 
 source_agent   | text                     |           |          | 
 source_item_id | uuid                     |           |          | 
 notes          | text                     |           |          | 
 is_current     | boolean                  |           | not null | true
 created_at     | timestamp with time zone |           | not null | now()
 superseded_at  | timestamp with time zone |           |          | 
 created_by     | text                     |           | not null | 
 pinned         | boolean                  |           |          | false
 updated_at     | timestamp with time zone |           |          | now()
Indexes:
    "site_specs_pkey" PRIMARY KEY, btree (id)
    "idx_site_specs_current" UNIQUE, btree (site_id, aspect) WHERE is_current = true
    "idx_site_specs_history" btree (site_id, aspect, created_at DESC)
    "idx_site_specs_lookup" btree (site_id) WHERE is_current = true
    "idx_site_specs_source_item" btree (source_item_id) WHERE source_item_id IS NOT NULL
    "idx_site_specs_updated" btree (updated_at DESC)
Foreign-key constraints:
    "site_specs_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
Triggers:
    trg_site_specs_updated_at BEFORE UPDATE ON site_specs FOR EACH ROW EXECUTE FUNCTION set_updated_at()
```

### `page_components`

```
                               Table "public.page_components"
        Column        |           Type           | Collation | Nullable |      Default      
----------------------+--------------------------+-----------+----------+-------------------
 id                   | uuid                     |           | not null | gen_random_uuid()
 page_id              | uuid                     |           |          | 
 component_id         | uuid                     |           |          | 
 position             | integer                  |           | not null | 
 slot_name            | character varying(100)   |           |          | 
 parent_instance_id   | uuid                     |           |          | 
 rendered_html        | text                     |           |          | 
 content_data         | jsonb                    |           |          | 
 content_hash         | character varying(64)    |           |          | 
 data_path            | character varying(500)   |           |          | 
 data_uuid            | uuid                     |           |          | gen_random_uuid()
 created_at           | timestamp with time zone |           |          | now()
 updated_at           | timestamp with time zone |           |          | now()
 build_status         | text                     |           |          | 'pending'::text
 reviewed_at          | timestamp with time zone |           |          | 
 reviewed_by          | text                     |           |          | 
 deploy_commit        | text                     |           |          | 
 research_id          | uuid                     |           |          | 
 sources_displayed    | boolean                  |           |          | false
 content_item_id      | uuid                     |           |          | 
 component_version_id | uuid                     |           |          | 
 schema_mode          | text                     |           |          | 
 locked_at            | timestamp with time zone |           |          | 
 locked_by            | text                     |           |          | 
 content_brief        | jsonb                    |           |          | 
 lock_type            | text                     |           |          | 
 lock_expires_at      | timestamp with time zone |           |          | 
Indexes:
    "page_components_pkey" PRIMARY KEY, btree (id)
    "idx_page_components_content" btree (content_item_id)
    "idx_page_components_locked" btree (locked_at) WHERE locked_at IS NOT NULL
    "idx_page_components_page" btree (page_id)
    "idx_page_components_parent" btree (parent_instance_id)
    "idx_page_components_schema_mode" btree (schema_mode) WHERE schema_mode IS NOT NULL
    "idx_page_components_status" btree (build_status)
    "idx_page_components_template" btree (component_id)
    "idx_page_components_timed_lock" btree (lock_expires_at) WHERE lock_type = 'timed'::text
Check constraints:
    "chk_page_components_lock_type" CHECK (lock_type IS NULL OR (lock_type = ANY (ARRAY['permanent'::text, 'timed'::text, 'review'::text])))
Foreign-key constraints:
    "page_components_component_id_fkey" FOREIGN KEY (component_id) REFERENCES content_components(id)
    "page_components_content_item_id_fkey" FOREIGN KEY (content_item_id) REFERENCES content_items(id) ON DELETE SET NULL
    "page_components_page_id_fkey" FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
    "page_components_parent_instance_id_fkey" FOREIGN KEY (parent_instance_id) REFERENCES page_components(id)
    "page_components_research_id_fkey" FOREIGN KEY (research_id) REFERENCES research_results(id) ON DELETE SET NULL
Referenced by:
    TABLE "link_registry" CONSTRAINT "link_registry_source_component_instance_id_fkey" FOREIGN KEY (source_component_instance_id) REFERENCES page_components(id) ON DELETE CASCADE
    TABLE "page_component_history" CONSTRAINT "page_component_history_component_id_fkey" FOREIGN KEY (component_id) REFERENCES page_components(id) ON DELETE SET NULL
    TABLE "page_components" CONSTRAINT "page_components_parent_instance_id_fkey" FOREIGN KEY (parent_instance_id) REFERENCES page_components(id)
    TABLE "research_results" CONSTRAINT "research_results_component_instance_id_fkey" FOREIGN KEY (component_instance_id) REFERENCES page_components(id) ON DELETE SET NULL
Triggers:
    trigger_auto_lock_on_deploy BEFORE UPDATE ON page_components FOR EACH ROW EXECUTE FUNCTION auto_lock_on_deploy()
```

### `pages`

```
                                          Table "public.pages"
         Column          |           Type           | Collation | Nullable |           Default           
-------------------------+--------------------------+-----------+----------+-----------------------------
 id                      | uuid                     |           | not null | gen_random_uuid()
 site_id                 | uuid                     |           |          | 
 name                    | character varying(255)   |           | not null | 
 url                     | character varying(500)   |           | not null | 
 title                   | character varying(500)   |           |          | 
 page_type               | character varying(50)    |           |          | 
 status                  | character varying(50)    |           |          | 'active'::character varying
 content_hash            | character varying(64)    |           |          | 
 meta_description        | text                     |           |          | 
 topics                  | text[]                   |           |          | 
 nav_label               | character varying(255)   |           |          | 
 nav_order               | integer                  |           |          | 100
 in_header               | boolean                  |           |          | true
 in_footer               | boolean                  |           |          | true
 last_built_at           | timestamp with time zone |           |          | 
 expires_at              | timestamp with time zone |           |          | 
 created_at              | timestamp with time zone |           |          | now()
 updated_at              | timestamp with time zone |           |          | now()
 build_status            | text                     |           |          | 'pending'::text
 deployed_at             | timestamp with time zone |           |          | 
 sections                | jsonb                    |           |          | '[]'::jsonb
 version                 | integer                  |           |          | 1
 rendered_header         | text                     |           |          | 
 rendered_footer         | text                     |           |          | 
 rendered_head           | text                     |           |          | 
 site_area_id            | uuid                     |           |          | 
 content_direction       | jsonb                    |           |          | 
 page_spec               | jsonb                    |           |          | 
 suppressed_sections     | jsonb                    |           |          | '[]'::jsonb
 built_from_plan_version | uuid                     |           |          | 
Indexes:
    "pages_pkey" PRIMARY KEY, btree (id)
    "idx_pages_area" btree (site_area_id) WHERE site_area_id IS NOT NULL
    "idx_pages_build_status" btree (site_id, build_status)
    "idx_pages_nav" btree (site_id, in_header, nav_order) WHERE status::text = 'active'::text
    "idx_pages_needs_build" btree (site_id) WHERE build_status = ANY (ARRAY['planned'::text, 'needs_rebuild'::text])
    "idx_pages_plan_version" btree (site_id, built_from_plan_version)
    "idx_pages_site" btree (site_id)
    "idx_pages_status" btree (status)
    "idx_pages_type" btree (page_type)
    "pages_site_id_name_key" UNIQUE CONSTRAINT, btree (site_id, name)
Check constraints:
    "chk_page_type_kebab_case" CHECK (page_type IS NULL OR page_type::text = ''::text OR page_type::text ~ '^[a-z][a-z0-9]*(-[a-z0-9]+)*$'::text)
Foreign-key constraints:
    "pages_site_area_id_fkey" FOREIGN KEY (site_area_id) REFERENCES site_areas(id)
    "pages_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE
Referenced by:
    TABLE "flow_pages" CONSTRAINT "flow_pages_page_id_fkey" FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
    TABLE "link_registry" CONSTRAINT "link_registry_source_page_id_fkey" FOREIGN KEY (source_page_id) REFERENCES pages(id) ON DELETE CASCADE
    TABLE "link_registry" CONSTRAINT "link_registry_target_page_id_fkey" FOREIGN KEY (target_page_id) REFERENCES pages(id)
    TABLE "page_component_history" CONSTRAINT "page_component_history_page_id_fkey" FOREIGN KEY (page_id) REFERENCES pages(id)
    TABLE "page_components" CONSTRAINT "page_components_page_id_fkey" FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
    TABLE "redirects" CONSTRAINT "redirects_source_page_id_fkey" FOREIGN KEY (source_page_id) REFERENCES pages(id)
    TABLE "research_results" CONSTRAINT "research_results_page_id_fkey" FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE
    TABLE "site_nav_items" CONSTRAINT "site_nav_items_page_id_fkey" FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE SET NULL
Triggers:
    trg_invalidate_nav_on_page_change AFTER INSERT OR DELETE OR UPDATE ON pages FOR EACH ROW EXECUTE FUNCTION invalidate_navigation_cache()
```

### `site_work_items`

```
                             Table "public.site_work_items"
      Column      |           Type           | Collation | Nullable |      Default      
------------------+--------------------------+-----------+----------+-------------------
 id               | uuid                     |           | not null | gen_random_uuid()
 site_id          | uuid                     |           | not null | 
 source           | text                     |           | not null | 
 item_type        | text                     |           | not null | 
 severity         | text                     |           | not null | 'medium'::text
 summary          | text                     |           | not null | 
 spec             | jsonb                    |           |          | '{}'::jsonb
 page_id          | uuid                     |           |          | 
 component_id     | uuid                     |           |          | 
 entity_id        | uuid                     |           |          | 
 affected_url     | text                     |           |          | 
 impact           | jsonb                    |           |          | 
 resolution_path  | text                     |           |          | 
 suggested_action | text                     |           |          | 
 priority         | integer                  |           |          | 100
 handler_agent    | text                     |           |          | 
 status           | text                     |           | not null | 'detected'::text
 created_by       | text                     |           | not null | 
 handled_by       | text                     |           |          | 
 approved_by      | text                     |           |          | 
 claimed_by       | text                     |           |          | 
 depends_on       | uuid[]                   |           |          | 
 parent_item_id   | uuid                     |           |          | 
 related_item_ids | uuid[]                   |           |          | 
 batch_id         | uuid                     |           |          | 
 attempt_count    | integer                  |           |          | 0
 max_attempts     | integer                  |           |          | 3
 result           | jsonb                    |           |          | '{}'::jsonb
 error            | text                     |           |          | 
 item_key         | text                     |           |          | 
 created_at       | timestamp with time zone |           |          | now()
 triaged_at       | timestamp with time zone |           |          | 
 claimed_at       | timestamp with time zone |           |          | 
 completed_at     | timestamp with time zone |           |          | 
 approval_mode    | text                     |           |          | 'auto'::text
 updated_at       | timestamp with time zone |           |          | now()
 pipeline         | text                     |           | not null | 'build'::text
Indexes:
    "site_work_items_pkey" PRIMARY KEY, btree (id)
    "idx_swi_batch" btree (batch_id) WHERE batch_id IS NOT NULL
    "idx_swi_claimed" btree (status, claimed_at) WHERE status = 'claimed'::text
    "idx_swi_completed" btree (completed_at) WHERE status = ANY (ARRAY['complete'::text, 'verified'::text, 'rejected'::text, 'wont_fix'::text])
    "idx_swi_dedup" UNIQUE, btree (site_id, item_key) WHERE item_key IS NOT NULL AND (status <> ALL (ARRAY['complete'::text, 'verified'::text, 'rejected'::text, 'wont_fix'::text, 'failed'::text, 'unresolved'::text]))
    "idx_swi_deps" gin (depends_on) WHERE depends_on IS NOT NULL
    "idx_swi_handler" btree (handler_agent, status) WHERE status = ANY (ARRAY['triaged'::text, 'approved'::text])
    "idx_swi_page" btree (page_id) WHERE page_id IS NOT NULL
    "idx_swi_pipeline" btree (pipeline, status)
    "idx_swi_site_pending" btree (site_id, priority) WHERE status = ANY (ARRAY['triaged'::text, 'approved'::text])
    "idx_swi_site_status" btree (site_id, status)
    "idx_work_items_blocked" btree (handler_agent) WHERE status = 'blocked'::text
Foreign-key constraints:
    "site_work_items_parent_item_id_fkey" FOREIGN KEY (parent_item_id) REFERENCES site_work_items(id)
    "site_work_items_site_id_fkey" FOREIGN KEY (site_id) REFERENCES sites(id)
Referenced by:
    TABLE "content_feed_items" CONSTRAINT "content_feed_items_work_item_id_fkey" FOREIGN KEY (work_item_id) REFERENCES site_work_items(id)
    TABLE "site_work_items" CONSTRAINT "site_work_items_parent_item_id_fkey" FOREIGN KEY (parent_item_id) REFERENCES site_work_items(id)
```
```

---

## Database capabilities

# Database capabilities

_What the database can do — installed extensions and helper functions. Reuse these before hand-rolling SQL (e.g. snapshot_agent for backups, pgvector for similarity)._

## Installed extensions (`\dx`)

```
                                     List of installed extensions
   Name    | Version |   Schema   |                            Description                            
-----------+---------+------------+-------------------------------------------------------------------
 pg_trgm   | 1.6     | public     | text similarity measurement and index searching based on trigrams
 pgcrypto  | 1.3     | public     | cryptographic functions
 plpgsql   | 1.0     | pg_catalog | PL/pgSQL procedural language
 uuid-ossp | 1.1     | public     | generate universally unique identifiers (UUIDs)
 vector    | 0.8.0   | public     | vector data type and ivfflat and hnsw access methods
(5 rows)
```

## Functions (`\df snapshot`)

```
                       List of functions
 Schema | Name | Result data type | Argument data types | Type 
--------+------+------------------+---------------------+------
(0 rows)
```

---

## Runtime evidence

_not available in the thin slice. For a real debug bundle this is the run trace, step sequence, errors, and logs correlated by `orchestration_id`._

---

## Pointers

- This bundle is a selection from 533 analysed files. In-scope: platform/orchestration/actions/v3_site_actions.go, platform/orchestration/actions/reconcile_section_data_action.go, platform/orchestration/actions/registry.go.
- Neighbourhood mode: call-graph slice (callees, callers, types used) of the in-scope symbols.
- Everything else is omitted. To pull more in, re-run with another `-scope path[:Symbol]`, or `-neighbour package` for the whole package.
