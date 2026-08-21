-- 532_tool_generator_replace_existing_optional_marker.sql
--
-- HOTFIX for seed 496 (bugs_open/331), same day. 496 mapped the per-item flag
-- as "replace_existing!" — the STRICT marker — believing strict meant "resolve
-- only from this path". It means that AND "fail extraction when unresolved"
-- (action_inputs.go's strict-enforcement branch), so from 12:12:12Z every
-- add_tool item whose spec does NOT carry replace_existing — the shape every
-- pre-496 producer emits, tool-suggester foremost — died at save_tool with
-- "strict '!' fields did not resolve via their explicit mapping", while the
-- work item read complete with error NULL. Measured by the webdesign lane
-- (items cd3812b5 absent→failed, 298fa5f8 explicit-false→built) within 90
-- minutes of the seed. TL-047's "absent ⇒ byte-identical (pinned)" held at the
-- ACTION layer; the seed's marker broke absence at the EXTRACTION layer, which
-- no action-level test can see.
--
-- The correct marker is "?" — optional-explicit (MarkedConfigKey, live in the
-- running v1.0.1321 via ecc419bd1): resolution is EXPLICIT-ONLY (the RFC_029
-- whole-tree search can never supply the flag — the property 496 wanted) and
-- non-resolution passes through to plain absence (the property 496 broke).
-- Config is live immediately; no image involved.

ROLLBACK;

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND (default_config #> '{workflow,steps,save_tool,config}') ? 'replace_existing!';
    IF n <> 1 THEN
        RAISE EXCEPTION '532: expected exactly 1 active tool-generator carrying the broken replace_existing! mapping, found % — re-read before applying', n;
    END IF;
END $$;

SELECT snapshot_agent('tool-generator',
    '532: hotfix 496 — replace_existing mapping marker ! -> ? (strict-on-absence broke every plain add_tool at extraction; optional-explicit keeps search exclusion, allows absence)');

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config #- '{workflow,steps,save_tool,config,replace_existing!}',
        '{workflow,steps,save_tool,config,replace_existing?}', '"input_data.spec.replace_existing"'::jsonb),
    updated_at = NOW()
WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,save_tool,config,replace_existing?}' = 'input_data.spec.replace_existing'
      AND NOT (default_config #> '{workflow,steps,save_tool,config}') ? 'replace_existing!';
    IF n <> 1 THEN
        RAISE EXCEPTION '532: post-condition failed';
    END IF;
    RAISE NOTICE '532: post-condition OK — marker is now ?, ! removed';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline','build',
    E'## HOTFIX: tool-generator save_tool replace_existing marker ! -> ? (bugs_open/331, migration 532)\n\n'
    'Seed 496 (12:12Z 2026-08-21) used the strict ! marker, which FAILS extraction when the field is absent — every plain add_tool (no replace_existing in spec) died at save_tool while its item read complete. Fixed 532: the ? optional-explicit marker keeps the search exclusion and allows absence. Items that failed in the window 12:12–<apply time> with "strict ''!'' fields did not resolve" need refiling.',
    '["build-pipeline","tool-generator","bugs_open/331"]'::jsonb,
    'migration','532_tool_generator_replace_existing_optional_marker.sql');

INSERT INTO schema_migrations (filename) VALUES ('532_tool_generator_replace_existing_optional_marker.sql');

COMMIT;
