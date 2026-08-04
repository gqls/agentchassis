-- 309_page_content_writer_plans_its_own_sections.sql
--
-- bugs_open/087 — "page-rebuild gives its writer no section plan, and nothing in
-- the writer builds one." Candidate D: make the WRITER build one, so no caller
-- can get it wrong.
--
-- THE DEFECT, restated at the level it actually lives. page-content-writer's
-- select_sections is an extract_fields whose every configured path is rooted in
-- either the link resolver's output or `input_data.section_plan`. A caller that
-- dispatches the writer without a section_plan produces no `sections_ready`, and
-- process_sections_loop — a loop over `sections_for_render.sections_ready` — dies:
--     key 'sections_ready' not found at position 1 in path
--     'sections_for_render.sections_ready'
-- A section_plan entry is not a section NAME; it is the rich object plan_sections
-- produces. No input-mapping one-liner can conjure one. Something must RUN
-- plan_sections.
--
-- WHY THIS SEED AND NOT A THIRD COPY OF MIGRATION 246. 246 applied the case's
-- candidate A to ONE caller. Measured on the live rows today (2026-08-04), four
-- agent_definitions reference page-content-writer:
--
--     page-build-handler       runs plan_sections, maps section_plan   OK
--     page-rebuild             runs plan_sections, maps section_plan   OK (mig 246)
--     pageflow-builder         NEITHER — build_pages_loop.start_step
--                              is write_page_content                    BROKEN
--     site-work-orchestrator   NEITHER — same shape, current_page
--                              comes from current_item.spec             BROKEN
--
-- Both broken ones are reachable: 075d_simple_maintain_trigger.sh dispatches
-- site-work-orchestrator directly, and pageflow-builder is the only value ever
-- recorded in site_specs.recommended_builder (1,216 rows, 2026-08-02). Patching
-- them the same way would leave FOUR hand-maintained copies of one planning step
-- and the door still open for caller #5. This seed closes the door instead.
--
--   NOTE ON THE CENSUS ABOVE, because the trap is silent: `default_config::text
--   LIKE '%plan_sections%'` is NOT a safe test for that step — `_` is a
--   single-character wildcard, so it also matches the substring "plan.sections"
--   inside `section_plan.sections_ready`, and page-content-writer reads as a
--   false positive. The table above was read from the step keys themselves.
--
-- THE CHANGE. Four new steps, one rewired scalar, one appended fallback path:
--
--     build_render_context  --next_step-->  check_section_plan   (was: resolve_links)
--
--     check_section_plan       conditional_branch
--       condition "input_data.section_plan.sections_ready"
--         truthy -> resolve_links      caller's plan kept VERBATIM
--         falsy  -> plan_sections
--     plan_sections            output_field: section_plan
--         -> check_planned_sections
--     check_planned_sections   conditional_branch
--       condition "section_plan.ready_count > 0"
--         truthy -> resolve_links
--         falsy  -> fail_no_ready_sections
--     fail_no_ready_sections   fail_workflow
--
--   + select_sections gains a FOURTH fallback path, "section_plan.sections_ready".
--
-- WHY A CONDITIONAL AND NOT AN UNCONDITIONAL PLAN. plan_sections calls
-- createDeferredItems. Running it when the caller already planned would
-- double-file deferred work items and discard a caller plan that may carry
-- content attachments (bugs_open/178's load_current_section_content). With the
-- conditional, the 100% of live traffic that arrives from page-build-handler
-- takes the truthy branch and behaves IDENTICALLY — one extra condition
-- evaluation, no new action run.
--
-- WHY check_planned_sections EXISTS — the anti-empty-page guard. With no ready
-- sections plan_sections returns {"sections_ready": [], "ready_count": 0,
-- "reason": "no sections to plan"} (plan_sections_action.go:869-873) — the KEY
-- EXISTS. So select_sections SUCCEEDS, the loop iterates zero times, compile_page
-- emits an empty page and the caller deploys it over a real one. page-build-handler
-- already refuses this case (check_has_ready_sections -> mark_no_ready_sections);
-- the writer had no equivalent. This is a DECLARED BEHAVIOUR CHANGE beyond 087's
-- own symptom: a writer call whose plan yields zero ready sections now FAILS
-- LOUDLY instead of producing an empty page. Unreachable from page-build-handler
-- (it gates before spawning the writer); reachable from page-rebuild and the two
-- broken callers, where an empty page overwriting a good one is the worse outcome.
--
-- WHY THE FOURTH select_sections PATH IS NEEDED. ExtractFieldsAction tries each
-- configured path directly and then with "input_data." PREFIXED — never stripped
-- (v3_site_actions.go:4284-4305). Existing paths 2 and 3 are already written with
-- the prefix, so NEITHER can ever reach a top-level section_plan. Appended, not
-- inserted, and guarded by a containment test, so it composes with the 192 lane's
-- owed post-roll removal of their path 3 in either order.
--
-- READ, NOT ASSUMED, before this seed was written:
--   * conditional_branch_action.go:305-315 — a bare dotted path with no operator
--     IS a truthiness check; valueIsTruthy (:527-551) is false for nil, "", empty
--     array, empty map, 0 and false. "Absent or empty" needs no new grammar and
--     no Go change.
--   * resolveFieldValue Strategy 5 (:396-411) recursively searches beneath the
--     base path, so a 192-WRAPPED caller plan still reads truthy — and it cannot
--     fire when no plan was supplied at all, because the base object is absent.
--   * datahelpers.ExtractNestedField (data_helpers.go:1199-1234) is strict
--     traversal + a .response unwrap, so path 4 resolves a top-level section_plan
--     exactly and nothing else.
--   * plan_sections_action.go:49-51 — Required is ["site_id"] ALONE; work_item_id
--     is Optional and createDeferredItems guards `if parentWorkItemID != ""`. The
--     file makes ZERO LLM calls: the self-plan branch is DB-only and costs no
--     credits.
--   * registry.go:65/71 — conditional_branch is canonical, `conditional` a
--     deprecated alias to the same handler (4 live steps use the canonical name,
--     113 the alias). fail_workflow is registered at :35 and live in
--     report-builder.fail_out.
--
-- DELIBERATELY NOT CHANGED — page-content-writer's resolve_links input_mapping
-- ("sections?": "input_data.section_plan.sections_ready"). On the self-plan branch
-- the resolver is handed nothing, because BOTH of FindByPath's prefix fallbacks
-- guard `i == 0` (content_search.go:70-95): `input_data` resolves at position 0,
-- so the miss at position 1 gets no fallback. Repointing it to the unprefixed
-- "section_plan.sections_ready" would fix that and is shape-agnostic (unlike the
-- nested repoint seed 308 correctly refused, it does not re-break on the roll).
-- It is OWED, not done, for three reasons: seed 308 pins that mapping deliberately
-- and the 192 lane is mid-flight on the Go half; the resulting degradation — no
-- internal CTA resolution — is exactly the cost the estate has ALREADY ACCEPTED
-- fleet-wide for the pre-roll window, so the self-plan branch is no worse than
-- every build running today; and the repoint trades an EXACT hit for a three-deep
-- FindByPath fallback on the one branch that currently works, which is not a trade
-- to make inside the change being used to prove 087.
--
-- BLAST RADIUS. Four callers, all measured. Truthy branch (page-build-handler,
-- page-rebuild) — no behaviour change. Falsy branch (pageflow-builder,
-- site-work-orchestrator, any direct dispatch) — from a guaranteed fatal to a
-- planned build, or to a named failure where it would have emitted an empty page.
--
-- Config-only: live on apply, no image roll — both actions are registered in the
-- running binary and exercised live today. Reversible via the snapshot below.
--
-- COUNCIL: out of scope. The gate reviews platform/, internal/, pkg/ (owner ruling
-- 2026-07-17) and 097_TRIGGER_council_review_v1.sh:127 refuses a submission
-- touching none of them, client-side, before spending credits. This is a seed.
--
-- VERIFY (see the DO block before COMMIT for the config-shape half; the live half
-- is bugs_open/087's own acceptance test, on a rebuild_policy='generic' page).

BEGIN;

SELECT snapshot_agent('page-content-writer',
    'pre-update: bugs_open/087 candidate D — the writer plans its own sections when a caller supplies none');

-- ============================================================================
-- 1. check_section_plan — keep a usable caller plan, otherwise plan locally.
--    A targeted ADD of one new key: it cannot disturb its siblings, unlike a
--    jsonb_set writing a literal object at the parent `steps` path.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_section_plan}',
        '{
            "action": "conditional_branch",
            "config": {
                "condition": "input_data.section_plan.sections_ready",
                "then_step": "resolve_links",
                "else_step": "plan_sections"
            },
            "description": "A caller-supplied usable section_plan is kept verbatim; anything else is planned locally (bugs_open/087)"
        }'::jsonb,
        true)
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- 2. plan_sections — build the plan the caller did not supply.
--    No error_step, deliberately: a planning failure should fail under its own
--    name, not degrade into an extract failure two steps later (bugs_closed/086).
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,plan_sections}',
        '{
            "action": "plan_sections",
            "config": {
                "site_id": "input_data.site_record.site_id",
                "sections": "input_data.current_page.sections",
                "page_name": "input_data.current_page.name"
            },
            "next_step": "check_planned_sections",
            "description": "Build the section plan the caller did not supply (bugs_open/087)",
            "output_field": "section_plan"
        }'::jsonb,
        true)
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- 3. check_planned_sections — refuse to compile an empty page.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,check_planned_sections}',
        '{
            "action": "conditional_branch",
            "config": {
                "condition": "section_plan.ready_count > 0",
                "then_step": "resolve_links",
                "else_step": "fail_no_ready_sections"
            },
            "description": "Zero ready sections would compile an EMPTY page and deploy it over a real one — fail instead (bugs_open/087)"
        }'::jsonb,
        true)
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- 4. fail_no_ready_sections — the deliberate FAILURE verdict.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,fail_no_ready_sections}',
        '{
            "action": "fail_workflow",
            "config": {
                "reason": "page-content-writer planned its own sections and none are ready — no page can be written. Check the page''s sections list and the components'' input_schema data requirements; see section_plan.reason and bugs_open/087."
            },
            "description": "End in a deliberate FAILURE rather than emitting an empty page (bugs_open/087)"
        }'::jsonb,
        true)
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- 5. Rewire build_render_context onto the new branch. One scalar.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,build_render_context,next_step}',
        '"check_section_plan"'::jsonb,
        false)
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND default_config #>> '{workflow,steps,build_render_context,next_step}' = 'resolve_links';

