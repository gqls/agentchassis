-- 497 ROLLBACK — restore the three stale `owners` entries and remove the rule comment.
--
-- Reverses 497 by the SAME four anchors, driven from the SAME eight variables, with
-- the directions swapped. Guarded on the POST-497 md5 so it refuses to run against a
-- text that is not what 497 produced, and it asserts the md5 returns exactly to 497's
-- pre-image.
--
-- ⚠ WHAT THIS CANNOT UNDO. 497's whole point is that `owners` is stamped into
-- `result.held_pair_escalation.owner` ONCE, at escalation time, and never revisited.
-- Any row that escalated while 497 was live keeps the CORRECTED string; any row that
-- escalated before it keeps the stale one. Reverting changes only what FUTURE ticks
-- write. That asymmetry is the reason 497 was worth applying before a tick, and it is
-- the reason this rollback is close to pointless after one — it is provided because a
-- migration without one is not reviewable, not because reverting would restore a prior
-- state. If you are reverting, prefer writing a forward migration that says why.

BEGIN;

DO $$
DECLARE
    em   text := chr(8212);   -- the em dash, spelled in ASCII so no anchor carries one
    q    text;
    back text;
    live_md5 text;
    n    int;
    old1 text; old2 text; old3 text; old4 text;
    new1 text; new2 text; new3 text; new4 text;
BEGIN
    -- the eight strings, identical to 497's. `old` = pre-497, `new` = post-497.
    old1 := $a$'bugs_open/201 lane $a$ || em || $a$ docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch'$a$;
    old2 := $a$'bugs_open/184 + bugs_open/201 lane $a$ || em || $a$ docs024_key_docs_latest/bugfix_201_page_content_writer_dispatch'$a$;
    old3 := $a$'(UNASSIGNED - claim this) check_page_component_status_drift.go added 2026-07-10, never touched since, no lane doc claims it'$a$;
    old4 := E'    owners (item_type, owner) AS (\n        VALUES\n';

    new1 := $a$'083_detected_findings_never_reach_a_handler (OPEN) + lane docs024_key_docs_latest/bugfix_277_required_fields_repair. The canary decision belongs to the 083 author. Context (279/284/290 contribution, 2026-08-18): the 0/6 is TYPE-specific, since this same handler completes 179 needs_page and 133 content_rewrite, and 3 of the 4 surviving failures are one day of a dead AI endpoint rather than a remit mismatch. Bug 201_page_content_writer_direct_dispatch_silently_or_fatally_no_ops, named here until 2026-08-19, is now CLOSED.'$a$;

    new2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. NB THE reason FIELD UNDERSTATES THIS PAIR: its floor text is true of the PAIR and is not why THESE rows are stuck. [MEASURED 2026-08-19] all 7 rows held on page-build-handler sit on rebuild_policy=owned pages, and the generic repair refuses an owned page BY DESIGN. The newer literal_markdown to page-rerender route stands at 8 complete on generic pages and 1 failed on the single owned page it tried, so re-pointing these 7 would produce 7 more failures, drag a healthy pair toward its floor, and repair nothing. The live candidate is the one already named in the what_to_do ELSE branch below: 295 fix candidate 3, route owned-page content findings to section_edit.'$a$;

    new3 := $a$'300_drift_findings_key_on_an_unstable_component_id_so_waiting_turns_them_into_failures (OPEN; fix live, council-approved) + lane docs024_key_docs_latest/bugfix_277_required_fields_repair. NB the CHECKER file check_page_component_status_drift.go is still untouched since 2026-07-10, so that half of the old UNASSIGNED note was true and remains true; what changed on 2026-08-19 is that the ITEM TYPE now has an owner, which it did not when 466 was written.'$a$;

    new4 := E'    owners (item_type, owner) AS (\n'
            '        -- NEVER write a bugs_open/ or bugs_closed/ PREFIX in this map (497).\n'
            '        -- That prefix is exactly what flips when the lane you are pointing at\n'
            '        -- SUCCEEDS, so a pointer written that way is built to rot at the moment\n'
            '        -- it starts to matter: all three entries below were dead or wrong by\n'
            '        -- 2026-08-19. Name a bug by NUMBER and SLUG and let the reader grep BOTH\n'
            '        -- directories; a bare number is ambiguous (184 alone names two unrelated\n'
            '        -- cases). Name the LANE DIRECTORY, which does not move when a bug closes.\n'
            '        -- This value is stamped into result.held_pair_escalation.owner ONCE, at\n'
            '        -- escalation time. A later migration does NOT retro-fix rows already\n'
            '        -- escalated, so a wrong entry here is a wrong instruction handed to a\n'
            '        -- person at the only moment they were going to read it.\n'
            '        VALUES\n';

    -- GUARD: refuse unless the live text is exactly what 497 produced.
    SELECT pre_query, md5(pre_query) INTO q, live_md5
      FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q IS NULL THEN
        RAISE EXCEPTION '497 ROLLBACK: no held-pair-canary-escalation row';
    END IF;

    IF live_md5 <> '406bd7571e60981fb062a2e2b3fac515' THEN
        RAISE EXCEPTION
          '497 ROLLBACK: ABORTING — the live pre_query is not what 497 produced (expected md5 406bd7571e60981fb062a2e2b3fac515 len 13106, found % len %). Either 497 was never applied, or another lane has edited the column since. Re-read the live column before reverting: a blind revert here would silently undo whatever landed after 497.',
          live_md5, length(q);
    END IF;

    n := (length(q) - length(replace(q, new4, ''))) / length(new4);
    IF n <> 1 THEN RAISE EXCEPTION '497 ROLLBACK: the 497 owners-header block occurs % times, expected 1', n; END IF;

    -- APPLY the reverse
    UPDATE scheduled_tasks
       SET pre_query = replace(replace(replace(replace(pre_query, new1, old1), new2, old2), new3, old3), new4, old4)
     WHERE name = 'held-pair-canary-escalation';

    -- VERIFY: we are back at 497's pre-image, byte for byte.
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF md5(q) <> '51547db8db7fc7caf6cca7c0b54c8ee3' THEN
        RAISE EXCEPTION
          '497 ROLLBACK VERIFY FAILED: expected 497 pre-image md5 51547db8db7fc7caf6cca7c0b54c8ee3 len 10566, got % len %. Do not commit.',
          md5(q), length(q);
    END IF;

    -- and the forward direction still works from here, which is what makes this a
    -- rollback rather than a one-way street
    back := replace(replace(replace(replace(q, old1, new1), old2, new2), old3, new3), old4, new4);
    IF md5(back) <> '406bd7571e60981fb062a2e2b3fac515' THEN
        RAISE EXCEPTION '497 ROLLBACK VERIFY: reverted text does not re-apply to 497 cleanly (got % len %)', md5(back), length(back);
    END IF;

    RAISE NOTICE '497 ROLLBACK OK: pre-image restored exactly (md5 % len %), and it re-applies to 497 cleanly.', md5(q), length(q);
END $$;

COMMIT;
