-- 415_content_gap_planner_max_tokens_at_the_key_the_code_reads.sql
--
-- CORRECTS MIGRATION 413. 413 raised max_tokens to 16000 at
-- `...plan_gaps.config.max_tokens` — a key NOTHING READS — while the key the
-- code actually resolves, `...plan_gaps.config.ai_service.max_tokens`, stayed at
-- 4000. This migration moves the headroom to the live key and deletes the dead
-- one. It is written as a superseding file rather than an edit to 413 because
-- 413 has already been applied.
--
-- WHY THIS IS URGENT AND NOT COSMETIC. 413 also switched the step to
-- claude-sonnet-5, and on that model a request omitting `thinking` runs ADAPTIVE
-- THINKING by default (claude-sonnet-4-6 ran it off). max_tokens caps thinking
-- PLUS response text together. So between 413 and this file the step ran with
-- thinking enabled against the ORIGINAL 4000 budget — precisely the truncation
-- configuration 413's banner said it existed to prevent. The JSON plan this step
-- returns can be cut mid-structure, and per CLAUDE.md output_tokens ==
-- max_tokens means the completion was CUT, not finished.
--
-- HOW THE MISTAKE SURVIVED ITS OWN GUARD — worth reading, because the guard
-- looked rigorous. 413's post-condition asserted `config.max_tokens >= 16000`
-- against the key it had just written, so it could not have come out any other
-- way; it measured the write, not the behaviour. The disconfirming check is to
-- assert on the key the RESOLVER reads, which is what this file does. A
-- `[MEASURED]` figure is only evidence if the measurement could have failed.
--
-- THE RESOLUTION ORDER, read from platform/orchestration/actions/ai_actions.go
-- rather than inferred from the field names:
--
--   agentConfig["max_tokens"]      -- agentConfig = agentDef.DefaultConfig, i.e.
--                                  -- the TOP LEVEL of default_config. Not the
--                                  -- step's config block, despite the name.
--   else aiServiceConfig["max_tokens"]
--                                  -- the MERGED ai_service: root `ai_service`
--                                  -- overlaid by the step's
--                                  -- config.ai_service (resolveAIServiceConfig)
--   else 2048                      -- the client's hardcoded transport floor
--
-- Measured on the live row before writing this file: top-level max_tokens unset,
-- root ai_service.max_tokens unset, step ai_service.max_tokens = 4000. So the
-- effective cap was 4000 and `config.max_tokens` was inert at any value.
--
-- WHY 16000 — unchanged from 413: it is the documented ceiling for NON-STREAMING
-- requests (above roughly that, HTTP timeouts replace truncation as the failure
-- mode, and this client does not stream), and this step's JSON answer measures
-- ~441 output tokens per call across 404 calls, leaving the remainder to
-- thinking. max_tokens is a CEILING, not a commitment — only generated tokens
-- are billed, so the headroom costs nothing until it is used.
--
-- ROLLBACK: 415_..._ROLLBACK.sql restores step ai_service.max_tokens to 4000.
-- Do NOT run it while the step is on claude-sonnet-5 — that recreates the
-- truncation configuration. Revert 413 first (which reverts the model), or leave
-- the headroom in place, since an unused ceiling is free.

SELECT snapshot_agent('content-gap-planner',
                      '415_content_gap_planner_max_tokens_at_the_key_the_code_reads.sql: pre-update');

BEGIN;

DO $$
DECLARE
    n_defs      integer;
    step_ai_max text;
    top_max     text;
BEGIN
    SELECT count(*) INTO n_defs
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF n_defs <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 415: expected exactly 1 live content-gap-planner definition, found %', n_defs;
    END IF;

    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->'ai_service'->>'max_tokens',
           default_config->>'max_tokens'
      INTO step_ai_max, top_max
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    -- A top-level max_tokens would OUTRANK the key this migration sets, making
    -- this change inert in exactly the way 413's was. Refuse rather than repeat
    -- the mistake one level up.
    IF top_max IS NOT NULL THEN
        RAISE EXCEPTION 'MIGRATION 415: default_config.max_tokens = % is set at the TOP LEVEL and outranks the step ai_service key this migration writes — setting the step key would be inert. Resolve the top-level value first.', top_max;
    END IF;

    RAISE NOTICE 'migration 415 pre-conditions OK: step ai_service.max_tokens = %, no top-level override', coalesce(step_ai_max,'(unset)');
END $$;

-- Set the headroom on the key the resolver reads, and drop the inert key 413
-- introduced so no future reader mistakes it for the live one.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_gaps,config,ai_service,max_tokens}',
        to_jsonb(16000)
    ) #- '{workflow,steps,plan_gaps,config,max_tokens}',
    version    = version + 1,
    updated_at = now()
WHERE type='content-gap-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- POST-CONDITIONS. These assert the RESOLVED value, not the written one.
DO $$
DECLARE
    step_ai_max  integer;
    dead_key     boolean;
    top_max      text;
    root_ai_max  text;
    effective    integer;
BEGIN
    SELECT (default_config->'workflow'->'steps'->'plan_gaps'->'config'->'ai_service'->>'max_tokens')::integer,
           (default_config->'workflow'->'steps'->'plan_gaps'->'config' ? 'max_tokens'),
           default_config->>'max_tokens',
           default_config->'ai_service'->>'max_tokens'
      INTO step_ai_max, dead_key, top_max, root_ai_max
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF dead_key THEN
        RAISE EXCEPTION 'MIGRATION 415: the inert config.max_tokens key is still present — a future reader would trust a number nothing reads';
    END IF;

    -- Recompute what ai_actions.go would resolve, in its own precedence order.
    effective := COALESCE(top_max::integer, step_ai_max, root_ai_max::integer, 2048);

    IF effective < 16000 THEN
        RAISE EXCEPTION 'MIGRATION 415: RESOLVED max_tokens is %, expected >= 16000. claude-sonnet-5 runs adaptive thinking by default and max_tokens caps thinking PLUS response together — this budget truncates the JSON plan.', effective;
    END IF;

    RAISE NOTICE 'migration 415 OK: resolved max_tokens = % (step ai_service key), inert key removed', effective;
END $$;

COMMIT;
