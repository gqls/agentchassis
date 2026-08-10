-- 377_council_gate_cache_breakpoint_reorder.sql
--
-- Hoists the shared evidence block to the FRONT of all 17 council-gate seat
-- templates and inserts the cache breakpoint after it, so Anthropic prompt
-- caching (LCO-008) can actually fire.
--
-- WHY THIS IS THE WHOLE FIX AND THE GO CHANGE WAS ONLY HALF. Measured 24h to
-- 2026-08-10: council-gate burned 11,632,762 INPUT tokens against 370,298
-- output — 97% input, ~85% of all fleet LLM spend. Each of 17 seats receives
-- ~100k input, and the seats share their entire evidence body. But their
-- measured common prefix was ZERO characters, because each seat's persona sat
-- at the top and pushed the shared body to a different offset. Anthropic
-- caching is a PREFIX match, so with the old ordering the client-side seam
-- could never have cached anything no matter how it was configured.
--
-- WHAT MAKES THIS SAFE TO CACHE AT ALL: the seats run SEQUENTIALLY
-- (review_editquality -> review_constitution -> ... -> review_guardian ->
-- council_decide, per next_step). The documented fan-out hazard — N parallel
-- requests all miss because none can read a cache the others are still
-- writing — does not apply to a chain. Seat 1 writes; seats 2..17 read.
--
-- THE TRANSFORMATION IS A MOVE, NOT A REWRITE. The 172-char block
--     ## Schema (the ONLY tables available to checks)
--     {{.schema_hint.text}}
--     ## The author's stated rationale
--     {{.input_data.rationale}}
--     ## The plan
--     {{.plan_persisted.plan_json}}
-- is byte-identical across all 17 seats (md5 574d945d97706890d6595a0f24c9a38f,
-- verified before writing this) and appears exactly once in each. It is
-- removed from its mid-template position and prepended, followed by the
-- marker. Net length change is +25 chars per seat — exactly the marker plus
-- two newlines — which is the arithmetic proof that nothing else moved.
--
-- BEHAVIOURAL NOTE, STATED RATHER THAN BURIED: each seat now reads the plan
-- BEFORE its own persona and instructions, where previously it read them
-- after. For long context this is generally the better ordering (instructions
-- last are the ones best followed), but it is a real change to how 17
-- reviewers are prompted and it is why this ships separately from the client
-- seam rather than bundled with it.
--
-- IN-FLIGHT RUNS: SAFE, and this was MEASURED rather than assumed.
--
-- The cautious version of this banner said "do not apply while a council run is
-- in flight, because a run mid-chain may reload step config between steps". That
-- turned out to be wrong, and the check that settled it is one query:
--
--   SELECT orchestration_id,
--          (workflow_plan->'steps'->'review_mission'->'config' ? 'prompt_template'),
--          length(workflow_plan->'steps'->'review_mission'->'config'->>'prompt_template')
--   FROM orchestration_states WHERE status NOT IN ('COMPLETED','FAILED','CANCELLED');
--
-- Both council runs live at the time carried their OWN copy of every seat
-- template inside orchestration_states.workflow_plan (4,530 chars for
-- review_mission, matching the then-current template exactly). An orchestration
-- executes from that captured plan, so editing agent_definitions cannot reach a
-- run already under way: in-flight reviews finish on the old prompts, and only
-- orchestrations created after this migration pick up the new ones.
--
-- This is worth knowing well beyond this migration — it is the difference
-- between "config edits are safe on a busy shared cluster" and "every config
-- edit needs a quiet window", and on a tree this many sessions share, the second
-- would mean waiting for a window that may never come.
--
-- What it does NOT license: editing a seat template and expecting an in-flight
-- run to pick it up. It will not. If you are fixing a prompt because a review is
-- going wrong RIGHT NOW, that run is already committed to the old text.

BEGIN;

-- Backup first. The ROLLBACK file restores from this table rather than trying
-- to reverse the move, because each seat's ORIGINAL block position differs and
-- a reverse-transform would have to rediscover 17 of them. Restoring bytes is
-- exact; re-deriving them is a second chance to be wrong.
DROP TABLE IF EXISTS bak_council_gate_config_20260810;
CREATE TABLE bak_council_gate_config_20260810 AS
SELECT id, type, version, default_config, now() AS backed_up_at
FROM agent_definitions
WHERE type = 'council-gate'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

