
<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dream spec / gap analysis / feasibility (one spec with status, not two documents)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 021 resolved decision 24 "One spec, not two"; 028: per-item status "not fully implemented yet — Phase 2"
- **what:** The full spec IS the dream; items carry status deployed/planned/blocked; gap analysis = blocked/planned subset; feasibility-recheck promotes blocked→planned when capability arrives; feasibility annotations prevent impossible work items. Older 002d dream_spec-in-content_data shape superseded by this. Phase 2 of 028 makes it mechanical.
- **sources:** 021#One Spec, Not Two; 028#The spec has status; 002d#Gap Analysis
- **relations:** fidelity dial; feasibility-recheck task
- **verify-later:** does feasibility-recheck scheduled task exist; per-item status columns

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Site spec unification (site_specs aspects as the one authoritative spec)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 021 target-state doc with immediate/short/medium ordering; backfill + fallback both recommended; newer pipeline already uses aspects
- **what:** One versioned spec per site as independent aspect rows (classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance), is_current/superseded_at per-aspect versioning with source/source_agent/source_item_id provenance; write_site_spec deep-merges so every row is a complete self-contained record (pruning-safe). content_data is legacy; read_site_spec falls back to it. Classifier writes intent, planner implements, design agent executes, audits enforce. The 15 content-strategy questions map onto aspects; deep research is a future classifier enrichment.
- **sources:** 021 full; P1#Site Specification System
- **relations:** 028 ownership; 030 strategic-vs-plan-time; dream spec
- **verify-later:** backfill done for legacy sites; read_site_spec fallback

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fidelity dial (locked/high/medium/low + no-adoption confidence mode)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 028 Phases: Phase 1 current (implicit high); explicit fidelity input is Phase 3; depends on per-item spec status (Phase 2)
- **what:** Controls how much aspiration reaches first deployment vs is deferred to the improvement loop, and the loop's promotion rate. locked = adopted-exact, no promotions; high (default with adoption) = faithful launch, ~one substantive change per audit cycle; medium = modest extensions; low = adoption as inspiration; blank domains reinterpret it as research-confidence tolerance. Lives on the trigger input + a build_policy/adoption_meta aspect.
- **sources:** 028#Fidelity, #Phased implementation
- **relations:** spec status; improvement loop rate
- **verify-later:** any fidelity/build_policy aspect in prod

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec aspect ownership and read-and-extend (anti-silent-overwrite)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028 ownership list; adoption-aware classifier prompt (migration 006) live; open question notes planner still writing design_intent pre-030
- **what:** Classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planners own site_plan. Classifier over adoption output is read-and-extend (preserve adopted dimensions, add strategy), never overwrite. Named failure modes: silent overwrite, confident fabrication on thin signal, default-to-brochure, reflexive upstream re-runs, schema-level commercial bias, adoption without strategic analysis.
- **sources:** 028#Who writes what, #Failure modes
- **relations:** 030 planner redirect to directives; composition self-heal
- **verify-later:** build-site-planner no longer writes design_intent/content_direction post-030

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Superseding a spec doesn't undo installed artefacts (re-queue rule)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028: gamesdesign remediation hit it (fresh specs + stale installed theme at sites.style_collection_id)
- **what:** Agents with install side-effects (composition, nav, pages, assets) write beyond site_specs and leave live pointers from long-lived tables; invalidating their spec must also queue the re-run work item (needs_composition etc.) — long-term the supersession itself should emit the recovery item. Test: does the agent write other tables AND does a live pointer reference the write?
- **sources:** 028#Failure modes (last)
- **relations:** install_site_composition; composition trigger matrix (027)
- **verify-later:** whether supersession-emits-item was ever built

<!-- SOURCE: U04_idea_uk.md -->
### Structured design_intent from the classifier (palette + typography reference_values)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Migration applied 2026-06-20; "Palette migration proven on a real build" 2026-06-21 — fresh idea.uk resolved `palette_source=design_intent_values`, parchment, "no invented blue".
- **what:** Root cause of generic-looking fresh builds: the classifier wrote design_intent colours as **prose** (hex buried in colour_mood sentences) while every consumer — the composition cascade and the analyze_design prompt — reads **structured** `design_intent.palette.reference_values` (8 slots) + `typography.reference_values`. The migration edits the classifier's classify_and_extract schema and adds a MANDATORY-fields bullet (all 8 slots as hex; style_direction must agree with the palette; never default to blue-and-grey), applied via snapshot_agent backup + exact-anchor replace() with a RAISE self-check. This single change is what makes both design stages agree (base = parchment, overlay starts from parchment).
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A + checklist); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md ("Direct consequence" section)
- **relations:** two-stage pipeline; mandatory-full overlay bug (this is precondition for its fix); prompt-migration discipline.
- **verify-later:** live classifier prompt contains the palette block; site 1244516d design_intent spec.

<!-- SOURCE: U04_idea_uk.md -->
### Phase 0 classifier-only positioning read
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** "Phase 0 result (2026-06-14 — ran, with the live site up)": faithful identity/classification/content_direction/design_intent specs, interactive-platform at ~0.91 confidence.
- **what:** Running just domain-research-classifier on a domain as a near-zero-cost positioning brief before committing to a build — its four spec aspects ARE the answer to "what does this site do for a stranger?". Caveats codified: a fresh read is NOT blank-slate (the classifier scrapes the live site up to 3 pages unless an adoption already ran); a generic name yields a generic name-only read, so hiding the live site removes signal not bias; a safe suppression trick exists (temporary blank nginx `location = /` — never touch DNS/nginx wholesale with a live Stripe webhook); the classifier's terminal needs_strategy item will flow into a full build if dispatch is running. Decision: leave the live site up.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 0); idea.uk/HANDOFF_chassis_site.md
- **relations:** standing-ambition mission (the read was faithful-but-backward-looking, which motivated it); fresh vs adoption.
- **verify-later:** idea.uk site 97ed2f64 spec rows (incl. the duplicate-spec tidy-up).

<!-- SOURCE: U04_idea_uk.md -->
### Build-standard classifier migration (best-in-class quality/fit, not scope)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "READY, PROVEN-CORRECT, NOT YET APPLIED… replace() simulated against the live prompt → lands cleanly" (2026-06-21); still unapplied in TODO P2 (2026-06-26).
- **what:** A migration prepending a "Build standard" block to the classifier prompt: every build should aim at best-in-class **quality and fit** for its vertical — explicitly not scope inflation. Its first draft failed on a multi-line anchor (whitespace mangling) and was fixed to a single-line anchor with a rollback proven clean — feeding the prompt-migration discipline. Test plan: fresh build first; confirm an adopted rebuild stays faithful rather than drifting.
- **sources:** idea.uk/HANDOFF(13).md (Migrations status); idea.uk/TODO_chassis_and_idea_uk(1).md#P2; idea.uk/running_notes(63).md (lll/mmm 06-20/21)
- **relations:** standing-ambition default; prompt-migration discipline.
- **verify-later:** whether migration_classifier_build_standard.sql was later applied (file itself lives outside this unit).

