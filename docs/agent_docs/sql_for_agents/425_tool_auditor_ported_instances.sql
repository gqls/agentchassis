-- 425_tool_auditor_ported_instances.sql
--
-- bugs_open/281 — the tool-auditor can now audit a PORTED tool instance, and
-- cannot be pointed at an arbitrary one.
--
-- WHAT WAS WRONG (mechanism 2 of the bug). load_tool resolved the tool by
-- component only:
--     WHERE cc.id = $1::uuid AND cc.is_active = true LIMIT 1
-- with params ["input_data.component_id"]. webdesign.co.uk's 63 ported tools
-- are 63 page instances of ONE shared content_components row (function
-- 'ported-page', 115 instances across two sites), each instance's real tool
-- living in ITS page_components.rendered_html. Pointed at that component the
-- join yields ~115 rows and LIMIT 1 picks one at random — the audit would run
-- against some other site's tool and complete successfully. And the prompt
-- interpolated {{.tool_data.html_template}}, which for a ported instance is
-- the shared passthrough wrapper, not the tool.
--
-- WHAT THIS CHANGES (all in the live tool-auditor row; no Go ships with it):
--   1. load_tool pins the INSTANCE: `AND pc.page_id = $2::uuid`, params
--      ["input_data.component_id", "input_data.spec.page_id"]. The whole item
--      spec reaches the handler as input_data.spec (build-dispatch-loop's
--      call_handler input_mapping, migration 051), and every producer of
--      improve_tool / audit_tool writes spec.page_id (check_tool_health,
--      check_tool_acceptance, the Tier-4 judge, this agent's own items). A
--      query_database param that resolves to nil is a HARD step error
--      (database_actions.go), which is the loud failure LANDMINES prescribes
--      over an optional-path silent skip — and the pre-flight below proves no
--      open item lacks the key before the change lands.
--   2. load_tool selects cc.component_level and a `source_html` column: the
--      html_template for a real fork (its contract, and what tool-improver
--      edits), the instance's rendered_html for a ported tool. display_name
--      falls back to the page-derived subject key for a ported instance
--      (check_tool_acceptance precedent — the shared row's display_name would
--      label 63 tools identically). The llm_audit prompt reads source_html.
--   3. Routing gate: create_items_loop starts at check_target_class. A ported
--      instance's findings ALL go to create_review_item (needs_human_review);
--      only a real fork keeps the confidence split into create_improve_item →
--      tool-improver. tool-improver rewrites content_components.html_template,
--      which for a ported instance is the wrapper shared by every ported page
--      (clobbered fleet-wide 2026-08-05 and 2026-08-14). The Go-side fence in
--      update_component_html refuses that write too; this keeps the item from
--      being minted at all.
--   4. Both create steps carry page_id at top level and item_key_suffix_field
--      = tool_data.page_id. Today's keys are SITE-wide (audit_fix_<domain>,
--      audit_review_<domain>) so two tools' findings on one site collide on
--      idx_swi_dedup and one is silently dropped (bugs_closed/154's second
--      finding, deliberately left there for this change). create_review_item's
--      spec gains page_id / page_name so a reviewer can find the instance.
--
-- ORDER. Config is live immediately; the widened Go producers ride the next
-- chassis roll. Applied before the roll: existing real-fork items already
-- carry spec.page_id, so nothing breaks; the audit for a fork is unchanged
-- (source_html == html_template for component_level='tool').

ROLLBACK;

BEGIN;

-- ── Pre-flight ───────────────────────────────────────────────────────────────
DO $$
DECLARE
    target_count int;
    needle_count int;
    open_without_page_id int;
