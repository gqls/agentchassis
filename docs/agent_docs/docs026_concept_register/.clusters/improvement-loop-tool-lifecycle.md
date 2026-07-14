# Cluster: improvement-loop-tool-lifecycle
Categories included: improvement-loop, tool-lifecycle, tool-library, tool-pipeline, new:component-lifecycle, new:games-lifecycle


<!-- SOURCE: U01_docs024_numbered_core.md -->
### QA three layers, group auditor agents, and the promotion pattern
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 002d architecture with live agents (design-audit-agent, visual-design-auditor, content-quality-auditor, site-review-agent)
- **what:** Layer 1 structural checks (algorithmic, every cycle), Layer 2 group LLM audits (shared context, ONE LLM call per group), Layer 3 strategic review. Group agents chosen over per-check agents (context reuse) and over a mini-action registry (every agent is an orchestrator); a check is promoted from action step to spawned agent by changing one workflow line, output_field unchanged. Site type enables audit groups via maintenance_profile.audit.
- **sources:** 002(4)/002d#Quality Assurance; #Promotion Pattern
- **relations:** improvement loop; audit-enforces-not-overrides
- **verify-later:** design-audit-agent workflow; maintenance_profile.audit config

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Audit enforces intent, doesn't override (chain of authority) + propose mode for spec-less sites
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 002d responsibility boundaries; resolved decision 20
- **what:** classifier decides intent → planner implements → composition installs → webdesign renders → audit checks build vs stated intent and emits work items; it never makes design decisions. Exception: where no classifier output exists, audit may PROPOSE a direction, flagged for HITL. Handlers never know their trigger (build/audit/manual all use the same webdesign-agent).
- **sources:** 002d#Responsibility Boundaries; 004#Human Direction Integration
- **relations:** dream spec; direction spec; 028 silent-override failure mode
- **verify-later:** auditor prompts reference design_intent/direction

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Improvement loop flow with pass cap, finding cap, and auto-reset ("sites breathe")
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 004 v3/v4 dated changes (2026-03-25/31); token math before/after (~88K→~30K)
- **what:** improvement-loop orchestrator: pass-limit gate (≥3 → complete_clean) → algorithmic discovery agents (quality/design/completeness) → LLM audits (TOP 5 findings each with current_value/acceptance_test/suggestion/max_fix_attempts) → triage → increment pass → insert needs_rerender p99 → dispatch. Auto-reset after 60 days / direction change / major rebuild / manual, pairing with lock expiry (the unimplemented half) to create the breathe rhythm.
- **sources:** 004 full; 031_LOCKS (the missing half)
- **relations:** section locking; direction spec; locks-should-expire question
- **verify-later:** improvement-loop workflow; audit_pass_count fields

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Discovery checks catalogue (quality/design/completeness) and ordering rule
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 004 tables with handlers per check; validate_component_standards sub-checks
- **what:** quality: broken_nav_links, placeholder_contact, generic_theme. design: hardcoded_section_colors, forced_text_colors, undeployed_assets, missing_css, validate_component_standards (9 sub-checks incl. unlinked components, slot mismatch, nav layout, unwanted search icon). completeness: empty_sections, empty_blog, orphan_pages. validate_component_standards runs BEFORE colour checks (structure before rendered-HTML checks). DiscoveryCheck interface: checks append WorkItemSpecs; the runner inserts with dedup — plugins must not insert their own rows.
- **sources:** 004#Discovery Agents, #Component Standards Validation; 016 §0 item 27 (interface shape)
- **relations:** two-strike dedup; handler routing
- **verify-later:** discovery_checks/ registry

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Blog listing rebuild and slot-detection strategy
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 103 handoff: rewritten + deployed with three bug fixes; 004 shows it in rerender-pages workflow
- **what:** rebuild_blog_listing runs in rerender-pages before get_pages: finds the actual listing slot via priority list → pages.sections → default; loads a template that genuinely has a {{range}} (guard against CSS-only components); ensures article links (SQL template patch + post-render safety net); writes content_data alongside rendered_html; computes read time from component lengths.
- **sources:** 103_blog_nav_handoff-2026-04-12.md; 004#Blog Listing Rebuild
- **relations:** empty_blog check → blog-content-planner
- **verify-later:** rebuild_blog_listing_action.go current form

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dispatch failures triage report and the bug/recommendation/gap classification gap (P10)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** 105 v4 (2026-04-15): P1–P6 deployed/done, P7/P8 human-gated, P9 fix written, P10 deferred with design note
- **what:** Systematic queue triage produced ten priority fixes: component_id nil in load_tool; retry-safe fork deploys; rate-limit errors classified transient; load_page_record page_id fallback; plan-then-reconcile for section data requests (auto-close stale); design chain unblocked; audit gap findings rerouted to needs_content_page. P10 names an architectural gap: auditors emit opinions (recommendations) that the pipeline auto-fixes as if bugs — proposed three-way classification (bug/recommendation/gap) with specialist agents + per-site approval mode (~1 week, deferred).
- **sources:** 105 full
- **relations:** write_audit_findings; approval model (P1 plan)
- **verify-later:** P9/P10 status since April

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery agents on dead/stub sites (noise at scale)
- **category:** improvement-loop
- **status-signal:** unknown
- **status-evidence:** "80+ needs_content_planning items for gamesdesign.co.uk, mostly [stale: triaged 48h+] … running on a sites row whose adoption had failed or been deleted" (2026-04-23, item 19, added in version (1) only)
- **what:** Discovery agents keep generating remediation items for sites that are deleted, stubs, or mid-adoption. Proposed precondition: skip site_ids with status deleted/archived, no current identity spec, or adoption in flight.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md item 19
- **relations:** two-strike pile-up; library-row cleanup pattern
- **verify-later:** discovery agent site-selection queries

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Recommendation specialist architecture (bug vs gap vs recommendation)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** "P10 … Deferred until HITL queue becomes a bottleneck" (April 2026); P9 (gap → needs_content_page) deployed
- **what:** LLM auditors mix factual bugs with opinions; the pipeline shouldn't auto-fix opinions. Proposed: finding_type classification (bug → auto-fix; gap → rebuild via needs_content_page — this part deployed as P9; recommendation → specialist agent decides apply/dismiss/escalate, e.g. identity-advisor for contact details), with per-site approval_mode (auto|review) gating.
- **sources:** HANDOFF-pipeline-triage-april-2026.md#P9, #P10; FOCUS_content_quality.md#machinery
- **relations:** content quality work order; two sources of truth for email
- **verify-later:** write_audit_findings_action.go Rule 4; finding_type field existence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### April 2026 pipeline triage fix set (P1–P5)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Triaged 57 build pipeline failures … wrote and deployed 7 code fixes (P1–P5, P9 across 12 files)" (April 2026)
- **what:** P1: component_id plumbed through create_work_item into site_work_items (unblocking tool-improver's load_tool). P2: idempotent tool fork deploy (reuse orphaned forks). P3: 429/rate-limit/billing errors classified transient in isAIUnavailable (items back to triaged without burning attempts; ~130 wasted attempts/day stopped). P4: load_page_record falls back page_name → page_id. P5: plan-then-reconcile for needs_section_data (auto-close stale requests when data arrives; create without duplicating when still missing) — "feedback loops need both directions".
- **sources:** HANDOFF-pipeline-triage-april-2026.md (whole)
- **relations:** section-data reconciler (P5's successor); two-strike rule
- **verify-later:** ai_errors.go isAIUnavailable patterns; plan_sections loadOpenSectionDataRequests/closeResolvedDataRequest

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component linking enrichment saga (component_id NULL on rebuilt pages)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Resolved on homepage. 7/7 linked" (2026-04-20) after fixes deployed 04-17/18; slot_name normalized differentiators-section → differentiators proving the data-component branch fires
- **what:** page_components.component_id was wiped on every rebuild because sections_metadata from the content writer carries only rendered_html — extractSectionsFromMetadata defaulted ComponentName to "section" and the enrichment guard skipped every row. Fixes: run enrichSectionsWithPlannedNames before enrichSectionsWithComponentIDs; prefer the HTML data-component attribute over metadata names; strip -section/-container/-wrapper/-block suffixes; log at Info with candidates_tried. Long-term structural fix (deferred): compile_page_sections should emit component_name per metadata entry. Stale pre-fix rows self-heal on next natural rebuild. Companion facts: `mode` has exactly one legal value "recreate"; build_mode is a dead parameter.
- **sources:** HANDOFF_2026-04-18_enrichment_bug_diagnosed_and_patched.md; HANDOFF_2026-04-19…md#1; HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md
- **relations:** data-component attribute contract (news-listing template fix); spec-is-primary-input contract
- **verify-later:** save_page_sections_action.go enrichment order; compile_page_sections metadata shape

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Improvement-sweep site starvation
- **category:** improvement-loop
- **status-signal:** unknown
- **status-evidence:** "Oldest updated_at site always wins; sites with frequent rebuilds dominate" — carried P3 across 04-17 → 04-20 handoffs, never picked up
- **what:** The improvement sweep's site selection starves some sites the same way find_dispatchable_site's arbitrary ordering does — scheduling fairness is an unowned concern across both loops.
- **sources:** HANDOFF_2026-04-17_triage_and_component_linking.md#known-issues; HANDOFF_2026-04-20…(2).md#5
- **relations:** dispatch chain fairness ORDER BY
- **verify-later:** improvement-sweep pre_query

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme-coherence audit guard (Q8)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** PLAN Q8 "A scheme-coherence check in the design auditor / improvement loop … Status: open (I)"; absent from the SCHEME CLOSE remaining-work map and every later position note — silently dropped after the paired-variable close.
- **what:** The proposed regression guard: an auditor/improvement-loop check flagging "section scheme does not match site scheme / unintended contrast" so the scheme→components fix cannot silently regress. Designed as fix-shape item 8, never specified or built; the eventual regression protection took a different form (contract in the creator prompt + the re-aimed fixer as mechanical enforcer), leaving no dedicated audit check.
- **sources:** PLAN_scheme_to_components(1).md#Q8 #Provisional-fix-shape; RUNBOOK_scheme_to_components(50).md#SCHEME-CLOSE (absence)
- **relations:** fixes-at-initial-render principle; fix_forced_text_colours re-aim (partial substitute).
- **verify-later:** design-auditor checks list — does any scheme-coherence rule exist?

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Improvement-loop colour/nav fixer suite (pre-re-aim state)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Notes (Sj) full read of the fixers; HANDOFF: "As aimed today they are scheme-blind and ENFORCE the component-owns contract (`fix_forced_text_colors` injects dark `--section-*` into is_dark_section components)."
- **what:** The established fixer infrastructure: color-variable-fixer agent runs `fix_hardcoded_colors` (dark background hex → `var(--color-primary)`; dark 2-stop gradients → primary/secondary; deliberately leaves `rgba(0,0,0,x)` overlays alone; fixes both template and rendered HTML) and `fix_forced_text_colors` (strips forced child text colours so elements inherit the chain, WCAG-validates ≥4.5, and — the superseded part — injected the white `--section-*` contract into is_dark_section components); nav-link-fixer runs `fix_nav_link_templates` (with the documented rule that `render_site_components force_rerender` must follow); `fix_component_template` routes on fix_type (inject_nav_flex_css, remove_element, align_slot_name, inject_responsive_css, repair_template_slots) — symptom fixes for exactly the dark fallback header's output. Running them as-was on idea.uk would have entrenched dark.
- **sources:** running_notes_scheme_to_components(55).md#Sj; HANDOFF_scheme_to_components_for_claude_code(1).md#Established
- **relations:** superseded by fix_forced_text_colours re-aim; fix_hardcoded_colors retained (its hex→primary mapping stays coherent with the paintPaletteBand class).
- **verify-later:** fix_harcoded_colours_action.go, fix_nav_link_templates_action.go, fix_component_template_action.go current behaviour.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### fix_forced_text_colours re-aim: painting classifier + declaration rewriter
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** gobatch_04(2) delivers whole-function Edit G + new classifier code, with "RESOLVED (2026-07-06): … switched to `var(--color-primary-text, …)`"; RUNBOOK 07-06 night: "Slice 3 (the re-aimed fixer): deploying per your note — confirm it built"; supervised first run still pending.
- **what:** The backstop fixer rebuilt around the new contract: a `paintClass` classifier (paintAmbient/paintPair/paintInk/paintPaletteBand) derives what a template's own CSS paints from regexes over its style blocks — never from `is_dark_section` (the parameter is kept but deliberately ignored so call sites stay unchanged); `rewriteSectionDeclarationsInHTML` converts literal `--section-*` declarations to the class-appropriate references (pair text, hero ink, on-colour family, color-mix derivatives) and deletes declarations from ambient non-painters; the proven literal-stripping machinery and the WCAG contrast gate are retained; the old contract-injector trio (`ensureSectionContractInHTML`/`injectSectionContract`/`sectionContractRe`) is deleted. `result.contractAdded` is repurposed as a rewrite counter (name kept per the no-rename rule, meaning shift noted).
- **sources:** gobatch_04_fixer_reaim(2).md; running_notes_scheme_to_components(55).md#Ud #Ue #Ug; SPEC_scheme_to_components.md#W5
- **relations:** section painting contract (the thing it enforces); supervised fixer first-run protocol; is_dark_section demotion.
- **verify-later:** fix_forced_text_colours_action.go deployed body (classifier present, injector trio gone); first live run's details JSON.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Fixes-land-at-initial-render principle (loop fixers backstop only)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Notes (Sk) "User steer … Fix at INITIAL RENDER (library + composition/renderer). The improvement-loop fixers stay AVAILABLE as a backstop but must NOT be *required* for new builds"; SPEC W5: "The fixers remain improvement-loop backstops — never required for a correct first render."
- **what:** A governing platform principle set mid-thread and carried into every artefact: correctness must come from the library, composition, and renderer at first render; improvement-loop fixers are post-hoc safety nets dispatched on audit findings, never a required step for a new build. This ruled out "make the fixers scheme-aware" as the primary mechanism and shaped where each fix landed (templates and prompts, not loop passes).
- **sources:** running_notes_scheme_to_components(55).md#Sk; SPEC_scheme_to_components.md#W5; HANDOFF_scheme_to_components_for_claude_code(1).md#Established (USER DIRECTION)
- **relations:** section painting contract; component-creator prompt re-aim; scheme-coherence guard (abandoned alternative).
- **verify-later:** whether any build workflow invokes colour fixers as a required step.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Supervised fixer first-run protocol (disposable specimen site)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK STEP D (2026-07-06 night+): "watch the re-aimed fixer act on a real site once, under supervision, before it is ever allowed near the improvement loop or a second site" — not yet run at last dated evidence.
- **what:** Rollout protocol for a re-aimed automated fixer: confirm the deployed pod carries it; capture the specimen's before-state (dartsonline, a disposable freshly-built non-target site, site_id 5fe8785b); spawn manually via the 016b harness; judge the returned per-component `details` JSON (declarations rewritten per class, literals stripped, contrast-gate skips) rather than the render; re-read components and diff; only then decide hand-needles vs fixer for the library tail, and only then allow the improvement loop near it. Includes the deferred plan for a guarded 7-table cascade delete of a messed-up disposable site (written against schema when needed, never ad-hoc). The sizing read found dartsonline a thin specimen (all components literal-free; only two literal text hexes).
- **sources:** RUNBOOK_scheme_to_components(50).md#STEP-D; stepD_and_pages_reads.sql; running_notes_scheme_to_components(55).md#Uj #Uk
- **relations:** fix_forced_text_colours re-aim; hazard/band tail; debugging harness (016b).
- **verify-later:** whether Step D ran; dartsonline component state; the cascade-delete script's existence.

<!-- SOURCE: U05_content_quality_linking.md -->
### Recommendation-specialist finding routing (bug / recommendation / gap)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** PLAN doc header: "Status: Proposed — not yet implemented"; FOCUS_content_quality(2): "identity-advisor does NOT exist. sites.approval_mode does NOT exist … PROPOSED, not built."
- **what:** Auditor findings mix factually-broken bugs with subjective recommendations; the pipeline treated both as content_rewrite, causing false-positive rewrites and HITL flooding. Proposal: auditors emit finding_type (bug|recommendation|gap); bugs → content_rewrite, gaps → rebuild (P9 shipped that part), recommendations → per-category specialist agents (identity-advisor for contact/email, tone-shift, content-strategist, component-template-fixer for cta, brand-designer) each returning apply/dismiss/escalate; sites.approval_mode ('auto'|'review') as the per-site HITL escape hatch; a learning loop feeding dismissed findings back into auditor prompts. Reality-checked this unit: the specialists and approval_mode were never built, and component-template-fixer in fact PUNTS on CTAs (cta_improvement → needs_review) — the PLAN's claim it "already handles CTA fixes" was wrong.
- **sources:** package_module/output_contexts/PLAN_design-note-recommendation-specialists.md; FOCUS_content_quality(2).md#machinery; running_notes_16(1).md#part-2
- **relations:** content-quality-auditor/visual-design-auditor; internal-link-resolver (took over the CTA-fix role); HITL.
- **verify-later:** write_audit_findings_action.go routing; whether finding_type exists in auditor prompts; sites schema for approval_mode.

<!-- SOURCE: U05_content_quality_linking.md -->
### Discovery-agent domain scoping + observe-only check enablement
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** running_notes_17(21) 2026-06-14: enumeration of the three discovery agents + "findings land status='detected' which is NOT claimable".
- **what:** Discovery checks live in domain-scoped analyst agents (quality-discovery-agent=build, design-discovery-agent=design, completeness-discovery-agent=content) via run_discovery_checks whose config `checks` array is the sole enablement switch; checks self-register via init() but run only when named. Findings insert with status='detected' (unclaimable), so a check can run observe-only while improvement-sweep (the triager) stays disabled — auditing signal with zero auto-action. Per-finding Pipeline/HandlerAgent from the returned WorkItemSpec survive insertWorkItem verbatim.
- **sources:** running_notes_17(21).md#wiring-confirmed + #§7-gate-RESOLVED; RUNBOOK_linking_phantom_fixes(7).md#7a
- **relations:** check_phantom_internal_links; improvement-sweep gating; sectionless check S1.
- **verify-later:** run_discovery_checks_action.go; the three agents' checks arrays.

<!-- SOURCE: U05_content_quality_linking.md -->
### improvement-sweep pause + gated re-enable sequencing
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** FOCUS_internal_linking(1) Operational note: "improvement-sweep scheduled_task is disabled (enabled=f, last completed 2026-05-08), intentionally paused during core build".
- **what:** The improvement-loop's triage sweep is deliberately off during core build; the detect→fix loop depends on it. Re-enable sequencing settled: first have the phantom_internal_links check enabled observe-only AND both surface handlers in place, watch one cycle clean, then re-enable — so resuming clears findings rather than accumulating them. improvement-loop confirmed as the sweep orchestrator (spawns discovery agents + build-dispatch-loop + triage_detected_items).
- **sources:** FOCUS_internal_linking(1).md#operational-note; RUNBOOK_linking_phantom_fixes(7).md#7b; running_notes_17(21).md 2026-06-14
- **relations:** observe-only enablement; internal-link-resolver (prerequisite handler).
- **verify-later:** scheduled_tasks improvement-sweep enabled flag; first post-enable triage cycle.

<!-- SOURCE: U05_content_quality_linking.md -->
### checkEmptyPageSections — dormant, half-superseded detector
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** running_notes_15(12) Part 6 RESULT: "the empty-page detector is dormant code"; Part 9: "stale, never-enabled code partially superseded by EmptySectionsCheck".
- **what:** A sub-check inside the validate_component_standards wrapper targeting "deployed/active pages with zero rendered components" — wired in code but the wrapper is enabled in no discovery agent, bundled with six unrelated sub-checks, scoped to miss `planned` pages, and routes to page-content-writer (which cannot persist/deploy). Its sibling EmptySectionsCheck got the handler fix + enablement; this one didn't. Kept as the canonical example of dormant check debt and why a dedicated check was written instead.
- **sources:** running_notes_15(12).md#part-6, #part-9
- **relations:** check_sectionless_pages (its replacement for the planned-page case); EmptySectionsCheck.
- **verify-later:** ComponentStandardsCheck.Run; any agent enabling validate_component_standards.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Improvement-loop colour fixers are scheme-blind (re-aim as backstop, not fix)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** "Running them as-is on idea.uk would ENTRENCH dark, not lighten it" (running_notes Sj); user steer Sk: fix at initial render, fixers stay backstop; re-aim listed as pending Go-batch work.
- **what:** Existing fixer machinery — color-variable-fixer (fix_hardcoded_colors: dark hex→var(--color-primary), leaves rgba overlays; fix_forced_text_colors: strips forced child text colours, WCAG-validates, but INJECTS the white --section-* contract for is_dark_section components), nav-link-fixer (+ its documented render_site_components force_rerender follow-up), fix_component_template symptom fixes — de-hardcodes template + rendered HTML but enforces the OLD component-owns-dark contract keyed on is_dark_section, pulling opposite to the chosen architecture. Decision: initial-render fixes are primary; fixers get re-aimed to enforce the same paired-variable contract as backstop (key on what the template paints, never is_dark_section).
- **sources:** running_notes(22).md Sj, Sk; RUNBOOK_scheme_to_components(18).md WHERE WE ARE (fixer re-aim in Go batch)
- **relations:** hazard/band split; paired-variable direction; D2b (later analogous prevention-over-cure shape).
- **verify-later:** fix_forced_text_colors / fix_hardcoded_colors current keying; whether re-aim shipped.

<!-- SOURCE: U09_adoption.md -->
### Dormant discovery check: checkEmptyPageSections / validate_component_standards wrapper
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** "`checkEmptyPageSections` is stale, never-enabled code partially superseded by `EmptySectionsCheck` (which got the handler fix + enablement)… the wrapper `validate_component_standards` is in no agent's checks array (enables_vcs=f ×3)" (running_notes_15 Part 9).
- **what:** An existing sub-check targeting "page with no rendered sections" that never ran in production: its wrapper is not in any discovery agent's checks array, it is bundled with six unrelated sub-checks, scoped to deployed/active only (missing `planned` pages), and routes to page-content-writer (a task specialist with no persistence/deploy — cannot rebuild a page). Documented as the rationale for writing the dedicated sectionless check instead of extending it. General mechanism: `run_discovery_checks` runs only the names listed in the step-config `checks` array.
- **sources:** running_notes_15(10)#part-6, #part-9
- **relations:** check_sectionless_pages (its replacement for this case); EmptySectionsCheck (empty-HTML variant, enabled)
- **verify-later:** discovery_checks registry; completeness/design/quality-discovery-agent checks arrays

<!-- SOURCE: U09_adoption.md -->
### needs_section_data semantics and the abandoned standalone handler
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** "SUPERSEDED 2026-05-06 by FOCUS_directory_builder_and_list_components.md… needs_section_data items are emitted with status='needs_human_review' directly. They mean 'couldn't resolve component or required field — get a human to look.' Not async dispatch." (FUTURE_section_data_handler(1) header).
- **what:** The abandoned idea of a dedicated `needs_section_data` handler agent that would fetch list data asynchronously. Corrected understanding: those items are HITL by design; the list-data mechanism is query.* resolution. Residual real issue: plan_sections conflates two sub-cases — component missing (should emit needs_component to component-creator) vs required field genuinely unsourceable (real HITL) — and 41 items were stuck across six sites. `reconcile_section_data_action.go` exists for the deferred-query case but is not wired into any agent (and was a red herring for the guides-hub defect).
- **sources:** old2/FUTURE_section_data_handler(1).md, old/FUTURE_DEPRECATED_WRONG_section_data_handler.md, FOCUS_directory_builder_and_list_components.md#what-about-needs_section_data, running_notes_14(25)#part-14o–14p
- **relations:** directory-builder; plan_sections required-field switch; cta_url deferral bug
- **verify-later:** whether the 41 stuck items were triaged; reconcile_section_data wiring status

<!-- SOURCE: U10_imagery.md -->
### Corrupted component templates and the quality→regeneration bridge
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "✅ Bridge self-heal proven END-TO-END… no human involvement at any step. Fleet-remaining corrupted: 7" (2026-07-10); "10/14 healed" in the handoff.
- **what:** 14 components fleet-wide had html_template saved as RENDERED OUTPUT (literal `<no value>`, zero `{{…}}` vars) — historical damage from the pre-validation component-generation era (created_from='generated', 2026-03-31→04-13); the modern writer's pre-store validation already rejects this class. Detection existed (compute_component_quality flags "0 template variables") and repair existed (needs_component_regeneration → component-creator); the missing piece was a ~200-line bridge: `check_component_template_corrupted` discovery check (cross-site guard since components are fleet-shared, cap 5/pass). Field-preservation guard rejections are handled by re-queuing with exact field names in spec.description (rendered into the creator prompt).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turns-11–16/#Turn-20, SQL_2026-07-10_register_component_template_corrupted.sql, SHOWCASE_technical_architecture.md#5
- **relations:** tool-library/component-creator contract; the flagship "self-healing fleet" showcase story.
- **verify-later:** check_component_template_corrupted.go; remaining corrupted count fleet-wide.

<!-- SOURCE: U10_imagery.md -->
### mark_item_failed error honesty (flag-before-complete)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Fix applied (SQL_2026-07-10_pagebuild_mark_item_failed.sql… verified)… Failed page builds are now VISIBLE instead of silently complete."
- **what:** page-build-handler's step-level error routing pointed at `complete_error`, a SUCCESS-labelled complete_workflow — so a real step failure completed the orchestration and the dispatcher stamped the item 'complete' with no error (the "no-op complete" anomaly, triggered by a Kafka reply flake). The established flag-the-item-BEFORE-completing pattern was extended to real errors: a `mark_item_failed` step (update_work_item_status → 'failed', attempt-counted) inserted ahead of complete_error with all 8 error pointers repointed. Workflow-config-only. A fleet-trust principle: "a fleet you can trust starts with a fleet that tells the truth about itself."
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-16, SQL_2026-07-10_pagebuild_mark_item_failed.sql, SHOWCASE_imagery_workstream.md#4
- **relations:** Kafka partition race (the trigger); CompleteWorkItemAction guard semantics; likely needed on other handler workflows.
- **verify-later:** page-build-handler workflow error_step pointers; failed-vs-complete item stats post-fix.

<!-- SOURCE: U11_traffic_probe.md -->
### backend_unreachable discovery check
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "backend_unreachable REWRITTEN against the real DiscoveryCheck interface … gofmt-clean. Enable by adding … to the discovery agent's default_config checks array" — written and interface-reconciled, enablement not claimed.
- **what:** A discovery_checks/ check giving the improvement loop eyes on the VM class: per-site, NOOPs unless deploy_config.target='vm'; probes the public https://<domain>/health; on failure returns a WorkItemSpec (source='discovery', item_type='backend_unreachable', item_key dedup against idx_swi_dedup's partial unique index → one open alert per site, no spam). ALERT not auto-fix: handler_agent empty because a down VM isn't chassis-fixable — sits visible at 'detected'; the P5 vmhost adapter becomes the handler later. SELF-CLEARING: resolves its own open item on recovery using the runner's transaction. A companion `missing_beacon` check (rendered index lacks the /api/hit img) was floated and not built.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-e/f, traffic_probe_plan(12).md#P4, HANDOFF#cross-thread
- **relations:** VM-hosted backend sites class (first-class sites coverage), scheduler-and-tasks, vmhost adapter
- **verify-later:** discovery_checks/check_backend_unreachable.go in chassis; discovery agent checks array

<!-- SOURCE: U12_docs024_archives.md -->
### Colour-fix algorithmic detail (countHardcodedColorComponents / findForcedTextColors)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Full detail present in `009b_improvement_loop_v2.md`; deleted outright from `009c_improvement_loop_v3.md` onward and absent from live `004_improvement_loop.md` (table-row summary only).
- **what:** Documented exact algorithmic mechanics for two `design-discovery-agent` colour checks: `hardcoded_section_colors` and `forced_text_colors` (parses `<style>` blocks, flags only child text elements, skips container/link rules), with a WCAG AA 4.5:1 contrast safety check and `--section-*` contract injection. Pruned from the docs from v3 onward in favour of a one-line table entry.
- **sources:** old/older1/009b_improvement_loop_v2.md#"Colour Fix Detail"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** color-variable-fixer handler; contracts-and-standards CSS variable contract
- **verify-later:** `fix_hardcoded_colors`/`findForcedTextColors` Go source accuracy.

<!-- SOURCE: U12_docs024_archives.md -->
### Per-site, per-audit-type cadence configuration (maintenance_profile.audit.{type})
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** Appears identically in v2-v4 as a "## Configuration" section, each time caveated "future enhancement." Absent from v5 and live, which document a simpler global 60-day auto-reset with no per-audit-type knobs.
- **what:** Three consecutive doc versions carried a designed-but-never-built configuration surface: per-site JSON config letting each audit type be individually enabled/disabled with its own re-run interval. Quietly dropped rather than implemented.
- **sources:** old/older1/009b_improvement_loop_v2.md#"Configuration"; old/older1/009d_improvement_loop_v4.md#"Configuration"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** Audit Pass Cap / Auto-reset mechanism (its replacement)
- **verify-later:** check `sites.settings.maintenance_profile` rows for leftover `audit.{type}` keys.

<!-- SOURCE: U12_docs024_archives.md -->
### Acceptance-test cheap-LLM verification call gating lock + retry
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Documented in `009c_improvement_loop_v3.md`/`009d_improvement_loop_v4.md`, incl. literal verification prompt. Live `004_improvement_loop.md` retains `acceptance_test` as a required field but documents no corresponding verification-call step.
- **what:** Each finding carried an `acceptance_test` enabling a cheap follow-up LLM call after a fix: feed fixed HTML back, get YES/NO, gating section lock (pass) or retry up to `max_fix_attempts` before escalating to `needs_human_review`. The field survived but the explicit verify-then-lock mechanism dropped out of documentation by v5/live.
- **sources:** old/older1/009c_improvement_loop_v3.md#"Structured Findings Format"; docs024_key_docs_latest/004_improvement_loop.md#"1. Finding Cap"
- **relations:** Section Locking; Finding Cap
- **verify-later:** search for a dedicated verification-call step (`verify_fix`/`check_acceptance_test`) in fixer code.

<!-- SOURCE: U12_docs024_archives.md -->
### Content-writer chrome double-injection bug and chrome-ownership rule
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** `009d` v4 "## Content Writer Chrome Fix (v4)": full bug narrative + cleanup; only a one-line changelog bullet survives into v5/live.
- **what:** A production bug rule: site chrome (header/footer/head) must be injected exactly once, only at the rerender/assembly step — never by the content writer. Fix set all three inject flags false on `page-content-writer`'s `compile_page` step plus a cleanup pass removing baked-in chrome components.
- **sources:** old/older1/009d_improvement_loop_v4.md#"Content Writer Chrome Fix (v4)"; docs024_key_docs_latest/004_improvement_loop.md (changelog line only)
- **relations:** contracts-and-standards (component/slot contract); site-component-linker
- **verify-later:** confirm `page-content-writer` inject flags remain false; check for reappearance of baked-in header/footer components.

<!-- SOURCE: U12_docs024_archives.md -->
### Audit finding dedup + blocked-item filtering algorithm (write_audit_findings)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** Full three-step algorithm documented in v3/v4; not present in v5/live (mentioned only in passing, then dropped even from the summary line).
- **what:** `write_audit_findings` was documented as implementing three dedup/safety layers: bulk-preloading blocked item keys, a broader item_type+page match against existing blocked items, and item-key-based dedup against pending items. This mechanism-level detail disappears from the documentation surface after v4.
- **sources:** old/older1/009d_improvement_loop_v4.md#"Finding Dedup and Blocked Item Filtering"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** Finding Cap; Triage Drain Controls
- **verify-later:** confirm `write_audit_findings` still implements bulk-preload + item_key pattern.

<!-- SOURCE: U12_docs024_archives.md -->
### Triage drain loop — structured audit findings, capped passes, section locking
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Live `009_model_infrastructure.md` "Done": structured findings, audit pass cap, section locking exclusion all checked off.
- **what:** Fix for unbounded audit/fix/re-audit token spend. Findings must carry `acceptance_test`/`acceptance_levels`/`minimum_required`. Audits capped at 3 numbered batches per site. Passing sections get `locked_at`; subsequent audits skip them; unlock is manual. Per-page sequential processing via `depends_on` prevents overlapping fixes.
- **sources:** old/older1/020d_gpu_and_model_infrastructure_v4.md#"Triage Drain Loop Fix"; docs024_key_docs_latest/009_model_infrastructure.md#"Decisions Made"
- **relations:** three-way audit-finding classification; GPU/AI-endpoint scheduling
- **verify-later:** `write_audit_findings_action.go`; section-lock column on `page_components`.

<!-- SOURCE: U12_docs024_archives.md -->
### wont_fix/superseded dedup and needs_section_data data-honesty pattern
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Described as "correct behaviour... the dedup system working" in the second draft.
- **what:** When a recurring issue is detected while an older item is stuck, the loop creates a new item and marks the old one `wont_fix` ("superseded by active duplicate") — expected noise, not a bug. `needs_section_data` items requiring unfabricatable data (bios, pricing, case studies) correctly route to `wont_fix`/HITL rather than inventing content.
- **sources:** old/older1/016_debugging_guide_v2_april26.md#"9. Specific Failure Patterns"
- **relations:** debugging playbook; needs_section_data triage
- **verify-later:** current dedup logic; HITL routing for `needs_section_data`.

<!-- SOURCE: U12_docs024_archives.md -->
### plan_sections pre-check → plan-then-reconcile evolution
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** v3 report P5 "DEPLOYED" (pre-check); live v4 shows the same P5 row "UPDATED" with a materially different mechanism.
- **what:** The original fix for wasteful LLM re-sends on sections with pending `needs_section_data` was a pre-check that simply skipped them. Revised to "plan-then-reconcile": ready sections auto-close stale data requests, deferred sections create new requests while skipping duplicates.
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P5)"; docs024_key_docs_latest/105_dispatch-pipeline-failures-report_v4.md#"Priority Fixes (P5)"
- **relations:** early pipeline-failure triage priorities; needs_section_data triage
- **verify-later:** current `plan_sections_action.go` logic for auto-closing `needs_section_data` items.

<!-- SOURCE: U12_docs024_archives.md -->
### Three-way audit-finding classification (bug / recommendation / gap)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** v3 report's P10 marked "DEFERRED... ~1 week project"; referenced independently, still not built, in `027_design_and_site_planner_v2.md` months later.
- **what:** Auditors currently produce findings the pipeline auto-fixes uniformly as if bugs, but many are opinions/recommendations — producing false-positive fix attempts. Proposed fix: three-way classification with dedicated specialist agents per category and per-site approval mode for recommendations.
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P10)"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"10. Open Design Areas"
- **relations:** audit gap-finding routing fix; triage drain loop
- **verify-later:** existence/status of `design-note-recommendation-specialists.md` or any implementing specialist agent.

<!-- SOURCE: U12_docs024_archives.md -->
### Audit gap-finding routing fix (existing-page gaps → needs_content_page)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** v3 report P9: "FIX WRITTEN — write_audit_findings_action.go: Rule 4 routes gap findings on existing pages to needs_content_page, not content_rewrite."
- **what:** Gap findings on existing pages were being routed to `content_rewrite` (edits, not rebuilds), causing validation-failed rewrites. Rule 4 redirects them to `needs_content_page` (full rebuild path).
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P9)"
- **relations:** three-way audit-finding classification; needs_content_page work-item type
- **verify-later:** current `write_audit_findings_action.go` Rule 4 logic.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Automation ratchet (per-capability trust levels)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §2 "Automation is a ratchet, not a switch"; a capability graduates `confirm-every → confirm-exceptions → notify → autonomous` "when evidence shows it reliable"
- **what:** Automation is not global; each capability (create action, provision nginx, reshard DB) carries its own trust level and graduates only on evidence. "Fully automated" is the union of individually-graduated capabilities. A trust ledger records each capability's level, gate policy, and supporting evidence.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#2, ED/MASTER_autonomous_build_and_operate(4).md#8.1
- **relations:** trust ledger; reliability cascade; bidirectional ratchet
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Reliability cascade (reuse → generate+verify → compete+judge → HITL)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §3 "Applied to every sub-task, in descending order of reliability"
- **what:** A per-task router for producing any unit of work in descending reliability order: known-good reuse, then generate+deterministic-verify, then compete-N-and-judge in a sandbox, then HITL. A verified recurring generated solution becomes candidate known-good and graduates its gate — feedback into the ratchet.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#3, ED/MASTER_autonomous_build_and_operate(4).md#7.2
- **relations:** cascade router; known-good library; multi-author generation
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Trust ledger + gate-policy engine
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7.2 "The trust ledger is the master knob"; §8.2 "Stand up the ledger (a table)"
- **what:** A per-capability store of automation level, gate policy, and supporting evidence, plus a small evaluator mapping (capability, trust level, stakes) → gate. It is the master knob: it governs both how conservatively a thing is produced and whether the result applies without a human.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** automation ratchet; cascade router; governance decision package
- **verify-later:** trust-ledger table (proposed)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Cascade router
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §6.3 "Cascade router — the per-task decision (reuse / generate / compete / HITL)"; §8.4 "the loop's least-bounded step"
- **what:** An action/agent that picks a cascade tier per leaf task from three inputs — the capability's verifiability/containment, its trust-ledger entry, and the task's stakes. Named as the loop's least-bounded step, so conservative-by-default and ledger-gated.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#8.4
- **relations:** reliability cascade; trust ledger; verification harness
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Bidirectional ratchet (trust can be lost)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7.5 "Trust can be lost, not only gained … the safety property that lets the ratchet advance at all"
- **what:** Feedback is two-directional: success accrues evidence toward graduation; repeated/severe failure drops the trust level, tightens the gate and raises the cascade floor. A regressing capability is automatically pulled back under supervision.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** automation ratchet; trust ledger
- **verify-later:** none

