# Register — business-strategy

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

30 concepts, consolidated from 84 raw extractions across units U01, U02, U03, U04, U06, U13, U17a, U18, U19, U20, U21, U22, U24b, U24e, U24f, U25, U26. (The cluster input file contained this category's raw blocks twice, back-to-back and byte-identical, in addition to genuine cross-unit duplication — both kinds are merged below.)

### BIZ-001 — Platform mission: best possible site per domain via one unified pipeline
- **status:** partial
- **status-evidence:** 028_platform_mission_and_pipeline_direction.md is a living document (second revision 2026-04-22), restated independently by the user again 2026-07-07; no later doc supersedes the mission itself, but delivery against it (fidelity dial, classifier) remains partial.
- **what:** The platform's anchoring mission, stated independently at least three times across different documents: given a domain name in any state, produce the best possible multipage website end-to-end through one agent graph/pipeline, with minimal human input. "Best" = most useful to probable visitors (measured by real engagement) AND best revenue via whatever model genuinely fits — the classifier decides the commercial shape, and defaulting to a generic brochure/consultancy site when no strong signal exists is a named failure mode to eliminate. One pipeline serves blank/adopted/missioned/replication domains, differing only in input material and a fidelity dial. Restated with additional scope by the user mid-thread (2026-07-07): vertical-targeted design/content, eventual exemplar-parsing (reasoning about why best-in-class sites work, not copying them), and enrichment via tools/blog/news/infographics from wider-world reasoning.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-mission, #commercial-viability-is-not-the-same-as-a-business-site, #failure-modes-we-want-to-eliminate; running_notes_scheme_to_components(55).md#Um
- **relations:** fidelity dial; classifier as strategic brain; interactive content generation; orchestrator conventions; adoption-pipeline (exemplar-parsing kinship)
- **verify-later:** site_specs classification aspect; domain-research-classifier; agent-creation guidelines doc; exemplar-researcher plans elsewhere in docs

### BIZ-002 — finetuning.uk RAG platform product strategy & business plan
- **status:** aspirational
- **status-evidence:** BUSINESS_PLAN_finetuning_uk.md last touched 2026-04-21 ("Numbers are planning estimates"); FOCUS doc decisions dated 2026-04-21 with shipping ladder explicitly "Aspirational dates, not promises"; no later doc in any unit records a milestone reached.
- **what:** The plan to turn finetuning.uk into a paid product, pivoted several times across its documented history: from "self-service fine-tuning SaaS" to a RAG-over-your-docs platform for technical-adjacent SMEs (10-50 people, UK/EU) whose differentiator is automatic data curation — parse/classify/dedupe/quality-score/PII-scan/inconsistency-flag with a visible curation report ("competitors treat bad data as the customer's problem; we treat it as the product"); and from "concierge first, UI later" to UI-first ("build our own cockpit"). Solo-operator business plan: tiers Trial/£199/£499/£1,499/mo plus concierge fees (£750 audit → £15-30k bespoke); gross margin 57-78% per Growth customer; break-even ~5 Growth customers; 12-month target £9-12k/month, ~£100k year-1 revenue; content-led cold acquisition only. Reuse map: same Ollama/Unsloth/export/eval plumbing as the internal flywheel; entirely new: multi-tenancy, billing, UI, support, legal. Explicit not-to-ship list (multi-tenant fine-tuning SaaS, public API). Week-1 technical items named: tenant_id on knowledge_base enforced in rag_lookup/rag_index; auth stack choice.
- **sources:** BUSINESS_PLAN_finetuning_uk.md; BUSINESS_PLAN_finetuning_uk(1).md#1,#5,#7; FOCUS_finetuning_flywheel_and_service(13).md#5,#7,#8,#8a,#10,#11; FOCUS_finetuning_flywheel_and_service(25).md#5-13; FOCUS_finetuning_flywheel_and_service_v1.md#8,#10 (superseded shape)
- **relations:** internal flywheel infra reuse table; knowledge_base tenant_id plan; RAG knowledge_base; UI-first decision
- **verify-later:** state of finetuning.uk site; any tenant_id on knowledge_base; any finetuning.uk app code; site_specs for finetuning.uk

### BIZ-003 — Five-layer platform stack (L0 chassis → L1 idea engine → L2 idea.uk → L3 vertical tools → L4 tool-rich sites → L5 VM backend deploy)
- **status:** partial
- **status-evidence:** Consolidation map dated 2026-06-04: Layer 0 EXISTS, Layer 1 BUILT, Layers 2-3 IN PROGRESS, Layers 4-5 FUTURE ("Thunder adapter is the seed").
- **what:** A planning frame presenting the whole enterprise as one stack, each layer a customer of the one below: L0 the chassis builds sites (exists); L1 the idea engine decides what's worth building (built — method + internal CLI + idea.uk); L2 idea.uk sells the idea engine externally (in progress, first to go live); L3 recommended tools get built for real, chassis-native (in progress, e.g. SFI26 Diff Alerts); L4 the idea engine becomes a planning input so any domain gets a tool-rich site (future — "the original problem statement"); L5 automated backend deployment onto VMs closes the last gap, since today's pipeline only deploys static sites (future — Thunder adapter is the seed). Natural build order: prove L1 → ship L2 → build L3 once → generalise into L4 → grow L5 from Thunder.
- **sources:** idea.uk/CONSOLIDATION_where_it_all_fits.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md; running_notes(44).md ("Consolidation map written")
- **relations:** Layer-5 persistent-service wrapper; SFI26 Diff Alerts; chassis-native idea engine (Phase D); Thunder adapter (docs033/035); service-deployer pattern (= the L5 gap)
- **verify-later:** existence of any service-deployer agent; site_plan aspects carrying blocked/planned tool items; thunder-adapter actions; CONSOLIDATION_where_it_all_fits.md (live doc)

### BIZ-004 — Payable-differentiator / moat framework (asset × AI capability × paying audience)
- **status:** deployed
- **status-evidence:** Encoded as the core principle of the shipped ideation method (PLAN_idea_uk §3/§5, idea_uk_method_v0); applied across 8 domain runs (testruns v0/v2); independently restated in running_notes(44) and as a standalone strategy doc (docs025 PLAN_simple_paid_multidomain_chat §10).
- **what:** The doctrine used to filter which domains/products are worth building: the AI model is never the differentiator (everyone has the same models) — the defensible unit is a hard-to-reproduce asset (proprietary/paid data feed, an owned process/output, a well-built tool, a commercial partnership, or early-mover timing on a new AI capability) combined with AI, aimed at an audience that will pay for something a free model with a good prompt cannot do. Honest moat verdict reached via the framework's own self-application to idea.uk: its durable advantages are currency (a maintained capability watchlist beating models' self-knowledge), verification-with-evidence, and the build bridge (can build the idea, not just describe it) — a process/freshness/integration advantage, not a static asset ("effort + freshness + integration, sustained by maintenance"). Includes the brand-fit corollary (treat the product collection as separate from the domain portfolio; match deliberately) and a cross-domain pattern discovered via repeated runs: wherever the underlying product has high margin, the seller already gives expert support away free (Bloomberg/Refinitiv, Open Bionics, Robotiq) — an almost-automatic cut for "help-you-buy-X" candidates. Worked examples applying the method: websitedesign.com (package the site-spec/plan as a starter prompt for Bolt/Lovable), gaswholesalers.com (buy oil/gas data feeds), agritec.uk (partnership vouchers).
- **sources:** idea.uk/PLAN_idea_uk(3).md#5; idea.uk/idea_uk_method_v0(3).md; idea.uk/running_notes(63).md (2026-05-27 arc); running_notes(44).md ("The differentiator framework", "Moat analysis (idea.uk)"); docs025.../PLAN_simple_paid_multidomain_chat(1).md#10
- **relations:** ideation method; capability watchlist; five-layer stack; idea generation method; verticals designed
- **verify-later:** whether the capability watchlist exists as a recurring workflow anywhere in scheduled_tasks/agent_definitions

