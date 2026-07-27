# Register — site-plan-and-reconciler

> **covers-through: 2026-07-13** · extraction freeze.
> Subsystems that shipped after this date may be absent from this file
> **entirely** — absence here is not evidence of absence in the platform. See `bugs_open/106`.

42 concepts, consolidated from 65 raw extractions across units U01, U02, U03, U04, U05,
U09, U10, U12, U13, U14, U15, U17a, U18, U19, U20, U23, U24a, U24d, U25, U26.

### PLAN-001 — Plan as declarative artefact + reconciler (Kubernetes-style desired-vs-realised)
- **status:** partial
- **status-evidence:** 029/030 "Phase 0 lands today" (2026-05-04 re-adopt verified dedup); Phase 1 schema committed via migration 031/040 (`site_plans`, `site_plan_pages`, `site_plan_sections`, `site_plan_directives` — row-per-thing over JSONB for 1000+ page scale); tables confirmed live in production by later units (site_plan_imagery and work-item flows reference them).
- **what:** The structural fix for the two-writer duplicate-pages bug (adoption + site-planner both wrote `pages` rows with no shared identity space): the planner stops emitting work items directly and instead writes desired state to a separate, versioned plan-domain schema (one `is_current` plan per site); a deterministic Go reconciler (`reconcile_site_plan`, no LLM) diffs plan-vs-realised and emits idempotent `needs_page:<name>` items for the diff only, with preference weights/cycle budget/dependency ordering. Versioning mirrors `site_specs` (is_current + superseded_at). Phase 2 vision: discoverers/auditors read the plan for sharper fitness checks.
- **sources:** docs024 029_site_plan_and_reconciler (multiple copies, U01/U17a); docs024 030_phase1_plan_and_reconciler (U01/U17a/U19); 007_adoption_pipeline_v4.patch; sql_for_tables/040_site_plans_schema.sql (U19)
- **relations:** CanonicalisePage (PLAN-002); built_from_plan_version (PLAN-004); directive cascade (PLAN-005); site-build-pipeline lineage (BLD register, predecessor generations)
- **verify-later:** reconcile_site_plan action name and code; site_plans/site_plan_pages/site_plan_sections/site_plan_directives tables live; pages.built_from_plan_version populated

### PLAN-002 — CanonicalisePage + ValidateRoles: deterministic page-shape vocabulary
- **status:** deployed
- **status-evidence:** Phase 0 "landed" (FOCUS_planner_ignores_adopted_state); Part A `-index`→`section-index` rule "written, unit-tested green, and deployed", verified via the 2026-05-28 production run (hubs deployed as section-index at nested URLs).
- **what:** A single canonicalisation helper (`datahelpers/page_canonical.go`: `ValidateRoles` + `CanonicalisePage`) maps a `(role, slug, parent_section)` descriptor to a canonical `(name, url, page_type)` triple — index→`/index.html`; `<slug>.html` content; `<section>-index`→`/<section>/index.html`; `tool-<slug>`/guide role → nested URLs — called from both adoption and planner surfaces. Part A's Rule 2 promotes a name ending `-index` with a non-leaf role to `section-index` (guarded by `isLeafRole`), recovering the LLM's one reliable signal (naming) when url/parent are omitted. Part B (de-hardcode the tools/guides/games vertical vocabulary in `nestedRoleFromURL`) remains unscoped. The two role-normalisers (`normaliseRole` routing-collapsed vs `normalisePageType` flavour-preserving) are intentionally layered — merging them was explicitly withdrawn as wrong. Residual: validator emits generic `section-index`, losing blog-index/entity-directory "flavour" — undecided whether the component resolver needs it.
- **sources:** HANDOFF_2026-05-25; FOCUS_chrome_templates_and_page_shape.md#fix-2; ARCHITECTURAL_TENSIONS(3).md#Tension-1/2; WM/029_site_plan_and_reconciler(1).md#fix; WM/030(4).md#Q3
- **relations:** two-canonicalisation-surfaces divergence (PLAN-003); Architectural Tension #1/#2 (PLAN-006/007); page_type vocabulary gap (PLAN-010)
- **verify-later:** page_role_validator.go Rule 2 + isLeafRole; page_canonical.go guide case; nestedRoleFromURL hardcoded verticals

### PLAN-003 — Two canonicalisation-surfaces divergence: WriteSitePlanAction vs SyncPagesToDBAction
- **status:** deployed
- **status-evidence:** running_notes_14(25/26) Part 1-5, Part 8: "Both the core fix (sync no longer flattens hubs) and the companion (built_from_plan_version set) are confirmed. Thread closed" (2026-05-28).
- **what:** `WriteSitePlanAction` ran `ValidateRoles + CanonicalisePage` (correct `section-index` hubs), while `SyncPagesToDBAction` ran `CanonicalisePage` alone on the raw `page_plan`, producing flat rows (e.g. `games-index` typed `content` → `/games-index.html`) that `upsertPage`'s `ON CONFLICT` then used to overwrite the correct adoption-time row — one logical page, two writers, divergent results (including `tool-game-*` double-prefix duplicates). Option 1 (make sync read the already-validated `site_plan_pages`) was rejected/withdrawn because `pageflow-builder` and other callers invoke sync with no plan ever written, so it would silently break them. Shipped fix (Option 2): sync runs the identical `ValidateRoles` pipeline too, unifying all five callers. Exposed the deliberate guides de-prefix trade-off (plan de-prefixes `guide-rng-design`; sync now agrees, surfaced not silent).
- **sources:** running_notes_14(20/25/26) Part 1-3, Part 5; 016 §9 three linked entries; HANDOFF_2026-05-25#confirmed-root-cause; adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md
- **relations:** CanonicalisePage (PLAN-002); built_from_plan_version stamp (PLAN-004); Architectural Tension #2 (PLAN-007)
- **verify-later:** site_db_actions.go SyncPagesToDBAction normalisation loop; whether pageflow-builder was ever deprecated

