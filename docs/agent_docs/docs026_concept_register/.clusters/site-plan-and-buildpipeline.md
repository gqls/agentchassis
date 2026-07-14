# Cluster: site-plan-and-buildpipeline
Categories included: site-plan-and-reconciler, new:page-build-pipeline, new:build-pipeline, new:rebuild-cascade, new:work-dispatch, new:work-item-integrity, new:work-item-system, new:dispatch-pipeline, new:site-build-pipeline, new:site-build-orchestration-generations, new:action-build-pipeline


<!-- SOURCE: U01_docs024_numbered_core.md -->
### built_from_plan_version deploy-time stamp replaces the deployed→needs_rebuild flip (Option B)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 016 §9 dedicated entry (2026-05-28) "Fix shipped", completing the HANDOFF_2026-05-07 deferred design
- **what:** upsertPage's blunt flip (deployed→needs_rebuild on every sync) stood in for the unbuilt drift stamp; Option B stamps built_from_plan_version at the UpdatePageStatusAction deployed chokepoint and makes sync fill-if-null, retiring the flip so drift detection flows through the reconciler's decideEmit. Lesson (checklist 22): a "bug" may be a half-implemented design — complete it, don't patch around it.
- **sources:** 016 §9 flip entry; 029/030 design
- **relations:** reconciler; tool-page churn
- **verify-later:** any direct build_status='deployed' writes bypassing the action

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption slug-mangling: two canonicalisation surfaces must agree
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 016 §9 chain of entries (2026-05-19→26): cause pinned to WriteSitePlanAction ValidateRoles strip + SyncPagesToDBAction canonicalising raw page_plan WITHOUT ValidateRoles; fix "CHOSEN" (option 2) not confirmed shipped
- **what:** ValidateRoles strips tool-/guide-/game- prefixes and -index; CanonicalisePage re-adds them only for tool/game/guide roles, so wrong page_types (hubs typed content, guides typed blog-post) permanently flatten names/URLs. sync_pages_to_db reads raw page_plan (not site_plan_pages), skips ValidateRoles, and its ON CONFLICT overwrites correct adoption-time rows — one logical page, two writers, divergent results (incl. tool-game-* double prefixes). Fix: run the identical ValidateRoles+CanonicalisePage pipeline in sync (works for all five callers incl. plan-less pageflow-builder); root fix upstream is correct page_type at adoption; endgame is 029's deterministic slug preservation.
- **sources:** 016 §9 three linked entries; 030 phase-0 result
- **relations:** CanonicalisePage; page_type vocabulary
- **verify-later:** SyncPagesToDBAction ValidateRoles call present?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Plan as declarative artefact + reconciler (Kubernetes-style desired-vs-realised)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030: Phase 0 done (2026-05-04 re-adopt verified dedup); Phase 1 schema/decisions committed; 007 patch describes reconciler emitting needs_page as current behaviour
- **what:** The planner stops emitting work items; it writes desired state to plan-domain tables (site_plans one-current-per-site, site_plan_pages, site_plan_sections, site_plan_directives) and a deterministic Go reconciler diffs plan vs pages and emits idempotent needs_page items (with preference weights, cycle budget, dependency ordering). Fixes the two-writer duplicate-pages structural bug (adoption + planner not sharing identity space). Phase 2: discoverers/auditors read the plan for sharper fitness checks.
- **sources:** 029 full; 030 full; 007_adoption_pipeline_v4.patch
- **relations:** CanonicalisePage; built_from_plan_version; directives
- **verify-later:** site_plans tables live; reconciler action name

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Strategic vs plan-time guidance split (site_plan_directives)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030 Q1/Q2 decisions + 007 patch stating planner "no longer overwrites" adoption specs; lock-transfer designed
- **what:** site_specs.design_intent/content_direction stay strategic (classifier/adoption-owned); the planner's per-build guidance flattens into row-shaped site_plan_directives (scope site/page/section, category, subject, directive, source, Pattern-A locks) read by downstream agents via a brief renderer. One LLM call still produces structure+design+content together (coherence over three-call split); only the write targets change. HITL locks transfer across plan rebuilds by composite key inside write_site_plan.
- **sources:** 030#Q1/Q2, #Strategic vs plan-time naming; 031(3)#Lock transfer
- **relations:** B-029-4 design-intent clobber (motivating bug); lock transfer
- **verify-later:** site_plan_directives populated; brief renderer helper

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Per-page brief generation (lazy) and the no-empty-slots acceptance test
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** 029 B-029-2 promotes it to Phase-1 acceptance test; briefs "lazy" design section
- **what:** Component templates are named slots; a per-page brief enumerates slot content, generated lazily at build time. Without briefs, component-author defaults leak (empty img src, /services.html CTAs on sites without services). Acceptance: a Phase-1 build produces no empty slots and no leaked defaults — unbriefed slots either don't render or error before deploy.
- **sources:** 029#B-029-2, #Per-page brief generation
- **relations:** directives; B-029 bug list (dup nav items; theme vars never written)
- **verify-later:** brief generation exists?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Architectural Tension #1 — infer-and-repair vs deterministic structure derivation
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Status 2026-05-25: Tension #1 has a deployed partial fix (Part A — ValidateRoles -index rule), pending a clean production test"
- **what:** The pipeline takes structural decisions (page role/type/URL) from LLM free-text labels then repairs with starved, vertical-hardcoded heuristics, producing silent structural corruption (section hubs flattened to content). Resolution principle: derive structure deterministically from the LLM's reliable signal — naming (`<section>-index` marks a hub, vertical-agnostically); schema-constrain generation to kill form errors (necessary but not sufficient); make fallback heuristics fail loud, never default to content. Explicit recommendation AGAINST a free parent-pointer tree (worst LLM reliability tier); a leaf's section, if needed, is a constrained choice over the enumerated hub set.
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-1; HANDOFF_2026-05-26 (page_type re-type as an instance)
- **relations:** Tension #2; page_type vocabulary gap; LLM reliability strategy (same principle, component scale)
- **verify-later:** ValidateRoles -index rule and de-hardcoded nestedRoleFromURL in page_role_validator.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Architectural Tension #2 — page identity derived in multiple places that undo each other
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Tension #2's residual confirmed cosmetic (see HANDOFF_2026-05-25)" but flavour-collapse residual "evidence-gated, not yet a code change" (2026-05-25)
- **what:** Adoption, planner-write and convergence each re-derive canonical page name/role/URL with no single owner, so a later stage can undo an earlier correct result (convergence preserved games-index; WriteSitePlanAction flattened it one step later). Principle: one canonical owner; canonicalisation idempotent on already-canonical input; downstream reads identity read-only. Part A made section indexes round-trip cleanly; the remaining residual is flavour collapse (validator emits generic section-index, losing blog-index/entity-directory flavour) — decide from a deployed run whether the component resolver needs the flavour before writing preservation code. Withdrawn: merging the two role-normalisers (intentionally layered).
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-2; HANDOFF_2026-05-26 (write vs sync canonicaliser divergence)
- **relations:** Tension #1; kebab/snake; canonicaliser divergence
- **verify-later:** CanonicalisePage/normaliseRole/normalisePageType in datahelpers/page_canonical.go; component resolver's page_type dependence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site_specs vs site_plan two-layer architecture + aspect ownership contract
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "build-site-planner workflow writes both shapes during transition (old site_specs/site_plan aspect AND new plan tables)" (undated FOCUS, references docs 028-030)
- **what:** site_specs = strategic, brand-level, slow-changing, one owning agent per aspect (classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planner owns the four plan tables). site_plan tables = per-build, row-shaped, rebuilt per plan. Three ownership rules (don't read what you didn't spec; don't overwrite another's aspect; write outputs to the spec) with the classifier read-and-extend carve-out. Decision rules and anti-patterns for where new data lives (specs vs directives vs sibling structured tables).
- **sources:** FOCUS_site_spec_vs_site_plan.md (whole); ASSESSMENT_imagery_phase_0_1…md#What-Phase-1-changes
- **relations:** directive cascade; lock transfer; imagery placement
- **verify-later:** site_plans/site_plan_pages/site_plan_sections/site_plan_directives tables; legacy site_plan aspect readers (pageflow-builder)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site_plan_directives cascade + brief renderer
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Reconciler is documented in doc 030 but the chassis-side implementation has been landing in stages"; brief renderer named as `datahelpers/page_brief.go` "per the work order"
- **what:** Cross-cutting guidance rows located by (scope site/page/section, scope_ref, category, subject) with HITL lock columns. Consumers never read rows directly: a Go brief renderer cascades site → page → section and applies cardinality semantics (single-valued subjects override at narrower scope; multi-valued accumulate), emitting short LLM-ready briefs. The pattern imagery/text/design guidance should all follow.
- **sources:** FOCUS_site_spec_vs_site_plan.md#directives; ASSESSMENT_imagery_phase_0_1…md#Amendments
- **relations:** lock transfer; site_plan_imagery sibling-table pattern
- **verify-later:** datahelpers/page_brief.go existence and consumers

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### HITL lock transfer across plan rebuilds
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** described as run "inside write_site_plan" per doc 030; extended for imagery + lock_type/expiry per 2026-05 patches ("transferDirectiveLocks carries lock_type/expiry — written (patch doc)")
- **what:** On plan rebuild, locked directives from the previous current plan are matched to new rows by composite key (scope, scope_ref, category, subject, ordering); locked_at/locked_by and HITL-edited text copy over (HITL wins); unmatched locks drop with a log, previous plan kept as history. Any sibling table wanting HITL adopts the same shape.
- **sources:** FOCUS_site_spec_vs_site_plan.md#Lock-transfer; FOCUS_adoption_faithfulness_via_locks(2).md#dependency-chain
- **relations:** adoption-faithfulness timed locks; site_plan_imagery
- **verify-later:** transferDirectiveLocks in write_site_plan action code

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Section-data deferral + reconciler loop
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "reconcile_section_data_action.go — new, not yet wired to a host"; pages_under_section implemented (2026-06-02)
- **what:** query.*-sourced section fields unresolvable at plan time defer as needs_section_data; the queryresolve package (pages_where_type, now pages_under_section joining site_areas) resolves them; a lightweight reconciler (not an LLM agent — the once-planned directory-builder was never built) rescans open items whose missing fields are all query-sourced and emits needs_page re-renders (dedup key page_rerender:<page>), leaving human-data items (team, pricing) in HITL. plan_sections closes items on re-render. Host (loop check or post-build finalize) still to pick.
- **sources:** HANDOFF_2026-06-02…md#2; FOCUS_internal_linking.md#4; HANDOFF-pipeline-triage-april-2026.md P5
- **relations:** P5 plan-then-reconcile; list hubs; self-contained components heuristic gap
- **verify-later:** reconcile_section_data host + registry entry; queryresolve switch cases

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### page_type vocabulary gap forcing game→tool re-type (Gap B)
- **category:** site-plan-and-reconciler
- **status-signal:** unknown
- **status-evidence:** "root cause is confirmed from the planner's response_text … there is no `game` [in the Canonical Page Types list], so every adopted game is forced to `tool`" (2026-05-26); "OPEN structurally; may have been addressed by the other-chat fixes … Verify post-deploy"
- **what:** The plan_site prompt's closed page-type list lacks `game`; the LLM keeps names faithfully but re-types game pages as tool; canonicalisation's tool branch then renames, and a page_type change (not a name change) is what duplicates pages — 5 duplicate game-*/tool-game-* pairs on gamesdesign. Also exposed: WriteSitePlanAction and sync_pages_to_db canonicalise the same tool-typed page differently (tool-auto-battler vs tool-game-auto-battler) — code read required before fixing. Verification queries recorded (stem-grouped pages; response_text page_type; composition install).
- **sources:** HANDOFF_2026-05-26…md#diagnosis, #Where-to-resume
- **relations:** Tension #1/#2; games content type; adoption faithfulness locks
- **verify-later:** run the three handoff queries on a post-2026-05-26 adoption; page_canonical.go call sites

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Section data source triad and reconcile_section_data
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** HANDOFF (2026-06-19): "`reconcile_section_data` IS wired — registry.go line 914 … description 'Re-trigger pages whose deferred section data is now query-resolvable'" (correcting a stale note that it was not wired).
- **what:** A component's content comes from one of three sources, and fixes differ per case: (1) query-resolvable section data (the tools/guides-list kind — the reconciler's scope: `ReconcileSectionDataAction` re-triggers pages whose deferred data has become resolvable), (2) a human-entered spec field (e.g. pricing tier_1_* from `site_specs.pricing` — the reconciler correctly skips these), (3) page-content-writer prose (LLM-generated). The differentiators investigation established the triad as the diagnostic frame — and then found the actual fault was in none of the sources (a key-naming mismatch). Incidental same-thread finding: `write_site_spec` errors "missing required fields: [spec_data]" on persist_mission/roadmap — the action input is spec_data but the column is `data` (site_specs is aspect + data jsonb, UNIQUE(site_id,aspect) WHERE is_current).
- **sources:** HANDOFF_idea_uk_differentiators_section_data.md; bundle3; running_notes_scheme_to_components(55).md#Sa #Sh (corrected facts)
- **relations:** array item-fields contract (the real fault); plan_sections deferral.
- **verify-later:** reconcile_section_data_action.go scope logic; registry.go wiring; site_specs schema.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Planner re-plan union safety (normaliseRealisedToPlanPage)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Checkpoint (Un) 2026-07-07: "normaliseRealisedToPlanPage (v3_site_actions.go:4383) exists so a re-plan LOADS realised pages …, converts them to plan-page shape CARRYING their sections, and UNIONS with the LLM proposal — its own comment: without carrying sections the upsert would clobber built pages."
- **what:** Site composition is whole-plan and LLM-driven: build-site-planner (consuming needs_site_plan) supersedes the current site_plans row and rewrites site_plan_pages + site_plan_sections. Re-running it is safe by design because load_existing_pages surfaces realised pages and the normaliser unions them (with their sections) into the new plan — built pages keep their composition while catalogued-but-uncomposed pages get composed. This makes "emit needs_site_plan" the structural route for composing missing pages, versus hand-INSERTing plan rows (which drifts nav/plan/page consistency).
- **sources:** running_notes_scheme_to_components(55).md#Un; stepF_replan_read.sql
- **relations:** planned-but-uncomposed pages gap; work-item crafting conventions.
- **verify-later:** v3_site_actions.go normaliseRealisedToPlanPage; build-site-planner workflow steps.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Planned-but-uncomposed pages gap (catalogued, never composed)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** Checkpoint (Ul): "the three planned pages have NO site_plan_sections rows; their pages.sections = []. Catalogued, never composed"; (Un) ends with the replan-read staged — the emit had not run at the unit's last dated note (2026-07-07).
- **what:** A distinct failure shape: pages rows exist with page_type and nav intent set (news-index, guides-index, tool-audience-check on idea.uk), so navigation links to them and 404s, but they carry empty sections and no plan rows — the LLM plan behind the current site_plans row never included them. A W6-style needs_page emit would build an empty page; the correct route is two-phase: planner re-run composes them (union-safe), then needs_page builds and deploys. Also surfaced the distinction between query-backed index pages (news/guides may be fed by the blog-listing mechanism) and static pages, and reuse of the already-deployed audience-check tool component.
- **sources:** running_notes_scheme_to_components(55).md#Uk #Ul #Um #Un; RUNBOOK_scheme_to_components(50).md#PLANNED-PAGES; stepD_and_pages_reads.sql (block B/C)
- **relations:** planner re-plan union safety; navigation (nav 404s); rebuild vs rerender.
- **verify-later:** idea.uk pages rows for the three; site_plan_sections presence; whether the needs_site_plan emit ran.

<!-- SOURCE: U04_idea_uk.md -->
### Section-data reconciler and the human-sourced-field boundary
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "reconcile_section_data IS wired (registry.go L914… 're-trigger pages whose deferred section data is now query-resolvable')" — correcting an earlier stale "built but unwired" note (rr, 2026-06-19).
- **what:** Deferred section data (needs_section_data) is re-triggered when it becomes *query-resolvable*; the boundary concept: **human-sourced** spec fields (e.g. pricing tiers from site_specs.pricing) are not query-resolvable, so the reconciler can never fill them — either capture the data into specs (the £29 into pricing) or the section shouldn't be on the page. The unresolved-CTA gating (render no button when no eligible destination page exists) is the same honest-degradation family, tied to the thin 4-page plan having no hub pages.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + empty-index content gaps); idea.uk/README_001_todo_list.md
- **relations:** item_fields fix; site-plan thinness; content-governance (pricing spec).
- **verify-later:** reconcile_section_data_action.go host wiring; idea.uk pricing spec.

<!-- SOURCE: U05_content_quality_linking.md -->
### Section-index hub canonicalisation divergence + plan-version stamping
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 5: "Both the core fix … and the companion … are confirmed. Thread closed."; Part 9/10 A1 VERIFIED CLOSED.
- **what:** Two canonicalisation surfaces disagreed: WriteSitePlanAction ran ValidateRoles+CanonicalisePage (hubs → section-index, nested URLs) while SyncPagesToDBAction ran CanonicalisePage alone on the raw page_plan — flattening hubs on every sync. Fix (Option 2): sync runs the identical pipeline (Option 1 — read site_plan_pages — rejected because active callers have no plan at sync time). Companions: built_from_plan_version stamped at deploy time in UpdatePageStatusAction (completing the deferred doc-029 design), upsertPage COALESCE fill-if-null, and removal of the deployed→needs_rebuild flip (a pre-design stand-in that over-fired).
- **sources:** running_notes_14(26).md#part-1-3, #part-8; site_db_actions/upsertPage references throughout
- **relations:** reconciler drift detection; adoption faithfulness convergence; A1 tool deploy failure.
- **verify-later:** SyncPagesToDBAction ValidateRoles call; UpdatePageStatusAction stamp; reconciler decideEmit.

<!-- SOURCE: U05_content_quality_linking.md -->
### Adoption-faithfulness convergence + the []map type-assertion keystone bug
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14n: "CONVERGED … the convergence/duplicate-page root cause … is RESOLVED on a clean run" (2026-06-05 17:26).
- **what:** The reconcile-plan-with-realised subsystem (Pass A unions adoption-locked realised pages missing from the LLM plan; Pass C2 drops planned pages whose topic stem collides with an existing page) had NEVER functioned since deploy: ValidateSitePlanAction asserted existing_pages as []interface{} while QueryDatabaseAction returns []map[string]interface{} — the assertion always failed silently, so convergence no-op'd for every site (bare-sibling guide duplicates, guides absent from plans). Fix: type-switch both shapes + a count log so an empty set is never silent; plus normaliseRealisedToPlanPage carrying sections/meta/nav_order so the union can't clobber adopted pages to empty (the union-clobber that had emptied the source-populated hubs). Multiple interim framings (054 not applied; lock-window) were corrected en route — 053/054 were live; the killer was the type bug.
- **sources:** running_notes_14(26).md#part-14h-14n
- **relations:** locks (adoption_locked first-plan branch; 90-day replan window non-functional); planner sibling-invention; empty-hub union clobber.
- **verify-later:** ValidateSitePlanAction extraction switch; reconcilePlanWithRealised counters in planner logs.

