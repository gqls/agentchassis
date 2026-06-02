clients_db=# SELECT type, description FROM agent_definitions;
type                  |                                                                                                                                                                                                                           description
---------------------------------------+-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
 researcher                            | Conducts thorough research and analysis
 site-component-architect              | Assembles empty HTML templates.
 content-researcher                    | Researches and gathers comprehensive information for website content
 webdesign-agent                       | Generates CSS stylesheets for sites. Accepts site_context or loads from DB. Analyzes design requirements and generates production CSS.
 thunder-reaper                        | Periodic sweep that finds Thunder Compute instances older than their max_uptime_hours and dispatches decommission_instance for each. Single-instance-per-tick; relies on scheduled_tasks pre_query to identify the next target.
 domain-research-classifier            | Researches a domain via web search and scrape, classifies site type, extracts identity signals, writes findings to site_specs, creates next work item.
 build-site-planner                    | Handler-mode site planner for the build dispatch pipeline. Reads research and briefing from site_specs, plans site structure via LLM, validates plan, syncs pages to DB, creates page content work items.
 reasoning                             | Performs logical analysis and decision making
 visual-design-auditor                 | Group auditor for visual design quality. Loads style collection and rendered HTML samples, runs algorithmic checks for colour consistency and spacing, then makes one LLM call for holistic visual assessment. Produces findings as work items.
 image-build-handler                   | Self-contained handler for image work items (logo, hero). Calls image-generator, stores asset in DB, deploys optimized image to git. Used by dispatch loop for needs_logo and needs_hero_image items.
 site-scraper                          | Scrapes live websites to extract design context. Uses webscrape adapter with Firecrawl. Outputs site_context for webdesign-agent.
 website-capture-firecrawl             | Captures website content using Firecrawl API with validation and extraction
 site-work-orchestrator                | Unified build/maintenance orchestrator. Builds sites from work items in site_work_items table. Processes items by priority, calling appropriate handler agents. Compatible with pageflow-builder's planner and content writer.
 page-rebuild                          | Rebuilds specific pages flagged as needs_rebuild on an existing site. Loads site context from DB, generates fresh content, deploys. Skips planning, asset generation, and CSS. Used for maintenance: fixing stale pages, adding new pages, refreshing content.
 component-quality-auditor             | Scores existing content_components and creates regeneration work items for low-quality ones. Runs periodically to keep the component library healthy.
 gpu-provisioner                       | Provisions a Thunder Compute instance by dispatching provision_instance to thunder-adapter. Awaits the response containing instance_ip, ssh_user, ssh_key_secret_name, provisioning_id, thunder_identifier, provisioned_at. Replaces migration-022 stub.
 content-creator-testimonials          | Specialized in writing authentic, emotionally resonant customer testimonials
 chief-strategist                      | Site planner that outputs pages with component types (v2 - unified architecture)
 rag-test-agent                        | Minimal deterministic agent exercising rag_index and rag_lookup to verify chassis registration. Takes input_data.content and input_data.query.
 quality-discovery-agent               | Scans sites for quality issues: broken nav links (#slug instead of /page.html), placeholder/fabricated contact info, generic unthemed CSS. All checks are algorithmic — no LLM budget needed. Writes findings to site_work_items.
 site-asset-renderer                   | Renders /assets/js/snippets.js for a site and commits to git. Deterministic — no LLM. Triggered when js_snippets or component set changes, or invoked by webdesign-agent.
 site-planner                          | Analyzes brief and plans site structure: pages, components, style collection, asset needs. Single LLM call to create comprehensive plan.
 portfolio-architect                   | Assembles portfolio/showcase sites with galleries, case studies, and visual layouts
 training-data-exporter                | Exports successful LLM calls from llm_call_log into training_exports schema (runs + rows tables) as ChatML + metadata records. Reads parameters from input_data (agent_type, step_name, model_filter, etc.). Must be invoked via training-data-export-orchestrator wrapper so it gets a dedicated pod.
 multipage-website-builder             | Builds large websites (20+ pages) using batched generation to avoid token limits
 med-price-scrape-orchestrator         | Spawns a temporary pod to scrape veterinary medicine prices.
 nav-updater                           | Refreshes navigation tables from current pages, re-renders header/footer components with updated nav, then reassembles and deploys all pages. Algorithmic only - no LLM calls. Used when pages are added/removed/renamed and nav needs updating across the site.
 domain-submitter                      | Entry point for new domain submissions. Creates site record, stores contact info, creates needs_domain_research work item for the classifier. Minimal input required — just the domain name.
 website-builder                       | Orchestrates complete website creation
 tool-deployer                         | Forks a library tool to a site. Creates the component fork, tool page, and page_component link. The page then flows through the normal render/deploy pipeline.
 content-quality-auditor               | Group auditor for content quality. Loads the site brief, page content samples, and target audience. Makes one LLM call to assess tone alignment, content gaps, CTA effectiveness, and differentiation. Produces findings as work items.
 med-export-orchestrator               | Spawns a temporary pod to export medicine prices as JSON to a configured site.
 model-trainer                         | Orchestrates the QLoRA training kickoff: prepare data, provision GPU, launch training. The training-monitor scheduled task handles polling and completion.
 deployer-agent                        | Commits final site files to a Git repository by calling the git-adapter.
 content-creator-contact               | Specialized in writing warm, inviting contact sections
 landing-page-builder                  | Orchestrates the complete landing page build workflow - spawns specialist agents and coordinates them to build conversion-focused landing pages
 page-content-writer                   | Writes content for a single page. Loads section components, renders templates with brief data, calls LLM for content sections needing original writing. Can spawn research-agent for research-backed content.
 content-feed-trigger                  | Heartbeat agent: finds sites recommended for news feeds and dispatches content-feed-orchestrator per site. Runs on a 6-hour schedule.
 ch-collector                          | Bulk collects all Companies House companies with SIC 75000 (veterinary activities) into local mirror table. Run periodically to keep the mirror fresh.
 training-launcher                     | STUB — SCPs scripts and dataset to the GPU instance, SSH-execs training in background, updates training_runs.status. Currently a no-op pass-through to unblock orchestrator testing; real implementation pending.
 domain-analyst                        | Analyzes domains and determines appropriate website type
 ch-matcher                            | Matches verified businesses against the local ch_vet_companies mirror table using postcode + name similarity. No API calls. Updates ch_vet_companies.matched_business_id for confirmed matches.
 content-gap-planner                   | Plans how to fill content gaps identified by audits. Reads gap descriptions and site context, decides whether to create new pages, add sections to existing pages, or update specs. Creates actionable work items for page-build-handler.
 med-url-discoverer                    | Discovers product URLs by scraping retailer category pages
 copywriter                            | Creates compelling marketing and content copy
 content-creator-contact               | Specialized in writing welcoming and effective contact page content
 site-publisher                        | Publishes websites to storage buckets
 training-data-preparer                | Worker: streams training_exports.rows to S3 JSONL and inserts a model_lifecycle.training_runs row in pending state.
 content_researcher                    | Researches content for websites
 content-creator-cta                   | Specialized in writing persuasive call-to-action sections with urgency and clear value
 content-creator                       | Advanced AI-powered content generation with memory and style adaptation
 blog-content-planner                  | Plans initial blog posts for a site based on its industry, services, and target audience. Reads site specs, checks existing posts, asks LLM to plan 3-4 relevant posts, creates page records and work items for page-build-handler.
 tool-improver                         | Incrementally improves deployed tools. Loads current HTML, applies LLM-driven fixes for rendering, mobile, UX, or accessibility issues, saves updated HTML, and triggers page re-render.
 domain-strategist                     | Determines optimal site strategy for a domain. Analyzes domain value, revenue models, competitive positioning. Outputs strategic guidance — site type, revenue model, content strategy, recommended page types. Does not design page architecture (that is the planner responsibility).
 multipage-website-builder             | Builds multi-page websites with consistent navigation and header injection
 page-build-handler                    | Wrapper handler for page-content-writer. Spawns the specialist, captures output, persists sections to page_components via save_page_sections, updates page status, assembles and deploys via page-rerender. Dispatch-loop compatible.
 design-discovery-agent                | Scans sites for design-domain issues: undeployed assets, missing CSS, colour problems, missing style collections, deactivated components, stale header/footer renders, shared style collections. All algorithmic — no LLM budget.
 content-creator-about                 | Specialized in writing compelling about page content that tells the story of a business or explains a concept
 site-architect                        | Plans website structure and navigation
 site-review-agent                     | Strategic site review. Compares current site against original brief and dream spec. Asks: is the site achieving its purpose? Calls content-quality-auditor for content assessment, then runs its own strategic alignment review. Produces work items for content rewrites, new pages, and tone shifts.
 visual-designer                       | Handles images, logos, and visual assets
 med-price-collector                   | Scrapes veterinary medicine prices from online pharmacies
 improvement-loop                      | Post-build quality improvement cycle. Runs discovery agents to find issues, triages findings, dispatches fixes via build-dispatch-loop, and triggers rerender when fixes complete.
 asset-deployer                        | Deploys a single image asset: downloads from S3, optimizes by purpose, commits to git. Reusable for any image deploy task.
 site-component-architect              | Assembles empty HTML templates from the in-house component library.
 site-component-linker                 | Links site_components rows to correct content_components from the style collection. Fixes NULL component_id that causes fallback rendering. Creates needs_rerender work item after linking.
 endpoint-health-checker               | Periodic health check for AI endpoints (Ollama, Claude). Pings each endpoint in ai_endpoint_health and updates healthy/error status. Triggered by scheduler every 30s.
 build-pipeline-trigger                | Heartbeat agent: seeds build queue entries, finds sites with pending work items, triggers dispatch loop. One site per invocation.
 training-data-export-orchestrator     | Thin wrapper that spawns training-data-exporter in a dedicated pod and waits for the export to complete. This is the agent triggered via Kafka for manual exports. Passes input_data through to the worker.
 feed-ingester                         | Fetches content from a single source (RSS, news search, LLM news, scrape) and writes to content_feed_items
 html-developer                        | Generates HTML/CSS/JS code for websites
 vet-practice-verifier                 | Verifies and enriches a single veterinary practice record by searching the web, scraping the practice website, and extracting structured data via LLM.
 site-adoption-orchestrator            | Spawns a temporary pod to run site-adoption-agent. Thin wrapper that gives each adoption its own Job pod for clean per-lifetime logs. Passes input_data through unchanged (target_url, destination_domain, etc).
 content-reviewer                      | Reviews page content for quality, accuracy, and brand alignment. Supports HITL mode (human review with edits) and auto-eval mode (LLM review with auto-approve or flag).
 internal-linker                       | Finds existing pages that should contextually link to an orphaned sub-page. Loads the target page and candidate pages with content samples, uses LLM to pick natural link placements, creates content_rewrite items for page-build-handler.
 area-sweep-orchestrator               | Loads un-swept postcode districts from the search_areas table and dispatches area-sweep-discoverer agents for each one.
 tool-generator                        | Creates a new interactive tool from scratch via LLM. Generates self-contained HTML/JS/CSS based on the tool description, saves as a content_component, and creates a tool page. Used when no suitable library tool exists to fork.
 rerender-site                         | Re-renders and deploys all pages for a site from stored components. Re-renders site-level components (header, footer, head), then spawns page-rerender agent for each page. Used after design changes, component updates, or nav changes.
 content-creator                       | Creates content for component-based pages (v2 - receives current_page object)
 ch-llm-reviewer                       | Reviews ambiguous Companies House matches using LLM judgment. Processes pending_llm_review entries from ch_vet_companies, classifies each as confirmed, rejected, or uncertain.
 med-url-mapper                        | Discovers product URLs across retailer sites using Firecrawl /map endpoint. Broader than category-page discovery.
 multipage-wrapper                     | Wraps single-page site into multi-page structure (index, about, contact)
 site-adoption-agent                   | Crawls an existing site, analyses structure and content, creates specs and work items to recreate it
 website-extract-structured            | Extracts structured data from websites
 content-creator-hero                  | Specialized in writing compelling hero sections with powerful headlines and engaging subheadlines
 med-json-exporter                     | Exports current medicine prices as JSON and commits to a site git repo. Configurable: domain, filters (species/category/retailer), output files, data path.
 intake-orchestrator                   | Entry point for site creation: classifies project type, runs briefing, spawns appropriate builder agent
 maintenance-triage                    | Scans deployed sites for maintenance issues (stale pages, missing content, orphan nav items). Populates the maintenance_queue and dispatches specialist agents to resolve issues. Can scan one site or all sites.
 site-design-planner                   | Resolves composition (palette, layout, typography) for a site BEFORE webdesign-agent renders. Tag-overlap matching against the layout library, font-family matching against typography library, priority-cascade extraction of palette from design_reference / mission / design_intent. Installs the composition atomically into css_themes + style_collections + sites.style_collection_id + resolved_composition spec. All deterministic Go actions — no LLM.
 html-assembler                        | Assembles final HTML from template and content, injects CSS themes and JS snippets.
 design-audit-agent                    | Top-level design audit orchestrator. Spawns visual-design-auditor for CSS/layout/colour checks and content-quality-auditor for tone/gaps/CTA checks. Aggregates findings and triages work items.
 med-url-discover-orchestrator         | Spawns a temporary pod to discover product URLs from retailer category pages.
 tool-auditor                          | LLM-based code review of deployed interactive tools. Reads full HTML/CSS/JS source, reasons through logic and layout, identifies bugs, mobile issues, UX problems, and accessibility gaps. Creates improve_tool items for fixable issues.
 color-variable-fixer                  | Replaces hardcoded hex colors in component inline <style> blocks with CSS variable references. Fixes templates (permanent) and rendered HTML (immediate). No content changes.
 briefing-agent                        | Executes briefing questionnaires - either via LLM inference or HITL collection
 spec-updater                          | Applies spec updates from audit findings. Reads aspect/field/value from work item spec, merges into site_specs with versioning. No LLM — mechanical merge only.
 rerender-pages                        | Re-assembles all deployed pages from stored components. Optionally re-renders site-level components (header/footer/head) first when refresh_site_components=true.
 html-developer-chunked                | Generates HTML in smaller chunks to handle large sites without token limits
 ch-accounts-fetcher                   | Fetches and parses Companies House filed accounts (iXBRL). Extracts net assets, total assets, employee count, turnover, profit/loss where available.
 page-rerender                         | Renders and deploys a single page from stored sections. Called by rerender-pages for each page in a loop.
 build-briefing-agent                  | Handler-mode briefing agent for the build dispatch pipeline. Reads domain research from site_specs, fetches the target builder questionnaire, uses LLM to answer all briefing questions, writes answers to site_specs, chains to site-planner.
 content-site-architect                | Assembles content/publishing sites with article grids, category nav, and ad zones
 med-url-map-orchestrator              | Spawns a temporary pod to discover product URLs via Firecrawl /map endpoint.
 component-template-fixer              | Applies targeted fixes to site_components and page_components: CSS injection (nav flex), element removal (search icon), slot_name alignment. Routes on spec.fix_type.
 site-classifier                       | Analyzes domain and objective to determine site type and recommend appropriate builder group
 pageflow-builder                      | Component-based website builder. Spawns specialist agents (planner, content writer, reviewer, deployer), uses DB components for structure, LLM only for content. Builds and deploys pages one at a time.
 css-patch-agent                       | Applies targeted CSS fixes from audit findings. Reads the current stylesheet, uses LLM to generate a specific patch, and deploys. Does not regenerate the full theme.
 image-generator                       | Creates images using AI generation
 generic                               | No-op agent. Scheduled tasks that do work in pre_query CTEs send messages here. The workflow completes immediately.
 tool-recreation-handler               | Recreates interactive tools, games, and applications from crawled source code. Two-stage: analyses the tool purpose and function, then generates working replacement code. Used for adoption of sites with JavaScript-heavy interactive pages.
 build-dispatch-loop                   | Processes one work item per invocation, then spawns itself if more remain. No loops or sub_workflows — each dispatch is a separate orchestration with clean logs.
 feed-triage                           | Scores ingested feed items for relevance to site spec — identity, vertical, values, legal rules. Pre-display gate.
 content-site-builder                  | Orchestrates the complete content/publishing site build workflow
 research-agent                        | Researches topics via web search, extracts relevant quotes, synthesizes findings with full source attribution. Stores results in research_results table for citation.
 chief-strategist                      | Creates a "first-principles" Build Plan (e.g., AIDA, PAS) from a simple objective.
 completeness-discovery-agent          | Scans sites for content completeness and integrity: empty sections, cross-site company name contamination, unrendered Go template syntax. All algorithmic — no LLM budget.
 landing-page-architect                | Assembles conversion-focused landing pages from component library (PAS, AIDA, etc.)
 section-editor                        | Performs granular edits to individual page sections. Updates content_data (source of truth) then re-renders from template with full site context. Edits survive future re-renders.
 tool-suggester                        | Evaluates what interactive tools would benefit a site based on its industry, services, audience, and existing pages. Uses LLM judgment — not limited to library catalogue. Creates add_tool work items for each recommendation.
 simple-content-writer-with-approval   | Generates content for organisations and waits for human approval before completion
 nav-link-fixer                        | Fixes broken navigation links in header/footer. Updates content_component templates to use {{.url}} instead of #{{.slug}}, then force re-renders site_components.
 business-intel                        | Business intelligence enrichment agent. Currently handles Companies House enrichment for verified businesses. Will expand to cover other data sources and verticals.
 web-search                            | Searches the internet for information
 site-deployer                         | Commits assembled site to git repository. Works with intake orchestrator flow.
 area-sweep-discoverer                 | Searches for veterinary practices within a UK postcode district and creates discovery candidates for unknown businesses.
 image-generator                       | Creates images using AI generation with S3 storage (orchestrator mode)
 content-creator-features              | Specialized in writing clear, customer-focused feature descriptions
 calculator                            | Performs mathematical calculations including addition, multiplication, and other operations
 vet-batch-processor                   | Loads pending verification tasks from the queue and processes them sequentially by spawning vet-practice-verifier agents. Designed for single-pod, low-throughput, polite data collection.
 component-creator                     | Generates new HTML component templates from section type descriptions. Processes needs_new_component work items. Stores results in content_components with selection metadata for future reuse.
 work-item-archiver                    | Archives terminal work items (complete, failed, wont_fix) older than 7 days to site_work_items_archive
 ch-detail-fetcher                     | Fetches detailed company data (profile, officers, PSC) from CH API for confirmed matches. Derives succession risk signals and stores in companies_house_data.
 content-writer                        | Creates website content from brief data and template requirements. Works with intake orchestrator flow.
 brand-designer                        | Analyzes domain, industry, and objectives to select or generate custom CSS themes and brand guidelines
 content-feed-orchestrator             | Per-site orchestrator: dispatches feed-ingesters for due sources, produces latest-news JSON, commits to git
 site-strategist                       | Creates strategic build plans using behavioral psychology and briefing data. Works with intake orchestrator flow.
 vet-pipeline-orchestrator             | Runs the full vet discovery pipeline: sweep areas, promote candidates, dispatch verifiers. Uses rolling execution so each run advances work from previous runs.
 content-creator-hero-without-research | Generates hero sections for websites without performing research; uses direct input only.
(138 rows)

