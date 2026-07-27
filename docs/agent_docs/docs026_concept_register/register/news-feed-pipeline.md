# Register — news-feed-pipeline

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

19 concepts, consolidated from 44 raw extractions (22 unique blocks, each duplicated
once in the source cluster file) across units U01, U02, U10, U12, U13, U17a, U18,
U19, U21, U23.

### NEWS-001 — News feed pipeline (sources → async ingest → triage → JSON render → commit)
- **status:** deployed
- **status-evidence:** "006 status table all ✅ with gaswholesalers evidence; scheduled 6-hourly"; robot-hands replication "9 content_sources seeded, 0 erroring" (2026-07-10); vonc RUNNING_NOTES 2026-06-25 "all confirmed active v1.0.1078" (read from agent definitions and action source).
- **what:** The end-to-end live chain: content-feed-trigger is a 6-hour heartbeat (scheduled_tasks name='content-feed-refresh') that finds sites whose classification spec recommends a news feed (content_features.news_feed.recommended) and need refresh, dispatching content-feed-orchestrator per site (max 5) → seed_sources/seed_content_sources (from the classification spec) → dispatch ingesters async per due source (feed-ingester: rss / news_search / api_news / scrape, routed by source_type) → feed-triage scores PRIOR runs' items (two-pass by design: ingest now, triage next run) → render_news_section builds latest-news.json (6 items) + news-archive.json (20 items, if a news-index page exists) from a Go struct → git_commit. Homepage snippet + /news.html listing page both client-fetch the JSON via the component's own extracted JS — news updates are decoupled from page rerender. Adding news to an existing site is enrichment, not re-plan: evaluate_news_feed writes classification.content_features.news_feed, news-section-addition amends the plan (RULE 11 places latest-news on the homepage), using content_sources as pure data rows (no new code). This is the platform's general model for any static-site runtime data feed (mirrored by the Phase-3 provocation pipeline and the med JSON export pattern).
- **sources:** 006_news_feed_pipeline (full); HANDOFF_robot_hands_rebuild.md#PIPELINE-RECON; old/README_news_pipeline.md; 087_feeds_triage_ingester_orchestrator_etc.sql; 089_latest_news.sql; 090_b_content_feed_trigger.sql; 090_content_feed_orchestrator.sql; docs/RUNNING_NOTES_vonc(36).md#2026-06-25/#2026-06-29
- **relations:** growth budget (content-governance CGV-004); content-gap-planner chain (CGV-015); source diversity (NEWS-004); content_sources & feed-item lifecycle schema (NEWS-018); Phase-3 provocation pipeline (clone)
- **verify-later:** content_sources/content_feed_items tables; content-feed-refresh task; render_news_section_action.go; feed-triage real implementation vs stub

### NEWS-002 — Feed triage: relevance + credibility + source-attribution provenance
- **status:** partial
- **status-evidence:** 006 marked deployed, but Known Open Issues: "credibility always 0 … fields exist but aren't being populated".
- **what:** LLM triage scores relevance 0-100 and credibility high/medium/low with attribution chain {original_source, found_via, source_tier} across a 6-tier source taxonomy; rejects fabricated URLs, nav links, uncorroborated low-credibility claims. Status lifecycle ingested→relevant/review/rejected→expired(30d).
- **sources:** 006#feed-triage, #Issues; #Resolved Decisions 47
- **relations:** diversity scoring plan (NEWS-005); Grok provider choice (NEWS-003); feed triage scoring repair (NEWS-007, an earlier, different bug in the same subsystem)
- **verify-later:** credibility population bug fixed?

### NEWS-003 — Real-time-search news providers (Grok Responses API decision)
- **status:** deployed
- **status-evidence:** 006 resolved decision 48; hallucinated-URL bug table entry.
- **what:** api_news sources route to xAI grok-4-1-fast via Responses API with web_search+x_search, OpenAI Responses (gpt-4.1-mini) or Perplexity sonar — all real-time search; chat-completions grok-3-mini hallucinated 2023 URLs and was dropped.
- **sources:** 006#fetch_llm_news provider routing
- **relations:** feed triage credibility (NEWS-002)
- **verify-later:** provider keys in personae-default-secrets