<!-- SOURCE: U09_adoption.md -->
### First-plan branch: "no current plan + pages exist ⇒ adopted pages"
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "054 `load_existing_pages` — partially live. The query emits `adoption_locked` but only via the first-plan branch: CASE WHEN NOT EXISTS (current is_current plan for this site) THEN true" (2026-06-05 verified landed state).
- **what:** Deterministic detection of the faithful first pass: when `load_existing_pages` finds no current site_plan but pages exist, all existing pages are flagged `adoption_locked=true` (only ever true after adoption; from-scratch sites have no pages before the planner's own sync). Convergence keys off this flag; a re-adoption from a cleared DB (or retiring the current plan) makes any site a "first pass" deterministically.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#verified-landed-state, verify_readoption_fix.sql, running_notes_14(25)#part-14i
- **relations:** reconcilePlanWithRealised convergence; verify_readoption gate G1/G2 (retire current plan to force first pass)
- **verify-later:** live load_existing_pages SQL in build-site-planner def

<!-- SOURCE: U09_adoption.md -->
### Planner ignores adopted state (generic-skeleton overlay)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Diagnosed 2026-05-19 ("build-site-planner independently generates a 9-page generic site skeleton that ignores the adopted pages"); addressed by the convergence work verified 2026-06-05 plus the "Existing Pages — ALREADY BUILT, PRESERVE EXACTLY" prompt block (v1.0.1047). Residual: prompt alone did not stop differently-slugged siblings (bare `economy-basics` beside `guide-economy-basics`) — that took Pass C2.
- **what:** Two confirmed mechanisms: (1) the planner planned from identity/archetype without reading realised state, inventing parallel pages (renamed tool dups, `post` placeholder from a prompt example); (2) ValidateRoles couldn't converge a childless plan (section-index promotion needs a child declaring ParentSection). Root cause per doc 029: two surfaces (adoption, planner) both write pages and queue work without a shared identity space. Fix: planner reads realised state and converges; reconciler is the sole work-item producer ("can't produce duplicates by construction").
- **sources:** FOCUS_planner_ignores_adopted_state.md, running_notes_14(25)#part-14c–14e, migration_cleanup_bare_guide_duplicates.sql
- **relations:** doc 029/030 declarative plan + reconciler; reconcilePlanWithRealised; nav dedup guard B-029-1
- **verify-later:** `plan_site` prompt existing-pages block in live build-site-planner def; llm_call_log for planner runs

<!-- SOURCE: U09_adoption.md -->
### reconcilePlanWithRealised convergence (Pass A union, rename snap-back, Pass C/C2 dedup)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "VERIFIED RESOLVED on a clean run (2026-06-05 17:26Z, corr 6381cb13)… guide-economy-basics…guide-skinner-box all as role=guide (5), with ZERO bare siblings… Pass A unioned the adopted guides into the plan and Pass C2 dropped the bare-sibling duplicates, both firing for the first time."
- **what:** Deterministic Go convergence in `ValidateSitePlanAction`/`v3_site_actions.go`, gated on `adoption_locked` pages: unions LLM-omitted adopted pages into the plan (via `normaliseRealisedToPlanPage`), snaps back renames, dedups section-stem collisions (`sectionStemOf`) and item-topic siblings (`itemStemOf` strips tool-/guide-/game- prefixes mirroring CanonicalisePage — Pass C2), and truncates preserving locked pages. It does not special-case adoption in Go — it preserves whatever the query flags.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md, running_notes_14(25)#part-14l–14n
- **relations:** first-plan branch; type-assertion inertness bug (kept it dead until 06-05); union-clobber carry fix
- **verify-later:** `v3_site_actions.go` reconcilePlanWithRealised, itemStemOf; planner log lines "existing pages loaded for convergence", "reconciled with adoption-locked pages"

<!-- SOURCE: U09_adoption.md -->
### Union-clobber bug and the carry fix (sections/meta_description/nav_order)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "on the first pass, every adopted page the LLM omitted was unioned with empty values and the sync clobbered its real sections/meta_description/nav_order to empty… Fix (both must land together): (a) load_existing_pages SELECT adds the fields… (b) normaliseRealisedToPlanPage carries them" — verified on the 2026-06-05 clean run; "the empty hubs were the union clobber… NOT a planner gap."
- **what:** Pass A's union originally emitted `sections: []` because the 054 query didn't select the fields, and `upsertPage`'s `ON CONFLICT … sections = EXCLUDED.sections` overwrote the adopted page's real values — the difference between a faithful first pass and one that wipes adopted content the LLM didn't re-list. The carry fix also reframed the "empty hubs" defect: source hubs are populated (`guides-index → ["guide-list"]` etc.); no separate hub-convergence step is needed for adopted sites.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#union-clobber, running_notes_14(25)#part-14i–14j, migration mentioned: migration_load_existing_pages_carry_fields.sql
- **relations:** upsertPage ON CONFLICT semantics; empty-hub clarification; convergence
- **verify-later:** load_existing_pages SELECT column list; normaliseRealisedToPlanPage in v3_site_actions.go

<!-- SOURCE: U09_adoption.md -->
### Canonical page-shape vocabulary (CanonicalisePage + ValidateRoles)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Phase 0 "landed" (FOCUS_planner_ignores_adopted_state); Part A `-index` rule "written, unit-tested green, and deployed" and verified via the 2026-05-28 run (CATALOGUE §0 "hubs deployed as section-index at nested URLs").
- **what:** One canonical name/URL/page_type vocabulary for logical pages (index `/index.html`; `<slug>.html` content; `<section>-index` → `/<section>/index.html`; `tool-<slug>` → `/tools/<slug>/index.html`; guide role → `/guides/<slug>/index.html`), implemented in `datahelpers.ValidateRoles` + `CanonicalisePage` (page_canonical.go). Part A adds Rule 2: a name ending `-index` with a non-leaf role is promoted to `section-index` (with an `isLeafRole` guard), recovering the LLM's reliable signal (the name) when url/parent are omitted. Part B (de-hard-code the tools/guides/games vertical vocabulary in `nestedRoleFromURL`) remains unscoped. The two role-normalisers (`normaliseRole` routing-collapsed vs `normalisePageType` flavour-preserving) are intentionally layered — merging them was withdrawn as wrong.
- **sources:** HANDOFF_2026-05-25, FOCUS_chrome_templates_and_page_shape.md#fix-2, running_notes_14(25)#part-1–5
- **relations:** sync canonicalisation divergence; adoption URL computation (flat, pre-canonicaliser); guide page_type
- **verify-later:** `page_role_validator.go` (Rule 2 + isLeafRole), `page_canonical.go` guide case, `nestedRoleFromURL` hardcoded verticals

<!-- SOURCE: U09_adoption.md -->
### Two-canonicalisation-surfaces divergence: SyncPagesToDB lacked ValidateRoles
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "Both the core fix (sync no longer flattens hubs) and the companion (built_from_plan_version set…) are confirmed. Thread closed." (running_notes_14 Part 5, 2026-05-28).
- **what:** `WriteSitePlanAction` ran ValidateRoles+CanonicalisePage → correct plan; `SyncPagesToDBAction` ran CanonicalisePage only, on the raw `page_plan` from collected data — so a `games-index` typed `content` flattened to `/games-index.html` and the upsert overwrote the correct adoption row. Fix chosen: Option 2 — sync runs the identical ValidateRoles pipeline (Option 1, reading site_plan_pages, would break the plan-less callers pageflow-builder/multipage-website-builder/site-work-orchestrator). Exposed the deliberate guides de-prefix trade-off (plan de-prefixes `guide-rng-design`; sync now agrees — surfaced, not silent).
- **sources:** running_notes_14(25)#part-1–3, HANDOFF_2026-05-25#confirmed-root-cause
- **relations:** canonical vocabulary; built_from_plan_version companion; ARCHITECTURAL_TENSIONS #2 (identity derived in multiple places)
- **verify-later:** `site_db_actions.go` SyncPagesToDBAction normalisation loop

<!-- SOURCE: U09_adoption.md -->
### built_from_plan_version drift stamp + removal of the deployed→needs_rebuild flip
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "Option B shipped (two files, coupled)… confirmed in production" (running_notes_14 Part 8–10; CATALOGUE A1 fix list, 2026-06-03).
- **what:** The intended doc-029 design — stamp `pages.built_from_plan_version` at build time and detect staleness in the reconciler — had been deferred; a stand-in `deployed → needs_rebuild` flip in `upsertPage` over-fired on every sync and churned pre-plan tool deploys. Completion: `UpdatePageStatusAction` stamps the current plan id on deploy; `upsertPage` COALESCE fill-if-null (never overwrite a real build version) and the flip removed; drift flows through the reconciler's `decideEmit`. Principle recorded: before fixing a misbehaving mechanism, check for deferred design debt — complete it rather than patch around it.
- **sources:** running_notes_14(25)#part-8, CATALOGUE(9)#family-a, old2/HANDOFF_2026-05-07(1)#5
- **relations:** doc 029 drift detection; A1 tool/game deploy failure; reconciler stale-page churn
- **verify-later:** `v3_site_actions.go` UpdatePageStatusAction deployed branch; `site_db_actions.go` upsertPage CASE

<!-- SOURCE: U09_adoption.md -->
### Bare-sibling duplicate pages (planner re-invents adopted topics)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "DECISIVE (llm_call_log plan_site @ 20:25:22): the planner WAS given the adopted guides and emitted economy-basics anyway → PROMPT-RULE gap… FIX (recommended, structural, Go): deterministic guard… drop a planned page whose topic STEM collides" — shipped as Pass C2 and verified on the 2026-06-05 clean run; cleanup migration applied.
- **what:** The planner proposed bare `economy-basics` etc. beside adopted `guide-economy-basics` — a differently-slugged sibling the "preserve existing pages" prompt rule did not stop. Deterministic Go stem-dedup (Pass C2, reusing CanonicalisePage's prefix stripping) is the guarantee; a prompt stopgap was optional. The durable cleanup migration also removes the bare rows from the current plan (reconciler would re-create them otherwise) and terminalises their work items (site_work_items.page_id has no FK).
- **sources:** running_notes_14(25)#part-14c–14e, migration_cleanup_bare_guide_duplicates.sql, FOCUS_adoption_faithfulness_via_locks(5).md#item-topic-sibling-dedup
- **relations:** planner ignores adopted state; convergence Pass C2; LLM-rule vs deterministic-guard principle
- **verify-later:** itemStemOf/Pass C2 in v3_site_actions.go

<!-- SOURCE: U09_adoption.md -->
### Adoption calls the canonicaliser + reconciler orphan pruning
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** "Adoption today doesn't go through this. It computes its own URL based on page_type only… This needs an additional reconciler pass: pages… with NO entry in site_plan_pages… should be soft-deleted or marked for removal. The reconciler… doesn't prune orphans. That's a follow-up." (FOCUS_chrome_templates Fix 2).
- **what:** Adoption's local URL computation (flat `/games.html` etc.) diverges from the canonicaliser the planner uses, producing duplicate logical pages (`games` + `games-index`) that ON CONFLICT can't match. Proposed: apply_adoption_plan calls CanonicalisePage; reconciler gains an orphan-pruning pass (pages absent from the current plan get archived); one-off cleanup migration. Partially overtaken by the convergence work (which unions/dedups at plan time) and the analyze_site prompt fix, but orphan pruning remains unbuilt — orphaned bare pages persisted after Pass C2 dropped them from the plan and needed manual cleanup.
- **sources:** FOCUS_chrome_templates_and_page_shape.md#fix-2, running_notes_14(25)#part-14l follow-up
- **relations:** canonical vocabulary; bare-sibling cleanup migration (the manual stand-in); page-cleanup pass idea in 05-07 Phase-2 candidates
- **verify-later:** apply_adoption_plan URL computation today; any reconciler pruning logic

<!-- SOURCE: U09_adoption.md -->
### Deferred plumbing stubs: scheduled reconciler tick, domain-aware ensure_pages
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** "6. Scheduled reconciler tick — Not built. Reconciler currently fires only when called by the planner… 7. ensure_pages should be domain-aware — Currently hardcoded in workflow JSON… Stub for the next discussion" (HANDOFF_2026-05-07(1)). A scheduled reconcile tick is later referenced as existing in emit_design guard rationale ("Plan-time, not reconcile-time, so the scheduled reconcile tick does not backfill") — status conflict to resolve in stage 2.
- **what:** Two small deferred items from Phase-1 deployment: a heartbeat scheduled_tasks row producing periodic reconcile passes (mirroring content-feed-trigger), and moving the hardcoded ensure_pages page list into strategist/briefing-written site_specs read at plan time.
- **sources:** old2/HANDOFF_2026-05-07(1)#6–7, FOCUS_design_composition_flow_and_adoption_fidelity(1).md#3A
- **relations:** reconcile_site_plan; build-pipeline-trigger cadence
- **verify-later:** scheduled_tasks for a reconcile tick; ensure_pages config source

<!-- SOURCE: U10_imagery.md -->
### needs_section_data resolution: reconciler, not an agent
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** "SUPERSEDED 2026-05-06 by FOCUS_directory_builder_and_list_components.md"; "Update 2026-05-27… Decision: a full LLM handler agent (and the directory-builder agent) is not needed for the query-resolvable cases" — the two decided pieces not marked built.
- **what:** `needs_section_data` items are emitted at needs_human_review meaning "couldn't resolve component or required field", not async dispatch; 41 items were stuck system-wide. Resolution machinery already exists (`queryresolve.Resolve`, only `pages_where_type:<type>` implemented; `pages_under_section` named but absent from the dispatch switch). The settled design: (1) implement pages_under_section in queryresolve; (2) a section-data reconciler (a resolver, not an agent) re-attempting open items through existing machinery, closing via closeResolvedDataRequest and flagging re-renders; genuinely-human data (spec-sourced) stays HITL. The originally-planned dedicated handler agent and the never-built `directory-builder` agent are documented dropped ideas.
- **sources:** FUTURE_section_data_handler_1_.md (header supersession + 2026-05-27 update + original)
- **relations:** abandoned: directory-builder agent; relates to list components inventory (~17 components) and page-build-handler.
- **verify-later:** queryresolve.go dispatch switch; count of stuck needs_section_data items.

<!-- SOURCE: U12_docs024_archives.md -->
### site_plan page-role enum naming (underscore → hyphen; index → landing)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Archive: `"section_index" | ... | "blog_post"`; live: `"section-index" | ... | "blog-post" | "landing"`.
- **what:** `site_plan_pages.role` vocabulary was originally underscore-separated with a bare `index` role for the homepage. Renamed to hyphenated form and the homepage role renamed to `landing`, matching kebab-case conventions elsewhere.
- **sources:** old/029_site_plan_and_reconciler.md#"role table"; docs024_key_docs_latest/029_site_plan_and_reconciler(2).md#"role table"
- **relations:** page_type vocabulary and kebab constraint (016 §6.5)
- **verify-later:** DB check constraint on `site_plan_pages.role`/`pages.page_type` for hyphenated values.

<!-- SOURCE: U12_docs024_archives.md -->
### site_plan_partials — single JSONB-blob partial storage (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "JSONB blobs were considered and rejected because at anticipated scale... loading whole blobs to read one slice is wasteful, surgical HITL edits become hard, and lock transfer at meaningful granularity is impossible."
- **what:** Archived Phase 1 plan proposed one table, `site_plan_partials`, storing each partial as a single versioned JSONB blob per plan. Abandoned for two normalized row-per-thing tables — `site_plan_sections` and `site_plan_directives` — enabling per-row HITL locking at 1000+ page scale.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"schema section"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"schema section"
- **relations:** lock transfer across plan rebuilds; lazy per-page brief generation (also abandoned)
- **verify-later:** confirm `site_plan_directives`/`site_plan_sections` tables exist, `site_plan_partials` does not.

<!-- SOURCE: U12_docs024_archives.md -->
### Three sequential per-partial plan-builder LLM calls (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "Earlier draft of this doc proposed three sequential LLM calls. Looking at the existing build-site-planner agent, that lean was wrong."
- **what:** Archived plan proposed splitting the plan-builder into three sequential LLM calls for independent retry granularity. Abandoned once it was noticed the production build-site-planner agent already produces all three coherently in one call with no evidence of retry-granularity problems.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q2. Plan-builder LLM tier"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Q2. Plan-builder LLM call shape"
- **relations:** site_plan_partials (abandoned)
- **verify-later:** build-site-planner agent_definitions workflow — confirm single LLM call shape.

<!-- SOURCE: U12_docs024_archives.md -->
### Separate BuildPageURL path-resolver helper (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "The earlier draft of this doc proposed a separate BuildPageURL helper... That argument was overly cautious... Consolidated."
- **what:** Archived plan proposed a brand-new ~50-line Go helper sibling to `page_canonical.go`. Abandoned as overly cautious: Phase 1 instead extends `CanonicalisePage` additively with an optional `ParentSection` field.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q3. URL paths"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Q3. URL paths"
- **relations:** site_plan page-role enum naming
- **verify-later:** `datahelpers/page_canonical.go` — confirm `CanonicalisePage` has `ParentSection`, no separate `BuildPageURL`.

<!-- SOURCE: U12_docs024_archives.md -->
### Lazy per-page brief generation via build_page_brief step (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Archive rollout step 8: "build_page_brief step in page-build-handler... generates site_plan_partials/page_brief:<name> if missing." Live replaces with a pure-Go brief renderer.
- **what:** Archived plan generated each page's brief lazily via an LLM step during page build. Abandoned for a deterministic, non-LLM Go helper that assembles a brief at read time by walking the directive cascade and applying cardinality rules.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"rollout table, step 7-8"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Directive cascade and brief assembly"
- **relations:** site_plan_partials (abandoned)
- **verify-later:** confirm `datahelpers/page_brief.go` exists; page-build-handler has no `build_page_brief` LLM step.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### FAQ empty-items bug: duplicate content-surface planning (Defect 1)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "Deployed status (2026-05-21) ... Prevention shipped on three fronts, all live" with confirmed-live flags and chassis v1.0.1029
- **what:** Pages were planned with both a freeform `generic-text-block` and a structured component (`faq`, `pricing`) intended to hold the same content, because the content-gap-planner's prompt example hardcoded `generic-text-block` and the site-planner's mappings omitted faq/pricing entirely; the content writer (proven correct by an isolated build test) then filled the freeform block and left the structured component empty. Fixed by editing both planner prompts and an archetype-aware `defaultSectionsForPage` Go backstop.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md, js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#empty-shells, js_snippets_news_gaswholesalers/old/faq_empty_items_prevention_findings(1).md
- **relations:** Display-name leak (Defect 2); "Renders empty" diagnostic method; per-section briefs gap; extractResponseContent flat-string hypothesis (superseded)
- **verify-later:** content-gap-planner and site-planner agent_definitions prompt_template, apply_gap_plan_action.go defaultSectionsForPage, chassis v1.0.1029

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Display-name leak into section arrays (Defect 2) + validate_components resolver
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "validate_components implemented in ValidateSitePlanAction (was a dead flag)... deployed in chassis v1.0.1029"
- **what:** A planner path could emit a component's `display_name` instead of its kebab `function` into a page's `sections` array, orphaning the page_component. Fixed by implementing the previously-dead `validate_components` config flag in `ValidateSitePlanAction`: a `componentNameResolver` resolves each section name (exact match → NormalizeComponentFunction → display/name lookup → drop+log if unresolvable). The gap-planner path (`applyNewPage`) doesn't route through `validate_site_plan`, so the same resolver had to be wired in separately there too.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#Fix-B-implementation, js_snippets_news_gaswholesalers/old/validate_components_implementation.md
- **relations:** FAQ duplicate content-surface bug; per-section briefs gap
- **verify-later:** ValidateSitePlanAction, loadComponentNameResolver, NormalizeComponentFunction

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Per-section briefs gap (planner depth) — bare section-name strings, no intent
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "Planner depth: per-section briefs + stale-plan write-back... planner needs to emit them" listed under "Open — structural (not blocking)"
- **what:** `site_plan.pages[].sections` is an array of bare strings with no per-section brief. This is the deeper cause behind Defect 1: without a brief, the writer cannot tell that `faq` and `generic-text-block` are competing surfaces. A consumer already exists (`plan_sections.sectionDescription`) but the planner never emits any of those shapes. Token-budget caveat: adding briefs to every section on a large site materially grows planner output size.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#Fix-C-stale-plan, js_snippets_news_gaswholesalers/old/site_planner_depth_and_freshness_concerns.md
- **relations:** FAQ duplicate content-surface bug; Post-build validation of structured components; validate_components resolver
- **verify-later:** load_page_sections_from_spec, plan_sections.sectionDescription resolver

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Stale site_plan — gap-planned pages never written back (Concern 2)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "gap-planned pages aren't written back to site_plan (faq was absent from the plan entirely)... apply_gap_plan should append new pages to site_plan" — not yet implemented
- **what:** Pages created after initial site planning get a `pages` row and nav entries but are never appended to `site_specs.site_plan`; the plan drifts from reality with every gap-added page. Proposed fix: `apply_gap_plan` deep-merges the new page into `site_specs.site_plan` (mirroring `enrich_news_feed`'s pattern), plus a periodic plan-reconciliation discovery check.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#stale-plan, js_snippets_news_gaswholesalers/old/site_planner_depth_and_freshness_concerns.md
- **relations:** Per-section briefs gap; page content-creation build pipeline trace
- **verify-later:** apply_gap_plan_action.go, enrich_news_feed deep-merge pattern

<!-- SOURCE: U13_docs024_small_dirs.md -->
### site_plan as authoritative build source, overwriting pages.sections
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "`load_page_sections_from_spec_action.go` ... CONFIRMED in code" (PLAN_tool_widget_clobber(9).md §2.4)
- **what:** The page-build pipeline's section authority is `site_specs.site_plan`, not `pages.sections` directly — the loader syncs the plan's sections back into `pages.sections` on every build where a plan entry exists, only falling back to `pages.sections` if the plan yields nothing. Consequence: any fix that only sets `pages.sections` inside a tool action is futile once a plan entry exists; a durable fix must add the tool/embed section to the planner's `site_plan` output itself.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.4, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-2
- **relations:** Tool widget clobber mechanism (M1); Canonical tool-page section-shape design question
- **verify-later:** `load_page_sections_from_spec_action.go`; whether `site_plan` now carries a tool/embed section entry for `page_type='tool'` pages

<!-- SOURCE: U13_docs024_small_dirs.md -->
### queryresolve reality-vs-invention architectural promise
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Stated as an existing architectural line: "queryresolve exists specifically to draw a line between content the LLM is allowed to write... and content that has a database answer"
- **what:** A specific agent responsibility (`queryresolve`) enforcing a hard boundary in the site-build pipeline between LLM-authored creative content and database-derived factual lists — framed as central to the platform's "avoid fabrication" mission alongside carving the build into specialists with non-overlapping responsibilities.
- **sources:** pitch/003thebiggerpicture.md
- **relations:** Fractal agent architecture claim; Design/composition work-item emission gap
- **verify-later:** queryresolve action implementation; `source: query.*` convention in page_components

<!-- SOURCE: U13_docs024_small_dirs.md -->
### New-domain build pipeline stage chain (domain-submitter → page-build-handler)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** Traced live from code/DB snapshot: "Confirmed: ReconcileSitePlanAction reconciles pages only" and "The chain is fully connected" — caveated as read from a 2026-05-21 DB backup snapshot, "may have drifted"
- **what:** The confirmed happy-path chain for building a brand-new domain: `domain-submitter` → `domain-research-classifier` → `domain-strategist` → `build-briefing-agent` → `build-site-planner` (plan_site → validate → write_site_plan → sync_pages → populate_nav → reconcile_site_plan) → `page-build-handler` per page → `rerender-pages`. Driven by the 30s `build-pipeline-trigger` heartbeat, with every stage's `create_work_item` defaulting to status `triaged` so the pipeline self-advances.
- **sources:** plainjanedomain/README.md
- **relations:** Design/composition work-item emission gap; queryresolve reality-vs-invention architectural promise
- **verify-later:** live SELECT type, status, image_tag FROM agent_definitions WHERE type IN (domain-submitter, domain-research-classifier, domain-strategist, build-briefing-agent, build-site-planner)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Design/composition work-item emission gap (planner reorg unclosed seam)
- **category:** site-plan-and-reconciler
- **status-signal:** unknown
- **status-evidence:** "So nothing in the build path appears to emit a needs_design/needs_composition trigger for a fresh domain... consistent with this being an unclosed seam from the planner reorg"
- **what:** A discovered structural risk: the legacy `WriteBuildItemsAction` emitted the full item set for a new build (`needs_page`, `needs_logo`/`needs_hero_image`, `needs_composition`, `needs_design`), but the Phase-1 replacement (`build-site-planner` → `write_site_plan` + `reconcile_site_plan`) emits only `needs_page` + `needs_rerender`. The only fallback is the improvement-loop's `design-discovery-agent` catching `missing_css` later — meaning a new site could deploy pages referencing a stylesheet that doesn't exist yet.
- **sources:** plainjanedomain/README.md
- **relations:** New-domain build pipeline stage chain; Site-chrome rendering gap (dartsonline) — same class of defect
- **verify-later:** ReconcileSitePlanAction, WriteSitePlanAction, WriteBuildItemsAction Go source; design-discovery-agent missing_css check

<!-- SOURCE: U14_docs019_runbooks.md -->
### Roadmap-phases scope decision gap (nav grounded in built reality)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 6 "PROMOTED 2026-07-07 — THE BUG IS PLATFORM-WIDE … 082_submit_domain_unified.sh accepts ONLY --mission … AND build-site-planner's prompt has NO ELSE BRANCH for the roadmap-authority block … an absent decision point, not a missing default."
- **what:** No submitted site ever gets a roadmap/phases decision: the submit script has no --roadmap path and the planner's phase-discipline instructions vanish (not degrade) without one — so commerce-shaped domains get aspirational full plans and nav links to unbuildable pages. Fix shape (relay-wide, by construction): a post-classification scope-decision hop writes a phased roadmap_brief (P1 content/guides/tools; P2 legal-gated affiliate; P3 catalogue); planner prompt gains the ELSE branch (default phase-1-only or HITL hold); nav generation grounded in the BUILT set regardless of plan. Guidelines 001 already define the roadmap/phases mechanism — the docs had it, intake didn't. The legal gate on P2 is named as the fix-loop council's first concrete reviewer job.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 6); docs019/RUNBOOK_diagnosis_fix_loop(9).md#root-context
- **relations:** F0 guides pilot (nav-vs-built strand); coverage baseline; council compliance reviewer
- **verify-later:** 082_submit_domain_unified.sh flags; build-site-planner roadmap_brief template block; nav-updater

<!-- SOURCE: U15_docs019_running_notes.md -->
### Roadmap-phase enforcement gap (builder item 6)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "VERIFIED IN CODE: 082_submit_domain_unified.sh — grep confirms ONLY --mission/--mission-file exist, no --roadmap anywhere. build-site-planner prompt — the {{if .roadmap_brief}}...{{end}} block has NO else" (NOTES_running_fixloop(9).md).
- **what:** A platform-wide defect (reclassified from a single-site fix into the builder thread's main queue item) where no domain-submission path ever produces a phased roadmap, so a site's Tier-3 roadmap phase rules simply vanish rather than degrade — an absent decision point, not a hidden mechanism. Fix shape: a new post-classification hop writing a phased roadmap for commerce-shaped domains, enforced at three existing relay-wide points (strategist prompt, planner deliverability validation, built-grounded nav) rather than per-site.
- **sources:** NOTES_running_fixloop(9).md "TWO CORRECTIONS: amendment path under-specified; bug is platform-wide"; NOTES_running_synthesis_v4(39).md 2026-07-07 mirror entry.
- **relations:** Diagnosis→fix loop workstream founding; work-item relay / builder-generations architecture; curated best-in-class standing expectation.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Site plan as declarative artefact + reconciler
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 029(1) "The shape that fixes this is the same pattern Kubernetes uses: a declarative artefact … plus a reconciler … Phase 0 lands today"
- **what:** Fixes the duplicate-pages bug where two surfaces (adoption + site-planner) both wrote `pages` rows without a shared identity space. The planner writes a declarative desired-state plan; a deterministic Go reconciler (`reconcile_site_plan`, no LLM) walks desired-vs-realised and emits `needs_page:<name>` for the diff only.
- **sources:** WM/029_site_plan_and_reconciler(1).md#why-this-exists, WM/029_site_plan_and_reconciler(1).md#phase-1-plan-as-declarative-artefact-reconciler-emits-work, WM/030_phase1_plan_and_reconciler(4).md#plan-builder-cascade-replaces-todays-site-planner-emit-and-queue
- **relations:** CanonicalisePage; plan-domain schema; LLM tiering; drift auditors
- **verify-later:** reconcile_site_plan action; site_plan_structure/pages; pages.built_from_plan_version

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### CanonicalisePage + role validator (deterministic page identity)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 029(1) Phase 0 "A single canonicalisation helper in datahelpers/ … called from both surfaces"; 030(4) Q3 role validator (Go)
- **what:** A single `datahelpers/page_canonical.go` helper maps a `(role, slug, parent_section)` descriptor to a canonical `(name, url, page_type)` triple, called from both adoption and planner surfaces. Phase 1 extends it with `ParentSection` and adds a role-validator that corrects LLM role mislabels deterministically before persisting.
- **sources:** WM/029_site_plan_and_reconciler(1).md#fix, WM/030_phase1_plan_and_reconciler(4).md#q3-url-paths-canonicalisepage-phase-0-helper-extended-linknav-agents-own-drift, WM/016_debugging_guide_v2_44.md#adoption-faithfulness
- **relations:** site plan reconciler; architectural tension #1/#2; adoption faithfulness strip bug
- **verify-later:** datahelpers/page_canonical.go; ValidateRoles; CanonicalisePage

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Plan-domain schema + directive cascade + brief assembly
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030(4) Q1 "separate site_plans schema, not site_specs aspects … four plan-domain tables, all row-shaped for scale"
- **what:** Phase 1 rejects reusing `site_specs` aspects in favour of normalised plan tables (`site_plans`, `site_plan_pages`, `site_plan_sections`, `site_plan_directives`) row-shaped for 1000+ page scale. Guidance lives in `site_plan_directives` at site/page/section scope; a Go brief renderer (`datahelpers/page_brief.go`) walks the cascade and applies single- vs multi-valued cardinality.
- **sources:** WM/030_phase1_plan_and_reconciler(4).md#q1-plan-storage-separate-site_plans-schema-not-site_specs-aspects, WM/030_phase1_plan_and_reconciler(4).md#directive-cascade-and-brief-assembly, WM/030_phase1_plan_and_reconciler(4).md#what-stays-in-site_specs
- **relations:** site plan reconciler; lock transfer; strategic-vs-plan-time naming split
- **verify-later:** site_plan_directives; datahelpers/page_brief.go; write_site_plan action

<!-- SOURCE: U18_sql_for_agents.md -->
### site-planner (single-LLM-call site plan)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** 022 shows model flip-flops (sonnet→haiku for cost, 040 haiku→sonnet because planning is "high-leverage"); 053 build-site-planner is the successor for work-item builds.
- **what:** v2 planner: one LLM call over brief + component library + style collections producing validated_plan, pages, style_collection, needs_logo/needs_images. The model-choice oscillation (cost vs quality on high-leverage decisions) is documented reasoning worth keeping.
- **sources:** 022_site_planner.sql; sql_for_agents_v2/022_site_planner.sql; 040_optimise_which_llms.sql
- **relations:** chief-strategist (predecessor), build-site-planner (successor), pageflow-builder (caller)
- **verify-later:** which planner the live pipelines invoke

<!-- SOURCE: U18_sql_for_agents.md -->
### build-site-planner + roadmap-overrides-components rule
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 053 shows the workflow rewired to the site_plans domain ("changed to: ... write_site_plan → sync_pages → populate_nav → reconcile_site_plan → complete"); plan_site runs on claude-opus-4-6; 067 adds thinking budget.
- **what:** Handler for needs_site_plan. Reads site_specs (identity/classification/briefing/strategy), loads component library and style collections, plans via LLM, validates, then writes into the site_plans domain and reconciles. Carries the ROADMAP OVERRIDE rule verbatim: "ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase... use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list... Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components... The roadmap is the authority for this site." Earlier form wrote plan/design_intent/content_direction specs + write_build_items (one needs_content_write per page).
- **sources:** 053_build_site_planner.sql; 108_site_plan_pages.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** site_plans/reconciler domain (docs 029/030); component selector creating needs_new_component items; roadmap spec aspect
- **verify-later:** write_site_plan + reconcile_site_plan actions; roadmap aspect producer

<!-- SOURCE: U18_sql_for_agents.md -->
### site_plan_pages schema repair (plan-domain drift)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 108 "Migration 033: Reconcile site_plan_pages columns + drop orphan site_plan_partials... every write_site_plan call to date has failed at the title-column error."
- **what:** Repairs drift between two drafts of the site-plan schema: adds title/meta_description/nav_label columns, drops page_data and the unused site_plan_partials table (directives are row-per-directive in site_plan_directives). Documents the CREATE TABLE IF NOT EXISTS silent-skip failure mode when a rewritten migration follows an applied earlier draft.
- **sources:** 108_site_plan_pages.sql
- **relations:** build-site-planner; migration-discipline concepts (124)
- **verify-later:** live \d site_plan_pages / site_plan_directives

<!-- SOURCE: U19_sql_tables_components.md -->
### site_plans declarative plan domain
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Migration 031 (both drafts) with detailed rationale referencing doc 030; later tables (site_plan_imagery, work-item flows) depend on and reference it.
- **what:** The plan is a separate versioned artefact from site_specs: site_plans (version anchor, one is_current per site), site_plan_pages (row per planned page: canonical name/role/slug/url, parent_section for section-index detection, nav flags), site_plan_sections (structural per-section rows carrying resolved component_version/palette/layout/typography ids for HTML data-* provenance), site_plan_directives. Row-per-thing chosen over JSONB blobs for 1000+ page scale and surgical HITL edits; versioning mirrors site_specs (is_current + superseded_at).
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql
- **relations:** site_specs (strategic vs operational boundary); reconciler; naming note that plan_sections/save_page_sections actions "share a noun and nothing else".
- **verify-later:** write_site_plan action; plan row counts per site.

<!-- SOURCE: U19_sql_tables_components.md -->
### Directive cascade and HITL lock transfer
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 040 second draft: scope_ref encoding, cardinality lookup in brief renderer, "write_site_plan... transfers the lock onto the equivalent new directive row" matched by (scope, scope_ref, category, subject, ordering).
- **what:** Design/content/voice/structural guidance stored row-per-directive at site/page/section scope; a Go brief renderer walks the cascade (site → page → section) and emits prompt-ready text — consumers never read directives directly. Cardinality (override vs accumulate) is renderer knowledge, not schema. Human-locked directives survive plan rebuilds via stable-composite-key lock transfer performed only by write_site_plan.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#site_plan_directives
- **relations:** Pattern A locks; site_plan_imagery (same pattern); doc 030 "Directive cascade and brief assembly".
- **verify-later:** brief renderer helper; lock-transfer code in write_site_plan.

<!-- SOURCE: U19_sql_tables_components.md -->
### Plan drift detection and reconciler scheduling
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** pages.built_from_plan_version + sites.last_reconciled_at columns with reconciler semantics documented; later migrations reset built_from_plan_version=NULL to force rebuilds.
- **what:** Each built page records the plan version that produced it; the reconciler diffs site_plan_pages against pages, flags pages whose plan version lags current (NULL = never built under a plan), and emits needs_page/rebuild work items. sites.last_reconciled_at lets the scheduled tick skip recently reconciled sites; deliberately no FK so hard-deleted plans read as drift.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#4 and #5; docs/agent_docs/sql_for_tables/003_pages.sql#rebuild-flips
- **relations:** site_plans domain; site_work_items; scheduler.
- **verify-later:** reconcile_site_plan action; scheduled reconciler task.

<!-- SOURCE: U19_sql_tables_components.md -->
### site_plan_partials with lazy page briefs (early plan shape)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** First draft of migration 031 defines site_plan_partials ('design_direction', 'content_strategy' eager; 'page_brief:<name>' lazy via build_page_brief); the second draft in the same file replaces it with site_plan_sections + site_plan_directives.
- **what:** The initial plan-domain design stored design direction, content strategy and per-page briefs as versioned JSONB partials, with lazy page briefs written on demand by page-build-handler. Superseded by the row-per-section/row-per-directive shape for scale and surgical edits.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#site_plan_partials
- **relations:** superseded by site_plan_sections + site_plan_directives.
- **verify-later:** whether site_plan_partials exists in production or only the directive shape shipped.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Multi-page site support (wrap_multipage, multipage-site-builder)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** 030 SQL creates multipage-site-builder (index/about/contact + privacy); 031 shows the wrap_multipage step after html_assembler with CollectedData trace; today's pages/site_plans domain is the successor.
- **what:** Extending the single-page pipeline to small multi-page sites: after assembly, a wrap_multipage action derives index/about/contact (and privacy) pages, and the deployer commits all files. The first step from "landing page generator" toward the current multi-page site model.
- **sources:** docs004_website_capture_project/007different_types_of_site/030_about_page_and_privacy.sql; docs004_website_capture_project/007different_types_of_site/031_about_page_multipage_site.md
- **relations:** successor: site_plans / pages domain (site-plan-and-reconciler docs 029/030); robot-hands 3-page build (earlier sibling).
- **verify-later:** wrap_multipage in registry.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Three section sources for a page build (aspect → pages.sections → plan tables)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Workflow dump + code read 2026-07-06: "load_spec_sections... reads site_specs aspect site_plan (AUTHORITATIVE) → fallback page_record.sections. The site_plan_sections TABLE is NOT read by this path."
- **what:** Page builds resolve their section list from, in order: the `site_specs` aspect `site_plan` (legacy blob, 5 sites carry one; vonc has none), `pages.sections` (jsonb fallback — what actually serves vonc; the newer planner dual-writes plan tables → pages.sections), and same-role sibling synthesis; the `site_plan_sections` table is written by the vonc-generation planner but not read by the build path. Three peer stores with unclear precedence caused ten silent no-op builds and two fixes landing in the wrong store (a plan-table row, then the pages.sections UPDATE that finally unblocked).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3; docs/RUNBOOK_phase2_provocation_js(29).md#update-2026-07-06
- **relations:** plan storage authority (029 Q1); complete_error silent no-ops; load_page_record lookup semantics
- **verify-later:** load_page_sections_from_spec_action.go source order; SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE 'site_plan%'

<!-- SOURCE: U23_docs_root_vonc.md -->
### Plan storage authority — 029 Q1 and the withdrawn table-first alteration
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** PLAN_dynamic_sections(4): "SUPERSEDED (2026-07-06, same day) — decision deferred to 029 Q1; alteration WITHDRAWN"; "Decision closed (2026-07-07): the user chose REVERT."
- **what:** After the silent no-ops, a decision was made (then withdrawn the same day) to make the `site_plans` family the authoritative plan store and alter `load_page_sections_from_spec` to read site_plan_sections first. Reading design doc 029 showed plan storage is its OPEN Q1 ("site_specs aspects vs new table", lean = partitioned site_plan_* aspects + a reconcile_site_plan action); three shapes coexist in production (legacy site_plan blob aspect ×5 sites; 029 partitioned aspects apparently unimplemented; the vonc-generation tables with pages.sections dual-write). The alteration was withdrawn and the repo file reverted (ORIGINAL.go; cluster reverts on next chassis push); evidence contributed to Q1: the table path now exists in production post-dating the lean. Store-agnostic preventions retained. Earlier draft (v2 of the plan) also named a `site_plan_directives` child table not mentioned in the final version.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#decision + #superseded; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-alteration-withdrawn + #2026-07-07-revert-decision; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3
- **relations:** three section sources; reconcile_site_plan (029); planner ≥1-section invariant
- **verify-later:** git history of load_page_sections_from_spec_action.go (reverted?); repo grep reconcile_site_plan; docs024 029 doc Q1 status

<!-- SOURCE: U23_docs_root_vonc.md -->
### Planner role-aware ≥1-section invariant + role→pipeline mapping
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** Backlog item 1 in HANDOFF §9; "Invariant refined: every planned page whose ROLE is built by page-build-handler must have ≥1 section" (Gate B, 2026-07-06) — nowhere claimed built.
- **what:** The June planner emitted all 8 vonc pages but skipped SECTIONS for exactly the two non-standard roles — blog-post (legitimate: the blog pipeline builds those) and section-index (the defect that caused the archive 404). Prevention: at plan-store time, every planned page whose role page-build-handler owns must have ≥1 section, with the role→pipeline mapping made explicit; plus auditor drift rule (pages.sections vs current plan) and post-deploy URL-presence checks per active page.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#gate-results; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-i-gate-results
- **relations:** complete_error family; section descriptor design; quality-auditor rules
- **verify-later:** site-planner agent_definition; site_plan_pages roles for recent sites

<!-- SOURCE: U23_docs_root_vonc.md -->
### Autonomous section composition — per-section descriptor {role, kind, data_feed}
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** PLAN_dynamic_sections_and_loaders(4) status "DESIGN"; gaps list "(1) Section descriptor... Without this the framework can't tell static from dynamic" — none of gaps 1–5 marked built.
- **what:** The framework (not a human) should decide, from the domain/site-spec, which sections a page has, each section's role (to prevent overlaps like provocation-card's mini-lobby vs lobby-grid), whether it is static (build-time content) or dynamic (runtime-filled from a feed), and which named feed — encoded as a per-section descriptor `{component_name, role, kind, data_feed}` on the plan, written by the site-planner, consumed by build AND maintenance flows. The plan not carrying `kind` is why the assembler dropped the runtime-filled shells. Includes a spec-level feed catalogue and quality-auditor maintenance detections (dropped-dynamic, overlap, deferral, empty-dynamic). The root design point: a data-driven component should DECLARE its runtime data dependency so the pipeline provisions feed + loader automatically.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#the-question + #structural-gaps; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-reframed; docs/RUNBOOK_phase2_provocation_js(29).md#how-a-component-should-declare
- **relations:** Tier E runtime-feed tier; loader-builder agent; static-vs-dynamic distinction; plan storage authority (where the descriptor lives follows 029 Q1)
- **verify-later:** site_plan_sections columns (kind/data_feed/role exist?); site-planner prompt/workflow

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### validate_components section-name resolver
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** validate_components_implementation: "Implements the dead `validate_components: true` flag … currently set true for site-planner but never read"; provides `loadComponentNameResolver` and a gated block for `ValidateSitePlanAction`; describes deploying and testing via the isolated-build harness (implies not yet live).
- **what:** A deterministic resolver that maps each site-plan section name to a real `content_components.function` — via normalisation, display-name lookup ("FAQ Section"→`faq`), and name lookup — dropping+logging unresolvable names so they don't orphan downstream `page_components`. Deliberately narrow: it does NOT deduplicate or make intent decisions (that's the planner prompt + per-section briefs). Must also run in `applyNewPage` (content-gap-planner path bypasses validate_site_plan).
- **sources:** js_snippets_news_gaswholesalers/old/validate_components_implementation(1).md#scope, #2-the-validation-block, #3-the-gap-planner-path
- **relations:** NormalizeComponentFunction; per-section briefs; content-gap-planner; component schema drift
- **verify-later:** ValidateSitePlanAction validate_components flag read; apply_gap_plan_action.go applyNewPage

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### SyncPagesToDBAction / WriteSitePlanAction canonicalisation divergence — Option 1 rejected
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 2: "**Option 1 (single source of truth):** sync reads identity from `site_plan_pages`... **Decision: Option 2**... Corrected an earlier framing that called Option 1 'the structural one' — Option 2 is the structural fix here; Option 1 is coupling."
- **what:** Two canonicalisation surfaces disagreed — `WriteSitePlanAction` ran `ValidateRoles + CanonicalisePage` (producing correct `section-index` hubs in `site_plan_pages`), while `SyncPagesToDBAction` ran `CanonicalisePage` alone on raw `page_plan` (producing flat `pages` rows), and `upsertPage`'s `ON CONFLICT` then overwrote the correct row with the flat one. Option 1 (make sync read the already-validated `site_plan_pages`) was rejected because `pageflow-builder` (confirmed active) and two other callers invoke sync with no plan ever written, so Option 1 would silently break them. The shipped fix (Option 2) runs `ValidateRoles` inside sync too, unifying the pipeline across all five callers.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 1–3
- **relations:** `pageflow-builder` deprecation (decoupled from this fix, tracked separately), guide page_type restructuring
- **verify-later:** `SyncPagesToDBAction`/`site_db_actions.go` current state; whether `pageflow-builder` was ever actually deprecated.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Adoption-faithfulness-via-locks convergence — confirmed INERT
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** running_notes_14(20) Part 14h: "TRUE root cause: `reconcilePlanWithRealised` gates on `rm[\"adoption_locked\"]`; the live `load_existing_pages` query does NOT emit `adoption_locked` ... lockedPages always empty -> reconcile ALWAYS no-ops." And: "`FOCUS_adoption_faithfulness_via_locks.md` status — convergence 'Inert until 054 + write_site_plan land.' ... LIVE STATE: lock tables have ONLY `locked_at`/`locked_by` — NO `lock_type`/`lock_expires_at` -> 053 NOT applied... 054 NOT applied."
- **what:** A designed subsystem meant to make adoption re-plans faithful to already-realised (locked) pages — schema migration 053 (lock_type/lock_expires_at columns), migration 054 (`load_existing_pages` emits `adoption_locked`), and `write_site_plan` locking logic — was found, on live inspection, to be entirely unapplied. The one piece that *was* built (`reconcilePlanWithRealised`'s convergence check in `v3_site_actions.go`) silently no-ops because its input is never populated. This directly explains two other defects in the same arc (the bare-guide duplicates, and 5 guide pages never being unioned into the plan).
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 14h; (references live `FOCUS_adoption_faithfulness_via_locks.md`, `031_locks(3).md`)
- **relations:** bare-guide duplicate pages; sync/write-site-plan divergence
- **verify-later:** whether migrations 053/054 have since been applied; current state of `write_site_plan_action.go`'s `transferDirectiveLocks`.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Deployed→needs_rebuild ON CONFLICT flip — pre-design stand-in later completed properly (Option B)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 8: "the flip is a pre-design *stand-in* for 're-sync invalidates deployed pages'... It over-fires (every deployed page, every sync) and mis-fires on pre-plan deploys (tools)... **Option B shipped**: COALESCE fill-if-null; removed the `deployed→needs_rebuild` CASE branch... Drift now flows through the reconciler's `decideEmit`."
- **what:** `upsertPage`'s `ON CONFLICT` branch that flipped any `deployed` page back to `needs_rebuild` on every sync was a workaround for a never-shipped design: `029`/`030` intended `built_from_plan_version` to be stamped at build time and drift detected by the reconciler, but the stamp was "explicitly deferred" per `HANDOFF_2026-05-07` #5 ("User explicitly OK'd this"). The investigation confirmed the flip should be completed as originally designed rather than patched around (rejecting a narrower "Option A: exclude tool/game from the flip" as entrenching the workaround) — shipped as the deploy-time stamp in `UpdatePageStatusAction` + COALESCE fill-if-null in `upsertPage`.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 8; CATALOGUE_gamesdesign_post_sync_fix_defects(4).md A1
- **relations:** A1 tool/game deploy-gap root cause (below)
- **verify-later:** `v3_site_actions.go` `UpdatePageStatusAction`, `site_db_actions.go` `upsertPage` current state.

<!-- SOURCE: U25_leopardess_social.md -->
### Page section source precedence and the plan-storage triple shape (029 Q1)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** HANDOFF §3: "Section sources … site_specs aspect site_plan (authoritative in code) → pages.sections (fallback) → sibling synthesis. The site_plan_sections table is NOT read by this path"; "Three shapes coexist in production … The decision belongs to the planner/reconciler thread."
- **what:** A page build reads sections from the site_specs 'site_plan' blob aspect first, then pages.sections, then same-role sibling layout synthesis — while the newer site_plans/site_plan_sections tables (which the vonc-generation planner writes, dual-writing pages.sections) are ignored by this path. A drafted table-first alteration was consciously withdrawn pending design doc 029's open Q1 (aspects vs table). Operational corollaries: the provocations-index unblock was a pages.sections UPDATE; reconcile_site_plan re-emits needs_page for any planned-but-unbuilt page every run (the standing needs_page:provocation trap — park it to detected after every vonc reconcile).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #9.7; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 (needs_page:provocation)
- **relations:** silent no-op success class; archetype hub build (used reconcile_site_plan properly); docs024 029/030
- **verify-later:** load_page_sections_from_spec_action.go; SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE 'site_plan%'; reconcile_site_plan grep

<!-- SOURCE: U26_misc_dirs.md -->
### Website-builder agent group (six-specialist pipeline)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Ran in production per basic_usage/001 ("Step-by-Step Guide to Your First Website Build", migrations 005/007/009 referenced); the current platform builds sites via the site_plans domain / webdesign-agent pipeline (002 spine, docs 029/030), which replaced this group.
- **what:** The original end-to-end website creation flow: an orchestrator agent calls domain-analyst (business categorisation via web-search) → site-architect (page structure, pausing for human approval) → fan-out of content-researcher + visual-designer (image search/generation, logo) → html-developer (per-page vanilla HTML/CSS fan-out) → site-publisher (s3_upload, preview URL). Seeded as agent_definitions + an agent_groups row; triggered by one spawn_group Kafka message.
- **sources:** docs/architecture/027-create-website-creation-system; docs/basic_usage/001basic_usage.txt; docs/basic_usage/003_dynamic_prompt_improvement#step-1.1
- **relations:** superseded by site_plans + webdesign-agent + design-composition pipeline; HITL pause in site-architect; result storage split
- **verify-later:** migrations 005/007/009 in platform/database/migrations/; whether group still seeded

<!-- SOURCE: U01_docs024_numbered_core.md -->
### built_from_plan_version deploy-time stamp replaces the deployed→needs_rebuild flip (Option B)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 016 §9 dedicated entry (2026-05-28) "Fix shipped", completing the HANDOFF_2026-05-07 deferred design
- **what:** upsertPage's blunt flip (deployed→needs_rebuild on every sync) stood in for the unbuilt drift stamp; Option B stamps built_from_plan_version at the UpdatePageStatusAction deployed chokepoint and makes sync fill-if-null, retiring the flip so drift detection flows through the reconciler's decideEmit. Lesson (checklist 22): a "bug" may be a half-implemented design — complete it, don't patch around it.
- **sources:** 016 §9 flip entry; 029/030 design
- **relations:** reconciler; tool-page churn
- **verify-later:** any direct build_status='deployed' writes bypassing the action

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Adoption slug-mangling: two canonicalisation surfaces must agree
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 016 §9 chain of entries (2026-05-19→26): cause pinned to WriteSitePlanAction ValidateRoles strip + SyncPagesToDBAction canonicalising raw page_plan WITHOUT ValidateRoles; fix "CHOSEN" (option 2) not confirmed shipped
- **what:** ValidateRoles strips tool-/guide-/game- prefixes and -index; CanonicalisePage re-adds them only for tool/game/guide roles, so wrong page_types (hubs typed content, guides typed blog-post) permanently flatten names/URLs. sync_pages_to_db reads raw page_plan (not site_plan_pages), skips ValidateRoles, and its ON CONFLICT overwrites correct adoption-time rows — one logical page, two writers, divergent results (incl. tool-game-* double prefixes). Fix: run the identical ValidateRoles+CanonicalisePage pipeline in sync (works for all five callers incl. plan-less pageflow-builder); root fix upstream is correct page_type at adoption; endgame is 029's deterministic slug preservation.
- **sources:** 016 §9 three linked entries; 030 phase-0 result
- **relations:** CanonicalisePage; page_type vocabulary
- **verify-later:** SyncPagesToDBAction ValidateRoles call present?

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Plan as declarative artefact + reconciler (Kubernetes-style desired-vs-realised)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030: Phase 0 done (2026-05-04 re-adopt verified dedup); Phase 1 schema/decisions committed; 007 patch describes reconciler emitting needs_page as current behaviour
- **what:** The planner stops emitting work items; it writes desired state to plan-domain tables (site_plans one-current-per-site, site_plan_pages, site_plan_sections, site_plan_directives) and a deterministic Go reconciler diffs plan vs pages and emits idempotent needs_page items (with preference weights, cycle budget, dependency ordering). Fixes the two-writer duplicate-pages structural bug (adoption + planner not sharing identity space). Phase 2: discoverers/auditors read the plan for sharper fitness checks.
- **sources:** 029 full; 030 full; 007_adoption_pipeline_v4.patch
- **relations:** CanonicalisePage; built_from_plan_version; directives
- **verify-later:** site_plans tables live; reconciler action name

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Strategic vs plan-time guidance split (site_plan_directives)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030 Q1/Q2 decisions + 007 patch stating planner "no longer overwrites" adoption specs; lock-transfer designed
- **what:** site_specs.design_intent/content_direction stay strategic (classifier/adoption-owned); the planner's per-build guidance flattens into row-shaped site_plan_directives (scope site/page/section, category, subject, directive, source, Pattern-A locks) read by downstream agents via a brief renderer. One LLM call still produces structure+design+content together (coherence over three-call split); only the write targets change. HITL locks transfer across plan rebuilds by composite key inside write_site_plan.
- **sources:** 030#Q1/Q2, #Strategic vs plan-time naming; 031(3)#Lock transfer
- **relations:** B-029-4 design-intent clobber (motivating bug); lock transfer
- **verify-later:** site_plan_directives populated; brief renderer helper

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Per-page brief generation (lazy) and the no-empty-slots acceptance test
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** 029 B-029-2 promotes it to Phase-1 acceptance test; briefs "lazy" design section
- **what:** Component templates are named slots; a per-page brief enumerates slot content, generated lazily at build time. Without briefs, component-author defaults leak (empty img src, /services.html CTAs on sites without services). Acceptance: a Phase-1 build produces no empty slots and no leaked defaults — unbriefed slots either don't render or error before deploy.
- **sources:** 029#B-029-2, #Per-page brief generation
- **relations:** directives; B-029 bug list (dup nav items; theme vars never written)
- **verify-later:** brief generation exists?

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Architectural Tension #1 — infer-and-repair vs deterministic structure derivation
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Status 2026-05-25: Tension #1 has a deployed partial fix (Part A — ValidateRoles -index rule), pending a clean production test"
- **what:** The pipeline takes structural decisions (page role/type/URL) from LLM free-text labels then repairs with starved, vertical-hardcoded heuristics, producing silent structural corruption (section hubs flattened to content). Resolution principle: derive structure deterministically from the LLM's reliable signal — naming (`<section>-index` marks a hub, vertical-agnostically); schema-constrain generation to kill form errors (necessary but not sufficient); make fallback heuristics fail loud, never default to content. Explicit recommendation AGAINST a free parent-pointer tree (worst LLM reliability tier); a leaf's section, if needed, is a constrained choice over the enumerated hub set.
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-1; HANDOFF_2026-05-26 (page_type re-type as an instance)
- **relations:** Tension #2; page_type vocabulary gap; LLM reliability strategy (same principle, component scale)
- **verify-later:** ValidateRoles -index rule and de-hardcoded nestedRoleFromURL in page_role_validator.go

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Architectural Tension #2 — page identity derived in multiple places that undo each other
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Tension #2's residual confirmed cosmetic (see HANDOFF_2026-05-25)" but flavour-collapse residual "evidence-gated, not yet a code change" (2026-05-25)
- **what:** Adoption, planner-write and convergence each re-derive canonical page name/role/URL with no single owner, so a later stage can undo an earlier correct result (convergence preserved games-index; WriteSitePlanAction flattened it one step later). Principle: one canonical owner; canonicalisation idempotent on already-canonical input; downstream reads identity read-only. Part A made section indexes round-trip cleanly; the remaining residual is flavour collapse (validator emits generic section-index, losing blog-index/entity-directory flavour) — decide from a deployed run whether the component resolver needs the flavour before writing preservation code. Withdrawn: merging the two role-normalisers (intentionally layered).
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-2; HANDOFF_2026-05-26 (write vs sync canonicaliser divergence)
- **relations:** Tension #1; kebab/snake; canonicaliser divergence
- **verify-later:** CanonicalisePage/normaliseRole/normalisePageType in datahelpers/page_canonical.go; component resolver's page_type dependence

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site_specs vs site_plan two-layer architecture + aspect ownership contract
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "build-site-planner workflow writes both shapes during transition (old site_specs/site_plan aspect AND new plan tables)" (undated FOCUS, references docs 028-030)
- **what:** site_specs = strategic, brand-level, slow-changing, one owning agent per aspect (classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planner owns the four plan tables). site_plan tables = per-build, row-shaped, rebuilt per plan. Three ownership rules (don't read what you didn't spec; don't overwrite another's aspect; write outputs to the spec) with the classifier read-and-extend carve-out. Decision rules and anti-patterns for where new data lives (specs vs directives vs sibling structured tables).
- **sources:** FOCUS_site_spec_vs_site_plan.md (whole); ASSESSMENT_imagery_phase_0_1…md#What-Phase-1-changes
- **relations:** directive cascade; lock transfer; imagery placement
- **verify-later:** site_plans/site_plan_pages/site_plan_sections/site_plan_directives tables; legacy site_plan aspect readers (pageflow-builder)

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### site_plan_directives cascade + brief renderer
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "Reconciler is documented in doc 030 but the chassis-side implementation has been landing in stages"; brief renderer named as `datahelpers/page_brief.go` "per the work order"
- **what:** Cross-cutting guidance rows located by (scope site/page/section, scope_ref, category, subject) with HITL lock columns. Consumers never read rows directly: a Go brief renderer cascades site → page → section and applies cardinality semantics (single-valued subjects override at narrower scope; multi-valued accumulate), emitting short LLM-ready briefs. The pattern imagery/text/design guidance should all follow.
- **sources:** FOCUS_site_spec_vs_site_plan.md#directives; ASSESSMENT_imagery_phase_0_1…md#Amendments
- **relations:** lock transfer; site_plan_imagery sibling-table pattern
- **verify-later:** datahelpers/page_brief.go existence and consumers

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### HITL lock transfer across plan rebuilds
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** described as run "inside write_site_plan" per doc 030; extended for imagery + lock_type/expiry per 2026-05 patches ("transferDirectiveLocks carries lock_type/expiry — written (patch doc)")
- **what:** On plan rebuild, locked directives from the previous current plan are matched to new rows by composite key (scope, scope_ref, category, subject, ordering); locked_at/locked_by and HITL-edited text copy over (HITL wins); unmatched locks drop with a log, previous plan kept as history. Any sibling table wanting HITL adopts the same shape.
- **sources:** FOCUS_site_spec_vs_site_plan.md#Lock-transfer; FOCUS_adoption_faithfulness_via_locks(2).md#dependency-chain
- **relations:** adoption-faithfulness timed locks; site_plan_imagery
- **verify-later:** transferDirectiveLocks in write_site_plan action code

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Section-data deferral + reconciler loop
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "reconcile_section_data_action.go — new, not yet wired to a host"; pages_under_section implemented (2026-06-02)
- **what:** query.*-sourced section fields unresolvable at plan time defer as needs_section_data; the queryresolve package (pages_where_type, now pages_under_section joining site_areas) resolves them; a lightweight reconciler (not an LLM agent — the once-planned directory-builder was never built) rescans open items whose missing fields are all query-sourced and emits needs_page re-renders (dedup key page_rerender:<page>), leaving human-data items (team, pricing) in HITL. plan_sections closes items on re-render. Host (loop check or post-build finalize) still to pick.
- **sources:** HANDOFF_2026-06-02…md#2; FOCUS_internal_linking.md#4; HANDOFF-pipeline-triage-april-2026.md P5
- **relations:** P5 plan-then-reconcile; list hubs; self-contained components heuristic gap
- **verify-later:** reconcile_section_data host + registry entry; queryresolve switch cases

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### page_type vocabulary gap forcing game→tool re-type (Gap B)
- **category:** site-plan-and-reconciler
- **status-signal:** unknown
- **status-evidence:** "root cause is confirmed from the planner's response_text … there is no `game` [in the Canonical Page Types list], so every adopted game is forced to `tool`" (2026-05-26); "OPEN structurally; may have been addressed by the other-chat fixes … Verify post-deploy"
- **what:** The plan_site prompt's closed page-type list lacks `game`; the LLM keeps names faithfully but re-types game pages as tool; canonicalisation's tool branch then renames, and a page_type change (not a name change) is what duplicates pages — 5 duplicate game-*/tool-game-* pairs on gamesdesign. Also exposed: WriteSitePlanAction and sync_pages_to_db canonicalise the same tool-typed page differently (tool-auto-battler vs tool-game-auto-battler) — code read required before fixing. Verification queries recorded (stem-grouped pages; response_text page_type; composition install).
- **sources:** HANDOFF_2026-05-26…md#diagnosis, #Where-to-resume
- **relations:** Tension #1/#2; games content type; adoption faithfulness locks
- **verify-later:** run the three handoff queries on a post-2026-05-26 adoption; page_canonical.go call sites

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Section data source triad and reconcile_section_data
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** HANDOFF (2026-06-19): "`reconcile_section_data` IS wired — registry.go line 914 … description 'Re-trigger pages whose deferred section data is now query-resolvable'" (correcting a stale note that it was not wired).
- **what:** A component's content comes from one of three sources, and fixes differ per case: (1) query-resolvable section data (the tools/guides-list kind — the reconciler's scope: `ReconcileSectionDataAction` re-triggers pages whose deferred data has become resolvable), (2) a human-entered spec field (e.g. pricing tier_1_* from `site_specs.pricing` — the reconciler correctly skips these), (3) page-content-writer prose (LLM-generated). The differentiators investigation established the triad as the diagnostic frame — and then found the actual fault was in none of the sources (a key-naming mismatch). Incidental same-thread finding: `write_site_spec` errors "missing required fields: [spec_data]" on persist_mission/roadmap — the action input is spec_data but the column is `data` (site_specs is aspect + data jsonb, UNIQUE(site_id,aspect) WHERE is_current).
- **sources:** HANDOFF_idea_uk_differentiators_section_data.md; bundle3; running_notes_scheme_to_components(55).md#Sa #Sh (corrected facts)
- **relations:** array item-fields contract (the real fault); plan_sections deferral.
- **verify-later:** reconcile_section_data_action.go scope logic; registry.go wiring; site_specs schema.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Planner re-plan union safety (normaliseRealisedToPlanPage)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Checkpoint (Un) 2026-07-07: "normaliseRealisedToPlanPage (v3_site_actions.go:4383) exists so a re-plan LOADS realised pages …, converts them to plan-page shape CARRYING their sections, and UNIONS with the LLM proposal — its own comment: without carrying sections the upsert would clobber built pages."
- **what:** Site composition is whole-plan and LLM-driven: build-site-planner (consuming needs_site_plan) supersedes the current site_plans row and rewrites site_plan_pages + site_plan_sections. Re-running it is safe by design because load_existing_pages surfaces realised pages and the normaliser unions them (with their sections) into the new plan — built pages keep their composition while catalogued-but-uncomposed pages get composed. This makes "emit needs_site_plan" the structural route for composing missing pages, versus hand-INSERTing plan rows (which drifts nav/plan/page consistency).
- **sources:** running_notes_scheme_to_components(55).md#Un; stepF_replan_read.sql
- **relations:** planned-but-uncomposed pages gap; work-item crafting conventions.
- **verify-later:** v3_site_actions.go normaliseRealisedToPlanPage; build-site-planner workflow steps.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Planned-but-uncomposed pages gap (catalogued, never composed)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** Checkpoint (Ul): "the three planned pages have NO site_plan_sections rows; their pages.sections = []. Catalogued, never composed"; (Un) ends with the replan-read staged — the emit had not run at the unit's last dated note (2026-07-07).
- **what:** A distinct failure shape: pages rows exist with page_type and nav intent set (news-index, guides-index, tool-audience-check on idea.uk), so navigation links to them and 404s, but they carry empty sections and no plan rows — the LLM plan behind the current site_plans row never included them. A W6-style needs_page emit would build an empty page; the correct route is two-phase: planner re-run composes them (union-safe), then needs_page builds and deploys. Also surfaced the distinction between query-backed index pages (news/guides may be fed by the blog-listing mechanism) and static pages, and reuse of the already-deployed audience-check tool component.
- **sources:** running_notes_scheme_to_components(55).md#Uk #Ul #Um #Un; RUNBOOK_scheme_to_components(50).md#PLANNED-PAGES; stepD_and_pages_reads.sql (block B/C)
- **relations:** planner re-plan union safety; navigation (nav 404s); rebuild vs rerender.
- **verify-later:** idea.uk pages rows for the three; site_plan_sections presence; whether the needs_site_plan emit ran.

<!-- SOURCE: U04_idea_uk.md -->
### Section-data reconciler and the human-sourced-field boundary
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "reconcile_section_data IS wired (registry.go L914… 're-trigger pages whose deferred section data is now query-resolvable')" — correcting an earlier stale "built but unwired" note (rr, 2026-06-19).
- **what:** Deferred section data (needs_section_data) is re-triggered when it becomes *query-resolvable*; the boundary concept: **human-sourced** spec fields (e.g. pricing tiers from site_specs.pricing) are not query-resolvable, so the reconciler can never fill them — either capture the data into specs (the £29 into pricing) or the section shouldn't be on the page. The unresolved-CTA gating (render no button when no eligible destination page exists) is the same honest-degradation family, tied to the thin 4-page plan having no hub pages.
- **sources:** idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md (open items + empty-index content gaps); idea.uk/README_001_todo_list.md
- **relations:** item_fields fix; site-plan thinness; content-governance (pricing spec).
- **verify-later:** reconcile_section_data_action.go host wiring; idea.uk pricing spec.

<!-- SOURCE: U05_content_quality_linking.md -->
### Section-index hub canonicalisation divergence + plan-version stamping
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 5: "Both the core fix … and the companion … are confirmed. Thread closed."; Part 9/10 A1 VERIFIED CLOSED.
- **what:** Two canonicalisation surfaces disagreed: WriteSitePlanAction ran ValidateRoles+CanonicalisePage (hubs → section-index, nested URLs) while SyncPagesToDBAction ran CanonicalisePage alone on the raw page_plan — flattening hubs on every sync. Fix (Option 2): sync runs the identical pipeline (Option 1 — read site_plan_pages — rejected because active callers have no plan at sync time). Companions: built_from_plan_version stamped at deploy time in UpdatePageStatusAction (completing the deferred doc-029 design), upsertPage COALESCE fill-if-null, and removal of the deployed→needs_rebuild flip (a pre-design stand-in that over-fired).
- **sources:** running_notes_14(26).md#part-1-3, #part-8; site_db_actions/upsertPage references throughout
- **relations:** reconciler drift detection; adoption faithfulness convergence; A1 tool deploy failure.
- **verify-later:** SyncPagesToDBAction ValidateRoles call; UpdatePageStatusAction stamp; reconciler decideEmit.

<!-- SOURCE: U05_content_quality_linking.md -->
### Adoption-faithfulness convergence + the []map type-assertion keystone bug
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14n: "CONVERGED … the convergence/duplicate-page root cause … is RESOLVED on a clean run" (2026-06-05 17:26).
- **what:** The reconcile-plan-with-realised subsystem (Pass A unions adoption-locked realised pages missing from the LLM plan; Pass C2 drops planned pages whose topic stem collides with an existing page) had NEVER functioned since deploy: ValidateSitePlanAction asserted existing_pages as []interface{} while QueryDatabaseAction returns []map[string]interface{} — the assertion always failed silently, so convergence no-op'd for every site (bare-sibling guide duplicates, guides absent from plans). Fix: type-switch both shapes + a count log so an empty set is never silent; plus normaliseRealisedToPlanPage carrying sections/meta/nav_order so the union can't clobber adopted pages to empty (the union-clobber that had emptied the source-populated hubs). Multiple interim framings (054 not applied; lock-window) were corrected en route — 053/054 were live; the killer was the type bug.
- **sources:** running_notes_14(26).md#part-14h-14n
- **relations:** locks (adoption_locked first-plan branch; 90-day replan window non-functional); planner sibling-invention; empty-hub union clobber.
- **verify-later:** ValidateSitePlanAction extraction switch; reconcilePlanWithRealised counters in planner logs.

<!-- SOURCE: U09_adoption.md -->
### First-plan branch: "no current plan + pages exist ⇒ adopted pages"
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "054 `load_existing_pages` — partially live. The query emits `adoption_locked` but only via the first-plan branch: CASE WHEN NOT EXISTS (current is_current plan for this site) THEN true" (2026-06-05 verified landed state).
- **what:** Deterministic detection of the faithful first pass: when `load_existing_pages` finds no current site_plan but pages exist, all existing pages are flagged `adoption_locked=true` (only ever true after adoption; from-scratch sites have no pages before the planner's own sync). Convergence keys off this flag; a re-adoption from a cleared DB (or retiring the current plan) makes any site a "first pass" deterministically.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#verified-landed-state, verify_readoption_fix.sql, running_notes_14(25)#part-14i
- **relations:** reconcilePlanWithRealised convergence; verify_readoption gate G1/G2 (retire current plan to force first pass)
- **verify-later:** live load_existing_pages SQL in build-site-planner def

<!-- SOURCE: U09_adoption.md -->
### Planner ignores adopted state (generic-skeleton overlay)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Diagnosed 2026-05-19 ("build-site-planner independently generates a 9-page generic site skeleton that ignores the adopted pages"); addressed by the convergence work verified 2026-06-05 plus the "Existing Pages — ALREADY BUILT, PRESERVE EXACTLY" prompt block (v1.0.1047). Residual: prompt alone did not stop differently-slugged siblings (bare `economy-basics` beside `guide-economy-basics`) — that took Pass C2.
- **what:** Two confirmed mechanisms: (1) the planner planned from identity/archetype without reading realised state, inventing parallel pages (renamed tool dups, `post` placeholder from a prompt example); (2) ValidateRoles couldn't converge a childless plan (section-index promotion needs a child declaring ParentSection). Root cause per doc 029: two surfaces (adoption, planner) both write pages and queue work without a shared identity space. Fix: planner reads realised state and converges; reconciler is the sole work-item producer ("can't produce duplicates by construction").
- **sources:** FOCUS_planner_ignores_adopted_state.md, running_notes_14(25)#part-14c–14e, migration_cleanup_bare_guide_duplicates.sql
- **relations:** doc 029/030 declarative plan + reconciler; reconcilePlanWithRealised; nav dedup guard B-029-1
- **verify-later:** `plan_site` prompt existing-pages block in live build-site-planner def; llm_call_log for planner runs

<!-- SOURCE: U09_adoption.md -->
### reconcilePlanWithRealised convergence (Pass A union, rename snap-back, Pass C/C2 dedup)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "VERIFIED RESOLVED on a clean run (2026-06-05 17:26Z, corr 6381cb13)… guide-economy-basics…guide-skinner-box all as role=guide (5), with ZERO bare siblings… Pass A unioned the adopted guides into the plan and Pass C2 dropped the bare-sibling duplicates, both firing for the first time."
- **what:** Deterministic Go convergence in `ValidateSitePlanAction`/`v3_site_actions.go`, gated on `adoption_locked` pages: unions LLM-omitted adopted pages into the plan (via `normaliseRealisedToPlanPage`), snaps back renames, dedups section-stem collisions (`sectionStemOf`) and item-topic siblings (`itemStemOf` strips tool-/guide-/game- prefixes mirroring CanonicalisePage — Pass C2), and truncates preserving locked pages. It does not special-case adoption in Go — it preserves whatever the query flags.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md, running_notes_14(25)#part-14l–14n
- **relations:** first-plan branch; type-assertion inertness bug (kept it dead until 06-05); union-clobber carry fix
- **verify-later:** `v3_site_actions.go` reconcilePlanWithRealised, itemStemOf; planner log lines "existing pages loaded for convergence", "reconciled with adoption-locked pages"

<!-- SOURCE: U09_adoption.md -->
### Union-clobber bug and the carry fix (sections/meta_description/nav_order)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "on the first pass, every adopted page the LLM omitted was unioned with empty values and the sync clobbered its real sections/meta_description/nav_order to empty… Fix (both must land together): (a) load_existing_pages SELECT adds the fields… (b) normaliseRealisedToPlanPage carries them" — verified on the 2026-06-05 clean run; "the empty hubs were the union clobber… NOT a planner gap."
- **what:** Pass A's union originally emitted `sections: []` because the 054 query didn't select the fields, and `upsertPage`'s `ON CONFLICT … sections = EXCLUDED.sections` overwrote the adopted page's real values — the difference between a faithful first pass and one that wipes adopted content the LLM didn't re-list. The carry fix also reframed the "empty hubs" defect: source hubs are populated (`guides-index → ["guide-list"]` etc.); no separate hub-convergence step is needed for adopted sites.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#union-clobber, running_notes_14(25)#part-14i–14j, migration mentioned: migration_load_existing_pages_carry_fields.sql
- **relations:** upsertPage ON CONFLICT semantics; empty-hub clarification; convergence
- **verify-later:** load_existing_pages SELECT column list; normaliseRealisedToPlanPage in v3_site_actions.go

<!-- SOURCE: U09_adoption.md -->
### Canonical page-shape vocabulary (CanonicalisePage + ValidateRoles)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Phase 0 "landed" (FOCUS_planner_ignores_adopted_state); Part A `-index` rule "written, unit-tested green, and deployed" and verified via the 2026-05-28 run (CATALOGUE §0 "hubs deployed as section-index at nested URLs").
- **what:** One canonical name/URL/page_type vocabulary for logical pages (index `/index.html`; `<slug>.html` content; `<section>-index` → `/<section>/index.html`; `tool-<slug>` → `/tools/<slug>/index.html`; guide role → `/guides/<slug>/index.html`), implemented in `datahelpers.ValidateRoles` + `CanonicalisePage` (page_canonical.go). Part A adds Rule 2: a name ending `-index` with a non-leaf role is promoted to `section-index` (with an `isLeafRole` guard), recovering the LLM's reliable signal (the name) when url/parent are omitted. Part B (de-hard-code the tools/guides/games vertical vocabulary in `nestedRoleFromURL`) remains unscoped. The two role-normalisers (`normaliseRole` routing-collapsed vs `normalisePageType` flavour-preserving) are intentionally layered — merging them was withdrawn as wrong.
- **sources:** HANDOFF_2026-05-25, FOCUS_chrome_templates_and_page_shape.md#fix-2, running_notes_14(25)#part-1–5
- **relations:** sync canonicalisation divergence; adoption URL computation (flat, pre-canonicaliser); guide page_type
- **verify-later:** `page_role_validator.go` (Rule 2 + isLeafRole), `page_canonical.go` guide case, `nestedRoleFromURL` hardcoded verticals

<!-- SOURCE: U09_adoption.md -->
### Two-canonicalisation-surfaces divergence: SyncPagesToDB lacked ValidateRoles
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "Both the core fix (sync no longer flattens hubs) and the companion (built_from_plan_version set…) are confirmed. Thread closed." (running_notes_14 Part 5, 2026-05-28).
- **what:** `WriteSitePlanAction` ran ValidateRoles+CanonicalisePage → correct plan; `SyncPagesToDBAction` ran CanonicalisePage only, on the raw `page_plan` from collected data — so a `games-index` typed `content` flattened to `/games-index.html` and the upsert overwrote the correct adoption row. Fix chosen: Option 2 — sync runs the identical ValidateRoles pipeline (Option 1, reading site_plan_pages, would break the plan-less callers pageflow-builder/multipage-website-builder/site-work-orchestrator). Exposed the deliberate guides de-prefix trade-off (plan de-prefixes `guide-rng-design`; sync now agrees — surfaced, not silent).
- **sources:** running_notes_14(25)#part-1–3, HANDOFF_2026-05-25#confirmed-root-cause
- **relations:** canonical vocabulary; built_from_plan_version companion; ARCHITECTURAL_TENSIONS #2 (identity derived in multiple places)
- **verify-later:** `site_db_actions.go` SyncPagesToDBAction normalisation loop

<!-- SOURCE: U09_adoption.md -->
### built_from_plan_version drift stamp + removal of the deployed→needs_rebuild flip
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "Option B shipped (two files, coupled)… confirmed in production" (running_notes_14 Part 8–10; CATALOGUE A1 fix list, 2026-06-03).
- **what:** The intended doc-029 design — stamp `pages.built_from_plan_version` at build time and detect staleness in the reconciler — had been deferred; a stand-in `deployed → needs_rebuild` flip in `upsertPage` over-fired on every sync and churned pre-plan tool deploys. Completion: `UpdatePageStatusAction` stamps the current plan id on deploy; `upsertPage` COALESCE fill-if-null (never overwrite a real build version) and the flip removed; drift flows through the reconciler's `decideEmit`. Principle recorded: before fixing a misbehaving mechanism, check for deferred design debt — complete it rather than patch around it.
- **sources:** running_notes_14(25)#part-8, CATALOGUE(9)#family-a, old2/HANDOFF_2026-05-07(1)#5
- **relations:** doc 029 drift detection; A1 tool/game deploy failure; reconciler stale-page churn
- **verify-later:** `v3_site_actions.go` UpdatePageStatusAction deployed branch; `site_db_actions.go` upsertPage CASE

<!-- SOURCE: U09_adoption.md -->
### Bare-sibling duplicate pages (planner re-invents adopted topics)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "DECISIVE (llm_call_log plan_site @ 20:25:22): the planner WAS given the adopted guides and emitted economy-basics anyway → PROMPT-RULE gap… FIX (recommended, structural, Go): deterministic guard… drop a planned page whose topic STEM collides" — shipped as Pass C2 and verified on the 2026-06-05 clean run; cleanup migration applied.
- **what:** The planner proposed bare `economy-basics` etc. beside adopted `guide-economy-basics` — a differently-slugged sibling the "preserve existing pages" prompt rule did not stop. Deterministic Go stem-dedup (Pass C2, reusing CanonicalisePage's prefix stripping) is the guarantee; a prompt stopgap was optional. The durable cleanup migration also removes the bare rows from the current plan (reconciler would re-create them otherwise) and terminalises their work items (site_work_items.page_id has no FK).
- **sources:** running_notes_14(25)#part-14c–14e, migration_cleanup_bare_guide_duplicates.sql, FOCUS_adoption_faithfulness_via_locks(5).md#item-topic-sibling-dedup
- **relations:** planner ignores adopted state; convergence Pass C2; LLM-rule vs deterministic-guard principle
- **verify-later:** itemStemOf/Pass C2 in v3_site_actions.go

<!-- SOURCE: U09_adoption.md -->
### Adoption calls the canonicaliser + reconciler orphan pruning
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** "Adoption today doesn't go through this. It computes its own URL based on page_type only… This needs an additional reconciler pass: pages… with NO entry in site_plan_pages… should be soft-deleted or marked for removal. The reconciler… doesn't prune orphans. That's a follow-up." (FOCUS_chrome_templates Fix 2).
- **what:** Adoption's local URL computation (flat `/games.html` etc.) diverges from the canonicaliser the planner uses, producing duplicate logical pages (`games` + `games-index`) that ON CONFLICT can't match. Proposed: apply_adoption_plan calls CanonicalisePage; reconciler gains an orphan-pruning pass (pages absent from the current plan get archived); one-off cleanup migration. Partially overtaken by the convergence work (which unions/dedups at plan time) and the analyze_site prompt fix, but orphan pruning remains unbuilt — orphaned bare pages persisted after Pass C2 dropped them from the plan and needed manual cleanup.
- **sources:** FOCUS_chrome_templates_and_page_shape.md#fix-2, running_notes_14(25)#part-14l follow-up
- **relations:** canonical vocabulary; bare-sibling cleanup migration (the manual stand-in); page-cleanup pass idea in 05-07 Phase-2 candidates
- **verify-later:** apply_adoption_plan URL computation today; any reconciler pruning logic

<!-- SOURCE: U09_adoption.md -->
### Deferred plumbing stubs: scheduled reconciler tick, domain-aware ensure_pages
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** "6. Scheduled reconciler tick — Not built. Reconciler currently fires only when called by the planner… 7. ensure_pages should be domain-aware — Currently hardcoded in workflow JSON… Stub for the next discussion" (HANDOFF_2026-05-07(1)). A scheduled reconcile tick is later referenced as existing in emit_design guard rationale ("Plan-time, not reconcile-time, so the scheduled reconcile tick does not backfill") — status conflict to resolve in stage 2.
- **what:** Two small deferred items from Phase-1 deployment: a heartbeat scheduled_tasks row producing periodic reconcile passes (mirroring content-feed-trigger), and moving the hardcoded ensure_pages page list into strategist/briefing-written site_specs read at plan time.
- **sources:** old2/HANDOFF_2026-05-07(1)#6–7, FOCUS_design_composition_flow_and_adoption_fidelity(1).md#3A
- **relations:** reconcile_site_plan; build-pipeline-trigger cadence
- **verify-later:** scheduled_tasks for a reconcile tick; ensure_pages config source

<!-- SOURCE: U10_imagery.md -->
### needs_section_data resolution: reconciler, not an agent
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** "SUPERSEDED 2026-05-06 by FOCUS_directory_builder_and_list_components.md"; "Update 2026-05-27… Decision: a full LLM handler agent (and the directory-builder agent) is not needed for the query-resolvable cases" — the two decided pieces not marked built.
- **what:** `needs_section_data` items are emitted at needs_human_review meaning "couldn't resolve component or required field", not async dispatch; 41 items were stuck system-wide. Resolution machinery already exists (`queryresolve.Resolve`, only `pages_where_type:<type>` implemented; `pages_under_section` named but absent from the dispatch switch). The settled design: (1) implement pages_under_section in queryresolve; (2) a section-data reconciler (a resolver, not an agent) re-attempting open items through existing machinery, closing via closeResolvedDataRequest and flagging re-renders; genuinely-human data (spec-sourced) stays HITL. The originally-planned dedicated handler agent and the never-built `directory-builder` agent are documented dropped ideas.
- **sources:** FUTURE_section_data_handler_1_.md (header supersession + 2026-05-27 update + original)
- **relations:** abandoned: directory-builder agent; relates to list components inventory (~17 components) and page-build-handler.
- **verify-later:** queryresolve.go dispatch switch; count of stuck needs_section_data items.

<!-- SOURCE: U12_docs024_archives.md -->
### site_plan page-role enum naming (underscore → hyphen; index → landing)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Archive: `"section_index" | ... | "blog_post"`; live: `"section-index" | ... | "blog-post" | "landing"`.
- **what:** `site_plan_pages.role` vocabulary was originally underscore-separated with a bare `index` role for the homepage. Renamed to hyphenated form and the homepage role renamed to `landing`, matching kebab-case conventions elsewhere.
- **sources:** old/029_site_plan_and_reconciler.md#"role table"; docs024_key_docs_latest/029_site_plan_and_reconciler(2).md#"role table"
- **relations:** page_type vocabulary and kebab constraint (016 §6.5)
- **verify-later:** DB check constraint on `site_plan_pages.role`/`pages.page_type` for hyphenated values.

<!-- SOURCE: U12_docs024_archives.md -->
### site_plan_partials — single JSONB-blob partial storage (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "JSONB blobs were considered and rejected because at anticipated scale... loading whole blobs to read one slice is wasteful, surgical HITL edits become hard, and lock transfer at meaningful granularity is impossible."
- **what:** Archived Phase 1 plan proposed one table, `site_plan_partials`, storing each partial as a single versioned JSONB blob per plan. Abandoned for two normalized row-per-thing tables — `site_plan_sections` and `site_plan_directives` — enabling per-row HITL locking at 1000+ page scale.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"schema section"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"schema section"
- **relations:** lock transfer across plan rebuilds; lazy per-page brief generation (also abandoned)
- **verify-later:** confirm `site_plan_directives`/`site_plan_sections` tables exist, `site_plan_partials` does not.

<!-- SOURCE: U12_docs024_archives.md -->
### Three sequential per-partial plan-builder LLM calls (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "Earlier draft of this doc proposed three sequential LLM calls. Looking at the existing build-site-planner agent, that lean was wrong."
- **what:** Archived plan proposed splitting the plan-builder into three sequential LLM calls for independent retry granularity. Abandoned once it was noticed the production build-site-planner agent already produces all three coherently in one call with no evidence of retry-granularity problems.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q2. Plan-builder LLM tier"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Q2. Plan-builder LLM call shape"
- **relations:** site_plan_partials (abandoned)
- **verify-later:** build-site-planner agent_definitions workflow — confirm single LLM call shape.

<!-- SOURCE: U12_docs024_archives.md -->
### Separate BuildPageURL path-resolver helper (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Live: "The earlier draft of this doc proposed a separate BuildPageURL helper... That argument was overly cautious... Consolidated."
- **what:** Archived plan proposed a brand-new ~50-line Go helper sibling to `page_canonical.go`. Abandoned as overly cautious: Phase 1 instead extends `CanonicalisePage` additively with an optional `ParentSection` field.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q3. URL paths"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Q3. URL paths"
- **relations:** site_plan page-role enum naming
- **verify-later:** `datahelpers/page_canonical.go` — confirm `CanonicalisePage` has `ParentSection`, no separate `BuildPageURL`.

<!-- SOURCE: U12_docs024_archives.md -->
### Lazy per-page brief generation via build_page_brief step (abandoned)
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** Archive rollout step 8: "build_page_brief step in page-build-handler... generates site_plan_partials/page_brief:<name> if missing." Live replaces with a pure-Go brief renderer.
- **what:** Archived plan generated each page's brief lazily via an LLM step during page build. Abandoned for a deterministic, non-LLM Go helper that assembles a brief at read time by walking the directive cascade and applying cardinality rules.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"rollout table, step 7-8"; docs024_key_docs_latest/030_phase1_plan_and_reconciler(5).md#"Directive cascade and brief assembly"
- **relations:** site_plan_partials (abandoned)
- **verify-later:** confirm `datahelpers/page_brief.go` exists; page-build-handler has no `build_page_brief` LLM step.

<!-- SOURCE: U13_docs024_small_dirs.md -->
### FAQ empty-items bug: duplicate content-surface planning (Defect 1)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "Deployed status (2026-05-21) ... Prevention shipped on three fronts, all live" with confirmed-live flags and chassis v1.0.1029
- **what:** Pages were planned with both a freeform `generic-text-block` and a structured component (`faq`, `pricing`) intended to hold the same content, because the content-gap-planner's prompt example hardcoded `generic-text-block` and the site-planner's mappings omitted faq/pricing entirely; the content writer (proven correct by an isolated build test) then filled the freeform block and left the structured component empty. Fixed by editing both planner prompts and an archetype-aware `defaultSectionsForPage` Go backstop.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md, js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#empty-shells, js_snippets_news_gaswholesalers/old/faq_empty_items_prevention_findings(1).md
- **relations:** Display-name leak (Defect 2); "Renders empty" diagnostic method; per-section briefs gap; extractResponseContent flat-string hypothesis (superseded)
- **verify-later:** content-gap-planner and site-planner agent_definitions prompt_template, apply_gap_plan_action.go defaultSectionsForPage, chassis v1.0.1029

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Display-name leak into section arrays (Defect 2) + validate_components resolver
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "validate_components implemented in ValidateSitePlanAction (was a dead flag)... deployed in chassis v1.0.1029"
- **what:** A planner path could emit a component's `display_name` instead of its kebab `function` into a page's `sections` array, orphaning the page_component. Fixed by implementing the previously-dead `validate_components` config flag in `ValidateSitePlanAction`: a `componentNameResolver` resolves each section name (exact match → NormalizeComponentFunction → display/name lookup → drop+log if unresolvable). The gap-planner path (`applyNewPage`) doesn't route through `validate_site_plan`, so the same resolver had to be wired in separately there too.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#Fix-B-implementation, js_snippets_news_gaswholesalers/old/validate_components_implementation.md
- **relations:** FAQ duplicate content-surface bug; per-section briefs gap
- **verify-later:** ValidateSitePlanAction, loadComponentNameResolver, NormalizeComponentFunction

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Per-section briefs gap (planner depth) — bare section-name strings, no intent
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "Planner depth: per-section briefs + stale-plan write-back... planner needs to emit them" listed under "Open — structural (not blocking)"
- **what:** `site_plan.pages[].sections` is an array of bare strings with no per-section brief. This is the deeper cause behind Defect 1: without a brief, the writer cannot tell that `faq` and `generic-text-block` are competing surfaces. A consumer already exists (`plan_sections.sectionDescription`) but the planner never emits any of those shapes. Token-budget caveat: adding briefs to every section on a large site materially grows planner output size.
- **sources:** js_snippets_news_gaswholesalers/FOCUS_faq_empty_items_and_page_content.md#Fix-C-stale-plan, js_snippets_news_gaswholesalers/old/site_planner_depth_and_freshness_concerns.md
- **relations:** FAQ duplicate content-surface bug; Post-build validation of structured components; validate_components resolver
- **verify-later:** load_page_sections_from_spec, plan_sections.sectionDescription resolver

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Stale site_plan — gap-planned pages never written back (Concern 2)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** TODO_remaining_work.md "gap-planned pages aren't written back to site_plan (faq was absent from the plan entirely)... apply_gap_plan should append new pages to site_plan" — not yet implemented
- **what:** Pages created after initial site planning get a `pages` row and nav entries but are never appended to `site_specs.site_plan`; the plan drifts from reality with every gap-added page. Proposed fix: `apply_gap_plan` deep-merges the new page into `site_specs.site_plan` (mirroring `enrich_news_feed`'s pattern), plus a periodic plan-reconciliation discovery check.
- **sources:** js_snippets_news_gaswholesalers/016_debugging_guide_addenda.md#stale-plan, js_snippets_news_gaswholesalers/old/site_planner_depth_and_freshness_concerns.md
- **relations:** Per-section briefs gap; page content-creation build pipeline trace
- **verify-later:** apply_gap_plan_action.go, enrich_news_feed deep-merge pattern

<!-- SOURCE: U13_docs024_small_dirs.md -->
### site_plan as authoritative build source, overwriting pages.sections
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** "`load_page_sections_from_spec_action.go` ... CONFIRMED in code" (PLAN_tool_widget_clobber(9).md §2.4)
- **what:** The page-build pipeline's section authority is `site_specs.site_plan`, not `pages.sections` directly — the loader syncs the plan's sections back into `pages.sections` on every build where a plan entry exists, only falling back to `pages.sections` if the plan yields nothing. Consequence: any fix that only sets `pages.sections` inside a tool action is futile once a plan entry exists; a durable fix must add the tool/embed section to the planner's `site_plan` output itself.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.4, tools/tool_widget_clobber/NOTES_running_tool_widget_investigation.md#Phase-2
- **relations:** Tool widget clobber mechanism (M1); Canonical tool-page section-shape design question
- **verify-later:** `load_page_sections_from_spec_action.go`; whether `site_plan` now carries a tool/embed section entry for `page_type='tool'` pages

<!-- SOURCE: U13_docs024_small_dirs.md -->
### queryresolve reality-vs-invention architectural promise
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Stated as an existing architectural line: "queryresolve exists specifically to draw a line between content the LLM is allowed to write... and content that has a database answer"
- **what:** A specific agent responsibility (`queryresolve`) enforcing a hard boundary in the site-build pipeline between LLM-authored creative content and database-derived factual lists — framed as central to the platform's "avoid fabrication" mission alongside carving the build into specialists with non-overlapping responsibilities.
- **sources:** pitch/003thebiggerpicture.md
- **relations:** Fractal agent architecture claim; Design/composition work-item emission gap
- **verify-later:** queryresolve action implementation; `source: query.*` convention in page_components

<!-- SOURCE: U13_docs024_small_dirs.md -->
### New-domain build pipeline stage chain (domain-submitter → page-build-handler)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** Traced live from code/DB snapshot: "Confirmed: ReconcileSitePlanAction reconciles pages only" and "The chain is fully connected" — caveated as read from a 2026-05-21 DB backup snapshot, "may have drifted"
- **what:** The confirmed happy-path chain for building a brand-new domain: `domain-submitter` → `domain-research-classifier` → `domain-strategist` → `build-briefing-agent` → `build-site-planner` (plan_site → validate → write_site_plan → sync_pages → populate_nav → reconcile_site_plan) → `page-build-handler` per page → `rerender-pages`. Driven by the 30s `build-pipeline-trigger` heartbeat, with every stage's `create_work_item` defaulting to status `triaged` so the pipeline self-advances.
- **sources:** plainjanedomain/README.md
- **relations:** Design/composition work-item emission gap; queryresolve reality-vs-invention architectural promise
- **verify-later:** live SELECT type, status, image_tag FROM agent_definitions WHERE type IN (domain-submitter, domain-research-classifier, domain-strategist, build-briefing-agent, build-site-planner)

<!-- SOURCE: U13_docs024_small_dirs.md -->
### Design/composition work-item emission gap (planner reorg unclosed seam)
- **category:** site-plan-and-reconciler
- **status-signal:** unknown
- **status-evidence:** "So nothing in the build path appears to emit a needs_design/needs_composition trigger for a fresh domain... consistent with this being an unclosed seam from the planner reorg"
- **what:** A discovered structural risk: the legacy `WriteBuildItemsAction` emitted the full item set for a new build (`needs_page`, `needs_logo`/`needs_hero_image`, `needs_composition`, `needs_design`), but the Phase-1 replacement (`build-site-planner` → `write_site_plan` + `reconcile_site_plan`) emits only `needs_page` + `needs_rerender`. The only fallback is the improvement-loop's `design-discovery-agent` catching `missing_css` later — meaning a new site could deploy pages referencing a stylesheet that doesn't exist yet.
- **sources:** plainjanedomain/README.md
- **relations:** New-domain build pipeline stage chain; Site-chrome rendering gap (dartsonline) — same class of defect
- **verify-later:** ReconcileSitePlanAction, WriteSitePlanAction, WriteBuildItemsAction Go source; design-discovery-agent missing_css check

<!-- SOURCE: U14_docs019_runbooks.md -->
### Roadmap-phases scope decision gap (nav grounded in built reality)
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 6 "PROMOTED 2026-07-07 — THE BUG IS PLATFORM-WIDE … 082_submit_domain_unified.sh accepts ONLY --mission … AND build-site-planner's prompt has NO ELSE BRANCH for the roadmap-authority block … an absent decision point, not a missing default."
- **what:** No submitted site ever gets a roadmap/phases decision: the submit script has no --roadmap path and the planner's phase-discipline instructions vanish (not degrade) without one — so commerce-shaped domains get aspirational full plans and nav links to unbuildable pages. Fix shape (relay-wide, by construction): a post-classification scope-decision hop writes a phased roadmap_brief (P1 content/guides/tools; P2 legal-gated affiliate; P3 catalogue); planner prompt gains the ELSE branch (default phase-1-only or HITL hold); nav generation grounded in the BUILT set regardless of plan. Guidelines 001 already define the roadmap/phases mechanism — the docs had it, intake didn't. The legal gate on P2 is named as the fix-loop council's first concrete reviewer job.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 6); docs019/RUNBOOK_diagnosis_fix_loop(9).md#root-context
- **relations:** F0 guides pilot (nav-vs-built strand); coverage baseline; council compliance reviewer
- **verify-later:** 082_submit_domain_unified.sh flags; build-site-planner roadmap_brief template block; nav-updater

<!-- SOURCE: U15_docs019_running_notes.md -->
### Roadmap-phase enforcement gap (builder item 6)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** "VERIFIED IN CODE: 082_submit_domain_unified.sh — grep confirms ONLY --mission/--mission-file exist, no --roadmap anywhere. build-site-planner prompt — the {{if .roadmap_brief}}...{{end}} block has NO else" (NOTES_running_fixloop(9).md).
- **what:** A platform-wide defect (reclassified from a single-site fix into the builder thread's main queue item) where no domain-submission path ever produces a phased roadmap, so a site's Tier-3 roadmap phase rules simply vanish rather than degrade — an absent decision point, not a hidden mechanism. Fix shape: a new post-classification hop writing a phased roadmap for commerce-shaped domains, enforced at three existing relay-wide points (strategist prompt, planner deliverability validation, built-grounded nav) rather than per-site.
- **sources:** NOTES_running_fixloop(9).md "TWO CORRECTIONS: amendment path under-specified; bug is platform-wide"; NOTES_running_synthesis_v4(39).md 2026-07-07 mirror entry.
- **relations:** Diagnosis→fix loop workstream founding; work-item relay / builder-generations architecture; curated best-in-class standing expectation.

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Site plan as declarative artefact + reconciler
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 029(1) "The shape that fixes this is the same pattern Kubernetes uses: a declarative artefact … plus a reconciler … Phase 0 lands today"
- **what:** Fixes the duplicate-pages bug where two surfaces (adoption + site-planner) both wrote `pages` rows without a shared identity space. The planner writes a declarative desired-state plan; a deterministic Go reconciler (`reconcile_site_plan`, no LLM) walks desired-vs-realised and emits `needs_page:<name>` for the diff only.
- **sources:** WM/029_site_plan_and_reconciler(1).md#why-this-exists, WM/029_site_plan_and_reconciler(1).md#phase-1-plan-as-declarative-artefact-reconciler-emits-work, WM/030_phase1_plan_and_reconciler(4).md#plan-builder-cascade-replaces-todays-site-planner-emit-and-queue
- **relations:** CanonicalisePage; plan-domain schema; LLM tiering; drift auditors
- **verify-later:** reconcile_site_plan action; site_plan_structure/pages; pages.built_from_plan_version

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### CanonicalisePage + role validator (deterministic page identity)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 029(1) Phase 0 "A single canonicalisation helper in datahelpers/ … called from both surfaces"; 030(4) Q3 role validator (Go)
- **what:** A single `datahelpers/page_canonical.go` helper maps a `(role, slug, parent_section)` descriptor to a canonical `(name, url, page_type)` triple, called from both adoption and planner surfaces. Phase 1 extends it with `ParentSection` and adds a role-validator that corrects LLM role mislabels deterministically before persisting.
- **sources:** WM/029_site_plan_and_reconciler(1).md#fix, WM/030_phase1_plan_and_reconciler(4).md#q3-url-paths-canonicalisepage-phase-0-helper-extended-linknav-agents-own-drift, WM/016_debugging_guide_v2_44.md#adoption-faithfulness
- **relations:** site plan reconciler; architectural tension #1/#2; adoption faithfulness strip bug
- **verify-later:** datahelpers/page_canonical.go; ValidateRoles; CanonicalisePage

<!-- SOURCE: U17a_docs019_archive_discussions_and_main.md -->
### Plan-domain schema + directive cascade + brief assembly
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** 030(4) Q1 "separate site_plans schema, not site_specs aspects … four plan-domain tables, all row-shaped for scale"
- **what:** Phase 1 rejects reusing `site_specs` aspects in favour of normalised plan tables (`site_plans`, `site_plan_pages`, `site_plan_sections`, `site_plan_directives`) row-shaped for 1000+ page scale. Guidance lives in `site_plan_directives` at site/page/section scope; a Go brief renderer (`datahelpers/page_brief.go`) walks the cascade and applies single- vs multi-valued cardinality.
- **sources:** WM/030_phase1_plan_and_reconciler(4).md#q1-plan-storage-separate-site_plans-schema-not-site_specs-aspects, WM/030_phase1_plan_and_reconciler(4).md#directive-cascade-and-brief-assembly, WM/030_phase1_plan_and_reconciler(4).md#what-stays-in-site_specs
- **relations:** site plan reconciler; lock transfer; strategic-vs-plan-time naming split
- **verify-later:** site_plan_directives; datahelpers/page_brief.go; write_site_plan action

<!-- SOURCE: U18_sql_for_agents.md -->
### site-planner (single-LLM-call site plan)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** 022 shows model flip-flops (sonnet→haiku for cost, 040 haiku→sonnet because planning is "high-leverage"); 053 build-site-planner is the successor for work-item builds.
- **what:** v2 planner: one LLM call over brief + component library + style collections producing validated_plan, pages, style_collection, needs_logo/needs_images. The model-choice oscillation (cost vs quality on high-leverage decisions) is documented reasoning worth keeping.
- **sources:** 022_site_planner.sql; sql_for_agents_v2/022_site_planner.sql; 040_optimise_which_llms.sql
- **relations:** chief-strategist (predecessor), build-site-planner (successor), pageflow-builder (caller)
- **verify-later:** which planner the live pipelines invoke

<!-- SOURCE: U18_sql_for_agents.md -->
### build-site-planner + roadmap-overrides-components rule
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 053 shows the workflow rewired to the site_plans domain ("changed to: ... write_site_plan → sync_pages → populate_nav → reconcile_site_plan → complete"); plan_site runs on claude-opus-4-6; 067 adds thinking budget.
- **what:** Handler for needs_site_plan. Reads site_specs (identity/classification/briefing/strategy), loads component library and style collections, plans via LLM, validates, then writes into the site_plans domain and reconciles. Carries the ROADMAP OVERRIDE rule verbatim: "ROADMAP OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase... use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list... Unknown section types are handled by the component selector downstream. Do NOT replace roadmap section_types with standard components... The roadmap is the authority for this site." Earlier form wrote plan/design_intent/content_direction specs + write_build_items (one needs_content_write per page).
- **sources:** 053_build_site_planner.sql; 108_site_plan_pages.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** site_plans/reconciler domain (docs 029/030); component selector creating needs_new_component items; roadmap spec aspect
- **verify-later:** write_site_plan + reconcile_site_plan actions; roadmap aspect producer

<!-- SOURCE: U18_sql_for_agents.md -->
### site_plan_pages schema repair (plan-domain drift)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 108 "Migration 033: Reconcile site_plan_pages columns + drop orphan site_plan_partials... every write_site_plan call to date has failed at the title-column error."
- **what:** Repairs drift between two drafts of the site-plan schema: adds title/meta_description/nav_label columns, drops page_data and the unused site_plan_partials table (directives are row-per-directive in site_plan_directives). Documents the CREATE TABLE IF NOT EXISTS silent-skip failure mode when a rewritten migration follows an applied earlier draft.
- **sources:** 108_site_plan_pages.sql
- **relations:** build-site-planner; migration-discipline concepts (124)
- **verify-later:** live \d site_plan_pages / site_plan_directives

<!-- SOURCE: U19_sql_tables_components.md -->
### site_plans declarative plan domain
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Migration 031 (both drafts) with detailed rationale referencing doc 030; later tables (site_plan_imagery, work-item flows) depend on and reference it.
- **what:** The plan is a separate versioned artefact from site_specs: site_plans (version anchor, one is_current per site), site_plan_pages (row per planned page: canonical name/role/slug/url, parent_section for section-index detection, nav flags), site_plan_sections (structural per-section rows carrying resolved component_version/palette/layout/typography ids for HTML data-* provenance), site_plan_directives. Row-per-thing chosen over JSONB blobs for 1000+ page scale and surgical HITL edits; versioning mirrors site_specs (is_current + superseded_at).
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql
- **relations:** site_specs (strategic vs operational boundary); reconciler; naming note that plan_sections/save_page_sections actions "share a noun and nothing else".
- **verify-later:** write_site_plan action; plan row counts per site.

<!-- SOURCE: U19_sql_tables_components.md -->
### Directive cascade and HITL lock transfer
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** 040 second draft: scope_ref encoding, cardinality lookup in brief renderer, "write_site_plan... transfers the lock onto the equivalent new directive row" matched by (scope, scope_ref, category, subject, ordering).
- **what:** Design/content/voice/structural guidance stored row-per-directive at site/page/section scope; a Go brief renderer walks the cascade (site → page → section) and emits prompt-ready text — consumers never read directives directly. Cardinality (override vs accumulate) is renderer knowledge, not schema. Human-locked directives survive plan rebuilds via stable-composite-key lock transfer performed only by write_site_plan.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#site_plan_directives
- **relations:** Pattern A locks; site_plan_imagery (same pattern); doc 030 "Directive cascade and brief assembly".
- **verify-later:** brief renderer helper; lock-transfer code in write_site_plan.

<!-- SOURCE: U19_sql_tables_components.md -->
### Plan drift detection and reconciler scheduling
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** pages.built_from_plan_version + sites.last_reconciled_at columns with reconciler semantics documented; later migrations reset built_from_plan_version=NULL to force rebuilds.
- **what:** Each built page records the plan version that produced it; the reconciler diffs site_plan_pages against pages, flags pages whose plan version lags current (NULL = never built under a plan), and emits needs_page/rebuild work items. sites.last_reconciled_at lets the scheduled tick skip recently reconciled sites; deliberately no FK so hard-deleted plans read as drift.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#4 and #5; docs/agent_docs/sql_for_tables/003_pages.sql#rebuild-flips
- **relations:** site_plans domain; site_work_items; scheduler.
- **verify-later:** reconcile_site_plan action; scheduled reconciler task.

<!-- SOURCE: U19_sql_tables_components.md -->
### site_plan_partials with lazy page briefs (early plan shape)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** First draft of migration 031 defines site_plan_partials ('design_direction', 'content_strategy' eager; 'page_brief:<name>' lazy via build_page_brief); the second draft in the same file replaces it with site_plan_sections + site_plan_directives.
- **what:** The initial plan-domain design stored design direction, content strategy and per-page briefs as versioned JSONB partials, with lazy page briefs written on demand by page-build-handler. Superseded by the row-per-section/row-per-directive shape for scale and surgical edits.
- **sources:** docs/agent_docs/sql_for_tables/040_site_plans_schema.sql#site_plan_partials
- **relations:** superseded by site_plan_sections + site_plan_directives.
- **verify-later:** whether site_plan_partials exists in production or only the directive shape shipped.

<!-- SOURCE: U20_legacy_docs_a.md -->
### Multi-page site support (wrap_multipage, multipage-site-builder)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** 030 SQL creates multipage-site-builder (index/about/contact + privacy); 031 shows the wrap_multipage step after html_assembler with CollectedData trace; today's pages/site_plans domain is the successor.
- **what:** Extending the single-page pipeline to small multi-page sites: after assembly, a wrap_multipage action derives index/about/contact (and privacy) pages, and the deployer commits all files. The first step from "landing page generator" toward the current multi-page site model.
- **sources:** docs004_website_capture_project/007different_types_of_site/030_about_page_and_privacy.sql; docs004_website_capture_project/007different_types_of_site/031_about_page_multipage_site.md
- **relations:** successor: site_plans / pages domain (site-plan-and-reconciler docs 029/030); robot-hands 3-page build (earlier sibling).
- **verify-later:** wrap_multipage in registry.

<!-- SOURCE: U23_docs_root_vonc.md -->
### Three section sources for a page build (aspect → pages.sections → plan tables)
- **category:** site-plan-and-reconciler
- **status-signal:** deployed
- **status-evidence:** Workflow dump + code read 2026-07-06: "load_spec_sections... reads site_specs aspect site_plan (AUTHORITATIVE) → fallback page_record.sections. The site_plan_sections TABLE is NOT read by this path."
- **what:** Page builds resolve their section list from, in order: the `site_specs` aspect `site_plan` (legacy blob, 5 sites carry one; vonc has none), `pages.sections` (jsonb fallback — what actually serves vonc; the newer planner dual-writes plan tables → pages.sections), and same-role sibling synthesis; the `site_plan_sections` table is written by the vonc-generation planner but not read by the build path. Three peer stores with unclear precedence caused ten silent no-op builds and two fixes landing in the wrong store (a plan-table row, then the pages.sections UPDATE that finally unblocked).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3; docs/RUNBOOK_phase2_provocation_js(29).md#update-2026-07-06
- **relations:** plan storage authority (029 Q1); complete_error silent no-ops; load_page_record lookup semantics
- **verify-later:** load_page_sections_from_spec_action.go source order; SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE 'site_plan%'

<!-- SOURCE: U23_docs_root_vonc.md -->
### Plan storage authority — 029 Q1 and the withdrawn table-first alteration
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** PLAN_dynamic_sections(4): "SUPERSEDED (2026-07-06, same day) — decision deferred to 029 Q1; alteration WITHDRAWN"; "Decision closed (2026-07-07): the user chose REVERT."
- **what:** After the silent no-ops, a decision was made (then withdrawn the same day) to make the `site_plans` family the authoritative plan store and alter `load_page_sections_from_spec` to read site_plan_sections first. Reading design doc 029 showed plan storage is its OPEN Q1 ("site_specs aspects vs new table", lean = partitioned site_plan_* aspects + a reconcile_site_plan action); three shapes coexist in production (legacy site_plan blob aspect ×5 sites; 029 partitioned aspects apparently unimplemented; the vonc-generation tables with pages.sections dual-write). The alteration was withdrawn and the repo file reverted (ORIGINAL.go; cluster reverts on next chassis push); evidence contributed to Q1: the table path now exists in production post-dating the lean. Store-agnostic preventions retained. Earlier draft (v2 of the plan) also named a `site_plan_directives` child table not mentioned in the final version.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#decision + #superseded; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-alteration-withdrawn + #2026-07-07-revert-decision; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3
- **relations:** three section sources; reconcile_site_plan (029); planner ≥1-section invariant
- **verify-later:** git history of load_page_sections_from_spec_action.go (reverted?); repo grep reconcile_site_plan; docs024 029 doc Q1 status

<!-- SOURCE: U23_docs_root_vonc.md -->
### Planner role-aware ≥1-section invariant + role→pipeline mapping
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** Backlog item 1 in HANDOFF §9; "Invariant refined: every planned page whose ROLE is built by page-build-handler must have ≥1 section" (Gate B, 2026-07-06) — nowhere claimed built.
- **what:** The June planner emitted all 8 vonc pages but skipped SECTIONS for exactly the two non-standard roles — blog-post (legitimate: the blog pipeline builds those) and section-index (the defect that caused the archive 404). Prevention: at plan-store time, every planned page whose role page-build-handler owns must have ≥1 section, with the role→pipeline mapping made explicit; plus auditor drift rule (pages.sections vs current plan) and post-deploy URL-presence checks per active page.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#gate-results; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-i-gate-results
- **relations:** complete_error family; section descriptor design; quality-auditor rules
- **verify-later:** site-planner agent_definition; site_plan_pages roles for recent sites

<!-- SOURCE: U23_docs_root_vonc.md -->
### Autonomous section composition — per-section descriptor {role, kind, data_feed}
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** PLAN_dynamic_sections_and_loaders(4) status "DESIGN"; gaps list "(1) Section descriptor... Without this the framework can't tell static from dynamic" — none of gaps 1–5 marked built.
- **what:** The framework (not a human) should decide, from the domain/site-spec, which sections a page has, each section's role (to prevent overlaps like provocation-card's mini-lobby vs lobby-grid), whether it is static (build-time content) or dynamic (runtime-filled from a feed), and which named feed — encoded as a per-section descriptor `{component_name, role, kind, data_feed}` on the plan, written by the site-planner, consumed by build AND maintenance flows. The plan not carrying `kind` is why the assembler dropped the runtime-filled shells. Includes a spec-level feed catalogue and quality-auditor maintenance detections (dropped-dynamic, overlap, deferral, empty-dynamic). The root design point: a data-driven component should DECLARE its runtime data dependency so the pipeline provisions feed + loader automatically.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#the-question + #structural-gaps; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-reframed; docs/RUNBOOK_phase2_provocation_js(29).md#how-a-component-should-declare
- **relations:** Tier E runtime-feed tier; loader-builder agent; static-vs-dynamic distinction; plan storage authority (where the descriptor lives follows 029 Q1)
- **verify-later:** site_plan_sections columns (kind/data_feed/role exist?); site-planner prompt/workflow

<!-- SOURCE: U24a_docs_archive_classic_and_docs024_misc.md -->
### validate_components section-name resolver
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** validate_components_implementation: "Implements the dead `validate_components: true` flag … currently set true for site-planner but never read"; provides `loadComponentNameResolver` and a gated block for `ValidateSitePlanAction`; describes deploying and testing via the isolated-build harness (implies not yet live).
- **what:** A deterministic resolver that maps each site-plan section name to a real `content_components.function` — via normalisation, display-name lookup ("FAQ Section"→`faq`), and name lookup — dropping+logging unresolvable names so they don't orphan downstream `page_components`. Deliberately narrow: it does NOT deduplicate or make intent decisions (that's the planner prompt + per-section briefs). Must also run in `applyNewPage` (content-gap-planner path bypasses validate_site_plan).
- **sources:** js_snippets_news_gaswholesalers/old/validate_components_implementation(1).md#scope, #2-the-validation-block, #3-the-gap-planner-path
- **relations:** NormalizeComponentFunction; per-section briefs; content-gap-planner; component schema drift
- **verify-later:** ValidateSitePlanAction validate_components flag read; apply_gap_plan_action.go applyNewPage

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### SyncPagesToDBAction / WriteSitePlanAction canonicalisation divergence — Option 1 rejected
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 2: "**Option 1 (single source of truth):** sync reads identity from `site_plan_pages`... **Decision: Option 2**... Corrected an earlier framing that called Option 1 'the structural one' — Option 2 is the structural fix here; Option 1 is coupling."
- **what:** Two canonicalisation surfaces disagreed — `WriteSitePlanAction` ran `ValidateRoles + CanonicalisePage` (producing correct `section-index` hubs in `site_plan_pages`), while `SyncPagesToDBAction` ran `CanonicalisePage` alone on raw `page_plan` (producing flat `pages` rows), and `upsertPage`'s `ON CONFLICT` then overwrote the correct row with the flat one. Option 1 (make sync read the already-validated `site_plan_pages`) was rejected because `pageflow-builder` (confirmed active) and two other callers invoke sync with no plan ever written, so Option 1 would silently break them. The shipped fix (Option 2) runs `ValidateRoles` inside sync too, unifying the pipeline across all five callers.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 1–3
- **relations:** `pageflow-builder` deprecation (decoupled from this fix, tracked separately), guide page_type restructuring
- **verify-later:** `SyncPagesToDBAction`/`site_db_actions.go` current state; whether `pageflow-builder` was ever actually deprecated.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Adoption-faithfulness-via-locks convergence — confirmed INERT
- **category:** site-plan-and-reconciler
- **status-signal:** abandoned
- **status-evidence:** running_notes_14(20) Part 14h: "TRUE root cause: `reconcilePlanWithRealised` gates on `rm[\"adoption_locked\"]`; the live `load_existing_pages` query does NOT emit `adoption_locked` ... lockedPages always empty -> reconcile ALWAYS no-ops." And: "`FOCUS_adoption_faithfulness_via_locks.md` status — convergence 'Inert until 054 + write_site_plan land.' ... LIVE STATE: lock tables have ONLY `locked_at`/`locked_by` — NO `lock_type`/`lock_expires_at` -> 053 NOT applied... 054 NOT applied."
- **what:** A designed subsystem meant to make adoption re-plans faithful to already-realised (locked) pages — schema migration 053 (lock_type/lock_expires_at columns), migration 054 (`load_existing_pages` emits `adoption_locked`), and `write_site_plan` locking logic — was found, on live inspection, to be entirely unapplied. The one piece that *was* built (`reconcilePlanWithRealised`'s convergence check in `v3_site_actions.go`) silently no-ops because its input is never populated. This directly explains two other defects in the same arc (the bare-guide duplicates, and 5 guide pages never being unioned into the plan).
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 14h; (references live `FOCUS_adoption_faithfulness_via_locks.md`, `031_locks(3).md`)
- **relations:** bare-guide duplicate pages; sync/write-site-plan divergence
- **verify-later:** whether migrations 053/054 have since been applied; current state of `write_site_plan_action.go`'s `transferDirectiveLocks`.

<!-- SOURCE: U24d_docs_archive_adoption_content_quality.md -->
### Deployed→needs_rebuild ON CONFLICT flip — pre-design stand-in later completed properly (Option B)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** running_notes_14(20) Part 8: "the flip is a pre-design *stand-in* for 're-sync invalidates deployed pages'... It over-fires (every deployed page, every sync) and mis-fires on pre-plan deploys (tools)... **Option B shipped**: COALESCE fill-if-null; removed the `deployed→needs_rebuild` CASE branch... Drift now flows through the reconciler's `decideEmit`."
- **what:** `upsertPage`'s `ON CONFLICT` branch that flipped any `deployed` page back to `needs_rebuild` on every sync was a workaround for a never-shipped design: `029`/`030` intended `built_from_plan_version` to be stamped at build time and drift detected by the reconciler, but the stamp was "explicitly deferred" per `HANDOFF_2026-05-07` #5 ("User explicitly OK'd this"). The investigation confirmed the flip should be completed as originally designed rather than patched around (rejecting a narrower "Option A: exclude tool/game from the flip" as entrenching the workaround) — shipped as the deploy-time stamp in `UpdatePageStatusAction` + COALESCE fill-if-null in `upsertPage`.
- **sources:** adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 8; CATALOGUE_gamesdesign_post_sync_fix_defects(4).md A1
- **relations:** A1 tool/game deploy-gap root cause (below)
- **verify-later:** `v3_site_actions.go` `UpdatePageStatusAction`, `site_db_actions.go` `upsertPage` current state.

<!-- SOURCE: U25_leopardess_social.md -->
### Page section source precedence and the plan-storage triple shape (029 Q1)
- **category:** site-plan-and-reconciler
- **status-signal:** partial
- **status-evidence:** HANDOFF §3: "Section sources … site_specs aspect site_plan (authoritative in code) → pages.sections (fallback) → sibling synthesis. The site_plan_sections table is NOT read by this path"; "Three shapes coexist in production … The decision belongs to the planner/reconciler thread."
- **what:** A page build reads sections from the site_specs 'site_plan' blob aspect first, then pages.sections, then same-role sibling layout synthesis — while the newer site_plans/site_plan_sections tables (which the vonc-generation planner writes, dual-writing pages.sections) are ignored by this path. A drafted table-first alteration was consciously withdrawn pending design doc 029's open Q1 (aspects vs table). Operational corollaries: the provocations-index unblock was a pages.sections UPDATE; reconcile_site_plan re-emits needs_page for any planned-but-unbuilt page every run (the standing needs_page:provocation trap — park it to detected after every vonc reconcile).
- **sources:** docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#3, #9.7; docs/social001_vonc_tiktok_social/minilobby_task/RUNBOOK_minilobby_task.md#0 (needs_page:provocation)
- **relations:** silent no-op success class; archetype hub build (used reconcile_site_plan properly); docs024 029/030
- **verify-later:** load_page_sections_from_spec_action.go; SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE 'site_plan%'; reconcile_site_plan grep

<!-- SOURCE: U26_misc_dirs.md -->
### Website-builder agent group (six-specialist pipeline)
- **category:** site-plan-and-reconciler
- **status-signal:** superseded
- **status-evidence:** Ran in production per basic_usage/001 ("Step-by-Step Guide to Your First Website Build", migrations 005/007/009 referenced); the current platform builds sites via the site_plans domain / webdesign-agent pipeline (002 spine, docs 029/030), which replaced this group.
- **what:** The original end-to-end website creation flow: an orchestrator agent calls domain-analyst (business categorisation via web-search) → site-architect (page structure, pausing for human approval) → fan-out of content-researcher + visual-designer (image search/generation, logo) → html-developer (per-page vanilla HTML/CSS fan-out) → site-publisher (s3_upload, preview URL). Seeded as agent_definitions + an agent_groups row; triggered by one spawn_group Kafka message.
- **sources:** docs/architecture/027-create-website-creation-system; docs/basic_usage/001basic_usage.txt; docs/basic_usage/003_dynamic_prompt_improvement#step-1.1
- **relations:** superseded by site_plans + webdesign-agent + design-composition pipeline; HITL pause in site-architect; result storage split
- **verify-later:** migrations 005/007/009 in platform/database/migrations/; whether group still seeded

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Rebuild vs rerender semantics and stale-render fossilisation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** CHECK 4 RESULTS: "the deployed pages are RERENDER output carrying stale stored renders … deployed hero consumes legacy `var(--accent-color, #0f3460)` … A full page-build-handler rebuild is required; `needs_rerender` would re-fossilise them" — settling the 016-vs-026 documented tension "by direct evidence".
- **what:** Two distinct page-refresh routes with different semantics: `needs_rerender` (handler rerender-pages) reassembles stored `page_components.rendered_html` and injects stored chrome — it does NOT re-render component templates, so template changes never land and old renders fossilise; a full build (`site_work_items` insert: pipeline=build, handler_agent=page-build-handler, status=triaged) re-runs plan_sections and re-renders everything. idea.uk had lived for weeks on reassemblies of early renders while the library advanced — the fossil tell is a legacy variable name in deployed HTML (`var(--accent-color`), and its count going to 0 is the proof a rebuild truly re-rendered. Docs 026 ("rerender reflects new template") and 016 ("patches in place") disagreed; evidence sided with 016. Related hazard from the migration sketch: a content rebuild can de-tool a tool page (page-content-writer regenerates from plan_sections, which does not know the interactive tool).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS #Migration-backfill; running_notes_scheme_to_components(55).md#So #Sh(migration route); HANDOFF_scheme_to_components_for_claude_code(1).md#Invariant (item 5)
- **relations:** dual chrome render paths; work-item crafting conventions; deployed-binary-predates-disk class.
- **verify-later:** rerender-pages vs page-build-handler workflow definitions in agent_definitions; 016/026 doc reconciliation.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### rerender-pages v6 workflow (refresh_site_components gate)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Tb): "Workflow (v6) fully read: gate `input_data.spec.refresh_site_components == true` → FORCED chrome render (header/footer/head) → js snippets render+commit → blog listing → get_pages (deployed+active) → create_rerender_items (per-page) → update_site_status deployed → complete."
- **what:** The site-wide rerender agent: one work item fans out to a forced chrome re-render (the only workflow passing `force_rerender: true` — pageflow-builder and site-work-orchestrator pass false, which explains fossilised chrome surviving full builds), JS snippet rendering, blog-listing rebuild, then per-page rerender items that the build dispatch loop drains; it ends by marking the site deployed. `spec.function`/`component_id` are consumed nowhere in v6. The real producer of such items is `store_generated_component` on regeneration (one deduped `needs_rerender` per affected site, item_key `component_regen_rerender:<uuid>`).
- **sources:** running_notes_scheme_to_components(55).md#Ta #Tb; w4b_02_read_triggers.sql; w4b_04_trigger_item.sql
- **relations:** dual chrome render paths; rebuild vs rerender semantics; work-item crafting conventions.
- **verify-later:** agent_definitions type='rerender-pages' version 6 default_config; check_refresh_components step.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### plan_sections field deferral semantics and needs_section_data escalation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (To): "Section BACK + both escalations self-closed = the deployed skip_field behaviour works"; W6.4: "two `needs_section_data` items in `needs_human_review` — `plan_sections` could not resolve `illustration_url` … and built each page WITHOUT the section."
- **what:** plan_sections resolves each schema field per its declared `source`; unresolvable required fields defer the WHOLE section, escalate a `needs_section_data` work item into needs_human_review, and the page builds without the section — a loud drop, not silent (guide refinement: fossil pages had been hiding the unresolved dependency). `on_missing: skip_field` is the established optional pattern: omit the field, let the template gate handle it. Edit A fixed the smell that a REQUIRED field with on_missing:skip_field fell to the default defer branch instead of honouring the declared intent. `closeResolvedDataRequest` self-closes escalations once the field resolves post-deploy.
- **sources:** gobatch_01_plan_sections.md#Edit-A; RUNBOOK_scheme_to_components(50).md#W6.4 #W7-FINDINGS; running_notes_scheme_to_components(55).md#Tg #Tl #To; w6_05_section_data_read.sql
- **relations:** image fields optional-with-gate; section data source triad; deployed-binary-predates-disk.
- **verify-later:** plan_sections_action.go on_missing switch (required branch skip_field case present); needs_section_data item lifecycle.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Array item-fields prompt contract (019 migration + ItemFields)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu) 2026-06-21: "Prompt migration already applied"; 019 migration NOTICE "prompt patched"; checkpoint (ss) documents the root cause and fragments verified at positions 2330/3402.
- **what:** Root cause of the differentiators empty cards: the page-content-writer prompt listed array fields with type only, never element shape, so the LLM guessed item keys — `title`/`body` against a template reading `name`/`description` renders empty; FAQ worked only because the natural guess happened to match. Fix: `plan_sections` gains `ItemFields []string` on `llmFieldSpec` via `extractArrayItemFields` (reads both `items` and `item_schema`, sorted for stable prompts); the 019 migration patches the prompt's What-To-Write line and makes the Output-Format JSON skeleton type-aware (`[{ "k": "..." }]` for arrays). The migration is order-independent with the Go deploy ({{if .item_fields}} is simply false until populated), idempotent via a sentinel, aborts if fragments moved, and ships a paired down-migration.
- **sources:** running_notes_checkpoint_ss(1).md; 019_pcw_prompt_item_fields.sql; plan_pcw_item_fields_fix(1).md; RUNBOOK_pcw_item_fields_fix.md
- **relations:** render-time item-key reconciler; component schema-template invariant; SQL change-management pattern.
- **verify-later:** agent_definitions page-content-writer prompt_template markers; plan_sections_action.go ItemFields population.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Render-time item-key reconciler (schema-sourced, non-fatal)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Checkpoint (uu): "Three artefacts now final in outputs … code awaits a chassis image bump" (2026-06-21); no later doc in this unit confirms the image bump for this specific change.
- **what:** A belt-and-braces safety net in `RenderComponentAction`: before the merge, `reconcileGeneratedItemKeys` remaps LLM-drifted array item keys onto the expected ones using case/separator-insensitive matching plus a synonym table (title/body → name/description etc.), never moving a synonym onto a key that is itself expected. Decision 1B hardened it to source expected keys from the component's own `input_schema` (fields with source:"llm" only) instead of the section plan — removing plan-freshness coupling and making the prompt change an optimisation, not a correctness requirement. Decision 2: unrecoverable misses ERROR-and-continue (a missing sub-field is cosmetic; failing a page build is higher blast-radius). Corrected content lands in both rendered HTML and persisted content_data. Cross-file deploy constraint: rides the same image as plan_sections' extractArrayItemFields.
- **sources:** running_notes_checkpoint_uu.md; running_notes_checkpoint_ss(1).md#Fix-delivered; RUNBOOK_pcw_item_fields_fix.md#4-Logs
- **relations:** array item-fields contract; component schema-template invariant; needs_llm routing.
- **verify-later:** v3_site_actions.go reconcileGeneratedItemKeys + wire-in; whether the carrying image shipped (log lines "reconcileGeneratedItemKeys" in writer pods).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### needs_llm routing via detectNeedsLLMContent
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (ss): "the writer sub-workflow … branches on `render_mode == 'agent' OR needs_llm == true`. `needs_llm` is computed by `detectNeedsLLMContent` (v3_site_actions.go ~4095), which returns true for any non-empty `input_schema`."
- **what:** How a section reaches the LLM generation path: the page-content-writer's `process_sections_loop` routes on render_mode OR the computed needs_llm flag, and because detectNeedsLLMContent returns true for any non-empty input_schema, template-mode components with schemas still get LLM content. This made an investigative render_mode flip harmless to revert (differentiators back to 'template') and explains why a 'template' component had generated content at all.
- **sources:** running_notes_checkpoint_ss(1).md#What-we-established #Correction-logged
- **relations:** section data source triad; array item-fields contract.
- **verify-later:** v3_site_actions.go detectNeedsLLMContent; writer sub-workflow branch config.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### No component-level regeneration trigger (whole-page rebuild remedy)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "**No component-level regeneration trigger exists** (user confirmed). So the remedy for the already-deployed broken cards is a whole-index `page-rebuild`, which regenerates *all* index sections … Accepted as the cost."
- **what:** A platform limitation shaping every content-fix decision: there is no mechanism to regenerate one component on one page; the only remedy for bad stored content is a full page rebuild, which rewrites every section's copy (copy churn on hero, FAQ, narrative accepted as cost). Repeatedly parked on the hygiene/backlog lists; interacts with rebuild-vs-rerender (rerender can't be used because it reassembles stored HTML).
- **sources:** running_notes_checkpoint_uu.md#Decisions-taken; RUNBOOK_pcw_item_fields_fix.md#3
- **relations:** rebuild vs rerender semantics; content-governance (regeneration).
- **verify-later:** whether any component-scoped regen item type has since appeared in site_work_items vocabulary.

<!-- SOURCE: U05_content_quality_linking.md -->
### Re-render vs rebuild distinction (which path fixes what)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §2 "Render vs rebuild — what fixes what"; captured in 002/016 per NOTES(44).
- **what:** A load-bearing operational distinction: re-render (page-rerender / rerender-pages) re-applies templates to component data stored at last build; only a rebuild (work item → build-dispatch-loop → page-build-handler → writer) re-runs plan_sections source resolution and the resolver. Consequences: header/footer fixes need only re-render (data rebuilt fresh in Go); hero CTAs and hub URLs need rebuilds (stored data still carries phantoms); P4.2 proved page_rerender preserves sections but does NOT re-resolve schema-sourced CTAs.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#2; NOTES(44) P4.2 result 2026-06-22; running_notes_17(21).md#re-render-mechanics
- **relations:** no-LLM re-render path; interactive clobber (why rebuilds are dangerous); work-item routing.
- **verify-later:** page-rerender vs page-build-handler workflows; rerender_single_page_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-build-handler build path
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Workflow chain confirmed repeatedly from live agent_definitions (HANDOFF_2026-06-09 Key references; NOTES(44) 2026-06-22 step-config dump).
- **what:** The per-page content build orchestrator: ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections (load_page_sections_from_spec) → plan_sections → check_has_ready_sections → spawn/call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender → deploy_page → complete. One linear flow, no item_type branch; deploys by spawning page-rerender + git commit, one commit per page. `spec.mode='recreate'` loads the adoption crawl to preserve original copy; `spec.suggestion` feeds writer rewrite_guidance.
- **sources:** HANDOFF_2026-06-09(2).md#key-references; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4; running_notes_17(21).md#page-build-handler-contract
- **relations:** page-content-writer; save_page_sections; silent-completion (complete_error exit); interactive clobber.
- **verify-later:** page-build-handler default_config.workflow.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer (task specialist, no persistence)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 9: writer def read — "no save_page_sections, no update_status, no deploy".
- **what:** The content-generation specialist: spawn_research → load_site_specs → prepare_link_context → build_render_context → process_sections_loop (render/generate per section) → resolve_links → select_sections → compile_page. It only produces content (per-section outputs + compiled sections_metadata); persistence and deploy live in the page-build-handler wrapper — routing a discovery item straight at the writer can never deploy a page (a documented stale-handler bug in a dormant check). Its `complete` step's singular output_field was the Part-1 trigger.
- **sources:** running_notes_15(12).md#part-9; HANDOFF_2026-06-09(2).md#key-references; NOTES(44) writer key findings
- **relations:** result-contract; resolver wiring; recreate mode; prepare_link_context gap.
- **verify-later:** page-content-writer default_config; compile_page_sections_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### save_page_sections: DELETE+INSERT persistence with layered guards
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 Part 4 "DONE — patched save_page_sections (Layers 1+2) deployed on v1.0.1077".
- **what:** The single save path for page sections (three callers: page-build-handler, page-rerender, tool-recreation-handler): reads structured sections_metadata (primary) or an HTML-parse fallback (saveSectionsExtractFromHTML — extended with a single-fragment fallback after the `<div>`-not-`<section>` tool loss), snapshots page_component_history, then DELETE+INSERT of the produced set. Guards accreted through this unit: the content-regression guard (existing stripped text >200 and new < existing/4 → error — correct to refuse a wipe, threshold scales with page size); Layer 1 interactivity guard (existing page interactive, new set not → "interactivity regression blocked"); Layer 2 carry-forward of non-spec interactive sections (keep/replace/re-append by slot); source_item_id stamping into history via config-driven work_item_id_field.
- **sources:** NOTES(44) 2026-06-24 patch sessions; HANDOFF_2026-06-15(2).md#3; game_lost_its_tool/001_context; running_notes_17(21).md#index-save-read
- **relations:** interactive clobber; index stale-rebuild defect; save-failure visibility; content_data⊕resolved_data model.
- **verify-later:** save_page_sections_action.go (guards at ~L251-287, DELETE+INSERT ~L322-393, history ~L296-310); page_component_history.source_item_id population.

<!-- SOURCE: U05_content_quality_linking.md -->
### Interactive-page clobber failure class (spec-planned rebuild drops the tool)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 "Part 4 — DONE … game-pathfinding A* tool re-created (interactive ~20KB, deployed 2026-06-26) and now protected from re-clobber".
- **what:** An interactive tool/game exists ONLY as bespoke `<canvas>`/JS markup in page_components.rendered_html — not in the page spec, not LLM-regeneratable. ANY full rebuild (needs_page/needs_content_page/content_rewrite/link_resolution_rebuild/admin regenerate) plans from the spec, omits the tool, and save's DELETE+INSERT drops it (a links-only maintenance task destroyed a working A* game). Text-based regression guards missed it because the loss is markup/JS, not prose. Fix landed at the save path (Layers 1+2 above), NOT routing (P4.2 falsified the page_rerender reroute) and NOT the planner (which traffics in section-name skeletons). Interactivity signal: rendered_html ILIKE canvas/game-container/tool-page (data-component alone is not a signal). Prior partial fix: findPreservedComponentIDs preserved only render_action components.
- **sources:** PLAN_pathfinding_missing_game.md; NOTES(44) 2026-06-22 clobber sessions; game_lost_its_tool/001_context; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4
- **relations:** save_page_sections guards; tool-recreation-handler; item_key mis-key (same page); sectionHasVisibleContent (second silent-drop path).
- **verify-later:** page_component_history for game-pathfinding; save_page_sections Layer 1/2 code; regression test: link rebuild on pathfinding blocks not clobbers.

<!-- SOURCE: U05_content_quality_linking.md -->
### No-LLM re-render path (rerender_page_sections, Part 2 / Option Y)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 2 — DEPLOYED 2026-06-21; image_landed verified; … finish P2.4–P2.7".
- **what:** A field-re-resolve + re-render capability that avoids the full LLM writer: an image landing or resolvable section data previously forced a full content rebuild (LLM spend + regression-guard exposure). New rerender_page_sections action re-renders ALL of a page's sections from stored content_data overlaid with FRESH resolved_data (reusing plan_sections' side-effect-free planSection/sourceResolver — route ii), renders via RenderTemplate, emits the exact sections_metadata shape save reads. Slotted into page-rerender as a pre-pass gated by spec.reason (image_landed / section_data_resolved); flag_page_image_rebuild + reconcile_section_data repointed to emit page_rerender-type items (closing their type/key mismatch). NULL content_data on any section → escalate the whole page to needs_page (self-healing one-time full rebuild that backfills content_data). Y-lean render context chosen after confirming templates use only content_data + CSS-var colours. Design alternatives recorded: Option X (no-LLM branch inside the writer) rejected; re-render-affected-section-only rejected in favour of re-render-all.
- **sources:** NOTES(44)#part-A sections (decision trail 2026-06-19→21); RUNBOOK_gamesdesign_index_rebuild(29).md#part-2; HANDOFF_page_pipeline(11).md#5
- **relations:** content_data⊕resolved_data model; re-render vs rebuild; P4.2 (does NOT re-resolve schema-sourced CTAs — that stayed with the writer path).
- **verify-later:** rerender_page_sections_action.go; page-rerender check_rerender_mode wiring; P2.4–P2.7 test outcomes.

<!-- SOURCE: U05_content_quality_linking.md -->
### content_data ⊕ resolved_data persistence model
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19: "UNKNOWN NOW RESOLVED … content_data IS complete enough to re-render from" (RenderComponentAction deliberate merge, per its comment).
- **what:** RenderComponentAction builds a section's content_data as LLM copy (content_from) overlaid with resolved_data (merge_with) — deliberately persisting resolved items/urls/labels alongside the copy, next to rendered_html. This is what makes no-LLM re-rendering possible (render again from stored content + fresh resolution). Corollary schema fact that cost a wrong turn: there is NO page_components.resolved_data column — resolved values live inside content_data.
- **sources:** NOTES(44) 2026-06-19; HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_17(21).md#schema-correction
- **relations:** no-LLM re-render; system-stats key-contract break (content_data keys vs template keys).
- **verify-later:** v3_site_actions.go RenderComponentAction (~L1372); page_components schema.

<!-- SOURCE: U05_content_quality_linking.md -->
### Index stale-rebuild defect (writer output ≠ save input path)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "index rebuild VERIFIED on a real build … sections_metadata array, sm_count=5 … Part 1 result contract working".
- **what:** The unit's opening mystery: index rebuilds completed and git-committed while all five page_components stayed frozen at 06-06. The investigation is a model falsification chain — load, concurrent deploy, claim-lease duration, caller timeout, component locks, and the content-regression guard (writer measured at 33k chars >> the 5760 threshold) were each raised and eliminated — landing on the writer's compiled result being replaced by the size-limit stub before save (the Part-1 result-contract bug). Resolved by the flatten fix; verified end-to-end 06-19 and 06-24 (deployed hero "Your Probability Maths Is Wrong", real hub CTAs).
- **sources:** HANDOFF_2026-06-15_index_stale_rebuild(2).md; NOTES_gamesdesign_silent_norebuild(44).md; running_notes_17(21).md#index-deep-dive
- **relations:** result-contract resolution; silent-completion; save-failure visibility; "git committed ≠ new content" heuristic.
- **verify-later:** orchestration_states 472eed7d/4e0b339a; page_components index timestamps.

<!-- SOURCE: U05_content_quality_linking.md -->
### Save-failure visibility fix (mark_save_failed) + engine error_step ambiguity
- **category:** NEW:page-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** page_build_handler_save_failure_visible.sql delivered 2026-06-15 with "unmet prerequisite (which error_step the engine reads)"; no later doc records applying it.
- **what:** Routes save_sections' error to a new mark_save_failed step (fail_work_item → needs_human_review) instead of complete_error, so a blocked/failed save surfaces instead of laundering into `complete`. Blocked on a real engine unknown: the save_sections step carries error_step in TWO places (step-level and config-level) and it is unconfirmed which the workflow engine honours for routing — "DO NOT GUESS". Companion (also unbuilt): gate deploy_page on sections_saved>0 so a no-write save can't re-commit stale components.
- **sources:** page_build_handler_save_failure_visible.sql; HANDOFF_2026-06-15(2).md#3-bugs; running_notes_17(21).md#FIX-written
- **relations:** silent-completion family; complete_error semantics (Fix B, deferred).
- **verify-later:** whether the SQL was ever applied; chassis engine error_step resolution (step.ErrorStep vs config["error_step"]).

<!-- SOURCE: U05_content_quality_linking.md -->
### Sectionless-page durability stack (2b sibling fallback + S1 check + S2 flag)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-09(2): "Durability code WRITTEN this session (NOT yet deployed)"; running_notes_16(1) carries it as deploy-pending; later docs never record S1 enablement.
- **what:** A planned page reaching build with empty pages.sections silently completed as success ("Content writer skipped — page has no sections defined"). Three-layer durability: 2b — load_page_sections_from_spec gains a final fallback synthesising the layout from a same-role sibling's modal section list (layout skeleton only, WARN-logged, writes pages.sections); S1 — new discovery check check_sectionless_pages flags current-plan pages with empty sections that a sibling can fix and re-triggers to page-build-handler; S2 — check_has_ready_sections ELSE repointed from complete_error to mark_no_sections (needs_human_review). Decisive build fact: pages.sections is the build-read field; site_plan_sections is NOT on the build path (plan hygiene only). Also documented: checkEmptyPageSections is dormant, half-superseded code (wrapper never enabled, wrong handler) — a dedicated check was chosen over reviving it.
- **sources:** running_notes_15(12).md (whole arc); package_module/running_notes_16_adoption_sections.md (same content); HANDOFF_2026-06-09(2).md
- **relations:** Fix A prerequisite; silent-completion; skinner-box case; complete_error semantics.
- **verify-later:** load_page_sections_from_spec_action.go 2b fallback deployed?; completeness-discovery-agent checks array contains "sectionless_pages"?; page-build-handler mark_no_sections step.

<!-- SOURCE: U05_content_quality_linking.md -->
### plan_sections field-source resolution semantics (on_missing, required, defer)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14p "RESOLVED 2026-06-06" with code-confirmed semantics.
- **what:** The engine semantics governing when a section renders, defers, or drops a field: query.* fields return a non-nil empty slice (never defer, never consult on_missing); on_missing defaults to skip_field; the REQUIRED-field switch has NO skip_field case, so a required field defaulting to skip_field falls to defer — the trap that hid the guides hub (guide-list cta_url required=true + unpopulated spec source deferred the whole section). Fix chosen at the component (required=false) not the engine (the defer-for-safety default is defensible). Related deferral machinery: needs_section_data items for query-resolvable gaps, with reconcile_section_data as the designed loop-closer — registered but STILL UNHOSTED (nothing calls it; query-resolvable items sit at needs_human_review).
- **sources:** running_notes_14(26).md#part-14o-14p; HANDOFF_2026-06-09(2).md#june-02-actions; running_notes_16_content_quality_and_internal_linking(1).md#carried-forward
- **relations:** B4/B5 (query fields + template gates); no-LLM re-render (reuses planSection); component schema contracts.
- **verify-later:** plan_sections_action.go on_missing switches; reconcile_section_data host (still none?); guide-list/blog-listing required flags.

<!-- SOURCE: U05_content_quality_linking.md -->
### sectionHasVisibleContent assembler filter
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "approx_visible_len = 0 … the filter correctly drops it"; "the filter is right".
- **what:** rerender_single_page's getPageSections strips style/script/tags/entities and DROPS any section with ≤10 visible chars (WARN-only). Verified correct for text-empty shells (system-stats), but recognised as a SECOND silent-drop path for interactive content independent of save_page_sections — a low-prose game could be stripped at assembly even after the carry-forward preserves it in the DB. Open question noted: should it share the Part-4 interactivity signal rather than a pure text heuristic (the same text-heuristic blind spot as the regression guard).
- **sources:** NOTES(44) 2026-06-24 system-stats/assembler sessions
- **relations:** interactive clobber; system-stats break; text-heuristic blind spot family.
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent; game-auto-battler visible_len.

<!-- SOURCE: U05_content_quality_linking.md -->
### A1 — adopted tools/games never deployed a file (parser + status-churn chain)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 10: "A1 VERIFIED CLOSED … all five games committed … tools deploy".
- **what:** No tool/game page produced a deployable file because saveSectionsExtractFromHTML extracted only `<section>` blocks while recreate_tool emits `<div class="tool-page">` → zero page_components → assemblePage returned "" → rerender skipped → no git commit, all while work items read complete. Fixed by the single-fragment fallback (whole fragment as one section, guarded against full documents), coupled with the deployed→needs_rebuild flip removal and deploy-time plan-version stamping. Established the durable read: getPageSections reads page_components, not pages.sections — "has sections" and "has rendered components" are different facts.
- **sources:** running_notes_14(26).md#part-7-10
- **relations:** save_page_sections; interactive clobber (later same-family loss); tool-recreation-handler.
- **verify-later:** saveSectionsExtractFromHTML fallback; deployed repo /tools//games/ trees.

<!-- SOURCE: U05_content_quality_linking.md -->
### UpdatePageStatusAction zero-component deploy guard (Option B)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2 "Option B delivered" + Part 11/12 arc; reaper comment later cites the hardening as in place.
- **what:** A page must never be marked deployed with zero real components: the deployed branch is guarded by pageHasComponents (EXISTS on page_components with non-null component_id + non-empty rendered_html); on zero components it refuses `deployed`, sets needs_rebuild + clears the plan stamp, fail-open on check errors. Keeps build_status honest as evidence for the reaper (the homepage had been 'deployed' with 0 components and no file).
- **sources:** running_notes_14(26).md#part-11-12
- **relations:** evidence-gated reaper; auto-complete false positive; built_from_plan_version stamp.
- **verify-later:** v3_site_actions.go UpdatePageStatusAction deployed branch.

<!-- SOURCE: U05_content_quality_linking.md -->
### Deploy-observability bookkeeping gap
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** NOTES(44) 2026-06-21: "Telemetry gap, non-blocking: page_components.deploy_commit blank and pages.last_built_at NULL though deployed_at is set — deploy step isn't writing those back."
- **what:** The deploy path sets deployed_at but never writes page_components.deploy_commit or pages.last_built_at, and content_hash is empty on investigated pages — so change detection falls back to updated_at + rendered_html length. Folded into a later deploy-observability fix; small but it repeatedly complicated verification.
- **sources:** NOTES(44) 2026-06-21 update; running_notes_17(21).md (content_hash note)
- **relations:** debugging heuristics (git committed ≠ new content); save_page_sections.
- **verify-later:** deploy_page/git_commit write-backs.

<!-- SOURCE: U23_docs_root_vonc.md -->
### render_mode derivation + LLM routing condition (migration 002)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Migration table 2026-06-24: "002 DONE — check_render_mode condition fixed"; deriveRenderMode code deployed 2026-06-24 ~15:00; hero LLM content confirmed on the rebuilt index.
- **what:** `StoreGeneratedComponentAction` originally hardcoded `render_mode='template'` on every component, making the LLM content path permanently unreachable; `deriveRenderMode(inputSchemaJSON)` now returns 'agent' iff any schema field has source='llm'. Separately, page-content-writer's `check_render_mode` condition was reading a never-populated field; migration 002 changed it to `current_section.llm_field_specs != null` (populated by plan_sections from the schema), routing any section with LLM fields to content generation for all sites. Note: render_mode is NOT what routes sections (a later red herring) — llm_field_specs is.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#4-render_mode-hardcoded + #migration-002-outcome + #2026-07-02-~19:20; docs/RUNBOOK_vonc_migrations(14).md#background-migration-002
- **relations:** render_mode sweep (dropped); plan_sections deferral (render_mode red herring)
- **verify-later:** store_generated_component_action.go deriveRenderMode; page-content-writer agent_definition check_render_mode condition

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component-table render_mode sweep (65 components) — dropped migration
- **category:** NEW:page-build-pipeline
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_vonc.md base vs (1) diff: "Migration 002 (render_mode sweep across 65 components) is DROPPED"; PLAN_vonc_next_steps(1): "The 65-component render_mode update is DROPPED; existing components are fine as-is."
- **what:** The first plan for fixing LLM routing was a DB sweep updating `render_mode` on 65 existing library components. Dropped once it was established that workflow routing reads `llm_field_specs` (set by plan_sections from the schema), not the stored render_mode — so only the agent_definition condition needed fixing and existing component rows were fine as-is. Captures the earliest documented shape of the fix; useful provenance for why component rows still carry historical render_mode values.
- **sources:** docs/RUNNING_NOTES_vonc.md#4 (pre-edit base); docs/PLAN_vonc_next_steps(1).md#p1; docs/RUNBOOK_vonc_migrations(1).md (earlier "Fix render_mode on components" migration heading, dropped from later versions)
- **relations:** render_mode derivation + routing condition (the replacement)
- **verify-later:** none (historical)

<!-- SOURCE: U23_docs_root_vonc.md -->
### plan_sections readiness triage and deferral semantics
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Confirmed in code 2026-07-02 (planSection read); fix validated end-to-end 2026-07-03 (index went 3 → 6 sections after populating cta spec + relaxing illustration field); 016b §9 entry.
- **what:** plan_sections classifies each planned section by resolving its schema fields: source=llm always available; query.*/renderer/static resolve at render time or fall back; any other source runs the resolver. A REQUIRED field whose source doesn't resolve hits the on_missing switch, whose `default:` case DEFERS the section ("default to defer for safety") — and empty on_missing defaults to skip_field, which is not a case in the required switch, so it defers. save_page_sections then persists only the ready set, dropping deferred sections' page instances. Authoring rule: never `required=true` + `on_missing=skip_field`; fix by populating the site data source or degrading the field.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:20 + #2026-07-02-~19:35; docs/016b_debugging_guide_merged(3).md#regenerated-content-section-is-deferred; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f
- **relations:** site_specs cta aspect; resolver asset kinds gap; plan-driven rebuild + clobber
- **verify-later:** plan_sections_action.go planSection on_missing switch; save_page_sections_action.go

<!-- SOURCE: U23_docs_root_vonc.md -->
### Plan-driven rebuild + interactive/deferred-section clobber (carry-forward fix)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** 016b: "Part 4... fix WRITTEN (un-deployed)" — Layer 1 interactivity guard + Layer 2 carry-forward in patched save_page_sections_action.go; the 2026-07-02 vonc rebuild demonstrated the drop live (6 planned → 3 saved, brief-explanation instance gone).
- **what:** A needs_page rebuild is PLAN-driven, not pending-driven: load sections from the plan → triage → the writer renders ALL ready planned sections → save_page_sections DELETE+INSERTs the page's components. Sections present in page_components but absent from the plan (interactive tools stored only as rendered_html) or deferred by triage get silently dropped. Fix (written, not deployed): interactivity-aware guard blocking a non-interactive set replacing a deployed interactive one, plus carry-forward of existing interactive sections; three callers to bump (page-build-handler, page-rerender, tool-recreation-handler); plus source_item_id stamping for traceability.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + 2026-06-24 update); docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:40 + #2026-07-02-~19:00; docs/PLAN_spark_provocation_pipeline.md#standing-constraints
- **relations:** plan_sections deferral; page_components single-writer (save_page_sections); interactive tool pages stored as rendered_html
- **verify-later:** save_page_sections_action.go (is the guard/carry-forward deployed?); page_component_history.source_item_id

<!-- SOURCE: U23_docs_root_vonc.md -->
### complete_error silent-success family (page build completes having built nothing)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Mechanism fully confirmed 2026-07-06 (workflow dump); the defect is live ("an error path implemented as a SUCCESSFUL completion"); preventions listed as backlog item 1 in the HANDOFF, not built.
- **what:** page-build-handler routes zero-ready-sections to a step literally named `complete_error` — a complete_workflow with success_message "Content writer skipped — page has no sections defined" — so builds against a section-less page complete cleanly having done nothing. Ten silent no-ops (33–65s completes) hid a 404 CTA destination for two weeks; a work-item result carrying ONLY `site_record` (healthy: `[sections_saved, deploy_result]`) is the diagnostic signature. Variants: plan row naming a nonexistent component also passes silently. Preventions (aspirational): complete_error fails loudly or raises needs_plan_sections; auditor linked+planned+URL-presence rules; `last_built_at` is never written by build or rerender (dead column — write it or drop it).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed + #2026-07-08; docs/016b_debugging_guide_merged(3).md#page-build-completes-having-built-nothing; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§9
- **relations:** three section sources; planner ≥1-section invariant; trust-the-artifact doctrine
- **verify-later:** page-build-handler default_config complete_error step; pages.last_built_at writes anywhere in the codebase

<!-- SOURCE: U23_docs_root_vonc.md -->
### load_page_record lookup semantics (name-first, page_id fallback)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** load_page_record_action.go read in full 2026-07-06: "Lookup priority: page_name (site_id+name) first; page_id only if name empty/bogus (nonPageNames)... returns sections PARSED FROM pages.sections + section_count."
- **what:** The build's page lookup resolves by page_name against `pages.name` first, falling back to page_id only for empty/bogus names, and returns the page's own `sections` jsonb with a count — which is what gates the zero-sections branch. Schema gotcha bundled with it: `pages` has `name` not `page_name`; work-item specs use domain/page_id/filename/page_name.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-load_page_record-read; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** complete_error family; three section sources
- **verify-later:** load_page_record_action.go nonPageNames list

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two re-render paths + assemble-only rerender distinction
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Doc-003-derived and header-confirmed 2026-07-09: rerender_single_page "confirmed ASSEMBLE-ONLY" from its own header; light-path escalation rule quoted from 003.
- **what:** Three distinct "rerender" operations that must not be conflated: (1) FULL rebuild — needs_page → page-build-handler → page-content-writer (LLM regenerates copy); (2) LIGHT re-render — rerender_page_sections behind a page_rerender item: re-renders every section from EXISTING content_data via RenderComponentAction, no LLM, and escalates the whole page to a full rebuild when content_data is NULL; (3) ASSEMBLE-ONLY — rerender_single_page (the habitual rerender-*.sh trigger): reassembles stored page_components.rendered_html + stored site_components chrome and deploys; template-only edits will NOT appear through it. Mode-B sections likely have NULL content_data, making the light path escalate — the deciding probe for edit sequencing.
- **sources:** docs/PLAN_provocation-card(3).md#method-corrected; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/016b_debugging_guide_merged(3).md#open-threads (Part 2)
- **relations:** sanctioned edit paths; assemble-time visible-content filter; two chrome assembly paths
- **verify-later:** rerender_page_sections_action.go escalation branch; page_rerender item routing

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Rebuild vs rerender semantics and stale-render fossilisation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** CHECK 4 RESULTS: "the deployed pages are RERENDER output carrying stale stored renders … deployed hero consumes legacy `var(--accent-color, #0f3460)` … A full page-build-handler rebuild is required; `needs_rerender` would re-fossilise them" — settling the 016-vs-026 documented tension "by direct evidence".
- **what:** Two distinct page-refresh routes with different semantics: `needs_rerender` (handler rerender-pages) reassembles stored `page_components.rendered_html` and injects stored chrome — it does NOT re-render component templates, so template changes never land and old renders fossilise; a full build (`site_work_items` insert: pipeline=build, handler_agent=page-build-handler, status=triaged) re-runs plan_sections and re-renders everything. idea.uk had lived for weeks on reassemblies of early renders while the library advanced — the fossil tell is a legacy variable name in deployed HTML (`var(--accent-color`), and its count going to 0 is the proof a rebuild truly re-rendered. Docs 026 ("rerender reflects new template") and 016 ("patches in place") disagreed; evidence sided with 016. Related hazard from the migration sketch: a content rebuild can de-tool a tool page (page-content-writer regenerates from plan_sections, which does not know the interactive tool).
- **sources:** RUNBOOK_scheme_to_components(50).md#CHECK-4-RESULTS #Migration-backfill; running_notes_scheme_to_components(55).md#So #Sh(migration route); HANDOFF_scheme_to_components_for_claude_code(1).md#Invariant (item 5)
- **relations:** dual chrome render paths; work-item crafting conventions; deployed-binary-predates-disk class.
- **verify-later:** rerender-pages vs page-build-handler workflow definitions in agent_definitions; 016/026 doc reconciliation.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### rerender-pages v6 workflow (refresh_site_components gate)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (Tb): "Workflow (v6) fully read: gate `input_data.spec.refresh_site_components == true` → FORCED chrome render (header/footer/head) → js snippets render+commit → blog listing → get_pages (deployed+active) → create_rerender_items (per-page) → update_site_status deployed → complete."
- **what:** The site-wide rerender agent: one work item fans out to a forced chrome re-render (the only workflow passing `force_rerender: true` — pageflow-builder and site-work-orchestrator pass false, which explains fossilised chrome surviving full builds), JS snippet rendering, blog-listing rebuild, then per-page rerender items that the build dispatch loop drains; it ends by marking the site deployed. `spec.function`/`component_id` are consumed nowhere in v6. The real producer of such items is `store_generated_component` on regeneration (one deduped `needs_rerender` per affected site, item_key `component_regen_rerender:<uuid>`).
- **sources:** running_notes_scheme_to_components(55).md#Ta #Tb; w4b_02_read_triggers.sql; w4b_04_trigger_item.sql
- **relations:** dual chrome render paths; rebuild vs rerender semantics; work-item crafting conventions.
- **verify-later:** agent_definitions type='rerender-pages' version 6 default_config; check_refresh_components step.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### plan_sections field deferral semantics and needs_section_data escalation
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Notes (To): "Section BACK + both escalations self-closed = the deployed skip_field behaviour works"; W6.4: "two `needs_section_data` items in `needs_human_review` — `plan_sections` could not resolve `illustration_url` … and built each page WITHOUT the section."
- **what:** plan_sections resolves each schema field per its declared `source`; unresolvable required fields defer the WHOLE section, escalate a `needs_section_data` work item into needs_human_review, and the page builds without the section — a loud drop, not silent (guide refinement: fossil pages had been hiding the unresolved dependency). `on_missing: skip_field` is the established optional pattern: omit the field, let the template gate handle it. Edit A fixed the smell that a REQUIRED field with on_missing:skip_field fell to the default defer branch instead of honouring the declared intent. `closeResolvedDataRequest` self-closes escalations once the field resolves post-deploy.
- **sources:** gobatch_01_plan_sections.md#Edit-A; RUNBOOK_scheme_to_components(50).md#W6.4 #W7-FINDINGS; running_notes_scheme_to_components(55).md#Tg #Tl #To; w6_05_section_data_read.sql
- **relations:** image fields optional-with-gate; section data source triad; deployed-binary-predates-disk.
- **verify-later:** plan_sections_action.go on_missing switch (required branch skip_field case present); needs_section_data item lifecycle.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Array item-fields prompt contract (019 migration + ItemFields)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu) 2026-06-21: "Prompt migration already applied"; 019 migration NOTICE "prompt patched"; checkpoint (ss) documents the root cause and fragments verified at positions 2330/3402.
- **what:** Root cause of the differentiators empty cards: the page-content-writer prompt listed array fields with type only, never element shape, so the LLM guessed item keys — `title`/`body` against a template reading `name`/`description` renders empty; FAQ worked only because the natural guess happened to match. Fix: `plan_sections` gains `ItemFields []string` on `llmFieldSpec` via `extractArrayItemFields` (reads both `items` and `item_schema`, sorted for stable prompts); the 019 migration patches the prompt's What-To-Write line and makes the Output-Format JSON skeleton type-aware (`[{ "k": "..." }]` for arrays). The migration is order-independent with the Go deploy ({{if .item_fields}} is simply false until populated), idempotent via a sentinel, aborts if fragments moved, and ships a paired down-migration.
- **sources:** running_notes_checkpoint_ss(1).md; 019_pcw_prompt_item_fields.sql; plan_pcw_item_fields_fix(1).md; RUNBOOK_pcw_item_fields_fix.md
- **relations:** render-time item-key reconciler; component schema-template invariant; SQL change-management pattern.
- **verify-later:** agent_definitions page-content-writer prompt_template markers; plan_sections_action.go ItemFields population.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### Render-time item-key reconciler (schema-sourced, non-fatal)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Checkpoint (uu): "Three artefacts now final in outputs … code awaits a chassis image bump" (2026-06-21); no later doc in this unit confirms the image bump for this specific change.
- **what:** A belt-and-braces safety net in `RenderComponentAction`: before the merge, `reconcileGeneratedItemKeys` remaps LLM-drifted array item keys onto the expected ones using case/separator-insensitive matching plus a synonym table (title/body → name/description etc.), never moving a synonym onto a key that is itself expected. Decision 1B hardened it to source expected keys from the component's own `input_schema` (fields with source:"llm" only) instead of the section plan — removing plan-freshness coupling and making the prompt change an optimisation, not a correctness requirement. Decision 2: unrecoverable misses ERROR-and-continue (a missing sub-field is cosmetic; failing a page build is higher blast-radius). Corrected content lands in both rendered HTML and persisted content_data. Cross-file deploy constraint: rides the same image as plan_sections' extractArrayItemFields.
- **sources:** running_notes_checkpoint_uu.md; running_notes_checkpoint_ss(1).md#Fix-delivered; RUNBOOK_pcw_item_fields_fix.md#4-Logs
- **relations:** array item-fields contract; component schema-template invariant; needs_llm routing.
- **verify-later:** v3_site_actions.go reconcileGeneratedItemKeys + wire-in; whether the carrying image shipped (log lines "reconcileGeneratedItemKeys" in writer pods).

<!-- SOURCE: U03_idea_uk_section_data.md -->
### needs_llm routing via detectNeedsLLMContent
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (ss): "the writer sub-workflow … branches on `render_mode == 'agent' OR needs_llm == true`. `needs_llm` is computed by `detectNeedsLLMContent` (v3_site_actions.go ~4095), which returns true for any non-empty `input_schema`."
- **what:** How a section reaches the LLM generation path: the page-content-writer's `process_sections_loop` routes on render_mode OR the computed needs_llm flag, and because detectNeedsLLMContent returns true for any non-empty input_schema, template-mode components with schemas still get LLM content. This made an investigative render_mode flip harmless to revert (differentiators back to 'template') and explains why a 'template' component had generated content at all.
- **sources:** running_notes_checkpoint_ss(1).md#What-we-established #Correction-logged
- **relations:** section data source triad; array item-fields contract.
- **verify-later:** v3_site_actions.go detectNeedsLLMContent; writer sub-workflow branch config.

<!-- SOURCE: U03_idea_uk_section_data.md -->
### No component-level regeneration trigger (whole-page rebuild remedy)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Checkpoint (uu): "**No component-level regeneration trigger exists** (user confirmed). So the remedy for the already-deployed broken cards is a whole-index `page-rebuild`, which regenerates *all* index sections … Accepted as the cost."
- **what:** A platform limitation shaping every content-fix decision: there is no mechanism to regenerate one component on one page; the only remedy for bad stored content is a full page rebuild, which rewrites every section's copy (copy churn on hero, FAQ, narrative accepted as cost). Repeatedly parked on the hygiene/backlog lists; interacts with rebuild-vs-rerender (rerender can't be used because it reassembles stored HTML).
- **sources:** running_notes_checkpoint_uu.md#Decisions-taken; RUNBOOK_pcw_item_fields_fix.md#3
- **relations:** rebuild vs rerender semantics; content-governance (regeneration).
- **verify-later:** whether any component-scoped regen item type has since appeared in site_work_items vocabulary.

