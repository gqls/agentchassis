-- 497 — held-pair-canary-escalation: all THREE entries in the `owners` map are stale,
--       and the map is stamped into a human's row ONCE, at escalation time
--
-- 453/466/479 are ledger-recorded and are NOT edited. This file rewrites the live
-- `scheduled_tasks.held-pair-canary-escalation` pre_query by SURGICAL REPLACEMENT on
-- four verbatim anchors, each asserted to occur exactly once, guarded on the whole
-- text's md5, and — see CONTROLS — proved to have changed NOTHING ELSE by putting the
-- old strings back and asserting the md5 returns to the pre-image. It transcribes
-- nothing: the 10,566-character body, most of it 466's `what_to_do` prose, is carried
-- through untouched.
--
-- ============================================================================
-- THE DEFECT
-- ============================================================================
-- `owners` maps an item_type to the lane that should receive its escalation. Its value
-- is written into `result.held_pair_escalation.owner` by the `escalated` UPDATE — i.e.
-- it is stamped into the row at the moment of escalation and NEVER REVISITED. A later
-- migration does not retro-fix a row that has already escalated. So a wrong entry is
-- not "stale config"; it is a wrong instruction handed to a person, once, permanently,
-- at the only moment they were going to read it.
--
-- [MEASURED 2026-08-19 21:15Z] All three entries are wrong, in two different ways:
--
--   1. `placeholder_contact`  -> 'bugs_open/201 lane ...'
--   2. `literal_markdown`     -> 'bugs_open/184 + bugs_open/201 lane ...'
--      `bugs_open/201` and `bugs_open/184` DO NOT EXIST. Verified at HEAD, not with
--      `ls`: `git ls-tree -r --name-only HEAD -- bugs_open/ bugs_closed/ | grep -E
--      '/(184|201)_'` returns three paths, all under `bugs_closed/`. (184 is one of
--      the documented ambiguous numbers — it names TWO unrelated cases; both closed.)
--   3. `page_component_status_drift` -> '(UNASSIGNED - claim this) ... no lane doc
--      claims it'. HALF true and half false, so it is corrected rather than replaced:
--      `check_page_component_status_drift.go` really is untouched since 2026-07-10
--      (`git log` on the file: one commit, 7813f3eb9, 2026-07-10) — but the ITEM TYPE
--      is now owned by `bugs_open/300`, which is open with its fix live and council-
--      approved. This is the entry with a PROVEN consumer: the only escalation this
--      family has ever produced (`page_component_status_drift -> component-template-
--      fixer`, 2026-08-17T12:57:43Z) carries exactly this string in its `result`, so a
--      human was already told "nobody owns this" about a bug that has a lane.
--
-- ============================================================================
-- THE STRUCTURAL HALF — why the entries rotted, and what stops it recurring
-- ============================================================================
-- The map hardcodes a `bugs_open/` PREFIX. That prefix is precisely the thing that
-- flips when the lane being pointed at SUCCEEDS. So this map was built to go stale at
-- exactly the moment its pointer starts to matter, and it did: 184 and 201 both closed
-- while findings were still queued against them. Correcting the three strings alone
-- would re-arm the same trap for the next close, so replacement 4 writes the rule into
-- the map itself: name a bug by NUMBER and SLUG, never by directory; name the lane
-- directory, which does not move.
--
-- ============================================================================
-- WHAT IS AND IS NOT CHANGED — the behavioural surface is NIL
-- ============================================================================
-- CHANGED: four string literals — three `VALUES` entries in a lookup CTE, plus an
-- inserted SQL comment. Nothing else, and the CONTROLS below prove it mechanically.
-- NOT CHANGED: no predicate, no clock, no threshold, no CTE, no row selection. This
-- migration cannot alter WHICH rows escalate, WHEN they escalate, or WHAT HAPPENS to
-- them. It alters only the sentence a human is handed. Deliberately so: two ticks are
-- imminent (`placeholder_contact`, 3 rows, 2026-08-20 12:57; `literal_markdown`, 7
-- rows, 2026-08-21 12:57) and they need the pointer fixed BEFORE they fire, so the
-- change made under that clock is the one that cannot move a row.
--
-- The `reason` field is left ALONE, and that is a decision, not an omission. For the
-- literal_markdown 7 it says "the pair succeeds below 25%, so the promoter has stopped
-- feeding it" — true of the PAIR, and not why THOSE rows are stuck. Making `reason`
-- say so would mean reading `rebuild_policy` inside this task: real new logic, on the
-- very seam `bugs_open/333` was filed for on 2026-08-19. So the correction rides in
-- the `owner` string, which is prose and costs nothing, and the logic question goes to
-- 333 where it belongs.
--
-- ============================================================================
-- CONTROLS — the reverse-replacement test, which is the load-bearing one
-- ============================================================================
-- A verify block that asserts "the new strings are present" is nearly vacuous: it
-- passes identically whether or not the replacement also mangled 10 KB of `what_to_do`
-- prose, which is the ONLY damage this file could plausibly do. So the verify block
-- puts each NEW string back to its OLD one and asserts the md5 returns to the
-- pre-image `51547db8db7fc7caf6cca7c0b54c8ee3`. That can fail — it fails if any other
-- byte moved — and it is what makes "nothing else changed" evidence rather than an
-- intention. It has already earned its place: it caught an asymmetric undo in this
-- file's own first draft, before the file was ever run.
--
-- Forward and reverse are driven from the SAME EIGHT VARIABLES, declared once. A
-- control written out by hand a second time tests the transcription, not the change.
--
-- ============================================================================
-- ORDER / ROLLBACK
-- ============================================================================
-- DB config: live at the task's next daily tick (12:57 UTC) after COMMIT. No binary
-- dependency, so no roll is needed and no image tag is involved. Rollback is a separate
-- file (497_..._ROLLBACK.sql), which reverses the same four anchors and asserts the
-- md5 returns to this pre-image. Reverting does NOT rewrite the `owner` string on any
-- row escalated in the meantime — by the same one-shot property that makes this
-- migration worth applying before the tick.

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
    -- ------------------------------------------------------------------
    -- the eight strings, declared ONCE and used by both directions
    -- ------------------------------------------------------------------
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

    -- ------------------------------------------------------------------
    -- GUARD: refuse unless the live text is exactly what this was written against
    -- ------------------------------------------------------------------
    SELECT pre_query, md5(pre_query) INTO q, live_md5
      FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q IS NULL THEN
        RAISE EXCEPTION '497: no held-pair-canary-escalation row — nothing to correct';
    END IF;

    IF live_md5 <> '51547db8db7fc7caf6cca7c0b54c8ee3' THEN
        RAISE EXCEPTION
          '497: ABORTING — the live pre_query is not the text this migration was written against (expected md5 51547db8db7fc7caf6cca7c0b54c8ee3 len 10566, found % len %). Another lane has edited it since 2026-08-19 21:15Z. Re-read the live column, re-derive this file against it, and do NOT force: overwriting is how one lane silently reverts another.',
          live_md5, length(q);
    END IF;

    -- Each anchor must occur EXACTLY once. A replace() that hits the wrong place, or
    -- nothing at all, still leaves valid SQL and a green-looking apply.
    n := (length(q) - length(replace(q, old1, ''))) / length(old1);
    IF n <> 1 THEN RAISE EXCEPTION '497: anchor 1 (placeholder_contact owner) occurs % times, expected 1', n; END IF;
    n := (length(q) - length(replace(q, old2, ''))) / length(old2);
    IF n <> 1 THEN RAISE EXCEPTION '497: anchor 2 (literal_markdown owner) occurs % times, expected 1', n; END IF;
    n := (length(q) - length(replace(q, old3, ''))) / length(old3);
    IF n <> 1 THEN RAISE EXCEPTION '497: anchor 3 (page_component_status_drift owner) occurs % times, expected 1', n; END IF;
    n := (length(q) - length(replace(q, old4, ''))) / length(old4);
    IF n <> 1 THEN RAISE EXCEPTION '497: anchor 4 (owners CTE header) occurs % times, expected 1', n; END IF;

    -- ------------------------------------------------------------------
    -- APPLY: the four replacements, in one statement
    -- ------------------------------------------------------------------
    UPDATE scheduled_tasks
       SET pre_query = replace(replace(replace(replace(pre_query, old1, new1), old2, new2), old3, new3), old4, new4)
     WHERE name = 'held-pair-canary-escalation';

    -- ------------------------------------------------------------------
    -- VERIFY. RAISE, not SELECT — a plain SELECT cannot stop the COMMIT.
    -- ------------------------------------------------------------------
    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    -- (a) cheap first-fails: old text gone, new text in
    IF position(old1 in q) <> 0 THEN RAISE EXCEPTION '497 VERIFY: old placeholder_contact owner still present'; END IF;
    IF position(old2 in q) <> 0 THEN RAISE EXCEPTION '497 VERIFY: old literal_markdown owner still present'; END IF;
    IF position(old3 in q) <> 0 THEN RAISE EXCEPTION '497 VERIFY: old drift owner still present'; END IF;
    IF position(new1 in q) =  0 THEN RAISE EXCEPTION '497 VERIFY: new placeholder_contact owner absent'; END IF;
    IF position(new2 in q) =  0 THEN RAISE EXCEPTION '497 VERIFY: new literal_markdown owner absent'; END IF;
    IF position(new3 in q) =  0 THEN RAISE EXCEPTION '497 VERIFY: new drift owner absent'; END IF;
    IF position(new4 in q) =  0 THEN RAISE EXCEPTION '497 VERIFY: the directory-prefix rule was not written into the map'; END IF;

    -- (b) the structural claim, asserted rather than asserted-about: the DEAD pointers
    --     are gone and only the LIVE one survives.
    --
    --     ⚠ A bare count of 'bugs_open/' is the WRONG assertion here, and it fired on
    --     this file's own first run: the rule comment inserted by replacement 4 says
    --     "NEVER write a bugs_open/ ... PREFIX", so the REMEDY TEXT contains the token
    --     the check was counting. Counting a token cannot distinguish a live pointer,
    --     a dead one, and a prohibition against writing one. So assert the three
    --     specific things that are actually claimed:
    n := (length(q) - length(replace(q, 'bugs_open/201', ''))) / length('bugs_open/201');
    IF n <> 0 THEN RAISE EXCEPTION '497 VERIFY: bugs_open/201 (CLOSED) still referenced % time(s)', n; END IF;

    n := (length(q) - length(replace(q, 'bugs_open/184', ''))) / length('bugs_open/184');
    IF n <> 0 THEN RAISE EXCEPTION '497 VERIFY: bugs_open/184 (CLOSED) still referenced % time(s)', n; END IF;

    --     bugs_open/300 is genuinely open and is cited twice in `what_to_do`; those two
    --     citations are CORRECT and must survive untouched. This is the positive
    --     control: it fails if the replacements ate the prose they must not touch.
    n := (length(q) - length(replace(q, 'bugs_open/300', ''))) / length('bugs_open/300');
    IF n <> 2 THEN RAISE EXCEPTION '497 VERIFY: expected the 2 live bugs_open/300 citations in what_to_do to survive, found %', n; END IF;

    -- (c) THE LOAD-BEARING CONTROL. Put the old strings back; the md5 must return to
    --     the pre-image. This is the only assertion here that can catch collateral
    --     damage to the 10 KB of `what_to_do` prose this file never intends to touch —
    --     every check above passes identically whether or not that prose survived.
    back := replace(replace(replace(replace(q, new1, old1), new2, old2), new3, old3), new4, old4);

    IF md5(back) <> '51547db8db7fc7caf6cca7c0b54c8ee3' THEN
        RAISE EXCEPTION
          '497 VERIFY: REVERSE-REPLACEMENT CONTROL FAILED. Undoing the four replacements does not reproduce the pre-image (expected md5 51547db8db7fc7caf6cca7c0b54c8ee3 len 10566, got % len %). Something OTHER than the four intended strings changed. Do not commit.',
          md5(back), length(back);
    END IF;

    RAISE NOTICE '497 OK: owners map corrected (3 entries), directory-prefix rule written in, reverse-replacement control reproduces the pre-image exactly. New md5 % len %.', md5(q), length(q);
END $$;

COMMIT;
