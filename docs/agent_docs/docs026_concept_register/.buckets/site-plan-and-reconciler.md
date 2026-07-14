
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
