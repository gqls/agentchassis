-- 516 ROLLBACK — restore the UNMARKED `related_pages` wire on both tool build
-- steps, i.e. put migration 211's shape back.
--
-- WHAT ROLLING BACK RESTORES, stated plainly so the decision is informed: the
-- unmarked wire resolves the same path when it exists AND falls through to the
-- whole-tree search when it does not — which is `bugs_open/330`, the defect 516
-- exists to close. Roll back only if the MARKED form is causing harm (e.g. a
-- consumer turns out to need the searched value after all), never as tidy-up.
--
-- Hand-run. Not swept by the runner (a _ROLLBACK sidecar is refused client-side).

BEGIN;

SELECT snapshot_agent('tool-generator',
                      '516_ROLLBACK_tool_related_pages_optional_explicit: pre-rollback');
SELECT snapshot_agent('tool-deployer',
                      '516_ROLLBACK_tool_related_pages_optional_explicit: pre-rollback');

-- Discovery-driven, exactly like the apply and for the same reason: a hand-typed
-- list of carriers is a snapshot of whatever the author's census could see, and
-- a rollback that misses a nested carrier leaves it converted with no record.
DO $$
DECLARE
    tgt record;
    found int := 0;
BEGIN
    FOR tgt IN
        WITH RECURSIVE walk AS (
            SELECT d.type AS agent_type, ARRAY[]::text[] AS path, d.default_config AS node
              FROM agent_definitions d
             WHERE d.is_active AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
            UNION ALL
            SELECT w.agent_type, w.path || e.key, e.value
              FROM walk w CROSS JOIN LATERAL jsonb_each(w.node) e
             WHERE jsonb_typeof(w.node) = 'object'
        )
        SELECT agent_type, path AS config_path, node ->> 'related_pages?' AS ref
          FROM walk
         WHERE jsonb_typeof(node) = 'object'
           AND path[array_length(path,1)] = 'config'
           AND node ? 'related_pages?'
         ORDER BY 1, 2
    LOOP
        found := found + 1;

        UPDATE agent_definitions
           SET default_config = jsonb_set(default_config,
                   tgt.config_path || ARRAY['related_pages'],
                   to_jsonb(tgt.ref), true),
               updated_at = NOW()
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

        UPDATE agent_definitions
           SET default_config = default_config #- (tgt.config_path || ARRAY['related_pages?']),
               updated_at = NOW()
         WHERE type = tgt.agent_type AND is_active
           AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;
    END LOOP;

    IF found = 0 THEN
        RAISE EXCEPTION '516 ROLLBACK: no related_pages? carrier found anywhere — nothing to roll '
            'back (516 was not applied, or someone has already reverted it)';
    END IF;
    RAISE NOTICE '516 ROLLBACK: reverted % carrier(s) to the unmarked 211 wire', found;
END $$;

DO $$
DECLARE
    leftover int;
    restored int;
BEGIN
    WITH RECURSIVE walk AS (
        SELECT d.type AS agent_type, ARRAY[]::text[] AS path, d.default_config AS node
          FROM agent_definitions d
         WHERE d.is_active AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
        UNION ALL
        SELECT w.agent_type, w.path || e.key, e.value
          FROM walk w CROSS JOIN LATERAL jsonb_each(w.node) e
         WHERE jsonb_typeof(w.node) = 'object'
    ), configs AS (
        SELECT node FROM walk
         WHERE jsonb_typeof(node) = 'object' AND path[array_length(path,1)] = 'config'
    )
    SELECT count(*) FILTER (WHERE node ? 'related_pages?'),
           count(*) FILTER (WHERE node ? 'related_pages')
      INTO leftover, restored FROM configs;

    IF leftover <> 0 THEN
        RAISE EXCEPTION '516 ROLLBACK VERIFY FAILED: % related_pages? wire(s) survive', leftover;
    END IF;
    IF restored < 2 THEN
        RAISE EXCEPTION '516 ROLLBACK VERIFY FAILED: only % unmarked wire(s) restored, expected >= 2', restored;
    END IF;
    RAISE NOTICE '516 ROLLBACK OK: % unmarked wire(s), 0 marked — bugs_open/330 is reachable again', restored;
END $$;

COMMIT;
