-- 357_page_component_artefact_archive_ROLLBACK.sql
--
-- Removes the bugs_open/229 page-side archive TRIGGERS + FUNCTION only.
-- Columns on page_component_history / page_components and every archived row
-- are KEPT — they hold recovered artefacts, and a rollback must not become
-- the loss it guards against (mig 344 precedent). The Go digest stamp keeps
-- writing harmlessly into the kept column if the image outlives the rollback.

BEGIN;

DROP TRIGGER IF EXISTS trg_page_component_artefact_archive_upd ON page_components;
DROP TRIGGER IF EXISTS trg_page_component_artefact_archive_del ON page_components;
DROP FUNCTION IF EXISTS page_component_artefact_archive();

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM pg_trigger
    WHERE tgrelid = 'page_components'::regclass
      AND tgname LIKE 'trg_page_component_artefact_archive%';
    IF n <> 0 THEN
        RAISE EXCEPTION 'rollback guard: % archive trigger(s) still present', n;
    END IF;
END $$;

COMMIT;
