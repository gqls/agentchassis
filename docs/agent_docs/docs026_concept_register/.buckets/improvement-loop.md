
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
