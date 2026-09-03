-- 750: make bugs_open/427's fight-calendar fix DURABLE by correcting the
--      AUTHORITATIVE site_plan_sections rows to match what the page already is
--
-- Context (2026-09-03). bugs_open/427's migrations 719/727/728 swapped
-- generic-text-block -> event-list and dropped `advertising` on
-- boxingonline.com/tool-fight-calendar. They wrote ONLY pages.sections.
--
-- pages.sections is TIER 3 (a materialised cache). The page build reads
-- load_page_sections_from_spec_action.go, whose priority order is:
--   1. site_plan_sections for the CURRENT plan, ORDER BY ordering   (:142-148)
--   2. site_specs.site_plan aspect
--   3. pages.sections
-- ...and it SYNCS THE WINNER DOWN over pages.sections (:558-570). So tier 1
-- still names [hero-tool, generic-text-block, advertising] and the next page
-- BUILD would resurrect them and delete event-list.
--
-- ⚠ 427 §19.2 named `sync_pages` as the reverting mechanism. That is WRONG,
-- though its conclusion (the migrations are transient) is RIGHT. sync_pages is
-- preceded by reconcilePlanWithRealised (v3_site_actions.go:7701-7724), which
-- SNAPS a deployed/needs_rebuild page's realised pages.sections back ONTO the
-- plan proposal -- so a re-plan launders the cache forward and is SAFE. The
-- danger is any page BUILD, which needs no re-plan at all. A reader of §19.2
-- would guard the wrong path.
--
-- This is the same trap migration 154 fixed on 2026-07-15 after 153 made
-- exactly this mistake. It has since COMPLETED twice more: robot-hands.com/
-- gripper-catalog has lost `gripper-spec-sheet` (the very component 154 was
-- written to rescue) and idea.uk/guides-index lost `guide-list`. In both, all
-- three stores now agree WITHOUT the corrected component.
--
-- SHAPE: an in-place RENAME, deliberately NOT 154's delete-renumber-insert.
-- assigned_fact_ids and subject are read POSITIONALLY by `ordering`
-- (load_page_sections_from_spec_action.go:164-186). These rows carry
-- assigned_fact_ids = '[]', which is NOT the same as NULL: '[]' means "this
-- section deliberately states no verified facts" (plan_sections_action.go
-- scopeItem returns early only on facts == nil), NULL means unscoped. A
-- delete+insert re-chooses that value; a rename cannot. Orderings stay
-- contiguous (0,1), so idx_site_plan_sections_key (plan_id, page_name,
-- ordering) is never in play.
--
-- SAFETY vs. per-plan immutability (reconcile_site_plan_action.go:596-601,
-- which decideEmit's restamp correctness rests on):
--   * boxingonline.com has EXACTLY ONE site_plans row. There is no superseded
--     plan whose rows could be falsified for a cohort.
--   * That plan is also the page's built_from_plan_version, so decideEmit
--     returns "skip_built" BEFORE it compares any section list.
--   * site_plan_sections is keyed per (plan_id, page_name), so no other page
--     on the site changes verdict.
--   * After this migration built_from_plan_version becomes a TRUE statement
--     about what was served, where today it is false.
-- The general question -- may a non-planner action mutate the current plan --
-- is architecture-scope and is going to the council/owner as an RFC. This file
-- is the single-page precedent 154 already set, not that decision.
--
-- WHAT THIS DOES NOT TOUCH, and why:
--   * pages.sections            -- already correct: ["hero-tool","event-list"]
--   * page_components           -- already correct: hero-tool@1, event-list@2
--   * site_specs.site_plan      -- this site has NO such aspect (0 rows).
--                                  153 wrote one because robot-hands HAD one;
--                                  creating one here would hand tier 2 a
--                                  permanent say over tier 3 on every page.
--   * the served artefact       -- expected byte-for-byte UNCHANGED. This
--                                  removes a latent revert; it must not alter
--                                  output. Any change to page_components
--                                  bytes or pages.updated_at is a FINDING.
--
-- RESIDUAL, stated not buried: this defends against the tier-1 loader. It does
-- NOT make the page immune to a genuine re-plan minting a NEW site_plans row,
-- which no per-plan migration can reach. That is the framework fix's half.
--
-- Verify by hand after applying (the loader's own query):
--   SELECT sps.component_name FROM site_plan_sections sps
--     JOIN site_plans sp ON sp.id = sps.plan_id
--    WHERE sp.site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
--      AND sp.is_current = true AND sps.page_name = 'tool-fight-calendar'
--    ORDER BY sps.ordering;
--   -- expect: hero-tool, event-list

BEGIN;

-- ── Pre-checks. Every one re-SELECTs the LIVE row. A bare SELECT cannot abort
-- a transaction (ON_ERROR_STOP ignores a non-empty result), so these are
-- DO/RAISE. And none of them inspects a variable this block itself set --
-- migration 747 was bitten by exactly that form today.
DO $pre$
DECLARE
    v_site      uuid := 'd2aa5206-73bc-4707-a69c-2702c1eb9152';
    v_page      text := 'tool-fight-calendar';
    v_plan      uuid;
    v_n         int;
    v_names     jsonb;
    v_orders    int[];
BEGIN
    SELECT count(*) INTO v_n FROM site_plans WHERE site_id = v_site;
    IF v_n <> 1 THEN
        RAISE EXCEPTION '750 ABORT: expected exactly 1 site_plans row for boxingonline.com, found %. The no-superseded-plan premise no longer holds; re-derive before applying.', v_n;
    END IF;

    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current = true;
    IF v_plan IS NULL THEN
        RAISE EXCEPTION '750 ABORT: no is_current plan for boxingonline.com';
    END IF;

    IF NOT EXISTS (SELECT 1 FROM pages
                    WHERE site_id = v_site AND name = v_page
                      AND built_from_plan_version = v_plan) THEN
        RAISE EXCEPTION '750 ABORT: page built_from_plan_version is not the current plan. decideEmit would no longer return skip_built and this edit is no longer inert for the reconciler.';
    END IF;

    -- tier 1 must be EXACTLY the stale triple, in order.
    SELECT jsonb_agg(component_name ORDER BY ordering), array_agg(ordering ORDER BY ordering)
      INTO v_names, v_orders
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = v_page;
    IF v_names IS DISTINCT FROM '["hero-tool","generic-text-block","advertising"]'::jsonb
       OR v_orders IS DISTINCT FROM ARRAY[0,1,2] THEN
        RAISE EXCEPTION '750 ABORT: tier-1 rows are % at orderings %, expected [hero-tool,generic-text-block,advertising] at [0,1,2]. Another session has changed the plan; re-derive.', v_names, v_orders;
    END IF;

    -- positional scoping must be the house shape; refuse rather than carry an
    -- unexpected value forward blind.
    SELECT count(*) INTO v_n FROM site_plan_sections
     WHERE plan_id = v_plan AND page_name = v_page
       AND (assigned_fact_ids IS DISTINCT FROM '[]'::jsonb
            OR subject IS NOT NULL
            OR component_version_id IS NOT NULL);
    IF v_n <> 0 THEN
        RAISE EXCEPTION '750 ABORT: % of the 3 plan rows carry a non-[] assigned_fact_ids, a subject, or a pinned component_version_id. The rename would silently re-scope or mis-pin; inspect them by hand.', v_n;
    END IF;

    -- tier 2 must be ABSENT: this migration deliberately does not create one.
    SELECT count(*) INTO v_n FROM site_specs WHERE site_id = v_site AND aspect = 'site_plan';
    IF v_n <> 0 THEN
        RAISE EXCEPTION '750 ABORT: a site_specs.site_plan aspect now exists (% rows). Tier 2 would outrank tier 3 and this migration no longer aligns all serving stores.', v_n;
    END IF;

    -- the cache and the live page must already BE the target.
    IF NOT EXISTS (SELECT 1 FROM pages WHERE site_id = v_site AND name = v_page
                     AND sections = '["hero-tool","event-list"]'::jsonb) THEN
        RAISE EXCEPTION '750 ABORT: pages.sections is not ["hero-tool","event-list"]. This migration aligns the plan TO the page; if the page has moved, the target is wrong.';
    END IF;

    SELECT count(*) INTO v_n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = v_page AND COALESCE(pc.build_status,'') <> 'removed';
    IF v_n <> 2 THEN
        RAISE EXCEPTION '750 ABORT: expected 2 live page_components, found %.', v_n;
    END IF;

    -- zero locked rows => MergeLockedPageSlots is the identity here, so the
    -- raw comparisons in the post-check are exactly what the drift check does.
    SELECT count(*) INTO v_n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = v_page
       AND COALESCE(pc.build_status,'') <> 'removed'
       AND NOT (pc.locked_at IS NULL
                OR (pc.lock_type = 'timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW()));
    IF v_n <> 0 THEN
        RAISE EXCEPTION '750 ABORT: % locked rows on the page. The post-check compares raw lists and would no longer match the drift check''s merged comparison.', v_n;
    END IF;
END
$pre$;

-- ── The write. Two statements, scoped to the CURRENT plan and this page only.
UPDATE site_plan_sections
   SET component_name = 'event-list'
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152' AND is_current = true)
   AND page_name = 'tool-fight-calendar'
   AND ordering = 1
   AND component_name = 'generic-text-block';

DELETE FROM site_plan_sections
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152' AND is_current = true)
   AND page_name = 'tool-fight-calendar'
   AND ordering = 2
   AND component_name = 'advertising';

-- ── Close this lane's own section_source_drift item, in the same transaction
-- as the alignment that resolves it, so neither can exist without the other.
UPDATE site_work_items
   SET status = 'complete',
       updated_at = NOW(),
       result = jsonb_build_object(
           'closed_by',  'migration 750',
           'closed_on',  '2026-09-03',
           'resolution', 'plan aligned to the live page; all serving stores now agree',
           'authority',  '["hero-tool","event-list"]'::jsonb,
           'cache',      '["hero-tool","event-list"]'::jsonb)
 WHERE site_id = 'd2aa5206-73bc-4707-a69c-2702c1eb9152'
   AND item_key = 'section_source_drift:tool-fight-calendar'
   AND status = 'needs_human_review';

-- ── Post-write verify. Re-reads live rows inside the transaction.
DO $post$
DECLARE
    v_site   uuid := 'd2aa5206-73bc-4707-a69c-2702c1eb9152';
    v_page   text := 'tool-fight-calendar';
    v_plan   uuid;
    v_names  jsonb;
    v_orders int[];
    v_cache  jsonb;
    v_n      int;
BEGIN
    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current = true;

    SELECT jsonb_agg(component_name ORDER BY ordering), array_agg(ordering ORDER BY ordering)
      INTO v_names, v_orders
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = v_page;
    IF v_names IS DISTINCT FROM '["hero-tool","event-list"]'::jsonb
       OR v_orders IS DISTINCT FROM ARRAY[0,1] THEN
        RAISE EXCEPTION '750 VERIFY FAILED: tier 1 reads % at orderings %, expected [hero-tool,event-list] at [0,1].', v_names, v_orders;
    END IF;

    -- the rename must have preserved the ROW, not replaced it.
    IF NOT EXISTS (SELECT 1 FROM site_plan_sections
                    WHERE id = 'd74518a8-03f9-4054-bd88-517aeae5f623'
                      AND component_name = 'event-list' AND ordering = 1) THEN
        RAISE EXCEPTION '750 VERIFY FAILED: row d74518a8 is not the event-list row at ordering 1 -- this was a replace, not a rename, and the positional scoping has been re-chosen.';
    END IF;

    SELECT count(*) INTO v_n FROM site_plan_sections
     WHERE plan_id = v_plan AND page_name = v_page
       AND (assigned_fact_ids IS DISTINCT FROM '[]'::jsonb OR subject IS NOT NULL);
    IF v_n <> 0 THEN
        RAISE EXCEPTION '750 VERIFY FAILED: % surviving rows no longer carry assigned_fact_ids=[] with a NULL subject.', v_n;
    END IF;

    -- the three-store agreement: this is check_section_source_drift's own
    -- orderedListsEqual over its two loaders, with the merge as identity.
    SELECT sections INTO v_cache FROM pages WHERE site_id = v_site AND name = v_page;
    IF v_cache IS DISTINCT FROM v_names THEN
        RAISE EXCEPTION '750 VERIFY FAILED: authority % <> cache %.', v_names, v_cache;
    END IF;

    -- the loader's own sync-down guard (:562), executed by hand: it must be a
    -- no-op, i.e. the next build would write nothing.
    IF EXISTS (SELECT 1 FROM pages WHERE site_id = v_site AND name = v_page
                 AND sections IS DISTINCT FROM v_names) THEN
        RAISE EXCEPTION '750 VERIFY FAILED: the loader sync-down would still fire.';
    END IF;

    -- the artefact must be untouched.
    SELECT count(*) INTO v_n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = v_page AND COALESCE(pc.build_status,'') <> 'removed';
    IF v_n <> 2 THEN
        RAISE EXCEPTION '750 VERIFY FAILED: live page_components is now %, expected 2 (this migration must not alter the artefact).', v_n;
    END IF;

    -- the drift item is closed.
    IF EXISTS (SELECT 1 FROM site_work_items
                WHERE site_id = v_site AND item_key = 'section_source_drift:tool-fight-calendar'
                  AND status = 'needs_human_review') THEN
        RAISE EXCEPTION '750 VERIFY FAILED: the section_source_drift item is still open.';
    END IF;
END
$post$;

COMMIT;
