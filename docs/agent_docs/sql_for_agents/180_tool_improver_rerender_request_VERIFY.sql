-- 180_tool_improver_rerender_request_VERIFY.sql
--
-- Read-only. Safe to run before OR after 180. Every row it prints should be
-- checked by eye — a migration that reports success is not the same as a
-- pipeline that works (the invariant bugs_open/024 was built on).

\echo '=== 1. Is 180 recorded in the ledger? (bugs_open/007 trap) ==='
SELECT filename, applied_at
FROM schema_migrations
WHERE filename LIKE '180_%';

\echo ''
\echo '=== 2. The patched step config — read it, do not trust a row count ==='
SELECT jsonb_pretty(default_config #> '{workflow,steps,create_rerender_item,config}')
FROM agent_definitions
WHERE type = 'tool-improver' AND is_active = true;

\echo ''
\echo '=== 3. Post-conditions, one boolean each (all must be t) ==='
SELECT
    default_config #>> '{workflow,steps,create_rerender_item,config,spec_literal,reason}'
        = 'section_data_resolved'                                    AS reason_stamped,
    default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,component_id}'
        = 'update_result.component_id'                               AS component_path_set,
    (default_config #> '{workflow,steps,create_rerender_item,config}' ->> 'recurrence_expected')::boolean
                                                                     AS recurrence_on,
    default_config #>> '{workflow,steps,create_rerender_item,config,item_key_suffix_field}'
        = 'update_result.component_id'                               AS key_scoped,
    -- Round 4's plan DELETED this step; this one must not.
    default_config #> '{workflow,steps}' ? 'create_rerender_item'     AS step_still_present,
    -- ...and complete's output_fields must still be able to reference it.
    default_config #> '{workflow,steps,complete,config,output_fields}' @> '["rerender_item"]'::jsonb
                                                                     AS output_field_ref_intact
FROM agent_definitions
WHERE type = 'tool-improver' AND is_active = true;

\echo ''
\echo '=== 4. The consuming half is still wired as assumed (NOT changed by 180) ==='
\echo '    rerender-pages must still read reason/component_id off the inbound spec.'
SELECT
    default_config #>> '{workflow,steps,create_rerender_items,config,reason}'       AS reads_reason,
    default_config #>> '{workflow,steps,create_rerender_items,config,component_id}' AS reads_component_id
FROM agent_definitions
WHERE type = 'rerender-pages' AND is_active = true;

\echo ''
\echo '=== 5. Guardtest fixture must be UNPATCHED (it must not dispatch real work) ==='
SELECT type,
       (default_config #> '{workflow,steps,create_rerender_item,config}' ? 'spec_literal') AS wrongly_patched
FROM agent_definitions
WHERE type = 'tool-improver-guardtest' AND is_active = true;

\echo ''
\echo '=== 6. THE ACTUAL PROOF — does the render match the template yet? ==='
\echo '    Both false before a fresh improve+rerender cycle; both true when 024 is fixed.'
\echo '    Match the SPECIFIC rule, never a generic CSS property (the T24/T28 trap).'
SELECT
    (cc.html_template  LIKE '%minmax(0, 2fr)%') AS template_has_fix,
    (pc.rendered_html  LIKE '%minmax(0, 2fr)%') AS render_has_fix,
    length(cc.html_template) AS template_len,
    length(pc.rendered_html) AS render_len
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
WHERE pc.id = '45229c85-5600-4e8c-b4ae-e8058f74b185';

\echo ''
\echo '=== 7. Re-render request history — later items must NOT be born unresolved ==='
SELECT item_type, status, left(summary, 70) AS summary, created_at
FROM site_work_items
WHERE page_id = 'f25dd4d8-6e25-44eb-a021-689d3057d7a3'
  AND item_type IN ('needs_rerender', 'page_rerender')
ORDER BY created_at DESC
LIMIT 8;
