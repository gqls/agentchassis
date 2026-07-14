# EXTRACTION U10 — docs024_key_docs_latest/imagery (imagery pipeline, phases 0–2H, best-in-class I0–I8, sprite sheets, product illustration, robot-hands rebuild)
Extracted 2026-07-13. Files in scope: 91. Concepts found: 69.

## Coverage
Paths relative to `docs/agent_docs/docs024_key_docs_latest/imagery/`.

| file | treatment |
|---|---|
| ANALYSIS_phase_2f_two_defects.md | full |
| CONTEXT_PACK_imagery_sprite_sheet.md | full |
| FOCUS_imagery_assessment.md | family-delta |
| FOCUS_imagery_assessment_1_.md | family-latest |
| FUTURE_section_data_handler_1_.md | full |
| HANDOFF_imagery_best_in_class.md | full |
| HANDOFF_robot_hands_rebuild.md | family-latest |
| PLAN_imagery_best_in_class.md | full |
| PLAN_imagery_loop_closure.md | family-latest |
| PLAN_imagery_phase_2g.md | family-latest |
| RUNBOOK_imagery_best_in_class.md | full |
| RUNNING_NOTES_imagery_best_in_class.md | full |
| SCOPE_I2_sprite_sheets.md | full |
| SHOWCASE_imagery_workstream.md | full |
| SHOWCASE_one_pager.md | full |
| SHOWCASE_technical_architecture.md | full |
| SQL_2026-07-08_robothands_mission_brief.sql | header-scan |
| SQL_2026-07-08_robothands_rebuild_prep.sql | header-scan |
| SQL_2026-07-09_register_image_source_unsatisfiable.sql | header-scan |
| SQL_2026-07-10_b7_layout_fix.sql | header-scan |
| SQL_2026-07-10_b7_layout_swap.sql | header-scan |
| SQL_2026-07-10_pagebuild_mark_item_failed.sql | header-scan |
| SQL_2026-07-10_register_component_template_corrupted.sql | header-scan |
| SQL_2026-07-10_robothands_imagery_style_guide.sql | header-scan |
| SQL_2026-07-11_asset_deployer_brand_head_mode.sql | header-scan |
| SQL_2026-07-12_add_sprite_sheet_kind.sql | header-scan |
| SQL_2026-07-12_asset_deployer_explicit_paths.sql | header-scan |
| SQL_2026-07-12_seed_robothands_sprite_sheet.sql | header-scan |
| STATUS_imagery_2026-05-08.md | family-latest |
| STATUS_imagery_2026-05-12.md | family-latest |
| TODO_imagery_followups.md | family-latest |
| old/001_image_model_comparison.md | full |
| old/ANALYSIS_phase_2f_two_defects.md | family-delta |
| old/FUTURE_data_graph_pipeline.md | full |
| old/HANDOFF_robot_hands_rebuild(1).md | family-delta |
| old/HANDOFF_robot_hands_rebuild.md | family-delta |
| old/PHASE_0_BUNDLE_README.md | full |
| old/PLAN_imagery_loop_closure(2).md | family-delta |
| old/PLAN_imagery_loop_closure(3).md | family-delta |
| old/PLAN_imagery_loop_closure(4).md | family-delta |
| old/PLAN_imagery_loop_closure(5).md | family-delta |
| old/PLAN_imagery_loop_closure(7).md | family-delta |
| old/PLAN_imagery_loop_closure(8).md | family-delta |
| old/PLAN_imagery_phase_2g.md | family-delta |
| old/README.md | full |
| old/README_news_pipeline.md | full |
| old/README_robot-hands-reboot.md | full |
| old/STATUS_affiliate_sites_2026-05-12.md | full |
| old/STATUS_imagery_2026-05-06.md | family-delta |
| old/STATUS_imagery_2026-05-08(1).md | family-delta |
| old/STATUS_imagery_2026-05-08(2).md | family-delta |
| old/STATUS_imagery_2026-05-08(3).md | family-delta |
| old/STATUS_imagery_2026-05-08(4).md | family-delta |
| old/STATUS_imagery_2026-05-08.md | family-delta |
| old/STATUS_imagery_2026-05-12.md | family-delta |
| old/TODO_imagery_followups(1).md | family-delta |
| old/TODO_imagery_followups(10).md | family-delta |
| old/TODO_imagery_followups(11).md | family-delta |
| old/TODO_imagery_followups(12).md | family-delta |
| old/TODO_imagery_followups(13).md | family-delta |
| old/TODO_imagery_followups(2).md | family-delta |
| old/TODO_imagery_followups(3).md | family-delta |
| old/TODO_imagery_followups(4).md | family-delta |
| old/TODO_imagery_followups(5).md | family-delta |
| old/TODO_imagery_followups(7).md | family-delta |
| old/TODO_imagery_followups(8).md | family-delta |
| old/TODO_imagery_followups(9).md | family-delta |
| old/TODO_imagery_followups.md | family-delta |
| old/illustration/PLAN_product_illustration.md | full |
| old/pageflow-builder_2026-05-12.txt | skipped-generated (53KB single-line agent-definition dump) |
| old/pageflow-builder_2026-05-12_NOTES(1).sql | full |
| old/phase1/check_image_url_404.go | header-scan |
| old/phase1/check_placeholder_image_in_use.go | header-scan |
| old/phase1/check_unfulfilled_image_prompt.go | header-scan |
| old/phase1/imagery_helpers.go | header-scan |
| old/phase1/phase_1_5_smoke_test_v2.sql | header-scan |
| old/phase1/phase_1_pre_migration_backup.sql | header-scan |
| old/phase1/phase_1_register_imagery_checks.sql | header-scan |
| old/phase2/2E/STATUS_imagery_2026-05-08.md | family-delta |
| old/phase2/2E/check_unfulfilled_image_prompt.go | header-scan |
| old/phase2/2E/imagery_helpers.go | header-scan |
| old/phase2/2E/phase_2e_deploy_image_asset_action.diff | header-scan |
| old/phase2/2E/phase_2e_image_build_handler_variant_path.sql | header-scan |
| old/phase2/2E/phase_2e_pre_migration_backup.sql | header-scan |
| old/phase2/2E/phase_2e_store_asset_action.diff | header-scan |
| old/phase_2g_step1_site_plan_imagery.sql | header-scan |
| old/planner_prompt_patch_imagery.md | full |
| old/verify_and_wire_lucide.md | full |

Family notes: the PLAN_imagery_loop_closure(2..8), STATUS_imagery_* and TODO_imagery_followups(1..13) families are progressive snapshots — the live copies are supersets; delta scans found one genuinely dropped concept (`check_legacy_image_prompts_aspect`, captured below as superseded) and the (2)-era "per-section granularity: deferred" decision that was later un-deferred (captured inside the 2G concepts). old/phase2/2E/STATUS and old/HANDOFF_robot_hands_rebuild(*) are strict subsets of their live counterparts.

## Concepts

### Imagery loop-closure programme (Phases 0–6)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Progress (updated 2026-05-14) — Phase 2G + Phase 2H are operationally verified end-to-end on robot-hands.com … Phases 3, 4, 5, 6 of the outer plan not started"; Phase 3 "⏸ deferred 2026-05-14".
- **what:** The sequenced master plan for closing the gap between what the planner/spec asks for in imagery and what is delivered: Phase 0 (wire unread data), 1 (algorithmic discovery checks), 2A–2H (schema + pipeline refactor + structured plan imagery + request shape), 3 (adoption mirror), 4 (text-only auditor awareness), 5 (vision LLM path), 6 (imagery-quality-auditor). Each phase shippable alone; LLM phases gated on algorithmic checks working.
- **sources:** PLAN_imagery_loop_closure.md#Progress, PLAN_imagery_loop_closure.md#Phase-summary-table, STATUS_imagery_2026-05-12.md#At-a-glance
- **relations:** superseded in spirit by the best-in-class programme (I0–I8) which renumbered to avoid collision; feeds imagery-quality-auditor, adoption image mirror.
- **verify-later:** phases table vs live code: `platform/orchestration/actions/generate_image_actions.go`, `discovery_checks/`, `assets` schema, image-build-handler agent_definitions row.

### imagery_direction prompt prepend (Phase 0.1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "0.1 — read imagery_direction ✅ delivered, verified 2026-05-08"; asset row shows origin_prompt beginning with the direction text.
- **what:** `generate_image` reads `site_specs.design_intent.imagery_direction` and prepends it to the subject prompt ("Style direction: … Subject: …", later unlabeled with a 200-char SDXL-aware sentence-boundary cap). Closed the "webdesign-agent writes imagery taste, image-generator ignores it" gap. Later superseded per-site by the imagery_style_guide when present (one brand voice, no double prepend).
- **sources:** PLAN_imagery_loop_closure.md#Phase-0, old/PHASE_0_BUNDLE_README.md, STATUS_imagery_2026-05-08.md#Today's-verification
- **relations:** imagery_style_guide brand guide; per-kind prompt gating (direction gated OFF icons/logos).
- **verify-later:** `getImageryDirectionForSite` / `composeImagePromptWithDirection` in generate_image_actions.go; assets.origin_prompt on recent generations.

### Asset provenance population (origin_prompt / origin_model)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "origin_model='sdxl' — Phase 0.2 column write happened" (2026-05-08); "origin_model propagation — assets.origin_model carries real provider/model" (2026-05-20 close-out).
- **what:** `store_asset` populates `assets.origin_prompt` (fixing a pre-existing bug where it was silently dropped — every row was NULL) and `assets.origin_model`; later extended so the adapter returns provider/model_id and workflows propagate it (`banana/gemini-3-pro-image-preview` vs `sdxl`) instead of a hardcoded literal. Provenance is the substrate for spec-vs-delivery audits.
- **sources:** old/PHASE_0_BUNDLE_README.md#Phase-0.2, TODO_imagery_followups.md#What-shipped-this-session, STATUS_imagery_2026-05-08.md
- **relations:** imagery discovery checks read it; imagery-quality-auditor (future) compares it to delivered image.
- **verify-later:** StoreAssetAction in v3_site_actions.go; `SELECT origin_model, origin_prompt FROM assets` distribution.

### Algorithmic imagery discovery checks (Phase 1 trio)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Phases 1.1–1.4 all "✅ delivered" 2026-05-08 (1.2/1.3 "partial — check fires; symptom-path site needed").
- **what:** Three Go discovery checks catch spec-to-delivery gaps without LLM cost: `unfulfilled_image_prompt` (planner asked, no asset), `placeholder_image_in_use` (fallback path rendered, no asset), `image_url_404` (HTML references an image no assets row backs, DB-only version). All follow the DiscoveryCheck interface, register via init(), and were appended to design-discovery-agent's run_checks. A longer wishlist of ~12 further checks (alt-text, dimensions, orphans, cross-site contamination, multi-image underfill) was catalogued and mostly remains unbuilt.
- **sources:** PLAN_imagery_loop_closure.md#Phase-1, FOCUS_imagery_assessment_1_.md#13.2, old/phase1/phase_1_register_imagery_checks.sql
- **relations:** check_unfulfilled_imagery_plan (2G successor for the new shape); image_source_unsatisfiable and component_template_corrupted (later siblings).
- **verify-later:** `platform/orchestration/actions/discovery_checks/check_*.go`; design-discovery-agent run_checks array.

