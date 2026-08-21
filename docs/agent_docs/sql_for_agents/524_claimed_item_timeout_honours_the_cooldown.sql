-- 524 — `claimed-item-timeout`'s own retry ladder learns the cooldown, so the
-- fifth writer stops re-triaging a timed-out item with no wait.
--
-- ── HOLD RELEASED 2026-08-21 ────────────────────────────────────────────────
-- Held from 2026-08-21 morning until now. The condition was `bugs_open/344`'s
-- candidate 1 being **LIVE — rolled and verified at the artefact**, not merely
-- committed (a distinction that would have released this a day early: the fix was
-- committed in the morning and the owner deferred the roll until the evening).
--
-- CONDITION MET: v1.0.1322 rolled; the whole chassis fleet reports one stamp
-- (`bac189921`, 59 pods) and `0f80f5ea1` is an ancestor of it. Verified live: the
-- close canary's attempts 1 and 2 re-triaged with their cooldown stamps and
-- SURVIVED the loop's completion call, and the natural census reads ZERO
-- false-green rows (`retry_after > completed_at`) since the roll.
--
-- So a `retry_after` stamped by THIS sweep can no longer be overwritten by
-- `mark_complete` seconds later, which is the only reason the stamp was withheld.
--
-- ⚠ Note what this file does NOT need, corrected from `bugs_open/341` §5b before
-- release: the sweep's two auto-COMPLETE arms do NOT need the completion
-- predicate. All three of its arms carry `WHERE wi.status = 'claimed'`, and a
-- ladder-re-triaged row is `triaged` with its claim columns cleared by the ladder
-- itself — measured: 0 rows at `status='claimed'` carry any `retry_after`, and 0
-- false-green rows are attributable to this sweep. Adding it would be dead SQL.
--
-- Idempotent: the pre-state check treats an already-applied pre_query as success.

BEGIN;

DO $do$
DECLARE
  v_pq      text;
  v_anchor  text := E'        END\n    WHERE status = ''claimed''\n      AND claimed_at < NOW() - INTERVAL ''40 minutes''';
  v_new     text;
  v_hits    int;
BEGIN
  SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name = 'claimed-item-timeout';
  IF v_pq IS NULL THEN
    RAISE EXCEPTION '524: scheduled_tasks row claimed-item-timeout not found';
  END IF;

  -- Already applied? Then stop, successfully.
  IF v_pq LIKE '%retry_after%' THEN
    RAISE NOTICE '524: pre_query already honours the cooldown — nothing to do';
    RETURN;
  END IF;

  -- PRE-STATE GATE: the fragment must exist exactly once, or the text this file was
  -- written against has changed and a blind edit would corrupt it.
  SELECT count(*) INTO v_hits FROM regexp_matches(v_pq, regexp_replace(v_anchor, '([().\[\]*+?^$|\\])', '\\\1', 'g'), 'g');
  IF v_hits <> 1 THEN
    RAISE EXCEPTION '524: ABORTING — the reset CTE anchor occurs % times, expected exactly 1. Another lane has edited claimed-item-timeout.pre_query since 2026-08-21; re-read the live column and re-derive this file. Do NOT force.', v_hits;
  END IF;

  -- THE EDIT. `retry_after` is stamped only on the non-terminal arm, and the minutes
  -- come from reaper_policies exactly as the Go ladder reads them (per-item_type row
  -- overriding the queue-wide '__default__', code default 30) — scaled by the attempt
  -- being consumed, so 30m then 60m on a max_attempts=3 item. NULL on the terminal arm
  -- because a `failed` row is not waiting for anything.
  -- NOTE THE COMMA after the error CASE's END. The first version of this file
  -- omitted it and produced syntactically invalid SQL that EVERY string-matching
  -- check below still passed — see the parse check at the end, which is why it
  -- exists.
  v_new := replace(
    v_pq,
    v_anchor,
    E'        END,\n'
    '        retry_after = CASE\n'
    '            WHEN attempt_count + 1 >= max_attempts THEN NULL\n'
    '            ELSE NOW() + make_interval(mins =>\n'
    '                 COALESCE((SELECT rp.backoff_minutes FROM reaper_policies rp\n'
    '                            WHERE rp.queue = ''site_work_items''\n'
    '                              AND rp.item_type IN (site_work_items.item_type, ''__default__'')\n'
    '                            ORDER BY (rp.item_type = site_work_items.item_type) DESC\n'
    '                            LIMIT 1), 30) * (attempt_count + 1))\n'
    '        END\n'
    || E'    WHERE status = ''claimed''\n      AND claimed_at < NOW() - INTERVAL ''40 minutes''');

  IF v_new = v_pq THEN
    RAISE EXCEPTION '524: replace() changed nothing despite a unique anchor match — refusing to record a no-op as applied';
  END IF;

  UPDATE scheduled_tasks SET pre_query = v_new, updated_at = now()
   WHERE name = 'claimed-item-timeout';
