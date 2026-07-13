-- verify_before_migration.sql — run on the live clients_db BEFORE applying the
-- doc_plans/doc_notes migration. 2026-07-04.
-- Two gates + one convention check. Paste results back.

-- 1. Name collision: expect zero rows.
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN ('doc_plans', 'doc_notes');

-- 2. Pipeline subject keys: the dump shows site_work_items.pipeline is
--    unconstrained text NOT NULL DEFAULT 'build' (no CHECK), so the live value
--    set is convention, not schema. These are the values doc subjects will use.
SELECT pipeline, count(*) AS items, min(created_at) AS first_seen, max(created_at) AS last_seen
FROM site_work_items
GROUP BY pipeline
ORDER BY items DESC;

-- 3. (Convention check) Confirm no CHECK constraint has appeared since the dump:
SELECT conname, pg_get_constraintdef(oid)
FROM pg_constraint
WHERE conrelid = 'site_work_items'::regclass
  AND contype = 'c';