WITH src AS (
    SELECT d.id, s.key AS step_key, s.value AS step_val
    FROM agent_definitions d,
         jsonb_each(d.default_config->'workflow'->'steps') s
    WHERE d.type = 'council-gate'
      AND d.is_active AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
),
rewritten AS (
    SELECT id,
           jsonb_object_agg(
               step_key,
               CASE
                 WHEN step_key LIKE 'review\_%'
                  AND step_val->'config' ? 'prompt_template'
                  AND position('## Schema (the ONLY tables available to checks)'
                               in step_val->'config'->>'prompt_template') > 0
                 THEN jsonb_set(
                        step_val,
                        '{config,prompt_template}',
                        to_jsonb(
                          -- shared block, hoisted
                          substring(step_val->'config'->>'prompt_template'
                                    from position('## Schema (the ONLY tables available to checks)'
                                                  in step_val->'config'->>'prompt_template')
                                    for  position('{{.plan_persisted.plan_json}}'
                                                  in step_val->'config'->>'prompt_template')
                                       - position('## Schema (the ONLY tables available to checks)'
                                                  in step_val->'config'->>'prompt_template')
                                       + length('{{.plan_persisted.plan_json}}'))
                          || E'\n<!--CACHE_BREAKPOINT-->\n'
                          -- remainder, with the block removed from where it was
                          || replace(
                               step_val->'config'->>'prompt_template',
                               substring(step_val->'config'->>'prompt_template'
                                         from position('## Schema (the ONLY tables available to checks)'
                                                       in step_val->'config'->>'prompt_template')
                                         for  position('{{.plan_persisted.plan_json}}'
                                                       in step_val->'config'->>'prompt_template')
                                            - position('## Schema (the ONLY tables available to checks)'
                                                       in step_val->'config'->>'prompt_template')
                                            + length('{{.plan_persisted.plan_json}}')),
                               '')
                        )
                      )
                 ELSE step_val
               END
           ) AS new_steps
    FROM src
    GROUP BY id
)
UPDATE agent_definitions d
SET default_config = jsonb_set(d.default_config, '{workflow,steps}', r.new_steps),
    version        = d.version + 1,
    updated_at     = now()
FROM rewritten r
WHERE d.id = r.id;

-- Verify with DO/RAISE. A block of SELECTs cannot stop the COMMIT —
-- ON_ERROR_STOP does not treat a result set as an error — so a SELECT-based
-- "verification" commits green regardless of what it found.
DO $$
DECLARE
    n_seats      integer;
    n_marker     integer;
    n_prefix     integer;
    n_placeholders integer;
BEGIN
    SELECT count(*),
           count(*) FILTER (WHERE s.value->'config'->>'prompt_template' LIKE '%<!--CACHE_BREAKPOINT-->%'),
           count(*) FILTER (WHERE s.value->'config'->>'prompt_template' LIKE '{{.schema_hint.text}}%'
                                  OR s.value->'config'->>'prompt_template' LIKE '## Schema%'),
           count(*) FILTER (WHERE s.value->'config'->>'prompt_template' LIKE '%{{.schema_hint.text}}%'
                              AND s.value->'config'->>'prompt_template' LIKE '%{{.input_data.rationale}}%'
                              AND s.value->'config'->>'prompt_template' LIKE '%{{.plan_persisted.plan_json}}%')
      INTO n_seats, n_marker, n_prefix, n_placeholders
    FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
    WHERE d.type='council-gate' AND d.is_active
      AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
      AND s.key LIKE 'review\_%';

    IF n_seats <> 17 THEN
        RAISE EXCEPTION 'MIGRATION 377: expected 17 review seats, found %', n_seats;
    END IF;
    IF n_marker <> n_seats THEN
        RAISE EXCEPTION 'MIGRATION 377: only %/% seats carry the cache marker', n_marker, n_seats;
    END IF;
    IF n_prefix <> n_seats THEN
        RAISE EXCEPTION 'MIGRATION 377: only %/% seats START with the shared block — a seat whose persona still leads caches NOTHING', n_prefix, n_seats;
    END IF;
    IF n_placeholders <> n_seats THEN
        RAISE EXCEPTION 'MIGRATION 377: only %/% seats retain all three placeholders — the move dropped content', n_placeholders, n_seats;
    END IF;

    RAISE NOTICE 'migration 377 OK: %/% seats hoisted, marked, and intact', n_seats, n_seats;
END $$;

-- THE ONE THAT MATTERS MOST, separated so its failure message is unambiguous:
-- every seat's cacheable prefix must be byte-identical, or each writes its own
-- entry and reads none — strictly worse than no caching, and silent.
DO $$
DECLARE
    n_distinct integer;
BEGIN
    SELECT count(DISTINCT md5(split_part(s.value->'config'->>'prompt_template', '<!--CACHE_BREAKPOINT-->', 1)))
      INTO n_distinct
    FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
    WHERE d.type='council-gate' AND d.is_active
      AND COALESCE(d.is_snapshot,false)=false AND d.deleted_at IS NULL
      AND s.key LIKE 'review\_%';

    IF n_distinct <> 1 THEN
        RAISE EXCEPTION
          'MIGRATION 377: % distinct cacheable prefixes across the seats (must be exactly 1). Every seat would write its own cache entry and read none — worse than no caching, and it would look like success.', n_distinct;
    END IF;
    RAISE NOTICE 'migration 377 OK: all seats share exactly 1 byte-identical cacheable prefix';
END $$;

COMMIT;