<!-- SOURCE: U18_sql_for_agents.md -->
### Discovery agents (design / quality / completeness) and the check registry
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 047/048 define them; 074 expands completeness checks; 142/146 still adding checks to design-discovery in 2026-07 ("run_discovery_checks warns... and skips unknown names" — the safe-rollout pattern).
- **what:** Read-only detectors that "find problems. They do not fix anything. They do not call other agents." Each runs run_discovery_checks with a named check list, writing findings to site_work_items (source='discovery', status='detected'). design: undeployed_assets, missing_css, duplicate_palette, missing_tools, tool_health, tool_acceptance, tool_acceptance_due. quality: broken_nav_links, placeholder_contact, generic_theme. completeness: empty_sections plus integrity checks — cross_site_contamination, unrendered_templates, missing_style_collection, deactivated_site_components. All algorithmic, no LLM budget. Unknown check names warn-and-skip, so SQL can enable a check before the Go ships.
- **sources:** 047_discovery_checks.sql; 048_discovery_agents.sql; 058_quality_checks_and_fixers.sql; 074_completeness_discovery_agent.sql; 142_enable_tool_acceptance_check.sql; 146_enable_tool_acceptance_due.sql
- **relations:** improvement-loop orchestrates them; fixer agents consume their items; check registry in discovery_checks.go
- **verify-later:** registered checks in run_discovery_checks_action.go / discovery_checks/*.go

<!-- SOURCE: U18_sql_for_agents.md -->
### improvement-loop (post-build discovery → triage → fix → rerender cycle)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 054 definition; 086 Part B adds audit_pass_count guard "stops after 3 passes"; 100 portfolio claims sites "receive autonomous content audits... on rolling schedules".
- **what:** Runs after initial build (or on schedule/manual trigger): spawns the three discovery agents, triage_detected_items promotes detected → triaged, and if anything was promoted inserts needs_rerender at priority 99 and fires build-dispatch-loop to process all fixes then rerender. 086's audit-pass cap plus section locking provide the loop's termination condition ("the triage drain").
- **sources:** 054_improvement_loop.sql; 086_visual_design_auditor.sql; 061_tool_deployer_and_discovery_agent.sql (flow diagram)
- **relations:** discovery agents, fixers, audit agents, locks
- **verify-later:** improvement-sweep scheduled task; triage_detected_items action; audit_pass_count in sites.settings

<!-- SOURCE: U18_sql_for_agents.md -->
### Fixer agents: color-variable-fixer, site-component-linker, component-template-fixer, css-patch-agent
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 056/064/076 definitions; all in 075's idle-timeout list; component-template-fixer gains note-writing in 132.
- **what:** Narrow algorithmic/LLM fixers dispatched from the queue: color-variable-fixer replaces hardcoded hex in component inline styles with CSS variables (fixes both templates permanently and rendered_html immediately); site-component-linker fixes NULL component_id causing fallback rendering; component-template-fixer applies targeted template surgery (nav flex CSS injection, element removal, slot_name alignment) routed on spec.fix_type; css-patch-agent LLM-patches the current stylesheet for spacing/responsive/layout issues without full regeneration (explicitly NOT theme redesign — that's webdesign-agent). All create deduplicated needs_rerender items only when they changed something.
- **sources:** 056_colour_variable_fixer.sql; 064_site_component_linker_and_fixer.sql; 076_css_patch_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** discovery checks (hardcoded_section_colors, unlinked components); rerender pipeline; audit findings
- **verify-later:** fix_hardcoded_colors, link_site_components, fix_component_template actions

<!-- SOURCE: U18_sql_for_agents.md -->
### Audit agent hierarchy (visual-design-auditor, content-quality-auditor, design-audit-agent, site-review-agent)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 066 defines the hierarchy; 084 patches prompts due to a real cost incident ("845 design-audit work items across 4 domains in ~10 days... cost explosion"); 086 excludes locked components from audit queries.
- **what:** LLM auditors layered above discovery: pattern is "algorithmic checks first, then ONE LLM call for subjective assessment, then write findings" (write_audit_findings). 084 makes findings structured and bounded: TOP 5 only, every finding must carry current_value, a concrete `acceptance_test` "that a DIFFERENT agent could verify without re-auditing", max_fix_attempts, and must skip what algorithmic checks already caught. site-review-agent adds strategic alignment review; unclassifiable gaps become needs_content_planning items for content-gap-planner.
- **sources:** 066_audit_agent_definitions.sql; 084_site_review_agents.sql; 086_visual_design_auditor.sql; 071_content_gap_planner.sql
- **relations:** locks (locked_at exclusion); improvement-loop pass cap; fixers consume findings
- **verify-later:** write_audit_findings action; current audit prompts vs 084 text

<!-- SOURCE: U19_sql_tables_components.md -->
### Improvement-sweep and build-pipeline-trigger scheduling
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Seeded tasks with evolving pre_queries: queue-size gate (skip when >20 open build items), round-robin site selection by least-recently-checked, skip sites with claimed items or locks.
- **what:** The improvement loop's cadence lives in scheduled_tasks: build-pipeline-trigger (2 min) finds sites with triaged/approved items and fires the dispatch loop; improvement-sweep (10 min) picks the next site for discovery checks, gated so discovery never floods an already-backed-up queue and locked sites are skipped. Both share the 'dispatch' concurrency group.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#seed-data and #improvement-sweep-fixes
- **relations:** scheduler; work queue; site-level lock.
- **verify-later:** current pre_query for improvement-sweep; discovery agent set.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Heartbeat maintenance model (findings-based, pre-work-items)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** v1 (docs017/021) findings table + domain orchestrators; v2 (docs017/022) full spawn chain modeled on "the vet-batch pattern" with budget management; explicitly replaced by 023: "maintenance-triage + page-rebuild → maintenance-batch-scheduler + site-work-orchestrator".
- **what:** The first full maintenance architecture: K8s CronJob (8h) → agent-chassis spawns maintenance-batch-scheduler → claims batch (FOR UPDATE SKIP LOCKED, batch_size controls concurrency) → per-site site-maintenance-orchestrator runs fix-pending → verify-previous → discover-due → triage cycle; discovery agents per domain (content/links/seo/compliance/structural) write maintenance_findings; triage (a step, not an agent) enriches with impact reads and classifies resolution path (auto_fix/suggest/flag/monitor/ignore); narrow fix agents resolve; cross-domain coordination only via side-effect findings with parent_finding_id — "no agent calls another agent for coordination." Daily maintenance-catch-all handles stale findings, HITL reminders, cross-site patterns, stuck recovery.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#8-Maintenance-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/021_maintenance_architecture_plan_v1.md
- **relations:** vet-batch-processor precedent (vet-med-pricing); unified work items (successor); scheduler-and-tasks; maintenance profile.
- **verify-later:** maintenance_findings/maintenance_tasks tables; maintenance-batch-scheduler agent history.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Unified build & maintenance work items (site_work_items)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** docs017/023: "Build and maintenance are the same process. A new site is a set of findings that need fixing"; full site_work_items DDL with item_key dedup, depends_on, parent_item_id, batch claiming; docs017/044 traces the working site-work-orchestrator step by step.
- **what:** The pivotal unification: every piece of work — building a page, fixing stale content, adding a tool, publishing an article — is one work item with source (planner/discovery/content_feed/manual/improvement/side_effect), domain, item_type, severity, spec JSONB, triage enrichment (impact, resolution_path, priority, handler_agent), lifecycle statuses detected→triaged→approved→claimed→in_progress→complete→pending_verify→verified, dependencies, dedup keys, attempt limits, and archival. The planner becomes a discovery agent writing 'needs_content_page' items; the same orchestrator/fix agents process build and maintenance; sites start minimal and improve incrementally, always left in a working state; per-page git commits. Old and new systems coexist (v2 intake routes between pageflow-builder and site-work-orchestrator). This is the direct ancestor of the current work-item lifecycle and improvement loop.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** heartbeat model (predecessor); maintenance_queue/page-rebuild (earlier still); work-item lifecycle in development-guide (current form); news feed → work items.
- **verify-later:** site_work_items vs current work_items table naming/shape; site-work-orchestrator vs current orchestrators.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Per-site maintenance profile with budgets
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** docs017/019b maintenance_profile JSON in sites.settings ("content every 7d, links every 8h... budget: llm_calls_per_cycle: 20, max_auto_fixes_per_cycle: 5"); 023 extends with content_feed cadence and time_sensitivity.
- **what:** Each site declares which maintenance domains run, at what cadence, with which sub-agents and regulatory bodies, plus hard budgets on LLM calls and auto-fixes per cycle — a finance site gets hourly high-sensitivity feeds and FCA compliance, a brochure site gets links+freshness only. Ancestor of the growth budget concept.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Per-Site-Configuration; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Per-Site-Configuration
- **relations:** content-governance growth budget (descendant); scheduler cadence.
- **verify-later:** maintenance_profile key in sites.settings rows.

<!-- SOURCE: U22_recent_small_docs.md -->
### maintenance-triage agent
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** Defined with dry_run mode; workflow scans sites, queues page_rebuild tasks, spawns page-rebuild agent; described alongside "for future use" queue.
- **what:** An orchestrator that scans deployed sites for maintenance issues (stale pages, missing pages, broken links, CSS drift), inserts tasks into `maintenance_queue`, then dispatches specialist agents (page-rebuild) per affected site. Supports dry_run (scan+queue without dispatch) and a configurable stale_threshold_days.
- **sources:** docs019_business/017_maintenance_triage_agent.sql
- **relations:** maintenance_queue, page-rebuild agent, improvement-loop
- **verify-later:** agent_definitions type='maintenance-triage'; actions scan_sites_for_maintenance/prepare_rebuild_dispatches

<!-- SOURCE: U22_recent_small_docs.md -->
### build-dispatch-loop self-chaining removal
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Fix: Loop back to load_next_item internally ... Status: Applied to production DB. Verified live definition matches migration."
- **what:** A fix (migration 063) removing the build-dispatch-loop's self-respawn pattern (spawn_next_dispatch → call_next_dispatch), which repeatedly left the parent stuck in AWAITING_RESPONSES when the child's Kafka response was lost to topic retention/pod restarts. Now loops back to load_next_item internally (9 steps vs 13), timeout bumped 900→1800s.
- **sources:** docs021.../024_handoff_summary_2026_03_02.md#fixes-applied
- **relations:** work-item lifecycle, dispatch loop, orchestration timeouts
- **verify-later:** build-dispatch-loop agent definition step count; migration 063

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Triage drain loop fix (bounded audit passes + section locking)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** 020c "Triage Drain Loop Fix": 845+ design-audit items across 4 domains in ~10 days, "The loop has no termination condition"; Fixes 1-5 (structured findings with acceptance_test, max 3 passes, section locking via `page_components.locked_at`, verify-against-criteria not re-audit, per-page `depends_on` sequencing); "~65-70%" token reduction.
- **what:** The audit→fix→re-audit loop ran unbounded, consuming most tokens. Fix: auditors emit structured findings with `acceptance_test`/`acceptance_levels`; cap at 3 audit passes per site producing numbered batches; lock passing sections (`locked_at`) so later audits skip them (unlock always manual); verify via a cheap acceptance-test call (not full re-audit); sequence same-page items via `depends_on`. This is the origin of the improvement-loop guardrails referenced across imagery/content docs.
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#triage-drain-loop-fix
- **relations:** section locking (page_components.locked_at); imagery audit loop 3-pass cap; recommendation specialists
- **verify-later:** audit agent acceptance_test emission; sites.settings audit pass cap

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel B — RAG knowledge base with nomic task prefixes
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4b "Step 3 (chassis integration) passed on 2026-04-21 … Flywheel B is done"; "Prefix patch deployed and verified live 2026-04-21"
- **what:** `knowledge_base` pgvector(768) table read/written by `rag_lookup`/`rag_index` actions on the cpu-ollama nomic-embed-text endpoint, with trigram fallback. Empirically established that nomic `search_document:`/`search_query:` task prefixes are load-bearing for ranking (French Bulldog BOAS test), now patched into production. Best practice: filter by metadata (vertical, component_type, source) first, then rank by similarity.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.2, #2.4b; #14 (Ollama specifics)
- **relations:** short-term lever paired with LoRA (long-term); the flagship of the finetuning.uk RAG product
- **verify-later:** platform/orchestration/actions/rag_actions.go (applyNomicPrefix); knowledge_base table; PATCH_rag_actions_nomic_prefixes

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Defect-cataloguing discipline (enumerate-before-fixing)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** `CATALOGUE_gamesdesign_post_sync_fix_defects.md` states its purpose explicitly: "Enumerate every observed defect as a *separate* item before fixing, so distinct causes are not conflated into one rolling investigation," with causes marked "tentative" until confirmed by reading source. Later revisions (`(4)`) show the discipline paying off: defects graduate from `[NEW]`/`[PARKED]` through `[FIX SHIPPED — PARTIALLY VERIFIED]` to `[VERIFIED CLOSED]` with a pinned, source-read cause replacing the original tentative one.
- **what:** A working method for a real adoption-run defect sweep: group symptoms into lettered families by shared mechanism (A deployment gaps, B link fallbacks, C list-component content, D section-data gaps, E content quality, F guide duplication, G design fidelity, H hygiene, I open unknowns, J dispatch throughput), triage by root cause not symptom, and forbid shipping a fix from a "tentative" cause without first reading the responsible action.
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects.md; adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md
- **relations:** running-notes debugging-log convention (below), silent-completion family
- **verify-later:** whether this catalogue format was formalised anywhere beyond this one adoption run.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Dormant discovery-check machinery (`checkEmptyPageSections` / `validate_component_standards`)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** running_notes_15(5) Part 6: "**RESULT (decisive):** ... `validate_component_standards` (its wrapper) is **not enabled in any** discovery agent (`enables_vcs=f` for all three)... The empty-page detector is **dormant code**, not a buggy check."
- **what:** A pre-existing check (`checkEmptyPageSections`, inside `ComponentStandardsCheck`/`validate_component_standards`) already targets exactly "page with no rendered sections," but was never added to any discovery agent's `checks` config array, so it has literally never fired in production — its 11 historical `needs_content_page` items were all traced to adoption-run/manual sources, none to this check. It was also found to be scoped too narrowly (`deployed`/`active` only, missing `planned`) and to recover by re-emitting a still-empty spec (would loop, not repair) — reasons a *new* dedicated check (`check_sectionless_pages.go`) was written instead of extending the dormant one.
- **sources:** adoption/running_notes_15_skinner_box_and_adoption_sections(5).md Parts 6–8
- **relations:** sectionless-page silent completion (above)
- **verify-later:** `discovery_checks/` registry current contents; whether `validate_component_standards` has since been enabled anywhere.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `improvement-sweep` scheduled task — deliberately disabled pending consumer readiness
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** running_notes_17(16): "Operational: `improvement-sweep` scheduled_task is **disabled** (`enabled=f`, last completed 2026-05-08), intentionally paused during core build... Before re-enabling: have the `phantom_internal_links` check enabled AND both handler agents in place (`nav-link-fixer` exists; `internal-link-resolver` is Step 3), so resuming the sweep clears findings rather than accumulating them." Later resolved to a specific enablement gate (§7) confirming per-finding routing survives `run_discovery_checks_action.go`'s pipeline stamping, so the check could finally be enabled "observe-only" without turning the sweep back on.
- **what:** The discover→triage→fix improvement loop's top-level scheduler is deliberately kept off while core build work is in flight, on the explicit policy that a discovery check should only be enabled once its handler agent actually exists — otherwise findings accumulate unconsumed rather than clearing. This is a recorded operational policy (not a bug) governing when automation is safe to turn back on.
- **sources:** content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md, "Policy settled" and "§7" sections
- **relations:** dormant discovery-check machinery; internal-link-resolver agent
- **verify-later:** current `enabled` state of `improvement-sweep` in `scheduled_tasks`.

<!-- SOURCE: U25_leopardess_social.md -->
### `needs_rebuild` is inert without an explicit work item
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK O3: "pages.build_status='needs_rebuild' does nothing on its own … This is why six pages have sat at needs_rebuild doing nothing" (2026-07-10).
- **what:** The build-dispatch-loop reads site_work_items and never scans pages; only write_build_items converts needs_rebuild to work items and it lives inside site-work-orchestrator/build-site-planner, not the loop. Operator remedy: INSERT site_work_items rows (needs_content_page → page-build-handler) explicitly. Related dispatch facts: claim_work_item blocks items whose handler_agent doesn't exist; unhealthy AI endpoints release items; the partial unique index idx_swi_dedup silently suppresses new items sharing (site_id, item_key) with an open one.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#O3, #O4, #landmines-1/9; docs/leopardessconsulting/HANDOFF.md#4.1
- **relations:** silent no-op success class; work-item dedup semantics
- **verify-later:** build-dispatch-loop workflow; write_build_items call sites

<!-- SOURCE: U25_leopardess_social.md -->
### build_status 'approved' invisibility defect and its layered fleet fix
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_minilobby §0: writer fix "CLOSED — fixed in chassis v1.0.1102, verified end-to-end 2026-07-10"; CHECK constraint "CLOSED 2026-07-11 — migration 049 applied + negative-tested"; drift check shipped v1.0.1104.
- **what:** apply_section_edit left an edited live section at build_status='approved' while every discovery check filters ='deployed' — a live section silently invisible to the whole audit surface (same shape as complete_error). Fixed at four layers: writer (UpdatePageStatusAction gained page_component_id_field, mirroring the deploy mark onto the named page_component; coordinator dataRefKeys registration), invariant (migration 049 CHECK constraint on the previously free-text column — invented statuses now fail loudly), detection (new check page_component_status_drift: unknown status → emit; pending-on-deployed-page → finding only, since a status flip would hide real staleness — 19 such rows surfaced across 5 sites), repair (new fixer repair_page_component_status with refusal guards). Leopardess RUNBOOK carries the interim two-statement manual repair as landmine 3.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.0, #8.1; docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#3-#4; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10
- **relations:** section-editor; fleet generalisation doctrine; discovery check wiring
- **verify-later:** v3_site_actions.go UpdatePageStatusAction; migration 049; check_page_component_status_drift.go

<!-- SOURCE: U25_leopardess_social.md -->
### Runtime-fill guards in discovery checks (three defused landmines)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-10 (night): v1.0.1105 "third guard PROVEN … the spawned pod logged all four exemptions by name"; guard split 3 emit / 2 skip verified live.
- **what:** Three discovery checks would have dismantled vonc's runtime-fill shells the moment the improvement loop switched on: component_template_corrupted (would regenerate shells into build-time copy), broken_template_slots (endless churn of version rows), and empty_sections (already enabled — caught live raising full LLM rebuilds of the shells within minutes of the first pass). Guard pattern: key on the data-runtime-fill marker (author's declared intent, never component names), exclude from work-item emission but record a Findings entry — a bare SQL NOT LIKE was rejected because silent skipping is the codebase's recurring failure mode. For runtime-fill shells `<no value>` IS the mechanism, not the defect.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 entries; docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#3 (#3, #14)
- **relations:** runtime-fill mechanism; Mode-B templates; fleet generalisation doctrine
- **verify-later:** check_component_template_corrupted.go, check_component_standards.go, check_empty_sections.go guards in the running binary

<!-- SOURCE: U25_leopardess_social.md -->
### Discovery-check wiring gaps (registered-not-enabled / enabled-not-implemented / sweep off)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** PLAN_generalise §3 #6/#7: "8 checks registered in Go, enabled in no agent — incl. sectionless_pages, the exact detector for the ten silent complete_error builds"; "3 enabled check names have no Go implementation"; improvement-sweep "disabled since 2026-05-02, so discovery runs only by manual trigger".
- **what:** The improvement loop's configuration surface has three drift modes: checks registered in Go but named in no agent's checks array (inert capability); check names enabled in agents with no implementation (runner warns "Unknown discovery check"); and the scheduler that drives the whole loop disabled for two months, making discovery manual-only — which is why design-discovery-agent had never run on vonc and undeployed_assets never fired. Enable decisions are deliberate per-agent config edits (three checks added to completeness-discovery-agent 2026-07-10, backed up first); the remaining candidates are listed survey-first.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#5; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#4; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12 (first design-discovery run on vonc ever)
- **relations:** silent no-op success class; runtime-fill guards; scheduler-and-tasks (improvement-sweep)
- **verify-later:** agent_definitions checks arrays vs Go-registered Name() set; scheduled_tasks improvement-sweep row

<!-- SOURCE: U25_leopardess_social.md -->
### Fleet generalisation doctrine (four rules + artifact verification)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** PLAN_generalise §2 (2026-07-10), applied to all 14 findings; "of thirteen findings exactly one is site-specific".
- **what:** The doctrine for turning incident fixes into fleet guarantees: (1) fix the writer, not the row — a psql repair scoped to a site_id is evidence of a bug, never its fix; (2) detect by contract, not by name — guard on declared markers, never component names; (3) surface, never silently skip — where a check must not act it emits a Finding; (4) every detection needs a fixer, or a written reason there is none. Fifth inherited rule: verify by artifact, never item status. The mini-lobby trim is the worked example: "a keyhole onto fleet-wide defects".
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#2, #3
- **relations:** operator discipline; build_status fix layers; runtime-fill guards
- **verify-later:** n/a (doctrine)

<!-- SOURCE: U25_leopardess_social.md -->
### Work-item dedup and two-strike semantics (partial index behaviour)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-10: "idx_swi_dedup is partial — it only covers non-terminal rows, and the two-strike rule only counts complete/failed. So a completeness pass … will re-raise the two rejected shell items."
- **what:** Dedup semantics that shape operations: the partial unique index suppresses a new work item only while an open one shares (site_id, item_key) — rejected/terminal items are outside it, so a re-run re-raises them (which made a rejected item's *absence* after the guarded build positive proof the guard worked). Leopardess side: the same index silently suppresses intended new items when an open twin exists. Co-page duplicate rule: a page rebuild is whole-page, so multiple empty_section items on one page are duplicates — close the second by artifact.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 (evening/night); docs/leopardessconsulting/RUNBOOK.md#landmine-9; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 (dispatch note)
- **relations:** needs_rebuild inert; discovery wiring; stuck-claim dispatch noise
- **verify-later:** idx_swi_dedup definition; two-strike logic in dispatch

<!-- SOURCE: U01_docs024_numbered_core.md -->
### QA three layers, group auditor agents, and the promotion pattern
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 002d architecture with live agents (design-audit-agent, visual-design-auditor, content-quality-auditor, site-review-agent)
- **what:** Layer 1 structural checks (algorithmic, every cycle), Layer 2 group LLM audits (shared context, ONE LLM call per group), Layer 3 strategic review. Group agents chosen over per-check agents (context reuse) and over a mini-action registry (every agent is an orchestrator); a check is promoted from action step to spawned agent by changing one workflow line, output_field unchanged. Site type enables audit groups via maintenance_profile.audit.
- **sources:** 002(4)/002d#Quality Assurance; #Promotion Pattern
- **relations:** improvement loop; audit-enforces-not-overrides
- **verify-later:** design-audit-agent workflow; maintenance_profile.audit config

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Audit enforces intent, doesn't override (chain of authority) + propose mode for spec-less sites
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 002d responsibility boundaries; resolved decision 20
- **what:** classifier decides intent → planner implements → composition installs → webdesign renders → audit checks build vs stated intent and emits work items; it never makes design decisions. Exception: where no classifier output exists, audit may PROPOSE a direction, flagged for HITL. Handlers never know their trigger (build/audit/manual all use the same webdesign-agent).
- **sources:** 002d#Responsibility Boundaries; 004#Human Direction Integration
- **relations:** dream spec; direction spec; 028 silent-override failure mode
- **verify-later:** auditor prompts reference design_intent/direction

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Improvement loop flow with pass cap, finding cap, and auto-reset ("sites breathe")
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 004 v3/v4 dated changes (2026-03-25/31); token math before/after (~88K→~30K)
- **what:** improvement-loop orchestrator: pass-limit gate (≥3 → complete_clean) → algorithmic discovery agents (quality/design/completeness) → LLM audits (TOP 5 findings each with current_value/acceptance_test/suggestion/max_fix_attempts) → triage → increment pass → insert needs_rerender p99 → dispatch. Auto-reset after 60 days / direction change / major rebuild / manual, pairing with lock expiry (the unimplemented half) to create the breathe rhythm.
- **sources:** 004 full; 031_LOCKS (the missing half)
- **relations:** section locking; direction spec; locks-should-expire question
- **verify-later:** improvement-loop workflow; audit_pass_count fields

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Discovery checks catalogue (quality/design/completeness) and ordering rule
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 004 tables with handlers per check; validate_component_standards sub-checks
- **what:** quality: broken_nav_links, placeholder_contact, generic_theme. design: hardcoded_section_colors, forced_text_colors, undeployed_assets, missing_css, validate_component_standards (9 sub-checks incl. unlinked components, slot mismatch, nav layout, unwanted search icon). completeness: empty_sections, empty_blog, orphan_pages. validate_component_standards runs BEFORE colour checks (structure before rendered-HTML checks). DiscoveryCheck interface: checks append WorkItemSpecs; the runner inserts with dedup — plugins must not insert their own rows.
- **sources:** 004#Discovery Agents, #Component Standards Validation; 016 §0 item 27 (interface shape)
- **relations:** two-strike dedup; handler routing
- **verify-later:** discovery_checks/ registry

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Blog listing rebuild and slot-detection strategy
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 103 handoff: rewritten + deployed with three bug fixes; 004 shows it in rerender-pages workflow
- **what:** rebuild_blog_listing runs in rerender-pages before get_pages: finds the actual listing slot via priority list → pages.sections → default; loads a template that genuinely has a {{range}} (guard against CSS-only components); ensures article links (SQL template patch + post-render safety net); writes content_data alongside rendered_html; computes read time from component lengths.
- **sources:** 103_blog_nav_handoff-2026-04-12.md; 004#Blog Listing Rebuild
- **relations:** empty_blog check → blog-content-planner
- **verify-later:** rebuild_blog_listing_action.go current form

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Dispatch failures triage report and the bug/recommendation/gap classification gap (P10)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** 105 v4 (2026-04-15): P1–P6 deployed/done, P7/P8 human-gated, P9 fix written, P10 deferred with design note
- **what:** Systematic queue triage produced ten priority fixes: component_id nil in load_tool; retry-safe fork deploys; rate-limit errors classified transient; load_page_record page_id fallback; plan-then-reconcile for section data requests (auto-close stale); design chain unblocked; audit gap findings rerouted to needs_content_page. P10 names an architectural gap: auditors emit opinions (recommendations) that the pipeline auto-fixes as if bugs — proposed three-way classification (bug/recommendation/gap) with specialist agents + per-site approval mode (~1 week, deferred).
- **sources:** 105 full
- **relations:** write_audit_findings; approval model (P1 plan)
- **verify-later:** P9/P10 status since April

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery agents on dead/stub sites (noise at scale)
- **category:** improvement-loop
- **status-signal:** unknown
- **status-evidence:** "80+ needs_content_planning items for gamesdesign.co.uk, mostly [stale: triaged 48h+] … running on a sites row whose adoption had failed or been deleted" (2026-04-23, item 19, added in version (1) only)
- **what:** Discovery agents keep generating remediation items for sites that are deleted, stubs, or mid-adoption. Proposed precondition: skip site_ids with status deleted/archived, no current identity spec, or adoption in flight.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md item 19
- **relations:** two-strike pile-up; library-row cleanup pattern
- **verify-later:** discovery agent site-selection queries

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Recommendation specialist architecture (bug vs gap vs recommendation)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** "P10 … Deferred until HITL queue becomes a bottleneck" (April 2026); P9 (gap → needs_content_page) deployed
- **what:** LLM auditors mix factual bugs with opinions; the pipeline shouldn't auto-fix opinions. Proposed: finding_type classification (bug → auto-fix; gap → rebuild via needs_content_page — this part deployed as P9; recommendation → specialist agent decides apply/dismiss/escalate, e.g. identity-advisor for contact details), with per-site approval_mode (auto|review) gating.
- **sources:** HANDOFF-pipeline-triage-april-2026.md#P9, #P10; FOCUS_content_quality.md#machinery
- **relations:** content quality work order; two sources of truth for email
- **verify-later:** write_audit_findings_action.go Rule 4; finding_type field existence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### April 2026 pipeline triage fix set (P1–P5)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Triaged 57 build pipeline failures … wrote and deployed 7 code fixes (P1–P5, P9 across 12 files)" (April 2026)
- **what:** P1: component_id plumbed through create_work_item into site_work_items (unblocking tool-improver's load_tool). P2: idempotent tool fork deploy (reuse orphaned forks). P3: 429/rate-limit/billing errors classified transient in isAIUnavailable (items back to triaged without burning attempts; ~130 wasted attempts/day stopped). P4: load_page_record falls back page_name → page_id. P5: plan-then-reconcile for needs_section_data (auto-close stale requests when data arrives; create without duplicating when still missing) — "feedback loops need both directions".
- **sources:** HANDOFF-pipeline-triage-april-2026.md (whole)
- **relations:** section-data reconciler (P5's successor); two-strike rule
- **verify-later:** ai_errors.go isAIUnavailable patterns; plan_sections loadOpenSectionDataRequests/closeResolvedDataRequest

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component linking enrichment saga (component_id NULL on rebuilt pages)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Resolved on homepage. 7/7 linked" (2026-04-20) after fixes deployed 04-17/18; slot_name normalized differentiators-section → differentiators proving the data-component branch fires
- **what:** page_components.component_id was wiped on every rebuild because sections_metadata from the content writer carries only rendered_html — extractSectionsFromMetadata defaulted ComponentName to "section" and the enrichment guard skipped every row. Fixes: run enrichSectionsWithPlannedNames before enrichSectionsWithComponentIDs; prefer the HTML data-component attribute over metadata names; strip -section/-container/-wrapper/-block suffixes; log at Info with candidates_tried. Long-term structural fix (deferred): compile_page_sections should emit component_name per metadata entry. Stale pre-fix rows self-heal on next natural rebuild. Companion facts: `mode` has exactly one legal value "recreate"; build_mode is a dead parameter.
- **sources:** HANDOFF_2026-04-18_enrichment_bug_diagnosed_and_patched.md; HANDOFF_2026-04-19…md#1; HANDOFF_2026-04-20_component_linking_resolved_mode_rewrite_bug(2).md
- **relations:** data-component attribute contract (news-listing template fix); spec-is-primary-input contract
- **verify-later:** save_page_sections_action.go enrichment order; compile_page_sections metadata shape

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Improvement-sweep site starvation
- **category:** improvement-loop
- **status-signal:** unknown
- **status-evidence:** "Oldest updated_at site always wins; sites with frequent rebuilds dominate" — carried P3 across 04-17 → 04-20 handoffs, never picked up
- **what:** The improvement sweep's site selection starves some sites the same way find_dispatchable_site's arbitrary ordering does — scheduling fairness is an unowned concern across both loops.
- **sources:** HANDOFF_2026-04-17_triage_and_component_linking.md#known-issues; HANDOFF_2026-04-20…(2).md#5
- **relations:** dispatch chain fairness ORDER BY
- **verify-later:** improvement-sweep pre_query

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Scheme-coherence audit guard (Q8)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** PLAN Q8 "A scheme-coherence check in the design auditor / improvement loop … Status: open (I)"; absent from the SCHEME CLOSE remaining-work map and every later position note — silently dropped after the paired-variable close.
- **what:** The proposed regression guard: an auditor/improvement-loop check flagging "section scheme does not match site scheme / unintended contrast" so the scheme→components fix cannot silently regress. Designed as fix-shape item 8, never specified or built; the eventual regression protection took a different form (contract in the creator prompt + the re-aimed fixer as mechanical enforcer), leaving no dedicated audit check.
- **sources:** PLAN_scheme_to_components(1).md#Q8 #Provisional-fix-shape; RUNBOOK_scheme_to_components(50).md#SCHEME-CLOSE (absence)
- **relations:** fixes-at-initial-render principle; fix_forced_text_colours re-aim (partial substitute).
- **verify-later:** design-auditor checks list — does any scheme-coherence rule exist?

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Improvement-loop colour/nav fixer suite (pre-re-aim state)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Notes (Sj) full read of the fixers; HANDOFF: "As aimed today they are scheme-blind and ENFORCE the component-owns contract (`fix_forced_text_colors` injects dark `--section-*` into is_dark_section components)."
- **what:** The established fixer infrastructure: color-variable-fixer agent runs `fix_hardcoded_colors` (dark background hex → `var(--color-primary)`; dark 2-stop gradients → primary/secondary; deliberately leaves `rgba(0,0,0,x)` overlays alone; fixes both template and rendered HTML) and `fix_forced_text_colors` (strips forced child text colours so elements inherit the chain, WCAG-validates ≥4.5, and — the superseded part — injected the white `--section-*` contract into is_dark_section components); nav-link-fixer runs `fix_nav_link_templates` (with the documented rule that `render_site_components force_rerender` must follow); `fix_component_template` routes on fix_type (inject_nav_flex_css, remove_element, align_slot_name, inject_responsive_css, repair_template_slots) — symptom fixes for exactly the dark fallback header's output. Running them as-was on idea.uk would have entrenched dark.
- **sources:** running_notes_scheme_to_components(55).md#Sj; HANDOFF_scheme_to_components_for_claude_code(1).md#Established
- **relations:** superseded by fix_forced_text_colours re-aim; fix_hardcoded_colors retained (its hex→primary mapping stays coherent with the paintPaletteBand class).
- **verify-later:** fix_harcoded_colours_action.go, fix_nav_link_templates_action.go, fix_component_template_action.go current behaviour.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### fix_forced_text_colours re-aim: painting classifier + declaration rewriter
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** gobatch_04(2) delivers whole-function Edit G + new classifier code, with "RESOLVED (2026-07-06): … switched to `var(--color-primary-text, …)`"; RUNBOOK 07-06 night: "Slice 3 (the re-aimed fixer): deploying per your note — confirm it built"; supervised first run still pending.
- **what:** The backstop fixer rebuilt around the new contract: a `paintClass` classifier (paintAmbient/paintPair/paintInk/paintPaletteBand) derives what a template's own CSS paints from regexes over its style blocks — never from `is_dark_section` (the parameter is kept but deliberately ignored so call sites stay unchanged); `rewriteSectionDeclarationsInHTML` converts literal `--section-*` declarations to the class-appropriate references (pair text, hero ink, on-colour family, color-mix derivatives) and deletes declarations from ambient non-painters; the proven literal-stripping machinery and the WCAG contrast gate are retained; the old contract-injector trio (`ensureSectionContractInHTML`/`injectSectionContract`/`sectionContractRe`) is deleted. `result.contractAdded` is repurposed as a rewrite counter (name kept per the no-rename rule, meaning shift noted).
- **sources:** gobatch_04_fixer_reaim(2).md; running_notes_scheme_to_components(55).md#Ud #Ue #Ug; SPEC_scheme_to_components.md#W5
- **relations:** section painting contract (the thing it enforces); supervised fixer first-run protocol; is_dark_section demotion.
- **verify-later:** fix_forced_text_colours_action.go deployed body (classifier present, injector trio gone); first live run's details JSON.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Fixes-land-at-initial-render principle (loop fixers backstop only)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Notes (Sk) "User steer … Fix at INITIAL RENDER (library + composition/renderer). The improvement-loop fixers stay AVAILABLE as a backstop but must NOT be *required* for new builds"; SPEC W5: "The fixers remain improvement-loop backstops — never required for a correct first render."
- **what:** A governing platform principle set mid-thread and carried into every artefact: correctness must come from the library, composition, and renderer at first render; improvement-loop fixers are post-hoc safety nets dispatched on audit findings, never a required step for a new build. This ruled out "make the fixers scheme-aware" as the primary mechanism and shaped where each fix landed (templates and prompts, not loop passes).
- **sources:** running_notes_scheme_to_components(55).md#Sk; SPEC_scheme_to_components.md#W5; HANDOFF_scheme_to_components_for_claude_code(1).md#Established (USER DIRECTION)
- **relations:** section painting contract; component-creator prompt re-aim; scheme-coherence guard (abandoned alternative).
- **verify-later:** whether any build workflow invokes colour fixers as a required step.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Supervised fixer first-run protocol (disposable specimen site)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK STEP D (2026-07-06 night+): "watch the re-aimed fixer act on a real site once, under supervision, before it is ever allowed near the improvement loop or a second site" — not yet run at last dated evidence.
- **what:** Rollout protocol for a re-aimed automated fixer: confirm the deployed pod carries it; capture the specimen's before-state (dartsonline, a disposable freshly-built non-target site, site_id 5fe8785b); spawn manually via the 016b harness; judge the returned per-component `details` JSON (declarations rewritten per class, literals stripped, contrast-gate skips) rather than the render; re-read components and diff; only then decide hand-needles vs fixer for the library tail, and only then allow the improvement loop near it. Includes the deferred plan for a guarded 7-table cascade delete of a messed-up disposable site (written against schema when needed, never ad-hoc). The sizing read found dartsonline a thin specimen (all components literal-free; only two literal text hexes).
- **sources:** RUNBOOK_scheme_to_components(50).md#STEP-D; stepD_and_pages_reads.sql; running_notes_scheme_to_components(55).md#Uj #Uk
- **relations:** fix_forced_text_colours re-aim; hazard/band tail; debugging harness (016b).
- **verify-later:** whether Step D ran; dartsonline component state; the cascade-delete script's existence.

<!-- SOURCE: U05_content_quality_linking.md -->
### Recommendation-specialist finding routing (bug / recommendation / gap)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** PLAN doc header: "Status: Proposed — not yet implemented"; FOCUS_content_quality(2): "identity-advisor does NOT exist. sites.approval_mode does NOT exist … PROPOSED, not built."
- **what:** Auditor findings mix factually-broken bugs with subjective recommendations; the pipeline treated both as content_rewrite, causing false-positive rewrites and HITL flooding. Proposal: auditors emit finding_type (bug|recommendation|gap); bugs → content_rewrite, gaps → rebuild (P9 shipped that part), recommendations → per-category specialist agents (identity-advisor for contact/email, tone-shift, content-strategist, component-template-fixer for cta, brand-designer) each returning apply/dismiss/escalate; sites.approval_mode ('auto'|'review') as the per-site HITL escape hatch; a learning loop feeding dismissed findings back into auditor prompts. Reality-checked this unit: the specialists and approval_mode were never built, and component-template-fixer in fact PUNTS on CTAs (cta_improvement → needs_review) — the PLAN's claim it "already handles CTA fixes" was wrong.
- **sources:** package_module/output_contexts/PLAN_design-note-recommendation-specialists.md; FOCUS_content_quality(2).md#machinery; running_notes_16(1).md#part-2
- **relations:** content-quality-auditor/visual-design-auditor; internal-link-resolver (took over the CTA-fix role); HITL.
- **verify-later:** write_audit_findings_action.go routing; whether finding_type exists in auditor prompts; sites schema for approval_mode.

<!-- SOURCE: U05_content_quality_linking.md -->
### Discovery-agent domain scoping + observe-only check enablement
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** running_notes_17(21) 2026-06-14: enumeration of the three discovery agents + "findings land status='detected' which is NOT claimable".
- **what:** Discovery checks live in domain-scoped analyst agents (quality-discovery-agent=build, design-discovery-agent=design, completeness-discovery-agent=content) via run_discovery_checks whose config `checks` array is the sole enablement switch; checks self-register via init() but run only when named. Findings insert with status='detected' (unclaimable), so a check can run observe-only while improvement-sweep (the triager) stays disabled — auditing signal with zero auto-action. Per-finding Pipeline/HandlerAgent from the returned WorkItemSpec survive insertWorkItem verbatim.
- **sources:** running_notes_17(21).md#wiring-confirmed + #§7-gate-RESOLVED; RUNBOOK_linking_phantom_fixes(7).md#7a
- **relations:** check_phantom_internal_links; improvement-sweep gating; sectionless check S1.
- **verify-later:** run_discovery_checks_action.go; the three agents' checks arrays.

<!-- SOURCE: U05_content_quality_linking.md -->
### improvement-sweep pause + gated re-enable sequencing
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** FOCUS_internal_linking(1) Operational note: "improvement-sweep scheduled_task is disabled (enabled=f, last completed 2026-05-08), intentionally paused during core build".
- **what:** The improvement-loop's triage sweep is deliberately off during core build; the detect→fix loop depends on it. Re-enable sequencing settled: first have the phantom_internal_links check enabled observe-only AND both surface handlers in place, watch one cycle clean, then re-enable — so resuming clears findings rather than accumulating them. improvement-loop confirmed as the sweep orchestrator (spawns discovery agents + build-dispatch-loop + triage_detected_items).
- **sources:** FOCUS_internal_linking(1).md#operational-note; RUNBOOK_linking_phantom_fixes(7).md#7b; running_notes_17(21).md 2026-06-14
- **relations:** observe-only enablement; internal-link-resolver (prerequisite handler).
- **verify-later:** scheduled_tasks improvement-sweep enabled flag; first post-enable triage cycle.

<!-- SOURCE: U05_content_quality_linking.md -->
### checkEmptyPageSections — dormant, half-superseded detector
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** running_notes_15(12) Part 6 RESULT: "the empty-page detector is dormant code"; Part 9: "stale, never-enabled code partially superseded by EmptySectionsCheck".
- **what:** A sub-check inside the validate_component_standards wrapper targeting "deployed/active pages with zero rendered components" — wired in code but the wrapper is enabled in no discovery agent, bundled with six unrelated sub-checks, scoped to miss `planned` pages, and routes to page-content-writer (which cannot persist/deploy). Its sibling EmptySectionsCheck got the handler fix + enablement; this one didn't. Kept as the canonical example of dormant check debt and why a dedicated check was written instead.
- **sources:** running_notes_15(12).md#part-6, #part-9
- **relations:** check_sectionless_pages (its replacement for the planned-page case); EmptySectionsCheck.
- **verify-later:** ComponentStandardsCheck.Run; any agent enabling validate_component_standards.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Improvement-loop colour fixers are scheme-blind (re-aim as backstop, not fix)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** "Running them as-is on idea.uk would ENTRENCH dark, not lighten it" (running_notes Sj); user steer Sk: fix at initial render, fixers stay backstop; re-aim listed as pending Go-batch work.
- **what:** Existing fixer machinery — color-variable-fixer (fix_hardcoded_colors: dark hex→var(--color-primary), leaves rgba overlays; fix_forced_text_colors: strips forced child text colours, WCAG-validates, but INJECTS the white --section-* contract for is_dark_section components), nav-link-fixer (+ its documented render_site_components force_rerender follow-up), fix_component_template symptom fixes — de-hardcodes template + rendered HTML but enforces the OLD component-owns-dark contract keyed on is_dark_section, pulling opposite to the chosen architecture. Decision: initial-render fixes are primary; fixers get re-aimed to enforce the same paired-variable contract as backstop (key on what the template paints, never is_dark_section).
- **sources:** running_notes(22).md Sj, Sk; RUNBOOK_scheme_to_components(18).md WHERE WE ARE (fixer re-aim in Go batch)
- **relations:** hazard/band split; paired-variable direction; D2b (later analogous prevention-over-cure shape).
- **verify-later:** fix_forced_text_colors / fix_hardcoded_colors current keying; whether re-aim shipped.

<!-- SOURCE: U09_adoption.md -->
### Dormant discovery check: checkEmptyPageSections / validate_component_standards wrapper
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** "`checkEmptyPageSections` is stale, never-enabled code partially superseded by `EmptySectionsCheck` (which got the handler fix + enablement)… the wrapper `validate_component_standards` is in no agent's checks array (enables_vcs=f ×3)" (running_notes_15 Part 9).
- **what:** An existing sub-check targeting "page with no rendered sections" that never ran in production: its wrapper is not in any discovery agent's checks array, it is bundled with six unrelated sub-checks, scoped to deployed/active only (missing `planned` pages), and routes to page-content-writer (a task specialist with no persistence/deploy — cannot rebuild a page). Documented as the rationale for writing the dedicated sectionless check instead of extending it. General mechanism: `run_discovery_checks` runs only the names listed in the step-config `checks` array.
- **sources:** running_notes_15(10)#part-6, #part-9
- **relations:** check_sectionless_pages (its replacement for this case); EmptySectionsCheck (empty-HTML variant, enabled)
- **verify-later:** discovery_checks registry; completeness/design/quality-discovery-agent checks arrays

<!-- SOURCE: U09_adoption.md -->
### needs_section_data semantics and the abandoned standalone handler
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** "SUPERSEDED 2026-05-06 by FOCUS_directory_builder_and_list_components.md… needs_section_data items are emitted with status='needs_human_review' directly. They mean 'couldn't resolve component or required field — get a human to look.' Not async dispatch." (FUTURE_section_data_handler(1) header).
- **what:** The abandoned idea of a dedicated `needs_section_data` handler agent that would fetch list data asynchronously. Corrected understanding: those items are HITL by design; the list-data mechanism is query.* resolution. Residual real issue: plan_sections conflates two sub-cases — component missing (should emit needs_component to component-creator) vs required field genuinely unsourceable (real HITL) — and 41 items were stuck across six sites. `reconcile_section_data_action.go` exists for the deferred-query case but is not wired into any agent (and was a red herring for the guides-hub defect).
- **sources:** old2/FUTURE_section_data_handler(1).md, old/FUTURE_DEPRECATED_WRONG_section_data_handler.md, FOCUS_directory_builder_and_list_components.md#what-about-needs_section_data, running_notes_14(25)#part-14o–14p
- **relations:** directory-builder; plan_sections required-field switch; cta_url deferral bug
- **verify-later:** whether the 41 stuck items were triaged; reconcile_section_data wiring status

<!-- SOURCE: U10_imagery.md -->
### Corrupted component templates and the quality→regeneration bridge
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "✅ Bridge self-heal proven END-TO-END… no human involvement at any step. Fleet-remaining corrupted: 7" (2026-07-10); "10/14 healed" in the handoff.
- **what:** 14 components fleet-wide had html_template saved as RENDERED OUTPUT (literal `<no value>`, zero `{{…}}` vars) — historical damage from the pre-validation component-generation era (created_from='generated', 2026-03-31→04-13); the modern writer's pre-store validation already rejects this class. Detection existed (compute_component_quality flags "0 template variables") and repair existed (needs_component_regeneration → component-creator); the missing piece was a ~200-line bridge: `check_component_template_corrupted` discovery check (cross-site guard since components are fleet-shared, cap 5/pass). Field-preservation guard rejections are handled by re-queuing with exact field names in spec.description (rendered into the creator prompt).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turns-11–16/#Turn-20, SQL_2026-07-10_register_component_template_corrupted.sql, SHOWCASE_technical_architecture.md#5
- **relations:** tool-library/component-creator contract; the flagship "self-healing fleet" showcase story.
- **verify-later:** check_component_template_corrupted.go; remaining corrupted count fleet-wide.

<!-- SOURCE: U10_imagery.md -->
### mark_item_failed error honesty (flag-before-complete)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Fix applied (SQL_2026-07-10_pagebuild_mark_item_failed.sql… verified)… Failed page builds are now VISIBLE instead of silently complete."
- **what:** page-build-handler's step-level error routing pointed at `complete_error`, a SUCCESS-labelled complete_workflow — so a real step failure completed the orchestration and the dispatcher stamped the item 'complete' with no error (the "no-op complete" anomaly, triggered by a Kafka reply flake). The established flag-the-item-BEFORE-completing pattern was extended to real errors: a `mark_item_failed` step (update_work_item_status → 'failed', attempt-counted) inserted ahead of complete_error with all 8 error pointers repointed. Workflow-config-only. A fleet-trust principle: "a fleet you can trust starts with a fleet that tells the truth about itself."
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-16, SQL_2026-07-10_pagebuild_mark_item_failed.sql, SHOWCASE_imagery_workstream.md#4
- **relations:** Kafka partition race (the trigger); CompleteWorkItemAction guard semantics; likely needed on other handler workflows.
- **verify-later:** page-build-handler workflow error_step pointers; failed-vs-complete item stats post-fix.

<!-- SOURCE: U11_traffic_probe.md -->
### backend_unreachable discovery check
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** running notes 2026-06-13(f): "backend_unreachable REWRITTEN against the real DiscoveryCheck interface … gofmt-clean. Enable by adding … to the discovery agent's default_config checks array" — written and interface-reconciled, enablement not claimed.
- **what:** A discovery_checks/ check giving the improvement loop eyes on the VM class: per-site, NOOPs unless deploy_config.target='vm'; probes the public https://<domain>/health; on failure returns a WorkItemSpec (source='discovery', item_type='backend_unreachable', item_key dedup against idx_swi_dedup's partial unique index → one open alert per site, no spam). ALERT not auto-fix: handler_agent empty because a down VM isn't chassis-fixable — sits visible at 'detected'; the P5 vmhost adapter becomes the handler later. SELF-CLEARING: resolves its own open item on recovery using the runner's transaction. A companion `missing_beacon` check (rendered index lacks the /api/hit img) was floated and not built.
- **sources:** traffic_probe_running_notes(28).md#2026-06-13-e/f, traffic_probe_plan(12).md#P4, HANDOFF#cross-thread
- **relations:** VM-hosted backend sites class (first-class sites coverage), scheduler-and-tasks, vmhost adapter
- **verify-later:** discovery_checks/check_backend_unreachable.go in chassis; discovery agent checks array

<!-- SOURCE: U12_docs024_archives.md -->
### Colour-fix algorithmic detail (countHardcodedColorComponents / findForcedTextColors)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Full detail present in `009b_improvement_loop_v2.md`; deleted outright from `009c_improvement_loop_v3.md` onward and absent from live `004_improvement_loop.md` (table-row summary only).
- **what:** Documented exact algorithmic mechanics for two `design-discovery-agent` colour checks: `hardcoded_section_colors` and `forced_text_colors` (parses `<style>` blocks, flags only child text elements, skips container/link rules), with a WCAG AA 4.5:1 contrast safety check and `--section-*` contract injection. Pruned from the docs from v3 onward in favour of a one-line table entry.
- **sources:** old/older1/009b_improvement_loop_v2.md#"Colour Fix Detail"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** color-variable-fixer handler; contracts-and-standards CSS variable contract
- **verify-later:** `fix_hardcoded_colors`/`findForcedTextColors` Go source accuracy.

<!-- SOURCE: U12_docs024_archives.md -->
### Per-site, per-audit-type cadence configuration (maintenance_profile.audit.{type})
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** Appears identically in v2-v4 as a "## Configuration" section, each time caveated "future enhancement." Absent from v5 and live, which document a simpler global 60-day auto-reset with no per-audit-type knobs.
- **what:** Three consecutive doc versions carried a designed-but-never-built configuration surface: per-site JSON config letting each audit type be individually enabled/disabled with its own re-run interval. Quietly dropped rather than implemented.
- **sources:** old/older1/009b_improvement_loop_v2.md#"Configuration"; old/older1/009d_improvement_loop_v4.md#"Configuration"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** Audit Pass Cap / Auto-reset mechanism (its replacement)
- **verify-later:** check `sites.settings.maintenance_profile` rows for leftover `audit.{type}` keys.

<!-- SOURCE: U12_docs024_archives.md -->
### Acceptance-test cheap-LLM verification call gating lock + retry
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** Documented in `009c_improvement_loop_v3.md`/`009d_improvement_loop_v4.md`, incl. literal verification prompt. Live `004_improvement_loop.md` retains `acceptance_test` as a required field but documents no corresponding verification-call step.
- **what:** Each finding carried an `acceptance_test` enabling a cheap follow-up LLM call after a fix: feed fixed HTML back, get YES/NO, gating section lock (pass) or retry up to `max_fix_attempts` before escalating to `needs_human_review`. The field survived but the explicit verify-then-lock mechanism dropped out of documentation by v5/live.
- **sources:** old/older1/009c_improvement_loop_v3.md#"Structured Findings Format"; docs024_key_docs_latest/004_improvement_loop.md#"1. Finding Cap"
- **relations:** Section Locking; Finding Cap
- **verify-later:** search for a dedicated verification-call step (`verify_fix`/`check_acceptance_test`) in fixer code.

<!-- SOURCE: U12_docs024_archives.md -->
### Content-writer chrome double-injection bug and chrome-ownership rule
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** `009d` v4 "## Content Writer Chrome Fix (v4)": full bug narrative + cleanup; only a one-line changelog bullet survives into v5/live.
- **what:** A production bug rule: site chrome (header/footer/head) must be injected exactly once, only at the rerender/assembly step — never by the content writer. Fix set all three inject flags false on `page-content-writer`'s `compile_page` step plus a cleanup pass removing baked-in chrome components.
- **sources:** old/older1/009d_improvement_loop_v4.md#"Content Writer Chrome Fix (v4)"; docs024_key_docs_latest/004_improvement_loop.md (changelog line only)
- **relations:** contracts-and-standards (component/slot contract); site-component-linker
- **verify-later:** confirm `page-content-writer` inject flags remain false; check for reappearance of baked-in header/footer components.

<!-- SOURCE: U12_docs024_archives.md -->
### Audit finding dedup + blocked-item filtering algorithm (write_audit_findings)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** Full three-step algorithm documented in v3/v4; not present in v5/live (mentioned only in passing, then dropped even from the summary line).
- **what:** `write_audit_findings` was documented as implementing three dedup/safety layers: bulk-preloading blocked item keys, a broader item_type+page match against existing blocked items, and item-key-based dedup against pending items. This mechanism-level detail disappears from the documentation surface after v4.
- **sources:** old/older1/009d_improvement_loop_v4.md#"Finding Dedup and Blocked Item Filtering"; docs024_key_docs_latest/004_improvement_loop.md
- **relations:** Finding Cap; Triage Drain Controls
- **verify-later:** confirm `write_audit_findings` still implements bulk-preload + item_key pattern.

<!-- SOURCE: U12_docs024_archives.md -->
### Triage drain loop — structured audit findings, capped passes, section locking
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Live `009_model_infrastructure.md` "Done": structured findings, audit pass cap, section locking exclusion all checked off.
- **what:** Fix for unbounded audit/fix/re-audit token spend. Findings must carry `acceptance_test`/`acceptance_levels`/`minimum_required`. Audits capped at 3 numbered batches per site. Passing sections get `locked_at`; subsequent audits skip them; unlock is manual. Per-page sequential processing via `depends_on` prevents overlapping fixes.
- **sources:** old/older1/020d_gpu_and_model_infrastructure_v4.md#"Triage Drain Loop Fix"; docs024_key_docs_latest/009_model_infrastructure.md#"Decisions Made"
- **relations:** three-way audit-finding classification; GPU/AI-endpoint scheduling
- **verify-later:** `write_audit_findings_action.go`; section-lock column on `page_components`.

<!-- SOURCE: U12_docs024_archives.md -->
### wont_fix/superseded dedup and needs_section_data data-honesty pattern
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Described as "correct behaviour... the dedup system working" in the second draft.
- **what:** When a recurring issue is detected while an older item is stuck, the loop creates a new item and marks the old one `wont_fix` ("superseded by active duplicate") — expected noise, not a bug. `needs_section_data` items requiring unfabricatable data (bios, pricing, case studies) correctly route to `wont_fix`/HITL rather than inventing content.
- **sources:** old/older1/016_debugging_guide_v2_april26.md#"9. Specific Failure Patterns"
- **relations:** debugging playbook; needs_section_data triage
- **verify-later:** current dedup logic; HITL routing for `needs_section_data`.

<!-- SOURCE: U12_docs024_archives.md -->
### plan_sections pre-check → plan-then-reconcile evolution
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** v3 report P5 "DEPLOYED" (pre-check); live v4 shows the same P5 row "UPDATED" with a materially different mechanism.
- **what:** The original fix for wasteful LLM re-sends on sections with pending `needs_section_data` was a pre-check that simply skipped them. Revised to "plan-then-reconcile": ready sections auto-close stale data requests, deferred sections create new requests while skipping duplicates.
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P5)"; docs024_key_docs_latest/105_dispatch-pipeline-failures-report_v4.md#"Priority Fixes (P5)"
- **relations:** early pipeline-failure triage priorities; needs_section_data triage
- **verify-later:** current `plan_sections_action.go` logic for auto-closing `needs_section_data` items.

<!-- SOURCE: U12_docs024_archives.md -->
### Three-way audit-finding classification (bug / recommendation / gap)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** v3 report's P10 marked "DEFERRED... ~1 week project"; referenced independently, still not built, in `027_design_and_site_planner_v2.md` months later.
- **what:** Auditors currently produce findings the pipeline auto-fixes uniformly as if bugs, but many are opinions/recommendations — producing false-positive fix attempts. Proposed fix: three-way classification with dedicated specialist agents per category and per-site approval mode for recommendations.
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P10)"; docs024_key_docs_latest/027_design_and_site_planner_v2.md#"10. Open Design Areas"
- **relations:** audit gap-finding routing fix; triage drain loop
- **verify-later:** existence/status of `design-note-recommendation-specialists.md` or any implementing specialist agent.

<!-- SOURCE: U12_docs024_archives.md -->
### Audit gap-finding routing fix (existing-page gaps → needs_content_page)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** v3 report P9: "FIX WRITTEN — write_audit_findings_action.go: Rule 4 routes gap findings on existing pages to needs_content_page, not content_rewrite."
- **what:** Gap findings on existing pages were being routed to `content_rewrite` (edits, not rebuilds), causing validation-failed rewrites. Rule 4 redirects them to `needs_content_page` (full rebuild path).
- **sources:** old/older1/105_dispatch-pipeline-failures-report_v3.md#"Priority Fixes (P9)"
- **relations:** three-way audit-finding classification; needs_content_page work-item type
- **verify-later:** current `write_audit_findings_action.go` Rule 4 logic.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Automation ratchet (per-capability trust levels)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §2 "Automation is a ratchet, not a switch"; a capability graduates `confirm-every → confirm-exceptions → notify → autonomous` "when evidence shows it reliable"
- **what:** Automation is not global; each capability (create action, provision nginx, reshard DB) carries its own trust level and graduates only on evidence. "Fully automated" is the union of individually-graduated capabilities. A trust ledger records each capability's level, gate policy, and supporting evidence.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#2, ED/MASTER_autonomous_build_and_operate(4).md#8.1
- **relations:** trust ledger; reliability cascade; bidirectional ratchet
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Reliability cascade (reuse → generate+verify → compete+judge → HITL)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §3 "Applied to every sub-task, in descending order of reliability"
- **what:** A per-task router for producing any unit of work in descending reliability order: known-good reuse, then generate+deterministic-verify, then compete-N-and-judge in a sandbox, then HITL. A verified recurring generated solution becomes candidate known-good and graduates its gate — feedback into the ratchet.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#3, ED/MASTER_autonomous_build_and_operate(4).md#7.2
- **relations:** cascade router; known-good library; multi-author generation
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Trust ledger + gate-policy engine
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7.2 "The trust ledger is the master knob"; §8.2 "Stand up the ledger (a table)"
- **what:** A per-capability store of automation level, gate policy, and supporting evidence, plus a small evaluator mapping (capability, trust level, stakes) → gate. It is the master knob: it governs both how conservatively a thing is produced and whether the result applies without a human.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#8.2
- **relations:** automation ratchet; cascade router; governance decision package
- **verify-later:** trust-ledger table (proposed)

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Cascade router
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §6.3 "Cascade router — the per-task decision (reuse / generate / compete / HITL)"; §8.4 "the loop's least-bounded step"
- **what:** An action/agent that picks a cascade tier per leaf task from three inputs — the capability's verifiability/containment, its trust-ledger entry, and the task's stakes. Named as the loop's least-bounded step, so conservative-by-default and ledger-gated.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.2, ED/MASTER_autonomous_build_and_operate(4).md#8.4
- **relations:** reliability cascade; trust ledger; verification harness
- **verify-later:** none

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Bidirectional ratchet (trust can be lost)
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §7.5 "Trust can be lost, not only gained … the safety property that lets the ratchet advance at all"
- **what:** Feedback is two-directional: success accrues evidence toward graduation; repeated/severe failure drops the trust level, tightens the gate and raises the cascade floor. A regressing capability is automatically pulled back under supervision.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** automation ratchet; trust ledger
- **verify-later:** none

<!-- SOURCE: U18_sql_for_agents.md -->
### Discovery agents (design / quality / completeness) and the check registry
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 047/048 define them; 074 expands completeness checks; 142/146 still adding checks to design-discovery in 2026-07 ("run_discovery_checks warns... and skips unknown names" — the safe-rollout pattern).
- **what:** Read-only detectors that "find problems. They do not fix anything. They do not call other agents." Each runs run_discovery_checks with a named check list, writing findings to site_work_items (source='discovery', status='detected'). design: undeployed_assets, missing_css, duplicate_palette, missing_tools, tool_health, tool_acceptance, tool_acceptance_due. quality: broken_nav_links, placeholder_contact, generic_theme. completeness: empty_sections plus integrity checks — cross_site_contamination, unrendered_templates, missing_style_collection, deactivated_site_components. All algorithmic, no LLM budget. Unknown check names warn-and-skip, so SQL can enable a check before the Go ships.
- **sources:** 047_discovery_checks.sql; 048_discovery_agents.sql; 058_quality_checks_and_fixers.sql; 074_completeness_discovery_agent.sql; 142_enable_tool_acceptance_check.sql; 146_enable_tool_acceptance_due.sql
- **relations:** improvement-loop orchestrates them; fixer agents consume their items; check registry in discovery_checks.go
- **verify-later:** registered checks in run_discovery_checks_action.go / discovery_checks/*.go

<!-- SOURCE: U18_sql_for_agents.md -->
### improvement-loop (post-build discovery → triage → fix → rerender cycle)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 054 definition; 086 Part B adds audit_pass_count guard "stops after 3 passes"; 100 portfolio claims sites "receive autonomous content audits... on rolling schedules".
- **what:** Runs after initial build (or on schedule/manual trigger): spawns the three discovery agents, triage_detected_items promotes detected → triaged, and if anything was promoted inserts needs_rerender at priority 99 and fires build-dispatch-loop to process all fixes then rerender. 086's audit-pass cap plus section locking provide the loop's termination condition ("the triage drain").
- **sources:** 054_improvement_loop.sql; 086_visual_design_auditor.sql; 061_tool_deployer_and_discovery_agent.sql (flow diagram)
- **relations:** discovery agents, fixers, audit agents, locks
- **verify-later:** improvement-sweep scheduled task; triage_detected_items action; audit_pass_count in sites.settings

<!-- SOURCE: U18_sql_for_agents.md -->
### Fixer agents: color-variable-fixer, site-component-linker, component-template-fixer, css-patch-agent
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 056/064/076 definitions; all in 075's idle-timeout list; component-template-fixer gains note-writing in 132.
- **what:** Narrow algorithmic/LLM fixers dispatched from the queue: color-variable-fixer replaces hardcoded hex in component inline styles with CSS variables (fixes both templates permanently and rendered_html immediately); site-component-linker fixes NULL component_id causing fallback rendering; component-template-fixer applies targeted template surgery (nav flex CSS injection, element removal, slot_name alignment) routed on spec.fix_type; css-patch-agent LLM-patches the current stylesheet for spacing/responsive/layout issues without full regeneration (explicitly NOT theme redesign — that's webdesign-agent). All create deduplicated needs_rerender items only when they changed something.
- **sources:** 056_colour_variable_fixer.sql; 064_site_component_linker_and_fixer.sql; 076_css_patch_agent.sql; 058_quality_checks_and_fixers.sql
- **relations:** discovery checks (hardcoded_section_colors, unlinked components); rerender pipeline; audit findings
- **verify-later:** fix_hardcoded_colors, link_site_components, fix_component_template actions

<!-- SOURCE: U18_sql_for_agents.md -->
### Audit agent hierarchy (visual-design-auditor, content-quality-auditor, design-audit-agent, site-review-agent)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** 066 defines the hierarchy; 084 patches prompts due to a real cost incident ("845 design-audit work items across 4 domains in ~10 days... cost explosion"); 086 excludes locked components from audit queries.
- **what:** LLM auditors layered above discovery: pattern is "algorithmic checks first, then ONE LLM call for subjective assessment, then write findings" (write_audit_findings). 084 makes findings structured and bounded: TOP 5 only, every finding must carry current_value, a concrete `acceptance_test` "that a DIFFERENT agent could verify without re-auditing", max_fix_attempts, and must skip what algorithmic checks already caught. site-review-agent adds strategic alignment review; unclassifiable gaps become needs_content_planning items for content-gap-planner.
- **sources:** 066_audit_agent_definitions.sql; 084_site_review_agents.sql; 086_visual_design_auditor.sql; 071_content_gap_planner.sql
- **relations:** locks (locked_at exclusion); improvement-loop pass cap; fixers consume findings
- **verify-later:** write_audit_findings action; current audit prompts vs 084 text

<!-- SOURCE: U19_sql_tables_components.md -->
### Improvement-sweep and build-pipeline-trigger scheduling
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** Seeded tasks with evolving pre_queries: queue-size gate (skip when >20 open build items), round-robin site selection by least-recently-checked, skip sites with claimed items or locks.
- **what:** The improvement loop's cadence lives in scheduled_tasks: build-pipeline-trigger (2 min) finds sites with triaged/approved items and fires the dispatch loop; improvement-sweep (10 min) picks the next site for discovery checks, gated so discovery never floods an already-backed-up queue and locked sites are skipped. Both share the 'dispatch' concurrency group.
- **sources:** docs/agent_docs/sql_for_tables/020_scheduled_tasks.sql#seed-data and #improvement-sweep-fixes
- **relations:** scheduler; work queue; site-level lock.
- **verify-later:** current pre_query for improvement-sweep; discovery agent set.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Heartbeat maintenance model (findings-based, pre-work-items)
- **category:** improvement-loop
- **status-signal:** superseded
- **status-evidence:** v1 (docs017/021) findings table + domain orchestrators; v2 (docs017/022) full spawn chain modeled on "the vet-batch pattern" with budget management; explicitly replaced by 023: "maintenance-triage + page-rebuild → maintenance-batch-scheduler + site-work-orchestrator".
- **what:** The first full maintenance architecture: K8s CronJob (8h) → agent-chassis spawns maintenance-batch-scheduler → claims batch (FOR UPDATE SKIP LOCKED, batch_size controls concurrency) → per-site site-maintenance-orchestrator runs fix-pending → verify-previous → discover-due → triage cycle; discovery agents per domain (content/links/seo/compliance/structural) write maintenance_findings; triage (a step, not an agent) enriches with impact reads and classifies resolution path (auto_fix/suggest/flag/monitor/ignore); narrow fix agents resolve; cross-domain coordination only via side-effect findings with parent_finding_id — "no agent calls another agent for coordination." Daily maintenance-catch-all handles stale findings, HITL reminders, cross-site patterns, stuck recovery.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/022_maintenance_architecture_plan_v2.md; docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#8-Maintenance-Agent-Family; docs017_legacy_agent_rules_images_design_keydocs/021_maintenance_architecture_plan_v1.md
- **relations:** vet-batch-processor precedent (vet-med-pricing); unified work items (successor); scheduler-and-tasks; maintenance profile.
- **verify-later:** maintenance_findings/maintenance_tasks tables; maintenance-batch-scheduler agent history.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Unified build & maintenance work items (site_work_items)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** docs017/023: "Build and maintenance are the same process. A new site is a set of findings that need fixing"; full site_work_items DDL with item_key dedup, depends_on, parent_item_id, batch claiming; docs017/044 traces the working site-work-orchestrator step by step.
- **what:** The pivotal unification: every piece of work — building a page, fixing stale content, adding a tool, publishing an article — is one work item with source (planner/discovery/content_feed/manual/improvement/side_effect), domain, item_type, severity, spec JSONB, triage enrichment (impact, resolution_path, priority, handler_agent), lifecycle statuses detected→triaged→approved→claimed→in_progress→complete→pending_verify→verified, dependencies, dedup keys, attempt limits, and archival. The planner becomes a discovery agent writing 'needs_content_page' items; the same orchestrator/fix agents process build and maintenance; sites start minimal and improve incrementally, always left in a working state; per-page git commits. Old and new systems coexist (v2 intake routes between pageflow-builder and site-work-orchestrator). This is the direct ancestor of the current work-item lifecycle and improvement loop.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** heartbeat model (predecessor); maintenance_queue/page-rebuild (earlier still); work-item lifecycle in development-guide (current form); news feed → work items.
- **verify-later:** site_work_items vs current work_items table naming/shape; site-work-orchestrator vs current orchestrators.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Per-site maintenance profile with budgets
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** docs017/019b maintenance_profile JSON in sites.settings ("content every 7d, links every 8h... budget: llm_calls_per_cycle: 20, max_auto_fixes_per_cycle: 5"); 023 extends with content_feed cadence and time_sensitivity.
- **what:** Each site declares which maintenance domains run, at what cadence, with which sub-agents and regulatory bodies, plus hard budgets on LLM calls and auto-fixes per cycle — a finance site gets hourly high-sensitivity feeds and FCA compliance, a brochure site gets links+freshness only. Ancestor of the growth budget concept.
- **sources:** docs017_legacy_agent_rules_images_design_keydocs/019b_agent_architecture_v5_with_tickets_news.md#Per-Site-Configuration; docs017_legacy_agent_rules_images_design_keydocs/023_maintenance_architecture_unified_v6.md#Per-Site-Configuration
- **relations:** content-governance growth budget (descendant); scheduler cadence.
- **verify-later:** maintenance_profile key in sites.settings rows.

<!-- SOURCE: U22_recent_small_docs.md -->
### maintenance-triage agent
- **category:** improvement-loop
- **status-signal:** aspirational
- **status-evidence:** Defined with dry_run mode; workflow scans sites, queues page_rebuild tasks, spawns page-rebuild agent; described alongside "for future use" queue.
- **what:** An orchestrator that scans deployed sites for maintenance issues (stale pages, missing pages, broken links, CSS drift), inserts tasks into `maintenance_queue`, then dispatches specialist agents (page-rebuild) per affected site. Supports dry_run (scan+queue without dispatch) and a configurable stale_threshold_days.
- **sources:** docs019_business/017_maintenance_triage_agent.sql
- **relations:** maintenance_queue, page-rebuild agent, improvement-loop
- **verify-later:** agent_definitions type='maintenance-triage'; actions scan_sites_for_maintenance/prepare_rebuild_dispatches

<!-- SOURCE: U22_recent_small_docs.md -->
### build-dispatch-loop self-chaining removal
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Fix: Loop back to load_next_item internally ... Status: Applied to production DB. Verified live definition matches migration."
- **what:** A fix (migration 063) removing the build-dispatch-loop's self-respawn pattern (spawn_next_dispatch → call_next_dispatch), which repeatedly left the parent stuck in AWAITING_RESPONSES when the child's Kafka response was lost to topic retention/pod restarts. Now loops back to load_next_item internally (9 steps vs 13), timeout bumped 900→1800s.
- **sources:** docs021.../024_handoff_summary_2026_03_02.md#fixes-applied
- **relations:** work-item lifecycle, dispatch loop, orchestration timeouts
- **verify-later:** build-dispatch-loop agent definition step count; migration 063

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Triage drain loop fix (bounded audit passes + section locking)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** 020c "Triage Drain Loop Fix": 845+ design-audit items across 4 domains in ~10 days, "The loop has no termination condition"; Fixes 1-5 (structured findings with acceptance_test, max 3 passes, section locking via `page_components.locked_at`, verify-against-criteria not re-audit, per-page `depends_on` sequencing); "~65-70%" token reduction.
- **what:** The audit→fix→re-audit loop ran unbounded, consuming most tokens. Fix: auditors emit structured findings with `acceptance_test`/`acceptance_levels`; cap at 3 audit passes per site producing numbered batches; lock passing sections (`locked_at`) so later audits skip them (unlock always manual); verify via a cheap acceptance-test call (not full re-audit); sequence same-page items via `depends_on`. This is the origin of the improvement-loop guardrails referenced across imagery/content docs.
- **sources:** old/older1/020c_gpu_and_model_infrastructure_v3.md#triage-drain-loop-fix
- **relations:** section locking (page_components.locked_at); imagery audit loop 3-pass cap; recommendation specialists
- **verify-later:** audit agent acceptance_test emission; sites.settings audit pass cap

<!-- SOURCE: U24b_docs_archive_finetuning.md -->
### Flywheel B — RAG knowledge base with nomic task prefixes
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** FOCUS(21) §2.4b "Step 3 (chassis integration) passed on 2026-04-21 … Flywheel B is done"; "Prefix patch deployed and verified live 2026-04-21"
- **what:** `knowledge_base` pgvector(768) table read/written by `rag_lookup`/`rag_index` actions on the cpu-ollama nomic-embed-text endpoint, with trigram fallback. Empirically established that nomic `search_document:`/`search_query:` task prefixes are load-bearing for ranking (French Bulldog BOAS test), now patched into production. Best practice: filter by metadata (vertical, component_type, source) first, then rank by similarity.
- **sources:** flywheel_docs/FOCUS_finetuning_flywheel_and_service(21).md#2.2, #2.4b; #14 (Ollama specifics)
- **relations:** short-term lever paired with LoRA (long-term); the flagship of the finetuning.uk RAG product
- **verify-later:** platform/orchestration/actions/rag_actions.go (applyNomicPrefix); knowledge_base table; PATCH_rag_actions_nomic_prefixes

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Defect-cataloguing discipline (enumerate-before-fixing)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** `CATALOGUE_gamesdesign_post_sync_fix_defects.md` states its purpose explicitly: "Enumerate every observed defect as a *separate* item before fixing, so distinct causes are not conflated into one rolling investigation," with causes marked "tentative" until confirmed by reading source. Later revisions (`(4)`) show the discipline paying off: defects graduate from `[NEW]`/`[PARKED]` through `[FIX SHIPPED — PARTIALLY VERIFIED]` to `[VERIFIED CLOSED]` with a pinned, source-read cause replacing the original tentative one.
- **what:** A working method for a real adoption-run defect sweep: group symptoms into lettered families by shared mechanism (A deployment gaps, B link fallbacks, C list-component content, D section-data gaps, E content quality, F guide duplication, G design fidelity, H hygiene, I open unknowns, J dispatch throughput), triage by root cause not symptom, and forbid shipping a fix from a "tentative" cause without first reading the responsible action.
- **sources:** adoption/CATALOGUE_gamesdesign_post_sync_fix_defects.md; adoption/CATALOGUE_gamesdesign_post_sync_fix_defects(4).md
- **relations:** running-notes debugging-log convention (below), silent-completion family
- **verify-later:** whether this catalogue format was formalised anywhere beyond this one adoption run.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Dormant discovery-check machinery (`checkEmptyPageSections` / `validate_component_standards`)
- **category:** improvement-loop
- **status-signal:** abandoned
- **status-evidence:** running_notes_15(5) Part 6: "**RESULT (decisive):** ... `validate_component_standards` (its wrapper) is **not enabled in any** discovery agent (`enables_vcs=f` for all three)... The empty-page detector is **dormant code**, not a buggy check."
- **what:** A pre-existing check (`checkEmptyPageSections`, inside `ComponentStandardsCheck`/`validate_component_standards`) already targets exactly "page with no rendered sections," but was never added to any discovery agent's `checks` config array, so it has literally never fired in production — its 11 historical `needs_content_page` items were all traced to adoption-run/manual sources, none to this check. It was also found to be scoped too narrowly (`deployed`/`active` only, missing `planned`) and to recover by re-emitting a still-empty spec (would loop, not repair) — reasons a *new* dedicated check (`check_sectionless_pages.go`) was written instead of extending the dormant one.
- **sources:** adoption/running_notes_15_skinner_box_and_adoption_sections(5).md Parts 6–8
- **relations:** sectionless-page silent completion (above)
- **verify-later:** `discovery_checks/` registry current contents; whether `validate_component_standards` has since been enabled anywhere.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### `improvement-sweep` scheduled task — deliberately disabled pending consumer readiness
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** running_notes_17(16): "Operational: `improvement-sweep` scheduled_task is **disabled** (`enabled=f`, last completed 2026-05-08), intentionally paused during core build... Before re-enabling: have the `phantom_internal_links` check enabled AND both handler agents in place (`nav-link-fixer` exists; `internal-link-resolver` is Step 3), so resuming the sweep clears findings rather than accumulating them." Later resolved to a specific enablement gate (§7) confirming per-finding routing survives `run_discovery_checks_action.go`'s pipeline stamping, so the check could finally be enabled "observe-only" without turning the sweep back on.
- **what:** The discover→triage→fix improvement loop's top-level scheduler is deliberately kept off while core build work is in flight, on the explicit policy that a discovery check should only be enabled once its handler agent actually exists — otherwise findings accumulate unconsumed rather than clearing. This is a recorded operational policy (not a bug) governing when automation is safe to turn back on.
- **sources:** content_quality_and_internal_linking/running_notes_17_internal_linking_phantom_fixes(16).md, "Policy settled" and "§7" sections
- **relations:** dormant discovery-check machinery; internal-link-resolver agent
- **verify-later:** current `enabled` state of `improvement-sweep` in `scheduled_tasks`.

<!-- SOURCE: U25_leopardess_social.md -->
### `needs_rebuild` is inert without an explicit work item
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK O3: "pages.build_status='needs_rebuild' does nothing on its own … This is why six pages have sat at needs_rebuild doing nothing" (2026-07-10).
- **what:** The build-dispatch-loop reads site_work_items and never scans pages; only write_build_items converts needs_rebuild to work items and it lives inside site-work-orchestrator/build-site-planner, not the loop. Operator remedy: INSERT site_work_items rows (needs_content_page → page-build-handler) explicitly. Related dispatch facts: claim_work_item blocks items whose handler_agent doesn't exist; unhealthy AI endpoints release items; the partial unique index idx_swi_dedup silently suppresses new items sharing (site_id, item_key) with an open one.
- **sources:** docs/leopardessconsulting/RUNBOOK.md#O3, #O4, #landmines-1/9; docs/leopardessconsulting/HANDOFF.md#4.1
- **relations:** silent no-op success class; work-item dedup semantics
- **verify-later:** build-dispatch-loop workflow; write_build_items call sites

<!-- SOURCE: U25_leopardess_social.md -->
### build_status 'approved' invisibility defect and its layered fleet fix
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_minilobby §0: writer fix "CLOSED — fixed in chassis v1.0.1102, verified end-to-end 2026-07-10"; CHECK constraint "CLOSED 2026-07-11 — migration 049 applied + negative-tested"; drift check shipped v1.0.1104.
- **what:** apply_section_edit left an edited live section at build_status='approved' while every discovery check filters ='deployed' — a live section silently invisible to the whole audit surface (same shape as complete_error). Fixed at four layers: writer (UpdatePageStatusAction gained page_component_id_field, mirroring the deploy mark onto the named page_component; coordinator dataRefKeys registration), invariant (migration 049 CHECK constraint on the previously free-text column — invented statuses now fail loudly), detection (new check page_component_status_drift: unknown status → emit; pending-on-deployed-page → finding only, since a status flip would hide real staleness — 19 such rows surfaced across 5 sites), repair (new fixer repair_page_component_status with refusal guards). Leopardess RUNBOOK carries the interim two-statement manual repair as landmine 3.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#8.0, #8.1; docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#3-#4; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10
- **relations:** section-editor; fleet generalisation doctrine; discovery check wiring
- **verify-later:** v3_site_actions.go UpdatePageStatusAction; migration 049; check_page_component_status_drift.go

<!-- SOURCE: U25_leopardess_social.md -->
### Runtime-fill guards in discovery checks (three defused landmines)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-10 (night): v1.0.1105 "third guard PROVEN … the spawned pod logged all four exemptions by name"; guard split 3 emit / 2 skip verified live.
- **what:** Three discovery checks would have dismantled vonc's runtime-fill shells the moment the improvement loop switched on: component_template_corrupted (would regenerate shells into build-time copy), broken_template_slots (endless churn of version rows), and empty_sections (already enabled — caught live raising full LLM rebuilds of the shells within minutes of the first pass). Guard pattern: key on the data-runtime-fill marker (author's declared intent, never component names), exclude from work-item emission but record a Findings entry — a bare SQL NOT LIKE was rejected because silent skipping is the codebase's recurring failure mode. For runtime-fill shells `<no value>` IS the mechanism, not the defect.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 entries; docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#3 (#3, #14)
- **relations:** runtime-fill mechanism; Mode-B templates; fleet generalisation doctrine
- **verify-later:** check_component_template_corrupted.go, check_component_standards.go, check_empty_sections.go guards in the running binary

<!-- SOURCE: U25_leopardess_social.md -->
### Discovery-check wiring gaps (registered-not-enabled / enabled-not-implemented / sweep off)
- **category:** improvement-loop
- **status-signal:** partial
- **status-evidence:** PLAN_generalise §3 #6/#7: "8 checks registered in Go, enabled in no agent — incl. sectionless_pages, the exact detector for the ten silent complete_error builds"; "3 enabled check names have no Go implementation"; improvement-sweep "disabled since 2026-05-02, so discovery runs only by manual trigger".
- **what:** The improvement loop's configuration surface has three drift modes: checks registered in Go but named in no agent's checks array (inert capability); check names enabled in agents with no implementation (runner warns "Unknown discovery check"); and the scheduler that drives the whole loop disabled for two months, making discovery manual-only — which is why design-discovery-agent had never run on vonc and undeployed_assets never fired. Enable decisions are deliberate per-agent config edits (three checks added to completeness-discovery-agent 2026-07-10, backed up first); the remaining candidates are listed survey-first.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#5; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#4; docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-12 (first design-discovery run on vonc ever)
- **relations:** silent no-op success class; runtime-fill guards; scheduler-and-tasks (improvement-sweep)
- **verify-later:** agent_definitions checks arrays vs Go-registered Name() set; scheduled_tasks improvement-sweep row

<!-- SOURCE: U25_leopardess_social.md -->
### Fleet generalisation doctrine (four rules + artifact verification)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** PLAN_generalise §2 (2026-07-10), applied to all 14 findings; "of thirteen findings exactly one is site-specific".
- **what:** The doctrine for turning incident fixes into fleet guarantees: (1) fix the writer, not the row — a psql repair scoped to a site_id is evidence of a bug, never its fix; (2) detect by contract, not by name — guard on declared markers, never component names; (3) surface, never silently skip — where a check must not act it emits a Finding; (4) every detection needs a fixer, or a written reason there is none. Fifth inherited rule: verify by artifact, never item status. The mini-lobby trim is the worked example: "a keyhole onto fleet-wide defects".
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/PLAN_generalise_fixes_to_fleet.md#2, #3
- **relations:** operator discipline; build_status fix layers; runtime-fill guards
- **verify-later:** n/a (doctrine)

<!-- SOURCE: U25_leopardess_social.md -->
### Work-item dedup and two-strike semantics (partial index behaviour)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** RUNNING_NOTES_minilobby 2026-07-10: "idx_swi_dedup is partial — it only covers non-terminal rows, and the two-strike rule only counts complete/failed. So a completeness pass … will re-raise the two rejected shell items."
- **what:** Dedup semantics that shape operations: the partial unique index suppresses a new work item only while an open one shares (site_id, item_key) — rejected/terminal items are outside it, so a re-run re-raises them (which made a rejected item's *absence* after the guarded build positive proof the guard worked). Leopardess side: the same index silently suppresses intended new items when an open twin exists. Co-page duplicate rule: a page rebuild is whole-page, so multiple empty_section items on one page are duplicates — close the second by artifact.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10 (evening/night); docs/leopardessconsulting/RUNBOOK.md#landmine-9; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 (dispatch note)
- **relations:** needs_rebuild inert; discovery wiring; stuck-claim dispatch noise
- **verify-later:** idx_swi_dedup definition; two-strike logic in dispatch

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Interactive-page de-tool hazard (content rebuild silently drops a tool/game)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** CONFIRMED 2026-06-22 (gamesdesign game-pathfinding, 18KB A* canvas overwritten 06-14); "fix pending" in 002/005/016; 016b v2 update: two-layer save_page_sections fix WRITTEN, un-deployed
- **what:** A tool lives as a section's rendered_html, not a planned section, so any full rebuild (needs_page or link_resolution_rebuild) regenerates from plan_sections and replaces it with generic-text-block; the prose-based content-regression guard doesn't catch markup/JS loss. Fix layers: interactivity-aware save guard + carry-forward of interactive sections in save_page_sections (written), source_item_id stamping into page_component_history, and routing link maintenance through a preserve-sections path (page_rerender ruled out for CTA re-resolution — it doesn't re-run link logic).
- **sources:** 005(1) hazard block; 002(4)#Interactive-page hazard; 016 §9 final entry; 016b Part 4
- **relations:** phantom-CTA resolution bug (separate); tool recreation mis-key (Part 3)
- **verify-later:** save_page_sections_action.go patched version deployed?; page_component_history.source_item_id population

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool doc header (sentinel comment; stripped at deploy)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 019 full lifecycle table incl. StripToolDocHeader call sites and tool_health checks
- **what:** Every new tool's script opens with one sentinel-delimited block (function/purpose/behaviour/inputs/outputs; never ids/dates; no */). It never ships: StripToolDocHeader runs at the three outbound assembly points; DB rendered_html retains it for audit parity. Creation gate validates presence; improver preserves/updates it; auditor audits code AGAINST its stated behaviour; malformed (opener without closer) is left in and flagged by tool_health.
- **sources:** 019#Tool Doc Header; 020 tool_health tier-1 checks
- **relations:** per-tool travelling docs (037); tool-auditor
- **verify-later:** platform/content/tool_doc_header.go; prompt migration applied

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool quality three tiers (structural / LLM audit / headless-browser future)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 020: tiers 1–2 automated live, tier 3 "planned"
- **what:** tool_health tier 1 (Go, free): deploy status, HTML/template present, script/style/@media, hex/external-dep warnings, doc-header checks — blockers create improve_tool. Tier 2: audit_tool queued (30-day/tool cooldown) → tool-auditor Sonnet code review across six categories, findings by confidence (certain/likely → improve_tool, possible → needs_human_review), quality_score 1-10 tracked. Tool removal is a human decision via dashboard.
- **sources:** 020#tool_health, #tool-auditor; 019#Tool Quality Standards
- **relations:** tool-improver; component-quality-auditor (sections)
- **verify-later:** check_tool_health.go cooldowns

