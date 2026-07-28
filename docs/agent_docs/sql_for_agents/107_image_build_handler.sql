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

---
-- a kind is another word for type e.g. hero, logo

-- phase_2h_step4_legacy_kind_defaults.sql
--
-- Phase 2H.4 — opt the legacy image-build-handler workflow steps into
-- per-kind generation defaults from generate_image_actions.go's
-- kindDefaults map.
--
-- generate_image_actions resolves `kind` from inputData["kind"] first,
-- then from params.StepConfig.Config["default_kind"]. The Phase 2G step 5
-- branch (call_imagery_gen) already passes kind via input_mapping. The
-- three legacy callers don't know about kind, so we add a default_kind
-- to their step config:
--
--   call_logo_gen      → default_kind = "logo"
--   call_hero_gen      → default_kind = "hero"
--   call_variant_gen   → default_kind = "hero"
--
-- After this migration, all four callers benefit from kind-aware defaults
-- (negative_prompt, cfg_scale, steps). The most visible win is logo
-- generation no longer producing human figures.
--
-- Idempotent. Backup per doc 009 convention.

\set ON_ERROR_STOP on

-- ── Backup ──

CREATE TABLE agent_def_image_build_handler_backup_20260513_pre_phase2h_kind_defaults AS
SELECT * FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

SELECT
    (SELECT COUNT(*) FROM agent_definitions
     WHERE type = 'image-build-handler' AND is_active = true) AS live,
    (SELECT COUNT(*) FROM agent_def_image_build_handler_backup_20260513_pre_phase2h_kind_defaults) AS backup;

-- ── Migration ──

BEGIN;

