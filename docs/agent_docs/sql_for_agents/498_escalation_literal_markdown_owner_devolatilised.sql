-- 498 — held-pair-canary-escalation: de-volatilise the `literal_markdown` owner string
--       that `497` wrote this morning, because three of its figures went stale in TWELVE HOURS
--
-- ============================================================================
-- WHY THIS EXISTS, AND IT IS NOT A NICE ONE
-- ============================================================================
-- `497` (applied 2026-08-20 ~08:00Z, four hours ago) fixed a map that had rotted because it
-- hardcoded a `bugs_open/` PREFIX — a value guaranteed to go stale at the moment the lane it
-- named succeeded. In the SAME migration I wrote into the SAME live config:
--
--   "[MEASURED 2026-08-19] all 7 rows held on page-build-handler sit on rebuild_policy=owned
--    pages ... The newer literal_markdown to page-rerender route stands at 8 complete on
--    generic pages and 1 failed on the single owned page it tried, so re-pointing these 7
--    would produce 7 more failures, drag a healthy pair toward its floor, and repair nothing."
--
-- Every clause of that was true when measured and THREE of them were false within twelve hours:
--
--   1. "all 7 rows HELD" — they are not held. [MEASURED 2026-08-20 08:11Z] between 07:20:42Z
--      and 07:23:58Z this morning all 7 were dispatched, refused by `OWNED_PAGE_GUARD`, and
--      terminated `wont_fix`. The held set for this pair is now EMPTY and the escalation the
--      lane had docketed for 2026-08-21 12:57Z will not fire.
--   2. "a HEALTHY pair" / the floor — the pair was 3 ok / 34 failed (8.1%, FLOOR-HELD) last
--      night and is 19 ok / 24 failed (44%, PROMOTABLE) now. 16 completions landed across 3
--      sites inside the 07:00Z hour. Its release is WHY the 7 were dispatched into the guard.
--   3. "drag a healthy pair toward its floor" — it cannot, and `301` is the reason. A refusal
--      now terminates `wont_fix`, which is excluded from BOTH the numerator and the denominator
--      of the promoter's rule (`bugs_closed/301` §, and `bugs_open/333` counts the terminations).
--      So the prediction's mechanism was already prevented by a fix that shipped two days ago.
--
-- THE LESSON, and it is the same lesson as `497`'s: **`497` fixed a self-staling POINTER and in
-- the same breath wrote self-staling FIGURES.** The distinction that matters is not
-- measured-vs-unmeasured — all of it was measured and dated. It is:
--
--   * a DATED OBSERVATION stays true for ever ("on 08-20 07:20Z, 7 rows were refused"), and
--   * a DESCRIPTION OF CURRENT STATE ("all 7 rows are held") is false the moment the state moves,
--     and a reader has no way to tell which kind they are looking at.
--
-- A one-shot annotation read by a human months later must carry the first kind and the durable
-- MECHANISM, never the second. That is what this migration changes it to.
--
-- ============================================================================
-- WHAT ELSE THIS CARRIES — the answer to the question the old string got wrong
-- ============================================================================
-- The old string said only what NOT to do. [MEASURED 2026-08-20] there is a candidate route and
-- it is the one `466`'s own `what_to_do` already names (`bugs_closed/295` fix candidate 3):
-- `section_edit` -> `section-editor` is **36 complete / 1 failed of 39 on `rebuild_policy=owned`
-- pages**, lifetime incl. archive — against `what_to_do`'s conservative "18 completes".
--
-- And the trap that would have made this another wrong recommendation was checked, not assumed.
-- `LANDMINES.md` records that a `section_edit` on a per-site TOOL FORK whose template carries
-- `{{.field}}` copy and whose `content_data` is `{}` re-renders every text node to EMPTY while
-- every floor passes. Six of the seven pages are `tool-*`, so that is a direct hit on shape.
-- It does NOT fire here, and for a reason worth writing down: the trap needs BOTH halves, and
-- on all 7 pages the `component_level='tool'` fork has `content_data='{}'` but **zero**
-- `{{.field}}` hits in its template — there is nothing for an empty `content_data` to fail to
-- fill. Better still, the literal markdown is not in the tool fork at all: it is in the
-- `ported-page` slot, a `component_level='section'` component whose `content_data` IS populated.
-- That is the ordinary, well-trodden target, i.e. the 36/1 population.
--
-- ⚠ The FIRST check I ran for that was near-vacuous and is recorded so nobody repeats it:
-- grepping `page_components.rendered_html` for `{{.` returns 0 whether or not the risk exists,
-- because `rendered_html` is the RENDERED OUTPUT — a field that resolved to empty leaves no
-- `{{.` behind either. The template is the only place the question can be asked.
--
-- STILL UNVERIFIED and stated as such in the new string: whether a producer can file a
-- `section_edit` item for a `literal_markdown`-shaped finding at all. That is `277`'s question.
--
-- ============================================================================
-- SCOPE / CONTROLS / ROLLBACK
-- ============================================================================
-- ONE string literal in the `owners` VALUES CTE. No predicate, clock, threshold, CTE or row
-- selection; it cannot change which rows escalate, when, or what happens to them. Same proven
-- pattern as `479`/`497`: one anchor asserted to occur exactly once, guarded on the whole text's
-- md5, forward and reverse driven from the same two variables, and a reverse-replacement control
-- that returns the 13,106-character body to `497`'s exact md5 — the only check that can catch
-- collateral damage to the ~10 KB of `what_to_do` prose. Rollback is a separate file.
--
-- Applies live at the next daily tick (12:57 UTC). No binary dependency.

