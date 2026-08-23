-- 569 — bugs_open/305: raise the `rewrite_negations` repair ceiling from 2000 to
-- 16000, because the ceiling is failing on exactly the pages the gate exists for.
--
-- WHAT 517 SAID, AND WHY IT WAS REASONABLE. 517 set `max_tokens` to 2000 with a
-- stated rationale: "the answer is a handful of replacement sentences, and a
-- section-sized ceiling would only buy room for the model to write an essay this
-- gate would then reject." It also wrote down the failure mode it was accepting:
-- "⚠ If a future model runs adaptive thinking out of this budget the call yields
-- no text, the repair fails CLOSED (splices nothing) and says
-- `repair_unavailable` — loud, not silent, which is the whole reason that status
-- exists." That instrument worked. This migration is what the instrument was for.
--
-- MEASURED 2026-08-23 ~12:35Z, over the live `copy_gate*` markers then present
-- (a rolling window — `orchestration_states` is reaped, and the population moved
-- 42 -> 44 during the measurement itself):
--
--   targets | markers |      outcome
--   --------+---------+--------------------
--         1 |      36 | repaired
--         2 |       2 | repaired
--         3 |       3 | repaired
--         5 |       2 | repaired
--         9 |       1 | repair_unavailable
--        10 |       1 | repair_unavailable
--
-- Not intermittent truncation — a hard capacity limit with NO exceptions either
-- side of it. Every marker with <=5 targets repaired; both with >=9 failed. The
-- two failures carry the ceiling in their own error text:
--   "response truncated: stop_reason=max_tokens (output_tokens=2000 reached the
--    configured cap, 2066 chars recovered)"   <- 10 targets
--   "... (output_tokens=2000 reached the configured cap, 0 chars recovered)"
--                                             <- 9 targets, no text at all
--
-- COST OF THE CEILING: those 2 markers hold **19 of the 78 targets** in the
-- window — **24.4%** of all repair targets are lost to it, against the 12 lost
-- to the accounting hole that 6e9cb411d fixes. The failure is total: the round
-- is discarded whole, so a page with 10 constructions gets ZERO repaired while a
-- page with 1 gets its one.
--
-- WHY 2000 CANNOT HOLD, mechanically, and it is not about essays:
--   (a) `max_tokens` is an enforced per-response ceiling THE MODEL IS NOT AWARE
--       OF. It cannot pace itself to fit; it is simply cut. So a ceiling that is
--       too small does not buy brevity, it buys nothing at all.
--   (b) the answer is O(targets), not O(1) — the model echoes `from` AND `to`
--       for each target, so a dense page's answer is several times a sparse
--       page's. "A handful of replacement sentences" is the 1-target case, which
--       is 36 of 44 markers and is why this held for two days.
--   (c) `claude-sonnet-5` with no `thinking` block set runs ADAPTIVE THINKING,
--       and thinking tokens are drawn from this same ceiling. That is 517's own
--       ⚠, and the 9-target run recovering 0 chars is what it looks like.
--
-- WHY 16000: it is what the SIBLING STEP IN THE SAME SUB-WORKFLOW already uses —
-- `generate_content`, which writes a whole section, is at 16000 while the step
-- that rewrites a few of its sentences was at 2000. It is also the documented
-- default for a non-streaming request. This keeps 517's instinct that the repair
-- should not exceed the writer; it just stops it being 8x smaller.
--
-- COST: output tokens are billed AS GENERATED, not as reserved — the ceiling is
-- a safety limit, not an allocation. The 36 one-target markers will emit exactly
-- what they emit today and cost exactly what they cost today. Only the dense
-- pages, which currently produce nothing, will spend more.
--
-- WHAT THIS DOES NOT FIX: the ceiling is still fixed while `plan.targets` is
-- unbounded (the page budget sets a TOLERANCE, not a cap — every hit past it
-- becomes a target). 16000 covers every page yet observed with a wide margin,
-- but a dense enough page would truncate again, and it would again fail loudly
-- as `repair_unavailable`. Chunking the targets across calls is the structural
-- answer and is a CODE change; it is recorded in bugs_open/305 §27, not smuggled
-- in here.
--
-- Config-only: live the moment it applies, no image, no roll. Needle-gated on
-- the CURRENT value of 2000, so a re-run is a 0-row no-op and this cannot
-- silently overwrite a later deliberate retune by another session.

BEGIN;

SELECT snapshot_agent('page-content-writer',
                      '569_rewrite_negations_ceiling_fits_a_dense_page.sql: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(default_config,
         '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,max_tokens}',
         '16000'::jsonb),
       updated_at = now()
 WHERE type = 'page-content-writer' AND is_active
   AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,action}' = 'rewrite_negations'
   AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,max_tokens}' = '2000';

DO $$
DECLARE n int; mt text; model text;
BEGIN
  SELECT count(*) INTO n FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
     AND default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,max_tokens}' = '16000';
  IF n <> 1 THEN
    RAISE EXCEPTION '569 FAILED: expected exactly 1 page-content-writer with rewrite_negations max_tokens=16000, got %', n;
  END IF;

  -- The rest of the ai_service block must be untouched: this migration changes
  -- ONE number, and a jsonb_set against a moved path would mint an orphan key
  -- rather than fail. Assert the neighbours the step actually needs.
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,rewrite_negations,config,ai_service,model}'
    INTO model FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF model IS DISTINCT FROM 'claude-sonnet-5' THEN
    RAISE EXCEPTION '569 FAILED: ai_service.model is % — the block moved or was rewritten, and 569 changed the wrong thing', COALESCE(model, '<null>');
  END IF;

  -- The sibling this value is anchored to. If generate_content is no longer at
  -- 16000, the "never exceed the writer" rationale above needs re-reading.
  SELECT default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,ai_service,max_tokens}'
    INTO mt FROM agent_definitions
   WHERE type = 'page-content-writer' AND is_active
     AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
  IF mt IS DISTINCT FROM '16000' THEN
    RAISE WARNING '569 NOTE: generate_content max_tokens is %, not 16000 — the anchor this value was chosen against has moved; re-read the rationale in 569.', COALESCE(mt, '<null>');
  END IF;

  RAISE NOTICE '569 OK: rewrite_negations repairs a dense page at 16000, anchored on generate_content';
END $$;

COMMIT;
