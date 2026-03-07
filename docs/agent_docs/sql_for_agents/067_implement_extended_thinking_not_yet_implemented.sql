-- ============================================================================
-- Enable extended thinking on classifier and planner
-- Run AFTER deploying the anthropic.go and ai_actions.go patches
--
-- budget_tokens: 10000 gives the model room to reason through
-- industry analysis, competitor assessment, and architectural decisions.
-- Adds ~30-60 seconds to each call. Cost: ~$0.30-0.50 per call.
-- ============================================================================

-- Classifier: add budget_tokens to ai_service config
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,classify_and_extract,config,ai_service,budget_tokens}',
        '10000'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'domain-research-classifier' AND deleted_at IS NULL;

-- Planner: add budget_tokens to ai_service config
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_site,config,ai_service,budget_tokens}',
        '10000'::jsonb
                     ),
    updated_at = NOW()
WHERE type = 'build-site-planner' AND deleted_at IS NULL;

-- Verify
SELECT type,
       default_config->'workflow'->'steps'->
       CASE type
           WHEN 'domain-research-classifier' THEN 'classify_and_extract'
           WHEN 'build-site-planner' THEN 'plan_site'
           END->'config'->'ai_service' as ai_service
FROM agent_definitions
WHERE type IN ('domain-research-classifier', 'build-site-planner')
  AND deleted_at IS NULL;

-- To disable extended thinking later (e.g. for cost control):
-- UPDATE agent_definitions
-- SET default_config = default_config #- '{workflow,steps,classify_and_extract,config,ai_service,budget_tokens}'
-- WHERE type = 'domain-research-classifier';