<!-- SOURCE: U05_content_quality_linking.md -->
### Re-render vs rebuild distinction (which path fixes what)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** RUNBOOK_linking_phantom_fixes(7) §2 "Render vs rebuild — what fixes what"; captured in 002/016 per NOTES(44).
- **what:** A load-bearing operational distinction: re-render (page-rerender / rerender-pages) re-applies templates to component data stored at last build; only a rebuild (work item → build-dispatch-loop → page-build-handler → writer) re-runs plan_sections source resolution and the resolver. Consequences: header/footer fixes need only re-render (data rebuilt fresh in Go); hero CTAs and hub URLs need rebuilds (stored data still carries phantoms); P4.2 proved page_rerender preserves sections but does NOT re-resolve schema-sourced CTAs.
- **sources:** RUNBOOK_linking_phantom_fixes(7).md#2; NOTES(44) P4.2 result 2026-06-22; running_notes_17(21).md#re-render-mechanics
- **relations:** no-LLM re-render path; interactive clobber (why rebuilds are dangerous); work-item routing.
- **verify-later:** page-rerender vs page-build-handler workflows; rerender_single_page_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-build-handler build path
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Workflow chain confirmed repeatedly from live agent_definitions (HANDOFF_2026-06-09 Key references; NOTES(44) 2026-06-22 step-config dump).
- **what:** The per-page content build orchestrator: ensure_site_record → load_page_record → check_page_found → load_existing_content → load_spec_sections (load_page_sections_from_spec) → plan_sections → check_has_ready_sections → spawn/call_content_writer → check_content_produced → validate_content → save_sections → update_status → spawn_rerender → deploy_page → complete. One linear flow, no item_type branch; deploys by spawning page-rerender + git commit, one commit per page. `spec.mode='recreate'` loads the adoption crawl to preserve original copy; `spec.suggestion` feeds writer rewrite_guidance.
- **sources:** HANDOFF_2026-06-09(2).md#key-references; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4; running_notes_17(21).md#page-build-handler-contract
- **relations:** page-content-writer; save_page_sections; silent-completion (complete_error exit); interactive clobber.
- **verify-later:** page-build-handler default_config.workflow.

