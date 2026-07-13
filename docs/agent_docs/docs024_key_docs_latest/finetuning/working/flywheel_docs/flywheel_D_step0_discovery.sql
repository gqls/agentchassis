-- ============================================================================
-- Flywheel D — Discovery: briefing-agent shape + training data availability
-- ============================================================================
-- Read-only. No writes. We run these, look at the output, then design.
--
-- Paste the four queries one at a time.
-- ============================================================================


-- ── 1. Current briefing-agent definition ──
-- We want: the LLM step name (that's what swap_agent_model targets),
-- the prompt template, the current model.

\echo '=== 1. briefing-agent definition ==='

SELECT
    type,
    is_active,
    is_snapshot,
    version,
    jsonb_pretty(
        jsonb_build_object(
            'workflow_start_step', default_config #> '{workflow,start_step}',
            'workflow_steps', (
                SELECT jsonb_object_agg(step_name, step_info)
                FROM (
                    SELECT
                        key as step_name,
                        jsonb_build_object(
                            'action', value->'action',
                            'next_step', value->'next_step',
                            'output_field', value->'output_field',
                            'ai_service', value #> '{config,ai_service}',
                            'has_prompt', (value #> '{config,prompt_template}') IS NOT NULL OR (value #> '{config,prompt}') IS NOT NULL
                        ) as step_info
                    FROM jsonb_each(default_config #> '{workflow,steps}')
                ) steps
            ),
            'top_level_model', default_config->'ai_service'->>'model',
            'top_level_provider', default_config->'ai_service'->>'provider'
        )
    ) as briefing_agent_shape
FROM agent_definitions
WHERE type = 'briefing-agent'
  AND is_active = true
  AND is_snapshot = false
  AND deleted_at IS NULL
ORDER BY version DESC
LIMIT 1;


-- ── 2. LLM steps inside briefing-agent ──
-- Zoom in on whichever steps use execute_llm_prompt.

\echo ''
\echo '=== 2. LLM-calling steps (full config) in briefing-agent ==='

SELECT
    step_name,
    jsonb_pretty(step_config) as config
FROM (
    SELECT
        key as step_name,
        value->'config' as step_config
    FROM agent_definitions,
         jsonb_each(default_config #> '{workflow,steps}')
    WHERE type = 'briefing-agent'
      AND is_active = true
      AND is_snapshot = false
      AND deleted_at IS NULL
      AND value->>'action' = 'execute_llm_prompt'
) llm_steps;


-- ── 3. briefing-agent call history in llm_call_log ──
-- How many successful calls? What's the shape of input/output?

\echo ''
\echo '=== 3a. Call counts by step ==='

SELECT
    step_name,
    count(*) as calls,
    count(*) FILTER (WHERE success) as successful,
    count(*) FILTER (WHERE NOT success) as failed,
    min(created_at) as first_call,
    max(created_at) as last_call,
    ROUND(AVG(latency_ms)) as avg_latency_ms,
    ROUND(AVG(input_tokens)) as avg_in_tokens,
    ROUND(AVG(output_tokens)) as avg_out_tokens
FROM llm_call_log
WHERE agent_type = 'briefing-agent'
GROUP BY step_name
ORDER BY calls DESC;

\echo ''
\echo '=== 3b. A single recent successful call (to see shape) ==='

SELECT
    agent_type,
    step_name,
    model,
    input_tokens,
    output_tokens,
    latency_ms,
    LEFT(prompt_rendered, 500) as prompt_head,
    LEFT(response_text, 500) as response_head
FROM llm_call_log
WHERE agent_type = 'briefing-agent'
  AND success = true
ORDER BY created_at DESC
LIMIT 1;


-- ── 4. Is mistral-small3.1 resolvable as an Ollama model? ──
-- We established earlier the adapter has mistral-small3.1:latest loaded.
-- Let's confirm nothing in the system blocks it.

\echo ''
\echo '=== 4. Are there any snapshots / alternate versions of briefing-agent we should know about ==='

SELECT type, version, is_active, is_snapshot, created_at
FROM agent_definitions
WHERE type = 'briefing-agent'
  AND deleted_at IS NULL
ORDER BY is_snapshot, version DESC;