-- ============================================================================
-- 6. select_sections gains the ONLY path that can see a writer-local plan.
--    APPEND, guarded by containment, so it is commutative with the 192 lane's
--    owed removal of their path 3 — neither ordering can duplicate or drop it.
-- ============================================================================

UPDATE agent_definitions
SET default_config = jsonb_set(
        default_config,
        '{workflow,steps,select_sections,config,fields,sections_ready}',
        (default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}')
            || '["section_plan.sections_ready"]'::jsonb,
        false)
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL
  AND NOT ((default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}')
           @> '["section_plan.sections_ready"]'::jsonb);

-- ============================================================================
-- 7. Stamp updated_at. There is NO trigger on agent_definitions — verified
--    2026-08-04 (`SELECT tgname FROM pg_trigger WHERE tgrelid='agent_definitions'
--    ::regclass AND NOT tgisinternal` returns NOTHING) — so the column is current
--    only if a seed sets it. Seeds 246 and 308 both do; the six statements above
--    did not, which left the row reading 09:01:35Z (another lane's write) while
--    carrying this lane's changes. That is the exact signal the next session uses
--    to ask "has anyone touched this?", so leaving it stale is not cosmetic.
-- ============================================================================

UPDATE agent_definitions
SET updated_at = NOW()
WHERE type = 'page-content-writer'
  AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

-- ============================================================================
-- VERIFY inside the transaction. A DO block that RAISES — a verify block made of
-- SELECTs cannot stop a COMMIT, because ON_ERROR_STOP ignores a non-empty result
-- (the RFC_006 lesson, CLAUDE.md / MEMORY).
-- ============================================================================

DO $$
DECLARE
    cfg          jsonb;
    step_count   int;
    paths        jsonb;
BEGIN
    SELECT default_config->'workflow' INTO cfg
    FROM agent_definitions
    WHERE type = 'page-content-writer'
      AND is_active AND COALESCE(is_snapshot, false) = false AND deleted_at IS NULL;

    IF cfg IS NULL THEN
        RAISE EXCEPTION '087/309: no live page-content-writer row';
    END IF;

    -- the four new steps exist
    IF NOT (cfg->'steps' ? 'check_section_plan'
        AND cfg->'steps' ? 'plan_sections'
        AND cfg->'steps' ? 'check_planned_sections'
        AND cfg->'steps' ? 'fail_no_ready_sections') THEN
        RAISE EXCEPTION '087/309: one of the four new steps is missing';
    END IF;

    SELECT count(*) INTO step_count FROM jsonb_object_keys(cfg->'steps');
    IF step_count <> 14 THEN
        RAISE EXCEPTION '087/309: expected 14 steps (10 original + 4), found %', step_count;
    END IF;

    -- the entry point is untouched.
    -- NOTE — every comparison below is IS DISTINCT FROM, not <>. A missing jsonb
    -- path yields NULL, and `NULL <> 'x'` is NULL, not TRUE, so a plain <> check
    -- against a key that does not exist can NEVER fire: it would read as a pass
    -- whatever the truth. That is a check whose result could not have come out
    -- otherwise, which is worth no more than no check at all. Caught here by
    -- exactly that: the loop's key is `iterate_over`, and an `items_field`
    -- assertion sat green against NULL until the live row was read.
    IF cfg->>'start_step' IS DISTINCT FROM 'spawn_research_agent' THEN
        RAISE EXCEPTION '087/309: start_step is % — it must stay spawn_research_agent',
            cfg->>'start_step';
    END IF;

    -- the graph is wired as designed
    IF cfg #>> '{steps,build_render_context,next_step}' IS DISTINCT FROM 'check_section_plan' THEN
        RAISE EXCEPTION '087/309: build_render_context.next_step is %, expected check_section_plan',
            cfg #>> '{steps,build_render_context,next_step}';
    END IF;
    IF cfg #>> '{steps,check_section_plan,config,then_step}' IS DISTINCT FROM 'resolve_links'
       OR cfg #>> '{steps,check_section_plan,config,else_step}' IS DISTINCT FROM 'plan_sections' THEN
        RAISE EXCEPTION '087/309: check_section_plan branches wrong';
    END IF;
    IF cfg #>> '{steps,plan_sections,next_step}' IS DISTINCT FROM 'check_planned_sections'
       OR cfg #>> '{steps,plan_sections,output_field}' IS DISTINCT FROM 'section_plan' THEN
        RAISE EXCEPTION '087/309: plan_sections wired wrong';
    END IF;
    IF cfg #>> '{steps,check_planned_sections,config,then_step}' IS DISTINCT FROM 'resolve_links'
       OR cfg #>> '{steps,check_planned_sections,config,else_step}' IS DISTINCT FROM 'fail_no_ready_sections' THEN
        RAISE EXCEPTION '087/309: check_planned_sections branches wrong';
    END IF;

    -- resolve_links is REACHED, and its mapping is untouched (see the header)
    IF cfg #>> '{steps,resolve_links,next_step}' IS DISTINCT FROM 'select_sections' THEN
        RAISE EXCEPTION '087/309: resolve_links.next_step moved';
    END IF;
    IF cfg #>> '{steps,resolve_links,config,input_mapping,sections?}'
       IS DISTINCT FROM 'input_data.section_plan.sections_ready' THEN
        RAISE EXCEPTION '087/309: resolve_links input_mapping was changed — this seed must not touch it';
    END IF;

    -- select_sections: the 192 lane's paths survive, ours is appended last,
    -- and their `required` opt-in is intact
    paths := cfg #> '{steps,select_sections,config,fields,sections_ready}';
    IF paths IS NULL OR jsonb_typeof(paths) IS DISTINCT FROM 'array' THEN
        RAISE EXCEPTION '087/309: select_sections sections_ready is not an array of paths';
    END IF;
    IF paths->>0 IS DISTINCT FROM 'resolved_links.response.link_resolution.sections_ready' THEN
        RAISE EXCEPTION '087/309: select_sections path 1 changed — the resolver must stay first';
    END IF;
    IF NOT (paths @> '["input_data.section_plan.sections_ready"]'::jsonb) THEN
        RAISE EXCEPTION '087/309: select_sections lost the caller-plan path';
    END IF;
    IF NOT (paths @> '["section_plan.sections_ready"]'::jsonb) THEN
        RAISE EXCEPTION '087/309: select_sections did not gain the writer-local path';
    END IF;
    IF paths->>(jsonb_array_length(paths) - 1) IS DISTINCT FROM 'section_plan.sections_ready' THEN
        RAISE EXCEPTION '087/309: the writer-local path must be LAST, found at another index';
    END IF;
    IF NOT COALESCE(cfg #> '{steps,select_sections,config,required}' @> '["sections_ready"]'::jsonb, false) THEN
        RAISE EXCEPTION '087/309: select_sections lost the 192 lane''s required opt-in';
    END IF;

    -- the loop itself is byte-untouched. The key is `iterate_over`, NOT
    -- `items_field` — page-rebuild's build_pages_loop uses the latter, this one
    -- does not, and asserting the wrong name is a check that cannot fail.
    IF cfg #>> '{steps,process_sections_loop,config,iterate_over}'
       IS DISTINCT FROM 'sections_for_render.sections_ready' THEN
        RAISE EXCEPTION '087/309: process_sections_loop iterate_over is %, expected sections_for_render.sections_ready',
            cfg #>> '{steps,process_sections_loop,config,iterate_over}';
    END IF;

    RAISE NOTICE '087/309 OK — 14 steps, branch wired, select_sections has % paths, resolve_links untouched',
        jsonb_array_length(paths);
END $$;

COMMIT;

-- ============================================================================
-- ROLLBACK, if needed (forward-only applies to git, not to a live config row):
--
--   BEGIN;
--   UPDATE agent_definitions SET default_config =
--       jsonb_set(
--         (default_config #- '{workflow,steps,check_section_plan}'
--                        #- '{workflow,steps,plan_sections}'
--                        #- '{workflow,steps,check_planned_sections}'
--                        #- '{workflow,steps,fail_no_ready_sections}'),
--         '{workflow,steps,build_render_context,next_step}', '"resolve_links"'::jsonb, false)
--   WHERE type='page-content-writer' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   -- then drop the last select_sections path:
--   UPDATE agent_definitions SET default_config = jsonb_set(default_config,
--       '{workflow,steps,select_sections,config,fields,sections_ready}',
--       (default_config #> '{workflow,steps,select_sections,config,fields,sections_ready}') - 'section_plan.sections_ready',
--       false)
--   WHERE type='page-content-writer' AND is_active
--     AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
--   COMMIT;
--
-- Or restore the snapshot taken at the top of this file.
-- ============================================================================
