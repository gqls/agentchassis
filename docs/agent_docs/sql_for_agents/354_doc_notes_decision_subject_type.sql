-- 354 — allow subject_type='decision' on doc_notes (RFC_015 stage 1)
--
-- WHY: RFC_015 (decision records — allow change, forbid regression, owner-
-- approved 2026-08-08) types decisions as doc_notes rows. The first four
-- (idea.uk, D-001..D-004, source='rfc015-staging') are subject_type=
-- 'component' because this constraint left them no honest slot — same
-- story, same shape, as migration 270's landmine rows.
--
-- SCOPE — additive and inert (2026-07-29 ruling: normal council gate, not an
-- RFC of its own). Nothing writes subject_type='decision' until operators or
-- agents type new decisions; no existing row or query changes meaning.
--
-- LANDMINE FOR READERS (same as 270's): filter on `categories ? 'decision'`,
-- NEVER on subject_type='decision' — the four staged rows are typed
-- 'component' and are real decisions. The CATEGORY is the stable interface;
-- both the citation gate (decision_guard.go) and the guards check
-- (check_decision_guards.go) already filter by category only.
--
-- RE-RUN SAFE: guard fails loudly if already applied or the shape drifted.

BEGIN;

DO $$
DECLARE
  found_def text;
BEGIN
  SELECT pg_get_constraintdef(oid) INTO found_def
    FROM pg_constraint
   WHERE conrelid = 'public.doc_notes'::regclass
     AND conname  = 'doc_notes_subject_type_check';

  IF found_def IS NULL THEN
    RAISE EXCEPTION 'doc_notes_subject_type_check absent — table not in the shape 354 expects; read it before applying';
  END IF;

  IF found_def LIKE '%decision%' THEN
    RAISE EXCEPTION 'doc_notes_subject_type_check already allows decision — 354 already applied; record with --record-only';
  END IF;

  IF found_def NOT LIKE '%component%' OR found_def NOT LIKE '%landmine%' THEN
    RAISE EXCEPTION 'constraint lacks component/landmine — vocabulary drifted from what 354 was written against';
  END IF;
END $$;

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
    'component'::text,
    'decision'::text
  ]));

COMMIT;
