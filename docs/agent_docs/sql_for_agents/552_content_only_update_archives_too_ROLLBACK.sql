-- FILE: docs/agent_docs/sql_for_agents/552_content_only_update_archives_too_ROLLBACK.sql
--
-- Revert bugs_open/355's archive widening: drop the content-only UPDATE
-- trigger. Run BY HAND — the runner never applies a ROLLBACK sidecar.
--
-- Archived rows are KEPT, per the 344/357 precedent: a rollback must not become
-- the loss it guards against. After this, content_data-only UPDATEs are once
-- again invisible to page_component_history (357's rendered_html-gated triggers
-- still stand), and the content-loss-check's coverage shrinks accordingly — its
-- next heartbeat doc_note will show the overwrite-pair volume falling, which is
-- the honest signal that this arm went dark.

\set ON_ERROR_STOP on

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'page_components'::regclass
                     AND tgname = 'trg_page_component_content_archive_upd') THEN
        RAISE EXCEPTION 'mig552 ROLLBACK: trigger not present — 552 is not applied';
    END IF;
END $$;

DROP TRIGGER trg_page_component_content_archive_upd ON page_components;

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'page_components'::regclass
                 AND tgname = 'trg_page_component_content_archive_upd') THEN
        RAISE EXCEPTION 'mig552 ROLLBACK verify FAILED: trigger still present';
    END IF;
    RAISE NOTICE 'mig552 ROLLBACK: content-only updates no longer archive (357''s triggers untouched)';
END $$;

COMMIT;
