-- 357_page_component_artefact_archive.sql
--
-- bugs_open/229 (owner ruling 2026-08-09: candidate 1 — extend the 344 shape):
-- `page_components.rendered_html` gets the same archive-on-destruction +
-- divergence-stamp guard that mig 344 gave `site_components`. Full design and
-- the measurements behind every decision:
-- docs024_key_docs_latest/bugfix_229_page_component_archive/PLAN_2026-08-09_….md
--
-- Shape, and where it deliberately differs from 344:
--
--   * ARCHIVE INTO THE EXISTING `page_component_history` (the ruling's word
--     "extend" is literal): five nullable columns carry the artefact arm; the
--     14.8k existing content_data rows and their four app writers are
--     untouched. Trigger rows are distinguishable by
--     source = 'artefact_archive_trigger'.
--   * A DELETE ARM. Page-side, DELETE is the dominant lifecycle (19,054
--     deletes vs 4,928 updates all-time — measured 2026-08-09): the
--     DELETE+INSERT rebuild family is how STY-025's interactive tools were
--     destroyed. No comparison is possible across delete+insert, so deletes
--     archive unconditionally (when the row held bytes).
--   * DELETE rows carry component_id NULL — the FK to the dying row would
--     reject the insert (ON DELETE SET NULL protects existing rows, not new
--     ones). Identity travels in the new slot_name/position columns instead.
--   * CASCADE SKIP, the one deliberate soft path: when a whole PAGE is
--     deleted, the cascade removes its components AFTER the pages row is
--     gone — an archive row can then neither resolve site_id nor satisfy
--     history's own FK to pages. The function returns without archiving in
--     exactly that case (full-page deletion is not the silent-section-wipe
--     class; 740 pages deleted all-time). Landmine, stated in STY-056.
--
-- FAIL-CLOSED (RFC_017 precedent, page-side evidence): write rate measured at
-- 27-290 rows/day — not the firehose the 226 scope-out feared; roles identical
-- to chrome's (clients_user owns everything); a broken history table halts
-- page writes LOUDLY rather than silently recreating the bug. Rollback is the
-- one-statement sidecar.
--
-- The digest means "reproducible from content_data". It is stamped by the Go
-- render/save paths ONLY (same-statement md5): save_page_sections,
-- rebuild_blog_listing, section_editor, create_report_page. Deliberately NOT
-- stamped: adopt_verbatim (ported bytes are not reproducible), the colour-fix
-- artefact rewriters, admin edits, raw psql — their content flags as
-- hand_patched at the next rebuild, which is the point.
--
-- ROLLBACK RECIPE (also in 357_page_component_artefact_archive_ROLLBACK.sql):
--   DROP TRIGGER IF EXISTS trg_page_component_artefact_archive_upd ON page_components;
--   DROP TRIGGER IF EXISTS trg_page_component_artefact_archive_del ON page_components;
--   DROP FUNCTION IF EXISTS page_component_artefact_archive();
-- Columns and archived data are KEPT — a rollback must not become the loss it
-- guards against (344 precedent).

BEGIN;

ALTER TABLE page_component_history
    ADD COLUMN rendered_html        text,
    ADD COLUMN rendered_html_digest text,
    ADD COLUMN divergence           text
        CONSTRAINT pch_divergence_check
        CHECK (divergence IS NULL OR divergence IN ('machine_made','hand_patched','unstamped')),
    ADD COLUMN application_name     text,
    ADD COLUMN op                   text
        CONSTRAINT pch_op_check
        CHECK (op IS NULL OR op IN ('overwrite','delete')),
    ADD COLUMN slot_name            varchar(100),
    ADD COLUMN position             integer;

ALTER TABLE page_components ADD COLUMN rendered_html_digest text;

CREATE FUNCTION page_component_artefact_archive() RETURNS trigger AS $fn$
DECLARE
    v_site_id uuid;
    v_verdict text;
BEGIN
    -- Resolve the site through the page. On a page-cascade delete the pages
    -- row is already gone: skip (see header — the FK forbids archiving here).
    SELECT p.site_id INTO v_site_id FROM pages p WHERE p.id = OLD.page_id;
    IF NOT FOUND THEN
        IF TG_OP = 'DELETE' THEN
            RETURN OLD;
        END IF;
        -- An UPDATE on a component whose page is missing should be impossible
        -- (FK); if it happens, fail closed rather than archive unattributably.
        RAISE EXCEPTION 'page_component_artefact_archive: page % not found for component %', OLD.page_id, OLD.id;
    END IF;

    v_verdict := CASE
        WHEN OLD.rendered_html_digest IS NULL THEN 'unstamped'
        WHEN OLD.rendered_html_digest = md5(OLD.rendered_html) THEN 'machine_made'
        ELSE 'hand_patched'
    END;

    INSERT INTO page_component_history (
        component_id, page_id, site_id, content_data, source,
        rendered_html, rendered_html_digest, divergence, application_name, op,
        slot_name, position
    ) VALUES (
        CASE WHEN TG_OP = 'DELETE' THEN NULL ELSE OLD.id END,
        OLD.page_id, v_site_id,
        COALESCE(OLD.content_data, '{}'::jsonb),
        'artefact_archive_trigger',
        OLD.rendered_html, OLD.rendered_html_digest, v_verdict,
        current_setting('application_name', true),
        CASE WHEN TG_OP = 'DELETE' THEN 'delete' ELSE 'overwrite' END,
        OLD.slot_name, OLD.position
    );

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$fn$ LANGUAGE plpgsql;