BEGIN;

DO $$
DECLARE
    q    text;
    back text;
    live_md5 text;
    n    int;
    old2 text;
    new2 text;
BEGIN
    -- 497's string (the one being replaced)
    old2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. NB THE reason FIELD UNDERSTATES THIS PAIR: its floor text is true of the PAIR and is not why THESE rows are stuck. [MEASURED 2026-08-19] all 7 rows held on page-build-handler sit on rebuild_policy=owned pages, and the generic repair refuses an owned page BY DESIGN. The newer literal_markdown to page-rerender route stands at 8 complete on generic pages and 1 failed on the single owned page it tried, so re-pointing these 7 would produce 7 more failures, drag a healthy pair toward its floor, and repair nothing. The live candidate is the one already named in the what_to_do ELSE branch below: 295 fix candidate 3, route owned-page content findings to section_edit.'$a$;

    -- the de-volatilised replacement: durable mechanism + DATED observations only
    new2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. FIRST SPLIT THE HELD ROWS BY pages.rebuild_policy: the reason field beside this one describes the PAIR, and an owned-page row is not waiting for a route - it is REFUSED BY DESIGN by the generic page-build repair, so it will never be repaired by being re-pointed at one. WORKED INSTANCE, 2026-08-20 07:20-07:24Z, written up in 277 and 333: this pair rose above the promoter floor, the promoter released the 7 owned-page rows it had been holding, and every one was refused by OWNED_PAGE_GUARD and terminated wont_fix. Note what that costs you as a reader: wont_fix is excluded from BOTH sides of the promoter rule (301), and idx_swi_dedup excludes it too so the detector re-files the same finding (333) - the loop therefore leaves no trace in this pair record, and a healthy-looking ratio here is not evidence that owned pages are being repaired. CANDIDATE ROUTE, measured 2026-08-20: section_edit to section-editor, 36 complete / 1 failed of 39 on rebuild_policy=owned pages lifetime incl. archive (295 fix candidate 3). Target the SECTION component that carries the prose, not a component_level=tool fork. section_edit REWRITES an existing component and cannot ADD one, which suits literal markdown. UNVERIFIED: whether a producer can file a section_edit item for this finding shape - that is 277 question.'$a$;

    -- GUARD
    SELECT pre_query, md5(pre_query) INTO q, live_md5
      FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q IS NULL THEN
        RAISE EXCEPTION '498: no held-pair-canary-escalation row';
    END IF;

    IF live_md5 <> '406bd7571e60981fb062a2e2b3fac515' THEN
        RAISE EXCEPTION
          '498: ABORTING — the live pre_query is not what 497 produced (expected md5 406bd7571e60981fb062a2e2b3fac515 len 13106, found % len %). Either 497 was not applied or another lane has edited the column since 2026-08-20 08:11Z. Re-read the live column and re-derive; do NOT force.',
          live_md5, length(q);
    END IF;

    n := (length(q) - length(replace(q, old2, ''))) / length(old2);
    IF n <> 1 THEN RAISE EXCEPTION '498: the 497 literal_markdown owner string occurs % times, expected 1', n; END IF;

    -- APPLY
    UPDATE scheduled_tasks
       SET pre_query = replace(pre_query, old2, new2)
     WHERE name = 'held-pair-canary-escalation';

    -- VERIFY
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF position(old2 in q) <> 0 THEN RAISE EXCEPTION '498 VERIFY: the 497 string is still present'; END IF;
    IF position(new2 in q) =  0 THEN RAISE EXCEPTION '498 VERIFY: the new string is absent'; END IF;

    -- the three stale claims must be GONE, named individually so this cannot pass by accident
    IF position('all 7 rows held on page-build-handler' in q) <> 0 THEN
        RAISE EXCEPTION '498 VERIFY: the stale "all 7 rows held" current-state claim survives';
    END IF;
    IF position('drag a healthy pair toward its floor' in q) <> 0 THEN
        RAISE EXCEPTION '498 VERIFY: the stale "drag a healthy pair toward its floor" prediction survives (301 prevents it)';
    END IF;
    IF position('8 complete on generic pages and 1 failed' in q) <> 0 THEN
        RAISE EXCEPTION '498 VERIFY: the stale page-rerender figures survive';
    END IF;

    -- 497's other two corrections must be UNTOUCHED — positive controls
    IF position('083_detected_findings_never_reach_a_handler (OPEN)' in q) = 0 THEN
        RAISE EXCEPTION '498 VERIFY: 497 placeholder_contact owner damaged';
    END IF;
    IF position('300_drift_findings_key_on_an_unstable_component_id' in q) = 0 THEN
        RAISE EXCEPTION '498 VERIFY: 497 drift owner damaged';
    END IF;
    IF position('NEVER write a bugs_open/ or bugs_closed/ PREFIX in this map (497).' in q) = 0 THEN
        RAISE EXCEPTION '498 VERIFY: 497 directory-prefix rule damaged';
    END IF;
    n := (length(q) - length(replace(q, 'bugs_open/300', ''))) / length('bugs_open/300');
    IF n <> 2 THEN RAISE EXCEPTION '498 VERIFY: expected the 2 live bugs_open/300 citations in what_to_do to survive, found %', n; END IF;

    -- THE LOAD-BEARING CONTROL: put it back and the md5 must return to 497's exact output.
    back := replace(q, new2, old2);
    IF md5(back) <> '406bd7571e60981fb062a2e2b3fac515' THEN
        RAISE EXCEPTION
          '498 VERIFY: REVERSE-REPLACEMENT CONTROL FAILED. Undoing the replacement does not reproduce 497 output (expected md5 406bd7571e60981fb062a2e2b3fac515 len 13106, got % len %). Something OTHER than the one intended string changed. Do not commit.',
          md5(back), length(back);
    END IF;

    RAISE NOTICE '498 OK: literal_markdown owner de-volatilised, three stale claims gone, 497 other corrections intact, reverse-replacement control reproduces 497 output exactly. New md5 % len %.', md5(q), length(q);
END $$;

COMMIT;
