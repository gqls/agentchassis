-- 471 — the floor-held escalation remedy must not send a human to fix a handler
--       that is REFUSING ON PURPOSE
--
-- Amends ONE prose string inside `held-pair-canary-escalation`.pre_query (set by
-- 466). No predicate is touched: the change is applied as a single `replace()`
-- against the live text, so "text-only" holds BY CONSTRUCTION rather than by a
-- control that could not have come out otherwise. (This lane has now written
-- three tautological controls — 430, 453, 466 — so the shape is deliberate.)
--
-- ============================================================================
-- WHY
-- ============================================================================
-- 466(b) gave floor-held pairs their own remedy text, ending:
--     "FIX THE HANDLER, or decide the pair is wrong and retire the producer."
-- That instruction is wrong for the very pair that is about to receive it.
--
-- [MEASURED 2026-08-18, 948 `failed` rows, site_work_items UNION archive — the
-- same lifetime population the floor reads after 465]
--     protective refusal (handler declined on purpose) .. 434  45.8%
--        of which `rebuild_policy=owned` .................. 418
--     transient / infra ................................. 234  24.7%
--     housekeeping / backfill (never an attempt) ........ 110  11.6%
--     GENUINE non-repair ................................ 170  17.9%
-- and 17.9% is an UPPER bound: ~9 rows are misfiled into it (4 "diagnosis
-- exceeded 75m", 5 "claims floor blocked"), both corrections pushing the same way.
--
-- The floor counts all 948 identically. `literal_markdown -> page-build-handler`
-- — 10 rows at `detected`, the ONLY floor-held pair with rows waiting — is
-- 3 successes / 36 failures, of which **16 are protective refusals** and 16
-- genuine. It escalates on the **2026-08-21 12:57Z** tick and would tell a human
-- to fix a handler that, in 44% of its failures, is doing exactly what it should.
--
-- The pair is still CORRECTLY held (3/(3+16) = 16%, under the 25% floor), and
-- **no pair anywhere flips verdict** under a refusal-excluding floor when the
-- promoter's FULL predicate is used (`c = 0 OR under-floor`) — measured, 0 flips.
-- So this migration deliberately does NOT change the gate's arithmetic.
--
-- ============================================================================
-- WHY THE CLASSIFIER IS NOT BEING PUT IN THE GATE
-- ============================================================================
-- Encoding `error ILIKE '%rebuild_policy=owned%'` into a live gate would make an
-- error message's WORDING load-bearing fleet-wide: reword the string, silently
-- change what the promoter dispatches, with no test that fails. That is one rung
-- worse than the "a source-scan test makes your COMMENTS load-bearing" family,
-- because the coupling crosses services. The sound fix is a STRUCTURED refusal
-- signal (a distinct terminal status, or a `result` key the handler sets), so the
-- gate reads an assertion instead of a sentence — a new shared vocabulary on a
-- shared seam, i.e. architecture-scope (owner rulings 2026-07-28 / 07-29), to be
-- taken with bugs_open/295 and NOT smuggled into a fourth revision of a pre_query.
--
-- In ADVISORY prose read by a human the same brittleness is tolerable: a
-- misclassification reads oddly, it cannot mis-gate anything. So the remedy text
-- tells the reader to partition the failures themselves and hands them the query.

BEGIN;

DO $$
DECLARE
    old_q      text;
    new_q      text;
    n          int;
    n_target   int;
    old_anchor text := $A$FIX THE HANDLER, or decide the pair is wrong and retire the producer.$A$;
    new_anchor text := $A$FIRST PARTITION THE FAILURES — the floor counts EVERY failed row alike, and most of them are not the handler failing. [MEASURED 2026-08-18, 948 failed rows across site_work_items + site_work_items_archive] only ~18% are genuine non-repairs: 46% are the handler CORRECTLY REFUSING (rebuild_policy=owned, overwrite REFUSED, section shrink), 25% transient/infra, 12% housekeeping that was never an attempt. Run this for THIS pair before blaming anything: SELECT count(*), left(error,90) FROM (SELECT item_type,handler_agent,status,error FROM site_work_items UNION ALL SELECT item_type,handler_agent,status,error FROM site_work_items_archive) z WHERE status=''failed'' AND item_type=<the pair item_type> AND handler_agent=<the pair handler_agent> GROUP BY 2 ORDER BY 1 DESC. If protective refusals dominate, the handler is behaving CORRECTLY, the floor is mis-holding this pair, and the real defect is bugs_open/295 (producer families dying on owned pages) — fix THAT, not the handler. Only if genuine non-repairs dominate: FIX THE HANDLER, or decide the pair is wrong and retire the producer. Method and the fleet-wide partition: docs024_key_docs_latest/bugfix_277_required_fields_repair/NOTES_required_fields_repair.md (2026-08-18).$A$;
BEGIN
    SELECT pre_query INTO old_q FROM scheduled_tasks WHERE name = 'held-pair-canary-escalation';
    IF old_q IS NULL THEN
        RAISE EXCEPTION '471: scheduled task held-pair-canary-escalation not found (or pre_query NULL). Nothing to amend.';
    END IF;

    -- PRECONDITION (can fail): the live text must be 466's, with the anchor
    -- appearing exactly once. If another session has revised this task since,
    -- the count changes and this stops rather than silently editing the wrong text.
    n := (length(old_q) - length(replace(old_q, old_anchor, ''))) / length(old_anchor);
    IF n <> 1 THEN
        RAISE EXCEPTION '471: PRECONDITION FAILED — expected exactly 1 occurrence of 466''s floor-held anchor, found %. The live pre_query is not the text this migration was written against; re-read it before applying.', n;
    END IF;

    new_q := replace(old_q, old_anchor, new_anchor);

    -- CONTROL 1, and it is the one that matters (can fail): the rewritten
    -- pre_query must still PARSE AND PLAN. EXPLAIN plans without executing, so
    -- this validates the entire statement — including that every apostrophe in
    -- the new prose is correctly doubled, which is the realistic way to break a
    -- string edit inside a nested SQL literal — while mutating NOTHING.
    -- An un-doubled quote makes this a syntax error and aborts the COMMIT.
    EXECUTE 'EXPLAIN ' || new_q;

    UPDATE scheduled_tasks
       SET pre_query = new_q, updated_at = now()
     WHERE name = 'held-pair-canary-escalation';

    -- CONTROL 2, POSITIVE (can fail): the text just fixed must be REACHABLE —
    -- i.e. some pair is actually floor-held, with rows waiting, or this
    -- migration corrects a string nothing will ever emit. The named case is
    -- `literal_markdown -> page-build-handler` (10 rows on 2026-08-18, escalating
    -- on the 08-21 12:57Z tick). If the pair has since been repaired or promoted
    -- this RAISEs, which is the correct outcome: re-measure before applying.
    SELECT count(*) INTO n_target
      FROM site_work_items wi
      CROSS JOIN LATERAL (
          SELECT count(*) FILTER (WHERE status IN ('complete','verified')) AS c,
                 count(*) FILTER (WHERE status = 'failed')                 AS f
          FROM (SELECT item_type, handler_agent, status FROM site_work_items
                UNION ALL
                SELECT item_type, handler_agent, status FROM site_work_items_archive) hh
          WHERE hh.item_type = wi.item_type AND hh.handler_agent = wi.handler_agent) h
     WHERE wi.status = 'detected'
       AND wi.item_type = 'literal_markdown'
       AND wi.handler_agent = 'page-build-handler'
       AND h.c > 0
       AND NOT ((h.c + h.f) < 5 OR h.c >= 0.25 * (h.c + h.f));

    IF n_target = 0 THEN
        RAISE EXCEPTION '471: POSITIVE CONTROL FAILED — no floor-held rows for literal_markdown->page-build-handler, so the remedy text this migration fixes is unreachable and the change is untested. Re-measure the held set before applying.';
    END IF;

    RAISE NOTICE '471: floor-held remedy now tells the reader to PARTITION the failures before blaming the handler. Text-only (single replace against the live pre_query); new statement parses and plans. Reachable: % row(s) floor-held for literal_markdown->page-build-handler, escalating on the 2026-08-21 12:57Z tick.', n_target;
END $$;

COMMIT;
