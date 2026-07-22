-- 195_tool_improver_deliver_via_section_editor.sql
--
-- features_open/009 (from bugs_closed/024) — Option A, owner-approved 2026-07-21.
--
-- WHAT THIS CHANGES. tool-improver's post-fix delivery step (create_rerender_item)
-- currently emits a generic `needs_rerender` handled by `rerender-pages`. For a
-- TOOL page that path is a dead end: `rerender_page_sections -> save_page_sections`
-- hard-REFUSES any `rebuild_policy='owned'` page (experience-loop guard rail 1,
-- migration 164, commit fb89f1071), and EVERY tool page is owned by definition
-- (`UPDATE pages SET rebuild_policy='owned' WHERE page_type='tool'`). Migration 180
-- made that request well-formed, but a well-formed request to a forbidden path
-- still cannot deliver (proven: a reason-bearing page_rerender reached
-- save_sections and was refused).
--
-- THE SANCTIONED PATH is the one the guard's own error names: the `section-editor`
-- agent (`load_edit_context -> apply_section_edit -> git_commit ->
-- update_page_status`), which re-renders a single section from its CURRENT template
-- and reassembles — not the DELETE-and-reinsert the guard exists to stop. Driven by
-- hand it delivered the benchmark fix LIVE (correlation
-- c3828d17-cba4-4325-87b3-84b972ec9c7e; rendered_html 9,901 -> 10,705, flex fix on
-- the live page).
--
-- HOW. This step now creates a `section_edit` work item whose handler_agent is
-- `section-editor`. The build-dispatch-loop is generic — load_work_items has no
-- handler filter, and spawn_handler spawns `agent_type_field:
-- current_item.handler_agent` — so it routes the item to section-editor and maps
-- `spec: current_item.spec` into input_data.spec. section-editor resolves its inputs
-- (edit_type, page_name, slot_name, field_updates) from input_data.spec via the
-- action framework's recursive lookup (verified live: a spec-nested drive
-- COMPLETED, correlation 8dfbb732).
--
-- WHY THE FIELDS RESOLVE:
--   * edit_type='content_edit', field_updates={}  — spec_literal constants. A
--     content_edit with empty field_updates is a PURE RE-RENDER from the current
--     template (applyContentEdit requires one of field_updates/replacement_content_data;
--     {} satisfies it and merges nothing). This is exactly what delivered the fix
--     by hand. The template tool-improver just wrote (update_component) is what
--     load_edit_context loads.
--   * page_name <- tool_data.page_name, slot_name <- tool_data.function  — spec_paths.
--     load_tool selects both (output_format=object); verified present and populated
--     in real runs. load_edit_context resolves a section by page_name+slot_name when
--     no page_component_id is given. spec_paths is a HARD ERROR if unresolved (by
--     design), so a missing identifier fails loudly rather than delivering nothing.
--   * item_key = section_edit_tool_fix_<domain>_<component_id>, component-scoped
--     (item_key_suffix_field), with recurrence_expected so a SUCCESSFUL predecessor
--     is not counted as an anti-churn strike (the two-strike rule; dedup unchanged).
--
-- 180's spec_literal(reason)/spec_paths(component_id) are REPLACED here — they were
-- for the generic rerender request, which is retired for tool-improver. The generic
-- machinery (and 180's affordances) keep their value for NON-tool generic producers;
-- this change touches only tool-improver's tail.
--
-- SAFETY. Worst case if the section_edit item is somehow not delivered: the tool
-- fix does not reach the page — i.e. the status quo before this change, no
-- regression. The step key (create_rerender_item), output_field (rerender_item) and
-- next_step (compose_note) are UNCHANGED, so complete.config.output_fields' reference
-- to "rerender_item" stays valid (no dangling reference). The name is now a
-- misnomer; renaming it is a separate, non-urgent tidy (would require rewiring
-- update_component.next_step) and is deliberately not done here to keep the change
-- surgical.
--
-- LIVE-IMMEDIATELY. This is a config change; it takes effect on the next
-- tool-improver run with no image roll. No Go ships with it (every mechanism it
-- uses is already in the running binary).

-- Defensive: clear any sticky aborted transaction in this session.
ROLLBACK;

BEGIN;

-- ── Pre-flight: exactly one un-migrated active row ──────────────────────────
DO $$
DECLARE
    target_count int;
    already_applied int;
BEGIN
    SELECT count(*) INTO target_count
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND COALESCE(is_snapshot,false) = false
      AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,create_rerender_item,config,handler_agent}' = 'rerender-pages';

    SELECT count(*) INTO already_applied
    FROM agent_definitions
    WHERE type = 'tool-improver'
      AND is_active = true
      AND COALESCE(is_snapshot,false) = false
      AND deleted_at IS NULL
      AND default_config #>> '{workflow,steps,create_rerender_item,config,handler_agent}' = 'section-editor';

    IF already_applied > 0 THEN
        RAISE EXCEPTION '195: already applied (% row(s) route to section-editor) — not idempotent by design', already_applied;
    END IF;

    IF target_count <> 1 THEN
        RAISE EXCEPTION '195: expected exactly 1 active tool-improver whose create_rerender_item routes to rerender-pages, found % — the roster moved (or 180 not applied), re-diff before applying', target_count;
    END IF;

    RAISE NOTICE '195: pre-flight OK — 1 target row, 0 already applied';