<!-- SOURCE: U04_idea_uk.md -->
### Content-rebuild de-tools tool pages (confirmed hazard)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "confirmed hazard, fix pending" (TODO P3, 2026-06-26).
- **what:** A needs_page / link_resolution_rebuild on a tool or game page regenerates the page from plan_sections, and the plan knows nothing about the interactive tool living in a section's rendered_html — so the tool is silently replaced with generated prose. Fix direction: route link maintenance through a preserve-sections re-render path, stamp source_item_id, add an interactivity-aware save guard. Flagged as a direct risk to idea.uk's post-P0 rebuild if tools land first.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md#P3; idea.uk/running_notes_2(6).md (backlog)
- **relations:** tool pipeline (005/016b/020/026 cross-refs); page-rerender vs page-build-handler distinction.
- **verify-later:** whether the preserve-sections path landed.

<!-- SOURCE: U05_content_quality_linking.md -->
### tool-recreation-handler (interactive rebuild path)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-26: "WI 6a28c4b3 completed … the playable A* tool is BACK".
- **what:** Recreates interactive tool/game pages from the adoption crawl: recreate_tool (Opus 64k, timeout 2400s) → check_tool_completeness → validate_tool → save_page_sections → update_page_status → spawn_rerender. Mode must be `recreate_tool` (not `recreate` — load_existing_content skips unless mode matches, a prior gotcha). Interactive pages are routed here from adoption via buildPageFeatureMap (T1 routing fix). One of the three save_page_sections callers, so it carries the Part-4 guards; re-creating a tool doesn't trip Layer 1 (new content IS interactive).
- **sources:** PLAN_pathfinding_missing_game.md#2; NOTES(44) 2026-06-25/26; HANDOFF_page_pipeline(11).md#4
- **relations:** interactive clobber; adoption pipeline T1 routing; item_key mis-key.
- **verify-later:** tool-recreation-handler default_config; apply_adoption_plan buildPageFeatureMap.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance criteria live in the tool's PLAN (fenced ```criteria JSON block)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECISION: acceptance criteria live in the tool's doc_plans PLAN" (2026-07-04 rev 4); consumed live by Tier 2 and Tier 4 checkers.
- **what:** The per-tool definition of *working*. Candidates judged on key/lifecycle/owner: `site_specs` — right machinery, wrong key (site-scoped; per-site copies drift); `site_plans`/directives — wrong lifecycle and owner (churniest artifact, planner-owned; "never store the bar in the artifact that regenerates most"); findings' `acceptance_test` — right pattern, wrong duration (dies with the work item; the standing criteria SEED it). The PLAN wins on all three axes. Format: a machine-extractable fenced ```criteria JSON block (tool-doc-header precedent), extracted by `load_doc_context` as `criteria_json`; lifts to a column only on volume. Per-site parametrisation goes to `direction.must_have`, not the PLAN.
- **sources:** PLAN_travelling_docs(6).md#where-acceptance-criteria-live; 001_README_acceptance_criteria.md; RUNNING_NOTES_travelling_docs(39).md#rev4
- **relations:** verification ladder; findings acceptance_test/max_fix_attempts (improvement-loop 004); direction.must_have.
- **verify-later:** criteria fence extraction in load_doc_context; `has_fence` on live PLANs.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria describe DELIVERED reality, not aspiration (Option B)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECIDED 2026-07-10: Option B — inline reality (user: 'I choose option B and surrender')"; migrations 143 (PLANs superseded) + 144 (composer fixed) applied and verified.
- **what:** The composer had asserted a designed-but-never-built JS extraction (`asset_loads /tools/assets/<fn>.js`) in every PLAN; Tier-2's first sweep failed every tool on it by construction. Principle on record: criteria must describe what the system delivers; aspirations live in roadmaps. If extraction ever ships, PLANs supersede forward again. Corollary: the composer's standard checks became boots/console/status/mobile-fit + optional interaction from real selectors.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5 (Option B block); HANDOFF_2026-07-10…md#§2; PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** js-not-extracted delivery gap; Tier-2 first sweep; PLAN supersede versioning.
- **verify-later:** current PLAN fences have no asset check; compose_plan prompt (four checks, inline delivery line).