-- Sanity: confirm the three target steps exist
DO $check$
BEGIN
    IF NOT (
        EXISTS (SELECT 1 FROM agent_definitions
                 WHERE type = 'image-build-handler' AND is_active = true
                   AND default_config #> '{workflow,steps,call_logo_gen}' IS NOT NULL)
        AND
        EXISTS (SELECT 1 FROM agent_definitions
                 WHERE type = 'image-build-handler' AND is_active = true
                   AND default_config #> '{workflow,steps,call_hero_gen}' IS NOT NULL)
        AND
        EXISTS (SELECT 1 FROM agent_definitions
                 WHERE type = 'image-build-handler' AND is_active = true
                   AND default_config #> '{workflow,steps,call_variant_gen}' IS NOT NULL)
    ) THEN
        RAISE EXCEPTION 'one or more of call_logo_gen/call_hero_gen/call_variant_gen missing — wrong agent state for 2H.4';
END IF;
END
$check$;

-- Three jsonb_set chained — set default_kind on each step's config.
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                jsonb_set(
                        default_config,
                        '{workflow,steps,call_logo_gen,config,default_kind}',
                        '"logo"'::jsonb
                ),
                '{workflow,steps,call_hero_gen,config,default_kind}',
                '"hero"'::jsonb
        ),
        '{workflow,steps,call_variant_gen,config,default_kind}',
        '"hero"'::jsonb
                     ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- Verify
DO $verify$
DECLARE
v_logo_kind    text;
    v_hero_kind    text;
    v_variant_kind text;
BEGIN
SELECT default_config #>> '{workflow,steps,call_logo_gen,config,default_kind}',
           default_config #>> '{workflow,steps,call_hero_gen,config,default_kind}',
           default_config #>> '{workflow,steps,call_variant_gen,config,default_kind}'
INTO v_logo_kind, v_hero_kind, v_variant_kind
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

IF v_logo_kind <> 'logo' THEN
        RAISE EXCEPTION 'call_logo_gen.default_kind is %, expected logo', v_logo_kind;
END IF;
    IF v_hero_kind <> 'hero' THEN
        RAISE EXCEPTION 'call_hero_gen.default_kind is %, expected hero', v_hero_kind;
END IF;
    IF v_variant_kind <> 'hero' THEN
        RAISE EXCEPTION 'call_variant_gen.default_kind is %, expected hero', v_variant_kind;
END IF;

    RAISE NOTICE 'phase_2h.4: default_kind set on legacy callers (logo, hero, hero)';
END
$verify$;

COMMIT;

---
-- ? in paths optional paths

-- phase_2g_step5_hotfix_optional_input_mapping.sql
--
-- Hotfix on top of phase_2g_step5: call_imagery_gen's input_mapping has
-- `constraints`, `style_hints`, and `kind` as REQUIRED, but step 4's
-- discovery check only emits these in the spec when the corresponding
-- site_plan_imagery columns are non-null. Most imagery rows have null
-- for both, so most specs are missing these fields, so call_imagery_gen
-- fails immediately on input resolution.
--
-- Verified by orchestration e98deca7-c9df-438a-b256-be1cd15579ec failing
-- at call_imagery_gen with error:
--   "input_mapping failed: source path 'input_data.spec.constraints'
--    not found for field 'constraints'"
--
-- Fix: rename the three fields in input_mapping with `?` suffix per the
-- chassis's established optional-field convention (variant chain uses
-- this pattern: "asset_key?" → "input_data.spec.asset_key").
--
-- Reversible. Backup taken per doc 009.

\set ON_ERROR_STOP on

CREATE TABLE agent_def_image_build_handler_backup_20260514_pre_phase2g_optional_mapping AS
SELECT * FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

SELECT
    (SELECT COUNT(*) FROM agent_definitions
     WHERE type = 'image-build-handler' AND is_active = true) AS live,
    (SELECT COUNT(*) FROM agent_def_image_build_handler_backup_20260514_pre_phase2g_optional_mapping) AS backup;

BEGIN;

-- Take the current input_mapping, drop the three required keys, add
-- them back with `?` suffix pointing to the same source paths.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_imagery_gen,config,input_mapping}',
        (
            (default_config #> '{workflow,steps,call_imagery_gen,config,input_mapping}')
                - 'kind' - 'style_hints' - 'constraints'
            ) || jsonb_build_object(
                'kind?',        'input_data.spec.kind',
                'style_hints?', 'input_data.spec.style_hints',
                'constraints?', 'input_data.spec.constraints'
                 ),
        false
                     ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- Verify
DO $verify$
DECLARE
v_mapping jsonb;
BEGIN
SELECT default_config #> '{workflow,steps,call_imagery_gen,config,input_mapping}'
INTO v_mapping
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

IF NOT (v_mapping ? 'kind?') THEN
        RAISE EXCEPTION 'kind? not present after migration';
END IF;
    IF NOT (v_mapping ? 'style_hints?') THEN
        RAISE EXCEPTION 'style_hints? not present';
END IF;
    IF NOT (v_mapping ? 'constraints?') THEN
        RAISE EXCEPTION 'constraints? not present';
END IF;
    IF v_mapping ? 'constraints' THEN
        RAISE EXCEPTION 'old required constraints key still present';
END IF;

    -- Required fields still required
    IF NOT (v_mapping ? 'prompt') THEN
        RAISE EXCEPTION 'prompt missing — migration corrupted the mapping';
END IF;
    IF NOT (v_mapping ? 'site_id') THEN
        RAISE EXCEPTION 'site_id missing — migration corrupted the mapping';
END IF;

    RAISE NOTICE 'call_imagery_gen input_mapping now: %', v_mapping;
END
$verify$;

COMMIT;

---
--
-- fix paths

-- phase_2g_step5_hotfix_store_asset_purpose.sql
--
-- Second hotfix on top of phase_2g_step5. The store_asset action does NOT
-- support a `purpose_field` config key — only the `*_field` keys actually
-- attested in the existing chains (asset_key_field, site_id_field,
-- data_field, origin_prompt_field). The migration assumed by analogy
-- that purpose_field would work too. It doesn't: store_asset can't resolve
-- purpose, writes a DB row with empty purpose, and the asset_stored
-- output mapping silently drops image_uri. Downstream call_asset_deployer
-- then fails reading asset_stored.image_uri.
--
-- Verified by orchestration 076b6f19-11ed-45d1-b250-4703e8f4aa2b:
--  - SDXL call succeeded (child orchestration 222edd25 COMPLETED at 10:10:58)
--  - assets row created with purpose=NULL (00caf435-e639-4e16-8e68-2a2f27cc12a3)
--  - parent orchestration FAILED at call_asset_deployer with input_mapping
--    error: 'asset_stored.image_uri' not found for field 's3_uri'
--
-- Fix: drop purpose_field, hardcode purpose: "hero" in both new store
-- steps. Mirrors store_variant_asset exactly except for update_site_brand_assets.
--
-- Limitation: kind=logo imagery work items (if any exist for this site)
-- will be stored with purpose="hero" — incorrect. For robot-hands.com,
-- the legacy logo asset (asset_key=logo, from 2026-05-08) already covers
-- the site_plan_imagery logo row, so no needs_imagery items of kind=logo
-- are pending. Other sites may need either the Go-side purpose_field fix
-- (proper) or kind-routing in the workflow (verbose) before they can use
-- this branch for logos.
--
-- Reversible. Backup taken.

\set ON_ERROR_STOP on

CREATE TABLE agent_def_image_build_handler_backup_20260514_pre_store_purpose_hotfix AS
SELECT * FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

BEGIN;

-- Drop purpose_field and add hardcoded purpose: "hero" in both store steps.
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,store_imagery_brand_asset,config}',
                (
                    (default_config #> '{workflow,steps,store_imagery_brand_asset,config}')
                        - 'purpose_field'
                    ) || jsonb_build_object('purpose', 'hero')
        ),
        '{workflow,steps,store_imagery_asset,config}',
        (
            (default_config #> '{workflow,steps,store_imagery_asset,config}')
                - 'purpose_field'
            ) || jsonb_build_object('purpose', 'hero')
                     ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- Verify
DO $verify$
DECLARE
v_brand_cfg  jsonb;
    v_plain_cfg  jsonb;
BEGIN
SELECT default_config #> '{workflow,steps,store_imagery_brand_asset,config}',
           default_config #> '{workflow,steps,store_imagery_asset,config}'
INTO v_brand_cfg, v_plain_cfg
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

IF v_brand_cfg ? 'purpose_field' THEN
        RAISE EXCEPTION 'store_imagery_brand_asset still has purpose_field';
END IF;
    IF v_plain_cfg ? 'purpose_field' THEN
        RAISE EXCEPTION 'store_imagery_asset still has purpose_field';
END IF;
    IF (v_brand_cfg ->> 'purpose') <> 'hero' THEN
        RAISE EXCEPTION 'store_imagery_brand_asset purpose is %, expected hero', (v_brand_cfg ->> 'purpose');
END IF;
    IF (v_plain_cfg ->> 'purpose') <> 'hero' THEN
        RAISE EXCEPTION 'store_imagery_asset purpose is %, expected hero', (v_plain_cfg ->> 'purpose');
END IF;
    -- update_site_brand_assets should still be true/false respectively
    IF (v_brand_cfg ->> 'update_site_brand_assets') <> 'true' THEN
        RAISE EXCEPTION 'store_imagery_brand_asset.update_site_brand_assets corrupted';
END IF;
    IF (v_plain_cfg ->> 'update_site_brand_assets') <> 'false' THEN
        RAISE EXCEPTION 'store_imagery_asset.update_site_brand_assets corrupted';
END IF;

    RAISE NOTICE 'hotfix applied: both store steps now use purpose=hero';
END
$verify$;

COMMIT;

---
--
-- phase_2g_followup_mark_work_item_complete.sql
--
-- Follow-up on Phase 2G step 5. The image-build-handler workflow completes
-- successfully and deploys an asset, but does not update its triggering
-- work item's status. Items sit in `detected` indefinitely after
-- successful processing — observable in robot-hands.com after 2026-05-14
-- end-to-end verification: 8 needs_imagery items, 7 still detected after
-- one was processed end-to-end.
--
-- Fix: insert a new `mark_work_item_complete` step between
-- `call_asset_deployer` and `complete`, using the new UpdateWorkItemStatus
-- action (added to v3_site_actions.go and registered as
-- "update_work_item_status" — see update_work_item_status_action.go).
--
-- The step is inserted in the SHARED tail of the workflow, so all four
-- branches (needs_imagery, variant, logo, hero) benefit from a single
-- insertion.
--
-- The action gracefully no-ops when input_data.work_item_id is absent,
-- so manual triggers without a work_item_id (e.g. ad-hoc kcat calls
-- bypassing dispatch) continue to work without error.
--
-- Reversible. Backup taken.
--
-- Prerequisite: UpdateWorkItemStatusAction Go code deployed to chassis.
-- Without that, this step would attempt to invoke an unknown action and
-- fail with action-not-registered.

\set ON_ERROR_STOP on

CREATE TABLE agent_def_image_build_handler_backup_20260514_pre_mark_work_item_complete AS
SELECT * FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

BEGIN;

-- Sanity: confirm the agent has the post-step-5 structure
DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'image-build-handler' AND is_active = true
           AND default_config #> '{workflow,steps,call_asset_deployer}' IS NOT NULL
           AND default_config #>> '{workflow,steps,call_asset_deployer,next_step}' = 'complete'
    ) THEN
        RAISE EXCEPTION 'image-build-handler does not have expected post-step-5 structure; call_asset_deployer.next_step is not "complete"';
