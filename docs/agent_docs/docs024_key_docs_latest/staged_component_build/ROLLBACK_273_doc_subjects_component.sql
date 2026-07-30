-- ROLLBACK for migration 273 — remove 'component' from both doc subject_type CHECKs.
--
-- WHY THIS FILE EXISTS: the council gate's debug_historian seat objected that 273 shipped
-- as a single forward-apply file with no separate verify and no rollback artifact,
-- contrary to the house discipline of keeping migration / verify / rollback as distinct
-- artifacts for production schema surgery. Correct. This is the rollback; the verify is
-- VERIFY_273_before_apply.sh beside it.
--
-- ⚠ NOT NUMBERED AND NOT IN sql_for_agents/ — DELIBERATELY, AND DO NOT MOVE IT THERE.
-- The migration runner applies EVERY pending .sql in that directory. A numbered rollback
-- sitting next to its own forward migration is a loaded gun: one `--apply` and the runner
-- applies the change and then reverts it, or reverts it long after something depends on it.
-- Run this by hand, deliberately:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < ROLLBACK_273_doc_subjects_component.sql
--
-- PRECONDITION THIS FILE ENFORCES RATHER THAN TRUSTS: narrowing a CHECK fails if any row
-- violates it, so a rollback while component docs exist would abort mid-transaction. The
-- guard below refuses FIRST, with a count and a clear reason, instead of letting Postgres
-- produce a constraint-violation error that reads like a corrupt table.
--
-- IT ALSO PRESERVES 'landmine' on doc_notes. The two tables do not carry the same
-- vocabulary (270 added 'landmine' to doc_notes only); rebuilding doc_notes' constraint
-- from doc_plans' array would orphan the live landmine corpus. Same landmine as 273's.

BEGIN;

DO $$
DECLARE
  plan_rows  bigint;
  note_rows  bigint;
  notes_def  text;
BEGIN
  SELECT count(*) INTO plan_rows FROM doc_plans WHERE subject_type = 'component';
  SELECT count(*) INTO note_rows FROM doc_notes WHERE subject_type = 'component';

  IF plan_rows > 0 OR note_rows > 0 THEN
    RAISE EXCEPTION
      'refusing to roll back: % doc_plans and % doc_notes rows still carry subject_type=''component''. Narrowing the CHECK would abort anyway. Decide what happens to those rows FIRST — they are somebody''s travelling docs, not debris.',
      plan_rows, note_rows;
  END IF;

  SELECT pg_get_constraintdef(oid) INTO notes_def
    FROM pg_constraint
   WHERE conrelid = 'public.doc_notes'::regclass
     AND conname  = 'doc_notes_subject_type_check';

  IF notes_def IS NULL THEN
    RAISE EXCEPTION 'doc_notes_subject_type_check is absent — the table is not in the shape this rollback was written against';
  END IF;

  IF notes_def NOT LIKE '%landmine%' THEN
    RAISE EXCEPTION 'doc_notes_subject_type_check does not allow landmine — this rollback would not restore it either; read the constraint before proceeding';
  END IF;
END $$;

-- Restore both constraints to their pre-273 definitions: doc_plans without 'component',
-- doc_notes without 'component' but WITH 'landmine' (its post-270 state).
ALTER TABLE public.doc_plans
  DROP CONSTRAINT doc_plans_subject_type_check;

ALTER TABLE public.doc_plans
  ADD CONSTRAINT doc_plans_subject_type_check
  CHECK (subject_type = ANY (ARRAY[
    'tool'::text,
    'pipeline'::text,
    'experience'::text,
    'action'::text,
    'experience-pattern'::text
  ]));

ALTER TABLE public.doc_notes
  DROP CONSTRAINT doc_notes_subject_type_check;

ALTER TABLE public.doc_notes
  ADD CONSTRAINT doc_notes_subject_type_check
  CHECK (subject_type = ANY (ARRAY[
    'tool'::text,
    'pipeline'::text,
    'experience'::text,
    'action'::text,
    'experience-pattern'::text,
    'landmine'::text
  ]));

COMMIT;

-- AFTER ROLLING BACK, the Go half must come out too, or you have inverted the split:
-- the Go gate would accept 'component' while the DB refuses it. Remove "component" from
-- validDocSubjectTypes in platform/orchestration/actions/doc_subjects_common.go and roll
-- an image. TestValidDocSubjectTypes_LockstepWithMigrationCheck will fail until you also
-- remove 273 from sql_for_agents/ (it reads the newest numbered migration), which is the
-- lockstep working as designed.
