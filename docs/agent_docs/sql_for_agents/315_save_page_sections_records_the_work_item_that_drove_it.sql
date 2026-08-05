-- 315_save_page_sections_records_the_work_item_that_drove_it.sql
--
-- bugs_closed/156 § "What remains", item 1. `save_page_sections` reads
-- `work_item_id_field` from its step config and uses it for TWO things:
--
--   1. `page_component_history.source_item_id` — attribution on the destructive
--      overwrite, which was added precisely so tracing a bad rebuild does not
--      have to be forensic-by-timing;
--   2. `context.work_item_id` in the duplicate-section record
--      (`CONTENT_DUPLICATE_SECTIONS_COLLAPSED`, register PBP-033).
--
-- The key was set by NOBODY — 6 of 6 call sites UNSET, measured 2026-08-05. So
-- both of those are NULL/empty on every save in the fleet.
--
-- HOW IT WAS FOUND, because it is the argument for behavioural induction and not
-- for more unit tests: 156's guard was induced in production on 2026-08-05 and
-- the record came back correct in every field EXCEPT `work_item_id`, which was
-- empty. Every unit test supplies the step config itself, so all 14 passed while
-- the field designed for producer-hunting arrived blank on the real path. Only
-- running it against the LIVE config showed it.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- THE CENSUS, AND THE TWO TRAPS IN TAKING IT
--
-- Trap 1 — the step is NESTED inside a loop sub_workflow in three of the six
-- callers, so a top-level `jsonb_object_keys(default_config->'workflow'->'steps')`
-- finds only THREE and reports a clean-looking half-answer. Same trap seed 312
-- records. Walk the whole document:
--
--     SELECT type FROM agent_definitions
--      WHERE default_config::text LIKE '%save_page_sections%'
--        AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--
-- Trap 2 — and this is why this seed does NOT set one value everywhere:
-- **THE CORRECT PATH IS NOT THE SAME FOR ALL CALLERS.** site-work-orchestrator's
-- `build_items_loop` iterates over `work_items.items`, so inside its sub_workflow
-- the work item IS `current_item` — proven by its sibling steps in the same
-- sub_workflow, which pass `work_item_id: current_item.id` to `fail_work_item`
-- and `complete_work_item`. Setting `input_data.work_item_id` there would be a
-- DEAD CONFIG KEY THAT LOOKS LIVE: harmless at runtime, and permanently
-- misleading to the next reader, who would see the key set and assume attribution
-- works. pageflow-builder and page-rebuild loop over `pages_to_build.pages`
-- instead, so `current_item` is a PAGE and the same substitution would be wrong
-- in the other direction.
--
-- ─────────────────────────────────────────────────────────────────────────────
-- WHAT THIS SEED CHANGES — three of six, each on evidence, not on symmetry
--
--   page-rerender          input_data.work_item_id   MEASURED: 736 of 776 retained
--                                                    runs carry it (the 40 without
--                                                    are the direct-spawn path,
--                                                    e.g. rerender_page_vonc.sh,
--                                                    which has no work item — the
--                                                    field is correctly empty there)
--   page-build-handler     input_data.work_item_id   MEASURED: 74 of 74
--   site-work-orchestrator current_item.id           PROVEN from the sub_workflow:
--                                                    its own fail/complete steps
--                                                    already treat current_item.id
--                                                    as the work item id
--
-- WHAT IT DELIBERATELY LEAVES UNSET, and why that is not laziness:
--
--   pageflow-builder, page-rebuild, tool-recreation-handler — **ZERO retained
--   orchestration rows** for all three, so there is no evidence the path resolves.
--   `orchestration_states` keeps terminal rows ~24h. `input_data.work_item_id` is
--   the only plausible path for the first two (their loop is over pages), but
--   "plausible" is exactly what the dead-key trap is made of. Close the gap by
--   MEASURING when they next run, then extend this seed:
--
--     SELECT initial_request_data->'__execution_context__'->'sender'->>'agent_type' AS agent,
--            count(*) AS runs,
--            count(*) FILTER (WHERE collected_data->'input_data'->>'work_item_id' IS NOT NULL) AS with_id
--       FROM orchestration_states
--      WHERE initial_request_data->'__execution_context__'->'sender'->>'agent_type'
--            IN ('pageflow-builder','page-rebuild','tool-recreation-handler')
--      GROUP BY 1;
--
-- BLAST RADIUS: additive. The key is read with ExtractNestedFieldString and an
-- unresolved path yields "", which leaves `sourceItemID` nil — i.e. exactly the
-- behaviour every caller has today. Nothing can regress; the only change is that
-- a value that was always NULL can now be non-NULL.
--
-- CONFIG, so LIVE IMMEDIATELY on the next dispatch — no chassis roll.
--
-- APPLY:
--   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
--     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - < 315_….sql

