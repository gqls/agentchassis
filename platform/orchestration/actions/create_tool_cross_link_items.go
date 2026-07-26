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
// Returns the number of items created. Errors are logged and swallowed per
// item: cross-linking is a follow-on nicety and must never fail a tool build.
func emitToolCrossLinkItems(ctx context.Context, db *sql.DB, logger *zap.Logger, req toolCrossLinkRequest) int {
	logger = logger.With(
		zap.String("tool", req.toolFunction),
		zap.String("tool_page_url", req.toolPageURL),
	)

	if req.toolFunction == "" || req.toolName == "" {
		return 0
	}
	if len(req.relatedPages) == 0 {
		logger.Info("emitToolCrossLinkItems: no related_pages on the suggestion, nothing to cross-link")
		return 0
	}
	if !strings.HasPrefix(req.toolPageURL, "/") {
		// Never guess. bugs_open/029 is exactly what guessing produced.
		logger.Warn("emitToolCrossLinkItems: refusing to emit without a real tool page URL",
			zap.String("received", req.toolPageURL))
		return 0
	}

	// --- Guard 2: prove the tool page will be served ---
	var buildStatus string
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(build_status, '') FROM pages WHERE id = $1`, req.toolPageID,
	).Scan(&buildStatus); err != nil {
		logger.Warn("emitToolCrossLinkItems: tool page row not readable, skipping", zap.Error(err))
		return 0
	}

	var dependsOn *string
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
			return 0
		}
		gate := "{" + gateID.String() + "}"
		dependsOn = &gate
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

		// The ON CONFLICT WHERE clause must imply idx_swi_dedup's predicate
		// (see workItemTerminalStatuses) — a hardcoded stale list here
		// previously made this insert fail 42P10 and only Warn-log.
		_, err := db.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by,
				item_key, depends_on
			) VALUES (
				$1, $2, 'build', 'content_rewrite', 'low',
				$3, $4::jsonb, $5, 110, 'page-build-handler', 'triaged', $2,
				$6, $7::uuid[]
			) ON CONFLICT (site_id, item_key)
			  WHERE item_key IS NOT NULL
			  AND status NOT IN (%s)
			  DO NOTHING
		`, sqlInList(workItemTerminalStatuses)),
			req.siteID, req.emittedBy,
			fmt.Sprintf("Add %s tool reference to %s page", req.toolName, pageName),
			string(spec), pageID, itemKey, dependsOn,
		)
		if err != nil {
			logger.Warn("emitToolCrossLinkItems: failed to create cross-link item",
				zap.String("page", pageName), zap.Error(err))
			continue
		}

		created++
		logger.Info("emitToolCrossLinkItems: created cross-link item",
			zap.String("page", pageName),
			zap.String("page_id", pageID.String()))
	}

	return created
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

		itemsCreated += emitToolCrossLinkItems(ctx, params.DB, logger, toolCrossLinkRequest{
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
