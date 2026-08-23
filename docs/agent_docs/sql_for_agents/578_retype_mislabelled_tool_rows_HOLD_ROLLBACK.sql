-- ROLLBACK for 578 — restore the mislabelled rows exactly as they were.
--
-- Restores component_id, content_data and component_version_id from the backup
-- table 578 wrote before touching anything. It deliberately does NOT restore
-- rendered_html, slot_name or position: 578 never changed them, and a rollback
-- that writes columns the forward migration did not touch can only introduce
-- damage — if a rebuild has run since, those columns legitimately hold newer
-- bytes and putting the old ones back would be a silent content revert.
--
-- ⚠ WHAT THIS RESTORES IS THE DEFECT. The rows go back to declaring themselves
-- the shared `hero` while storing a whole interactive tool. That is a coherent
-- thing to want (it is the state the fleet ran in for months, and every mechanism
-- has been coping with it) but do it knowingly: the schema check will resume
-- filing false "missing headline" findings, and the rows become unrepairable
-- again because any backfill that gives them content_data would regenerate a
-- title band over the tool.

BEGIN;

DO $$
DECLARE n int;
BEGIN
    IF to_regclass('public.page_components_backup_357_20260823') IS NULL THEN
        RAISE EXCEPTION 'no backup table — 578 was never applied here, or the backup was dropped. '
                        'Refusing to guess what these rows used to be.';
    END IF;

    UPDATE page_components pc
       SET component_id         = b.component_id,
           content_data         = b.content_data,
           component_version_id = b.component_version_id
      FROM page_components_backup_357_20260823 b
     WHERE pc.id = b.id;

    GET DIAGNOSTICS n = ROW_COUNT;
    RAISE NOTICE 'bugs_open/357 phase 3 rollback: % row(s) restored to their pre-repair identity', n;
END $$;

COMMIT;

-- The backup table is deliberately LEFT IN PLACE. It is the only record of what
-- these rows looked like before, it costs almost nothing, and a rollback is
-- exactly the moment someone will want to look at it again.
