-- 499 — held-pair-canary-escalation: `498`'s candidate route is WRONG for the population
--       it is attached to, and the corrected guidance is a TEST, not a target
--
-- ============================================================================
-- WHY, AND IT IS THE THIRD CORRECTION TO THIS ONE STRING IN ONE DAY
-- ============================================================================
-- `498` (08:40Z) replaced `497`'s stale figures with a candidate route:
--
--   "CANDIDATE ROUTE, measured 2026-08-20: section_edit to section-editor, 36 complete /
--    1 failed of 39 on rebuild_policy=owned pages lifetime incl. archive (295 fix
--    candidate 3). Target the SECTION component that carries the prose, not a
--    component_level=tool fork."
--
-- The 36/1 figure is correct and the route genuinely works on owned pages. **The
-- instruction attached to it is wrong for `literal_markdown` on the pages that produced
-- this residual**, and it is wrong in the one direction that matters: it names as the
-- target the single component that CANNOT be re-rendered.
--
-- [MEASURED 2026-08-20 ~09:30Z] Reading the actual findings on the 7 rows, which I had
-- not done when I wrote `498`:
--
--   * all 7 have `source: rendered_html`, `pattern: code_span`, `slot: ported-page`,
--     `field: (empty — rendered_html)`. The matches are backticked code tokens in ported
--     technical prose: `fetch()`, `feTurbulence`, `ease-in-out`, `33%`.
--   * the detector scans BOTH surfaces (`literalMarkdownFinding.Source` is
--     `content_data | rendered_html`). These fired on the SECOND one.
--   * the `ported-page` component's `content_data` is `{schema, sha256, source, qa_tier,
--     generator}` — 215 bytes of METADATA. There is no prose in it.
--   * its template's only field is `{{.body}}`, and `body` is not a key.
--
-- **So every content_data-based repair is inapplicable BY CONSTRUCTION**, including both
-- routes this lane has proposed: `473`'s rerender-from-content_data, and `498`'s
-- `section_edit` + `strip_literal_markdown` (which calls
-- `StripLiteralMarkdownFromContentData` on a map holding no prose). The content lives
-- ONLY in `page_components.rendered_html`.
--
-- PROVEN, not reasoned. The real template rendered against the real `content_data`, using
-- production's own engine and option (`text/template`, `Option("missingkey=zero")`,
-- `component_library.go:861`): **4,665 bytes out, 188 in the body region, ZERO
-- non-whitespace visible characters, `err=<nil>`.** The control — the same template with a
-- generic page's `content_data`, which DOES carry `body` — renders 11,035 bytes with 6,568
-- of prose. Same component, same template, two payloads, opposite results.
--
-- ============================================================================
-- WHAT IS **NOT** CLAIMED — the alarming version of this is wrong too
-- ============================================================================
-- This is NOT "100 pages are one edit away from being blanked". `apply_section_edit` calls
-- `enforceSingleSlotFloors` (`section_editor_actions.go:451` ->
-- `single_slot_floors.go:161`), whose axis is VISIBLE text with style and script content
-- excluded, engaging above `minShrinkGuardVisibleChars` (200) on the existing side. Against
-- thousands of existing visible characters and ZERO incoming, it refuses and writes
-- nothing: *"apply_section_edit: SLOT FLOOR REFUSED ... the existing component still
-- stands."* So the outcome of following `498`'s instruction would be a **third refusal
-- mode**, not damage. The pages are protected; they are simply not repairable this way.
--
-- ============================================================================
-- THE CORRECTED GUIDANCE IS A TEST, NOT A TARGET
-- ============================================================================
-- `497` named a dead pointer. `498` named stale figures. Both were corrected by writing a
-- better VALUE. This one is corrected by writing a **question the reader must answer about
-- their own case**, because the failure was not a wrong value — it was a right value
-- applied to a population it does not fit. A named target is only ever right for the
-- population the author was looking at; a test travels.
--
-- The test: **read the finding's `source`, then ask whether the target component's
-- `content_data` can reproduce its `rendered_html`** — i.e. whether the template's fields
-- are actually keys in `content_data`. If the finding is in `rendered_html` and the
-- content lives only there, NO content_data-based route can repair it, however good that
-- route's record looks on other pages.
--
-- ============================================================================
-- SCOPE / CONTROLS / ROLLBACK
-- ============================================================================
-- ONE string literal in the `owners` VALUES CTE. No predicate, clock, threshold, CTE or
-- row selection. Same pattern as `479`/`497`/`498`: one anchor asserted unique, md5-guarded,
-- forward and reverse from the same two variables, reverse-replacement control returning
-- the 13,746-character body to `498`'s exact md5, and mutation-proven. Live at the next
-- daily tick. No binary dependency.

BEGIN;

DO $$
DECLARE
    q    text;
    back text;
    live_md5 text;
    n    int;
    old2 text;   -- 498's string
    new2 text;   -- this one
BEGIN
    old2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. FIRST SPLIT THE HELD ROWS BY pages.rebuild_policy: the reason field beside this one describes the PAIR, and an owned-page row is not waiting for a route - it is REFUSED BY DESIGN by the generic page-build repair, so it will never be repaired by being re-pointed at one. WORKED INSTANCE, 2026-08-20 07:20-07:24Z, written up in 277 and 333: this pair rose above the promoter floor, the promoter released the 7 owned-page rows it had been holding, and every one was refused by OWNED_PAGE_GUARD and terminated wont_fix. Note what that costs you as a reader: wont_fix is excluded from BOTH sides of the promoter rule (301), and idx_swi_dedup excludes it too so the detector re-files the same finding (333) - the loop therefore leaves no trace in this pair record, and a healthy-looking ratio here is not evidence that owned pages are being repaired. CANDIDATE ROUTE, measured 2026-08-20: section_edit to section-editor, 36 complete / 1 failed of 39 on rebuild_policy=owned pages lifetime incl. archive (295 fix candidate 3). Target the SECTION component that carries the prose, not a component_level=tool fork. section_edit REWRITES an existing component and cannot ADD one, which suits literal markdown. UNVERIFIED: whether a producer can file a section_edit item for this finding shape - that is 277 question.'$a$;

    new2 := $a$'TWO different bugs, both OPEN. ROUTING: 333_producers_route_content_findings_at_page_build_handler_without_reading_rebuild_policy. REPAIR DESIGN: 277_required_fields_missing_has_no_repair_handler_fleet_wide, lane docs024_key_docs_latest/bugfix_277_required_fields_repair. Bugs 184 and 201, named here until 2026-08-19, are BOTH CLOSED. BEFORE YOU ROUTE ANYTHING, RUN THIS TEST ON YOUR OWN ROWS - it is what three corrections to this entry in one day were about. (1) READ THE FINDING SOURCE. spec findings carry source = content_data OR rendered_html; the detector scans both. (2) ASK WHETHER content_data CAN REPRODUCE rendered_html for the target component - are the template fields actually keys in content_data? If the finding is in rendered_html and the content lives only there, NO content_data-based route can repair it, however good that route record looks elsewhere. WORKED INSTANCE, 2026-08-20: all 7 owned-page rows held here were source=rendered_html, pattern=code_span, slot=ported-page - backticked code tokens in ported prose. That component content_data holds only metadata and its template only field is body, which is not a key; rendering it produces ZERO visible characters (measured against production own engine, missingkey=zero). So BOTH routes this lane proposed were inapplicable by construction: 473 rerender-from-content_data, and section_edit with strip_literal_markdown, which strips a map holding no prose. Not dangerous, just useless - enforceSingleSlotFloors measures VISIBLE text and refuses at zero, leaving the component standing. NOT A GENERAL VERDICT: section_edit to section-editor really is 36 complete / 1 failed of 39 on owned pages, and is the right route where content_data CAN fill the template. The predicate is not ownership, it is whether the content is reachable from content_data at all.'$a$;

    SELECT pre_query, md5(pre_query) INTO q, live_md5
      FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF q IS NULL THEN
        RAISE EXCEPTION '499: no held-pair-canary-escalation row';
    END IF;

    IF live_md5 <> 'ddd0c894845fa77edcd2d9f213a4d890' THEN
        RAISE EXCEPTION
          '499: ABORTING — the live pre_query is not what 498 produced (expected md5 ddd0c894845fa77edcd2d9f213a4d890 len 13746, found % len %). Re-read the live column and re-derive; do NOT force.',
          live_md5, length(q);
    END IF;

    n := (length(q) - length(replace(q, old2, ''))) / length(old2);
    IF n <> 1 THEN RAISE EXCEPTION '499: the 498 literal_markdown owner string occurs % times, expected 1', n; END IF;

    UPDATE scheduled_tasks
       SET pre_query = replace(pre_query, old2, new2)
     WHERE name = 'held-pair-canary-escalation';

    SELECT pre_query INTO q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';

    IF position(old2 in q) <> 0 THEN RAISE EXCEPTION '499 VERIFY: the 498 string is still present'; END IF;
    IF position(new2 in q) =  0 THEN RAISE EXCEPTION '499 VERIFY: the new string is absent'; END IF;

    -- the wrong instruction must be GONE, named so this cannot pass by accident
    IF position('Target the SECTION component that carries the prose' in q) <> 0 THEN
        RAISE EXCEPTION '499 VERIFY: the misleading "target the SECTION component" instruction survives';
    END IF;
    -- and the corrected test must be PRESENT
    IF position('ASK WHETHER content_data CAN REPRODUCE rendered_html' in q) = 0 THEN
        RAISE EXCEPTION '499 VERIFY: the corrected content-surface test was not written in';
    END IF;
    -- the 36/1 record must SURVIVE, explicitly not retracted (it is true and useful)
    IF position('36 complete / 1 failed of 39 on owned pages' in q) = 0 THEN
        RAISE EXCEPTION '499 VERIFY: the section_edit record was dropped — it is correct and must stay';
    END IF;

    -- 497's other corrections, untouched — positive controls
    IF position('083_detected_findings_never_reach_a_handler (OPEN)' in q) = 0 THEN
        RAISE EXCEPTION '499 VERIFY: 497 placeholder_contact owner damaged';
    END IF;
    IF position('300_drift_findings_key_on_an_unstable_component_id' in q) = 0 THEN
        RAISE EXCEPTION '499 VERIFY: 497 drift owner damaged';
    END IF;
    n := (length(q) - length(replace(q, 'bugs_open/300', ''))) / length('bugs_open/300');
    IF n <> 2 THEN RAISE EXCEPTION '499 VERIFY: expected the 2 live bugs_open/300 citations to survive, found %', n; END IF;

    -- THE LOAD-BEARING CONTROL
    back := replace(q, new2, old2);
    IF md5(back) <> 'ddd0c894845fa77edcd2d9f213a4d890' THEN
        RAISE EXCEPTION
          '499 VERIFY: REVERSE-REPLACEMENT CONTROL FAILED. Undoing does not reproduce 498 output (expected ddd0c894845fa77edcd2d9f213a4d890 len 13746, got % len %). Something OTHER than the one intended string changed. Do not commit.',
          md5(back), length(back);
    END IF;

    RAISE NOTICE '499 OK: the wrong target replaced by the content-surface TEST, the 36/1 record kept, 497 corrections intact, reverse control reproduces 498 exactly. New md5 % len %.', md5(q), length(q);
END $$;

COMMIT;
