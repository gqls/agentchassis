# Register — research-agents

7 concepts, consolidated from 20 raw extractions (10 unique blocks, each
appearing twice due to exact whole-block duplication in the cluster input file)
across units U14, U15, U16, U18, U19, U21, U22, U24c, U24e, U26.

### RES-001 — vertical-exemplar-researcher — the exemplar-research relay hop
- **status:** deployed
- **status-evidence:** Three independent extraction units converge on the same verified result: builder_route(21) "§B4 CLOSED — QUALITY VERIFIED 2026-07-06 … ✔ CONSUMPTION PROVEN: the strategy's gap_opportunity QUOTES the hop … ✔ TRANSMISSION THREE HOPS DEEP"; NOTES_running_synthesis_v4(39) "Landscape verified: three real vertical leaders; causal synthesis (reasons not copies); confidence 0.82. Strategy QUOTES the hop and builds the moat on it" (2026-07-06); HANDOFF_builder_thread confirms "LIVE and quality-verified end to end on dartsonline.com".
- **what:** A new relay hop (`needs_vertical_research` → `vertical-exemplar-researcher`) inserted between the domain classifier and the strategist, closing a gap where the classifier captured `competitors_found` names but nothing ever researched them. Reuse-only agent (one DB row, zero new Go): reads specs → LLM selects 3 of the vertical's best exemplar sites (flat keys, own domain forbidden) → runs 3 deliberately shallow Firecrawl crawls (budget: limit 6, markdown, main-content only, depth 1 — vs adoption's one-site-deep 30/rawHtml/4) → synthesises per-exemplar success factors, cross-exemplar patterns, and a differentiation opportunity (reasons, not copies) → writes `site_specs` `aspect=vertical_landscape` → chains to `needs_strategy`. Verified end-to-end on dartsonline.com: real vertical leaders selected, causal synthesis (confidence 0.82), and the strategy step demonstrably quoted the landscape when building its differentiator. Inserted via reroute (classifier chains needs_vertical_research; researcher chains needs_strategy onward; priority 7 below strategy's 8); an optional strategist prompt nudge makes the strategy step weigh the new aspect. First deployment stalled on an `image_tag` default-value trap (seed migration copied `agent_definitions` columns from a donor missing the spawn-consumed `command`/`image_tag` columns, defaulting to the stale `latest` tag) — fixed by copying from a fresher donor, flagged as a recurring trap.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4; NOTES_running_synthesis_v4(39).md (2026-07-04 through 07-06, §B4 sequence); README_flows.md; HANDOFF_builder_thread.md#2; NNN_seed_vertical_exemplar_researcher(2).sql; NNN_reroute_classifier_to_vertical_research.sql; NNN_strategist_vertical_landscape_nudge.sql
- **relations:** work-item relay spine; adoption fidelity (contrasting crawl budget); coverage baseline (curated-list reuse candidate); image_tag trap (its spawn incident); roadmap-phase enforcement gap; site-spec-and-classifier
- **verify-later:** NNN_seed_vertical_exemplar_researcher.sql; vertical_landscape spec rows in site_specs; reroute migration chain_config

### RES-002 — research-agent (cited web research into research_results)
- **status:** deployed
- **status-evidence:** Defined active v1.0.575 (sql_for_agents_v2/024); idle timeout set (075); classifier v2 (003) depends on it. An earlier, separately-documented lineage (docs015 data-flow verification, docs012/017 legacy architecture) independently confirms the same step-chain live and states the standing principle: "Research is cited — all LLM-generated content must cite sources, which are stored."
- **what:** Web-search research specialist, present since the platform's legacy architecture and continuing as a current agent: composes a search query from raw inputs (extract_topic → build_search_query), web-searches, selects top URLs, batch-scrapes them via the webscrape adapter (prepare_urls → batch_webscrape), formats findings with snippet context, synthesises a JSON summary (key points, recommendations, confidence, full source attribution), and persists to research_results — returning a research_id / `[0][1]`-style citation markers that page-content-writer and other prompts consume. Spawned by page-content-writer and site-classifier v2; backing store for research-driven components (FAQ, long-form).
- **sources:** sql_for_agents_v2/024_research_agent.sql; 003_site_classifier.sql; docs015_data_flow_verification/001_data_flow_verification.md; docs012.../012_summary_of_all_before_this_in_this_folder.md#research-agent; docs017.../019b_agent_architecture_v5_with_tickets_news.md#Active-Agents
- **relations:** research_results table (RES-003); render_mode needs_research; batch_webscrape action; adapters (webscrape)
- **verify-later:** research_results table usage; prepare_urls/batch_webscrape/format_research_content in registry

### RES-003 — research_results with source attribution
- **status:** deployed
- **status-evidence:** Table created in 004 PART 5 with sources JSONB format (url, title, domain, accessed_at, quotes, relevance_score); 009 patches add result_type and data/findings columns the code expects; training exports read result_type='tool_recreation_training'.
- **what:** Research findings persisted per site/page/component with full source attribution and expiry (expires_at refresh signal); page_components.research_id links content to the research that informed it, with sources_displayed controlling on-page attribution. Also doubles as generic typed result storage (result_type), e.g. tool recreation training triples.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART5; docs/agent_docs/sql_for_tables/009_research_results.sql; docs/agent_docs/sql_for_tables/025_llm_call_log_rag_knowledge_base.sql#exports
- **relations:** research-agent (RES-002, its writer); content grounding; finetuning flywheel (training triples); content_items origin_research_id
- **verify-later:** research-agent writers; result_type vocabulary

### RES-004 — Chat differentiator ideation agent
- **status:** aspirational
- **status-evidence:** "A low-risk, internal use of the agent framework ... Also still needs work — treat the output as candidates, not commitments."
- **what:** A proposed internal agent (runs on our own data, no isolation concerns) that, given a domain + audience, runs the asset × AI-capability combination and proposes ranked candidate payable differentiators split into "test now (cheap)" vs "score/consider (expensive)", each naming the asset and capability it depends on. Can spawn sub-agents to research willingness-to-pay or check whether a data feed exists/what it costs. Re-runnable across all domains whenever a new AI capability is added — the mechanism for catching early-adopter opportunities. Idea generator feeding human judgement, not an automated builder.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#11
- **relations:** payable-differentiator framework; Capability watchlist (RES-006)
- **verify-later:** any ideation/differentiator agent definition

### RES-005 — Wayback/archive.org grounding method + limitation
- **status:** partial
- **status-evidence:** running_notes 2026-06-13(b): "archive.org: Claude CAN web_fetch archive pages but ONLY when a search surfaces the exact URL; canNOT enumerate CDX on demand and the sandbox can't reach archive.org directly."
- **what:** Each probe page is grounded in the domain's old vertical via a Wayback snapshot. Constraint: the sandbox can't reach archive.org directly and can't enumerate CDX on demand, so grounding a NEW domain requires the operator to supply the Wayback URL/snapshot, or Claude uses web search + the domain name.
- **sources:** traffic_probe_running_notes(27).md#2026-06-13-b; traffic_probe_runbook.md#3
- **relations:** feeds intent-probe page content
- **verify-later:** archive.org.results/{relojistas,wayfaringlondoner}

### RES-006 — Capability watchlist + real-world event watchlist (dual standing research workflows)
- **status:** aspirational
- **status-evidence:** `running_notes(44).md`: "The capability list in the method is a starter; the watchlist workflow itself isn't designed" (open thread, never closed within this file); later: "Real-world event watchlist promoted to a second standing workflow... Both are recurring research workflows that fire re-runs of ideation."
- **what:** Two proposed recurring background research workflows: (1) a capability watchlist tracking new AI capabilities that beat the model's self-knowledge (agentic browsing, million-token contexts, real-time voice, etc.) — the "early-adopter mechanism"; (2) an event/window watchlist tracking scheme deadlines, regulation changes and application windows per domain (proven by the agritec SFI26 Window 1 case, which turned a "consider later" candidate into "test now"). Both are meant to trigger automatic re-runs of the ideation method across domains, but the trigger mechanism itself was never designed/built within this archive's timeframe.
- **sources:** `running_notes(44).md` ("Capability watchlist warrants its own workflow", "Watchlist should track scheme/event windows, not just AI capabilities", "Real-world event watchlist promoted to a second standing workflow")
- **relations:** idea generation method; Chat differentiator ideation agent (RES-004)
- **verify-later:** whether any scheduled_task / agent implements this in the live chassis

### RES-007 — Deep-research domain insight agent
- **status:** abandoned
- **status-evidence:** 016 designs a "domain-insight-agent" deciding when deep social research pays ("Value Multiple: 50-100x"); tied to the abandoned domain-flipping context, though its research-orchestration DNA resembles the later research agents.
- **what:** A strategic classifier that assesses whether a domain/topic merits multi-platform deep research (Reddit/LinkedIn/Twitter/Facebook/YouTube community mining, influencer mapping, sentiment threading) versus standard development, then deploys the appropriate research agent squad to synthesise unique content, tools and FAQs from real community pain points — the claimed competitive moat over single-LLM or SEO-tool approaches.
- **sources:** docs/architecture/016-competitive-advantge.md#enhanced-domain-analysis-agent,#deep-research-workflows-by-domain-type
- **relations:** domain value maximisation; topic amplifier engine; current research-agents lineage (RES-001/RES-002)
- **verify-later:** n/a
