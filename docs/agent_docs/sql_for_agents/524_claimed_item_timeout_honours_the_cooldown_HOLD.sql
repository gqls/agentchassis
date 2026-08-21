-- 524 — `claimed-item-timeout`'s own retry ladder learns the cooldown, so the
-- fifth writer stops re-triaging a timed-out item with no wait.
--
-- ⚠⚠ _HOLD — DO NOT APPLY YET. Built, parse-checked and committed; deliberately
-- withheld. The name is the mechanism: `run-migrations.sh`'s SIDECAR_RE excludes
-- `_HOLD.sql` from `--apply` while still listing it, because a banner in a comment
-- cannot stop another lane's sweep and an ordering-critical file must not depend on
-- someone reading the top of it.
--
-- RELEASE CONDITION: `bugs_open/344` resolved — or its candidate 1 (completion
-- refuses while `retry_after` is in the future) live.
--
-- WHY, and it is an interaction rather than a fault in this file. 344: the dispatch
-- loop's `mark_complete` overwrites a ladder-re-triaged item to `complete` ~2 s after
-- the failure write, because the child ends via a success-labelled `complete_workflow`
-- and `triaged` is not in the completion guard. That was harmless before
-- `bugs_open/307`, and is load-bearing now that the ladder makes `triaged` the
-- post-failure state. The queryable fingerprint is `retry_after > completed_at`.
--
-- THIS FILE WOULD WIDEN THAT. Stamping `retry_after` from the sweep creates exactly
-- the row shape 344 destroys — so applying it before 344 is decided converts a defect
-- confined to the Go ladder into one that also reaches every claim timeout. The fix
-- is correct; the ORDER is what is wrong, and holding is cheaper than discovering it
-- in a census later.
--
-- Release checklist: (1) 344 closed or candidate 1 live; (2) re-run the pre-state
-- gate below — this file was derived against the pre_query as at 2026-08-21 and
-- ABORTS if another lane has edited it since; (3) rename off `_HOLD`, apply, record.
--
-- bugs_open/341 candidate 2, narrowed on measurement (see the CORRECTION below).
-- bugs_open/307 built one failure-write contract in Go and this SQL sweep — enabled,
-- every 120 s — kept its own copy of the ladder outside it. Of the ways the two
-- diverged, exactly ONE is a real defect, and this file closes that one.
--
-- ── CORRECTION TO bugs_open/341, WHICH I WROTE ─────────────────────────────────
-- 341 lists "no terminal-DECISION guard" as a defect of this sweep, on the grounds
-- that its predicate is "`WHERE status='claimed'` only". That is wrong, and it is
-- wrong in the safe direction: `status = 'claimed'` is STRICTLY STRONGER than the
-- decision guard for this purpose. A row a handler has deliberately parked is, by
-- definition, no longer `claimed` — so this UPDATE cannot reach it, and the race is
-- closed by the WHERE evaluating at execution time rather than by any list. The Go
-- path needs an explicit `status NOT IN (…)` because it is handed a work_item_id and
-- writes whatever it finds; this sweep SELECTS its own population and has already
-- excluded every decision status by construction.
--
-- So this file adds NO guard. Adding one would be ~7 statuses of dead SQL implying a
-- risk that does not exist here, which is its own kind of misinformation.
--
-- ── WHAT IS GENUINELY DIVERGENT, and what this fixes ──────────────────────────
--   * NO COOLDOWN  ← fixed here. The reset re-triages with `retry_after` unset, so a
--     timed-out item is immediately re-claimable. During an infrastructure outage
--     that means ~40-minute cycles burning the attempt budget against a dead
--     dependency: slower than the pre-307 Go path's instant re-claim, but the same
--     defect, and the same ending (`failed` at max_attempts inside the outage).
--     [MEASURED 2026-08-19] 23 `failed` rows in 14 days carry this sweep's own
--     `Claim timed out%` error; 18 are `tool-auditor` at attempt_count=3.
--   * NO TRANSIENT RELEASE — NOT fixed here, and cannot be: burst detection reads
--     `agent_error_log` and layers two Go classifiers. Re-expressing that in SQL is a
--     second implementation of one rule, which is the drift class 307 exists to end.
--     A claim timeout during an outage therefore still spends an attempt — but now it
--     spends them 30 and 60 minutes apart instead of back to back, which is what
--     candidate 2 was for. The rest stays with 341 candidate 1.
--   * LEAVES `handled_by` NULL — deliberately unchanged. Writing it would make the
--     census honest going forward but would change what `handled_by IS NULL` means
--     mid-stream for every query already written against it (307 §2.2 and two other
--     lanes). `error NOT LIKE 'Claim timed out%'` remains the documented discriminator.
--
-- ── HOW IT EDITS, and why not as a rewrite ────────────────────────────────────
-- A verbatim-anchored `replace()`, not a full-value overwrite. The council's
-- `debug_historian` seat objected (corr 4cdec68b, high) that migration 506 replaced an
-- embedded query wholesale with no pre-state check, in tables with drift history. This
-- file answers that by CONSTRUCTION rather than by a gate: it can only touch the exact
-- fragment it names, it asserts that fragment occurs exactly once BEFORE writing, and
-- it never retypes the other ~6 kB of the pre_query — so a concurrent edit elsewhere in
-- that column survives untouched instead of being silently reverted.
--
-- NOT named `*_claimed_item_timeout_generic_evidence.sql` ON PURPOSE:
-- `claim_timeout_exclusion_lockstep_test.go` globs that pattern, takes the
-- lexicographically newest and reads its FIRST `item_type NOT IN (…)` as the live
-- exclusion list. This file does not touch the exclusion list, so it must not present
-- itself as the newest authority on it — `482` stays that file.
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