### NEWS-004 — Render source-diversity interleaving
- **status:** aspirational
- **status-evidence:** 006 status table ✅; decision recorded after single-source domination.
- **stage2-verified (2026-07-14):** deployed → aspirational — render_news_section_action.go:loadNewsItems (line 338+) has no ROW_NUMBER()/PARTITION BY source_id anywhere in file (grep 0 hits); actual query is a plain ORDER BY status/relevance_score/published_at with LIMIT — no per-source cap/interleaving logic exists. Present-tense-plan false positive.
- **what:** loadNewsItems uses ROW_NUMBER() OVER (PARTITION BY source_id) ordered by source_rank then recency so each source contributes at most ~2 of 6 display slots; with topic-focused sources this also yields topical diversity.
- **sources:** 006#Render action source diversity, #Content Diversity §6
- **relations:** topic-focused source splitting (planned, NEWS-005); pipeline architecture (NEWS-001)
- **verify-later:** render_news_section_action.go query

### NEWS-005 — Content diversity & original research pipeline (readership segments, timelines, scenario analysis, engagement)
- **status:** aspirational
- **status-evidence:** 006 "What's Not Built Yet" lists every piece (topic splitting ready-to-implement; article-rewriter/feed-publisher/feed-lifecycle blocked/unbuilt).
- **what:** Planned evolution: topic-focused source splitting (SQL-only), coverage-gap pre-fetch step, multi-language regional discovery with triage translation, triage diversity scoring, research-agent multi-step investigations (fact/history/quotes/numbers) → writer targeted per readership segment (procurement/ops/trading/strategy) → eval agent quality gate → publish; continuous annotated timelines with pattern recognition; if/then scenario analysis (no predictions); client-side engagement measurement feeding content planning.
- **sources:** 006#Expansion Roadmap, #Content Diversity & Research Pipeline
- **relations:** research-agents; batch API integration; mid-era news design (NEWS-016, an earlier vision covering similar unbuilt ground: article-rewriter, publisher, lifecycle decay)
- **verify-later:** none built — check for article-rewriter definition

### NEWS-006 — News publishing gap (curation → deployed posts)
- **status:** aspirational
- **status-evidence:** "News — pipeline exists, publishing doesn't … The pipeline ends at curation" (2026-05).
- **what:** Ingestion/triage/diversity produce latest-news.json per site but nothing turns curated items into deployed blog posts; Path B connects news ingestion to page deployment via page-content-writer with a news-feed input, passing the site's deployed tool list for cross-linking.
- **sources:** FOCUS_interactive_content_generation(4).md#News, #Path-B
- **relations:** feed triage fixes (NEWS-007); topic splitting (NEWS-005); "Insights section" Tier-2 target (NEWS-010)
- **verify-later:** whether an article-publishing step now exists in news pipeline

### NEWS-007 — Feed triage scoring repair (config reads + wrapper unwrap)
- **status:** deployed
- **status-evidence:** "Triage is working. As of last check: 41 relevant, 23 rejected, 232 ingested (backlog clearing at 15 items per cycle)" (2026-04-17).
- **what:** 200+ items sat unscored since April 2nd due to three stacked bugs: LLM output truncation (max_items 50 → 15; max_tokens → 8192), config literal invisible to inputs.GetInt (use GetIntField on StepConfig.Config), and the execute_llm_prompt wrapper map ({type,result}) never unwrapped. Topic splitting of the single Grok source into topic-focused sources planned (SQL-only).
- **sources:** HANDOFF_2026-04-17_triage_and_component_linking.md#1, #4
- **relations:** chassis input conventions; content-feed-trigger workflow bug; feed triage credibility bug (NEWS-002, a later, different bug in the same triage stage)
- **verify-later:** feed_triage_actions.go; content_feed_items backlog state

### NEWS-008 — News pipeline replication and the news enrichment pattern
- **status:** deployed
- **status-evidence:** "News pipeline: LIVE and healthy (replicate gas)" (2026-05-20 recon); robot-hands "9 content_sources seeded, 0 erroring" (2026-07-10).
- **what:** content_sources rows of four parallel types (rss, news_search, api_news with grok/web-search tools, scrape) serve as the replication template for adding news to a new site — pure data rows, no new code. evaluate_news_feed writes classification.content_features.news_feed; news-section-addition amends the plan (RULE 11 places latest-news on the homepage). Two distinct components serve it (latest-news card grid on index; news-listing full page). Item expiry happens via status transition; the expires_at column exists but is unpopulated.
- **sources:** HANDOFF_robot_hands_rebuild.md#PIPELINE-RECON, old/README_news_pipeline.md, PLAN_imagery_best_in_class.md#Phase-I0-status
- **relations:** deploy_page files_field dependency (NEWS-012, news JS silently dropped otherwise); news imagery builds on it; pipeline architecture (NEWS-001)
- **verify-later:** robot-hands content_sources rows; content_feed_items lifecycle counts

