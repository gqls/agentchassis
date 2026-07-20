-- 180_tool_improver_rerender_request_ROLLBACK.sql
--
-- Reverses 180 by STRIPPING the four keys it added, returning
-- create_rerender_item's config to its pre-180 shape.
--
-- WHY STRIP RATHER THAN RESTORE FROM SNAPSHOT: snapshot_agent captures the
-- WHOLE agent row. Restoring it would also revert any change another thread
-- made to tool-improver after 180 landed — and the roster moves fast (this
-- agent's error_step routing changed twice on 2026-07-20 alone, commit
-- 40c8f00b4). Stripping exactly the keys 180 added touches nothing else.
-- Same reasoning as bugs_open/019's migration 177 rollback.
--
-- EFFECT OF ROLLING BACK: the Go affordances are additive and default-off, so
-- removing these keys returns the step to supplying no reason and no
-- component_id — i.e. it RE-OPENS bugs_open/024 (re-renders go back to
-- assemble-from-stale and report success). That is the intended pre-180 state,
-- but do not roll back expecting a neutral outcome.

ROLLBACK;

BEGIN;

-- Refuse to run if 180 was never applied, so a mistaken rollback is loud
-- rather than a silent 0-row success.
DO $$
DECLARE
    patched int;
BEGIN
    SELECT count(*) INTO patched
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND default_config #> '{workflow,steps,create_rerender_item,config}' ? 'spec_literal';

    IF patched <> 1 THEN
        RAISE EXCEPTION '180 ROLLBACK: expected exactly 1 patched active tool-improver row, found % — 180 may not be applied, or the roster moved', patched;
    END IF;
END $$;

SELECT snapshot_agent(
    'tool-improver',
    'ROLLBACK of 180: strip the rerender-request keys (re-opens bugs_open/024)'
);

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,create_rerender_item,config}',
        (default_config #> '{workflow,steps,create_rerender_item,config}')
            - 'recurrence_expected'
            - 'spec_literal'
            - 'spec_paths'
            - 'item_key_suffix_field'
    ),
    updated_at = NOW()
WHERE type = 'tool-improver'
  AND is_active = true
  AND default_config #> '{workflow,steps,create_rerender_item,config}' ? 'spec_literal'
RETURNING
    type,
    NOT (default_config #> '{workflow,steps,create_rerender_item,config}' ? 'spec_literal')        AS literal_gone,
    NOT (default_config #> '{workflow,steps,create_rerender_item,config}' ? 'spec_paths')          AS paths_gone,
    NOT (default_config #> '{workflow,steps,create_rerender_item,config}' ? 'recurrence_expected') AS recurrence_gone,
    NOT (default_config #> '{workflow,steps,create_rerender_item,config}' ? 'item_key_suffix_field') AS suffix_gone,
    -- The step itself must survive the rollback.
    default_config #> '{workflow,steps}' ? 'create_rerender_item'                                  AS step_still_present;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'pipeline',
    'build',
    E'## ROLLED BACK: tool-improver rerender request (180)\n\n'
    'The four keys added by 180 (recurrence_expected, spec_literal, spec_paths, '
    'item_key_suffix_field) were stripped from create_rerender_item. The step '
    'again supplies no spec.reason and no spec.component_id, so re-renders '
    'return to assemble-from-stale while reporting success — bugs_open/024 is '
    're-opened by this rollback.\n\n'
    'Categories: fix, migration',
    '["fix", "migration"]'::jsonb,
    'migration',
    '180_tool_improver_rerender_request_ROLLBACK.sql'
);

DELETE FROM schema_migrations
WHERE filename = '180_tool_improver_rerender_request.sql';

COMMIT;
