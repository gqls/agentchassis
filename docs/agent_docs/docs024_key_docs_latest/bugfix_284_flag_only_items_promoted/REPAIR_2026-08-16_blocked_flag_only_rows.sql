-- ============================================================================
-- ⚠ SUPERSEDED — ALREADY EXECUTED. DO NOT RUN.
--
-- The repair this file describes was carried out on 2026-08-16 as migration
-- `docs/agent_docs/sql_for_agents/442_repair_flag_only_rows_blocked_by_the_old_promoter.sql`
-- (applied AND recorded in schema_migrations, with its own ROLLBACK sidecar).
-- All 60 rows are repaired and stamped `result.repair_284`; this file's predicate
-- (`status='blocked' AND handler_agent=''`) now matches 0 rows, so running it
-- would be a harmless no-op — but it would also look like the repair had not
-- happened yet.
--
-- WHY THERE ARE TWO: the session that executed the repair wrote 442 without first
-- reading this lane's own directory, where this file was already waiting. That was
-- a process miss, logged in NOTES and WRONG_CALLS.md. The two files reach the same
-- end state by the same discipline; 442 is the one in the ledger, so 442 is the
-- record. Rollback via 442's sidecar, not this file's.
-- ============================================================================

-- REPAIR — bugs_open/284: the 60 flag-only rows wrongly stamped `blocked`.
--
-- ⚠ DO NOT RUN THIS UNTIL THE GUARD IS LIVE. Until `7027a2801` is in the running
-- agent-chassis binary, every row this repairs is promoted again by the OLD
-- promoter and blocked again on the next claim — you will have spent a write and
-- moved nothing. The gate below refuses to run in that case; it is not advice.
--
-- Written to the council's own standard (debug_historian, advisory objection on
-- correlation c22998e8, 2026-08-16): a mechanically-counted pre-state, a guarded
-- idempotent UPDATE, a verify that can FAIL the transaction, and a rollback file
-- beside it (ROLLBACK_2026-08-16_blocked_flag_only_rows.sql).
--
-- PRECONDITION, checked by hand before you open this file — the binary must say so
-- itself, per service, never per fleet:
--
--   kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 \
--     | grep -m1 'build provenance'
--   git merge-base --is-ancestor 7027a2801 <the stamped commit> && echo LIVE
--
-- Then set the marker below to that stamped commit. The transaction ABORTS if you
-- leave it unset — a repair that cannot name the binary it is safe against is the
-- thing this file exists to prevent.

\set ON_ERROR_STOP on

-- ⚠ FIXED 2026-08-17: this was `\set guard_commit ''`, a **psql client** variable,
-- while the DO block below reads `current_setting('myvars.guard_commit')`, a **server**
-- GUC. The two never met, so the gate raised unconditionally and this script could
-- never run even when it should have. It failed CLOSED, which is the safe direction and
-- is why nothing was damaged — but a guard whose passing path was never induced is not
-- a guard, it is a comment that aborts. Logged in WRONG_CALLS 2026-08-17.
-- Set the pod-verified commit as a real session GUC, e.g.:
--   SET myvars.guard_commit = '6a782274b';
SET myvars.guard_commit = '';   -- ← put the pod-verified commit sha here

BEGIN;

-- ── 0. The gate. A SELECT cannot stop a COMMIT (ON_ERROR_STOP ignores a non-empty
--       result set), so this is a DO block that RAISES. LANDMINES, migration guards.
DO $$
BEGIN
    IF current_setting('myvars.guard_commit', true) IS NULL
       OR current_setting('myvars.guard_commit', true) = '' THEN
        RAISE EXCEPTION
            'REFUSING: guard_commit is unset. Verify the running binary carries 7027a2801 '
            '(build provenance line + git merge-base --is-ancestor), then set it. '
            'Repairing before the guard is live re-blocks every row on the next claim.';
    END IF;
END $$;

-- ── 1. PRE-STATE, counted mechanically and kept in the transaction, so the verify
--       at the end compares against a number nobody typed.
CREATE TEMP TABLE repair_284_pre AS
SELECT id, item_type, status, error, spec->>'check' AS producing_check
FROM site_work_items
WHERE status = 'blocked'
  AND handler_agent = ''
  AND error = 'No handler_agent set — item cannot be routed to any agent';

\echo '── pre-state (expected 2026-08-16: capability_gap 18, image_url_404 40) ──'
SELECT item_type, count(*) FROM repair_284_pre GROUP BY 1 ORDER BY 2 DESC;

-- ── 2. A full row-level backup of exactly what we are about to touch. Cheap, and
--       the rollback file reads from it.
CREATE TABLE IF NOT EXISTS repair_284_backup AS
SELECT * FROM site_work_items WHERE 1 = 0;

