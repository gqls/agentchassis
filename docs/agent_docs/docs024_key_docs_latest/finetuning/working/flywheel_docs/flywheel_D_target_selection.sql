-- ============================================================================
-- Flywheel D — Target agent selection based on logged call volume
-- ============================================================================
-- Find which agents have the richest training data already accumulated.
-- We want: high volume, high success rate, short-to-medium output length
-- (easier to eval), structured JSON where possible (automatic scoring),
-- and low reasoning complexity (local models can handle).
-- ============================================================================


-- ── 1. Agents by call volume and recency ──

\echo '=== 1. Agents by logged call volume ==='

SELECT
    CASE WHEN agent_type = '' OR agent_type IS NULL THEN '(empty)' ELSE agent_type END as agent_type,
    step_name,
    count(*) as calls,
    count(*) FILTER (WHERE success) as ok,
    count(*) FILTER (WHERE NOT success) as failed,
    ROUND(AVG(input_tokens)) as avg_in,
    ROUND(AVG(output_tokens)) as avg_out,
    ROUND(AVG(latency_ms)) as avg_ms,
    MIN(created_at)::date as first,
    MAX(created_at)::date as last
FROM llm_call_log
GROUP BY agent_type, step_name
ORDER BY calls DESC
LIMIT 25;


-- ── 2. Zoom on anything design-related ──

\echo ''
\echo '=== 2. Design-related or content-related agents specifically ==='

SELECT
    agent_type,
    step_name,
    count(*) as calls,
    count(*) FILTER (WHERE success) as ok,
    ROUND(AVG(output_tokens)) as avg_out,
    max(created_at) as most_recent
FROM llm_call_log
WHERE agent_type ILIKE '%design%'
   OR agent_type ILIKE '%content%'
   OR agent_type ILIKE '%writer%'
   OR agent_type ILIKE '%audit%'
   OR agent_type ILIKE '%plan%'
   OR agent_type ILIKE '%classif%'
GROUP BY agent_type, step_name
ORDER BY calls DESC;


-- ── 3. Sample a recent response from each of the top 5 agent types ──
-- Just to see what the outputs look like.

\echo ''
\echo '=== 3. One recent successful response per top agent ==='

WITH top_agents AS (
    SELECT agent_type, step_name
    FROM llm_call_log
    WHERE success = true
      AND agent_type IS NOT NULL AND agent_type != ''
    GROUP BY agent_type, step_name
    ORDER BY count(*) DESC
    LIMIT 5
),
latest_each AS (
    SELECT DISTINCT ON (l.agent_type, l.step_name)
        l.agent_type,
        l.step_name,
        l.model,
        l.output_tokens,
        l.response_text
    FROM llm_call_log l
    JOIN top_agents t ON t.agent_type = l.agent_type AND t.step_name = l.step_name
    WHERE l.success = true
    ORDER BY l.agent_type, l.step_name, l.created_at DESC
)
SELECT
    agent_type,
    step_name,
    model,
    output_tokens,
    LEFT(response_text, 300) as response_head
FROM latest_each;


-- ── 4. Which of those agents have the simplest workflows? ──
-- Short responses + JSON output + single LLM step = easiest to eval.

\echo ''
\echo '=== 4. Candidate agents: output looks like JSON, under 500 tokens, recent activity ==='

SELECT
    agent_type,
    step_name,
    count(*) FILTER (WHERE success) as successful_calls,
    ROUND(AVG(output_tokens)) as avg_out,
    count(*) FILTER (WHERE response_text LIKE '{%' OR response_text LIKE '[%') as json_like,
    max(created_at) as last_call
FROM llm_call_log
WHERE agent_type IS NOT NULL AND agent_type != ''
  AND success = true
GROUP BY agent_type, step_name
HAVING count(*) FILTER (WHERE success) >= 5
   AND AVG(output_tokens) < 800
ORDER BY successful_calls DESC;
