-- 435_tool_generator_adopt_existing_page_HOLD.sql
--
-- ⚠ HOLD — DO NOT APPLY until the chassis binary carrying bugs_open/286's Go
-- (create_tool_component `adopt_existing_page`, register TL-044) has ROLLED.
-- The `_HOLD` suffix keeps the migration runner's --apply from taking it
-- (SIDECAR_RE excludes it; MEMORY "migration-runner-practice"). Applied against
-- the OLD binary the key is simply unread (the strict config check for this
-- action does not exist — create_tool_component has no ConfigKeys declaration),
-- so nothing breaks, but the pilot would then run and collide exactly as before
-- and burn a generator round. Image before config.
--
-- To apply after the roll: verify the binary
--   kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq "<the 286 commit sha>" /proc/1/exe   (+ an absent control)
-- then rename to 435_tool_generator_adopt_existing_page.sql (drop _HOLD, fix the
-- two filename literals below) and apply.
--
-- What it does: sets `adopt_existing_page: true` on tool-generator's save_tool
-- ONLY. Consumer census (TL-044): this is the single live step naming
-- create_tool_component. With the flag, an add_tool item whose function matches
-- an existing tool page ATTACHES to that page (deploy_tool's identity + role
-- machinery) instead of dying on pages_site_id_name_key. Greenfield tools are
-- unaffected in effect (creation goes through UpsertPageForRole with the same
-- columns). Config is live immediately once applied.

ROLLBACK;

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,save_tool,action}' = 'create_tool_component'
      AND NOT (default_config #> '{workflow,steps,save_tool,config}') ? 'adopt_existing_page';
    IF n <> 1 THEN
        RAISE EXCEPTION '435: expected exactly 1 active tool-generator with save_tool=create_tool_component and no adopt_existing_page key, found % — re-read before applying', n;
    END IF;
END $$;

SELECT snapshot_agent('tool-generator',
    '435: bugs_open/286 — save_tool opts into adopt_existing_page (attach a same-URL rebuild to the existing tool page instead of colliding)');

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,save_tool,config,adopt_existing_page}', 'true'::jsonb),
    updated_at = NOW()
WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,save_tool,action}' = 'create_tool_component';

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND (default_config #> '{workflow,steps,save_tool,config,adopt_existing_page}') = 'true'::jsonb;
    IF n <> 1 THEN
        RAISE EXCEPTION '435: post-condition failed — % rows carry adopt_existing_page=true', n;
    END IF;
    RAISE NOTICE '435: post-condition OK';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline','build',
    E'## tool-generator save_tool: adopt_existing_page ON (bugs_open/286, migration 435)\n\n'
    'create_tool_component now attaches a same-URL rebuild to the EXISTING tool page (resolveToolPageIdentity + UpsertPageForRole, Refresh:[]) instead of failing on pages_site_id_name_key and deleting its own component. This is the ported-tool replacement route (webdesign_tool_rebuilds lane): after the generator deploys, the lane retires the ported slot and re-renders as a separate verified step.',
    '["build-pipeline","tool-generator","bugs_open/286"]'::jsonb,
    'migration','435_tool_generator_adopt_existing_page_HOLD.sql');

INSERT INTO schema_migrations (filename) VALUES ('435_tool_generator_adopt_existing_page_HOLD.sql');

COMMIT;