BEGIN
    SELECT count(*) INTO target_count
    FROM agent_definitions
    WHERE type = 'tool-auditor'
      AND is_active = true
      AND COALESCE(is_snapshot,false) = false
      AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id"]'::jsonb;
    IF target_count <> 1 THEN
        RAISE EXCEPTION '425: expected exactly 1 active un-migrated tool-auditor (load_tool params = [input_data.component_id]), found % — re-diff before applying', target_count;
    END IF;

    -- The prompt needle must occur exactly once, or the replace is not what it claims.
    SELECT (length(default_config::text) - length(replace(default_config::text, '{{.tool_data.html_template}}', '')))
           / length('{{.tool_data.html_template}}')
      INTO needle_count
    FROM agent_definitions
    WHERE type = 'tool-auditor' AND is_active AND COALESCE(is_snapshot,false) = false AND deleted_at IS NULL;
    IF needle_count <> 1 THEN
        RAISE EXCEPTION '425: expected exactly 1 occurrence of {{.tool_data.html_template}} in the tool-auditor config, found %', needle_count;
    END IF;

    -- The instance pin makes spec.page_id REQUIRED. Every open item this agent
    -- (or tool-improver, same shape in 426) could be dispatched must carry it.
    SELECT count(*) INTO open_without_page_id
    FROM site_work_items
    WHERE item_type IN ('audit_tool', 'improve_tool')
      AND status NOT IN ('complete','verified','rejected','wont_fix','failed','unresolved','cancelled')
      AND NOT (COALESCE(spec, '{}'::jsonb) ? 'page_id');
    IF open_without_page_id > 0 THEN
        RAISE EXCEPTION '425: % open audit_tool/improve_tool item(s) lack spec.page_id and would hard-fail load_tool after this change — backfill or cancel them first (SELECT id, item_type, item_key FROM site_work_items WHERE item_type IN (''audit_tool'',''improve_tool'') AND status NOT IN (''complete'',''verified'',''rejected'',''wont_fix'',''failed'',''unresolved'',''cancelled'') AND NOT (spec ? ''page_id''))', open_without_page_id;
    END IF;

    RAISE NOTICE '425: pre-flight OK — 1 target row, 1 prompt needle, 0 open items without spec.page_id';
END $$;

SELECT snapshot_agent(
    'tool-auditor',
    '425: bugs_open/281 — instance-pinned load_tool, source_html prompt source, ported findings routed to human review, per-page item keys'
);

-- ── 1+2. load_tool: instance pin + component_level + source_html + display_name ──
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,load_tool,config,query}',
            to_jsonb(
                'SELECT cc.id::text AS component_id, cc.function, cc.component_level, '
             || 'CASE WHEN cc.component_level = ''tool'' THEN COALESCE(cc.display_name, cc.function) ELSE regexp_replace(p.name, ''^tool-'', '''') END AS display_name, '
             || 'cc.html_template, COALESCE(pc.rendered_html, '''') AS rendered_html, '
             || 'CASE WHEN cc.component_level = ''tool'' THEN cc.html_template ELSE COALESCE(pc.rendered_html, '''') END AS source_html, '
             || 'cc.description, p.id::text AS page_id, p.name AS page_name, p.url AS page_url, COALESCE(p.build_status, '''') AS build_status '
             || 'FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id JOIN pages p ON pc.page_id = p.id '
             || 'WHERE cc.id = $1::uuid AND pc.page_id = $2::uuid AND cc.is_active = true LIMIT 1'
            )
        ),
        '{workflow,steps,load_tool,config,params}',
        '["input_data.component_id", "input_data.spec.page_id"]'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-auditor'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL
  AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id"]'::jsonb;

-- ── 2. llm_audit reads source_html (needle-gated: exactly one occurrence, proven above) ──
UPDATE agent_definitions
SET default_config = replace(default_config::text, '{{.tool_data.html_template}}', '{{.tool_data.source_html}}')::jsonb,
    updated_at = NOW()
WHERE type = 'tool-auditor'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL
  AND default_config::text LIKE '%{{.tool_data.html_template}}%';