<!-- SOURCE: U08_travelling_docs.md -->
### The tool verification ladder (Tiers 0–4)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "The verification ladder is whole (Tier 0/1/2/4) and closes on both outcomes" — RUNBOOK position 2026-07-12; Tier 3 remains Phase B.
- **what:** Cheap-to-expensive tiers, each catching a different class: Tier 0 generation-time output integrity (`HasToolDocHeader` gate + `check_tool_completeness`, deliberately flags-but-passes); Tier 1 structural post-deploy (`check_tool_health`); Tier 2 static contract-presence against deployed HTML (anchor rule); Tier 3 acceptance audit (`tool-auditor` vs criteria — Phase B, unbuilt extension); Tier 4 behavioural — drive the deployed tool in headless Chromium until criteria pass. Standing rule: never read a Tier-2 pass as "the tool works" — that claim belongs to Tier 4.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_travelling_docs(6).md#tool-assurance; OVERVIEW_self_verifying_tools.md#mechanism-2
- **relations:** every tier concept below; "passed checks ≠ working".
- **verify-later:** check_tool_completeness + check_tool_health + discovery_checks/check_tool_acceptance.go + browser-runner adapter, all in the chassis repo.

<!-- SOURCE: U08_travelling_docs.md -->
### "Completeness + validation passed" ≠ working — twice demonstrated
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN(6) rollout outcomes: "The June recreation introduced the economy-simulator's two bugs and passed; run 2 of the repair faithfully recreated them and passed."
- **what:** The standing empirical argument for the behavioural tier: structural/validation checks measure output integrity, not behaviour. The same game shipped broken twice while passing every existing check — the June 2026-06-05 recreation introduced the bugs (proven from tool_recreation_training rows and the origin game.js which has neither bug), and repair run 2 recreated them while its own note truthfully said "completeness + validation passed".
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45; OVERVIEW_self_verifying_tools.md#problem
- **relations:** Tier 4; seam rule; economy-simulator case.
- **verify-later:** tool_recreation_training rows for page d9a8e6e8 dated 2026-06-05.

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-2 static acceptance checker (discovery check `tool_acceptance`)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Stage 5 — LIVE 2026-07-10 ✓ (first sweep proven)" — run cd0d9731 on v1.0.1107 produced exactly the pre-verified findings (2 improve_tool items + 2 acceptance-fail notes, check-level precision confirmed).
- **what:** A browserless discovery check (sibling of `tool_health` in `discovery_checks/`): loads the current PLAN's criteria fence, fetches the deployed page (bounded 12s/2MB, cached per run), and evaluates the statically-visible subset under the anchor rule, plus shell checks (tool-doc header not leaked, no `<no value>` residue). No criteria → a `needs_criteria` note (30-day cooldown), never a fake pass. Failures → one improve_tool item (criteria embedded as `acceptance_test`, 7-day cooldown, cancelled items excluded since migration 146's correct-while-touching) + an acceptance-fail note. Scope limit by construction: only generator-created tools have content_components rows; adopted/recreated page-section tools are invisible to Tier 2.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5; RUNNING_NOTES_travelling_docs(39).md#stage-5-built,#stage-5-live; HANDOFF_2026-07-10…md#§1,§2
- **relations:** anchor rule; migration 142; Tier 4 (reaches page-section tools via pages).
- **verify-later:** `discovery_checks/check_tool_acceptance.go`; design-discovery-agent run_checks list.

<!-- SOURCE: U08_travelling_docs.md -->
### The anchor rule — static checks confirm, never refute
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "STAGE-5 RULE SETTLED" 2026-07-09 after the #tableWrap inspection (empty div filled by JS); implemented + unit-tested incl. the founding cases.
- **what:** Validate only a criteria selector's ANCHOR (leftmost id/class token) against `html_template`: `#tableWrap` exists ⇒ `#tableWrap tr` passes (rows are JS-built; Tier 4 asserts them for real); `#xpTableBody` exists nowhere ⇒ fails ⇒ drop or -EDIT. Static validation can confirm a selector but never refute one — never delete a check merely because the DOM is constructed at runtime. Motivated by the composer inventing selectors it ASSERTS on while copying real ones it ACTS on; the remedy is a check made by the system on itself, not a sterner prompt. Implementation detail banked: CSS class tokens are whitespace-delimited (Go regexp `\b` wrongly splits on hyphens).
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5-rule; RUNNING_NOTES_travelling_docs(39).md#rev39,#rev40; OVERVIEW_self_verifying_tools.md#tier-2
- **relations:** composer selector-invention incident; Tier 4 runtime assertions; tool-auditor (same logic belongs there — unbuilt).
- **verify-later:** anchor extraction + class-token comparison in check_tool_acceptance.go tests.

<!-- SOURCE: U08_travelling_docs.md -->
### Composer selector invention — caught twice, machine-corrected
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** CONFIRMED 2026-07-09 (`#xpTableBody`/`#statsStrip` token_anywhere=f); second sighting caught by Tier 2 itself 2026-07-10 (kebab `#drop-chance` vs real camelCase `#dropChance`).
- **what:** The PLAN-composer LLM invented DOM ids for assertion targets despite an explicit "never invent a selector" instruction — the rule held for controls it acts on and failed for things it asserts on. First instance corrected by a guarded supersede migration that itself initially refused a valid runtime selector (leading to the anchor rule); second caught automatically by the live Tier-2 sweep and corrected by migration 143. Demonstrates the design stance: hallucination is countered by verification, not prompt escalation.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#stage-5-live; 0NN_supersede_xp_curve_plan_selectors(2).sql; RUNBOOK_travelling_docs(38).md#task-3-proven
- **relations:** anchor rule; supersede versioning (correction recorded as a NOTES entry — "the travelling-docs loop applied to itself").
- **verify-later:** xp-curve PLAN v1→v2 chain + its correction note in doc_notes.

<!-- SOURCE: U08_travelling_docs.md -->
### tool-acceptance-agent — Tier 4 self-driving orchestrator
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** First machine acceptance-run note (run bf330ac6, 2026-07-12, "Tier-4 acceptance PASSED — all 3 evaluated checks"); fail path proven live via a controlled reverted test (failed=1, improve_tool_created=true, full teardown verified).
- **what:** An agent (migration 145) closing the loop with zero humans: `ensure_site_record → load_docs → request_browser_run (Kafka await; resolves the tool's deployed URL from pages itself; NO-OP skips without awaiting when the PLAN has no criteria) → judge_acceptance_results → complete`. Judge recomputes the verdict from results: all pass → acceptance-run note; any fail → acceptance-fail note + ONE improve_tool item (criteria embedded as acceptance_test, handler tool-improver); component-less recreated/adopted tools get the note but no item — logged honestly for manual routing. Trigger 087 (dry-run default).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tool-acceptance-agent-built,#tier-4-self-driving,#fail-path-proven; README_summary_paragraph2_for_discussion.md; 087_TRIGGER_tool_acceptance.sh (header)
- **relations:** browser-runner adapter; acceptance iteration loop; continuous sweep.
- **verify-later:** `platform/orchestration/actions/tool_acceptance_actions.go`; agent_definitions row tool-acceptance-agent; migration 145.

