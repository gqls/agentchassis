-- ============================================================================
-- Phase 0 imagery work — pre-migration backup
--
-- Backs up the three agent_definitions rows that phase_0_combined_migration.sql
-- will modify. Run this BEFORE applying that migration.
--
-- Naming: follows the per-agent scoped pattern from 006_news_feed_pipeline_v2.md
--         (agent_def_<short>_backup_<YYYYMMDD>). The "_pre_phase0_imagery"
--         suffix disambiguates from any other 20260505 backups that may exist
--         or be created today by unrelated work.
--
-- Discipline (per 009_model_infrastructure.md):
--   - No DROP TABLE IF EXISTS. If a name collides, pick a new suffix.
--   - The collision IS the safety net.
-- ============================================================================

BEGIN;

-- ----------------------------------------------------------------------------
-- 1. site-work-orchestrator
-- ----------------------------------------------------------------------------
CREATE TABLE agent_def_site_work_orchestrator_backup_20260505_pre_phase0_imagery AS
SELECT * FROM agent_definitions
WHERE type = 'site-work-orchestrator';

-- ----------------------------------------------------------------------------
-- 2. pageflow-builder
-- ----------------------------------------------------------------------------
CREATE TABLE agent_def_pageflow_builder_backup_20260505_pre_phase0_imagery AS
SELECT * FROM agent_definitions
WHERE type = 'pageflow-builder';

-- ----------------------------------------------------------------------------
-- 3. image-build-handler
-- ----------------------------------------------------------------------------
CREATE TABLE agent_def_image_build_handler_backup_20260505_pre_phase0_imagery AS
SELECT * FROM agent_definitions
WHERE type = 'image-build-handler';

-- ----------------------------------------------------------------------------
-- Verification — each backup table should contain exactly one row matching
-- the corresponding live agent_definitions row
-- ----------------------------------------------------------------------------
SELECT 'site-work-orchestrator' AS agent,
       (SELECT COUNT(*) FROM agent_definitions WHERE type = 'site-work-orchestrator') AS live,
       (SELECT COUNT(*) FROM agent_def_site_work_orchestrator_backup_20260505_pre_phase0_imagery) AS backup
UNION ALL
SELECT 'pageflow-builder',
       (SELECT COUNT(*) FROM agent_definitions WHERE type = 'pageflow-builder'),
       (SELECT COUNT(*) FROM agent_def_pageflow_builder_backup_20260505_pre_phase0_imagery)
UNION ALL
SELECT 'image-build-handler',
       (SELECT COUNT(*) FROM agent_definitions WHERE type = 'image-build-handler'),
       (SELECT COUNT(*) FROM agent_def_image_build_handler_backup_20260505_pre_phase0_imagery);

COMMIT;


-- ============================================================================
-- RESTORE — only if reverting Phase 0 SQL changes
--
-- Pattern matches 006_news_feed_pipeline_v2.md restoration:
--   UPDATE agent_definitions
--   SET default_config = (SELECT default_config FROM <backup_table> LIMIT 1),
--       updated_at = NOW()
--   WHERE type = '<agent_type>';
--
-- This restores default_config only — leaves anything else (resource_limits,
-- topic_pattern, etc.) at whatever current values are. If a full restore is
-- needed, use the row directly from the backup table.
-- ============================================================================
-- BEGIN;
--
-- UPDATE agent_definitions
-- SET default_config = (
--     SELECT default_config
--     FROM agent_def_site_work_orchestrator_backup_20260505_pre_phase0_imagery
--     LIMIT 1
-- ),
--     updated_at = NOW()
-- WHERE type = 'site-work-orchestrator';
--
-- UPDATE agent_definitions
-- SET default_config = (
--     SELECT default_config
--     FROM agent_def_pageflow_builder_backup_20260505_pre_phase0_imagery
--     LIMIT 1
-- ),
--     updated_at = NOW()
-- WHERE type = 'pageflow-builder';
--
-- UPDATE agent_definitions
-- SET default_config = (
--     SELECT default_config
--     FROM agent_def_image_build_handler_backup_20260505_pre_phase0_imagery
--     LIMIT 1
-- ),
--     updated_at = NOW()
-- WHERE type = 'image-build-handler';
--
-- COMMIT;

---
-- combined migration
-- ============================================================================
-- Migration: Phase 0 imagery work — combined
--
-- Bundles three concerns that all touch the same six agent-definition rows:
--
--   1. Phase 0.1 — pass site_id through to image-generator so it can read
--      design_intent.imagery_direction from site_specs.
--      (6 jsonb_set: input_mapping additions)
--
--   2. Phase 0.2 — set origin_model literal on store_asset steps so the
--      assets table records what produced each image.
--      (6 jsonb_set: store_asset config additions)
--
--   3. Phase 0.2 follow-up — normalise origin_prompt_field on the two
--      orchestrator parents so origin_prompt records the actual composed
--      prompt sent to the model (post-Phase-0.1 prefix), not the un-composed
--      plan prompt. image-build-handler is already correct (uses
--      image_result.prompt) — only site-work-orchestrator and pageflow-builder
--      need this fix.
--      (4 jsonb_set: origin_prompt_field path changes)
--
-- One transaction. All three concerns must ship together because:
--   - Section 1's Go change reads site_id from input_mapping
--   - Section 2's Go change reads origin_model from config
--   - Section 3 ensures the prompt actually recorded reflects what was sent
--
-- Rollback at the bottom of file.
-- ============================================================================

BEGIN;

-- ----------------------------------------------------------------------------
-- SECTION 1 — site_id passthrough to image-generator
-- (Phase 0.1; 6 surgical updates across 3 agents)
-- ----------------------------------------------------------------------------

