-- 413_content_gap_planner_to_sonnet_5.sql
--
-- Moves content-gap-planner's `plan_gaps` step from claude-sonnet-4-6 to
-- claude-sonnet-5, and raises that step's max_tokens from 4000 to 16000 because
-- the model change silently turns adaptive thinking ON.
--
-- WHY THIS MIGRATION EXISTS AT ALL — it is the precondition for 414. Migration
-- 414 adds the LCO-008 cache breakpoint to this agent, and the entire economic
-- case for that marker rests on the 1-hour cache TTL shipped in v1.0.1301
-- (platform/aiservice/anthropic.go, `cacheTTL = "1h"`). Measured 2026-08-15 on
-- llm_call_log, 391 repeat-prefix pairs over 3 days: 1.0% of repeats fall within
-- 5 minutes, 99.7% within an hour. Break-even is ~22% at the 5m TTL and ~53% at
-- 1h. So at 5m the marker would COST ~24% more than not caching; at 1h it is a
-- large win. The marker is only worth adding if the 1h TTL genuinely applies.
--
-- AND THAT IS THE PROBLEM THIS FIXES. The 1h bucket was proven on 2026-08-15
-- against claude-sonnet-5 only — the probe recorded in anthropic.go returned
-- "ephemeral_1h_input_tokens": 6003 on that model. On claude-sonnet-4-6 the same
-- probe confirmed the field is ACCEPTED (HTTP 200, no 400) but happened to return
-- a cache READ with 0 in both creation buckets, so it did NOT re-prove the bucket.
-- Every scrap of post-roll cache evidence in the estate comes from sonnet-5:
-- all 17 council-gate seats run claude-sonnet-5, and council-gate is the only
-- agent in the fleet with a marker (89.6M cache reads over 3 days). Marking
-- content-gap-planner while it ran sonnet-4-6 would have made it the fleet's
-- first 1h-TTL user on a model where the bucket is unproven — betting the
-- marker's whole payoff on an unverified assumption. Moving the agent onto the
-- model where the bucket IS proven removes that bet rather than hedging it.
-- Owner decision, 2026-08-15.
--
-- ⚠ THE max_tokens CHANGE IS NOT OPTIONAL — IT IS WHAT KEEPS THIS SAFE.
--
-- On claude-sonnet-4-6, a request that omits the `thinking` parameter runs with
-- thinking OFF. On claude-sonnet-5, the same request runs ADAPTIVE THINKING by
-- default. This client never sends `thinking` unless `budget_tokens` is set
-- (platform/aiservice/anthropic.go: the field is only populated from
-- options["budget_tokens"]), and this step's ai_service config does not set it.
-- So the model swap alone silently enables thinking.
--
-- max_tokens is a hard cap on thinking PLUS response text together. This step
-- must return a complete JSON object; a budget sized for the answer alone can
-- now be consumed by thinking and truncate the JSON mid-structure. That failure
-- is real and already instrumented here — the client raises on
-- stop_reason == "max_tokens", and CLAUDE.md's standing rule is that
-- output_tokens == max_tokens means the completion was CUT, not finished.
--
-- WHY 16000, AND WHY NOT SIMPLY DISABLE THINKING INSTEAD. Disabling is not
-- reachable from this client: its only thinking path emits the
-- {"type": "enabled", "budget_tokens": N} form, which claude-sonnet-5 REJECTS
-- with a 400 (manual extended thinking is removed on that model), and there is
-- no code path that emits {"type": "disabled"}. Headroom is therefore the only
-- lever available without a Go change, a build and a fleet roll. 16000 is chosen
-- as the documented ceiling for NON-STREAMING requests — above roughly that,
-- SDK/HTTP timeouts become the failure mode instead of truncation, and this
-- client does not stream. Measured headroom: 178,363 output tokens across 404
-- calls is ~441 tokens per call for the JSON itself, so 16000 leaves the whole
-- remainder to thinking.
--
-- max_tokens is a CEILING, not a commitment — only tokens actually generated are
-- billed, so raising it costs nothing by itself and buys the truncation margin.
--
-- COST, STATED HONESTLY IN BOTH DIRECTIONS. claude-sonnet-5 uses a different
-- tokenizer: the same prompt text becomes roughly 30% MORE tokens than on
-- claude-sonnet-4-6. Sticker price is identical ($3/M input, $15/M output), so
-- that 30% is a real increase taken on its own, and adaptive thinking adds
-- output tokens this step did not previously spend. Against that: 414's caching
-- removes ~82% of the ~74% of each prompt that is shared prefix, which more than
-- covers the tokenizer inflation. There is also a temporary tailwind that must
-- NOT be mistaken for the steady state — claude-sonnet-5 is on introductory
-- pricing of $2/M input and $10/M output through 2026-08-31, after which it
-- returns to $3/$15. Re-measure in September rather than trusting an August
-- figure.
--
-- WHAT IS NOT A RISK HERE, checked rather than assumed: this client
-- deliberately never sends temperature, top_p or top_k (see the comment at
-- platform/aiservice/anthropic.go, "Temperature is intentionally NOT sent"), and
-- non-default sampling parameters are exactly what claude-sonnet-5 rejects with
-- a 400. Nothing in this step's config sets budget_tokens either. So the swap
-- carries no 400 surface; its risk is entirely the truncation one handled above.
--
-- VERIFY AFTER APPLYING — the truncation check is the one that matters:
--
--   SELECT created_at, model, input_tokens, output_tokens,
--          coalesce(cache_read_input_tokens,0) AS reads
--   FROM llm_call_log
--   WHERE agent_type='content-gap-planner' AND created_at > now() - interval '2 hours'
--   ORDER BY created_at;
--
-- output_tokens at or near the configured max_tokens means the JSON was CUT.
--
-- ROLLBACK: 413_content_gap_planner_to_sonnet_5_ROLLBACK.sql restores both
-- fields together. Do not revert the model without also reverting max_tokens —
-- they are one change (see the thinking note above).

