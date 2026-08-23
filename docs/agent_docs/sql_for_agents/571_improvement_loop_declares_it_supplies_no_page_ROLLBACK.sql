-- ROLLBACK for 571 — removes the `page_id?` / `component_id?` declarations from
-- improvement-loop's two site-wide `create_work_item` steps.
--
-- ⚠ WHAT ROLLING BACK RESTORES IS THE DEFECT, not a neutral prior state. With the markers
-- gone, both fields fall through to the whole-tree search again — and on a sweep whose tree
-- holds a discovery report, that search sees dozens of distinct page ids. Post-flip
-- (RFC_029 step 5, live v1.0.1323+) it will REFUSE them and write a
-- RESOLVER_CONFLICTING_CANDIDATES row per occurrence rather than guess, so the rollback is
-- safe — but it re-opens the noise, and on any binary predating the flip it would restore
-- the silent substitution of an arbitrary finding's page id onto a site-wide item.
--
-- Only roll back if 571's declaration is itself shown to be wrong — e.g. these steps turn
-- out to need a page after all, in which case the right move is to wire the correct path,
-- not to remove the declaration.

BEGIN;

DO $do$
DECLARE
    removed int := 0;
    tgt     record;
BEGIN
    FOR tgt IN
        SELECT s.key AS step_name
          FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'create_work_item'
           AND (s.value->'config' ? 'page_id?' OR s.value->'config' ? 'component_id?')
         ORDER BY s.key
    LOOP
        UPDATE agent_definitions
           SET default_config =
                   (default_config #- ARRAY['workflow','steps',tgt.step_name,'config','page_id?'])
                                   #- ARRAY['workflow','steps',tgt.step_name,'config','component_id?'],
               updated_at = NOW()
         WHERE type = 'improvement-loop' AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
        removed := removed + 1;
        RAISE NOTICE '571 ROLLBACK: improvement-loop.% markers removed', tgt.step_name;
    END LOOP;

    IF removed = 0 THEN
        RAISE NOTICE '571 ROLLBACK: nothing to remove — 571 is not applied (or was already rolled back)';
    END IF;

    -- Verify the removal AND that nothing else was lost.
    IF EXISTS (
        SELECT 1 FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'create_work_item'
           AND (s.value->'config' ? 'page_id?' OR s.value->'config' ? 'component_id?')
    ) THEN
        RAISE EXCEPTION '571 ROLLBACK: a marked key survives';
    END IF;

    IF EXISTS (
        SELECT 1 FROM agent_definitions d, jsonb_each(d.default_config->'workflow'->'steps') s
         WHERE d.type = 'improvement-loop' AND d.is_active
           AND COALESCE(d.is_snapshot, false) = false AND d.deleted_at IS NULL
           AND s.value->>'action' = 'create_work_item'
           AND NOT (s.value->'config' ? 'site_id' AND s.value->'config' ? 'item_type')
    ) THEN
        RAISE EXCEPTION '571 ROLLBACK: a pre-existing config key was lost';
    END IF;
END
$do$;

COMMIT;