<!-- SOURCE: U05_content_quality_linking.md -->
### page-content-writer (task specialist, no persistence)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 9: writer def read — "no save_page_sections, no update_status, no deploy".
- **what:** The content-generation specialist: spawn_research → load_site_specs → prepare_link_context → build_render_context → process_sections_loop (render/generate per section) → resolve_links → select_sections → compile_page. It only produces content (per-section outputs + compiled sections_metadata); persistence and deploy live in the page-build-handler wrapper — routing a discovery item straight at the writer can never deploy a page (a documented stale-handler bug in a dormant check). Its `complete` step's singular output_field was the Part-1 trigger.
- **sources:** running_notes_15(12).md#part-9; HANDOFF_2026-06-09(2).md#key-references; NOTES(44) writer key findings
- **relations:** result-contract; resolver wiring; recreate mode; prepare_link_context gap.
- **verify-later:** page-content-writer default_config; compile_page_sections_action.go.

<!-- SOURCE: U05_content_quality_linking.md -->
### save_page_sections: DELETE+INSERT persistence with layered guards
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 Part 4 "DONE — patched save_page_sections (Layers 1+2) deployed on v1.0.1077".
- **what:** The single save path for page sections (three callers: page-build-handler, page-rerender, tool-recreation-handler): reads structured sections_metadata (primary) or an HTML-parse fallback (saveSectionsExtractFromHTML — extended with a single-fragment fallback after the `<div>`-not-`<section>` tool loss), snapshots page_component_history, then DELETE+INSERT of the produced set. Guards accreted through this unit: the content-regression guard (existing stripped text >200 and new < existing/4 → error — correct to refuse a wipe, threshold scales with page size); Layer 1 interactivity guard (existing page interactive, new set not → "interactivity regression blocked"); Layer 2 carry-forward of non-spec interactive sections (keep/replace/re-append by slot); source_item_id stamping into history via config-driven work_item_id_field.
- **sources:** NOTES(44) 2026-06-24 patch sessions; HANDOFF_2026-06-15(2).md#3; game_lost_its_tool/001_context; running_notes_17(21).md#index-save-read
- **relations:** interactive clobber; index stale-rebuild defect; save-failure visibility; content_data⊕resolved_data model.
- **verify-later:** save_page_sections_action.go (guards at ~L251-287, DELETE+INSERT ~L322-393, history ~L296-310); page_component_history.source_item_id population.

