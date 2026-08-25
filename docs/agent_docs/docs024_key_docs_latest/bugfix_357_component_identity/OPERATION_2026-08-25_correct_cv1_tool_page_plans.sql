-- bugs_open/357 — correct the two cv1.co.uk tool-page plans, and re-queue their recreation.
--
-- NOT a platform migration and deliberately NOT numbered into sql_for_agents/: this is a
-- one-off DATA correction on two rows of ONE site, run to give phase 2 its first firing.
-- Owner decision 2026-08-25 ("correct the two cv1 plans"), taken over the two alternatives
-- (lower the floor for tool recreation; fix apply_adoption_plan first).
--
-- WHY THIS IS A CORRECTION AND NOT A WORKAROUND. apply_adoption_plan classified both pages
-- as tools and queued them to tool-recreation-handler, whose save can only ever emit ONE
-- section; the same action, in the same transaction, wrote them multi-entry `pages.sections`
-- plans. Its own analysis spec for tool-example says `"self_contained": true`. The plan and
-- the route contradict each other, and the route is the one borne out by the artefact: the
-- recreation produced a single 21,265-byte self-contained fragment. This brings the plan
-- into line with what the page IS.
--
-- WHAT IT DOES NOT DO: it does not touch prune_floor_ratio, any agent config, any other
-- site, or any page_components row. Both pages hold ZERO component rows, so there is
-- nothing to lose. `planned[0]` is PRESERVED on both, which matters: the slot name comes
-- from it, so index keeps `hero` and therefore reproduces the real 357 scenario end to end
-- (a whole tool in a slot named hero) rather than a sanitised version of it.
--
-- Floor arithmetic after this, checked at the source (prune_floor.go:89 — ratio() returns 1
-- when Stored <= 0; :128 — Below is `ratio < floor`, so exactly-at passes):
--   cohort `sections`         = 1 confirmed / 0 stored  -> 1.00  passes
--   cohort `planned sections` = 1 confirmed / 1 planned -> 1.00  passes
--
-- ROLLBACK: OPERATION_2026-08-25_correct_cv1_tool_page_plans_ROLLBACK.sql (exact inverse of
-- the plan edit; it does NOT delete rows this run's recreation may have written by then).

BEGIN;

DO $$
DECLARE
    v_site   uuid := '8c3e9118-2455-4f0d-b01a-5dcde13dcf99';
    v_batch  uuid := gen_random_uuid();
    n        int;
    r        record;
BEGIN
    -- Refuse if the ground has moved since this file was written. The build cascade for
    -- this site was still running when it was authored (needs_briefing), and the site
    -- planner also writes pages.sections -- so "the plan is what I think it is" is exactly
    -- the assumption that must be checked rather than trusted.
    SELECT count(*) INTO n FROM pages
     WHERE site_id = v_site AND name IN ('index','tool-example');
    IF n <> 2 THEN
        RAISE EXCEPTION 'expected exactly 2 cv1 pages named index/tool-example, found %', n;
    END IF;

    SELECT count(*) INTO n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site;
    IF n <> 0 THEN
        RAISE EXCEPTION 'cv1 now holds % page_components row(s). This operation was authorised '
                        'on the basis that both pages are EMPTY and there is nothing to lose; '
                        'something has built them since. Re-read before proceeding.', n;
    END IF;

    -- The correction. planned[0] is preserved on both.
    UPDATE pages SET sections = '["hero"]'::jsonb, updated_at = now()
     WHERE site_id = v_site AND name = 'index';
    UPDATE pages SET sections = '["generic-text-block"]'::jsonb, updated_at = now()
     WHERE site_id = v_site AND name = 'tool-example';

    -- Re-queue the recreation, reusing each page's ORIGINAL spec so mode/interactive_features
    -- travel unchanged -- the point is to re-run the same route, not a different one.
    INSERT INTO site_work_items (
        site_id, source, pipeline, item_type, severity, summary,
        spec, page_id, priority, handler_agent, status, created_by, item_key, batch_id
    )
    SELECT v_site, 'adoption', 'build', 'needs_tool_recreation', 'medium',
           'Re-run tool recreation for ' || p.name || ' (357 phase-2 first firing; plan corrected to one section)',
           old.spec, p.id, 5, 'tool-recreation-handler', 'triaged', 'bugs_open/357',
           'needs_page:' || p.name, v_batch
      FROM pages p
      JOIN LATERAL (
            SELECT wi.spec FROM site_work_items wi
             WHERE wi.page_id = p.id AND wi.item_type = 'needs_tool_recreation'
             ORDER BY wi.created_at DESC LIMIT 1
      ) old ON true
     WHERE p.site_id = v_site AND p.name IN ('index','tool-example');

    GET DIAGNOSTICS n = ROW_COUNT;
    IF n <> 2 THEN
        RAISE EXCEPTION 'expected to queue 2 recreations, queued % -- an open row with the same '
                        'item_key would be refused by idx_swi_dedup', n;
    END IF;

    -- Prove the plan now matches what the route can produce.
    FOR r IN
        SELECT p.name, jsonb_array_length(p.sections) AS planned
          FROM pages p WHERE p.site_id = v_site AND p.name IN ('index','tool-example')
    LOOP
        IF r.planned <> 1 THEN
            RAISE EXCEPTION 'page % still plans % sections; the floor would refuse again', r.name, r.planned;
        END IF;
        RAISE NOTICE 'cv1.co.uk/%: plan corrected to 1 section, recreation re-queued', r.name;
    END LOOP;
END $$;

COMMIT;
