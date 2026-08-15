-- 426_tool_improver_instance_pinned_load.sql
--
-- bugs_open/281 — tool-improver loads the tool INSTANCE it was sent, and its
-- delivery step names the instance's real slot.
--
-- DEFENCE IN DEPTH, stated honestly. After 425 (tool-auditor) and the widened
-- Go producers (check_tool_health, check_tool_acceptance), tool-improver should
-- never RECEIVE an item for a ported instance: those file as ported_tool_fix /
-- needs_human_review with no handler. And the Go fence in update_component_html
-- refuses to overwrite a shared non-tool component regardless. This migration
-- closes the two remaining ways the improver itself could act on the wrong
-- instance if something hand-filed or pre-existing reaches it:
--
--   1. load_tool had the same component-only `WHERE cc.id = $1 LIMIT 1` as the
--      auditor's — pointed at a shared component it loads an arbitrary page's
--      context (page_id, page_name) and then DELIVERS the fix to that page. It
--      now pins `pc.page_id = $2::uuid` from input_data.spec.page_id (the whole
--      spec reaches the handler; every producer writes spec.page_id; 425's
--      pre-flight already proved no open item lacks it — re-proved here) and
--      selects pc.slot_name.
--   2. The delivery step (migration 195) bound spec_paths.slot_name to
--      tool_data.function — right for a fork only by the convention that a
--      tool's slot is named after its function, and wrong for a ported instance
--      (its slot is 'ported-page'). It now reads the ACTUAL placement,
--      tool_data.slot_name, and its dedup key suffix is the page rather than
--      the (possibly shared) component id.
--
-- The fix prompt keeps reading {{.tool_data.html_template}}: tool-improver's
-- OUTPUT replaces html_template, so it must review the template it will replace
-- (a fork's contract). This is deliberately NOT the auditor's source_html swap.
--
-- Config is live immediately; no Go ships with it.

ROLLBACK;

BEGIN;

DO $$
DECLARE
    target_count int;
    open_without_page_id int;
BEGIN
    SELECT count(*) INTO target_count
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND COALESCE(is_snapshot,false) = false
      AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id"]'::jsonb
      AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}' = 'tool_data.function';
    IF target_count <> 1 THEN
        RAISE EXCEPTION '426: expected exactly 1 active un-migrated tool-improver (component-only load_tool, slot_name <- tool_data.function), found % — re-diff before applying', target_count;
    END IF;

    SELECT count(*) INTO open_without_page_id
    FROM site_work_items
    WHERE item_type IN ('audit_tool', 'improve_tool')
      AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled')
      AND NOT (COALESCE(spec, '{}'::jsonb) ? 'page_id');
    IF open_without_page_id > 0 THEN
        RAISE EXCEPTION '426: % open audit_tool/improve_tool item(s) lack spec.page_id and would hard-fail load_tool after this change — backfill or cancel them first', open_without_page_id;
    END IF;

    RAISE NOTICE '426: pre-flight OK — 1 target row, 0 open items without spec.page_id';
END $$;

SELECT snapshot_agent(
    'tool-improver',
    '426: bugs_open/281 — instance-pinned load_tool (+slot_name), delivery slot from the actual placement, per-page delivery key'
);

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,load_tool,config,query}',
            to_jsonb(
                'SELECT cc.id::text as component_id, cc.function, cc.display_name, cc.html_template, cc.description, cc.semantic_tags::text as tags, cc.is_dark_section, cc.component_level, pc.slot_name, p.url as page_url, p.id::text as page_id, p.name as page_name '
             || 'FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id '
             || 'WHERE cc.id = $1::uuid AND pc.page_id = $2::uuid AND cc.is_active = true LIMIT 1'
            )
        ),
        '{workflow,steps,load_tool,config,params}',
        '["input_data.component_id", "input_data.spec.page_id"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-improver'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id"]'::jsonb;

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}',
            '"tool_data.slot_name"'::jsonb
        ),
        '{workflow,steps,create_rerender_item,config,item_key_suffix_field}',
        '"tool_data.page_id"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-improver'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}' = 'tool_data.function';

DO $$
DECLARE
    ok_count int;
BEGIN
    SELECT count(*) INTO ok_count
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND COALESCE(is_snapshot,false) = false
      AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id", "input_data.spec.page_id"]'::jsonb
      AND default_config #>> '{workflow,steps,load_tool,config,query}' LIKE '%pc.page_id = $2::uuid%'
      AND default_config #>> '{workflow,steps,load_tool,config,query}' LIKE '%pc.slot_name%'
      AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}' = 'tool_data.slot_name'
      AND default_config #>> '{workflow,steps,create_rerender_item,config,item_key_suffix_field}' = 'tool_data.page_id'
      -- The fix prompt still reads the template it replaces.
      AND default_config::text LIKE '%{{.tool_data.html_template}}%';
    IF ok_count <> 1 THEN
        RAISE EXCEPTION '426: post-condition failed — % fully-migrated tool-improver rows, expected 1', ok_count;
    END IF;
    RAISE NOTICE '426: post-condition OK';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'pipeline',
    'build',
    E'## tool-improver loads the tool INSTANCE it was sent (bugs_open/281, migration 426)\n\n'
    'load_tool pins component_id AND spec.page_id (it used to LIMIT 1 an arbitrary placement of a '
    'shared component) and selects the placement''s slot_name; the section-editor delivery step '
    'now names that actual slot instead of assuming slot == function, and dedups per page. '
    'Defence in depth: after 425 and the widened producers a ported instance never routes here, '
    'and update_component_html refuses to overwrite a shared non-tool component regardless.',
    '["build-pipeline", "tool-improver", "bugs_open/281"]'::jsonb,
    'migration',
    '426_tool_improver_instance_pinned_load.sql'
);

INSERT INTO schema_migrations (filename)
VALUES ('426_tool_improver_instance_pinned_load.sql');

COMMIT;