SELECT snapshot_agent('content-gap-planner',
                      '413_content_gap_planner_to_sonnet_5.sql: pre-update');

BEGIN;

-- PRE-CONDITIONS.
DO $$
DECLARE
    n_defs     integer;
    cur_model  text;
    cur_max    text;
    has_budget boolean;
BEGIN
    SELECT count(*) INTO n_defs
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF n_defs <> 1 THEN
        RAISE EXCEPTION 'MIGRATION 413: expected exactly 1 live content-gap-planner definition, found %', n_defs;
    END IF;

    SELECT default_config->'workflow'->'steps'->'plan_gaps'->'config'->'ai_service'->>'model',
           default_config->'workflow'->'steps'->'plan_gaps'->'config'->>'max_tokens',
           (default_config->'workflow'->'steps'->'plan_gaps'->'config'->'ai_service' ? 'budget_tokens')
      INTO cur_model, cur_max, has_budget
    FROM agent_definitions
    WHERE type='content-gap-planner' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    IF cur_model IS DISTINCT FROM 'claude-sonnet-4-6' THEN
        RAISE EXCEPTION 'MIGRATION 413: plan_gaps model is %, expected claude-sonnet-4-6 — another lane has already changed it; re-read before applying', cur_model;
    END IF;

    -- budget_tokens would make the client send the {"type":"enabled"} thinking
    -- form, which claude-sonnet-5 rejects with a 400 on EVERY call. If a lane
    -- has added one since this migration was written, stop rather than ship a
    -- config that fails outright.
    IF has_budget THEN
        RAISE EXCEPTION 'MIGRATION 413: plan_gaps ai_service now sets budget_tokens — the client would send manual extended thinking, which claude-sonnet-5 rejects with a 400 on every call. Remove it before switching model.';
    END IF;

    RAISE NOTICE 'migration 413 pre-conditions OK: model=%, max_tokens=%, no budget_tokens', cur_model, coalesce(cur_max,'(unset)');
END $$;

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,plan_gaps,config,ai_service,model}',
            to_jsonb('claude-sonnet-5'::text)
        ),
        '{workflow,steps,plan_gaps,config,max_tokens}',
        to_jsonb(16000)
    ),
    version    = version + 1,
    updated_at = now()
WHERE type='content-gap-planner' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- POST-CONDITIONS. Both fields must move together — see the banner.
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

    IF new_model <> 'claude-sonnet-5' THEN
        RAISE EXCEPTION 'MIGRATION 413: model is % after update, expected claude-sonnet-5', new_model;
    END IF;

    -- The headroom assertion. A model change that lands WITHOUT it silently
    -- enables adaptive thinking against the old budget, and the failure appears
    -- as truncated JSON rather than as an error here.
    IF new_max < 16000 THEN
        RAISE EXCEPTION 'MIGRATION 413: max_tokens is % after update, expected >= 16000. claude-sonnet-5 runs adaptive thinking by default and max_tokens caps thinking PLUS response together — this budget would truncate the JSON plan.', new_max;
    END IF;

    RAISE NOTICE 'migration 413 OK: model=%, max_tokens=% (thinking headroom in place)', new_model, new_max;
END $$;

COMMIT;