<!-- SOURCE: U05_content_quality_linking.md -->
### Guide as first-class page_type (classifier vocabulary + canonical URLs)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14j: "guides typed page_type=guide directly … migration_adoption_add_guide_page_type.sql worked".
- **what:** Guides were folded into blog-post by the site-adoption-agent's analyze_site enum, defeating query.pages_where_type:guide list resolution. Structural fix over band-aid: `guide` added to the adoption classifier enum + guidance, existing rows re-typed, URLs migrated /blog/guide-<slug>.html → /guides/<slug>/index.html (page_canonical.go already had the guide case). Classification geography documented: analyze_site LLM emits per-page page_type+url; site-classifier is site-type only; build-site-planner has a staler vocabulary but preserves existing pages verbatim; pages.page_type has a kebab-format check, no value allowlist.
- **sources:** running_notes_14(26).md#part-13-14c
- **relations:** Tier-D lists; adoption pipeline; bare-duplicate cleanup.
- **verify-later:** site-adoption-agent analyze_site prompt enum; pages.page_type values on a fresh adoption.

<!-- SOURCE: U14_docs019_runbooks.md -->
### site_type taxonomy drift between classifier and strategist (Q8)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B2 "vocabulary drift between hops — the classifier's site_type set (brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social) vs the strategist's canonical set (brochure|authority-portal|local-directory|review-site|content-hub|landing-page|portfolio). Two taxonomies for the same concept, one spec chain." Queued item 3.
- **what:** Two adjacent relay hops use different canonical vocabularies for the same site_type concept flowing through one spec chain — a contract-drift hygiene defect awaiting a one-canonical-set decision plus two snapshot migrations.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2 (Q8); docs019/RUNBOOK_builder_route(21).md#queue (item 3)
- **relations:** two front doors Q5; workflow result contract (drift class)
- **verify-later:** classifier and strategist prompt enumerations

