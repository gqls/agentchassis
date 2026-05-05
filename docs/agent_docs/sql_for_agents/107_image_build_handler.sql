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