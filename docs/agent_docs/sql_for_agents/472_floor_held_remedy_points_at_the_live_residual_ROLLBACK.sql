-- 472 ROLLBACK — restore 471's pointer.
-- ⚠ Reverting reinstates a reference to `bugs_open/295`, which DOES NOT EXIST
-- (295 closed 2026-08-17 and lives in bugs_closed/). There is no reason to run
-- this except to unwind 472 as part of unwinding 471.
BEGIN;
DO $$
DECLARE
    old_q text; n int;
    new_anchor text := $A$and the handler does not need repairing — the PAGE needs a different ROUTE. See bugs_closed/295 (CLOSED 2026-08-17: the owned-page refusal is no longer silent, it files an owned_page_review row — but that made the refusal VISIBLE, it did not make the page get FIXED, and the item still fails by design). Its fix candidate 3 is UNTOUCHED and is the live remedy: route content findings on owned pages to section_edit, which demonstrably works on them (18 completes). ⚠ apply_section_edit is right for REWRITING an existing component and a DEAD END for ADDING a section to an owned page — grep LANDMINES for apply_section_edit before choosing. Check for an existing owned_page_review row for the page first.$A$;
    old_anchor text := $A$and the real defect is bugs_open/295 (producer families dying on owned pages) — fix THAT, not the handler.$A$;
BEGIN
    SELECT pre_query INTO old_q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    n := (length(old_q) - length(replace(old_q, new_anchor, ''))) / length(new_anchor);
    IF n <> 1 THEN
        RAISE EXCEPTION '472 ROLLBACK: expected exactly 1 occurrence of 472''s text, found %.', n;
    END IF;
    UPDATE scheduled_tasks SET pre_query = replace(old_q, new_anchor, old_anchor), updated_at = now()
     WHERE name = 'held-pair-canary-escalation';
    RAISE NOTICE '472 ROLLBACK: 471 pointer restored (and it names a path that does not exist).';
END $$;
COMMIT;
