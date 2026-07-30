-- 273 — allow subject_type='component' on doc_plans and doc_notes
--
-- WHY THIS FILE EXISTS: features_open/027 / docs024_key_docs_latest/staged_component_build.
-- A section component needs the same travelling docs a tool already gets (TL-017): a PLAN
-- carrying its criteria fence, and append-only NOTES carrying every repair. Today it cannot
-- have either — both CHECK constraints refuse 'component', so the insert fails outright.
--
-- THE ORDERING RULE, AND WHY IT IS NOT OPTIONAL. This migration is HALF the change. The
-- other half is `validDocSubjectTypes` in platform/orchestration/actions/doc_subjects_common.go,
-- which gates every doc action (write_doc_plan, append_doc_note, load_doc_context,
-- persist_diagnosis_note). Applying this file against an image whose Go list still lacks
-- 'component' recreates migration 184's split contract — the DB accepts what every Go gate
-- refuses — which is bugs_open/064, filed because 184 did exactly this and left its own
-- seeded action docs unreachable. So: IMAGE FIRST, THEN THIS MIGRATION.
--
-- Mitigating fact, stated so nobody over-reacts to an early apply: nothing writes
-- subject_type='component' yet, so a widened CHECK ahead of the image is INERT rather than
-- broken. The split only bites when a producer exists, and the first producer is P1 step 2
-- of the lane's PLAN. This is why the file is numbered normally rather than withheld.
--
-- LOCKSTEP: TestValidDocSubjectTypes_LockstepWithMigrationCheck parses the NEWEST numbered
-- migration that re-creates doc_plans_subject_type_check and fails the build if its ARRAY
-- does not equal validDocSubjectTypes. That is why this file and the Go edit are in ONE
-- commit — landing either alone reddens HEAD for every other session on this shared tree.
--
-- LANDMINE THIS FILE PAYS FOR ONCE: the two tables do NOT carry the same vocabulary.
-- doc_notes also allows 'landmine' (migration 270, 07-30). Re-adding doc_notes' constraint
-- from doc_plans' array — the obvious way to make them "agree" — would drop 'landmine' and
-- orphan the live landmine corpus (57 rows when measured 2026-07-30, written by other
-- threads and synced from LANDMINES.md by scripts/landmines-sync.py). Read the constraint
-- you are replacing; the tables are deliberately not identical.
--
-- SCOPE — additive and inert. No existing row changes; no existing query changes meaning; a
-- consumer filtering on subject_type cannot see the new value until something writes it.
-- Under the owner ruling of 2026-07-29 §1 that makes this normal council-gate scope and NOT
-- an RFC: it adds an opt-in capability rather than changing what doc_plans/doc_notes
-- GUARANTEE to anyone already using them.
--
-- RE-RUN SAFE: each half guards, then drops by name and re-adds the whole definition, so a
-- second apply converges rather than accumulating.

BEGIN;

-- Guard both constraints before touching either, so a half-applied state is impossible.
DO $$
DECLARE
  plans_def text;
  notes_def text;
BEGIN
  SELECT pg_get_constraintdef(oid) INTO plans_def
    FROM pg_constraint
   WHERE conrelid = 'public.doc_plans'::regclass
     AND conname  = 'doc_plans_subject_type_check';

  SELECT pg_get_constraintdef(oid) INTO notes_def
    FROM pg_constraint
   WHERE conrelid = 'public.doc_notes'::regclass
     AND conname  = 'doc_notes_subject_type_check';

  IF plans_def IS NULL OR notes_def IS NULL THEN
    RAISE EXCEPTION 'a doc subject_type CHECK is absent — the tables are not in the shape migration 273 was written against; read them before applying';
  END IF;

  IF plans_def LIKE '%component%' AND notes_def LIKE '%component%' THEN
    RAISE EXCEPTION 'both CHECKs already allow component — migration 273 is already applied; record it with --record-only rather than re-running';
  END IF;

  -- 270 added 'landmine' to doc_notes ONLY. If that is missing, this migration would
  -- silently narrow doc_notes and orphan the landmine corpus. Refuse instead.
  IF notes_def NOT LIKE '%landmine%' THEN
    RAISE EXCEPTION 'doc_notes_subject_type_check does not allow landmine — migration 270 has not been applied here, and 273 assumes it; apply 270 first or this file drops a value it never added';
  END IF;
END $$;

ALTER TABLE public.doc_plans
  DROP CONSTRAINT doc_plans_subject_type_check;

ALTER TABLE public.doc_plans
  ADD CONSTRAINT doc_plans_subject_type_check
  CHECK (subject_type = ANY (ARRAY[
    'tool'::text,
    'pipeline'::text,
    'experience'::text,
    'action'::text,
    'experience-pattern'::text,
    'component'::text
  ]));

-- doc_notes keeps 'landmine' — see the landmine note in the header.
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
    'landmine'::text,
    'component'::text
  ]));

COMMIT;