### NEWS-009 — Price-news TTL and news→infographic enhancements
- **status:** aspirational
- **status-evidence:** "Both (a) and (b) are NICE-TO-HAVE… backlog" (2026-05-20, user-stated not urgent).
- **what:** (a) Price-aware filtering with short expiry: classify fetched news for price-movement items and expire them after 1-2 days via the existing-but-unused expires_at column plus a topics-based triage tag; per-site vertical. (b) News→infographic: pick 1-2 items, research the subject, generate an infographic — ties into the imagery infographic kind, research adapters, and the data-graph pipeline when data-driven.
- **sources:** HANDOFF_robot_hands_rebuild.md#NEWS-ENHANCEMENTS, old/README_news_pipeline.md
- **relations:** data-graph pipeline; news imagery partially absorbs (b)
- **verify-later:** expires_at population; any price-tagging triage rule

### NEWS-010 — "Insights section" as the Tier-2 news-feed expansion target
- **status:** superseded
- **status-evidence:** Archive Tier 2 = "Insights section... Future"; live Tier 2 = "News listing page... ✅ Working," curated/rewritten-article idea folded into Tier 3.
- **what:** The original three-tier roadmap treated a dedicated `/insights/` section of rewritten, curated articles as the second expansion tier after homepage snippets. When the archive-first news-index/listing page was actually built, it took the Tier-2 slot instead, and the "curated rewritten articles" idea was pushed down into Tier 3, where `article-rewriter` and `feed-publisher` remain listed as not-yet-built in both versions.
- **sources:** old/older1/006_news_feed_pipeline.md#"Expansion Roadmap"; docs024_key_docs_latest/006_news_feed_pipeline_v2.md#"Expansion Roadmap"
- **relations:** article-rewriter/feed-publisher agents (still unbuilt); news publishing gap (NEWS-006)
- **verify-later:** check whether a `/insights/` route or `article-rewriter` agent definition exists anywhere