### Asset locking (2A) and hard-vs-soft lock semantics
- **category:** locks
- **status-signal:** deployed
- **status-evidence:** "2A — assets.locked_at + locked_by ✅ delivered 2026-05-09"; docstring "v3 (final, applied 2026-05-08)".
- **what:** `assets` gains `locked_at timestamptz` + `locked_by text` + partial index, mirroring `page_components` exactly. Canonical lock model (settled after three docstring iterations): detection via `locked_at IS NULL`; classification hard (admin/admin-removed/checkpoint) vs soft (deploy/manual/auditor names) via `locked_by`; NO time-based expiry exists in production. Human uploads/locked assets are excluded from auditor queries and regeneration.
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2A, PLAN_imagery_loop_closure.md#2A, old/README.md
- **relations:** logo permanence (D5) is the first real consumer; timed lock-expiry project (deferred).
- **verify-later:** `check_component_lock.go`; assets table DDL; the store-asset lock guard `WHERE assets.locked_at IS NULL`.

### Timed lock-expiry project (deferred)
- **category:** locks
- **status-signal:** aspirational
- **status-evidence:** "Approved policy (2026-05-08): implement timed expiry as a focused future project… Sequenced after the imagery loop work completes."
- **what:** One migration adding `lock_type` + `lock_expires_at` to all four Pattern A tables (page_components, site_components, site_plan_directives, assets); auto-lock writers default from a policy table ('admin' permanent, 'deploy' timed/30, auditor approvals timed/90); ~8–10 callsite filter expansions; CheckComponentLock extended; new `expired_review_locks` discovery check. Restores the rhythm doc 004 v4 designed, of which only the audit-pass-counter-reset half shipped.
- **sources:** old/README.md, STATUS_imagery_2026-05-08.md#Lock-expiry-investigation, PLAN_imagery_loop_closure.md#Decisions
- **relations:** references LOCKS_should_locks_expire.md (outside this unit); asset locking 2A.
- **verify-later:** whether lock_type/lock_expires_at columns exist on any Pattern A table.

### asset_key multi-image model (Phases 2B–2D)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2B ✅ 2026-05-09; 2C ✅ deployed 2026-05-09; 2D ✅ applied 2026-05-10" with migration sanity checks documented.
- **what:** Broke the one-asset-per-purpose-per-site constraint: `assets.asset_key` column (backfilled from purpose), new unique index `(site_id, asset_key) WHERE active`, StoreAssetAction ON CONFLICT switched to asset_key, then the old `(site_id, purpose)` unique index dropped. Enables N heroes/icons/illustrations per site, with `(purpose, asset_key)` split (canonical hero = hero/hero; variant = hero/hero_about). Strict production apply order documented (2A→2B→2C deploy→verify→2D).
- **sources:** STATUS_imagery_2026-05-08.md#Phase-2B/2C/2D, PLAN_imagery_loop_closure.md#Phase-2, old/phase2/2E/phase_2e_store_asset_action.diff
- **relations:** hero-variant routing (2E) consumes it; DeployedWebPath derives filenames from asset_key.
- **verify-later:** `\d assets` indexes; StoreAssetAction ON CONFLICT target.

### Hero-variant routing through image-build-handler (2E)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2E ✅ delivered, verified 2026-05-12" with the full hero_about end-to-end trace.
- **what:** Made `hero_<page>` variants routable: `check_unfulfilled_image_prompt` classifies logo / hero_home / hero_<page> into needs_logo / needs_hero_image / unfulfilled_hero_variant; new `hasActiveAssetForAssetKey` helper (purpose-level check gave false positives); `deploy_image_asset` derives per-variant paths (`assets/images/hero-about.jpg`, `_`→`-`); StoreAssetAction gains `asset_key_field` JSONPath config; a third variant branch added to the image-build-handler workflow (spawn/call/store/deploy) leaving logo/hero branches untouched.
- **sources:** STATUS_imagery_2026-05-12.md#Phase-2E, old/phase2/2E/check_unfulfilled_image_prompt.go, old/phase2/2E/phase_2e_image_build_handler_variant_path.sql
- **relations:** needs_imagery branch (2G.5) later sits in front of it; known gap — variant chain doesn't pass site_id (imagery_direction not prepended for variants).
- **verify-later:** image-build-handler workflow branches in agent_definitions; imagery_helpers.go.

### Spawned asset-deployer deploy pattern / storage-env isolation (2F)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2F ✅ deployed + verified 2026-05-12"; boxed warning 2026-07-10 "Where deploys run (by design, don't 'fix' this)".
- **what:** The chassis pod deliberately carries NO `IMAGE_BUCKET`, so it builds no storage client and inline `deploy_image_asset` fails there ("storage client not available") — by design. Deploys run in a spawned `asset-deployer` child into which `spawn_actions.go` injects S3/B2 env via the `isStorageEnabledAgent` list. 2F replaced three inline deploy step pairs in image-build-handler with spawn+call pairs targeting asset-deployer. Hand-triggering asset-deployer standalone fails because it skipped the injection — a triggering mistake, not a bug.
- **sources:** PLAN_imagery_loop_closure.md#2F, PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS, STATUS_imagery_2026-05-08.md#[BLOCKER]-Storage-architecture-mismatch
- **relations:** brand_head mode rides the same agent; storage-architecture (doc 032).
- **verify-later:** `agentbase/agent.go:294`, `spawn_actions.go` isStorageEnabledAgent, 107_image_build_handler.sql:725 comment.

### Snapshot-shadowing agent-definition loader defect
- **category:** system-architecture
- **status-signal:** deployed
- **status-evidence:** "Snapshot-shadowing hypothesis confirmed via SQL… Two loaders patched… Builds deployed as v1.0.1006"; "Snapshot audit closed clean" (2026-05-12).
- **what:** `snapshot_agent()` inserts snapshots at version+1000, so any loader using `ORDER BY version DESC LIMIT 1` without filtering `is_snapshot`/`is_active` reads the snapshot instead of the active row — every snapshot silently shadows its agent until the loader is fixed. `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` were patched; other loaders were already correct. Structural residue: snapshot retention policy and a single AgentDefinitionRepository remain open.
- **sources:** ANALYSIS_phase_2f_two_defects.md#Defect-1, STATUS_imagery_2026-05-12.md#Loader-snapshot-defect
- **relations:** model-infrastructure snapshot/rollback (021_model_swap_and_rollback.sql); parked "deep discussion" trio with the consumer-group race.
- **verify-later:** grep `FROM agent_definitions` across Go for is_snapshot filters; snapshot row counts.

### Kafka per-spawn response-topic partition race
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** "did not reproduce on second run… monitor" (2026-05-11); HANDOFF 2026-07-12 "Kafka per-job response-topic partition race — transient; now surfaces as failed items (mark_item_failed fix) instead of silent successes."
- **what:** An adapter (git-adapter observed; kafka-go LeastBytes balancer) occasionally writes its response to partition 1 of a single-partition per-spawn topic, losing the reply — work succeeds but the orchestration times out/fails. Root cause suspected stale partition metadata for just-created topics. Never structurally fixed; consequence downgraded from silent-success to visible failed items by the mark_item_failed pattern. The same race killed a content-writer reply and produced the "no-op complete" anomaly.
- **sources:** ANALYSIS_phase_2f_two_defects.md#Defect-2, RUNNING_NOTES_imagery_best_in_class.md#Turn-16, HANDOFF_imagery_best_in_class.md#Open-threads
- **relations:** mark_item_failed error-honesty fix; consumer-group race (separate doc, chassis replicas=1).
- **verify-later:** `platform/kafka/producer.go` balancer; adapter logs for "topic partition not found".

### site_plan_imagery table (2G.1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.1 — site_plan_imagery table ✅ delivered 2026-05-12"; live `chk_kind` re-verified 2026-07-08.
- **what:** Sibling of site_plan_directives holding structured imagery requirements: scope (site|page|section) + scope_ref, key (→asset_key), kind CHECK enum (logo|hero|illustration|icon|infographic, later +sprite_sheet), required prompt, JSONB style_hints/constraints, ordering, source CHECK (llm|classifier|manual|adoption), lock columns with the same lock-transfer treatment, unique on (plan_id, scope, COALESCE(scope_ref,''), key). `product` deliberately excluded (products come from the affiliate_products resolver, not the planner). Kind enum is mirrored in Go (`validImageryKinds`) — constraint and mirror change together.
- **sources:** PLAN_imagery_phase_2g.md#Schema, old/phase_2g_step1_site_plan_imagery.sql, SQL_2026-07-12_add_sprite_sheet_kind.sql
- **relations:** planner imagery block writes it; check_unfulfilled_imagery_plan reads it; five-place new-kind checklist.
- **verify-later:** `\d site_plan_imagery`; chk_kind vs validImageryKinds in write_site_plan_action.go (~line 183).

### Planner imagery block (2G.3 prompt extension)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.3 ✅ delivered 2026-05-13 (with max_tokens bump + path fix)"; 2026-07-08 ground truth "planner prompt carries the Imagery Block + decomposition rule; max_tokens now 16000".
- **what:** build-site-planner's JSON output gains an `imagery` key (site[] / pages{} / sections{} entries with key, kind, prompt, optional style_hints/constraints) in the same LLM call as pages/design_direction — a single call, no separate imagery planner. Replaces the flat `image_prompts:{logo,hero_home}` contract that had hero/logo-only names baked in. max_tokens raised 4000→8000 (JSON truncation on a 14-page roadmap) and later to 16000. Legacy image_prompts continues to be emitted during transition.
- **sources:** PLAN_imagery_phase_2g.md#Planner-output-shape, PLAN_imagery_loop_closure.md#Application-status, FOCUS_imagery_assessment_1_.md#4.1
- **relations:** one-entry-one-image decomposition rule; planner key stability problem; sprite_sheet planner emission (future).
- **verify-later:** build-site-planner default_config prompt_template "## Imagery Block"; sql_for_agents/053 patches.

### flattenImageryBlock write path + imagery lock transfer (2G.2)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "✅ deployed 2026-05-12; path fix on 2026-05-13 (function looked up data['imagery'] at top level rather than walking wrapper shapes via findDirectiveTree)".
- **what:** `write_site_plan` walks the planner's imagery block and inserts site_plan_imagery rows in the same transaction as pages/sections/directives (`flattenImageryBlock` + `insertImageryRow` enforcing the kind enum), and transfers locks from the previous current plan's locked imagery rows matched on (scope, scope_ref, key) — locked HITL prompt edits survive plan rebuilds.
- **sources:** PLAN_imagery_phase_2g.md#write_site_plan-extension, PLAN_imagery_loop_closure.md#2G
- **relations:** site_plan_imagery table; content-governance lock semantics.
- **verify-later:** write_site_plan_action.go flattenImageryBlock/insertImageryRow; lock-transfer behaviour on replan.

