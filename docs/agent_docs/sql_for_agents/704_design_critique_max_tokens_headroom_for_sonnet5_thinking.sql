-- 704_design_critique_max_tokens_headroom_for_sonnet5_thinking.sql — the missing half of 662.
--
-- WHY: mig 662 moved the 018 critic's critique step to anthropic/claude-sonnet-5 carrying
-- max_tokens 6000 — a budget sized against the answer alone. On sonnet-5 a request that
-- omits `thinking` runs ADAPTIVE THINKING by default, and max_tokens caps thinking PLUS
-- response together, so a structured report can be cut mid-structure while the step reads
-- complete (the estate's output_tokens == max_tokens rule). This client cannot disable
-- thinking on sonnet-5 (its only thinking path emits budget_tokens, which sonnet-5 400s),
-- so headroom is the sole lever without a Go change. The council's round-2 REVISE on
-- 52c9a201 (2026-09-02 14:01Z, gating objection) flagged exactly this unaddressed risk.
--
-- WHY 16000: the documented ceiling for NON-STREAMING requests (above roughly that, HTTP
-- timeouts replace truncation; this client does not stream). max_tokens is a CEILING, not
-- a commitment — unused headroom is unbilled. Worked pattern: mig 415 (content_gap_planner,
-- bugfix_209), whose landmine also warns the step's bare config.max_tokens key is INERT —
-- this file writes the key the resolver actually reads (step config.ai_service.max_tokens;
-- resolveAIServiceConfig precedence: top-level default_config.max_tokens -> merged
-- ai_service.max_tokens -> hardcoded 2048), and asserts the RESOLVED value.
--
-- Pre-state [MEASURED 2026-09-02]: top NULL, root ai_service NULL, step_ai 6000,
-- no budget_tokens anywhere on the step. Exactly one live row.
-- ROLLBACK: jsonb_set the same path back to 6000 — but not while the step is on sonnet-5
-- (that recreates the truncation config); or restore the snapshot taken below.

SELECT snapshot_agent('design-critique-agent',
                      '704_design_critique_max_tokens_headroom_for_sonnet5_thinking.sql: pre-update');

BEGIN;

DO $$
DECLARE
    n_defs   integer;
    top_max  text;
    has_bt   boolean;
BEGIN
    SELECT count(*) INTO n_defs
    FROM agent_definitions
    WHERE type='design-critique-agent' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF n_defs <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 704: expected exactly 1 live design-critique-agent definition, found %', n_defs;
    END IF;

    SELECT default_config->>'max_tokens',
           (default_config->'workflow'->'steps'->'critique'->'config'->'ai_service' ? 'budget_tokens')
      INTO top_max, has_bt
    FROM agent_definitions
    WHERE type='design-critique-agent' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF top_max IS NOT NULL THEN
        RAISE EXCEPTION 'MIGRATION 704: default_config.max_tokens = % at the TOP LEVEL outranks the step ai_service key this migration writes — the write would be inert. Resolve the top-level value first.', top_max;
    END IF;
    IF has_bt THEN
        RAISE EXCEPTION 'MIGRATION 704: the critique step ai_service carries budget_tokens — sonnet-5 400s on that; this migration must not proceed until it is removed.';
    END IF;
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,critique,config,ai_service,max_tokens}',
        to_jsonb(16000))
WHERE type='design-critique-agent' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- POST-CONDITION: assert the RESOLVED value in the resolver's own precedence order,
-- never merely the key just written.
DO $$
DECLARE
    step_ai_max integer;
    top_max     text;
    root_ai_max text;
    effective   integer;
BEGIN
    SELECT (default_config->'workflow'->'steps'->'critique'->'config'->'ai_service'->>'max_tokens')::integer,
           default_config->>'max_tokens',
           default_config->'ai_service'->>'max_tokens'
      INTO step_ai_max, top_max, root_ai_max
    FROM agent_definitions
    WHERE type='design-critique-agent' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    effective := COALESCE(top_max::integer, step_ai_max, root_ai_max::integer, 2048);

    IF effective IS DISTINCT FROM 16000 THEN
        RAISE EXCEPTION 'MIGRATION 704: RESOLVED max_tokens is %, expected 16000. sonnet-5 adaptive thinking shares this cap with the response — an undersized budget truncates the report.', effective;
    END IF;
    RAISE NOTICE 'migration 704 OK: resolved max_tokens = 16000 (step ai_service key)';
END $$;

COMMIT;
