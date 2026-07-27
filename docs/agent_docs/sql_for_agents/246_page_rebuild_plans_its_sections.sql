-- 246_page_rebuild_plans_its_sections.sql
--
-- bugs_open/087 — `page-rebuild` gives its writer no section plan, and nothing
-- in the writer builds one. Fix candidate A, the file's own smallest change.
--
-- THE DEFECT. `page-content-writer.select_sections` is an `extract_fields` over
-- two sources — the link resolver's output, and `input_data.section_plan.
-- sections_ready`. On the rebuild path BOTH are empty: page-rebuild supplies no
-- `section_plan`, so the resolver returns an empty `sections_ready`, and
-- page-rebuild's `write_page_content` input_mapping never mapped one.
-- `extract_fields` then writes `sections_for_render` WITHOUT the key, and the
-- loop over `sections_for_render.sections_ready` dies:
--     key 'sections_ready' not found at position 1 in path
--     'sections_for_render.sections_ready'
--
-- WHY THIS IS SAFE — the two open questions the case named as blockers, both
-- settled by reading the action (2026-07-27), not by trying it:
--
--   Q1 "does plan_sections tolerate an absent work_item_id?"  YES.
--      plan_sections_action.go:50-52 — Required is ["site_id"] ALONE;
--      work_item_id is Optional. createDeferredItems (:1824-1830) guards
--      `if parentWorkItemID != ""`, so an absent one leaves parentID nil and
--      deferred items simply get no parent. The rebuild flow has no work item
--      and does not need one.
--
--   Q2 "where does `sections` come from on the rebuild path?"  current_page.sections,
--      which is EXACTLY the shape the action documents for itself. Its own
--      header example (:22-25) is `"sections": "page_record.sections"`, and the
--      parser (:649-664) accepts []interface{} of strings, []string, or a JSON
--      string. `current_page.sections` is ["hero-about","content-block-about",…]
--      — the first case. filterSiteLevelSections then strips any header/footer
--      names, which is the one hazard in feeding it a raw page section list.
--
-- The case worried these two might block candidate A. Neither does.
--
-- BONUS, worth knowing: even in the worst case this is an improvement. If
-- `current_page.sections` were empty, plan_sections returns
-- `{"sections_ready": [], …, "reason": "no sections to plan"}` (:673-681) — the
-- KEY EXISTS. So the loop iterates zero times instead of dying on a missing
-- path. The fatal error this bug reports cannot recur in that shape.
--
-- BLAST RADIUS: page-rebuild only, and it is DORMANT — 0 runs in the 13 days
-- orchestration_states retains, other than the case's own test. The routine path
-- (build-dispatch-loop → page-build-handler) already runs plan_sections and is
-- untouched by this.
--
-- Config-only: live on apply, no image roll. Reversible via the snapshot below.
--
-- NOT DONE HERE: candidates B (route rebuilds through page-build-handler) and C
-- (retire page-rebuild entirely). Both change what other lanes' scripts do and
-- the case says they want an owner call. C is the one that removes the class.

BEGIN;

SELECT snapshot_agent('page-rebuild',
    'pre-update: bugs_open/087 candidate A — give the rebuild loop a plan_sections step');

-- 1. Insert the plan_sections step into the loop's sub-workflow.
--    A targeted ADD of one new key: it cannot disturb its siblings, unlike a
--    jsonb_set writing a literal object at the parent `steps` path.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,build_pages_loop,config,sub_workflow,steps,plan_sections}',
      '{
         "action": "plan_sections",
         "config": {
           "site_id":   "site_record.site_id",
           "sections":  "current_page.sections",
           "page_name": "current_page.name"
         },
         "next_step": "write_page_content",
         "description": "Resolve section data requirements for this page so the writer receives a real section_plan (bugs_open/087)",
         "output_field": "section_plan"
       }'::jsonb,
      true)
WHERE type='page-rebuild' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 2. Make it the entry point. It was write_page_content.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,build_pages_loop,config,sub_workflow,start_step}',
      '"plan_sections"'::jsonb, false)
WHERE type='page-rebuild' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- 3. Hand the plan to the writer. One key added to input_mapping — NOT a
--    replacement of the map, whose eight existing entries must all survive.
UPDATE agent_definitions
SET default_config = jsonb_set(default_config,
      '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config,input_mapping,section_plan}',
      '"section_plan"'::jsonb, true),
    updated_at = NOW()
WHERE type='page-rebuild' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

-- ── VERIFY (all three must read correctly before COMMIT) ────────────────────
\echo '=== 1. start_step must now be plan_sections ==='
SELECT default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->>'start_step' AS start_step
FROM agent_definitions WHERE type='page-rebuild' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

\echo '=== 2. the step graph — plan_sections must head it ==='
SELECT k||'  ->  '||coalesce(v->>'next_step','(terminal)') AS step
FROM agent_definitions,
     jsonb_each(default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps') AS e(k,v)
WHERE type='page-rebuild' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
ORDER BY 1;

\echo '=== 3. input_mapping must have NINE keys — the original eight PLUS section_plan ==='
SELECT count(*) AS mapping_keys,
       bool_or(k='section_plan') AS has_section_plan
FROM agent_definitions,
     jsonb_object_keys(default_config->'workflow'->'steps'->'build_pages_loop'->'config'->'sub_workflow'->'steps'->'write_page_content'->'config'->'input_mapping') AS k
WHERE type='page-rebuild' AND is_active
  AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;

COMMIT;

-- ── ROLLBACK, if the live verification fails ────────────────────────────────
-- restore_agent_snapshot() if it exists, or by hand:
--   UPDATE agent_definitions SET default_config = jsonb_set(
--     default_config #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,plan_sections}'
--                    #- '{workflow,steps,build_pages_loop,config,sub_workflow,steps,write_page_content,config,input_mapping,section_plan}',
--     '{workflow,steps,build_pages_loop,config,sub_workflow,start_step}', '"write_page_content"'::jsonb, false)
--   WHERE type='page-rebuild' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