END IF;
END
$check$;

-- 1. Add the new mark_work_item_complete step.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,mark_work_item_complete}',
        $step$
            {
    "action": "update_work_item_status",
        "config": {
        "work_item_id_field": "input_data.work_item_id",
        "status": "complete",
        "skip_if_missing": true
    },
    "next_step": "complete",
    "error_step": "complete",
    "description": "Mark the triggering site_work_items row as complete. Gracefully no-ops if input_data.work_item_id is absent. error_step is `complete` (not complete_error) because failure to update the work item should not fail the asset workflow — the asset is already deployed.",
    "output_field": "work_item_marked_complete"
}
$step$::jsonb
       ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- 2. Redirect call_asset_deployer.next_step from "complete" to
--    "mark_work_item_complete". This is in the shared tail, so all four
--    branches (needs_imagery, variant, logo, hero) benefit at once.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_asset_deployer,next_step}',
        '"mark_work_item_complete"'::jsonb
                     ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- Verify
DO $verify$
DECLARE
v_steps     jsonb;
    v_next_step text;
BEGIN
SELECT default_config #> '{workflow,steps}'
INTO v_steps
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

IF NOT (v_steps ? 'mark_work_item_complete') THEN
        RAISE EXCEPTION 'mark_work_item_complete step missing after migration';