END
$do$;

-- ── Verify (DO/RAISE — a SELECT cannot stop a COMMIT) ────────────────────────
DO $do$
DECLARE
  v_pq text;
BEGIN
  SELECT pre_query INTO v_pq FROM scheduled_tasks WHERE name = 'claimed-item-timeout';

  IF v_pq NOT LIKE '%retry_after = CASE%' THEN
    RAISE EXCEPTION '524: the reset CTE does not stamp retry_after';
  END IF;
  IF v_pq NOT LIKE '%reaper_policies%' THEN
    RAISE EXCEPTION '524: the backoff is not read from reaper_policies — a literal would be the third hand-rolled copy';
  END IF;

  -- The clauses that were already load-bearing must survive the edit. A replace()
  -- cannot drop them, but asserting it is what makes that a fact rather than a claim.
  IF v_pq NOT LIKE '%completed_by_orchestration%'
     OR v_pq NOT LIKE '%completed_by_evidence%'
     OR v_pq NOT LIKE '%INTERVAL ''40 minutes''%'
     OR v_pq NOT LIKE '%item_type NOT IN%' THEN
    RAISE EXCEPTION '524: the edit lost a pre-existing clause of the sweep — it must differ by the retry_after stamp ALONE';
  END IF;

  IF to_regclass('reaper_policies') IS NULL THEN
    RAISE EXCEPTION '524: reaper_policies is missing — migration 335 must be applied first';
  END IF;
  IF NOT EXISTS (SELECT 1 FROM information_schema.columns
                  WHERE table_name = 'site_work_items' AND column_name = 'retry_after') THEN
    RAISE EXCEPTION '524: site_work_items.retry_after is missing — migration 505 must be applied first';
  END IF;

  -- ── THE PARSE CHECK, and it is the load-bearing one ────────────────────────
  -- Every assertion above is a STRING match, and a pre_query is DATA to this
  -- migration: it parses only when the scheduled task next RUNS, 120 s later, in
  -- a job nobody is watching. The first version of this file inserted the new
  -- assignment without a comma after the error CASE's END — invalid SQL — and
  -- ALL of the LIKE checks above passed on it. A sweep that fails to parse does
  -- not raise anything a human sees; it simply stops reclaiming timed-out items,
  -- which looks exactly like "no items are timing out".
  --
  -- EXPLAIN parses and plans without executing, even for data-modifying CTEs, so
  -- it is a genuine syntax gate and it is safe here.
  BEGIN
    EXECUTE 'EXPLAIN ' || v_pq;
  EXCEPTION WHEN others THEN
    RAISE EXCEPTION '524: the rewritten pre_query DOES NOT PARSE (%). The sweep would silently stop reclaiming timed-out items. SQLSTATE %', SQLERRM, SQLSTATE;
  END;
END
$do$;

COMMIT;