-- 1a. image-build-handler.call_logo_gen
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_gen,config,input_mapping,site_id}',
        '"site_record.site_id"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'image-build-handler';

-- 1b. image-build-handler.call_hero_gen
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_hero_gen,config,input_mapping,site_id}',
        '"site_record.site_id"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'image-build-handler';

-- 1c. site-work-orchestrator.generate_hero_image
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_hero_image,config,input_mapping,site_id}',
        '"site_record.site_id"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'site-work-orchestrator';

-- 1d. site-work-orchestrator.call_logo_generation
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_generation,config,input_mapping,site_id}',
        '"site_record.site_id"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'site-work-orchestrator';

-- 1e. pageflow-builder.generate_hero_image
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,generate_hero_image,config,input_mapping,site_id}',
        '"site_record.site_id"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'pageflow-builder';

-- 1f. pageflow-builder.call_logo_generation
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_generation,config,input_mapping,site_id}',
        '"site_record.site_id"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'pageflow-builder';


-- ----------------------------------------------------------------------------
-- SECTION 2 — origin_model literal on store_asset configs
-- (Phase 0.2; 6 surgical updates across 3 agents)
--
-- Today's only generation backend is Stability hosted SDXL. When provider
-- routing lands (PLAN Phase 4), these literals get replaced with
-- origin_model_field paths that point to the dynamic model in the response.
-- ----------------------------------------------------------------------------

-- 2a. image-build-handler.store_hero_asset
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_hero_asset,config,origin_model}',
        '"sdxl"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'image-build-handler';

-- 2b. image-build-handler.store_logo_asset
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_logo_asset,config,origin_model}',
        '"sdxl"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'image-build-handler';

-- 2c. site-work-orchestrator.store_hero_asset
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_hero_asset,config,origin_model}',
        '"sdxl"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'site-work-orchestrator';

-- 2d. site-work-orchestrator.store_logo_asset
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_logo_asset,config,origin_model}',
        '"sdxl"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'site-work-orchestrator';

-- 2e. pageflow-builder.store_hero_asset
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_hero_asset,config,origin_model}',
        '"sdxl"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'pageflow-builder';

-- 2f. pageflow-builder.store_logo_asset
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_logo_asset,config,origin_model}',
        '"sdxl"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'pageflow-builder';


-- ----------------------------------------------------------------------------
-- SECTION 3 — normalise origin_prompt_field to capture the composed prompt
-- (Phase 0.2 follow-up; 4 surgical updates across 2 agents)
--
-- Today: site-work-orchestrator and pageflow-builder set
--   origin_prompt_field = "site_plan.image_prompts.hero_home" (or .logo)
-- This records what the planner asked for, NOT what was actually sent
-- to the model after Phase 0.1's imagery_direction prefix is composed in.
--
-- After this migration: origin_prompt records what the generator returned,
-- which is the composed prompt the model actually saw. Better for audit,
-- better for future iterations of the imagery audit work.
--
-- image-build-handler is already correct (uses image_result.prompt) so
-- no change needed there.
--
-- The output_mapping on these workflows already populates {hero,logo}_result.prompt
-- via "prompt": "generate.response.prompt", so the path is valid.
-- ----------------------------------------------------------------------------

-- 3a. site-work-orchestrator.store_hero_asset.origin_prompt_field
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_hero_asset,config,origin_prompt_field}',
        '"hero_result.prompt"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'site-work-orchestrator';

-- 3b. site-work-orchestrator.store_logo_asset.origin_prompt_field
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_logo_asset,config,origin_prompt_field}',
        '"logo_result.prompt"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'site-work-orchestrator';

-- 3c. pageflow-builder.store_hero_asset.origin_prompt_field
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_hero_asset,config,origin_prompt_field}',
        '"hero_result.prompt"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'pageflow-builder';

-- 3d. pageflow-builder.store_logo_asset.origin_prompt_field
UPDATE agent_definitions SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_logo_asset,config,origin_prompt_field}',
        '"logo_result.prompt"'::jsonb,
        true
                                              ), updated_at = NOW()
WHERE type = 'pageflow-builder';


-- ----------------------------------------------------------------------------
-- VERIFICATION (run after commit; expects 16 rows, all populated)
-- ----------------------------------------------------------------------------
-- WITH paths AS (
--     SELECT 'image-build-handler' AS t, '{workflow,steps,call_logo_gen,config,input_mapping,site_id}'::text[] AS p, 's1' AS section
--     UNION ALL SELECT 'image-build-handler', '{workflow,steps,call_hero_gen,config,input_mapping,site_id}'::text[], 's1'
--     UNION ALL SELECT 'site-work-orchestrator', '{workflow,steps,generate_hero_image,config,input_mapping,site_id}'::text[], 's1'
--     UNION ALL SELECT 'site-work-orchestrator', '{workflow,steps,call_logo_generation,config,input_mapping,site_id}'::text[], 's1'
--     UNION ALL SELECT 'pageflow-builder', '{workflow,steps,generate_hero_image,config,input_mapping,site_id}'::text[], 's1'
--     UNION ALL SELECT 'pageflow-builder', '{workflow,steps,call_logo_generation,config,input_mapping,site_id}'::text[], 's1'
--     UNION ALL SELECT 'image-build-handler', '{workflow,steps,store_hero_asset,config,origin_model}'::text[], 's2'
--     UNION ALL SELECT 'image-build-handler', '{workflow,steps,store_logo_asset,config,origin_model}'::text[], 's2'
--     UNION ALL SELECT 'site-work-orchestrator', '{workflow,steps,store_hero_asset,config,origin_model}'::text[], 's2'
--     UNION ALL SELECT 'site-work-orchestrator', '{workflow,steps,store_logo_asset,config,origin_model}'::text[], 's2'
--     UNION ALL SELECT 'pageflow-builder', '{workflow,steps,store_hero_asset,config,origin_model}'::text[], 's2'
--     UNION ALL SELECT 'pageflow-builder', '{workflow,steps,store_logo_asset,config,origin_model}'::text[], 's2'
--     UNION ALL SELECT 'site-work-orchestrator', '{workflow,steps,store_hero_asset,config,origin_prompt_field}'::text[], 's3'
--     UNION ALL SELECT 'site-work-orchestrator', '{workflow,steps,store_logo_asset,config,origin_prompt_field}'::text[], 's3'
--     UNION ALL SELECT 'pageflow-builder', '{workflow,steps,store_hero_asset,config,origin_prompt_field}'::text[], 's3'
--     UNION ALL SELECT 'pageflow-builder', '{workflow,steps,store_logo_asset,config,origin_prompt_field}'::text[], 's3'
-- )
-- SELECT paths.section, paths.t AS agent_type, paths.p AS path,
--        ad.default_config #>> paths.p AS value
--   FROM paths
--   JOIN agent_definitions ad ON ad.type = paths.t
-- ORDER BY paths.section, paths.t, paths.p;

