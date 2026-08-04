-- 311_retire_the_192_wrapper_shim_path.sql
--
-- bugs_open/192, cleanup. Seed 308 added a THIRD fallback path to
-- page-content-writer.select_sections:
--
--     input_data.section_plan.section_plan.sections_ready
--
-- It existed only to tolerate the wrapper that load_current_section_content
-- wrote while the old binary was still serving — it declared itself TEMPORARY
-- and self-retiring, because fallback paths are tried in order and the FLAT
-- path sits ahead of it. The Go fix is now live (chassis v1.0.1250, both
-- replicas pod-verified 2026-08-04: "resolved via no configured path" 1/1,
-- "keys present at this level" 2/2, positive control
-- "ExtractFieldsAction: Found via input_data prefix" 1/1), so the wrapper is
-- gone at source and this path is dead config. Dead config is folklore: the
-- next reader cannot tell a live fallback from a fossil.
--
-- REMOVAL IS BY VALUE, NOT BY INDEX, and that is load-bearing. Seed 309
-- (another lane, page-content-writer plans its own sections) APPENDS a fourth
-- path `section_plan.sections_ready` to the same array. Whether 309 lands
-- before or after this file is not knowable from here — it was unrecorded in
-- schema_migrations when this was written — so an index-based delete would
-- remove whichever element happened to sit at that position. Filtering on the
-- literal is correct under either ordering, and is idempotent.
--
-- This file deliberately does NOT assert a final array length: 309 may or may
-- not have run. It asserts the two things that actually matter — the shim is
-- gone, and the flat path that replaces it is present.
--
-- Number claimed from the LEDGER, not from `ls`: schema_migrations' highest was
-- 308 (this lane's own), while 309 and 310 already existed on disk unrecorded
-- from other lanes. A number on disk is not a number taken, and a number taken
-- is not a number free.

BEGIN;

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,select_sections,config,fields,sections_ready}',
        (
            SELECT COALESCE(jsonb_agg(p ORDER BY ord), '[]'::jsonb)
            FROM jsonb_array_elements(
                     default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}'
                 ) WITH ORDINALITY AS t(p, ord)
            WHERE p <> '"input_data.section_plan.section_plan.sections_ready"'::jsonb
        ),
        false),
    updated_at = NOW()
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND (default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}')
      @> '["input_data.section_plan.section_plan.sections_ready"]'::jsonb;

-- Verify with DO/RAISE, never bare SELECTs: ON_ERROR_STOP ignores a non-empty
-- result set, so a SELECT below an UPDATE cannot stop the COMMIT.
DO $$
DECLARE
    paths jsonb;
    req   jsonb;
BEGIN
    SELECT default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}',
           default_config #> '{workflow,steps,select_sections,config,required}'
    INTO paths, req
    FROM agent_definitions
    WHERE type = 'page-content-writer'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF paths @> '["input_data.section_plan.section_plan.sections_ready"]'::jsonb THEN
        RAISE EXCEPTION 'the 192 wrapper shim is still present: %', paths;
    END IF;

    -- The path the shim was standing in for MUST still be there, or this file
    -- has removed the fallback instead of the fossil.
    IF NOT (paths @> '["input_data.section_plan.sections_ready"]'::jsonb) THEN
        RAISE EXCEPTION 'the FLAT path is gone — removal hit the wrong element: %', paths;
    END IF;

    -- And the resolver-augmented path must still lead, or link resolution is
    -- silently skipped on every build.
    IF paths->>0 <> 'resolved_links.response.link_resolution.sections_ready' THEN
        RAISE EXCEPTION 'path 1 is no longer the resolver-augmented path: %', paths->>0;
    END IF;

    -- 308's opt-in must survive this edit. It is what makes a future shape
    -- drift fail AT the extraction instead of two steps downstream.
    IF req IS NULL OR NOT (req ? 'sections_ready') THEN
        RAISE EXCEPTION 'the required opt-in was lost: %', req;
    END IF;
END $$;

SELECT jsonb_pretty(default_config #> '{workflow,steps,select_sections,config}') AS select_sections_config
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;
