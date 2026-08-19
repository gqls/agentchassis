-- 496_tool_generator_replace_existing_HOLD.sql
--
-- ⚠ HOLD until the chassis roll that carries create_tool_component's
-- replace_existing arm (bugs_open/331, register TL-047) has ROLLED. The `_HOLD`
-- suffix keeps the migration runner's --apply from taking it (SIDECAR_RE
-- excludes it; MEMORY "migration-runner-practice"). Image before config.
--
-- Safe-to-apply-early? Technically yes — the key is an INPUT MAPPING, not a
-- config switch: on the old binary `replace_existing` is not in the action's
-- ActionInputSpec, so the mapped value is simply unread (create_tool_component
-- has no ConfigKeys declaration, so there is no strict-config refusal either).
-- But applying it early invites a lane to file a `replace_existing: true` item
-- against the old binary, which would then take the already_exists short-circuit
-- and burn a generator round having changed nothing. So: HOLD anyway.
--
-- To apply after the roll: verify the binary carries the fix commit
--   (probe the STAMP, then `git merge-base --is-ancestor <fix commit> <stamp>`,
--   with a junk-hex control — never "is my sha in the binary"),
-- then rename to 496_tool_generator_replace_existing.sql (drop _HOLD, fix the
-- two filename literals below) and apply.
--
-- What it does: adds ONE step-config line to tool-generator's save_tool —
--   "replace_existing!": "input_data.spec.replace_existing"
-- The trailing `!` is the STRICT marker (RFC_029 §9 D3): the field resolves ONLY
-- from that explicit path and never meets the whole-tree recursive search, so a
-- stray `replace_existing` elsewhere in collected data can never arm the
-- regeneration. Authority is therefore PER ITEM: an add_tool work item whose
-- spec carries `"replace_existing": true` regenerates the site's existing tool
-- component for that function IN PLACE (same component_id, placements'
-- rendered_html rewritten in one transaction, page_component_history archives
-- the old bytes); every other item — including a duplicate add_tool from the
-- suggester — still takes today's already_exists no-op. Greenfield builds and
-- the adopt_existing_page (286) route are unaffected. Consumer census: this is
-- the single live step naming create_tool_component (TL-044/TL-047).
-- Config is live immediately once applied.

ROLLBACK;

BEGIN;

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,save_tool,action}' = 'create_tool_component'
      AND NOT (default_config #> '{workflow,steps,save_tool,config}') ? 'replace_existing!'
      AND NOT (default_config #> '{workflow,steps,save_tool,config}') ? 'replace_existing';
    IF n <> 1 THEN
        RAISE EXCEPTION '496: expected exactly 1 active tool-generator with save_tool=create_tool_component and no replace_existing mapping, found % — re-read before applying', n;
    END IF;
END $$;

SELECT snapshot_agent('tool-generator',
    '496: bugs_open/331 — save_tool maps replace_existing! from input_data.spec.replace_existing (per-item in-place regeneration of the site''s own tool)');

UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
        '{workflow,steps,save_tool,config,replace_existing!}', '"input_data.spec.replace_existing"'::jsonb),
    updated_at = NOW()
WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,save_tool,action}' = 'create_tool_component';

DO $$
DECLARE n int;
BEGIN
    SELECT count(*) INTO n FROM agent_definitions
    WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,save_tool,config,replace_existing!}' = 'input_data.spec.replace_existing';
    IF n <> 1 THEN
        RAISE EXCEPTION '496: post-condition failed — % rows carry the replace_existing! mapping', n;
    END IF;
    RAISE NOTICE '496: post-condition OK';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES ('pipeline','build',
    E'## tool-generator save_tool: replace_existing mapped per item (bugs_open/331, migration 496)\n\n'
    'create_tool_component now REGENERATES the site''s existing tool component IN PLACE when the add_tool item''s spec carries replace_existing: true — same component_id, the live placement''s rendered_html rewritten in the same transaction (page_component_history archives the old bytes), no second row, no name/index collision, no by-hand slot retire. Items without the flag are unchanged (the already_exists probe stays the per-site throttle). The three-step hand recipe in the webdesign_tool_rebuilds RUNBOOK (deactivate, rename, retire-before-rerender) is retired for re-fixes once this is applied on a binary carrying TL-047.',
    '["build-pipeline","tool-generator","bugs_open/331"]'::jsonb,
    'migration','496_tool_generator_replace_existing.sql');

INSERT INTO schema_migrations (filename) VALUES ('496_tool_generator_replace_existing.sql');

COMMIT;