INSERT INTO repair_284_backup
SELECT * FROM site_work_items WHERE id IN (SELECT id FROM repair_284_pre);

-- ── 3. The repair. Idempotent by construction: the predicate names the state being
--       left, so a second run touches 0 rows. Per TYPE, because the right resting
--       state differs and a single sweep would be wrong for one of them.
--
--       capability_gap -> 'deferred': the status this item TYPE means everywhere
--       else (remit.go's CapabilityGapItem, WriteBuildItemsAction), and where the
--       other 19 healthy rows already sit.
UPDATE site_work_items
SET status = 'deferred', error = NULL, updated_at = now()
WHERE id IN (SELECT id FROM repair_284_pre WHERE item_type = 'capability_gap');

--       image_url_404 -> 'detected': its producers' designed flag-only resting
--       state, which the guard now makes genuinely inert. ⚠ This check has NO
--       retraction path (0 `result.Resolved` sites, unlike its three siblings), so
--       these rows will sit at `detected` until a human acts. That is the check's
--       pre-existing design; do not read a persistent count as the guard failing.
UPDATE site_work_items
SET status = 'detected', error = NULL, updated_at = now()
WHERE id IN (SELECT id FROM repair_284_pre WHERE item_type = 'image_url_404');

-- ── 4. NOT SWEPT, deliberately: the two hand-inserted rows (page_rerender,
--       needs_experience_plan — no spec.original_pipeline, no triaged_at,
--       created_by naming a session). They came from the SECOND path, and their
--       sessions may have meant them to dispatch — in which case the repair is a
--       handler_agent, not a status. Judge them individually:
--
--   SELECT id, item_type, summary, created_by, spec FROM repair_284_pre p
--     JOIN site_work_items w USING (id)
--    WHERE p.item_type NOT IN ('capability_gap','image_url_404');

-- ── 5. VERIFY, and it must be able to FAIL. Three properties: every targeted row
--       moved; nothing outside the target set moved; no row of ours is still
--       carrying the routing error.
DO $$
DECLARE
    still_blocked int;
    landed        int;
    stray_error   int;
BEGIN
    SELECT count(*) INTO still_blocked
    FROM site_work_items w JOIN repair_284_pre p USING (id)
    WHERE p.item_type IN ('capability_gap','image_url_404') AND w.status = 'blocked';

    SELECT count(*) INTO landed
    FROM site_work_items w JOIN repair_284_pre p USING (id)
    WHERE (p.item_type = 'capability_gap' AND w.status = 'deferred')
       OR (p.item_type = 'image_url_404'  AND w.status = 'detected');

    SELECT count(*) INTO stray_error
    FROM site_work_items w JOIN repair_284_pre p USING (id)
    WHERE p.item_type IN ('capability_gap','image_url_404')
      AND w.error = 'No handler_agent set — item cannot be routed to any agent';

    IF still_blocked > 0 THEN
        RAISE EXCEPTION 'VERIFY FAILED: % targeted row(s) still blocked', still_blocked;
    END IF;
    IF stray_error > 0 THEN
        RAISE EXCEPTION 'VERIFY FAILED: % row(s) still carry the routing error', stray_error;
    END IF;
    IF landed <> (SELECT count(*) FROM repair_284_pre
                   WHERE item_type IN ('capability_gap','image_url_404')) THEN
        RAISE EXCEPTION 'VERIFY FAILED: % landed, % targeted — a row went somewhere unintended',
            landed, (SELECT count(*) FROM repair_284_pre
                      WHERE item_type IN ('capability_gap','image_url_404'));
    END IF;

    RAISE NOTICE 'VERIFY PASSED: % row(s) repaired, 0 still blocked, 0 stray errors', landed;
END $$;

-- ── 5b. INDUCE THE GATE BOTH WAYS before trusting it — the omission that made the
--       original version of this file inert. It must ABORT with the GUC empty, and
--       PROCEED with it set:
--         SET myvars.guard_commit = '';            -- expect: REFUSING …
--         SET myvars.guard_commit = '6a782274b';   -- expect: no refusal
--       A gate only ever tested on its failing side cannot be distinguished from one
--       that always fails.

-- ── 6. The control that proves the verify could have failed. Induce it once, on a
--       throwaway transaction, BEFORE you trust the pass:
--         BEGIN; UPDATE site_work_items SET status='blocked'
--          WHERE id = (SELECT id FROM repair_284_pre LIMIT 1);
--         -- re-run block 5 → must RAISE. ROLLBACK;

COMMIT;

-- ── 7. After the commit, the standing check: nothing should ever return to this
--       state once the guard is live. A non-zero result here later is a REGRESSION,
--       or the second (hand-insert) path firing again.
--
--   SELECT item_type, count(*) FROM site_work_items
--    WHERE status='blocked' AND handler_agent=''
--    GROUP BY 1;
