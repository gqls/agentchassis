-- 537 ROLLBACK — remove the `commit_sha?` declaration from build-dispatch-loop's
-- `mark_complete` step, returning the field to the whole-tree search.
--
-- WHAT ROLLING BACK RESTORES, stated plainly so the decision is informed: the
-- unwired field is resolved by `findFieldRecursive`, which inside a
-- multi-iteration loop collects one `commit_sha` per iteration and picks a
-- sorted-first winner. That is `bugs_open/334`, including the cross-iteration
-- case where one item's commit is attached to a DIFFERENT item entirely
-- (worked example: item `cc1db035`, a tool-generator `add_tool` wearing a
-- `section-editor` iteration's commit). Roll back only if the declared path is
-- actively wrong — never as tidy-up.
--
-- Hand-run. Not swept by the runner (a _ROLLBACK sidecar is refused client-side).

BEGIN;

SELECT snapshot_agent('build-dispatch-loop',
                      '537_ROLLBACK_bdl_mark_complete_declares_commit_sha: pre-rollback');

DO $$
DECLARE
    step_path text[];
BEGIN
    SELECT path INTO step_path FROM (
        WITH RECURSIVE walk AS (
            SELECT ARRAY[]::text[] AS path, d.default_config AS node
              FROM agent_definitions d
             WHERE d.type = 'build-dispatch-loop' AND d.is_active
               AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
            UNION ALL
            SELECT w.path || e.key, e.value
              FROM walk w CROSS JOIN LATERAL jsonb_each(w.node) e
             WHERE jsonb_typeof(w.node) = 'object'
        )
        SELECT path FROM walk
         WHERE jsonb_typeof(node) = 'object' AND node ? 'commit_sha?'
    ) z;

    IF step_path IS NULL THEN
        RAISE EXCEPTION '537 ROLLBACK: no commit_sha? wire found in build-dispatch-loop — nothing '
            'to roll back (537 was not applied, or someone has already reverted it)';
    END IF;

    UPDATE agent_definitions
       SET default_config = default_config #- (step_path || ARRAY['commit_sha?']),
           updated_at = NOW()
     WHERE type = 'build-dispatch-loop' AND is_active
       AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;

    RAISE NOTICE '537 ROLLBACK: removed commit_sha? from %', array_to_string(step_path, '.');
END $$;

DO $$
DECLARE
    still int;
BEGIN
    SELECT count(*) INTO still FROM (
        WITH RECURSIVE walk AS (
            SELECT ARRAY[]::text[] AS path, d.default_config AS node
              FROM agent_definitions d
             WHERE d.type = 'build-dispatch-loop' AND d.is_active
               AND COALESCE(d.is_snapshot,false) = false AND d.deleted_at IS NULL
            UNION ALL
            SELECT w.path || e.key, e.value
              FROM walk w CROSS JOIN LATERAL jsonb_each(w.node) e
             WHERE jsonb_typeof(w.node) = 'object'
        )
        SELECT 1 FROM walk WHERE jsonb_typeof(node) = 'object' AND node ? 'commit_sha?'
    ) z;

    IF still <> 0 THEN
        RAISE EXCEPTION '537 ROLLBACK VERIFY FAILED: % commit_sha? wire(s) survive', still;
    END IF;
    RAISE NOTICE '537 ROLLBACK OK: bugs_open/334 is reachable again';
END $$;

COMMIT;
