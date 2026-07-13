# PLAN — Imagery Audit and Fix Loop

Sequenced work plan for closing the gap between what the planner / spec / other agents ask for in imagery, and what is actually delivered. Builds on `FOCUS_imagery_assessment.md`.

Sister docs (read alongside):

- `PLAN_imagery_phase_2g.md` — detailed plan for Phase 2G (current work; structured imagery rows in the plan domain).
- `PLAN_product_illustration.md` — sibling: product imagery via `affiliate_products` resolver. Not part of this loop, but interacts at the renderer level.
- `STATUS_imagery_2026-05-12.md` — current operational status with verification trail.
- `FOCUS_prompt_composition_pattern.md` — considered opinion on the text-vs-image cascade asymmetry surfaced during 2G step 5 design. Captures why we deferred matching text's pattern and what better composition might look like.

---

## Decisions taken

| Question | Decision |
|---|---|
| Imagery audit — extend existing or new agent? | New agent: **`imagery-quality-auditor`**, sibling of `visual-design-auditor` under `design-audit-agent`. |
| Max regeneration attempts per finding | **2** (matches the existing `max_fix_attempts: 2` on structured findings). |
| Asset locking | **Mirror `page_components` exactly**: `locked_at timestamptz` + `locked_by text` on `assets`, same query exclusion patterns. Hard-vs-soft via `locked_by` (per 013 / 031). Timed expiry deferred. |
| Per-section / per-component image granularity | **Un-deferred (2026-05-12).** The original plan deferred this; Phase 2G now delivers it via `site_plan_imagery` with `scope ∈ {site, page, section}`. Section scope is wired but populated sparingly in v1. |
| Structured imagery — site_specs or site_plan? | **`site_plan` domain.** Sibling table `site_plan_imagery`, not a `site_specs` aspect. Resolved 2026-05-12; see `PLAN_imagery_phase_2g.md` § "Decisions taken". |
| Migration off legacy `image_prompts` | **Operational deregistration, not a dedicated check.** Reframed 2026-05-13: "is a site on legacy?" is not a useful signal — legacy isn't inherently a fault. What matters is whether something is broken, and the existing checks already detect their respective brokenness against each path. Both checks run in parallel during transition; once `unfulfilled_image_prompt` reliably finds zero gaps across active sites, pull it from `design-discovery-agent.run_checks` — one string out of a JSON array, no code change. |
| Planner — single LLM call or two? | **Single.** The imagery block is a new key in the existing plan-builder JSON output. No separate "imagery planner" call. |
| Step 4 — `asset_key` derivation from imagery row | **Option (a): `asset_key = imagery.key` directly.** Predictable, matches legacy `hero_home → asset_key=hero_home`. Collision detection logs warnings rather than blocking; the planner prompt steers toward unique keys. Resolved 2026-05-13. |
| Step 4 — new `item_type` or reuse `unfulfilled_hero_variant`? | **New `needs_imagery`.** The variant item_type assumes "hero" semantics in its name, wrong for logo/illustration/icon/infographic. Cleaner semantic boundary for future kind-specific handling. Resolved 2026-05-13. |
| Step 4 — per-pass cap | **20 items/pass.** Sites with rich imagery complete over multiple discovery passes. Higher priority (site logo, index hero) emitted first via SQL ORDER BY. Resolved 2026-05-13. |
| Step 4 — dedup key shape | **`needs_imagery:<scope>:<scope_ref\|->:<key>`.** Deterministically derivable from any imagery row's columns; dash placeholder keeps the format positional for site-scope rows. Resolved 2026-05-13. |
| Step 5 — extend variant branch matcher or new branch alongside? | **New branch alongside.** Variant chain's hardcoded `purpose="hero"` and brand-update behaviour assumes hero-only; cleaner separation now than forking later when kind-specific handling lands. Six new workflow steps, shared asset-deployer tail. Resolved 2026-05-13. |
| Step 5 — `update_site_brand_assets` rule for new branch | **Option (b): site-scope OR (page-scope AND `scope_ref=index` AND `kind=hero`).** Preserves today's behaviour where `hero_home` updates brand_assets. Boolean computed by step 4's discovery check and carried as `spec.brand_update`; workflow routes on it. Resolved 2026-05-13. |
| Step 5 — image prompt cascade | **Defer.** Keep the single-prepend `imagery_direction` cascade for v1. The richer composition (matching text's pattern) is a separate phase; see `FOCUS_prompt_composition_pattern.md` for considered opinion on why copying the text pattern is the wrong target. Resolved 2026-05-13. |
| Plan-version handling on imagery deltas | **Discovery check is sufficient for v1; reconciler extension is a small follow-up.** The check reads only current-plan imagery rows, so stale plans are naturally ignored. Reconciler-driven emission removes the lag between plan write and discovery firing — worth doing but not blocking. Resolved 2026-05-13. |
| 2H — request shape v1 field set | **negative_prompt, seed, reference_image_uri, cfg_scale, steps.** `style_preset`, `samples`, `safety_mode` deferred. Style_preset specifically has no adapter mapping code today; adding it is its own decision (what does each kind map to in Stability's preset vocabulary?). Resolved 2026-05-13. |
| 2H — defaults location | **Go-side per-kind map, not a config table.** Simpler, no migration. Move to a table if per-vertical or per-site overrides matter. Spec values from caller override defaults. Resolved 2026-05-13. |
| 2G step 5 — kind=logo handling in new branch | **Deferred.** Hotfix hardcodes `purpose: "hero"` in both store steps because `store_asset.purpose_field` isn't a supported config key. kind=logo items route through the legacy `unfulfilled_image_prompt → needs_logo` path for now. Proper fix is a small Go change to `store_asset` (add `purpose_field` lookup), or workflow-side kind routing. Resolved 2026-05-14. |
| Step 5 `input_mapping` discipline | **Use `?` suffix for optional spec fields.** `constraints`, `style_hints`, and `kind` originally listed as required in `call_imagery_gen.input_mapping`, but step 4 only emits these when source `site_plan_imagery` columns are non-null. Items with null source columns failed at extraction. Convention going forward: any spec field that may be absent on legitimate items gets `?`. Verified by orchestration `e98deca7`. Resolved 2026-05-14. |

---

## Progress (updated 2026-05-14)

Phase 2G + Phase 2H are operationally verified end-to-end on robot-hands.com. First asset through the new path: `a12b5d71-3069-45b9-b170-1da476afbcd3` (asset_key `hero_home`, purpose `hero`), generated 2026-05-14 10:22 via orchestration `b53b4515-7c67-4dce-bd82-339370d16f49`. Deployed file visible at `robot-hands.com/assets/images/hero-home.jpg`. Phases 3, 4, 5, 6 of the outer plan not started.

### Application status (2026-05-14)

| Artefact | Type | Applied | Notes |
|---|---|---|---|
| `phase_2g_step3_planner_imagery_prompt.sql` | SQL | ✅ | Plus max_tokens bump and corrective patch — see step 3 narrative below. |
| `phase_2g_step3_planner_maxtokens_bump.sql` | SQL | ✅ | Raised plan_site `max_tokens` from 4000 to 8000 after JSON truncation seen on 14-page roadmap. |
| `flattenImageryBlock_patch.go` | Go | ✅ | Confirmed live by `site_plan_imagery` rows landing after planner trigger. |
| `check_unfulfilled_imagery_plan.go` | Go | ✅ | Confirmed live by `unfulfilled_imagery_plan` emitting 8 work items on first discovery run. |
| `phase_2g_step5_image_build_handler_needs_imagery.sql` | SQL | ✅ | Confirmed live — orchestration `b53b4515` routed through `check_item_type_imagery → spawn_image_gen_imagery → call_imagery_gen → check_imagery_brand_update → store_imagery_brand_asset → spawn_asset_deployer → call_asset_deployer → complete`. |
| `phase_2g_step5_hotfix_optional_input_mapping.sql` | SQL | ✅ | Added 2026-05-14 — required step-5 `input_mapping` fields `constraints`, `style_hints`, `kind` made optional via `?` suffix. Caught by orchestration `e98deca7` failing at `call_imagery_gen` with "source path 'input_data.spec.constraints' not found". |
| `phase_2g_step5_hotfix_store_asset_purpose.sql` | SQL | ✅ | Added 2026-05-14 — `purpose_field` is NOT supported by `store_asset` action; replaced with literal `purpose: "hero"` in both store steps. Caught by orchestration `076b6f19` failing at `call_asset_deployer` with "source path 'asset_stored.image_uri' not found" because the failed config caused store_asset to write a partial row without populating its output_field. Side effect: kind=logo items can't currently use the new branch — see Open items. |
| `phase_2g_step4_register_check.sql` | SQL | ✅ | `unfulfilled_imagery_plan` registered in `design-discovery-agent.run_checks`. |
| `generate_image_actions.go` (Phase 2H) | Go | ✅ | `kindDefaults` map, `resolveKind`, `parseAspectRatio`, plus extended imageData in `GenerateImageAction`. Chassis pod confirmed running new binary. |
| `dynamic_adapter.go` (Phase 2H) | Go | ❓ unconfirmed | Whether the adapter binary specifically carries the 2H changes is unclear — the `generateImage: stability request` log line wasn't found on `kubectl logs -l app=image-generator-adapter` post-rollout. May be deployed under a different label, OR the adapter wasn't rebuilt alongside the chassis. Asset generation succeeded regardless, but per-kind cfg_scale/steps and the negative_prompt application at Stability may still be inactive. Worth verifying which deployment carries which binary. |
| `phase_2h_step4_legacy_kind_defaults.sql` | SQL | ⏳ pending | Adds `default_kind` to legacy `call_logo_gen` / `call_hero_gen` / `call_variant_gen` step configs so legacy callers also pick up kind-derived defaults. Optional — only matters once the kind-routing question is solved for kind=logo items. |
| `site_plans_unique_current_index.sql` | SQL | ❓ unconfirmed | Adds revert function + history view (partial index was already present). |

To confirm operational state at any time, the three-line UNION query continues to be useful — see status doc.

### End-to-end verification trace (2026-05-14)

Reference for future debugging — this is the full happy path observed:

| Time | Event |
|---|---|
| 10:22:06 | Parent orchestration `b53b4515` created — image-build-handler kicked off for work item `19a4b7cd` (kind=hero, asset_key=hero_home, brand_update=true). |
| 10:22:06 | `ensure_site_record` → loaded site context. |
| 10:22:06 | `check_item_type_imagery` → matched `needs_imagery`, routed to `spawn_image_gen_imagery`. |
| 10:22:06 | `spawn_image_gen_imagery` → spawned image-generator pod. |
| 10:22:24 | Child orchestration `a64f9d18` created in the spawned image-generator. |
| 10:22:24 | `call_imagery_gen` invoked — `imagery_direction` prepended (Phase 0.1 path), full prompt sent to adapter. |
| 10:22:34 | Adapter completed Stability call, returned image_uri (~10s SDXL generation). Asset row `a12b5d71` written to `assets` with correct purpose, asset_key, origin_prompt. |
| 10:22:36 | Child orchestration COMPLETED. |
| 10:23:04 | Parent orchestration COMPLETED — asset deployed to git as `assets/images/hero-home.jpg`. |

Total wall time: 58 seconds for one hero generation including spawn-pod cold-start.

### Phases delivered

| Phase | Status | Verified |
|---|---|---|
| 0.1 — read imagery_direction | ✅ delivered | 2026-05-08 |
| 0.2 — populate origin_model / origin_prompt | ✅ delivered | 2026-05-08 |
| 1.1 — `check_unfulfilled_image_prompt` | ✅ delivered | 2026-05-08 |
| 1.2 — `check_placeholder_image_in_use` | ✅ delivered | partial (check fires; symptom-path site needed) |
| 1.3 — `check_image_url_404` | ✅ delivered | partial (as above) |
| 1.4 — register checks in design-discovery-agent | ✅ delivered | 2026-05-08 |
| 1.5 — handler dispatch smoke test (+ hotfix) | ✅ delivered | 2026-05-08 |
| 2A — `assets.locked_at` + `locked_by` columns | ✅ delivered | 2026-05-09 |
| 2B — `assets.asset_key` column + new unique index | ✅ delivered | 2026-05-09 |
| 2C — `store_asset` writes asset_key; ON CONFLICT switch | ✅ delivered | 2026-05-09 |
| 2D — drop old `(site_id, purpose)` unique constraint | ✅ delivered | 2026-05-10 |
| 2E — variant path through image-build-handler | ✅ delivered | 2026-05-12 |
| 2F — image-build-handler spawns asset-deployer | ✅ delivered | 2026-05-12 |
| 2G.1 — `site_plan_imagery` table | ✅ delivered | 2026-05-12 |
| 2G.2 — `write_site_plan` extension + `flattenImageryBlock` | ✅ delivered | 2026-05-12 |
| 2G.3 — planner prompt extension | ✅ delivered | 2026-05-13 (with max_tokens bump + path fix) |
| 2G.4 — `check_unfulfilled_imagery_plan` discovery check | ✅ delivered | 2026-05-14 (8 work items emitted on first run; correct priority ordering) |
| 2G.5 — image-build-handler `needs_imagery` branch | ✅ delivered | 2026-05-14 (with two hotfixes — optional input_mapping + store_asset purpose; see Application Status table) |
| 2G.6 — legacy `image_prompts` age-out | ❌ retired as scoped | Reframed 2026-05-13; see Decisions table |
| 2H — image generator request shape | ✅ delivered (action layer) | 2026-05-14 partial — chassis confirmed running new code; adapter binary unconfirmed (see Application Status table) |
| 3 — adoption image mirror | ⏸ deferred 2026-05-14 | Not cancelled. Reason: adopted sites currently carry minimal or no imagery, so the visual-continuity gain Phase 3 provides doesn't apply to any active customer. Revisit when adoption volume includes image-heavy sites, or when a client specifically raises brand-continuity concerns about regenerated imagery. The `reference_image_uri` plumbing through 2G + 2H is preserved as a forward-compat hook — no architectural debt accumulating from the wait. |
| 4 — visual auditor sees imagery (text-only) | — not started | — |
| 5 — vision-capable LLM path | — not started | — |
| 6 — `imagery-quality-auditor` agent | — not started | — |

### Known issues, not blocking

Carried forward from prior status notes and the 2026-05-12 / 2026-05-13 verification runs:

- **Dispatch loop not claiming triaged image items.** When page work is also in queue, image work items sit in `triaged` indefinitely. Manual triggering works; investigation pending. Not blocking 2G (the loop is currently off by design).
- **Retry semantics gap.** Failed items sit `failed` until manually cleaned. Dedup index allows duplicates between `complete` rows but not between `triaged`+`failed`, which makes resets awkward. Worth pinning down the exact index definition before leaning on it.
- **`content_components.description` missing on four rows.** Verification of step 2 surfaced `<no value>` template warnings for `ad_zone_inline`, `category-listing`, `content-listing`, `featured-article`. The component lines render as e.g. `content-listing (Article Grid): content-listing - <no value>` in the planner prompt. Data hygiene fix — one UPDATE per row populating `description` with a one-sentence summary.
- **`llm_call_log.agent_type` populated as empty across all calls.** `params.AgentType` isn't being passed through to `LogLLMCall` (also flagged in doc 009). Increasingly painful for diagnostic queries.
- **FAILED orchestrations accumulate in `orchestration_states`.** No auto-cleanup or admin wrapper. Worth periodic archival sibling of `site_work_items_archive`, or an admin function to purge FAILED rows older than a threshold.
- **Variant chain doesn't pass `site_id` to image-generator.** As a result, hero variants don't get `design_intent.imagery_direction` prepended. Discovered during step 5 design. Cosmetic — variants still generate — but a small one-line fix to `call_variant_gen.input_mapping` in the image-build-handler workflow.
- **`store_asset` doesn't support `purpose_field` config key.** Discovered 2026-05-14 during step 5 verification. Only `asset_key_field`, `site_id_field`, `data_field`, `origin_prompt_field` are attested `*_field` variants. The needs_imagery branch currently hardcodes `purpose: "hero"` in both store steps as a workaround, which means **kind=logo work items can't use the new branch yet** — they would write the asset row with purpose=hero and be deployed as JPG instead of PNG-with-transparency. Two ways to fix properly: (a) add `purpose_field` lookup support to the `store_asset` action — small Go change, ~5 lines; (b) branch the workflow by kind before the store step — verbose but no Go change. Until either is done, route any kind=logo imagery via the legacy `unfulfilled_image_prompt` → `needs_logo` path (which still works).
- **image-build-handler doesn't update `site_work_items.status` on completion.** When triggered manually (bypassing dispatch), the work item stays in `detected` after the orchestration completes and the asset deploys. Dispatch normally does the status update — but the handler itself should also do it when `input_data.work_item_id` is set. Otherwise: future discovery passes see the asset exists and don't re-emit, but the work item lingers forever in `detected`. Manual cleanup query in the debugging guide. Real fix: add a `mark_work_item_complete` step before the `complete` terminal step in image-build-handler's workflow, conditional on `input_data.work_item_id != null`. Same pattern presumably needed by other agents that may be triggered outside dispatch.
- **Adapter deployment vs chassis deployment.** Discovered 2026-05-14 during Phase 2H verification. The image-generator adapter (`dynamic_adapter.go`) appears to be deployed as a separate Kubernetes resource from the chassis (`generate_image_actions.go`). After a chassis rebuild + rollout, the adapter may still be running an older binary. Symptom: the `generateImage: stability request` log line added in 2H wasn't found on adapter pods. Worth documenting which deployment carries which binary, and adding the adapter to the rebuild/rollout sequence.

---

## Sequencing principles

- Each phase is shippable on its own. No phase depends on a later phase.
- Schema changes are isolated to dedicated sub-phases; behavioural phases run against a stable schema.
- LLM-side changes (Phase 4+) are gated on the algorithmic checks (Phase 1) working — the auditor should not have to re-discover what algorithms already caught.
- Each step within a phase is sized to a single PR / single deploy. No step touches more than 2-3 files.
- Reuse is called out per step. New code is the exception, not the default.

---

## Phase 0 — Wire what's already written but unread ✅ delivered

Goal: stop ignoring data we already have. Demonstrates the path is live, no schema change.

### 0.1 Read `imagery_direction` into the image prompt ✅

`platform/orchestration/actions/generate_image_actions.go`. After resolving the prompt via three-tier priority, look up `site_specs.aspect = 'design_intent'` for the current site, extract `data.imagery_direction`, prepend as `"Style direction: <direction>\n\nSubject: <prompt>"`. Verified on the robot-hands.com hero generation — `imagery_direction` tokens appear in `assets.origin_prompt`.

### 0.2 Populate `origin_model` in `store_asset` ✅

Reads the active model from `agent_config["model"]` or agent_definition's `default_config.model`. Writes to `assets.origin_model`. Verified `'sdxl'` populating on new hero/logo rows.

---

## Phase 1 — Algorithmic discovery checks for spec-to-delivery gap ✅ delivered

Goal: catch simple structural mismatches without LLM cost. All three live in `platform/orchestration/actions/discovery_checks/`, follow the existing `DiscoveryCheck` interface, register via `init()`.

### 1.1 `check_unfulfilled_image_prompt.go` ✅

Detects `site_specs.site_plan.image_prompts` keys with no matching `assets` row. Routes to existing `image-build-handler`. Spec copies `image_prompts` so the handler workflow finds them at `input_data.spec.image_prompts.hero_home`.

### 1.2 `check_placeholder_image_in_use.go` ✅

Detects `page_components.rendered_html` containing fallback paths (`/assets/images/hero.jpg`, `/assets/images/logo.png`) with no matching `assets` row. Same handler.

### 1.3 `check_image_url_404.go` (DB-only version) ✅

Detects `<img>` references to `/assets/images/X.{jpg,png,webp}` where X doesn't match any `assets.purpose` (or now `asset_key`) for the site. Best-effort routing.

### 1.4 Wire all three into `design-discovery-agent` ✅

Appended to `default_config.workflow.steps.run_checks.config.checks`. Config-only change.

### 1.5 `image-build-handler` smoke test ✅ (with hotfix)

Verified end-to-end. **Unplanned hotfix:** `image-build-handler.call_logo_gen` and `call_hero_gen` were missing the `output_mapping` block that `pageflow-builder` had — without it, the generator's URL lived three levels deeper than `store_asset` expected, so every dispatch-path image asset silently failed to store. SQL-only fix; files `phase_1_5_pre_migration_backup.sql`, `phase_1_5_image_build_handler_output_mapping.sql`.

---

## Phase 2 — Schema groundwork and pipeline refactor ✅ 2A–2F delivered; 2G in flight; 2H not started

Phase 2 expanded beyond its original sketch. The sub-phases below reflect what actually shipped. 2A–2D are the original "locking and multi-image readiness" work, broken into deployable units. 2E–2F handled variant routing and the storage-credential mismatch. 2G is the substantive new work that this section originally deferred.

### 2A — Asset locking columns ✅

`assets.locked_at timestamptz`, `assets.locked_by text`, partial index `idx_assets_locked WHERE locked_at IS NOT NULL`. Mirrors `page_components` exactly. Lock semantics: detection via `locked_at IS NULL`; classification via `locked_by` value (hard = `admin`/`admin-removed`/`checkpoint`; soft = `deploy`/`manual`/auditor names/other). Reference helper `check_component_lock.go`. No time-based expiry today.

### 2B — `asset_key` column for multi-image readiness ✅

`assets.asset_key text` (nullable), backfilled from `purpose` for existing rows. New unique index `idx_assets_site_asset_key_unique (site_id, asset_key) WHERE asset_key IS NOT NULL AND status = 'active'`. Old `(site_id, purpose)` constraint stayed in place during this phase. Both coexisted harmlessly while `asset_key = purpose` for everything.

### 2C — `store_asset` writes `asset_key`; ON CONFLICT switch ✅

Go-only change to `StoreAssetAction`. Accepts optional `asset_key` config; defaults to `purpose`. ON CONFLICT target switched from `(site_id, purpose)` to `(site_id, asset_key)`. No behaviour change at deploy time — forward-compatible scaffolding for 2D.

### 2D — Drop old `(site_id, purpose)` unique constraint ✅

`DROP INDEX idx_assets_site_purpose_unique;`. Sanity checks abort the migration if any active row has `purpose IS NOT NULL AND asset_key IS NULL`, or if `idx_assets_site_asset_key_unique` is missing. Multi-image case (same `purpose`, distinct `asset_key`) becomes possible at the schema level.

### 2E — Variant path through `image-build-handler` ✅

Makes hero variants (`hero_about`, `hero_tools`, `hero_matchmatrix`, ...) routable end-to-end. Five coordinated changes:

1. `check_unfulfilled_image_prompt.go` — variants now route to `image-build-handler` (previously flag-only). Spec includes `asset_key` and `prompt` directly. `(purpose, asset_key)` split: canonical hero = `purpose=hero, asset_key=hero`; variant `hero_about` = `purpose=hero, asset_key=hero_about`.
2. `imagery_helpers.go` — new `hasActiveAssetForAssetKey`. The classifier switched away from `hasActiveAssetForPurpose` (which would report "hero exists" when only `asset_key=hero` exists — wrong for `asset_key=hero_about`).
3. `deploy_image_asset_action.go` — accepts `asset_key`. When provided AND differs from purpose, derives per-variant deploy path: `assets/images/<asset_key with _ → ->.{ext}`.
4. `StoreAssetAction` — accepts `asset_key_field` config (JSONPath lookup), needed for the variant workflow which passes asset_key through `input_data.spec.asset_key`.
5. `image-build-handler` workflow migration — third branch for variant items. Existing logo and hero paths unmodified.

### 2F — `image-build-handler` spawns `asset-deployer` ✅

Resolves the storage architecture mismatch: chassis pod had no S3 credentials, so inline `deploy_image_asset` calls ran in-process and failed with "storage client not available". The migration replaces three inline deploy step pairs with spawn+call patterns targeting `asset-deployer`. Storage env flows naturally because `asset-deployer` is in `isStorageEnabledAgent` — the spawn action injects S3_ENDPOINT, IMAGE_BUCKET into the child pod's environment.

**Side-fix during 2F verification:** loader-snapshot defect. Two agent-definition loaders used `ORDER BY version DESC LIMIT 1` without filtering snapshots; they were picking up rollback snapshots with version+1000 offset instead of the active row. Patched `processor.go::loadAgentDefinition` and `spawn_actions.go::getAgentDefinition` to add `is_active = true AND (is_snapshot IS NULL OR is_snapshot = false)`. Snapshot audit on production data returned zero affected rows outside the test site.

**Adapter side-fix:** `imagegenerator/dynamic_adapter.go` — HTTP timeout raised from 30s to 120s (SDXL response body reads were exceeding 30s causing tear-downs); zap marshal error fixed by replacing `zap.Any("request", req)` with three safe fields.

### 2G — Structured imagery in the plan domain — steps 1-3 delivered; 4-5 designed; 6 not started

Detail in `PLAN_imagery_phase_2g.md`. Six components:

1. **New table `site_plan_imagery`** — schema applied 2026-05-12. ✅
2. **`write_site_plan` extension** — `flattenImageryBlock` + `insertImageryRow` + lock transfer. ✅ deployed 2026-05-12; path fix on 2026-05-13 (the function was looking up `data["imagery"]` at top level rather than walking the wrapper shapes via `findDirectiveTree` — same pattern the directive writer uses).
3. **Planner prompt extension** — taught build-site-planner to emit the `imagery` block. ✅ delivered 2026-05-13. Two corrective patches surfaced during verification: max_tokens raised from 4000 to 8000 (the imagery block plus 14-page roadmap output blew the previous cap, causing JSON truncation in `validate_site_plan`), and the `flattenImageryBlock` path fix noted above.
4. **`check_unfulfilled_imagery_plan` discovery check** — Go check sibling of `check_unfulfilled_image_prompt`. Reads current plan's imagery rows, emits one `needs_imagery` work item per row whose `asset_key` is missing. Per-pass cap of 20; priority ordered (site logo → index hero → page-scope hero → page-scope non-hero → section-scope). Spec carries `scope`, `key`, `kind`, `asset_key`, `purpose`, `prompt`, `brand_update`, plus optional `style_hints` / `constraints`. 📦 designed 2026-05-13.
5. **`image-build-handler` `needs_imagery` branch** — new branch alongside the existing variant/logo/hero chains rather than extending the variant matcher. Six new workflow steps, shared asset-deployer tail. Reads `purpose` from spec (so logo/hero/illustration/icon/infographic all flow through the same chain); routes brand-asset update via the `spec.brand_update` boolean computed by step 4. 📦 designed 2026-05-13.
6. **Legacy `image_prompts` age-out (no dedicated check)** — original plan called for `check_legacy_image_prompts_aspect` to detect sites still on the legacy path. Reframed 2026-05-13: "is this site on legacy?" isn't a useful signal because legacy isn't inherently a fault. What matters is whether something is broken, and `unfulfilled_image_prompt` already detects its respective gaps. Both checks run in parallel during transition. When `unfulfilled_image_prompt` reliably finds zero gaps across all active sites (which happens naturally once every site has had a planner run post-2G.3), it gets pulled from the `design-discovery-agent.run_checks` list — one operational decision, no code. Could be a year from now; could be never. Running both is cheap.

#### Step 5 workflow (post-deploy)

The new `needs_imagery` branch sits in front of the existing routing tree. Legacy `unfulfilled_hero_variant`, `needs_logo`, `needs_hero_image` items still fall through to their existing chains.

```
ensure_site_record
    │
    ▼
check_item_type_imagery ──[needs_imagery?]── yes ──► spawn_image_gen_imagery
    │                                                  │
    │ no                                               ▼
    │                                              call_imagery_gen
    │                                                  │  (site_id passed → imagery_direction prepended)
    │                                                  │  (kind/style_hints/constraints pass-through)
    │                                                  ▼
    │                                              check_imagery_brand_update
    │                                                  │
    │                                                  ├─[spec.brand_update=true]──► store_imagery_brand_asset
    │                                                  │                                   │
    │                                                  └─[else]──► store_imagery_asset    │
    │                                                                       │             │
    ▼                                                                       └─────┬───────┘
check_item_type ──[variant?]── yes ──► spawn_image_gen_variant ... store_variant_asset ──┤
    │                                                                                     │
    │ no                                                                                  │
    ▼                                                                                     │
check_logo_or_hero ──[logo?]── yes ──► spawn_image_gen ... store_logo_asset ─────────────┤
    │                                                                                     │
    │ no                                                                                  │
    ▼                                                                                     │
spawn_image_gen_hero ... store_hero_asset ───────────────────────────────────────────────┤
                                                                                          │
                                                                                          ▼
                                                                                spawn_asset_deployer
                                                                                          │
                                                                                          ▼
                                                                                 call_asset_deployer
                                                                                          │
                                                                                          ▼
                                                                                       complete
```

All four routing branches (needs_imagery, variant, logo, hero) converge on the same `spawn_asset_deployer → call_asset_deployer → complete` tail, so the deployer chain is unchanged. The asset-deployer reads `asset_key` to derive the per-variant deploy path (e.g. `assets/images/hero-about.jpg`), which works identically for `needs_imagery` items.

#### Step 4 priority bands and cap

The discovery check emits at most 20 items per pass, ordered by classifier-assigned priority. Bands mirror the legacy `classifyPromptKey` bands in `check_unfulfilled_image_prompt.go` so the two checks produce comparable queue orderings during the transition window:

| Imagery row pattern | Priority | Severity | Notes |
|---|---|---|---|
| `scope=page` AND `scope_ref=index` AND `kind=hero` | 65 | high | Mirrors `hero_home` |
| `scope=site` AND `kind=logo` | 70 | high | Mirrors `logo` |
| `scope=site` (other kinds) | 75 | high | Site-wide supporting imagery |
| `scope=page` AND `kind=hero` (non-index) | 80 | medium | Variant heroes |
| `scope=page` (non-hero) | 90 | medium | Page decoratives |
| `scope=section` (any kind) | 100 | low | Decoratives, icons, infographics |

#### Image prompt cascade — deferred

Step 5 keeps the existing single-prepend cascade (subject prompt + `imagery_direction` prepended via `composeImagePromptWithDirection`). It does not match the richer composition pattern `page-content-writer` uses for text generation. That asymmetry is intentional; see `FOCUS_prompt_composition_pattern.md` for the considered opinion on why the text pattern is not a good target to copy and what a better image cascade might look like. The eventual landing place for image cascade is more likely Phase 2H (parameter-shaping in the image-generator request) than a step-5-style prompt extension.

### 2H — Image generator request shape — design confirmed, code awaiting adapter source

Scope confirmed 2026-05-13. Five sub-phases, all Go-side (no SQL migrations except for the small workflow tweak in 2H.4).

**In scope for 2H v1:**

- `negative_prompt` — biggest visible quality lift, particularly for logos (forbid people/text/watermarks).
- `seed` — reproducibility for HITL retry and deterministic testing.
- `reference_image_uri` — passes through but no callers use it yet; unblocks Phase 3.1 img2img.
- `cfg_scale` and `steps` — per-kind SDXL knobs.

**Deferred to 2H follow-up or later:**

- `style_preset` — no adapter mapping exists today (confirmed 2026-05-13). Stability-specific, finicky behaviour; adding the field means deciding what each kind maps to and verifying against Stability's documented presets.
- `samples` — multi-output, low priority.
- `safety_mode` — Stability-specific, low priority for v1.
- Multi-provider routing — separate plan.
- The composer/envelope cascade — separate phase; see `FOCUS_prompt_composition_pattern.md`.

**Sub-phases:**

1. **2H.1 — Extend `GenerateImageAction` to read new fields from `inputData`.** Apply kind-aware defaults when fields are absent. Include them in `imageData` (the body.data payload).
2. **2H.2 — Adapter accepts and forwards new fields.** `platform/adapters/imagegenerator/dynamic_adapter.go` parses the extended `body.data` and maps it onto the Stability API call.
3. **2H.3 — Per-kind defaults table.** A Go-side `map[string]imageDefaults{}` keyed by `kind`. Logos get `negative_prompt: "people, faces, text, signature, watermark"`. Icons get tighter aspect and background suppression. Heroes get the existing behaviour as their default.
4. **2H.4 — Legacy callers pick up kind-derived defaults too.** Add `kind` to the `input_mapping` of `call_logo_gen`, `call_hero_gen`, `call_variant_gen`. Three tiny workflow JSON edits. After this, even legacy paths benefit from kind-derived defaults.
5. **2H.5 — End-to-end test on robot-hands.com.** Trigger one item of each kind. Verify `assets.origin_prompt` shows the constructed prompt; eyeball results (especially logos — biggest visible win expected).

**Design decisions taken:**

- **Defaults live Go-side** (per-kind map), not in a config table. Simpler, no migration needed. Move to a table if per-vertical or per-site overrides matter later.
- **Spec overrides Go defaults.** Caller-supplied fields in `inputData` take precedence. Use case: HITL operator retries with a specific seed.
- **`style_hints` and `constraints` from the imagery row become functional here.** Step 4 plumbs them through; 2H wires them to actual generation parameters. `constraints.no_text` → appends to negative_prompt; `style_hints.aspect_ratio` → drives width/height.

**Blocker before code lands:**

Need the current source of `platform/adapters/imagegenerator/dynamic_adapter.go` to write 2H.2. Specifically: the struct or map type the adapter unmarshals `body.data` into, the function that builds the Stability HTTP request, and any existing field-mapping config.

---

## Phase 3 — Adoption image mirror and many-images-per-page foundation — not started

Goal: stop discarding crawled imagery; lay the foundation for pages that carry many images each, not just one hero.

**Scope expanded (2026-05-12).** The original Phase 3 was solely the adoption mirror. With Phase 2G now delivering scoped imagery rows and the "many images per page" direction surfaced, Phase 3 also takes on the per-component imagery wiring that the original plan deferred. The adoption mirror remains the core deliverable; the additions below extend its reach.

### 3.1 New action `mirror_adoption_images`

`platform/orchestration/actions/mirror_adoption_images_action.go`.

Inputs: `site_id`, `research_results_id` (or read latest `research_results` row with `result_type = 'adoption_crawl'`).

Behaviour:

1. Read crawl result, extract image URLs from `images[]` and link-extension matches.
2. For each: download (existing `webscrape` or direct HTTP via circuit breaker), upload to S3 via `storage.S3Client.Upload`.
3. Insert one `assets` row per image:
   - `origin_type = 'adopted'`
   - `origin_url = <source URL>`
   - `purpose = 'adopted_image'`
   - `asset_key = 'adopted:<filename or hash>'`
   - `attribution`, `license` from any signal we can derive (often blank).

Reuse: S3 client, HTTP client with circuit breaker (`platform/resilience`), data extraction pattern from existing scrape-result handlers.

Limits: cap per site (e.g. 50 images) and per-image size (e.g. 5MB). Log skipped items; don't fail the action for individual download failures.

### 3.2 Wire into `apply_adoption_plan`

`agent_definitions` row for `site-adoption-agent`. New step `mirror_adoption_images` between `apply_adoption_plan` and `complete`. Acceptance: adopt a 12-image site → 12 `assets` rows with `origin_type='adopted'` after adoption completes.

### 3.3 Discovery check `check_crawled_images_discarded.go` (backfill)

Detects sites with `research_results.result_type = 'adoption_crawl'` containing image URLs but no `assets` rows with `origin_type = 'adopted'` for that site. Emits `item_type = 'crawled_images_discarded'` routed to a new `adoption-image-mirror` agent (3.4).

Sites adopted before Phase 3 lands need this. After all old sites are processed, the check becomes a no-op for new sites.

### 3.4 New agent `adoption-image-mirror`

Specialist category, orchestrator processing mode. One-step wrapper around 3.1's action:

```
ensure_site_record → mirror_adoption_images → complete
```

We make it an agent (not just a dispatch-loop action call) because that's the project pattern: every dispatched work item type maps to an agent.

### 3.5 Per-component image declarations begin to consume `site_plan_imagery` (NEW for the refresh)

The original Phase 3 stopped at 3.4. The "many images per page" direction extends Phase 3 to wire scoped imagery rows through to components.

`content_components.input_schema` v2 already supports image fields with `source: "site_assets.{key}"` resolvers. Today only `hero_image` uses this; most components ignore it. Phase 3.5 brings these declarations alive:

- **Components declare what they need.** A `team-grid` component declares `member_avatars: array of image with source: site_plan_imagery.page.{page}.team_avatar.{slug}`. A `services-grid` declares `service_icons` from `section.{page}:{section_id}.icon.{slug}`.
- **Renderer resolves them.** `BuildRenderContextAction` extends its image-URL extraction to read `site_plan_imagery` rows by scoped `asset_key`, then look up matching `assets` rows.
- **Discovery walks the gap.** Phase 2G's `check_unfulfilled_imagery_plan` already emits work items for any imagery row without a matching asset — so when a component declares 6 team avatars and the planner produces 6 page-scope imagery rows, the discovery check emits 6 work items, image-build-handler generates each, asset-deployer writes them to git as `assets/images/page/about/team-avatar-<slug>.jpg`.

This is the structural foundation for pages with many images. The actual generation chain is unchanged — variant routing handles arbitrary asset_keys. What's new is (a) components knowing how to consume scoped imagery and (b) the renderer doing the lookup.

Bounded scope for v1:

- One additional component type wired up as proof: probably `team-grid` (lowest risk, clearest signal).
- Renderer extraction extended to the new resolver pattern.
- No icon-resolver or infographic-generator yet (separate plans; deferred to Phase 3-followups or later).

### 3.6 Reference imagery available to image-generator (deferred hook)

No code change yet. Once `assets` has `origin_type='adopted'` rows from 3.1, downstream work (img2img, style references, future per-section illustrations) can read them.

Future hook: when image-generator gains `reference_image_uri` support (Phase 2H), prefer adopted images for matching purpose / section. The plan-builder can also use adopted images as visual reference when composing per-page imagery prompts — "match the visual style of `<adopted image URL>`" becomes a soft directive on relevant rows.

---

## Phase 4 — Visual auditor extension (text-only, cheapest LLM-side win) — not started

Goal: make the existing visual-design-auditor aware that imagery exists, without a vision model.

### 4.1 Extend `load_design_context` SQL

Add subqueries to the auditor's existing big SELECT:

- `(SELECT json_agg(json_build_object('purpose', purpose, 'asset_key', asset_key, 'url', url, 'origin_prompt', origin_prompt)) FROM assets WHERE site_id = s.id AND status = 'active' AND locked_at IS NULL AND origin_type IN ('generated', 'adopted')) as assets`
- `(SELECT data->>'imagery_direction' FROM site_specs WHERE site_id = s.id AND aspect = 'design_intent' AND is_current = true) as imagery_direction`
- Once 2G's planner extension is delivering: also load `site_plan_imagery` rows for the current plan so the auditor can see what was *supposed* to be there.

Risk: query becomes large. If it breaks the action's prompt template, split into two query steps.

### 4.2 Update auditor LLM prompt

Add `IMAGERY` as the 6th check category. Provide the new context. Pass algorithmic-check results through to avoid double-flagging.

### 4.3 Tune for false positives

Run against 5–10 existing sites without enabling fixes. Read findings. Triage obvious false positives against the prompt. Acceptance: ≥80% of imagery findings on a random sample are accurate.

---

## Phase 5 — Vision-capable LLM path — not started

Goal: foundational capability for an auditor to actually look at images. Required by Phase 6.

### 5.1 Extend `aiservice.AIService` interface

Add `GenerateTextWithImages(ctx, prompt, imageURLs, options)` or accept `image_urls` in the existing options map.

### 5.2 Implement Anthropic vision

`platform/aiservice/anthropic.go`. Build messages with image content blocks per Anthropic's vision API. Image URLs must be publicly accessible — our presigned B2 URLs work (valid 7 days), but expired ones won't. Auditor should refresh presigned URLs immediately before the call.

### 5.3 Workflow action wrapper

Lean toward extending `execute_llm_prompt` with an `image_urls_field` config rather than a new action — fewer registry entries, same workflow shape.

### 5.4 Cost monitoring

Vision calls are more expensive per call. Tag `vision_call: true` in `llm_call_log` metadata so cost analysis can separate them.

---

## Phase 6 — `imagery-quality-auditor` agent — not started

Goal: separate agent dedicated to imagery audit, vision-capable, sibling of `visual-design-auditor`.

### 6.1 Agent definition

Analyst category, orchestrator processing mode. Resource limits match `visual-design-auditor`. Workflow:

```
ensure_site_record
  → load_imagery_context        (SQL: assets, site_plan_imagery, imagery_direction, identity)
  → check_has_imagery           (conditional — skip if no assets to audit)
  → run_imagery_audit_vision    (execute_llm_prompt with image_urls)
  → write_audit_findings        (existing action, source = 'imagery-quality-audit')
  → complete
```

### 6.2 Audit prompt

Follow the existing TOP 5 / acceptance_test contract used by `visual-design-auditor`. Imagery-specific categories:

- `direction_mismatch` — image doesn't reflect `imagery_direction` or `style_hints`
- `brand_mismatch` — colour palette of generated image clashes with site palette
- `inconsistency` — multiple hero variants don't share a visual style
- `quality` — visible artefacts, garbled text in generated image, AI giveaways
- `inappropriate` — generated content unsuitable for the site's audience or tone

`max_fix_attempts: 2`. Handler routing:

- Direction / brand / inconsistency → `image-build-handler` regenerates with updated prompt
- Quality → `image-build-handler` regenerates with a different seed (depends on Phase 2H)
- Inappropriate → `image-build-handler` regenerates with stronger negative prompts; if 2 attempts fail → `needs_human_review`

### 6.3 Lock honouring

`load_imagery_context` filters `WHERE locked_at IS NULL AND origin_type IN ('generated', 'adopted')`. Human uploads (`origin_type = 'uploaded'`) excluded.

### 6.4 Hook into `design-audit-agent`

Parallel to existing visual and content auditor calls. Sequence: visual → content → imagery → triage. Identical pattern; copy-modify.

### 6.5 Pass count and rate limiting

Imagery findings count toward existing 3-pass cap on the improvement loop. `max_fix_attempts: 2` per finding. No new infrastructure.

### 6.6 Gated rollout

Deploy with `is_active = false` in `design-audit-agent` workflow. Verify on a test site. Flip live.

---

## Future direction — pages with many images (post-Phase-3)

The "single hero + logo" world is ending. With Phase 2G delivering scoped imagery rows and Phase 3.5 wiring per-component declarations, the natural growth is:

- **Pages declare richer imagery needs.** A services page wants per-service icons + a hero. A team page wants per-person portraits. A blog post wants a hero plus inline illustrations. The planner emits one `site_plan_imagery` row per slot.
- **Components own their imagery contracts.** `input_schema` declarations become the source of truth for what a component needs. Planner reads them when planning; renderer reads them when rendering; discovery walks them when auditing. No more silent fallthroughs to `/assets/images/hero.jpg`.
- **Generation scales horizontally.** Each imagery row is one work item; image-build-handler's variant chain already handles arbitrary asset_keys. A 30-image page is 30 parallel work items, not a special pipeline.
- **Audit becomes per-image.** With each row independently addressable, `imagery-quality-auditor` (Phase 6) can flag "the third team avatar looks AI-glossy" rather than "imagery looks off." Findings route back to the specific row.

Specific follow-on work this enables, not yet planned:

- **Icon-resolver path.** Features components declare `icon: string` but render no SVG. Once components consume scoped imagery rows, `kind=icon` rows can land — either inline SVG (LLM-generated or library-looked-up) or small raster. The icon component / template helper sits on top of this.
- **Infographic generator.** Separate generator agent producing SVG from structured data. Wired through the same `site_plan_imagery` shape with `kind=infographic`. Plan-builder emits one row per chart.
- **Multi-image components.** `team-grid.member_avatars: array of image` is one example. Once 3.5 wires the resolver pattern, any component declaring `array of image` works the same way.
- **Img2img reference paths.** Adopted images (3.1) become reference inputs for fresh generations. The plan-builder can flag "use this adopted image as reference style" on specific imagery rows.

These are all deferred but no longer blocked by structural gaps. The plumbing landing in Phases 2G and 3 is the foundation.

---

## What we're explicitly NOT doing in this plan

Updated for the refresh — several previously-deferred items have been pulled forward into Phase 2G or 3.

Still deferred:

- **Multi-provider image generation.** Stability remains the only provider. Banana Pro / FLUX / Midjourney is a separate plan, dependent on Phase 2H's request shape work.
- **True SVG icon generation.** `kind=icon` in the imagery schema produces raster in v1. Real SVG (vector, scalable, crisp at any size) needs a different generator.
- **Infographic generator agent.** `kind=infographic` is in the enum but renders as `illustration` for now. Real infographics need data-driven generators (chart libraries, structured input). Sketched in the future-direction section.
- **Vision-based hero variant deduplication** ("the same image is on three pages"). Natural extension of Phase 6, low priority until per-section imagery has been live for a while.
- **Per-vertical LoRA fine-tunes.** Pre-existing plan in `018_canine_biology.md`, unblocked by Phase 5's vision work but not dependent on it.
- **Updating pageflow-builder.** Decision (per `PLAN_imagery_phase_2g.md`): pageflow-builder is being left behind. Sites built via it keep using the legacy `check_unfulfilled_image_prompt` path until they age out.

No longer deferred (moved into active plans):

- ~~Per-section / per-component imagery~~ — delivered structurally by Phase 2G; component-side wiring is Phase 3.5.
- ~~Richer planner output~~ — Phase 2G's `imagery` block replaces the `{logo, hero_home}` contract.
- ~~Reading `design_intent.imagery_direction`~~ — delivered in Phase 0.1.

---

## Phase summary table

| Phase | Goal | New files | Modified files | Schema changes | Agents touched |
|---|---|---|---|---|---|
| 0 ✅ | Wire imagery_direction; populate origin_model | none | `generate_image_actions.go`, `store_asset_action.go` | none | none |
| 1 ✅ | Algorithmic discovery checks | three `check_*.go` files | `agent_definitions` (design-discovery-agent), `image-build-handler` (hotfix) | none | design-discovery-agent (config), image-build-handler (hotfix) |
| 2A–D ✅ | Asset locking + multi-image schema | none | `assets` table | `locked_at`, `locked_by`, `asset_key`, new index; old constraint dropped | none |
| 2E ✅ | Variant routing | `imagery_helpers.go` | `check_unfulfilled_image_prompt.go`, `deploy_image_asset_action.go`, `StoreAssetAction`, `image-build-handler` workflow | none | image-build-handler |
| 2F ✅ | Spawn asset-deployer | none | `image-build-handler` workflow, `processor.go`, `spawn_actions.go`, `dynamic_adapter.go` | none | image-build-handler, asset-deployer (input_schema) |
| 2G ✅ | Structured imagery in plan domain | discovery check, planner imagery block, image-build-handler branch | `write_site_plan_action.go`, planner prompt template, `image-build-handler` workflow, `design-discovery-agent` run list | `site_plan_imagery` table | build-site-planner, design-discovery-agent (config), image-build-handler |
| 2H | Image generator request shape | none | `generate_image_actions.go`, `dynamic_adapter.go` | none | image-generator |
| 3 | Adoption mirror + per-component wiring | `mirror_adoption_images_action.go`, `check_crawled_images_discarded.go`, new agent | `agent_definitions` (site-adoption-agent), `BuildRenderContextAction`, one example component's input_schema | none | site-adoption-agent, new `adoption-image-mirror` |
| 4 | Visual auditor sees imagery | none | `agent_definitions` (visual-design-auditor) | none | visual-design-auditor |
| 5 | Vision-capable LLM path | maybe `execute_llm_prompt_with_vision_action.go` | `aiservice/interface.go`, `aiservice/anthropic.go`, possibly `execute_llm_prompt_action.go` | none | none directly |
| 6 | imagery-quality-auditor agent | none (workflow in agent_definitions) | `agent_definitions` (design-audit-agent) | none | new `imagery-quality-auditor`, design-audit-agent |

---

## Risks and mitigations

| Risk | Phase | Mitigation |
|---|---|---|
| Discovery check spec format mismatch with image-build-handler | 1.5 ✅ | Resolved during Phase 1 verification — hotfix added output_mapping. |
| `assets` UNIQUE constraint blocks adoption mirror | 2B → 3.1 ✅ | 2B/2D landed; constraint replaced with `(site_id, asset_key)` form. |
| Planner emits imagery rows that violate DB CHECK constraints | 2G | Prompt teaches the LLM the allowed values verbatim; `validate_site_plan` adds a pre-write structural check before `write_site_plan` runs. |
| Legacy and new imagery checks both fire on the same gap | 2G transition | Both checks call `hasActiveAssetForAssetKey` before emitting. If the legacy check produces an asset first, the new check skips emission for that asset_key (and vice versa). Dedup index handles work-item-level collisions. No double work, no double cost. |
| Visual auditor false positives on imagery findings | 4.3 | Manual sample run on 5-10 sites before enabling fixes. |
| Vision API URL expiry | 5.2 | Refresh presigned URLs immediately before the call. |
| Audit pass thrash on imagery findings | 6.5 | `max_fix_attempts: 2` per finding; existing 3-pass site cap. Both already in place. |
| Cost spike from vision calls | 5.4, 6 | Tag `vision_call: true` in `llm_call_log`. Monitor in week 1 of Phase 6 rollout. |
| Adoption mirror downloads exceed B2 storage budget | 3.1 | Per-site image cap (50); per-image size cap (5MB); skip-don't-fail on individual failures. |

---

## Open items / future phases (not part of this plan)

- **Image generator request shape (Phase 2H).** Needed before generation can vary by `kind`. Currently every image is an SDXL `cfg_scale: 7, steps: 30` call.
- **Multi-provider image generation.** Banana Pro, Midjourney, FLUX. Separate plan; depends on 2H.
- **Icon library and inline SVG resolver.** Foundation laid by Phase 3.5; actual icon path is a separate plan.
- **Infographic generator.** Separate plan, may share scaffolding with icon work.
- **Per-vertical LoRA fine-tunes** (`018_canine_biology.md`). Unblocked by Phase 5's vision work but not dependent on it.
- **Vision-based hero variant deduplication.** Natural extension of Phase 6.
- **Dispatch loop priority tuning.** Known issue: triaged image items wait behind page work indefinitely. Investigation pending; not blocking 2G.
- **Retry semantics for failed work items.** Failed items sit `failed` until manually cleaned. Worth pinning down the dedup index definition exactly before leaning on it further.
- **`content_components.description` backfill** (small data hygiene). Four rows: `ad_zone_inline`, `category-listing`, `content-listing`, `featured-article` — populate with one-sentence summaries so the planner prompt template renders cleanly.
- **`idx_site_plans_current` already exists.** The partial unique index in `site_plans_unique_current_index.sql` is redundant — the index was added in a prior migration not yet reflected in the file index. The revert function `revert_site_plan_to()` and the `site_plans_history` view from that file are still net-new and worth applying.
- **`llm_call_log.agent_type` populated as empty across all calls.** `params.AgentType` isn't being passed through to `LogLLMCall` (also called out in doc 009 as a known low-priority fix). Increasingly painful for diagnostic queries — the orchestration_states table works around it via `owner_agent_type`, but anywhere we filter llm_call_log by agent is broken. Small chassis-side fix.
- **FAILED orchestrations accumulate in `orchestration_states`.** No auto-cleanup or admin wrapper. Probably cosmetic, but the table grows and diagnostic queries get noisier. Worth either a periodic archival job (sibling of `site_work_items_archive`) or at least an admin function to purge FAILED rows older than a threshold.
- **Variant chain doesn't pass `site_id` to image-generator.** Discovered during 2G step 5 design. `call_variant_gen.input_mapping` is missing `site_id`, so variant heroes don't get `imagery_direction` prepended via `getImageryDirectionForSite`. One-line fix to the image-build-handler workflow. The new `needs_imagery` branch already passes site_id correctly.
- **Image prompt composition pattern revisit.** Step 5 deferred matching the page-content-writer's text cascade pattern; see `FOCUS_prompt_composition_pattern.md`. Considered opinion: the text pattern itself is fragile and shouldn't be copied wholesale. A composer step that produces a parameter envelope (for both text and images) is the strongest candidate when this is revisited. Likely lands in Phase 2H or a sibling phase rather than as an extension of 2G.
- **Unify legacy and new imagery routing under `needs_imagery`.** A future refactor option surfaced during 2G.6 reframing. Today the legacy check emits three item_types (`needs_logo`, `needs_hero_image`, `unfulfilled_hero_variant`) routing through three branches in image-build-handler. Could be collapsed: have `check_unfulfilled_image_prompt` emit `needs_imagery` items with the right spec.kind / spec.purpose / spec.asset_key fields, then retire the three legacy branches. One-line change to `classifyPromptKey`. Reduces image-build-handler workflow surface area. Not urgent — both paths work and don't conflict — but worth doing eventually as part of simplifying the handler. "Always fix legacy with modern" framing.
