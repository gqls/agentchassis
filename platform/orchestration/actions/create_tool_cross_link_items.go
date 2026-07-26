// FILE: platform/orchestration/actions/create_tool_cross_link_items.go
//
// Tool cross-linking: content_rewrite work items that tell page-build-handler
// to weave a contextual reference to a tool into an EXISTING page's copy.
//
// bugs_open/029 — a tool page's URL CANNOT be constructed from the tool's
// function name. The three build paths produce three different shapes for the
// same input (`deploy_tool_to_site` strips the `tool-` prefix, the seeded
// library forks keep it, `CanonicalisePage` appends `/index.html`), so the URL
// has to be LOOKED UP from pages.url — and that lookup is only meaningful once
// the tool page row exists. Constructing `/tools/{function}.html` at SUGGESTION
// time, as this file used to, was wrong on all three shapes: 0 of 27 emitted
// items resolved to a real page, and page-content-writer obeyed the fabricated
// URL because the item's own acceptance_test demanded it.
//
// So the emitter moved to the tool BUILD paths, which have the real URL they
// just created:
//
//	emitToolCrossLinkItems — shared emitter. Takes the REAL pages.url of an
//	    existing tool page and never constructs one. Called from
//	    deploy_tool_action.go and create_tool_component_action.go after the
//	    page row is inserted.
//
//	CreateToolCrossLinkItemsAction — the original suggestion-time action, kept
//	    registered and made FAIL-SAFE: it now resolves each tool to a real page
//	    and skips the ones that have none, so it can no longer fabricate a URL
//	    wherever it is invoked. Its workflow step is removed from tool-suggester
//	    by migration 211; this action remains only so that a stale or restored
//	    config naming it cannot invalidate a workflow (bugs_closed/017) or
//	    resurrect the defect.
//
// Registration:
//
//	"create_tool_cross_link_items": {
//	    Handler:     CreateToolCrossLinkItemsAction,
//	    Category:    "site",
//	    Description: "Create content_rewrite items to cross-link tools from related pages",
//	    IsLocal:     true,
//	},

package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// Shared emitter
// ============================================================================

// toolCrossLinkRequest describes one tool whose existing pages should gain a
// contextual reference to it.
//
// toolPageURL is the page's ACTUAL pages.url — the caller must have read or
// written it, never derived it from toolFunction. An empty value is refused.
type toolCrossLinkRequest struct {
	siteID       uuid.UUID
	toolFunction string
	toolName     string
	toolDesc     string
	toolPageID   uuid.UUID
	toolPageURL  string
	relatedPages []string
	// emittedBy names the actor for the source/created_by columns, e.g.
	// "tool-deployer". Cross-link rows are identified by their item_key
	// prefix (tool_crosslink:), not by source — see itemKey below.
	emittedBy string
}

// crossLinkFailedStatuses are the work-item statuses from which a tool page's
// content build will not resume on its own. A cross-link gated behind such an
// item would wait forever, so we decline to emit instead — no link is a
// non-event, a permanent link to an undeployed page is bugs_open/029's damage
// arriving by a second route.
var crossLinkFailedStatuses = []string{"failed", "rejected", "cancelled", "wont_fix", "unresolved"}

// toolPageLive reports whether a page's build_status means the URL is already
// served. needs_rebuild counts: the page was deployed and is queued for a
// refresh, so the link resolves today.
func toolPageLive(buildStatus string) bool {
	return buildStatus == "deployed" || buildStatus == "needs_rebuild"
}