COMMIT;


-- ============================================================================
-- ROLLBACK (run only if reverting)
-- ============================================================================
-- BEGIN;
--
-- -- Section 1 rollback: remove site_id from image-generator input_mappings
-- UPDATE agent_definitions SET default_config =
--     default_config #- '{workflow,steps,call_logo_gen,config,input_mapping,site_id}'
--                    #- '{workflow,steps,call_hero_gen,config,input_mapping,site_id}',
--     updated_at = NOW()
--   WHERE type = 'image-build-handler';
-- UPDATE agent_definitions SET default_config =
--     default_config #- '{workflow,steps,generate_hero_image,config,input_mapping,site_id}'
--                    #- '{workflow,steps,call_logo_generation,config,input_mapping,site_id}',
--     updated_at = NOW()
--   WHERE type = 'site-work-orchestrator';
-- UPDATE agent_definitions SET default_config =
--     default_config #- '{workflow,steps,generate_hero_image,config,input_mapping,site_id}'
--                    #- '{workflow,steps,call_logo_generation,config,input_mapping,site_id}',
--     updated_at = NOW()
--   WHERE type = 'pageflow-builder';
--
-- -- Section 2 rollback: remove origin_model literals
-- UPDATE agent_definitions SET default_config =
--     default_config #- '{workflow,steps,store_hero_asset,config,origin_model}'
--                    #- '{workflow,steps,store_logo_asset,config,origin_model}',
--     updated_at = NOW()
--   WHERE type IN ('image-build-handler', 'site-work-orchestrator', 'pageflow-builder');
--
-- -- Section 3 rollback: restore prior origin_prompt_field values
-- UPDATE agent_definitions SET default_config = jsonb_set(
--     default_config,
--     '{workflow,steps,store_hero_asset,config,origin_prompt_field}',
--     '"site_plan.image_prompts.hero_home"'::jsonb
-- ), updated_at = NOW()
--   WHERE type IN ('site-work-orchestrator', 'pageflow-builder');
-- UPDATE agent_definitions SET default_config = jsonb_set(
--     default_config,
--     '{workflow,steps,store_logo_asset,config,origin_prompt_field}',
--     '"site_plan.image_prompts.logo"'::jsonb
-- ), updated_at = NOW()
--   WHERE type IN ('site-work-orchestrator', 'pageflow-builder');
--
-- COMMIT;

---

-- ============================================================================
-- Phase 1.5 hotfix — add output_mapping to image-build-handler image-generator
-- calls so store_asset can read image_result.image_url correctly.
--
-- Problem found in production verification (2026-05-07):
--   image-build-handler.call_logo_gen and call_hero_gen have no output_mapping
--   on their call_agent step. The image-generator response gets stored under
--   image_result wholesale, deeply nested as
--   image_result.response.generate.response.image_url.
--   store_asset's data_field reads image_result.image_url and finds nothing,
--   so it returns {stored: false, reason: "no asset URL found"}.
--   Result: image-generator runs, uploads to S3, returns the URL, but no
--   asset row is ever created.
--
-- Pattern: identical output_mapping to what pageflow-builder.generate_hero_image
-- and pageflow-builder.call_logo_generation already use.
--
-- Idempotent: uses jsonb_set with create_missing=true. Re-running has no
-- additional effect since the same value is set.
-- ============================================================================

BEGIN;

-- Add output_mapping to call_hero_gen
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_hero_gen,config,output_mapping}',
        '{"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}'::jsonb,
        true
                     ),
    updated_at = NOW()
WHERE type = 'image-build-handler';

-- Add output_mapping to call_logo_gen
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_logo_gen,config,output_mapping}',
        '{"prompt": "generate.response.prompt", "image_uri": "generate.response.image_uri", "image_url": "generate.response.image_url", "generated_at": "generate.response.generated_at"}'::jsonb,
        true
                     ),
    updated_at = NOW()
WHERE type = 'image-build-handler';


