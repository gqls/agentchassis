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