<!-- SOURCE: U05_content_quality_linking.md -->
### Interactive-page clobber failure class (spec-planned rebuild drops the tool)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** HANDOFF_page_pipeline(11) §2 "Part 4 — DONE … game-pathfinding A* tool re-created (interactive ~20KB, deployed 2026-06-26) and now protected from re-clobber".
- **what:** An interactive tool/game exists ONLY as bespoke `<canvas>`/JS markup in page_components.rendered_html — not in the page spec, not LLM-regeneratable. ANY full rebuild (needs_page/needs_content_page/content_rewrite/link_resolution_rebuild/admin regenerate) plans from the spec, omits the tool, and save's DELETE+INSERT drops it (a links-only maintenance task destroyed a working A* game). Text-based regression guards missed it because the loss is markup/JS, not prose. Fix landed at the save path (Layers 1+2 above), NOT routing (P4.2 falsified the page_rerender reroute) and NOT the planner (which traffics in section-name skeletons). Interactivity signal: rendered_html ILIKE canvas/game-container/tool-page (data-component alone is not a signal). Prior partial fix: findPreservedComponentIDs preserved only render_action components.
- **sources:** PLAN_pathfinding_missing_game.md; NOTES(44) 2026-06-22 clobber sessions; game_lost_its_tool/001_context; RUNBOOK_gamesdesign_index_rebuild(29).md#part-4
- **relations:** save_page_sections guards; tool-recreation-handler; item_key mis-key (same page); sectionHasVisibleContent (second silent-drop path).
- **verify-later:** page_component_history for game-pathfinding; save_page_sections Layer 1/2 code; regression test: link rebuild on pathfinding blocks not clobbers.

<!-- SOURCE: U05_content_quality_linking.md -->
### No-LLM re-render path (rerender_page_sections, Part 2 / Option Y)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 2 — DEPLOYED 2026-06-21; image_landed verified; … finish P2.4–P2.7".
- **what:** A field-re-resolve + re-render capability that avoids the full LLM writer: an image landing or resolvable section data previously forced a full content rebuild (LLM spend + regression-guard exposure). New rerender_page_sections action re-renders ALL of a page's sections from stored content_data overlaid with FRESH resolved_data (reusing plan_sections' side-effect-free planSection/sourceResolver — route ii), renders via RenderTemplate, emits the exact sections_metadata shape save reads. Slotted into page-rerender as a pre-pass gated by spec.reason (image_landed / section_data_resolved); flag_page_image_rebuild + reconcile_section_data repointed to emit page_rerender-type items (closing their type/key mismatch). NULL content_data on any section → escalate the whole page to needs_page (self-healing one-time full rebuild that backfills content_data). Y-lean render context chosen after confirming templates use only content_data + CSS-var colours. Design alternatives recorded: Option X (no-LLM branch inside the writer) rejected; re-render-affected-section-only rejected in favour of re-render-all.
- **sources:** NOTES(44)#part-A sections (decision trail 2026-06-19→21); RUNBOOK_gamesdesign_index_rebuild(29).md#part-2; HANDOFF_page_pipeline(11).md#5
- **relations:** content_data⊕resolved_data model; re-render vs rebuild; P4.2 (does NOT re-resolve schema-sourced CTAs — that stayed with the writer path).
- **verify-later:** rerender_page_sections_action.go; page-rerender check_rerender_mode wiring; P2.4–P2.7 test outcomes.

<!-- SOURCE: U05_content_quality_linking.md -->
### content_data ⊕ resolved_data persistence model
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19: "UNKNOWN NOW RESOLVED … content_data IS complete enough to re-render from" (RenderComponentAction deliberate merge, per its comment).
- **what:** RenderComponentAction builds a section's content_data as LLM copy (content_from) overlaid with resolved_data (merge_with) — deliberately persisting resolved items/urls/labels alongside the copy, next to rendered_html. This is what makes no-LLM re-rendering possible (render again from stored content + fresh resolution). Corollary schema fact that cost a wrong turn: there is NO page_components.resolved_data column — resolved values live inside content_data.
- **sources:** NOTES(44) 2026-06-19; HANDOFF_2026-06-15(2).md#schema-corrections; running_notes_17(21).md#schema-correction
- **relations:** no-LLM re-render; system-stats key-contract break (content_data keys vs template keys).
- **verify-later:** v3_site_actions.go RenderComponentAction (~L1372); page_components schema.

<!-- SOURCE: U05_content_quality_linking.md -->
### Index stale-rebuild defect (writer output ≠ save input path)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "index rebuild VERIFIED on a real build … sections_metadata array, sm_count=5 … Part 1 result contract working".
- **what:** The unit's opening mystery: index rebuilds completed and git-committed while all five page_components stayed frozen at 06-06. The investigation is a model falsification chain — load, concurrent deploy, claim-lease duration, caller timeout, component locks, and the content-regression guard (writer measured at 33k chars >> the 5760 threshold) were each raised and eliminated — landing on the writer's compiled result being replaced by the size-limit stub before save (the Part-1 result-contract bug). Resolved by the flatten fix; verified end-to-end 06-19 and 06-24 (deployed hero "Your Probability Maths Is Wrong", real hub CTAs).
- **sources:** HANDOFF_2026-06-15_index_stale_rebuild(2).md; NOTES_gamesdesign_silent_norebuild(44).md; running_notes_17(21).md#index-deep-dive
- **relations:** result-contract resolution; silent-completion; save-failure visibility; "git committed ≠ new content" heuristic.
- **verify-later:** orchestration_states 472eed7d/4e0b339a; page_components index timestamps.

<!-- SOURCE: U05_content_quality_linking.md -->
### Save-failure visibility fix (mark_save_failed) + engine error_step ambiguity
- **category:** NEW:page-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** page_build_handler_save_failure_visible.sql delivered 2026-06-15 with "unmet prerequisite (which error_step the engine reads)"; no later doc records applying it.
- **what:** Routes save_sections' error to a new mark_save_failed step (fail_work_item → needs_human_review) instead of complete_error, so a blocked/failed save surfaces instead of laundering into `complete`. Blocked on a real engine unknown: the save_sections step carries error_step in TWO places (step-level and config-level) and it is unconfirmed which the workflow engine honours for routing — "DO NOT GUESS". Companion (also unbuilt): gate deploy_page on sections_saved>0 so a no-write save can't re-commit stale components.
- **sources:** page_build_handler_save_failure_visible.sql; HANDOFF_2026-06-15(2).md#3-bugs; running_notes_17(21).md#FIX-written
- **relations:** silent-completion family; complete_error semantics (Fix B, deferred).
- **verify-later:** whether the SQL was ever applied; chassis engine error_step resolution (step.ErrorStep vs config["error_step"]).

<!-- SOURCE: U05_content_quality_linking.md -->
### Sectionless-page durability stack (2b sibling fallback + S1 check + S2 flag)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** HANDOFF_2026-06-09(2): "Durability code WRITTEN this session (NOT yet deployed)"; running_notes_16(1) carries it as deploy-pending; later docs never record S1 enablement.
- **what:** A planned page reaching build with empty pages.sections silently completed as success ("Content writer skipped — page has no sections defined"). Three-layer durability: 2b — load_page_sections_from_spec gains a final fallback synthesising the layout from a same-role sibling's modal section list (layout skeleton only, WARN-logged, writes pages.sections); S1 — new discovery check check_sectionless_pages flags current-plan pages with empty sections that a sibling can fix and re-triggers to page-build-handler; S2 — check_has_ready_sections ELSE repointed from complete_error to mark_no_sections (needs_human_review). Decisive build fact: pages.sections is the build-read field; site_plan_sections is NOT on the build path (plan hygiene only). Also documented: checkEmptyPageSections is dormant, half-superseded code (wrapper never enabled, wrong handler) — a dedicated check was chosen over reviving it.
- **sources:** running_notes_15(12).md (whole arc); package_module/running_notes_16_adoption_sections.md (same content); HANDOFF_2026-06-09(2).md
- **relations:** Fix A prerequisite; silent-completion; skinner-box case; complete_error semantics.
- **verify-later:** load_page_sections_from_spec_action.go 2b fallback deployed?; completeness-discovery-agent checks array contains "sectionless_pages"?; page-build-handler mark_no_sections step.

<!-- SOURCE: U05_content_quality_linking.md -->
### plan_sections field-source resolution semantics (on_missing, required, defer)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 14p "RESOLVED 2026-06-06" with code-confirmed semantics.
- **what:** The engine semantics governing when a section renders, defers, or drops a field: query.* fields return a non-nil empty slice (never defer, never consult on_missing); on_missing defaults to skip_field; the REQUIRED-field switch has NO skip_field case, so a required field defaulting to skip_field falls to defer — the trap that hid the guides hub (guide-list cta_url required=true + unpopulated spec source deferred the whole section). Fix chosen at the component (required=false) not the engine (the defer-for-safety default is defensible). Related deferral machinery: needs_section_data items for query-resolvable gaps, with reconcile_section_data as the designed loop-closer — registered but STILL UNHOSTED (nothing calls it; query-resolvable items sit at needs_human_review).
- **sources:** running_notes_14(26).md#part-14o-14p; HANDOFF_2026-06-09(2).md#june-02-actions; running_notes_16_content_quality_and_internal_linking(1).md#carried-forward
- **relations:** B4/B5 (query fields + template gates); no-LLM re-render (reuses planSection); component schema contracts.
- **verify-later:** plan_sections_action.go on_missing switches; reconcile_section_data host (still none?); guide-list/blog-listing required flags.

<!-- SOURCE: U05_content_quality_linking.md -->
### sectionHasVisibleContent assembler filter
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-24: "approx_visible_len = 0 … the filter correctly drops it"; "the filter is right".
- **what:** rerender_single_page's getPageSections strips style/script/tags/entities and DROPS any section with ≤10 visible chars (WARN-only). Verified correct for text-empty shells (system-stats), but recognised as a SECOND silent-drop path for interactive content independent of save_page_sections — a low-prose game could be stripped at assembly even after the carry-forward preserves it in the DB. Open question noted: should it share the Part-4 interactivity signal rather than a pure text heuristic (the same text-heuristic blind spot as the regression guard).
- **sources:** NOTES(44) 2026-06-24 system-stats/assembler sessions
- **relations:** interactive clobber; system-stats break; text-heuristic blind spot family.
- **verify-later:** rerender_single_page_action.go sectionHasVisibleContent; game-auto-battler visible_len.

<!-- SOURCE: U05_content_quality_linking.md -->
### A1 — adopted tools/games never deployed a file (parser + status-churn chain)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 10: "A1 VERIFIED CLOSED … all five games committed … tools deploy".
- **what:** No tool/game page produced a deployable file because saveSectionsExtractFromHTML extracted only `<section>` blocks while recreate_tool emits `<div class="tool-page">` → zero page_components → assemblePage returned "" → rerender skipped → no git commit, all while work items read complete. Fixed by the single-fragment fallback (whole fragment as one section, guarded against full documents), coupled with the deployed→needs_rebuild flip removal and deploy-time plan-version stamping. Established the durable read: getPageSections reads page_components, not pages.sections — "has sections" and "has rendered components" are different facts.
- **sources:** running_notes_14(26).md#part-7-10
- **relations:** save_page_sections; interactive clobber (later same-family loss); tool-recreation-handler.
- **verify-later:** saveSectionsExtractFromHTML fallback; deployed repo /tools//games/ trees.

<!-- SOURCE: U05_content_quality_linking.md -->
### UpdatePageStatusAction zero-component deploy guard (Option B)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2 "Option B delivered" + Part 11/12 arc; reaper comment later cites the hardening as in place.
- **what:** A page must never be marked deployed with zero real components: the deployed branch is guarded by pageHasComponents (EXISTS on page_components with non-null component_id + non-empty rendered_html); on zero components it refuses `deployed`, sets needs_rebuild + clears the plan stamp, fail-open on check errors. Keeps build_status honest as evidence for the reaper (the homepage had been 'deployed' with 0 components and no file).
- **sources:** running_notes_14(26).md#part-11-12
- **relations:** evidence-gated reaper; auto-complete false positive; built_from_plan_version stamp.
- **verify-later:** v3_site_actions.go UpdatePageStatusAction deployed branch.

