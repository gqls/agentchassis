-- 209_experience_gauntlet_round5_gaps.sql
--
-- The vonc-spark-game re-plan (corr fcdf8e72, fired after 207 landed) reached
-- 5 REVISE rounds without approval and ESCALATED (max_rounds hit) — a
-- designed circuit-breaker, not a bug: complete_escalated's own description
-- says "the disagreement IS the round-boundary decision menu... read
-- council_report artifacts". Round 5's objections were substantive and
-- narrow (journeys 1, feasibility 2, contracts 3 — honesty/mvp both
-- approved), converging steadily across rounds (feasibility dropped its
-- backend-doesn't-exist objection entirely after 207 landed in round 3).
--
-- compose/load_context read NO prior council_report or rejected-plan
-- history (verified: compose.config.input_fields = [experience_context,
-- input_data] only; load_context's query pulls live site state, not
-- artifact history) — so a bare re-fire starts genuinely blind and would
-- likely re-discover the same gaps at real credit cost. This migration
-- folds round 5's three concrete, addressable gaps into the compose
-- Decisions channel, same mechanism as 197/207, so the next fire's FIRST
-- draft already accounts for them:
--
--   (1) gauntlet-interface's live enter-button handler actively simulates a
--       round (starts the timer, scrolls, focuses an objective) instead of
--       showing the honest offline/disabled state — this is the component
--       built in the 2026-07-22 partial remedy and needs sequencing to
--       change, not just new markup.
--   (2) the EXACT verified JSON response shapes for round/position/defend —
--       supplied here first-hand from the tools-api source + a live
--       round-trip test (2026-07-25), not inferred. This is the precise gap
--       the contracts seat named ("access path is not exact").
--   (3) two named existing-loader gaps the contracts seat found by reading
--       real source: provocations-archive-loader has no detail_body/slug/
--       class-split logic, and tool-arena-interface's own source was never
--       supplied to a prior round for verification.
--
-- One surgical replace() appending a new decision paragraph after 207's
-- liveness-evidence text (same anchor point, so it composes cleanly).
--
-- Workstream: gauntlet_dead_cta. Config-only, live immediately.
-- ROLLBACK: snapshot below.

BEGIN;

SELECT snapshot_agent('experience-planner', '209_experience_gauntlet_round5_gaps: pre-update');

UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,compose,config,prompt_template}',
         to_jsonb(
           replace(
             default_config->'workflow'->'steps'->'compose'->'config'->>'prompt_template',
             $OLD$Every one of these was exercised for real, not asserted.$OLD$,
             $NEW$Every one of these was exercised for real, not asserted. ROUND-5 GAPS (from the council's own 2026-07-25 escalation, corr fcdf8e72 — address these directly rather than rediscovering them): (a) the EXACT verified JSON response shapes, taken from the tools-api source and a live round-trip, so the data contract can be exact rather than approximate: POST .../round -> {"round_id":"<uuid>","provocation":{"eyebrow":"...","headline":"...","body":"...","primary_cta":{"label":"...","url":"..."},"secondary_cta":{"label":"...","url":"..."},"stats":[{"value":"...","label":"..."}]}} (provocation is the verbatim 'today' object from /data/provocations.json — do not invent extra fields); POST .../position -> {"counter_position":"...","challenge":"..."} (two string fields, nothing else); POST .../defend -> {"verdict":"...","reasons":"..."} (two string fields, nothing else). (b) gauntlet-interface's CURRENT live js_content actively simulates entering a round on its enter-button click (starts the real timer, scrolls to the challenge panel, focuses the first objective) — this is the 2026-07-22 partial-remedy build and it must be SEQUENCED for change (not left as-is), so that in disabled/offline-scaffolding mode the click instead sets the honest offline status text and does nothing else; never let a click both look disabled and still run the old simulate-a-round behaviour. (c) two existing loaders/components need their own gaps named as sequenced steps if the plan touches them: provocations-archive-loader's current source only does cloneNode+setText for date/title/teaser/stat plus a conditional href — it has no detail_body/slug read and sets no --linked/--static class, so any archive-detail journey needs an explicit loader-modification step, not an assumption that it already splits entries; and any Journey referencing tool-arena-interface must have that component's actual html_template/js_content pulled into context and quoted before the plan asserts its selectors exist.$NEW$
           )
         ),
         true)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE rp text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'compose'->'config'->>'prompt_template' INTO rp
    FROM agent_definitions
   WHERE type='experience-planner' AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF position('ROUND-5 GAPS' in rp) = 0 THEN
    RAISE EXCEPTION '209: round-5 gaps text did not land';
  END IF;
  IF position('"counter_position":"...","challenge":"..."' in rp) = 0 THEN
    RAISE EXCEPTION '209: exact API shapes missing';
  END IF;
  IF position('actively simulates entering a round' in rp) = 0 THEN
    RAISE EXCEPTION '209: enter-button gap missing';
  END IF;
  -- untouched neighbours (197 + 207) must survive
  IF position('https://tools.apis.uk — DECIDED and LIVE' in rp) = 0 THEN
    RAISE EXCEPTION '209: 207 liveness evidence was disturbed';
  END IF;
  IF position('REAL DEBATE against an AI opponent' in rp) = 0 THEN
    RAISE EXCEPTION '209: D1-REVISED (197) was disturbed';
  END IF;
  IF position('hollow shell' in rp) = 0 THEN
    RAISE EXCEPTION '209: diagnosis line 3 (197) was disturbed';
  END IF;
  IF position('D2 (REVISED 2026-07-18' in rp) = 0 THEN
    RAISE EXCEPTION '209: D2 was disturbed';
  END IF;
  IF position('LENGTH AND QUOTING DISCIPLINE' in rp) = 0 THEN
    RAISE EXCEPTION '209: length discipline (176) was disturbed';
  END IF;
  IF position('<!-- END EXPERIENCE_PLAN -->' in rp) = 0 THEN
    RAISE EXCEPTION '209: output-format trailer was disturbed';
  END IF;
END $$;

COMMIT;