END IF;

SELECT v_steps #>> '{call_asset_deployer,next_step}' INTO v_next_step;
IF v_next_step <> 'mark_work_item_complete' THEN
        RAISE EXCEPTION 'call_asset_deployer.next_step is %, expected mark_work_item_complete', v_next_step;
END IF;

SELECT v_steps #>> '{mark_work_item_complete,next_step}' INTO v_next_step;
IF v_next_step <> 'complete' THEN
        RAISE EXCEPTION 'mark_work_item_complete.next_step is %, expected complete', v_next_step;
END IF;

    -- Confirm the terminal "complete" step still exists
    IF NOT (v_steps ? 'complete') THEN
        RAISE EXCEPTION 'terminal "complete" step missing — migration may have corrupted the workflow';
END IF;

    RAISE NOTICE 'phase_2g followup: mark_work_item_complete inserted into image-build-handler tail';
END
$verify$;

COMMIT;

---
-- mark site work item failed after image run

-- phase_2g_followup_mark_work_item_failed.sql
--
-- Companion to phase_2g_followup_mark_work_item_complete.sql. Same
-- pattern, error side of the workflow:
--
-- Insert a new step `mark_work_item_failed` immediately before the
-- existing `complete_error` terminal. Redirect every step whose
-- `error_step` currently points at `complete_error` to point at the new
-- step instead. The new step then routes to `complete_error`.
--
-- After this, any failure in image-build-handler that would have hit
-- `complete_error` first updates its triggering site_work_items row to
-- status='failed' (and increments attempt_count), then terminates.
-- Dispatch retry semantics that key on attempt_count now have correct
-- input. Failed orchestrations are also auditable from the work item
-- itself via result.completed_by_orchestration_id.
--
-- The action gracefully no-ops if `input_data.work_item_id` is absent,
-- so manual triggers without a work_item_id continue to work.
--
-- The new step's own error_step is `complete_error` (not itself) — if
-- the status update itself fails, terminate normally rather than loop.
--
-- Steps whose error_step is currently `spawn_asset_deployer` (the
-- store_*_asset steps) are NOT changed. They route through deploy first
-- to try to salvage the image URL, and if THAT chain ultimately fails,
-- it lands at `mark_work_item_failed` via the redirected
-- call_asset_deployer error path. Same effect, no special-case.
--
-- Reversible. Backup taken.
--
-- Prerequisite: UpdateWorkItemStatusAction Go code deployed AND
-- registered (same prereq as the success-side migration).

