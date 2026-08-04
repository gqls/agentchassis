-- 308_select_sections_tolerates_the_wrapper_and_opts_into_required.sql
--
-- bugs_open/192 — every page build in the fleet fails at process_sections_loop
-- with "key 'sections_ready' not found at position 1 in path
-- 'sections_for_render.sections_ready'".
--
-- Cause: page-build-handler's load_current_section_content step (seed 299,
-- bugs_open/178) declares output_field: section_plan, and the action returns
-- {section_plan, applied, reason|matched} on EVERY path including its eight
-- "pass-throughs". coordinator.go storeActionResult stores a return value
-- WHOLESALE under output_field, so the flat plan is replaced by a wrapper on
-- every build in every mode. page-content-writer's select_sections then finds
-- nothing at input_data.section_plan.sections_ready (the plan moved to
-- .section_plan.section_plan.sections_ready), writes an empty
-- sections_for_render, and the NEXT step fails naming the missing key rather
-- than the failed extraction.
--
-- ORDER — this seed deliberately runs BEFORE its image, inverting the house
-- "image first, then seeds" rule. That rule exists because a seed naming an
-- unregistered ACTION fails at runtime. This seed names no action: change 1 is
-- plain data, and change 2 sets a config key the running binary provably
-- ignores (ExtractFieldsAction reads only "fields", "field_map" and "defaults"
-- — v3_site_actions.go:4244/4272/4333/4344). Running it now restores page
-- builds on the binary already deployed; the Go fix follows on the next roll.
--
-- 1. A THIRD fallback path that tolerates the wrapper. TEMPORARY. Fallback
--    paths are tried in order, so the moment the Go fix rolls and the plan is
--    flat again, path 2 wins and path 3 is structurally dead. Remove it in a
--    follow-up seed once the roll is pod-verified — it is folklore otherwise.
--
-- 2. Opt select_sections into extract_fields' new "required" list. INERT until
--    the roll, then permanent: it converts "no path resolved" from a silent
--    omission that fails two steps later into a step failure that names the
--    field, the paths tried and what was in scope. Default OFF fleet-wide, per
--    the 2026-08-02 owner ruling on shared-seam authority.
--
-- NOT changed, deliberately: page-content-writer's resolve_links input_mapping
-- ("sections?": "input_data.section_plan.sections_ready") is broken by the same
-- wrapper, which is why path 1 comes back null and the link resolver is handed
-- no sections. It is left alone because input_mapping has no ordered fallback:
-- pointing it at the nested path would fix it today and SILENTLY re-break it
-- the moment the Go fix lands, with no error to notice. It self-heals on the
-- roll. Until then internal CTA resolution is degraded on every build — a known,
-- stated cost of the pre-roll window, tracked in bugs_open/192.
--
-- Verify: bugs_open/192's own check — a dispatched build reaches compile_page,
-- and collected_data ? 'sections_for_render' carries a non-empty sections_ready.

BEGIN;

-- ============================================================================
-- select_sections: third fallback path + required
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,select_sections,config}',
        (default_config #> '{workflow,steps,select_sections,config}')
        || jsonb_build_object(
            'fields', jsonb_build_object(
                'sections_ready', jsonb_build_array(
                    'resolved_links.response.link_resolution.sections_ready',
                    'input_data.section_plan.sections_ready',
                    -- TEMPORARY (bugs_open/192): tolerates the wrapper the
                    -- running binary still writes. Dead once the unwrap rolls.
                    'input_data.section_plan.section_plan.sections_ready'
                )
            ),
            'required', jsonb_build_array('sections_ready')
        )
    ),
    updated_at = NOW()
WHERE type = 'page-content-writer'
  AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- Verification is a DO/RAISE, not a bare SELECT: ON_ERROR_STOP ignores a
-- non-empty result set, so a SELECT below the UPDATE cannot stop the COMMIT.
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
      AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF paths IS NULL OR jsonb_array_length(paths) <> 3 THEN
        RAISE EXCEPTION 'select_sections fallback list is % , want 3 paths', paths;
    END IF;
    IF paths->>0 <> 'resolved_links.response.link_resolution.sections_ready' THEN
        RAISE EXCEPTION 'path 1 reordered (%) — the resolver-augmented path must stay first', paths->>0;
    END IF;
    IF paths->>1 <> 'input_data.section_plan.sections_ready' THEN
        RAISE EXCEPTION 'path 2 reordered (%) — the FLAT path must precede the wrapper shim, or the shim never dies', paths->>1;
    END IF;
    IF paths->>2 <> 'input_data.section_plan.section_plan.sections_ready' THEN
        RAISE EXCEPTION 'path 3 (the wrapper shim) missing or wrong: %', paths->>2;
    END IF;
    IF req IS NULL OR NOT (req ? 'sections_ready') THEN
        RAISE EXCEPTION 'required did not take: %', req;
    END IF;
END $$;

-- Nothing else in the fleet opts in, and nothing else may have been disturbed.
DO $$
DECLARE
    opted_in int;
BEGIN
    SELECT count(*) INTO opted_in
    FROM agent_definitions ad, jsonb_each(ad.default_config->'workflow'->'steps') s
    WHERE ad.is_active AND COALESCE(ad.is_snapshot, false) = false AND ad.deleted_at IS NULL
      AND s.value->>'action' = 'extract_fields'
      AND s.value->'config' ? 'required';

    IF opted_in <> 1 THEN
        RAISE EXCEPTION 'expected exactly 1 extract_fields step opted into required (page-content-writer/select_sections), found %', opted_in;
    END IF;
END $$;

SELECT jsonb_pretty(default_config #> '{workflow,steps,select_sections,config}') AS select_sections_config
FROM agent_definitions
WHERE type = 'page-content-writer'
  AND is_active = true AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

COMMIT;
