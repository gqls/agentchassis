# Cluster: business-strategy-products
Categories included: business-strategy, new:idea-product, new:business-intelligence-platform, new:conversion-playbooks, new:portfolio-evolution, new:vertical-knowledge-architecture, payments, new:legal-and-compliance, new:legal-liability, new:marketing, new:seo, new:affiliate-and-products, new:affiliate-commerce


<!-- SOURCE: U01_docs024_numbered_core.md -->
### Platform mission and the single unified pipeline
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 028 (living document, rev 2026-04-22) is the stated check-yourself-against anchor
- **what:** Mission: given domain names in any state, produce the best possible website end-to-end with minimal human input, "best" = most useful to probable visitors measured by real engagement AND best revenue via whatever model genuinely fits. One pipeline for blank/adopted/missioned/replication domains — differing only in input material and the fidelity dial. Revenue model shapes the site (default-to-brochure/consultancy is a named failure mode); classifier decides the commercial shape.
- **sources:** 028#The mission, #Commercial viability
- **relations:** fidelity dial; classifier as strategic brain
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### finetuning.uk self-service product strategy
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Not started. Questions to answer before scoping"; 12 decisions dated 2026-04-21; shipping ladder "Aspirational dates, not promises"
- **what:** finetuning.uk as both a credible knowledge site and a revenue product: flagship = RAG platform with data curation as a first-class visible feature (parse/classify/dedupe/quality-score/PII-scan/inconsistency-flag pipeline reusing the framework), concierge-onboarded then self-serve; tiers from £199/mo platform to £15-30k bespoke; target user technical-adjacent SMEs; UI-first build as own operational cockpit; explicit not-to-ship list (multi-tenant fine-tuning SaaS, public API). Differentiation from positioning (UK residency, opinionated simplicity, self-improvement loop) not engineering.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#5, #7, #8, #8a, #10, #11
- **relations:** internal flywheel infra reuse table (#6); knowledge_base tenant_id plan
- **verify-later:** state of finetuning.uk site; any tenant_id on knowledge_base

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Platform mission restatement (plan and build websites from a domain)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Checkpoint (Um), restated by the user 2026-07-07: "intelligently PLAN and BUILD multipage websites from a domain name — targeted design/content per vertical, eventually parsing best-in-class exemplars (reasoning from why they work, not copying), adding tools/blog/news/infographics from wider-world reasoning."
- **what:** The system's purpose as the user states it mid-thread when re-anchoring scope: an agent system that plans and builds whole multipage sites from a domain name, with vertical-targeted design and content, future exemplar-parsing (reasoning about why best-in-class sites work rather than copying), and enrichment via tools, blog, news and infographics; supported by close agent/message logging, an agent-creation guidelines doc, and distinct low-overlap agent responsibilities with sub-agents for research-before-content.
- **sources:** running_notes_scheme_to_components(55).md#Um
- **relations:** orchestrator conventions; research-agents; adoption-pipeline (exemplar parsing kinship).
- **verify-later:** agent-creation guidelines doc; exemplar-researcher plans elsewhere in docs.

<!-- SOURCE: U04_idea_uk.md -->
### Five-layer platform stack (chassis → idea engine → idea.uk → vertical tools → tool-rich sites → VM backend deploy)
- **category:** business-strategy
- **status-signal:** partial
- **status-evidence:** "Where it all fits" map dated 2026-06-04: Layer 0 EXISTS, Layer 1 BUILT, Layers 2–3 IN PROGRESS, Layers 4–5 FUTURE ("Thunder adapter is the seed").
- **what:** A consolidation model presenting the whole enterprise as one stack: the chassis builds sites (L0); the idea engine decides what's worth building (L1); idea.uk sells that externally (L2); recommended tools get built for real, chassis-native (L3); the engine becomes a planning input so any domain gets a tool-rich site (L4 — "the original problem statement"); and automated backend deployment onto VMs closes the last gap (L5). Each layer is a customer of the one below.
- **sources:** idea.uk/CONSOLIDATION_where_it_all_fits.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md
- **relations:** Layer-5 persistent-service wrapper; SFI26 Diff Alerts; chassis-native idea engine (Phase D); Thunder adapter (docs033/035).
- **verify-later:** existence of any service-deployer agent; site_plan aspects carrying blocked/planned tool items; thunder-adapter actions.

<!-- SOURCE: U04_idea_uk.md -->
### Differentiator framework — payable idea = hard-to-reproduce asset × current AI capability
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** PLAN_idea_uk §3/§5 (framework in use); method encodes it as the core principle; testruns v0/v2 applied it across 8 domain runs.
- **what:** The AI model is never the differentiator (everyone has the same models); the defensible unit is an asset × capability aimed at an audience that will pay, doing something a free model with a good prompt cannot. Honest moat verdict: the durable advantages are **currency** (a maintained capability watchlist beating models' self-knowledge), **verification with evidence**, and the **build bridge** (we can build the idea, not just describe it) — a process/freshness/integration advantage, not a static asset. Includes the brand-fit corollary (treat the product collection as separate from the domain portfolio; match deliberately).
- **sources:** idea.uk/PLAN_idea_uk(3).md#5; idea.uk/idea_uk_method_v0(3).md; idea.uk/running_notes(63).md (2026-05-27 arc)
- **relations:** ideation method; capability watchlist; five-layer stack; paid multi-domain chat plan (§10 of that doc).
- **verify-later:** whether the capability watchlist exists as a recurring workflow anywhere in scheduled_tasks/agent_definitions.

<!-- SOURCE: U04_idea_uk.md -->
### Sale-readiness / separability discipline (assets as data, minimal identifiable dependency set)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** PLAN_idea_uk §2 rule "keep our asset list as input data, never built into the method"; RUNBOOK_idea_uk Notes: "the engine takes assets as data and the billing sits behind a provider interface, so idea.uk remains a separable unit".
- **what:** idea.uk is built to be sold as a working unit: business assets are always passed in as data (so the same engine serves internal domains and strangers), the set of workflows/actions it uses is kept identifiable and minimal, and billing sits behind a provider interface. The standalone Go service honours this (stdlib-only, file store, FakeProvider fallback).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/RUNBOOK_idea_uk(9).md; idea.uk/idea_uk_architecture_and_deployment(6).md#1
- **relations:** provider abstraction (payments); engine Go port.
- **verify-later:** golang_files/engine.go input contract; billing.go Provider interface.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk as an instance of the paid multi-domain chat plan (day-pass lineage)
- **category:** business-strategy
- **status-signal:** superseded
- **status-evidence:** PLAN_idea_uk §2 "idea.uk is itself an instance of the paid multi-domain chat"; the built product ended up a report service, not a chat domain — the worker/paywall/day-pass reuse never happened in the shipped form.
- **what:** idea.uk originated as one configured domain of a planned "simple paid multi-domain chat" product (edge worker + paywall + day-pass), with the ideation method as its bound tool. The 2026-05-27 running-notes arc covers day-pass economics, per-domain monetisation by domain type, and serverless-edge vs central-nginx topology. The shipped idea.uk deliberately diverged: it is NOT edge-shaped (minutes-long background job → always-on service).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/running_notes(63).md (2026-05-27, "Pivot to simple paid multi-domain chat", "Topology note: idea.uk is NOT pure-static/edge")
- **relations:** PLAN_simple_paid_multidomain_chat.md (outside this unit); hosting split concept below.
- **verify-later:** whether the chat/day-pass product exists anywhere else in docs/ (other units).

<!-- SOURCE: U04_idea_uk.md -->
### Voluntary pay and "free goes" rejected → free taster + paid report
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** idea_uk_open_discussion §5 (2026-05-28): "probably not a good idea in this form… Drop voluntary pay and the multi-free-go idea. The taster is the better hook."
- **what:** Voluntary-pay ("pay if satisfied") and N-free-goes monetisation were analysed and rejected (abuse risk, no demand signal, trivially circumvented). Replaced by the pattern that shipped: a free, cheap (~£0.02) audience-check taster as proof-of-value plus a £29 full report with refund guarantee.
- **sources:** idea.uk/idea_uk_open_discussion.md#5; idea.uk/running_notes(63).md ("Day-pass collapses payment complexity", CHECKPOINT 2026-05-28 §4)
- **relations:** audience-check taster endpoint; pricing decisions.
- **verify-later:** n/a (business decision).

<!-- SOURCE: U04_idea_uk.md -->
### Unit economics, pricing, and sourcing decisions (incl. self-hosting deferred)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** running_notes SESSION DECISIONS LOG 2026-06-11; open_discussion §§1–2, 6 with verified May-2026 pricing; £29 live and proven with a real card 2026-06-14.
- **what:** Per-run engine cost ~£0.40–0.60 (verify step dominates; optimisable to ~£0.20–0.30 via Haiku scoring + prompt caching); Stripe UK fees 1.5%+£0.20, break-even ~£0.72, worst-case refund cost ~£1.43; price settled at **pay-per-idea, cost-plus, £29 flat** (not B2B SaaS for the ideation product itself). Self-hosted LLMs analysed and deferred ("a 2027 decision, not a 2026 one") — commercial frontier models win at this volume, and open-weight models are weakest exactly at the cut step's ruthlessness.
- **sources:** idea.uk/idea_uk_open_discussion.md#1-2,6; idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md (pricing checkpoint)
- **relations:** Stripe webhook pattern; engine model line-up.
- **verify-later:** REPORT_PRICE_GBP env on the live box.

<!-- SOURCE: U06_finetuning.md -->
### finetuning.uk product strategy (RAG platform flagship, data curation as the product)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN "Last touched: 2026-04-21… Numbers are planning estimates"; FOCUS(25) §7 decisions 11-12; no later doc records any milestone reached.
- **what:** The external-product thesis: finetuning.uk becomes a RAG platform for technical-adjacent SMEs (10-50 people, knowledge-intensive, UK/EU) whose differentiator is automatic data curation (parse/classify/dedup/quality-score/PII-scan/inconsistency-flag with a visible curation report) — "competitors treat bad data as the customer's problem; we treat it as the product." RAG chosen ahead of text/image LoRA tiers (users arrive with docs, not training pairs; the infra is built). Self-service *fine-tuning* SaaS is explicitly deferred/not-shipped. Reuse map: same Ollama/Unsloth/export/eval plumbing; entirely new: multi-tenancy, billing, UI, support, legal. Week-1 technical items named: tenant_id on knowledge_base enforced in rag_lookup/rag_index; auth stack choice.
- **sources:** BUSINESS_PLAN_finetuning_uk.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#5-13
- **relations:** knowledge_base RAG; internal flywheel (shared plumbing); UI-first decision
- **verify-later:** tenant_id on knowledge_base; any finetuning.uk app code; site_specs for finetuning.uk

<!-- SOURCE: U06_finetuning.md -->
### finetuning.uk business plan (pricing, unit economics, milestones)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN changelog: "2026-04-21 — Initial draft… refine as data arrives"; no subsequent updates in this tree.
- **what:** Solo-operator plan: tiers Trial/£199/£499/£1,499/Enterprise plus concierge fees (£750 audit → £15-30k bespoke); gross margin 57-78% per Growth customer; break-even ~5 Growth customers; 12-month target £9-12k/month and ~£100k year-1 revenue; content-led cold acquisition only (the framework as content engine is the claimed structural moat); interim gigs capped at 50% of time; assumption list with explicit 60-day tests; milestone/decision gates at months 1/3/6/12. Superseded staging inside the family: v1's "concierge first, UI later" three-tier structure was replaced by the UI-first "build our own cockpit" revision (2026-04-21 fourth pass).
- **sources:** BUSINESS_PLAN_finetuning_uk.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service_v1.md#8,#10 (superseded shape); working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#8,#10
- **relations:** product strategy; vonc/business-strategy docs elsewhere
- **verify-later:** nothing technical; check for any later business docs superseding this

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Building-and-hosting as a service via chat (build-as-a-service reframing)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Recorded because it sharply reframes the satellite ... (Discussion artefact; revisit as it firms up.)" (PLAN_isolated_chat_environment(5).md §12)
- **what:** A worked example where a site's own chatbot becomes the intake + orchestration front-end to the whole build platform offered as a service: a prospective customer types a domain + spec into an existing site's chat box, and the full build pipeline runs on the satellite (not core) to produce a new hosted site — itself shipped with its own chat box (explicit recursion). Reframes the satellite from an isolation nicety into a required second, customer-facing instance of the whole platform, and surfaces net-new concerns: cost/abuse exposure from anonymous builds, need for accounts/billing/quota gating.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#12
- **relations:** Isolated chat/satellite architecture (Y-copy); Operator-vs-vendor business model fork; Conversational build-intake via briefing-agent chat; Agent-to-adapter maturation path
- **verify-later:** conversational `briefing-agent` reusing `018_briefing_questionnaire`/`002_intake_orchestrator` is proposed but not confirmed to exist

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Audience-tuned elevator pitch variants (V1-V4 method)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Four fully-written 25-95 word variants plus 10-second openers, each annotated with likely follow-up question and delivery notes — a finished deliverable, not a plan
- **what:** A pitch-crafting technique producing four distinct ~25-35-second verbal pitches for the same underlying platform, each tuned to a different audience (technical-peer contrarian opener; commercial/investor asset-framing; mixed-audience concrete-first default; written/cold-context compressed version), plus even-shorter 10-second openers. Explicitly engineers what's deliberately left out to keep each version deliverable in one breath.
- **sources:** pitch/002_substrate_framing_elevator_pitches.md#V1-V4, pitch/002_substrate_framing_elevator_pitches.md#Notes on delivery
- **relations:** Substrate-vs-application pitch framing; Fractal agent architecture claim
- **verify-later:** n/a (pitch-writing artefact, not code)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Substrate-vs-application pitch framing
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** Framed explicitly as "Experimental alternative framing... To compare against, not necessarily replace, the original pitch doc"
- **what:** Repositions the same system from "I built a website builder" (a product) to "I built a domain-agnostic distributed agent orchestration substrate, and the website builder is one demonstration of it among five" (an asset/infrastructure claim). Chosen per audience: website framing for commercial/marketing-tech roles, substrate framing for AI-infrastructure roles. Comes with an explicit "words to use / avoid" list.
- **sources:** pitch/framework_pitch_substrate_framing.md#§1,§6,§9, pitch/002_substrate_framing_elevator_pitches.md#final notes
- **relations:** Fractal agent architecture claim; Honest-delta disclosure discipline; Audience-tuned elevator pitch variants
- **verify-later:** n/a

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Fractal agent architecture claim (self-similar recursive orchestration)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Backed by a traced real production call chain "seven levels deep with identical code paths at each level" (intake-orchestrator → site-work-orchestrator → build-dispatch-loop → page-build-handler → page-content-writer → research-agent → web-fetcher)
- **what:** The claim that every agent in the system is itself an orchestrator using identical primitives (spawn/call/claim/complete, same Kafka topic conventions, same orchestration_states shape) at every depth of the spawn tree, with no architectural distinction between a "top-level orchestrator" and a "leaf specialist." Framed as the single most defensible/highest-risk word in the pitch, contrasted against single-process Python frameworks (LangGraph/CrewAI/AutoGen).
- **sources:** pitch/framework_pitch_substrate_framing.md#§2, pitch/002_substrate_framing_elevator_pitches.md#final notes, pitch/framework_pitch_reference.md#§3.1
- **relations:** Substrate-vs-application pitch framing; Multi-cluster agent dispatch contract
- **verify-later:** orchestration_states.parent_orchestration_id chain query; agent_spawn_history

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Honest-delta disclosure discipline (built vs admitted-not-built table)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** A fully filled-in comparison table ("Honest Delta — What's Built vs What the Architecture Admits") described as "the load-bearing piece... what protects you from over-claiming"
- **what:** A pitch-integrity practice: pair every ambitious architectural claim with an explicit, honest ledger of what's actually proven in production versus merely structurally possible versus not built at all. Extends into a parallel "Honest Weak Points" section (solo development, documentation drift, schema drift, race conditions, incomplete migrations, no formal test suite) each with a pre-written honest-framing answer.
- **sources:** pitch/framework_pitch_substrate_framing.md#§7,§6.4, pitch/framework_pitch_reference.md#§10
- **relations:** Substrate-vs-application pitch framing; Fractal agent architecture claim
- **verify-later:** n/a

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Operator-vs-vendor business model fork / sell-the-framework separability
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "that resolves the strategic fork cleanly (operator-primary at scale, vendor-optional per domain)." (stripe/001commentary.md#§13); independently reached in the isolated-chat-environment discussion: "Design the seam now (entitlement abstraction + the two check points + entitlement state on client/network); build billing depth later." (tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md §13)
- **what:** Identifies that "operate thousands of domains yourself" and "sell the whole framework/instance to a buyer" pull toward opposite technical choices (centralised multi-tenant efficiency vs. clean separable sellable units), resolved as operator-primary at scale with per-domain vendor-optionality: the unit of blast-radius isolation (the satellite) is distinguished from the unit of separability-for-sale (the domain, partitioned within a satellite's stores via `site_id`/`network_id`). Recommends honouring five seams cheaply now — ownership on site rows, an entitlement gate at build-submission and at maintenance-run, a pluggable billing adapter, credential parameterisation, and a build-tier/cost-profile flag — because retrofitting separability later is "a forensic untangling."
- **sources:** stripe/001commentary.md#§13, stripe/001commentary.md#final turn, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13
- **relations:** Isolated chat/satellite architecture (Y-copy); Ownership hierarchy reuse for entitlement scoping; Entitlement gate architecture; Building-and-hosting as a service via chat
- **verify-later:** every domain-scoped table's site_id/network_id discipline; export-and-re-point procedure for per-domain sale

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Website platform mission (best site per domain, one pipeline)
- **category:** business-strategy
- **status-signal:** partial
- **status-evidence:** 028 header "Living document. Second revision (2026-04-22)"; "produce the best possible website for each … with minimal human input"
- **what:** The anchoring *why*: given any domain, produce the best possible site end-to-end through one agent graph, where "best" = most useful to probable visitors (measured by engagement) and best revenue via whatever model genuinely fits. Commercial viability ≠ a brochure "business site"; defaulting to consultancy/services/contact when the signal is absent is an explicit failure mode to counter.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-mission, WM/028_platform_mission_and_pipeline_direction.md#commercial-viability-is-not-the-same-as-a-business-site, WM/028_platform_mission_and_pipeline_direction.md#failure-modes-we-want-to-eliminate
- **relations:** classifier strategic brain; fidelity dial; interactive content generation
- **verify-later:** site_specs classification aspect; domain-research-classifier

<!-- SOURCE: U18_sql_for_agents.md -->
### domain-strategist (strategy vs architecture separation)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 060 definition with explicit responsibility statement; work item chain needs_strategy → needs_briefing.
- **what:** Handler for needs_strategy items. Determines the strategy for a domain — canonical site_type, revenue model, content strategy, page_type recommendations, tone/positioning — and writes site_specs aspect "strategy". Explicit contract: does NOT design page architecture; "The planner has final say... may agree, adjust, or override"; does NOT overwrite the researcher's "classification" aspect.
- **sources:** 060_domain_strategist.sql
- **relations:** build-site-planner reads strategy; domain-research-classifier upstream
- **verify-later:** strategy aspect consumption in plan_site prompt

<!-- SOURCE: U18_sql_for_agents.md -->
### Portfolio/use-case spec seeds (ai-agent-orchestration.com)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 100 INSERTs site_specs 'portfolio' aspect with five dated case studies claiming operational metrics ("Six production sites deployed and self-maintaining... under 4 hours" domain-to-live).
- **what:** Marketing-facing data seed whose case studies double as a platform capability inventory circa file-100: autonomous multi-site pipeline (30+ agents), tool generation + cross-linking, vet data platform, news aggregation with credibility scoring, and the orchestration layer itself (Kafka/Postgres/K8s, hot-swappable SQL workflow definitions, fuel budgets). Useful as documentary status evidence for many other concepts, not ground truth.
- **sources:** 100_portfolio_use_cases_etc.sql
- **relations:** nearly every pipeline concept above; site-case-studies
- **verify-later:** claims vs stage-2 code/DB verification

<!-- SOURCE: U19_sql_tables_components.md -->
### AI persona team and departments marketing model
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Applied site_specs updates (075a/076) injecting team and departments JSONB for ai-agent-orchestration.com, finetuning.uk, leopardessconsulting.co.uk, with audience-tuned copy per site.
- **what:** The platform presents itself through named AI managing-agent personas — Archivist (Research), Sentinel (Quality), Quartermaster (Operations) — alongside the human principal, plus an 8-department / 70+ agent structure with per-department agent counts and capability summaries. Stored as identity-spec data consumed by the content writer for team/departments sections; departments-grid component renders it as the leadership-team alternative.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#076
- **relations:** site_specs identity; departments-grid component; pitch/business docs.
- **verify-later:** rendered team/departments sections on the three sites.

<!-- SOURCE: U20_legacy_docs_a.md -->
### EBORG — evidence-based organisational planning (venture concept)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** Appears only as the demo business for HITL content approval (0103, full pitch text in the trigger script); never seen in later doc eras.
- **what:** A business idea used as the HITL demo client: map every role/responsibility/objective in an organisation and pair each with a framework of AI agents that gather research, assess options, and provide evidence-based reasoning — "human-centered, continuously learning organisation". Also spawned the simple-content-writer-with-approval agent.
- **sources:** docs002_hitl_parallel/README.0103.hitl_start_message.md; docs002_hitl_parallel/README.0102.hitl_agent_definitions
- **relations:** HITL content approval group (content-approval-hitl); thematically echoes the later council-of-experts idea in docs026 stage 3.
- **verify-later:** none (idea registry only).

<!-- SOURCE: U20_legacy_docs_a.md -->
### WordPress handoff (XML export, plugin shortcodes, SQL brand injection)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** Detailed plan in 004 (WordPress Formatter agent, wordpress-export.xml, [wpforms] shortcodes for stubs, wp_options SQL injection for branding); frontend-framework survey in 005; never mentioned again anywhere.
- **what:** Client-handoff strategy: transpile a generated site into a single WordPress import file so a client's developer gets a standard maintainable WP site in minutes; complex components become plugin shortcodes; brand colours/fonts injected into theme settings via one SQL file. Part of a broader survey of exit routes (traditional CMS vs SaaS builders vs headless/Jamstack).
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.005.frontend_frameworks.md
- **relations:** business-strategy (client/exit strategy); deployment-github (the retained path).
- **verify-later:** none — abandoned idea registry.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Two-tier commercialisation model (sell output → sell setup)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** docs016/004 dated 2026-03-03 "Working notes for strategic direction": three-tier model trimmed to two ("drop the domain selling tier"); practical next step "produce 10 websites in a niche... validate with real money".
- **what:** Frank commercial assessment: framework differentiators are real infrastructure (K8s/Kafka/Postgres, data-driven workflows, multi-cluster, chassis pattern) but lack docs/community; revenue paths ranked (website service most mature; SEO content; document processing needs domain partner; framework sales longest); recommended model — run the service in a chosen niche to accumulate live outputs, then sell the whole setup as a business-in-a-box (£5-25K) once 20-50 outputs prove it, repeating per product; canine project reframed as portfolio/demo spend. Open decisions: niche, sellable quality, solo vs partner, runway.
- **sources:** docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** business-strategy category (pitch, domain strategy); early portfolio inventory; canine biology demotion.
- **verify-later:** n/a (strategy).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Early portfolio inventory (honest capability notes)
- **category:** business-strategy
- **status-signal:** unknown
- **status-evidence:** docs018/008b raw notes: "None of our sites get leads at the moment so we can't say they do... we'd rather sell the sites achievement at the moment"; lists leopardessconsulting, vetcomparison.uk, wykefarm.co.uk, mortgagecalculator, website-design.com.
- **what:** A candid snapshot of what existed circa Feb 2026: leopardessconsulting built and evolving over days; a veterinary price-comparison site plus vet search/scrape/data-collection service; wykefarm.co.uk farm site (biodiversity content); a quickly-built but rough mortgage calculator site; website-design.com with functional tools (paste boards, mind maps, colour tools) but poor polish; framework scaling claim "several thousand agents". Useful ground truth for verifying which case-study sites actually functioned.
- **sources:** docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** site-case-studies (leopardess, vet, wyke); vet-med-pricing; dynamic-applications (website-design.com tools).
- **verify-later:** sites table rows for each named domain.

<!-- SOURCE: U22_recent_small_docs.md -->
### Verticals designed (revenue models + knowledge clusters)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** Revenue projection tables labelled "months 12-18" with market data "verified through research"; no live-site revenue claimed.
- **what:** Five verticals worked out with specific knowledge clusters, source lists, page-type libraries, monetisation, and 24-month revenue projections: veterinary/vetcomparison.uk (insurance affiliate £15-35 + listings, £1,960-7,875/mo), energy/gaswholesalers.com (qualified leads £30-60, £1,250-5,350), finance_mortgage/mortgagecalculator.co.uk (broker leads £50-150, £16,500-44,000 — highest value), seasonal_gifts/xmaspresents.com (affiliate 3-17%), plus a "sell the domain not develop" premium pathway (design.co.uk £20-100k).
- **sources:** docs021.../020_vertical_cluster_architecture.md#3, docs021.../025_session_handoff_vertical_architecture.md#verticals-designed, docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture, domain content strategy framework, premium-domain pathway
- **verify-later:** vertical_registry monetisation_config; any live vetcomparison/gaswholesalers/mortgagecalculator sites

<!-- SOURCE: U22_recent_small_docs.md -->
### Domain content strategy framework (15-question)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "For the content generation pipeline, the 15-question framework should feed into the briefing/research phase" — prescriptive/should, not implemented.
- **what:** A systematic three-layer, 15-question methodology for deciding what content a domain needs to compete: Layer 1 (who visits, intent, satisfaction, money flow), Layer 2 (competitor pages, buying journey, real questions, bookmarkable hook), Layer 3 (best page on the topic, original element, format, next action). Worked examples for gaswholesalers.com and vetcomparison.uk with verified lead/affiliate rates. Questions 5-7 require real competitive research.
- **sources:** docs022.../001_domain_content_strategy_framework.md, docs022.../002_domain_content_strategy_framework_v2.md
- **relations:** domain-strategist prompt, deep research domain authority, site classifier
- **verify-later:** domain-strategist agent prompt; briefing/research phase incorporating the framework

<!-- SOURCE: U22_recent_small_docs.md -->
### Deep research domain authority strategy
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "The multi-cluster knowledge base approach lets you build that deep knowledge layer for any domain" — strategy doc, canine project cited as proof-of-concept not production.
- **what:** The thesis that content wins on E-E-A-T by synthesising primary/authoritative sources (BSAVA, Ofgem, PRA/FCA, swap-rate data) into knowledge consumers can't easily find, rather than rephrasing published synthesis. A repeatable 6-step pipeline (niche mapping → primary-source identification → multi-cluster KB construction → gap identification → content architecture → generation from KB) creates a defensible moat: depth consistency, cross-cluster synthesis, and update efficiency competitors can't copy by rewriting one article.
- **sources:** docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture, domain content strategy framework, canine biology KB
- **verify-later:** research-agent primary-source handling; knowledge_base source_authority weighting

<!-- SOURCE: U22_recent_small_docs.md -->
### Content-site valuation model (24-32x)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "using current market multiples of 24-32x monthly profit (Empire Flippers averaging 24x, premium 30-35x)" — used throughout as a projection basis.
- **what:** The valuation basis underpinning the domain portfolio strategy: content/affiliate sites sell at ~24-32x monthly profit, so a £1,500-3,000/mo site is worth ~£36k-96k. Combined with verified per-niche lead/affiliate rates to project each domain's asset value and justify the knowledge-base investment. The portfolio is framed as the testing ground toward a £25k+ annual revenue target and a two-tier service→pipeline-sale model.
- **sources:** docs022.../002_domain_content_strategy_framework_v2.md#monetisation, docs021.../025_session_handoff_vertical_architecture.md#market-data-verified
- **relations:** verticals designed, commercial model (chatbot docs), premium-domain pathway
- **verify-later:** n/a (business assumption)

<!-- SOURCE: U22_recent_small_docs.md -->
### Building-and-hosting-as-a-service via chat (recursive platform)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Recorded because it sharply reframes the satellite ... (Discussion artefact; revisit as it firms up.)"
- **what:** A worked example where a chat box on design.co.uk becomes the intake+orchestration front-end to the whole build platform offered as a service: conversational briefing (a briefing-agent interview replacing the static form) → satellite intake orchestrator → full build workflows on the satellite → hybrid S3+lambda hosting → the new site itself gets a chatbot (recursion). Requires the full chassis on the satellite (rules out Option X for this use case) and surfaces new SaaS concerns: cost/abuse gating, accounts/billing/quotas, feeding reusable building blocks one-directionally from core.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#12
- **relations:** isolated chat environment (Y-copy), briefing-agent/intake orchestrator, commercial model, simple paid multi-domain chat
- **verify-later:** briefing-agent conversational mode; satellite intake orchestrator

<!-- SOURCE: U22_recent_small_docs.md -->
### Payable-differentiator framework (asset × AI capability)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "This is the method ... It is a starting point and needs more thought and testing — the menus below are incomplete, the scoring is rough."
- **what:** A method for justifying a paid chat when there's no proprietary data: value comes not from the model (everyone has it) but from a hard-to-reproduce asset (proprietary/paid data feed, an owned process/output, a well-built tool, a commercial partnership, or early access to a new AI capability) combined with AI for a paying audience. Maintain two menus (assets; AI-capabilities-worth-using-now) and pair one of each per domain. Worked examples: websitedesign.com (package our site-spec/plan as a starter prompt for Bolt/Lovable — strongest), gaswholesalers.com (buy oil/gas data feeds), agritec.uk (partnership vouchers). Prioritise by reproducibility, willingness-to-pay, build cost, cross-domain reuse; test willingness-to-pay before building.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#10
- **relations:** simple paid multi-domain chat, ideation agent, verticals designed
- **verify-later:** n/a (strategy)

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### finetuning.uk self-service RAG product (business strategy)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN(1) §1 "A RAG platform for SMEs"; FOCUS(21) §5 "This is a product decision, not a technical one" and §10 shipping ladder with "Aspirational dates, not promises"
- **what:** The plan to turn finetuning.uk into a paid product. Direction pivoted several times: from "self-service fine-tuning SaaS" → RAG-over-your-docs platform with automatic data curation as the named differentiator; from "concierge first, UI later" → UI-first ("build the cockpit we use ourselves"); target user refined up from non-technical owners to technical-adjacent SME ops leads. Pricing £199–1,499/mo + setup fees; £9-12k/mo solo-operator target; £100k year-1 projection.
- **sources:** BUSINESS_PLAN_finetuning_uk(1).md#1, #5, #7; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#5, #7, #8, #10
- **relations:** reuses internal-flywheel infra (Ollama, Unsloth, rag_index); replacement direction for the abandoned "self-serve fine-tuning" pitch
- **verify-later:** finetuning.uk site_specs; tenant_id on knowledge_base (not yet built)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Moat / differentiator framework (asset × AI × audience-that-pays)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "A payable idea = asset × AI × audience-that-pays. Five asset types: proprietary data, owned process/output, well-built tool, partnership, early-mover timing on a new capability."
- **what:** A doctrine used to filter which domains/products are worth building: the AI model is never the differentiator, the asset it's applied to is. Honest verdict reached during idea.uk's own self-analysis: its moat is "effort + freshness + integration... sustained by maintenance, not a static asset," not a structural moat. Cross-domain pattern discovered via repeated method runs: "wherever the underlying product has high margin, the seller already gives expert support away free" (Bloomberg/Refinitiv, Open Bionics, Robotiq) — an almost-automatic cut for "help-you-buy-X" candidates.
- **sources:** `running_notes(44).md` ("The differentiator framework", "Moat analysis (idea.uk)", "New cross-domain pattern: high-margin-product sellers bundle support free")
- **relations:** idea generation method
- **verify-later:** n/a (a design doctrine, not code)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Five-layer consolidation model (L0–L5)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "Consolidation map written... wrote CONSOLIDATION_where_it_all_fits.md — five layers."
- **what:** A planning frame reconciling the idea.uk work with the wider platform roadmap: L0 chassis (exists, builds static sites), L1 idea engine (built, standalone: method + internal CLI + idea.uk), L2 idea.uk product (in progress, first to go live), L3 vertical tools (in progress, chassis-native, e.g. SFI26 Diff Alerts), L4 tool-rich site building for any domain (future — the idea engine becomes a *planning input* to the chassis site builder), L5 automated VM backend deployment (future — today's pipeline only deploys static→B2; provisioning+deploying a persistent backend is the gap Thunder is the seed of). States the natural build order is "prove L1 → ship L2 → build L3 once → generalise into L4 → grow L5 from Thunder."
- **sources:** `running_notes(44).md` ("Consolidation map written")
- **relations:** service-deployer pattern (= the L5 gap); chassis-native idea engine (= L4); idea.uk product (= L2)
- **verify-later:** `CONSOLIDATION_where_it_all_fits.md` (live doc, out of this unit's scope)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### SaaS commercial model — operator-primary, vendor-optional, with entitlement seams
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Resolved direction" language for the model itself, but the concrete implementation ("Design the seam now... build the billing depth later") is explicitly deferred — "None of these constrain the build as long as separability and the seams above are honoured."
- **what:** A commercial model for operating the platform at scale while keeping individual sites/backends sellable: **operator-primary, vendor-optional** — operate thousands of domains directly, with the option to sell a domain plus its backend (the common case) or, rarely, the whole framework/instance. The key structural insight is that the unit of blast-radius isolation (the satellite/cluster) and the unit of sale-separability (the domain) are different granularities — operating thousands of domains does not require thousands of clusters; it requires clean per-domain partitioning (keyed on `site_id`/`domain`) plus the ability to extract one domain's artifact, data, and credentials at sale time. Five seams are flagged as cheap now / expensive to retrofit: `owner_id` on site rows; an entitlement check at both build-submission and maintenance-run (never calling Stripe directly — always through a pluggable billing-adapter interface); credential parameterization everywhere (no hardcoded keys); and a build-tier/cost-profile flag (`saas_cheap` vs `portfolio`) driving cheaper model/batching choices so low-price builds retain margin.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md §13
- **relations:** isolated chat satellite architecture (above, same document)
- **verify-later:** whether an owner_id/entitlement layer or billing adapter exists anywhere in the schema/codebase today

<!-- SOURCE: U25_leopardess_social.md -->
### Data-sovereignty / pilot-first / startup fast-start commercial positioning
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK H6/H8/H9 all "resolved 2026-07-10" as *drafted positioning*, each explicitly "Not yet done for a client".
- **what:** Three owner-confirmed engagement angles drafted into specs/portfolio.json use_cases with honest labelling: (A8) data-sovereignty as a capability built *with* a client during a scoped engagement, never a standing guarantee; (H6) pilot-first engagement ladder (bounded fixed-price pilot → licence/day-rate/retainer decided by what the pilot reveals); (H9) startups building agent products start from the platform's already-solved operational layer (state, retries, HITL, no-redeploy workflow changes). Plus register-reconciliation generalisation ("your list vs an authoritative register" as the general shape of the Companies House work).
- **sources:** docs/leopardessconsulting/specs/portfolio.json#use_cases; docs/leopardessconsulting/RUNBOOK.md#H6-H9; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6, #Turn-7
- **relations:** per-step model routing; no-tenant-isolation; claim-evidence audit rule
- **verify-later:** site_specs aspect 'portfolio' current row

<!-- SOURCE: U25_leopardess_social.md -->
### UK-sovereign stack exploration (deferred)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK Reference: "explicitly deferred to a separate chat by the owner, 2026-07-10 … Do not start this unprompted."
- **what:** Future exercise: a fully UK-hosted compute+storage+model stack. Baseline facts captured so the future thread doesn't re-derive them: compute Rackspace UK; storage Backblaze us-east-005 (US); Anthropic and Google models US; self-hosted path sits wherever the cluster is.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Reference; docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P6
- **relations:** data-sovereignty positioning; storage-architecture
- **verify-later:** memory `uk-sovereign-stack-exploration`

<!-- SOURCE: U26_misc_dirs.md -->
### WordPress export agent and content-subscription plugin
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 008 designs it enthusiastically ("This is GENIUS!"); 009 immediately deconstructs it ("You're not unique in 'AI builds WordPress sites'. That market is saturated... Only add WordPress if they're begging for it") — never mentioned again anywhere.
- **what:** An agent converting generated HTML sites into installable WordPress themes + WXR content exports, paired with a WP plugin subscribing to the platform for auto-published fresh content (recurring revenue). Explicitly killed by the competitive analysis in 009, which redirected differentiation toward "sites that update themselves" / continuous content ecosystems.
- **sources:** docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md#wordpress-export-agent-design; docs/architecture/009-wordpress-discussion#the-hard-truth
- **relations:** HTML-first delivery; living-content differentiation (which the platform did pursue via news feed / content pipelines)
- **verify-later:** n/a (never built)

<!-- SOURCE: U26_misc_dirs.md -->
### Domain value maximisation pipeline (domain flipping)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 010/011/015 lay out the strategy against the user's real domain portfolio (collateralfinancing.com, holidaytime.com, websitedesign.com...) with 48-hour development timelines; no later doc pursues domain flipping — the platform pivoted to operating its own sites.
- **what:** Use the agent platform to develop parked domains into sites with content, traffic and revenue to multiply sale value (naked $500 → revenue-bearing $10k+): domain classification (brandable/exact-match/local/product), tiered portfolio treatment, 48h batch development, monetisation setup (leads/affiliate/ads), and "self-selling" footers that market the build service from every developed domain.
- **sources:** docs/architecture/010-domain-value-maximisation.md; docs/architecture/011-example-domains; docs/architecture/015-underserved-niche.md#your-domain-portfolio-is-your-marketplace
- **relations:** deep-research domain insight agent; underserved-niche strategy; site-case-studies (the surviving practice of operating exemplar sites)
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### Underserved-niche and vertical showcase strategy
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 015/016 propose niches (compliance docs, local-business packages, academic assistants, affiliate content) and per-industry showcase domains with a workflow marketplace ("DIY $500 / DFY $200mo / White Label $2000mo"); positioning discussion only, no implementation trail.
- **what:** Rather than competing with Temporal/LangChain/Zapier broadly, own narrow verticals where multi-agent coordination wins: each showcase domain demos an industry solution (legal docs, restaurant launch, real-estate listings) and funnels to purchasable workflows. Includes the pricing-tier and "Business-in-a-Box" (site + content pipeline + email + social) framings, and the investor-demo positioning of the framework as the star with swappable use cases.
- **sources:** docs/architecture/015-underserved-niche.md; docs/architecture/016-competitive-advantge.md#who-actually-pays-for-ai-sites; docs/architecture/012-investors.md#the-portfolio-approach
- **relations:** domain value maximisation; EBORG organizational OS
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### AI-native orchestration positioning (vs Temporal/Airflow)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 012/014 are interview/investor argumentation ("You could build this on Temporal, but it would be like using Kubernetes to run a single container"); the accompanying Temporal/Airflow adapter agents were never built.
- **what:** The articulated build-vs-buy rationale for the platform: AI-specific needs (dynamic JSON workflows without deployment, token/fuel tracking, prompt management, AI-failure handling, multi-tenant agent isolation, workflows spawning from AI decisions) justify a purpose-built orchestrator. Includes proposed Temporal-adapter and Airflow-adapter agents to bridge into enterprise workflow estates as a migration path ("we don't replace your existing systems, we enhance them").
- **sources:** docs/architecture/012-investors.md#better-answer; docs/architecture/014-Temporal-Airflow-adapters.md
- **relations:** adapters (current adapter guide is a different, real lineage); distributed embedded orchestration
- **verify-later:** confirm no temporal/airflow adapter code exists

<!-- SOURCE: U26_misc_dirs.md -->
### EBORG — Evidence-Based Organisational Planning
- **category:** business-strategy
- **status-signal:** unknown
- **status-evidence:** The Nov 2025 HITL demo is branded "For EBORG" with a full pitch paragraph in start_hitl_workflow.sh ("mapping every role, responsibility, and objective, then pairing each with a framework of AI agents"); no other doc in this unit elaborates whether it became a product.
- **what:** A product concept: organisations map roles/responsibilities/objectives and pair each with AI agents that gather research, assess options and provide evidence-based reasoning — "a human-centered, continuously learning organisation" and the concrete descendant of 016's Organizational OS idea. Used as the demo business in the HITL content-approval workflow.
- **sources:** docs/humanintheloop/start_hitl_workflow.sh; docs/humanintheloop/hitl_agent_definition.sql (header comment); docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** cross-domain intelligence network; HITL content-approval demo
- **verify-later:** any EBORG references in business/vonc docs (other units)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Platform mission and the single unified pipeline
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 028 (living document, rev 2026-04-22) is the stated check-yourself-against anchor
- **what:** Mission: given domain names in any state, produce the best possible website end-to-end with minimal human input, "best" = most useful to probable visitors measured by real engagement AND best revenue via whatever model genuinely fits. One pipeline for blank/adopted/missioned/replication domains — differing only in input material and the fidelity dial. Revenue model shapes the site (default-to-brochure/consultancy is a named failure mode); classifier decides the commercial shape.
- **sources:** 028#The mission, #Commercial viability
- **relations:** fidelity dial; classifier as strategic brain
- **verify-later:** —

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### finetuning.uk self-service product strategy
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Not started. Questions to answer before scoping"; 12 decisions dated 2026-04-21; shipping ladder "Aspirational dates, not promises"
- **what:** finetuning.uk as both a credible knowledge site and a revenue product: flagship = RAG platform with data curation as a first-class visible feature (parse/classify/dedupe/quality-score/PII-scan/inconsistency-flag pipeline reusing the framework), concierge-onboarded then self-serve; tiers from £199/mo platform to £15-30k bespoke; target user technical-adjacent SMEs; UI-first build as own operational cockpit; explicit not-to-ship list (multi-tenant fine-tuning SaaS, public API). Differentiation from positioning (UK residency, opinionated simplicity, self-improvement loop) not engineering.
- **sources:** FOCUS_finetuning_flywheel_and_service(13).md#5, #7, #8, #8a, #10, #11
- **relations:** internal flywheel infra reuse table (#6); knowledge_base tenant_id plan
- **verify-later:** state of finetuning.uk site; any tenant_id on knowledge_base

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Platform mission restatement (plan and build websites from a domain)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Checkpoint (Um), restated by the user 2026-07-07: "intelligently PLAN and BUILD multipage websites from a domain name — targeted design/content per vertical, eventually parsing best-in-class exemplars (reasoning from why they work, not copying), adding tools/blog/news/infographics from wider-world reasoning."
- **what:** The system's purpose as the user states it mid-thread when re-anchoring scope: an agent system that plans and builds whole multipage sites from a domain name, with vertical-targeted design and content, future exemplar-parsing (reasoning about why best-in-class sites work rather than copying), and enrichment via tools, blog, news and infographics; supported by close agent/message logging, an agent-creation guidelines doc, and distinct low-overlap agent responsibilities with sub-agents for research-before-content.
- **sources:** running_notes_scheme_to_components(55).md#Um
- **relations:** orchestrator conventions; research-agents; adoption-pipeline (exemplar parsing kinship).
- **verify-later:** agent-creation guidelines doc; exemplar-researcher plans elsewhere in docs.

<!-- SOURCE: U04_idea_uk.md -->
### Five-layer platform stack (chassis → idea engine → idea.uk → vertical tools → tool-rich sites → VM backend deploy)
- **category:** business-strategy
- **status-signal:** partial
- **status-evidence:** "Where it all fits" map dated 2026-06-04: Layer 0 EXISTS, Layer 1 BUILT, Layers 2–3 IN PROGRESS, Layers 4–5 FUTURE ("Thunder adapter is the seed").
- **what:** A consolidation model presenting the whole enterprise as one stack: the chassis builds sites (L0); the idea engine decides what's worth building (L1); idea.uk sells that externally (L2); recommended tools get built for real, chassis-native (L3); the engine becomes a planning input so any domain gets a tool-rich site (L4 — "the original problem statement"); and automated backend deployment onto VMs closes the last gap (L5). Each layer is a customer of the one below.
- **sources:** idea.uk/CONSOLIDATION_where_it_all_fits.md; idea.uk/PARALLEL_engine_deployment_and_layer5.md
- **relations:** Layer-5 persistent-service wrapper; SFI26 Diff Alerts; chassis-native idea engine (Phase D); Thunder adapter (docs033/035).
- **verify-later:** existence of any service-deployer agent; site_plan aspects carrying blocked/planned tool items; thunder-adapter actions.

<!-- SOURCE: U04_idea_uk.md -->
### Differentiator framework — payable idea = hard-to-reproduce asset × current AI capability
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** PLAN_idea_uk §3/§5 (framework in use); method encodes it as the core principle; testruns v0/v2 applied it across 8 domain runs.
- **what:** The AI model is never the differentiator (everyone has the same models); the defensible unit is an asset × capability aimed at an audience that will pay, doing something a free model with a good prompt cannot. Honest moat verdict: the durable advantages are **currency** (a maintained capability watchlist beating models' self-knowledge), **verification with evidence**, and the **build bridge** (we can build the idea, not just describe it) — a process/freshness/integration advantage, not a static asset. Includes the brand-fit corollary (treat the product collection as separate from the domain portfolio; match deliberately).
- **sources:** idea.uk/PLAN_idea_uk(3).md#5; idea.uk/idea_uk_method_v0(3).md; idea.uk/running_notes(63).md (2026-05-27 arc)
- **relations:** ideation method; capability watchlist; five-layer stack; paid multi-domain chat plan (§10 of that doc).
- **verify-later:** whether the capability watchlist exists as a recurring workflow anywhere in scheduled_tasks/agent_definitions.

<!-- SOURCE: U04_idea_uk.md -->
### Sale-readiness / separability discipline (assets as data, minimal identifiable dependency set)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** PLAN_idea_uk §2 rule "keep our asset list as input data, never built into the method"; RUNBOOK_idea_uk Notes: "the engine takes assets as data and the billing sits behind a provider interface, so idea.uk remains a separable unit".
- **what:** idea.uk is built to be sold as a working unit: business assets are always passed in as data (so the same engine serves internal domains and strangers), the set of workflows/actions it uses is kept identifiable and minimal, and billing sits behind a provider interface. The standalone Go service honours this (stdlib-only, file store, FakeProvider fallback).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/RUNBOOK_idea_uk(9).md; idea.uk/idea_uk_architecture_and_deployment(6).md#1
- **relations:** provider abstraction (payments); engine Go port.
- **verify-later:** golang_files/engine.go input contract; billing.go Provider interface.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk as an instance of the paid multi-domain chat plan (day-pass lineage)
- **category:** business-strategy
- **status-signal:** superseded
- **status-evidence:** PLAN_idea_uk §2 "idea.uk is itself an instance of the paid multi-domain chat"; the built product ended up a report service, not a chat domain — the worker/paywall/day-pass reuse never happened in the shipped form.
- **what:** idea.uk originated as one configured domain of a planned "simple paid multi-domain chat" product (edge worker + paywall + day-pass), with the ideation method as its bound tool. The 2026-05-27 running-notes arc covers day-pass economics, per-domain monetisation by domain type, and serverless-edge vs central-nginx topology. The shipped idea.uk deliberately diverged: it is NOT edge-shaped (minutes-long background job → always-on service).
- **sources:** idea.uk/PLAN_idea_uk(3).md#2; idea.uk/running_notes(63).md (2026-05-27, "Pivot to simple paid multi-domain chat", "Topology note: idea.uk is NOT pure-static/edge")
- **relations:** PLAN_simple_paid_multidomain_chat.md (outside this unit); hosting split concept below.
- **verify-later:** whether the chat/day-pass product exists anywhere else in docs/ (other units).

<!-- SOURCE: U04_idea_uk.md -->
### Voluntary pay and "free goes" rejected → free taster + paid report
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** idea_uk_open_discussion §5 (2026-05-28): "probably not a good idea in this form… Drop voluntary pay and the multi-free-go idea. The taster is the better hook."
- **what:** Voluntary-pay ("pay if satisfied") and N-free-goes monetisation were analysed and rejected (abuse risk, no demand signal, trivially circumvented). Replaced by the pattern that shipped: a free, cheap (~£0.02) audience-check taster as proof-of-value plus a £29 full report with refund guarantee.
- **sources:** idea.uk/idea_uk_open_discussion.md#5; idea.uk/running_notes(63).md ("Day-pass collapses payment complexity", CHECKPOINT 2026-05-28 §4)
- **relations:** audience-check taster endpoint; pricing decisions.
- **verify-later:** n/a (business decision).

<!-- SOURCE: U04_idea_uk.md -->
### Unit economics, pricing, and sourcing decisions (incl. self-hosting deferred)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** running_notes SESSION DECISIONS LOG 2026-06-11; open_discussion §§1–2, 6 with verified May-2026 pricing; £29 live and proven with a real card 2026-06-14.
- **what:** Per-run engine cost ~£0.40–0.60 (verify step dominates; optimisable to ~£0.20–0.30 via Haiku scoring + prompt caching); Stripe UK fees 1.5%+£0.20, break-even ~£0.72, worst-case refund cost ~£1.43; price settled at **pay-per-idea, cost-plus, £29 flat** (not B2B SaaS for the ideation product itself). Self-hosted LLMs analysed and deferred ("a 2027 decision, not a 2026 one") — commercial frontier models win at this volume, and open-weight models are weakest exactly at the cut step's ruthlessness.
- **sources:** idea.uk/idea_uk_open_discussion.md#1-2,6; idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md (pricing checkpoint)
- **relations:** Stripe webhook pattern; engine model line-up.
- **verify-later:** REPORT_PRICE_GBP env on the live box.

<!-- SOURCE: U06_finetuning.md -->
### finetuning.uk product strategy (RAG platform flagship, data curation as the product)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN "Last touched: 2026-04-21… Numbers are planning estimates"; FOCUS(25) §7 decisions 11-12; no later doc records any milestone reached.
- **what:** The external-product thesis: finetuning.uk becomes a RAG platform for technical-adjacent SMEs (10-50 people, knowledge-intensive, UK/EU) whose differentiator is automatic data curation (parse/classify/dedup/quality-score/PII-scan/inconsistency-flag with a visible curation report) — "competitors treat bad data as the customer's problem; we treat it as the product." RAG chosen ahead of text/image LoRA tiers (users arrive with docs, not training pairs; the infra is built). Self-service *fine-tuning* SaaS is explicitly deferred/not-shipped. Reuse map: same Ollama/Unsloth/export/eval plumbing; entirely new: multi-tenancy, billing, UI, support, legal. Week-1 technical items named: tenant_id on knowledge_base enforced in rag_lookup/rag_index; auth stack choice.
- **sources:** BUSINESS_PLAN_finetuning_uk.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#5-13
- **relations:** knowledge_base RAG; internal flywheel (shared plumbing); UI-first decision
- **verify-later:** tenant_id on knowledge_base; any finetuning.uk app code; site_specs for finetuning.uk

<!-- SOURCE: U06_finetuning.md -->
### finetuning.uk business plan (pricing, unit economics, milestones)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN changelog: "2026-04-21 — Initial draft… refine as data arrives"; no subsequent updates in this tree.
- **what:** Solo-operator plan: tiers Trial/£199/£499/£1,499/Enterprise plus concierge fees (£750 audit → £15-30k bespoke); gross margin 57-78% per Growth customer; break-even ~5 Growth customers; 12-month target £9-12k/month and ~£100k year-1 revenue; content-led cold acquisition only (the framework as content engine is the claimed structural moat); interim gigs capped at 50% of time; assumption list with explicit 60-day tests; milestone/decision gates at months 1/3/6/12. Superseded staging inside the family: v1's "concierge first, UI later" three-tier structure was replaced by the UI-first "build our own cockpit" revision (2026-04-21 fourth pass).
- **sources:** BUSINESS_PLAN_finetuning_uk.md; working/flywheel_docs/FOCUS_finetuning_flywheel_and_service_v1.md#8,#10 (superseded shape); working/flywheel_docs/FOCUS_finetuning_flywheel_and_service(25).md#8,#10
- **relations:** product strategy; vonc/business-strategy docs elsewhere
- **verify-later:** nothing technical; check for any later business docs superseding this

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Building-and-hosting as a service via chat (build-as-a-service reframing)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Recorded because it sharply reframes the satellite ... (Discussion artefact; revisit as it firms up.)" (PLAN_isolated_chat_environment(5).md §12)
- **what:** A worked example where a site's own chatbot becomes the intake + orchestration front-end to the whole build platform offered as a service: a prospective customer types a domain + spec into an existing site's chat box, and the full build pipeline runs on the satellite (not core) to produce a new hosted site — itself shipped with its own chat box (explicit recursion). Reframes the satellite from an isolation nicety into a required second, customer-facing instance of the whole platform, and surfaces net-new concerns: cost/abuse exposure from anonymous builds, need for accounts/billing/quota gating.
- **sources:** tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#12
- **relations:** Isolated chat/satellite architecture (Y-copy); Operator-vs-vendor business model fork; Conversational build-intake via briefing-agent chat; Agent-to-adapter maturation path
- **verify-later:** conversational `briefing-agent` reusing `018_briefing_questionnaire`/`002_intake_orchestrator` is proposed but not confirmed to exist

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Audience-tuned elevator pitch variants (V1-V4 method)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Four fully-written 25-95 word variants plus 10-second openers, each annotated with likely follow-up question and delivery notes — a finished deliverable, not a plan
- **what:** A pitch-crafting technique producing four distinct ~25-35-second verbal pitches for the same underlying platform, each tuned to a different audience (technical-peer contrarian opener; commercial/investor asset-framing; mixed-audience concrete-first default; written/cold-context compressed version), plus even-shorter 10-second openers. Explicitly engineers what's deliberately left out to keep each version deliverable in one breath.
- **sources:** pitch/002_substrate_framing_elevator_pitches.md#V1-V4, pitch/002_substrate_framing_elevator_pitches.md#Notes on delivery
- **relations:** Substrate-vs-application pitch framing; Fractal agent architecture claim
- **verify-later:** n/a (pitch-writing artefact, not code)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Substrate-vs-application pitch framing
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** Framed explicitly as "Experimental alternative framing... To compare against, not necessarily replace, the original pitch doc"
- **what:** Repositions the same system from "I built a website builder" (a product) to "I built a domain-agnostic distributed agent orchestration substrate, and the website builder is one demonstration of it among five" (an asset/infrastructure claim). Chosen per audience: website framing for commercial/marketing-tech roles, substrate framing for AI-infrastructure roles. Comes with an explicit "words to use / avoid" list.
- **sources:** pitch/framework_pitch_substrate_framing.md#§1,§6,§9, pitch/002_substrate_framing_elevator_pitches.md#final notes
- **relations:** Fractal agent architecture claim; Honest-delta disclosure discipline; Audience-tuned elevator pitch variants
- **verify-later:** n/a

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Fractal agent architecture claim (self-similar recursive orchestration)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Backed by a traced real production call chain "seven levels deep with identical code paths at each level" (intake-orchestrator → site-work-orchestrator → build-dispatch-loop → page-build-handler → page-content-writer → research-agent → web-fetcher)
- **what:** The claim that every agent in the system is itself an orchestrator using identical primitives (spawn/call/claim/complete, same Kafka topic conventions, same orchestration_states shape) at every depth of the spawn tree, with no architectural distinction between a "top-level orchestrator" and a "leaf specialist." Framed as the single most defensible/highest-risk word in the pitch, contrasted against single-process Python frameworks (LangGraph/CrewAI/AutoGen).
- **sources:** pitch/framework_pitch_substrate_framing.md#§2, pitch/002_substrate_framing_elevator_pitches.md#final notes, pitch/framework_pitch_reference.md#§3.1
- **relations:** Substrate-vs-application pitch framing; Multi-cluster agent dispatch contract
- **verify-later:** orchestration_states.parent_orchestration_id chain query; agent_spawn_history

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Honest-delta disclosure discipline (built vs admitted-not-built table)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** A fully filled-in comparison table ("Honest Delta — What's Built vs What the Architecture Admits") described as "the load-bearing piece... what protects you from over-claiming"
- **what:** A pitch-integrity practice: pair every ambitious architectural claim with an explicit, honest ledger of what's actually proven in production versus merely structurally possible versus not built at all. Extends into a parallel "Honest Weak Points" section (solo development, documentation drift, schema drift, race conditions, incomplete migrations, no formal test suite) each with a pre-written honest-framing answer.
- **sources:** pitch/framework_pitch_substrate_framing.md#§7,§6.4, pitch/framework_pitch_reference.md#§10
- **relations:** Substrate-vs-application pitch framing; Fractal agent architecture claim
- **verify-later:** n/a

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Operator-vs-vendor business model fork / sell-the-framework separability
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "that resolves the strategic fork cleanly (operator-primary at scale, vendor-optional per domain)." (stripe/001commentary.md#§13); independently reached in the isolated-chat-environment discussion: "Design the seam now (entitlement abstraction + the two check points + entitlement state on client/network); build billing depth later." (tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md §13)
- **what:** Identifies that "operate thousands of domains yourself" and "sell the whole framework/instance to a buyer" pull toward opposite technical choices (centralised multi-tenant efficiency vs. clean separable sellable units), resolved as operator-primary at scale with per-domain vendor-optionality: the unit of blast-radius isolation (the satellite) is distinguished from the unit of separability-for-sale (the domain, partitioned within a satellite's stores via `site_id`/`network_id`). Recommends honouring five seams cheaply now — ownership on site rows, an entitlement gate at build-submission and at maintenance-run, a pluggable billing adapter, credential parameterisation, and a build-tier/cost-profile flag — because retrofitting separability later is "a forensic untangling."
- **sources:** stripe/001commentary.md#§13, stripe/001commentary.md#final turn, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13
- **relations:** Isolated chat/satellite architecture (Y-copy); Ownership hierarchy reuse for entitlement scoping; Entitlement gate architecture; Building-and-hosting as a service via chat
- **verify-later:** every domain-scoped table's site_id/network_id discipline; export-and-re-point procedure for per-domain sale

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Website platform mission (best site per domain, one pipeline)
- **category:** business-strategy
- **status-signal:** partial
- **status-evidence:** 028 header "Living document. Second revision (2026-04-22)"; "produce the best possible website for each … with minimal human input"
- **what:** The anchoring *why*: given any domain, produce the best possible site end-to-end through one agent graph, where "best" = most useful to probable visitors (measured by engagement) and best revenue via whatever model genuinely fits. Commercial viability ≠ a brochure "business site"; defaulting to consultancy/services/contact when the signal is absent is an explicit failure mode to counter.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-mission, WM/028_platform_mission_and_pipeline_direction.md#commercial-viability-is-not-the-same-as-a-business-site, WM/028_platform_mission_and_pipeline_direction.md#failure-modes-we-want-to-eliminate
- **relations:** classifier strategic brain; fidelity dial; interactive content generation
- **verify-later:** site_specs classification aspect; domain-research-classifier

<!-- SOURCE: U18_sql_for_agents.md -->
### domain-strategist (strategy vs architecture separation)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 060 definition with explicit responsibility statement; work item chain needs_strategy → needs_briefing.
- **what:** Handler for needs_strategy items. Determines the strategy for a domain — canonical site_type, revenue model, content strategy, page_type recommendations, tone/positioning — and writes site_specs aspect "strategy". Explicit contract: does NOT design page architecture; "The planner has final say... may agree, adjust, or override"; does NOT overwrite the researcher's "classification" aspect.
- **sources:** 060_domain_strategist.sql
- **relations:** build-site-planner reads strategy; domain-research-classifier upstream
- **verify-later:** strategy aspect consumption in plan_site prompt

<!-- SOURCE: U18_sql_for_agents.md -->
### Portfolio/use-case spec seeds (ai-agent-orchestration.com)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** 100 INSERTs site_specs 'portfolio' aspect with five dated case studies claiming operational metrics ("Six production sites deployed and self-maintaining... under 4 hours" domain-to-live).
- **what:** Marketing-facing data seed whose case studies double as a platform capability inventory circa file-100: autonomous multi-site pipeline (30+ agents), tool generation + cross-linking, vet data platform, news aggregation with credibility scoring, and the orchestration layer itself (Kafka/Postgres/K8s, hot-swappable SQL workflow definitions, fuel budgets). Useful as documentary status evidence for many other concepts, not ground truth.
- **sources:** 100_portfolio_use_cases_etc.sql
- **relations:** nearly every pipeline concept above; site-case-studies
- **verify-later:** claims vs stage-2 code/DB verification

<!-- SOURCE: U19_sql_tables_components.md -->
### AI persona team and departments marketing model
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** Applied site_specs updates (075a/076) injecting team and departments JSONB for ai-agent-orchestration.com, finetuning.uk, leopardessconsulting.co.uk, with audience-tuned copy per site.
- **what:** The platform presents itself through named AI managing-agent personas — Archivist (Research), Sentinel (Quality), Quartermaster (Operations) — alongside the human principal, plus an 8-department / 70+ agent structure with per-department agent counts and capability summaries. Stored as identity-spec data consumed by the content writer for team/departments sections; departments-grid component renders it as the leadership-team alternative.
- **sources:** docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql#076
- **relations:** site_specs identity; departments-grid component; pitch/business docs.
- **verify-later:** rendered team/departments sections on the three sites.

<!-- SOURCE: U20_legacy_docs_a.md -->
### EBORG — evidence-based organisational planning (venture concept)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** Appears only as the demo business for HITL content approval (0103, full pitch text in the trigger script); never seen in later doc eras.
- **what:** A business idea used as the HITL demo client: map every role/responsibility/objective in an organisation and pair each with a framework of AI agents that gather research, assess options, and provide evidence-based reasoning — "human-centered, continuously learning organisation". Also spawned the simple-content-writer-with-approval agent.
- **sources:** docs002_hitl_parallel/README.0103.hitl_start_message.md; docs002_hitl_parallel/README.0102.hitl_agent_definitions
- **relations:** HITL content approval group (content-approval-hitl); thematically echoes the later council-of-experts idea in docs026 stage 3.
- **verify-later:** none (idea registry only).

<!-- SOURCE: U20_legacy_docs_a.md -->
### WordPress handoff (XML export, plugin shortcodes, SQL brand injection)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** Detailed plan in 004 (WordPress Formatter agent, wordpress-export.xml, [wpforms] shortcodes for stubs, wp_options SQL injection for branding); frontend-framework survey in 005; never mentioned again anywhere.
- **what:** Client-handoff strategy: transpile a generated site into a single WordPress import file so a client's developer gets a standard maintainable WP site in minutes; complex components become plugin shortcodes; brand colours/fonts injected into theme settings via one SQL file. Part of a broader survey of exit routes (traditional CMS vs SaaS builders vs headless/Jamstack).
- **sources:** docs004_website_capture_project/website_analysis/README.004.backend.summary_ideas.md; docs004_website_capture_project/website_analysis/README.005.frontend_frameworks.md
- **relations:** business-strategy (client/exit strategy); deployment-github (the retained path).
- **verify-later:** none — abandoned idea registry.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Two-tier commercialisation model (sell output → sell setup)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** docs016/004 dated 2026-03-03 "Working notes for strategic direction": three-tier model trimmed to two ("drop the domain selling tier"); practical next step "produce 10 websites in a niche... validate with real money".
- **what:** Frank commercial assessment: framework differentiators are real infrastructure (K8s/Kafka/Postgres, data-driven workflows, multi-cluster, chassis pattern) but lack docs/community; revenue paths ranked (website service most mature; SEO content; document processing needs domain partner; framework sales longest); recommended model — run the service in a chosen niche to accumulate live outputs, then sell the whole setup as a business-in-a-box (£5-25K) once 20-50 outputs prove it, repeating per product; canine project reframed as portfolio/demo spend. Open decisions: niche, sellable quality, solo vs partner, runway.
- **sources:** docs016_dogs_medicine_pathways/004_medical_business_reality_assessment.md; docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** business-strategy category (pitch, domain strategy); early portfolio inventory; canine biology demotion.
- **verify-later:** n/a (strategy).

<!-- SOURCE: U21_legacy_docs_b.md -->
### Early portfolio inventory (honest capability notes)
- **category:** business-strategy
- **status-signal:** unknown
- **status-evidence:** docs018/008b raw notes: "None of our sites get leads at the moment so we can't say they do... we'd rather sell the sites achievement at the moment"; lists leopardessconsulting, vetcomparison.uk, wykefarm.co.uk, mortgagecalculator, website-design.com.
- **what:** A candid snapshot of what existed circa Feb 2026: leopardessconsulting built and evolving over days; a veterinary price-comparison site plus vet search/scrape/data-collection service; wykefarm.co.uk farm site (biodiversity content); a quickly-built but rough mortgage calculator site; website-design.com with functional tools (paste boards, mind maps, colour tools) but poor polish; framework scaling claim "several thousand agents". Useful ground truth for verifying which case-study sites actually functioned.
- **sources:** docs018_rerendering/008b_my_notes_what_I_can_do
- **relations:** site-case-studies (leopardess, vet, wyke); vet-med-pricing; dynamic-applications (website-design.com tools).
- **verify-later:** sites table rows for each named domain.

<!-- SOURCE: U22_recent_small_docs.md -->
### Verticals designed (revenue models + knowledge clusters)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** Revenue projection tables labelled "months 12-18" with market data "verified through research"; no live-site revenue claimed.
- **what:** Five verticals worked out with specific knowledge clusters, source lists, page-type libraries, monetisation, and 24-month revenue projections: veterinary/vetcomparison.uk (insurance affiliate £15-35 + listings, £1,960-7,875/mo), energy/gaswholesalers.com (qualified leads £30-60, £1,250-5,350), finance_mortgage/mortgagecalculator.co.uk (broker leads £50-150, £16,500-44,000 — highest value), seasonal_gifts/xmaspresents.com (affiliate 3-17%), plus a "sell the domain not develop" premium pathway (design.co.uk £20-100k).
- **sources:** docs021.../020_vertical_cluster_architecture.md#3, docs021.../025_session_handoff_vertical_architecture.md#verticals-designed, docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture, domain content strategy framework, premium-domain pathway
- **verify-later:** vertical_registry monetisation_config; any live vetcomparison/gaswholesalers/mortgagecalculator sites

<!-- SOURCE: U22_recent_small_docs.md -->
### Domain content strategy framework (15-question)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "For the content generation pipeline, the 15-question framework should feed into the briefing/research phase" — prescriptive/should, not implemented.
- **what:** A systematic three-layer, 15-question methodology for deciding what content a domain needs to compete: Layer 1 (who visits, intent, satisfaction, money flow), Layer 2 (competitor pages, buying journey, real questions, bookmarkable hook), Layer 3 (best page on the topic, original element, format, next action). Worked examples for gaswholesalers.com and vetcomparison.uk with verified lead/affiliate rates. Questions 5-7 require real competitive research.
- **sources:** docs022.../001_domain_content_strategy_framework.md, docs022.../002_domain_content_strategy_framework_v2.md
- **relations:** domain-strategist prompt, deep research domain authority, site classifier
- **verify-later:** domain-strategist agent prompt; briefing/research phase incorporating the framework

<!-- SOURCE: U22_recent_small_docs.md -->
### Deep research domain authority strategy
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "The multi-cluster knowledge base approach lets you build that deep knowledge layer for any domain" — strategy doc, canine project cited as proof-of-concept not production.
- **what:** The thesis that content wins on E-E-A-T by synthesising primary/authoritative sources (BSAVA, Ofgem, PRA/FCA, swap-rate data) into knowledge consumers can't easily find, rather than rephrasing published synthesis. A repeatable 6-step pipeline (niche mapping → primary-source identification → multi-cluster KB construction → gap identification → content architecture → generation from KB) creates a defensible moat: depth consistency, cross-cluster synthesis, and update efficiency competitors can't copy by rewriting one article.
- **sources:** docs022.../003_deep_domain_research_authority.md
- **relations:** vertical knowledge architecture, domain content strategy framework, canine biology KB
- **verify-later:** research-agent primary-source handling; knowledge_base source_authority weighting

<!-- SOURCE: U22_recent_small_docs.md -->
### Content-site valuation model (24-32x)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "using current market multiples of 24-32x monthly profit (Empire Flippers averaging 24x, premium 30-35x)" — used throughout as a projection basis.
- **what:** The valuation basis underpinning the domain portfolio strategy: content/affiliate sites sell at ~24-32x monthly profit, so a £1,500-3,000/mo site is worth ~£36k-96k. Combined with verified per-niche lead/affiliate rates to project each domain's asset value and justify the knowledge-base investment. The portfolio is framed as the testing ground toward a £25k+ annual revenue target and a two-tier service→pipeline-sale model.
- **sources:** docs022.../002_domain_content_strategy_framework_v2.md#monetisation, docs021.../025_session_handoff_vertical_architecture.md#market-data-verified
- **relations:** verticals designed, commercial model (chatbot docs), premium-domain pathway
- **verify-later:** n/a (business assumption)

<!-- SOURCE: U22_recent_small_docs.md -->
### Building-and-hosting-as-a-service via chat (recursive platform)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Recorded because it sharply reframes the satellite ... (Discussion artefact; revisit as it firms up.)"
- **what:** A worked example where a chat box on design.co.uk becomes the intake+orchestration front-end to the whole build platform offered as a service: conversational briefing (a briefing-agent interview replacing the static form) → satellite intake orchestrator → full build workflows on the satellite → hybrid S3+lambda hosting → the new site itself gets a chatbot (recursion). Requires the full chassis on the satellite (rules out Option X for this use case) and surfaces new SaaS concerns: cost/abuse gating, accounts/billing/quotas, feeding reusable building blocks one-directionally from core.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#12
- **relations:** isolated chat environment (Y-copy), briefing-agent/intake orchestrator, commercial model, simple paid multi-domain chat
- **verify-later:** briefing-agent conversational mode; satellite intake orchestrator

<!-- SOURCE: U22_recent_small_docs.md -->
### Payable-differentiator framework (asset × AI capability)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "This is the method ... It is a starting point and needs more thought and testing — the menus below are incomplete, the scoring is rough."
- **what:** A method for justifying a paid chat when there's no proprietary data: value comes not from the model (everyone has it) but from a hard-to-reproduce asset (proprietary/paid data feed, an owned process/output, a well-built tool, a commercial partnership, or early access to a new AI capability) combined with AI for a paying audience. Maintain two menus (assets; AI-capabilities-worth-using-now) and pair one of each per domain. Worked examples: websitedesign.com (package our site-spec/plan as a starter prompt for Bolt/Lovable — strongest), gaswholesalers.com (buy oil/gas data feeds), agritec.uk (partnership vouchers). Prioritise by reproducibility, willingness-to-pay, build cost, cross-domain reuse; test willingness-to-pay before building.
- **sources:** docs025.../PLAN_simple_paid_multidomain_chat(1).md#10
- **relations:** simple paid multi-domain chat, ideation agent, verticals designed
- **verify-later:** n/a (strategy)

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### finetuning.uk self-service RAG product (business strategy)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** BUSINESS_PLAN(1) §1 "A RAG platform for SMEs"; FOCUS(21) §5 "This is a product decision, not a technical one" and §10 shipping ladder with "Aspirational dates, not promises"
- **what:** The plan to turn finetuning.uk into a paid product. Direction pivoted several times: from "self-service fine-tuning SaaS" → RAG-over-your-docs platform with automatic data curation as the named differentiator; from "concierge first, UI later" → UI-first ("build the cockpit we use ourselves"); target user refined up from non-technical owners to technical-adjacent SME ops leads. Pricing £199–1,499/mo + setup fees; £9-12k/mo solo-operator target; £100k year-1 projection.
- **sources:** BUSINESS_PLAN_finetuning_uk(1).md#1, #5, #7; flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#5, #7, #8, #10
- **relations:** reuses internal-flywheel infra (Ollama, Unsloth, rag_index); replacement direction for the abandoned "self-serve fine-tuning" pitch
- **verify-later:** finetuning.uk site_specs; tenant_id on knowledge_base (not yet built)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Moat / differentiator framework (asset × AI × audience-that-pays)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "A payable idea = asset × AI × audience-that-pays. Five asset types: proprietary data, owned process/output, well-built tool, partnership, early-mover timing on a new capability."
- **what:** A doctrine used to filter which domains/products are worth building: the AI model is never the differentiator, the asset it's applied to is. Honest verdict reached during idea.uk's own self-analysis: its moat is "effort + freshness + integration... sustained by maintenance, not a static asset," not a structural moat. Cross-domain pattern discovered via repeated method runs: "wherever the underlying product has high margin, the seller already gives expert support away free" (Bloomberg/Refinitiv, Open Bionics, Robotiq) — an almost-automatic cut for "help-you-buy-X" candidates.
- **sources:** `running_notes(44).md` ("The differentiator framework", "Moat analysis (idea.uk)", "New cross-domain pattern: high-margin-product sellers bundle support free")
- **relations:** idea generation method
- **verify-later:** n/a (a design doctrine, not code)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Five-layer consolidation model (L0–L5)
- **category:** business-strategy
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md` "Consolidation map written... wrote CONSOLIDATION_where_it_all_fits.md — five layers."
- **what:** A planning frame reconciling the idea.uk work with the wider platform roadmap: L0 chassis (exists, builds static sites), L1 idea engine (built, standalone: method + internal CLI + idea.uk), L2 idea.uk product (in progress, first to go live), L3 vertical tools (in progress, chassis-native, e.g. SFI26 Diff Alerts), L4 tool-rich site building for any domain (future — the idea engine becomes a *planning input* to the chassis site builder), L5 automated VM backend deployment (future — today's pipeline only deploys static→B2; provisioning+deploying a persistent backend is the gap Thunder is the seed of). States the natural build order is "prove L1 → ship L2 → build L3 once → generalise into L4 → grow L5 from Thunder."
- **sources:** `running_notes(44).md` ("Consolidation map written")
- **relations:** service-deployer pattern (= the L5 gap); chassis-native idea engine (= L4); idea.uk product (= L2)
- **verify-later:** `CONSOLIDATION_where_it_all_fits.md` (live doc, out of this unit's scope)

<!-- SOURCE: U24f_docs_archive_remaining_small.md -->
### SaaS commercial model — operator-primary, vendor-optional, with entitlement seams
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** "Resolved direction" language for the model itself, but the concrete implementation ("Design the seam now... build the billing depth later") is explicitly deferred — "None of these constrain the build as long as separability and the seams above are honoured."
- **what:** A commercial model for operating the platform at scale while keeping individual sites/backends sellable: **operator-primary, vendor-optional** — operate thousands of domains directly, with the option to sell a domain plus its backend (the common case) or, rarely, the whole framework/instance. The key structural insight is that the unit of blast-radius isolation (the satellite/cluster) and the unit of sale-separability (the domain) are different granularities — operating thousands of domains does not require thousands of clusters; it requires clean per-domain partitioning (keyed on `site_id`/`domain`) plus the ability to extract one domain's artifact, data, and credentials at sale time. Five seams are flagged as cheap now / expensive to retrofit: `owner_id` on site rows; an entitlement check at both build-submission and maintenance-run (never calling Stripe directly — always through a pluggable billing-adapter interface); credential parameterization everywhere (no hardcoded keys); and a build-tier/cost-profile flag (`saas_cheap` vs `portfolio`) driving cheaper model/batching choices so low-price builds retain margin.
- **sources:** docs/_archive/agent_docs/docs025_ai_chatbot_idea_uk/excellent_discussions/PLAN_isolated_chat_environment(3).md §13
- **relations:** isolated chat satellite architecture (above, same document)
- **verify-later:** whether an owner_id/entitlement layer or billing adapter exists anywhere in the schema/codebase today

<!-- SOURCE: U25_leopardess_social.md -->
### Data-sovereignty / pilot-first / startup fast-start commercial positioning
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK H6/H8/H9 all "resolved 2026-07-10" as *drafted positioning*, each explicitly "Not yet done for a client".
- **what:** Three owner-confirmed engagement angles drafted into specs/portfolio.json use_cases with honest labelling: (A8) data-sovereignty as a capability built *with* a client during a scoped engagement, never a standing guarantee; (H6) pilot-first engagement ladder (bounded fixed-price pilot → licence/day-rate/retainer decided by what the pilot reveals); (H9) startups building agent products start from the platform's already-solved operational layer (state, retries, HITL, no-redeploy workflow changes). Plus register-reconciliation generalisation ("your list vs an authoritative register" as the general shape of the Companies House work).
- **sources:** docs/leopardessconsulting/specs/portfolio.json#use_cases; docs/leopardessconsulting/RUNBOOK.md#H6-H9; docs/leopardessconsulting/RUNNING_NOTES.md#Turn-6, #Turn-7
- **relations:** per-step model routing; no-tenant-isolation; claim-evidence audit rule
- **verify-later:** site_specs aspect 'portfolio' current row

<!-- SOURCE: U25_leopardess_social.md -->
### UK-sovereign stack exploration (deferred)
- **category:** business-strategy
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK Reference: "explicitly deferred to a separate chat by the owner, 2026-07-10 … Do not start this unprompted."
- **what:** Future exercise: a fully UK-hosted compute+storage+model stack. Baseline facts captured so the future thread doesn't re-derive them: compute Rackspace UK; storage Backblaze us-east-005 (US); Anthropic and Google models US; self-hosted path sits wherever the cluster is.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#Reference; docs/leopardessconsulting/AUDIT_verified_facts.md#4b-P6
- **relations:** data-sovereignty positioning; storage-architecture
- **verify-later:** memory `uk-sovereign-stack-exploration`

<!-- SOURCE: U26_misc_dirs.md -->
### WordPress export agent and content-subscription plugin
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 008 designs it enthusiastically ("This is GENIUS!"); 009 immediately deconstructs it ("You're not unique in 'AI builds WordPress sites'. That market is saturated... Only add WordPress if they're begging for it") — never mentioned again anywhere.
- **what:** An agent converting generated HTML sites into installable WordPress themes + WXR content exports, paired with a WP plugin subscribing to the platform for auto-published fresh content (recurring revenue). Explicitly killed by the competitive analysis in 009, which redirected differentiation toward "sites that update themselves" / continuous content ecosystems.
- **sources:** docs/architecture/008-start-with-plain-old-html-js-css-to-wordpress.md#wordpress-export-agent-design; docs/architecture/009-wordpress-discussion#the-hard-truth
- **relations:** HTML-first delivery; living-content differentiation (which the platform did pursue via news feed / content pipelines)
- **verify-later:** n/a (never built)

<!-- SOURCE: U26_misc_dirs.md -->
### Domain value maximisation pipeline (domain flipping)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 010/011/015 lay out the strategy against the user's real domain portfolio (collateralfinancing.com, holidaytime.com, websitedesign.com...) with 48-hour development timelines; no later doc pursues domain flipping — the platform pivoted to operating its own sites.
- **what:** Use the agent platform to develop parked domains into sites with content, traffic and revenue to multiply sale value (naked $500 → revenue-bearing $10k+): domain classification (brandable/exact-match/local/product), tiered portfolio treatment, 48h batch development, monetisation setup (leads/affiliate/ads), and "self-selling" footers that market the build service from every developed domain.
- **sources:** docs/architecture/010-domain-value-maximisation.md; docs/architecture/011-example-domains; docs/architecture/015-underserved-niche.md#your-domain-portfolio-is-your-marketplace
- **relations:** deep-research domain insight agent; underserved-niche strategy; site-case-studies (the surviving practice of operating exemplar sites)
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### Underserved-niche and vertical showcase strategy
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 015/016 propose niches (compliance docs, local-business packages, academic assistants, affiliate content) and per-industry showcase domains with a workflow marketplace ("DIY $500 / DFY $200mo / White Label $2000mo"); positioning discussion only, no implementation trail.
- **what:** Rather than competing with Temporal/LangChain/Zapier broadly, own narrow verticals where multi-agent coordination wins: each showcase domain demos an industry solution (legal docs, restaurant launch, real-estate listings) and funnels to purchasable workflows. Includes the pricing-tier and "Business-in-a-Box" (site + content pipeline + email + social) framings, and the investor-demo positioning of the framework as the star with swappable use cases.
- **sources:** docs/architecture/015-underserved-niche.md; docs/architecture/016-competitive-advantge.md#who-actually-pays-for-ai-sites; docs/architecture/012-investors.md#the-portfolio-approach
- **relations:** domain value maximisation; EBORG organizational OS
- **verify-later:** n/a

<!-- SOURCE: U26_misc_dirs.md -->
### AI-native orchestration positioning (vs Temporal/Airflow)
- **category:** business-strategy
- **status-signal:** abandoned
- **status-evidence:** 012/014 are interview/investor argumentation ("You could build this on Temporal, but it would be like using Kubernetes to run a single container"); the accompanying Temporal/Airflow adapter agents were never built.
- **what:** The articulated build-vs-buy rationale for the platform: AI-specific needs (dynamic JSON workflows without deployment, token/fuel tracking, prompt management, AI-failure handling, multi-tenant agent isolation, workflows spawning from AI decisions) justify a purpose-built orchestrator. Includes proposed Temporal-adapter and Airflow-adapter agents to bridge into enterprise workflow estates as a migration path ("we don't replace your existing systems, we enhance them").
- **sources:** docs/architecture/012-investors.md#better-answer; docs/architecture/014-Temporal-Airflow-adapters.md
- **relations:** adapters (current adapter guide is a different, real lineage); distributed embedded orchestration
- **verify-later:** confirm no temporal/airflow adapter code exists

<!-- SOURCE: U26_misc_dirs.md -->
### EBORG — Evidence-Based Organisational Planning
- **category:** business-strategy
- **status-signal:** unknown
- **status-evidence:** The Nov 2025 HITL demo is branded "For EBORG" with a full pitch paragraph in start_hitl_workflow.sh ("mapping every role, responsibility, and objective, then pairing each with a framework of AI agents"); no other doc in this unit elaborates whether it became a product.
- **what:** A product concept: organisations map roles/responsibilities/objectives and pair each with AI agents that gather research, assess options and provide evidence-based reasoning — "a human-centered, continuously learning organisation" and the concrete descendant of 016's Organizational OS idea. Used as the demo business in the HITL content-approval workflow.
- **sources:** docs/humanintheloop/start_hitl_workflow.sh; docs/humanintheloop/hitl_agent_definition.sql (header comment); docs/architecture/016-competitive-advantge.md#the-organizational-os-concept
- **relations:** cross-domain intelligence network; HITL content-approval demo
- **verify-later:** any EBORG references in business/vonc docs (other units)

<!-- SOURCE: U04_idea_uk.md -->
### Ideation method v0→v3 (staged, multi-model, web-verified pipeline)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Method doc carries v1/v2/v3 patches with dated rationale; engine validated end-to-end live 2026-06-04; live product runs it per paid order.
- **what:** The runnable method: (1) frame AND challenge the audience; (2) generate candidates across four lenses — demand, generalist-failure, frontier, outcome — plus the original asset×capability sweep; (3) cut each candidate against the *specific named free substitute* with a different model, incl. the seller-bundles-support-free check; (4) verify survivors with web research, evidence attached; (5) score; (6) rank and split test-now vs consider. Version history is itself conceptual: v1 added Durability + the specific-substitute cut; v2 added multi-lens generation + audience-fit challenge (single-lens generation diagnosed as "narrow — supply-side only"); v3 added the Risk column. Dogfooding rule: if the method can't find an advancing candidate for idea.uk itself, it isn't good enough.
- **sources:** idea.uk/idea_uk_method_v0(3).md; idea.uk/idea_uk_testrun_v2.md; idea.uk/KEY_DOC_idea_method_prompt.md; idea.uk/running_notes(63).md (method-run checkpoints)
- **relations:** operator-risk column; cross-vendor critique; capability watchlist; engine implementations.
- **verify-later:** golang_files/prompts.go step prompts vs the method doc.

<!-- SOURCE: U04_idea_uk.md -->
### Operator-risk column: hazard scored separately from fitness, with gates
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Debugging-guide item 23 (2026-05-28) documents the addition end-to-end; A1 acceptance notes the live report carries "Operator risk: N/5" with auto-flags.
- **what:** A sixth scored dimension, Risk to the operator (1–5, 5 safest), scoring the CONSEQUENCE of being wrong, not the probability. Deliberately **not** added to the fitness sum and **not** in the Def≥3∧Will≥3 gate: Risk=1 (regulated professions) is dropped automatically into a visible "Dropped for operator risk" section; Risk≤2 advances but flagged "needs liability work before building" with the cheapest_test forced to demand PII + solicitor-reviewed T&Cs first; Risk breaks ties toward safer builds. Generalisable lesson: when a scoring system recommends actions to an operator who carries downstream exposure, hazard must be a separate scored dimension — fitness sums cannot see it. First real effect: paused the SFI single-farm assessment.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md (item 23); idea.uk/idea_uk_method_v0(3).md (Risk rubric); idea.uk/LIABILITY_AND_TERMS(2).md (header)
- **relations:** liability framework; SFI26 Diff Alerts swap; ideation method.
- **verify-later:** engine.go `scored` struct + riskNote; idea_method_runner.py parity.

<!-- SOURCE: U04_idea_uk.md -->
### Capability watchlist + real-world event-window watchlist
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** PLAN_idea_uk §8: "the capability watchlist runs as its own recurring research workflow" (stance accepted); no evidence either watchlist was ever built as a workflow.
- **what:** Two maintained lists that feed re-runs of ideation: (1) AI capabilities worth using now, grouped by what specialism does that generalists don't — the mechanism for being early to ideas a new capability just unlocked, and the single strongest durable advantage; (2) real-world event windows per domain (scheme deadlines, regulation changes — e.g. SFI26 Window 1), because timing changes what's actionable. The capability menu v1 ships inside the method/prompts; the *recurring maintenance workflows* remain unbuilt.
- **sources:** idea.uk/idea_uk_method_v0(3).md (capability list v1); idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md ("Watchlist should track scheme/event windows")
- **relations:** differentiator framework (currency moat); scheduler-and-tasks (would host the recurrence).
- **verify-later:** scheduled_tasks / agent_definitions for any watchlist workflow.

<!-- SOURCE: U04_idea_uk.md -->
### Cross-vendor critique (the cut step on a different vendor)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Architecture doc §9: `[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic` stderr line added; env-switched via OPENAI_API_KEY.
- **what:** The method's quality gate (the cut) is run by a different model from the generator — ideally a different **vendor** — so the method isn't one model marking its own work. Implemented as an optional OpenAI branch on the cut step (OPENAI_API_KEY + OPENAI_CRITIQUE_MODEL); same-vendor fallback still uses a different model (Sonnet vs Opus). Cross-vendor comparison flagged as an untested open experiment.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#9; idea.uk/idea_uk_method_v0(3).md (diff-model markers); idea.uk/RUNBOOK_idea_uk(9).md (go-live step 5)
- **relations:** ideation method; multi-model ensemble moat claim.
- **verify-later:** engine.go cut-step branch.

<!-- SOURCE: U04_idea_uk.md -->
### Engine implementations: single-shot prompt → Python runner → Go engine (with LLM feature upgrade)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** "Ported the idea.uk tooling from Python to Go (platform is Go throughout)" (running notes 2026-05-28ff); A1 DONE 2026-05-28 with live validation 2026-06-04; python files retained as superseded parity copies.
- **what:** Three coexisting expressions of the method: a paste-anywhere single-shot prompt (weakest — one model marks its own work), the Python `idea_method_runner.py`/`idea_service.py` originals (superseded), and the shipped Go engine (`engine.go`+`prompts.go`, stdlib-only, no SDKs, offline `GOPROXY=off` build). The A1 upgrade set the model line-up (Opus for generate/verify, Sonnet for cut/score, all env-overridable) and added extended thinking per step (off for brainstorm breadth), prompt caching on static system blocks, `web_search_20260209` + code-execution filtering on verify, and a `WEB_SEARCH_MAX_USES` budget (raised 6→12 after a quota-exhausted run left premises "provisional").
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1; idea.uk/golang_files/engine.go (header); idea.uk/python_files/idea_method_runner.py (header); idea.uk/RUNBOOK_idea_uk.md (base — Python era, family-delta)
- **relations:** LLM API shape disciplines (the three bugs found during validation); ideation method.
- **verify-later:** engine.go callClaudeOpts / usesAdaptiveThinking.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk service: request-then-confirm flow, REVIEW_BEFORE_PAY, AUTO_DELIVER, capacity cap
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Live and earning; full flow proven end-to-end with a real card 2026-06-14 ("LIVE BUG RESOLVED: paid + report delivered end-to-end").
- **what:** The order state machine: visitor `/request` (free) → operator confirm/decline → pay → fulfil. Two switchable shapes: charge-first (engine runs after payment; AUTO_DELIVER=false holds the report for operator review) and **review-before-pay** (default from 2026-06-11: confirm runs the engine first, operator reviews the draft, `/approve` sends the pay link — money is only taken after the operator has seen the deliverable). `MAX_ACTIVE_ORDERS` caps in-flight orders so capacity can't be oversold; `/capacity` exposes it. Orders live in a JSON file store (`/var/lib/idea/orders.json`) — deliberately no DB on the exposed box.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (flow + 2026-06-11 update); idea.uk/golang_files/service.go (header); idea.uk/idea_uk_architecture_and_deployment(6).md#5
- **relations:** Stripe webhook truth; B2 dead-drop persistence design (future DB); liability framework (operator review as mitigation).
- **verify-later:** service.go state machine + service_test.go (19+ checks).

<!-- SOURCE: U04_idea_uk.md -->
### Free audience-check taster endpoint
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** A2 DONE 2026-05-28 with acceptance ticks; live on the page; taster now logs the result (2026-06-11 checkpoint b).
- **what:** `/audience-check` — the method's step 1 (audience challenge + 2–3 alternative audiences) exposed as a free, no-auth, ~£0.02/run, ~10s taster: the conversion hook that replaced voluntary-pay. Per-IP sliding-window rate limiting (3/h, 20/day) with Retry-After; XSS-escaped HTML fragment for direct innerHTML insertion; TASTER_ENABLED kill switch; each run logs business/audience/result as market intelligence.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A2; idea.uk/golang_files/audience_check.go (header); idea.uk/idea_uk_open_discussion.md#5
- **relations:** voluntary-pay rejection; ideation method step 1.
- **verify-later:** audience_check.go limiter + tests.

<!-- SOURCE: U04_idea_uk.md -->
### Click-through operator approval links (HMAC per-order tokens)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Checkpoint 2026-06-11 (f)/(g): built, then "click-through confirmed working by user".
- **what:** Request/review emails carry links to a page with Confirm/Approve/Decline buttons. The link carries an HMAC(order id, INTERNAL_API_KEY) token authorising that one order only; the link opens a **safe GET page** (mail-scanner prefetch can't trigger anything) and the action fires only on a button POST; actions stay gated by order status so a token can't double-fire. Curl + X-Internal-Key remains the fallback.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (2026-06-11 update); idea.uk/running_notes(63).md (checkpoints f–g)
- **relations:** request-then-confirm flow; hitl (same shape as approval flows).
- **verify-later:** service.go token mint/verify.

<!-- SOURCE: U04_idea_uk.md -->
### Fake-door → intent-capture-first launch discipline
- **category:** NEW:idea-product
- **status-signal:** superseded
- **status-evidence:** PLAN §7 step 4 ("intent capture first, no payment"); superseded by the live request-then-confirm flow with real Stripe (the fakedoor page became the embedded landing page).
- **what:** Launch pattern: a static landing page offering the report at a flat price, capturing intent without charging ("we reply within 24h with a confirmed slot + payment link, or a polite decline") — deliberately avoiding charge-then-fail refund overhead — with a visible monthly slot count to throttle demand to manual capacity. Also prescribed as a parallel track for the strongest single-domain candidate (agritec SFI26 checker). The page evolved into the embedded `page.html` of the live service.
- **sources:** idea.uk/PLAN_idea_uk(3).md#7; idea.uk/running_notes(63).md ("Built the idea.uk fake-door page", "Fake-door modified to intent-capture-only"); idea.uk/idea_uk_fakedoor(9).html (deployment notes header)
- **relations:** request-then-confirm flow (its successor); demand-test philosophy in the method's cheapest_test.
- **verify-later:** n/a (historical).

<!-- SOURCE: U04_idea_uk.md -->
### Deliverable quality standards for reports and product emails
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** BUGS_idea_uk 2026-06-11 entries all marked "Fixed this build" with standing "for future builds" rules.
- **what:** Standing rules distilled from report-email review: every customer-facing string in plain English for a non-technical owner (jargon/acronyms treated as defects); every standalone deliverable opens with a one-paragraph plain summary of what it is; rejected options always say what the thing was and why it died; deliverables get a deliberate professional design distinct from marketing surfaces (the £29 report email: navy/gold/serif "sheet" look, unlike the landing page); illustrative examples must not leak into generated output (audience-anchored generation). Transport rule: any HTML email must be base64/quoted-printable encoded (the SMTP 998-octet line-fold corrupted raw HTML mid-tag).
- **sources:** idea.uk/BUGS_idea_uk(4).md; idea.uk/RUNBOOK_idea_uk(9).md ("HTML emails are base64-encoded")
- **relations:** content-quality (platform analogue); transactional email realities.
- **verify-later:** service.go b64Body; report HTML renderer.

<!-- SOURCE: U04_idea_uk.md -->
### Chassis-native idea engine (Phase D `idea-orchestrator`)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK Phase D "Not started yet… needs a schema pass first"; architecture doc §8 declines to write the SQL until the action contracts are read.
- **what:** The second way to run the method: as a chassis agent + workflow reusing existing actions almost 1:1 (execute_llm_prompt for frame/generate/cut/score, web_search for verify, HITL actions for the operator gate, store_result/write_site_spec for persistence) — for running the method internally across our own domains on schedule (the Layer-4 planning input). The billing half deliberately stays in the standalone service ("a product/payment concern, not an agent workflow"). Bundle 2 packages exactly this port task.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#8; idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase D); idea.uk/BUNDLE_2_chassis_idea_engine_workflow.md
- **relations:** five-layer stack (L4); development-guide conventions (every agent an orchestrator; spawn sub-agents).
- **verify-later:** agent_definitions for any idea-orchestrator; the docubundle context file.

<!-- SOURCE: U04_idea_uk.md -->
### Multi-tenant branded intake pages on one central engine (white-label Option C)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** open_discussion §7: Option C "RECOMMENDED… Want me to do this in the next round?" — never built.
- **what:** Other sites offer the ideation product via their own branded static request page (built through the normal pipeline, own price/copy) POSTing to the central service with a tenant_id; per-tenant Stripe branding; iframe and CNAME/reverse-proxy options analysed and rejected. Needs ~100–200 lines (tenant field on Order, tenants config, tenant-aware /request). Shape A (site IS the service) vs Shape B (request panel on a content site) hosting split defined in the architecture doc; a forked-component tool is explicitly the wrong model for a server-side paid engine — sites only ever *link* to it.
- **sources:** idea.uk/idea_uk_open_discussion.md#7; idea.uk/idea_uk_architecture_and_deployment(6).md#7
- **relations:** tool-library boundary (why the engine is not a content_component); site_plan blocked/planned mechanism.
- **verify-later:** service.go for any tenant handling (expect none).

<!-- SOURCE: U04_idea_uk.md -->
### Real-door streaming progress page + programmatic refund endpoint (Phase A3/A4)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK A3/A4 have outputs+acceptance defined but no DONE mark; refunds confirmed manual-only in the Stripe section ("There is no refund code").
- **what:** A3: post-payment page polls `/status/{order_id}` and renders live engine progress ("generating… cutting… verifying claim 1 of N"), report renders in-browser — the "real door" UX (option (a) of the real-door analysis; the honest 72h email model shipped instead). A4: operator-gated `/refund` calling Stripe POST /v1/refunds and marking the order refunded — refunds today are manual dashboard clicks and the app doesn't record them.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A3-A4; idea.uk/idea_uk_open_discussion.md#3-4; idea.uk/RUNBOOK_idea_uk(9).md (Refunds — manual)
- **relations:** request-then-confirm flow; Stripe pattern.
- **verify-later:** service.go routes (expect no /status, /refund).

<!-- SOURCE: U04_idea_uk.md -->
### SFI26 Diff Alerts (first vertical tool) — replacing the single-farm assessment
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** "Tool swapped 2026-05-28… paused on liability grounds"; Phase C fully specified (C1–C5) with no build evidence; the base DEVELOPMENT_RUNBOOK still carries the original single-farm Phase C (family-delta capture of the abandoned product).
- **what:** The first Layer-3 vertical tool: a subscription digest for UK farm advisors summarising what changed in Defra/RPA SFI26 guidance, from a versioned scraped corpus, with every change cited to source+version, weekly, operator-reviewed for 8 issues before auto-send. Scored 19/25 with Risk 4. It replaced the **SFI26 single-farm assessment** (abandoned/backlogged: Risk 2 — a wrong number could cost a farmer £5–50k), the first product decision the Risk column changed. Chassis-native by design (recurring, per-user state, scheduled), the opposite plumbing to standalone idea.uk.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase C + swap note); idea.uk/DEVELOPMENT_RUNBOOK.md (base — original single-farm Phase C); idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 3)
- **relations:** operator-risk column; liability framework (SFI T&Cs draft); vet-med-pricing (sibling scraping shape).
- **verify-later:** any SFI corpus/agent in the repo or DB (expect none).

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk standalone service page-serving and deploy gotchas
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Debugging guide §11 added for the idea.uk service; each gotcha tied to a fixed live incident.
- **what:** The operational failure catalogue for the single-binary service: every served path needs an explicit mux handler (bare 404s on linked pages); `writeHTML` fragments vs the `a.page()` full-page brand wrapper (navigation targets must wrap; injected fragments must not); startup templating of `CONTACT_EMAIL`/`MONTH_SLOTS` placeholders; systemd EnvironmentFile keeps inline comments (crash-loop + nginx 502); certbot failure made non-fatal in setup.sh; replace a running binary by scp-to-temp + `mv -f` (text-file-busy); Let's Encrypt rejects placeholder emails.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md#11; idea.uk/golang_files/README_setup_SETUP.md; idea.uk/BUGS_idea_uk(4).md (mobile safe-area padding)
- **relations:** setup.sh; VM launch plan.
- **verify-later:** service.go routes() vs page.html hrefs.

<!-- SOURCE: U04_idea_uk.md -->
### Ideation method v0→v3 (staged, multi-model, web-verified pipeline)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Method doc carries v1/v2/v3 patches with dated rationale; engine validated end-to-end live 2026-06-04; live product runs it per paid order.
- **what:** The runnable method: (1) frame AND challenge the audience; (2) generate candidates across four lenses — demand, generalist-failure, frontier, outcome — plus the original asset×capability sweep; (3) cut each candidate against the *specific named free substitute* with a different model, incl. the seller-bundles-support-free check; (4) verify survivors with web research, evidence attached; (5) score; (6) rank and split test-now vs consider. Version history is itself conceptual: v1 added Durability + the specific-substitute cut; v2 added multi-lens generation + audience-fit challenge (single-lens generation diagnosed as "narrow — supply-side only"); v3 added the Risk column. Dogfooding rule: if the method can't find an advancing candidate for idea.uk itself, it isn't good enough.
- **sources:** idea.uk/idea_uk_method_v0(3).md; idea.uk/idea_uk_testrun_v2.md; idea.uk/KEY_DOC_idea_method_prompt.md; idea.uk/running_notes(63).md (method-run checkpoints)
- **relations:** operator-risk column; cross-vendor critique; capability watchlist; engine implementations.
- **verify-later:** golang_files/prompts.go step prompts vs the method doc.

<!-- SOURCE: U04_idea_uk.md -->
### Operator-risk column: hazard scored separately from fitness, with gates
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Debugging-guide item 23 (2026-05-28) documents the addition end-to-end; A1 acceptance notes the live report carries "Operator risk: N/5" with auto-flags.
- **what:** A sixth scored dimension, Risk to the operator (1–5, 5 safest), scoring the CONSEQUENCE of being wrong, not the probability. Deliberately **not** added to the fitness sum and **not** in the Def≥3∧Will≥3 gate: Risk=1 (regulated professions) is dropped automatically into a visible "Dropped for operator risk" section; Risk≤2 advances but flagged "needs liability work before building" with the cheapest_test forced to demand PII + solicitor-reviewed T&Cs first; Risk breaks ties toward safer builds. Generalisable lesson: when a scoring system recommends actions to an operator who carries downstream exposure, hazard must be a separate scored dimension — fitness sums cannot see it. First real effect: paused the SFI single-farm assessment.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md (item 23); idea.uk/idea_uk_method_v0(3).md (Risk rubric); idea.uk/LIABILITY_AND_TERMS(2).md (header)
- **relations:** liability framework; SFI26 Diff Alerts swap; ideation method.
- **verify-later:** engine.go `scored` struct + riskNote; idea_method_runner.py parity.

<!-- SOURCE: U04_idea_uk.md -->
### Capability watchlist + real-world event-window watchlist
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** PLAN_idea_uk §8: "the capability watchlist runs as its own recurring research workflow" (stance accepted); no evidence either watchlist was ever built as a workflow.
- **what:** Two maintained lists that feed re-runs of ideation: (1) AI capabilities worth using now, grouped by what specialism does that generalists don't — the mechanism for being early to ideas a new capability just unlocked, and the single strongest durable advantage; (2) real-world event windows per domain (scheme deadlines, regulation changes — e.g. SFI26 Window 1), because timing changes what's actionable. The capability menu v1 ships inside the method/prompts; the *recurring maintenance workflows* remain unbuilt.
- **sources:** idea.uk/idea_uk_method_v0(3).md (capability list v1); idea.uk/PLAN_idea_uk(3).md#8; idea.uk/running_notes(63).md ("Watchlist should track scheme/event windows")
- **relations:** differentiator framework (currency moat); scheduler-and-tasks (would host the recurrence).
- **verify-later:** scheduled_tasks / agent_definitions for any watchlist workflow.

<!-- SOURCE: U04_idea_uk.md -->
### Cross-vendor critique (the cut step on a different vendor)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Architecture doc §9: `[cut] cross-vendor: OpenAI (gpt-4o)` / `[cut] same-vendor: Anthropic` stderr line added; env-switched via OPENAI_API_KEY.
- **what:** The method's quality gate (the cut) is run by a different model from the generator — ideally a different **vendor** — so the method isn't one model marking its own work. Implemented as an optional OpenAI branch on the cut step (OPENAI_API_KEY + OPENAI_CRITIQUE_MODEL); same-vendor fallback still uses a different model (Sonnet vs Opus). Cross-vendor comparison flagged as an untested open experiment.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#9; idea.uk/idea_uk_method_v0(3).md (diff-model markers); idea.uk/RUNBOOK_idea_uk(9).md (go-live step 5)
- **relations:** ideation method; multi-model ensemble moat claim.
- **verify-later:** engine.go cut-step branch.

<!-- SOURCE: U04_idea_uk.md -->
### Engine implementations: single-shot prompt → Python runner → Go engine (with LLM feature upgrade)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** "Ported the idea.uk tooling from Python to Go (platform is Go throughout)" (running notes 2026-05-28ff); A1 DONE 2026-05-28 with live validation 2026-06-04; python files retained as superseded parity copies.
- **what:** Three coexisting expressions of the method: a paste-anywhere single-shot prompt (weakest — one model marks its own work), the Python `idea_method_runner.py`/`idea_service.py` originals (superseded), and the shipped Go engine (`engine.go`+`prompts.go`, stdlib-only, no SDKs, offline `GOPROXY=off` build). The A1 upgrade set the model line-up (Opus for generate/verify, Sonnet for cut/score, all env-overridable) and added extended thinking per step (off for brainstorm breadth), prompt caching on static system blocks, `web_search_20260209` + code-execution filtering on verify, and a `WEB_SEARCH_MAX_USES` budget (raised 6→12 after a quota-exhausted run left premises "provisional").
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A1; idea.uk/golang_files/engine.go (header); idea.uk/python_files/idea_method_runner.py (header); idea.uk/RUNBOOK_idea_uk.md (base — Python era, family-delta)
- **relations:** LLM API shape disciplines (the three bugs found during validation); ideation method.
- **verify-later:** engine.go callClaudeOpts / usesAdaptiveThinking.

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk service: request-then-confirm flow, REVIEW_BEFORE_PAY, AUTO_DELIVER, capacity cap
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Live and earning; full flow proven end-to-end with a real card 2026-06-14 ("LIVE BUG RESOLVED: paid + report delivered end-to-end").
- **what:** The order state machine: visitor `/request` (free) → operator confirm/decline → pay → fulfil. Two switchable shapes: charge-first (engine runs after payment; AUTO_DELIVER=false holds the report for operator review) and **review-before-pay** (default from 2026-06-11: confirm runs the engine first, operator reviews the draft, `/approve` sends the pay link — money is only taken after the operator has seen the deliverable). `MAX_ACTIVE_ORDERS` caps in-flight orders so capacity can't be oversold; `/capacity` exposes it. Orders live in a JSON file store (`/var/lib/idea/orders.json`) — deliberately no DB on the exposed box.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (flow + 2026-06-11 update); idea.uk/golang_files/service.go (header); idea.uk/idea_uk_architecture_and_deployment(6).md#5
- **relations:** Stripe webhook truth; B2 dead-drop persistence design (future DB); liability framework (operator review as mitigation).
- **verify-later:** service.go state machine + service_test.go (19+ checks).

<!-- SOURCE: U04_idea_uk.md -->
### Free audience-check taster endpoint
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** A2 DONE 2026-05-28 with acceptance ticks; live on the page; taster now logs the result (2026-06-11 checkpoint b).
- **what:** `/audience-check` — the method's step 1 (audience challenge + 2–3 alternative audiences) exposed as a free, no-auth, ~£0.02/run, ~10s taster: the conversion hook that replaced voluntary-pay. Per-IP sliding-window rate limiting (3/h, 20/day) with Retry-After; XSS-escaped HTML fragment for direct innerHTML insertion; TASTER_ENABLED kill switch; each run logs business/audience/result as market intelligence.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A2; idea.uk/golang_files/audience_check.go (header); idea.uk/idea_uk_open_discussion.md#5
- **relations:** voluntary-pay rejection; ideation method step 1.
- **verify-later:** audience_check.go limiter + tests.

<!-- SOURCE: U04_idea_uk.md -->
### Click-through operator approval links (HMAC per-order tokens)
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Checkpoint 2026-06-11 (f)/(g): built, then "click-through confirmed working by user".
- **what:** Request/review emails carry links to a page with Confirm/Approve/Decline buttons. The link carries an HMAC(order id, INTERNAL_API_KEY) token authorising that one order only; the link opens a **safe GET page** (mail-scanner prefetch can't trigger anything) and the action fires only on a button POST; actions stay gated by order status so a token can't double-fire. Curl + X-Internal-Key remains the fallback.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (2026-06-11 update); idea.uk/running_notes(63).md (checkpoints f–g)
- **relations:** request-then-confirm flow; hitl (same shape as approval flows).
- **verify-later:** service.go token mint/verify.

<!-- SOURCE: U04_idea_uk.md -->
### Fake-door → intent-capture-first launch discipline
- **category:** NEW:idea-product
- **status-signal:** superseded
- **status-evidence:** PLAN §7 step 4 ("intent capture first, no payment"); superseded by the live request-then-confirm flow with real Stripe (the fakedoor page became the embedded landing page).
- **what:** Launch pattern: a static landing page offering the report at a flat price, capturing intent without charging ("we reply within 24h with a confirmed slot + payment link, or a polite decline") — deliberately avoiding charge-then-fail refund overhead — with a visible monthly slot count to throttle demand to manual capacity. Also prescribed as a parallel track for the strongest single-domain candidate (agritec SFI26 checker). The page evolved into the embedded `page.html` of the live service.
- **sources:** idea.uk/PLAN_idea_uk(3).md#7; idea.uk/running_notes(63).md ("Built the idea.uk fake-door page", "Fake-door modified to intent-capture-only"); idea.uk/idea_uk_fakedoor(9).html (deployment notes header)
- **relations:** request-then-confirm flow (its successor); demand-test philosophy in the method's cheapest_test.
- **verify-later:** n/a (historical).

<!-- SOURCE: U04_idea_uk.md -->
### Deliverable quality standards for reports and product emails
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** BUGS_idea_uk 2026-06-11 entries all marked "Fixed this build" with standing "for future builds" rules.
- **what:** Standing rules distilled from report-email review: every customer-facing string in plain English for a non-technical owner (jargon/acronyms treated as defects); every standalone deliverable opens with a one-paragraph plain summary of what it is; rejected options always say what the thing was and why it died; deliverables get a deliberate professional design distinct from marketing surfaces (the £29 report email: navy/gold/serif "sheet" look, unlike the landing page); illustrative examples must not leak into generated output (audience-anchored generation). Transport rule: any HTML email must be base64/quoted-printable encoded (the SMTP 998-octet line-fold corrupted raw HTML mid-tag).
- **sources:** idea.uk/BUGS_idea_uk(4).md; idea.uk/RUNBOOK_idea_uk(9).md ("HTML emails are base64-encoded")
- **relations:** content-quality (platform analogue); transactional email realities.
- **verify-later:** service.go b64Body; report HTML renderer.

<!-- SOURCE: U04_idea_uk.md -->
### Chassis-native idea engine (Phase D `idea-orchestrator`)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK Phase D "Not started yet… needs a schema pass first"; architecture doc §8 declines to write the SQL until the action contracts are read.
- **what:** The second way to run the method: as a chassis agent + workflow reusing existing actions almost 1:1 (execute_llm_prompt for frame/generate/cut/score, web_search for verify, HITL actions for the operator gate, store_result/write_site_spec for persistence) — for running the method internally across our own domains on schedule (the Layer-4 planning input). The billing half deliberately stays in the standalone service ("a product/payment concern, not an agent workflow"). Bundle 2 packages exactly this port task.
- **sources:** idea.uk/idea_uk_architecture_and_deployment(6).md#8; idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase D); idea.uk/BUNDLE_2_chassis_idea_engine_workflow.md
- **relations:** five-layer stack (L4); development-guide conventions (every agent an orchestrator; spawn sub-agents).
- **verify-later:** agent_definitions for any idea-orchestrator; the docubundle context file.

<!-- SOURCE: U04_idea_uk.md -->
### Multi-tenant branded intake pages on one central engine (white-label Option C)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** open_discussion §7: Option C "RECOMMENDED… Want me to do this in the next round?" — never built.
- **what:** Other sites offer the ideation product via their own branded static request page (built through the normal pipeline, own price/copy) POSTing to the central service with a tenant_id; per-tenant Stripe branding; iframe and CNAME/reverse-proxy options analysed and rejected. Needs ~100–200 lines (tenant field on Order, tenants config, tenant-aware /request). Shape A (site IS the service) vs Shape B (request panel on a content site) hosting split defined in the architecture doc; a forked-component tool is explicitly the wrong model for a server-side paid engine — sites only ever *link* to it.
- **sources:** idea.uk/idea_uk_open_discussion.md#7; idea.uk/idea_uk_architecture_and_deployment(6).md#7
- **relations:** tool-library boundary (why the engine is not a content_component); site_plan blocked/planned mechanism.
- **verify-later:** service.go for any tenant handling (expect none).

<!-- SOURCE: U04_idea_uk.md -->
### Real-door streaming progress page + programmatic refund endpoint (Phase A3/A4)
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** DEVELOPMENT_RUNBOOK A3/A4 have outputs+acceptance defined but no DONE mark; refunds confirmed manual-only in the Stripe section ("There is no refund code").
- **what:** A3: post-payment page polls `/status/{order_id}` and renders live engine progress ("generating… cutting… verifying claim 1 of N"), report renders in-browser — the "real door" UX (option (a) of the real-door analysis; the honest 72h email model shipped instead). A4: operator-gated `/refund` calling Stripe POST /v1/refunds and marking the order refunded — refunds today are manual dashboard clicks and the app doesn't record them.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md#A3-A4; idea.uk/idea_uk_open_discussion.md#3-4; idea.uk/RUNBOOK_idea_uk(9).md (Refunds — manual)
- **relations:** request-then-confirm flow; Stripe pattern.
- **verify-later:** service.go routes (expect no /status, /refund).

<!-- SOURCE: U04_idea_uk.md -->
### SFI26 Diff Alerts (first vertical tool) — replacing the single-farm assessment
- **category:** NEW:idea-product
- **status-signal:** aspirational
- **status-evidence:** "Tool swapped 2026-05-28… paused on liability grounds"; Phase C fully specified (C1–C5) with no build evidence; the base DEVELOPMENT_RUNBOOK still carries the original single-farm Phase C (family-delta capture of the abandoned product).
- **what:** The first Layer-3 vertical tool: a subscription digest for UK farm advisors summarising what changed in Defra/RPA SFI26 guidance, from a versioned scraped corpus, with every change cited to source+version, weekly, operator-reviewed for 8 issues before auto-send. Scored 19/25 with Risk 4. It replaced the **SFI26 single-farm assessment** (abandoned/backlogged: Risk 2 — a wrong number could cost a farmer £5–50k), the first product decision the Risk column changed. Chassis-native by design (recurring, per-user state, scheduled), the opposite plumbing to standalone idea.uk.
- **sources:** idea.uk/DEVELOPMENT_RUNBOOK(3).md (Phase C + swap note); idea.uk/DEVELOPMENT_RUNBOOK.md (base — original single-farm Phase C); idea.uk/CONSOLIDATION_where_it_all_fits.md (Layer 3)
- **relations:** operator-risk column; liability framework (SFI T&Cs draft); vet-med-pricing (sibling scraping shape).
- **verify-later:** any SFI corpus/agent in the repo or DB (expect none).

<!-- SOURCE: U04_idea_uk.md -->
### idea.uk standalone service page-serving and deploy gotchas
- **category:** NEW:idea-product
- **status-signal:** deployed
- **status-evidence:** Debugging guide §11 added for the idea.uk service; each gotcha tied to a fixed live incident.
- **what:** The operational failure catalogue for the single-binary service: every served path needs an explicit mux handler (bare 404s on linked pages); `writeHTML` fragments vs the `a.page()` full-page brand wrapper (navigation targets must wrap; injected fragments must not); startup templating of `CONTACT_EMAIL`/`MONTH_SLOTS` placeholders; systemd EnvironmentFile keeps inline comments (crash-loop + nginx 502); certbot failure made non-fatal in setup.sh; replace a running binary by scp-to-temp + `mv -f` (text-file-busy); Let's Encrypt rejects placeholder emails.
- **sources:** idea.uk/016_debugging_guide_v2_32(1).md#11; idea.uk/golang_files/README_setup_SETUP.md; idea.uk/BUGS_idea_uk(4).md (mobile safe-area padding)
- **relations:** setup.sh; VM launch plan.
- **verify-later:** service.go routes() vs page.html hrefs.

<!-- SOURCE: U22_recent_small_docs.md -->
### business_intel schema (multi-vertical business intelligence platform)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Agent definitions seeded with `status = 'experimental'`; verification_progress/discovery_stats views described as "used by bulk script monitoring"; ~3,500 vet practices already loaded.
- **what:** A separate `business_intel` schema inside `clients_db` (distinct from the public website-builder schema) that models businesses for data-collection verticals. Layered design: universal `businesses` table + `business_verticals` registry, with 1:1 vertical detail tables (`vet_practice_details`), pricing, products, people, and provenance. Seeded verticals: veterinary, online-pharmacy, seaweed-farming.
- **sources:** docs019_business/001_business_intel_schema.sql#layers, docs019_business/002_business_intel_actions.md#load_business_record
- **relations:** vet-practice-verifier, area-sweep discovery, vet-med-pricing (business_prices/product_prices)
- **verify-later:** schema `business_intel` in clients_db; tables businesses, business_verticals, vet_practice_details, business_prices, products, product_prices

<!-- SOURCE: U22_recent_small_docs.md -->
### Business verticals registry (business_intel)
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `INSERT INTO business_intel.business_verticals ... ON CONFLICT (slug) DO NOTHING` seeds veterinary/online-pharmacy/seaweed-farming with `default_agent_type`.
- **what:** A `business_verticals` table keying each collection vertical by slug, display name, `default_agent_type` (e.g. `vet-practice-verifier`, `pharmacy-price-monitor`), and per-vertical `collection_config` JSONB. Businesses and collection tasks reference the vertical; used to scope which agent handles a business type. Distinct from the docs021 knowledge `vertical_registry`.
- **sources:** docs019_business/001_business_intel_schema.sql#seed-verticals
- **relations:** vertical_registry (docs021 — different table, same "vertical" idea), collection_tasks
- **verify-later:** business_intel.business_verticals rows and default_agent_type usage

<!-- SOURCE: U22_recent_small_docs.md -->
### Data observations provenance model
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `store_business_verification` inserts a `data_observations` row per agent run with `raw_data JSONB`, source, confidence, orchestration_id.
- **what:** Every scrape/search/submission is recorded as a `data_observations` row carrying raw + extracted data, source type/name/url, extraction confidence, and the producing `orchestration_id`. Provides an audit trail and change history for business facts, separate from the current values on the business record. Temporal staleness columns (first_seen/last_confirmed/missed_count/is_stale) track freshness on prices and contacts.
- **sources:** docs019_business/001_business_intel_schema.sql#layer3, docs019_business/009_discovery_candidates.sql#temporal-tracking
- **relations:** business_intel schema, vet-practice-verifier
- **verify-later:** business_intel.data_observations; is_stale/missed_count columns; stale_contacts view

<!-- SOURCE: U22_recent_small_docs.md -->
### collection_tasks queue + batch claiming
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `load_business_batch` claims pending tasks via `FOR UPDATE SKIP LOCKED`; unique partial index prevents duplicate pending tasks per business+task_type.
- **what:** A `collection_tasks` queue (task_type initial_verification/price_refresh/status_check/discovery; status pending→in_progress→completed/failed/needs_review; priority). Agents claim batches atomically with SKIP LOCKED and reset orphaned in_progress rows after crashes. `ensure_collection_tasks` backfills tasks for pending businesses.
- **sources:** docs019_business/001_business_intel_schema.sql#collection_tasks, docs019_business/002_business_intel_actions.md#load_business_batch, docs019_business/015_collection_tasks.sql
- **relations:** vet-batch-processor, maintenance_queue (same claim pattern)
- **verify-later:** business_intel.collection_tasks; idx_collection_tasks_unique_pending; load_business_batch action

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-practice-verifier agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Seeded `status='experimental'`; SQL file is a long trail of iterative production fixes (Go-template dots, ai_service path, scraped_data path, prepare_context step) implying live debugging, plus "stuck at scrape_website - timeout goroutine lost" cleanup.
- **what:** Single-practice orchestrator workflow: load_business → search_practice (web_search) → scrape_website → prepare_context → extract_and_reconcile (LLM JSON extraction of business/vet_details/prices/staff/contacts) → store_results → scan_discoveries. Runs on claude-haiku-4-5. Callable standalone or spawned by the batch processor.
- **sources:** docs019_business/004_vet_practice_verifier.sql, docs019_business/002_business_intel_actions.md#store_business_verification
- **relations:** vet-batch-processor, prepare_extraction_context, scan_discovery_candidates, discovery_candidates
- **verify-later:** agent_definitions type='vet-practice-verifier'; actions load_business_record/store_business_verification

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-batch-processor agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Three documented fix rounds (spawn verifier first; continue_on_error; "remove loop — loop steps can't re-expand", max_iterations 50→250) show it broke and was reworked in production.
- **what:** Single-pod orchestrator that claims a batch of pending verification tasks and processes them sequentially, spawning one reusable `vet-practice-verifier` and calling it per business. Designed for polite, low-throughput collection; drains the queue by re-running.
- **sources:** docs019_business/003_vet_batch_processor.sql, docs019_business/005_initial_messaging.md, docs019_business/006_initial_messaging.sh
- **relations:** vet-practice-verifier, collection_tasks, vet-pipeline-orchestrator
- **verify-later:** agent_definitions type='vet-batch-processor'; loop step re-expansion behaviour in chassis

<!-- SOURCE: U22_recent_small_docs.md -->
### Geographic area-sweep discovery system
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** "We currently have around 3,500 practices and estimate ~5,000 in the UK"; converted from fire-and-forget dispatch to spawn+loop; costs "3,402 credits out of our 100k/month budget".
- **what:** Two-agent system (`area-sweep-orchestrator` + `area-sweep-discoverer`) that sweeps every UK postcode district (3,402 seeded in `search_areas`) via Firecrawl search, skips directory/aggregator domains, checks results against existing businesses and candidates, and inserts new finds into `discovery_candidates`. Go actions: load_unswept_areas, dispatch_area_discoverers, process_area_sweep.
- **sources:** docs019_business/011_area_sweep_discovery_system.md, docs019_business/010_district_search_areas_uk.sql, docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** discovery_candidates, vet-pipeline-orchestrator, search_result_cache
- **verify-later:** business_intel.search_areas seed count; actions load_unswept_areas/dispatch_area_discoverers/process_area_sweep

<!-- SOURCE: U22_recent_small_docs.md -->
### Discovery candidates + promotion pipeline
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** `promote_candidates` PL/pgSQL loops pending candidates → businesses with dedup/dismiss logic; status flow pending→matched/promoted/dismissed/needs_enrichment.
- **what:** `discovery_candidates` stores practices found in search results that don't match existing records, with match_method/confidence and group detection. A promotion routine inserts website-bearing candidates into `businesses` (status 'pending'), skips URL duplicates and directory-title junk, and queues them for verification. `search_result_cache` stores raw results for later mining.
- **sources:** docs019_business/009_discovery_candidates.sql, docs019_business/012_promote_candidates.sql
- **relations:** area-sweep discovery, collection_tasks, scan_discovery_candidates
- **verify-later:** business_intel.discovery_candidates; promote_candidates action; discovery_summary/discovery_stats views

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-pipeline-orchestrator (rolling pipeline)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Multiple reworks: fire-and-forget → spawn+loop coordinator → thin coordinator calling child orchestrators; timeouts bumped to 12h sweep / 6h verify.
- **what:** A rolling coordinator that advances work from previous runs each time it runs: load unswept areas → sweep → promote discovery_candidates → ensure_collection_tasks → run batch verification. Evolved from firing Kafka messages to spawning/awaiting child orchestrators (area-sweep + batch-processor) with promotion between.
- **sources:** docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** area-sweep discovery, vet-batch-processor, promote_candidates
- **verify-later:** agent_definitions type='vet-pipeline-orchestrator'; ensure_collection_tasks action

<!-- SOURCE: U22_recent_small_docs.md -->
### prepare_extraction_context / scan_discovery_candidates actions
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Wired into vet-practice-verifier via UPDATE migrations adding prepare_context and scan_discoveries steps; "reuses the skipDomains map from scan_discovery_candidates.go".
- **what:** Two supporting Go actions in the vet pipeline: `prepare_extraction_context` formats search results + scraped content (max_content_length/max_snippets) into a clean `extraction_context` for the LLM step; `scan_discovery_candidates` scans a verifier's search results for unknown practices (skipping aggregator domains) and inserts them into discovery_candidates. Both illustrate the "complexity in Go actions, thin workflow" convention.
- **sources:** docs019_business/004_vet_practice_verifier.sql#prepare_context, #scan_discoveries, docs019_business/011_area_sweep_discovery_system.md
- **relations:** vet-practice-verifier, discovery_candidates, area-sweep process_area_sweep (shares skipDomains)
- **verify-later:** actions prepare_extraction_context, scan_discovery_candidates; skipDomains map

<!-- SOURCE: U22_recent_small_docs.md -->
### business_intel schema (multi-vertical business intelligence platform)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Agent definitions seeded with `status = 'experimental'`; verification_progress/discovery_stats views described as "used by bulk script monitoring"; ~3,500 vet practices already loaded.
- **what:** A separate `business_intel` schema inside `clients_db` (distinct from the public website-builder schema) that models businesses for data-collection verticals. Layered design: universal `businesses` table + `business_verticals` registry, with 1:1 vertical detail tables (`vet_practice_details`), pricing, products, people, and provenance. Seeded verticals: veterinary, online-pharmacy, seaweed-farming.
- **sources:** docs019_business/001_business_intel_schema.sql#layers, docs019_business/002_business_intel_actions.md#load_business_record
- **relations:** vet-practice-verifier, area-sweep discovery, vet-med-pricing (business_prices/product_prices)
- **verify-later:** schema `business_intel` in clients_db; tables businesses, business_verticals, vet_practice_details, business_prices, products, product_prices

<!-- SOURCE: U22_recent_small_docs.md -->
### Business verticals registry (business_intel)
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `INSERT INTO business_intel.business_verticals ... ON CONFLICT (slug) DO NOTHING` seeds veterinary/online-pharmacy/seaweed-farming with `default_agent_type`.
- **what:** A `business_verticals` table keying each collection vertical by slug, display name, `default_agent_type` (e.g. `vet-practice-verifier`, `pharmacy-price-monitor`), and per-vertical `collection_config` JSONB. Businesses and collection tasks reference the vertical; used to scope which agent handles a business type. Distinct from the docs021 knowledge `vertical_registry`.
- **sources:** docs019_business/001_business_intel_schema.sql#seed-verticals
- **relations:** vertical_registry (docs021 — different table, same "vertical" idea), collection_tasks
- **verify-later:** business_intel.business_verticals rows and default_agent_type usage

<!-- SOURCE: U22_recent_small_docs.md -->
### Data observations provenance model
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `store_business_verification` inserts a `data_observations` row per agent run with `raw_data JSONB`, source, confidence, orchestration_id.
- **what:** Every scrape/search/submission is recorded as a `data_observations` row carrying raw + extracted data, source type/name/url, extraction confidence, and the producing `orchestration_id`. Provides an audit trail and change history for business facts, separate from the current values on the business record. Temporal staleness columns (first_seen/last_confirmed/missed_count/is_stale) track freshness on prices and contacts.
- **sources:** docs019_business/001_business_intel_schema.sql#layer3, docs019_business/009_discovery_candidates.sql#temporal-tracking
- **relations:** business_intel schema, vet-practice-verifier
- **verify-later:** business_intel.data_observations; is_stale/missed_count columns; stale_contacts view

<!-- SOURCE: U22_recent_small_docs.md -->
### collection_tasks queue + batch claiming
- **category:** NEW:business-intelligence-platform
- **status-signal:** deployed
- **status-evidence:** `load_business_batch` claims pending tasks via `FOR UPDATE SKIP LOCKED`; unique partial index prevents duplicate pending tasks per business+task_type.
- **what:** A `collection_tasks` queue (task_type initial_verification/price_refresh/status_check/discovery; status pending→in_progress→completed/failed/needs_review; priority). Agents claim batches atomically with SKIP LOCKED and reset orphaned in_progress rows after crashes. `ensure_collection_tasks` backfills tasks for pending businesses.
- **sources:** docs019_business/001_business_intel_schema.sql#collection_tasks, docs019_business/002_business_intel_actions.md#load_business_batch, docs019_business/015_collection_tasks.sql
- **relations:** vet-batch-processor, maintenance_queue (same claim pattern)
- **verify-later:** business_intel.collection_tasks; idx_collection_tasks_unique_pending; load_business_batch action

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-practice-verifier agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Seeded `status='experimental'`; SQL file is a long trail of iterative production fixes (Go-template dots, ai_service path, scraped_data path, prepare_context step) implying live debugging, plus "stuck at scrape_website - timeout goroutine lost" cleanup.
- **what:** Single-practice orchestrator workflow: load_business → search_practice (web_search) → scrape_website → prepare_context → extract_and_reconcile (LLM JSON extraction of business/vet_details/prices/staff/contacts) → store_results → scan_discoveries. Runs on claude-haiku-4-5. Callable standalone or spawned by the batch processor.
- **sources:** docs019_business/004_vet_practice_verifier.sql, docs019_business/002_business_intel_actions.md#store_business_verification
- **relations:** vet-batch-processor, prepare_extraction_context, scan_discovery_candidates, discovery_candidates
- **verify-later:** agent_definitions type='vet-practice-verifier'; actions load_business_record/store_business_verification

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-batch-processor agent
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Three documented fix rounds (spawn verifier first; continue_on_error; "remove loop — loop steps can't re-expand", max_iterations 50→250) show it broke and was reworked in production.
- **what:** Single-pod orchestrator that claims a batch of pending verification tasks and processes them sequentially, spawning one reusable `vet-practice-verifier` and calling it per business. Designed for polite, low-throughput collection; drains the queue by re-running.
- **sources:** docs019_business/003_vet_batch_processor.sql, docs019_business/005_initial_messaging.md, docs019_business/006_initial_messaging.sh
- **relations:** vet-practice-verifier, collection_tasks, vet-pipeline-orchestrator
- **verify-later:** agent_definitions type='vet-batch-processor'; loop step re-expansion behaviour in chassis

<!-- SOURCE: U22_recent_small_docs.md -->
### Geographic area-sweep discovery system
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** "We currently have around 3,500 practices and estimate ~5,000 in the UK"; converted from fire-and-forget dispatch to spawn+loop; costs "3,402 credits out of our 100k/month budget".
- **what:** Two-agent system (`area-sweep-orchestrator` + `area-sweep-discoverer`) that sweeps every UK postcode district (3,402 seeded in `search_areas`) via Firecrawl search, skips directory/aggregator domains, checks results against existing businesses and candidates, and inserts new finds into `discovery_candidates`. Go actions: load_unswept_areas, dispatch_area_discoverers, process_area_sweep.
- **sources:** docs019_business/011_area_sweep_discovery_system.md, docs019_business/010_district_search_areas_uk.sql, docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** discovery_candidates, vet-pipeline-orchestrator, search_result_cache
- **verify-later:** business_intel.search_areas seed count; actions load_unswept_areas/dispatch_area_discoverers/process_area_sweep

<!-- SOURCE: U22_recent_small_docs.md -->
### Discovery candidates + promotion pipeline
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** `promote_candidates` PL/pgSQL loops pending candidates → businesses with dedup/dismiss logic; status flow pending→matched/promoted/dismissed/needs_enrichment.
- **what:** `discovery_candidates` stores practices found in search results that don't match existing records, with match_method/confidence and group detection. A promotion routine inserts website-bearing candidates into `businesses` (status 'pending'), skips URL duplicates and directory-title junk, and queues them for verification. `search_result_cache` stores raw results for later mining.
- **sources:** docs019_business/009_discovery_candidates.sql, docs019_business/012_promote_candidates.sql
- **relations:** area-sweep discovery, collection_tasks, scan_discovery_candidates
- **verify-later:** business_intel.discovery_candidates; promote_candidates action; discovery_summary/discovery_stats views

<!-- SOURCE: U22_recent_small_docs.md -->
### vet-pipeline-orchestrator (rolling pipeline)
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Multiple reworks: fire-and-forget → spawn+loop coordinator → thin coordinator calling child orchestrators; timeouts bumped to 12h sweep / 6h verify.
- **what:** A rolling coordinator that advances work from previous runs each time it runs: load unswept areas → sweep → promote discovery_candidates → ensure_collection_tasks → run batch verification. Evolved from firing Kafka messages to spawning/awaiting child orchestrators (area-sweep + batch-processor) with promotion between.
- **sources:** docs019_business/014_vet_pipeline_orchestrator.sql
- **relations:** area-sweep discovery, vet-batch-processor, promote_candidates
- **verify-later:** agent_definitions type='vet-pipeline-orchestrator'; ensure_collection_tasks action

<!-- SOURCE: U22_recent_small_docs.md -->
### prepare_extraction_context / scan_discovery_candidates actions
- **category:** NEW:business-intelligence-platform
- **status-signal:** partial
- **status-evidence:** Wired into vet-practice-verifier via UPDATE migrations adding prepare_context and scan_discoveries steps; "reuses the skipDomains map from scan_discovery_candidates.go".
- **what:** Two supporting Go actions in the vet pipeline: `prepare_extraction_context` formats search results + scraped content (max_content_length/max_snippets) into a clean `extraction_context` for the LLM step; `scan_discovery_candidates` scans a verifier's search results for unknown practices (skipping aggregator domains) and inserts them into discovery_candidates. Both illustrate the "complexity in Go actions, thin workflow" convention.
- **sources:** docs019_business/004_vet_practice_verifier.sql#prepare_context, #scan_discoveries, docs019_business/011_area_sweep_discovery_system.md
- **relations:** vet-practice-verifier, discovery_candidates, area-sweep process_area_sweep (shares skipDomains)
- **verify-later:** actions prepare_extraction_context, scan_discovery_candidates; skipDomains map

<!-- SOURCE: U20_legacy_docs_a.md -->
### Playbook > Strategic Pattern > Component hierarchy (Librarian as system brain)
- **category:** NEW:conversion-playbooks
- **status-signal:** abandoned
- **status-evidence:** Extensive design (Playbooks/Strategic_Patterns/Pattern_Component_Slots/Components schema, success_score feedback) across website_analysis 001–003; no implementation era follows — the MVP path (chief-strategist + in-house components) shipped instead, and the schema never reappears.
- **what:** "Strategy-to-website engine": the library stores *business solutions*, not just components — Playbooks (objective+vertical strategies with success scores, e.g. affiliate product-review), containing Strategic Patterns (comparison-table, best-of listicle), containing Components. Learn loop classifies scraped winners into this hierarchy; Execute loop queries "best playbook for objective X in vertical Y" and assembles it; A/B results feed success_score back. The Librarian is the sole read/write gatekeeper; "the link is the database schema".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** behavioural models library (the surviving cousin); site-spec-and-classifier archetype system is the spiritual live successor; affiliate content-type placement knowledge (reviews/comparisons/listicles) embedded here.
- **verify-later:** confirm no Playbooks/Strategic_Patterns tables exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Behavioural models library and functional component labelling
- **category:** NEW:conversion-playbooks
- **status-signal:** partial
- **status-evidence:** "PAS" shipped as a real input (`"model": "PAS"` in mvp-site-builder trigger messages; chief-strategist prompt takes {{.model}}); the wider library (AIDA, Fogg B=MAP, Cialdini, Hook) and deep inference labelling remained design.
- **what:** Components are labelled by *behavioural function*, not visual pattern: not "hero" but "attention_capture"/"problem_statement"/"social_proof", drawn from marketing science (AIDA, PAS, Fogg Behaviour Model, Cialdini's persuasion principles, the Hook model). Build plans map a chosen behavioural model to a sequence of functional sections; the architect assembles "a psychological argument, not just a visual page". Self-critiques recorded: inference black-box risk (LLM can't reliably tell "agitation" from "interest"), theory-vs-reality gap, new-generic monoculture trap.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md; docs004_website_capture_project/website_analysis/README.007.behavioural_models.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** MVF; data-function contract; content strategy in current content pipeline (content-quality docs) is the descendant.
- **verify-later:** whether current build plans still carry a behavioural model field.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Minimal Viable Funnel (pragmatic-first Day-1 build)
- **category:** NEW:conversion-playbooks
- **status-signal:** superseded
- **status-evidence:** Fully built as mvp-site-builder (boxing-tickets.com runs); superseded within docs004 itself by the briefing→specialist-architect pipeline and later by the current work-item site build.
- **what:** Anti-boil-the-ocean strategy: start with one behavioural model (PAS) and three generic in-house components (problem/agitate/solution blocks) so a strategically coherent landing page can be built with zero scraped data — solving the cold-start problem. Scraping demoted to an "iteration engine" suggesting upgrades.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md#minimal-viable-funnel; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** MVP site builder pipeline; intelligent fallback; in-house forge.
- **verify-later:** —

<!-- SOURCE: U20_legacy_docs_a.md -->
### Strategic fallback stubs for non-replicable components
- **category:** NEW:conversion-playbooks
- **status-signal:** abandoned
- **status-evidence:** Design-only ("Store 'Stubs' with 'Fallbacks'… two-pronged output") in website_analysis 001/003; no stub tables or developer-task topics appear later.
- **what:** When ingestion finds a component it can't replicate (e.g. a mortgage calculator), record a Stub with its *strategic goal* (lead-gen-quote) and a linked simple fallback component (CTA form). The live site ships the working fallback; simultaneously a developer task goes to a HITL queue ("developer.tasks.required") to build the real thing as v2. The site is always complete and strategically sound.
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#strategic-fallback; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** dynamic-applications (the current interactive app generation finally addresses "non-replicable dynamic apps"); HITL queue.
- **verify-later:** none — idea registry.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Playbook > Strategic Pattern > Component hierarchy (Librarian as system brain)
- **category:** NEW:conversion-playbooks
- **status-signal:** abandoned
- **status-evidence:** Extensive design (Playbooks/Strategic_Patterns/Pattern_Component_Slots/Components schema, success_score feedback) across website_analysis 001–003; no implementation era follows — the MVP path (chief-strategist + in-house components) shipped instead, and the schema never reappears.
- **what:** "Strategy-to-website engine": the library stores *business solutions*, not just components — Playbooks (objective+vertical strategies with success scores, e.g. affiliate product-review), containing Strategic Patterns (comparison-table, best-of listicle), containing Components. Learn loop classifies scraped winners into this hierarchy; Execute loop queries "best playbook for objective X in vertical Y" and assembles it; A/B results feed success_score back. The Librarian is the sole read/write gatekeeper; "the link is the database schema".
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md; docs004_website_capture_project/website_analysis/README.002.summary_of_plan_agents_groups.md
- **relations:** behavioural models library (the surviving cousin); site-spec-and-classifier archetype system is the spiritual live successor; affiliate content-type placement knowledge (reviews/comparisons/listicles) embedded here.
- **verify-later:** confirm no Playbooks/Strategic_Patterns tables exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Behavioural models library and functional component labelling
- **category:** NEW:conversion-playbooks
- **status-signal:** partial
- **status-evidence:** "PAS" shipped as a real input (`"model": "PAS"` in mvp-site-builder trigger messages; chief-strategist prompt takes {{.model}}); the wider library (AIDA, Fogg B=MAP, Cialdini, Hook) and deep inference labelling remained design.
- **what:** Components are labelled by *behavioural function*, not visual pattern: not "hero" but "attention_capture"/"problem_statement"/"social_proof", drawn from marketing science (AIDA, PAS, Fogg Behaviour Model, Cialdini's persuasion principles, the Hook model). Build plans map a chosen behavioural model to a sequence of functional sections; the architect assembles "a psychological argument, not just a visual page". Self-critiques recorded: inference black-box risk (LLM can't reliably tell "agitation" from "interest"), theory-vs-reality gap, new-generic monoculture trap.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md; docs004_website_capture_project/website_analysis/README.007.behavioural_models.md; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md
- **relations:** MVF; data-function contract; content strategy in current content pipeline (content-quality docs) is the descendant.
- **verify-later:** whether current build plans still carry a behavioural model field.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Minimal Viable Funnel (pragmatic-first Day-1 build)
- **category:** NEW:conversion-playbooks
- **status-signal:** superseded
- **status-evidence:** Fully built as mvp-site-builder (boxing-tickets.com runs); superseded within docs004 itself by the briefing→specialist-architect pipeline and later by the current work-item site build.
- **what:** Anti-boil-the-ocean strategy: start with one behavioural model (PAS) and three generic in-house components (problem/agitate/solution blocks) so a strategically coherent landing page can be built with zero scraped data — solving the cold-start problem. Scraping demoted to an "iteration engine" suggesting upgrades.
- **sources:** docs004_website_capture_project/website_analysis/README.006.visual_to_code.md#minimal-viable-funnel; docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** MVP site builder pipeline; intelligent fallback; in-house forge.
- **verify-later:** —

<!-- SOURCE: U20_legacy_docs_a.md -->
### Strategic fallback stubs for non-replicable components
- **category:** NEW:conversion-playbooks
- **status-signal:** abandoned
- **status-evidence:** Design-only ("Store 'Stubs' with 'Fallbacks'… two-pronged output") in website_analysis 001/003; no stub tables or developer-task topics appear later.
- **what:** When ingestion finds a component it can't replicate (e.g. a mortgage calculator), record a Stub with its *strategic goal* (lead-gen-quote) and a linked simple fallback component (CTA form). The live site ships the working fallback; simultaneously a developer task goes to a HITL queue ("developer.tasks.required") to build the real thing as v2. The site is always complete and strategically sound.
- **sources:** docs004_website_capture_project/website_analysis/README.001.prompt_of_objectives.md#strategic-fallback; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** dynamic-applications (the current interactive app generation finally addresses "non-replicable dynamic apps"); HITL queue.
- **verify-later:** none — idea registry.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Pragmatic Evolution model (explore/exploit portfolio cohorts)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** Full strategy synthesis in 008/014 (cohorts: top-10% untouched, middle-40% careful P1-P2 A/B tests, bottom-50% high-velocity churn; site-specific optimisation "no monoculture"); no subsequent doc era operates a portfolio this way — the platform pivoted to per-site quality loops.
- **what:** An evolutionary algorithm over a large site portfolio: select worst performers, radically mutate them with mixed component "genes", evaluate fitness after 3 months. Critique recorded and resolved into an explore/exploit design: attribution black hole and SEO destabilisation confine chaos to a "loser" cohort where attribution is deliberately ignored; winners graduate. Winning changes are applied only to individual sites where they actually won, and content evolves on a separate continuous track from layout to protect SEO.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** improvement-loop (the live per-site discovery→audit→fix cycle is the surviving descendant in spirit); traffic-analytics (fitness signal dependency).
- **verify-later:** none — strategy registry; check if any cohort/experiment tables exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Hypothesis priority list (learn loop as idea generator, not fact finder)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** 008/014: "All scraped data is treated as messy, high-correlation ideas, not truth… Librarian generates a Hypothesis Priority List (P1–P5)"; scorecard interrogation of sites against all behavioural models; no implementation follows.
- **what:** Epistemics for the scraping programme: accept that ingestion finds correlation ("cargo cults"), rank target sites by external success metrics (Ahrefs/Semrush APIs via an seo_api_adapter), interrogate each against every behavioural model to produce confidence scorecards, and emit a prioritised backlog of testable hypotheses for the Evolve loop to convert into causation.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** Prospector/seo_api_adapter (never built); llm-quality-testing shares the evaluation mindset.
- **verify-later:** none.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Pragmatic Evolution Engine (portfolio build/learn/test/optimize)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** docs009/001 full 4-phase plan ("Internal Library of Effectiveness", "Controlled Evolutionary Cohorts"); no subsequent doc implements cohort testing, manifests, or the Librarian.
- **what:** The founding mission statement for a large-scale website portfolio: Phase 1 pragmatic-first MVP builds from behavioural models (AIDA/PAS) with intelligent component fallback; Phase 2 "Idea Generator" evidence gathering from winner sites (Prospector via Ahrefs-type metrics, Capture Bot producing dom+screenshot+layout_map "Rosetta Stone", Pattern Deconstructor VLM scoring components against behavioural models, Librarian producing a Hypothesis Priority List); Phase 3 large-scale single-variable A/B cohort tests turning correlation into causation, with content and layout evolved on separate tracks for SEO stability; Phase 4 site-specific optimization (winners applied only where they won — no monoculture), manifest.json component "genes" per site, git_hook_adapter flagging human-edited repos as desynchronized for HITL review, and exporter agents (WordPress XML/SQL) for client handoff.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Core-Mission; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-2; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-4
- **relations:** site interrogation/pattern library; adoption-pipeline; improvement-loop is the maintenance-shaped descendant; llm-quality-testing.
- **verify-later:** any manifest.json in site repos; git_hook_adapter; cohort/experiment tables (expected absent).

<!-- SOURCE: U20_legacy_docs_a.md -->
### Pragmatic Evolution model (explore/exploit portfolio cohorts)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** Full strategy synthesis in 008/014 (cohorts: top-10% untouched, middle-40% careful P1-P2 A/B tests, bottom-50% high-velocity churn; site-specific optimisation "no monoculture"); no subsequent doc era operates a portfolio this way — the platform pivoted to per-site quality loops.
- **what:** An evolutionary algorithm over a large site portfolio: select worst performers, radically mutate them with mixed component "genes", evaluate fitness after 3 months. Critique recorded and resolved into an explore/exploit design: attribution black hole and SEO destabilisation confine chaos to a "loser" cohort where attribution is deliberately ignored; winners graduate. Winning changes are applied only to individual sites where they actually won, and content evolves on a separate continuous track from layout to protect SEO.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** improvement-loop (the live per-site discovery→audit→fix cycle is the surviving descendant in spirit); traffic-analytics (fitness signal dependency).
- **verify-later:** none — strategy registry; check if any cohort/experiment tables exist.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Hypothesis priority list (learn loop as idea generator, not fact finder)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** 008/014: "All scraped data is treated as messy, high-correlation ideas, not truth… Librarian generates a Hypothesis Priority List (P1–P5)"; scorecard interrogation of sites against all behavioural models; no implementation follows.
- **what:** Epistemics for the scraping programme: accept that ingestion finds correlation ("cargo cults"), rank target sites by external success metrics (Ahrefs/Semrush APIs via an seo_api_adapter), interrogate each against every behavioural model to produce confidence scorecards, and emit a prioritised backlog of testable hypotheses for the Evolve loop to convert into causation.
- **sources:** docs004_website_capture_project/website_analysis/README.008.evolutionary_algorithm_of_site_portfolio.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md; docs004_website_capture_project/website_analysis/README.003.summary_for_development.md
- **relations:** Prospector/seo_api_adapter (never built); llm-quality-testing shares the evaluation mindset.
- **verify-later:** none.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Pragmatic Evolution Engine (portfolio build/learn/test/optimize)
- **category:** NEW:portfolio-evolution
- **status-signal:** abandoned
- **status-evidence:** docs009/001 full 4-phase plan ("Internal Library of Effectiveness", "Controlled Evolutionary Cohorts"); no subsequent doc implements cohort testing, manifests, or the Librarian.
- **what:** The founding mission statement for a large-scale website portfolio: Phase 1 pragmatic-first MVP builds from behavioural models (AIDA/PAS) with intelligent component fallback; Phase 2 "Idea Generator" evidence gathering from winner sites (Prospector via Ahrefs-type metrics, Capture Bot producing dom+screenshot+layout_map "Rosetta Stone", Pattern Deconstructor VLM scoring components against behavioural models, Librarian producing a Hypothesis Priority List); Phase 3 large-scale single-variable A/B cohort tests turning correlation into causation, with content and layout evolved on separate tracks for SEO stability; Phase 4 site-specific optimization (winners applied only where they won — no monoculture), manifest.json component "genes" per site, git_hook_adapter flagging human-edited repos as desynchronized for HITL review, and exporter agents (WordPress XML/SQL) for client handoff.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Core-Mission; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-2; docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#Phase-4
- **relations:** site interrogation/pattern library; adoption-pipeline; improvement-loop is the maintenance-shaped descendant; llm-quality-testing.
- **verify-later:** any manifest.json in site repos; git_hook_adapter; cohort/experiment tables (expected absent).

<!-- SOURCE: U22_recent_small_docs.md -->
### Vertical knowledge architecture
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "The architecture ... is now designed but not yet fully implemented"; implementation todo is phased 0-8 with only Phase 0 partially begun.
- **what:** The strategic pivot from a flat build pipeline to routing each domain to a specialised vertical (veterinary, energy_wholesale, finance_mortgage, seasonal_gifts, generic) that maintains its own knowledge-base collection, research strategy, content patterns, and monetisation config. Verticals are logical first (shared infra, knowledge_base collections + rag_lookup/index), physical later (dispatch_agent). Compounding moat: the tenth domain benefits from the first nine's research.
- **sources:** docs021.../020_vertical_cluster_architecture.md, docs021.../025_session_handoff_vertical_architecture.md, docs020.../011_where_we_are.md
- **relations:** vertical_registry, research/build cluster separation, RAG knowledge_base, site classifier vertical output
- **verify-later:** vertical_registry table; agent_definitions tagged 'vertical-orch'; knowledge_base collection usage

<!-- SOURCE: U22_recent_small_docs.md -->
### vertical_registry table + knowledge-base provenance extensions
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "These additions to the database, not yet applied"; seed INSERT for 5 verticals given as a to-do.
- **what:** A `vertical_registry` table mapping vertical_slug → research/build orchestrator types, knowledge_collection, research_sources, content_patterns/page_type_library, monetisation_config, refresh_schedule, and maturity_stage; plus knowledge_base column extensions (source_authority 1-5, source_url, source_date, vertical_slug, knowledge_type) for provenance-weighted retrieval. Seeds veterinary/energy/mortgage/seasonal/generic.
- **sources:** docs021.../020_vertical_cluster_architecture.md#8, docs021.../026_implementation_todo_vertical_architecture(2).md#2.1
- **relations:** vertical knowledge architecture, rag_lookup min_authority, business_verticals (different table)
- **verify-later:** vertical_registry table; knowledge_base.source_authority/vertical_slug/knowledge_type columns

<!-- SOURCE: U22_recent_small_docs.md -->
### Research/build cluster separation
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "Phase 2: Physical separation ... implementation deferred"; described as designed, Phase 8 in the todo.
- **what:** A two-cluster model separating messy/slow/shared research (web scraping, PDF parsing, LLM knowledge extraction, validation, indexing) from structured/fast/per-site build. They communicate only via the shared knowledge_base (Postgres reads/writes) and Kafka orchestration; build dispatches a research request when it hits a knowledge gap. Justified by independent scaling, failure isolation, clean logs, tighter build-cluster network policy, and clean cost attribution.
- **sources:** docs021.../020_vertical_cluster_architecture.md#5, docs021.../025_session_handoff_vertical_architecture.md#layer-5
- **relations:** vertical knowledge architecture, DispatchAgentAction, vertical research handler
- **verify-later:** any research-cluster agent definitions; dispatch_agent used for research requests

<!-- SOURCE: U22_recent_small_docs.md -->
### Vertical research handler + knowledge accumulation loop
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "Phase 4 (Research Handler)" is an unchecked todo; `needs_vertical_research` work-item type not yet in the schema.
- **what:** A `vertical-research-handler` agent that processes `needs_vertical_research` work items (identify sources → scrape → parse → LLM extract structured knowledge chunks → validate quality/confidence → rag_index with source_authority/vertical_slug). Realises the knowledge-accumulation loop: first domain bears foundational research cost, gaps become research items at priority 1-4 that content items (priority 10-17) depend on, and indexed knowledge benefits all future domains in the vertical.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#phase-4, docs021.../020_vertical_cluster_architecture.md#6
- **relations:** research/build separation, rag_index, work-item lifecycle, knowledge-indexer agent
- **verify-later:** needs_vertical_research item type; vertical-research-handler agent; check_knowledge_coverage action

<!-- SOURCE: U22_recent_small_docs.md -->
### Vertical knowledge architecture
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "The architecture ... is now designed but not yet fully implemented"; implementation todo is phased 0-8 with only Phase 0 partially begun.
- **what:** The strategic pivot from a flat build pipeline to routing each domain to a specialised vertical (veterinary, energy_wholesale, finance_mortgage, seasonal_gifts, generic) that maintains its own knowledge-base collection, research strategy, content patterns, and monetisation config. Verticals are logical first (shared infra, knowledge_base collections + rag_lookup/index), physical later (dispatch_agent). Compounding moat: the tenth domain benefits from the first nine's research.
- **sources:** docs021.../020_vertical_cluster_architecture.md, docs021.../025_session_handoff_vertical_architecture.md, docs020.../011_where_we_are.md
- **relations:** vertical_registry, research/build cluster separation, RAG knowledge_base, site classifier vertical output
- **verify-later:** vertical_registry table; agent_definitions tagged 'vertical-orch'; knowledge_base collection usage

<!-- SOURCE: U22_recent_small_docs.md -->
### vertical_registry table + knowledge-base provenance extensions
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "These additions to the database, not yet applied"; seed INSERT for 5 verticals given as a to-do.
- **what:** A `vertical_registry` table mapping vertical_slug → research/build orchestrator types, knowledge_collection, research_sources, content_patterns/page_type_library, monetisation_config, refresh_schedule, and maturity_stage; plus knowledge_base column extensions (source_authority 1-5, source_url, source_date, vertical_slug, knowledge_type) for provenance-weighted retrieval. Seeds veterinary/energy/mortgage/seasonal/generic.
- **sources:** docs021.../020_vertical_cluster_architecture.md#8, docs021.../026_implementation_todo_vertical_architecture(2).md#2.1
- **relations:** vertical knowledge architecture, rag_lookup min_authority, business_verticals (different table)
- **verify-later:** vertical_registry table; knowledge_base.source_authority/vertical_slug/knowledge_type columns

<!-- SOURCE: U22_recent_small_docs.md -->
### Research/build cluster separation
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "Phase 2: Physical separation ... implementation deferred"; described as designed, Phase 8 in the todo.
- **what:** A two-cluster model separating messy/slow/shared research (web scraping, PDF parsing, LLM knowledge extraction, validation, indexing) from structured/fast/per-site build. They communicate only via the shared knowledge_base (Postgres reads/writes) and Kafka orchestration; build dispatches a research request when it hits a knowledge gap. Justified by independent scaling, failure isolation, clean logs, tighter build-cluster network policy, and clean cost attribution.
- **sources:** docs021.../020_vertical_cluster_architecture.md#5, docs021.../025_session_handoff_vertical_architecture.md#layer-5
- **relations:** vertical knowledge architecture, DispatchAgentAction, vertical research handler
- **verify-later:** any research-cluster agent definitions; dispatch_agent used for research requests

<!-- SOURCE: U22_recent_small_docs.md -->
### Vertical research handler + knowledge accumulation loop
- **category:** NEW:vertical-knowledge-architecture
- **status-signal:** aspirational
- **status-evidence:** "Phase 4 (Research Handler)" is an unchecked todo; `needs_vertical_research` work-item type not yet in the schema.
- **what:** A `vertical-research-handler` agent that processes `needs_vertical_research` work items (identify sources → scrape → parse → LLM extract structured knowledge chunks → validate quality/confidence → rag_index with source_authority/vertical_slug). Realises the knowledge-accumulation loop: first domain bears foundational research cost, gaps become research items at priority 1-4 that content items (priority 10-17) depend on, and indexed knowledge benefits all future domains in the vertical.
- **sources:** docs021.../026_implementation_todo_vertical_architecture(2).md#phase-4, docs021.../020_vertical_cluster_architecture.md#6
- **relations:** research/build separation, rag_index, work-item lifecycle, knowledge-indexer agent
- **verify-later:** needs_vertical_research item type; vertical-research-handler agent; check_knowledge_coverage action

<!-- SOURCE: U04_idea_uk.md -->
### Stripe integration pattern: webhook as the only source of truth
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** Live £29 payments proven end-to-end 2026-06-14 (incl. resolving the stray-character webhook-secret incident); full setup documented from the real dashboards.
- **what:** The reference payments pattern proven by idea.uk: entitlement/fulfilment granted only on a signature-verified `checkout.session.completed` (browser redirects prove nothing); webhook handling idempotent via an event-dedup table; a **restricted** API key scoped to Checkout Sessions:Write only; test and live are separate accounts with separate webhook destinations and secrets ("a sandbox webhook does not cover live"); the signing secret must be byte-exact (one pasted stray character 400'd every event and stalled a paid order — recovered by resending the event); Stripe keeps its fee on refunds; no SDK — raw HTTP + HMAC verify.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (Stripe billing — setup + troubleshooting); idea.uk/PLAN_stripe_billing_integration(3).md (idea.uk reference block); idea.uk/golang_files/billing.go (header)
- **relations:** platform billing plan (adopts these principles); request-then-confirm flow.
- **verify-later:** billing.go webhook verify (HMAC-SHA256 over timestamp+body, constant-time compare).

<!-- SOURCE: U04_idea_uk.md -->
### Platform Stripe billing integration plan (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** "the auth service has a subscription scaffold… but no working payment integration — no Stripe SDK, no checkout creation, no webhooks"; every DDL marked PROPOSED.
- **what:** The chassis-wide billing design for the build/host/chat product: truth lives in the auth DB mutated only by verified webhooks; the chassis gates on a one-way-fed `client_entitlements` cache (Kafka entitlement-changed events + reconciliation sweep) because the maintenance heartbeat can't call auth per site; two charge shapes — recurring tier subscription per client and a one-off **$5 build credit** (Checkout mode=payment, consumed via the atomic-claim idiom); build-submission gate reuses the `approval_mode` hold; provider interface from day one. idea.uk is the cited working reference for the one-off path.
- **sources:** idea.uk/PLAN_stripe_billing_integration(3).md; idea.uk/RUNBOOK_idea_uk(9).md (reference implementation)
- **relations:** Stripe webhook pattern; admin-dashboard-and-api (auth service); scheduler heartbeats.
- **verify-later:** auth service repo subscriptions tables; any client_entitlements table (expect absent).

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Entitlement gate architecture (build-submission + maintenance-run gates)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN doc §8 describes both gates in the future tense as design; build order (§10) lists them as unbuilt steps
- **what:** Two entitlement checkpoints reusing existing chassis mechanisms: (1) build-submission gate — a new `pending_entitlement` hold state on `site_work_items.approval_mode` (mirroring the existing hitl/pending_review pattern), parking the first expensive work item until a billing check clears, with atomic credit consumption via the same UPDATE...RETURNING idiom as `claim_work_item`; (2) maintenance-run gate — a join-filter added to the three heartbeat selection queries requiring `maintenance_active`, valuable even before any domain is sold.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§8, stripe/001commentary.md#final turns
- **relations:** Two-plane billing architecture; Ownership hierarchy reuse for entitlement scoping; One-off credit vs recurring subscription billing model
- **verify-later:** site_work_items.approval_mode values; build-pipeline-trigger/improvement-loop/content-feed-trigger selection SQL

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two-plane billing architecture (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §3: "Truth = auth DB, mutated only by webhooks... Gate reads = chassis client_entitlements cache, fed one-directionally from auth" — all proposed, table marked PROPOSED
- **what:** Splits billing across two databases/services with one directional bridge: the auth service owns billing truth (subscriptions, credits, webhook-driven events); the chassis reads a local `client_entitlements` cache table fed by an entitlement-changed Kafka event plus a reconciliation sweep backstop — required because the maintenance heartbeat must join across thousands of sites per tick.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§3,§5
- **relations:** Entitlement gate architecture; Isolated chat/satellite architecture (Y-copy); Pluggable billing provider abstraction
- **verify-later:** proposed table client_entitlements; entitlement-changed Kafka event/consumer (not built)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pluggable billing provider abstraction (Stripe as implementation #1)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §4 gives a full Go interface sketch explicitly labelled "Sketch, not final"; current code has "no Stripe SDK" imported at all
- **what:** A `Provider` interface (`EnsureCustomer`, `CreateSubscriptionCheckout`, `CreateOneOffCheckout`, `CreatePortalSession`, `CancelSubscription`, `ParseWebhook`) behind which Stripe is the first implementation, normalising provider-specific webhook payloads into a provider-agnostic `Event` type. Justified as "zero retrofit cost" specifically because no Stripe integration exists yet.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§4,#TL;DR
- **relations:** Two-plane billing architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Webhook-as-only-source-of-truth billing principle
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "service.go imports only uuid, zap, time, context, fmt — no Stripe SDK. CreateSubscription is a bare DB insert that sets Status = 'active' with no payment step... There's no webhook handler anywhere." Confirmed by direct code read
- **what:** The organising design principle for the billing plan: client-side success redirects must never grant entitlement — only a signature-verified Stripe webhook, deduplicated by `provider_event_id`, may mutate entitlement state. Directly motivated by the audited finding that today `status = active` merely means "a row exists" with zero payment verification.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§2,#Appendix, stripe/001commentary.md#Stripe audit turn
- **relations:** Existing but non-functional auth-service subscription scaffold; Two-plane billing architecture
- **verify-later:** internal/auth-service/subscription/service.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### One-off credit vs recurring subscription billing model
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §7 and §5 `billing_credits` DDL are both marked PROPOSED; no credit ledger exists in code
- **what:** Two distinct charge shapes: recurring (maintenance/tier subscription, reusing the existing but non-functional subscription scaffold) and one-off (the $5-per-site build and first-site-free grant, modelled as a `billing_credits` ledger — granted/consumed counts per client). Build proceeds only once a credit is atomically consumed via the entitlement gate.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§5,§7, stripe/001commentary.md#pricing discussion turn
- **relations:** Entitlement gate architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** proposed billing_credits, billing_events tables (auth DB)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Existing but non-functional auth-service subscription scaffold
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "GetUsageStats returns hardcoded zeros with the comment 'returning mock data,' which makes CheckQuota always pass... repository.go mixes ? (MySQL-style) placeholders... with $1 (Postgres-style)... a strong sign this module has never actually been exercised." (stripe/PLAN_stripe_billing_integration.md#§1,§Appendix); independently confirmed from the tools-workstream side (PLAN_isolated_chat_environment(5).md §13: "Billing is scaffolded, not wired," correcting that same document's own earlier "billing largely exists" assumption)
- **what:** A pre-existing `subscription` package in the auth service (models, repository, service, handlers) with a `subscriptions` table, tier constants (free/basic/premium/enterprise), `stripe_customer_id`/`stripe_subscription_id` columns, a `CheckoutSession` type, and JWT claims already carrying `client_id`+`tier` — all reusable — but verified as not wired: no Stripe SDK import anywhere, `CreateSubscription` is a bare insert with no payment step, no webhook handler exists, `CheckUsage`/`GetUsageStats` returns hardcoded zeros so quota checks always pass, and a placeholder-dialect inconsistency in `repository.go` is a strong sign the module has never run against a live database. Security consequence: any entitlement gate trusting `subscription.status` today only reflects "a row exists," not "payment cleared."
- **sources:** stripe/PLAN_stripe_billing_integration.md#§1,§Appendix, stripe/001commentary.md#Stripe audit turn, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13
- **relations:** Webhook-as-only-source-of-truth billing principle; Pluggable billing provider abstraction; Ownership hierarchy reuse for entitlement scoping; Entitlement gate architecture
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go; presence/absence of a Stripe webhook handler

<!-- SOURCE: U22_recent_small_docs.md -->
### Commercial model + entitlement seams (billing adapter)
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "billing/identity is mostly reuse, not new" — the auth service already has a `subscriptions` table with `stripe_customer_id`, tier definitions, JWT carrying client_id+tier; "live checkout-session creation and webhooks were not evident ... verify before relying."
- **what:** The saleability design: operator-primary (operate thousands of domains), vendor-optional (sell a domain + its backend, rarely the whole framework). Isolation unit = the satellite; separability unit = the domain (partition by site_id/domain, extractable + swappable credentials). Seams to honour now: ownership via existing clients→networks→sites hierarchy (re-parent network_id to sell), a pluggable billing adapter (Stripe first, generalise stripe_* columns to provider_*), two entitlement gates (build-submission reusing site_work_items.approval_mode → a pending_entitlement hold; maintenance-run filtering the heartbeat site-selection queries as a cost valve), a saas_cheap-vs-portfolio build-tier riding the existing batch/sync rail, and snapshot-able building blocks for whole-instance sales.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#13, docs025.../PLAN_simple_paid_multidomain_chat(1).md#2
- **relations:** auth-service subscriptions, site_work_items.approval_mode, batch processing (scheduled→batch), building-as-a-service
- **verify-later:** auth subscriptions table + Stripe webhook wiring; site_work_items.approval_mode; heartbeat site-selection queries

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### REVIEW_BEFORE_PAY billing flow supersedes charge-first flow
- **category:** payments
- **status-signal:** partial
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)": "Supersedes the older Flow/Email/AUTO_DELIVER notes above where they differ... `REVIEW_BEFORE_PAY` (default on)."
- **what:** idea.uk's original flow charged the customer first (Stripe Checkout), then ran the engine, then optionally held for operator review before emailing (`AUTO_DELIVER`). This was replaced by a `REVIEW_BEFORE_PAY` switch (default on): the operator's `/confirm` now *runs the engine first* and holds the draft for review; only after the operator approves does the buyer get a pay link — no money is taken until a human has seen the actual output. The original charge-first flow is kept as a fallback (`REVIEW_BEFORE_PAY=false`) "if engine cost ever spikes." A click-through token-based approve/decline UI (HMAC per order) was added on top to remove the need for curl+API-key.
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)"
- **relations:** idea.uk product; Stripe webhook-as-truth pattern
- **verify-later:** `idea-go/service.go` `REVIEW_BEFORE_PAY` branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Stripe webhook-as-truth billing pattern (idea.uk lightweight variant)
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Stripe billing — setup" section: live keys, live webhook destination IDs, "Billing follows the PLAN_stripe_billing_integration.md principles but in the lightweight pay-per-idea shape... proven end-to-end with a real card on 2026-06-14."
- **what:** idea.uk's billing never trusts a browser redirect; only a signature-verified `checkout.session.completed` webhook (deduped by event id) marks an order paid and triggers delivery. Uses a Stripe **restricted API key** scoped to `Checkout Sessions → Write` only (least privilege — no refunds, no customer/product read access needed since Checkout uses inline `price_data`). Refunds are manual-only in the Stripe dashboard (no `/refund` endpoint exists). This is presented explicitly as the lightweight, one-off-payment implementation of the same principles as the full chassis-wide Stripe plan (see separate entry) — webhook-is-truth, idempotent, provider behind an interface (FakeProvider swap for local testing).
- **sources:** `RUNBOOK_idea_uk(10).md` §"Stripe billing — setup" (webhook destination IDs, account IDs, restricted-key scoping, troubleshooting runbook for a real signature-mismatch incident on 2026-06-14)
- **relations:** chassis-wide Stripe billing integration plan (supersedes/generalizes); REVIEW_BEFORE_PAY flow
- **verify-later:** Stripe dashboard accounts `acct_1RNfPY08YuzM2cqf` (test) / `acct_1RNfPL02nQ76FNif` (live)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-wide Stripe billing integration plan (client_entitlements cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** Doc self-describes as "PROPOSED" throughout ("Schema caveat: this plan is written from the auth subscription Go models, not the auth DB migrations... Every DDL below is PROPOSED"). No claim of implementation for the chassis-wide version (idea.uk implemented its own lighter variant instead — see above).
- **what:** A designed-not-built architecture for platform-wide billing: auth service owns billing truth (subscriptions, one-off credits, webhook-verified events only), chassis reads through a one-directionally-fed cache table `client_entitlements` (never calls auth synchronously from the hot path), with two gates — a low-volume build-submission gate (`approval_mode='pending_entitlement'`) and a high-volume maintenance-run gate (join-filter on heartbeat queries). Covers both recurring (maintenance/tier) and one-off ($5 build credit) charge shapes, a provider interface abstracting Stripe, and a verified-findings appendix showing the existing auth subscription code is a non-functional scaffold (`CreateSubscription` stamps `status=active` with no payment, no Stripe SDK, no webhook handler, mock usage stats, and a `?`/`$1` placeholder-dialect mismatch implying the code was never run against one DB engine).
- **sources:** `docubundle_idea_golive/package_module/output_contexts/PLAN_stripe_billing_integration.md` (packaged context-pack snapshot, 390 lines); archive `PLAN_stripe_billing_integration(2).md` (identical to live `(3).md`)
- **relations:** Stripe webhook-as-truth pattern (idea.uk's lighter realisation of the same principles); isolated-chat-environment commercial model (referenced, live doc)
- **verify-later:** `internal/auth-service/subscription/{models,repository,service,handlers}.go` — confirm whether the scaffold described (no Stripe SDK, mock usage stats, dialect mismatch) is still the current state

<!-- SOURCE: U04_idea_uk.md -->
### Stripe integration pattern: webhook as the only source of truth
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** Live £29 payments proven end-to-end 2026-06-14 (incl. resolving the stray-character webhook-secret incident); full setup documented from the real dashboards.
- **what:** The reference payments pattern proven by idea.uk: entitlement/fulfilment granted only on a signature-verified `checkout.session.completed` (browser redirects prove nothing); webhook handling idempotent via an event-dedup table; a **restricted** API key scoped to Checkout Sessions:Write only; test and live are separate accounts with separate webhook destinations and secrets ("a sandbox webhook does not cover live"); the signing secret must be byte-exact (one pasted stray character 400'd every event and stalled a paid order — recovered by resending the event); Stripe keeps its fee on refunds; no SDK — raw HTTP + HMAC verify.
- **sources:** idea.uk/RUNBOOK_idea_uk(9).md (Stripe billing — setup + troubleshooting); idea.uk/PLAN_stripe_billing_integration(3).md (idea.uk reference block); idea.uk/golang_files/billing.go (header)
- **relations:** platform billing plan (adopts these principles); request-then-confirm flow.
- **verify-later:** billing.go webhook verify (HMAC-SHA256 over timestamp+body, constant-time compare).

<!-- SOURCE: U04_idea_uk.md -->
### Platform Stripe billing integration plan (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** "the auth service has a subscription scaffold… but no working payment integration — no Stripe SDK, no checkout creation, no webhooks"; every DDL marked PROPOSED.
- **what:** The chassis-wide billing design for the build/host/chat product: truth lives in the auth DB mutated only by verified webhooks; the chassis gates on a one-way-fed `client_entitlements` cache (Kafka entitlement-changed events + reconciliation sweep) because the maintenance heartbeat can't call auth per site; two charge shapes — recurring tier subscription per client and a one-off **$5 build credit** (Checkout mode=payment, consumed via the atomic-claim idiom); build-submission gate reuses the `approval_mode` hold; provider interface from day one. idea.uk is the cited working reference for the one-off path.
- **sources:** idea.uk/PLAN_stripe_billing_integration(3).md; idea.uk/RUNBOOK_idea_uk(9).md (reference implementation)
- **relations:** Stripe webhook pattern; admin-dashboard-and-api (auth service); scheduler heartbeats.
- **verify-later:** auth service repo subscriptions tables; any client_entitlements table (expect absent).

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Entitlement gate architecture (build-submission + maintenance-run gates)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN doc §8 describes both gates in the future tense as design; build order (§10) lists them as unbuilt steps
- **what:** Two entitlement checkpoints reusing existing chassis mechanisms: (1) build-submission gate — a new `pending_entitlement` hold state on `site_work_items.approval_mode` (mirroring the existing hitl/pending_review pattern), parking the first expensive work item until a billing check clears, with atomic credit consumption via the same UPDATE...RETURNING idiom as `claim_work_item`; (2) maintenance-run gate — a join-filter added to the three heartbeat selection queries requiring `maintenance_active`, valuable even before any domain is sold.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§8, stripe/001commentary.md#final turns
- **relations:** Two-plane billing architecture; Ownership hierarchy reuse for entitlement scoping; One-off credit vs recurring subscription billing model
- **verify-later:** site_work_items.approval_mode values; build-pipeline-trigger/improvement-loop/content-feed-trigger selection SQL

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two-plane billing architecture (auth-service truth + chassis entitlement cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §3: "Truth = auth DB, mutated only by webhooks... Gate reads = chassis client_entitlements cache, fed one-directionally from auth" — all proposed, table marked PROPOSED
- **what:** Splits billing across two databases/services with one directional bridge: the auth service owns billing truth (subscriptions, credits, webhook-driven events); the chassis reads a local `client_entitlements` cache table fed by an entitlement-changed Kafka event plus a reconciliation sweep backstop — required because the maintenance heartbeat must join across thousands of sites per tick.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§3,§5
- **relations:** Entitlement gate architecture; Isolated chat/satellite architecture (Y-copy); Pluggable billing provider abstraction
- **verify-later:** proposed table client_entitlements; entitlement-changed Kafka event/consumer (not built)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Pluggable billing provider abstraction (Stripe as implementation #1)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §4 gives a full Go interface sketch explicitly labelled "Sketch, not final"; current code has "no Stripe SDK" imported at all
- **what:** A `Provider` interface (`EnsureCustomer`, `CreateSubscriptionCheckout`, `CreateOneOffCheckout`, `CreatePortalSession`, `CancelSubscription`, `ParseWebhook`) behind which Stripe is the first implementation, normalising provider-specific webhook payloads into a provider-agnostic `Event` type. Justified as "zero retrofit cost" specifically because no Stripe integration exists yet.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§4,#TL;DR
- **relations:** Two-plane billing architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Webhook-as-only-source-of-truth billing principle
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "service.go imports only uuid, zap, time, context, fmt — no Stripe SDK. CreateSubscription is a bare DB insert that sets Status = 'active' with no payment step... There's no webhook handler anywhere." Confirmed by direct code read
- **what:** The organising design principle for the billing plan: client-side success redirects must never grant entitlement — only a signature-verified Stripe webhook, deduplicated by `provider_event_id`, may mutate entitlement state. Directly motivated by the audited finding that today `status = active` merely means "a row exists" with zero payment verification.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§2,#Appendix, stripe/001commentary.md#Stripe audit turn
- **relations:** Existing but non-functional auth-service subscription scaffold; Two-plane billing architecture
- **verify-later:** internal/auth-service/subscription/service.go

<!-- SOURCE: U13_docs024_small_dirs.md -->
### One-off credit vs recurring subscription billing model
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** PLAN §7 and §5 `billing_credits` DDL are both marked PROPOSED; no credit ledger exists in code
- **what:** Two distinct charge shapes: recurring (maintenance/tier subscription, reusing the existing but non-functional subscription scaffold) and one-off (the $5-per-site build and first-site-free grant, modelled as a `billing_credits` ledger — granted/consumed counts per client). Build proceeds only once a credit is atomically consumed via the entitlement gate.
- **sources:** stripe/PLAN_stripe_billing_integration.md#§5,§7, stripe/001commentary.md#pricing discussion turn
- **relations:** Entitlement gate architecture; Existing but non-functional auth-service subscription scaffold
- **verify-later:** proposed billing_credits, billing_events tables (auth DB)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Existing but non-functional auth-service subscription scaffold
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "GetUsageStats returns hardcoded zeros with the comment 'returning mock data,' which makes CheckQuota always pass... repository.go mixes ? (MySQL-style) placeholders... with $1 (Postgres-style)... a strong sign this module has never actually been exercised." (stripe/PLAN_stripe_billing_integration.md#§1,§Appendix); independently confirmed from the tools-workstream side (PLAN_isolated_chat_environment(5).md §13: "Billing is scaffolded, not wired," correcting that same document's own earlier "billing largely exists" assumption)
- **what:** A pre-existing `subscription` package in the auth service (models, repository, service, handlers) with a `subscriptions` table, tier constants (free/basic/premium/enterprise), `stripe_customer_id`/`stripe_subscription_id` columns, a `CheckoutSession` type, and JWT claims already carrying `client_id`+`tier` — all reusable — but verified as not wired: no Stripe SDK import anywhere, `CreateSubscription` is a bare insert with no payment step, no webhook handler exists, `CheckUsage`/`GetUsageStats` returns hardcoded zeros so quota checks always pass, and a placeholder-dialect inconsistency in `repository.go` is a strong sign the module has never run against a live database. Security consequence: any entitlement gate trusting `subscription.status` today only reflects "a row exists," not "payment cleared."
- **sources:** stripe/PLAN_stripe_billing_integration.md#§1,§Appendix, stripe/001commentary.md#Stripe audit turn, tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md#13
- **relations:** Webhook-as-only-source-of-truth billing principle; Pluggable billing provider abstraction; Ownership hierarchy reuse for entitlement scoping; Entitlement gate architecture
- **verify-later:** internal/auth-service/subscription/{models,repository,service,handlers}.go; presence/absence of a Stripe webhook handler

<!-- SOURCE: U22_recent_small_docs.md -->
### Commercial model + entitlement seams (billing adapter)
- **category:** payments
- **status-signal:** partial
- **status-evidence:** "billing/identity is mostly reuse, not new" — the auth service already has a `subscriptions` table with `stripe_customer_id`, tier definitions, JWT carrying client_id+tier; "live checkout-session creation and webhooks were not evident ... verify before relying."
- **what:** The saleability design: operator-primary (operate thousands of domains), vendor-optional (sell a domain + its backend, rarely the whole framework). Isolation unit = the satellite; separability unit = the domain (partition by site_id/domain, extractable + swappable credentials). Seams to honour now: ownership via existing clients→networks→sites hierarchy (re-parent network_id to sell), a pluggable billing adapter (Stripe first, generalise stripe_* columns to provider_*), two entitlement gates (build-submission reusing site_work_items.approval_mode → a pending_entitlement hold; maintenance-run filtering the heartbeat site-selection queries as a cost valve), a saas_cheap-vs-portfolio build-tier riding the existing batch/sync rail, and snapshot-able building blocks for whole-instance sales.
- **sources:** docs025.../PLAN_isolated_chat_environment(4).md#13, docs025.../PLAN_simple_paid_multidomain_chat(1).md#2
- **relations:** auth-service subscriptions, site_work_items.approval_mode, batch processing (scheduled→batch), building-as-a-service
- **verify-later:** auth subscriptions table + Stripe webhook wiring; site_work_items.approval_mode; heartbeat site-selection queries

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### REVIEW_BEFORE_PAY billing flow supersedes charge-first flow
- **category:** payments
- **status-signal:** partial
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)": "Supersedes the older Flow/Email/AUTO_DELIVER notes above where they differ... `REVIEW_BEFORE_PAY` (default on)."
- **what:** idea.uk's original flow charged the customer first (Stripe Checkout), then ran the engine, then optionally held for operator review before emailing (`AUTO_DELIVER`). This was replaced by a `REVIEW_BEFORE_PAY` switch (default on): the operator's `/confirm` now *runs the engine first* and holds the draft for review; only after the operator approves does the buyer get a pay link — no money is taken until a human has seen the actual output. The original charge-first flow is kept as a fallback (`REVIEW_BEFORE_PAY=false`) "if engine cost ever spikes." A click-through token-based approve/decline UI (HMAC per order) was added on top to remove the need for curl+API-key.
- **sources:** `RUNBOOK_idea_uk(10).md` "Status & operating update (2026-06-11)"
- **relations:** idea.uk product; Stripe webhook-as-truth pattern
- **verify-later:** `idea-go/service.go` `REVIEW_BEFORE_PAY` branch

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Stripe webhook-as-truth billing pattern (idea.uk lightweight variant)
- **category:** payments
- **status-signal:** deployed
- **status-evidence:** `RUNBOOK_idea_uk(10).md` "Stripe billing — setup" section: live keys, live webhook destination IDs, "Billing follows the PLAN_stripe_billing_integration.md principles but in the lightweight pay-per-idea shape... proven end-to-end with a real card on 2026-06-14."
- **what:** idea.uk's billing never trusts a browser redirect; only a signature-verified `checkout.session.completed` webhook (deduped by event id) marks an order paid and triggers delivery. Uses a Stripe **restricted API key** scoped to `Checkout Sessions → Write` only (least privilege — no refunds, no customer/product read access needed since Checkout uses inline `price_data`). Refunds are manual-only in the Stripe dashboard (no `/refund` endpoint exists). This is presented explicitly as the lightweight, one-off-payment implementation of the same principles as the full chassis-wide Stripe plan (see separate entry) — webhook-is-truth, idempotent, provider behind an interface (FakeProvider swap for local testing).
- **sources:** `RUNBOOK_idea_uk(10).md` §"Stripe billing — setup" (webhook destination IDs, account IDs, restricted-key scoping, troubleshooting runbook for a real signature-mismatch incident on 2026-06-14)
- **relations:** chassis-wide Stripe billing integration plan (supersedes/generalizes); REVIEW_BEFORE_PAY flow
- **verify-later:** Stripe dashboard accounts `acct_1RNfPY08YuzM2cqf` (test) / `acct_1RNfPL02nQ76FNif` (live)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-wide Stripe billing integration plan (client_entitlements cache)
- **category:** payments
- **status-signal:** aspirational
- **status-evidence:** Doc self-describes as "PROPOSED" throughout ("Schema caveat: this plan is written from the auth subscription Go models, not the auth DB migrations... Every DDL below is PROPOSED"). No claim of implementation for the chassis-wide version (idea.uk implemented its own lighter variant instead — see above).
- **what:** A designed-not-built architecture for platform-wide billing: auth service owns billing truth (subscriptions, one-off credits, webhook-verified events only), chassis reads through a one-directionally-fed cache table `client_entitlements` (never calls auth synchronously from the hot path), with two gates — a low-volume build-submission gate (`approval_mode='pending_entitlement'`) and a high-volume maintenance-run gate (join-filter on heartbeat queries). Covers both recurring (maintenance/tier) and one-off ($5 build credit) charge shapes, a provider interface abstracting Stripe, and a verified-findings appendix showing the existing auth subscription code is a non-functional scaffold (`CreateSubscription` stamps `status=active` with no payment, no Stripe SDK, no webhook handler, mock usage stats, and a `?`/`$1` placeholder-dialect mismatch implying the code was never run against one DB engine).
- **sources:** `docubundle_idea_golive/package_module/output_contexts/PLAN_stripe_billing_integration.md` (packaged context-pack snapshot, 390 lines); archive `PLAN_stripe_billing_integration(2).md` (identical to live `(3).md`)
- **relations:** Stripe webhook-as-truth pattern (idea.uk's lighter realisation of the same principles); isolated-chat-environment commercial model (referenced, live doc)
- **verify-later:** `internal/auth-service/subscription/{models,repository,service,handlers}.go` — confirm whether the scaffold described (no Stripe SDK, mock usage stats, dialect mismatch) is still the current state

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### LIABILITY_AND_TERMS and legal pages (terms, refund, privacy) — AI-disclosure requirement
- **category:** NEW:legal-and-compliance
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "/terms and /refund-policy pages written + served," "privacy policy added; terms hardened (AI disclaimer)" — all shipped as live routes, explicitly flagged as drafts pending a "~£200-500 fixed-fee UK solicitor review needed before going live."
- **what:** Three plain-language legal pages built directly into the idea.uk Go binary (`termsBody`/`refundBody`/`privacyBody` constants, `{{EMAIL}}` templated at serve time): terms (explicitly states reports are AI-generated and AI "can be confidently wrong and invent facts... treat everything as to-be-checked... entirely your responsibility and not ours"), refund policy (14-day no-reason refund plus fault/non-delivery refund), and a UK-GDPR-shaped privacy policy naming processors (Stripe, Anthropic) and flagging the US data-transfer point. Grew out of a liability analysis (`LIABILITY_AND_TERMS.md`) triggered directly by the Risk-column near-miss (SFI single-farm assessment) — identifies the real legal exposure as common-law negligent misstatement (Hedley Byrne) rather than any formal regulatory regime, since SFI navigation itself isn't formally regulated.
- **sources:** `running_notes(44).md` (three consecutive 2026-06-05 checkpoints)
- **relations:** Risk-as-hazard scoring dimension (the trigger); idea.uk product
- **verify-later:** whether solicitor review has actually happened; `/terms`, `/refund-policy`, `/privacy` routes live

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### LIABILITY_AND_TERMS and legal pages (terms, refund, privacy) — AI-disclosure requirement
- **category:** NEW:legal-and-compliance
- **status-signal:** deployed
- **status-evidence:** `running_notes(44).md`: "/terms and /refund-policy pages written + served," "privacy policy added; terms hardened (AI disclaimer)" — all shipped as live routes, explicitly flagged as drafts pending a "~£200-500 fixed-fee UK solicitor review needed before going live."
- **what:** Three plain-language legal pages built directly into the idea.uk Go binary (`termsBody`/`refundBody`/`privacyBody` constants, `{{EMAIL}}` templated at serve time): terms (explicitly states reports are AI-generated and AI "can be confidently wrong and invent facts... treat everything as to-be-checked... entirely your responsibility and not ours"), refund policy (14-day no-reason refund plus fault/non-delivery refund), and a UK-GDPR-shaped privacy policy naming processors (Stripe, Anthropic) and flagging the US data-transfer point. Grew out of a liability analysis (`LIABILITY_AND_TERMS.md`) triggered directly by the Risk-column near-miss (SFI single-farm assessment) — identifies the real legal exposure as common-law negligent misstatement (Hedley Byrne) rather than any formal regulatory regime, since SFI navigation itself isn't formally regulated.
- **sources:** `running_notes(44).md` (three consecutive 2026-06-05 checkpoints)
- **relations:** Risk-as-hazard scoring dimension (the trigger); idea.uk product
- **verify-later:** whether solicitor review has actually happened; `/terms`, `/refund-policy`, `/privacy` routes live

<!-- SOURCE: U04_idea_uk.md -->
### Liability framework: risk-tiered mitigations, disclaimers, PII, and draft T&Cs
- **category:** NEW:legal-liability
- **status-signal:** partial
- **status-evidence:** /terms, /refund-policy, /privacy live on the service (2026-06-05) with AI-can-be-wrong wording; solicitor review explicitly still pending (A6 open); PII quote a kickoff item, no policy on file evidenced.
- **what:** The full liability posture for AI-analysis products: negligent misstatement (Hedley Byrne) named as the real exposure route; disclaimers must be conspicuous and *proximate* (in the report itself, top-of-report box, not just site footer); every claim cited + date-stamped, versioned corpus as audit trail, 6-year input/output retention; operator review of early deliveries; generous visible refunds; PII insurance with the AI-assisted-human-reviewed framing disclosed honestly; limited company for payment. Draft starter T&Cs for both idea.uk (information-not-advice, liability capped at fee) and the sharper SFI product (verify-before-acting obligations, exclusions list). Policy pages served from string constants through the brand wrapper with an {{EMAIL}} token; UK-GDPR privacy naming Stripe/Anthropic as processors. The mission's "never verdicts → opinion+evidence+questions" framing is the same posture applied to site content.
- **sources:** idea.uk/LIABILITY_AND_TERMS(2).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 update); idea.uk/DEVELOPMENT_RUNBOOK(3).md#A6-A7
- **relations:** operator-risk column (what triggers "needs liability work"); idea.uk mission; SFI Diff Alerts.
- **verify-later:** live /terms content; any PII policy record.

<!-- SOURCE: U04_idea_uk.md -->
### Liability framework: risk-tiered mitigations, disclaimers, PII, and draft T&Cs
- **category:** NEW:legal-liability
- **status-signal:** partial
- **status-evidence:** /terms, /refund-policy, /privacy live on the service (2026-06-05) with AI-can-be-wrong wording; solicitor review explicitly still pending (A6 open); PII quote a kickoff item, no policy on file evidenced.
- **what:** The full liability posture for AI-analysis products: negligent misstatement (Hedley Byrne) named as the real exposure route; disclaimers must be conspicuous and *proximate* (in the report itself, top-of-report box, not just site footer); every claim cited + date-stamped, versioned corpus as audit trail, 6-year input/output retention; operator review of early deliveries; generous visible refunds; PII insurance with the AI-assisted-human-reviewed framing disclosed honestly; limited company for payment. Draft starter T&Cs for both idea.uk (information-not-advice, liability capped at fee) and the sharper SFI product (verify-before-acting obligations, exclusions list). Policy pages served from string constants through the brand wrapper with an {{EMAIL}} token; UK-GDPR privacy naming Stripe/Anthropic as processors. The mission's "never verdicts → opinion+evidence+questions" framing is the same posture applied to site content.
- **sources:** idea.uk/LIABILITY_AND_TERMS(2).md; idea.uk/idea_uk_architecture_and_deployment(6).md (2026-06-05 update); idea.uk/DEVELOPMENT_RUNBOOK(3).md#A6-A7
- **relations:** operator-risk column (what triggers "needs liability work"); idea.uk mission; SFI Diff Alerts.
- **verify-later:** live /terms content; any PII policy record.

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Marketing as work items + OpenClaw adapter
- **category:** NEW:marketing
- **status-signal:** aspirational
- **status-evidence:** P1 marketing section entirely future (agents/adapter unbuilt)
- **what:** SEM campaigns, landing pages, email sequences, social content, schema markup, ad copy all as work items with dedicated handler agents; an openclaw-adapter (adapter service, self-hosted) translates structured campaign specs to external platforms (Google/Meta/LinkedIn) and returns metrics; marketing-discovery-agent finds gaps (GBP, schema, page-2 rankings, competitor ads); SEM setup is HITL-gated.
- **sources:** P1#Marketing: SEM, Outbound, and Growth
- **relations:** work-item system extensibility
- **verify-later:** none built

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Marketing as work items + OpenClaw adapter
- **category:** NEW:marketing
- **status-signal:** aspirational
- **status-evidence:** P1 marketing section entirely future (agents/adapter unbuilt)
- **what:** SEM campaigns, landing pages, email sequences, social content, schema markup, ad copy all as work items with dedicated handler agents; an openclaw-adapter (adapter service, self-hosted) translates structured campaign specs to external platforms (Google/Meta/LinkedIn) and returns metrics; marketing-discovery-agent finds gaps (GBP, schema, page-2 rankings, competitor ads); SEM setup is HITL-gated.
- **sources:** P1#Marketing: SEM, Outbound, and Growth
- **relations:** work-item system extensibility
- **verify-later:** none built

<!-- SOURCE: U21_legacy_docs_b.md -->
### SEO content agent
- **category:** NEW:seo
- **status-signal:** aspirational
- **status-evidence:** docs017/019b "seo-content-agent | LLM for generation, algorithmic for validation | New — runs after page content is written"; seo-discovery-agent in maintenance Phase 0; slot exists in component-builder-v2 sketch.
- **what:** A post-content sweep owning meta titles/descriptions, structured data/JSON-LD, robots directives, canonical URLs and Open Graph across all pages, with algorithmic validation and LLM generation; complemented in maintenance by sitemap-sync, schema validation, and meta-freshness discovery plus sitemap-regenerator and schema-fixer fix agents. No dedicated SEO category exists in the current taxonomy despite recurring SEO responsibilities across eras.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#SEO-Content-Agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Fix-Agents
- **relations:** meta-manager (docs018/007); link technical types; site-finalizer sitemap generation.
- **verify-later:** any seo agent definitions; sitemap.xml generation code path.

<!-- SOURCE: U21_legacy_docs_b.md -->
### SEO content agent
- **category:** NEW:seo
- **status-signal:** aspirational
- **status-evidence:** docs017/019b "seo-content-agent | LLM for generation, algorithmic for validation | New — runs after page content is written"; seo-discovery-agent in maintenance Phase 0; slot exists in component-builder-v2 sketch.
- **what:** A post-content sweep owning meta titles/descriptions, structured data/JSON-LD, robots directives, canonical URLs and Open Graph across all pages, with algorithmic validation and LLM generation; complemented in maintenance by sitemap-sync, schema validation, and meta-freshness discovery plus sitemap-regenerator and schema-fixer fix agents. No dedicated SEO category exists in the current taxonomy despite recurring SEO responsibilities across eras.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#SEO-Content-Agent; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Fix-Agents
- **relations:** meta-manager (docs018/007); link technical types; site-finalizer sitemap generation.
- **verify-later:** any seo agent definitions; sitemap.xml generation code path.

<!-- SOURCE: U19_sql_tables_components.md -->
### Affiliate and products domain
- **category:** NEW:affiliate-and-products
- **status-signal:** partial
- **status-evidence:** Full schema in 004 (products, product_assets, affiliate_programs, affiliate_products, link_registry.affiliate_product_id + requires_disclosure); 043 (2026) still references "the affiliate_products resolver" as the source of product imagery, so the domain is alive but no seeds/operations appear in this unit.
- **what:** Commerce layer: first-party products (pricing incl. price_display "From £99", SEO fields, per-site slug uniqueness) with asset junctions; affiliate networks (tracking param templates, commission terms, API refs) and affiliate_products with cached network data + custom editorial overlay (pros/cons/verdict/rating, content_status cached→enhanced→reviewed) and availability checking. Link registry marks affiliate links and FTC/ASA disclosure requirements.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART7-11; docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql#kind-comment
- **relations:** link-management (registry); product-card/product-grid library components; imagery (product images excluded from planner).
- **verify-later:** affiliate_products resolver code; any populated programs.

<!-- SOURCE: U19_sql_tables_components.md -->
### Affiliate and products domain
- **category:** NEW:affiliate-and-products
- **status-signal:** partial
- **status-evidence:** Full schema in 004 (products, product_assets, affiliate_programs, affiliate_products, link_registry.affiliate_product_id + requires_disclosure); 043 (2026) still references "the affiliate_products resolver" as the source of product imagery, so the domain is alive but no seeds/operations appear in this unit.
- **what:** Commerce layer: first-party products (pricing incl. price_display "From £99", SEO fields, per-site slug uniqueness) with asset junctions; affiliate networks (tracking param templates, commission terms, API refs) and affiliate_products with cached network data + custom editorial overlay (pros/cons/verdict/rating, content_status cached→enhanced→reviewed) and availability checking. Link registry marks affiliate links and FTC/ASA disclosure requirements.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART7-11; docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql#kind-comment
- **relations:** link-management (registry); product-card/product-grid library components; imagery (product images excluded from planner).
- **verify-later:** affiliate_products resolver code; any populated programs.

<!-- SOURCE: U10_imagery.md -->
### Affiliate sites programme and the query.affiliate_products resolver gap
- **category:** NEW:affiliate-commerce
- **status-signal:** aspirational
- **status-evidence:** "This is not the active workstream right now — a holding doc" (2026-05-12); affiliate_products "Zero rows today"; resolver "a wired socket with no plug".
- **what:** The affiliate vision (boxing tickets, darts gear, lead-gen) with three vertical shapes (pure-product / event-ticket / lead-generation) and a layered build path (one product on one page → ingestion + editorial enrichment → imagery via illustrations → event/lead verticals). Substantial scaffolding exists — affiliate_products/affiliate_programs tables, five product components (product-card-with-cta declares `source: query.affiliate_products` with typed image_url; product-specs schema effectively empty), link_registry disclosure flags, the med-* scraper family as an ingestion model — but no program integration, no resolver populating the declared source, no editorial pipeline, no calendar/event infrastructure.
- **sources:** old/STATUS_affiliate_sites_2026-05-12.md, STATUS_imagery_2026-05-12.md#Component-audit-finding, FOCUS_imagery_assessment_1_.md#3.2
- **relations:** product illustration plugs in as a resolver precedence rule; link-management (doc 024); vet-med-pricing med-* agents as pattern.
- **verify-later:** affiliate_products row count; any resolver handling query.affiliate_products in queryresolve/sourceResolver.

<!-- SOURCE: U10_imagery.md -->
### Affiliate sites programme and the query.affiliate_products resolver gap
- **category:** NEW:affiliate-commerce
- **status-signal:** aspirational
- **status-evidence:** "This is not the active workstream right now — a holding doc" (2026-05-12); affiliate_products "Zero rows today"; resolver "a wired socket with no plug".
- **what:** The affiliate vision (boxing tickets, darts gear, lead-gen) with three vertical shapes (pure-product / event-ticket / lead-generation) and a layered build path (one product on one page → ingestion + editorial enrichment → imagery via illustrations → event/lead verticals). Substantial scaffolding exists — affiliate_products/affiliate_programs tables, five product components (product-card-with-cta declares `source: query.affiliate_products` with typed image_url; product-specs schema effectively empty), link_registry disclosure flags, the med-* scraper family as an ingestion model — but no program integration, no resolver populating the declared source, no editorial pipeline, no calendar/event infrastructure.
- **sources:** old/STATUS_affiliate_sites_2026-05-12.md, STATUS_imagery_2026-05-12.md#Component-audit-finding, FOCUS_imagery_assessment_1_.md#3.2
- **relations:** product illustration plugs in as a resolver precedence rule; link-management (doc 024); vet-med-pricing med-* agents as pattern.
- **verify-later:** affiliate_products row count; any resolver handling query.affiliate_products in queryresolve/sourceResolver.
