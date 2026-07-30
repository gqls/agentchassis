-- 271 — the council gate's schema_hint carries the landmine index
--
-- D10(c), owner ruling 2026-07-29. The proposal (PROPOSAL_D9_landmines_as_a_
-- footprinted_corpus.md §4c) asked for "add doc_notes to the schema hint" so the
-- seats can reach the prose corpus that D8 established is otherwise invisible to
-- every one of them.
--
-- WHY THIS IS ONE STEP AND NOT SEVENTEEN. `load_schema_hint` is a single
-- query_database step whose output field `schema_hint` is ALREADY listed in every
-- review seat's `input_fields`. Changing this one query therefore reaches all
-- seats without touching a single seat config — which matters, because CLAUDE.md
-- forbids hand-patching the gate roster (099_SYNC_gate_roster.py owns it, and two
-- hand-maintained rosters is the exact drift class this council reviews for).
--
-- WHAT CHANGES
--   1. `doc_notes` joins the ten tables whose columns are listed, so a seat
--      writing a check against it stops hallucinating columns — the stated
--      purpose of this step.
--   2. A compact LANDMINE INDEX is appended: one line per subject_key, with the
--      query to pull the full body. Index, not bodies — see the size note.
--
-- MEASURED COST, not estimated (2026-07-29): the hint goes from 5,855 to ~11,413
-- bytes, i.e. +5,558, and it is injected into ~17 seat prompts per submission.
-- Acceptable at 16 landmines. NOT acceptable at 100.
--
-- THE FOLLOW-UP, PROVEN IN PRINCIPLE AND DELIBERATELY NOT SHIPPED HERE:
-- relevance-gating. Binding the submission as a param and filtering
-- `WHERE $1::text LIKE '%' || subject_key || '%'` was tested read-only against
-- the most recent real submission and cut 16 landmines to **2** — and both were
-- genuinely on point (it touched palette code, and `palette_specialised_slots.go`
-- carries a landmine). It is not shipped because this step has NO `error_step`:
-- if a new param path failed to resolve, load_schema_hint would fail and take
-- EVERY council submission with it. Whether an unresolvable param binds NULL or
-- raises is [UNVERIFIED], and a blast radius of "all reviews" is not the place to
-- find out. Ship it once that is checked, with a length floor of >= 6 on
-- subject_key (`Bash` and `cmd/` are 4 and would match prose).
--
-- SCOPE: additive and inert in the sense that matters — no seat's prompt,
-- roster, footprint or decision rule changes, and a seat that ignores the new
-- section behaves exactly as before. It does change what every seat SEES, so it
-- is recorded here and in register DOC-067 rather than treated as a config tweak.
--
-- RE-RUN SAFE: guarded on the marker text; raises if already applied.

\set ON_ERROR_STOP on

BEGIN;

-- Snapshot first. NOTE the two-arg form writes to agent_definitions_backup, NOT
-- an is_snapshot row in agent_definitions — see the `snapshot_agent` landmine.
SELECT snapshot_agent('council-gate', '271_schema_hint_landmines');

DO $guard$
DECLARE
  q text;
BEGIN
  SELECT default_config->'workflow'->'steps'->'load_schema_hint'->'config'->>'query'
    INTO q
    FROM agent_definitions
   WHERE type = 'council-gate' AND is_active AND deleted_at IS NULL
     AND COALESCE(is_snapshot, false) = false;

  IF q IS NULL THEN
    RAISE EXCEPTION 'council-gate has no load_schema_hint step — the workflow has changed since migration 271 was written; read it before applying';
  END IF;

  IF q LIKE '%LANDMINES%' THEN
    RAISE EXCEPTION 'load_schema_hint already carries the landmine index — already applied; use --record-only';
  END IF;

  IF q NOT LIKE '%information_schema.columns%' THEN
    RAISE EXCEPTION 'load_schema_hint is not the information_schema query this migration expects — read it before applying';
  END IF;
END
$guard$;

UPDATE agent_definitions
SET default_config = jsonb_set(
      jsonb_set(
        default_config,
        '{workflow,steps,load_schema_hint,config,query}',
        to_jsonb($q$
SELECT string_agg(t.line, chr(10)) AS text FROM (
  SELECT 0 AS ord,
         table_name || '(' || string_agg(column_name || ' ' || data_type, ', ' ORDER BY ordinal_position) || ')' AS line
    FROM information_schema.columns
   WHERE table_schema = 'public'
     AND table_name IN ('pages','sites','site_plans','site_plan_pages','site_work_items',
                        'content_components','page_components','agent_definitions',
                        'diagnosis_artifacts','agent_error_log','doc_notes')
   GROUP BY table_name
  UNION ALL
  SELECT 1, chr(10) || 'LANDMINES -- traps that mislead BEFORE any symptom, keyed by the thing they guard. If this change touches one, read it before judging: SELECT body FROM doc_notes WHERE subject_key = $$<key>$$ AND categories ? $$landmine$$'
  UNION ALL
  SELECT 2, '  ' || subject_key || ' -- ' || left(replace(split_part(body, chr(10), 1), chr(10), ' '), 110)
    FROM (SELECT DISTINCT ON (subject_key) subject_key, body
            FROM doc_notes
           WHERE categories ? 'landmine'
           ORDER BY subject_key, created_at DESC) d
) t
$q$::text),
        true),
      '{workflow,steps,load_schema_hint,description}',
      to_jsonb('Live table/column list (information_schema) so reviewer checks stop hallucinating columns, PLUS the landmine index from doc_notes (categories ? landmine) so a seat can see that the thing being changed carries a known trap. Index only — pull the body with the query in the header line.'::text),
      true),
    updated_at = now()
WHERE type = 'council-gate' AND is_active AND deleted_at IS NULL
  AND COALESCE(is_snapshot, false) = false;

COMMIT;