<!-- SOURCE: U08_travelling_docs.md -->
### Continuous acceptance — the `tool_acceptance_due` periodic sweep
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Built + migration 146 applied 2026-07-12, but "v1.0.1111 … the continuous sweep is NOT in the binary" (untracked-file trap); "GATE: continuous acceptance activates on the next image built from 83ba9bd4+" (T11, 2026-07-13 — state at unit close).
- **what:** A discovery check that emits one `acceptance_run` work item per active tool with a deployed page and current PLAN criteria, unless a verdict landed within 7 days or a run is open. Design calls: post-creation/post-improve hooks deliberately NOT used (they'd fire before the page redeploys — creation ends at 'planned', improve merely queues a rerender; the sweep only ever sees deployed pages); items emitted straight to `triaged` (acceptance needs no human judgment; `detected` items were observed sitting unswept); priority 90 so acceptance tests the NEW page after builds/rerenders.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tier-4-continuous,#v1.0.1111; HANDOFF_2026-07-10…md#T10,T11; README_summary_paragraph2_for_discussion.md
- **relations:** tool-acceptance-agent; untracked-file deploy trap; improve_tool cooldown (cancelled items excluded).
- **verify-later:** `discovery_checks/check_tool_acceptance_due.go` in the deployed image; first unattended acceptance-run note.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance iteration loop — iterate until criteria pass
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Both halves proven separately (fail path via controlled test 2026-07-12; fix agents write notes); "let a REAL failure flow through to tool-improver and back" still open at unit close.
- **what:** deploy → acceptance run → failing criterion → `improve_tool` item (criterion as `acceptance_test`, bounded by `max_fix_attempts`) → fixer loads PLAN+NOTES first → fix → append note → redeploy → re-run. Criteria hold the bar still across iterations; NOTES stop iterations fighting each other. *Working* = criteria pass. The one link proven only with a synthetic input is a real failure flowing through tool-improver and back.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_tool_acceptance_runner(2).md#flow; OVERVIEW_self_verifying_tools.md#autonomous-loop
- **relations:** findings acceptance_test pattern (improvement-loop); tool-improver; continuous sweep.
- **verify-later:** an improve_tool item with source 'acceptance' processed end-to-end by tool-improver.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria contract v0 (check-type vocabulary + profiles)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** P0 implements 3 of 7 check types; "the composer emitted "action":"select" … a verb the Tier-4 criteria vocabulary must now define" (open).
- **what:** The machine-readable criteria schema: `profiles: [desktop, mobile]`; check types `selector_exists`, `selector_count`, `no_console_errors`, `asset_loads`, `interaction` (fill/click/select steps + expect), `no_horizontal_overflow`, `page_status_ok`. Deterministic only in v0 — no LLM drives the browser. Desktop = Chromium 1366×900; mobile = one stable Playwright device descriptor (emulation first; real devices out of scope). Phasing P0 boot checks → P1 interpreter+mobile → P2 interactions → P3 screenshots (via the existing Backblaze deploy path) → P4 optional LLM-exploratory mode.
- **sources:** PLAN_tool_acceptance_runner(2).md#criteria-contract,#profiles,#phasing; RUNBOOK_travelling_docs(38).md#stage-6
- **relations:** browser-runner adapter (P0); multi-page tool criteria (open question — url_role field).
- **verify-later:** criteria interpreter coverage in run_checks_action.go; whether "select" verb was added.

<!-- SOURCE: U08_travelling_docs.md -->
### Multi-page tool documentation prerequisites
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK §5.4 "Multi-page prerequisite: preserve-sections re-render + interactivity-aware save guard (pending) before scaling page counts."
- **what:** Multi-page tools add a "Page set & inter-page contract" PLAN section (URLs, shared state keys, data feeds) and may need per-page checks (a `url_role` field). Scaling page counts is explicitly gated on the pending preserve-sections re-render and interactivity-aware save guard.
- **sources:** RUNBOOK_travelling_docs(38).md#§2,#§5; PLAN_travelling_docs(6).md#tool-assurance; PLAN_tool_acceptance_runner(2).md#open-questions
- **relations:** interactive-section clobber (Part 4) below; criteria contract.
- **verify-later:** save_page_sections interactivity guard deployment status.

<!-- SOURCE: U08_travelling_docs.md -->
### Recreation writes page sections — component-less tools and their visibility gap
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Established by query 2026-07-09 ("pages.sections is EMPTY … the 32 KB game body exists only as deployed HTML in the sites repo"); Tier-2 scope note 2026-07-10.
- **what:** `tool-recreation-handler` ends save_page_sections → update_status → deploy_page and never creates a `content_components` row — adopted/recreated tools exist only as page sections + deployed HTML (source in adoption-crawl research_results: adoption_crawl full markdown+rawHTML, adoption_page per-page; `spec.mode="recreate"` is the handshake set by apply_adoption_plan). Consequences: no component address for tool-improver; invisible to Tier 2 by construction (Tier 4 reaches them via pages); NOTES subject must be pipeline-scoped. `site_plan_sections` is site-plan STRUCTURE, not HTML.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev41,#rev42,#rev43; HANDOFF_2026-07-09…_1_.md#§4; RUNBOOK_travelling_docs(38).md#task-5-record
- **relations:** dangling-doc rule; adoption pipeline (007); Tier-2 scope limit.
- **verify-later:** tool-recreation-handler workflow steps; research_results result_types.

<!-- SOURCE: U09_adoption.md -->
### Tools/games behavioural QA loop (planned)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "PLAN_tools_games_behavioral_qa_loop.md (this session) — a standalone QA/maintenance loop that builds out the planned-but-unbuilt Tier 3 (headless behavioral testing) and adds a games lifecycle… Phased; first cut Phase 0+1" (HANDOFF_2026-06-06).
- **what:** A standalone QA loop for deployed interactive tools/games, motivated by real defects (Jelly Invaders degrading over time, P2P host replies not reaching mobile clients, untested cross-browser/mobile variants). Referenced from the adoption thread as FUTURE work; the plan doc itself lives elsewhere.
- **sources:** HANDOFF_2026-06-06#future, HANDOFF_2026-06-09#later-parked
- **relations:** tool-recreation-handler output quality (Family I1); 019/020 tool library/lifecycle Tier 3
- **verify-later:** PLAN_tools_games_behavioral_qa_loop.md (outside this unit)

<!-- SOURCE: U09_adoption.md -->
### Validation observability: structured rejection logging (recordValidationRejection)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Validation observability deployed: store_generated_component_action.go recordValidationRejection writes a structured agent_error_log row on every pre-store rejection… guide-list's attempt-1 failure was captured exactly — one SQL row, no pod-log forensics" (Session 5, 2026-05-11).
- **what:** Every component pre-store validation rejection writes a structured agent_error_log row (severity warning for bookkeeping vs error for structural; orphan/unknown field names as typed JSONB arrays), replacing pod-log forensics. Companion pattern: the retry budget of 3 is calibrated for the single-bookkeeping-orphan failure class seen in Tier-D regens (tool-list missed card_link_label, guide-list read_guide_label); a central label registry would prevent the class entirely (idea, not built).
- **sources:** FOCUS_directory_builder_and_list_components.md#tier-d-converge
- **relations:** component-creator; chrome-template gate (would reuse the same gate/log shape)
- **verify-later:** store_generated_component_action.go recordValidationRejection

<!-- SOURCE: U12_docs024_archives.md -->
### Mandatory minimum tool-suggestion count (2–5, no "suggest zero" option)
- **category:** tool-lifecycle
- **status-signal:** superseded
- **status-evidence:** Archive: "It returns 2–5 suggestions." Live: "It can return 0-5 suggestions. Returning zero is correct when no tools are appropriate."
- **what:** The earliest `tool-suggester` design forced the LLM to always propose at least two tools per site. Replaced by an explicit zero-is-valid design, directly tied to the same failure class as `matchToolToSite` (irrelevant tools forced onto sites).
- **sources:** old/older1/012_tool_lifecycle_guide.md#"Agent: tool-suggester"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Agent: tool-suggester"
- **relations:** tag-based deterministic tool-to-site matching (above)
- **verify-later:** check tool-suggester's current prompt for the zero-suggestions instruction.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool acceptance runner (headless-browser acceptance testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: initial plan (P0 not started)." (tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md, header)
- **what:** Tier-4 rung of the tool verification ladder: a `browser-runner-adapter` (Playwright+Chromium, mirroring analyser-adapter) drives a deployed tool page under desktop+mobile profiles, judges declared criteria (selector_exists, no_console_errors, asset_loads, interaction, no_horizontal_overflow, page_status_ok) pass/fail, feeding failures back as `improve_tool` work items until criteria pass. Criteria live in the tool's travelling PLAN as a criteria block.
- **sources:** tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md#Aim, #Criteria-contract, #Phasing
- **relations:** Behavioral QA loop for tools & games (this is the deterministic v0 layer); tool-lifecycle (020); a recent repo commit ("browser-runner-adapter: commit the full Tier-4 adapter") suggests adapter code may already exist — verify
- **verify-later:** `browser-runner-adapter` deployment; `tool-acceptance-agent` orchestrator; `max_fix_attempts` convention

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool widget clobber mechanism (M1: DELETE+INSERT rebuild wipes side-written components)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "M1 (clobber) is a confirmed latent defect that does not explain these pages but would bite once M2 is fixed. Fixes drafted, not implemented." (PLAN_tool_widget_clobber(9).md)
- **what:** `save_page_sections_action.go` rebuilds a page's `page_components` by `DELETE FROM page_components WHERE page_id=$1` then re-INSERTs only the sections the content writer supplied. Any side-written row not in that list — including a tool/game widget inserted by `create_tool_component`/`deploy_tool` at position 2 — is destroyed on the next `needs_content_page` build. A content-regression guard exists but compares only visible text length after stripping tags, so it is structurally blind to script-heavy widgets. Old content is snapshotted to `page_component_history` before delete, so wipes are recoverable/detectable.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.1-2.2,#3,#7, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-1
- **relations:** Two divergent tool-creation paths; site_plan as authoritative build source; Canonical tool-page section-shape design question; Recreation-loss defect
- **verify-later:** `save_page_sections_action.go` regression guard/delete lines; `page_component_history` rows with `source='save_page_sections_overwrite'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two divergent tool-creation paths (novel vs fork)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** table in §2.3 of PLAN_tool_widget_clobber(9).md documents both paths as existing, currently-running code
- **what:** `create_tool_component_action.go` (the "novel" path) never sets `pages.sections`, leaving it default `[]`; `deploy_tool_action.go` (the "fork" path) sets `pages.sections` to `["hero-tool","tool-guide-intro","<toolFunction>","tool-cta"]`. Both side-write the widget into `page_components` at position 2 and queue `needs_content_page`. The novel path is more exposed to the clobber mechanism since the widget is a member of no section list anywhere.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.3,#9
- **relations:** Tool widget clobber mechanism (M1); Canonicalise tool page identity (T3)
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`; `idx_cc_tool_function_unique` partial index behaviour

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-adoption detection check (T2 — check_tool_recreation_needed.go)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2: "T2 — check_tool_recreation_needed.go ... Deployed. Backfills automatically on next discovery run, if recreation works."
- **what:** A new `discovery_checks` package check: finds `page_type IN ('tool','game')`, `status='active'` pages with no widget, sources `interactive_features` from adoption findings via the same canonical-name transform as T1, and emits `needs_tool_recreation` (7-day per-page cooldown). Pages with no captured features are surfaced but deferred to the tool-suggester/generation path. Doubles as the backfill mechanism for pre-existing widget-less pages.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2
- **relations:** Adoption interactivity misroute (T1); check_tool_health blind spot
- **verify-later:** `check_tool_recreation_needed.go`; `idx_swi_dedup`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Recreation-loss defect (correctly-routed recreation still produces no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "Not confirmed: that widgets now actually deploy... correct routing → completed recreation → no widget... Hold the trigger." (HANDOFF §1,§3)
- **what:** Query K showed all five games on gamesdesign.co.uk — which had routed correctly to `tool-recreation-handler` all along and whose recreation work items completed — had no deployed widget component and no inline `<script>` section. So the routing fix (T1) is necessary but not sufficient; something downstream prevents the widget from landing. Diagnosis was interrupted mid-investigation when a parallel adoption chat reset the underlying state, so the exact mechanism remains open.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§2.9,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§5
- **relations:** Tool widget clobber mechanism (M1); tool-game-* duplicate pages; Post-adoption detection check (T2)
- **verify-later:** re-run queries R1-R3/L/M/N1/N2 against current gamesdesign.co.uk state; check `page_component_history` for a clobber snapshot on a game page

<!-- SOURCE: U13_docs024_small_dirs.md -->
### tool-game-* duplicate pages (T5)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "T5 — tool-game-* duplicate pages (step 8) ... Pending re-observe ... May have been wiped by step-9 reset" (PLAN_tool_widget_clobber(9).md §5b)
- **what:** Five `page_type=tool`, `build_status=planned` pages surfaced that duplicate the five existing games by name (`tool-game-<name>`). Candidate mechanisms: `tool-recreation-handler` building a separate page instead of populating the original interactive page, or a planner/reconciler role-divergence in the `029` canonicalisation family. The duplicates vanished in the step-9 state reset before their origin could be confirmed.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§6
- **relations:** Recreation-loss defect; Canonicalise tool page identity (T3)
- **verify-later:** query M (who created tool-game-* pages) re-pointed at current state

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonicalise tool page identity across surfaces (T3)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "T3 ... Open, independent ... Low risk; can land at any time." (HANDOFF §6)
- **what:** `create_tool_component` and `deploy_tool` build page name/url/page_type ad hoc, diverging from the canonical `datahelpers.CanonicalisePage` helper that adoption and the planner already use. Proposed fix: route both tool actions through the same canonical helper. Flagged as a gap in `029`'s Phase-0 deliverable list, which covered only two other files.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2,§6,§8
- **relations:** Two divergent tool-creation paths; tool-game-* duplicate pages
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`, `datahelpers/page_canonical.go`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonical tool-page section-shape design question and fix options
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Fix options (structural-first; not yet implemented)" (PLAN_tool_widget_clobber(9).md §5)
- **what:** Raises and answers (as a design decision, not yet built) whether a tool page even wants generic hero/guide-intro/CTA sections, or just the widget. Three options recorded: (1) make the widget a first-class section in whichever authority the build reads; (2) right-size the tool page's canonical section list; (3) make `save_page_sections` structure-aware as a safety net. Recommended: 1+2 together with 3 as a guard. Notes `content_guidance` already instructs the writer not to regenerate the widget, but the writer has no mechanical way to honour that.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §4,§5
- **relations:** Tool widget clobber mechanism (M1); site_plan as build authority; content-governance
- **verify-later:** whether `plan_sections_action.go` now emits a tool/embed section for `page_type='tool'` pages

<!-- SOURCE: U13_docs024_small_dirs.md -->
### check_tool_health INNER JOIN blind spot
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "check_tool_health blind spot. Its INNER JOIN content_components → page_components means a tool with no linked page_components row ... is invisible" (PLAN_tool_widget_clobber(9).md §8)
- **what:** The Tier-1 tool health check joins `content_components` to `page_components` with an INNER JOIN, so a `page_type='tool'` page with zero linked components (post-clobber, or never-generated) is invisible to it and the check silently reports "no tools" as a pass. T2 partially closes this by detecting the same condition independently, but the original check itself was not corrected.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §9
- **relations:** Post-adoption detection check (T2); Recreation-loss defect
- **verify-later:** `check_tool_health.go` join logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### forked_from NULL collision risk on novel tools
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "forked_from NULL on novel tools ... Two sites generating the same function would collide. Latent; not today's bug." (PLAN_tool_widget_clobber(9).md §8)
- **what:** `create_tool_component` omits `forked_from`, so novel/generated tools are classified as library tools by the partial unique index `idx_cc_tool_function_unique (function) WHERE component_level='tool' AND forked_from IS NULL AND is_active`. Two different sites independently generating a tool with the same function name would collide.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8
- **relations:** Two divergent tool-creation paths
- **verify-later:** `idx_cc_tool_function_unique` definition; whether any collision has actually occurred

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Behavioral QA loop for tools & games (Tier 3+ headless-browser testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: proposed (2026-06-06)." (PLAN_tools_games_behavioral_qa_loop.md, header)
- **what:** A standalone, slower-cadence QA loop that runs generated tools/games in an isolated multi-engine Playwright pod over time under synthetic drive, to catch defect classes invisible to a single render/screenshot: temporal degradation, cross-browser divergence, mobile-specific layout/touch bugs, and multi-context networked/relay failures. Correctness judged via a three-layer oracle: generic deterministic invariants, type-specific assertions, and LLM-as-judge over a screenshot/video series — with auto-fix gated to high-confidence deterministic findings only. Reuses the existing check→work-item→improver pipeline.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md#1-Why,#4-The-headless-pod,#5-The-oracle-problem,#10-Phasing
- **relations:** Tool acceptance runner (this loop is the heavier behavioral/temporal successor); Games quality lifecycle parity; tool-lifecycle (020)
- **verify-later:** whether any phase has been built; `qa_runs`/`last_qa_at` storage location

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tool-doc header rollout (provenance + stripped headers)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** thin_slice(27) "Tool-doc header rollout (2026-06-11) — apply order is load-bearing. … Three stages; do not reorder — the gate without the prompt fails every generation, and the stamps without the columns fail every insert." No completion claim in this unit's files.
- **what:** Rollout procedure for tool documentation headers: (1) provenance columns on content_components (source_agent_type, source_orchestration_id), (2) anchored idempotent prompt updates adding the `=== tool-doc ===` header requirement (abort if prompts drifted), (3) one binary release (tool_doc_header.go + five action edits) so headers are stamped in the DB template but STRIPPED from shipped pages/CDN assets, with a tool_health no_doc_header WARNING converging old tools on the normal sweep — no retrofit campaign.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#tool-doc-header-rollout
- **relations:** doc_plans/doc_notes (the tools thread's later system); tiered tool acceptance
- **verify-later:** content_components source_% columns; '=== tool-doc ===' in html_template rows; tool_health sweep items

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tiered tool acceptance (static contract check + browser-runner)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Their Stages 5–6 define a TIERED ACCEPTANCE system for tools: a static Tier-2 contract-presence check and a Tier-4 browser-runner adapter (Chromium+Playwright, Kafka request/response per the 035 Adapter Guide) — their 'loop for complicated tools' is acceptance/verification + docs, NOT a rival diagnosis loop."
- **what:** The tools thread's acceptance ladder for generated tools, recorded here as a shared component: Tier-2 static contract-presence verification and a Tier-4 browser-runner adapter executing tools in real Chromium — also earmarked as a future verification service for fix-loop F1 fixes touching pages and as a council reviewer's instrument.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** council of reviewers; tool pipeline; adapters (035 guide)
- **verify-later:** browser-runner adapter existence; tool acceptance stages in the tools thread docs

<!-- SOURCE: U15_docs019_running_notes.md -->
### Tool-doc header system
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DONE 2026-06-11: 019/020 CONSOLIDATED — splice files retired" (principles(59)); status marked apply-ready and untouched in v2(36) small-pending list.
- **what:** A standardised 6-12 line sentinel-delimited header block written into every generated tool's script (purpose, behavioural invariants, no-external-calls, version marker) at creation time, stripped at deploy-assembly (three call sites: single-page rerender, `collectJSAssets`, bulk rerender) so it never ships to visitors but is retained in the DB `html_template` for audit/parse parity. Enforced via a hard `HasToolDocHeader` gate in `create_tool_component`, tool-generator/tool-improver prompt edits, and two new `tool_health` tier-1 checks (`no_doc_header` warning, `malformed_doc_header` error). Paired with new `source_agent_type`/`source_orchestration_id` provenance columns on `content_components`, mirroring `knowledge_base`'s existing provenance pair.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tool-doc entries (multiple DONE items).
- **relations:** JS content separation contract; doc claim-verification convention; canonical-doc-home discipline.
- **verify-later:** Whether the rollout (provenance migration → prompts SQL → binary release) was ever applied — repeatedly flagged as "apply-ready, not yet applied" across all later notes files through 2026-07-06.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Fork-divergence detection for library tools
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "IMMEDIATE WIN INSTEAD: FORK-DIVERGENCE detection — pure SQL discovery check (tier-1, zero cost)" (principles(59)).
- **what:** A proposed zero-cost SQL discovery check comparing a deployed fork's `html_template` hash against its `forked_from` library original to answer "which forks are unmodified / safe to bulk-push a library change" — deliberately deferred building full code-symbol indexing of tools (each tool is one IIFE, thin symbol pickings; tool discovery already solved via `semantic_tags`) until a concrete consumer needs it.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tools/provenance/docs design entry.
- **relations:** Tool-doc header system; JS content separation contract.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Tool page missing widget (M1 clobber vs M2 misroute)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 016 addendum(4) "RESOLVED 2026-05-26 → b1 … key the feature map by the canonical name in buildPageFeatureMap"; companion PLAN_tool_widget_clobber.md
- **what:** A `page_type='tool'` page rendering a description but no widget has two causes needing different fixes: M1 clobber (`SavePageSectionsAction` deletes page_components and its content-regression guard can't see a script-heavy widget) vs M2 never-generated (adoption recreate has no parse stage). For gamesdesign, root cause was a misroute: `buildPageFeatureMap` keys by raw page name while the route looks up canonicalised (`tool-`-prefixed) names.
- **sources:** WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#root-cause-m2-corrected-after-verification, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#diagnostic-recipe-read-only-30-seconds, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** CanonicalisePage; adoption parse-stage; site plan reconciler; interactive content
- **verify-later:** buildPageFeatureMap; tool-recreation-handler; SavePageSectionsAction; PLAN_tool_widget_clobber.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool quality tiers: tool-auditor (Tier 2 LLM review), tool-improver, acceptance checks (Tier 2 static) and tool-acceptance-agent (Tier 4 browser runs)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 088 (tool-auditor); 142 enables tool_acceptance check (2026-07-10, doc_notes entry "unit tests... green"); 145 inserts tool-acceptance-agent; 146 makes Tier 4 continuous via tool_acceptance_due sweep.
- **what:** Layered tool verification. Tier 1: check_tool_health structural checks. Tier 2 (LLM): tool-auditor reads full HTML/CSS/JS and reasons through logic/mobile/UX/accessibility, creating improve_tool or needs_human_review items. Tier 2 (static): check_tool_acceptance asserts the PLAN's criteria fence against the deployed page under the ANCHOR RULE ("validate a selector's leftmost id/class token, never the whole path; confirm, never refute; -EDIT ids skipped"). Tier 4: tool-acceptance-agent drives the deployed tool in headless Chromium via the browser-runner adapter against PLAN criteria — "the tier that turns 'deployed' into 'works'" — pass → acceptance-run note; fail → acceptance-fail note + one improve_tool item carrying criteria as acceptance_test. tool-improver executes improve_tool fixes. 7-day cooldowns; cancelled items excluded from cooldown (146).
- **sources:** 088_tool_auditor_agent.sql; 142_enable_tool_acceptance_check.sql; 145_tool_acceptance_agent.sql; 146_enable_tool_acceptance_due.sql; 062_tool_suggester_and_improver.sql
- **relations:** travelling PLAN criteria fences; design-discovery-agent hosts the checks; browser-runner adapter
- **verify-later:** request_browser_run / judge_acceptance_results actions; check_tool_acceptance.go anchor rule; browser-runner adapter deployment

<!-- SOURCE: U18_sql_for_agents.md -->
### Acceptance-criteria honesty: invented selectors and inline-delivery decisions
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 136 (2026-07-09) repairs the first machine-written PLAN's invented ids (#xpTableBody→#tableWrap tbody, #statsStrip→#statRow); 143/144 (2026-07-10) "PLANs surrender to delivered reality" — asset extraction "was designed but never built", so criteria drop asset_loads and the composer prompt is corrected.
- **what:** Two recurring failure classes in machine-written acceptance criteria, and their durable remedies: (1) composers invent selectors they ASSERT on even while obeying never-invent for controls they ACT on — remedy is Tier-2 static validation of criteria selectors against html_template (anchor rule), not sterner prompts; (2) criteria must describe what the system DELIVERS, not aspirations — the /tools/assets/<fn>.js extraction path was never built, all JS ships inline, so PLANs and the composer prompt were superseded to inline delivery ("born honest"). Also note the abandoned mechanism: Path-1 tool asset extraction on rerender.
- **sources:** 136_supersede_xp_curve_plan_selectors.sql; 143_supersede_plans_inline_delivery.sql; 144_composer_inline_delivery.sql; 113_site_asset_renderer.sql (the extraction design it contradicts)
- **relations:** travelling docs supersede pattern; tool acceptance tiers
- **verify-later:** whether asset extraction ever ships (would trigger forward supersede)

<!-- SOURCE: U19_sql_tables_components.md -->
### Component quality tracking (0–100 score)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "None of these fields are required by the existing pipeline — they are additive... selector will use them when present and ignore when NULL" (005 ~9848).
- **what:** Additive quality fields on content_components computed by a compute_component_quality action, with indexes for auditor queries (below threshold OR unscored) and planner preference (higher quality per function). Distinct from avg_quality_score in the selector metadata set.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-quality-tracking
- **relations:** component selector metadata; improvement loop auditors.
- **verify-later:** compute_component_quality action in registry; populated quality_score values.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component versioning (component_versions)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** Table created in schema-mode migration (008 PART 3); page_components.component_version_id exists in live dump with comment "if versioning enabled".
- **what:** Versioned snapshots of component templates (html_template, css_template, input_schema per version_number) so strict-mode pages could pin a specific template version. Referenced as an optional backup target in later template-fix migrations; unclear whether any writer maintains it.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#PART3; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** schema-mode subsystem (abandoned); site_plan_sections.component_version_id (planner provenance).
- **verify-later:** row count in component_versions; writers in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Tool library fork-on-deploy model
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** forked_from column, partial unique index on function scoped to canonical tools, and the later constraint amendment "Forks (forked_from IS NOT NULL) are excluded from the uniqueness check" fixing the add_tool failure on gamedesign.uk.
- **what:** Library tools are canonical rows (component_level='tool', forked_from IS NULL); deploying to a site copies the row as a fork (forked_from = library id) referenced by page_components. Library changes never cascade to forks; fleet updates go through per-site work items. Uniqueness of `function` applies only to active canonical tools so many site forks can share a function; forks are only ever addressed by component_id.
- **sources:** docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#fork-constraint-fix; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#idx_cc_tool_function_unique
- **relations:** component library; seeded tool library; improvement-loop fleet updates.
- **verify-later:** deployer fork-copy code; fork counts per library tool.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component regeneration in place (store_generated_component mechanics)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 083 result: brief-explanation updated in place (same id, created_at unchanged, status 'regenerated', component_versions snapshot, needs_rerender raised) — "matches the documented behaviour (003 §348)".
- **what:** store_generated_component looks up an existing component by the LLM's EMITTED `function` (forked_from IS NULL); if found, it snapshots the current row to component_versions (MAX+1), UPDATEs in place (component_id preserved → all page/site FKs keep resolving), sets template/schema/js_content/render_mode/is_active, then markPagesPendingRebuild raises ONE needs_rerender per affected site. Determinism hazard: regeneration keys on the emitted function name — an unpinned LLM can emit a different name and INSERT a stray duplicate (the 081 'general-hero' incident); pin the function in the description. Pre-store validation rejects `<no value>` templates and checks placeholder/schema parity.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-06-30-~18:35 + #2026-07-01-~12:46; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e
- **relations:** shared library guard; component-quality-auditor; call_agent contract validation (the trigger saga)
- **verify-later:** store_generated_component_action.go lookup + snapshot + markPagesPendingRebuild; component_versions rows

<!-- SOURCE: U23_docs_root_vonc.md -->
### component-quality-auditor auto-regeneration threshold
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Read from its default_config 2026-06-29: creates needs_component_regeneration items only for quality_score < 50, handler component-creator, spec {function, component_id, quality_score, quality_issues}.
- **what:** The auditor raises regeneration work items for low-quality components — but its strict `< 50` condition meant the three vonc shells scoring EXACTLY 50 were never auto-picked-up (explaining zero queued items and requiring manual triggers). Its item shape confirms the designed regen path keys on function and routes to component-creator. Boundary-condition gap worth a rule review; also the future home of the autonomy plan's maintenance detections.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~21:00; docs/PLAN_dynamic_sections_and_loaders(4).md#maintenance
- **relations:** component regeneration in place; autonomous section composition (auditor rules gap 4)
- **verify-later:** component-quality-auditor default_config condition; quality_score distribution at exactly 50

<!-- SOURCE: U23_docs_root_vonc.md -->
### Store-path template validation (+ pending <script>-balance hardening)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Existing checks confirmed in code 2026-06-29 (`<no value>` rejection, placeholder/schema parity, unclosed `<style>`, section/div presence); the `<script>` balance check + separateInlineJS truncation warning remain "STILL MISSING" / backlog item 2 on 2026-07-09.
- **what:** store_generated_component's pre-store validation gate rejects Mode-A/B artifacts and unclosed `<style>` but NOT an unclosed `<script>` — the gap that let provocation-card ship a truncated inline script that swallowed the page footer at render. Hardening definition: add a `<script>` open/close balance check (reject or flag-for-regeneration) plus a truncation warning in separateInlineJS. Prevents the class "truncated template ships and breaks the page".
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-07-03-~13:25 (hardening def); docs/RUNBOOK_phase2_provocation_js(29).md#appendix-g; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** legacy un-extracted shells (the truncation instance); Mode A/B taxonomy
- **verify-later:** store_generated_component_action.go validation block for script balance

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Tool widget clobber (save_page_sections DELETE+INSERT destroys widgets)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.2: `SavePageSectionsAction` does `DELETE FROM page_components WHERE page_id=$1` then re-inserts only writer sections; content-regression guard "compares visible text length … structurally blind to tools"; M1 confirmed latent, "Fixes drafted, not implemented".
- **what:** Two writers collide: create_tool_component/deploy_tool side-writes a tool widget into `page_components`, but the authoritative build rebuilds page_components by DELETE+INSERT from the section list (whose authority is `site_plan`, synced into `pages.sections`). The widget isn't in that list, so the first `needs_content_page` build deletes it (snapshotted to `page_component_history` with `source='save_page_sections_overwrite'`). Fix options: make the widget a first-class site_plan section (preferred), right-size the tool page, or make save_page_sections structure-aware.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2-confirmed-findings, #5-fix-options; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md
- **relations:** load_page_sections_from_spec authority; adoption interactivity misroute; live tools/tool_widget_clobber/ set
- **verify-later:** save_page_sections_action.go; load_page_sections_from_spec_action.go; create_tool_component_action.go vs deploy_tool_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adoption interactivity misroute (canonical prefix desync)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.7 "b1 confirmed"; HANDOFF T1 "Deployed": `buildPageFeatureMap` keyed the feature map by raw `fm["page"]` while the routing loop looked up `CanonicalisePage`-canonicalised names (tool branch adds `tool-` prefix), so tool lookups always missed → empty `Features` → static `needs_content_page` route; games (already `game-` prefixed) matched.
- **what:** Adopted tool pages rendered as static description pages because the feature-map key (bare slug) never matched the canonicalised lookup key (with `tool-` prefix). Fixed (T1) by keying `buildPageFeatureMap` on the canonical name resolved from `plan["pages"]`, so routing and content attachment both land in the same namespace. Self-contained one-function change.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.7-resolved, #5b-tasks; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t1
- **relations:** CanonicalisePage (029); tool-recreation-handler; T3 canonicalise create_tool_component/deploy_tool
- **verify-later:** apply_adoption_plan_action.go buildPageFeatureMap; datahelpers/page_canonical.go CanonicalisePage

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### tool-recreation-handler + recreation discovery check (T2)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF T2 "Deployed": `check_tool_recreation_needed.go` (discovery_checks) detects `page_type IN ('tool','game')` active pages with no tool/game component and no inline `<script>`, sources `interactive_features` from adoption findings by canonical name, emits `needs_tool_recreation:<page>` (7-day cooldown). tool-recreation-handler workflow: `recreate_tool`(execute_llm_prompt)→`check_tool_completeness`→`spawn_rerender`→page-rerender.
- **what:** A registered agent that LLM-recreates interactive widgets for pages adoption captured as text-only, plus a discovery check that backfills widget-less interactive pages automatically on the next scheduled run. Item key deliberately distinct from adoption's `needs_page:<name>` to avoid collision.
- **sources:** tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t2; tools/PLAN_tool_widget_clobber(11).md#5b-tasks
- **relations:** adoption interactivity misroute (T1); recreation-loss defect (T4); check_tool_health blind spot
- **verify-later:** check_tool_recreation_needed.go; tool-recreation-handler agent_definition; check_tool_completeness_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Recreation-loss defect (correct routing yields no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** HANDOFF §3 / PLAN(11) §2.8 (step 8): five games routed `needs_tool_recreation → tool-recreation-handler` and completed (query I), yet all five are widget-less (`has_widget_component=f, has_script_section=f`), plus five new `tool-game-*` duplicate planned pages; step 9 state reset (L/M/N returned 0 rows) left diagnosis incomplete.
- **what:** Even correctly-routed, completed recreation didn't land a deployed widget — a second active defect downstream of routing (T4), blocking. Candidate mechanisms: recreation mis-targeted a parallel `tool-game-*` page, M1 clobber, handler completed without persisting, or a snippets false-negative (inline `<script>` extracted to `/assets/js/snippets.js`). Must be diagnosed before bulk-triggering backfill.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.8-blocking, #2.9-state-reset; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#3
- **relations:** tool widget clobber (M1); tool-recreation-handler; snippets extraction mechanism
- **verify-later:** tool-recreation-handler recreate_tool→check_tool_completeness→spawn_rerender; page_component_history source values

<!-- SOURCE: U25_leopardess_social.md -->
### Mode-B rendered-artifact templates (components stored as rendered output)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** VERDICT §2 (2026-07-09): "rendered_html == html_template with all '<no value>' removed — which the byte counts confirm exactly"; "they are rendered outputs stored as source templates".
- **what:** A component corruption class: html_template full of bare `<no value>`, zero {{.}} slots, empty input_schema — the stored template IS a rendered artifact. Consequences: render is a pure function of the template (predictable to the byte — used twice as an acceptance test); content_data is dead weight; repair_template_slots cannot repair them (zero `</no>` tags → needs_regeneration); for runtime-fill shells the emptiness is accidentally exactly what the loader needs, so regeneration must consciously re-establish the empty-shell contract or sections ship with baked copy.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#2, #8.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-24
- **relations:** runtime-fill guards; component selector/creator (regeneration path); problem-category taxonomy (empty-shell/mode-b-template)
- **verify-later:** components with `<no value>` and 0 schema fields fleet-wide

<!-- SOURCE: U25_leopardess_social.md -->
### Component-creator invocation contract (dual placement + quote-free description)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** NOTES_brief-explanation 083 (2026-07-01) "SUCCEEDED (in-place UPDATE)" via dual placement; framework fix "PATCH_validate_input_contract.go — drafted, not deployed" (HANDOFF §9.3).
- **what:** Manually invoking component-creator (spawn+call) must satisfy BOTH the input_contract (top-level required fields — call_agent validates against top-level extracted fields) AND the workflow's field paths (input_data.spec.*): the working pattern places section_type both top-level and inside spec, pins the function name in the description so the store lands as an in-place UPDATE (else a stray component INSERTs), and keeps the description quote-free to survive the kcat/JSON pipeline. The generic build-dispatch-loop cannot satisfy top-level-required contracts (same class); the durable fix — contract validator accepting top-level OR input_data.spec.{field} — is drafted, not deployed. Regeneration semantics: UPDATE-in-place keyed by function, component_versions snapshot, auto needs_rerender per affected site, store validation rejects `<no value>` templates.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md (081/082/083 arc); docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Component-creator-input
- **relations:** component selector/creator; shared component library semantics
- **verify-later:** call_agent contract validation code; PATCH_validate_input_contract.go status

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Interactive-page de-tool hazard (content rebuild silently drops a tool/game)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** CONFIRMED 2026-06-22 (gamesdesign game-pathfinding, 18KB A* canvas overwritten 06-14); "fix pending" in 002/005/016; 016b v2 update: two-layer save_page_sections fix WRITTEN, un-deployed
- **what:** A tool lives as a section's rendered_html, not a planned section, so any full rebuild (needs_page or link_resolution_rebuild) regenerates from plan_sections and replaces it with generic-text-block; the prose-based content-regression guard doesn't catch markup/JS loss. Fix layers: interactivity-aware save guard + carry-forward of interactive sections in save_page_sections (written), source_item_id stamping into page_component_history, and routing link maintenance through a preserve-sections path (page_rerender ruled out for CTA re-resolution — it doesn't re-run link logic).
- **sources:** 005(1) hazard block; 002(4)#Interactive-page hazard; 016 §9 final entry; 016b Part 4
- **relations:** phantom-CTA resolution bug (separate); tool recreation mis-key (Part 3)
- **verify-later:** save_page_sections_action.go patched version deployed?; page_component_history.source_item_id population

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool doc header (sentinel comment; stripped at deploy)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 019 full lifecycle table incl. StripToolDocHeader call sites and tool_health checks
- **what:** Every new tool's script opens with one sentinel-delimited block (function/purpose/behaviour/inputs/outputs; never ids/dates; no */). It never ships: StripToolDocHeader runs at the three outbound assembly points; DB rendered_html retains it for audit parity. Creation gate validates presence; improver preserves/updates it; auditor audits code AGAINST its stated behaviour; malformed (opener without closer) is left in and flagged by tool_health.
- **sources:** 019#Tool Doc Header; 020 tool_health tier-1 checks
- **relations:** per-tool travelling docs (037); tool-auditor
- **verify-later:** platform/content/tool_doc_header.go; prompt migration applied

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool quality three tiers (structural / LLM audit / headless-browser future)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 020: tiers 1–2 automated live, tier 3 "planned"
- **what:** tool_health tier 1 (Go, free): deploy status, HTML/template present, script/style/@media, hex/external-dep warnings, doc-header checks — blockers create improve_tool. Tier 2: audit_tool queued (30-day/tool cooldown) → tool-auditor Sonnet code review across six categories, findings by confidence (certain/likely → improve_tool, possible → needs_human_review), quality_score 1-10 tracked. Tool removal is a human decision via dashboard.
- **sources:** 020#tool_health, #tool-auditor; 019#Tool Quality Standards
- **relations:** tool-improver; component-quality-auditor (sections)
- **verify-later:** check_tool_health.go cooldowns

<!-- SOURCE: U04_idea_uk.md -->
### Content-rebuild de-tools tool pages (confirmed hazard)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "confirmed hazard, fix pending" (TODO P3, 2026-06-26).
- **what:** A needs_page / link_resolution_rebuild on a tool or game page regenerates the page from plan_sections, and the plan knows nothing about the interactive tool living in a section's rendered_html — so the tool is silently replaced with generated prose. Fix direction: route link maintenance through a preserve-sections re-render path, stamp source_item_id, add an interactivity-aware save guard. Flagged as a direct risk to idea.uk's post-P0 rebuild if tools land first.
- **sources:** idea.uk/TODO_chassis_and_idea_uk(1).md#P3; idea.uk/running_notes_2(6).md (backlog)
- **relations:** tool pipeline (005/016b/020/026 cross-refs); page-rerender vs page-build-handler distinction.
- **verify-later:** whether the preserve-sections path landed.

<!-- SOURCE: U05_content_quality_linking.md -->
### tool-recreation-handler (interactive rebuild path)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-26: "WI 6a28c4b3 completed … the playable A* tool is BACK".
- **what:** Recreates interactive tool/game pages from the adoption crawl: recreate_tool (Opus 64k, timeout 2400s) → check_tool_completeness → validate_tool → save_page_sections → update_page_status → spawn_rerender. Mode must be `recreate_tool` (not `recreate` — load_existing_content skips unless mode matches, a prior gotcha). Interactive pages are routed here from adoption via buildPageFeatureMap (T1 routing fix). One of the three save_page_sections callers, so it carries the Part-4 guards; re-creating a tool doesn't trip Layer 1 (new content IS interactive).
- **sources:** PLAN_pathfinding_missing_game.md#2; NOTES(44) 2026-06-25/26; HANDOFF_page_pipeline(11).md#4
- **relations:** interactive clobber; adoption pipeline T1 routing; item_key mis-key.
- **verify-later:** tool-recreation-handler default_config; apply_adoption_plan buildPageFeatureMap.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance criteria live in the tool's PLAN (fenced ```criteria JSON block)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECISION: acceptance criteria live in the tool's doc_plans PLAN" (2026-07-04 rev 4); consumed live by Tier 2 and Tier 4 checkers.
- **what:** The per-tool definition of *working*. Candidates judged on key/lifecycle/owner: `site_specs` — right machinery, wrong key (site-scoped; per-site copies drift); `site_plans`/directives — wrong lifecycle and owner (churniest artifact, planner-owned; "never store the bar in the artifact that regenerates most"); findings' `acceptance_test` — right pattern, wrong duration (dies with the work item; the standing criteria SEED it). The PLAN wins on all three axes. Format: a machine-extractable fenced ```criteria JSON block (tool-doc-header precedent), extracted by `load_doc_context` as `criteria_json`; lifts to a column only on volume. Per-site parametrisation goes to `direction.must_have`, not the PLAN.
- **sources:** PLAN_travelling_docs(6).md#where-acceptance-criteria-live; 001_README_acceptance_criteria.md; RUNNING_NOTES_travelling_docs(39).md#rev4
- **relations:** verification ladder; findings acceptance_test/max_fix_attempts (improvement-loop 004); direction.must_have.
- **verify-later:** criteria fence extraction in load_doc_context; `has_fence` on live PLANs.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria describe DELIVERED reality, not aspiration (Option B)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DECIDED 2026-07-10: Option B — inline reality (user: 'I choose option B and surrender')"; migrations 143 (PLANs superseded) + 144 (composer fixed) applied and verified.
- **what:** The composer had asserted a designed-but-never-built JS extraction (`asset_loads /tools/assets/<fn>.js`) in every PLAN; Tier-2's first sweep failed every tool on it by construction. Principle on record: criteria must describe what the system delivers; aspirations live in roadmaps. If extraction ever ships, PLANs supersede forward again. Corollary: the composer's standard checks became boots/console/status/mobile-fit + optional interaction from real selectors.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5 (Option B block); HANDOFF_2026-07-10…md#§2; PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** js-not-extracted delivery gap; Tier-2 first sweep; PLAN supersede versioning.
- **verify-later:** current PLAN fences have no asset check; compose_plan prompt (four checks, inline delivery line).

<!-- SOURCE: U08_travelling_docs.md -->
### The tool verification ladder (Tiers 0–4)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "The verification ladder is whole (Tier 0/1/2/4) and closes on both outcomes" — RUNBOOK position 2026-07-12; Tier 3 remains Phase B.
- **what:** Cheap-to-expensive tiers, each catching a different class: Tier 0 generation-time output integrity (`HasToolDocHeader` gate + `check_tool_completeness`, deliberately flags-but-passes); Tier 1 structural post-deploy (`check_tool_health`); Tier 2 static contract-presence against deployed HTML (anchor rule); Tier 3 acceptance audit (`tool-auditor` vs criteria — Phase B, unbuilt extension); Tier 4 behavioural — drive the deployed tool in headless Chromium until criteria pass. Standing rule: never read a Tier-2 pass as "the tool works" — that claim belongs to Tier 4.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_travelling_docs(6).md#tool-assurance; OVERVIEW_self_verifying_tools.md#mechanism-2
- **relations:** every tier concept below; "passed checks ≠ working".
- **verify-later:** check_tool_completeness + check_tool_health + discovery_checks/check_tool_acceptance.go + browser-runner adapter, all in the chassis repo.

<!-- SOURCE: U08_travelling_docs.md -->
### "Completeness + validation passed" ≠ working — twice demonstrated
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN(6) rollout outcomes: "The June recreation introduced the economy-simulator's two bugs and passed; run 2 of the repair faithfully recreated them and passed."
- **what:** The standing empirical argument for the behavioural tier: structural/validation checks measure output integrity, not behaviour. The same game shipped broken twice while passing every existing check — the June 2026-06-05 recreation introduced the bugs (proven from tool_recreation_training rows and the origin game.js which has neither bug), and repair run 2 recreated them while its own note truthfully said "completeness + validation passed".
- **sources:** PLAN_travelling_docs(6).md#rollout-outcomes; RUNNING_NOTES_travelling_docs(39).md#rev45; OVERVIEW_self_verifying_tools.md#problem
- **relations:** Tier 4; seam rule; economy-simulator case.
- **verify-later:** tool_recreation_training rows for page d9a8e6e8 dated 2026-06-05.

<!-- SOURCE: U08_travelling_docs.md -->
### Tier-2 static acceptance checker (discovery check `tool_acceptance`)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Stage 5 — LIVE 2026-07-10 ✓ (first sweep proven)" — run cd0d9731 on v1.0.1107 produced exactly the pre-verified findings (2 improve_tool items + 2 acceptance-fail notes, check-level precision confirmed).
- **what:** A browserless discovery check (sibling of `tool_health` in `discovery_checks/`): loads the current PLAN's criteria fence, fetches the deployed page (bounded 12s/2MB, cached per run), and evaluates the statically-visible subset under the anchor rule, plus shell checks (tool-doc header not leaked, no `<no value>` residue). No criteria → a `needs_criteria` note (30-day cooldown), never a fake pass. Failures → one improve_tool item (criteria embedded as `acceptance_test`, 7-day cooldown, cancelled items excluded since migration 146's correct-while-touching) + an acceptance-fail note. Scope limit by construction: only generator-created tools have content_components rows; adopted/recreated page-section tools are invisible to Tier 2.
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5; RUNNING_NOTES_travelling_docs(39).md#stage-5-built,#stage-5-live; HANDOFF_2026-07-10…md#§1,§2
- **relations:** anchor rule; migration 142; Tier 4 (reaches page-section tools via pages).
- **verify-later:** `discovery_checks/check_tool_acceptance.go`; design-discovery-agent run_checks list.

<!-- SOURCE: U08_travelling_docs.md -->
### The anchor rule — static checks confirm, never refute
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "STAGE-5 RULE SETTLED" 2026-07-09 after the #tableWrap inspection (empty div filled by JS); implemented + unit-tested incl. the founding cases.
- **what:** Validate only a criteria selector's ANCHOR (leftmost id/class token) against `html_template`: `#tableWrap` exists ⇒ `#tableWrap tr` passes (rows are JS-built; Tier 4 asserts them for real); `#xpTableBody` exists nowhere ⇒ fails ⇒ drop or -EDIT. Static validation can confirm a selector but never refute one — never delete a check merely because the DOM is constructed at runtime. Motivated by the composer inventing selectors it ASSERTS on while copying real ones it ACTS on; the remedy is a check made by the system on itself, not a sterner prompt. Implementation detail banked: CSS class tokens are whitespace-delimited (Go regexp `\b` wrongly splits on hyphens).
- **sources:** RUNBOOK_travelling_docs(38).md#stage-5-rule; RUNNING_NOTES_travelling_docs(39).md#rev39,#rev40; OVERVIEW_self_verifying_tools.md#tier-2
- **relations:** composer selector-invention incident; Tier 4 runtime assertions; tool-auditor (same logic belongs there — unbuilt).
- **verify-later:** anchor extraction + class-token comparison in check_tool_acceptance.go tests.

<!-- SOURCE: U08_travelling_docs.md -->
### Composer selector invention — caught twice, machine-corrected
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** CONFIRMED 2026-07-09 (`#xpTableBody`/`#statsStrip` token_anywhere=f); second sighting caught by Tier 2 itself 2026-07-10 (kebab `#drop-chance` vs real camelCase `#dropChance`).
- **what:** The PLAN-composer LLM invented DOM ids for assertion targets despite an explicit "never invent a selector" instruction — the rule held for controls it acts on and failed for things it asserts on. First instance corrected by a guarded supersede migration that itself initially refused a valid runtime selector (leading to the anchor rule); second caught automatically by the live Tier-2 sweep and corrected by migration 143. Demonstrates the design stance: hallucination is countered by verification, not prompt escalation.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev34,#rev35,#stage-5-live; 0NN_supersede_xp_curve_plan_selectors(2).sql; RUNBOOK_travelling_docs(38).md#task-3-proven
- **relations:** anchor rule; supersede versioning (correction recorded as a NOTES entry — "the travelling-docs loop applied to itself").
- **verify-later:** xp-curve PLAN v1→v2 chain + its correction note in doc_notes.

<!-- SOURCE: U08_travelling_docs.md -->
### tool-acceptance-agent — Tier 4 self-driving orchestrator
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** First machine acceptance-run note (run bf330ac6, 2026-07-12, "Tier-4 acceptance PASSED — all 3 evaluated checks"); fail path proven live via a controlled reverted test (failed=1, improve_tool_created=true, full teardown verified).
- **what:** An agent (migration 145) closing the loop with zero humans: `ensure_site_record → load_docs → request_browser_run (Kafka await; resolves the tool's deployed URL from pages itself; NO-OP skips without awaiting when the PLAN has no criteria) → judge_acceptance_results → complete`. Judge recomputes the verdict from results: all pass → acceptance-run note; any fail → acceptance-fail note + ONE improve_tool item (criteria embedded as acceptance_test, handler tool-improver); component-less recreated/adopted tools get the note but no item — logged honestly for manual routing. Trigger 087 (dry-run default).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tool-acceptance-agent-built,#tier-4-self-driving,#fail-path-proven; README_summary_paragraph2_for_discussion.md; 087_TRIGGER_tool_acceptance.sh (header)
- **relations:** browser-runner adapter; acceptance iteration loop; continuous sweep.
- **verify-later:** `platform/orchestration/actions/tool_acceptance_actions.go`; agent_definitions row tool-acceptance-agent; migration 145.

