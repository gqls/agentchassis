// FILE: platform/orchestration/actions/create_tool_component_action.go
//
// Creates a new tool component from LLM-generated HTML, creates a tool page,
// and links them via page_components. Used by tool-generator for tools
// that don't exist in the library.
//
// This is the "create from scratch" counterpart to deploy_tool_to_site
// (which forks existing library tools).
//
// After creation, this action also:
//   - Sets rendered_html on the page_component (so the tool is visible immediately)
//   - Creates a needs_content_page work item (so hero/intro/CTA sections get written)
//   - Creates a companion guide page + work item (same pattern as deploy_tool_action)
//
// Registration:
//
//	"create_tool_component": {
//	    Handler:     CreateToolComponentAction,
//	    Category:    "site",
//	    Description: "Create a new tool component from generated HTML and set up its page",
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
	"github.com/gqls/agentchassis/platform/content"
	"github.com/gqls/agentchassis/platform/orchestration/datahelpers"
	"go.uber.org/zap"
)

// ============================================================================
// ActionInputSpec
// ============================================================================

var CreateToolComponentInputSpec = datahelpers.ActionInputSpec{
	Required: []string{"site_id", "html_content", "function", "display_name"},
	// related_pages: the add_tool item's own suggestion field. Read here so the
	// tool's cross-link items can be emitted from THIS path, which knows the
	// page's real URL, instead of at suggestion time where it can only be
	// guessed (bugs_open/029).
	// replace_existing (bugs_open/331, TL-047): per-ITEM authority to regenerate
	// the site's existing tool component for this function IN PLACE instead of
	// taking the already_exists short-circuit. Mapped STRICTLY by the generator
	// seed ("replace_existing!": "input_data.spec.replace_existing") so the
	// whole-tree search (RFC_029) can never supply it; absent ⇒ today's path.
	Optional:   []string{"description", "category", "related_pages", "replace_existing"},
	Defaults:   map[string]interface{}{"category": "interactive"},
	Deprecated: map[string]string{},
	// enforce_instance_scope: arms the instance-scope birth guard
	// (tool_birth_instance_scope.go). DECLARED HERE IN THE SAME COMMIT AS ITS
	// READER — bugs_closed/336 is what happens otherwise: a key declared on the
	// wrong action's spec passes every instrument (binary literal present,
	// git log -S names the reader) while validation refuses it at dispatch.
	// Setting: not an input field, so ConfigKeys per the RenderComponent
	// precedent; it does not move the RFC_022 optional-key budget.
	ConfigKeys: []string{"enforce_instance_scope"},
}

func init() {
	datahelpers.RegisterActionInputSpec("create_tool_component", CreateToolComponentInputSpec)
}

// ============================================================================
// ACTION: create_tool_component
// ============================================================================