\set ON_ERROR_STOP on

CREATE TABLE agent_def_image_build_handler_backup_20260514_pre_mark_work_item_failed AS
SELECT * FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

BEGIN;

-- Sanity: the success-side migration must already have applied
-- (mark_work_item_complete exists). Otherwise we're applying out of order.
DO $check$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'image-build-handler' AND is_active = true
           AND default_config #> '{workflow,steps,mark_work_item_complete}' IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'mark_work_item_complete missing — apply phase_2g_followup_mark_work_item_complete.sql first';
END IF;
    IF NOT EXISTS (
        SELECT 1 FROM agent_definitions
         WHERE type = 'image-build-handler' AND is_active = true
           AND default_config #> '{workflow,steps,complete_error}' IS NOT NULL
    ) THEN
        RAISE EXCEPTION 'complete_error terminal missing from workflow';
END IF;
END
$check$;

-- 1. Add the new mark_work_item_failed step.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,mark_work_item_failed}',
        $step$
            {
    "action": "update_work_item_status",
        "config": {
        "work_item_id_field": "input_data.work_item_id",
        "status": "failed",
        "skip_if_missing": true
    },
    "next_step": "complete_error",
    "error_step": "complete_error",
    "description": "Mark the triggering site_work_items row as failed before terminating. Gracefully no-ops if input_data.work_item_id is absent. error_step is complete_error so a failure of the status-update itself terminates normally rather than looping.",
    "output_field": "work_item_marked_failed"
}
$step$::jsonb
       ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- 2. Redirect every step's error_step from "complete_error" to
--    "mark_work_item_failed". Each step listed explicitly so the
--    migration is reviewable; spawn_*_asset steps' error_step is
--    deliberately NOT changed (those route to spawn_asset_deployer for
--    salvage and converge on call_asset_deployer's redirected error
--    path anyway).
UPDATE agent_definitions
SET default_config =
        jsonb_set(
                jsonb_set(
                        jsonb_set(
                                jsonb_set(
                                        jsonb_set(
                                                default_config,
                                                '{workflow,steps,call_hero_gen,error_step}',
                                                '"mark_work_item_failed"'::jsonb
                                        ),
                                        '{workflow,steps,call_logo_gen,error_step}',
                                        '"mark_work_item_failed"'::jsonb
                                ),
                                '{workflow,steps,call_variant_gen,error_step}',
                                '"mark_work_item_failed"'::jsonb
                        ),
                        '{workflow,steps,call_imagery_gen,error_step}',
                        '"mark_work_item_failed"'::jsonb
                ),
                '{workflow,steps,call_asset_deployer,error_step}',
                '"mark_work_item_failed"'::jsonb
        ),
    updated_at = now()
WHERE type = 'image-build-handler'
  AND is_active = true;

-- Verify
DO $verify$
DECLARE
v_steps      jsonb;
    v_err_step   text;
BEGIN
SELECT default_config #> '{workflow,steps}'
INTO v_steps
FROM agent_definitions
WHERE type = 'image-build-handler' AND is_active = true;

-- New step exists
IF NOT (v_steps ? 'mark_work_item_failed') THEN
        RAISE EXCEPTION 'mark_work_item_failed step missing after migration';
