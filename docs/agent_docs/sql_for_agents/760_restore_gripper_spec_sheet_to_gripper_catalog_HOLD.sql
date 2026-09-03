-- 760 (_HOLD — DO NOT APPLY): restore `gripper-spec-sheet` to the CURRENT plan's
--     section rows for robot-hands.com/gripper-catalog
--
-- ═══ WHY THIS IS _HOLD AND NOT A MIGRATION YOU MAY RUN ═══════════════════════
--
-- It is held on an OWNER RULING, not on incompleteness. Three questions must be
-- answered first, and the third would make this file moot:
--
--   q1  RFC_064 §7 q2 — may a non-planner write withdraw a page's
--       `built_from_plan_version`? WITHOUT that, this migration is a NO-OP in
--       effect: the page is stamped at the current plan and `build_status =
--       'deployed'`, so reconcile_site_plan_action.go's decideEmit returns
--       `skip_built` BEFORE it compares any section list (:612-614). The
--       corrected plan would sit in the database and never reach a visitor.
--       ⚠ This is the sharp difference from migration 750, which needed no
--       rebuild because it aligned the plan to a page that was ALREADY right.
--
--   q2  Does the build path reach an archived page at all? `pages.status =
--       'archived'` here. UNKNOWN to the author; not assumed either way.
--
--   q3  Should this page be serving at all? It is `archived` and returns
--       HTTP 200 — one of NINE open `archived_page_still_serving` items
--       (bugs_closed/359's detector, migration 648), filed 2026-08-26 and
--       2026-09-02, ALL still `detected`, none triaged. **A "retire it
--       properly" answer to q3 moots this whole file.** Answering q1 alone
--       implicitly un-retires the page, which is why the three go up together.
--
-- ═══ WHAT IS BEING RESTORED, AND ON WHOSE AUTHORITY ══════════════════════════
--
-- `gripper-spec-sheet` was put on this page on 2026-07-24 by
-- `docs/agent_docs/docs024_key_docs_latest/robot_hands/SQL_2026-07-24_r9_gripper_catalog_real_grid.sql`
-- ("R9"), at position 3, deliberately INSTEAD of `product-grid` — that migration's
-- own comment records the reason: product-grid's e-commerce fields (price, rating,
-- badge, image) would be empty for grippers and "invite fabrication", where spec
-- cards are "the honest fit". The owning lane's HANDOFF frames it as an owner call
-- ("owner chose extend-don't-soften again"), and that lane confirmed on
-- 2026-09-03, after re-reading its own SUMMARY/README files, that there is NO
-- later reversal.
--
-- R9 wrote `pages.sections` (tier 3) and not `site_plan_sections` (tier 1). The
-- next page BUILD synced tier 1 down over the cache and the section was destroyed.
-- That is `bugs_open/469`'s completed loss, and it is the SECOND time this exact
-- component has been lost this way on this site — migration `154` (2026-07-15) was
-- written to rescue it from migration `153` making the identical mistake.
--
-- ═══ WHY MIGRATION 750'S TEMPLATE DOES NOT TRANSFER ══════════════════════════
--
--   750 (boxingonline)                     | this file
--   ---------------------------------------|--------------------------------------
--   in-place RENAME at a fixed `ordering`  | INSERT, with two rows shifting down
--   plan aligned to an ALREADY-CORRECT page| page AND plan both wrong
--   exactly 1 site_plans row (pre-checked) | FIVE site_plans rows
--   artefact byte-for-byte unchanged       | the page must REBUILD to render
--
-- On the RENUMBERING risk, enumerated for THIS page rather than carried across
-- from 750's general warning (`ordering` is a positional join key for four
-- consumers) — [MEASURED 2026-09-03]:
--   * `assigned_fact_ids` — NULL on all four rows. `'[]'` vs NULL is a real
--     distinction (`'[]'` = "states no verified facts", NULL = unscoped); here
--     every row is NULL, so a shift re-points nothing.
--   * `subject`            — NULL on all four rows.
--   * `site_plan_imagery`  — this page's ONLY row is scope='page',
--     scope_ref='gripper-catalog'. NOT the `<page>:<ordinal>` section form, so a
--     shift cannot silently re-point a section figure.
--   * `page_components.position` — rewritten by the rebuild that this correction
--     requires anyway; it is not a store this file writes.
-- So the shift is low-risk ON THIS PAGE. The general warning stands.
--
-- On per-plan immutability (reconcile_site_plan_action.go:596-601, which
-- decideEmit's restamp correctness rests on): the four superseded plans have
-- **0 pages stamped to them**; all 51 stamped pages point at the current plan
-- 7a40a0f9. This file touches ONLY the current plan, and rows are keyed
-- (plan_id, page_name), so no other page's verdict moves.
--
-- Tier 2 (`site_specs.site_plan`) EXISTS on this site and is current (14 pages),
-- but its `gripper-catalog` entry has NO `sections` key — so it does not serve
-- (loadAspectSections requires a non-nil array; the loader's type assertion
-- fails on null). This file deliberately does NOT create one: giving tier 2 a
-- say here would hand it permanent precedence over tier 3 for every page.
--
-- ═══ WHAT THIS FILE DELIBERATELY DOES NOT DO ════════════════════════════════
--   * it does NOT touch `pages.sections`   — the build's sync-down writes it,
--                                            and writing the cache is the very
--                                            mistake that caused this bug twice
--   * it does NOT touch `page_components`  — the rebuild composes them
--   * it does NOT withdraw the build stamp — that is q1, the owner's to rule
--   * it does NOT create a tier-2 aspect entry
--
-- ═══ APPLY ORDER, IF AND WHEN RULED ═════════════════════════════════════════
--   1. owner rules q1 (and q2/q3 do not moot it)
--   2. rename this file without `_HOLD`, ADD the stamp-withdrawal statements the
--      ruling licenses, and re-run the pre-checks — they re-derive live state
--   3. apply, then drive a rebuild of the page and verify AT THE ARTEFACT:
--        ./scripts/probe-page-url.sh robot-hands.com gripper-catalog
--      and confirm a `gripper-spec-sheet` row exists in page_components.
--      ⚠ The plan being right is NOT the verification. The page is.
--
-- Verify the plan side by hand after applying (the loader's own query):
--   SELECT sps.component_name FROM site_plan_sections sps
--     JOIN site_plans sp ON sp.id = sps.plan_id
--    WHERE sp.site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92'
--      AND sp.is_current = true AND sps.page_name = 'gripper-catalog'
--    ORDER BY sps.ordering;
--   -- expect: hero, generic-text-block, gripper-spec-sheet, info-card-grid, call-to-action

BEGIN;

-- ── Pre-checks. Every one re-SELECTs the LIVE row, and none inspects a variable
-- this block itself set. DO/RAISE, never bare SELECTs: ON_ERROR_STOP ignores a
-- non-empty result, so a verify block of SELECTs cannot abort the COMMIT.
DO $pre$
DECLARE
    v_site   uuid := '00ff3af5-dad8-4770-9f70-3edc267a3c92';
    v_page   text := 'gripper-catalog';
    v_plan   uuid;
    v_n      int;
    v_names  jsonb;
    v_orders int[];
BEGIN
    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current = true;
    IF v_plan IS NULL THEN
        RAISE EXCEPTION '760 ABORT: no is_current plan for robot-hands.com';
    END IF;

    -- The superseded plans must still hold nothing. This is the immutability
    -- premise, re-derived rather than quoted: 750 could assert "exactly one
    -- plan"; here there are five and the safety comes from where the STAMPS point.
    SELECT count(*) INTO v_n FROM pages p
     WHERE p.site_id = v_site AND p.built_from_plan_version IS NOT NULL
       AND p.built_from_plan_version <> v_plan;
    IF v_n <> 0 THEN
        RAISE EXCEPTION '760 ABORT: % page(s) are stamped to a SUPERSEDED plan. Mutating the current plan is still safe for them, but the premise recorded in this header (all stamps point at the current plan) no longer holds; re-derive before applying.', v_n;
    END IF;

    -- tier 1 must be EXACTLY the four-section list, in order, at 0..3.
    SELECT jsonb_agg(component_name ORDER BY ordering), array_agg(ordering ORDER BY ordering)
      INTO v_names, v_orders
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = v_page;
    IF v_names IS DISTINCT FROM '["hero","generic-text-block","info-card-grid","call-to-action"]'::jsonb
       OR v_orders IS DISTINCT FROM ARRAY[0,1,2,3] THEN
        RAISE EXCEPTION '760 ABORT: tier-1 rows are % at orderings %, expected [hero,generic-text-block,info-card-grid,call-to-action] at [0,1,2,3]. Another session has changed the plan, or the repair has already run; re-derive.', v_names, v_orders;
    END IF;

    -- The positional scoping must be the shape the header enumerated. Refuse
    -- rather than shift rows whose scoping this file has not reasoned about.
    SELECT count(*) INTO v_n FROM site_plan_sections
     WHERE plan_id = v_plan AND page_name = v_page
       AND (assigned_fact_ids IS NOT NULL OR subject IS NOT NULL
            OR component_version_id IS NOT NULL OR palette_id IS NOT NULL
            OR layout_id IS NOT NULL OR typography_set_id IS NOT NULL);
    IF v_n <> 0 THEN
        RAISE EXCEPTION '760 ABORT: % of the 4 plan rows now carry a non-NULL assigned_fact_ids/subject/component_version_id/palette_id/layout_id/typography_set_id. The renumbering below would silently re-point them; inspect by hand.', v_n;
    END IF;

    -- No SECTION-scoped imagery on this page: `<page>:<ordinal>` is a positional
    -- join key and a shift would re-point every figure.
    SELECT count(*) INTO v_n FROM site_plan_imagery
     WHERE plan_id = v_plan AND scope_ref LIKE (v_page || ':%');
    IF v_n <> 0 THEN
        RAISE EXCEPTION '760 ABORT: % section-scoped site_plan_imagery row(s) (scope_ref LIKE ''%%:<ordinal>'') exist for this page. Shifting orderings would re-point them.', v_n;
    END IF;

    -- The component being restored must still exist and be usable.
    IF NOT EXISTS (SELECT 1 FROM content_components WHERE function = 'gripper-spec-sheet') THEN
        RAISE EXCEPTION '760 ABORT: no content_components row with function=gripper-spec-sheet. The component this restores no longer exists.';
    END IF;

    -- The cache must still be the four-section list. If it is not, the page has
    -- moved and this file's target is wrong.
    IF NOT EXISTS (SELECT 1 FROM pages WHERE site_id = v_site AND name = v_page
                     AND sections = '["hero","generic-text-block","info-card-grid","call-to-action"]'::jsonb) THEN
        RAISE EXCEPTION '760 ABORT: pages.sections is no longer the four-section list. Re-derive the target.';
    END IF;

    -- Zero locked rows => MergeLockedPageSlots is the identity, so the raw
    -- comparisons in the post-check mean what check_section_source_drift means.
    SELECT count(*) INTO v_n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = v_page
       AND COALESCE(pc.build_status,'') <> 'removed'
       AND NOT (pc.locked_at IS NULL
                OR (pc.lock_type = 'timed' AND pc.lock_expires_at IS NOT NULL AND pc.lock_expires_at < NOW()));
    IF v_n <> 0 THEN
        RAISE EXCEPTION '760 ABORT: % locked row(s) on the page. The post-check compares raw lists and would no longer match the drift check''s merged comparison.', v_n;
    END IF;
END
$pre$;

-- ── The write. Three statements, scoped to the CURRENT plan and this page only.
--
-- SHIFT VIA AN OFFSET, NOT `ordering + 1` IN PLACE. idx_site_plan_sections_key
-- is a NON-deferrable UNIQUE index on (plan_id, page_name, ordering), and a
-- single statement moving 2->3 while 3 still exists can raise a spurious
-- violation depending on row order. Parking the rows at +1000 first makes the
-- collision structurally impossible rather than order-dependent.

UPDATE site_plan_sections
   SET ordering = ordering + 1000
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true)
   AND page_name = 'gripper-catalog'
   AND ordering >= 2;

INSERT INTO site_plan_sections (plan_id, page_name, ordering, component_name)
VALUES ((SELECT id FROM site_plans WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true),
        'gripper-catalog', 2, 'gripper-spec-sheet');

UPDATE site_plan_sections
   SET ordering = ordering - 999
 WHERE plan_id = (SELECT id FROM site_plans WHERE site_id = '00ff3af5-dad8-4770-9f70-3edc267a3c92' AND is_current = true)
   AND page_name = 'gripper-catalog'
   AND ordering >= 1000;

-- ── Post-write verify. Re-reads live rows inside the transaction.
DO $post$
DECLARE
    v_site   uuid := '00ff3af5-dad8-4770-9f70-3edc267a3c92';
    v_page   text := 'gripper-catalog';
    v_plan   uuid;
    v_names  jsonb;
    v_orders int[];
    v_n      int;
BEGIN
    SELECT id INTO v_plan FROM site_plans WHERE site_id = v_site AND is_current = true;

    SELECT jsonb_agg(component_name ORDER BY ordering), array_agg(ordering ORDER BY ordering)
      INTO v_names, v_orders
      FROM site_plan_sections WHERE plan_id = v_plan AND page_name = v_page;
    IF v_names IS DISTINCT FROM '["hero","generic-text-block","gripper-spec-sheet","info-card-grid","call-to-action"]'::jsonb
       OR v_orders IS DISTINCT FROM ARRAY[0,1,2,3,4] THEN
        RAISE EXCEPTION '760 VERIFY FAILED: tier 1 reads % at orderings %, expected the 5-section list at [0,1,2,3,4].', v_names, v_orders;
    END IF;

    -- The two shifted rows must be the SAME ROWS, moved — not replacements.
    -- A delete-and-reinsert would re-choose assigned_fact_ids, which is the
    -- distinction migration 750's header exists to protect.
    IF NOT EXISTS (SELECT 1 FROM site_plan_sections
                    WHERE id = '8654754d-9e17-490f-80a3-4826ea628d1b'
                      AND component_name = 'info-card-grid' AND ordering = 3) THEN
        RAISE EXCEPTION '760 VERIFY FAILED: row 8654754d is not info-card-grid at ordering 3 — this was a replace, not a move.';
    END IF;
    IF NOT EXISTS (SELECT 1 FROM site_plan_sections
                    WHERE id = '0f34d9f1-68a5-4fa2-8d88-4e8223f2650c'
                      AND component_name = 'call-to-action' AND ordering = 4) THEN
        RAISE EXCEPTION '760 VERIFY FAILED: row 0f34d9f1 is not call-to-action at ordering 4 — this was a replace, not a move.';
    END IF;

    -- No row parked at the offset survived.
    SELECT count(*) INTO v_n FROM site_plan_sections
     WHERE plan_id = v_plan AND page_name = v_page AND ordering >= 1000;
    IF v_n <> 0 THEN
        RAISE EXCEPTION '760 VERIFY FAILED: % row(s) still parked at the +1000 offset.', v_n;
    END IF;

    -- Scoping columns must still be untouched on every surviving row.
    SELECT count(*) INTO v_n FROM site_plan_sections
     WHERE plan_id = v_plan AND page_name = v_page
       AND (assigned_fact_ids IS NOT NULL OR subject IS NOT NULL);
    IF v_n <> 0 THEN
        RAISE EXCEPTION '760 VERIFY FAILED: % row(s) acquired a non-NULL assigned_fact_ids or subject.', v_n;
    END IF;

    -- This file must NOT have touched the cache or the artefact. Both are the
    -- rebuild's job, and writing the cache here is the mistake that caused the bug.
    IF NOT EXISTS (SELECT 1 FROM pages WHERE site_id = v_site AND name = v_page
                     AND sections = '["hero","generic-text-block","info-card-grid","call-to-action"]'::jsonb) THEN
        RAISE EXCEPTION '760 VERIFY FAILED: pages.sections changed. This migration must not write the cache.';
    END IF;
    SELECT count(*) INTO v_n FROM page_components pc JOIN pages p ON p.id = pc.page_id
     WHERE p.site_id = v_site AND p.name = v_page AND COALESCE(pc.build_status,'') <> 'removed';
    IF v_n <> 4 THEN
        RAISE EXCEPTION '760 VERIFY FAILED: live page_components is now %, expected 4 (this migration must not alter the artefact).', v_n;
    END IF;

    -- AND THE HONEST POSTCONDITION: the plan and the cache now DISAGREE on
    -- purpose. That is a section_source_drift finding by construction, and it is
    -- the correct intermediate state — the page has not rebuilt yet. It becomes
    -- a DEFECT if no rebuild follows.
    RAISE NOTICE '760 APPLIED: tier 1 now names 5 sections, pages.sections still names 4. This is DELIBERATE and TRANSIENT. The page MUST be rebuilt or the correction never reaches a visitor — and until it is, check_section_source_drift will (correctly) flag this page.';
END
$post$;

COMMIT;
