-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================

-- content-reviewer
-- Reviews content, may involve HITL
UPDATE agent_definitions
SET input_contract = '{
    "required": ["current_page", "page_content"],
    "optional": ["reviewed_brief"]
}'::jsonb,
    output_contract = '{
    "produces": ["review_result", "approved", "feedback"]
}'::jsonb
WHERE type = 'content-reviewer';

