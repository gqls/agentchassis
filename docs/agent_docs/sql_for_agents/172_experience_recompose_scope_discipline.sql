-- 172_experience_recompose_scope_discipline.sql — Experience Loop, CP2 unblock #2.
--
-- PROBLEM (observed, run fbe12212, 2026-07-19): the council ran all 5 rounds and
-- escalated at the cap. Round 4 reached 3-of-4 approve. It never converged
-- because the MVP referee made substantially the SAME scope-cut objection in
-- rounds 1, 2, 3 and 5 — "defer the tool-arena rebuild / Journey C; the core
-- loop only needs arena.timer_seconds" — and the recompose step never acted on
-- it. Meanwhile each recompose answered a feasibility/journeys objection by
-- ADDING specification, which enlarged scope and re-triggered the MVP referee.
-- Round 4 -> 5 is the clean instance: journeys and mvp both reached approve in
-- round 4, then the recompose that answered feasibility pushed mvp back to
-- object.
--
-- ROOT CAUSE (prompt-level, not code): the recompose prompt labelled that seat
-- '## MVP referee said (advisory)'. That is true of its VERDICT rights (it
-- cannot veto) but false of its POWER: decideCouncil
-- (diagnose_council_decide_action.go:263-267) returns "revise" on ANY
-- reviewer's `object`, so an advisory seat's objection blocks approval exactly
-- as hard as a veto seat's. We told the composer the objection was optional,
-- and it treated it as optional — correctly, given what it was told.
--
-- FIX, two parts, both in the recompose prompt_template:
--   1. Relabel the section so the composer knows the seat blocks.
--   2. Add a SCOPE DISCIPLINE rule: every MVP-referee objection must be either
--      APPLIED (move to LATER) or REBUTTED in §4 in one sentence; and answering
--      one critic must never enlarge the MVP cut to satisfy another.
--
-- Deliberately NOT done: silencing the seat by excluding advisory reviewers from
-- the revise trigger. That needs Go + an image roll, and it would suppress a
-- critic that has been substantively RIGHT every round — the plan really is
-- over-scoped. Owner chose this direction 2026-07-19.
--
-- Applied as targeted replace() on the stored template rather than re-authoring
-- the whole ~6KB prompt, so it cannot silently drop unrelated prompt text.
-- Config-only: live on commit, no image roll. Seed 167 carries the same change
-- in-place so a re-apply cannot clobber it.

BEGIN;

SELECT snapshot_agent('experience-planner', 'pre-update: 172 recompose scope discipline')
WHERE EXISTS (SELECT 1 FROM agent_definitions
              WHERE type='experience-planner' AND is_active
                AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL);

-- 1. Relabel the MVP referee section.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,recompose,config,prompt_template}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'recompose'->'config'->>'prompt_template',
           '## MVP referee said (advisory)',
           '## MVP referee said (NO veto, but its objections BLOCK approval just like any other seat — see SCOPE DISCIPLINE)'
         )),
         false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. Insert SCOPE DISCIPLINE immediately before LENGTH DISCIPLINE.
UPDATE agent_definitions
   SET default_config = jsonb_set(
         default_config,
         '{workflow,steps,recompose,config,prompt_template}',
         to_jsonb(replace(
           default_config->'workflow'->'steps'->'recompose'->'config'->>'prompt_template',
           'LENGTH DISCIPLINE (this loop has truncated before',
           'SCOPE DISCIPLINE (run 7 escalated on exactly this): the council decides by "ANY objection => revise", so the MVP referee''s objection blocks approval as hard as a veto. Its scope cuts are NOT optional advice. For EVERY MVP-referee objection you must do one of two things, visibly: (a) MOVE the named item to the LATER list, or (b) keep it and state in §4, in one sentence, why the core loop genuinely cannot play without it. Silently keeping it is not an option and has already cost four rounds. Above all: answering one critic must NEVER enlarge the MVP cut to satisfy another — if fixing a feasibility or journeys objection requires adding scope, prefer CUTTING the feature that caused the objection over specifying it further. A smaller plan that all four seats approve beats a complete plan that never converges.' || chr(10) || chr(10) ||
           'LENGTH DISCIPLINE (this loop has truncated before'
         )),
         false)
 WHERE type='experience-planner'
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- Assert both edits landed exactly once, and that nothing else was lost.
DO $$
DECLARE
  t text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'recompose'->'config'->>'prompt_template'
    INTO t
    FROM agent_definitions
   WHERE type='experience-planner'
     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

  IF t IS NULL THEN
    RAISE EXCEPTION 'recompose prompt_template missing';
  END IF;
  IF position('SCOPE DISCIPLINE' in t) = 0 THEN
    RAISE EXCEPTION 'SCOPE DISCIPLINE rule not inserted';
  END IF;
  IF position('LENGTH DISCIPLINE' in t) = 0 THEN
    RAISE EXCEPTION 'LENGTH DISCIPLINE rule lost — replace() clobbered it';
  END IF;
  IF position('## MVP referee said (advisory)' in t) > 0 THEN
    RAISE EXCEPTION 'old advisory label still present';
  END IF;
  IF position('BLOCK approval just like any other seat' in t) = 0 THEN
    RAISE EXCEPTION 'MVP referee section not relabelled';
  END IF;
  IF position('END EXPERIENCE_PLAN' in t) = 0 THEN
    RAISE EXCEPTION 'output contract text lost from prompt';
  END IF;
END $$;

COMMIT;

-- Rollback: restore the snapshot taken above, or reverse both replace() calls.