-- ----------------------------------------------------------------------------
-- Verification — both steps should now have output_mapping with the four keys.
-- ----------------------------------------------------------------------------
SELECT
    type,
    jsonb_pretty(default_config #> '{workflow,steps,call_hero_gen,config,output_mapping}') AS hero_output_mapping,
    jsonb_pretty(default_config #> '{workflow,steps,call_logo_gen,config,output_mapping}') AS logo_output_mapping
FROM agent_definitions
WHERE type = 'image-build-handler';

COMMIT;


-- ============================================================================
-- ROLLBACK — removes the output_mapping fields, restoring prior shape.
-- ============================================================================
-- BEGIN;
-- UPDATE agent_definitions
-- SET default_config = default_config
--     #- '{workflow,steps,call_hero_gen,config,output_mapping}'
--     #- '{workflow,steps,call_logo_gen,config,output_mapping}',
--     updated_at = NOW()
-- WHERE type = 'image-build-handler';
-- COMMIT;

---

-- ============================================================================
-- Phase 2E — add variant path to image-build-handler workflow
--
-- Existing workflow has two branches under check_item_type:
--   * spawn_image_gen     → call_logo_gen     → store_logo_asset → deploy_logo
--   * spawn_image_gen_hero → call_hero_gen    → store_hero_asset → deploy_hero
--
-- Phase 2E adds a third branch for unfulfilled_hero_variant work items:
--   * spawn_image_gen_variant → call_variant_gen → store_variant_asset → deploy_variant
--
-- The variant branch reads input_data.spec.prompt directly (avoiding workflow
-- dynamic-key lookup), passes asset_key to store_asset (Phase 2C-aware) and
-- deploy_image_asset (Phase 2E action change), and uses purpose=hero for
-- consistent dimensions/audit grouping.
--
-- Existing logo and hero paths are NOT modified — backward compatibility
-- preserved for any in-flight work items still using the canonical spec
-- format.
--
-- Order of operations:
--   1. Update check_item_type's condition to route variants to the new branch.
--   2. Add four new step entries: spawn_image_gen_variant, call_variant_gen,
--      store_variant_asset, deploy_variant.
--
-- Both updates use jsonb_set on default_config; idempotent across re-runs.
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ----------------------------------------------------------------------------
-- Step 1: Update check_item_type to add the variant branch.
-- ----------------------------------------------------------------------------
-- Current condition: "input_data.item_type == 'needs_logo' OR input_data.purpose == 'logo'"
--   then_step: spawn_image_gen        (logo path)
--   else_step: spawn_image_gen_hero   (hero path)
--
-- Phase 2E: replace check_item_type with a chain of two conditionals so we
-- get three branches without losing the existing routing.
--
-- The new structure:
--   check_item_type → conditional checking variant first
--                  ├─ true  → spawn_image_gen_variant
--                  └─ false → check_logo_or_hero (existing logical split)
--                              ├─ true (logo) → spawn_image_gen
--                              └─ false (hero) → spawn_image_gen_hero

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_item_type}',
        '{
            "action": "conditional",
            "config": {
                "condition": "input_data.item_type == ''unfulfilled_hero_variant''",
                "then_step": "spawn_image_gen_variant",
                "else_step": "check_logo_or_hero"
            },
            "description": "Route to variant, logo, or hero generation path"
        }'::jsonb
                     )
WHERE type = 'image-build-handler';


-- Step 1b: Add the secondary conditional that preserves the original split.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_logo_or_hero}',
        '{
            "action": "conditional",
            "config": {
                "condition": "input_data.item_type == ''needs_logo'' OR input_data.purpose == ''logo''",
                "then_step": "spawn_image_gen",
                "else_step": "spawn_image_gen_hero"
            },
            "description": "Existing logo-or-hero split preserved from pre-Phase-2E"
        }'::jsonb
                     )
WHERE type = 'image-build-handler';


-- ----------------------------------------------------------------------------
-- Step 2: Add the four variant-path steps.
-- ----------------------------------------------------------------------------

-- 2a. spawn_image_gen_variant
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_image_gen_variant}',
        '{
            "action": "spawn_agent",
            "config": {
                "role": "image_generator",
                "agent_type": "image-generator"
            },
            "next_step": "call_variant_gen",
            "description": "Spawn image generator for a hero variant",
            "output_field": "image_gen_agent"
        }'::jsonb
                     )
WHERE type = 'image-build-handler';


-- 2b. call_variant_gen — reads input_data.spec.prompt directly (no dynamic
-- key lookup needed since the Phase 2E spec format includes the prompt at
-- a known path).
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_variant_gen}',
        '{
            "action": "call_agent",
            "config": {
                "agent_type": "image-generator",
                "target_role": "image_generator",
                "input_mapping": {
                    "prompt": "input_data.spec.prompt",
                    "site_plan": "input_data.spec"
                },
                "output_mapping": {
                    "prompt": "generate.response.prompt",
                    "image_uri": "generate.response.image_uri",
                    "image_url": "generate.response.image_url",
                    "generated_at": "generate.response.generated_at"
                },
                "timeout_seconds": 120
            },
            "next_step": "store_variant_asset",
            "error_step": "complete_error",
            "description": "Generate hero variant image",
            "output_field": "image_result"
        }'::jsonb
                     )
WHERE type = 'image-build-handler';


-- 2c. store_variant_asset — uses literal purpose="hero" (all variants are hero
-- variants today) and asset_key_field for path-based lookup of the variant
-- identifier from input_data.spec.asset_key. Phase 2E adds asset_key_field
-- support to StoreAssetAction.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,store_variant_asset}',
        '{
            "action": "store_asset",
            "config": {
                "asset_type": "image",
                "purpose": "hero",
                "asset_key_field": "input_data.spec.asset_key",
                "data_field": "image_result.image_url",
                "origin_type": "generated",
                "origin_model": "sdxl",
                "site_id_field": "site_record.site_id",
                "origin_prompt_field": "image_result.prompt",
                "update_site_brand_assets": false
            },
            "next_step": "deploy_variant",
            "error_step": "deploy_variant",
            "description": "Store variant in assets table. Error step still deploys.",
            "output_field": "asset_stored"
        }'::jsonb
                     )
WHERE type = 'image-build-handler';