### BIZ-005 — Sale-readiness / separability discipline (idea.uk)
- **status:** deployed
- **status-evidence:** PLAN_idea_uk §2 rule "keep our asset list as input data, never built into the method"; RUNBOOK_idea_uk Notes: "the engine takes assets as data and the billing sits behind a provider interface, so idea.uk remains a separable unit".
- **what:** idea.uk is built to be sold as a working unit: business assets are always passed in as data (so the same engine serves internal domains and strangers), the set of workflows/actions it uses is kept identifiable and minimal, and billing sits behind a provider interface. The standalone Go service honours this (stdlib-only, file store, FakeProvider fallback).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/RUNBOOK_idea_uk(9).md; idea.uk/idea_uk_architecture_and_deployment(6).md#1
- **relations:** provider abstraction (payments); engine Go port; PAY-005 Pluggable billing provider abstraction
- **verify-later:** golang_files/engine.go input contract; billing.go Provider interface

### BIZ-006 — idea.uk as an instance of the paid multi-domain chat plan (day-pass lineage)
- **status:** superseded
- **status-evidence:** PLAN_idea_uk §2 "idea.uk is itself an instance of the paid multi-domain chat"; the built product ended up a report service, not a chat domain — the worker/paywall/day-pass reuse never happened in the shipped form.
- **what:** idea.uk originated as one configured domain of a planned "simple paid multi-domain chat" product (edge worker + paywall + day-pass), with the ideation method as its bound tool. The 2026-05-27 running-notes arc covers day-pass economics, per-domain monetisation by domain type, and serverless-edge vs central-nginx topology. The shipped idea.uk deliberately diverged: it is NOT edge-shaped (minutes-long background job → always-on service).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/running_notes(63).md (2026-05-27, "Pivot to simple paid multi-domain chat", "Topology note: idea.uk is NOT pure-static/edge")
- **relations:** PLAN_simple_paid_multidomain_chat.md; BIZ-004 payable-differentiator framework
- **verify-later:** whether the chat/day-pass product exists anywhere else in docs

