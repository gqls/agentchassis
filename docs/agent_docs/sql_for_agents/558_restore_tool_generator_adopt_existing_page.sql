-- 558_restore_tool_generator_adopt_existing_page.sql
--
-- RESTORE of migration 435's flag after an UN-SNAPSHOTTED removal (2026-08-22,
-- webdesign_tool_rebuilds lane; incident recorded in bugs_open/360's lane NOTES
-- and bugs_open/363-to-be if filed separately).
--
-- What happened: tool-generator save_tool carried adopt_existing_page=true from
-- 435 (2026-08-16) through every snapshotted migration up to and including 516
-- (pre-516 snapshot, state of 2026-08-21 13:50:42Z, still shows true). Some time
-- between 516's apply (~16:55Z 2026-08-21) and 2026-08-22 08:36:05Z (the live
-- row's updated_at) the key was REMOVED: no agent_definitions_backup snapshot,
-- no schema_migrations row, and the whole-config diff against the pre-516
-- snapshot shows exactly three key changes — related_pages -> related_pages?
-- (516's own surgical rename) and adopt_existing_page GONE (nobody's). 516
-- itself cannot have done it (jsonb_set one key + #- one key, read before this
-- file was written).
--
-- Effect while absent: create_tool_component falls back to CREATE-page and any
-- add_tool for an EXISTING tool page dies at save_tool with
-- pages_site_id_name_key 23505 while the work item reads complete/error NULL.
-- Measured casualty: item 21ab0704-a62e-4b1c-9a86-ae7af488d825 (2026-08-22
-- 11:28:53Z, tool-blueprint-compiler, Phase C #1) — the first adopt-route build
-- attempted after 516's apply. The running binary (v1.0.1322) still carries the
-- adopt literal (probed with a negative control), so config restore is the
-- whole fix.
--
-- Same statements as 435 (pre-guard: key absent; snapshot; one jsonb_set;
-- post-guard; doc_note), new number because forward-only and because
-- schema_migrations already carries 435's row.

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
        RAISE EXCEPTION '558: expected exactly 1 active tool-generator with save_tool=create_tool_component and no adopt_existing_page key, found % — the flag may already be back; re-read before applying', n;
    END IF;
END $$;

SELECT snapshot_agent('tool-generator',
    '558: restore adopt_existing_page after un-snapshotted removal (window 2026-08-21 ~16:55Z to 2026-08-22 08:36:05Z; first casualty item 21ab0704)');

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
        RAISE EXCEPTION '558: post-condition failed — % rows carry adopt_existing_page=true', n;
    END IF;
    RAISE NOTICE '558: post-condition OK';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline','build',
    E'## tool-generator save_tool: adopt_existing_page RESTORED (migration 558)\n\n'
    'The 435 flag was removed by an un-snapshotted write between 2026-08-21 ~16:55Z and 2026-08-22 08:36:05Z (writer unidentified; no backup row, no schema_migrations row). While absent, every add_tool for an existing tool page died at save_tool with pages_site_id_name_key 23505, item complete/error NULL. Restored by the webdesign_tool_rebuilds lane; first casualty item 21ab0704 (tool-blueprint-compiler) refiled after this. If you removed the key DELIBERATELY, say so in doc_notes and to the lane — and take a snapshot next time.',
    '["build-pipeline","tool-generator","bugs_open/286","webdesign_tool_rebuilds"]'::jsonb,
    'migration','558_restore_tool_generator_adopt_existing_page.sql');

INSERT INTO schema_migrations (filename) VALUES ('558_restore_tool_generator_adopt_existing_page.sql');

COMMIT;
