-- Fix intake-orchestrator to use confirmed_type.recommended_builder instead of classification
-- This ensures the HITL confirmation is actually respected

-- Update the spawn_builder step to use confirmed_type
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,spawn_builder,config,agent_type_field}',
        '"confirmed_type.recommended_builder"'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'intake-orchestrator';

-- Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'spawn_builder'->'config' as spawn_builder_config
FROM agent_definitions
WHERE type = 'intake-orchestrator';

-- Also verify pageflow-builder exists and is active
SELECT
    type,
    display_name,
    version,
    is_active,
    status
FROM agent_definitions
WHERE type = 'pageflow-builder';

-- Check all active builders (what intake-orchestrator will discover)
SELECT
    type,
    display_name,
    version,
    is_active
FROM agent_definitions
WHERE type LIKE '%-builder'
  AND is_active = true
ORDER BY type;

---

update paths again

-- Fix hitl_confirm_type default_from paths for site_type field
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,hitl_confirm_type,config,fields,0,default_from}',
        '"classification.response.result.site_type"'::jsonb
                     )
WHERE type = 'intake-orchestrator';

-- Fix hitl_confirm_type default_from paths for recommended_builder field
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,hitl_confirm_type,config,fields,1,default_from}',
        '"classification.response.result.recommended_builder"'::jsonb
                     )
WHERE type = 'intake-orchestrator';

-- Verify the changes
SELECT
    type,
    default_config->'workflow'->'steps'->'hitl_confirm_type'->'config'->'fields'->0->'default_from' as field_0_default,
    default_config->'workflow'->'steps'->'hitl_confirm_type'->'config'->'fields'->1->'default_from' as field_1_default,
    default_config->'workflow'->'steps'->'fetch_questionnaire'->'config'->'agent_type_field' as fetch_q_agent_type
FROM agent_definitions
WHERE type = 'intake-orchestrator';

-- path change
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,fetch_questionnaire,config,agent_type_field}',
        '"confirmed_type.recommended_builder"'::jsonb
                     )
WHERE type = 'intake-orchestrator';


---


increase timeout

-- ===========================================================================
-- FIX: Increase timeout for call_builder in intake-orchestrator
-- ===========================================================================

-- Step 1: Check current config
SELECT
    type,
    default_config->'workflow'->'steps'->'call_builder'->'config' as call_builder_config
FROM agent_definitions
WHERE type = 'intake-orchestrator';

-- Step 2: Update timeout to 2 hours (7200 seconds) for HITL workflows
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,call_builder,config}',
        COALESCE(default_config->'workflow'->'steps'->'call_builder'->'config', '{}'::jsonb) ||
        '{
          "timeout_seconds": 7200,
          "max_retries": 10
        }'::jsonb
                     )
WHERE type = 'intake-orchestrator'
  AND default_config->'workflow'->'steps'->'call_builder' IS NOT NULL;

-- Step 3: Verify the change
SELECT
    type,
    default_config->'workflow'->'steps'->'call_builder'->'config'->'timeout_seconds' as timeout_seconds,
    default_config->'workflow'->'steps'->'call_builder'->'config'->'max_retries' as max_retries
FROM agent_definitions
WHERE type = 'intake-orchestrator';