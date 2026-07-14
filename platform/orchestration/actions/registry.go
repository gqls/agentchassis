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
	"derive_brand_head_assets": {
		Handler:     DeriveBrandHeadAssetsAction,
		Category:    "image",
		Description: "Derive favicon + OG card from the site logo and commit to git",
		IsLocal:     true,
	},
	"emit_sprite_css": {
		Handler:     EmitSpriteCSSAction,
		Category:    "site",
		Description: "Generate and commit sprites.css from the verified sprite-sheet plan",
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
	"load_existing_component": {
		Handler:     LoadExistingComponentAction,
		Category:    "site",
		Description: "Load existing component field names for regeneration preservation",
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
	"rerender_page_sections": {
		Handler:     RerenderPageSectionsAction,
		Category:    "site",
		Description: "Re-render a page's sections from stored content_data + fresh resolved fields (no LLM)",
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
	"request_browser_run": {
		Handler:     RequestBrowserRunAction,
		Category:    "tools",
		Description: "Ask the browser-runner adapter to drive a tool's deployed page against its PLAN criteria; awaits the results (no-op skip when the PLAN has no criteria)",
		IsLocal:     true,
	},
	"judge_acceptance_results": {
		Handler:     JudgeAcceptanceResultsAction,
		Category:    "tools",
		Description: "Turn a browser run into acceptance-run/acceptance-fail doc_notes and an improve_tool item carrying the criteria as acceptance_test",
		IsLocal:     true,
	},
	"analyse_repo_local": {
		Handler:     AnalyseRepoLocalAction,
		Category:    "code",
		Description: "Fetch a repo at ref to a local temp dir (read-only tarball) and analyse it in-process; returns the analyser Output with a real local root + commit_sha (for the diagnose loop's body reads). Read-only.",
		IsLocal:     true,
	},

	// =========================================================================
	// DIAGNOSE — the read-only, workflow-driven diagnosis loop (engine: pkg/diagnose)
	//   gather: analyse_repo_local → lookup_code_symbols → diagnose_load_runtime
	//           → diagnose_assemble_bundle
	//   verdict: execute_llm_prompt (verdict prompt; NOT a diagnose action)
	//   control: diagnose_route (guards + re-scope; loops back to assemble | → emit)
	//   output:  diagnose_emit (formats the human-facing diagnosis; read-only)
	// =========================================================================
	"diagnose_load_runtime": {
		Handler:     DiagnoseLoadRuntimeAction,
		Category:    "diagnose",
		Description: "Read runtime evidence (agent_error_log, site_work_items, orchestration_states) for a site/correlation; read-only",
		IsLocal:     true,
	},
	"diagnose_assemble_bundle": {
		Handler:     DiagnoseAssembleBundleAction,
		Category:    "diagnose",
		Description: "Compose the diagnosis bundle: hypothesis + in-scope symbol bodies (read at commit_sha) + runtime evidence, for the verdict step",
		IsLocal:     true,
	},
	"diagnose_route": {
		Handler:     DiagnoseRouteAction,
		Category:    "diagnose",
		Description: "Per-iteration loop controller: apply the engine guards + call-graph re-scope to the verdict, return a next_step override to iterate (back to gather) or stop (to emit); read-only",
		IsLocal:     true,
	},
	"diagnose_emit": {
		Handler:     DiagnoseEmitAction,
		Category:    "diagnose",
		Description: "Format the human-facing diagnosis + evidence trail from the loop's terminal state; read-only, never a fix",
		IsLocal:     true,
	},
	"diagnose_persist_fix_plan": {
		Handler:     DiagnosePersistFixPlanAction,
		Category:    "diagnose",
		Description: "Validate the fix-proposer's constrained edit plan and persist it to diagnosis_artifacts (kind=fix_plan); writes no code — F1.1b turns an approved plan into a branch",
		IsLocal:     true,
	},
	"diagnose_council_decide": {
		Handler:     DiagnoseCouncilDecideAction,
		Category:    "diagnose",
		Description: "Deterministic council decision over the fix-plan reviewers' verdicts (veto→rejected, object→revise, else approved; hard_veto_from honored); persists kind=council_report; sets should_revise/should_reframe for the decision router",
		IsLocal:     true,
	},
	"diagnose_run_checks": {
		Handler:     DiagnoseRunChecksAction,
		Category:    "diagnose",
		Description: "Run the council reviewers' read-only verification queries (checks: [{sql,why}]) under the data_request containment (lint, READ ONLY tx, EXPLAIN gate, capped rows) and hand results to the next repropose",
		IsLocal:     true,
	},
	"diagnose_escalate": {
		Handler:     DiagnoseEscalateAction,
		Category:    "diagnose",
		Description: "Persist the human hand-off package (kind=escalation: decision + diagnosis + final plan + reviews) when the council rejects or the revise budget exhausts; a first-class success terminal, not a failure",
		IsLocal:     true,
	},
	"diagnose_prepare_fix_commit": {
		Handler:     DiagnosePrepareFixCommitAction,
		Category:    "diagnose",
		Description: "F1.1b(c) safety core: validate the fix-implementer's whole-file outputs against the approved plan's HARD file allowlist (reject out-of-plan, incomplete, empty, no-op) and assemble the git-adapter commit + PR payloads",
		IsLocal:     true,
	},
	"diagnose_build_gate": {
		Handler:     DiagnoseBuildGateAction,
		Category:    "diagnose",
		Description: "F1.1b(c) build gate (owner: no PRs for broken code): run gofmt (changed files) + targeted go build for the fix branch in a golang k8s Job; returns passed/log as a RESULT the workflow routes on — red never becomes a PR",
		IsLocal:     true,
	},
	"diagnose_read_repo_files": {
		Handler:     DiagnoseReadRepoFilesAction,
		Category:    "diagnose",
		Description: "Fetch the CURRENT bodies of the approved plan's modify/add files via the GitHub contents API (read token from the spawn gate) so sketch_to_files rewrites real code; a modify file that 404s is a hard error",
		IsLocal:     true,
	},
	"git_adapter_request": {
		Handler:     GitAdapterRequestAction,
		Category:    "site",
		Description: "Generic git-adapter caller (allowlisted verbs: commit, create_branch, create_pull_request) with data assembled from config paths/literals; awaits the adapter response; the write credential never leaves the adapter",
		IsLocal:     true,
	},
	"fixloop_digest": {
		Handler:     FixloopDigestAction,
		Category:    "diagnose",
		Description: "The awareness surface: deterministic digest (no LLM) of fix-loop runs, council decisions, gate/PR outcomes, and agent-config snapshots in a window; persisted to doc_notes (pipeline/diagnose, categories digest+fixloop)",
		IsLocal:     true,
	},
	"diagnose_triage": {
		Handler:     DiagnoseTriageAction,
		Category:    "diagnose",
		Description: "Deterministic router (no LLM) from the operational immune system into the fix loop: escalates loud-failure PATTERNS (deduped, capped) to needs_diagnosis and surfaces capability_gap (no-handler-yet) items to the roadmap; one doc_note per sweep",
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
	// DOCUMENTATION PROJECT - Tool plans, lasting, travelling plans e.g. with acceptance criteria
	// =========================================================================
	"persist_diagnosis_note": {
		Handler:     PersistDiagnosisNoteAction,
		Category:    "documentation",
		Description: "Persist a diagnosis report as a NOTES entry when the subject is explicit",
		IsLocal:     true,
	},
	"write_doc_plan": {
		Handler:     WriteDocPlanAction,
		Category:    "documentation",
		Description: "Supersede-write the current PLAN doc for a tool/pipeline subject",
		IsLocal:     true,
	},
	"append_doc_note": {
		Handler:     AppendDocNoteAction,
		Category:    "documentation",
		Description: "Append one NOTES entry (row) for a tool/pipeline subject",
		IsLocal:     true,
	},
	"load_doc_context": {
		Handler:     LoadDocContextAction,
		Category:    "documentation",
		Description: "Load current PLAN + latest NOTES for a tool/pipeline subject",
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
