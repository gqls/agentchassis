# Register — site-spec-and-classifier

22 concepts, consolidated from 60 raw extractions (the cluster input file contains
the entire site-spec-and-classifier block set duplicated exactly twice — 30 unique
raw blocks appearing twice each — plus further cross-unit duplication of the same
concepts under different titles) across units U01, U04, U05, U14, U16, U17a, U18,
U19, U20, U21, U22, U23, U24e, U25.

### SPEC-001 — Dream spec / gap analysis / feasibility: one spec with per-item status, not two documents
- **status:** aspirational
- **status-evidence:** 021 resolved decision 24 "One spec, not two"; 028: per-item status "not fully implemented yet — Phase 2"; an independent, archived proposal (docs022) reached the identical design and was never merged in.
- **what:** The full spec IS the dream; items carry status deployed/planned/blocked. Gap analysis is the blocked/planned subset; a feasibility-recheck task promotes blocked→planned when capability arrives, preventing impossible work items from being emitted. The design was proposed at least twice independently (021/028's aspect model, and a separate archived docs022 "unified spec" proposal covering classification/identity/design_intent/content_direction/pages/features/SEO/maintenance) before either shipped the mechanical per-item status Phase 2 — the older 002d dream_spec-in-content_data shape is superseded by this.
- **sources:** 021#One Spec, Not Two; 028#The spec has status; 002d#Gap Analysis; docs022.../old/004_classifier_notes.md
- **relations:** fidelity dial (SPEC-003); feasibility-recheck task; site_specs aspect store (SPEC-002)
- **verify-later:** does feasibility-recheck scheduled task exist; per-item status columns on site_specs

### SPEC-002 — Site spec unification: site_specs aspect-versioned store as the one authoritative spec
- **status:** deployed
- **status-evidence:** Table created in the Phase-0 migration (019); a live pg_dump backup (bk_site_specs.sql) shows the production shape including `pinned`, with extensive backfills for real sites — more specific and dated than the 021 target-state doc's "partial" framing.
- **what:** All strategic site specification lives as (site_id, aspect, data JSONB) rows — classification, identity, strategy, tone, design_intent, content_direction, growth_config, adoption_source, etc. — with provenance (source enum: classifier/adoption/hitl/planner/improvement/seed/manual/rollback/fork/recovery; source_agent; source_item_id) and history via is_current + superseded_at (unique current row per site+aspect). `write_site_spec` deep-merges partials so every row is self-contained and pruning-safe; `pinned` (Phase 4) prevents agents overriding human-set specs. Classifier writes intent, planner implements, design agent executes, audits enforce; content_data is legacy and read_site_spec falls back to it.
- **sources:** 021 full; P1#Site Specification System; docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql; docs/agent_docs/sql_for_tables/bk_site_specs.sql
- **relations:** doc 028 ownership model (SPEC-004); strategic-vs-plan-time split (doc 030); dream spec / per-item status (SPEC-001)
- **verify-later:** write_site_spec action; pinned enforcement in writers; backfill completeness for legacy sites

### SPEC-003 — Fidelity dial (locked/high/medium/low + no-adoption confidence mode)
- **status:** partial
- **status-evidence:** 028: "Fidelity … Five values, with high as the default when adoption evidence is present"; "Phase 1 current (implicit high); explicit fidelity input is Phase 3; depends on per-item spec status (Phase 2)".
- **what:** A dial controlling how much aspirational extension reaches the first build vs how faithfully it matches the strongest evidence (usually adoption): locked = adopted-exact/no promotions, high (default with adoption) = faithful launch with ~one substantive change per audit cycle, medium = modest extensions, low = adoption as inspiration; blank domains reinterpret it as research-confidence tolerance. Only Phase 1 (implicit high, coarse prompt behaviour) exists; it is meant to live on the trigger input plus a build_policy/adoption_meta aspect once built.
- **sources:** 028#Fidelity, #Phased implementation; WM/028_platform_mission_and_pipeline_direction.md#fidelity-controlling-how-much-aspiration-reaches-the-first-build
- **relations:** the adoption-pipeline's identical dial from the other side of the same doc (adoption-pipeline ADO-011); spec-has-status (SPEC-001); classifier-as-strategic-brain (SPEC-011)
- **verify-later:** any fidelity/build_policy aspect in prod

### SPEC-004 — Spec aspect ownership and read-and-extend (anti-silent-overwrite)
- **status:** deployed
- **status-evidence:** 028 ownership list; adoption-aware classifier prompt (migration 006) live; open question notes the planner still writes design_intent pre-030.
- **what:** Classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planners own site_plan. The classifier over adoption output is read-and-extend (preserve adopted dimensions, add strategy), never overwrite. Named failure modes: silent overwrite, confident fabrication on thin signal, default-to-brochure, reflexive upstream re-runs, schema-level commercial bias, adoption without strategic analysis.
- **sources:** 028#Who writes what, #Failure modes
- **relations:** doc 030 planner redirect to directives; composition self-heal; classifier-as-strategic-brain (SPEC-011)
- **verify-later:** confirm build-site-planner no longer writes design_intent/content_direction post-030

### SPEC-005 — Superseding a spec doesn't undo installed artefacts (re-queue rule)
- **status:** deployed
- **status-evidence:** 028: gamesdesign remediation hit it in practice (fresh specs written but a stale installed theme remained pinned at sites.style_collection_id).
- **what:** Agents with install side-effects (composition, nav, pages, assets) write beyond site_specs and leave live pointers in long-lived tables; invalidating their spec must also queue the re-run work item (needs_composition etc.), and long-term the supersession itself should emit the recovery item automatically. Test for whether a writer needs this: does the agent write other tables AND does a live pointer reference the write?
- **sources:** 028#Failure modes (last)
- **relations:** install_site_composition; composition trigger matrix (doc 027)
- **verify-later:** whether supersession-emits-item was ever built

### SPEC-006 — Structured design_intent from the classifier (palette + typography reference_values)
- **status:** deployed
- **status-evidence:** Migration applied 2026-06-20; "Palette migration proven on a real build" 2026-06-21 — fresh idea.uk resolved `palette_source=design_intent_values`, parchment, "no invented blue".
- **what:** Root cause of generic-looking fresh builds: the classifier wrote design_intent colours as prose (hex buried in colour_mood sentences) while every consumer — the composition cascade and the analyze_design prompt — reads structured `design_intent.palette.reference_values` (8 slots) + `typography.reference_values`. The migration edited the classifier's classify_and_extract schema and added a MANDATORY-fields bullet (all 8 slots as hex; style_direction must agree with the palette; never default to blue-and-grey) via snapshot_agent backup + exact-anchor replace() with a self-check.
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md
- **relations:** two-stage design pipeline; mandatory-full overlay bug (this is its precondition fix); prompt-migration discipline
- **verify-later:** live classifier prompt contains the palette block; site 1244516d design_intent spec

### SPEC-007 — Phase 0 classifier-only positioning read
- **status:** deployed
- **status-evidence:** "Phase 0 result (2026-06-14 — ran, with the live site up)": faithful identity/classification/content_direction/design_intent specs, interactive-platform at ~0.91 confidence.
- **what:** Running just domain-research-classifier on a domain is a near-zero-cost positioning brief before committing to a full build — its four spec aspects answer "what does this site do for a stranger?". Codified caveats: a fresh read is not blank-slate (the classifier scrapes the live site up to 3 pages unless an adoption already ran); a generic name yields a generic name-only read; a safe suppression trick exists (temporary blank nginx `location = /`, never touching DNS/nginx wholesale with a live Stripe webhook); the classifier's terminal needs_strategy item will flow into a full build if dispatch is running.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 0); idea.uk/HANDOFF_chassis_site.md
- **relations:** standing-ambition mission; fresh vs adoption convergence (adoption-pipeline ADO-010)
- **verify-later:** idea.uk site 97ed2f64 spec rows (incl. the duplicate-spec tidy-up)

### SPEC-008 — Build-standard classifier migration (best-in-class quality/fit, not scope)
- **status:** partial
- **status-evidence:** "READY, PROVEN-CORRECT, NOT YET APPLIED… replace() simulated against the live prompt → lands cleanly" (2026-06-21); still unapplied in TODO P2 (2026-06-26).
- **what:** A migration prepending a "Build standard" block to the classifier prompt: every build should aim at best-in-class quality and fit for its vertical — explicitly not scope inflation. Its first draft failed on a multi-line anchor (whitespace mangling) and was fixed to a single-line anchor with a proven-clean rollback, feeding the prompt-migration discipline.
- **sources:** idea.uk/HANDOFF(13).md (Migrations status); idea.uk/TODO_chassis_and_idea_uk(1).md#P2; idea.uk/running_notes(63).md
- **relations:** standing-ambition default; prompt-migration discipline
- **verify-later:** whether migration_classifier_build_standard.sql was later applied

### SPEC-009 — Guide as a first-class page_type (classifier vocabulary + canonical URLs)
- **status:** deployed
- **status-evidence:** "guides typed page_type=guide directly … migration_adoption_add_guide_page_type.sql worked" (2026-06-05 re-adoption).
- **what:** Guides were folded into blog-post by the site-adoption-agent's analyze_site enum, defeating `query.pages_where_type:guide` list resolution. The structural fix over the band-aid: add `guide` to the adoption classifier enum + guidance, re-type existing rows, and migrate URLs from /blog/guide-<slug>.html to /guides/<slug>/index.html. Classification geography documented alongside it: analyze_site (LLM) emits per-page page_type+url; site-classifier is site-type only; build-site-planner has a staler vocabulary but preserves existing pages verbatim; pages.page_type has only a kebab-format CHECK, no value allowlist.
- **sources:** running_notes_14(26).md#part-13-14c
- **relations:** Tier-D lists; bare-duplicate cleanup; same fix described from the adoption-mechanism side (adoption-pipeline ADO-015)
- **verify-later:** site-adoption-agent analyze_site prompt enum; pages.page_type values on a fresh adoption

### SPEC-010 — site_type taxonomy drift between classifier and strategist, and the consolidation queue
- **status:** partial
- **status-evidence:** builder_route(21) §B2: "vocabulary drift between hops — the classifier's site_type set (brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social) vs the strategist's canonical set (brochure|authority-portal|local-directory|review-site|content-hub|landing-page|portfolio). Two taxonomies for the same concept, one spec chain." Queued as open work, not executed in these files.
- **what:** Two adjacent relay hops use different canonical vocabularies for the same site_type concept flowing through one spec chain — a contract-drift hygiene defect. The queued brief also covers a second overlap: domain-research-classifier (relay-native) vs site-classifier (pageflow/intake path, which doesn't use work-item triage and keys hitl_confirm_type off confirmed_type.recommended_builder) — diff both rows, map dependency points, merge additions both ways with snapshot migrations, deprecate only at zero usage.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2 (Q8), #queue (item 3); HANDOFF_builder_thread.md#3; README_flows.md (Q8 note)
- **relations:** two front doors question; workflow result contract (same drift class); classifier lineage (SPEC-012)
- **verify-later:** classifier and strategist prompt enumerations; both classifier rows; intake usage query on orchestration_states

### SPEC-011 — Classifier as strategic brain (always runs full)
- **status:** partial
- **status-evidence:** 028: "The classifier … runs on every site entering the pipeline, and it always does its full job … Adoption does not shortcut it"; Phase 1 current, Phases 2–5 not implemented.
- **what:** The domain-research-classifier decides what a site should be on every site; adoption and operator-mission inputs are weighted, not bypasses. It is not constrained to current capability — best-version items it can't build yet are marked `blocked` for feasibility-recheck rather than silently dropped. Silent override is the named failure mode this guards against.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-classifier-is-the-strategic-brain, #input-sources-and-their-weight, #phased-implementation
- **relations:** website mission; fidelity dial (SPEC-003); spec-has-status (SPEC-001); adoption → classifier handoff (adoption-pipeline ADO-006)
- **verify-later:** domain-research-classifier agent_definition; migration 006

### SPEC-012 — Classifier lineage: site-classifier v1 (single Haiku label) → v2 (research-backed domain_profile) → domain-research-classifier
- **status:** deployed
- **status-evidence:** 049 header documents the current agent's pipeline position "first agent after seed_build_queue"; 067 adds an extended-thinking budget to its classify_and_extract step — the most recent, most specific evidence for what's live today, superseding the two earlier forms.
- **what:** Classification evolved through three generations. v1 (docs006/007) was a single Haiku call mapping domain+objective to {landing, content, portfolio, brochure} + recommended_builder, with explicit responsibility fences (never picks pages or style_collection). v2 turned it into an orchestrator (Haiku research brief → research-agent web investigation → Sonnet synthesis) producing a backward-compatible site_type plus a rich domain_profile (business identity, tone, visual_direction, image_guidance, strategic analysis). The current domain-research-classifier is the work-item-pipeline successor: it handles needs_domain_research, researches via web search and scrape, classifies site type, extracts identity signals, writes the identity/classification site_specs aspects, and creates the next work item (needs_briefing, later needs_strategy).
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; 003_site_classifier.sql; 049_domain_research_classifier.sql; 060_domain_strategist.sql; 067_implement_extended_thinking_not_yet_implemented.sql; docs006_workflow_builder/003_current_state_of_agents.sql
- **relations:** domain_profile as ancestor of site_specs aspects; site_type taxonomy drift (SPEC-010); intake orchestrator (SPEC-015)
- **verify-later:** live site-classifier definition vs domain-research-classifier; who consumes domain_profile today; current next-item wiring (strategy vs briefing)

### SPEC-013 — spec-updater (mechanical site_specs merge from findings)
- **status:** unknown
- **status-evidence:** 072 definition; no later patches confirming live behaviour appear in this unit.
- **what:** Handler for needs_spec_update items: applies {aspect, field, suggested_value} to site_specs with the WriteSiteSpecAction versioning pattern, with no LLM involved. Description-only items complete as "needs human review" — "the complexity is in the Go action, not in the workflow."
- **sources:** 072_spec_updater_agent.sql
- **relations:** content-gap-planner and audits emit its items; site_specs supersede-versioning (SPEC-002)
- **verify-later:** update_site_spec_from_item action

### SPEC-014 — Specialist architects per site type (legacy)
- **status:** partial
- **status-evidence:** 023 SQL: landing-page-architect created (renamed copy of site-component-architect), content-site-architect created with content-specific components, portfolio-architect created "for future use".
- **what:** One architect agent per site type, each with its own default sections and component_category filter into the library; the alternative "one architect, differentiated by build plan" was debated and the group-per-project-type model won conceptually at the time.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** groups as project recipes; conditional_call_agent routing
- **verify-later:** the three architect rows; content-site component rows

### SPEC-015 — Intake orchestrator workflow (classify → brief → spawn builder) — legacy
- **status:** superseded
- **status-evidence:** docs006/011: "✅ WORKING SYSTEM: HITL Orchestration with Multi-Agent Workflow" listing the 11-step intake workflow; docs017/023 later introduces intake-orchestrator-v2 routing.
- **what:** The entry-point orchestration: spawn/call site-classifier → fetch_available_builders from DB → HITL confirm site type (human can override classifier and builder choice) → fetch builder questionnaire → briefing agent fills it → HITL review brief → spawn and call the chosen builder. Established the pattern of human quality-gates before expensive generation, ancestor of the current relay/work-item pipeline.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#Working-Agents; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Naming-and-Coexistence
- **relations:** classifier lineage (SPEC-012); briefing agent; HITL protocol
- **verify-later:** intake-orchestrator workflow JSON in agent_definitions; whether v2 routing exists

### SPEC-016 — Feasibility / blocked-handler pattern
- **status:** partial
- **status-evidence:** "the claim action catches it ... Item marked 'blocked', error='Handler agent not registered' ... weeks later ... Feasibility-recheck task finds it ... promoted to 'triaged'."
- **what:** A pattern where planners/discovery emit work items naming an intended handler even if that agent doesn't exist yet. The claim action checks agent_definitions; unknown handlers mark the item `blocked` with an error, and a periodic feasibility-recheck task promotes blocked items to `triaged` once the handler is deployed. A Go `check_feasibility` action can also pre-block at spec time based on the agent registry.
- **sources:** docs022.../old/004_classifier_notes.md#feasibility-assessment
- **relations:** dream spec / per-item status (SPEC-001); work-item lifecycle; tool-lifecycle
- **verify-later:** claim action handler-existence check; feasibility-recheck scheduled task

### SPEC-017 — write_site_spec spec_data string coercion (bugfix)
- **status:** deployed
- **status-evidence:** RUNBOOK_vonc_session: "FIXED, deployed"; migration table: code deployed 2026-06-24 ~15:00.
- **what:** `WriteSiteSpecAction` hard type-asserted `spec_data` to a map, rejecting the plain-string mission_brief/roadmap_brief the domain-submitter workflow resolves ("spec_data must be a JSON object, got string"). Fix: a coercion block — JSON string → parse; plain string → wrap as `{"text": value}` (matching the classifier prompt's read pattern); objects pass through unchanged.
- **sources:** docs/RUNBOOK_vonc_session(1).md#1; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc(36).md#1
- **relations:** handoff document convention; data-shape/contract-drift debugging family; mission+roadmap aspects (SPEC-021)
- **verify-later:** platform/orchestration/actions/site_spec_actions.go WriteSiteSpecAction coercion block

### SPEC-018 — Chassis-native idea engine (Phase D / Layer 4)
- **status:** aspirational
- **status-evidence:** "the chassis version is one idea-orchestrator agent + one workflow reusing these [existing actions], NOT a port of engine.go. Did NOT write the SQL — needs a schema pass first."
- **what:** A mapped-but-unbuilt plan to express the idea-generation method as chassis actions instead of the standalone Go/Python engine: execute_llm_prompt for generate/cut/verify/score, web_search/scrape_web/firecrawl_* for verify, and request_human_input/create_approval_request/await_approval/process_approval_decision for the operator confirm+review gate, identified explicitly as "literally HITL." Distinguishes Shape A (the site IS the service, like idea.uk) from Shape B (a static request-a-report page posting to one central service), since the engine is server-side and minutes-long.
- **sources:** running_notes(44).md ("Wrote the architecture & deployment guide; clarified hosting + OpenAI")
- **relations:** idea generation method (site-case-studies CASE-011); HITL; tool-lifecycle
- **verify-later:** whether an idea-orchestrator agent_definition or workflow exists

### SPEC-019 — Email identity in site_spec — deterministic address encoding + per-site `email` aspect
- **status:** aspirational
- **status-evidence:** "FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md... Proposed `email` data... so a FUTURE email-provisioner agent... can create per-domain forwarders."
- **what:** A proposed platform-wide convention for how any generated site gets an inbound/outbound email identity: a deterministic encoding (lowercase domain, `.`→`-`, `@<operator-domain>`), stored rather than derived-on-read to allow per-site overrides and handle rare collisions, plus a new `email` aspect on site_specs carrying status/address/from/reply_to/provider/forwards_to, reusing the existing deployed/planned/blocked + feasibility-recheck state machine.
- **sources:** running_notes(44).md ("FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md")
- **relations:** site_specs aspect list (SPEC-002); catch-all email forwarding, abandoned (SPEC-020)
- **verify-later:** whether `email` was actually added to the aspect list; EMAIL_identity_in_site_spec.md (live doc)

### SPEC-020 — Catch-all email forwarding — abandoned in favour of specific per-site forwarders
- **status:** abandoned
- **status-evidence:** "CHECKPOINT 2026-06-06 — inbound test FAILED (No Such User Here): catch-all not catching"; and again "inbound still bouncing... root cause = Default Address not forwarding" — two consecutive real-world failures.
- **what:** The initial plan used a domain-level catch-all (cPanel Default Address / "Forward All Email for a Domain") so any encoded address would work without per-site setup. This repeatedly bounced because the mail backend only routes truly-unmatched addresses through the default address, which was itself misconfigured. Design refinement: prefer specific per-site forwarders (created when a site is published) over a server catch-all — only forward addresses that exist, no backscatter.
- **sources:** running_notes(44).md (two consecutive checkpoints, 2026-06-06)
- **relations:** email identity in site_spec, the design this feeds back into (SPEC-019)
- **verify-later:** current leopardess.uk cPanel Default Address / Forwarders configuration

### SPEC-021 — Mission + roadmap as site_specs aspects (strategy-driven site intake)
- **status:** deployed
- **status-evidence:** 004_submit_vonc_trigger.sh ("Tier 3 submission: domain + mission + roadmap + briefs") exists and vonc.com was built from it; 003d specifies persist_mission/persist_roadmap via the existing write_site_spec action.
- **what:** Strategic context travels as input_data.mission (positioning, differentiators, tone, target users, core concepts, measurable objectives) and input_data.roadmap (phases with per-page purpose, section_types and content_context), persisted to site_specs aspects 'mission' and 'roadmap'. The classifier is told not to discover business type from the domain for mission-driven sites; the planner builds only the current phase and outputs section_types, not component names; content writers draw voice from mission and per-page content_context — explicitly requiring no new tables, no chassis code, no RAG for v1.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Approach, #What-goes-where; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh
- **relations:** component selector/creator; roadmap phase advancement (SPEC-022); vonc.com v1 site
- **verify-later:** site_specs aspects mission/roadmap for vonc; intake-orchestrator/domain-submitter workflow steps

### SPEC-022 — Roadmap phase advancement and automated strategic review
- **status:** aspirational
- **status-evidence:** 003d "Phase advancement (later) … Manual for now"; the earlier 003 v1 sketched the full automated loop ("scheduled agent … compare actuals vs targets … propose phase advancement") which later versions dropped to a one-liner.
- **what:** Phases advance by updating the roadmap aspect (current phase → complete, next → active) and re-triggering planning; measurable objectives in the mission aspect (DAU, completion rates, session duration, share rate) indicate when. The fuller vision — a scheduled strategic-review agent closing strategy → build → measure → adjust — was designed and then deliberately parked, not forgotten.
- **sources:** docs/social001_vonc_tiktok_social/003_spark_strategic_planning_architecture.md#Future-automated-strategic-review; docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Phase-advancement
- **relations:** mission/roadmap aspects (SPEC-021); traffic-analytics (the missing measurement half)
- **verify-later:** any analytics source; scheduler entries for strategic review (expect none)
