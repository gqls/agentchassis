# Register — imagery

64 concepts, consolidated from 92 raw extractions across units U01, U02, U03, U05, U09, U10, U18, U19, U20, U21, U22, U23, U24a, U25.

> Note on source duplication: the cluster input file for this unit contained the *entire* raw block set duplicated verbatim (same `SOURCE` tags, byte-identical text, appearing twice back-to-back) — a mechanical artefact of cluster assembly, not two independent extractions. Those exact duplicate pairs were collapsed before any semantic merging; the raw-block counts below and in the final report treat each such pair as one raw block. Genuine cross-unit duplication (the same mechanism independently described by 2-6 different extraction units) is heavy on this topic as expected and is merged per-concept below.

### IMG-001 — Imagery loop-closure master plan (Phases 0–6)
- **status:** partial
- **status-evidence:** Most complete snapshot (U10, 2026-05-14): "Phase 2G + Phase 2H are operationally verified end-to-end on robot-hands.com … Phases 3, 4, 5, 6 of the outer plan not started." Superseded in spirit from 2026-07-08 by the best-in-class programme (IMG-003), which renumbered phases to avoid collision.
- **what:** The sequenced master plan (`PLAN_imagery_loop_closure.md`) for closing the gap between what a site's imagery spec/plan asks for and what actually gets delivered: Phase 0 wires unread data (imagery_direction, origin_model) into generation; Phase 1 adds algorithmic no-LLM discovery checks; Phases 2A-2H are a schema + pipeline refactor (asset locking, asset_key multi-image identity, structured `site_plan_imagery` plan rows, richer request shape); Phase 3 is the adoption image mirror; Phase 4 gives the text-only visual auditor imagery awareness; Phase 5 is the vision-capable LLM path; Phase 6 is a dedicated `imagery-quality-auditor` agent. Each phase is independently shippable; the LLM-cost phases (4-6) are explicitly gated on the algorithmic checks working first. Decisions locked early: a separate auditor agent (not extending the existing one), max 2 regen attempts, asset locking mirrors `page_components` exactly, per-section imagery granularity initially deferred (later reversed, see IMG-012/014).
- **sources:** PLAN_imagery_loop_closure.md (whole, unit U10); PLAN_imagery_loop_closure.md#part-1/#part-2 (unit U09); PLAN_imagery_loop_closure.md + ASSESSMENT_imagery_phase_0_1_vs_phase_1_architecture.md (unit U02); STATUS_imagery_2026-05-12.md#At-a-glance
- **relations:** every Phase-N concept below (IMG-005 through IMG-024); superseded-in-spirit by IMG-003 (best-in-class programme); IMG-002 (pre-plan assessment it responds to)
- **verify-later:** phases table vs live code: `platform/orchestration/actions/generate_image_actions.go`, `discovery_checks/`, `assets` schema, image-build-handler agent_definitions row

### IMG-002 — Imagery subsystem pre-plan assessment (single hardcoded adapter baseline)
- **status:** superseded
- **status-evidence:** Descriptive baseline (Part 1 of PLAN_imagery_loop_closure.md) written before the phased plan; every gap it lists is tracked as its own phase concept above/below.
- **what:** The as-is snapshot that motivated IMG-001: one `DynamicImageAdapter` hardcoded to Stability SDXL (fixed request body — no negative prompt, seed, img2img, LoRA, variants); `assets` UNIQUE(site_id, purpose) blocking multi-image purposes; the planner asking for exactly `{logo, hero_home}`; components declaring only `hero_image`; features' `icon` strings rendering nowhere; misleading unread image-adapter config files; two image-generator agent rows (one a placeholder).
- **sources:** PLAN_imagery_loop_closure.md#part-1, old2/PLAN_imagery_loop_closure(1).md
- **relations:** IMG-001 (the plan this baseline motivated); IMG-009 (asset_key fixes the purpose-uniqueness gap); IMG-021 (adoption mirror fixes crawled-imagery loss); IMG-050 (per-vertical LoRA)
- **verify-later:** internal/adapters/imagegenerator/dynamic_adapter.go; assets unique indexes; ImagePurposes map

### IMG-003 — Imagery best-in-class programme (G1–G9, D1–D8, phases I0–I8)
- **status:** partial
- **status-evidence:** HANDOFF 2026-07-12: "Phase I0 ✅ COMPLETE… Phase I1 ✅ COMPLETE, LIVE-VERIFIED… Phase I2 ⏳ IN PROGRESS… Phases I3–I8 not started."
- **what:** The 2026-07-08 successor programme raising fleet visual quality to best-in-class, superseding the loop-closure programme's later phases. Nine goals (brand kit/logo permanence, data-accurate infographics, content-linked card imagery, graphic artefacts/sprites, copyright-safe product sketches, news imagery, performance budgets, accessibility/OG surface, quality loop) governed by eight user-confirmed design decisions: D1 code-rendered charts (never diffusion), D2 two lanes (plan-driven vs content-driven), D3 kind batches as text+CHECK constraint, D4 brand guide as data, D5 logo lock, D6 dual icon strategy, D7 product-sketch constraints, D8 deploy-enforced performance budgets. Phases I0-I8 are each acceptance-gated on robot-hands.com, with a companion running-notes/runbook/handoff/showcase document set maintained every turn.
- **sources:** PLAN_imagery_best_in_class.md, HANDOFF_imagery_best_in_class.md, RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** builds on IMG-001; IMG-004 (Lane B/D2); IMG-038 (I1 style guide); IMG-039 (D5 logo lock); IMG-043 (I2 sprites); IMG-044/045/046/047/048 (I3/I4/I5/I6/I7); IMG-058 (I0 finding); RUNBOOK human-gate model (IMG-063)
- **verify-later:** phase status blocks vs live DB/site state; open runbook items B4/B5/B9/B10/B11

### IMG-004 — Two lanes of imagery: plan-driven vs content-driven (Lane B)
- **status:** aspirational
- **status-evidence:** D2 "confirmed" 2026-07-08 as a design decision; Lane B storage decision ("generic entity_type + entity_id columns on assets") also confirmed but not implemented; Phase I3 (first Lane B consumer) "not started" as of 2026-07-12.
- **what:** Everything built through Phase 2G/I0 is plan-driven — a fixed imagery list decided once at plan time. Card images, news charts, and product sketches are content-driven instead: attached to articles/news items/products, arriving continuously after the plan, with prompts composed from the content itself plus the brand guide. Lane B generalises the existing affiliate `custom_image_id` pattern via new `entity_type`+`entity_id` columns on `assets`, per-entity work item types, and content-sweeping discovery checks — sharing all generation/deploy/audit machinery downstream of the work item with Lane A (plan-driven).
- **sources:** PLAN_imagery_best_in_class.md#3/#8, RUNNING_NOTES_imagery_best_in_class.md#Turn-2
- **relations:** IMG-044 (content-linked card imagery, I3), IMG-045 (news imagery, I5), IMG-047 (product sketches, I6) are its planned instances; IMG-003 (parent programme)
- **verify-later:** assets table for entity_type/entity_id columns (expected absent as of extraction)

### IMG-005 — imagery_direction prompt prepend + origin_model/origin_prompt provenance (Phase 0/0.1/0.2)
- **status:** deployed
- **status-evidence:** "0.1 — read imagery_direction ✅ delivered, verified 2026-05-08"; "origin_model='sdxl' — Phase 0.2 column write happened" (2026-05-08); Phase 0.1 (SQL migration 107) wires site_id through six call sites. Later superseded per-site by IMG-038 (imagery_style_guide) when present, avoiding double-prepend.
- **what:** `generate_image` reads `site_specs.design_intent.imagery_direction` and prepends it to the subject prompt ("Style direction: … Subject: …", later capped at 200 chars on an SDXL-aware sentence boundary), closing the "webdesign-agent writes imagery taste, image-generator ignores it" gap. In the same phase, `store_asset` was fixed to actually populate `assets.origin_prompt` (previously silently dropped — every row was NULL) and `assets.origin_model`; later extended so the adapter returns real provider/model_id (`banana/gemini-3-pro-image-preview` vs `sdxl`) instead of a hardcoded literal. Required coordinated Go+SQL shipping across image-build-handler, site-work-orchestrator, and pageflow-builder. Provenance is the substrate for every later spec-vs-delivery audit. Side benefit: pulled planner-invented hero prompts back toward the adopted look (partial mitigation of IMG-006/Bug 4).
- **sources:** PLAN_imagery_loop_closure.md#Phase-0 (U02/U10); old/PHASE_0_BUNDLE_README.md#Phase-0.2; STATUS_imagery_2026-05-08.md; 107_image_build_handler.sql Sections 1-3 (U18); 029_image_generator.sql
- **relations:** IMG-038 (imagery_style_guide supersedes this at runtime); IMG-006 (Bug 4, partially mitigated); IMG-007 (discovery checks read origin_model/origin_prompt); IMG-024 (future auditor compares provenance to delivered image)
- **verify-later:** `getImageryDirectionForSite`/`composeImagePromptWithDirection` in generate_image_actions.go; `SELECT origin_model, origin_prompt FROM assets` distribution

