
<!-- SOURCE: U14_docs019_runbooks.md -->
### vertical-exemplar-researcher — the exemplar-research relay hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "§B4 CLOSED — QUALITY VERIFIED 2026-07-06 … ✔ CONSUMPTION PROVEN: the strategy's gap_opportunity QUOTES the hop … ✔ TRANSMISSION THREE HOPS DEEP".
- **what:** The first new build of the builder route: a reuse-only agent (one DB row, zero new Go) inserted as needs_vertical_research between classifier and strategist. Twelve-step workflow: read specs → LLM exemplar selection (3 of the vertical's best sites, flat keys, own domain forbidden) → 3× shallow firecrawl + format → synthesis LLM (per-exemplar success factors, cross-exemplar patterns, adopt/adapt/avoid lessons, differentiation opportunity — REASONS NOT COPIES) → write_site_spec aspect=vertical_landscape → chain needs_strategy. Verified end-to-end on dartsonline: real vertical leaders selected, causal synthesis, quoted by the strategy, differentiator surfaced in the plan. Design calls: shallow-many vs adoption's deep-one; specs-not-messages; strategist prompt nudge so the research is read.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (design calls, change-set, re-run verification)
- **relations:** work-item relay spine; adoption fidelity; coverage baseline (curated-list reuse candidate); image_tag trap (its spawn incident)
- **verify-later:** NNN_seed_vertical_exemplar_researcher.sql; vertical_landscape spec rows; reroute migration chain_config

<!-- SOURCE: U15_docs019_running_notes.md -->
### Vertical-exemplar-researcher / competitor synthesis hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** "§B4 CLOSED on quality... Landscape verified: three real vertical leaders; causal synthesis (reasons not copies); confidence 0.82. Strategy QUOTES the hop and builds the moat on it" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** A new relay hop (`needs_vertical_research` → `vertical-exemplar-researcher`) inserted between the domain classifier and the strategist to close a gap where the classifier captured `competitors_found` names but nothing ever researched them: it runs shallow crawls of 3 vertical exemplars (vs. adoption's one deep crawl of the site itself), synthesises causal reasons (not copied content) into a `site_specs` row (`aspect=vertical_landscape`), which the strategist prompt reads wholesale and demonstrably used to shape a real site's differentiator. Its first live deployment stalled because the seed migration copied `agent_definitions` columns from a donor missing the spawn-consumed `command`/`image_tag` columns (defaulted to the stale `latest` image tag) — fixed by copying from a fresher donor and flagged as a recurring `image_tag` default-value trap.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 through 07-06 (§B4 sequence, full).
- **relations:** Work-item relay / builder-generations architecture; roadmap-phase enforcement gap.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Vertical-exemplar research hop (best-of-niche synthesis into vertical_landscape)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread: "§B4 vertical-exemplar-researcher LIVE and quality-verified end to end on dartsonline.com … causal synthesis (confidence 0.82) → strategy QUOTING the landscape".
- **what:** A new relay hop between classification and strategy: find the vertical's best existing sites, read three of them shallowly (deliberate budget: limit 6, markdown, main-content only, depth 1 — vs adoption's one-site-deep 30/rawHtml/4), and distil WHY they succeed — reasons, not copies — into spec aspect vertical_landscape for the strategist and planner. Reuse-only (every step an existing action; the whole agent is one DB row, no Go, no image build); written as a spec because specs are the per-site shared memory across hops; inserted via reroute (classifier chains needs_vertical_research; researcher chains needs_strategy onward; priority 7 below strategy's 8 in the ascending ladder); an optional strategist prompt nudge makes the strategy step weigh the new aspect (research nobody reads is wasted). First bare-domain→deployed-site milestone followed.
- **sources:** README_flows.md (the plain-language explainer); NNN_seed_vertical_exemplar_researcher(2).sql; NNN_reroute_classifier_to_vertical_research.sql; NNN_strategist_vertical_landscape_nudge.sql; HANDOFF_builder_thread.md#2
- **relations:** relay spine; adoption pipeline (contrasting crawl budget); site-spec-and-classifier
- **verify-later:** vertical-exemplar-researcher row; vertical_landscape aspects in site_specs

<!-- SOURCE: U18_sql_for_agents.md -->
### research-agent (cited web research into research_results)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Defined active v1.0.575 (v2/024); idle timeout set in 075; classifier v2 (003) depends on it.
- **what:** Web-search research specialist that extracts relevant quotes, synthesises findings with full source attribution and stores in a research_results table for citation ([0], [1] markers consumed by page-content-writer prompts).
- **sources:** sql_for_agents_v2/024_research_agent.sql; 024_research_agent.sql; 003_site_classifier.sql
- **relations:** spawned by page-content-writer and site-classifier v2
- **verify-later:** research_results table usage

<!-- SOURCE: U19_sql_tables_components.md -->
### research_results with source attribution
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Table created in 004 PART 5 with sources JSONB format (url, title, domain, accessed_at, quotes, relevance_score); 009 patches add result_type and data/findings columns the code expects; training exports read result_type='tool_recreation_training'.
- **what:** Research findings persisted per site/page/component with full source attribution and expiry (expires_at refresh signal); page_components.research_id links content to the research that informed it, with sources_displayed controlling on-page attribution. Also doubles as generic typed result storage (result_type) e.g. tool recreation training triples.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART5; docs/agent_docs/sql_for_tables/009_research_results.sql; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#exports
- **relations:** content grounding; finetuning flywheel (training triples); content_items origin_research_id.
- **verify-later:** research-agent writers; result_type vocabulary.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Research agent with cited sources
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** docs015/001 step-by-step verified pipeline (extract_topic → build_search_query → web_search → prepare_urls → batch_webscrape → format_research_content → synthesize → insert_research_result); docs012/010 principle "Research is cited — all LLM-generated content must cite sources, which are stored."
- **what:** A self-contained research agent: composes a search query from raw inputs, searches, selects top URLs, batch-scrapes them via the webscrape adapter, formats findings with snippet context, synthesizes a JSON summary (key points, recommendations, confidence), and persists to research_results with full source list — returning a research_id that content sections reference. Backing store for research-driven components (FAQ, long-form).
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#research-agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** render_mode needs_research; batch_webscrape action; adapters (webscrape); current research-agents category.
- **verify-later:** research_results table; prepare_urls/batch_webscrape/format_research_content in registry.

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat differentiator ideation agent
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** "A low-risk, internal use of the agent framework ... Also still needs work — treat the output as candidates, not commitments."
- **what:** A proposed internal agent (runs on our own data, no isolation concerns) that, given a domain + audience, runs the asset × AI-capability combination and proposes ranked candidate payable differentiators split into "test now (cheap)" vs "score/consider (expensive)", each naming the asset and capability it depends on. Can spawn sub-agents to research willingness-to-pay or check whether a data feed exists/what it costs. Re-runnable across all domains whenever a new AI capability is added — the mechanism for catching early-adopter opportunities. Idea generator feeding human judgement, not an automated builder.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#11
- **relations:** payable-differentiator framework, research-agents
- **verify-later:** any ideation/differentiator agent definition

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Wayback/archive.org grounding method + limitation
- **category:** research-agents
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(b) "archive.org: Claude CAN web_fetch archive pages but ONLY when a search surfaces the exact URL; canNOT enumerate CDX on demand and the sandbox can't reach archive.org directly".
- **what:** Each probe page is grounded in the domain's old vertical via a Wayback snapshot. Constraint: the sandbox can't reach archive.org directly and can't enumerate CDX on demand, so grounding a NEW domain requires the operator to supply the Wayback URL/snapshot, or Claude uses web search + the domain name.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_runbook.md#3
- **relations:** feeds intent-probe page content
- **verify-later:** archive.org.results/{relojistas,wayfaringlondoner}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Capability watchlist + real-world event watchlist (dual standing research workflows)
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "The capability list in the method is a starter; the watchlist workflow itself isn't designed" (open thread, never closed within this file); later: "Real-world event watchlist promoted to a second standing workflow... Both are recurring research workflows that fire re-runs of ideation."
- **what:** Two proposed recurring background research workflows: (1) a capability watchlist tracking new AI capabilities that beat the model's self-knowledge (agentic browsing, million-token contexts, real-time voice, etc.) — the "early-adopter mechanism"; (2) an event/window watchlist tracking scheme deadlines, regulation changes and application windows per domain (proven by the agritec SFI26 Window 1 case, which turned a "consider later" candidate into "test now"). Both are meant to trigger automatic re-runs of the ideation method across domains, but the trigger mechanism itself was never designed/built within this archive's timeframe.
- **sources:** `running_notes(44).md` ("Capability watchlist warrants its own workflow", "Watchlist should track scheme/event windows, not just AI capabilities", "Real-world event watchlist promoted to a second standing workflow")
- **relations:** idea generation method
- **verify-later:** whether any scheduled_task / agent implements this in the live chassis

<!-- SOURCE: U26_misc_dirs.md -->
### Deep-research domain insight agent
- **category:** research-agents
- **status-signal:** abandoned
- **status-evidence:** 016 designs a "domain-insight-agent" deciding when deep social research pays ("Value Multiple: 50-100x"); tied to the abandoned domain-flipping context, though its research-orchestration DNA resembles the later research agents.
- **what:** A strategic classifier that assesses whether a domain/topic merits multi-platform deep research (Reddit/LinkedIn/Twitter/Facebook/YouTube community mining, influencer mapping, sentiment threading) versus standard development, then deploys the appropriate research agent squad to synthesise unique content, tools and FAQs from real community pain points — the claimed competitive moat over single-LLM or SEO-tool approaches.
- **sources:** docs/architecture/016-competitive-advantge.md#enhanced-domain-analysis-agent; docs/architecture/016-competitive-advantge.md#deep-research-workflows-by-domain-type
- **relations:** domain value maximisation; topic amplifier engine; current research-agents lineage
- **verify-later:** n/a

<!-- SOURCE: U14_docs019_runbooks.md -->
### vertical-exemplar-researcher — the exemplar-research relay hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "§B4 CLOSED — QUALITY VERIFIED 2026-07-06 … ✔ CONSUMPTION PROVEN: the strategy's gap_opportunity QUOTES the hop … ✔ TRANSMISSION THREE HOPS DEEP".
- **what:** The first new build of the builder route: a reuse-only agent (one DB row, zero new Go) inserted as needs_vertical_research between classifier and strategist. Twelve-step workflow: read specs → LLM exemplar selection (3 of the vertical's best sites, flat keys, own domain forbidden) → 3× shallow firecrawl + format → synthesis LLM (per-exemplar success factors, cross-exemplar patterns, adopt/adapt/avoid lessons, differentiation opportunity — REASONS NOT COPIES) → write_site_spec aspect=vertical_landscape → chain needs_strategy. Verified end-to-end on dartsonline: real vertical leaders selected, causal synthesis, quoted by the strategy, differentiator surfaced in the plan. Design calls: shallow-many vs adoption's deep-one; specs-not-messages; strategist prompt nudge so the research is read.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (design calls, change-set, re-run verification)
- **relations:** work-item relay spine; adoption fidelity; coverage baseline (curated-list reuse candidate); image_tag trap (its spawn incident)
- **verify-later:** NNN_seed_vertical_exemplar_researcher.sql; vertical_landscape spec rows; reroute migration chain_config

<!-- SOURCE: U15_docs019_running_notes.md -->
### Vertical-exemplar-researcher / competitor synthesis hop
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** "§B4 CLOSED on quality... Landscape verified: three real vertical leaders; causal synthesis (reasons not copies); confidence 0.82. Strategy QUOTES the hop and builds the moat on it" (NOTES_running_synthesis_v4(39).md, 2026-07-06).
- **what:** A new relay hop (`needs_vertical_research` → `vertical-exemplar-researcher`) inserted between the domain classifier and the strategist to close a gap where the classifier captured `competitors_found` names but nothing ever researched them: it runs shallow crawls of 3 vertical exemplars (vs. adoption's one deep crawl of the site itself), synthesises causal reasons (not copied content) into a `site_specs` row (`aspect=vertical_landscape`), which the strategist prompt reads wholesale and demonstrably used to shape a real site's differentiator. Its first live deployment stalled because the seed migration copied `agent_definitions` columns from a donor missing the spawn-consumed `command`/`image_tag` columns (defaulted to the stale `latest` image tag) — fixed by copying from a fresher donor and flagged as a recurring `image_tag` default-value trap.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 through 07-06 (§B4 sequence, full).
- **relations:** Work-item relay / builder-generations architecture; roadmap-phase enforcement gap.

<!-- SOURCE: U16_docs019_design_plans.md -->
### Vertical-exemplar research hop (best-of-niche synthesis into vertical_landscape)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** HANDOFF_builder_thread: "§B4 vertical-exemplar-researcher LIVE and quality-verified end to end on dartsonline.com … causal synthesis (confidence 0.82) → strategy QUOTING the landscape".
- **what:** A new relay hop between classification and strategy: find the vertical's best existing sites, read three of them shallowly (deliberate budget: limit 6, markdown, main-content only, depth 1 — vs adoption's one-site-deep 30/rawHtml/4), and distil WHY they succeed — reasons, not copies — into spec aspect vertical_landscape for the strategist and planner. Reuse-only (every step an existing action; the whole agent is one DB row, no Go, no image build); written as a spec because specs are the per-site shared memory across hops; inserted via reroute (classifier chains needs_vertical_research; researcher chains needs_strategy onward; priority 7 below strategy's 8 in the ascending ladder); an optional strategist prompt nudge makes the strategy step weigh the new aspect (research nobody reads is wasted). First bare-domain→deployed-site milestone followed.
- **sources:** README_flows.md (the plain-language explainer); NNN_seed_vertical_exemplar_researcher(2).sql; NNN_reroute_classifier_to_vertical_research.sql; NNN_strategist_vertical_landscape_nudge.sql; HANDOFF_builder_thread.md#2
- **relations:** relay spine; adoption pipeline (contrasting crawl budget); site-spec-and-classifier
- **verify-later:** vertical-exemplar-researcher row; vertical_landscape aspects in site_specs

<!-- SOURCE: U18_sql_for_agents.md -->
### research-agent (cited web research into research_results)
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Defined active v1.0.575 (v2/024); idle timeout set in 075; classifier v2 (003) depends on it.
- **what:** Web-search research specialist that extracts relevant quotes, synthesises findings with full source attribution and stores in a research_results table for citation ([0], [1] markers consumed by page-content-writer prompts).
- **sources:** sql_for_agents_v2/024_research_agent.sql; 024_research_agent.sql; 003_site_classifier.sql
- **relations:** spawned by page-content-writer and site-classifier v2
- **verify-later:** research_results table usage

<!-- SOURCE: U19_sql_tables_components.md -->
### research_results with source attribution
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** Table created in 004 PART 5 with sources JSONB format (url, title, domain, accessed_at, quotes, relevance_score); 009 patches add result_type and data/findings columns the code expects; training exports read result_type='tool_recreation_training'.
- **what:** Research findings persisted per site/page/component with full source attribution and expiry (expires_at refresh signal); page_components.research_id links content to the research that informed it, with sources_displayed controlling on-page attribution. Also doubles as generic typed result storage (result_type) e.g. tool recreation training triples.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART5; docs/agent_docs/sql_for_tables/009_research_results.sql; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#exports
- **relations:** content grounding; finetuning flywheel (training triples); content_items origin_research_id.
- **verify-later:** research-agent writers; result_type vocabulary.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Research agent with cited sources
- **category:** research-agents
- **status-signal:** deployed
- **status-evidence:** docs015/001 step-by-step verified pipeline (extract_topic → build_search_query → web_search → prepare_urls → batch_webscrape → format_research_content → synthesize → insert_research_result); docs012/010 principle "Research is cited — all LLM-generated content must cite sources, which are stored."
- **what:** A self-contained research agent: composes a search query from raw inputs, searches, selects top URLs, batch-scrapes them via the webscrape adapter, formats findings with snippet context, synthesizes a JSON summary (key points, recommendations, confidence), and persists to research_results with full source list — returning a research_id that content sections reference. Backing store for research-driven components (FAQ, long-form).
- **sources:** docs015_data_flow_verification/001_data_flow_verification.md; docs012_site_maps_and_components/012_summary_of_all_before_this_in_this_folder.md#research-agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** render_mode needs_research; batch_webscrape action; adapters (webscrape); current research-agents category.
- **verify-later:** research_results table; prepare_urls/batch_webscrape/format_research_content in registry.

<!-- SOURCE: U22_recent_small_docs.md -->
### Chat differentiator ideation agent
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** "A low-risk, internal use of the agent framework ... Also still needs work — treat the output as candidates, not commitments."
- **what:** A proposed internal agent (runs on our own data, no isolation concerns) that, given a domain + audience, runs the asset × AI-capability combination and proposes ranked candidate payable differentiators split into "test now (cheap)" vs "score/consider (expensive)", each naming the asset and capability it depends on. Can spawn sub-agents to research willingness-to-pay or check whether a data feed exists/what it costs. Re-runnable across all domains whenever a new AI capability is added — the mechanism for catching early-adopter opportunities. Idea generator feeding human judgement, not an automated builder.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#11
- **relations:** payable-differentiator framework, research-agents
- **verify-later:** any ideation/differentiator agent definition

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### Wayback/archive.org grounding method + limitation
- **category:** research-agents
- **status-signal:** partial
- **status-evidence:** running_notes 2026-06-13(b) "archive.org: Claude CAN web_fetch archive pages but ONLY when a search surfaces the exact URL; canNOT enumerate CDX on demand and the sandbox can't reach archive.org directly".
- **what:** Each probe page is grounded in the domain's old vertical via a Wayback snapshot. Constraint: the sandbox can't reach archive.org directly and can't enumerate CDX on demand, so grounding a NEW domain requires the operator to supply the Wayback URL/snapshot, or Claude uses web search + the domain name.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b, traffic_probe_runbook.md#3
- **relations:** feeds intent-probe page content
- **verify-later:** archive.org.results/{relojistas,wayfaringlondoner}

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Capability watchlist + real-world event watchlist (dual standing research workflows)
- **category:** research-agents
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "The capability list in the method is a starter; the watchlist workflow itself isn't designed" (open thread, never closed within this file); later: "Real-world event watchlist promoted to a second standing workflow... Both are recurring research workflows that fire re-runs of ideation."
- **what:** Two proposed recurring background research workflows: (1) a capability watchlist tracking new AI capabilities that beat the model's self-knowledge (agentic browsing, million-token contexts, real-time voice, etc.) — the "early-adopter mechanism"; (2) an event/window watchlist tracking scheme deadlines, regulation changes and application windows per domain (proven by the agritec SFI26 Window 1 case, which turned a "consider later" candidate into "test now"). Both are meant to trigger automatic re-runs of the ideation method across domains, but the trigger mechanism itself was never designed/built within this archive's timeframe.
- **sources:** `running_notes(44).md` ("Capability watchlist warrants its own workflow", "Watchlist should track scheme/event windows, not just AI capabilities", "Real-world event watchlist promoted to a second standing workflow")
- **relations:** idea generation method
- **verify-later:** whether any scheduled_task / agent implements this in the live chassis

<!-- SOURCE: U26_misc_dirs.md -->
### Deep-research domain insight agent
- **category:** research-agents
- **status-signal:** abandoned
- **status-evidence:** 016 designs a "domain-insight-agent" deciding when deep social research pays ("Value Multiple: 50-100x"); tied to the abandoned domain-flipping context, though its research-orchestration DNA resembles the later research agents.
- **what:** A strategic classifier that assesses whether a domain/topic merits multi-platform deep research (Reddit/LinkedIn/Twitter/Facebook/YouTube community mining, influencer mapping, sentiment threading) versus standard development, then deploys the appropriate research agent squad to synthesise unique content, tools and FAQs from real community pain points — the claimed competitive moat over single-LLM or SEO-tool approaches.
- **sources:** docs/architecture/016-competitive-advantge.md#enhanced-domain-analysis-agent; docs/architecture/016-competitive-advantge.md#deep-research-workflows-by-domain-type
- **relations:** domain value maximisation; topic amplifier engine; current research-agents lineage
- **verify-later:** n/a