### PLAN-004 — built_from_plan_version drift stamp + removal of the deployed→needs_rebuild flip (Option B)
- **status:** deployed
- **status-evidence:** "Option B shipped (two files, coupled)... confirmed in production" (running_notes_14 Part 8-12, CATALOGUE A1 fix list, 2026-06-03); 016 §9 dedicated entry (2026-05-28) "Fix shipped".
- **what:** `upsertPage`'s blunt `deployed→needs_rebuild` flip on every sync was a pre-design stand-in for the doc-029/030 intent — stamp `pages.built_from_plan_version` at build time and detect staleness in the reconciler — which had been explicitly deferred (HANDOFF_2026-05-07 #5). It over-fired (every deployed page, every sync) and mis-fired on pre-plan tool/game deploys. Completion: `UpdatePageStatusAction` stamps the current plan id at the deployed chokepoint; `upsertPage` COALESCE fill-if-null (never overwrite a real build version); the flip removed; drift now flows through the reconciler's `decideEmit`. `sites.last_reconciled_at` lets a scheduled tick skip recently-reconciled sites (no FK, so hard-deleted plans read as drift). Lesson recorded twice: a "bug" may be a half-implemented design — complete it, don't patch around it.
- **sources:** running_notes_14(25/26) Part 8-12; CATALOGUE_gamesdesign_post_sync_fix_defects(4/9)#family-a A1; old2/HANDOFF_2026-05-07(1)#5; sql_for_tables/040_site_plans_schema.sql#4-5
- **relations:** two-canonicalisation-surfaces divergence (PLAN-003); adoption-faithfulness convergence (PLAN-013); UpdatePageStatusAction zero-component guard (PBP register)
- **verify-later:** v3_site_actions.go UpdatePageStatusAction deployed branch; site_db_actions.go upsertPage CASE; scheduled reconciler tick existence

### PLAN-005 — Strategic vs plan-time guidance split: site_plan_directives cascade + brief renderer + HITL lock transfer
- **status:** partial
- **status-evidence:** 030 Q1/Q2 decisions committed; "Reconciler is documented in doc 030 but the chassis-side implementation has been landing in stages"; lock transfer "run inside write_site_plan" per doc 030, extended for imagery + lock_type/expiry per 2026-05 patches.
- **what:** `site_specs.design_intent/content_direction` stay strategic and slow-changing (classifier/adoption-owned); the planner's per-build guidance flattens into row-shaped `site_plan_directives` (scope site/page/section, category, subject, directive, source, Pattern-A locks). Consumers never read rows directly: a Go brief renderer (`datahelpers/page_brief.go`) cascades site→page→section and applies cardinality semantics (single-valued override at narrower scope; multi-valued accumulate), emitting short LLM-ready briefs — the pattern imagery/text/design guidance should all follow. On plan rebuild, locked directives from the previous current plan match new rows by composite key (scope, scope_ref, category, subject, ordering); HITL-edited text/locked_at/locked_by copy over (HITL wins); unmatched locks drop with a log; previous plan kept as history. One LLM call still produces structure+design+content together (coherence over a three-call split) — only the write targets changed.
- **sources:** 030#Q1/Q2; FOCUS_site_spec_vs_site_plan.md (directives + lock-transfer sections); sql_for_tables/040_site_plans_schema.sql#site_plan_directives; FOCUS_adoption_faithfulness_via_locks(2/5).md
- **relations:** plan as declarative artefact (PLAN-001); site_specs/site_plan two-layer ownership (PLAN-008); per-section briefs gap (PLAN-025)
- **verify-later:** datahelpers/page_brief.go existence and consumers; transferDirectiveLocks in write_site_plan action code; site_plan_directives populated in production

### PLAN-006 — Architectural Tension #1: infer-and-repair vs deterministic structure derivation
- **status:** partial
- **status-evidence:** "Status 2026-05-25: Tension #1 has a deployed partial fix (Part A — ValidateRoles -index rule), pending a clean production test."
- **what:** The pipeline took structural decisions (page role/type/URL) from LLM free-text labels then repaired them with starved, vertical-hardcoded heuristics, producing silent structural corruption (section hubs flattened to content). Resolution principle: derive structure deterministically from the LLM's one reliable signal — naming (`<section>-index` marks a hub, vertical-agnostically); schema-constrain generation to kill form errors (necessary but not sufficient); make fallback heuristics fail loud, never default to content. Explicit recommendation against a free parent-pointer tree (worst LLM reliability tier) — a leaf's section, if needed, is a constrained choice over the enumerated hub set.
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-1; HANDOFF_2026-05-26 (page_type re-type as an instance)
- **relations:** CanonicalisePage (PLAN-002); page_type vocabulary gap (PLAN-010); Tension #2 (PLAN-007)
- **verify-later:** ValidateRoles -index rule and de-hardcoded nestedRoleFromURL in page_role_validator.go

### PLAN-007 — Architectural Tension #2: page identity derived in multiple places that undo each other
- **status:** partial
- **status-evidence:** "Tension #2's residual confirmed cosmetic" (HANDOFF_2026-05-25); flavour-collapse residual "evidence-gated, not yet a code change" (2026-05-25).
- **what:** Adoption, planner-write and convergence each re-derive canonical page name/role/URL with no single owner, so a later stage can undo an earlier correct result (convergence preserved `games-index`; `WriteSitePlanAction` flattened it one step later). Principle: one canonical owner; canonicalisation idempotent on already-canonical input; downstream reads identity read-only. Part A made section indexes round-trip cleanly; the remaining residual is flavour collapse (validator emits generic `section-index`, losing blog-index/entity-directory flavour). Withdrawn idea: merging the two role-normalisers (intentionally layered, not a bug).
- **sources:** ARCHITECTURAL_TENSIONS(3).md#Tension-2; HANDOFF_2026-05-26 (write vs sync canonicaliser divergence)
- **relations:** Tension #1 (PLAN-006); two-canonicalisation-surfaces divergence (PLAN-003)
- **verify-later:** CanonicalisePage/normaliseRole/normalisePageType in datahelpers/page_canonical.go; component resolver's page_type dependence

### PLAN-008 — site_specs vs site_plan two-layer architecture + aspect ownership contract
- **status:** partial
- **status-evidence:** "build-site-planner workflow writes both shapes during transition (old site_specs/site_plan aspect AND new plan tables)" (undated FOCUS doc referencing docs 028-030).
- **what:** `site_specs` = strategic, brand-level, slow-changing, one owning agent per aspect (classifier owns identity/classification/content_direction/design_intent/seo/maintenance; adoption owns site_archetype/design_reference; strategist owns strategy; planner owns the four plan tables). `site_plan` tables = per-build, row-shaped, rebuilt per plan. Three ownership rules: don't read what you didn't spec; don't overwrite another's aspect; write outputs to the spec — with a classifier read-and-extend carve-out. Decision rules and anti-patterns govern where new data lives (specs vs directives vs sibling structured tables).
- **sources:** FOCUS_site_spec_vs_site_plan.md (whole); ASSESSMENT_imagery_phase_0_1…md#What-Phase-1-changes
- **relations:** directive cascade (PLAN-005); plan as declarative artefact (PLAN-001); lock transfer (PLAN-005)
- **verify-later:** site_plans/site_plan_pages/site_plan_sections/site_plan_directives tables; legacy site_plan aspect readers (pageflow-builder)