-- 2d. deploy_variant — uses literal purpose="hero" and asset_key_field for
-- per-variant filename derivation (Phase 2E action change).
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_variant}',
        '{
            "action": "deploy_image_asset",
            "config": {
                "purpose": "hero",
                "domain_field": "site_record.domain",
                "uri_field": "asset_stored.image_uri",
                "asset_key_field": "input_data.spec.asset_key"
            },
            "next_step": "complete",
            "error_step": "complete_error",
            "description": "Download variant from S3, optimize, git commit. Per-variant deploy path derived from asset_key (e.g. assets/images/hero-about.jpg).",
            "output_field": "deploy_result"
        }'::jsonb
                     )
WHERE type = 'image-build-handler';


-- ----------------------------------------------------------------------------
-- Verification — confirm all steps exist and the routing reads correctly
-- ----------------------------------------------------------------------------
SELECT type, version,
       jsonb_pretty(default_config #> '{workflow,steps,check_item_type}')         AS check_item_type,
       jsonb_pretty(default_config #> '{workflow,steps,check_logo_or_hero}')      AS check_logo_or_hero
FROM agent_definitions
WHERE type = 'image-build-handler';

SELECT 'variant steps present' AS check,
       (default_config #> '{workflow,steps,spawn_image_gen_variant}') IS NOT NULL AS spawn,
       (default_config #> '{workflow,steps,call_variant_gen}')        IS NOT NULL AS call_gen,
       (default_config #> '{workflow,steps,store_variant_asset}')     IS NOT NULL AS store,
       (default_config #> '{workflow,steps,deploy_variant}')          IS NOT NULL AS deploy
FROM agent_definitions
WHERE type = 'image-build-handler';

SELECT 'existing steps preserved' AS check,
       (default_config #> '{workflow,steps,spawn_image_gen}')      IS NOT NULL AS logo_spawn,
       (default_config #> '{workflow,steps,call_logo_gen}')        IS NOT NULL AS logo_call,
       (default_config #> '{workflow,steps,store_logo_asset}')     IS NOT NULL AS logo_store,
       (default_config #> '{workflow,steps,deploy_logo}')          IS NOT NULL AS logo_deploy,
       (default_config #> '{workflow,steps,spawn_image_gen_hero}') IS NOT NULL AS hero_spawn,
       (default_config #> '{workflow,steps,call_hero_gen}')        IS NOT NULL AS hero_call,
       (default_config #> '{workflow,steps,store_hero_asset}')     IS NOT NULL AS hero_store,
       (default_config #> '{workflow,steps,deploy_hero}')          IS NOT NULL AS hero_deploy
FROM agent_definitions
WHERE type = 'image-build-handler';

COMMIT;


-- ============================================================================
-- ROLLBACK
-- ============================================================================
-- Restore from backup table (see phase_2e_pre_migration_backup.sql) or
-- explicitly remove the variant steps and restore the old check_item_type:
--
-- BEGIN;
-- UPDATE agent_definitions
-- SET default_config = default_config
--     #- '{workflow,steps,check_logo_or_hero}'
--     #- '{workflow,steps,spawn_image_gen_variant}'
--     #- '{workflow,steps,call_variant_gen}'
--     #- '{workflow,steps,store_variant_asset}'
--     #- '{workflow,steps,deploy_variant}';
--
-- UPDATE agent_definitions
-- SET default_config = jsonb_set(default_config, '{workflow,steps,check_item_type}',
--   '{
--       "action": "conditional",
--       "config": {
--           "condition": "input_data.item_type == ''needs_logo'' OR input_data.purpose == ''logo''",
--           "else_step": "spawn_image_gen_hero",
--           "then_step": "spawn_image_gen"
--       },
--       "description": "Route to logo or hero generation path"
--   }'::jsonb)
-- WHERE type = 'image-build-handler';
-- COMMIT;

----
-- fix image s3client paths
-- ============================================================================
-- Phase 2F: image-build-handler → spawn asset-deployer for deploy steps
-- ============================================================================
-- Background
-- ----------
-- image-build-handler currently runs `deploy_image_asset` inline in three
-- steps: deploy_logo, deploy_hero, deploy_variant. Inline execution happens
-- inside the chassis pod, which by design does NOT carry storage env vars
-- (different operations write to different buckets and may run at different
-- times). The result is "storage client not available" failures.
--
-- spawn_actions.go injects storage env vars (S3_ENDPOINT, IMAGE_BUCKET,
-- ASSETS_BUCKET, B2_*, AWS_*) into spawned children whose agent_type is in
-- isStorageEnabledAgent OR whose category is orchestrator/code-driven.
-- asset-deployer is already in isStorageEnabledAgent (id e9a9bac9-…), with a
-- workflow that runs deploy_image_asset against those credentials.
--
-- This migration replaces the three inline deploys with a single shared
-- spawn_asset_deployer + call_asset_deployer pair. All three branches
-- (logo / hero / variant) converge here because the input shape is uniform:
--   - domain     ← site_record.domain
--   - s3_uri     ← asset_stored.image_uri
--   - purpose    ← asset_stored.purpose
--   - asset_key  ← input_data.spec.asset_key  (optional, ? on missing)
--
-- Two coordinated changes — applied as a single transaction with snapshots
-- taken first so revert_agent() can roll either agent back independently if
-- something downstream surfaces.
--
-- Affected agents:
--   asset-deployer       e9a9bac9-dfe4-4aca-8f32-19738ac265c6
--   image-build-handler  04b10d94-11ee-447c-9ff9-7924b8e9897c
-- ============================================================================

\set ON_ERROR_STOP on

BEGIN;

-- ── 1. Snapshots ────────────────────────────────────────────────────────────
-- snapshot_agent() inserts an is_snapshot=true copy of the current row,
-- pointing previous_version_id back at the active row. revert_agent() reads
-- the latest snapshot and restores default_config from it.

SELECT snapshot_agent('asset-deployer')      AS asset_deployer_snapshot_id;
SELECT snapshot_agent('image-build-handler') AS image_build_handler_snapshot_id;


-- ── 2. asset-deployer: accept asset_key ─────────────────────────────────────
-- 2a. Declare asset_key as an optional input on the agent's input_contract.

UPDATE agent_definitions
SET input_contract = jsonb_set(
        input_contract,
        '{optional}',
        '["deploy_path", "purpose", "asset_key"]'::jsonb
                     ),
    updated_at = NOW()
WHERE id = 'e9a9bac9-dfe4-4aca-8f32-19738ac265c6'
  AND type = 'asset-deployer'
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 2b. Make the deploy_asset step's deploy_image_asset action extract asset_key
-- from input_data. The action already has asset_key in DeployImageAssetInputSpec
-- (Optional list); this just adds it to the workflow-level input_fields
-- declaration so ExtractActionInputs picks it up.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,deploy_asset,config,input_fields}',
        '["s3_uri", "deploy_path", "purpose", "domain", "asset_key"]'::jsonb
                     ),
    updated_at = NOW()
WHERE id = 'e9a9bac9-dfe4-4aca-8f32-19738ac265c6'
  AND type = 'asset-deployer'
  AND (is_snapshot IS NULL OR is_snapshot = false);


-- ── 3. image-build-handler: replace inline deploys with spawn+call ──────────
-- 3a. Remove the three inline deploy steps.

UPDATE agent_definitions
SET default_config = (
    default_config
        #- '{workflow,steps,deploy_logo}'
        #- '{workflow,steps,deploy_hero}'
        #- '{workflow,steps,deploy_variant}'
    ),
    updated_at = NOW()
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND type = 'image-build-handler'
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 3b. Add the spawn step. role naming follows the same convention as the
-- existing spawn_image_gen / call_*_gen pairs in this same workflow
-- (target_role lookup, fixed type).

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_asset_deployer}',
        $json$
            {
            "action": "spawn_agent",
        "config": {
                "role": "asset_deployer",
        "agent_type": "asset-deployer"
    },
            "next_step": "call_asset_deployer",
            "description": "Spawn asset-deployer in its own Job pod (gets storage env vars via spawn_actions.go isStorageEnabledAgent gate).",
            "output_field": "asset_deployer_agent"
        }
        $json$::jsonb
    ),
    updated_at = NOW()
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND type = 'image-build-handler'
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 3c. Add the call step. asset_key is optional (?) — present for variants and
-- the canonical logo/hero paths via the classifier, omitted on any future
-- caller that doesn't set spec.asset_key.

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_asset_deployer}',
        $json$
            {
            "action": "call_agent",
        "config": {
                "agent_type": "asset-deployer",
        "target_role": "asset_deployer",
        "input_mapping": {
                    "domain": "site_record.domain",
        "s3_uri": "asset_stored.image_uri",
        "purpose": "asset_stored.purpose",
        "asset_key?": "input_data.spec.asset_key"
    },
                "timeout_seconds": 180
            },
            "next_step": "complete",
            "error_step": "complete_error",
            "description": "Call asset-deployer to download from S3, optimize by purpose, commit to git. asset_key drives per-variant deploy path when distinct from purpose (e.g. assets/images/hero-about.jpg).",
            "output_field": "deploy_result"
        }
        $json$::jsonb
    ),
    updated_at = NOW()
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND type = 'image-build-handler'
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 3d. Repoint the three store steps' next_step / error_step at the new spawn
-- entry. error_step matches the prior semantics: if store fails, still try to
-- deploy (the storage URI is reachable through fallbacks even when the row
-- insert errors).

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        jsonb_set(
                                jsonb_set(
                                        jsonb_set(
                                                default_config,
                                                '{workflow,steps,store_logo_asset,next_step}',
                                                '"spawn_asset_deployer"'::jsonb
                                        ),
                                        '{workflow,steps,store_logo_asset,error_step}',
                                        '"spawn_asset_deployer"'::jsonb
                                ),
                                '{workflow,steps,store_hero_asset,next_step}',
                                '"spawn_asset_deployer"'::jsonb
                        ),
                        '{workflow,steps,store_hero_asset,error_step}',
                        '"spawn_asset_deployer"'::jsonb
                ),
                '{workflow,steps,store_variant_asset,next_step}',
                '"spawn_asset_deployer"'::jsonb
        ),
        '{workflow,steps,store_variant_asset,error_step}',
        '"spawn_asset_deployer"'::jsonb
                     ),
    updated_at = NOW()
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND type = 'image-build-handler'
  AND (is_snapshot IS NULL OR is_snapshot = false);


-- ── 4. Verification ─────────────────────────────────────────────────────────
-- These selects are read-only and run before COMMIT so a failure here aborts
-- the migration.

-- 4a. asset-deployer accepts asset_key on the contract.
SELECT
    'asset-deployer.input_contract.optional' AS check_name,
    input_contract->'optional'               AS value
FROM agent_definitions
WHERE id = 'e9a9bac9-dfe4-4aca-8f32-19738ac265c6'
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 4b. asset-deployer's deploy_asset step extracts asset_key.
SELECT
    'asset-deployer.deploy_asset.input_fields' AS check_name,
    default_config->'workflow'->'steps'->'deploy_asset'->'config'->'input_fields' AS value
FROM agent_definitions
WHERE id = 'e9a9bac9-dfe4-4aca-8f32-19738ac265c6'
  AND (is_snapshot IS NULL OR is_snapshot = false);

-- 4c. The three old inline deploy steps are gone (expect 0).
SELECT
    'image-build-handler orphan deploys (expect 0)' AS check_name,
    count(*) AS count
FROM agent_definitions,
    jsonb_object_keys(default_config->'workflow'->'steps') AS k
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND k IN ('deploy_logo', 'deploy_hero', 'deploy_variant');

-- 4d. The new spawn+call pair is present (expect 2).
SELECT
    'image-build-handler new spawn+call (expect 2)' AS check_name,
    count(*) AS count
FROM agent_definitions,
    jsonb_object_keys(default_config->'workflow'->'steps') AS k
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND k IN ('spawn_asset_deployer', 'call_asset_deployer');

-- 4e. All three store_*_asset steps now route to spawn_asset_deployer.
SELECT
    'image-build-handler store routing'                           AS check_name,
    key                                                           AS step_name,
    value->>'next_step'                                           AS next_step,
    value->>'error_step'                                          AS error_step
FROM agent_definitions,
    jsonb_each(default_config->'workflow'->'steps')
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND (is_snapshot IS NULL OR is_snapshot = false)
  AND key IN ('store_logo_asset', 'store_hero_asset', 'store_variant_asset')
ORDER BY key;

-- 4f. The new call_asset_deployer's input_mapping is what we wrote.
SELECT
    'image-build-handler call_asset_deployer.input_mapping' AS check_name,
    default_config->'workflow'->'steps'->'call_asset_deployer'->'config'->'input_mapping' AS value
FROM agent_definitions
WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'
  AND (is_snapshot IS NULL OR is_snapshot = false);

COMMIT;


-- ============================================================================
-- Rollback (do NOT run as part of the migration — separate session)
-- ============================================================================
-- Each agent has its own snapshot. revert_agent restores from the latest one
-- and deletes the snapshot. Run separately if needed:
--
--     SELECT revert_agent('image-build-handler');
--     SELECT revert_agent('asset-deployer');
--
-- Order doesn't matter — they're independent rows. Note revert_agent only
-- restores default_config; the input_contract change on asset-deployer is
-- NOT covered by snapshot_agent's restore path. If you need to revert that
-- too, do it manually:
--
--     UPDATE agent_definitions
--     SET input_contract = jsonb_set(
--             input_contract,
--             '{optional}',
--             '["deploy_path", "purpose"]'::jsonb
--         ),
--         updated_at = NOW()
--     WHERE id = 'e9a9bac9-dfe4-4aca-8f32-19738ac265c6'
--       AND (is_snapshot IS NULL OR is_snapshot = false);

---
-- workflow changes

-- phase_2g_step5_image_build_handler_needs_imagery.sql
--
-- Phase 2G step 5 — teach image-build-handler to process needs_imagery
-- work items (emitted by step 4's check_unfulfilled_imagery_plan).
--
-- Approach: NEW BRANCH ALONGSIDE the existing variant chain rather than
-- extending the variant branch's matcher. Justification: the next phase
-- of imagery work (kind-specific behaviour — icon transparency, logo
-- composition rules, infographic SVG output) wants separation from the
-- variant chain's hero-only assumptions. Forking now is cheaper than
-- forking later.
--
-- Workflow additions:
--   1. check_item_type_imagery  — routes needs_imagery items to new branch
--   2. spawn_image_gen_imagery  — spawn image-generator
--   3. call_imagery_gen         — invoke generator with prompt + site_id
--                                  + kind/style_hints/constraints (the
--                                  latter three pass-through until the
--                                  cascade is enriched in a follow-up)
--   4. check_imagery_brand_update — routes on spec.brand_update boolean
--                                    computed by step 4
--   5. store_imagery_brand_asset  — store with update_site_brand_assets=true
--   6. store_imagery_asset        — store with update_site_brand_assets=false
--
-- Both store steps feed into the shared spawn_asset_deployer → call_asset_deployer
-- → complete tail, so the asset-deployer chain is unchanged.
--
-- ensure_site_record's next_step changes from check_item_type to
-- check_item_type_imagery. check_item_type_imagery's else_step is
-- check_item_type, preserving the existing routing for legacy item types.
--
-- Note on store_asset config: uses purpose_field and asset_key_field for
-- dynamic spec lookup. asset_key_field is already established (variant chain
-- uses it). purpose_field is by convention same pattern; if the store_asset
-- action doesn't recognise it, the migration verification will surface a
-- failure in the first run.
--
-- Run AFTER step 4 (check_unfulfilled_imagery_plan.go deployed) so that
-- needs_imagery items exist to be consumed.

\set ON_ERROR_STOP on

-- ── Backup ──

CREATE TABLE agent_def_image_build_handler_backup_20260513_pre_phase2g_step5 AS
SELECT * FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

SELECT
    (SELECT COUNT(*) FROM agent_definitions
     WHERE type = 'image-build-handler' AND is_active = true) AS live,
    (SELECT COUNT(*) FROM agent_def_image_build_handler_backup_20260513_pre_phase2g_step5) AS backup;

-- ── Migration ──

BEGIN;

-- Sanity: target row exists and currently has the variant branch
-- (confirms we're patching the post-Phase-2E state, not an older snapshot)
DO $check$
DECLARE
v_has_variant boolean;
BEGIN
SELECT default_config #> '{workflow,steps,check_item_type}' IS NOT NULL
       AND default_config #> '{workflow,steps,call_variant_gen}' IS NOT NULL
INTO v_has_variant
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

IF NOT v_has_variant THEN
        RAISE EXCEPTION 'image-build-handler does not have post-Phase-2E variant chain; check the agent_definition state before migrating';
END IF;
END
$check$;

-- 1. Add the six new workflow steps in one jsonb merge.
--    Merging `||` onto an existing object key adds/replaces keys, leaving
--    other siblings untouched.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps}',
        (default_config #> '{workflow,steps}') || $new_steps$
            {
    "check_item_type_imagery": {
        "action": "conditional",
        "config": {
            "condition": "input_data.item_type == 'needs_imagery'",
        "then_step": "spawn_image_gen_imagery",
        "else_step": "check_item_type"
    },
        "description": "Phase 2G step 5: route needs_imagery items to the structured imagery branch. Legacy item types fall through to check_item_type (variant/logo/hero)."
    },
    "spawn_image_gen_imagery": {
        "action": "spawn_agent",
        "config": {
            "role": "image_generator",
            "agent_type": "image-generator"
        },
        "next_step": "call_imagery_gen",
        "description": "Spawn image generator for a needs_imagery item (Phase 2G step 5)",
        "output_field": "image_gen_agent"
    },
    "call_imagery_gen": {
        "action": "call_agent",
        "config": {
            "agent_type": "image-generator",
            "target_role": "image_generator",
            "input_mapping": {
                "prompt": "input_data.spec.prompt",
                "site_id": "site_record.site_id",
                "kind": "input_data.spec.kind",
                "style_hints": "input_data.spec.style_hints",
                "constraints": "input_data.spec.constraints",
                "site_plan": "input_data.spec"
            },
            "output_mapping": {
                "prompt": "generate.response.prompt",
                "image_uri": "generate.response.image_uri",
                "image_url": "generate.response.image_url",
                "generated_at": "generate.response.generated_at"
            },
            "timeout_seconds": 120
        },
        "next_step": "check_imagery_brand_update",
        "error_step": "complete_error",
        "description": "Generate imagery (Phase 2G step 5). site_id passes so generate_image_actions can prepend design_intent.imagery_direction. kind/style_hints/constraints pass through for future cascade enrichment.",
        "output_field": "image_result"
    },
    "check_imagery_brand_update": {
        "action": "conditional",
        "config": {
            "condition": "input_data.spec.brand_update == true",
            "then_step": "store_imagery_brand_asset",
            "else_step": "store_imagery_asset"
        },
        "description": "Phase 2G step 5: route based on the brand_update flag set by the discovery check (rule b: site-scope OR canonical index hero)."
    },
    "store_imagery_brand_asset": {
        "action": "store_asset",
        "config": {
            "asset_type": "image",
            "purpose_field": "input_data.spec.purpose",
            "asset_key_field": "input_data.spec.asset_key",
            "data_field": "image_result.image_url",
            "origin_type": "generated",
            "origin_model": "sdxl",
            "site_id_field": "site_record.site_id",
            "origin_prompt_field": "image_result.prompt",
            "update_site_brand_assets": true
        },
        "next_step": "spawn_asset_deployer",
        "error_step": "spawn_asset_deployer",
        "description": "Store imagery asset and update site_brand_assets (logo or canonical index hero — rule b). Error step still deploys.",
        "output_field": "asset_stored"
    },
    "store_imagery_asset": {
        "action": "store_asset",
        "config": {
            "asset_type": "image",
            "purpose_field": "input_data.spec.purpose",
            "asset_key_field": "input_data.spec.asset_key",
            "data_field": "image_result.image_url",
            "origin_type": "generated",
            "origin_model": "sdxl",
            "site_id_field": "site_record.site_id",
            "origin_prompt_field": "image_result.prompt",
            "update_site_brand_assets": false
        },
        "next_step": "spawn_asset_deployer",
        "error_step": "spawn_asset_deployer",
        "description": "Store imagery asset without touching site_brand_assets (page-scope non-index, section-scope, non-hero kinds). Error step still deploys.",
        "output_field": "asset_stored"
    }
}
$new_steps$::jsonb
       ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- 2. Redirect ensure_site_record.next_step to the new imagery router.
--    The new router falls through to the old check_item_type for any
--    non-needs_imagery item, so legacy paths still work.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,ensure_site_record,next_step}',
        '"check_item_type_imagery"'::jsonb
                     ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- ── Verification ──

DO $verify$
DECLARE
v_steps        jsonb;
    v_next_step    text;
    v_branch_count int;
BEGIN
SELECT default_config #> '{workflow,steps}'
INTO v_steps
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

-- All six new steps must exist
FOREACH v_next_step IN ARRAY ARRAY[
        'check_item_type_imagery',
        'spawn_image_gen_imagery',
        'call_imagery_gen',
        'check_imagery_brand_update',
        'store_imagery_brand_asset',
        'store_imagery_asset'
    ]
    LOOP
        IF NOT (v_steps ? v_next_step) THEN
            RAISE EXCEPTION 'Step % not present after migration', v_next_step;
END IF;
END LOOP;

    -- Redirected entry point
SELECT v_steps #>> '{ensure_site_record,next_step}' INTO v_next_step;
IF v_next_step <> 'check_item_type_imagery' THEN
        RAISE EXCEPTION 'ensure_site_record.next_step is %, expected check_item_type_imagery', v_next_step;
END IF;

    -- Legacy branches intact
    FOREACH v_next_step IN ARRAY ARRAY[
        'check_item_type',
        'check_logo_or_hero',
        'spawn_image_gen_variant',
        'call_variant_gen',
        'store_variant_asset',
        'store_logo_asset',
        'store_hero_asset',
        'spawn_asset_deployer',
        'call_asset_deployer',
        'complete'
    ]
    LOOP
        IF NOT (v_steps ? v_next_step) THEN
            RAISE EXCEPTION 'Legacy step % missing after migration — backup may have been clobbered', v_next_step;
END IF;
END LOOP;

    -- Branch count should be: 10 existing + 6 new = 16
SELECT count(*) INTO v_branch_count FROM jsonb_object_keys(v_steps);
IF v_branch_count < 16 THEN
        RAISE EXCEPTION 'Step count after migration is %, expected at least 16', v_branch_count;
END IF;

    RAISE NOTICE 'phase_2g_step5: image-build-handler extended; total workflow steps now %', v_branch_count;
END
$verify$;

COMMIT;