func CreateToolComponentAction(ctx context.Context, params ActionParams) (interface{}, error) {
	logger := params.Logger.With(
		zap.String("action", "create_tool_component"),
	)

	logger.Info("CreateToolComponentAction: Starting")

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
		CreateToolComponentInputSpec,
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

	htmlContent := inputs.Get("html_content")
	if htmlContent == "" {
		return nil, fmt.Errorf("html_content is empty — LLM generation may have failed")
	}

	// Strip markdown fences if LLM included them despite instructions
	htmlContent = datahelpers.StripCodeFences(htmlContent)

	// Tool-doc header gate (019 §Tool Doc Header). All tool-generator output
	// passes through here, so this is the contract's enforcement point (the
	// generator workflow has no check_tool_completeness step — that action
	// rides the tool-recreation flow only). Hard fail: the workflow's
	// error_step fires and the work item carries the reason.
	// DEPLOYMENT ORDER: apply the tool-generator prompt update (which teaches
	// the header) before or with the binary carrying this gate, or every
	// generation fails here.
	if !content.HasToolDocHeader(htmlContent) {
		return nil, fmt.Errorf("generated tool HTML lacks the tool-doc header (%s ... %s) — prompt not yet updated, or the LLM dropped it; regenerate (019 §Tool Doc Header)",
			content.ToolDocOpen, content.ToolDocClose)
	}

	// Completeness gate (bugs_open/021 INSTANCE 1 / bugs_open/046). HasToolDocHeader
	// above only proves the doc-header SENTINELS are present, and they sit at the
	// TOP of the <script> — so a generation cut off mid-stream at the TAIL keeps its
	// header and sails through. That is exactly how bugs_open/046's 8 tools were
	// born truncated and shipped broken JavaScript to 6 live customer domains.
	//
	// componentTemplateValid is the SAME absolute structural predicate the schema
	// loader applies (plan_sections_action.go: every paired tag balanced in markup
	// context and the template ends on a closed tag). Applying it at the birth
	// WRITE means a tool that would be dropped at LOAD can never be born in the
	// first place — one contract, both seams. Hard fail: the workflow's error_step
	// routes the item to needs_human_review carrying this reason (same path as the
	// header gate above).
	//
	// The refusal names the MEASURED signals — which pair is unbalanced, at what
	// counts, and whether the template ends cleanly — and does not assert a cause.
	// The previous message claimed "the generation was cut" unconditionally, which
	// its own context payload (ends_cleanly: true) disproved on the bugs_open/303
	// false positives; a wrong asserted cause sent the next reader hunting a
	// token-limit truncation that had not happened. If this fires with
	// ends_cleanly true AND balanced counts, something new is wrong — read
	// toolTemplateValid before trusting any message here.
	if !componentTemplateValid(htmlContent, "tool") {
		var imbalance []string
		for _, tb := range content.StructuralTagCounts(htmlContent) {
			if tb.Opens > tb.Closes {
				imbalance = append(imbalance, fmt.Sprintf("%s: %d open vs %d close", tb.Open, tb.Opens, tb.Closes))
			}
		}
		signals := "ends mid-token (last chars: " + tailForMessage(htmlContent) + ")"
		if len(imbalance) > 0 {
			signals = "unbalanced in markup context: " + strings.Join(imbalance, ", ")
		}
		recordComponentWriteRejection(
			ctx, logger, params,
			actionProvenance{
				AgentType: params.ExecutionContext.Sender.AgentType,
				StepName:  params.ExecutionContext.StepName,
				Action:    "create_tool_component",
			},
			fmt.Sprintf("tool birth refused for site %s: generated HTML is structurally incomplete — %s", siteIDStr, signals),
			"tool_birth_truncation_blocked",
			"error",
			map[string]interface{}{
				"site_id":       siteIDStr,
				"html_length":   len(htmlContent),
				"ends_cleanly":  endsCleanly(htmlContent),
				"tag_imbalance": imbalance,
			},
		)
		return nil, fmt.Errorf("refusing to persist structurally incomplete tool for site %s: %s — if a generation was genuinely cut, llm_call_log's output_tokens will be at or near max_tokens; check that before re-rolling (bugs_open/303)",
			siteIDStr, signals)
	}

	function := inputs.Get("function")
	displayName := inputs.Get("display_name")
	description := inputs.Get("description")
	category := inputs.Get("category")
	if category == "" {
		category = "interactive"
	}

	// Read config for nav/page settings
	config := params.StepConfig.Config
	navSection := "Tools"
	if v, ok := config["nav_section"].(string); ok && v != "" {
		navSection = v
	}
	inHeader := true
	if v, ok := config["in_header"].(bool); ok {
		inHeader = v
	}
	inFooter := false
	if v, ok := config["in_footer"].(bool); ok {
		inFooter = v
	}
	// adopt_existing_page (opt-in, default OFF — the 2026-08-02 §2 shape): when
	// true, a page already holding this tool's canonical or legacy name is
	// attached to instead of collided with — resolveToolPageIdentity +
	// UpsertPageForRole, the same machinery deploy_tool uses. When false (every
	// existing caller), behaviour is unchanged: a bare INSERT that errors on
	// pages_site_id_name_key. The flag exists for the ported-tool replacement
	// route (bugs_open/286): a same-URL rebuild is impossible without it.
	adoptExistingPage := false
	if v, ok := config["adopt_existing_page"].(bool); ok {
		adoptExistingPage = v
	}

	// --- Load site domain ---
	var siteDomain string
	err = params.DB.QueryRowContext(ctx,
		`SELECT domain FROM sites WHERE id = $1`, siteID,
	).Scan(&siteDomain)
	if err != nil {
		return nil, fmt.Errorf("failed to load site domain: %w", err)
	}

	// Sanitise function name
	function = sanitiseFunction(function)

	// Build component name (unique per site)
	domainSlug := strings.ReplaceAll(siteDomain, ".", "-")
	componentName := fmt.Sprintf("%s-%s", function, domainSlug)

	logger.Info("CreateToolComponentAction: Creating tool",
		zap.String("site_id", siteIDStr),
		zap.String("domain", siteDomain),
		zap.String("function", function),
		zap.String("display_name", displayName),
		zap.String("component_name", componentName),
	)

	// --- Instance-scope birth guard (bugs_open/283 flow half) ---
	// Placed AFTER sanitiseFunction (the token embeds the sanitised name) and
	// BEFORE the already-exists branch, so the replace_existing regeneration
	// receives scoped bytes too. renderedHTML is the occurrence-0 binding for
	// the verbatim rendered_html writes below.
	scopedHTML, renderedHTML, scopeInfo, scopeRefuse := ScopeToolBirthTemplate(
		htmlContent, function, enforceInstanceScope(config), logger)
	if scopeRefuse != nil {
		recordComponentWriteRejection(
			ctx, logger, params,
			actionProvenance{
				AgentType: params.ExecutionContext.Sender.AgentType,
				StepName:  params.ExecutionContext.StepName,
				Action:    "create_tool_component",
			},
			fmt.Sprintf("tool birth refused for site %s (%s): %v", siteIDStr, function, scopeRefuse),
			"tool_birth_instance_scope_refused",
			"error",
			map[string]interface{}{
				"site_id":  siteIDStr,
				"function": function,
				"detail":   fmt.Sprintf("%v", scopeInfo["instance_scope_defect"]),
			},
		)
		return nil, scopeRefuse
	}
	htmlContent = scopedHTML
	logger.Info("CreateToolComponentAction: instance-scope birth guard",
		zap.Any("scope", scopeInfo))

	// --- Check if already exists ---
	var existingID string
	err = params.DB.QueryRowContext(ctx, `
		SELECT cc.id::text FROM content_components cc
		JOIN page_components pc ON pc.component_id = cc.id
		JOIN pages p ON pc.page_id = p.id
		WHERE cc.function = $1
		  AND cc.component_level = 'tool'
		  AND p.site_id = $2
		  AND cc.is_active = true
		LIMIT 1
	`, function, siteID).Scan(&existingID)

	if err == nil && existingID != "" {
		// REPLACE arm (bugs_open/331): the item asked for a regeneration of the
		// tool it knows the site already has — regenerate the incumbent in place
		// (create_tool_component_regenerate.go). Without the flag the
		// short-circuit below stands: it is the per-site throttle.
		if replaceExistingRequested(inputs) {
			return regenerateToolComponentInPlace(ctx, params, logger, toolRegenerateRequest{
				siteID:       siteID,
				incumbentID:  existingID,
				function:     function,
				displayName:  displayName,
				description:  description,
				category:     category,
				htmlContent:  htmlContent,
				renderedHTML: renderedHTML,
			})
		}
		logger.Info("CreateToolComponentAction: Tool already exists for this site",
			zap.String("existing_id", existingID))
		return map[string]interface{}{
			"already_exists": true,
			"component_id":   existingID,
			"function":       function,
		}, nil
	}

	// --- Library claim check (RFC_036 §9.3) ---
	// idx_cc_tool_function_unique is partial on `component_level='tool' AND
	// forked_from IS NULL AND is_active`, so if a LIBRARY tool already claims
	// this function, the bare INSERT below dies on SQLSTATE 23505 at save_tool
	// — while the work item reports complete (RFC_036 §1). A site-specific
	// native build of a tool the library also offers IS a site copy of that
	// tool — which is what forked_from means everywhere else, and the shape
	// deploy_tool_to_site already creates (deploy_tool_action.go) — so the new
	// row forks from the library entry and the index keeps guaranteeing what
	// it exists for: one canonical library entry per tool function. The
	// section-level sibling of this collision is bugs_open/311 (CLC-020),
	// resolved differently on measured grounds recorded in RFC_036 §11.
	// A lookup read error degrades to today's behaviour (no fork → the index
	// refuses LOUDLY on a real collision), never to a silent guess.
	var libraryToolFork interface{} // nil unless a library entry claims the function
	{
		var libID string
		lookupErr := params.DB.QueryRowContext(ctx, `
			SELECT id::text FROM content_components
			WHERE function = $1
			  AND component_level = 'tool'
			  AND forked_from IS NULL
			  AND is_active = true
			LIMIT 1
		`, function).Scan(&libID)
		switch {
		case lookupErr == nil && libID != "":
			libraryToolFork = libID
			logger.Info("CreateToolComponentAction: library tool claims this function — new component forks from it (RFC_036 §9.3)",
				zap.String("function", function),
				zap.String("library_tool_id", libID))
		case lookupErr != nil && lookupErr != sql.ErrNoRows:
			logger.Warn("CreateToolComponentAction: library-claim lookup failed — proceeding without fork; a real collision will be refused loudly by idx_cc_tool_function_unique",
				zap.String("function", function),
				zap.Error(lookupErr))
		}
	}

	// --- Create component ---
	componentID := uuid.New()

	// source_* = creation provenance (NNN_add_component_provenance.sql),
	// mirroring knowledge_base's pair. Values from the execution context the
	// action already holds — apply that migration before this binary deploys.
	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO content_components (
			id, name, display_name, function, component_level,
			category, description, html_template,
			is_active, is_dark_section, created_from,
			source_agent_type, source_orchestration_id, forked_from
		) VALUES ($1, $2, $3, $4, 'tool', $5, $6, $7, true, false, 'generated', $8, $9, $10::uuid)
	`, componentID, componentName, displayName, function,
		category, description, htmlContent,
		nullIfEmpty(params.ExecutionContext.Sender.AgentType),
		nullIfEmpty(params.ExecutionContext.OrchestrationID),
		libraryToolFork)
	if err != nil {
		return nil, fmt.Errorf("failed to create tool component: %w", err)
	}

	logger.Info("CreateToolComponentAction: Component created",
		zap.String("component_id", componentID.String()))

	// --- Create page ---
	// Canonical identity via CanonicalisePage (guard rail 2 / TL-003): the old
	// flat /tools/<function>.html ignored the site's canonical URL shape and
	// every birth needed a manual reconcile. sanitiseFunction guarantees the
	// tool- prefix, so the canonical name equals function and the acceptance
	// coupling (pages.name == content_components.function) holds.
	pageID := uuid.New()
	pageName, pageURL, _ := datahelpers.CanonicalisePage(datahelpers.PageDescriptor{
		Role:     "tool",
		Slug:     function,
		FlatURLs: siteUsesFlatURLs(ctx, params.DB, siteID, logger),
	})
	if pageName == "" {
		// bugs_open/080: this used to fall back to the hand-rolled flat
		// /tools/<function>.html — the divergent-identity shape the bug is
		// about, kept alive as an else-branch. An identity that cannot be
		// canonicalised is refused, not hand-rolled; sanitiseFunction has
		// already run, so reaching this means the function was empty or
		// reduced to nothing.
		params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
		return nil, fmt.Errorf("tool function %q failed canonicalisation", function)
	}
	pageTitle := fmt.Sprintf("%s | %s", displayName, navSection)

	// bugs_open/103, SECOND CALL SITE — and bugs_open/339, where 103's remedy was
	// found blind. This action's `description` is the tool's BUILD BRIEF by
	// definition (datahelpers/meta_description.go's own header says exactly that
	// of a component_level='tool' row), so it is NEVER a candidate for public
	// copy: 339 measured the current brief population at 200-320 chars, under the
	// guard's 320 threshold and matching none of its July-census briefMarkers, so
	// nine live tool pages published their spec verbatim. A guard re-fitted to
	// today's brief style would go stale the same way; not offering the brief at
	// all is the version that cannot. The empty candidate makes
	// PublicMetaDescription vet only the COMPOSED side (a brief-shaped composed
	// line is a caller bug and must yield "" rather than publish), and `replaced`
	// is definitionally false, so the old rejection log has nothing to report.
	// If a future path wants real marketing copy on a tool page, it needs its own
	// vetted field; this column's candidate slot is closed on purpose.
	toolMeta, _ := datahelpers.PublicMetaDescription(
		"",
		composedToolMetaDescription(displayName),
	)

	// in_header / in_footer are written EXPLICITLY (bugs_open/149 A4's recordable
	// half). They used to be omitted, so both inherited the column default — which
	// is TRUE for each — while this action had already computed inHeader/inFooter
	// from its own step config and used them only to decide whether to touch nav.
	// A step configured `in_footer: false` therefore produced a row saying
	// `in_footer = true`: the writer discarded its own decision, and the flag
	// recorded nothing.
	//
	// This is a prerequisite for the classifyPagesForNav fix in the same commit,
	// not a tidy-up: that fix makes these flags load-bearing — they are what puts
	// a /tools/ page into the footer nav — so an inherited default would now
	// silently place pages the config asked to keep out.
	pageAdopted := false
	if adoptExistingPage {
		// Replacement arm (bugs_open/286). An existing row under either name
		// shape keeps its stored identity (bugs_open/080's rule — writing the
		// canonical name past a legacy row would mint a second page); only no
		// row at all mints the canonical shape.
		pageName, pageURL, err = resolveToolPageIdentity(ctx, params.DB, siteID, function,
			siteUsesFlatURLs(ctx, params.DB, siteID, logger))
		if err != nil {
			params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
			return nil, fmt.Errorf("failed to resolve tool page identity: %w", err)
		}
		toolPage, uerr := UpsertPageForRole(ctx, params.DB, PageRoleUpsert{
			SiteID:   siteID,
			Domain:   siteDomain,
			Name:     pageName,
			PageType: "tool",
			Source:   "tool-generator",
			// Declared: the role is this arm's constant, so an unshipped page of
			// another role holding this name is one this arm may take over (RFC_010).
			AdoptUnshippedRows: true,
			Columns: []PageColumn{
				Col("url", pageURL),
				Col("title", pageTitle),
				Col("nav_order", 200),
				Col("in_header", inHeader),
				Col("in_footer", inFooter),
				Col("meta_description", toolMeta),
				Col("build_status", "planned"),
				Col("status", "active"),
			},
			// Empty on purpose: a live same-role page is attached to as it
			// stands — the replacement route retires the ported slot and
			// re-renders separately, and nothing about the served page may
			// change as a side effect of the component write.
			Refresh: []string{},
		}, logger)
		if uerr != nil {
			params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
			return nil, fmt.Errorf("failed to create tool page: %w", uerr)
		}
		if toolPage.Refused() {
			params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
			return nil, fmt.Errorf("tool page %q not created: %s", pageName, toolPage.Reason)
		}
		pageID = toolPage.PageID
		pageAdopted = toolPage.Outcome != PageRoleCreated
	} else {
		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO pages (
				id, site_id, name, url, title,
				page_type, status, build_status, nav_order, meta_description,
				in_header, in_footer
			) VALUES ($1, $2, $3, $4, $5, 'tool', 'active', 'planned', 200, $6, $7, $8)
		`, pageID, siteID, pageName, pageURL, pageTitle, toolMeta, inHeader, inFooter)
		if err != nil {
			// If page creation fails, clean up the component
			params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
			return nil, fmt.Errorf("failed to create tool page: %w", err)
		}
	}

	logger.Info("CreateToolComponentAction: Page created",
		zap.String("page_id", pageID.String()),
		zap.String("url", pageURL),
		zap.Bool("page_adopted", pageAdopted))

	// --- Link component to page ---
	// Position 2 (same as deploy_tool_action): tool widget sits between intro and CTA
	// Set rendered_html so the tool is visible immediately
	// Set slot_name to the function (component naming contract)
	//
	// bugs_open/362: this writer was one of the two that persisted rendered_html
	// with NO dead-link repair — deferred on 2026-08-02 because the repair then
	// deleted JS-built anchors from tool scripts (bugs_open/180), unblocked the
	// same day by LNK-029's span-aware repair, and unwired for twenty days
	// because the documents that said "wait" outlived the wait. Placed after the
	// pages row exists (this tool's own URL is in the index) and before the
	// write, so what is repaired is what is persisted. Fail-open by the seam's
	// own design; the JS-built-anchor probe lives in
	// tool_writer_link_repair_362_test.go.
	renderedHTML = repairComponentHTMLBeforePersist(ctx, params, siteID,
		siteDomain, pageName, pageURL, "create_tool_component", renderedHTML, logger)
	_, err = params.DB.ExecContext(ctx, `
		INSERT INTO page_components (
			page_id, component_id, position, slot_name,
			rendered_html, content_data, build_status
		) VALUES ($1, $2, 2, $3, $4, '{}'::jsonb, 'deployed')
	`, pageID, componentID, function, renderedHTML)
	if err != nil {
		// Clean up on failure. A page this call ADOPTED pre-existed the call
		// and is live — deleting it here would destroy a served page to tidy
		// up a failed link; only a row this call created may be removed.
		if !pageAdopted {
			params.DB.ExecContext(ctx, `DELETE FROM pages WHERE id = $1`, pageID)
		}
		params.DB.ExecContext(ctx, `DELETE FROM content_components WHERE id = $1`, componentID)
		return nil, fmt.Errorf("failed to link component to page: %w", err)
	}

	// --- Ask for the nav rebuild the declaration needs (bugs_open/149 A6) ---
	//
	// This replaced addToolToNav on 2026-07-31. That function hand-wrote a
	// site_nav_items row into a bespoke `tools` group typed `primary`, i.e. the
	// header — which the platform's own classifier bars for tool pages ("tier 4,
	// never primary") and which the next populate_nav_tables run DELETED without
	// replacing. Two writers of a derived table, one of which cannot express what
	// the other writes. See nav_rebuild_request.go for the full argument, including
	// why writing the row here would silence check_orphan_pages.
	if inHeader || inFooter {
		RequestNavRebuild(ctx, params.DB, NavRebuildRequest{
			SiteID:      siteID,
			PageID:      pageID,
			PageName:    pageName,
			PageURL:     pageURL,
			InHeader:    inHeader,
			InFooter:    inFooter,
			RequestedBy: "tool-generator",
		}, logger)
	}

	// --- Ask for content around the tool, if the page can take any ---
	//
	// bugs_open/177. This arm declares no sections above (see the pages INSERT),
	// so the item it used to raise unconditionally could never be satisfied: all
	// nine tool_content items ever minted died in needs_human_review, and one of
	// them blocked a dependent long enough to stall the fleet. The guard asks the
	// same sources page-build-handler will ask and declines when the answer is
	// "nothing a writer could write". See tool_content_item.go.
	contentItem := raiseToolContentItem(ctx, params, logger, toolContentItemRequest{
		siteID:       siteID,
		pageID:       pageID,
		pageName:     pageName,
		toolFunction: function,
		displayName:  displayName,
		description:  description,
		pageURL:      pageURL,
		source:       "tool-generator",
	})

	// --- Cross-link the tool from its related pages (bugs_open/029) ---
	// Emitted HERE, after the page row and its content build item exist, so the
	// rewrite instruction carries the page's real URL and only runs once the
	// page is live. This path's URL shape (CanonicalisePage → /tools/x/index.html)
	// is one of the three that the old suggestion-time constructor got wrong.
	crossLinksAdded := emitToolCrossLinkItems(ctx, params, logger, toolCrossLinkRequest{
		siteID:       siteID,
		toolFunction: function,
		toolName:     displayName,
		toolDesc:     description,
		toolPageID:   pageID,
		toolPageURL:  pageURL,
		relatedPages: relatedPagesFromInputs(inputs, params.CollectedData),
		emittedBy:    "tool-generator",
	})

	// --- Create companion guide page ---
	// Identity via the shared canonicaliser (bugs_open/080) — byte-identical to
	// the old sprintf for every existing guide, and the single authority now.
	guideName, guideURL, guideIDErr := companionGuideIdentity(pageName)
	if guideIDErr != nil {
		logger.Warn("CreateToolComponentAction: companion guide identity failed canonicalisation, guide skipped (non-fatal)",
			zap.Error(guideIDErr))
	}
	guideTitle := fmt.Sprintf("Understanding %s | Guide", displayName)

	// bugs_open/175. Byte-identical to deploy_tool_action's companion-guide upsert
	// (two files, one statement) and carrying the same defect: `page_type` was in
	// the INSERT and not in the DO UPDATE SET, so a collision re-titled somebody
	// else's page and left it typed as it was. Both now go through the one seam —
	// page_role_upsert.go — which is also what stops the two copies drifting apart
	// again.
	//
	// Refresh is `title` ALONE, deliberately: a guide already written must not lose
	// its sections to a tool regeneration. That is the existing behaviour, kept.
	var guidePage PageRoleResult
	guideErr := guideIDErr
	if guideErr == nil {
		guidePage, guideErr = UpsertPageForRole(ctx, params.DB, PageRoleUpsert{
			SiteID:             siteID,
			Domain:             siteDomain,
			Name:               guideName,
			PageType:           "blog-post",
			Source:             "tool-generator",
			AdoptUnshippedRows: true,
			Columns: []PageColumn{
				Col("url", guideURL),
				Col("title", guideTitle),
				Col("nav_order", 200),
				Col("in_header", false),
				Col("in_footer", false),
				Col("meta_description", composedGuideMetaDescription(displayName)),
				JSONCol("sections", `["hero", "article-body", "call-to-action"]`),
				Col("build_status", "planned"),
				Col("status", "active"),
			},
			Refresh: []string{"title"},
		}, logger)
	}

	switch {
	case guideErr != nil:
		logger.Warn("CreateToolComponentAction: Failed to create companion guide page (non-fatal)", zap.Error(guideErr))
	case guidePage.Refused():
		logger.Warn("CreateToolComponentAction: companion guide refused — a live page holds that name under another role",
			zap.String("guide_name", guideName),
			zap.String("existing_page_type", guidePage.ExistingType),
			zap.String("reason", guidePage.Reason))
	default:
		guidePageID := guidePage.PageID
		// Create work item for the guide article
		guideSpec, _ := json.Marshal(map[string]interface{}{
			"page_name":         guideName,
			"page_id":           guidePageID.String(),
			"page_type":         "blog-post",
			"tool_function":     function,
			"tool_display_name": displayName,
			"tool_page_url":     pageURL,
			"source":            "tool-generator",
			// bugs_open/271: "suggestion" is the key page-build-handler reads
			// and forwards to the writer prompt; "content_guidance" had none.
			"suggestion": fmt.Sprintf(
				"Write an in-depth guide about %s. Explain the concept, why it matters, common mistakes people make, "+
					"and practical tips. At relevant points, reference the interactive %s tool at %s — "+
					"e.g. 'Use our %s to see how this applies to your situation.' "+
					"The article should stand alone as useful content even without the tool.",
				strings.TrimPrefix(displayName, "UK "),
				strings.ToLower(displayName), pageURL,
				strings.ToLower(displayName),
			),
		})

		_, err = params.DB.ExecContext(ctx, `
			INSERT INTO site_work_items (
				site_id, source, pipeline, item_type, severity, summary,
				spec, page_id, priority, handler_agent, status, created_by, item_key
			) VALUES (
				$1, 'tool-generator', 'build', 'needs_content_page', 'low',
				$2, $3::jsonb, $4, 70, 'page-build-handler', 'triaged', 'tool-generator', $5
			) ON CONFLICT DO NOTHING
		`, siteID,
			fmt.Sprintf("Write companion guide: %s", guideTitle),
			string(guideSpec), guidePageID,
			fmt.Sprintf("tool_guide:%s:%s", function, siteID),
		)
		if err != nil {
			logger.Warn("CreateToolComponentAction: Failed to create guide work item (non-fatal)", zap.Error(err))
		} else {
			logger.Info("CreateToolComponentAction: Companion guide created",
				zap.String("guide_page_id", guidePageID.String()),
				zap.String("guide_url", guideURL))
		}
	}

	logger.Info("CreateToolComponentAction: Complete",
		zap.String("component_id", componentID.String()),
		zap.String("page_id", pageID.String()),
		zap.String("page_url", pageURL),
		zap.String("function", function))

	result := map[string]interface{}{
		"component_id":      componentID.String(),
		"page_id":           pageID.String(),
		"page_url":          pageURL,
		"function":          function,
		"display_name":      displayName,
		"guide_url":         guideURL,
		"needs_rerender":    true,
		"generated":         true,
		"page_adopted":      pageAdopted,
		"cross_links_added": crossLinksAdded,
		"content_item":      contentItem,
	}
	for k, v := range scopeInfo {
		result[k] = v
	}
	return result, nil
}

// addToolToNav was DELETED on 2026-07-31 (bugs_open/149 A6). It hand-wrote a
// site_nav_items row into a `tools` group typed `primary` — the header — for a
// page type the platform's own classifier bars from primary nav, and the next
// populate_nav_tables run deleted the row without replacing it. Its replacement
// is RequestNavRebuild (nav_rebuild_request.go): declare membership on the page
// row, let the one derivation write the nav row, and re-render the chrome that
// makes the link real. Do not reintroduce a second writer of site_nav_items —
// that is the defect, not the fix.

// sanitiseFunction ensures function name is valid kebab-case with tool- prefix
func sanitiseFunction(function string) string {
	function = strings.ToLower(strings.TrimSpace(function))
	function = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			return r
		}
		if r == ' ' || r == '_' {
			return '-'
		}
		return -1
	}, function)
	if !strings.HasPrefix(function, "tool-") {
		function = "tool-" + function
	}
	return function
}