### PLAN-009 — Section-data deferral + reconciler (reconcile_section_data / needs_section_data)
- **status:** deployed
- **status-evidence:** HANDOFF (2026-06-19), corrected against an earlier stale note: "`reconcile_section_data` IS wired — registry.go line 914 … 'Re-trigger pages whose deferred section data is now query-resolvable'" — superseding the 2026-05-27 "not yet wired"/"decision: no dedicated handler agent needed" framing.
- **what:** A component's content comes from one of three sources with different fix shapes: (1) query-resolvable section data (tools/guides-list kind) — `ReconcileSectionDataAction` re-triggers pages whose deferred data has become resolvable via the `queryresolve` package (`pages_where_type`, later `pages_under_section` joining site_areas); (2) a human-entered spec field (e.g. pricing tiers from `site_specs.pricing`) — the reconciler correctly and permanently skips these, leaving them in HITL; (3) page-content-writer prose (LLM-generated). The originally-planned dedicated LLM handler agent and a never-built `directory-builder` agent were deliberately dropped in favour of this lightweight, non-LLM resolver, which rescans open items and emits `needs_page` re-renders (dedup key `page_rerender:<page>`), closing via `closeResolvedDataRequest`.
- **sources:** HANDOFF_idea_uk_differentiators_section_data.md; running_notes_scheme_to_components(55).md#Sa/#Sh; idea.uk/RUNBOOK_idea_uk_chassis_site_and_vm_deploy(25).md; FUTURE_section_data_handler_1_.md; FOCUS_internal_linking.md#4
- **relations:** needs_section_data classification (work-dispatch register WDS-E12); plan_sections deferral semantics (page-build-pipeline register)
- **verify-later:** reconcile_section_data_action.go scope logic; registry.go wiring; queryresolve.go dispatch switch; count of stuck needs_section_data items

### PLAN-010 — page_type vocabulary gap forcing game→tool re-type (Gap B)
- **status:** unknown
- **status-evidence:** "root cause confirmed from the planner's response_text … there is no `game` [in the Canonical Page Types list], so every adopted game is forced to `tool`" (2026-05-26); "OPEN structurally; may have been addressed by the other-chat fixes … Verify post-deploy."
- **what:** The `plan_site` prompt's closed page-type list lacks `game`; the LLM keeps names faithfully but re-types game pages as `tool`; canonicalisation's tool branch then renames, and a page_type change (not a name change) is what duplicates pages — 5 duplicate `game-*`/`tool-game-*` pairs observed on gamesdesign. Also exposed: `WriteSitePlanAction` and `sync_pages_to_db` canonicalise the same tool-typed page differently (`tool-auto-battler` vs `tool-game-auto-battler`).
- **sources:** HANDOFF_2026-05-26…md#diagnosis, #Where-to-resume
- **relations:** Architectural Tension #1 (PLAN-006); adoption faithfulness locks (PLAN-013)
- **verify-later:** run the three handoff verification queries on a post-2026-05-26 adoption; page_canonical.go call sites

### PLAN-011 — Planner re-plan union safety (normaliseRealisedToPlanPage)
- **status:** deployed
- **status-evidence:** Checkpoint 2026-07-07: "normaliseRealisedToPlanPage (v3_site_actions.go:4383) exists so a re-plan LOADS realised pages …, converts them to plan-page shape CARRYING their sections, and UNIONS with the LLM proposal — its own comment: without carrying sections the upsert would clobber built pages."
- **what:** Site composition is whole-plan and LLM-driven: `build-site-planner` (consuming `needs_site_plan`) supersedes the current `site_plans` row and rewrites `site_plan_pages` + `site_plan_sections`. Re-running it is safe by design because `load_existing_pages` surfaces realised pages and the normaliser unions them (with their sections) into the new plan — built pages keep their composition while catalogued-but-uncomposed pages get composed. This makes "emit needs_site_plan" the structural route for composing missing pages, versus hand-inserting plan rows (which drifts nav/plan/page consistency).
- **sources:** running_notes_scheme_to_components(55).md#Un; stepF_replan_read.sql
- **relations:** planned-but-uncomposed pages gap (PLAN-012); union-clobber bug (PLAN-016)
- **verify-later:** v3_site_actions.go normaliseRealisedToPlanPage; build-site-planner workflow steps

### PLAN-012 — Planned-but-uncomposed pages gap (catalogued, never composed)
- **status:** aspirational
- **status-evidence:** Checkpoint: "the three planned pages have NO site_plan_sections rows; their pages.sections = []. Catalogued, never composed" — the corrective replan-read staged but not confirmed run as of 2026-07-07.
- **what:** A distinct failure shape from empty-sections builds: `pages` rows exist with page_type and nav intent set (so navigation links to them and 404s), but they carry empty sections and no plan rows — the LLM plan behind the current `site_plans` row never included them. A naive `needs_page` emit would build an empty page; the correct route is two-phase — planner re-run composes them (union-safe per PLAN-011), then `needs_page` builds and deploys.
- **sources:** running_notes_scheme_to_components(55).md#Uk/#Ul/#Um/#Un; RUNBOOK_scheme_to_components(50).md#PLANNED-PAGES; stepD_and_pages_reads.sql
- **relations:** planner re-plan union safety (PLAN-011); navigation 404s; rebuild vs rerender (page-build-pipeline register)
- **verify-later:** idea.uk pages rows for the three; site_plan_sections presence; whether the needs_site_plan emit ran

### PLAN-013 — Adoption-faithfulness convergence (reconcilePlanWithRealised): sequential root-cause fixes
- **status:** deployed
- **status-evidence:** "VERIFIED RESOLVED on a clean run (2026-06-05 17:26Z, corr 6381cb13) … ZERO bare siblings … CONVERGED … the convergence/duplicate-page root cause … is RESOLVED."
- **what:** A designed subsystem making adoption re-plans faithful to already-realised (locked) pages was broken through two sequential, independently-diagnosed root causes before it ever worked: (1) an earlier live-inspection found `reconcilePlanWithRealised` gates on `rm["adoption_locked"]`, but the live `load_existing_pages` query never emitted that flag — lockedPages was always empty, so convergence always no-op'd (confirmed "INERT"; migrations 053/054 not yet applied at that point); (2) after 053/054 landed, `ValidateSitePlanAction` still asserted `existing_pages` as `[]interface{}` while `QueryDatabaseAction` returns `[]map[string]interface{}` — the type assertion always failed silently, so convergence still no-op'd for every site. Deterministic Go convergence (in `v3_site_actions.go`), gated on `adoption_locked` pages: Pass A unions LLM-omitted adopted pages into the plan (via `normaliseRealisedToPlanPage`), snaps back renames, Pass C/C2 dedups section-stem and item-topic sibling collisions, truncates preserving locked pages. Fix: type-switch both shapes + a count log so an empty set is never silent again.
- **sources:** running_notes_14(20/25/26)#part-14h-14n; adoption/running_notes_14_sync_fix_and_adoption_rerun(20).md Part 14h; FOCUS_adoption_faithfulness_via_locks(5).md
- **relations:** first-plan branch (PLAN-014); union-clobber bug (PLAN-016); bare-sibling duplicates (PLAN-017)
- **verify-later:** ValidateSitePlanAction extraction switch; reconcilePlanWithRealised counters in planner logs; current state of write_site_plan_action.go transferDirectiveLocks

