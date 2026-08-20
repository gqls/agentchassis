-- 498 ROLLBACK — restore `497`'s `literal_markdown` owner string.
--
-- Reverses 498 on the same anchor, driven from the same two variables with the directions
-- swapped. Guarded on the POST-498 md5, and asserts the md5 returns exactly to 497's output.
--
-- ⚠ READ THIS BEFORE REVERTING. What 498 removed was not an opinion, it was three claims that
-- had already been measured false:
--   * "all 7 rows held on page-build-handler" — the 7 terminated `wont_fix` 2026-08-20 07:20-07:24Z;
--   * "a healthy pair" / floor-held — the pair went 8.1% -> 44% overnight;
--   * "drag a healthy pair toward its floor" — `301` makes that mechanically impossible, since a
--     refusal now terminates `wont_fix`, which is excluded from both sides of the promoter rule.
-- Reverting restores all three into a one-shot annotation that a human reads at escalation time.
-- If you need 497's wording back for provenance, read it here or in `git show`; you almost
-- certainly do not need it LIVE. And as with 497, reverting cannot rewrite the string on a row
-- that has already escalated — it changes only what future ticks write.

BEGIN;

DO $$
DECLARE
    q    text;
    back text;
    live_md5 text;
    n    int;
    old2 text;   -- 497's string
    new2 text;   -- 498's string
BEGIN
    old2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. NB THE reason FIELD UNDERSTATES THIS PAIR: its floor text is true of the PAIR and is not why THESE rows are stuck. [MEASURED 2026-08-19] all 7 rows held on page-build-handler sit on rebuild_policy=owned pages, and the generic repair refuses an owned page BY DESIGN. The newer literal_markdown to page-rerender route stands at 8 complete on generic pages and 1 failed on the single owned page it tried, so re-pointing these 7 would produce 7 more failures, drag a healthy pair toward its floor, and repair nothing. The live candidate is the one already named in the what_to_do ELSE branch below: 295 fix candidate 3, route owned-page content findings to section_edit.'$a$;

    new2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. FIRST SPLIT THE HELD ROWS BY pages.rebuild_policy: the reason field beside this one describes the PAIR, and an owned-page row is not waiting for a route - it is REFUSED BY DESIGN by the generic page-build repair, so it will never be repaired by being re-pointed at one. WORKED INSTANCE, 2026-08-20 07:20-07:24Z, written up in 277 and 333: this pair rose above the promoter floor, the promoter released the 7 owned-page rows it had been holding, and every one was refused by OWNED_PAGE_GUARD and terminated wont_fix. Note what that costs you as a reader: wont_fix is excluded from BOTH sides of the promoter rule (301), and idx_swi_dedup excludes it too so the detector re-files the same finding (333) - the loop therefore leaves no trace in this pair record, and a healthy-looking ratio here is not evidence that owned pages are being repaired. CANDIDATE ROUTE, measured 2026-08-20: section_edit to section-editor, 36 complete / 1 failed of 39 on rebuild_policy=owned pages lifetime incl. archive (295 fix candidate 3). Target the SECTION component that carries the prose, not a component_level=tool fork. section_edit REWRITES an existing component and cannot ADD one, which suits literal markdown. UNVERIFIED: whether a producer can file a section_edit item for this finding shape - that is 277 question.'$a$;

    SELECT pre_query, md5(pre_query) INTO q, live_md5
      FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q IS NULL THEN
        RAISE EXCEPTION '498 ROLLBACK: no held-pair-canary-escalation row';
    END IF;

    IF live_md5 <> 'ddd0c894845fa77edcd2d9f213a4d890' THEN
        RAISE EXCEPTION
          '498 ROLLBACK: ABORTING — the live pre_query is not what 498 produced (expected md5 ddd0c894845fa77edcd2d9f213a4d890 len 13746, found % len %). Either 498 was never applied, or another lane has edited the column since. A blind revert here would silently undo whatever landed after 498.',
          live_md5, length(q);
    END IF;

    n := (length(q) - length(replace(q, new2, ''))) / length(new2);
    IF n <> 1 THEN RAISE EXCEPTION '498 ROLLBACK: the 498 string occurs % times, expected 1', n; END IF;

    UPDATE scheduled_tasks
       SET pre_query = replace(pre_query, new2, old2)
     WHERE name = 'held-pair-canary-escalation';

    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF md5(q) <> '406bd7571e60981fb062a2e2b3fac515' THEN
        RAISE EXCEPTION
          '498 ROLLBACK VERIFY FAILED: expected 497 output md5 406bd7571e60981fb062a2e2b3fac515 len 13106, got % len %. Do not commit.',
          md5(q), length(q);
    END IF;

    back := replace(q, old2, new2);
    IF md5(back) <> 'ddd0c894845fa77edcd2d9f213a4d890' THEN
        RAISE EXCEPTION '498 ROLLBACK VERIFY: reverted text does not re-apply to 498 cleanly (got % len %)', md5(back), length(back);
    END IF;

    RAISE NOTICE '498 ROLLBACK OK: 497 output restored exactly (md5 % len %), and it re-applies to 498 cleanly.', md5(q), length(q);
END $$;

COMMIT;