### check_unfulfilled_imagery_plan discovery check (2G.4)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.4 ✅ delivered 2026-05-14 (8 work items emitted on first run; correct priority ordering)"; Pipeline:"build" fix confirmed in code 2026-07-08.
- **what:** Walks the current plan's site_plan_imagery rows, emits one `needs_imagery` work item per row lacking a matching active asset (via hasActiveAssetForAssetKey), capped at 20/pass, priority-banded (site logo 70 → index hero 65 → site other 75 → page hero 80 → page other 90 → section 100) mirroring legacy classifyPromptKey bands. `computeAssetKey` namespaces deeper keys (`page.about.illustration_team_values`, `section.home.2.icon_precision`) while keeping hero/logo names flat for backward-compatible deploy paths. Dedup key `needs_imagery:<scope>:<scope_ref|->:<key>`.
- **sources:** PLAN_imagery_phase_2g.md#Discovery-check-1, PLAN_imagery_loop_closure.md#Decisions/#2G, TODO_imagery_followups.md#7
- **relations:** legacy unfulfilled_image_prompt runs in parallel during transition (both call hasActiveAssetForAssetKey to avoid double work); pipeline-field fix.
- **verify-later:** check_unfulfilled_imagery_plan.go (hardcoded Pipeline "build"); design-discovery-agent run_checks.

### needs_imagery branch in image-build-handler (2G.5)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2G.5 ✅ delivered 2026-05-14 (with two hotfixes — optional input_mapping + store_asset purpose)"; first asset a12b5d71 through the new path 2026-05-14.
- **what:** A new branch alongside (not extending) the variant chain: check_item_type_imagery → spawn_image_gen_imagery → call_imagery_gen (site_id passed so imagery_direction prepends; kind/style_hints/constraints pass through) → brand-update conditional store → shared spawn_asset_deployer tail. Brand-asset update routed by a `spec.brand_update` boolean computed at discovery (site scope OR index-page hero). Hotfixes established the `?`-suffix optional input_mapping convention and exposed that store_asset lacked `purpose_field` (initially hardcoding purpose:"hero", blocking kind=logo items — later fixed by the purpose_field workflow fix, 2026-05-20). A future refactor option is recorded: collapse the three legacy branches into needs_imagery ("always fix legacy with modern").
- **sources:** PLAN_imagery_loop_closure.md#2G/#Step-5-workflow, PLAN_imagery_phase_2g.md#image-build-handler-extension, TODO_imagery_followups.md#What-shipped-this-session
- **relations:** hero-variant branch (2E); `?` optional-mapping convention; purpose_field fix.
- **verify-later:** image-build-handler workflow JSON; store_imagery_asset purpose_field config.

### Legacy image_prompts age-out check (check_legacy_image_prompts_aspect)
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** Old plan versions: "Migration off legacy image_prompts: Age-out via check_legacy_image_prompts_aspect, registered last"; live plan: "2G.6 ❌ retired as scoped. Reframed 2026-05-13… one string out of a JSON array, no code change."
- **what:** Originally a dedicated discovery check emitting `needs_replan` for sites still on `site_specs.site_plan.image_prompts` (deliberately registered LAST to avoid replan churn before the planner extension shipped). Reframed and retired: "is a site on legacy?" is not a fault signal — the existing checks already detect brokenness on both paths; migration became an operational deregistration decision (pull `unfulfilled_image_prompt` from run_checks once it reliably finds zero gaps).
- **sources:** old/PLAN_imagery_loop_closure(3).md#Decisions, PLAN_imagery_phase_2g.md#Discovery-check-2, PLAN_imagery_loop_closure.md#Decisions (2026-05-13 reframe)
- **relations:** superseded by "operational deregistration, not a dedicated check"; transition dual-check running.
- **verify-later:** confirm no check_legacy_image_prompts_aspect.go exists; whether unfulfilled_image_prompt is still registered.

### pageflow-builder retirement
- **category:** imagery
- **status-signal:** superseded
- **status-evidence:** "Decision marker: 2026-05-12 — user agreed pageflow-builder is being left behind"; snapshot saved to pageflow-builder_2026-05-12.txt.
- **what:** The legacy monolithic site builder (inline deploy_image_asset, hardcoded generate_logo/generate_hero_image, sequential 20-iteration page loop, writes site_specs.site_plan directly bypassing the plan-domain tables) is deliberately not extended with the 2G imagery shape. Architecture converges on build-site-planner/plan-builder + triaged work items + page-build-handler + image-build-handler. Sites it built stay on the legacy check path until they age out; a full row snapshot exists as the recovery reference. The classifier's `recommended_builder` default was a noted loose end.
- **sources:** PLAN_imagery_phase_2g.md#On-leaving-pageflow-builder-behind, old/pageflow-builder_2026-05-12_NOTES(1).sql, old/pageflow-builder_2026-05-12.txt
- **relations:** superseded by plan-domain + dispatch architecture; robot-hands rebuild dropped its recommended_builder key.
- **verify-later:** pageflow-builder agent_definitions row status; any remaining live traffic.

### Image-generator request shape + per-kind defaults (Phase 2H)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "2H ✅ delivered (action layer) 2026-05-14 partial — chassis confirmed; adapter binary unconfirmed"; later provider work (2026-05-20) shipped the adapter side.
- **what:** Extends the generation request beyond {prompt,width,height}: v1 fields negative_prompt, seed, reference_image_uri (pass-through), cfg_scale, steps; Go-side `kindDefaults` map per kind (logos get people/text/watermark negative prompts; icons tighter aspect; heroes unchanged) with caller spec overriding defaults; style_hints.aspect_ratio drives dimensions and constraints feed the negative prompt. style_preset/samples/safety_mode deferred. Defaults deliberately live in Go, not a config table.
- **sources:** PLAN_imagery_loop_closure.md#2H, STATUS_imagery_2026-05-12.md#Phase-2H-(proposed), TODO_imagery_followups.md#4
- **relations:** provider abstraction; parseAspectRatio whitelist fix; constraints "informational only" decision.
- **verify-later:** kindDefaults/resolveKind/parseAspectRatio in generate_image_actions.go; adapter field mapping in dynamic_adapter.go.

### parseAspectRatio SDXL v1.0 whitelist snap
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Elevated HIGH 2026-05-18 ("16:9 → 1024×576, SDXL rejects"); Turn 24 (2026-07-11) refers to "pre-SDXL-snap-fix residue", implying the fix landed.
- **what:** parseAspectRatio snapped to multiples of 64 rather than SDXL v1.0's strict dimension whitelist (1024×1024, 1152×896, 1344×768, …), so planner-emitted aspect_ratio hints produced rejected sizes and blocked hero generation — a regression enabled by the item-4 prompt patch (heroes previously fell through to valid kindDefaults). Fix: snap to the nearest whitelist pair matching the requested orientation.
- **sources:** TODO_imagery_followups.md#5, RUNNING_NOTES_imagery_best_in_class.md#Turn-24
- **relations:** Phase 2H request shape; planner prompt patch change 1 (aspect moved into style_hints).
- **verify-later:** whitelist logic in generate_image_actions.go; test 16:9→1344×768.

### Adoption image mirror (Phase 3)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "3 — adoption image mirror ⏸ deferred 2026-05-14. Not cancelled… reference_image_uri plumbing preserved as a forward-compat hook."
- **what:** Stop discarding crawled imagery on adoption: new `mirror_adoption_images` action (download crawl images, upload to S3, insert assets rows with origin_type='adopted', origin_url, attribution/license; caps 50 images/site, 5MB each), wired into apply_adoption_plan; backfill check `check_crawled_images_discarded` routed to a new one-step `adoption-image-mirror` agent. Adopted images become img2img/style references and auditor signals. Deferred because current adopted sites carry minimal imagery.
- **sources:** PLAN_imagery_loop_closure.md#Phase-3, FOCUS_imagery_assessment_1_.md#7/#9-item-9
- **relations:** reference-image style anchoring; adoption-pipeline category (site crawling).
- **verify-later:** existence of mirror_adoption_images_action.go, adoption-image-mirror agent row; assets rows with origin_type='adopted'.

### Visual auditor imagery awareness (Phase 4, text-only)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "4 — visual auditor sees imagery (text-only) — not started".
- **what:** Extend visual-design-auditor's load_design_context SQL to include assets rows (unlocked, generated/adopted), imagery_direction, and site_plan_imagery, and add IMAGERY as a sixth check category with algorithmic-check results passed through to avoid double-flagging; tune on 5–10 sites before enabling fixes (≥80% accuracy target). Today the auditor's context contains zero image data — it cannot notice a missing or off-brief hero.
- **sources:** PLAN_imagery_loop_closure.md#Phase-4, FOCUS_imagery_assessment_1_.md#13.1/#13.3
- **relations:** imagery-quality-auditor (option B chosen as eventual answer); design-composition auditors.
- **verify-later:** visual-design-auditor load_design_context SQL in agent_definitions.

### Vision-capable LLM path (Phase 5)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "5 — vision-capable LLM path — not started".
- **what:** Foundational capability for auditors to actually look at images: extend aiservice.AIService with image inputs, implement Anthropic vision content blocks, prefer extending execute_llm_prompt with an image_urls_field over a new action, refresh presigned URLs immediately before calls, tag vision_call:true in llm_call_log for cost separation.
- **sources:** PLAN_imagery_loop_closure.md#Phase-5
- **relations:** required by imagery-quality-auditor and sprite-sheet vision auto-verify (I2.4/I8).
- **verify-later:** aiservice interface; anthropic.go vision support.

### imagery-quality-auditor agent (Phase 6 / I8)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "6 — imagery-quality-auditor agent — not started"; I8 in the best-in-class plan also not started.
- **what:** A vision-capable sibling of visual-design-auditor under design-audit-agent, dedicated to imagery: categories direction_mismatch / brand_mismatch / inconsistency / quality / inappropriate; max_fix_attempts 2; findings route to image-build-handler regeneration (different prompt/seed/negative prompt) escalating to needs_human_review; honours locks and origin_type='uploaded'; gated rollout. I8 adds sprite-sheet cell verification and brand-guide reference comparison. Chosen over extending the existing auditor (separate TOP-5 cap; only imagery pays vision cost).
- **sources:** PLAN_imagery_loop_closure.md#Phase-6, FOCUS_imagery_assessment_1_.md#13.4, PLAN_imagery_best_in_class.md#Phase-I8
- **relations:** vision path (Phase 5); imagery_style_guide as the audit standard; improvement-loop pass caps.
- **verify-later:** no imagery-quality-auditor row in agent_definitions (expected absent).

