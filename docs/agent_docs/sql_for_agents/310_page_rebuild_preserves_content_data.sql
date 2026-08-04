-- 310_page_rebuild_preserves_content_data.sql
--
-- bugs_open/194 — four of six save_page_sections callers never map
-- sections_metadata, so every save through them writes content_data = NULL.
-- This seed fixes ONE of the four: page-rebuild.
--
-- FOUND BY bugs_open/087's own acceptance run, which is why it is fixed here and
-- not deferred: that run (correlation 3fdf4acf-5f96-49f9-8801-28047aae92ef,
-- 2026-08-04 09:47-09:50Z) rebuilt vetcomparison.uk/tool-cma-obligation-checker-guide
-- correctly in every other respect — all steps COMPLETED, page serves 200 with all
-- three components — and NULLed content_data on all three slots (644 / 3810 / 420
-- chars before, NULL after). Leaving a live page unable to rerender would be damage
-- recorded and not repaired.
--
-- MECHANISM. SavePageSectionsAction writes content_data on insert
-- (save_page_sections_action.go:685-687) from the section's own content_data, which
-- reaches it through the `sections_metadata_field` config key. Read off the live
-- rows — note the step is nested inside a loop sub_workflow in four of the six, so
-- a top-level jsonb_each MISSES them and you need
-- jsonb_path_query(default_config, '$.**.steps.*'):
--
--     page-build-handler        page_content.response.sections_metadata   preserved
--     page-rerender             rerender_sections.sections_metadata       preserved
--     page-rebuild              ABSENT                                    NULLed  <- this seed
--     pageflow-builder          ABSENT                                    NULLed
--     site-work-orchestrator    ABSENT                                    NULLed
--     tool-recreation-handler   ABSENT                                    NULLed
--
-- THE DATA IS THERE AND SIMPLY NOT MAPPED. page-content-writer's compile_page output
-- on the run above has keys: page_html, page_name, section_count, sections_metadata.
-- page-rebuild stores the writer's reply under output_field: page_content, so
-- `page_content.response.sections_metadata` — byte-identical to the value
-- page-build-handler already uses — resolves here too.
--
-- SCOPE, deliberately narrow. The other three callers are NOT touched:
-- pageflow-builder and site-work-orchestrator are the same two dormant callers
-- bugs_open/087 found broken in the other direction, and tool-recreation-handler
-- runs a different writer flow whose response shape this lane has NOT read —
-- copying the key there would be an unmeasured claim. bugs_open/194 carries them.
--
-- Config-only: live on apply, no image roll. One added key; no existing key is
-- rewritten, so the other three config entries cannot be disturbed.
--
-- VERIFY: the DO block below checks the shape. The behavioural proof is a re-run of
-- the same page requiring BOTH halves — length(rendered_html) AND
-- length(content_data::text) non-null on every slot, at the new run's updated_at.
-- Disconfirming outcome, stated in advance: if content_data is STILL NULL after
-- this, the writer's reply is not reaching the save step by this path and the key
-- name is not the fault.

BEGIN;

SELECT snapshot_agent('page-rebuild',
    'pre-update: bugs_open/194 — save_sections never mapped sections_metadata, so every rebuild NULLed content_data');

-- updated_at is set explicitly: there is NO trigger on agent_definitions (verified
-- 2026-08-04 — pg_trigger has no non-internal row for the table), so the column is
-- current only if the seed sets it, and a stale one reads as "nobody has touched
-- this row" to the next session.
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,sections_metadata_field}',
        '"page_content.response.sections_metadata"'::jsonb,
        true),
    updated_at = NOW()
WHERE type = 'page-rebuild'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

DO $$
DECLARE
    cfg jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections}'
    INTO cfg
    FROM agent_definitions
    WHERE type = 'page-rebuild'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '194/310: page-rebuild has no build_pages_loop.save_sections step';
    END IF;

    -- IS DISTINCT FROM, not <>: a missing jsonb path is NULL and `NULL <> 'x'` is
    -- NULL, so a plain <> against an absent key can never fire. (The trap that sat
    -- green in seed 309's first draft.)
    IF cfg #>> '{config,sections_metadata_field}'
       IS DISTINCT FROM 'page_content.response.sections_metadata' THEN
        RAISE EXCEPTION '194/310: sections_metadata_field is %, expected page_content.response.sections_metadata',
            cfg #>> '{config,sections_metadata_field}';
    END IF;

    -- the three pre-existing keys must survive: this is an ADD, not a rewrite
    IF cfg #>> '{config,html_field}'         IS DISTINCT FROM 'assembled_page.html'
       OR cfg #>> '{config,site_id_field}'   IS DISTINCT FROM 'site_record.site_id'
       OR cfg #>> '{config,page_name_field}' IS DISTINCT FROM 'current_page.name' THEN
        RAISE EXCEPTION '194/310: an existing save_sections config key was disturbed — %', cfg->'config';
    END IF;

    IF cfg->>'action' IS DISTINCT FROM 'save_page_sections'
       OR cfg->>'next_step' IS DISTINCT FROM 'update_page_status' THEN
        RAISE EXCEPTION '194/310: save_sections step shape changed';
    END IF;

    RAISE NOTICE '194/310 OK — sections_metadata_field added, 3 existing keys intact';
END $$;

COMMIT;

-- ROLLBACK if needed:
--   UPDATE agent_definitions SET default_config = default_config
--     #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,save_sections,config,sections_metadata_field}'
--   WHERE type='page-rebuild' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