-- ── 3. Routing gate: ported instance → review item, fork → confidence split ──
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,create_items_loop,config,sub_workflow,steps,check_target_class}',
            jsonb_build_object(
                'action', 'conditional',
                'config', jsonb_build_object(
                    'condition', 'tool_data.component_level != tool',
                    'then_step', 'create_review_item',
                    'else_step', 'check_confidence'
                ),
                'description', 'A ported instance (not a tool-level fork) never routes to tool-improver: its fix cannot be written to the shared component (bugs_open/281)'
            )
        ),
        '{workflow,steps,create_items_loop,config,sub_workflow,start_step}',
        '"check_target_class"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-auditor'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,start_step}' = 'check_confidence';

-- ── 4. Per-page item keys + page identity on both create steps ──
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            jsonb_set(
                jsonb_set(
                    default_config,
                    '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,page_id}',
                    '"tool_data.page_id"'::jsonb
                ),
                '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,item_key_suffix_field}',
                '"tool_data.page_id"'::jsonb
            ),
            '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,page_id}',
            '"tool_data.page_id"'::jsonb
        ),
        '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,item_key_suffix_field}',
        '"tool_data.page_id"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-auditor'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL;

UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,spec_data,page_id}',
            '"tool_data.page_id"'::jsonb
        ),
        '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,spec_data,page_name}',
        '"tool_data.page_name"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-auditor'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL;

-- ── Post-condition ───────────────────────────────────────────────────────────
DO $$
DECLARE
    ok_count int;
BEGIN
    SELECT count(*) INTO ok_count
    FROM agent_definitions
    WHERE type = 'tool-auditor'
      AND is_active = true
      AND COALESCE(is_snapshot,false) = false
      AND deleted_at IS NULL
      AND default_config #> '{workflow,steps,load_tool,config,params}' = '["input_data.component_id", "input_data.spec.page_id"]'::jsonb
      AND default_config #>> '{workflow,steps,load_tool,config,query}' LIKE '%pc.page_id = $2::uuid%'
      AND default_config #>> '{workflow,steps,load_tool,config,query}' LIKE '%AS source_html%'
      AND default_config #>> '{workflow,steps,llm_audit,config,prompt_template}' LIKE '%{{.tool_data.source_html}}%'
      AND default_config::text NOT LIKE '%{{.tool_data.html_template}}%'
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,start_step}' = 'check_target_class'
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,check_target_class,config,then_step}' = 'create_review_item'
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,check_target_class,config,else_step}' = 'check_confidence'
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_improve_item,config,item_key_suffix_field}' = 'tool_data.page_id'
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,item_key_suffix_field}' = 'tool_data.page_id'
      AND default_config #>> '{workflow,steps,create_items_loop,config,sub_workflow,steps,create_review_item,config,spec_data,page_id}' = 'tool_data.page_id';
    IF ok_count <> 1 THEN
        RAISE EXCEPTION '425: post-condition failed — % fully-migrated tool-auditor rows, expected 1', ok_count;
    END IF;
    RAISE NOTICE '425: post-condition OK';
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'pipeline',
    'build',
    E'## tool-auditor audits a tool INSTANCE, and a ported instance never routes to tool-improver (bugs_open/281, migration 425)\n\n'
    'load_tool now pins the instance (component_id AND spec.page_id) — pointed at the shared '
    'ported-page component it used to LIMIT 1 an arbitrary page. The LLM reviews source_html: '
    'the html_template for a real fork, the page''s rendered_html for a ported instance. '
    'Findings on a ported instance all become needs_human_review items; only a tool-level fork''s '
    'certain/likely findings become improve_tool for tool-improver, whose writeback is the shared '
    'template and clobbered every ported page on 2026-08-05 and 2026-08-14. Item keys are per page '
    '(item_key_suffix_field = page_id) instead of per site.',
    '["build-pipeline", "tool-auditor", "bugs_open/281"]'::jsonb,
    'migration',
    '425_tool_auditor_ported_instances.sql'
);

INSERT INTO schema_migrations (filename)
VALUES ('425_tool_auditor_ported_instances.sql');

COMMIT;
