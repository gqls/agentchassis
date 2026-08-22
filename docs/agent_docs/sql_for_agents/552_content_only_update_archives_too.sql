-- FILE: docs/agent_docs/sql_for_agents/552_content_only_update_archives_too.sql
--
-- bugs_open/355 (owner directive 2026-08-22: ship the content-loss detector AND
-- its reader) — close the archive's content-only blind spot first, because the
-- detector reads the archive and inherits every hole in it.
--
-- THE HOLE, in one sentence: migration 357's UPDATE trigger is
-- `AFTER UPDATE OF rendered_html ... WHEN (NEW.rendered_html IS DISTINCT FROM
-- OLD.rendered_html)`, so an UPDATE that changes `content_data` WITHOUT touching
-- the artefact archives nothing — the pre-image of the very column the
-- content-loss class is about is the one change the archive cannot see.
--
-- WHO ACTUALLY WRITES THAT SHAPE, measured 2026-08-22 at HEAD: every Go
-- in-place writer sets rendered_html in the same UPDATE (adopt_verbatim :512,
-- create_report_page :224, rebuild_blog_listing :336, section_editor :1472/:1482/
-- :1507, webdesignport import :217) — EXCEPT the admin handler
-- (internal/core-manager/admin/page_admin_handlers.go :326-361), whose SET list
-- is built dynamically and can be content_data-only. Plus any raw psql, and any
-- future writer. Narrow today; structural tomorrow — and the point of the
-- detector is precisely the writers nobody is thinking about.
--
-- WHAT THIS ADDS: a second UPDATE trigger on the SAME function, firing exactly
-- when the first one cannot:
--   * content_data actually changed, AND
--   * rendered_html did NOT change in the same statement (if it did, 357's
--     trigger archives the row already — the WHEN clauses are mutually
--     exclusive on that predicate, so no UPDATE can archive twice), AND
--   * the outgoing content_data held something (archiving '{}' pre-images is
--     noise; a row gaining its first keys has lost nothing).
--
-- DELIBERATE REDUNDANCY, stated so it is not "fixed" later: a content-only
-- archive row carries OLD.rendered_html, which on this arm equals the bytes
-- still live on the row. Sharing 357's function unchanged was chosen over a
-- parameterised variant because the archive function is fail-closed
-- (RFC_017 posture: a broken history table must halt page writes loudly), and a
-- second code path through it is a second thing that can halt them. Volume on
-- this arm today is admin edits only; the content-loss-check's daily heartbeat
-- reports pair volumes, so growth is watched, not assumed.
--
-- FAIL-CLOSED, like 357: the INSERT failing aborts the triggering UPDATE.
-- Rollback: 552_content_only_update_archives_too_ROLLBACK.sql (drops the
-- trigger only; archived rows are KEPT — a rollback must not become the loss
-- it guards against).

\set ON_ERROR_STOP on

BEGIN;

-- Pre-flight: the function and the 357 triggers this composes with must exist,
-- and this trigger must not already exist (idempotence guard).
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'page_component_artefact_archive') THEN
        RAISE EXCEPTION 'mig552: page_component_artefact_archive() missing — migration 357 is not applied; this file extends it and cannot precede it';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'page_components'::regclass
                     AND tgname = 'trg_page_component_artefact_archive_upd') THEN
        RAISE EXCEPTION 'mig552: 357''s UPDATE trigger is missing — the mutual-exclusion argument below assumes it exists';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgrelid = 'page_components'::regclass
                 AND tgname = 'trg_page_component_content_archive_upd') THEN
        RAISE EXCEPTION 'mig552: already applied (trg_page_component_content_archive_upd exists)';
    END IF;
END $$;

CREATE TRIGGER trg_page_component_content_archive_upd
AFTER UPDATE OF content_data ON page_components
FOR EACH ROW
WHEN (OLD.content_data IS NOT NULL AND OLD.content_data <> '{}'::jsonb
      AND NEW.content_data IS DISTINCT FROM OLD.content_data
      AND NEW.rendered_html IS NOT DISTINCT FROM OLD.rendered_html)
EXECUTE FUNCTION page_component_artefact_archive();