<!-- SOURCE: U16_docs019_design_plans.md -->
### Classifier consolidation + site_type taxonomy alignment (queued work)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_builder_thread Q5/Q8 — a queued brief with a read-first plan, not executed in these files.
- **what:** Two classifiers overlap: domain-research-classifier (newer, relay-native) and site-classifier (pageflow/intake path, which does NOT use work-item triage — its differences may be load-bearing: intake's hitl_confirm_type keys off confirmed_type.recommended_builder). Brief: diff both rows, map dependency points, check hard before changing; merge additions both ways with snapshot migrations; deprecate only at zero usage. Behind it: the classifier and strategist use different canonical site_type vocabularies in the same spec chain — one decision, two snapshot prompt migrations.
- **sources:** HANDOFF_builder_thread.md#3; README_flows.md (Q8 note)
- **relations:** adoption-writes-first; vertical-exemplar hop
- **verify-later:** both classifier rows; intake usage query on orchestration_states

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Classifier as strategic brain (always runs full)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "The classifier … runs on every site entering the pipeline, and it always does its full job … Adoption does not shortcut it"; Phase 1 current, Phases 2–5 not implemented
- **what:** The `domain-research-classifier` decides what a site *should be* on every site; adoption/operator-mission are weighted inputs, not bypasses. It is not constrained to current capability — best-version items it can't build yet are marked `blocked` for `feasibility-recheck`. Silent override is the failure mode.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-classifier-is-the-strategic-brain-it-always-runs-in-full, WM/028_platform_mission_and_pipeline_direction.md#input-sources-and-their-weight, WM/028_platform_mission_and_pipeline_direction.md#phased-implementation
- **relations:** website mission; fidelity dial; spec-has-status; adoption pipeline
- **verify-later:** domain-research-classifier agent_definition; migration 006

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Fidelity dial (locked/high/medium/low)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "Fidelity … Five values, with high as the default when adoption evidence is present"; "depends on per-item status on specs … Phase 3"
- **what:** A dial controlling how much aspirational extension reaches the first build vs how faithfully it matches the strongest evidence (usually adoption): `locked` (exact, no promotion), `high`, `medium`, `low`; no-adoption reinterprets it as a confidence tolerance. Currently only implicit `high` (Phase 1).
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#fidelity-controlling-how-much-aspiration-reaches-the-first-build
- **relations:** spec-has-status; classifier strategic brain; adoption faithfulness locks
- **verify-later:** proposed adoption_meta/build_policy spec aspect

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Spec has per-item status — one spec, not two
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "one spec, not two. Items have status (deployed / planned / blocked) … It is not fully implemented yet … planned to be implemented in Phase 2"
- **what:** The dream is the full spec; the build is its non-blocked subset. Per-item status makes the dream-vs-build distinction mechanical — the build pipeline builds only `deployed`, `feasibility-recheck` promotes `blocked→planned`. Each spec row records source/source_agent/source_item_id for provenance; agents read-and-extend, never silently overwrite.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-spec-has-status-deployed-planned-blocked, WM/028_platform_mission_and_pipeline_direction.md#who-writes-what-who-doesnt-override
- **relations:** references doc 021; fidelity dial; feasibility-recheck
- **verify-later:** site_specs is_current/superseded_at; feasibility-recheck task

<!-- SOURCE: U18_sql_for_agents.md -->
### site-classifier → research-backed classification with domain_profile
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** 003 header: "Changes site-classifier from a single Haiku LLM guess into a research-backed orchestrator"; later 049 creates domain-research-classifier for the work-item pipeline, which takes over first-stage classification.
- **what:** Evolution of classification: v1 was one Haiku call mapping domain+objective to {landing, content, portfolio, brochure} + recommended_builder. v2 (file 003) made it an orchestrator: Haiku research brief → research-agent web investigation → Sonnet synthesis producing backward-compatible site_type plus a rich domain_profile (business identity, tone, visual_direction, image_guidance, strategic analysis). Explicit responsibility fences: does NOT pick pages or style_collection (planner's job) but DOES provide design inputs consumed by planner, image-generator, webdesign-agent, page-content-writer.
- **sources:** 003_site_classifier.sql; sql_for_agents_v1/003_site_classifier.sql; sql_for_agents_v2/000_backup_agents.sql
- **relations:** succeeded by domain-research-classifier (work-item pipeline); domain_profile is ancestor of site_specs aspects
- **verify-later:** live site-classifier definition vs domain-research-classifier; who consumes domain_profile today

<!-- SOURCE: U18_sql_for_agents.md -->
### domain-research-classifier (work-item first stage)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 049 header documents pipeline position "first agent after seed_build_queue"; 067 adds extended-thinking budget to its classify_and_extract step (conditional on patch deploy).
- **what:** Handler for needs_domain_research: researches a domain via web search and scrape, classifies site type, extracts identity signals, writes site_specs aspects "identity" and "classification", creates the next work item (needs_briefing; later needs_strategy per 060).
- **sources:** 049_domain_research_classifier.sql; 060_domain_strategist.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** successor of site-classifier v2; site_specs aspect model
- **verify-later:** current next-item wiring (strategy vs briefing)

<!-- SOURCE: U18_sql_for_agents.md -->
### spec-updater (mechanical site_specs merge from findings)
- **category:** site-spec-and-classifier
- **status-signal:** unknown
- **status-evidence:** 072 definition; no later patches in this unit.
- **what:** Handler for needs_spec_update items: applies {aspect, field, suggested_value} to site_specs with the WriteSiteSpecAction versioning pattern. No LLM. Description-only items complete as "needs human review". "The complexity is in the Go action, not in the workflow."
- **sources:** 072_spec_updater_agent.sql
- **relations:** content-gap-planner and audits emit its items; site_specs supersede-versioning
- **verify-later:** update_site_spec_from_item action

<!-- SOURCE: U19_sql_tables_components.md -->
### site_specs aspect-versioned specification store
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Table created in Phase-0 migration (019); live pg_dump backup (bk_site_specs.sql) shows the production shape including pinned; extensive backfills for real sites.
- **what:** All strategic site specification lives as (site_id, aspect, data JSONB) rows — identity, strategy, tone, design_intent, content_direction, growth_config, adoption_source — with provenance (source enum: classifier/adoption/hitl/planner/improvement/seed/manual/rollback/fork/recovery; source_agent; source_item_id) and history via is_current + superseded_at (unique current per site+aspect). write_site_spec deep-merges partials so each row is self-contained. `pinned` (Phase 4) prevents agents overriding human-set specs.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql; docs/agent_docs/sql_for_tables/bk_site_specs.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a-team-data
- **relations:** site_plans (operational counterpart); site snapshots capture current specs; identity enrichment (departments/team).
- **verify-later:** write_site_spec action; pinned enforcement in writers.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site classifier and site_type taxonomy
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** site-classifier agent SQL with the landing/content/portfolio/brochure taxonomy and recommended_group mapping, plus a template-flattening prompt fix (evidence it was actually run); taxonomy names the live classifier architecture (docs 021).
- **what:** A lightweight LLM agent classifying a project into site types — landing (conversion single-CTA), content (publishing/ads/SEO), portfolio (showcase), brochure/directory (multi-page business / listings) — with confidence, reasoning, detected signals, and a recommended builder group. The direct ancestor of the platform's archetype/classification system.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** successor: site-spec-and-classifier (classification architecture, archetype); briefing agent; intake orchestrator.
- **verify-later:** site-classifier agent row; current archetype enum vs this 4-type taxonomy.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Specialist architects per site type
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 023 SQL: landing-page-architect created (renamed copy of site-component-architect), content-site-architect created with content-site components (article grid, sidebar, ad zones, category nav), portfolio-architect created "for future use".
- **what:** One architect agent per site type, each with its own default sections and component_category filter into the library; the alternative "one architect, differentiated by build plan" was debated (025) and the group-per-project-type model won conceptually.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** groups as project recipes; conditional_call_agent routing.
- **verify-later:** the three architect rows; content-site component rows.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Site classifier agent
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/003 SQL defines 'site-classifier' with a Haiku prompt classifying landing/content/portfolio/brochure; docs015/004 confirms "Single LLM call → outputs ONE site_type... ONE recommended_builder"; current system uses the multi-aspect site-spec classifier (docs024 021).
- **what:** A lightweight LLM agent that classifies a domain+objective into a site type (landing/content/portfolio/brochure) with confidence, reasoning, detected industry and signals, and recommends the corresponding builder group. Its single-label output was later superseded by the richer site-spec aspect classification.
- **sources:** docs006_workflow_builder/003_current_state_of_agents.sql#2-SITE-CLASSIFIER; docs007_brochure_builder/001_brochure_builder_plan.md#Classification-Signals; docs015_data_flow_verification/004_builder_flow.md
- **relations:** intake orchestrator; HITL type confirmation; successor: site-spec-and-classifier architecture.
- **verify-later:** agent_definitions 'site-classifier' vs current classifier agents.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Intake orchestrator workflow (classify → brief → spawn builder)
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/011: "✅ WORKING SYSTEM: HITL Orchestration with Multi-Agent Workflow" listing the 11-step intake workflow with two HITL pauses; docs017/023 later introduces intake-orchestrator-v2 routing.
- **what:** The entry-point orchestration: spawn/call site-classifier → fetch_available_builders from DB → HITL confirm site type (human can override classifier and builder choice) → fetch builder questionnaire → briefing agent fills it → HITL review brief → spawn and call the chosen builder. Established the pattern of human quality-gates before expensive generation.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#Working-Agents; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Naming-and-Coexistence
- **relations:** site classifier; briefing agent; intake-orchestrator-v2; HITL protocol.
- **verify-later:** intake-orchestrator workflow JSON in agent_definitions; whether v2 routing exists.

<!-- SOURCE: U22_recent_small_docs.md -->
### Unified site spec (status-tagged single document)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** Doc lives under `docs022_domain_authority/old/`; "Extend the current classifier output ... Steps 1-3 can happen incrementally" — proposed, and archived.
- **what:** A proposal for the site-classifier to emit one unified spec covering classification, identity, design_intent, content_direction, pages, features, SEO, and maintenance_profile — every item tagged `status` (deployed/planned/blocked). The "dream" is the whole doc; the "build" is the non-blocked subset. Downstream agents (planner enriches rather than decides pages; design/content agents implement explicit intent; audit agents treat the spec as ground truth; HITL edits it).
- **sources:** docs022.../old/004_classifier_notes.md
- **relations:** site-classifier vertical/disposition output, feasibility/blocked-handler pattern, design_intent, HITL
- **verify-later:** site_specs.spec_type='unified_spec'; classifier identity/design_intent/content_direction fields

<!-- SOURCE: U22_recent_small_docs.md -->
### Feasibility / blocked-handler pattern
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "the claim action catches it ... Item marked 'blocked', error='Handler agent not registered' ... weeks later ... Feasibility-recheck task finds it ... promoted to 'triaged'." Describes an existing dispatch/claim mechanism.
- **what:** A pattern where planners/discovery emit work items naming an intended handler even if that agent doesn't exist yet. The claim action checks agent_definitions; unknown handlers mark the item `blocked` with an error; a periodic feasibility-recheck task promotes blocked items to `triaged` once the handler is deployed. A Go `check_feasibility` action can also pre-block at spec time based on the agent registry.
- **sources:** docs022.../old/004_classifier_notes.md#feasibility-assessment
- **relations:** unified site spec, work-item lifecycle, tool-lifecycle
- **verify-later:** claim action handler-existence check; feasibility-recheck scheduled task

<!-- SOURCE: U23_docs_root_vonc.md -->
### write_site_spec spec_data string coercion
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session: "FIXED, deployed"; migration table: code deployed 2026-06-24 ~15:00.
- **what:** `WriteSiteSpecAction` hard type-asserted `spec_data` to a map, rejecting the plain-string `mission_brief`/`roadmap_brief` the domain-submitter workflow resolves ("spec_data must be a JSON object, got string"). Fix: a coercion block — JSON string → parse; plain string → wrap as `{"text": value}` (matching the classifier prompt's `{{.site_specs.specs.mission_brief.text}}` read); objects pass through. The HANDOFF doc for this bug is also a worked example of the evidence-only handoff pattern (symptom carried, cause left to be read from code).
- **sources:** docs/RUNBOOK_vonc_session(1).md#1; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc(36).md#1
- **relations:** handoff document convention; data-shape/contract-drift debugging family
- **verify-later:** platform/orchestration/actions/site_spec_actions.go WriteSiteSpecAction coercion block

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-native idea engine (Phase D / Layer 4)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "the chassis version is one idea-orchestrator agent + one workflow reusing these [existing actions], NOT a port of engine.go. Did NOT write the SQL — needs a schema pass first (check-schema-before-SQL)."
- **what:** A mapped-but-unbuilt plan to express the idea-generation method as chassis actions rather than the standalone Go/Python engine: `execute_llm_prompt` for generate/cut/verify/score, `web_search`/`scrape_web`/`firecrawl_*` for verify, and — notably — `request_human_input`/`create_approval_request`/`await_approval`/`process_approval_decision` for the operator confirm+review gate, explicitly identified as "literally HITL." Distinguishes two shapes for applying the method to a domain: Shape A (the site IS the service, like idea.uk) vs Shape B (a static "request a report" page posting to one central service) — because the engine is server-side and minutes-long, it cannot be a forked `content_components` client-side tool the way other tools are.
- **sources:** `running_notes(44).md` ("Wrote the architecture & deployment guide; clarified hosting + OpenAI")
- **relations:** idea generation method; HITL (docs002_hitl_parallel); tool-lifecycle (contrast with deploy_tool_to_site)
- **verify-later:** whether an `idea-orchestrator` agent_definition or workflow exists

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Email identity in site_spec — deterministic address encoding + per-site `email` aspect
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md... Proposed `email` data... so a FUTURE email-provisioner agent (design only; catch-all makes it unnecessary now) can create per-domain forwarders."
- **what:** A proposed platform-wide convention for how any generated site gets an inbound/outbound email identity: a deterministic encoding (lowercase domain, `.`→`-`, `@<operator-domain>`, e.g. `agritec.uk` → `agritec-uk@leopardess.uk`), stored (not derived-on-read) to allow per-site overrides and to handle rare collisions; a new `email` aspect on `site_specs` (joining the existing classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance aspects) carrying status/address/from/reply_to/provider/forwards_to, reusing the spec's existing deployed/planned/blocked + feasibility-recheck state machine.
- **sources:** `running_notes(44).md` ("FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md")
- **relations:** site-spec-and-classifier (021 aspect list); catch-all email routing (superseded sub-concept, below)
- **verify-later:** whether `email` was actually added to the 021 aspect list; `EMAIL_identity_in_site_spec.md` (live doc)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Catch-all email forwarding — abandoned in favour of specific per-site forwarders
- **category:** site-spec-and-classifier
- **status-signal:** abandoned
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-06-06 — inbound test FAILED (No Such User Here): catch-all not catching"; and again "inbound still bouncing (No Such User); root cause = Default Address not forwarding" — two consecutive real-world failures of the originally-planned mechanism.
- **what:** The initial plan used a domain-level catch-all (cPanel "Default Address" / "Forward All Email for a Domain") so any `<encoded>@leopardess.uk` address would work without per-site setup. In practice this repeatedly bounced with "No Such User Here" because the mail backend delivers known mailboxes locally and only routes truly-unmatched addresses through the default address, which itself was misconfigured/pointed at the wrong of two confusingly similar domains (`leopardess.uk` vs `leopardess.co.uk`). Design refinement recorded explicitly: "prefer SPECIFIC per-site forwarders (created when a site is published) over a server catch-all — only forward addresses that exist, no backscatter, and it's exactly what the future email-provisioner agent does."
- **sources:** `running_notes(44).md` (two consecutive checkpoints, 2026-06-06)
- **relations:** Email identity in site_spec (the design this discovery feeds back into)
- **verify-later:** current leopardess.uk cPanel Default Address / Forwarders configuration

<!-- SOURCE: U25_leopardess_social.md -->
### Mission + roadmap as site_specs aspects (strategy-driven site intake)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 004_submit_vonc_trigger.sh ("Tier 3 submission: domain + mission + roadmap + briefs") exists and vonc.com was built from it; 003d specifies persist_mission/persist_roadmap via the existing write_site_spec action.
- **what:** Strategic context travels as input_data.mission (positioning, differentiators, tone, target users, core concepts, measurable objectives) and input_data.roadmap (phases with per-page purpose, section_types and content_context), persisted to site_specs aspects 'mission' and 'roadmap'. The classifier is told not to discover business type from the domain for mission-driven sites (site_type "interactive-platform"); the planner builds only the current phase and outputs section_types, not component names; content writers draw voice from mission and per-page content_context. Explicitly requires no new tables, no chassis code, no RAG for v1.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Approach, #What-goes-where, #Pipeline-changes; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh
- **relations:** component selector/creator; phase advancement loop; vonc.com v1 site
- **verify-later:** site_specs aspects mission/roadmap for vonc; intake-orchestrator/domain-submitter workflow steps

<!-- SOURCE: U25_leopardess_social.md -->
### Roadmap phase advancement and automated strategic review
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 003d "Phase advancement (later) … Manual for now"; 003 (earlier version) sketched the full automated loop ("scheduled agent … compare actuals vs targets … propose phase advancement") which later versions dropped to a one-liner.
- **what:** Phases advance by updating the roadmap aspect (current phase → complete, next → active) and re-triggering planning; measurable objectives in the mission aspect (DAU, completion rates, session duration, share rate) tell you when. The fuller vision — a scheduled strategic-review agent closing strategy → build → measure → adjust — was designed in 003 v1 and deferred; the delta is the record that it was consciously parked, not forgotten.
- **sources:** docs/social001_vonc_tiktok_social/003_spark_strategic_planning_architecture.md#Future-automated-strategic-review (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Phase-advancement
- **relations:** mission/roadmap aspects; traffic-analytics (the missing measurement half)
- **verify-later:** any analytics source; scheduler entries for strategic review (expect none)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dream spec / gap analysis / feasibility (one spec with status, not two documents)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 021 resolved decision 24 "One spec, not two"; 028: per-item status "not fully implemented yet — Phase 2"
- **what:** The full spec IS the dream; items carry status deployed/planned/blocked; gap analysis = blocked/planned subset; feasibility-recheck promotes blocked→planned when capability arrives; feasibility annotations prevent impossible work items. Older 002d dream_spec-in-content_data shape superseded by this. Phase 2 of 028 makes it mechanical.
- **sources:** 021#One Spec, Not Two; 028#The spec has status; 002d#Gap Analysis
- **relations:** fidelity dial; feasibility-recheck task
- **verify-later:** does feasibility-recheck scheduled task exist; per-item status columns

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Site spec unification (site_specs aspects as the one authoritative spec)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 021 target-state doc with immediate/short/medium ordering; backfill + fallback both recommended; newer pipeline already uses aspects
- **what:** One versioned spec per site as independent aspect rows (classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance), is_current/superseded_at per-aspect versioning with source/source_agent/source_item_id provenance; write_site_spec deep-merges so every row is a complete self-contained record (pruning-safe). content_data is legacy; read_site_spec falls back to it. Classifier writes intent, planner implements, design agent executes, audits enforce. The 15 content-strategy questions map onto aspects; deep research is a future classifier enrichment.
- **sources:** 021 full; P1#Site Specification System
- **relations:** 028 ownership; 030 strategic-vs-plan-time; dream spec
- **verify-later:** backfill done for legacy sites; read_site_spec fallback

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fidelity dial (locked/high/medium/low + no-adoption confidence mode)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 028 Phases: Phase 1 current (implicit high); explicit fidelity input is Phase 3; depends on per-item spec status (Phase 2)
- **what:** Controls how much aspiration reaches first deployment vs is deferred to the improvement loop, and the loop's promotion rate. locked = adopted-exact, no promotions; high (default with adoption) = faithful launch, ~one substantive change per audit cycle; medium = modest extensions; low = adoption as inspiration; blank domains reinterpret it as research-confidence tolerance. Lives on the trigger input + a build_policy/adoption_meta aspect.
- **sources:** 028#Fidelity, #Phased implementation
- **relations:** spec status; improvement loop rate
- **verify-later:** any fidelity/build_policy aspect in prod

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Spec aspect ownership and read-and-extend (anti-silent-overwrite)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028 ownership list; adoption-aware classifier prompt (migration 006) live; open question notes planner still writing design_intent pre-030
- **what:** Classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planners own site_plan. Classifier over adoption output is read-and-extend (preserve adopted dimensions, add strategy), never overwrite. Named failure modes: silent overwrite, confident fabrication on thin signal, default-to-brochure, reflexive upstream re-runs, schema-level commercial bias, adoption without strategic analysis.
- **sources:** 028#Who writes what, #Failure modes
- **relations:** 030 planner redirect to directives; composition self-heal
- **verify-later:** build-site-planner no longer writes design_intent/content_direction post-030

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Superseding a spec doesn't undo installed artefacts (re-queue rule)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 028: gamesdesign remediation hit it (fresh specs + stale installed theme at sites.style_collection_id)
- **what:** Agents with install side-effects (composition, nav, pages, assets) write beyond site_specs and leave live pointers from long-lived tables; invalidating their spec must also queue the re-run work item (needs_composition etc.) — long-term the supersession itself should emit the recovery item. Test: does the agent write other tables AND does a live pointer reference the write?
- **sources:** 028#Failure modes (last)
- **relations:** install_site_composition; composition trigger matrix (027)
- **verify-later:** whether supersession-emits-item was ever built

<!-- SOURCE: U04_idea_uk.md -->
### Structured design_intent from the classifier (palette + typography reference_values)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Migration applied 2026-06-20; "Palette migration proven on a real build" 2026-06-21 — fresh idea.uk resolved `palette_source=design_intent_values`, parchment, "no invented blue".
- **what:** Root cause of generic-looking fresh builds: the classifier wrote design_intent colours as **prose** (hex buried in colour_mood sentences) while every consumer — the composition cascade and the analyze_design prompt — reads **structured** `design_intent.palette.reference_values` (8 slots) + `typography.reference_values`. The migration edits the classifier's classify_and_extract schema and adds a MANDATORY-fields bullet (all 8 slots as hex; style_direction must agree with the palette; never default to blue-and-grey), applied via snapshot_agent backup + exact-anchor replace() with a RAISE self-check. This single change is what makes both design stages agree (base = parchment, overlay starts from parchment).
- **sources:** idea.uk/migration_domain_research_classifier_structured_design_intent.sql; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Stage A + checklist); idea.uk/DESIGN_PIPELINE_two_track_investigation(5).md ("Direct consequence" section)
- **relations:** two-stage pipeline; mandatory-full overlay bug (this is precondition for its fix); prompt-migration discipline.
- **verify-later:** live classifier prompt contains the palette block; site 1244516d design_intent spec.

<!-- SOURCE: U04_idea_uk.md -->
### Phase 0 classifier-only positioning read
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** "Phase 0 result (2026-06-14 — ran, with the live site up)": faithful identity/classification/content_direction/design_intent specs, interactive-platform at ~0.91 confidence.
- **what:** Running just domain-research-classifier on a domain as a near-zero-cost positioning brief before committing to a build — its four spec aspects ARE the answer to "what does this site do for a stranger?". Caveats codified: a fresh read is NOT blank-slate (the classifier scrapes the live site up to 3 pages unless an adoption already ran); a generic name yields a generic name-only read, so hiding the live site removes signal not bias; a safe suppression trick exists (temporary blank nginx `location = /` — never touch DNS/nginx wholesale with a live Stripe webhook); the classifier's terminal needs_strategy item will flow into a full build if dispatch is running. Decision: leave the live site up.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (Phase 0); idea.uk/HANDOFF_chassis_site.md
- **relations:** standing-ambition mission (the read was faithful-but-backward-looking, which motivated it); fresh vs adoption.
- **verify-later:** idea.uk site 97ed2f64 spec rows (incl. the duplicate-spec tidy-up).

<!-- SOURCE: U04_idea_uk.md -->
### Build-standard classifier migration (best-in-class quality/fit, not scope)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "READY, PROVEN-CORRECT, NOT YET APPLIED… replace() simulated against the live prompt → lands cleanly" (2026-06-21); still unapplied in TODO P2 (2026-06-26).
- **what:** A migration prepending a "Build standard" block to the classifier prompt: every build should aim at best-in-class **quality and fit** for its vertical — explicitly not scope inflation. Its first draft failed on a multi-line anchor (whitespace mangling) and was fixed to a single-line anchor with a rollback proven clean — feeding the prompt-migration discipline. Test plan: fresh build first; confirm an adopted rebuild stays faithful rather than drifting.
- **sources:** idea.uk/HANDOFF(13).md (Migrations status); idea.uk/TODO_chassis_and_idea_uk(1).md#P2; idea.uk/running_notes(63).md (lll/mmm 06-20/21)
- **relations:** standing-ambition default; prompt-migration discipline.
- **verify-later:** whether migration_classifier_build_standard.sql was later applied (file itself lives outside this unit).

<!-- SOURCE: U05_content_quality_linking.md -->
### Guide as first-class page_type (classifier vocabulary + canonical URLs)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14j: "guides typed page_type=guide directly … migration_adoption_add_guide_page_type.sql worked".
- **what:** Guides were folded into blog-post by the site-adoption-agent's analyze_site enum, defeating query.pages_where_type:guide list resolution. Structural fix over band-aid: `guide` added to the adoption classifier enum + guidance, existing rows re-typed, URLs migrated /blog/guide-<slug>.html → /guides/<slug>/index.html (page_canonical.go already had the guide case). Classification geography documented: analyze_site LLM emits per-page page_type+url; site-classifier is site-type only; build-site-planner has a staler vocabulary but preserves existing pages verbatim; pages.page_type has a kebab-format check, no value allowlist.
- **sources:** running_notes_14(26).md#part-13-14c
- **relations:** Tier-D lists; adoption pipeline; bare-duplicate cleanup.
- **verify-later:** site-adoption-agent analyze_site prompt enum; pages.page_type values on a fresh adoption.

<!-- SOURCE: U14_docs019_runbooks.md -->
### site_type taxonomy drift between classifier and strategist (Q8)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B2 "vocabulary drift between hops — the classifier's site_type set (brochure|landing|portfolio|content|ecommerce|tools|interactive-platform|social) vs the strategist's canonical set (brochure|authority-portal|local-directory|review-site|content-hub|landing-page|portfolio). Two taxonomies for the same concept, one spec chain." Queued item 3.
- **what:** Two adjacent relay hops use different canonical vocabularies for the same site_type concept flowing through one spec chain — a contract-drift hygiene defect awaiting a one-canonical-set decision plus two snapshot migrations.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2 (Q8); docs019/RUNBOOK_builder_route(21).md#queue (item 3)
- **relations:** two front doors Q5; workflow result contract (drift class)
- **verify-later:** classifier and strategist prompt enumerations

<!-- SOURCE: U16_docs019_design_plans.md -->
### Classifier consolidation + site_type taxonomy alignment (queued work)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** HANDOFF_builder_thread Q5/Q8 — a queued brief with a read-first plan, not executed in these files.
- **what:** Two classifiers overlap: domain-research-classifier (newer, relay-native) and site-classifier (pageflow/intake path, which does NOT use work-item triage — its differences may be load-bearing: intake's hitl_confirm_type keys off confirmed_type.recommended_builder). Brief: diff both rows, map dependency points, check hard before changing; merge additions both ways with snapshot migrations; deprecate only at zero usage. Behind it: the classifier and strategist use different canonical site_type vocabularies in the same spec chain — one decision, two snapshot prompt migrations.
- **sources:** HANDOFF_builder_thread.md#3; README_flows.md (Q8 note)
- **relations:** adoption-writes-first; vertical-exemplar hop
- **verify-later:** both classifier rows; intake usage query on orchestration_states

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Classifier as strategic brain (always runs full)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "The classifier … runs on every site entering the pipeline, and it always does its full job … Adoption does not shortcut it"; Phase 1 current, Phases 2–5 not implemented
- **what:** The `domain-research-classifier` decides what a site *should be* on every site; adoption/operator-mission are weighted inputs, not bypasses. It is not constrained to current capability — best-version items it can't build yet are marked `blocked` for `feasibility-recheck`. Silent override is the failure mode.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-classifier-is-the-strategic-brain-it-always-runs-in-full, WM/028_platform_mission_and_pipeline_direction.md#input-sources-and-their-weight, WM/028_platform_mission_and_pipeline_direction.md#phased-implementation
- **relations:** website mission; fidelity dial; spec-has-status; adoption pipeline
- **verify-later:** domain-research-classifier agent_definition; migration 006

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Fidelity dial (locked/high/medium/low)
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "Fidelity … Five values, with high as the default when adoption evidence is present"; "depends on per-item status on specs … Phase 3"
- **what:** A dial controlling how much aspirational extension reaches the first build vs how faithfully it matches the strongest evidence (usually adoption): `locked` (exact, no promotion), `high`, `medium`, `low`; no-adoption reinterprets it as a confidence tolerance. Currently only implicit `high` (Phase 1).
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#fidelity-controlling-how-much-aspiration-reaches-the-first-build
- **relations:** spec-has-status; classifier strategic brain; adoption faithfulness locks
- **verify-later:** proposed adoption_meta/build_policy spec aspect

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Spec has per-item status — one spec, not two
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 028 "one spec, not two. Items have status (deployed / planned / blocked) … It is not fully implemented yet … planned to be implemented in Phase 2"
- **what:** The dream is the full spec; the build is its non-blocked subset. Per-item status makes the dream-vs-build distinction mechanical — the build pipeline builds only `deployed`, `feasibility-recheck` promotes `blocked→planned`. Each spec row records source/source_agent/source_item_id for provenance; agents read-and-extend, never silently overwrite.
- **sources:** WM/028_platform_mission_and_pipeline_direction.md#the-spec-has-status-deployed-planned-blocked, WM/028_platform_mission_and_pipeline_direction.md#who-writes-what-who-doesnt-override
- **relations:** references doc 021; fidelity dial; feasibility-recheck
- **verify-later:** site_specs is_current/superseded_at; feasibility-recheck task

<!-- SOURCE: U18_sql_for_agents.md -->
### site-classifier → research-backed classification with domain_profile
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** 003 header: "Changes site-classifier from a single Haiku LLM guess into a research-backed orchestrator"; later 049 creates domain-research-classifier for the work-item pipeline, which takes over first-stage classification.
- **what:** Evolution of classification: v1 was one Haiku call mapping domain+objective to {landing, content, portfolio, brochure} + recommended_builder. v2 (file 003) made it an orchestrator: Haiku research brief → research-agent web investigation → Sonnet synthesis producing backward-compatible site_type plus a rich domain_profile (business identity, tone, visual_direction, image_guidance, strategic analysis). Explicit responsibility fences: does NOT pick pages or style_collection (planner's job) but DOES provide design inputs consumed by planner, image-generator, webdesign-agent, page-content-writer.
- **sources:** 003_site_classifier.sql; sql_for_agents_v1/003_site_classifier.sql; sql_for_agents_v2/000_backup_agents.sql
- **relations:** succeeded by domain-research-classifier (work-item pipeline); domain_profile is ancestor of site_specs aspects
- **verify-later:** live site-classifier definition vs domain-research-classifier; who consumes domain_profile today

<!-- SOURCE: U18_sql_for_agents.md -->
### domain-research-classifier (work-item first stage)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 049 header documents pipeline position "first agent after seed_build_queue"; 067 adds extended-thinking budget to its classify_and_extract step (conditional on patch deploy).
- **what:** Handler for needs_domain_research: researches a domain via web search and scrape, classifies site type, extracts identity signals, writes site_specs aspects "identity" and "classification", creates the next work item (needs_briefing; later needs_strategy per 060).
- **sources:** 049_domain_research_classifier.sql; 060_domain_strategist.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** successor of site-classifier v2; site_specs aspect model
- **verify-later:** current next-item wiring (strategy vs briefing)

<!-- SOURCE: U18_sql_for_agents.md -->
### spec-updater (mechanical site_specs merge from findings)
- **category:** site-spec-and-classifier
- **status-signal:** unknown
- **status-evidence:** 072 definition; no later patches in this unit.
- **what:** Handler for needs_spec_update items: applies {aspect, field, suggested_value} to site_specs with the WriteSiteSpecAction versioning pattern. No LLM. Description-only items complete as "needs human review". "The complexity is in the Go action, not in the workflow."
- **sources:** 072_spec_updater_agent.sql
- **relations:** content-gap-planner and audits emit its items; site_specs supersede-versioning
- **verify-later:** update_site_spec_from_item action

<!-- SOURCE: U19_sql_tables_components.md -->
### site_specs aspect-versioned specification store
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** Table created in Phase-0 migration (019); live pg_dump backup (bk_site_specs.sql) shows the production shape including pinned; extensive backfills for real sites.
- **what:** All strategic site specification lives as (site_id, aspect, data JSONB) rows — identity, strategy, tone, design_intent, content_direction, growth_config, adoption_source — with provenance (source enum: classifier/adoption/hitl/planner/improvement/seed/manual/rollback/fork/recovery; source_agent; source_item_id) and history via is_current + superseded_at (unique current per site+aspect). write_site_spec deep-merges partials so each row is self-contained. `pinned` (Phase 4) prevents agents overriding human-set specs.
- **sources:** docs/agent_docs/sql_for_tables/019_site_specs_page_component_history.sql; docs/agent_docs/sql_for_tables/bk_site_specs.sql; docs/agent_docs/sql_for_tables/018_site_work_items.sql#075a-team-data
- **relations:** site_plans (operational counterpart); site snapshots capture current specs; identity enrichment (departments/team).
- **verify-later:** write_site_spec action; pinned enforcement in writers.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Site classifier and site_type taxonomy
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** site-classifier agent SQL with the landing/content/portfolio/brochure taxonomy and recommended_group mapping, plus a template-flattening prompt fix (evidence it was actually run); taxonomy names the live classifier architecture (docs 021).
- **what:** A lightweight LLM agent classifying a project into site types — landing (conversion single-CTA), content (publishing/ads/SEO), portfolio (showcase), brochure/directory (multi-page business / listings) — with confidence, reasoning, detected signals, and a recommended builder group. The direct ancestor of the platform's archetype/classification system.
- **sources:** docs004_website_capture_project/007different_types_of_site/029.intake_and_groups.sql; docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** successor: site-spec-and-classifier (classification architecture, archetype); briefing agent; intake orchestrator.
- **verify-later:** site-classifier agent row; current archetype enum vs this 4-type taxonomy.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Specialist architects per site type
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** 023 SQL: landing-page-architect created (renamed copy of site-component-architect), content-site-architect created with content-site components (article grid, sidebar, ad zones, category nav), portfolio-architect created "for future use".
- **what:** One architect agent per site type, each with its own default sections and component_category filter into the library; the alternative "one architect, differentiated by build plan" was debated (025) and the group-per-project-type model won conceptually.
- **sources:** docs004_website_capture_project/006semantic_themes/README.023.specialist_site_architects.md; docs004_website_capture_project/007different_types_of_site/025.agent_group_discussion
- **relations:** groups as project recipes; conditional_call_agent routing.
- **verify-later:** the three architect rows; content-site component rows.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Site classifier agent
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/003 SQL defines 'site-classifier' with a Haiku prompt classifying landing/content/portfolio/brochure; docs015/004 confirms "Single LLM call → outputs ONE site_type... ONE recommended_builder"; current system uses the multi-aspect site-spec classifier (docs024 021).
- **what:** A lightweight LLM agent that classifies a domain+objective into a site type (landing/content/portfolio/brochure) with confidence, reasoning, detected industry and signals, and recommends the corresponding builder group. Its single-label output was later superseded by the richer site-spec aspect classification.
- **sources:** docs006_workflow_builder/003_current_state_of_agents.sql#2-SITE-CLASSIFIER; docs007_brochure_builder/001_brochure_builder_plan.md#Classification-Signals; docs015_data_flow_verification/004_builder_flow.md
- **relations:** intake orchestrator; HITL type confirmation; successor: site-spec-and-classifier architecture.
- **verify-later:** agent_definitions 'site-classifier' vs current classifier agents.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Intake orchestrator workflow (classify → brief → spawn builder)
- **category:** site-spec-and-classifier
- **status-signal:** superseded
- **status-evidence:** docs006/011: "✅ WORKING SYSTEM: HITL Orchestration with Multi-Agent Workflow" listing the 11-step intake workflow with two HITL pauses; docs017/023 later introduces intake-orchestrator-v2 routing.
- **what:** The entry-point orchestration: spawn/call site-classifier → fetch_available_builders from DB → HITL confirm site type (human can override classifier and builder choice) → fetch builder questionnaire → briefing agent fills it → HITL review brief → spawn and call the chosen builder. Established the pattern of human quality-gates before expensive generation.
- **sources:** docs006_workflow_builder/011_working_landing_page_builder.md#Working-Agents; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Naming-and-Coexistence
- **relations:** site classifier; briefing agent; intake-orchestrator-v2; HITL protocol.
- **verify-later:** intake-orchestrator workflow JSON in agent_definitions; whether v2 routing exists.

<!-- SOURCE: U22_recent_small_docs.md -->
### Unified site spec (status-tagged single document)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** Doc lives under `docs022_domain_authority/old/`; "Extend the current classifier output ... Steps 1-3 can happen incrementally" — proposed, and archived.
- **what:** A proposal for the site-classifier to emit one unified spec covering classification, identity, design_intent, content_direction, pages, features, SEO, and maintenance_profile — every item tagged `status` (deployed/planned/blocked). The "dream" is the whole doc; the "build" is the non-blocked subset. Downstream agents (planner enriches rather than decides pages; design/content agents implement explicit intent; audit agents treat the spec as ground truth; HITL edits it).
- **sources:** docs022.../old/004_classifier_notes.md
- **relations:** site-classifier vertical/disposition output, feasibility/blocked-handler pattern, design_intent, HITL
- **verify-later:** site_specs.spec_type='unified_spec'; classifier identity/design_intent/content_direction fields

<!-- SOURCE: U22_recent_small_docs.md -->
### Feasibility / blocked-handler pattern
- **category:** site-spec-and-classifier
- **status-signal:** partial
- **status-evidence:** "the claim action catches it ... Item marked 'blocked', error='Handler agent not registered' ... weeks later ... Feasibility-recheck task finds it ... promoted to 'triaged'." Describes an existing dispatch/claim mechanism.
- **what:** A pattern where planners/discovery emit work items naming an intended handler even if that agent doesn't exist yet. The claim action checks agent_definitions; unknown handlers mark the item `blocked` with an error; a periodic feasibility-recheck task promotes blocked items to `triaged` once the handler is deployed. A Go `check_feasibility` action can also pre-block at spec time based on the agent registry.
- **sources:** docs022.../old/004_classifier_notes.md#feasibility-assessment
- **relations:** unified site spec, work-item lifecycle, tool-lifecycle
- **verify-later:** claim action handler-existence check; feasibility-recheck scheduled task

<!-- SOURCE: U23_docs_root_vonc.md -->
### write_site_spec spec_data string coercion
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_vonc_session: "FIXED, deployed"; migration table: code deployed 2026-06-24 ~15:00.
- **what:** `WriteSiteSpecAction` hard type-asserted `spec_data` to a map, rejecting the plain-string `mission_brief`/`roadmap_brief` the domain-submitter workflow resolves ("spec_data must be a JSON object, got string"). Fix: a coercion block — JSON string → parse; plain string → wrap as `{"text": value}` (matching the classifier prompt's `{{.site_specs.specs.mission_brief.text}}` read); objects pass through. The HANDOFF doc for this bug is also a worked example of the evidence-only handoff pattern (symptom carried, cause left to be read from code).
- **sources:** docs/RUNBOOK_vonc_session(1).md#1; docs/HANDOFF_vonc_write_site_spec_spec_data.md; docs/RUNNING_NOTES_vonc(36).md#1
- **relations:** handoff document convention; data-shape/contract-drift debugging family
- **verify-later:** platform/orchestration/actions/site_spec_actions.go WriteSiteSpecAction coercion block

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Chassis-native idea engine (Phase D / Layer 4)
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "the chassis version is one idea-orchestrator agent + one workflow reusing these [existing actions], NOT a port of engine.go. Did NOT write the SQL — needs a schema pass first (check-schema-before-SQL)."
- **what:** A mapped-but-unbuilt plan to express the idea-generation method as chassis actions rather than the standalone Go/Python engine: `execute_llm_prompt` for generate/cut/verify/score, `web_search`/`scrape_web`/`firecrawl_*` for verify, and — notably — `request_human_input`/`create_approval_request`/`await_approval`/`process_approval_decision` for the operator confirm+review gate, explicitly identified as "literally HITL." Distinguishes two shapes for applying the method to a domain: Shape A (the site IS the service, like idea.uk) vs Shape B (a static "request a report" page posting to one central service) — because the engine is server-side and minutes-long, it cannot be a forked `content_components` client-side tool the way other tools are.
- **sources:** `running_notes(44).md` ("Wrote the architecture & deployment guide; clarified hosting + OpenAI")
- **relations:** idea generation method; HITL (docs002_hitl_parallel); tool-lifecycle (contrast with deploy_tool_to_site)
- **verify-later:** whether an `idea-orchestrator` agent_definition or workflow exists

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Email identity in site_spec — deterministic address encoding + per-site `email` aspect
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** `running_notes(44).md`: "FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md... Proposed `email` data... so a FUTURE email-provisioner agent (design only; catch-all makes it unnecessary now) can create per-domain forwarders."
- **what:** A proposed platform-wide convention for how any generated site gets an inbound/outbound email identity: a deterministic encoding (lowercase domain, `.`→`-`, `@<operator-domain>`, e.g. `agritec.uk` → `agritec-uk@leopardess.uk`), stored (not derived-on-read) to allow per-site overrides and to handle rare collisions; a new `email` aspect on `site_specs` (joining the existing classification/identity/strategy/design_intent/content_direction/site_plan/seo/maintenance aspects) carrying status/address/from/reply_to/provider/forwards_to, reusing the spec's existing deployed/planned/blocked + feasibility-recheck state machine.
- **sources:** `running_notes(44).md` ("FRAMEWORK DESIGN (written this turn): EMAIL_identity_in_site_spec.md")
- **relations:** site-spec-and-classifier (021 aspect list); catch-all email routing (superseded sub-concept, below)
- **verify-later:** whether `email` was actually added to the 021 aspect list; `EMAIL_identity_in_site_spec.md` (live doc)

<!-- SOURCE: U24e_docs_archive_idea_uk.md -->
### Catch-all email forwarding — abandoned in favour of specific per-site forwarders
- **category:** site-spec-and-classifier
- **status-signal:** abandoned
- **status-evidence:** `running_notes(44).md` "CHECKPOINT 2026-06-06 — inbound test FAILED (No Such User Here): catch-all not catching"; and again "inbound still bouncing (No Such User); root cause = Default Address not forwarding" — two consecutive real-world failures of the originally-planned mechanism.
- **what:** The initial plan used a domain-level catch-all (cPanel "Default Address" / "Forward All Email for a Domain") so any `<encoded>@leopardess.uk` address would work without per-site setup. In practice this repeatedly bounced with "No Such User Here" because the mail backend delivers known mailboxes locally and only routes truly-unmatched addresses through the default address, which itself was misconfigured/pointed at the wrong of two confusingly similar domains (`leopardess.uk` vs `leopardess.co.uk`). Design refinement recorded explicitly: "prefer SPECIFIC per-site forwarders (created when a site is published) over a server catch-all — only forward addresses that exist, no backscatter, and it's exactly what the future email-provisioner agent does."
- **sources:** `running_notes(44).md` (two consecutive checkpoints, 2026-06-06)
- **relations:** Email identity in site_spec (the design this discovery feeds back into)
- **verify-later:** current leopardess.uk cPanel Default Address / Forwarders configuration

<!-- SOURCE: U25_leopardess_social.md -->
### Mission + roadmap as site_specs aspects (strategy-driven site intake)
- **category:** site-spec-and-classifier
- **status-signal:** deployed
- **status-evidence:** 004_submit_vonc_trigger.sh ("Tier 3 submission: domain + mission + roadmap + briefs") exists and vonc.com was built from it; 003d specifies persist_mission/persist_roadmap via the existing write_site_spec action.
- **what:** Strategic context travels as input_data.mission (positioning, differentiators, tone, target users, core concepts, measurable objectives) and input_data.roadmap (phases with per-page purpose, section_types and content_context), persisted to site_specs aspects 'mission' and 'roadmap'. The classifier is told not to discover business type from the domain for mission-driven sites (site_type "interactive-platform"); the planner builds only the current phase and outputs section_types, not component names; content writers draw voice from mission and per-page content_context. Explicitly requires no new tables, no chassis code, no RAG for v1.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Approach, #What-goes-where, #Pipeline-changes; docs/social001_vonc_tiktok_social/trigger_script/004_submit_vonc_trigger.sh
- **relations:** component selector/creator; phase advancement loop; vonc.com v1 site
- **verify-later:** site_specs aspects mission/roadmap for vonc; intake-orchestrator/domain-submitter workflow steps

<!-- SOURCE: U25_leopardess_social.md -->
### Roadmap phase advancement and automated strategic review
- **category:** site-spec-and-classifier
- **status-signal:** aspirational
- **status-evidence:** 003d "Phase advancement (later) … Manual for now"; 003 (earlier version) sketched the full automated loop ("scheduled agent … compare actuals vs targets … propose phase advancement") which later versions dropped to a one-liner.
- **what:** Phases advance by updating the roadmap aspect (current phase → complete, next → active) and re-triggering planning; measurable objectives in the mission aspect (DAU, completion rates, session duration, share rate) tell you when. The fuller vision — a scheduled strategic-review agent closing strategy → build → measure → adjust — was designed in 003 v1 and deferred; the delta is the record that it was consciously parked, not forgotten.
- **sources:** docs/social001_vonc_tiktok_social/003_spark_strategic_planning_architecture.md#Future-automated-strategic-review (family-delta); docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#Phase-advancement
- **relations:** mission/roadmap aspects; traffic-analytics (the missing measurement half)
- **verify-later:** any analytics source; scheduler entries for strategic review (expect none)