// emitToolCrossLinkItems creates one content_rewrite item per related page,
// each carrying the tool page's real URL.
//
// Two guards, both load-bearing for bugs_open/029:
//
//  1. The URL is a parameter, never constructed here. An empty or non-absolute
//     URL is refused outright rather than guessed at.
//
//  2. The items are GATED on the tool page actually going live. If the page is
//     not yet deployed, each emitted item depends_on the open work item that
//     will deploy it (load_work_item_actions.go only dispatches an item whose
//     depends_on rows are all complete/verified). If no such item exists — or
//     it has already failed — nothing is emitted, because we cannot show the
//     link will ever resolve. That is the leopardess failure: the tool was
//     never built, and the reference shipped anyway.
//
// A DECLINED emit is recorded durably, not just logged (agent_error_log via
// recordCrossLinkSkip). Every path that returns 0 is a cross-link the site was
// meant to get and did not, and "someone reads the pod logs" is not a
// surfacing mechanism — the council gate objected to exactly that, and the
// platform's own history (bugs_open/034) is of silent drops nobody could count.
//
// Returns the number of items created. Errors are logged and swallowed per
// item: cross-linking is a follow-on nicety and must never fail a tool build.
func emitToolCrossLinkItems(ctx context.Context, params ActionParams, logger *zap.Logger, req toolCrossLinkRequest) int {
	db := params.DB
	logger = logger.With(
		zap.String("tool", req.toolFunction),
		zap.String("tool_page_url", req.toolPageURL),
	)

	if req.toolFunction == "" || req.toolName == "" {
		return 0
	}
	if len(req.relatedPages) == 0 {
		// Not a defect: a suggestion may legitimately name no related pages.
		// Recorded at 'info' so the count is still visible, since "the LLM
		// stopped filling the field in" and "there was nothing to link" are
		// indistinguishable from the outside.
		logger.Info("emitToolCrossLinkItems: no related_pages on the suggestion, nothing to cross-link")
		recordCrossLinkSkip(ctx, params, logger, req, "no_related_pages", "info",
			"suggestion carried no related_pages, so no cross-links were emitted")
		return 0
	}
	if !strings.HasPrefix(req.toolPageURL, "/") {
		// Never guess. bugs_open/029 is exactly what guessing produced.
		logger.Warn("emitToolCrossLinkItems: refusing to emit without a real tool page URL",
			zap.String("received", req.toolPageURL))
		recordCrossLinkSkip(ctx, params, logger, req, "no_real_tool_page_url", "warning",
			"caller supplied no absolute tool page URL; refusing to construct one (bugs_open/029)")
		return 0
	}

	// --- Guard 2: prove the tool page will be served ---
	var buildStatus string
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(build_status, '') FROM pages WHERE id = $1`, req.toolPageID,
	).Scan(&buildStatus); err != nil {
		logger.Warn("emitToolCrossLinkItems: tool page row not readable, skipping", zap.Error(err))
		recordCrossLinkSkip(ctx, params, logger, req, "tool_page_unreadable", "warning", err.Error())
		return 0
	}

	var dependsOn []uuid.UUID
	if !toolPageLive(buildStatus) {
		var gateID uuid.UUID
		err := db.QueryRowContext(ctx, fmt.Sprintf(`
			SELECT id FROM site_work_items
			WHERE site_id = $1
			  AND page_id = $2
			  AND item_type = 'needs_content_page'
			  AND status NOT IN (%s)
			ORDER BY created_at DESC
			LIMIT 1
		`, sqlInList(crossLinkFailedStatuses)), req.siteID, req.toolPageID).Scan(&gateID)
		if err != nil {
			logger.Info("emitToolCrossLinkItems: tool page is not live and has no open build item — not emitting cross-links",
				zap.String("build_status", buildStatus),
				zap.Error(err))
			recordCrossLinkSkip(ctx, params, logger, req, "tool_page_will_not_go_live", "warning",
				fmt.Sprintf("tool page build_status=%q and no open needs_content_page item to gate on — cross-links withheld rather than pointed at a page that may never deploy", buildStatus))
			return 0
		}
		dependsOn = []uuid.UUID{gateID}
		logger.Info("emitToolCrossLinkItems: gating cross-links behind the tool page build",
			zap.String("build_status", buildStatus),
			zap.String("depends_on", gateID.String()))
	}

	// --- Resolve related page names to ids ---
	pageMap := make(map[string]uuid.UUID)
	rows, err := db.QueryContext(ctx, `
		SELECT name, id FROM pages
		WHERE site_id = $1 AND status = 'active'
	`, req.siteID)
	if err != nil {
		logger.Warn("emitToolCrossLinkItems: failed to load pages", zap.Error(err))
		recordCrossLinkSkip(ctx, params, logger, req, "page_map_unreadable", "warning", err.Error())
		return 0
	}
	for rows.Next() {
		var name string
		var id uuid.UUID
		if rows.Scan(&name, &id) == nil {
			pageMap[name] = id
		}
	}
	rows.Close()

	created := 0
	for _, pageName := range req.relatedPages {
		pageName = strings.TrimSpace(pageName)
		if pageName == "" {
			continue
		}

		// Tool-to-tool cross-linking isn't useful, and the tool's own page
		// must never link to itself.
		if strings.HasPrefix(pageName, "tool-") {
			logger.Info("emitToolCrossLinkItems: skipping tool page", zap.String("page_name", pageName))
			continue
		}

		pageID, exists := pageMap[pageName]
		if !exists {
			logger.Info("emitToolCrossLinkItems: related page not found, skipping",
				zap.String("page_name", pageName))
			continue
		}
		if pageID == req.toolPageID {
			continue
		}

		spec, _ := json.Marshal(map[string]interface{}{
			"page":        pageName,
			"page_name":   pageName, // dispatch loop maps spec.page_name → input_data.page_name
			"page_id":     pageID.String(),
			"description": fmt.Sprintf("Add a contextual reference to the %s tool on this page where it naturally fits the topic.", req.toolName),
			"suggestion": fmt.Sprintf(
				"Weave a natural reference to '%s' (%s) into the existing content. "+
					"Link to %s — use that URL exactly as written, do not alter or invent one. "+
					"The reference should feel like a helpful suggestion, not an advertisement — "+
					"e.g. 'Use our %s to see how this applies to your situation' or "+
					"'Try our %s to estimate your costs'. "+
					"Place it where it's most contextually relevant to what the page is discussing. "+
					"Do NOT add a new section — integrate it into existing paragraphs or after a relevant point.",
				req.toolName, req.toolDesc, req.toolPageURL,
				strings.ToLower(req.toolName), strings.ToLower(req.toolName),
			),
			"acceptance_test": fmt.Sprintf(
				"Page contains at least one inline link to %s with anchor text referencing the tool",
				req.toolPageURL,
			),
			"source":            req.emittedBy,
			"cross_link":        true,
			"tool_function":     req.toolFunction,
			"tool_display_name": req.toolName,
			// The URL machine-readably, so a sweep can check the link without
			// parsing prose (the prose copy above is what the writer obeys).
			"tool_page_url":    req.toolPageURL,
			"tool_page_id":     req.toolPageID.String(),
			"work_item_type":   "content_rewrite",
			"max_fix_attempts": 1,
		})

		// item_key keeps the tool_crosslink: namespace rather than the row's
		// own item_type (the usual workItemKey contract). Deliberate, and the
		// two reasons are worth stating: a cross-link is a distinct dedup unit
		// (tool × page) that must NOT collapse into the page-level
		// content_rewrite namespace, and live non-terminal rows already hold
		// their slot under this key — renaming it now would duplicate them.
		itemKey := fmt.Sprintf("tool_crosslink:%s:%s:%s", req.toolFunction, pageName, req.siteID)

		// Insert through the CENTRAL helper rather than a local INSERT. This is
		// the one place the ON CONFLICT clause and workItemTerminalStatuses are
		// kept in lockstep with idx_swi_dedup (the 42P10 class), and it brings
		// the two-strike anti-churn with it. The file used to carry its own
		// copy of that clause; three call sites of a copied dedup rule is
		// exactly the drift the index/Go-list contract warns about.
		//
		// recurrenceExpected stays FALSE (the default): a cross-link is a
		// detected gap, not a repeatable action request. A completed
		// predecessor means the reference is already woven, so suppressing the
		// re-request is correct, and a third attempt should surface as
		// 'unresolved' rather than churn the writer.
		pid := pageID
		inserted, err := withWorkItemTx(ctx, db, logger, workItem{
			siteID:       req.siteID,
			source:       req.emittedBy,
			pipeline:     "build",
			itemType:     "content_rewrite",
			severity:     "low",
			summary:      fmt.Sprintf("Add %s tool reference to %s page", req.toolName, pageName),
			spec:         string(spec),
			pageID:       &pid,
			priority:     110,
			handlerAgent: "page-build-handler",
			status:       "triaged",
			createdBy:    req.emittedBy,
			itemKey:      itemKey,
			dependsOn:    dependsOn,
		})
		if err != nil {
			logger.Warn("emitToolCrossLinkItems: failed to create cross-link item",
				zap.String("page", pageName), zap.Error(err))
			recordCrossLinkSkip(ctx, params, logger, req, "insert_failed", "warning",
				fmt.Sprintf("page %s: %v", pageName, err))
			continue
		}
		if !inserted {
			// Suppressed by dedup or anti-churn — an existing row already
			// covers this (tool, page). Not a loss, so not recorded.
			logger.Info("emitToolCrossLinkItems: cross-link already covered by an existing item",
				zap.String("page", pageName))
			continue
		}

		created++
		logger.Info("emitToolCrossLinkItems: created cross-link item",
			zap.String("page", pageName),
			zap.String("page_id", pageID.String()))
	}

	return created
}

// withWorkItemTx runs insertWorkItem in its own transaction. insertWorkItem
// takes a *sql.Tx because its callers batch many items; this emitter inserts
// one at a time from an action that has only a *sql.DB, and a failure here must
// never roll back the tool build that already succeeded.
func withWorkItemTx(ctx context.Context, db *sql.DB, logger *zap.Logger, item workItem) (bool, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	inserted, err := insertWorkItem(ctx, tx, item, logger)
	if err != nil {
		_ = tx.Rollback()
		return false, err
	}
	if err := tx.Commit(); err != nil {
		return false, err
	}
	return inserted, nil
}

// recordCrossLinkSkip makes a declined emit COUNTABLE. Every early return in
// emitToolCrossLinkItems means a page that was meant to reference a tool will
// not, and a log line is not a record: nothing can query it, nothing ages it,
// and the fleet has already been bitten by drops that only existed in logs
// (bugs_open/034). agent_error_log is the platform's existing durable channel
// for exactly this — same table recordComponentWriteRejection writes to.
//
// Deliberately NOT a work item: there is no handler that could action "the
// cross-link was withheld", and filing an item nobody can close is the
// bugs_open/077 shape (a queue entry whose handler has no remit).
func recordCrossLinkSkip(
	ctx context.Context,
	params ActionParams,
	logger *zap.Logger,
	req toolCrossLinkRequest,
	code string,
	severity string,
	detail string,
) {
	recordComponentWriteRejection(
		ctx, params.DB, logger, params,
		actionProvenance{
			AgentType: params.ExecutionContext.Sender.AgentType,
			StepName:  params.ExecutionContext.StepName,
			Action:    "emit_tool_cross_link_items",
		},
		fmt.Sprintf("tool cross-links not emitted for %s on site %s: %s",
			req.toolFunction, req.siteID, detail),
		"tool_crosslink_not_emitted:"+code,
		severity,
		map[string]interface{}{
			"site_id":         req.siteID.String(),
			"tool_function":   req.toolFunction,
			"tool_page_url":   req.toolPageURL,
			"tool_page_id":    req.toolPageID.String(),
			"related_pages":   req.relatedPages,
			"emitted_by":      req.emittedBy,
			"skip_reason":     code,
			"bug_reference":   "bugs_open/029",
			"related_pages_n": len(req.relatedPages),
		},
	)
}

// relatedPagesFromSpec reads a suggestion's related_pages, accepting the shapes
// it actually arrives in: the []interface{} of a decoded jsonb array, a
// []string, or a JSON-encoded string.
func relatedPagesFromSpec(raw interface{}) []string {
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case string:
		var out []string
		if json.Unmarshal([]byte(v), &out) == nil {
			return out
		}
	}
	return nil
}

// relatedPagesFromInputs reads the suggestion's related_pages for a tool BUILD
// action. Migration 211 wires "related_pages": "input_data.spec.related_pages"
// into both build steps' config (resolved by ExtractActionInputs Strategy 0);
// the direct read is the fallback for a config that predates it, so the fix
// works on an unmigrated agent definition too.
func relatedPagesFromInputs(inputs *datahelpers.ActionInputs, collected map[string]interface{}) []string {
	if pages := relatedPagesFromSpec(inputs.GetRaw("related_pages")); len(pages) > 0 {
		return pages
	}
	return relatedPagesFromSpec(datahelpers.ExtractNestedField(collected, "input_data.spec.related_pages"))
}

// resolveToolPageURL looks up the page a tool is deployed on. The join through
// page_components is the only reliable route: pages.name and pages.url both
// vary by build path, but the component's `function` is the naming contract.
func resolveToolPageURL(ctx context.Context, db *sql.DB, siteID uuid.UUID, toolFunction string) (uuid.UUID, string, bool) {
	var pageID uuid.UUID
	var url string
	err := db.QueryRowContext(ctx, `
		SELECT p.id, p.url
		FROM pages p
		JOIN page_components pc ON pc.page_id = p.id
		JOIN content_components cc ON cc.id = pc.component_id
		WHERE p.site_id = $1
		  AND p.status = 'active'
		  AND cc.function = $2
		  AND cc.component_level = 'tool'
		ORDER BY (p.build_status = 'deployed') DESC, p.created_at ASC
		LIMIT 1
	`, siteID, toolFunction).Scan(&pageID, &url)
	if err == nil && url != "" {
		return pageID, url, true
	}

	// Fallback: the page may exist before its component is linked. Both known
	// naming shapes, still read from pages.url — never constructed.
	err = db.QueryRowContext(ctx, `
		SELECT id, url FROM pages
		WHERE site_id = $1 AND status = 'active'
		  AND name IN ($2, $3)
		ORDER BY (build_status = 'deployed') DESC, created_at ASC
		LIMIT 1
	`, siteID, toolFunction, strings.TrimPrefix(toolFunction, "tool-")).Scan(&pageID, &url)
	if err == nil && url != "" {
		return pageID, url, true
	}
	return uuid.Nil, "", false
}

// ============================================================================
// ActionInputSpec
// ============================================================================

var CreateToolCrossLinkItemsInputSpec = datahelpers.ActionInputSpec{
	Required:   []string{"site_id"},
	Optional:   []string{"suggestions"},
	Defaults:   map[string]interface{}{},
	Deprecated: map[string]string{},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_tool_cross_link_items", CreateToolCrossLinkItemsInputSpec)
}

// ============================================================================
// ACTION: create_tool_cross_link_items
// ============================================================================

// CreateToolCrossLinkItemsAction is retained as a fail-safe, not as the
// emitter — see the file header. Called at suggestion time (its original
// wiring) it now finds no tool page and emits nothing; called after a build it
// emits correctly against the real URL.
func CreateToolCrossLinkItemsAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "create_tool_cross_link_items"),
	)

	logger.Info("CreateToolCrossLinkItemsAction: Starting")

	if params.ExecutionContext.Action == "initialize" {
		return map[string]interface{}{"status": "initialized"}, nil
	}
	if params.DB == nil {
		return nil, fmt.Errorf("database connection required")
	}

	// --- Resolve inputs ---
	inputs, err := datahelpers.ExtractActionInputs(
		params.CollectedData,
		params.StepConfig.Config,
		CreateToolCrossLinkItemsInputSpec,
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

	// --- Extract suggestions from collected data ---
	// The suggestions come from evaluation.result.suggestions (the same array
	// that create_items_loop iterates over)
	suggestionsRaw := datahelpers.ExtractNestedField(params.CollectedData, "evaluation.result.suggestions")
	if suggestionsRaw == nil {
		logger.Info("CreateToolCrossLinkItemsAction: No suggestions found, skipping")
		return map[string]interface{}{"items_created": 0, "reason": "no suggestions"}, nil
	}

	suggestions, ok := suggestionsRaw.([]interface{})
	if !ok {
		logger.Info("CreateToolCrossLinkItemsAction: Suggestions not an array, skipping",
			zap.String("type", fmt.Sprintf("%T", suggestionsRaw)))
		return map[string]interface{}{"items_created": 0, "reason": "suggestions not array"}, nil
	}

	itemsCreated := 0
	skippedNoPage := 0
	for _, sugRaw := range suggestions {
		sug, ok := sugRaw.(map[string]interface{})
		if !ok {
			continue
		}

		toolName, _ := sug["name"].(string)
		toolFunction, _ := sug["function"].(string)
		toolDescription, _ := sug["description"].(string)

		if toolName == "" || toolFunction == "" {
			continue
		}

		relatedPages := relatedPagesFromSpec(sug["related_pages"])
		if len(relatedPages) == 0 {
			continue
		}

		// The whole of bugs_open/029 in one branch: no page row, no URL, no
		// item. The tool build paths emit these once the page exists.
		toolPageID, toolPageURL, found := resolveToolPageURL(ctx, params.DB, siteID, toolFunction)
		if !found {
			skippedNoPage++
			logger.Info("CreateToolCrossLinkItemsAction: tool has no page yet — the build path will emit its cross-links",
				zap.String("tool", toolFunction))
			continue
		}

		itemsCreated += emitToolCrossLinkItems(ctx, params, logger, toolCrossLinkRequest{
			siteID:       siteID,
			toolFunction: toolFunction,
			toolName:     toolName,
			toolDesc:     toolDescription,
			toolPageID:   toolPageID,
			toolPageURL:  toolPageURL,
			relatedPages: relatedPages,
			emittedBy:    "tool-suggester",
		})
	}

	logger.Info("CreateToolCrossLinkItemsAction: Complete",
		zap.Int("items_created", itemsCreated),
		zap.Int("skipped_tool_has_no_page", skippedNoPage),
		zap.Int("suggestions_processed", len(suggestions)))

	return map[string]interface{}{
		"items_created":            itemsCreated,
		"skipped_tool_has_no_page": skippedNoPage,
	}, nil
}