BEGIN;

-- 1. page-rerender — the caller whose induction exposed the gap
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,save_sections,config,work_item_id_field}',
        '"input_data.work_item_id"'::jsonb,
        true)
WHERE type = 'page-rerender'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 2. page-build-handler — 74 of 74 retained runs carry the value
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,save_sections,config,work_item_id_field}',
        '"input_data.work_item_id"'::jsonb,
        true)
WHERE type = 'page-build-handler'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- 3. site-work-orchestrator — DIFFERENT PATH ON PURPOSE. Its loop iterates over
--    work items, so current_item IS the work item (its own fail_work_item and
--    complete_work_item steps in this same sub_workflow prove it).
UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,work_item_id_field}',
        '"current_item.id"'::jsonb,
        true)
WHERE type = 'site-work-orchestrator'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- VERIFY as DO/RAISE, never a block of SELECTs: ON_ERROR_STOP does not fire on a
-- non-empty result set, so a SELECT-based verify block cannot stop the COMMIT.
DO $$
DECLARE v text;
BEGIN
    SELECT default_config #>> '{workflow,steps,save_sections,config,work_item_id_field}'
      INTO v FROM agent_definitions WHERE type='page-rerender'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v IS DISTINCT FROM 'input_data.work_item_id' THEN
        RAISE EXCEPTION '315: page-rerender work_item_id_field is %, expected input_data.work_item_id', v;
    END IF;

    SELECT default_config #>> '{workflow,steps,save_sections,config,work_item_id_field}'
      INTO v FROM agent_definitions WHERE type='page-build-handler'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v IS DISTINCT FROM 'input_data.work_item_id' THEN
        RAISE EXCEPTION '315: page-build-handler work_item_id_field is %, expected input_data.work_item_id', v;
    END IF;

    SELECT default_config #>> '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,config,work_item_id_field}'
      INTO v FROM agent_definitions WHERE type='site-work-orchestrator'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v IS DISTINCT FROM 'current_item.id' THEN
        RAISE EXCEPTION '315: site-work-orchestrator work_item_id_field is %, expected current_item.id (its loop is over WORK ITEMS, not pages)', v;
    END IF;

    -- Guard the thing that actually breaks if a path is fat-fingered: the step
    -- must still BE a save_page_sections step after the jsonb_set. create_missing
    -- happily builds a whole branch of nonsense from a wrong path, which is the
    -- failure mode this check exists for.
    SELECT default_config #>> '{workflow,steps,save_sections,action}'
      INTO v FROM agent_definitions WHERE type='page-rerender'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v IS DISTINCT FROM 'save_page_sections' THEN
        RAISE EXCEPTION '315: page-rerender save_sections action is now %, the path was wrong', v;
    END IF;

    SELECT default_config #>> '{workflow,steps,build_items_loop,config,sub_workflow,steps,save_sections,action}'
      INTO v FROM agent_definitions WHERE type='site-work-orchestrator'
       AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
    IF v IS DISTINCT FROM 'save_page_sections' THEN
        RAISE EXCEPTION '315: site-work-orchestrator nested save_sections action is now %, the path was wrong', v;
    END IF;

    RAISE NOTICE '315: three call sites set, each on its own measured path; three left UNSET with no retained runs to verify them';
END $$;

COMMIT;
