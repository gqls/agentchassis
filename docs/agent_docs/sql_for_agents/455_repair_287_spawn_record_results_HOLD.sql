-- 455 — bugs_open/287 (spawn_record): repair the historical work-item results whose
-- true reply is STILL RECOVERABLE, and mark what was replaced
--
-- ⚠⚠ HOLD: NEEDS THE OWNER'S GO. This is a bulk UPDATE of historical rows in a live
-- table; the lane's README put the choice to the owner with counts rather than
-- assuming it (repair the recoverable minority, or leave history annotated and rely
-- on the reader warnings). Nothing here is required to stop the defect — that is
-- already done and proven live by migrations 448/452.
--
-- WHAT IS BROKEN: between the 2026-08-15 10:14Z roll and migration 452 (applied
-- 2026-08-17 16:28:57Z), dispatch-loop completions stored `spawn_agent`'s ack —
-- {role, topics, agent_id, agent_type} — as `site_work_items.result` instead of the
-- handler's reply. The work itself happened; only the record is someone else's.
--
-- THE POPULATION, MEASURED 2026-08-17 ~16:45Z (re-run the same CTE before applying —
-- it grew from 2,259 at 12:40 to 3,330 as pre-452 runs drained):
--   * 3,330 items carry the spawn-record shape;
--   * 303 of them can be repaired, and this file repairs EXACTLY those.
--   * The remaining ~3,027 are NOT recoverable: their parent orchestration's
--     `collected_data` no longer holds the reply (bugs_open/289's aggregation
--     rewrote these keys; long-lived runs lose them). They stay as they are, and the
--     287 §6a reader guidance plus the LANDMINE cover them: never read an item's
--     `result` as evidence of what the work did — verify at the artefact.
--
-- WHY THE JOIN IS EXACT, NOT PROBABILISTIC (this is the part worth reviewing): the
-- spawn record's own `topics.requests` string embeds the parent correlation's first 8
-- hex and the iteration index, e.g.
--   job.<corr8>-<x>-<handler>-process_item_iter_<N>_spawn_handler.requests
-- An 8-hex prefix alone COULD collide, so it is not trusted on its own: every row is
-- additionally required to satisfy
--   parent.collected_data->('process_item_item_'||N)->>'id' = item.id
-- i.e. the parent must itself name THIS item as the one it processed in THAT
-- iteration. Under that predicate all 303 candidate rows matched and all 303 carry a
-- `response` at `handler_result_<N>` — so the reply being written back is the reply
-- that parent got for this item, per row, not a nearest-neighbour guess.
--
-- WHAT IT WRITES: `result` becomes the recovered `handler_result_<N>` object (which is
-- `call_agent`'s step record with the handler's reply at `.response` — the same shape
-- post-452 completions now store), PLUS two provenance keys:
--   `_replaced_spawn_record` — the exact value being overwritten, so this is reversible
--                              from the row itself (the ROLLBACK sidecar uses it);
--   `_repaired_by`           — '455' and the date, so a reader can tell a repaired row
--                              from one that was recorded correctly at the time.
--   `_completed_at_before_repair` — the row's true completion time, because the
--                              `updated_at` trigger cannot be held (see next block).
-- Nothing else on the row changes: status, error, item_key, item_type are untouched.
--
-- ⚠ THE `updated_at` TRIGGER IS UNCONDITIONAL — MEASURED, NOT ASSUMED (2026-08-17):
--   trg on site_work_items = `BEGIN NEW.updated_at = NOW(); RETURN NEW; END;`
-- so `SET updated_at = updated_at` CANNOT hold the timestamp; the trigger overwrites
-- it. My first draft of this file asserted the timestamp had NOT moved and would
-- therefore have RAISED and aborted on every run — caught by reading the trigger and
-- dry-running in a rolled-back transaction before this file was ever offered for
-- approval (lane NOTES, 2026-08-17).
--
-- WHAT THAT COSTS, AND WHAT IS DONE ABOUT IT. Three readers key on `updated_at`:
--   (1) the hourly item census in bugfix_287_spawn_record/RUNBOOK — 303 repaired rows
--       would appear as completions in the repair hour, which would look exactly like
--       a spike of correctly-recorded completions and could be mistaken for evidence
--       the fix worked. **After applying this, add `AND NOT (result ? '_repaired_by')`
--       to that census.** The repair must not be able to flatter the fix.
--   (2) staleness/reaper reporting (bugs_closed/213's mechanism keys on `updated_at`).
--       These rows are `complete` — terminal — so no reaper claims them; the effect is
--       on reports, not on routing.
--   (3) anything answering "what completed recently".
-- Mitigation taken instead of fighting the trigger: the row keeps its true completion
-- time inside the result as `_completed_at_before_repair`, so no information is lost
-- and (1) and (3) can recover it per row.
--
-- IF THE OWNER PREFERS TIMESTAMP FIDELITY over avoiding DDL, the alternative is to
-- wrap the UPDATE in `ALTER TABLE site_work_items DISABLE TRIGGER <name>;` …
-- `ENABLE TRIGGER`, inside this same transaction (DDL is transactional, so an abort
-- restores it). NOT chosen here: for the duration of the statement it would also
-- suppress the bump for every CONCURRENT writer on a table the dispatch loop writes
-- to constantly, which trades a reporting artefact on 303 known rows for a silent one
-- on an unknown set.
--
-- Idempotent: fenced on the spawn-record shape still being present; a replay matches
-- no row. Rollback sidecar restores from `_replaced_spawn_record`.

BEGIN;

-- The needles, computed once so the UPDATE and the verify agree on the count.
CREATE TEMP TABLE repair_287_needles AS
WITH bad AS (
    SELECT id, updated_at,
           substring(result->'topics'->>'requests' from '^job\.([0-9a-f]{8})') AS corr8,
           substring(result->'topics'->>'requests' from '_iter_([0-9]+)_')     AS iter
    FROM site_work_items
    WHERE status = 'complete'
      AND result ? 'topics' AND result ? 'agent_id' AND result ? 'agent_type'
)
SELECT b.id                                              AS item_id,
       b.updated_at                                      AS updated_at_before,
       os.collected_data -> ('handler_result_' || b.iter) AS recovered_reply
FROM bad b
JOIN orchestration_states os
  ON left(os.correlation_id::text, 8) = b.corr8
 AND b.iter IS NOT NULL
 AND os.collected_data -> ('process_item_item_' || b.iter) ->> 'id' = b.id::text
WHERE os.collected_data -> ('handler_result_' || b.iter) ? 'response';

SELECT count(*) AS needles_found FROM repair_287_needles;

UPDATE site_work_items w
   SET result = n.recovered_reply
                || jsonb_build_object(
                     '_replaced_spawn_record', w.result,
                     '_repaired_by', '455_repair_287_spawn_record_results (2026-08-17)',
                     '_completed_at_before_repair', to_jsonb(n.updated_at_before)
                   )
  FROM repair_287_needles n
 WHERE w.id = n.item_id
   AND w.result ? 'topics' AND w.result ? 'agent_id';   -- replay fence

-- Verify (DO/RAISE — a SELECT cannot stop the COMMIT). Three post-conditions:
-- every needle repaired; none still spawn-shaped; every row kept its true
-- completion time (the trigger's bump is expected and is NOT treated as a failure).
DO $$
DECLARE expected int; repaired int; still_bad int; moved int;
BEGIN
    SELECT count(*) INTO expected FROM repair_287_needles;
    IF expected = 0 THEN
        RAISE EXCEPTION '455: zero needles — either already applied, or the join predicate no longer matches. Re-measure before forcing.';
    END IF;

    SELECT count(*) INTO repaired
      FROM repair_287_needles n JOIN site_work_items w ON w.id = n.item_id
     WHERE w.result ? '_replaced_spawn_record' AND w.result ? 'response';
    IF repaired <> expected THEN
        RAISE EXCEPTION '455: repaired % of % needles', repaired, expected;
    END IF;

    SELECT count(*) INTO still_bad
      FROM repair_287_needles n JOIN site_work_items w ON w.id = n.item_id
     WHERE w.result ? 'topics' AND w.result ? 'agent_id';
    IF still_bad > 0 THEN
        RAISE EXCEPTION '455: % repaired rows still carry the spawn-record shape', still_bad;
    END IF;

    -- The trigger bumps updated_at unconditionally (header). Assert the true
    -- completion time survived on every repaired row instead of forbidding the bump.
    SELECT count(*) INTO moved
      FROM repair_287_needles n JOIN site_work_items w ON w.id = n.item_id
     WHERE NOT (w.result ? '_completed_at_before_repair');
    IF moved > 0 THEN
        RAISE EXCEPTION '455: % rows lost their pre-repair completion time', moved;
    END IF;

    RAISE NOTICE '455: repaired % rows; updated_at was bumped by the trigger, true time kept at result->_completed_at_before_repair', repaired;
END $$;

INSERT INTO doc_notes (subject_type, subject_key, body, categories, source, created_by)
SELECT 'pipeline', 'build',
$note$## 455: repaired the recoverable spawn-record work-item results — bugs_open/287
Between the 2026-08-15 roll and migration 452, dispatch-loop completions stored spawn_agent's ack as site_work_items.result instead of the handler's reply. This repaired the subset whose reply is still recoverable from the parent orchestration, joined EXACTLY (the parent must name this item as the one it processed in that iteration), writing the recovered reply plus `_replaced_spawn_record` (reversible from the row) and `_repaired_by`. The larger remainder is NOT recoverable — their parents' collected_data no longer holds the reply — and stays annotated: never read an item's result as evidence of what the work did; verify at the artefact.
Categories: migration$note$,
'["migration"]'::jsonb, 'agent', 'bugfix-287-spawn-record-lane'
WHERE NOT EXISTS (
    SELECT 1 FROM doc_notes WHERE body LIKE '## 455: repaired the recoverable spawn-record work-item results%'
);

COMMIT;