### PLAN-014 — First-plan branch: "no current plan + pages exist ⇒ adopted pages"
- **status:** deployed
- **status-evidence:** "054 `load_existing_pages` — partially live. The query emits `adoption_locked` but only via the first-plan branch: CASE WHEN NOT EXISTS (current is_current plan for this site) THEN true" (2026-06-05 verified landed state).
- **what:** Deterministic detection of the faithful first pass: when `load_existing_pages` finds no current site_plan but pages exist, all existing pages are flagged `adoption_locked=true` (only ever true after adoption; from-scratch sites have no pages before the planner's own sync). Convergence keys off this flag; a re-adoption from a cleared DB (or retiring the current plan) makes any site a "first pass" deterministically.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#verified-landed-state; verify_readoption_fix.sql; running_notes_14(25)#part-14i
- **relations:** adoption-faithfulness convergence (PLAN-013)
- **verify-later:** live load_existing_pages SQL in build-site-planner def

### PLAN-015 — Planner ignores adopted state (generic-skeleton overlay)
- **status:** superseded
- **status-evidence:** Diagnosed 2026-05-19 ("build-site-planner independently generates a 9-page generic site skeleton that ignores the adopted pages"); addressed by the convergence work verified 2026-06-05 plus the "Existing Pages — ALREADY BUILT, PRESERVE EXACTLY" prompt block (v1.0.1047).
- **what:** Two confirmed mechanisms: (1) the planner planned from identity/archetype without reading realised state, inventing parallel pages; (2) ValidateRoles couldn't converge a childless plan (section-index promotion needs a child declaring ParentSection). Root cause per doc 029: two surfaces (adoption, planner) both write pages and queue work without a shared identity space. Residual after the prompt fix: it did not stop differently-slugged siblings (bare `economy-basics` beside `guide-economy-basics`) — that took Pass C2 (PLAN-017).
- **sources:** FOCUS_planner_ignores_adopted_state.md; running_notes_14(25)#part-14c-14e; migration_cleanup_bare_guide_duplicates.sql
- **relations:** plan as declarative artefact (PLAN-001); adoption-faithfulness convergence (PLAN-013); bare-sibling duplicates (PLAN-017)
- **verify-later:** plan_site prompt existing-pages block in live build-site-planner def; llm_call_log for planner runs

### PLAN-016 — Union-clobber bug and the carry fix (sections/meta_description/nav_order)
- **status:** deployed
- **status-evidence:** "on the first pass, every adopted page the LLM omitted was unioned with empty values and the sync clobbered its real sections/meta_description/nav_order to empty … verified on the 2026-06-05 clean run."
- **what:** Pass A's union originally emitted `sections: []` because the 054 query didn't select the fields, and `upsertPage`'s `ON CONFLICT … sections = EXCLUDED.sections` overwrote the adopted page's real values — the difference between a faithful first pass and one that wipes adopted content the LLM didn't re-list. Fix (both parts must land together): `load_existing_pages` SELECT adds the fields; `normaliseRealisedToPlanPage` carries them. Reframed the separate "empty hubs" defect: source hubs are populated; no separate hub-convergence step is needed for adopted sites.
- **sources:** FOCUS_adoption_faithfulness_via_locks(5).md#union-clobber; running_notes_14(25)#part-14i-14j; migration_load_existing_pages_carry_fields.sql
- **relations:** adoption-faithfulness convergence (PLAN-013); planner re-plan union safety (PLAN-011)
- **verify-later:** load_existing_pages SELECT column list; normaliseRealisedToPlanPage in v3_site_actions.go

### PLAN-017 — Bare-sibling duplicate pages (planner re-invents adopted topics)
- **status:** deployed
- **status-evidence:** "DECISIVE (llm_call_log plan_site @ 20:25:22): the planner WAS given the adopted guides and emitted economy-basics anyway → PROMPT-RULE gap" — shipped as Pass C2 and verified on the 2026-06-05 clean run; cleanup migration applied.
- **what:** The planner proposed bare `economy-basics` etc. beside adopted `guide-economy-basics` — a differently-slugged sibling the "preserve existing pages" prompt rule did not stop. Deterministic Go stem-dedup (Pass C2, reusing CanonicalisePage's prefix stripping) is the guarantee; a prompt stopgap was optional. The durable cleanup migration removes bare rows from the current plan (the reconciler would otherwise re-create them) and terminalises their work items.
- **sources:** running_notes_14(25)#part-14c-14e; migration_cleanup_bare_guide_duplicates.sql; FOCUS_adoption_faithfulness_via_locks(5).md#item-topic-sibling-dedup
- **relations:** planner ignores adopted state (PLAN-015); adoption-faithfulness convergence (PLAN-013)
- **verify-later:** itemStemOf/Pass C2 in v3_site_actions.go

### PLAN-018 — Adoption calls the canonicaliser + reconciler orphan pruning (unbuilt)
- **status:** aspirational
- **status-evidence:** "Adoption today doesn't go through this. It computes its own URL based on page_type only … The reconciler … doesn't prune orphans. That's a follow-up" (FOCUS_chrome_templates Fix 2).
- **what:** Adoption's local URL computation (flat `/games.html` etc.) diverges from the canonicaliser the planner uses, producing duplicate logical pages that ON CONFLICT can't match. Proposed: `apply_adoption_plan` calls `CanonicalisePage`; reconciler gains an orphan-pruning pass (pages absent from the current plan get archived/soft-deleted). Partially overtaken by the convergence work (unions/dedups at plan time), but orphan pruning remains unbuilt — orphaned bare pages persisted after Pass C2 dropped them from the plan and needed manual cleanup.
- **sources:** FOCUS_chrome_templates_and_page_shape.md#fix-2; running_notes_14(25)#part-14l follow-up
- **relations:** CanonicalisePage (PLAN-002); bare-sibling cleanup migration (PLAN-017)
- **verify-later:** apply_adoption_plan URL computation today; any reconciler pruning logic

### PLAN-019 — Deferred plumbing stubs: scheduled reconciler tick, domain-aware ensure_pages
- **status:** aspirational
- **status-evidence:** "Scheduled reconciler tick — Not built. Reconciler currently fires only when called by the planner … ensure_pages should be domain-aware — Currently hardcoded in workflow JSON" (HANDOFF_2026-05-07(1)). Conflicting later reference: an emit_design guard rationale mentions "the scheduled reconcile tick does not backfill" as if one exists — status conflict flagged for stage 2.
- **what:** Two small deferred Phase-1 items: a heartbeat `scheduled_tasks` row producing periodic reconcile passes (mirroring content-feed-trigger), and moving the hardcoded `ensure_pages` page list into strategist/briefing-written site_specs read at plan time.
- **sources:** old2/HANDOFF_2026-05-07(1)#6-7; FOCUS_design_composition_flow_and_adoption_fidelity(1).md#3A
- **relations:** plan as declarative artefact (PLAN-001); build-pipeline-trigger cadence (build-pipeline register)
- **verify-later:** scheduled_tasks for a reconcile tick; ensure_pages config source; resolve the built-vs-not-built status conflict

### PLAN-020 — site_plan page-role enum naming (underscore → hyphen; index → landing)
- **status:** superseded
- **status-evidence:** Archive: `"section_index" | ... | "blog_post"`; live: `"section-index" | ... | "blog-post" | "landing"`.
- **what:** `site_plan_pages.role` vocabulary was originally underscore-separated with a bare `index` role for the homepage. Renamed to hyphenated form and the homepage role renamed to `landing`, matching kebab-case conventions elsewhere.
- **sources:** old/029_site_plan_and_reconciler.md#"role table"; docs024_key_docs_latest/029_site_plan_and_reconciler(2).md#"role table"
- **relations:** page_type vocabulary and kebab constraint (016 §6.5)
- **verify-later:** DB check constraint on site_plan_pages.role/pages.page_type for hyphenated values

### PLAN-021 — site_plan_partials: single JSONB-blob partial storage (abandoned)
- **status:** abandoned
- **status-evidence:** Live doc: "JSONB blobs were considered and rejected because at anticipated scale ... loading whole blobs to read one slice is wasteful, surgical HITL edits become hard, and lock transfer at meaningful granularity is impossible." First draft of migration 031/040 defined it; the second draft in the same migration file replaces it.
- **what:** Archived Phase 1 plan proposed one table, `site_plan_partials`, storing each partial (design_direction, content_strategy eager; `page_brief:<name>` lazy) as a single versioned JSONB blob per plan. Abandoned for two normalized row-per-thing tables — `site_plan_sections` and `site_plan_directives` — enabling per-row HITL locking at 1000+ page scale.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"schema section"; sql_for_tables/040_site_plans_schema.sql#site_plan_partials
- **relations:** lock transfer across plan rebuilds (PLAN-005); lazy per-page brief generation (PLAN-024, also abandoned)
- **verify-later:** confirm site_plan_directives/site_plan_sections tables exist, site_plan_partials does not

### PLAN-022 — Three sequential per-partial plan-builder LLM calls (abandoned)
- **status:** abandoned
- **status-evidence:** Live: "Earlier draft of this doc proposed three sequential LLM calls. Looking at the existing build-site-planner agent, that lean was wrong."
- **what:** Archived plan proposed splitting the plan-builder into three sequential LLM calls for independent retry granularity. Abandoned once it was noticed the production build-site-planner agent already produces all three (structure+design+content) coherently in one call with no evidence of retry-granularity problems.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q2. Plan-builder LLM tier"
- **relations:** site_plan_partials (PLAN-021)
- **verify-later:** build-site-planner agent_definitions workflow — confirm single LLM call shape

### PLAN-023 — Separate BuildPageURL path-resolver helper (abandoned)
- **status:** abandoned
- **status-evidence:** Live: "The earlier draft of this doc proposed a separate BuildPageURL helper... That argument was overly cautious... Consolidated."
- **what:** Archived plan proposed a brand-new ~50-line Go helper sibling to `page_canonical.go`. Abandoned as overly cautious: Phase 1 instead extends `CanonicalisePage` additively with an optional `ParentSection` field.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"Q3. URL paths"
- **relations:** site_plan page-role enum naming (PLAN-020); CanonicalisePage (PLAN-002)
- **verify-later:** datahelpers/page_canonical.go — confirm CanonicalisePage has ParentSection, no separate BuildPageURL

### PLAN-024 — Lazy per-page brief generation via build_page_brief step (abandoned; replaced by deterministic Go brief renderer)
- **status:** abandoned
- **status-evidence:** Archive rollout step 8: "build_page_brief step in page-build-handler... generates site_plan_partials/page_brief:<name> if missing." Live doc replaces with a pure-Go brief renderer; 029 B-029-2 had promoted the LLM-lazy version to a Phase-1 acceptance test before the design changed.
- **what:** Archived plan generated each page's brief lazily via an LLM step during page build, motivated by a no-empty-slots acceptance test (component-author defaults leaking — empty img src, /services.html CTAs on sites without services). Abandoned for a deterministic, non-LLM Go helper (PLAN-005's brief renderer) that assembles a brief at read time by walking the directive cascade and applying cardinality rules.
- **sources:** old/FOCUS_ADOPTION_030_phase1_plan_and_reconciler.md#"rollout table, step 7-8"; 029#B-029-2
- **relations:** directive cascade + brief renderer (PLAN-005); site_plan_partials (PLAN-021); per-section briefs gap (PLAN-025)
- **verify-later:** confirm datahelpers/page_brief.go exists; page-build-handler has no build_page_brief LLM step

### PLAN-025 — Per-section briefs gap (planner depth): bare section-name strings, no intent
- **status:** aspirational
- **status-evidence:** TODO_remaining_work.md: "Planner depth: per-section briefs + stale-plan write-back... planner needs to emit them" listed under "Open — structural (not blocking)."
- **what:** `site_plan.pages[].sections` is an array of bare strings with no per-section brief — the deeper structural cause behind the FAQ duplicate-content-surface defect (PLAN-026): without a brief, the writer cannot tell that `faq` and `generic-text-block` are competing surfaces for the same content. A consumer already exists (`plan_sections.sectionDescription`) but the planner never emits any of those shapes. Token-budget caveat: briefing every section on a large site materially grows planner output size.
- **sources:** FOCUS_faq_empty_items_and_page_content.md#Fix-C-stale-plan; site_planner_depth_and_freshness_concerns.md
- **relations:** FAQ duplicate content-surface bug (PLAN-026); lazy brief generation (PLAN-024, abandoned alternative); validate_components resolver (PLAN-027)
- **verify-later:** load_page_sections_from_spec; plan_sections.sectionDescription resolver

### PLAN-026 — FAQ empty-items bug: duplicate content-surface planning (Defect 1)
- **status:** deployed
- **status-evidence:** "Deployed status (2026-05-21) ... Prevention shipped on three fronts, all live", chassis v1.0.1029.
- **what:** Pages were planned with both a freeform `generic-text-block` and a structured component (`faq`, `pricing`) intended to hold the same content, because the content-gap-planner's prompt example hardcoded `generic-text-block` and the site-planner's mappings omitted faq/pricing entirely; the content writer then filled the freeform block and left the structured component empty. Fixed by editing both planner prompts and an archetype-aware `defaultSectionsForPage` Go backstop.
- **sources:** FOCUS_faq_empty_items_and_page_content.md; 016_debugging_guide_addenda.md#empty-shells; faq_empty_items_prevention_findings(1).md
- **relations:** display-name leak / validate_components resolver (PLAN-027); per-section briefs gap (PLAN-025)
- **verify-later:** content-gap-planner and site-planner agent_definitions prompt_template; apply_gap_plan_action.go defaultSectionsForPage; chassis v1.0.1029

### PLAN-027 — Display-name leak into section arrays + validate_components resolver
- **status:** deployed
- **status-evidence:** "validate_components implemented in ValidateSitePlanAction (was a dead flag)... deployed in chassis v1.0.1029." (An earlier framing of the same mechanism described `validate_components: true` as "currently set true for site-planner but never read" and not yet live — the dated v1.0.1029 deployment claim is the more specific/later evidence and is preferred.)
- **what:** A planner path could emit a component's `display_name` instead of its kebab `function` into a page's `sections` array, orphaning the page_component. Fixed by implementing the previously-dead `validate_components` config flag in `ValidateSitePlanAction`: a `componentNameResolver` resolves each section name (exact match → NormalizeComponentFunction → display/name lookup → drop+log if unresolvable). Deliberately narrow — it does not deduplicate or make intent decisions (that's the planner prompt + per-section briefs). The gap-planner path (`applyNewPage`) doesn't route through `validate_site_plan`, so the resolver had to be wired in separately there too.
- **sources:** FOCUS_faq_empty_items_and_page_content.md#Fix-B-implementation; validate_components_implementation(1).md#scope
- **relations:** FAQ duplicate content-surface bug (PLAN-026); per-section briefs gap (PLAN-025)
- **verify-later:** ValidateSitePlanAction validate_components flag read; loadComponentNameResolver/NormalizeComponentFunction; apply_gap_plan_action.go applyNewPage

### PLAN-028 — Stale site_plan: gap-planned pages never written back (Concern 2)
- **status:** aspirational
- **status-evidence:** TODO_remaining_work.md: "gap-planned pages aren't written back to site_plan (faq was absent from the plan entirely)... apply_gap_plan should append new pages to site_plan" — not yet implemented.
- **what:** Pages created after initial site planning get a `pages` row and nav entries but are never appended to `site_specs.site_plan`; the plan drifts from reality with every gap-added page. Proposed fix: `apply_gap_plan` deep-merges the new page into `site_specs.site_plan` (mirroring `enrich_news_feed`'s pattern), plus a periodic plan-reconciliation discovery check.
- **sources:** 016_debugging_guide_addenda.md#stale-plan; site_planner_depth_and_freshness_concerns.md
- **relations:** per-section briefs gap (PLAN-025); page content-creation build pipeline trace
- **verify-later:** apply_gap_plan_action.go; enrich_news_feed deep-merge pattern

### PLAN-029 — site_plan as authoritative build source, overwriting pages.sections
- **status:** deployed
- **status-evidence:** "`load_page_sections_from_spec_action.go` ... CONFIRMED in code" (PLAN_tool_widget_clobber(9).md §2.4).
- **what:** The page-build pipeline's section authority is `site_specs.site_plan`, not `pages.sections` directly — the loader syncs the plan's sections back into `pages.sections` on every build where a plan entry exists, only falling back to `pages.sections` if the plan yields nothing. Consequence: any fix that only sets `pages.sections` inside a tool action is futile once a plan entry exists; a durable fix must add the tool/embed section to the planner's `site_plan` output itself.
- **sources:** tools/tool_widget_clobber/PLAN_tool_widget_clobber(9).md#2.4; NOTES_running_tool_widget_investigation.md#Phase-2
- **relations:** three section sources for a page build (PLAN-038); tool widget clobber mechanism
- **verify-later:** load_page_sections_from_spec_action.go; whether site_plan now carries a tool/embed section entry for page_type='tool' pages

### PLAN-030 — queryresolve reality-vs-invention architectural promise
- **status:** deployed
- **status-evidence:** Stated as an existing architectural line: "queryresolve exists specifically to draw a line between content the LLM is allowed to write... and content that has a database answer."
- **what:** A specific agent responsibility (`queryresolve`) enforcing a hard boundary between LLM-authored creative content and database-derived factual lists — framed as central to the platform's "avoid fabrication" mission, alongside carving the build into specialists with non-overlapping responsibilities.
- **sources:** pitch/003thebiggerpicture.md
- **relations:** section-data deferral + reconciler (PLAN-009); fractal agent architecture claim
- **verify-later:** queryresolve action implementation; source: query.* convention in page_components

### PLAN-031 — New-domain build pipeline stage chain (domain-submitter → page-build-handler)
- **status:** partial
- **status-evidence:** Traced live from code/DB snapshot: "Confirmed: ReconcileSitePlanAction reconciles pages only" and "The chain is fully connected" — caveated as read from a 2026-05-21 DB backup snapshot, "may have drifted."
- **what:** The confirmed happy-path chain for building a brand-new domain: `domain-submitter` → `domain-research-classifier` → `domain-strategist` → `build-briefing-agent` → `build-site-planner` (plan_site → validate → write_site_plan → sync_pages → populate_nav → reconcile_site_plan) → `page-build-handler` per page → `rerender-pages`, driven by the 30s `build-pipeline-trigger` heartbeat, with every stage's `create_work_item` defaulting to status `triaged` so the pipeline self-advances.
- **sources:** plainjanedomain/README.md
- **relations:** work-item relay spine (build-pipeline register); design/composition emission gap (PLAN-032)
- **verify-later:** live SELECT type, status, image_tag FROM agent_definitions for the named stages

### PLAN-032 — Design/composition work-item emission gap (planner reorg unclosed seam)
- **status:** deployed
- **status-evidence:** "So nothing in the build path appears to emit a needs_design/needs_composition trigger for a fresh domain... consistent with this being an unclosed seam from the planner reorg."
- **stage2-verified (2026-07-14):** unknown → deployed — platform/orchestration/actions/emit_design_items_action.go now exists, explicitly built to close this exact gap (header comment: 'build-site-planner (Phase 1) moved its terminal step to write_site_plan + reconcile_site_plan, neither of which emits the design trigger. This action restores it'). Registered as 'emit_de...
- **what:** A discovered structural risk: the legacy `WriteBuildItemsAction` emitted the full item set for a new build (`needs_page`, `needs_logo`/`needs_hero_image`, `needs_composition`, `needs_design`), but the Phase-1 replacement (`build-site-planner` → `write_site_plan` + `reconcile_site_plan`) emits only `needs_page` + `needs_rerender`. The only fallback is the improvement-loop's `design-discovery-agent` catching `missing_css` later — meaning a new site could deploy pages referencing a stylesheet that doesn't exist yet.
- **sources:** plainjanedomain/README.md
- **relations:** new-domain build pipeline stage chain (PLAN-031); site-chrome rendering gap (dartsonline), same class of defect
- **verify-later:** ReconcileSitePlanAction, WriteSitePlanAction, WriteBuildItemsAction Go source; design-discovery-agent missing_css check

### PLAN-033 — Roadmap-phases scope decision gap (nav grounded in built reality)
- **status:** partial
- **status-evidence:** "VERIFIED IN CODE: 082_submit_domain_unified.sh — grep confirms ONLY --mission/--mission-file exist, no --roadmap anywhere. build-site-planner prompt — the {{if .roadmap_brief}}...{{end}} block has NO else" (2026-07-07, promoted platform-wide from a single-site fix).
- **what:** No submitted site ever gets a roadmap/phases decision: the submit script has no `--roadmap` path and the planner's phase-discipline instructions vanish (not degrade) without one — so commerce-shaped domains get aspirational full plans and nav links to unbuildable pages. Fix shape (relay-wide, by construction): a post-classification scope-decision hop writes a phased `roadmap_brief` (P1 content/guides/tools; P2 legal-gated affiliate; P3 catalogue); planner prompt gains the ELSE branch (default phase-1-only or HITL hold); nav generation grounded in the BUILT set regardless of plan. Guidelines 001 already define the roadmap/phases mechanism — the docs had it, intake didn't.
- **sources:** docs019/RUNBOOK_builder_route(21).md#queue (item 6); NOTES_running_fixloop(9).md; NOTES_running_synthesis_v4(39).md 2026-07-07
- **relations:** F0 guides pilot; coverage baseline (build-pipeline register); council compliance reviewer
- **verify-later:** 082_submit_domain_unified.sh flags; build-site-planner roadmap_brief template block; nav-updater

### PLAN-034 — site-planner (v2, single-LLM-call site plan)
- **status:** superseded
- **status-evidence:** 022 shows model flip-flops (sonnet→haiku for cost, 040 haiku→sonnet because planning is "high-leverage"); 053 build-site-planner is the successor for work-item builds.
- **what:** v2 planner: one LLM call over brief + component library + style collections producing validated_plan, pages, style_collection, needs_logo/needs_images. The model-choice oscillation (cost vs quality on high-leverage decisions) is documented reasoning worth keeping.
- **sources:** 022_site_planner.sql; sql_for_agents_v2/022_site_planner.sql; 040_optimise_which_llms.sql
- **relations:** chief-strategist (predecessor); build-site-planner (PLAN-035, successor); pageflow-builder (caller, build-pipeline register)
- **verify-later:** which planner the live pipelines invoke

### PLAN-035 — build-site-planner + roadmap-overrides-components rule
- **status:** deployed
- **status-evidence:** 053 shows the workflow rewired to the site_plans domain ("write_site_plan → sync_pages → populate_nav → reconcile_site_plan → complete"); plan_site runs on claude-opus-4-6; 067 adds thinking budget.
- **what:** Handler for `needs_site_plan`. Reads site_specs (identity/classification/briefing/strategy), loads component library and style collections, plans via LLM, validates, then writes into the site_plans domain and reconciles. Carries the ROADMAP OVERRIDE rule verbatim: "Build ONLY the pages listed in the current phase... use EXACTLY the section_types listed — even if they do not appear in the Available Section Components list... The roadmap is the authority for this site." Earlier form wrote plan/design_intent/content_direction specs + write_build_items.
- **sources:** 053_build_site_planner.sql; 108_site_plan_pages.sql; 067_implement_extended_thinking_not_yet_implemented.sql
- **relations:** site_plans/reconciler domain (PLAN-001); roadmap-phases scope decision gap (PLAN-033); component selector needs_new_component items
- **verify-later:** write_site_plan + reconcile_site_plan actions; roadmap aspect producer

### PLAN-036 — site_plan_pages schema repair (plan-domain drift)
- **status:** deployed
- **status-evidence:** 108 "Migration 033: Reconcile site_plan_pages columns + drop orphan site_plan_partials... every write_site_plan call to date has failed at the title-column error."
- **what:** Repairs drift between two drafts of the site-plan schema: adds title/meta_description/nav_label columns, drops page_data and the unused site_plan_partials table (directives are row-per-directive in site_plan_directives). Documents the `CREATE TABLE IF NOT EXISTS` silent-skip failure mode when a rewritten migration follows an applied earlier draft.
- **sources:** 108_site_plan_pages.sql
- **relations:** build-site-planner (PLAN-035); site_plan_partials (PLAN-021, abandoned); migration-discipline concepts
- **verify-later:** live \d site_plan_pages / site_plan_directives

### PLAN-037 — Multi-page site support (wrap_multipage, multipage-site-builder)
- **status:** superseded
- **status-evidence:** 030 SQL creates multipage-site-builder (index/about/contact + privacy); 031 shows the wrap_multipage step after html_assembler; today's pages/site_plans domain is the successor.
- **what:** Extending the single-page pipeline to small multi-page sites: after assembly, a `wrap_multipage` action derives index/about/contact (and privacy) pages, and the deployer commits all files. The first step from "landing page generator" toward the current multi-page site model.
- **sources:** docs004_website_capture_project/007different_types_of_site/030_about_page_and_privacy.sql; 031_about_page_multipage_site.md
- **relations:** successor: site_plans/pages domain (PLAN-001); builder generation lineage (build-pipeline register)
- **verify-later:** wrap_multipage in registry

### PLAN-038 — Three section sources for a page build + plan storage triple shape (029 Q1)
- **status:** superseded
- **status-evidence:** Workflow dump + code read 2026-07-06: "load_spec_sections... reads site_specs aspect site_plan (AUTHORITATIVE) → fallback page_record.sections. The site_plan_sections TABLE is NOT read by this path." Design doc 029's Q1 ("site_specs aspects vs new table") remains its own open question as of the same date.
- **stage2-verified (2026-07-14):** partial → superseded — load_page_sections_from_spec_action.go (current tree, last touched by commit 9255620 on 2026-07-07, present on HEAD of branch 085_debug_and_feature_loops) now reads site_plan_sections FIRST as 'authoritative' (header comment + code lines 109-130), contradicting the doc's central claim 'The site_plan_sections TABLE i...
- **what:** Page builds resolve their section list, in order, from: the `site_specs` aspect `site_plan` (legacy blob, ~5 sites carry one), `pages.sections` (jsonb fallback — what actually serves most sites; the newer planner dual-writes plan tables → pages.sections), and same-role sibling synthesis; the `site_plan_sections` table (written by the newer planner generation) is NOT read by this build path at all. Three peer stores with unclear precedence caused ten silent no-op builds and two fixes landing in the wrong store before the pages.sections UPDATE that finally unblocked. Operational corollary: `reconcile_site_plan` re-emits `needs_page` for any planned-but-unbuilt page every run (a standing trap for pages parked pending a decision).
- **sources:** docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-06-mechanism-fully-confirmed; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§3/§9.7; docs/social001_vonc_tiktok_social/minilobby_task/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md
- **relations:** plan storage authority (PLAN-039); complete_error silent no-ops (page-build-pipeline register); site_plan as authoritative build source (PLAN-029)
- **verify-later:** load_page_sections_from_spec_action.go source order; SELECT DISTINCT aspect FROM site_specs WHERE aspect LIKE 'site_plan%'; reconcile_site_plan grep

### PLAN-039 — Plan storage authority: 029 Q1 and the withdrawn table-first alteration
- **status:** superseded
- **status-evidence:** "SUPERSEDED (2026-07-06, same day) — decision deferred to 029 Q1; alteration WITHDRAWN"; "Decision closed (2026-07-07): the user chose REVERT."
- **what:** After the silent no-ops (PLAN-038), a decision was made (then withdrawn the same day) to make the `site_plans` family authoritative and alter `load_page_sections_from_spec` to read `site_plan_sections` first. Design doc 029's Q1 ("site_specs aspects vs new table") remained genuinely open; three shapes coexist in production (legacy blob aspect ×5 sites; 029's partitioned aspects apparently unimplemented; the newer-generation tables with pages.sections dual-write). The alteration was reverted (cluster reverts on next chassis push); the table path existing in production post-dates the original design lean, which is evidence for Q1 but doesn't resolve it.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#decision/#superseded; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-07-alteration-withdrawn/#revert-decision
- **relations:** three section sources (PLAN-038); plan as declarative artefact (PLAN-001)
- **verify-later:** git history of load_page_sections_from_spec_action.go (reverted?); docs024 029 doc Q1 status

### PLAN-040 — Planner role-aware ≥1-section invariant + role→pipeline mapping
- **status:** aspirational
- **status-evidence:** "Invariant refined: every planned page whose ROLE is built by page-build-handler must have ≥1 section" (Gate B, 2026-07-06) — nowhere claimed built.
- **what:** A planner run emitted all pages for a site but skipped SECTIONS for exactly the two non-standard roles — `blog-post` (legitimate: the blog pipeline builds those) and `section-index` (the defect that caused a 404). Prevention: at plan-store time, every planned page whose role page-build-handler owns must have ≥1 section, with the role→pipeline mapping made explicit; plus an auditor drift rule (pages.sections vs current plan) and post-deploy URL-presence checks per active page.
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#gate-results; docs/HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9
- **relations:** complete_error family (page-build-pipeline register); section descriptor design (PLAN-041)
- **verify-later:** site-planner agent_definition; site_plan_pages roles for recent sites

### PLAN-041 — Autonomous section composition: per-section descriptor {role, kind, data_feed}
- **status:** aspirational
- **status-evidence:** PLAN_dynamic_sections_and_loaders(4) status "DESIGN"; gaps list "(1) Section descriptor... Without this the framework can't tell static from dynamic" — none of gaps 1-5 marked built.
- **what:** The framework (not a human) should decide, from the domain/site-spec, which sections a page has, each section's role (to prevent overlaps, e.g. a mini-lobby vs a lobby-grid), whether it is static (build-time content) or dynamic (runtime-filled from a feed), and which named feed — encoded as a per-section descriptor `{component_name, role, kind, data_feed}` on the plan, written by the site-planner, consumed by build AND maintenance flows. The plan not carrying `kind` is why an assembler previously dropped runtime-filled shells. Includes a spec-level feed catalogue and quality-auditor maintenance detections (dropped-dynamic, overlap, deferral, empty-dynamic).
- **sources:** docs/PLAN_dynamic_sections_and_loaders(4).md#the-question/#structural-gaps; docs/RUNNING_NOTES_vonc_v2(28).md#2026-07-04-reframed
- **relations:** ≥1-section invariant (PLAN-040); plan storage authority (PLAN-039)
- **verify-later:** site_plan_sections columns (kind/data_feed/role exist?); site-planner prompt/workflow

### PLAN-042 — Website-builder agent group (six-specialist pipeline)
- **status:** superseded
- **status-evidence:** Ran in production per basic_usage/001 (migrations 005/007/009 referenced); the current platform builds sites via the site_plans domain / webdesign-agent pipeline (docs 029/030), which replaced this group.
- **what:** The original end-to-end website creation flow: an orchestrator agent calls domain-analyst (business categorisation via web-search) → site-architect (page structure, pausing for human approval) → fan-out of content-researcher + visual-designer (image search/generation, logo) → html-developer (per-page vanilla HTML/CSS fan-out) → site-publisher (s3_upload, preview URL). Seeded as agent_definitions + an agent_groups row; triggered by one spawn_group Kafka message.
- **sources:** docs/architecture/027-create-website-creation-system; docs/basic_usage/001basic_usage.txt
- **relations:** superseded by site_plans + webdesign-agent + design-composition pipeline; builder generation lineage (build-pipeline register)
- **verify-later:** migrations 005/007/009 in platform/database/migrations/; whether group still seeded
