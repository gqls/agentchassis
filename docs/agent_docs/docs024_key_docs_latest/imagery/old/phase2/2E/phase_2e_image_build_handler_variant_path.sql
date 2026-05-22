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
