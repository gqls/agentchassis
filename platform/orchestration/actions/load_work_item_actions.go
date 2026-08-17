// FILE: platform/orchestration/actions/load_work_item_actions.go
// Work item actions for the unified build/maintenance queue (site_work_items table).
//
// Actions:
//   WriteBuildItemsAction    — queries pages from DB, creates work items in site_work_items
//   LoadWorkItemsAction      — loads pending items for a site, respecting dependencies
//   CompleteWorkItemAction   — marks a work item as complete with result data
//   FailWorkItemAction       — marks a work item as failed with error info
//
// These actions run alongside existing maintenance_queue actions. No migration needed.

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	checks "github.com/gqls/agentchassis/platform/orchestration/actions/discovery_checks"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpecs
// ============================================================================

var WriteBuildItemsInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"site_id"},
	Optional:    []string{"site_plan", "batch_id"},
	Defaults:    map[string]interface{}{},
	Deprecated:  map[string]string{},
}

// LoadWorkItemsInputSpec — only site_id needs resolution from collectedData.
// Filter fields (item_pipeline, handler_agent, max_items) are pure config literals
// and are read directly from params.StepConfig.Config in the action body.
// NOTE: "pipeline" is NOT listed here because site_record.pipeline would be picked
// up by nested lookup (field name collision — see checklist section on this).
var LoadWorkItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

var CompleteWorkItemInputSpec = datahelpers.ActionInputSpec{
	CheckConfig: true,
	Required:    []string{"work_item_id"},
	Optional:    []string{"result", "commit_sha"},
	Defaults:    map[string]interface{}{},
	Deprecated: map[string]string{
		"work_item_id_field": "work_item_id",
		"commit_sha_field":   "commit_sha",
		"result_field":       "result",
	},
}

// FailWorkItemInputSpec — work_item_id needs path resolution.
// error_message is a config literal (e.g. "Content review not approved")
// and is read directly from params.StepConfig.Config.
var FailWorkItemInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"work_item_id"},
	Optional: []string{},
	Defaults: map[string]interface{}{},
	Deprecated: map[string]string{
		"work_item_id_field": "work_item_id",
	},
}

func init() {
	datahelpers.RegisterActionInputSpec("write_build_items", WriteBuildItemsInputSpec)
	datahelpers.RegisterActionInputSpec("load_work_items", LoadWorkItemsInputSpec)
	datahelpers.RegisterActionInputSpec("complete_work_item", CompleteWorkItemInputSpec)
	datahelpers.RegisterActionInputSpec("fail_work_item", FailWorkItemInputSpec)
}

