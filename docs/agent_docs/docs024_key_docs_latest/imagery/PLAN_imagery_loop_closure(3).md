# PLAN — Imagery Audit and Fix Loop

Sequenced work plan for closing the gap between what the planner / spec / other agents ask for in imagery, and what is actually delivered. Builds on `FOCUS_imagery_assessment.md`.

Sister docs (read alongside):

- `PLAN_imagery_phase_2g.md` — detailed plan for Phase 2G (current work; structured imagery rows in the plan domain).
- `PLAN_product_illustration.md` — sibling: product imagery via `affiliate_products` resolver. Not part of this loop, but interacts at the renderer level.
- `STATUS_imagery_2026-05-12.md` — current operational status with verification trail.

---

## Decisions taken

| Question | Decision |
|---|---|
| Imagery audit — extend existing or new agent? | New agent: **`imagery-quality-auditor`**, sibling of `visual-design-auditor` under `design-audit-agent`. |
| Max regeneration attempts per finding | **2** (matches the existing `max_fix_attempts: 2` on structured findings). |
| Asset locking | **Mirror `page_components` exactly**: `locked_at timestamptz` + `locked_by text` on `assets`, same query exclusion patterns. Hard-vs-soft via `locked_by` (per 013 / 031). Timed expiry deferred. |
| Per-section / per-component image granularity | **Un-deferred (2026-05-12).** The original plan deferred this; Phase 2G now delivers it via `site_plan_imagery` with `scope ∈ {site, page, section}`. Section scope is wired but populated sparingly in v1. |
| Structured imagery — site_specs or site_plan? | **`site_plan` domain.** Sibling table `site_plan_imagery`, not a `site_specs` aspect. Resolved 2026-05-12; see `PLAN_imagery_phase_2g.md` § "Decisions taken". |
| Migration off legacy `image_prompts` | **Age-out** via `check_legacy_image_prompts_aspect`, registered last. |
| Planner — single LLM call or two? | **Single.** The imagery block is a new key in the existing plan-builder JSON output. No separate "imagery planner" call. |

---

## Progress (updated 2026-05-12)

Phases 0, 1, and the full 2A–2F sub-tree are delivered and end-to-end verified. Phase 2G is in flight (step 3 of 6 — the planner prompt extension — actively being drafted). Phases 2H, 3, 4, 5, 6 not started.

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
| 2G — structured imagery in the plan domain | ⏳ in flight | step 3 of 6 (planner prompt extension) |
| 2H — image generator request shape | — not started | — |
| 3 — adoption image mirror | — not started | — |
| 4 — visual auditor sees imagery (text-only) | — not started | — |
| 5 — vision-capable LLM path | — not started | — |
| 6 — `imagery-quality-auditor` agent | — not started | — |

### Known issues, not blocking

Carried forward from prior status notes and the 2026-05-12 verification run:

- **Dispatch loop not claiming triaged image items.** When page work is also in queue, image work items sit in `triaged` indefinitely. Manual triggering works; investigation pending. Not blocking 2G (the loop is currently off by design).
- **Retry semantics gap.** Failed items sit `failed` until manually cleaned. Dedup index allows duplicates between `complete` rows but not between `triaged`+`failed`, which makes resets awkward. Worth pinning down the exact index definition before leaning on it.
- **`content_components.description` missing on four rows.** Verification of step 2 surfaced `<no value>` template warnings for `ad_zone_inline`, `category-listing`, `content-listing`, `featured-article`. The component lines render as e.g. `content-listing (Article Grid): content-listing - <no value>` in the planner prompt. Data hygiene fix — one UPDATE per row populating `description` with a one-sentence summary.

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

### 2G — Structured imagery in the plan domain ⏳ in flight

Detail in `PLAN_imagery_phase_2g.md`. Six components:

1. New table `site_plan_imagery` — schema applied 2026-05-12. ✅
2. `write_site_plan` extension — `flattenImageryBlock` + `insertImageryRow` + lock transfer. ✅ deployed (dormant until step 3).
3. **Planner prompt extension** — teach build-site-planner to emit the `imagery` block. **Currently being drafted.** This is the first behavioural change in 2G.
4. New discovery check `check_unfulfilled_imagery_plan` — walks `site_plan_imagery` rows, emits work items.
5. `image-build-handler` extension — accepts `kind`, `style_hints`, `constraints`; routes through the existing variant chain.
6. Age-out check `check_legacy_image_prompts_aspect` — written but not registered until 1–5 are stable.

### 2H — Image generator request shape — not started

Goal: extend the image-generator request shape beyond `{prompt, width, height}` to support `negative_prompt`, `style_preset`, `seed`, `reference_image_uri`, `samples`, `safety_mode`. Most are no-ops on most providers; just get passed through if supported. Unblocks per-kind generation tuning, img2img references (Phase 3 dependency), and provider differentiation. Not blocking 2G.

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
| 2G ⏳ | Structured imagery in plan domain | discovery check, age-out check | `write_site_plan_action.go`, planner prompt template, `image-build-handler` workflow | `site_plan_imagery` table | build-site-planner, design-discovery-agent (config), image-build-handler |
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
| Age-out check fires before planner is reliably emitting imagery | 2G.6 | Check registered last; only after 2G.1–5 stable. |
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
