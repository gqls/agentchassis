-- 413_content_gap_planner_to_sonnet_5_ROLLBACK.sql
--
-- Restores content-gap-planner's `plan_gaps` step to claude-sonnet-4-6 with
-- max_tokens 4000.
--
-- ⚠ REVERT 414 FIRST IF IT HAS BEEN APPLIED. The cache breakpoint added by
-- migration 414 is only economically sound at the 1-hour cache TTL, and the 1h
-- bucket is proven on claude-sonnet-5 but NOT on claude-sonnet-4-6 (the
-- 2026-08-15 probe confirmed the field is accepted on 4-6 without re-proving the
-- bucket). Rolling the model back while leaving the marker in place recreates
-- exactly the untested combination migration 413 existed to avoid — and if the
-- TTL silently behaves as 5 minutes there, the marker costs ~24% MORE than no
-- caching at all, with a measured 1.0% hit rate against a ~22% break-even. The
-- failure is silent: writes with almost no reads look like ordinary traffic.
--
--   ./414_content_gap_planner_cache_breakpoint_ROLLBACK.sql   -- run this first
--   ./413_content_gap_planner_to_sonnet_5_ROLLBACK.sql        -- then this
--
-- BOTH FIELDS MOVE TOGETHER, and that is the whole point of this file. On
-- claude-sonnet-4-6 a request that omits `thinking` runs with thinking OFF, so
-- the 16000 headroom is unnecessary there; on claude-sonnet-5 it is what stops
-- adaptive thinking truncating the JSON plan. Reverting the model while leaving
-- max_tokens raised is harmless (it is only a ceiling, and unused ceiling is not
-- billed). Reverting max_tokens while leaving the model on claude-sonnet-5 is
-- NOT harmless — that is the truncation configuration. If you only want one
-- half, take the model back and leave the headroom.

BEGIN;

DO $$
DECLARE
    cur_model  text;
    has_marker boolean;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->'ai_service'->>'model',
           (default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'prompt_template'
                LIKE '%<!--CACHE_BREAKPOINT-->%')
      INTO cur_model, has_marker
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF cur_model IS DISTINCT FROM 'claude-sonnet-5' THEN
        RAISE EXCEPTION 'ROLLBACK 413: model is % — already rolled back, or never applied', coalesce(cur_model,'(null)');
    END IF;

    IF has_marker THEN
        RAISE EXCEPTION 'ROLLBACK 413: the cache breakpoint from migration 414 is still present. Run 414_content_gap_planner_cache_breakpoint_ROLLBACK.sql FIRST — caching on claude-sonnet-4-6 relies on a 1h TTL bucket that has never been proven on that model, and if it behaves as 5m the marker costs ~24%% more than no caching, silently.';
    END IF;

    RAISE NOTICE 'rollback 413: reverting model to claude-sonnet-4-6 and max_tokens to 4000';
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,plan_gaps,config,ai_service,model}',
            to_jsonb('claude-sonnet-4-6'::text)
        ),
        '{workflow,steps,plan_gaps,config,max_tokens}',
        to_jsonb(4000)
    ),
    version    = version + 1,
    updated_at = now()
WHERE type='content-gap-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE
    new_model text;
    new_max   integer;
BEGIN
    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->'ai_service'->>'model',
           (default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'max_tokens')::integer
      INTO new_model, new_max
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF new_model <> 'claude-sonnet-4-6' OR new_max <> 4000 THEN
        RAISE EXCEPTION 'ROLLBACK 413: expected claude-sonnet-4-6 / 4000, got % / %', new_model, new_max;
    END IF;

    RAISE NOTICE 'rollback 413 OK: model=%, max_tokens=%', new_model, new_max;
END $$;

COMMIT;