### IMG-006 — Planner ignores site_archetype imagery constraints (Bug 4)
- **status:** aspirational
- **status-evidence:** "site_archetype.design.imagery … says 'minimal icons/diagrams, no decorative photography'. The planner's site_plan still produced lavish hero prompts" (2026-04-23); no fix-verification recorded in scope since.
- **stage2-verified (2026-07-14):** unknown → aspirational — Current planner prompt (docs/agent_docs/sql_for_agents/053_build_site_planner.sql, plan_site step ~line 855+) has 0 hits for 'site_archetype'/'archetype' anywhere in the prompt_template, and always sets needs_images:true, needs_logo:true per RULE 7 — confirms the Bug-4 fix (read site_archetype.design.imagery, suppre...
- **what:** The planner invents hero image prompts that contradict the adopted archetype's imagery stance. The proposed fix shape is for the planner prompt to read `site_archetype.design.imagery` and set `needs_images=false` when it says none/minimal. Phase 0.1's imagery_direction prepend (IMG-005) only partially mitigates the symptom (it steers style, it doesn't suppress unwanted images).
- **sources:** HANDOFF_2026-04-23(1).md Bug 4; ASSESSMENT_imagery_phase_0_1…md#Bug-4
- **relations:** IMG-005 (partial mitigation); adoption faithfulness (broader category)
- **verify-later:** plan_site prompt for archetype imagery constraint

### IMG-007 — Algorithmic imagery discovery checks (Phase 1 trio)
- **status:** deployed
- **status-evidence:** "Phases 1.1–1.4 all ✅ delivered 2026-05-08 (1.2/1.3 partial — check fires; symptom-path site needed)"; dispatch doc separately shows the checks live and routing work items.
- **what:** Three no-LLM Go discovery checks catching spec-to-delivery gaps at zero LLM cost: `unfulfilled_image_prompt` (planner asked for an image, no asset exists), `placeholder_image_in_use` (a hardcoded fallback path is rendered in `rendered_html` with no backing asset — the silent-failure case), and `image_url_404` (HTML references an image URL no `assets` row backs, DB-only detection). All follow the `DiscoveryCheck` interface, register via `init()`, and were appended to design-discovery-agent's `run_checks`. A ~12-item wishlist of further checks (alt-text, dimensions, orphans, cross-site contamination, multi-image underfill) was catalogued and mostly remains unbuilt. These three checks emit `needs_imagery`/`needs_hero_image`/`needs_logo` items to image-build-handler via the dispatch loop.
- **sources:** PLAN_imagery_loop_closure.md#Phase-1 (U02/U10); FOCUS_imagery_assessment_1_.md#13.2; old/phase1/phase_1_register_imagery_checks.sql; FOCUS_dispatch_diagnostic(4).md#why-stuck/#Q4
- **relations:** IMG-015 (check_unfulfilled_imagery_plan, the 2G.4 structured successor for the new site_plan_imagery shape); IMG-059 (image_source_unsatisfiable, a later sibling); pipeline soft-label bug; baked-fallback problem (IMG-051)
- **verify-later:** `platform/orchestration/actions/discovery_checks/check_*.go`; design-discovery-agent run_checks array; unfulfilled_imagery_plan pipeline value

### IMG-008 — Asset locking mirrors page_components (Phase 2A)
- **status:** deployed
- **status-evidence:** Planned in Phase 2; FOCUS_adoption_faithfulness (2026-05-19) states migration 053 adds `lock_type`+`lock_expires_at` "on page_components, site_components, site_plan_directives, assets … written, ready to apply."
- **what:** `assets` gains `locked_at`/`lock_type` using the same vocabulary and exclusion predicate as `page_components`, so audits and discovery checks skip locked assets (e.g. an approved, permanent logo). This is the schema foundation IMG-039 (logo permanence) and every later "honour locks" clause depends on.
- **sources:** PLAN_imagery_loop_closure.md#Phase-2; FOCUS_adoption_faithfulness_via_locks(2).md#implementation-plan
- **relations:** IMG-039 (logo permanence consumes this); IMG-009 (asset_key, same migration family); IMG-021 (adoption mirror also proposed to use asset_key namespacing)
- **verify-later:** assets table columns and indexes; lock-exclusion predicate reused verbatim from page_components

### IMG-009 — asset_key multi-image model (Phase 2B–2D)
- **status:** deployed
- **status-evidence:** "2B ✅ 2026-05-09; 2C ✅ deployed 2026-05-09; 2D ✅ applied 2026-05-10", with live psql migration output pasted (11 rows backfilled, backup table `assets_backup_20260508_pre_phase2d`, ON_ERROR_STOP guard, old unique index dropped after a straggler sanity check).
- **what:** Broke the one-asset-per-purpose-per-site constraint. New `assets.asset_key` column (backfilled from `purpose`), new unique index `(site_id, asset_key) WHERE active`, `StoreAssetAction`'s ON CONFLICT target switched to `asset_key`, then the old `(site_id, purpose)` unique index dropped. Enables N heroes/icons/illustrations per site — canonical hero is `hero/hero`, a variant is `hero/hero_about` (the `(purpose, asset_key)` split). Strict production apply order was documented and followed: 2A→2B→2C deploy→verify→2D.
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2B/2C/2D; PLAN_imagery_loop_closure.md#Phase-2; old/phase2/2E/phase_2e_store_asset_action.diff (U10); docs/agent_docs/sql_for_tables/041_assets.sql#Phase2B/#Phase2D (U19)
- **relations:** IMG-010 (hero-variant routing consumes it); IMG-054 (DeployedWebPath derives filenames from asset_key); IMG-012 (site_plan_imagery.key → namespaced asset_key)
- **verify-later:** `\d assets` indexes; StoreAssetAction ON CONFLICT target; idx_assets_site_asset_key_unique

### IMG-010 — Hero-variant routing through image-build-handler (Phase 2E)
- **status:** deployed
- **status-evidence:** "2E ✅ delivered, verified 2026-05-12" with a full hero_about end-to-end trace.
- **what:** Made `hero_<page>` variants routable: `check_unfulfilled_image_prompt` classifies logo / hero_home / hero_<page> into needs_logo / needs_hero_image / unfulfilled_hero_variant; a new `hasActiveAssetForAssetKey` helper replaced a purpose-level check that gave false positives; `deploy_image_asset` derives per-variant paths (`assets/images/hero-about.jpg`, underscores to hyphens); `StoreAssetAction` gains an `asset_key_field` JSONPath config; a third variant branch was added to the image-build-handler workflow (spawn/call/store/deploy), leaving the logo/hero branches untouched. Known gap left open: the variant chain doesn't pass `site_id`, so imagery_direction isn't prepended for variants.
- **sources:** STATUS_imagery_2026-05-12.md#Phase-2E; old/phase2/2E/check_unfulfilled_image_prompt.go; old/phase2/2E/phase_2e_image_build_handler_variant_path.sql
- **relations:** IMG-016 (needs_imagery branch, 2G.5, later sits in front of this chain); IMG-009 (asset_key it consumes)
- **verify-later:** image-build-handler workflow branches in agent_definitions; imagery_helpers.go

### IMG-011 — Spawned asset-deployer deploy pattern / storage-env isolation (Phase 2F)
- **status:** deployed
- **status-evidence:** "2F ✅ deployed + verified 2026-05-12"; boxed warning re-confirmed 2026-07-10, "Where deploys run (by design, don't 'fix' this)"; independently confirmed by the leopardess audit (AUDIT_verified_facts#D8 finding 1).
- **what:** The base agent-chassis pod deliberately carries no `IMAGE_BUCKET` and builds no storage client, so inline `deploy_image_asset` fails there with "storage client not available" — by design, not a bug. Real deploys spawn an `asset-deployer` child agent, into which `spawn_actions.go` injects S3/B2 env via the `isStorageEnabledAgent` list (documented at `107_image_build_handler.sql:725`). Phase 2F replaced three inline deploy step-pairs in image-build-handler with spawn+call pairs targeting asset-deployer. Hand-triggering asset-deployer standalone reproduces a spurious "Storage client not configured" failure because it skips the injection — a triggering mistake, not a defect.
- **sources:** PLAN_imagery_loop_closure.md#2F; PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS; STATUS_imagery_2026-05-08.md#[BLOCKER]-Storage-architecture-mismatch (U10); docs/leopardessconsulting/AUDIT_verified_facts.md#D8, RUNNING_NOTES.md#Turn-10 (U25)
- **relations:** IMG-040 (brand_head mode rides the same asset-deployer agent); IMG-054 (two-URL model, same audit); storage-architecture (S3/B2 credentials, cross-category)
- **verify-later:** `agentbase/agent.go:294`; `spawn_actions.go` isStorageEnabledAgent list; `107_image_build_handler.sql:725` comment

### IMG-012 — site_plan_imagery structured plan table (Phase 2G.1)
- **status:** deployed
- **status-evidence:** "2G.1 — site_plan_imagery table ✅ delivered 2026-05-12"; live `chk_kind` constraint re-verified 2026-07-08; live JOIN in production code confirmed by 2026-06-02 ("site_plan_imagery.key = assets.asset_key").
- **what:** A sibling of `site_plan_directives` holding structured per-image plan requirements at site/page/section scope: `scope`(site|page|section)+`scope_ref` (NULL / page_name / `page:ordinal` for sections, enforced by `chk_scope_ref_consistency`), `key` (→asset_key stem), `kind` CHECK enum (logo|hero|illustration|icon|infographic, later +sprite_sheet — mirrored in Go as `validImageryKinds`, constraint and mirror must change together), required `prompt`, JSONB `style_hints`/`constraints` that cascade additively with directives' imagery_direction, `ordering`, `source` CHECK (llm|classifier|manual|adoption), and HITL lock columns with the same lock-transfer treatment as directives. Unique on `(plan_id, scope, COALESCE(scope_ref,''), key)`. `product` is deliberately excluded — product imagery comes from the affiliate_products resolver, not the planner. Table is the successor to the legacy `site_specs.site_plan.image_prompts` flat dictionary; scoped page rows are what drive per-page heroes.
- **sources:** PLAN_imagery_phase_2g.md#Schema; old/phase_2g_step1_site_plan_imagery.sql; SQL_2026-07-12_add_sprite_sheet_kind.sql (U10); docs/agent_docs/sql_for_tables/043_site_plan_imagery.sql (U19); FOCUS_site_spec_vs_site_plan.md#where-imagery-lives, HANDOFF_2026-06-02…md#fix (U02)
- **relations:** IMG-013 (write path that populates it); IMG-014 (planner prompt that supplies it); IMG-015 (discovery check that reads it); IMG-051 (per-page hero resolver consumes it); five-place new-kind checklist (IMG-031)
- **verify-later:** `\d site_plan_imagery`; chk_kind vs validImageryKinds in write_site_plan_action.go (~line 183); table population on a fresh build

### IMG-013 — flattenImageryBlock write path + imagery lock transfer (Phase 2G.2)
- **status:** deployed
- **status-evidence:** "✅ deployed 2026-05-12; path fix on 2026-05-13 (function looked up data['imagery'] at top level rather than walking wrapper shapes via findDirectiveTree)".
- **what:** `write_site_plan` walks the planner's `imagery` JSON block and inserts `site_plan_imagery` rows in the same transaction as pages/sections/directives (`flattenImageryBlock` + `insertImageryRow`, enforcing the kind enum), and transfers locks from the previous current plan's locked imagery rows matched on `(scope, scope_ref, key)` — so locked HITL prompt edits survive plan rebuilds.
- **sources:** PLAN_imagery_phase_2g.md#write_site_plan-extension; PLAN_imagery_loop_closure.md#2G
- **relations:** IMG-012 (the table it writes to); content-governance lock semantics (cross-category)
- **verify-later:** write_site_plan_action.go flattenImageryBlock/insertImageryRow; lock-transfer behaviour on replan

### IMG-014 — Planner imagery block prompt extension (Phase 2G.3)
- **status:** deployed
- **status-evidence:** "2G.3 ✅ delivered 2026-05-13 (with max_tokens bump + path fix)"; 2026-07-08 ground truth confirms "planner prompt carries the Imagery Block + decomposition rule; max_tokens now 16000".
- **what:** build-site-planner's JSON output gains an `imagery` key (`site[]`/`pages{}`/`sections{}` entries with key, kind, prompt, optional style_hints/constraints) produced in the same single LLM call as pages/design_direction — no separate imagery-planning call. Replaces the flat `image_prompts:{logo,hero_home}` contract that had hero/logo-only names baked in. `max_tokens` was raised 4000→8000 (JSON truncation observed on a 14-page roadmap) and later to 16000. The legacy `image_prompts` output continued to be emitted during the transition period.
- **sources:** PLAN_imagery_phase_2g.md#Planner-output-shape; PLAN_imagery_loop_closure.md#Application-status; FOCUS_imagery_assessment_1_.md#4.1
- **relations:** IMG-032 (one-entry-one-image decomposition rule, same prompt patch); IMG-033 (planner key stability problem); sprite_sheet planner emission
- **verify-later:** build-site-planner default_config prompt_template "## Imagery Block"; sql_for_agents/053 patches

### IMG-015 — check_unfulfilled_imagery_plan discovery check (Phase 2G.4)
- **status:** deployed
- **status-evidence:** "2G.4 ✅ delivered 2026-05-14 (8 work items emitted on first run; correct priority ordering)"; the `Pipeline:"build"` hardcoding fix confirmed live in code 2026-07-08 (an earlier version emitted with `pipeline='design'`, a stuck-work-item bug).
- **what:** Walks the current plan's `site_plan_imagery` rows and emits one `needs_imagery` work item per row lacking a matching active asset (via `hasActiveAssetForAssetKey`), capped at 20/pass, priority-banded (site logo 70 → index hero 65 → site other 75 → page hero 80 → page other 90 → section 100), mirroring the legacy `classifyPromptKey` bands. `computeAssetKey` namespaces deeper keys (`page.about.illustration_team_values`, `section.home.2.icon_precision`) while keeping hero/logo names flat for backward-compatible deploy paths. Dedup key `needs_imagery:<scope>:<scope_ref|->:<key>`. Ran in parallel with the legacy `unfulfilled_image_prompt` check during the transition (both call `hasActiveAssetForAssetKey` to avoid double work).
- **sources:** PLAN_imagery_phase_2g.md#Discovery-check-1; PLAN_imagery_loop_closure.md#Decisions/#2G; TODO_imagery_followups.md#7; FOCUS_dispatch_diagnostic(4).md#Q4 (U02, pipeline emission bug)
- **relations:** IMG-007 (Phase-1 trio it succeeds for the structured shape); IMG-012 (table it reads); IMG-016 (needs_imagery items it feeds)
- **verify-later:** check_unfulfilled_imagery_plan.go (hardcoded Pipeline "build"); design-discovery-agent run_checks

### IMG-016 — needs_imagery branch in image-build-handler (Phase 2G.5)
- **status:** deployed
- **status-evidence:** "2G.5 ✅ delivered 2026-05-14 (with two hotfixes — optional input_mapping + store_asset purpose)"; first asset `a12b5d71` through the new path 2026-05-14; 057/107 SQL migrations define/extend the handler.
- **what:** A new branch alongside (not replacing) the Phase-2E variant chain: `check_item_type_imagery` → `spawn_image_gen_imagery` → `call_imagery_gen` (site_id passed so imagery_direction prepends; kind/style_hints/constraints pass through) → brand-update conditional store → shared `spawn_asset_deployer` tail. A `spec.brand_update` boolean, computed at discovery time (site scope OR index-page hero), routes whether the stored asset also updates site brand assets. Two hotfixes: established the `?`-suffix optional input_mapping convention, and exposed that `store_asset` lacked a `purpose_field` (it initially hardcoded purpose:"hero", blocking kind=logo items — fixed 2026-05-20). A future refactor option is recorded: collapse the three legacy branches into needs_imagery ("always fix legacy with modern").
- **sources:** PLAN_imagery_loop_closure.md#2G/#Step-5-workflow; PLAN_imagery_phase_2g.md#image-build-handler-extension; TODO_imagery_followups.md#What-shipped-this-session (U10); 057_image_build_handler.sql; 107_image_build_handler.sql "phase_2g_step5" section (U18)
- **relations:** IMG-010 (hero-variant branch it sits in front of); IMG-015 (emits the items it consumes); `?` optional-mapping convention; purpose_field fix
- **verify-later:** image-build-handler workflow JSON; store_imagery_asset purpose_field config

### IMG-017 — Legacy image_prompts age-out check — retired/reframed (Phase 2G.6)
- **status:** superseded
- **status-evidence:** Old plan versions: "Migration off legacy image_prompts: Age-out via check_legacy_image_prompts_aspect, registered last"; live plan: "2G.6 ❌ retired as scoped. Reframed 2026-05-13 … one string out of a JSON array, no code change."
- **what:** Originally a dedicated discovery check that would emit `needs_replan` for sites still on `site_specs.site_plan.image_prompts` (deliberately registered LAST to avoid replan churn before the planner extension shipped). Reframed and retired: "is a site on legacy?" was judged not a fault signal in itself — the existing checks already detect brokenness on both paths. Migration became a purely operational deregistration decision (pull `unfulfilled_image_prompt` from run_checks once it reliably finds zero gaps) rather than a shipped check.
- **sources:** old/PLAN_imagery_loop_closure(3).md#Decisions; PLAN_imagery_phase_2g.md#Discovery-check-2; PLAN_imagery_loop_closure.md#Decisions (2026-05-13 reframe)
- **relations:** superseded by "operational deregistration, not a dedicated check"; IMG-007 (the dual-check transition period it concerns)
- **verify-later:** confirm no check_legacy_image_prompts_aspect.go exists; whether unfulfilled_image_prompt is still registered

### IMG-018 — Image-generator request shape + per-kind defaults (Phase 2H)
- **status:** deployed
- **status-evidence:** "2H ✅ delivered (action layer) 2026-05-14 partial — chassis confirmed; adapter binary unconfirmed"; the adapter side shipped with the 2026-05-20 provider work (IMG-030).
- **what:** Extends the generation request beyond `{prompt,width,height}`: v1 fields negative_prompt, seed, reference_image_uri (pass-through), cfg_scale, steps. A Go-side `kindDefaults` map supplies per-kind defaults (logos get people/text/watermark negative prompts; icons get a tighter aspect; heroes unchanged), overridable by the caller's spec; `style_hints.aspect_ratio` drives dimensions and `constraints` feed the negative prompt. `style_preset`/samples/safety_mode were deferred. Defaults deliberately live in Go, not a config table.
- **sources:** PLAN_imagery_loop_closure.md#2H; STATUS_imagery_2026-05-12.md#Phase-2H-(proposed); TODO_imagery_followups.md#4
- **relations:** IMG-030 (provider abstraction, adapter-side completion); IMG-019 (parseAspectRatio whitelist fix, a regression in this phase's territory); "constraints informational only" decision
- **verify-later:** kindDefaults/resolveKind/parseAspectRatio in generate_image_actions.go; adapter field mapping in dynamic_adapter.go

### IMG-019 — parseAspectRatio SDXL v1.0 whitelist snap fix
- **status:** deployed
- **status-evidence:** Elevated HIGH 2026-05-18 ("16:9 → 1024×576, SDXL rejects"); a 2026-07-11 note refers to "pre-SDXL-snap-fix residue", implying the fix landed and is now historical.
- **what:** `parseAspectRatio` snapped requested dimensions to multiples of 64 rather than SDXL v1.0's strict dimension whitelist (1024×1024, 1152×896, 1344×768, …), so planner-emitted `aspect_ratio` hints produced rejected sizes and blocked hero generation — a regression enabled by the Phase 2H/planner prompt patch (heroes previously fell through to valid kindDefaults instead of an aspect hint). Fix: snap to the nearest whitelist pair matching the requested orientation.
- **sources:** TODO_imagery_followups.md#5; RUNNING_NOTES_imagery_best_in_class.md#Turn-24
- **relations:** IMG-018 (Phase 2H request shape); IMG-032 (planner prompt patch that moved aspect into style_hints, the trigger)
- **verify-later:** whitelist logic in generate_image_actions.go; test 16:9→1344×768

### IMG-020 — Build-time imagery trigger: emit_imagery_items + imageryplan.go shared selection
- **status:** deployed
- **status-evidence:** "The imagery trigger had the identical bug and is fixed in the same deploy… deployed 2026-05-26, chassis v1.0.1047" (FOCUS_design_composition §3A).
- **what:** `write_site_plan` recorded planner image requests into `site_plan_imagery` (via flattenImageryBlock, IMG-013), but nothing on the build path acted on them at plan time — `needs_imagery` previously came only from the loop's `unfulfilled_imagery_plan` check (capped 20/pass, running after the fact). `emit_imagery_items_action.go` closes this by emitting at plan time itself; `imageryplan.go` is a shared package holding row selection, priority/severity classification, the brand_update rule, item_key, and spec body — used by BOTH the build-time emitter (status `triaged`) and the loop-time check (status `detected`), the anti-drift pattern of one source of truth for two call sites. Priority bands run index hero 65 … section 98, ahead of the terminal needs_rerender at 99. Known asymmetry: emit_imagery has no site-level no-backfill guard, unlike the equivalent emit_design.
- **sources:** FOCUS_design_composition_flow_and_adoption_fidelity(1).md#3A; README_difference_between_work_site_orchestrator_and_build_site_planner.md
- **relations:** IMG-001 (loop-closure programme, same-shaped gap as design); IMG-015 (the loop-time check sharing imageryplan.go); IMG-064 (imagery work-item economy, this is one hop in the chain)
- **verify-later:** emit_imagery_items_action.go, imageryplan.go; site_plan_imagery rows on a fresh build

### IMG-021 — Adoption image mirror (Phase 3 / Lane hook)
- **status:** aspirational
- **status-evidence:** "3 — adoption image mirror ⏸ deferred 2026-05-14. Not cancelled … reference_image_uri plumbing preserved as a forward-compat hook"; "Adopted images are persisted but only as historical record for now" — no deployment claim anywhere in scope.
- **what:** Stops discarding crawled imagery at site adoption. A new `mirror_adoption_images` action would download crawl images (capped ~50/site, 5MB each), upload to S3, and insert `assets` rows with `origin_type='adopted'`, `asset_key='adopted:<filename>'`, `origin_url`, and attribution/license — wired into `apply_adoption_plan` plus a backfill discovery check `check_crawled_images_discarded` routed to a new one-step `adoption-image-mirror` agent. Adopted images would become img2img/style references and future auditor signal. Deferred specifically because current adopted sites carry minimal imagery, not because the design is blocked.
- **sources:** PLAN_imagery_loop_closure.md#Phase-3 (U02/U10); FOCUS_imagery_assessment_1_.md#7/#9-item-9
- **relations:** IMG-009 (asset_key namespacing it would use); IMG-049 (reference-image anchoring, its consumer); adoption-pipeline category (site crawling, cross-category)
- **verify-later:** existence of mirror_adoption_images_action.go, adoption-image-mirror agent row; assets rows with origin_type='adopted'

### IMG-022 — Visual auditor imagery awareness (Phase 4, text-only)
- **status:** aspirational
- **status-evidence:** "4 — visual auditor sees imagery (text-only) — not started."
- **what:** Would extend visual-design-auditor's `load_design_context` SQL to include `assets` rows (unlocked, generated/adopted), `imagery_direction`, and `site_plan_imagery`, and add IMAGERY as a sixth check category, with algorithmic-check results passed through to avoid double-flagging. Would need tuning on 5-10 sites before enabling fixes (≥80% accuracy target). Today the auditor's context contains zero image data — it cannot notice a missing or off-brief hero.
- **sources:** PLAN_imagery_loop_closure.md#Phase-4; FOCUS_imagery_assessment_1_.md#13.1/#13.3
- **relations:** IMG-024 (imagery-quality-auditor, option B chosen as the eventual full answer instead of extending this one); design-composition auditors (cross-category)
- **verify-later:** visual-design-auditor load_design_context SQL in agent_definitions

### IMG-023 — Vision-capable LLM path (Phase 5)
- **status:** aspirational
- **status-evidence:** "5 — vision-capable LLM path — not started" (both U02 and U10 sources, no deployment claim).
- **what:** Foundational capability required before any auditor can actually look at images: extend `aiservice.AIService` with image inputs (`GenerateTextWithImages`), implement Anthropic vision content blocks (URL source), prefer extending `execute_llm_prompt` with an `image_urls_field` config over a wholly new action, refresh presigned URLs immediately before the call (freshness requirement), and tag `vision_call:true` in `llm_call_log` for cost separation.
- **sources:** PLAN_imagery_loop_closure.md#Phase-5
- **relations:** IMG-024 (imagery-quality-auditor, its main consumer); sprite-sheet vision auto-verify (IMG-043's deferred I2.4/I8 step); llm_call_log
- **verify-later:** platform/aiservice/anthropic.go vision support

### IMG-024 — imagery-quality-auditor agent (Phase 6 / I8)
- **status:** aspirational
- **status-evidence:** "6 — imagery-quality-auditor agent — not started"; I8 in the best-in-class plan also "not started" as of 2026-07-12.
- **what:** A planned vision-capable sibling of visual-design-auditor under design-audit-agent, dedicated to imagery: findings categories direction_mismatch / brand_mismatch / inconsistency / quality / inappropriate; max_fix_attempts 2; findings route to image-build-handler for regeneration (different prompt/seed/negative prompt), escalating to needs_human_review; honours asset locks and origin_type='uploaded' (never touches human uploads); counts toward the existing 3-pass audit cap; gated rollout via design-audit-agent. I8 (the best-in-class renumbering) adds sprite-sheet cell verification and brand-guide reference comparison. Chosen deliberately over extending the existing text-only auditor (IMG-022) — a separate TOP-5 finding cap, and only imagery pays the vision cost.
- **sources:** PLAN_imagery_loop_closure.md#Phase-6 (U02/U10); FOCUS_imagery_assessment_1_.md#13.4; PLAN_imagery_best_in_class.md#Phase-I8
- **relations:** IMG-023 (vision path it requires); IMG-039 (imagery_style_guide as its eventual audit standard); IMG-005 (provenance it would compare against); IMG-008 (asset locking it must honour)
- **verify-later:** no imagery-quality-auditor row expected yet in agent_definitions

### IMG-025 — Image generation as parameter shaping (deferred composer/envelope step)
- **status:** aspirational
- **status-evidence:** "The composer step is its own work" (recorded during Phase 2G step 5 design, ~2026-05); separately, "Step 5 — image prompt cascade: Defer … see FOCUS_prompt_composition_pattern.md for why copying the text pattern is the wrong target" (resolved 2026-05-13). Partially realised since by IMG-031's per-kind gating (composition-by-kind, not a full cascade).
- **what:** Unlike text, images have a ~77-token CLIP prompt budget and no "don't" understanding, so composition for images means deriving parameters — subject, negative_prompt from kind, style_preset/LoRA from imagery_direction, reference_image_uri from adopted images, aspect/cfg/steps per kind — not blending prose the way page-content-writer composes text. A cheap `compose_image_request` step (Go rules, or a small LLM) producing a parameter envelope before image-generator is the candidate design, judged as belonging with Phase 2H request-shape work but deliberately not built (copying the text-composition pattern was judged the wrong target for images).
- **sources:** FOCUS_prompt_composition_pattern.md#What-this-means-for-images (U02); PLAN_imagery_loop_closure.md#Decisions/#Image-prompt-cascade—deferred/#Open-items (U10)
- **relations:** IMG-018 (Phase 2H request shape); IMG-031 (per-kind gating, the partial realisation); mega-prompt fragility (envelope pattern B, cross-category)
- **verify-later:** whether any composer/envelope step or compose_image_request action exists

### IMG-026 — Icon generation lessons and image-model comparison
- **status:** deployed
- **status-evidence:** "Verdict (2026-05-18): Fix A is insufficient… SDXL ignores style instructions"; "Final icon batch state (verified 2026-05-20): all six icons banana/gemini-3-pro-image-preview, visual gate passed."
- **what:** SDXL is the wrong tool for flat-vector icons — it has a strong photorealism bias on concrete subjects, drifts on multi-panel prompts, and has no real transparency. A full model comparison (SDXL, SD3.5, FLUX schnell/dev/pro, DALL-E 3, Imagen 3, Nano Banana Pro 2, Midjourney, LLM-SVG) ranked FLUX schnell cheapest-good and Banana best for reference-conditioned sibling consistency; the decision was to plumb reference images AND switch icon generation to Banana. A first attempt (Fix A, prompt-only rescue) was tried and judged insufficient before this switch. Related fixes: a `purpose_field` so icons store as purpose=icon (240×240, not hero's 1600×900); kindDefaults icon dimensions; a jpg-vs-png note for thin line art.
- **sources:** TODO_imagery_followups.md#23; old/001_image_model_comparison.md; TODO_imagery_followups.md#Final-icon-batch-state
- **relations:** IMG-030 (provider abstraction this decision drove); IMG-028 (transparency abandonment); IMG-027 (LLM-SVG sleeper option); IMG-049 (reference-image anchoring)
- **verify-later:** ImagePurposes["icon"]; assets rows origin_model for icon assets

### IMG-027 — LLM-generated SVG icon path (sleeper option)
- **status:** aspirational
- **status-evidence:** "Sleeper option for later: LLM-generated SVG for icons… Worth a focused experiment once the immediate work is shipped" (2026-05-18).
- **what:** Icons are vector by nature; an LLM (Claude/GPT) writing SVG markup directly bypasses the entire convince-a-diffusion-model problem at roughly $0.001-0.005/icon, crisp at any size, no copyright concern. Was the analyst's original recommendation before the user chose the Banana raster route (IMG-026); retained as a possible future replacement of the raster icon pipeline.
- **sources:** TODO_imagery_followups.md#23 (options c1/c2, recommendation); FOCUS_imagery_assessment_1_.md#9-item-6
- **relations:** superseded for now by IMG-026 (Banana raster icons); IMG-029 (Lucide covers UI chrome regardless of this path)
- **verify-later:** none — idea only, no code expected

### IMG-028 — Diffusion transparency abandoned → flat-grey chip icons
- **status:** deployed
- **status-evidence:** "Transparency abandoned as too fragile — image models paint a transparency checkerboard into RGB (confirmed: icon_cycle_time mode=RGB has_alpha=False). Decision: option 2, 'embrace the box'" (2026-05-20).
- **what:** Image models cannot produce true alpha; requesting transparent backgrounds yields a painted checkerboard baked into RGB. Locked decision: icons generate on a flat selectable grey background (#EEEEEE bg / #4A4A4A line) and are presented inside a styled CSS chip; the planner prompt and all existing icon specs were patched accordingly. The same lesson recurs for sprite sheets (flat selectable background, never transparent).
- **sources:** TODO_imagery_followups.md#Icon-background-resolution; SCOPE_I2_sprite_sheets.md#3 (planner prompt); CONTEXT_PACK_imagery_sprite_sheet.md#Attach—docs
- **relations:** IMG-026 (icon lessons); IMG-043 (sprite-sheet prompt rules inherit this); CSS chip styling (left as site-template work)
- **verify-later:** planner prompt icon-background wording; icon assets' actual backgrounds

### IMG-029 — Lucide icon strategy and validator wiring
- **status:** partial
- **status-evidence:** "Lucide validator written (lucide_icons.go) — NOT yet wired" (2026-05-20), re-verified still unwired 2026-07-08; listed as an open I0 close-out item in the best-in-class plan.
- **what:** The features grid renders icons as Lucide webfont glyphs (`<i data-lucide="{{.icon}}">`), not generated raster — the generated icon pipeline was never the right tool for UI chrome. Missing icons in practice are LLM-invented Lucide names that don't exist. The fix design is a single-source allowlist that is both the prompt's choice list AND a pre-store `SanitizeFeatureIcons` sweep, plus an optional render-time net; the allowlist must be verified against the bundled Lucide version. Blocked on identifying the content-generation step that fills features `content_data`. Icon strategy stays deliberately dual (D6): Lucide for UI chrome, generated sprites/raster for decorative glyphs.
- **sources:** TODO_imagery_followups.md#features-component-icons; old/verify_and_wire_lucide.md; PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** IMG-043 (sprite sheets cover decorative glyphs, the other half of D6); IMG-027 (LLM-SVG, an alternate icon path)
- **verify-later:** callers of SanitizeFeatureIcons/ValidateLucideIcon outside lucide_icons.go

### IMG-030 — Image provider abstraction and kind→provider routing
- **status:** partial
- **status-evidence:** 2026-05-20 close-out: "Provider abstraction — internal/adapters/imagegenerator/{provider,stability,banana}. dynamic_adapter.go routes by kind: icon → Banana, everything else → Stability. Proven end-to-end." The A6 extension (2026-07): "Committed in dynamic_adapter.go; NOT deployed to cluster (needs a Makefile build-from-local-filesystem)."
- **what:** The originally hardcoded Stability-only adapter (env-driven; the image-adapter.yaml/agent-definition config blocks are misleading and unread) was refactored into provider packages with kind-based routing. As shipped 2026-05-20: flat kinds (icon, sprite_sheet) → Google Banana `gemini-3-pro-image-preview`; photographic kinds → Stability SDXL. The provider Request carries `ReferenceImageURIs` (Banana native reference-image support, the mechanism brand consistency depends on — Stability has no equivalent on its standard REST endpoint). A6 (2026-07) extends the Banana route to logo/illustration/infographic as well, leaving only hero/photographic kinds on SDXL — committed in code but, per the leopardess audit, not yet deployed to the cluster (needs a Makefile build-from-local-filesystem, per this repo's build practice). Imagery kinds are constrained twice — `chk_kind` (DB) and `validImageryKinds` (Go, in write_site_plan_action.go) — so changing the kind set always needs a migration + Go edit together. Known opens: Stability provider timeout 60s vs the old 120s; circuit breaker not threaded into provider clients.
- **sources:** TODO_imagery_followups.md#What-shipped-this-session; FOCUS_imagery_assessment_1_.md#1.1–1.3; PLAN_imagery_best_in_class.md#2 (U10); docs/leopardessconsulting/RUNNING_NOTES.md#Turn-4/#Turn-7, RUNBOOK.md#O5/#landmine-5, PLAN_leopardess_rebuild.md#L2 (U25)
- **relations:** IMG-026 (icon lessons that drove the initial split); IMG-018 (2H request shape); IMG-012 (chk_kind/validImageryKinds dual constraint); IMG-039 (logo candidate generation, a consumer)
- **verify-later:** internal/adapters/imagegenerator/ package layout; adapter switch cases; timeout value; dynamic_adapter.go routing on the running pod vs the repo (A6 deploy gap)

### IMG-031 — Per-kind prompt gating and the five-place new-kind checklist
- **status:** deployed
- **status-evidence:** "directionAppliesToKind … gates design_intent.imagery_direction OFF icons/logos" (shipped 2026-05-20); "PROVEN on real generations (icons carried palette, not medium)" (2026-07-11); sprite gating fix commit 4629aa17 proven in origin_prompt (2026-07 Turn 31).
- **what:** Photographic brand direction contaminates non-photographic kinds (prepending it to an icon prompt makes the model paint a photo around the icon), so prompt composition gates per kind: hero/illustration/infographic get medium+mood+palette; icon and sprite_sheet get palette only; logo gets nothing. Two gating functions (`directionAppliesToKind`, `styleGuide.directionForKind`) plus the DB kind constraint, the Go kind mirror, the adapter switch, and the ImagePurposes map form a standing five-place checklist that any new imagery kind must touch — the lesson learned hard during sprite-sheet integration (I2.0).
- **sources:** TODO_imagery_followups.md#What-shipped-this-session; HANDOFF_imagery_best_in_class.md#Mechanisms; RUNNING_NOTES_imagery_best_in_class.md#Turn-29
- **relations:** IMG-038 (imagery_style_guide supplies the gated content); IMG-043 (sprite_sheet contamination near-miss, the cautionary case); IMG-012 (chk_kind, one of the five places)
- **verify-later:** both gating functions list identical kind sets; grep for the five places whenever a new kind is added

### IMG-032 — One-entry-one-image decomposition rule (planner prompt patch)
- **status:** deployed
- **status-evidence:** planner_prompt_patch changes applied ("planner prompt carries the Imagery Block + decomposition rule", verified live 2026-07-08).
- **what:** A work item must describe exactly one deliverable: prompts must never ask for "a set of six icons" (SDXL/diffusion models render one six-panel image — superficially successful but unusable). The planner prompt teaches per-entry single-image prompts, bans plural/counting phrasing (RULE 16), biases toward over-decomposition (unused icons are cheap; botched multi-panel images are expensive), moves aspect ratio into `style_hints.aspect_ratio` (the field Go actually reads), and demotes `constraints` to "informational only, reserved". The `icon_cross_technology` six-panel artifact and its cleanup SQL are the canonical cautionary example.
- **sources:** old/planner_prompt_patch_imagery.md; TODO_imagery_followups.md#25/#4
- **relations:** IMG-014 (planner imagery block, same prompt); IMG-019 (SDXL whitelist fix, downstream of the aspect_ratio move); multi-entry sections remain the canonical way to express multiple images at one scope
- **verify-later:** RULE 16 in the live planner prompt; absence of "set of" phrasing in site_plan_imagery.prompt

### IMG-033 — Planner key stability across replans
- **status:** aspirational
- **status-evidence:** "Symptom (2026-05-15): previous plan keyed hero_canonical; new plan called it brand_hero_canonical … discovery emitted a fresh work item"; no fix recorded since, though stale keys from this bug were cleaned up during the best-in-class rebuild.
- **what:** The planner LLM freely chooses imagery `key` values, so replans rename equivalent concepts, discovery sees the "new" name as missing, and regenerations/orphan assets accumulate. Fix options ranked: (a) pass the previous plan's keys into the prompt with a reuse rule (lowest effort), (b) a canonical key dictionary, (c) semantic concept matching at discovery time. None confirmed shipped.
- **sources:** TODO_imagery_followups.md#26; RUNNING_NOTES_imagery_best_in_class.md#Turn-24 (stale failed rows closed)
- **relations:** IMG-014 (planner imagery block); IMG-061 (orphaned generated assets, one source of the waste)
- **verify-later:** whether the planner prompt includes previous-plan keys; duplicate-concept assets on replanned sites

### IMG-034 — generate_image action + image-generator adapter pipeline (legacy Stability-only flow)
- **status:** superseded
- **status-evidence:** docs002/0100: "Image creation is now working" (deployed at the time); the current taxonomy names site_plan_imagery (IMG-012 onward) as the successor pipeline; docs017's exact patches ("generate_hero_image → store_hero_asset → deploy_hero_image (NEW) → templates use {{.hero_url}} → /assets/images/hero.jpg") describe the same generation predating asset_key/site_plan_imagery.
- **what:** The first-generation image workflow: `GenerateImageAction` resolves prompts (template-rendered from CollectedData) and sends them to a shared adapter topic (`system.adapter.image-generator.requests`) consumed by 3 load-balanced Python adapter replicas (a Kafka consumer group), which call Stability AI, upload PNG to S3/Backblaze under `images/{client_id}/{date}/{image_id}`, and respond to the requesting agent's topic; a circuit breaker guarded API failures. `store_generated_image`/an early StoreAssetAction persisted the S3 URI into `assets` and into `sites.content_data` by purpose; `deploy_image_asset` downloaded from S3, optimised for web (resize per purpose), and base64-committed via git-adapter. A notable bug/fix on record: `GenerateImageAction` originally bypassed the image-generator *agent* and posted straight to the adapter topic — corrected so the agent properly orchestrates (parent → agent → adapter → agent → parent).
- **sources:** docs001_flow_general/README.095.a.image_handling.git.057_image.md; README.095c.image_handling_topics.md; README.097a.imagecreationandstorageflow.md; README.096b.robothandswebsite.md (U20); docs017_legacy_agent_rules_images_design_keydocs/002_pageflow_image_changes.md, 001_changes_needed.md (U21)
- **relations:** successor: IMG-012 through IMG-020 (site_plan_imagery pipeline); IMG-016 (image-build-handler, its architectural descendant); adapter microservice pattern (cross-category)
- **verify-later:** internal/adapters/imagegenerator; whether Stability AI config survives; current imagery pipeline tables vs this legacy shape

### IMG-035 — Image storage and display URL strategy (S3/B2 dual URI)
- **status:** deployed
- **status-evidence:** README.099a recommends and implements dual URIs: "image_uri (s3:// for storage reference), image_url (https:// for web use)"; robot-hands pages of that era embedded presigned URLs.
- **what:** Generated images return both an `s3://` URI (storage reference) and a public HTTPS/CDN URL for embedding in HTML. Options canvassed at the time: public-bucket/CDN (chosen), presigned URLs (rejected for the expiry problem on permanent sites — the same problem later rediscovered and fixed properly, see IMG-054), base64 embedding, and an image proxy service. Backblaze B2 public bucket setup was documented as part of this decision.
- **sources:** docs001_flow_general/README.099a.image_storage_and_display_urls.md; docs002_hitl_parallel/README.0100.state_of_play_for_creating_website.md
- **relations:** IMG-054 (the durable two-URL model that later actually solved the expiry problem this strategy anticipated); storage-architecture (S3/B2 credentials, cross-category)
- **verify-later:** ConvertS3URIToPublicURL or equivalent; current image URL scheme on live sites

### IMG-036 — pageflow-builder retirement
- **status:** superseded
- **status-evidence:** "Decision marker: 2026-05-12 — user agreed pageflow-builder is being left behind"; snapshot saved to pageflow-builder_2026-05-12.txt.
- **what:** The legacy monolithic site builder — inline `deploy_image_asset`, hardcoded `generate_logo`/`generate_hero_image`, a sequential 20-iteration page loop, and writes directly to `site_specs.site_plan` bypassing the plan-domain tables — is deliberately not extended with the 2G imagery shape. Architecture converges instead on build-site-planner/plan-builder + triaged work items + page-build-handler + image-build-handler. Sites pageflow-builder already built stay on the legacy check path until they age out; a full row snapshot exists as the recovery reference. The classifier's `recommended_builder` default was noted as a loose end.
- **sources:** PLAN_imagery_phase_2g.md#On-leaving-pageflow-builder-behind; old/pageflow-builder_2026-05-12_NOTES(1).sql; old/pageflow-builder_2026-05-12.txt
- **relations:** superseded by the plan-domain + dispatch architecture (IMG-012 onward); robot-hands rebuild dropped its recommended_builder key
- **verify-later:** pageflow-builder agent_definitions row status; any remaining live traffic

### IMG-037 — Assets table with full provenance (schema)
- **status:** deployed
- **status-evidence:** 004 PART 6 creates `assets` with origin tracking; later phases (locks, asset_key) applied to a live table already holding 11 rows. docs012/010's new-table list (assets, products, product_assets, affiliate_programs, affiliate_products) shows the assets side shipped while the product/affiliate side remained design.
- **what:** The all-binary-assets table (image/video/document/logo/favicon) recording storage location (provider/path/url), file metadata, and full provenance: `origin_type` (generated/uploaded/scraped/stock/affiliate/derived), `origin_url`/`origin_prompt`/`origin_model`, `origin_asset_id` for derivations, an `alterations` history JSONB, and attribution/license fields. A `purpose` field ('hero', 'og_image', …) drives placement, later joined by `asset_key` (IMG-009) for multi-image identity. The companion products/product_assets/affiliate_programs/affiliate_products design landed for e-commerce/review sites; products-as-entities superseded the standalone product tables.
- **sources:** docs/agent_docs/sql_for_components/004_component_architecture_schema.sql#PART6; docs/agent_docs/sql_for_tables/041_assets.sql (U19); docs012_site_maps_and_components/010_component_and_site_architecture.md#New-Tables; docs017.../002_pageflow_image_changes.md (U21)
- **relations:** IMG-009 (asset_key extension); IMG-008 (lock columns extension); entity-data (products as entities, cross-category); link_registry affiliate fields (cross-category)
- **verify-later:** StoreAssetAction; storage_provider values in use; products/affiliate_programs existence

### IMG-038 — imagery_style_guide — per-site brand guide as data (Phase I1)
- **status:** deployed
- **status-evidence:** "✅ PHASE I1 COMPLETE — LIVE-VERIFIED 2026-07-11 … per-site imagery_style_guide driving generation with per-kind gating — PROVEN on real output."
- **what:** A `site_specs` aspect `{palette, medium, mood, avoid, reference_asset_keys}` distilled from design_intent, read by `generate_image` for every generation: photographic kinds get medium+mood+palette prepended, icons/sprite sheets get palette only, logos get nothing (mirrors IMG-031's gating). The guide supersedes the free-text `imagery_direction` prepend (IMG-005) when present, avoiding a double prepend. `avoid` terms feed the negative prompt — a stronger channel than positive pleading. `reference_asset_keys` resolve to stable `s3://` URIs (presigned URLs stripped back to bucket/key so anchors outlive the 7-day signature) and flow to Banana as style anchors. Described as the single biggest lever for a consistent professional look, deliberately per-site so sites diverge on purpose.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-17/#Turn-24; SQL_2026-07-10_robothands_imagery_style_guide.sql; SHOWCASE_technical_architecture.md#4
- **relations:** IMG-005 (superseded free-text prepend); IMG-031 (per-kind gating it supplies content for); IMG-049 (reference-image anchoring); IMG-024 (future audit standard)
- **verify-later:** imagery_style_guide.go; robot-hands site_specs aspect row; "+style_guide" log lines

### IMG-039 — Logo permanence: generate → human-approve → lock (D5)
- **status:** deployed
- **status-evidence:** "Logo: user-approved, LOCKED (assets.locked_at, lock_type=permanent); store-guard refuses overwrites" (2026-07-11, B6 done); concretely exercised on the leopardess site (six candidates generated 2026-07-10, owner chose c2, exact approved PNG deployed live and byte-verified).
- **what:** One consistent logo for the life of a site is a policy, not a generation feature: the logo is generated, a human approves it via the runbook eyeball ritual (A3), `locked_at` is set, and the assets-upsert's `WHERE assets.locked_at IS NULL` guard refuses any future overwrite; auditors and regeneration paths must skip locked assets. Favicon and OG card are derived from the approved logo, never independently generated (IMG-040). robot-hands' May-8 logo was approved as-is and locked. The concrete brand-mark commissioning practice observed on the leopardess site: generate N candidates through the same model/key the chassis uses, save all prompts for pipeline reproduction, judge by small-size survival (a favicon is 16px — solid-fill marks survive, line-art dissolves), and require human approval for what is a for-the-life-of-the-site decision — the owner approves a *specific image*, not a prompt, since regenerating "the same" prompt yields a mark they never saw.
- **sources:** PLAN_imagery_best_in_class.md#D5/#Phase-I1; RUNBOOK_imagery_best_in_class.md#B6; RUNNING_NOTES_imagery_best_in_class.md#Turn-24 (U10); docs/leopardessconsulting/logo_candidates/PROMPTS.md, RUNNING_NOTES.md#Turn-8/#Turn-9, REPLICATION_in_chassis.md#4 (U25)
- **relations:** IMG-008 (locking schema this consumes); IMG-040 (brand-head derivation consumes the locked logo); IMG-030 (provider/model used); hitl checkpoint_for_review surface (cross-category)
- **verify-later:** robot-hands logo asset locked_at/lock_type; store guard in StoreAssetAction; site_plan_imagery logo row for leopardess

### IMG-040 — Brand-head derived assets (favicon + OG card) — derive_brand_head_assets action
- **status:** deployed
- **status-evidence:** "favicon.png/og-card.png serve 200; og:image + twitter:card injected into every head at render time" (I1 verification, 2026-07-11).
- **what:** `derive_brand_head_assets` deterministically derives a favicon (64×64 square resize) and an OG card (1200×630, logo centred on a solid brand-palette colour; gradients rejected) from the locked logo bytes — no LLM — commits both to the site repo and records provenance rows (`origin_model='derived-from-logo'`). `injectBrandHeadTags` in render_site_components injects favicon/OG/Twitter head tags fleet-wide, idempotently. Runs via a `brand_head` mode branch on asset-deployer (IMG-011), dispatched by a `needs_brand_head_assets` work item — a reusable pattern flagged as a candidate for auto-emit right after logo lock.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-25/#Turn-26/#Turn-27; SQL_2026-07-11_asset_deployer_brand_head_mode.sql
- **relations:** IMG-039 (logo permanence, its precondition); IMG-011 (asset-deployer agent it rides); IMG-041 (the manual workaround used where this action wasn't reachable)
- **verify-later:** derive_brand_head_assets action registration; asset-deployer check_mode branch; live favicon/og-card on robot-hands

### IMG-041 — Manual brand-asset commit workaround (brand-asset derivation gap)
- **status:** partial
- **status-evidence:** REPLICATION summary table (leopardess): favicon + OG-card derivation "[gap] — small; currently manual"; RUNBOOK H4 "resolved 2026-07-10 … deployed live … all verified 200 and byte-identical."
- **what:** Documents a site (leopardess) where favicon/OG generation from an approved logo was performed manually rather than through IMG-040's chassis action: background knockout to transparency, gold normalisation to brand hex, a flood-filled silhouette favicon for 16/32px, a multi-size .ico, an opaque apple-touch icon, and a 1200×630 OG card were all hand-derived. Delivery used `commit_brand_assets.sh`, sending the same git-adapter commit message `deploy_image_asset` would send, for pre-approved images with no generation step to spawn. `deploy_brand_asset.sh` notes asset-deployer never rewrites `assets.url` without an `asset_id` (a landmine for this workaround). This documents a genuine reconciliation point for stage 2: whether this predates IMG-040's chassis-wide action, or reflects a site/workflow that couldn't reach it at the time.
- **sources:** docs/leopardessconsulting/RUNNING_NOTES.md#Turn-9; RUNBOOK.md#O7; scripts/commit_brand_assets.sh, deploy_brand_asset.sh (headers)
- **relations:** IMG-040 (the chassis-native action this substitutes for); IMG-030 (kind routing, A6); IMG-054 (two-URL model)
- **verify-later:** internal/adapters/git/github_client.go CommitToRepo path prefixing; whether IMG-040's derive_brand_head_assets action predates or postdates this workaround chronologically

### IMG-042 — Header logo resolution from plan imagery
- **status:** deployed
- **status-evidence:** "header resolves the locked logo from plan imagery (logo-img live in the served header)" (2026-07-11); fix commit b00c150b.
- **what:** The header is a site component rendered by `render_site_components`, untouched by the page-level resolver fixes (IMG-051/IMG-058), and previously read the never-populated `sites.logo_url` — so sites showed a text mark despite a deployed logo file. Fix: `loadSiteDataFull` resolves the site-scope logo from `site_plan_imagery`→`assets` via `storage.DeployedWebPath` (never `assets.url`), keeping `sites.logo_url` only as a legacy fallback. Closed a long-standing "logo-in-header resolution gap" carried since 2026-05-27.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-23; FOCUS_imagery_assessment_1_.md#5.1 (gap origin); PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** IMG-058 (image-role resolver, the page-side sibling fix); IMG-054 (DeployedWebPath convention)
- **verify-later:** loadSiteDataFull logo resolution; served header `<img>` on fleet sites

### IMG-043 — Sprite-sheet bullets and list treatment (Phase I2)
- **status:** partial
- **status-evidence:** "Phase I2 ⏳ IN PROGRESS — the active phase … I2.0 ✅ … I2.1 ⏳ REGEN IN FLIGHT (Turn 31) … I2.2 NEXT BUILDABLE NOW" (2026-07-12/13).
- **what:** One coherent N×M glyph grid per site (3×3 @ 768², 256px cells; generated via Banana, harnessing the model's natural gridded-image tendency), sliced by CSS `background-position`; bullets/nav consume it via `::before` and `.sprite-<name>` classes — one generation, one asset, one stylesheet, no Go image cropping. A delivery-format question was resolved twice: sprite CSS ships as a separate committed `/assets/css/sprites.css` + head `<link>`, not through `css_snippets` (a GLOBAL library with no site scoping — the per-site committed bundle is the house pattern). Cell-content alignment is the main risk, mitigated by an ordered-grid prompt plus a human eyeball-and-assign gate (B11, `cell_names_verified` flag); vision auto-verify is deferred to I2.4/I8. The first generation was near-perfect (all 9 glyphs in reading order); its deploy failure spawned the ExtractActionInputs lesson (cross-category).
- **sources:** SCOPE_I2_sprite_sheets.md; CONTEXT_PACK_imagery_sprite_sheet.md; PLAN_imagery_best_in_class.md#Phase-I2; RUNNING_NOTES_imagery_best_in_class.md#Turns-28–32; SQL_2026-07-12_seed_robothands_sprite_sheet.sql
- **relations:** IMG-031 (five-place kind checklist, I2.0); IMG-040 (brand-head commit pattern reused for sprites.css); IMG-028 (flat-background lesson reapplied); IMG-029 (D6 dual icon strategy, the other half)
- **verify-later:** chk_kind includes sprite_sheet; sprite-sheet-main.png 768×768 on robot-hands; sprites.css emit action existence

### IMG-044 — Content-linked card imagery (Phase I3)
- **status:** aspirational
- **status-evidence:** "Phases I3–I8: not started" (2026-07-12); the card-crop approach was confirmed as a design decision on 2026-07-08.
- **what:** Every linking card (blog index, news feed, tool directory) would carry an image reflecting the content behind it, sharing a visual family with the content page. Confirmed approach: the card image is the article's asset re-cropped per purpose (one generation yields article hero, card crop ~800×450 WebP, OG crop) rather than a sibling generation. Would be the first real Lane B (content-driven, IMG-004) consumer, and would clear the one remaining empty image slot on robot-hands (learning-center-index listing card).
- **sources:** PLAN_imagery_best_in_class.md#Phase-I3; RUNNING_NOTES_imagery_best_in_class.md#Turn-2/#Turn-13
- **relations:** IMG-004 (Lane B, its architecture); IMG-045 (news imagery reuses its mechanics); IMG-048 (performance budgets set the card byte ceiling ≤60KB)
- **verify-later:** card kind/purpose in ImagePurposes (expected absent yet)

### IMG-045 — News imagery (Phase I5)
- **status:** aspirational
- **status-evidence:** I5 "not started"; the freshness rule was confirmed as a design decision 2026-07-08 ("no SLA … configurable news_image_grace_interval … working suggestion 6h").
- **what:** News ingestion would attach a per-item imagery decision via a small LLM classification: `chart` (data-driven story → IMG-046 pipeline), illustration/photo (→ IMG-044's pipeline), or none. Feed cards and article pages would share the artefact. No SLA is planned (ingest ~2×/day); after a configurable grace interval an item would fall back to a brand-kit-derived default image so the feed never shows an empty slot.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I5; RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** IMG-046 (data-graph pipeline, its data-story branch); IMG-044 (card imagery, its photo/illustration branch); IMG-038 (brand kit, the fallback source)
- **verify-later:** none yet — design only

### IMG-046 — Data-graph / chart pipeline (code-rendered, never diffusion) (Phase I4)
- **status:** aspirational
- **status-evidence:** "Status: scoping only — not built" (2026-05-20); "I4 not started" (2026-07-12); the runtime decision (go-echarts in-chassis) was confirmed 2026-07-08.
- **what:** A hard constraint drives this: diffusion models cannot plot real data — they fabricate values. Charts must instead be a separate three-stage pipeline: fetch real series (EIA/FRED/per-vertical free-tier sources, stored for reproducibility + attribution) → code-render (go-echarts; a static SVG/PNG must always exist as fallback) → an LLM editorial layer only for titles/callouts/annotations, never data values. Would need a `data-chart-generator`-shaped agent, and deliberately does NOT add `chart` to the `site_plan_imagery` kind enum — charts are a Lane B (IMG-004) artefact. `infographic` stays decorative-Banana and must never carry real numbers. See CHRT-001 (register/data-charts.md) for the later, more concrete Go SVG-emitter design that operationalises this same constraint.
- **sources:** old/FUTURE_data_graph_pipeline.md; PLAN_imagery_best_in_class.md#Phase-I4/#D1/D3; TODO_imagery_followups.md#Future-workstream (U10); imagery/old/TODO_imagery_followups(15).md#future-workstream-data-graphs, imagery/old/PLAN_imagery_phase_2g(1).md#what-2g-doesnt-include (U24a)
- **relations:** IMG-045 (news imagery consumes it for data-driven stories); CHRT-001 (the concrete implementation plan, in the new:data-charts register); IMG-012 (infographic kind, deliberately kept separate from chart)
- **verify-later:** no chart pipeline code expected in the imagery adapters; go-echarts dependency presence in go.mod; FUTURE_data_graph_pipeline.md (live); infographic-generator agent (expected absent)

### IMG-047 — Product illustration pipeline (copyright-safe sketches) (Phase I6)
- **status:** aspirational
- **status-evidence:** "Planned but not built (design work already done — reuse it)" (2026-07-08); "I6 not started."
- **what:** Would generate stylised product illustrations to avoid copyright/trade-dress exposure from scraped affiliate photos: a discovery check `check_product_without_custom_illustration` (per-pass cap ~20), a `product-illustration-handler` agent delegating to image-build-handler, a `link_asset_to_product` action setting `affiliate_products.custom_image_id`, with renderer precedence custom_image_id → cached_image_url. Stylisation is a hard-coded constraint, not a knob (D7): medium varies by product category (CAD-like / pencil / watercolour), with an altered viewpoint, an in-context setting, and no brand markings; img2img from the cached photo is explicitly v2-only under the derivative-work framing. 3D product reconstruction was explicitly parked as unnecessary once sketches cover the need.
- **sources:** old/illustration/PLAN_product_illustration.md; PLAN_imagery_best_in_class.md#Phase-I6/#D7; STATUS_imagery_2026-05-12.md#Component-audit-finding
- **relations:** affiliate sites programme (resolver dependency, cross-category); IMG-004 (Lane B instance); product components' query.affiliate_products socket
- **verify-later:** affiliate_products.custom_image_id usage; existence of the handler agent (expected absent)

### IMG-048 — Image performance budgets (Phase I7 / D8)
- **status:** aspirational
- **status-evidence:** Budgets "confirmed as proposed" 2026-07-08 (hero ≤180KB, card ≤60KB, sprite ≤80KB, index above-fold ≤600KB); "I7 not started."
- **what:** Per-kind byte and dimension budgets would be enforced at deploy time (extending ImagePurposes with ceilings; WebP for photographic kinds) and policed by a new `image_weight_over_budget` discovery check routed to asset-deployer for re-optimisation; responsive srcset + lazy loading in image-bearing templates; alt text required at generation time plus an `image_alt_text_missing` check; sprites amortise small art into one download by design.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I7/#D8; RUNBOOK_imagery_best_in_class.md#B5
- **relations:** OptimizeImageForWeb/ImagePurposes (existing enforcement point it extends); accessibility goal G8; IMG-044 (card byte ceiling)
- **verify-later:** ImagePurposes byte-ceiling fields (expected absent)

### IMG-049 — Reference-image style anchoring
- **status:** partial
- **status-evidence:** Item 24 decision 2026-05-18 ("plumb reference-image AND switch icon model"); Banana native path live via style guide reference_asset_keys (I1, 2026-07-11); IP-Adapter/img2img paths not built.
- **what:** Conditioning generations on a reference image for sibling consistency. Three techniques were ranked (img2img subject-anchor; IP-Adapter style-anchor — not on Stability's standard REST endpoint; LoRA highest fidelity) alongside three reference-provenance options (generate-one-then-derive; per-site curated style library; system-wide per-kind house style). What actually shipped is the Banana-native form: approved reference assets resolved to stable `s3://` URIs flow as style anchors for photographic kinds (via IMG-038's imagery_style_guide). Schema hooks (`reference_image_uri` field, `origin_asset_id`, `alterations` JSONB) exist ahead of the fuller img2img/IP-Adapter/LoRA paths.
- **sources:** TODO_imagery_followups.md#24; PLAN_imagery_best_in_class.md#D4; RUNNING_NOTES_imagery_best_in_class.md#Turn-17
- **relations:** IMG-038 (imagery_style_guide, the shipped mechanism); IMG-021 (adoption mirror as a reference source); IMG-047 (product-sketch img2img, v2)
- **verify-later:** Banana provider reference handling; whether any non-Banana reference path exists

### IMG-050 — Per-vertical LoRA fine-tunes
- **status:** aspirational
- **status-evidence:** "Status: planned, not started … the fine-tuning plan presupposes an adapter that can take a model field; our adapter cannot" (assessment); best-in-class plan explicitly keeps it "as a future consistency upgrade once I1's reference-image approach shows its limits."
- **what:** A plan to train per-vertical image LoRAs (60-90 curated, captioned images, SDXL/PixArt base preferred over FLUX for clean diagrams, roughly £35-95 for a first pass) for consistent visual style — e.g. veterinary/biological illustration (anatomical cross-sections, pathway diagrams, procedure illustrations, infographics for the canine-biology vertical, 018), or per-vertical energy infographics. Served via serverless (Replicate/RunPod) rather than an in-cluster GPU initially. Blocked historically on the adapter's model-selection plumbing; now deliberately deprioritised behind the cheaper reference-image approach (IMG-049).
- **sources:** FOCUS_imagery_assessment_1_.md#8; PLAN_imagery_loop_closure.md#Open-items; PLAN_imagery_best_in_class.md#6 (U10); docs023.../018_canine_biology.md#7 (U22)
- **relations:** IMG-049 (reference-image anchoring, the substitute); canine-biology vertical (cross-category); model-infrastructure training (cross-category)
- **verify-later:** none — not started

### IMG-051 — Per-page hero resolver + flag_page_image_rebuild rebuild trigger (baked-fallback fix, June 2026)
- **status:** deployed
- **status-evidence:** "the fixes are in production" (2026-06-02, HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md); independently corroborated: "the 'hero resolver' (June-02) is about hero IMAGES, not CTA URLs" (running_notes_16, explicitly disambiguating it from a costly earlier CTA-URL bug conflation). Delivered as edited plan_sections_action.go + new flag_page_image_rebuild_action.go (2026-05-27).
- **what:** Root-cause triad diagnosed across multiple sites: (1) `site_plan_imagery` plans imagery per page, but `store_asset` wrote `content_data[<purpose>_url]` keyed only by purpose, so every page's hero overwrote the single site-wide `hero_url` last-write-wins; (2) the first page render happens before imagery generation completes, so the `on_missing` static fallback path (`/assets/images/hero.jpg`) gets baked into `rendered_html`; (3) the terminal rerender path patches stored HTML without re-resolving, so a later-landed asset never reaches an already-rendered page. Fix: `ensureAssets` in `plan_sections_action.go` resolves this page's hero via a `site_plan_imagery` JOIN `assets` (page scope) and the site logo from site scope; a new `flag_page_image_rebuild_action.go` flags the page `needs_rebuild` and emits `needs_page` at priority 99 (dedup key `page_rerender:<page>`) as a terminal image-build-handler step, so the page re-resolves *through* `plan_sections` once its asset lands — closing the asset→render timing loop for page-scoped imagery. The logo/header path was deliberately left out of scope here (it renders via `render_site_components`, not `plan_sections` — see IMG-042 for its later fix).
- **sources:** HANDOFF_2026-06-02_hero_resolver_and_section_data_reconciler.md#1 (U02); 016 §9 deployed-hero-images entry, 002(4) flag_page_image_rebuild → page_rerender (U01); running_notes_16_content_quality_and_internal_linking(1).md#part-1, NOTES(44) Part A, HANDOFF_2026-06-09(2).md#june-02-actions (U05); FOCUS_imagery_assessment_1_.md#5.1-Decision, SHOWCASE_technical_architecture.md#3, TODO_imagery_followups.md#12 (U10)
- **relations:** IMG-012 (site_plan_imagery, the table this resolver joins); IMG-052 (the legacy content_data shadow this fix coexists with); IMG-055/IMG-056 (the section-scope extension of this same resolver); IMG-058 (the July follow-on rewrite covering the remaining incompatible patterns); IMG-060 (the general "rerender reassembles" principle this fix routes around, page-scope only)
- **verify-later:** registry.go entries for flag_page_image_rebuild/reconcile_section_data; hero component input_schema (field vs hardcoded template); ensureAssets page-aware behaviour on current code

### IMG-052 — Legacy site-level hero_url shadow (content_data last-write-wins)
- **status:** deployed
- **status-evidence:** "hero_url lives at SITE level (sites.content_data …), merged beneath section data; template `or` picks it over the presigned background_image … last-write-wins per purpose → site-wide hero currently = hero-about's image everywhere" (idea.uk investigation, 2026-07).
- **what:** A legacy mechanism, still present after IMG-051's June fix, that both once saved renders and continues to distort them: image deploys historically wrote `purpose+"_url"` keys (e.g. `hero_url`) into site-level `sites.content_data`, which the ContentData-priority merge (`component_library.go` ~line 736) supplies to templates *ahead of* the schema-resolved `site_assets.hero` value. Consequence: hero renders stayed immune to presigned-URL expiry (they're local-path), but every page on an affected site shows the same last-written hero image, with the per-page hero assets generated correctly but sitting unconsumed. Banked as a known quirk — per-page heroes fully winning over this shadow was left as a later improvement — and is render-neutral with respect to the presigned-URL localisation fix (IMG-053).
- **sources:** running_notes_scheme_to_components(55).md#Tx/#Ty/#Tz/#Ub; w9_02_deployer_and_shadow.sql; w8_09_hero_exposure.sql
- **relations:** IMG-051 (the resolver this shadow undermines); IMG-053 (presigned-URL fix, render-neutral to this); IMG-058 (the July resolver rewrite that fully defeats this shadow via a "designed authoritative overlay")
- **verify-later:** sites.content_data hero_url keys; component_library.go merge priority; whether per-page heroes fully win on any site yet

### IMG-053 — Presigned-URL expiry and deploy-time asset localisation (Edit F + w9_04 backfill)
- **status:** deployed
- **status-evidence:** "w9_06: t/f/f both pages … localisation verified on content; Edit F deployed → recurrence prevention live → THREAD CLOSED" (2026-07-06); "w9_04 RETURNING 'UPDATE 18, every url now /assets/images/…'."
- **what:** A whole failure class on idea.uk: `assets.url` stored the presigned B2/S3 URL from generation (`X-Amz-Expires=604800` — dies in seven days), while the asset-deployer had already committed the optimised file into the site repo under a key-derived local name. Renders resolving from `assets.url` therefore embedded URLs that die; heroes had escaped only by being shadowed by the legacy local path (IMG-052). Two-sided fix: `w9_04` backfill flipped all 18 idea.uk rows to `/assets/images/<key-hyphenated>.jpg`, preserving the unsigned S3 object path into a new `storage_path` (+`storage_provider`) column, followed by a rebuild; `Edit F` made `deploy_image_asset` record the committed local URL on the asset row at every future deploy, for ALL kinds, best-effort (a recording failure must never fail the deploy itself). Applies platform-wide to any site lacking the legacy hero_url shadow that had accidentally masked the bug.
- **sources:** w9_03_assets_schema_and_inventory.sql; w9_04_backfill_flip.sql; gobatch_03_deploy_asset_localise(1).md; running_notes_scheme_to_components(55).md#Tu/#Tw/#Tz/#Ua/#Ub/#Ud
- **relations:** IMG-052 (legacy shadow, render-neutral counterpart); IMG-054 (the general two-URL model this instance-fixes); storage-architecture (S3/B2 refs preserved in storage_path, cross-category)
- **verify-later:** deploy_image_asset_action.go post-commit UPDATE; assets rows url vs storage_path forms across sites

### IMG-054 — DeployedWebPath committed-path convention (the two-URL asset serving model)
- **status:** deployed
- **status-evidence:** "Verified: zero X-Amz URLs appear in any deployed page's HTML" (2026-07-10, after a wasted debugging turn); leopardess AUDIT_verified_facts D8 independently reaches the same conclusion and includes a documented self-correction ("an earlier 'asset deploy is broken platform-wide' alarm was withdrawn after re-investigation").
- **what:** Every generated image has two URLs by design: `assets.url` is a 7-day presigned S3 URL (SigV4 hard-caps expiry at 604800s — a throwaway source handle, never used to render), and the durable committed git path `/assets/images/<asset-key>.<ext>`, derived by `storage.DeployedWebPath(asset_key, purpose)` — the single source of truth shared by both the deployer and the resolver so the two paths cannot drift. Pages serve via GitHub Actions → Backblaze B2 → a Cloudflare Worker that re-signs each GET server-side. Debugging rule established from this: get the real asset_key/purpose from the DB and curl the derived path directly; a presigned URL sitting in `assets.url` is cosmetic staleness, not a broken image (roughly 83 stale presigned rows platform-wide are cosmetic-only; the w9_04-style url-flip backfill, IMG-053, is optional cleanup, not a fix requirement).
- **sources:** PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS; RUNNING_NOTES_imagery_best_in_class.md#Turn-8; SHOWCASE_technical_architecture.md#3 (U10); docs/leopardessconsulting/AUDIT_verified_facts.md#D8, RUNNING_NOTES.md#Turn-10, RUNBOOK.md#landmine-6/7 (U25)
- **relations:** IMG-011 (spawn-time storage env injection, same audit); IMG-053 (the instance-level presigned-expiry fix); IMG-058 (image-role resolver emits these paths); IMG-042 (header logo resolution uses the same convention)
- **verify-later:** storage.DeployedWebPath/AssetKeyFilename; deploy_image_asset_action.go url-flip (~line 250); assets.url vs storage_path across sites

### IMG-055 — Section-scope imagery pipeline (plan → emit → generate → deploy → rebuild) — idea.uk verification
- **status:** deployed
- **status-evidence:** "The flow already exists (05-26 handoff + code): write_site_plan → site_plan_imagery rows … → emit_imagery_items … → needs_imagery (priority ≤98) → image-build-handler … → flag_page_image_rebuild → needs_page (99) → rebuild resolves the URL"; exercised end-to-end (assets active in ~3 minutes each, 2026-07-03).
- **what:** A concrete, site-specific verification of the general imagery work-item chain (see IMG-064 for the umbrella concept) on idea.uk: the planner writes `site_plan_imagery` rows (scope site/page/section; `scope_ref` `page:ordinal` for sections; key; kind hero/logo/icon/illustration/infographic; authored prompt — the table has NO description column; ordering; source), the gap-driven `emit_imagery_items` emits `needs_imagery` items only where no asset exists, image-build-handler's 25-step workflow generates, stores, brand-checks, and spawns the asset-deployer (download S3 → optimise by purpose → commit to the site git repo, key-named files with `_`→`-`.jpg), then `flag_page_image_rebuild` emits `needs_page` so `plan_sections` re-resolves the now-present asset. Concrete finding: on idea.uk the pipeline never fired for a brief-explanation illustration simply because the planner never requested one (16 rows: 5 heroes, 10 icons, logo) — not a pipeline defect. Hygiene note: ordinal-based `scope_ref`s drift when plans reorder; resolution is by key, not ordinal.
- **sources:** RUNBOOK_scheme_to_components(50).md#W7/#W7-0.3/0.4-RESULTS; w7b_01_imagery.sql; running_notes_scheme_to_components(55).md#Th/#Tj/#Ty/#Tz
- **relations:** IMG-064 (the general umbrella chain this instance verifies); IMG-053 (presigned-URL expiry, discovered during the same investigation); IMG-056 (the resolution gap this pipeline exposed for section-scope imagery)
- **verify-later:** site_plan_imagery schema + emit_imagery_items step in build-site-planner; image-build-handler workflow steps

### IMG-056 — ensureAssets scope-resolution gap: hero/logo-only surfacing blocks section/illustration imagery (Edit B / kind-alias fix)
- **status:** deployed
- **status-evidence:** Observed as an unresolved gap on vonc 2026-07-02 ("resolver ensureAssets only surfaces hero/logo, so site_assets.illustration can't reach them"); fixed on idea.uk (~2026-07-06, "Edit B LIVE (tools t)"; "BRIEF-EXPLANATION CLOSED … illustration renders on index AND tools; Edit B fine throughout"); the same underlying resolver/kind-alias mechanism reused successfully on the minilobby build (2026-07-11, "src filled").
- **what:** The structural gap that made section/illustration imagery unresolvable across at least three sites: `ensureAssets` (`plan_sections_action.go`) loaded only the page hero and site logo into the resolver's assets map (plus a legacy content_data fallback), so `site_assets.<key>` for section-scope imagery — or any `site_assets.illustration`-sourced field — could never resolve even when the illustration asset existed and was deployed (illustration_game_master, illustration_gauntlet_cta, etc. on vonc, confirmed active with deployed files). The interim workaround on vonc was making such fields optional (text-only render). The proper fix (Edit B, idea.uk) adds a third query (site_plan_imagery scope='section', scope_ref LIKE page||':%', joined to active assets), mapping BOTH by key (per-key schema paths, e.g. icon sets) AND by kind first-wins alias (generic `site_assets.illustration` paths) — modelled on the existing hero block. A two-day "index miss" observed right after this deployment turned out to be a probe artefact (grepping for the asset key string, when storage objects are UUID-named) — a debugging lesson, not a regression. The same kind-alias resolution mechanism was later exercised again on minilobby: a schema field sourced `site_assets.illustration` resolves via `site_plan_imagery` kind-alias rows to a deployed asset path, with stale pre-resolver renders fixed by a light `page_rerender` (reason `image_landed`, no LLM); the archetype hub separately consumed 8 orphaned icons via page-scope hero alias rows.
- **sources:** gobatch_01_plan_sections.md#Edit-B; RUNBOOK_scheme_to_components(50).md#W7-CODE-FINDING; running_notes_scheme_to_components(55).md#Ti/#Tt/#Tu/#Tv (U03); docs/RUNNING_NOTES_vonc(36).md#2026-07-02/#2026-07-03, RUNBOOK_phase2_provocation_js(29).md#appendix-f, HANDOFF_2026-07-09_vonc_spark_minilobby_trim(2).md#§9 (U23); docs/social001_vonc_tiktok_social/minilobby_task/RUNNING_NOTES_minilobby_task.md#2026-07-10/11/12, VERDICT_minilobby_trim_method.md#8.5a, tool_docs/NOTES_brief-explanation(5).md (U25)
- **relations:** IMG-055 (the pipeline this gap sits inside); IMG-051 (the page-scope sibling fix this extends to section scope); IMG-059 (image_source_unsatisfiable, the platform-wide guarantee this gap motivated); IMG-058 (the July robot-hands resolver rewrite, a parallel/possibly-overlapping fix — reconcile in stage 2 whether these are the same code path)
- **verify-later:** plan_sections_action.go ensureAssets section query and kind-alias logic; rendered brief-explanation `<img>` src on index/tools; whether ensureAssets extension here and IMG-058's ImageRoleForPath are the same or parallel fixes

### IMG-057 — flag_page_image_rebuild section-scope mapping (Edit H)
- **status:** partial
- **status-evidence:** RUNBOOK 07-06 night: "Slice 4c: PENDING APPLY — gobatch_05_flag_section_scope.md now carries Edit H … AND Edit H2"; the companion step-description SQL landed ("4c step description UPDATE 1") but the code edit awaited the next commit/image at time of writing.
- **what:** `flag_page_image_rebuild` no-ops for non-page scope, so section-scope imagery landings never triggered the page rebuild that would surface them (observed live: zero flag-created `needs_page` in 30h after two illustrations landed on idea.uk). Section `scope_ref`s carry the page as a prefix (e.g. `index:1`), so the fix is a prefix-split: map scope 'section' to its page and fall through to the existing page path — no new emit code required. Edit H2 + slice4c additionally align the file header comment and the agent-definition step description with the new behaviour (a cosmetic-drift discipline: descriptions must match deployed behaviour).
- **sources:** gobatch_05_flag_section_scope.md; slice4c_step_description.sql; running_notes_scheme_to_components(55).md#Tp/#Ui/#Uj
- **relations:** IMG-051 (flag_page_image_rebuild, the mechanism being extended); IMG-056 (the section-scope resolution gap this rebuild-trigger fix pairs with)
- **verify-later:** flag_page_image_rebuild_action.go deployed body; image-build-handler flag_rebuild step description text

### IMG-058 — Image-role alias resolver + authoritative overlay (Phase I0 hero-resolution rewrite, July)
- **status:** deployed
- **status-evidence:** "RESOLVED 2026-07-10 — the finding took THREE fixes, all deployed/applied … 16 distinct hero files, each referenced by exactly one page."
- **what:** A later (robot-hands, July) discovery that there was no single contract for how a component gets its image — three incompatible patterns coexisted: the legacy site-wide `content_data.hero_url` field (IMG-052); preset `site_assets.background/product_screenshot/...` sources that nothing generates; and components with no image slot at all — meaning per-page heroes generated correctly but never rendered (same placeholder everywhere, or an empty src). Fix (three parts): `imageryplan.ImageRoleForPath`, a shared alias table normalising generic image field names to the "hero" role; a page-aware `ensureAssets` resolving page hero → site hero fallback; and `planSection` injecting the resolved hero under legacy alias keys into `resolved_data`, which the renderer merges LAST — the designed authoritative overlay that defeats the stale site-wide `hero_url`. Planner-side key alignment was explicitly rejected as structurally impossible (the component is selected after planning, so the planner can't know its field names in advance).
- **sources:** FOCUS_imagery_assessment_1_.md#5.1; RUNNING_NOTES_imagery_best_in_class.md#Turns-5–10; PLAN_imagery_best_in_class.md#I0-FINDING
- **relations:** IMG-051 (the June page-scope fix this rewrite supersedes/extends); IMG-052 (the shadow this overlay defeats); IMG-059 (image_source_unsatisfiable, this fix's forward guarantee); IMG-042 (header logo, the sibling fix outside this resolver's scope); IMG-056 (possible overlap — reconcile in stage 2)
- **verify-later:** imageryplan package + test; plan_sections_action.go resolve()/ensureAssets/planSection

### IMG-059 — image_source_unsatisfiable discovery check
- **status:** deployed
- **status-evidence:** Registered 2026-07-09/10; "live but has produced 0 flags (heroes all resolve now) — expected."
- **what:** Flags any component `input_schema` image field sourced from a `site_assets.<path>` that no asset key, plan imagery row, or image-role alias can supply — a systematic guarantee that the empty-src class of failure (IMG-056, IMG-058) is caught on every future domain instead of relying on eyeballing. Flag-only (`needs_human_review`, no automated handler), deduped per site/page/function/path, capped 25/pass. Substitutes for the structurally-impossible planner-side guard (component chosen after planning); shares the alias table with the resolver so the two mechanisms cannot drift apart.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-7; SQL_2026-07-09_register_image_source_unsatisfiable.sql
- **relations:** IMG-058 (image-role resolver, shares its alias table); IMG-061 (orphaned assets, the flag-side complement); services-hero orphan case
- **verify-later:** check_image_source_unsatisfiable.go; design-discovery-agent run_checks

### IMG-060 — Rerender reassembles, it does not re-resolve
- **status:** partial
- **status-evidence:** FOCUS §5.1 root cause 3 (2026-05-27); a 2026-07-10 investigation flagged what looked like a "NEW, SEPARATE ISSUE — rerender-completeness gap" but later narrowed the specific instance to corrupted templates; the underlying property nonetheless stands, restated as: "needs_rerender and the colour/CSS fixers regex-patch stored rendered_html; they do not re-run plan_sections."
- **what:** A general architectural property, not a single bug: the terminal rerender path reassembles existing `rendered_html` rather than re-running section resolution, so values that later land in `content_data` (resolved images, alt text) never reach the HTML without a full page rebuild through `plan_sections`. `flag_page_image_rebuild` (IMG-051) routes page-scoped imagery specifically around this limitation, but site-level components (header/footer) and non-hero inline sections remain exposed to it. Also noted as a corollary: `page_components.rendered_html` is a snapshot, not a view — template changes don't reach deployed pages without a rebuild either.
- **sources:** FOCUS_imagery_assessment_1_.md#5.1; RUNNING_NOTES_imagery_best_in_class.md#Turn-10/#Turn-11; HANDOFF_robot_hands_rebuild.md#Also-watch
- **relations:** IMG-051 (flag_page_image_rebuild, the routed-around exception); corrupted templates (a misdiagnosis neighbour, cross-category); styling-render-pipeline (cross-category)
- **verify-later:** which paths re-run plan_sections vs regex-patch rendered_html

### IMG-061 — Orphaned generated assets (component consumes nothing)
- **status:** partial
- **status-evidence:** "services-hero template has no `<img>`; the generated hero_services.jpg is never consumed" (2026-07-09); a full orphan check remains "deferred"; the services-imagery-coverage question has been open since 2026-05-15.
- **what:** Generation waste in two forms: the plan requests imagery no selected component can display (e.g. a services hero with no template slot), or assets outlive replans (stale sprite-sheet-main.jpg clutter, superseded May-era assets). Detection is only partially covered today by `undeployed_assets` and `image_source_unsatisfiable` (IMG-059) — a dedicated orphan/asset-supersession check and a repo cleanup pass remain open, as does the planner-side question of which pages should get heroes at all.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-5/#Turn-31; TODO_imagery_followups.md#21; PLAN_imagery_best_in_class.md#I0-finding-(c)
- **relations:** IMG-059 (image_source_unsatisfiable, the flag-side complement); IMG-033 (planner key stability, another orphan source)
- **verify-later:** assets rows with no referencing rendered_html; site-repo files with no assets row

### IMG-062 — Components declare imagery contracts / many-images-per-page direction
- **status:** aspirational
- **status-evidence:** Phase 3.5 was "NEW for the refresh" while the parent Phase 3 was deferred; framed as "Future direction — pages with many images (post-Phase-3)."
- **what:** `content_components.input_schema` v2 already supports typed image fields with arbitrary `source` resolvers, but only `hero_image` actually uses it. The direction: components would own their imagery contracts (team-grid declares `member_avatars` arrays; services-grid declares per-service icons), the renderer resolves scoped `site_plan_imagery` rows by asset_key, discovery walks the declared gaps, and generation scales horizontally (a 30-image page becomes 30 work items through the unchanged chain). Would enable per-image audit and retire silent fallthroughs to `/assets/images/hero.jpg`. The features `icon` string contract being one-sided (declared but never wired to a renderer) is the standing counter-example, resolved separately via Lucide (IMG-029) rather than this direction.
- **sources:** PLAN_imagery_loop_closure.md#3.5/#Future-direction; FOCUS_imagery_assessment_1_.md#4.2/#9-item-5
- **relations:** IMG-004 (Lane B); IMG-044 (card imagery); contracts-and-standards (input_schema slot specs, cross-category)
- **verify-later:** any component beyond hero_image with resolved image declarations

### IMG-063 — Human taste-gate operating model (runbook rituals)
- **status:** convention
- **status-evidence:** RUNBOOK structure in active use: standing rituals A1-A5 + a one-off queue B1-B11, most B-items closed with dates; "Humans only at the taste layer" (showcase doc).
- **stage2-verified (2026-07-14):** deployed → convention — Runbook operating model / division-of-labour doctrine (humans-at-taste-layer), no built artifact claimed; author's own verify-later says 'process concept'.
- **what:** The imagery workstream's division of labour: agents do all authoring/migrations/deploy-prep; humans do credentials, backups approval, budget sign-off, and visual approval gates specifically — logo approval (IMG-039, once, then locked), sprite-sheet cell verification (IMG-043, assigning true meanings after generation), and sampled page eyeballs at each phase's acceptance gate. These gates are deliberately the most expensive part of each phase; generation is never trusted to self-judge taste.
- **sources:** RUNBOOK_imagery_best_in_class.md; SHOWCASE_imagery_workstream.md#Why-it's-interesting; SCOPE_I2_sprite_sheets.md#Phasing
- **relations:** IMG-039 (logo approval gate); IMG-043 (sprite eyeball gate B11); hitl category (broader HITL machinery, cross-category)
- **verify-later:** n/a — process concept; check runbook item states for current status

### IMG-064 — Imagery work-item economy end-to-end chain
- **status:** deployed
- **status-evidence:** SHOWCASE technical architecture (verified against production 2026-07-10): full diagram planner → site_plan_imagery → needs_imagery → image-build-handler → provider adapter → store_asset → asset-deployer → flag_page_image_rebuild → resolver → rendered page; "~90s prompt → git commit."
- **what:** The consolidated, operating imagery pipeline as a single nameable chain, including its dedup property (a partial unique index lets discovery checks re-run forever without creating duplicate work) and its honest-state property (`mark_item_failed` rather than silent drops). This is the umbrella concept the individual phase concepts (IMG-005 through IMG-020, IMG-051 onward) compose into, and the shape any new imagery capability (sprites, cards, sketches) is expected to ride rather than reinvent. IMG-055 is a concrete, site-specific (idea.uk) verification of this same chain.
- **sources:** SHOWCASE_technical_architecture.md#2/#3; SHOWCASE_one_pager.md; PLAN_imagery_best_in_class.md#2
- **relations:** every imagery pipeline concept above; IMG-055 (concrete instance verification); development-guide work-item lifecycle (cross-category)
- **verify-later:** end-to-end trace on a fresh needs_imagery item