<!-- SOURCE: U05_content_quality_linking.md -->
### Deploy-observability bookkeeping gap
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** NOTES(44) 2026-06-21: "Telemetry gap, non-blocking: page_components.deploy_commit blank and pages.last_built_at NULL though deployed_at is set — deploy step isn't writing those back."
- **what:** The deploy path sets deployed_at but never writes page_components.deploy_commit or pages.last_built_at, and content_hash is empty on investigated pages — so change detection falls back to updated_at + rendered_html length. Folded into a later deploy-observability fix; small but it repeatedly complicated verification.
- **sources:** NOTES(44) 2026-06-21 update; running_notes_17(21).md (content_hash note)
- **relations:** debugging heuristics (git committed ≠ new content); save_page_sections.
- **verify-later:** deploy_page/git_commit write-backs.

<!-- SOURCE: U23_docs_root_vonc.md -->
### render_mode derivation + LLM routing condition (migration 002)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Migration table 2026-06-24: "002 DONE — check_render_mode condition fixed"; deriveRenderMode code deployed 2026-06-24 ~15:00; hero LLM content confirmed on the rebuilt index.
- **what:** `StoreGeneratedComponentAction` originally hardcoded `render_mode='template'` on every component, making the LLM content path permanently unreachable; `deriveRenderMode(inputSchemaJSON)` now returns 'agent' iff any schema field has source='llm'. Separately, page-content-writer's `check_render_mode` condition was reading a never-populated field; migration 002 changed it to `current_section.llm_field_specs != null` (populated by plan_sections from the schema), routing any section with LLM fields to content generation for all sites. Note: render_mode is NOT what routes sections (a later red herring) — llm_field_specs is.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#4-render_mode-hardcoded + #migration-002-outcome + #2026-07-02-~19:20; docs/RUNBOOK_vonc_migrations(14).md#background-migration-002
- **relations:** render_mode sweep (dropped); plan_sections deferral (render_mode red herring)
- **verify-later:** store_generated_component_action.go deriveRenderMode; page-content-writer agent_definition check_render_mode condition

<!-- SOURCE: U23_docs_root_vonc.md -->
### Component-table render_mode sweep (65 components) — dropped migration
- **category:** NEW:page-build-pipeline
- **status-signal:** abandoned
- **status-evidence:** RUNNING_NOTES_vonc.md base vs (1) diff: "Migration 002 (render_mode sweep across 65 components) is DROPPED"; PLAN_vonc_next_steps(1): "The 65-component render_mode update is DROPPED; existing components are fine as-is."
- **what:** The first plan for fixing LLM routing was a DB sweep updating `render_mode` on 65 existing library components. Dropped once it was established that workflow routing reads `llm_field_specs` (set by plan_sections from the schema), not the stored render_mode — so only the agent_definition condition needed fixing and existing component rows were fine as-is. Captures the earliest documented shape of the fix; useful provenance for why component rows still carry historical render_mode values.
- **sources:** docs/RUNNING_NOTES_vonc.md#4 (pre-edit base); docs/PLAN_vonc_next_steps(1).md#p1; docs/RUNBOOK_vonc_migrations(1).md (earlier "Fix render_mode on components" migration heading, dropped from later versions)
- **relations:** render_mode derivation + routing condition (the replacement)
- **verify-later:** none (historical)

<!-- SOURCE: U23_docs_root_vonc.md -->
### plan_sections readiness triage and deferral semantics
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Confirmed in code 2026-07-02 (planSection read); fix validated end-to-end 2026-07-03 (index went 3 → 6 sections after populating cta spec + relaxing illustration field); 016b §9 entry.
- **what:** plan_sections classifies each planned section by resolving its schema fields: source=llm always available; query.*/renderer/static resolve at render time or fall back; any other source runs the resolver. A REQUIRED field whose source doesn't resolve hits the on_missing switch, whose `default:` case DEFERS the section ("default to defer for safety") — and empty on_missing defaults to skip_field, which is not a case in the required switch, so it defers. save_page_sections then persists only the ready set, dropping deferred sections' page instances. Authoring rule: never `required=true` + `on_missing=skip_field`; fix by populating the site data source or degrading the field.
- **sources:** docs/RUNNING_NOTES_vonc(36).md#2026-07-02-~19:20 + #2026-07-02-~19:35; docs/016b_debugging_guide_merged(3).md#regenerated-content-section-is-deferred; docs/RUNBOOK_phase2_provocation_js(29).md#appendix-f
- **relations:** site_specs cta aspect; resolver asset kinds gap; plan-driven rebuild + clobber
- **verify-later:** plan_sections_action.go planSection on_missing switch; save_page_sections_action.go

<!-- SOURCE: U23_docs_root_vonc.md -->
### Plan-driven rebuild + interactive/deferred-section clobber (carry-forward fix)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** 016b: "Part 4... fix WRITTEN (un-deployed)" — Layer 1 interactivity guard + Layer 2 carry-forward in patched save_page_sections_action.go; the 2026-07-02 vonc rebuild demonstrated the drop live (6 planned → 3 saved, brief-explanation instance gone).
- **what:** A needs_page rebuild is PLAN-driven, not pending-driven: load sections from the plan → triage → the writer renders ALL ready planned sections → save_page_sections DELETE+INSERTs the page's components. Sections present in page_components but absent from the plan (interactive tools stored only as rendered_html) or deferred by triage get silently dropped. Fix (written, not deployed): interactivity-aware guard blocking a non-interactive set replacing a deployed interactive one, plus carry-forward of existing interactive sections; three callers to bump (page-build-handler, page-rerender, tool-recreation-handler); plus source_item_id stamping for traceability.
- **sources:** docs/016b_debugging_guide_merged(3).md#open-threads (Part 4 + 2026-06-24 update); docs/RUNNING_NOTES_vonc(36).md#2026-07-01-~13:40 + #2026-07-02-~19:00; docs/PLAN_spark_provocation_pipeline.md#standing-constraints
- **relations:** plan_sections deferral; page_components single-writer (save_page_sections); interactive tool pages stored as rendered_html
- **verify-later:** save_page_sections_action.go (is the guard/carry-forward deployed?); page_component_history.source_item_id

<!-- SOURCE: U23_docs_root_vonc.md -->
### complete_error silent-success family (page build completes having built nothing)
- **category:** NEW:page-build-pipeline
- **status-signal:** partial
- **status-evidence:** Mechanism fully confirmed 2026-07-06 (workflow dump); the defect is live ("an error path implemented as a SUCCESSFUL completion"); preventions listed as backlog item 1 in the HANDOFF, not built.
- **what:** page-build-handler routes zero-ready-sections to a step literally named `complete_error` — a complete_workflow with success_message "Content writer skipped — page has no sections defined" — so builds against a section-less page complete cleanly having done nothing. Ten silent no-ops (33–65s completes) hid a 404 CTA destination for two weeks; a work-item result carrying ONLY `site_record` (healthy: `[sections_saved, deploy_result]`) is the diagnostic signature. Variants: plan row naming a nonexistent component also passes silently. Preventions (aspirational): complete_error fails loudly or raises needs_plan_sections; auditor linked+planned+URL-presence rules; `last_built_at` is never written by build or rerender (dead column — write it or drop it).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed + #2026-07-08; docs/016b_debugging_guide_merged(3).md#page-build-completes-having-built-nothing; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3 + #§9
- **relations:** three section sources; planner ≥1-section invariant; trust-the-artifact doctrine
- **verify-later:** page-build-handler default_config complete_error step; pages.last_built_at writes anywhere in the codebase

<!-- SOURCE: U23_docs_root_vonc.md -->
### load_page_record lookup semantics (name-first, page_id fallback)
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** load_page_record_action.go read in full 2026-07-06: "Lookup priority: page_name (site_id+name) first; page_id only if name empty/bogus (nonPageNames)... returns sections PARSED FROM pages.sections + section_count."
- **what:** The build's page lookup resolves by page_name against `pages.name` first, falling back to page_id only for empty/bogus names, and returns the page's own `sections` jsonb with a count — which is what gates the zero-sections branch. Schema gotcha bundled with it: `pages` has `name` not `page_name`; work-item specs use domain/page_id/filename/page_name.
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-load_page_record-read; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§7
- **relations:** complete_error family; three section sources
- **verify-later:** load_page_record_action.go nonPageNames list

<!-- SOURCE: U23_docs_root_vonc.md -->
### Two re-render paths + assemble-only rerender distinction
- **category:** NEW:page-build-pipeline
- **status-signal:** deployed
- **status-evidence:** Doc-003-derived and header-confirmed 2026-07-09: rerender_single_page "confirmed ASSEMBLE-ONLY" from its own header; light-path escalation rule quoted from 003.
- **what:** Three distinct "rerender" operations that must not be conflated: (1) FULL rebuild — needs_page → page-build-handler → page-content-writer (LLM regenerates copy); (2) LIGHT re-render — rerender_page_sections behind a page_rerender item: re-renders every section from EXISTING content_data via RenderComponentAction, no LLM, and escalates the whole page to a full rebuild when content_data is NULL; (3) ASSEMBLE-ONLY — rerender_single_page (the habitual rerender-*.sh trigger): reassembles stored page_components.rendered_html + stored site_components chrome and deploys; template-only edits will NOT appear through it. Mode-B sections likely have NULL content_data, making the light path escalate — the deciding probe for edit sequencing.
- **sources:** docs/PLAN_provocation-card(3).md#method-corrected; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§4; docs/016b_debugging_guide_merged(3).md#open-threads (Part 2)
- **relations:** sanctioned edit paths; assemble-time visible-content filter; two chrome assembly paths
- **verify-later:** rerender_page_sections_action.go escalation branch; page_rerender item routing

