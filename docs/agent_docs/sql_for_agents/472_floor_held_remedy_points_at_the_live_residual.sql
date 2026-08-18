-- 472 — correct 471's onward pointer: `bugs_open/295` DOES NOT EXIST
--
-- Same shape as 471: a single `replace()` against the live `pre_query` of
-- `held-pair-canary-escalation`, so text-only holds by construction. No
-- predicate is touched.
--
-- ============================================================================
-- WHY — I repeated a reference from a handoff without enumerating it
-- ============================================================================
-- 471 (applied ~30 minutes earlier) told the human reading a floor-held
-- escalation that "the real defect is bugs_open/295 (producer families dying on
-- owned pages) — fix THAT, not the handler". Both halves are wrong:
--
--   1. **295 is CLOSED**, and has been since 2026-08-17 21:35 UTC. It lives at
--      `bugs_closed/295_HANDOFF_2026-08-17_the_owned_page_guard_in_save_page_sections_kills_the_work_item_and_files_no_review_row.md`.
--      A human following the payload would have grepped `bugs_open/` and found
--      nothing — a dead end inside an instruction that only fires when someone
--      is already stuck.
--   2. **What 295 fixed is not what the reader needs.** It made the refusal
--      VISIBLE (the guard now files an `owned_page_review` row where there had
--      been zero for all history). It deliberately did NOT stop the item
--      failing, and it did not make the page get repaired.
--
-- The LIVE residual — named in 295 itself as "not addressed here, still open" —
-- is **fix candidate 3: route content findings on owned pages to `section_edit`,
-- which demonstrably works on them (18 completes)**. That is the actionable
-- remedy, and it is what the payload should say.
--
-- I took `bugs_open/295` from my own handoff and carried it into a live payload
-- without checking the path still resolved — the "ground every figure against the
-- live system before repeating it from another doc" rule, missed on a reference
-- rather than a figure. Logged in WRONG_CALLS.md (2026-08-18).
--
-- ⚠ Also folded in, because it is the trap immediately downstream: `295` records
-- that `apply_section_edit` is right for REWRITING an existing component and a
-- **dead end for ADDING a section** to an owned page. Sending someone to the
-- route without that caveat swaps one dead end for another.

BEGIN;

DO $$
DECLARE
    old_q      text;
    n          int;
    old_anchor text := $A$and the real defect is bugs_open/295 (producer families dying on owned pages) — fix THAT, not the handler.$A$;
    new_anchor text := $A$and the handler does not need repairing — the PAGE needs a different ROUTE. See bugs_closed/295 (CLOSED 2026-08-17: the owned-page refusal is no longer silent, it files an owned_page_review row — but that made the refusal VISIBLE, it did not make the page get FIXED, and the item still fails by design). Its fix candidate 3 is UNTOUCHED and is the live remedy: route content findings on owned pages to section_edit, which demonstrably works on them (18 completes). ⚠ apply_section_edit is right for REWRITING an existing component and a DEAD END for ADDING a section to an owned page — grep LANDMINES for apply_section_edit before choosing. Check for an existing owned_page_review row for the page first.$A$;
BEGIN
    SELECT pre_query INTO old_q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF old_q IS NULL THEN
        RAISE EXCEPTION '472: scheduled task held-pair-canary-escalation not found.';
    END IF;

    -- PRECONDITION (can fail): 471's text must be live, exactly once.
    n := (length(old_q) - length(replace(old_q, old_anchor, ''))) / length(old_anchor);
    IF n <> 1 THEN
        RAISE EXCEPTION '472: PRECONDITION FAILED — expected exactly 1 occurrence of 471''s bugs_open/295 pointer, found %. Either 471 is not applied or the task has been revised since; read the live pre_query first.', n;
    END IF;

    -- CONTROL (can fail): the rewritten statement must still parse and plan.
    -- EXPLAIN plans without executing, so nothing is mutated and an undoubled
    -- apostrophe in the new prose aborts the COMMIT.
    EXECUTE 'EXPLAIN ' || replace(old_q, old_anchor, new_anchor);

    UPDATE scheduled_tasks
       SET pre_query = replace(old_q, old_anchor, new_anchor), updated_at = now()
     WHERE name = 'held-pair-canary-escalation';

    RAISE NOTICE '472: floor-held remedy now points at bugs_closed/295 and its LIVE residual (fix candidate 3, route to section_edit), not at a bugs_open path that does not exist.';
END $$;

COMMIT;
