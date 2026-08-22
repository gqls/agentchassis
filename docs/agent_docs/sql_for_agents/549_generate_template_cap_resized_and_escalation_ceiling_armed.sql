-- 549_generate_template_cap_resized_and_escalation_ceiling_armed.sql
--
-- bugs_open/337: component-creator's generate_template step is capped at
-- ai_service.max_tokens = 16000, and ONE section type (loans-credit-health-check)
-- reliably produces more — nine truncations across three work items, every one
-- cut at output_tokens=16000 with 46,441-48,817 chars recovered, three full
-- generations burned per item, two live pages shipped hollow. The fleet headroom
-- monitor (LCO-007, fleet-step-token-pressure) has flagged the step since
-- 2026-08-18: "T generate_template@16000 — n=229, p95 92.4%, peak 100.0%,
-- truncated 9".
--
-- TWO HALVES, one migration (the 484 shape — raise what restores function AND
-- arm what closes the door):
--
--   1. max_tokens 16000 -> 24000. Sized from measurement, not taste: successful
--      calls' p95 is 13,633 (85% of the old cap) and their maximum 15,374 (96%),
--      so the old cap was tight for the step's ORDINARY work, not just the
--      failing section; the cut generations extrapolate to ~19-22k tokens
--      (~47.5k chars cut mid-document at ~2.95 chars/token). 24000 clears the
--      ordinary distribution by ~75% and the estimated failing need with margin.
--      The sibling whole-component writers were levelled long ago and this step
--      was missed: generate_tool_html and improve_tool run at 32000 (migration
--      168 lineage, bugs_closed/012's levelling table), recreate_tool at 64000.
--
--   2. max_tokens_ceiling 32000 — NEW KEY, read by the bugs_open/337 escalation
--      in execute_llm_prompt (platform/orchestration/actions/
--      truncation_escalation.go, same commit): when a call comes back truncated
--      (typed TruncatedError) and the ceiling exceeds the cap the call was sent
--      with, the step retries ONCE at the ceiling instead of burning its
--      remaining attempts on a deterministic refusal. INERT UNTIL THE NEXT
--      CHASSIS ROLL — deliberately: applied before the roll, half 1 is live
--      immediately (the old binary reads ai_service.max_tokens) and this key
--      waits for the code that consumes it. That is staged arming, not dead
--      config; there is NO ordering constraint in either direction.
--
-- WHY 32000 AND NOT MORE: the chassis does not stream (anthropic.go's single
-- http.Client, 600s timeout), so the real output bound is wall-clock. This
-- step measures 92-127 tok/s on claude-sonnet-4-6 (llm_call_log latencies: the
-- nine 16,000-token cut calls each took 165-170s). 32000 tokens at the observed
-- WORST throughput (91.7 tok/s) is ~349s — clock-safe; 40000+ at a conservative
-- 60 tok/s crosses 600s and converts a loud truncation into a silent clock
-- death, which LCO-007's own C-vs-T doctrine says is the strictly worse trade.
--
-- WHY THE ROUTINE CAP IS NOT SIMPLY 32000: an escalation is a FORENSIC EVENT —
-- the cut first call logs an "ESCALATED (bugs_open/337: ...)" llm_call_log row,
-- so demand above 24000 stays visible and queryable. A step parked at the
-- ceiling has no headroom signal left except the next truncation.
--
-- MODEL INTERPLAY (LANDMINES "claude-sonnet-4-6 -> claude-sonnet-5 turns
-- adaptive thinking ON"): these numbers are sized for sonnet-4-6, where
-- max_tokens is visible text. If this step is ever moved to sonnet-5, the cap
-- becomes a THINKING+TEXT budget and both numbers must be re-derived.
--
-- 067-SWEEP NOTE (bugs_closed/067: a cap defect on one step means sweep every
-- step of the agent): generate_template is component-creator's ONLY
-- execute_llm_prompt step, so this migration IS the whole sweep for this agent.
-- [MEASURED 2026-08-22, after the council's prior_art_librarian seat objected
-- that this claim carried no measurement tag — it was right, the claim had been
-- asserted from a step-name list rather than queried]:
--   SELECT s.key, s.value->>'action', s.value->'config'->'ai_service'->>'max_tokens'
--   FROM agent_definitions a, LATERAL jsonb_each(a.default_config->'workflow'->'steps') s
--   WHERE a.type='component-creator' AND a.is_active
--     AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;
-- returns SIX steps: generate_template (execute_llm_prompt, cap 16000) plus
-- complete/ensure_site_record/load_existing_component/read_site_spec/
-- store_component — none of which is an LLM action and none of which carries a
-- cap. The query could have returned a second LLM step and did not.
--
-- Scoped by type + live-row predicate, pre-state gated, DO/RAISE verify
-- asserting the RESOLVED value (the 415 pattern — never assert the key you just
-- wrote), snapshot first, rollback sidecar. Config half LIVE ON APPLY.

SELECT snapshot_agent('component-creator', 'migration 549: pre-update (bugs_open/337 cap resize + escalation ceiling)');

BEGIN;

-- ── Pre-conditions ──────────────────────────────────────────────────────────
DO $$
DECLARE
    live_rows integer;
    step_cap  text;
    top_cap   text;
    ceiling   text;
BEGIN
    SELECT count(*) INTO live_rows
    FROM agent_definitions
    WHERE type='component-creator' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF live_rows IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION 'MIGRATION 549: expected exactly 1 live component-creator row, found % — a second active row would make this UPDATE ambiguous. Resolve first.', live_rows;
    END IF;

    SELECT default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens}',
           default_config->>'max_tokens',
           default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens_ceiling}'
    INTO step_cap, top_cap, ceiling
    FROM agent_definitions
    WHERE type='component-creator' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    -- Refuse to overwrite a value someone else has already moved (the 484 gate).
    IF step_cap IS DISTINCT FROM '16000' THEN
        RAISE EXCEPTION 'MIGRATION 549: generate_template ai_service.max_tokens is % not 16000 — another change landed first. Re-derive this migration against the live value.', COALESCE(step_cap,'ABSENT');
    END IF;

    -- A top-level max_tokens OUTRANKS the step ai_service key in the resolver
    -- (ai_actions.go:358-364), which would make this write inert while its own
    -- post-condition read back the written key happily (LANDMINES: migration
    -- 413's exact failure). Refuse rather than write a key nothing reads.
    IF top_cap IS NOT NULL THEN
        RAISE EXCEPTION 'MIGRATION 549: default_config.max_tokens = % is set at the TOP LEVEL and outranks the step ai_service key this migration writes. Resolve the top-level value first.', top_cap;
    END IF;

    -- Double-apply refusal.
    IF ceiling IS NOT NULL THEN
        RAISE EXCEPTION 'MIGRATION 549: max_tokens_ceiling already present (%) — refusing to double-apply.', ceiling;
    END IF;
END $$;

-- ── The write: both keys in one statement ──────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(default_config,
            '{workflow,steps,generate_template,config,ai_service,max_tokens}',
            to_jsonb(24000)),
        '{workflow,steps,generate_template,config,ai_service,max_tokens_ceiling}',
        to_jsonb(32000)),
    version = version + 1,
    updated_at = now()
