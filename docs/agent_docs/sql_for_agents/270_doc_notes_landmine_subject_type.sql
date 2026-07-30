-- 270 — allow subject_type='landmine' on doc_notes
--
-- WHY THIS FILE EXISTS: owner ruling D10 (2026-07-29) made
-- docs024_key_docs_latest/LANDMINES.md the system of record for landmines, synced
-- into doc_notes by scripts/landmines-sync.py so council seats and agents can read
-- them. The sync needs a subject_type for rows whose footprint is a FILE, a TABLE,
-- a COMMAND or a service — none of which is a tool, pipeline, experience or action.
--
-- The proposal this implements (architecture_review/PROPOSAL_D9_landmines_as_a_
-- footprinted_corpus.md, §3) states "No migration needed; the table, both indexes
-- and the practice already exist". That is true of the CATEGORY and false of the
-- SUBJECT TYPE: doc_notes_subject_type_check restricts subject_type to
-- tool|pipeline|experience|action|experience-pattern. Measured 2026-07-29 — and it
-- is precisely why the 7 landmine rows written 07-27/28 by two other threads are
-- subject_type='action': the constraint left them no choice, so a landmine about
-- `cmd/` or `postgres-clients-0` had nowhere honest to go.
--
-- SCOPE — additive and inert. Nothing reads subject_type='landmine' until the sync
-- writes it; no existing row changes; no existing query changes meaning. Under the
-- owner ruling of 2026-07-29 §1 that makes this normal council-gate scope, not an
-- RFC: it adds an opt-in capability, it does not change what doc_notes GUARANTEES
-- to anyone already using it.
--
-- LANDMINE FOR WHOEVER QUERIES THESE ROWS: filter on `categories ? 'landmine'`,
-- NOT on subject_type='landmine'. The 7 pre-existing rows are subject_type='action'
-- and they are real landmines; typing them differently would be rewriting another
-- thread's records to suit this migration. The CATEGORY is the stable interface,
-- the subject_type is only accurate typing for new rows.
--
-- RE-RUN SAFE: drops the constraint by name only if present, then re-adds it. The
-- ADD is the whole definition, so applying twice converges rather than accumulating.

BEGIN;

-- Guard: fail loudly if the constraint is not the shape this migration expects,
-- rather than silently replacing something a later migration changed underneath us.
DO $$
DECLARE
  found_def text;
BEGIN
  SELECT pg_get_constraintdef(oid) INTO found_def
    FROM pg_constraint
   WHERE conrelid = 'public.doc_notes'::regclass
     AND conname  = 'doc_notes_subject_type_check';

  IF found_def IS NULL THEN
    RAISE EXCEPTION 'doc_notes_subject_type_check is absent — doc_notes is not in the shape migration 270 was written against; read the table before applying';
  END IF;

  IF found_def LIKE '%landmine%' THEN
    RAISE EXCEPTION 'doc_notes_subject_type_check already allows landmine — this migration is already applied; record it with --record-only rather than re-running';
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
    'landmine'::text
  ]));

COMMIT;