<!-- SOURCE: U14_docs019_runbooks.md -->
### Builder route method — map what exists before building (§B0 census)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "Rule honoured: map what EXISTS against what the problem statement wants BEFORE creating anything. Sources: the 147-row agent_definitions census (2026-07-03)"; §B0 findings enumerated.
- **what:** The builder route's opening method: an inventory matrix of problem-statement capabilities (intake, research, planning, design, content, tools, feeds, infographics, build/deploy, improvement, observability) against the ~147 existing agent types. Findings: every section except infographics has agents; the real defect is ~8 overlapping top-tier "build the site" orchestrators; the per-section content family is already prototyped; genuine gaps are the infographics owner and the success-factor synthesis step. Liveness comes from pump + handler references, not the status column.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B0; docs019/RUNBOOK_builder_route(21).md#B0-findings
- **relations:** three builder generations; work-item relay spine; vertical-exemplar researcher (the gap filled)
- **verify-later:** agent_definitions census queries; duplicate-row Q1

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three coexisting builder generations
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B1 "Three generations coexist: GEN-1 (template era) … GEN-2 (in-memory multipage) … GEN-3 (component/spec/DB era — the LIVE architecture): pageflow-builder v20 (ACTIVE)" (dumps read 2026-07-04).
- **what:** The archaeology of site building: GEN-1 template chains (strategist→architect→writer→html-assembler→site-deployer), GEN-2 in-memory multipage (chief-strategist→content loop→assemble→deployer-agent, no components/specs/review), GEN-3 component/spec/DB (pageflow-builder v20's full inline build; site-work-orchestrator as its queue-native sibling with dynamic per-item handlers and maintenance mode). Explains duplicate deployers (Q3: site-deployer serves GEN-1; deployer-agent GEN-2/3) and frames consolidation.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#open-questions (Q3)
- **relations:** builder census; work-item relay spine (the decision among them)
- **verify-later:** workflow dumps of the nine builders; pageflow-builder v20 definition

<!-- SOURCE: U14_docs019_runbooks.md -->
### The work-item relay spine (baton/hop model)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B3 "DECISION (pre-stated rule fires): the relay reaches page-build-handler natively ⇒ THE SPINE = THE WORK-ITEM RELAY"; MILESTONE 2026-07-06 "first end-to-end domain→deployed site through the relay" (dartsonline.com).
- **what:** The settled build architecture: work moves as a relay of site_work_items batons — each names a handler_agent; the 30s pump claims unclaimed batons and spawns the named agent; the agent does one job, writes findings to site_specs (the site's shared notebook — spec-not-message, the 1.27MB lesson), creates the next baton, stops. Full chain: domain-submitter/adoption → classifier → (vertical research) → strategist → briefing → build-site-planner (emits needs_page/design/imagery/rerender items) → page-build-handler per page → rerender/deploy. Observed extra hops: needs_composition→site-design-planner, needs_design→webdesign-agent, needs_imagery→image-build-handler, needs_rerender→rerender-pages; page items are item_type needs_page. pageflow-builder survives as intake's initial-build convenience.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3; docs019/RUNBOOK_builder_route(21).md#B4 (plain-language explainer, map corrections); docs019/RUNBOOK_builder_route(21).md#milestone
- **relations:** build pump + immune system; builder generations; roadmap scope-decision gap; site quality programme (first output's gaps)
- **verify-later:** load_work_item_actions.go routing; the 37-row dartsonline item chain

<!-- SOURCE: U14_docs019_runbooks.md -->
### Build pump and the queue immune system
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "the scheduler fires build-pipeline-trigger EVERY 30s … The queue's immune system is all ENABLED: claimed-item-timeout (evidence-based auto-complete …), feasibility-recheck, stale-orchestration-reaper, stale-work-item-reaper (48h), work-item-archiver, database-cleanup. FLAG: improvement-sweep is DISABLED."
- **what:** What drives the relay: scheduled build-pipeline-trigger (30s, pre_query gated, concurrency dispatch/8) → build-dispatch-loop → atomic claim → spawn dynamic handler → complete/fail → touch scheduled_tasks. The immune system self-heals the queue (claimed-item-timeout does evidence-based auto-complete, its SQL documenting the gamesdesign false-positive lesson; feasibility-recheck unblocks when handlers appear; reapers and archiver bound staleness). Standing flag: improvement-sweep is disabled platform-wide, so the improvement loop is not running; content-feed-refresh is enabled 6-hourly.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2
- **relations:** work-item relay spine; needs_diagnosis intake (rides the same machinery); site quality LEG 6
- **verify-later:** scheduled_tasks rows (build-pipeline-trigger, improvement-sweep enabled flags); claimed-item-timeout SQL

<!-- SOURCE: U14_docs019_runbooks.md -->
### Two front doors and duplicate classifiers (Q5)
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "Two front doors, two classifiers (overlap)"; queue item 2 "[MAIN] Q5 front-door consolidation — two classifiers, one responsibility" (queued, undecided).
- **what:** Intake exists twice: the queue door (domain-submitter → work-item relay with domain-research-classifier) and intake-orchestrator v3 (HITL: site-classifier → confirm type → questionnaire → briefing-agent → spawn dynamic builder). site-classifier and domain-research-classifier hold the same responsibility; the classifier prompt hardcodes recommended_builder="pageflow-builder"; intake carries orphaned rerender steps. Consolidation direction (deprecate the intake door vs align contracts) is an open decision.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#queue (item 2)
- **relations:** work-item relay spine; site_type taxonomy drift; adoption fidelity inversion
- **verify-later:** intake-orchestrator usage evidence (orchestration_names ILIKE intake); site-classifier workflow

<!-- SOURCE: U14_docs019_runbooks.md -->
### image_tag 'latest' stale-default trap
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "INCIDENT 2026-07-06 — first claim STALLED … THE ONE REAL DIFFERENCE: image_tag='latest' (column default) … the registry's latest is an ANCIENT chassis build … FIX APPLIED … NEW PARKED TRAP (systemic): agent_definitions.image_tag DEFAULTS to 'latest' — every future seeded agent inherits it."
- **what:** Seeded agents inherit image_tag='latest', which points at an ancient pre-architecture chassis build (boots the retired generic.process consumer regardless of env) — the newly seeded researcher stalled on it. Immediate fix: copy image columns from a live donor in every seed. Systemic options parked: repoint/retire `latest`, ALTER the column default, or a New Agent checklist line. Rollback convention is the same lever inverted: revert by repointing image_tag to the prior tag. Same staleness class as the HEAD-pinned index. Follow-up question: does deploy bulk-bump pinned tags (all five tool rows updated at once suggests yes)?
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (incident); docs019/RUNBOOK_builder_route(21).md#queue (item 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#8 (rollback)
- **relations:** stale-corpus class; standing evidence rules (seed hygiene)
- **verify-later:** agent_definitions image_tag column default; whether redeploy-agents bumps rows

<!-- SOURCE: U14_docs019_runbooks.md -->
### Coverage baseline — guides, tools, news, curated top-N on most sites
- **category:** NEW:build-pipeline
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 7 "standing expectation going forward is most sites should carry guides + tools + news + a curated (LLM-picked, non-affiliate) top-N list … the curated-list mechanism, which IS new"; "STANDING EXPECTATION HOME: 001_development_guide … NOT the per-message prompt (decays), NOT the constitution (dev method)."
- **what:** A platform content-coverage policy: most sites should carry guides, tools, news, and a curated non-affiliate top-N list of the vertical's best products/services with outbound links; "pages need not be original to be best-in-class — genuinely useful common content counts". Enforcement points are the strategist/planner prompts (relay-wide-fixes-every-site logic); the curated-list mechanism is the one genuinely new build (reuse candidates: research-agent or the exemplar-researcher crawl pattern feeding a curation step). The mechanism for guides/tools/news EXISTS (gamesdesign, gaswholesalers prove it) — dartsonline's absence is a broken route, not a missing feature.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** F0 guides pilot (the broken route); roadmap gap (same enforcement points); site quality LEG 5
- **verify-later:** 001 guideline amendment; strategist/planner prompt coverage clauses

<!-- SOURCE: U18_sql_for_agents.md -->
### pageflow-builder (component-based site build orchestration)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Still being patched in Phase 0 imagery work 2026-05-05 (107 backs up its row before migration); 026 documents the full live step chain.
- **what:** The central v2-era builder, renamed from multipage-website-builder v3. Spawns planner/content-writer/reviewer/deployer, then: ensure_site_record → call_site_planner → store brief+plan → sync_pages_to_db → populate_nav → asset steps → select_style_collection → set_default_components → render_site_components → get_pages_to_build (filters by build_status) → build_pages_loop (write → review → assemble → deploy per page) → apply_site_design (CSS) → trigger_site_deploy (Cloudflare). The known hazard that sync_pages_to_db can reset page statuses is documented in-file.
- **sources:** 026_pageflow_builder.sql; sql_for_agents_v2/026_pageflow_builder.sql; 107_image_build_handler.sql (backup section)
- **relations:** parallel/legacy path beside site-work-orchestrator and build-dispatch-loop; uses site-planner, page-content-writer, content-reviewer, deployer-agent
- **verify-later:** whether new sites still route through pageflow-builder or only via the work-item pipeline

<!-- SOURCE: U18_sql_for_agents.md -->
### page-content-writer (section-by-section content generation)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Continuously patched from v2 era through 069 (reads site_specs), 107-era imagery direction; 075 gives it idle_timeout 180.
- **what:** Writes one page section-by-section: spawn_research_agent → load_page_components → build_render_context → process_sections_loop (per-section LLM call constrained to that component's `llm_field_specs`) → compile_page. The prompt is a major behavioural contract: official-contact-only rule, internal-link constraint to listed pages, content_direction/imagery_direction from site_specs, admin content briefs, "Recreate Mode" for adopted sites (adapt original page markdown), and an 18-rule anti-fabrication list (no invented people/testimonials/statistics/case studies; "ALWAYS better to be honest and general than specific and fabricated").
- **sources:** 023_page_content_writer_agent.sql; sql_for_agents_v2/023_page_content_writer_agent.sql; 069_blog_posts.sql
- **relations:** called by pageflow-builder, page-build-handler; feeds save_page_sections/page_components; anti-fabrication rules relate to content-governance
- **verify-later:** live prompt_template vs the 023 copies; llm_field_specs source in content_components

<!-- SOURCE: U18_sql_for_agents.md -->
### site-work-orchestrator (unified build/maintenance over site_work_items)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 045 definition; row backed up and patched in Phase 0 imagery (107, 2026-05-05).
- **what:** Orchestrator that builds sites from prioritized site_work_items rows, calling appropriate handler agents per item, "compatible with pageflow-builder's planner and content writer". The first expression of the unified build/maintenance queue idea, later refined into the one-item-at-a-time build-dispatch-loop.
- **sources:** 045_site_work_orchestrator.sql; 107_image_build_handler.sql
- **relations:** site_work_items table; build-dispatch-loop (leaner successor/sibling); discovery agents write into its queue
- **verify-later:** which orchestrator the live triggers use; site_work_items schema

<!-- SOURCE: U18_sql_for_agents.md -->
### Work-item build pipeline: domain-submitter → dispatch loop → handler agents
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 051/052 definitions; 075 sets idle timeouts across the whole handler fleet; 146 still adding items into the same queue in 2026-07.
- **what:** The current architecture. domain-submitter (068) creates a site record + needs_domain_research item from just a domain. build-pipeline-trigger (052) is a 30-min heartbeat: seeds the build queue, finds one site with pending items, fires build-dispatch-loop (051), which loads the highest-priority claimable item, claims it, spawns+calls the handler agent, marks complete, and if items remain spawns a FRESH dispatch loop (separate orchestration, clean logs). Item chain for a new site: needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → needs_content_page (per page) → images → needs_rerender. Concurrency safety via claim_work_item; health-gating via ai_endpoint_health before claiming.
- **sources:** 051_build_dispatch_loop.sql; 052_build_pipeline_trigger.sql; 068_domain_submitter_agent.sql; 085_ai_endpoint_health_checker.sql
- **relations:** every handler agent below; scheduler-and-tasks (CronJob trigger); replaces intake-orchestrator
- **verify-later:** LoadWorkItemsAction first_item patch; claim semantics; current item_type → handler_agent routing table

<!-- SOURCE: U18_sql_for_agents.md -->
### page-build-handler (content-page handler with section planning and validation gates)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 065 documents the evolved workflow with plan_sections/validate_content and error paths; 070 notes empty_sections handler switched to it.
- **what:** Wrapper solving "specialist vs handler": page-content-writer generates but doesn't persist, so this handler loads page + specs, plan_sections resolves data sources per section (creating deferred items when sections aren't ready), calls the writer, validate_content checks placeholders/templates/cross-site contamination (blockers → mark_needs_review), then save_page_sections, update_page_status, and deploys via page-rerender. Earlier version (055) was simpler (no plan/validate steps, no deploy).
- **sources:** 065_page_build_handler_wrapper.sql; 055_page_build_handler.sql; 070_blog_content_planner.sql
- **relations:** page-content-writer, page-rerender; content_rewrite items route here; needs_new_component items from plan_sections
- **verify-later:** plan_sections + validate_page_content actions; deferred-item creation

<!-- SOURCE: U18_sql_for_agents.md -->
### page-rebuild (rebuild pages without re-planning)
- **category:** NEW:build-pipeline
- **status-signal:** unknown
- **status-evidence:** 039 full definition with detailed reuse/skip lists; no later references found in this unit.
- **what:** Rebuilds specific pages (build_status='needs_rebuild') on an existing site loading all context from DB given a domain, explicitly skipping planner, sync_pages_to_db, asset generation, component rendering, CSS and nav (all already done) while reusing the standard build-loop agents. Documents design principles: agent owns its domain; spawnable not standalone; reuse before creating; complexity in Go.
- **sources:** 039_page_rebuild_agent.sql
- **relations:** pageflow-builder (same loop, different input_mapping via rebuild_context); load_site_for_rebuild action
- **verify-later:** whether page-rebuild survived the dispatch-loop refactor

<!-- SOURCE: U14_docs019_runbooks.md -->
### Builder route method — map what exists before building (§B0 census)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) "Rule honoured: map what EXISTS against what the problem statement wants BEFORE creating anything. Sources: the 147-row agent_definitions census (2026-07-03)"; §B0 findings enumerated.
- **what:** The builder route's opening method: an inventory matrix of problem-statement capabilities (intake, research, planning, design, content, tools, feeds, infographics, build/deploy, improvement, observability) against the ~147 existing agent types. Findings: every section except infographics has agents; the real defect is ~8 overlapping top-tier "build the site" orchestrators; the per-section content family is already prototyped; genuine gaps are the infographics owner and the success-factor synthesis step. Liveness comes from pump + handler references, not the status column.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B0; docs019/RUNBOOK_builder_route(21).md#B0-findings
- **relations:** three builder generations; work-item relay spine; vertical-exemplar researcher (the gap filled)
- **verify-later:** agent_definitions census queries; duplicate-row Q1

<!-- SOURCE: U14_docs019_runbooks.md -->
### Three coexisting builder generations
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B1 "Three generations coexist: GEN-1 (template era) … GEN-2 (in-memory multipage) … GEN-3 (component/spec/DB era — the LIVE architecture): pageflow-builder v20 (ACTIVE)" (dumps read 2026-07-04).
- **what:** The archaeology of site building: GEN-1 template chains (strategist→architect→writer→html-assembler→site-deployer), GEN-2 in-memory multipage (chief-strategist→content loop→assemble→deployer-agent, no components/specs/review), GEN-3 component/spec/DB (pageflow-builder v20's full inline build; site-work-orchestrator as its queue-native sibling with dynamic per-item handlers and maintenance mode). Explains duplicate deployers (Q3: site-deployer serves GEN-1; deployer-agent GEN-2/3) and frames consolidation.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#open-questions (Q3)
- **relations:** builder census; work-item relay spine (the decision among them)
- **verify-later:** workflow dumps of the nine builders; pageflow-builder v20 definition

<!-- SOURCE: U14_docs019_runbooks.md -->
### The work-item relay spine (baton/hop model)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B3 "DECISION (pre-stated rule fires): the relay reaches page-build-handler natively ⇒ THE SPINE = THE WORK-ITEM RELAY"; MILESTONE 2026-07-06 "first end-to-end domain→deployed site through the relay" (dartsonline.com).
- **what:** The settled build architecture: work moves as a relay of site_work_items batons — each names a handler_agent; the 30s pump claims unclaimed batons and spawns the named agent; the agent does one job, writes findings to site_specs (the site's shared notebook — spec-not-message, the 1.27MB lesson), creates the next baton, stops. Full chain: domain-submitter/adoption → classifier → (vertical research) → strategist → briefing → build-site-planner (emits needs_page/design/imagery/rerender items) → page-build-handler per page → rerender/deploy. Observed extra hops: needs_composition→site-design-planner, needs_design→webdesign-agent, needs_imagery→image-build-handler, needs_rerender→rerender-pages; page items are item_type needs_page. pageflow-builder survives as intake's initial-build convenience.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B3; docs019/RUNBOOK_builder_route(21).md#B4 (plain-language explainer, map corrections); docs019/RUNBOOK_builder_route(21).md#milestone
- **relations:** build pump + immune system; builder generations; roadmap scope-decision gap; site quality programme (first output's gaps)
- **verify-later:** load_work_item_actions.go routing; the 37-row dartsonline item chain

<!-- SOURCE: U14_docs019_runbooks.md -->
### Build pump and the queue immune system
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** builder_route(21) §B2 "the scheduler fires build-pipeline-trigger EVERY 30s … The queue's immune system is all ENABLED: claimed-item-timeout (evidence-based auto-complete …), feasibility-recheck, stale-orchestration-reaper, stale-work-item-reaper (48h), work-item-archiver, database-cleanup. FLAG: improvement-sweep is DISABLED."
- **what:** What drives the relay: scheduled build-pipeline-trigger (30s, pre_query gated, concurrency dispatch/8) → build-dispatch-loop → atomic claim → spawn dynamic handler → complete/fail → touch scheduled_tasks. The immune system self-heals the queue (claimed-item-timeout does evidence-based auto-complete, its SQL documenting the gamesdesign false-positive lesson; feasibility-recheck unblocks when handlers appear; reapers and archiver bound staleness). Standing flag: improvement-sweep is disabled platform-wide, so the improvement loop is not running; content-feed-refresh is enabled 6-hourly.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B2
- **relations:** work-item relay spine; needs_diagnosis intake (rides the same machinery); site quality LEG 6
- **verify-later:** scheduled_tasks rows (build-pipeline-trigger, improvement-sweep enabled flags); claimed-item-timeout SQL

<!-- SOURCE: U14_docs019_runbooks.md -->
### Two front doors and duplicate classifiers (Q5)
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "Two front doors, two classifiers (overlap)"; queue item 2 "[MAIN] Q5 front-door consolidation — two classifiers, one responsibility" (queued, undecided).
- **what:** Intake exists twice: the queue door (domain-submitter → work-item relay with domain-research-classifier) and intake-orchestrator v3 (HITL: site-classifier → confirm type → questionnaire → briefing-agent → spawn dynamic builder). site-classifier and domain-research-classifier hold the same responsibility; the classifier prompt hardcodes recommended_builder="pageflow-builder"; intake carries orphaned rerender steps. Consolidation direction (deprecate the intake door vs align contracts) is an open decision.
- **sources:** docs019/RUNBOOK_builder_route(21).md#B1; docs019/RUNBOOK_builder_route(21).md#queue (item 2)
- **relations:** work-item relay spine; site_type taxonomy drift; adoption fidelity inversion
- **verify-later:** intake-orchestrator usage evidence (orchestration_names ILIKE intake); site-classifier workflow

<!-- SOURCE: U14_docs019_runbooks.md -->
### image_tag 'latest' stale-default trap
- **category:** NEW:build-pipeline
- **status-signal:** partial
- **status-evidence:** builder_route(21) "INCIDENT 2026-07-06 — first claim STALLED … THE ONE REAL DIFFERENCE: image_tag='latest' (column default) … the registry's latest is an ANCIENT chassis build … FIX APPLIED … NEW PARKED TRAP (systemic): agent_definitions.image_tag DEFAULTS to 'latest' — every future seeded agent inherits it."
- **what:** Seeded agents inherit image_tag='latest', which points at an ancient pre-architecture chassis build (boots the retired generic.process consumer regardless of env) — the newly seeded researcher stalled on it. Immediate fix: copy image columns from a live donor in every seed. Systemic options parked: repoint/retire `latest`, ALTER the column default, or a New Agent checklist line. Rollback convention is the same lever inverted: revert by repointing image_tag to the prior tag. Same staleness class as the HEAD-pinned index. Follow-up question: does deploy bulk-bump pinned tags (all five tool rows updated at once suggests yes)?
- **sources:** docs019/RUNBOOK_builder_route(21).md#B4 (incident); docs019/RUNBOOK_builder_route(21).md#queue (item 1); docs019/RUNBOOK_gamesdesign_index_rebuild.md#8 (rollback)
- **relations:** stale-corpus class; standing evidence rules (seed hygiene)
- **verify-later:** agent_definitions image_tag column default; whether redeploy-agents bumps rows

<!-- SOURCE: U14_docs019_runbooks.md -->
### Coverage baseline — guides, tools, news, curated top-N on most sites
- **category:** NEW:build-pipeline
- **status-signal:** aspirational
- **status-evidence:** builder_route(21) queue item 7 "standing expectation going forward is most sites should carry guides + tools + news + a curated (LLM-picked, non-affiliate) top-N list … the curated-list mechanism, which IS new"; "STANDING EXPECTATION HOME: 001_development_guide … NOT the per-message prompt (decays), NOT the constitution (dev method)."
- **what:** A platform content-coverage policy: most sites should carry guides, tools, news, and a curated non-affiliate top-N list of the vertical's best products/services with outbound links; "pages need not be original to be best-in-class — genuinely useful common content counts". Enforcement points are the strategist/planner prompts (relay-wide-fixes-every-site logic); the curated-list mechanism is the one genuinely new build (reuse candidates: research-agent or the exemplar-researcher crawl pattern feeding a curation step). The mechanism for guides/tools/news EXISTS (gamesdesign, gaswholesalers prove it) — dartsonline's absence is a broken route, not a missing feature.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 7)
- **relations:** F0 guides pilot (the broken route); roadmap gap (same enforcement points); site quality LEG 5
- **verify-later:** 001 guideline amendment; strategist/planner prompt coverage clauses

<!-- SOURCE: U18_sql_for_agents.md -->
### pageflow-builder (component-based site build orchestration)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Still being patched in Phase 0 imagery work 2026-05-05 (107 backs up its row before migration); 026 documents the full live step chain.
- **what:** The central v2-era builder, renamed from multipage-website-builder v3. Spawns planner/content-writer/reviewer/deployer, then: ensure_site_record → call_site_planner → store brief+plan → sync_pages_to_db → populate_nav → asset steps → select_style_collection → set_default_components → render_site_components → get_pages_to_build (filters by build_status) → build_pages_loop (write → review → assemble → deploy per page) → apply_site_design (CSS) → trigger_site_deploy (Cloudflare). The known hazard that sync_pages_to_db can reset page statuses is documented in-file.
- **sources:** 026_pageflow_builder.sql; sql_for_agents_v2/026_pageflow_builder.sql; 107_image_build_handler.sql (backup section)
- **relations:** parallel/legacy path beside site-work-orchestrator and build-dispatch-loop; uses site-planner, page-content-writer, content-reviewer, deployer-agent
- **verify-later:** whether new sites still route through pageflow-builder or only via the work-item pipeline

<!-- SOURCE: U18_sql_for_agents.md -->
### page-content-writer (section-by-section content generation)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** Continuously patched from v2 era through 069 (reads site_specs), 107-era imagery direction; 075 gives it idle_timeout 180.
- **what:** Writes one page section-by-section: spawn_research_agent → load_page_components → build_render_context → process_sections_loop (per-section LLM call constrained to that component's `llm_field_specs`) → compile_page. The prompt is a major behavioural contract: official-contact-only rule, internal-link constraint to listed pages, content_direction/imagery_direction from site_specs, admin content briefs, "Recreate Mode" for adopted sites (adapt original page markdown), and an 18-rule anti-fabrication list (no invented people/testimonials/statistics/case studies; "ALWAYS better to be honest and general than specific and fabricated").
- **sources:** 023_page_content_writer_agent.sql; sql_for_agents_v2/023_page_content_writer_agent.sql; 069_blog_posts.sql
- **relations:** called by pageflow-builder, page-build-handler; feeds save_page_sections/page_components; anti-fabrication rules relate to content-governance
- **verify-later:** live prompt_template vs the 023 copies; llm_field_specs source in content_components

<!-- SOURCE: U18_sql_for_agents.md -->
### site-work-orchestrator (unified build/maintenance over site_work_items)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 045 definition; row backed up and patched in Phase 0 imagery (107, 2026-05-05).
- **what:** Orchestrator that builds sites from prioritized site_work_items rows, calling appropriate handler agents per item, "compatible with pageflow-builder's planner and content writer". The first expression of the unified build/maintenance queue idea, later refined into the one-item-at-a-time build-dispatch-loop.
- **sources:** 045_site_work_orchestrator.sql; 107_image_build_handler.sql
- **relations:** site_work_items table; build-dispatch-loop (leaner successor/sibling); discovery agents write into its queue
- **verify-later:** which orchestrator the live triggers use; site_work_items schema

<!-- SOURCE: U18_sql_for_agents.md -->
### Work-item build pipeline: domain-submitter → dispatch loop → handler agents
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 051/052 definitions; 075 sets idle timeouts across the whole handler fleet; 146 still adding items into the same queue in 2026-07.
- **what:** The current architecture. domain-submitter (068) creates a site record + needs_domain_research item from just a domain. build-pipeline-trigger (052) is a 30-min heartbeat: seeds the build queue, finds one site with pending items, fires build-dispatch-loop (051), which loads the highest-priority claimable item, claims it, spawns+calls the handler agent, marks complete, and if items remain spawns a FRESH dispatch loop (separate orchestration, clean logs). Item chain for a new site: needs_domain_research → needs_strategy → needs_briefing → needs_site_plan → needs_content_page (per page) → images → needs_rerender. Concurrency safety via claim_work_item; health-gating via ai_endpoint_health before claiming.
- **sources:** 051_build_dispatch_loop.sql; 052_build_pipeline_trigger.sql; 068_domain_submitter_agent.sql; 085_ai_endpoint_health_checker.sql
- **relations:** every handler agent below; scheduler-and-tasks (CronJob trigger); replaces intake-orchestrator
- **verify-later:** LoadWorkItemsAction first_item patch; claim semantics; current item_type → handler_agent routing table

<!-- SOURCE: U18_sql_for_agents.md -->
### page-build-handler (content-page handler with section planning and validation gates)
- **category:** NEW:build-pipeline
- **status-signal:** deployed
- **status-evidence:** 065 documents the evolved workflow with plan_sections/validate_content and error paths; 070 notes empty_sections handler switched to it.
- **what:** Wrapper solving "specialist vs handler": page-content-writer generates but doesn't persist, so this handler loads page + specs, plan_sections resolves data sources per section (creating deferred items when sections aren't ready), calls the writer, validate_content checks placeholders/templates/cross-site contamination (blockers → mark_needs_review), then save_page_sections, update_page_status, and deploys via page-rerender. Earlier version (055) was simpler (no plan/validate steps, no deploy).
- **sources:** 065_page_build_handler_wrapper.sql; 055_page_build_handler.sql; 070_blog_content_planner.sql
- **relations:** page-content-writer, page-rerender; content_rewrite items route here; needs_new_component items from plan_sections
- **verify-later:** plan_sections + validate_page_content actions; deferred-item creation

<!-- SOURCE: U18_sql_for_agents.md -->
### page-rebuild (rebuild pages without re-planning)
- **category:** NEW:build-pipeline
- **status-signal:** unknown
- **status-evidence:** 039 full definition with detailed reuse/skip lists; no later references found in this unit.
- **what:** Rebuilds specific pages (build_status='needs_rebuild') on an existing site loading all context from DB given a domain, explicitly skipping planner, sync_pages_to_db, asset generation, component rendering, CSS and nav (all already done) while reusing the standard build-loop agents. Documents design principles: agent owns its domain; spawnable not standalone; reuse before creating; complexity in Go.
- **sources:** 039_page_rebuild_agent.sql
- **relations:** pageflow-builder (same loop, different input_mapping via rebuild_context); load_site_for_rebuild action
- **verify-later:** whether page-rebuild survived the dispatch-loop refactor

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F3 scoped reason-stamped rerender (dependent-page scoping + reason propagation)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "F3 PROVEN END-TO-END (scoping to exactly the five, reason propagation, gate, section-rerender execution)" (NOTES §9v); fleet on v1.0.1088 (§9r).
- **what:** Three coupled changes making a component regen's triggered re-render actually repair sections, scoped to the blast radius: F3a — create_rerender_items accepts `reason`+`component_id`; when reason ∈ {section_data_resolved, image_landed} AND component_id set, it queries the component's dependent pages (page_components JOIN pages) and creates reason-stamped `page_rerender` items only for those; no signals → unchanged assemble-only all-pages behaviour. F3b — store re-adds `reason: section_data_resolved` to its needs_rerender spec. F3c — rerender-pages step config maps `reason`/`component_id` from `input_data.spec`. Either half alone degrades safely to assemble-only. Accepted residual: rerender_page_sections re-renders ALL sections of each dependent page (documented blast radius that later stamped the gauntlet onto robot-hands).
- **sources:** NOTES(43).md §9b, §9p–§9r, §9v; RUNBOOK(49).md Part A; w4b_03_read_rerender_config.sql
- **relations:** assemble-only vs section re-render; F8 step 4 reused it; F6 (dedup/counter flaws in the same action).
- **verify-later:** create_rerender_items_action.go (InputSpec reason/component_id; dependent scoping; rows.Close before INSERT loop); rerender-pages default_config create_rerender_items step (5 keys).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Assemble-only vs section re-render distinction (the `reason` gate)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "a bare needs_rerender is assemble-only and just re-ships the empty shell" (NOTES §6); gate `reason ∈ {image_landed, section_data_resolved}` confirmed from rerender_page_sections source (§2 Correction 3).
- **what:** Two fundamentally different re-render depths: an assemble-only pass re-assembles/re-ships existing `rendered_html` (cannot repair content), while a section re-render (`rerender_page_sections`, gated on reason image_landed/section_data_resolved) regenerates each section's HTML from stored `content_data` against the current template, no LLM. A reason dropped anywhere along the chain (as rerender-pages originally did) silently downgrades to assemble-only — an inert "fix" that was caught only by checking the consuming step's config. Central operational lesson of the thread.
- **sources:** NOTES(43).md §2 Correction 3, §6, §9a; PLAN(1).md Phase 4; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F3 (makes the reason survive); rerender fossilisation (site-level analogue); carry-forward path.
- **verify-later:** rerender_page_sections_action.go reason gate; page-rerender workflow's routing on spec.reason.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Carry-forward path and the carry fingerprint
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "per-page save-style row updates + identical bytes = the rerender_page_sections CARRY path fingerprint" (NOTES §9u), confirmed by data §9v.
- **what:** In rerender_page_sections, a section that fails `planSection` readiness (unresolvable required field, missing component, empty template) is carried: `carryStoredSection` re-emits the stored HTML, save_sections writes it back with a fresh per-page `updated_at` but identical bytes. Protective against shipping worse output, but it re-fossilises stale/empty renders forever when readiness is permanently blocked — and its diagnostic signature (fresh distinct timestamps + one shared md5) is how the recovery stall was pinned to cta_url readiness rather than the rerender chain.
- **sources:** NOTES(43).md §9u, §9v; BUNDLE(3).md §3 (rerender partly protective)
- **relations:** section readiness model (the gate it consults); F5 (the added-required-field cause); recovery playbook.
- **verify-later:** rerender_page_sections_action.go: planSection, carryStoredSection, NULL pre-check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Auto-escalation: empty content_data → needs_page writer rebuild
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "Auto-escalations fired as designed: page-rerender created needs_page for matchmatrix (17:31) and gripper-selection-guide (17:37); writer rebuilt gripper-selection-guide 18:04" (NOTES §9aw).
- **what:** rerender_page_sections' NULL pre-check escalates a whole page to a `needs_page` item (handler page-build-handler, spec `{reason: content_data_backfill, page_name}`) when any section lacks `content_data` — routing un-re-renderable pages to a full writer rebuild instead of carrying garbage. Deliberately exploited during F8 recovery both as the rebuild mechanism and as the source of a correctly-shaped needs_page item to clone (never guess a spec).
- **sources:** NOTES(43).md §9ap, §9aw, §9bd; RUNBOOK(49).md Part C Steps 8–9
- **relations:** carry-forward path (its sibling branch); work-item spec-cloning discipline; matchmatrix planning-data gap (an escalation that no-ops).
- **verify-later:** rerender_page_sections NULL pre-check/escalate branch; page-build-handler content_data_backfill flow.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Rerender fossilisation (reassembly re-ships stale renders; template changes need full rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "deployed hero consumes legacy var(--accent-color…) — sections are stale old-template renders… A full page-build-handler rebuild is required; needs_rerender would re-fossilise" (RUNBOOK_scheme_to_components(18) Check 4a, evidenced from deployed HTML).
- **what:** The rerender handler reassembles stored `page_components.rendered_html` and injects `site_components.rendered_html` — it does not re-render component templates — so sites can live indefinitely on reassemblies of early renders while the library advances (idea.uk served renders of long-inactive components). Template/library changes only reach pages through a section re-render or a full page-build-handler rebuild; chrome only through render_site_components. Settled a documented 016-vs-026 doc contradiction by direct evidence.
- **sources:** RUNBOOK_scheme_to_components(18).md Check 3/4 RESULTS; running_notes_scheme_to_components(22).md Sh (migration route), So
- **relations:** assemble-only vs section re-render; chrome refresh gating; W6 rebuild sequencing.
- **verify-later:** rerender_single_page_action.go; which paths run RenderComponent vs re-ship stored HTML.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Chrome refresh gating (render_site_components, force_rerender, repoint-before-rerender)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "renderAndStoreSiteComponent joins the PINNED site_components.component_id with no is_active filter, and without force_rerender SKIPS non-empty slots" (RUNBOOK_scheme(18) W4b, code fact); repoint executed UPDATE 1 ×2.
- **what:** Site chrome (header/footer/head) is rendered and stored by render_site_components, a path separate from page builds (page-build-handler is not among its six invoking agents — a rebuild never refreshes chrome). Its join honours the pinned component_id with no is_active filter and skips non-empty slots unless `force_rerender: true` (only rerender-pages v6 passes it, gated on `spec.refresh_site_components`). Operational consequence: repoint the pinned rows to the fixed/active components BEFORE forcing the re-render, and refresh stored chrome before a rebuild so later automatic rerenders can't re-inject stale dark chrome.
- **sources:** RUNBOOK_scheme_to_components(18).md W4b + STEP-1 RESULTS; running_notes(22).md Sy–Ta; w4b_03_read_rerender_config.sql
- **relations:** chrome linkage tangle; rerender fossilisation; F3 (same rerender-pages agent).
- **verify-later:** render_site_components_action.go:345–430; rerender-pages v6 check_refresh_components gate.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### page-build-handler writes only planned sections (sections=0 → silent no-op rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "matchmatrix sections_listed=0 vs siblings 9/5 ⇒ the page-build-handler iterates the planned section list and had nothing to write — planning-data gap on a legacy page — PARKED" (NOTES §9bh).
- **what:** The writer rebuild flow iterates the page's planned section list; a legacy page with an empty `pages.sections` plan completes its needs_page item without writing anything — a silent no-op rebuild indistinguishable from success at the item level (double no-op observed before diagnosis; page_type hypothesis falsified first). Remediation options (planner reconcile/adopt, or retire the page) parked as hygiene.
- **sources:** NOTES(43).md §9bf–§9bh; RUNBOOK(49).md Step 12(a)
- **relations:** auto-escalation; loose status semantics; site-plan-and-reconciler (planner as another item producer, §9at).
- **verify-later:** page-build-handler section iteration; matchmatrix pages.sections.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### F3 scoped reason-stamped rerender (dependent-page scoping + reason propagation)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "F3 PROVEN END-TO-END (scoping to exactly the five, reason propagation, gate, section-rerender execution)" (NOTES §9v); fleet on v1.0.1088 (§9r).
- **what:** Three coupled changes making a component regen's triggered re-render actually repair sections, scoped to the blast radius: F3a — create_rerender_items accepts `reason`+`component_id`; when reason ∈ {section_data_resolved, image_landed} AND component_id set, it queries the component's dependent pages (page_components JOIN pages) and creates reason-stamped `page_rerender` items only for those; no signals → unchanged assemble-only all-pages behaviour. F3b — store re-adds `reason: section_data_resolved` to its needs_rerender spec. F3c — rerender-pages step config maps `reason`/`component_id` from `input_data.spec`. Either half alone degrades safely to assemble-only. Accepted residual: rerender_page_sections re-renders ALL sections of each dependent page (documented blast radius that later stamped the gauntlet onto robot-hands).
- **sources:** NOTES(43).md §9b, §9p–§9r, §9v; RUNBOOK(49).md Part A; w4b_03_read_rerender_config.sql
- **relations:** assemble-only vs section re-render; F8 step 4 reused it; F6 (dedup/counter flaws in the same action).
- **verify-later:** create_rerender_items_action.go (InputSpec reason/component_id; dependent scoping; rows.Close before INSERT loop); rerender-pages default_config create_rerender_items step (5 keys).

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Assemble-only vs section re-render distinction (the `reason` gate)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "a bare needs_rerender is assemble-only and just re-ships the empty shell" (NOTES §6); gate `reason ∈ {image_landed, section_data_resolved}` confirmed from rerender_page_sections source (§2 Correction 3).
- **what:** Two fundamentally different re-render depths: an assemble-only pass re-assembles/re-ships existing `rendered_html` (cannot repair content), while a section re-render (`rerender_page_sections`, gated on reason image_landed/section_data_resolved) regenerates each section's HTML from stored `content_data` against the current template, no LLM. A reason dropped anywhere along the chain (as rerender-pages originally did) silently downgrades to assemble-only — an inert "fix" that was caught only by checking the consuming step's config. Central operational lesson of the thread.
- **sources:** NOTES(43).md §2 Correction 3, §6, §9a; PLAN(1).md Phase 4; RUNBOOK_pre_cleanup_backup.md §The problem
- **relations:** F3 (makes the reason survive); rerender fossilisation (site-level analogue); carry-forward path.
- **verify-later:** rerender_page_sections_action.go reason gate; page-rerender workflow's routing on spec.reason.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Carry-forward path and the carry fingerprint
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "per-page save-style row updates + identical bytes = the rerender_page_sections CARRY path fingerprint" (NOTES §9u), confirmed by data §9v.
- **what:** In rerender_page_sections, a section that fails `planSection` readiness (unresolvable required field, missing component, empty template) is carried: `carryStoredSection` re-emits the stored HTML, save_sections writes it back with a fresh per-page `updated_at` but identical bytes. Protective against shipping worse output, but it re-fossilises stale/empty renders forever when readiness is permanently blocked — and its diagnostic signature (fresh distinct timestamps + one shared md5) is how the recovery stall was pinned to cta_url readiness rather than the rerender chain.
- **sources:** NOTES(43).md §9u, §9v; BUNDLE(3).md §3 (rerender partly protective)
- **relations:** section readiness model (the gate it consults); F5 (the added-required-field cause); recovery playbook.
- **verify-later:** rerender_page_sections_action.go: planSection, carryStoredSection, NULL pre-check.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Auto-escalation: empty content_data → needs_page writer rebuild
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "Auto-escalations fired as designed: page-rerender created needs_page for matchmatrix (17:31) and gripper-selection-guide (17:37); writer rebuilt gripper-selection-guide 18:04" (NOTES §9aw).
- **what:** rerender_page_sections' NULL pre-check escalates a whole page to a `needs_page` item (handler page-build-handler, spec `{reason: content_data_backfill, page_name}`) when any section lacks `content_data` — routing un-re-renderable pages to a full writer rebuild instead of carrying garbage. Deliberately exploited during F8 recovery both as the rebuild mechanism and as the source of a correctly-shaped needs_page item to clone (never guess a spec).
- **sources:** NOTES(43).md §9ap, §9aw, §9bd; RUNBOOK(49).md Part C Steps 8–9
- **relations:** carry-forward path (its sibling branch); work-item spec-cloning discipline; matchmatrix planning-data gap (an escalation that no-ops).
- **verify-later:** rerender_page_sections NULL pre-check/escalate branch; page-build-handler content_data_backfill flow.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Rerender fossilisation (reassembly re-ships stale renders; template changes need full rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "deployed hero consumes legacy var(--accent-color…) — sections are stale old-template renders… A full page-build-handler rebuild is required; needs_rerender would re-fossilise" (RUNBOOK_scheme_to_components(18) Check 4a, evidenced from deployed HTML).
- **what:** The rerender handler reassembles stored `page_components.rendered_html` and injects `site_components.rendered_html` — it does not re-render component templates — so sites can live indefinitely on reassemblies of early renders while the library advances (idea.uk served renders of long-inactive components). Template/library changes only reach pages through a section re-render or a full page-build-handler rebuild; chrome only through render_site_components. Settled a documented 016-vs-026 doc contradiction by direct evidence.
- **sources:** RUNBOOK_scheme_to_components(18).md Check 3/4 RESULTS; running_notes_scheme_to_components(22).md Sh (migration route), So
- **relations:** assemble-only vs section re-render; chrome refresh gating; W6 rebuild sequencing.
- **verify-later:** rerender_single_page_action.go; which paths run RenderComponent vs re-ship stored HTML.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### Chrome refresh gating (render_site_components, force_rerender, repoint-before-rerender)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "renderAndStoreSiteComponent joins the PINNED site_components.component_id with no is_active filter, and without force_rerender SKIPS non-empty slots" (RUNBOOK_scheme(18) W4b, code fact); repoint executed UPDATE 1 ×2.
- **what:** Site chrome (header/footer/head) is rendered and stored by render_site_components, a path separate from page builds (page-build-handler is not among its six invoking agents — a rebuild never refreshes chrome). Its join honours the pinned component_id with no is_active filter and skips non-empty slots unless `force_rerender: true` (only rerender-pages v6 passes it, gated on `spec.refresh_site_components`). Operational consequence: repoint the pinned rows to the fixed/active components BEFORE forcing the re-render, and refresh stored chrome before a rebuild so later automatic rerenders can't re-inject stale dark chrome.
- **sources:** RUNBOOK_scheme_to_components(18).md W4b + STEP-1 RESULTS; running_notes(22).md Sy–Ta; w4b_03_read_rerender_config.sql
- **relations:** chrome linkage tangle; rerender fossilisation; F3 (same rerender-pages agent).
- **verify-later:** render_site_components_action.go:345–430; rerender-pages v6 check_refresh_components gate.

<!-- SOURCE: U07_content_quality2_overwrite.md -->
### page-build-handler writes only planned sections (sections=0 → silent no-op rebuild)
- **category:** NEW:rebuild-cascade
- **status-signal:** deployed
- **status-evidence:** "matchmatrix sections_listed=0 vs siblings 9/5 ⇒ the page-build-handler iterates the planned section list and had nothing to write — planning-data gap on a legacy page — PARKED" (NOTES §9bh).
- **what:** The writer rebuild flow iterates the page's planned section list; a legacy page with an empty `pages.sections` plan completes its needs_page item without writing anything — a silent no-op rebuild indistinguishable from success at the item level (double no-op observed before diagnosis; page_type hypothesis falsified first). Remediation options (planner reconcile/adopt, or retire the page) parked as hygiene.
- **sources:** NOTES(43).md §9bf–§9bh; RUNBOOK(49).md Step 12(a)
- **relations:** auto-escalation; loose status semantics; site-plan-and-reconciler (planner as another item producer, §9at).
- **verify-later:** page-build-handler section iteration; matchmatrix pages.sections.

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Work-item state machine (detected → triaged → claimed → complete/failed)
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Phase 2G + 2H verified end-to-end at scale. Seven hero items … processed through the full chain … without manual intervention" (2026-05-15)
- **what:** `detected` is a valid intermediate state: discovery emits at detected; design-audit-agent's `triage_detected_items` step (registry.go:722) promotes to triaged; dispatch claims only triaged/approved (partial indexes idx_swi_handler / idx_swi_site_pending); handlers mark complete/failed (mark_work_item_complete / mark_work_item_failed steps). There is NO automated coupling between discovery and audit — items sit in detected until an audit runs. Admin-created items insert directly at triaged.
- **sources:** FOCUS_dispatch_diagnostic(4).md#TL;DR, #Evidence-trail; HANDOFF-pipeline-triage-april-2026.md
- **relations:** dispatch chain; auto-triage open question; two-strike rule; silent completion
- **verify-later:** registry.go triage_detected_items; site_work_items partial indexes; design-audit-agent workflow

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatch chain: build-pipeline-trigger → find_dispatchable_site → build-dispatch-loop
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "researched in depth this session" (2026-05-15) with the actual selection SQL quoted; scheduled_tasks row build-pipeline-trigger every 30s
- **what:** The scheduler fires build-pipeline-trigger, whose find_dispatchable_site step picks ONE site per tick (DISTINCT ON with no outer ORDER BY — effectively arbitrary among eligible sites) and spawns build-dispatch-loop scoped to it, which loads up to 5 items (pipeline='build') and claims/spawns handlers. Throughput cap ~5 items per site per 30s, one site at a time. build-pipeline-trigger doesn't write orchestration_states, making its decisions untraceable.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** NOT EXISTS blocker; Bug 3 site-targeting; fairness ORDER BY improvement
- **verify-later:** scheduled_tasks 'build-pipeline-trigger' row; build-pipeline-trigger / build-dispatch-loop agent definitions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### NOT EXISTS whole-site claim blocker
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "A single stuck claimed item on a site excludes the entire site from dispatch consideration until it clears … by design … but it makes stuck claims a system-stopping condition" (2026-05-15)
- **what:** find_dispatchable_site's NOT EXISTS clause excludes any site with ANY item in status='claimed' — an absolute blocker, not a deprioritiser. Prevents racing claims mid-execution but converts one dead handler into a site-wide stall. Proposed (cheap, high-leverage, not built): watchdog that resets claims older than ~15 min.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3
- **relations:** claim-timeout sweeper absence (Bug 2); dispatcher stall (Bug 1)
- **verify-later:** whether an auto-reset sweeper now exists

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### pipeline column as soft routing label
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "Decision reached (2026-05-15, with user): leave the field as a soft, currently-unused routing label … Not implemented yet."
- **what:** `site_work_items.pipeline` (renamed from `domain`; default 'build') is a coarse label allowing pipeline-specific dispatchers, but only build-dispatch-loop exists — 'design' and 'maintenance' items sit dormant. It duplicates what handler_agent already implies, with nothing keeping them in sync (the unfulfilled_imagery_plan check emitted pipeline='design' and stalled). Decided: discovery checks write 'build'; loosen the dispatcher to accept any value. Stale `target_domain` config keyword survives the rename.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q4
- **relations:** unfulfilled_imagery_plan check; dispatch chain
- **verify-later:** current unfulfilled_imagery_plan INSERT; build-dispatch-loop load_items item_pipeline config

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Silent completion pathology and the positive-evidence rule
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "observed and characterised; not yet fixed" (captured 2026-04-19); mode 2 later "already confirmed fixed" per FOCUS_content_quality (2026-06-09); modes 1/3 ran at 66×/47× per week (2026-04-20)
- **what:** Three modes mark work complete that isn't: reaper auto-completion on lost responses; validate_content failures inconsistently routed to complete instead of needs_human_review; 40-minute blind reaper marking claim-timeouts complete instead of resetting to triaged. Root flaw: "we're done trying" treated as "the work is done". Fix rule: complete only on explicit success response OR positive DB evidence (page_components rows, build_status='deployed', git commit). Symptoms attempt_count=0-on-success and updated_at<claimed_at belong to the same semantic muddle.
- **sources:** FOCUS_page_build_handler_silent_completion.md (whole); HANDOFF_2026-04-20_error_investigations.md#2, #3; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#C
- **relations:** claim-timeout mechanism; validate_page_content gate; two-strike rule
- **verify-later:** reaper code paths setting status='complete'; whether modes 1 and 3 were fixed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatcher response-stall and missing claim/orchestration timeout cleanup
- **category:** NEW:work-dispatch
- **status-signal:** unknown
- **status-evidence:** "Bug 1 … Blocker for autonomous cascade completion" and "Bug 2 — No claim-timeout / orchestration-timeout cleanup" (2026-04-23); every cascade needed "manual dispatcher pokes"
- **what:** build-dispatch-loop orchestrations stall at process_item_iter_N_call_handler even when the handler response arrived (suspects: Kafka consumer reconnect failure; mark_complete not firing); with no sweeper, claimed items and AWAITING_RESPONSES orchestrations accumulate forever and block sites (compounding the NOT EXISTS blocker). Fix shapes: consumer reconnect detection, periodic claim-release sweeper, force-fail of timed-out orchestrations.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md Bugs 1-2
- **relations:** NOT EXISTS blocker; consumer group race; silent completion
- **verify-later:** any sweeper added after 2026-04-23; kafka client reconnect handling

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### build-pipeline-trigger site targeting via pre_query
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Bug 3 … Scheduler-driven dispatcher invocations all default to system.internal site_id … Fix shape: Add a pre_query" (2026-04-23)
- **what:** The scheduled dispatcher fires with no site targeting so it lands on system.internal and no-ops while real sites wait. Proposed pre_query on the scheduled_tasks row selecting sites with open build items so one dispatcher fires per site.
- **sources:** HANDOFF_2026-04-23(1).md Bug 3
- **relations:** dispatch chain; find_dispatchable_site arbitrariness
- **verify-later:** scheduled_tasks.build-pipeline-trigger pre_query column value

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two-strike rule for work items
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Two-strike rule — FINAL DECISION … Decided NOT to weaken" (2026-04-23); born-unresolved pile-up pattern noted 2026-04-17
- **what:** insertWorkItem marks a new item `unresolved` when 2 prior items with the same item_key ended (complete + failed both count), breaking discover↔fix loops. Cost: items born unresolved accumulate; re-cascades hit strikes from a previous run's completes. The sanctioned fix is item_key cascade_run_id scoping (deferred), not weakening the rule. Centralised `workItemTerminalStatuses` const (work_items_common.go) keeps the dedup index and ON CONFLICT predicates from drifting.
- **sources:** HANDOFF_2026-04-23(1).md #Two-strike, deploy table; HANDOFF-pipeline-triage-april-2026.md#patterns
- **relations:** idx_swi_dedup migration 012; discovery noise on dead sites
- **verify-later:** work_items_common.go; whether cascade_run_id scoping landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery auto-triage and scheduled-audit open questions
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Q1 — discovery emissions auto-triage (still open); Q2 — scheduled audit runs (still open)" (2026-05-15)
- **what:** Should low-risk discovery emissions (e.g. needs_imagery) auto-triage via a per-check `auto_triage_emissions` flag rather than waiting for an audit run? And is design-audit-agent scheduled anywhere, or is triage operator-driven? Both parked; determine before more discovery checks ship.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q1, #Q2
- **relations:** work-item state machine
- **verify-later:** scheduled_tasks rows for design-audit-agent; DiscoveryCheck interface

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Work-item state machine (detected → triaged → claimed → complete/failed)
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Phase 2G + 2H verified end-to-end at scale. Seven hero items … processed through the full chain … without manual intervention" (2026-05-15)
- **what:** `detected` is a valid intermediate state: discovery emits at detected; design-audit-agent's `triage_detected_items` step (registry.go:722) promotes to triaged; dispatch claims only triaged/approved (partial indexes idx_swi_handler / idx_swi_site_pending); handlers mark complete/failed (mark_work_item_complete / mark_work_item_failed steps). There is NO automated coupling between discovery and audit — items sit in detected until an audit runs. Admin-created items insert directly at triaged.
- **sources:** FOCUS_dispatch_diagnostic(4).md#TL;DR, #Evidence-trail; HANDOFF-pipeline-triage-april-2026.md
- **relations:** dispatch chain; auto-triage open question; two-strike rule; silent completion
- **verify-later:** registry.go triage_detected_items; site_work_items partial indexes; design-audit-agent workflow

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatch chain: build-pipeline-trigger → find_dispatchable_site → build-dispatch-loop
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "researched in depth this session" (2026-05-15) with the actual selection SQL quoted; scheduled_tasks row build-pipeline-trigger every 30s
- **what:** The scheduler fires build-pipeline-trigger, whose find_dispatchable_site step picks ONE site per tick (DISTINCT ON with no outer ORDER BY — effectively arbitrary among eligible sites) and spawns build-dispatch-loop scoped to it, which loads up to 5 items (pipeline='build') and claims/spawns handlers. Throughput cap ~5 items per site per 30s, one site at a time. build-pipeline-trigger doesn't write orchestration_states, making its decisions untraceable.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3; HANDOFF_2026-04-23(1).md Bug 3
- **relations:** NOT EXISTS blocker; Bug 3 site-targeting; fairness ORDER BY improvement
- **verify-later:** scheduled_tasks 'build-pipeline-trigger' row; build-pipeline-trigger / build-dispatch-loop agent definitions

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### NOT EXISTS whole-site claim blocker
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "A single stuck claimed item on a site excludes the entire site from dispatch consideration until it clears … by design … but it makes stuck claims a system-stopping condition" (2026-05-15)
- **what:** find_dispatchable_site's NOT EXISTS clause excludes any site with ANY item in status='claimed' — an absolute blocker, not a deprioritiser. Prevents racing claims mid-execution but converts one dead handler into a site-wide stall. Proposed (cheap, high-leverage, not built): watchdog that resets claims older than ~15 min.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q3
- **relations:** claim-timeout sweeper absence (Bug 2); dispatcher stall (Bug 1)
- **verify-later:** whether an auto-reset sweeper now exists

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### pipeline column as soft routing label
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "Decision reached (2026-05-15, with user): leave the field as a soft, currently-unused routing label … Not implemented yet."
- **what:** `site_work_items.pipeline` (renamed from `domain`; default 'build') is a coarse label allowing pipeline-specific dispatchers, but only build-dispatch-loop exists — 'design' and 'maintenance' items sit dormant. It duplicates what handler_agent already implies, with nothing keeping them in sync (the unfulfilled_imagery_plan check emitted pipeline='design' and stalled). Decided: discovery checks write 'build'; loosen the dispatcher to accept any value. Stale `target_domain` config keyword survives the rename.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q4
- **relations:** unfulfilled_imagery_plan check; dispatch chain
- **verify-later:** current unfulfilled_imagery_plan INSERT; build-dispatch-loop load_items item_pipeline config

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Silent completion pathology and the positive-evidence rule
- **category:** NEW:work-dispatch
- **status-signal:** partial
- **status-evidence:** "observed and characterised; not yet fixed" (captured 2026-04-19); mode 2 later "already confirmed fixed" per FOCUS_content_quality (2026-06-09); modes 1/3 ran at 66×/47× per week (2026-04-20)
- **what:** Three modes mark work complete that isn't: reaper auto-completion on lost responses; validate_content failures inconsistently routed to complete instead of needs_human_review; 40-minute blind reaper marking claim-timeouts complete instead of resetting to triaged. Root flaw: "we're done trying" treated as "the work is done". Fix rule: complete only on explicit success response OR positive DB evidence (page_components rows, build_status='deployed', git commit). Symptoms attempt_count=0-on-success and updated_at<claimed_at belong to the same semantic muddle.
- **sources:** FOCUS_page_build_handler_silent_completion.md (whole); HANDOFF_2026-04-20_error_investigations.md#2, #3; HANDOFF_2026-04-20_composition_deployed_design_stuck.md#C
- **relations:** claim-timeout mechanism; validate_page_content gate; two-strike rule
- **verify-later:** reaper code paths setting status='complete'; whether modes 1 and 3 were fixed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Dispatcher response-stall and missing claim/orchestration timeout cleanup
- **category:** NEW:work-dispatch
- **status-signal:** unknown
- **status-evidence:** "Bug 1 … Blocker for autonomous cascade completion" and "Bug 2 — No claim-timeout / orchestration-timeout cleanup" (2026-04-23); every cascade needed "manual dispatcher pokes"
- **what:** build-dispatch-loop orchestrations stall at process_item_iter_N_call_handler even when the handler response arrived (suspects: Kafka consumer reconnect failure; mark_complete not firing); with no sweeper, claimed items and AWAITING_RESPONSES orchestrations accumulate forever and block sites (compounding the NOT EXISTS blocker). Fix shapes: consumer reconnect detection, periodic claim-release sweeper, force-fail of timed-out orchestrations.
- **sources:** HANDOFF_2026-04-23_dispatch_reliability_and_008_validated(1).md Bugs 1-2
- **relations:** NOT EXISTS blocker; consumer group race; silent completion
- **verify-later:** any sweeper added after 2026-04-23; kafka client reconnect handling

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### build-pipeline-trigger site targeting via pre_query
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Bug 3 … Scheduler-driven dispatcher invocations all default to system.internal site_id … Fix shape: Add a pre_query" (2026-04-23)
- **what:** The scheduled dispatcher fires with no site targeting so it lands on system.internal and no-ops while real sites wait. Proposed pre_query on the scheduled_tasks row selecting sites with open build items so one dispatcher fires per site.
- **sources:** HANDOFF_2026-04-23(1).md Bug 3
- **relations:** dispatch chain; find_dispatchable_site arbitrariness
- **verify-later:** scheduled_tasks.build-pipeline-trigger pre_query column value

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Two-strike rule for work items
- **category:** NEW:work-dispatch
- **status-signal:** deployed
- **status-evidence:** "Two-strike rule — FINAL DECISION … Decided NOT to weaken" (2026-04-23); born-unresolved pile-up pattern noted 2026-04-17
- **what:** insertWorkItem marks a new item `unresolved` when 2 prior items with the same item_key ended (complete + failed both count), breaking discover↔fix loops. Cost: items born unresolved accumulate; re-cascades hit strikes from a previous run's completes. The sanctioned fix is item_key cascade_run_id scoping (deferred), not weakening the rule. Centralised `workItemTerminalStatuses` const (work_items_common.go) keeps the dedup index and ON CONFLICT predicates from drifting.
- **sources:** HANDOFF_2026-04-23(1).md #Two-strike, deploy table; HANDOFF-pipeline-triage-april-2026.md#patterns
- **relations:** idx_swi_dedup migration 012; discovery noise on dead sites
- **verify-later:** work_items_common.go; whether cascade_run_id scoping landed

<!-- SOURCE: U02_docs024_focus_handoff.md -->
### Discovery auto-triage and scheduled-audit open questions
- **category:** NEW:work-dispatch
- **status-signal:** aspirational
- **status-evidence:** "Q1 — discovery emissions auto-triage (still open); Q2 — scheduled audit runs (still open)" (2026-05-15)
- **what:** Should low-risk discovery emissions (e.g. needs_imagery) auto-triage via a per-check `auto_triage_emissions` flag rather than waiting for an audit run? And is design-audit-agent scheduled anywhere, or is triage operator-driven? Both parked; determine before more discovery checks ship.
- **sources:** FOCUS_dispatch_diagnostic(4).md#Q1, #Q2
- **relations:** work-item state machine
- **verify-later:** scheduled_tasks rows for design-audit-agent; DiscoveryCheck interface

<!-- SOURCE: U05_content_quality_linking.md -->
### Silent-completion failure family ("work reports success but doesn't happen")
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** running_notes_15(12) Part 11 "modes 1–3 resolved; one residual gap"; Fix B "deferred, low urgency (monitor=0)"; NOTES(44) documents further members through June.
- **what:** The unit's unifying defect class: work items reach `complete` while the artifact was never produced. Members catalogued: the result-stub drop (Part 1); complete_error being a SUCCESS-labelled complete_workflow for sectionless/skip paths; error_step laundering a genuine save failure into complete; the old reaper auto-complete on lost response; claim-timeout marked complete; complete_work_item clobbering deliberate flags; deploy re-committing stale components ("git committed ≠ new content"). Doctrine that emerged: trust rendered HTML / DB state over work-item status; a blocked/failed step must surface as a non-terminal status, never `complete`.
- **sources:** running_notes_15(12).md#part-10-12; HANDOFF_2026-06-09(2).md#RESOLVED; NOTES(44) passim; page_build_handler_save_failure_visible.sql (header)
- **relations:** every fix below: evidence-gated reaper, Fix A, mark_no_sections, mark_save_failed, positive-evidence monitor. FOCUS_page_build_handler_silent_completion.md is the home doc.
- **verify-later:** FOCUS_page_build_handler_silent_completion.md; positive-evidence monitor query results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Evidence-gated claimed-item-timeout reaper (positive-evidence completion + reset)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2: "Option A APPLIED + verified live (enabled=t, interval 120s, new pre_query 14:09)"; running_notes_15(12) Part 11 re-confirms.
- **what:** The claimed-item-timeout scheduled task's SQL pre_query auto-completes a stuck claim ONLY with positive artifact evidence specific to the item type (page_components updated after claim for needs_content_page — the v2 migration decoupled it from the untrustworthy build_status='deployed' flag; deployed_at for page_rerender), else resets: attempt_count+1, back to triaged (or failed at max). Replaced the loose "any page updated since claim" auto-complete that falsely completed the gamesdesign homepage build. The reset branch made a separately-planned stale-claim watchdog redundant (reuse-not-build).
- **sources:** running_notes_14(26).md#part-11-12; running_notes_15(12).md#part-11; HANDOFF_2026-06-09(2).md#FOCUS-modes
- **relations:** silent-completion family; UpdatePageStatusAction 0-component guard (keeps the evidence honest).
- **verify-later:** scheduled_tasks claimed-item-timeout pre_query; migration_claimed_item_timeout_evidence_v2.sql.

<!-- SOURCE: U05_content_quality_linking.md -->
### complete_work_item flag-preservation guard (Fix A)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-09(2): "Fix A marked applied" in the FOCUS doc; listed in the deploy batch ("Fix A must ship with S2").
- **what:** CompleteWorkItemAction did an unconditional UPDATE to status='complete', so the dispatch loop's mark_complete clobbered deliberate handler-set flags (needs_human_review from mark_needs_review / mark_no_sections). The guard adds `AND status NOT IN (<flagged/terminal set>)` and returns completed=rows>0. Confirmed necessary by inference: the skinner-box sectionless retry proved complete_error → dispatch mark_complete fires. Prerequisite for S2 and for the existing HITL flag to be effective.
- **sources:** running_notes_15(12).md#part-11-12; HANDOFF_2026-06-09(2).md
- **relations:** silent-completion family; sectionless durability stack; workItemTerminalStatuses (needs_human_review deliberately NON-terminal).
- **verify-later:** load_work_item_actions.go CompleteWorkItemAction WHERE clause.

<!-- SOURCE: U05_content_quality_linking.md -->
### item_key canonicalization (Part 3/B) + dedup namespace decisions
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 3 — CODE PREPARED, not applied — apply after Part 2 verifies".
- **what:** item_key prefixes drifted from item_type across creators, causing two confirmed bugs: (1) content rebuilds not co-deduping (needs_page:<name> vs page_rerender:<name> for the same work → double builds); (2) adoption keying BOTH needs_tool_recreation and needs_content_page as needs_page:<name> → unique-index collision silently drops one (observed live: the pathfinding tool-recreation item mis-keyed). Fix: a plain `workItemKey(itemType, target)` builder in work_items_common.go, tool branch → its own namespace; content branch DECIDED (Option B) to stay in the needs_page namespace, preserving the deliberate doc-029 planner co-dedup — the prefix==item_type invariant carries one documented exception. Doctrine until shipped: route/diagnose by item_type → handler_agent, never by item_key.
- **sources:** NOTES(44)#item_key-contract + Part B sections; RUNBOOK_gamesdesign_index_rebuild(29).md#part-3; HANDOFF_page_pipeline(11).md#6
- **relations:** dedup index; work-item routing; adoption apply_adoption_plan.
- **verify-later:** work_items_common.go workItemKey; apply_adoption_plan_action.go lines ~627–655; P3.2 survey results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item dedup index + two-strike anti-churn rule
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 8 "Key enabling discovery (insertWorkItem …): a built-in two-strike rule".
- **what:** idx_swi_dedup is a partial UNIQUE(site_id,item_key) over non-terminal statuses only — terminal rows are excluded so a completed key can be requeued cleanly; ON CONFLICT DO NOTHING is the safe insert idiom. insertWorkItem adds a two-strike rule (an item_key with ≥2 terminal attempts in 7 days inserts as `unresolved`; <3h after a terminal item is suppressed), so discovery checks need no anti-churn logic of their own. A non-terminal flag (needs_human_review) deliberately holds the dedup slot, preventing re-trigger loops.
- **sources:** running_notes_15(12).md#part-8; HANDOFF_page_pipeline(11).md#schema-gotchas; RUNBOOK_linking_phantom_fixes(7).md#5
- **relations:** item_key canonicalization; sectionless durability stack.
- **verify-later:** insertWorkItem two-strike logic; idx_swi_dedup definition.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item routing map (item_type → handler agent)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19 "Corrected trigger (confirmed from live config, not inferred)".
- **what:** needs_page / needs_content_page / link_resolution_rebuild → page-build-handler (full content path through the writer); page_rerender → page-rerender (assemble-from-DB + deploy; after Part 2 also the no-LLM re-render pre-pass); needs_rerender → rerender-pages (site loop that mints per-page page_rerender items via create_rerender_items); needs_tool_recreation → tool-recreation-handler. build-dispatch-loop claims status triaged/approved only; discovery findings land 'detected' (unclaimable). page-build-handler does NOT branch on item_type — dispatch metadata only; it reads spec.page_id/page_name/mode/suggestion.
- **sources:** NOTES(44) 2026-06-19; RUNBOOK_linking_phantom_fixes(7).md#5 handler facts; running_notes_17(21).md#page-build-handler-contract
- **relations:** re-render vs rebuild; interactive clobber (link_resolution_rebuild routed to the full builder by design); dispatch throughput.
- **verify-later:** build-dispatch-loop claim SQL; agent workflow defs for the four handlers.

<!-- SOURCE: U09_adoption.md -->
### Silent-completion family: "complete" means "we stopped", not "the work succeeded"
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** "2026-06-09 update: modes 1–3 are now resolved in current code; one residual gap remains (`complete_work_item` clobber)… Fix A (applied 2026-06-09)… Fix B (deferred, low urgency given monitor=0)."
- **what:** The architectural flaw that a work item reaches `status='complete'` without the work succeeding, in several modes: (1) reaper auto-complete on lost handler responses ("Auto-completed: work verified done despite lost response"), (2) validate_content failures routed to complete, (3) claim-timeout marked complete instead of reset, plus the dispatch-level variants — the unguarded `CompleteWorkItemAction` clobbering handler-set `needs_human_review` flags (Fix A: status guard applied) and `complete_error` being a SUCCESS-labelled `complete_workflow` on genuine-failure paths (Fix B, deferred). Modes 1–3 resolved via the evidence-gated reaper; the rule is: complete only on explicit handler success OR positive DB evidence. The gamesdesign homepage (deployed+stamped in DB, no file in repo) and guide-skinner-box were direct consequences.
- **sources:** FOCUS_page_build_handler_silent_completion(1).md, HANDOFF_2026-06-09, running_notes_15(10)#part-10–12, CATALOGUE(9)#A4
- **relations:** claimed-item-timeout reaper; positive-evidence completion; sectionless-page durability (S2 depends on Fix A); work-item lifecycle (001)
- **verify-later:** `load_work_item_actions.go` CompleteWorkItemAction guard; page-build-handler `complete_error` semantics; monitor query results

<!-- SOURCE: U09_adoption.md -->
### claimed-item-timeout scheduled task: evidence-gated completion + stale-claim reset
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "Option A APPLIED + verified live: the v2 migration is in place — claimed-item-timeout shows enabled=t, interval_seconds=120, and the new 'provably done on the specific targeted artifact' pre_query" (2026-06-04); "Mode 1… RESOLVED… Mode 3… RESOLVED" (2026-06-09 re-verification).
- **what:** A scheduled task whose SQL pre_query both (a) auto-completes a stuck claimed item only with positive artifact evidence — `page_components` with `component_id` + non-empty `rendered_html` + `updated_at > claimed_at` for needs_content_page, `deployed_at > claimed_at` for page_rerender, head-slot update for needs_design — and (b) resets stale claims (>40 min, no evidence) to `triaged` (or `failed` at max_attempts) with attempt_count+1. The reset CTE IS the Lever-C claim watchdog the dispatch doc designed — building a separate watchdog was explicitly cancelled as duplication ("REVISED 2026-06-04 — DO NOT BUILD THIS. The reset already exists."). Evidence deliberately prefers ground-truth artifacts over the untrustworthy `build_status='deployed'` flag.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#decision, running_notes_14(25)#part-12, FOCUS_page_build_handler_silent_completion(1).md#update
- **relations:** silent-completion family; Option B deployed-guard keeps the flag honest; dispatch NOT-EXISTS deadlock (the reset unfreezes the site)
- **verify-later:** scheduled_tasks row `claimed-item-timeout` pre_query (v2: page_components evidence for needs_content_page); 40-min threshold tuning note (~25 min floor above the 1200s call_handler)

<!-- SOURCE: U09_adoption.md -->
### Positive-evidence deploy guard (0-component page never marked deployed)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "(B) DONE — v3_site_actions.go patch delivered… UpdatePageStatusAction now calls pageHasComponents(pageID) before marking deployed; if 0 rendered components it refuses and flips to needs_rebuild (clearing the stamp)" (CATALOGUE A4, applied 2026-06-04).
- **what:** `pageHasComponents` (EXISTS on page_components with non-null component_id + non-empty rendered_html) gates the `deployed` status write; a 0-component page flips to `needs_rebuild` with stamp cleared so the reconciler rebuilds instead of skipping. Fail-open on check error so a transient failure can't halt legitimate deploys. Makes `build_status='deployed'` trustworthy for downstream evidence checks.
- **sources:** CATALOGUE(9)#A4, running_notes_14(25)#part-12-addendum-2
- **relations:** silent-completion family; claimed-item-timeout evidence
- **verify-later:** v3_site_actions.go pageHasComponents + UpdatePageStatusAction deployed branch

<!-- SOURCE: U05_content_quality_linking.md -->
### Silent-completion failure family ("work reports success but doesn't happen")
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** running_notes_15(12) Part 11 "modes 1–3 resolved; one residual gap"; Fix B "deferred, low urgency (monitor=0)"; NOTES(44) documents further members through June.
- **what:** The unit's unifying defect class: work items reach `complete` while the artifact was never produced. Members catalogued: the result-stub drop (Part 1); complete_error being a SUCCESS-labelled complete_workflow for sectionless/skip paths; error_step laundering a genuine save failure into complete; the old reaper auto-complete on lost response; claim-timeout marked complete; complete_work_item clobbering deliberate flags; deploy re-committing stale components ("git committed ≠ new content"). Doctrine that emerged: trust rendered HTML / DB state over work-item status; a blocked/failed step must surface as a non-terminal status, never `complete`.
- **sources:** running_notes_15(12).md#part-10-12; HANDOFF_2026-06-09(2).md#RESOLVED; NOTES(44) passim; page_build_handler_save_failure_visible.sql (header)
- **relations:** every fix below: evidence-gated reaper, Fix A, mark_no_sections, mark_save_failed, positive-evidence monitor. FOCUS_page_build_handler_silent_completion.md is the home doc.
- **verify-later:** FOCUS_page_build_handler_silent_completion.md; positive-evidence monitor query results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Evidence-gated claimed-item-timeout reaper (positive-evidence completion + reset)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_14(26) Part 12 addendum 2: "Option A APPLIED + verified live (enabled=t, interval 120s, new pre_query 14:09)"; running_notes_15(12) Part 11 re-confirms.
- **what:** The claimed-item-timeout scheduled task's SQL pre_query auto-completes a stuck claim ONLY with positive artifact evidence specific to the item type (page_components updated after claim for needs_content_page — the v2 migration decoupled it from the untrustworthy build_status='deployed' flag; deployed_at for page_rerender), else resets: attempt_count+1, back to triaged (or failed at max). Replaced the loose "any page updated since claim" auto-complete that falsely completed the gamesdesign homepage build. The reset branch made a separately-planned stale-claim watchdog redundant (reuse-not-build).
- **sources:** running_notes_14(26).md#part-11-12; running_notes_15(12).md#part-11; HANDOFF_2026-06-09(2).md#FOCUS-modes
- **relations:** silent-completion family; UpdatePageStatusAction 0-component guard (keeps the evidence honest).
- **verify-later:** scheduled_tasks claimed-item-timeout pre_query; migration_claimed_item_timeout_evidence_v2.sql.

<!-- SOURCE: U05_content_quality_linking.md -->
### complete_work_item flag-preservation guard (Fix A)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** HANDOFF_2026-06-09(2): "Fix A marked applied" in the FOCUS doc; listed in the deploy batch ("Fix A must ship with S2").
- **what:** CompleteWorkItemAction did an unconditional UPDATE to status='complete', so the dispatch loop's mark_complete clobbered deliberate handler-set flags (needs_human_review from mark_needs_review / mark_no_sections). The guard adds `AND status NOT IN (<flagged/terminal set>)` and returns completed=rows>0. Confirmed necessary by inference: the skinner-box sectionless retry proved complete_error → dispatch mark_complete fires. Prerequisite for S2 and for the existing HITL flag to be effective.
- **sources:** running_notes_15(12).md#part-11-12; HANDOFF_2026-06-09(2).md
- **relations:** silent-completion family; sectionless durability stack; workItemTerminalStatuses (needs_human_review deliberately NON-terminal).
- **verify-later:** load_work_item_actions.go CompleteWorkItemAction WHERE clause.

<!-- SOURCE: U05_content_quality_linking.md -->
### item_key canonicalization (Part 3/B) + dedup namespace decisions
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** HANDOFF_page_pipeline(11) §2: "Part 3 — CODE PREPARED, not applied — apply after Part 2 verifies".
- **what:** item_key prefixes drifted from item_type across creators, causing two confirmed bugs: (1) content rebuilds not co-deduping (needs_page:<name> vs page_rerender:<name> for the same work → double builds); (2) adoption keying BOTH needs_tool_recreation and needs_content_page as needs_page:<name> → unique-index collision silently drops one (observed live: the pathfinding tool-recreation item mis-keyed). Fix: a plain `workItemKey(itemType, target)` builder in work_items_common.go, tool branch → its own namespace; content branch DECIDED (Option B) to stay in the needs_page namespace, preserving the deliberate doc-029 planner co-dedup — the prefix==item_type invariant carries one documented exception. Doctrine until shipped: route/diagnose by item_type → handler_agent, never by item_key.
- **sources:** NOTES(44)#item_key-contract + Part B sections; RUNBOOK_gamesdesign_index_rebuild(29).md#part-3; HANDOFF_page_pipeline(11).md#6
- **relations:** dedup index; work-item routing; adoption apply_adoption_plan.
- **verify-later:** work_items_common.go workItemKey; apply_adoption_plan_action.go lines ~627–655; P3.2 survey results.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item dedup index + two-strike anti-churn rule
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** running_notes_15(12) Part 8 "Key enabling discovery (insertWorkItem …): a built-in two-strike rule".
- **what:** idx_swi_dedup is a partial UNIQUE(site_id,item_key) over non-terminal statuses only — terminal rows are excluded so a completed key can be requeued cleanly; ON CONFLICT DO NOTHING is the safe insert idiom. insertWorkItem adds a two-strike rule (an item_key with ≥2 terminal attempts in 7 days inserts as `unresolved`; <3h after a terminal item is suppressed), so discovery checks need no anti-churn logic of their own. A non-terminal flag (needs_human_review) deliberately holds the dedup slot, preventing re-trigger loops.
- **sources:** running_notes_15(12).md#part-8; HANDOFF_page_pipeline(11).md#schema-gotchas; RUNBOOK_linking_phantom_fixes(7).md#5
- **relations:** item_key canonicalization; sectionless durability stack.
- **verify-later:** insertWorkItem two-strike logic; idx_swi_dedup definition.

<!-- SOURCE: U05_content_quality_linking.md -->
### Work-item routing map (item_type → handler agent)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** NOTES(44) 2026-06-19 "Corrected trigger (confirmed from live config, not inferred)".
- **what:** needs_page / needs_content_page / link_resolution_rebuild → page-build-handler (full content path through the writer); page_rerender → page-rerender (assemble-from-DB + deploy; after Part 2 also the no-LLM re-render pre-pass); needs_rerender → rerender-pages (site loop that mints per-page page_rerender items via create_rerender_items); needs_tool_recreation → tool-recreation-handler. build-dispatch-loop claims status triaged/approved only; discovery findings land 'detected' (unclaimable). page-build-handler does NOT branch on item_type — dispatch metadata only; it reads spec.page_id/page_name/mode/suggestion.
- **sources:** NOTES(44) 2026-06-19; RUNBOOK_linking_phantom_fixes(7).md#5 handler facts; running_notes_17(21).md#page-build-handler-contract
- **relations:** re-render vs rebuild; interactive clobber (link_resolution_rebuild routed to the full builder by design); dispatch throughput.
- **verify-later:** build-dispatch-loop claim SQL; agent workflow defs for the four handlers.

<!-- SOURCE: U09_adoption.md -->
### Silent-completion family: "complete" means "we stopped", not "the work succeeded"
- **category:** NEW:work-item-integrity
- **status-signal:** partial
- **status-evidence:** "2026-06-09 update: modes 1–3 are now resolved in current code; one residual gap remains (`complete_work_item` clobber)… Fix A (applied 2026-06-09)… Fix B (deferred, low urgency given monitor=0)."
- **what:** The architectural flaw that a work item reaches `status='complete'` without the work succeeding, in several modes: (1) reaper auto-complete on lost handler responses ("Auto-completed: work verified done despite lost response"), (2) validate_content failures routed to complete, (3) claim-timeout marked complete instead of reset, plus the dispatch-level variants — the unguarded `CompleteWorkItemAction` clobbering handler-set `needs_human_review` flags (Fix A: status guard applied) and `complete_error` being a SUCCESS-labelled `complete_workflow` on genuine-failure paths (Fix B, deferred). Modes 1–3 resolved via the evidence-gated reaper; the rule is: complete only on explicit handler success OR positive DB evidence. The gamesdesign homepage (deployed+stamped in DB, no file in repo) and guide-skinner-box were direct consequences.
- **sources:** FOCUS_page_build_handler_silent_completion(1).md, HANDOFF_2026-06-09, running_notes_15(10)#part-10–12, CATALOGUE(9)#A4
- **relations:** claimed-item-timeout reaper; positive-evidence completion; sectionless-page durability (S2 depends on Fix A); work-item lifecycle (001)
- **verify-later:** `load_work_item_actions.go` CompleteWorkItemAction guard; page-build-handler `complete_error` semantics; monitor query results

<!-- SOURCE: U09_adoption.md -->
### claimed-item-timeout scheduled task: evidence-gated completion + stale-claim reset
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "Option A APPLIED + verified live: the v2 migration is in place — claimed-item-timeout shows enabled=t, interval_seconds=120, and the new 'provably done on the specific targeted artifact' pre_query" (2026-06-04); "Mode 1… RESOLVED… Mode 3… RESOLVED" (2026-06-09 re-verification).
- **what:** A scheduled task whose SQL pre_query both (a) auto-completes a stuck claimed item only with positive artifact evidence — `page_components` with `component_id` + non-empty `rendered_html` + `updated_at > claimed_at` for needs_content_page, `deployed_at > claimed_at` for page_rerender, head-slot update for needs_design — and (b) resets stale claims (>40 min, no evidence) to `triaged` (or `failed` at max_attempts) with attempt_count+1. The reset CTE IS the Lever-C claim watchdog the dispatch doc designed — building a separate watchdog was explicitly cancelled as duplication ("REVISED 2026-06-04 — DO NOT BUILD THIS. The reset already exists."). Evidence deliberately prefers ground-truth artifacts over the untrustworthy `build_status='deployed'` flag.
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md#decision, running_notes_14(25)#part-12, FOCUS_page_build_handler_silent_completion(1).md#update
- **relations:** silent-completion family; Option B deployed-guard keeps the flag honest; dispatch NOT-EXISTS deadlock (the reset unfreezes the site)
- **verify-later:** scheduled_tasks row `claimed-item-timeout` pre_query (v2: page_components evidence for needs_content_page); 40-min threshold tuning note (~25 min floor above the 1200s call_handler)

<!-- SOURCE: U09_adoption.md -->
### Positive-evidence deploy guard (0-component page never marked deployed)
- **category:** NEW:work-item-integrity
- **status-signal:** deployed
- **status-evidence:** "(B) DONE — v3_site_actions.go patch delivered… UpdatePageStatusAction now calls pageHasComponents(pageID) before marking deployed; if 0 rendered components it refuses and flips to needs_rebuild (clearing the stamp)" (CATALOGUE A4, applied 2026-06-04).
- **what:** `pageHasComponents` (EXISTS on page_components with non-null component_id + non-empty rendered_html) gates the `deployed` status write; a 0-component page flips to `needs_rebuild` with stamp cleared so the reconciler rebuilds instead of skipping. Fail-open on check error so a transient failure can't halt legitimate deploys. Makes `build_status='deployed'` trustworthy for downstream evidence checks.
- **sources:** CATALOGUE(9)#A4, running_notes_14(25)#part-12-addendum-2
- **relations:** silent-completion family; claimed-item-timeout evidence
- **verify-later:** v3_site_actions.go pageHasComponents + UpdatePageStatusAction deployed branch

<!-- SOURCE: U01_docs024_numbered_core.md -->
### `pipeline` column as work-item routing namespace
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5): renamed from `domain` (clash with site domain); dispatch loop filters pipeline='build'
- **what:** site_work_items.pipeline ('build','design','maintenance'…) is the dispatch routing namespace, never the website domain. Everything in the initial build must be pipeline='build'; items emitted straight-to-triaged must set it at emission (triage rewrites it for detected items). Historical trap: dispatch once passed the namespace to handlers as the site domain.
- **sources:** 001(5)#Work item pipeline must be "build"; 016 Schema reminders
- **relations:** dispatch loop; schema-drift rule
- **verify-later:** find_dispatchable_site / LoadWorkItemsAction filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Terminal work items: every pipeline ends with assembly + deployment
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5) rule with the minimum brochure chain
- **what:** The planner/WriteBuildItems must create terminal items (needs_rerender, priority ~20/99) that re-render site components, assemble pages from page_components + site_components, git commit, and trigger deploy — otherwise the pipeline produces data but no website.
- **sources:** 001(5)#Every pipeline must end with assembly; 004 step 9 (needs_rerender priority 99)
- **relations:** commit-is-deploy; rerender agents
- **verify-later:** WriteBuildItemsAction

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item blocking/unblocking and the `unresolved` two-strike mechanism
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 006 "Two-strike dedup ✅ Deployed"; 001(5) describes insertWorkItem mechanics
- **what:** Three block causes (missing handler → feasibility-recheck auto-promotes; spec blocked → manual; manual). insertWorkItem suppresses re-detection within 3h of a terminal duplicate and, after 2+ terminal attempts in 7 days, creates new items as status `unresolved` — visible, not dispatched. `wont_fix` + "superseded by active duplicate" is the dedup system working, not a bug.
- **sources:** 001(5)#Work Item Lifecycle; 006 issues table (12k duplicate cleanup); 016 §9 wont_fix entry
- **relations:** idx_swi_dedup; feasibility-recheck
- **verify-later:** load_work_item_actions.go insertWorkItem

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Unified build & maintenance via site_work_items (single queue, same code)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4): "Build and maintenance are the same process"; live lifecycle steps 3/7/11 "same code"
- **what:** Every piece of work is a site_work_items row (source, pipeline, item_type, severity, spec jsonb, handler_agent, status enum incl. needs_human_review, priority, depends_on uuid[], item_key dedup, result with commit_sha). New site = planner-written items; maintenance = discovery-written items; same orchestrator/dispatch/handlers. Cross-domain coordination happens only through the table (side_effect items with parent_item_id), never agent-to-agent calls.
- **sources:** 002(4)#Unified Build and Maintenance, #site_work_items; 003(8)#Cross-Domain Coordination
- **relations:** dispatch loop; work-item state machine (016); P1 expansion sources table
- **verify-later:** site_work_items schema incl. depends_on, item_key

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item routing: content rebuild vs re-render (needs_page / page_rerender / needs_rerender / link_resolution_rebuild)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4) routing table dated with 2026-06-22 hazard confirmation
- **what:** Route by item_type never item_key; the distinction is whether copy is regenerated. needs_page/needs_content_page → full LLM rebuild via page-build-handler; page_rerender (reason image_landed/section_data_resolved) → no-LLM re-resolve + re-render from stored content_data; needs_rerender → batch reassembly; link_resolution_rebuild is INTENDED links-only but runs the full writer (hazard). page_rerender on a NULL-content_data section escalates the page to needs_page (backfills content_data).
- **sources:** 002(4)#Work-item routing; 003(8)#Source of truth principle (two re-render paths)
- **relations:** interactive-page de-tool hazard; content_data source of truth
- **verify-later:** rerender_page_sections action; flag_page_image_rebuild/reconcile_section_data emitters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item state machine, transition ownership, and site-exclusion by stuck claim
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 016 dedicated section with ownership table and the 2026-05-14 mis-diagnosis lesson
- **what:** detected→(design-audit-agent triage_detected_items)→triaged→(build-dispatch-loop claim)→claimed→(handler)→complete/failed; admin inserts at triaged. Most common "won't dispatch" cause: one stuck claimed item excludes the ENTIRE site via find_dispatchable_site's NOT EXISTS. Debugging trap #1: don't infer writers from readers (indexes show the read path; grep the verb for the writer).
- **sources:** 016#Work item lifecycle; #Site excluded from dispatch
- **relations:** two-strike; claimed-item-timeout; operator reset SQL
- **verify-later:** triage_detected_items registration (registry.go:722)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### needs_section_data: resolvable-by-query vs genuinely-human, and the section-data reconciler direction
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** 016 §9 (2026-05-27 direction): pages_where_type implemented, pages_under_section named-but-unimplemented; reconciler recorded in FUTURE_section_data_handler
- **what:** Some needs_section_data items are human-only (team, pricing); list/grid sections source from query.* and resolve mechanically once pages exist — but the unimplemented query name or pre-active timing defers them anyway. Read spec.missing[].source to classify. Direction: implement pages_under_section + a lightweight resolver (not an LLM agent) that re-attempts open items via queryresolve, closes via closeResolvedDataRequest, and flags re-render.
- **sources:** 016 §9 needs_section_data entry; 002(4) reconcile_section_data → page_rerender emitters
- **relations:** input schema v2; deferral-drop bug
- **verify-later:** queryresolve vocabulary today

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Side-effect rules engine (deterministic follow-on items)
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** P1 table of triggers; some live (side_effect source in 002/003 contracts), full Go rules engine unconfirmed
- **what:** After each handler completes, deterministic Go rules (not LLM) emit follow-ons: new page → needs_nav_update + needs_sitemap; deletion → redirects; CSS change → needs_rerender; milestone item types → needs_snapshot.
- **sources:** P1#Side effects; 003(8)#Cross-Domain Coordination
- **relations:** cross-domain coordination; snapshots triggers
- **verify-later:** rules engine implementation

<!-- SOURCE: U01_docs024_numbered_core.md -->
### `pipeline` column as work-item routing namespace
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5): renamed from `domain` (clash with site domain); dispatch loop filters pipeline='build'
- **what:** site_work_items.pipeline ('build','design','maintenance'…) is the dispatch routing namespace, never the website domain. Everything in the initial build must be pipeline='build'; items emitted straight-to-triaged must set it at emission (triage rewrites it for detected items). Historical trap: dispatch once passed the namespace to handlers as the site domain.
- **sources:** 001(5)#Work item pipeline must be "build"; 016 Schema reminders
- **relations:** dispatch loop; schema-drift rule
- **verify-later:** find_dispatchable_site / LoadWorkItemsAction filters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Terminal work items: every pipeline ends with assembly + deployment
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 001(5) rule with the minimum brochure chain
- **what:** The planner/WriteBuildItems must create terminal items (needs_rerender, priority ~20/99) that re-render site components, assemble pages from page_components + site_components, git commit, and trigger deploy — otherwise the pipeline produces data but no website.
- **sources:** 001(5)#Every pipeline must end with assembly; 004 step 9 (needs_rerender priority 99)
- **relations:** commit-is-deploy; rerender agents
- **verify-later:** WriteBuildItemsAction

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item blocking/unblocking and the `unresolved` two-strike mechanism
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 006 "Two-strike dedup ✅ Deployed"; 001(5) describes insertWorkItem mechanics
- **what:** Three block causes (missing handler → feasibility-recheck auto-promotes; spec blocked → manual; manual). insertWorkItem suppresses re-detection within 3h of a terminal duplicate and, after 2+ terminal attempts in 7 days, creates new items as status `unresolved` — visible, not dispatched. `wont_fix` + "superseded by active duplicate" is the dedup system working, not a bug.
- **sources:** 001(5)#Work Item Lifecycle; 006 issues table (12k duplicate cleanup); 016 §9 wont_fix entry
- **relations:** idx_swi_dedup; feasibility-recheck
- **verify-later:** load_work_item_actions.go insertWorkItem

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Unified build & maintenance via site_work_items (single queue, same code)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4): "Build and maintenance are the same process"; live lifecycle steps 3/7/11 "same code"
- **what:** Every piece of work is a site_work_items row (source, pipeline, item_type, severity, spec jsonb, handler_agent, status enum incl. needs_human_review, priority, depends_on uuid[], item_key dedup, result with commit_sha). New site = planner-written items; maintenance = discovery-written items; same orchestrator/dispatch/handlers. Cross-domain coordination happens only through the table (side_effect items with parent_item_id), never agent-to-agent calls.
- **sources:** 002(4)#Unified Build and Maintenance, #site_work_items; 003(8)#Cross-Domain Coordination
- **relations:** dispatch loop; work-item state machine (016); P1 expansion sources table
- **verify-later:** site_work_items schema incl. depends_on, item_key

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item routing: content rebuild vs re-render (needs_page / page_rerender / needs_rerender / link_resolution_rebuild)
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 002(4) routing table dated with 2026-06-22 hazard confirmation
- **what:** Route by item_type never item_key; the distinction is whether copy is regenerated. needs_page/needs_content_page → full LLM rebuild via page-build-handler; page_rerender (reason image_landed/section_data_resolved) → no-LLM re-resolve + re-render from stored content_data; needs_rerender → batch reassembly; link_resolution_rebuild is INTENDED links-only but runs the full writer (hazard). page_rerender on a NULL-content_data section escalates the page to needs_page (backfills content_data).
- **sources:** 002(4)#Work-item routing; 003(8)#Source of truth principle (two re-render paths)
- **relations:** interactive-page de-tool hazard; content_data source of truth
- **verify-later:** rerender_page_sections action; flag_page_image_rebuild/reconcile_section_data emitters

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Work-item state machine, transition ownership, and site-exclusion by stuck claim
- **category:** NEW:work-item-system
- **status-signal:** deployed
- **status-evidence:** 016 dedicated section with ownership table and the 2026-05-14 mis-diagnosis lesson
- **what:** detected→(design-audit-agent triage_detected_items)→triaged→(build-dispatch-loop claim)→claimed→(handler)→complete/failed; admin inserts at triaged. Most common "won't dispatch" cause: one stuck claimed item excludes the ENTIRE site via find_dispatchable_site's NOT EXISTS. Debugging trap #1: don't infer writers from readers (indexes show the read path; grep the verb for the writer).
- **sources:** 016#Work item lifecycle; #Site excluded from dispatch
- **relations:** two-strike; claimed-item-timeout; operator reset SQL
- **verify-later:** triage_detected_items registration (registry.go:722)

<!-- SOURCE: U01_docs024_numbered_core.md -->
### needs_section_data: resolvable-by-query vs genuinely-human, and the section-data reconciler direction
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** 016 §9 (2026-05-27 direction): pages_where_type implemented, pages_under_section named-but-unimplemented; reconciler recorded in FUTURE_section_data_handler
- **what:** Some needs_section_data items are human-only (team, pricing); list/grid sections source from query.* and resolve mechanically once pages exist — but the unimplemented query name or pre-active timing defers them anyway. Read spec.missing[].source to classify. Direction: implement pages_under_section + a lightweight resolver (not an LLM agent) that re-attempts open items via queryresolve, closes via closeResolvedDataRequest, and flags re-render.
- **sources:** 016 §9 needs_section_data entry; 002(4) reconcile_section_data → page_rerender emitters
- **relations:** input schema v2; deferral-drop bug
- **verify-later:** queryresolve vocabulary today

<!-- SOURCE: U01_docs024_numbered_core.md -->
### Side-effect rules engine (deterministic follow-on items)
- **category:** NEW:work-item-system
- **status-signal:** partial
- **status-evidence:** P1 table of triggers; some live (side_effect source in 002/003 contracts), full Go rules engine unconfirmed
- **what:** After each handler completes, deterministic Go rules (not LLM) emit follow-ons: new page → needs_nav_update + needs_sitemap; deletion → redirects; CSS change → needs_rerender; milestone item types → needs_snapshot.
- **sources:** P1#Side effects; 003(8)#Cross-Domain Coordination
- **relations:** cross-domain coordination; snapshots triggers
- **verify-later:** rules engine implementation

<!-- SOURCE: U09_adoption.md -->
### Dispatch throughput: one-site-per-tick + absolute NOT-EXISTS exclusion
- **category:** NEW:dispatch-pipeline
- **status-signal:** partial
- **status-evidence:** "Investigation complete; Lever C… effectively already implemented; the remaining actions are… the guardrails (timeout alignment, fairness ORDER BY, git-adapter retry), which are the real remaining dispatch work" (FOCUS_dispatch(3), revised 2026-06-04).
- **what:** `find_dispatchable_site` selects `DISTINCT ON (site_id) … LIMIT 1` with a `NOT EXISTS` clause excluding a site entirely while any item is `claimed` — so tools dispatch serially within a site (~5-min Opus builds; 11 tools ≈ an hour minimum), a dead handler freezes the whole site until the reaper resets the claim (observed 47–67 min gaps), and lowest-UUID sites can starve others (no fairness ordering; 8 concurrent triggers converge on one site). Levers: A multi-site decoupling (lower priority — cadence is already 30s/8, corrected from stale doc values), B bounded per-site concurrency (K=2–3, gated on OOM guardrails), C claim watchdog (already exists). Scheduler timeout mismatch (300s task vs 900/1200/4200s work) is a latent over-spawn/OOM risk; guardrails: accurate handler memory requests (Pending is the safe failure mode vs Evicted), timeout alignment, fairness ORDER BY. Claim mechanics: claim is an atomic conditional UPDATE; `claimed_by` is the agent type, not an orchestration id — the orchestration-liveness fast path needs a schema addition (follow-up).
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md, CATALOGUE(9)#family-j, running_notes_14(25)#part-9
- **relations:** claimed-item-timeout reset; git-adapter race (gates Lever A); build-pipeline-trigger scheduled task; stale agent-definition descriptions mislead (build-dispatch-loop described as self-spawning single-item, actually a 5-iteration loop)
- **verify-later:** find_dispatchable_site SQL (line ~276 NOT EXISTS); scheduled_tasks build-pipeline-trigger row; handler pod requests/limits

<!-- SOURCE: U09_adoption.md -->
### Dispatch throughput: one-site-per-tick + absolute NOT-EXISTS exclusion
- **category:** NEW:dispatch-pipeline
- **status-signal:** partial
- **status-evidence:** "Investigation complete; Lever C… effectively already implemented; the remaining actions are… the guardrails (timeout alignment, fairness ORDER BY, git-adapter retry), which are the real remaining dispatch work" (FOCUS_dispatch(3), revised 2026-06-04).
- **what:** `find_dispatchable_site` selects `DISTINCT ON (site_id) … LIMIT 1` with a `NOT EXISTS` clause excluding a site entirely while any item is `claimed` — so tools dispatch serially within a site (~5-min Opus builds; 11 tools ≈ an hour minimum), a dead handler freezes the whole site until the reaper resets the claim (observed 47–67 min gaps), and lowest-UUID sites can starve others (no fairness ordering; 8 concurrent triggers converge on one site). Levers: A multi-site decoupling (lower priority — cadence is already 30s/8, corrected from stale doc values), B bounded per-site concurrency (K=2–3, gated on OOM guardrails), C claim watchdog (already exists). Scheduler timeout mismatch (300s task vs 900/1200/4200s work) is a latent over-spawn/OOM risk; guardrails: accurate handler memory requests (Pending is the safe failure mode vs Evicted), timeout alignment, fairness ORDER BY. Claim mechanics: claim is an atomic conditional UPDATE; `claimed_by` is the agent type, not an orchestration id — the orchestration-liveness fast path needs a schema addition (follow-up).
- **sources:** FOCUS_dispatch_throughput_and_claim_watchdog(3).md, CATALOGUE(9)#family-j, running_notes_14(25)#part-9
- **relations:** claimed-item-timeout reset; git-adapter race (gates Lever A); build-pipeline-trigger scheduled task; stale agent-definition descriptions mislead (build-dispatch-loop described as self-spawning single-item, actually a 5-iteration loop)
- **verify-later:** find_dispatchable_site SQL (line ~276 NOT EXISTS); scheduled_tasks build-pipeline-trigger row; handler pod requests/limits

<!-- SOURCE: U21_legacy_docs_b.md -->
### MVP build squad lineage (chief-strategist → architect → content-creator → deployer)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs009/001 defines the 4 MVP agents with Kafka payloads; docs015/004 shows the mature descendant ("pageflow-builder: ensure_site_record → call_site_planner → sync_pages → populate_nav → image generation → style collection → build_pages_loop").
- **what:** The builder pipeline's evolutionary line: mvp-site-builder (4 agents, single page) → landing-page-builder / content-site-builder (specialist architects per site type) → multipage-website-builder (batching, then sequential loop) → pageflow-builder (DB-backed pages, per-page loop with review and git commit) → site-work-orchestrator (work-item driven). Each generation kept strategist/planner, writer, assembler, deployer roles while moving state from CollectedData into the database.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-Agent-chief-strategist; docs006_workflow_builder/003_current_state_of_agents.sql#SPECIALIST-ARCHITECT-SYSTEM; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** unified work items (final form); loop action; site-planner; deployment-github.
- **verify-later:** which builder agent_definitions still exist and which have traffic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Sequential page generation (Phase 0 multipage fix)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs010/019 "Current Problem: Multipage-website-builder tries to generate 4 pages at once — race conditions... Solution: make it work like landing-page-builder: sequential, one page at a time"; two-delay spawn fix noted FIXED in docs009/003.
- **what:** Replacing parallel batch spawning with a strategist-planned page list iterated by the loop action (research → write per page), a wrap_multipage action generating navigation and collecting assets, and spawn timing fixed by double initialization delays. The stabilization step that made multipage builds reliable enough for everything after.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Status
- **relations:** loop action; batched generation (predecessor); pageflow-builder (successor).
- **verify-later:** wrap_multipage action; spawn_actions.go delay logic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Selective rebuild via build_status
- **category:** NEW:site-build-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/004 ("get_pages_to_build filters by build_status IN ('planned','needs_rebuild')"); docs015/003 documents the stale-page trap ("If the site planner didn't include use-cases in the new plan... it shows in nav but has stale content").
- **what:** Two orthogonal page state columns — status (active/deleted/needs_attention lifecycle) and build_status (planned/needs_rebuild/deployed) — let rebuilds touch only marked pages. Known failure mode: pages absent from a new plan silently keep old content while remaining in nav. Ancestor of work-item-driven rebuild targeting.
- **sources:** docs018_rerendering/004_trigger_just_pages_that_need_rebuild.md; docs015_data_flow_verification/003_temp_doc_rebuild_flow.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** page-rebuild agent; maintenance_queue (proto-work-items); work items.
- **verify-later:** get_pages_to_build action; build_status usage today.

<!-- SOURCE: U21_legacy_docs_b.md -->
### MVP build squad lineage (chief-strategist → architect → content-creator → deployer)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs009/001 defines the 4 MVP agents with Kafka payloads; docs015/004 shows the mature descendant ("pageflow-builder: ensure_site_record → call_site_planner → sync_pages → populate_nav → image generation → style collection → build_pages_loop").
- **what:** The builder pipeline's evolutionary line: mvp-site-builder (4 agents, single page) → landing-page-builder / content-site-builder (specialist architects per site type) → multipage-website-builder (batching, then sequential loop) → pageflow-builder (DB-backed pages, per-page loop with review and git commit) → site-work-orchestrator (work-item driven). Each generation kept strategist/planner, writer, assembler, deployer roles while moving state from CollectedData into the database.
- **sources:** docs009_site_interrogation_and_solutions/001_gemini_discussions_multipage.md#1-Agent-chief-strategist; docs006_workflow_builder/003_current_state_of_agents.sql#SPECIALIST-ARCHITECT-SYSTEM; docs015_data_flow_verification/004_builder_flow.md; docs017_legacy_agent_rules_images_design_keydocs/044_data_flow_trace_maintenance_work_items.md
- **relations:** unified work items (final form); loop action; site-planner; deployment-github.
- **verify-later:** which builder agent_definitions still exist and which have traffic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Sequential page generation (Phase 0 multipage fix)
- **category:** NEW:site-build-pipeline
- **status-signal:** superseded
- **status-evidence:** docs010/019 "Current Problem: Multipage-website-builder tries to generate 4 pages at once — race conditions... Solution: make it work like landing-page-builder: sequential, one page at a time"; two-delay spawn fix noted FIXED in docs009/003.
- **what:** Replacing parallel batch spawning with a strategist-planned page list iterated by the loop action (research → write per page), a wrap_multipage action generating navigation and collecting assets, and spawn timing fixed by double initialization delays. The stabilization step that made multipage builds reliable enough for everything after.
- **sources:** docs010_multitrack_flows_persona_architecture/019_start_here_document.md; docs009_site_interrogation_and_solutions/003_claude_save_point.md#Status
- **relations:** loop action; batched generation (predecessor); pageflow-builder (successor).
- **verify-later:** wrap_multipage action; spawn_actions.go delay logic.

<!-- SOURCE: U21_legacy_docs_b.md -->
### Selective rebuild via build_status
- **category:** NEW:site-build-pipeline
- **status-signal:** deployed
- **status-evidence:** docs018/004 ("get_pages_to_build filters by build_status IN ('planned','needs_rebuild')"); docs015/003 documents the stale-page trap ("If the site planner didn't include use-cases in the new plan... it shows in nav but has stale content").
- **what:** Two orthogonal page state columns — status (active/deleted/needs_attention lifecycle) and build_status (planned/needs_rebuild/deployed) — let rebuilds touch only marked pages. Known failure mode: pages absent from a new plan silently keep old content while remaining in nav. Ancestor of work-item-driven rebuild targeting.
- **sources:** docs018_rerendering/004_trigger_just_pages_that_need_rebuild.md; docs015_data_flow_verification/003_temp_doc_rebuild_flow.md; docs017_legacy_agent_rules_images_design_keydocs/041_page_rebuild_action.md
- **relations:** page-rebuild agent; maintenance_queue (proto-work-items); work items.
- **verify-later:** get_pages_to_build action; build_status usage today.

<!-- SOURCE: U15_docs019_running_notes.md -->
### Work-item relay / three-generation builder architecture
- **category:** NEW:site-build-orchestration-generations
- **status-signal:** partial
- **status-evidence:** "THREE generations coexist... GEN-3 component/spec/DB era = pageflow-builder v20 ACTIVE + site-work-orchestrator (queue-native sibling)... §B3 CLOSED: spine = the work-item relay" (NOTES_running_synthesis_v4(39).md, 2026-07-04).
- **what:** A builder-thread inventory found three coexisting generations of "build a site" orchestration on the platform (GEN-1 template era; GEN-2 in-memory multipage v1≈v2; GEN-3 component/spec/DB era) with ~8 overlapping top-level "build the site" orchestrators, only one of which (`pageflow-builder`) is the active monolith. Separately, a queue-native work-item relay (`domain-submitter → needs_domain_research → build-dispatch-loop → domain-research-classifier → needs_strategy → domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan → build-site-planner → needs_page/needs_content_page → page-build-handler`) was traced end-to-end via `reconcile_site_plan`'s routing table and confirmed to reach the builder NATIVELY — established as the real spine, with `pageflow-builder` demoted to "intake convenience." A commented-out `"tool"` route in the same routing table is the mechanism gap blocking tool/infographics pages from the relay.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B0" through "§B3 CLOSED" entries.
- **relations:** Roadmap-phase enforcement gap; adoption pipeline; vertical-exemplar-researcher hop; site-quality programme handoff.
- **verify-later:** `RUNBOOK_builder_route.md`, `load_work_item_actions.go` routing table, the un-consolidated GEN-1/2 legacy orchestrators (Q1/Q5 consolidation candidates, left open).

<!-- SOURCE: U15_docs019_running_notes.md -->
### Work-item relay / three-generation builder architecture
- **category:** NEW:site-build-orchestration-generations
- **status-signal:** partial
- **status-evidence:** "THREE generations coexist... GEN-3 component/spec/DB era = pageflow-builder v20 ACTIVE + site-work-orchestrator (queue-native sibling)... §B3 CLOSED: spine = the work-item relay" (NOTES_running_synthesis_v4(39).md, 2026-07-04).
- **what:** A builder-thread inventory found three coexisting generations of "build a site" orchestration on the platform (GEN-1 template era; GEN-2 in-memory multipage v1≈v2; GEN-3 component/spec/DB era) with ~8 overlapping top-level "build the site" orchestrators, only one of which (`pageflow-builder`) is the active monolith. Separately, a queue-native work-item relay (`domain-submitter → needs_domain_research → build-dispatch-loop → domain-research-classifier → needs_strategy → domain-strategist → needs_briefing → build-briefing-agent → needs_site_plan → build-site-planner → needs_page/needs_content_page → page-build-handler`) was traced end-to-end via `reconcile_site_plan`'s routing table and confirmed to reach the builder NATIVELY — established as the real spine, with `pageflow-builder` demoted to "intake convenience." A commented-out `"tool"` route in the same routing table is the mechanism gap blocking tool/infographics pages from the relay.
- **sources:** NOTES_running_synthesis_v4(39).md 2026-07-04 "§B0" through "§B3 CLOSED" entries.
- **relations:** Roadmap-phase enforcement gap; adoption pipeline; vertical-exemplar-researcher hop; site-quality programme handoff.
- **verify-later:** `RUNBOOK_builder_route.md`, `load_work_item_actions.go` routing table, the un-consolidated GEN-1/2 legacy orchestrators (Q1/Q5 consolidation candidates, left open).

<!-- SOURCE: U22_recent_small_docs.md -->
### Automated Go action build pipeline (compiler pod)
- **category:** NEW:action-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** "This is a medium-term investment"; whole doc is a design with a numbered ordered rollout, no deployment claim.
- **what:** A design for an in-cluster compiler pod that watches git for LLM-written Go action files, compiles the full chassis, runs tests, has a second-LLM review stage, builds an image via kaniko, and deploys per an HITL dial (manual→staging→auto) with rollback via recorded previous_tag. Uses an `action_build_jobs` job/audit table; git stays the source of truth, replacing GitHub Actions. Closes the loop: LLM identifies missing capability → writes action → compiled/tested/deployed → wires into workflow JSON.
- **sources:** docs020.../002_automated_go_action_create_and_build_pipeline.md
- **relations:** modular discovery-check registry (init() pattern), HITL, tool-lifecycle
- **verify-later:** action_build_jobs table; any compiler-service/ deployment

<!-- SOURCE: U22_recent_small_docs.md -->
### Automated Go action build pipeline (compiler pod)
- **category:** NEW:action-build-pipeline
- **status-signal:** aspirational
- **status-evidence:** "This is a medium-term investment"; whole doc is a design with a numbered ordered rollout, no deployment claim.
- **what:** A design for an in-cluster compiler pod that watches git for LLM-written Go action files, compiles the full chassis, runs tests, has a second-LLM review stage, builds an image via kaniko, and deploys per an HITL dial (manual→staging→auto) with rollback via recorded previous_tag. Uses an `action_build_jobs` job/audit table; git stays the source of truth, replacing GitHub Actions. Closes the loop: LLM identifies missing capability → writes action → compiled/tested/deployed → wires into workflow JSON.
- **sources:** docs020.../002_automated_go_action_create_and_build_pipeline.md
- **relations:** modular discovery-check registry (init() pattern), HITL, tool-lifecycle
- **verify-later:** action_build_jobs table; any compiler-service/ deployment