### NEWS-011 — News rendering three-layer architecture (data / behaviour / structure+style)
- **status:** deployed
- **status-evidence:** "Discovered and fixed 2026-05-19/20 during the gaswholesalers.com news rollout"; FOCUS doc header "Consolidated from ... the fix-plan half of findings_and_plan_news_visual.md".
- **what:** News (and any data-driven component) rendering splits into three independently produced/deployed layers: Data (content_feed_items → /data/*.json via render_news_section), Behaviour (content_components.js_content → /tools/assets/{function}.js via rerender_single_page), and Structure+style (html_template + css_snippets, inlined per page). They connect only at runtime in the browser via fetch. This separation is deliberate and is what allows multiple independent news views per site.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Rendering-layer, js_snippets_news_gaswholesalers/old/006_news_feed_pipeline_addendum_rendering.md
- **relations:** files_field deploy mechanism (NEWS-012); two-news-components pattern (NEWS-013); component asset coupling gap
- **verify-later:** render_news_section action, rerender_single_page action, content_components.js_content/html_template columns, css_snippets table

### NEWS-012 — files_field vs content_field git_commit deploy bug (component JS assets silently dropped)
- **status:** deployed
- **status-evidence:** "Verification: Single-page rerenders of index and news both returned files_count: 2" — fixed via jsonb_set, no code change, structural (site-wide) fix.
- **what:** The `page-rerender` workflow's `deploy_page` step was configured with `content_field: "rendered_page.html"` (HTML only) instead of `files_field: "rendered_page.files"` (HTML + all component JS). `git_commit`'s `extractFilesForGit` had three extraction methods and the wrong one was selected, so every component's `js_content` was computed but discarded before ever reaching git — for every site, since inception. Fixed by a config-only jsonb_set edit; applies structurally to all components/sites, present and future.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Resolved, js_snippets_news_gaswholesalers/old/006_news_feed_pipeline_addendum_rendering.md, js_snippets_news_gaswholesalers/TODO_remaining_work.md#Done
- **relations:** news rendering three-layer architecture (NEWS-011); rendered_html snapshot-not-view pattern
- **verify-later:** page-rerender agent_definition deploy_page step config, git_commit action extractFilesForGit

### NEWS-013 — Two distinct news components as a multi-view pattern (latest-news vs news-listing)
- **status:** deployed
- **status-evidence:** Table of both components' function/JS/data pairing verified live on gaswholesalers.com.
- **what:** `latest-news` (homepage card grid, curated top 6) and `news-listing` (full archive list) are two separate content_components rows, each with its own template, JS, and data file — not duplicates. The architecture generalizes: adding a new filtered/styled news view requires only a new content_components row + CSS + a data-producing step; the deploy mechanism is generic over component function name, no workflow change needed.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#The-two-news-components-and-their-pairings, js_snippets_news_gaswholesalers/old/002_how_webdesign_handles_snippets.md
- **relations:** news rendering three-layer architecture (NEWS-011); component asset coupling gap
- **verify-later:** content_components rows id 77dafa26 (latest-news), 11d4dc21 (news-listing)

### NEWS-014 — rerender-pages refresh-flag coupling (three concerns behind one flag)
- **status:** aspirational
- **status-evidence:** "Future improvement (note for backlog)... Low effort, modest value, do it next time we touch rerender-pages versions" — not implemented as of TODO_remaining_work.md.
- **what:** The `rerender-pages` workflow ties three conceptually-independent refresh operations (site components re-render, JS snippets rebuild+deploy, blog-listing rebuild) behind a single `refresh_site_components` boolean. Proposed fix: split into three independent flags.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#rerender-pages-workflow-findings, js_snippets_news_gaswholesalers/old/rerender_pages_workflow_findings.md
- **relations:** rebuild_blog_listing news-index gap (NEWS-015); two rerender paths (NEWS-016)
- **verify-later:** grep/inspect `rerender-pages`; `refresh_site_components`

### NEWS-015 — rebuild_blog_listing does not handle news-index pages (silent no-op gap)
- **status:** partial
- **status-evidence:** "the step is a silent no-op. It logs 'No blog page found, skipping' ... news visuals do get updated [via later page_rerender]. But there is no equivalent news-listing rebuild".
- **what:** `RebuildBlogListingAction`'s `findBlogPage` only matches `page_type='blog-index'` or `name='blog' AND page_type='content'` — never `page_type='news-index'` or `name='news'`. On news-only sites the step silently no-ops; would need a parallel `rebuild_news_listing`/`findNewsPage`.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_news_rendering_and_component_assets.md#Finding-1, js_snippets_news_gaswholesalers/old/rerender_pages_workflow_findings.md
- **relations:** rerender-pages refresh-flag coupling (NEWS-014)
- **verify-later:** grep/inspect `RebuildBlogListingAction`; `findBlogPage`; `page_type='blog-index'`

### NEWS-016 — Two rerender trigger paths (site-wide work-item batch vs single-page orchestration-only)
- **status:** deployed
- **status-evidence:** HANDOFF doc "Key facts to carry forward: Two rerender paths: site-wide rerender-pages creates site_work_items (item_type=page_rerender); single-page page-rerender is an orchestration only (no work item)".
- **what:** Site-wide `rerender-pages` creates `site_work_items` rows (batch, dispatched over time, load-bearing on reaper/OOM fragility); single-page `page-rerender` is triggered as a direct orchestration with no work-item row, used for quick manual/test verification of a fix.
- **sources:** js_snippets_news_gaswholesalers/old/HANDOFF_2026-05-21_faq_prevention_and_news.md
- **relations:** rerender-pages refresh-flag coupling (NEWS-014); reaper mechanisms and gap
- **verify-later:** grep/inspect `rerender-pages`; `site_work_items`; `page-rerender`

### NEWS-017 — Blog-listing / orphan-page routing session handoff
- **status:** deployed
- **status-evidence:** 102_blog_handoff header "Session Handoff — April 10 2026"; "Ready to Deploy (files generated, not yet applied)".
- **stage2-verified (2026-07-14):** partial → deployed — check_orphan_pages.go:12-155 implements exactly the 3-route classification (blog-post rerender via companion action; nav_drift->HandlerAgent nav-updater at line 125; needs_internal_links->HandlerAgent internal-linker at line 155). rebuild_blog_listing_action.go and check_orphan_pages.go both have commit histories (e...
- **what:** A dated operational handoff fixing blog-listing rendering (slot-name mismatch, empty-schema CSS-only template, missing article links) and reclassifying orphan pages into three routes (blog-post→rerender, nav-flags→nav-drift→nav-updater, no-nav→needs_internal_links→internal-linker). Documents self-hosted GitHub Actions runner deploy, the page-build-handler `error_step`-placement fix (46 validation crashes), the dedup pattern, and a future Mistral-Small-on-CPU internal-linker.
- **sources:** ED/102_blog_handoff-2026-04-10.md#completed-this-session, #ready-to-deploy-files-generated-not-yet-applied, #remaining-unresolved-groups-not-yet-addressed
- **relations:** work item lifecycle/unresolved; deployment-github; nav sync; link management
- **verify-later:** rebuild_blog_listing_action.go; check_orphan_pages.go; github-actions-runner; nav-updater/internal-linker

### NEWS-018 — News feed pipeline: content_sources and feed-item lifecycle schema
- **status:** deployed
- **status-evidence:** content_sources DDL (twice-iterated) + seed_boxing_sources function; content_feed_items in 018; applied handler-routing fixes and content-feed-refresh task (6h); live Grok config update to grok-4-1-fast with search_tools for gaswholesalers.
- **what:** Per-site content sources with typed configs — news_search (web search adapter), rss, api_news (LLM news via xAI/Grok incl. prompt_template, hours_lookback, search_tools), scrape (Firecrawl), api_data (structured APIs like BoE rates) — scheduled by fetch_interval/next_fetch_at with error tracking. Fetched items flow through content_feed_items' separate lifecycle (ingested→filtered→relevant→queued→published/rejected/expired/duplicate) with per-site relevance scoring, entity cross-referencing and dedup, becoming a site_work_items row only at publish time. Routing contract: missing_news_sources / stale_news_section / all_sources_erroring → content-feed-orchestrator; missing_news_section → content-gap-planner. Client-side rendering: news sections render from a static JSON artefact deployed alongside the site (/data/latest-news.json), so news refresh is a data commit, not a page rebuild — component ships noscript fallback, date humanisation (formatNewsDate), and canonical CSS in css_snippets.
- **sources:** docs/agent_docs/sql_for_tables/027_content_sources_table.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#028_news_feed_handler_routing_fixes; docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#news-feed-js; docs/agent_docs/sql_for_tables/044_css_snippets.sql
- **relations:** work queue; latest-news client rendering; scheduler; pipeline architecture (NEWS-001); med JSON export (same static-JSON publishing pattern)
- **verify-later:** content-feed-orchestrator workflow; feed item volumes; JSON writer for /data/latest-news.json; per-site adoption

### NEWS-019 — News & content feed pipeline (mid-era design, superseded)
- **status:** superseded
- **status-evidence:** v1 in docs017/030, refined in 019b (sub-agents feed-ingester/deduplicator/triage/article-rewriter/publisher/lifecycle), restructured in 023 ("Feed items go through ingestion, filtering, deduplication and relevance scoring before they become publishable" with work_item linkage); today's news-feed-pipeline (NEWS-001) is the deployed descendant.
- **what:** Per-site content sources (RSS/API/scrape/entity_event) polled on configurable intervals → raw content_feed_items → dedup (near-duplicate headline detection) → LLM triage (relevance, urgency, angle for THIS site) → article-rewriter producing original articles in site voice with entity cross-links and required disclaimers → publication as pages → time-based lifecycle decay (featured 0-24h → current → aging → archive → prune, with per-site-type pacing and event-calendar coupling). Later revision: publishable items become site_work_items (handler article-writer) and rewritten articles become entities; news display is a design concern owned by the component/theme system, not the feed.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/030_news_feeds_v1.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#7-News-and-Content-Feed-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Content-Feed-Items
- **relations:** entity news_triggers; work items; pipeline architecture (NEWS-001, successor); content diversity roadmap (NEWS-005, revives similar unbuilt ambitions)
- **verify-later:** content_sources/content_feed_items current schema vs these designs
