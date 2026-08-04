-- 313_tool_recreation_declares_it_has_no_sections_metadata.sql
--
-- bugs_open/194, and the one thing its council round (b6023fc1) asked for that was
-- not already in the change. The `bug_historian` seat's "missing" note:
--
--   "Whether tool-recreation-handler and any other future SavePageSectionsAction
--    caller whose collected_data shape structurally cannot resolve the default
--    should get [a declaration] seeded as a *mandatory* follow-up decision point,
--    or whether that decision is left to drift the way sections_metadata_field
--    itself drifted since 2026-02-18."
--
-- It should not be left to drift, so it is seeded here. This is the whole point of
-- the field: tool-recreation-handler's NULL content_data is CORRECT — it recreates
-- a whole-page tool as one HTML blob, its step graph has no writer call anywhere
-- (recreate_tool -> validate_tool -> save from validation_result.clean_html), and
-- rerender_page_sections_action.go:318 already exempts a self-contained tool
-- section from the missing-content escalation. Before this seed that fact lived in
-- two Go comments and a bug file. After it, the caller says so itself — which is
-- RFC_010's rule (owner ruling 2026-08-02): "a comment is not a control on a tree
-- this many sessions share".
--
-- WHAT IT DOES, and what it deliberately does not. `expects_no_sections_metadata`
-- suppresses two things: the default-path lookup, and the CONTENT_DATA_REGRESSION
-- record. It is NOT the refusal — `require_sections_metadata` is the opposite
-- declaration and is still seeded on nobody.
--
-- INERT UNTIL THE NEXT CHASSIS ROLL, and no ordering constraint is claimed (the
-- 2026-07-29 owner ruling retired condition (1) of the ordering exemption, and
-- there genuinely is none here). The running binary does not read this key, so it
-- ignores it; the new binary reads it. Both directions are safe, in either order.
-- Applying it now rather than after the roll means the config cannot be forgotten
-- in the gap — which is the exact failure this whole bug is about.
--
-- BELT AND BRACES, NOT LOAD-BEARING. Even with the key absent, the default
-- structurally cannot resolve on this caller: datahelpers.ExtractNestedField
-- (data_helpers.go:1199-1234) walks top-level keys plus a `.response` auto-unwrap
-- and has no input_data fallback, and this agent's collected_data has no
-- page_content key at any depth. So this seed changes no behaviour. It changes what
-- a READER of the config can tell without reading two Go files — see the
-- prior_art_librarian seat's related point that a load-bearing claim should be
-- checkable in the schema rather than trusted from prose.

BEGIN;

SELECT snapshot_agent('tool-recreation-handler',
    'pre-update: bugs_open/194 — declare that this caller has no sections_metadata by design (PBP-031)');

-- NOTE the path: this caller's save_sections is at the TOP level of the workflow,
-- not inside a loop sub_workflow like four of the other five. Verified, not copied
-- from its siblings — copying a sibling's path is the shape of the bug this seed
-- belongs to.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,save_sections,config,expects_no_sections_metadata}',
        'true'::jsonb,
        true),
    updated_at = NOW()
WHERE type = 'tool-recreation-handler'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,save_sections}'
    INTO cfg FROM agent_definitions
    WHERE type = 'tool-recreation-handler'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '194/313: tool-recreation-handler has no top-level save_sections step';
    END IF;

    -- IS DISTINCT FROM, not <>: a missing jsonb path is NULL and `NULL <> 'x'` is
    -- NULL, so a plain <> against an absent key can never fire.
    IF cfg #>> '{config,expects_no_sections_metadata}' IS DISTINCT FROM 'true' THEN
        RAISE EXCEPTION '194/313: expects_no_sections_metadata is %, expected true',
            COALESCE(cfg #>> '{config,expects_no_sections_metadata}', '<NULL>');
    END IF;

    -- The declaration and the refusal are opposites; holding both would be
    -- incoherent. Assert the refusal is NOT here, so a later hand-edit that adds it
    -- has to notice this line.
    IF cfg #>> '{config,require_sections_metadata}' IS NOT NULL THEN
        RAISE EXCEPTION '194/313: this step declares BOTH expects_no_sections_metadata and require_sections_metadata — %', cfg->'config';
    END IF;

    -- the four pre-existing keys must survive: this is an ADD, not a rewrite.
    -- page_name_field is page_record.name here — NOT current_page.name (page-*) and
    -- NOT current_item.spec.name (site-work-orchestrator). Three callers, three
    -- different values; asserted verbatim.
    IF cfg #>> '{config,html_field}'         IS DISTINCT FROM 'validation_result.clean_html'
       OR cfg #>> '{config,site_id_field}'   IS DISTINCT FROM 'site_record.site_id'
       OR cfg #>> '{config,page_name_field}' IS DISTINCT FROM 'page_record.name'
       OR cfg #>> '{config,error_step}'      IS DISTINCT FROM 'complete_error' THEN
        RAISE EXCEPTION '194/313: an existing save_sections config key was disturbed — %', cfg->'config';
    END IF;

    RAISE NOTICE '194/313 OK — declaration added, 4 existing keys intact, refusal absent';
END $$;

COMMIT;

-- ROLLBACK if needed:
--   UPDATE agent_definitions SET default_config = default_config
--     #- '{workflow,steps,save_sections,config,expects_no_sections_metadata}'
--   WHERE type='tool-recreation-handler' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