### BIZ-007 — Voluntary pay and "free goes" rejected → free taster + paid report
- **status:** abandoned
- **status-evidence:** idea_uk_open_discussion §5 (2026-05-28): "probably not a good idea in this form… Drop voluntary pay and the multi-free-go idea. The taster is the better hook."
- **what:** Voluntary-pay ("pay if satisfied") and N-free-goes monetisation were analysed and rejected (abuse risk, no demand signal, trivially circumvented). Replaced by the pattern that shipped: a free, cheap (~£0.02) audience-check taster as proof-of-value plus a £29 full report with refund guarantee.
- **sources:** idea.uk/idea_uk_open_discussion.md#5; idea.uk/running_notes(63).md ("Day-pass collapses payment complexity", CHECKPOINT 2026-05-28 §4)
- **relations:** IDEA-007 Free audience-check taster endpoint; pricing decisions
- **verify-later:** n/a (business decision)

### BIZ-008 — Unit economics, pricing, and sourcing decisions (idea.uk, incl. self-hosting deferred)
- **status:** deployed
- **status-evidence:** running_notes SESSION DECISIONS LOG 2026-06-11; open_discussion §§1-2,6 with verified May-2026 pricing; £29 live and proven with a real card 2026-06-14.
- **what:** Per-run engine cost ~£0.40-0.60 (verify step dominates; optimisable to ~£0.20-0.30 via Haiku scoring + prompt caching); Stripe UK fees 1.5%+£0.20, break-even ~£0.72, worst-case refund cost ~£1.43; price settled at pay-per-idea, cost-plus, £29 flat (not B2B SaaS for the ideation product itself). Self-hosted LLMs analysed and deferred ("a 2027 decision, not a 2026 one") — commercial frontier models win at this volume, and open-weight models are weakest exactly at the cut step's ruthlessness.
- **sources:** idea.uk/idea_uk_open_discussion.md#1-2,6; idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md (pricing checkpoint)
- **relations:** PAY-001 Stripe webhook-as-truth pattern; engine model line-up
- **verify-later:** REPORT_PRICE_GBP env on the live box