END $$;

-- Snapshot before any agent_definitions change (platform convention).
SELECT snapshot_agent(
    'tool-improver',
    'features_open/009 (Option A): deliver the tool fix via the section-editor (apply_section_edit), not the generic rerender path the ownership guard forbids for tool pages'
);

-- ── The change: replace the delivery step's config (needle-gated on the
--    pre-state handler_agent) and update its description. Step key, output_field
--    and next_step are unchanged. ─────────────────────────────────────────────
UPDATE agent_definitions
SET default_config = jsonb_set(
        jsonb_set(
            default_config,
            '{workflow,steps,create_rerender_item,config}',
            jsonb_build_object(
                'source',                'tool-improver',
                'page_id',               'tool_data.page_id',
                'site_id',               'input_data.site_id',
                'summary',               'Deliver tool fix to the live page via the section-editor (apply_section_edit)',
                'priority',              50,
                'severity',              'medium',
                'item_type',             'section_edit',
                'item_domain',           'build',
                'handler_agent',         'section-editor',
                -- content_edit + empty field_updates = pure re-render from the
                -- current (just-written) template. Literals, so they ride in the
                -- item's spec and section-editor reads them from input_data.spec.
                'spec_literal',          jsonb_build_object('edit_type', 'content_edit', 'field_updates', '{}'::jsonb),
                -- section-editor resolves the target by page_name + slot_name.
                -- Both come from load_tool's tool_data (function == slot_name for a
                -- tool). Unresolved = hard error in create_work_item, by design.
                'spec_paths',            jsonb_build_object('page_name', 'tool_data.page_name', 'slot_name', 'tool_data.function'),
                -- Component-scoped dedup key; recurrence_expected so a SUCCESSFUL
                -- predecessor is not an anti-churn strike (dedup itself unchanged).
                'item_key_prefix',       'section_edit_tool_fix',
                'item_key_suffix_field', 'update_result.component_id',
                'recurrence_expected',   true
            )
        ),
        '{workflow,steps,create_rerender_item,description}',
        '"Deliver the tool fix via the section-editor (apply_section_edit) — the sanctioned path for rebuild_policy=owned tool pages"'::jsonb
    ),
    updated_at = NOW()
WHERE type = 'tool-improver'
  AND is_active = true
  AND COALESCE(is_snapshot,false) = false
  AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,create_rerender_item,config,handler_agent}' = 'rerender-pages'