-- ---------------------------------------------------------------------------
-- VERIFY — induced probe, DO/RAISE, in 357's own style: every arm exercised on
-- a THROWAWAY row, each check selecting its archive row BY ITS BYTES (357's
-- probe failed its first apply on ORDER BY created_at — same xact, one
-- timestamp). Probe self-cleans.
-- ---------------------------------------------------------------------------
DO $probe$
DECLARE
    v_page_id uuid;
    v_pc_id   uuid;
    n         int;
    v_op      text;
    v_html    text;
BEGIN
    SELECT id INTO v_page_id FROM pages ORDER BY created_at LIMIT 1;
    IF v_page_id IS NULL THEN
        RAISE EXCEPTION 'mig552 probe: no pages row to hang the probe on';
    END IF;

    INSERT INTO page_components (page_id, position, slot_name, rendered_html, content_data, build_status)
    VALUES (v_page_id, 9998, 'mig552_probe', 'PROBE-HTML', '{"k":"v1"}'::jsonb, 'pending')
    RETURNING id INTO v_pc_id;

    -- 1. Content-only change on a non-empty pre-image: MUST archive, once,
    --    op=overwrite, carrying the pre-image content and the (unchanged) html.
    UPDATE page_components SET content_data = '{"k":"v2"}'::jsonb WHERE id = v_pc_id;
    SELECT count(*) INTO n FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe';
    IF n <> 1 THEN
        RAISE EXCEPTION 'mig552 probe: content-only update archived % row(s), expected 1', n;
    END IF;
    SELECT op, rendered_html INTO v_op, v_html FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe'
       AND content_data = '{"k":"v1"}'::jsonb;
    IF v_op IS DISTINCT FROM 'overwrite' OR v_html IS DISTINCT FROM 'PROBE-HTML' THEN
        RAISE EXCEPTION 'mig552 probe: content-only archive row is (op %, html %) — expected (overwrite, PROBE-HTML)', v_op, v_html;
    END IF;

    -- 2. No-op content touch: MUST NOT archive (IS DISTINCT gate).
    UPDATE page_components SET content_data = content_data WHERE id = v_pc_id;
    SELECT count(*) INTO n FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe';
    IF n <> 1 THEN
        RAISE EXCEPTION 'mig552 probe: no-op content touch archived (now % rows) — WHEN gate wrong', n;
    END IF;

    -- 3. Both change in one statement: exactly ONE archive row (357''s trigger
    --    takes it; this one must stand down — the mutual-exclusion check).
    UPDATE page_components SET rendered_html = 'PROBE-HTML-2', content_data = '{"k":"v3"}'::jsonb
     WHERE id = v_pc_id;
    SELECT count(*) INTO n FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe';
    IF n <> 2 THEN
        RAISE EXCEPTION 'mig552 probe: both-columns update produced % total row(s), expected 2 (no double-archive)', n;
    END IF;
    -- and the row it produced came from 357''s arm: pre-image html, v2 content.
    SELECT count(*) INTO n FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe'
       AND rendered_html = 'PROBE-HTML' AND content_data = '{"k":"v2"}'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION 'mig552 probe: both-columns archive row not found by its bytes — wrong arm fired';
    END IF;

    -- 4. Empty pre-image gains keys: MUST NOT archive ('{}' gate).
    UPDATE page_components SET content_data = '{}'::jsonb WHERE id = v_pc_id;  -- archives v3 (3rd row, legitimate)
    UPDATE page_components SET content_data = '{"k":"v4"}'::jsonb WHERE id = v_pc_id;  -- pre-image {} — must NOT archive
    SELECT count(*) INTO n FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe';
    IF n <> 3 THEN
        RAISE EXCEPTION 'mig552 probe: empty-pre-image update archived (% total rows, expected 3)', n;
    END IF;

    -- 5. Status-only touch: MUST NOT archive (UPDATE OF column list).
    UPDATE page_components SET build_status = 'deployed' WHERE id = v_pc_id;
    SELECT count(*) INTO n FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe';
    IF n <> 3 THEN
        RAISE EXCEPTION 'mig552 probe: status-only touch archived (% total rows)', n;
    END IF;

    -- Self-clean: the DELETE below fires 357''s delete arm (one more archive
    -- row) — remove the probe''s history in the same breath.
    DELETE FROM page_components WHERE id = v_pc_id;
    DELETE FROM page_component_history
     WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig552_probe';
END $probe$;

COMMIT;