END IF;

    -- New step's targets correct
    IF (v_steps #>> '{mark_work_item_failed,next_step}') <> 'complete_error' THEN
        RAISE EXCEPTION 'mark_work_item_failed.next_step is %, expected complete_error',
                        v_steps #>> '{mark_work_item_failed,next_step}';
END IF;

    -- Each call_*_gen and call_asset_deployer redirected (explicit per step)
    v_err_step := v_steps #>> '{call_hero_gen,error_step}';
    IF v_err_step <> 'mark_work_item_failed' THEN
        RAISE EXCEPTION 'call_hero_gen.error_step is %, expected mark_work_item_failed', v_err_step;
END IF;

    v_err_step := v_steps #>> '{call_logo_gen,error_step}';
    IF v_err_step <> 'mark_work_item_failed' THEN
        RAISE EXCEPTION 'call_logo_gen.error_step is %, expected mark_work_item_failed', v_err_step;
END IF;

    v_err_step := v_steps #>> '{call_variant_gen,error_step}';
    IF v_err_step <> 'mark_work_item_failed' THEN
        RAISE EXCEPTION 'call_variant_gen.error_step is %, expected mark_work_item_failed', v_err_step;
END IF;

    v_err_step := v_steps #>> '{call_imagery_gen,error_step}';
    IF v_err_step <> 'mark_work_item_failed' THEN
        RAISE EXCEPTION 'call_imagery_gen.error_step is %, expected mark_work_item_failed', v_err_step;
END IF;

    v_err_step := v_steps #>> '{call_asset_deployer,error_step}';
    IF v_err_step <> 'mark_work_item_failed' THEN
        RAISE EXCEPTION 'call_asset_deployer.error_step is %, expected mark_work_item_failed', v_err_step;
END IF;

    -- Terminal complete_error still exists
    IF NOT (v_steps ? 'complete_error') THEN
        RAISE EXCEPTION 'complete_error terminal missing — migration corrupted workflow';
END IF;
    -- mark_work_item_complete (from sibling migration) still intact
    IF NOT (v_steps ? 'mark_work_item_complete') THEN
        RAISE EXCEPTION 'mark_work_item_complete missing — migration corrupted workflow';
END IF;

    RAISE NOTICE 'phase_2g followup: mark_work_item_failed inserted; 5 error_step redirections applied';
END
$verify$;

COMMIT;

--

-- unhardcode sdxl

    -- backup
clients_db=# SELECT snapshot_agent('image-build-handler', 'unhardcode sdxl image handler type');
NOTICE:  Snapshot captured: type=image-build-handler, source_version=1, source_id=04b10d94-11ee-447c-9ff9-7924b8e9897c, reason=unhardcode sdxl image handler type
            snapshot_agent
--------------------------------------
 04b10d94-11ee-447c-9ff9-7924b8e9897c
(1 row)


-- origin_model_workflow_propagation.sql  (v3 — column confirmed: default_config)
--
-- Propagate the real provider/model into assets.origin_model instead of the
-- hardcoded "sdxl". Workflow-JSON only — StoreAssetAction already supports
-- origin_model_field (v3_site_actions.go:2252-2257); the literal had to be
-- removed because it takes precedence over the field.
--
-- CONFIRMED against the live row (04b10d94-…, version 1):
--   • column : default_config   (NOT *_workflow — all three are null here)
--   • key    : type = 'image-build-handler'
--   • path   : {workflow,steps,<step>,config,...}  (wrapper present)
--   • single active row, version = 1

BEGIN;

-- 0. Snapshot before mutating.
SELECT snapshot_agent('image-build-handler', 'unhardcode sdxl image handler type');

-- 1. Sanity: expect all five to read "sdxl".
SELECT
    default_config #> '{workflow,steps,store_hero_asset,config,origin_model}'          AS hero,
  default_config #> '{workflow,steps,store_logo_asset,config,origin_model}'          AS logo,
  default_config #> '{workflow,steps,store_imagery_asset,config,origin_model}'       AS imagery,
  default_config #> '{workflow,steps,store_variant_asset,config,origin_model}'       AS variant,
  default_config #> '{workflow,steps,store_imagery_brand_asset,config,origin_model}' AS imagery_brand
FROM agent_definitions
WHERE type = 'image-build-handler' AND deleted_at IS NULL;

-- ── Part 1: add origin_model to the four call_*_gen output_mappings ──────────
UPDATE agent_definitions SET default_config =
                                 jsonb_set(jsonb_set(jsonb_set(jsonb_set(
                                                                       default_config,
                                                                       '{workflow,steps,call_hero_gen,config,output_mapping,origin_model}',    '"generate.response.origin_model"'::jsonb, true),
                                                               '{workflow,steps,call_logo_gen,config,output_mapping,origin_model}',    '"generate.response.origin_model"'::jsonb, true),
                                                     '{workflow,steps,call_imagery_gen,config,output_mapping,origin_model}', '"generate.response.origin_model"'::jsonb, true),
                                           '{workflow,steps,call_variant_gen,config,output_mapping,origin_model}', '"generate.response.origin_model"'::jsonb, true)
WHERE type = 'image-build-handler' AND deleted_at IS NULL;

-- ── Part 2: add origin_model_field to the five store_*_asset steps ───────────
UPDATE agent_definitions SET default_config =
                                 jsonb_set(jsonb_set(jsonb_set(jsonb_set(jsonb_set(
                                                                                 default_config,
                                                                                 '{workflow,steps,store_hero_asset,config,origin_model_field}',          '"image_result.origin_model"'::jsonb, true),
                                                                         '{workflow,steps,store_logo_asset,config,origin_model_field}',          '"image_result.origin_model"'::jsonb, true),
                                                               '{workflow,steps,store_imagery_asset,config,origin_model_field}',       '"image_result.origin_model"'::jsonb, true),
                                                     '{workflow,steps,store_variant_asset,config,origin_model_field}',       '"image_result.origin_model"'::jsonb, true),
                                           '{workflow,steps,store_imagery_brand_asset,config,origin_model_field}', '"image_result.origin_model"'::jsonb, true)
WHERE type = 'image-build-handler' AND deleted_at IS NULL;

-- ── Part 2b: remove the "sdxl" literal so the field is actually consulted ────
UPDATE agent_definitions SET default_config =
                                 default_config
    #- '{workflow,steps,store_hero_asset,config,origin_model}'
    #- '{workflow,steps,store_logo_asset,config,origin_model}'
    #- '{workflow,steps,store_imagery_asset,config,origin_model}'
    #- '{workflow,steps,store_variant_asset,config,origin_model}'
    #- '{workflow,steps,store_imagery_brand_asset,config,origin_model}'
WHERE type = 'image-build-handler' AND deleted_at IS NULL;

-- 2. Verify: literal gone, field + mapping present (spot-check imagery step).
SELECT
    default_config #> '{workflow,steps,store_imagery_asset,config,origin_model}'       AS imagery_literal_should_be_null,
  default_config #> '{workflow,steps,store_imagery_asset,config,origin_model_field}' AS imagery_field,
  default_config #> '{workflow,steps,call_imagery_gen,config,output_mapping,origin_model}' AS imagery_mapping,
  default_config #> '{workflow,steps,store_hero_asset,config,origin_model}'          AS hero_literal_should_be_null,
  default_config #> '{workflow,steps,store_hero_asset,config,origin_model_field}'    AS hero_field
FROM agent_definitions
WHERE type = 'image-build-handler' AND deleted_at IS NULL;
-- Expect: *_literal_should_be_null = NULL
--         imagery_field   = "image_result.origin_model"
--         hero_field      = "image_result.origin_model"
--         imagery_mapping = "generate.response.origin_model"

-- COMMIT;  -- uncomment when the verify output looks right
ROLLBACK;   -- safe default: review first, then switch to COMMIT

---

-- ============================================================================
-- Wire flag_page_image_rebuild into image-build-handler
-- DB: templates_db   Table: agent_definitions   type = 'image-build-handler'
-- Row id (from the provided dump): 04b10d94-11ee-447c-9ff9-7924b8e9897c
--
-- >>> SUPERSEDED IN PART — DO NOT REPLAY THIS SECTION AS WRITTEN (2026-07-28).
-- >>> The `"error_step": "complete"` below was disabled by seed 220
-- >>> (bugs_closed/086: it would end the orchestration GREEN on a failure) and
-- >>> then re-pointed at `mark_work_item_failed` by seed 259 on the owner's
-- >>> ruling. Section 0's precondition ("expect flag_rebuild_exists = f") is a
-- >>> printed expectation, not a guard — the UPDATE runs regardless, so a
-- >>> replay would silently restore the routing both later seeds removed.
-- >>> Replay only with the error_step line changed to mark_work_item_failed.
--
-- Adds a terminal step `flag_rebuild` after `mark_work_item_complete` and
-- redirects mark_work_item_complete.next_step → flag_rebuild. Every path
-- (needs_imagery, logo, hero, variant) converges at mark_work_item_complete,
-- so all flow through flag_rebuild; the action no-ops unless the work item's
-- spec.scope == 'page' (legacy logo/hero items have no spec.scope).
--
-- Config paths resolve via ExtractActionInputs Strategy 0 (dot-paths from
-- collected data): site_id ← site_record.site_id, scope/scope_ref ←
-- input_data.spec.*. The InputSpec field names (site_id/scope/scope_ref) match
-- these config keys.
-- ============================================================================

-- ── 0. Inspect first ────────────────────────────────────────────────────────
-- Confirm exactly one row, and the current terminal chain.
SELECT id, type, image_tag,
       default_config #>> '{workflow,steps,mark_work_item_complete,next_step}' AS mwc_next,
       (default_config #> '{workflow,steps,flag_rebuild}') IS NOT NULL          AS flag_rebuild_exists
FROM agent_definitions
WHERE type = 'image-build-handler';
-- Expect: one row, mwc_next = 'complete', flag_rebuild_exists = f.
-- If more than one row, scope the UPDATE below by id instead of type.

-- ── 1. Add the flag_rebuild step + redirect mark_work_item_complete ──────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
                default_config,
                '{workflow,steps,flag_rebuild}',
                '{
                    "action": "flag_page_image_rebuild",
                    "config": {
                        "site_id":   "site_record.site_id",
                        "scope":     "input_data.spec.scope",
                        "scope_ref": "input_data.spec.scope_ref"
                    },
                    "next_step": "complete",
                    "error_step": "complete",
                    "description": "Page-scoped imagery: flag the page needs_rebuild and emit needs_page so plan_sections re-resolves the now-present asset. No-ops for non-page-scoped (logo/legacy). error_step is complete so a failure here does not fail the asset workflow.",
                    "output_field": "rebuild_flagged"
                }'::jsonb,
                true
        ),
        '{workflow,steps,mark_work_item_complete,next_step}',
        '"flag_rebuild"'::jsonb,
        false
                     ),
    updated_at = now()
