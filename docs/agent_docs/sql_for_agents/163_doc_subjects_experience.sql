-- 163_doc_subjects_experience.sql — allow subject_type='experience' on the
-- travelling docs (Experience Loop, Phase 1 / RUNBOOK T1.1).
--
-- doc_plans/doc_notes CHECK constraints allow only ('tool','pipeline'); the
-- Experience Loop stores EXPERIENCE_PLANs as doc_plans rows with
-- subject_type='experience' (PLAN_experience_loop.md §3-A). Additive only:
-- existing rows untouched, no code change required (the doc actions pass
-- subject_type through from step config).
--
-- Subject-key conventions for 'experience' (RUNBOOK T1.2):
--   * subject_key = kebab experience name, e.g. 'vonc-spark-game'
--   * drill/test subjects use suffix '-drill' and are superseded + their work
--     items cancelled after use (RUNBOOK T5.4)
--
-- Applied out of band (psql -f + ledger row same sitting per
-- bugs_open/aaa_fails_to_mend/007).

BEGIN;

ALTER TABLE doc_plans DROP CONSTRAINT doc_plans_subject_type_check;
ALTER TABLE doc_plans ADD CONSTRAINT doc_plans_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text, 'pipeline'::text, 'experience'::text]));

ALTER TABLE doc_notes DROP CONSTRAINT doc_notes_subject_type_check;
ALTER TABLE doc_notes ADD CONSTRAINT doc_notes_subject_type_check
  CHECK (subject_type = ANY (ARRAY['tool'::text, 'pipeline'::text, 'experience'::text]));

INSERT INTO doc_notes (id, subject_type, subject_key, body, categories, source, created_by)
VALUES (
  gen_random_uuid(),
  'pipeline', 'build',
  '## Travelling docs now accept subject_type=experience
Observed: the Experience Loop needs EXPERIENCE_PLAN rows (journeys, promise ledger, data contracts, MVP cut, journey criteria) in doc_plans, but both tables'' CHECK constraints allowed only tool|pipeline.
Fix: constraints extended to tool|pipeline|experience (migration 163). Additive; no code change; the partial-unique one-current-per-subject index is type-agnostic and needs nothing.
Conventions: subject_key = kebab experience name (pilot: vonc-spark-game); drills use suffix -drill and are superseded after use.
Categories: fix',
  '["fix"]'::jsonb,
  'migration', '163_doc_subjects_experience'
);

COMMIT;