// ============================================================================
// ACTION: write_build_items
// Used by: site-work-orchestrator (after sync_pages_to_db)
// ============================================================================
//
// Queries pages from the pages table (build_status = 'planned') and creates
// one work item per page. The spec field contains the full page record from DB,
// matching exactly what get_pages_to_build returns — so current_item.spec has
// the same fields as current_page in pageflow-builder (id, name, url, title,
// sections, page_type, etc).
//
// Also creates tracking items for site-wide tasks (logo, hero, design) based
// on flags in the site_plan.
//
// Data inputs:
//   - site_id (required) — UUID of the site (pages must already be synced)
//   - site_plan (optional) — planner output for asset/design flags
//   - batch_id (optional) — UUID to group items; generated if not provided
func WriteBuildItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("WriteBuildItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		WriteBuildItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Check if site is locked — skip all automated work
	var siteLocked bool
	if err := params.DB.QueryRowContext(ctx, `
		SELECT locked_at IS NOT NULL FROM sites WHERE id = $1
	`, siteID).Scan(&siteLocked); err == nil && siteLocked {
		logger.Info("LoadWorkItemsAction: site is locked, skipping",
			zap.String("site_id", siteIDStr))
		return map[string]interface{}{
			"items":          []interface{}{},
			"count":          0,
			"skipped_reason": "site_locked",
		}, nil
	}

	batchID := uuid.New()
	if b := inputs.Get("batch_id"); b != "" {
		if parsed, err := uuid.Parse(b); err == nil {
			batchID = parsed
		}
	}

	// ---- Query pages from DB (same data as get_pages_to_build) ----
	//
	// Match both 'planned' (fresh sites) and 'needs_rebuild' (re-plans,
	// improvement-loop triggers, re-classification). sync_pages_to_db
	// sets existing pages to 'needs_rebuild' when a planner re-plans
	// them; the old filter on 'planned' alone would miss those.
	// includeOwned=false: a rebuild_policy='owned' page must not be given a
	// needs_page item at all (bugs_open/208). This is the same refusal
	// ReconcileSitePlanAction already makes for its own route — it emits
	// owned_page_review instead of needs_page (reconcile_site_plan_action.go:232-270)
	// — and this action is a needs_page producer that had no such guard.
	//
	// Whether the item is then DANGEROUS depends on which consumer picks it up,
	// and the two differ (measured 2026-08-06): page-build-handler runs
	// save_sections -> update_status -> deploy_page, so its ownership refusal
	// fires BEFORE anything is committed and that route is safe; but
	// site-work-orchestrator's build_items_loop runs assemble_page -> deploy_page
	// (git_commit) -> save_sections, so there the commit lands first. Not filing
	// the item is what makes the difference irrelevant.
	pages, err := queryPagesForBuild(ctx, params.DB, siteID,
		[]string{"planned", "needs_rebuild"}, false, false, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to query pages: %w", err)
	}

	// Zero pages is NOT an early exit. needs_composition, needs_design,
	// and needs_rerender are site-scoped and must queue regardless of
	// per-page work. The per-page loop below is a no-op when pages is
	// empty; execution continues to the site-level inserts that follow.
	if len(pages) == 0 {
		logger.Info("WriteBuildItemsAction: no per-page builds needed; queuing site-level items only",
			zap.String("site_id", siteIDStr),
		)
	}

	// ---- Extract plan flags for asset/design items ----
	planData := inputs.GetMap("site_plan")
	if planData == nil {
		planData = make(map[string]interface{})
	}
	if resp, ok := planData["response"].(map[string]interface{}); ok {
		if pd, ok := resp["plan_data"].(map[string]interface{}); ok {
			planData = pd
		} else {
			planData = resp
		}
	}

	needsLogo, _ := planData["needs_logo"].(bool)
	needsImages, _ := planData["needs_images"].(bool)

	// ---- Insert work items in a transaction ----
	tx, err := params.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	inserted := 0
	skipped := 0

	// Content page items
	for i, page := range pages {
		pageName, _ := page["name"].(string)
		if pageName == "" {
			continue
		}

		pageIDStr, _ := page["id"].(string)
		var pageIDPtr *uuid.UUID
		if parsed, err := uuid.Parse(pageIDStr); err == nil {
			pageIDPtr = &parsed
		}

		pageType, _ := page["page_type"].(string)
		if pageType == "" {
			pageType = "content"
		}
		pageType = normalizePageType(pageType) // already exists at line 784

		// Known available builders — update this map when adding new builder agents
		type builderInfo struct {
			handler  string
			itemType string
		}
		availableBuilders := map[string]builderInfo{
			"content":    {handler: "page-build-handler", itemType: "needs_content_page"},
			"index":      {handler: "page-build-handler", itemType: "needs_content_page"},
			"landing":    {handler: "page-build-handler", itemType: "needs_content_page"},
			"blog-index": {handler: "page-build-handler", itemType: "needs_content_page"},
			"blog-post":  {handler: "page-build-handler", itemType: "needs_content_page"},
			// bugs_open/206: directory-build-handler ensures the page's plan
			// layout (ensure_page_section_layout, defaulting to
			// ["hero","directory-listing"]) then delegates the actual build
			// to page-build-handler — see agent_definitions seed
			// 001_directory_build_handler.sql. entity-page stays unavailable
			// below: practice pages are deliberately on hold pending more
			// source data (features_open/021's own P1), so no builder for it
			// yet — building one now would be ahead of a demonstrated need.
			"entity-directory": {handler: "directory-build-handler", itemType: "needs_directory"},
			// Add here as builders become available:
			// "entity-page":      {handler: "entity-page-build-handler", itemType: "needs_entity_page"},
			// "tool":             {handler: "tool-build-handler", itemType: "needs_tool_page"},
			// # "blog-index":       {handler: "blog-build-handler", itemType: "needs_blog_index"},
			// # "blog-post":        {handler: "blog-build-handler", itemType: "needs_blog_post"},
			// and news
		}

		// Known page types whose builders don't exist yet
		unavailableBuilders := map[string]string{
			"tool":        "tool-builder",
			"entity-page": "entity-page-builder",
		}

		handlerAgent := "page-build-handler"
		itemType := "needs_content_page"

		if info, available := availableBuilders[pageType]; available {
			handlerAgent = info.handler
			itemType = info.itemType
		} else if neededBuilder, known := unavailableBuilders[pageType]; known {
			// Known type but builder not available — log deferred capability gap
			gapSpec, _ := json.Marshal(map[string]interface{}{
				"page_name":      pageName,
				"page_type":      pageType,
				"builder_needed": neededBuilder,
				"page_spec":      page,
			})

			ok, err := insertWorkItem(ctx, tx, workItem{
				siteID:       siteID,
				source:       "planner",
				pipeline:     "build",
				itemType:     "capability_gap",
				severity:     "low",
				summary:      fmt.Sprintf("Page '%s' needs %s (not yet available)", pageName, neededBuilder),
				spec:         string(gapSpec),
				pageID:       pageIDPtr,
				priority:     200,
				handlerAgent: neededBuilder,
				status:       "deferred",
				createdBy:    "site-planner",
				itemKey:      fmt.Sprintf("capability_gap:%s:%s", pageType, pageName),
				batchID:      batchID,
			}, logger)
			if err != nil {
				logger.Warn("WriteBuildItemsAction: Failed to insert capability_gap",
					zap.String("page", pageName), zap.Error(err))
			} else if ok {
				logger.Info("WriteBuildItemsAction: Recorded capability gap",
					zap.String("page", pageName),
					zap.String("page_type", pageType),
					zap.String("builder_needed", neededBuilder),
				)
			}
			continue // Skip — don't create a dispatch work item for this page
		} else {
			// Completely unknown page_type — fall through to default handler
			logger.Warn("WriteBuildItemsAction: Unknown page_type, using page-build-handler",
				zap.String("page_type", pageType),
				zap.String("page", pageName),
			)
		}

		specJSON, err := json.Marshal(page)
		if err != nil {
			logger.Warn("WriteBuildItemsAction: marshal error",
				zap.String("page", pageName), zap.Error(err))
			continue
		}

		ok, err := insertWorkItem(ctx, tx, workItem{
			siteID:       siteID,
			source:       "planner",
			pipeline:     "build",
			itemType:     itemType,
			severity:     "high",
			summary:      fmt.Sprintf("Build page: %s", pageName),
			spec:         string(specJSON),
			pageID:       pageIDPtr,
			priority:     10 + i,
			handlerAgent: handlerAgent,
			status:       "triaged",
			createdBy:    "site-planner",
			itemKey:      fmt.Sprintf("needs_page:%s", pageName),
			batchID:      batchID,
		}, logger)
		if err != nil {
			logger.Warn("WriteBuildItemsAction: insert error",
				zap.String("page", pageName), zap.Error(err))
			continue
		}
		if ok {
			inserted++
		} else {
			skipped++
		}
	}

	// Asset tracking items.
	//
	// image-build-handler REQUIRES spec.image_prompts.<key> for the logo and
	// hero branches (call_logo_gen maps "prompt" from
	// input_data.spec.image_prompts.logo, call_hero_gen from …hero_home, and
	// input_mapping is a strict allow-list), while needs_logo/needs_images are
	// booleans the planner sets INDEPENDENTLY of whether it wrote any prompt —
	// plan["image_prompts"] is defaulted to an empty map at
	// v3_site_actions.go:3107 by a separate branch. So a plan saying
	// needs_logo:true with no logo prompt files an item that can never be
	// handled: it dies at input extraction with "source path
	// 'input_data.spec.image_prompts.logo' not found" (bugs_open/210).
	//
	// Do not file a generation request we already know is unconsumable.
	if needsLogo {
		insertImageBuildItem(ctx, tx, imageBuildItemArgs{
			db: params.DB, siteID: siteID, batchID: batchID, planData: planData,
			purpose: "logo", promptKey: "logo",
			itemType: "needs_logo", itemKey: "needs_logo",
			summary: "Generate site logo", severity: "high",
		}, &inserted, logger)
	}
	if needsImages {
		insertImageBuildItem(ctx, tx, imageBuildItemArgs{
			db: params.DB, siteID: siteID, batchID: batchID, planData: planData,
			purpose: "hero", promptKey: "hero_home",
			itemType: "needs_hero_image", itemKey: "needs_hero:home",
			summary: "Generate hero image", severity: "medium",
		}, &inserted, logger)
	}

	// (helpers for the two blocks above live at the bottom of this file)

	// Design tracking item
	// The existing block just inserted needs_design. Replace with the two-step
	// sequence below — insert needs_composition first, look up its id, then
	// insert needs_design with depends_on pointing at it.

	// Composition tracking item — site-design-planner resolves palette/layout/
	// typography BEFORE webdesign-agent renders. Priority 7 = one before design.
	compSpec, _ := json.Marshal(map[string]interface{}{
		"reason": "build_pipeline_initial_composition",
	})
	_, err = insertWorkItem(ctx, tx, workItem{
		siteID:       siteID,
		source:       "planner",
		pipeline:     "build",
		itemType:     "needs_composition",
		severity:     "high",
		summary:      "Resolve palette/layout/typography composition for the site",
		spec:         string(compSpec),
		priority:     7,
		handlerAgent: "site-design-planner",
		status:       "triaged",
		createdBy:    "site-planner",
		itemKey:      "needs_composition",
		batchID:      batchID,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("insert needs_composition: %w", err)
	}

	// Look up the composition item's id so we can wire needs_design's
	// depends_on at it. insertWorkItem doesn't return the id (it returns
	// only inserted/suppressed bool), and the two-strike / dedup logic
	// means we can't assume we just inserted a fresh row. The partial
	// unique index guarantees at most one open row for this (site_id,
	// item_key) — that's the composition item we want to depend on.
	//
	// If a suppressed duplicate exists (same item_key, still open), we
	// depend on that existing row — the design item waits for whichever
	// composition item is actually active. If NO open composition item
	// exists (both this insert AND any prior were terminal), we skip the
	// dependency — design will run without waiting, which is the closest
	// thing to correct behaviour: composition already ran or won't run.
	var compositionIDPtr *uuid.UUID
	var fetchedID uuid.UUID
	// Note: predicate must match idx_swi_dedup's exclusion set — we want
	// "open" items, i.e. rows the dedup index covers. Built from
	// workItemTerminalStatuses so it stays aligned with insertWorkItem's
	// ON CONFLICT WHERE clause and migration 012's index predicate.
	compLookupSQL := fmt.Sprintf(`
		SELECT id FROM site_work_items
		WHERE site_id = $1
		  AND item_key = 'needs_composition'
		  AND status NOT IN (%s)
		ORDER BY created_at DESC
		LIMIT 1
	`, sqlInList(workItemTerminalStatuses))
	err = tx.QueryRowContext(ctx, compLookupSQL, siteID).Scan(&fetchedID)
	if err == nil {
		compositionIDPtr = &fetchedID
		inserted++ // Count the composition item as an insertion for summary purposes
		logger.Info("WriteBuildItemsAction: composition item queued",
			zap.String("composition_item_id", fetchedID.String()),
		)
	} else if err == sql.ErrNoRows {
		logger.Info("WriteBuildItemsAction: no open composition item found — design will run without dependency",
			zap.String("site_id", siteIDStr),
		)
	} else {
		logger.Warn("WriteBuildItemsAction: composition item lookup failed — design will run without dependency",
			zap.Error(err),
		)
	}

	// Build the depends_on slice for the design item. Nil/empty slice means
	// no dependency — insertWorkItem handles both identically (the
	// dependsOnStr branch at line 911 only runs if len > 0).
	var designDepends []uuid.UUID
	if compositionIDPtr != nil {
		designDepends = []uuid.UUID{*compositionIDPtr}
	}

	// Design tracking item — now gated by composition completion.
	// Kept at priority 8 so the ordering signal is preserved even for
	// operators reading the queue.
	ok, _ := insertWorkItem(ctx, tx, workItem{
		siteID: siteID, source: "planner", pipeline: "build",
		itemType: "needs_design", severity: "high",
		summary: "Generate site stylesheet", spec: "{}",
		priority: 8, handlerAgent: "webdesign-agent",
		status: "triaged", createdBy: "site-planner",
		itemKey: "needs_design", batchID: batchID,
		dependsOn: designDepends, // ← NEW: wait for composition
	}, logger)
	if ok {
		inserted++
	}

	// Rerender tracking item — assembles pages from components after content + design
	rerenderSpec, _ := json.Marshal(map[string]interface{}{
		"refresh_site_components": true,
	})
	ok, _ = insertWorkItem(ctx, tx, workItem{
		siteID: siteID, source: "planner", pipeline: "build",
		itemType: "needs_rerender", severity: "medium",
		summary:  "Re-assemble and deploy all pages after build completes",
		spec:     string(rerenderSpec),
		priority: 20, handlerAgent: "rerender-pages",
		status: "triaged", createdBy: "site-planner",
		itemKey: "needs_rerender", batchID: batchID,
	}, logger)
	if ok {
		inserted++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit work items: %w", err)
	}

	logger.Info("WriteBuildItemsAction: Complete",
		zap.Int("items_inserted", inserted),
		zap.Int("items_skipped", skipped),
		zap.Int("pages_found", len(pages)),
		zap.String("batch_id", batchID.String()),
	)

	return map[string]interface{}{
		"items_inserted": inserted,
		"items_skipped":  skipped,
		"batch_id":       batchID.String(),
		"total_items":    inserted + skipped,
		"page_count":     len(pages),
	}, nil
}

// setRoutingField exposes one first-class routing column on the item map,
// resolved COLUMN FIRST and falling back to the same key inside spec.
//
// The fallback is what makes a single dispatcher path correct for both
// populations: input_mapping resolves exactly one source path per destination
// and has no coalesce syntax, so without this the config would have to choose
// which half of the fleet's work items it could route (bugs_open/154).
//
// A blank column with no spec key writes NOTHING — the key must stay absent
// rather than become "", because the dispatch mapping marks these optional
// ("component_id?") and an optional path that RESOLVES to empty is passed on
// as an empty string, where a path that is MISSING is skipped. Handlers gate on
// presence (create_rerender_items: `componentIDStr != ""`), so materialising ""
// would turn "not supplied" into "supplied as empty" for every item that has
// neither source.
func setRoutingField(item map[string]interface{}, key, columnValue string, logger *zap.Logger) {
	if columnValue != "" {
		item[key] = columnValue
		return
	}
	spec, ok := item["spec"].(map[string]interface{})
	if !ok {
		return
	}
	// Only strings pass through: these are id/URL fields, and a non-string here
	// means the spec is carrying something else under the same name.
	if fromSpec, present := spec[key].(string); present && fromSpec != "" {
		item[key] = fromSpec
		return
	}
	// %T only — the TYPE, never the value. spec is caller-supplied and can carry
	// site content, so the shape is what is diagnostic here and the content is
	// what must not reach a log line.
	if unusable, present := spec[key]; present {
		logger.Warn("LoadWorkItemsAction: spec key is not a non-empty string — routing field left unset",
			zap.String("key", key),
			zap.String("spec_type", fmt.Sprintf("%T", unusable)))
	}
}

// aliasGuidanceIntoSuggestion makes the dead guidance spelling live
// (bugs_open/271): four emitters wrote spec.content_guidance — the content-gap
// pipeline (apply_gap_plan), tool-generator, tool-deployer and the tool-content
// raiser — while every reader reads spec.suggestion. The brief was therefore
// discarded before any prompt was built, and the rewrite happened anyway,
// steered by writer_block and the existing page, and reported complete.
//
// Same root constraint as setRoutingField above: input_mapping resolves exactly
// one source path per destination and has no coalesce syntax, so "read either
// key" cannot be expressed in config. It has to happen here, at the one point
// every dispatched item passes through.
//
// IN-MEMORY ONLY — the DB row is never touched, so historical specs stay
// byte-identical for any out-of-repo reader, and this is not a migration.
//
// Writes AT MOST the one key "suggestion": never when suggestion already holds
// a value (of any type), never from an empty or non-string guidance, and never
// as "" — an optional mapping path that RESOLVES to empty is forwarded as an
// empty string where a MISSING path is skipped, and handlers gate on presence
// (see the note on setRoutingField).
//
// suggestion is a PROSE channel, which is what makes this safe where
// backfilling component_id was not: enumerated over every active
// agent_definitions row on 2026-08-15, its only live readers are
// page-build-handler's optional "rewrite_guidance?" mapping (which reaches
// page-content-writer's "## Rewrite Guidance" prompt block) and prompt lines in
// content-gap-planner and css-patch-agent. None gates scope, routing, or any
// branch.
func aliasGuidanceIntoSuggestion(spec map[string]interface{}) {
	if existing, present := spec["suggestion"]; present {
		if s, ok := existing.(string); !ok || s != "" {
			// A real value, or a shape this function will not judge. Either way
			// the author's own key wins — the alias never overwrites.
			return
		}
	}
	guidance, ok := spec["content_guidance"].(string)
	if !ok || guidance == "" {
		return
	}
	spec["suggestion"] = guidance
}

// ============================================================================
// ACTION: load_work_items
// Used by: site-work-orchestrator (to get items to process)
// ============================================================================
//
// Returns items ordered by priority. Filters out items whose dependencies
// haven't completed. Each item's spec field is a parsed map (not JSON string)
// so dot-notation access works: current_item.spec.name, current_item.spec.id.
//
// FIRST-CLASS ROUTING COLUMNS (bugs_open/154). site_work_items carries
// component_id, entity_id and affected_url as real columns, but this action
// used to expose only page_id — so the ONLY path a dispatcher could reference
// was current_item.spec.<key>, i.e. a copy the creating agent had to remember
// to duplicate into the spec JSONB. A creator that populated the column and
// not the blob produced an item that was structurally undispatchable: the
// dispatch loop's optional "component_id?" mapping silently skipped, and the
// handler's first query_database died on an unresolved param. That is what
// killed every tool-auditor-raised improve_tool item at tool-improver's
// load_tool step, while items from three other creators ran clean.
//
// So each of these is exposed top-level and resolved COLUMN FIRST, falling
// back to spec.<key>. One path — current_item.component_id — is then correct
// for both populations, which is what lets the dispatch mapping be a single
// path (input_mapping has no coalesce syntax; see input_contracts).
//
// Deliberately NOT applied to page_id, which stays column-only. It is already
// exposed, so there is no bug to close, and measured 2026-07-31 there are 218
// rows with a NULL column and a spec.page_id — every one of which would newly
// gain current_item.page_id. Widening what reaches a handler changes it
// without editing it; that is a change to make for a reason, not for symmetry.
//
// The spec map is NEVER mutated for ROUTING/ID keys, and that rule is
// load-bearing. Backfilling the resolved value into spec was the other
// candidate and it is unsafe: rerender-pages reads
// input_data.spec.component_id, and create_rerender_items gates
// `scoped := (reason == "section_data_resolved" || reason == "image_landed")
// && componentIDStr != ""` on it — so writing into spec could silently flip a
// site-wide rerender into a component-scoped one. Top-level exposure leaves
// every existing spec.* reader reading exactly what it reads today.
//
// ONE NARROW EXCEPTION, added 2026-08-15: aliasGuidanceIntoSuggestion
// (bugs_open/271) fills spec.suggestion from spec.content_guidance when
// suggestion is absent. The component_id hazard does not transfer, and the
// discriminating test is what the key DOES, not that the write is additive:
// suggestion is a prose channel whose only live readers (enumerated by query
// 2026-08-15 over every active agent_definitions row — page-build-handler's
// optional "rewrite_guidance?" mapping, plus prompt lines in
// content-gap-planner and css-patch-agent) render it into a prompt. None gates
// scope, routing, or any branch. A future candidate that fails THAT test
// belongs under the original rule, not under this exception.
//
// Data inputs (via ActionInputSpec):
//   - site_id (required) — resolved from collectedData via path
//
// Config literals (read directly, NOT through ExtractActionInputs):
//   - item_pipeline (optional) — filter by work item pipeline (e.g. "build")
//     Named item_pipeline to avoid collision with site_record.pipeline
//   - handler_agent (optional) — filter by handler agent type
//   - max_items (optional, default 50)
func LoadWorkItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("LoadWorkItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// site_id needs path resolution from collectedData
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		LoadWorkItemsInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	siteIDStr := inputs.Get("site_id")
	siteID, err := uuid.Parse(siteIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid site_id: %w", err)
	}

	// Filter params are config literals — read directly to avoid
	// nested lookup collisions (e.g. "pipeline" → site_record.pipeline)
	config := params.StepConfig.Config
	pipelineFilter, _ := config["item_pipeline"].(string)
	// Also accept old key for backwards compat during transition:
	if pipelineFilter == "" {
		pipelineFilter, _ = config["item_domain"].(string)
	}
	handlerFilter, _ := config["handler_agent"].(string)
	maxItems := datahelpers.GetIntField(config, "max_items", 50)

	query := `
		SELECT
			wi.id, wi.site_id, wi.source, wi.pipeline, wi.item_type,
			wi.severity, wi.summary, wi.spec, wi.page_id,
			wi.priority, wi.handler_agent, wi.status, wi.item_key,
			wi.batch_id, wi.attempt_count,
			COALESCE(wi.approval_mode, 'auto') as approval_mode,
			wi.component_id, wi.entity_id, wi.affected_url
		FROM site_work_items wi
		WHERE wi.site_id = $1
		  AND wi.status IN ('triaged', 'approved')
		  AND wi.attempt_count < wi.max_attempts
		  AND (COALESCE(wi.approval_mode, 'auto') = 'auto' OR wi.status = 'approved')
		  AND (
		    wi.depends_on IS NULL 
		    OR NOT EXISTS (
		      SELECT 1 FROM unnest(wi.depends_on) dep_id
		      WHERE dep_id NOT IN (
		        SELECT id FROM site_work_items
		        WHERE site_id = $1 AND status IN ('complete', 'verified')
		      )
		    )
		  )
	`

	args := []interface{}{siteID}
	argIdx := 2

	if pipelineFilter != "" {
		query += fmt.Sprintf(" AND wi.pipeline = $%d", argIdx)
		args = append(args, pipelineFilter)
		argIdx++
	}

	if handlerFilter != "" {
		query += fmt.Sprintf(" AND wi.handler_agent = $%d", argIdx)
		args = append(args, handlerFilter)
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY wi.priority ASC, wi.created_at ASC LIMIT $%d", argIdx)
	args = append(args, maxItems)

	rows, err := params.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var workItems []interface{}
	handlerSet := make(map[string]bool)
	droppedRows := 0

	for rows.Next() {
		var (
			id, wiSiteID               uuid.UUID
			source, pipeline, itemType string
			severity, summary          string
			specJSON                   []byte
			pageID                     *uuid.UUID
			priority                   int
			status                     string
			itemKey                    *string
			batchID                    *uuid.UUID
			attemptCount               int
			approvalMode               string
			// First-class routing columns — all nullable, hence pointers.
			componentID, entityID *uuid.UUID
			affectedURL           sql.NullString
		)
		// Nullable in the schema until migration 217 (bugs_closed/078) gave it
		// NOT NULL DEFAULT ''. Scanned as NullString regardless: a plain string
		// here turned one NULL row into a fleet-wide build outage twice in 24
		// hours, because the failed scan dropped the row via the `continue`
		// below while find_dispatchable_site still counted it — so the trigger
		// re-picked the same site every 120s forever. Never re-tighten this.
		var handlerAgent sql.NullString

		err := rows.Scan(
			&id, &wiSiteID, &source, &pipeline, &itemType,
			&severity, &summary, &specJSON, &pageID,
			&priority, &handlerAgent, &status, &itemKey,
			&batchID, &attemptCount, &approvalMode,
			&componentID, &entityID, &affectedURL,
		)
		if err != nil {
			// Error, not Warn: this branch drops a row that the site selector
			// has ALREADY counted as dispatchable, so the loop can complete
			// having claimed nothing while the row stays 'triaged'. A Warn
			// inside a `continue` is invisible at fleet scale (bugs_closed/078).
			droppedRows++
			logger.Error("LoadWorkItemsAction: work item row DROPPED on scan error — "+
				"site was selected as dispatchable but this item cannot be loaded",
				zap.Error(err),
				zap.String("site_id", siteID.String()),
			)
			continue
		}

		// Parse spec into map so current_item.spec.name works in dot-notation
		var specData interface{}
		if len(specJSON) > 0 {
			var specMap map[string]interface{}
			if err := json.Unmarshal(specJSON, &specMap); err == nil {
				// bugs_open/271: the guidance spelling four emitters wrote has
				// no reader; alias it onto the one every reader reads. Inside
				// the successful-unmarshal branch on purpose — a spec that
				// failed to parse is passed through as its raw string, exactly
				// as before.
				aliasGuidanceIntoSuggestion(specMap)
				specData = specMap
			} else {
				specData = string(specJSON)
			}
		}

		item := map[string]interface{}{
			"id":            id.String(),
			"site_id":       wiSiteID.String(),
			"source":        source,
			"pipeline":      pipeline,
			"item_type":     itemType,
			"severity":      severity,
			"summary":       summary,
			"spec":          specData,
			"priority":      priority,
			"handler_agent": handlerAgent.String, // "" when NULL — see the scan note above
			"status":        status,
			"attempt_count": attemptCount,
			"approval_mode": approvalMode,
		}

		// First-class routing columns, column-first with a spec.<key> fallback,
		// so ONE dispatcher path serves both populations. See the header note:
		// page_id is deliberately excluded, and spec is never written to.
		setRoutingField(item, "component_id", uuidPtrString(componentID), logger)
		setRoutingField(item, "entity_id", uuidPtrString(entityID), logger)
		setRoutingField(item, "affected_url", affectedURL.String, logger)

		if pageID != nil {
			item["page_id"] = pageID.String()
		}
		if itemKey != nil {
			item["item_key"] = *itemKey
		}
		if batchID != nil {
			item["batch_id"] = batchID.String()
		}

		workItems = append(workItems, item)
		handlerSet[handlerAgent.String] = true
	}

	var agents []string
	for agent := range handlerSet {
		agents = append(agents, agent)
	}

	logger.Info("LoadWorkItemsAction: Complete",
		zap.Int("items_loaded", len(workItems)),
		zap.Int("rows_dropped", droppedRows),
		zap.Strings("handler_agents", agents),
	)

	// Loading nothing while dropping rows is the livelock signature of
	// bugs_closed/078: the site was selected as dispatchable, the loop claims
	// nothing, the rows stay 'triaged', and the trigger re-picks this same site
	// on the next tick. Say so once, loudly, rather than leaving it to be found
	// by hand 22 minutes into an outage.
	if droppedRows > 0 && len(workItems) == 0 {
		logger.Error("LoadWorkItemsAction: LIVELOCK RISK — every candidate row was dropped, "+
			"so this site will be re-selected indefinitely with nothing to dispatch",
			zap.String("site_id", siteID.String()),
			zap.Int("rows_dropped", droppedRows),
		)
	}

	// first_item: convenience field for single-item dispatch (avoids array indexing)
	var firstItem interface{}
	if len(workItems) > 0 {
		firstItem = workItems[0]
	}

	return map[string]interface{}{
		"items":          workItems,
		"item_count":     len(workItems),
		"has_items":      len(workItems) > 0,
		"handler_agents": agents,
		"first_item":     firstItem,
		// Surfaced into collected_data so "the selector called this site
		// dispatchable and the loader loaded nothing" is answerable from the
		// orchestration row, instead of only from pod logs that have rotated.
		"rows_dropped": droppedRows,
	}, nil
}

// ============================================================================
// ACTION: complete_work_item
// Used by: site-work-orchestrator loop (after page is built and committed)
// ============================================================================
func CompleteWorkItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("CompleteWorkItemAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		CompleteWorkItemInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	itemIDStr := inputs.Get("work_item_id")

	// Fallback: config value may be a dot-notation path (e.g. "current_fix_item.id")
	// that ExtractActionInputs returned as a literal. Resolve against collectedData.
	if itemIDStr == "" || strings.Contains(itemIDStr, ".") {
		if resolved := resolveConfigPath(params.StepConfig.Config, "work_item_id", params.CollectedData, logger); resolved != "" {
			itemIDStr = resolved
		}
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id: %w", err)
	}

	resultData := inputs.GetMap("result")
	if resultData == nil {
		resultData = make(map[string]interface{})
	}
	if sha := inputs.Get("commit_sha"); sha != "" {
		resultData["commit_sha"] = sha
	}

	agentType := "unknown"
	if params.ExecutionContext.Sender.AgentType != "" {
		agentType = params.ExecutionContext.Sender.AgentType
	}

	// Completion gate 1: a response arriving is not the work succeeding. The
	// dispatch loop calls complete_work_item for every handler saga that came
	// back, but the envelope's response_status only records DELIVERY — the
	// saga's own verdict lives at response.status. A workflow that never ran at
	// all (unregistered or mistyped action → WORKFLOW_INVALID) returns
	// response.status='failed', and used to be stamped 'complete' alongside the
	// very error proving nothing was done. The item_key dedup then suppressed
	// re-detection, so the loop believed the defect class was handled: 54 items
	// across 6 sites by the 2026-07-18 sweep. This runs before the verifier
	// because a saga that failed outright is not worth verifying.
	detail, failed, unknownVerdict := handlerReportedFailure(resultData)
	if unknownVerdict != "" {
		recordUnknownVerdict(ctx, params, itemID, unknownVerdict, logger)
	}
	if failed {
		failedJSON, mErr := json.Marshal(resultData)
		if mErr != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", mErr)
		}
		return failUnverifiedCompletion(ctx, params.DB, itemID, agentType, string(failedJSON),
			"completion blocked: handler saga reported failure: "+detail,
			"handler_reported_failure", logger)
	}

	// Completion gates 1b and 2, both keyed on the item's TYPE so they share one
	// lookup (see verifyBeforeComplete):
	//   1b — the handler reported it changed nothing, for the item types that have
	//        opted in. A saga that altered no artefact cannot have repaired one, and
	//        for the type opted in today the route provably cannot repair it at all
	//        (bugs_open/213 D1, complete_work_item_no_change.go).
	//   2  — for item types with a registered verifier, confirm the defect is
	//        actually gone. A handler saga can return success without touching the
	//        defect (no-op paths exit through success-labelled complete_workflow
	//        steps) — see complete_work_item_verification.go for the policy.
	// resultData is passed in because the handler's own report is not yet stored:
	// this function marshals it below, AFTER the gates.
	verification, mayComplete, abstained := verifyBeforeComplete(ctx, params.DB, itemID, resultData, logger)
	if verification != nil {
		resultData["_verification"] = verification
	}
	if abstained != nil {
		// Gate 1b opted in for this type but could not read the payload. The item
		// still completes — an unreadable payload is not evidence of a no-op — so
		// this is recorded on a queryable surface rather than only logged, for the
		// same reason recordUnknownVerdict is (a pod log does not survive a roll).
		recordUnknownNoChangeShape(ctx, params, itemID, *abstained, logger)
	}

	resultJSON, err := json.Marshal(resultData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	if !mayComplete {
		// Two distinct causes reach here since RFC_017 (the defect persisting, and
		// verification being unable to run at all under a fail-closed type); the
		// message and reason code must say which, because this text is what a human
		// reads off the item's error column.
		msg, reason := blockedCompletionReason(verification)
		return failUnverifiedCompletion(ctx, params.DB, itemID, agentType, string(resultJSON),
			msg, reason, logger)
	}

	// Guard: do NOT overwrite a status a handler deliberately set to a flagged
	// or terminal state. The dispatch loop calls complete_work_item on every
	// successful handler saga, and page-build-handler's complete_error is a
	// SUCCESS-labelled complete_workflow — so without this guard a handler that
	// flagged its item needs_human_review (mark_needs_review for validation
	// failures; mark_no_sections for a sectionless page with no sibling layout)
	// would be re-stamped 'complete' here, silently undoing the flag. Completing
	// from an in-progress status (claimed/triaged/approved/detected/…) is
	// unaffected; only deliberate flags/terminals are preserved.
	res, err := params.DB.ExecContext(ctx, `
		UPDATE site_work_items
		SET status = 'complete',
		    result = $2::jsonb,
		    completed_at = NOW(),
		    handled_by = $3
		WHERE id = $1
		  AND status NOT IN ('needs_human_review','failed','unresolved','rejected','wont_fix','verified','blocked')
	`, itemID, string(resultJSON), agentType)
	if err != nil {
		return nil, fmt.Errorf("failed to complete work item: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		logger.Info("CompleteWorkItemAction: skipped — item already in a flagged/terminal status, not overwriting",
			zap.String("item_id", itemIDStr))
		return map[string]interface{}{
			"completed": false,
			"item_id":   itemIDStr,
			"reason":    "already_flagged_or_terminal",
		}, nil
	}

	logger.Info("CompleteWorkItemAction: Done", zap.String("item_id", itemIDStr))

	return map[string]interface{}{
		"completed": true,
		"item_id":   itemIDStr,
	}, nil
}

// ============================================================================
// ACTION: fail_work_item
// ============================================================================
func FailWorkItemAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger
	logger.Info("FailWorkItemAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		FailWorkItemInputSpec,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("input extraction failed: %w", err)
	}

	itemIDStr := inputs.Get("work_item_id")

	// Fallback: config value may be a dot-notation path (e.g. "current_fix_item.id")
	// that ExtractActionInputs returned as a literal. Resolve against collectedData.
	if itemIDStr == "" || strings.Contains(itemIDStr, ".") {
		if resolved := resolveConfigPath(params.StepConfig.Config, "work_item_id", params.CollectedData, logger); resolved != "" {
			itemIDStr = resolved
		}
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid work_item_id: %w", err)
	}

	// error_message is a config literal (e.g. "Content review not approved"),
	// not a path — read directly from config to avoid path resolution
	// Start with the config literal as default
	errorMsg, _ := params.StepConfig.Config["error_message"].(string)

	// routeToErrorStep stores this at __step_error.message in collected_data.
	// Prefer the actual error from the failed step (stored by routeToErrorStep)
	if stepError := datahelpers.ExtractNestedFieldString(params.CollectedData, "__step_error.message"); stepError != "" {
		errorMsg = stepError
		logger.Info("FailWorkItemAction: using real error from __step_error",
			zap.String("error", errorMsg))
	}

	agentType := "unknown"
	if params.ExecutionContext.Sender.AgentType != "" {
		agentType = params.ExecutionContext.Sender.AgentType
	}

	// Check for status_override — used for HITL gates like needs_human_review.
	// When set, the item gets that status directly without incrementing attempt_count.
	statusOverride, _ := params.StepConfig.Config["status_override"].(string)

	var newStatus string
	var attemptsLeft int

	if statusOverride != "" {
		err = params.DB.QueryRowContext(ctx, `
			UPDATE site_work_items
			SET error = $2,
			    status = $3,
			    handled_by = $4
			WHERE id = $1
			RETURNING status, max_attempts - attempt_count
		`, itemID, errorMsg, statusOverride, agentType).Scan(&newStatus, &attemptsLeft)

		logger.Info("FailWorkItemAction: using status_override",
			zap.String("item_id", itemIDStr),
			zap.String("status_override", statusOverride))
	} else if isAIUnavailable(fmt.Errorf("%s", errorMsg)) {
		// AI endpoint unavailable (connection refused, credit exhaustion, etc.)
		// Release back to triaged WITHOUT incrementing attempt_count.
		// The item will be retried when the endpoint recovers.
		err = params.DB.QueryRowContext(ctx, `
			UPDATE site_work_items
			SET error = $2,
			    status = 'triaged',
			    claimed_by = NULL,
			    claimed_at = NULL,
			    handled_by = NULL,
			    updated_at = NOW()
			WHERE id = $1
			RETURNING status, max_attempts - attempt_count
		`, itemID, "AI unavailable: "+errorMsg).Scan(&newStatus, &attemptsLeft)

		logger.Info("FailWorkItemAction: AI unavailable — released to triaged without counting attempt",
			zap.String("item_id", itemIDStr),
			zap.String("error", errorMsg))
	} else {
		err = params.DB.QueryRowContext(ctx, `
			UPDATE site_work_items
			SET attempt_count = attempt_count + 1,
			    error = $2,
			    status = CASE
			        WHEN attempt_count + 1 >= max_attempts THEN 'failed'
			        ELSE 'triaged'
			    END,
			    handled_by = $3
			WHERE id = $1
			RETURNING status, max_attempts - (attempt_count)
		`, itemID, errorMsg, agentType).Scan(&newStatus, &attemptsLeft)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update work item: %w", err)
	}

	logger.Info("FailWorkItemAction: Done",
		zap.String("item_id", itemIDStr),
		zap.String("new_status", newStatus),
		zap.Int("attempts_left", attemptsLeft),
	)

	return map[string]interface{}{
		"item_id":       itemIDStr,
		"new_status":    newStatus,
		"attempts_left": attemptsLeft,
		"will_retry":    newStatus == "triaged",
	}, nil
}

// ============================================================================
// Internal helpers
// ============================================================================

// resolveConfigPath checks if a config field contains a dot-notation path
// (e.g. "current_fix_item.id") and resolves it against collectedData.
// Used by CompleteWorkItemAction and FailWorkItemAction where config values
// may be paths into loop variables that ExtractActionInputs doesn't resolve.
func resolveConfigPath(config map[string]interface{}, field string, collectedData map[string]interface{}, logger *zap.Logger) string {
	pathStr, ok := config[field].(string)
	if !ok || !strings.Contains(pathStr, ".") {
		return ""
	}

	val := getNestedValue(collectedData, pathStr)
	if val == nil {
		logger.Warn("resolveConfigPath: path not found in collectedData",
			zap.String("field", field),
			zap.String("path", pathStr))
		return ""
	}

	if str, ok := val.(string); ok {
		logger.Info("resolveConfigPath: resolved",
			zap.String("field", field),
			zap.String("path", pathStr),
			zap.String("value", str))
		return str
	}
	return ""
}

type workItem struct {
	siteID       uuid.UUID
	source       string
	pipeline     string
	itemType     string
	severity     string
	summary      string
	spec         string
	pageID       *uuid.UUID
	componentID  *uuid.UUID
	priority     int
	handlerAgent string
	status       string
	createdBy    string
	itemKey      string
	batchID      uuid.UUID
	dependsOn    []uuid.UUID

	// parentItemID links this item to the item that caused it to be raised.
	//
	// It exists because its ABSENCE is why three call sites walked round the
	// shared door. `apply_gap_plan_action.go` needs `parent_item_id` on the items
	// it raises, this struct had no field for it, so all three hand-rolled their
	// own `INSERT … ON CONFLICT DO NOTHING` — inheriting none of the anti-churn,
	// none of the honest boolean, and a bare conflict target that does not match
	// idx_swi_dedup's partial predicate. Two council seats on correlation
	// a5b70424 objected to exactly that (bugs_open/091). A missing field on a
	// shared helper is not a small gap: it is a standing invitation to fork it.
	parentItemID *uuid.UUID

	// recurrenceExpected marks an item whose RE-REQUEST is normal, rather than
	// evidence that previous handlers failed to fix something.
	//
	// The anti-churn block below (within-cycle suppression + the two-strike
	// label) reads a repeat item_key as "we keep fixing this and it keeps coming
	// back". That is right for a DETECTED DEFECT, which recurs only because a
	// detector found it again. It is wrong for an ACTION REQUEST — "re-render
	// this page after I changed its template" — where a completed predecessor
	// means the request SUCCEEDED, and asking again is the normal course of
	// business, not a third strike.
	//
	// Conflating the two silently broke the tool fix loop: two SUCCESSFUL
	// re-renders poisoned a shared item_key, so every later re-render on that
	// site was born 'unresolved' and never dispatched, and durable template
	// fixes never reached the live page for three cycles (bugs_open/024).
	//
	// Opt-in, default false: existing callers keep today's behaviour exactly.
	// Dedup is NOT waived by this flag — idx_swi_dedup still refuses a second
	// OPEN item for the same (site_id, item_key). Only the anti-churn heuristics
	// are skipped.
	recurrenceExpected bool
}

// conflictPolicy decides what happens when a keyed item arrives and an OPEN item
// already holds that key.
//
// It is a PARAMETER of writeWorkItem rather than a field on workItem, and that is
// deliberate. As a field, a caller could set it and still call insertWorkItem —
// whose single bool cannot express a refresh, so the caller would silently be told
// "nothing happened" about a write that did happen. As a parameter, the old
// function cannot receive it and the mistake does not compile.
type conflictPolicy int

const (
	// dropOnConflict is today's behaviour, unchanged and the default: the open
	// item wins and the arriving finding is discarded. Right when the second
	// finding SAYS THE SAME THING as the first — which is the common case, and
	// why this is the default.
	dropOnConflict conflictPolicy = iota

	// refreshOnConflict brings the open item's DESCRIPTION up to date instead of
	// discarding the finding. For a detector keyed per SITE rather than per
	// finding, a second, DIFFERENT finding arriving while an item is open is not
	// a duplicate — dropping it leaves a durable record that describes something
	// that is no longer true, and the human who opens it is sent to the wrong
	// place (bugs_open/091: four of five open stale_evidence items named the
	// wrong facts on 2026-08-02).
	//
	// The row stays SINGLE — dedup is not waived, and this is not a route to
	// filling the needs_human_review queue (bugs_open/033's owner ruling).
	refreshOnConflict
)

// workItemWrite is the honest outcome of a work-item write. It exists because a
// bool cannot carry it: `ON CONFLICT … DO UPDATE` affects a row, so RowsAffected
// would read "true" for a write that created nothing — reintroducing, inside
// 091's own fix, the exact dishonesty 091 was filed about.
//
// Recorded() is what a caller usually wants ("does a durable record now describe
// my finding?"). Inserted is what a field NAMED "created" must be set from.
type workItemWrite struct {
	// Inserted: a NEW row was written.
	Inserted bool
	// Refreshed: an existing OPEN row's summary/spec were brought up to date.
	// Only reachable under refreshOnConflict.
	Refreshed bool
	// BornBlocked: the row WAS written, but at status 'blocked' instead of the
	// dispatchable status the caller asked for, because its handler_agent names
	// no registered agent (bugs_open/291). The finding is durably recorded —
	// Inserted stays true — and feasibility-recheck promotes the row the moment
	// an agent of that name is registered.
	BornBlocked bool
}

// Recorded reports whether a durable record now describes this finding, however
// it got there.
func (w workItemWrite) Recorded() bool { return w.Inserted || w.Refreshed }

// insertWorkItem writes a work item and reports whether a NEW ROW was created.
// It is the entry point for every caller that does not care about refreshes,
// which is all of them but one; its behaviour is byte-identical to before the
// conflict policy existed.
func insertWorkItem(ctx context.Context, tx *sql.Tx, item workItem, logger *zap.Logger) (bool, error) {
	w, err := writeWorkItem(ctx, tx, item, dropOnConflict, logger)
	return w.Inserted, err
}

// writeWorkItem is insertWorkItem with the conflict policy made explicit and the
// outcome told in full. See conflictPolicy for why the policy is a parameter.
func writeWorkItem(ctx context.Context, tx *sql.Tx, item workItem, policy conflictPolicy, logger *zap.Logger) (workItemWrite, error) {
	var batchIDPtr *uuid.UUID
	if item.batchID != uuid.Nil {
		batchIDPtr = &item.batchID
	}

	// --- Two-strike rule: suppress within-cycle, label unresolved after 2 attempts ---
	// Skipped entirely for recurrence-expected items (see workItem.recurrenceExpected):
	// for an action request a terminal predecessor is a SUCCESS, not a strike, and
	// suppressing the re-request drops the very work that success depends on.
	if item.itemKey != "" && !item.recurrenceExpected {
		var terminalCount int
		var newestAge float64 // hours since most recent terminal item

		err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*),
			       COALESCE(EXTRACT(EPOCH FROM (NOW() - MAX(created_at))) / 3600.0, 999)
			FROM site_work_items
			WHERE site_id = $1
			  AND item_key = $2
			  AND status IN ('complete', 'failed')
			  AND created_at > NOW() - INTERVAL '7 days'
		`, item.siteID, item.itemKey).Scan(&terminalCount, &newestAge)

		if err == nil && terminalCount > 0 {
			// Within-cycle suppress: most recent terminal item is < 3h old
			if newestAge < 3.0 {
				logger.Info("insertWorkItem: suppressed — terminal item too recent",
					zap.String("item_key", item.itemKey),
					zap.Float64("age_hours", newestAge),
				)
				return workItemWrite{}, nil
			}

			// Two strikes: handler had 2 chances and the issue persists.
			// Mark as unresolved so it's visible for investigation.
			if terminalCount >= 2 {
				logger.Info("insertWorkItem: marking unresolved — 2 prior attempts did not fix the issue",
					zap.String("item_key", item.itemKey),
					zap.Int("prior_attempts", terminalCount),
					zap.String("original_handler", item.handlerAgent),
				)
				item.status = "unresolved"
				item.summary = fmt.Sprintf("[unresolved after %d attempts] %s", terminalCount, item.summary)
			}
		}
	}

	// --- Routability at the door (bugs_open/291) ---
	// A row born in a dispatchable status whose handler_agent names no registered
	// agent can only ever become claim's handler-not-registered `blocked` — after
	// a wasted dispatch pick and a misleading interval at 'triaged'. Stamp it
	// `blocked` at birth instead, with the SAME error text claim writes, so the
	// outcome is identical but earlier, cheaper, and honest. This runs AFTER the
	// two-strike block so it sees the final status.
	//
	// Demote, never refuse: a refusal here would lose the finding to a pod log —
	// discovery sweeps log-and-continue on insert error, and config-driven
	// create_work_item loops run continue_on_error — whereas a blocked row is
	// durable, visible (idx_work_items_blocked), and self-healing: the
	// feasibility-recheck task (600s) promotes it to 'triaged' the moment an
	// agent of that name is registered, which also absorbs the platform's
	// image-before-seeds ordering. Claim's own branch remains the backstop for
	// the raw-INSERT writers that bypass this door.
	//
	// The probe is claim's own predicate, rendered from the one shared helper —
	// see workItemHandlerRegisteredSQL for why it deliberately ignores
	// is_active/is_snapshot. The trigger set is exactly CHECK 443's statuses;
	// see workItemStatusRequiresRegisteredHandler for why it must not widen.
	// Kill-switch (council round 1, guardian seat): a rollback lever for the
	// guard itself, independent of any binary roll. Ships ARMED — this is not an
	// opt-in (the owner has ruled against default-OFF switches that rot
	// unexercised); it exists so an operator can disarm the door probe fleet-wide
	// with a redeploy-free env change if it ever misbehaves, falling back to the
	// claim-time backstop, which is the exact pre-guard behaviour.
	demotedUnroutable := false
	if item.handlerAgent != "" && workItemStatusRequiresRegisteredHandler(item.status) &&
		os.Getenv("DISABLE_UNREGISTERED_HANDLER_DEMOTION") == "" {
		var registered bool
		if probeErr := tx.QueryRowContext(ctx,
			"SELECT "+workItemHandlerRegisteredSQL("$1"), item.handlerAgent,
		).Scan(&registered); probeErr != nil {
			// Mirror claim's posture on a failed handler read: log and fall
			// through — the claim-time branch remains the backstop.
			logger.Warn("writeWorkItem: handler registration probe failed, skipping door guard",
				zap.String("handler_agent", item.handlerAgent),
				zap.Error(probeErr))
		} else if !registered {
			logger.Warn("writeWorkItem: handler not registered — item born blocked",
				zap.String("handler_agent", item.handlerAgent),
				zap.String("item_type", item.itemType),
				zap.String("item_key", item.itemKey),
				zap.String("requested_status", item.status))
			item.status = "blocked"
			demotedUnroutable = true
		}
	}

	var itemKeyPtr *string
	if item.itemKey != "" {
		itemKeyPtr = &item.itemKey
	}

	var dependsOnStr *string
	if len(item.dependsOn) > 0 {
		ids := "{"
		for i, id := range item.dependsOn {
			if i > 0 {
				ids += ","
			}
			ids += id.String()
		}
		ids += "}"
		dependsOnStr = &ids
	}

	args := []interface{}{
		item.siteID, item.source, item.pipeline, item.itemType, item.severity,
		item.summary, item.spec,
		item.pageID, item.componentID, item.priority, item.handlerAgent, item.status, item.createdBy,
		itemKeyPtr, batchIDPtr, dependsOnStr,
	}

	// parent_item_id — and, since bugs_open/291, error — are appended ONLY when
	// needed, so a caller that never asked for either sends the identical
	// statement with the identical sixteen arguments it always sent. That is not
	// tidiness — WIDENING A SHARED STATEMENT BREAKS CALLERS NOBODY EDITED. Adding
	// a column unconditionally failed twenty tests across eight files in lanes
	// other sessions are working right now (sqlmock matches the argument COUNT),
	// none of which had anything to do with this change. A seam whose cost lands
	// on people who did not opt into it is the drift this repo's council exists
	// to refuse.
	extraCols, extraVals, argN := "", "", 16
	if item.parentItemID != nil {
		argN++
		extraCols += ", parent_item_id"
		extraVals += fmt.Sprintf(", $%d", argN)
		args = append(args, *item.parentItemID)
	}
	if demotedUnroutable {
		argN++
		extraCols += ", error"
		extraVals += fmt.Sprintf(", $%d", argN)
		args = append(args, "Handler agent not registered: "+item.handlerAgent)
	}

	insertSQL := fmt.Sprintf(`
		INSERT INTO site_work_items (
			site_id, source, pipeline, item_type, severity, summary, spec,
			page_id, component_id, priority, handler_agent, status, created_by,
			item_key, batch_id, depends_on%s
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7::jsonb,
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16::uuid[]%s
		)
		ON CONFLICT (site_id, item_key)
			WHERE item_key IS NOT NULL
			  AND status NOT IN (%s)
		DO NOTHING
	`, extraCols, extraVals, sqlInList(workItemTerminalStatuses))

	result, err := tx.ExecContext(ctx, insertSQL, args...)

	if err != nil {
		return workItemWrite{}, fmt.Errorf("insert failed for %s: %w", item.itemKey, err)
	}

	rows, _ := result.RowsAffected()
	if rows > 0 {
		logger.Info("Work item inserted",
			zap.String("item_key", item.itemKey),
			zap.String("handler", item.handlerAgent),
			zap.String("status", item.status),
			zap.Int("priority", item.priority),
		)
		return workItemWrite{Inserted: true, BornBlocked: demotedUnroutable}, nil
	}

	// Nothing was inserted: an OPEN item already holds this key. Under the
	// default policy that is the end of it — and the FINDING is gone with it.
	if policy != refreshOnConflict {
		return workItemWrite{}, nil
	}
	return refreshOpenWorkItem(ctx, tx, item, logger)
}

// workItemHeldStatuses is the set of statuses at which a handler has actually
// TAKEN an item and may be acting on its spec right now. It is the complement of
// "queued but unowned" — `triaged`, `approved`, `needs_human_review` and
// `detected` are all waiting for someone, so none of them is held.
//
// Used by refreshOpenWorkItem (bugs_open/091): a repeat finding may bring an open
// item's description up to date, but never one that is being worked, because the
// handler read the spec when it claimed the row and would not see the change.
//
// This is NOT interchangeable with workItemTerminalStatuses and must never be
// folded into it: a held item is emphatically not done, and adding it there would
// take it out of idx_swi_dedup's predicate and permit a second OPEN row for the
// same key — the very thing the dedup exists to stop.
//
// ⚠ IT BELONGS IN work_items_common.go, beside workItemTerminalStatuses and
// workItemDispatchableStatuses, and it is here only because that file had another
// session's uncommitted work in it when this landed (2026-08-03) and a pathspec
// commit cannot leave a same-file passenger behind. Move it when that file is
// clean; the lists are easier to keep in lockstep when they are visible together.
var workItemHeldStatuses = []string{
	"claimed",
	"diagnosing",
}

// refreshOpenWorkItemSQL is a function rather than an inline string so the test
// that asserts the guard clauses can read the STATEMENT THAT RUNS, instead of
// re-deriving its own copy and agreeing with itself.
//
// THE STATUS LISTS ARE PARAMETERS, NOT INTERPOLATED TEXT — council `8e7357ae`,
// constitution seat: *"SCHEMA FIRST, PARAMETERISED ALWAYS requires values pass as
// query parameters, never interpolated into a template string."* It is right, and
// the objection is worth more than the one line it costs: `sqlInList` exists only
// because insertWorkItem's `ON CONFLICT … WHERE` predicate MUST be literal text
// for partial-index inference to resolve idx_swi_dedup. This is a plain `WHERE` —
// it has no such constraint, so the exemption does not extend to it. The array
// literal is built the way `depends_on` already is a few lines above, which is the
// existing machinery for exactly this.
//
// It RETURNS the refreshed row's status, which is not decoration: it is what lets
// the writer say WHICH row it changed without a second query, and so warn when a
// refresh lands on an item sitting in the human-review queue (see below).
func refreshOpenWorkItemSQL() string {
	return `
		UPDATE site_work_items
		   SET summary    = $3,
		       spec       = $4::jsonb,
		       updated_at = NOW()
		 WHERE site_id  = $1
		   AND item_key = $2
		   AND status <> ALL($5::text[])
		   AND status <> ALL($6::text[])
	 RETURNING status
	`
}

// sqlTextArrayLiteral formats a Go string slice as a Postgres array literal for
// use as a QUERY PARAMETER (`$n::text[]`), not for interpolation into SQL text.
// Same shape as the `depends_on` literal insertWorkItem already builds.
func sqlTextArrayLiteral(values []string) string {
	return "{" + strings.Join(values, ",") + "}"
}

// refreshOpenWorkItem brings the OPEN item holding this key up to date with the
// finding that just arrived, and reports whether it actually did.
//
// A SEPARATE STATEMENT, not a `DO UPDATE` clause on the insert above, for two
// reasons that are worth more than the extra round trip:
//
//   - the insert every other caller uses stays byte-identical, so this seam
//     cannot change what happens to the ~20 call sites that never asked for it;
//   - `DO UPDATE` makes RowsAffected() 1 for an update, so the single bool
//     insertWorkItem returns would start reading "true" for a write that created
//     nothing. That is precisely the false report bugs_open/091 was filed about,
//     arriving through 091's own fix.
//
// It updates the DESCRIPTION and nothing else. status, priority, handler_agent
// and severity may all have been moved by a human on the open row; the arriving
// finding owns what it observed, not how somebody decided to handle it.
//
// THE PREDICATE IS THE CONCURRENCY GUARD. No row is locked between the failed
// insert and this update, so the item may go terminal in between. It is not
// re-opened: the same terminal-status list that idx_swi_dedup uses means a
// terminal row simply does not match, zero rows update, and the caller is told
// nothing was recorded — which is true, and which the next sweep will fix by
// inserting cleanly. A row a handler currently HOLDS (claimed, diagnosing) is
// skipped for the same reason in reverse: changing the spec under a running
// handler is a different bad state, and this seam must not create one.
// THE NO-OP IS LOUD, AND IT IS LOUD HERE RATHER THAN AT THE CALLER — council
// `8e7357ae`, bug_historian seat: *"the ONLY fail-loud surface for that case is the
// one caller's own log.Warn … any future refreshOnConflict adopter who doesn't
// replicate the Recorded()-check-and-warn pattern reproduces 091 silently."* That
// is 091's own shape one level down — a shared mechanism whose correctness depends
// on each caller remembering something — so the warning belongs to the writer. A
// caller may add its own; it can no longer be the only one.
func refreshOpenWorkItem(ctx context.Context, tx *sql.Tx, item workItem, logger *zap.Logger) (workItemWrite, error) {
	if item.itemKey == "" {
		return workItemWrite{}, nil
	}

	var refreshedStatus string
	err := tx.QueryRowContext(ctx, refreshOpenWorkItemSQL(),
		item.siteID, item.itemKey, item.summary, item.spec,
		sqlTextArrayLiteral(workItemTerminalStatuses),
		sqlTextArrayLiteral(workItemHeldStatuses),
	).Scan(&refreshedStatus)

	if errors.Is(err, sql.ErrNoRows) {
		logger.Warn("writeWorkItem: FINDING NOT RECORDED — nothing inserted and nothing "+
			"refreshed. An item already holds this key and could not be updated: it went "+
			"terminal between the two statements, or a handler is holding it. The durable "+
			"record still describes the EARLIER finding (bugs_open/091)",
			zap.String("item_key", item.itemKey),
			zap.String("item_type", item.itemType),
			zap.String("site_id", item.siteID.String()),
		)
		return workItemWrite{}, nil
	}
	if err != nil {
		return workItemWrite{}, fmt.Errorf("refresh failed for %s: %w", item.itemKey, err)
	}

	// A refresh that lands on a human-review row rewrites what someone may be
	// reading. It is deliberate — needs_human_review is a queue, not a claim, and
	// leaving the record FALSE is worse — but it is the one judgement in this seam
	// the council asked to see confirmed rather than assumed (guardian, architecture),
	// so it says so every time it happens rather than only in a design document.
	if refreshedStatus == "needs_human_review" {
		logger.Warn("writeWorkItem: refreshed the open work item with a newer finding — "+
			"the row is in the HUMAN-REVIEW queue, so its description changed under any "+
			"reader. Deliberate (bugs_open/091): a record that is silently WRONG is worse "+
			"than one that silently changes",
			zap.String("item_key", item.itemKey),
			zap.String("item_type", item.itemType),
		)
		return workItemWrite{Refreshed: true}, nil
	}

	logger.Info("writeWorkItem: refreshed the open work item with a newer finding",
		zap.String("item_key", item.itemKey),
		zap.String("item_type", item.itemType),
		zap.String("status", refreshedStatus),
	)
	return workItemWrite{Refreshed: true}, nil
}

// normalizePageType applies the same kebab-case normalization used for
// component functions. LLMs sometimes output entity_directory instead of
// entity-directory, or BlogIndex instead of blog-index.
func normalizePageType(pt string) string {
	pt = strings.ToLower(pt)
	pt = strings.ReplaceAll(pt, "_", "-")
	pt = strings.ReplaceAll(pt, " ", "-")
	return pt
}

// ---------------------------------------------------------------------------
// bugs_open/210 — a logo/hero item is only well-formed if it carries a prompt
// ---------------------------------------------------------------------------

type imageBuildItemArgs struct {
	// db is the READ handle for the brand-identity default; tx is the insert
	// path. They are deliberately separate: the default's queries must not
	// join the work-item transaction, whose rollback would otherwise discard
	// nothing useful but would widen its lock window for a read that cannot
	// affect it.
	db        *sql.DB
	siteID    uuid.UUID
	batchID   uuid.UUID
	planData  map[string]interface{}
	purpose   string // assets.purpose — "logo" / "hero"
	promptKey string // key inside plan.image_prompts — "logo" / "hero_home"
	itemType  string
	itemKey   string
	summary   string
	severity  string
}

// imagePromptFromPlan reads one prompt out of a plan's image_prompts map.
// Returns "" when the map is absent, null, not an object, or the key is
// missing or non-string — every one of which reaches this code as a plain
// absence rather than an error, which is why the caller must check.
func imagePromptFromPlan(planData map[string]interface{}, promptKey string) string {
	prompts, ok := planData["image_prompts"].(map[string]interface{})
	if !ok {
		return ""
	}
	s, _ := prompts[promptKey].(string)
	return s
}

// insertImageBuildItem files a generation request, supplying the prompt the
// handler requires — from the plan when it has one, and from the site's brand
// identity when it does not.
//
// It can no longer file an item without a prompt, which is the defect
// bugs_open/210 records: image-build-handler's logo/hero branches map "prompt"
// as a REQUIRED field, so a promptless item dies at input extraction. The
// default comes from the owner's 2026-08-09 ruling — see the branch below.
func insertImageBuildItem(ctx context.Context, tx *sql.Tx, a imageBuildItemArgs, inserted *int, logger *zap.Logger) {
	prompt := imagePromptFromPlan(a.planData, a.promptKey)

	item := workItem{
		siteID: a.siteID, source: "planner", pipeline: "build",
		itemType: a.itemType, severity: a.severity,
		priority: 5, status: "triaged", createdBy: "site-planner",
		itemKey: a.itemKey, batchID: a.batchID,
	}

	spec := map[string]interface{}{
		"purpose":       a.purpose,
		"prompt_key":    a.promptKey,
		"image_prompts": a.planData["image_prompts"],
	}

	promptSource := checks.BrandPromptSourcePlanned
	if prompt == "" {
		// OWNER RULING 2026-08-09 (bugs_open/210): default rather than block.
		// ~2,000 domains to populate and no capacity to approve that many
		// logos, so the human path is an OVERRIDE, not a gate. Same helper and
		// same reasoning as check_placeholder_image_in_use — brand identity,
		// never the imagery direction that logos are excluded from.
		prompt = checks.DefaultBrandImagePrompt(ctx, a.db, a.siteID, a.purpose, logger)
		promptSource = checks.BrandPromptSourceDefault
		logger.Info("WriteBuildItemsAction: plan asks for an image but carries no prompt — "+
			"using the brand-identity default",
			zap.String("site_id", a.siteID.String()),
			zap.String("purpose", a.purpose),
			zap.String("prompt_key", a.promptKey))
	}

	// Both key shapes, matching check_unfulfilled_image_prompt's dual format:
	// the two legacy handler branches read the nested image_prompts.<key>, the
	// two Phase-2E branches read the flat spec.prompt. Writing both keeps this
	// producer correct under either, and makes the deferred convergence
	// (IMG-069's open question) a config-only change.
	spec["prompt"] = prompt
	spec["prompt_source"] = promptSource
	if _, ok := spec["image_prompts"].(map[string]interface{}); !ok {
		spec["image_prompts"] = map[string]interface{}{}
	}
	if m, ok := spec["image_prompts"].(map[string]interface{}); ok {
		m[a.promptKey] = prompt
	}
	specJSON, _ := json.Marshal(spec)
	item.spec = string(specJSON)
	item.summary = a.summary
	item.handlerAgent = "image-build-handler"

	ok, err := insertWorkItem(ctx, tx, item, logger)
	if err != nil {
		logger.Warn("WriteBuildItemsAction: image item insert error",
			zap.String("item_type", a.itemType), zap.Error(err))
		return
	}
	if ok {
		*inserted++
	}
}
