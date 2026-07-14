| CGV-001 | Section-editor architecture (content_data as source of truth) | deployed | Single-section edits via content_data update + re-render, never HTML patching | content-governance.md |
| CGV-002 | Granular editing spectrum & three edit paths | deployed | Routing model: direct edit / brief regenerate / page regenerate / direction propagate | content-governance.md |
| CGV-003 | Spatial addressing for natural-language editing | partial | data-pc-id/slot/position shipped at section level; element-level addressing unfulfilled | content-governance.md |
| CGV-004 | Page growth budget (growth_config, three-tier weekly limits) | deployed | CheckPageGrowthBudget caps content/blog/structural page creation per site per week | content-governance.md |
| CGV-005 | Human direction channels and the pinned direction spec | partial | Work-item / direction-update / reference-suggestion channels; direction resets audit pass | content-governance.md |
| CGV-006 | Two sources of truth for site contact email | unknown | sites.email vs site_specs.identity.email can drift; COALESCE band-aid, no consolidation | content-governance.md |
| CGV-007 | Standing-ambition default in the mission aspect | aspirational | Proposed domain-submitter default mission_brief so builds lead the vertical, not mirror it | content-governance.md |
| CGV-008 | Optimistic-lock co-management of shared rows across parallel chats | deployed | WHERE updated_at=<last-known> UPDATE; 0 rows means stop and coordinate | content-governance.md |
| CGV-009 | Snapshot-before-change backup conventions | deployed | snapshot_agent, manual component_versions inserts, CTAS bak tables before every mutation | content-governance.md |
| CGV-010 | Silent-fallback link family (phantom /contact.html, /services.html) | unknown | Components link to nonexistent pages instead of resolving or degrading gracefully | content-governance.md |
| CGV-011 | Content-regression guard on section save | deployed | save_page_sections blocks thinner regenerations from overwriting richer live content | content-governance.md |
| CGV-012 | Standards curation & governance — concern curators | aspirational | One flat curator agent per top-level concern, reusing the auditor pattern | content-governance.md |
| CGV-013 | Coordinator role (arbitrates and frames) | aspirational | Thin layer above curators owning taxonomy, cross-concern conflicts, human framing | content-governance.md |
| CGV-014 | vonc.com mini-lobby content-edit re-render scope-boundary question | deployed | content_data-only edit rule established; correct path for a structural trim left unclear | content-governance.md |
| CGV-015 | Blog/content planning agents (blog-content-planner, content-gap-planner, internal-linker) | deployed | LLM planners turning content gaps into pages, sections, or internal links | content-governance.md |
| CGV-016 | page_components.build_status CHECK constraint | deployed | Restricts build_status to a fixed enum after an invented 'approved' value hid a section | content-governance.md |
| CGV-017 | Schema-mode strict/flexible subsystem | abandoned | Approval-snapshot lock regime built, never wired up, dropped 2026-07-09 | content-governance.md |
| CGV-018 | content_items reusable content layer | unknown | Typed reusable content rows (headline/tagline/faq) built but apparently never written to | content-governance.md |
| CGV-019 | page_component_history full-snapshot content history | deployed | Full content_data snapshot before every write; rollback/audit substrate for edits | content-governance.md |
| CGV-020 | Section governance columns: content_brief and suppressed_sections | deployed | content_brief enables regeneration; suppressed_sections stops discovery resurrecting removals | content-governance.md |
| CGV-021 | Page-content-writer + admin content brief regeneration flow | partial | Bridge doc: admin edits brief -> Regenerate -> content_rewrite item with brief in prompt | content-governance.md |
| CGV-022 | Legal content agent + legal constraint rules | aspirational | Jurisdiction-aware legal pages and machine-readable disclaimer/forbidden-phrase rules | content-governance.md |
| CGV-023 | Content review flow with rejection -> needs_attention | deployed | HITL/auto-eval gate; rejected pages marked needs_attention and queued for maintenance | content-governance.md |
| CGV-024 | maintenance_queue + claim/complete/fail functions | partial | Generic site-maintenance work queue with SKIP LOCKED claim/complete/fail functions | content-governance.md |
| CGV-025 | maintenance_queue as generic install/uninstall trigger surface | partial | Reused maintenance_queue as a generic per-site add-on trigger, first for chatbot install | content-governance.md |
| CGV-026 | Recommendation-specialist architecture (bug vs recommendation vs gap) | abandoned | Proposed finding_type routing (bug/gap/recommendation) with approval_mode gate; never built | content-governance.md |
| CGV-027 | Privacy posture (no cookies/JS/IP; UK GDPR/PECR) | deployed | Low-risk privacy stance baked into the traffic-probe engine and pages | content-governance.md |
| CGV-028 | site_specs `pinned` flag not honoured by the write path | partial | WriteSiteSpec ignores pinned; only disabled improvement-sweep currently protects specs | content-governance.md |
| CQ-001 | Content-quality defect catalogue (gamesdesign.co.uk) | partial | Maintained catalogue of hero-CTA, brand-suffix, empty-footer/description defects | content-quality.md |
| CQ-002 | validate_page_content gate (pre-deploy content validator) | deployed | Blocker validator (placeholder/contamination/links/email) routing failures to human review | content-quality.md |
| CQ-003 | Shared-component regen clobber failure mode | deployed | Regenerating a shared component silently emptied every dependent page using old field names | content-quality.md |
| CQ-004 | Recovery playbook for stranded dependents (Route A vs Route B) | deployed | Full writer rebuild vs re-key + scoped re-render to recover pages after a contract change | content-quality.md |
| CQ-005 | F8 — shared-component contamination (three carriers) | partial | Site-specific product pitch baked into a shared component via fallbacks/merge/llm_guidance | content-quality.md |
| CQ-006 | Neutralize-in-place remediation pattern | deployed | Surgical jsonb patch to strip contamination when no clean component restore point exists | content-quality.md |
| CQ-007 | Adoption content-quality defect families (polish batch) | unknown | Brand-suffix titles, empty footer contact, duplicated H1s, empty meta descriptions | content-quality.md |
| CQ-008 | Post-build validation of structured components (Fix D) | aspirational | Proposed post-build assertion that required structured fields are actually populated | content-quality.md |
| CQ-009 | Site-quality programme handoff | partial | Baseline measurement (0 nav/img/svg/script) triggered a dedicated site-quality runbook | content-quality.md |
| CQ-010 | Placeholder-content suppression sweep | deployed | Placeholder text hidden behind HTML comment + needs_human_review item + rerender | content-quality.md |
| CQ-011 | Audited content pipeline (persona -> research -> draft -> veracity/copyright audits) | aspirational | Orchestrated content generation with fact-check and plagiarism/copyright audit stages | content-quality.md |
| CQ-012 | Prompt composition asymmetry (text cascade vs image) | aspirational | Deliberate choice to keep single-prepend image cascade separate from text composition | content-quality.md |
| CQ-013 | Input sanitisation (sanitizeValue, Cc/Cf stripping, NFD survives) | deployed | Engine strips control/format chars, collapses whitespace correctly, defers NFC to collector | content-quality.md |
| CQ-014 | component-template-fixer CTA reuse assumption — corrected | superseded | Plan wrongly assumed CTA-fix reuse existed; agent actually punts CTAs to needs_review | content-quality.md |
| CQ-015 | identity-advisor agent and sites.approval_mode gate — never built | abandoned | Confirmed absent: three-way finding_type routing and its specialist agents never built | content-quality.md |
| CQ-016 | LLM fabrication classes in self-built site content | deployed | Fictional staff, fake taxonomies, nonexistent capabilities invented and later removed | content-quality.md |
| CQ-017 | Anti-hype voice and claim-discipline spec | deployed | Reusable voice contract: banned hype language, smallest-true-claim, CTA governance | content-quality.md |
| NEWS-001 | News feed pipeline (sources -> async ingest -> triage -> JSON render -> commit) | deployed | 6h heartbeat -> orchestrator -> ingesters -> triage -> render_news_section -> git commit | news-feed-pipeline.md |
| NEWS-002 | Feed triage: relevance + credibility + source-attribution provenance | partial | LLM scores relevance/credibility/attribution; credibility field never actually populated | news-feed-pipeline.md |
| NEWS-003 | Real-time-search news providers (Grok Responses API decision) | deployed | api_news routes to Grok/OpenAI/Perplexity real-time search after chat-completions hallucinated URLs | news-feed-pipeline.md |
| NEWS-004 | Render source-diversity interleaving | deployed | ROW_NUMBER partition by source caps any one source at ~2 of 6 display slots | news-feed-pipeline.md |
| NEWS-005 | Content diversity & original research pipeline | aspirational | Roadmap: topic splitting, readership-segment writers, timelines, scenario analysis | news-feed-pipeline.md |
| NEWS-006 | News publishing gap (curation -> deployed posts) | aspirational | Curated news items never become deployed blog posts; Path B design to close the gap | news-feed-pipeline.md |
| NEWS-007 | Feed triage scoring repair (config reads + wrapper unwrap) | deployed | Three stacked bugs left 200+ items unscored; truncation/config/wrapper-unwrap fixes | news-feed-pipeline.md |
| NEWS-008 | News pipeline replication and the news enrichment pattern | deployed | content_sources rows as a pure-data replication template for adding news to a new site | news-feed-pipeline.md |
| NEWS-009 | Price-news TTL and news->infographic enhancements | aspirational | Nice-to-have backlog: short-expiry price news and news-driven infographic generation | news-feed-pipeline.md |
| NEWS-010 | "Insights section" as the Tier-2 news-feed expansion target | superseded | Planned curated-articles tier was displaced by the archive-first news-index page | news-feed-pipeline.md |
| NEWS-011 | News rendering three-layer architecture (data/behaviour/structure+style) | deployed | JSON data, component JS, and template/CSS deploy independently, joining only in-browser | news-feed-pipeline.md |
| NEWS-012 | files_field vs content_field git_commit deploy bug | deployed | deploy_page misconfigured field silently dropped all component JS from git since inception | news-feed-pipeline.md |
| NEWS-013 | Two distinct news components as a multi-view pattern | deployed | latest-news and news-listing are separate components, template for future filtered views | news-feed-pipeline.md |
| NEWS-014 | rerender-pages refresh-flag coupling | aspirational | One boolean conflates three independent refresh operations; split proposed, not done | news-feed-pipeline.md |
| NEWS-015 | rebuild_blog_listing does not handle news-index pages | partial | findBlogPage never matches news-index pages, so news-only sites silently no-op | news-feed-pipeline.md |
| NEWS-016 | Two rerender trigger paths (site-wide batch vs single-page orchestration) | deployed | rerender-pages creates work items; page-rerender is a direct no-work-item orchestration | news-feed-pipeline.md |
| NEWS-017 | Blog-listing / orphan-page routing session handoff | partial | Dated fix package for blog-listing rendering and three-way orphan-page reclassification | news-feed-pipeline.md |
| NEWS-018 | News feed pipeline: content_sources and feed-item lifecycle schema | deployed | Typed source configs and the ingested->published/expired/duplicate item lifecycle | news-feed-pipeline.md |
| NEWS-019 | News & content feed pipeline (mid-era design, superseded) | superseded | Earlier article-rewriter/publisher/lifecycle-decay design, ancestor of the deployed pipeline | news-feed-pipeline.md |
| TRF-001 | Traffic-probe mission/program (residual-traffic intent capture) | deployed | Parked domains get a probe page inviting one action to rank real rebuild demand | traffic-analytics.md |
| TRF-002 | Wayback grounding of probe pages | deployed | archive.org snapshot fixes vertical/language/invited-action before building a probe page | traffic-analytics.md |
| TRF-003 | Probe page pattern — one invited action, plausible framing | deployed | One-line tagline, single action, no JS/cookies, framing follows the domain's heritage | traffic-analytics.md |
| TRF-004 | Minimal-data privacy posture (UK GDPR/PECR) | deployed | No cookies/JS/IP/names logged; declared load-invariant even under traffic pressure | traffic-analytics.md |
| TRF-005 | Intent event record (fields, omissions, landing_query enrichment) | deployed | One row per submission with kind/value/ref_host/country; no IP/UA/cookies ever recorded | traffic-analytics.md |
| TRF-006 | Visit beacon and events-per-1k metric | deployed | No-JS 1x1 beacon counts human visits as denominator for the core intent-rate metric | traffic-analytics.md |
| TRF-007 | Capture-side input sanitisation with deferred normalisation | deployed | Engine strips Cc/Cf and collapses whitespace; NFC/lowercasing deferred to the collector | traffic-analytics.md |
| TRF-008 | /events export endpoint and checkpoint contract | deployed | Key-gated NDJSON export with strictly-after since param, lock-free, duplicate-free pulls | traffic-analytics.md |
| TRF-009 | Access-log passive harvest and /access-digest | deployed | Engine reads its own nginx log for referer/404/UA signals the event stream can't see | traffic-analytics.md |
| TRF-010 | intent_events table with structural idempotency | deployed | engine_event_id UNIQUE + ON CONFLICT makes overlapping collector pulls safely idempotent | traffic-analytics.md |
| TRF-011 | Intent collection topology (collector under wrapper-orchestrator) | partial | Thin orchestrator spawns a collector worker to pull events/stats from all VM-hosted sites | traffic-analytics.md |
| TRF-012 | Ingest validation contract | aspirational | Full spec for what the collector must enforce (shape checks, dedupe, NFC); not yet enabled | traffic-analytics.md |
| TRF-013 | intent_site_stats visit-count snapshot | partial | One-row-per-host cumulative /stats snapshot feeding the events-per-1k rate calculation | traffic-analytics.md |
| TRF-014 | Ranking queries and graduation criteria | partial | Six read-only queries answer "is there demand"; graduation threshold still a proposal | traffic-analytics.md |
| TRF-015 | intent-probe content component | deployed | New content_components row: plain HTML form + beacon, no JS, capturing anonymous intent | traffic-analytics.md |
| TRF-016 | Probe content restraint — no results, no imagery, no anchoring | deployed | Deliberately no results page and no photos in v1 so displayed content can't bias the signal | traffic-analytics.md |
| TRF-017 | Traffic-claim verification and the bot-vs-human verdict method | deployed | Method for testing marketplace visit claims against beacon and access-log ground truth | traffic-analytics.md |
| TRF-018 | Global bot-IP blocklist (Thread D) | aspirational | Idea to block illegitimate-crawler IPs globally across all boxes from the access-digest rollup | traffic-analytics.md |
| TRF-019 | Relojistas static-rebuild manifest (Thread A) | aspirational | Plan to package a domain's heritage/RSS/inbound-link signals into a static multi-page rebuild | traffic-analytics.md |
| TRF-020 | Domain shortlist and selection policy | deployed | Ranked parked-domain export and a policy for choosing which domains to probe first | traffic-analytics.md |
| TRF-021 | Per-domain notes and living-docs convention | deployed | Every probe domain gets a living notes file; project knowledge lives in plan/runbook/notes | traffic-analytics.md |
| TPI-001 | Audio-monitoring topic discovery with auto-spawned topic agents | abandoned | Podcast transcription -> novel-topic detection -> auto-spawned monitoring agent, unbuilt | topic-intelligence.md |
| TPI-002 | Topic amplifier / deep digger engine | abandoned | Full engineering design for topic collection/verification/dedup; no implementation trace | topic-intelligence.md |
| TPI-003 | Cross-domain intelligence network and subscription tiers | abandoned | Vision of sibling-domain intelligence sharing and paid subscription tiers; pure vision | topic-intelligence.md |
