-- 344_site_component_history_divergence_guard.sql
-- bugs_open/226: a chrome rebuild silently discards hand-patched content —
-- renderAndStoreSiteComponent replaces site_components.rendered_html outright,
-- and content that exists ONLY in the stored artefact (a psql replace() like
-- mig 268's footer note, an admin edit, a chrome-fix append) dies with no
-- record. Measured twice on oufe.com, one rebuild, unnoticed eight days.
--
-- This file is the DB half of the fix (council corr cffbfec4-3bec-4577-8844-d17c546ded3e):
--
--   1. site_component_history — archives the OUTGOING rendered_html whenever
--      an UPDATE replaces it with something different. House shape from
--      page_component_history (entity refs + payload + discriminator), but
--      this one archives the ARTEFACT — no rendered_html archive existed
--      anywhere before this table.
--   2. site_components.rendered_html_digest — the render path (Go half, rides
--      the next chassis build) stamps md5(rendered_html) in the SAME statement
--      that stores the bytes. digest = md5(bytes) thereafter means "these are
--      exactly the bytes the render path wrote"; a mismatch means some other
--      writer patched the artefact since. DELIBERATELY NOT BACKFILLED: stamping
--      existing rows would declare every unknown hand-patch machine-made and
--      silence the detector. NULL classifies as 'unstamped' and converges as
--      the fleet re-renders (46 of 57 rows unstamped at authoring).
--   3. trg_site_component_archive — the archive is a TRIGGER, not Go call-site
--      edits, because the writer inventory has six classes (render overwrite,
--      relink-erase, set/replace, append, core-manager admin SQL, raw psql)
--      and only a trigger sees all six — raw psql being the class that caused
--      the bug. FAIL-CLOSED on purpose: an overwrite that cannot archive
--      aborts loudly. Do NOT "fix" a failing chrome rebuild by dropping the
--      trigger — it is refusing to destroy an unarchived artefact.
--
-- LANDMINE (also appended to LANDMINES.md): this trigger is invisible to every
-- grep of Go. The write site in render_site_components_action.go carries a
-- pointing comment; the register entry is STY-054.
--
-- Ordering: this file is live the moment it applies — BEFORE the bugs_open/117
-- staleness wave (built 2026-08-08, rides the next roll) rebuilds the
-- unstamped fleet. Applying first makes that wave the first thing archived
-- rather than the last thing lost. The Go half is inert until its image rolls;
-- nothing here depends on it.
--
-- Apply: kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--          psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < this_file
-- Then:  ./scripts/migration/run-migrations.sh --record-only <this_file> --note "..."
-- Rollback: 344_site_component_history_divergence_guard_ROLLBACK.sql (drops
--           trigger+function only; the table and column hold archived
--           artefacts and are kept).

BEGIN;

CREATE TABLE site_component_history (
    id                    uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    site_component_id     uuid REFERENCES site_components(id) ON DELETE SET NULL,
    site_id               uuid NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    slot_name             varchar(100) NOT NULL,
    component_id          uuid,
    rendered_html         text NOT NULL,
    render_inputs         jsonb,
    rendered_html_digest  text,
    -- What the OUTGOING artefact was, judged against its own stamp:
    --   machine_made — bytes are exactly what the render path last wrote
    --   hand_patched — bytes differ from the stamp: someone patched the artefact
    --   unstamped    — pre-fix row (or non-render writer since); cannot tell
    divergence            text NOT NULL CHECK (divergence IN ('machine_made','hand_patched','unstamped')),
    -- current_setting('application_name'): 'psql' for hand runs; the Go pool's
    -- name for pipeline writes. Advisory provenance, not authority.
    application_name      text,
    -- true when the replacement was NULL/erase (the relinkSiteComponent arm),
    -- not an overwrite with new bytes.
    new_is_null           boolean NOT NULL DEFAULT false,
    created_at            timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_sch_site ON site_component_history (site_id, created_at DESC);
CREATE INDEX idx_sch_component ON site_component_history (site_component_id, created_at DESC);

ALTER TABLE site_components ADD COLUMN rendered_html_digest text;

CREATE FUNCTION site_component_history_archive() RETURNS trigger AS $fn$
BEGIN
    INSERT INTO site_component_history (
        site_component_id, site_id, slot_name, component_id,
        rendered_html, render_inputs, rendered_html_digest,
        divergence, application_name, new_is_null
    ) VALUES (
        OLD.id, OLD.site_id, OLD.slot_name, OLD.component_id,
        OLD.rendered_html, OLD.render_inputs, OLD.rendered_html_digest,
        CASE
            WHEN OLD.rendered_html_digest IS NULL THEN 'unstamped'
            WHEN OLD.rendered_html_digest = md5(OLD.rendered_html) THEN 'machine_made'
            ELSE 'hand_patched'
        END,
        current_setting('application_name', true),
        NEW.rendered_html IS NULL
    );
    RETURN NEW;
END;
$fn$ LANGUAGE plpgsql;

-- AFTER + FOR EACH ROW: any INSERT failure aborts the triggering UPDATE
-- (fail-closed — see header). The WHEN gate means a byte-identical rewrite or
-- a status-only touch archives nothing.
CREATE TRIGGER trg_site_component_archive
AFTER UPDATE OF rendered_html ON site_components
FOR EACH ROW
WHEN (OLD.rendered_html IS NOT NULL AND OLD.rendered_html <> ''
      AND NEW.rendered_html IS DISTINCT FROM OLD.rendered_html)
EXECUTE FUNCTION site_component_history_archive();

-- ---------------------------------------------------------------------------
-- VERIFY — induced probe, DO/RAISE (a bare SELECT cannot stop the COMMIT).
-- Exercises: the negative (no-op rewrite archives nothing), the positive
-- (a differing overwrite archives the original bytes as 'unstamped'), and the
-- restore path. Probe rows are removed; the live row's bytes end exactly as
-- they began, and updated_at is never touched (these UPDATEs do not set it).
-- ---------------------------------------------------------------------------
DO $probe$
DECLARE
    probe_id uuid;
    orig     text;
    n        int;
BEGIN
    SELECT id, rendered_html INTO probe_id, orig
    FROM site_components
    WHERE rendered_html IS NOT NULL AND rendered_html <> ''
    ORDER BY site_id, slot_name
    LIMIT 1;

    IF probe_id IS NULL THEN
        RAISE EXCEPTION 'mig344 probe: no populated site_components row to exercise the trigger';
    END IF;

    -- Negative control first: a byte-identical rewrite must NOT archive.
    UPDATE site_components SET rendered_html = rendered_html WHERE id = probe_id;
    SELECT count(*) INTO n FROM site_component_history WHERE site_component_id = probe_id;
    IF n <> 0 THEN
        RAISE EXCEPTION 'mig344 probe: no-op rewrite archived % row(s); WHEN gate is wrong', n;
    END IF;

    -- Positive: a differing overwrite archives the ORIGINAL bytes, unstamped.
    UPDATE site_components
    SET rendered_html = rendered_html || '<!-- mig344 probe -->'
    WHERE id = probe_id;
    SELECT count(*) INTO n FROM site_component_history
    WHERE site_component_id = probe_id AND divergence = 'unstamped' AND rendered_html = orig;
    IF n <> 1 THEN
        RAISE EXCEPTION 'mig344 probe: expected exactly 1 archived row holding the original bytes, got %', n;
    END IF;

    -- Restore (fires the trigger again, archiving the probe-marked bytes).
    UPDATE site_components SET rendered_html = orig WHERE id = probe_id;
    SELECT count(*) INTO n FROM site_component_history WHERE site_component_id = probe_id;
    IF n <> 2 THEN
        RAISE EXCEPTION 'mig344 probe: expected 2 archive rows after restore, got %', n;
    END IF;

    -- The live row is byte-identical to where it started; drop the probe rows.
    DELETE FROM site_component_history WHERE site_component_id = probe_id;
END;
$probe$;

COMMIT;
