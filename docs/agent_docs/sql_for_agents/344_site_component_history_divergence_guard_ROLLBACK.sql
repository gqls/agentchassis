-- 344_site_component_history_divergence_guard_ROLLBACK.sql
-- Rolls back the ACTIVE half of mig 344 only: the trigger and its function.
--
-- The table (site_component_history) and column
-- (site_components.rendered_html_digest) are KEPT deliberately — the table
-- holds archived artefacts, and a rollback that dropped it would itself be
-- the silent loss bugs_open/226 exists to end. The column is inert without
-- the trigger and the Go half. If removal is truly wanted after the archived
-- rows have been reviewed:
--   DROP TABLE site_component_history;
--   ALTER TABLE site_components DROP COLUMN rendered_html_digest;
-- — by hand, eyes open, never from this file.
--
-- After this rollback, chrome overwrites behave exactly as before mig 344
-- (unconditional replace, no archive). The Go half's digest stamp writes a
-- value into a column nothing reads, and its classification SELECT reads
-- NULLs — both harmless.

BEGIN;

DROP TRIGGER IF EXISTS trg_site_component_archive ON site_components;
DROP FUNCTION IF EXISTS site_component_history_archive();

COMMIT;