### BIZ-009 — Building-and-hosting as a service via chat (recursive satellite platform)
- **status:** aspirational
- **status-evidence:** "Recorded because it sharply reframes the satellite ... (Discussion artefact; revisit as it firms up.)" — recorded twice, independently, in two different document trees (tools/tool_widget_clobber and docs025), each still marked as unresolved discussion.
- **what:** A worked example where a site's own chatbot becomes the intake + orchestration front-end to the whole build platform offered as a service: a prospective customer types a domain + spec into an existing site's (or design.co.uk's) chat box; conversational briefing (a briefing-agent interview replacing the static form) triggers the full build pipeline on the satellite (not core) to produce a new hosted site — itself shipped with its own chat box (explicit recursion). Requires the full chassis on the satellite (rules out lighter isolation options). Reframes the satellite from an isolation nicety into a required second, customer-facing instance of the whole platform, and surfaces net-new concerns: cost/abuse exposure from anonymous builds, need for accounts/billing/quota gating, feeding reusable building blocks one-directionally from core.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#12; docs025.../PLAN_isolated_chat_environment(4).md#12
- **relations:** Isolated chat/satellite architecture (Y-copy); BIZ-014 Operator-vs-vendor business model fork; briefing-agent/intake orchestrator; PAY-003 Entitlement gate architecture
- **verify-later:** conversational briefing-agent reusing 018_briefing_questionnaire/002_intake_orchestrator is proposed but not confirmed to exist; satellite intake orchestrator

### BIZ-010 — Audience-tuned elevator pitch variants (V1-V4 method)
- **status:** deployed
- **status-evidence:** Four fully-written 25-95 word variants plus 10-second openers, each annotated with likely follow-up question and delivery notes — a finished deliverable, not a plan.
- **what:** A pitch-crafting technique producing four distinct ~25-35-second verbal pitches for the same underlying platform, each tuned to a different audience (technical-peer contrarian opener; commercial/investor asset-framing; mixed-audience concrete-first default; written/cold-context compressed version), plus even-shorter 10-second openers. Explicitly engineers what's deliberately left out to keep each version deliverable in one breath.
- **sources:** pitch/002_substrate_framing_elevator_pitches.md#V1-V4, #Notes on delivery
- **relations:** BIZ-011 Substrate-vs-application pitch framing; BIZ-012 Fractal agent architecture claim
- **verify-later:** n/a (pitch-writing artefact, not code)

### BIZ-011 — Substrate-vs-application pitch framing
- **status:** aspirational
- **status-evidence:** Framed explicitly as "Experimental alternative framing... To compare against, not necessarily replace, the original pitch doc".
- **what:** Repositions the same system from "I built a website builder" (a product) to "I built a domain-agnostic distributed agent orchestration substrate, and the website builder is one demonstration of it among five" (an asset/infrastructure claim). Chosen per audience: website framing for commercial/marketing-tech roles, substrate framing for AI-infrastructure roles. Comes with an explicit "words to use / avoid" list.
- **sources:** pitch/framework_pitch_substrate_framing.md#§1,§6,§9; pitch/002_substrate_framing_elevator_pitches.md#final notes
- **relations:** BIZ-012 Fractal agent architecture claim; BIZ-013 Honest-delta disclosure discipline; BIZ-010 Audience-tuned elevator pitch variants
- **verify-later:** n/a

### BIZ-012 — Fractal agent architecture claim (self-similar recursive orchestration)
- **status:** deployed
- **status-evidence:** Backed by a traced real production call chain "seven levels deep with identical code paths at each level" (intake-orchestrator → site-work-orchestrator → build-dispatch-loop → page-build-handler → page-content-writer → research-agent → web-fetcher).
- **what:** The claim that every agent in the system is itself an orchestrator using identical primitives (spawn/call/claim/complete, same Kafka topic conventions, same orchestration_states shape) at every depth of the spawn tree, with no architectural distinction between a "top-level orchestrator" and a "leaf specialist." Framed as the single most defensible/highest-risk word in the pitch, contrasted against single-process Python frameworks (LangGraph/CrewAI/AutoGen).
- **sources:** pitch/framework_pitch_substrate_framing.md#§2; pitch/002_substrate_framing_elevator_pitches.md#final notes; pitch/framework_pitch_reference.md#§3.1
- **relations:** BIZ-011 Substrate-vs-application pitch framing; Multi-cluster agent dispatch contract
- **verify-later:** orchestration_states.parent_orchestration_id chain query; agent_spawn_history

### BIZ-013 — Honest-delta disclosure discipline (built vs admitted-not-built table)
- **status:** convention
- **status-evidence:** A fully filled-in comparison table ("Honest Delta — What's Built vs What the Architecture Admits") described as "the load-bearing piece... what protects you from over-claiming".
- **stage2-verified (2026-07-14):** deployed → convention — Honest-delta disclosure is a pitch-writing/documentation discipline (a comparison table in a doc), not a code/infra artifact; stage1 itself lists verify-later: n/a.
- **what:** A pitch-integrity practice: pair every ambitious architectural claim with an explicit, honest ledger of what's actually proven in production versus merely structurally possible versus not built at all. Extends into a parallel "Honest Weak Points" section (solo development, documentation drift, schema drift, race conditions, incomplete migrations, no formal test suite) each with a pre-written honest-framing answer.
- **sources:** pitch/framework_pitch_substrate_framing.md#§7,§6.4; pitch/framework_pitch_reference.md#§10
- **relations:** BIZ-011 Substrate-vs-application pitch framing; BIZ-012 Fractal agent architecture claim
- **verify-later:** n/a

### BIZ-014 — Operator-vs-vendor business model fork / SaaS commercial model with entitlement seams
- **status:** aspirational
- **status-evidence:** "Resolved direction" language for the model itself, in two independently-arrived-at documents ("that resolves the strategic fork cleanly (operator-primary at scale, vendor-optional per domain)" — stripe/001commentary.md§13; "Design the seam now... build billing depth later" — PLAN_isolated_chat_environment§13); the concrete implementation is explicitly deferred in both.
- **what:** Identifies that "operate thousands of domains yourself" and "sell the whole framework/instance to a buyer" pull toward opposite technical choices (centralised multi-tenant efficiency vs. clean separable sellable units), resolved as operator-primary at scale with per-domain vendor-optionality. The key structural insight: the unit of blast-radius isolation (the satellite/cluster) is distinct from the unit of separability-for-sale (the domain), partitioned within a satellite's stores via site_id/network_id/domain — operating thousands of domains does not require thousands of clusters. Recommends honouring five seams cheaply now because retrofitting separability later is "a forensic untangling": ownership on site rows (reusing the existing clients→networks→sites hierarchy), an entitlement gate at both build-submission and maintenance-run (never calling Stripe directly — always through a pluggable billing-adapter interface), credential parameterisation everywhere, and a build-tier/cost-profile flag (saas_cheap vs portfolio) driving cheaper model/batching choices so low-price builds retain margin.
- **sources:** stripe/001commentary.md#§13, #final turn; tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13; docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md §13
- **relations:** Isolated chat/satellite architecture (Y-copy); Ownership hierarchy reuse for entitlement scoping; PAY-003 Entitlement gate architecture; BIZ-009 Building-and-hosting as a service via chat
- **verify-later:** every domain-scoped table's site_id/network_id discipline; export-and-re-point procedure for per-domain sale; whether an owner_id/entitlement layer or billing adapter exists anywhere in the schema/codebase today

### BIZ-015 — domain-strategist (strategy vs architecture separation)
- **status:** deployed
- **status-evidence:** 060 definition with explicit responsibility statement; work item chain needs_strategy → needs_briefing.
- **what:** Handler for needs_strategy items. Determines the strategy for a domain — canonical site_type, revenue model, content strategy, page_type recommendations, tone/positioning — and writes site_specs aspect "strategy". Explicit contract: does NOT design page architecture; "The planner has final say... may agree, adjust, or override"; does NOT overwrite the researcher's "classification" aspect.
- **sources:** 060_domain_strategist.sql
- **relations:** build-site-planner reads strategy; domain-research-classifier upstream; BIZ-023 Domain content strategy framework
- **verify-later:** strategy aspect consumption in plan_site prompt

### BIZ-016 — Portfolio/use-case spec seeds (ai-agent-orchestration.com)
- **status:** deployed
- **status-evidence:** 100 INSERTs site_specs 'portfolio' aspect with five dated case studies claiming operational metrics ("Six production sites deployed and self-maintaining... under 4 hours" domain-to-live).
- **what:** Marketing-facing data seed whose case studies double as a platform capability inventory circa file-100: autonomous multi-site pipeline (30+ agents), tool generation + cross-linking, vet data platform, news aggregation with credibility scoring, and the orchestration layer itself (Kafka/Postgres/K8s, hot-swappable SQL workflow definitions, fuel budgets). Useful as documentary status evidence for many other concepts, not ground truth.
- **sources:** 100_portfolio_use_cases_etc.sql
- **relations:** nearly every pipeline concept above; BIZ-021 Early portfolio inventory
- **verify-later:** claims vs stage-2 code/DB verification

### BIZ-017 — AI persona team and departments marketing model
- **status:** deployed
- **status-evidence:** Applied site_specs updates (075a/076) injecting team and departments JSONB for ai-agent-orchestration.com, finetuning.uk, leopardessconsulting.co.uk, with audience-tuned copy per site.
- **what:** The platform presents itself through named AI managing-agent personas — Archivist (Research), Sentinel (Quality), Quartermaster (Operations) — alongside the human principal, plus an 8-department / 70+ agent structure with per-department agent counts and capability summaries. Stored as identity-spec data consumed by the content writer for team/departments sections; departments-grid component renders it as the leadership-team alternative.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#076
- **relations:** site_specs identity; departments-grid component; pitch/business docs
- **verify-later:** rendered team/departments sections on the three sites

### BIZ-018 — EBORG — Evidence-Based Organisational Planning (venture concept)
- **status:** unknown
- **status-evidence:** Appears in two independent doc eras: as the demo business for HITL content approval (0103, full pitch text in the trigger script) and later, separately, branded "For EBORG" in start_hitl_workflow.sh with the same pitch paragraph, described as "the concrete descendant of 016's Organizational OS idea"; no doc in either era confirms it became a real product.
- **what:** A business idea used as the HITL demo client: map every role/responsibility/objective in an organisation and pair each with a framework of AI agents that gather research, assess options, and provide evidence-based reasoning — "a human-centered, continuously learning organisation." Also spawned the simple-content-writer-with-approval agent. Thematically echoes the later council-of-experts idea in docs026 stage 3.
- **sources:** docs002_hitl_parallel/README.0103.hitl_start_message.md; docs002_hitl_parallel/README.0102.hitl_agent_definitions; docs/humanintheloop/start_hitl_workflow.sh; docs/humanintheloop/hitl_agent_definition.sql (header comment); docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** HITL content approval group (content-approval-hitl); cross-domain intelligence network
- **verify-later:** any EBORG references in business/vonc docs outside this cluster

### BIZ-019 — WordPress export/handoff idea (XML export, plugin shortcodes, subscription plugin) — abandoned
- **status:** abandoned
- **status-evidence:** Detailed plan in docs004/004-005 (WordPress Formatter agent, wordpress-export.xml, SQL brand injection) followed much later by a fuller docs/architecture/008-009 design (WXR export + subscription plugin) that is itself immediately killed by competitive analysis ("You're not unique in 'AI builds WordPress sites'. That market is saturated... Only add WordPress if they're begging for it"); never mentioned again in either lineage.
- **what:** Two related conceptions of the same exit idea, roughly a generation apart: (1) an early client-handoff strategy — transpile a generated site into a single WordPress import file so a client's developer gets a standard maintainable WP site in minutes, complex components as plugin shortcodes, brand colours/fonts injected via one SQL file, part of a broader survey of exit routes (traditional CMS vs SaaS builders vs headless/Jamstack); (2) a later, more enthusiastic design — an agent converting generated HTML sites into installable WordPress themes + WXR content exports, paired with a WP plugin subscribing to the platform for auto-published fresh content (recurring revenue) — explicitly deconstructed and killed in the very next document, which redirected differentiation toward "sites that update themselves" / continuous content ecosystems (the path the platform did pursue via news feed / content pipelines).
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.005.frontend_frameworks.md; docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md#wordpress-export-agent-design; docs/architecture/009-wordpress-discussion#the-hard-truth
- **relations:** business-strategy (client/exit strategy); deployment-github (the retained path); HTML-first delivery; living-content differentiation
- **verify-later:** none — abandoned idea registry

### BIZ-020 — Two-tier commercialisation model (sell output → sell setup)
- **status:** aspirational
- **status-evidence:** docs016/004 dated 2026-03-03 "Working notes for strategic direction": three-tier model trimmed to two ("drop the domain selling tier"); practical next step "produce 10 websites in a niche... validate with real money".
- **what:** Frank commercial assessment: framework differentiators are real infrastructure (K8s/Kafka/Postgres, data-driven workflows, multi-cluster, chassis pattern) but lack docs/community; revenue paths ranked (website service most mature; SEO content; document processing needs domain partner; framework sales longest); recommended model — run the service in a chosen niche to accumulate live outputs, then sell the whole setup as a business-in-a-box (£5-25K) once 20-50 outputs prove it, repeating per product; canine project reframed as portfolio/demo spend. Open decisions: niche, sellable quality, solo vs partner, runway.
- **sources:** docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** BIZ-021 early portfolio inventory; canine biology demotion
- **verify-later:** n/a (strategy)

### BIZ-021 — Early portfolio inventory (honest capability notes)
- **status:** unknown
- **status-evidence:** docs018/008b raw notes: "None of our sites get leads at the moment so we can't say they do... we'd rather sell the sites achievement at the moment"; lists leopardessconsulting, vetcomparison.uk, wykefarm.co.uk, mortgagecalculator, website-design.com.
- **what:** A candid snapshot of what existed circa Feb 2026: leopardessconsulting built and evolving over days; a veterinary price-comparison site plus vet search/scrape/data-collection service; wykefarm.co.uk farm site (biodiversity content); a quickly-built but rough mortgage calculator site; website-design.com with functional tools (paste boards, mind maps, colour tools) but poor polish; framework scaling claim "several thousand agents". Useful ground truth for verifying which case-study sites actually functioned.
- **sources:** docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** site-case-studies (leopardess, vet, wyke); vet-med-pricing; dynamic-applications (website-design.com tools); BIZ-016 Portfolio/use-case spec seeds
- **verify-later:** sites table rows for each named domain

### BIZ-022 — Verticals designed (revenue models + knowledge clusters)
- **status:** aspirational
- **status-evidence:** Revenue projection tables labelled "months 12-18" with market data "verified through research"; no live-site revenue claimed.
- **what:** Five verticals worked out with specific knowledge clusters, source lists, page-type libraries, monetisation, and 24-month revenue projections: veterinary/vetcomparison.uk (insurance affiliate £15-35 + listings, £1,960-7,875/mo), energy/gaswholesalers.com (qualified leads £30-60, £1,250-5,350), finance_mortgage/mortgagecalculator.co.uk (broker leads £50-150, £16,500-44,000 — highest value), seasonal_gifts/xmaspresents.com (affiliate 3-17%), plus a "sell the domain not develop" premium pathway (design.co.uk £20-100k).
- **sources:** docs021.../020_vertical_cluster_architecture.md#3; docs021.../025_session_handoff_vertical_architecture.md#verticals-designed; docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture (VKA-001); BIZ-023 Domain content strategy framework; BIZ-025 Content-site valuation model
- **verify-later:** vertical_registry monetisation_config; any live vetcomparison/gaswholesalers/mortgagecalculator sites

### BIZ-023 — Domain content strategy framework (15-question)
- **status:** aspirational
- **status-evidence:** "For the content generation pipeline, the 15-question framework should feed into the briefing/research phase" — prescriptive/should, not implemented.
- **what:** A systematic three-layer, 15-question methodology for deciding what content a domain needs to compete: Layer 1 (who visits, intent, satisfaction, money flow), Layer 2 (competitor pages, buying journey, real questions, bookmarkable hook), Layer 3 (best page on the topic, original element, format, next action). Worked examples for gaswholesalers.com and vetcomparison.uk with verified lead/affiliate rates. Questions 5-7 require real competitive research.
- **sources:** docs022.../001_domain_content_strategy_framework.md; docs022.../002_domain_content_strategy_framework_v2.md
- **relations:** domain-strategist prompt (BIZ-015); deep research domain authority (BIZ-024); site classifier
- **verify-later:** domain-strategist agent prompt; briefing/research phase incorporating the framework

### BIZ-024 — Deep research domain authority strategy
- **status:** aspirational
- **status-evidence:** "The multi-cluster knowledge base approach lets you build that deep knowledge layer for any domain" — strategy doc, canine project cited as proof-of-concept not production.
- **what:** The thesis that content wins on E-E-A-T by synthesising primary/authoritative sources (BSAVA, Ofgem, PRA/FCA, swap-rate data) into knowledge consumers can't easily find, rather than rephrasing published synthesis. A repeatable 6-step pipeline (niche mapping → primary-source identification → multi-cluster KB construction → gap identification → content architecture → generation from KB) creates a defensible moat: depth consistency, cross-cluster synthesis, and update efficiency competitors can't copy by rewriting one article.
- **sources:** docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture (VKA-001); domain content strategy framework (BIZ-023); canine biology KB
- **verify-later:** research-agent primary-source handling; knowledge_base source_authority weighting

### BIZ-025 — Content-site valuation model (24-32x)
- **status:** aspirational
- **status-evidence:** "using current market multiples of 24-32x monthly profit (Empire Flippers averaging 24x, premium 30-35x)" — used throughout as a projection basis.
- **what:** The valuation basis underpinning the domain portfolio strategy: content/affiliate sites sell at ~24-32x monthly profit, so a £1,500-3,000/mo site is worth ~£36k-96k. Combined with verified per-niche lead/affiliate rates to project each domain's asset value and justify the knowledge-base investment. The portfolio is framed as the testing ground toward a £25k+ annual revenue target and a two-tier service→pipeline-sale model.
- **sources:** docs022.../002_domain_content_strategy_framework_v2.md#monetisation; docs021.../025_session_handoff_vertical_architecture.md#market-data-verified
- **relations:** BIZ-022 Verticals designed; BIZ-020 Two-tier commercialisation model
- **verify-later:** n/a (business assumption)

### BIZ-026 — Data-sovereignty / pilot-first / startup fast-start commercial positioning
- **status:** aspirational
- **status-evidence:** RUNBOOK H6/H8/H9 all "resolved 2026-07-10" as drafted positioning, each explicitly "Not yet done for a client".
- **what:** Three owner-confirmed engagement angles drafted into specs/portfolio.json use_cases with honest labelling: (A8) data-sovereignty as a capability built with a client during a scoped engagement, never a standing guarantee; (H6) pilot-first engagement ladder (bounded fixed-price pilot → licence/day-rate/retainer decided by what the pilot reveals); (H9) startups building agent products start from the platform's already-solved operational layer (state, retries, HITL, no-redeploy workflow changes). Plus register-reconciliation generalisation ("your list vs an authoritative register" as the general shape of the Companies House work).
- **sources:** docs/leopardessconsulting/specs/portfolio.json#use_cases; docs/leopardessconsulting/RUNBOOK.md#H6-H9; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6,#Turn-7
- **relations:** per-step model routing; no-tenant-isolation; claim-evidence audit rule
- **verify-later:** site_specs aspect 'portfolio' current row

### BIZ-027 — UK-sovereign stack exploration (deferred)
- **status:** aspirational
- **status-evidence:** RUNBOOK Reference: "explicitly deferred to a separate chat by the owner, 2026-07-10 … Do not start this unprompted."
- **what:** Future exercise: a fully UK-hosted compute+storage+model stack. Baseline facts captured so the future thread doesn't re-derive them: compute Rackspace UK; storage Backblaze us-east-005 (US); Anthropic and Google models US; self-hosted path sits wherever the cluster is.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Reference; docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P6
- **relations:** BIZ-026 data-sovereignty positioning; storage-architecture
- **verify-later:** memory uk-sovereign-stack-exploration

### BIZ-028 — Domain value maximisation pipeline (domain flipping)
- **status:** abandoned
- **status-evidence:** 010/011/015 lay out the strategy against the user's real domain portfolio (collateralfinancing.com, holidaytime.com, websitedesign.com...) with 48-hour development timelines; no later doc pursues domain flipping — the platform pivoted to operating its own sites.
- **what:** Use the agent platform to develop parked domains into sites with content, traffic and revenue to multiply sale value (naked $500 → revenue-bearing $10k+): domain classification (brandable/exact-match/local/product), tiered portfolio treatment, 48h batch development, monetisation setup (leads/affiliate/ads), and "self-selling" footers that market the build service from every developed domain.
- **sources:** docs/architecture/010-domain-value-maximisation.md; docs/architecture/011-example-domains; docs/architecture/015-underserved-niche.md#your-domain-portfolio-is-your-marketplace
- **relations:** deep-research domain insight agent; BIZ-029 underserved-niche strategy; site-case-studies
- **verify-later:** n/a

### BIZ-029 — Underserved-niche and vertical showcase strategy
- **status:** abandoned
- **status-evidence:** 015/016 propose niches (compliance docs, local-business packages, academic assistants, affiliate content) and per-industry showcase domains with a workflow marketplace ("DIY $500 / DFY $200mo / White Label $2000mo"); positioning discussion only, no implementation trail.
- **what:** Rather than competing with Temporal/LangChain/Zapier broadly, own narrow verticals where multi-agent coordination wins: each showcase domain demos an industry solution (legal docs, restaurant launch, real-estate listings) and funnels to purchasable workflows. Includes the pricing-tier and "Business-in-a-Box" (site + content pipeline + email + social) framings, and the investor-demo positioning of the framework as the star with swappable use cases.
- **sources:** docs/architecture/015-underserved-niche.md; docs/architecture/016-competitive-advantge.md#who-actually-pays-for-ai-sites; docs/architecture/012-investors.md#the-portfolio-approach
- **relations:** BIZ-028 domain value maximisation; EBORG organizational OS (BIZ-018)
- **verify-later:** n/a

### BIZ-030 — AI-native orchestration positioning (vs Temporal/Airflow)
- **status:** abandoned
- **status-evidence:** 012/014 are interview/investor argumentation ("You could build this on Temporal, but it would be like using Kubernetes to run a single container"); the accompanying Temporal/Airflow adapter agents were never built.
- **what:** The articulated build-vs-buy rationale for the platform: AI-specific needs (dynamic JSON workflows without deployment, token/fuel tracking, prompt management, AI-failure handling, multi-tenant agent isolation, workflows spawning from AI decisions) justify a purpose-built orchestrator. Includes proposed Temporal-adapter and Airflow-adapter agents to bridge into enterprise workflow estates as a migration path ("we don't replace your existing systems, we enhance them").
- **sources:** docs/architecture/012-investors.md#better-answer; docs/architecture/014-Temporal-Airflow-adapters.md
- **relations:** adapters (current adapter guide is a different, real lineage); distributed embedded orchestration
- **verify-later:** confirm no temporal/airflow adapter code exists