<!-- SOURCE: U08_travelling_docs.md -->
### Continuous acceptance — the `tool_acceptance_due` periodic sweep
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Built + migration 146 applied 2026-07-12, but "v1.0.1111 … the continuous sweep is NOT in the binary" (untracked-file trap); "GATE: continuous acceptance activates on the next image built from 83ba9bd4+" (T11, 2026-07-13 — state at unit close).
- **what:** A discovery check that emits one `acceptance_run` work item per active tool with a deployed page and current PLAN criteria, unless a verdict landed within 7 days or a run is open. Design calls: post-creation/post-improve hooks deliberately NOT used (they'd fire before the page redeploys — creation ends at 'planned', improve merely queues a rerender; the sweep only ever sees deployed pages); items emitted straight to `triaged` (acceptance needs no human judgment; `detected` items were observed sitting unswept); priority 90 so acceptance tests the NEW page after builds/rerenders.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#tier-4-continuous,#v1.0.1111; HANDOFF_2026-07-10…md#T10,T11; README_summary_paragraph2_for_discussion.md
- **relations:** tool-acceptance-agent; untracked-file deploy trap; improve_tool cooldown (cancelled items excluded).
- **verify-later:** `discovery_checks/check_tool_acceptance_due.go` in the deployed image; first unattended acceptance-run note.

<!-- SOURCE: U08_travelling_docs.md -->
### Acceptance iteration loop — iterate until criteria pass
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Both halves proven separately (fail path via controlled test 2026-07-12; fix agents write notes); "let a REAL failure flow through to tool-improver and back" still open at unit close.
- **what:** deploy → acceptance run → failing criterion → `improve_tool` item (criterion as `acceptance_test`, bounded by `max_fix_attempts`) → fixer loads PLAN+NOTES first → fix → append note → redeploy → re-run. Criteria hold the bar still across iterations; NOTES stop iterations fighting each other. *Working* = criteria pass. The one link proven only with a synthetic input is a real failure flowing through tool-improver and back.
- **sources:** RUNBOOK_travelling_docs(38).md#§5; PLAN_tool_acceptance_runner(2).md#flow; OVERVIEW_self_verifying_tools.md#autonomous-loop
- **relations:** findings acceptance_test pattern (improvement-loop); tool-improver; continuous sweep.
- **verify-later:** an improve_tool item with source 'acceptance' processed end-to-end by tool-improver.

<!-- SOURCE: U08_travelling_docs.md -->
### Criteria contract v0 (check-type vocabulary + profiles)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** P0 implements 3 of 7 check types; "the composer emitted "action":"select" … a verb the Tier-4 criteria vocabulary must now define" (open).
- **what:** The machine-readable criteria schema: `profiles: [desktop, mobile]`; check types `selector_exists`, `selector_count`, `no_console_errors`, `asset_loads`, `interaction` (fill/click/select steps + expect), `no_horizontal_overflow`, `page_status_ok`. Deterministic only in v0 — no LLM drives the browser. Desktop = Chromium 1366×900; mobile = one stable Playwright device descriptor (emulation first; real devices out of scope). Phasing P0 boot checks → P1 interpreter+mobile → P2 interactions → P3 screenshots (via the existing Backblaze deploy path) → P4 optional LLM-exploratory mode.
- **sources:** PLAN_tool_acceptance_runner(2).md#criteria-contract,#profiles,#phasing; RUNBOOK_travelling_docs(38).md#stage-6
- **relations:** browser-runner adapter (P0); multi-page tool criteria (open question — url_role field).
- **verify-later:** criteria interpreter coverage in run_checks_action.go; whether "select" verb was added.

<!-- SOURCE: U08_travelling_docs.md -->
### Multi-page tool documentation prerequisites
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** RUNBOOK §5.4 "Multi-page prerequisite: preserve-sections re-render + interactivity-aware save guard (pending) before scaling page counts."
- **what:** Multi-page tools add a "Page set & inter-page contract" PLAN section (URLs, shared state keys, data feeds) and may need per-page checks (a `url_role` field). Scaling page counts is explicitly gated on the pending preserve-sections re-render and interactivity-aware save guard.
- **sources:** RUNBOOK_travelling_docs(38).md#§2,#§5; PLAN_travelling_docs(6).md#tool-assurance; PLAN_tool_acceptance_runner(2).md#open-questions
- **relations:** interactive-section clobber (Part 4) below; criteria contract.
- **verify-later:** save_page_sections interactivity guard deployment status.

<!-- SOURCE: U08_travelling_docs.md -->
### Recreation writes page sections — component-less tools and their visibility gap
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Established by query 2026-07-09 ("pages.sections is EMPTY … the 32 KB game body exists only as deployed HTML in the sites repo"); Tier-2 scope note 2026-07-10.
- **what:** `tool-recreation-handler` ends save_page_sections → update_status → deploy_page and never creates a `content_components` row — adopted/recreated tools exist only as page sections + deployed HTML (source in adoption-crawl research_results: adoption_crawl full markdown+rawHTML, adoption_page per-page; `spec.mode="recreate"` is the handshake set by apply_adoption_plan). Consequences: no component address for tool-improver; invisible to Tier 2 by construction (Tier 4 reaches them via pages); NOTES subject must be pipeline-scoped. `site_plan_sections` is site-plan STRUCTURE, not HTML.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev41,#rev42,#rev43; HANDOFF_2026-07-09…_1_.md#§4; RUNBOOK_travelling_docs(38).md#task-5-record
- **relations:** dangling-doc rule; adoption pipeline (007); Tier-2 scope limit.
- **verify-later:** tool-recreation-handler workflow steps; research_results result_types.

<!-- SOURCE: U09_adoption.md -->
### Tools/games behavioural QA loop (planned)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "PLAN_tools_games_behavioral_qa_loop.md (this session) — a standalone QA/maintenance loop that builds out the planned-but-unbuilt Tier 3 (headless behavioral testing) and adds a games lifecycle… Phased; first cut Phase 0+1" (HANDOFF_2026-06-06).
- **what:** A standalone QA loop for deployed interactive tools/games, motivated by real defects (Jelly Invaders degrading over time, P2P host replies not reaching mobile clients, untested cross-browser/mobile variants). Referenced from the adoption thread as FUTURE work; the plan doc itself lives elsewhere.
- **sources:** HANDOFF_2026-06-06#future, HANDOFF_2026-06-09#later-parked
- **relations:** tool-recreation-handler output quality (Family I1); 019/020 tool library/lifecycle Tier 3
- **verify-later:** PLAN_tools_games_behavioral_qa_loop.md (outside this unit)

<!-- SOURCE: U09_adoption.md -->
### Validation observability: structured rejection logging (recordValidationRejection)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Validation observability deployed: store_generated_component_action.go recordValidationRejection writes a structured agent_error_log row on every pre-store rejection… guide-list's attempt-1 failure was captured exactly — one SQL row, no pod-log forensics" (Session 5, 2026-05-11).
- **what:** Every component pre-store validation rejection writes a structured agent_error_log row (severity warning for bookkeeping vs error for structural; orphan/unknown field names as typed JSONB arrays), replacing pod-log forensics. Companion pattern: the retry budget of 3 is calibrated for the single-bookkeeping-orphan failure class seen in Tier-D regens (tool-list missed card_link_label, guide-list read_guide_label); a central label registry would prevent the class entirely (idea, not built).
- **sources:** FOCUS_directory_builder_and_list_components.md#tier-d-converge
- **relations:** component-creator; chrome-template gate (would reuse the same gate/log shape)
- **verify-later:** store_generated_component_action.go recordValidationRejection

<!-- SOURCE: U12_docs024_archives.md -->
### Mandatory minimum tool-suggestion count (2–5, no "suggest zero" option)
- **category:** tool-lifecycle
- **status-signal:** superseded
- **status-evidence:** Archive: "It returns 2–5 suggestions." Live: "It can return 0-5 suggestions. Returning zero is correct when no tools are appropriate."
- **what:** The earliest `tool-suggester` design forced the LLM to always propose at least two tools per site. Replaced by an explicit zero-is-valid design, directly tied to the same failure class as `matchToolToSite` (irrelevant tools forced onto sites).
- **sources:** old/older1/012_tool_lifecycle_guide.md#"Agent: tool-suggester"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Agent: tool-suggester"
- **relations:** tag-based deterministic tool-to-site matching (above)
- **verify-later:** check tool-suggester's current prompt for the zero-suggestions instruction.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool acceptance runner (headless-browser acceptance testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: initial plan (P0 not started)." (tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md, header)
- **what:** Tier-4 rung of the tool verification ladder: a `browser-runner-adapter` (Playwright+Chromium, mirroring analyser-adapter) drives a deployed tool page under desktop+mobile profiles, judges declared criteria (selector_exists, no_console_errors, asset_loads, interaction, no_horizontal_overflow, page_status_ok) pass/fail, feeding failures back as `improve_tool` work items until criteria pass. Criteria live in the tool's travelling PLAN as a criteria block.
- **sources:** tools/tool_acceptance_runner/PLAN_tool_acceptance_runner.md#Aim, #Criteria-contract, #Phasing
- **relations:** Behavioral QA loop for tools & games (this is the deterministic v0 layer); tool-lifecycle (020); a recent repo commit ("browser-runner-adapter: commit the full Tier-4 adapter") suggests adapter code may already exist — verify
- **verify-later:** `browser-runner-adapter` deployment; `tool-acceptance-agent` orchestrator; `max_fix_attempts` convention

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Tool widget clobber mechanism (M1: DELETE+INSERT rebuild wipes side-written components)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "M1 (clobber) is a confirmed latent defect that does not explain these pages but would bite once M2 is fixed. Fixes drafted, not implemented." (PLAN_tool_widget_clobber(9).md)
- **what:** `save_page_sections_action.go` rebuilds a page's `page_components` by `DELETE FROM page_components WHERE page_id=$1` then re-INSERTs only the sections the content writer supplied. Any side-written row not in that list — including a tool/game widget inserted by `create_tool_component`/`deploy_tool` at position 2 — is destroyed on the next `needs_content_page` build. A content-regression guard exists but compares only visible text length after stripping tags, so it is structurally blind to script-heavy widgets. Old content is snapshotted to `page_component_history` before delete, so wipes are recoverable/detectable.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.1-2.2,#3,#7, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-1
- **relations:** Two divergent tool-creation paths; site_plan as authoritative build source; Canonical tool-page section-shape design question; Recreation-loss defect
- **verify-later:** `save_page_sections_action.go` regression guard/delete lines; `page_component_history` rows with `source='save_page_sections_overwrite'`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Two divergent tool-creation paths (novel vs fork)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** table in §2.3 of PLAN_tool_widget_clobber(9).md documents both paths as existing, currently-running code
- **what:** `create_tool_component_action.go` (the "novel" path) never sets `pages.sections`, leaving it default `[]`; `deploy_tool_action.go` (the "fork" path) sets `pages.sections` to `["hero-tool","tool-guide-intro","<toolFunction>","tool-cta"]`. Both side-write the widget into `page_components` at position 2 and queue `needs_content_page`. The novel path is more exposed to the clobber mechanism since the widget is a member of no section list anywhere.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.3,#9
- **relations:** Tool widget clobber mechanism (M1); Canonicalise tool page identity (T3)
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`; `idx_cc_tool_function_unique` partial index behaviour

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Post-adoption detection check (T2 — check_tool_recreation_needed.go)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2: "T2 — check_tool_recreation_needed.go ... Deployed. Backfills automatically on next discovery run, if recreation works."
- **what:** A new `discovery_checks` package check: finds `page_type IN ('tool','game')`, `status='active'` pages with no widget, sources `interactive_features` from adoption findings via the same canonical-name transform as T1, and emits `needs_tool_recreation` (7-day per-page cooldown). Pages with no captured features are surfaced but deferred to the tool-suggester/generation path. Doubles as the backfill mechanism for pre-existing widget-less pages.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2
- **relations:** Adoption interactivity misroute (T1); check_tool_health blind spot
- **verify-later:** `check_tool_recreation_needed.go`; `idx_swi_dedup`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Recreation-loss defect (correctly-routed recreation still produces no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "Not confirmed: that widgets now actually deploy... correct routing → completed recreation → no widget... Hold the trigger." (HANDOFF §1,§3)
- **what:** Query K showed all five games on gamesdesign.co.uk — which had routed correctly to `tool-recreation-handler` all along and whose recreation work items completed — had no deployed widget component and no inline `<script>` section. So the routing fix (T1) is necessary but not sufficient; something downstream prevents the widget from landing. Diagnosis was interrupted mid-investigation when a parallel adoption chat reset the underlying state, so the exact mechanism remains open.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§2.9,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§5
- **relations:** Tool widget clobber mechanism (M1); tool-game-* duplicate pages; Post-adoption detection check (T2)
- **verify-later:** re-run queries R1-R3/L/M/N1/N2 against current gamesdesign.co.uk state; check `page_component_history` for a clobber snapshot on a game page

<!-- SOURCE: U13_docs024_small_dirs.md -->
### tool-game-* duplicate pages (T5)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "T5 — tool-game-* duplicate pages (step 8) ... Pending re-observe ... May have been wiped by step-9 reset" (PLAN_tool_widget_clobber(9).md §5b)
- **what:** Five `page_type=tool`, `build_status=planned` pages surfaced that duplicate the five existing games by name (`tool-game-<name>`). Candidate mechanisms: `tool-recreation-handler` building a separate page instead of populating the original interactive page, or a planner/reconciler role-divergence in the `029` canonicalisation family. The duplicates vanished in the step-9 state reset before their origin could be confirmed.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §2.8,§5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §3,§6
- **relations:** Recreation-loss defect; Canonicalise tool page identity (T3)
- **verify-later:** query M (who created tool-game-* pages) re-pointed at current state

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonicalise tool page identity across surfaces (T3)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "T3 ... Open, independent ... Low risk; can land at any time." (HANDOFF §6)
- **what:** `create_tool_component` and `deploy_tool` build page name/url/page_type ad hoc, diverging from the canonical `datahelpers.CanonicalisePage` helper that adoption and the planner already use. Proposed fix: route both tool actions through the same canonical helper. Flagged as a gap in `029`'s Phase-0 deliverable list, which covered only two other files.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §5b, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §2,§6,§8
- **relations:** Two divergent tool-creation paths; tool-game-* duplicate pages
- **verify-later:** `create_tool_component_action.go`, `deploy_tool_action.go`, `datahelpers/page_canonical.go`

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Canonical tool-page section-shape design question and fix options
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Fix options (structural-first; not yet implemented)" (PLAN_tool_widget_clobber(9).md §5)
- **what:** Raises and answers (as a design decision, not yet built) whether a tool page even wants generic hero/guide-intro/CTA sections, or just the widget. Three options recorded: (1) make the widget a first-class section in whichever authority the build reads; (2) right-size the tool page's canonical section list; (3) make `save_page_sections` structure-aware as a safety net. Recommended: 1+2 together with 3 as a guard. Notes `content_guidance` already instructs the writer not to regenerate the widget, but the writer has no mechanical way to honour that.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §4,§5
- **relations:** Tool widget clobber mechanism (M1); site_plan as build authority; content-governance
- **verify-later:** whether `plan_sections_action.go` now emits a tool/embed section for `page_type='tool'` pages

<!-- SOURCE: U13_docs024_small_dirs.md -->
### check_tool_health INNER JOIN blind spot
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "check_tool_health blind spot. Its INNER JOIN content_components → page_components means a tool with no linked page_components row ... is invisible" (PLAN_tool_widget_clobber(9).md §8)
- **what:** The Tier-1 tool health check joins `content_components` to `page_components` with an INNER JOIN, so a `page_type='tool'` page with zero linked components (post-clobber, or never-generated) is invisible to it and the check silently reports "no tools" as a pass. T2 partially closes this by detecting the same condition independently, but the original check itself was not corrected.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8, tools/tool_widget_clobber/HANDOFF_2026-05-26_tool_routing_fix_deployed.md §9
- **relations:** Post-adoption detection check (T2); Recreation-loss defect
- **verify-later:** `check_tool_health.go` join logic

<!-- SOURCE: U13_docs024_small_dirs.md -->
### forked_from NULL collision risk on novel tools
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** "forked_from NULL on novel tools ... Two sites generating the same function would collide. Latent; not today's bug." (PLAN_tool_widget_clobber(9).md §8)
- **what:** `create_tool_component` omits `forked_from`, so novel/generated tools are classified as library tools by the partial unique index `idx_cc_tool_function_unique (function) WHERE component_level='tool' AND forked_from IS NULL AND is_active`. Two different sites independently generating a tool with the same function name would collide.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md §8
- **relations:** Two divergent tool-creation paths
- **verify-later:** `idx_cc_tool_function_unique` definition; whether any collision has actually occurred

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Behavioral QA loop for tools & games (Tier 3+ headless-browser testing)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Status: proposed (2026-06-06)." (PLAN_tools_games_behavioral_qa_loop.md, header)
- **what:** A standalone, slower-cadence QA loop that runs generated tools/games in an isolated multi-engine Playwright pod over time under synthetic drive, to catch defect classes invisible to a single render/screenshot: temporal degradation, cross-browser divergence, mobile-specific layout/touch bugs, and multi-context networked/relay failures. Correctness judged via a three-layer oracle: generic deterministic invariants, type-specific assertions, and LLM-as-judge over a screenshot/video series — with auto-fix gated to high-confidence deterministic findings only. Reuses the existing check→work-item→improver pipeline.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md#1-Why,#4-The-headless-pod,#5-The-oracle-problem,#10-Phasing
- **relations:** Tool acceptance runner (this loop is the heavier behavioral/temporal successor); Games quality lifecycle parity; tool-lifecycle (020)
- **verify-later:** whether any phase has been built; `qa_runs`/`last_qa_at` storage location

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tool-doc header rollout (provenance + stripped headers)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** thin_slice(27) "Tool-doc header rollout (2026-06-11) — apply order is load-bearing. … Three stages; do not reorder — the gate without the prompt fails every generation, and the stamps without the columns fail every insert." No completion claim in this unit's files.
- **what:** Rollout procedure for tool documentation headers: (1) provenance columns on content_components (source_agent_type, source_orchestration_id), (2) anchored idempotent prompt updates adding the `=== tool-doc ===` header requirement (abort if prompts drifted), (3) one binary release (tool_doc_header.go + five action edits) so headers are stamped in the DB template but STRIPPED from shipped pages/CDN assets, with a tool_health no_doc_header WARNING converging old tools on the normal sweep — no retrofit campaign.
- **sources:** docs019/RUNBOOK_thin_slice(27).md#tool-doc-header-rollout
- **relations:** doc_plans/doc_notes (the tools thread's later system); tiered tool acceptance
- **verify-later:** content_components source_% columns; '=== tool-doc ===' in html_template rows; tool_health sweep items

<!-- SOURCE: U14_docs019_runbooks.md -->
### Tiered tool acceptance (static contract check + browser-runner)
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** diagnosis_fix_loop(9) "Their Stages 5–6 define a TIERED ACCEPTANCE system for tools: a static Tier-2 contract-presence check and a Tier-4 browser-runner adapter (Chromium+Playwright, Kafka request/response per the 035 Adapter Guide) — their 'loop for complicated tools' is acceptance/verification + docs, NOT a rival diagnosis loop."
- **what:** The tools thread's acceptance ladder for generated tools, recorded here as a shared component: Tier-2 static contract-presence verification and a Tier-4 browser-runner adapter executing tools in real Chromium — also earmarked as a future verification service for fix-loop F1 fixes touching pages and as a council reviewer's instrument.
- **sources:** docs019/RUNBOOK_diagnosis_fix_loop(9).md#what-already-exists; docs019/RUNBOOK_diagnosis_fix_loop(9).md#collision-surface
- **relations:** council of reviewers; tool pipeline; adapters (035 guide)
- **verify-later:** browser-runner adapter existence; tool acceptance stages in the tools thread docs

<!-- SOURCE: U15_docs019_running_notes.md -->
### Tool-doc header system
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** "DONE 2026-06-11: 019/020 CONSOLIDATED — splice files retired" (principles(59)); status marked apply-ready and untouched in v2(36) small-pending list.
- **what:** A standardised 6-12 line sentinel-delimited header block written into every generated tool's script (purpose, behavioural invariants, no-external-calls, version marker) at creation time, stripped at deploy-assembly (three call sites: single-page rerender, `collectJSAssets`, bulk rerender) so it never ships to visitors but is retained in the DB `html_template` for audit/parse parity. Enforced via a hard `HasToolDocHeader` gate in `create_tool_component`, tool-generator/tool-improver prompt edits, and two new `tool_health` tier-1 checks (`no_doc_header` warning, `malformed_doc_header` error). Paired with new `source_agent_type`/`source_orchestration_id` provenance columns on `content_components`, mirroring `knowledge_base`'s existing provenance pair.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tool-doc entries (multiple DONE items).
- **relations:** JS content separation contract; doc claim-verification convention; canonical-doc-home discipline.
- **verify-later:** Whether the rollout (provenance migration → prompts SQL → binary release) was ever applied — repeatedly flagged as "apply-ready, not yet applied" across all later notes files through 2026-07-06.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Fork-divergence detection for library tools
- **category:** tool-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "IMMEDIATE WIN INSTEAD: FORK-DIVERGENCE detection — pure SQL discovery check (tier-1, zero cost)" (principles(59)).
- **what:** A proposed zero-cost SQL discovery check comparing a deployed fork's `html_template` hash against its `forked_from` library original to answer "which forks are unmodified / safe to bulk-push a library change" — deliberately deferred building full code-symbol indexing of tools (each tool is one IIFE, thin symbol pickings; tool discovery already solved via `semantic_tags`) until a concrete consumer needs it.
- **sources:** NOTES_running_synthesis_principles(59) 2026-06-11 tools/provenance/docs design entry.
- **relations:** Tool-doc header system; JS content separation contract.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Tool page missing widget (M1 clobber vs M2 misroute)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** 016 addendum(4) "RESOLVED 2026-05-26 → b1 … key the feature map by the canonical name in buildPageFeatureMap"; companion PLAN_tool_widget_clobber.md
- **what:** A `page_type='tool'` page rendering a description but no widget has two causes needing different fixes: M1 clobber (`SavePageSectionsAction` deletes page_components and its content-regression guard can't see a script-heavy widget) vs M2 never-generated (adoption recreate has no parse stage). For gamesdesign, root cause was a misroute: `buildPageFeatureMap` keys by raw page name while the route looks up canonicalised (`tool-`-prefixed) names.
- **sources:** WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#root-cause-m2-corrected-after-verification, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#diagnostic-recipe-read-only-30-seconds, WM/016_debugging_guide_addendum_adopted_tools_no_widget(4).md#potential-solutions
- **relations:** CanonicalisePage; adoption parse-stage; site plan reconciler; interactive content
- **verify-later:** buildPageFeatureMap; tool-recreation-handler; SavePageSectionsAction; PLAN_tool_widget_clobber.md

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool quality tiers: tool-auditor (Tier 2 LLM review), tool-improver, acceptance checks (Tier 2 static) and tool-acceptance-agent (Tier 4 browser runs)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 088 (tool-auditor); 142 enables tool_acceptance check (2026-07-10, doc_notes entry "unit tests... green"); 145 inserts tool-acceptance-agent; 146 makes Tier 4 continuous via tool_acceptance_due sweep.
- **what:** Layered tool verification. Tier 1: check_tool_health structural checks. Tier 2 (LLM): tool-auditor reads full HTML/CSS/JS and reasons through logic/mobile/UX/accessibility, creating improve_tool or needs_human_review items. Tier 2 (static): check_tool_acceptance asserts the PLAN's criteria fence against the deployed page under the ANCHOR RULE ("validate a selector's leftmost id/class token, never the whole path; confirm, never refute; -EDIT ids skipped"). Tier 4: tool-acceptance-agent drives the deployed tool in headless Chromium via the browser-runner adapter against PLAN criteria — "the tier that turns 'deployed' into 'works'" — pass → acceptance-run note; fail → acceptance-fail note + one improve_tool item carrying criteria as acceptance_test. tool-improver executes improve_tool fixes. 7-day cooldowns; cancelled items excluded from cooldown (146).
- **sources:** 088_tool_auditor_agent.sql; 142_enable_tool_acceptance_check.sql; 145_tool_acceptance_agent.sql; 146_enable_tool_acceptance_due.sql; 062_tool_suggester_and_improver.sql
- **relations:** travelling PLAN criteria fences; design-discovery-agent hosts the checks; browser-runner adapter
- **verify-later:** request_browser_run / judge_acceptance_results actions; check_tool_acceptance.go anchor rule; browser-runner adapter deployment

<!-- SOURCE: U18_sql_for_agents.md -->
### Acceptance-criteria honesty: invented selectors and inline-delivery decisions
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 136 (2026-07-09) repairs the first machine-written PLAN's invented ids (#xpTableBody→#tableWrap tbody, #statsStrip→#statRow); 143/144 (2026-07-10) "PLANs surrender to delivered reality" — asset extraction "was designed but never built", so criteria drop asset_loads and the composer prompt is corrected.
- **what:** Two recurring failure classes in machine-written acceptance criteria, and their durable remedies: (1) composers invent selectors they ASSERT on even while obeying never-invent for controls they ACT on — remedy is Tier-2 static validation of criteria selectors against html_template (anchor rule), not sterner prompts; (2) criteria must describe what the system DELIVERS, not aspirations — the /tools/assets/<fn>.js extraction path was never built, all JS ships inline, so PLANs and the composer prompt were superseded to inline delivery ("born honest"). Also note the abandoned mechanism: Path-1 tool asset extraction on rerender.
- **sources:** 136_supersede_xp_curve_plan_selectors.sql; 143_supersede_plans_inline_delivery.sql; 144_composer_inline_delivery.sql; 113_site_asset_renderer.sql (the extraction design it contradicts)
- **relations:** travelling docs supersede pattern; tool acceptance tiers
- **verify-later:** whether asset extraction ever ships (would trigger forward supersede)

<!-- SOURCE: U19_sql_tables_components.md -->
### Component quality tracking (0–100 score)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** "None of these fields are required by the existing pipeline — they are additive... selector will use them when present and ignore when NULL" (005 ~9848).
- **what:** Additive quality fields on content_components computed by a compute_component_quality action, with indexes for auditor queries (below threshold OR unscored) and planner preference (higher quality per function). Distinct from avg_quality_score in the selector metadata set.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#component-quality-tracking
- **relations:** component selector metadata; improvement loop auditors.
- **verify-later:** compute_component_quality action in registry; populated quality_score values.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component versioning (component_versions)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** Table created in schema-mode migration (008 PART 3); page_components.component_version_id exists in live dump with comment "if versioning enabled".
- **what:** Versioned snapshots of component templates (html_template, css_template, input_schema per version_number) so strict-mode pages could pin a specific template version. Referenced as an optional backup target in later template-fix migrations; unclear whether any writer maintains it.
- **sources:** docs/agent_docs/sql_for_tables/008_page_components_and_schema_mode.sql#PART3; docs/agent_docs/sql_for_tables/005c_bk_page_components.sql
- **relations:** schema-mode subsystem (abandoned); site_plan_sections.component_version_id (planner provenance).
- **verify-later:** row count in component_versions; writers in Go.

<!-- SOURCE: U19_sql_tables_components.md -->
### Tool library fork-on-deploy model
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** forked_from column, partial unique index on function scoped to canonical tools, and the later constraint amendment "Forks (forked_from IS NOT NULL) are excluded from the uniqueness check" fixing the add_tool failure on gamedesign.uk.
- **what:** Library tools are canonical rows (component_level='tool', forked_from IS NULL); deploying to a site copies the row as a fork (forked_from = library id) referenced by page_components. Library changes never cascade to forks; fleet updates go through per-site work items. Uniqueness of `function` applies only to active canonical tools so many site forks can share a function; forks are only ever addressed by component_id.
- **sources:** docs/agent_docs/sql_for_tools/002_tool_migration.sql; docs/agent_docs/sql_for_tables/005_content_components.sql#fork-constraint-fix; docs/agent_docs/sql_for_tables/005b_bk_content_components.sql#idx_cc_tool_function_unique
- **relations:** component library; seeded tool library; improvement-loop fleet updates.
- **verify-later:** deployer fork-copy code; fork counts per library tool.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component regeneration in place (store_generated_component mechanics)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** 083 result: brief-explanation updated in place (same id, created_at unchanged, status 'regenerated', component_versions snapshot, needs_rerender raised) — "matches the documented behaviour (003 §348)".
- **what:** store_generated_component looks up an existing component by the LLM's EMITTED `function` (forked_from IS NULL); if found, it snapshots the current row to component_versions (MAX+1), UPDATEs in place (component_id preserved → all page/site FKs keep resolving), sets template/schema/js_content/render_mode/is_active, then markPagesPendingRebuild raises ONE needs_rerender per affected site. Determinism hazard: regeneration keys on the emitted function name — an unpinned LLM can emit a different name and INSERT a stray duplicate (the 081 'general-hero' incident); pin the function in the description. Pre-store validation rejects `<no value>` templates and checks placeholder/schema parity.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-06-30-~18:35 + #2026-07-01-~12:46; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-e
- **relations:** shared library guard; component-quality-auditor; call_agent contract validation (the trigger saga)
- **verify-later:** store_generated_component_action.go lookup + snapshot + markPagesPendingRebuild; component_versions rows

<!-- SOURCE: U23_docs_root_vonc.md -->
### component-quality-auditor auto-regeneration threshold
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** Read from its default_config 2026-06-29: creates needs_component_regeneration items only for quality_score < 50, handler component-creator, spec {function, component_id, quality_score, quality_issues}.
- **what:** The auditor raises regeneration work items for low-quality components — but its strict `< 50` condition meant the three vonc shells scoring EXACTLY 50 were never auto-picked-up (explaining zero queued items and requiring manual triggers). Its item shape confirms the designed regen path keys on function and routes to component-creator. Boundary-condition gap worth a rule review; also the future home of the autonomy plan's maintenance detections.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~21:00; docs/PLAN_dynamic_sections_and_loaders(4).md#maintenance
- **relations:** component regeneration in place; autonomous section composition (auditor rules gap 4)
- **verify-later:** component-quality-auditor default_config condition; quality_score distribution at exactly 50

<!-- SOURCE: U23_docs_root_vonc.md -->
### Store-path template validation (+ pending <script>-balance hardening)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** Existing checks confirmed in code 2026-06-29 (`<no value>` rejection, placeholder/schema parity, unclosed `<style>`, section/div presence); the `<script>` balance check + separateInlineJS truncation warning remain "STILL MISSING" / backlog item 2 on 2026-07-09.
- **what:** store_generated_component's pre-store validation gate rejects Mode-A/B artifacts and unclosed `<style>` but NOT an unclosed `<script>` — the gap that let provocation-card ship a truncated inline script that swallowed the page footer at render. Hardening definition: add a `<script>` open/close balance check (reject or flag-for-regeneration) plus a truncation warning in separateInlineJS. Prevents the class "truncated template ships and breaks the page".
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-06-29-~20:30 + #2026-07-03-~13:25 (hardening def); docs/RUNBOOK_phase2_provocation_js(29).md#appendix-g; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** legacy un-extracted shells (the truncation instance); Mode A/B taxonomy
- **verify-later:** store_generated_component_action.go validation block for script balance

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Tool widget clobber (save_page_sections DELETE+INSERT destroys widgets)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.2: `SavePageSectionsAction` does `DELETE FROM page_components WHERE page_id=$1` then re-inserts only writer sections; content-regression guard "compares visible text length … structurally blind to tools"; M1 confirmed latent, "Fixes drafted, not implemented".
- **what:** Two writers collide: create_tool_component/deploy_tool side-writes a tool widget into `page_components`, but the authoritative build rebuilds page_components by DELETE+INSERT from the section list (whose authority is `site_plan`, synced into `pages.sections`). The widget isn't in that list, so the first `needs_content_page` build deletes it (snapshotted to `page_component_history` with `source='save_page_sections_overwrite'`). Fix options: make the widget a first-class site_plan section (preferred), right-size the tool page, or make save_page_sections structure-aware.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2-confirmed-findings, #5-fix-options; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md
- **relations:** load_page_sections_from_spec authority; adoption interactivity misroute; live tools/tool_widget_clobber/ set
- **verify-later:** save_page_sections_action.go; load_page_sections_from_spec_action.go; create_tool_component_action.go vs deploy_tool_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Adoption interactivity misroute (canonical prefix desync)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** PLAN_tool_widget_clobber(11) §2.7 "b1 confirmed"; HANDOFF T1 "Deployed": `buildPageFeatureMap` keyed the feature map by raw `fm["page"]` while the routing loop looked up `CanonicalisePage`-canonicalised names (tool branch adds `tool-` prefix), so tool lookups always missed → empty `Features` → static `needs_content_page` route; games (already `game-` prefixed) matched.
- **what:** Adopted tool pages rendered as static description pages because the feature-map key (bare slug) never matched the canonicalised lookup key (with `tool-` prefix). Fixed (T1) by keying `buildPageFeatureMap` on the canonical name resolved from `plan["pages"]`, so routing and content attachment both land in the same namespace. Self-contained one-function change.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.7-resolved, #5b-tasks; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t1
- **relations:** CanonicalisePage (029); tool-recreation-handler; T3 canonicalise create_tool_component/deploy_tool
- **verify-later:** apply_adoption_plan_action.go buildPageFeatureMap; datahelpers/page_canonical.go CanonicalisePage

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### tool-recreation-handler + recreation discovery check (T2)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** HANDOFF T2 "Deployed": `check_tool_recreation_needed.go` (discovery_checks) detects `page_type IN ('tool','game')` active pages with no tool/game component and no inline `<script>`, sources `interactive_features` from adoption findings by canonical name, emits `needs_tool_recreation:<page>` (7-day cooldown). tool-recreation-handler workflow: `recreate_tool`(execute_llm_prompt)→`check_tool_completeness`→`spawn_rerender`→page-rerender.
- **what:** A registered agent that LLM-recreates interactive widgets for pages adoption captured as text-only, plus a discovery check that backfills widget-less interactive pages automatically on the next scheduled run. Item key deliberately distinct from adoption's `needs_page:<name>` to avoid collision.
- **sources:** tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#t2; tools/PLAN_tool_widget_clobber(11).md#5b-tasks
- **relations:** adoption interactivity misroute (T1); recreation-loss defect (T4); check_tool_health blind spot
- **verify-later:** check_tool_recreation_needed.go; tool-recreation-handler agent_definition; check_tool_completeness_action.go

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### Recreation-loss defect (correct routing yields no deployed widget)
- **category:** tool-lifecycle
- **status-signal:** unknown
- **status-evidence:** HANDOFF §3 / PLAN(11) §2.8 (step 8): five games routed `needs_tool_recreation → tool-recreation-handler` and completed (query I), yet all five are widget-less (`has_widget_component=f, has_script_section=f`), plus five new `tool-game-*` duplicate planned pages; step 9 state reset (L/M/N returned 0 rows) left diagnosis incomplete.
- **what:** Even correctly-routed, completed recreation didn't land a deployed widget — a second active defect downstream of routing (T4), blocking. Candidate mechanisms: recreation mis-targeted a parallel `tool-game-*` page, M1 clobber, handler completed without persisting, or a snippets false-negative (inline `<script>` extracted to `/assets/js/snippets.js`). Must be diagnosed before bulk-triggering backfill.
- **sources:** tools/PLAN_tool_widget_clobber(11).md#2.8-blocking, #2.9-state-reset; tools/HANDOFF_2026-05-26_tool_routing_fix_deployed(1).md#3
- **relations:** tool widget clobber (M1); tool-recreation-handler; snippets extraction mechanism
- **verify-later:** tool-recreation-handler recreate_tool→check_tool_completeness→spawn_rerender; page_component_history source values

<!-- SOURCE: U25_leopardess_social.md -->
### Mode-B rendered-artifact templates (components stored as rendered output)
- **category:** tool-lifecycle
- **status-signal:** deployed
- **status-evidence:** VERDICT §2 (2026-07-09): "rendered_html == html_template with all '<no value>' removed — which the byte counts confirm exactly"; "they are rendered outputs stored as source templates".
- **what:** A component corruption class: html_template full of bare `<no value>`, zero {{.}} slots, empty input_schema — the stored template IS a rendered artifact. Consequences: render is a pure function of the template (predictable to the byte — used twice as an acceptance test); content_data is dead weight; repair_template_slots cannot repair them (zero `</no>` tags → needs_regeneration); for runtime-fill shells the emptiness is accidentally exactly what the loader needs, so regeneration must consciously re-establish the empty-shell contract or sections ship with baked copy.
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#2, #8.6; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocation-card(12).md#2026-06-24
- **relations:** runtime-fill guards; component selector/creator (regeneration path); problem-category taxonomy (empty-shell/mode-b-template)
- **verify-later:** components with `<no value>` and 0 schema fields fleet-wide

<!-- SOURCE: U25_leopardess_social.md -->
### Component-creator invocation contract (dual placement + quote-free description)
- **category:** tool-lifecycle
- **status-signal:** partial
- **status-evidence:** NOTES_brief-explanation 083 (2026-07-01) "SUCCEEDED (in-place UPDATE)" via dual placement; framework fix "PATCH_validate_input_contract.go — drafted, not deployed" (HANDOFF §9.3).
- **what:** Manually invoking component-creator (spawn+call) must satisfy BOTH the input_contract (top-level required fields — call_agent validates against top-level extracted fields) AND the workflow's field paths (input_data.spec.*): the working pattern places section_type both top-level and inside spec, pins the function name in the description so the store lands as an in-place UPDATE (else a stray component INSERTs), and keeps the description quote-free to survive the kcat/JSON pipeline. The generic build-dispatch-loop cannot satisfy top-level-required contracts (same class); the durable fix — contract validator accepting top-level OR input_data.spec.{field} — is drafted, not deployed. Regeneration semantics: UPDATE-in-place keyed by function, component_versions snapshot, auto needs_rerender per affected site, store validation rejects `<no value>` templates.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md (081/082/083 arc); docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#8; docs/social001_vonc_tiktok_social/tool_docs/SPEC_provocations-archive-list(1).md#Component-creator-input
- **relations:** component selector/creator; shared component library semantics
- **verify-later:** call_agent contract validation code; PATCH_validate_input_contract.go status

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fork-on-deploy tool ownership model
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019: "This is deliberate. A bad library change shouldn't break ten sites simultaneously."
- **what:** Library tools (component_level='tool', forked_from NULL) are blueprints never referenced by pages; deployment forks a copy per site (forked_from set) and the site owns it — library changes never cascade; pushing improvements to sites is per-site work items. Orphan-fork retry safety: two-stage existing-fork check (P105 fix); GetComponentByFunction excludes forks.
- **sources:** 019#Core Concept; 105 item 6 fix; 020 tool-deployer
- **relations:** tool-improver divergence; component regen (library-level, forked_from NULL)
- **verify-later:** deploy_tool_action.go two-stage check

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Never load html_template in listing queries (storage discipline)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019 heading "Rule: never load html_template in listing or discovery queries" + query audit section
- **what:** Tool/component templates are large; listing and discovery queries must select metadata only, loading html_template only for the specific row being rendered/forked. When to split template from component table is an anticipated (not yet needed) refactor.
- **sources:** 019_tool_library(2).md#Storage and Query Patterns
- **relations:** —
- **verify-later:** listing queries in tool-suggester load_library_tools

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component selector / creator: section_type vs function split
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** 007 Phase 3 items; component-creator live (016b incidents reference it; selection metadata columns specced); selector "integrates into plan_sections as a fallback path"; 036 FINDING: current resolution is direct function lookup, scorer "not exercised"
- **what:** Splits "what role does this section play" (section_type) from "which template" (function). Planner emits section_types; a scoring selector (suitable_site_types/page_types, content_shape, visual_density, usage_count, avg_quality_score, created_from) picks the variant; no candidate → needs_new_component work item → component-creator generates against the full component contract prompt and stores with metadata; quality feedback loop from auditor scores creates a fitness landscape. Backward compatible: direct function lookup remains path 1.
- **sources:** 007#Component Selector and Creator; 036 §7 (scorer not on path)
- **relations:** component regeneration; component creation contract prompt
- **verify-later:** section_type/selection-metadata columns exist; selector wired in plan_sections?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component-creator agent (observed-pattern section components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** context-aware generation deployed 2026-04-17 (reads mission_brief/design_intent/content_direction; max_tokens 16000); regeneration workflow path noted missing ("component-creator only handles needs_new_component") 2026-04-17
- **what:** Generates new section component templates (hero, feature-grid, etc. — distinct from tool-generator) when a page build meets an unfamiliar section type; prompt carries the full component contract and tiered field classification. Known gap at the time: no delete-old→create-new→rerender regeneration path for quality-auditor findings; StoreGeneratedComponentAction later gained a create-OR-regenerate path (Track 2, 2026-04-20) but not deactivated-row resurrection (unique-name collisions need ad-hoc DELETE).
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #Pending; HANDOFF_2026-04-20_error_investigations.md#historical
- **relations:** quality tracking; validation gates; LLM reliability
- **verify-later:** component-creator workflow; store_generated regen path today

<!-- SOURCE: U05_content_quality_linking.md -->
### system-stats component key-contract break (regen renames fields, dependents empty)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "concluded for gamesdesign (closed-by-removal); shared-component fix OUT OF SCOPE … flag the platform bug to its owners".
- **what:** A durable cross-site platform bug found via gamesdesign's empty stats band: component-creator regenerated the shared system-stats component renaming its schema fields (stat_1_number → stat1_value etc.), then re-rendered every dependent from its EXISTING un-migrated content_data — all 5 live instances (across multiple sites) went text-empty in one 16ms batch. The regen mechanism exists but doesn't migrate dependents' content_data on a field rename. Side findings: usage_count is a stale counter (claimed 22, live 5); component_versions is now populated by component-creator (future reverts possible; here only one version existed — no revert target); a concurrent-chat co-management protocol was applied (freshness probes before any shared-component write).
- **sources:** NOTES(44) 2026-06-24 system-stats sessions; RUNBOOK_gamesdesign_index_rebuild(29).md#part-5; HANDOFF_page_pipeline(11).md#2
- **relations:** sectionHasVisibleContent filter (correctly hid the empty shell); writer↔component-schema binding; component regeneration flow (026).
- **verify-later:** content_components fdd92ad4 current schema; the 5 broken instances on other sites; component-creator regen migration behaviour.

<!-- SOURCE: U05_content_quality_linking.md -->
### Tier-D list components (items array from real pages) vs numbered-flat fabrication
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 13 addendum: migration_game_list_tier_d.sql delivered/applied; Part 14g "pre-check shows game-list_pre_037, guide-list_pre_037, tool-list all query_sourced=t".
- **what:** The list-component contract: a Tier-D component sources `items` from query.pages_where_type:<type> (real realised pages), vs the legacy numbered-flat anti-pattern (game1_title…game6_* all source:llm) that fabricated and duplicated entries. game-list was migrated to tool-list-parity (identical field vocabulary so the writer/merge path treats all lists the same); richness deliberately simplified because pages carries only url/title/meta_description. guide-list was Tier-D but starved by the page_type vocabulary gap.
- **sources:** running_notes_14(26).md#part-13; PLAN_b4_b5_hubs_and_link_resolver(3).md#problem
- **relations:** guide page_type; B4/B5 hub links; component schema contracts (003).
- **verify-later:** game-list_pre_037 input_schema/items source; queryresolve resolvePagesWhereType.

<!-- SOURCE: U08_travelling_docs.md -->
### create_tool_component updates in place by function; unique index covers active library originals
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Side finding rev 33 (same function re-run → one component row, same id); index predicate read 2026-07-07.
- **what:** `idx_cc_tool_function_unique` = UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND is_active=true — uniqueness covers ACTIVE LIBRARY ORIGINALS only (duplicate function rows are forks/inactive versions), and `create_tool_component` updates an existing function in place rather than duplicating. Vindicates function-keyed docs (they span all instances). Also banked: content_components has NO site_id column (site scoping via page_components/site_components only); created_from CHECK {manual,generated,adopted,tool,forked}.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23,#rev28,#rev33; HANDOFF_2026-07-08…md#§7
- **relations:** doc subject convention; provenance columns.
- **verify-later:** pg_indexes indexdef for idx_cc_tool_function_unique.

<!-- SOURCE: U09_adoption.md -->
### Tier-D list components: queryresolve + items-array contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** "v1 (inline resolution in plan_sections + merge in page-content-writer) is what's deployed" (FOCUS_directory_builder); tool-list verified end-to-end Step 3; game-list Tier-D migration "delivered… validated… live on commit" (2026-06-04); guide-list resolving after guide re-type.
- **what:** List/directory/grid components declare one `items` array field with `source: query.<name>` (e.g. `query.pages_where_type:tool`) resolved by the `queryresolve` Go package (registry of concrete SQL queries, hard cap 24/default 12, `status IN ('active','deployed')`) at plan_sections time — a deliberate contract change from doc 003's aspirational "at render time" (no render-time template engine exists). Templates use `{{range .items}}`; the query DSL is code-registered, never LLM-written. The dedicated `directory-builder` agent (doc 002 Phase 2 name) is the deferred v2 — a thin wrapper over the same resolver adding re-triggerability; hybrid chosen over inline-forever or agent-first.
- **sources:** FOCUS_directory_builder_and_list_components.md, migration_game_list_tier_d.sql, FOCUS_component_schema_patterns.md
- **relations:** numbered-flat anti-pattern (what it replaces); Step 2/3 anti-fabrication path; directory-builder v2 (aspirational)
- **verify-later:** queryresolve/queryresolve.go; tool-list (migration 041) and game-list/guide-list schemas in content_components

<!-- SOURCE: U12_docs024_archives.md -->
### Tag-based deterministic tool-to-site matching (matchToolToSite)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive `findMissingTools` matches via site-type/industry affinity plus a "universal tools" carve-out; live `020_tool_lifecycle(2).md` Bug History: "the matchToolToSite function classified security/password/privacy as universal, deploying a password checker to every site (including gas wholesalers). Fixed by removing tag-based matching entirely."
- **what:** The original tool-suggestion mechanism was a deterministic Go function comparing a library tool's `semantic_tags` against a site's type/industry. This produced the documented failure mode and was replaced entirely by `tool-suggester`, an LLM-judgment agent that can suggest zero tools.
- **sources:** old/older1/010_tool_library_guide.md#"Deploying automatically via discovery"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Bug history"
- **relations:** tool-suggester agent; mandatory minimum tool-suggestion count (below)
- **verify-later:** confirm `matchToolToSite` function/code path has actually been removed.

<!-- SOURCE: U12_docs024_archives.md -->
### Planned assets-table template/JS split for large tools (superseded plan)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive: for tools >200KB, split JS into a separate `assets`-table file — "This isn't built yet." Live: "A template/JS split IS built — but for the component-creator pipeline... not via the assets table this section once envisioned."
- **what:** The original plan routed oversized tool templates through the `assets` table/S3 pipeline. What was actually built instead — only for component-creator (games/feeds/explorers), not tools — is a `js_content` column on `content_components` populated by `separateInlineJS()`. Live docs warn against applying this to tools without first fixing two known gaps.
- **sources:** old/older1/010_tool_library_guide.md#"When to split template from component"; docs024_key_docs_latest/019_tool_library(2).md#"When to split template from component"
- **relations:** JS Content Separation Contract (003); component-creator pipeline
- **verify-later:** `SELECT count(*) FROM content_components WHERE component_level='tool' AND js_content IS NOT NULL`.

<!-- SOURCE: U12_docs024_archives.md -->
### Component selector by functional requirement
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)"; no later confirmation found.
- **what:** Proposed capability-based search over `content_components` — finding a component by what it does rather than by name/category — paired with section recipes.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4"
- **relations:** section recipes for adoption; tool-suggester's LLM-judgment matching
- **verify-later:** any capability/tag-based component search implementation in `component_library.go`.

<!-- SOURCE: U16_docs019_design_plans.md -->
### JS tools documentation and provenance gap
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** FOCUS_js_tools_documentation: "Status: flagged 2026-06-09. Not started."
- **what:** The platform's JS tools have no prose docs and no code-symbol provenance; the only documentation is origin history (site/plan specs). Three separated needs: prose documentation (language-agnostic rag path, the main gap), code-symbol provenance (waits on the analyser adapter's JS parser drop-in), origin history (exists, a seed not a substitute). Open: docs' git home, a coverage signal, and whether docs and symbols share a tool identity key.
- **sources:** FOCUS_js_tools_documentation.md
- **relations:** analyser adapter polyglot seam; documentation indexing; tool-doc header contract
- **verify-later:** where JS tool sources live; any tool docs collection

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Known-good solution library
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §5 "(NEW) … proven solutions captured as reusable, parameterised templates"; §6.3 "reuse pgvector + a parameterised-solutions table"
- **what:** A store of proven solutions as reusable parameterised templates, indexed by capability/domain, versioned, carrying the conformance + outcome evidence that justified capture. It is the substrate the cascade's tier-1 reuse draws on — derived success promoted to authored-reusable, gated multi-instance to avoid codifying a fluke.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** reliability cascade; authored-vs-derived context (artifacts→docs arrow)
- **verify-later:** pgvector; a parameterised-solutions table (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### component-creator (LLM component template generation) + CSS variable naming contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 093 definition (needs_new_component handler); 123 patches its prompt with the STRICT RULE on variable names, showing live iteration.
- **what:** Generates reusable HTML component templates from section-type descriptions, storing them in content_components with selection metadata (section_type etc.) for reuse. 123 hardens the prompt: only `--color-{role}` variables from the enumerated list are permitted; invented names like --primary-color "are WRONG and will produce broken output because they are undefined in every deployed stylesheet". Closes the loop with build-site-planner's roadmap rule: unknown roadmap section_types become needs_new_component items handled here.
- **sources:** 093_component_creator.sql; 123_component_creator.sql; 053_build_site_planner.sql
- **relations:** component contracts (003 docs); component-quality-auditor; component selector
- **verify-later:** store_generated_component action; component_selector.go; selection metadata columns

<!-- SOURCE: U18_sql_for_agents.md -->
### component-quality-auditor (library health scoring)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 102 definition + one-shot backfill item scoring every existing component.
- **what:** Periodically scores content_components via compute_component_quality and creates needs_component_regeneration items for low scorers — keeps the shared component library healthy rather than only fixing per-site instances.
- **sources:** 102_component_quality_auditor.sql
- **relations:** component-creator (regeneration handler); improvement loop
- **verify-later:** compute_component_quality scoring criteria

<!-- SOURCE: U19_sql_tables_components.md -->
### Component library (content_components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Live pg_dump backup (005b) shows full production shape incl. forked_from, is_dark_section, chk_function_kebab_case; 000 dump shows 41 template components pre-070.
- **what:** Single table of reusable renderables: name, html_template, input_schema, `function` (identity), display_name, category, semantic_tags, component_level (site/page/section/element/head/header/footer/tool), render_mode, is_active, is_dark_section, forked_from. Everything the platform renders — sections, headers, footers, heads, tools — is a row here. Seeds added missing section types (hero variants, contact, features, social-proof, cta, about, departments-grid) as the planner LLM demanded them.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql; docs/agent_docs/sql_for_components/007_add_components.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component naming contract; render modes; tool fork model; component selector metadata.
- **verify-later:** content_components table in clients_db; component_renderer/compile_page_sections Go actions.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component selector metadata and scoring
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** ALTER TABLE adds + two idempotent backfill migrations mapping every existing component to a section_type; selector indexes created.
- **what:** Columns that let a selector score components for a slot: section_type (kebab), suitable_site_types / suitable_page_types (JSONB arrays, GIN indexed), content_shape, visual_density (low/medium/high), usage_count (battle-testedness), avg_quality_score (0–1 auditor feedback), created_from (manual/generated/adopted provenance). Backfill maps hero variants → 'hero', page heroes → page purpose, catch-all → function.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#~9100-9700
- **relations:** component library; component quality tracking; site-plan sections resolution.
- **verify-later:** selector Go code reading these columns; non-NULL section_type coverage.

<!-- SOURCE: U19_sql_tables_components.md -->
### Seeded standalone tool library
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Seed INSERTs with full inline templates: AB test calculator, password entropy meter, plus placeholders (favicon generator, bayesian ranking, clip-path builder, meme generator); three finance batches (stamp duty, mortgage affordability, repayment, overpayment, bridging loan, BTL investor, equity release).
- **what:** Canonical interactive tools stored whole in content_components as `<style>+<main>+<script>` with render_mode='standalone' — no template substitution; site head/header/footer are injected by compile_page_sections; CSS uses var(--color-*) so branding comes from the site stylesheet. Finance calculators are self-contained UK-market tools (SDLT bands, amortization schedules, retained-interest bridging maths).
- **sources:** docs/agent_docs/sql_for_tools/001_initial_toolset.sql; docs/agent_docs/sql_for_tools/003_finance_tools_batch1.sql; docs/agent_docs/sql_for_tools/005_finance_tools_batch3.sql
- **relations:** fork-on-deploy model; CSS variable contract.
- **verify-later:** library tool rows and their deployment forks on live sites.

<!-- SOURCE: U19_sql_tables_components.md -->
### system.internal canonical library site
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Migration 025 "Creates the system.internal site for hosting library-level work items"; migration 042 targets guide-list regeneration work items at system.internal.
- **what:** A synthetic site record that owns library-level work (component regeneration, library maintenance) so the ordinary site_work_items/dispatch machinery can operate on the shared component library exactly as it does on a customer site.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components and #migration-042
- **relations:** site_work_items queue; component regeneration via component-creator.
- **verify-later:** sites row for system.internal; work items with that site_id.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Semantic component library with vector embeddings
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Component schema with clip_embedding_vector and `SELECT … ORDER BY (clip_embedding <=> [vector])` queries (docs003); design tokens + S3 asset paths per component; never evidenced as populated — the shipped library became content_components without embeddings.
- **what:** Vision of a Postgres/pgvector library of deconstructed web components: cleaned HTML/CSS with CSS-variable design tokens, behaviour JS modules, screenshots in S3, semantic labels (layout_purpose, funnel_stage), and CLIP embeddings enabling similarity search ("find a hero that feels 'rustic brewery'"). The Librarian was the sole writer; Architect/Publisher queried it.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md#librarian; docs003_firecrawl/README.0124.11_agent_summary.md; docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql (store_component step)
- **relations:** successors: content_components / tool-library + tool registry matching; embeddings idea resurfaces in contextkit (diagnosis-loop).
- **verify-later:** pgvector extension usage; any table with embedding columns from this era.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Intelligent fallback component matching (P1/P2/P3)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** assemble_from_library action implemented per docs (P1 exact function match → P3 generic-text-block) and the Generic Text Block fallback component INSERTed (017); mvp-site-builder ran on it.
- **what:** The site architect resolves each build-plan section against the component library in tiers: P1 perfect function match, P2 similar purpose, P3 generic fallback — guaranteeing the site always builds. Fallback component and base head/CSS components seeded in content_components.
- **sources:** docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** assemble_from_library action; content_components; tool-library matching (live successor).
- **verify-later:** assemble_from_library in registry; fallback rows in content_components.

<!-- SOURCE: U20_legacy_docs_a.md -->
### In-House Forge — content_components with data-function semantic contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** content_components seeded in 017 (Generic Text Block, Document Head with base CSS…); the *current* page-content-writer "loads component DEFINITIONS from content_components (template, input_schema, category, description)" (100_content_page_build_handler_flow.md) — the table survived into the live pipeline.
- **what:** The platform's own component library: rows with name, function (semantic purpose), html_template with {{placeholders}}, input_schema (the content contract), category and semantic tags. HTML slots carry data-function/data-semantic-purpose attributes forming a shared contract: architects build empty containers, the content pipeline independently fills them by function. Directly ancestral to today's component contracts and slot specs.
- **sources:** docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** contracts-and-standards (component contracts/slot specs — live successor); tool-library component library; intelligent fallback.
- **verify-later:** content_components schema now vs then; data-function attributes in current templates.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Shared component library semantics + field-set guard + neutral-base/fork rule
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Guard read in code (store_generated_component L315-335, referencing the fdd92ad4 incident); shared-component clobber check VERDICT 2026-07-04: "no contamination; brief-explanation base is neutral"; rule recorded as standing.
- **what:** Components with `forked_from IS NULL` are a cross-site SHARED library keyed by `function` (brief-explanation is shared by vonc + idea.uk + robot-hands). A deliberate guard blocks regenerations that DROP or RENAME existing fields on a shared component (the fdd92ad4 system-stats incident: an in-place field rename silently emptied every dependent); pure field ADDITION passes. Standing rule derived: regenerate a shared base only for neutral, purely-additive improvements; site-specific voice must FORK (`forked_from = base_id`) — the "deliberate migration" the code prescribes; direct SQL UPDATEs bypass both the guard and component_versions snapshots. An optional multi-site regen gate (`allow_shared_base_regen`) was considered and HELD.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-shared-component-clobber-check + #2026-07-04-verdict + #2026-07-04-lobby-grid-verified (store analysis); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** component regeneration in place; content-governance (voice leakage); section descriptor (fork-vs-base per site should live on the plan)
- **verify-later:** store_generated_component_action.go field-set guard; content_components forked_from usage

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-probe capture component
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "intent-probe INSERTED into the live library (INSERT 0 1) … second run's INSERT 0 0 is the ON CONFLICT idempotency".
- **what:** A NEW content-library section (after STEP-ZERO found nothing reusable among 83 sections) rendering the invited-action page: no-JS HTML `<form>` POST + 1×1 beacon `<img>`, CSS-var theming, Component Input Schema v2. v1 limit: single text-input action (search/freetext kinds); the {{range}}-based categories variant deferred until the renderer's array handling is verified.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_plan(11).md#p3, traffic_probe_running_notes(27).md#2026-06-11-component-live
- **relations:** carries requires-backend tag; hand-instanced for relojistas/wayfaringlondoner
- **verify-later:** content_components row `intent-probe`; intent_probe_component.sql

<!-- SOURCE: U25_leopardess_social.md -->
### Component selector + component creator (self-extending component library)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** component-creator demonstrably runs in production (NOTES_brief-explanation 083 regen 2026-07-01; archive-list creation 2026-07-06); the selector scoring/metadata design (suitable_site_types, usage_count, avg_quality_score, fitness landscape) is specified in 003d with a build sequence — deployment state of the scoring path unverified.
- **what:** New site types work without special-casing: the planner outputs section_types (structural need), the component selector queries content_components by section_type scoring site-type match + quality + usage, and a "no suitable component" result raises needs_new_component for the component-creator, which LLM-generates html_template + input_schema under the component contract and stores with selection metadata (created_from='generated', quality NULL, usage 0). Components then compete on audit scores — a fitness landscape where good templates survive and spread; second builds reuse everything.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-solution, #Component-library-growth; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md
- **relations:** component creation contract; component-creator invocation contract; shared component library semantics; tool-lifecycle
- **verify-later:** content_components columns section_type/suitable_site_types/usage_count/avg_quality_score; selector Go function; planner prompt

<!-- SOURCE: U25_leopardess_social.md -->
### Shared component library semantics (field-set guard, neutral base, fork rule)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** NOTES_brief-explanation(5) 2026-07-09: "store_generated_component blocks a regeneration that DROPS or RENAMES a field on a shared component — the guard exists because of the fdd92ad4 system-stats incident"; brief-explanation verified shared across vonc + idea.uk + robot-hands with a neutral base.
- **what:** Components with forked_from IS NULL form a cross-site shared library keyed by function. Rules: regenerating a shared base is safe only for neutral, purely-additive changes (the field guard blocks drops/renames — renaming a field once silently emptied every dependent); site-specific voice must fork (forked_from = base_id); direct SQL UPDATEs bypass both the guard and component_versions snapshotting, so hand edits must snapshot manually and check D4-style "is it really single-site" first.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md#2026-07-09; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #5-D4; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#6.3
- **relations:** component selector/creator; three-per-row rule (shared grids untouchable); per-site style fork (same never-edit-shared principle)
- **verify-later:** store_generated_component field guard; forked_from distribution across content_components

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Fork-on-deploy tool ownership model
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019: "This is deliberate. A bad library change shouldn't break ten sites simultaneously."
- **what:** Library tools (component_level='tool', forked_from NULL) are blueprints never referenced by pages; deployment forks a copy per site (forked_from set) and the site owns it — library changes never cascade; pushing improvements to sites is per-site work items. Orphan-fork retry safety: two-stage existing-fork check (P105 fix); GetComponentByFunction excludes forks.
- **sources:** 019#Core Concept; 105 item 6 fix; 020 tool-deployer
- **relations:** tool-improver divergence; component regen (library-level, forked_from NULL)
- **verify-later:** deploy_tool_action.go two-stage check

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Never load html_template in listing queries (storage discipline)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 019 heading "Rule: never load html_template in listing or discovery queries" + query audit section
- **what:** Tool/component templates are large; listing and discovery queries must select metadata only, loading html_template only for the specific row being rendered/forked. When to split template from component table is an anticipated (not yet needed) refactor.
- **sources:** 019_tool_library(2).md#Storage and Query Patterns
- **relations:** —
- **verify-later:** listing queries in tool-suggester load_library_tools

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Component selector / creator: section_type vs function split
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** 007 Phase 3 items; component-creator live (016b incidents reference it; selection metadata columns specced); selector "integrates into plan_sections as a fallback path"; 036 FINDING: current resolution is direct function lookup, scorer "not exercised"
- **what:** Splits "what role does this section play" (section_type) from "which template" (function). Planner emits section_types; a scoring selector (suitable_site_types/page_types, content_shape, visual_density, usage_count, avg_quality_score, created_from) picks the variant; no candidate → needs_new_component work item → component-creator generates against the full component contract prompt and stores with metadata; quality feedback loop from auditor scores creates a fitness landscape. Backward compatible: direct function lookup remains path 1.
- **sources:** 007#Component Selector and Creator; 036 §7 (scorer not on path)
- **relations:** component regeneration; component creation contract prompt
- **verify-later:** section_type/selection-metadata columns exist; selector wired in plan_sections?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Component-creator agent (observed-pattern section components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** context-aware generation deployed 2026-04-17 (reads mission_brief/design_intent/content_direction; max_tokens 16000); regeneration workflow path noted missing ("component-creator only handles needs_new_component") 2026-04-17
- **what:** Generates new section component templates (hero, feature-grid, etc. — distinct from tool-generator) when a page build meets an unfamiliar section type; prompt carries the full component contract and tiered field classification. Known gap at the time: no delete-old→create-new→rerender regeneration path for quality-auditor findings; StoreGeneratedComponentAction later gained a create-OR-regenerate path (Track 2, 2026-04-20) but not deactivated-row resurrection (unique-name collisions need ad-hoc DELETE).
- **sources:** HANDOFF_2026-04-17_component_rendering_js_separation_quality.md#1, #Pending; HANDOFF_2026-04-20_error_investigations.md#historical
- **relations:** quality tracking; validation gates; LLM reliability
- **verify-later:** component-creator workflow; store_generated regen path today

<!-- SOURCE: U05_content_quality_linking.md -->
### system-stats component key-contract break (regen renames fields, dependents empty)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "concluded for gamesdesign (closed-by-removal); shared-component fix OUT OF SCOPE … flag the platform bug to its owners".
- **what:** A durable cross-site platform bug found via gamesdesign's empty stats band: component-creator regenerated the shared system-stats component renaming its schema fields (stat_1_number → stat1_value etc.), then re-rendered every dependent from its EXISTING un-migrated content_data — all 5 live instances (across multiple sites) went text-empty in one 16ms batch. The regen mechanism exists but doesn't migrate dependents' content_data on a field rename. Side findings: usage_count is a stale counter (claimed 22, live 5); component_versions is now populated by component-creator (future reverts possible; here only one version existed — no revert target); a concurrent-chat co-management protocol was applied (freshness probes before any shared-component write).
- **sources:** NOTES(44) 2026-06-24 system-stats sessions; RUNBOOK_gamesdesign_index_rebuild(29).md#part-5; HANDOFF_page_pipeline(11).md#2
- **relations:** sectionHasVisibleContent filter (correctly hid the empty shell); writer↔component-schema binding; component regeneration flow (026).
- **verify-later:** content_components fdd92ad4 current schema; the 5 broken instances on other sites; component-creator regen migration behaviour.

<!-- SOURCE: U05_content_quality_linking.md -->
### Tier-D list components (items array from real pages) vs numbered-flat fabrication
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 13 addendum: migration_game_list_tier_d.sql delivered/applied; Part 14g "pre-check shows game-list_pre_037, guide-list_pre_037, tool-list all query_sourced=t".
- **what:** The list-component contract: a Tier-D component sources `items` from query.pages_where_type:<type> (real realised pages), vs the legacy numbered-flat anti-pattern (game1_title…game6_* all source:llm) that fabricated and duplicated entries. game-list was migrated to tool-list-parity (identical field vocabulary so the writer/merge path treats all lists the same); richness deliberately simplified because pages carries only url/title/meta_description. guide-list was Tier-D but starved by the page_type vocabulary gap.
- **sources:** running_notes_14(26).md#part-13; PLAN_b4_b5_hubs_and_link_resolver(3).md#problem
- **relations:** guide page_type; B4/B5 hub links; component schema contracts (003).
- **verify-later:** game-list_pre_037 input_schema/items source; queryresolve resolvePagesWhereType.

<!-- SOURCE: U08_travelling_docs.md -->
### create_tool_component updates in place by function; unique index covers active library originals
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Side finding rev 33 (same function re-run → one component row, same id); index predicate read 2026-07-07.
- **what:** `idx_cc_tool_function_unique` = UNIQUE(function) WHERE component_level='tool' AND forked_from IS NULL AND is_active=true — uniqueness covers ACTIVE LIBRARY ORIGINALS only (duplicate function rows are forks/inactive versions), and `create_tool_component` updates an existing function in place rather than duplicating. Vindicates function-keyed docs (they span all instances). Also banked: content_components has NO site_id column (site scoping via page_components/site_components only); created_from CHECK {manual,generated,adopted,tool,forked}.
- **sources:** RUNNING_NOTES_travelling_docs(39).md#rev22,#rev23,#rev28,#rev33; HANDOFF_2026-07-08…md#§7
- **relations:** doc subject convention; provenance columns.
- **verify-later:** pg_indexes indexdef for idx_cc_tool_function_unique.

<!-- SOURCE: U09_adoption.md -->
### Tier-D list components: queryresolve + items-array contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** "v1 (inline resolution in plan_sections + merge in page-content-writer) is what's deployed" (FOCUS_directory_builder); tool-list verified end-to-end Step 3; game-list Tier-D migration "delivered… validated… live on commit" (2026-06-04); guide-list resolving after guide re-type.
- **what:** List/directory/grid components declare one `items` array field with `source: query.<name>` (e.g. `query.pages_where_type:tool`) resolved by the `queryresolve` Go package (registry of concrete SQL queries, hard cap 24/default 12, `status IN ('active','deployed')`) at plan_sections time — a deliberate contract change from doc 003's aspirational "at render time" (no render-time template engine exists). Templates use `{{range .items}}`; the query DSL is code-registered, never LLM-written. The dedicated `directory-builder` agent (doc 002 Phase 2 name) is the deferred v2 — a thin wrapper over the same resolver adding re-triggerability; hybrid chosen over inline-forever or agent-first.
- **sources:** FOCUS_directory_builder_and_list_components.md, migration_game_list_tier_d.sql, FOCUS_component_schema_patterns.md
- **relations:** numbered-flat anti-pattern (what it replaces); Step 2/3 anti-fabrication path; directory-builder v2 (aspirational)
- **verify-later:** queryresolve/queryresolve.go; tool-list (migration 041) and game-list/guide-list schemas in content_components

<!-- SOURCE: U12_docs024_archives.md -->
### Tag-based deterministic tool-to-site matching (matchToolToSite)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive `findMissingTools` matches via site-type/industry affinity plus a "universal tools" carve-out; live `020_tool_lifecycle(2).md` Bug History: "the matchToolToSite function classified security/password/privacy as universal, deploying a password checker to every site (including gas wholesalers). Fixed by removing tag-based matching entirely."
- **what:** The original tool-suggestion mechanism was a deterministic Go function comparing a library tool's `semantic_tags` against a site's type/industry. This produced the documented failure mode and was replaced entirely by `tool-suggester`, an LLM-judgment agent that can suggest zero tools.
- **sources:** old/older1/010_tool_library_guide.md#"Deploying automatically via discovery"; docs024_key_docs_latest/020_tool_lifecycle(2).md#"Bug history"
- **relations:** tool-suggester agent; mandatory minimum tool-suggestion count (below)
- **verify-later:** confirm `matchToolToSite` function/code path has actually been removed.

<!-- SOURCE: U12_docs024_archives.md -->
### Planned assets-table template/JS split for large tools (superseded plan)
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Archive: for tools >200KB, split JS into a separate `assets`-table file — "This isn't built yet." Live: "A template/JS split IS built — but for the component-creator pipeline... not via the assets table this section once envisioned."
- **what:** The original plan routed oversized tool templates through the `assets` table/S3 pipeline. What was actually built instead — only for component-creator (games/feeds/explorers), not tools — is a `js_content` column on `content_components` populated by `separateInlineJS()`. Live docs warn against applying this to tools without first fixing two known gaps.
- **sources:** old/older1/010_tool_library_guide.md#"When to split template from component"; docs024_key_docs_latest/019_tool_library(2).md#"When to split template from component"
- **relations:** JS Content Separation Contract (003); component-creator pipeline
- **verify-later:** `SELECT count(*) FROM content_components WHERE component_level='tool' AND js_content IS NOT NULL`.

<!-- SOURCE: U12_docs024_archives.md -->
### Component selector by functional requirement
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** Listed under "Phase 4 (later)"; no later confirmation found.
- **what:** Proposed capability-based search over `content_components` — finding a component by what it does rather than by name/category — paired with section recipes.
- **sources:** old/older1/FOCUS_design_and_styling_adoption WORK_PLAN.md#"Phase 4"
- **relations:** section recipes for adoption; tool-suggester's LLM-judgment matching
- **verify-later:** any capability/tag-based component search implementation in `component_library.go`.

<!-- SOURCE: U16_docs019_design_plans.md -->
### JS tools documentation and provenance gap
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** FOCUS_js_tools_documentation: "Status: flagged 2026-06-09. Not started."
- **what:** The platform's JS tools have no prose docs and no code-symbol provenance; the only documentation is origin history (site/plan specs). Three separated needs: prose documentation (language-agnostic rag path, the main gap), code-symbol provenance (waits on the analyser adapter's JS parser drop-in), origin history (exists, a seed not a substitute). Open: docs' git home, a coverage signal, and whether docs and symbols share a tool identity key.
- **sources:** FOCUS_js_tools_documentation.md
- **relations:** analyser adapter polyglot seam; documentation indexing; tool-doc header contract
- **verify-later:** where JS tool sources live; any tool docs collection

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Known-good solution library
- **category:** tool-library
- **status-signal:** aspirational
- **status-evidence:** MASTER(4) §5 "(NEW) … proven solutions captured as reusable, parameterised templates"; §6.3 "reuse pgvector + a parameterised-solutions table"
- **what:** A store of proven solutions as reusable parameterised templates, indexed by capability/domain, versioned, carrying the conformance + outcome evidence that justified capture. It is the substrate the cascade's tier-1 reuse draws on — derived success promoted to authored-reusable, gated multi-instance to avoid codifying a fluke.
- **sources:** ED/MASTER_autonomous_build_and_operate(4).md#5, ED/MASTER_autonomous_build_and_operate(4).md#6.3, ED/MASTER_autonomous_build_and_operate(4).md#7.5
- **relations:** reliability cascade; authored-vs-derived context (artifacts→docs arrow)
- **verify-later:** pgvector; a parameterised-solutions table (proposed)

<!-- SOURCE: U18_sql_for_agents.md -->
### component-creator (LLM component template generation) + CSS variable naming contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 093 definition (needs_new_component handler); 123 patches its prompt with the STRICT RULE on variable names, showing live iteration.
- **what:** Generates reusable HTML component templates from section-type descriptions, storing them in content_components with selection metadata (section_type etc.) for reuse. 123 hardens the prompt: only `--color-{role}` variables from the enumerated list are permitted; invented names like --primary-color "are WRONG and will produce broken output because they are undefined in every deployed stylesheet". Closes the loop with build-site-planner's roadmap rule: unknown roadmap section_types become needs_new_component items handled here.
- **sources:** 093_component_creator.sql; 123_component_creator.sql; 053_build_site_planner.sql
- **relations:** component contracts (003 docs); component-quality-auditor; component selector
- **verify-later:** store_generated_component action; component_selector.go; selection metadata columns

<!-- SOURCE: U18_sql_for_agents.md -->
### component-quality-auditor (library health scoring)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** 102 definition + one-shot backfill item scoring every existing component.
- **what:** Periodically scores content_components via compute_component_quality and creates needs_component_regeneration items for low scorers — keeps the shared component library healthy rather than only fixing per-site instances.
- **sources:** 102_component_quality_auditor.sql
- **relations:** component-creator (regeneration handler); improvement loop
- **verify-later:** compute_component_quality scoring criteria

<!-- SOURCE: U19_sql_tables_components.md -->
### Component library (content_components)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Live pg_dump backup (005b) shows full production shape incl. forked_from, is_dark_section, chk_function_kebab_case; 000 dump shows 41 template components pre-070.
- **what:** Single table of reusable renderables: name, html_template, input_schema, `function` (identity), display_name, category, semantic_tags, component_level (site/page/section/element/head/header/footer/tool), render_mode, is_active, is_dark_section, forked_from. Everything the platform renders — sections, headers, footers, heads, tools — is a row here. Seeds added missing section types (hero variants, contact, features, social-proof, cta, about, departments-grid) as the planner LLM demanded them.
- **sources:** docs/agent_docs/sql_for_tables/005b_bk_content_components.sql; docs/agent_docs/sql_for_components/007_add_components.sql; docs/agent_docs/sql_for_tables/000_content_components_backup_070_refactor.sql
- **relations:** component naming contract; render modes; tool fork model; component selector metadata.
- **verify-later:** content_components table in clients_db; component_renderer/compile_page_sections Go actions.

<!-- SOURCE: U19_sql_tables_components.md -->
### Component selector metadata and scoring
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** ALTER TABLE adds + two idempotent backfill migrations mapping every existing component to a section_type; selector indexes created.
- **what:** Columns that let a selector score components for a slot: section_type (kebab), suitable_site_types / suitable_page_types (JSONB arrays, GIN indexed), content_shape, visual_density (low/medium/high), usage_count (battle-testedness), avg_quality_score (0–1 auditor feedback), created_from (manual/generated/adopted provenance). Backfill maps hero variants → 'hero', page heroes → page purpose, catch-all → function.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#~9100-9700
- **relations:** component library; component quality tracking; site-plan sections resolution.
- **verify-later:** selector Go code reading these columns; non-NULL section_type coverage.

<!-- SOURCE: U19_sql_tables_components.md -->
### Seeded standalone tool library
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Seed INSERTs with full inline templates: AB test calculator, password entropy meter, plus placeholders (favicon generator, bayesian ranking, clip-path builder, meme generator); three finance batches (stamp duty, mortgage affordability, repayment, overpayment, bridging loan, BTL investor, equity release).
- **what:** Canonical interactive tools stored whole in content_components as `<style>+<main>+<script>` with render_mode='standalone' — no template substitution; site head/header/footer are injected by compile_page_sections; CSS uses var(--color-*) so branding comes from the site stylesheet. Finance calculators are self-contained UK-market tools (SDLT bands, amortization schedules, retained-interest bridging maths).
- **sources:** docs/agent_docs/sql_for_tools/001_initial_toolset.sql; docs/agent_docs/sql_for_tools/003_finance_tools_batch1.sql; docs/agent_docs/sql_for_tools/005_finance_tools_batch3.sql
- **relations:** fork-on-deploy model; CSS variable contract.
- **verify-later:** library tool rows and their deployment forks on live sites.

<!-- SOURCE: U19_sql_tables_components.md -->
### system.internal canonical library site
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Migration 025 "Creates the system.internal site for hosting library-level work items"; migration 042 targets guide-list regeneration work items at system.internal.
- **what:** A synthetic site record that owns library-level work (component regeneration, library maintenance) so the ordinary site_work_items/dispatch machinery can operate on the shared component library exactly as it does on a customer site.
- **sources:** docs/agent_docs/sql_for_tables/005_content_components.sql#migration-025-library-components and #migration-042
- **relations:** site_work_items queue; component regeneration via component-creator.
- **verify-later:** sites row for system.internal; work items with that site_id.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Semantic component library with vector embeddings
- **category:** tool-library
- **status-signal:** superseded
- **status-evidence:** Component schema with clip_embedding_vector and `SELECT … ORDER BY (clip_embedding <=> [vector])` queries (docs003); design tokens + S3 asset paths per component; never evidenced as populated — the shipped library became content_components without embeddings.
- **what:** Vision of a Postgres/pgvector library of deconstructed web components: cleaned HTML/CSS with CSS-variable design tokens, behaviour JS modules, screenshots in S3, semantic labels (layout_purpose, funnel_stage), and CLIP embeddings enabling similarity search ("find a hero that feels 'rustic brewery'"). The Librarian was the sole writer; Architect/Publisher queried it.
- **sources:** docs003_firecrawl/README.0120.11_agent_website_framework.md#librarian; docs003_firecrawl/README.0124.11_agent_summary.md; docs004_website_capture_project/playwright/website_builder_orchestration_agent.sql (store_component step)
- **relations:** successors: content_components / tool-library + tool registry matching; embeddings idea resurfaces in contextkit (diagnosis-loop).
- **verify-later:** pgvector extension usage; any table with embedding columns from this era.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Intelligent fallback component matching (P1/P2/P3)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** assemble_from_library action implemented per docs (P1 exact function match → P3 generic-text-block) and the Generic Text Block fallback component INSERTed (017); mvp-site-builder ran on it.
- **what:** The site architect resolves each build-plan section against the component library in tiers: P1 perfect function match, P2 similar purpose, P3 generic fallback — guaranteeing the site always builds. Fallback component and base head/CSS components seeded in content_components.
- **sources:** docs004_website_capture_project/website_analysis/README.010.pragmatic_first_steps.md; docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.014.summary_next_steps.md
- **relations:** assemble_from_library action; content_components; tool-library matching (live successor).
- **verify-later:** assemble_from_library in registry; fallback rows in content_components.

<!-- SOURCE: U20_legacy_docs_a.md -->
### In-House Forge — content_components with data-function semantic contract
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** content_components seeded in 017 (Generic Text Block, Document Head with base CSS…); the *current* page-content-writer "loads component DEFINITIONS from content_components (template, input_schema, category, description)" (100_content_page_build_handler_flow.md) — the table survived into the live pipeline.
- **what:** The platform's own component library: rows with name, function (semantic purpose), html_template with {{placeholders}}, input_schema (the content contract), category and semantic tags. HTML slots carry data-function/data-semantic-purpose attributes forming a shared contract: architects build empty containers, the content pipeline independently fills them by function. Directly ancestral to today's component contracts and slot specs.
- **sources:** docs004_website_capture_project/website_analysis/README.017.base_components.md; docs004_website_capture_project/website_analysis/README.011.mvp_content_generation_workflow.md; docs001_flow_general/100_content_page_build_handler_flow.md
- **relations:** contracts-and-standards (component contracts/slot specs — live successor); tool-library component library; intelligent fallback.
- **verify-later:** content_components schema now vs then; data-function attributes in current templates.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Shared component library semantics + field-set guard + neutral-base/fork rule
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** Guard read in code (store_generated_component L315-335, referencing the fdd92ad4 incident); shared-component clobber check VERDICT 2026-07-04: "no contamination; brief-explanation base is neutral"; rule recorded as standing.
- **what:** Components with `forked_from IS NULL` are a cross-site SHARED library keyed by `function` (brief-explanation is shared by vonc + idea.uk + robot-hands). A deliberate guard blocks regenerations that DROP or RENAME existing fields on a shared component (the fdd92ad4 system-stats incident: an in-place field rename silently emptied every dependent); pure field ADDITION passes. Standing rule derived: regenerate a shared base only for neutral, purely-additive improvements; site-specific voice must FORK (`forked_from = base_id`) — the "deliberate migration" the code prescribes; direct SQL UPDATEs bypass both the guard and component_versions snapshots. An optional multi-site regen gate (`allow_shared_base_regen`) was considered and HELD.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-shared-component-clobber-check + #2026-07-04-verdict + #2026-07-04-lobby-grid-verified (store analysis); docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§8
- **relations:** component regeneration in place; content-governance (voice leakage); section descriptor (fork-vs-base per site should live on the plan)
- **verify-later:** store_generated_component_action.go field-set guard; content_components forked_from usage

<!-- SOURCE: U24c_docs_archive_traffic_probe.md -->
### intent-probe capture component
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** running_notes 2026-06-11 "intent-probe INSERTED into the live library (INSERT 0 1) … second run's INSERT 0 0 is the ON CONFLICT idempotency".
- **what:** A NEW content-library section (after STEP-ZERO found nothing reusable among 83 sections) rendering the invited-action page: no-JS HTML `<form>` POST + 1×1 beacon `<img>`, CSS-var theming, Component Input Schema v2. v1 limit: single text-input action (search/freetext kinds); the {{range}}-based categories variant deferred until the renderer's array handling is verified.
- **sources:** traffic_probe_running_notes(27).md#2026-06-10-p3-repo-selection, traffic_probe_plan(11).md#p3, traffic_probe_running_notes(27).md#2026-06-11-component-live
- **relations:** carries requires-backend tag; hand-instanced for relojistas/wayfaringlondoner
- **verify-later:** content_components row `intent-probe`; intent_probe_component.sql

<!-- SOURCE: U25_leopardess_social.md -->
### Component selector + component creator (self-extending component library)
- **category:** tool-library
- **status-signal:** partial
- **status-evidence:** component-creator demonstrably runs in production (NOTES_brief-explanation 083 regen 2026-07-01; archive-list creation 2026-07-06); the selector scoring/metadata design (suitable_site_types, usage_count, avg_quality_score, fitness landscape) is specified in 003d with a build sequence — deployment state of the scoring path unverified.
- **what:** New site types work without special-casing: the planner outputs section_types (structural need), the component selector queries content_components by section_type scoring site-type match + quality + usage, and a "no suitable component" result raises needs_new_component for the component-creator, which LLM-generates html_template + input_schema under the component contract and stores with selection metadata (created_from='generated', quality NULL, usage 0). Components then compete on audit scores — a fitness landscape where good templates survive and spread; second builds reuse everything.
- **sources:** docs/social001_vonc_tiktok_social/003d_spark_strategic_planning_architecture.md#The-solution, #Component-library-growth; docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md; docs/social001_vonc_tiktok_social/tool_docs/NOTES_provocations-archive-list.md
- **relations:** component creation contract; component-creator invocation contract; shared component library semantics; tool-lifecycle
- **verify-later:** content_components columns section_type/suitable_site_types/usage_count/avg_quality_score; selector Go function; planner prompt

<!-- SOURCE: U25_leopardess_social.md -->
### Shared component library semantics (field-set guard, neutral base, fork rule)
- **category:** tool-library
- **status-signal:** deployed
- **status-evidence:** NOTES_brief-explanation(5) 2026-07-09: "store_generated_component blocks a regeneration that DROPS or RENAMES a field on a shared component — the guard exists because of the fdd92ad4 system-stats incident"; brief-explanation verified shared across vonc + idea.uk + robot-hands with a neutral base.
- **what:** Components with forked_from IS NULL form a cross-site shared library keyed by function. Rules: regenerating a shared base is safe only for neutral, purely-additive changes (the field guard blocks drops/renames — renaming a field once silently emptied every dependent); site-specific voice must fork (forked_from = base_id); direct SQL UPDATEs bypass both the guard and component_versions snapshotting, so hand edits must snapshot manually and check D4-style "is it really single-site" first.
- **sources:** docs/social001_vonc_tiktok_social/tool_docs/NOTES_brief-explanation(5).md#2026-07-09; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #5-D4; docs/social001_vonc_tiktok_social/minilobby_task/VERDICT_minilobby_trim_method.md#6.3
- **relations:** component selector/creator; three-per-row rule (shared grids untouchable); per-site style fork (same never-edit-shared principle)
- **verify-later:** store_generated_component field guard; forked_from distribution across content_components

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool pipeline end-to-end (suggest → route → generate/fork → cross-link → rewrite)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 005(1): "Pipeline Status: Fully Operational… all work without manual intervention" with per-site verified results
- **what:** check_missing_tools → tool-suggester (LLM judgement, 0-5 suggestions, library-vs-novel routing via check_is_library) → tool-deployer (fork) or tool-generator (novel) → create_cross_links (content_rewrite items per related page, item_key tool_crosslink:*, tool- pages filtered) → dispatch → page-build-handler threads rewrite_guidance (`input_data.spec.suggestion`) into the writer's nested loop prompt → rerender. The writer prompt lives deep in sub_workflow nesting — top-level jsonb_each misses it (072 trap).
- **sources:** 005_tool_pipeline(1).md full; 020 agents detail
- **relations:** de-tool hazard; fork-on-deploy; tool doc header
- **verify-later:** migrations 070–073 applied; cross-link items in prod

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tools pipeline (suggest / deploy-fork / generate / improve / audit)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "Five agents, two discovery checks, full lifecycle documented … Tier 3 (headless browser visual testing) is planned, not built"; but Path D flags interactive behaviour "reportedly currently don't work" (2026-05-14)
- **what:** tool-suggester (LLM over spec aspects + library, 0-5 suggestions with library_source routing), tool-deployer (library fork with forked_from + tool page + companion guide), tool-generator (novel LLM tool, same wiring), tool-improver (issue-driven rewrite), tool_health Tier-1 structural check + tool-auditor Tier-2 LLM review with confidence-split routing. Missing vs the four-stage pattern: no parse stage (source tools not read), loose source-tool fidelity. Fork-retry idempotency fixed (P2: reuse orphaned forks; GetComponentByFunction excludes forks).
- **sources:** FOCUS_interactive_content_generation(4).md#Tools; HANDOFF-pipeline-triage-april-2026.md P1/P2
- **relations:** games gap (copies this shape); library model; quality model
- **verify-later:** actual tool interactivity failures (Path D); tool_health/tool-auditor definitions

<!-- SOURCE: U08_travelling_docs.md -->
### Tool creation never enqueues the final page deploy (planned-pages gap)
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Gap on record: tool creation ends at `complete` without enqueuing a page_rerender item — the pages deploy only when something else sweeps" (2026-07-10; pages hand-deployed by inserting page_rerender items).
- **what:** tool-generator creates component + page + nav but leaves the page `build_status='planned'`; nothing enqueues the render+deploy hop, so new tool pages 404 until an unrelated sweep. Recorded follow-up: a `create_rerender_item` tail on tool-generator. Interacts with acceptance timing (the reason post-creation acceptance hooks were rejected in favour of the deployed-pages-only sweep).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#both-planned-tool-pages-deployed; HANDOFF_2026-07-10…md#§5.3
- **relations:** continuous sweep design; build/rerender pipeline.
- **verify-later:** whether tool-generator gained a rerender tail.

<!-- SOURCE: U08_travelling_docs.md -->
### Inline-JS extraction ("Path 1" /tools/assets/<fn>.js) — designed, partly real, not on the deploy path
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** 016b v3 entry (separateInlineJS/collectJSAssets exist and are correct for the store path); Tier-2 first sweep proved the deploy path ships JS INLINE and never references the asset (js-not-extracted, Option B superseded the criteria).
- **what:** The store path's `separateInlineJS` extracts a bare inline `<script>` into `js_content`, replaced by a `<script src="/tools/assets/{function}.js">` reference deployed by `collectJSAssets` — but only for attribute-less tags, and legacy/seeded rows predate it (empty shells with raw inline scripts; provocation-card additionally truncated mid-script — store validation checks unclosed `<style>` but not `<script>`). Meanwhile the generator/deploy route for new tools ships everything inline, so "Path 1 extraction" is delivered reality nowhere on that route. Hardening recorded: script-balance check at store time; regenerate broken shells through the current path.
- **sources:** 016b_debugging_guide_7_3_(7).md#js-not-extracted-entry; RUNBOOK_travelling_docs(38).md#stage-5 (pre-verification); PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** delivered-reality principle (Option B); empty-shell/mode-b categories; vonc case evidence.
- **verify-later:** store_generated_component_action.go separateInlineJS; whether extraction ever ships.

<!-- SOURCE: U10_imagery.md -->
### deploy_page files_field contract (co-located JS must ship)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "CRITICAL deploy dependency: page-rerender deploy_page step MUST use files_field:'rendered_page.files' (NOT content_field)… fix was applied during the gas rollout 2026-05-19/20 — VERIFY it's still in the current config."
- **what:** If page deploys use content_field (HTML only), component JS (/tools/assets/*.js) is silently dropped — news sections render empty and interactive tools ship as shells. The files_field form carries the full file set. Related evidence: tool generation works but deploy is where tools stalled (gas-unit-converter built with real JS but stuck build_status='pending'); the working-tools acceptance is deployed page + committed JS + resolving links, never "component generated".
- **sources:** HANDOFF_robot_hands_rebuild.md#Tools/#News-pipeline, TODO_imagery_followups.md#17
- **relations:** robot-hands rebuild hard requirement; render_css_from_spec fallback gap (page-level CSS silently not shipped) noted alongside.
- **verify-later:** page-rerender deploy_page config; tool page build_status across sites.

<!-- SOURCE: U14_docs019_runbooks.md -->
### Commented-out tool route and the planned-tool-page seam
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B3 "COMMENTED-OUT FUTURE ROUTES present: entity-directory, entity-page, and 'tool' → tool-build-handler (needs_tool_page)"; §B5 "ON HOLD: coordination with the parallel tools chat … The §B5 interface — how a PLANNED tool page reaches the pipeline — is a JOINT decision".
- **what:** The relay's reconcile routing table carries a commented "tool" → tool-build-handler route, so planned tool pages (e.g. dartsonline's headline tool-setup-builder differentiator) ship as prose via page-build-handler. Design fork recorded for the joint decision: (i) thin tool-build-handler driving generation into the synced page (page-creation conflict); (ii) tool-generator gains an existing-page mode; (iii) most reuse-shaped — no handler, a relay hop runs tool-suggester after site_plan and its pipeline owns page creation end-to-end. Accepted sequencing: ship prose first, upgrade later.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3 (C3); docs019/RUNBOOK_builder_route(21).md#B4 (§B5 candidate); docs019/RUNBOOK_builder_route(21).md#B5
- **relations:** work-item relay; tool pipeline (active suggester/generator/deployer); thread-boundary convention
- **verify-later:** load_work_item_actions.go commented routes; the joint-seam decision record

<!-- SOURCE: U15_docs019_running_notes.md -->
### Reuse-checking retrieval architecture
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Checking for reuse is a retrieval problem with a judgement tail, not a generation problem... A maintained capability catalog... turns the first reuse question into a lookup" (principles(59) §Reuse-checking).
- **what:** A framing (partially realised in the actual contextkit/code_symbols build) that reuse-checking should be almost entirely algorithmic: a maintained signature/type/call-graph index answers "have we solved this?" as a query rather than a whole-codebase read; exact-duplicate detection is algorithmic/high-precision (fingerprinting), "similar" detection is semantic/fuzzy (embeddings); a cheap model should narrow candidates for recall, never decide; and any reuse index rots like any derived artifact, needing incremental refresh keyed to real ground-truth cases (past duplications caught in review), since the dangerous error (a missed match) leaves no trace.
- **sources:** NOTES_running_synthesis_principles(59) §Reuse-checking (finding code that already solves the problem).
- **relations:** B4a embedding-quality evaluation finding; code-context retrieval infrastructure.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Toolchain validator + repo read/search (net-new for code)
- **category:** tool-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) §4 "Validator changes kind: from contract checks … to a toolchain validator … the most important new piece"
- **what:** Low-regret net-new pieces for a self-coding pipeline: a toolchain validator giving ground-truth `go build/vet/test` + SQL dry-run pass/fail, a repo read/search capability (automating today's manual STEP ZERO), edits-against-existing-files rather than whole-file regeneration, and shared-repo serialization. The write→validate→regenerate loop, "broken output never overwrites," locks, HITL gating, and git→actions→backblaze deploy all transfer.
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#3, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#4, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2
- **relations:** verification harness; STEP ZERO; self-dev coordination positions
- **verify-later:** existing loop primitives; component_versions; needs_human_review

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool pipeline: tool-suggester → tool-generator/tool-deployer → cross-linking
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 062/062b/098 definitions and patches; 100 portfolio: "Six industry-specific tools deployed... Tool references automatically woven into 18 content pages. Full pipeline... runs autonomously."
- **what:** tool-suggester (evaluate_tools handler) uses LLM judgment over specs+pages to decide which interactive tools would genuinely help a site (not limited to library catalogue), creating add_tool items; tool-deployer forks a library tool to the site (component fork + tool page + page_component link, then normal render/deploy); tool-generator creates new tool HTML from brand context (and since 131 writes a travelling PLAN); 098 adds cross-linking — suggestions carry related_pages, and create_tool_cross_link_items generates content_rewrite items so page-build-handler weaves tool references into existing copy. missing_tools discovery check auto-seeds add_tool items.
- **sources:** 062_tool_suggester_and_improver.sql; 062b_tool_deployer_and_generator_agent.sql; 098_tool_suggester_cross_linking.sql; 061_tool_deployer_and_discovery_agent.sql
- **relations:** tool-library; tool acceptance tiers; travelling docs
- **verify-later:** deploy_tool_to_site action; create_tool_cross_link_items

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Tool pipeline end-to-end (suggest → route → generate/fork → cross-link → rewrite)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 005(1): "Pipeline Status: Fully Operational… all work without manual intervention" with per-site verified results
- **what:** check_missing_tools → tool-suggester (LLM judgement, 0-5 suggestions, library-vs-novel routing via check_is_library) → tool-deployer (fork) or tool-generator (novel) → create_cross_links (content_rewrite items per related page, item_key tool_crosslink:*, tool- pages filtered) → dispatch → page-build-handler threads rewrite_guidance (`input_data.spec.suggestion`) into the writer's nested loop prompt → rerender. The writer prompt lives deep in sub_workflow nesting — top-level jsonb_each misses it (072 trap).
- **sources:** 005_tool_pipeline(1).md full; 020 agents detail
- **relations:** de-tool hazard; fork-on-deploy; tool doc header
- **verify-later:** migrations 070–073 applied; cross-link items in prod

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Tools pipeline (suggest / deploy-fork / generate / improve / audit)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "Five agents, two discovery checks, full lifecycle documented … Tier 3 (headless browser visual testing) is planned, not built"; but Path D flags interactive behaviour "reportedly currently don't work" (2026-05-14)
- **what:** tool-suggester (LLM over spec aspects + library, 0-5 suggestions with library_source routing), tool-deployer (library fork with forked_from + tool page + companion guide), tool-generator (novel LLM tool, same wiring), tool-improver (issue-driven rewrite), tool_health Tier-1 structural check + tool-auditor Tier-2 LLM review with confidence-split routing. Missing vs the four-stage pattern: no parse stage (source tools not read), loose source-tool fidelity. Fork-retry idempotency fixed (P2: reuse orphaned forks; GetComponentByFunction excludes forks).
- **sources:** FOCUS_interactive_content_generation(4).md#Tools; HANDOFF-pipeline-triage-april-2026.md P1/P2
- **relations:** games gap (copies this shape); library model; quality model
- **verify-later:** actual tool interactivity failures (Path D); tool_health/tool-auditor definitions

<!-- SOURCE: U08_travelling_docs.md -->
### Tool creation never enqueues the final page deploy (planned-pages gap)
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Gap on record: tool creation ends at `complete` without enqueuing a page_rerender item — the pages deploy only when something else sweeps" (2026-07-10; pages hand-deployed by inserting page_rerender items).
- **what:** tool-generator creates component + page + nav but leaves the page `build_status='planned'`; nothing enqueues the render+deploy hop, so new tool pages 404 until an unrelated sweep. Recorded follow-up: a `create_rerender_item` tail on tool-generator. Interacts with acceptance timing (the reason post-creation acceptance hooks were rejected in favour of the deployed-pages-only sweep).
- **sources:** RUNNING_NOTES_travelling_docs(39).md#both-planned-tool-pages-deployed; HANDOFF_2026-07-10…md#§5.3
- **relations:** continuous sweep design; build/rerender pipeline.
- **verify-later:** whether tool-generator gained a rerender tail.

<!-- SOURCE: U08_travelling_docs.md -->
### Inline-JS extraction ("Path 1" /tools/assets/<fn>.js) — designed, partly real, not on the deploy path
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** 016b v3 entry (separateInlineJS/collectJSAssets exist and are correct for the store path); Tier-2 first sweep proved the deploy path ships JS INLINE and never references the asset (js-not-extracted, Option B superseded the criteria).
- **what:** The store path's `separateInlineJS` extracts a bare inline `<script>` into `js_content`, replaced by a `<script src="/tools/assets/{function}.js">` reference deployed by `collectJSAssets` — but only for attribute-less tags, and legacy/seeded rows predate it (empty shells with raw inline scripts; provocation-card additionally truncated mid-script — store validation checks unclosed `<style>` but not `<script>`). Meanwhile the generator/deploy route for new tools ships everything inline, so "Path 1 extraction" is delivered reality nowhere on that route. Hardening recorded: script-balance check at store time; regenerate broken shells through the current path.
- **sources:** 016b_debugging_guide_7_3_(7).md#js-not-extracted-entry; RUNBOOK_travelling_docs(38).md#stage-5 (pre-verification); PLAN_travelling_docs(6).md#rollout-outcomes
- **relations:** delivered-reality principle (Option B); empty-shell/mode-b categories; vonc case evidence.
- **verify-later:** store_generated_component_action.go separateInlineJS; whether extraction ever ships.

<!-- SOURCE: U10_imagery.md -->
### deploy_page files_field contract (co-located JS must ship)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "CRITICAL deploy dependency: page-rerender deploy_page step MUST use files_field:'rendered_page.files' (NOT content_field)… fix was applied during the gas rollout 2026-05-19/20 — VERIFY it's still in the current config."
- **what:** If page deploys use content_field (HTML only), component JS (/tools/assets/*.js) is silently dropped — news sections render empty and interactive tools ship as shells. The files_field form carries the full file set. Related evidence: tool generation works but deploy is where tools stalled (gas-unit-converter built with real JS but stuck build_status='pending'); the working-tools acceptance is deployed page + committed JS + resolving links, never "component generated".
- **sources:** HANDOFF_robot_hands_rebuild.md#Tools/#News-pipeline, TODO_imagery_followups.md#17
- **relations:** robot-hands rebuild hard requirement; render_css_from_spec fallback gap (page-level CSS silently not shipped) noted alongside.
- **verify-later:** page-rerender deploy_page config; tool page build_status across sites.

<!-- SOURCE: U14_docs019_runbooks.md -->
### Commented-out tool route and the planned-tool-page seam
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) §B3 "COMMENTED-OUT FUTURE ROUTES present: entity-directory, entity-page, and 'tool' → tool-build-handler (needs_tool_page)"; §B5 "ON HOLD: coordination with the parallel tools chat … The §B5 interface — how a PLANNED tool page reaches the pipeline — is a JOINT decision".
- **what:** The relay's reconcile routing table carries a commented "tool" → tool-build-handler route, so planned tool pages (e.g. dartsonline's headline tool-setup-builder differentiator) ship as prose via page-build-handler. Design fork recorded for the joint decision: (i) thin tool-build-handler driving generation into the synced page (page-creation conflict); (ii) tool-generator gains an existing-page mode; (iii) most reuse-shaped — no handler, a relay hop runs tool-suggester after site_plan and its pipeline owns page creation end-to-end. Accepted sequencing: ship prose first, upgrade later.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3 (C3); docs019/RUNBOOK_builder_route(21).md#B4 (§B5 candidate); docs019/RUNBOOK_builder_route(21).md#B5
- **relations:** work-item relay; tool pipeline (active suggester/generator/deployer); thread-boundary convention
- **verify-later:** load_work_item_actions.go commented routes; the joint-seam decision record

<!-- SOURCE: U15_docs019_running_notes.md -->
### Reuse-checking retrieval architecture
- **category:** tool-pipeline
- **status-signal:** partial
- **status-evidence:** "Checking for reuse is a retrieval problem with a judgement tail, not a generation problem... A maintained capability catalog... turns the first reuse question into a lookup" (principles(59) §Reuse-checking).
- **what:** A framing (partially realised in the actual contextkit/code_symbols build) that reuse-checking should be almost entirely algorithmic: a maintained signature/type/call-graph index answers "have we solved this?" as a query rather than a whole-codebase read; exact-duplicate detection is algorithmic/high-precision (fingerprinting), "similar" detection is semantic/fuzzy (embeddings); a cheap model should narrow candidates for recall, never decide; and any reuse index rots like any derived artifact, needing incremental refresh keyed to real ground-truth cases (past duplications caught in review), since the dangerous error (a missed match) leaves no trace.
- **sources:** NOTES_running_synthesis_principles(59) §Reuse-checking (finding code that already solves the problem).
- **relations:** B4a embedding-quality evaluation finding; code-context retrieval infrastructure.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Toolchain validator + repo read/search (net-new for code)
- **category:** tool-pipeline
- **status-signal:** aspirational
- **status-evidence:** FOCUS_self_development(1) §4 "Validator changes kind: from contract checks … to a toolchain validator … the most important new piece"
- **what:** Low-regret net-new pieces for a self-coding pipeline: a toolchain validator giving ground-truth `go build/vet/test` + SQL dry-run pass/fail, a repo read/search capability (automating today's manual STEP ZERO), edits-against-existing-files rather than whole-file regeneration, and shared-repo serialization. The write→validate→regenerate loop, "broken output never overwrites," locks, HITL gating, and git→actions→backblaze deploy all transfer.
- **sources:** ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#3, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#4, ED/FOCUS_self_development_coding_pipeline_reasoning(1).md#2
- **relations:** verification harness; STEP ZERO; self-dev coordination positions
- **verify-later:** existing loop primitives; component_versions; needs_human_review

<!-- SOURCE: U18_sql_for_agents.md -->
### Tool pipeline: tool-suggester → tool-generator/tool-deployer → cross-linking
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** 062/062b/098 definitions and patches; 100 portfolio: "Six industry-specific tools deployed... Tool references automatically woven into 18 content pages. Full pipeline... runs autonomously."
- **what:** tool-suggester (evaluate_tools handler) uses LLM judgment over specs+pages to decide which interactive tools would genuinely help a site (not limited to library catalogue), creating add_tool items; tool-deployer forks a library tool to the site (component fork + tool page + page_component link, then normal render/deploy); tool-generator creates new tool HTML from brand context (and since 131 writes a travelling PLAN); 098 adds cross-linking — suggestions carry related_pages, and create_tool_cross_link_items generates content_rewrite items so page-build-handler weaves tool references into existing copy. missing_tools discovery check auto-seeds add_tool items.
- **sources:** 062_tool_suggester_and_improver.sql; 062b_tool_deployer_and_generator_agent.sql; 098_tool_suggester_cross_linking.sql; 061_tool_deployer_and_discovery_agent.sql
- **relations:** tool-library; tool acceptance tiers; travelling docs
- **verify-later:** deploy_tool_to_site action; create_tool_cross_link_items

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared content-component reuse model (one content_components row, N page_components instances)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** Stated as settled platform fact throughout ("one `content_components` row, N `page_components` instances", BUNDLE(3) §1); vonc.com/index arrived mid-recovery as "healthy sixth dependent" proving live reuse.
- **what:** A section component is a single shared `content_components` row (keyed by `function`, with `section_type`, `input_schema.fields`, `html_template`, `is_active`, `forked_from`) reused across pages and sites; each page stores its own `content_data` in `page_components`. Any change to the shared row therefore has a cross-site blast radius — the structural precondition for both incidents in this thread.
- **sources:** BUNDLE(3).md §1; NOTES(43).md §1, §9z (vonc sixth dependent); HANDOFF(7).md §Platform operating model
- **relations:** clobber failure mode; F4 fork-vs-match; F8 contamination; optimistic-lock co-management.
- **verify-later:** content_components + page_components schemas; idx_cc_selector (section_type, component_level) partial index.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### StoreGeneratedComponentAction regeneration branch
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Correction 6 — StoreGeneratedComponentAction is the rename writer" (NOTES §2); insert/dedup shape read from the live file (NOTES §9q).
- **what:** On storing a generated component whose `function` matches an existing row (`WHERE function=$1 AND forked_from IS NULL`, deliberately is_active-agnostic since 2026-05-06), the action snapshots the old schema/template to `component_versions`, UPDATEs the shared row in place (same component_id, so dependents follow), marks dependents pending (`markPagesPendingRebuild` — build_status only, no rendered_html), and raises one deduped `needs_rerender` work item per affected site via `createRerenderWorkItem`. Pre-fix, that item carried no `reason`, making the triggered re-render assemble-only and unable to repair anything.
- **sources:** NOTES(43).md §2 Correction 6, §9h, §9q; BUNDLE(3).md §3
- **relations:** F1 guard lives in its validation block; F3b re-added the reason to its spec; F4 (its function-keyed lookup is the fork vector).
- **verify-later:** store_generated_component_action.go (existence check L198–207; regen branch; markPagesPendingRebuild; createRerenderWorkItem NOT EXISTS dedup).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1 field-contract guard (reject regens that rename/drop retained fields)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE… 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o, 2026-07-02 16:39); RUNBOOK(49) Part A "Fixes deployed + proven".
- **what:** In StoreGeneratedComponentAction's Layer-1 validation, on `isRegeneration` the guard diffs old vs new `input_schema.fields` names (`schemaFieldSet` helper); any retained field that disappears becomes a blockingIssue routed through `recordValidationRejection` into `agent_error_log` — additions allowed, renames/drops rejected before any snapshot/UPDATE. Converts silent stranding into a loud, queryable rejection naming the stranded fields. Design choice: preserve-the-contract strict-reject backstop, not a per-dependent migration.
- **sources:** F1_store_generated_component_action.patch; NOTES(43).md §9, §9a, §9o; RUNBOOK(49).md Part A
- **relations:** F1-prompt (generation-time complement so name-preserving regens succeed); F5 (proposed extension); F8 (guard checks names only — its blind spot).
- **verify-later:** store_generated_component_action.go guard block + schemaFieldSet; agent_error_log rows error_code component_validation_rejected; store_generated_component_guard_test.go.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1-prompt generation-time field-name preservation (loader + dormant rule + function pin)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F1-prompt 1a/1b+2 done (validator passes; loader executes; verify row good)" (NOTES §9k); Tier 3a two regens preserved names with md5-verified template change (§9l).
- **what:** Three coupled pieces so regens preserve names instead of being rejected: (1) `load_existing_component` Go action — looks up the canonical component by `section_type` (is_active, forked_from IS NULL, component_level='section'), outputs `existing_component.field_names` + `function`; advisory, never errors (no match → empty map → rule dormant). (2) A snapshot-first, anchored, idempotent, drift-checked SQL migration wiring the step before generate_template and inserting a dormant `{{if .existing_component.field_names}}` prompt rule: reuse these exact names, MAY add, MUST NOT rename/drop. (3) `F1prompt2_pin_function.sql` pins `{{.existing_component.function}}` so the store matches the same row (the F4 mitigation). Option A (pre-generation lookup by section_type) chosen after `\d content_components` confirmed section_type is queryable.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c–§9f, §9m; RUNBOOK(49).md Part A
- **relations:** F1 guard (backstop); F4 (pin is its mitigation); prompt-migration convention; deploy-ordering gate (its 9i failure).
- **verify-later:** load_existing_component_action.go + registry.go (IsLocal:true); component-creator default_config (top-level prompt_template; load_existing_component step; input_fields incl. existing_component).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Store-driven retry on field-drift rejection (Option B)
- **category:** NEW:component-lifecycle
- **status-signal:** abandoned
- **status-evidence:** "Two options… user chose A" (NOTES §9d–§9e); Option B never mentioned again after 9e.
- **what:** The rejected alternative to Option A: on a field-drift rejection the guard would return the existing field names, and a store_component error edge would loop back to generate_template with the names injected, retrying once. Judged heavier (reject-with-retry-data + loop-guarded workflow edge) but authoritative (no key guessing); dropped when section_type proved a stable pre-generation lookup key.
- **sources:** NOTES(43).md §9d, §9e
- **relations:** F1-prompt Option A (superseding choice).
- **verify-later:** absence: component-creator workflow should have no store_component→generate_template error edge.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F4 — regen-vs-create keyed on the LLM-chosen function (silent fork)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "F4 (structural finding): regen-vs-create is keyed on the LLM-chosen function → nondeterministic. Miss case = silent FORK" (NOTES §9m); pin migration applied §9n; "store-side advisory FLAGGED as follow-up, not built".
- **what:** Whether a store is a regeneration or a creation depends on whether the LLM happened to choose the existing `function` name — a miss silently creates a parallel active duplicate for a section_type (library fragmentation; guard bypassed by fork; selector nondeterminism). Observed live in F2 testing (fork 80222fc1). Mitigation shipped: prompt pin of `{{.existing_component.function}}`; store-side advisory (warn when function misses but an active same-section_type row exists) deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate. A suspected live-fork case (duplicate hero rows) was later softened to old manual seeding.
- **sources:** NOTES(43).md §9m, §9n, §9al, §9am; RUNBOOK(49).md Part E
- **relations:** F1-prompt function pin; StoreGeneratedComponentAction lookup; F2 methodology (exposed it).
- **verify-later:** duplicate non-forked function rows in content_components; whether any store-side advisory exists.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F5 — regen-added required fallback-less fields strand renderability
- **category:** NEW:component-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Flagged F5: extend the F1 guard… Not built now" (NOTES §9v); still an open Part E flag in RUNBOOK(49).
- **what:** The incident's second facet: the 15:06 regen also ADDED `cta_url` (required, Tier-C source, no fallback) that no affected site's specs could satisfy — renames strand stored content, required additions strand renderability (sections permanently not-ready → carried forever). Proposed guard extension: reject, or force optional/fallbacked, any added required field on a regeneration.
- **sources:** NOTES(43).md §9v; RUNBOOK(49).md Part E F5; HANDOFF(7).md §Flags
- **relations:** F1 guard; section readiness model; carry-forward path.
- **verify-later:** store_generated_component_action.go — absence of an added-required-field check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F7 — unguarded template swap in update_component_html (residual)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "the snapshot INSERT is ALREADY FIXED in current code… Residual: no placeholder⇄schema sync validation on template swaps" (NOTES §9ak); Part E flag open.
- **what:** `update_component_html` swaps a shared component's template (snapshotting versions — its old silent snapshot failure on the removed `version_note` column is fixed) but performs no placeholder⇄schema agreement validation and no field-contract guard, leaving a second, unguarded write path to shared components. The original F7 framing (an unversioned live swap on hero) was investigated and softened; the residual is the missing validation.
- **sources:** NOTES(43).md §9aj, §9ak, §9am; RUNBOOK(49).md Part E F7; HANDOFF(7).md §Flags
- **relations:** F1 guard (candidate extension target); component versioning.
- **verify-later:** update_component_html_action.go — snapshot INSERT columns; absence of schema-sync validation.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Component versioning via component_versions (and unversioned-write provenance)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "Component updated AGAIN 2026-07-03 13:22:44 with NO v2 row… unversioned write path provenance OPEN" (NOTES §9an); manual snapshots v2/v3 taken to backfill.
- **what:** `component_versions` snapshots (component_id, version_number, schema, template, change_description, changed_by, change_source) are the change history for shared components; `change_source` records the triggering work item's source (useful provenance). Coverage is incomplete: some write paths historically failed silently or bypass versioning entirely, so manual mirror-the-working-insert snapshots are the established compensation before risky writes, and zero-version updates are treated as an investigation smell.
- **sources:** NOTES(43).md §9k, §9an, §9ao, §9bd; RUNBOOK(49).md Part C Step 6
- **relations:** F7; snapshot-before-change conventions; F8 remediation (v2/v3 snapshots).
- **verify-later:** component_versions rows for fdd92ad4 and brief-explanation; snapshotComponentVersion call sites.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### llm_guidance as a per-field generation-steering surface
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Writer prompt renders per-field guidance ⇒ every writer pass on any site was instructed to write vonc's product" (NOTES §9ba); writer config read §9az.
- **what:** Each `input_schema.fields` entry may carry `llm_guidance`, which page-content-writer renders into its generate_content prompt as the field's instruction (alongside name/type/required/description; fallback values notably never enter the prompt). On a shared component this is the highest-leverage contamination/steering surface — it shapes all future content on every consuming site — and therefore must be site-neutral while preserving structural guidance (word counts, `<em>` accent rule).
- **sources:** NOTES(43).md §9az–§9bb; RUNBOOK(49).md Part C Step 7 (the 11 neutral strings)
- **relations:** F8 carrier 3; page-content-writer prompt assembly; F8 lint scope.
- **verify-later:** page-content-writer default_config generate_content prompt (llm_field_specs block); brief-explanation field attrs.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Superseded hypothesis: update_component_html re-renders dependents inline
- **category:** NEW:component-lifecycle
- **status-signal:** superseded
- **status-evidence:** Original BUNDLE base: "the regeneration then re-renders every dependent page… rewritten together at 15:06:12.956, roughly 16ms after the component's own update"; Correction 1: "update_component_html is clean… the synchronized timestamp is just the pending-flag, not a render".
- **what:** The initial working theory held that update_component_html performed an inline dependent re-render (inferred from the ~16ms synchronized timestamps) and was the clobber writer. Disproved by reading the action: it only snapshots, swaps, and flags pending; the blame moved through RenderComponentAction and component-creator's workflow (both cleared) to StoreGeneratedComponentAction. Worth keeping as the exemplar of the thread's core epistemics: seven early inferences were each corrected against code/data before any fix shipped ("distrust each early inference until verified").
- **sources:** BUNDLE_component_regen_clobber.md §1 (base version); NOTES(43).md §2 Corrections 1–7, §3
- **relations:** clobber failure mode; StoreGeneratedComponentAction (actual writer).
- **verify-later:** n/a (historical).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Shared content-component reuse model (one content_components row, N page_components instances)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** Stated as settled platform fact throughout ("one `content_components` row, N `page_components` instances", BUNDLE(3) §1); vonc.com/index arrived mid-recovery as "healthy sixth dependent" proving live reuse.
- **what:** A section component is a single shared `content_components` row (keyed by `function`, with `section_type`, `input_schema.fields`, `html_template`, `is_active`, `forked_from`) reused across pages and sites; each page stores its own `content_data` in `page_components`. Any change to the shared row therefore has a cross-site blast radius — the structural precondition for both incidents in this thread.
- **sources:** BUNDLE(3).md §1; NOTES(43).md §1, §9z (vonc sixth dependent); HANDOFF(7).md §Platform operating model
- **relations:** clobber failure mode; F4 fork-vs-match; F8 contamination; optimistic-lock co-management.
- **verify-later:** content_components + page_components schemas; idx_cc_selector (section_type, component_level) partial index.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### StoreGeneratedComponentAction regeneration branch
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Correction 6 — StoreGeneratedComponentAction is the rename writer" (NOTES §2); insert/dedup shape read from the live file (NOTES §9q).
- **what:** On storing a generated component whose `function` matches an existing row (`WHERE function=$1 AND forked_from IS NULL`, deliberately is_active-agnostic since 2026-05-06), the action snapshots the old schema/template to `component_versions`, UPDATEs the shared row in place (same component_id, so dependents follow), marks dependents pending (`markPagesPendingRebuild` — build_status only, no rendered_html), and raises one deduped `needs_rerender` work item per affected site via `createRerenderWorkItem`. Pre-fix, that item carried no `reason`, making the triggered re-render assemble-only and unable to repair anything.
- **sources:** NOTES(43).md §2 Correction 6, §9h, §9q; BUNDLE(3).md §3
- **relations:** F1 guard lives in its validation block; F3b re-added the reason to its spec; F4 (its function-keyed lookup is the fork vector).
- **verify-later:** store_generated_component_action.go (existence check L198–207; regen branch; markPagesPendingRebuild; createRerenderWorkItem NOT EXISTS dedup).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1 field-contract guard (reject regens that rename/drop retained fields)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F2 COMPLETE… 3b reject ✔ (live firing, three-level visibility, zero mutation)" (NOTES §9o, 2026-07-02 16:39); RUNBOOK(49) Part A "Fixes deployed + proven".
- **what:** In StoreGeneratedComponentAction's Layer-1 validation, on `isRegeneration` the guard diffs old vs new `input_schema.fields` names (`schemaFieldSet` helper); any retained field that disappears becomes a blockingIssue routed through `recordValidationRejection` into `agent_error_log` — additions allowed, renames/drops rejected before any snapshot/UPDATE. Converts silent stranding into a loud, queryable rejection naming the stranded fields. Design choice: preserve-the-contract strict-reject backstop, not a per-dependent migration.
- **sources:** F1_store_generated_component_action.patch; NOTES(43).md §9, §9a, §9o; RUNBOOK(49).md Part A
- **relations:** F1-prompt (generation-time complement so name-preserving regens succeed); F5 (proposed extension); F8 (guard checks names only — its blind spot).
- **verify-later:** store_generated_component_action.go guard block + schemaFieldSet; agent_error_log rows error_code component_validation_rejected; store_generated_component_guard_test.go.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F1-prompt generation-time field-name preservation (loader + dormant rule + function pin)
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "F1-prompt 1a/1b+2 done (validator passes; loader executes; verify row good)" (NOTES §9k); Tier 3a two regens preserved names with md5-verified template change (§9l).
- **what:** Three coupled pieces so regens preserve names instead of being rejected: (1) `load_existing_component` Go action — looks up the canonical component by `section_type` (is_active, forked_from IS NULL, component_level='section'), outputs `existing_component.field_names` + `function`; advisory, never errors (no match → empty map → rule dormant). (2) A snapshot-first, anchored, idempotent, drift-checked SQL migration wiring the step before generate_template and inserting a dormant `{{if .existing_component.field_names}}` prompt rule: reuse these exact names, MAY add, MUST NOT rename/drop. (3) `F1prompt2_pin_function.sql` pins `{{.existing_component.function}}` so the store matches the same row (the F4 mitigation). Option A (pre-generation lookup by section_type) chosen after `\d content_components` confirmed section_type is queryable.
- **sources:** F1prompt_component_creator_preserve_field_names(1).sql; NOTES(43).md §9c–§9f, §9m; RUNBOOK(49).md Part A
- **relations:** F1 guard (backstop); F4 (pin is its mitigation); prompt-migration convention; deploy-ordering gate (its 9i failure).
- **verify-later:** load_existing_component_action.go + registry.go (IsLocal:true); component-creator default_config (top-level prompt_template; load_existing_component step; input_fields incl. existing_component).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Store-driven retry on field-drift rejection (Option B)
- **category:** NEW:component-lifecycle
- **status-signal:** abandoned
- **status-evidence:** "Two options… user chose A" (NOTES §9d–§9e); Option B never mentioned again after 9e.
- **what:** The rejected alternative to Option A: on a field-drift rejection the guard would return the existing field names, and a store_component error edge would loop back to generate_template with the names injected, retrying once. Judged heavier (reject-with-retry-data + loop-guarded workflow edge) but authoritative (no key guessing); dropped when section_type proved a stable pre-generation lookup key.
- **sources:** NOTES(43).md §9d, §9e
- **relations:** F1-prompt Option A (superseding choice).
- **verify-later:** absence: component-creator workflow should have no store_component→generate_template error edge.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F4 — regen-vs-create keyed on the LLM-chosen function (silent fork)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "F4 (structural finding): regen-vs-create is keyed on the LLM-chosen function → nondeterministic. Miss case = silent FORK" (NOTES §9m); pin migration applied §9n; "store-side advisory FLAGGED as follow-up, not built".
- **what:** Whether a store is a regeneration or a creation depends on whether the LLM happened to choose the existing `function` name — a miss silently creates a parallel active duplicate for a section_type (library fragmentation; guard bypassed by fork; selector nondeterminism). Observed live in F2 testing (fork 80222fc1). Mitigation shipped: prompt pin of `{{.existing_component.function}}`; store-side advisory (warn when function misses but an active same-section_type row exists) deliberately advisory-only and unbuilt, since multiple components per section_type can be legitimate. A suspected live-fork case (duplicate hero rows) was later softened to old manual seeding.
- **sources:** NOTES(43).md §9m, §9n, §9al, §9am; RUNBOOK(49).md Part E
- **relations:** F1-prompt function pin; StoreGeneratedComponentAction lookup; F2 methodology (exposed it).
- **verify-later:** duplicate non-forked function rows in content_components; whether any store-side advisory exists.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F5 — regen-added required fallback-less fields strand renderability
- **category:** NEW:component-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Flagged F5: extend the F1 guard… Not built now" (NOTES §9v); still an open Part E flag in RUNBOOK(49).
- **what:** The incident's second facet: the 15:06 regen also ADDED `cta_url` (required, Tier-C source, no fallback) that no affected site's specs could satisfy — renames strand stored content, required additions strand renderability (sections permanently not-ready → carried forever). Proposed guard extension: reject, or force optional/fallbacked, any added required field on a regeneration.
- **sources:** NOTES(43).md §9v; RUNBOOK(49).md Part E F5; HANDOFF(7).md §Flags
- **relations:** F1 guard; section readiness model; carry-forward path.
- **verify-later:** store_generated_component_action.go — absence of an added-required-field check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F7 — unguarded template swap in update_component_html (residual)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "the snapshot INSERT is ALREADY FIXED in current code… Residual: no placeholder⇄schema sync validation on template swaps" (NOTES §9ak); Part E flag open.
- **what:** `update_component_html` swaps a shared component's template (snapshotting versions — its old silent snapshot failure on the removed `version_note` column is fixed) but performs no placeholder⇄schema agreement validation and no field-contract guard, leaving a second, unguarded write path to shared components. The original F7 framing (an unversioned live swap on hero) was investigated and softened; the residual is the missing validation.
- **sources:** NOTES(43).md §9aj, §9ak, §9am; RUNBOOK(49).md Part E F7; HANDOFF(7).md §Flags
- **relations:** F1 guard (candidate extension target); component versioning.
- **verify-later:** update_component_html_action.go — snapshot INSERT columns; absence of schema-sync validation.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Component versioning via component_versions (and unversioned-write provenance)
- **category:** NEW:component-lifecycle
- **status-signal:** partial
- **status-evidence:** "Component updated AGAIN 2026-07-03 13:22:44 with NO v2 row… unversioned write path provenance OPEN" (NOTES §9an); manual snapshots v2/v3 taken to backfill.
- **what:** `component_versions` snapshots (component_id, version_number, schema, template, change_description, changed_by, change_source) are the change history for shared components; `change_source` records the triggering work item's source (useful provenance). Coverage is incomplete: some write paths historically failed silently or bypass versioning entirely, so manual mirror-the-working-insert snapshots are the established compensation before risky writes, and zero-version updates are treated as an investigation smell.
- **sources:** NOTES(43).md §9k, §9an, §9ao, §9bd; RUNBOOK(49).md Part C Step 6
- **relations:** F7; snapshot-before-change conventions; F8 remediation (v2/v3 snapshots).
- **verify-later:** component_versions rows for fdd92ad4 and brief-explanation; snapshotComponentVersion call sites.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### llm_guidance as a per-field generation-steering surface
- **category:** NEW:component-lifecycle
- **status-signal:** deployed
- **status-evidence:** "Writer prompt renders per-field guidance ⇒ every writer pass on any site was instructed to write vonc's product" (NOTES §9ba); writer config read §9az.
- **what:** Each `input_schema.fields` entry may carry `llm_guidance`, which page-content-writer renders into its generate_content prompt as the field's instruction (alongside name/type/required/description; fallback values notably never enter the prompt). On a shared component this is the highest-leverage contamination/steering surface — it shapes all future content on every consuming site — and therefore must be site-neutral while preserving structural guidance (word counts, `<em>` accent rule).
- **sources:** NOTES(43).md §9az–§9bb; RUNBOOK(49).md Part C Step 7 (the 11 neutral strings)
- **relations:** F8 carrier 3; page-content-writer prompt assembly; F8 lint scope.
- **verify-later:** page-content-writer default_config generate_content prompt (llm_field_specs block); brief-explanation field attrs.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Superseded hypothesis: update_component_html re-renders dependents inline
- **category:** NEW:component-lifecycle
- **status-signal:** superseded
- **status-evidence:** Original BUNDLE base: "the regeneration then re-renders every dependent page… rewritten together at 15:06:12.956, roughly 16ms after the component's own update"; Correction 1: "update_component_html is clean… the synchronized timestamp is just the pending-flag, not a render".
- **what:** The initial working theory held that update_component_html performed an inline dependent re-render (inferred from the ~16ms synchronized timestamps) and was the clobber writer. Disproved by reading the action: it only snapshots, swaps, and flags pending; the blame moved through RenderComponentAction and component-creator's workflow (both cleared) to StoreGeneratedComponentAction. Worth keeping as the exemplar of the thread's core epistemics: seven early inferences were each corrected against code/data before any fix shipped ("distrust each early inference until verified").
- **sources:** BUNDLE_component_regen_clobber.md §1 (base version); NOTES(43).md §2 Corrections 1–7, §3
- **relations:** clobber failure mode; StoreGeneratedComponentAction (actual writer).
- **verify-later:** n/a (historical).

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Games quality lifecycle parity (new game_health / game-auditor / game-behavioral-tester / game-improver)
- **category:** NEW:games-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Games currently have no quality lifecycle. Add the analogues, reusing tool shapes wherever possible" (PLAN_tools_games_behavioral_qa_loop.md §7)
- **what:** Proposes mirroring the entire tool-lifecycle quality apparatus for games: `check_game_health.go` as the Tier-1 analogue of tool_health; `game-auditor` as the Tier-2 analogue of tool-auditor; `game-behavioral-tester` sharing the QA-loop harness with game-specific invariants; `game-improver` as the fix handler for `improve_game` items. Explicitly conditional on first confirming games are modelled compatibly (component_level/page_type, fork model) so the tool pipeline can be forked rather than rewritten.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md §7,§11
- **relations:** Behavioral QA loop for tools & games; tool-lifecycle (020)
- **verify-later:** whether `component_level='game'`/`page_type='game'` schema support already exists

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Games quality lifecycle parity (new game_health / game-auditor / game-behavioral-tester / game-improver)
- **category:** NEW:games-lifecycle
- **status-signal:** aspirational
- **status-evidence:** "Games currently have no quality lifecycle. Add the analogues, reusing tool shapes wherever possible" (PLAN_tools_games_behavioral_qa_loop.md §7)
- **what:** Proposes mirroring the entire tool-lifecycle quality apparatus for games: `check_game_health.go` as the Tier-1 analogue of tool_health; `game-auditor` as the Tier-2 analogue of tool-auditor; `game-behavioral-tester` sharing the QA-loop harness with game-specific invariants; `game-improver` as the fix handler for `improve_game` items. Explicitly conditional on first confirming games are modelled compatibly (component_level/page_type, fork model) so the tool pipeline can be forked rather than rewritten.
- **sources:** tools/tool_widget_clobber/PLAN_tools_games_behavioral_qa_loop.md §7,§11
- **relations:** Behavioral QA loop for tools & games; tool-lifecycle (020)
- **verify-later:** whether `component_level='game'`/`page_type='game'` schema support already exists