### Image provider abstraction and kind→provider routing
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** 2026-05-20 close-out: "Provider abstraction — internal/adapters/imagegenerator/{provider,stability,banana}. dynamic_adapter.go routes by kind: icon → Banana, everything else → Stability. Proven end-to-end."
- **what:** The originally hardcoded Stability-only adapter (env-driven; the image-adapter.yaml/agent-definition config blocks are misleading and unread) was refactored into provider packages with kind-based routing: flat kinds (icon, later logo/illustration/infographic per a committed-but-then-pending routing change, and sprite_sheet) → Google Banana `gemini-3-pro-image-preview`; photographic kinds → Stability SDXL. The provider Request carries ReferenceImageURIs (Banana native reference-image support). Known opens: Stability provider timeout 60s vs old 120s; circuit breaker not threaded into provider clients.
- **sources:** TODO_imagery_followups.md#What-shipped-this-session, FOCUS_imagery_assessment_1_.md#1.1–1.3, PLAN_imagery_best_in_class.md#2
- **relations:** icon model lessons drove it; 2H request shape; multi-provider routing beyond two providers still deferred.
- **verify-later:** internal/adapters/imagegenerator/ package layout; adapter switch cases; timeout value.

### Icon generation lessons and image-model comparison
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Verdict (2026-05-18): Fix A is insufficient… SDXL ignores style instructions"; "Final icon batch state (verified 2026-05-20): all six icons banana/gemini-3-pro-image-preview, visual gate passed."
- **what:** SDXL is the wrong tool for flat-vector icons — strong photorealism bias on concrete subjects, multi-panel drift, no real transparency. A full model comparison (SDXL, SD3.5, FLUX schnell/dev/pro, DALL-E 3, Imagen 3, Nano Banana Pro 2, Midjourney, LLM-SVG) ranked FLUX schnell cheapest-good and Banana best for reference-conditioned sibling consistency; decision: plumb reference images AND switch icon generation to Banana. Related fixes: purpose_field so icons store as purpose=icon (240×240, not hero 1600×900); kindDefaults icon dimensions; jpg-vs-png note for thin line art.
- **sources:** TODO_imagery_followups.md#23, old/001_image_model_comparison.md, TODO_imagery_followups.md#Final-icon-batch-state
- **relations:** provider abstraction; transparency abandonment; LLM-SVG sleeper option; reference-image anchoring.
- **verify-later:** ImagePurposes["icon"]; assets rows origin_model for icon assets.

### LLM-generated SVG icon path (sleeper option)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Sleeper option for later: LLM-generated SVG for icons… Worth a focused experiment once the immediate work is shipped" (2026-05-18).
- **what:** Icons are vector by nature; an LLM (Claude/GPT) writing SVG markup directly bypasses the entire convince-a-diffusion-model problem at ~$0.001–0.005/icon, crisp at any size, no copyright concern. Was the analyst's original recommendation (c2) before the user chose the Banana route; retained as a possible future replacement of the raster icon pipeline. Implies per-kind generation pipeline routing.
- **sources:** TODO_imagery_followups.md#23 (options c1/c2, recommendation), FOCUS_imagery_assessment_1_.md#9-item-6
- **relations:** superseded for now by Banana raster icons; Lucide covers UI chrome regardless.
- **verify-later:** none (idea only).

