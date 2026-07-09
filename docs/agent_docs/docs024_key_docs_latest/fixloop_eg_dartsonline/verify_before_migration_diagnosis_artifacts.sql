-- verify_before_migration_diagnosis_artifacts.sql — run on the live clients_db
-- BEFORE applying 0NN_diagnosis_artifacts.sql. 2026-07-09.
-- Follows the travelling-docs precedent (verify_before_migration.sql).
-- Three gates. Paste results back into NOTES_running_fixloop(10).md.

-- 1. Name collision: expect ZERO rows.
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name = 'diagnosis_artifacts';

-- 2. Index-name collision: expect ZERO rows. (Names are global per schema, so a
--    clean table check alone is not sufficient.)
SELECT indexname
FROM pg_indexes
WHERE schemaname = 'public'
  AND indexname IN (
    'idx_diagnosis_artifacts_correlation',
    'idx_diagnosis_artifacts_bundle_current',
    'idx_diagnosis_artifacts_expiry'
  );

-- 3. Convention check — the pipeline namespace F0.1c will use. `pipeline` is
--    unconstrained text NOT NULL DEFAULT 'build' on site_work_items (no CHECK),
--    so 'diagnose' needs no schema change; confirm no CHECK has appeared since,
--    and see whether 'diagnose' is already in live use.
SELECT pipeline, count(*) AS items, min(created_at) AS first_seen, max(created_at) AS last_seen
FROM site_work_items
GROUP BY pipeline
ORDER BY items DESC;

SELECT conname, pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conrelid = 'site_work_items'::regclass
  AND contype = 'c';