RETURNING
    type,
    default_config #>> '{workflow,steps,create_rerender_item,config,handler_agent}'                  AS handler,
    default_config #>> '{workflow,steps,create_rerender_item,config,item_type}'                      AS item_type,
    default_config #>> '{workflow,steps,create_rerender_item,config,spec_literal,edit_type}'         AS edit_type,
    default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,page_name}'           AS page_name_path,
    default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}'           AS slot_name_path,
    default_config #>> '{workflow,steps,create_rerender_item,config,item_key_suffix_field}'          AS key_suffix,
    (default_config #> '{workflow,steps,create_rerender_item,config}' ->> 'recurrence_expected')::boolean AS recurrence,
    -- The step must still exist and still be referenced (no dangling reference).
    (default_config #> '{workflow,steps}' ? 'create_rerender_item')                                 AS step_present,
    (default_config #> '{workflow,steps,create_rerender_item,config}' ? 'field_updates')            AS has_field_updates_key;

-- ── Post-condition: exactly one row, changed correctly ──────────────────────
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
      AND default_config #>> '{workflow,steps,create_rerender_item,config,handler_agent}' = 'section-editor'
      AND default_config #>> '{workflow,steps,create_rerender_item,config,item_type}' = 'section_edit'
      AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_literal,edit_type}' = 'content_edit'
      AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,page_name}' = 'tool_data.page_name'
      AND default_config #>> '{workflow,steps,create_rerender_item,config,spec_paths,slot_name}' = 'tool_data.function'
      AND (default_config #> '{workflow,steps,create_rerender_item,config}' ->> 'recurrence_expected')::boolean IS TRUE
      AND default_config #> '{workflow,steps}' ? 'create_rerender_item'
      AND default_config #> '{workflow,steps,complete,config,output_fields}' @> '"rerender_item"'::jsonb;

    IF ok_count <> 1 THEN
        RAISE EXCEPTION '195: post-condition failed — % fully-migrated rows, expected 1', ok_count;
    END IF;

    RAISE NOTICE '195: post-condition OK — 1 row routes tool fixes to section-editor, step still present and referenced';
END $$;

-- Workflow-altering migration leaves a ('pipeline','build') note.
INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
VALUES (
    'pipeline',
    'build',
    E'## tool-improver now delivers tool fixes via the SECTION-EDITOR (features_open/009, Option A)\n\n'
    'Observed (bugs_closed/024): a tool-improver template fix was written correctly '
    'to content_components.html_template and NEVER reached the live page. Six '
    'defects were diagnosed on the GENERIC rerender path — but the real blocker is '
    'that the generic section-render->save path is DELIBERATELY forbidden for tool '
    'pages: save_page_sections hard-refuses rebuild_policy=owned pages (experience-'
    'loop guard rail 1, migration 164), and every tool page is owned by definition. '
    'A reason-bearing page_rerender (what migration 180 produces) reached '
    'save_sections and was refused — the first proof run to get that far.\n\n'
    'Fix (Option A, owner-approved): the delivery step now emits a `section_edit` '
    'work item routed to the `section-editor` agent (apply_section_edit / '
    'content_edit), the guard''s own sanctioned path, which re-renders a single '
    'section from its CURRENT template and reassembles — not the DELETE-and-reinsert '
    'the guard exists to stop. Driven by hand, that path delivered the benchmark fix '
    'LIVE (rendered_html 9,901 -> 10,705). The build-dispatch-loop routes the item '
    'generically (spawn_handler uses current_item.handler_agent).\n\n'
    '180''s generic-rerender request is retired FOR TOOL-IMPROVER by this change; its '
    'affordances (spec_literal/spec_paths/item_key_suffix_field/recurrence_expected) '
    'keep their value for non-tool generic producers.\n\n'
    'Verified: pending the first real run — apply, then drive improve -> section_edit '
    '-> section-editor and confirm the benchmark''s rendered .ltb-row-grid rule leaves '
    'display:grid and mobile-fit@mobile passes.\n\n'
    'Categories: fix, migration',
    '["fix", "migration"]'::jsonb,
    'migration',
    '195_tool_improver_deliver_via_section_editor.sql'
);

INSERT INTO schema_migrations (filename)
VALUES ('195_tool_improver_deliver_via_section_editor.sql');

COMMIT;