### Diffusion transparency abandoned → flat-grey chip icons
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Transparency abandoned as too fragile — image models paint a transparency checkerboard into RGB (confirmed: icon_cycle_time mode=RGB has_alpha=False). Decision: option 2, 'embrace the box'" (2026-05-20).
- **what:** Image models cannot produce true alpha; requesting transparent backgrounds yields painted checkerboards. Locked decision: icons generate on a flat selectable grey background (#EEEEEE bg / #4A4A4A line) and are presented inside a styled CSS chip; the planner prompt and all existing icon specs were patched accordingly. The lesson recurs for sprite sheets (flat selectable background, NOT transparent).
- **sources:** TODO_imagery_followups.md#Icon-background-resolution, SCOPE_I2_sprite_sheets.md#3 (planner prompt), CONTEXT_PACK_imagery_sprite_sheet.md#Attach—docs
- **relations:** icon lessons; sprite-sheet prompt rules; CSS chip styling was left as site-template work.
- **verify-later:** planner prompt icon-background wording; icon assets' actual backgrounds.

### Per-kind prompt gating and the five-place new-kind checklist
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "directionAppliesToKind… gates design_intent.imagery_direction OFF icons/logos" (shipped 2026-05-20); "PROVEN on real generations (icons carried palette, not medium)" (2026-07-11); sprite gating fix commit 4629aa17 proven in origin_prompt (Turn 31).
- **what:** Photographic brand direction contaminates non-photographic kinds (prepending it to an icon prompt makes the model paint a photo around the icon), so prompt composition gates per kind: hero/illustration/infographic get medium+mood+palette; icon and sprite_sheet palette only; logo nothing. Two gating functions (`directionAppliesToKind`, `styleGuide.directionForKind`) plus the DB constraint, Go mirror, adapter switch and ImagePurposes form the standing five-place checklist any new imagery kind must touch — the I2.0 lesson.
- **sources:** TODO_imagery_followups.md#What-shipped-this-session, HANDOFF_imagery_best_in_class.md#Mechanisms, RUNNING_NOTES_imagery_best_in_class.md#Turn-29
- **relations:** imagery_style_guide supplies the gated content; sprite_sheet contamination near-miss is the cautionary case.
- **verify-later:** both gating functions list identical kind sets; grep for the five places when a new kind exists.

### One-entry-one-image decomposition rule (planner prompt patch)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** planner_prompt_patch changes applied ("planner prompt carries the Imagery Block + decomposition rule", verified live 2026-07-08).
- **what:** A work item describes one deliverable: prompts must never ask for "a set of six icons" (SDXL renders one six-panel image — unusable but superficially successful). The planner prompt teaches per-entry single-image prompts, bans plural/counting phrasing (RULE 16), biases toward over-decomposition (unused icons are cheap, botched multi-panels expensive), moves aspect ratio to style_hints.aspect_ratio (the key Go reads) and demotes constraints to "informational only, reserved". The icon_cross_technology six-panel artifact and its cleanup SQL are the canonical example.
- **sources:** old/planner_prompt_patch_imagery.md, TODO_imagery_followups.md#25/#4
- **relations:** planner imagery block; SDXL whitelist fix; multi-entry sections remain the canonical way to express multiple images at one scope.
- **verify-later:** RULE 16 in the live planner prompt; absence of "set of" in site_plan_imagery.prompt.

### Planner key stability across replans
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Symptom (2026-05-15): previous plan keyed hero_canonical; new plan called it brand_hero_canonical… discovery emitted a fresh work item"; no fix recorded since.
- **what:** The planner LLM freely chooses imagery `key` values, so replans rename equivalent concepts, discovery sees missing assets, and generations/orphan assets accumulate. Fix options ranked: (a) pass old plan's keys into the prompt with a reuse rule (lowest effort), (b) canonical key dictionary, (c) semantic concept matching at discovery time. Stale keys from this bug were cleaned up during the best-in-class rebuild.
- **sources:** TODO_imagery_followups.md#26, RUNNING_NOTES_imagery_best_in_class.md#Turn-24 (stale failed rows closed)
- **relations:** planner imagery block; replan-driven waste; asset orphan cleanup.
- **verify-later:** whether the planner prompt includes previous-plan keys; duplicate-concept assets on replanned sites.

### Lucide icon strategy and validator wiring
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Lucide validator written (lucide_icons.go) — NOT yet wired" (2026-05-20), re-verified unwired 2026-07-08; still listed as an open I0 close-out item.
- **what:** The features grid renders icons as Lucide webfont glyphs (`<i data-lucide="{{.icon}}">`), not generated raster — the generated icon pipeline was never the right tool for it. Missing icons are LLM-invented Lucide names. Fix design: a single-source allowlist that is both the prompt's choice list and a pre-store `SanitizeFeatureIcons` sweep, plus optional render-time net; the allowlist must be verified against the bundled Lucide version. Blocked on identifying the content-generation step that fills features content_data. Icon strategy stays dual (D6): Lucide for UI chrome, generated sprites/raster for decorative glyphs.
- **sources:** TODO_imagery_followups.md#features-component-icons, old/verify_and_wire_lucide.md, PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** sprite sheets cover decorative glyphs; robot-hands rebuild was to carry the wiring.
- **verify-later:** callers of SanitizeFeatureIcons/ValidateLucideIcon outside lucide_icons.go.

### Data-graph / chart pipeline (code-rendered, never diffusion)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Status: scoping only — not built" (2026-05-20); I4 "not started" as of 2026-07-12; runtime decision confirmed 2026-07-08 (go-echarts in-chassis).
- **what:** Hard constraint: diffusion models cannot plot real data — they fabricate values. Charts are a separate three-stage pipeline: fetch real series (EIA/FRED/per-vertical free-tier sources, stored for reproducibility + attribution) → code-render (go-echarts; static SVG/PNG always exists as fallback) → LLM editorial layer only (titles, callouts, annotations — never data values). Needs a `data-chart-generator`-shaped agent and deliberately does NOT add `chart` to site_plan_imagery kinds (charts are Lane B artefacts); `infographic` stays decorative-Banana and must never carry real numbers.
- **sources:** old/FUTURE_data_graph_pipeline.md, PLAN_imagery_best_in_class.md#Phase-I4/#D1/D3, TODO_imagery_followups.md#Future-workstream
- **relations:** news imagery (I5) consumes it for data-driven stories; RUNBOOK B4 data-source keys.
- **verify-later:** no chart pipeline code expected; go-echarts dependency absence.

### Product illustration pipeline (copyright-safe sketches)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Planned but not built (design work already done — reuse it)" (2026-07-08); I6 not started.
- **what:** Generate stylised product illustrations to avoid copyright/trade-dress exposure from scraped affiliate photos: discovery check `check_product_without_custom_illustration` (per-pass cap ~20), `product-illustration-handler` agent delegating to image-build-handler, `link_asset_to_product` action setting affiliate_products.custom_image_id, renderer precedence custom_image_id → cached_image_url. Stylisation is a hard-coded constraint, not a knob (D7): medium by product category (CAD-like / pencil / watercolour), altered viewpoint, in-context setting, no brand markings; img2img from the cached photo is v2-only under the derivative-work framing.
- **sources:** old/illustration/PLAN_product_illustration.md, PLAN_imagery_best_in_class.md#Phase-I6/#D7, STATUS_imagery_2026-05-12.md#Component-audit-finding
- **relations:** affiliate sites programme (resolver dependency); product components' query.affiliate_products socket; 3D reconstruction explicitly parked.
- **verify-later:** affiliate_products.custom_image_id usage; existence of the handler agent (expected absent).

### Affiliate sites programme and the query.affiliate_products resolver gap
- **category:** NEW:affiliate-commerce
- **status-signal:** aspirational
- **status-evidence:** "This is not the active workstream right now — a holding doc" (2026-05-12); affiliate_products "Zero rows today"; resolver "a wired socket with no plug".
- **what:** The affiliate vision (boxing tickets, darts gear, lead-gen) with three vertical shapes (pure-product / event-ticket / lead-generation) and a layered build path (one product on one page → ingestion + editorial enrichment → imagery via illustrations → event/lead verticals). Substantial scaffolding exists — affiliate_products/affiliate_programs tables, five product components (product-card-with-cta declares `source: query.affiliate_products` with typed image_url; product-specs schema effectively empty), link_registry disclosure flags, the med-* scraper family as an ingestion model — but no program integration, no resolver populating the declared source, no editorial pipeline, no calendar/event infrastructure.
- **sources:** old/STATUS_affiliate_sites_2026-05-12.md, STATUS_imagery_2026-05-12.md#Component-audit-finding, FOCUS_imagery_assessment_1_.md#3.2
- **relations:** product illustration plugs in as a resolver precedence rule; link-management (doc 024); vet-med-pricing med-* agents as pattern.
- **verify-later:** affiliate_products row count; any resolver handling query.affiliate_products in queryresolve/sourceResolver.

### needs_section_data resolution: reconciler, not an agent
- **category:** site-plan-and-reconciler
- **status-signal:** aspirational
- **status-evidence:** "SUPERSEDED 2026-05-06 by FOCUS_directory_builder_and_list_components.md"; "Update 2026-05-27… Decision: a full LLM handler agent (and the directory-builder agent) is not needed for the query-resolvable cases" — the two decided pieces not marked built.
- **what:** `needs_section_data` items are emitted at needs_human_review meaning "couldn't resolve component or required field", not async dispatch; 41 items were stuck system-wide. Resolution machinery already exists (`queryresolve.Resolve`, only `pages_where_type:<type>` implemented; `pages_under_section` named but absent from the dispatch switch). The settled design: (1) implement pages_under_section in queryresolve; (2) a section-data reconciler (a resolver, not an agent) re-attempting open items through existing machinery, closing via closeResolvedDataRequest and flagging re-renders; genuinely-human data (spec-sourced) stays HITL. The originally-planned dedicated handler agent and the never-built `directory-builder` agent are documented dropped ideas.
- **sources:** FUTURE_section_data_handler_1_.md (header supersession + 2026-05-27 update + original)
- **relations:** abandoned: directory-builder agent; relates to list components inventory (~17 components) and page-build-handler.
- **verify-later:** queryresolve.go dispatch switch; count of stuck needs_section_data items.

### Imagery best-in-class programme (G1–G9, D1–D8, phases I0–I8)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** HANDOFF 2026-07-12: "Phase I0 ✅ COMPLETE… Phase I1 ✅ COMPLETE, LIVE-VERIFIED… Phase I2 ⏳ IN PROGRESS… Phases I3–I8 not started."
- **what:** The 2026-07-08 successor programme raising fleet visual quality to best-in-class: nine goals (brand kit/logo permanence, data-accurate infographics, content-linked card imagery, graphic artefacts/sprites, copyright-safe product sketches, news imagery, performance budgets, accessibility/OG surface, quality loop) governed by eight user-confirmed design decisions (D1 code-rendered charts, D2 two lanes, D3 kind batches as text+CHECK, D4 brand guide as data, D5 logo lock, D6 dual icon strategy, D7 sketch constraints, D8 deploy-enforced budgets). Phases I0–I8, each acceptance-gated on robot-hands.com; companion running-notes/runbook/handoff/showcase document set maintained every turn.
- **sources:** PLAN_imagery_best_in_class.md, HANDOFF_imagery_best_in_class.md, RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** builds on the loop-closure programme; RUNBOOK human-gate model; showcase docs quote its numbers.
- **verify-later:** phase status blocks vs live DB/site state; open runbook items B4/B5/B9/B10/B11.

### Two lanes of imagery: plan-driven vs content-driven (Lane B)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** D2 "confirmed" 2026-07-08 as a decision; Lane B storage decision "generic entity_type + entity_id columns on assets — confirmed"; I3 not started.
- **what:** Everything built so far is plan-driven (fixed list decided at plan time). Card images, news charts, and product sketches are content-driven — attached to articles/news items/products, arriving continuously after the plan, prompts composed from the content itself plus the brand guide. Lane B generalises the affiliate custom_image_id pattern via entity_type+entity_id columns on assets, per-entity work item types, and content-sweeping discovery checks, sharing all generation/deploy/audit machinery downstream of the work item.
- **sources:** PLAN_imagery_best_in_class.md#3/#8, RUNNING_NOTES_imagery_best_in_class.md#Turn-2
- **relations:** content-linked card imagery (I3), news imagery (I5), product sketches (I6) are its instances.
- **verify-later:** assets table for entity_type/entity_id columns (expected absent yet).

### imagery_style_guide — per-site brand guide as data (I1)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "✅ PHASE I1 COMPLETE — LIVE-VERIFIED 2026-07-11… per-site imagery_style_guide driving generation with per-kind gating — PROVEN on real output."
- **what:** A site_specs aspect {palette, medium, mood, avoid, reference_asset_keys} distilled from design_intent, read by generate_image for every generation: photographic kinds get medium+mood+palette prepended, icons/sprite sheets palette only, logos nothing; the guide supersedes free-text imagery_direction when present; `avoid` terms feed the negative prompt (stronger channel than positive pleading); reference_asset_keys resolve to stable s3:// URIs (presigned URLs stripped back to bucket/key so anchors outlive the 7-day signature) and flow to Banana as style anchors. The single biggest lever for consistent professional look, per-site so sites diverge deliberately.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-17/#Turn-24, SQL_2026-07-10_robothands_imagery_style_guide.sql, SHOWCASE_technical_architecture.md#4
- **relations:** per-kind gating; reference-image anchoring; supersedes-at-runtime the Phase 0.1 free-text prepend.
- **verify-later:** imagery_style_guide.go; robot-hands site_specs aspect row; +style_guide log lines.

### Logo permanence: generate → human-approve → lock (D5)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Logo: user-approved, LOCKED (assets.locked_at, lock_type=permanent); store-guard refuses overwrites" (2026-07-11, B6 done).
- **what:** One consistent logo for the life of a site is a policy, not a generation feature: the logo is generated, a human approves it via the runbook (A3 eyeball ritual), `locked_at` is set, and the assets upsert's `WHERE assets.locked_at IS NULL` guard refuses any future overwrite; auditors and regeneration paths must skip locked assets. Favicon and OG card are derived from the approved logo, never independently generated. robot-hands' May-8 logo was approved as-is and locked.
- **sources:** PLAN_imagery_best_in_class.md#D5/#Phase-I1, RUNBOOK_imagery_best_in_class.md#B6, RUNNING_NOTES_imagery_best_in_class.md#Turn-24
- **relations:** asset locking 2A supplies the columns; brand-head derivation consumes the locked logo.
- **verify-later:** robot-hands logo asset locked_at/lock_type; store guard in StoreAssetAction.

### Brand-head derived assets (favicon + OG card)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "favicon.png/og-card.png serve 200; og:image + twitter:card injected into every head at render time" (I1 verification, 2026-07-11).
- **what:** `derive_brand_head_assets` action deterministically derives favicon (64×64 square resize) and OG card (1200×630, logo centred on a solid brand-palette colour; gradients rejected) from the locked logo bytes — no LLM — commits both to the site repo and records provenance rows (origin_model='derived-from-logo'). `injectBrandHeadTags` in render_site_components injects favicon/OG/Twitter head tags fleet-wide, idempotently. Runs via a `brand_head` mode branch on asset-deployer dispatched by a `needs_brand_head_assets` work item — the reusable pattern for any site (candidate auto-emit after logo lock).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-25/#Turn-26/#Turn-27, SQL_2026-07-11_asset_deployer_brand_head_mode.sql
- **relations:** logo permanence; sprite CSS head-link reuses the same injection + commit shape.
- **verify-later:** derive_brand_head_assets action registration; asset-deployer check_mode branch; live favicon/og-card on robot-hands.

### Header logo resolution from plan imagery
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "header resolves the locked logo from plan imagery (logo-img live in the served header)" (2026-07-11); fix commit b00c150b.
- **what:** The header is a site component rendered by `render_site_components`, untouched by the page-level resolver fixes, and read the never-populated `sites.logo_url` — so sites showed a text mark despite a deployed logo file. Fix: `loadSiteDataFull` resolves the site-scope logo from site_plan_imagery→assets via `storage.DeployedWebPath` (never assets.url), keeping sites.logo_url as legacy fallback. Closed the long-standing "logo-in-header resolution gap" carried since 2026-05-27.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-23, FOCUS_imagery_assessment_1_.md#5.1 (gap origin), PLAN_imagery_best_in_class.md#Phase-I0
- **relations:** image-role resolver (page-side sibling); DeployedWebPath convention.
- **verify-later:** loadSiteDataFull logo resolution; served header `<img>` on fleet sites.

### Sprite-sheet bullets and list treatment (I2)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "Phase I2 ⏳ IN PROGRESS — the active phase… I2.0 ✅… I2.1 ⏳ REGEN IN FLIGHT (Turn 31)… I2.2 NEXT BUILDABLE NOW" (2026-07-12/13).
- **what:** One coherent N×M glyph grid per site (3×3 @ 768², 256px cells; Banana — harnessing the model's gridded-image tendency), sliced by CSS `background-position`; bullets/nav via `::before` and `.sprite-<name>` classes — one generation, one asset, one stylesheet, no Go image cropping. Delivery deviation resolved twice: sprite CSS ships as a separate committed `/assets/css/sprites.css` + head `<link>` (css_snippets is a GLOBAL library with no site scoping; the per-site committed bundle is the house pattern). Cell-content alignment is THE risk, mitigated by ordered-grid prompt + human eyeball-and-assign gate (B11, cell_names_verified flag); vision auto-verify deferred to I2.4/I8. First generation was near-perfect (all 9 glyphs in reading order); its deploy failure spawned the ExtractActionInputs lesson.
- **sources:** SCOPE_I2_sprite_sheets.md, CONTEXT_PACK_imagery_sprite_sheet.md, PLAN_imagery_best_in_class.md#Phase-I2, RUNNING_NOTES_imagery_best_in_class.md#Turns-28–32, SQL_2026-07-12_seed_robothands_sprite_sheet.sql
- **relations:** five-place kind checklist (I2.0); brand-head commit pattern reused for sprites.css; referenced PLAN_imagery_sprite_sheet.md lives outside this unit.
- **verify-later:** chk_kind includes sprite_sheet; sprite-sheet-main.png 768×768 on robot-hands; sprites.css emit action existence.

### Content-linked card imagery (I3)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Phases I3–I8: not started" (2026-07-12); card-crop decision confirmed 2026-07-08.
- **what:** Every linking card (blog index, news feed, tool directory) carries an image reflecting the content behind it, sharing a visual family with the content page. Confirmed approach: the card image is the article's asset re-cropped per purpose (one generation yields article hero, card crop ~800×450 WebP, OG crop), not a sibling generation. First real Lane B consumer; also clears the one remaining empty image slot on robot-hands (learning-center-index listing card).
- **sources:** PLAN_imagery_best_in_class.md#Phase-I3, RUNNING_NOTES_imagery_best_in_class.md#Turn-2/#Turn-13
- **relations:** two lanes (Lane B); news imagery reuses its mechanics; performance budgets set the card byte ceiling (≤60KB).
- **verify-later:** card kind/purpose in ImagePurposes (expected absent yet).

### News imagery (I5)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** I5 not started; freshness rule confirmed 2026-07-08 ("no SLA… configurable news_image_grace_interval… working suggestion 6h").
- **what:** News ingestion attaches a per-item imagery decision via a small LLM classification: `chart` (data-driven story → I4 pipeline), illustration/photo (I3 pipeline), or none. Feed cards and article pages share the artefact. No SLA (ingest ~2×/day); after a configurable grace interval an item falls back to a brand-kit-derived default image so the feed never shows an empty slot.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I5, RUNNING_NOTES_imagery_best_in_class.md#Decision-log
- **relations:** data-graph pipeline (I4), card imagery (I3), brand kit (I1); news→infographic backlog enhancement.
- **verify-later:** none yet (design only).

### Image performance budgets (I7 / D8)
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** budgets "confirmed as proposed" 2026-07-08 (hero ≤180KB, card ≤60KB, sprite ≤80KB, index above-fold ≤600KB); I7 not started.
- **what:** Per-kind byte and dimension budgets enforced at deploy (extend ImagePurposes with ceilings; WebP for photographic kinds) and policed by a new `image_weight_over_budget` discovery check routed to asset-deployer re-optimisation; responsive srcset + lazy loading in image-bearing templates; alt text required at generation time plus an `image_alt_text_missing` check; sprites amortise small art into one download.
- **sources:** PLAN_imagery_best_in_class.md#Phase-I7/#D8, RUNBOOK_imagery_best_in_class.md#B5
- **relations:** OptimizeImageForWeb/ImagePurposes (existing enforcement point); accessibility goal G8.
- **verify-later:** ImagePurposes byte-ceiling fields (expected absent).

### Image-role alias resolver + authoritative overlay
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "RESOLVED 2026-07-10 — the finding took THREE fixes, all deployed/applied… 16 distinct hero files, each referenced by exactly one page."
- **what:** There was no single contract for how a component gets its image — three incompatible patterns (legacy content_data.hero_url site-wide field; preset `site_assets.background/product_screenshot/...` sources nothing generates; components with no image slot) meant per-page heroes generated correctly but never rendered (same placeholder everywhere / empty src). Fix: `imageryplan.ImageRoleForPath` shared alias table normalising generic image field names to the "hero" role; page-aware `ensureAssets` resolving page hero → site hero fallback; `planSection` injecting the resolved hero under legacy alias keys into resolved_data, which the renderer merges LAST (the designed authoritative overlay), defeating the stale site-wide hero_url. Planner-side key alignment was rejected as structurally impossible (component selected after planning).
- **sources:** FOCUS_imagery_assessment_1_.md#5.1, RUNNING_NOTES_imagery_best_in_class.md#Turns-5–10, PLAN_imagery_best_in_class.md#I0-FINDING
- **relations:** image_source_unsatisfiable check is its guarantee for future domains; DeployedWebPath was fix 2 of the trio; corrupted templates were fix 3.
- **verify-later:** imageryplan package + test; plan_sections_action.go resolve()/ensureAssets/planSection.

### DeployedWebPath committed-path convention (the two-URL serving model)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Verified: zero X-Amz URLs appear in any deployed page's HTML" (2026-07-10 warning box, corrected after a wasted debugging turn).
- **what:** Every generated image has two URLs: `assets.url` is a 7-day presigned S3 URL (SigV4 hard protocol ceiling — a throwaway source handle, never used to render) and the durable committed git path `/assets/images/<asset-key>.<ext>` derived by `storage.DeployedWebPath(asset_key, purpose)` — the single source of truth shared by deployer and resolver so they cannot drift. Pages serve via GitHub Actions → Backblaze B2 → a Cloudflare worker that re-signs each GET server-side. Debugging rule: get the real asset_key/purpose from the DB and curl the derived path; a presigned URL in assets.url is cosmetic staleness, not a broken image.
- **sources:** PLAN_imagery_best_in_class.md#HOW-IMAGE-SERVING-ACTUALLY-WORKS, RUNNING_NOTES_imagery_best_in_class.md#Turn-8, SHOWCASE_technical_architecture.md#3
- **relations:** storage-architecture (worker.js, buckets); image-role resolver emits these paths; leopardess AUDIT_verified_facts D8 is the full write-up.
- **verify-later:** storage.DeployedWebPath/AssetKeyFilename; deploy_image_asset_action.go url-flip (~line 250).

### flag_page_image_rebuild re-render trigger
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** "Delivered 2026-05-27 as the edited plan_sections_action.go and new flag_page_image_rebuild_action.go."
- **what:** Because needs_imagery runs after the page first renders, the fallback bakes into rendered_html and terminal rerenders reassemble without re-resolving. A terminal step in image-build-handler flags the page needs_rebuild and emits `needs_page` at priority 99 for page-scoped imagery, so the page re-resolves *through* plan_sections after its asset lands — closing the asset→render timing loop (the general asset→rerender coupling for site-level components remains an open item).
- **sources:** FOCUS_imagery_assessment_1_.md#5.1-Decision, SHOWCASE_technical_architecture.md#3, TODO_imagery_followups.md#12
- **relations:** image-role resolver (same fix bundle); rerender-reassembles-not-resolves root cause 3.
- **verify-later:** flag_page_image_rebuild_action.go; image-build-handler terminal step.

### image_source_unsatisfiable discovery check
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** Registered 2026-07-09/10; "live but has produced 0 flags (heroes all resolve now) — expected."
- **what:** Flags component input_schema image fields sourced from a `site_assets.<path>` that no asset key, plan imagery row, or image-role alias can supply — the systematic guarantee that the empty-src class of failure is caught on every future domain instead of eyeballed. Flag-only (needs_human_review, no handler), dedup per site/page/function/path, cap 25/pass. Substituted for the structurally-impossible planner-side guard (component chosen after planning). Shares the alias table with the resolver so the two cannot drift.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-7, SQL_2026-07-09_register_image_source_unsatisfiable.sql
- **relations:** image-role resolver; services-hero orphan case (generated hero no component consumes).
- **verify-later:** check_image_source_unsatisfiable.go; design-discovery-agent run_checks.

### Corrupted component templates and the quality→regeneration bridge
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "✅ Bridge self-heal proven END-TO-END… no human involvement at any step. Fleet-remaining corrupted: 7" (2026-07-10); "10/14 healed" in the handoff.
- **what:** 14 components fleet-wide had html_template saved as RENDERED OUTPUT (literal `<no value>`, zero `{{…}}` vars) — historical damage from the pre-validation component-generation era (created_from='generated', 2026-03-31→04-13); the modern writer's pre-store validation already rejects this class. Detection existed (compute_component_quality flags "0 template variables") and repair existed (needs_component_regeneration → component-creator); the missing piece was a ~200-line bridge: `check_component_template_corrupted` discovery check (cross-site guard since components are fleet-shared, cap 5/pass). Field-preservation guard rejections are handled by re-queuing with exact field names in spec.description (rendered into the creator prompt).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turns-11–16/#Turn-20, SQL_2026-07-10_register_component_template_corrupted.sql, SHOWCASE_technical_architecture.md#5
- **relations:** tool-library/component-creator contract; the flagship "self-healing fleet" showcase story.
- **verify-later:** check_component_template_corrupted.go; remaining corrupted count fleet-wide.

### mark_item_failed error honesty (flag-before-complete)
- **category:** improvement-loop
- **status-signal:** deployed
- **status-evidence:** "Fix applied (SQL_2026-07-10_pagebuild_mark_item_failed.sql… verified)… Failed page builds are now VISIBLE instead of silently complete."
- **what:** page-build-handler's step-level error routing pointed at `complete_error`, a SUCCESS-labelled complete_workflow — so a real step failure completed the orchestration and the dispatcher stamped the item 'complete' with no error (the "no-op complete" anomaly, triggered by a Kafka reply flake). The established flag-the-item-BEFORE-completing pattern was extended to real errors: a `mark_item_failed` step (update_work_item_status → 'failed', attempt-counted) inserted ahead of complete_error with all 8 error pointers repointed. Workflow-config-only. A fleet-trust principle: "a fleet you can trust starts with a fleet that tells the truth about itself."
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-16, SQL_2026-07-10_pagebuild_mark_item_failed.sql, SHOWCASE_imagery_workstream.md#4
- **relations:** Kafka partition race (the trigger); CompleteWorkItemAction guard semantics; likely needed on other handler workflows.
- **verify-later:** page-build-handler workflow error_step pointers; failed-vs-complete item stats post-fix.

### Work-item re-drive and zombie-claim operational semantics
- **category:** debugging
- **status-signal:** partial
- **status-evidence:** Standing lessons recorded Turns 31–32 (2026-07-12/13); "the zombie-claim dispatch stall was the single biggest time cost of the 2026-07-09/10 verification" (B9, still open).
- **what:** Hard-won dispatch mechanics: a claimed item stuck >~10 min blocks its ENTIRE site via find_dispatchable_site's NOT-EXISTS clause (standing unstick UPDATE; real fix = reaper cadence + per-item-type circuit breaker, TODO 6/10/11); re-driving an item requires resetting `attempt_count=0` and claim metadata, not just status (capped items are silently excluded — dispatch looks dead but is correctly idle); a just-finished orchestration's tail can re-stamp a freshly-reset item complete (state-machine race); manually-inserted items are NOT auto-triaged (insert as triaged); dedup is a partial unique (site_id, item_key) over non-terminal statuses whose exact semantics made resets awkward. Historical: dispatch once didn't claim triaged imagery items behind page work; fairness/observability gaps (outer ORDER BY, trigger not writing orchestration_states) remain listed.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-32, HANDOFF_imagery_best_in_class.md#Mechanisms, TODO_imagery_followups.md#6/#8/#9/#10, RUNBOOK_imagery_best_in_class.md#B9
- **relations:** mark_item_failed; state-machine corruption on failed items (claim metadata not cleared); scheduler-and-tasks.
- **verify-later:** find_dispatchable_site SQL; reaper cadence; idx_swi_dedup definition.

### Pipeline field as soft routing label
- **category:** debugging
- **status-signal:** deployed
- **status-evidence:** "check_unfulfilled_imagery_plan.go hardcodes Pipeline: 'build' — the 2026-05-17 fix is in the code" (verified 2026-07-08); Part B dispatcher-filter loosening scoped alongside.
- **what:** Discovery checks running under design-discovery-agent inherited pipeline='design', which build-dispatch-loop's item_pipeline filter silently excluded — needs_imagery items required manual UPDATEs to dispatch. Two-part fix: checks write Pipeline:"build" at source (pipeline is the destination handler's side, not the origin's), and the dispatcher's filter was removed so any future mismatched emission still dispatches. The field survives as a soft routing label for possible future multi-pipeline dispatchers.
- **sources:** TODO_imagery_followups.md#7, RUNNING_NOTES_imagery_best_in_class.md#Turn-2 (verification)
- **relations:** work-item dispatch semantics; design-discovery-agent context.
- **verify-later:** build-dispatch-loop load_items config; Pipeline literal in imagery checks.

### ExtractActionInputs Strategy-0 explicit dot-paths lesson
- **category:** development-guide
- **status-signal:** partial
- **status-evidence:** "FIXED workflow-only: SQL_2026-07-12_asset_deployer_explicit_paths.sql… Standing lesson: give ExtractActionInputs actions explicit dot-paths; never trust the search" — with the dispatch-shape (`input_data.spec.*`) gap recorded, not fixed.
- **what:** ExtractActionInputs' aggressive recursive field search matched a stale `purpose` elsewhere in collected_data, so the sprite sheet deployed as a 900×900 hero-config JPG despite the child receiving purpose='sprite_sheet' — explicit Strategy-0 dot-path config values are resolved first and win. Latent siblings: items dispatched via build-dispatch-loop carry payload under input_data.spec.* which the explicit paths miss; historical spawned deploys may have silently used hero dimensions (May icons' file dims worth checking).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-31, SQL_2026-07-12_asset_deployer_explicit_paths.sql, HANDOFF_imagery_best_in_class.md#I2.1
- **relations:** dispatch input contract; deploy_image_asset defaults ("purpose":"hero").
- **verify-later:** asset-deployer deploy_asset step config paths; datahelpers extraction strategies.

### No runtime re-compose path — layout change via the 025 FK-swap pattern
- **category:** design-composition
- **status-signal:** partial
- **status-evidence:** "B7 COMPLETED 2026-07-10 evening — via the 025 FK-swap pattern… there is no runtime re-compose path (deliberate deferral). NEW OPEN ITEM: build a proper runtime re-compose mode."
- **what:** Changing an existing site's layout is deliberately unsupported at runtime: install_site_composition refuses when a style_collection exists, and fork_theme_from_site's install mode was removed 2026-04-19. The sanctioned workaround is a targeted `css_themes.layout_id` FK swap (backup + verify) followed by a webdesign-agent CSS re-render + page rerenders. Root cause of the B7 brochure fallback: robot-hands' old-format classification lacked `industry_tags`, so the scheme-aware matcher had nothing to score — while the layout library already held the right answer (tool-portal-dark, itself grown from a prior instance of the same gap: the library learns).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turns-18–22, SQL_2026-07-10_b7_layout_fix.sql, SQL_2026-07-10_b7_layout_swap.sql, PLAN_imagery_best_in_class.md#B7
- **relations:** design-composition doc 027 matcher; needs_new_layout_candidate → library-growth loop; classification format drift (also caused the missing news flag).
- **verify-later:** install_site_composition refusal; robot-hands css_themes.layout_id = tool-portal-dark.

### robot-hands.com rebuild (testbed case study)
- **category:** site-case-studies
- **status-signal:** deployed
- **status-evidence:** "Phase I0 ✅ COMPLETE. 33-page rebuild w/ live news (9 sources); 16 distinct per-page git-path heroes, zero expiring URLs" (2026-07-12).
- **what:** The adopted site's content layer was broken (10 zero-component pages, NULL content, features content in a pre-drift schema) while the imagery pipeline was correct — so it was rebuilt from scratch with news scope: supersede adoption-residue aspects, news-enable classification, add a mission_brief aspect, retire stale items, insert+manually-triage a needs_site_plan trigger, then a fresh 33-page plan (29 imagery rows) built unattended through dispatch. A 2026-05-20 audit first said "PATCH, do not re-plan" (foundation sound, build broken); the 2026-07-08 decision superseded it with full re-plan. Hard requirements: tools must actually work (deployed JS, resolving links) and it is the acceptance surface for all I-phases.
- **sources:** HANDOFF_robot_hands_rebuild.md, SQL_2026-07-08_robothands_rebuild_prep.sql, SQL_2026-07-08_robothands_mission_brief.sql, RUNNING_NOTES_imagery_best_in_class.md#Turns-2–4
- **relations:** news enrichment pattern; schema-contract drift bug (features {title,description} vs {icon,name,description}); orphan pre-rebuild pages cleanup still open.
- **verify-later:** current plan 7a40a0f9; page/component counts; tool pages' build_status.

### News pipeline replication and the news enrichment pattern
- **category:** news-feed-pipeline
- **status-signal:** deployed
- **status-evidence:** "News pipeline: LIVE and healthy (replicate gas)" (2026-05-20 recon); robot-hands "9 content_sources seeded, 0 erroring" (2026-07-10).
- **what:** The live chain content-feed-trigger → content-feed-orchestrator → feed-ingester → feed-triage → render_news_section (→ /data/latest-news.json + news-archive.json), with content_sources rows of four parallel types (rss, news_search, api_news with grok/web-search tools, scrape) as the replication template — pure data rows, no new code. Adding news to an existing site is enrichment, not re-plan: evaluate_news_feed writes classification.content_features.news_feed, news-section-addition amends the plan (RULE 11 places latest-news on the homepage). Two distinct components serve it (latest-news card grid on index; news-listing full page). Item expiry happens via status transition; the expires_at column exists but is unpopulated.
- **sources:** HANDOFF_robot_hands_rebuild.md#PIPELINE-RECON, old/README_news_pipeline.md, PLAN_imagery_best_in_class.md#Phase-I0-status
- **relations:** deploy_page files_field dependency (news JS silently dropped otherwise); news imagery (I5) builds on it.
- **verify-later:** robot-hands content_sources rows; content_feed_items lifecycle counts.

### Price-news TTL and news→infographic enhancements
- **category:** news-feed-pipeline
- **status-signal:** aspirational
- **status-evidence:** "Both (a) and (b) are NICE-TO-HAVE… backlog" (2026-05-20, user-stated not urgent).
- **what:** (a) Price-aware filtering with short expiry: classify fetched news for price-movement items and expire them after 1–2 days via the existing-but-unused expires_at column plus a topics-based triage tag; per-site vertical. (b) News→infographic: pick 1–2 items, research the subject, generate an infographic — ties into the imagery infographic kind, research adapters, and the data-graph pipeline when data-driven.
- **sources:** HANDOFF_robot_hands_rebuild.md#NEWS-ENHANCEMENTS, old/README_news_pipeline.md
- **relations:** data-graph pipeline; news imagery I5 partially absorbs (b).
- **verify-later:** expires_at population; any price-tagging triage rule.

### deploy_page files_field contract (co-located JS must ship)
- **category:** tool-pipeline
- **status-signal:** deployed
- **status-evidence:** "CRITICAL deploy dependency: page-rerender deploy_page step MUST use files_field:'rendered_page.files' (NOT content_field)… fix was applied during the gas rollout 2026-05-19/20 — VERIFY it's still in the current config."
- **what:** If page deploys use content_field (HTML only), component JS (/tools/assets/*.js) is silently dropped — news sections render empty and interactive tools ship as shells. The files_field form carries the full file set. Related evidence: tool generation works but deploy is where tools stalled (gas-unit-converter built with real JS but stuck build_status='pending'); the working-tools acceptance is deployed page + committed JS + resolving links, never "component generated".
- **sources:** HANDOFF_robot_hands_rebuild.md#Tools/#News-pipeline, TODO_imagery_followups.md#17
- **relations:** robot-hands rebuild hard requirement; render_css_from_spec fallback gap (page-level CSS silently not shipped) noted alongside.
- **verify-later:** page-rerender deploy_page config; tool page build_status across sites.

### Reference-image style anchoring
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** Item 24 decision 2026-05-18 ("plumb reference-image AND switch icon model"); Banana native path live via style guide reference_asset_keys (I1, 2026-07-11); IP-Adapter/img2img paths not built.
- **what:** Conditioning generations on a reference image for sibling consistency. Three techniques ranked (img2img subject-anchor; IP-Adapter style-anchor, not on Stability's standard REST endpoint; LoRA highest); three reference-provenance options (generate-one-then-derive; per-site curated style library; system-wide per-kind house style). What shipped is the Banana-native form: approved reference assets resolved to stable s3:// URIs flow as style anchors for photographic kinds. Schema hooks (reference_image_uri field, origin_asset_id, alterations JSONB) exist ahead of the fuller paths.
- **sources:** TODO_imagery_followups.md#24, PLAN_imagery_best_in_class.md#D4, RUNNING_NOTES_imagery_best_in_class.md#Turn-17
- **relations:** imagery_style_guide; adoption mirror as a reference source; product-sketch img2img v2.
- **verify-later:** Banana provider reference handling; whether any non-Banana reference path exists.

### Per-vertical LoRA fine-tunes
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Status: planned, not started… the fine-tuning plan presupposes an adapter that can take a model field; our adapter cannot" (assessment); best-in-class explicitly keeps it "as a future consistency upgrade once I1's reference-image approach shows its limits".
- **what:** Training per-vertical image LoRAs (60–90 curated images, SDXL/PixArt base, ~£35–95 first pass, per 018_canine_biology) for consistent visual style (vet diagrams, energy infographics). Blocked historically on model-selection plumbing; now deliberately deprioritised behind the cheaper reference-image approach.
- **sources:** FOCUS_imagery_assessment_1_.md#8, PLAN_imagery_loop_closure.md#Open-items, PLAN_imagery_best_in_class.md#6
- **relations:** canine-biology (018); model-infrastructure training; reference-image anchoring as the substitute.
- **verify-later:** none (not started).

### Prompt-composition composer/envelope revisit
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** "Step 5 — image prompt cascade: Defer… see FOCUS_prompt_composition_pattern.md for why copying the text pattern is the wrong target" (resolved 2026-05-13).
- **what:** The image path keeps a single-prepend cascade rather than matching page-content-writer's richer text composition — a considered asymmetry: the text pattern itself is judged fragile and not worth copying. The strongest future candidate is a composer step producing a parameter envelope for both text and images, likely landing in a 2H-sibling phase. Partially realised since by the style-guide gating (which is composition-by-kind, not a full cascade).
- **sources:** PLAN_imagery_loop_closure.md#Decisions/#Image-prompt-cascade—deferred/#Open-items
- **relations:** FOCUS_prompt_composition_pattern.md (outside unit); imagery_style_guide; 2H request shape.
- **verify-later:** whether any composer/envelope step exists.

### Components declare imagery contracts / many-images-per-page direction
- **category:** imagery
- **status-signal:** aspirational
- **status-evidence:** Phase 3.5 "NEW for the refresh" but Phase 3 deferred; "Future direction — pages with many images (post-Phase-3)".
- **what:** content_components.input_schema v2 already supports typed image fields with arbitrary `source` resolvers, but only hero_image uses it. The direction: components own their imagery contracts (team-grid declares member_avatars arrays; services-grid declares per-service icons), the renderer resolves scoped site_plan_imagery rows by asset_key, discovery walks the declared gaps, and generation scales horizontally (a 30-image page is 30 work items through the unchanged chain). Enables per-image audit and retires silent fallthroughs to /assets/images/hero.jpg. The features `icon` string contract being one-sided (no renderer wiring) is the standing counter-example, resolved separately via Lucide.
- **sources:** PLAN_imagery_loop_closure.md#3.5/#Future-direction, FOCUS_imagery_assessment_1_.md#4.2/#9-item-5
- **relations:** Lane B; card imagery; contracts-and-standards (input_schema slot specs).
- **verify-later:** any component beyond hero_image with resolved image declarations.

### Context-bundle seeding for fresh agent threads
- **category:** documentation-system
- **status-signal:** deployed
- **status-evidence:** RUNBOOK A4 standing ritual: "z_bundles/imagery_seed_docs/imagery_bundle.sh… Output lands at z_bundles/imagery_bundle.md"; used to seed Turn 1.
- **what:** A repeatable script assembles the workstream's context bundle (key docs + live schema/runtime sections queried from the cluster) for cold-starting a fresh agent session; run after credential refresh or the DB sections come out empty. Paired with the document-set discipline: PLAN (map) / RUNNING_NOTES (turn-by-turn evidence) / RUNBOOK (human task queue) / HANDOFF (single cold-start entry point, updated every turn) / SHOWCASE (shareable summaries).
- **sources:** RUNBOOK_imagery_best_in_class.md#A4, HANDOFF_imagery_best_in_class.md#Document-map, CONTEXT_PACK_imagery_sprite_sheet.md
- **relations:** documentation-system conventions (travelling docs); the CONTEXT_PACK is the sprite-specific instance.
- **verify-later:** z_bundles/imagery_seed_docs/imagery_bundle.sh existence.

### API keys logged in plaintext (scrub + rotate)
- **category:** database-and-infrastructure
- **status-signal:** deployed
- **status-evidence:** "SECURITY — STILL OPEN, STILL HIGHEST PRIORITY" (2026-05-20) → "✅ DONE 2026-07-08. Keys rotated (user confirmed)."
- **what:** STABILITY_API_KEY and BANANA_API_KEY were logged in plaintext at info level (adapter init env-dump zap.Any, zap.String("apiKey"), plus B2 keys in NewS3Client debug logs) — with billing exposure on the paid Banana tier. Carried as highest-priority for seven weeks across handoffs; resolved by scrubbing to scoped fields and rotating both keys.
- **sources:** TODO_imagery_followups.md#SECURITY, HANDOFF_robot_hands_rebuild.md#Carried-forward, RUNBOOK_imagery_best_in_class.md#B1
- **relations:** provider abstraction (the code carrying the logging); credentials handling conventions.
- **verify-later:** dynamic_adapter.go logging fields; no raw keys in current logs.

### Manual agent-trigger pattern (kcat orchestrate; never hand-roll spawn+call)
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "the documented system.intake pattern is STALE (topic doesn't exist) — the working mechanism is the kcat trigger script pattern" (Turn 18); "Do NOT hand-roll spawn_agent+call_agent inline workflows" (Turn 26 lesson).
- **what:** Manually triggering agents means an `action=orchestrate` envelope to `system.agent.generic.requests` with config.agent_type + input_data — known-good for improvement-loop, webdesign-agent, rerender-pages. Hand-crafted inline spawn+call parents fail because the spawned child runs its workflow on INIT and idles before the call arrives; work destined for spawned handlers must route through work items + dispatch instead.
- **sources:** HANDOFF_imagery_best_in_class.md#Mechanisms, RUNNING_NOTES_imagery_best_in_class.md#Turn-18/#Turn-26
- **relations:** dispatch input contract; brand-head activation was the proving case.
- **verify-later:** 033_rerender_pages_trigger.sh precedent; system.intake topic absence.

### Dispatch input contract for handler agents
- **category:** contracts-and-standards
- **status-signal:** deployed
- **status-evidence:** Recorded as a hard-won mechanism 2026-07-12: handlers invoked by build-dispatch-loop receive `input_data.spec`, `input_data.site_id`, `input_data.domain`, `input_data.item_type`.
- **what:** The canonical payload shape a dispatched handler sees; step conditions and input paths must be written against it (e.g. asset-deployer's check_mode tests both `input_data.spec.mode` and `input_data.mode` to cover dispatch and direct-call shapes). Divergence between dispatch-shape (`input_data.spec.*`) and direct-call shape (`input_data.*`) is a live source of latent extraction bugs.
- **sources:** HANDOFF_imagery_best_in_class.md#Mechanisms, SQL_2026-07-11_asset_deployer_brand_head_mode.sql, SQL_2026-07-12_asset_deployer_explicit_paths.sql (NOTE block)
- **relations:** ExtractActionInputs lesson; work-item dispatch semantics.
- **verify-later:** build-dispatch-loop spawn payload construction.

### psql read-only PreToolUse gate
- **category:** development-guide
- **status-signal:** deployed
- **status-evidence:** "added a PreToolUse permission hook (.claude/hooks/psql_readonly_gate.py)… tested against a 20-case matrix and proven live" (Turn 3, 2026-07-08).
- **what:** Agent-session tooling: a hook auto-approves read-only SELECT/`\d` psql via the exact kubectl-exec form while mutations still prompt the human — reducing friction for the DB ground-truth checks every session performs. Session auth expires ~daily (runbook A1 re-login ritual).
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-3, HANDOFF_imagery_best_in_class.md#Mechanisms, RUNBOOK_imagery_best_in_class.md#A1
- **relations:** context-bundle seeding; operator runbook rituals.
- **verify-later:** .claude/hooks/psql_readonly_gate.py; settings.local.json hook wiring.

### Human taste-gate operating model (runbook rituals)
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** RUNBOOK structure in active use: standing rituals A1–A5 + one-off queue B1–B11, most B-items closed with dates; "Humans only at the taste layer" (showcase).
- **what:** The imagery workstream's division of labour: agents do all authoring/migrations/deploy-prep; humans do credentials, backups approval, budget sign-off, and visual approval gates — logo approval (once, then locked), sprite-sheet cell verification (assign true meanings after generation), and sampled page eyeballs per phase acceptance. Gates are deliberately the phases' biggest cost; generation is never trusted to self-judge taste.
- **sources:** RUNBOOK_imagery_best_in_class.md, SHOWCASE_imagery_workstream.md#Why-it's-interesting, SCOPE_I2_sprite_sheets.md#Phasing
- **relations:** logo permanence; sprite eyeball gate B11; hitl category (broader HITL machinery).
- **verify-later:** n/a (process concept); runbook item states.

### Imagery work-item economy end-to-end chain
- **category:** imagery
- **status-signal:** deployed
- **status-evidence:** SHOWCASE technical architecture (verified against production 2026-07-10): full diagram planner → site_plan_imagery → needs_imagery → image-build-handler → provider adapter → store_asset → asset-deployer → flag_page_image_rebuild → resolver → rendered page; "~90 s prompt → git commit".
- **what:** The consolidated, operating imagery pipeline as a single nameable chain, including its dedup property (partial unique index lets checks re-run forever) and honest-state property (mark_item_failed). This is the umbrella concept the individual phase concepts compose into, and the shape any new imagery capability (sprites, cards, sketches) must ride.
- **sources:** SHOWCASE_technical_architecture.md#2/#3, SHOWCASE_one_pager.md, PLAN_imagery_best_in_class.md#2
- **relations:** every imagery concept above; development-guide work-item lifecycle.
- **verify-later:** end-to-end trace on a fresh needs_imagery item.

### Rerender reassembles, it does not re-resolve
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** FOCUS §5.1 root cause 3 (2026-05-27); Turn 10 "NEW, SEPARATE ISSUE… rerender-completeness gap" — later narrowed: the specific 2026-07-10 instance was corrupted templates, but the underlying property stands ("needs_rerender and the colour/CSS fixers regex-patch stored rendered_html; they do not re-run plan_sections").
- **what:** The terminal rerender path reassembles existing rendered_html rather than re-running section resolution, so values that later land in content_data (resolved images, alt text) do not reach the HTML without a page rebuild through plan_sections. flag_page_image_rebuild routes page-scoped imagery around this; site-level components (header/footer) and non-hero inline sections remain exposed. Also noted: page_components.rendered_html is a snapshot, not a view — template changes don't reach deployed pages without a rebuild.
- **sources:** FOCUS_imagery_assessment_1_.md#5.1, RUNNING_NOTES_imagery_best_in_class.md#Turn-10/#Turn-11, HANDOFF_robot_hands_rebuild.md#Also-watch
- **relations:** flag_page_image_rebuild; corrupted templates (the misdiagnosis neighbour); styling-render-pipeline.
- **verify-later:** which paths re-run plan_sections vs regex-patch rendered_html.

### Orphaned generated assets (component consumes nothing)
- **category:** imagery
- **status-signal:** partial
- **status-evidence:** "services-hero template has no `<img>`; the generated hero_services.jpg is never consumed" (2026-07-09); full orphan check "deferred"; services coverage question open since item 21 (2026-05-15).
- **what:** Generation waste where the plan requests imagery no selected component can display (services hero), or assets outlive replans (stale sprite-sheet-main.jpg clutter; superseded May assets). Detection partially covered by undeployed_assets and image_source_unsatisfiable; a dedicated orphan/asset-supersession check and repo cleanup pass remain open, as does the planner-side question of which pages get heroes at all.
- **sources:** RUNNING_NOTES_imagery_best_in_class.md#Turn-5/#Turn-31, TODO_imagery_followups.md#21, PLAN_imagery_best_in_class.md#I0-finding-(c)
- **relations:** image_source_unsatisfiable (the flag side); planner key stability (another orphan source).
- **verify-later:** assets rows with no referencing rendered_html; site-repo files with no assets row.

## Dropped / superseded ideas ledger (quick index)
- `check_legacy_image_prompts_aspect` — designed, then retired as scoped (see concept above).
- `directory-builder` agent and a dedicated `needs_section_data` handler agent — decided in project docs, never built, superseded by queryresolve + reconciler design.
- Per-section imagery "deferred" (loop-closure v2 decision) — reversed 2026-05-12 by Phase 2G.
- Diffusion transparency for icons — abandoned (flat-grey chip).
- SDXL for flat icons — abandoned (Banana routing).
- Fix-A prompt-only icon rescue — tried, judged insufficient 2026-05-18.
- Planner-side guard "don't plan heroes for hero-less components" — structurally impossible (component chosen post-plan); substituted by image_source_unsatisfiable.
- Hand-rolled spawn+call trigger workflows — anti-pattern, replaced by work-item dispatch routing.
- `system.intake` trigger topic — stale/dead; kcat orchestrate envelope is the working path.
- pageflow-builder — retired from new-feature work, snapshot preserved.
- 3D product reconstruction — parked (sketches cover the need).
- fork_theme_from_site install mode — removed 2026-04-19; FK-swap is the sanctioned workaround.
