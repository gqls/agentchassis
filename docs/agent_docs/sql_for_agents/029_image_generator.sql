-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================


-- image-generator
-- Generates images using AI
UPDATE agent_definitions
SET input_contract = '{
    "required": [],
    "optional": ["page", "site_record", "reviewed_brief", "prompt", "image_prompts"]
}'::jsonb,
    output_contract = '{
    "produces": ["image_url", "image_data"]
}'::jsonb
WHERE type = 'image-generator';

