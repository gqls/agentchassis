-- 471 ROLLBACK — restore 466's floor-held remedy sentence.
--
-- ⚠ Reverting reinstates a MEASURED misdirection: the floor-held escalation
-- payload goes back to "FIX THE HANDLER, or decide the pair is wrong and retire
-- the producer" for a pair whose failures are 44% the handler CORRECTLY refusing
-- to clobber owned pages (`literal_markdown -> page-build-handler`, 16 of 36
-- protective, measured 2026-08-18). Nothing about the GATE changes either way —
-- 471 is text-only — so the only thing reverting buys is a shorter payload.
--
-- Applies the inverse single replace(), and asserts the 471 text is present
-- exactly once first, so it cannot silently no-op or edit a third revision.

BEGIN;

DO $$
DECLARE
    old_q      text;
    n          int;
    new_anchor text := $A$FIRST PARTITION THE FAILURES — the floor counts EVERY failed row alike, and most of them are not the handler failing. [MEASURED 2026-08-18, 948 failed rows across site_work_items + site_work_items_archive] only ~18% are genuine non-repairs: 46% are the handler CORRECTLY REFUSING (rebuild_policy=owned, overwrite REFUSED, section shrink), 25% transient/infra, 12% housekeeping that was never an attempt. Run this for THIS pair before blaming anything: SELECT count(*), left(error,90) FROM (SELECT item_type,handler_agent,status,error FROM site_work_items UNION ALL SELECT item_type,handler_agent,status,error FROM site_work_items_archive) z WHERE status=''failed'' AND item_type=<the pair item_type> AND handler_agent=<the pair handler_agent> GROUP BY 2 ORDER BY 1 DESC. If protective refusals dominate, the handler is behaving CORRECTLY, the floor is mis-holding this pair, and the real defect is bugs_open/295 (producer families dying on owned pages) — fix THAT, not the handler. Only if genuine non-repairs dominate: FIX THE HANDLER, or decide the pair is wrong and retire the producer. Method and the fleet-wide partition: docs024_key_docs_latest/bugfix_277_required_fields_repair/NOTES_required_fields_repair.md (2026-08-18).$A$;
    old_anchor text := $A$FIX THE HANDLER, or decide the pair is wrong and retire the producer.$A$;
BEGIN
    SELECT pre_query INTO old_q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF old_q IS NULL THEN
        RAISE EXCEPTION '471 ROLLBACK: scheduled task held-pair-canary-escalation not found.';
    END IF;

    n := (length(old_q) - length(replace(old_q, new_anchor, ''))) / length(new_anchor);
    IF n <> 1 THEN
        RAISE EXCEPTION '471 ROLLBACK: expected exactly 1 occurrence of 471''s text, found %. Either 471 was never applied or the task has been revised since; read the live pre_query before reverting.', n;
    END IF;

    UPDATE scheduled_tasks
       SET pre_query = replace(old_q, new_anchor, old_anchor), updated_at = now()
     WHERE name = 'held-pair-canary-escalation';

    RAISE NOTICE '471 ROLLBACK: 466 floor-held remedy text restored. The misdirection described in 471 is back.';
END $$;

COMMIT;