-- AFTER + FOR EACH ROW: an INSERT failure aborts the triggering statement
-- (fail-closed). The WHEN gates mean byte-identical rewrites, status-only
-- touches, and empty-artefact rows archive nothing.
CREATE TRIGGER trg_page_component_artefact_archive_upd
AFTER UPDATE OF rendered_html ON page_components
FOR EACH ROW
WHEN (OLD.rendered_html IS NOT NULL AND OLD.rendered_html <> ''
      AND NEW.rendered_html IS DISTINCT FROM OLD.rendered_html)
EXECUTE FUNCTION page_component_artefact_archive();

CREATE TRIGGER trg_page_component_artefact_archive_del
AFTER DELETE ON page_components
FOR EACH ROW
WHEN (OLD.rendered_html IS NOT NULL AND OLD.rendered_html <> '')
EXECUTE FUNCTION page_component_artefact_archive();

-- ---------------------------------------------------------------------------
-- VERIFY — induced probe, DO/RAISE. Exercises every arm on a THROWAWAY row
-- (the existing 14.8k history rows mean the negative control must count only
-- source='artefact_archive_trigger' rows, never totals): negative (no-op
-- rewrite), machine_made overwrite, hand_patched overwrite, DELETE archive
-- with component_id NULL. Probe rows self-delete; no live row is touched.
-- ---------------------------------------------------------------------------
DO $probe$
DECLARE
    v_page_id uuid;
    v_pc_id   uuid;
    n         int;
    v_div     text;
    v_op      text;
    v_comp    uuid;
    v_bytes   text;
BEGIN
    SELECT id INTO v_page_id FROM pages ORDER BY created_at LIMIT 1;
    IF v_page_id IS NULL THEN
        RAISE EXCEPTION 'mig357 probe: no pages row to hang the probe on';
    END IF;

    INSERT INTO page_components (page_id, position, slot_name, rendered_html, content_data, build_status)
    VALUES (v_page_id, 9999, 'mig357_probe', 'PROBE-ORIGINAL', '{}'::jsonb, 'pending')
    RETURNING id INTO v_pc_id;

    -- Stamp it the way the render path will (digest-only UPDATE must not fire
    -- the trigger — it is UPDATE OF rendered_html).
    UPDATE page_components SET rendered_html_digest = md5(rendered_html) WHERE id = v_pc_id;
    SELECT count(*) INTO n FROM page_component_history
    WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig357_probe';
    IF n <> 0 THEN
        RAISE EXCEPTION 'mig357 probe: digest-only stamp fired the trigger (% rows)', n;
    END IF;

    -- Negative: byte-identical rewrite archives nothing.
    UPDATE page_components SET rendered_html = rendered_html WHERE id = v_pc_id;
    SELECT count(*) INTO n FROM page_component_history
    WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig357_probe';
    IF n <> 0 THEN
        RAISE EXCEPTION 'mig357 probe: no-op rewrite archived % row(s); WHEN gate wrong', n;
    END IF;

    -- Machine-made overwrite: outgoing bytes match their stamp.
    -- NOTE: every archive row in this transaction shares one created_at
    -- (now() = xact start), so ORDER BY created_at cannot distinguish them —
    -- the first apply attempt failed on exactly that. Each check therefore
    -- selects its row BY ITS BYTES (row identity, not ordering).
    UPDATE page_components SET rendered_html = 'PROBE-PATCHED' WHERE id = v_pc_id;
    SELECT divergence, op, component_id INTO v_div, v_op, v_comp
    FROM page_component_history
    WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig357_probe'
      AND rendered_html = 'PROBE-ORIGINAL';
    IF v_div IS DISTINCT FROM 'machine_made' OR v_op IS DISTINCT FROM 'overwrite' OR v_comp IS DISTINCT FROM v_pc_id THEN
        RAISE EXCEPTION 'mig357 probe: machine overwrite recorded (%, %, %) — expected (machine_made, overwrite, probe id)', v_div, v_op, v_comp;
    END IF;

    -- Hand-patched overwrite: stored bytes no longer match the stale stamp.
    UPDATE page_components SET rendered_html = 'PROBE-FINAL' WHERE id = v_pc_id;
    SELECT divergence INTO v_div
    FROM page_component_history
    WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig357_probe'
      AND rendered_html = 'PROBE-PATCHED';
    IF v_div IS DISTINCT FROM 'hand_patched' THEN
        RAISE EXCEPTION 'mig357 probe: patched overwrite classified %, expected hand_patched', v_div;
    END IF;

    -- DELETE arm: archives the final bytes, component_id NULL (FK), op delete.
    DELETE FROM page_components WHERE id = v_pc_id;
    SELECT divergence, op, component_id, rendered_html INTO v_div, v_op, v_comp, v_bytes
    FROM page_component_history
    WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig357_probe'
      AND rendered_html = 'PROBE-FINAL';
    IF v_op IS DISTINCT FROM 'delete' OR v_comp IS NOT NULL OR v_bytes IS DISTINCT FROM 'PROBE-FINAL' THEN
        RAISE EXCEPTION 'mig357 probe: delete arm recorded (op %, component %, bytes %)', v_op, v_comp, v_bytes;
    END IF;

    SELECT count(*) INTO n FROM page_component_history
    WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig357_probe';
    IF n <> 3 THEN
        RAISE EXCEPTION 'mig357 probe: expected exactly 3 archive rows (machine, patched, delete), found %', n;
    END IF;

    -- Self-clean: the probe leaves no rows behind.
    DELETE FROM page_component_history
    WHERE source = 'artefact_archive_trigger' AND slot_name = 'mig357_probe';
END $probe$;

COMMIT;
