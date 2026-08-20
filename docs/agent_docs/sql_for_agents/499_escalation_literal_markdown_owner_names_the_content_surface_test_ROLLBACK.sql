-- 499 ROLLBACK — restore `498`'s `literal_markdown` owner string.
--
-- Reverses 499 on the same anchor, from the same two variables, directions swapped.
-- Guarded on the POST-499 md5; asserts the md5 returns exactly to 498's output.
--
-- ⚠ READ BEFORE REVERTING. What 499 removed was a specific, actionable instruction that is
-- WRONG for the population this entry is attached to: *"Target the SECTION component that
-- carries the prose."* On the pages that produced this residual, that component's
-- `content_data` holds only metadata (`{schema, sha256, source, qa_tier, generator}`) and
-- its template's only field is `{{.body}}`, which is not a key — rendering it produces ZERO
-- visible characters (measured with production's own engine and `missingkey=zero`). The
-- findings themselves are `source: rendered_html`, so no content_data-based route reaches
-- them at all. Restoring that sentence puts a wrong target back into a one-shot annotation
-- a human reads at escalation time.
--
-- 499 did NOT retract the 36/1 `section_edit` record — that is correct and is kept. It
-- replaced a TARGET with a TEST. If you are reverting because you disagree with the test,
-- write a forward migration saying why; a revert restores the wrong target as a side effect.

BEGIN;

DO $$
DECLARE
    q    text;
    back text;
    live_md5 text;
    n    int;
    old2 text;   -- 498's string
    new2 text;   -- 499's string
BEGIN
    old2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. FIRST SPLIT THE HELD ROWS BY pages.rebuild_policy: the reason field beside this one describes the PAIR, and an owned-page row is not waiting for a route - it is REFUSED BY DESIGN by the generic page-build repair, so it will never be repaired by being re-pointed at one. WORKED INSTANCE, 2026-08-20 07:20-07:24Z, written up in 277 and 333: this pair rose above the promoter floor, the promoter released the 7 owned-page rows it had been holding, and every one was refused by OWNED_PAGE_GUARD and terminated wont_fix. Note what that costs you as a reader: wont_fix is excluded from BOTH sides of the promoter rule (301), and idx_swi_dedup excludes it too so the detector re-files the same finding (333) - the loop therefore leaves no trace in this pair record, and a healthy-looking ratio here is not evidence that owned pages are being repaired. CANDIDATE ROUTE, measured 2026-08-20: section_edit to section-editor, 36 complete / 1 failed of 39 on rebuild_policy=owned pages lifetime incl. archive (295 fix candidate 3). Target the SECTION component that carries the prose, not a component_level=tool fork. section_edit REWRITES an existing component and cannot ADD one, which suits literal markdown. UNVERIFIED: whether a producer can file a section_edit item for this finding shape - that is 277 question.'$a$;

    new2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. BEFORE YOU ROUTE ANYTHING, RUN THIS TEST ON YOUR OWN ROWS - it is what three corrections to this entry in one day were about. (1) READ THE FINDING SOURCE. spec findings carry source = content_data OR rendered_html; the detector scans both. (2) ASK WHETHER content_data CAN REPRODUCE rendered_html for the target component - are the template fields actually keys in content_data? If the finding is in rendered_html and the content lives only there, NO content_data-based route can repair it, however good that route record looks elsewhere. WORKED INSTANCE, 2026-08-20: all 7 owned-page rows held here were source=rendered_html, pattern=code_span, slot=ported-page - backticked code tokens in ported prose. That component content_data holds only metadata and its template only field is body, which is not a key; rendering it produces ZERO visible characters (measured against production own engine, missingkey=zero). So BOTH routes this lane proposed were inapplicable by construction: 473 rerender-from-content_data, and section_edit with strip_literal_markdown, which strips a map holding no prose. Not dangerous, just useless - enforceSingleSlotFloors measures VISIBLE text and refuses at zero, leaving the component standing. NOT A GENERAL VERDICT: section_edit to section-editor really is 36 complete / 1 failed of 39 on owned pages, and is the right route where content_data CAN fill the template. The predicate is not ownership, it is whether the content is reachable from content_data at all.'$a$;

    SELECT pre_query, md5(pre_query) INTO q, live_md5
      FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q IS NULL THEN
        RAISE EXCEPTION '499 ROLLBACK: no held-pair-canary-escalation row';
    END IF;

    IF live_md5 <> '0d72b423ef41447c98e64d7d894d6f56' THEN
        RAISE EXCEPTION
          '499 ROLLBACK: ABORTING — the live pre_query is not what 499 produced (expected md5 0d72b423ef41447c98e64d7d894d6f56 len 13938, found % len %). Either 499 was never applied or another lane has edited the column since.',
          live_md5, length(q);
    END IF;

    n := (length(q) - length(replace(q, new2, ''))) / length(new2);
    IF n <> 1 THEN RAISE EXCEPTION '499 ROLLBACK: the 499 string occurs % times, expected 1', n; END IF;

    UPDATE scheduled_tasks
       SET pre_query = replace(pre_query, new2, old2)
     WHERE name = 'held-pair-canary-escalation';

    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF md5(q) <> 'ddd0c894845fa77edcd2d9f213a4d890' THEN
        RAISE EXCEPTION
          '499 ROLLBACK VERIFY FAILED: expected 498 output md5 ddd0c894845fa77edcd2d9f213a4d890 len 13746, got % len %. Do not commit.',
          md5(q), length(q);
    END IF;

    back := replace(q, old2, new2);
    IF md5(back) <> '0d72b423ef41447c98e64d7d894d6f56' THEN
        RAISE EXCEPTION '499 ROLLBACK VERIFY: reverted text does not re-apply to 499 cleanly (got % len %)', md5(back), length(back);
    END IF;

    RAISE NOTICE '499 ROLLBACK OK: 498 output restored exactly (md5 % len %), and it re-applies to 499 cleanly.', md5(q), length(q);
END $$;

COMMIT;
