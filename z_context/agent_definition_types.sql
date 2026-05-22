clients_db=# SELECT type, version, description FROM agent_definitions;
type                  | version |                                                                                                                                                                                                                           description
---------------------------------------+---------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 researcher                            |       1 | Conducts thorough research and analysis
 site-component-architect              |       2 | Assembles empty HTML templates.
 content-researcher                    |       1 | Researches and gathers comprehensive information for website content
 build-site-planner                    |       1 | Handler-mode site planner for the build dispatch pipeline. Reads research and briefing from site_specs, plans site structure via LLM, validates plan, syncs pages to DB, creates page content work items.
 reasoning                             |       1 | Performs logical analysis and decision making
 thunder-reaper                        |       1 | Periodic sweep that finds Thunder Compute instances older than their max_uptime_hours and dispatches decommission_instance for each. Single-instance-per-tick; relies on scheduled_tasks pre_query to identify the next target.
 domain-research-classifier            |       1 | Researches a domain via web search and scrape, classifies site type, extracts identity signals, writes findings to site_specs, creates next work item.
 visual-design-auditor                 |       1 | Group auditor for visual design quality. Loads style collection and rendered HTML samples, runs algorithmic checks for colour consistency and spacing, then makes one LLM call for holistic visual assessment. Produces findings as work items.
 webdesign-agent                       |       1 | Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS.
 image-build-handler                   |       1 | Self-contained handler for image work items (logo, hero). Calls image-generator, stores asset in DB, deploys optimized image to git. Used by dispatch loop for needs_logo and needs_hero_image items.
 website-capture-firecrawl             |       1 | Captures website content using Firecrawl API with validation and extraction
 page-rebuild                          |       1 | Rebuilds specific pages flagged as needs_rebuild on an existing site. Loads site context from DB, generates fresh content, deploys. Skips planning, asset generation, and CSS. Used for maintenance: fixing stale pages, adding new pages, refreshing content.
 content-creator-testimonials          |       1 | Specialized in writing authentic, emotionally resonant customer testimonials
 site-scraper                          |       1 | Scrapes live websites to extract design context. Uses webscrape adapter with Firecrawl. Outputs site_context for webdesign-agent.
 site-work-orchestrator                |       1 | Unified build/maintenance orchestrator. Builds sites from work items in site_work_items table. Processes items by priority, calling appropriate handler agents. Compatible with pageflow-builder's planner and content writer.
 component-quality-auditor             |       1 | Scores existing content_components and creates regeneration work items for low-quality ones. Runs periodically to keep the component library healthy.
 chief-strategist                      |       2 | Site planner that outputs pages with component types (v2 - unified architecture)
 rag-test-agent                        |       1 | Minimal deterministic agent exercising rag_index and rag_lookup to verify chassis registration. Takes input_data.content and input_data.query.
 gpu-provisioner                       |       1 | Provisions a Thunder Compute instance by dispatching provision_instance to thunder-adapter. Awaits the response containing instance_ip, ssh_user, ssh_key_secret_name, provisioning_id, thunder_identifier, provisioned_at. Replaces migration-022 stub.
 site-asset-renderer                   |       1 | Renders /assets/js/snippets.js for a site and commits to git. Deterministic — no LLM. Triggered when js_snippets or component set changes, or invoked by webdesign-agent.
 site-planner                          |       1 | Analyzes brief and plans site structure: pages, components, style collection, asset needs. Single LLM call to create comprehensive plan.
 training-data-exporter                |       1 | Exports successful LLM calls from llm_call_log into training_exports schema (runs + rows tables) as ChatML + metadata records. Reads parameters from input_data (agent_type, step_name, model_filter, etc.). Must be invoked via training-data-export-orchestrator wrapper so it gets a dedicated pod.
 model-trainer                         |       1 | Orchestrates the QLoRA training kickoff: prepare data, provision GPU, launch training. The training-monitor scheduled task handles polling and completion.
 quality-discovery-agent               |       1 | Scans sites for quality issues: broken nav links (#slug instead of /page.html), placeholder/fabricated contact info, generic unthemed CSS. All checks are algorithmic — no LLM budget needed. Writes findings to site_work_items.
 training-launcher                     |       1 | STUB — SCPs scripts and dataset to the GPU instance, SSH-execs training in background, updates training_runs.status. Currently a no-op pass-through to unblock orchestrator testing; real implementation pending.
 portfolio-architect                   |       1 | Assembles portfolio/showcase sites with galleries, case studies, and visual layouts
 multipage-website-builder             |       1 | Builds large websites (20+ pages) using batched generation to avoid token limits
 ch-matcher                            |       1 | Matches verified businesses against the local ch_vet_companies mirror table using postcode + name similarity. No API calls. Updates ch_vet_companies.matched_business_id for confirmed matches.
 nav-updater                           |       1 | Refreshes navigation tables from current pages, re-renders header/footer components with updated nav, then reassembles and deploys all pages. Algorithmic only - no LLM calls. Used when pages are added/removed/renamed and nav needs updating across the site.
 content-gap-planner                   |       1 | Plans how to fill content gaps identified by audits. Reads gap descriptions and site context, decides whether to create new pages, add sections to existing pages, or update specs. Creates actionable work items for page-build-handler.
 med-url-discoverer                    |       1 | Discovers product URLs by scraping retailer category pages
 med-export-orchestrator               |       1 | Spawns a temporary pod to export medicine prices as JSON to a configured site.
 copywriter                            |       1 | Creates compelling marketing and content copy
 landing-page-builder                  |       1 | Orchestrates the complete landing page build workflow - spawns specialist agents and coordinates them to build conversion-focused landing pages
 content-feed-trigger                  |       1 | Heartbeat agent: finds sites recommended for news feeds and dispatches content-feed-orchestrator per site. Runs on a 6-hour schedule.
 site-publisher                        |       1 | Publishes websites to storage buckets
 content-creator-contact               |       1 | Specialized in writing warm, inviting contact sections
 med-price-scrape-orchestrator         |       1 | Spawns a temporary pod to scrape veterinary medicine prices.
 training-data-preparer                |       1 | Worker: streams training_exports.rows to S3 JSONL and inserts a model_lifecycle.training_runs row in pending state.
 deployer-agent                        |       1 | Commits final site files to a Git repository by calling the git-adapter.
 content_researcher                    |       1 | Researches content for websites
 content-creator                       |       1 | Advanced AI-powered content generation with memory and style adaptation
 content-creator-cta                   |       1 | Specialized in writing persuasive call-to-action sections with urgency and clear value
 page-content-writer                   |       2 | Writes content for a single page. Loads section components, renders templates with brief data, calls LLM for content sections needing original writing. Can spawn research-agent for research-backed content.
 blog-content-planner                  |       1 | Plans initial blog posts for a site based on its industry, services, and target audience. Reads site specs, checks existing posts, asks LLM to plan 3-4 relevant posts, creates page records and work items for page-build-handler.
 domain-strategist                     |       1 | Determines optimal site strategy for a domain. Analyzes domain value, revenue models, competitive positioning. Outputs strategic guidance — site type, revenue model, content strategy, recommended page types. Does not design page architecture (that is the planner responsibility).
 content-creator-about                 |       1 | Specialized in writing compelling about page content that tells the story of a business or explains a concept
 multipage-website-builder             |       2 | Builds multi-page websites with consistent navigation and header injection
 visual-designer                       |       1 | Handles images, logos, and visual assets
 ch-collector                          |       1 | Bulk collects all Companies House companies with SIC 75000 (veterinary activities) into local mirror table. Run periodically to keep the mirror fresh.
 domain-analyst                        |       1 | Analyzes domains and determines appropriate website type
 site-component-architect              |       1 | Assembles empty HTML templates from the in-house component library.
 html-developer                        |       1 | Generates HTML/CSS/JS code for websites
 tool-deployer                         |       1 | Forks a library tool to a site. Creates the component fork, tool page, and page_component link. The page then flows through the normal render/deploy pipeline.
 asset-deployer                        |       1 | Deploys a single image asset: downloads from S3, optimizes by purpose, commits to git. Reusable for any image deploy task.
 site-component-linker                 |       1 | Links site_components rows to correct content_components from the style collection. Fixes NULL component_id that causes fallback rendering. Creates needs_rerender work item after linking.
 website-builder                       |       1 | Orchestrates complete website creation
 content-quality-auditor               |       1 | Group auditor for content quality. Loads the site brief, page content samples, and target audience. Makes one LLM call to assess tone alignment, content gaps, CTA effectiveness, and differentiation. Produces findings as work items.
 content-reviewer                      |       5 | Reviews page content for quality, accuracy, and brand alignment. Supports HITL mode (human review with edits) and auto-eval mode (LLM review with auto-approve or flag).
 internal-linker                       |       1 | Finds existing pages that should contextually link to an orphaned sub-page. Loads the target page and candidate pages with content samples, uses LLM to pick natural link placements, creates content_rewrite items for page-build-handler.
 tool-generator                        |       1 | Creates a new interactive tool from scratch via LLM. Generates self-contained HTML/JS/CSS based on the tool description, saves as a content_component, and creates a tool page. Used when no suitable library tool exists to fork.
 content-creator-hero                  |       1 | Specialized in writing compelling hero sections with powerful headlines and engaging subheadlines
 design-discovery-agent                |       1 | Scans sites for design-domain issues: undeployed assets, missing CSS, colour problems, missing style collections, deactivated components, stale header/footer renders, shared style collections. All algorithmic — no LLM budget.
 tool-improver                         |       1 | Incrementally improves deployed tools. Loads current HTML, applies LLM-driven fixes for rendering, mobile, UX, or accessibility issues, saves updated HTML, and triggers page re-render.
 improvement-loop                      |       1 | Post-build quality improvement cycle. Runs discovery agents to find issues, triages findings, dispatches fixes via build-dispatch-loop, and triggers rerender when fixes complete.
 domain-submitter                      |       1 | Entry point for new domain submissions. Creates site record, stores contact info, creates needs_domain_research work item for the classifier. Minimal input required — just the domain name.
 content-creator-contact               |       2 | Specialized in writing welcoming and effective contact page content
 med-json-exporter                     |       1 | Exports current medicine prices as JSON and commits to a site git repo. Configurable: domain, filters (species/category/retailer), output files, data path.
 site-adoption-agent                   |       1 | Crawls an existing site, analyses structure and content, creates specs and work items to recreate it
 med-price-collector                   |       1 | Scrapes veterinary medicine prices from online pharmacies
 site-architect                        |       1 | Plans website structure and navigation
 vet-practice-verifier                 |       1 | Verifies and enriches a single veterinary practice record by searching the web, scraping the practice website, and extracting structured data via LLM.
 intake-orchestrator                   |       3 | Entry point for site creation: classifies project type, runs briefing, spawns appropriate builder agent
 site-review-agent                     |       1 | Strategic site review. Compares current site against original brief and dream spec. Asks: is the site achieving its purpose? Calls content-quality-auditor for content assessment, then runs its own strategic alignment review. Produces work items for content rewrites, new pages, and tone shifts.
 med-url-discover-orchestrator         |       1 | Spawns a temporary pod to discover product URLs from retailer category pages.
 feed-ingester                         |       1 | Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items
 site-adoption-orchestrator            |       1 | Spawns a temporary pod to run site-adoption-agent. Thin wrapper that gives each adoption its own Job pod for clean per-lifetime logs. Passes input_data through unchanged (target_url, destination_domain, etc).
 tool-auditor                          |       1 | LLM-based code review of deployed interactive tools. Reads full HTML/CSS/JS source, reasons through logic and layout, identifies bugs, mobile issues, UX problems, and accessibility gaps. Creates improve_tool items for fixable issues.
 briefing-agent                        |       1 | Executes briefing questionnaires - either via LLM inference or HITL collection
 rerender-site                         |       1 | Re-renders and deploys all pages for a site from stored components. Re-renders site-level components (header, footer, head), then spawns page-rerender agent for each page. Used after design changes, component updates, or nav changes.
 html-developer-chunked                |       1 | Generates HTML in smaller chunks to handle large sites without token limits
 maintenance-triage                    |       1 | Scans deployed sites for maintenance issues (stale pages, missing content, orphan nav items). Populates the maintenance_queue and dispatches specialist agents to resolve issues. Can scan one site or all sites.
 ch-accounts-fetcher                   |       1 | Fetches and parses Companies House filed accounts (iXBRL). Extracts net assets, total assets, employee count, turnover, profit/loss where available.
 content-creator                       |       2 | Creates content for component-based pages (v2 - receives current_page object)
 area-sweep-orchestrator               |       1 | Loads un-swept postcode districts from the search_areas table and dispatches area-sweep-discoverer agents for each one.
 page-build-handler                    |       1 | Wrapper handler for page-content-writer. Spawns the specialist, captures output, persists sections to page_components via save_page_sections, updates page status, assembles and deploys via page-rerender. Dispatch-loop compatible.
 ch-llm-reviewer                       |       1 | Reviews ambiguous Companies House matches using LLM judgment. Processes pending_llm_review entries from ch_vet_companies, classifies each as confirmed, rejected, or uncertain.
 med-url-mapper                        |       1 | Discovers product URLs across retailer sites using Firecrawl /map endpoint. Broader than category-page discovery.
 rerender-pages                        |       6 | Re-assembles all deployed pages from stored components. Optionally re-renders site-level components (header/footer/head) first when refresh_site_components=true.
 site-design-planner                   |       1 | Resolves composition (palette, layout, typography) for a site BEFORE webdesign-agent renders. Tag-overlap matching against the layout library, font-family matching against typography library, priority-cascade extraction of palette from design_reference / mission / design_intent. Installs the composition atomically into css_themes + style_collections + sites.style_collection_id + resolved_composition spec. All deterministic Go actions — no LLM.
 spec-updater                          |       1 | Applies spec updates from audit findings. Reads aspect/field/value from work item spec, merges into site_specs with versioning. No LLM — mechanical merge only.
 color-variable-fixer                  |       1 | Replaces hardcoded hex colors in component inline <style> blocks with CSS variable references. Fixes templates (permanent) and rendered HTML (immediate). No content changes.
 build-briefing-agent                  |       1 | Handler-mode briefing agent for the build dispatch pipeline. Reads domain research from site_specs, fetches the target builder questionnaire, uses LLM to answer all briefing questions, writes answers to site_specs, chains to site-planner.
 component-template-fixer              |       1 | Applies targeted fixes to site_components and page_components: CSS injection (nav flex), element removal (search icon), slot_name alignment. Routes on spec.fix_type.
 multipage-wrapper                     |       1 | Wraps single-page site into multi-page structure (index, about, contact)
 training-data-export-orchestrator     |       1 | Thin wrapper that spawns training-data-exporter in a dedicated pod and waits for the export to complete. This is the agent triggered via Kafka for manual exports. Passes input_data through to the worker.
 generic                               |       1 | No-op agent. Scheduled tasks that do work in pre_query CTEs send messages here. The workflow completes immediately.
 content-site-architect                |       1 | Assembles content/publishing sites with article grids, category nav, and ad zones
 website-extract-structured            |       1 | Extracts structured data from websites
 html-assembler                        |       1 | Assembles final HTML from template and content, injects CSS themes and JS snippets.
 endpoint-health-checker               |       1 | Periodic health check for AI endpoints (Ollama, Claude). Pings each endpoint in ai_endpoint_health and updates healthy/error status. Triggered by scheduler every 30s.
 build-pipeline-trigger                |       1 | Heartbeat agent: seeds build queue entries, finds sites with pending work items, triggers dispatch loop. One site per invocation.
 design-audit-agent                    |       1 | Top-level design audit orchestrator. Spawns visual-design-auditor for CSS/layout/colour checks and content-quality-auditor for tone/gaps/CTA checks. Aggregates findings and triages work items.
 tool-recreation-handler               |       1 | Recreates interactive tools, games, and applications from crawled source code. Two-stage: analyses the tool purpose and function, then generates working replacement code. Used for adoption of sites with JavaScript-heavy interactive pages.
 site-classifier                       |       1 | Analyzes domain and objective to determine site type and recommend appropriate builder group
 css-patch-agent                       |       1 | Applies targeted CSS fixes from audit findings. Reads the current stylesheet, uses LLM to generate a specific patch, and deploys. Does not regenerate the full theme.
 build-dispatch-loop                   |       1 | Processes one work item per invocation, then spawns itself if more remain. No loops or sub_workflows — each dispatch is a separate orchestration with clean logs.
 feed-triage                           |       1 | Scores ingested feed items for relevance to site spec — identity, vertical, values, legal rules. Pre-display gate.
 page-rerender                         |       1 | Renders and deploys a single page from stored sections. Called by rerender-pages for each page in a loop.
 content-site-builder                  |       1 | Orchestrates the complete content/publishing site build workflow
 research-agent                        |       1 | Researches topics via web search, extracts relevant quotes, synthesizes findings with full source attribution. Stores results in research_results table for citation.
 chief-strategist                      |       1 | Creates a "first-principles" Build Plan (e.g., AIDA, PAS) from a simple objective.
 med-url-map-orchestrator              |       1 | Spawns a temporary pod to discover product URLs via Firecrawl /map endpoint.
 completeness-discovery-agent          |       1 | Scans sites for content completeness and integrity: empty sections, cross-site company name contamination, unrendered Go template syntax. All algorithmic — no LLM budget.
 landing-page-architect                |       1 | Assembles conversion-focused landing pages from component library (PAS, AIDA, etc.)
 pageflow-builder                      |      20 | Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time.
 section-editor                        |       1 | Performs granular edits to individual page sections. Updates content_data (source of truth) then re-renders from template with full site context. Edits survive future re-renders.
 tool-suggester                        |       1 | Evaluates what interactive tools would benefit a site based on its industry, services, audience, and existing pages. Uses LLM judgment — not limited to library catalogue. Creates add_tool work items for each recommendation.
 image-generator                       |       1 | Creates images using AI generation
 simple-content-writer-with-approval   |       1 | Generates content for organisations and waits for human approval before completion
 nav-link-fixer                        |       1 | Fixes broken navigation links in header/footer. Updates content_component templates to use {{.url}} instead of #{{.slug}}, then force re-renders site_components.
 business-intel                        |       1 | Business intelligence enrichment agent. Currently handles Companies House enrichment for verified businesses. Will expand to cover other data sources and verticals.
 web-search                            |       1 | Searches the internet for information
 site-deployer                         |       1 | Commits assembled site to git repository. Works with intake orchestrator flow.
 area-sweep-discoverer                 |       1 | Searches for veterinary practices within a UK postcode district and creates discovery candidates for unknown businesses.
 image-generator                       |       2 | Creates images using AI generation with S3 storage (orchestrator mode)
 content-creator-features              |       1 | Specialized in writing clear, customer-focused feature descriptions
 calculator                            |       1 | Performs mathematical calculations including addition, multiplication, and other operations
 vet-batch-processor                   |       1 | Loads pending verification tasks from the queue and processes them sequentially by spawning vet-practice-verifier agents. Designed for single-pod, low-throughput, polite data collection.
 component-creator                     |       1 | Generates new HTML component templates from section type descriptions. Processes needs_new_component work items. Stores results in content_components with selection metadata for future reuse.
 work-item-archiver                    |       1 | Archives terminal work items (complete, failed, wont_fix) older than 7 days to site_work_items_archive
 ch-detail-fetcher                     |       1 | Fetches detailed company data (profile, officers, PSC) from CH API for confirmed matches. Derives succession risk signals and stores in companies_house_data.
 content-writer                        |       1 | Creates website content from brief data and template requirements. Works with intake orchestrator flow.
 brand-designer                        |       1 | Analyzes domain, industry, and objectives to select or generate custom CSS themes and brand guidelines
 content-feed-orchestrator             |       1 | Per-site orchestrator: dispatches feed-ingesters for due sources, produces latest-news JSON, commits to git
 site-strategist                       |       1 | Creates strategic build plans using behavioral psychology and briefing data. Works with intake orchestrator flow.
 vet-pipeline-orchestrator             |       1 | Runs the full vet discovery pipeline: sweep areas, promote candidates, dispatch verifiers. Uses rolling execution so each run advances work from previous runs.
 content-creator-hero-without-research |       1 | Generates hero sections for websites without performing research; uses direct input only.
(138 rows)
