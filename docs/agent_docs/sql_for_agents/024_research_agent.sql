-- ============================================================================
-- Add Input Contracts to Existing Agents
-- ============================================================================
-- These contracts define what each agent expects to receive.
-- Contract validation will fail fast with clear error messages when required
-- fields are missing.
-- ============================================================================

-- research-agent
-- Researches products, competitors, etc.
UPDATE agent_definitions
SET input_contract = '{
    "required": ["research_request"],
    "optional": ["site_record", "page", "domain"]
}'::jsonb,
    output_contract = '{
    "produces": ["research_results", "products", "competitor_data"]
}'::jsonb
WHERE type = 'research-agent';
