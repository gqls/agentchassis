
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