WHERE type='component-creator' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── Post-conditions: assert the RESOLVED value, not the written key ─────────
DO $$
DECLARE
    top_cap    text;
    step_cap   text;
    root_cap   text;
    resolved   integer;
    ceiling    text;
    dead_key   text;
    n_updated  integer;
BEGIN
    SELECT count(*) INTO n_updated
    FROM agent_definitions
    WHERE type='component-creator' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens_ceiling}' = '32000';
    IF n_updated IS DISTINCT FROM 1 THEN
        RAISE EXCEPTION 'MIGRATION 549: expected exactly 1 row carrying the new ceiling, found %.', n_updated;
    END IF;

    SELECT default_config->>'max_tokens',
           default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens}',
           default_config#>>'{ai_service,max_tokens}',
           default_config#>>'{workflow,steps,generate_template,config,ai_service,max_tokens_ceiling}',
           default_config#>>'{workflow,steps,generate_template,config,max_tokens}'
    INTO top_cap, step_cap, root_cap, ceiling, dead_key
    FROM agent_definitions
    WHERE type='component-creator' AND is_active
      AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

    -- The resolver's own precedence order (ai_actions.go:358-364 + the
    -- aiservice fallback): top-level, then step ai_service (overlaying root),
    -- then the provider default.
    resolved := COALESCE(top_cap::integer, step_cap::integer, root_cap::integer, 2048);
    IF resolved IS DISTINCT FROM 24000 THEN
        RAISE EXCEPTION 'MIGRATION 549: RESOLVED max_tokens is %, expected 24000 — a higher-precedence key is outranking the write.', resolved;
    END IF;

    IF ceiling IS DISTINCT FROM '32000' THEN
        RAISE EXCEPTION 'MIGRATION 549: max_tokens_ceiling is %, expected 32000.', COALESCE(ceiling,'ABSENT');
    END IF;

    -- Negative control: the write must not have minted the inert
    -- config.max_tokens spelling (LANDMINES: that key reads NULL for every
    -- consumer and looks exactly like the live one).
    IF dead_key IS NOT NULL THEN
        RAISE EXCEPTION 'MIGRATION 549: inert config.max_tokens key present (%) — wrong depth was written.', dead_key;
    END IF;
END $$;

COMMIT;
