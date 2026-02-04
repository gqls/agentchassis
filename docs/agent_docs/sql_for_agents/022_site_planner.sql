-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================

-- site-planner
-- Plans site structure, pages, and components
UPDATE agent_definitions
SET input_contract = '{
    "required": ["site_record"],
    "optional": ["input_data", "reviewed_brief"]
}'::jsonb,
    output_contract = '{
    "produces": ["validated_plan", "pages", "style_collection", "needs_logo", "needs_images"]
}'::jsonb
WHERE type = 'site-planner';


-- switch to haiku

-- =============================================================
-- 1. Switch site-planner model from sonnet to haiku
--
-- The site-planner does structured planning (component selection,
-- page layout) which haiku handles well. This reduces cost and
-- is less likely to hit rate limits during overload.
-- =============================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,ai_service,model}',
        '"claude-haiku-4-5"'
                     ),
    updated_at = now()
WHERE type = 'site-planner';

-- Verify
SELECT
    type,
    default_config->'workflow'->'steps'->'plan_site'->'config'->'ai_service'->>'model' as planner_model
FROM agent_definitions
WHERE type = 'site-planner';