WHERE type = 'image-build-handler';
-- (swap the WHERE to `WHERE id = '04b10d94-11ee-447c-9ff9-7924b8e9897c'` if
--  step 0 showed more than one row.)

-- ── 2. Verify ────────────────────────────────────────────────────────────────
SELECT default_config #>> '{workflow,steps,mark_work_item_complete,next_step}'      AS mwc_next,
       default_config #>> '{workflow,steps,flag_rebuild,action}'                    AS fr_action,
       default_config #>> '{workflow,steps,flag_rebuild,next_step}'                 AS fr_next,
       default_config #>> '{workflow,steps,flag_rebuild,config,scope_ref}'          AS fr_scope_ref
FROM agent_definitions
WHERE type = 'image-build-handler';
-- Expect: mwc_next = 'flag_rebuild', fr_action = 'flag_page_image_rebuild',
--         fr_next = 'complete', fr_scope_ref = 'input_data.spec.scope_ref'.

-- ============================================================================
-- Registry entries (registry.go) — add for BOTH new local actions:
--
--   "flag_page_image_rebuild": {
--       Handler:     FlagPageImageRebuildAction,
--       Category:    "site",
--       Description: "Re-render a page after its image asset lands so the hero resolves",
--       IsLocal:     true,
--   },
--   "reconcile_section_data": {
--       Handler:     ReconcileSectionDataAction,
--       Category:    "site",
--       Description: "Re-trigger pages whose deferred section data is now query-resolvable",
--       IsLocal:     true,
--   },
-- ============================================================================